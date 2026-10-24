package main

// Sanitised, line-based terminal input over a bufio.Reader. Every prompt
// loops until it gets valid input, the user cancels, or input ends (EOF).

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// errCancel signals that the user backed out of a prompt.
var errCancel = errors.New("cancelled")

// console wraps a buffered input stream and an output writer.
type console struct {
	in  *bufio.Reader
	out io.Writer

	raw     bool   // read key by key (interactive terminal with raw-mode support)
	screen  bool   // output is a terminal, so clearing the screen is meaningful
	keyBar  bool   // a person types at a terminal: show the key bar (keybar.go)
	hideBar bool   // the learner switched the key bar off in Settings
	onClear func() // redraws the current screen after Ctrl+L (may be nil)

	// Live editor state, guarded by mu, so a terminal resize can redraw
	// the screen and the half-typed line from another goroutine.
	mu        sync.Mutex
	editing   bool
	curPrompt string
	curBuf    []rune
	cur       int // cursor position in curBuf
	curRow    int // terminal row of the cursor, from the input's first row
	shown     int
	inputRows int // rows the last answered line took (for erasing the key bar)

	// Readline's memory: earlier answers for ↑ and ↓ (this session only),
	// where ↑ has got to, the line you were typing, and the last cut.
	history   []string
	histPos   int
	histDraft []rune
	killed    []rune
	searches  []string   // the pager's / history, kept apart as less does
	inline    bool       // editing the pager's status line: one row only
	cols      func() int // current layout width (nil: 80)

	// pending is an answer typed ahead at a pager prompt; the next
	// readLine returns it instead of waiting for input.
	pending *string

	// prefill is a key typed in the full-screen pager to answer what
	// comes next: the next raw prompt starts with it, ready to go on.
	prefill []rune

	// onResize, when set, redraws a full-screen view after the terminal
	// is resized (called under mu, before any other redraw).
	onResize func()

	// enterRaw switches to single keys (tests replace it); nil means the
	// real terminal.
	enterRaw func() (func(), error)

	// size reports the terminal's columns and rows (tests replace it);
	// nil means the real terminal.
	size func() (cols, rows int, ok bool)

	// closed is set once input has ended (Ctrl+D or end of a pipe). It is
	// sticky, so every later prompt also sees the end of input and the
	// app always reaches its commit-and-exit path.
	closed bool
}

// setOnClear installs (or removes, with nil) the redraw hook, under the
// lock shared with the resize handler.
func (c *console) setOnClear(f func()) {
	c.mu.Lock()
	c.onClear = f
	c.mu.Unlock()
}

// setOnResize installs (or removes) the full-screen redraw hook.
func (c *console) setOnResize(f func()) {
	c.mu.Lock()
	c.onResize = f
	c.mu.Unlock()
}

// width is the current layout width for messages.
func (c *console) width() int {
	if c.cols != nil {
		return c.cols()
	}
	return 80
}

// say prints a status message with a marker, wrapped to the terminal.
func (c *console) say(marker, text string, style func(string) string) {
	indent := 3 + visibleLen(marker) // "  ✓ "; emoji markers are two columns wide
	for i, l := range wrap(text, c.width()-indent, "") {
		lead := strings.Repeat(" ", indent)
		if i == 0 {
			lead = "  " + marker + " "
		}
		fmt.Fprintln(c.out, lead+style(l))
	}
}

// Feedback lines, one visual vocabulary for the whole app.
func (c *console) ok(format string, args ...any) {
	c.say(sty.Green("✓"), fmt.Sprintf(format, args...), func(s string) string { return s })
}
func (c *console) warn(format string, args ...any) {
	c.say(sty.Yellow("!"), fmt.Sprintf(format, args...), sty.Yellow)
}
func (c *console) fail(format string, args ...any) {
	c.say(sty.Red("✗"), fmt.Sprintf(format, args...), sty.Red)
}
func (c *console) note(format string, args ...any) {
	c.say(sty.Dim("·"), fmt.Sprintf(format, args...), sty.Dim)
}

// layoutWidth reports the current layout width to package-level helpers
// (set by main; 80 when unset).
var layoutWidth = func() int { return 80 }

