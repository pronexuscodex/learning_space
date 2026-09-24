// Command academy is a zero-dependency terminal ledger for a private
// "Systems & AI Academy" curriculum.
//
// All state lives in a single flat JSON file (academy_campus_registry.json)
// stored next to the binary. The file is human-readable by design: you can
// open, diff, or version-control it at any time. Writes are atomic: the new
// state is written to a temporary file in the same directory, fsynced, and
// renamed over the old file, so a crash mid-write can never leave a
// half-written registry behind.
//
// Build:
//
//	go build -trimpath -ldflags="-s -w" -o academy .
//
// Run:
//
//	./academy                      # registry beside the binary
//	./academy -registry ./my.json  # explicit registry path
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

// registryFileName is the on-disk name of the flat JSON database.
const registryFileName = "academy_campus_registry.json"

// schemaVersion is bumped whenever the JSON layout changes incompatibly.
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
	ID         int          `json:"id"`
	Title      string       `json:"title"`
	Status     Status       `json:"status"`
	Literature []Literature `json:"required_literature"`
	Labs       []Lab        `json:"labs"`
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

// validate checks structural invariants after loading from disk and repairs
// the lab ID counter if it has drifted behind existing IDs.
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
	for _, t := range r.Tracks {
		for _, s := range t.Stages {
			if stageIDs[s.ID] {
				return fmt.Errorf("duplicate stage id %d", s.ID)
			}
			stageIDs[s.ID] = true
			if !s.Status.Valid() {
				return fmt.Errorf("stage %d has invalid status %q", s.ID, s.Status)
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
				if l.ID > maxLab {
					maxLab = l.ID
				}
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
		return Stage{ID: id, Title: title, Status: StatusActive, Literature: lit, Labs: []Lab{}}
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

// ---------------------------------------------------------------------------
// Console input (raw line polling over bufio)
// ---------------------------------------------------------------------------

// errCancel signals that the user backed out of a prompt.
var errCancel = errors.New("cancelled")

// console wraps a buffered input stream and an output writer.
type console struct {
	in  *bufio.Reader
	out io.Writer
}

// sanitize strips control characters (tabs become spaces) and surrounding
// whitespace from a raw input line.
func sanitize(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case r == utf8.RuneError, unicode.IsControl(r):
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

// isCancel reports whether the input is a cancel token.
func isCancel(s string) bool {
	s = strings.ToLower(s)
	return s == "q" || s == ":q"
}

// readLine prints prompt and returns one sanitized line. It returns io.EOF
// when input is exhausted (Ctrl-D / closed pipe).
func (c *console) readLine(prompt string) (string, error) {
	fmt.Fprint(c.out, prompt)
	line, err := c.in.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line != "" {
			return sanitize(line), nil // final unterminated line
		}
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(c.out)
		}
		return "", err
	}
	return sanitize(line), nil
}

// promptText asks for free text. ":q" cancels. Required fields reject blank
// input; over-long input is rejected rather than silently truncated.
func (c *console) promptText(label string, maxLen int, required bool) (string, error) {
	for {
		s, err := c.readLine(label + " (:q to cancel): ")
		if err != nil {
			return "", err
		}
		switch {
		case strings.EqualFold(s, ":q"):
			return "", errCancel
		case s == "" && required:
			fmt.Fprintln(c.out, "  ! This field is required.")
		case utf8.RuneCountInString(s) > maxLen:
			fmt.Fprintf(c.out, "  ! Too long (%d chars, max %d).\n", utf8.RuneCountInString(s), maxLen)
		default:
			return s, nil
		}
	}
}

// promptInt asks for an integer accepted by valid. "q" cancels.
func (c *console) promptInt(label string, valid func(int) error) (int, error) {
	for {
		s, err := c.readLine(label + " (q to cancel): ")
		if err != nil {
			return 0, err
		}
		if isCancel(s) {
			return 0, errCancel
		}
		n, convErr := strconv.Atoi(strings.TrimPrefix(s, "#"))
		if convErr != nil {
			fmt.Fprintf(c.out, "  ! %q is not a whole number.\n", s)
			continue
		}
		if vErr := valid(n); vErr != nil {
			fmt.Fprintf(c.out, "  ! %v\n", vErr)
			continue
		}
		return n, nil
	}
}

// promptChoice shows a numbered list and returns the chosen zero-based index.
func (c *console) promptChoice(label string, options []string) (int, error) {
	for i, opt := range options {
		fmt.Fprintf(c.out, "  %d) %s\n", i+1, opt)
	}
	n, err := c.promptInt(label, func(n int) error {
		if n < 1 || n > len(options) {
			return fmt.Errorf("choose a number between 1 and %d", len(options))
		}
		return nil
	})
	return n - 1, err
}

// parseHours accepts decimal hours ("1.5") or a Go duration ("1h30m", "45m").
func parseHours(s string) (float64, error) {
	if h, err := strconv.ParseFloat(s, 64); err == nil {
		if math.IsNaN(h) || math.IsInf(h, 0) {
			return 0, errors.New("not a finite number")
		}
		return h, nil
	}
	d, err := time.ParseDuration(strings.ToLower(strings.ReplaceAll(s, " ", "")))
	if err != nil {
		return 0, fmt.Errorf("%q is not a number of hours (try 2.5 or 1h30m)", s)
	}
	return d.Hours(), nil
}

// promptHours asks for an hour amount within [min, max]. When allowZero is
// false the value must be strictly positive.
func (c *console) promptHours(label string, max float64, allowZero bool) (float64, error) {
	for {
		s, err := c.readLine(label + " (q to cancel): ")
		if err != nil {
			return 0, err
		}
		if isCancel(s) {
			return 0, errCancel
		}
		if s == "" && allowZero {
			return 0, nil
		}
		h, perr := parseHours(s)
		switch {
		case perr != nil:
			fmt.Fprintf(c.out, "  ! %v\n", perr)
		case h < 0:
			fmt.Fprintln(c.out, "  ! Hours cannot be negative.")
		case h == 0 && !allowZero:
			fmt.Fprintln(c.out, "  ! Hours must be greater than zero.")
		case h > max:
			fmt.Fprintf(c.out, "  ! Maximum is %.0f hours per entry.\n", max)
		default:
			return roundHours(h), nil
		}
	}
}

// confirm asks a yes/no question and loops until it gets one.
func (c *console) confirm(label string) (bool, error) {
	for {
		s, err := c.readLine(label + " [y/n]: ")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(s) {
		case "y", "yes":
			return true, nil
		case "n", "no", "q":
			return false, nil
		}
		fmt.Fprintln(c.out, "  ! Please answer y or n.")
	}
}

// roundHours keeps hours at two decimal places to avoid float drift.
func roundHours(h float64) float64 { return math.Round(h*100) / 100 }

// ---------------------------------------------------------------------------
// Application
// ---------------------------------------------------------------------------

// App owns the in-memory registry. mu guards reg and dirty so a signal
// handler can safely commit while the main loop is blocked on input.
type App struct {
	mu    sync.Mutex
	reg   *Registry
	path  string
	dirty bool
	con   *console
}

// commit atomically persists the registry to disk.
func (a *App) commit() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reg.LastCommit = time.Now().UTC().Truncate(time.Second)
	if err := atomicWriteJSON(a.path, a.reg); err != nil {
		return err
	}
	a.dirty = false
	return nil
}

// mutate runs fn under the lock and marks state dirty.
func (a *App) mutate(fn func(r *Registry)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	fn(a.reg)
	a.dirty = true
}

func (a *App) printf(format string, args ...any) { fmt.Fprintf(a.con.out, format, args...) }

const rule = "════════════════════════════════════════════════════════════════════════"

func (a *App) printMenu() {
	a.printf("\n%s\n  SYSTEMS & AI ACADEMY  ::  Campus Registry\n%s\n", rule, rule)
	a.printf("  1) View Campus Ledger\n")
	a.printf("  2) Enroll in a New Lab\n")
	a.printf("  3) Log Study/Lab Hours\n")
	a.printf("  4) Advance Academic Status\n")
	a.printf("  5) Atomic Commit & Exit\n")
	a.printf("  6) Checkpoint (commit and keep working)\n")
	a.printf("  7) Exit WITHOUT saving\n")
	if a.dirty {
		a.printf("  [uncommitted changes]\n")
	}
}

// statusBadge renders a compact marker for a status.
func statusBadge(s Status) string {
	if s == StatusGraduated {
		return "[GRADUATED]"
	}
	return "[ACTIVE]   "
}

// viewLedger prints every track, stage, reading list and lab.
func (a *App) viewLedger() {
	a.mu.Lock()
	defer a.mu.Unlock()

	var grand float64
	for _, t := range a.reg.Tracks {
		var trackHours float64
		a.printf("\n%s\n TRACK %s: %s\n%s\n", rule, t.ID, strings.ToUpper(t.Name), rule)
		for si := range t.Stages {
			s := &t.Stages[si]
			read, total := literatureProgress(s)
			hours := stageHours(s)
			trackHours += hours
			a.printf("\n %s Stage %d: %s\n", statusBadge(s.Status), s.ID, s.Title)
			a.printf("   Literature %d/%d read | %d lab(s) | %.2fh logged\n", read, total, len(s.Labs), hours)
			for _, lit := range s.Literature {
				mark := " "
				if lit.Read {
					mark = "x"
				}
				a.printf("     [%s] %s — %s (%s)\n", mark, lit.Title, lit.Author, lit.Kind)
			}
			if len(s.Labs) == 0 {
				a.printf("     · no labs enrolled yet\n")
			}
			for _, l := range s.Labs {
				a.printf("     ▸ Lab #%d  %-32s %8.2fh  %-14s %s\n",
					l.ID, l.Name, l.HoursLogged, l.CompilationStatus, l.Status)
				if l.Architecture != "" {
					a.printf("         arch: %s\n", l.Architecture)
				}
			}
		}
		a.printf("\n   Track %s cumulative: %.2fh\n", t.ID, roundHours(trackHours))
		grand += trackHours
	}
	a.printf("\n%s\n  Campus total: %.2fh across %d lab(s)  |  last commit: %s\n",
		rule, roundHours(grand), a.reg.labCount(), formatTime(a.reg.LastCommit))
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Local().Format("2006-01-02 15:04")
}

// listStages prints a compact stage index, optionally restricted to a track.
func (a *App) listStages(track *Track) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, t := range a.reg.Tracks {
		if track != nil && t.ID != track.ID {
			continue
		}
		a.printf("  Track %s — %s\n", t.ID, t.Name)
		for _, s := range t.Stages {
			a.printf("    %d) %s  %s\n", s.ID, statusBadge(s.Status), s.Title)
		}
	}
}

