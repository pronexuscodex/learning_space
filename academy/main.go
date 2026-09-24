// Command academy is a zero-dependency terminal ledger for a private
// "Systems & AI Academy" curriculum.
//
// All progress lives in a single flat JSON file (academy_campus_registry.json)
// stored next to the binary. The file is human-readable by design: you can
// open, diff, or version-control it at any time. Writes are atomic: the new
// state is written to a temporary file in the same directory, fsynced, and
// renamed over the old file, so a crash mid-write can never leave a
// half-written registry behind.
//
// Teaching content (concept explanations, resources, lab blueprints, quiz
// questions) is compiled into the binary; see curriculum.go.
//
// Source layout (one package, one build target):
//
//	main.go        schema, storage layer, entry point
//	console.go     sanitised line-based input over bufio
//	ui.go          ANSI styling, bars, wrapping, sparklines
//	app.go         the interactive ledger actions
//	studyhall.go   concept reader, resource library, blueprints, quizzes
//	curriculum.go  the Study Hall knowledge base
//
// Build:
//
//	go build -trimpath -ldflags="-s -w" -o academy .
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// registryFileName is the on-disk name of the flat JSON database.
const registryFileName = "academy_campus_registry.json"

// schemaVersion is bumped whenever the JSON layout changes incompatibly.
// Version 1 files without concepts_studied load fine: the field is additive.
const schemaVersion = 1

// Input and value limits. Anything outside these is rejected, not truncated.
const (
	maxNameLen       = 80
	maxNotesLen      = 600
	maxLogNoteLen    = 200
	maxInitialHours  = 10000.0
	maxHoursPerEntry = 24.0
)

// ---------------------------------------------------------------------------
// Data schema
// ---------------------------------------------------------------------------

// Status is the academic verification status shared by stages and labs.
type Status string

const (
	StatusActive    Status = "Active Research"
	StatusGraduated Status = "Mastered/Graduated"
)

// Valid reports whether s is a known status.
func (s Status) Valid() bool { return s == StatusActive || s == StatusGraduated }

// Toggled returns the opposite status.
func (s Status) Toggled() Status {
	if s == StatusGraduated {
		return StatusActive
	}
	return StatusGraduated
}

// CompileStatus tracks whether a lab's code actually builds.
type CompileStatus string

const (
	CompileNotBuilt CompileStatus = "Not Compiled"
	CompileFailing  CompileStatus = "Build Failing"
	CompileOK       CompileStatus = "Compiles"
	CompileTested   CompileStatus = "Tests Passing"
)

// compileStatuses lists every CompileStatus in menu order.
var compileStatuses = []CompileStatus{CompileNotBuilt, CompileFailing, CompileOK, CompileTested}

// Valid reports whether c is a known compilation status.
func (c CompileStatus) Valid() bool {
	for _, known := range compileStatuses {
		if c == known {
			return true
		}
	}
	return false
}

// Literature is a single required book or paper for a stage.
type Literature struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Kind   string `json:"kind"` // "Book" or "Paper"
	Read   bool   `json:"read"`
}

// HourEntry is one logged block of focused time on a lab.
type HourEntry struct {
	At    time.Time `json:"at"`
	Hours float64   `json:"hours"`
	Note  string    `json:"note,omitempty"`
}

// Lab is a hands-on engineering project built under a stage.
type Lab struct {
	ID                int           `json:"id"`
	Name              string        `json:"name"`
	Architecture      string        `json:"architecture_notes"`
	HoursLogged       float64       `json:"hours_logged"`
	CompilationStatus CompileStatus `json:"compilation_status"`
	Status            Status        `json:"status"`
	EnrolledAt        time.Time     `json:"enrolled_at"`
	Log               []HourEntry   `json:"hour_log"`
}

// Stage is one numbered unit of the curriculum.
type Stage struct {
	ID              int          `json:"id"`
	Title           string       `json:"title"`
	Status          Status       `json:"status"`
	Literature      []Literature `json:"required_literature"`
	ConceptsStudied []string     `json:"concepts_studied"`
	Labs            []Lab        `json:"labs"`
}

// hasStudied reports whether the named concept is marked as understood.
func (s *Stage) hasStudied(concept string) bool {
	for _, c := range s.ConceptsStudied {
		if c == concept {
			return true
		}
	}
	return false
}

