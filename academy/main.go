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
//	studyhall.go   concept reader, glossary, resources, blueprints, quizzes
//	curriculum.go  the Study Hall knowledge base (technical layer)
//	explainers.go  beginner layer: analogies, real-life examples, glossary
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
	"sort"
	"strings"
	"syscall"
	"time"
)

// registryFileName is the on-disk name of the flat JSON database.
const registryFileName = "academy_campus_registry.json"

// schemaVersion is bumped whenever the JSON layout changes incompatibly.
// Older files are upgraded on load by migrate(); see v1StageRenumber.
const schemaVersion = 2

// v1StageRenumber maps schema-1 stage IDs (the original seven stages) to
// their place in the 16-stage curriculum introduced by schema 2.
var v1StageRenumber = map[int]int{1: 5, 2: 6, 3: 7, 4: 8, 5: 13, 6: 15, 7: 16}

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
	ID              int               `json:"id"`
	Title           string            `json:"title"`
	Status          Status            `json:"status"`
	Literature      []Literature      `json:"required_literature"`
	ConceptsStudied []string          `json:"concepts_studied"`
	ExercisesDone   []string          `json:"exercises_done"`
	Notes           map[string]string `json:"notes"` // concept → the learner's own explanation
	MasteryCheck    *MasteryCheck     `json:"mastery_check,omitempty"`
	TypeInDone      bool              `json:"type_in_done,omitempty"`
	Labs            []Lab             `json:"labs"`
}

// contains reports whether list holds item.
func contains(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}

// setMember adds or removes item from list, keeping it free of duplicates.
func setMember(list []string, item string, present bool) []string {
	kept := list[:0]
	for _, x := range list {
		if x != item {
			kept = append(kept, x)
		}
	}
	if present {
		kept = append(kept, item)
	}
	return kept
}

// hasStudied reports whether the named concept is marked as understood.
func (s *Stage) hasStudied(concept string) bool { return contains(s.ConceptsStudied, concept) }

// setStudied marks or unmarks a concept as understood.
func (s *Stage) setStudied(concept string, studied bool) {
	s.ConceptsStudied = setMember(s.ConceptsStudied, concept, studied)
}

// exerciseKey identifies exercise i (0-based) of a concept in exercises_done.
func exerciseKey(concept string, i int) string { return fmt.Sprintf("%s #%d", concept, i+1) }

// hasDone reports whether an exercise is ticked off.
func (s *Stage) hasDone(concept string, i int) bool {
	return contains(s.ExercisesDone, exerciseKey(concept, i))
}

// setDone ticks or unticks an exercise.
func (s *Stage) setDone(concept string, i int, done bool) {
	s.ExercisesDone = setMember(s.ExercisesDone, exerciseKey(concept, i), done)
}

