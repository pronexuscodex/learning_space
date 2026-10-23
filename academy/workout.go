package main

// The Daily Workout: one guided session a day that strings the strongest
// study habits together, in the order that works: retrieve, recall,
// learn, practise (mixed), reflect.

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var workoutInfo = map[string]struct{ Title, Why string }{
	StepReview: {"Warm-up review",
		"Retrieve before you learn anything new. Spaced retrieval is the best-proven way to remember."},
	StepRecall: {"Blank-page recall",
		"Write what you remember about one concept, then check. Pulling it out of your head beats rereading."},
	StepLearn: {"Learn one new idea",
		"Read one concept, then explain it in your own words. Producing an explanation is what makes it stick."},
	StepPractice: {"Mixed practice",
		"An exercise picked from everything you have understood so far, mixing topics once you have several. Mixing feels harder, and that difficulty is what teaches."},
	StepReflect: {"Reflect",
		"Two honest questions. Noticing how you learn is how you get better at learning."},
}

// workoutLine is the one-line status shown on the dashboard.
func (a *App) workoutLine() string {
	a.mu.Lock()
	day := a.reg.todayWorkout(time.Now())
	a.mu.Unlock()
	parts := ""
	for i, step := range workoutSteps {
		mark := sty.Gray("○")
		if day != nil && day.has(step) {
			mark = sty.Green("●")
		}
		if i > 0 {
			parts += " "
		}
		parts += mark
	}
	if day != nil && day.complete() {
		return "  " + sty.Bold(sty.Green("Today's workout done")) + " " + parts
	}
	return "  " + sty.Green("➜ Today's workout") + " " + parts + " " + sty.Gray("(press j)")
}

// dailyWorkout runs today's remaining workout steps.
func (a *App) dailyWorkout() error {
	w := a.cols()
	a.println(heading("TODAY'S WORKOUT", sty.Green, w))
	for _, l := range wrap("Think first, look it up second, ask an AI last, and only to check your own answer. About 30 to 45 minutes. Each step is saved as you go, so you can stop and come back.", w-4, "  ") {
		a.println(l)
	}
	learnedStage := -1
	for n, step := range workoutSteps {
		a.mu.Lock()
		done := a.reg.workoutFor(time.Now()).has(step)
		a.mu.Unlock()
		if done {
			continue
		}
		info := workoutInfo[step]
		if empty := a.emptyStep(step); empty != "" {
			// Nothing to do in this step today: say so and move on,
			// rather than asking "Ready?" first.
			// Recorded as skipped too: the checklist shows it as "–", not
			// as work done, and it does not count towards the streak.
			a.mutate(func(r *Registry) {
				if day := r.workoutFor(time.Now()); !day.has(step) {
					day.Done = append(day.Done, step)
					day.Skipped = append(day.Skipped, step)
				}
			})
			a.con.say(sty.Gray("✓"), fmt.Sprintf("%s · %s: %s", fmt.Sprintf("Step %d", n+1), info.Title, empty), sty.Gray)
			continue
		}
		a.println("")
		a.println(a.workoutChecklist())
		a.printf("\n  %s %s %s\n", sty.Bold(sty.Green(fmt.Sprintf("Step %d of %d", n+1, len(workoutSteps)))), sty.Gray("·"), sty.Bold(info.Title))
		for _, l := range wrap(info.Why, w-6, "    ") {
			a.println(sty.Gray(l))
		}
		choice, err := a.con.promptChoice("Ready?", []string{"Start", "Skip this step today", "Stop for now (progress is kept)"})
		if errors.Is(err, errCancel) || (err == nil && choice == 2) {
			a.con.note("Workout paused. Press j on the menu to continue where you left off.")
			return nil
		}
		if err != nil {
			return err
		}
		if choice == 0 {
			stage, err := a.runWorkoutStep(step, learnedStage)
			if err != nil && !errors.Is(err, errCancel) {
				return err
			}
			if stage >= 0 {
				learnedStage = stage
			}
		}
		a.mutate(func(r *Registry) {
			day := r.workoutFor(time.Now())
			if !day.has(step) {
				day.Done = append(day.Done, step)
			}
			if choice == 1 && !contains(day.Skipped, step) {
				day.Skipped = append(day.Skipped, step)
			}
		})
	}
	a.println("")
	a.println(a.workoutChecklist())
	a.mu.Lock()
	n := a.reg.workoutsCompleted()
	a.mu.Unlock()
	a.con.ok("Workout complete: %d so far. Same time tomorrow builds the habit.", n)
	a.announceAchievements()
	return nil
}