// setStudied marks or unmarks a concept as understood.
func (s *Stage) setStudied(concept string, studied bool) {
	kept := s.ConceptsStudied[:0]
	for _, c := range s.ConceptsStudied {
		if c != concept {
			kept = append(kept, c)
		}
	}
	if studied {
		kept = append(kept, concept)
	}
	s.ConceptsStudied = kept
}

// conceptProgress returns (studied, total) against the Study Hall guide.
// Only names that still exist in the guide are counted.
func conceptProgress(s *Stage) (int, int) {
	g, ok := guideFor(s.ID)
	if !ok {
		return 0, 0
	}
	n := 0
	for _, c := range g.Concepts {
		if s.hasStudied(c.Name) {
			n++
		}
	}
	return n, len(g.Concepts)
}

// Track groups stages under a single theme.
type Track struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Stages []Stage `json:"stages"`
}

// Registry is the complete persisted state of the academy.
type Registry struct {
	SchemaVersion int       `json:"schema_version"`
	LastCommit    time.Time `json:"last_commit"`
	NextLabID     int       `json:"next_lab_id"`
	Tracks        []Track   `json:"tracks"`
}

// findStage returns the stage with the given global ID and its track.
func (r *Registry) findStage(id int) (*Track, *Stage) {
	for ti := range r.Tracks {
		t := &r.Tracks[ti]
		for si := range t.Stages {
			if t.Stages[si].ID == id {
				return t, &t.Stages[si]
			}
		}
	}
	return nil, nil
}

// findLab returns the lab with the given ID and the stage that owns it.
func (r *Registry) findLab(id int) (*Stage, *Lab) {
	for ti := range r.Tracks {
		for si := range r.Tracks[ti].Stages {
			s := &r.Tracks[ti].Stages[si]
			for li := range s.Labs {
				if s.Labs[li].ID == id {
					return s, &s.Labs[li]
				}
			}
		}
	}
	return nil, nil
}

// labCount returns the total number of labs across all tracks.
func (r *Registry) labCount() int {
	n := 0
	for _, t := range r.Tracks {
		for _, s := range t.Stages {
			n += len(s.Labs)
		}
	}
	return n
}

// stageHours sums the hours of every lab in a stage.
func stageHours(s *Stage) float64 {
	var h float64
	for _, l := range s.Labs {
		h += l.HoursLogged
	}
	return roundHours(h)
}

// literatureProgress returns (read, total) for a stage's reading list.
func literatureProgress(s *Stage) (int, int) {
	read := 0
	for _, lit := range s.Literature {
		if lit.Read {
			read++
		}
	}
	return read, len(s.Literature)
}

// campusStats is the aggregate shown on the dashboard.
type campusStats struct {
	hours                     float64
	labs, labsGrad            int
	stages, stagesGrad        int
	texts, textsRead          int
	concepts, conceptsStudied int
	activity                  []float64 // hours per day, oldest first
	streak                    int       // consecutive active days
}

// stats aggregates progress across the whole campus. activityDays sets how
// many days of history are bucketed for the sparkline and streak.
func (r *Registry) stats(now time.Time, activityDays int) campusStats {
	cs := campusStats{activity: make([]float64, activityDays)}
	today := dayStart(now)
	for ti := range r.Tracks {
		for si := range r.Tracks[ti].Stages {
			s := &r.Tracks[ti].Stages[si]
			cs.stages++
			if s.Status == StatusGraduated {
				cs.stagesGrad++
			}
			read, total := literatureProgress(s)
			cs.textsRead += read
			cs.texts += total
			studied, concepts := conceptProgress(s)
			cs.conceptsStudied += studied
			cs.concepts += concepts
			for _, l := range s.Labs {
				cs.labs++
				cs.hours += l.HoursLogged
				if l.Status == StatusGraduated {
					cs.labsGrad++
				}
				for _, e := range l.Log {
					ago := int(math.Round(today.Sub(dayStart(e.At)).Hours() / 24))
					if ago >= 0 && ago < activityDays {
						cs.activity[activityDays-1-ago] += e.Hours
					}
				}
			}
		}
	}
	cs.hours = roundHours(cs.hours)

	// Streak: consecutive days with logged time, ending today (or yesterday,
	// so the streak survives until you study today).
	i := activityDays - 1
	if i >= 0 && cs.activity[i] == 0 {
		i--
	}
	for ; i >= 0 && cs.activity[i] > 0; i-- {
		cs.streak++
	}
	return cs
}

