package main

import (
	"strings"
	"testing"
	"time"
)

func TestNextStepsFreshCampus(t *testing.T) {
	reg := seedRegistry()
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.Local)
	steps := reg.nextSteps(now)
	if len(steps) != 2 || steps[0].Kind != SuggestConcept || steps[0].Stage != 0 || steps[0].Concept != 0 {
		t.Fatalf("fresh campus should start with Stage 0's first concept, got %+v", steps)
	}
	if steps[1].Kind != SuggestFocus {
		t.Fatalf("an idle day should suggest a focus session, got %+v", steps[1])
	}
}

func TestNextStepsOrderAndLimit(t *testing.T) {
	reg := seedRegistry()
	reg.ClassicMode = true
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.Local)
	_, s := reg.findStage(0)
	g, _ := guideFor(0)
	s.setStudied(g.Concepts[0].Name, true) // unlocks due review cards

	steps := reg.nextSteps(now)
	if len(steps) != 4 {
		t.Fatalf("expected the 4-step limit, got %d: %+v", len(steps), steps)
	}
	want := []string{SuggestReview, SuggestConcept, SuggestExercise, SuggestTypeIn}
	for i, k := range want {
		if steps[i].Kind != k {
			t.Fatalf("step %d = %s, want %s (%+v)", i, steps[i].Kind, k, steps)
		}
	}
	if steps[1].Concept != 1 || steps[2].Concept != 0 || steps[2].Exercise != 0 {
		t.Fatalf("wrong concept/exercise picked: %+v", steps)
	}

	// Once every concept is understood the mastery check is suggested.
	for _, c := range g.Concepts {
		s.setStudied(c.Name, true)
	}
	found := false
	for _, st := range reg.nextSteps(now) {
		found = found || (st.Kind == SuggestMastery && st.Stage == 0)
	}
	if !found {
		t.Fatal("a fully understood stage should suggest its mastery check")
	}
}

func TestCurrentStageRespectsPrereqs(t *testing.T) {
	reg := seedRegistry()
	if cs := reg.currentStage(); cs == nil || cs.ID != 0 {
		t.Fatalf("current stage = %+v, want Stage 0", cs)
	}
	_, s := reg.findStage(3)
	g, _ := guideFor(3)
	s.setStudied(g.Concepts[0].Name, true)
	if cs := reg.currentStage(); cs.ID != 3 { // a started stage beats an unstarted one
		t.Fatalf("a started stage should be current, got %d", cs.ID)
	}
	if reg.prereqsReady(16) {
		t.Fatal("the last stage should not be ready on a fresh campus")
	}
}

func TestSearchRequiresEveryWordAndRanksTitles(t *testing.T) {
	hits := search("memory cache")
	if len(hits) == 0 {
		t.Fatal("expected matches for 'memory cache'")
	}
	for _, h := range hits[:min(5, len(hits))] {
		if h.Title == "" || h.Stage < NoStage {
			t.Fatalf("hit without title or stage: %+v", h)
		}
	}
	for i := 1; i < len(hits); i++ {
		if hits[i].score > hits[i-1].score {
			t.Fatal("hits are not sorted by score")
		}
	}
	if len(search("memory zzqqxx")) != 0 {
		t.Fatal("every word must match")
	}
	if search("   ") != nil {
		t.Fatal("a blank query should find nothing")
	}
	// A glossary term's own name outranks passing mentions.
	top := search("mutex")
	if len(top) == 0 || !strings.Contains(strings.ToLower(top[0].Title), "mutex") {
		t.Fatalf("expected a mutex title first, got %+v", top[:min(3, len(top))])
	}
}

func TestSnippetCentresOnTheWord(t *testing.T) {
	text := strings.Repeat("filler ", 30) + "target word here"
	s := snippet(text, "target", 40)
	if !strings.HasPrefix(s, "…") || !strings.Contains(s, "target") || visibleLen(s) > 40 {
		t.Fatalf("snippet = %q", s)
	}
	if got := snippet("short text", "absent", 40); got != "short text" {
		t.Fatalf("snippet without match = %q", got)
	}
}

func TestDailyMinutesAndWeekSummary(t *testing.T) {
	reg := seedRegistry()
	now := time.Date(2026, 9, 24, 18, 0, 0, 0, time.Local)
	reg.StudySessions = []StudySession{
		{Start: now.Add(-time.Hour), Minutes: 25, Stage: 1},
		{Start: now.AddDate(0, 0, -2), Minutes: 50, Stage: NoStage},
		{Start: now.AddDate(0, 0, -9), Minutes: 30, Stage: 2},
		{Start: now.AddDate(0, 0, -40), Minutes: 99}, // outside the window
	}
	reg.ReviewHistory[now.Format("2006-01-02")] = 10
	reg.ReviewHistory[now.AddDate(0, 0, -8).Format("2006-01-02")] = 4

	days := reg.dailyMinutes(now, 14)
	if days[13] != 35 || days[11] != 50 || days[4] != 30 || days[5] != 4 {
		t.Fatalf("daily minutes = %v", days)
	}
	ws := reg.weekSummary(now)
	if ws.minutes != 85 || ws.prevMinutes != 34 || ws.activeDays != 2 || ws.prevDays != 2 {
		t.Fatalf("week summary minutes/days = %+v", ws)
	}
	if ws.reviews != 10 || ws.prevReviews != 4 || ws.sessions != 2 || ws.prevSess != 1 {
		t.Fatalf("week summary counts = %+v", ws)
	}
}

func TestStudySessionsCountInStats(t *testing.T) {
	reg := seedRegistry()
	now := time.Now()
	before := reg.stats(now, 60)
	reg.StudySessions = append(reg.StudySessions, StudySession{Start: now, Minutes: 90, Stage: 1})
	after := reg.stats(now, 60)
	if after.hours-before.hours != 1.5 {
		t.Fatalf("a 90-minute session should add 1.5h, got %.2f", after.hours-before.hours)
	}
	if after.activity[len(after.activity)-1] <= 0 || after.streak < 1 {
		t.Fatalf("today's session should count as activity: %+v", after.activity[len(after.activity)-3:])
	}
}

func TestExportMarkdown(t *testing.T) {
	reg := seedRegistry()
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.Local)
	if md := reg.exportMarkdown(now); !strings.HasPrefix(md, "# ") || strings.Contains(md, "## Stage") {
		t.Fatalf("fresh export should have a title and no stages:\n%s", md)
	}
	_, s := reg.findStage(5)
	g, _ := guideFor(5)
	s.setStudied(g.Concepts[0].Name, true)
	reg.Notebook = append(reg.Notebook, NotebookEntry{At: now, Stage: 5, Ref: "type-in", Kind: NotePrediction, Text: "It prints 42"})
	md := reg.exportMarkdown(now)
	for _, want := range []string{"## Stage 5", g.Concepts[0].Name, "It prints 42"} {
		if !strings.Contains(md, want) {
			t.Fatalf("export missing %q:\n%s", want, md)
		}
	}
}
