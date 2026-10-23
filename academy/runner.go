package main

// The checked-exercise runner: compile the learner's C file with the C
// compiler on their machine, run it against a check's cases, and say in
// words what went wrong.
//
// What it does and does not protect against. The program is the
// learner's own code, run on their own computer, as if they had typed the
// commands themselves; the academy never runs code from anywhere else.
// Each run gets a fresh temporary folder as its working directory, a
// minimal environment (none of the learner's variables, so no tokens or
// keys), no standard input beyond the case's, a time limit, and an output
// cap; on Linux, macOS and the BSDs the whole process group is killed at
// the limit, so a program that forks cannot outlive the check. It is not
// a sandbox: the program could still read or write the learner's files or
// use the network, exactly like any program they compile themselves.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	checkCaseTimeout = 2 * time.Second // per case; tests shorten it
	compileTimeout   = 60 * time.Second
)

const (
	maxCheckStdout = 1 << 20  // 1 MB of output is far more than any case needs
	maxCheckStderr = 64 << 10 // compiler and sanitizer messages
)

var (
	errNoCompiler = errors.New("no C compiler found: install one (Linux: sudo apt install build-essential, or your distribution's equivalent; macOS: xcode-select --install), or set CC")
	errNoRunner   = errors.New("checked exercises run on Linux, macOS and the BSDs; on Windows, install WSL (the Windows Subsystem for Linux) and run the academy there")
)

// CaseResult is the outcome of one case.
type CaseResult struct {
	Case    CheckCase
	Pass    bool
	Stdout  string
	Stderr  string
	Exit    int
	Problem string // why it failed, in words ("" when it passed)
}

// CheckRun is the outcome of running one check.
type CheckRun struct {
	Check      Check
	Compiler   string // the first line of "cc --version"
	Sanitizers bool   // built with AddressSanitizer and UndefinedBehaviorSanitizer
	Leaks      bool   // the leak checker was on (for checks about memory, where it works)
	Built      bool
	BuildLog   string // compiler messages: warnings count as errors
	Cases      []CaseResult
	Passed     int
}

// OK reports whether the program built and passed every case.
func (r CheckRun) OK() bool { return r.Built && r.Passed == len(r.Check.Cases) }

// compiler is the C compiler to run: a program and the words that come
// before the compiler's own arguments ($CC may be "ccache gcc").
type compiler struct {
	path string
	pre  []string
}

func (c compiler) command(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, c.path, append(append([]string{}, c.pre...), args...)...)
}

func (c compiler) key() string { return strings.Join(append([]string{c.path}, c.pre...), " ") }

// findCompiler returns the C compiler to use: $CC if it names one on the
// PATH, otherwise cc, gcc or clang.
func findCompiler(getenv func(string) string) (compiler, error) {
	if f := strings.Fields(getenv("CC")); len(f) > 0 {
		if p, err := exec.LookPath(f[0]); err == nil {
			return compiler{p, f[1:]}, nil
		}
	}
	for _, n := range []string{"cc", "gcc", "clang"} {
		if p, err := exec.LookPath(n); err == nil {
			return compiler{path: p}, nil
		}
	}
	return compiler{}, errNoCompiler
}

// compilerVersion is the first line of "cc --version", for the record.
func compilerVersion(cc compiler) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := cc.command(ctx, "--version").Output()
	if err != nil {
		return filepath.Base(cc.path)
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(line)
}

var sanitizerProbe sync.Map // compiler key → bool

// sanitizersWork reports whether cc can build and run a program with the
// address and undefined-behaviour sanitizers here. Some systems (minimal
// containers, some distributions) lack their runtime libraries.
func sanitizersWork(cc compiler) bool {
	if v, ok := sanitizerProbe.Load(cc.key()); ok {
		return v.(bool)
	}
	ok := func() bool {
		dir, err := os.MkdirTemp("", "academy-probe-")
		if err != nil {
			return false
		}
		defer os.RemoveAll(dir)
		src := filepath.Join(dir, "probe.c")
		if os.WriteFile(src, []byte("int main(void) { return 0; }\n"), 0o600) != nil {
			return false
		}
		ctx, cancel := context.WithTimeout(context.Background(), compileTimeout)
		defer cancel()
		prog := filepath.Join(dir, "probe")
		args := append(sanitizerFlags(), "-o", prog, src)
		if cc.command(ctx, args...).Run() != nil {
			return false
		}
		cmd := exec.CommandContext(ctx, prog)
		cmd.Env = childEnv(dir, false)
		return cmd.Run() == nil
	}()
	sanitizerProbe.Store(cc.key(), ok)
	return ok
}

