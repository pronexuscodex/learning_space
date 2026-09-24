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
// already delivering single keys without echo.
func (c *console) editLine(prompt string) (string, error) {
	var buf []rune
	fmt.Fprint(c.out, prompt)
	redraw := func() {
		fmt.Fprint(c.out, "\r\x1b[K"+lastLine(prompt)+string(buf))
	}
	for {
		r, _, err := c.in.ReadRune()
		if err != nil {
			fmt.Fprintln(c.out)
			if err == io.EOF && len(buf) > 0 {
				return sanitize(string(buf)), nil
			}
			return "", err
		}
		switch {
		case r == '\r' || r == '\n':
			fmt.Fprint(c.out, "\n")
			return sanitize(string(buf)), nil
		case r == keyCtrlD:
			if len(buf) == 0 {
				fmt.Fprintln(c.out)
				return "", io.EOF
			}
		case r == keyCtrlC:
			// Only reached if the terminal does not raise SIGINT itself.
			fmt.Fprintln(c.out)
			return "", io.EOF
		case r == keyBackspace || r == keyDelete:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				redraw()
			}
		case r == keyCtrlU:
			buf = buf[:0]
			redraw()
		case r == keyCtrlW:
			buf = []rune(strings.TrimRight(string(buf), " "))
			for len(buf) > 0 && buf[len(buf)-1] != ' ' {
				buf = buf[:len(buf)-1]
			}
			redraw()
		case r == keyCtrlL:
			c.clearScreen()
			fmt.Fprint(c.out, prompt+string(buf))
		case r == keyEscape:
			c.skipEscapeSequence()
		case r == '\t':
			buf = append(buf, ' ')
			fmt.Fprint(c.out, " ")
		case unicode.IsPrint(r):
			buf = append(buf, r)
			fmt.Fprint(c.out, string(r))
		}
	}
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
