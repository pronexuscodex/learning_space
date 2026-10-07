package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// studied returns a fresh registry with the first n concepts of a stage
// marked as understood.
func studied(t *testing.T, stage, n int) (*Registry, StageGuide) {
	t.Helper()
	r := seedRegistry()
	g, ok := guideFor(stage)
	if !ok || len(g.Concepts) < n {
		t.Fatalf("stage %d has fewer than %d concepts", stage, n)
	}
	_, s := r.findStage(stage)
	for _, c := range g.Concepts[:n] {
		s.setStudied(c.Name, true)
	}
	return r, g
}

func TestRedoComesBackAfterHelp(t *testing.T) {
	r, g := studied(t, 0, 1)
	c := g.Concepts[0]
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	key := exerciseKey(c.Name, 0)

	r.Solves = append(r.Solves, SolveRecord{Key: key, Stage: 0, Concept: c.Name, Index: 0, At: now, How: SolvedHelp})
	if !r.redoPending(key) {
		t.Fatal("an exercise solved with help must be pending a redo")
	}
	if got := r.redos(now.Add(24*time.Hour), true); len(got) != 0 {
		t.Fatalf("a redo must not be due after one day, got %d", len(got))
	}
	if got := r.redos(now.Add(redoAfter), true); len(got) != 1 || got[0].Key != key {
		t.Fatalf("the redo must be due after %v, got %+v", redoAfter, got)
	}

	// Practice picks the due redo before anything else.
	p, ok := r.practicePick(now.Add(redoAfter), -1, func(int) int { return 0 })
	if !ok || p.Redo == nil || p.Stage != 0 || p.Exercise != 0 {
		t.Fatalf("practicePick must return the due redo, got %+v ok=%v", p, ok)
	}

	// Solving it alone later clears the redo and earns the achievement.
	r.Solves = append(r.Solves, SolveRecord{Key: key, Stage: 0, Concept: c.Name, At: now.Add(4 * 24 * time.Hour), How: SolvedAlone})
	if r.redoPending(key) || !r.redidAlone() {
		t.Fatal("solving alone after help must clear the redo and count as redone alone")
	}
	if sr := r.selfReliance(); sr.own() != 1 || sr.total() != 1 {
		t.Fatalf("self-reliance must use the latest solve per exercise, got %+v", sr)
	}
}

func TestDocsCountAsYourOwnWork(t *testing.T) {
	r := seedRegistry()
	r.Solves = []SolveRecord{{Key: "x #1", At: time.Now(), How: SolvedDocs}}
	if r.redoPending("x #1") {
		t.Error("reading the documentation is the engineer's own work, not help that needs a redo")
	}
}

func TestPracticeInterleavesTopics(t *testing.T) {
	r, _ := studied(t, 0, 2)
	g1, _ := guideFor(1)
	_, s1 := r.findStage(1)
	s1.setStudied(g1.Concepts[0].Name, true)
	now := time.Now()
	for i := 0; i < 5; i++ {
		p, ok := r.practicePick(now, 0, func(n int) int { return i % n })
		if !ok || p.Stage == 0 {
			t.Fatalf("with another stage available, practice must avoid stage 0; got %+v", p)
		}
	}
	// With only the avoided stage left, it is used rather than nothing.
	r2, _ := studied(t, 0, 1)
	if p, ok := r2.practicePick(now, 0, func(int) int { return 0 }); !ok || p.Stage != 0 {
		t.Fatalf("falling back to the same stage, got %+v ok=%v", p, ok)
	}
	// Nothing understood: nothing to practise.
	if _, ok := seedRegistry().practicePick(now, -1, func(int) int { return 0 }); ok {
		t.Error("nothing understood, yet practicePick found an exercise")
	}
}

func TestRecallSchedule(t *testing.T) {
	r, g := studied(t, 0, 2)
	now := time.Now()
	if got := r.recallCandidates(now); len(got) != 2 {
		t.Fatalf("both understood concepts are ready for a first recall, got %d", len(got))
	}
	r.Recalls = append(r.Recalls, RecallRecord{Stage: 0, Concept: g.Concepts[0].Name, At: now, Score: RecallMost, Text: "x"})
	got := r.recallCandidates(now)
	if len(got) != 1 || got[0].Name != g.Concepts[1].Name {
		t.Fatalf("a concept recalled well today is not due again yet; got %+v", got)
	}
	got = r.recallCandidates(now.Add(recallInterval(RecallMost)))
	if len(got) != 2 || got[0].Name != g.Concepts[0].Name {
		t.Fatalf("an overdue recall comes before a first one; got %+v", got)
	}
	if recallInterval(RecallLittle) >= recallInterval(RecallAll) {
		t.Error("weak recalls must come back sooner than strong ones")
	}
}

