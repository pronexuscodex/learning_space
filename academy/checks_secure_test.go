package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range secureSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var secureSolutions = map[string]string{
	"tlv_parse": `#include <stdio.h>
#include <stdlib.h>

#define LIMIT 1048576
#define MAX_DEPTH 8

/* parse checks (and, when print is set, prints) the records in p[0..n). */
static int parse(const unsigned char *p, size_t n, int depth, int print) {
    while (n > 0) {
        if (n < 3) return 0;
        unsigned type = p[0];
        size_t len = (size_t)p[1] << 8 | p[2];
        p += 3;
        n -= 3;
        if (len > n) return 0;
        const unsigned char *v = p;
        p += len;
        n -= len;
        int ind = 2 * depth;
        switch (type) {
        case 1:
            if (print) {
                printf("%*stext \"", ind, "");
                for (size_t i = 0; i < len; i++) {
                    unsigned c = v[i];
                    if (c >= ' ' && c <= '~' && c != '"' && c != '\\') putchar((int)c);
                    else printf("\\x%02x", c);
                }
                printf("\"\n");
            }
            break;
        case 2:
            if (len != 4) return 0;
            if (print)
                printf("%*snumber %lu\n", ind, "",
                       (unsigned long)v[0] << 24 | (unsigned long)v[1] << 16 | (unsigned long)v[2] << 8 | v[3]);
            break;
        case 3:
            if (depth >= MAX_DEPTH) return 0;
            if (print) printf("%*slist {\n", ind, "");
            if (!parse(v, len, depth + 1, print)) return 0;
            if (print) printf("%*s}\n", ind, "");
            break;
        default:
            if (print) printf("%*sskip type %u, %zu bytes\n", ind, "", type, len);
        }
    }
    return 1;
}

int main(void) {
    unsigned char *buf = malloc(LIMIT + 1);
    if (!buf) return 1;
    size_t n = 0, got;
    while (n <= LIMIT && (got = fread(buf + n, 1, LIMIT + 1 - n, stdin)) > 0) n += got;
    if (n > LIMIT) {
        fprintf(stderr, "tlv_parse: input longer than %d bytes\n", LIMIT);
        free(buf);
        return 1;
    }
    if (!parse(buf, n, 0, 0)) {
        fprintf(stderr, "tlv_parse: malformed input\n");
        free(buf);
        return 1;
    }
    parse(buf, n, 0, 1);
    free(buf);
    return 0;
}
`,
	"secret_scan": `#define _POSIX_C_SOURCE 200809L
#include <ctype.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct finding { size_t col; const char *kind; };

static struct finding *found;
static size_t nfound, capfound;

static void add(size_t col, const char *kind) {
    if (nfound == capfound) {
        capfound = capfound ? 2 * capfound : 16;
        found = realloc(found, capfound * sizeof *found);
        if (!found) exit(1);
    }
    found[nfound].col = col;
    found[nfound].kind = kind;
    nfound++;
}

static int alnum(int c) { return isalnum((unsigned char)c); }
static int ident(int c) { return alnum(c) || c == '_'; }
static int upper_digit(int c) { return (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'); }

static size_t find(const char *s, size_t len, size_t from, const char *needle) {
    size_t nl = strlen(needle);
    for (size_t i = from; i + nl <= len; i++)
        if (memcmp(s + i, needle, nl) == 0) return i;
    return (size_t)-1;
}

/* prefix, then exactly n bytes for which ok holds, alone in its word. */
static void prefixed(const char *s, size_t len, const char *prefix, size_t n, int (*ok)(int), const char *kind) {
    size_t pl = strlen(prefix);
    for (size_t i = 0; i + pl + n <= len; i++) {
        if (memcmp(s + i, prefix, pl) != 0 || (i > 0 && alnum(s[i - 1]))) continue;
        size_t end = i + pl + n, k;
        if (end < len && alnum(s[end])) continue;
        for (k = i + pl; k < end && ok(s[k]); k++) {
        }
        if (k == end) add(i + 1, kind);
    }
}

static int has_keyword(const char *s, size_t n) {
    static const char *words[] = {"password", "secret", "token", "api_key"};
    for (int w = 0; w < 4; w++) {
        size_t wl = strlen(words[w]);
        for (size_t i = 0; i + wl <= n; i++) {
            size_t k = 0;
            while (k < wl && tolower((unsigned char)s[i + k]) == words[w][k]) k++;
            if (k == wl) return 1;
        }
    }
    return 0;
}

static int by_col(const void *a, const void *b) {
    const struct finding *x = a, *y = b;
    return (x->col > y->col) - (x->col < y->col);
}

static void scan_line(const char *s, size_t len) {
    nfound = 0;
    prefixed(s, len, "AKIA", 16, upper_digit, "aws-access-key");
    prefixed(s, len, "ghp_", 36, alnum, "github-token");
    size_t b = find(s, len, 0, "-----BEGIN ");
    if (b != (size_t)-1 && find(s, len, b, "PRIVATE KEY-----") != (size_t)-1) add(b + 1, "private-key");
    size_t specific = nfound;
    for (size_t i = 0; i < len;) {
        if (!ident(s[i]) || (i > 0 && ident(s[i - 1]))) {
            i++;
            continue;
        }
        size_t j = i;
        while (j < len && ident(s[j])) j++;
        size_t next = j;
        if (has_keyword(s + i, j - i)) {
            size_t k = j;
            int sep = 0;
            if (k < len && (s[k] == '"' || s[k] == '\'')) k++;
            while (k < len && (s[k] == ' ' || s[k] == '\t')) k++;
            if (k + 1 < len && s[k] == ':' && s[k + 1] == '=') {
                k += 2;
                sep = 1;
            } else if (k < len && (s[k] == '=' || s[k] == ':')) {
                k++;
                sep = 1;
            }
            if (sep) {
                while (k < len && (s[k] == ' ' || s[k] == '\t')) k++;
                if (k < len && (s[k] == '"' || s[k] == '\'')) {
                    size_t start = k + 1, stop = start;
                    while (stop < len && s[stop] != s[k]) stop++;
                    if (stop < len) {
                        int inside = 0;
                        for (size_t f = 0; f < specific; f++)
                            if (found[f].col - 1 >= start && found[f].col - 1 < stop) inside = 1;
                        if (stop - start >= 8 && !inside) add(start + 1, "assignment");
                        next = stop + 1;
                    }
                }
            }
        }
        i = next;
    }
    if (nfound > 1) qsort(found, nfound, sizeof *found, by_col);
}

int main(void) {
    char *line = NULL;
    size_t cap = 0;
    ssize_t n;
    long lineno = 0, total = 0;
    while ((n = getline(&line, &cap, stdin)) != -1) {
        lineno++;
        size_t len = (size_t)n;
        if (len > 0 && line[len - 1] == '\n') len--;
        scan_line(line, len);
        for (size_t f = 0; f < nfound; f++) printf("%ld:%zu %s\n", lineno, found[f].col, found[f].kind);
        total += (long)nfound;
    }
    printf("findings: %ld\n", total);
    free(line);
    free(found);
    return 0;
}
`,
	"sha256": `#include <ctype.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const uint32_t K[64] = {
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
};

static const uint32_t H0[8] = {
    0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
};

#define ROTR(x, n) (((x) >> (n)) | ((x) << (32 - (n))))

struct sha { uint32_t h[8]; unsigned char block[64]; size_t used; uint64_t total; };

static void compress(uint32_t h[8], const unsigned char *b) {
    uint32_t w[64];
    for (int t = 0; t < 16; t++)
        w[t] = (uint32_t)b[4 * t] << 24 | (uint32_t)b[4 * t + 1] << 16 | (uint32_t)b[4 * t + 2] << 8 | b[4 * t + 3];
    for (int t = 16; t < 64; t++) {
        uint32_t s0 = ROTR(w[t - 15], 7) ^ ROTR(w[t - 15], 18) ^ (w[t - 15] >> 3);
        uint32_t s1 = ROTR(w[t - 2], 17) ^ ROTR(w[t - 2], 19) ^ (w[t - 2] >> 10);
        w[t] = s1 + w[t - 7] + s0 + w[t - 16];
    }
    uint32_t a = h[0], bb = h[1], c = h[2], d = h[3], e = h[4], f = h[5], g = h[6], hh = h[7];
    for (int t = 0; t < 64; t++) {
        uint32_t t1 = hh + (ROTR(e, 6) ^ ROTR(e, 11) ^ ROTR(e, 25)) + ((e & f) ^ (~e & g)) + K[t] + w[t];
        uint32_t t2 = (ROTR(a, 2) ^ ROTR(a, 13) ^ ROTR(a, 22)) + ((a & bb) ^ (a & c) ^ (bb & c));
        hh = g;
        g = f;
        f = e;
        e = d + t1;
        d = c;
        c = bb;
        bb = a;
        a = t1 + t2;
    }
    h[0] += a;
    h[1] += bb;
    h[2] += c;
    h[3] += d;
    h[4] += e;
    h[5] += f;
    h[6] += g;
    h[7] += hh;
}

static void update(struct sha *s, const unsigned char *p, size_t n) {
    s->total += n;
    while (n > 0) {
        size_t take = 64 - s->used;
        if (take > n) take = n;
        memcpy(s->block + s->used, p, take);
        s->used += take;
        p += take;
        n -= take;
        if (s->used == 64) {
            compress(s->h, s->block);
            s->used = 0;
        }
    }
}

static void finish(struct sha *s, char hex[65]) {
    uint64_t bits = s->total * 8;
    unsigned char pad = 0x80, zero = 0, len[8];
    update(s, &pad, 1);
    while (s->used != 56) update(s, &zero, 1);
    for (int i = 0; i < 8; i++) len[i] = (unsigned char)(bits >> (56 - 8 * i));
    update(s, len, 8);
    for (int i = 0; i < 8; i++) snprintf(hex + 8 * i, 9, "%08x", (unsigned)s->h[i]);
}

static int hash_file(FILE *f, char hex[65]) {
    static unsigned char buf[65536];
    struct sha s = {{0}, {0}, 0, 0};
    memcpy(s.h, H0, sizeof H0);
    size_t got;
    while ((got = fread(buf, 1, sizeof buf, f)) > 0) update(&s, buf, got);
    if (ferror(f)) return 0;
    finish(&s, hex);
    return 1;
}

static int usage(void) {
    fprintf(stderr, "usage: sha256 [FILE]  or  sha256 -c HEX FILE\n");
    return 2;
}

static int digest_of(const char *path, char hex[65]) {
    FILE *f = fopen(path, "rb");
    if (!f) {
        perror(path);
        return 0;
    }
    int ok = hash_file(f, hex);
    fclose(f);
    if (!ok) perror(path);
    return ok;
}

int main(int argc, char **argv) {
    char hex[65];
    if (argc == 1) {
        if (!hash_file(stdin, hex)) return 1;
        printf("%s\n", hex);
        return 0;
    }
    if (argc == 2 && strcmp(argv[1], "-c") != 0) {
        if (!digest_of(argv[1], hex)) return 1;
        printf("%s  %s\n", hex, argv[1]);
        return 0;
    }
    if (argc != 4 || strcmp(argv[1], "-c") != 0) return usage();
    const char *want = argv[2];
    if (strlen(want) != 64) return usage();
    for (int i = 0; i < 64; i++)
        if (!isxdigit((unsigned char)want[i])) return usage();
    if (!digest_of(argv[3], hex)) return 1;
    for (int i = 0; i < 64; i++)
        if (tolower((unsigned char)want[i]) != hex[i]) {
            fprintf(stderr, "%s: FAILED\n", argv[3]);
            return 1;
        }
    printf("%s: OK\n", argv[3]);
    return 0;
}
`,
}

func TestSecureChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := secureSolutions
	mutate := func(src, from, to string) string {
		t.Helper()
		out := strings.Replace(src, from, to, 1)
		if out == src {
			t.Fatalf("the mutation %q did not apply", from)
		}
		return out
	}
	old := checkCaseTimeout
	checkCaseTimeout = time.Second
	defer func() { checkCaseTimeout = old }()

	tlv := ref["tlv_parse"]
	// Clamping a length that is too long instead of refusing it. (Simply
	// trusting it reads past the input, which the sanitizers catch, but
	// without them what is read there is chance.)
	failsWith(t, "tlv_parse", mutate(tlv, "        if (len > n) return 0;", "        if (len > n) len = n;"), "a length past the end", "")
	// Length bytes read through a signed char.
	failsWith(t, "tlv_parse", mutate(tlv, "size_t len = (size_t)p[1] << 8 | p[2];", "size_t len = (size_t)((char)p[1] << 8 | (char)p[2]);"), "lengths of 128 and more", "")
	// No depth limit.
	failsWith(t, "tlv_parse", mutate(tlv, "            if (depth >= MAX_DEPTH) return 0;\n", ""), "nine nested lists", "")
	// Printing while parsing, so a bad file is half printed.
	failsWith(t, "tlv_parse", mutate(mutate(tlv, "if (!parse(buf, n, 0, 0)) {", "if (!parse(buf, n, 0, 1)) {"), "    parse(buf, n, 0, 1);\n", ""), "an error after valid records", "")
	// Control bytes printed raw.
	failsWith(t, "tlv_parse", mutate(tlv, "if (c >= ' ' && c <= '~' && c != '\"' && c != '\\\\')", "if (c != '\"' && c != '\\\\')"), "text that needs escaping", "")
	// A number of any length.
	failsWith(t, "tlv_parse", mutate(tlv, "            if (len != 4) return 0;\n", ""), "a number of 3 bytes", "")
	// No size limit.
	failsWith(t, "tlv_parse", mutate(tlv, "    if (n > LIMIT) {", "    if (n > LIMIT + 1) {"), "over a megabyte", "")

	scan := ref["secret_scan"]
	// No check of the character after the key.
	failsWith(t, "secret_scan", mutate(scan, "        if (end < len && alnum(s[end])) continue;\n", ""), "lookalikes", "")
	// Keywords compared case-sensitively.
	failsWith(t, "secret_scan", mutate(scan, "tolower((unsigned char)s[i + k]) == words[w][k]", "s[i + k] == words[w][k]"), "upper and mixed case names", "")
	// Only the first finding on each line.
	failsWith(t, "secret_scan", mutate(scan, "for (size_t f = 0; f < nfound; f++) printf", "for (size_t f = 0; f < nfound && f < 1; f++) printf"), "two on one line", "")
	// A fixed 4 KB line buffer.
	fixed := mutate(mutate(scan, "    while ((n = getline(&line, &cap, stdin)) != -1) {", "    static char buf[4096];\n    (void)cap;\n    while (fgets(buf, sizeof buf, stdin) && (line = buf, n = (ssize_t)strlen(buf), 1)) {"), "    free(line);\n", "")
	failsWith(t, "secret_scan", fixed, "a very long line", "")
	// One secret reported twice.
	failsWith(t, "secret_scan", mutate(scan, "if (stop - start >= 8 && !inside)", "if (stop - start >= 8 && (inside || 1))"), "a key in an assignment", "")
	// Columns counted from 0.
	failsWith(t, "secret_scan", mutate(scan, "        if (k == end) add(i + 1, kind);", "        if (k == end) add(i, kind);"), "the starter's example", "")

	sha := ref["sha256"]
	// The length in bytes instead of bits.
	failsWith(t, "sha256", mutate(sha, "uint64_t bits = s->total * 8;", "uint64_t bits = s->total;"), "the starter's example", "")
	// Padding that does not add a block when the length no longer fits.
	failsWith(t, "sha256", mutate(sha, "while (s->used != 56) update(s, &zero, 1);", "while (s->used < 56) update(s, &zero, 1);"), "56 bytes", "")
	// A wrong rotation in the message schedule.
	failsWith(t, "sha256", mutate(sha, "ROTR(w[t - 15], 7)", "ROTR(w[t - 15], 8)"), "the starter's example", "")
	// Reading binary data as text.
	asText := mutate(sha, "    size_t got;\n    while ((got = fread(buf, 1, sizeof buf, f)) > 0) update(&s, buf, got);", "    while (fgets((char *)buf, sizeof buf, f)) update(&s, buf, strlen((char *)buf));")
	failsWith(t, "sha256", asText, "a file", "")
	// Hex compared case-sensitively.
	failsWith(t, "sha256", mutate(sha, "if (tolower((unsigned char)want[i]) != hex[i]) {", "if (want[i] != hex[i]) {"), "a matching checksum", "")
	// A mismatch that still exits 0.
	failsWith(t, "sha256", mutate(sha, "            fprintf(stderr, \"%s: FAILED\\n\", argv[3]);\n            return 1;", "            printf(\"%s: FAILED\\n\", argv[3]);\n            return 0;"), "a tampered file", "")
	// A malformed checksum treated as a mismatch.
	failsWith(t, "sha256", mutate(sha, "    if (strlen(want) != 64) return usage();\n    for (int i = 0; i < 64; i++)\n        if (!isxdigit((unsigned char)want[i])) return usage();\n", ""), "a malformed checksum", "")
}
