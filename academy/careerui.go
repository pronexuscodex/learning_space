package main

// Screens of the career center (see careers.go).

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"time"
)

// careerOptions are the career center's menu, in order.
var careerOptions = []string{
	"Choose your target role (what the job is, what interviews test)",
	"Readiness report (what is missing, step by step)",
	"Portfolio projects (build what employers look for)",
	"Coding interview patterns (16 patterns, 32 practice problems)",
	"System design (the framework and 9 classic questions)",
	"Behavioral interviews and your story bank (STAR)",
	"Mock interview (timed, scored against a rubric)",
	"Resume builder (bullets for your projects)",
	"Job search (kit, applications tracker, tips)",
	"Export your career kit (Markdown)",
	"Back to the main menu",
}

// careerCenter is the career center's main screen.
func (a *App) careerCenter() error {
	for {
		a.renderCareerHome()
		n, err := a.con.promptChoice("Career center", careerOptions)
		if errors.Is(err, errCancel) {
			return nil
		}
		if err != nil {
			return err
		}
		switch n {
		case 0:
			err = a.chooseRole()
		case 1:
			a.paged(true, a.renderReadiness)
		case 2:
			err = a.portfolio()
		case 3:
			err = a.codingPractice()
		case 4:
			err = a.systemDesign()
		case 5:
			err = a.storyBank()
		case 6:
			err = a.mockInterview()
		case 7:
			err = a.resumeBuilder()
		case 8:
			err = a.jobSearch()
		case 9:
			err = a.exportCareer()
		default:
			return nil
		}
		if err != nil && !errors.Is(err, errCancel) {
			return err
		}
	}
}

// targetRole returns the learner's target role.
func (a *App) targetRole() (Role, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return roleByID(a.reg.Career.Role)
}

// subhead prints a bold subheading, wrapped to the screen.
func (a *App) subhead(title string, color func(string) string) {
	a.println("")
	for _, l := range wrap(title, a.cols()-4, "") {
		a.println("  " + sty.Bold(color(l)))
	}
}

// grayWrap prints gray text wrapped with an indent.
func (a *App) grayWrap(text, indent string) {
	for _, l := range wrap(text, a.cols()-len(indent)-2, indent) {
		a.println(sty.Gray(l))
	}
}

// renderCareerHome prints the career center's header: the target role and
// how ready the learner is.
func (a *App) renderCareerHome() {
	w := a.cols()
	a.println("")
	a.println(heading("GET HIRED · the career center", sty.Green, w))
	role, ok := a.targetRole()
	if !ok {
		for _, l := range wrap("Companies hire engineers who can show three things: solid foundations, projects they built and finished, and that they can solve problems out loud in an interview. This center turns your learning into all three. Start by choosing the job you want [1].", w-4, "  ") {
			a.println(l)
		}
		a.println("")
		return
	}
	a.mu.Lock()
	rd := a.reg.readiness(role)
	steps := a.reg.careerSteps(role, rd)
	a.mu.Unlock()
	a.printf("  %s %s\n", sty.Gray("Target role:"), sty.Bold(role.Name))
	a.printf("  %s %s %s\n", sty.Gray("Readiness:  "), bar(int(rd.Total*100), 100, min(30, w-26), sty.Green), sty.Bold(percent(rd.Total)))
	parts := fmt.Sprintf("foundations %s · portfolio %s · practice %s · job-search kit %s",
		percent(rd.Foundations), percent(rd.Portfolio), percent(rd.Practice), percent(rd.Kit))
	for _, l := range wrap(parts, w-4, "  ") {
		a.println(sty.Gray(l))
	}
	for _, l := range wrap(rd.verdict(), w-4, "  ") {
		a.println(sty.Green(l))
	}
	if len(steps) > 0 {
		for i, l := range wrap(steps[0], w-12, "") {
			lead := "  " + sty.Bold(sty.Yellow("Next ▸")) + " "
			if i > 0 {
				lead = "         "
			}
			a.println(lead + l)
		}
	}
	a.println("")
}

// renderRole prints everything about one role.
func (a *App) renderRole(role Role) {
	w := a.cols()
	para := func(title string, color func(string) string, text string) {
		a.subhead(title, color)
		for _, l := range wrap(text, w-6, "    ") {
			a.println(l)
		}
	}
	list := func(title string, color func(string) string, items []string) {
		a.subhead(title, color)
		for _, it := range items {
			for i, l := range wrap(it, w-8, "") {
				lead := "    • "
				if i > 0 {
					lead = "      "
				}
				a.println(lead + l)
			}
		}
	}
	a.println("")
	a.println(heading(strings.ToUpper(role.Name), sty.Green, w))
	para("The work", sty.Cyan, role.Does)
	para("Who hires", sty.Cyan, role.Hires)
	list("Job titles to search for", sty.Cyan, role.Titles)
	list("What the interviews test", sty.Yellow, role.Screens)
	list("A typical interview process", sty.Yellow, role.Loop)

	a.subhead("Stages that prepare you", sty.Magenta)
	a.mu.Lock()
	for _, id := range role.Stages {
		g, ok := guideFor(id)
		_, s := a.reg.findStage(id)
		if !ok || s == nil {
			continue
		}
		done := 0
		for _, c := range g.Concepts {
			if s.hasStudied(c.Name) {
				done++
			}
		}
		a.printf("    %s %s %s\n", bar(done, len(g.Concepts), 8, sty.Green), sty.Gray(fmt.Sprintf("%d/%d", done, len(g.Concepts))),
			truncate(fmt.Sprintf("Stage %d · %s", id, s.Title), w-20))
	}
	a.mu.Unlock()
	list("Portfolio projects employers value", sty.Magenta, role.Projects)
	para("Skills to show on your resume (only what you can be interviewed on)", sty.Magenta, strings.Join(role.Keywords, " · "))
	para("Advice", sty.Green, role.Advice)
}

