package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range neuralSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var neuralSolutions = map[string]string{
	"autodiff": `#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

enum { INPUT, ADD, SUB, MUL, TANH, EXP, RELU };

struct node { char name[16]; int op, a, b; double val, grad; };

static struct node *nodes;
static int n, cap;

/* A small open-addressing hash table from names to node indexes. */
#define TABLE 65536
static int table[TABLE];

static unsigned hash(const char *s) {
    unsigned h = 2166136261u;
    while (*s) h = (h ^ (unsigned char)*s++) * 16777619u;
    return h;
}

static int *slot(const char *name) {
    for (unsigned h = hash(name) % TABLE;; h = (h + 1) % TABLE)
        if (table[h] == 0 || strcmp(nodes[table[h] - 1].name, name) == 0) return &table[h];
}

static int lookup(const char *name) {
    int v = *slot(name);
    return v - 1;
}

static int add_node(const char *name) {
    int *s = slot(name);
    if (*s != 0) return -1;
    if (n == cap) {
        cap = cap ? cap * 2 : 1024;
        nodes = realloc(nodes, cap * sizeof *nodes);
        if (!nodes) exit(1);
    }
    memset(&nodes[n], 0, sizeof nodes[n]);
    strcpy(nodes[n].name, name);
    *s = n + 1;
    return n++;
}

static int fail(const char *msg, int line) {
    fprintf(stderr, "autodiff: line %d: %s\n", line, msg);
    return 1;
}

int main(void) {
    char buf[256], a[16], b[16], c[16], d[16];
    int line = 0, inputs[4096], ninputs = 0, assigned = 0;
    while (fgets(buf, sizeof buf, stdin)) {
        line++;
        double v;
        if (sscanf(buf, "input %15s %lf", a, &v) == 2) {
            int i = add_node(a);
            if (i < 0) return fail("defined twice", line);
            nodes[i].op = INPUT;
            nodes[i].val = v;
            if (ninputs < 4096) inputs[ninputs++] = i;
            continue;
        }
        int got = sscanf(buf, "%15s = %15s %15s %15s", a, b, c, d);
        if (got < 3) {
            if (strspn(buf, " \t\r\n") == strlen(buf)) continue;
            return fail("not a statement", line);
        }
        static const char *names[] = {"", "add", "sub", "mul", "tanh", "exp", "relu"};
        int op = 0;
        for (int k = 1; k <= 6; k++)
            if (strcmp(b, names[k]) == 0) op = k;
        if (op == 0) return fail("unknown operation", line);
        int binary = op == ADD || op == SUB || op == MUL;
        if (got != (binary ? 4 : 3)) return fail("wrong number of arguments", line);
        int x = lookup(c), y = binary ? lookup(d) : -1;
        if (x < 0 || (binary && y < 0)) return fail("a name used before it is defined", line);
        int i = add_node(a);
        if (i < 0) return fail("defined twice", line);
        nodes[i].op = op;
        nodes[i].a = x;
        nodes[i].b = y;
        double xv = nodes[x].val, yv = binary ? nodes[y].val : 0;
        switch (op) {
        case ADD: nodes[i].val = xv + yv; break;
        case SUB: nodes[i].val = xv - yv; break;
        case MUL: nodes[i].val = xv * yv; break;
        case TANH: nodes[i].val = tanh(xv); break;
        case EXP: nodes[i].val = exp(xv); break;
        case RELU: nodes[i].val = xv > 0 ? xv : 0; break;
        }
        assigned = 1;
    }
    if (!assigned) return fail("no assignment", line);
    nodes[n - 1].grad = 1;
    for (int i = n - 1; i >= 0; i--) {
        struct node *p = &nodes[i];
        double g = p->grad;
        switch (p->op) {
        case ADD: nodes[p->a].grad += g; nodes[p->b].grad += g; break;
        case SUB: nodes[p->a].grad += g; nodes[p->b].grad -= g; break;
        case MUL: nodes[p->a].grad += g * nodes[p->b].val; nodes[p->b].grad += g * nodes[p->a].val; break;
        case TANH: nodes[p->a].grad += g * (1 - p->val * p->val); break;
        case EXP: nodes[p->a].grad += g * p->val; break;
        case RELU: if (nodes[p->a].val > 0) nodes[p->a].grad += g; break;
        }
    }
    printf("value %.6f\n", nodes[n - 1].val);
    for (int k = 0; k < ninputs; k++) printf("grad %s %.6f\n", nodes[inputs[k]].name, 0.0 + nodes[inputs[k]].grad);
    free(nodes);
    return 0;
}
`,
	"attention": `#include <math.h>
#include <stdio.h>
#include <stdlib.h>

int main(void) {
    int n, d;
    if (scanf("%d %d", &n, &d) != 2 || n < 1 || d < 1) {
        fprintf(stderr, "attention: the first line must be n and d, both at least 1\n");
        return 1;
    }
    double *m = malloc(3 * (size_t)n * d * sizeof *m), *s = malloc(n * sizeof *s);
    if (!m || !s) return 1;
    for (size_t i = 0; i < 3 * (size_t)n * d; i++)
        if (scanf("%lf", &m[i]) != 1) {
            fprintf(stderr, "attention: expected %d rows of %d numbers for each of Q, K and V\n", n, d);
            return 1;
        }
    double *q = m, *k = m + (size_t)n * d, *v = m + 2 * (size_t)n * d;
    double scale = 1 / sqrt(d);
    for (int i = 0; i < n; i++) {
        double mx = -INFINITY, sum = 0;
        for (int j = 0; j <= i; j++) {
            double dot = 0;
            for (int t = 0; t < d; t++) dot += q[i * d + t] * k[j * d + t];
            s[j] = dot * scale;
            if (s[j] > mx) mx = s[j];
        }
        for (int j = 0; j <= i; j++) {
            s[j] = exp(s[j] - mx);
            sum += s[j];
        }
        for (int t = 0; t < d; t++) {
            double o = 0;
            for (int j = 0; j <= i; j++) o += s[j] / sum * v[j * d + t];
            printf(t ? " %.4f" : "%.4f", o);
        }
        printf("\n");
    }
    free(m);
    free(s);
    return 0;
}
`,
	"bpe": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int count[1256][1256];

int main(int argc, char **argv) {
    char *end;
    long merges = argc == 2 ? strtol(argv[1], &end, 10) : -1;
    if (argc != 2 || end == argv[1] || *end || merges < 0 || merges > 1000) {
        fprintf(stderr, "usage: bpe MERGES (0 to 1000)\n");
        return 2;
    }
    size_t cap = 1 << 16, n = 0;
    unsigned char *buf = malloc(cap);
    if (!buf) return 1;
    size_t got;
    while ((got = fread(buf + n, 1, cap - n, stdin)) > 0) {
        n += got;
        if (n == cap) {
            cap *= 2;
            buf = realloc(buf, cap);
            if (!buf) return 1;
        }
    }
    int *tok = malloc((n + 1) * sizeof *tok);
    if (!tok) return 1;
    for (size_t i = 0; i < n; i++) tok[i] = buf[i];
    int size = 256 + (int)merges;
    for (int m = 0; m < merges; m++) {
        for (int a = 0; a < size; a++) memset(count[a], 0, size * sizeof count[a][0]);
        for (size_t i = 0; i + 1 < n; i++) count[tok[i]][tok[i + 1]]++;
        int bl = -1, br = -1, best = 1;
        for (int a = 0; a < size; a++)
            for (int b = 0; b < size; b++)
                if (count[a][b] > best) {
                    best = count[a][b];
                    bl = a;
                    br = b;
                }
        if (bl < 0) break;
        int id = 256 + m;
        printf("%d = %d %d (%d)\n", id, bl, br, best);
        size_t w = 0;
        for (size_t i = 0; i < n;) {
            if (i + 1 < n && tok[i] == bl && tok[i + 1] == br) {
                tok[w++] = id;
                i += 2;
            } else {
                tok[w++] = tok[i++];
            }
        }
        n = w;
    }
    printf("tokens %zu\n", n);
    free(buf);
    free(tok);
    return 0;
}
`,
}

func TestNeuralChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := neuralSolutions
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

	// Gradients assigned with = instead of accumulated with +=.
	failsWith(t, "autodiff", mutate("autodiff", "case MUL: nodes[p->a].grad += g * nodes[p->b].val; nodes[p->b].grad += g * nodes[p->a].val; break;",
		"case MUL: nodes[p->a].grad = g * nodes[p->b].val; nodes[p->b].grad = g * nodes[p->a].val; break;"), "a value used twice", "")
	// tanh's derivative from the input instead of the output.
	failsWith(t, "autodiff", mutate("autodiff", "nodes[p->a].grad += g * (1 - p->val * p->val);", "nodes[p->a].grad += g * (1 - nodes[p->a].val * nodes[p->a].val);"), "a tanh neuron", "")
	// sub passing +g to its second argument.
	failsWith(t, "autodiff", mutate("autodiff", "nodes[p->b].grad -= g; break;", "nodes[p->b].grad += g; break;"), "subtraction and exp", "")
	// ReLU passing gradient even when its input was negative.
	failsWith(t, "autodiff", mutate("autodiff", "case RELU: if (nodes[p->a].val > 0) nodes[p->a].grad += g; break;", "case RELU: nodes[p->a].grad += g; break;"), "a dead ReLU", "")
	// Undefined names silently read as node 0.
	failsWith(t, "autodiff", mutate("autodiff", "if (x < 0 || (binary && y < 0)) return fail", "if (0) return fail"), "a name used before it is defined", "")

	// No causal mask: every row sees the future.
	noMask := strings.ReplaceAll(ref["attention"], "j <= i", "j < n")
	failsWith(t, "attention", noMask, "the future is masked", "")
	// No 1/sqrt(d) scaling.
	failsWith(t, "attention", mutate("attention", "double scale = 1 / sqrt(d);", "double scale = 1;"), "scaled by the square root of d", "")
	// Softmax without subtracting the maximum: exp overflows to inf.
	failsWith(t, "attention", mutate("attention", "s[j] = exp(s[j] - mx);", "s[j] = exp(s[j]);"), "huge scores", "")

	// Overlapping replacements: aaaa becomes three tokens.
	failsWith(t, "bpe", mutate("bpe", "                tok[w++] = id;\n                i += 2;", "                tok[w++] = id;\n                i += 1;"), "overlapping pairs", "")
	// Ties going to the last pair found, not the first.
	failsWith(t, "bpe", mutate("bpe", "if (count[a][b] > best) {", "if (count[a][b] >= best && count[a][b] > 1) {"), "ties go to the smallest pair", "")
	// Merging pairs that occur only once.
	failsWith(t, "bpe", mutate("bpe", "int bl = -1, br = -1, best = 1;", "int bl = -1, br = -1, best = 0;"), "stopping early", "")
}
