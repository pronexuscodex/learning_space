package main

// Guidance and insight: what to do next, search across the curriculum,
// study sessions, activity history and a Markdown export. Everything here
// is pure logic over the registry; the screens live in insightsui.go.

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// StudySession is one focus-timer session.
type StudySession struct {
	Start   time.Time `json:"start"`
	Minutes float64   `json:"minutes"`
	Stage   int       `json:"stage"` // 0: general study
	Note    string    `json:"note,omitempty"`
}

// ---------------------------------------------------------------------------
// What's next
// ---------------------------------------------------------------------------

// Suggestion kinds.
const (
	SuggestReview   = "review"
	SuggestConcept  = "concept"
	SuggestExercise = "exercise"
	SuggestMastery  = "mastery"
	SuggestTypeIn   = "type-in"
	SuggestFocus    = "focus"
)

// Suggestion is one recommended next step.
type Suggestion struct {
	Kind     string
	Title    string
	Why      string
	Stage    int
	Concept  int // index within the stage's guide
	Exercise int // index within the concept's exercises
}

// stageOrder is the recommended order of study: Foundations first, then
// Systems and Software, then AI.
func (r *Registry) stageOrder() []*Stage {
	var out []*Stage
	for ti := range r.Tracks {
		for si := range r.Tracks[ti].Stages {
			out = append(out, &r.Tracks[ti].Stages[si])
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// prereqsReady reports whether at least half of every prerequisite
// stage's concepts are understood.
func (r *Registry) prereqsReady(stageID int) bool {
	for _, p := range stagePrereqs[stageID] {
		_, ps := r.findStage(p)
		if ps == nil {
			continue
		}
		if studied, total := conceptProgress(ps); studied*2 < total {
			return false
		}
	}
	return true
}

// currentStage is where the learner is working: the first stage (in study
// order) that has been started but not fully understood; otherwise the
// first unstarted stage whose prerequisites are ready.
func (r *Registry) currentStage() *Stage {
	var fresh *Stage
	for _, s := range r.stageOrder() {
		studied, total := conceptProgress(s)
		if total == 0 || studied == total {
			continue
		}
		if studied > 0 {
			return s
		}
		if fresh == nil && r.prereqsReady(s.ID) {
			fresh = s
		}
	}
	return fresh
}

// nextSteps returns up to four suggestions, most important first:
// reviews that are due (they protect everything already learned), the
// next concept, an open exercise, a mastery check or type-in that is
// ready, and a nudge to keep the streak alive.
func (r *Registry) nextSteps(now time.Time) []Suggestion {
	var out []Suggestion
	if due := len(r.dueCards(now)); due > 0 {
		out = append(out, Suggestion{Kind: SuggestReview,
			Title: fmt.Sprintf("Review %d due card(s)", due),
			Why:   "Reviews protect what you have already learned, so they come first. Most sessions take 5–15 minutes."})
	}

	if s := r.currentStage(); s != nil {
		g, _ := guideFor(s.ID)
		for i, c := range g.Concepts {
			if !s.hasStudied(c.Name) {
				why := "The next concept in your current stage."
				if studied, _ := conceptProgress(s); studied == 0 {
					why = "A good place to start: the stages this one builds on are covered."
				}
				out = append(out, Suggestion{Kind: SuggestConcept, Stage: s.ID, Concept: i,
					Title: fmt.Sprintf("Study “%s” (Stage %d)", c.Name, s.ID), Why: why})
				break
			}
		}
	}

	// The lowest open exercise of the earliest understood concept: warm-ups
	// before practice before real-world.
exercises:
	for _, s := range r.stageOrder() {
		g, ok := guideFor(s.ID)
		if !ok {
			continue
		}
		for level := 0; level < 3; level++ {
			for ci, c := range g.Concepts {
				if !s.hasStudied(c.Name) || level >= len(c.Exercises) || s.hasDone(c.Name, level) {
					continue
				}
				out = append(out, Suggestion{Kind: SuggestExercise, Stage: s.ID, Concept: ci, Exercise: level,
					Title: fmt.Sprintf("%s exercise: %s (Stage %d)", levelBadgePlain(c.Exercises[level].Level), c.Name, s.ID),
					Why:   "You understand the idea; practice is what makes it stick."})
				break exercises
			}
		}
	}

	for _, s := range r.stageOrder() {
		studied, total := conceptProgress(s)
		if total > 0 && studied == total && (s.MasteryCheck == nil || !s.MasteryCheck.Passed) {
			out = append(out, Suggestion{Kind: SuggestMastery, Stage: s.ID,
				Title: fmt.Sprintf("Take the mastery check for Stage %d", s.ID),
				Why:   "Every concept is understood; a mixed exam shows what has really stuck."})
			break
		}
	}

	if r.ClassicMode {
		for _, s := range r.stageOrder() {
			if studied, _ := conceptProgress(s); studied > 0 && !s.TypeInDone {
				out = append(out, Suggestion{Kind: SuggestTypeIn, Stage: s.ID,
					Title: fmt.Sprintf("Type in this stage's listing (Stage %d)", s.ID),
					Why:   "Predict, type it by hand, run it, compare."})
				break
			}
		}
	}

	if cs := r.stats(now, 60); cs.activity[len(cs.activity)-1] == 0 && r.ReviewHistory[now.Format("2006-01-02")] == 0 {
		why := "Even one focused session keeps momentum."
		if cs.streak > 0 {
			why = fmt.Sprintf("Keep your %d-day streak alive: any session or review today counts.", cs.streak)
		}
		out = append(out, Suggestion{Kind: SuggestFocus, Title: "Start a 25-minute focus session", Why: why})
	}

	if len(out) > 4 {
		out = out[:4]
	}
	return out
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// Search hit kinds.
const (
	HitConcept   = "concept"
	HitGlossary  = "glossary"
	HitResource  = "resource"
	HitClassic   = "classic"
	HitBlueprint = "blueprint"
)

// Hit is one search result.
type Hit struct {
	Kind    string
	Stage   int
	Title   string
	Snippet string
	Concept int // for concept hits
	score   int
}

// snippet returns a short excerpt of text around the first query word.
func snippet(text, word string, width int) string {
	flat := strings.Join(strings.Fields(text), " ")
	i := strings.Index(strings.ToLower(flat), word)
	if i < 0 {
		return truncate(flat, width)
	}
	start := max(0, i-width/3)
	for start > 0 && flat[start-1] != ' ' { // don't start mid-word
		start--
	}
	out := flat[start:]
	if start > 0 {
		out = "…" + out
	}
	return truncate(out, width)
}

// search finds curriculum content containing every word of the query
// (case-insensitive). Name and title matches rank above body matches.
func search(query string) []Hit {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil
	}
	matches := func(texts ...string) bool {
		all := strings.ToLower(strings.Join(texts, " "))
		for _, w := range words {
			if !strings.Contains(all, w) {
				return false
			}
		}
		return true
	}
	titleScore := func(title string) int {
		t := strings.ToLower(title)
		n := 0
		for _, w := range words {
			if strings.Contains(t, w) {
				n += 10
			}
		}
		return n
	}

	var hits []Hit
	ids := make([]int, 0, len(curriculum))
	for id := range curriculum {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		g := curriculum[id]
		for ci, c := range g.Concepts {
			body := strings.Join([]string{c.Summary, c.Analogy, c.Example, c.Body, c.MentalModel}, " ")
			if matches(c.Name, body) {
				hits = append(hits, Hit{Kind: HitConcept, Stage: id, Concept: ci, Title: c.Name,
					Snippet: snippet(body, words[0], 100), score: 5 + titleScore(c.Name)})
			}
		}
		for _, t := range g.Glossary {
			if matches(t.Word, t.Meaning) {
				hits = append(hits, Hit{Kind: HitGlossary, Stage: id, Title: t.Word, Snippet: t.Meaning, score: 3 + titleScore(t.Word)})
			}
		}
		for _, r := range g.Resources {
			if matches(r.Title, r.Note, r.Kind) {
				hits = append(hits, Hit{Kind: HitResource, Stage: id, Title: r.Title, Snippet: snippet(r.Note, words[0], 100), score: 2 + titleScore(r.Title)})
			}
		}
		for _, b := range g.Blueprints {
			if matches(b.Name, b.Brief) {
				hits = append(hits, Hit{Kind: HitBlueprint, Stage: id, Title: b.Name, Snippet: b.Brief, score: 2 + titleScore(b.Name)})
			}
		}
		if cg, _, ok := classicFor(id); ok {
			for _, cr := range []ClassicRead{cg.Anchor, cg.Classic, cg.Source} {
				if matches(cr.Title, cr.Author, cr.Why) {
					hits = append(hits, Hit{Kind: HitClassic, Stage: id, Title: cr.Title, Snippet: cr.Author, score: 2 + titleScore(cr.Title)})
				}
			}
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	return hits
}

// ---------------------------------------------------------------------------
// Activity history
// ---------------------------------------------------------------------------

// dailyMinutes returns study minutes per local day (lab hours, focus
// sessions, and one minute per review card) for the last n days, oldest
// first, ending today.
func (r *Registry) dailyMinutes(now time.Time, n int) []float64 {
	out := make([]float64, n)
	today := dayStart(now)
	add := func(at time.Time, minutes float64) {
		ago := int(math.Round(today.Sub(dayStart(at)).Hours() / 24))
		if ago >= 0 && ago < n {
			out[n-1-ago] += minutes
		}
	}
	for ti := range r.Tracks {
		for _, s := range r.Tracks[ti].Stages {
			for _, l := range s.Labs {
				for _, e := range l.Log {
					add(e.At, e.Hours*60)
				}
			}
		}
	}
	for _, ss := range r.StudySessions {
		add(ss.Start, ss.Minutes)
	}
	for day, cards := range r.ReviewHistory {
		if t, err := time.ParseInLocation("2006-01-02", day, time.Local); err == nil {
			add(t, float64(cards))
		}
	}
	return out
}

// weekSummary compares the last 7 days with the 7 before them.
type weekSummary struct {
	minutes, prevMinutes float64
	activeDays, prevDays int
	reviews, prevReviews int
	sessions, prevSess   int
}

func (r *Registry) weekSummary(now time.Time) weekSummary {
	var w weekSummary
	days := r.dailyMinutes(now, 14)
	for i, m := range days {
		if i >= 7 {
			w.minutes += m
			if m > 0 {
				w.activeDays++
			}
		} else {
			w.prevMinutes += m
			if m > 0 {
				w.prevDays++
			}
		}
	}
	today := dayStart(now)
	for day, cards := range r.ReviewHistory {
		t, err := time.ParseInLocation("2006-01-02", day, time.Local)
		if err != nil {
			continue
		}
		switch ago := int(math.Round(today.Sub(t).Hours() / 24)); {
		case ago >= 0 && ago < 7:
			w.reviews += cards
		case ago >= 7 && ago < 14:
			w.prevReviews += cards
		}
	}
	for _, ss := range r.StudySessions {
		switch ago := int(math.Round(today.Sub(dayStart(ss.Start)).Hours() / 24)); {
		case ago >= 0 && ago < 7:
			w.sessions++
		case ago >= 7 && ago < 14:
			w.prevSess++
		}
	}
	return w
}

// hardestCards returns the review cards forgotten most often.
func (r *Registry) hardestCards(n int) []Card {
	var cards []Card
	for _, c := range r.unlockedCards() {
		if r.Reviews[c.ID].Lapses > 0 {
			cards = append(cards, c)
		}
	}
	sort.SliceStable(cards, func(i, j int) bool { return r.Reviews[cards[i].ID].Lapses > r.Reviews[cards[j].ID].Lapses })
	if len(cards) > n {
		cards = cards[:n]
	}
	return cards
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

// exportMarkdown renders the learner's progress, own-words explanations,
// lab notebook and labs as a Markdown document.
func (r *Registry) exportMarkdown(now time.Time) string {
	var b strings.Builder
	cs := r.stats(now, 60)
	fmt.Fprintf(&b, "# My Systems & AI Academy notes\n\n_Exported %s._\n\n", now.Format("2006-01-02 15:04"))
	fmt.Fprintf(&b, "- Study time: **%.1f h** · streak: **%d day(s)**\n", cs.hours, cs.streak)
	fmt.Fprintf(&b, "- Concepts understood: **%d/%d** · mastered: **%d** · exercises: **%d/%d**\n",
		cs.conceptsStudied, cs.concepts, cs.mastered, cs.exercisesDone, cs.exercises)
	fmt.Fprintf(&b, "- Review deck: %d card(s), %d due\n\n", len(r.unlockedCards()), cs.due)

	for _, s := range r.stageOrder() {
		g, _ := guideFor(s.ID)
		notes := r.notebookFor(s.ID, "")
		started := len(s.ConceptsStudied) > 0 || len(s.ExercisesDone) > 0 || len(s.Labs) > 0 || len(notes) > 0
		if !started {
			continue
		}
		_, _, mastered, total := r.stageMastery(s)
		fmt.Fprintf(&b, "## Stage %d · %s\n\n", s.ID, s.Title)
		fmt.Fprintf(&b, "Status: %s · mastered %d/%d", s.Status, mastered, total)
		if s.MasteryCheck != nil {
			fmt.Fprintf(&b, " · mastery check %d/%d", s.MasteryCheck.Score, s.MasteryCheck.Total)
		}
		b.WriteString("\n\n")
		for _, c := range g.Concepts {
			if !s.hasStudied(c.Name) {
				continue
			}
			level := r.conceptMastery(s, c)
			fmt.Fprintf(&b, "### %s — %s\n\n", c.Name, masteryNames[level])
			fmt.Fprintf(&b, "> %s\n\n", c.MentalModel)
			if note := s.Notes[c.Name]; note != "" {
				fmt.Fprintf(&b, "**In my own words:** %s\n\n", note)
			}
			var done []string
			for i, e := range c.Exercises {
				if s.hasDone(c.Name, i) {
					done = append(done, levelBadgePlain(e.Level))
				}
			}
			if len(done) > 0 {
				fmt.Fprintf(&b, "Exercises done: %s\n\n", strings.Join(done, ", "))
			}
		}
		if len(notes) > 0 {
			b.WriteString("### Lab notebook\n\n")
			for _, n := range notes {
				fmt.Fprintf(&b, "- %s · **%s** (%s): %s\n", n.At.Local().Format("2006-01-02"), n.Kind, n.Ref, n.Text)
			}
			b.WriteString("\n")
		}
		if len(s.Labs) > 0 {
			b.WriteString("### Labs\n\n")
			for _, l := range s.Labs {
				fmt.Fprintf(&b, "- **%s**: %.1f h, %s, %s", l.Name, l.HoursLogged, l.CompilationStatus, l.Status)
				if l.Architecture != "" {
					fmt.Fprintf(&b, ". %s", l.Architecture)
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