// chooseRole lets the learner read about each role and pick a target.
func (a *App) chooseRole() error {
	var names []string
	for _, r := range roles {
		names = append(names, r.Name)
	}
	a.println("")
	n, err := a.con.promptChoice("Read about which role?", names)
	if err != nil {
		return err
	}
	role := roles[n]
	a.paged(false, func() { a.renderRole(role) })
	a.println("")
	if cur, ok := a.targetRole(); ok && cur.ID == role.ID {
		a.con.note("This is already your target role.")
		return nil
	}
	yes, err := a.con.confirm("Make " + role.Name + " your target role?")
	if err != nil || !yes {
		return err
	}
	a.mutate(func(r *Registry) { r.Career.Role = role.ID })
	a.con.ok("Target role: %s. Your readiness and next steps now follow it.", sty.Bold(role.Name))
	return nil
}

// renderReadiness prints the full readiness report.
func (a *App) renderReadiness() {
	w := a.cols()
	role, ok := a.targetRole()
	a.println("")
	a.println(heading("READINESS REPORT", sty.Green, w))
	if !ok {
		a.con.note("Choose a target role first [1].")
		return
	}
	a.mu.Lock()
	rd := a.reg.readiness(role)
	steps := a.reg.careerSteps(role, rd)
	kit := append([]string(nil), a.reg.Career.Kit...)
	a.mu.Unlock()
	a.printf("  %s\n", sty.Bold(truncate(role.Name, w-4)))
	a.printf("  %s %s\n", bar(int(rd.Total*100), 100, min(20, w-12), sty.Green), sty.Bold(percent(rd.Total)))
	for _, l := range wrap(rd.verdict()+" No app can promise a job; this measures what interviews and hiring managers check.", w-4, "  ") {
		a.println(sty.Gray(l))
	}
	part := func(name string, weight, v float64, detail []string) {
		note := ""
		if w >= 64 {
			note = sty.Gray(fmt.Sprintf("(%d%% of the total)", int(weight*100)))
		}
		a.printf("\n  %s %s %s %s\n", sty.Bold(padRight(name, 19)), bar(int(v*100), 100, min(12, w-30), sty.Green), padRight(percent(v), 5), note)
		for _, d := range detail {
			for _, l := range wrap(d, w-8, "    ") {
				a.println(l)
			}
		}
	}
	part("Foundations", weightFoundations, rd.Foundations, []string{fmt.Sprintf("%d of %d concepts understood in this role's stages.", rd.Concepts, rd.ConceptsTotal)})
	pf := []string{fmt.Sprintf("Shipped (tests passing or graduated): %d. Aim for %d strong projects.", len(rd.Shipped), targetProjects)}
	if len(rd.Shipped) > 0 {
		pf = append(pf, "✓ "+strings.Join(rd.Shipped, ", "))
	}
	if len(rd.Building) > 0 {
		pf = append(pf, "In progress: "+strings.Join(rd.Building, ", "))
	}
	part("Portfolio", weightPortfolio, rd.Portfolio, pf)
	pr := []string{fmt.Sprintf("Coding mocks %d/%d · stories %d/%d", rd.Coding, targetCoding, rd.Stories, targetStories)}
	if role.Design {
		pr[0] += fmt.Sprintf(" · system design mocks %d/%d", rd.Design, targetDesign)
	}
	part("Interview practice", weightPractice, rd.Practice, pr)
	var kd []string
	for _, k := range jobKit {
		mark := sty.Gray("○")
		if contains(kit, k.ID) {
			mark = sty.Green("✓")
		}
		kd = append(kd, mark+" "+k.Name)
	}
	part("Job-search kit", weightKit, rd.Kit, kd)

	a.subhead("What to do next, most useful first", sty.Yellow)
	for i, s := range steps {
		for j, l := range wrap(s, w-9, "") {
			lead := fmt.Sprintf("    %d. ", i+1)
			if j > 0 {
				lead = "       "
			}
			a.println(lead + l)
		}
	}
}

// findBlueprint finds a blueprint by name in any stage.
func findBlueprint(name string) (int, Blueprint, bool) {
	for id := 0; id <= lastStageID(); id++ {
		g, ok := guideFor(id)
		if !ok {
			continue
		}
		for _, b := range g.Blueprints {
			if strings.EqualFold(b.Name, name) {
				return id, b, true
			}
		}
	}
	return 0, Blueprint{}, false
}

// portfolioChecklist is what makes a project count with employers.
var portfolioChecklist = []string{
	"A README that says what it is and why you built it, in the first three lines",
	"How to build and run it in one or two commands",
	"Tests, run automatically on every push (a CI badge)",
	"A diagram or a short description of the design and its trade-offs",
	"Results: a benchmark, a screenshot or a short demo",
	"What you learned, and what you would do next",
	"A clean history of small commits with clear messages",
}

// labState says where the learner is with a named project.
func (a *App) labState(name string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, t := range a.reg.Tracks {
		for _, s := range t.Stages {
			for _, l := range s.Labs {
				if strings.EqualFold(l.Name, name) {
					if l.CompilationStatus == CompileTested || l.Status == StatusGraduated {
						return sty.Green("✓ shipped")
					}
					return sty.Yellow(fmt.Sprintf("◐ building · %.1f h", l.HoursLogged))
				}
			}
		}
	}
	return sty.Gray("○ not started")
}

