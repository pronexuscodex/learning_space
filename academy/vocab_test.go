package main

import (
	"strings"
	"testing"
	"time"
)

func TestVocabularyIsComplete(t *testing.T) {
	if len(vocab) < 200 {
		t.Fatalf("expected at least 200 words, got %d", len(vocab))
	}
	cats := map[string]bool{}
	for _, c := range vocabCats {
		cats[c] = true
	}
	names := map[string]string{}
	for _, w := range vocab {
		if w.Word == "" || w.Means == "" || w.Like == "" || w.Example == "" || w.Watch == "" {
			t.Errorf("%q is missing a field", w.Word)
		}
		if !cats[w.Cat] {
			t.Errorf("%q has unknown category %q", w.Word, w.Cat)
		}
		if w.Stage < NoStage || w.Stage > 16 {
			t.Errorf("%q has stage %d", w.Word, w.Stage)
		}
		for _, n := range append([]string{w.Word}, w.Also...) {
			key := strings.ToLower(n)
			if other, dup := names[key]; dup && other != w.Word {
				t.Errorf("name %q belongs to both %q and %q", n, other, w.Word)
			}
			names[key] = w.Word
		}
	}
	for _, w := range vocab {
		for _, s := range w.See {
			if _, ok := lookupWord(s); !ok {
				t.Errorf("%q links to unknown word %q", w.Word, s)
			}
		}
	}
	for _, c := range vocabCats {
		if len(wordsByCat(c)) == 0 {
			t.Errorf("category %q is empty", c)
		}
	}
}

func TestVocabularyLookupAndSearch(t *testing.T) {
	w, ok := lookupWord("type conversion")
	if !ok || w.Word != "Type casting" {
		t.Fatalf("lookup by alias = %q, %v", w.Word, ok)
	}
	hits := searchWords("cast")
	if len(hits) == 0 || hits[0].Word != "Type casting" {
		t.Fatalf("search 'cast' = %v", hits)
	}
	if got := searchWords("pointer"); len(got) == 0 || got[0].Word != "Pointer" {
		t.Fatal("an exact name should rank first")
	}
	if len(searchWords("zzqqxx")) != 0 || searchWords(" ") != nil {
		t.Fatal("nonsense and blank queries find nothing")
	}
	found := wordsIn(5, nil, "We cast the float to an int and store it in a variable.")
	var got []string
	for _, f := range found {
		got = append(got, f.Word)
	}
	joined := strings.Join(got, ",")
	for _, want := range []string{"Type casting", "Variable", "Integer", "Floating point"} {
		if !strings.Contains(joined, want) {
			t.Errorf("wordsIn missed %s: %v", want, got)
		}
	}
	if len(wordsIn(2, nil, "variable, integer, pointer, recursion")) != 2 {
		t.Error("wordsIn should respect its limit")
	}
	if len(wordsIn(5, map[string]bool{"variable": true}, "a variable")) != 0 {
		t.Error("wordsIn should skip words in skip")
	}
	if len(wordsIn(5, map[string]bool{"type": true}, "each data type")) != 0 {
		t.Error("an entry whose alias is in skip should be skipped")
	}
	if len(wordsIn(5, nil, "we build a list and call it")) != 0 {
		t.Error("everyday words should not be flagged in prose")
	}
	d1 := wordOfTheDay(time.Date(2026, 9, 24, 8, 0, 0, 0, time.Local))
	d1b := wordOfTheDay(time.Date(2026, 9, 24, 22, 0, 0, 0, time.Local))
	d2 := wordOfTheDay(time.Date(2026, 9, 25, 8, 0, 0, 0, time.Local))
	if d1.Word != d1b.Word || d1.Word == d2.Word {
		t.Fatal("word of the day should be stable within a day and change the next")
	}
}

func TestWordDeckCards(t *testing.T) {
	reg := seedRegistry()
	reg.WordDeck = []string{"type casting", "pointer", "no such word"}
	var words int
	for _, c := range reg.unlockedCards() {
		if c.Kind == CardWord {
			words++
			if !strings.HasPrefix(c.ID, "word :: ") || c.A == "" {
				t.Fatalf("bad word card %+v", c)
			}
		}
	}
	if words != 2 {
		t.Fatalf("expected 2 word cards, got %d", words)
	}
	if !reg.inDeck("Pointer") {
		t.Fatal("inDeck should ignore case")
	}
}
