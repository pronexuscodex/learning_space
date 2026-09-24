package main

// A tiny line editor for interactive terminals. It reads one key at a time
// so that shell-style shortcuts work immediately:
//
//	Ctrl+L       clear the screen and redraw (the menu, then your prompt)
//	Backspace    delete the previous character
//	Ctrl+U       erase the whole line        Ctrl+W   erase the previous word
//	Ctrl+D       end of input (on an empty line)
//	Ctrl+C       commit and exit (handled by the signal handler)
//
// Arrow keys and other escape sequences are ignored instead of appearing
// as ^[[A. Where raw input is unavailable (pipes, Windows), input stays
// line-based and Ctrl+L followed by Enter still clears the screen.

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode"
)

// clearSeq moves the cursor home and clears the visible screen.
const clearSeq = "\x1b[H\x1b[2J"

// Key codes handled by the editor.
const (
	keyCtrlC     = 0x03
	keyCtrlD     = 0x04
	keyBackspace = 0x08
	keyCtrlL     = 0x0c
	keyCtrlU     = 0x15
	keyCtrlW     = 0x17
	keyEscape    = 0x1b
	keyDelete    = 0x7f
)

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
	restore, err := enableRawInput(int(os.Stdin.Fd()))
	if err != nil {
		c.raw = false // fall back to line mode for the rest of the session
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
	c.editing, c.curPrompt, c.curBuf = true, prompt, nil
	fmt.Fprint(c.out, prompt)
	c.shown = visibleLen(lastLine(prompt)) // columns used by the input line as drawn
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
	defer func() { c.curBuf = buf }()
	if err != nil {
		fmt.Fprintln(c.out)
		if err == io.EOF && len(buf) > 0 {
			return sanitize(string(buf)), true, nil
		}
		return "", true, err
	}
	switch {
	case r == '\r' || r == '\n':
		fmt.Fprint(c.out, "\n")
		return sanitize(string(buf)), true, nil
	case r == keyCtrlD:
		if len(buf) == 0 {
			fmt.Fprintln(c.out)
			return "", true, io.EOF
		}
	case r == keyCtrlC:
		// Only reached if the terminal does not raise SIGINT itself.
		fmt.Fprintln(c.out)
		return "", true, io.EOF
	case r == keyBackspace || r == keyDelete:
		if len(buf) > 0 {
			buf = buf[:len(buf)-1]
			c.redrawLine(buf)
		}
	case r == keyCtrlU:
		buf = buf[:0]
		c.redrawLine(buf)
	case r == keyCtrlW:
		buf = []rune(strings.TrimRight(string(buf), " "))
		for len(buf) > 0 && buf[len(buf)-1] != ' ' {
			buf = buf[:len(buf)-1]
		}
		c.redrawLine(buf)
	case r == keyCtrlL:
		c.clearScreen()
		fmt.Fprint(c.out, c.curPrompt+string(buf))
		c.shown = visibleLen(lastLine(c.curPrompt) + string(buf))
	case r == keyEscape:
		c.skipEscapeSequence()
	case r == '\t':
		buf = append(buf, ' ')
		fmt.Fprint(c.out, " ")
		c.shown++
	case unicode.IsPrint(r):
		buf = append(buf, r)
		fmt.Fprint(c.out, string(r))
		c.shown += runeWidth(r)
	}
	return "", false, nil
}

// redrawLine repaints the input line, even when it has wrapped onto several
// terminal rows: move up to its first row, clear to the end of the screen,
// and draw it again at the current width. Caller holds c.mu.
func (c *console) redrawLine(buf []rune) {
	if cols := c.termCols(); cols > 0 && c.shown > 0 {
		if up := (c.shown - 1) / cols; up > 0 {
			fmt.Fprintf(c.out, "\x1b[%dA", up)
		}
	}
	line := lastLine(c.curPrompt) + string(buf)
	fmt.Fprint(c.out, "\r\x1b[J"+line)
	c.shown = visibleLen(line)
}

// redrawForResize re-lays out the current screen after the terminal is
// resized. Only screens with a redraw hook (the main menu) are redrawn
// straight away; every other screen picks up the new size when it is next
// drawn, or on Ctrl+L.
func (c *console) redrawForResize() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.editing || c.onClear == nil {
		return
	}
	c.clearScreen()
	fmt.Fprint(c.out, c.curPrompt+string(c.curBuf))
	c.shown = visibleLen(lastLine(c.curPrompt) + string(c.curBuf))
}

// termCols is the terminal's real width in columns (0 if unknown), used to
// work out how many rows a wrapped input line occupies.
func (c *console) termCols() int {
	if !c.screen {
		return 0
	}
	if cols, _, ok := terminalSize(int(os.Stdout.Fd())); ok {
		return cols
	}
	return 0
}

// skipEscapeSequence consumes the rest of an escape sequence such as an
// arrow key (ESC [ A). A lone Esc press is simply ignored.
func (c *console) skipEscapeSequence() {
	if c.in.Buffered() == 0 {
		return
	}
	r, _, err := c.in.ReadRune()
	if err != nil || (r != '[' && r != 'O') {
		return
	}
	for c.in.Buffered() > 0 {
		b, err := c.in.ReadByte()
		if err != nil || (b >= 0x40 && b <= 0x7e) {
			return
		}
	}
}

// lastLine returns the part of a prompt after its final newline, which is
// the part that shares the line with the input.
func lastLine(s string) string {
	if i := strings.LastIndex(s, "\n"); i >= 0 {
		return s[i+1:]
	}
	return s
}
