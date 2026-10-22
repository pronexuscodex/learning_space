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

// pagerPromptWords are the words of the old "Enter more · b back · /word
// find · q stop · or type your choice" line, in any of its forms. None of
// them may be shown between pages any more.
var pagerPromptWords = []string{"Enter more", "Enter for more", "Enter to continue", "Enter continue", "q stop", "or type your choice", "/word find", "b back"}

func assertNoPagerPrompt(t *testing.T, out string) {
	t.Helper()
	plain := stripANSI(out)
	for _, w := range pagerPromptWords {
		if strings.Contains(plain, w) {
			t.Errorf("a pagination prompt (%q) was shown:\n%s", w, plain)
		}
	}
}

// Enter shows the next page at once, with no prompt or key list printed
// under a page. On a terminal the line the reader pressed Enter on is
// erased, so each page follows the last seamlessly; with output going to
// a file or pipe, nothing extra is written.
func TestPagerEnterAdvancesWithoutPrompt(t *testing.T) {
	t.Setenv("LINES", "12") // pages of 10 lines
	t.Setenv("TERM", "xterm-256color")
	const erase = "\x1b[1A\r\x1b[2K"
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
		got := out.String()
		assertNoPagerPrompt(t, got)
		if !strings.Contains(got, "line 25") || a.con.pending != nil {
			t.Errorf("screen=%v: three Enters must show all 25 lines and type nothing ahead:\n%s", screen, got)
		}
		if want := map[bool]int{true: 3, false: 0}[screen]; strings.Count(got, erase) != want {
			t.Errorf("screen=%v: %d answered lines erased, want %d", screen, strings.Count(got, erase), want)
		}
		// Nothing at all is printed between a page and the next one.
		between := map[bool]string{true: "line 10\n" + erase + "line 11\n", false: "line 10\nline 11\n"}[screen]
		if !strings.Contains(got, between) {
			t.Errorf("screen=%v: the next page must follow the last line directly (%q):\n%q", screen, between, got)
		}
	}
}

// A real lesson: Enter pages through a concept card without any prompt,
// and the questions that follow the card are still asked and answered as
// before; an Enter that turned a page is never taken as their answer.
func TestLessonEnterAdvancesThenQuestionsStillWork(t *testing.T) {
	t.Setenv("LINES", "12")
	t.Setenv("TERM", "xterm-256color")
	g, _ := guideFor(1)

	// The card as one block, to count its pages (10 lines each).
	var whole bytes.Buffer
	dry := &App{reg: seedRegistry(), width: 80, con: &console{in: bufio.NewReader(strings.NewReader("n\n")), out: &whole}}
	if err := dry.studyConceptAt(1, g, 0); err != nil {
		t.Fatal(err)
	}
	card, _, ok := strings.Cut(whole.String(), "  "+sty.Bold("Do you understand it"))
	if !ok {
		t.Fatalf("the card must end with the understanding question:\n%s", whole.String())
	}
	cardLines := strings.Split(strings.TrimRight(card, "\n"), "\n")
	enters := (len(cardLines)+9)/10 - 1
	if enters < 2 {
		t.Fatalf("the card should span several pages, got %d lines", len(cardLines))
	}
	var lastLine string // the card's last line of text (not its border)
	for i := len(cardLines) - 1; i >= 0 && lastLine == ""; i-- {
		if l := strings.Trim(stripANSI(cardLines[i]), " ┃┗━"); l != "" {
			lastLine = l
		}
	}

	var out bytes.Buffer
	input := strings.Repeat("\n", enters) + "y\n" + "Code becomes machine instructions\n"
	a := &App{reg: seedRegistry(), width: 80, pager: true,
		con: &console{in: bufio.NewReader(strings.NewReader(input)), out: &out, screen: true}}
	if err := a.studyConceptAt(1, g, 0); err != nil {
		t.Fatal(err)
	}
	shown := stripANSI(out.String())
	assertNoPagerPrompt(t, out.String())
	if !strings.Contains(shown, lastLine) {
		t.Errorf("Enter must page through to the card's last line (%q):\n%s", lastLine, shown)
	}
	if strings.Contains(shown, "you read") {
		t.Error("paging with Enter reads the card to the end; it must not count as leaving early")
	}
	if !strings.Contains(shown, "Do you understand it") || !strings.Contains(shown, "Your explanation") {
		t.Errorf("the card's questions must still be asked:\n%s", shown)
	}
	_, s := a.reg.findStage(1)
	if !s.hasStudied(g.Concepts[0].Name) || s.Notes[g.Concepts[0].Name] != "Code becomes machine instructions" {
		t.Error("the answers typed after paging must be taken as before")
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
	assertNoPagerPrompt(t, out)

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
