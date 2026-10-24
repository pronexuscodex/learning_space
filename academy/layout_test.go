package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisplayWidth(t *testing.T) {
	cases := map[string]int{"abc": 3, "héllo": 5, "📕 Book": 7, "日本": 4, "\x1b[1mbold\x1b[0m": 4, "é": 1}
	for s, want := range cases {
		if got := visibleLen(s); got != want {
			t.Errorf("visibleLen(%q) = %d, want %d", s, got, want)
		}
	}
	if got := truncate("📕📕📕📕", 5); visibleLen(got) > 5 {
		t.Errorf("truncate overflowed: %q", got)
	}
}

func TestWrapNeverExceedsWidth(t *testing.T) {
	text := "A path like /tmp/claude-0/-home-user-learning-space/very/long/unbroken/path/name.json must still fit, as must 📕 wide glyphs."
	for _, w := range []int{12, 20, 37, 60} {
		for _, l := range wrap(text, w, "  ") {
			if visibleLen(l) > w {
				t.Errorf("width %d: line %q is %d wide", w, l, visibleLen(l))
			}
		}
	}
}

func TestFlowPacksItemsWithinWidth(t *testing.T) {
	items := []string{"Hours 0.00", "Labs 0", "Stages 0/16", "Texts 0/49", "Concepts 0/82"}
	for _, w := range []int{20, 40, 80} {
		lines := flow(items, " │ ", w, "  ")
		joined := strings.Join(lines, "\n")
		for _, it := range items {
			if !strings.Contains(joined, it) {
				t.Errorf("width %d lost item %q", w, it)
			}
		}
		for _, l := range lines {
			if visibleLen(l) > w {
				t.Errorf("width %d: %q is %d wide", w, l, visibleLen(l))
			}
		}
	}
	if n := len(flow(items, " │ ", 200, "")); n != 1 {
		t.Errorf("wide terminal should use one line, got %d", n)
	}
}

func TestPromptLabelWrapsLongQuestions(t *testing.T) {
	saved := layoutWidth
	defer func() { layoutWidth = saved }()
	layoutWidth = func() int { return 40 }
	p := promptLabel("Do you understand it well enough to explain it to a friend?", "y/n")
	for _, l := range strings.Split(p, "\n") {
		if visibleLen(l) > 40 {
			t.Errorf("prompt line %q is %d wide", l, visibleLen(l))
		}
	}
	if !strings.HasSuffix(stripANSI(p), "› ") {
		t.Error("the input must stay on the prompt's last line")
	}
}

func TestWrapKeepsCodeVerbatim(t *testing.T) {
	text := "Declare it first:\n\n  int count = 3;   // a comment\n  double x;\n\nThen use it\nin a sentence."
	got := wrap(text, 40, "  ")
	want := []string{"  Declare it first:", "", "    int count = 3;   // a comment", "    double x;", "", "  Then use it in a sentence."}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("wrap =\n%q\nwant\n%q", got, want)
	}
	for _, l := range wrap("  "+strings.Repeat("x", 100), 40, "") {
		if visibleLen(l) > 40 {
			t.Fatalf("long code line not broken: %q", l)
		}
	}
}

// Every concept card, in every stage, fits the terminal at the widths the
// academy supports: nothing is cut off at the right edge.
func TestConceptCardsFitTheScreen(t *testing.T) {
	reg := seedRegistry()
	for _, w := range []int{40, 60, 80} {
		for id, g := range curriculum {
			for i, c := range g.Concepts {
				var out bytes.Buffer
				a := &App{reg: reg, width: w, con: &console{out: &out}}
				builds := reg.prereqsOf(c.Name)
				a.renderConcept(c, i+1, len(g.Concepts), g.Glossary, make([]bool, len(c.Exercises)), "", 0, builds, []string{"○ a checked exercise (not run yet)"})
				for _, l := range strings.Split(out.String(), "\n") {
					if n := visibleLen(l); n > w {
						t.Errorf("%d columns, Stage %d %q: a line is %d wide: %q", w, id, c.Name, n, stripANSI(l))
					}
				}
			}
		}
	}
}

