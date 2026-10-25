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

// The home screen shows everything by default: the full dashboard and
// every screen. Settings offers the simple home, which shows three
// numbers, where you are and a short menu, with the guide in sight for
// new learners.
func TestHomeScreens(t *testing.T) {
	for _, r := range []*Registry{seedRegistry(), demoRegistry(time.Now())} {
		home := renderHome(t, r, 80)
		for _, want := range []string{"MAIN MENU", "[0] Start Here", "[w] Tech watch", "[8] Undo this session", "[q] Exit", "Campus"} {
			if !strings.Contains(home, want) {
				t.Errorf("the default home lacks %q:\n%s", want, home)
			}
		}
		if strings.Contains(home, "Tidy") || strings.Contains(home, "[t]") {
			t.Errorf("Tidy screen is gone, but the home still offers it:\n%s", home)
		}
	}

	newbie := seedRegistry()
	newbie.Settings.SimpleHome = true
	simpleNew := renderHome(t, newbie, 80)
	for _, want := range []string{"No streak yet", "Today 0 min", "Nothing due", "[j] Today's workout", "[0] Start Here", "[?] All screens & keys", "[q] Exit"} {
		if !strings.Contains(simpleNew, want) {
			t.Errorf("simple home for a new learner lacks %q:\n%s", want, simpleNew)
		}
	}
	for _, hidden := range []string{"[w]", "[2]", "[8]", "Labs", "MAIN MENU"} {
		if strings.Contains(simpleNew, hidden) {
			t.Errorf("simple home shows %q, which belongs to All screens", hidden)
		}
	}

	learner := demoRegistry(time.Now())
	learner.Settings.SimpleHome = true
	simple := renderHome(t, learner, 80)
	if strings.Contains(simple, "[0] Start Here") {
		t.Error("Start Here leaves the short menu once a few concepts are understood")
	}
	if !strings.Contains(simple, "cards due") || !strings.Contains(simple, "Now Stage") {
		t.Errorf("simple home lacks the due count or the current stage:\n%s", simple)
	}

	for _, cols := range []int{minWidth, 60, maxWidth} {
		for _, l := range strings.Split(renderHome(t, seedRegistry(), cols), "\n") {
			if visibleLen(l) > cols {
				t.Errorf("cols %d: %q overflows", cols, l)
			}
		}
	}
}