// portfolio shows the target role's portfolio projects and starts one.
func (a *App) portfolio() error {
	role, ok := a.targetRole()
	if !ok {
		a.con.note("Choose a target role first [1]: each role has its own portfolio projects.")
		return nil
	}
	w := a.cols()
	render := func() {
		a.println("")
		a.println(heading("PORTFOLIO · "+strings.ToUpper(role.Name), sty.Magenta, w))
		for _, l := range wrap("Hiring managers look at what you finished. Two or three complete, well-explained projects beat ten half-done ones. Each project below is a lab blueprint from the curriculum: start it here, log your hours [3] and mark its tests passing [4] when it is done.", w-4, "  ") {
			a.println(sty.Gray(l))
		}
		for i, name := range role.Projects {
			stage, b, found := findBlueprint(name)
			if !found {
				continue
			}
			head := sty.Cyan(fmt.Sprintf("[%d]", i+1)) + " " + sty.Bold(name)
			if state := a.labState(name); visibleLen(head)+visibleLen(state)+4 <= w {
				a.printf("\n  %s  %s\n", head, state)
			} else {
				a.printf("\n  %s\n      %s\n", truncateStyled(head, w-2), state)
			}
			a.printf("      %s\n", sty.Gray(fmt.Sprintf("Stage %d", stage)))
			for _, l := range wrap(b.Brief, w-8, "      ") {
				a.println(l)
			}
			for _, l := range wrap("Milestones: "+strings.Join(b.Milestones, " → "), w-8, "      ") {
				a.println(sty.Gray(l))
			}
		}
		a.subhead("What makes a project count with employers", sty.Yellow)
		for _, c := range portfolioChecklist {
			for i, l := range wrap(c, w-10, "") {
				lead := "    ☐ "
				if i > 0 {
					lead = "      "
				}
				a.println(lead + l)
			}
		}
	}
	a.paged(false, render)
	a.println("")
	s, err := a.con.readLine(promptLabel("Number to start that project as a lab", "q back"))
	if err != nil || s == "" || isCancel(s) {
		return err
	}
	var n int
	if _, err := fmt.Sscan(s, &n); err != nil || n < 1 || n > len(role.Projects) {
		a.con.warn("Choose a project number from 1 to %d.", len(role.Projects))
		return nil
	}
	stage, b, _ := findBlueprint(role.Projects[n-1])
	notes := b.Brief + " Milestones: " + strings.Join(b.Milestones, " → ")
	if len([]rune(notes)) > maxNotesLen {
		notes = b.Brief
	}
	id, err := a.addLab(stage, b.Name, notes, 0)
	if errors.Is(err, errDuplicateLab) {
		a.con.note("You have already started %s. Log hours with [3]; mark its tests passing with [4] when done.", b.Name)
		return nil
	}
	if err != nil {
		return err
	}
	a.con.ok("Started lab #%d %s (Stage %d). Log hours with [3] on the main menu.", id, sty.Bold(b.Name), stage)
	return nil
}

// codingPractice shows the method and the patterns, and runs practice.
func (a *App) codingPractice() error {
	a.paged(false, a.renderInterviewMethod)
	for {
		var names []string
		for _, p := range codingPatterns {
			names = append(names, p.Name)
		}
		a.println("")
		n, err := a.con.promptChoice("Which pattern?", names)
		if err != nil {
			return err
		}
		p := codingPatterns[n]
		a.paged(false, func() { a.renderPattern(p) })
		var drills []string
		for _, d := range p.Drills {
			drills = append(drills, "Practise: "+d.Title)
		}
		a.println("")
		d, err := a.con.promptChoice("Practise a problem?", drills)
		if errors.Is(err, errCancel) {
			continue
		}
		if err != nil {
			return err
		}
		if err := a.codingSession(p, p.Drills[d]); err != nil {
			return err
		}
	}
}

// renderInterviewMethod prints how to work any coding problem.
func (a *App) renderInterviewMethod() {
	w := a.cols()
	a.println("")
	a.println(heading("CODING INTERVIEWS · the method", sty.Cyan, w))
	for _, l := range wrap("Most interview problems are variations on a small number of patterns. Learn to spot the pattern, and follow the same steps every time, out loud.", w-4, "  ") {
		a.println(sty.Gray(l))
	}
	a.println("")
	for i, m := range interviewMethod {
		a.printf("  %s %s\n", sty.Cyan(fmt.Sprintf("%d.", i+1)), sty.Bold(m.Step))
		for _, l := range wrap(m.How, w-8, "     ") {
			a.println(l)
		}
	}
	a.subhead("Habits that make the difference", sty.Yellow)
	for _, t := range interviewTips {
		for i, l := range wrap(t, w-8, "") {
			lead := "    • "
			if i > 0 {
				lead = "      "
			}
			a.println(lead + l)
		}
	}
}

// renderPattern prints one pattern and its practice problems.
func (a *App) renderPattern(p Pattern) {
	w := a.cols()
	a.println("")
	a.println(heading(strings.ToUpper(p.Name), sty.Cyan, w))
	field := func(label string, color func(string) string, text string) {
		a.subhead(label, color)
		for _, l := range wrap(text, w-6, "    ") {
			a.println(l)
		}
	}
	field("When to reach for it", sty.Yellow, p.Signal)
	field("The idea", sty.Cyan, p.Idea)
	field("How", sty.Cyan, strings.Join(p.Steps, " → "))
	field("⚠ Watch out", sty.Red, p.Pitfall)
	field("Usual cost", sty.Gray, p.Cost)
	a.subhead("Practice problems", sty.Green)
	for i, d := range p.Drills {
		for j, l := range wrap(d.Title, w-10, "") {
			lead := "    " + sty.Cyan(fmt.Sprintf("%d.", i+1)) + " "
			if j > 0 {
				lead = "       "
			}
			a.println(lead + l)
		}
	}
}