// workoutChecklist renders today's steps with their state.
func (a *App) workoutChecklist() string {
	a.mu.Lock()
	day := a.reg.todayWorkout(time.Now())
	a.mu.Unlock()
	var items []string
	for _, step := range workoutSteps {
		mark, label := sty.Gray("○"), workoutInfo[step].Title
		switch {
		case day != nil && contains(day.Skipped, step):
			mark, label = sty.Gray("–"), sty.Gray(label)
		case day != nil && day.has(step):
			mark = sty.Green("●")
		}
		items = append(items, mark+" "+label)
	}
	out := ""
	for i, l := range flow(items, "   ", a.cols(), "  ") {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}

// runWorkoutStep runs one step. It returns the stage of a newly learned
// concept (so practice can pick a different topic), or -1.
func (a *App) runWorkoutStep(step string, learnedStage int) (int, error) {
	now := time.Now()
	switch step {
	case StepReview:
		a.mu.Lock()
		due := len(a.reg.dueCards(now))
		a.mu.Unlock()
		if due == 0 {
			a.con.ok("Nothing is due for review. Your memory is up to date.")
			return -1, nil
		}
		return -1, a.dailyReview()
	case StepRecall:
		a.mu.Lock()
		cands := a.reg.recallCandidates(now)
		a.mu.Unlock()
		if len(cands) == 0 {
			a.con.note("Nothing to recall yet: once you understand a concept, it comes back here.")
			return -1, nil
		}
		return -1, a.freeRecall(cands[0].Stage, cands[0].Concept)
	case StepLearn:
		a.mu.Lock()
		stage, ci, ok := a.reg.nextConcept()
		a.mu.Unlock()
		if !ok {
			a.con.ok("Every available concept is understood. Use the time for a mastery check or a lab.")
			return -1, nil
		}
		g, _ := guideFor(stage)
		return stage, a.studyConceptAt(stage, g, ci)
	case StepPractice:
		a.mu.Lock()
		role, hasRole := roleByID(a.reg.Career.Role)
		redos := len(a.reg.interviewRedos(now))
		a.mu.Unlock()
		if hasRole {
			interview := "An interview problem for " + role.Name + " (25–45 min, aloud)"
			if redos > 0 {
				interview = "An interview problem to redo alone (" + count(redos, "is", "are") + " due)"
			}
			n, err := a.con.promptChoice("Practise what?", []string{"An exercise from the curriculum (mixed topics)", interview})
			if err != nil {
				return -1, err
			}
			if n == 1 {
				return -1, a.interviewProblem()
			}
		}
		a.mu.Lock()
		p, ok := a.reg.practicePick(now, learnedStage, rand.Intn)
		a.mu.Unlock()
		if !ok {
			a.con.note("No open exercises yet: understand a concept first, and its exercises appear here.")
			return -1, nil
		}
		g, _ := guideFor(p.Stage)
		c := g.Concepts[p.Concept]
		if p.Redo != nil {
			a.con.say(sty.Cyan("↻"), fmt.Sprintf("A redo from a blank page: on %s you solved this with help. Now do it alone.", p.Redo.At.Format("2 January")), sty.Cyan)
		} else {
			a.con.note("From Stage %d · %s", p.Stage, c.Name)
		}
		return -1, a.workExercise(p.Stage, c, p.Exercise)
	case StepReflect:
		return -1, a.reflect()
	}
	return -1, nil
}

// reflect asks the two end-of-session questions.
func (a *App) reflect() error {
	hardest, err := a.con.promptText("What was the hardest thing today, and why? (Enter to skip)", maxNotesLen, false)
	if err != nil {
		return err
	}
	asked, err := a.con.confirm("Did you ask an AI for an answer today before trying it yourself?")
	if err != nil {
		return err
	}
	a.mutate(func(r *Registry) {
		day := r.workoutFor(time.Now())
		day.Hardest = hardest
		day.AskedFirst = &asked
	})
	if asked {
		a.con.note("Thanks for being honest. Next time try the order engineers use: 20 minutes of your own attempt, then the docs, and only then ask an AI to check your answer, not to write it.")
	} else {
		a.con.ok("You did the thinking yourself. That is how understanding is built.")
	}
	return nil
}

// emptyStep says why a workout step has nothing to do today, or "" when
// it has.
func (a *App) emptyStep(step string) string {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	switch step {
	case StepReview:
		if len(a.reg.dueCards(now)) == 0 {
			return "nothing is due, your memory is up to date"
		}
	case StepRecall:
		if len(a.reg.recallCandidates(now)) == 0 {
			return "nothing to recall yet; once you understand a concept, it comes back here"
		}
	case StepLearn:
		if _, _, ok := a.reg.nextConcept(); !ok {
			return "every available concept is understood; use the time for a mastery check or a lab"
		}
	case StepPractice:
		if _, ok := a.reg.practicePick(now, -1, func(int) int { return 0 }); !ok {
			if _, hasRole := roleByID(a.reg.Career.Role); !hasRole {
				return "no open exercises yet; understand a concept first and its exercises appear here"
			}
		}
	}
	return ""
}
