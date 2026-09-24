package main

// Study Hall: per-stage concept explanations, glossary, resource library,
// lab blueprints (one-step enrolment) and self-check quizzes.

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// studyHall picks a stage and loops over its learning menu.
func (a *App) studyHall() error {
	a.println(heading("STUDY HALL · learn, practise, build", sty.Blue, a.cols()))
	stageID, err := a.pickStage(nil)
	if err != nil {
		return err
	}
	return a.studyHallFor(stageID)
}

// studyHallFor opens the Study Hall of one stage.
func (a *App) studyHallFor(stageID int) error {
	guide, ok := guideFor(stageID)
	if !ok {
		a.con.note("No Study Hall material for stage %d yet.", stageID)
		return nil
	}

	for {
		a.renderStageSyllabus(stageID, guide)
		a.mu.Lock()
		_, st := a.reg.findStage(stageID)
		exDone, exTotal := exerciseProgress(st)
		a.mu.Unlock()
		choice, err := a.con.promptChoice("Study Hall", []string{
			"Study a concept",
			fmt.Sprintf("%s %s", sty.Bold("Exercise gym"), sty.Gray(fmt.Sprintf("(%d/%d done · warm-up → practice → real-world)", exDone, exTotal))),
			fmt.Sprintf("Glossary %s", sty.Gray(fmt.Sprintf("(%d words explained simply)", len(guide.Glossary)))),
			fmt.Sprintf("Resource library %s", sty.Gray(fmt.Sprintf("(%d)", len(guide.Resources)))),
			fmt.Sprintf("Lab blueprints %s", sty.Gray(fmt.Sprintf("(%d, enroll with one keystroke)", len(guide.Blueprints)))),
			fmt.Sprintf("Self-check quiz %s", sty.Gray(fmt.Sprintf("(%d questions)", len(guide.Quiz)))),
			fmt.Sprintf("%s %s", sty.Bold("Mastery check"), sty.Gray(fmt.Sprintf("(mixed exam, %d%% to pass)", int(masteryPassMark*100)))),
			fmt.Sprintf("%s %s", sty.Bold("Classic corner"), sty.Gray("(anchor book, classic text, real source, type-in lab)")),
			"Back to main menu",
		})
		if err != nil {
			return err
		}
		switch choice {
		case 0:
			err = a.studyConcept(stageID, guide)
		case 1:
			err = a.exerciseGym(stageID, guide)
		case 2:
			a.paged(true, func() { a.showGlossary(guide) })
		case 3:
			a.paged(true, func() { a.showResources(guide) })
		case 4:
			err = a.showBlueprints(stageID, guide)
		case 5:
			err = a.runQuiz(guide)
		case 6:
			err = a.masteryCheck(stageID)
		case 7:
			err = a.classicCorner(stageID)
		default:
			return nil
		}
		switch {
		case errors.Is(err, errCancel):
			a.con.note("Cancelled.")
		case err != nil:
			return err
		}
	}
}

