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
