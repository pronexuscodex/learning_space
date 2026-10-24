package main

// The key bar: one line at the foot of the screen, just above the prompt,
// naming the keys that work right now, like the bottom lines of nano and
// of the 1980s word processors. It disappears once the line is answered,
// so a long session does not leave a trail of them behind.

import (
	"fmt"
	"os"
	"strings"
)

// barKey is one entry of the key bar. Raw entries work only with the line
// editor (single keys); the rest also work when the terminal hands over
// whole lines.
type barKey struct {
	key, what string
	raw       bool
}

// Ordered by importance: on a narrow screen the bar keeps the first ones.
var (
	editBar = []barKey{
		{"Enter", "answer", false},
		{"←→", "move", true},
		{"↑↓", "earlier answers", true},
		{"^A ^E", "start, end", true},
		{"^W ^U", "cut word, line", true},
		{"^Y", "paste", true},
		{"^L", "redraw", true},
	}
	menuBar = []barKey{
		{"key+Enter", "open", false},
		{"?", "every key", false},
		{"0", "start here", false},
		{"q", "quit", false},
		{"↑", "last choice", true},
		{"^L", "redraw", true},
	}
	paragraphBar = []barKey{
		{"Enter", "next line", false},
		{"empty line", "finish", false},
		{":q", "cancel", false},
		{"←→", "move", true},
		{"↑", "earlier", true},
		{"^U", "cut", true},
	}
)

// barLine renders the key bar for a prompt, fitted to the screen, or ""
// when there is to be none.
func (c *console) barLine(keys []barKey) string {
	if !c.keyBar || c.hideBar || len(keys) == 0 || os.Getenv("TERM") == "dumb" {
		return ""
	}
	room := c.width() - 2
	if cols := c.termCols(); cols > 0 {
		room = min(room, cols-3) // never wrap: erasing counts on one row
	}
	var parts []string
	used := 0
	for _, k := range keys {
		if k.raw && !c.raw {
			continue
		}
		n := visibleLen(k.key) + 1 + visibleLen(k.what)
		if len(parts) > 0 {
			n += 2
		}
		if used+n > room {
			break
		}
		used += n
		parts = append(parts, sty.Bold(sty.Cyan(k.key))+" "+sty.Gray(k.what))
	}
	if len(parts) == 0 {
		return ""
	}
	return "  " + strings.Join(parts, "  ")
}

// eraseBar deletes the key bar, which is rows lines above the cursor,
// and puts the cursor back where the answer left it. A bar that has
// scrolled off the screen (a very tall prompt) is left alone.
func (c *console) eraseBar(rows int) {
	if _, h, ok := c.termSize(); !ok || rows >= h || rows < 1 {
		return
	}
	fmt.Fprintf(c.out, "\x1b[%dA\r\x1b[M", rows) // DL: the lines below move up
	if rows > 1 {
		fmt.Fprintf(c.out, "\x1b[%dB", rows-1)
	}
}
