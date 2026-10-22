package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestChecksAreWellFormed(t *testing.T) {
	ids := map[string]bool{}
	for _, ch := range checks {
		if ch.ID == "" || ids[ch.ID] || strings.ContainsAny(ch.ID, " /\\") {
			t.Errorf("bad or duplicate check ID %q", ch.ID)
		}
		ids[ch.ID] = true
		if id, _, ok := conceptAt(ch.Concept); !ok || id != ch.Stage {
			t.Errorf("%s: concept %q is not in Stage %d", ch.ID, ch.Concept, ch.Stage)
		}
		if !strings.HasSuffix(ch.File, ".c") || strings.ContainsAny(ch.File, "/\\") {
			t.Errorf("%s: file %q must be a plain .c name", ch.ID, ch.File)
		}
		if ch.Title == "" || ch.Task == "" || ch.Starter == "" || len(ch.Cases) == 0 || len(ch.Hints) == 0 {
			t.Errorf("%s: title, task, starter, cases and hints are all required", ch.ID)
		}
		names := map[string]bool{}
		for _, c := range ch.Cases {
			if c.Name == "" || names[c.Name] {
				t.Errorf("%s: missing or duplicate case name %q", ch.ID, c.Name)
			}
			names[c.Name] = true
			if c.WantExit != 0 && c.Want != "" {
				t.Errorf("%s/%s: an error case should expect no standard output", ch.ID, c.Name)
			}
		}
		if ch.Harness != "" && strings.Contains(ch.Starter, "main(") {
			t.Errorf("%s: the checker supplies main, so the starter must not have one", ch.ID)
		}
	}
	if _, ok := checkByID("hello"); !ok || len(checksFor(1)) != len(checks) {
		t.Error("lookup helpers")
	}
}

func TestOutputDiff(t *testing.T) {
	for _, tc := range []struct{ want, got, problem string }{
		{"a\nb\n", "a\nb\n", ""},
		{"a\nb\n", "a  \r\nb\n\n\n", ""}, // invisible differences are ignored
		{"a\nb\n", "a\nc\n", `line 2 should be "b", but was "c"`},
		{"a\nb\n", "a\n", `output stopped early: line 2 should be "b"`},
		{"a\n", "a\nb\n", `extra output from line 2: "b"`},
		{"", "oops\n", `printed "oops", but nothing was expected on standard output`},
	} {
		if p := outputDiff(tc.want, tc.got); p != tc.problem {
			t.Errorf("outputDiff(%q, %q) = %q, want %q", tc.want, tc.got, p, tc.problem)
		}
	}
}

func TestCheckLogAndStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), checkLogFileName)
	if ev, err := readCheckLog(path); err != nil || ev != nil {
		t.Fatalf("missing log: %v %v", ev, err)
	}
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	for _, e := range []CheckEvent{
		{ID: "swap", At: now, Kind: checkEventRun, Built: true, Passed: 3, Total: 5},
		{ID: "swap", At: now.Add(time.Minute), Kind: checkEventHint, Hint: 1},
		{ID: "swap", At: now.Add(2 * time.Minute), Kind: checkEventRun, Built: true, Passed: 5, Total: 5, OK: true, HintsSeen: 1},
		{ID: "hello", At: now, Kind: checkEventRun, Built: true, Passed: 1, Total: 1, OK: true},
	} {
		if err := appendCheckEvent(path, e); err != nil {
			t.Fatal(err)
		}
	}
	// A torn last line (a crash mid-write) is skipped, not fatal.
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString(`{"id":"swap","kind":"ru`)
	f.Close()

	events, err := readCheckLog(path)
	if err != nil || len(events) != 4 {
		t.Fatalf("read %d events, err %v", len(events), err)
	}
	st := checkStatuses(events)
	if s := st["swap"]; s.Runs != 2 || s.Best != 5 || !s.Passed || s.OnOwn || s.Hints != 1 || !s.LastOK {
		t.Errorf("swap status = %+v (passed with a hint: not on its own)", s)
	}
	if s := st["hello"]; !s.Passed || !s.OnOwn {
		t.Errorf("hello status = %+v (passed with no hints)", s)
	}
}

