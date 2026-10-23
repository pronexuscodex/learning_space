package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range softwareSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var softwareSolutions = map[string]string{
	"git_blob": `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct sha1 { uint32_t h[5]; unsigned char block[64]; size_t used; uint64_t length; };

static uint32_t rotl(uint32_t x, int n) { return (x << n) | (x >> (32 - n)); }

static void compress(struct sha1 *s, const unsigned char *p) {
    uint32_t w[80];
    for (int t = 0; t < 16; t++)
        w[t] = (uint32_t)p[4 * t] << 24 | (uint32_t)p[4 * t + 1] << 16 | (uint32_t)p[4 * t + 2] << 8 | p[4 * t + 3];
    for (int t = 16; t < 80; t++) w[t] = rotl(w[t - 3] ^ w[t - 8] ^ w[t - 14] ^ w[t - 16], 1);
    uint32_t a = s->h[0], b = s->h[1], c = s->h[2], d = s->h[3], e = s->h[4];
    for (int t = 0; t < 80; t++) {
        uint32_t f, k;
        if (t < 20) { f = (b & c) | (~b & d); k = 0x5A827999; }
        else if (t < 40) { f = b ^ c ^ d; k = 0x6ED9EBA1; }
        else if (t < 60) { f = (b & c) | (b & d) | (c & d); k = 0x8F1BBCDC; }
        else { f = b ^ c ^ d; k = 0xCA62C1D6; }
        uint32_t tmp = rotl(a, 5) + f + e + k + w[t];
        e = d; d = c; c = rotl(b, 30); b = a; a = tmp;
    }
    s->h[0] += a; s->h[1] += b; s->h[2] += c; s->h[3] += d; s->h[4] += e;
}

static void update(struct sha1 *s, const unsigned char *p, size_t n) {
    s->length += n;
    while (n > 0) {
        size_t take = 64 - s->used < n ? 64 - s->used : n;
        memcpy(s->block + s->used, p, take);
        s->used += take; p += take; n -= take;
        if (s->used == 64) { compress(s, s->block); s->used = 0; }
    }
}

static void finish(struct sha1 *s, unsigned char out[20]) {
    uint64_t bits = s->length * 8;
    unsigned char pad = 0x80, zero = 0, len[8];
    update(s, &pad, 1);
    while (s->used != 56) update(s, &zero, 1);
    for (int i = 0; i < 8; i++) len[i] = (unsigned char)(bits >> (56 - 8 * i));
    update(s, len, 8);
    for (int i = 0; i < 20; i++) out[i] = (unsigned char)(s->h[i / 4] >> (24 - 8 * (i % 4)));
}

int main(void) {
    size_t n = 0, cap = 1 << 16;
    unsigned char *data = malloc(cap);
    if (data == NULL) return 1;
    size_t got;
    while ((got = fread(data + n, 1, cap - n, stdin)) > 0) {
        n += got;
        if (n == cap) {
            cap *= 2;
            unsigned char *more = realloc(data, cap);
            if (more == NULL) return 1;
            data = more;
        }
    }
    struct sha1 s = {{0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0}, {0}, 0, 0};
    char header[32];
    int h = snprintf(header, sizeof header, "blob %zu", n);
    update(&s, (const unsigned char *)header, (size_t)h + 1);   /* the zero byte too */
    update(&s, data, n);
    unsigned char digest[20];
    finish(&s, digest);
    for (int i = 0; i < 20; i++) printf("%02x", digest[i]);
    printf("\n");
    free(data);
    return 0;
}
`,
	"make_order": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define BUCKETS (1 << 18)

struct node { char name[64]; int has_rule, state, *deps, ndeps, next; };

static struct node *nodes;
static int count, capacity, head[BUCKETS];

static unsigned hash(const char *s) {
    unsigned h = 2166136261u;
    for (; *s; s++) h = (h ^ (unsigned char)*s) * 16777619u;
    return h & (BUCKETS - 1);
}

static int lookup(const char *name) {
    unsigned h = hash(name);
    for (int i = head[h]; i != -1; i = nodes[i].next)
        if (strcmp(nodes[i].name, name) == 0) return i;
    if (count == capacity) {
        capacity = capacity ? capacity * 2 : 1024;
        nodes = realloc(nodes, (size_t)capacity * sizeof *nodes);
        if (nodes == NULL) exit(3);
    }
    struct node *n = &nodes[count];
    memset(n, 0, sizeof *n);
    snprintf(n->name, sizeof n->name, "%s", name);
    n->next = head[h];
    head[h] = count;
    return count++;
}

/* 0 not started, 1 in progress, 2 done; returns 0 on a cycle */
static int visit(int t) {
    if (nodes[t].state == 2) return 1;
    if (nodes[t].state == 1) return 0;
    nodes[t].state = 1;
    for (int i = 0; i < nodes[t].ndeps; i++) {
        int d = nodes[t].deps[i];
        if (nodes[d].has_rule && !visit(d)) return 0;
    }
    nodes[t].state = 2;
    puts(nodes[t].name);
    return 1;
}

int main(int argc, char **argv) {
    if (argc < 2) { fprintf(stderr, "usage: make_order TARGET < rules\n"); return 2; }
    for (int i = 0; i < BUCKETS; i++) head[i] = -1;
    char line[4097];
    while (fgets(line, sizeof line, stdin)) {
        char *colon = strchr(line, ':');
        if (colon == NULL) continue;
        *colon = '\0';
        char *name = strtok(line, " \t\n");
        if (name == NULL) continue;
        int t = lookup(name);
        nodes[t].has_rule = 1;
        int cap = 0;
        for (char *d = strtok(colon + 1, " \t\n"); d; d = strtok(NULL, " \t\n")) {
            int di = lookup(d);
            if (nodes[t].ndeps == cap) {
                cap = cap ? cap * 2 : 4;
                nodes[t].deps = realloc(nodes[t].deps, (size_t)cap * sizeof(int));
                if (nodes[t].deps == NULL) return 3;
            }
            nodes[t].deps[nodes[t].ndeps++] = di;
        }
    }
    int target = -1;
    for (int i = 0; i < count; i++)
        if (strcmp(nodes[i].name, argv[1]) == 0) target = i;
    if (target == -1 || !nodes[target].has_rule) { fprintf(stderr, "make_order: no rule to make %s\n", argv[1]); return 2; }
    if (!visit(target)) { fflush(stdout); fprintf(stderr, "make_order: circular dependency\n"); return 1; }
    for (int i = 0; i < count; i++) free(nodes[i].deps);
    free(nodes);
    return 0;
}
`,
	"bisect": `long first_bad(long good, long bad, int (*is_bad)(long)) {
    while (bad - good > 1) {
        long mid = good + (bad - good) / 2;
        if (is_bad(mid)) bad = mid;
        else good = mid;
    }
    return bad;
}
`,
	"nfa_match": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* Thompson's construction and simulation, after Russ Cox. */
enum { MATCH = 256, SPLIT = 257, ANY = 258 };

struct state { int c; struct state *out, *out1; int lastlist; };

static struct state *states;
static int nstates, listid;

static struct state *state(int c, struct state *out, struct state *out1) {
    struct state *s = &states[nstates++];
    s->c = c; s->out = out; s->out1 = out1; s->lastlist = 0;
    return s;
}

/* A fragment: its start, and the dangling pointers to patch (a list
   threaded through the pointers themselves). */
union ptrlist { union ptrlist *next; struct state *s; };
struct frag { struct state *start; union ptrlist *out; };

static union ptrlist *list1(struct state **p) { union ptrlist *l = (union ptrlist *)p; l->next = NULL; return l; }
static void patch(union ptrlist *l, struct state *s) {
    for (union ptrlist *next; l; l = next) { next = l->next; l->s = s; }
}
static union ptrlist *append(union ptrlist *a, union ptrlist *b) {
    union ptrlist *old = a;
    while (a->next) a = a->next;
    a->next = b;
    return old;
}

/* Infix to postfix with '.' as explicit concatenation; NULL if invalid.
   Literal '.' in the pattern becomes ANY, encoded as byte 1. */
static int *postfix(const char *re) {
    size_t n = strlen(re);
    int *out = malloc((2 * n + 2) * sizeof *out), *o = out;
    struct { int alt, atom; } paren[1000], *p = paren;
    int alt = 0, atom = 0;
    if (out == NULL) return NULL;
    for (; *re; re++) {
        switch (*re) {
        case '(':
            if (atom > 1) { --atom; *o++ = '.'; }
            if (p >= paren + 1000) { free(out); return NULL; }
            p->alt = alt; p->atom = atom; p++;
            alt = 0; atom = 0;
            break;
        case '|':
            if (atom == 0) { free(out); return NULL; }
            while (--atom > 0) *o++ = '.';
            alt++;
            break;
        case ')':
            if (p == paren || atom == 0) { free(out); return NULL; }
            while (--atom > 0) *o++ = '.';
            for (; alt > 0; alt--) *o++ = '|';
            --p; alt = p->alt; atom = p->atom; atom++;
            break;
        case '*': case '+': case '?':
            if (atom == 0) { free(out); return NULL; }
            *o++ = *re;
            break;
        default:
            if (atom > 1) { --atom; *o++ = '.'; }
            *o++ = *re == '.' ? ANY : (unsigned char)*re;
            atom++;
            break;
        }
    }
    if (p != paren || atom == 0) { free(out); return NULL; }
    while (--atom > 0) *o++ = '.';
    for (; alt > 0; alt--) *o++ = '|';
    *o = 0;
    return out;
}

static struct state *compile(const int *post, size_t len) {
    struct frag *stack = malloc((len + 1) * sizeof *stack), *sp = stack, e1, e2;
    struct state *s;
    for (; *post; post++) {
        switch (*post) {
        case '.': e2 = *--sp; e1 = *--sp; patch(e1.out, e2.start); *sp++ = (struct frag){e1.start, e2.out}; break;
        case '|': e2 = *--sp; e1 = *--sp; s = state(SPLIT, e1.start, e2.start); *sp++ = (struct frag){s, append(e1.out, e2.out)}; break;
        case '?': e1 = *--sp; s = state(SPLIT, e1.start, NULL); *sp++ = (struct frag){s, append(e1.out, list1(&s->out1))}; break;
        case '*': e1 = *--sp; s = state(SPLIT, e1.start, NULL); patch(e1.out, s); *sp++ = (struct frag){s, list1(&s->out1)}; break;
        case '+': e1 = *--sp; s = state(SPLIT, e1.start, NULL); patch(e1.out, s); *sp++ = (struct frag){e1.start, list1(&s->out1)}; break;
        default: s = state(*post, NULL, NULL); *sp++ = (struct frag){s, list1(&s->out)}; break;
        }
    }
    e1 = *--sp;
    patch(e1.out, state(MATCH, NULL, NULL));
    free(stack);
    return e1.start;
}

struct list { struct state **s; int n; };

static void addstate(struct list *l, struct state *s) {
    if (s == NULL || s->lastlist == listid) return;
    s->lastlist = listid;
    if (s->c == SPLIT) { addstate(l, s->out); addstate(l, s->out1); return; }
    l->s[l->n++] = s;
}

static int matches(struct state *start, const char *line, struct list *a, struct list *b) {
    listid++; a->n = 0; addstate(a, start);
    for (; *line; line++) {
        int c = (unsigned char)*line;
        listid++; b->n = 0;
        for (int i = 0; i < a->n; i++)
            if (a->s[i]->c == c || a->s[i]->c == ANY) addstate(b, a->s[i]->out);
        struct list t = *a; *a = *b; *b = t;
    }
    for (int i = 0; i < a->n; i++) if (a->s[i]->c == MATCH) return 1;
    return 0;
}

int main(int argc, char **argv) {
    if (argc < 2) { fprintf(stderr, "usage: nfa_match PATTERN\n"); return 2; }
    int *post = postfix(argv[1]);
    if (post == NULL) { fprintf(stderr, "nfa_match: invalid pattern\n"); return 2; }
    size_t len = 0;
    while (post[len]) len++;
    states = calloc(len + 2, sizeof *states);
    struct state *start = compile(post, len);
    struct list a = {malloc((len + 2) * sizeof(struct state *)), 0}, b = {malloc((len + 2) * sizeof(struct state *)), 0};
    size_t cap = 1 << 16;
    char *line = malloc(cap);
    int c, n = 0;
    while ((c = getchar()) != EOF) {
        if ((size_t)n + 1 >= cap) { cap *= 2; line = realloc(line, cap); }
        if (c != '\n') { line[n++] = (char)c; continue; }
        line[n] = 0;
        puts(matches(start, line, &a, &b) ? "match" : "no");
        n = 0;
    }
    free(line); free(a.s); free(b.s); free(states); free(post);
    return 0;
}
`,
	"turing": `#include <stdio.h>
#include <stdlib.h>

#define CELLS 200001
#define HALT 26

struct rule { unsigned char write; signed char move; unsigned char next; };

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    long limit = atol(argv[1]);
    struct rule rule[27][2] = {{{0, 0, 0}}};
    char s[8], r[8], w[8], m[8], nx[8];
    int start = -1;
    while (scanf("%7s %7s %7s %7s %7s", s, r, w, m, nx) == 5) {
        int st = s[0] - 'A', sym = r[0] - '0';
        if (start < 0) start = st;
        rule[st][sym].write = (unsigned char)(w[0] - '0');
        rule[st][sym].move = m[0] == 'L' ? -1 : 1;
        rule[st][sym].next = (unsigned char)(nx[0] == 'H' ? HALT : nx[0] - 'A');
    }
    unsigned char *tape = calloc(CELLS, 1);
    long pos = CELLS / 2, steps = 0;
    int state = start;
    while (steps < limit) {
        const struct rule *ru = &rule[state][tape[pos]];
        tape[pos] = ru->write;
        pos += ru->move;
        state = ru->next;
        steps++;
        if (state == HALT) break;
    }
    if (state == HALT) {
        long ones = 0;
        for (long i = 0; i < CELLS; i++) ones += tape[i];
        printf("halted after %ld steps, %ld ones\n", steps, ones);
    } else {
        printf("still running after %ld steps\n", limit);
    }
    free(tape);
    return 0;
}
`,
	"dpll": `#include <stdio.h>
#include <stdlib.h>

static int nvars, nclauses, *lits, *start, *value, *trail, ntrail;

static int lit_value(int l) {   /* 1 true, 0 false, -1 unassigned */
    int v = value[abs(l)];
    if (v < 0) return -1;
    return (l > 0) == v;
}

static void assign(int l) { value[abs(l)] = l > 0; trail[ntrail++] = abs(l); }

/* Unit propagation; 0 on a conflict. */
static int propagate(void) {
    for (int changed = 1; changed;) {
        changed = 0;
        for (int c = 0; c < nclauses; c++) {
            int unassigned = 0, last = 0, sat = 0;
            for (int i = start[c]; i < start[c + 1]; i++) {
                int v = lit_value(lits[i]);
                if (v == 1) { sat = 1; break; }
                if (v < 0) { unassigned++; last = lits[i]; }
            }
            if (sat) continue;
            if (unassigned == 0) return 0;
            if (unassigned == 1) { assign(last); changed = 1; }
        }
    }
    return 1;
}

static int choose(void) {   /* a literal from the shortest open clause, or 0 */
    int best = 0, bestlen = 1 << 30;
    for (int c = 0; c < nclauses; c++) {
        int len = 0, pick = 0, sat = 0;
        for (int i = start[c]; i < start[c + 1]; i++) {
            int v = lit_value(lits[i]);
            if (v == 1) { sat = 1; break; }
            if (v < 0) { len++; pick = lits[i]; }
        }
        if (!sat && len < bestlen) { bestlen = len; best = pick; }
    }
    return best;
}

static int solve(void) {
    int mark = ntrail;
    if (!propagate()) goto undo;
    int l = choose();
    if (l == 0) return 1;
    for (int k = 0; k < 2; k++) {
        int m = ntrail;
        assign(k == 0 ? l : -l);
        if (solve()) return 1;
        while (ntrail > m) value[trail[--ntrail]] = -1;
    }
undo:
    while (ntrail > mark) value[trail[--ntrail]] = -1;
    return 0;
}

int main(void) {
    int c;
    while ((c = getchar()) == 'c' || c == '\n') {
        if (c == 'c') while ((c = getchar()) != '\n' && c != EOF) {}
    }
    if (c != 'p' || scanf(" cnf %d %d", &nvars, &nclauses) != 2) return 1;
    int cap = 1024, n = 0, k = 0, x;
    lits = malloc((size_t)cap * sizeof *lits);
    start = malloc(((size_t)nclauses + 1) * sizeof *start);
    start[0] = 0;
    while (k < nclauses && scanf("%d", &x) == 1) {
        if (x == 0) { start[++k] = n; continue; }
        if (n == cap) { cap *= 2; lits = realloc(lits, (size_t)cap * sizeof *lits); }
        lits[n++] = x;
    }
    nclauses = k;
    value = malloc(((size_t)nvars + 1) * sizeof *value);
    trail = malloc(((size_t)nvars + 1) * sizeof *trail);
    for (int i = 0; i <= nvars; i++) value[i] = -1;
    if (!solve()) { puts("UNSAT"); return 0; }
    puts("SAT");
    const char *sep = "";
    for (int i = 1; i <= nvars; i++)
        if (value[i] == 1) { printf("%s%d", sep, i); sep = " "; }
    printf("\n");
    free(lits); free(start); free(value); free(trail);
    return 0;
}
`,
}

