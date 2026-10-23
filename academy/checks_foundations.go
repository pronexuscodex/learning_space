package main

// Checked exercises added when Stages 2 to 8 were deepened: data
// structures and algorithms (Stage 2), discrete mathematics (Stage 3),
// computer architecture (Stage 4), computer systems (Stage 5) and
// operating systems (Stage 6), storage engines (Stage 7) and networks
// (Stage 8). Most read standard input and are timed on
// inputs big enough that only the efficient idea passes.

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"
)

var foundationsChecks = []Check{
	{
		ID: "duplicates", Stage: 2, Concept: "Big-O: Measuring Growth",
		Title: "The first repeated number",
		Task:  "Read n (0 to 200,000) and then n integers. Reading from left to right, print the first number that appears for the second time, or none if every number is different. Two nested loops are O(n²): fine for a hundred numbers, far too slow for 200,000. Find an O(n log n) way.",
		File:  "duplicates.c",
		Starter: `/* Checked exercise: duplicates
 * Input:  n, then n integers (any int, negative too).
 * Output: the number whose second appearance comes first, or "none".
 *   5
 *   3 1 4 1 3      prints 1: its second copy (position 3) comes before
 *                  the second 3 (position 4)
 * Comparing every pair is O(n^2): 200,000 numbers make 20 billion pairs.
 * Sort a copy instead (qsort is O(n log n)), remembering where each
 * number came from.
 */
`,
		Cases: []CheckCase{
			{Name: "the example", Stdin: "5\n3 1 4 1 3\n", Want: "1\n",
				Explain: "1 and 3 both repeat, but the second 1 comes first (position 3, before the second 3 at position 4)."},
			{Name: "all different", Stdin: "4\n10 20 30 40\n", Want: "none\n"},
			{Name: "no numbers", Stdin: "0\n", Want: "none\n"},
			{Name: "extremes and negatives", Stdin: "6\n-5 2147483647 -2147483648 7 -2147483648 -5\n", Want: "-2147483648\n",
				Explain: "A comparison function that returns a - b overflows for numbers this far apart: return (a > b) - (a < b)."},
			{Name: "200,000 different numbers", Stdin: duplicatesInput(200000, -1, -1), Want: "none\n",
				Explain: "Comparing every pair is 20 billion comparisons. Sort (value, position) pairs with qsort, so equal values sit side by side."},
			{Name: "a late repeat", Stdin: duplicatesInput(200000, 199998, 199999), Want: fmt.Sprintln(duplicatesValue(199998))},
		},
		Hints: []string{
			"Make an array of structs { int value; int pos; } and qsort it by value, then by position.",
			"After sorting, every group of equal values sits together. For each group with two or more entries, its second entry's pos is when that value repeats; keep the smallest such pos.",
			"Write the comparison as (x > y) - (x < y): x - y overflows when one is very negative and the other very positive.",
		},
	},
	{
		ID: "hanoi", Stage: 2, Concept: "Recursion & Divide and Conquer",
		Title: "The Towers of Hanoi",
		Task:  "Read n (0 to 16): n disks, smallest numbered 1, start on peg A. Print every move that brings them all to peg C, one per line as \"disk from to\" (\"1 A C\"), never putting a disk on a smaller one. Use the recursive plan: move the top n − 1 disks to the spare peg, move disk n, move the n − 1 back on top. It takes 2ⁿ − 1 moves.",
		File:  "hanoi.c",
		Starter: `/* Checked exercise: hanoi
 * Input:  n (0 to 16).
 * Output: the moves taking n disks from peg A to peg C, one per line:
 *   1 A B
 *   2 A C
 *   1 B C          (n = 2)
 * void hanoi(int n, char from, char to, char via):
 *   if n is 0, there is nothing to move (the base case);
 *   otherwise: hanoi(n - 1, from, via, to); print the move of disk n;
 *              hanoi(n - 1, via, to, from);
 */
`,
		Cases: []CheckCase{
			{Name: "one disk", Stdin: "1\n", Want: "1 A C\n"},
			{Name: "two disks", Stdin: "2\n", Want: "1 A B\n2 A C\n1 B C\n",
				Explain: "Disk 1 goes to the spare peg B first, so disk 2 can go straight to C."},
			{Name: "three disks", Stdin: "3\n", Want: hanoiMoves(3)},
			{Name: "no disks", Stdin: "0\n", Want: "",
				Explain: "With 0 disks there is nothing to do: the base case prints nothing."},
			{Name: "sixteen disks", Stdin: "16\n", Want: hanoiMoves(16),
				Explain: "65,535 moves: 2¹⁶ − 1. The recursion is only 16 calls deep, however many moves it prints."},
		},
		Hints: []string{
			"Trust the smaller call: if hanoi(n - 1, …) moves n − 1 disks correctly, three lines finish the job.",
			"Print the move of disk n between the two recursive calls: printf(\"%d %c %c\\n\", n, from, to).",
		},
	},
	{
		ID: "bsearch", Stage: 2, Concept: "Searching: Linear & Binary Search",
		Title: "Binary search, first match",
		Task:  "Read n (0 to 200,000), then n integers in non-decreasing order (repeats allowed), then q (up to 100,000), then q numbers to look up. For each, print the position (from 0) of its first occurrence, or -1 if it is not there. A linear scan per lookup is too slow; halve the range instead, and when you find a match, keep looking to its left.",
		File:  "bsearch.c",
		Starter: `/* Checked exercise: bsearch
 * Input:   n, then n sorted ints (repeats allowed), then q, then q ints.
 * Output:  for each of the q ints, the index of its FIRST occurrence,
 *          or -1, one per line.
 *   5
 *   1 3 3 3 7
 *   3
 *   3 4 1          prints 1, -1 and 0
 * Keep an invariant and check every branch against it, for example:
 *   everything before low is < x, everything from high on is >= x.
 * Compute the middle as low + (high - low) / 2.
 */
`,
		Cases: []CheckCase{
			{Name: "the example", Stdin: "5\n1 3 3 3 7\n3\n3 4 1\n", Want: "1\n-1\n0\n",
				Explain: "With repeats, finding a 3 is not enough: report the first one, at index 1."},
			{Name: "edges", Stdin: "4\n2 4 6 8\n6\n2 8 1 9 5 6\n", Want: "0\n3\n-1\n-1\n-1\n2\n",
				Explain: "Check the first and last items, numbers smaller and larger than every item, and a gap in the middle."},
			{Name: "empty array", Stdin: "0\n2\n5 -5\n", Want: "-1\n-1\n"},
			{Name: "all the same", Stdin: "6\n9 9 9 9 9 9\n2\n9 10\n", Want: "0\n-1\n",
				Explain: "A plain binary search returns whichever 9 it meets first, often index 2: after a match, keep searching to the left."},
			{Name: "200,000 numbers, 100,000 lookups", Stdin: bsearchInput(), Want: bsearchWant(),
				Explain: "A linear scan per lookup is up to 20 billion comparisons; binary search needs about 18 per lookup."},
		},
		Hints: []string{
			"Find the first position whose value is >= x (lower bound); x is there only if that position is < n and holds x.",
			"low = 0, high = n; while (low < high) { mid = low + (high - low) / 2; if (a[mid] < x) low = mid + 1; else high = mid; }",
			"200,000 ints fit easily in a static array (or one malloc); read them with scanf(\"%d\", …).",
		},
	},
	{
		ID: "intervals", Stage: 2, Concept: "Greedy Algorithms",
		Title: "The most meetings in one room",
		Task:  "Read n (0 to 200,000), then n meetings, each a start and an end time (0 ≤ start < end ≤ 1,000,000,000). A meeting holds the room from its start up to its end, so another may start exactly when it ends. Print the largest number of meetings that fit in one room without overlapping. Sort by end time and take each meeting that starts after the last one taken ends; other greedy rules look right and fail.",
		File:  "intervals.c",
		Starter: `/* Checked exercise: intervals
 * Input:  n, then n lines "start end".
 * Output: the most meetings that fit in one room without overlapping.
 *   5
 *   0 3
 *   2 5
 *   4 7
 *   6 9
 *   8 10           prints 3 (0-3, 4-7 and 8-10)
 * A meeting may start exactly when the previous one ends.
 * The safe greedy choice: always take the meeting that ENDS first.
 */
`,
		Cases: []CheckCase{
			{Name: "the example", Stdin: "5\n0 3\n2 5\n4 7\n6 9\n8 10\n", Want: "3\n"},
			{Name: "back to back", Stdin: "3\n0 5\n5 10\n10 15\n", Want: "3\n",
				Explain: "A meeting that starts exactly when the last one ends does not overlap it: compare start >= last end."},
			{Name: "one long meeting", Stdin: "3\n0 100\n1 2\n3 4\n", Want: "2\n",
				Explain: "Taking the meeting that starts first (0-100) blocks the other two: sort by end time, not start time."},
			{Name: "the shortest is not the best", Stdin: "3\n0 6\n5 8\n7 12\n", Want: "2\n",
				Explain: "The shortest meeting (5-8) overlaps both others; taking the one that ends first (0-6) leaves room for 7-12."},
			{Name: "none", Stdin: "0\n", Want: "0\n"},
			{Name: "200,000 meetings", Stdin: intervalsInput(), Want: fmt.Sprintln(intervalsWant()),
				Explain: "Sort once with qsort (O(n log n)), then take meetings in a single pass."},
		},
		Hints: []string{
			"Store the meetings in an array of structs { int start, end; } and qsort it by end.",
			"Keep the end of the last meeting taken (start it below every time, such as -1), and take a meeting when its start is >= that end.",
		},
	},
	{
		ID: "edit_distance", Stage: 2, Concept: "Dynamic Programming",
		Title: "Edit distance",
		Task:  "Read lines in pairs until the end of the input, each line up to 2,000 characters (possibly empty). For each pair, print the edit distance: the fewest single-character insertions, deletions and substitutions that turn the first line into the second. The plain recursion is exponential; fill the table of distances between prefixes instead.",
		File:  "edit_distance.c",
		Starter: `/* Checked exercise: edit_distance
 * Input:  lines in pairs: a word or sentence, then the one to compare.
 * Output: one edit distance per pair.
 *   kitten
 *   sitting        prints 3 (k->s, e->i, insert g)
 * d[i][j] = the distance between the first i characters of a and the
 * first j characters of b:
 *   d[i][0] = i, d[0][j] = j,
 *   d[i][j] = min(d[i-1][j] + 1, d[i][j-1] + 1,
 *                 d[i-1][j-1] + (a[i-1] != b[j-1]))
 * Two rows of the table are enough.
 */
`,
		Cases: []CheckCase{
			{Name: "kitten to sitting", Stdin: "kitten\nsitting\n", Want: "3\n"},
			{Name: "several pairs", Stdin: "flaw\nlawn\nsame\nsame\nabc\n\n\nxyz\n", Want: "2\n0\n3\n3\n",
				Explain: "An empty line is a word of length 0: its distance to any word is that word's length."},
			{Name: "a sentence", Stdin: "the quick brown fox\nthe quack brown box\n", Want: "2\n"},
			{Name: "2,000 characters each", Stdin: editDistanceInput(), Want: fmt.Sprintln(editDistanceWant()),
				Explain: "The plain recursion tries every way to align the two lines: exponential. The table has 2,001 × 2,001 cells, each filled once."},
		},
		Hints: []string{
			"Read each line with fgets into a buffer of 2,002 bytes or more, and strip the newline: s[strcspn(s, \"\\n\")] = '\\0'.",
			"A full table of 2,001 × 2,001 ints is 16 MB: too big for the stack. Use two rows (prev and cur) of 2,001 ints, swapping them after each row.",
		},
	},
	{
		ID: "sat_count", Stage: 3, Concept: "Logic & Boolean Algebra",
		Title: "Count the rows that make a formula true",
		Task:  "Read a formula in conjunctive normal form: n variables (1 to 16) and m clauses (0 to 200), each clause a list of literals ended by 0, where k means variable k is true and -k means it is false. A clause is true when at least one of its literals is; the formula is true when every clause is. Print how many of the 2ⁿ rows of the truth table make it true.",
		File:  "sat_count.c",
		Starter: `/* Checked exercise: sat_count
 * Input:  n m, then m clauses, each a list of literals ended by 0.
 *         k means "variable k is true", -k means "variable k is false".
 *   2 2
 *   1 2 0          (x1 OR x2)
 *   -1 -2 0        AND (NOT x1 OR NOT x2)
 * Output: the number of rows (out of 2^n) where every clause is true:
 *   2              (x1 x2 = 1 0 and 0 1: exclusive or)
 * Row r (0 to 2^n - 1) gives variable k the value of bit k - 1 of r.
 * A clause of no literals (just 0) is false in every row.
 */
`,
		Cases: []CheckCase{
			{Name: "exclusive or", Stdin: "2 2\n1 2 0\n-1 -2 0\n", Want: "2\n"},
			{Name: "no clauses", Stdin: "3 0\n", Want: "8\n",
				Explain: "With no clauses every row satisfies the formula: all 2³ = 8 of them."},
			{Name: "a contradiction", Stdin: "1 2\n1 0\n-1 0\n", Want: "0\n"},
			{Name: "an empty clause", Stdin: "2 1\n0\n", Want: "0\n",
				Explain: "A clause is true when at least one literal is true; with no literals it can never be true."},
			{Name: "the last variable", Stdin: "16 1\n16 0\n", Want: "32768\n",
				Explain: "Variables are numbered from 1: variable 16 is bit 15 of the row. Half of the 65,536 rows set it."},
			{Name: "16 variables, 120 clauses", Stdin: satInput(), Want: fmt.Sprintln(satWant())},
		},
		Hints: []string{
			"Store the clauses as one array of literals, with the 0s kept as separators.",
			"For each row r from 0 to (1 << n) - 1: a literal k is true when (r >> (k - 1)) & 1 is 1; a literal -k is true when that bit is 0.",
		},
	},
	{
		ID: "bitsets", Stage: 3, Concept: "Sets, Functions & Relations",
		Title: "Sets as bits",
		Task:  "Read two lines, each a set of numbers from 0 to 63 separated by spaces (a line may be empty, and a number may repeat). Store each set in a uint64_t, one bit per member, and print five lines: the union, the intersection, the first minus the second, and the second minus the first (each in increasing order, separated by spaces, or - for the empty set), then subset if the first set is a subset of the second, or not a subset.",
		File:  "bitsets.c",
		Starter: `/* Checked exercise: bitsets
 * Input:  two lines of numbers from 0 to 63 (a line may be empty).
 *   1 3 5 7
 *   3 4 5
 * Output: union, intersection, A - B, B - A, then subset or not a subset:
 *   1 3 4 5 7
 *   3 5
 *   1 7
 *   4
 *   not a subset
 * Member i is bit i: set |= (uint64_t)1 << i.  (1 << i is an int: wrong
 * from i = 31 on.)  Union |, intersection &, difference a & ~b.
 */
#include <stdint.h>
`,
		Cases: []CheckCase{
			{Name: "the example", Stdin: "1 3 5 7\n3 4 5\n", Want: "1 3 4 5 7\n3 5\n1 7\n4\nnot a subset\n"},
			{Name: "a subset, with repeats", Stdin: "2 2 9\n9 2 30\n", Want: "2 9 30\n2 9\n-\n30\nsubset\n",
				Explain: "Setting a bit twice changes nothing: a set has no duplicates. {2, 9} is inside {2, 9, 30}."},
			{Name: "the empty set", Stdin: "\n5\n", Want: "5\n-\n-\n5\nsubset\n",
				Explain: "The empty set is a subset of every set, and A − B is empty when A is."},
			{Name: "high members", Stdin: "0 31 32 63\n31 63\n", Want: "0 31 32 63\n31 63\n0 32\n-\nnot a subset\n",
				Explain: "1 << 31 and above overflow an int: shift a uint64_t, (uint64_t)1 << i."},
		},
		Hints: []string{
			"Read each line with fgets, then take numbers from it with strtol in a loop, so an empty line is simply a set with no members.",
			"To print a set, test bits 0 to 63 in order: if (s >> i & 1) prints i.",
			"A is a subset of B when (a & b) == a, or equivalently (a & ~b) == 0.",
		},
	},
	{
		ID: "binomial", Stage: 3, Concept: "Counting & Combinatorics",
		Title: "Exact binomial coefficients",
		Task:  "Read lines of two numbers n and k (0 ≤ k ≤ n ≤ 67) until the end of the input, and print C(n, k), the number of ways to choose k items from n, exactly. Every answer fits in an unsigned 64-bit integer, but n! does not from n = 21, and multiplying before dividing overflows near n = 67. Build Pascal's triangle with additions.",
		File:  "binomial.c",
		Starter: `/* Checked exercise: binomial
 * Input:  lines "n k", 0 <= k <= n <= 67.
 * Output: C(n, k) exactly, one per line.
 *   5 2            prints 10
 * C(67, 33) = 14226520737620288370 is the largest answer: it fits in an
 * unsigned long long (print it with %llu), but 21! does not.
 * Pascal's triangle: C(n, 0) = C(n, n) = 1 and
 *   C(n, k) = C(n - 1, k - 1) + C(n - 1, k)
 * Only additions, and no sum ever exceeds the answer.
 */
`,
		Cases: []CheckCase{
			{Name: "small", Stdin: "5 2\n4 0\n4 4\n6 3\n", Want: "10\n1\n1\n20\n"},
			{Name: "the lottery", Stdin: "49 6\n", Want: "13983816\n"},
			{Name: "just past 20!", Stdin: "21 1\n21 10\n", Want: "21\n352716\n",
				Explain: "21! does not fit in 64 bits, and unsigned arithmetic wraps silently: n! / (k!(n − k)!) gives nonsense from here on."},
			{Name: "the largest", Stdin: "67 33\n67 34\n66 33\n", Want: "14226520737620288370\n14226520737620288370\n7219428434016265740\n",
				Explain: "c = c * (n − k + i) / i overflows before it divides here: the product is about 4.8 × 10²⁰. Pascal's triangle only adds."},
			{Name: "a whole row", Stdin: binomialRowInput(60), Want: binomialRowWant(60)},
		},
		Hints: []string{
			"A static unsigned long long table[68][68], filled once at the start, answers every line by lookup.",
			"Fill row n from row n − 1: table[n][0] = table[n][n] = 1, and table[n][k] = table[n-1][k-1] + table[n-1][k] in between.",
		},
	},
	{
		ID: "modpow", Stage: 3, Concept: "Modular Arithmetic & Number Theory",
		Title: "Fast modular power",
		Task:  "Read lines of three numbers a, e and m (0 ≤ a, e ≤ 10¹⁸, 1 ≤ m < 2³²) until the end of the input, and print aᵉ mod m for each. Multiplying a by itself e times is far too slow for e = 10¹⁸; square and multiply instead, reducing mod m at every step so that every product fits in 64 bits.",
		File:  "modpow.c",
		Starter: `/* Checked exercise: modpow
 * Input:  lines "a e m": 0 <= a, e <= 10^18, 1 <= m < 2^32.
 * Output: a^e mod m, one per line.
 *   2 10 1000      prints 24 (1024 mod 1000)
 * Read and compute with uint64_t (scanf with "%llu" into unsigned long
 * long works too). Reduce a mod m first: then every product of two
 * remainders is below 2^64.
 * Square and multiply: while e > 0, if e is odd multiply the result by
 * a; square a; halve e.
 */
#include <stdint.h>
`,
		Cases: []CheckCase{
			{Name: "small powers", Stdin: "2 10 1000\n3 4 5\n7 0 13\n", Want: "24\n1\n1\n",
				Explain: "Anything to the power 0 is 1."},
			{Name: "mod 1", Stdin: "5 0 1\n123 456 1\n", Want: "0\n0\n",
				Explain: "Every number is 0 mod 1, even a⁰ = 1: reduce the result mod m too."},
			{Name: "huge exponents", Stdin: "2 1000000000000000000 1000000007\n3 999999999999999999 4294967291\n", Want: modpowWant([][3]uint64{{2, 1000000000000000000, 1000000007}, {3, 999999999999999999, 4294967291}}),
				Explain: "10¹⁸ multiplications would take decades: squaring needs about 60 steps for an exponent this size."},
			{Name: "a bigger than m", Stdin: "1000000000000000000 5 4294967291\n4294967290 2 4294967291\n", Want: modpowWant([][3]uint64{{1000000000000000000, 5, 4294967291}, {4294967290, 2, 4294967291}}),
				Explain: "Reduce a mod m before multiplying: 10¹⁸ squared does not fit in 64 bits, but (10¹⁸ mod m) squared does."},
		},
		Hints: []string{
			"uint64_t result = 1 % m; a %= m; then loop while e > 0.",
			"(m − 1)² is below 2⁶⁴ when m < 2³², so a * a % m and result * a % m are exact in uint64_t.",
		},
	},
	{
		ID: "inversions", Stage: 3, Concept: "Recurrences: The Cost of Recursion",
		Title: "Count the inversions",
		Task:  "Read n (0 to 200,000) and n integers, and print the number of inversions: pairs of positions i < j whose values are out of order (a[i] > a[j]). Checking every pair is O(n²). Adapt merge sort, whose recurrence T(n) = 2T(n/2) + n gives O(n log n): while merging, each item taken from the right half jumps ahead of every item still waiting in the left half. The count can exceed 2 billion.",
		File:  "inversions.c",
		Starter: `/* Checked exercise: inversions
 * Input:  n, then n integers.
 * Output: the number of pairs i < j with a[i] > a[j].
 *   5
 *   2 4 1 3 5      prints 3: (2,1), (4,1) and (4,3)
 * Merge sort, counting as you merge: when the next smallest item comes
 * from the RIGHT half, it was behind every item left in the LEFT half,
 * so add (number of items left in the left half).
 * 200,000 numbers in reverse order have about 2 * 10^10 inversions: use
 * long long.
 */
`,
		Cases: []CheckCase{
			{Name: "the example", Stdin: "5\n2 4 1 3 5\n", Want: "3\n"},
			{Name: "sorted and empty", Stdin: "4\n1 2 3 4\n", Want: "0\n"},
			{Name: "nothing", Stdin: "0\n", Want: "0\n"},
			{Name: "equal values", Stdin: "4\n2 2 1 1\n", Want: "4\n",
				Explain: "Equal values are not out of order: only a[i] > a[j] counts. When merging, take from the left half on a tie."},
			{Name: "200,000 in reverse", Stdin: reverseInput(200000), Want: "19999900000\n",
				Explain: "n(n − 1)/2 inversions: about 2 × 10¹⁰, more than an int can hold."},
			{Name: "200,000 random", Stdin: inversionsInput(), Want: fmt.Sprintln(inversionsWant()),
				Explain: "Comparing every pair is 2 × 10¹⁰ comparisons; merge sort needs about 3.5 million."},
		},
		Hints: []string{
			"Write long long sort_count(int *a, int *tmp, int n): sort and count the left half, the right half, then merge into tmp, counting, and copy back.",
			"On a tie take the left item first; then only strictly smaller right items count as inversions.",
		},
	},
	{
		ID: "euler", Stage: 3, Concept: "Graph Theory: Degrees, Trees & Paths",
		Title: "Can you draw it without lifting the pen?",
		Task:  "Read an undirected graph: n vertices (1 to 100,000, numbered from 1) and m edges (0 to 200,000) as pairs u v, possibly repeated, never from a vertex to itself. Print circuit if a walk can use every edge exactly once and end where it started, path u v (the two ends, smaller first) if it can only do so between two different vertices, or none. The edges must all be connected; vertices with no edges do not matter. With no edges, print circuit.",
		File:  "euler.c",
		Starter: `/* Checked exercise: euler
 * Input:  n m, then m lines "u v" (1 <= u, v <= n, u != v).
 * Output: circuit, path u v, or none.
 * Euler (1736): if the edges are connected,
 *   every vertex has even degree        -> circuit
 *   exactly two have odd degree         -> path between those two
 *   otherwise                           -> none
 * If the edges are in two or more separate pieces -> none.
 * Count degrees in an array; check connectivity with a BFS (or a
 * union-find) from any vertex that has an edge.
 */
`,
		Cases: []CheckCase{
			{Name: "a triangle", Stdin: "3 3\n1 2\n2 3\n3 1\n", Want: "circuit\n"},
			{Name: "the house", Stdin: "5 8\n1 2\n2 3\n3 4\n4 1\n1 3\n2 4\n3 5\n4 5\n", Want: "path 1 2\n",
				Explain: "Vertices 1 and 2 have degree 3, every other vertex an even degree: start at one odd vertex and finish at the other."},
			{Name: "Königsberg", Stdin: "4 7\n1 2\n1 2\n1 3\n1 3\n1 4\n2 4\n3 4\n", Want: "none\n",
				Explain: "Four vertices of odd degree: no walk crosses every bridge once. Repeated edges count once each."},
			{Name: "two islands", Stdin: "6 6\n1 2\n2 3\n3 1\n4 5\n5 6\n6 4\n", Want: "none\n",
				Explain: "Every degree is even, but the edges form two separate triangles: one walk cannot cover both. Check connectivity too."},
			{Name: "lonely vertices", Stdin: "10 2\n3 7\n7 9\n", Want: "path 3 9\n",
				Explain: "Vertices with no edges do not stop the drawing: check connectivity only among vertices that have edges."},
			{Name: "no edges", Stdin: "3 0\n", Want: "circuit\n"},
			{Name: "a big ring", Stdin: eulerRingInput(100000), Want: "circuit\n",
				Explain: "A recursive depth-first search 100,000 calls deep can overflow the stack: use a queue (BFS) or union-find."},
		},
		Hints: []string{
			"Store the edges in two arrays and build adjacency lists with an offset array (count each vertex's degree, take prefix sums, then fill), or use union-find on the edges directly.",
			"Union-find: parent[v] = v at first; find follows parents to the root; for each edge, union the two roots. Connected means every vertex with an edge has the same root.",
		},
	},
	{
		ID: "birthday", Stage: 3, Concept: "Probability for Programmers",
		Title: "The birthday problem, exactly",
		Task:  "Read lines with a number k (0 to 1000) until the end of the input, and for each print the probability, with 6 decimals, that at least two of k people share a birthday, assuming 365 equally likely birthdays. Compute the chance that all k birthdays differ, 365/365 × 364/365 × …, and take it from 1.",
		File:  "birthday.c",
		Starter: `/* Checked exercise: birthday
 * Input:  lines with k, 0 <= k <= 1000.
 * Output: P(at least two of k share a birthday), printf("%.6f\n", p).
 *   23             prints 0.507297
 * P(all different) = (365/365) * (364/365) * ... * ((365 - k + 1)/365)
 * Careful: 364 / 365 in C is integer division, which is 0.
 * With more than 365 people, two must share (the pigeonhole principle).
 */
`,
		Cases: []CheckCase{
			{Name: "twenty-three people", Stdin: "23\n", Want: "0.507297\n",
				Explain: "364 / 365 is 0 in integer arithmetic: write 364.0 / 365, or use doubles throughout."},
			{Name: "nobody and one person", Stdin: "0\n1\n", Want: "0.000000\n0.000000\n"},
			{Name: "a few more", Stdin: "2\n10\n30\n41\n50\n70\n", Want: "0.002740\n0.116948\n0.706316\n0.903152\n0.970374\n0.999160\n"},
			{Name: "more than 365", Stdin: "100\n366\n1000\n", Want: "1.000000\n1.000000\n1.000000\n",
				Explain: "With 366 people two must share a birthday (the pigeonhole principle); a negative factor such as (365 − 366)/365 must never enter the product."},
		},
		Hints: []string{
			"double all_different = 1.0; for (int i = 0; i < k && i < 365; i++) all_different *= (365.0 - i) / 365.0;",
			"For k > 365 the answer is exactly 1: stop the loop at 365, where the factor reaches 0.",
		},
	},
	{
		ID: "bases", Stage: 4, Concept: "Binary & Number Representation",
		Title: "Decimal, hex and binary",
		Task:  "Read lines, each a number from 0 to 4,294,967,295 written in decimal (42), hexadecimal (0x2A or 0x2a) or binary (0b101010), and print it in all three: 42 0x2a 0b101010, with lowercase hex digits and no leading zeros (0 is 0 0x0 0b0). Parse and print binary yourself: printf has %x but no binary format.",
		File:  "bases.c",
		Starter: `/* Checked exercise: bases
 * Input:  lines with a number: 42, 0x2A or 0b101010 (up to 2^32 - 1).
 * Output: decimal, hex and binary on one line:
 *   42 0x2a 0b101010
 * Hold the value in a uint32_t (or unsigned long long while parsing).
 * Parsing: value = value * base + digit, for base 10, 16 or 2.
 * Printing binary: find the highest 1 bit, then print every bit from it
 * down to bit 0. Zero is the special case: 0b0.
 */
#include <stdint.h>
`,
		Cases: []CheckCase{
			{Name: "the same number three ways", Stdin: "42\n0x2A\n0b101010\n", Want: "42 0x2a 0b101010\n42 0x2a 0b101010\n42 0x2a 0b101010\n",
				Explain: "Hex digits A to F may be upper or lower case on input; print them in lower case."},
			{Name: "zero and one", Stdin: "0\n0b1\n0x0\n", Want: "0 0x0 0b0\n1 0x1 0b1\n0 0x0 0b0\n",
				Explain: "Zero has no highest 1 bit: print it as 0b0, not as an empty string."},
			{Name: "powers of two", Stdin: "256\n0x80000000\n", Want: "256 0x100 0b100000000\n2147483648 0x80000000 0b10000000000000000000000000000000\n"},
			{Name: "the largest", Stdin: "4294967295\n0xffffffff\n", Want: "4294967295 0xffffffff 0b11111111111111111111111111111111\n4294967295 0xffffffff 0b11111111111111111111111111111111\n",
				Explain: "4294967295 does not fit in an int (it prints as -1): use uint32_t, and %u or PRIu32 to print it."},
		},
		Hints: []string{
			"Read each line with fgets, check whether it starts with 0x/0X or 0b/0B, and choose the base; then loop over the remaining characters.",
			"A hex digit's value: c - '0' for 0–9, tolower(c) - 'a' + 10 for a–f.",
			"For binary output, i from 31 down to 0: skip leading zeros until the first 1 bit, then print every bit.",
		},
	},
	{
		ID: "nand_gates", Stage: 4, Concept: "Logic Gates & Circuits",
		Title: "Everything from NAND",
		Task:  "The checker gives you one gate, int nand(int a, int b), and counts how often it is called. Using only nand (no &, |, ^, !, ~ or arithmetic on the bits), write int not_(int a), and_(int a, int b), or_(int a, int b) and xor_(int a, int b), with the fewest possible NAND gates (1, 2, 3 and 4), and a full adder int full_add(int a, int b, int cin, int *cout) built from your gates. The checker adds every pair of 8-bit numbers with your full adder. No main.",
		File:  "nand_gates.c",
		Starter: `/* Checked exercise: nand_gates
 * The checker provides:   int nand(int a, int b);
 * Write, using only nand() and your own gates (inputs and outputs are 0 or 1):
 *   int not_(int a);                 1 NAND
 *   int and_(int a, int b);          2 NANDs
 *   int or_(int a, int b);           3 NANDs
 *   int xor_(int a, int b);          4 NANDs
 *   int full_add(int a, int b, int cin, int *cout);
 *       returns the sum bit and sets *cout to the carry out
 * No main.
 */
int nand(int a, int b);
`,
		Harness: `#include <stdio.h>
#include <string.h>

static long calls;
int nand(int a, int b) { calls++; return !(a && b); }

int not_(int a);
int and_(int a, int b);
int or_(int a, int b);
int xor_(int a, int b);
int full_add(int a, int b, int cin, int *cout);

static void table(const char *name, int (*gate)(int, int)) {
    long most = 0;
    for (int a = 0; a <= 1; a++)
        for (int b = 0; b <= 1; b++) {
            calls = 0;
            int out = gate(a, b);
            if (calls > most) most = calls;
            printf("%s(%d, %d) = %d\n", name, a, b, out);
        }
    printf("NAND gates: %ld\n", most);
}

static int not_2(int a, int b) { (void)b; return not_(a); }

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "not") == 0) {
        long most = 0;
        for (int a = 0; a <= 1; a++) {
            calls = 0;
            printf("not(%d) = %d\n", a, not_2(a, 0));
            if (calls > most) most = calls;
        }
        printf("NAND gates: %ld\n", most);
    } else if (strcmp(argv[1], "and") == 0) table("and", and_);
    else if (strcmp(argv[1], "or") == 0) table("or", or_);
    else if (strcmp(argv[1], "xor") == 0) table("xor", xor_);
    else if (strcmp(argv[1], "adder") == 0) {
        for (int x = 0; x < 256; x++)
            for (int y = 0; y < 256; y++) {
                int carry = 0, sum = 0;
                calls = 0;
                for (int i = 0; i < 8; i++) {
                    int c;
                    sum |= full_add(x >> i & 1, y >> i & 1, carry, &c) << i;
                    carry = c;
                }
                sum |= carry << 8;
                if (sum != x + y) {
                    printf("%d + %d gave %d\n", x, y, sum);
                    return 0;
                }
                if (calls == 0) {
                    printf("the adder never called nand\n");
                    return 0;
                }
            }
        printf("all 65536 sums correct\n");
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "not", Args: []string{"not"}, Want: "not(0) = 1\nnot(1) = 0\nNAND gates: 1\n",
				Explain: "A NAND with both inputs tied together is a NOT: nand(a, a)."},
			{Name: "and", Args: []string{"and"}, Want: "and(0, 0) = 0\nand(0, 1) = 0\nand(1, 0) = 0\nand(1, 1) = 1\nNAND gates: 2\n",
				Explain: "AND is NOT NAND: nand(nand(a, b), nand(a, b)) uses 3 gates; reuse the inner result instead: x = nand(a, b); nand(x, x)."},
			{Name: "or", Args: []string{"or"}, Want: "or(0, 0) = 0\nor(0, 1) = 1\nor(1, 0) = 1\nor(1, 1) = 1\nNAND gates: 3\n",
				Explain: "De Morgan: a OR b = NOT(NOT a AND NOT b) = nand(not a, not b)."},
			{Name: "xor", Args: []string{"xor"}, Want: "xor(0, 0) = 0\nxor(0, 1) = 1\nxor(1, 0) = 1\nxor(1, 1) = 0\nNAND gates: 4\n",
				Explain: "Share the first gate: m = nand(a, b); then nand(nand(a, m), nand(b, m)). Building it from AND, OR and NOT takes more."},
			{Name: "an 8-bit ripple-carry adder", Args: []string{"adder"}, Want: "all 65536 sums correct\n",
				Explain: "sum = a XOR b XOR cin; carry out = (a AND b) OR (cin AND (a XOR b))."},
		},
		Hints: []string{
			"Write each gate on paper as a little circuit of NANDs first, and count them.",
			"The full adder needs no new tricks: two half adders (an XOR and an AND each) and an OR for the two carries.",
		},
	},
	{
		ID: "alu", Stage: 4, Concept: "Building an ALU: Adders, Subtraction & Flags",
		Title: "An ALU with flags",
		Task:  "Write uint32_t alu(uint32_t a, uint32_t b, char op, struct flags *f) for op '+', '-', '&', '|' and '^', setting f->z (the result is 0), f->n (bit 31 of the result is 1), f->c (for +: a carry out of bit 31; for -: a borrow, that is a < b as unsigned numbers) and f->v (signed overflow: the result as a signed number is wrong). For &, | and ^, c and v are 0. Compute the flags from the bits, as the hardware does; no main.",
		File:  "alu.c",
		Starter: `/* Checked exercise: alu
 * struct flags as below; write alu(). No main.
 *   op '+': a + b    op '-': a - b    '&' '|' '^': bitwise
 *   z: result == 0          n: bit 31 of the result
 *   c: '+' carry out of bit 31; '-' borrow (a < b unsigned); else 0
 *   v: '+' a and b have the same sign bit and the result's differs
 *      '-' a and b have different sign bits and the result's differs
 *          from a's;  else 0
 * Unsigned arithmetic wraps around, so a + b and a - b are always
 * defined on uint32_t.
 */
#include <stdint.h>

struct flags { int z, n, c, v; };

uint32_t alu(uint32_t a, uint32_t b, char op, struct flags *f);
`,
		Harness: `#include <stdint.h>
#include <stdio.h>
#include <string.h>

struct flags { int z, n, c, v; };
uint32_t alu(uint32_t a, uint32_t b, char op, struct flags *f);

static void show(uint32_t a, char op, uint32_t b) {
    struct flags f = {9, 9, 9, 9};
    uint32_t r = alu(a, b, op, &f);
    printf("%08x %c %08x = %08x  Z%d N%d C%d V%d\n", (unsigned)a, op, (unsigned)b, (unsigned)r, f.z, f.n, f.c, f.v);
}

/* The flags computed the long way, with 64-bit arithmetic. */
static struct flags expect(uint32_t a, uint32_t b, char op, uint32_t *r) {
    struct flags f = {0, 0, 0, 0};
    int64_t sa = (int32_t)a, sb = (int32_t)b, s = 0;
    switch (op) {
    case '+': *r = a + b; f.c = ((uint64_t)a + b) >> 32; s = sa + sb; f.v = s != (int32_t)*r; break;
    case '-': *r = a - b; f.c = a < b; s = sa - sb; f.v = s != (int32_t)*r; break;
    case '&': *r = a & b; break;
    case '|': *r = a | b; break;
    default:  *r = a ^ b; break;
    }
    f.z = *r == 0;
    f.n = *r >> 31;
    return f;
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "add") == 0) {
        show(2, '+', 3);
        show(0x7fffffff, '+', 1);
        show(0xffffffff, '+', 1);
        show(0x80000000, '+', 0x80000000);
    } else if (strcmp(argv[1], "sub") == 0) {
        show(7, '-', 5);
        show(5, '-', 7);
        show(0x80000000, '-', 1);
        show(0, '-', 0);
        show(0x7fffffff, '-', 0xffffffff);
    } else if (strcmp(argv[1], "logic") == 0) {
        show(0xf0f0f0f0, '&', 0x0f0f0f0f);
        show(0x80000000, '|', 1);
        show(0xffffffff, '^', 0x0000ffff);
    } else if (strcmp(argv[1], "random") == 0) {
        uint64_t x = 88172645463325252u;
        const char *ops = "+-&|^";
        for (int i = 0; i < 1000000; i++) {
            x ^= x << 13; x ^= x >> 7; x ^= x << 17;
            uint32_t a = (uint32_t)x, b = (uint32_t)(x >> 32), want;
            if (i % 4 == 0) b = a ^ (b & 0x80000001u);   /* near-equal values */
            char op = ops[i % 5];
            struct flags g = {9, 9, 9, 9}, e = expect(a, b, op, &want);
            uint32_t got = alu(a, b, op, &g);
            if (got != want || g.z != e.z || g.n != e.n || g.c != e.c || g.v != e.v) {
                printf("wrong: ");
                show(a, op, b);
                printf("want   %08x %c %08x = %08x  Z%d N%d C%d V%d\n", (unsigned)a, op, (unsigned)b, (unsigned)want, e.z, e.n, e.c, e.v);
                return 0;
            }
        }
        printf("1000000 operations correct\n");
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "addition", Args: []string{"add"}, Want: "00000002 + 00000003 = 00000005  Z0 N0 C0 V0\n7fffffff + 00000001 = 80000000  Z0 N1 C0 V1\nffffffff + 00000001 = 00000000  Z1 N0 C1 V0\n80000000 + 80000000 = 00000000  Z1 N0 C1 V1\n",
				Explain: "Carry is about unsigned numbers (ffffffff + 1 wraps to 0); overflow is about signed ones (7fffffff + 1, the largest int plus one, becomes negative)."},
			{Name: "subtraction", Args: []string{"sub"}, Want: "00000007 - 00000005 = 00000002  Z0 N0 C0 V0\n00000005 - 00000007 = fffffffe  Z0 N1 C1 V0\n80000000 - 00000001 = 7fffffff  Z0 N0 C0 V1\n00000000 - 00000000 = 00000000  Z1 N0 C0 V0\n7fffffff - ffffffff = 80000000  Z0 N1 C1 V1\n",
				Explain: "For subtraction, c is the borrow: 1 when a < b as unsigned numbers. 5 − 7 borrows (c = 1) but is a correct signed result (−2), so v = 0."},
			{Name: "logic", Args: []string{"logic"}, Want: "f0f0f0f0 & 0f0f0f0f = 00000000  Z1 N0 C0 V0\n80000000 | 00000001 = 80000001  Z0 N1 C0 V0\nffffffff ^ 0000ffff = ffff0000  Z0 N1 C0 V0\n",
				Explain: "Logic operations clear c and v, and set z and n from the result like any other."},
			{Name: "a million random operations", Args: []string{"random"}, Want: "1000000 operations correct\n"},
		},
		Hints: []string{
			"Carry for +: the result is smaller than a (r < a) exactly when the addition wrapped around.",
			"Overflow for +: ((a ^ r) & (b ^ r)) >> 31: the result's sign differs from both inputs' signs. For −: ((a ^ b) & (a ^ r)) >> 31.",
			"Set every flag on every call: the checker fills them with 9 first.",
		},
	},
	{
		ID: "number_fsm", Stage: 4, Concept: "Finite State Machines & Sequential Logic",
		Title: "Validate numbers with a state machine",
		Task:  "Read lines (up to 10,000 characters each) and print valid or invalid for each, according to whether the whole line is a decimal number: an optional sign, then digits with an optional fractional part (12, 12., 12.5) or a dot and digits (.5), then optionally e or E, an optional sign, and at least one digit. Nothing else is allowed, not even spaces. Write it as a state machine: an enum of states and a switch.",
		File:  "number_fsm.c",
		Starter: `/* Checked exercise: number_fsm
 * Input:  lines; output valid or invalid for each.
 * Valid:   0  -12  +3.14  .5  5.  1e10  2E-3  6.02e+23
 * Invalid: (empty)  .  e5  1e  1e+  +-1  1.2.3  12a  " 1"  1e2.5
 * Suggested states:
 *   START, SIGN, INT (digits), DOT_NO_INT (a "." with no digits before),
 *   FRAC (after the "."), EXP (after e/E), EXP_SIGN, EXP_INT, ERROR
 * Accepting states: INT, FRAC, EXP_INT.
 * Draw the arrows first: which characters move from each state where?
 */
`,
		Cases: []CheckCase{
			{Name: "plain numbers", Stdin: "0\n-12\n+3.14\n.5\n5.\n", Want: "valid\nvalid\nvalid\nvalid\nvalid\n",
				Explain: `"5." and ".5" are both numbers, but "." alone is not: a dot needs a digit on at least one side.`},
			{Name: "exponents", Stdin: "1e10\n2E-3\n6.02e+23\n1e\n1e+\ne5\n1e2.5\n", Want: "valid\nvalid\nvalid\ninvalid\ninvalid\ninvalid\ninvalid\n",
				Explain: "After e there must be at least one digit (a sign alone is not enough), and the exponent is a whole number: no dot."},
			{Name: "not numbers", Stdin: "\n.\n+-1\n1.2.3\n12a\n 1\n1 \n-\n+.\n", Want: "invalid\ninvalid\ninvalid\ninvalid\ninvalid\ninvalid\ninvalid\ninvalid\ninvalid\n",
				Explain: "An empty line, a lone sign or dot, two signs, two dots, letters and spaces are all invalid: only accepting states at the end of the line mean valid."},
			{Name: "edge cases", Stdin: "+.5\n-0.0e-0\n00012\n.e1\n", Want: "valid\nvalid\nvalid\ninvalid\n"},
			{Name: "many lines", Stdin: numberFSMInput(), Want: numberFSMWant(),
				Explain: "Every one of these lines was generated from the rules; compare the first mismatch with your arrows."},
		},
		Hints: []string{
			"Strip the newline (s[strcspn(s, \"\\n\")] = '\\0'), run the machine over every character, then check the final state.",
			"From START: a sign goes to SIGN, a digit to INT, a dot to DOT_NO_INT, anything else to ERROR. From INT: digits stay, a dot goes to FRAC, e/E goes to EXP.",
			"Once in ERROR, stay there: the rest of the line cannot make it valid.",
		},
	},
	{
		ID: "cpu", Stage: 4, Concept: "The CPU: Datapath & Instruction Cycle",
		Title: "Emulate a tiny CPU",
		Task:  "Read a program for a tiny CPU from standard input, one instruction per line, numbered from 0, and run it. Registers r0 to r3 start at 0 and hold 64-bit signed integers. Instructions: LOAD r n (r = n), ADD r s (r = r + s), SUB r s (r = r − s), JZ r t (if r is 0, continue at instruction t), JMP t, PRINT r (print the value on its own line) and HALT. Running past the last instruction also stops. If the program runs more than 1,000,000 instructions, print too many steps and stop.",
		File:  "cpu.c",
		Starter: `/* Checked exercise: cpu
 * A program (up to 1000 lines), for example a countdown from 3:
 *   LOAD r0 3          0
 *   LOAD r1 1          1
 *   PRINT r0           2
 *   SUB r0 r1          3
 *   JZ r0 6            4
 *   JMP 2              5
 *   HALT               6      prints 3, 2 and 1
 * Fetch the instruction at pc, decode it (a switch on the opcode),
 * execute it, and move pc on: pc + 1, or the jump target.
 * Parse every line into an array of structs before running.
 */
`,
		Cases: []CheckCase{
			{Name: "count down", Stdin: "LOAD r0 3\nLOAD r1 1\nPRINT r0\nSUB r0 r1\nJZ r0 6\nJMP 2\nHALT\n", Want: "3\n2\n1\n"},
			{Name: "sum 1 to 100", Stdin: cpuSumProgram, Want: "5050\n",
				Explain: "A jump goes to the instruction with that number, counting from 0."},
			{Name: "fibonacci", Stdin: cpuFibProgram, Want: "0\n1\n1\n2\n3\n5\n8\n13\n21\n34\n",
				Explain: "ADD and SUB take two registers: ADD r2 r1 means r2 = r2 + r1."},
			{Name: "fall off the end", Stdin: "LOAD r3 -7\nPRINT r3\nPRINT r2\n", Want: "-7\n0\n",
				Explain: "Registers start at 0, and running past the last instruction stops the program."},
			{Name: "an endless loop", Stdin: "LOAD r0 1\nJMP 0\n", Want: "too many steps\n",
				Explain: "Count every instruction executed, and stop after 1,000,000."},
			{Name: "big numbers", Stdin: "LOAD r0 4000000000\nLOAD r1 4000000000\nADD r0 r1\nPRINT r0\nHALT\n", Want: "8000000000\n",
				Explain: "Registers are 64-bit: use long long, read with strtoll or %lld."},
		},
		Hints: []string{
			"struct insn { int op; int a; long long b; }: for each line, sscanf the opcode word, then its operands; turn \"r2\" into 2 by skipping the r.",
			"while (pc >= 0 && pc < count) { if (++steps > 1000000) { too many steps } switch (code[pc].op) { … } }",
		},
	},
	{
		ID: "rv_decode", Stage: 4, Concept: "Instruction Sets & Machine Code",
		Title: "Disassemble RISC-V",
		Task:  "Read lines, each a 32-bit RISC-V instruction as 8 hex digits (such as 00500093), and print it as assembly with numbered registers and decimal immediates: add, sub, and, or and xor as add x3, x1, x2; addi, andi, ori and xori as addi x1, x0, 5. Print unknown for any other instruction. The immediate is a signed 12-bit number in bits 20–31.",
		File:  "rv_decode.c",
		Starter: `/* Checked exercise: rv_decode
 * Input:  lines of 8 hex digits, one 32-bit instruction each.
 * Output: one line of assembly each, or "unknown":
 *   00500093       addi x1, x0, 5
 *   fff08113       addi x2, x1, -1
 *   402081b3       sub x3, x1, x2
 * Fields:  opcode = w & 0x7f      rd  = (w >> 7) & 0x1f
 *          funct3 = (w >> 12) & 7 rs1 = (w >> 15) & 0x1f
 *          rs2 = (w >> 20) & 0x1f funct7 = w >> 25
 *          imm = the top 12 bits, as a SIGNED number
 * opcode 0x33 (R-type): funct3 0 add (funct7 0) or sub (funct7 0x20),
 *   4 xor, 6 or, 7 and (funct7 0)
 * opcode 0x13 (I-type): funct3 0 addi, 4 xori, 6 ori, 7 andi
 */
#include <stdint.h>
`,
		Cases: []CheckCase{
			{Name: "assembled by LLVM", Stdin: "00500093\nfff08113\n002081b3\n402081b3\n00c5f533\n00c5e533\n01df4fb3\n",
				Want:    "addi x1, x0, 5\naddi x2, x1, -1\nadd x3, x1, x2\nsub x3, x1, x2\nand x10, x11, x12\nor x10, x11, x12\nxor x31, x30, x29\n",
				Explain: "These words came from llvm-mc, LLVM's own assembler. add and sub differ only in bit 30, inside funct7."},
			{Name: "immediates at the limits", Stdin: "7ff30293\n80030293\n0ff47393\nff046393\nfff44393\n00000013\n",
				Want:    "addi x5, x6, 2047\naddi x5, x6, -2048\nandi x7, x8, 255\nori x7, x8, -16\nxori x7, x8, -1\naddi x0, x0, 0\n",
				Explain: "The immediate is signed: 0x800 in the top 12 bits is −2048, not 2048. Shift the word as an int32_t, or subtract 4096 when bit 31 is set."},
			{Name: "not ours", Stdin: "00311093\n00000000\n42008033\nffffffff\n",
				Want:    "unknown\nunknown\nunknown\nunknown\n",
				Explain: "slli (funct3 1), an all-zero word, an R-type with a funct7 that is neither 0 nor 0x20, and all ones: anything outside the nine instructions is unknown."},
			{Name: "a thousand instructions", Stdin: rvInput(), Want: rvWant()},
		},
		Hints: []string{
			"Read each line with scanf(\"%x\", &w) into an unsigned int, or strtoul(line, NULL, 16).",
			"Decide by opcode first, then funct3, then (for R-type) funct7; every combination you do not handle falls through to unknown.",
		},
	},
	{
		ID: "linkage", Stage: 5, Concept: "Linking, Loading & Shared Libraries",
		Title: "Keep your names to yourself",
		Task:  "Write a tick counter: void tick(void) adds one, int ticks(void) returns the count, void reset(void) sets it back to 0. The checker links your file with its own, which already has a global variable named count and a function named total, so every other name in your file must be private to it. No main.",
		File:  "linkage.c",
		// Some toolchains (Apple's among them) still merge an uninitialised
		// global with another file's definition as a "common" symbol, hiding
		// the clash; -fno-common makes every platform report it, as C requires.
		Flags: []string{"-fno-common"},
		Starter: `/* Checked exercise: linkage
 * Write these three functions; no main:
 *   void tick(void);    add one to the counter
 *   int  ticks(void);   the counter's value
 *   void reset(void);   back to 0
 * The checker's own file, linked with yours, already defines
 *   int count;          and          int total(void);
 * A global name defined in two files is a link error ("multiple
 * definition"). Anything the outside world does not need to see should
 * be static: private to this file.
 */
`,
		Harness: `#include <stdio.h>
#include <string.h>

int count = 100;               /* the checker's own count */
int total(void) { return 7; }  /* and its own total */

void tick(void);
int ticks(void);
void reset(void);

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "three") == 0) {
        tick(); tick(); tick();
        printf("%d\n", ticks());
    } else if (strcmp(argv[1], "reset") == 0) {
        tick(); tick(); reset(); tick();
        printf("%d\n", ticks());
    } else if (strcmp(argv[1], "theirs") == 0) {
        tick();
        printf("%d %d %d\n", ticks(), count, total());
    } else return 2;
    return 0;
}
`,
		BuildHints: []BuildHint{
			{"multiple definition of", "Your file and the checker's both define the same global name. Make yours static (static int count;): it becomes private to your file, and the clash disappears."},
			{"duplicate symbol", "Your file and the checker's both define the same global name. Make yours static (static int count;): it becomes private to your file, and the clash disappears."},
			{"undefined reference", "The linker found no definition: define tick, ticks and reset, spelled exactly like that, and do not make those three static."},
			{"Undefined symbols", "The linker found no definition: define tick, ticks and reset, spelled exactly like that, and do not make those three static."},
		},
		Cases: []CheckCase{
			{Name: "three ticks", Args: []string{"three"}, Want: "3\n"},
			{Name: "reset", Args: []string{"reset"}, Want: "1\n"},
			{Name: "the checker's names are untouched", Args: []string{"theirs"}, Want: "1 100 7\n",
				Explain: "The checker's own count must still be 100: your counter is a different variable, private to your file."},
		},
		Hints: []string{
			"static int count; at the top of your file is a variable only your file can see; the checker's count is a different variable.",
			"Helper functions can be static too: static void bump(int by) { … }. Only tick, ticks and reset must be visible to the checker.",
		},
	},
	{
		ID: "checked_add", Stage: 5, Concept: "What the Optimizer Does",
		Title: "Overflow checks that cannot overflow",
		Task:  "Write int checked_add(int a, int b, int *sum) and int checked_mul(int a, int b, int *product): each returns 1 and stores the result when it fits in an int, or returns 0 and leaves the output alone when it would overflow. The check itself must not overflow: signed overflow is undefined behaviour, the undefined-behaviour sanitizer stops the program the moment it happens, and the optimizer is allowed to delete a check written after it. No main.",
		File:  "checked_add.c",
		Starter: `/* Checked exercise: checked_add
 *   int checked_add(int a, int b, int *sum);
 *   int checked_mul(int a, int b, int *product);
 * Return 1 and store the result if it fits in an int; return 0 and
 * leave *sum or *product unchanged if it would overflow.
 * Test BEFORE computing:  a + b overflows when b > 0 && a > INT_MAX - b,
 * or when b < 0 && a < INT_MIN - b. Computing a + b first and looking at
 * the result is too late: the overflow itself is undefined behaviour.
 * (__builtin_add_overflow and __builtin_mul_overflow, in gcc and clang,
 * are allowed too.)
 */
#include <limits.h>
`,
		Harness: `#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

int checked_add(int a, int b, int *sum);
int checked_mul(int a, int b, int *product);

static void add(int a, int b) {
    int r = 12345;
    int ok = checked_add(a, b, &r);
    printf("%d + %d: %s %d\n", a, b, ok ? "ok" : "overflow", r);
}

static void mul(int a, int b) {
    int r = 12345;
    int ok = checked_mul(a, b, &r);
    printf("%d * %d: %s %d\n", a, b, ok ? "ok" : "overflow", r);
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "add") == 0) {
        add(1, 2); add(INT_MAX, 1); add(INT_MIN, -1); add(INT_MAX, INT_MIN);
        add(-5, 5); add(INT_MAX, 0); add(INT_MIN, INT_MIN);
    } else if (strcmp(argv[1], "mul") == 0) {
        mul(6, 7); mul(INT_MAX, 2); mul(46340, 46340); mul(46341, 46341);
        mul(INT_MIN, -1); mul(-1, INT_MIN); mul(-1, INT_MAX); mul(0, INT_MIN); mul(INT_MIN, 1);
    } else if (strcmp(argv[1], "random") == 0) {
        uint64_t x = 88172645463325252u;
        for (int i = 0; i < 1000000; i++) {
            x ^= x << 13; x ^= x >> 7; x ^= x << 17;
            int a = (int)(uint32_t)x, b = (int)(uint32_t)(x >> 32);
            if (i % 3 == 0) b = (int)((uint32_t)b >> (i % 31));   /* smaller values too */
            int r1 = 12345, r2 = 12345;
            int64_t s = (int64_t)a + b, p = (int64_t)a * b;
            int okadd = s >= INT_MIN && s <= INT_MAX, okmul = p >= INT_MIN && p <= INT_MAX;
            int gotadd = checked_add(a, b, &r1), gotmul = checked_mul(a, b, &r2);
            if (gotadd != okadd || (okadd && r1 != (int)s) || (!okadd && r1 != 12345)) {
                printf("checked_add(%d, %d) is wrong\n", a, b);
                return 0;
            }
            if (gotmul != okmul || (okmul && r2 != (int)p) || (!okmul && r2 != 12345)) {
                printf("checked_mul(%d, %d) is wrong\n", a, b);
                return 0;
            }
        }
        printf("1000000 pairs correct\n");
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "additions at the edges", Args: []string{"add"},
				Want:    "1 + 2: ok 3\n2147483647 + 1: overflow 12345\n-2147483648 + -1: overflow 12345\n2147483647 + -2147483648: ok -1\n-5 + 5: ok 0\n2147483647 + 0: ok 2147483647\n-2147483648 + -2147483648: overflow 12345\n",
				Explain: "Both directions overflow: INT_MAX + 1 and INT_MIN − 1. On overflow, leave the output untouched (the checker set it to 12345)."},
			{Name: "multiplications at the edges", Args: []string{"mul"},
				Want:    "6 * 7: ok 42\n2147483647 * 2: overflow 12345\n46340 * 46340: ok 2147395600\n46341 * 46341: overflow 12345\n-2147483648 * -1: overflow 12345\n-1 * -2147483648: overflow 12345\n-1 * 2147483647: ok -2147483647\n0 * -2147483648: ok 0\n-2147483648 * 1: ok -2147483648\n",
				Explain: "INT_MIN × −1 is the trap: −INT_MIN is one more than INT_MAX. Dividing to test (a > INT_MAX / b) needs care with negative numbers and with 0."},
			{Name: "a million random pairs", Args: []string{"random"}, Want: "1000000 pairs correct\n"},
		},
		Hints: []string{
			"Addition: if (b > 0 ? a > INT_MAX - b : a < INT_MIN - b) return 0; INT_MAX - b cannot overflow when b > 0, nor INT_MIN - b when b < 0.",
			"Multiplication without division: compute the product in a long long, which holds any product of two ints exactly, then check the range before storing.",
			"Or use the compiler's builtins: if (__builtin_mul_overflow(a, b, &r)) return 0;",
		},
	},
	{
		ID: "arena", Stage: 6, Concept: "Memory Allocators: How malloc Works",
		Title: "An arena allocator",
		Task:  "Write an arena: arena_init takes one block of size bytes from malloc; arena_alloc hands out the next n bytes of it, at an address that is a multiple of 16, or returns NULL when they do not fit; arena_reset makes the whole block free again; arena_destroy frees it. arena_alloc must never call malloc, and a request too big to fit must give NULL, however big. Keep struct arena exactly as given: the checker's file uses it too. No main.",
		File:  "arena.c",
		Leaks: true,
		Starter: `/* Checked exercise: arena
 * One malloc up front; every allocation after it is a pointer bump.
 *   int   arena_init(struct arena *a, size_t size);  1 on success, 0 if malloc failed
 *   void *arena_alloc(struct arena *a, size_t n);    next n bytes, 16-byte aligned, or NULL
 *   void  arena_reset(struct arena *a);              everything free again
 *   void  arena_destroy(struct arena *a);            give the block back to free
 * Align the address, not just the offset: round (uintptr_t)(base + used)
 * up to a multiple of 16 with (p + 15) & ~(uintptr_t)15.
 * Beware: used + n can wrap around past SIZE_MAX to a small number.
 * Compare n with what is left instead.
 */
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

struct arena {
    char *base;   /* the block, from one malloc */
    size_t size;  /* its size in bytes */
    size_t used;  /* bytes handed out so far, padding included */
};
`,
		Harness: `#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

struct arena {
    char *base;
    size_t size;
    size_t used;
};

int arena_init(struct arena *a, size_t size);
void *arena_alloc(struct arena *a, size_t n);
void arena_reset(struct arena *a);
void arena_destroy(struct arena *a);

struct node { struct node *next; long value; };

static int inside(struct arena *a, void *p, size_t n) {
    char *c = p;
    return c >= a->base && c + n <= a->base + a->size;
}

static void report(struct arena *a, size_t n) {
    void *p = arena_alloc(a, n);
    if (p == NULL) { printf("%zu null\n", n); return; }
    if ((uintptr_t)p % 16 != 0) printf("%zu not aligned\n", n);
    else if (!inside(a, p, n)) printf("%zu outside the block\n", n);
    else { memset(p, 0x5a, n); printf("%zu ok\n", n); }
}

int main(int argc, char **argv) {
    struct arena a;
    if (argc < 2) return 2;
    if (!arena_init(&a, strcmp(argv[1], "nodes") == 0 ? 1000000 * sizeof(struct node) : 256)) {
        printf("arena_init failed\n");
        return 0;
    }
    if (strcmp(argv[1], "sizes") == 0) {
        size_t sizes[] = {1, 3, 24, 100, 7};
        unsigned char *p[5];
        char *end = NULL;
        for (int i = 0; i < 5; i++) {
            p[i] = arena_alloc(&a, sizes[i]);
            if (p[i] == NULL) { printf("block %d: NULL, but it fits\n", i); goto done; }
            if ((uintptr_t)p[i] % 16 != 0) { printf("block %d: address not a multiple of 16\n", i); goto done; }
            if (!inside(&a, p[i], sizes[i])) { printf("block %d: outside the arena's block\n", i); goto done; }
            if (end != NULL && (char *)p[i] < end) { printf("block %d overlaps the one before\n", i); goto done; }
            end = (char *)p[i] + sizes[i];
            memset(p[i], i + 1, sizes[i]);
        }
        for (int i = 0; i < 5; i++)
            for (size_t j = 0; j < sizes[i]; j++)
                if (p[i][j] != i + 1) { printf("block %d was overwritten\n", i); goto done; }
        printf("5 blocks: aligned, inside, apart\n");
    } else if (strcmp(argv[1], "full") == 0) {
        report(&a, 200); report(&a, 100); report(&a, 40); report(&a, 8);
    } else if (strcmp(argv[1], "huge") == 0) {
        report(&a, 16); report(&a, SIZE_MAX); report(&a, SIZE_MAX - 8); report(&a, 240); report(&a, 1);
    } else if (strcmp(argv[1], "reset") == 0) {
        void *first = arena_alloc(&a, 64);
        arena_alloc(&a, 64);
        arena_alloc(&a, 64);
        arena_reset(&a);
        void *again = arena_alloc(&a, 200);
        printf(again == first ? "the block is reused\n" : "after reset, the first allocation should start at the beginning again\n");
    } else if (strcmp(argv[1], "nodes") == 0) {
        for (int round = 0; round < 2; round++) {
            struct node *head = NULL;
            for (long i = 0; i < 1000000; i++) {
                struct node *n = arena_alloc(&a, sizeof *n);
                if (n == NULL) { printf("node %ld: NULL, but the arena has room for a million\n", i); goto done; }
                n->value = i;
                n->next = head;
                head = n;
            }
            long sum = 0;
            for (struct node *n = head; n != NULL; n = n->next) sum += n->value;
            printf("%ld\n", sum);
            arena_reset(&a);
        }
    } else return 2;
done:
    arena_destroy(&a);
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "sizes 1, 3, 24, 100 and 7", Args: []string{"sizes"}, Want: "5 blocks: aligned, inside, apart\n",
				Explain: "Every block must start at a multiple of 16, inside the one block from arena_init, after the end of the block before it."},
			{Name: "until it is full", Args: []string{"full"}, Want: "200 ok\n100 null\n40 ok\n8 null\n",
				Explain: "In a 256-byte arena, after 200 bytes the next block starts at 208: 100 bytes do not fit, 40 do (208 + 40 = 248), and then 8 do not (256 + 8 > 256). A failed request must not use up any space."},
			{Name: "a request too big to add", Args: []string{"huge"}, Want: "16 ok\n18446744073709551615 null\n18446744073709551607 null\n240 ok\n1 null\n",
				Explain: "used + n wraps around past SIZE_MAX to a small number, and the check passes. Compare n with what is left, a->size - start, instead: that cannot overflow."},
			{Name: "reset", Args: []string{"reset"}, Want: "the block is reused\n"},
			{Name: "a million nodes, twice", Args: []string{"nodes"}, Want: "499999500000\n499999500000\n",
				Explain: "A million 16-byte nodes fill the arena exactly: no padding is needed when every size is already a multiple of 16."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"uintptr_t p = (uintptr_t)(a->base + a->used); p = (p + 15) & ~(uintptr_t)15; is the next aligned address; its offset is p - (uintptr_t)a->base.",
			"Fail when start > a->size || n > a->size - start: that order of checks never overflows. Only then set a->used = start + n.",
			"arena_reset only sets used back to 0; arena_destroy calls free(a->base) once.",
		},
	},
	{
		ID: "pipeline", Stage: 6, Concept: "File Descriptors, Files & Pipes",
		Title: "Connect two programs with a pipe",
		Task:  "./pipeline A ARGS... '|' B ARGS... runs A | B as a shell does: create a pipe, start both programs as child processes with fork and execvp, standard output of A going into the pipe and standard input of B coming out of it, then wait for both and return B's exit status. If exec fails, the child reports why on standard error and exits with 127. Without a | argument (or with nothing on either side), print a usage message to standard error and return 2.",
		File:  "pipeline.c",
		Starter: posixDefine + `/* Checked exercise: pipeline
 * Usage: ./pipeline echo hi '|' tr a-z A-Z
 * 1. Find the "|" argument and replace it with NULL: argv[1] and the
 *    word after the "|" are now two argument vectors for execvp.
 * 2. pipe(fds): fds[0] is the read end, fds[1] the write end.
 * 3. Fork twice. In the first child dup2(fds[1], 1); in the second
 *    dup2(fds[0], 0). Each child then closes both fds and calls execvp.
 * 4. The parent closes both fds too, then waits for both children.
 * A pipe only reports end-of-file when every copy of its write end is
 * closed, so one forgotten close makes the reader wait forever.
 */
`,
		Cases: []CheckCase{
			{Name: "echo hi | tr a-z A-Z", Args: []string{"echo", "hi", "|", "tr", "a-z", "A-Z"}, Want: "HI\n"},
			{Name: "echo hello | cat", Args: []string{"echo", "hello", "|", "cat"}, Want: "hello\n",
				Explain: "cat finishes only at end-of-file, which comes when every copy of the write end is closed: in the parent, and in the second child after its dup2."},
			{Name: "more than the pipe holds", Args: []string{"seq", "1", "100000", "|", "tail", "-n", "1"}, Want: "100000\n",
				Explain: "seq prints about 575 KB, far more than the pipe's 64 KiB: start both children before waiting for either, or the first blocks on a full pipe while you wait for it."},
			{Name: "yes | head -n 3", Args: []string{"yes", "|", "head", "-n", "3"}, Want: "y\ny\ny\n",
				Explain: "yes never stops by itself: it ends with SIGPIPE when nobody can read the pipe any more. If the parent (or yes itself) still holds the read end, yes blocks on a full pipe instead and nothing ends."},
			{Name: "B's exit status", Args: []string{"true", "|", "false"}, WantExit: 1,
				Explain: "Return the exit status of the second program, decoded with WIFEXITED and WEXITSTATUS, like a shell does."},
			{Name: "only B's exit status counts", Args: []string{"false", "|", "true"}},
			{Name: "a program that does not exist", Args: []string{"echo", "hi", "|", "/no/such/program"}, WantExit: 127, Stderr: true,
				Explain: "execvp only returns when it fails: then perror and _exit(127) in the child."},
			{Name: "no pipe", Args: []string{"echo", "hi"}, WantExit: 2, Stderr: true},
			{Name: "nothing after the pipe", Args: []string{"echo", "hi", "|"}, WantExit: 2, Stderr: true},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Count the open copies of the write end: the pipe makes one; each fork copies it. Every copy except the one dup2 put on A's descriptor 1 must be closed.",
			"After dup2(fds[1], 1) the pipe is reachable as descriptor 1, so close(fds[0]) and close(fds[1]) lose nothing.",
			"Call fflush(stdout) before fork if you have printed anything, and use _exit in a child whose exec failed.",
		},
	},
	{
		ID: "graceful", Stage: 6, Concept: "Signals",
		Title: "Stop when asked",
		Task:  "Write int install_stop_handlers(void), which makes SIGTERM and SIGINT stop the program politely instead of killing it (return 0, or -1 if sigaction fails), and int stop_signal(void), which returns 0 until one of them arrives and then the signal's number. The handler may only record the signal. Install it with sigaction and without SA_RESTART, so that a blocking call such as read returns early (with errno EINTR) and the program notices. No main: the checker sends the signals.",
		File:  "graceful.c",
		Starter: posixDefine + `/* Checked exercise: graceful
 *   int install_stop_handlers(void);  SIGTERM and SIGINT; 0, or -1 on failure
 *   int stop_signal(void);            0 until one arrives, then its number
 * A handler can run between any two instructions, even inside printf or
 * malloc, so it must only set a flag:
 *   static volatile sig_atomic_t got;
 *   static void on_stop(int sig) { got = sig; }
 * Install it with sigaction, sa_flags = 0. (signal() on Linux and macOS
 * sets SA_RESTART: a blocking read would then carry on waiting.)
 */
#include <signal.h>
`,
		Harness: posixDefine + `#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <string.h>
#include <sys/wait.h>
#include <time.h>
#include <unistd.h>

int install_stop_handlers(void);
int stop_signal(void);

static void nap(long ms) {
    struct timespec t = {ms / 1000, (ms % 1000) * 1000000L};
    nanosleep(&t, NULL);
}

/* child runs in a new process: it signals that it is ready, then waits
   for SIGTERM in a busy loop or in a blocking read. */
static void child(int ready, int blocking) {
    int p[2];
    if (install_stop_handlers() != 0 || pipe(p) != 0) _exit(3);
    if (write(ready, "r", 1) != 1) _exit(3);
    if (blocking) {
        char c;
        ssize_t n = read(p[0], &c, 1);   /* nobody writes: only a signal ends this */
        if (n < 0 && errno == EINTR)
            printf("child: read interrupted, stop %d\n", stop_signal());
        else
            printf("child: read returned %zd\n", n);
    } else {
        volatile unsigned long spins = 0;
        while (stop_signal() == 0) spins++;
        printf("child: stop %d\n", stop_signal());
    }
    fflush(stdout);
    _exit(0);
}

static void from_outside(int blocking) {
    int ready[2];
    char c;
    if (pipe(ready) != 0) return;
    fflush(stdout);
    pid_t pid = fork();
    if (pid == 0) { close(ready[0]); child(ready[1], blocking); }
    close(ready[1]);
    if (read(ready[0], &c, 1) != 1) { printf("the child could not install its handlers\n"); return; }
    int status = 0;
    pid_t done = 0;
    for (int i = 0; i < 20 && done == 0; i++) {   /* SIGTERM every 100 ms, for 2 s */
        kill(pid, SIGTERM);
        nap(100);
        done = waitpid(pid, &status, WNOHANG);
    }
    if (done == 0) {
        kill(pid, SIGKILL);
        waitpid(pid, &status, 0);
        printf(blocking ? "the read was restarted after SIGTERM, so the child never noticed\n"
                        : "the child kept running after SIGTERM\n");
    } else if (WIFSIGNALED(status)) {
        printf("child killed by signal %d\n", WTERMSIG(status));
    } else {
        printf("child exited %d\n", WEXITSTATUS(status));
    }
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "nothing") == 0) {
        if (install_stop_handlers() != 0) { printf("install_stop_handlers failed\n"); return 0; }
        printf("stop %d\n", stop_signal());
    } else if (strcmp(argv[1], "term") == 0 || strcmp(argv[1], "int") == 0) {
        if (install_stop_handlers() != 0) { printf("install_stop_handlers failed\n"); return 0; }
        fflush(stdout);
        raise(argv[1][0] == 't' ? SIGTERM : SIGINT);
        printf("stop %d\nstill running, free to clean up\n", stop_signal());
    } else if (strcmp(argv[1], "busy") == 0) {
        from_outside(0);
    } else if (strcmp(argv[1], "blocked") == 0) {
        from_outside(1);
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "no signal yet", Args: []string{"nothing"}, Want: "stop 0\n"},
			{Name: "SIGTERM", Args: []string{"term"}, Want: "stop 15\nstill running, free to clean up\n",
				Explain: "Catch SIGTERM with sigaction: by default it kills the program. The handler records the signal and returns; it must not exit or print."},
			{Name: "SIGINT", Args: []string{"int"}, Want: "stop 2\nstill running, free to clean up\n",
				Explain: "Ctrl-C sends SIGINT (2): install the same handler for it."},
			{Name: "from another process, while working", Args: []string{"busy"}, Want: "child: stop 15\nchild exited 0\n",
				Explain: "kill(pid, SIGTERM) from another process must set the flag a running loop reads. Make the flag volatile sig_atomic_t, or the compiler may keep a copy of it in a register."},
			{Name: "from another process, while blocked in read", Args: []string{"blocked"}, Want: "child: read interrupted, stop 15\nchild exited 0\n",
				Explain: "Without SA_RESTART, the signal makes read return -1 with errno EINTR. signal(), or sa_flags = SA_RESTART, restarts the read instead, and a server blocked in read or accept never notices it was asked to stop."},
		},
		Timeout: 6 * time.Second,
		Hints: []string{
			"struct sigaction sa; memset(&sa, 0, sizeof sa); sa.sa_handler = on_stop; sigemptyset(&sa.sa_mask); then sigaction(SIGTERM, &sa, NULL) and the same for SIGINT.",
			"The handler only stores: got = sig. stop_signal just returns got.",
			"printf, malloc and exit are not async-signal-safe: never call them from a handler.",
		},
	},
	{
		ID: "buffer_pool", Stage: 7, Concept: "Pages, Records & the Buffer Pool",
		Title: "A buffer pool with LRU eviction",
		Task:  "Simulate a database's buffer pool. The first line of input is the number of frames and a mode, trace or summary; each line after it is r P (read page P), w P (write page P: it becomes dirty), p P (pin P: read it and keep it from being evicted) or u P (unpin P once). A read, write or pin of a page already in the pool is a hit. A miss loads the page into a free frame, or else evicts the least recently used unpinned page, writing it back first if it is dirty; if every frame is pinned, nothing is loaded. In trace mode print one line per miss (see the starter); in both modes finish with hits H, misses M, writes W.",
		File:  "buffer_pool.c",
		Starter: `/* Checked exercise: buffer_pool
 * Input:  3 trace          frames, and trace or summary
 *         r 1              read page 1
 *         w 2              write page 2 (dirty from now on)
 *         p 3              pin page 3 (an access too)
 *         u 3              unpin page 3 (not an access)
 * On a miss, in trace mode, print one of:
 *   load 4                       a frame was free
 *   evict 2, load 4              2 was the least recently used unpinned page
 *   write 2, evict 2, load 4     ... and it was dirty: written back first
 *   no frame for 4               every frame is pinned: nothing loaded
 * At the end, in both modes:
 *   hits H, misses M, writes W   (a "no frame" access counts as a miss)
 * Each frame needs: page number, pin count, dirty flag, time of last use.
 */
#include <stdio.h>
`,
		Cases: []CheckCase{
			{Name: "three frames", Stdin: "3 trace\nr 1\nr 2\nr 3\nr 1\nr 4\nr 2\nr 1\n",
				Want:    "load 1\nload 2\nload 3\nevict 2, load 4\nevict 3, load 2\nhits 2, misses 5, writes 0\n",
				Explain: "Reading 1 again made it the most recently used, so 4 evicts 2, the least recently used; then 2 evicts 3."},
			{Name: "dirty pages are written back", Stdin: "2 trace\nw 1\nr 2\nr 3\nw 2\nr 1\nr 4\n",
				Want:    "load 1\nload 2\nwrite 1, evict 1, load 3\nevict 3, load 1\nwrite 2, evict 2, load 4\nhits 1, misses 5, writes 2\n",
				Explain: "Page 1 was written, so evicting it writes it back. Page 1 comes back clean. Page 2 became dirty on the hit w 2."},
			{Name: "a pinned page stays", Stdin: "2 trace\np 1\nr 2\nr 3\nr 4\nu 1\nr 5\n",
				Want:    "load 1\nload 2\nevict 2, load 3\nevict 3, load 4\nevict 1, load 5\nhits 0, misses 5, writes 0\n",
				Explain: "While 1 is pinned it is never the victim, even though it is the least recently used. Once unpinned, it goes first."},
			{Name: "every frame pinned", Stdin: "2 trace\np 1\np 2\nr 3\np 1\nu 1\nr 3\nu 1\nr 3\n",
				Want:    "load 1\nload 2\nno frame for 3\nno frame for 3\nevict 1, load 3\nhits 1, misses 5, writes 0\n",
				Explain: "Page 1 was pinned twice, so it needs two unpins. Until then, there is no frame for page 3: print that and load nothing."},
			{Name: "a scan floods LRU", Stdin: "4 summary\n" + strings.Repeat("r 1\nr 2\nr 3\nr 4\nr 5\n", 20),
				Want:    "hits 0, misses 100, writes 0\n",
				Explain: "Scanning five pages in a loop through four frames always evicts exactly the page needed next: LRU's worst case."},
			{Name: "a million accesses", Stdin: bufferPoolInput(), Want: bufferPoolWant(),
				Explain: "Use the same rules as the small cases: a hit refreshes the time of last use, and only a dirty page is written back."},
		},
		Hints: []string{
			"Keep a counter that goes up by one per access; a frame's last use is the counter's value at its last access, and the victim is the unpinned frame with the smallest one.",
			"Look for the page first (a hit); on a miss, look for a free frame before choosing a victim.",
			"u P does nothing unless P is in the pool with a pin count above 0, and it does not count as an access.",
		},
	},
	{
		ID: "bloom", Stage: 7, Concept: "LSM-Trees",
		Title: "A Bloom filter",
		Task:  "Write a Bloom filter of m bits with k hash functions: bloom_init clears m bits (allocated with malloc or calloc, rounded up to whole bytes), bloom_add sets the k bits for a key, bloom_maybe returns 1 if all k bits for the key are set and 0 otherwise, and bloom_free frees the bits. It must never say 0 for a key that was added, and with 10 bits per key and k = 7 it should say 1 for fewer than 1.5% of the keys that were not. Keep struct bloom exactly as given. No main.",
		File:  "bloom.c",
		Leaks: true,
		Starter: `/* Checked exercise: bloom
 *   int  bloom_init(struct bloom *b, size_t m, int k);   1 on success
 *   void bloom_add(struct bloom *b, const char *key);
 *   int  bloom_maybe(const struct bloom *b, const char *key);
 *   void bloom_free(struct bloom *b);
 * Bit i lives in byte i / 8, at position i % 8.
 * k hash functions from two: compute h1 and h2 of the key (for example
 * 64-bit FNV-1a, and FNV-1a with a different starting value), then use
 * positions (h1 + j * h2) % m for j = 0 .. k-1. Make h2 odd.
 * FNV-1a: h = 14695981039346656037; for each byte: h ^= byte; h *= 1099511628211;
 */
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

struct bloom {
    unsigned char *bits;  /* m bits, (m + 7) / 8 bytes */
    size_t m;             /* number of bits */
    int k;                /* number of hash functions */
};
`,
		Harness: `#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

struct bloom {
    unsigned char *bits;
    size_t m;
    int k;
};

int bloom_init(struct bloom *b, size_t m, int k);
void bloom_add(struct bloom *b, const char *key);
int bloom_maybe(const struct bloom *b, const char *key);
void bloom_free(struct bloom *b);

static size_t ones(const struct bloom *b) {
    size_t n = 0;
    for (size_t i = 0; i < b->m; i++) n += (b->bits[i / 8] >> (i % 8)) & 1;
    return n;
}

int main(int argc, char **argv) {
    struct bloom b;
    char key[32];
    if (argc < 2) return 2;
    if (strcmp(argv[1], "one") == 0) {
        if (!bloom_init(&b, 1000, 7)) { printf("bloom_init failed\n"); return 0; }
        printf("empty: %s\n", bloom_maybe(&b, "hello") ? "maybe" : "no");
        bloom_add(&b, "hello");
        size_t n = ones(&b);
        printf("bits set: %s\n", n >= 1 && n <= 7 ? "1 to 7" : "not 1 to 7");
        printf("hello: %s\n", bloom_maybe(&b, "hello") ? "maybe" : "no");
    } else if (strcmp(argv[1], "added") == 0) {
        if (!bloom_init(&b, 1000000, 7)) { printf("bloom_init failed\n"); return 0; }
        for (int i = 0; i < 100000; i++) { snprintf(key, sizeof key, "key%d", i); bloom_add(&b, key); }
        for (int i = 0; i < 100000; i++) {
            snprintf(key, sizeof key, "key%d", i);
            if (!bloom_maybe(&b, key)) { printf("%s was added, but bloom_maybe says no\n", key); goto done; }
        }
        printf("all 100000 keys: maybe\n");
    } else if (strcmp(argv[1], "rate") == 0) {
        if (!bloom_init(&b, 1000000, 7)) { printf("bloom_init failed\n"); return 0; }
        for (int i = 0; i < 100000; i++) { snprintf(key, sizeof key, "user:%06d", i); bloom_add(&b, key); }
        double set = ones(&b) / 1e6;
        if (set < 0.45 || set > 0.56) printf("%.1f%% of the bits are set, expected about 50%%\n", set * 100);
        else printf("about half the bits set\n");
        int fp = 0;
        for (int i = 0; i < 100000; i++) { snprintf(key, sizeof key, "item:%06d", i); fp += bloom_maybe(&b, key); }
        if (fp < 1500) printf("false positives under 1.5%%\n");
        else printf("false positives: %.2f%%\n", fp / 1000.0);
    } else if (strcmp(argv[1], "odd") == 0) {
        if (!bloom_init(&b, 1001, 3)) { printf("bloom_init failed\n"); return 0; }
        for (int i = 0; i < 100; i++) { snprintf(key, sizeof key, "k%d", i); bloom_add(&b, key); }
        for (int i = 0; i < 100; i++) {
            snprintf(key, sizeof key, "k%d", i);
            if (!bloom_maybe(&b, key)) { printf("%s was added, but bloom_maybe says no\n", key); goto done; }
        }
        if (ones(&b) > 300) { printf("100 keys set more than 3 bits each\n"); goto done; }
        printf("1001 bits: all 100 keys found\n");
    } else return 2;
done:
    bloom_free(&b);
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "one key", Args: []string{"one"}, Want: "empty: no\nbits set: 1 to 7\nhello: maybe\n",
				Explain: "A new filter has every bit clear (calloc, or memset after malloc), and one key sets at most k bits."},
			{Name: "no false negatives", Args: []string{"added"}, Want: "all 100000 keys: maybe\n",
				Explain: "bloom_add and bloom_maybe must compute exactly the same k positions, and set and test the same bit: byte i / 8, bit i % 8."},
			{Name: "the false-positive rate", Args: []string{"rate"}, Want: "about half the bits set\nfalse positives under 1.5%\n",
				Explain: "With 10 bits per key and k = 7, theory says 0.82%. A higher rate means the positions are not independent: a hash that ignores part of the key, or k positions that are all the same."},
			{Name: "a size that is not a multiple of 8", Args: []string{"odd"}, Want: "1001 bits: all 100 keys found\n",
				Explain: "1,001 bits need 126 bytes: (m + 7) / 8, not m / 8."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Set bit i: b->bits[i / 8] |= (unsigned char)(1u << (i % 8)); test it: (b->bits[i / 8] >> (i % 8)) & 1.",
			"Use uint64_t for the hashes, so that the multiplication wraps around instead of overflowing, then take % b->m.",
			"bloom_init: b->bits = calloc((m + 7) / 8, 1); remember m and k in the struct.",
		},
	},
	{
		ID: "hash_join", Stage: 7, Concept: "Query Planning & Join Algorithms",
		Title: "A hash join",
		Task:  "Read two tables from standard input and print their join on the key. Each table is a line with its number of rows, then one row per line: an integer key (64-bit, possibly negative) and a word of up to 15 characters. For each row of the second table, in order, print key a b for every row of the first table with the same key, in the first table's order. The tables can have 100,000 rows each, so compare keys through a hash table, not every pair.",
		File:  "hash_join.c",
		Starter: `/* Checked exercise: hash_join
 * Input:   3          rows of table A
 *          1 apple
 *          2 pear
 *          1 plum
 *          2          rows of table B
 *          1 red
 *          3 blue
 * Output:  1 apple red
 *          1 plum red
 * Build: put every row of A into a hash table keyed on the key, keeping
 * rows with the same key in A's order. Probe: for each row of B, look up
 * its key and print every match.
 * A nested loop compares 100,000 x 100,000 = ten billion pairs: too slow.
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "the example", Stdin: "3\n1 apple\n2 pear\n1 plum\n2\n1 red\n3 blue\n", Want: "1 apple red\n1 plum red\n",
				Explain: "For key 1, both rows of A match, printed in A's order: apple, then plum."},
			{Name: "no matches", Stdin: "2\n1 a\n2 b\n2\n3 c\n4 d\n", Want: ""},
			{Name: "an empty table", Stdin: "0\n2\n1 a\n2 b\n", Want: ""},
			{Name: "duplicates on both sides", Stdin: "3\n5 x\n5 y\n6 z\n3\n5 p\n6 q\n5 r\n", Want: "5 x p\n5 y p\n6 z q\n5 x r\n5 y r\n",
				Explain: "Every pair of rows with the same key is one line of output: two A rows times two B rows with key 5 make four lines."},
			{Name: "negative and huge keys", Stdin: "3\n-7 neg\n9223372036854775807 max\n-9223372036854775808 min\n3\n-9223372036854775808 m\n-7 n\n9223372036854775807 x\n",
				Want:    "-9223372036854775808 min m\n-7 neg n\n9223372036854775807 max x\n",
				Explain: "In C, key % buckets is negative for a negative key: an index before the start of the array. Convert the key to an unsigned type before hashing."},
			{Name: "100,000 rows each", Stdin: hashJoinInput(), Want: hashJoinWant(),
				Explain: "Ten billion comparisons take minutes; building a hash table of A and probing it once per row of B takes milliseconds."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Read with scanf(\"%lld %15s\", &key, word); store A's rows in an array, and the hash table as bucket heads plus a next index per row.",
			"To keep A's order within a bucket, insert A's rows in reverse order at the head of each chain, or keep a tail pointer per bucket.",
			"Choose more buckets than rows (a power of two, such as 262,144) and mix the key's bits: (uint64_t)key * 0x9E3779B97F4A7C15 >> 46 for 2^18 buckets.",
		},
	},
	{
		ID: "routes", Stage: 8, Concept: "IP Addressing, Routing & NAT",
		Title: "Longest-prefix match",
		Task:  "Route packets. The first line of input is the number of routes; each route is a prefix such as 10.1.0.0/16 and a next-hop name. Every line after the routes is an IPv4 address: print the next hop of the longest prefix that contains it, or unreachable if none does. A prefix's host bits do not matter (10.1.2.3/8 means 10.0.0.0/8), /0 matches every address, and if one prefix appears twice the later route wins. There can be 100,000 routes and 100,000 addresses.",
		File:  "routes.c",
		Starter: `/* Checked exercise: routes
 * Input:   3                   number of routes
 *          10.0.0.0/8 A
 *          10.1.0.0/16 B
 *          0.0.0.0/0 D
 *          10.1.2.3            addresses, until the end of input
 *          8.8.8.8
 * Output:  B                   10.1.0.0/16 is the longest match
 *          D
 * An address is 32 bits: (a << 24) | (b << 16) | (c << 8) | d, as a
 * uint32_t. The mask of a /n prefix is n one-bits, then zeros. Careful:
 * shifting a 32-bit value by 32 is undefined in C, so /0 needs its own case.
 * Comparing each address with every route is 10^10 steps for the big case:
 * use a binary trie (one level per bit), or a hash table per length.
 */
#include <stdint.h>
#include <stdio.h>
`,
		Cases: []CheckCase{
			{Name: "the most specific route wins", Stdin: "4\n10.0.0.0/8 A\n10.1.0.0/16 B\n192.168.1.0/24 C\n0.0.0.0/0 D\n10.1.2.3\n10.200.0.1\n192.168.1.77\n8.8.8.8\n192.168.2.1\n",
				Want:    "B\nA\nC\nD\nD\n",
				Explain: "10.1.2.3 matches 10.0.0.0/8, 10.1.0.0/16 and the default; the longest, /16, wins. 192.168.2.1 is outside 192.168.1.0/24."},
			{Name: "no default route", Stdin: "1\n10.0.0.0/8 A\n11.0.0.1\n10.255.255.255\n", Want: "unreachable\nA\n"},
			{Name: "/32 and /0", Stdin: "2\n0.0.0.0/0 X\n1.2.3.4/32 Y\n1.2.3.4\n1.2.3.5\n255.255.255.255\n0.0.0.0\n", Want: "Y\nX\nX\nX\n",
				Explain: "A /0 mask is 0, and a /32 mask is all ones. 0xFFFFFFFFu << 32 is undefined in C, not 0: handle /0 separately."},
			{Name: "host bits in a prefix", Stdin: "1\n10.1.2.3/8 A\n10.9.9.9\n", Want: "A\n",
				Explain: "Mask the prefix's address too: 10.1.2.3/8 is the network 10.0.0.0/8."},
			{Name: "the later route wins", Stdin: "2\n10.0.0.0/8 A\n10.0.0.0/8 B\n10.1.1.1\n", Want: "B\n"},
			{Name: "100,000 routes", Stdin: routesInput(), Want: routesWant(),
				Explain: "Comparing every address with every route is ten billion comparisons. A trie finds the longest match in at most 32 steps."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Parse with sscanf(line, \"%u.%u.%u.%u/%d\", &a, &b, &c, &d, &len) and check the result is 5 (4 for an address).",
			"A binary trie: each node has two children and an optional next hop. Insert walks the first len bits from the top bit down; lookup walks the address's bits, remembering the last next hop it passed.",
			"The mask of /n, for n from 1 to 32: 0xFFFFFFFFu << (32 - n). For n = 0 it is 0.",
		},
	},
	{
		ID: "retry", Stage: 8, Concept: "Timeouts, Retries & Idempotency",
		Title: "Retry with backoff and jitter",
		Task:  "Write int with_retries(void *op, int max_attempts, long base_ms, long cap_ms). It calls attempt(op), which returns 0 for success, 1 for a temporary failure and 2 for a permanent one, up to max_attempts times. Return at once on success (0) or a permanent failure (2); after a temporary failure, if attempts remain, wait jitter(d) milliseconds with wait_ms, where d is base_ms for the first retry and doubles for each one after it, never above cap_ms, then try again. If every attempt failed temporarily, return 1. The checker supplies attempt, jitter and wait_ms, and records every call. No main.",
		File:  "retry.c",
		Starter: `/* Checked exercise: retry
 * Supplied by the checker:
 *   int  attempt(void *op);    0 success, 1 temporary failure, 2 permanent
 *   long jitter(long max);     a random wait between 0 and max
 *   void wait_ms(long ms);     sleep (the checker only records it)
 * Write:
 *   int with_retries(void *op, int max_attempts, long base_ms, long cap_ms);
 * Retry only temporary failures, and never wait after the last attempt.
 * Waits: jitter(d) with d = base, 2 x base, 4 x base, ... at most cap.
 * Doubling a long 60 times overflows it: stop doubling once d reaches cap.
 */
int attempt(void *op);
long jitter(long max);
void wait_ms(long ms);
`,
		Harness: `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int with_retries(void *op, int max_attempts, long base_ms, long cap_ms);

static const char *script;   /* one letter per attempt: S, T or P; the last repeats */
static int calls, waits;
static long last_jitter = -1;
static int bad_wait;

int attempt(void *op) {
    (void)op;
    int n = (int)strlen(script);
    char c = script[calls < n ? calls : n - 1];
    if (calls > waits) printf("tried again without waiting\n");
    calls++;
    return c == 'S' ? 0 : c == 'P' ? 2 : 1;
}

long jitter(long max) {
    printf("jitter(%ld) ", max);
    last_jitter = max / 3;
    return last_jitter;
}

void wait_ms(long ms) {
    if (ms != last_jitter) bad_wait = 1;
    printf("wait %ld\n", ms);
    last_jitter = -1;
    waits++;
}

int main(int argc, char **argv) {
    if (argc < 5) return 2;
    script = argv[1];
    int r = with_retries(&calls, atoi(argv[2]), atol(argv[3]), atol(argv[4]));
    if (bad_wait) printf("wait_ms must be given the value jitter returned\n");
    if (waits >= calls && calls > 0) printf("waited after the last attempt\n");
    printf("%d attempts, result %d\n", calls, r);
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "success at once", Args: []string{"S", "5", "100", "10000"}, Want: "1 attempts, result 0\n"},
			{Name: "two failures, then success", Args: []string{"TTS", "5", "100", "10000"},
				Want:    "jitter(100) wait 33\njitter(200) wait 66\n3 attempts, result 0\n",
				Explain: "The first retry waits jitter(base), the next jitter(2 × base). Pass jitter's result to wait_ms."},
			{Name: "a permanent failure is not retried", Args: []string{"TP", "5", "100", "10000"},
				Want:    "jitter(100) wait 33\n2 attempts, result 2\n",
				Explain: "2 means retrying cannot help (a bad request, say): return it at once."},
			{Name: "out of attempts", Args: []string{"T", "4", "100", "10000"},
				Want:    "jitter(100) wait 33\njitter(200) wait 66\njitter(400) wait 133\n4 attempts, result 1\n",
				Explain: "Four attempts need only three waits: there is no point waiting after the last one."},
			{Name: "the cap", Args: []string{"T", "6", "100", "500"},
				Want:    "jitter(100) wait 33\njitter(200) wait 66\njitter(400) wait 133\njitter(500) wait 166\njitter(500) wait 166\n6 attempts, result 1\n",
				Explain: "The doubling stops at cap_ms: 100, 200, 400, then 500 instead of 800."},
			{Name: "seventy attempts", Args: []string{"T", "70", "1000", "60000"}, Want: retryWant(70, 1000, 60000),
				Explain: "1000 × 2^63 does not fit in a long: signed overflow is undefined behaviour. Stop doubling once d reaches the cap (if d > cap / 2, d = cap)."},
			{Name: "one attempt only", Args: []string{"T", "1", "100", "1000"}, Want: "1 attempts, result 1\n"},
		},
		Hints: []string{
			"Loop for i from 1 to max_attempts: r = attempt(op); if r != 1 return r; if i == max_attempts return 1; wait, then double d.",
			"wait_ms(jitter(d)) passes the random value straight on.",
			"Double safely: d = (d > cap_ms / 2) ? cap_ms : d * 2;",
		},
	},
	{
		ID: "ring", Stage: 8, Concept: "Partitioning & Consistent Hashing",
		Title: "A consistent-hash ring",
		Task:  "Place keys on a consistent-hash ring. Input: the number of virtual nodes per node, v; the number of nodes and their names; the number of keys and the keys, one per line. Node name N has v points on the ring, the hashes of the strings N#0, N#1 … N#(v-1). A key belongs to the node of the first point whose hash is at least the key's hash, wrapping around to the smallest point if there is none. Print each key's node, in order. The hash is 32-bit FNV-1a, given in the starter; when two points have the same hash, the node listed first wins. There can be 100,000 points and 100,000 keys.",
		File:  "ring.c",
		Starter: `/* Checked exercise: ring
 * Input:   2            virtual nodes per node
 *          3 A B C      nodes
 *          2            keys
 *          user:1
 *          user:2
 * Output:  the node of each key, one per line.
 * Points: fnv1a("A#0"), fnv1a("A#1"), fnv1a("B#0") ... Sort them, then
 * for each key binary-search for the first point >= fnv1a(key); past the
 * end, wrap around to the first point. Equal hashes: the node listed
 * first in the input wins (sort by hash, then by node position).
 * Names and keys are at most 63 characters.
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static uint32_t fnv1a(const char *s) {
    uint32_t h = 2166136261u;
    for (; *s; s++) {
        h ^= (unsigned char)*s;
        h *= 16777619u;
    }
    return h;
}
`,
		Cases: []CheckCase{
			{Name: "three nodes", Stdin: ringInput(1, []string{"A", "B", "C"}, ringKeys(12, "user:")), Want: ringWant(1, []string{"A", "B", "C"}, ringKeys(12, "user:")),
				Explain: "Hash each key, find the first point at or after it, and wrap to the smallest point when the key's hash is above every point."},
			{Name: "virtual nodes", Stdin: ringInput(50, []string{"alpha", "beta", "gamma", "delta"}, ringKeys(40, "k")), Want: ringWant(50, []string{"alpha", "beta", "gamma", "delta"}, ringKeys(40, "k")),
				Explain: "Every node has v points, named node#0 to node#(v-1), with the number in decimal."},
			{Name: "one node", Stdin: ringInput(3, []string{"solo"}, ringKeys(5, "x")), Want: ringWant(3, []string{"solo"}, ringKeys(5, "x"))},
			{Name: "100,000 points", Stdin: ringInput(100, ringNodes(1000), ringKeys(100000, "key-")), Want: ringWant(100, ringNodes(1000), ringKeys(100000, "key-")),
				Explain: "Scanning all 100,000 points for each of 100,000 keys is ten billion steps: sort the points once, then binary-search for each key."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Build the point string with snprintf(buf, sizeof buf, \"%s#%d\", name, i).",
			"Store points as structs { uint32_t hash; int node; } and sort with qsort; break ties on node so the first-listed node comes first.",
			"Binary search: lo = 0, hi = n; while lo < hi, mid = lo + (hi - lo) / 2, and keep the half that contains the first hash >= h. If lo == n, use point 0.",
		},
	},
}

// duplicatesValue is the number at position i of duplicatesInput.
func duplicatesValue(i int) int { return (i*7919)%400009 - 200000 }

// duplicatesInput is n distinct numbers, except that position b repeats
// the value at position a (when a >= 0).
func duplicatesInput(n, a, b int) string {
	var s strings.Builder
	fmt.Fprintln(&s, n)
	for i := 0; i < n; i++ {
		v := duplicatesValue(i)
		if i == b && a >= 0 {
			v = duplicatesValue(a)
		}
		fmt.Fprint(&s, v, " ")
	}
	s.WriteString("\n")
	return s.String()
}

// hanoiMoves is the full list of moves for n disks from A to C.
func hanoiMoves(n int) string {
	var b strings.Builder
	var move func(n int, from, to, via byte)
	move = func(n int, from, to, via byte) {
		if n == 0 {
			return
		}
		move(n-1, from, via, to)
		fmt.Fprintf(&b, "%d %c %c\n", n, from, to)
		move(n-1, via, to, from)
	}
	move(n, 'A', 'C', 'B')
	return b.String()
}

// bsearchData is the sorted array and the lookups of the large bsearch
// case, the same every run.
func bsearchData() (a, q []int) {
	r := rand.New(rand.NewPCG(2, 2))
	a = make([]int, 200000)
	v := -1000000
	for i := range a {
		v += r.IntN(3) // steps of 0, 1 or 2: repeats and gaps
		a[i] = v
	}
	q = make([]int, 100000)
	for i := range q {
		q[i] = r.IntN(220000) - 1010000
	}
	return a, q
}

func bsearchInput() string {
	a, q := bsearchData()
	var s strings.Builder
	fmt.Fprintln(&s, len(a))
	for _, v := range a {
		fmt.Fprint(&s, v, " ")
	}
	fmt.Fprintf(&s, "\n%d\n", len(q))
	for _, v := range q {
		fmt.Fprint(&s, v, " ")
	}
	s.WriteString("\n")
	return s.String()
}

func bsearchWant() string {
	a, q := bsearchData()
	var s strings.Builder
	for _, x := range q {
		i, found := slices.BinarySearch(a, x) // the first index with a[i] >= x
		if !found {
			i = -1
		}
		fmt.Fprintln(&s, i)
	}
	return s.String()
}

// intervalsData is the large intervals case, the same every run.
func intervalsData() [][2]int {
	r := rand.New(rand.NewPCG(3, 3))
	m := make([][2]int, 200000)
	for i := range m {
		start := r.IntN(999000000)
		m[i] = [2]int{start, start + 1 + r.IntN(1000000)}
	}
	return m
}

func intervalsInput() string {
	var s strings.Builder
	m := intervalsData()
	fmt.Fprintln(&s, len(m))
	for _, x := range m {
		fmt.Fprintln(&s, x[0], x[1])
	}
	return s.String()
}

func intervalsWant() int {
	m := intervalsData()
	slices.SortFunc(m, func(x, y [2]int) int { return x[1] - y[1] })
	n, last := 0, -1
	for _, x := range m {
		if x[0] >= last {
			n, last = n+1, x[1]
		}
	}
	return n
}

// editDistancePair is two related 2,000-character lines: the second is
// the first with about one character in ten changed, dropped or added.
func editDistancePair() (string, string) {
	r := rand.New(rand.NewPCG(4, 4))
	const letters = "abcdefghij"
	a := make([]byte, 2000)
	for i := range a {
		a[i] = letters[r.IntN(len(letters))]
	}
	var b []byte
	for _, c := range a {
		switch r.IntN(30) {
		case 0: // changed
			b = append(b, letters[r.IntN(len(letters))])
		case 1: // dropped
		case 2: // an extra letter
			b = append(b, c, letters[r.IntN(len(letters))])
		default:
			b = append(b, c)
		}
	}
	return string(a), string(b[:min(len(b), 2000)])
}

func editDistanceInput() string {
	a, b := editDistancePair()
	return a + "\n" + b + "\n"
}

func editDistanceWant() int {
	a, b := editDistancePair()
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			sub := prev[j-1]
			if a[i-1] != b[j-1] {
				sub++
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, sub)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

// satData is the large sat_count case: 120 random clauses of up to three
// literals over 16 variables.
func satData() (n int, clauses [][]int) {
	r := rand.New(rand.NewPCG(5, 5))
	n = 16
	for i := 0; i < 120; i++ {
		var c []int
		for j := 0; j <= r.IntN(3); j++ {
			v := 1 + r.IntN(n)
			if r.IntN(2) == 0 {
				v = -v
			}
			c = append(c, v)
		}
		clauses = append(clauses, c)
	}
	return n, clauses
}

func satInput() string {
	n, clauses := satData()
	var s strings.Builder
	fmt.Fprintln(&s, n, len(clauses))
	for _, c := range clauses {
		for _, l := range c {
			fmt.Fprint(&s, l, " ")
		}
		s.WriteString("0\n")
	}
	return s.String()
}

func satWant() int {
	n, clauses := satData()
	count := 0
	for row := 0; row < 1<<n; row++ {
		all := true
		for _, c := range clauses {
			any := false
			for _, l := range c {
				v := l
				if v < 0 {
					v = -v
				}
				bit := row>>(v-1)&1 == 1
				if bit == (l > 0) {
					any = true
				}
			}
			all = all && any
		}
		if all {
			count++
		}
	}
	return count
}

// binomial is C(n, k) by Pascal's triangle, in uint64 (exact up to n = 67).
func binomial(n, k int) uint64 {
	row := []uint64{1}
	for i := 1; i <= n; i++ {
		next := make([]uint64, i+1)
		next[0], next[i] = 1, 1
		for j := 1; j < i; j++ {
			next[j] = row[j-1] + row[j]
		}
		row = next
	}
	return row[k]
}

func binomialRowInput(n int) string {
	var s strings.Builder
	for k := 0; k <= n; k++ {
		fmt.Fprintln(&s, n, k)
	}
	return s.String()
}

func binomialRowWant(n int) string {
	var s strings.Builder
	for k := 0; k <= n; k++ {
		fmt.Fprintln(&s, binomial(n, k))
	}
	return s.String()
}

// modpowWant is aᵉ mod m for each triple, by square and multiply.
func modpowWant(cases [][3]uint64) string {
	var s strings.Builder
	for _, c := range cases {
		a, e, m := c[0]%c[2], c[1], c[2]
		result := 1 % m
		for e > 0 {
			if e&1 == 1 {
				result = result * a % m
			}
			a = a * a % m
			e >>= 1
		}
		fmt.Fprintln(&s, result)
	}
	return s.String()
}

func reverseInput(n int) string {
	var s strings.Builder
	fmt.Fprintln(&s, n)
	for i := n; i >= 1; i-- {
		fmt.Fprint(&s, i, " ")
	}
	s.WriteString("\n")
	return s.String()
}

func inversionsData() []int {
	r := rand.New(rand.NewPCG(6, 6))
	a := make([]int, 200000)
	for i := range a {
		a[i] = r.IntN(100000) - 50000 // repeats on purpose
	}
	return a
}

func inversionsInput() string {
	a := inversionsData()
	var s strings.Builder
	fmt.Fprintln(&s, len(a))
	for _, v := range a {
		fmt.Fprint(&s, v, " ")
	}
	s.WriteString("\n")
	return s.String()
}

// inversionsWant counts inversions with a Fenwick tree over the values.
func inversionsWant() int64 {
	a := inversionsData()
	const off, size = 50001, 100002
	tree := make([]int64, size+1)
	var total int64
	for i := len(a) - 1; i >= 0; i-- {
		// count the values already seen (to the right) that are smaller
		for j := a[i] + off - 1; j > 0; j -= j & -j {
			total += tree[j]
		}
		for j := a[i] + off; j <= size; j += j & -j {
			tree[j]++
		}
	}
	return total
}

// eulerRingInput is a cycle through all n vertices, in a scrambled
// order so that the vertex numbers jump around.
func eulerRingInput(n int) string {
	r := rand.New(rand.NewPCG(7, 7))
	order := r.Perm(n)
	var s strings.Builder
	fmt.Fprintln(&s, n, n)
	for i := range order {
		fmt.Fprintln(&s, order[i]+1, order[(i+1)%n]+1)
	}
	return s.String()
}

// numberFSMLines are generated test lines for number_fsm: valid numbers
// built from the grammar, and near misses.
func numberFSMLines() []string {
	r := rand.New(rand.NewPCG(8, 8))
	digits := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte('0' + r.IntN(10))
		}
		return string(b)
	}
	var lines []string
	for i := 0; i < 400; i++ {
		var s string
		if r.IntN(3) == 0 {
			s += string("+-"[r.IntN(2)])
		}
		switch r.IntN(3) {
		case 0:
			s += digits(1 + r.IntN(5))
		case 1:
			s += digits(1+r.IntN(5)) + "." + digits(r.IntN(4))
		default:
			s += "." + digits(1+r.IntN(4))
		}
		if r.IntN(2) == 0 {
			s += string("eE"[r.IntN(2)])
			if r.IntN(2) == 0 {
				s += string("+-"[r.IntN(2)])
			}
			s += digits(1 + r.IntN(3))
		}
		if r.IntN(3) == 0 { // break it in one place
			b := []byte(s)
			at := r.IntN(len(b) + 1)
			junk := string(".eE+- x9"[r.IntN(8)])
			s = string(b[:at]) + junk + string(b[at:])
		}
		lines = append(lines, s)
	}
	lines = append(lines, strings.Repeat("7", 9997)+"e1") // 9,999 characters
	return lines
}

