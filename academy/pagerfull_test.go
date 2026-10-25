package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// fullRun shows 40 numbered lines in the full-screen pager (a 12-row
// terminal: 11 rows of text and the status line), pressing keys.
func fullRun(t *testing.T, keys string, finalPause bool) (string, *App) {
	t.Helper()
	t.Setenv("LINES", "12")
	t.Setenv("COLUMNS", "100")
	t.Setenv("TERM", "xterm-256color")
	var out bytes.Buffer
	a := &App{
		reg:   seedRegistry(),
		width: 100,
		pager: true,
		con: &console{
			in: bufio.NewReader(strings.NewReader(keys)), out: &out,
			raw: true, screen: true,
			enterRaw: func() (func(), error) { return func() {}, nil },
		},
	}
	a.paged(finalPause, func() {
		for i := 1; i <= 40; i++ {
			a.println(fmt.Sprintf("line %d", i))
		}
	})
	return out.String(), a
}

// Like man: the text fills the screen, only a status line sits under it,
// keys act at once, and the screen is restored when the reader leaves.
func TestFullScreenPager(t *testing.T) {
	out, a := fullRun(t, " q", false)
	for _, want := range []string{altScreenOn, "lines 1–11 of 40", "lines 12–22 of 40", altScreenOff} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out, "Enter more") || strings.Contains(out, "or type your choice") {
		t.Error("the full-screen pager has no prompt line in the text")
	}
	// After it, the part last on screen is shown again for what follows.
	after := out[strings.LastIndex(out, altScreenOff):]
	if !strings.Contains(after, "line 12\n") || !strings.Contains(after, "line 22") || strings.Contains(after, "line 23") {
		t.Errorf("the page last read must be shown after leaving:\n%q", after)
	}
	if a.con.prefill != nil {
		t.Error("q leaves without typing anything ahead")
	}

	// End shows (END); arrows and b move; g goes to the start.
	out, _ = fullRun(t, "G\x1b[A"+"bgq", true)
	for _, want := range []string{"(END) lines 30–40 of 40", "lines 29–39 of 40", "lines 18–28 of 40"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Count(out, "lines 1–11 of 40") != 2 {
		t.Error("g must go back to the start")
	}
	if after := out[strings.LastIndex(out, altScreenOff):]; strings.Contains(after, "line") {
		t.Error("with finalPause (the menu comes next), nothing is printed after leaving, as with man")
	}

	// /word finds, n and N move between matches.
	out, _ = fullRun(t, "/line 3\rnNq", false)
	for _, want := range []string{"“line 3” 1/11", "“line 3” 2/11"} {
		if !strings.Contains(out, want) {
			t.Errorf("search: missing %q", want)
		}
	}

	// The search line has the editing keys of every other prompt: here
	// "ine 3" is typed, Ctrl+A goes to the start, l is put in front, and
	// ↑ at the next / brings the search back.
	out, a = fullRun(t, "/ine 3\x01l\r/\x1b[A\rq", false)
	if strings.Count(out, "“line 3” 1/11") != 2 {
		t.Errorf("search editing and history: %q", out)
	}
	if len(a.con.history) != 0 || len(a.con.searches) != 1 {
		t.Errorf("searches have their own history: answers %q, searches %q", a.con.history, a.con.searches)
	}
	if _, a = fullRun(t, "/x\x7f\x7fq", false); a.con.inline {
		t.Error("Backspace on an empty search line cancels it, as in less")
	}

	// A number leaves and answers the next prompt.
	_, a = fullRun(t, "7\n", false)
	if string(a.con.prefill) != "7" {
		t.Fatalf("a number must be kept for the next prompt, got %q", string(a.con.prefill))
	}
	n, err := a.con.promptChoice("Pick", []string{"a", "b", "c", "d", "e", "f", "g", "h"})
	if err != nil || n != 6 {
		t.Errorf("the 7 typed in the pager must answer the next prompt: got %d, %v", n, err)
	}

	// h shows the keys.
	out, _ = fullRun(t, "h q", false)
	if !strings.Contains(out, "READING A LONG SCREEN") {
		t.Error("h must show the keys")
	}
}

func TestFullScreenStatusFits(t *testing.T) {
	t.Setenv("LINES", "12")
	for _, cols := range []int{40, 60, 80, 130} {
		t.Setenv("COLUMNS", fmt.Sprint(cols))
		p := &fullPager{a: &App{con: &console{}}, lines: make([]string, 40)}
		p.find.current = -1
		for _, top := range []int{0, 29} {
			p.top = top
			if w := visibleLen(p.status()); w >= cols {
				t.Errorf("%d columns: status line is %d wide", cols, w)
			}
		}
	}
}