// listLabs prints every lab; it returns false if none exist.
func (a *App) listLabs() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.reg.labCount() == 0 {
		a.printf("  No labs enrolled yet. Use option 2 to enroll one.\n")
		return false
	}
	for _, t := range a.reg.Tracks {
		for _, s := range t.Stages {
			for _, l := range s.Labs {
				a.printf("    #%-3d %-32s stage %d  %8.2fh  %s\n", l.ID, l.Name, s.ID, l.HoursLogged, l.Status)
			}
		}
	}
	return true
}

// pickTrack prompts for a track by letter ("A") or list number ("1").
func (a *App) pickTrack() (*Track, error) {
	a.mu.Lock()
	for i, t := range a.reg.Tracks {
		a.printf("  %d) Track %s: %s\n", i+1, t.ID, t.Name)
	}
	a.mu.Unlock()
	for {
		s, err := a.con.readLine("Track letter or number (q to cancel): ")
		if err != nil {
			return nil, err
		}
		if isCancel(s) {
			return nil, errCancel
		}
		a.mu.Lock()
		for i := range a.reg.Tracks {
			t := &a.reg.Tracks[i]
			if strings.EqualFold(s, t.ID) || s == strconv.Itoa(i+1) {
				a.mu.Unlock()
				return t, nil
			}
		}
		a.mu.Unlock()
		a.printf("  ! %q is not a track.\n", s)
	}
}

