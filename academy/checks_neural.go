package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Checked exercises for Stage 15 (neural networks from scratch): the
// machinery under every framework, in C.
var neuralChecks = []Check{
	{
		ID: "autodiff", Stage: 15, Concept: "Computational Graphs & Reverse-Mode Autodiff",
		Title: "Reverse-mode automatic differentiation",
		Task:  "Read a straight-line program from standard input, one statement per line: input NAME VALUE declares an input, and NAME = OP A [B] computes a value, where OP is add, sub or mul (two arguments) or tanh, exp or relu (one), and every argument is a name defined on an earlier line. The output is the last name assigned. Evaluate the program forwards, then run reverse-mode differentiation backwards and print value V (the output) and then grad NAME G for every input, in the order they were declared, all with %.6f; a gradient that is exactly zero prints as 0.000000, never -0.000000. A name used before it is defined, defined twice, an unknown OP or a program with no assignment gets a message on standard error and exit status 1, with nothing printed. Programs can have 20,000 statements.",
		File:  "autodiff.c",
		Starter: `/* Checked exercise: autodiff
 *   input x 3          output:  value 15.000000
 *   input w 2                   grad x 5.000000
 *   a = mul x w                 grad w 6.000000
 *   b = add a x
 *   y = add b a        y = 2xw + x: dy/dx = 2w + 1 = 5, dy/dw = 2x = 6
 * Each statement is a node: its op, its argument nodes, its value, and
 * its gradient (start every gradient at 0). Forwards: compute values in
 * order. Backwards: set the output's gradient to 1, then visit the
 * statements in reverse order (the program order is already a
 * topological order) and ADD each node's contribution to its arguments:
 *   add: g_a += g          mul: g_a += g * b      tanh: g_a += g * (1 - t*t)
 *   sub: g_a += g, g_b -= g                       exp:  g_a += g * e^a
 *   relu: g_a += g if a > 0
 * += matters: a value used twice gets gradient along both paths.
 * Then print 0.0 + g for each input: adding 0.0 turns -0.0 into 0.0.
 */
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Libs: []string{"-lm"},
		Cases: []CheckCase{
			{Name: "a value used twice", Stdin: "input x 3\nw = mul x x\ny = add w x\n", Want: autodiffWant("input x 3\nw = mul x x\ny = add w x\n"),
				Explain: "y = x² + x, so dy/dx = 2x + 1 = 7. x reaches y along three edges (twice through the mul, once through the add): gradients that are assigned with = instead of added with += give the wrong answer."},
			{Name: "the starter's example", Stdin: "input x 3\ninput w 2\na = mul x w\nb = add a x\ny = add b a\n", Want: autodiffWant("input x 3\ninput w 2\na = mul x w\nb = add a x\ny = add b a\n")},
			{Name: "a tanh neuron", Stdin: "input x1 0.5\ninput x2 -1.5\ninput w1 0.8\ninput w2 0.3\ninput b 0.1\np = mul x1 w1\nq = mul x2 w2\ns = add p q\nz = add s b\ny = tanh z\n", Want: autodiffWant("input x1 0.5\ninput x2 -1.5\ninput w1 0.8\ninput w2 0.3\ninput b 0.1\np = mul x1 w1\nq = mul x2 w2\ns = add p q\nz = add s b\ny = tanh z\n"),
				Explain: "tanh's derivative is 1 − tanh², computed from the value the forward pass already has."},
			{Name: "subtraction and exp", Stdin: "input a 1.5\ninput b 0.5\nd = sub a b\ne = exp d\nf = mul e b\n", Want: autodiffWant("input a 1.5\ninput b 0.5\nd = sub a b\ne = exp d\nf = mul e b\n"),
				Explain: "sub sends −g to its second argument; b gets both that and the gradient through the final mul."},
			{Name: "a dead ReLU", Stdin: "input x -2\ninput w 3\nz = mul x w\ny = relu z\n", Want: autodiffWant("input x -2\ninput w 3\nz = mul x w\ny = relu z\n"),
				Explain: "z = −6, so the ReLU passes no gradient back: both gradients are exactly zero. A ReLU that stays negative for every input never learns again, which is why it is called dead."},
			{Name: "an input that is never used", Stdin: "input x 2\ninput unused 5\ny = mul x x\n", Want: autodiffWant("input x 2\ninput unused 5\ny = mul x x\n")},
			{Name: "a name used before it is defined", Stdin: "input x 1\ny = add x z\n", WantExit: 1, Stderr: true},
			{Name: "an unknown operation", Stdin: "input x 1\ny = sqrt x\n", WantExit: 1, Stderr: true},
			{Name: "twenty thousand statements", Gen: generated(autodiffManyCase),
				Explain: "A deep chain that keeps reusing earlier values: a single backward pass in reverse program order gets every gradient. Re-running the forward pass per input, or searching the whole program for each name, is far too slow."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Read each line with fgets and parse it with sscanf: \"input %15s %lf\" or \"%15s = %15s %15s %15s\" (check how many fields matched).",
			"Keep the statements in an array of structs; look names up in a small hash table (or, at first, a linear search, which is fine for short programs).",
			"The gradient array is separate from the values and starts at zero; only the output's gradient starts at 1.",
		},
	},
	{
		ID: "attention", Stage: 15, Concept: "Attention & the Transformer",
		Title: "Causal self-attention",
		Task:  "Read n and d on the first line, then the n rows of Q, the n rows of K and the n rows of V, each row d numbers. Compute causal scaled dot-product attention: for each position i, the scores s_j = (q_i · k_j) / √d for j = 0 … i only (a token never sees the future), the weights softmax(s) computed stably (subtract the largest score), and output_i = Σ_j weight_j · v_j. Print the n output rows, each value with %.4f separated by spaces. n and d go up to 512 and 128. Missing numbers, or n or d below 1, get a message on standard error and exit status 1, with nothing printed.",
		File:  "attention.c",
		Starter: `/* Checked exercise: attention (one causal head)
 *   input:   2 2        Q rows, then K rows, then V rows
 *            1 0
 *            0 1
 *            1 0
 *            1 1
 *            1 2
 *            3 4
 *   output:  1.0000 2.0000          row 0 sees only itself
 *            2.3395 3.3395          row 1 leans towards v1
 * For each i:
 *   s_j = dot(q_i, k_j) / sqrt(d)   for j = 0 .. i   (the causal mask)
 *   m = max s_j;  w_j = exp(s_j - m);  w_j /= sum of w   (stable softmax)
 *   out_i = sum over j of w_j * v_j
 * Scores of 1,000 or more appear in the tests: exp(1000) is inf.
 */
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
`,
		Libs: []string{"-lm"},
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: "2 2\n1 0\n0 1\n1 0\n1 1\n1 2\n3 4\n", Want: attentionWant("2 2\n1 0\n0 1\n1 0\n1 1\n1 2\n3 4\n"),
				Explain: "Row 1's query (0, 1) matches k1 = (1, 1) better than k0 = (1, 0): scores 0.7071 and 0, weights 0.67 and 0.33, so the output leans towards v1."},
			{Name: "the future is masked", Gen: generated(func() (string, string) { return attentionCase(6, 4, 1, 15) }),
				Explain: "Row 0 can only see itself, so its output is exactly v0; every row i mixes only v0 … vi."},
			{Name: "scaled by the square root of d", Gen: generated(func() (string, string) { return attentionCase(8, 64, 1, 16) }),
				Explain: "With d = 64 the raw dot products are about 8 times too large; without dividing by √d = 8 the softmax picks one token and ignores the rest."},
			{Name: "huge scores", Gen: generated(func() (string, string) { return attentionCase(6, 4, 400, 17) }),
				Explain: "Scores reach the thousands: exp of them is inf, and inf / inf is NaN. Subtracting the largest score first makes the biggest term exp(0) = 1."},
			{Name: "a missing number", Stdin: "2 2\n1 0\n0 1\n1 0\n", WantExit: 1, Stderr: true},
			{Name: "512 tokens", Gen: generated(func() (string, string) { return attentionCase(512, 64, 1, 18) }),
				Explain: "512 × 512 scores of 64 terms each: 17 million multiplications, under a second in C."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Read everything with scanf(\"%lf\", ...) into three n × d arrays allocated with malloc; check that each read returned 1.",
			"A scores buffer of n doubles is enough: fill it for one row i, take its max, exponentiate, sum, then mix the V rows.",
			"The causal mask is just the loop bound: j runs from 0 to i, not to n − 1.",
		},
	},
	{
		ID: "bpe", Stage: 15, Concept: "Tokenisation & Embeddings",
		Title: "Byte-pair encoding",
		Task:  "./bpe MERGES reads text from standard input and learns a byte-pair encoding. Start with one token per byte (ids 0 to 255). Then, up to MERGES times: count every adjacent pair of tokens (at every position, overlaps included), pick the pair with the highest count (on a tie, the smallest left id, then the smallest right id), stop early if no pair occurs at least twice, and replace its occurrences from left to right, without overlapping, by a new token whose id is 256 for the first merge, 257 for the next, and so on. Print each merge as ID = LEFT RIGHT (COUNT), then tokens N, the number of tokens left. Inputs are up to 100 KB, with up to 300 merges. MERGES that is not a whole number from 0 to 1,000 gets a message on standard error and exit status 2.",
		File:  "bpe.c",
		Starter: `/* Checked exercise: bpe (byte-pair encoding, as in GPT's tokenizers)
 *   printf 'aaabdaaabac' | ./bpe 3
 *   256 = 97 97 (4)        "aa" occurs at 4 positions (overlaps count)
 *   257 = 97 98 (2)        "ab" ties with "aa"+"a"; the smaller left id wins
 *   258 = 256 257 (2)      then "aa" + "ab": "aaab"
 *   tokens 5
 * Tokens are ints in an array (start: the input's bytes). With at most
 * 256 + 1,000 ids, a 2D array of pair counts works:
 *   static int count[1256][1256];
 * Each merge: zero the counts, count the pairs, find the best one,
 * then rewrite the array in place: read i, write w; when tokens[i],
 * tokens[i+1] is the pair, write the new id and skip two.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the classic example", Args: []string{"3"}, Stdin: "aaabdaaabac", Want: bpeWant("aaabdaaabac", 3),
				Explain: "The example from the BPE literature: aa first (it occurs at 4 positions, counting the overlap in aaa); then ab, which ties with aa + a at 2 and wins on the smaller left id (97 < 256); then aa + ab, which is aaab."},
			{Name: "overlapping pairs", Args: []string{"1"}, Stdin: "aaaa", Want: bpeWant("aaaa", 1),
				Explain: "aa is counted at 3 positions, but replacing left to right without overlap turns aaaa into two tokens, not three."},
			{Name: "ties go to the smallest pair", Args: []string{"2"}, Stdin: "cdcdabab", Want: bpeWant("cdcdabab", 2),
				Explain: "ab and cd both occur twice: the smaller left id, a (97) against c (99), goes first."},
			{Name: "stopping early", Args: []string{"10"}, Stdin: "abcdefg", Want: bpeWant("abcdefg", 10),
				Explain: "No pair occurs twice: no merge is worth making, so none is printed and all 7 bytes stay tokens."},
			{Name: "not a number", Args: []string{"lots"}, Stdin: "ab", WantExit: 2, Stderr: true},
			{Name: "English text", Args: []string{"300"}, Gen: generated(bpeEnglishCase),
				Explain: "About 100 KB of English: the first merges are \"e \", \"th\", \"s \" and the like. 300 passes over 100,000 tokens: counting with a 2D array keeps each pass fast."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Read the whole input with fread into a buffer, then copy each byte into an int array.",
			"Find the best pair by scanning the count table in order of left id, then right id, keeping only a strictly higher count: that is the tie rule.",
			"Only pairs that occur at least twice are merged; stop when the best count is below 2.",
		},
	},
}

// autodiffRun evaluates and differentiates a valid program, returning the
// output value and each input's name and gradient, in declaration order.
func autodiffRun(prog string) (float64, []string, []float64) {
	type node struct {
		op   string
		a, b int
		val  float64
	}
	var nodes []node
	idx := map[string]int{}
	var inputs []int
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(prog), "\n") {
		f := strings.Fields(line)
		if f[0] == "input" {
			v, _ := strconv.ParseFloat(f[2], 64)
			idx[f[1]] = len(nodes)
			inputs, names = append(inputs, len(nodes)), append(names, f[1])
			nodes = append(nodes, node{op: "input", val: v})
			continue
		}
		n := node{op: f[2], a: idx[f[3]], b: -1}
		x := nodes[n.a].val
		switch n.op {
		case "add", "sub", "mul":
			n.b = idx[f[4]]
			y := nodes[n.b].val
			n.val = map[string]float64{"add": x + y, "sub": x - y, "mul": x * y}[n.op]
		case "tanh":
			n.val = math.Tanh(x)
		case "exp":
			n.val = math.Exp(x)
		case "relu":
			n.val = math.Max(x, 0)
		}
		idx[f[0]] = len(nodes)
		nodes = append(nodes, n)
	}
	g := make([]float64, len(nodes))
	g[len(nodes)-1] = 1
	for i := len(nodes) - 1; i >= 0; i-- {
		n := nodes[i]
		switch n.op {
		case "add":
			g[n.a] += g[i]
			g[n.b] += g[i]
		case "sub":
			g[n.a] += g[i]
			g[n.b] -= g[i]
		case "mul":
			g[n.a] += g[i] * nodes[n.b].val
			g[n.b] += g[i] * nodes[n.a].val
		case "tanh":
			g[n.a] += g[i] * (1 - n.val*n.val)
		case "exp":
			g[n.a] += g[i] * n.val
		case "relu":
			if nodes[n.a].val > 0 {
				g[n.a] += g[i]
			}
		}
	}
	grads := make([]float64, len(inputs))
	for k, i := range inputs {
		grads[k] = g[i] + 0.0
	}
	return nodes[len(nodes)-1].val, names, grads
}

// autodiffWant is the expected output for a valid program.
func autodiffWant(prog string) string {
	v, names, grads := autodiffRun(prog)
	var b strings.Builder
	fmt.Fprintf(&b, "value %.6f\n", v)
	for i, n := range names {
		g := grads[i]
		if g == 0 {
			g = 0 // never -0
		}
		fmt.Fprintf(&b, "grad %s %.6f\n", n, g)
	}
	return b.String()
}

// autodiffManyCase is a 20,000-statement program over 8 inputs that keeps
// reusing recent values through tanh, so values and gradients stay in
// range. Seeds are tried until no gradient is close enough to zero for
// rounding to decide its printed sign.
func autodiffManyCase() (string, string) {
	for seed := int64(15); ; seed++ {
		r := rand.New(rand.NewSource(seed))
		var b strings.Builder
		names := []string{}
		for i := range 8 {
			n := fmt.Sprintf("x%d", i)
			fmt.Fprintf(&b, "input %s %.3f\n", n, r.Float64()*2-1)
			names = append(names, n)
		}
		ops := []string{"add", "sub", "mul", "tanh", "tanh", "add"}
		for i := 0; i < 20000-8; i++ {
			pick := func() string { return names[len(names)-1-r.Intn(min(len(names), 12))] }
			op := ops[r.Intn(len(ops))]
			n := fmt.Sprintf("v%d", i)
			if op == "tanh" {
				fmt.Fprintf(&b, "%s = tanh %s\n", n, pick())
			} else {
				fmt.Fprintf(&b, "%s = %s %s %s\n", n, op, pick(), pick())
			}
			names = append(names, n)
			if i%3 == 2 { // squash, so nothing grows without bound
				m := fmt.Sprintf("t%d", i)
				fmt.Fprintf(&b, "%s = tanh %s\n", m, n)
				names = append(names, m)
				i++
			}
		}
		prog := b.String()
		_, _, grads := autodiffRun(prog)
		ok := true
		for _, g := range grads {
			if g != 0 && math.Abs(g) < 1e-4 {
				ok = false
			}
		}
		if ok {
			return prog, autodiffWant(prog)
		}
	}
}

// attentionWant computes causal attention exactly as the task describes.
func attentionWant(in string) string {
	f := strings.Fields(in)
	n, _ := strconv.Atoi(f[0])
	d, _ := strconv.Atoi(f[1])
	num := func(i int) float64 { v, _ := strconv.ParseFloat(f[2+i], 64); return v }
	at := func(m, i, j int) float64 { return num(m*n*d + i*d + j) }
	var b strings.Builder
	s := make([]float64, n)
	for i := range n {
		m := math.Inf(-1)
		for j := 0; j <= i; j++ {
			dot := 0.0
			for k := range d {
				dot += at(0, i, k) * at(1, j, k)
			}
			s[j] = dot / math.Sqrt(float64(d))
			m = math.Max(m, s[j])
		}
		sum := 0.0
		for j := 0; j <= i; j++ {
			s[j] = math.Exp(s[j] - m)
			sum += s[j]
		}
		for k := range d {
			o := 0.0
			for j := 0; j <= i; j++ {
				o += s[j] / sum * at(2, j, k)
			}
			if k > 0 {
				b.WriteByte(' ')
			}
			fmt.Fprintf(&b, "%.4f", o)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// attentionCase is random Q and K (scaled by size, to make huge scores)
// and V in [0.1, 1.1], so no output is near zero.
func attentionCase(n, d int, size float64, seed int64) (string, string) {
	r := rand.New(rand.NewSource(seed))
	var b strings.Builder
	fmt.Fprintf(&b, "%d %d\n", n, d)
	for m := range 3 {
		for range n {
			for k := range d {
				if k > 0 {
					b.WriteByte(' ')
				}
				v := r.NormFloat64() * size
				if m == 2 {
					v = 0.1 + r.Float64()
				}
				fmt.Fprintf(&b, "%.3f", v)
			}
			b.WriteByte('\n')
		}
	}
	return b.String(), attentionWant(b.String())
}

// bpeWant learns merges exactly as the task describes.
func bpeWant(text string, merges int) string {
	toks := make([]int, len(text))
	for i := range len(text) {
		toks[i] = int(text[i])
	}
	var b strings.Builder
	size := 256 + merges
	count := make([]int, size*size)
	for m := range merges {
		clear(count)
		for i := 0; i+1 < len(toks); i++ {
			count[toks[i]*size+toks[i+1]]++
		}
		best, bestN := -1, 1
		for p, c := range count {
			if c > bestN {
				best, bestN = p, c
			}
		}
		if best < 0 {
			break
		}
		l, rt, id := best/size, best%size, 256+m
		fmt.Fprintf(&b, "%d = %d %d (%d)\n", id, l, rt, bestN)
		w := 0
		for i := 0; i < len(toks); {
			if i+1 < len(toks) && toks[i] == l && toks[i+1] == rt {
				toks[w] = id
				i += 2
			} else {
				toks[w] = toks[i]
				i++
			}
			w++
		}
		toks = toks[:w]
	}
	fmt.Fprintf(&b, "tokens %d\n", len(toks))
	return b.String()
}

// bpeEnglishCase is about 100 KB of English from the academy's own
// lessons: varied, real prose, and the same on every machine.
func bpeEnglishCase() (string, string) {
	var b strings.Builder
	for _, id := range []int{0, 1, 2, 3, 13, 14} {
		for _, c := range curriculum[id].Concepts {
			for _, t := range []string{c.Summary, c.Body, c.Analogy, c.Example} {
				for _, r := range t {
					if r < 128 { // plain ASCII, so the test text is the same bytes everywhere
						b.WriteRune(r)
					}
				}
				b.WriteByte('\n')
			}
			if b.Len() > 100_000 {
				text := b.String()[:100_000]
				return text, bpeWant(text, 300)
			}
		}
	}
	text := b.String()
	return text, bpeWant(text, 300)
}
