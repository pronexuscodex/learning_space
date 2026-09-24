package main

// The interactive ledger: dashboard, menu, and every mutating action.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// App owns the in-memory registry. mu guards reg and dirty so a signal
// handler can safely commit while the main loop is blocked on input.
type App struct {
	mu    sync.Mutex
	reg   *Registry
	path  string
	dirty bool
	con   *console
	width int
}

// commit atomically persists the registry to disk.
func (a *App) commit() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reg.LastCommit = time.Now().UTC().Truncate(time.Second)
	if err := atomicWriteJSON(a.path, a.reg); err != nil {
		return err
	}
	a.dirty = false
	return nil
}

// mutate runs fn under the lock and marks state dirty.
func (a *App) mutate(fn func(r *Registry)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	fn(a.reg)
	a.dirty = true
}

func (a *App) printf(format string, args ...any) { fmt.Fprintf(a.con.out, format, args...) }
func (a *App) println(s string)                  { fmt.Fprintln(a.con.out, s) }

// ---------------------------------------------------------------------------
// Badges
// ---------------------------------------------------------------------------

func statusPill(s Status) string {
	if s == StatusGraduated {
		return sty.Bold(sty.Green("★ GRADUATED"))
	}
	return sty.Bold(sty.Yellow("◐ ACTIVE RESEARCH"))
}

func statusMark(s Status) string {
	if s == StatusGraduated {
		return sty.Green("★ graduated")
	}
	return sty.Yellow("◐ active")
}

func compileBadge(c CompileStatus) string {
	switch c {
	case CompileTested:
		return sty.Green("✔ tests passing")
	case CompileOK:
		return sty.Cyan("● compiles")
	case CompileFailing:
		return sty.Red("✗ build failing")
	default:
		return sty.Gray("○ not compiled")
	}
}

// ---------------------------------------------------------------------------
// Dashboard & menu
// ---------------------------------------------------------------------------

func (a *App) printDashboard() {
	a.mu.Lock()
	cs := a.reg.stats(time.Now(), 60)
	dirty := a.dirty
	a.mu.Unlock()

	stat := func(label, value string) string { return sty.Gray(label) + " " + sty.Bold(value) }
	a.println("")
	sep := sty.Gray("  │  ")
	a.println("  " + strings.Join([]string{
		stat("Hours", fmt.Sprintf("%.2f", cs.hours)),
		stat("Labs", fmt.Sprintf("%d", cs.labs)) + sty.Gray(fmt.Sprintf(" (%d★)", cs.labsGrad)),
		stat("Stages", fmt.Sprintf("%d/%d", cs.stagesGrad, cs.stages)),
		stat("Texts", fmt.Sprintf("%d/%d", cs.textsRead, cs.texts)),
	}, sep))
	a.println("  " + strings.Join([]string{
		stat("Concepts", fmt.Sprintf("%d/%d", cs.conceptsStudied, cs.concepts)),
		stat("Exercises", fmt.Sprintf("%d/%d", cs.exercisesDone, cs.exercises)),
	}, sep))

	done := cs.textsRead + cs.conceptsStudied + cs.exercisesDone + cs.stagesGrad
	total := cs.texts + cs.concepts + cs.exercises + cs.stages
	streak := sty.Gray("no streak yet")
	if cs.streak > 0 {
		streak = sty.Bold(sty.Yellow(fmt.Sprintf("▲ %d-day streak", cs.streak)))
	}
	a.printf("  %s %s %s   %s %s  %s\n",
		sty.Gray("Campus"), bar(done, total, 24, sty.Cyan), sty.Bold(pct(done, total)),
		sty.Gray("Last 14d"), sparkline(cs.activity[len(cs.activity)-14:]), streak)
	if dirty {
		a.printf("  %s\n", sty.Yellow("● uncommitted changes"))
	}
}

func (a *App) printMenu() {
	a.printDashboard()
	item := func(key, label string) string { return sty.Cyan("["+key+"]") + " " + label }
	rows := [][2]string{
		{item("1", "View Campus Ledger"), item("5", "Atomic Commit & Exit")},
		{item("2", "Enroll in a New Lab"), item("6", sty.Bold("Study Hall")+sty.Gray(" · learn & practise"))},
		{item("3", "Log Study/Lab Hours"), item("7", "Checkpoint (save, keep going)")},
		{item("4", "Advance Academic Status"), item("8", sty.Gray("Exit without saving"))},
	}
	edge := sty.Gray
	a.println("  " + edge("┌─ ") + sty.Bold("MAIN MENU") + " " + edge(strings.Repeat("─", 58)))
	a.println("  " + edge("│ ") + item("0", sty.Bold(sty.Green("Start Here"))+sty.Gray(" · new? how to learn with this academy")))
	for _, r := range rows {
		a.println("  " + edge("│ ") + padRight(r[0], 32) + r[1])
	}
	a.println("  " + edge("└"+strings.Repeat("─", 70)))
}

