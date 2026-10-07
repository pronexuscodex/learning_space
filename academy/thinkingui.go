package main

// Screens for thinking for yourself: the hint ladder, the honest solve
// report, and blank-page recall. The data and rules are in thinking.go.

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// hint-ladder rungs: each gives away a little more of the thinking.
var rungNames = []string{"", "Questions to think with", "Where to look it up", "The hint"}

// hintLadder offers help one rung at a time and returns the highest rung
// climbed. Rung 1 asks the questions an engineer asks; rung 2 points to
// where the answer can be found; only rung 3 gives the hint. In Classic
// Mode the hint stays behind the struggle clock.
func (a *App) hintLadder(stageID int, c Concept, e Exercise, key string, opened time.Time, classic bool) (int, error) {
	rung := 0
	a.println("")
	for _, l := range wrap("Stuck? Climb one rung at a time: questions first, then where to look, and only then the hint. The less help you take, the more of it becomes yours.", a.cols()-4, "  ") {
		a.println(sty.Gray(l))
	}
	for rung < 3 {
		next := rung + 1
		opts := []string{
			"Back to work: I'll keep thinking",
			fmt.Sprintf("Rung %d · %s", next, rungNames[next]),
		}
		choice, err := a.con.promptChoice("Your move", opts)
		if err != nil {
			return rung, err
		}
		if choice == 0 {
			return rung, nil
		}
		if next == 3 && classic {
			show, err := a.hintGate(stageID, key, opened, e.Level)
			if err != nil {
				return rung, err
			}
			if !show {
				continue
			}
		}
		rung = next
		a.showRung(stageID, c, e, rung)
	}
	return rung, nil
}

// showRung prints one rung of the hint ladder.
func (a *App) showRung(stageID int, c Concept, e Exercise, rung int) {
	w := a.cols()
	a.println("")
	title := fmt.Sprintf("Rung %d · %s", rung, rungNames[rung])
	bullet := func(text string) {
		for i, l := range wrap(text, w-8, "") {
			lead := "    "
			if i == 0 {
				lead = "  " + sty.Yellow("•") + " "
			}
			a.println(lead + l)
		}
	}
	switch rung {
	case 1:
		a.printf("  %s\n", sty.Bold(sty.Yellow("🧭 "+title)))
		bullet("What exactly is asked? Say it in one sentence, out loud or on paper.")
		bullet("What do you know, what is unknown, and what would a correct answer look like?")
		if c.MentalModel != "" {
			bullet(fmt.Sprintf("Which idea is this about? The mental model of “%s”: %s", c.Name, c.MentalModel))
		}
		bullet("Can you solve a smaller version first: the smallest input, one case, one line?")
		bullet("What could you print, draw or measure to see what is really happening?")
		bullet("Explain your attempt, step by step, to a rubber duck. Where do you hesitate? That is where the gap is.")
	case 2:
		a.printf("  %s\n", sty.Bold(sty.Yellow("🔎 "+title)))
		where := fmt.Sprintf("Your concept card: “%s”", c.Name)
		if c.UnderTheHood != "" {
			where += ", especially Under the hood"
		}
		bullet(where + ".")
		if g, ok := guideFor(stageID); ok {
			for _, t := range termsIn(g.Glossary, e.Task) {
				bullet(sty.Bold(t.Word) + ": " + t.Meaning)
			}
		}
		if l, ok := conceptLinks[c.Name]; ok {
			if l.Read != "" {
				bullet("Go deeper: " + l.Read)
			}
			if len(l.Related) > 0 {
				bullet("Related ideas you may have studied: " + strings.Join(l.Related, ", ") + ".")
			}
		}
		bullet("Primary sources: man pages (man 3 printf, man 2 open), the language or library reference, the RFC or standard. Look up the concept, not the answer to this exercise.")
	case 3:
		for i, l := range wrap(e.Hint, w-14, "") {
			lead := "           "
			if i == 0 {
				lead = sty.Bold(sty.Yellow("💡 Hint")) + "    "
			}
			a.printf("  %s%s\n", lead, sty.Yellow(l))
		}
	}
}