// validNumber is the number_fsm grammar, written as a regular expression's
// equivalent by hand: [+-]? (d+ (. d*)? | . d+) ([eE] [+-]? d+)?
func validNumber(s string) bool {
	i, n := 0, len(s)
	isDigit := func(c byte) bool { return c >= '0' && c <= '9' }
	if i < n && (s[i] == '+' || s[i] == '-') {
		i++
	}
	intDigits := 0
	for i < n && isDigit(s[i]) {
		i, intDigits = i+1, intDigits+1
	}
	fracDigits := 0
	if i < n && s[i] == '.' {
		i++
		for i < n && isDigit(s[i]) {
			i, fracDigits = i+1, fracDigits+1
		}
	}
	if intDigits == 0 && fracDigits == 0 {
		return false
	}
	if i < n && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < n && (s[i] == '+' || s[i] == '-') {
			i++
		}
		expDigits := 0
		for i < n && isDigit(s[i]) {
			i, expDigits = i+1, expDigits+1
		}
		if expDigits == 0 {
			return false
		}
	}
	return i == n
}

func numberFSMInput() string { return strings.Join(numberFSMLines(), "\n") + "\n" }

func numberFSMWant() string {
	var b strings.Builder
	for _, l := range numberFSMLines() {
		if validNumber(l) {
			b.WriteString("valid\n")
		} else {
			b.WriteString("invalid\n")
		}
	}
	return b.String()
}

