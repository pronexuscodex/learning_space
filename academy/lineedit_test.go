package main

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

// edit feeds raw keystrokes to the line editor and returns the line, the
// error, and everything the editor wrote to the screen.
func edit(t *testing.T, keys string, onClear func()) (string, error, string) {
	t.Helper()
	var out bytes.Buffer
	c := &console{in: bufio.NewReader(strings.NewReader(keys)), out: &out, screen: true, onClear: onClear}
	line, err := c.editLine("› ")
	return line, err, out.String()
}

func TestLineEditorKeys(t *testing.T) {
	cases := []struct{ keys, want string }{
		{"hello\r", "hello"},
		{"helo\x7flo\r", "hello"},           // backspace
		{"wrong words\x15right\r", "right"}, // Ctrl+U
		{"two words\x17word\r", "two word"}, // Ctrl+W
		{"a\x1b[Ab\x1b[1;5Cc\r", "abc"},     // ↑ with no history, Ctrl+→ at the end: nothing
		{"x\ty\n", "x y"},                   // tab → space, \n ends
		{"  padded  \r", "padded"},          // sanitised
		{"héllo 😀\r", "héllo 😀"},            // multi-byte runes
	}
	for _, tc := range cases {
		if got, err, _ := edit(t, tc.keys, nil); err != nil || got != tc.want {
			t.Errorf("keys %q: got %q (err %v), want %q", tc.keys, got, err, tc.want)
		}
	}
}

func TestLineEditorCtrlL(t *testing.T) {
	redrawn := 0
	got, err, screen := edit(t, "ab\x0cc\r", func() { redrawn++ })
	if err != nil || got != "abc" {
		t.Fatalf("got %q, %v", got, err)
	}
	if !strings.Contains(screen, clearSeq) || redrawn != 1 {
		t.Fatalf("Ctrl+L should clear once and redraw once; redrawn=%d screen=%q", redrawn, screen)
	}
	if !strings.Contains(screen[strings.Index(screen, clearSeq):], "› ab") {
		t.Fatal("prompt and typed text should be redrawn after clearing")
	}
}

func TestLineEditorEOF(t *testing.T) {
	if _, err, _ := edit(t, "\x04", nil); err != io.EOF {
		t.Fatalf("Ctrl+D on an empty line should be EOF, got %v", err)
	}
	if got, err, _ := edit(t, "abc\x04\r", nil); err != nil || got != "abc" {
		t.Fatalf("Ctrl+D mid-line should be ignored, got %q %v", got, err)
	}
	if got, err, _ := edit(t, "partial", nil); err != nil || got != "partial" {
		t.Fatalf("input ending without Enter should return the text, got %q %v", got, err)
	}
}

func TestCookedCtrlLClearsAndAsksAgain(t *testing.T) {
	var out bytes.Buffer
	redrawn := 0
	c := &console{in: bufio.NewReader(strings.NewReader("\x0c\nreal answer\n")), out: &out, screen: true, onClear: func() { redrawn++ }}
	got, err := c.readLine("› ")
	if err != nil || got != "real answer" || redrawn != 1 || !strings.Contains(out.String(), clearSeq) {
		t.Fatalf("got %q err=%v redrawn=%d", got, err, redrawn)
	}
}