func TestEnsureStarterNeverOverwrites(t *testing.T) {
	ch, _ := checkByID("fix_average")
	path := filepath.Join(t.TempDir(), "Stage 01", ch.File)
	if created, err := ensureStarter(path, ch); err != nil || !created {
		t.Fatalf("first call: created=%v err=%v", created, err)
	}
	os.WriteFile(path, []byte("my work"), 0o644)
	if created, err := ensureStarter(path, ch); err != nil || created {
		t.Fatalf("second call: created=%v err=%v", created, err)
	}
	if b, _ := os.ReadFile(path); string(b) != "my work" {
		t.Errorf("the learner's file was overwritten: %q", b)
	}
}

func TestLabsDir(t *testing.T) {
	docs := func() (string, error) { return "/home/ana/Documents", nil }
	env := map[string]string{}
	get := func(k string) string { return env[k] }
	if d := labsDirFor("/data/reg.json", get, docs); d != filepath.Join("/home/ana/Documents", labsFolderName) {
		t.Errorf("default: %s", d)
	}
	env["ACADEMY_HOME"] = "/srv/academy"
	if d := labsDirFor("/data/reg.json", get, docs); d != filepath.Join("/srv/academy", labsFolderName) {
		t.Errorf("ACADEMY_HOME: %s", d)
	}
	env["ACADEMY_LABS"] = "/tmp/labs"
	if d := labsDirFor("/data/reg.json", get, docs); d != "/tmp/labs" {
		t.Errorf("ACADEMY_LABS: %s", d)
	}
}

// referenceSolutions pass every case of their check.
var referenceSolutions = map[string]string{
	"hello": `#include <stdio.h>

int main(void) {
    printf("Hello, world!\n");
    return 0;
}
`,
	"fizzbuzz": `#include <stdio.h>

int main(void) {
    for (int i = 1; i <= 100; i++) {
        if (i % 15 == 0) {
            printf("FizzBuzz\n");
        } else if (i % 3 == 0) {
            printf("Fizz\n");
        } else if (i % 5 == 0) {
            printf("Buzz\n");
        } else {
            printf("%d\n", i);
        }
    }
    return 0;
}
`,
	"temperature": `double c_to_f(double c) { return c * 9.0 / 5.0 + 32.0; }
double f_to_c(double f) { return (f - 32.0) * 5.0 / 9.0; }
`,
	"swap": `void swap(int *a, int *b) {
    int t = *a;
    *a = *b;
    *b = t;
}
`,
	"letters": `#include <ctype.h>
#include <stdio.h>

int main(void) {
    long counts[26] = {0};
    int c;
    while ((c = getchar()) != EOF) {
        if (c < 128 && isalpha(c)) {
            counts[tolower(c) - 'a']++;
        }
    }
    for (int i = 0; i < 26; i++) {
        if (counts[i] > 0) {
            printf("%c %ld\n", 'a' + i, counts[i]);
        }
    }
    return 0;
}
`,
	"safe_number": `#include <ctype.h>
#include <errno.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>

int main(void) {
    char line[256];
    if (fgets(line, sizeof line, stdin) == NULL) {
        fprintf(stderr, "error: no input\n");
        return 1;
    }
    char *end;
    errno = 0;
    long v = strtol(line, &end, 10);
    if (end == line) {
        fprintf(stderr, "error: not a number\n");
        return 1;
    }
    while (*end != '\0' && isspace((unsigned char)*end)) {
        end++;
    }
    if (*end != '\0') {
        fprintf(stderr, "error: not a number\n");
        return 1;
    }
    if (errno == ERANGE || v < INT_MIN || v > INT_MAX) {
        fprintf(stderr, "error: out of range\n");
        return 1;
    }
    printf("ok %d\n", (int)v);
    return 0;
}
`,
	"fix_average": `#include <stdio.h>

double average(const int *a, int n) {
    int sum = 0;
    for (int i = 0; i < n; i++) {
        sum += a[i];
    }
    return (double)sum / n;
}

int main(void) {
    int scores[] = {90, 72, 85, 60, 99};
    printf("%.1f\n", average(scores, 5));
    int pair[] = {1, 2};
    printf("%.1f\n", average(pair, 2));
    return 0;
}
`,
}

// needCompiler skips runner tests where there is nothing to run them with.
func needCompiler(t *testing.T) {
	t.Helper()
	if !runnerSupported {
		t.Skip("the runner needs Linux, macOS or a BSD")
	}
	if _, err := findCompiler(os.Getenv); err != nil {
		t.Skip("no C compiler on this machine")
	}
}

