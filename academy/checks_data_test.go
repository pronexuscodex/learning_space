package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range dataSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var dataSolutions = map[string]string{
	"ols": `#include <stdio.h>
#include <stdlib.h>

int main(void) {
    size_t n = 0, cap = 1024;
    double *x = malloc(cap * sizeof *x), *y = malloc(cap * sizeof *y);
    if (!x || !y) return 1;
    double a, b;
    int got;
    while ((got = scanf("%lf %lf", &a, &b)) == 2) {
        if (n == cap) {
            cap *= 2;
            double *nx = realloc(x, cap * sizeof *x), *ny = realloc(y, cap * sizeof *y);
            if (!nx || !ny) return 1;
            x = nx;
            y = ny;
        }
        x[n] = a;
        y[n] = b;
        n++;
    }
    if (got != EOF) {
        fprintf(stderr, "ols: every line must be two numbers\n");
        return 1;
    }
    if (n < 2) {
        fprintf(stderr, "ols: need at least two points\n");
        return 1;
    }
    double mx = 0, my = 0;
    for (size_t i = 0; i < n; i++) {
        mx += x[i];
        my += y[i];
    }
    mx /= n;
    my /= n;
    double sxx = 0, sxy = 0, syy = 0;
    for (size_t i = 0; i < n; i++) {
        double dx = x[i] - mx, dy = y[i] - my;
        sxx += dx * dx;
        sxy += dx * dy;
        syy += dy * dy;
    }
    if (sxx == 0 || syy == 0) {
        fprintf(stderr, "ols: all x (or all y) values are the same\n");
        return 1;
    }
    double slope = sxy / sxx;
    printf("slope %.4f\nintercept %.4f\nr2 %.4f\n", slope, my - slope * mx, sxy * sxy / (sxx * syy));
    free(x);
    free(y);
    return 0;
}
`,
	"kmeans": `#include <stdio.h>
#include <stdlib.h>

int main(void) {
    int k;
    if (scanf("%d", &k) != 1 || k < 1) {
        fprintf(stderr, "kmeans: the first line must be k, at least 1\n");
        return 1;
    }
    size_t n = 0, cap = 1024;
    double *x = malloc(cap * sizeof *x), *y = malloc(cap * sizeof *y);
    if (!x || !y) return 1;
    double a, b;
    int got;
    while ((got = scanf("%lf %lf", &a, &b)) == 2) {
        if (n == cap) {
            cap *= 2;
            double *nx = realloc(x, cap * sizeof *x), *ny = realloc(y, cap * sizeof *y);
            if (!nx || !ny) return 1;
            x = nx;
            y = ny;
        }
        x[n] = a;
        y[n] = b;
        n++;
    }
    if (got != EOF) {
        fprintf(stderr, "kmeans: every point must be two numbers\n");
        return 1;
    }
    if (n < (size_t)k) {
        fprintf(stderr, "kmeans: fewer points than clusters\n");
        return 1;
    }
    double *cx = malloc(k * sizeof *cx), *cy = malloc(k * sizeof *cy);
    double *sx = malloc(k * sizeof *sx), *sy = malloc(k * sizeof *sy);
    int *count = calloc(k, sizeof *count), *assign = malloc(n * sizeof *assign);
    if (!cx || !cy || !sx || !sy || !count || !assign) return 1;
    for (int c = 0; c < k; c++) {
        cx[c] = x[c];
        cy[c] = y[c];
    }
    for (size_t i = 0; i < n; i++) assign[i] = -1;
    for (int round = 0; round < 1000; round++) {
        int changed = 0;
        for (size_t i = 0; i < n; i++) {
            int best = 0;
            double bestd = 0;
            for (int c = 0; c < k; c++) {
                double dx = x[i] - cx[c], dy = y[i] - cy[c], d = dx * dx + dy * dy;
                if (c == 0 || d < bestd) {
                    best = c;
                    bestd = d;
                }
            }
            if (assign[i] != best) {
                assign[i] = best;
                changed = 1;
            }
        }
        if (!changed) break;
        for (int c = 0; c < k; c++) {
            sx[c] = sy[c] = 0;
            count[c] = 0;
        }
        for (size_t i = 0; i < n; i++) {
            sx[assign[i]] += x[i];
            sy[assign[i]] += y[i];
            count[assign[i]]++;
        }
        for (int c = 0; c < k; c++)
            if (count[c] > 0) {
                cx[c] = sx[c] / count[c];
                cy[c] = sy[c] / count[c];
            }
    }
    double inertia = 0;
    for (size_t i = 0; i < n; i++) {
        double dx = x[i] - cx[assign[i]], dy = y[i] - cy[assign[i]];
        inertia += dx * dx + dy * dy;
    }
    for (int c = 0; c < k; c++) printf("centre %d: %.4f %.4f n=%d\n", c, cx[c], cy[c], count[c]);
    printf("inertia %.4f\n", inertia);
    free(x); free(y); free(cx); free(cy); free(sx); free(sy); free(count); free(assign);
    return 0;
}
`,
	"metrics": `#include <stdio.h>
#include <stdlib.h>

struct pred { double score; int label; };

static int byscore(const void *a, const void *b) {
    double x = ((const struct pred *)a)->score, y = ((const struct pred *)b)->score;
    return (x > y) - (x < y);
}

static void show(const char *name, int ok, double v) {
    if (ok) printf("%s %.4f\n", name, v);
    else printf("%s n/a\n", name);
}

int main(int argc, char **argv) {
    char *end;
    double t = argc == 2 ? strtod(argv[1], &end) : 0;
    if (argc != 2 || end == argv[1] || *end) {
        fprintf(stderr, "usage: metrics THRESHOLD\n");
        return 2;
    }
    size_t n = 0, cap = 1024;
    struct pred *p = malloc(cap * sizeof *p);
    if (!p) return 1;
    long tp = 0, fp = 0, fn = 0, tn = 0;
    int label, got;
    double s;
    while ((got = scanf("%d %lf", &label, &s)) == 2) {
        if (label != 0 && label != 1) {
            fprintf(stderr, "metrics: a label must be 0 or 1\n");
            return 1;
        }
        if (n == cap) {
            cap *= 2;
            struct pred *np = realloc(p, cap * sizeof *p);
            if (!np) return 1;
            p = np;
        }
        p[n].score = s;
        p[n].label = label;
        n++;
        int flagged = s >= t;
        if (label && flagged) tp++;
        else if (label) fn++;
        else if (flagged) fp++;
        else tn++;
    }
    if (got != EOF) {
        fprintf(stderr, "metrics: every line must be a label and a score\n");
        return 1;
    }
    qsort(p, n, sizeof *p, byscore);
    long long twice = 0, negbelow = 0;
    for (size_t i = 0; i < n;) {
        size_t j = i;
        long long pos = 0, neg = 0;
        for (; j < n && p[j].score == p[i].score; j++) {
            if (p[j].label) pos++;
            else neg++;
        }
        twice += pos * (2 * negbelow + neg);
        negbelow += neg;
        i = j;
    }
    long positives = tp + fn, negatives = fp + tn;
    double precision = tp + fp ? (double)tp / (tp + fp) : 0;
    double recall = positives ? (double)tp / positives : 0;
    double f1 = precision + recall > 0 ? 2 * precision * recall / (precision + recall) : 0;
    printf("TP %ld FP %ld FN %ld TN %ld\n", tp, fp, fn, tn);
    printf("accuracy %.4f\n", n ? (double)(tp + tn) / n : 0);
    show("precision", tp + fp > 0, precision);
    show("recall", positives > 0, recall);
    show("F1", tp + fp > 0 && positives > 0, f1);
    show("AUC", positives > 0 && negatives > 0, positives && negatives ? (double)twice / (2.0 * positives * negatives) : 0);
    free(p);
    return 0;
}
`,
}

func TestDataChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := dataSolutions
	mutate := func(id, from, to string) string {
		t.Helper()
		out := strings.Replace(ref[id], from, to, 1)
		if out == ref[id] {
			t.Fatalf("%s: the mutation %q did not apply", id, from)
		}
		return out
	}
	old := checkCaseTimeout
	checkCaseTimeout = time.Second
	defer func() { checkCaseTimeout = old }()

	// The textbook shortcut from raw sums: catastrophic cancellation for timestamps.
	shortcut := mutate("ols", "    double slope = sxy / sxx;\n", `    double Sx = 0, Sy = 0, Sxx = 0, Sxy = 0;
    for (size_t i = 0; i < n; i++) {
        Sx += x[i];
        Sy += y[i];
        Sxx += x[i] * x[i];
        Sxy += x[i] * y[i];
    }
    double slope = (n * Sxy - Sx * Sy) / (n * Sxx - Sx * Sx);
`)
	failsWith(t, "ols", shortcut, "Unix timestamps", "")
	// r2 from Sxy / (Sxx·Syy) without squaring Sxy: wrong, and negative for a falling line.
	failsWith(t, "ols", mutate("ols", "sxy * sxy / (sxx * syy)", "sxy / (sxx * syy)"), "a falling line", "")
	// No check for a vertical line: inf or nan printed.
	failsWith(t, "ols", mutate("ols", "if (sxx == 0 || syy == 0) {", "if (syy == 0) {"), "every x the same", "")
	// A fixed array: the hundred thousand points overflow it.
	failsWith(t, "ols", mutate("ols", "        if (n == cap) {\n            cap *= 2;", "        if (n == cap) {\n            return 1;\n            cap *= 2;"), "a hundred thousand points", "")

	// Ties going to the last centre.
	failsWith(t, "kmeans", mutate("kmeans", "if (c == 0 || d < bestd) {", "if (c == 0 || d <= bestd) {"), "ties go to the first centre", "")
	// Dividing by an empty cluster's zero count.
	failsWith(t, "kmeans", mutate("kmeans", "            if (count[c] > 0) {\n", "            if (1) {\n"), "an empty cluster", "")
	// A fixed number of rounds instead of running to convergence.
	failsWith(t, "kmeans", mutate("kmeans", "round < 1000", "round < 2"), "twenty thousand points", "")
	// No check that there are at least k points: reads past the end.
	failsWith(t, "kmeans", mutate("kmeans", "    if (n < (size_t)k) {", "    if (0) {"), "more clusters than points", "")

	// Ties counted as wins (or losses) instead of halves.
	failsWith(t, "metrics", mutate("metrics", "twice += pos * (2 * negbelow + neg);", "twice += pos * (2 * negbelow);"), "tied scores count half", "")
	// A strict threshold.
	failsWith(t, "metrics", mutate("metrics", "int flagged = s >= t;", "int flagged = s > t;"), "the threshold is inclusive", "")
	// Precision printed even when nothing was flagged.
	failsWith(t, "metrics", mutate("metrics", `show("precision", tp + fp > 0, precision);`, `show("precision", 1, (double)tp / (tp + fp));`), "nothing flagged", "")
	// AUC by comparing every positive with every negative: too slow.
	quadratic := mutate("metrics", "    qsort(p, n, sizeof *p, byscore);\n    long long twice = 0, negbelow = 0;\n", `    long long twice = 0, negbelow = 0;
    (void)byscore;
    for (size_t i = 0; i < n; i++)
        for (size_t j = 0; j < n; j++)
            if (p[i].label && !p[j].label)
                twice += p[i].score > p[j].score ? 2 : p[i].score == p[j].score;
    n = 0;
`)
	failsWith(t, "metrics", quadratic, "two hundred thousand predictions", "did not finish")
}