var leakProbe sync.Map // compiler key → bool

// leaksWork reports whether the sanitizers' leak checker works here. It
// needs to stop the program's threads with ptrace at exit, which some
// containers and macOS do not allow.
func leaksWork(cc compiler) bool {
	if v, ok := leakProbe.Load(cc.key()); ok {
		return v.(bool)
	}
	ok := func() bool {
		dir, err := os.MkdirTemp("", "academy-probe-")
		if err != nil {
			return false
		}
		defer os.RemoveAll(dir)
		src := filepath.Join(dir, "leak.c")
		leaky := "#include <stdlib.h>\nint main(void) { void *volatile p = malloc(64); (void)p; p = 0; return 0; }\n"
		if os.WriteFile(src, []byte(leaky), 0o600) != nil {
			return false
		}
		ctx, cancel := context.WithTimeout(context.Background(), compileTimeout)
		defer cancel()
		prog := filepath.Join(dir, "leak")
		args := append(sanitizerFlags(), "-o", prog, src)
		if cc.command(ctx, args...).Run() != nil {
			return false
		}
		cmd := exec.CommandContext(ctx, prog)
		cmd.Env = childEnv(dir, true)
		out, err := cmd.CombinedOutput()
		return err != nil && strings.Contains(string(out), "LeakSanitizer: detected memory leaks")
	}()
	leakProbe.Store(cc.key(), ok)
	return ok
}

func sanitizerFlags() []string {
	return []string{"-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fno-omit-frame-pointer"}
}

// childEnv is the whole environment a learner's program gets. Leak
// checking is on only for checks about memory, and only where it works
// (leaksWork).
func childEnv(dir string, leaks bool) []string {
	asan := "ASAN_OPTIONS=detect_leaks=0"
	if leaks {
		asan = "ASAN_OPTIONS=detect_leaks=1"
	}
	return []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + dir,
		"TMPDIR=" + dir,
		"LANG=C",
		"LC_ALL=C",
		asan,
		"UBSAN_OPTIONS=print_stacktrace=1",
	}
}

// cappedBuffer keeps at most max bytes and notes whether more came.
type cappedBuffer struct {
	bytes.Buffer
	max  int
	over bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if room := b.max - b.Len(); len(p) > room {
		b.over = true
		if room > 0 {
			b.Buffer.Write(p[:room])
		}
		return len(p), nil // keep the program running until the limit stops it
	}
	return b.Buffer.Write(p)
}