func TestReadKeyParsesSpecialKeys(t *testing.T) {
	c := &console{in: bufio.NewReader(strings.NewReader("\x1b[A\x1b[B\x1b[5~\x1b[6~\x1bOH\x1b[F\x1b[Cx"))}
	want := []string{keyUp, keyDown, keyPageUp, keyPageDown, keyHome, keyEnd, ""}
	for _, w := range want {
		if _, name, _ := c.readKey(); name != w {
			t.Errorf("got %q, want %q", name, w)
		}
	}
	if r, name, _ := c.readKey(); r != 'x' || name != "" {
		t.Errorf("a plain key after an escape sequence: got %q %q", r, name)
	}
}

// resizeReader asks for a resize redraw before handing over its keys, as
// a window resize would while the pager waits for a key.
type resizeReader struct {
	con  *console
	keys string
	done bool
}

func (r *resizeReader) Read(b []byte) (int, error) {
	if !r.done {
		r.done = true
		r.con.redrawForResize()
	}
	return strings.NewReader(r.keys).Read(b)
}

func TestFullScreenRedrawsOnResize(t *testing.T) {
	t.Setenv("LINES", "12")
	t.Setenv("COLUMNS", "100")
	t.Setenv("TERM", "xterm-256color")
	var out bytes.Buffer
	c := &console{out: &out, raw: true, screen: true, enterRaw: func() (func(), error) { return func() {}, nil }}
	c.in = bufio.NewReader(&resizeReader{con: c, keys: "q"})
	a := &App{reg: seedRegistry(), width: 100, pager: true, con: c}
	a.paged(true, func() {
		for i := 1; i <= 40; i++ {
			a.println(fmt.Sprintf("line %d", i))
		}
	})
	if n := strings.Count(out.String(), "lines 1–11 of 40"); n != 2 {
		t.Errorf("a resize must redraw the page at once: %d draws", n)
	}
	if c.onResize != nil {
		t.Error("the resize hook must be removed when the pager closes")
	}
}

// At 80 columns the page keeps its name on every page (it does not vanish
// as the line numbers grow), narrow screens still name Space, and the last
// page offers b back instead of a Space that goes nowhere.
func TestFullScreenStatusIsSteady(t *testing.T) {
	t.Setenv("LINES", "30")
	t.Setenv("COLUMNS", "80")
	p := &fullPager{a: &App{con: &console{}}, lines: make([]string, 269), home: true, title: "START HERE · how to learn wi…"}
	p.find.current = -1
	for _, top := range []int{0, 116, 240} {
		p.top = top
		s := stripANSI(p.status())
		if !strings.HasPrefix(s, " START HERE · lines") && !strings.HasPrefix(s, " START HERE · (END)") {
			t.Errorf("top %d: the page's name is missing: %q", top, s)
		}
		if strings.Contains(s, "how to") {
			t.Errorf("top %d: a cut-off subtitle: %q", top, s)
		}
		if end := strings.Contains(s, "(END)"); end == strings.Contains(s, "Space next") {
			t.Errorf("top %d: Space next must be offered on every page but the last: %q", top, s)
		}
	}
	for _, cols := range []int{40, 50, 60} {
		t.Setenv("COLUMNS", fmt.Sprint(cols))
		p.top = 116
		if s := stripANSI(p.status()); !strings.Contains(s, "Space next") {
			t.Errorf("%d columns: Space is not named: %q", cols, s)
		}
	}
}

// Every key that leaves a menu page must open something at the menu: a
// key with no screen behind it closes the page and only gets a warning.
func TestHomeKeysAreMenuKeys(t *testing.T) {
	a := &App{reg: seedRegistry(), con: &console{out: &bytes.Buffer{}}, width: 80}
	menu := map[rune]bool{'r': true, 's': true, '5': true, '7': true} // aliases: r = 9, s = /, 5 = q; 7 saves
	for _, g := range a.allScreens() {
		for _, e := range g.entries {
			text := stripANSI(e.text)
			if i, j := strings.Index(text, "["), strings.Index(text, "]"); i >= 0 && j == i+2 {
				menu[rune(text[i+1])] = true
			}
		}
	}
	for _, r := range homeKeys {
		if !menu[r] {
			t.Errorf("homeKeys has %q, but no menu item opens with it", r)
		}
	}
}
