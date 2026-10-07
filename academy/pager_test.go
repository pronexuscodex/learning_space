package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// pagerApp is an App that pages, reading keys from input, with pages of
// eight lines.
func pagerApp(t *testing.T, input string) (*App, *bytes.Buffer) {
	t.Helper()
	t.Setenv("LINES", "10")
	var out bytes.Buffer
	a := &App{
		con:   &console{in: bufio.NewReader(strings.NewReader(input)), out: &out},
		width: 80,
		pager: true,
	}
	return a, &out
}

// longScreen prints n numbered lines.
func longScreen(a *App, n int) func() {
	return func() {
		for i := 1; i <= n; i++ {
			a.printf("line %d\n", i)
		}
	}
}

// A choice typed at the first pager prompt skips the remaining pages and
// is answered by the prompt that follows, without being typed again.
func TestPagerPassesTypedChoiceToNextPrompt(t *testing.T) {
	a, out := pagerApp(t, "7\n")
	a.paged(false, longScreen(a, 40))
	if strings.Contains(out.String(), "line 9\n") {
		t.Fatalf("paging should stop once a choice is typed:\n%s", out)
	}
	n, err := a.con.promptInt("Which?", func(int) error { return nil })
	if err != nil || n != 7 {
		t.Fatalf("next prompt got %d, %v; want 7 typed at the pager", n, err)
	}
}

// Enter still pages through, and q still stops without leaving anything
// behind for the next prompt.
func TestPagerEnterAndQuit(t *testing.T) {
	a, out := pagerApp(t, "\nq\n5\n")
	a.paged(false, longScreen(a, 40))
	if !strings.Contains(out.String(), "line 16\n") || strings.Contains(out.String(), "line 17\n") {
		t.Fatalf("want two pages shown, then a stop:\n%s", out)
	}
	n, err := a.con.promptInt("Which?", func(int) error { return nil })
	if err != nil || n != 5 {
		t.Fatalf("next prompt got %d, %v; want 5 read from input", n, err)
	}
}

// After the final pause, a typed choice goes to the menu that follows.
func TestPagerFinalPauseHandsOnChoice(t *testing.T) {
	a, _ := pagerApp(t, "3\n")
	a.paged(true, longScreen(a, 4))
	s, err := a.con.readLine("academy › ")
	if err != nil || s != "3" {
		t.Fatalf("menu got %q, %v; want 3", s, err)
	}
}

// Picking a word from a long list works straight from the first page.
func TestPickWordFromFirstPage(t *testing.T) {
	words := make([]Word, 30)
	for i := range words {
		words[i] = Word{Word: fmt.Sprintf("word%d", i+1), Means: "something"}
	}
	a, out := pagerApp(t, "3\n")
	w, ok, err := a.pickWord(words, "Which word?")
	if err != nil || !ok || w.Word != "word3" {
		t.Fatalf("got %q, %v, %v; want word3", w.Word, ok, err)
	}
	if strings.Contains(out.String(), "word30") {
		t.Fatalf("the rest of the list should be skipped:\n%s", out)
	}
}

// The hint about typing a choice is shown only where it fits.
func TestPagerPromptFitsWidth(t *testing.T) {
	for _, cols := range []int{40, 60, 80, 130} {
		a := &App{width: cols, con: &console{}}
		for _, last := range []bool{false, true} {
			if w := visibleLen(a.pagerPrompt(last)); w > cols {
				t.Errorf("%d columns: pager prompt is %d wide", cols, w)
			}
		}
	}
	if !strings.Contains(stripANSI((&App{width: 80, con: &console{}}).pagerPrompt(false)), "or type your choice") {
		t.Error("80 columns should mention typing a choice")
	}
	// A narrow terminal (a phone, a split pane) still learns it can type.
	for _, cols := range []int{40, 48} {
		for _, last := range []bool{false, true} {
			if p := stripANSI((&App{width: cols, con: &console{}}).pagerPrompt(last)); !strings.Contains(p, "or type") {
				t.Errorf("%d columns should still mention typing: %q", cols, p)
			}
		}
	}
}
