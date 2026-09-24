package main

// Retention engine: spaced-repetition review cards, their scheduler, and
// the mastery ladder. It is built on well-established findings from
// learning science:
//   - retrieval practice: recalling an answer strengthens memory far more
//     than re-reading it;
//   - spacing: reviews spread over growing intervals beat cramming;
//   - interleaving: mixing topics in a session improves transfer.

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// ReviewCard is the persisted scheduling state of one review card.
type ReviewCard struct {
	Due          time.Time `json:"due"`
	IntervalDays float64   `json:"interval_days"`
	Ease         float64   `json:"ease"`
	Reps         int       `json:"reps"`   // consecutive successful reviews
	Lapses       int       `json:"lapses"` // times forgotten
	LastReviewed time.Time `json:"last_reviewed"`
}

// Review grades, as offered to the learner.
const (
	GradeAgain = 1 // forgot
	GradeHard  = 2 // recalled with serious effort
	GradeGood  = 3 // recalled correctly
	GradeEasy  = 4 // recalled instantly
)

const (
	startEase   = 2.5
	minEase     = 1.3
	maxInterval = 365.0

	// Mastery thresholds, in days between successful reviews.
	retainedDays = 7.0
	masteredDays = 21.0
)

// schedule returns the card's new state after a review with grade g.
// It is a compact variant of the SM-2 algorithm used by most
// spaced-repetition software.
func schedule(c ReviewCard, g int, now time.Time) ReviewCard {
	if c.Ease == 0 {
		c.Ease = startEase
	}
	c.LastReviewed = now
	switch g {
	case GradeAgain:
		c.Lapses++
		c.Reps = 0
		c.Ease = math.Max(minEase, c.Ease-0.2)
		c.IntervalDays = 0
		c.Due = now.Add(10 * time.Minute)
		return c
	case GradeHard:
		if c.Reps == 0 {
			c.IntervalDays = 1
		} else {
			c.IntervalDays = math.Max(1, c.IntervalDays*1.2)
		}
		c.Ease = math.Max(minEase, c.Ease-0.15)
	case GradeEasy:
		if c.Reps == 0 {
			c.IntervalDays = 4
		} else {
			c.IntervalDays = math.Max(c.IntervalDays+1, c.IntervalDays*c.Ease*1.3)
		}
		c.Ease += 0.15
	default: // GradeGood
		switch c.Reps {
		case 0:
			c.IntervalDays = 1
		case 1:
			c.IntervalDays = 3
		default:
			c.IntervalDays = math.Max(c.IntervalDays+1, c.IntervalDays*c.Ease)
		}
	}
	c.Reps++
	c.IntervalDays = math.Min(maxInterval, math.Round(c.IntervalDays*10)/10)
	c.Due = dayStart(now).AddDate(0, 0, int(math.Round(c.IntervalDays)))
	return c
}

// previewInterval describes when a card would next be due for each grade.
func previewInterval(c ReviewCard, g int, now time.Time) string {
	n := schedule(c, g, now)
	if g == GradeAgain {
		return "again soon"
	}
	d := int(math.Round(n.IntervalDays))
	switch {
	case d <= 1:
		return "1 day"
	case d < 30:
		return fmt.Sprintf("%d days", d)
	case d < 365:
		return fmt.Sprintf("%.1f months", float64(d)/30)
	}
	return "1 year"
}

// Card kinds.
const (
	CardKeyIdea  = "key idea"
	CardRealLife = "real life"
	CardWarmUp   = "warm-up"
	CardQuiz     = "quiz"
)

// Card is one review prompt derived from the curriculum.
type Card struct {
	ID      string
	StageID int
	Concept string // empty for stage quiz cards
	Kind    string
	Q, A    string
}

// conceptCards returns the review cards for one concept. The answers come
// straight from the curriculum, so they are always consistent with it.
func conceptCards(stageID int, c Concept) []Card {
	cards := []Card{
		{ID: c.Name + " :: key idea", StageID: stageID, Concept: c.Name, Kind: CardKeyIdea,
			Q: fmt.Sprintf("Explain the key idea of “%s” in one or two sentences.", c.Name),
			A: c.Summary + "\n\nMental model: " + c.MentalModel},
		{ID: c.Name + " :: real life", StageID: stageID, Concept: c.Name, Kind: CardRealLife,
			Q: fmt.Sprintf("Give a real-world example of “%s”, and say how the idea shows up in it.", c.Name),
			A: c.Example},
	}
	for _, e := range c.Exercises {
		if e.Level == LevelWarmUp {
			cards = append(cards, Card{ID: c.Name + " :: warm-up", StageID: stageID, Concept: c.Name, Kind: CardWarmUp,
				Q: e.Task, A: e.Hint})
			break
		}
	}
	return cards
}

