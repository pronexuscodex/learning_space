package main

// A minimal pager. Long screens (the ledger, Start Here, glossaries,
// concept cards) are far taller than a terminal, and the menu that follows
// them would otherwise push them straight off the screen. When a person
// is typing at a real terminal, such screens are shown one page at a time,
// and Enter shows the next page; no prompt is printed between pages.

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
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

// paged renders a screen and shows it one page at a time: Enter shows the
// next page. No prompt or key list is printed under a page; the reader
// just presses Enter to read on. The same keys as before still work when
// typed: b to go back, < and > for the start and end, /word to find a
// word (/ alone finds the next match, highlighted), q to stop. Anything
// else typed at a pause answers whatever comes next (a number from the
// list, a menu key), so a long list need not be paged to the end. With
// finalPause, it also waits after the last page, so the screen stays
// visible until the learner presses Enter.
func (a *App) paged(finalPause bool, render func()) {
	a.pagedWith(pageOpts{finalPause: finalPause}, render)
}

// pageOpts adjust how pagedWith shows a screen.
type pageOpts struct {
	finalPause bool // wait after the last page
	home       bool // opened from the main menu: its keys leave the page and go there at once
	recap      bool // the caller sums up when the reader leaves early: do not reprint the view
	start      int  // the line to open at
}

// pageResult says how the reader left a paged screen.
type pageResult struct {
	early bool // left before reaching the end
	top   int  // the first line on screen when they left
	end   int  // the line after the last one on screen
	read  int  // how far they got, in percent
	typed bool // left by typing an answer to what comes next
}

// pagedWith is paged with options, reporting how the reader left.
func (a *App) pagedWith(o pageOpts, render func()) pageResult {
	if !a.pager {
		render()
		return pageResult{read: 100}
	}
	var buf bytes.Buffer
	out := a.con.out
	a.con.out = &buf
	render()
	a.con.out = out

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	// Like man: a long text (or one to read before going back to the
	// menu) gets the whole screen, when keys can be read one at a time.
	if a.fullScreenOK() && (o.finalPause || len(lines) > termHeight()-2) {
		if res, ok := a.pageFullScreen(lines, o, render); ok {
			return res
		}
	}
	finalPause := o.finalPause
	pageSize := max(5, termHeight()-2)
	top := min(max(0, o.start), max(0, len(lines)-pageSize))
	result := func(end int, typed bool) pageResult {
		return pageResult{early: end < len(lines), top: top, end: end, read: end * 100 / max(1, len(lines)), typed: typed}
	}
	var find pagerSearch
	draw, clear := true, false
	after := "" // a note shown under the page, once
	for {
		end := min(top+pageSize, len(lines))
		if draw {
			if clear && a.con.screen {
				a.con.clearForRedraw()
			}
			page := lines[top:end]
			if find.query != "" {
				page = make([]string, end-top)
				for i, l := range lines[top:end] {
					page[i] = highlightMatches(l, find.query)
				}
			}
			a.println(strings.Join(page, "\n"))
			if after != "" {
				a.con.note(after)
			}
		}
		draw, clear, after = true, false, ""
		last := end == len(lines)
		if last && !finalPause && find.query == "" {
			return result(end, false)
		}

		// Wait for the reader without printing any prompt: the page simply
		// stops, and Enter shows the next one.
		s, err := a.con.readLine("")
		if err != nil {
			return result(end, false)
		}
		// The line the reader answered on is erased, so the next page
		// continues the text seamlessly.
		a.con.eraseLastLine()
		switch {
		case isCancel(s):
			return result(end, false)
		case s == "":
			if last {
				return result(end, false)
			}
			top = end
		case (s == "b" || s == "B") && top > 0:
			top, clear = max(0, top-pageSize), true
		case s == "<":
			top, clear = 0, true
		case s == ">":
			top, clear = max(0, len(lines)-pageSize), true
		case strings.HasPrefix(s, "/") && (len(s) > 1 || find.query != ""):
			if len(s) > 1 {
				find = newPagerSearch(lines, s[1:])
			}
			at, note := find.next(top)
			if at < 0 {
				a.println(sty.Yellow("  " + note))
				draw = false
				continue
			}
			// The match, with a little context above, and where it is
			// among the matches, since no status line shows that.
			where := fmt.Sprintf("“%s” %d/%d", truncate(find.query, 16), find.current+1, len(find.matches))
			if note != "" {
				where += " · " + note
			}
			top, clear, after = max(0, at-2), true, where
		default:
			// Anything else is the answer to what comes next: no need to
			// page to the end first.
			a.con.typeAhead(s)
			return result(end, true)
		}
	}
}

