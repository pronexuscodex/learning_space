package main

// Motivation and safety: weekly goals, the roadmap's stage states,
// achievements, and rotating backups of the registry. Pure logic; the
// screens live in goalsui.go.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Weekly goals
// ---------------------------------------------------------------------------

// Goals are the learner's weekly targets. Zero means "no goal".
type Goals struct {
	Minutes int `json:"minutes"` // study minutes per week
	Days    int `json:"days"`    // active study days per week
	Cards   int `json:"cards"`   // review cards per week
}

// Suggested starting goals: about 25 minutes on most days, and ten
// review cards a day.
var defaultGoals = Goals{Minutes: 150, Days: 5, Cards: 70}

// weekProgress is this calendar week's progress (Monday to today).
type weekProgress struct {
	minutes float64
	days    int
	cards   int
	dayOf   int // 1 = Monday … 7 = Sunday
}

func (r *Registry) weekToDate(now time.Time) weekProgress {
	wp := weekProgress{dayOf: (int(now.Weekday())+6)%7 + 1}
	for _, m := range r.dailyMinutes(now, wp.dayOf) {
		wp.minutes += m
		if m > 0 {
			wp.days++
		}
	}
	for i := 0; i < wp.dayOf; i++ {
		wp.cards += r.ReviewHistory[now.AddDate(0, 0, -i).Format("2006-01-02")]
	}
	return wp
}

// onTrack reports whether done is at least the share of goal that the
// elapsed part of the week calls for.
func onTrack(done float64, goal, dayOf int) bool {
	return goal <= 0 || done >= float64(goal)*float64(dayOf)/7
}

// ---------------------------------------------------------------------------
// Roadmap
// ---------------------------------------------------------------------------

// Stage states on the roadmap, in order of progress.
const (
	StageLocked     = iota // prerequisites not yet covered
	StageReady             // prerequisites covered, not started
	StageInProgress        // some concepts understood
	StageUnderstood        // every concept understood
	StageMastered          // mastery check passed or stage graduated
)

var stageStateNames = []string{"locked", "ready", "in progress", "understood", "mastered"}
var stageStateGlyphs = []string{"·", "○", "◐", "◉", "★"}

// stageState places a stage on the roadmap.
func (r *Registry) stageState(s *Stage) int {
	if s.Status == StatusGraduated || (s.MasteryCheck != nil && s.MasteryCheck.Passed) {
		return StageMastered
	}
	studied, total := conceptProgress(s)
	switch {
	case total > 0 && studied == total:
		return StageUnderstood
	case studied > 0:
		return StageInProgress
	case r.prereqsReady(s.ID):
		return StageReady
	}
	return StageLocked
}

