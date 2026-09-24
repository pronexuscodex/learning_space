package main

// Screens for the programmer's dictionary.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// wordOfTheDayLine is the dashboard's word of the day.
func (a *App) wordOfTheDayLine() string {
	w := wordOfTheDay(time.Now())
	width := a.cols()
	label := "🔤 Word of the day: "
	if width < 60 {
		label = "🔤 Today: "
	}
	tail := sty.Gray("  (d)")
	room := width - 2 - visibleLen(label) - visibleLen(tail)
	head := sty.Cyan(label) + sty.Bold(truncate(w.Word, room))
	if mean := room - visibleLen(w.Word) - 3; mean >= 12 {
		head += sty.Gray(" — " + truncate(w.Means, mean))
	}
	return "  " + head + tail
}

// renderWord prints one dictionary entry as a card.
func (a *App) renderWord(w Word, inDeck bool) {
	width := a.cols()
	edge := sty.Cyan("┃")
	a.println("")
	a.println("  " + sty.Cyan("┏━ ") + sty.Bold(sty.Cyan(truncate(w.Word, width-5))))
	meta := w.Cat
	if w.Stage > 0 {
		meta += fmt.Sprintf(" · taught in Stage %d", w.Stage)
	}
	for _, l := range wrap(meta, width-6, "") {
		a.printf("  %s %s\n", edge, sty.Gray(l))
	}
	if len(w.Also) > 0 {
		for _, l := range wrap("also: "+strings.Join(w.Also, ", "), width-6, "") {
			a.printf("  %s %s\n", edge, sty.Gray(l))
		}
	}
	section := func(title string, color func(string) string, text string) {
		a.printf("  %s\n  %s %s\n", edge, edge, sty.Bold(color(title)))
		for _, l := range wrap(text, width-6, "") {
			a.printf("  %s %s\n", edge, l)
		}
	}
	section("What it means", sty.Cyan, w.Means)
	section("Think of it like", sty.Yellow, w.Like)
	a.printf("  %s\n  %s %s\n", edge, edge, sty.Bold(sty.Green("In code")))
	for _, line := range strings.Split(w.Example, "\n") {
		for _, part := range hardBreak(line, width-8) {
			a.printf("  %s   %s\n", edge, sty.Green(part))
		}
	}
	section("⚠ Watch out", sty.Red, w.Watch)
	if len(w.See) > 0 {
		a.printf("  %s\n", edge)
		var items []string
		for i, s := range w.See {
			items = append(items, sty.Cyan(fmt.Sprintf("[%d]", i+1))+" "+s)
		}
		lines := flow(items, "  ", width-4-12, "")
		for i, l := range lines {
			lead := "            "
			if i == 0 {
				lead = sty.Bold("Related") + "     "
			}
			a.printf("  %s %s%s\n", edge, lead, l)
		}
	}
	deck := sty.Gray("not in your review deck")
	if inDeck {
		deck = sty.Green("✓ in your review deck")
	}
	a.printf("  %s\n  %s %s\n", edge, edge, deck)
	a.println("  " + sty.Cyan("┗"+strings.Repeat("━", width-4)))
}

// showWord displays an entry and lets the learner follow related words
// and add or remove it from the review deck.
func (a *App) showWord(w Word) error {
	for {
		a.mu.Lock()
		in := a.reg.inDeck(w.Word)
		a.mu.Unlock()
		a.paged(false, func() { a.renderWord(w, in) })
		action := "r add to review deck"
		if in {
			action = "r remove from deck"
		}
		hint := action + " · Enter back"
		if len(w.See) > 0 {
			hint = fmt.Sprintf("1-%d related · ", len(w.See)) + hint
		}
		s, err := a.con.readLine(promptLabel("Word", hint))
		if err != nil {
			return err
		}
		s = strings.ToLower(strings.TrimSpace(s))
		switch {
		case s == "" || isCancel(s):
			return nil
		case s == "r":
			key := strings.ToLower(w.Word)
			a.mutate(func(r *Registry) { r.WordDeck = setMember(r.WordDeck, key, !in) })
			if in {
				a.con.ok("Removed %s from your review deck.", w.Word)
			} else {
				a.con.ok("Added %s: it will come up in the Daily Review [9], spaced out over time.", w.Word)
			}
		default:
			n, convErr := strconv.Atoi(s)
			if convErr != nil || n < 1 || n > len(w.See) {
				a.con.warn("Choose a related word by number, r, or press Enter.")
				continue
			}
			next, ok := lookupWord(w.See[n-1])
			if !ok {
				continue
			}
			w = next
		}
	}
}