// Programs for the cpu check.
const cpuSumProgram = `LOAD r0 100
LOAD r1 0
LOAD r2 1
JZ r0 7
ADD r1 r0
SUB r0 r2
JMP 3
PRINT r1
HALT
`

const cpuFibProgram = `LOAD r0 0
LOAD r1 1
LOAD r3 10
PRINT r0
LOAD r2 0
ADD r2 r0
ADD r2 r1
LOAD r0 0
ADD r0 r1
LOAD r1 0
ADD r1 r2
LOAD r2 1
SUB r3 r2
JZ r3 15
JMP 3
HALT
`

// rvWords is the large rv_decode case: a mix of the nine instructions,
// with random registers and immediates, and some words that are not
// among them.
func rvWords() []uint32 {
	r := rand.New(rand.NewPCG(9, 9))
	words := make([]uint32, 1000)
	for i := range words {
		rd, rs1, rs2 := uint32(r.IntN(32)), uint32(r.IntN(32)), uint32(r.IntN(32))
		switch r.IntN(4) {
		case 0, 1: // R-type
			f3 := []uint32{0, 0, 4, 6, 7}[r.IntN(5)]
			f7 := uint32(0)
			if f3 == 0 && r.IntN(2) == 0 {
				f7 = 0x20
			}
			words[i] = f7<<25 | rs2<<20 | rs1<<15 | f3<<12 | rd<<7 | 0x33
		case 2: // I-type
			f3 := []uint32{0, 4, 6, 7}[r.IntN(4)]
			imm := uint32(r.IntN(4096))
			words[i] = imm<<20 | rs1<<15 | f3<<12 | rd<<7 | 0x13
		default: // anything at all
			words[i] = r.Uint32()
		}
	}
	return words
}