// promptLabel renders "  Label (hint) › ". Long labels wrap to the
// screen's width, keeping room on the last line to type; a label that is
// a list of choices ("a add · e edit · q back") breaks only between
// choices. A hint that does not fit goes on its own line above.
func promptLabel(label, hint string) string {
	// One set of brackets: "Which concept? (1–12, q cancel)", not
	// "Which concept? (1–12) (q cancel)".
	if i := strings.LastIndex(label, " ("); hint != "" && i >= 0 && strings.HasSuffix(label, ")") && !strings.ContainsAny(label[i+2:len(label)-1], "()") {
		label, hint = label[:i], label[i+2:len(label)-1]+", "+hint
	}
	w := layoutWidth()
	room := max(24, w-12) // the input shares the last line
	lines := breakChoices(label, room)
	for i := range lines {
		lines[i] = "  " + sty.Bold(lines[i])
	}
	s := strings.Join(lines, "\n") // long questions wrap; the input stays on the last line
	if hint != "" {
		if visibleLen(lines[len(lines)-1])+visibleLen(hint)+6 <= w-6 { // still room to type
			s += " " + sty.Gray("("+hint+")")
		} else {
			// Too long to share the line: the hint goes on its own line
			// above the question, never dropped (it often says what to type).
			var above []string
			for _, l := range breakChoices("("+hint+")", w-4) {
				above = append(above, sty.Gray("  "+l))
			}
			s = strings.Join(above, "\n") + "\n" + s
		}
	}
	return s + sty.Cyan(" › ")
}

// breakChoices wraps text to width; when it is a list of choices joined
// by " · ", it breaks only between choices.
func breakChoices(text string, width int) []string {
	if !strings.Contains(text, " · ") {
		return wrap(text, width, "")
	}
	var lines []string
	for _, l := range flow(strings.Split(text, " · "), " · ", width, "") {
		lines = append(lines, wrap(l, width, "")...) // a single choice wider than the line still wraps
	}
	return lines
}

// sanitize strips control characters (tabs become spaces), invalid UTF-8
// and surrounding whitespace from a raw input line.
func sanitize(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case r == utf8.RuneError, unicode.IsControl(r):
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

// isCancel reports whether the input is a cancel token.
func isCancel(s string) bool {
	s = strings.ToLower(s)
	return s == "q" || s == ":q"
}

// readLine prints prompt and returns one sanitized line. It returns io.EOF
// when input is exhausted (Ctrl-D / closed pipe). Ctrl+L clears the screen.
// The key bar above the prompt shows the editing keys.
func (c *console) readLine(prompt string) (string, error) {
	return c.readLineKeys(prompt, editBar)
}

// readLineKeys is readLine with the key bar's entries chosen by the
// caller (nil for none).
func (c *console) readLineKeys(prompt string, keys []barKey) (string, error) {
	if c.pending != nil {
		s := *c.pending
		c.pending = nil
		fmt.Fprintln(c.out, prompt+s) // show the answer where it is used
		return s, nil
	}
	if c.closed {
		return "", io.EOF
	}
	bar := ""
	if prompt != "" {
		bar = c.barLine(keys)
	}
	if bar != "" {
		prompt = bar + "\n" + prompt
	}
	var s string
	var err error
	c.inputRows = 0
	if c.raw {
		s, err = c.readLineRaw(prompt)
	} else {
		s, err = c.readLineCooked(prompt)
	}
	if errors.Is(err, io.EOF) {
		c.closed = true
	}
	if bar != "" && err == nil && c.inputRows > 0 {
		c.eraseBar(strings.Count(prompt, "\n") + c.inputRows)
	}
	return s, err
}

// readLineCooked reads a whole line at once, as the terminal delivers it.
// A line containing Ctrl+L (typed, then Enter) clears the screen and asks
// again.
func (c *console) readLineCooked(prompt string) (string, error) {
	for {
		fmt.Fprint(c.out, prompt)
		line, err := c.in.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && line != "" {
				return sanitize(line), nil // final unterminated line
			}
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(c.out)
			}
			return "", err
		}
		if strings.ContainsRune(line, keyCtrlL) {
			c.clearScreen()
			continue
		}
		c.inputRows = 1 // the terminal echoed the line: count its rows
		if cols := c.termCols(); cols > 0 {
			if n := visibleLen(lastLine(prompt) + strings.TrimRight(line, "\r\n")); n > 0 {
				c.inputRows = (n-1)/cols + 1
			}
		}
		return sanitize(line), nil
	}
}

// promptText asks for free text. ":q" cancels. Required fields reject blank
// input; over-long input is rejected rather than silently truncated.
func (c *console) promptText(label string, maxLen int, required bool) (string, error) {
	for {
		s, err := c.readLine(promptLabel(label, ":q cancel"))
		if err != nil {
			return "", err
		}
		switch {
		case strings.EqualFold(s, ":q"):
			return "", errCancel
		case s == "" && required:
			c.warn("This field is required.")
		case utf8.RuneCountInString(s) > maxLen:
			c.warn("Too long (%d chars, max %d).", utf8.RuneCountInString(s), maxLen)
		default:
			return s, nil
		}
	}
}