// missingPrereqs lists the prerequisite stages that are not yet half
// understood.
func (r *Registry) missingPrereqs(stageID int) []int {
	var out []int
	for _, p := range stagePrereqs[stageID] {
		if _, ps := r.findStage(p); ps != nil {
			if studied, total := conceptProgress(ps); studied*2 < total {
				out = append(out, p)
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Achievements
// ---------------------------------------------------------------------------

// Achievement is a milestone worth celebrating.
type Achievement struct {
	ID, Title, Desc string
	earned          func(r *Registry, cs campusStats) bool
}

// anyStage reports whether some stage satisfies fn.
func (r *Registry) anyStage(fn func(s *Stage) bool) bool {
	for _, s := range r.stageOrder() {
		if fn(s) {
			return true
		}
	}
	return false
}

// countStages counts the stages that satisfy fn.
func (r *Registry) countStages(fn func(s *Stage) bool) int {
	n := 0
	for _, s := range r.stageOrder() {
		if fn(s) {
			n++
		}
	}
	return n
}

func (r *Registry) totalReviews() int {
	n := 0
	for _, c := range r.ReviewHistory {
		n += c
	}
	return n
}

var achievements = []Achievement{
	{"first-concept", "First Light", "Understood your first concept.",
		func(r *Registry, cs campusStats) bool { return cs.conceptsStudied >= 1 }},
	{"first-exercise", "Hands On", "Finished your first exercise.",
		func(r *Registry, cs campusStats) bool { return cs.exercisesDone >= 1 }},
	{"ten-concepts", "Getting Serious", "Understood 10 concepts.",
		func(r *Registry, cs campusStats) bool { return cs.conceptsStudied >= 10 }},
	{"fifty-exercises", "Practice Makes Permanent", "Finished 50 exercises.",
		func(r *Registry, cs campusStats) bool { return cs.exercisesDone >= 50 }},
	{"streak-3", "Warming Up", "Studied 3 days in a row.",
		func(r *Registry, cs campusStats) bool { return cs.streak >= 3 }},
	{"streak-7", "One Full Week", "Studied 7 days in a row.",
		func(r *Registry, cs campusStats) bool { return cs.streak >= 7 }},
	{"streak-30", "Habit Formed", "Studied 30 days in a row.",
		func(r *Registry, cs campusStats) bool { return cs.streak >= 30 }},
	{"reviews-100", "Memory Keeper", "Reviewed 100 cards.",
		func(r *Registry, cs campusStats) bool { return r.totalReviews() >= 100 }},
	{"reviews-1000", "Long-Term Memory", "Reviewed 1,000 cards.",
		func(r *Registry, cs campusStats) bool { return r.totalReviews() >= 1000 }},
	{"focus-10", "Deep Worker", "Completed 10 focus sessions.",
		func(r *Registry, cs campusStats) bool { return len(r.StudySessions) >= 10 }},
	{"hours-100", "Centurion", "Logged 100 hours of study.",
		func(r *Registry, cs campusStats) bool { return cs.hours >= 100 }},
	{"texts-5", "Bookworm", "Read 5 of the required texts.",
		func(r *Registry, cs campusStats) bool { return cs.textsRead >= 5 }},
	{"first-lab", "Builder", "Enrolled in your first lab.",
		func(r *Registry, cs campusStats) bool { return cs.labs >= 1 }},
	{"lab-graduated", "Shipped It", "Graduated a lab.",
		func(r *Registry, cs campusStats) bool { return cs.labsGrad >= 1 }},
	{"lookout", "Lookout", "Saved 10 items to your tech-watch log, each with why it matters.",
		func(r *Registry, cs campusStats) bool { return len(r.Watch) >= 10 }},
	{"wordsmith", "Wordsmith", "Added 20 dictionary words to your review deck.",
		func(r *Registry, cs campusStats) bool { return len(r.WordDeck) >= 20 }},
	{"curator", "Curator", "Added 5 of your own resources.",
		func(r *Registry, cs campusStats) bool { return len(r.MyResources) >= 5 }},
	{"type-in", "Typed It In", "Completed a Classic Mode type-in.",
		func(r *Registry, cs campusStats) bool {
			return r.anyStage(func(s *Stage) bool { return s.TypeInDone })
		}},
	{"stage-understood", "Whole Picture", "Understood every concept in a stage.",
		func(r *Registry, cs campusStats) bool {
			return r.anyStage(func(s *Stage) bool { st, t := conceptProgress(s); return t > 0 && st == t })
		}},
	{"concept-mastered", "It Stuck", "Took a concept all the way to mastered.",
		func(r *Registry, cs campusStats) bool { return cs.mastered >= 1 }},
	{"mastery-passed", "Proven", "Passed a stage's mastery check.",
		func(r *Registry, cs campusStats) bool {
			return r.anyStage(func(s *Stage) bool { return s.MasteryCheck != nil && s.MasteryCheck.Passed })
		}},
	{"track-mastered", "Track Complete", "Mastered every stage of a track.",
		func(r *Registry, cs campusStats) bool {
			for ti := range r.Tracks {
				all := len(r.Tracks[ti].Stages) > 0
				for si := range r.Tracks[ti].Stages {
					all = all && r.stageState(&r.Tracks[ti].Stages[si]) == StageMastered
				}
				if all {
					return true
				}
			}
			return false
		}},
	{"graduate", "Graduate", "Mastered all sixteen stages.",
		func(r *Registry, cs campusStats) bool {
			return r.countStages(func(s *Stage) bool { return r.stageState(s) == StageMastered }) == len(r.stageOrder())
		}},
}

// awardAchievements records every newly earned achievement and returns
// them in list order.
func (r *Registry) awardAchievements(now time.Time) []Achievement {
	if r.Achievements == nil {
		r.Achievements = map[string]time.Time{}
	}
	cs := r.stats(now, 60)
	var fresh []Achievement
	for _, a := range achievements {
		if _, have := r.Achievements[a.ID]; have || !a.earned(r, cs) {
			continue
		}
		r.Achievements[a.ID] = now.UTC().Truncate(time.Second)
		fresh = append(fresh, a)
	}
	return fresh
}

// ---------------------------------------------------------------------------
// Backups
// ---------------------------------------------------------------------------

const (
	backupDirName = "academy_backups"
	backupKeep    = 10
)

// backupDir is where backups of the registry at path are kept.
func backupDir(path string) string { return filepath.Join(filepath.Dir(path), backupDirName) }

// backupRegistry copies the current registry file into the backup folder
// (unless it is identical to the newest backup) and keeps the newest
// backupKeep copies. A missing registry is not an error.
func backupRegistry(path string, now time.Time) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	dir := backupDir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	existing, err := listBackups(path)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		if last, err := os.ReadFile(existing[0]); err == nil && string(last) == string(data) {
			return nil // nothing changed since the last backup
		}
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	name := filepath.Join(dir, fmt.Sprintf("%s-%s.json", base, now.Format("20060102-150405.000")))
	if err := atomicWriteFile(name, data); err != nil {
		return err
	}
	all, err := listBackups(path)
	if err != nil {
		return err
	}
	for _, old := range all[min(len(all), backupKeep):] {
		os.Remove(old)
	}
	return nil
}

// listBackups returns the backups of the registry at path, newest first.
func listBackups(path string) ([]string, error) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	matches, err := filepath.Glob(filepath.Join(backupDir(path), base+"-*.json"))
	if err != nil {
		return nil, err
	}
	// The timestamp format sorts lexically.
	sort.Sort(sort.Reverse(sort.StringSlice(matches)))
	return matches, nil
}

// restoreBackup replaces the registry at path with a backup after checking
// that the backup loads cleanly. The current file is backed up first, so a
// restore can itself be undone.
func restoreBackup(path, backup string, now time.Time) error {
	if _, err := os.Stat(backup); err != nil {
		return err
	}
	if _, _, err := loadRegistry(backup); err != nil {
		return fmt.Errorf("backup is not usable: %w", err)
	}
	data, err := os.ReadFile(backup)
	if err != nil {
		return err
	}
	if err := backupRegistry(path, now); err != nil {
		return fmt.Errorf("back up the current registry first: %w", err)
	}
	return atomicWriteFile(path, data)
}

// backupCommand implements -backups and -restore and returns the exit code.
func backupCommand(path string, list bool, restore string) int {
	backups, err := listBackups(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		return 1
	}
	if list {
		if len(backups) == 0 {
			fmt.Printf("No backups yet in %s. One is made each time you commit.\n", backupDir(path))
			return 0
		}
		fmt.Printf("Backups of %s, newest first:\n", path)
		for i, b := range backups {
			info := ""
			if st, err := os.Stat(b); err == nil {
				info = fmt.Sprintf("  (%s, %d KB)", st.ModTime().Format("2006-01-02 15:04:05"), (st.Size()+1023)/1024)
			}
			fmt.Printf("  %2d. %s%s\n", i+1, filepath.Base(b), info)
		}
		fmt.Println("Restore one with: -restore <number>")
		if restore == "" {
			return 0
		}
	}
	src := restore
	if n, err := strconv.Atoi(restore); err == nil {
		if n < 1 || n > len(backups) {
			fmt.Fprintf(os.Stderr, "✗ there is no backup %d (see -backups)\n", n)
			return 1
		}
		src = backups[n-1]
	}
	if err := restoreBackup(path, src, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "✗ restore failed: %v\n", err)
		return 1
	}
	fmt.Printf("✓ Restored %s from %s.\n  The registry you replaced was backed up first, so this can be undone too.\n", path, filepath.Base(src))
	return 0
}
