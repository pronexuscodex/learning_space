package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWeekToDate(t *testing.T) {
	reg := seedRegistry()
	thu := time.Date(2026, 9, 24, 18, 0, 0, 0, time.Local) // a Thursday
	reg.StudySessions = []StudySession{
		{Start: thu.Add(-time.Hour), Minutes: 30},
		{Start: thu.AddDate(0, 0, -3), Minutes: 20},  // Monday: counts
		{Start: thu.AddDate(0, 0, -4), Minutes: 100}, // last Sunday: does not
	}
	reg.ReviewHistory[thu.Format("2006-01-02")] = 12 // also 12 minutes
	reg.ReviewHistory[thu.AddDate(0, 0, -5).Format("2006-01-02")] = 50

	wp := reg.weekToDate(thu)
	if wp.dayOf != 4 || wp.minutes != 62 || wp.days != 2 || wp.cards != 12 {
		t.Fatalf("week to date = %+v", wp)
	}
	if onTrack(40, 150, 4) || !onTrack(90, 150, 4) || !onTrack(0, 0, 1) {
		t.Fatal("onTrack should compare against the elapsed share of the week")
	}
}

func TestStageStatesAndPrereqs(t *testing.T) {
	reg := seedRegistry()
	_, s1 := reg.findStage(1)
	_, s2 := reg.findStage(2)
	if reg.stageState(s1) != StageReady || reg.stageState(s2) != StageLocked {
		t.Fatalf("fresh: stage 1 %d, stage 2 %d", reg.stageState(s1), reg.stageState(s2))
	}
	if m := reg.missingPrereqs(2); len(m) != 1 || m[0] != 1 {
		t.Fatalf("stage 2 should need stage 1, got %v", m)
	}
	g, _ := guideFor(1)
	s1.setStudied(g.Concepts[0].Name, true)
	if reg.stageState(s1) != StageInProgress {
		t.Fatal("one concept understood should mean in progress")
	}
	for _, c := range g.Concepts {
		s1.setStudied(c.Name, true)
	}
	if reg.stageState(s1) != StageUnderstood || reg.stageState(s2) != StageReady {
		t.Fatal("finishing stage 1 should open stage 2")
	}
	s1.MasteryCheck = &MasteryCheck{Passed: true}
	if reg.stageState(s1) != StageMastered {
		t.Fatal("a passed mastery check should mean mastered")
	}
}

func TestAchievementsAwardOnce(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range achievements {
		if a.ID == "" || a.Title == "" || a.Desc == "" || a.earned == nil || seen[a.ID] {
			t.Fatalf("bad or duplicate achievement %+v", a)
		}
		seen[a.ID] = true
	}
	reg := seedRegistry()
	now := time.Now()
	if got := reg.awardAchievements(now); len(got) != 0 {
		t.Fatalf("a fresh campus earned %v", got)
	}
	_, s := reg.findStage(1)
	g, _ := guideFor(1)
	s.setStudied(g.Concepts[0].Name, true)
	got := reg.awardAchievements(now)
	if len(got) != 1 || got[0].ID != "first-concept" {
		t.Fatalf("expected First Light, got %v", got)
	}
	if again := reg.awardAchievements(now); len(again) != 0 {
		t.Fatal("achievements must be awarded only once")
	}
	if _, ok := reg.Achievements["first-concept"]; !ok {
		t.Fatal("the award should be recorded")
	}
}

func TestBackupsRotateAndRestore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, registryFileName)
	if err := backupRegistry(path, time.Now()); err != nil {
		t.Fatalf("a missing registry is not an error: %v", err)
	}

	reg := seedRegistry()
	write := func(goalMinutes int) {
		reg.Goals.Minutes = goalMinutes
		if err := atomicWriteJSON(path, reg); err != nil {
			t.Fatal(err)
		}
	}
	base := time.Date(2026, 9, 24, 10, 0, 0, 0, time.Local)
	write(1)
	for i := 0; i < 2; i++ { // unchanged file: only one backup
		if err := backupRegistry(path, base.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	if b, _ := listBackups(path); len(b) != 1 {
		t.Fatalf("identical content should be backed up once, got %d", len(b))
	}
	for i := 2; i < 15; i++ {
		write(i)
		if err := backupRegistry(path, base.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	backups, _ := listBackups(path)
	if len(backups) != backupKeep {
		t.Fatalf("expected %d backups, got %d", backupKeep, len(backups))
	}
	if !strings.HasSuffix(backups[0], "-20260924-100014.000.json") {
		t.Fatalf("newest first expected, got %s", filepath.Base(backups[0]))
	}

	// Restore the oldest kept backup (goal 5) over the current file (goal 14).
	if err := restoreBackup(path, backups[len(backups)-1], base.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	restored, _, err := loadRegistry(path)
	if err != nil || restored.Goals.Minutes != 5 {
		t.Fatalf("restore gave goal %d, err %v", restored.Goals.Minutes, err)
	}

	// A broken or missing backup is refused and the registry is untouched.
	bad := filepath.Join(dir, "bad.json")
	os.WriteFile(bad, []byte("{not json"), 0o644)
	if restoreBackup(path, bad, base.Add(2*time.Minute)) == nil || restoreBackup(path, filepath.Join(dir, "nope.json"), base) == nil {
		t.Fatal("unusable backups must be refused")
	}
	if again, _, _ := loadRegistry(path); again.Goals.Minutes != 5 {
		t.Fatal("a refused restore must not change the registry")
	}
}

func TestRefreshFromSeedKeepsProgress(t *testing.T) {
	reg := seedRegistry()
	_, s := reg.findStage(1)
	// An older registry: the Python-era title and reading list, one book read.
	s.Title = "Programming Fundamentals"
	s.Literature = []Literature{{Title: "Think Python", Author: "Allen B. Downey", Kind: "Book", Read: true}}
	s.setStudied("Values, Types & Variables", true)

	if n := reg.refreshFromSeed(seedRegistry()); n != 1 {
		t.Fatalf("refreshed %d stages, want 1", n)
	}
	if s.Title != "Programming Fundamentals in C" {
		t.Fatalf("title = %q", s.Title)
	}
	if len(s.Literature) != 4 || !s.Literature[0].Read || s.Literature[0].Title != "Think Python" {
		t.Fatalf("literature = %+v", s.Literature)
	}
	if !s.hasStudied("Values, Types & Variables") {
		t.Fatal("progress lost")
	}
	if n := reg.refreshFromSeed(seedRegistry()); n != 0 {
		t.Fatal("a second refresh should change nothing")
	}
}