// runSource writes src as the check's file and runs the check.
func runSource(t *testing.T, id, src string) CheckRun {
	t.Helper()
	ch, ok := checkByID(id)
	if !ok {
		t.Fatalf("no check %q", id)
	}
	path := filepath.Join(t.TempDir(), ch.File)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	run, err := runCheck(ch, path)
	if err != nil {
		t.Fatalf("%s: %v", id, err)
	}
	return run
}

func TestReferenceSolutionsPass(t *testing.T) {
	needCompiler(t)
	for _, ch := range checks {
		src, ok := referenceSolutions[ch.ID]
		if !ok {
			t.Errorf("%s has no reference solution in the tests", ch.ID)
			continue
		}
		run := runSource(t, ch.ID, src)
		if !run.OK() {
			t.Errorf("%s: reference solution failed (built=%v)\n%s", ch.ID, run.Built, run.BuildLog)
			for _, c := range run.Cases {
				if !c.Pass {
					t.Errorf("  %s: %s\n  stderr: %s", c.Case.Name, c.Problem, c.Stderr)
				}
			}
		}
		if run.Compiler == "" {
			t.Errorf("%s: compiler version not recorded", ch.ID)
		}
	}
}

// failsWith runs src and checks that the named case fails with a problem
// containing want.
func failsWith(t *testing.T, id, src, caseName, want string) CheckRun {
	t.Helper()
	run := runSource(t, id, src)
	if run.OK() {
		t.Fatalf("%s: a wrong program passed", id)
	}
	for _, c := range run.Cases {
		if c.Case.Name == caseName {
			if c.Pass || !strings.Contains(c.Problem, want) {
				t.Errorf("%s/%s: pass=%v problem=%q, want it to mention %q", id, caseName, c.Pass, c.Problem, want)
			}
			return run
		}
	}
	t.Fatalf("%s: no case %q ran (built=%v)\n%s", id, caseName, run.Built, run.BuildLog)
	return run
}

func TestWrongProgramsFailWithReasons(t *testing.T) {
	needCompiler(t)
	old := checkCaseTimeout
	checkCaseTimeout = time.Second
	defer func() { checkCaseTimeout = old }()

	// The XOR trick zeroes a variable swapped with itself.
	failsWith(t, "swap", "void swap(int *a, int *b) { *a ^= *b; *b ^= *a; *a ^= *b; }\n",
		"a variable with itself", `line 1 should be "4", but was "0"`)

	// Integer division: 9 / 5 is 1.
	failsWith(t, "temperature", "double c_to_f(double c) { return 9 / 5 * c + 32; }\ndouble f_to_c(double f) { return (f - 32) * 5.0 / 9.0; }\n",
		"c_to_f(100)", `"212.00", but was "132.00"`)

	// An endless loop is stopped by the time limit.
	failsWith(t, "hello", "int main(void) { for (;;) {} }\n", "prints the greeting", "did not finish")

	// A crash is reported as one (without sanitizers it is a signal).
	run := runSource(t, "hello", "#include <stdio.h>\nint main(void) { int *p = NULL; printf(\"%d\\n\", *p); return 0; }\n")
	if run.OK() || !run.Built || run.Cases[0].Pass {
		t.Fatal("a NULL dereference passed")
	}
	if p := run.Cases[0].Problem; !strings.Contains(p, "Sanitizer") && !strings.Contains(p, "crashed") {
		t.Errorf("NULL dereference: problem = %q", p)
	}

	// The unfixed starter fails: the sanitizer catches the read past the end.
	ch, _ := checkByID("fix_average")
	run = runSource(t, "fix_average", ch.Starter)
	if run.OK() {
		t.Fatal("the buggy starter passed")
	}
	if run.Sanitizers && !strings.Contains(run.Cases[0].Problem, "AddressSanitizer") {
		t.Errorf("starter with sanitizers: problem = %q, want an AddressSanitizer report", run.Cases[0].Problem)
	}

	// An error printed to stdout instead of stderr is caught.
	failsWith(t, "safe_number", "#include <stdio.h>\nint main(void) { printf(\"error\\n\"); return 1; }\n",
		"letters", "nothing was expected on standard output")

	// A warning is an error: the build fails and says why.
	run = runSource(t, "hello", "#include <stdio.h>\nint main(void) { int unused; printf(\"Hello, world!\\n\"); return 0; }\n")
	if run.Built || !strings.Contains(run.BuildLog, "unused") {
		t.Errorf("a warning should fail the build: built=%v log=%q", run.Built, run.BuildLog)
	}

	// A main in a harness check is explained.
	run = runSource(t, "swap", "void swap(int *a, int *b) { int t = *a; *a = *b; *b = t; }\nint main(void) { return 0; }\n")
	if run.Built || !strings.Contains(run.BuildLog, "must not define main") {
		t.Errorf("main in a harness check: built=%v log=%q", run.Built, run.BuildLog)
	}
}

