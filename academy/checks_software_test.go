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
}