// TestReadlineKeys checks the editing keys shared with bash, zsh, Python
// and psql, one at a time.
func TestReadlineKeys(t *testing.T) {
	cases := []struct{ name, keys, want string }{
		{"← then insert", "helo\x1b[Dl\r", "hello"},
		{"Ctrl+B Ctrl+F", "ac\x02b\x06d\r", "abcd"},
		{"Ctrl+A inserts at the start", "world\x01hello \r", "hello world"},
		{"Home and End keys", "bc\x1b[Ha\x1b[Fd\r", "abcd"},
		{"VT100 Home and End", "bc\x1bOHa\x1bOFd\r", "abcd"},
		{"Home as ESC [1~", "bc\x1b[1~a\r", "abc"},
		{"Backspace before the cursor", "abxc\x02\x7f\r", "abc"},
		{"Delete under the cursor", "abxc\x02\x02\x1b[3~\r", "abc"},
		{"Ctrl+D under the cursor", "abxc\x02\x02\x04\r", "abc"},
		{"Ctrl+K cuts to the end", "keep this cut\x1b[1;5D\x0b\r", "keep this"},
		{"Ctrl+U cuts to the start only", "cut keep\x1b[1;5D\x15\r", "keep"},
		{"Ctrl+W cuts the word before the cursor", "one two three\x1b[1;5D\x17\r", "one three"},
		{"Ctrl+Y pastes the cut", "world hello\x17\x01\x19 \r", "hello world"},
		{"Alt+B and Alt+F", "a c\x1bb\x1bbX\x1bf\x1bfb\r", "Xa cb"},
		{"Ctrl+→ moves a word right", "one two\x01\x1b[1;5C!\r", "one! two"},
		{"Enter in the middle keeps the whole line", "whole line\x01\r", "whole line"},
		{"a lone Esc is ignored", "a\x1bb\x1b\r", "a"},
		{"unknown sequences are ignored", "a\x1b[5~\x1b[2~b\r", "ab"},
		{"wide characters", "😀b\x1b[Da\r", "😀ab"},
	}
	for _, tc := range cases {
		got, err, screen := edit(t, tc.keys, nil)
		if err != nil || got != tc.want {
			t.Errorf("%s: keys %q gave %q (err %v), want %q", tc.name, tc.keys, got, err, tc.want)
		}
		if strings.Contains(screen, "^[") {
			t.Errorf("%s: an escape sequence was echoed: %q", tc.name, screen)
		}
	}
}

// TestReadlineHistory checks ↑ and ↓ over earlier answers, keeping the
// line being typed.
func TestReadlineHistory(t *testing.T) {
	var out bytes.Buffer
	c := &console{in: bufio.NewReader(strings.NewReader(
		"first\r" +
			"second\r" +
			"second\r" + // repeats are kept once
			"\x1b[A\x1b[A\r" + // ↑↑: first
			"\x10!\r" + // Ctrl+P on its own: first again, edited
			"draft\x1b[A\x1b[B\r" + // ↑ then ↓: back to the draft
			"\x1b[A\x1b[A\x1b[A\x1b[A\x1b[A\x1b[A\r" + // past the oldest: stays there
			"   \r" + // blank: not remembered
			"\x1b[A\x0e\x0e\r")), // Ctrl+N past the newest: the (empty) draft
		out: &out, screen: true}
	want := []string{"first", "second", "second", "first", "first!", "draft", "first", "", ""}
	for i, w := range want {
		got, err := c.editLine("› ")
		if err != nil || got != w {
			t.Fatalf("line %d: got %q (%v), want %q", i+1, got, err, w)
		}
	}
	if strings.Join(c.history, ",") != "first,second,first,first!,draft,first" { // like bash with ignoredups
		t.Fatalf("history %q", c.history)
	}
}

// TestCursorOnWrappedLines checks that the cursor is put back in the right
// row and column when the input wraps past the terminal's width.
func TestCursorOnWrappedLines(t *testing.T) {
	var out bytes.Buffer
	c := &console{in: bufio.NewReader(strings.NewReader("abcdefghijklmnopqrstuvwxyz\x01\x05\x1b[D\x1b[D\r")), out: &out, screen: true,
		size: func() (int, int, bool) { return 10, 24, true }}
	got, err := c.editLine("› ") // 2 + 26 columns: three rows of ten
	if err != nil || got != "abcdefghijklmnopqrstuvwxyz" {
		t.Fatalf("got %q, %v", got, err)
	}
	screen := out.String()
	// Ctrl+A from the end (row 2): up two rows, column 2 after the prompt.
	if !strings.Contains(screen, "\x1b[2A\r\x1b[J› abcdefghijklmnopqrstuvwxyz\x1b[2A\r\x1b[2C") {
		t.Fatalf("Ctrl+A should redraw from the first row and put the cursor after the prompt: %q", screen)
	}
	// ← twice from the end: column 26 is row 2, column 6.
	if !strings.HasSuffix(screen, "\r\x1b[6C\x1b[2A\r\x1b[J› abcdefghijklmnopqrstuvwxyz\n") {
		t.Fatalf("Enter should move to the end before the new line: %q", screen)
	}
}

