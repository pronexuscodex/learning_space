package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every role points at real stages and real lab blueprints, so the
// portfolio and readiness always work.
func TestCareerContentIsComplete(t *testing.T) {
	ids := map[string]bool{}
	for _, r := range roles {
		if ids[r.ID] {
			t.Errorf("duplicate role ID %q", r.ID)
		}
		ids[r.ID] = true
		if r.Name == "" || r.Does == "" || r.Hires == "" || r.Advice == "" || len(r.Titles) == 0 || len(r.Screens) == 0 || len(r.Loop) == 0 || len(r.Keywords) == 0 {
			t.Errorf("%s: a field is empty", r.ID)
		}
		for _, s := range r.Stages {
			if _, ok := guideFor(s); !ok {
				t.Errorf("%s: unknown stage %d", r.ID, s)
			}
		}
		if len(r.Projects) < targetProjects {
			t.Errorf("%s: needs at least %d portfolio projects", r.ID, targetProjects)
		}
		for _, p := range r.Projects {
			if _, _, ok := findBlueprint(p); !ok {
				t.Errorf("%s: project %q is not a lab blueprint", r.ID, p)
			}
		}
	}
	if len(codingPatterns) < 12 {
		t.Errorf("expected at least 12 coding patterns, got %d", len(codingPatterns))
	}
	for _, p := range codingPatterns {
		if p.Signal == "" || p.Idea == "" || p.Pitfall == "" || p.Cost == "" || len(p.Steps) == 0 || len(p.Drills) < 2 {
			t.Errorf("pattern %q is incomplete", p.Name)
		}
		for _, d := range p.Drills {
			if d.Task == "" || d.Examples == "" || d.Target == "" || len(d.Hints) != 3 {
				t.Errorf("%s / %s: needs a task, an example, a target and three hints", p.Name, d.Title)
			}
		}
	}
	for _, p := range designPrompts {
		if p.Prompt == "" || p.Trap == "" || len(p.Ask) == 0 || len(p.Core) == 0 || len(p.Deep) == 0 || len(p.Stages) == 0 {
			t.Errorf("design question %q is incomplete", p.Name)
		}
		for _, s := range p.Stages {
			if _, ok := guideFor(s); !ok {
				t.Errorf("%s: unknown stage %d", p.Name, s)
			}
		}
	}
	themes := map[string]bool{}
	for _, th := range behavioralThemes {
		if themes[th.ID] || th.Question == "" || th.Tests == "" {
			t.Errorf("behavioral theme %q is duplicated or incomplete", th.ID)
		}
		themes[th.ID] = true
	}
	kit := map[string]bool{}
	for _, k := range jobKit {
		if kit[k.ID] || k.How == "" {
			t.Errorf("kit item %q is duplicated or incomplete", k.ID)
		}
		kit[k.ID] = true
	}
}

