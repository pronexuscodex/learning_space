package main

// Checked exercises for the Foundations track's theory stages: data
// structures and algorithms (Stage 2), discrete mathematics (Stage 3) and
// computer architecture (Stage 4). Most read standard input and are timed
// on inputs big enough that only the efficient idea passes.

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
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
