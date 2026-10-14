package main

// A minimal pager. Long screens (the ledger, Start Here, glossaries,
// concept cards) are far taller than a terminal, and the menu that follows
// them would otherwise push them straight off the screen. When a person
// is typing at a real terminal, such screens are shown one page at a time.

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

// paged renders a screen and shows it one page at a time, like man and
// less: Enter for the next page, b to go back, < and > for the start and
// end, /word to find a word (/ alone finds the next match, highlighted),
// q to stop. Anything else typed at the prompt answers whatever comes
// next (a number from the list, a menu key), so a long list need not be
// paged to the end. With finalPause, it also waits after the last page,
// so the screen stays visible until the learner is ready to go on.
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
	top := 0
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
			return
		}

		s, err := a.con.readLine(a.pagerPrompt(top, end, len(lines), &find) + " ")
		if err != nil {
			return
		}
		// Like man and less: the prompt disappears once answered, so it
		// never sits in the middle of the text.
		a.con.eraseLastLine()
		switch {
		case isCancel(s):
			return
		case s == "":
			if last {
				return
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
			top, clear, after = max(0, at-2), true, note // the match, with a little context above
		default:
			// Anything else is the answer to what comes next: no need to
			// page to the end first.
			a.con.typeAhead(s)
			return
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

// pagerPrompt is the status line: where you are, and the keys. It has a
// long, a medium and a short form, for the width of the terminal.
func (a *App) pagerPrompt(top, end, total int, find *pagerSearch) string {
	k := func(key, what string) string { return sty.Bold(sty.Cyan(key)) + sty.Gray(what) }
	line := func(items ...string) string {
		return sty.Gray("  ── ") + strings.Join(items, sty.Gray(" · ")) + sty.Gray(" ──")
	}
	where := fmt.Sprintf("%d%%", end*100/max(1, total))
	if end == total {
		where = "end"
	}
	if find.current >= 0 && len(find.matches) > 0 {
		where = fmt.Sprintf("“%s” %d/%d · %s", truncate(find.query, 16), find.current+1, len(find.matches), where)
	}
	at := sty.Gray(where)
	search := k("/word", " find")
	if find.current >= 0 && len(find.matches) > 1 {
		search = k("/", " next match")
	}
	var forms []string
	if end == total {
		forms = []string{
			line(at, k("Enter", " to continue"), k("b", " back"), search),
			line(at, k("Enter", " continue"), k("b", " back")),
			line(at, k("Enter", "")),
		}
	} else {
		forms = []string{
			line(at, k("Enter", " more"), k("b", " back"), search, k("q", " stop"), sty.Gray("or type your choice")),
			line(at, k("Enter", " more"), k("b", " back"), search, k("q", " stop")),
			line(at, k("Enter", " more"), k("q", " stop")),
			line(k("Enter", ""), k("q", "")),
		}
	}
	for _, f := range forms {
		if visibleLen(f)+1 < a.cols() {
			return f
		}
	}
	return forms[len(forms)-1]
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
