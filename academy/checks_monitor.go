package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// Checked exercises for Stage 19 (network defence and monitoring):
// tamper-evident logs, multi-pattern signature matching and two login
// detection rules, in C.
var monitorChecks = []Check{
	{
		ID: "log_chain", Stage: 19, Concept: "Tamper-Evident Logs & Trustworthy Time",
		Title: "A hash-chained log",
		Task:  "./log_chain seal reads log messages, one per line, and prints each as HASH MESSAGE: HASH is the lowercase hex SHA-256 of the previous line's HASH (64 zeros before the first line), a newline, and the message. ./log_chain verify [NEWEST] reads sealed lines and recomputes every hash; if all match, and the last HASH equals NEWEST when it is given, it prints ok N lines. Otherwise it prints, on standard error, the number of the first line that fails (a hash that does not match, or a line that is not 64 lowercase hex digits, a space and a message), and exits with status 1, printing nothing on standard output. Messages can be empty, contain any byte except a newline, and be 100,000 bytes long; there can be 100,000 lines. Any other arguments get a usage message and exit status 2.",
		File:  "log_chain.c",
		Starter: `/* Checked exercise: log_chain
 *   printf 'login ana\nlogout ana\n' | ./log_chain seal
 *   <64 hex digits> login ana
 *   <64 hex digits> logout ana
 * h1 = SHA256("000...000" "\n" "login ana")    (64 zeros, as text)
 * h2 = SHA256(hex(h1) "\n" "logout ana")
 * Because each hash covers the one before, changing, removing or
 * reordering any line breaks every hash after it. Removing lines at
 * the end breaks nothing, which is why verify can take NEWEST, the
 * last hash as published somewhere else.
 * Reuse the SHA-256 from your Stage 17 sha256.c. Lines can be very
 * long: read them with getline.
 */
#define _POSIX_C_SOURCE 200809L
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"seal"}, Stdin: "login ana\nlogout ana\n", Want: chainSeal("login ana\nlogout ana\n")},
			{Name: "empty messages", Args: []string{"seal"}, Stdin: "first\n\nthird\n", Want: chainSeal("first\n\nthird\n"),
				Explain: "An empty line is a message too: hash the previous hash, a newline and nothing."},
			{Name: "an intact log", Args: []string{"verify"}, Gen: generated(func() (string, string) {
				in := chainSeal(chainMessages(50, 81))
				return in, "ok 50 lines\n"
			})},
			{Name: "one changed word", Args: []string{"verify"}, Gen: generated(func() (string, string) {
				s := strings.Split(chainSeal(chainMessages(50, 82)), "\n")
				s[29] += "!"
				return strings.Join(s, "\n"), ""
			}), WantExit: 1, Stderr: true,
				Explain: "Line 30's message was changed and its old hash kept: the recomputed hash no longer matches."},
			{Name: "a deleted line", Args: []string{"verify"}, Gen: generated(func() (string, string) {
				s := strings.Split(chainSeal(chainMessages(50, 83)), "\n")
				return strings.Join(append(s[:20:20], s[21:]...), "\n"), ""
			}), WantExit: 1, Stderr: true,
				Explain: "Line 21 was removed: every line is still well formed, but line 21's hash was computed from the removed line's hash. Hashing each message alone, without the previous hash, cannot see this."},
			{Name: "two lines swapped", Args: []string{"verify"}, Gen: generated(func() (string, string) {
				s := strings.Split(chainSeal(chainMessages(50, 84)), "\n")
				s[10], s[11] = s[11], s[10]
				return strings.Join(s, "\n"), ""
			}), WantExit: 1, Stderr: true},
			{Name: "the end cut off", Args: []string{"verify", chainNewest(chainMessages(50, 85))}, Gen: generated(func() (string, string) {
				s := strings.Split(chainSeal(chainMessages(50, 85)), "\n")
				return strings.Join(s[:45], "\n") + "\n", ""
			}), WantExit: 1, Stderr: true,
				Explain: "The last 5 lines were removed. The chain that is left is perfectly valid; only the newest hash, kept elsewhere, shows it is short."},
			{Name: "the end cut off, with nothing to compare", Args: []string{"verify"}, Gen: generated(func() (string, string) {
				s := strings.Split(chainSeal(chainMessages(50, 85)), "\n")
				return strings.Join(s[:45], "\n") + "\n", "ok 45 lines\n"
			}),
				Explain: "Without NEWEST, a log cut short verifies: this is the chain's limit, and why the newest hash must be published somewhere else."},
			{Name: "upper-case hex", Args: []string{"verify"}, Gen: generated(func() (string, string) {
				return strings.ToUpper(chainSeal("x\n")[:64]) + " x\n", ""
			}), WantExit: 1, Stderr: true,
				Explain: "The format is lowercase hex: accept exactly one spelling, so two different texts never verify as the same log."},
			{Name: "a very long message", Args: []string{"seal"}, Gen: generated(func() (string, string) {
				in := "short\n" + strings.Repeat("m", 100000) + "\nafter\n"
				return in, chainSeal(in)
			}),
				Explain: "A 100,000-byte message: a fixed buffer for fgets splits it in two, and every hash from there on is wrong."},
			{Name: "a hundred thousand lines", Args: []string{"verify"}, Gen: generated(func() (string, string) {
				msgs := chainMessages(100000, 86)
				return chainSeal(msgs), "ok 100000 lines\n"
			})},
			{Name: "no mode", WantExit: 2, Stderr: true},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Keep the previous hash as 64 hex characters; for each line, hash prev + \"\\n\" + message, then write the new hex into prev.",
			"In verify, split each line at its first space after checking that the first 64 characters are 0-9 and a-f.",
			"Count lines as you go: the first failure's number is the one to report.",
		},
	},
	{
		ID: "multi_match", Stage: 19, Concept: "Intrusion Detection: Signatures & Anomalies",
		Title: "Many signatures in one pass",
		Task:  "./multi_match PATTERNS reads signatures from the file PATTERNS, one per line (1 to 64 bytes each, at most 5,000 of them and 40,000 bytes in all), and counts how often each occurs in standard input, overlapping occurrences included: aa occurs 3 times in aaaa. Print COUNT PATTERN for each signature in file order, then total N, the sum of the counts. The text can be 16 MB of any bytes, zeros and newlines included, and must be read in a single pass whatever the number of signatures (Aho-Corasick). A blank line or a line longer than 64 bytes in PATTERNS gets a message on standard error and exit status 1; a missing argument, exit status 2.",
		File:  "multi_match.c",
		Starter: `/* Checked exercise: multi_match (how signature engines search)
 *   PATTERNS:  he / she / his / hers   (one per line)
 *   text:      ushers
 *   output:    1 he
 *              1 she
 *              0 his
 *              1 hers
 *              total 3
 * Aho-Corasick:
 *   1. Build a trie of the patterns: next[state][byte], and a list of
 *      which patterns end at each state.
 *   2. Breadth first from the root, give each state a fail link: the
 *      longest proper suffix of its string that is also in the trie.
 *      Fill missing next[][] entries from the fail state's, so that
 *      matching never backtracks, and let each state also report what
 *      its fail state reports ("he" ends inside "she").
 *   3. Read the text once: state = next[state][byte]; count every
 *      pattern the state reports.
 * At most 40,000 pattern bytes means at most 40,001 states: an int
 * table next[40001][256] (about 40 MB) is fine.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"sigs.txt"}, Files: map[string]string{"sigs.txt": "he\nshe\nhis\nhers\n"}, Stdin: "ushers", Want: matchWant("he\nshe\nhis\nhers\n", "ushers")},
			{Name: "one inside another", Args: []string{"sigs.txt"}, Files: map[string]string{"sigs.txt": "she\nhe\ne\nhers\n"}, Stdin: "she sells shells; hers", Want: matchWant("she\nhe\ne\nhers\n", "she sells shells; hers"),
				Explain: "he and e end inside she: a state must also report the patterns of its fail links, or the shorter signatures are missed whenever a longer one matches."},
			{Name: "overlaps", Args: []string{"sigs.txt"}, Files: map[string]string{"sigs.txt": "aa\naaa\nab\n"}, Stdin: "aaaaab", Want: matchWant("aa\naaa\nab\n", "aaaaab"),
				Explain: "aa occurs 4 times in aaaaa: after a match, keep going from the same state instead of starting again at the root."},
			{Name: "binary data", Args: []string{"sigs.txt"}, Files: map[string]string{"sigs.txt": "MZ\nPK\nELF\n"}, Gen: generated(func() (string, string) {
				in := "\x7fELF\x02\x01\x00\x00MZ\x90\x00\x03\x00PK\x03\x04\x00\x00\x7fELF"
				return in, matchWant("MZ\nPK\nELF\n", in)
			}),
				Explain: "File signatures in binary data, with zero bytes between them: read with fread and keep the length, since strlen and string functions stop at the first zero."},
			{Name: "repeated signatures", Args: []string{"sigs.txt"}, Files: map[string]string{"sigs.txt": "cat\ndog\ncat\n"}, Stdin: "cat dog cat", Want: matchWant("cat\ndog\ncat\n", "cat dog cat")},
			{Name: "a blank signature", Args: []string{"sigs.txt"}, Files: map[string]string{"sigs.txt": "ok\n\nfine\n"}, Stdin: "x", WantExit: 1, Stderr: true,
				Explain: "An empty signature would match everywhere: refuse it rather than flood the counts."},
			{Name: "five thousand signatures", Args: []string{"sigs.txt"}, Files: map[string]string{"sigs.txt": matchManyPatterns()}, Gen: generated(matchManyCase),
				Explain: "5,000 signatures and 16 MB of text: searching for each one separately reads 80 GB, but one Aho-Corasick pass reads 16 MB."},
			{Name: "no arguments", WantExit: 2, Stderr: true},
		},
		Timeout: 4 * time.Second, // the reference takes under 1 s with the sanitizers; one search per signature, about 16 s
		Hints: []string{
			"Read the whole text with fread into one malloc'd buffer, growing it as you go; keep its length, never call strlen on it.",
			"For the breadth-first pass, a plain int array used as a queue is enough: every state goes in once.",
			"Keep, per state, a small list of pattern numbers (or a link to the next state that ends a pattern), and add 1 to each pattern's count as you pass it.",
		},
	},
	{
		ID: "auth_detect", Stage: 19, Concept: "Detection Engineering with MITRE ATT&CK",
		Title: "Two login detection rules",
		Task:  "Read login events from standard input, one per line: TIME ADDRESS USER RESULT, where TIME is whole seconds (never decreasing), RESULT is ok or fail, and the other fields have no spaces. For each event, apply these rules, using the window of the 600 seconds up to it (times greater than TIME − 600). single-account: this fail brings the user's failures in the window to 5 or more; print TIME single-account USER. many-accounts: this fail brings the number of different users with a failure from this ADDRESS in the window to 10 or more; print TIME many-accounts ADDRESS. A rule that fired for a user or an address stays quiet for that same user or address until 600 seconds after it fired. success-after-alert: an ok for a user within 600 seconds after a single-account alert for that user; print TIME success-after-alert USER ADDRESS, once per alert. When one event fires two rules, print single-account first. Finally print alerts N. A malformed line, or a time earlier than the one before, gets a message on standard error and exit status 1. There can be 50,000 events, 2,000 users and 1,000 addresses.",
		File:  "auth_detect.c",
		Starter: `/* Checked exercise: auth_detect
 *   input:   100 10.0.0.5 ana fail
 *            110 10.0.0.5 ana fail
 *            120 10.0.0.5 ana fail
 *            130 10.0.0.5 ana fail
 *            140 10.0.0.5 ana fail
 *            150 10.0.0.5 ana ok
 *   output:  140 single-account ana
 *            150 success-after-alert ana 10.0.0.5
 *            alerts 2
 * The two failure rules see different shapes of trouble: many failures
 * for one account, and failures spread over many accounts from one
 * address. Neither sees the other's.
 * For each user, keep the times of recent failures (drop those at or
 * before TIME - 600); for each address, the users it failed for and
 * when each last failed. Linear searches through arrays are fast
 * enough at these sizes.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: "100 10.0.0.5 ana fail\n110 10.0.0.5 ana fail\n120 10.0.0.5 ana fail\n130 10.0.0.5 ana fail\n140 10.0.0.5 ana fail\n150 10.0.0.5 ana ok\n", Want: authWant("100 10.0.0.5 ana fail\n110 10.0.0.5 ana fail\n120 10.0.0.5 ana fail\n130 10.0.0.5 ana fail\n140 10.0.0.5 ana fail\n150 10.0.0.5 ana ok\n")},
			{Name: "ordinary mistakes", Gen: generated(func() (string, string) {
				var b strings.Builder
				for i := 0; i < 40; i++ {
					fmt.Fprintf(&b, "%d 10.1.0.%d user%d fail\n%d 10.1.0.%d user%d ok\n", 1000+i*30, i, i, 1010+i*30, i, i)
				}
				return b.String(), authWant(b.String())
			}),
				Explain: "Forty people each mistype a password once, from their own addresses: no rule may fire. A detection that fires on normal life gets switched off."},
			{Name: "the window's edge", Stdin: "0 10.0.0.9 bo fail\n100 10.0.0.9 bo fail\n200 10.0.0.9 bo fail\n300 10.0.0.9 bo fail\n600 10.0.0.9 bo fail\n601 10.0.0.9 bo fail\n", Want: authWant("0 10.0.0.9 bo fail\n100 10.0.0.9 bo fail\n200 10.0.0.9 bo fail\n300 10.0.0.9 bo fail\n600 10.0.0.9 bo fail\n601 10.0.0.9 bo fail\n"),
				Explain: "At time 600 the failure at 0 is exactly 600 seconds old and no longer counts (the window is times greater than TIME − 600), so the fifth failure in the window comes only at 601."},
			{Name: "many accounts from one address", Gen: generated(func() (string, string) {
				var b strings.Builder
				for i := 0; i < 15; i++ {
					fmt.Fprintf(&b, "%d 203.0.113.50 acct%02d fail\n", 5000+i*3, i)
				}
				return b.String(), authWant(b.String())
			}),
				Explain: "One failure each for 15 accounts from one address: the many-accounts rule fires at the tenth, and the per-account rule, rightly, never does."},
			{Name: "many failures, one account", Gen: generated(func() (string, string) {
				var b strings.Builder
				for i := 0; i < 12; i++ {
					fmt.Fprintf(&b, "%d 198.51.100.7 root fail\n", 7000+i*5)
				}
				return b.String(), authWant(b.String())
			}),
				Explain: "Twelve failures for one account from one address: one single-account alert, then quiet for 600 seconds. Counting failures instead of different users would also fire many-accounts here, and alerting at every failure after the fifth would raise eight alerts for one event."},
			{Name: "success after an alert", Gen: generated(func() (string, string) {
				var b strings.Builder
				for i := 0; i < 6; i++ {
					fmt.Fprintf(&b, "%d 192.0.2.44 cy fail\n", 100+i*10)
				}
				b.WriteString("200 192.0.2.44 cy ok\n300 192.0.2.44 cy ok\n2000 192.0.2.44 cy ok\n")
				return b.String(), authWant(b.String())
			}),
				Explain: "A success right after a burst of failures is the alert most worth waking someone for: the burst may have worked. It is reported once, at 200; the ok at 2000 is long after the alert."},
			{Name: "time going backwards", Stdin: "100 10.0.0.1 ana fail\n90 10.0.0.1 ana fail\n", WantExit: 1, Stderr: true,
				Explain: "Out-of-order events would make every window wrong; refuse them, and fix the clocks or the sorting upstream."},
			{Name: "a malformed line", Stdin: "100 10.0.0.1 ana\n", WantExit: 1, Stderr: true},
			{Name: "a day of logins", Gen: generated(authManyCase),
				Explain: "50,000 events from 2,000 users with two attacks hidden among them."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Read with fgets and sscanf(line, \"%ld %63s %63s %7s %c\", ...) == 4; check RESULT is exactly ok or fail.",
			"For a user, a ring or a growing array of failure times; drop from the front while the oldest is <= TIME - 600, then count.",
			"Keep, per user and per address, the time each rule last fired (or -1), and test TIME >= that + 600 before firing again.",
		},
	},
}

func sha256HexOf(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// chainSeal seals messages as the task describes.
func chainSeal(in string) string {
	var b strings.Builder
	prev := strings.Repeat("0", 64)
	for _, m := range strings.Split(strings.TrimSuffix(in, "\n"), "\n") {
		prev = sha256HexOf(prev + "\n" + m)
		b.WriteString(prev + " " + m + "\n")
	}
	return b.String()
}

func chainNewest(msgs string) string {
	s := chainSeal(msgs)
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	return lines[len(lines)-1][:64]
}

// chainMessages is n log messages, the same for the same seed.
func chainMessages(n int, seed int64) string {
	rng := rand.New(rand.NewSource(seed))
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "%d event=login user=u%d result=%s\n", 1792834867+i*7, rng.Intn(300), []string{"ok", "ok", "ok", "fail"}[rng.Intn(4)])
	}
	return b.String()
}

// matchWant counts overlapping occurrences with Aho-Corasick, as the
// C program must (searching for each pattern separately is too slow
// for the big case even here).
func matchWant(patterns, text string) string {
	pats := strings.Split(strings.TrimSuffix(patterns, "\n"), "\n")
	type node struct {
		next map[byte]int
		fail int
		out  []int
	}
	nodes := []node{{next: map[byte]int{}}}
	for i, p := range pats {
		v := 0
		for k := 0; k < len(p); k++ {
			u, ok := nodes[v].next[p[k]]
			if !ok {
				nodes = append(nodes, node{next: map[byte]int{}})
				u = len(nodes) - 1
				nodes[v].next[p[k]] = u
			}
			v = u
		}
		nodes[v].out = append(nodes[v].out, i)
	}
	queue := []int{}
	for _, u := range nodes[0].next {
		queue = append(queue, u)
	}
	sort.Ints(queue)
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		keys := make([]int, 0, len(nodes[v].next))
		for c := range nodes[v].next {
			keys = append(keys, int(c))
		}
		sort.Ints(keys)
		for _, ci := range keys {
			c := byte(ci)
			u := nodes[v].next[c]
			f := nodes[v].fail
			for {
				if w, ok := nodes[f].next[c]; ok && w != u {
					nodes[u].fail = w
					break
				}
				if f == 0 {
					nodes[u].fail = 0
					break
				}
				f = nodes[f].fail
			}
			nodes[u].out = append(nodes[u].out, nodes[nodes[u].fail].out...)
			queue = append(queue, u)
		}
	}
	counts := make([]int, len(pats))
	v := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		for {
			if u, ok := nodes[v].next[c]; ok {
				v = u
				break
			}
			if v == 0 {
				break
			}
			v = nodes[v].fail
		}
		for _, p := range nodes[v].out {
			counts[p]++
		}
	}
	var b strings.Builder
	total := 0
	for i, p := range pats {
		fmt.Fprintf(&b, "%d %s\n", counts[i], p)
		total += counts[i]
	}
	fmt.Fprintf(&b, "total %d\n", total)
	return b.String()
}

// matchManyPatterns is 5,000 short signatures (about 35 KB), built at
// startup, so kept small.
func matchManyPatterns() string {
	rng := rand.New(rand.NewSource(91))
	var b strings.Builder
	for i := 0; i < 5000; i++ {
		n := 4 + rng.Intn(7)
		for k := 0; k < n; k++ {
			b.WriteByte("abcdefghij"[rng.Intn(10)])
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func matchManyCase() (string, string) {
	rng := rand.New(rand.NewSource(92))
	text := make([]byte, 16<<20)
	for i := range text {
		text[i] = "abcdefghij"[rng.Intn(10)]
	}
	return string(text), matchWant(matchManyPatterns(), string(text))
}

// authWant applies the three rules exactly as the task describes.
func authWant(in string) string {
	type userState struct {
		fails     []int
		lastAlert int
		reported  bool
	}
	type addrState struct {
		lastFail  map[string]int
		lastAlert int
	}
	users := map[string]*userState{}
	addrs := map[string]*addrState{}
	var b strings.Builder
	n := 0
	for _, line := range strings.Split(strings.TrimSuffix(in, "\n"), "\n") {
		f := strings.Fields(line)
		var t int
		fmt.Sscan(f[0], &t)
		addr, user, result := f[1], f[2], f[3]
		u := users[user]
		if u == nil {
			u = &userState{lastAlert: -1}
			users[user] = u
		}
		a := addrs[addr]
		if a == nil {
			a = &addrState{lastFail: map[string]int{}, lastAlert: -1}
			addrs[addr] = a
		}
		if result == "ok" {
			if u.lastAlert >= 0 && !u.reported && t-u.lastAlert <= 600 {
				fmt.Fprintf(&b, "%d success-after-alert %s %s\n", t, user, addr)
				u.reported = true
				n++
			}
			continue
		}
		u.fails = append(u.fails, t)
		for len(u.fails) > 0 && u.fails[0] <= t-600 {
			u.fails = u.fails[1:]
		}
		if len(u.fails) >= 5 && (u.lastAlert < 0 || t >= u.lastAlert+600) {
			fmt.Fprintf(&b, "%d single-account %s\n", t, user)
			u.lastAlert, u.reported = t, false
			n++
		}
		a.lastFail[user] = t
		distinct := 0
		for _, last := range a.lastFail {
			if last > t-600 {
				distinct++
			}
		}
		if distinct >= 10 && (a.lastAlert < 0 || t >= a.lastAlert+600) {
			fmt.Fprintf(&b, "%d many-accounts %s\n", t, addr)
			a.lastAlert = t
			n++
		}
	}
	fmt.Fprintf(&b, "alerts %d\n", n)
	return b.String()
}

// authManyCase is a simulated day: ordinary logins with occasional
// mistakes, plus one burst on a single account and one address failing
// for many accounts.
func authManyCase() (string, string) {
	rng := rand.New(rand.NewSource(93))
	type ev struct {
		t                  int
		addr, user, result string
	}
	var evs []ev
	for len(evs) < 49000 {
		u := rng.Intn(2000)
		t := rng.Intn(86000)
		addr := fmt.Sprintf("10.%d.%d.%d", u%4, u/250, rng.Intn(4))
		if rng.Intn(10) == 0 {
			evs = append(evs, ev{t, addr, fmt.Sprintf("u%d", u), "fail"})
		}
		evs = append(evs, ev{t + 3, addr, fmt.Sprintf("u%d", u), "ok"})
	}
	for k := 0; k < 40; k++ {
		evs = append(evs, ev{30000 + k*4, "198.51.100.23", "u17", "fail"})
	}
	evs = append(evs, ev{30200, "198.51.100.23", "u17", "ok"})
	for k := 0; k < 300; k++ {
		evs = append(evs, ev{50000 + k*2, "203.0.113.99", fmt.Sprintf("u%d", rng.Intn(2000)), "fail"})
	}
	sort.SliceStable(evs, func(i, j int) bool { return evs[i].t < evs[j].t })
	var b strings.Builder
	for _, e := range evs {
		fmt.Fprintf(&b, "%d %s %s %s\n", e.t, e.addr, e.user, e.result)
	}
	return b.String(), authWant(b.String())
}