// Each Stage 9 to 12 check fails the typical wrong answer, for the reason
// its lesson teaches.
func TestSoftwareChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := softwareSolutions
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

	// Padding to 56 with an off-by-one: only the boundary lengths fail.
	failsWith(t, "git_blob", mutate("git_blob", "    while (s->used != 56) update(s, &zero, 1);", "    while (s->used != 56 && s->used != 57) update(s, &zero, 1);"), "a 56-byte message", "")
	// The length in bytes instead of bits.
	failsWith(t, "git_blob", mutate("git_blob", "    uint64_t bits = s->length * 8;", "    uint64_t bits = s->length;"), "an empty file", "")
	// The header without its zero byte.
	failsWith(t, "git_blob", mutate("git_blob", "(size_t)h + 1);", "(size_t)h);"), "hello and a newline", "")
	// Reading text with fgets stops nothing, but strlen does: zero bytes cut the file short.
	failsWith(t, "git_blob", mutate("git_blob", "    update(&s, data, n);", "    data[n < cap ? n : cap - 1] = 0;\n    update(&s, data, strlen((char *)data));"), "every byte value, 2 MB", "")

	// No cycle detection: the recursion never ends.
	failsWith(t, "make_order", mutate("make_order", "    if (nodes[t].state == 1) return 0;\n", ""), "a cycle", "")
	// Printing a target before its prerequisites (pre-order).
	failsWith(t, "make_order", strings.Replace(mutate("make_order", "    nodes[t].state = 1;\n", "    nodes[t].state = 1;\n    puts(nodes[t].name);\n"), "    nodes[t].state = 2;\n    puts(nodes[t].name);", "    nodes[t].state = 2;", 1), "the example", "")
	// Forgetting that a target was done: shared prerequisites printed twice.
	failsWith(t, "make_order", strings.Replace(mutate("make_order", "    if (nodes[t].state == 2) return 1;\n", ""), "    nodes[t].state = 2;\n", "    nodes[t].state = 0;\n", 1), "a shared prerequisite, once", "")

	// (good + bad) / 2 overflows near LONG_MAX.
	overflow := mutate("bisect", "        long mid = good + (bad - good) / 2;", "        long mid = (good + bad) / 2;")
	if run := runSource(t, "bisect", overflow); run.Sanitizers {
		for _, c := range run.Cases {
			if c.Case.Name == "near LONG_MAX" && (c.Pass || !strings.Contains(c.Problem, "UndefinedBehaviorSanitizer")) {
				t.Errorf("bisect: (good + bad) / 2 must be reported as undefined behaviour: pass=%v problem=%q", c.Pass, c.Problem)
			}
		}
	}
	// Testing every version in turn.
	failsWith(t, "bisect", "long first_bad(long good, long bad, int (*is_bad)(long)) {\n    for (long v = good + 1; v < bad; v++)\n        if (is_bad(v)) return v;\n    return bad;\n}\n", "a thousand commits", "")
	// Testing the known endpoints too.
	failsWith(t, "bisect", mutate("bisect", "    while (bad - good > 1) {", "    if (is_bad(good)) return good;\n    while (bad - good > 1) {"), "eight versions", "")

	// A backtracking matcher: correct, but exponential on nested repetition.
	backtrack := `#include <stdio.h>
#include <string.h>

/* Matches the subset without parentheses or |: c, ., c*, c+, c? */
static int here(const char *re, const char *s) {
    if (*re == 0) return *s == 0;
    int one = *s && (re[0] == '.' || re[0] == *s);
    if (re[1] == '*') return here(re + 2, s) || (one && here(re, s + 1));
    if (re[1] == '+') return one && (here(re + 2, s + 1) || here(re, s + 1));
    if (re[1] == '?') return here(re + 2, s) || (one && here(re + 2, s + 1));
    return one && here(re + 1, s + 1);
}

int main(int argc, char **argv) {
    char line[200000];
    if (argc < 2) return 2;
    while (fgets(line, sizeof line, stdin)) {
        line[strcspn(line, "\n")] = 0;
        puts(here(argv[1], line) ? "match" : "no");
    }
    return 0;
}
`
	failsWith(t, "nfa_match", backtrack, "precedence", "")
	// Matching anywhere in the line instead of the whole line.
	failsWith(t, "nfa_match", mutate("nfa_match", "    for (int i = 0; i < a->n; i++) if (a->s[i]->c == MATCH) return 1;\n    return 0;", "    for (int i = 0; i < a->n; i++) if (a->s[i]->c == MATCH) return 1;\n    return a->n > 0;"), "stars", "")
	// Accepting a star with nothing before it.
	failsWith(t, "nfa_match", mutate("nfa_match", "        case '*': case '+': case '?':\n            if (atom == 0) { free(out); return NULL; }", "        case '*': case '+': case '?':\n            if (atom == 0) break;"), "a star with nothing before it", "")

	// Not counting the move into H as a step.
	failsWith(t, "turing", mutate("turing", "        steps++;\n        if (state == HALT) break;", "        if (state == HALT) break;\n        steps++;"), "the two-state busy beaver", "")
	// Counting every 1 written instead of the 1s left on the tape.
	failsWith(t, "turing", strings.NewReplacer(
		"    long pos = CELLS / 2, steps = 0;", "    long pos = CELLS / 2, steps = 0, written = 0;",
		"        tape[pos] = ru->write;", "        tape[pos] = ru->write;\n        written += ru->write;",
		"        for (long i = 0; i < CELLS; i++) ones += tape[i];", "        ones = written;",
	).Replace(ref["turing"]), "the four-state busy beaver", "")

	// Trying every assignment and checking at the end: correct, but 2⁷²⁹ tries.
	enumerate := strings.NewReplacer(
		"    int mark = ntrail;\n    if (!propagate()) goto undo;\n    int l = choose();\n    if (l == 0) return 1;",
		"    int mark = ntrail;\n    (void)choose;\n    if (ntrail == nvars) { if (!propagate()) goto undo; return 1; }\n    int l = 1;\n    while (value[l] >= 0) l++;",
	).Replace(ref["dpll"])
	if enumerate == ref["dpll"] {
		t.Fatal("dpll: the enumeration mutation did not apply")
	}
	failsWith(t, "dpll", enumerate, "the hardest Sudoku", "did not finish")
	// Forgetting to undo a failed branch.
	failsWith(t, "dpll", strings.NewReplacer(
		"        while (ntrail > m) value[trail[--ntrail]] = -1;\n", "        (void)m;\n",
		"undo:\n    while (ntrail > mark) value[trail[--ntrail]] = -1;\n", "undo:\n    (void)mark;\n",
	).Replace(ref["dpll"]), "four pigeons, three holes", "")
	// Reading clauses line by line.
	failsWith(t, "dpll", mutate("dpll", "        if (x == 0) { start[++k] = n; continue; }", "        if (x == 0) { start[++k] = n; while ((c = getchar()) != '\\n' && c != EOF) {} continue; }"), "clauses over several lines", "")
}
