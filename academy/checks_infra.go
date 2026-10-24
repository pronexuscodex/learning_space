package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Checked exercises for Stage 16 (AI infrastructure): the memory
// arithmetic of serving and training models, in C, with no GPU needed.
var infraChecks = []Check{
	{
		ID: "quantize", Stage: 16, Concept: "Quantisation & Number Formats",
		Title: "int8 quantisation, per row or per tensor",
		Task:  "./quantize MODE reads R and C on the first line, then R rows of C numbers, and quantises them to int8. MODE is row (one scale for each row) or tensor (one scale for the whole matrix). In double precision throughout: m is the largest absolute value in the row (or the matrix), scale = m / 127.0, and each q = round(w / scale) (C's round: halves away from zero), clamped to −127 … 127; when m is 0, the scale is 0 and every q is 0. Print one line per row: its scale with %.6e, then its q values as integers, separated by spaces. Then print max error E and mean error E, with %.3e, where each error is |w − scale × q| and the mean is their sum, added in row order, divided by R × C. R and C go up to 256 and 512. A MODE other than row or tensor gets a usage message on standard error and exit status 2; missing numbers, or R or C below 1, get a message on standard error and exit status 1, with nothing printed.",
		File:  "quantize.c",
		Starter: `/* Checked exercise: quantize (int8, as used to shrink LLM weights)
 *   printf '2 3\n0.5 -1.0 0.25\n2.0 0.1 -0.7\n' | ./quantize row
 *   7.874016e-03 64 -127 32
 *   1.574803e-02 127 6 -44
 *   max error 7.087e-03
 *   mean error 3.084e-03
 * Row 0: m = 1.0, scale = 1/127; 0.5 / scale = 63.5, which rounds to 64.
 * Dequantised, 64 * scale = 0.50394: an error of 0.0039.
 *   scale = m / 127.0       q = round(w / scale), clamped to -127..127
 * m is the largest |w| (fabs): a large negative weight counts too. With
 * MODE tensor, m is taken over the whole matrix, so one outlier sets
 * every row's scale. A row of zeros must not divide by zero.
 */
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Libs: []string{"-lm"},
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"row"}, Stdin: "2 3\n0.5 -1.0 0.25\n2.0 0.1 -0.7\n", Want: quantizeWant("2 3\n0.5 -1.0 0.25\n2.0 0.1 -0.7\n", true)},
			{Name: "rounding, not truncation", Args: []string{"row"}, Stdin: "1 5\n1.0 0.3 -0.3 0.5 -0.996\n", Want: quantizeWant("1 5\n1.0 0.3 -0.3 0.5 -0.996\n", true),
				Explain: "0.3 × 127 = 38.1 and 0.5 × 127 = 63.5: round gives 38 and 64, a cast to int gives 38 and 63. Truncation always moves towards zero, so its errors do not cancel: they add up in every dot product."},
			{Name: "a negative outlier", Args: []string{"row"}, Stdin: "2 4\n0.1 -2.0 0.3 0.05\n-0.2 0.4 -0.1 0.2\n", Want: quantizeWant("2 4\n0.1 -2.0 0.3 0.05\n-0.2 0.4 -0.1 0.2\n", true),
				Explain: "Row 0's largest magnitude is −2.0. Taking the largest value (0.3) instead of the largest absolute value makes −2.0 / scale = −847, far outside int8."},
			{Name: "one scale for the whole matrix", Args: []string{"tensor"}, Gen: generated(func() (string, string) { return quantizeCase(8, 16, 3, 41, false) }),
				Explain: "Three outliers of about 1.0 among weights of about 0.02 set the single scale near 0.008, so most weights become −5 … 5, barely 4 of the 8 bits: compare the errors with the same matrix in row mode."},
			{Name: "the same matrix, one scale per row", Args: []string{"row"}, Gen: generated(func() (string, string) { return quantizeCase(8, 16, 3, 41, true) }),
				Explain: "Only the rows that hold an outlier keep a coarse scale; every other row gets the full −127 … 127 range, and the mean error drops to a third."},
			{Name: "a row of zeros", Args: []string{"row"}, Stdin: "2 3\n0 0 0\n0.5 0.25 -0.125\n", Want: quantizeWant("2 3\n0 0 0\n0.5 0.25 -0.125\n", true),
				Explain: "m = 0 gives a scale of 0, and 0 / 0 is NaN: a row of zeros must be special-cased (pruned weights and padding are real rows of zeros)."},
			{Name: "an unknown mode", Args: []string{"int4"}, Stdin: "1 1\n1\n", WantExit: 2, Stderr: true},
			{Name: "a missing number", Args: []string{"row"}, Stdin: "2 2\n1 2\n3\n", WantExit: 1, Stderr: true},
			{Name: "256 × 512 weights", Args: []string{"row"}, Gen: generated(func() (string, string) { return quantizeCase(256, 512, 40, 42, true) })},
		},
		Hints: []string{
			"Read all R × C numbers into one malloc'd array of doubles first, checking that every scanf returned 1, so that an error prints nothing.",
			"Compute the scale for each row (or once for the matrix) before printing that row; keep the error sum and the maximum in doubles as you go.",
			"Print q as an int (%d): printing a double that is −0.0 with %.0f gives -0.",
		},
	},
	{
		ID: "kv_blocks", Stage: 16, Concept: "KV Cache, Batching & PagedAttention",
		Title: "A paged KV-cache allocator",
		Task:  "./kv_blocks BLOCKS SIZE manages a KV cache of BLOCKS blocks (numbered 0 to BLOCKS − 1) that hold SIZE tokens each, as vLLM does. Commands come one per line on standard input. new S starts an empty sequence named S. append S N adds N tokens to S: if its last block is only partly full and is shared with another sequence, S first gets a private copy of it (copy-on-write), and then new blocks are added as the tokens need them; every block is taken from the free blocks, lowest number first (the copy first). If there are not enough free blocks for all of it, print out of memory S and change nothing. fork S T creates T sharing all of S's blocks and its length (prefix sharing). free S ends S: each of its blocks loses one user and becomes free when it has none. show S prints S len L blocks followed by its block numbers in order. stats prints free F shared X, where X counts the blocks with two or more users. An unknown command, sequence or number, or a sequence name that already exists, gets a message on standard error and exit status 1 at once. BLOCKS goes up to 4,096 and SIZE up to 1,024; names are up to 15 characters, with up to 1,000 sequences alive; there can be 50,000 commands. BLOCKS or SIZE that is not a whole number from 1 to those limits gets a usage message and exit status 2.",
		File:  "kv_blocks.c",
		Starter: `/* Checked exercise: kv_blocks (PagedAttention's block manager)
 *   ./kv_blocks 8 4         8 blocks of 4 tokens
 *   new a
 *   append a 6              blocks 0 1   (4 + 2 tokens)
 *   fork a b                b shares 0 and 1
 *   append b 1              block 1 is shared and partly full:
 *                           b copies it into 2, then appends there
 *   show a                  a len 6 blocks 0 1
 *   show b                  b len 7 blocks 0 2
 *   stats                   free 5 shared 1     (block 0 has two users)
 * Keep, for every block, its number of users (0 means free). For every
 * sequence: its name, its length, and its list of block numbers.
 *   blocks needed for L tokens = (L + SIZE - 1) / SIZE
 *   the last block is partly full when L % SIZE != 0
 * Work out everything an append needs before changing anything.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"8", "4"}, Stdin: "new a\nappend a 6\nfork a b\nappend b 1\nshow a\nshow b\nstats\n", Want: kvWant(8, 4, "new a\nappend a 6\nfork a b\nappend b 1\nshow a\nshow b\nstats\n")},
			{Name: "copy-on-write", Args: []string{"16", "4"}, Stdin: "new p\nappend p 10\nfork p c1\nfork p c2\nappend c1 3\nappend c2 1\nappend p 2\nshow p\nshow c1\nshow c2\nstats\n", Want: kvWant(16, 4, "new p\nappend p 10\nfork p c1\nfork p c2\nappend c1 3\nappend c2 1\nappend p 2\nshow p\nshow c1\nshow c2\nstats\n"),
				Explain: "Three sequences share a prompt of 10 tokens: blocks 0 and 1 are full and stay shared, block 2 holds 2 tokens. c1 and c2 each copy block 2 before writing; by the time p appends, it is block 2's only user and writes in place. Without the copy, c1's tokens would overwrite the prompt everyone else is reading."},
			{Name: "a full last block is not copied", Args: []string{"8", "4"}, Stdin: "new a\nappend a 8\nfork a b\nappend b 1\nshow a\nshow b\nstats\n", Want: kvWant(8, 4, "new a\nappend a 8\nfork a b\nappend b 1\nshow a\nshow b\nstats\n"),
				Explain: "a's two blocks are full, so b's new token goes into a fresh block and nothing needs copying: b keeps sharing both."},
			{Name: "free keeps shared blocks", Args: []string{"8", "2"}, Stdin: "new a\nappend a 4\nfork a b\nfree a\nstats\nnew c\nappend c 3\nshow b\nshow c\nstats\n", Want: kvWant(8, 2, "new a\nappend a 4\nfork a b\nfree a\nstats\nnew c\nappend c 3\nshow b\nshow c\nstats\n"),
				Explain: "Freeing a only drops a user from blocks 0 and 1: b still reads them. Freeing them outright would give them to c, and b's keys and values would be overwritten."},
			{Name: "out of memory changes nothing", Args: []string{"4", "4"}, Stdin: "new a\nappend a 9\nnew b\nappend b 5\nstats\nappend b 4\nshow b\nstats\n", Want: kvWant(4, 4, "new a\nappend a 9\nnew b\nappend b 5\nstats\nappend b 4\nshow b\nstats\n"),
				Explain: "b's 5 tokens need two blocks and only one is free: the append must fail as a whole, with no block taken. Taking the one free block and then failing would leak it."},
			{Name: "lowest free block first", Args: []string{"8", "1"}, Stdin: "new a\nappend a 3\nnew b\nappend b 2\nfree a\nnew c\nappend c 4\nshow b\nshow c\n", Want: kvWant(8, 1, "new a\nappend a 3\nnew b\nappend b 2\nfree a\nnew c\nappend c 4\nshow b\nshow c\n")},
			{Name: "a sequence that does not exist", Args: []string{"8", "4"}, Stdin: "new a\nappend z 1\nshow a\n", WantExit: 1, Stderr: true,
				Explain: "The program stops at the bad command: show a never runs."},
			{Name: "a name used twice", Args: []string{"8", "4"}, Stdin: "new a\nnew a\n", WantExit: 1, Stderr: true},
			{Name: "a bad block count", Args: []string{"0", "4"}, Stdin: "", WantExit: 2, Stderr: true},
			{Name: "fifty thousand commands", Args: []string{"4096", "16"}, Gen: generated(kvManyCase),
				Explain: "Hundreds of requests arriving, forking, growing and finishing, as on a busy inference server."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"An array users[BLOCKS] of ints is the whole allocator: finding the lowest free block is a loop for the first 0.",
			"Each sequence needs its block list to grow: a malloc'd int array of BLOCKS entries is simple and big enough.",
			"For append, count first: copy = (L % SIZE != 0 && users[last] > 1), extra = blocks for L + N minus blocks for L. Only when copy + extra blocks are free, take them.",
		},
	},
	{
		ID: "allreduce", Stage: 16, Concept: "Parallelism Across GPUs: Data, Tensor & Pipeline",
		Title: "Ring all-reduce",
		Task:  "Read N and L on the first line, then N lines of L integers: the gradient of each of N workers. Simulate the ring all-reduce that sums them. Split each vector into N chunks, chunk k being the elements from k × L / N up to (k + 1) × L / N (integer division, so sizes can differ by one). Reduce-scatter: in step s = 0 … N − 2, every worker i sends its chunk (i − s) mod N to worker (i + 1) mod N, which adds it into its own copy of that chunk. All-gather: in step s = 0 … N − 2, every worker i sends its chunk (i + 1 − s) mod N to worker (i + 1) mod N, which replaces its copy. Print after reduce-scatter, then each worker as worker I: and its L values; then after all-gather and each worker the same way; then worker I sent E values for each worker, E being the total size of the chunks it sent. Values fit in 32 bits but their sums may not. N goes up to 64 and L up to 2,000. Missing numbers, or N or L below 1, get a message on standard error and exit status 1, with nothing printed.",
		File:  "allreduce.c",
		Starter: `/* Checked exercise: allreduce (how data-parallel training sums gradients)
 *   input:  3 3          after reduce-scatter
 *           1 2 3        worker 0: 1 15 12
 *           4 5 6        worker 1: 5 5 18
 *           7 8 9        worker 2: 12 13 9
 *                        after all-gather
 *                        worker 0: 12 15 18
 *                        ...                 (all three the same)
 *                        worker 0 sent 4 values
 *                        ...
 * After reduce-scatter, worker i holds the complete sum of chunk
 * (i + 1) mod N; all-gather passes those complete chunks round the ring.
 * In C, (i - s) % N is negative when i < s: write (i - s + N) % N.
 * Each worker sends only 2 (N - 1) / N of its vector, whatever N is.
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: "3 3\n1 2 3\n4 5 6\n7 8 9\n", Want: allreduceWant("3 3\n1 2 3\n4 5 6\n7 8 9\n"),
				Explain: "Each worker holds a different partly summed vector after reduce-scatter: worker 0 has the full sum of chunk 1 (15), worker 1 of chunk 2 (18), worker 2 of chunk 0 (12)."},
			{Name: "four workers", Stdin: "4 8\n1 0 0 0 0 0 0 1\n0 2 0 0 0 0 2 0\n0 0 3 0 0 3 0 0\n0 0 0 4 4 0 0 0\n", Want: allreduceWant("4 8\n1 0 0 0 0 0 0 1\n0 2 0 0 0 0 2 0\n0 0 3 0 0 3 0 0\n0 0 0 4 4 0 0 0\n"),
				Explain: "With 4 workers there are 3 steps of each phase; the chunk a worker sends moves back by one each step. Sending (i + s) instead of (i − s) passes on a chunk the worker never received."},
			{Name: "uneven chunks", Gen: generated(func() (string, string) { return allreduceCase(3, 7, 100, 51) }),
				Explain: "7 elements in 3 chunks: 0 … 1, 2 … 3, 4 … 6 (sizes 2, 2, 3). Assuming L / N elements per chunk drops the last one."},
			{Name: "sums beyond 32 bits", Gen: generated(func() (string, string) { return allreduceCase(5, 10, 2000000000, 52) }),
				Explain: "Five values near 2 × 10⁹ add up to 10¹⁰, more than an int holds (2,147,483,647): use long long for the sums."},
			{Name: "one worker", Stdin: "1 3\n5 6 7\n", Want: allreduceWant("1 3\n5 6 7\n"),
				Explain: "With one worker there are no steps at all, and it sends nothing."},
			{Name: "a missing number", Stdin: "2 2\n1 2\n3\n", WantExit: 1, Stderr: true},
			{Name: "sixteen workers", Gen: generated(func() (string, string) { return allreduceCase(16, 200, 1000000, 53) }),
				Explain: "Every worker sends 2 × 15 / 16 of 200 values: 375, almost twice its vector, however many workers join."},
		},
		Hints: []string{
			"Keep N vectors of long long in one malloc'd array: v[i * L + k] is element k of worker i.",
			"A function giving chunk c's start (c * L / N) and end ((c + 1) * L / N) keeps the steps short.",
			"Within one step, the chunk worker i receives is never the one it sends, so updating in place is safe.",
		},
	},
}

// quantizeWant quantises exactly as the task describes.
func quantizeWant(in string, perRow bool) string {
	f := strings.Fields(in)
	r, _ := strconv.Atoi(f[0])
	c, _ := strconv.Atoi(f[1])
	w := make([]float64, r*c)
	for i := range w {
		w[i], _ = strconv.ParseFloat(f[2+i], 64)
	}
	maxAbs := func(xs []float64) float64 {
		m := 0.0
		for _, x := range xs {
			m = math.Max(m, math.Abs(x))
		}
		return m
	}
	all := maxAbs(w)
	var b strings.Builder
	maxErr, sum := 0.0, 0.0
	for i := 0; i < r; i++ {
		row := w[i*c : (i+1)*c]
		m := all
		if perRow {
			m = maxAbs(row)
		}
		scale := m / 127.0
		fmt.Fprintf(&b, "%.6e", scale)
		for _, x := range row {
			q := 0
			if m > 0 {
				q = int(math.Max(-127, math.Min(127, math.Round(x/scale))))
			}
			fmt.Fprintf(&b, " %d", q)
			e := math.Abs(x - scale*float64(q))
			maxErr = math.Max(maxErr, e)
			sum += e
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "max error %.3e\nmean error %.3e\n", maxErr, sum/float64(r*c))
	return b.String()
}

// quantizeCase is an r × c matrix of small normal weights with a few
// large outliers, as real LLM weight matrices have.
func quantizeCase(r, c, outliers int, seed int64, perRow bool) (string, string) {
	rng := rand.New(rand.NewSource(seed))
	w := make([]float64, r*c)
	for i := range w {
		w[i] = rng.NormFloat64() * 0.02
	}
	for k := 0; k < outliers; k++ {
		w[rng.Intn(len(w))] = (0.8 + 0.4*rng.Float64()) * float64(1-2*rng.Intn(2))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d %d\n", r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if j > 0 {
				b.WriteString(" ")
			}
			b.WriteString(strconv.FormatFloat(w[i*c+j], 'f', 6, 64))
		}
		b.WriteString("\n")
	}
	in := b.String()
	return in, quantizeWant(in, perRow)
}

// kvWant runs the block manager the task describes, stopping at the
// first bad command as the program must.
func kvWant(blocks, size int, in string) string {
	users := make([]int, blocks)
	type seq struct {
		n      int
		blocks []int
	}
	seqs := map[string]*seq{}
	var b strings.Builder
	need := func(n int) int { return (n + size - 1) / size }
	free := func() int {
		f := 0
		for _, u := range users {
			if u == 0 {
				f++
			}
		}
		return f
	}
	take := func() int {
		for i, u := range users {
			if u == 0 {
				users[i] = 1
				return i
			}
		}
		panic("no free block")
	}
	for _, line := range strings.Split(strings.TrimSpace(in), "\n") {
		f := strings.Fields(line)
		switch {
		case len(f) == 2 && f[0] == "new" && seqs[f[1]] == nil:
			seqs[f[1]] = &seq{}
		case len(f) == 3 && f[0] == "append" && seqs[f[1]] != nil:
			s := seqs[f[1]]
			n, _ := strconv.Atoi(f[2])
			copyLast := s.n%size != 0 && users[s.blocks[len(s.blocks)-1]] > 1
			extra := need(s.n+n) - need(s.n)
			cost := extra
			if copyLast {
				cost++
			}
			if cost > free() {
				fmt.Fprintf(&b, "out of memory %s\n", f[1])
				continue
			}
			if copyLast {
				last := len(s.blocks) - 1
				users[s.blocks[last]]--
				s.blocks[last] = take()
			}
			for k := 0; k < extra; k++ {
				s.blocks = append(s.blocks, take())
			}
			s.n += n
		case len(f) == 3 && f[0] == "fork" && seqs[f[1]] != nil && seqs[f[2]] == nil:
			s := seqs[f[1]]
			for _, blk := range s.blocks {
				users[blk]++
			}
			seqs[f[2]] = &seq{n: s.n, blocks: append([]int(nil), s.blocks...)}
		case len(f) == 2 && f[0] == "free" && seqs[f[1]] != nil:
			for _, blk := range seqs[f[1]].blocks {
				users[blk]--
			}
			delete(seqs, f[1])
		case len(f) == 2 && f[0] == "show" && seqs[f[1]] != nil:
			s := seqs[f[1]]
			fmt.Fprintf(&b, "%s len %d blocks", f[1], s.n)
			for _, blk := range s.blocks {
				fmt.Fprintf(&b, " %d", blk)
			}
			b.WriteString("\n")
		case len(f) == 1 && f[0] == "stats":
			shared := 0
			for _, u := range users {
				if u > 1 {
					shared++
				}
			}
			fmt.Fprintf(&b, "free %d shared %d\n", free(), shared)
		default:
			return b.String()
		}
	}
	return b.String()
}

// kvLive is how many requests kvManyCase keeps alive at most.
const kvLive = 280

// kvManyCase simulates a busy server: requests with shared system
// prompts arrive, fork for parallel samples, grow a few tokens at a time
// and finish.
func kvManyCase() (string, string) {
	rng := rand.New(rand.NewSource(61))
	var b strings.Builder
	var live []string
	length := map[string]int{} // as if no append ran out of memory
	next := 0
	start := func() string {
		name := fmt.Sprintf("r%d", next)
		next++
		live = append(live, name)
		return name
	}
	end := func(k int) {
		fmt.Fprintf(&b, "free %s\n", live[k])
		live = append(live[:k], live[k+1:]...)
	}
	for lines := 0; lines < 50000; lines++ {
		switch r := rng.Intn(100); {
		case len(live) < 2 || r < 6 && len(live) < kvLive:
			name := start()
			n := 1 + rng.Intn(300)
			fmt.Fprintf(&b, "new %s\nappend %s %d\n", name, name, n)
			length[name] = n
			lines++
		case r < 9 && len(live) < kvLive:
			from := live[rng.Intn(len(live))]
			name := start()
			fmt.Fprintf(&b, "fork %s %s\n", from, name)
			length[name] = length[from]
		case r < 11:
			end(rng.Intn(len(live)))
		case r < 12:
			b.WriteString("stats\n")
		case r < 14:
			fmt.Fprintf(&b, "show %s\n", live[rng.Intn(len(live))])
		default:
			k := rng.Intn(len(live))
			if length[live[k]] > 500 { // a finished answer
				end(k)
				continue
			}
			n := 1 + rng.Intn(20)
			fmt.Fprintf(&b, "append %s %d\n", live[k], n)
			length[live[k]] += n
		}
	}
	b.WriteString("stats\n")
	in := b.String()
	return in, kvWant(4096, 16, in)
}

// allreduceWant simulates the ring exactly as the task describes.
func allreduceWant(in string) string {
	f := strings.Fields(in)
	n, _ := strconv.Atoi(f[0])
	l, _ := strconv.Atoi(f[1])
	v := make([]int64, n*l)
	for i := range v {
		v[i], _ = strconv.ParseInt(f[2+i], 10, 64)
	}
	lo := func(c int) int { return c * l / n }
	hi := func(c int) int { return (c + 1) * l / n }
	sent := make([]int, n)
	var b strings.Builder
	dump := func(title string) {
		b.WriteString(title + "\n")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "worker %d:", i)
			for k := 0; k < l; k++ {
				fmt.Fprintf(&b, " %d", v[i*l+k])
			}
			b.WriteString("\n")
		}
	}
	for s := 0; s < n-1; s++ {
		for i := 0; i < n; i++ {
			c, j := ((i-s)%n+n)%n, (i+1)%n
			for k := lo(c); k < hi(c); k++ {
				v[j*l+k] += v[i*l+k]
			}
			sent[i] += hi(c) - lo(c)
		}
	}
	dump("after reduce-scatter")
	for s := 0; s < n-1; s++ {
		for i := 0; i < n; i++ {
			c, j := ((i+1-s)%n+n)%n, (i+1)%n
			for k := lo(c); k < hi(c); k++ {
				v[j*l+k] = v[i*l+k]
			}
			sent[i] += hi(c) - lo(c)
		}
	}
	dump("after all-gather")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "worker %d sent %d values\n", i, sent[i])
	}
	return b.String()
}

// allreduceCase is n workers' random gradients of length l, values in
// −max … max.
func allreduceCase(n, l int, max int64, seed int64) (string, string) {
	rng := rand.New(rand.NewSource(seed))
	var b strings.Builder
	fmt.Fprintf(&b, "%d %d\n", n, l)
	for i := 0; i < n; i++ {
		for k := 0; k < l; k++ {
			if k > 0 {
				b.WriteString(" ")
			}
			x := rng.Int63n(max + 1)
			if max >= 1000000000 {
				x = max - rng.Int63n(1000) // near the top, so the sums overflow an int
			} else if rng.Intn(2) == 0 {
				x = -x
			}
			fmt.Fprintf(&b, "%d", x)
		}
		b.WriteString("\n")
	}
	in := b.String()
	return in, allreduceWant(in)
}
