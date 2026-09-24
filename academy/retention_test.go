package main

import (
	"testing"
	"time"
)

func TestScheduleGrowsIntervalsAndResetsOnLapse(t *testing.T) {
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.Local)
	var c ReviewCard
	var got []float64
	var reviewedAt time.Time
	for i := 0; i < 5; i++ {
		reviewedAt = now
		c = schedule(c, GradeGood, now)
		got = append(got, c.IntervalDays)
		now = c.Due
	}
	// 1, 3, then growing by the ease factor (2.5).
	if got[0] != 1 || got[1] != 3 || got[2] != 7.5 || got[3] <= got[2] || got[4] <= got[3] {
		t.Fatalf("intervals = %v", got)
	}
	if !c.Due.After(reviewedAt.AddDate(0, 0, 20)) {
		t.Fatalf("due %v not far enough in the future", c.Due)
	}

	lapsed := schedule(c, GradeAgain, now)
	if lapsed.Reps != 0 || lapsed.Lapses != 1 || lapsed.IntervalDays != 0 || lapsed.Ease >= c.Ease {
		t.Fatalf("lapse not applied: %+v", lapsed)
	}
	if !lapsed.Due.Before(now.Add(time.Hour)) {
		t.Fatal("a forgotten card should come back within the hour")
	}

	easy := schedule(ReviewCard{}, GradeEasy, now)
	hard := schedule(ReviewCard{}, GradeHard, now)
	if !(easy.IntervalDays > hard.IntervalDays) || easy.Ease <= hard.Ease {
		t.Fatalf("easy %+v should outpace hard %+v", easy, hard)
	}
	for i := 0; i < 50; i++ {
		easy = schedule(easy, GradeEasy, now)
	}
	if easy.IntervalDays > maxInterval {
		t.Fatalf("interval %v exceeds cap", easy.IntervalDays)
	}
}

func TestDeckUnlocksAndDueCards(t *testing.T) {
	reg := seedRegistry()
	now := time.Now()
	if n := len(reg.unlockedCards()); n != 0 {
		t.Fatalf("fresh campus has %d unlocked cards", n)
	}
	_, s := reg.findStage(2)
	g, _ := guideFor(2)
	s.setStudied(g.Concepts[0].Name, true)
	cards := reg.unlockedCards()
	if len(cards) != 3 {
		t.Fatalf("one understood concept should unlock 3 cards, got %d", len(cards))
	}
	if len(reg.dueCards(now)) != 3 {
		t.Fatal("new cards should be due immediately")
	}
	reg.Reviews[cards[0].ID] = schedule(ReviewCard{}, GradeGood, now)
	if len(reg.dueCards(now)) != 2 {
		t.Fatal("a reviewed card should leave the due list")
	}

	// Quiz cards unlock once half the stage's concepts are understood.
	for _, c := range g.Concepts[:3] {
		s.setStudied(c.Name, true)
	}
	quiz := 0
	for _, c := range reg.unlockedCards() {
		if c.Kind == CardQuiz {
			quiz++
		}
	}
	if quiz != len(g.Quiz) {
		t.Fatalf("expected %d quiz cards unlocked, got %d", len(g.Quiz), quiz)
	}
}

func TestMasteryLadder(t *testing.T) {
	reg := seedRegistry()
	_, s := reg.findStage(1)
	g, _ := guideFor(1)
	c := g.Concepts[0]
	step := func(want int) {
		t.Helper()
		if got := reg.conceptMastery(s, c); got != want {
			t.Fatalf("mastery = %s, want %s", masteryNames[got], masteryNames[want])
		}
	}
	step(MasteryNew)
	s.setStudied(c.Name, true)
	step(MasteryUnderstood)
	s.setDone(c.Name, 0, true)
	s.setDone(c.Name, 1, true)
	step(MasteryPractised)
	for _, card := range conceptCards(1, c) {
		reg.Reviews[card.ID] = ReviewCard{IntervalDays: 8, Reps: 3}
	}
	step(MasteryRetained)
	for _, card := range conceptCards(1, c) {
		reg.Reviews[card.ID] = ReviewCard{IntervalDays: 25, Reps: 4}
	}
	step(MasteryRetained) // long retention, but one exercise still open
	s.setDone(c.Name, 2, true)
	step(MasteryMastered)
}

func TestConceptNamesUniqueAndLinksValid(t *testing.T) {
	seen := map[string]int{}
	for id, g := range curriculum {
		for _, c := range g.Concepts {
			if prev, dup := seen[c.Name]; dup {
				t.Errorf("concept %q appears in stages %d and %d", c.Name, prev, id)
			}
			seen[c.Name] = id
		}
	}
	for name := range seen {
		l, ok := conceptLinks[name]
		if !ok || l.Read == "" || len(l.Related) == 0 {
			t.Errorf("concept %q needs related concepts and a go-deeper pointer", name)
			continue
		}
		for _, r := range l.Related {
			if _, ok := seen[r]; !ok {
				t.Errorf("concept %q links to unknown concept %q", name, r)
			}
			if r == name {
				t.Errorf("concept %q links to itself", name)
			}
		}
	}
	for name := range conceptLinks {
		if _, ok := seen[name]; !ok {
			t.Errorf("links entry %q matches no concept", name)
		}
	}
	for id := range curriculum {
		pre, ok := stagePrereqs[id]
		if !ok {
			t.Errorf("stage %d has no prerequisites entry", id)
		}
		for _, p := range pre {
			if p >= id {
				t.Errorf("stage %d lists later stage %d as a prerequisite", id, p)
			}
		}
	}
}

func TestEveryResourceIsHTTPSAndTitled(t *testing.T) {
	for _, l := range allResourceLinks() {
		if l.Title == "" || len(l.URL) < 12 || l.URL[:8] != "https://" {
			t.Errorf("stage %d resource %q has a bad URL %q", l.Stage, l.Title, l.URL)
		}
	}
}
