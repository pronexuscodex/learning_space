package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range numericSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var numericSolutions = map[string]string{
	"softmax": `#include <math.h>
#include <stdio.h>
#include <stdlib.h>

#define MAXN 64

/* Prints the softmax and the loss for one line; returns 0 if it is invalid. */
static int line(char *p) {
    double z[MAXN + 1];
    int n = 0;
    for (;;) {
        char *end;
        double v = strtod(p, &end);
        if (end == p) break;
        if (n == MAXN + 1) return 0;
        z[n++] = v;
        p = end;
    }
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') p++;
    if (*p != '\0' || n < 3) return 0;
    double c = z[0];
    int k = (int)c;
    if (c != k || k < 0 || k >= n - 1) return 0;
    double *l = z + 1;
    n--;
    double m = l[0];
    for (int i = 1; i < n; i++)
        if (l[i] > m) m = l[i];
    double s = 0;
    for (int i = 0; i < n; i++) s += exp(l[i] - m);
    for (int i = 0; i < n; i++) printf("%.6f ", exp(l[i] - m) / s);
    printf("loss=%.6f\n", m + log(s) - l[k]);
    return 1;
}

int main(void) {
    char buf[4096];
    int bad = 0;
    while (fgets(buf, sizeof buf, stdin)) {
        if (!line(buf)) {
            fprintf(stderr, "softmax: not a valid line: %s", buf);
            bad = 1;
        }
    }
    return bad;
}
`,
	"power_iter": `#include <math.h>
#include <stdio.h>
#include <stdlib.h>

static void mul(int n, const double *a, const double *v, double *w) {
    for (int i = 0; i < n; i++) {
        double s = 0;
        for (int j = 0; j < n; j++) s += a[i * n + j] * v[j];
        w[i] = s;
    }
}

static double dot(int n, const double *x, const double *y) {
    double s = 0;
    for (int i = 0; i < n; i++) s += x[i] * y[i];
    return s;
}

int main(void) {
    int n;
    if (scanf("%d", &n) != 1 || n < 1 || n > 500) {
        fprintf(stderr, "power_iter: expected a size from 1 to 500\n");
        return 1;
    }
    double *a = malloc(sizeof(double) * n * n);
    double *v = malloc(sizeof(double) * n);
    double *w = malloc(sizeof(double) * n);
    if (!a || !v || !w) return 1;
    for (int i = 0; i < n * n; i++)
        if (scanf("%lf", &a[i]) != 1) {
            fprintf(stderr, "power_iter: the matrix is incomplete\n");
            return 1;
        }
    for (int i = 0; i < n; i++)
        for (int j = 0; j < i; j++)
            if (a[i * n + j] != a[j * n + i]) {
                fprintf(stderr, "power_iter: not symmetric at (%d, %d)\n", i, j);
                return 1;
            }

    for (int i = 0; i < n; i++) v[i] = i + 1;
    double norm = sqrt(dot(n, v, v));
    for (int i = 0; i < n; i++) v[i] /= norm;
    for (int step = 0; step < 100000; step++) {
        mul(n, a, v, w);
        double lambda = dot(n, v, w);
        norm = sqrt(dot(n, w, w));
        if (lambda < 0) norm = -norm;
        double change = 0;
        for (int i = 0; i < n; i++) {
            w[i] /= norm;
            double d = fabs(w[i] - v[i]);
            if (d > change) change = d;
        }
        double *t = v;
        v = w;
        w = t;
        if (change <= 1e-12) break;
    }
    mul(n, a, v, w);
    printf("%.6f\n", dot(n, v, w));
    for (int i = 0; i < n; i++)
        if (fabs(v[i]) > 1e-6) {
            if (v[i] < 0)
                for (int j = 0; j < n; j++) v[j] = -v[j];
            break;
        }
    for (int i = 0; i < n; i++) printf(i ? " %.6f" : "%.6f", v[i]);
    printf("\n");
    free(a);
    free(v);
    free(w);
    return 0;
}
`,
	"kahan_sum": `#include <stdio.h>
#include <stdlib.h>

int main(void) {
    char buf[256];
    float naive = 0, sum = 0, c = 0;
    double dbl = 0;
    int bad = 0;
    while (fgets(buf, sizeof buf, stdin)) {
        char *end, *end2;
        float x = strtof(buf, &end);
        long count = 1;
        if (end == buf) {
            bad = 1;
            continue;
        }
        while (*end == ' ' || *end == '\t') end++;
        if (*end != '\n' && *end != '\0') {
            count = strtol(end, &end2, 10);
            if (end2 == end || count < 1) {
                bad = 1;
                continue;
            }
            while (*end2 == ' ' || *end2 == '\t' || *end2 == '\r') end2++;
            if (*end2 != '\n' && *end2 != '\0') {
                bad = 1;
                continue;
            }
        }
        for (long i = 0; i < count; i++) {
            naive += x;
            float y = x - c;
            float t = sum + y;
            c = (t - sum) - y;
            sum = t;
            dbl += x;
        }
    }
    if (bad) {
        fprintf(stderr, "kahan_sum: every line must be a number and an optional count of at least 1\n");
        return 1;
    }
    printf("naive %.4f\nkahan %.4f\ndouble %.4f\n", naive, sum, dbl);
    return 0;
}
`,
}

func TestNumericChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := numericSolutions
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

	// The textbook formula, without subtracting the maximum: inf / inf.
	naive := strings.ReplaceAll(mutate("softmax", "    for (int i = 1; i < n; i++)\n        if (l[i] > m) m = l[i];\n", ""), "double m = l[0];", "double m = 0;")
	failsWith(t, "softmax", naive, "large logits", "")
	// The loss as -log(p): infinite once p rounds to 0.
	failsWith(t, "softmax", mutate("softmax", `printf("loss=%.6f\n", m + log(s) - l[k]);`, `printf("loss=%.6f\n", -log(exp(l[k] - m) / s));`), "a hopeless prediction", "")
	// Subtracting the minimum instead of the maximum still overflows.
	failsWith(t, "softmax", mutate("softmax", "if (l[i] > m) m = l[i];", "if (l[i] < m) m = l[i];"), "a hopeless prediction", "")
	// Counting the class from 1.
	failsWith(t, "softmax", mutate("softmax", "- l[k]);", "- l[k - 1 < 0 ? 0 : k - 1]);"), "the class is counted from 0", "")
	// No range check on the class.
	failsWith(t, "softmax", mutate("softmax", "k < 0 || k >= n - 1", "k < 0"), "a class out of range", "")

	// The size of A·v instead of v·A·v: a negative eigenvalue comes out positive.
	failsWith(t, "power_iter", mutate("power_iter", `printf("%.6f\n", dot(n, v, w));`, `printf("%.6f\n", sqrt(dot(n, w, w)));`), "a negative eigenvalue", "")
	// Starting from all ones: stuck on the wrong eigenvector.
	failsWith(t, "power_iter", mutate("power_iter", "v[i] = i + 1;", "v[i] = 1;"), "a start on the wrong eigenvector", "")
	// No negation for a negative λ: v flips sign every step and runs every one of the 100,000.
	failsWith(t, "power_iter", mutate("power_iter", "        if (lambda < 0) norm = -norm;\n", "        (void)lambda;\n"), "a 300 × 300 matrix", "")
	// A fixed few steps instead of iterating to convergence.
	failsWith(t, "power_iter", mutate("power_iter", "step < 100000", "step < 20"), "a slowly converging matrix", "")
	// Stopping when λ settles: it converges twice as fast as v, so v is still off in the sixth decimal.
	lambdaStop := mutate("power_iter", "    for (int step = 0; step < 100000; step++) {\n", "    double prev = 1e300;\n    for (int step = 0; step < 100000; step++) {\n")
	lambdaStop = strings.Replace(lambdaStop, "        if (change <= 1e-12) break;", "        (void)change;\n        if (fabs(lambda - prev) <= 1e-12) break;\n        prev = lambda;", 1)
	failsWith(t, "power_iter", lambdaStop, "a slowly converging matrix", "")
	// No symmetry check.
	failsWith(t, "power_iter", mutate("power_iter", "            if (a[i * n + j] != a[j * n + i]) {", "            if (0) {"), "not symmetric", "")

	// The correction applied with the wrong sign: the loss is added twice instead of cancelled.
	failsWith(t, "kahan_sum", mutate("kahan_sum", "float y = x - c;", "float y = x + c;"), "ten million tenths", "")
	// The "algebraically zero" correction dropped, as -ffast-math would.
	failsWith(t, "kahan_sum", mutate("kahan_sum", "            c = (t - sum) - y;\n", ""), "ten million tenths", "")
	// The double sum reading the decimal as a double, not the same float.
	failsWith(t, "kahan_sum", mutate("kahan_sum", "            dbl += x;", "            dbl += 0.1;"), "ten million tenths", "")
	// The count kept in a short: a hundred million wraps around.
	failsWith(t, "kahan_sum", mutate("kahan_sum", "for (long i = 0; i < count; i++)", "for (long i = 0; i < (short)count; i++)"), "a hundred million", "")
	// A zero count accepted.
	failsWith(t, "kahan_sum", mutate("kahan_sum", "count < 1", "count < 0"), "a zero count", "")
}
