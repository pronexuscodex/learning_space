package main

// A small line editor for interactive terminals, with the keys of GNU
// readline (bash, Python, psql) and zsh's Emacs mode, so the habits learnt
// here work everywhere else, and the other way round:
//
//	← →  Ctrl+B Ctrl+F    one character left or right
//	Ctrl+← →  Alt+B Alt+F one word left or right
//	Home End  Ctrl+A Ctrl+E  the start or the end of the line
//	Backspace             delete the character before the cursor
//	Delete  Ctrl+D        delete the character under the cursor
//	                      (Ctrl+D on an empty line: end of input)
//	Ctrl+U  Ctrl+K        cut to the start or to the end of the line
//	Ctrl+W                cut the word before the cursor
//	Ctrl+Y                paste what was cut last
//	↑ ↓  Ctrl+P Ctrl+N    earlier and later answers (history)
//	Ctrl+L                clear the screen and redraw (the menu, then your prompt)
//	Ctrl+C                commit and exit (handled by the signal handler)
//
// Other escape sequences are ignored instead of appearing as ^[[5~. Where
// raw input is unavailable (pipes, Windows), input stays line-based and
// Ctrl+L followed by Enter still clears the screen.

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"
	"unicode"
)

// clearSeq moves the cursor home and clears the visible screen.
const clearSeq = "\x1b[H\x1b[2J"

// Key codes handled by the editor.
const (
	keyCtrlA     = 0x01
	keyCtrlB     = 0x02
	keyCtrlC     = 0x03
	keyCtrlD     = 0x04
	keyCtrlE     = 0x05
	keyCtrlF     = 0x06
	keyBackspace = 0x08
	keyCtrlK     = 0x0b
	keyCtrlL     = 0x0c
	keyCtrlN     = 0x0e
	keyCtrlP     = 0x10
	keyCtrlU     = 0x15
	keyCtrlW     = 0x17
	keyCtrlY     = 0x19
	keyEscape    = 0x1b
	keyDelete    = 0x7f
)

// historySize is how many earlier answers ↑ can bring back.
const historySize = 200

// The active raw-mode restore function, so that exits and signals can put
// the terminal back the way they found it.
var (
	termMu      sync.Mutex
	termRestore func()
)

// restoreTerminal undoes raw mode if it is active. Safe to call any time.
func restoreTerminal() {
	termMu.Lock()
	defer termMu.Unlock()
	if termRestore != nil {
		termRestore()
		termRestore = nil
	}
}

// clearScreen wipes the terminal (only when output is a terminal) and runs
// the redraw hook, if any.
func (c *console) clearScreen() {
	if c.screen {
		fmt.Fprint(c.out, clearSeq)
	}
	if c.onClear != nil {
		c.onClear()
	}
}

// readLineRaw reads one line key by key with the shortcuts above.
func (c *console) readLineRaw(prompt string) (string, error) {
	restore, err := c.rawMode()
	if err != nil {
		c.raw = false // fall back to line mode for the rest of the session
		c.prefill = nil
		return c.readLineCooked(prompt)
	}
	termMu.Lock()
	termRestore = restore
	termMu.Unlock()
	defer restoreTerminal()
	return c.editLine(prompt)
}

// editLine is the key-by-key editing loop; it assumes the terminal is
// already delivering single keys without echo. Screen writes happen under
// c.mu so that a resize redraw (redrawForResize) never interleaves with
// them.
func (c *console) editLine(prompt string) (string, error) {
	c.mu.Lock()
	c.editing, c.curPrompt, c.curBuf = true, prompt, c.prefill
	c.cur, c.curRow = len(c.curBuf), 0
	c.histPos, c.histDraft = len(c.history), nil
	c.prefill = nil
	c.paint(prompt)
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.editing = false
		c.mu.Unlock()
	}()

	for {
		r, _, err := c.in.ReadRune() // blocks without holding the lock
		c.mu.Lock()
		line, done, err := c.editKey(r, err)
		c.mu.Unlock()
		if done {
			return line, err
		}
	}
}