func TestReadiness(t *testing.T) {
	r := seedRegistry()
	role, _ := roleByID("backend")
	rd := r.readiness(role)
	if rd.Total != 0 || rd.ConceptsTotal == 0 {
		t.Fatalf("a new learner starts at 0%% with concepts to learn: %+v", rd)
	}
	steps := r.careerSteps(role, rd)
	if len(steps) == 0 || !strings.Contains(steps[0], "Learn") {
		t.Fatalf("the foundations weigh most, so they come first: %v", steps)
	}

	// Understand every concept of the role's stages.
	for _, id := range role.Stages {
		g, _ := guideFor(id)
		_, s := r.findStage(id)
		for _, c := range g.Concepts {
			s.setStudied(c.Name, true)
		}
	}
	// Ship one role project, start another, ship one of your own.
	_, s9 := r.findStage(9)
	s9.Labs = append(s9.Labs,
		Lab{ID: 1, Name: "Production-Ready To-Do API", CompilationStatus: CompileTested, Status: StatusActive},
		Lab{ID: 2, Name: "Open-Source Contribution", CompilationStatus: CompileOK, Status: StatusActive},
		Lab{ID: 3, Name: "My own idea", CompilationStatus: CompileNotBuilt, Status: StatusGraduated})
	for i := 0; i < targetCoding; i++ {
		r.Career.Mocks = append(r.Career.Mocks, MockRecord{Kind: MockCoding, Of: 4})
	}
	r.Career.Kit = []string{"resume", "github"}
	rd = r.readiness(role)
	if rd.Foundations != 1 {
		t.Errorf("foundations = %v, want 1", rd.Foundations)
	}
	if want := 1.5 / targetProjects; rd.Portfolio != want || len(rd.Shipped) != 2 || len(rd.Building) != 1 {
		t.Errorf("portfolio = %v (shipped %v, building %v), want %v", rd.Portfolio, rd.Shipped, rd.Building, want)
	}
	if rd.Practice < 0.33 || rd.Practice > 0.34 { // coding done; stories and design not yet
		t.Errorf("practice = %v, want 1/3", rd.Practice)
	}
	want := weightFoundations + weightPortfolio*0.5 + weightPractice/3 + weightKit*2/float64(len(jobKit))
	if diff := rd.Total - want; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("total = %v, want %v", rd.Total, want)
	}
	// Practice now lacks the most (0.20 × 2/3 > 0.25 × 1/2): a story first,
	// then finishing the project already started.
	steps = r.careerSteps(role, rd)
	if len(steps) < 2 || !strings.Contains(steps[0], "STAR") || !strings.Contains(steps[1], "Open-Source Contribution") {
		t.Errorf("the weakest part comes first: %v", steps)
	}
	if rd.verdict() == "" {
		t.Error("there is always a verdict")
	}
}