// promptInt asks for an integer accepted by valid. "q" cancels.
func (c *console) promptInt(label string, valid func(int) error) (int, error) {
	for {
		s, err := c.readLine(promptLabel(label, "q cancel"))
		if err != nil {
			return 0, err
		}
		if isCancel(s) {
			return 0, errCancel
		}
		n, convErr := strconv.Atoi(strings.TrimPrefix(s, "#"))
		if convErr != nil {
			c.warn("%q is not a whole number.", s)
			continue
		}
		if vErr := valid(n); vErr != nil {
			c.warn("%v", vErr)
			continue
		}
		return n, nil
	}
}

// promptChoice shows a numbered list and returns the chosen zero-based index.
func (c *console) promptChoice(label string, options []string) (int, error) {
	// Numbers are padded to the widest ("[9] " lines up with "[10]"). A
	// long option wraps to a second line under its own text; on narrow
	// screens, where that would double the menu's height, it is cut.
	numW := len(fmt.Sprintf("[%d]", len(options)))
	for i, opt := range options {
		num := fmt.Sprintf("[%d]", i+1)
		pad := strings.Repeat(" ", numW-len(num))
		room := c.width() - 5 - numW
		first := true
		for _, part := range strings.Split(opt, "\n") { // an option may span lines
			part = strings.TrimLeft(part, " ")
			lines := []string{part}
			if visibleLen(part) > room {
				lines = []string{truncate(stripANSI(part), room)}
				if c.width() >= 60 {
					lines = wrap(stripANSI(part), room, "")
					if len(lines) > 2 {
						lines = []string{lines[0], truncate(strings.Join(lines[1:], " "), room)}
					}
				}
			}
			for _, l := range lines {
				if first {
					fmt.Fprintf(c.out, "    %s%s %s\n", sty.Cyan(num), pad, l)
					first = false
				} else {
					fmt.Fprintf(c.out, "    %s %s\n", strings.Repeat(" ", numW), l)
				}
			}
		}
	}
	n, err := c.promptInt(label, func(n int) error {
		if n < 1 || n > len(options) {
			return fmt.Errorf("choose a number between 1 and %d", len(options))
		}
		return nil
	})
	return n - 1, err
}

// parseHours accepts decimal hours ("1.5") or a Go duration ("1h30m", "45m").
func parseHours(s string) (float64, error) {
	if h, err := strconv.ParseFloat(s, 64); err == nil {
		if math.IsNaN(h) || math.IsInf(h, 0) {
			return 0, errors.New("not a finite number")
		}
		return h, nil
	}
	d, err := time.ParseDuration(strings.ToLower(strings.ReplaceAll(s, " ", "")))
	if err != nil {
		return 0, fmt.Errorf("%q is not a number of hours (try 2.5 or 1h30m)", s)
	}
	return d.Hours(), nil
}

// promptHours asks for an hour amount up to max. When allowZero is true a
// blank answer means zero; otherwise the value must be strictly positive.
func (c *console) promptHours(label, hint string, max float64, allowZero bool) (float64, error) {
	for {
		s, err := c.readLine(promptLabel(label, hint+", q cancel"))
		if err != nil {
			return 0, err
		}
		if isCancel(s) {
			return 0, errCancel
		}
		if s == "" && allowZero {
			return 0, nil
		}
		h, perr := parseHours(s)
		switch {
		case perr != nil:
			c.warn("%v", perr)
		case h < 0:
			c.warn("Hours cannot be negative.")
		case h == 0 && !allowZero:
			c.warn("Hours must be greater than zero.")
		case h > max:
			c.warn("Maximum is %.0f hours per entry.", max)
		default:
			return roundHours(h), nil
		}
	}
}

// confirm asks a yes/no question and loops until it gets one.
func (c *console) confirm(label string) (bool, error) {
	for {
		s, err := c.readLine(promptLabel(label, "y/n"))
		if err != nil {
			return false, err
		}
		switch strings.ToLower(s) {
		case "y", "yes":
			return true, nil
		case "n", "no", "q":
			return false, nil
		}
		c.warn("Please answer y or n.")
	}
}

// eraseLastLine removes the line just answered (the cursor is at the
// start of the line below it), as man and less do with their prompt, so
// the next output takes its place. Only on a real terminal.
func (c *console) eraseLastLine() {
	if c.screen && os.Getenv("TERM") != "dumb" {
		fmt.Fprint(c.out, "\x1b[1A\r\x1b[2K")
	}
}

// clearForRedraw clears a terminal before a screen is drawn again from
// its start (the pager going back or jumping to a match).
func (c *console) clearForRedraw() {
	if c.screen && os.Getenv("TERM") != "dumb" {
		fmt.Fprint(c.out, clearSeq)
	}
}

// typeAhead makes s the answer to the next prompt.
func (c *console) typeAhead(s string) { c.pending = &s }
