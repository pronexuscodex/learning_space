package main

// Screens for weekly goals, the roadmap and achievements.

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Weekly goals
// ---------------------------------------------------------------------------

// goalsLine is the dashboard's weekly-goal tracker ("" when no goals).
func (a *App) goalsLine() []string {
	a.mu.Lock()
	g := a.reg.Goals
	wp := a.reg.weekToDate(time.Now())
	a.mu.Unlock()
	if g == (Goals{}) {
		return nil
	}
	item := func(label string, done float64, goal int, unit string) string {
		if goal <= 0 {
			return ""
		}
		color := sty.Yellow
		switch {
		case done >= float64(goal):
			color = sty.Green
		case onTrack(done, goal, wp.dayOf):
			color = sty.Cyan
		}
		text := sty.Gray(label) + " " + bar(int(done), goal, 6, color) + " " + sty.Bold(fmt.Sprintf("%.0f/%d", done, goal)) + sty.Gray(unit)
		if done >= float64(goal) {
			text += sty.Green(" ✓")
		}
		return text
	}
	return flow([]string{
		sty.Bold("This week"),
		item("time", wp.minutes, g.Minutes, " min"),
		item("days", float64(wp.days), g.Days, ""),
		item("cards", float64(wp.cards), g.Cards, ""),
	}, "   ", a.cols(), "  ")
}