func TestCareerDataIsRepaired(t *testing.T) {
	r := seedRegistry()
	r.Career = Career{
		Role:         "astronaut",
		Kit:          []string{"resume", "resume", "unknown"},
		Stories:      []Story{{Title: " "}, {Title: "Kept", Theme: "hard"}},
		Applications: []Application{{Company: ""}, {Company: "Acme", Stage: "Ghosted"}},
		Mocks:        []MockRecord{{Kind: "karaoke", Of: 4}, {Kind: MockCoding, Score: 9, Of: 4, Minutes: -3}},
		Bullets:      map[string]string{"x": " ", "y": "Built y"},
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var back Registry
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if err := back.validate(); err != nil {
		t.Fatal(err)
	}
	c := back.Career
	if c.Role != "" || len(c.Kit) != 1 || len(c.Stories) != 1 || len(c.Applications) != 1 || c.Applications[0].Stage != "Applied" {
		t.Errorf("not repaired: %+v", c)
	}
	if len(c.Mocks) != 1 || c.Mocks[0].Score != 4 || c.Mocks[0].Minutes != 0 || len(c.Bullets) != 1 {
		t.Errorf("mocks or bullets not repaired: %+v %v", c.Mocks, c.Bullets)
	}
}

func TestCareerHelpers(t *testing.T) {
	if _, ok := responseRate(nil); ok {
		t.Error("no applications: no rate")
	}
	rate, _ := responseRate([]Application{{Stage: "Saved"}, {Stage: "Applied"}, {Stage: "Screen"}, {Stage: "Rejected"}, {Stage: "Offer"}})
	if rate != 0.5 {
		t.Errorf("rate = %v, want 0.5 (2 of 4 sent; saved ones are not sent)", rate)
	}
	// pickDrill avoids problems already practised.
	var mocks []MockRecord
	for _, p := range codingPatterns {
		for _, d := range p.Drills {
			mocks = append(mocks, MockRecord{Kind: MockCoding, Topic: p.Name + ": " + d.Title})
		}
	}
	last := mocks[len(mocks)-1]
	mocks = mocks[:len(mocks)-1]
	p, d := pickDrill(mocks, func(int) int { return 0 })
	if p.Name+": "+d.Title != last.Topic {
		t.Errorf("the only fresh problem must be picked, got %s: %s", p.Name, d.Title)
	}
	if mdCell("a|b\nc") != `a\|b c` {
		t.Error("table cells must escape pipes and newlines")
	}
}

// The career center from start to finish, with typed answers.
func TestCareerCenterScreens(t *testing.T) {
	input := strings.Join([]string{
		"1", "1", "y", // choose the backend role
		"2",      // readiness report
		"3", "1", // start the first portfolio project
		"4", "1", "1", "", "d", "y", "y", "n", "y", "q", // a coding problem: one hint, then the rubric
		"6", "a", "1", "Learning C fast", "New job, new language", "Ship a tool in C", "Read K&R, built small programs daily", "Shipped in two weeks", "q",
		"9", "a", "Acme", "Backend Engineer", "", "k1", "1", "3", "", "q", // job search: an application, a kit item, an update
		"8", "1", "Built a REST API in Go with tests", // resume bullet
		"10", // export
		"11", // back
	}, "\n") + "\n"
	dir := t.TempDir()
	var out bytes.Buffer
	a := &App{reg: seedRegistry(), path: filepath.Join(dir, "reg.json"), width: 80,
		con: &console{in: bufio.NewReader(strings.NewReader(input)), out: &out}}
	if err := a.careerCenter(); err != nil {
		t.Fatalf("career center: %v\n%s", err, out.String())
	}
	c := a.reg.Career
	if c.Role != "backend" {
		t.Errorf("role = %q", c.Role)
	}
	if len(c.Mocks) != 1 || c.Mocks[0].Score != 3 || c.Mocks[0].Kind != MockCoding {
		t.Errorf("mocks = %+v", c.Mocks)
	}
	if len(c.Stories) != 1 || c.Stories[0].Theme != behavioralThemes[0].ID {
		t.Errorf("stories = %+v", c.Stories)
	}
	if len(c.Applications) != 1 || c.Applications[0].Stage != "Screen" || !contains(c.Kit, "resume") {
		t.Errorf("applications = %+v, kit = %v", c.Applications, c.Kit)
	}
	if c.Bullets["Production-Ready To-Do API"] == "" {
		t.Errorf("bullets = %v", c.Bullets)
	}
	if _, l := a.reg.findLab(1); l == nil || l.Name != "Production-Ready To-Do API" {
		t.Error("the portfolio project must be started as a lab")
	}
	md, err := os.ReadFile(filepath.Join(dir, "academy_career.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Target role: Backend Software Engineer", "Built a REST API in Go with tests", "Learning C fast", "| Acme |"} {
		if !strings.Contains(string(md), want) {
			t.Errorf("export is missing %q", want)
		}
	}
	shown := stripANSI(out.String())
	for _, want := range []string{"GET HIRED", "READINESS REPORT", "PORTFOLIO", "CODING INTERVIEWS", "Hint 1:", "JOB SEARCH", "RESUME BUILDER"} {
		if !strings.Contains(shown, want) {
			t.Errorf("screen %q was not shown", want)
		}
	}
}

// Every career screen fits the narrowest supported terminal.
func TestCareerScreensFit(t *testing.T) {
	r := seedRegistry()
	r.Career.Role = "security"
	for _, cols := range []int{40, 80} {
		var out bytes.Buffer
		a := &App{reg: r, width: cols, con: &console{out: &out}}
		a.renderCareerHome()
		a.renderReadiness()
		a.renderInterviewMethod()
		a.renderDesignBasics()
		a.renderStoryBank()
		a.renderJobSearch()
		for _, role := range roles {
			a.renderRole(role)
		}
		for _, p := range codingPatterns {
			a.renderPattern(p)
		}
		for _, p := range designPrompts {
			a.renderDesignPrompt(p, true)
		}
		for _, line := range strings.Split(out.String(), "\n") {
			if w := visibleLen(line); w > cols {
				t.Errorf("%d columns: line is %d wide: %q", cols, w, stripANSI(line))
			}
		}
	}
}
