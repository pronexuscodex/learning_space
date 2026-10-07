package main

// The interactive ledger: dashboard, menu, and every mutating action.

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	width int  // fixed layout width when output is not a terminal
	pager bool // page long screens (stdin is a terminal)

	opMu      sync.Mutex
	interrupt func() // set while Ctrl+C should cancel an operation, not quit

	// Saving is automatic. The registry as it was when the academy opened
	// is kept, so "undo this session" can put it back, and one backup is
	// made per session (the first save), so backups span days, not clicks.
	original []byte // nil: there was no registry file
	backedUp bool
	onDisk   bool // the file has been written this session
}

// onInterrupt makes Ctrl+C call f instead of committing and quitting, until
// it is called again with nil.
func (a *App) onInterrupt(f func()) {
	a.opMu.Lock()
	a.interrupt = f
	a.opMu.Unlock()
}

// interruptOp cancels the running operation, if there is one.
func (a *App) interruptOp() bool {
	a.opMu.Lock()
	f := a.interrupt
	a.interrupt = nil
	a.opMu.Unlock()
	if f == nil {
		return false
	}
	f()
	return true
}

// commit atomically persists the registry to disk. The first save of a
// session backs up the file as it was before.
func (a *App) commit() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reg.LastCommit = time.Now().UTC().Truncate(time.Second)
	if !a.backedUp {
		if err := backupRegistry(a.path, time.Now()); err != nil {
			// A failed backup must not block saving the learner's work.
			a.con.warn("Could not back up the previous registry: %v", err)
		}
		a.backedUp = true
	}
	if err := atomicWriteJSON(a.path, a.reg); err != nil {
		return err
	}
	a.dirty = false
	a.onDisk = true
	return nil
}

// autosave saves when something changed. It runs after every action, so
// closing the window or a crash loses at most the action in progress.
func (a *App) autosave() {
	a.mu.Lock()
	dirty := a.dirty
	a.mu.Unlock()
	if !dirty {
		return
	}
	if err := a.commit(); err != nil {
		a.con.fail("Could not save your progress: %v. It is still in memory; the academy will try again after your next action.", err)
	}
}

// rememberOriginal keeps the registry file as it is now, for undo.
func (a *App) rememberOriginal() {
	if data, err := os.ReadFile(a.path); err == nil {
		a.original = data
	}
}

// undoSession puts the registry back as it was when the academy opened.
func (a *App) undoSession() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.onDisk {
		return nil // nothing was written this session
	}
	if a.original == nil {
		err := os.Remove(a.path)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	return atomicWriteFile(a.path, a.original)
}

// mutate runs fn under the lock and marks state dirty.
func (a *App) mutate(fn func(r *Registry)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	fn(a.reg)
	a.dirty = true
}

