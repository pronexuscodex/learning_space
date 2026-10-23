package main

// Checked exercises for the Software, Security & Theory track (Stages 9
// to 12), added when those stages were deepened.

import (
	"crypto/sha1"
	"fmt"
	"math/rand/v2"
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