// ---------------------------------------------------------------------------
// 1. Campus ledger
// ---------------------------------------------------------------------------

func (a *App) viewLedger() {
	a.mu.Lock()
	defer a.mu.Unlock()
	w := a.width

	var grand float64
	for ti := range a.reg.Tracks {
		t := &a.reg.Tracks[ti]
		color := trackColor(t.ID)
		a.println(heading(fmt.Sprintf("TRACK %s · %s", t.ID, strings.ToUpper(t.Name)), color, w))

		var trackHours float64
		grad := 0
		for si := range t.Stages {
			s := &t.Stages[si]
			trackHours += stageHours(s)
			if s.Status == StatusGraduated {
				grad++
			}
			a.renderStageCard(s, color)
		}
		a.printf("\n  %s %s %d/%d stages graduated · %s\n",
			color(sty.Bold("Track "+t.ID)), bar(grad, len(t.Stages), 16, color),
			grad, len(t.Stages), sty.Bold(fmt.Sprintf("%.2fh", roundHours(trackHours))))
		grand += trackHours
	}

	cs := a.reg.stats(time.Now(), 60)
	a.println("\n" + sty.Gray(strings.Repeat("─", w)))
	a.printf("  %s %s across %d lab(s)   %s %s   %s %s\n",
		sty.Gray("Campus total"), sty.Bold(fmt.Sprintf("%.2fh", roundHours(grand))), a.reg.labCount(),
		sty.Gray("Last 14d"), sparkline(cs.activity[len(cs.activity)-14:]),
		sty.Gray("last commit"), formatTime(a.reg.LastCommit))
}

// renderStageCard prints one stage as a left-bordered card. Caller holds mu.
func (a *App) renderStageCard(s *Stage, color func(string) string) {
	w := a.width
	edge := color("│")
	line := func(content string) { a.println("  " + edge + " " + content) }

	read, total := literatureProgress(s)
	studied, concepts := conceptProgress(s)
	exDone, exTotal := exerciseProgress(s)

	a.println("")
	a.printf("  %s %s %s %s\n", color("╭─"), color(sty.Bold(fmt.Sprintf("Stage %d", s.ID))),
		sty.Gray("·"), sty.Bold(truncate(s.Title, w-18)))
	line(statusPill(s.Status) + sty.Gray(fmt.Sprintf("   %.2fh logged · %d lab(s)", stageHours(s), len(s.Labs))))
	line(padRight(sty.Gray("Reading"), 10) + bar(read, total, 20, sty.Green) + " " + fmt.Sprintf("%d/%d", read, total))
	line(padRight(sty.Gray("Concepts"), 10) + bar(studied, concepts, 20, sty.Blue) + " " + fmt.Sprintf("%d/%d", studied, concepts))
	line(padRight(sty.Gray("Exercises"), 10) + bar(exDone, exTotal, 20, sty.Magenta) + " " + fmt.Sprintf("%d/%d", exDone, exTotal))

	for _, lit := range s.Literature {
		mark := sty.Gray("○")
		if lit.Read {
			mark = sty.Green("✔")
		}
		line("  " + mark + " " + truncate(lit.Title+" — "+lit.Author, w-22) + sty.Gray(" · "+lit.Kind))
	}

	if len(s.Labs) == 0 {
		line(sty.Gray("  ▸ no labs yet — enroll one [2] or pick a blueprint in the Study Hall [6]"))
	} else {
		line(sty.Bold("Labs"))
	}
	for _, l := range s.Labs {
		line(fmt.Sprintf("  %s %s %s %s  %s  %s",
			color("▸"),
			sty.Gray(padRight(fmt.Sprintf("#%d", l.ID), 4)),
			sty.Bold(padRight(truncate(l.Name, 28), 28)),
			sty.Bold(fmt.Sprintf("%7.2fh", l.HoursLogged)),
			padRight(compileBadge(l.CompilationStatus), 16),
			statusMark(l.Status)))
		if l.Architecture != "" {
			line(sty.Gray("      ↳ " + truncate(l.Architecture, w-14)))
		}
	}
	a.println("  " + color("╰─"))
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Local().Format("2006-01-02 15:04")
}

// ---------------------------------------------------------------------------
// Pickers
// ---------------------------------------------------------------------------