// runCheck compiles src (and the check's other files, which sit beside
// it) and runs it against the check's cases.
func runCheck(ch Check, src string) (CheckRun, error) {
	run := CheckRun{Check: ch}
	if !runnerSupported {
		return run, errNoRunner
	}
	cc, err := findCompiler(os.Getenv)
	if err != nil {
		return run, err
	}
	srcDir := filepath.Dir(src)
	sources := []string{src}
	for _, f := range ch.Extra {
		if strings.HasSuffix(f.Name, ".c") {
			sources = append(sources, filepath.Join(srcDir, f.Name))
		}
	}
	for _, f := range append([]string{src}, extraPaths(srcDir, ch)...) {
		if _, err := os.Stat(f); err != nil {
			return run, fmt.Errorf("cannot read your file %s: %w", f, err)
		}
	}
	run.Compiler = compilerVersion(cc)
	run.Sanitizers = sanitizersWork(cc)
	run.Leaks = ch.Leaks && run.Sanitizers && leaksWork(cc)

	dir, err := os.MkdirTemp("", "academy-check-")
	if err != nil {
		return run, err
	}
	defer os.RemoveAll(dir)
	prog := filepath.Join(dir, strings.TrimSuffix(ch.File, ".c"))
	args := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-g", "-O0"}
	if run.Sanitizers {
		args = append(args, sanitizerFlags()...)
	}
	args = append(args, ch.Flags...)
	args = append(args, "-I", srcDir, "-o", prog)
	args = append(args, sources...)
	if ch.Harness != "" {
		h := filepath.Join(dir, "check_main.c")
		if err := os.WriteFile(h, []byte(ch.Harness), 0o600); err != nil {
			return run, err
		}
		args = append(args, h)
	}
	// Libraries go after the sources: the linker only takes from a library
	// the symbols that files before it still need.
	args = append(args, ch.Libs...)

	ctx, cancel := context.WithTimeout(context.Background(), compileTimeout)
	defer cancel()
	build := cc.command(ctx, args...)
	build.Dir = srcDir
	log := &cappedBuffer{max: maxCheckStderr}
	build.Stdout, build.Stderr = log, log
	err = build.Run()
	run.BuildLog = strings.TrimSpace(log.String())
	if err != nil {
		if ch.Harness != "" && (strings.Contains(run.BuildLog, "multiple definition of `main'") || strings.Contains(run.BuildLog, "duplicate symbol") && strings.Contains(run.BuildLog, "_main")) {
			run.BuildLog += "\n\nYour file must not define main: the checker supplies its own and calls your functions."
		}
		for _, h := range ch.BuildHints {
			if strings.Contains(run.BuildLog, h.Match) {
				run.BuildLog += "\n\n" + h.Note
			}
		}
		if ctx.Err() != nil {
			run.BuildLog += "\n\nThe compiler did not finish within a minute."
		}
		return run, nil
	}
	run.Built = true

	timeout := max(checkCaseTimeout, ch.Timeout)
	for i, c := range ch.Cases {
		// Each case runs in a folder of its own holding only its files.
		caseDir := filepath.Join(dir, fmt.Sprintf("case-%d", i+1))
		if err := os.Mkdir(caseDir, 0o700); err != nil {
			return run, err
		}
		for name, content := range c.Files {
			f := filepath.Join(caseDir, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(f), 0o700); err != nil {
				return run, err
			}
			if err := os.WriteFile(f, []byte(content), 0o600); err != nil {
				return run, err
			}
		}
		r := runCase(prog, caseDir, c, timeout, run.Leaks)
		if r.Pass {
			run.Passed++
		}
		run.Cases = append(run.Cases, r)
	}
	return run, nil
}

// extraPaths are a check's other files, beside the main one in dir.
func extraPaths(dir string, ch Check) []string {
	var out []string
	for _, f := range ch.Extra {
		out = append(out, filepath.Join(dir, f.Name))
	}
	return out
}

// runCase runs the program once and judges the result.
func runCase(prog, dir string, c CheckCase, timeout time.Duration, leaks bool) CaseResult {
	c = c.resolved()
	res := CaseResult{Case: c}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, prog, c.Args...)
	cmd.Dir = dir
	cmd.Env = childEnv(dir, leaks)
	cmd.Stdin = strings.NewReader(c.Stdin)
	stdout := &cappedBuffer{max: maxCheckStdout}
	stderr := &cappedBuffer{max: maxCheckStderr}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	cmd.WaitDelay = time.Second
	prepareProcess(cmd)
	err := cmd.Run()
	cleanupProcess(cmd)
	res.Stdout, res.Stderr = stdout.String(), stderr.String()
	if cmd.ProcessState != nil {
		res.Exit = cmd.ProcessState.ExitCode()
	}

	var exitErr *exec.ExitError
	switch {
	case sanitizerReport(res.Stderr) != "": // even if the report ran past the time limit
		res.Problem = sanitizerReport(res.Stderr)
	case ctx.Err() != nil:
		res.Problem = fmt.Sprintf("did not finish within %s: an endless loop, or waiting for input that never comes", timeout)
	case stdout.over:
		res.Problem = "printed more than 1 MB: probably an endless loop"
	case err != nil && !errors.As(err, &exitErr):
		res.Problem = "could not run the program: " + err.Error()
	case res.Exit < 0:
		res.Problem = "crashed (" + cmd.ProcessState.String() + ")"
	case res.Exit != c.WantExit:
		res.Problem = fmt.Sprintf("main returned %d, expected %d", res.Exit, c.WantExit)
	default:
		res.Problem = outputDiff(c.Want, res.Stdout)
		if res.Problem == "" && c.Stderr && strings.TrimSpace(res.Stderr) == "" {
			res.Problem = "printed nothing to standard error: an error message belongs there"
		}
		if res.Problem == "" {
			res.Problem = filesDiff(dir, c.WantFiles)
		}
	}
	res.Pass = res.Problem == ""
	return res
}