// setGoals explains weekly goals and lets the learner change them.
func (a *App) setGoals() error {
	a.println(heading("WEEKLY GOALS · small, steady, every week", sty.Green, a.cols()))
	a.println("")
	intro := `Weekly goals turn good intentions into a plan you can see. The
dashboard shows how this week (Monday to Sunday) is going: cyan means on
track for the days gone so far, yellow means behind, green means done.

- Aim small. 150 minutes a week is about 25 minutes on most days, and that is enough to finish the whole academy steadily.
- Days matter more than hours. Five short days beat one long day, because spacing is what makes memories last.
- Cards count reviews from the Daily Review. About ten a day keeps the deck healthy.`
	for _, l := range wrap(intro, a.cols()-4, "  ") {
		a.println(l)
	}

	a.mu.Lock()
	cur := a.reg.Goals
	a.mu.Unlock()
	if cur == (Goals{}) {
		cur = defaultGoals
		a.println("")
		a.con.note("No goals yet. Press Enter to accept each suggestion, or type your own (0 = no goal).")
	} else {
		a.println("")
		a.con.note("Press Enter to keep a goal, or type a new number (0 = no goal).")
	}
	ask := func(label string, current, limit int) (int, error) {
		for {
			s, err := a.con.readLine(promptLabel(fmt.Sprintf("%s (Enter = %d)", label, current), "q cancel"))
			if err != nil {
				return 0, err
			}
			if isCancel(s) {
				return 0, errCancel
			}
			if s == "" {
				return current, nil
			}
			if n, err := strconv.Atoi(s); err == nil && n >= 0 && n <= limit {
				return n, nil
			}
			a.con.warn("Enter a whole number from 0 to %d.", limit)
		}
	}
	var g Goals
	var err error
	if g.Minutes, err = ask("Study minutes per week", cur.Minutes, 7*24*60); err != nil {
		return err
	}
	if g.Days, err = ask("Study days per week", cur.Days, 7); err != nil {
		return err
	}
	if g.Cards, err = ask("Review cards per week", cur.Cards, 10000); err != nil {
		return err
	}
	a.mutate(func(r *Registry) { r.Goals = g })
	if g == (Goals{}) {
		a.con.ok("Goals cleared.")
	} else {
		a.con.ok("Goals saved. They appear on the dashboard; commit to keep them.")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Roadmap
// ---------------------------------------------------------------------------

func stateStyle(state int) func(string) string {
	return []func(string) string{sty.Gray, sty.Cyan, sty.Yellow, sty.Blue, sty.Green}[state]
}

// renderRoadmap draws every stage by track with its state, concept
// progress and missing prerequisites.
func (a *App) renderRoadmap() {
	a.mu.Lock()
	defer a.mu.Unlock()
	w := a.cols()
	cur := a.reg.currentStage()

	a.println(heading("ROADMAP · where you are, what opens next", sty.Cyan, w))

	// The right-hand label: "you are here" or the missing prerequisites.
	tails := map[int]string{}
	tailWidth := 0
	for _, s := range a.reg.stageOrder() {
		switch {
		case cur != nil && cur.ID == s.ID:
			tails[s.ID] = sty.Bold(sty.Yellow("➜ you are here"))
		case a.reg.stageState(s) == StageLocked:
			var need []string
			for _, p := range a.reg.missingPrereqs(s.ID) {
				need = append(need, strconv.Itoa(p))
			}
			tails[s.ID] = sty.Gray("needs " + strings.Join(need, ", "))
		}
		tailWidth = max(tailWidth, visibleLen(tails[s.ID]))
	}
	// "  ★ 16  " + title + " " + bar(8) + " " + "3/5" (5 columns), then "  " + label.
	room := w - 8 - 15
	if room-tailWidth-2 >= 14 {
		room -= tailWidth + 2
	} else {
		tails = nil // too narrow: the glyphs and legend carry the state
	}

	for ti := range a.reg.Tracks {
		t := &a.reg.Tracks[ti]
		a.printf("\n  %s\n", trackColor(t.ID)(sty.Bold(truncate(fmt.Sprintf("Track %s · %s", t.ID, t.Name), w-2))))
		for si := range t.Stages {
			s := &t.Stages[si]
			state := a.reg.stageState(s)
			color := stateStyle(state)
			studied, total := conceptProgress(s)
			line := "  " + color(stageStateGlyphs[state]) + " " + sty.Bold(fmt.Sprintf("%2d", s.ID)) + "  " +
				color(padRight(truncate(s.Title, room), room)) + " " + bar(studied, total, 8, color) + " " +
				sty.Gray(fmt.Sprintf("%d/%d", studied, total))
			if tail := tails[s.ID]; tail != "" {
				line = padRight(line, w-tailWidth) + tail
			}
			a.println(line)
		}
	}
	a.println("")
	var legend []string
	for i, name := range stageStateNames {
		legend = append(legend, stateStyle(i)(stageStateGlyphs[i])+" "+sty.Gray(name))
	}
	for _, l := range flow(legend, "   ", w, "  ") {
		a.println(l)
	}
	for _, l := range wrap("A stage opens once half of each stage it builds on is understood. Locked stages can still be explored; they will simply make more sense later.", w-4, "  ") {
		a.println(sty.Gray(l))
	}
}

// roadmap shows the map and opens a chosen stage's Study Hall.
func (a *App) roadmap() error {
	a.paged(false, a.renderRoadmap)
	a.println("")
	id, err := a.con.promptInt("Open which stage in the Study Hall?", func(id int) error {
		a.mu.Lock()
		defer a.mu.Unlock()
		if _, s := a.reg.findStage(id); s == nil {
			return fmt.Errorf("there is no stage %d", id)
		}
		return nil
	})
	if err != nil {
		return err
	}
	a.mu.Lock()
	missing := a.reg.missingPrereqs(id)
	a.mu.Unlock()
	if len(missing) > 0 {
		var need []string
		for _, p := range missing {
			need = append(need, strconv.Itoa(p))
		}
		a.con.note("Stage %d builds on stage(s) %s. Explore freely, but those will make it easier.", id, strings.Join(need, ", "))
	}
	return a.studyHallFor(id)
}

// ---------------------------------------------------------------------------
// Achievements
// ---------------------------------------------------------------------------

// announceAchievements awards newly earned achievements and celebrates them.
func (a *App) announceAchievements() {
	a.mu.Lock()
	fresh := a.reg.awardAchievements(time.Now())
	if len(fresh) > 0 {
		a.dirty = true
	}
	a.mu.Unlock()
	for _, ach := range fresh {
		a.con.ok("%s %s: %s", sty.Bold(sty.Yellow("🏆 Achievement unlocked!")), sty.Bold(ach.Title), ach.Desc)
	}
}

// showAchievements lists every achievement, earned ones first.
func (a *App) showAchievements() {
	a.mu.Lock()
	earned := map[string]time.Time{}
	for k, v := range a.reg.Achievements {
		earned[k] = v
	}
	a.mu.Unlock()
	w := a.cols()

	a.println(heading(fmt.Sprintf("ACHIEVEMENTS · %d of %d", len(earned), len(achievements)), sty.Yellow, w))
	a.println("")
	for pass := 0; pass < 2; pass++ {
		for _, ach := range achievements {
			at, have := earned[ach.ID]
			if have != (pass == 0) {
				continue
			}
			glyph, title := sty.Gray("○"), sty.Gray(ach.Title)
			when := ""
			if have {
				glyph, title = sty.Yellow("★"), sty.Bold(sty.Yellow(ach.Title))
				when = sty.Gray(" · " + at.Local().Format("2006-01-02"))
			}
			a.printf("  %s %s%s\n", glyph, title, when)
			for _, l := range wrap(ach.Desc, w-7, "     ") {
				a.println(sty.Gray(l))
			}
		}
	}
}
