package main

// Daily Review (spaced repetition) and the per-stage Mastery Check.

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// masteryBadge renders a mastery level as a coloured glyph.
func masteryBadge(level int) string {
	g := masteryGlyphs[level]
	switch level {
	case MasteryMastered:
		return sty.Bold(sty.Green(g))
	case MasteryRetained:
		return sty.Green(g)
	case MasteryPractised:
		return sty.Cyan(g)
	case MasteryUnderstood:
		return sty.Yellow(g)
	}
	return sty.Gray(g)
}

// masteryLegend explains the glyphs in one line.
func masteryLegend() string {
	parts := make([]string, len(masteryNames))
	for i, n := range masteryNames {
		parts[i] = masteryBadge(i) + " " + sty.Gray(n)
	}
	return strings.Join(parts, "  ")
}

// dailyReview runs a spaced-repetition session over every due card,
// interleaved across stages.
func (a *App) dailyReview() error {
	a.println(heading("DAILY REVIEW · remember what you learned", sty.Green, a.width))
	a.mu.Lock()
	now := time.Now()
	due := a.reg.dueCards(now)
	unlocked := len(a.reg.unlockedCards())
	a.mu.Unlock()

	if unlocked == 0 {
		a.con.note("Your review deck is empty. Mark a concept as understood in the Study Hall [6] and its cards join the deck.")
		return nil
	}
	if len(due) == 0 {
		a.con.ok("Nothing is due. All %d cards are scheduled for later. Come back tomorrow!", unlocked)
		return nil
	}

	a.printf("  %s\n", sty.Bold(fmt.Sprintf("%d card(s) due out of %d.", len(due), unlocked)))
	for _, l := range wrap("Recall each answer before you reveal it; that effort is what builds memory.", a.width-4, "  ") {
		a.println(l)
	}
	a.printf("  %s\n", sty.Gray("Grade honestly: 1 forgot · 2 hard · 3 good · 4 easy. q stops (progress is kept)."))

	// Interleave: shuffle so stages and card types mix.
	rand.Shuffle(len(due), func(i, j int) { due[i], due[j] = due[j], due[i] })

	queue := due
	reviewed, correct := 0, 0
	requeued := map[string]bool{}
	for len(queue) > 0 {
		card := queue[0]
		queue = queue[1:]
		grade, err := a.reviewCard(card, reviewed+1, reviewed+1+len(queue))
		if errors.Is(err, errCancel) {
			break
		}
		if err != nil {
			return err
		}
		reviewed++
		if grade >= GradeHard {
			correct++
		}
		// Forgotten cards come back once more at the end of the session.
		if grade == GradeAgain && !requeued[card.ID] {
			requeued[card.ID] = true
			queue = append(queue, card)
		}
	}

	if reviewed > 0 {
		a.mu.Lock()
		left := len(a.reg.dueCards(time.Now().Add(24 * time.Hour)))
		a.mu.Unlock()
		a.printf("\n  %s %s %s   %s\n", sty.Gray("Session"), bar(correct, reviewed, 20, sty.Green),
			sty.Bold(fmt.Sprintf("%d/%d recalled", correct, reviewed)),
			sty.Gray(fmt.Sprintf("%d card(s) due by tomorrow", left)))
	}
	return nil
}