// renderStageSyllabus prints the stage overview and the concept checklist.
func (a *App) renderStageSyllabus(stageID int, g StageGuide) {
	a.mu.Lock()
	t, s := a.reg.findStage(stageID)
	color := trackColor(t.ID)
	title := s.Title
	level := map[string]int{}
	dots := map[string]string{}
	for _, c := range g.Concepts {
		level[c.Name] = a.reg.conceptMastery(s, c)
		dots[c.Name] = exerciseDots(s, c)
	}
	points, maxPoints, mastered, total := a.reg.stageMastery(s)
	check := s.MasteryCheck
	var prereqs []string
	weak := 0
	for _, pid := range stagePrereqs[stageID] {
		if _, ps := a.reg.findStage(pid); ps != nil {
			prereqs = append(prereqs, fmt.Sprintf("Stage %d (%s)", pid, ps.Title))
			if st, tot := conceptProgress(ps); st*2 < tot {
				weak++
			}
		}
	}
	a.mu.Unlock()

	a.printf("\n  %s %s %s\n", color(sty.Bold(fmt.Sprintf("Stage %d", stageID))), sty.Gray("·"), sty.Bold(truncate(title, a.cols()-14)))
	for _, l := range wrap(g.Overview, a.cols()-4, "  ") {
		a.println(sty.Italic(l))
	}
	if len(prereqs) > 0 {
		a.println("")
		for i, l := range wrap(strings.Join(prereqs, ", "), a.cols()-16, "") {
			lead := "              "
			if i == 0 {
				lead = sty.Bold(sty.Blue("🧱 Builds on")) + "  "
			}
			a.printf("  %s%s\n", lead, sty.Gray(l))
		}
		if weak > 0 {
			for _, l := range wrap("Tip: you have covered less than half of some of these. A quick visit there first will make this stage easier.", a.cols()-16, "") {
				a.printf("  %s%s\n", strings.Repeat(" ", 14), sty.Yellow(l))
			}
		}
	}
	if len(g.Outcomes) > 0 {
		a.printf("\n  %s\n", sty.Bold(sty.Green("After this stage you'll be able to:")))
		for _, o := range g.Outcomes {
			for i, l := range wrap(o, a.cols()-8, "      ") {
				if i == 0 {
					l = "    " + sty.Green("✓") + " " + strings.TrimLeft(l, " ")
				}
				a.println(l)
			}
		}
	}
	checkNote := sty.Gray("mastery check not taken")
	if check != nil {
		if check.Passed {
			checkNote = sty.Green(fmt.Sprintf("mastery check passed %d/%d", check.Score, check.Total))
		} else {
			checkNote = sty.Yellow(fmt.Sprintf("mastery check %d/%d, not yet passed", check.Score, check.Total))
		}
	}
	w := a.cols()
	a.println("")
	for _, l := range flow([]string{
		sty.Gray("Mastery") + " " + bar(points, maxPoints, max(6, min(20, w-30)), sty.Green) + " " + sty.Bold(fmt.Sprintf("%d/%d mastered", mastered, total)),
		checkNote,
	}, "   ", w, "  ") {
		a.println(l)
	}
	legend := make([]string, 0, len(masteryNames)+1)
	for i, n := range masteryNames {
		legend = append(legend, masteryBadge(i)+" "+sty.Gray(n))
	}
	legend = append(legend, sty.Gray("●●● = exercises done"))
	for _, l := range flow(legend, "  ", w, "  ") {
		a.println(l)
	}
	for i, c := range g.Concepts {
		head := fmt.Sprintf("    %s %s %s %s", masteryBadge(level[c.Name]), color(fmt.Sprintf("%d.", i+1)), dots[c.Name], sty.Bold(truncate(c.Name, w-15)))
		if room := w - visibleLen(head) - 3; room >= 16 {
			head += " " + sty.Gray("— "+truncate(c.Summary, room))
		}
		a.println(head)
	}
	a.println("")
}

// studyConcept renders one concept card, then either marks it understood
// (asking the learner to explain it in their own words) or lets them
// refresh their explanation.
func (a *App) studyConcept(stageID int, g StageGuide) error {
	names := make([]string, len(g.Concepts))
	for i, c := range g.Concepts {
		names[i] = c.Name
	}
	idx, err := a.con.promptChoice("Concept", names)
	if err != nil {
		return err
	}
	return a.studyConceptAt(stageID, g, idx)
}