// pickStage prompts for a stage ID, optionally restricted to one track.
func (a *App) pickStage(track *Track) (int, error) {
	a.listStages(track)
	return a.con.promptInt("Stage ID", func(id int) error {
		a.mu.Lock()
		defer a.mu.Unlock()
		t, s := a.reg.findStage(id)
		if s == nil || (track != nil && t.ID != track.ID) {
			return fmt.Errorf("stage %d is not in the list above", id)
		}
		return nil
	})
}

// pickLab prompts for an existing lab ID.
func (a *App) pickLab() (int, error) {
	if !a.listLabs() {
		return 0, errCancel
	}
	return a.con.promptInt("Lab ID", func(id int) error {
		a.mu.Lock()
		defer a.mu.Unlock()
		if _, l := a.reg.findLab(id); l == nil {
			return fmt.Errorf("no lab with id %d", id)
		}
		return nil
	})
}

// enrollLab gathers parameters for a new lab and appends it to a stage.
func (a *App) enrollLab() error {
	a.printf("\n-- Enroll in a New Lab --\n")

	track, err := a.pickTrack()
	if err != nil {
		return err
	}

	stageID, err := a.pickStage(track)
	if err != nil {
		return err
	}

	var name string
	for {
		name, err = a.con.promptText("Lab name", maxNameLen, true)
		if err != nil {
			return err
		}
		a.mu.Lock()
		_, s := a.reg.findStage(stageID)
		dup := false
		for _, l := range s.Labs {
			if strings.EqualFold(l.Name, name) {
				dup = true
				break
			}
		}
		a.mu.Unlock()
		if !dup {
			break
		}
		a.printf("  ! Stage %d already has a lab named %q.\n", stageID, name)
	}

	notes, err := a.con.promptText("Architecture notes (optional)", maxNotesLen, false)
	if err != nil {
		return err
	}
	hours, err := a.con.promptHours(fmt.Sprintf("Initial lab hours [0-%.0f, blank = 0]", maxInitialHours), maxInitialHours, true)
	if err != nil {
		return err
	}

	var id int
	a.mutate(func(r *Registry) {
		_, s := r.findStage(stageID)
		now := time.Now().UTC().Truncate(time.Second)
		lab := Lab{
			ID:                r.NextLabID,
			Name:              name,
			Architecture:      notes,
			HoursLogged:       hours,
			CompilationStatus: CompileNotBuilt,
			Status:            StatusActive,
			EnrolledAt:        now,
			Log:               []HourEntry{},
		}
		if hours > 0 {
			lab.Log = append(lab.Log, HourEntry{At: now, Hours: hours, Note: "initial hours"})
		}
		s.Labs = append(s.Labs, lab)
		id = r.NextLabID
		r.NextLabID++
	})
	a.printf("  ✓ Enrolled lab #%d %q in stage %d.\n", id, name, stageID)
	return nil
}

