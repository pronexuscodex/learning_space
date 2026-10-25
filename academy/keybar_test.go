package main

import (
	"strings"
	"testing"
)

// How to get around is never drawn in grey (SGR 90), which some terminal
// palettes make unreadable (Solarized Dark draws it in the background
// colour): the key bar's words and a prompt's hint use the terminal's own
// text colour, in every theme.
func TestNavigationTextIsNotGrey(t *testing.T) {
	saved := sty
	defer func() { sty = saved }()
	for _, th := range themes {
		sty = Style{on: true, theme: th.ID}
		c := &console{screen: true, raw: true, keyBar: true, cols: func() int { return 80 }}
		for name, s := range map[string]string{
			"key bar":     c.barLine(editBar),
			"menu bar":    c.barLine(menuBar),
			"hint":        promptLabel("Which concept?", "q cancel"),
			"long hint":   promptLabel("Pick one", strings.Repeat("type a number, ", 6)+"q cancel"),
			"merged hint": promptLabel("Which concept? (1–12)", "q cancel"),
		} {
			if strings.Contains(s, "\x1b[90m") || strings.Contains(s, "\x1b[2m") {
				t.Errorf("theme %s, %s: navigation text is grey or faint: %q", th.ID, name, s)
			}
			if !strings.Contains(stripANSI(s), "cancel") && !strings.Contains(stripANSI(s), "Enter") && !strings.Contains(stripANSI(s), "every key") {
				t.Errorf("theme %s, %s: lost its words: %q", th.ID, name, stripANSI(s))
			}
		}
	}
}
