package main

// Screens for checked exercises: the list in the Study Hall, one
// exercise with its file, hints and runs, and "academy -check" for the
// terminal next to the learner's editor.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// homeTilde shortens a path under the home folder to ~/…, for display.
func homeTilde(p string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.Join("~", rel)
		}
	}
	return p
}

// wrapPath breaks a path into lines of at most width columns, after a
// separator where it can, so a long path never runs off a narrow screen.
func wrapPath(p string, width int) []string {
	width = max(width, 8)
	var lines []string
	for visibleLen(p) > width {
		// Break after a separator (either kind: paths may come from any
		// system); with none in reach, cut anywhere.
		cut := strings.LastIndexAny(p[:min(len(p), width)], `/\`) + 1
		if cut <= 1 {
			cut = len(truncateBytes(p, width))
		}
		lines = append(lines, p[:cut])
		p = p[cut:]
	}
	return append(lines, p)
}

// truncateBytes is the longest prefix of s that is at most width columns.
func truncateBytes(s string, width int) string {
	n := 0
	for i, r := range s {
		if n+visibleLen(string(r)) > width {
			return s[:i]
		}
		n += visibleLen(string(r))
	}
	return s
}

// checkSubject is what a check gives evidence for: its concept, or the
// stage as a whole for a project.
func checkSubject(ch Check) string {
	if ch.Project {
		return fmt.Sprintf("the Stage %d project", ch.Stage)
	}
	return "“" + ch.Concept + "”"
}

// checkMark is a check's state as a mark and a few words.
func checkMark(s CheckStatus) (string, string) {
	switch {
	case s.OnOwn:
		return sty.Green("✓"), "passed on your own"
	case s.Passed:
		return sty.Green("✓"), "passed, with hints"
	case s.Runs > 0 && !s.LastBuilt && s.LastCount == 0:
		return sty.Yellow("◐"), "does not compile yet"
	case s.Runs > 0:
		return sty.Yellow("◐"), fmt.Sprintf("%d/%d cases passing", s.LastCount, s.Total)
	}
	return sty.Gray("○"), "not run yet"
}

// loadCheckStatuses reads the check log, or returns empty statuses. It
// also keeps them on the registry, for What's next.
func (a *App) loadCheckStatuses() map[string]CheckStatus {
	events, err := readCheckLog(checkLogPath(a.path))
	if err != nil {
		a.con.warn("Could not read the check log: %v", err)
	}
	st := checkStatuses(events)
	a.mu.Lock()
	a.reg.checkStatus = st
	a.mu.Unlock()
	return st
}

// checkedExercises lists a stage's checked exercises.
func (a *App) checkedExercises(stageID int) error {
	list := checksFor(stageID)
	for {
		st := a.loadCheckStatuses()
		a.println(heading("CHECKED EXERCISES · your code, the real compiler", sty.Green, a.cols()))
		for _, l := range wrap("Write each program in your own editor. The academy compiles it with the C compiler on your computer, with warnings as errors and memory checks on where available, then runs it against test cases. A pass is recorded as evidence.", a.cols()-4, "  ") {
			a.println(l)
		}
		a.println("")
		options := make([]string, 0, len(list)+1)
		for _, ch := range list {
			mark, words := checkMark(st[ch.ID])
			options = append(options, fmt.Sprintf("%s %s %s", mark, sty.Bold(ch.Title), sty.Gray("· "+words)))
		}
		options = append(options, "Back")
		n, err := a.con.promptChoice("Exercise", options)
		if errors.Is(err, errCancel) || (err == nil && n == len(list)) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := a.checkScreen(list[n]); err != nil && !errors.Is(err, errCancel) {
			return err
		}
	}
}

// checkScreen shows one checked exercise and runs it on request.
func (a *App) checkScreen(ch Check) error {
	file := checkFile(a.path, ch)
	created, err := ensureStarters(a.path, ch)
	if err != nil {
		a.con.fail("Could not create the files: %v", err)
		return nil
	}
	for _, f := range created {
		a.con.ok("Created %s for you.", filepath.Base(f))
	}
	for {
		st := a.loadCheckStatuses()[ch.ID]
		w := a.cols()
		a.println("")
		a.printf("  %s\n", sty.Bold(sty.Green(truncate(ch.Title, w-4))))
		for _, l := range wrap(fmt.Sprintf("Stage %d · %s", ch.Stage, checkSubject(ch)), w-4, "  ") {
			a.println(sty.Gray(l))
		}
		for _, l := range wrap(ch.Task, w-4, "  ") {
			a.println(l)
		}
		a.println("")
		files := append([]string{file}, extraPaths(filepath.Dir(file), ch)...)
		label := "Your file:"
		if len(files) > 1 {
			label = "Your files:"
		}
		a.printf("  %s\n", sty.Bold(label))
		for _, l := range wrapPath(homeTilde(filepath.Dir(file))+string(filepath.Separator), w-6) {
			a.printf("    %s\n", l)
		}
		for _, f := range files {
			a.printf("      %s\n", truncate(filepath.Base(f), w-8))
		}
		for _, l := range wrap(fmt.Sprintf("Edit it in any editor and save. Then run the check here, or in another terminal: academy -check %s", ch.ID), w-4, "  ") {
			a.println(sty.Gray(l))
		}
		mark, words := checkMark(st)
		a.printf("  %s %s %s\n", sty.Bold("Status:"), mark, words)
		for i := 0; i < st.Hints && i < len(ch.Hints); i++ {
			for j, l := range wrap(ch.Hints[i], w-12, "") {
				lead := "         "
				if j == 0 {
					lead = sty.Yellow(fmt.Sprintf("  Hint %d ", i+1))
				}
				a.printf("  %s%s\n", lead, l)
			}
		}
		a.println("")

		options := []string{sty.Bold("Run the check")}
		if st.Hints < len(ch.Hints) {
			options = append(options, fmt.Sprintf("Reveal hint %d of %d %s", st.Hints+1, len(ch.Hints), sty.Gray("(a pass after a hint is recorded as \"with hints\")")))
		}
		options = append(options, "Back")
		n, err := a.con.promptChoice("Checked exercise", options)
		if err != nil {
			return err
		}
		switch {
		case n == 0:
			a.runCheckNow(ch, file, st.Hints)
		case n == 1 && st.Hints < len(ch.Hints):
			if err := appendCheckEvent(checkLogPath(a.path), CheckEvent{ID: ch.ID, At: time.Now(), Kind: checkEventHint, Hint: st.Hints + 1}); err != nil {
				a.con.fail("Could not record the hint: %v", err)
			}
		default:
			return nil
		}
	}
}

// runCheckNow runs a check, shows the report and records the run.
func (a *App) runCheckNow(ch Check, file string, hints int) {
	a.con.note("Compiling and running %s…", ch.File)
	run, err := runCheck(ch, file)
	if err != nil {
		a.con.fail("%v", err)
		return
	}
	a.paged(false, func() { writeCheckRun(a.con.out, run, a.cols()) })
	if err := appendCheckEvent(checkLogPath(a.path), runEvent(run, hints, time.Now())); err != nil {
		a.con.warn("The result could not be recorded: %v", err)
	}
}

// writeCheckRun prints a run's report: how it was built, then each case.
func writeCheckRun(w io.Writer, run CheckRun, width int) {
	ch := run.Check
	checks := "memory and undefined-behaviour checks (AddressSanitizer, UBSan)"
	switch {
	case !run.Sanitizers:
		checks = "no memory checks (the sanitizers are not available with this compiler)"
	case run.Leaks:
		checks += " and the leak checker"
	case ch.Leaks:
		checks += "; the leak checker does not work here, so leaks are not caught"
	}
	fmt.Fprintln(w)
	for _, l := range wrap(fmt.Sprintf("Built with %s: -std=c17 -Wall -Wextra -Werror, %s.", run.Compiler, checks), width-4, "  ") {
		fmt.Fprintln(w, sty.Gray(l))
	}
	if !run.Built {
		fmt.Fprintf(w, "\n  %s %s\n\n", sty.Red("✗"), sty.Bold("It did not compile. Warnings count as errors here."))
		lines := strings.Split(run.BuildLog, "\n")
		for i, l := range lines {
			if i == 40 {
				fmt.Fprintf(w, "    %s\n", sty.Gray(fmt.Sprintf("… %d more lines", len(lines)-40)))
				break
			}
			// Wrapped, never cut: the end of a compiler message is
			// often the part that explains it.
			for _, part := range wrap(l, width-6, "") {
				fmt.Fprintf(w, "    %s\n", part)
			}
		}
		fmt.Fprintln(w)
		for _, l := range wrap("Read the first error first: it gives the file, line and column. Later errors are often caused by it.", width-4, "  ") {
			fmt.Fprintln(w, l)
		}
		return
	}
	fmt.Fprintln(w)
	for _, r := range run.Cases {
		if r.Pass {
			fmt.Fprintf(w, "  %s %s\n", sty.Green("✓"), r.Case.Name)
			continue
		}
		fmt.Fprintf(w, "  %s %s\n", sty.Red("✗"), sty.Bold(r.Case.Name))
		if in := caseInput(r.Case); in != "" {
			fmt.Fprintf(w, "      %s %s\n", sty.Gray("input:"), truncate(in, width-14))
		}
		for _, l := range wrap(r.Problem, width-8, "      ") {
			fmt.Fprintln(w, l)
		}
		if r.Case.Explain != "" {
			for _, l := range wrap(r.Case.Explain, width-8, "      ") {
				fmt.Fprintln(w, sty.Yellow(l))
			}
		}
	}
	fmt.Fprintln(w)
	if run.OK() {
		all := fmt.Sprintf("All %d cases pass.", len(ch.Cases))
		if len(ch.Cases) == 1 {
			all = "It passes."
		}
		fmt.Fprintf(w, "  %s %s\n", sty.Green("✓"), sty.Bold(all+" Recorded as evidence for "+checkSubject(ch)+"."))
		return
	}
	fmt.Fprintf(w, "  %s\n", sty.Bold(fmt.Sprintf("%d of %d cases pass.", run.Passed, len(ch.Cases))))
}

// caseInput describes what a case gives the program, for the report.
func caseInput(c CheckCase) string {
	c = c.resolved()
	var parts []string
	if len(c.Args) > 0 {
		parts = append(parts, "arguments "+strings.Join(c.Args, " "))
	}
	switch {
	case c.Stdin == "" && c.Want == "" && c.WantExit != 0:
		parts = append(parts, "no input at all")
	case c.Stdin != "":
		in := c.Stdin
		if len(in) > 60 {
			in = fmt.Sprintf("%s… (%d bytes)", in[:40], len(c.Stdin))
		}
		parts = append(parts, fmt.Sprintf("%q", in))
	}
	return strings.Join(parts, ", ")
}

// checkCommand is "academy -check": list the checks, or run one and
// record the result. It returns the exit code: 0 when the check passes.
func checkCommand(regPath, id string, out io.Writer, width int) int {
	logPath := checkLogPath(regPath)
	events, err := readCheckLog(logPath)
	if err != nil {
		fmt.Fprintf(out, "✗ could not read %s: %v\n", logPath, err)
		return 1
	}
	st := checkStatuses(events)
	if id == "list" {
		fmt.Fprintf(out, "Checked exercises (files in %s):\n\n", homeTilde(labsDir(regPath)))
		for _, ch := range checks {
			mark, words := checkMark(st[ch.ID])
			fmt.Fprintf(out, "  %s %-12s Stage %d  %s %s\n", mark, ch.ID, ch.Stage, ch.Title, sty.Gray("· "+words))
		}
		fmt.Fprintln(out, "\nRun one with: academy -check ID")
		return 0
	}
	ch, ok := checkByID(id)
	if !ok {
		fmt.Fprintf(out, "✗ no checked exercise %q (academy -check list shows them all)\n", id)
		return 2
	}
	file := checkFile(regPath, ch)
	created, err := ensureStarters(regPath, ch)
	if err != nil {
		fmt.Fprintf(out, "✗ could not create the files: %v\n", err)
		return 1
	}
	if len(created) > 0 {
		for _, f := range created {
			fmt.Fprintf(out, "Created %s.\n", homeTilde(f))
		}
		fmt.Fprintf(out, "Write your program there, save, and run academy -check %s again.\n", id)
		for _, l := range wrap(ch.Task, width-4, "  ") {
			fmt.Fprintln(out, l)
		}
		return 1
	}
	run, err := runCheck(ch, file)
	if err != nil {
		fmt.Fprintf(out, "✗ %v\n", err)
		return 1
	}
	fmt.Fprintf(out, "%s · %s\n", sty.Bold(ch.Title), homeTilde(file))
	writeCheckRun(out, run, width)
	if err := appendCheckEvent(logPath, runEvent(run, st[ch.ID].Hints, time.Now())); err != nil {
		fmt.Fprintf(out, "! the result could not be recorded: %v\n", err)
	}
	if run.OK() {
		return 0
	}
	return 1
}
