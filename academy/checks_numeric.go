package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Checked exercises for Stage 13 (mathematical foundations): numerical
// code whose answers depend on doing the arithmetic in a safe order.
var numericChecks = []Check{
	{
		ID: "softmax", Stage: 13, Concept: "Probability, Softmax & Cross-Entropy",
		Title: "Softmax and cross-entropy that never overflow",
		Task:  "Each line of standard input holds the index of the true class (counting from 0), then two or more logits, separated by spaces. For each line, print the softmax probabilities with %.6f separated by spaces, then loss= and the cross-entropy loss −log p[true] with %.6f. Logits can be as large as ±1000, where exp overflows (or underflows to 0), so subtract the largest logit before exponentiating, and compute the loss from log-sum-exp rather than by taking the log of a probability. A line with a missing or out-of-range class, fewer than two logits, or something that is not a number gets a message on standard error and makes the program exit 1 at the end.",
		File:  "softmax.c",
		Starter: `/* Checked exercise: softmax
 *   input:   0 1000 1001 1002
 *   output:  0.090031 0.244728 0.665241 loss=2.407606
 * softmax(z)[i] = exp(z[i]) / sum_j exp(z[j])
 * exp(710) is inf, and exp(-800) is 0, so compute with m = max(z):
 *   s = sum_j exp(z[j] - m)        the largest term is exp(0) = 1
 *   p[i] = exp(z[i] - m) / s
 *   loss = m + log(s) - z[true]     (log-sum-exp minus the true logit)
 * Read a whole line with fgets, then take numbers from it with strtod,
 * checking its end pointer to detect something that is not a number.
 * At most 64 logits per line, and lines are shorter than 4096 bytes.
 */
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
`,
		Libs: []string{"-lm"},
		Cases: []CheckCase{
			{Name: "small logits", Stdin: "0 1 2 3\n2 0 0 0 0\n", Want: softmaxWant("0 1 2 3\n2 0 0 0 0\n"),
				Explain: "Equal logits give equal probabilities: four zeros give 0.25 each, and the loss is log 4."},
			{Name: "large logits", Stdin: "0 1000 1001 1002\n", Want: softmaxWant("0 1000 1001 1002\n"),
				Explain: "exp(1000) is inf, and inf / inf is NaN. Softmax only depends on the differences between logits, so subtracting the largest (1002) changes nothing but makes every exponent at most 0."},
			{Name: "very negative logits", Stdin: "2 -1000 -1001 -1002\n", Want: softmaxWant("2 -1000 -1001 -1002\n"),
				Explain: "exp(-1000) underflows to 0, and 0 / 0 is NaN too. After subtracting the largest logit (-1000), the biggest term is exp(0) = 1."},
			{Name: "a hopeless prediction", Stdin: "0 0 800\n", Want: softmaxWant("0 0 800\n"),
				Explain: "p[0] = exp(-800) rounds to 0, so -log p[0] would be infinite. Log-sum-exp gives the true loss, 800: m + log(s) - z[true] = 800 + 0 - 0."},
			{Name: "the class is counted from 0", Stdin: "1 2 1\n", Want: softmaxWant("1 2 1\n"),
				Explain: "Class 1 is the second logit (1), whose probability is 0.268941, so the loss is 1.313262."},
			{Name: "a class out of range", Stdin: "3 1 2 3\n", WantExit: 1, Stderr: true,
				Explain: "With three logits the classes are 0, 1 and 2: report the line on standard error and exit 1."},
			{Name: "something that is not a number", Stdin: "0 1 two 3\n", WantExit: 1, Stderr: true},
			{Name: "many lines", Gen: generated(softmaxManyCase),
				Explain: "Two thousand lines with up to 64 logits each, some shifted by ±1000: the same answer, to six decimals, as the mathematically exact one."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Parse with a loop: char *end; double v = strtod(p, &end); if (end == p) the rest is not a number (or the line ended); then p = end.",
			"The class index is read as a number too: check that it is a whole number from 0 to n - 1 once you know n.",
			"Never compute exp(z[i]) alone: always exp(z[i] - m). And never log(p[true]): it is -inf as soon as p[true] rounds to 0.",
		},
	},
	{
		ID: "power_iter", Stage: 13, Concept: "Eigenvectors & the SVD",
		Title: "The largest eigenvalue by power iteration",
		Task:  "Read a symmetric matrix from standard input: its size n (1 to 500) on the first line, then n rows of n numbers. Find the eigenvalue of largest absolute value, and its eigenvector, by power iteration: start from v = (1, 2, …, n) normalised to length 1; each step, compute w = A·v, λ = v·w, then v = w / ‖w‖, negated when λ < 0 (so v keeps its direction even when λ is negative); stop when no component of v changed by more than 1e-12, or after 100,000 steps. Print λ = v·A·v for the final v with %.6f on one line, then the components of v with %.6f separated by spaces, after negating the whole vector if needed so that its first component larger than 1e-6 in size is positive. A matrix that is not symmetric (or is incomplete) gets a message on standard error and exit status 1.",
		File:  "power_iter.c",
		Starter: `/* Checked exercise: power_iter
 *   input:     2
 *              2 1
 *              1 2
 *   output:    3.000000
 *              0.707107 0.707107
 * Power iteration: multiplying by A again and again stretches every
 * vector towards the eigenvector with the largest |eigenvalue|.
 *   v = (1, 2, ..., n) / its length
 *   repeat:  w = A v;  lambda = v . w;  v_new = w / |w|
 *            if lambda < 0, v_new = -v_new
 *            stop when max_i |v_new[i] - v[i]| <= 1e-12 (or 100,000 steps)
 * A matrix of 500 x 500 doubles is 2 MB: allocate it with malloc.
 */
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
`,
		Libs: []string{"-lm"},
		Cases: []CheckCase{
			{Name: "a 2 × 2 matrix", Stdin: "2\n2 1\n1 2\n", Want: powerIterWant("2\n2 1\n1 2\n"),
				Explain: "The eigenvalues of [[2, 1], [1, 2]] are 3 and 1; the eigenvector of 3 is (1, 1) / √2."},
			{Name: "a start on the wrong eigenvector", Stdin: "2\n1 -2\n-2 1\n", Want: powerIterWant("2\n1 -2\n-2 1\n"),
				Explain: "(1, 1) is exactly the eigenvector of -1, so power iteration from all ones would never leave it; (1, 2) has a component along (1, -1), whose eigenvalue 3 then takes over."},
			{Name: "a negative eigenvalue", Stdin: "2\n-4 1\n1 2\n", Want: powerIterWant("2\n-4 1\n1 2\n"),
				Explain: "The largest in size is -1 - √10 = -4.162278. ‖A·v‖ / ‖v‖ would say +4.16: the sign comes from v·A·v. Without the negation, v flips sign every step, so it never meets the stopping rule and runs all 100,000 steps."},
			{Name: "a 3 × 3 matrix", Stdin: "3\n4 1 0\n1 3 1\n0 1 2\n", Want: powerIterWant("3\n4 1 0\n1 3 1\n0 1 2\n")},
			{Name: "a slowly converging matrix", Stdin: "3\n10 0.5 0.2\n0.5 9.5 0.3\n0.2 0.3 1\n", Want: powerIterWant("3\n10 0.5 0.2\n0.5 9.5 0.3\n0.2 0.3 1\n"),
				Explain: "The top two eigenvalues are close (10.32 and 9.19), so each step shrinks the error only by their ratio: stopping after a fixed few steps, or when λ alone stops changing, leaves v wrong in the sixth decimal."},
			{Name: "not symmetric", Stdin: "2\n1 2\n3 4\n", WantExit: 1, Stderr: true,
				Explain: "A[0][1] = 2 but A[1][0] = 3: this method (and the meaning of v·A·v) assumes a symmetric matrix."},
			{Name: "a 300 × 300 matrix", Gen: generated(powerIterBigCase),
				Explain: "Random noise plus one strong pattern with a negative eigenvalue: the 300 numbers of the vector must all match. Each step is 90,000 multiplications, so a v that flips sign every step and runs all 100,000 steps takes far too long."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Read with scanf(\"%lf\", ...) and check that it returned 1 each time; then compare A[i][j] with A[j][i] exactly (the tests use the same text for both).",
			"Keep two vectors, v and w, and swap their roles rather than allocating a new one each step.",
			"Fix the sign of the final vector by finding the first i with fabs(v[i]) > 1e-6 and negating everything if v[i] < 0.",
		},
	},
	{
		ID: "kahan_sum", Stage: 13, Concept: "Floating Point for Numerical Computing",
		Title: "Summation that does not drift",
		Task:  "Each line of standard input holds a number, optionally followed by a repeat count (0.1 10000000 means add 0.1 ten million times; no count means once). Convert each number to float once, with strtof, and add it, the given number of times, to three sums: a plain float sum, a float sum with Kahan's compensated summation (y = x - c; t = sum + y; c = (t - sum) - y; sum = t), and a double sum of the same float values. Print naive, kahan and double with their sums in %.4f, one per line. Counts go up to 100 million. If any line is not a number with an optional whole count of at least 1, print a message on standard error and exit 1 without printing the sums.",
		File:  "kahan_sum.c",
		Starter: `/* Checked exercise: kahan_sum
 *   input:   0.1 10000000
 *   output:  naive 1087937.0000
 *            kahan 1000000.0000
 *            double 1000000.0149
 * Kahan's algorithm keeps the low bits each addition loses in c:
 *   y = x - c        the next number, corrected by the last loss
 *   t = sum + y      big + small: the low bits of y are lost here
 *   c = (t - sum) - y    what was actually added, minus what should
 *                        have been: the loss (algebraically 0!)
 *   sum = t
 * Keep sum, c, y and t all float. Build without -ffast-math: it lets
 * the compiler simplify (t - sum) - y to 0.
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "ten million tenths", Gen: kahanCase("0.1 10000000\n"),
				Explain: "0.1f is 0.100000001490116. Once the plain sum is large, each 0.1 is rounded to the sum's coarser spacing and the total drifts 8.8% high; Kahan's correction recovers the lost bits."},
			{Name: "absorption", Gen: kahanCase("16777216\n1 1000\n"),
				Explain: "Above 2²⁴ = 16,777,216 a float's spacing is 2: 16777216 + 1 rounds back to 16777216 a thousand times over. Kahan keeps the 1 in c and adds 2 at a time."},
			{Name: "a cancellation Kahan cannot save", Gen: kahanCase("100000000\n1\n-100000000\n"),
				Explain: "The true sum is 1. The double sum gets it, but both float sums give 0: after the 1 is absorbed, Kahan holds -1 in c, and then y = -100000000 - (-1) is rounded back to -100000000 because the spacing there is 8. Neumaier's variant fixes this case; the test asks for classic Kahan."},
			{Name: "mixed values", Gen: kahanCase("0.3 1000000\n-0.1 2000000\n1e-3 5000000\n12345.678\n")},
			{Name: "a hundred million", Gen: kahanCase("1 100000000\n"),
				Explain: "The plain float sum stops at 16,777,216, where adding 1 no longer changes it; the Kahan and double sums reach 100,000,000. That is a hundred million steps: keep the loop tight."},
			{Name: "a bad count", Stdin: "0.5 lots\n", WantExit: 1, Stderr: true},
			{Name: "a zero count", Stdin: "0.5 0\n", WantExit: 1, Stderr: true},
			{Name: "many lines", Gen: generated(kahanManyCase),
				Explain: "Thousands of lines of different sizes and signs: every sum must match to the last digit, which only an exactly specified algorithm can."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"strtof(p, &end) reads the number; then strtol(end, &end2, 10) the count, if anything but spaces and the newline follows.",
			"c holds what the last addition lost, with the sign that makes the next correction a subtraction: y = x - c, not x + c.",
			"Print a float with %.4f directly: printf promotes it to double, exactly.",
		},
	},
}

// softmaxLine is the expected output for one valid line of the softmax
// check: probabilities, then the cross-entropy loss via log-sum-exp.
func softmaxLine(class int, z []float64) string {
	m := z[0]
	for _, v := range z {
		m = math.Max(m, v)
	}
	s := 0.0
	for _, v := range z {
		s += math.Exp(v - m)
	}
	var b strings.Builder
	for _, v := range z {
		fmt.Fprintf(&b, "%.6f ", math.Exp(v-m)/s)
	}
	fmt.Fprintf(&b, "loss=%.6f\n", m+math.Log(s)-z[class])
	return b.String()
}

// softmaxWant computes the expected output for valid input lines.
func softmaxWant(in string) string {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(in, "\n"), "\n") {
		f := strings.Fields(line)
		class, _ := strconv.Atoi(f[0])
		z := make([]float64, len(f)-1)
		for i, s := range f[1:] {
			z[i], _ = strconv.ParseFloat(s, 64)
		}
		b.WriteString(softmaxLine(class, z))
	}
	return b.String()
}

// softmaxManyCase builds 2,000 lines of logits within 20 of each other
// (so no probability is exactly 1 and no loss rounds to -0), sometimes
// shifted by ±1000 so that only a max-subtracting softmax survives.
func softmaxManyCase() (string, string) {
	r := rand.New(rand.NewSource(13))
	var in strings.Builder
	for range 2000 {
		n := 2 + r.Intn(63)
		shift := []float64{0, 0, 1000, -1000}[r.Intn(4)]
		fmt.Fprintf(&in, "%d", r.Intn(n))
		for range n {
			fmt.Fprintf(&in, " %.3f", shift+r.Float64()*20-10)
		}
		in.WriteByte('\n')
	}
	return in.String(), softmaxWant(in.String())
}

// powerIter runs the power iteration exactly as the task describes and
// returns λ and the sign-normalised eigenvector.
func powerIter(a [][]float64) (float64, []float64) {
	n := len(a)
	mul := func(v []float64) []float64 {
		w := make([]float64, n)
		for i := range n {
			for j := range n {
				w[i] += a[i][j] * v[j]
			}
		}
		return w
	}
	dot := func(x, y []float64) float64 {
		s := 0.0
		for i := range x {
			s += x[i] * y[i]
		}
		return s
	}
	v := make([]float64, n)
	for i := range v {
		v[i] = float64(i + 1)
	}
	norm := math.Sqrt(dot(v, v))
	for i := range v {
		v[i] /= norm
	}
	for range 100000 {
		w := mul(v)
		lambda := dot(v, w)
		norm := math.Sqrt(dot(w, w))
		if lambda < 0 {
			norm = -norm
		}
		change := 0.0
		for i := range w {
			w[i] /= norm
			change = math.Max(change, math.Abs(w[i]-v[i]))
		}
		v = w
		if change <= 1e-12 {
			break
		}
	}
	lambda := dot(v, mul(v))
	for _, x := range v {
		if math.Abs(x) > 1e-6 {
			if x < 0 {
				for i := range v {
					v[i] = -v[i]
				}
			}
			break
		}
	}
	return lambda, v
}

// powerIterWant computes the expected output for a valid matrix.
func powerIterWant(in string) string {
	f := strings.Fields(in)
	n, _ := strconv.Atoi(f[0])
	a := make([][]float64, n)
	for i := range a {
		a[i] = make([]float64, n)
		for j := range a[i] {
			a[i][j], _ = strconv.ParseFloat(f[1+i*n+j], 64)
		}
	}
	lambda, v := powerIter(a)
	var b strings.Builder
	fmt.Fprintf(&b, "%.6f\n", lambda)
	for i, x := range v {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%.6f", x)
	}
	b.WriteByte('\n')
	return b.String()
}

// powerIterBigCase is a 300 × 300 symmetric matrix: noise in [-1, 1]
// minus 40·uuᵀ for a fixed unit vector u, so one negative eigenvalue
// (near -42) clearly dominates the noise's (within about ±20).
func powerIterBigCase() (string, string) {
	const n = 300
	r := rand.New(rand.NewSource(13))
	u := make([]float64, n)
	norm := 0.0
	for i := range u {
		u[i] = 1 + r.Float64()
		norm += u[i] * u[i]
	}
	norm = math.Sqrt(norm)
	a := make([][]float64, n)
	for i := range a {
		a[i] = make([]float64, n)
	}
	for i := range n {
		for j := i; j < n; j++ {
			x := r.Float64()*2 - 1 - 40*u[i]*u[j]/(norm*norm)
			// Round to the printed precision, so the program reads
			// exactly the matrix the expected answer was computed for.
			x, _ = strconv.ParseFloat(strconv.FormatFloat(x, 'f', 6, 64), 64)
			a[i][j], a[j][i] = x, x
		}
	}
	var in strings.Builder
	fmt.Fprintf(&in, "%d\n", n)
	for i := range a {
		for j, x := range a[i] {
			if j > 0 {
				in.WriteByte(' ')
			}
			in.WriteString(strconv.FormatFloat(x, 'f', 6, 64))
		}
		in.WriteByte('\n')
	}
	return in.String(), powerIterWant(in.String())
}

// kahanWant computes the three sums for valid input lines, in float32
// exactly as the task specifies.
func kahanWant(in string) string {
	var naive, sum, c float32
	var dbl float64
	for _, line := range strings.Split(strings.TrimSuffix(in, "\n"), "\n") {
		f := strings.Fields(line)
		x64, _ := strconv.ParseFloat(f[0], 32)
		x := float32(x64)
		count := 1
		if len(f) > 1 {
			count, _ = strconv.Atoi(f[1])
		}
		for range count {
			naive += x
			y := x - c
			t := sum + y
			c = (t - sum) - y
			sum = t
			dbl += float64(x)
		}
	}
	return fmt.Sprintf("naive %.4f\nkahan %.4f\ndouble %.4f\n", float64(naive), float64(sum), dbl)
}

// kahanCase defers the sums, which can take a hundred million steps,
// until the case runs.
func kahanCase(in string) func() (string, string) {
	return generated(func() (string, string) { return in, kahanWant(in) })
}

// kahanManyCase builds 3,000 lines of values of very different sizes
// and signs, some repeated, so that the three sums differ.
func kahanManyCase() (string, string) {
	r := rand.New(rand.NewSource(13))
	var in strings.Builder
	for range 3000 {
		v := (r.Float64()*2 - 0.8) * math.Pow(10, float64(r.Intn(9)-4))
		in.WriteString(strconv.FormatFloat(v, 'g', 7, 64))
		if r.Intn(3) == 0 {
			fmt.Fprintf(&in, " %d", 1+r.Intn(5000))
		}
		in.WriteByte('\n')
	}
	return in.String(), kahanWant(in.String())
}