func rvInput() string {
	var b strings.Builder
	for _, w := range rvWords() {
		fmt.Fprintf(&b, "%08x\n", w)
	}
	return b.String()
}

// rvDecode is the reference disassembler for the nine instructions.
func rvDecode(w uint32) string {
	opcode, rd, f3 := w&0x7f, w>>7&0x1f, w>>12&7
	rs1, rs2, f7 := w>>15&0x1f, w>>20&0x1f, w>>25
	switch opcode {
	case 0x33:
		names := map[uint32]string{0: "add", 4: "xor", 6: "or", 7: "and"}
		name, ok := names[f3]
		switch {
		case ok && f7 == 0:
		case f3 == 0 && f7 == 0x20:
			name = "sub"
		default:
			return "unknown"
		}
		return fmt.Sprintf("%s x%d, x%d, x%d", name, rd, rs1, rs2)
	case 0x13:
		names := map[uint32]string{0: "addi", 4: "xori", 6: "ori", 7: "andi"}
		name, ok := names[f3]
		if !ok {
			return "unknown"
		}
		return fmt.Sprintf("%s x%d, x%d, %d", name, rd, rs1, int32(w)>>20)
	}
	return "unknown"
}

func rvWant() string {
	var b strings.Builder
	for _, w := range rvWords() {
		b.WriteString(rvDecode(w) + "\n")
	}
	return b.String()
}