// reviewCard shows one card, lets the learner attempt it, reveals the
// answer (with their own notes), and records the grade.
func (a *App) reviewCard(card Card, n, total int) (int, error) {
	a.mu.Lock()
	st := a.reg.Reviews[card.ID]
	_, stage := a.reg.findStage(card.StageID)
	note := ""
	if stage != nil && card.Kind == CardKeyIdea {
		note = stage.Notes[card.Concept] // compare your words with the key idea
	}
	a.mu.Unlock()

	w := a.width
	edge := sty.Green("┃")
	a.println("")
	source := fmt.Sprintf("Stage %d", card.StageID)
	if card.Concept != "" {
		source += " · " + card.Concept
	}
	a.printf("  %s %s %s %s\n", sty.Green("┏━"), sty.Gray(fmt.Sprintf("%d/%d ·", n, total)),
		sty.Bold(sty.Green(strings.ToUpper(card.Kind))), sty.Gray(source))
	for _, l := range wrap(card.Q, w-6, "") {
		a.printf("  %s %s\n", edge, sty.Bold(l))
	}
	a.println("  " + sty.Green("┗"+strings.Repeat("━", w-4)))

	attempt, err := a.con.readLine(promptLabel("Your answer", "type it or think it, Enter to reveal, q stop"))
	if err != nil {
		return 0, err
	}
	if isCancel(attempt) {
		return 0, errCancel
	}

	a.printf("  %s\n", sty.Bold(sty.Cyan("Answer")))
	for _, l := range wrap(card.A, w-6, "    ") {
		a.println(sty.Cyan(l))
	}
	if note != "" {
		a.printf("  %s\n", sty.Bold(sty.Magenta("📝 In your own words")))
		for _, l := range wrap(note, w-6, "    ") {
			a.println(sty.Magenta(l))
		}
	}

	now := time.Now()
	options := []string{
		sty.Red("Forgot") + sty.Gray(" · see it again soon"),
		sty.Yellow("Hard") + sty.Gray(" · next in "+previewInterval(st, GradeHard, now)),
		sty.Green("Good") + sty.Gray(" · next in "+previewInterval(st, GradeGood, now)),
		sty.Cyan("Easy") + sty.Gray(" · next in "+previewInterval(st, GradeEasy, now)),
	}
	idx, err := a.con.promptChoice("How well did you recall it?", options)
	if err != nil {
		return 0, err
	}
	grade := idx + 1
	a.mutate(func(r *Registry) {
		r.Reviews[card.ID] = schedule(r.Reviews[card.ID], grade, now)
		r.ReviewHistory[now.Format("2006-01-02")]++
	})
	return grade, nil
}

// masteryCheck is an interleaved stage exam drawn from every card the stage
// can produce. Passing needs masteryPassMark.
func (a *App) masteryCheck(stageID int) error {
	cards := stageCards(stageID)
	if len(cards) == 0 {
		a.con.note("No questions for this stage yet.")
		return nil
	}

	a.mu.Lock()
	_, s := a.reg.findStage(stageID)
	studied, total := conceptProgress(s)
	prev := s.MasteryCheck
	a.mu.Unlock()

	a.println("")
	a.printf("  %s\n", sty.Bold(sty.Green("MASTERY CHECK · mixed questions from the whole stage")))
	if prev != nil {
		verdict := sty.Yellow("not passed")
		if prev.Passed {
			verdict = sty.Green("passed")
		}
		a.printf("  %s %d/%d on %s (%s)\n", sty.Gray("Last attempt:"), prev.Score, prev.Total, formatTime(prev.TakenAt), verdict)
	}
	if studied < total {
		a.con.warn("You have understood %d of %d concepts; the check covers all of them.", studied, total)
	}
	a.printf("  %s\n", sty.Gray(fmt.Sprintf("Answer in your head (or type), reveal, and judge yourself honestly. %d%% or more passes.", int(masteryPassMark*100))))
	ok, err := a.con.confirm("Start the mastery check?")
	if err != nil || !ok {
		return err
	}

	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	if len(cards) > masteryCheckSize {
		cards = cards[:masteryCheckSize]
	}

	score := 0
	for i, c := range cards {
		a.printf("\n  %s %s\n", sty.Bold(sty.Green(fmt.Sprintf("Q%d/%d", i+1, len(cards)))), sty.Gray(c.Kind))
		for _, l := range wrap(c.Q, a.width-6, "  ") {
			a.println(sty.Bold(l))
		}
		attempt, err := a.con.readLine(promptLabel("Your answer", "Enter to reveal, q abandons the check"))
		if err != nil {
			return err
		}
		if isCancel(attempt) {
			a.con.note("Check abandoned; nothing recorded.")
			return nil
		}
		for _, l := range wrap(c.A, a.width-8, "    ") {
			a.println(sty.Cyan(l))
		}
		got, err := a.con.confirm("Did you get the essentials right?")
		if err != nil {
			return err
		}
		if got {
			score++
		}
	}

	passed := float64(score) >= masteryPassMark*float64(len(cards))
	a.mutate(func(r *Registry) {
		_, s := r.findStage(stageID)
		s.MasteryCheck = &MasteryCheck{TakenAt: time.Now().UTC().Truncate(time.Second), Score: score, Total: len(cards), Passed: passed}
	})
	a.printf("\n  %s %s %s\n", sty.Gray("Score"), bar(score, len(cards), 20, sty.Green), sty.Bold(fmt.Sprintf("%d/%d", score, len(cards))))
	if passed {
		a.con.ok("Mastery check passed. Keep up your Daily Review so it stays that way, and retake this in a month.")
	} else {
		a.con.note("Not yet. Re-study the concepts you missed, do their exercises, and try again in a few days.")
	}
	return nil
}
