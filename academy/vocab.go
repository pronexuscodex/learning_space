package main

// The programmer's dictionary: jargon explained for people who have never
// heard it before. Every entry says what the word means in plain words,
// gives an everyday analogy, shows it in code, and warns about the usual
// confusion. Entries live in vocab_*.go, grouped by category.

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

// Word is one dictionary entry.
type Word struct {
	Word    string
	Also    []string // other names and spellings, used for search and detection
	Cat     string
	Stage   int    // the stage that teaches it most (0: everywhere)
	Means   string // plain meaning, one to three sentences
	Like    string // an everyday analogy
	Example string // code or a usage example; may span lines
	Watch   string // the usual confusion or mistake
	See     []string
}

// Dictionary categories, in display order.
var vocabCats = []string{
	"Basics", "Types & data", "Functions & structure", "Memory & C",
	"Tools & workflow", "Systems & concurrency", "Networks & web",
	"Data & databases", "Algorithms", "Security", "AI & ML", "Jargon & culture",
}

// vocab is the whole dictionary, gathered from the category files.
var vocab = func() []Word {
	var all []Word
	for _, part := range [][]Word{vocabBasics, vocabTypes, vocabFunctions, vocabMemory, vocabTools,
		vocabSystems, vocabNetworks, vocabData, vocabAlgorithms, vocabSecurity, vocabAI, vocabCulture} {
		all = append(all, part...)
	}
	sort.SliceStable(all, func(i, j int) bool { return strings.ToLower(all[i].Word) < strings.ToLower(all[j].Word) })
	return all
}()

// lookupWord finds an entry by its word or one of its other names.
func lookupWord(name string) (Word, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, w := range vocab {
		if strings.ToLower(w.Word) == name {
			return w, true
		}
		for _, a := range w.Also {
			if strings.ToLower(a) == name {
				return w, true
			}
		}
	}
	return Word{}, false
}

// searchWords returns entries matching the query: exact names first, then
// names that start with it, then names containing it, then meanings that
// contain every query word.
func searchWords(query string) []Word {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	var exact, prefix, contains, body []Word
	for _, w := range vocab {
		names := append([]string{w.Word}, w.Also...)
		best := 4
		for _, n := range names {
			n = strings.ToLower(n)
			switch {
			case n == q:
				best = min(best, 0)
			case strings.HasPrefix(n, q):
				best = min(best, 1)
			case strings.Contains(n, q):
				best = min(best, 2)
			}
		}
		if best == 4 {
			text := strings.ToLower(w.Means + " " + w.Like)
			all := true
			for _, f := range strings.Fields(q) {
				all = all && strings.Contains(text, f)
			}
			if all {
				best = 3
			}
		}
		switch best {
		case 0:
			exact = append(exact, w)
		case 1:
			prefix = append(prefix, w)
		case 2:
			contains = append(contains, w)
		case 3:
			body = append(body, w)
		}
	}
	return append(append(append(exact, prefix...), contains...), body...)
}

// wordsByCat returns the entries of one category, alphabetically.
func wordsByCat(cat string) []Word {
	var out []Word
	for _, w := range vocab {
		if w.Cat == cat {
			out = append(out, w)
		}
	}
	return out
}

// wordRegex caches one pattern per dictionary name.
var wordRegex = map[string]*regexp.Regexp{}