// dayStart truncates t to local midnight.
func dayStart(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

// validate checks structural invariants after loading from disk, normalises
// nil slices, and repairs the lab ID counter if it has drifted.
func (r *Registry) validate() error {
	if r.SchemaVersion != schemaVersion {
		return fmt.Errorf("unsupported schema_version %d (expected %d)", r.SchemaVersion, schemaVersion)
	}
	if len(r.Tracks) == 0 {
		return errors.New("registry contains no tracks")
	}
	stageIDs := map[int]bool{}
	labIDs := map[int]bool{}
	maxLab := 0
	for ti := range r.Tracks {
		for si := range r.Tracks[ti].Stages {
			s := &r.Tracks[ti].Stages[si]
			if stageIDs[s.ID] {
				return fmt.Errorf("duplicate stage id %d", s.ID)
			}
			stageIDs[s.ID] = true
			if !s.Status.Valid() {
				return fmt.Errorf("stage %d has invalid status %q", s.ID, s.Status)
			}
			if s.ConceptsStudied == nil {
				s.ConceptsStudied = []string{}
			}
			if s.Labs == nil {
				s.Labs = []Lab{}
			}
			for _, l := range s.Labs {
				if labIDs[l.ID] {
					return fmt.Errorf("duplicate lab id %d", l.ID)
				}
				labIDs[l.ID] = true
				if !l.Status.Valid() {
					return fmt.Errorf("lab %d has invalid status %q", l.ID, l.Status)
				}
				if !l.CompilationStatus.Valid() {
					return fmt.Errorf("lab %d has invalid compilation_status %q", l.ID, l.CompilationStatus)
				}
				if l.HoursLogged < 0 || math.IsNaN(l.HoursLogged) || math.IsInf(l.HoursLogged, 0) {
					return fmt.Errorf("lab %d has invalid hours_logged %v", l.ID, l.HoursLogged)
				}
				maxLab = max(maxLab, l.ID)
			}
		}
	}
	if r.NextLabID <= maxLab {
		r.NextLabID = maxLab + 1
	}
	return nil
}

// seedRegistry builds the default two-track curriculum.
func seedRegistry() *Registry {
	book := func(title, author string) Literature { return Literature{Title: title, Author: author, Kind: "Book"} }
	paper := func(title, author string) Literature { return Literature{Title: title, Author: author, Kind: "Paper"} }
	stage := func(id int, title string, lit ...Literature) Stage {
		return Stage{ID: id, Title: title, Status: StatusActive, Literature: lit, ConceptsStudied: []string{}, Labs: []Lab{}}
	}

	return &Registry{
		SchemaVersion: schemaVersion,
		NextLabID:     1,
		Tracks: []Track{
			{
				ID:   "A",
				Name: "System Core Foundations",
				Stages: []Stage{
					stage(1, "The Iron Layer (Low-Level Systems & Compilers)",
						book("Computer Systems: A Programmer's Perspective", "Bryant & O'Hallaron"),
						book("Crafting Interpreters", "Robert Nystrom"),
						book("Compilers: Principles, Techniques, and Tools", "Aho, Lam, Sethi & Ullman"),
					),
					stage(2, "Operating Systems Internals & Memory Layouts",
						book("Operating Systems: Three Easy Pieces", "Arpaci-Dusseau & Arpaci-Dusseau"),
						paper("What Every Programmer Should Know About Memory", "Ulrich Drepper"),
						book("Understanding the Linux Kernel", "Bovet & Cesati"),
					),
					stage(3, "Storage Engines & State Persistence",
						book("Database Internals", "Alex Petrov"),
						book("Designing Data-Intensive Applications", "Martin Kleppmann"),
						paper("The Log-Structured Merge-Tree (LSM-Tree)", "O'Neil, Cheng, Gawlick & O'Neil"),
					),
					stage(4, "Networks, Sockets, & Distributed Topology",
						book("TCP/IP Illustrated, Vol. 1", "W. Richard Stevens"),
						book("UNIX Network Programming, Vol. 1", "W. Richard Stevens"),
						paper("Time, Clocks, and the Ordering of Events in a Distributed System", "Leslie Lamport"),
						paper("In Search of an Understandable Consensus Algorithm (Raft)", "Ongaro & Ousterhout"),
					),
				},
			},
			{
				ID:   "B",
				Name: "Advanced AI & Hardware Stack",
				Stages: []Stage{
					stage(5, "Mathematical Foundations (Matrix Calculus & Linear Algebra)",
						book("Linear Algebra Done Right", "Sheldon Axler"),
						book("Mathematics for Machine Learning", "Deisenroth, Faisal & Ong"),
						paper("The Matrix Calculus You Need for Deep Learning", "Parr & Howard"),
					),
					stage(6, "Neural Architectures & Autograd from Scratch",
						book("Deep Learning", "Goodfellow, Bengio & Courville"),
						paper("Automatic Differentiation in Machine Learning: a Survey", "Baydin, Pearlmutter, Radul & Siskind"),
						paper("Attention Is All You Need", "Vaswani et al."),
					),
					stage(7, "AI Infrastructure, CUDA, & Memory-Bound Inference",
						book("Programming Massively Parallel Processors", "Hwu, Kirk & El Hajj"),
						paper("Roofline: An Insightful Visual Performance Model", "Williams, Waterman & Patterson"),
						paper("FlashAttention: Fast and Memory-Efficient Exact Attention", "Dao et al."),
						paper("Efficient Memory Management for LLM Serving with PagedAttention", "Kwon et al."),
					),
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Storage layer
// ---------------------------------------------------------------------------

// defaultRegistryPath places the registry next to the executable. When the
// binary lives in a temporary build directory (as with `go run`), it falls
// back to the current working directory so data is not lost on exit.
func defaultRegistryPath() string {
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		dir := filepath.Dir(exe)
		tmp := filepath.Clean(os.TempDir())
		if !strings.HasPrefix(dir, tmp) && !strings.Contains(dir, "go-build") {
			return filepath.Join(dir, registryFileName)
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, registryFileName)
	}
	return registryFileName
}

// loadRegistry reads the registry from path. If the file does not exist, a
// freshly seeded registry is returned with seeded=true. A corrupt file is an
// error: it is never silently overwritten.
func loadRegistry(path string) (reg *Registry, seeded bool, err error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return seedRegistry(), true, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}
	reg = &Registry{}
	if err := json.Unmarshal(data, reg); err != nil {
		return nil, false, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := reg.validate(); err != nil {
		return nil, false, fmt.Errorf("validate %s: %w", path, err)
	}
	return reg, false, nil
}

// atomicWriteJSON serializes v and swaps it into place at path.
//
// Sequence: write temp file in the same directory -> fsync -> close ->
// rename over target (atomic on POSIX and on NTFS via MoveFileEx) -> fsync
// the directory so the rename itself is durable.
func atomicWriteJSON(path string, v any) (err error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".academy-registry-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("fsync temp file: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err = os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err = os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename into place: %w", err)
	}

	// Best effort: directory fsync is unsupported on some platforms.
	if d, derr := os.Open(dir); derr == nil {
		_ = d.Sync()
		d.Close()
	}
	return nil
}

// roundHours keeps hours at two decimal places to avoid float drift.
func roundHours(h float64) float64 { return math.Round(h*100) / 100 }

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

func main() {
	pathFlag := flag.String("registry", "", "path to the JSON registry (default: "+registryFileName+" beside the binary)")
	noColor := flag.Bool("no-color", false, "disable ANSI colours (also honours NO_COLOR)")
	flag.Parse()

	sty = Style{on: colorEnabled(*noColor)}

	path := *pathFlag
	if path == "" {
		path = defaultRegistryPath()
	}
	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ bad registry path: %v\n", err)
		os.Exit(1)
	}

	reg, seeded, err := loadRegistry(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n  Refusing to start so the existing file is not overwritten.\n", err)
		os.Exit(1)
	}

	app := &App{
		reg:   reg,
		path:  path,
		dirty: seeded,
		con:   &console{in: bufio.NewReader(os.Stdin), out: os.Stdout},
		width: termWidth(),
	}

	fmt.Print(banner())
	if seeded {
		app.con.note("No registry found. Seeded a new campus at %s", path)
		if err := app.commit(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ could not create registry: %v\n", err)
			os.Exit(1)
		}
	} else {
		app.con.note("Loaded registry from %s", path)
	}

	// Commit on Ctrl-C / SIGTERM as well, so an interrupt never loses work.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nInterrupt received — committing.")
		app.commitAndExit()
	}()

	app.run()
}