// bufferPoolOps is the trace of the big buffer_pool case: mostly hot
// pages with a long tail, a third of them writes, and some pinned
// stretches.
func bufferPoolOps() []string {
	r := rand.New(rand.NewPCG(10, 10))
	var ops []string
	var pinned []int
	for len(ops) < 1000000 {
		page := r.IntN(100)
		if r.IntN(5) == 0 {
			page = r.IntN(10000)
		}
		switch x := r.IntN(100); {
		case x < 2 && len(pinned) < 8:
			pinned = append(pinned, page)
			ops = append(ops, fmt.Sprintf("p %d", page))
		case x < 4 && len(pinned) > 0:
			ops = append(ops, fmt.Sprintf("u %d", pinned[0]))
			pinned = pinned[1:]
		case x < 35:
			ops = append(ops, fmt.Sprintf("w %d", page))
		default:
			ops = append(ops, fmt.Sprintf("r %d", page))
		}
	}
	return ops
}

func bufferPoolInput() string {
	return "64 summary\n" + strings.Join(bufferPoolOps(), "\n") + "\n"
}

// bufferPoolSim runs the buffer_pool rules over ops with the given number
// of frames and returns the summary line.
func bufferPoolSim(frames int, ops []string) string {
	type frame struct {
		page, pins int
		dirty      bool
		used       int
	}
	var pool []*frame
	find := func(page int) *frame {
		for _, f := range pool {
			if f.page == page {
				return f
			}
		}
		return nil
	}
	hits, misses, writes, clock := 0, 0, 0, 0
	for _, op := range ops {
		var kind string
		var page int
		fmt.Sscanf(op, "%s %d", &kind, &page)
		if kind == "u" {
			if f := find(page); f != nil && f.pins > 0 {
				f.pins--
			}
			continue
		}
		clock++
		f := find(page)
		if f != nil {
			hits++
		} else {
			misses++
			if len(pool) < frames {
				f = &frame{page: page}
				pool = append(pool, f)
			} else {
				var victim *frame
				for _, g := range pool {
					if g.pins == 0 && (victim == nil || g.used < victim.used) {
						victim = g
					}
				}
				if victim == nil {
					continue
				}
				if victim.dirty {
					writes++
				}
				*victim = frame{page: page}
				f = victim
			}
		}
		f.used = clock
		if kind == "w" {
			f.dirty = true
		}
		if kind == "p" {
			f.pins++
		}
	}
	return fmt.Sprintf("hits %d, misses %d, writes %d\n", hits, misses, writes)
}

