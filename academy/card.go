package main

// The share card: a compact, screenshot-friendly summary of a learner's
// progress, with the project's address, for posting on social media
// ("#LearnInPublic"). Press k on the menu, or run academy -card.

import (
	"fmt"
	"strings"
	"time"
)

const projectURL = "github.com/pronexuscodex/learning_space"

// shortTrackNames label the card's track bars.
var shortTrackNames = map[string]string{"F": "Foundations", "A": "Systems", "S": "Software", "B": "AI", "D": "Security"}

// shareCard renders the card for a terminal cols wide.
func (r *Registry) shareCard(now time.Time, cols int) []string {
	cs := r.stats(now, 400)
	inner := max(24, min(cols-4, 44)) // text width inside "│ " and " │"

	var body []string
	add := func(s string) { body = append(body, s) }
	row := func(label, value string) {
		add(padRight(sty.Gray(label), 17) + value)
	}

	add(sty.Bold(sty.Cyan("SYSTEMS & AI ACADEMY")))
	add(sty.Gray("A CS degree in your terminal"))
	add("")
	row("Concepts", fmt.Sprintf("%s / %d", sty.Bold(fmt.Sprint(cs.conceptsStudied)), cs.concepts))
	row("Exercises", fmt.Sprintf("%s / %d", sty.Bold(fmt.Sprint(cs.exercisesDone)), cs.exercises))
	row("Stages done", fmt.Sprintf("%s / %d", sty.Bold(fmt.Sprint(cs.stagesGrad)), cs.stages))
	row("Study time", sty.Bold(fmt.Sprintf("%.1f h", cs.hours)))
	streak := "—"
	if cs.streak > 0 {
		streak = sty.Bold(sty.Yellow(fmt.Sprintf("%d day%s", cs.streak, plural(cs.streak))))
	}
	row("Streak", streak)
	row("Achievements", fmt.Sprintf("%s / %d", sty.Bold(fmt.Sprint(len(r.Achievements))), len(achievements)))
	add("")

	// One bar per track: concepts studied.
	barW := max(6, inner-20)
	for _, t := range r.Tracks {
		studied, total := 0, 0
		for i := range t.Stages {
			s, c := conceptProgress(&t.Stages[i])
			studied += s
			total += c
		}
		name, ok := shortTrackNames[t.ID]
		if !ok {
			name = truncate(t.Name, 12)
		}
		add(padRight(trackColor(t.ID)(name), 13) + bar(studied, total, barW, trackColor(t.ID)) + " " + padRight(pct(studied, total), 4))
	}
	if s := r.currentStage(); s != nil {
		add("")
		add(sty.Gray("Now: ") + truncate(fmt.Sprintf("Stage %d · %s", s.ID, s.Title), inner-5))
	}

	out := []string{"╭" + strings.Repeat("─", inner+2) + "╮"}
	for _, l := range body {
		out = append(out, "│ "+padRight(l, inner)+" │")
	}
	out = append(out, "╰"+strings.Repeat("─", inner+2)+"╯")
	if footer := "Free & open source: "; visibleLen(footer+projectURL) < cols {
		out = append(out, sty.Gray(footer)+projectURL)
	} else {
		out = append(out, projectURL)
	}
	return out
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// showShareCard prints the card with a hint on how to share it.
func (a *App) showShareCard() {
	a.mu.Lock()
	lines := a.reg.shareCard(time.Now(), a.cols())
	a.mu.Unlock()
	a.println("")
	for _, l := range lines {
		a.println(l)
	}
	a.println("")
	for _, l := range wrap("Take a screenshot and share your progress with #LearnInPublic. From a terminal, academy -card prints this card.", a.cols()-2, "  ") {
		a.println(sty.Gray(l))
	}
}