// hintGate applies the Classic Mode struggle clock to the hint and reports
// whether it may be shown now.
func (a *App) hintGate(stageID int, key string, opened time.Time, level string) (bool, error) {
	a.mu.Lock()
	stuck := a.reg.stuckLogged(key)
	a.mu.Unlock()
	ok, left := hintUnlocked(opened, level, stuck, time.Now())
	if ok {
		return true, nil
	}
	a.con.say(sty.Yellow("🔒"), fmt.Sprintf("The hint unlocks in %d more minute(s). Productive struggle is where the learning happens.", int(left.Minutes())+1), sty.Yellow)
	choice, err := a.con.promptChoice("What now?", []string{
		"Keep working (come back later)",
		"Log what I have tried so far",
		"I'm truly stuck: write down what I tried and unlock the hint now",
	})
	if err != nil {
		return false, err
	}
	switch choice {
	case 1:
		text, err := a.con.promptText("What have you tried?", maxNotesLen, true)
		if err != nil {
			return false, err
		}
		a.addNote(stageID, key, NoteTried, text)
		a.con.ok("Logged. Keep going; the clock is still running.")
	case 2:
		for {
			text, err := a.con.promptText("What did you try, and where exactly are you stuck?", maxNotesLen, true)
			if err != nil {
				return false, err
			}
			if utf8.RuneCountInString(text) < 20 {
				a.con.warn("Say a little more (at least a sentence). Describing the problem often solves it.")
				continue
			}
			a.addNote(stageID, key, NoteStuck, text)
			return true, nil
		}
	}
	return false, nil
}

// recordSolve asks how the exercise was solved and records it. A solve that
// leaned on the hint or on someone else comes back later, to be redone from
// a blank page.
func (a *App) recordSolve(stageID int, c Concept, i, rung int) error {
	opts := make([]string, len(solvedOptions))
	for j, o := range solvedOptions {
		opts[j] = o.Label
	}
	a.con.note("Be honest: this is only for you, and it decides what comes back for practice.")
	choice, err := a.con.promptChoice("How did you get there?", opts)
	if errors.Is(err, errCancel) {
		return nil // the exercise stays done; nothing recorded
	}
	if err != nil {
		return err
	}
	how := solvedOptions[choice].How
	if rung >= 3 && how == SolvedAlone {
		how = SolvedHint // the hint was read, so it was not entirely alone
		a.con.note("You opened the hint, so this one counts as solved with the hint.")
	}
	now := time.Now()
	rec := SolveRecord{Key: exerciseKey(c.Name, i), Stage: stageID, Concept: c.Name, Index: i, At: now, How: how, Rung: rung}
	a.mutate(func(r *Registry) { r.Solves = append(r.Solves, rec) })
	switch how {
	case SolvedAlone:
		a.con.ok("Solved on your own. That is the real thing.")
	case SolvedDocs:
		a.con.ok("Finding it in the docs is exactly what engineers do all day.")
	default:
		due := now.Add(redoAfter)
		a.con.say(sty.Cyan("↻"), fmt.Sprintf("Close the solution now. On %s this exercise comes back: redo it from a blank page, without help. Doing it alone is what makes it yours.", due.Format("Monday 2 January")), sty.Cyan)
	}
	return nil
}

// readParagraph reads lines until an empty one, up to maxLen characters.
func (a *App) readParagraph(label string, maxLen int) (string, error) {
	a.printf("  %s\n", sty.Gray(label))
	var lines []string
	total := 0
	for {
		l, err := a.con.readLine(sty.Cyan("  │ "))
		if err != nil {
			return strings.Join(lines, "\n"), err
		}
		if l == "" {
			break
		}
		if strings.EqualFold(l, ":q") {
			return "", errCancel
		}
		total += utf8.RuneCountInString(l) + 1
		if total > maxLen {
			a.con.warn("That is the limit (%d characters); finishing here.", maxLen)
			break
		}
		lines = append(lines, l)
	}
	return strings.Join(lines, "\n"), nil
}

