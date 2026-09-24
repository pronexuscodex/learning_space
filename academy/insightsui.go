package main

// Screens for What's next, Search, the Focus timer, the Progress report
// and the Markdown export.

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// What's next
// ---------------------------------------------------------------------------

// whatsNext shows the recommended next steps and runs the chosen one.
func (a *App) whatsNext() error {
	a.mu.Lock()
	steps := a.reg.nextSteps(time.Now())
	a.mu.Unlock()

	a.println(heading("WHAT'S NEXT · your best next steps", sty.Green, a.cols()))
	if len(steps) == 0 {
		a.con.ok("You are all caught up: nothing due and nothing open. Browse the Study Hall [6] or start a focus session [f].")
		return nil
	}
	for i, s := range steps {
		a.println("")
		for j, l := range wrap(s.Title, a.cols()-5, "") {
			num := "  "
			if j == 0 {
				num = fmt.Sprintf("%d.", i+1)
			}
			a.printf("  %s %s\n", sty.Green(num), sty.Bold(l))
		}
		for _, l := range wrap(s.Why, a.cols()-8, "     ") {
			a.println(sty.Gray(l))
		}
	}
	a.println("")
	n, err := a.con.promptInt("Start which one?", func(n int) error {
		if n < 1 || n > len(steps) {
			return fmt.Errorf("choose 1 to %d", len(steps))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return a.doSuggestion(steps[n-1])
}

// doSuggestion jumps straight into a suggested activity.
func (a *App) doSuggestion(s Suggestion) error {
	g, _ := guideFor(s.Stage)
	switch s.Kind {
	case SuggestReview:
		return a.dailyReview()
	case SuggestConcept:
		return a.studyConceptAt(s.Stage, g, s.Concept)
	case SuggestExercise:
		return a.workExercise(s.Stage, g.Concepts[s.Concept], s.Exercise)
	case SuggestMastery:
		return a.masteryCheck(s.Stage)
	case SuggestTypeIn:
		_, ti, _ := classicFor(s.Stage)
		return a.typeInLab(s.Stage, ti)
	case SuggestFocus:
		return a.focusTimer()
	case SuggestWatch:
		return a.techWatch()
	}
	return nil
}

// nextHint is the one-line pointer shown under the dashboard.
func (a *App) nextHint() string {
	a.mu.Lock()
	steps := a.reg.nextSteps(time.Now())
	a.mu.Unlock()
	if len(steps) == 0 {
		return ""
	}
	return sty.Green("➜ Next:") + " " + truncate(steps[0].Title, a.cols()-24) + sty.Gray("  (press n)")
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func (a *App) searchScreen() error {
	a.println(heading("SEARCH · concepts, glossary, resources, classics", sty.Blue, a.cols()))
	q, err := a.con.promptText("Search for", 80, true)
	if err != nil {
		return err
	}
	hits := search(q)
	a.mu.Lock()
	hits = append(a.reg.myResourceHits(q), hits...)
	a.mu.Unlock()
	if len(hits) == 0 {
		a.con.note("Nothing matches %q. Try fewer or shorter words.", q)
		return nil
	}
	shown := hits
	if len(shown) > 15 {
		shown = shown[:15]
	}
	a.printf("\n  %s\n", sty.Gray(fmt.Sprintf("%d match(es)%s", len(hits), map[bool]string{true: ", showing the best 15", false: ""}[len(hits) > 15])))
	for i, h := range shown {
		kind := map[string]func(string) string{HitConcept: sty.Blue, HitGlossary: sty.Cyan, HitResource: sty.Magenta, HitClassic: sty.Yellow, HitBlueprint: sty.Green, HitWord: sty.Cyan, HitMine: sty.Yellow}[h.Kind]
		head := fmt.Sprintf("%s %s %s", kind(padRight(h.Kind, 9)), sty.Gray(fmt.Sprintf("Stage %-2d", h.Stage)), sty.Bold(h.Title))
		a.printf("  %s %s\n", sty.Cyan(fmt.Sprintf("%2d.", i+1)), truncate(stripANSI(head), a.cols()-7))
		if h.Snippet != "" {
			a.printf("      %s\n", sty.Gray(truncate(h.Snippet, a.cols()-8)))
		}
	}
	a.println("")
	n, err := a.con.promptInt("Open which one?", func(n int) error {
		if n < 1 || n > len(shown) {
			return fmt.Errorf("choose 1 to %d", len(shown))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return a.openHit(shown[n-1])
}

// openHit shows a search result in the right place.
func (a *App) openHit(h Hit) error {
	g, _ := guideFor(h.Stage)
	switch h.Kind {
	case HitConcept:
		return a.studyConceptAt(h.Stage, g, h.Concept)
	case HitGlossary:
		for _, t := range g.Glossary {
			if t.Word == h.Title {
				a.printf("\n  %s %s\n", sty.Bold(sty.Cyan(t.Word)), sty.Gray(fmt.Sprintf("(Stage %d glossary)", h.Stage)))
				for _, l := range wrap(t.Meaning, a.cols()-6, "    ") {
					a.println(l)
				}
			}
		}
	case HitResource:
		for _, r := range g.Resources {
			if r.Title == h.Title {
				a.printf("\n  %s %s\n", sty.Magenta(r.Kind), sty.Bold(r.Title))
				if r.URL != "" {
					a.printf("    %s\n", sty.Under(sty.Cyan(r.URL)))
				}
				for _, l := range wrap(r.Note, a.cols()-6, "    ") {
					a.println(sty.Gray(l))
				}
			}
		}
	case HitClassic:
		return a.classicCorner(h.Stage)
	case HitMine:
		a.mu.Lock()
		var found *MyResource
		for i := range a.reg.MyResources {
			if a.reg.MyResources[i].ID == h.Concept {
				m := a.reg.MyResources[i]
				found = &m
			}
		}
		a.mu.Unlock()
		if found != nil {
			return a.editResourceFlow(*found)
		}
	case HitWord:
		if w, ok := lookupWord(h.Title); ok {
			return a.showWord(w)
		}
	case HitBlueprint:
		return a.showBlueprints(h.Stage, g)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Focus timer
// ---------------------------------------------------------------------------

// focusTimer runs a Pomodoro-style study session and logs it.
func (a *App) focusTimer() error {
	a.println(heading("FOCUS TIMER · one task, no distractions", sty.Yellow, a.cols()))
	a.mu.Lock()
	cur := a.reg.currentStage()
	a.mu.Unlock()
	def := 0
	if cur != nil {
		def = cur.ID
	}

	label := "Stage to log it under (0 = general"
	if def > 0 {
		label += fmt.Sprintf(", Enter = Stage %d", def)
	}
	label += ")"
	stage := def
	for {
		s, err := a.con.readLine(promptLabel(label, "q cancel"))
		if err != nil {
			return err
		}
		if isCancel(s) {
			return errCancel
		}
		if s == "" {
			break
		}
		n, convErr := strconv.Atoi(s)
		a.mu.Lock()
		_, st := a.reg.findStage(n)
		a.mu.Unlock()
		if convErr == nil && (n == 0 || st != nil) {
			stage = n
			break
		}
		a.con.warn("Enter a stage number from 1 to 16, or 0 for general study.")
	}

	minutes := 25
	for {
		s, err := a.con.readLine(promptLabel("Minutes (Enter = 25)", "q cancel"))
		if err != nil {
			return err
		}
		if isCancel(s) {
			return errCancel
		}
		if s == "" {
			break
		}
		if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 180 {
			minutes = n
			break
		}
		a.con.warn("Choose between 1 and 180 minutes.")
	}
	goal, err := a.con.promptText("Goal for this session (Enter to skip)", maxLogNoteLen, false)
	if err != nil {
		return err
	}

	start := time.Now()
	var worked time.Duration
	if a.con.raw && a.con.screen {
		worked, err = a.runLiveTimer(time.Duration(minutes)*time.Minute, goal)
		if err != nil {
			return err
		}
	} else {
		a.con.note("Started at %s; planned end %s. Press Enter when you finish.", start.Format("15:04"), start.Add(time.Duration(minutes)*time.Minute).Format("15:04"))
		if _, err := a.con.readLine(promptLabel("Finished?", "Enter")); err != nil {
			return err
		}
		worked = min(time.Since(start), 4*time.Hour)
	}

	logged := math.Round(worked.Minutes()*10) / 10
	if logged < 1 {
		a.con.note("Less than a minute, so nothing was logged.")
		return nil
	}
	note, err := a.con.promptText("What did you get done? (Enter to skip)", maxLogNoteLen, false)
	if err != nil && !errors.Is(err, errCancel) {
		return err
	}
	if note == "" {
		note = goal
	}
	a.mutate(func(r *Registry) {
		r.StudySessions = append(r.StudySessions, StudySession{Start: start.UTC().Truncate(time.Second), Minutes: logged, Stage: stage, Note: note})
	})
	where := "general study"
	if stage > 0 {
		where = fmt.Sprintf("Stage %d", stage)
	}
	a.mu.Lock()
	today := a.reg.dailyMinutes(time.Now(), 1)[0]
	a.mu.Unlock()
	a.con.ok("Logged %.0f minute(s) to %s. Study time today: %.0f minute(s).", logged, where, today)
	return nil
}

// runLiveTimer shows a live countdown and returns the time actually spent
// working (pauses excluded). Keys: p pause/resume, s stop early, Ctrl+L
// redraw; Enter logs the session once time is up.
func (a *App) runLiveTimer(total time.Duration, goal string) (time.Duration, error) {
	restore, err := enableRawInput(int(os.Stdin.Fd()))
	if err != nil {
		return 0, err
	}
	termMu.Lock()
	termRestore = restore
	termMu.Unlock()
	defer restoreTerminal()

	// One reader goroutine; it stops after delivering a key that ends the
	// session, so it never steals input from the next prompt.
	keys := make(chan rune)
	go func() {
		defer close(keys)
		for {
			r, _, err := a.con.in.ReadRune()
			if err != nil {
				return
			}
			keys <- r
			if r == '\r' || r == '\n' || r == 's' || r == 'S' || r == 'q' || r == 'Q' {
				return
			}
		}
	}()

	var worked time.Duration
	paused, done := false, false
	last := time.Now()
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()

	render := func() {
		left := max(0, total-worked).Round(time.Second)
		clock := fmt.Sprintf("%02d:%02d", int(left.Minutes()), int(left.Seconds())%60)
		var line string
		switch {
		case done:
			line = sty.Bold(sty.Green("✓ Time's up!")) + " Press Enter to log it."
		case paused:
			line = sty.Yellow("⏸ paused ") + sty.Bold(clock) + sty.Gray(" · p resume · s stop")
		default:
			line = sty.Yellow("⏱ ") + sty.Bold(clock) + sty.Gray(" left · p pause · s stop")
			if goal != "" {
				line += sty.Gray(" · ") + goal
			}
		}
		fmt.Fprint(a.con.out, "\r\x1b[K  "+truncateStyled(line, a.cols()-3))
	}

	a.println("")
	render()
	for {
		select {
		case <-tick.C:
			now := time.Now()
			if !paused && !done {
				worked += now.Sub(last)
				if worked >= total {
					worked, done = total, true
					fmt.Fprint(a.con.out, "\a") // terminal bell
				}
			}
			last = now
			render()
		case r, ok := <-keys:
			if !ok {
				fmt.Fprintln(a.con.out)
				return worked, nil
			}
			now := time.Now()
			if !paused && !done {
				worked += now.Sub(last)
			}
			last = now
			switch r {
			case 'p', 'P', ' ':
				if !done {
					paused = !paused
				}
			case 's', 'S', 'q', 'Q', '\r', '\n':
				fmt.Fprintln(a.con.out)
				if !done && r != 's' && r != 'S' && r != 'q' && r != 'Q' {
					// Enter before time is up: finish early.
					a.con.note("Finished early.")
				}
				return min(worked, total), nil
			case keyCtrlL:
				fmt.Fprint(a.con.out, clearSeq)
			}
			render()
		}
	}
}

// truncateStyled cuts a styled line to n visible columns, dropping styles
// when it has to cut.
func truncateStyled(s string, n int) string {
	if visibleLen(s) <= n {
		return s
	}
	return truncate(stripANSI(s), n)
}

// ---------------------------------------------------------------------------
// Progress report
// ---------------------------------------------------------------------------

// heatCell renders one day of the activity calendar.
func heatCell(minutes float64) string {
	switch {
	case minutes <= 0:
		return sty.Gray("·")
	case minutes < 15:
		return sty.Green("░")
	case minutes < 45:
		return sty.Green("▒")
	case minutes < 90:
		return sty.Green("▓")
	}
	return sty.Bold(sty.Green("█"))
}

func (a *App) progressReport() {
	w := a.cols()
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()

	a.println(heading("PROGRESS REPORT", sty.Green, w))

	// Activity calendar: one column per week (Monday first), newest right.
	weeks := max(4, min(26, (w-8)/2))
	weekday := (int(now.Weekday()) + 6) % 7 // Monday = 0
	days := (weeks-1)*7 + weekday + 1
	minutes := a.reg.dailyMinutes(now, days)
	a.printf("\n  %s\n", sty.Bold(fmt.Sprintf("Activity, last %d weeks", weeks)))
	names := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	for row := 0; row < 7; row++ {
		label := "  "
		if row%2 == 0 {
			label = names[row]
		}
		line := "  " + sty.Gray(label) + " "
		for col := 0; col < weeks; col++ {
			i := col*7 + row
			if i >= days {
				line += "  " // future days this week
				continue
			}
			line += heatCell(minutes[i]) + " "
		}
		a.println(line)
	}
	a.printf("  %s %s %s %s %s %s %s\n", sty.Gray("less"), heatCell(0), heatCell(5), heatCell(20), heatCell(60), heatCell(120), sty.Gray("more (minutes/day)"))

	// This week against last week.
	ws := a.reg.weekSummary(now)
	trend := func(cur, prev float64) string {
		switch {
		case cur > prev:
			return sty.Green("▲")
		case cur < prev:
			return sty.Red("▼")
		}
		return sty.Gray("=")
	}
	a.printf("\n  %s\n", sty.Bold("Last 7 days vs the 7 before"))
	row := func(label, cur, prev string, c, p float64) {
		if w < 50 { // narrow: drop the unit from the previous value
			prev = strings.TrimSuffix(prev, " min")
		}
		a.printf("    %s %s %s %s\n", padRight(label, 15), sty.Bold(padRight(cur, 9)), trend(c, p), sty.Gray("was "+prev))
	}
	row("Study time", fmt.Sprintf("%.0f min", ws.minutes), fmt.Sprintf("%.0f min", ws.prevMinutes), ws.minutes, ws.prevMinutes)
	row("Active days", fmt.Sprint(ws.activeDays), fmt.Sprint(ws.prevDays), float64(ws.activeDays), float64(ws.prevDays))
	row("Cards reviewed", fmt.Sprint(ws.reviews), fmt.Sprint(ws.prevReviews), float64(ws.reviews), float64(ws.prevReviews))
	row("Focus sessions", fmt.Sprint(ws.sessions), fmt.Sprint(ws.prevSess), float64(ws.sessions), float64(ws.prevSess))

	// Time and mastery per track.
	a.printf("\n  %s\n", sty.Bold("By track"))
	for ti := range a.reg.Tracks {
		t := &a.reg.Tracks[ti]
		var hours float64
		points, maxPoints := 0, 0
		for si := range t.Stages {
			s := &t.Stages[si]
			hours += stageHours(s)
			p, m, _, _ := a.reg.stageMastery(s)
			points, maxPoints = points+p, maxPoints+m
		}
		for _, ss := range a.reg.StudySessions {
			if tr, _ := a.reg.findStage(ss.Stage); tr != nil && tr.ID == t.ID {
				hours += ss.Minutes / 60
			}
		}
		color := trackColor(t.ID)
		for _, l := range flow([]string{
			color(sty.Bold("Track " + t.ID)),
			sty.Gray("mastery") + " " + bar(points, maxPoints, max(6, min(16, w-40)), color) + " " + pct(points, maxPoints),
			sty.Gray("time") + " " + sty.Bold(fmt.Sprintf("%.1fh", hours)),
		}, "  ", w, "    ") {
			a.println(l)
		}
	}

	// Memory health of the review deck.
	cards := a.reg.unlockedCards()
	mature, young, lapses := 0, 0, 0
	for _, c := range cards {
		st := a.reg.Reviews[c.ID]
		switch {
		case st.IntervalDays >= masteredDays:
			mature++
		case st.Reps > 0:
			young++
		}
		lapses += st.Lapses
	}
	a.printf("\n  %s\n", sty.Bold("Review deck"))
	for _, l := range flow([]string{
		fmt.Sprintf("%d card(s)", len(cards)),
		sty.Green(fmt.Sprintf("%d mature (21+ days)", mature)),
		sty.Cyan(fmt.Sprintf("%d learning", young)),
		sty.Gray(fmt.Sprintf("%d new", len(cards)-mature-young)),
		sty.Yellow(fmt.Sprintf("%d lapse(s)", lapses)),
	}, sty.Gray(" · "), w, "    ") {
		a.println(l)
	}
	if hard := a.reg.hardestCards(5); len(hard) > 0 {
		a.printf("  %s\n", sty.Bold("Most often forgotten: worth re-reading"))
		for _, c := range hard {
			name := c.Concept
			if name == "" {
				name = fmt.Sprintf("Stage %d quiz", c.StageID)
			}
			a.printf("    %s %s %s\n", sty.Yellow(fmt.Sprintf("%d×", a.reg.Reviews[c.ID].Lapses)), truncate(name, w-24), sty.Gray("· "+c.Kind))
		}
	}

	// Goals and achievements.
	if g := a.reg.Goals; g != (Goals{}) {
		wp := a.reg.weekToDate(now)
		a.printf("\n  %s %s\n", sty.Bold("This week's goals"), sty.Gray(fmt.Sprintf("(day %d of 7)", wp.dayOf)))
		goal := func(label string, done float64, goal int) {
			if goal <= 0 {
				return
			}
			state := sty.Yellow("behind")
			switch {
			case done >= float64(goal):
				state = sty.Green("✓ done")
			case onTrack(done, goal, wp.dayOf):
				state = sty.Cyan("on track")
			}
			a.printf("    %s %s %s\n", padRight(label, 15), sty.Bold(padRight(fmt.Sprintf("%.0f/%d", done, goal), 11)), state)
		}
		goal("Study minutes", wp.minutes, g.Minutes)
		goal("Study days", float64(wp.days), g.Days)
		goal("Review cards", float64(wp.cards), g.Cards)
	}
	a.printf("\n  %s %s %s\n", sty.Bold("Achievements"), sty.Yellow(fmt.Sprintf("%d of %d", len(a.reg.Achievements), len(achievements))), sty.Gray("(press a)"))
	a.println("")
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

func (a *App) exportNotes() error {
	a.mu.Lock()
	md := a.reg.exportMarkdown(time.Now())
	a.mu.Unlock()
	path := filepath.Join(filepath.Dir(a.path), "academy_notes.md")
	if err := atomicWriteFile(path, []byte(md)); err != nil {
		return fmt.Errorf("export: %w", err)
	}
	lines := strings.Count(md, "\n")
	a.con.ok("Exported %d lines of notes to %s", lines, path)
	a.con.note("It is plain Markdown: open it in any editor, Obsidian, Notion, or put it in Git.")
	return nil
}