// A bullet's continuation lines, indented by two spaces in the source,
// reflow with the bullet; an unindented line after a bullet starts a new
// paragraph instead of joining it.
func TestWrapBulletContinuationsAndParagraphs(t *testing.T) {
	text := "- first item that goes on\n  and continues here;\n- second item.\nA new paragraph."
	got := wrap(text, 80, "")
	want := []string{"• first item that goes on and continues here;", "• second item.", "A new paragraph."}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("wrap =\n%q\nwant\n%q", got, want)
	}
}

// No prose word in a lesson is ever split across two lines at a narrow
// width: only code and words longer than the line may be broken.
func TestLessonProseNeverSplitsWords(t *testing.T) {
	const w = 40
	for id, g := range curriculum {
		for _, c := range g.Concepts {
			for _, text := range []string{c.Body, c.UnderTheHood, c.Analogy, c.Example} {
				// The prose words, straight from the source: every line
				// except code (indented, and not continuing a bullet).
				var want []string
				inBullet := false
				for _, raw := range strings.Split(text, "\n") {
					switch {
					case strings.HasPrefix(raw, "- "):
						inBullet = true
					case strings.HasPrefix(raw, "  ") && !(inBullet && len(raw) > 2 && raw[2] != ' '):
						inBullet = false
						continue
					case !strings.HasPrefix(raw, "  "):
						inBullet = false
					}
					want = append(want, strings.Fields(strings.TrimPrefix(raw, "- "))...)
				}
				got := map[string]bool{}
				for _, l := range wrap(text, w-4, "  ") {
					for _, f := range strings.Fields(strings.TrimPrefix(strings.TrimSpace(l), "• ")) {
						got[f] = true
					}
				}
				for _, word := range want {
					if visibleLen(word) < w-8 && !got[word] {
						t.Errorf("Stage %d %q: %q is split across lines at %d columns", id, c.Name, word, w)
						break
					}
				}
			}
		}
	}
}

// A stage's glossary stays readable at every width: no meaning is squeezed
// into a column so narrow that words split or lines start with a comma.
func TestGlossaryNeverSplitsWords(t *testing.T) {
	for _, w := range []int{40, 80} {
		for id, g := range curriculum {
			var out bytes.Buffer
			a := &App{reg: seedRegistry(), width: w, con: &console{out: &out}}
			a.showGlossary(g)
			text := stripANSI(out.String())
			got := map[string]bool{}
			for _, l := range strings.Split(text, "\n") {
				if s := strings.TrimSpace(l); s != "" && (strings.ContainsAny(s[:1], ",;:)") || s == "." || strings.HasPrefix(s, ". ")) {
					t.Errorf("%d columns, Stage %d: a line starts with punctuation: %q", w, id, l)
				}
				for _, f := range strings.Fields(l) {
					got[f] = true
				}
			}
			for _, term := range g.Glossary {
				for _, word := range strings.Fields(term.Meaning) {
					if !got[word] {
						t.Errorf("%d columns, Stage %d %q: %q is split across lines", w, id, term.Word, word)
						break
					}
				}
			}
		}
	}
}

// Every diagram fits a concept card on an 80-column terminal whole: 80
// columns leave 79 for the app and 73 beside the card's edge, and a
// clipped diagram hides the part that explains it.
func TestDiagramsFitEightyColumns(t *testing.T) {
	for id, g := range curriculum {
		for _, c := range g.Concepts {
			for _, l := range dedent(c.Diagram) {
				if n := visibleLen(l); n > 73 {
					t.Errorf("Stage %d %q: a diagram line is %d wide (73 at most): %q", id, c.Name, n, l)
				}
			}
		}
	}
}
