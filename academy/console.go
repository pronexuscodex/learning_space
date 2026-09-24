package main

// Sanitised, line-based terminal input over a bufio.Reader. Every prompt
// loops until it gets valid input, the user cancels, or input ends (EOF).

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
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
}

// Feedback lines, one visual vocabulary for the whole app.
func (c *console) ok(format string, args ...any) {
	fmt.Fprintf(c.out, "  %s %s\n", sty.Green("✓"), fmt.Sprintf(format, args...))
}
func (c *console) warn(format string, args ...any) {
	fmt.Fprintf(c.out, "  %s %s\n", sty.Yellow("!"), sty.Yellow(fmt.Sprintf(format, args...)))
}
func (c *console) fail(format string, args ...any) {
	fmt.Fprintf(c.out, "  %s %s\n", sty.Red("✗"), sty.Red(fmt.Sprintf(format, args...)))
}
func (c *console) note(format string, args ...any) {
	fmt.Fprintf(c.out, "  %s\n", sty.Dim("· "+fmt.Sprintf(format, args...)))
}

// promptLabel renders "  Label (hint) › ".
func promptLabel(label, hint string) string {
	s := "  " + sty.Bold(label)
	if hint != "" {
		s += " " + sty.Gray("("+hint+")")
	}
	return s + sty.Cyan(" › ")
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
// when input is exhausted (Ctrl-D / closed pipe).
func (c *console) readLine(prompt string) (string, error) {
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
	return sanitize(line), nil
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
	for i, opt := range options {
		fmt.Fprintf(c.out, "    %s %s\n", sty.Cyan(fmt.Sprintf("[%d]", i+1)), opt)
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