// studyConceptAt opens concept idx of a stage directly (used by the Study
// Hall, What's next and Search).
func (a *App) studyConceptAt(stageID int, g StageGuide, idx int) error {
	c := g.Concepts[idx]
	a.mu.Lock()
	_, s := a.reg.findStage(stageID)
	already := s.hasStudied(c.Name)
	note := s.Notes[c.Name]
	level := a.reg.conceptMastery(s, c)
	done := make([]bool, len(c.Exercises))
	for i := range c.Exercises {
		done[i] = s.hasDone(c.Name, i)
	}
	a.mu.Unlock()
	a.paged(false, func() { a.renderConcept(c, idx+1, len(g.Concepts), g.Glossary, done, note, level) })

	if !already {
		yes, err := a.con.confirm("Do you understand it well enough to explain it to a friend?")
		if err != nil || !yes {
			if err == nil {
				a.con.note("No rush. Re-read the analogy, try the warm-up exercise, then come back.")
			}
			return err
		}
		a.con.note("Now explain it in your own words; putting it into your own words is what makes it stick.")
		text, err := a.con.promptText("Your explanation (Enter to skip)", maxNotesLen, false)
		if err != nil && !errors.Is(err, errCancel) {
			return err
		}
		a.mutate(func(r *Registry) {
			_, s := r.findStage(stageID)
			s.setStudied(c.Name, true)
			if text != "" {
				s.Notes[c.Name] = text
			}
		})
		a.con.ok("%s marked as understood. Its %d review cards join your Daily Review [9].", sty.Bold(c.Name), len(conceptCards(stageID, c)))
		return nil
	}

	choice, err := a.con.promptChoice("Next", []string{
		"Continue",
		"Write or update my own explanation",
		"Mark as needing review (removes its cards from the Daily Review)",
	})
	if err != nil {
		return err
	}
	switch choice {
	case 1:
		text, err := a.con.promptText("Your explanation", maxNotesLen, true)
		if err != nil {
			return err
		}
		a.mutate(func(r *Registry) {
			_, s := r.findStage(stageID)
			s.Notes[c.Name] = text
		})
		a.con.ok("Explanation saved. You will see it when this concept comes up in your reviews.")
	case 2:
		a.mutate(func(r *Registry) {
			_, s := r.findStage(stageID)
			s.setStudied(c.Name, false)
		})
		a.con.ok("%s moved back to review.", sty.Bold(c.Name))
	}
	return nil
}