// pagerSearch is a search within the paged text.
type pagerSearch struct {
	query   string
	matches []int // line numbers that contain the query
	current int   // index into matches of the last one shown, -1 before the first
}

func newPagerSearch(lines []string, query string) pagerSearch {
	f := pagerSearch{query: query, current: -1}
	q := []rune(strings.ToLower(query))
	for i, l := range lines {
		if len(matchSpans(stripANSI(l), q)) > 0 {
			f.matches = append(f.matches, i)
		}
	}
	return f
}

// next moves to the next match at or after line from (after the current
// one, when there is one), wrapping to the top. It returns the line, or -1
// with a message when there is nothing to show.
func (f *pagerSearch) next(from int) (int, string) {
	if len(f.matches) == 0 {
		return -1, fmt.Sprintf("“%s” is not on this screen", f.query)
	}
	start := from
	if f.current >= 0 {
		start = f.matches[f.current] + 1
	}
	for i, m := range f.matches {
		if m >= start {
			f.current = i
			return m, ""
		}
	}
	f.current = 0
	return f.matches[0], "Back to the first match."
}

// prev moves to the match before the current one (or before line from),
// wrapping to the last.
func (f *pagerSearch) prev(from int) (int, string) {
	if len(f.matches) == 0 {
		return -1, fmt.Sprintf("“%s” is not on this screen", f.query)
	}
	before := from
	if f.current >= 0 {
		before = f.matches[f.current]
	}
	for i := len(f.matches) - 1; i >= 0; i-- {
		if f.matches[i] < before {
			f.current = i
			return f.matches[i], ""
		}
	}
	f.current = len(f.matches) - 1
	return f.matches[f.current], "Back to the last match."
}

// matchSpans returns the rune ranges of plain where q (lower case) occurs,
// ignoring case.
func matchSpans(plain string, q []rune) [][2]int {
	if len(q) == 0 {
		return nil
	}
	r := []rune(plain)
	var spans [][2]int
	for i := 0; i+len(q) <= len(r); {
		ok := true
		for j, c := range q {
			if unicode.ToLower(r[i+j]) != c {
				ok = false
				break
			}
		}
		if ok {
			spans = append(spans, [2]int{i, i + len(q)})
			i += len(q)
		} else {
			i++
		}
	}
	return spans
}

// highlightMatches shows every occurrence of query in a (possibly styled)
// line in reverse video, keeping the line's own colours.
func highlightMatches(line, query string) string {
	if !sty.on {
		return line
	}
	spans := matchSpans(stripANSI(line), []rune(strings.ToLower(query)))
	if len(spans) == 0 {
		return line
	}
	var b strings.Builder
	vis, next := 0, 0 // visible rune index; next span
	for i := 0; i < len(line); {
		if line[i] == '\x1b' { // copy an escape sequence untouched
			j := i + 1
			if j < len(line) && line[j] == '[' {
				j++
				for j < len(line) && (line[j] < 0x40 || line[j] > 0x7e) {
					j++
				}
				j++
			}
			j = min(j, len(line))
			b.WriteString(line[i:j])
			i = j
			continue
		}
		if next < len(spans) && vis == spans[next][0] {
			b.WriteString("\x1b[7m")
		}
		_, size := utf8.DecodeRuneInString(line[i:])
		b.WriteString(line[i : i+size])
		i += size
		vis++
		if next < len(spans) && vis == spans[next][1] {
			b.WriteString("\x1b[27m")
			next++
		}
	}
	return b.String()
}
