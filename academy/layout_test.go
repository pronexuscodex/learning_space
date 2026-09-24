package main

import (
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