// renderConcept prints a concept as a heavy-bordered reading card, going
// from intuition (analogy, real life) to precision (details, diagram).
func (a *App) renderConcept(c Concept, n, total int, glossary []Term, done []bool, note string, level int) {
	w := a.cols()
	edge := sty.Blue("┃")
	blank := func() { a.printf("  %s\n", edge) }
	section := func(title string, color func(string) string) {
		blank()
		a.printf("  %s %s\n", edge, sty.Bold(color(title)))
	}
	paragraph := func(text string, style func(string) string) {
		for _, l := range wrap(text, w-6, "") {
			a.printf("  %s %s\n", edge, style(l))
		}
	}
	plainText := func(s string) string { return s }

	a.println("")
	for i, l := range flow([]string{
		sty.Gray(fmt.Sprintf("Concept %d/%d ·", n, total)),
		sty.Bold(sty.Blue(truncate(c.Name, w-8))),
		masteryBadge(level) + " " + sty.Gray(masteryNames[level]),
	}, " ", w-4, "") {
		lead := sty.Blue("┃")
		if i == 0 {
			lead = sty.Blue("┏━")
		}
		a.printf("  %s %s\n", lead, l)
	}
	paragraph(c.Summary, sty.Italic)

	if c.Analogy != "" {
		section("💬 In plain words", sty.Magenta)
		paragraph(c.Analogy, plainText)
	}
	if c.Example != "" {
		section("🌍 Real life", sty.Green)
		paragraph(c.Example, plainText)
	}
	section("🔍 The details", sty.Cyan)
	paragraph(c.Body, plainText)
	if c.Diagram != "" {
		blank()
		clipped := false
		for _, l := range dedent(c.Diagram) {
			if visibleLen(l) > w-6 {
				l, clipped = truncate(l, w-6), true
			}
			a.printf("  %s   %s\n", edge, sty.Cyan(l))
		}
		if clipped {
			paragraph("(diagram clipped: widen the terminal and press Ctrl+L to see it whole)", sty.Gray)
		}
	}
	if c.MentalModel != "" {
		section("◆ Mental model", sty.Yellow)
		paragraph(c.MentalModel, sty.Yellow)
	}
	if c.TryIt != "" {
		section("▶ Try it (optional)", sty.Green)
		paragraph(c.TryIt, plainText)
	}
	if len(c.Exercises) > 0 {
		section("🏋 Exercises", sty.Magenta)
		for i, e := range c.Exercises {
			mark := sty.Gray("○")
			if i < len(done) && done[i] {
				mark = sty.Green("✔")
			}
			a.printf("  %s %s %s\n", edge, mark, levelBadge(e.Level))
			for _, l := range wrap(e.Task, w-10, "") {
				a.printf("  %s     %s\n", edge, l)
			}
		}
		blank()
		paragraph("Hints and check-off: Study Hall → Exercise gym.", sty.Gray)
	}
	if l, ok := conceptLinks[c.Name]; ok {
		if len(l.Related) > 0 {
			section("🔗 Connects to", sty.Blue)
			for _, name := range l.Related {
				tag := fmt.Sprintf(" (Stage %d)", stageOfConcept(name))
				a.printf("  %s   %s %s\n", edge, sty.Blue("↔"), truncate(name, w-10-len(tag))+sty.Gray(tag))
			}
		}
		if l.Read != "" {
			section("📚 Go deeper", sty.Blue)
			paragraph(l.Read, plainText)
		}
	}
	if note != "" {
		section("📝 In your own words", sty.Magenta)
		paragraph(note, sty.Magenta)
	}
	if terms := termsIn(glossary, c.Summary, c.Analogy, c.Example, c.Body); len(terms) > 0 {
		section("📖 Words to know", sty.Blue)
		for _, t := range terms {
			for i, l := range wrap(t.Meaning, w-10-len(t.Word), "") {
				lead := strings.Repeat(" ", len(t.Word)+3)
				if i == 0 {
					lead = sty.Bold(t.Word) + " — "
				}
				a.printf("  %s   %s%s\n", edge, lead, sty.Gray(l))
			}
		}
	}
	a.println("  " + sty.Blue("┗"+strings.Repeat("━", w-4)))
}

// levelBadge renders an exercise level as a coloured label.
func levelBadge(level string) string {
	switch level {
	case LevelWarmUp:
		return sty.Bold(sty.Green("● WARM-UP")) + sty.Gray(" · no code needed")
	case LevelPractice:
		return sty.Bold(sty.Yellow("● PRACTICE")) + sty.Gray(" · a small program")
	case LevelRealWorld:
		return sty.Bold(sty.Red("● REAL-WORLD")) + sty.Gray(" · a real situation")
	default:
		return sty.Bold(level)
	}
}

// exerciseDots shows a concept's exercises as ●/○, one per exercise.
// Caller holds mu.
func exerciseDots(s *Stage, c Concept) string {
	var b strings.Builder
	for i := range c.Exercises {
		if s.hasDone(c.Name, i) {
			b.WriteString(sty.Magenta("●"))
		} else {
			b.WriteString(sty.Gray("○"))
		}
	}
	return b.String()
}

// exerciseGym lets the learner pick an exercise, reveal its hint and tick
// it off. q at any prompt returns to the Study Hall menu.
func (a *App) exerciseGym(stageID int, g StageGuide) error {
	for {
		a.println("")
		a.printf("  %s\n", sty.Bold(sty.Magenta("EXERCISE GYM · try first, then peek at the hint")))
		a.mu.Lock()
		_, s := a.reg.findStage(stageID)
		options := make([]string, len(g.Concepts))
		for i, c := range g.Concepts {
			options[i] = exerciseDots(s, c) + " " + c.Name
		}
		a.mu.Unlock()

		ci, err := a.con.promptChoice("Concept", options)
		if errors.Is(err, errCancel) {
			return nil
		}
		if err != nil {
			return err
		}
		c := g.Concepts[ci]

		a.mu.Lock()
		_, s = a.reg.findStage(stageID)
		exOptions := make([]string, len(c.Exercises))
		for i, e := range c.Exercises {
			mark := sty.Gray("○")
			if s.hasDone(c.Name, i) {
				mark = sty.Green("✔")
			}
			exOptions[i] = mark + " " + levelBadge(e.Level) + "\n        " + truncate(e.Task, a.cols()-12)
		}
		a.mu.Unlock()

		a.printf("\n  %s\n", sty.Bold(c.Name))
		ei, err := a.con.promptChoice("Exercise", exOptions)
		if errors.Is(err, errCancel) {
			continue
		}
		if err != nil {
			return err
		}
		if err := a.workExercise(stageID, c, ei); err != nil && !errors.Is(err, errCancel) {
			return err
		}
	}
}

