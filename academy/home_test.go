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

// The next step is the last line of the full home, right above the
// prompt, so a 24-row terminal shows it even when the dashboard has
// scrolled away.
func TestNextStepIsBesideThePrompt(t *testing.T) {
	for _, cols := range []int{minWidth, 80, maxWidth} {
		for _, r := range []*Registry{seedRegistry(), demoRegistry(time.Now())} {
			lines := strings.Split(strings.TrimRight(renderHome(t, r, cols), "\n"), "\n")
			last := lines[len(lines)-1]
			if !strings.Contains(last, "Today's workout") && !strings.Contains(last, "Next") {
				t.Errorf("cols %d: the last line of the home is %q, not the next step", cols, last)
			}
			if !strings.Contains(lines[len(lines)-2], "└") {
				t.Errorf("cols %d: the next step should come straight after the menu, got %q above it", cols, lines[len(lines)-2])
			}
		}
	}
}

// The dashboard says what each number counts, in the words the other
// screens use: graduated stages and labs (not a bare ★, which the Study
// Hall uses for a passed check), "Reading" as in the Ledger, and hours
// for the 14-day line (the streak also counts reviews).
func TestDashboardWords(t *testing.T) {
	home := renderHome(t, demoRegistry(time.Now()), 110)
	for _, want := range []string{"Stages graduated", "graduated)", "Reading", "Hours, 14 days", "Mastered"} {
		if !strings.Contains(home, want) {
			t.Errorf("dashboard lacks %q:\n%s", want, home)
		}
	}
	for _, old := range []string{"★)", "Texts", "Last 14d"} {
		if strings.Contains(home, old) {
			t.Errorf("dashboard still says %q", old)
		}
	}
	if stageStateNames[StageMastered] == "mastered" {
		t.Error("a stage's star is a passed check; \"mastered\" is reserved for concept mastery")
	}
}
