package main

// Thinking for yourself: the data and rules behind the habits that build
// real understanding, and that handing your thinking to a tool (a search
// engine, a forum, an AI) quietly skips.
//
//   - Solve log: after each exercise the learner says honestly how they got
//     there. An exercise solved with the hint or with someone else's help
//     (an AI, a forum answer) comes back after a few days, to be redone from
//     a blank page: being able to do it alone is the goal.
//   - Blank-page recall: write down everything you remember about a concept
//     before looking (free recall, the strongest form of retrieval
//     practice), then check and score yourself.
//   - Calibration: say how sure you are before revealing a review answer.
//     Knowing what you don't know is a skill of its own (metacognition).
//   - Daily Workout: one guided session a day that strings these together.

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// How an exercise was solved, as the learner reports it.
const (
	SolvedAlone = "alone" // from my own head
	SolvedDocs  = "docs"  // with documentation, books, man pages or my notes
	SolvedHint  = "hint"  // with the exercise's hint
	SolvedHelp  = "help"  // with an AI, a forum answer or someone else's solution
)

// solvedOptions are the choices offered, in order, with their meaning.
var solvedOptions = []struct{ How, Label string }{
	{SolvedAlone, "On my own, from my head"},
	{SolvedDocs, "With documentation, a book, man pages or my notes"},
	{SolvedHint, "With the hint"},
	{SolvedHelp, "With an AI, a forum answer or someone else's solution"},
}

// redoAfter is how long a helped solve waits before it comes back to be
// redone from a blank page: long enough that it is recall, not memory of
// the answer.
const redoAfter = 3 * 24 * time.Hour

// SolveRecord is one honest report of how an exercise got solved.
type SolveRecord struct {
	Key     string    `json:"key"` // exerciseKey(concept, index)
	Stage   int       `json:"stage"`
	Concept string    `json:"concept"`
	Index   int       `json:"index"`
	At      time.Time `json:"at"`
	How     string    `json:"how"`
	Rung    int       `json:"rung"` // highest hint-ladder rung climbed (0: none)
}

// helped reports whether the solve leaned on the answer rather than on the
// learner's own reasoning (documentation counts as the learner's own work:
// reading the manual is what engineers do).
func (s SolveRecord) helped() bool { return s.How == SolvedHint || s.How == SolvedHelp }

// Redo is an exercise to redo from a blank page.
type Redo struct {
	SolveRecord
	Due time.Time
}

// latestSolves returns the most recent solve for every exercise key.
func (r *Registry) latestSolves() map[string]SolveRecord {
	last := map[string]SolveRecord{}
	for _, s := range r.Solves {
		if prev, ok := last[s.Key]; !ok || !s.At.Before(prev.At) {
			last[s.Key] = s
		}
	}
	return last
}

