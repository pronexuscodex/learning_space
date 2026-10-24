package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Checked exercises for Stage 14 (statistics and classical machine
// learning): the models and measurements behind the libraries.
var dataChecks = []Check{
	{
		ID: "ols", Stage: 14, Concept: "Linear & Logistic Regression",
		Title: "A regression line that survives big numbers",
		Task:  "Read (x, y) pairs from standard input, one per line, and fit the least-squares line y = slope·x + intercept. Print slope, intercept and r2 (the coefficient of determination), each with %.4f. Compute it from centred sums, so that x values such as Unix timestamps (around 1.7 × 10⁹) do not lose every digit: x̄ and ȳ first, then Sxx = Σ(x − x̄)², Sxy = Σ(x − x̄)(y − ȳ) and Syy = Σ(y − ȳ)²; slope = Sxy / Sxx, intercept = ȳ − slope·x̄, r2 = Sxy² / (Sxx·Syy). Fewer than two points, x values that are all the same (a vertical line has no slope), y values that are all the same, or a line that is not two numbers get a message on standard error and exit status 1, with nothing printed.",
		File:  "ols.c",
		Starter: `/* Checked exercise: ols (ordinary least squares)
 *   input:   1 3        output:  slope 2.0000
 *            2 5                 intercept 1.0000
 *            3 7                 r2 1.0000
 * The textbook shortcut  slope = (n·Σxy − Σx·Σy) / (n·Σx² − (Σx)²)
 * subtracts two enormous, nearly equal numbers when x is large (a
 * timestamp squared is about 3 × 10^18), and the answer is noise.
 * Centre first: subtract the means, then sum the products. That
 * means reading the data twice, so keep it in a growing array
 * (malloc, then realloc to double the size when it is full).
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "a perfect line", Stdin: "1 3\n2 5\n3 7\n4 9\n5 11\n", Want: olsWant("1 3\n2 5\n3 7\n4 9\n5 11\n"),
				Explain: "Every point is on y = 2x + 1, so the slope is 2, the intercept 1 and r2 exactly 1."},
			{Name: "noisy points", Stdin: "1 2.1\n2 3.9\n3 6.2\n4 7.8\n5 10.1\n6 12.2\n", Want: olsWant("1 2.1\n2 3.9\n3 6.2\n4 7.8\n5 10.1\n6 12.2\n")},
			{Name: "a falling line", Stdin: "0 10\n1 8.5\n2 7.2\n3 5.4\n4 4.1\n", Want: olsWant("0 10\n1 8.5\n2 7.2\n3 5.4\n4 4.1\n"),
				Explain: "A negative slope gives a negative Sxy, but r2 = Sxy² / (Sxx·Syy) is still between 0 and 1."},
			{Name: "Unix timestamps", Stdin: olsTimestamps, Want: olsWant(olsTimestamps),
				Explain: "x is around 1.7 × 10⁹, so Σx² is about 2.9 × 10¹⁹: beyond the 2⁵³ that a double holds exactly. The shortcut formula's two huge terms cancel into rounding error; centred sums keep every digit, and the slope comes out exactly 0.5."},
			{Name: "every x the same", Stdin: "3 1\n3 2\n3 5\n", WantExit: 1, Stderr: true,
				Explain: "Sxx = 0: the points stand on a vertical line, which has no slope. Report it instead of dividing by zero."},
			{Name: "a single point", Stdin: "1 2\n", WantExit: 1, Stderr: true},
			{Name: "something that is not a number", Stdin: "1 2\n2 three\n", WantExit: 1, Stderr: true},
			{Name: "a hundred thousand points", Gen: generated(olsManyCase),
				Explain: "100,000 noisy points around y = 3x − 7: the array must grow as it is read."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Read pairs with scanf(\"%lf %lf\", &x, &y) == 2 in a loop, and check what stopped it: EOF is fine, a failed match is an error.",
			"Two passes over the array: one for the means, one for Sxx, Sxy and Syy.",
			"Check Sxx == 0 and Syy == 0 before dividing; both are exact zeros when all values are equal.",
		},
	},
	{
		ID: "kmeans", Stage: 14, Concept: "Trees, Ensembles & Clustering",
		Title: "k-means clustering",
		Task:  "Read k on the first line, then 2D points, one \"x y\" per line. Run k-means: the starting centres are the first k points; then repeat (assign every point to its nearest centre by squared distance, a tie going to the lower-numbered centre; move every centre to the mean of its points, a centre with no points staying where it is) until no point changes centre, or 1,000 rounds. Print each centre as centre I: X Y n=N, with X and Y in %.4f and centres counted from 0, then inertia (the sum of squared distances from each point to its centre) with %.4f. k below 1, fewer points than k, or a line that is not two numbers get a message on standard error and exit status 1, with nothing printed.",
		File:  "kmeans.c",
		Starter: `/* Checked exercise: kmeans
 *   input:  2            output:  centre 0: 0.5000 0.0000 n=2
 *           0 0                   centre 1: 2.0000 0.0000 n=1
 *           2 0                   inertia 0.5000
 *           1 0
 * Start from the first k points as the centres. Each round:
 *   1. assign each point to the centre with the smallest squared
 *      distance (dx*dx + dy*dy); on a tie, the lower-numbered centre
 *   2. move each centre to the mean of its points (an empty centre
 *      stays put: dividing by its 0 points would give NaN)
 * Stop when no assignment changed (or after 1,000 rounds), then print.
 * The first round always "changes" every assignment from nothing.
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "two clear groups", Stdin: kmeansSmall, Want: kmeansWant(kmeansSmall),
				Explain: "The first two points are both in the lower-left group, so one centre has to travel: after a few rounds each group has its own centre."},
			{Name: "ties go to the first centre", Stdin: "2\n0 0\n2 0\n1 0\n", Want: kmeansWant("2\n0 0\n2 0\n1 0\n"),
				Explain: "(1, 0) is exactly as far from (0, 0) as from (2, 0): the task gives it to centre 0, which then moves to (0.5, 0)."},
			{Name: "an empty cluster", Stdin: "3\n0 0\n0 0\n10 10\n10 11\n1 1\n", Want: kmeansWant("3\n0 0\n0 0\n10 10\n10 11\n1 1\n"),
				Explain: "Centres 0 and 1 start at the same point and every tie goes to centre 0, so in the first round centre 1 gets no points. It must stay at (0, 0), not become NaN from 0 / 0; in the next round, with centre 0 moved away, the two points at (0, 0) are its own."},
			{Name: "k is 0", Stdin: "0\n1 1\n", WantExit: 1, Stderr: true},
			{Name: "more clusters than points", Stdin: "3\n1 1\n2 2\n", WantExit: 1, Stderr: true},
			{Name: "twenty thousand points", Gen: generated(kmeansManyCase),
				Explain: "Eight blobs, k = 8, started from the first 8 points: it takes several rounds to settle, so stopping after a fixed few gives different centres. Look at the answer, too: three centres share one blob while other blobs share a centre. That is a local optimum, and the reason real libraries start from k-means++ and run several times."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Keep the points in growing arrays, an int array for the assignments (start it at -1), and arrays of k sums and counts.",
			"Compare squared distances with < (strictly less): the first centre that reaches the minimum keeps the point, which is the tie rule.",
			"The inertia is computed once, at the end, from the final assignments and centres.",
		},
	},
	{
		ID: "metrics", Stage: 14, Concept: "Evaluating Models Honestly",
		Title: "Precision, recall and AUC",
		Task:  "./metrics THRESHOLD reads predictions from standard input, one \"label score\" per line (label 0 or 1, score a number), and evaluates them. A prediction is positive when its score is at least THRESHOLD. Print the confusion matrix as TP n FP n FN n TN n, then accuracy, precision, recall, F1 and AUC, one per line, each with %.4f, or n/a where it is undefined: precision when nothing is predicted positive, recall when there are no positives, F1 when either of those is n/a (and 0.0000 when both are 0), AUC when there are no positives or no negatives. AUC is the probability that a random positive scores higher than a random negative, a tie counting as half. Inputs can have 200,000 lines: compare every positive with every negative and it is far too slow, so sort by score. A label other than 0 or 1, or a line that is not a label and a number, gets a message on standard error and exit status 1, with nothing printed.",
		File:  "metrics.c",
		Starter: `/* Checked exercise: metrics
 *   ./metrics 0.5  <  predictions
 *   1 0.9                         TP 2 FP 1 FN 1 TN 2
 *   0 0.8                         accuracy 0.6667
 *   1 0.7                         precision 0.6667
 *   0 0.3                         recall 0.6667
 *   1 0.2                         F1 0.6667
 *   0 0.1                         AUC 0.6667
 * precision = TP / (TP + FP)      recall = TP / (TP + FN)
 * F1 = 2·precision·recall / (precision + recall)
 * AUC in O(n log n): sort by score (qsort); walk from the lowest
 * score up, a group of equal scores at a time. Each positive in a
 * group beats every negative below the group, and ties half of the
 * negatives in it. AUC = that total / (positives × negatives).
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "a small example", Args: []string{"0.5"}, Stdin: metricsSmall, Want: metricsWant(metricsSmall, 0.5),
				Explain: "Of the 3 predicted positive, 2 are; of the 3 real positives, 2 were caught. AUC: of the 9 positive–negative pairs, the positive scores higher in 6."},
			{Name: "the threshold is inclusive", Args: []string{"0.7"}, Stdin: metricsSmall, Want: metricsWant(metricsSmall, 0.7),
				Explain: "A score equal to the threshold (0.7) counts as positive: \"at least THRESHOLD\"."},
			{Name: "tied scores count half", Args: []string{"0.5"}, Stdin: "1 0.5\n0 0.5\n1 0.9\n0 0.1\n", Want: metricsWant("1 0.5\n0 0.5\n1 0.9\n0 0.1\n", 0.5),
				Explain: "The pair (1 0.5, 0 0.5) is a tie and counts as half a win: AUC = (3 + 0.5) / 4 = 0.875."},
			{Name: "nothing flagged", Args: []string{"0.95"}, Stdin: metricsSmall, Want: metricsWant(metricsSmall, 0.95),
				Explain: "No score reaches 0.95, so precision is 0 / 0: n/a, not a crash or nan. Recall is a real 0, and F1 is n/a because precision is."},
			{Name: "only one class", Args: []string{"0.5"}, Stdin: "0 0.2\n0 0.7\n0 0.4\n", Want: metricsWant("0 0.2\n0 0.7\n0 0.4\n", 0.5),
				Explain: "With no positives at all, recall and AUC have no meaning: n/a."},
			{Name: "a bad label", Args: []string{"0.5"}, Stdin: "1 0.9\n2 0.4\n", WantExit: 1, Stderr: true},
			{Name: "an imbalanced set", Args: []string{"0.5"}, Gen: generated(func() (string, string) { return metricsCase(20000, 0.01, 14, 0.5) }),
				Explain: "1% positives: accuracy is high whatever the model does, so precision and recall are the numbers to read."},
			{Name: "two hundred thousand predictions", Args: []string{"0.3"}, Gen: generated(func() (string, string) { return metricsCase(200000, 0.5, 15, 0.3) }),
				Explain: "About 100,000 positives and 100,000 negatives: comparing every pair is 10¹⁰ comparisons, minutes of work. Sorting once is well under a second. Scores repeat (three decimals), so the tie rule matters here too."},
		},
		Timeout: 5 * time.Second,
		Hints: []string{
			"Store each line as a struct { double score; int label; } in a growing array, and qsort it by score.",
			"Count TP, FP, FN and TN in the same pass that reads the input; only AUC needs the sort.",
			"Keep the AUC total doubled (2 per win, 1 per tie) in a long long, so it stays exact, and divide at the end.",
		},
	},
}

// olsTimestamps are points on y = 3 + 0.5·(x − 1,700,000,000) at Unix
// timestamps, where the uncentred formula cancels catastrophically.
var olsTimestamps = func() string {
	var b strings.Builder
	for i := range 10 {
		fmt.Fprintf(&b, "%d %g\n", 1700000000+i, 3+0.5*float64(i))
	}
	return b.String()
}()

// olsFit is the centred least-squares fit the task describes.
func olsFit(xs, ys []float64) (slope, intercept, r2 float64) {
	n := float64(len(xs))
	mx, my := 0.0, 0.0
	for i := range xs {
		mx += xs[i]
		my += ys[i]
	}
	mx, my = mx/n, my/n
	var sxx, sxy, syy float64
	for i := range xs {
		dx, dy := xs[i]-mx, ys[i]-my
		sxx += dx * dx
		sxy += dx * dy
		syy += dy * dy
	}
	slope = sxy / sxx
	return slope, my - slope*mx, sxy * sxy / (sxx * syy)
}

// olsWant is the expected output for valid input.
func olsWant(in string) string {
	var xs, ys []float64
	for _, line := range strings.Split(strings.TrimSpace(in), "\n") {
		f := strings.Fields(line)
		x, _ := strconv.ParseFloat(f[0], 64)
		y, _ := strconv.ParseFloat(f[1], 64)
		xs, ys = append(xs, x), append(ys, y)
	}
	s, i, r := olsFit(xs, ys)
	return fmt.Sprintf("slope %.4f\nintercept %.4f\nr2 %.4f\n", s, i, r)
}

// olsManyCase is 100,000 noisy points around y = 3x − 7.
func olsManyCase() (string, string) {
	r := rand.New(rand.NewSource(14))
	var in strings.Builder
	for range 100000 {
		x := r.Float64() * 100
		fmt.Fprintf(&in, "%.3f %.3f\n", x, 3*x-7+r.NormFloat64()*5)
	}
	return in.String(), olsWant(in.String())
}

var kmeansSmall = "2\n1 1\n1.5 2\n9 9\n8.5 9.5\n1 0.5\n9.5 8\n2 1.5\n8 8.5\n"

// kmeansRun is k-means exactly as the task describes.
func kmeansRun(k int, xs, ys []float64) (cx, cy []float64, count []int, inertia float64) {
	n := len(xs)
	cx, cy = append([]float64(nil), xs[:k]...), append([]float64(nil), ys[:k]...)
	assign := make([]int, n)
	for i := range assign {
		assign[i] = -1
	}
	count = make([]int, k)
	for round := 0; round < 1000; round++ {
		changed := false
		for i := range n {
			best, bestD := 0, math.Inf(1)
			for c := range k {
				dx, dy := xs[i]-cx[c], ys[i]-cy[c]
				if d := dx*dx + dy*dy; d < bestD {
					best, bestD = c, d
				}
			}
			if assign[i] != best {
				assign[i], changed = best, true
			}
		}
		if !changed {
			break
		}
		sx, sy := make([]float64, k), make([]float64, k)
		count = make([]int, k)
		for i := range n {
			sx[assign[i]] += xs[i]
			sy[assign[i]] += ys[i]
			count[assign[i]]++
		}
		for c := range k {
			if count[c] > 0 {
				cx[c], cy[c] = sx[c]/float64(count[c]), sy[c]/float64(count[c])
			}
		}
	}
	for i := range n {
		dx, dy := xs[i]-cx[assign[i]], ys[i]-cy[assign[i]]
		inertia += dx*dx + dy*dy
	}
	return cx, cy, count, inertia
}

// kmeansWant is the expected output for valid input.
func kmeansWant(in string) string {
	lines := strings.Split(strings.TrimSpace(in), "\n")
	k, _ := strconv.Atoi(lines[0])
	var xs, ys []float64
	for _, l := range lines[1:] {
		f := strings.Fields(l)
		x, _ := strconv.ParseFloat(f[0], 64)
		y, _ := strconv.ParseFloat(f[1], 64)
		xs, ys = append(xs, x), append(ys, y)
	}
	cx, cy, count, inertia := kmeansRun(k, xs, ys)
	var b strings.Builder
	for c := range k {
		fmt.Fprintf(&b, "centre %d: %.4f %.4f n=%d\n", c, cx[c], cy[c], count[c])
	}
	fmt.Fprintf(&b, "inertia %.4f\n", inertia)
	return b.String()
}

// kmeansManyCase is 20,000 points in eight blobs, shuffled, with k = 8.
func kmeansManyCase() (string, string) {
	r := rand.New(rand.NewSource(14))
	var in strings.Builder
	in.WriteString("8\n")
	for range 20000 {
		blob := r.Intn(8)
		cx, cy := float64(blob%4)*10, float64(blob/4)*12
		fmt.Fprintf(&in, "%.3f %.3f\n", cx+r.NormFloat64()*1.5, cy+r.NormFloat64()*1.5)
	}
	return in.String(), kmeansWant(in.String())
}

var metricsSmall = "1 0.9\n0 0.8\n1 0.7\n0 0.3\n1 0.2\n0 0.1\n"

// metricsWant is the expected output for valid predictions at threshold t.
func metricsWant(in string, t float64) string {
	type pred struct {
		score float64
		pos   bool
	}
	var ps []pred
	var tp, fp, fn, tn int
	for _, line := range strings.Split(strings.TrimSpace(in), "\n") {
		f := strings.Fields(line)
		s, _ := strconv.ParseFloat(f[1], 64)
		pos := f[0] == "1"
		ps = append(ps, pred{s, pos})
		switch flagged := s >= t; {
		case pos && flagged:
			tp++
		case pos:
			fn++
		case flagged:
			fp++
		default:
			tn++
		}
	}
	na := func(ok bool, v float64) string {
		if !ok {
			return "n/a"
		}
		return fmt.Sprintf("%.4f", v)
	}
	precision := float64(tp) / float64(tp+fp)
	recall := float64(tp) / float64(tp+fn)
	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}
	// AUC: doubled wins (2 a win, 1 a tie), exact in integers.
	sort.Slice(ps, func(i, j int) bool { return ps[i].score < ps[j].score })
	var twice, negBelow int64
	for i := 0; i < len(ps); {
		j := i
		var pos, neg int64
		for ; j < len(ps) && ps[j].score == ps[i].score; j++ {
			if ps[j].pos {
				pos++
			} else {
				neg++
			}
		}
		twice += pos * (2*negBelow + neg)
		negBelow += neg
		i = j
	}
	positives, negatives := int64(tp+fn), int64(fp+tn)
	auc := float64(twice) / float64(2*positives*negatives)
	var b strings.Builder
	fmt.Fprintf(&b, "TP %d FP %d FN %d TN %d\n", tp, fp, fn, tn)
	fmt.Fprintf(&b, "accuracy %.4f\n", float64(tp+tn)/float64(len(ps)))
	fmt.Fprintf(&b, "precision %s\n", na(tp+fp > 0, precision))
	fmt.Fprintf(&b, "recall %s\n", na(tp+fn > 0, recall))
	fmt.Fprintf(&b, "F1 %s\n", na(tp+fp > 0 && tp+fn > 0, f1))
	fmt.Fprintf(&b, "AUC %s\n", na(positives > 0 && negatives > 0, auc))
	return b.String()
}

// metricsCase is n predictions with the given share of positives, from a
// model that is good but not perfect, with scores to three decimals.
func metricsCase(n int, share float64, seed int64, t float64) (string, string) {
	r := rand.New(rand.NewSource(seed))
	var in strings.Builder
	for range n {
		pos := r.Float64() < share
		mean := 0.35
		if pos {
			mean = 0.6
		}
		s := min(1, max(0, mean+r.NormFloat64()*0.18))
		label := 0
		if pos {
			label = 1
		}
		fmt.Fprintf(&in, "%d %.3f\n", label, s)
	}
	return in.String(), metricsWant(in.String(), t)
}