func bufferPoolWant() string { return bufferPoolSim(64, bufferPoolOps()) }

// hashJoinData is the big hash_join case: two tables of 100,000 rows with
// keys drawn from two million values, so about 5,000 rows match.
func hashJoinData() (a, b []int64) {
	r := rand.New(rand.NewPCG(11, 11))
	for i := 0; i < 100000; i++ {
		a = append(a, int64(r.IntN(2000000))-1000000)
		b = append(b, int64(r.IntN(2000000))-1000000)
	}
	return a, b
}

func hashJoinInput() string {
	a, b := hashJoinData()
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d\n", len(a))
	for i, k := range a {
		fmt.Fprintf(&sb, "%d a%d\n", k, i)
	}
	fmt.Fprintf(&sb, "%d\n", len(b))
	for i, k := range b {
		fmt.Fprintf(&sb, "%d b%d\n", k, i)
	}
	return sb.String()
}

func hashJoinWant() string {
	a, b := hashJoinData()
	rows := map[int64][]int{}
	for i, k := range a {
		rows[k] = append(rows[k], i)
	}
	var sb strings.Builder
	for j, k := range b {
		for _, i := range rows[k] {
			fmt.Fprintf(&sb, "%d a%d b%d\n", k, i, j)
		}
	}
	return sb.String()
}

