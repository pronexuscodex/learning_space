package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEveryStageHasAClassicCorner(t *testing.T) {
	for id := range curriculum {
		cg, ti, ok := classicFor(id)
		if !ok {
			t.Errorf("stage %d has no classic corner or type-in", id)
			continue
		}
		for label, r := range map[string]ClassicRead{"anchor": cg.Anchor, "classic": cg.Classic, "source": cg.Source} {
			if r.Title == "" || r.Author == "" || r.Why == "" || r.How == "" {
				t.Errorf("stage %d %s is incomplete: %+v", id, label, r)
			}
			if r.URL != "" && !strings.HasPrefix(r.URL, "https://") {
				t.Errorf("stage %d %s URL must be https: %s", id, label, r.URL)
			}
			if label != "anchor" && r.URL == "" {
				t.Errorf("stage %d %s must link to a free copy", id, label)
			}
		}
		if ti.File == "" || ti.Run == "" || ti.Predict == "" || ti.Lesson == "" || ti.Code == "" || ti.Expected == "" {
			t.Errorf("stage %d type-in is incomplete", id)
		}
		if len(strings.Split(ti.Code, "\n")) > 60 {
			t.Errorf("stage %d type-in is too long to type comfortably", id)
		}
	}
}

func TestStruggleClock(t *testing.T) {
	opened := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	if ok, left := hintUnlocked(opened, LevelPractice, false, opened.Add(10*time.Minute)); ok || left != 20*time.Minute {
		t.Fatalf("practice hint after 10 min: ok=%v left=%v", ok, left)
	}
	if ok, _ := hintUnlocked(opened, LevelPractice, false, opened.Add(30*time.Minute)); !ok {
		t.Fatal("practice hint should unlock at 30 min")
	}
	if ok, _ := hintUnlocked(opened, LevelWarmUp, false, opened.Add(10*time.Minute)); !ok {
		t.Fatal("warm-up hint should unlock at 10 min")
	}
	if ok, _ := hintUnlocked(opened, LevelRealWorld, true, opened.Add(time.Minute)); !ok {
		t.Fatal("logging being stuck should unlock the hint early")
	}
	if ok, _ := hintUnlocked(opened, LevelRealWorld, false, opened.Add(44*time.Minute)); ok {
		t.Fatal("real-world hint must stay locked before 45 min")
	}
}

func TestNotebookRoundTrip(t *testing.T) {
	reg := seedRegistry()
	reg.ClassicMode = true
	key := exerciseKey("Hash Tables", 2)
	reg.ExerciseOpened[key] = time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	reg.Notebook = append(reg.Notebook,
		NotebookEntry{Stage: 2, Ref: key, Kind: NotePlan, Text: "normalise, then dict"},
		NotebookEntry{Stage: 2, Ref: key, Kind: NoteStuck, Text: "unsure how to trim unicode spaces"},
		NotebookEntry{Stage: 3, Ref: "type-in", Kind: NotePrediction, Text: "about 1,100 primes"})
	_, s := reg.findStage(2)
	s.TypeInDone = true

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatal(err)
	}
	var got Registry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !got.ClassicMode || len(got.notebookFor(2, key)) != 2 || len(got.notebookFor(3, "")) != 1 {
		t.Fatalf("notebook or mode lost: %+v", got.Notebook)
	}
	if !got.stuckLogged(key) || got.stuckLogged(exerciseKey("Hash Tables", 0)) {
		t.Fatal("stuck flag is per exercise")
	}
	if _, s := got.findStage(2); !s.TypeInDone {
		t.Fatal("type-in completion lost")
	}
}
