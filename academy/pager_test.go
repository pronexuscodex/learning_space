package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// A choice typed at a pager pause is not lost: paging stops and the typed
// answer goes to the next prompt, so a long list need not be paged to the
// end before picking from it.
func TestPagerTypeAhead(t *testing.T) {
	t.Setenv("LINES", "12") // pages of 10 lines
	var out bytes.Buffer
	a := &App{
		reg:   seedRegistry(),
		con:   &console{in: bufio.NewReader(strings.NewReader("3\n")), out: &out},
		width: 80,
		pager: true,
	}
	a.paged(false, func() {
		for i := 1; i <= 40; i++ {
			a.println(fmt.Sprintf("line %d", i))
		}
	})
	if strings.Contains(out.String(), "line 11") {
		t.Fatal("paging must stop once a choice is typed")
	}
	n, err := a.con.promptChoice("Pick", []string{"a", "b", "c", "d"})
	if err != nil || n != 2 {
		t.Fatalf("the typed 3 must answer the next prompt: got %d, %v", n, err)
	}

	// q still just stops paging, and Enter still pages on.
	out.Reset()
	a.con = &console{in: bufio.NewReader(strings.NewReader("\nq\n")), out: &out}
	a.paged(false, func() {
		for i := 1; i <= 40; i++ {
			a.println(fmt.Sprintf("line %d", i))
		}
	})
	if !strings.Contains(out.String(), "line 20") || strings.Contains(out.String(), "line 21") || a.con.pending != nil {
		t.Fatalf("Enter pages on and q stops without typing ahead:\n%s", out.String())
	}
}

// Like man and less, an answered pager prompt is erased on a terminal, so
// "Enter for more" never stays in the middle of the text; with output
// going to a file or pipe, nothing extra is written.
func TestPagerErasesAnsweredPrompt(t *testing.T) {
	t.Setenv("LINES", "12")
	t.Setenv("TERM", "xterm-256color")
	for _, screen := range []bool{true, false} {
		var out bytes.Buffer
		a := &App{
			reg:   seedRegistry(),
			con:   &console{in: bufio.NewReader(strings.NewReader("\n\n\n\n")), out: &out, screen: screen},
			width: 80,
			pager: true,
		}
		a.paged(true, func() {
			for i := 1; i <= 25; i++ {
				a.println(fmt.Sprintf("line %d", i))
			}
		})
		const erase = "\x1b[1A\r\x1b[2K"
		got := strings.Count(out.String(), erase)
		if want := map[bool]int{true: 3, false: 0}[screen]; got != want {
			t.Errorf("screen=%v: %d prompts erased, want %d", screen, got, want)
		}
		if screen && !strings.Contains(out.String(), "more · ") && !strings.Contains(out.String(), "for more") {
			t.Error("the prompt must still be shown before it is answered")
		}
	}
}

// pagerRun pages 40 numbered lines (pages of 10) with the given input and
// returns what was shown and the app, for checking type-ahead.
func pagerRun(t *testing.T, input string, finalPause bool) (string, *App) {
	t.Helper()
	t.Setenv("LINES", "12")
	var out bytes.Buffer
	a := &App{
		reg:   seedRegistry(),
		con:   &console{in: bufio.NewReader(strings.NewReader(input)), out: &out},
		width: 100,
		pager: true,
	}
	a.paged(finalPause, func() {
		for i := 1; i <= 40; i++ {
			a.println(fmt.Sprintf("line %d", i))
		}
	})
	return out.String(), a
}

func TestPagerKeys(t *testing.T) {
	// b goes back a page; on the first page it is left for what comes next.
	out, a := pagerRun(t, "\nb\nq\n", false)
	if strings.Count(out, "line 1\n") != 2 || a.con.pending != nil {
		t.Errorf("b must show the first page again:\n%s", out)
	}
	_, a = pagerRun(t, "b\n", false)
	if a.con.pending == nil || *a.con.pending != "b" {
		t.Error("b on the first page has nowhere to go back to, so it answers the next prompt")
	}

	// > jumps to the end, < back to the start.
	out, _ = pagerRun(t, ">\n<\nq\n", true)
	if !strings.Contains(out, "line 40") || strings.Contains(out, "line 21\n") || strings.Count(out, "line 1\n") != 2 {
		t.Errorf("> must show the last page and < the first, without the pages between:\n%s", out)
	}
	if !strings.Contains(out, "end") {
		t.Error("the last page's prompt must say it is the end")
	}

	// /word finds the next line containing it, / alone the one after,
	// wrapping to the first match at the end.
	out, _ = pagerRun(t, "/LINE 3\n"+strings.Repeat("/\n", 11)+"q\n", false)
	for _, want := range []string{"“LINE 3” 1/11", "“LINE 3” 11/11", "Back to the first match."} {
		if !strings.Contains(out, want) {
			t.Errorf("search must show %q:\n%s", want, out)
		}
	}
	out, a = pagerRun(t, "/nothing like this\nq\n", false)
	if !strings.Contains(out, "is not on this screen") || a.con.pending != nil {
		t.Errorf("a word that is not there must say so:\n%s", out)
	}

	// / alone with no search yet is the main menu's Search key.
	_, a = pagerRun(t, "/\n", false)
	if a.con.pending == nil || *a.con.pending != "/" {
		t.Error("/ alone must answer the next prompt when nothing has been searched")
	}
}

func TestPagerPromptFits(t *testing.T) {
	for _, cols := range []int{40, 60, 80, 130} {
		a := &App{width: cols, con: &console{out: &bytes.Buffer{}}}
		for _, end := range []int{10, 40} {
			for _, f := range []pagerSearch{{current: -1}, {query: "stack", matches: []int{1, 5}, current: 0}} {
				if p := a.pagerPrompt(0, end, 40, &f); visibleLen(p) >= cols {
					t.Errorf("%d columns: prompt is %d wide: %q", cols, visibleLen(p), stripANSI(p))
				}
			}
		}
	}
}

func TestHighlightKeepsColours(t *testing.T) {
	old := sty
	sty = Style{on: true}
	defer func() { sty = old }()
	line := sty.Cyan("┃") + " The Stack and the " + sty.Bold("stack frame")
	got := highlightMatches(line, "stack")
	if stripANSI(got) != stripANSI(line) {
		t.Fatalf("highlighting must not change the text: %q", stripANSI(got))
	}
	if strings.Count(got, "\x1b[7m") != 2 || strings.Count(got, "\x1b[27m") != 2 {
		t.Errorf("both matches, whatever their case, must be highlighted: %q", got)
	}
	if !strings.Contains(got, "\x1b[36m┃") {
		t.Error("the line's own colours must be kept")
	}
	if highlightMatches(line, "queue") != line {
		t.Error("a line without the word must be unchanged")
	}
}
