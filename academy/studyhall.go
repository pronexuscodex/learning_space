package main

// Study Hall: per-stage concept explanations, resource library, lab
// blueprints (one-step enrolment) and self-check quizzes.

import (
	"errors"
	"fmt"
	"strings"
)

// studyHall picks a stage and loops over its learning menu.
func (a *App) studyHall() error {
	a.println(heading("STUDY HALL · concepts, resources & blueprints", sty.Blue, a.width))
	stageID, err := a.pickStage(nil)
	if err != nil {
		return err
	}
	guide, ok := guideFor(stageID)
	if !ok {
		a.con.note("No Study Hall material for stage %d yet.", stageID)
		return nil
	}

	for {
		a.renderStageSyllabus(stageID, guide)
		choice, err := a.con.promptChoice("Study Hall", []string{
			"Study a concept",
			fmt.Sprintf("Resource library %s", sty.Gray(fmt.Sprintf("(%d)", len(guide.Resources)))),
			fmt.Sprintf("Lab blueprints %s", sty.Gray(fmt.Sprintf("(%d, enroll with one keystroke)", len(guide.Blueprints)))),
			fmt.Sprintf("Self-check quiz %s", sty.Gray(fmt.Sprintf("(%d questions)", len(guide.Quiz)))),
			"Back to main menu",
		})
		if err != nil {
			return err
		}
		switch choice {
		case 0:
			err = a.studyConcept(stageID, guide)
		case 1:
			a.showResources(guide)
		case 2:
			err = a.showBlueprints(stageID, guide)
		case 3:
			err = a.runQuiz(guide)
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
	studiedSet := map[string]bool{}
	for _, c := range g.Concepts {
		studiedSet[c.Name] = s.hasStudied(c.Name)
	}
	studied, total := conceptProgress(s)
	a.mu.Unlock()

	a.printf("\n  %s %s %s\n", color(sty.Bold(fmt.Sprintf("Stage %d", stageID))), sty.Gray("·"), sty.Bold(title))
	for _, l := range wrap(g.Overview, a.width-4, "  ") {
		a.println(sty.Italic(l))
	}
	a.printf("\n  %s %s %d/%d\n", sty.Gray("Concepts"), bar(studied, total, 20, sty.Blue), studied, total)
	for i, c := range g.Concepts {
		mark := sty.Gray("○")
		if studiedSet[c.Name] {
			mark = sty.Green("✔")
		}
		a.printf("    %s %s %s %s\n", mark, color(fmt.Sprintf("%d.", i+1)), sty.Bold(c.Name), sty.Gray("— "+truncate(c.Summary, a.width-len(c.Name)-14)))
	}
	a.println("")
}

// studyConcept renders one concept card and offers to mark it understood.
func (a *App) studyConcept(stageID int, g StageGuide) error {
	names := make([]string, len(g.Concepts))
	for i, c := range g.Concepts {
		names[i] = c.Name
	}
	idx, err := a.con.promptChoice("Concept", names)
	if err != nil {
		return err
	}
	c := g.Concepts[idx]
	a.renderConcept(c, idx+1, len(g.Concepts))

	a.mu.Lock()
	_, s := a.reg.findStage(stageID)
	already := s.hasStudied(c.Name)
	a.mu.Unlock()

	label := "Mark as understood?"
	if already {
		label = "Already understood. Move back to 'needs review'?"
	}
	yes, err := a.con.confirm(label)
	if err != nil || !yes {
		return err
	}
	a.mutate(func(r *Registry) {
		_, s := r.findStage(stageID)
		s.setStudied(c.Name, !already)
	})
	if already {
		a.con.ok("%s moved back to review.", sty.Bold(c.Name))
	} else {
		a.con.ok("%s marked as understood.", sty.Bold(c.Name))
	}
	return nil
}

// renderConcept prints a concept as a heavy-bordered reading card.
func (a *App) renderConcept(c Concept, n, total int) {
	w := a.width
	edge := sty.Blue("┃")
	a.println("")
	a.printf("  %s %s %s\n", sty.Blue("┏━"), sty.Gray(fmt.Sprintf("Concept %d/%d ·", n, total)), sty.Bold(sty.Blue(c.Name)))
	a.printf("  %s %s\n", edge, sty.Italic(c.Summary))
	a.printf("  %s\n", edge)
	for _, l := range wrap(c.Body, w-6, "") {
		a.printf("  %s %s\n", edge, l)
	}
	if c.Diagram != "" {
		a.printf("  %s\n", edge)
		for _, l := range dedent(c.Diagram) {
			a.printf("  %s   %s\n", edge, sty.Cyan(l))
		}
	}
	if c.MentalModel != "" {
		a.printf("  %s\n", edge)
		for i, l := range wrap(c.MentalModel, w-24, "") {
			lead := "              "
			if i == 0 {
				lead = sty.Bold(sty.Yellow("◆ Mental model")) + " "
			} else {
				lead += " "
			}
			a.printf("  %s %s%s\n", edge, lead, sty.Yellow(l))
		}
	}
	if c.TryIt != "" {
		for i, l := range wrap(c.TryIt, w-24, "") {
			lead := "              "
			if i == 0 {
				lead = padRight(sty.Bold(sty.Green("▶ Try it")), 14) + " "
			} else {
				lead += " "
			}
			a.printf("  %s %s%s\n", edge, lead, l)
		}
	}
	a.println("  " + sty.Blue("┗"+strings.Repeat("━", w-4)))
}

// showResources prints the resource library grouped by kind.
func (a *App) showResources(g StageGuide) {
	a.println("")
	a.printf("  %s\n", sty.Bold(sty.Blue("RESOURCE LIBRARY")))
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
			a.printf("    %s %s\n", sty.Blue("◇"), sty.Bold(r.Title))
			if r.URL != "" {
				a.printf("      %s\n", sty.Under(sty.Cyan(r.URL)))
			}
			for _, l := range wrap(r.Note, a.width-8, "      ") {
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
		for _, l := range wrap(b.Brief, a.width-8, "      ") {
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
		for _, l := range wrap(q.Q, a.width-6, "  ") {
			a.println(sty.Bold(l))
		}
		s, err := a.con.readLine(promptLabel("Think, then press Enter to reveal", "q stop"))
		if err != nil {
			return err
		}
		if isCancel(s) {
			break
		}
		for _, l := range wrap(q.A, a.width-8, "    ") {
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