// workExercise shows one exercise in full, offers the hint, and toggles
// its done state. In Classic Mode the hint sits behind the struggle clock,
// and the learner keeps a lab notebook: a plan before, observations after.
func (a *App) workExercise(stageID int, c Concept, i int) error {
	e := c.Exercises[i]
	key := exerciseKey(c.Name, i)
	w := a.cols()
	edge := sty.Magenta("┃")

	a.mu.Lock()
	classic := a.reg.ClassicMode
	opened, seen := a.reg.ExerciseOpened[key]
	notes := a.reg.notebookFor(stageID, key)
	a.mu.Unlock()

	a.println("")
	a.printf("  %s %s %s\n", sty.Magenta("┏━"), levelBadge(e.Level), sty.Gray("· "+c.Name))
	for _, l := range wrap(e.Task, w-6, "") {
		a.printf("  %s %s\n", edge, l)
	}
	a.println("  " + sty.Magenta("┗"+strings.Repeat("━", w-4)))

	if classic {
		for _, n := range notes {
			a.printf("  %s %s\n", sty.Yellow("📓 "+strings.ToUpper(n.Kind)), sty.Gray(formatTime(n.At)))
			for _, l := range wrap(n.Text, w-8, "     ") {
				a.println(l)
			}
		}
		if !seen {
			opened = time.Now()
			a.mutate(func(r *Registry) { r.ExerciseOpened[key] = opened })
			a.printf("  %s %s\n", classicBadge(), sty.Yellow(fmt.Sprintf("Struggle clock started: the hint unlocks in %d minutes.", struggleMinutes[e.Level])))
			plan, err := a.con.promptText("Before you start: your plan or prediction (Enter to skip)", maxNotesLen, false)
			if err != nil {
				return err
			}
			a.addNote(stageID, key, NotePlan, plan)
		}
	}

	show, err := a.con.confirm("Show the hint?")
	if err != nil {
		return err
	}
	if show && classic {
		a.mu.Lock()
		stuck := a.reg.stuckLogged(key)
		a.mu.Unlock()
		if ok, left := hintUnlocked(opened, e.Level, stuck, time.Now()); !ok {
			a.printf("  %s %s\n", sty.Yellow("🔒"), sty.Yellow(fmt.Sprintf("The hint unlocks in %d more minute(s). Productive struggle is where the learning happens.", int(left.Minutes())+1)))
			choice, err := a.con.promptChoice("What now?", []string{
				"Keep working (come back later)",
				"Log what I have tried so far",
				"I'm truly stuck: write down what I tried and unlock the hint now",
			})
			if err != nil {
				return err
			}
			switch choice {
			case 0:
				show = false
			case 1:
				text, err := a.con.promptText("What have you tried?", maxNotesLen, true)
				if err != nil {
					return err
				}
				a.addNote(stageID, key, NoteTried, text)
				a.con.ok("Logged. Keep going; the clock is still running.")
				show = false
			case 2:
				for {
					text, err := a.con.promptText("What did you try, and where exactly are you stuck?", maxNotesLen, true)
					if err != nil {
						return err
					}
					if len([]rune(text)) < 20 {
						a.con.warn("Say a little more (at least a sentence). Describing the problem often solves it.")
						continue
					}
					a.addNote(stageID, key, NoteStuck, text)
					break
				}
			}
		}
	}
	if show {
		a.println("")
		for i, l := range wrap(e.Hint, w-14, "") {
			lead := "           "
			if i == 0 {
				lead = sty.Bold(sty.Yellow("💡 Hint")) + "    "
			}
			a.printf("  %s%s\n", lead, sty.Yellow(l))
		}
	}

	a.mu.Lock()
	_, s := a.reg.findStage(stageID)
	already := s.hasDone(c.Name, i)
	a.mu.Unlock()

	label := "Mark this exercise as done?"
	if already {
		label = "Already done. Untick it?"
	}
	yes, err := a.con.confirm(label)
	if err != nil || !yes {
		return err
	}
	if classic && !already {
		observed, err := a.con.promptText("What happened? Did your plan or prediction hold? (Enter to skip)", maxNotesLen, false)
		if err != nil {
			return err
		}
		a.addNote(stageID, key, NoteObserved, observed)
	}
	a.mutate(func(r *Registry) {
		_, s := r.findStage(stageID)
		s.setDone(c.Name, i, !already)
	})
	if already {
		a.con.ok("Exercise unticked.")
	} else {
		a.con.ok("Nice work! %s exercise done.", levelBadgePlain(e.Level))
	}
	return nil
}