// everydayNames are dictionary names that are also ordinary English, so
// they are not flagged when they merely appear in prose. They can still be
// looked up and searched.
var everydayNames = map[string]bool{
	"build": true, "call": true, "value": true, "block": true, "test": true, "log": true,
	"branch": true, "merge": true, "commit": true, "stream": true, "path": true, "program": true,
	"library": true, "package": true, "class": true, "object": true, "method": true, "model": true,
	"token": true, "set": true, "list": true, "map": true, "index": true, "hack": true, "deploy": true,
	"release": true, "ship": true, "event": true, "lock": true, "race": true, "port": true,
	"request": true, "response": true, "pipe": true, "container": true, "record": true, "link": true,
	"condition": true, "memory": true, "loop": true, "error": true, "bug": true, "code": true,
	"software": true, "name": true, "type": true, "comment": true, "function": true, "return": true,
	"address": true, "buffer": true, "cache": true, "graph": true, "tree": true, "node": true,
	"root": true, "leaf": true, "edge": true, "key": true, "query": true, "session": true,
	"statement": true, "expression": true, "literal": true, "argument": true, "parameter": true,
	"interface": true, "module": true, "import": true, "prompt": true, "shell": true, "console": true,
	"environment": true, "config": true, "settings": true, "docs": true, "review": true, "free": true,
	"image": true, "character": true, "string": true, "text": true, "body": true, "headers": true,
	"client": true, "server": true, "domain": true, "page": true, "swap": true, "thread": true,
	"threads": true, "process": true, "processes": true, "protocol": true, "backup": true,
	"snapshot": true, "migration": true, "schema": true, "sorting": true, "sort": true, "hash": true,
	"validation": true, "exploit": true, "login": true, "cipher": true, "salt": true, "sandbox": true,
	"training": true, "train": true, "fit": true, "prediction": true, "serving": true, "loss": true,
	"objective": true, "gradient": true, "vector": true, "matrix": true, "hype": true, "idiom": true,
	"legacy": true, "workaround": true, "hacky": true, "records": true, "structure": true,
	"make": true, "black": true, "false": true, "true": true, "none": true, "get": true, "post": true,
	"select": true, "future": true, "promise": true, "patch": true, "source": true, "stack": true,
	"reference": true, "register": true, "remainder": true, "precision": true, "scope": true,
	"global": true, "double": true, "leak": true, "parallel": true, "concurrent": true, "abstract": true,
	"pure": true, "invoke": true, "routine": true, "procedure": true, "editor": true, "production": true,
	"prod": true, "pipeline": true, "overflow": true, "header": true, "dry": true, "kiss": true,
	"raise": true, "throw": true, "rest": true, "testing": true, "weights": true, "biases": true,
	"defect": true, "convert": true, "assign": true, "mod": true, "nit": true, "wip": true, "str": true,
	"var": true, "arg": true, "lib": true, "args": true, "deps": true,
}

// wordsIn returns dictionary entries mentioned in the texts (by any of
// their names), at most limit, skipping the ones in skip.
func wordsIn(limit int, skip map[string]bool, texts ...string) []Word {
	joined := strings.Join(texts, " ")
	var found []Word
	for _, w := range vocab {
		names := append([]string{w.Word}, w.Also...)
		skipped := false
		for _, n := range names {
			skipped = skipped || skip[strings.ToLower(n)]
		}
		if skipped {
			continue
		}
		for _, n := range names {
			if len(n) < 3 || everydayNames[strings.ToLower(n)] { // too short or too common to spot reliably
				continue
			}
			re, ok := wordRegex[n]
			if !ok {
				re = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(n) + `(s|es)?\b`)
				wordRegex[n] = re
			}
			if re.MatchString(joined) {
				found = append(found, w)
				break
			}
		}
		if len(found) == limit {
			break
		}
	}
	return found
}

// wordOfTheDay picks the same word for everyone on a given day, cycling
// through the whole dictionary.
func wordOfTheDay(now time.Time) Word {
	days := int(dayStart(now).Sub(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)).Hours() / 24)
	n := len(vocab)
	return vocab[((days*37)%n+n)%n] // 37 is coprime with the size, so neighbours differ
}

// CardWord is the review-card kind for dictionary words.
const CardWord = "word"

// wordCard is the review card for a dictionary word.
func wordCard(w Word) Card {
	a := w.Means
	if w.Example != "" {
		a += "\n\nExample: " + strings.ReplaceAll(w.Example, "\n", " ⏎ ")
	}
	return Card{ID: "word :: " + strings.ToLower(w.Word), StageID: w.Stage, Concept: w.Word, Kind: CardWord,
		Q: "What does “" + w.Word + "” mean? Say it plainly and give an example.", A: a}
}

// inDeck reports whether a word is in the learner's review deck.
func (r *Registry) inDeck(word string) bool {
	return contains(r.WordDeck, strings.ToLower(word))
}
