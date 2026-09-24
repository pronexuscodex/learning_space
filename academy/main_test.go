package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// Every seeded stage must have complete Study Hall content.
func TestCurriculumCoversEveryStage(t *testing.T) {
	reg := seedRegistry()
	for _, tr := range reg.Tracks {
		for _, s := range tr.Stages {
			g, ok := guideFor(s.ID)
			if !ok {
				t.Fatalf("stage %d has no guide", s.ID)
			}
			if g.Overview == "" || len(g.Concepts) == 0 || len(g.Resources) == 0 || len(g.Blueprints) == 0 || len(g.Quiz) == 0 {
				t.Errorf("stage %d guide is incomplete", s.ID)
			}
			names := map[string]bool{}
			for _, c := range g.Concepts {
				if names[c.Name] {
					t.Errorf("stage %d: duplicate concept %q", s.ID, c.Name)
				}
				names[c.Name] = true
				if c.Body == "" || c.Summary == "" || c.MentalModel == "" {
					t.Errorf("stage %d concept %q is missing text", s.ID, c.Name)
				}
			}
			for _, r := range g.Resources {
				if r.URL != "" && !strings.HasPrefix(r.URL, "https://") {
					t.Errorf("stage %d resource %q: URL must be https", s.ID, r.Title)
				}
			}
			for _, b := range g.Blueprints {
				if utf8.RuneCountInString(b.Name) > maxNameLen || utf8.RuneCountInString(b.Brief) > maxNotesLen {
					t.Errorf("stage %d blueprint %q exceeds input limits", s.ID, b.Name)
				}
			}
		}
	}
}

func TestWrapRespectsWidthAndBullets(t *testing.T) {
	text := "alpha beta gamma delta epsilon zeta eta theta\nsoft break\n\n- bullet one that is fairly long indeed\nnext para"
	lines := wrap(text, 20, "  ")
	for _, l := range lines {
		if n := utf8.RuneCountInString(l); n > 20 {
			t.Errorf("line %q is %d runes, want <= 20", l, n)
		}
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "• bullet") {
		t.Errorf("bullet not rendered:\n%s", joined)
	}
	if !strings.Contains(joined, "theta soft") {
		t.Errorf("soft break not joined:\n%s", joined)
	}
}

func TestVisibleLenIgnoresANSI(t *testing.T) {
	s := Style{on: true}
	if got := visibleLen(s.Bold(s.Red("héllo"))); got != 5 {
		t.Fatalf("visibleLen = %d, want 5", got)
	}
}