// Rubrics: honest yes/no checks after a mock interview.
var (
	codingRubric = []string{
		"I asked clarifying questions and worked an example by hand before coding",
		"I said a brute-force approach and its cost before optimising",
		"My solution was correct on my examples and edge cases, traced by hand",
		"I stated the time and space complexity, and it matches the target",
	}
	designRubric = []string{
		"I clarified requirements and scale before drawing anything",
		"I made rough estimates (requests per second, storage)",
		"I defined the API and the data model",
		"I walked one request through a clear high-level design",
		"I went deep on the hardest part and named the trade-offs",
	}
	behavioralRubric = []string{
		"My answer followed Situation, Task, Action, Result",
		"The actions were mine (\"I\"), specific and concrete",
		"The result had evidence (a number or a clear outcome) and a lesson",
		"Spoken aloud, it took about two minutes",
	}
)

// scoreRubric asks each rubric question and returns the points met.
func (a *App) scoreRubric(rubric []string) (int, error) {
	a.println("")
	a.con.note("Score yourself honestly: the point is to find what to practise.")
	score := 0
	for _, q := range rubric {
		yes, err := a.con.confirm(q + "?")
		if err != nil {
			return score, err
		}
		if yes {
			score++
		}
	}
	return score, nil
}

// timed runs the clock for a mock: it shows when to finish and waits for
// Enter, returning the minutes taken.
func (a *App) timed(limit int, what string) (int, error) {
	start := time.Now()
	a.println("")
	a.printf("  %s started at %s · aim to finish by %s (%d min)\n", sty.Bold(what), start.Format("15:04"),
		start.Add(time.Duration(limit)*time.Minute).Format("15:04"), limit)
	_, err := a.con.readLine(promptLabel("Work on paper or in a plain editor; press Enter when you are done", "q give up"))
	return int(time.Since(start).Minutes() + 0.5), err
}

// codingSession presents one problem with its hint ladder and records it
// as a coding mock when the learner scores it.
func (a *App) codingSession(p Pattern, d Drill) error {
	w := a.cols()
	a.println("")
	a.println(heading("PROBLEM · "+strings.ToUpper(d.Title), sty.Yellow, w))
	for _, l := range wrap(d.Task, w-4, "  ") {
		a.println(l)
	}
	a.println("")
	a.printf("  %s\n", sty.Bold("Example"))
	for _, line := range strings.Split(d.Examples, "\n") {
		for _, l := range wrap(line, w-6, "    ") {
			a.println(sty.Green(l))
		}
	}
	for _, l := range wrap("Think first: follow the method and give it a real try (15–20 minutes) before taking a hint. Being stuck for a while is how the pattern sticks.", w-4, "  ") {
		a.println(sty.Gray(l))
	}
	start := time.Now()
	hints := 0
	for hints < len(d.Hints) {
		s, err := a.con.readLine(promptLabel(fmt.Sprintf("Enter for hint %d of %d · d when you are done", hints+1, len(d.Hints)), "q stop"))
		if err != nil {
			return err
		}
		if isCancel(s) {
			return nil
		}
		if strings.EqualFold(s, "d") {
			break
		}
		for _, l := range wrap(fmt.Sprintf("Hint %d: %s", hints+1, d.Hints[hints]), w-6, "    ") {
			a.println(sty.Yellow(l))
		}
		hints++
	}
	if hints == len(d.Hints) {
		if _, err := a.con.readLine(promptLabel("No more hints: finish it, then press Enter", "q stop")); err != nil {
			return err
		}
	}
	a.printf("\n  %s %s\n", sty.Bold("Aim for:"), d.Target)
	minutes := int(time.Since(start).Minutes() + 0.5)
	score, err := a.scoreRubric(codingRubric)
	if err != nil {
		return err
	}
	a.mutate(func(r *Registry) {
		r.Career.Mocks = append(r.Career.Mocks, MockRecord{Kind: MockCoding, Topic: p.Name + ": " + d.Title, At: time.Now(), Minutes: minutes, Score: score, Of: len(codingRubric)})
	})
	note := ""
	if hints > 0 {
		note = fmt.Sprintf(" You used %d hint%s: try a problem of the same pattern again in a few days without them.", hints, plural(hints))
	}
	a.con.ok("Recorded: %d/%d on the rubric, %d min.%s", score, len(codingRubric), minutes, note)
	return nil
}

// systemDesign shows the framework, the building blocks and the classic
// questions.
func (a *App) systemDesign() error {
	a.paged(false, a.renderDesignBasics)
	for {
		var names []string
		for _, p := range designPrompts {
			names = append(names, p.Name)
		}
		a.println("")
		n, err := a.con.promptChoice("Study which question?", names)
		if err != nil {
			return err
		}
		p := designPrompts[n]
		a.paged(false, func() { a.renderDesignPrompt(p, true) })
		a.println("")
		yes, err := a.con.confirm("Try it as a timed mock interview (45 min)?")
		if err != nil {
			return err
		}
		if yes {
			if err := a.designMock(p); err != nil {
				return err
			}
		}
	}
}