// levelBadgePlain is the level name without decoration, for sentences.
func levelBadgePlain(level string) string {
	switch level {
	case LevelWarmUp:
		return "Warm-up"
	case LevelPractice:
		return "Practice"
	case LevelRealWorld:
		return "Real-world"
	}
	return level
}

// termRegex caches one case-insensitive, plural-tolerant pattern per term.
var termRegex = map[string]*regexp.Regexp{}

// termsIn returns the glossary entries whose word appears in any text.
func termsIn(glossary []Term, texts ...string) []Term {
	joined := strings.Join(texts, " ")
	var found []Term
	for _, t := range glossary {
		re, ok := termRegex[t.Word]
		if !ok {
			re = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(t.Word) + `(s|es)?\b`)
			termRegex[t.Word] = re
		}
		if re.MatchString(joined) {
			found = append(found, t)
		}
	}
	return found
}

// showGlossary prints a stage's jargon in plain English.
func (a *App) showGlossary(g StageGuide) {
	a.println("")
	a.printf("  %s\n", sty.Bold(sty.Blue("GLOSSARY · words explained simply")))
	width := 0
	for _, t := range g.Glossary {
		width = max(width, len([]rune(t.Word)))
	}
	for _, t := range g.Glossary {
		for i, l := range wrap(t.Meaning, a.cols()-width-10, "") {
			lead := strings.Repeat(" ", width)
			if i == 0 {
				lead = padRight(t.Word, width)
			}
			a.printf("    %s  %s\n", sty.Bold(sty.Cyan(lead)), l)
		}
	}
	a.println("")
}

// showStartHere prints the orientation guide for new learners.
func (a *App) showStartHere() {
	a.println(heading("START HERE · how to learn with this academy", sty.Green, a.cols()))
	a.println("")
	for _, l := range wrap(startHere, a.cols()-4, "  ") {
		t := strings.TrimSpace(l)
		if isHeadingLine(t) {
			a.println("  " + sty.Bold(sty.Green(t)))
			continue
		}
		a.println(l)
	}
	a.println("")
	a.con.note("Next step: open the Study Hall [6], pick Stage 1 or 5, and study concept 1.")
}