// cols is the layout width. On a terminal it is measured live, so after a
// resize the next screen (or Ctrl+L) fits the new size.
func (a *App) cols() int {
	if a.con.screen {
		return termWidth()
	}
	return a.width
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
	classic := a.reg.ClassicMode
	a.mu.Unlock()
	w := a.cols()

	stat := func(label, value string) string { return sty.Gray(label) + " " + sty.Bold(value) }
	due := stat("Due", strconv.Itoa(cs.due))
	if cs.due > 0 {
		due = sty.Gray("Due") + " " + sty.Bold(sty.Yellow(strconv.Itoa(cs.due)))
	}
	a.println("")
	for _, l := range flow([]string{
		stat("Study", fmt.Sprintf("%.1f h", cs.hours)),
		stat("Labs", fmt.Sprintf("%d", cs.labs)) + sty.Gray(fmt.Sprintf(" (%d★)", cs.labsGrad)),
		stat("Stages", fmt.Sprintf("%d/%d", cs.stagesGrad, cs.stages)),
		stat("Texts", fmt.Sprintf("%d/%d", cs.textsRead, cs.texts)),
		stat("Concepts", fmt.Sprintf("%d/%d", cs.conceptsStudied, cs.concepts)),
		stat("Exercises", fmt.Sprintf("%d/%d", cs.exercisesDone, cs.exercises)),
		stat("Mastered", fmt.Sprintf("%d/%d", cs.mastered, cs.concepts)),
		due,
	}, sty.Gray("  │  "), w, "  ") {
		a.println(l)
	}

	// Campus progress weights long-term mastery most: each concept climbs
	// four mastery levels, plus reading and graduated stages.
	done := cs.masteryPoints + cs.textsRead + cs.stagesGrad
	total := cs.masteryMax + cs.texts + cs.stages
	streak := sty.Gray("no streak yet")
	if cs.streak > 0 {
		streak = sty.Bold(sty.Yellow(fmt.Sprintf("▲ %d-day streak", cs.streak)))
	}
	barWidth := max(10, min(24, w-20))
	for _, l := range flow([]string{
		sty.Gray("Campus") + " " + bar(done, total, barWidth, sty.Cyan) + " " + sty.Bold(pct(done, total)),
		sty.Gray("Last 14d") + " " + sparkline(cs.activity[len(cs.activity)-14:]),
		streak,
	}, "   ", w, "  ") {
		a.println(l)
	}
	if classic {
		for _, l := range flow([]string{classicBadge(), sty.Gray("hints behind the struggle clock"), sty.Gray("notebook on"), sty.Gray("type-ins in each Classic corner")}, sty.Gray(" · "), w, "  ") {
			a.println(l)
		}
	}
	for _, l := range a.goalsLine() {
		a.println(l)
	}
	a.println(a.wordOfTheDayLine())
	a.mu.Lock()
	day := a.reg.todayWorkout(time.Now())
	a.mu.Unlock()
	if day == nil || !day.complete() {
		a.println(a.workoutLine())
	} else if h := a.nextHint(); h != "" {
		a.println("  " + h)
	}
}