// quizCards returns a stage's quiz questions as review cards.
func quizCards(stageID int, g StageGuide) []Card {
	var cards []Card
	for i, q := range g.Quiz {
		cards = append(cards, Card{ID: fmt.Sprintf("stage %d :: quiz #%d", stageID, i+1), StageID: stageID, Kind: CardQuiz, Q: q.Q, A: q.A})
	}
	return cards
}

// stageCards returns every card a stage can produce, unlocked or not.
func stageCards(stageID int) []Card {
	g, ok := guideFor(stageID)
	if !ok {
		return nil
	}
	var cards []Card
	for _, c := range g.Concepts {
		cards = append(cards, conceptCards(stageID, c)...)
	}
	return append(cards, quizCards(stageID, g)...)
}

// unlockedCards returns the cards that belong in the learner's review deck:
// a concept's cards once it is marked understood, and a stage's quiz cards
// once at least half of its concepts are understood.
func (r *Registry) unlockedCards() []Card {
	var cards []Card
	for ti := range r.Tracks {
		for si := range r.Tracks[ti].Stages {
			s := &r.Tracks[ti].Stages[si]
			g, ok := guideFor(s.ID)
			if !ok {
				continue
			}
			for _, c := range g.Concepts {
				if s.hasStudied(c.Name) {
					cards = append(cards, conceptCards(s.ID, c)...)
				}
			}
			if studied, total := conceptProgress(s); total > 0 && studied*2 >= total {
				cards = append(cards, quizCards(s.ID, g)...)
			}
		}
	}
	return cards
}

// dueCards returns unlocked cards due at now; cards never reviewed are due.
// Cards are ordered by due time, so overdue cards come first.
func (r *Registry) dueCards(now time.Time) []Card {
	var due []Card
	for _, c := range r.unlockedCards() {
		st, ok := r.Reviews[c.ID]
		if !ok || !st.Due.After(now) {
			due = append(due, c)
		}
	}
	sort.SliceStable(due, func(i, j int) bool {
		return r.Reviews[due[i].ID].Due.Before(r.Reviews[due[j].ID].Due)
	})
	return due
}

// Mastery levels, from nothing to long-term mastery.
const (
	MasteryNew = iota
	MasteryUnderstood
	MasteryPractised
	MasteryRetained
	MasteryMastered
)

var masteryNames = []string{"new", "understood", "practised", "retained", "mastered"}
var masteryGlyphs = []string{"○", "◔", "◑", "◕", "●"}

// conceptMastery places a concept on the mastery ladder:
//   - understood: marked as understood;
//   - practised:  understood, with at least two exercises done;
//   - retained:   practised, and every review card remembered at ≥ 7-day spacing;
//   - mastered:   all exercises done, and every card remembered at ≥ 21-day spacing.
func (r *Registry) conceptMastery(s *Stage, c Concept) int {
	if !s.hasStudied(c.Name) {
		return MasteryNew
	}
	done := 0
	for i := range c.Exercises {
		if s.hasDone(c.Name, i) {
			done++
		}
	}
	if done < 2 {
		return MasteryUnderstood
	}
	minInterval := math.Inf(1)
	for _, card := range conceptCards(s.ID, c) {
		st, ok := r.Reviews[card.ID]
		if !ok {
			return MasteryPractised
		}
		minInterval = math.Min(minInterval, st.IntervalDays)
	}
	switch {
	case done == len(c.Exercises) && minInterval >= masteredDays:
		return MasteryMastered
	case minInterval >= retainedDays:
		return MasteryRetained
	}
	return MasteryPractised
}

// stageMastery returns (sum of mastery levels, max possible, mastered count, concept count).
func (r *Registry) stageMastery(s *Stage) (points, maxPoints, mastered, total int) {
	g, ok := guideFor(s.ID)
	if !ok {
		return
	}
	for _, c := range g.Concepts {
		m := r.conceptMastery(s, c)
		points += m
		if m == MasteryMastered {
			mastered++
		}
	}
	return points, MasteryMastered * len(g.Concepts), mastered, len(g.Concepts)
}

// MasteryCheck records the latest stage exam.
type MasteryCheck struct {
	TakenAt time.Time `json:"taken_at"`
	Score   int       `json:"score"`
	Total   int       `json:"total"`
	Passed  bool      `json:"passed"`
}

// masteryPassMark is the fraction of the mastery check needed to pass.
const masteryPassMark = 0.8

// masteryCheckSize caps how many questions one mastery check asks.
const masteryCheckSize = 12
