package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range infraSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var infraSolutions = map[string]string{
	"quantize": `#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static double max_abs(const double *w, int n) {
    double m = 0;
    for (int i = 0; i < n; i++)
        if (fabs(w[i]) > m) m = fabs(w[i]);
    return m;
}

int main(int argc, char **argv) {
    if (argc != 2 || (strcmp(argv[1], "row") != 0 && strcmp(argv[1], "tensor") != 0)) {
        fprintf(stderr, "usage: quantize row|tensor\n");
        return 2;
    }
    int per_row = strcmp(argv[1], "row") == 0;
    int r, c;
    if (scanf("%d %d", &r, &c) != 2 || r < 1 || c < 1) {
        fprintf(stderr, "quantize: the first line must be R and C, both at least 1\n");
        return 1;
    }
    double *w = malloc((size_t)r * c * sizeof *w);
    if (!w) return 1;
    for (size_t i = 0; i < (size_t)r * c; i++)
        if (scanf("%lf", &w[i]) != 1) {
            fprintf(stderr, "quantize: expected %d rows of %d numbers\n", r, c);
            return 1;
        }
    double all = max_abs(w, r * c), max_err = 0, sum = 0;
    for (int i = 0; i < r; i++) {
        const double *row = w + (size_t)i * c;
        double m = per_row ? max_abs(row, c) : all;
        double scale = m / 127.0;
        printf("%.6e", scale);
        for (int j = 0; j < c; j++) {
            int q = 0;
            if (m > 0) {
                double x = round(row[j] / scale);
                q = x > 127 ? 127 : x < -127 ? -127 : (int)x;
            }
            printf(" %d", q);
            double e = fabs(row[j] - scale * q);
            if (e > max_err) max_err = e;
            sum += e;
        }
        printf("\n");
    }
    printf("max error %.3e\nmean error %.3e\n", max_err, sum / ((double)r * c));
    free(w);
    return 0;
}
`,
	"kv_blocks": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MAX_SEQS 1000

struct seq { char name[16]; int len, nblocks, *blocks; };

static struct seq seqs[MAX_SEQS];
static int nseqs, nblocks, size, *users;

static struct seq *find(const char *name) {
    for (int i = 0; i < nseqs; i++)
        if (strcmp(seqs[i].name, name) == 0) return &seqs[i];
    return NULL;
}

static struct seq *create(const char *name) {
    if (find(name) || nseqs == MAX_SEQS) return NULL;
    struct seq *s = &seqs[nseqs++];
    strcpy(s->name, name);
    s->len = s->nblocks = 0;
    s->blocks = malloc(nblocks * sizeof *s->blocks);
    if (!s->blocks) exit(1);
    return s;
}

static int count_free(void) {
    int f = 0;
    for (int i = 0; i < nblocks; i++) f += users[i] == 0;
    return f;
}

static int take(void) {
    for (int i = 0; i < nblocks; i++)
        if (users[i] == 0) {
            users[i] = 1;
            return i;
        }
    return -1;
}

static int need(int len) { return (len + size - 1) / size; }

static int number(const char *s, long lo, long hi, long *out) {
    char *end;
    *out = strtol(s, &end, 10);
    return end != s && *end == 0 && *out >= lo && *out <= hi;
}

static int fail(const char *msg, const char *line) {
    fprintf(stderr, "kv_blocks: %s: %s", msg, line);
    return 1;
}

int main(int argc, char **argv) {
    long b, sz;
    if (argc != 3 || !number(argv[1], 1, 4096, &b) || !number(argv[2], 1, 1024, &sz)) {
        fprintf(stderr, "usage: kv_blocks BLOCKS SIZE (1 to 4096, 1 to 1024)\n");
        return 2;
    }
    nblocks = (int)b;
    size = (int)sz;
    users = calloc(nblocks, sizeof *users);
    if (!users) return 1;
    char line[128], cmd[16], x[32], y[32];
    while (fgets(line, sizeof line, stdin)) {
        int got = sscanf(line, "%15s %31s %31s", cmd, x, y);
        if (got <= 0) continue;
        struct seq *s = got >= 2 ? find(x) : NULL;
        long n;
        if (got == 2 && strcmp(cmd, "new") == 0) {
            if (strlen(x) > 15 || !create(x)) return fail("cannot create", line);
        } else if (got == 3 && strcmp(cmd, "append") == 0) {
            if (!s || !number(y, 1, 1000000, &n)) return fail("bad append", line);
            int copy = s->len % size != 0 && users[s->blocks[s->nblocks - 1]] > 1;
            int extra = need(s->len + (int)n) - need(s->len);
            if (copy + extra > count_free()) {
                printf("out of memory %s\n", s->name);
                continue;
            }
            if (copy) {
                users[s->blocks[s->nblocks - 1]]--;
                s->blocks[s->nblocks - 1] = take();
            }
            for (int k = 0; k < extra; k++) s->blocks[s->nblocks++] = take();
            s->len += (int)n;
        } else if (got == 3 && strcmp(cmd, "fork") == 0) {
            if (!s || strlen(y) > 15) return fail("bad fork", line);
            int from = (int)(s - seqs);
            struct seq *t = create(y);
            if (!t) return fail("cannot create", line);
            s = &seqs[from];
            t->len = s->len;
            t->nblocks = s->nblocks;
            for (int k = 0; k < s->nblocks; k++) {
                t->blocks[k] = s->blocks[k];
                users[s->blocks[k]]++;
            }
        } else if (got == 2 && strcmp(cmd, "free") == 0) {
            if (!s) return fail("no such sequence", line);
            for (int k = 0; k < s->nblocks; k++) users[s->blocks[k]]--;
            free(s->blocks);
            *s = seqs[--nseqs];
        } else if (got == 2 && strcmp(cmd, "show") == 0) {
            if (!s) return fail("no such sequence", line);
            printf("%s len %d blocks", s->name, s->len);
            for (int k = 0; k < s->nblocks; k++) printf(" %d", s->blocks[k]);
            printf("\n");
        } else if (got == 1 && strcmp(cmd, "stats") == 0) {
            int shared = 0;
            for (int i = 0; i < nblocks; i++) shared += users[i] > 1;
            printf("free %d shared %d\n", count_free(), shared);
        } else {
            return fail("unknown command", line);
        }
    }
    for (int i = 0; i < nseqs; i++) free(seqs[i].blocks);
    free(users);
    return 0;
}
`,
	"allreduce": `#include <stdio.h>
#include <stdlib.h>

static int n, l;

static int lo(int c) { return (int)((long)c * l / n); }
static int hi(int c) { return (int)((long)(c + 1) * l / n); }

static void dump(const char *title, const long long *v) {
    printf("%s\n", title);
    for (int i = 0; i < n; i++) {
        printf("worker %d:", i);
        for (int k = 0; k < l; k++) printf(" %lld", v[(size_t)i * l + k]);
        printf("\n");
    }
}

int main(void) {
    if (scanf("%d %d", &n, &l) != 2 || n < 1 || l < 1) {
        fprintf(stderr, "allreduce: the first line must be N and L, both at least 1\n");
        return 1;
    }
    long long *v = malloc((size_t)n * l * sizeof *v);
    long *sent = calloc(n, sizeof *sent);
    if (!v || !sent) return 1;
    for (size_t i = 0; i < (size_t)n * l; i++)
        if (scanf("%lld", &v[i]) != 1) {
            fprintf(stderr, "allreduce: expected %d lines of %d integers\n", n, l);
            return 1;
        }
    for (int s = 0; s < n - 1; s++)
        for (int i = 0; i < n; i++) {
            int c = (i - s + n) % n, j = (i + 1) % n;
            for (int k = lo(c); k < hi(c); k++) v[(size_t)j * l + k] += v[(size_t)i * l + k];
            sent[i] += hi(c) - lo(c);
        }
    dump("after reduce-scatter", v);
    for (int s = 0; s < n - 1; s++)
        for (int i = 0; i < n; i++) {
            int c = (i + 1 - s + n) % n, j = (i + 1) % n;
            for (int k = lo(c); k < hi(c); k++) v[(size_t)j * l + k] = v[(size_t)i * l + k];
            sent[i] += hi(c) - lo(c);
        }
    dump("after all-gather", v);
    for (int i = 0; i < n; i++) printf("worker %d sent %ld values\n", i, sent[i]);
    free(v);
    free(sent);
    return 0;
}
`,
}

func TestInfraChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := infraSolutions
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

	// Truncating towards zero instead of rounding.
	failsWith(t, "quantize", mutate("quantize", "double x = round(row[j] / scale);", "double x = (int)(row[j] / scale);"), "rounding, not truncation", "")
	// The largest value instead of the largest magnitude.
	failsWith(t, "quantize", mutate("quantize", "if (fabs(w[i]) > m) m = fabs(w[i]);", "if (w[i] > m) m = w[i];"), "a negative outlier", "")
	// Ignoring the mode: always one scale per row.
	failsWith(t, "quantize", mutate("quantize", "double m = per_row ? max_abs(row, c) : all;", "double m = per_row || all >= 0 ? max_abs(row, c) : all;"), "one scale for the whole matrix", "")
	// Dividing by zero on a row of zeros.
	failsWith(t, "quantize", mutate("quantize", "            if (m > 0) {\n", "            {\n"), "a row of zeros", "")
	// 128 levels instead of 127: the largest weight becomes 128, outside int8.
	failsWith(t, "quantize", mutate("quantize", "double scale = m / 127.0;", "double scale = m / 128.0;"), "the starter's example", "")

	// No copy-on-write: appending into a shared block.
	failsWith(t, "kv_blocks", mutate("kv_blocks", "int copy = s->len % size != 0 && users[s->blocks[s->nblocks - 1]] > 1;", "int copy = 0;"), "copy-on-write", "")
	// Copying the last block even when it is full.
	failsWith(t, "kv_blocks", mutate("kv_blocks", "int copy = s->len % size != 0 && users[s->blocks[s->nblocks - 1]] > 1;", "int copy = s->nblocks > 0 && users[s->blocks[s->nblocks - 1]] > 1;"), "a full last block is not copied", "")
	// free releasing blocks that another sequence still uses.
	failsWith(t, "kv_blocks", mutate("kv_blocks", "for (int k = 0; k < s->nblocks; k++) users[s->blocks[k]]--;", "for (int k = 0; k < s->nblocks; k++) users[s->blocks[k]] = 0;"), "free keeps shared blocks", "")
	// Taking blocks one by one and failing halfway.
	failsWith(t, "kv_blocks", mutate("kv_blocks", "            if (copy + extra > count_free()) {\n                printf(\"out of memory %s\\n\", s->name);\n                continue;\n            }\n",
		"            if (copy + extra > count_free()) {\n                printf(\"out of memory %s\\n\", s->name);\n                if (count_free() > 0) take();\n                continue;\n            }\n"), "out of memory changes nothing", "")
	// The highest free block instead of the lowest.
	failsWith(t, "kv_blocks", mutate("kv_blocks", "    for (int i = 0; i < nblocks; i++)\n        if (users[i] == 0) {", "    for (int i = nblocks - 1; i >= 0; i--)\n        if (users[i] == 0) {"), "lowest free block first", "")
	// Carrying on after a bad command.
	failsWith(t, "kv_blocks", mutate("kv_blocks", `if (!s || !number(y, 1, 1000000, &n)) return fail("bad append", line);`, `if (!s || !number(y, 1, 1000000, &n)) { fail("bad append", line); continue; }`), "a sequence that does not exist", "")

	// Overwriting instead of adding during reduce-scatter.
	failsWith(t, "allreduce", mutate("allreduce", "v[(size_t)j * l + k] += v[(size_t)i * l + k];", "v[(size_t)j * l + k] = v[(size_t)i * l + k];"), "the starter's example", "")
	// Sending the wrong chunk: i + s instead of i − s.
	failsWith(t, "allreduce", mutate("allreduce", "int c = (i - s + n) % n, j = (i + 1) % n;", "int c = (i + s) % n, j = (i + 1) % n;"), "four workers", "")
	// N steps instead of N − 1: chunks get counted twice.
	failsWith(t, "allreduce", mutate("allreduce", "    for (int s = 0; s < n - 1; s++)\n        for (int i = 0; i < n; i++) {\n            int c = (i - s", "    for (int s = 0; s < n; s++)\n        for (int i = 0; i < n; i++) {\n            int c = (i - s"), "the starter's example", "")
	// Even chunks of L / N: the remainder is never sent.
	failsWith(t, "allreduce", mutate("allreduce", "static int hi(int c) { return (int)((long)(c + 1) * l / n); }", "static int hi(int c) { return lo(c) + l / n; }"), "uneven chunks", "")
	// int sums that overflow.
	overflow := strings.NewReplacer("long long *v", "int *v", "%lld", "%d", "const long long *v", "const int *v").Replace(ref["allreduce"])
	failsWith(t, "allreduce", overflow, "sums beyond 32 bits", "")
}