func TestCalibrationVerdicts(t *testing.T) {
	var c Calibration
	if c.verdict() != "" {
		t.Error("no data must give no verdict")
	}
	for i := 0; i < 10; i++ {
		c.record(ConfSure, i < 6) // certain, right 60%
	}
	if v := c.verdict(); !strings.HasPrefix(v, "Overconfident") {
		t.Errorf("60%% when certain must read as overconfident, got %q", v)
	}
	var good Calibration
	for i := 0; i < 10; i++ {
		good.record(ConfSure, true)
		good.record(ConfGuess, i < 3)
	}
	if v := good.verdict(); !strings.HasPrefix(v, "Well calibrated") {
		t.Errorf("got %q", v)
	}
	good.record(7, true) // out of range: ignored, no panic
	if good.answers() != 20 {
		t.Errorf("answers = %d, want 20", good.answers())
	}
}

func TestWorkoutDays(t *testing.T) {
	r := seedRegistry()
	now := time.Now()
	if r.todayWorkout(now) != nil {
		t.Fatal("no workout before one is started")
	}
	d := r.workoutFor(now)
	for _, s := range workoutSteps[:len(workoutSteps)-1] {
		d.Done = append(d.Done, s)
	}
	if d.complete() || r.workoutsCompleted() != 0 {
		t.Fatal("a workout with a step left is not complete")
	}
	d.Done = append(d.Done, StepReflect)
	no := false
	d.AskedFirst = &no
	if !r.workoutFor(now).complete() || r.workoutsCompleted() != 1 || r.thinkFirstDays() != 1 {
		t.Fatal("finishing every step completes today's workout")
	}
	if len(r.Workouts) != 1 {
		t.Fatalf("workoutFor must reuse today's entry, got %d entries", len(r.Workouts))
	}
}

func TestThinkingDataSurvivesAndIsRepaired(t *testing.T) {
	r := seedRegistry()
	r.Solves = []SolveRecord{{Key: "a #1", How: SolvedAlone, At: time.Now()}, {Key: "b #1", How: "copied"}, {How: SolvedHint}}
	r.Recalls = []RecallRecord{{Concept: "x", Score: 9}, {Concept: "y", Score: RecallSome}}
	r.Calibration = Calibration{{-3, 1}, {2, 5}, {4, 4}}
	r.Settings = Settings{Theme: "neon", PlainSymbols: true}
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
	if len(back.Solves) != 1 || len(back.Recalls) != 1 {
		t.Errorf("invalid solves or recalls must be dropped: %d solves, %d recalls", len(back.Solves), len(back.Recalls))
	}
	if back.Calibration[0][0] != 0 || back.Calibration[0][1] != 0 || back.Calibration[1][1] != 2 {
		t.Errorf("calibration counts must be non-negative with correct ≤ answers, got %v", back.Calibration)
	}
	if back.Settings.Theme != ThemeDefault || !back.Settings.PlainSymbols {
		t.Errorf("an unknown theme falls back to the default, other settings are kept: %+v", back.Settings)
	}
}

func TestPlainSymbolsKeepWidth(t *testing.T) {
	in := "│ 💡 Hint │ 📝 note │ 🏆 ok │ ⚠ done"
	out := plainText(in)
	if visibleLen(out) != visibleLen(in) {
		t.Errorf("width changed: %d → %d (%q)", visibleLen(in), visibleLen(out), out)
	}
	if strings.ContainsFunc(out, func(r rune) bool { return r >= 0x1f000 }) {
		t.Errorf("emoji left in %q", out)
	}
	if plainText("plain text") != "plain text" {
		t.Error("text without emoji must pass through unchanged")
	}
}

func TestThemesRemapColours(t *testing.T) {
	for _, th := range themes {
		s := Style{on: true, theme: th.ID}
		if got := stripANSI(s.Red("x") + s.Gray("y") + s.Bold("z")); got != "xyz" {
			t.Errorf("%s: text must survive styling, got %q", th.Name, got)
		}
	}
	if s := (Style{on: true, theme: ThemeMono}); strings.Contains(s.Green("ok"), "\x1b[32m") {
		t.Error("monochrome must not emit colour codes")
	}
	if s := (Style{on: true, theme: ThemeColorblind}); !strings.Contains(s.Red("x"), "38;5;208") {
		t.Error("colour-blind theme must draw red as orange")
	}
	if s := (Style{on: true, theme: ThemeMono}); !strings.Contains(s.Bold("b"), "\x1b[1m") {
		t.Error("monochrome keeps bold")
	}
}