// renderDesignBasics prints the framework, building blocks and numbers.
func (a *App) renderDesignBasics() {
	w := a.cols()
	a.println("")
	a.println(heading("SYSTEM DESIGN · the framework", sty.Blue, w))
	for _, l := range wrap("System design rounds have no single right answer. Interviewers watch how you clarify, estimate, structure and weigh trade-offs. Use the same steps every time, and keep an eye on the clock.", w-4, "  ") {
		a.println(sty.Gray(l))
	}
	a.println("")
	for i, s := range designSteps {
		a.printf("  %s %s %s\n", sty.Cyan(fmt.Sprintf("%d.", i+1)), sty.Bold(s.Step), sty.Gray("("+s.Minutes+" min)"))
		for _, l := range wrap(s.How, w-8, "     ") {
			a.println(l)
		}
	}
	a.subhead("Building blocks", sty.Yellow)
	for _, b := range designBlocks {
		for i, l := range wrap(sty.Bold(b.Name)+": "+b.What, w-8, "") {
			lead := "    • "
			if i > 0 {
				lead = "      "
			}
			a.println(lead + l)
		}
	}
	a.subhead("Numbers to know", sty.Yellow)
	for _, n := range latencyNumbers {
		if w >= 60 {
			a.printf("    %s %s\n", padRight(n.What, 36), sty.Cyan(n.Time))
			continue
		}
		for _, l := range wrap(n.What+": "+n.Time, w-6, "    ") {
			a.println(l)
		}
	}
}

// renderDesignPrompt prints one design question; with answers, also what
// a strong design covers.
func (a *App) renderDesignPrompt(p DesignPrompt, answers bool) {
	w := a.cols()
	a.println("")
	a.println(heading("DESIGN · "+strings.ToUpper(p.Name), sty.Blue, w))
	for _, l := range wrap(p.Prompt, w-4, "  ") {
		a.println(sty.Bold(l))
	}
	list := func(title string, color func(string) string, items []string) {
		a.subhead(title, color)
		for _, it := range items {
			for i, l := range wrap(it, w-8, "") {
				lead := "    • "
				if i > 0 {
					lead = "      "
				}
				a.println(lead + l)
			}
		}
	}
	list("Questions worth asking first", sty.Yellow, p.Ask)
	if !answers {
		return
	}
	list("What a strong design covers", sty.Green, p.Core)
	list("Likely deep dives", sty.Cyan, p.Deep)
	a.subhead("⚠ The usual mistake", sty.Red)
	for _, l := range wrap(p.Trap, w-6, "    ") {
		a.println(l)
	}
	var stages []string
	for _, id := range p.Stages {
		stages = append(stages, fmt.Sprint(id))
	}
	a.println("")
	a.grayWrap("The academy teaches the pieces in Stage "+strings.Join(stages, " and ")+".", "  ")
}

// designMock runs a timed design interview on a question.
func (a *App) designMock(p DesignPrompt) error {
	a.paged(false, func() { a.renderDesignPrompt(p, false) })
	for _, l := range wrap("Draw boxes and arrows on paper and talk through each step of the framework aloud, as if an interviewer were listening. The full checklist appears when you finish.", a.cols()-4, "  ") {
		a.println(sty.Gray(l))
	}
	minutes, err := a.timed(45, "System design: "+p.Name)
	if err != nil {
		return err
	}
	a.paged(false, func() { a.renderDesignPrompt(p, true) })
	score, err := a.scoreRubric(designRubric)
	if err != nil {
		return err
	}
	a.mutate(func(r *Registry) {
		r.Career.Mocks = append(r.Career.Mocks, MockRecord{Kind: MockDesign, Topic: p.Name, At: time.Now(), Minutes: minutes, Score: score, Of: len(designRubric)})
	})
	a.con.ok("Recorded: %d/%d on the rubric, %d min.", score, len(designRubric), minutes)
	return nil
}

// renderStoryBank prints the STAR guide and which themes have stories.
func (a *App) renderStoryBank() {
	w := a.cols()
	a.mu.Lock()
	stories := append([]Story(nil), a.reg.Career.Stories...)
	a.mu.Unlock()
	a.println("")
	a.println(heading("BEHAVIORAL INTERVIEWS · your story bank", sty.Yellow, w))
	for _, l := range wrap("Every interview loop has questions like \"tell me about a time…\". They test how you work with people and problems. Answer with a story in four parts:", w-4, "  ") {
		a.println(sty.Gray(l))
	}
	for _, s := range starGuide {
		for i, l := range wrap(sty.Bold(s.Part)+": "+s.How, w-8, "") {
			lead := "    • "
			if i > 0 {
				lead = "      "
			}
			a.println(lead + l)
		}
	}
	a.println("")
	for _, t := range behavioralTips {
		for _, l := range wrap(t, w-6, "    ") {
			a.println(sty.Gray(l))
		}
	}
	a.subhead("The questions, and your stories for them", sty.Yellow)
	for _, t := range behavioralThemes {
		count := 0
		for _, s := range stories {
			if s.Theme == t.ID {
				count++
			}
		}
		mark := sty.Gray("○")
		if count > 0 {
			mark = sty.Green("✓")
		}
		a.printf("    %s %s\n", mark, sty.Bold(t.Name))
		for _, l := range wrap("\""+t.Question+"\" They look for: "+t.Tests, w-10, "      ") {
			a.println(sty.Gray(l))
		}
	}
	a.subhead(fmt.Sprintf("Your stories (%d)", len(stories)), sty.Yellow)
	if len(stories) == 0 {
		a.grayWrap("None yet. Press a to write your first one.", "    ")
	}
	for i, s := range stories {
		a.printf("    %s %s %s\n", sty.Cyan(fmt.Sprintf("%d.", i+1)), sty.Bold(truncate(s.Title, w-30)), sty.Gray("· "+themeName(s.Theme)))
	}
}

