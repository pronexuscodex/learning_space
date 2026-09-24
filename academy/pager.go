package main

// A minimal pager. Long screens (the ledger, Start Here, glossaries,
// concept cards) are far taller than a terminal, and the menu that follows
// them would otherwise push them straight off the screen. When a person
// is typing at a real terminal, such screens are shown one page at a time.

import (
	"bytes"
	"os"
	"strconv"
	"strings"
)

// interactive reports whether stdin is a terminal (not a pipe or file).
func interactive() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// stdoutIsTerminal reports whether output goes to a terminal.
func stdoutIsTerminal() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// termHeight returns the terminal's live height, else $LINES, else 24.
func termHeight() int {
	if _, h, ok := terminalSize(int(os.Stdout.Fd())); ok && h >= 10 {
		return h
	}
	if h, err := strconv.Atoi(os.Getenv("LINES")); err == nil && h >= 10 {
		return h
	}
	return 24
}

// paged renders a screen and shows it one page at a time. With finalPause,
// it also waits after the last page, so the screen stays visible until
// the learner is ready to return to the menu.
func (a *App) paged(finalPause bool, render func()) {
	if !a.pager {
		render()
		return
	}
	var buf bytes.Buffer
	out := a.con.out
	a.con.out = &buf
	render()
	a.con.out = out

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	pageSize := max(5, termHeight()-2)
	for start := 0; start < len(lines); start += pageSize {
		end := min(start+pageSize, len(lines))
		a.println(strings.Join(lines[start:end], "\n"))
		last := end == len(lines)
		if last && !finalPause {
			return
		}
		prompt := sty.Gray("  ── ") + sty.Bold(sty.Cyan("Enter")) + sty.Gray(" for more · ") + sty.Bold(sty.Cyan("q")) + sty.Gray(" to stop ──")
		if last {
			prompt = sty.Gray("  ── end · press ") + sty.Bold(sty.Cyan("Enter")) + sty.Gray(" to continue ──")
		}
		s, err := a.con.readLine(prompt + " ")
		if err != nil || isCancel(s) {
			return
		}
	}
}
