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