// logHours adds an increment of focused time to an existing lab.
func (a *App) logHours() error {
	a.printf("\n-- Log Study/Lab Hours --\n")
	labID, err := a.pickLab()
	if err != nil {
		return err
	}
	hours, err := a.con.promptHours(fmt.Sprintf("Hours to add (e.g. 1.5 or 1h30m, max %.0f)", maxHoursPerEntry), maxHoursPerEntry, false)
	if err != nil {
		return err
	}
	note, err := a.con.promptText("Session note (optional)", maxLogNoteLen, false)
	if err != nil {
		return err
	}

	var total float64
	var name string
	a.mutate(func(r *Registry) {
		_, l := r.findLab(labID)
		l.HoursLogged = roundHours(l.HoursLogged + hours)
		l.Log = append(l.Log, HourEntry{At: time.Now().UTC().Truncate(time.Second), Hours: hours, Note: note})
		total, name = l.HoursLogged, l.Name
	})
	a.printf("  ✓ +%.2fh on %q — cumulative %.2fh.\n", hours, name, total)
	return nil
}

// advanceStatus is the sub-menu for status transitions.
func (a *App) advanceStatus() error {
	a.printf("\n-- Advance Academic Status --\n")
	choice, err := a.con.promptChoice("Action", []string{
		"Toggle a Stage (Active Research <-> Mastered/Graduated)",
		"Toggle a Lab (Active Research <-> Mastered/Graduated)",
		"Mark required literature read/unread",
		"Set a Lab's compilation status",
	})
	if err != nil {
		return err
	}
	switch choice {
	case 0:
		return a.toggleStage()
	case 1:
		return a.toggleLab()
	case 2:
		return a.toggleLiterature()
	default:
		return a.setCompileStatus()
	}
}

func (a *App) toggleStage() error {
	id, err := a.pickStage(nil)
	if err != nil {
		return err
	}

	a.mu.Lock()
	_, s := a.reg.findStage(id)
	next := s.Status.Toggled()
	read, total := literatureProgress(s)
	activeLabs := 0
	for _, l := range s.Labs {
		if l.Status != StatusGraduated {
			activeLabs++
		}
	}
	a.mu.Unlock()

	if next == StatusGraduated && (read < total || activeLabs > 0) {
		a.printf("  ! Stage %d still has %d unread text(s) and %d active lab(s).\n", id, total-read, activeLabs)
		ok, err := a.con.confirm("  Graduate anyway?")
		if err != nil {
			return err
		}
		if !ok {
			return errCancel
		}
	}
	a.mutate(func(r *Registry) {
		_, s := r.findStage(id)
		s.Status = next
	})
	a.printf("  ✓ Stage %d is now %s.\n", id, next)
	return nil
}