// ipString formats a 32-bit IPv4 address.
func ipString(a uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", a>>24, a>>16&255, a>>8&255, a&255)
}

type route struct {
	net  uint32
	bits int
	hop  string
}

// routesData is the big routes case: 100,000 random prefixes from /8 to
// /32, most of them inside a few busy blocks so that they nest, and
// 100,000 addresses, a third of them inside a route on purpose.
func routesData() ([]route, []uint32) {
	r := rand.New(rand.NewPCG(12, 12))
	var routes []route
	for i := 0; i < 100000; i++ {
		bits := 8 + r.IntN(25)
		addr := r.Uint32()
		if r.IntN(4) != 0 {
			addr = 10<<24 | addr&0x00FFFFFF // nested routes inside 10.0.0.0/8
		}
		routes = append(routes, route{addr, bits, fmt.Sprintf("r%d", i)})
	}
	var addrs []uint32
	for i := 0; i < 100000; i++ {
		a := r.Uint32()
		if i%3 == 0 {
			rt := routes[r.IntN(len(routes))]
			m := ^uint32(0) << (32 - rt.bits) // bits is 8 to 32
			a = rt.net&m | a&^m               // inside rt, at a random host
		} else if i%3 == 1 {
			a = 10<<24 | a&0x00FFFFFF
		}
		addrs = append(addrs, a)
	}
	return routes, addrs
}