// TestKeyBar checks the bar above prompts: what it shows, where, and that
// it is gone once the line is answered.
func TestKeyBar(t *testing.T) {
	sized := func() (int, int, bool) { return 80, 24, true }
	t.Run("raw", func(t *testing.T) {
		var out bytes.Buffer
		c := &console{in: bufio.NewReader(strings.NewReader("hi\r")), out: &out, screen: true, raw: true, keyBar: true, size: sized,
			enterRaw: func() (func(), error) { return func() {}, nil }}
		got, err := c.readLine("  Name › ")
		if err != nil || got != "hi" {
			t.Fatalf("got %q, %v", got, err)
		}
		screen := stripANSI(out.String())
		if !strings.HasPrefix(screen, "  Enter answer  ←→ move  ↑↓ earlier answers  ^A ^E start, end") || !strings.Contains(screen, "\n  Name › hi") {
			t.Fatalf("the bar should sit on the line above the prompt: %q", screen)
		}
		if !strings.HasSuffix(out.String(), "\n\x1b[2A\r\x1b[M\x1b[1B") {
			t.Fatalf("after Enter the bar should be deleted, two rows up: %q", out.String())
		}
	})
	t.Run("narrow", func(t *testing.T) {
		c := &console{screen: true, raw: true, keyBar: true, cols: func() int { return 40 }, size: func() (int, int, bool) { return 40, 24, true }}
		bar := stripANSI(c.barLine(editBar))
		if visibleLen(bar) > 38 || !strings.HasPrefix(bar, "  Enter answer") {
			t.Fatalf("bar too wide or missing Enter at 40 columns: %q", bar)
		}
	})
	t.Run("line mode shows only keys that work there", func(t *testing.T) {
		c := &console{screen: true, keyBar: true}
		if bar := stripANSI(c.barLine(menuBar)); strings.Contains(bar, "^L") || strings.Contains(bar, "↑") || !strings.Contains(bar, "q quit") {
			t.Fatalf("line-mode bar %q", bar)
		}
	})
	t.Run("off", func(t *testing.T) {
		for _, c := range []*console{
			{screen: true, raw: true},                              // not a person at a terminal
			{screen: true, raw: true, keyBar: true, hideBar: true}, // switched off in Settings
		} {
			if bar := c.barLine(editBar); bar != "" {
				t.Fatalf("expected no bar, got %q", bar)
			}
		}
		var out bytes.Buffer
		c := &console{in: bufio.NewReader(strings.NewReader("\n")), out: &out, screen: true, keyBar: true}
		if _, err := c.readLine(""); err != nil || out.String() != "" {
			t.Fatalf("the pager's empty prompt has no bar: %q", out.String())
		}
	})
	t.Run("cooked rows", func(t *testing.T) {
		var out bytes.Buffer
		c := &console{in: bufio.NewReader(strings.NewReader(strings.Repeat("x", 30) + "\n")), out: &out, screen: true, keyBar: true,
			cols: func() int { return 20 }, size: func() (int, int, bool) { return 20, 24, true }}
		if _, err := c.readLine("› "); err != nil {
			t.Fatal(err)
		}
		// "› " + 30 characters is two rows at 20 columns, plus the bar.
		if !strings.HasSuffix(out.String(), "\x1b[3A\r\x1b[M\x1b[2B") {
			t.Fatalf("cooked erase: %q", out.String())
		}
	})
	t.Run("too tall to erase", func(t *testing.T) {
		var out bytes.Buffer
		c := &console{out: &out, size: func() (int, int, bool) { return 80, 3, true }}
		c.eraseBar(3)
		if out.Len() != 0 {
			t.Fatalf("a bar scrolled off the screen must be left alone: %q", out.String())
		}
	})
}
