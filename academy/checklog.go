package main

// Where checked exercises live, and the record of every run.
//
// The learner's C files are their own work, so like the PDF library they
// live where people keep documents:
//
//	Documents/Academy Labs/
//	    Stage 01/
//	        hello.c
//	        fizzbuzz.c
//
// $ACADEMY_LABS chooses another folder; with $ACADEMY_HOME set, the labs
// sit inside it. Starter files are created once and never overwritten.
//
// Every run and every hint revealed is appended, one JSON object per line,
// to academy_checks.jsonl beside the registry. It is a separate file so
// that "academy -check" in another terminal never races the app saving the
// registry, and appending never rewrites what is already recorded.

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	labsFolderName   = "Academy Labs"
	checkLogFileName = "academy_checks.jsonl"
)

// labsDir is the folder for the learner's checked-exercise files.
func labsDir(regPath string) string { return labsDirFor(regPath, os.Getenv, documentsDir) }

func labsDirFor(regPath string, getenv func(string) string, docs func() (string, error)) string {
	if d := getenv("ACADEMY_LABS"); d != "" {
		return d
	}
	if home := getenv("ACADEMY_HOME"); home != "" {
		return filepath.Join(home, labsFolderName)
	}
	if d, err := docs(); err == nil {
		return filepath.Join(d, labsFolderName)
	}
	return filepath.Join(filepath.Dir(regPath), labsFolderName)
}

// checkFile is where the learner writes a check's program.
func checkFile(regPath string, ch Check) string {
	return filepath.Join(labsDir(regPath), fmt.Sprintf("Stage %02d", ch.Stage), ch.File)
}

// ensureStarters creates the check's main file and its other files, each
// with its starter, where they do not exist yet. It returns the files it
// created.
func ensureStarters(regPath string, ch Check) ([]string, error) {
	var created []string
	main := checkFile(regPath, ch)
	files := []struct {
		path string
		ch   Check
	}{{main, ch}}
	for _, f := range ch.Extra {
		files = append(files, struct {
			path string
			ch   Check
		}{filepath.Join(filepath.Dir(main), f.Name), Check{Starter: f.Starter}})
	}
	for _, f := range files {
		made, err := ensureStarter(f.path, f.ch)
		if err != nil {
			return created, err
		}
		if made {
			created = append(created, f.path)
		}
	}
	return created, nil
}

// ensureStarter creates the check's file with its starter content if it
// does not exist yet. It never overwrites the learner's work.
func ensureStarter(path string, ch Check) (created bool, err error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := f.WriteString(ch.Starter); err != nil {
		f.Close()
		return false, err
	}
	return true, f.Close()
}

// CheckEvent is one line of the check log.
type CheckEvent struct {
	ID   string    `json:"id"`
	At   time.Time `json:"at"`
	Kind string    `json:"kind"` // "run" or "hint"

	// Kind "hint": how many hints are now revealed.
	Hint int `json:"hint,omitempty"`

	// Kind "run": the evidence.
	Built      bool   `json:"built,omitempty"`
	Passed     int    `json:"passed,omitempty"`
	Total      int    `json:"total,omitempty"`
	OK         bool   `json:"ok,omitempty"`
	HintsSeen  int    `json:"hints_seen,omitempty"` // hints revealed before this run
	Compiler   string `json:"compiler,omitempty"`
	Sanitizers bool   `json:"sanitizers,omitempty"`
}

const (
	checkEventRun  = "run"
	checkEventHint = "hint"
)

func checkLogPath(regPath string) string {
	return filepath.Join(filepath.Dir(regPath), checkLogFileName)
}

// appendCheckEvent adds one event to the log, creating it if needed.
func appendCheckEvent(path string, e CheckEvent) error {
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// readCheckLog reads every event. A missing log is empty; a line that is
// not a valid event (a torn write after a crash) is skipped.
func readCheckLog(path string) ([]CheckEvent, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []CheckEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		var e CheckEvent
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.ID != "" && (e.Kind == checkEventRun || e.Kind == checkEventHint) {
			out = append(out, e)
		}
	}
	return out, sc.Err()
}

// CheckStatus sums up the log for one check.
type CheckStatus struct {
	Runs      int
	Best      int // most cases passed in one run
	Total     int
	Passed    bool // some run passed every case
	OnOwn     bool // some run passed every case before any hint was revealed
	Hints     int  // hints revealed
	Last      time.Time
	LastOK    bool
	LastCount int  // cases passed in the latest run
	LastBuilt bool // the latest run compiled
}

// checkStatuses sums up the log, check by check.
func checkStatuses(events []CheckEvent) map[string]CheckStatus {
	out := map[string]CheckStatus{}
	for _, e := range events {
		s := out[e.ID]
		switch e.Kind {
		case checkEventHint:
			s.Hints = max(s.Hints, e.Hint)
		case checkEventRun:
			s.Runs++
			s.Best = max(s.Best, e.Passed)
			s.Total = e.Total
			s.Last, s.LastOK, s.LastCount, s.LastBuilt = e.At, e.OK, e.Passed, e.Built
			if e.OK {
				s.Passed = true
				s.OnOwn = s.OnOwn || e.HintsSeen == 0
			}
		}
		out[e.ID] = s
	}
	return out
}

// checkStatusesAt sums up the check log beside the registry at regPath.
// An unreadable log counts as empty: it only ever adds evidence.
func checkStatusesAt(regPath string) map[string]CheckStatus {
	events, _ := readCheckLog(checkLogPath(regPath))
	return checkStatuses(events)
}

// runEvent turns a run into its log line.
func runEvent(run CheckRun, hints int, at time.Time) CheckEvent {
	return CheckEvent{ID: run.Check.ID, At: at, Kind: checkEventRun, Built: run.Built,
		Passed: run.Passed, Total: len(run.Check.Cases), OK: run.OK(), HintsSeen: hints,
		Compiler: run.Compiler, Sanitizers: run.Sanitizers}
}