// A program that starts a background child cannot outlive the check.
func TestRunnerKillsTheProcessGroup(t *testing.T) {
	needCompiler(t)
	old := checkCaseTimeout
	checkCaseTimeout = time.Second
	defer func() { checkCaseTimeout = old }()
	marker := filepath.Join(t.TempDir(), "survived")
	src := `#include <stdio.h>
#include <unistd.h>
int main(void) {
    if (fork() == 0) {
        sleep(2);
        FILE *f = fopen("` + marker + `", "w");
        if (f) fclose(f);
        return 0;
    }
    for (;;) {}
}
`
	failsWith(t, "hello", src, "prints the greeting", "did not finish")
	time.Sleep(2500 * time.Millisecond)
	if _, err := os.Stat(marker); err == nil {
		t.Error("the forked child outlived the check")
	}
}

// The program sees none of the learner's environment variables.
func TestRunnerEnvironmentIsMinimal(t *testing.T) {
	needCompiler(t)
	t.Setenv("ACADEMY_TEST_SECRET", "hunter2")
	run := runSource(t, "hello", `#include <stdio.h>
#include <stdlib.h>
int main(void) {
    printf("%s\n", getenv("ACADEMY_TEST_SECRET") ? "leaked" : "Hello, world!");
    return 0;
}
`)
	if !run.OK() {
		t.Errorf("the environment leaked into the program: %+v", run.Cases)
	}
}

func TestSanitizerReport(t *testing.T) {
	asan := `=================================================================
==2979==ERROR: AddressSanitizer: stack-buffer-overflow on address 0x7f3b7aa00054 at pc 0x5596975613b5 bp 0x7ffc703e9550 sp 0x7ffc703e9540
READ of size 4 at 0x7f3b7aa00054 thread T0
    #0 0x5596975613b4 in average /home/ana/Academy Labs/Stage 01/fix_average.c:11
    #1 0x55969756164b in main /home/ana/Academy Labs/Stage 01/fix_average.c:18
`
	if got, want := sanitizerReport(asan), "AddressSanitizer found a stack-buffer-overflow, READ of size 4, in average, fix_average.c:11"; got != want {
		t.Errorf("ASan: got %q, want %q", got, want)
	}
	ubsan := "/home/ana/Academy Labs/Stage 01/hello.c:2:33: runtime error: load of null pointer of type 'int'\n    #0 0x5653919c6244 in main\n"
	if got, want := sanitizerReport(ubsan), "UndefinedBehaviorSanitizer: load of null pointer of type 'int', at hello.c:2:33"; got != want {
		t.Errorf("UBSan: got %q, want %q", got, want)
	}
	if got := sanitizerReport("warning: something else\n"); got != "" {
		t.Errorf("no report: got %q", got)
	}
}

// The Study Hall screens: open a check, reveal a hint, run it, go back.
func TestCheckedExercisesScreen(t *testing.T) {
	needCompiler(t)
	dir := t.TempDir()
	t.Setenv("ACADEMY_LABS", filepath.Join(dir, "labs"))
	input := strings.Join([]string{
		"1", // hello
		"2", // reveal hint 1
		"1", // run: the starter has no main, so it does not build
		"3", // back
		"8", // back to the Study Hall
	}, "\n") + "\n"
	var out strings.Builder
	a := &App{reg: seedRegistry(), path: filepath.Join(dir, "reg.json"), width: 80,
		con: &console{in: bufio.NewReader(strings.NewReader(input)), out: &out}}
	if err := a.checkedExercises(1); err != nil {
		t.Fatalf("checked exercises: %v\n%s", err, out.String())
	}
	text := stripANSI(out.String())
	for _, want := range []string{"Created", "hello.c", "Hint 1", "It did not compile", "academy -check hello"} {
		if !strings.Contains(text, want) {
			t.Errorf("screen lacks %q:\n%s", want, text)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "labs", "Stage 01", "hello.c")); err != nil {
		t.Errorf("starter not created: %v", err)
	}
	events, _ := readCheckLog(checkLogPath(a.path))
	if len(events) != 2 || events[0].Kind != checkEventHint || events[1].Kind != checkEventRun || events[1].HintsSeen != 1 || events[1].Built {
		t.Errorf("log = %+v, want a hint then a failed run that saw 1 hint", events)
	}
}

