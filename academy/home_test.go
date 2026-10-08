package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func renderHome(t *testing.T, r *Registry, cols int) string {
	t.Helper()
	var out bytes.Buffer
	a := &App{reg: r, con: &console{out: &out}, width: cols}
	a.printMenu()
	return stripANSI(out.String())
}

// The simple home shows three numbers, where you are and a short menu;
// the guide stays in sight for new learners; Settings brings back the
// detailed dashboard and every menu item.
func TestHomeScreens(t *testing.T) {
	newbie := renderHome(t, seedRegistry(), 80)
	for _, want := range []string{"No streak yet", "Today 0 min", "Nothing due", "[j] Today's workout", "[0] Start Here", "[?] All screens & keys", "[5] Exit"} {
		if !strings.Contains(newbie, want) {
			t.Errorf("simple home for a new learner lacks %q:\n%s", want, newbie)
		}
	}
	for _, hidden := range []string{"[w]", "[2]", "[8]", "Labs", "MAIN MENU"} {
		if strings.Contains(newbie, hidden) {
			t.Errorf("simple home shows %q, which belongs to All screens", hidden)
		}
	}

	learner := demoRegistry(time.Now())
	simple := renderHome(t, learner, 80)
	if strings.Contains(simple, "[0] Start Here") {
		t.Error("Start Here leaves the short menu once a few concepts are understood")
	}
	if !strings.Contains(simple, "cards due") || !strings.Contains(simple, "Now Stage") {
		t.Errorf("simple home lacks the due count or the current stage:\n%s", simple)
	}

	learner.Settings.DetailedHome = true
	detailed := renderHome(t, learner, 80)
	for _, want := range []string{"MAIN MENU", "[w] Tech watch", "[8] Undo this session", "Campus"} {
		if !strings.Contains(detailed, want) {
			t.Errorf("detailed home lacks %q", want)
		}
	}

	for _, cols := range []int{minWidth, 60, maxWidth} {
		for _, l := range strings.Split(renderHome(t, seedRegistry(), cols), "\n") {
			if visibleLen(l) > cols {
				t.Errorf("cols %d: %q overflows", cols, l)
			}
		}
	}
}
