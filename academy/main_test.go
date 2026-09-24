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
	_, s := reg.findStage(6)
	s.setStudied("Attention & the Transformer", true)
	if err := atomicWriteJSON(path, reg); err != nil {
		t.Fatal(err)
	}
	got, seeded, err := loadRegistry(path)
	if err != nil || seeded {
		t.Fatalf("load: seeded=%v err=%v", seeded, err)
	}
	if _, s := got.findStage(6); !s.hasStudied("Attention & the Transformer") {
		t.Fatal("concept progress lost in round trip")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("temp files left behind: %v", entries)
	}

	// A registry written before concepts_studied existed must still load.
	legacy := `{"schema_version":1,"next_lab_id":1,"tracks":[{"id":"A","name":"x","stages":[{"id":1,"title":"t","status":"Active Research","required_literature":[],"labs":null}]}]}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _, err := loadRegistry(path)
	if err != nil {
		t.Fatalf("legacy load: %v", err)
	}
	if st := old.Tracks[0].Stages[0]; st.ConceptsStudied == nil || st.Labs == nil {
		t.Fatal("nil slices were not normalised")
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