// freeRecall runs a blank-page recall of one concept: write everything you
// remember, then compare with the card and score yourself.
func (a *App) freeRecall(stageID int, ci int) error {
	g, ok := guideFor(stageID)
	if !ok || ci < 0 || ci >= len(g.Concepts) {
		return nil
	}
	c := g.Concepts[ci]
	w := a.cols()
	a.println(heading("BLANK-PAGE RECALL", sty.Magenta, w))
	a.printf("\n  %s %s\n", sty.Bold(c.Name), sty.Gray(fmt.Sprintf("· Stage %d", stageID)))
	for _, l := range wrap("Without looking anything up, write everything you remember: what it is, why it matters, how it works underneath, and an example. Pulling it out of your own head is what makes memory last; rereading only feels like learning.", w-4, "  ") {
		a.println(l)
	}
	text, err := a.readParagraph("Write as many lines as you like; an empty line finishes. :q cancels.", maxRecallLen)
	if err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		a.con.note("Nothing written. Even a few words count: try again when you are ready.")
		return nil
	}

	a.println("")
	a.printf("  %s\n", sty.Bold(sty.Cyan("Now compare with the card")))
	check := func(label, body string) {
		if body == "" {
			return
		}
		a.printf("  %s\n", sty.Cyan(label))
		for _, l := range wrap(body, w-6, "    ") {
			a.println(l)
		}
	}
	check("In one line", c.Summary)
	check("Mental model", c.MentalModel)
	if terms := termsIn(g.Glossary, c.Body); len(terms) > 0 {
		words := make([]string, len(terms))
		for j, t := range terms {
			words[j] = t.Word
		}
		check("Key words: did you use them?", strings.Join(words, ", "))
	}
	choice, err := a.con.promptChoice("How much did you get?", []string{
		sty.Red("Hardly anything") + sty.Gray(" · re-read the card today, recall again tomorrow"),
		sty.Yellow("Some of it") + sty.Gray(" · again in 3 days"),
		sty.Green("Most of it") + sty.Gray(" · again in 10 days"),
		sty.Cyan("All of it") + sty.Gray(" · again in a month"),
	})
	if errors.Is(err, errCancel) {
		return nil
	}
	if err != nil {
		return err
	}
	score := choice + 1
	a.mutate(func(r *Registry) {
		r.Recalls = append(r.Recalls, RecallRecord{Stage: stageID, Concept: c.Name, At: time.Now(), Score: score, Text: text})
	})
	if score <= RecallSome {
		a.con.note("That's useful to know. Re-read the card now (Study Hall [6]); tomorrow's recall will go better.")
	} else {
		a.con.ok("Recall saved: %s. Retrieval like this is the strongest study habit there is.", recallScoreNames[score])
	}
	return nil
}

// thinkingReport is the Progress report's "Thinking for yourself" section.
// The caller holds a.mu.
func (a *App) thinkingReport(w int, now time.Time) {
	a.printf("\n  %s\n", sty.Bold("Thinking for yourself"))
	line := func(label, value, note string) {
		out := "    " + padRight(label, 17) + value
		if note != "" {
			out += " " + sty.Gray(note)
		}
		if visibleLen(out) <= w {
			a.println(out)
			return
		}
		// Narrow screen: the label on its own line, the rest wrapped below.
		a.println("    " + label)
		for _, l := range flow(append([]string{value}, strings.Fields(sty.Gray(note))...), " ", w, "      ") {
			a.println(l)
		}
	}
	sr := a.reg.selfReliance()
	if t := sr.total(); t > 0 {
		line("Solved yourself", sty.Bold(fmt.Sprintf("%d of %d", sr.own(), t))+" "+bar(sr.own(), t, max(6, min(16, w-40)), sty.Green), pct(sr.own(), t))
		line("How", fmt.Sprintf("%d alone · %d docs · %d hint · %d help", sr.Alone, sr.Docs, sr.Hint, sr.Help), "")
	} else {
		line("Solved yourself", sty.Gray("no exercises reported yet"), "")
	}
	if due, all := len(a.reg.redos(now, true)), len(a.reg.redos(now, false)); all > 0 {
		line("Redo list", sty.Bold(fmt.Sprintf("%d due", due)), fmt.Sprintf("of %d waiting to be redone alone", all))
	}
	if n := len(a.reg.Recalls); n > 0 {
		sum := 0
		for _, r := range a.reg.Recalls {
			sum += r.Score
		}
		avg := float64(sum) / float64(n)
		line("Recalls", sty.Bold(fmt.Sprint(n)), fmt.Sprintf("average: %s", recallScoreNames[int(avg+0.5)]))
	}
	if c := a.reg.Calibration; c.answers() > 0 {
		var parts []string
		for lvl, name := range confNames {
			if acc, ok := c.accuracy(lvl); ok {
				parts = append(parts, fmt.Sprintf("%s %s", name, percent(acc)))
			}
		}
		value := sty.Gray(fmt.Sprintf("%d rated answers; 5 per level needed", c.answers()))
		if len(parts) > 0 {
			value = strings.Join(parts, " · ")
		}
		line("Calibration", value, "")
		if v := c.verdict(); v != "" {
			for _, l := range wrap(v, w-8, "      ") {
				a.println(sty.Cyan(l))
			}
		}
	}
	if n := a.reg.workoutsCompleted(); n > 0 || len(a.reg.Workouts) > 0 {
		line("Workouts", sty.Bold(fmt.Sprint(n)), fmt.Sprintf("completed · %d day(s) you thought first", a.reg.thinkFirstDays()))
	}
}