func TestCheckCommand(t *testing.T) {
	needCompiler(t)
	dir := t.TempDir()
	t.Setenv("ACADEMY_LABS", filepath.Join(dir, "labs"))
	reg := filepath.Join(dir, "reg.json")
	var out strings.Builder
	if code := checkCommand(reg, "nope", &out, 80); code != 2 {
		t.Errorf("unknown ID: exit %d", code)
	}
	if code := checkCommand(reg, "swap", &out, 80); code != 1 || !strings.Contains(out.String(), "Created") {
		t.Errorf("first run should create the file and stop: exit %d\n%s", code, out.String())
	}
	file := filepath.Join(dir, "labs", "Stage 01", "swap.c")
	os.WriteFile(file, []byte(referenceSolutions["swap"]), 0o644)
	out.Reset()
	if code := checkCommand(reg, "swap", &out, 80); code != 0 || !strings.Contains(out.String(), "All 5 cases pass") {
		t.Errorf("solution: exit %d\n%s", code, out.String())
	}
	out.Reset()
	if code := checkCommand(reg, "list", &out, 80); code != 0 || !strings.Contains(out.String(), "passed on your own") {
		t.Errorf("list: exit %d\n%s", code, out.String())
	}
}

// Where the sanitizers are missing, checks still build and run without them.
func TestRunnerWithoutSanitizers(t *testing.T) {
	needCompiler(t)
	real, _ := findCompiler(os.Getenv)
	if len(real.pre) > 0 {
		t.Skip("CC is a wrapper command")
	}
	fake := filepath.Join(t.TempDir(), "nosan-cc")
	script := "#!/bin/sh\nfor a in \"$@\"; do case \"$a\" in -fsanitize=*) echo 'cc: error: unsupported option' >&2; exit 1;; esac; done\nexec " + real.path + " \"$@\"\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CC", fake)
	run := runSource(t, "swap", referenceSolutions["swap"])
	if run.Sanitizers || !run.OK() {
		t.Errorf("without sanitizers: sanitizers=%v ok=%v\n%s", run.Sanitizers, run.OK(), run.BuildLog)
	}
	var out strings.Builder
	writeCheckRun(&out, run, 80)
	if !strings.Contains(out.String(), "no memory checks") {
		t.Errorf("the report should say the memory checks were off:\n%s", out.String())
	}
}

func TestFindCompilerHonoursCC(t *testing.T) {
	needCompiler(t)
	env := map[string]string{"CC": "env gcc"}
	cc, err := findCompiler(func(k string) string { return env[k] })
	if err != nil || filepath.Base(cc.path) != "env" || len(cc.pre) != 1 || cc.pre[0] != "gcc" {
		t.Errorf("CC=%q: got %+v, %v", env["CC"], cc, err)
	}
	env["CC"] = "no-such-compiler-here"
	if cc, err := findCompiler(func(k string) string { return env[k] }); err != nil || len(cc.pre) != 0 {
		t.Errorf("a CC that is not installed should fall back to cc: %+v, %v", cc, err)
	}
}

func TestWrapPath(t *testing.T) {
	p := "/tmp/academy-labs-ootp9n8d/Stage 01/fix_average.c"
	for _, w := range []int{10, 20, 34, 80} {
		lines := wrapPath(p, w)
		if strings.Join(lines, "") != p {
			t.Errorf("width %d: %q does not rejoin to the path", w, lines)
		}
		for _, l := range lines {
			if visibleLen(l) > w {
				t.Errorf("width %d: line %q is too wide", w, l)
			}
		}
	}
	if got := wrapPath(p, 34); len(got) != 2 || got[0] != "/tmp/academy-labs-ootp9n8d/" {
		t.Errorf("should break after a separator: %q", got)
	}
}