// redos lists exercises whose latest solve was helped, due first. Only
// those already due by now are returned when dueOnly is set.
func (r *Registry) redos(now time.Time, dueOnly bool) []Redo {
	var out []Redo
	for _, s := range r.latestSolves() {
		if !s.helped() {
			continue
		}
		due := s.At.Add(redoAfter)
		if dueOnly && due.After(now) {
			continue
		}
		out = append(out, Redo{s, due})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Due.Equal(out[j].Due) {
			return out[i].Due.Before(out[j].Due)
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// redoPending reports whether the exercise is waiting to be redone.
func (r *Registry) redoPending(key string) bool {
	s, ok := r.latestSolves()[key]
	return ok && s.helped()
}

// SelfReliance summarises how exercises got solved (latest solve of each).
type SelfReliance struct {
	Alone, Docs, Hint, Help int
}

func (s SelfReliance) total() int { return s.Alone + s.Docs + s.Hint + s.Help }

// own is the number solved by the learner's own reasoning (alone or docs).
func (s SelfReliance) own() int { return s.Alone + s.Docs }

func (r *Registry) selfReliance() SelfReliance {
	var s SelfReliance
	for _, rec := range r.latestSolves() {
		switch rec.How {
		case SolvedAlone:
			s.Alone++
		case SolvedDocs:
			s.Docs++
		case SolvedHint:
			s.Hint++
		case SolvedHelp:
			s.Help++
		}
	}
	return s
}

// ---------------------------------------------------------------------------
// Blank-page recall
// ---------------------------------------------------------------------------

// Recall scores, self-assessed after comparing with the concept card.
const (
	RecallLittle = 1 + iota
	RecallSome
	RecallMost
	RecallAll
)

var recallScoreNames = []string{"", "hardly anything", "some of it", "most of it", "all of it"}

// maxRecallLen bounds a written recall.
const maxRecallLen = 4000

// RecallRecord is one blank-page recall attempt.
type RecallRecord struct {
	Stage   int       `json:"stage"`
	Concept string    `json:"concept"`
	At      time.Time `json:"at"`
	Score   int       `json:"score"` // RecallLittle … RecallAll
	Text    string    `json:"text"`
}

// recallInterval is when a concept is ready for another recall, by score:
// weak recalls return the next day, strong ones weeks later.
func recallInterval(score int) time.Duration {
	switch score {
	case RecallLittle:
		return 24 * time.Hour
	case RecallSome:
		return 3 * 24 * time.Hour
	case RecallMost:
		return 10 * 24 * time.Hour
	}
	return 30 * 24 * time.Hour
}

// RecallCandidate is a studied concept that is ready to be recalled.
type RecallCandidate struct {
	Stage   int
	Concept int // index in the stage guide
	Name    string
	Last    *RecallRecord // nil: never recalled
	Due     time.Time
}

// recallCandidates lists understood concepts that are due for a recall,
// most overdue first; never-recalled concepts are due a day after they
// were understood (the timing is unknown, so they count as due now).
func (r *Registry) recallCandidates(now time.Time) []RecallCandidate {
	last := map[string]*RecallRecord{}
	for i := range r.Recalls {
		rec := &r.Recalls[i]
		k := recallKey(rec.Stage, rec.Concept)
		if prev, ok := last[k]; !ok || !rec.At.Before(prev.At) {
			last[k] = rec
		}
	}
	var out []RecallCandidate
	for _, s := range r.stageOrder() {
		g, ok := guideFor(s.ID)
		if !ok {
			continue
		}
		for ci, c := range g.Concepts {
			if !s.hasStudied(c.Name) {
				continue
			}
			cand := RecallCandidate{Stage: s.ID, Concept: ci, Name: c.Name, Last: last[recallKey(s.ID, c.Name)]}
			if cand.Last != nil {
				cand.Due = cand.Last.At.Add(recallInterval(cand.Last.Score))
				if cand.Due.After(now) {
					continue
				}
			}
			out = append(out, cand)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		// Weakest and longest-waiting first; never-recalled after due ones.
		a, b := out[i], out[j]
		if (a.Last == nil) != (b.Last == nil) {
			return a.Last != nil
		}
		return a.Due.Before(b.Due)
	})
	return out
}

func recallKey(stage int, concept string) string {
	return strings.ToLower(concept) + "@" + strconv.Itoa(stage)
}

// ---------------------------------------------------------------------------
// Confidence calibration
// ---------------------------------------------------------------------------

// Confidence levels asked before an answer is revealed.
const (
	ConfGuess = iota
	ConfFair
	ConfSure
)

var confNames = []string{"guessing", "fairly sure", "certain"}

// Calibration counts, per confidence level, the answers given and how many
// were recalled correctly: Calibration[level] = [answers, correct].
type Calibration [3][2]int

func (c *Calibration) record(level int, correct bool) {
	if level < ConfGuess || level > ConfSure {
		return
	}
	c[level][0]++
	if correct {
		c[level][1]++
	}
}

// accuracy is the share recalled correctly at a level, and whether there
// are enough answers (5) to say anything.
func (c Calibration) accuracy(level int) (float64, bool) {
	n, right := c[level][0], c[level][1]
	if n < 5 {
		return 0, false
	}
	return float64(right) / float64(n), true
}

func (c Calibration) answers() int { return c[0][0] + c[1][0] + c[2][0] }

// verdict is a one-sentence reading of the calibration, or "" when there
// is not enough data yet.
func (c Calibration) verdict() string {
	sure, okSure := c.accuracy(ConfSure)
	guess, okGuess := c.accuracy(ConfGuess)
	switch {
	case okSure && sure < 0.8:
		return "Overconfident: when you feel certain you are right only " + percent(sure) + " of the time. Slow down and test yourself before trusting that feeling."
	case okGuess && guess > 0.7:
		return "Underconfident: even your guesses are right " + percent(guess) + " of the time. You know more than you think."
	case okSure:
		return "Well calibrated: when you feel certain you are right " + percent(sure) + " of the time."
	}
	return ""
}

// ---------------------------------------------------------------------------
// Daily Workout
// ---------------------------------------------------------------------------

// Workout steps, in order.
const (
	StepReview   = "review"
	StepRecall   = "recall"
	StepLearn    = "learn"
	StepPractice = "practice"
	StepReflect  = "reflect"
)

var workoutSteps = []string{StepReview, StepRecall, StepLearn, StepPractice, StepReflect}

// WorkoutDay is one day's guided session.
type WorkoutDay struct {
	Date       string   `json:"date"` // local date, 2006-01-02
	Done       []string `json:"done"` // completed (or skipped) steps
	Skipped    []string `json:"skipped,omitempty"`
	Hardest    string   `json:"hardest,omitempty"` // reflection: the hardest thing today
	AskedFirst *bool    `json:"asked_ai_first,omitempty"`
}

func (w *WorkoutDay) has(step string) bool { return contains(w.Done, step) }

func (w *WorkoutDay) complete() bool {
	for _, s := range workoutSteps {
		if !w.has(s) {
			return false
		}
	}
	return true
}

// workoutFor returns today's workout, creating it when missing.
func (r *Registry) workoutFor(now time.Time) *WorkoutDay {
	day := dayStart(now).Format("2006-01-02")
	for i := range r.Workouts {
		if r.Workouts[i].Date == day {
			return &r.Workouts[i]
		}
	}
	r.Workouts = append(r.Workouts, WorkoutDay{Date: day, Done: []string{}})
	return &r.Workouts[len(r.Workouts)-1]
}

// todayWorkout returns today's workout without creating one.
func (r *Registry) todayWorkout(now time.Time) *WorkoutDay {
	day := dayStart(now).Format("2006-01-02")
	for i := range r.Workouts {
		if r.Workouts[i].Date == day {
			return &r.Workouts[i]
		}
	}
	return nil
}

// workoutsCompleted counts days whose workout was finished.
func (r *Registry) workoutsCompleted() int {
	n := 0
	for i := range r.Workouts {
		if r.Workouts[i].complete() {
			n++
		}
	}
	return n
}

// thinkFirstDays counts finished workouts where the learner tried before
// asking an AI.
func (r *Registry) thinkFirstDays() int {
	n := 0
	for _, w := range r.Workouts {
		if w.AskedFirst != nil && !*w.AskedFirst {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

// Colour themes.
const (
	ThemeDefault    = ""
	ThemeContrast   = "contrast"
	ThemeColorblind = "colorblind"
	ThemeMono       = "mono"
)

var themes = []struct{ ID, Name, About string }{
	{ThemeDefault, "Default", "the standard colours"},
	{ThemeContrast, "High contrast", "brighter text, no faint grey"},
	{ThemeColorblind, "Colour-blind friendly", "blue and orange instead of green and red"},
	{ThemeMono, "Monochrome", "no colour, only bold and layout"},
}

// Settings are the learner's display and practice preferences.
type Settings struct {
	Theme        string `json:"theme,omitempty"`
	PlainSymbols bool   `json:"plain_symbols,omitempty"` // replace emoji with plain characters
	NoConfidence bool   `json:"no_confidence,omitempty"` // skip "how sure are you?" in reviews
	DetailedHome bool   `json:"detailed_home,omitempty"` // the full dashboard and menu instead of the simple home
}

// normalize drops values a hand-edited file might hold.
func (s *Settings) normalize() {
	for _, t := range themes {
		if s.Theme == t.ID {
			return
		}
	}
	s.Theme = ThemeDefault
}

// normalizeThinking repairs the thinking-for-yourself data on load.
func (r *Registry) normalizeThinking() {
	valid := map[string]bool{SolvedAlone: true, SolvedDocs: true, SolvedHint: true, SolvedHelp: true}
	kept := r.Solves[:0]
	for _, s := range r.Solves {
		if valid[s.How] && s.Key != "" {
			kept = append(kept, s)
		}
	}
	r.Solves = kept
	recalls := r.Recalls[:0]
	for _, rec := range r.Recalls {
		if rec.Score >= RecallLittle && rec.Score <= RecallAll && rec.Concept != "" {
			recalls = append(recalls, rec)
		}
	}
	r.Recalls = recalls
	for i := range r.Calibration {
		for j := range r.Calibration[i] {
			r.Calibration[i][j] = max(0, r.Calibration[i][j])
		}
		r.Calibration[i][1] = min(r.Calibration[i][1], r.Calibration[i][0])
	}
	for i := range r.Workouts {
		if r.Workouts[i].Done == nil {
			r.Workouts[i].Done = []string{}
		}
	}
	r.Settings.normalize()
}

func percent(f float64) string { return strconv.Itoa(int(f*100+0.5)) + "%" }

// ---------------------------------------------------------------------------
// Choosing what to learn and practise
// ---------------------------------------------------------------------------

// nextConcept is the next concept to learn: the first one not yet
// understood in the current stage.
func (r *Registry) nextConcept() (stage, concept int, ok bool) {
	s := r.currentStage()
	if s == nil {
		return 0, 0, false
	}
	g, found := guideFor(s.ID)
	if !found {
		return 0, 0, false
	}
	for i, c := range g.Concepts {
		if !s.hasStudied(c.Name) {
			return s.ID, i, true
		}
	}
	return 0, 0, false
}

// Practice is an exercise chosen for practice.
type Practice struct {
	Stage, Concept, Exercise int
	Redo                     *Redo // set when it is a redo from a blank page
}

// practicePick chooses today's exercise: a redo that is due comes first;
// otherwise the easiest open exercise of an understood concept, preferably
// from a different stage than avoidStage (mixing topics, interleaving,
// teaches more than practising one topic in a block). pick chooses among
// candidates (pass rand.Intn; tests pass a fixed function).
func (r *Registry) practicePick(now time.Time, avoidStage int, pick func(int) int) (Practice, bool) {
	for _, rd := range r.redos(now, true) {
		g, ok := guideFor(rd.Stage)
		if !ok {
			continue
		}
		for ci, c := range g.Concepts {
			if c.Name == rd.Concept && rd.Index < len(c.Exercises) {
				rd := rd
				return Practice{Stage: rd.Stage, Concept: ci, Exercise: rd.Index, Redo: &rd}, true
			}
		}
	}
	var other, same []Practice
	for _, s := range r.stageOrder() {
		g, ok := guideFor(s.ID)
		if !ok {
			continue
		}
		for ci, c := range g.Concepts {
			if !s.hasStudied(c.Name) {
				continue
			}
			for ei := range c.Exercises {
				if s.hasDone(c.Name, ei) {
					continue
				}
				p := Practice{Stage: s.ID, Concept: ci, Exercise: ei}
				if s.ID == avoidStage {
					same = append(same, p)
				} else {
					other = append(other, p)
				}
				break // the easiest open exercise of each concept
			}
		}
	}
	if len(other) > 0 {
		return other[pick(len(other))], true
	}
	if len(same) > 0 {
		return same[pick(len(same))], true
	}
	return Practice{}, false
}

// redidAlone reports whether some exercise was solved with help and later
// solved again by the learner's own reasoning.
func (r *Registry) redidAlone() bool {
	helped := map[string]time.Time{}
	for _, s := range r.Solves {
		if s.helped() {
			if t, ok := helped[s.Key]; !ok || s.At.Before(t) {
				helped[s.Key] = s.At
			}
		}
	}
	for _, s := range r.Solves {
		if t, ok := helped[s.Key]; ok && !s.helped() && s.At.After(t) {
			return true
		}
	}
	return false
}
