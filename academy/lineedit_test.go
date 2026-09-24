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
		{"a\x1b[Ab\x1b[1;5Cc\r", "abc"},     // arrow keys ignored
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
