package main

// Checked exercises for the Software, Security & Theory track (Stages 9
// to 12), added when those stages were deepened.

import (
	"crypto/sha1"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"
	"time"
)

var softwareChecks = []Check{
	{
		ID: "git_blob", Stage: 9, Concept: "Version Control with Git",
		Title: "Git's object ID, by hand",
		Task:  "Print the ID Git gives a file: read all of standard input (any bytes, possibly megabytes), and print the SHA-1 of the header \"blob \", the length in decimal, a zero byte, and then the content, as 40 lowercase hex digits and a newline. Implement SHA-1 yourself; no libraries beyond the standard C library. Your output must match git hash-object for every file.",
		File:  "git_blob.c",
		Starter: `/* Checked exercise: git_blob
 * git hash-object FILE  ==  SHA-1( "blob " + length + "\0" + content )
 * SHA-1 in brief (FIPS 180-4):
 *   h0..h4 = 0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0
 *   Pad the message: one 0x80 byte, zeros until the length is 56 mod 64,
 *   then the message length in BITS as a 64-bit big-endian number.
 *   For each 64-byte block: w[0..15] = the block as big-endian 32-bit
 *   words; w[t] = rotl1(w[t-3] ^ w[t-8] ^ w[t-14] ^ w[t-16]) for t < 80.
 *   80 rounds with a..e = h0..h4:
 *     t <  20: f = (b & c) | (~b & d),          k = 0x5A827999
 *     t <  40: f = b ^ c ^ d,                   k = 0x6ED9EBA1
 *     t <  60: f = (b & c) | (b & d) | (c & d), k = 0x8F1BBCDC
 *     t <  80: f = b ^ c ^ d,                   k = 0xCA62C1D6
 *     tmp = rotl5(a) + f + e + k + w[t]; e = d; d = c; c = rotl30(b);
 *     b = a; a = tmp;
 *   then h0 += a, h1 += b, ... h4 += e. The digest is h0..h4, big-endian.
 * Use uint32_t for everything: unsigned arithmetic wraps, as SHA-1 needs.
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "an empty file", Stdin: "", Want: gitBlobID(""),
				Explain: "Even an empty file has a header: SHA-1 of \"blob 0\" and a zero byte is e69de29b…, the ID of every empty file in every Git repository."},
			{Name: "hello and a newline", Stdin: "hello\n", Want: gitBlobID("hello\n")},
			{Name: "no newline at the end", Stdin: "abc", Want: gitBlobID("abc")},
			{Name: "a 55-byte message", Stdin: strings.Repeat("x", 47), Want: gitBlobID(strings.Repeat("x", 47)),
				Explain: "Header (8 bytes) plus 47 is 55: the 0x80 byte and the 8-byte length just fit in the same 64-byte block."},
			{Name: "a 56-byte message", Stdin: strings.Repeat("y", 48), Want: gitBlobID(strings.Repeat("y", 48)),
				Explain: "56 bytes leave no room for the 8-byte length after the 0x80 byte: the padding needs a whole second block."},
			{Name: "a 64-byte message", Stdin: strings.Repeat("z", 56), Want: gitBlobID(strings.Repeat("z", 56)),
				Explain: "A message that fills a block exactly still gets a full block of padding after it."},
			{Name: "a three-digit length", Stdin: strings.Repeat("q", 111), Want: gitBlobID(strings.Repeat("q", 111)),
				Explain: "The header holds the length in decimal, so it grows with the number of digits: \"blob 111\" is 8 characters before the zero byte."},
			{Name: "every byte value, 2 MB", Gen: generated(func() (string, string) { b := allBytes(2 << 20); return b, gitBlobID(b) }),
				Explain: "Read the input as bytes (fread), not as text: zero bytes and bytes above 127 are content too. The length in bits is above 2³², so it needs 64 bits."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Read everything first (fread into a buffer you grow with realloc): the header needs the length before any content is hashed.",
			"rotl(x, n) is (x << n) | (x >> (32 - n)) on a uint32_t; n is always between 1 and 31 here, so neither shift is undefined.",
			"Feed the hash in pieces with a 64-byte buffer: update(header), update(content), then finish by adding the padding and the bit length.",
		},
	},
	{
		ID: "make_order", Stage: 9, Concept: "Build Systems & Dependencies",
		Title: "The order make builds in",
		Task:  "./make_order TARGET reads rules from standard input, one per line: a target name, a colon, and its prerequisites separated by spaces (a name with no rule is a source file). Print every target that TARGET needs, TARGET included, in the order make would build them: depth-first, prerequisites in the order listed, each target after all its prerequisites and only once; source files are not printed. If the targets TARGET needs form a cycle, print a message to standard error and return 1; if TARGET has no rule, or no TARGET is given, print a message to standard error and return 2. There can be 100,000 rules; lines are at most 4,096 characters, names at most 63.",
		File:  "make_order.c",
		Starter: `/* Checked exercise: make_order
 * Input (standard input):         ./make_order app  prints:
 *   app: main.o util.o              main.o
 *   main.o: main.c util.h           util.o
 *   util.o: util.c util.h           app
 * Blank lines are allowed; "lib:" (no prerequisites) is a rule too.
 * Depth-first search with three states per target: not started, in
 * progress, done. Meeting a target that is "in progress" means a cycle.
 * Print a target when it becomes "done".
 * 100,000 rules: look names up with a hash table, not by scanning.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the example", Args: []string{"app"}, Stdin: "app: main.o util.o\nmain.o: main.c util.h\nutil.o: util.c util.h\n", Want: "main.o\nutil.o\napp\n"},
			{Name: "a shared prerequisite, once", Args: []string{"all"}, Stdin: "all: a b\na: common\nb: common\ncommon: x.c\n", Want: "common\na\nb\nall\n",
				Explain: "common is needed twice but built once, before the first target that needs it."},
			{Name: "the listed order", Args: []string{"all"}, Stdin: "a:\nall: c b a\nb:\n\nc: a\n", Want: "a\nc\nb\nall\n",
				Explain: "Prerequisites are visited in the order listed (c, b, a), each one's own prerequisites first: c needs a, so a comes before c."},
			{Name: "a target lower down", Args: []string{"util.o"}, Stdin: "app: main.o util.o\nmain.o: main.c\nutil.o: util.c\n", Want: "util.o\n"},
			{Name: "a cycle", Args: []string{"a"}, Stdin: "a: b\nb: c\nc: a\n", WantExit: 1, Stderr: true,
				Explain: "a needs b needs c needs a: meeting a target that is still in progress means there is no possible order."},
			{Name: "a target that needs itself", Args: []string{"loop"}, Stdin: "loop: loop\n", WantExit: 1, Stderr: true},
			{Name: "a cycle elsewhere", Args: []string{"app"}, Stdin: "app: main.c\nx: y\ny: x\n", Want: "app\n",
				Explain: "Only what TARGET needs matters: a cycle among other targets is not an error here."},
			{Name: "no rule for the target", Args: []string{"nothing"}, Stdin: "app: main.c\n", WantExit: 2, Stderr: true},
			{Name: "no target", Stdin: "app: main.c\n", WantExit: 2, Stderr: true},
			{Name: "100,000 rules", Args: []string{"all"}, Gen: generated(func() (string, string) { return makeOrderInput(), makeOrderWant() }),
				Explain: "A visited check that scans everything built so far, or a name lookup that scans every rule, makes 10¹⁰ steps here: use a hash table for both."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Give every name an index in a hash table the first time you see it; store each rule's prerequisites as a list of indexes.",
			"visit(t): if done, return; if in progress, it is a cycle; mark in progress, visit each prerequisite that has a rule, mark done, print it.",
			"Read lines with fgets into a 4,097-byte buffer, split at the colon, then split the rest with strtok on spaces.",
		},
	},
	{
		ID: "bisect", Stage: 9, Concept: "Debugging Tools: gdb, Core Dumps & git bisect",
		Title: "git bisect, as a function",
		Task:  "Write long first_bad(long good, long bad, int (*is_bad)(long)). Versions good and bad are already known (good < bad, version good works, version bad does not), and once a version is bad every later one is too. Return the first bad version, calling is_bad only for versions strictly between good and bad, and at most ⌈log₂(bad − good)⌉ times. Versions can be as large as LONG_MAX. No main: the checker supplies is_bad and counts its calls.",
		File:  "bisect.c",
		Starter: `/* Checked exercise: bisect
 *   long first_bad(long good, long bad, int (*is_bad)(long));
 * Keep the invariant: good is good, bad is bad. While they are more than
 * 1 apart, test the version halfway between and move one end to it.
 * When bad == good + 1, bad is the first bad version.
 * The midpoint of two big longs: good + (bad - good) / 2, never
 * (good + bad) / 2, which overflows.
 */
`,
		Harness: `#include <limits.h>
#include <stdio.h>
#include <stdlib.h>

long first_bad(long good, long bad, int (*is_bad)(long));

static long lo, hi, answer;
static int calls, outside;

static int is_bad(long v) {
    calls++;
    if (v <= lo || v >= hi) outside = 1;
    return v >= answer;
}

static int limit(long n) {   /* ceil(log2 n) */
    int k = 0;
    while (n > 1) { n = (n + 1) / 2; k++; }
    return k;
}

int main(int argc, char **argv) {
    if (argc < 4) return 2;
    lo = strtol(argv[1], NULL, 0);
    hi = strtol(argv[2], NULL, 0);
    answer = strtol(argv[3], NULL, 0);
    long got = first_bad(lo, hi, is_bad);
    printf("first bad: %ld\n", got);
    if (outside) printf("tested a version that is not strictly between good and bad\n");
    if (calls > limit(hi - lo)) printf("%d tests, more than the %d needed\n", calls, limit(hi - lo));
    else printf("tests: within the limit\n");
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "eight versions", Args: []string{"0", "8", "5"}, Want: "first bad: 5\ntests: within the limit\n"},
			{Name: "neighbours", Args: []string{"41", "42", "42"}, Want: "first bad: 42\ntests: within the limit\n",
				Explain: "When good and bad are adjacent there is nothing to test: bad is the answer, with no call to is_bad."},
			{Name: "the last version", Args: []string{"0", "1000", "1000"}, Want: "first bad: 1000\ntests: within the limit\n",
				Explain: "Every version before 1000 is good: the answer can be bad itself, which you must never test."},
			{Name: "the first version", Args: []string{"0", "1000", "1"}, Want: "first bad: 1\ntests: within the limit\n"},
			{Name: "a thousand commits", Args: []string{"0", "1000", "637"}, Want: "first bad: 637\ntests: within the limit\n",
				Explain: "Halving the range each time needs ⌈log₂ 1000⌉ = 10 tests; testing every version in turn needs hundreds."},
			{Name: "near LONG_MAX", Args: []string{"0x4000000000000000", "0x7fffffffffffffff", "0x4000000000003039"}, Want: "first bad: 4611686018427400249\ntests: within the limit\n",
				Explain: "(good + bad) / 2 overflows a long here: signed overflow is undefined behaviour. good + (bad - good) / 2 cannot overflow."},
		},
		Hints: []string{
			"while (bad - good > 1) { long mid = good + (bad - good) / 2; if (is_bad(mid)) bad = mid; else good = mid; } return bad;",
			"good and bad themselves are known: testing them wastes a test, and the checker counts it as outside the range.",
		},
	},
	{
		ID: "nfa_match", Stage: 11, Concept: "NFAs, DFAs & How Regex Engines Work",
		Title: "A regex engine that never backtracks",
		Task:  "./nfa_match PATTERN reads lines from standard input and prints match or no for each, depending on whether the whole line matches PATTERN. Patterns use letters, digits and other printable characters as literals, . for any character, postfix *, + and ?, | for alternation (lowest precedence) and parentheses for grouping. Run the pattern as an NFA, keeping the set of states it could be in, so that no pattern and no input takes exponential time: lines can be 100,000 characters long. An invalid pattern (an unmatched parenthesis, or *, + or ? with nothing before it) gets a message on standard error and exit status 2.",
		File:  "nfa_match.c",
		Starter: `/* Checked exercise: nfa_match
 * ./nfa_match 'ab*c'  <  lines   prints match / no for each line.
 * The whole line must match, as if the pattern were ^(PATTERN)$.
 * One way, after Thompson (1968) and Russ Cox's "Regular Expression
 * Matching Can Be Simple And Fast":
 *   1. Rewrite the pattern in postfix with explicit concatenation:
 *      a(b|c)*d  ->  a b c | * . d .
 *   2. Build an NFA from the postfix with a stack of fragments: each
 *      operator pops one or two fragments and pushes a new one.
 *   3. Simulate: a list of current states; for each input character,
 *      the list of next states (follow split states at once). Use a
 *      "last seen" mark per state so each state is added once per step.
 * The line matches if the match state is in the list at the end.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "stars", Args: []string{"ab*c"}, Stdin: "ac\nabc\nabbbbc\nabd\n\nxabc\n", Want: "match\nmatch\nmatch\nno\nno\nno\n",
				Explain: "The whole line must match: \"xabc\" contains abc but does not match ab*c from its first character."},
			{Name: "alternation and grouping", Args: []string{"(cat|dog)s?"}, Stdin: "cat\ndogs\ncats\ncow\ncatdog\n", Want: "match\nmatch\nmatch\nno\nno\n"},
			{Name: "precedence", Args: []string{"ab|cd"}, Stdin: "ab\ncd\nabd\nacd\n", Want: "match\nmatch\nno\nno\n",
				Explain: "| binds loosest: ab|cd is (ab)|(cd), not a(b|c)d."},
			{Name: "plus, question mark and dot", Args: []string{"(ab)+.?"}, Stdin: "\nab\nababX\naba\nabXY\n", Want: "no\nmatch\nmatch\nmatch\nno\n"},
			{Name: "nested repetition", Gen: generated(func() (string, string) { return nfaMatchCase("(a+)+b", 30) }), Args: []string{"(a+)+b"},
				Explain: "A backtracking matcher tries every way to split the a's between the two +'s: 2³⁰ tries. Tracking the set of possible states reads each character once."},
			{Name: "long lines", Gen: generated(func() (string, string) { return nfaMatchCase("(a|b)*a(a|b)(a|b)(a|b)", 100000) }), Args: []string{"(a|b)*a(a|b)(a|b)(a|b)"}},
			{Name: "an unmatched parenthesis", Args: []string{"(ab"}, Stdin: "ab\n", WantExit: 2, Stderr: true},
			{Name: "a star with nothing before it", Args: []string{"*a"}, Stdin: "a\n", WantExit: 2, Stderr: true},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"In the postfix conversion, track for each open parenthesis how many atoms and alternatives you have seen; an operator with zero atoms before it is an error.",
			"An NFA state: a character to match (or ANY, SPLIT or MATCH) and one or two out-pointers. A fragment is a start state plus the list of dangling out-pointers to patch.",
			"Adding a state to a list: if it is a SPLIT, add both of its outs instead (recursively); skip a state already added in this step.",
		},
	},
	{
		ID: "turing", Stage: 11, Concept: "Turing Machines & Computability",
		Title: "Run a Turing machine",
		Task:  "./turing LIMIT reads a two-symbol Turing machine from standard input, one rule per line: STATE READ WRITE MOVE NEXT, such as A 0 1 R B (MOVE is L or R; the state named H means halt). Start in the state of the first rule, on a blank tape of 0s. Run until the machine halts or LIMIT steps have been taken, then print halted after N steps, K ones, or still running after LIMIT steps. Moving into H counts as a step. The head stays within 100,000 cells of where it started; LIMIT can be 100 million.",
		File:  "turing.c",
		Starter: `/* Checked exercise: turing
 * ./turing 1000  <  machine
 *   A 0 1 R B        in state A reading 0: write 1, move right, go to B
 *   A 1 1 L B
 *   B 0 1 L A
 *   B 1 1 R H        H: halt
 * prints:  halted after 6 steps, 4 ones
 * A table rule[state][symbol] with the write, the move and the next
 * state; states are single letters. The tape: an array of 200,001
 * cells with the head starting in the middle.
 * LIMIT can be 100,000,000: keep the inner loop simple.
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "the two-state busy beaver", Args: []string{"1000"}, Stdin: "A 0 1 R B\nA 1 1 L B\nB 0 1 L A\nB 1 1 R H\n", Want: turingWant("A 0 1 R B\nA 1 1 L B\nB 0 1 L A\nB 1 1 R H\n", 1000),
				Explain: "The move into H is the last step: 6 steps in all, leaving four 1s on the tape."},
			{Name: "the three-state busy beaver", Args: []string{"1000"}, Stdin: turingBB3, Want: turingWant(turingBB3, 1000)},
			{Name: "the four-state busy beaver", Args: []string{"1000"}, Stdin: turingBB4, Want: turingWant(turingBB4, 1000),
				Explain: "A rule may write 0 over a 1: count the 1s left on the tape at the end, not the times a 1 was written."},
			{Name: "a machine that never halts", Args: []string{"1000"}, Stdin: "A 0 1 R A\nA 1 1 R A\n", Want: "still running after 1000 steps\n",
				Explain: "Stop after LIMIT steps: some machines run forever, and no program can tell in general which ones."},
			{Name: "a limit hit exactly", Args: []string{"5"}, Stdin: "A 0 1 R B\nA 1 1 L B\nB 0 1 L A\nB 1 1 R H\n", Want: "still running after 5 steps\n"},
			{Name: "the five-state champion", Args: []string{"100000000"}, Stdin: turingBB5, Want: "halted after 47176870 steps, 4098 ones\n",
				Explain: "47 million steps: an inner loop that looks rules up by searching the list, or that resizes the tape on every step, runs out of time."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Map the state letter to an index (c - 'A'), and H to a special value; store rule[state][symbol] as three small numbers.",
			"Count the steps in a long: the four-state machine is small, but limits can exceed 2³¹ for other machines.",
			"Count the ones at the end by scanning the tape once.",
		},
	},
	{
		ID: "dpll", Stage: 11, Concept: "Reductions & NP-Completeness",
		Title: "A SAT solver",
		Task:  "Read a formula in DIMACS CNF from standard input (comment lines start with c; the header p cnf VARIABLES CLAUSES; then each clause as non-zero integers, a negative one meaning NOT, ending with 0) and decide whether it can be satisfied. Print UNSAT, or SAT followed by a line listing the variables that are true, in increasing order, separated by spaces (an empty line if none). Every satisfiable formula in the tests has exactly one solution. One formula is a Sudoku with 729 variables: trying all assignments is hopeless, so use DPLL with unit propagation.",
		File:  "dpll.c",
		Starter: `/* Checked exercise: dpll
 * DPLL(formula, assignment):
 *   repeat unit propagation: a clause with every literal false but one
 *     unassigned literal forces that literal true; a clause with every
 *     literal false is a conflict: return UNSAT for this branch
 *   if every clause has a true literal, return SAT
 *   pick an unassigned variable v; try v = true, then v = false
 *   (undo the assignments made in a branch when it fails)
 * Literal l is true when value[abs(l)] == (l > 0).
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "a chain of implications", Stdin: "c x1, x1 -> x2, not both x2 and x3\np cnf 3 3\n1 0\n-1 2 0\n-2 -3 0\n", Want: "SAT\n1 2\n",
				Explain: "Unit propagation alone solves this: x1 is forced, then x2, and then x3 must be false."},
			{Name: "a contradiction", Stdin: "p cnf 1 2\n1 0\n-1 0\n", Want: "UNSAT\n"},
			{Name: "nothing true", Stdin: "p cnf 2 2\n-1 0\n-2 0\n", Want: "SAT\n\n"},
			{Name: "clauses over several lines", Stdin: "p cnf 3 4\n-1\n0 2 3\n0 -3 -2 0\n2\n0\n", Want: "SAT\n2\n",
				Explain: "A clause ends at its 0, not at the end of the line: read numbers, not lines."},
			{Name: "four pigeons, three holes", Stdin: pigeonholeCNF(4, 3), Want: "UNSAT\n",
				Explain: "No assignment works, so the search must try every branch and backtrack correctly out of each one."},
			{Name: "the hardest Sudoku", Gen: generated(func() (string, string) { return sudokuCNF(sudokuInkala) }),
				Explain: "2⁷²⁹ assignments cannot be tried one by one. Unit propagation fills in every forced cell; the search then needs only hundreds of decisions."},
			{Name: "six pigeons, five holes", Stdin: pigeonholeCNF(6, 5), Want: "UNSAT\n"},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Store the clauses in one array of literals with a start index per clause; scan them all in each propagation round (simple, and fast enough here).",
			"Keep a trail: every variable you assign, in order. To undo a branch, unassign everything after the trail position where it began.",
			"Branch on a variable from the shortest clause that is not yet satisfied: it is the most constrained choice.",
		},
	},
}

// gitBlobID is the ID git hash-object gives a file with this content.
func gitBlobID(content string) string {
	return fmt.Sprintf("%x\n", sha1.Sum([]byte(fmt.Sprintf("blob %d\x00%s", len(content), content))))
}

// makeOrderRules is the big make_order case: 1,000 layers of 100 targets,
// each needing three targets of the layer below; the last layer needs
// source files; "all" needs the whole first layer.
func makeOrderRules() (order []string, rules map[string][]string) {
	r := rand.New(rand.NewPCG(13, 13))
	const layers, width = 1000, 100
	rules = map[string][]string{}
	name := func(l, i int) string { return fmt.Sprintf("t%d_%d", l, i) }
	var top []string
	for i := 0; i < width; i++ {
		top = append(top, name(0, i))
	}
	rules["all"] = top
	order = append(order, "all")
	for l := 0; l < layers; l++ {
		for i := 0; i < width; i++ {
			var deps []string
			if l == layers-1 {
				deps = []string{fmt.Sprintf("src%d.c", i)}
			} else {
				for _, j := range r.Perm(width)[:3] {
					deps = append(deps, name(l+1, j))
				}
			}
			rules[name(l, i)] = deps
			order = append(order, name(l, i))
		}
	}
	r.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	return order, rules
}

func makeOrderInput() string {
	order, rules := makeOrderRules()
	var sb strings.Builder
	for _, t := range order {
		fmt.Fprintf(&sb, "%s: %s\n", t, strings.Join(rules[t], " "))
	}
	return sb.String()
}

func makeOrderWant() string {
	_, rules := makeOrderRules()
	done := map[string]bool{}
	var sb strings.Builder
	var visit func(string)
	visit = func(t string) {
		if done[t] {
			return
		}
		deps, ok := rules[t]
		if !ok {
			return
		}
		done[t] = true // no cycles in this graph
		for _, d := range deps {
			visit(d)
		}
		sb.WriteString(t + "\n")
	}
	visit("all")
	return sb.String()
}

// nfaMatchCase builds lines for a pattern: strings of a's of growing
// length, some with a final b, and the matches Go's RE2 engine finds.
func nfaMatchCase(pattern string, n int) (stdin, want string) {
	re := regexp.MustCompile("^(?:" + pattern + ")$")
	var in, out strings.Builder
	r := rand.New(rand.NewPCG(14, 14))
	for _, k := range []int{n, n - 1, n / 2, 1, n} {
		line := strings.Repeat("a", k)
		if k != n && r.IntN(2) == 0 { // the first line of n a's never matches (a+)+b
			line += "b"
		}
		if pattern != "(a+)+b" { // mix in b's for the long-line pattern
			b := []byte(line)
			for i := range b {
				if r.IntN(3) == 0 {
					b[i] = 'b'
				}
			}
			line = string(b)
		}
		in.WriteString(line + "\n")
		if re.MatchString(line) {
			out.WriteString("match\n")
		} else {
			out.WriteString("no\n")
		}
	}
	return in.String(), out.String()
}

const (
	turingBB3 = "A 0 1 R B\nA 1 1 R H\nB 0 0 R C\nB 1 1 R B\nC 0 1 L C\nC 1 1 L A\n"
	turingBB4 = "A 0 1 R B\nA 1 1 L B\nB 0 1 L A\nB 1 0 L C\nC 0 1 R H\nC 1 1 L D\nD 0 1 R D\nD 1 0 R A\n"
	turingBB5 = "A 0 1 R B\nA 1 1 L C\nB 0 1 R C\nB 1 1 R B\nC 0 1 R D\nC 1 0 L E\nD 0 1 L A\nD 1 1 L D\nE 0 1 R H\nE 1 0 L A\n"
)

// turingWant runs a machine in the turing check's format.
func turingWant(machine string, limit int) string {
	type rule struct {
		write, move int
		next        byte
	}
	rules := map[[2]byte]rule{}
	var start byte
	for _, line := range strings.Split(strings.TrimSpace(machine), "\n") {
		f := strings.Fields(line)
		if start == 0 {
			start = f[0][0]
		}
		move := 1
		if f[3] == "L" {
			move = -1
		}
		rules[[2]byte{f[0][0], f[1][0]}] = rule{int(f[2][0] - '0'), move, f[4][0]}
	}
	tape := map[int]int{}
	pos, state := 0, start
	for steps := 1; steps <= limit; steps++ {
		r := rules[[2]byte{state, byte('0' + tape[pos])}]
		tape[pos] = r.write
		pos += r.move
		state = r.next
		if state == 'H' {
			ones := 0
			for _, v := range tape {
				ones += v
			}
			return fmt.Sprintf("halted after %d steps, %d ones\n", steps, ones)
		}
	}
	return fmt.Sprintf("still running after %d steps\n", limit)
}

// pigeonholeCNF says that p pigeons sit in h holes, each in at least one
// hole and no two in the same one: unsatisfiable when p > h.
func pigeonholeCNF(p, h int) string {
	v := func(i, j int) int { return i*h + j + 1 }
	var cl []string
	for i := 0; i < p; i++ {
		var c []string
		for j := 0; j < h; j++ {
			c = append(c, fmt.Sprint(v(i, j)))
		}
		cl = append(cl, strings.Join(c, " ")+" 0")
	}
	for j := 0; j < h; j++ {
		for a := 0; a < p; a++ {
			for b := a + 1; b < p; b++ {
				cl = append(cl, fmt.Sprintf("-%d -%d 0", v(a, j), v(b, j)))
			}
		}
	}
	return fmt.Sprintf("c %d pigeons, %d holes\np cnf %d %d\n%s\n", p, h, p*h, len(cl), strings.Join(cl, "\n"))
}

// sudokuInkala is the puzzle Arto Inkala published in 2012.
const sudokuInkala = "800000000003600000070090200050007000000045700000100030001000068008500010090000400"

// sudokuCNF encodes a Sudoku as SAT (variable 81r + 9c + d + 1 means
// cell (r, c) holds digit d + 1) and lists the true variables of its
// unique solution.
func sudokuCNF(puzzle string) (stdin, want string) {
	v := func(r, c, d int) int { return 81*r + 9*c + d + 1 }
	var cl []string
	add := func(lits ...int) {
		var s []string
		for _, l := range lits {
			s = append(s, fmt.Sprint(l))
		}
		cl = append(cl, strings.Join(s, " ")+" 0")
	}
	var groups [][][2]int
	for i := 0; i < 9; i++ {
		var row, col, box [][2]int
		for j := 0; j < 9; j++ {
			row = append(row, [2]int{i, j})
			col = append(col, [2]int{j, i})
			box = append(box, [2]int{3*(i/3) + j/3, 3*(i%3) + j%3})
		}
		groups = append(groups, row, col, box)
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			var all []int
			for d := 0; d < 9; d++ {
				all = append(all, v(r, c, d))
				for e := d + 1; e < 9; e++ {
					add(-v(r, c, d), -v(r, c, e))
				}
			}
			add(all...)
		}
	}
	for _, g := range groups {
		for d := 0; d < 9; d++ {
			var all []int
			for a := range g {
				all = append(all, v(g[a][0], g[a][1], d))
				for b := a + 1; b < 9; b++ {
					add(-v(g[a][0], g[a][1], d), -v(g[b][0], g[b][1], d))
				}
			}
			add(all...)
		}
	}
	grid := []byte(puzzle)
	for i, ch := range grid {
		if ch != '0' {
			add(v(i/9, i%9, int(ch-'1')))
		}
	}
	solveSudoku(grid)
	var trues []string
	for i, ch := range grid {
		trues = append(trues, fmt.Sprint(v(i/9, i%9, int(ch-'1'))))
	}
	return fmt.Sprintf("c Sudoku\np cnf 729 %d\n%s\n", len(cl), strings.Join(cl, "\n")),
		"SAT\n" + strings.Join(trues, " ") + "\n"
}

// solveSudoku fills grid by backtracking; it reports whether it could.
func solveSudoku(grid []byte) bool {
	for i, ch := range grid {
		if ch != '0' {
			continue
		}
		r, c := i/9, i%9
		for d := byte('1'); d <= '9'; d++ {
			ok := true
			for k := 0; k < 9 && ok; k++ {
				br, bc := 3*(r/3)+k/3, 3*(c/3)+k%3
				ok = grid[9*r+k] != d && grid[9*k+c] != d && grid[9*br+bc] != d
			}
			if ok {
				grid[i] = d
				if solveSudoku(grid) {
					return true
				}
			}
		}
		grid[i] = '0'
		return false
	}
	return true
}