func TestParseHours(t *testing.T) {
	cases := map[string]float64{"1.5": 1.5, "1h30m": 1.5, "45m": 0.75, "2H": 2}
	for in, want := range cases {
		got, err := parseHours(in)
		if err != nil || got != want {
			t.Errorf("parseHours(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"NaN", "Inf", "abc", ""} {
		if _, err := parseHours(bad); err == nil {
			t.Errorf("parseHours(%q) should fail", bad)
		}
	}
}

func TestAtomicWriteRoundTripAndLegacyLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, registryFileName)

	reg := seedRegistry()
	_, s := reg.findStage(15)
	s.setStudied("Attention & the Transformer", true)
	s.setDone("Attention & the Transformer", 2, true)
	if err := atomicWriteJSON(path, reg); err != nil {
		t.Fatal(err)
	}
	got, res, err := loadRegistry(path)
	if err != nil || res != (loadResult{}) {
		t.Fatalf("load: result=%+v err=%v", res, err)
	}
	if _, s := got.findStage(15); !s.hasStudied("Attention & the Transformer") || !s.hasDone("Attention & the Transformer", 2) {
		t.Fatal("concept or exercise progress lost in round trip")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("temp files left behind: %v", entries)
	}

	// A schema-1 registry (7 stages, no concepts_studied) must migrate:
	// stage IDs move to their new places, progress and labs survive, and
	// the new stages are merged in.
	legacy := `{"schema_version":1,"next_lab_id":2,"tracks":[
	  {"id":"A","name":"Systems","stages":[{"id":1,"title":"Iron","status":"Mastered/Graduated","required_literature":[],"labs":[
	    {"id":1,"name":"Tiny C","architecture_notes":"","hours_logged":3,"compilation_status":"Compiles","status":"Active Research","enrolled_at":"2026-01-01T00:00:00Z","hour_log":[]}]}]},
	  {"id":"B","name":"AI","stages":[{"id":6,"title":"Neural","status":"Active Research","required_literature":[],"concepts_studied":["Attention & the Transformer"],"labs":null}]}]}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	old, res, err := loadRegistry(path)
	if err != nil {
		t.Fatalf("legacy load: %v", err)
	}
	if !res.Migrated || res.NewStages != 14 {
		t.Fatalf("result = %+v, want migrated with 14 new stages", res)
	}
	// Titles follow the curriculum; status and labs are the learner's own.
	if _, s := old.findStage(5); s == nil || s.Title != "The Iron Layer (Low-Level Systems & Compilers)" || s.Status != StatusGraduated || len(s.Labs) != 1 || len(s.Literature) != 3 {
		t.Fatalf("stage 1 did not become stage 5 with its progress: %+v", s)
	}
	if _, s := old.findStage(15); s == nil || !s.hasStudied("Attention & the Transformer") || s.Labs == nil {
		t.Fatal("stage 6 did not become stage 15 with its progress")
	}
	if tr, s := old.findStage(1); s == nil || tr.ID != "F" || old.Tracks[0].ID != "F" {
		t.Fatal("Foundations track missing or not first")
	}
	var ids []string
	for _, tr := range old.Tracks {
		ids = append(ids, tr.ID)
	}
	if strings.Join(ids, "") != "FASB" {
		t.Fatalf("track order = %v, want F A S B", ids)
	}
	for _, id := range []int{2, 3, 4, 6, 7, 8} {
		if _, s := old.findStage(id); s == nil {
			t.Errorf("stage %d missing after reconcile", id)
		}
	}
	if _, s := old.findStage(5); old.Tracks[1].Stages[0].ID != 5 || s == nil {
		t.Error("existing stage should come first in its track")
	}

}

func TestStatsActivityAndStreak(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.Local)
	reg := seedRegistry()
	_, s := reg.findStage(1)
	s.Labs = append(s.Labs, Lab{ID: 1, Status: StatusActive, CompilationStatus: CompileOK, HoursLogged: 6, Log: []HourEntry{
		{At: now.Add(-48 * time.Hour), Hours: 1},
		{At: now.Add(-24 * time.Hour), Hours: 2},
		{At: now, Hours: 3},
	}})
	cs := reg.stats(now, 14)
	if cs.streak != 3 {
		t.Errorf("streak = %d, want 3", cs.streak)
	}
	if last := cs.activity[13]; last != 3 {
		t.Errorf("today's activity = %v, want 3", last)
	}
	if cs.labs != 1 || cs.hours != 6 {
		t.Errorf("labs=%d hours=%v", cs.labs, cs.hours)
	}
}

// Beginners must never meet a concept without its plain-words layer.
func TestEveryConceptHasBeginnerExplanations(t *testing.T) {
	for id, g := range curriculum {
		if len(g.Outcomes) == 0 || len(g.Glossary) < 5 {
			t.Errorf("stage %d: needs outcomes and at least 5 glossary terms", id)
		}
		for _, c := range g.Concepts {
			if c.Analogy == "" || c.Example == "" {
				t.Errorf("stage %d concept %q: missing analogy or real-life example", id, c.Name)
			}
			levels := []string{LevelWarmUp, LevelPractice, LevelRealWorld}
			if len(c.Exercises) != len(levels) {
				t.Errorf("stage %d concept %q: has %d exercises, want warm-up, practice and real-world", id, c.Name, len(c.Exercises))
				continue
			}
			for i, e := range c.Exercises {
				if e.Level != levels[i] || e.Task == "" || e.Hint == "" {
					t.Errorf("stage %d concept %q exercise %d is incomplete", id, c.Name, i+1)
				}
			}
		}
	}
	keyed := map[string]bool{}
	for name := range plainWords {
		keyed[name] = true
	}
	for name := range conceptExercises {
		keyed[name] = true
	}
	for name := range keyed {
		found := false
		for _, g := range curriculum {
			for _, c := range g.Concepts {
				found = found || c.Name == name
			}
		}
		if !found {
			t.Errorf("explainer or exercise set %q matches no concept (renamed?)", name)
		}
	}
}

func TestTermsInMatchesWholeWordsAndPlurals(t *testing.T) {
	gl := []Term{{Word: "Register"}, {Word: "Cache"}, {Word: "ABI"}}
	got := termsIn(gl, "The CPU has 16 registers and a big cache.", "Nothing about tabIs.")
	if len(got) != 2 || got[0].Word != "Register" || got[1].Word != "Cache" {
		t.Fatalf("termsIn = %+v", got)
	}
}

// Every guide must belong to a seeded stage, and IDs must run 1..N.
func TestGuidesMatchSeedStages(t *testing.T) {
	reg := seedRegistry()
	n := 0
	for _, tr := range reg.Tracks {
		n += len(tr.Stages)
	}
	if n != len(curriculum) {
		t.Fatalf("%d seeded stages but %d guides", n, len(curriculum))
	}
	for id := 1; id <= n; id++ {
		if _, s := reg.findStage(id); s == nil {
			t.Errorf("stage %d is not seeded", id)
		}
		if _, ok := guideFor(id); !ok {
			t.Errorf("stage %d has no guide", id)
		}
	}
}
