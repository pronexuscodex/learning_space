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