// editKey applies one key to the line being edited. Caller holds c.mu.
func (c *console) editKey(r rune, err error) (string, bool, error) {
	buf := c.curBuf
	if err != nil {
		c.toEnd()
		fmt.Fprintln(c.out)
		if err == io.EOF && len(buf) > 0 {
			return sanitize(string(buf)), true, nil
		}
		return "", true, err
	}
	if r == keyEscape {
		r = escapeKey(c.readEscape())
	}
	switch {
	case r == '\r' || r == '\n':
		c.toEnd()
		c.inputRows = c.curRow + 1
		if !c.inline {
			fmt.Fprint(c.out, "\n")
		}
		c.remember(string(buf))
		return sanitize(string(buf)), true, nil
	case r == keyCtrlD && len(buf) == 0:
		fmt.Fprintln(c.out)
		return "", true, io.EOF
	case r == keyCtrlC:
		// Only reached if the terminal does not raise SIGINT itself.
		c.toEnd()
		fmt.Fprintln(c.out)
		return "", true, io.EOF
	case r == keyBackspace || r == keyDelete:
		if c.cur > 0 {
			c.setLine(slices.Delete(buf, c.cur-1, c.cur), c.cur-1)
		}
	case r == keyCtrlD || r == keyDeleteUnder:
		if c.cur < len(buf) {
			c.setLine(slices.Delete(buf, c.cur, c.cur+1), c.cur)
		}
	case r == keyCtrlA:
		c.moveTo(0)
	case r == keyCtrlE:
		c.moveTo(len(buf))
	case r == keyCtrlB:
		c.moveTo(max(0, c.cur-1))
	case r == keyCtrlF:
		c.moveTo(min(len(buf), c.cur+1))
	case r == keyWordLeft:
		c.moveTo(wordStart(buf, c.cur))
	case r == keyWordRight:
		c.moveTo(wordEnd(buf, c.cur))
	case r == keyCtrlU:
		c.cut(0, c.cur)
	case r == keyCtrlK:
		c.cut(c.cur, len(buf))
	case r == keyCtrlW:
		c.cut(wordStart(buf, c.cur), c.cur)
	case r == keyCtrlY:
		c.insert(c.killed...)
	case r == keyCtrlP:
		c.recall(-1)
	case r == keyCtrlN:
		c.recall(+1)
	case r == keyCtrlL:
		c.clearScreen()
		c.paint(c.curPrompt)
	case r == '\t':
		c.insert(' ')
	case unicode.IsPrint(r):
		c.insert(r)
	}
	return "", false, nil
}

// Keys that arrive as escape sequences, given codes beyond the last
// Unicode character so that editKey can treat them like control keys.
const (
	keyNone        = -1
	keyDeleteUnder = unicode.MaxRune + iota
	keyWordLeft
	keyWordRight
)

// escapeKey maps the escape sequences of the usual terminals (xterm,
// the Linux console, VT100 application mode, macOS Terminal) to keys.
func escapeKey(seq string) rune {
	switch seq {
	case "[A", "OA":
		return keyCtrlP
	case "[B", "OB":
		return keyCtrlN
	case "[C", "OC":
		return keyCtrlF
	case "[D", "OD":
		return keyCtrlB
	case "[H", "OH", "[1~", "[7~":
		return keyCtrlA
	case "[F", "OF", "[4~", "[8~":
		return keyCtrlE
	case "[3~":
		return keyDeleteUnder
	case "[1;5D", "[1;3D", "b", "B":
		return keyWordLeft
	case "[1;5C", "[1;3C", "f", "F":
		return keyWordRight
	}
	return keyNone
}

// insert types runes at the cursor. At the end of the line they are simply
// printed; in the middle the line is redrawn.
func (c *console) insert(rs ...rune) {
	if len(rs) == 0 {
		return
	}
	if c.inline { // the status line holds one row
		room := c.termCols() - 1
		if room <= 0 {
			room = 62
		}
		if c.shown+visibleLen(string(rs)) > room {
			return
		}
	}
	if c.cur == len(c.curBuf) {
		c.curBuf = append(c.curBuf, rs...)
		c.cur = len(c.curBuf)
		fmt.Fprint(c.out, string(rs))
		c.shown += visibleLen(string(rs))
		c.curRow = c.endRow()
		return
	}
	c.setLine(slices.Insert(c.curBuf, c.cur, rs...), c.cur+len(rs))
}

// cut removes buf[from:to] and keeps it for Ctrl+Y, as readline does.
func (c *console) cut(from, to int) {
	if from >= to {
		return
	}
	c.killed = slices.Clone(c.curBuf[from:to])
	c.setLine(slices.Delete(c.curBuf, from, to), from)
}

// setLine replaces the line and the cursor position, and redraws.
func (c *console) setLine(buf []rune, cur int) {
	c.curBuf, c.cur = buf, cur
	c.redrawLine()
}

// moveTo puts the cursor at position i of the line.
func (c *console) moveTo(i int) {
	if i != c.cur {
		c.cur = i
		c.redrawLine()
	}
}

// toEnd moves the cursor to the end of the line, so that Enter starts the
// next line below all of it, not in the middle of a wrapped answer.
func (c *console) toEnd() {
	if c.cur < len(c.curBuf) {
		c.moveTo(len(c.curBuf))
	}
}

// recall steps through earlier answers (dir -1) or back towards the line
// being typed (dir +1), which is kept while you look.
func (c *console) recall(dir int) {
	pos := c.histPos + dir
	if pos < 0 || pos > len(c.history) {
		return
	}
	if c.histPos == len(c.history) {
		c.histDraft = slices.Clone(c.curBuf)
	}
	c.histPos = pos
	line := c.histDraft
	if pos < len(c.history) {
		line = []rune(c.history[pos])
	}
	c.setLine(slices.Clone(line), len(line))
}