// showResources prints the resource library grouped by kind.
func (a *App) showResources(g StageGuide) {
	a.println("")
	for _, l := range flow([]string{sty.Bold(sty.Blue("RESOURCE LIBRARY")), sty.Gray("links verified " + resourcesVerifiedOn), sty.Gray("re-check any time: academy -check-links")}, sty.Gray(" · "), a.cols(), "  ") {
		a.println(l)
	}
	order := []string{"Course", "Book", "Video", "Article", "Paper", "Tool", "Site"}
	for _, kind := range order {
		first := true
		for _, r := range g.Resources {
			if r.Kind != kind {
				continue
			}
			if first {
				a.printf("\n  %s\n", sty.Magenta(sty.Bold(strings.ToUpper(kind)+"S")))
				first = false
			}
			for i, l := range wrap(r.Title, a.cols()-8, "") {
				lead := "  "
				if i == 0 {
					lead = sty.Blue("◇") + " "
				}
				a.printf("    %s%s\n", lead, sty.Bold(l))
			}
			if r.URL != "" {
				a.printf("      %s\n", sty.Under(sty.Cyan(r.URL)))
			}
			for _, l := range wrap(r.Note, a.cols()-8, "      ") {
				a.println(sty.Gray(l))
			}
		}
	}
	a.println("")
}

// showBlueprints lists suggested labs and optionally enrolls one.
func (a *App) showBlueprints(stageID int, g StageGuide) error {
	a.println("")
	a.printf("  %s\n", sty.Bold(sty.Green("LAB BLUEPRINTS")))
	names := make([]string, len(g.Blueprints))
	for i, b := range g.Blueprints {
		names[i] = b.Name
		a.printf("\n  %s %s\n", sty.Green(fmt.Sprintf("[%d]", i+1)), sty.Bold(b.Name))
		for _, l := range wrap(b.Brief, a.cols()-8, "      ") {
			a.println(l)
		}
		for m, step := range b.Milestones {
			a.printf("      %s %s\n", sty.Gray(fmt.Sprintf("%d.", m+1)), sty.Gray(step))
		}
	}
	a.println("")

	yes, err := a.con.confirm("Enroll one of these as a lab?")
	if err != nil || !yes {
		return err
	}
	idx, err := a.con.promptChoice("Blueprint", names)
	if err != nil {
		return err
	}
	b := g.Blueprints[idx]
	notes := b.Brief + " Milestones: " + strings.Join(b.Milestones, " → ")
	if len([]rune(notes)) > maxNotesLen {
		notes = b.Brief
	}
	id, err := a.addLab(stageID, b.Name, notes, 0)
	if errors.Is(err, errDuplicateLab) {
		a.con.warn("%v", err)
		return nil
	}
	if err != nil {
		return err
	}
	a.con.ok("Enrolled lab #%d %s in stage %d. Log hours with [3].", id, sty.Bold(b.Name), stageID)
	return nil
}

// runQuiz walks the flashcards, revealing each answer on Enter, and scores
// the user's own assessment. Scores are not persisted.
func (a *App) runQuiz(g StageGuide) error {
	if len(g.Quiz) == 0 {
		a.con.note("No questions for this stage yet.")
		return nil
	}
	score, asked := 0, 0
	for i, q := range g.Quiz {
		a.printf("\n  %s\n", sty.Bold(sty.Magenta(fmt.Sprintf("Q%d/%d", i+1, len(g.Quiz)))))
		for _, l := range wrap(q.Q, a.cols()-6, "  ") {
			a.println(sty.Bold(l))
		}
		s, err := a.con.readLine(promptLabel("Think, then press Enter to reveal", "q stop"))
		if err != nil {
			return err
		}
		if isCancel(s) {
			break
		}
		for _, l := range wrap(q.A, a.cols()-8, "    ") {
			a.println(sty.Green(l))
		}
		got, err := a.con.confirm("Did you get it?")
		if err != nil {
			return err
		}
		asked++
		if got {
			score++
		}
	}
	if asked > 0 {
		a.printf("\n  %s %s %s\n", sty.Gray("Score"), bar(score, asked, 20, sty.Magenta),
			sty.Bold(fmt.Sprintf("%d/%d", score, asked)))
		if score < asked {
			a.con.note("Revisit the concepts behind the ones you missed.")
		}
	}
	return nil
}