// storyBank is the behavioral screen: read, add, view and delete stories.
func (a *App) storyBank() error {
	for {
		a.paged(false, a.renderStoryBank)
		a.println("")
		s, err := a.con.readLine(promptLabel("a add a story · number to read or delete one", "q back"))
		if err != nil || s == "" || isCancel(s) {
			return err
		}
		if strings.EqualFold(s, "a") {
			if err := a.addStory(""); err != nil && !errors.Is(err, errCancel) {
				return err
			}
			continue
		}
		var n int
		a.mu.Lock()
		total := len(a.reg.Career.Stories)
		a.mu.Unlock()
		if _, err := fmt.Sscan(s, &n); err != nil || n < 1 || n > total {
			a.con.warn("Type a, or a story number from 1 to %d.", total)
			continue
		}
		if err := a.showStory(n - 1); err != nil {
			return err
		}
	}
}

// addStory asks for a STAR story; theme may be preset.
func (a *App) addStory(theme string) error {
	if theme == "" {
		var names []string
		for _, t := range behavioralThemes {
			names = append(names, t.Name+sty.Gray(" · "+t.Question))
		}
		n, err := a.con.promptChoice("Which question does it answer?", names)
		if err != nil {
			return err
		}
		theme = behavioralThemes[n].ID
	}
	a.con.note("Keep each part short: a sentence or two. :q cancels.")
	ask := func(label string) (string, error) { return a.con.promptText(label, maxNotesLen, true) }
	var st Story
	var err error
	if st.Title, err = a.con.promptText("A short title (e.g. \"The cache that lost writes\")", maxNameLen, true); err != nil {
		return err
	}
	for _, f := range []struct {
		label string
		dst   *string
	}{{"Situation", &st.Situation}, {"Task (your responsibility)", &st.Task}, {"Action (what you did)", &st.Action}, {"Result (and what you learned)", &st.Result}} {
		if *f.dst, err = ask(f.label); err != nil {
			return err
		}
	}
	st.Theme, st.At = theme, time.Now()
	a.mutate(func(r *Registry) { r.Career.Stories = append(r.Career.Stories, st) })
	a.con.ok("Saved \"%s\" to your story bank. Say it aloud a few times this week.", st.Title)
	return nil
}

// showStory prints a story and offers to delete it.
func (a *App) showStory(i int) error {
	a.mu.Lock()
	st := a.reg.Career.Stories[i]
	a.mu.Unlock()
	w := a.cols()
	a.println("")
	a.printf("  %s %s\n", sty.Bold(sty.Yellow(st.Title)), sty.Gray("· "+themeName(st.Theme)))
	for _, p := range [][2]string{{"Situation", st.Situation}, {"Task", st.Task}, {"Action", st.Action}, {"Result", st.Result}} {
		for j, l := range wrap(sty.Bold(p[0])+": "+p[1], w-6, "") {
			lead := "    "
			if j > 0 {
				lead = "      "
			}
			a.println(lead + l)
		}
	}
	a.println("")
	del, err := a.con.confirm("Delete this story?")
	if err != nil || !del {
		return err
	}
	a.mutate(func(r *Registry) {
		if i < len(r.Career.Stories) {
			r.Career.Stories = append(r.Career.Stories[:i], r.Career.Stories[i+1:]...)
		}
	})
	a.con.ok("Deleted.")
	return nil
}

// mockInterview runs a timed mock of the chosen kind.
func (a *App) mockInterview() error {
	a.mu.Lock()
	mocks := append([]MockRecord(nil), a.reg.Career.Mocks...)
	a.mu.Unlock()
	a.println("")
	a.println(heading("MOCK INTERVIEW", sty.Yellow, a.cols()))
	for _, l := range wrap("Interviews are a skill that improves with practice under realistic conditions: a clock, no help, and talking aloud. Better still, ask a friend to play the interviewer and read the problem to you.", a.cols()-4, "  ") {
		a.println(sty.Gray(l))
	}
	if len(mocks) > 0 {
		a.printf("\n  %s\n", sty.Bold("Your recent mocks"))
		for _, m := range mocks[max(0, len(mocks)-5):] {
			a.printf("    %s %s %s %s\n", sty.Gray(m.At.Format("Jan 2")), padRight(m.Kind, 10),
				sty.Cyan(fmt.Sprintf("%d/%d", m.Score, m.Of)), truncate(m.Topic, a.cols()-30))
		}
	}
	a.println("")
	n, err := a.con.promptChoice("Which kind?", []string{
		"Coding (a random problem, 45 min)",
		"System design (a random question, 45 min)",
		"Behavioral (a random question, about 2 min aloud)",
	})
	if err != nil {
		return err
	}
	switch n {
	case 0:
		p, d := pickDrill(mocks, rand.IntN)
		a.con.note("Pattern hidden: spotting it is part of the test. Timer starts now.")
		hidden := Pattern{Name: p.Name}
		return a.codingSession(hidden, d)
	case 1:
		p := designPrompts[rand.IntN(len(designPrompts))]
		return a.designMock(p)
	default:
		return a.behavioralMock(behavioralThemes[rand.IntN(len(behavioralThemes))])
	}
}

// pickDrill chooses a coding problem, preferring ones not yet practised.
func pickDrill(mocks []MockRecord, intn func(int) int) (Pattern, Drill) {
	done := map[string]bool{}
	for _, m := range mocks {
		done[m.Topic] = true
	}
	type pick struct{ p, d int }
	var fresh, all []pick
	for pi, p := range codingPatterns {
		for di, d := range p.Drills {
			all = append(all, pick{pi, di})
			if !done[p.Name+": "+d.Title] {
				fresh = append(fresh, pick{pi, di})
			}
		}
	}
	if len(fresh) == 0 {
		fresh = all
	}
	c := fresh[intn(len(fresh))]
	return codingPatterns[c.p], codingPatterns[c.p].Drills[c.d]
}