// remember adds an answer to the history, skipping blanks and repeats.
func (c *console) remember(line string) {
	line = sanitize(line)
	if line == "" || (len(c.history) > 0 && c.history[len(c.history)-1] == line) {
		return
	}
	c.history = append(c.history, line)
	if len(c.history) > historySize {
		c.history = c.history[len(c.history)-historySize:]
	}
}

// wordStart is where the word before position i begins (Ctrl+W, Alt+B):
// skip spaces, then everything up to the previous space.
func wordStart(buf []rune, i int) int {
	for i > 0 && buf[i-1] == ' ' {
		i--
	}
	for i > 0 && buf[i-1] != ' ' {
		i--
	}
	return i
}

// wordEnd is where the word after position i ends (Alt+F).
func wordEnd(buf []rune, i int) int {
	for i < len(buf) && buf[i] == ' ' {
		i++
	}
	for i < len(buf) && buf[i] != ' ' {
		i++
	}
	return i
}

// paint writes text (the whole prompt, or its last line) followed by the
// input, then moves the cursor back to its place in the input. The input
// may wrap onto several terminal rows; c.curRow records the row the cursor
// ends on, counted from the input's first row. Caller holds c.mu.
func (c *console) paint(text string) {
	fmt.Fprint(c.out, text+string(c.curBuf))
	c.shown = visibleLen(lastLine(c.curPrompt) + string(c.curBuf))
	end := c.endRow()
	c.curRow = end
	if c.cur >= len(c.curBuf) {
		return
	}
	col := visibleLen(lastLine(c.curPrompt) + string(c.curBuf[:c.cur]))
	row := 0
	if cols := c.termCols(); cols > 0 {
		row, col = col/cols, col%cols
	}
	if up := end - row; up > 0 {
		fmt.Fprintf(c.out, "\x1b[%dA", up)
	}
	fmt.Fprint(c.out, "\r")
	if col > 0 {
		fmt.Fprintf(c.out, "\x1b[%dC", col)
	}
	c.curRow = row
}

// endRow is the row the end of the input is on. A line that exactly fills
// its last row leaves the cursor there (terminals wrap only when the next
// character comes), hence shown-1.
func (c *console) endRow() int {
	if cols := c.termCols(); cols > 0 && c.shown > 0 {
		return (c.shown - 1) / cols
	}
	return 0
}

// redrawLine repaints the input line, even when it has wrapped onto several
// terminal rows: move up to its first row, clear to the end of the screen,
// and draw it again at the current width. Caller holds c.mu.
func (c *console) redrawLine() {
	if c.curRow > 0 {
		fmt.Fprintf(c.out, "\x1b[%dA", c.curRow)
	}
	fmt.Fprint(c.out, "\r\x1b[J")
	c.paint(lastLine(c.curPrompt))
}

// redrawForResize re-lays out the current screen after the terminal is
// resized. Only screens with a redraw hook (the main menu) are redrawn
// straight away; every other screen picks up the new size when it is next
// drawn, or on Ctrl+L.
func (c *console) redrawForResize() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.onResize != nil { // a full-screen view redraws itself
		c.onResize()
		return
	}
	if !c.editing || c.onClear == nil {
		return
	}
	c.clearScreen()
	c.paint(c.curPrompt)
}

// termCols is the terminal's real width in columns (0 if unknown), used to
// work out how many rows a wrapped input line occupies.
func (c *console) termCols() int {
	if !c.screen {
		return 0
	}
	if cols, _, ok := c.termSize(); ok {
		return cols
	}
	return 0
}

// termSize is the terminal's size in columns and rows.
func (c *console) termSize() (cols, rows int, ok bool) {
	if c.size != nil {
		return c.size()
	}
	return terminalSize(int(os.Stdout.Fd()))
}

// readEscape reads the rest of an escape sequence after Esc: "[A" for ↑,
// "[3~" for Delete, "b" for Alt+B. A lone Esc press gives "".
func (c *console) readEscape() string {
	if c.in.Buffered() == 0 {
		return ""
	}
	r, _, err := c.in.ReadRune()
	if err != nil {
		return ""
	}
	if r != '[' && r != 'O' {
		return string(r) // Alt+key
	}
	seq := []byte{byte(r)}
	for c.in.Buffered() > 0 {
		b, err := c.in.ReadByte()
		if err != nil {
			break
		}
		seq = append(seq, b)
		if b >= 0x40 && b <= 0x7e {
			break
		}
	}
	return string(seq)
}

// lastLine returns the part of a prompt after its final newline, which is
// the part that shares the line with the input.
func lastLine(s string) string {
	if i := strings.LastIndex(s, "\n"); i >= 0 {
		return s[i+1:]
	}
	return s
}