func (a *App) toggleLab() error {
	id, err := a.pickLab()
	if err != nil {
		return err
	}

	a.mu.Lock()
	_, l := a.reg.findLab(id)
	next := l.Status.Toggled()
	cs := l.CompilationStatus
	a.mu.Unlock()

	if next == StatusGraduated && cs != CompileOK && cs != CompileTested {
		a.printf("  ! Lab #%d is %q.\n", id, cs)
		ok, err := a.con.confirm("  Graduate a lab that does not compile?")
		if err != nil {
			return err
		}
		if !ok {
			return errCancel
		}
	}
	a.mutate(func(r *Registry) {
		_, l := r.findLab(id)
		l.Status = next
	})
	a.printf("  ✓ Lab #%d is now %s.\n", id, next)
	return nil
}

func (a *App) toggleLiterature() error {
	stageID, err := a.pickStage(nil)
	if err != nil {
		return err
	}

	a.mu.Lock()
	_, s := a.reg.findStage(stageID)
	options := make([]string, len(s.Literature))
	for i, lit := range s.Literature {
		mark := " "
		if lit.Read {
			mark = "x"
		}
		options[i] = fmt.Sprintf("[%s] %s — %s", mark, lit.Title, lit.Author)
	}
	a.mu.Unlock()

	if len(options) == 0 {
		a.printf("  Stage %d has no required literature.\n", stageID)
		return nil
	}
	idx, err := a.con.promptChoice("Toggle which text", options)
	if err != nil {
		return err
	}

	var title string
	var read bool
	a.mutate(func(r *Registry) {
		_, s := r.findStage(stageID)
		lit := &s.Literature[idx]
		lit.Read = !lit.Read
		title, read = lit.Title, lit.Read
	})
	state := "unread"
	if read {
		state = "read"
	}
	a.printf("  ✓ %q marked %s.\n", title, state)
	return nil
}

func (a *App) setCompileStatus() error {
	id, err := a.pickLab()
	if err != nil {
		return err
	}
	options := make([]string, len(compileStatuses))
	for i, cs := range compileStatuses {
		options[i] = string(cs)
	}
	idx, err := a.con.promptChoice("Compilation status", options)
	if err != nil {
		return err
	}
	a.mutate(func(r *Registry) {
		_, l := r.findLab(id)
		l.CompilationStatus = compileStatuses[idx]
	})
	a.printf("  ✓ Lab #%d compilation status: %s.\n", id, compileStatuses[idx])
	return nil
}

// commitAndExit persists state and terminates the process.
func (a *App) commitAndExit() {
	if err := a.commit(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ commit failed: %v\n  (in-memory changes were NOT saved)\n", err)
		os.Exit(1)
	}
	a.printf("✓ Registry committed atomically to %s\n", a.path)
	os.Exit(0)
}

// run is the interactive loop. Ctrl-D (EOF) is treated as commit & exit.
func (a *App) run() {
	for {
		a.printMenu()
		choice, err := a.con.readLine("academy> ")
		if err != nil {
			a.printf("Input closed — committing.\n")
			a.commitAndExit()
		}

		var actionErr error
		switch choice {
		case "1":
			a.viewLedger()
		case "2":
			actionErr = a.enrollLab()
		case "3":
			actionErr = a.logHours()
		case "4":
			actionErr = a.advanceStatus()
		case "5":
			a.commitAndExit()
		case "6":
			if err := a.commit(); err != nil {
				a.printf("  ✗ checkpoint failed: %v\n", err)
			} else {
				a.printf("  ✓ Checkpoint written to %s\n", a.path)
			}
		case "7":
			ok, err := a.con.confirm("Discard all uncommitted changes and exit?")
			if err != nil {
				a.commitAndExit()
			}
			if ok {
				a.printf("Exited without saving.\n")
				os.Exit(0)
			}
		case "":
			// Blank line: just redraw the menu.
		default:
			a.printf("  ! %q is not a menu option. Choose 1-7.\n", choice)
		}

		switch {
		case errors.Is(actionErr, errCancel):
			a.printf("  · Cancelled — nothing changed.\n")
		case errors.Is(actionErr, io.EOF):
			a.printf("Input closed — committing.\n")
			a.commitAndExit()
		case actionErr != nil:
			a.printf("  ✗ %v\n", actionErr)
		}
	}
}

func main() {
	pathFlag := flag.String("registry", "", "path to the JSON registry (default: "+registryFileName+" beside the binary)")
	flag.Parse()

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
	}

	if seeded {
		fmt.Printf("· No registry found. Seeded a new campus at %s\n", path)
		if err := app.commit(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ could not create registry: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("· Loaded registry from %s\n", path)
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