// exerciseProgress returns (done, total) across all of a stage's concepts.
func exerciseProgress(s *Stage) (int, int) {
	g, ok := guideFor(s.ID)
	if !ok {
		return 0, 0
	}
	done, total := 0, 0
	for _, c := range g.Concepts {
		for i := range c.Exercises {
			total++
			if s.hasDone(c.Name, i) {
				done++
			}
		}
	}
	return done, total
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
	SchemaVersion int                   `json:"schema_version"`
	LastCommit    time.Time             `json:"last_commit"`
	NextLabID     int                   `json:"next_lab_id"`
	Tracks        []Track               `json:"tracks"`
	Reviews       map[string]ReviewCard `json:"reviews"`        // card ID → spaced-repetition state
	ReviewHistory map[string]int        `json:"review_history"` // local date → cards reviewed that day
	TidyScreen    bool                  `json:"tidy_screen"`    // start each action on a clean screen
	StudySessions []StudySession        `json:"study_sessions"` // focus-timer sessions
	Goals         Goals                 `json:"goals"`          // weekly targets (zero: none)
	Achievements  map[string]time.Time  `json:"achievements"`   // achievement ID → when earned

	// Classic Mode (see classic.go).
	ClassicMode    bool                 `json:"classic_mode"`
	Notebook       []NotebookEntry      `json:"notebook"`
	ExerciseOpened map[string]time.Time `json:"exercise_opened"` // exercise key → first opened (the struggle clock)
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
	exercises, exercisesDone  int
	masteryPoints, masteryMax int
	mastered                  int
	due                       int
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
			done, exercises := exerciseProgress(s)
			cs.exercisesDone += done
			cs.exercises += exercises
			pts, maxPts, mastered, _ := r.stageMastery(s)
			cs.masteryPoints += pts
			cs.masteryMax += maxPts
			cs.mastered += mastered
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
	// Focus-timer sessions count as study time too.
	for _, ss := range r.StudySessions {
		cs.hours += ss.Minutes / 60
		ago := int(math.Round(today.Sub(dayStart(ss.Start)).Hours() / 24))
		if ago >= 0 && ago < activityDays {
			cs.activity[activityDays-1-ago] += ss.Minutes / 60
		}
	}
	cs.hours = roundHours(cs.hours)
	cs.due = len(r.dueCards(now))

	// Streak: consecutive days with logged hours or reviews, ending today
	// (or yesterday, so the streak survives until you study today).
	active := func(i int) bool {
		day := today.AddDate(0, 0, i-(activityDays-1)).Format("2006-01-02")
		return cs.activity[i] > 0 || r.ReviewHistory[day] > 0
	}
	i := activityDays - 1
	if i >= 0 && !active(i) {
		i--
	}
	for ; i >= 0 && active(i); i-- {
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
			if s.ExercisesDone == nil {
				s.ExercisesDone = []string{}
			}
			if s.Notes == nil {
				s.Notes = map[string]string{}
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
	if r.Reviews == nil {
		r.Reviews = map[string]ReviewCard{}
	}
	if r.ReviewHistory == nil {
		r.ReviewHistory = map[string]int{}
	}
	if r.StudySessions == nil {
		r.StudySessions = []StudySession{}
	}
	if r.Notebook == nil {
		r.Notebook = []NotebookEntry{}
	}
	if r.ExerciseOpened == nil {
		r.ExerciseOpened = map[string]time.Time{}
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
		return Stage{ID: id, Title: title, Status: StatusActive, Literature: lit,
			ConceptsStudied: []string{}, ExercisesDone: []string{}, Notes: map[string]string{}, Labs: []Lab{}}
	}

	return &Registry{
		SchemaVersion:  schemaVersion,
		NextLabID:      1,
		Reviews:        map[string]ReviewCard{},
		ReviewHistory:  map[string]int{},
		Notebook:       []NotebookEntry{},
		ExerciseOpened: map[string]time.Time{},
		Tracks: []Track{
			{
				ID:   "F",
				Name: "Foundations of Computing",
				Stages: []Stage{
					stage(1, "Programming Fundamentals",
						book("How to Design Programs", "Felleisen, Findler, Flatt & Krishnamurthi"),
						book("Structure and Interpretation of Computer Programs", "Abelson & Sussman"),
						book("Think Python", "Allen B. Downey"),
					),
					stage(2, "Data Structures & Algorithms",
						book("Introduction to Algorithms (CLRS)", "Cormen, Leiserson, Rivest & Stein"),
						book("The Algorithm Design Manual", "Steven Skiena"),
						book("Grokking Algorithms", "Aditya Bhargava"),
					),
					stage(3, "Discrete Mathematics, Logic & Probability",
						book("Mathematics for Computer Science", "Lehman, Leighton & Meyer"),
						book("How to Prove It", "Daniel Velleman"),
						book("Discrete Mathematics and Its Applications", "Kenneth Rosen"),
					),
					stage(4, "Digital Logic & Computer Architecture",
						book("Code: The Hidden Language of Computer Hardware and Software", "Charles Petzold"),
						book("Digital Design and Computer Architecture", "Harris & Harris"),
						book("Computer Organization and Design", "Patterson & Hennessy"),
					),
				},
			},
			{
				ID:   "A",
				Name: "System Core Foundations",
				Stages: []Stage{
					stage(5, "The Iron Layer (Low-Level Systems & Compilers)",
						book("Computer Systems: A Programmer's Perspective", "Bryant & O'Hallaron"),
						book("Crafting Interpreters", "Robert Nystrom"),
						book("Compilers: Principles, Techniques, and Tools", "Aho, Lam, Sethi & Ullman"),
					),
					stage(6, "Operating Systems Internals & Memory Layouts",
						book("Operating Systems: Three Easy Pieces", "Arpaci-Dusseau & Arpaci-Dusseau"),
						paper("What Every Programmer Should Know About Memory", "Ulrich Drepper"),
						book("Understanding the Linux Kernel", "Bovet & Cesati"),
					),
					stage(7, "Storage Engines & State Persistence",
						book("Database Internals", "Alex Petrov"),
						book("Designing Data-Intensive Applications", "Martin Kleppmann"),
						paper("The Log-Structured Merge-Tree (LSM-Tree)", "O'Neil, Cheng, Gawlick & O'Neil"),
					),
					stage(8, "Networks, Sockets, & Distributed Topology",
						book("TCP/IP Illustrated, Vol. 1", "W. Richard Stevens"),
						book("UNIX Network Programming, Vol. 1", "W. Richard Stevens"),
						paper("Time, Clocks, and the Ordering of Events in a Distributed System", "Leslie Lamport"),
						paper("In Search of an Understandable Consensus Algorithm (Raft)", "Ongaro & Ousterhout"),
					),
				},
			},
			{
				ID:   "S",
				Name: "Software, Security & Theory",
				Stages: []Stage{
					stage(9, "Software Engineering & Professional Tools",
						book("The Pragmatic Programmer", "Hunt & Thomas"),
						book("A Philosophy of Software Design", "John Ousterhout"),
						book("Pro Git", "Chacon & Straub"),
					),
					stage(10, "Security & Cryptography",
						book("Security Engineering", "Ross Anderson"),
						book("Serious Cryptography", "Jean-Philippe Aumasson"),
						book("The Web Application Hacker's Handbook", "Stuttard & Pinto"),
					),
					stage(11, "Theory of Computation & Complexity",
						book("Introduction to the Theory of Computation", "Michael Sipser"),
						paper("On Computable Numbers, with an Application to the Entscheidungsproblem", "Alan Turing"),
					),
					stage(12, "Programming Languages & Paradigms",
						book("Essentials of Programming Languages", "Friedman & Wand"),
						book("Types and Programming Languages", "Benjamin Pierce"),
						paper("Why Functional Programming Matters", "John Hughes"),
					),
				},
			},
			{
				ID:   "B",
				Name: "Advanced AI & Hardware Stack",
				Stages: []Stage{
					stage(13, "Mathematical Foundations (Matrix Calculus & Linear Algebra)",
						book("Linear Algebra Done Right", "Sheldon Axler"),
						book("Mathematics for Machine Learning", "Deisenroth, Faisal & Ong"),
						paper("The Matrix Calculus You Need for Deep Learning", "Parr & Howard"),
					),
					stage(14, "Probability, Statistics & Classical Machine Learning",
						book("An Introduction to Statistical Learning", "James, Witten, Hastie & Tibshirani"),
						book("Think Stats", "Allen B. Downey"),
						book("Pattern Recognition and Machine Learning", "Christopher Bishop"),
					),
					stage(15, "Neural Architectures & Autograd from Scratch",
						book("Deep Learning", "Goodfellow, Bengio & Courville"),
						paper("Automatic Differentiation in Machine Learning: a Survey", "Baydin, Pearlmutter, Radul & Siskind"),
						paper("Attention Is All You Need", "Vaswani et al."),
					),
					stage(16, "AI Infrastructure, CUDA, & Memory-Bound Inference",
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

// loadResult says what loadRegistry had to do beyond reading the file.
type loadResult struct {
	Seeded    bool // no file existed; a fresh campus was created
	Migrated  bool // the file used an older schema and was upgraded
	NewStages int  // stages added from the current curriculum
}

// loadRegistry reads the registry from path. If the file does not exist, a
// freshly seeded registry is returned. Older schemas are migrated, and any
// stages the curriculum gained since the file was written are merged in.
// A corrupt file is an error: it is never silently overwritten.
func loadRegistry(path string) (*Registry, loadResult, error) {
	var res loadResult
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		res.Seeded = true
		return seedRegistry(), res, nil
	}
	if err != nil {
		return nil, res, fmt.Errorf("read %s: %w", path, err)
	}
	reg := &Registry{}
	if err := json.Unmarshal(data, reg); err != nil {
		return nil, res, fmt.Errorf("parse %s: %w", path, err)
	}
	if res.Migrated, err = reg.migrate(); err != nil {
		return nil, res, fmt.Errorf("migrate %s: %w", path, err)
	}
	res.NewStages = reg.reconcile(seedRegistry())
	if err := reg.validate(); err != nil {
		return nil, res, fmt.Errorf("validate %s: %w", path, err)
	}
	return reg, res, nil
}

// migrate upgrades an older schema in place and reports whether it did.
func (r *Registry) migrate() (bool, error) {
	switch r.SchemaVersion {
	case schemaVersion:
		return false, nil
	case 1:
		// Schema 2 inserted the Foundations track in front of the original
		// stages, so their IDs move. Labs keep their own IDs.
		for ti := range r.Tracks {
			for si := range r.Tracks[ti].Stages {
				s := &r.Tracks[ti].Stages[si]
				if to, ok := v1StageRenumber[s.ID]; ok {
					s.ID = to
				}
			}
		}
		r.SchemaVersion = 2
		return true, nil
	default:
		return false, fmt.Errorf("unsupported schema_version %d (this build understands up to %d)", r.SchemaVersion, schemaVersion)
	}
}

// reconcile merges in any tracks and stages from seed that r lacks, then
// orders tracks and stages as the seed does (unknown ones keep their place
// at the end). Existing progress is never touched. It returns the number of
// stages added.
func (r *Registry) reconcile(seed *Registry) int {
	added := 0
	for _, st := range seed.Tracks {
		ti := -1
		for i := range r.Tracks {
			if r.Tracks[i].ID == st.ID {
				ti = i
				break
			}
		}
		if ti < 0 {
			r.Tracks = append(r.Tracks, st)
			added += len(st.Stages)
			continue
		}
		t := &r.Tracks[ti]
		for _, ss := range st.Stages {
			if _, existing := r.findStage(ss.ID); existing == nil {
				t.Stages = append(t.Stages, ss)
				added++
			}
		}
		sortByRank(t.Stages, func(s Stage) int {
			for i, ss := range st.Stages {
				if ss.ID == s.ID {
					return i
				}
			}
			return len(st.Stages) + s.ID
		})
	}
	trackRank := func(t Track) int {
		for i, st := range seed.Tracks {
			if st.ID == t.ID {
				return i
			}
		}
		return len(seed.Tracks)
	}
	sortByRank(r.Tracks, trackRank)
	return added
}

// sortByRank stably sorts items by the rank function.
func sortByRank[T any](items []T, rank func(T) int) {
	sort.SliceStable(items, func(i, j int) bool { return rank(items[i]) < rank(items[j]) })
}

// atomicWriteJSON serializes v and swaps it into place at path.
//
// Sequence: write temp file in the same directory -> fsync -> close ->
// rename over target (atomic on POSIX and on NTFS via MoveFileEx) -> fsync
// the directory so the rename itself is durable.
func atomicWriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return atomicWriteFile(path, append(data, '\n'))
}

// atomicWriteFile writes data to path with the same temp-file, fsync and
// rename sequence, so the file is never seen half-written.
func atomicWriteFile(path string, data []byte) (err error) {
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
	checkLinks := flag.Bool("check-links", false, "verify every resource URL over the network, then exit")
	listBackupsFlag := flag.Bool("backups", false, "list the automatic backups of the registry, then exit")
	restoreFlag := flag.String("restore", "", "restore the registry from a backup (a number from -backups, or a file path), then exit")
	flag.Parse()

	sty = Style{on: colorEnabled(*noColor)}

	if *checkLinks {
		if failed := runLinkCheck(os.Stdout); failed > 0 {
			os.Exit(1)
		}
		return
	}

	path := *pathFlag
	if path == "" {
		path = defaultRegistryPath()
	}
	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ bad registry path: %v\n", err)
		os.Exit(1)
	}

	if *listBackupsFlag || *restoreFlag != "" {
		os.Exit(backupCommand(path, *listBackupsFlag, *restoreFlag))
	}

	reg, loaded, err := loadRegistry(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n  Refusing to start so the existing file is not overwritten.\n", err)
		os.Exit(1)
	}

	reg.awardAchievements(time.Now()) // silently backfill earlier milestones

	app := &App{
		reg:   reg,
		path:  path,
		dirty: loaded.Seeded || loaded.Migrated || loaded.NewStages > 0,
		con: &console{
			in:     bufio.NewReader(os.Stdin),
			out:    os.Stdout,
			raw:    interactive() && rawInputSupported(int(os.Stdin.Fd())),
			screen: stdoutIsTerminal(),
		},
		width: termWidth(),
		pager: interactive(),
	}

	app.con.cols = app.cols
	layoutWidth = app.cols
	if app.con.raw && app.con.screen {
		if resized := watchResize(); resized != nil {
			go func() {
				for range resized {
					// Let a burst of resize events (dragging a window edge) settle.
					time.Sleep(80 * time.Millisecond)
					for len(resized) > 0 {
						<-resized
					}
					app.con.redrawForResize()
				}
			}()
		}
	}
	fmt.Print(banner(app.cols()))
	if loaded.Seeded {
		app.con.note("No registry found. Seeded a new campus at %s", path)
		app.con.ok("Welcome! New here? Press %s for the Start Here guide.", sty.Bold(sty.Green("0")))
		if err := app.commit(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ could not create registry: %v\n", err)
			os.Exit(1)
		}
	} else {
		app.con.note("Loaded registry from %s", path)
		if loaded.Migrated {
			app.con.ok("Upgraded your registry to the 16-stage curriculum; your progress and labs are kept.")
		}
		if loaded.NewStages > 0 {
			app.con.ok("%d new stage(s) added to your campus. Commit (5 or 7) to save them.", loaded.NewStages)
		}
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