// filesDiff checks the files a case must leave behind, in name order, and
// describes the first difference ("" when they all match).
func filesDiff(dir string, want map[string]string) string {
	names := make([]string, 0, len(want))
	for name := range want {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil {
			return fmt.Sprintf("the file %s should exist afterwards, but does not", name)
		}
		if string(got) != want[name] {
			if d := outputDiff(want[name], string(got)); d != "" {
				return "in the file " + name + ": " + d
			}
			return fmt.Sprintf("the file %s should hold %d bytes, but holds %d", name, len(want[name]), len(got))
		}
	}
	return ""
}

// sanitizerReport sums up an AddressSanitizer or UndefinedBehaviorSanitizer
// report in one line: what went wrong and where ("AddressSanitizer found a
// stack-buffer-overflow, READ of size 4, in average, fix_average.c:11"),
// or "" when there is none.
func sanitizerReport(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if _, after, ok := strings.Cut(l, "ERROR: AddressSanitizer: "); ok {
			kind, _, _ := strings.Cut(after, " on ") // drop the raw addresses
			parts := []string{"AddressSanitizer found a " + kind}
			for _, next := range lines[i+1:] {
				next = strings.TrimSpace(next)
				switch {
				case strings.HasPrefix(next, "READ of size") || strings.HasPrefix(next, "WRITE of size"):
					what, _, _ := strings.Cut(next, " at ")
					parts = append(parts, what)
				case strings.HasPrefix(next, "#"):
					if where := frameWhere(next); where != "" {
						parts = append(parts, where)
					}
					return strings.Join(parts, ", ")
				}
			}
			return strings.Join(parts, ", ")
		}
		if strings.Contains(l, "ERROR: LeakSanitizer: detected memory leaks") {
			return leakSummary(lines[i+1:])
		}
		if loc, msg, ok := strings.Cut(l, ": runtime error: "); ok {
			return "UndefinedBehaviorSanitizer: " + msg + ", at " + filepath.Base(loc)
		}
	}
	return ""
}

// leakSummary sums up a LeakSanitizer report: how much leaked, and the
// learner's line that allocated the first block.
func leakSummary(lines []string) string {
	out := "LeakSanitizer found memory that was never freed"
	where := ""
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if after, ok := strings.CutPrefix(l, "SUMMARY: AddressSanitizer: "); ok {
			out += ": " + strings.TrimSuffix(after, ".")
			break
		}
		if where != "" || !strings.HasPrefix(l, "#") || strings.Contains(l, "sanitizer") || strings.Contains(l, "asan") {
			continue // the allocator's own frames come first
		}
		if _, after, ok := strings.Cut(l, " in "); ok {
			if fn, _, _ := strings.Cut(after, " "); !allocator[fn] {
				where = frameWhere(l)
			}
		}
	}
	if where != "" {
		out += ", allocated " + where
	}
	return out
}

// allocator names the C library's allocation functions, skipped when
// finding the learner's own line in a leak report.
var allocator = map[string]bool{"malloc": true, "calloc": true, "realloc": true, "strdup": true, "strndup": true, "reallocarray": true}

// frameWhere turns a stack frame ("#0 0x55… in average /path/f.c:11") into
// "in average, f.c:11".
func frameWhere(frame string) string {
	_, after, ok := strings.Cut(frame, " in ")
	if !ok {
		return ""
	}
	fn, loc, ok := strings.Cut(after, " ")
	if !ok {
		return "in " + fn
	}
	return fmt.Sprintf("in %s, %s", fn, filepath.Base(strings.TrimSpace(loc)))
}

// normOutput ignores what nobody can see: carriage returns, spaces at the
// ends of lines and blank lines at the end.
func normOutput(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// outputDiff describes the first difference between the expected and the
// actual output, or returns "" when they match.
func outputDiff(want, got string) string {
	w, g := normOutput(want), normOutput(got)
	for i := 0; i < len(w) || i < len(g); i++ {
		switch {
		case i >= len(g):
			return fmt.Sprintf("output stopped early: line %d should be %q", i+1, w[i])
		case i >= len(w):
			if len(w) == 0 {
				return fmt.Sprintf("printed %q, but nothing was expected on standard output", g[i])
			}
			return fmt.Sprintf("extra output from line %d: %q", i+1, g[i])
		case w[i] != g[i]:
			return fmt.Sprintf("line %d should be %q, but was %q", i+1, w[i], g[i])
		}
	}
	return ""
}
