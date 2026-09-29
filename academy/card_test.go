package main

import (
	"strings"
	"testing"
	"time"
)

// The card must fit every terminal width, label every track and carry
// the project's address, because it is made to be shared.
func TestShareCardFits(t *testing.T) {
	old := sty
	sty = Style{on: true}
	defer func() { sty = old }()
	now := time.Now()
	reg := demoRegistry(now)
	for _, cols := range []int{minWidth, 40, 60, 80, maxWidth} {
		lines := reg.shareCard(now, cols)
		for _, l := range lines {
			if w := visibleLen(l); w > cols {
				t.Errorf("cols %d: line is %d wide: %q", cols, w, stripANSI(l))
			}
		}
		text := stripANSI(strings.Join(lines, "\n"))
		for _, want := range []string{"SYSTEMS & AI ACADEMY", "Streak", projectURL} {
			if !strings.Contains(text, want) {
				t.Errorf("cols %d: card lacks %q", cols, want)
			}
		}
		for _, tr := range reg.Tracks {
			if !strings.Contains(text, shortTrackNames[tr.ID]+" ") {
				t.Errorf("cols %d: track %s has no short name on the card", cols, tr.ID)
			}
		}
	}
}