// printMenu draws the dashboard and the menu. Wide terminals get two
// columns; narrow ones get a single column, and hints are dropped first.
func (a *App) printMenu() {
	a.printDashboard()
	w := a.cols()
	a.mu.Lock()
	dueNow := len(a.reg.dueCards(time.Now()))
	classic := a.reg.ClassicMode
	tidy := a.reg.TidyScreen
	a.mu.Unlock()

	item := func(key, label string) string { return sty.Cyan("["+key+"]") + " " + label }
	hint := func(s string) string { return sty.Gray(" · " + s) }
	review := sty.Bold(sty.Green("Daily Review"))
	if dueNow > 0 {
		review += " " + sty.Bold(sty.Yellow(fmt.Sprintf("(%d due)", dueNow)))
	}
	mode := sty.Gray("off")
	if classic {
		mode = sty.Bold(sty.Yellow("ON"))
	}
	tidyMode := sty.Gray("off")
	if tidy {
		tidyMode = sty.Bold(sty.Green("on"))
	}
	type entry struct{ text, extra, short string } // extra is shown only when it fits; short is for narrow screens
	groups := []struct {
		title   string
		color   func(string) string
		entries []entry
	}{
		{"LEARN", sty.Green, []entry{
			{item("0", sty.Bold(sty.Green("Start Here"))), hint("begin here"), "Start Here"},
			{item("j", sty.Bold(sty.Green("Today's workout"))), hint("guided daily session"), "Workout"},
			{item("6", sty.Bold("Study Hall")), hint("learn & practise"), "Study Hall"},
			{item("9", review), "", "Review"},
			{item("n", sty.Bold(sty.Green("What's next"))), hint("best next step"), "What's next"},
			{item("m", "Roadmap"), hint(fmt.Sprintf("all %d stages", len(curriculum))), "Roadmap"},
			{item("d", sty.Bold("Dictionary")), hint("every word explained"), "Dictionary"},
			{item("/", "Search"), hint("find any topic"), "Search"},
		}},
		{"PRACTISE & TRACK", sty.Cyan, []entry{
			{item("f", "Focus timer"), hint("Pomodoro"), "Focus timer"},
			{item("2", "Start a lab project"), "", "Start lab"},
			{item("3", "Log study hours"), "", "Log hours"},
			{item("4", "Update progress"), hint("graduate, mark read"), "Update"},
			{item("1", "Ledger"), hint("every stage at a glance"), "Ledger"},
			{item("g", "Weekly goals"), "", "Goals"},
			{item("p", "Progress report"), hint("calendar & trends"), "Progress"},
			{item("a", "Achievements"), "", "Achievements"},
			{item("k", "Share card"), hint("post your progress"), "Share card"},
		}},
		{"RESOURCES", sty.Magenta, []entry{
			{item("l", sty.Bold("Library")), hint("download PDFs"), "Library"},
			{item("+", "My resources"), hint("your finds"), "My resources"},
			{item("w", sty.Bold("Tech watch")), hint("keep up with trends"), "Tech watch"},
			{item("x", "Export notes"), hint("to Markdown"), "Export"},
		}},
		{"SETTINGS & EXIT", sty.Yellow, []entry{
			{item("o", "Settings"), hint("theme, symbols, modes"), "Settings"},
			{item("?", "Keys & shortcuts"), hint("Ctrl+L clears"), "Keys"},
			{item("c", "Classic Mode "+mode), "", "Classic " + stripANSI(mode)},
			{item("t", "Tidy screen "+tidyMode), "", "Tidy " + stripANSI(tidyMode)},
			{item("5", "Exit"), hint("progress saves itself"), "Exit"},
			{item("8", sty.Gray("Undo this session")), "", "Undo session"},
		}},
	}

	edge := sty.Gray
	inner := w - 4 // room after "  │ "
	title := "┌─ " + sty.Bold("MAIN MENU") + " "
	a.println("  " + edge("┌─ ") + sty.Bold("MAIN MENU") + " " + edge(strings.Repeat("─", max(0, w-2-visibleLen(title)))))
	show := func(e entry, room int) string {
		if visibleLen(e.text+e.extra) <= room {
			return e.text + e.extra
		}
		return truncate(stripANSI(e.text), room)
	}
	// One, two or three columns, depending on the width.
	const colWidth = 32
	columns := 1
	switch {
	case inner >= 3*colWidth:
		columns = 3
	case inner >= colWidth+34:
		columns = 2
	}
	narrow := columns == 1 && inner >= 2*17 // two compact columns with short labels
	for gi, g := range groups {
		if gi > 0 {
			a.println("  " + edge("│"))
		}
		a.println("  " + edge("│ ") + g.color(sty.Bold(truncate(g.title, inner))))
		if narrow {
			half := inner / 2
			for i := 0; i < len(g.entries); i += 2 {
				key := func(e entry) string { return strings.SplitN(stripANSI(e.text), " ", 2)[0] }
				cell := func(e entry, room int) string {
					label := e.short
					if key(e) == "[9]" && dueNow > 0 {
						label += fmt.Sprintf(" (%d)", dueNow)
					}
					return sty.Cyan(key(e)) + " " + truncate(label, room-visibleLen(key(e))-1)
				}
				line := padRight(cell(g.entries[i], half-1), half)
				if i+1 < len(g.entries) {
					line += cell(g.entries[i+1], inner-half)
				}
				a.println("  " + edge("│ ") + line)
			}
			continue
		}
		for i := 0; i < len(g.entries); i += columns {
			line := ""
			for c := 0; c < columns && i+c < len(g.entries); c++ {
				last := c == columns-1 || i+c == len(g.entries)-1
				if last {
					line += show(g.entries[i+c], inner-c*colWidth)
				} else {
					line += padRight(show(g.entries[i+c], colWidth-1), colWidth)
				}
			}
			a.println("  " + edge("│ ") + line)
		}
	}
	a.println("  " + edge("└"+strings.Repeat("─", max(0, w-3))))
}