// listStages prints a compact stage index, optionally restricted to a track.
func (a *App) listStages(track *Track) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, t := range a.reg.Tracks {
		if track != nil && t.ID != track.ID {
			continue
		}
		color := trackColor(t.ID)
		a.printf("    %s\n", color(sty.Bold("Track "+t.ID+" · "+t.Name)))
		for si := range t.Stages {
			s := &t.Stages[si]
			studied, concepts := conceptProgress(s)
			exDone, exTotal := exerciseProgress(s)
			mark := sty.Yellow("◐")
			if s.Status == StatusGraduated {
				mark = sty.Green("★")
			}
			a.printf("      %s %s %s %s\n", color(padRight(fmt.Sprintf("[%d]", s.ID), 4)), mark,
				padRight(truncate(s.Title, 50), 50),
				sty.Gray(fmt.Sprintf("concepts %d/%d · ex %d/%d", studied, concepts, exDone, exTotal)))
		}
	}
}

// listLabs prints every lab; it returns false if none exist.
func (a *App) listLabs() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.reg.labCount() == 0 {
		a.con.note("No labs enrolled yet. Use [2] or a Study Hall blueprint to enroll one.")
		return false
	}
	for _, t := range a.reg.Tracks {
		color := trackColor(t.ID)
		for _, s := range t.Stages {
			for _, l := range s.Labs {
				a.printf("    %s %s %s %s  %s\n",
					color(padRight(fmt.Sprintf("#%d", l.ID), 4)),
					padRight(truncate(l.Name, 32), 32),
					sty.Gray(fmt.Sprintf("stage %d", s.ID)),
					sty.Bold(fmt.Sprintf("%8.2fh", l.HoursLogged)),
					statusMark(l.Status))
			}
		}
	}
	return true
}

// pickTrack prompts for a track by letter ("A") or list number ("1").
func (a *App) pickTrack() (*Track, error) {
	a.mu.Lock()
	for i, t := range a.reg.Tracks {
		color := trackColor(t.ID)
		a.printf("    %s %s\n", color(fmt.Sprintf("[%d|%s]", i+1, t.ID)), color(t.Name))
	}
	a.mu.Unlock()
	for {
		s, err := a.con.readLine(promptLabel("Track", "letter or number, q cancel"))
		if err != nil {
			return nil, err
		}
		if isCancel(s) {
			return nil, errCancel
		}
		a.mu.Lock()
		for i := range a.reg.Tracks {
			t := &a.reg.Tracks[i]
			if strings.EqualFold(s, t.ID) || s == strconv.Itoa(i+1) {
				a.mu.Unlock()
				return t, nil
			}
		}
		a.mu.Unlock()
		a.con.warn("%q is not a track.", s)
	}
}

// pickStage prompts for a stage ID, optionally restricted to one track.
func (a *App) pickStage(track *Track) (int, error) {
	a.listStages(track)
	return a.con.promptInt("Stage ID", func(id int) error {
		a.mu.Lock()
		defer a.mu.Unlock()
		t, s := a.reg.findStage(id)
		if s == nil || (track != nil && t.ID != track.ID) {
			return fmt.Errorf("stage %d is not in the list above", id)
		}
		return nil
	})
}