// behavioralMock asks one behavioral question and offers to save the
// answer as a story.
func (a *App) behavioralMock(t BehavioralTheme) error {
	a.println("")
	a.printf("  %s\n", sty.Bold(sty.Yellow("\""+t.Question+"\"")))
	a.println(sty.Gray("    They look for: " + t.Tests))
	minutes, err := a.timed(3, "Behavioral answer")
	if err != nil {
		return err
	}
	score, err := a.scoreRubric(behavioralRubric)
	if err != nil {
		return err
	}
	a.mutate(func(r *Registry) {
		r.Career.Mocks = append(r.Career.Mocks, MockRecord{Kind: MockBehavioral, Topic: t.Name, At: time.Now(), Minutes: minutes, Score: score, Of: len(behavioralRubric)})
	})
	a.con.ok("Recorded: %d/%d on the rubric.", score, len(behavioralRubric))
	yes, err := a.con.confirm("Write this story down for your story bank?")
	if err != nil || !yes {
		return err
	}
	return a.addStory(t.ID)
}

// resumeBuilder shows the rules and writes a bullet for a project.
func (a *App) resumeBuilder() error {
	type proj struct {
		name  string
		stage int
	}
	var projects []proj
	a.mu.Lock()
	for _, t := range a.reg.Tracks {
		for _, s := range t.Stages {
			for _, l := range s.Labs {
				projects = append(projects, proj{l.Name, s.ID})
			}
		}
	}
	bullets := map[string]string{}
	for k, v := range a.reg.Career.Bullets {
		bullets[k] = v
	}
	a.mu.Unlock()
	w := a.cols()
	a.paged(false, func() {
		a.println("")
		a.println(heading("RESUME BUILDER", sty.Magenta, w))
		for _, r := range resumeRules {
			for i, l := range wrap(r, w-8, "") {
				lead := "    • "
				if i > 0 {
					lead = "      "
				}
				a.println(lead + l)
			}
		}
		a.subhead("Weak and strong bullets", sty.Yellow)
		for _, e := range resumeExamples {
			for _, l := range wrap("✗ "+e.Weak, w-8, "    ") {
				a.println(sty.Red(l))
			}
			for _, l := range wrap("✓ "+e.Strong, w-8, "    ") {
				a.println(sty.Green(l))
			}
			a.println("")
		}
		a.printf("  %s\n", sty.Bold(sty.Yellow("Your projects")))
		if len(projects) == 0 {
			a.grayWrap("No labs yet: start a portfolio project [3] first.", "    ")
		}
		for i, p := range projects {
			for j, l := range wrap(p.name, w-10, "") {
				lead := "    " + sty.Cyan(fmt.Sprintf("%d.", i+1)) + " "
				if j > 0 {
					lead = "       "
				}
				a.println(lead + sty.Bold(l))
			}
			if b := bullets[p.name]; b != "" {
				for _, l := range wrap(b, w-10, "       ") {
					a.println(sty.Green(l))
				}
			} else {
				a.println(sty.Gray("       no bullet yet"))
			}
		}
	})
	if len(projects) == 0 {
		return nil
	}
	a.println("")
	s, err := a.con.readLine(promptLabel("Number of a project to write its bullet", "q back"))
	if err != nil || s == "" || isCancel(s) {
		return err
	}
	var n int
	if _, err := fmt.Sscan(s, &n); err != nil || n < 1 || n > len(projects) {
		a.con.warn("Choose a project number from 1 to %d.", len(projects))
		return nil
	}
	p := projects[n-1]
	if _, b, ok := findBlueprint(p.name); ok {
		for _, l := range wrap("The project: "+b.Brief, w-6, "    ") {
			a.println(sty.Gray(l))
		}
	}
	for _, l := range wrap("Formula: what you built + how (technology and the key idea) + the result, with an honest number. Start with a verb.", w-6, "    ") {
		a.println(sty.Yellow(l))
	}
	text, err := a.con.promptText("Bullet (Enter keeps the current one)", maxNotesLen, false)
	if err != nil || text == "" {
		return err
	}
	a.mutate(func(r *Registry) {
		if r.Career.Bullets == nil {
			r.Career.Bullets = map[string]string{}
		}
		r.Career.Bullets[p.name] = text
	})
	a.con.ok("Saved. Export your career kit [10] to put it on your resume.")
	return nil
}

// jobSearch is the kit checklist, the applications tracker and the tips.
func (a *App) jobSearch() error {
	for {
		a.paged(false, a.renderJobSearch)
		a.println("")
		s, err := a.con.readLine(promptLabel("a add application · number to update one · k1–k7 tick a kit item", "q back"))
		if err != nil || s == "" || isCancel(s) {
			return err
		}
		s = strings.ToLower(strings.TrimSpace(s))
		switch {
		case s == "a":
			if err := a.addApplication(); err != nil && !errors.Is(err, errCancel) {
				return err
			}
		case strings.HasPrefix(s, "k"):
			var n int
			if _, err := fmt.Sscan(s[1:], &n); err != nil || n < 1 || n > len(jobKit) {
				a.con.warn("Type k and a kit number from 1 to %d, such as k1.", len(jobKit))
				continue
			}
			id := jobKit[n-1].ID
			a.mutate(func(r *Registry) { r.Career.Kit = setMember(r.Career.Kit, id, !contains(r.Career.Kit, id)) })
		default:
			var n int
			a.mu.Lock()
			total := len(a.reg.Career.Applications)
			a.mu.Unlock()
			if _, err := fmt.Sscan(s, &n); err != nil || n < 1 || n > total {
				a.con.warn("Type a, k1–k%d, or an application number.", len(jobKit))
				continue
			}
			if err := a.updateApplication(n - 1); err != nil && !errors.Is(err, errCancel) {
				return err
			}
		}
	}
}