// pickWord lists entries and returns the chosen one.
func (a *App) pickWord(words []Word, label string) (Word, bool, error) {
	a.paged(false, func() {
		width := a.cols()
		for i, w := range words {
			num := sty.Cyan(fmt.Sprintf("%3d.", i+1))
			room := width - 7 - visibleLen(w.Word) - 3
			mean := ""
			if room > 12 {
				mean = sty.Gray(" — " + truncate(w.Means, room))
			}
			a.printf("  %s %s%s\n", num, sty.Bold(w.Word), mean)
		}
	})
	a.println("")
	n, err := a.con.promptInt(label, func(n int) error {
		if n < 1 || n > len(words) {
			return fmt.Errorf("choose 1 to %d", len(words))
		}
		return nil
	})
	if err != nil {
		return Word{}, false, err
	}
	return words[n-1], true, nil
}

// dictionary is the main dictionary screen.
func (a *App) dictionary() error {
	a.println(heading("PROGRAMMER'S DICTIONARY · every word explained", sty.Cyan, a.cols()))
	a.println("")
	for _, l := range wrap(fmt.Sprintf("%d words programmers use without explaining them, from \"variable\" and \"type casting\" to \"segfault\", \"race condition\" and \"yak shaving\". Each one has a plain meaning, an everyday comparison, real code, and the mistake people usually make with it.", len(vocab)), a.cols()-4, "  ") {
		a.println(l)
	}
	for {
		a.println("")
		a.println(a.wordOfTheDayLine())
		s, err := a.con.readLine(promptLabel("Type a word to look up · c categories · a A–Z · t today's word · m my deck", "q back"))
		if err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		var chosen Word
		var ok bool
		switch strings.ToLower(s) {
		case "", "q", ":q", "b":
			return nil
		case "t":
			chosen, ok = wordOfTheDay(time.Now()), true
		case "a":
			chosen, ok, err = a.pickWord(vocab, "Open which word?")
		case "c":
			for i, c := range vocabCats {
				a.printf("  %s %s %s\n", sty.Cyan(fmt.Sprintf("[%d]", i+1)), c, sty.Gray(fmt.Sprintf("(%d)", len(wordsByCat(c)))))
			}
			n, perr := a.con.promptInt("Category", func(n int) error {
				if n < 1 || n > len(vocabCats) {
					return fmt.Errorf("choose 1 to %d", len(vocabCats))
				}
				return nil
			})
			if perr != nil {
				err = perr
				break
			}
			a.println(heading(strings.ToUpper(vocabCats[n-1]), sty.Cyan, a.cols()))
			chosen, ok, err = a.pickWord(wordsByCat(vocabCats[n-1]), "Open which word?")
		case "m":
			a.mu.Lock()
			var deck []Word
			for _, name := range a.reg.WordDeck {
				if w, found := lookupWord(name); found {
					deck = append(deck, w)
				}
			}
			a.mu.Unlock()
			if len(deck) == 0 {
				a.con.note("Your deck has no words yet. Open a word and press r to add it; it will then come up in the Daily Review.")
				continue
			}
			chosen, ok, err = a.pickWord(deck, "Open which word?")
		default:
			hits := searchWords(s)
			switch {
			case len(hits) == 0:
				a.con.note("No entry for %q yet. Try a shorter form (\"cast\" instead of \"casting\"), or use search [/] for the whole curriculum.", s)
				continue
			case len(hits) == 1 || strings.EqualFold(hits[0].Word, s):
				chosen, ok = hits[0], true
			default:
				if len(hits) > 15 {
					hits = hits[:15]
				}
				chosen, ok, err = a.pickWord(hits, "Open which word?")
			}
		}
		if err != nil {
			if errors.Is(err, errCancel) {
				continue
			}
			return err
		}
		if ok {
			if err := a.showWord(chosen); err != nil {
				return err
			}
		}
	}
}