// ---------------------------------------------------------------------------
// 1. Campus ledger
// ---------------------------------------------------------------------------

func (a *App) viewLedger() {
	a.mu.Lock()
	defer a.mu.Unlock()
	w := a.cols()

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
		a.println("")
		for _, l := range flow([]string{
			color(sty.Bold("Track "+t.ID)) + " " + bar(grad, len(t.Stages), min(16, max(6, w-30)), color),
			fmt.Sprintf("%d/%d stages graduated", grad, len(t.Stages)),
			sty.Bold(fmt.Sprintf("%.2fh", roundHours(trackHours))),
		}, sty.Gray(" · "), w, "  ") {
			a.println(l)
		}
		grand += trackHours
	}

	cs := a.reg.stats(time.Now(), 60)
	a.println("\n" + sty.Gray(strings.Repeat("─", w)))
	for _, l := range flow([]string{
		sty.Gray("Campus total") + " " + sty.Bold(fmt.Sprintf("%.2fh", roundHours(grand))) + fmt.Sprintf(" across %d lab(s)", a.reg.labCount()),
		sty.Gray("Last 14d") + " " + sparkline(cs.activity[len(cs.activity)-14:]),
		sty.Gray("last saved") + " " + formatTime(a.reg.LastCommit),
	}, "   ", w, "  ") {
		a.println(l)
	}
}