// renderJobSearch prints the kit, the tracker and the tips.
func (a *App) renderJobSearch() {
	w := a.cols()
	a.mu.Lock()
	kit := append([]string(nil), a.reg.Career.Kit...)
	apps := append([]Application(nil), a.reg.Career.Applications...)
	a.mu.Unlock()
	a.println("")
	a.println(heading("JOB SEARCH", sty.Green, w))
	a.printf("  %s\n", sty.Bold(sty.Yellow("Your job-search kit")))
	for i, k := range jobKit {
		mark := sty.Gray("○")
		if contains(kit, k.ID) {
			mark = sty.Green("✓")
		}
		for j, l := range wrap(k.Name, w-12, "") {
			if j == 0 {
				a.printf("    %s %s %s\n", sty.Cyan(fmt.Sprintf("k%d", i+1)), mark, sty.Bold(l))
			} else {
				a.printf("         %s\n", sty.Bold(l))
			}
		}
		for _, l := range wrap(k.How, w-10, "         ") {
			a.println(sty.Gray(l))
		}
	}
	a.subhead(fmt.Sprintf("Applications (%d)", len(apps)), sty.Yellow)
	if len(apps) == 0 {
		a.grayWrap("None yet. Press a to add one when you apply.", "    ")
	} else {
		counts := map[string]int{}
		for _, ap := range apps {
			counts[ap.Stage]++
		}
		var parts []string
		for _, st := range appStages {
			if counts[st] > 0 {
				parts = append(parts, fmt.Sprintf("%s %d", st, counts[st]))
			}
		}
		a.println(sty.Gray("    " + strings.Join(parts, " · ")))
		if rate, ok := responseRate(apps); ok {
			a.println(sty.Gray("    Response rate: " + percent(rate) + " of applications reached a screen or further."))
		}
		for i, ap := range apps {
			a.printf("    %s %s %s %s\n", sty.Cyan(fmt.Sprintf("%d.", i+1)), sty.Bold(truncate(ap.Company, 24)),
				sty.Gray(truncate(ap.Role, max(10, w-50))), sty.Yellow(ap.Stage))
			if ap.Next != "" {
				for _, l := range wrap("next: "+ap.Next, w-12, "       ") {
					a.println(sty.Gray(l))
				}
			}
		}
	}
	a.subhead("How to run the search", sty.Yellow)
	for _, t := range jobSearchTips {
		for i, l := range wrap(sty.Bold(t.Title)+": "+t.How, w-8, "") {
			lead := "    • "
			if i > 0 {
				lead = "      "
			}
			a.println(lead + l)
		}
	}
}

// responseRate is the share of submitted applications that reached a
// screen or further.
func responseRate(apps []Application) (float64, bool) {
	sent, heard := 0, 0
	for _, ap := range apps {
		switch ap.Stage {
		case "Saved":
			continue
		case "Screen", "Assessment", "Interviews", "Offer":
			heard++
		}
		sent++
	}
	if sent == 0 {
		return 0, false
	}
	return float64(heard) / float64(sent), true
}

// addApplication adds an application to the tracker.
func (a *App) addApplication() error {
	company, err := a.con.promptText("Company", maxNameLen, true)
	if err != nil {
		return err
	}
	role, err := a.con.promptText("Role (job title)", maxNameLen, true)
	if err != nil {
		return err
	}
	next, err := a.con.promptText("Next step (e.g. \"follow up on the 14th\"; Enter to skip)", maxNotesLen, false)
	if err != nil {
		return err
	}
	now := time.Now()
	a.mutate(func(r *Registry) {
		r.Career.Applications = append(r.Career.Applications, Application{Company: company, Role: role, Stage: "Applied", Next: next, Applied: now, Updated: now})
	})
	a.con.ok("Added %s · %s.", company, role)
	return nil
}

// updateApplication moves an application to a new stage.
func (a *App) updateApplication(i int) error {
	a.mu.Lock()
	ap := a.reg.Career.Applications[i]
	a.mu.Unlock()
	a.printf("\n  %s · %s · now %s\n", sty.Bold(ap.Company), ap.Role, sty.Yellow(ap.Stage))
	n, err := a.con.promptChoice("New stage", appStages)
	if err != nil {
		return err
	}
	next, err := a.con.promptText("Next step (Enter to keep: "+truncate(ap.Next, 30)+")", maxNotesLen, false)
	if err != nil {
		return err
	}
	a.mutate(func(r *Registry) {
		if i >= len(r.Career.Applications) {
			return
		}
		p := &r.Career.Applications[i]
		p.Stage, p.Updated = appStages[n], time.Now()
		if next != "" {
			p.Next = next
		}
	})
	switch appStages[n] {
	case "Offer":
		a.con.ok("Congratulations! Ask for it in writing and for time to decide. See the tips on offers below.")
	case "Rejected":
		a.con.note("It happens to everyone. If you can, ask what to improve; then keep the pipeline full.")
	default:
		a.con.ok("Updated.")
	}
	return nil
}

// exportCareer writes the career kit as Markdown next to the registry.
func (a *App) exportCareer() error {
	a.mu.Lock()
	md := a.reg.careerMarkdown(time.Now())
	a.mu.Unlock()
	path := filepath.Join(filepath.Dir(a.path), "academy_career.md")
	if err := atomicWriteFile(path, []byte(md)); err != nil {
		return fmt.Errorf("export: %w", err)
	}
	a.con.ok("Exported your career kit to %s", path)
	a.con.note("Copy the bullets into your resume, the stories into your interview notes, and keep it with your applications.")
	return nil
}
