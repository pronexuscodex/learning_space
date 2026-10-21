package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

// On a page opened from the main menu, a menu key leaves it and is the
// menu's answer at once (no Enter): the page promises "type any key".
func TestHomePagesTakeMenuKeys(t *testing.T) {
	t.Setenv("LINES", "12")
	t.Setenv("TERM", "xterm-256color")
	for _, key := range []string{"d", "j", "g", "k", "6"} {
		var out bytes.Buffer
		a := &App{reg: seedRegistry(), width: 100, pager: true, con: &console{
			in: bufio.NewReader(strings.NewReader(key)), out: &out, raw: true, screen: true,
			enterRaw: func() (func(), error) { return func() {}, nil }}}
		a.pagedHome(func() {
			for i := 1; i <= 40; i++ {
				a.println(fmt.Sprintf("line %d", i))
			}
		})
		if a.con.pending == nil || *a.con.pending != key || a.con.prefill != nil {
			t.Errorf("%q on a home page must go straight to the menu (pending %v, prefill %q)", key, a.con.pending, string(a.con.prefill))
		}
	}
}

// Leaving a concept card early sums it up instead of asking straight away
// whether a half-read card is understood.
func TestConceptCardEarlyExit(t *testing.T) {
	t.Setenv("LINES", "12")
	var out bytes.Buffer
	a := &App{reg: seedRegistry(), width: 80, pager: true,
		con: &console{in: bufio.NewReader(strings.NewReader("q\n3\n")), out: &out}}
	g, _ := guideFor(1)
	if err := a.studyConceptAt(1, g, 0); err != nil {
		t.Fatal(err)
	}
	shown := stripANSI(out.String())
	if !strings.Contains(shown, "you read") || !strings.Contains(shown, "Key idea:") || strings.Contains(shown, "Do you understand it") {
		t.Errorf("an early exit must show the recap, not the question:\n%s", shown)
	}
	if _, s := a.reg.findStage(1); s.hasStudied(g.Concepts[0].Name) {
		t.Error("going back must not mark the concept understood")
	}

	// "I understand it" goes straight to the explanation, without asking again.
	out.Reset()
	a.con = &console{in: bufio.NewReader(strings.NewReader("q\n2\nIt is translated to machine code\n")), out: &out}
	if err := a.studyConceptAt(1, g, 0); err != nil {
		t.Fatal(err)
	}
	if _, s := a.reg.findStage(1); !s.hasStudied(g.Concepts[0].Name) || strings.Contains(stripANSI(out.String()), "Do you understand it") {
		t.Error("\"I understand it\" must mark the concept without asking again")
	}
}

func TestWorkoutSkipsEmptySteps(t *testing.T) {
	var out bytes.Buffer
	a := &App{reg: seedRegistry(), width: 80, con: &console{in: bufio.NewReader(strings.NewReader("3\n")), out: &out}}
	if err := a.dailyWorkout(); err != nil {
		t.Fatal(err)
	}
	shown := stripANSI(out.String())
	if strings.Count(shown, "Ready?") != 1 || !strings.Contains(shown, "Step 1 · Warm-up review: nothing is due") {
		t.Errorf("a new learner's empty steps pass without a question:\n%s", shown)
	}
	day := a.reg.todayWorkout(time.Now())
	if day == nil || !day.has(StepReview) || !day.has(StepRecall) || day.has(StepLearn) {
		t.Errorf("empty steps count as done, the rest wait: %+v", day)
	}
}

func TestPromptsBreakBetweenChoices(t *testing.T) {
	old := layoutWidth
	defer func() { layoutWidth = old }()
	layoutWidth = func() int { return 80 }
	p := stripANSI(promptLabel("Type a word to look up · c categories · a A–Z · t today's word · m my deck", "q back"))
	for _, line := range strings.Split(p, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "›"))
		if strings.HasSuffix(trimmed, " t") || strings.HasPrefix(trimmed, "today's") {
			t.Errorf("a choice was split across lines:\n%s", p)
		}
	}
	if q := stripANSI(promptLabel("Do you understand it well enough to explain it to a friend?", "y/n")); strings.Contains(q, "\n") {
		t.Errorf("a question that fits stays on one line with its hint:\n%s", q)
	}
}

func TestStageOverviewShownOnce(t *testing.T) {
	var out bytes.Buffer
	a := &App{reg: seedRegistry(), width: 80,
		con: &console{in: bufio.NewReader(strings.NewReader("3\nq\n")), out: &out}}
	if err := a.studyHallFor(1); err != nil {
		t.Fatal(err)
	}
	g, _ := guideFor(1)
	first := strings.Fields(g.Overview)[0:6]
	if n := strings.Count(stripANSI(out.String()), strings.Join(first, " ")); n != 1 {
		t.Errorf("the overview must be shown once, got %d", n)
	}
	if strings.Contains(stripANSI(out.String()), "Back to main menu") {
		t.Error("q goes back; no Back item")
	}
}
