package main

import (
	"os"
	"testing"
	"time"
)

// TestWriteDemoRegistry writes a realistic registry for screenshots and
// demos: a learner a few weeks in, with Stage 0 mastered, Stage 1 half
// done, reviews, a streak, labs, notes, goals and achievements. It only
// runs when ACADEMY_DEMO_OUT names the file to write:
//
//	ACADEMY_DEMO_OUT=/tmp/demo.json go test -run TestWriteDemoRegistry
func TestWriteDemoRegistry(t *testing.T) {
	out := os.Getenv("ACADEMY_DEMO_OUT")
	if out == "" {
		t.Skip("set ACADEMY_DEMO_OUT to write the demo registry")
	}
	now := time.Now()
	reg := demoRegistry(now)
	if err := atomicWriteJSON(out, reg); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadRegistry(out); err != nil {
		t.Fatalf("demo registry does not load: %v", err)
	}
}

func demoRegistry(now time.Time) *Registry {
	reg := seedRegistry()
	day := func(ago int, hour int) time.Time {
		d := dayStart(now).AddDate(0, 0, -ago)
		return d.Add(time.Duration(hour) * time.Hour).UTC()
	}

	// Stage 0: every concept understood and practised, check passed.
	_, s0 := reg.findStage(0)
	g0, _ := guideFor(0)
	for i, c := range g0.Concepts {
		s0.setStudied(c.Name, true)
		for e := range c.Exercises {
			if e < 2 || i%2 == 0 {
				s0.setDone(c.Name, e, true)
			}
		}
		for j, card := range conceptCards(0, c) {
			interval := 12.0 // retained
			if i%2 == 0 {
				interval = []float64{25, 30, 40}[j%3] // mastered: every card at 21+ days
			}
			reg.Reviews[card.ID] = ReviewCard{Due: day(-int(interval)/2, 0), IntervalDays: interval, Ease: 2.6, Reps: 4, LastReviewed: day(int(interval)/2, 9)}
		}
	}
	s0.Notes[g0.Concepts[0].Name] = "Bits are switches; a byte is 8 of them. Meaning comes from how you read them: 65 is 'A' as text."
	s0.MasteryCheck = &MasteryCheck{TakenAt: day(3, 20), Score: 11, Total: 12, Passed: true}
	s0.TypeInDone = true
	s0.Literature[0].Read = true
	reg.Notebook = append(reg.Notebook,
		NotebookEntry{At: day(10, 19), Stage: 0, Ref: "type-in", Kind: NotePrediction, Text: "It adds 5+4+3+2+1, so 15; 7 steps per loop, 5 loops, plus 3: 38 steps."},
		NotebookEntry{At: day(10, 20), Stage: 0, Ref: "type-in", Kind: NoteObserved, Text: "Exactly 15 and 38 steps. I mistyped 420 as 402 first: it stored into the wrong cell."},
	)

	// Stage 1: half of the concepts, some exercises, some cards due today.
	_, s1 := reg.findStage(1)
	g1, _ := guideFor(1)
	for i, c := range g1.Concepts[:4] {
		s1.setStudied(c.Name, true)
		s1.setDone(c.Name, 0, true)
		if i < 2 {
			s1.setDone(c.Name, 1, true)
		}
		for j, card := range conceptCards(1, c) {
			if (i+j)%2 == 0 {
				reg.Reviews[card.ID] = ReviewCard{Due: day(0, 0), IntervalDays: 3, Ease: 2.5, Reps: 2, LastReviewed: day(3, 9)}
			}
		}
	}
	s1.Notes[g1.Concepts[1].Name] = "In C a variable is a sized box at an address; 7/2 is 3 because both are ints."
	reg.ExerciseOpened[exerciseKey(g1.Concepts[4].Name, 1)] = now.Add(-12 * time.Minute).UTC()

	// Stage 2: just started.
	_, s2 := reg.findStage(2)
	g2, _ := guideFor(2)
	s2.setStudied(g2.Concepts[0].Name, true)

	// Six weeks of history: most days a little, a current streak of 12.
	for ago := 0; ago < 42; ago++ {
		if ago > 12 && ago%5 == 0 {
			continue
		}
		reg.ReviewHistory[day(ago, 0).Local().Format("2006-01-02")] = 6 + (ago*7)%13
		if ago%3 == 0 {
			reg.StudySessions = append(reg.StudySessions, StudySession{Start: day(ago, 18), Minutes: float64(25 + (ago*11)%35), Stage: []int{0, 1, 1, 2}[ago%4]})
		}
	}

	// Two labs with hour logs.
	_, s := reg.findStage(1)
	s.Labs = append(s.Labs, Lab{ID: 1, Name: "Tiny CPU Emulator (in C)", Architecture: "Stage 0's type-in grown into a real VM: 8 instructions, a loader, a single-step debugger",
		CompilationStatus: CompileTested, Status: StatusActive, EnrolledAt: day(20, 18)})
	_, s5 := reg.findStage(2)
	s5.Labs = append(s5.Labs, Lab{ID: 2, Name: "Route Planner", Architecture: "Dijkstra on a city map, then A*", CompilationStatus: CompileOK, Status: StatusActive, EnrolledAt: day(6, 19)})
	for i := range []int{0, 1} {
		l := &[]*Lab{&s.Labs[0], &s5.Labs[0]}[i]
		for ago := 1; ago < 20; ago += 3 {
			h := 0.5 + float64((ago*3)%4)*0.5
			(*l).Log = append((*l).Log, HourEntry{At: day(ago, 20), Hours: h, Note: "worked on it"})
			(*l).HoursLogged += h
		}
	}
	reg.NextLabID = 3

	reg.Goals = Goals{Minutes: 180, Days: 5, Cards: 70}
	reg.WordDeck = []string{"type casting", "pointer", "undefined behavior", "stack"}
	reg.MyResources = []MyResource{{ID: 1, Stage: 1, Kind: "Video", Title: "Ben Eater: 8-bit computer series", URL: "https://eater.net/8bit", Note: "Watch after Stage 0's CPU concept", Added: day(8, 12)}}
	reg.Watch = []WatchItem{{ID: 1, Added: day(2, 21), Title: "io_uring: faster Linux I/O", Topic: "Systems", Why: "Stage 0's I/O and system calls, taken further", Status: WatchRead, Ring: "assess"}}
	reg.awardAchievements(now)
	reg.LastCommit = now.UTC().Truncate(time.Second)
	return reg
}