func routesInput() string {
	routes, addrs := routesData()
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d\n", len(routes))
	for _, rt := range routes {
		fmt.Fprintf(&sb, "%s/%d %s\n", ipString(rt.net), rt.bits, rt.hop)
	}
	for _, a := range addrs {
		sb.WriteString(ipString(a) + "\n")
	}
	return sb.String()
}

func routesWant() string {
	routes, addrs := routesData()
	mask := func(bits int) uint32 {
		if bits == 0 {
			return 0
		}
		return ^uint32(0) << (32 - bits)
	}
	table := make([]map[uint32]string, 33)
	for i := range table {
		table[i] = map[uint32]string{}
	}
	for _, rt := range routes {
		table[rt.bits][rt.net&mask(rt.bits)] = rt.hop
	}
	var sb strings.Builder
	for _, a := range addrs {
		hop := "unreachable"
		for bits := 32; bits >= 0; bits-- {
			if h, ok := table[bits][a&mask(bits)]; ok {
				hop = h
				break
			}
		}
		sb.WriteString(hop + "\n")
	}
	return sb.String()
}

// retryWant is the retry harness's output when every attempt fails
// temporarily: waits of jitter(d), where the harness's jitter returns d / 3.
func retryWant(attempts int, base, limit int64) string {
	var sb strings.Builder
	d := base
	for i := 1; i < attempts; i++ {
		fmt.Fprintf(&sb, "jitter(%d) wait %d\n", d, d/3)
		if d > limit/2 {
			d = limit
		} else {
			d *= 2
		}
	}
	fmt.Fprintf(&sb, "%d attempts, result 1\n", attempts)
	return sb.String()
}

func fnv1a32(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func ringNodes(n int) []string {
	var out []string
	for i := 1; i <= n; i++ {
		out = append(out, fmt.Sprintf("n%d", i))
	}
	return out
}

func ringKeys(n int, prefix string) []string {
	var out []string
	for i := 0; i < n; i++ {
		out = append(out, fmt.Sprintf("%s%d", prefix, i*7919%1000003))
	}
	return out
}

func ringInput(v int, nodes, keys []string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d\n%d %s\n%d\n", v, len(nodes), strings.Join(nodes, " "), len(keys))
	for _, k := range keys {
		sb.WriteString(k + "\n")
	}
	return sb.String()
}

func ringWant(v int, nodes, keys []string) string {
	type point struct {
		hash uint32
		node int
	}
	var pts []point
	for n, name := range nodes {
		for i := 0; i < v; i++ {
			pts = append(pts, point{fnv1a32(fmt.Sprintf("%s#%d", name, i)), n})
		}
	}
	slices.SortStableFunc(pts, func(a, b point) int {
		if a.hash != b.hash {
			if a.hash < b.hash {
				return -1
			}
			return 1
		}
		return a.node - b.node
	})
	var sb strings.Builder
	for _, k := range keys {
		h := fnv1a32(k)
		i, _ := slices.BinarySearchFunc(pts, h, func(p point, h uint32) int {
			if p.hash < h {
				return -1
			}
			if p.hash > h {
				return 1
			}
			return 0
		})
		for i > 0 && pts[i-1].hash == h { // the first of equal hashes
			i--
		}
		if i == len(pts) {
			i = 0
		}
		sb.WriteString(nodes[pts[i].node] + "\n")
	}
	return sb.String()
}