// renderStageCard prints one stage as a left-bordered card that fits the
// current width. Caller holds mu.
func (a *App) renderStageCard(s *Stage, color func(string) string) {
	w := a.cols()
	inner := w - 4 // room after "  │ "
	edge := color("│")
	line := func(content string) { a.println("  " + edge + " " + content) }
	flowed := func(sep string, items ...string) {
		for _, l := range flow(items, sep, inner, "") {
			line(l)
		}
	}

	read, total := literatureProgress(s)
	studied, concepts := conceptProgress(s)
	exDone, exTotal := exerciseProgress(s)
	barW := max(6, min(20, inner-16))

	a.println("")
	a.printf("  %s %s %s %s\n", color("╭─"), color(sty.Bold(fmt.Sprintf("Stage %d", s.ID))),
		sty.Gray("·"), sty.Bold(truncate(s.Title, w-16)))
	flowed("   ", statusPill(s.Status), sty.Gray(fmt.Sprintf("%.2fh logged · %d lab(s)", stageHours(s), len(s.Labs))))
	line(padRight(sty.Gray("Reading"), 10) + bar(read, total, barW, sty.Green) + " " + fmt.Sprintf("%d/%d", read, total))
	line(padRight(sty.Gray("Concepts"), 10) + bar(studied, concepts, barW, sty.Blue) + " " + fmt.Sprintf("%d/%d", studied, concepts))
	line(padRight(sty.Gray("Exercises"), 10) + bar(exDone, exTotal, barW, sty.Magenta) + " " + fmt.Sprintf("%d/%d", exDone, exTotal))
	if g, ok := guideFor(s.ID); ok {
		var glyphs strings.Builder
		for _, c := range g.Concepts {
			glyphs.WriteString(masteryBadge(a.reg.conceptMastery(s, c)))
		}
		_, _, mastered, total := a.reg.stageMastery(s)
		check := sty.Gray("check not taken")
		if s.MasteryCheck != nil && s.MasteryCheck.Passed {
			check = sty.Green("✓ check passed")
		} else if s.MasteryCheck != nil {
			check = sty.Yellow(fmt.Sprintf("check %d/%d", s.MasteryCheck.Score, s.MasteryCheck.Total))
		}
		typeIn := ""
		if s.TypeInDone {
			typeIn = sty.Green("⌨ type-in ✓")
		}
		flowed("  ", padRight(sty.Gray("Mastery"), 10)+glyphs.String()+" "+fmt.Sprintf("%d/%d", mastered, total), check, typeIn)
	}

	for _, lit := range s.Literature {
		mark := sty.Gray("○")
		if lit.Read {
			mark = sty.Green("✔")
		}
		line("  " + mark + " " + truncate(lit.Title+" — "+lit.Author, inner-4-len(lit.Kind)-3) + sty.Gray(" · "+lit.Kind))
	}

	if len(s.Labs) == 0 {
		for i, l := range wrap("no labs yet — enroll one [2] or pick a blueprint in the Study Hall [6]", inner-4, "") {
			lead := "    "
			if i == 0 {
				lead = "  ▸ "
			}
			line(sty.Gray(lead + l))
		}
	} else {
		line(sty.Bold("Labs"))
	}
	for _, l := range s.Labs {
		id := sty.Gray(padRight(fmt.Sprintf("#%d", l.ID), 4))
		hours := sty.Bold(fmt.Sprintf("%.2fh", l.HoursLogged))
		if inner >= 72 { // one row: name, hours, build, status
			nameW := min(28, inner-46)
			line(fmt.Sprintf("  %s %s %s %s  %s  %s", color("▸"), id,
				sty.Bold(padRight(truncate(l.Name, nameW), nameW)), padRight(hours, 8),
				padRight(compileBadge(l.CompilationStatus), 16), statusMark(l.Status)))
		} else { // narrow: name on one row, details on the next
			line(fmt.Sprintf("  %s %s %s", color("▸"), id, sty.Bold(truncate(l.Name, inner-10))))
			for _, d := range flow([]string{hours, compileBadge(l.CompilationStatus), statusMark(l.Status)}, sty.Gray(" · "), inner, "        ") {
				line(d)
			}
		}
		if l.Architecture != "" {
			line(sty.Gray("      ↳ " + truncate(l.Architecture, inner-8)))
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
	w := a.cols()
	for _, t := range a.reg.Tracks {
		if track != nil && t.ID != track.ID {
			continue
		}
		color := trackColor(t.ID)
		a.printf("    %s\n", color(sty.Bold(truncate("Track "+t.ID+" · "+t.Name, w-6))))
		for si := range t.Stages {
			s := &t.Stages[si]
			studied, concepts := conceptProgress(s)
			exDone, exTotal := exerciseProgress(s)
			state := a.reg.stageState(s)
			mark := stateStyle(state)(stageStateGlyphs[state])
			// Right-aligned in a fixed column, so two-digit counts line up.
			counts := fmt.Sprintf("%25s", fmt.Sprintf("concepts %d/%d · ex %d/%d", studied, concepts, exDone, exTotal))
			titleW := w - 14 - visibleLen(counts) - 1
			if titleW < 24 { // too narrow for the counts: give the title the room
				titleW, counts = w-14, ""
			}
			a.printf("      %s %s %s %s\n", color(padRight(fmt.Sprintf("[%d]", s.ID), 4)), mark,
				padRight(truncate(s.Title, titleW), titleW), sty.Gray(counts))
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
	w := a.cols()
	for _, t := range a.reg.Tracks {
		color := trackColor(t.ID)
		for _, s := range t.Stages {
			for _, l := range s.Labs {
				nameW := max(12, min(32, w-36))
				for i, row := range flow([]string{
					color(padRight(fmt.Sprintf("#%d", l.ID), 4)) + " " + padRight(truncate(l.Name, nameW), nameW),
					sty.Gray(fmt.Sprintf("stage %d", s.ID)) + " " + sty.Bold(fmt.Sprintf("%.2fh", l.HoursLogged)),
					statusMark(l.Status),
				}, "  ", w-4, "") {
					if i > 0 {
						row = "     " + row
					}
					a.printf("    %s\n", row)
				}
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
	legend := make([]string, len(stageStateNames))
	for i, name := range stageStateNames {
		legend[i] = stateStyle(i)(stageStateGlyphs[i]) + " " + sty.Gray(name)
	}
	for _, l := range flow(legend, "  ", a.cols(), "    ") {
		a.println(l)
	}
	return a.con.promptInt("Which stage?", func(id int) error {
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
	a.println(heading("ENROLL IN A NEW LAB", sty.Green, a.cols()))
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
	a.println(heading("LOG STUDY / LAB HOURS", sty.Yellow, a.cols()))
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
	a.println(heading("ADVANCE ACADEMIC STATUS", sty.Magenta, a.cols()))
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
	checkPassed := s.MasteryCheck != nil && s.MasteryCheck.Passed
	activeLabs := 0
	for _, l := range s.Labs {
		if l.Status != StatusGraduated {
			activeLabs++
		}
	}
	a.mu.Unlock()

	if next == StatusGraduated && (read < total || studied < concepts || exDone < exTotal || activeLabs > 0 || !checkPassed) {
		a.con.warn("Stage %d still has %d unread text(s), %d unstudied concept(s), %d open exercise(s) and %d active lab(s).",
			id, total-read, concepts-studied, exTotal-exDone, activeLabs)
		if !checkPassed {
			a.con.warn("Its mastery check has not been passed yet (Study Hall → Mastery check).")
		}
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
	restoreTerminal()
	if err := a.commit(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ saving failed: %v\n  (your latest changes were NOT saved)\n", err)
		os.Exit(1)
	}
	a.con.ok("Progress saved. See you next time!")
	os.Exit(0)
}

// showShortcuts lists every key the academy understands.
func (a *App) showShortcuts() {
	a.println(heading("KEYS & SHORTCUTS", sty.Cyan, a.cols()))
	a.println("")
	row := func(key, what string) {
		for i, l := range wrap(what, a.cols()-20, "") {
			k := ""
			if i == 0 {
				k = key
			}
			a.printf("    %s %s\n", sty.Bold(sty.Cyan(padRight(k, 14))), l)
		}
	}
	a.printf("  %s\n", sty.Bold("While typing (any prompt)"))
	row("Ctrl+L", "clear the screen and redraw, keeping what you have typed")
	row("Backspace", "delete the previous character")
	row("Ctrl+U", "erase the whole line")
	row("Ctrl+W", "erase the previous word")
	row("Ctrl+D", "end input: saves and exits (on an empty line)")
	row("Ctrl+C", "save and exit (closing the window saves too)")
	row("q  or  :q", "cancel the current prompt (q for numbers, :q for text)")
	a.println("")
	a.printf("  %s\n", sty.Bold("Main menu"))
	row("0 – 9", "the menu options")
	row("r", "Daily Review (same as 9)")
	row("n", "What's next: your best next steps, one key to start")
	row("/  or  s", "search concepts, glossary, resources and classics")
	row("f", "focus timer: a timed study session that is logged")
	row("p", "progress report: activity calendar, weekly trend, weak cards")
	row("x", "export all your notes and progress to a Markdown file")
	row("d", "programmer's dictionary: jargon explained with analogies and code")
	row("w", "tech watch: the method, live headlines, a watch log and your tech radar")
	row("+", "my resources: add, edit, import and export your own finds")
	row("l", "library: download free books and papers as PDFs and open them")
	row("m", "roadmap: every stage's state and what it builds on; open any Study Hall")
	row("g", "weekly goals: minutes, days and review cards, tracked on the dashboard")
	row("a", "achievements")
	row("o", "settings: colour theme, plain symbols, confidence ratings, modes")
	row("j", "today's workout: review, recall, learn, mixed practice, reflect")
	row("k", "share card: your progress, ready to screenshot and post")
	row("c", "Classic Mode on/off")
	row("t", "tidy screen on/off: start every action on a clean screen")
	row("clear / cls", "clear the screen now")
	row("? / h", "this list")
	a.println("")
	a.printf("  %s\n", sty.Bold("Long screens"))
	row("Enter", "next page")
	row("q", "stop paging and continue")
	a.println("")
	a.printf("  %s\n", sty.Gray(truncate(versionLine(), a.cols()-4)))
	if !a.con.raw {
		for _, l := range wrap("Your terminal is in line mode (for example on Windows, or when input is piped), so Ctrl+L takes effect after you press Enter.", a.cols()-6, "  ") {
			a.println(sty.Gray(l))
		}
	}
}

// run is the interactive loop. Ctrl-D (EOF) is treated as save & exit.
func (a *App) run() {
	for {
		a.printMenu()
		a.con.setOnClear(a.printMenu) // Ctrl+L and resizes redraw the menu on a clean screen
		choice, err := a.con.readLine(sty.Bold(sty.Cyan("  academy")) + sty.Cyan(" › "))
		a.con.setOnClear(nil)

		// Tidy screen: each chosen action starts on a clean screen, and its
		// output (including its final message) stays visible above the menu.
		a.mu.Lock()
		tidy := a.reg.TidyScreen
		a.mu.Unlock()
		if tidy && a.con.screen && choice != "" {
			fmt.Fprint(a.con.out, clearSeq)
		}
		if err != nil {
			a.con.note("Input closed — saving.")
			a.commitAndExit()
		}

		var actionErr error
		switch strings.ToLower(choice) {
		case "0":
			a.paged(true, a.showStartHere)
		case "9", "r":
			actionErr = a.dailyReview()
		case "c":
			actionErr = a.toggleClassicMode()
		case "clear", "cls":
			if a.con.screen {
				fmt.Fprint(a.con.out, clearSeq)
			}
		case "t":
			a.mutate(func(r *Registry) { r.TidyScreen = !r.TidyScreen })
			a.mu.Lock()
			on := a.reg.TidyScreen
			a.mu.Unlock()
			if on {
				a.con.ok("Tidy screen on: every action now starts on a clean screen. Press t again to turn it off.")
			} else {
				a.con.ok("Tidy screen off: earlier output stays visible above the menu.")
			}
		case "?", "h", "help":
			a.paged(true, a.showShortcuts)
		case "n":
			actionErr = a.whatsNext()
		case "j":
			actionErr = a.dailyWorkout()
		case "o":
			actionErr = a.settingsScreen()
		case "/", "s", "search":
			actionErr = a.searchScreen()
		case "f":
			actionErr = a.focusTimer()
		case "p":
			a.paged(true, a.progressReport)
		case "x":
			actionErr = a.exportNotes()
		case "m":
			actionErr = a.roadmap()
		case "d":
			actionErr = a.dictionary()
		case "w":
			actionErr = a.techWatch()
		case "+", "my":
			actionErr = a.myResources()
		case "l":
			actionErr = a.library(AllStages)
		case "g":
			actionErr = a.setGoals()
		case "a":
			a.paged(true, a.showAchievements)
		case "k":
			a.showShareCard()
		case "1":
			a.paged(true, a.viewLedger)
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
		case "7", "save":
			if err := a.commit(); err != nil {
				a.con.fail("Saving failed: %v", err)
			} else {
				a.con.ok("Saved. (The academy also saves by itself after every action.)")
			}
		case "8":
			ok, err := a.con.confirm("Undo everything you did since you opened the academy, and exit?")
			if err != nil {
				a.commitAndExit()
			}
			if ok {
				restoreTerminal()
				if err := a.undoSession(); err != nil {
					fmt.Fprintf(os.Stderr, "✗ could not undo: %v\n", err)
					os.Exit(1)
				}
				a.con.note("This session was undone. Your progress is as it was when you opened the academy.")
				os.Exit(0)
			}
		case "":
			// Blank line: just redraw the menu.
		default:
			a.con.warn("%q is not a menu option. Choose 0-9, a letter from the menu, or ? for help.", choice)
		}

		switch {
		case errors.Is(actionErr, errCancel):
			a.con.note("Cancelled — nothing changed.")
		case errors.Is(actionErr, io.EOF):
			a.con.note("Input closed — saving.")
			a.commitAndExit()
		case actionErr != nil:
			a.con.fail("%v", actionErr)
		}
		a.announceAchievements()
		a.autosave()
	}
}