// pickLab prompts for an existing lab ID.
func (a *App) pickLab() (int, error) {
	if !a.listLabs() {
		return 0, errCancel
	}
	return a.con.promptInt("Lab ID", func(id int) error {
		a.mu.Lock()
		defer a.mu.Unlock()
		if _, l := a.reg.findLab(id); l == nil {
			return fmt.Errorf("no lab with id %d", id)
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// 2. Enroll
// ---------------------------------------------------------------------------

// errDuplicateLab is returned when a stage already has a lab of that name.
var errDuplicateLab = errors.New("duplicate lab name")

// addLab appends a new lab to a stage and returns its ID.
func (a *App) addLab(stageID int, name, notes string, hours float64) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, s := a.reg.findStage(stageID)
	if s == nil {
		return 0, fmt.Errorf("stage %d not found", stageID)
	}
	for _, l := range s.Labs {
		if strings.EqualFold(l.Name, name) {
			return 0, fmt.Errorf("%w: stage %d already has %q", errDuplicateLab, stageID, l.Name)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	lab := Lab{
		ID:                a.reg.NextLabID,
		Name:              name,
		Architecture:      notes,
		HoursLogged:       hours,
		CompilationStatus: CompileNotBuilt,
		Status:            StatusActive,
		EnrolledAt:        now,
		Log:               []HourEntry{},
	}
	if hours > 0 {
		lab.Log = append(lab.Log, HourEntry{At: now, Hours: hours, Note: "initial hours"})
	}
	s.Labs = append(s.Labs, lab)
	a.reg.NextLabID++
	a.dirty = true
	return lab.ID, nil
}

func (a *App) enrollLab() error {
	a.println(heading("ENROLL IN A NEW LAB", sty.Green, a.width))
	track, err := a.pickTrack()
	if err != nil {
		return err
	}
	stageID, err := a.pickStage(track)
	if err != nil {
		return err
	}

	var name string
	for {
		if name, err = a.con.promptText("Lab name", maxNameLen, true); err != nil {
			return err
		}
		a.mu.Lock()
		_, s := a.reg.findStage(stageID)
		dup := false
		for _, l := range s.Labs {
			dup = dup || strings.EqualFold(l.Name, name)
		}
		a.mu.Unlock()
		if !dup {
			break
		}
		a.con.warn("Stage %d already has a lab named %q.", stageID, name)
	}

	notes, err := a.con.promptText("Architecture notes (optional)", maxNotesLen, false)
	if err != nil {
		return err
	}
	hours, err := a.con.promptHours("Initial lab hours", fmt.Sprintf("0-%.0f, blank = 0", maxInitialHours), maxInitialHours, true)
	if err != nil {
		return err
	}

	id, err := a.addLab(stageID, name, notes, hours)
	if err != nil {
		return err
	}
	a.con.ok("Enrolled lab #%d %s in stage %d.", id, sty.Bold(name), stageID)
	return nil
}

// ---------------------------------------------------------------------------
// 3. Log hours
// ---------------------------------------------------------------------------

func (a *App) logHours() error {
	a.println(heading("LOG STUDY / LAB HOURS", sty.Yellow, a.width))
	labID, err := a.pickLab()
	if err != nil {
		return err
	}
	hours, err := a.con.promptHours("Hours to add", fmt.Sprintf("1.5 or 1h30m, max %.0f", maxHoursPerEntry), maxHoursPerEntry, false)
	if err != nil {
		return err
	}
	note, err := a.con.promptText("Session note (optional)", maxLogNoteLen, false)
	if err != nil {
		return err
	}

	var total float64
	var name string
	a.mutate(func(r *Registry) {
		_, l := r.findLab(labID)
		l.HoursLogged = roundHours(l.HoursLogged + hours)
		l.Log = append(l.Log, HourEntry{At: time.Now().UTC().Truncate(time.Second), Hours: hours, Note: note})
		total, name = l.HoursLogged, l.Name
	})
	a.con.ok("+%.2fh on %s — cumulative %s.", hours, sty.Bold(name), sty.Bold(fmt.Sprintf("%.2fh", total)))
	return nil
}

// ---------------------------------------------------------------------------
// 4. Advance status
// ---------------------------------------------------------------------------

func (a *App) advanceStatus() error {
	a.println(heading("ADVANCE ACADEMIC STATUS", sty.Magenta, a.width))
	choice, err := a.con.promptChoice("Action", []string{
		"Toggle a Stage  (Active Research ⇄ Mastered/Graduated)",
		"Toggle a Lab    (Active Research ⇄ Mastered/Graduated)",
		"Mark required literature read/unread",
		"Set a Lab's compilation status",
	})
	if err != nil {
		return err
	}
	switch choice {
	case 0:
		return a.toggleStage()
	case 1:
		return a.toggleLab()
	case 2:
		return a.toggleLiterature()
	default:
		return a.setCompileStatus()
	}
}

func (a *App) toggleStage() error {
	id, err := a.pickStage(nil)
	if err != nil {
		return err
	}

	a.mu.Lock()
	_, s := a.reg.findStage(id)
	next := s.Status.Toggled()
	read, total := literatureProgress(s)
	studied, concepts := conceptProgress(s)
	exDone, exTotal := exerciseProgress(s)
	activeLabs := 0
	for _, l := range s.Labs {
		if l.Status != StatusGraduated {
			activeLabs++
		}
	}
	a.mu.Unlock()

	if next == StatusGraduated && (read < total || studied < concepts || exDone < exTotal || activeLabs > 0) {
		a.con.warn("Stage %d still has %d unread text(s), %d unstudied concept(s), %d open exercise(s) and %d active lab(s).",
			id, total-read, concepts-studied, exTotal-exDone, activeLabs)
		ok, err := a.con.confirm("Graduate anyway?")
		if err != nil {
			return err
		}
		if !ok {
			return errCancel
		}
	}
	a.mutate(func(r *Registry) {
		_, s := r.findStage(id)
		s.Status = next
	})
	a.con.ok("Stage %d is now %s.", id, statusPill(next))
	return nil
}

func (a *App) toggleLab() error {
	id, err := a.pickLab()
	if err != nil {
		return err
	}

	a.mu.Lock()
	_, l := a.reg.findLab(id)
	next := l.Status.Toggled()
	cs := l.CompilationStatus
	a.mu.Unlock()

	if next == StatusGraduated && cs != CompileOK && cs != CompileTested {
		a.con.warn("Lab #%d is %q.", id, cs)
		ok, err := a.con.confirm("Graduate a lab that does not compile?")
		if err != nil {
			return err
		}
		if !ok {
			return errCancel
		}
	}
	a.mutate(func(r *Registry) {
		_, l := r.findLab(id)
		l.Status = next
	})
	a.con.ok("Lab #%d is now %s.", id, statusPill(next))
	return nil
}

func (a *App) toggleLiterature() error {
	stageID, err := a.pickStage(nil)
	if err != nil {
		return err
	}

	a.mu.Lock()
	_, s := a.reg.findStage(stageID)
	options := make([]string, len(s.Literature))
	for i, lit := range s.Literature {
		mark := sty.Gray("○")
		if lit.Read {
			mark = sty.Green("✔")
		}
		options[i] = fmt.Sprintf("%s %s %s", mark, lit.Title, sty.Gray("— "+lit.Author))
	}
	a.mu.Unlock()

	if len(options) == 0 {
		a.con.note("Stage %d has no required literature.", stageID)
		return nil
	}
	idx, err := a.con.promptChoice("Toggle which text", options)
	if err != nil {
		return err
	}

	var title string
	var read bool
	a.mutate(func(r *Registry) {
		_, s := r.findStage(stageID)
		lit := &s.Literature[idx]
		lit.Read = !lit.Read
		title, read = lit.Title, lit.Read
	})
	if read {
		a.con.ok("%s marked %s.", sty.Bold(title), sty.Green("read"))
	} else {
		a.con.ok("%s marked %s.", sty.Bold(title), sty.Yellow("unread"))
	}
	return nil
}

func (a *App) setCompileStatus() error {
	id, err := a.pickLab()
	if err != nil {
		return err
	}
	options := make([]string, len(compileStatuses))
	for i, cs := range compileStatuses {
		options[i] = compileBadge(cs)
	}
	idx, err := a.con.promptChoice("Compilation status", options)
	if err != nil {
		return err
	}
	a.mutate(func(r *Registry) {
		_, l := r.findLab(id)
		l.CompilationStatus = compileStatuses[idx]
	})
	a.con.ok("Lab #%d compilation status: %s.", id, compileBadge(compileStatuses[idx]))
	return nil
}

// ---------------------------------------------------------------------------
// Commit & main loop
// ---------------------------------------------------------------------------

// commitAndExit persists state and terminates the process.
func (a *App) commitAndExit() {
	if err := a.commit(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ commit failed: %v\n  (in-memory changes were NOT saved)\n", err)
		os.Exit(1)
	}
	a.con.ok("Registry committed atomically to %s", a.path)
	os.Exit(0)
}

// run is the interactive loop. Ctrl-D (EOF) is treated as commit & exit.
func (a *App) run() {
	for {
		a.printMenu()
		choice, err := a.con.readLine(sty.Bold(sty.Cyan("  academy")) + sty.Cyan(" › "))
		if err != nil {
			a.con.note("Input closed — committing.")
			a.commitAndExit()
		}

		var actionErr error
		switch strings.ToLower(choice) {
		case "0":
			a.showStartHere()
		case "1":
			a.viewLedger()
		case "2":
			actionErr = a.enrollLab()
		case "3":
			actionErr = a.logHours()
		case "4":
			actionErr = a.advanceStatus()
		case "5":
			a.commitAndExit()
		case "6":
			actionErr = a.studyHall()
		case "7":
			if err := a.commit(); err != nil {
				a.con.fail("Checkpoint failed: %v", err)
			} else {
				a.con.ok("Checkpoint written to %s", a.path)
			}
		case "8":
			ok, err := a.con.confirm("Discard all uncommitted changes and exit?")
			if err != nil {
				a.commitAndExit()
			}
			if ok {
				a.con.note("Exited without saving.")
				os.Exit(0)
			}
		case "":
			// Blank line: just redraw the menu.
		default:
			a.con.warn("%q is not a menu option. Choose 0-8.", choice)
		}

		switch {
		case errors.Is(actionErr, errCancel):
			a.con.note("Cancelled — nothing changed.")
		case errors.Is(actionErr, io.EOF):
			a.con.note("Input closed — committing.")
			a.commitAndExit()
		case actionErr != nil:
			a.con.fail("%v", actionErr)
		}
	}
}
