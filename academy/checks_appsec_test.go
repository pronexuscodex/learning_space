package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range appsecSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var appsecSolutions = map[string]string{
	"pbkdf2": `#define _POSIX_C_SOURCE 200809L
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

static void init(struct sha *s) {
    static const uint32_t h0[8] = {0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19};
    memcpy(s->h, h0, sizeof h0);
    s->used = 0;
    s->total = 0;
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

static void finish(struct sha *s, unsigned char out[32]) {
    uint64_t bits = s->total * 8;
    unsigned char pad = 0x80, zero = 0, len[8];
    update(s, &pad, 1);
    while (s->used != 56) update(s, &zero, 1);
    for (int i = 0; i < 8; i++) len[i] = (unsigned char)(bits >> (56 - 8 * i));
    update(s, len, 8);
    for (int i = 0; i < 8; i++)
        for (int j = 0; j < 4; j++) out[4 * i + j] = (unsigned char)(s->h[i] >> (24 - 8 * j));
}

/* The inner and outer states after hashing the padded key, computed once. */
struct hmac { struct sha inner, outer; };

static void hmac_init(struct hmac *m, const unsigned char *key, size_t klen) {
    unsigned char k[64] = {0}, pad[64];
    if (klen > 64) {
        struct sha s;
        init(&s);
        update(&s, key, klen);
        finish(&s, k);
    } else if (klen) {
        memcpy(k, key, klen);
    }
    init(&m->inner);
    for (int i = 0; i < 64; i++) pad[i] = k[i] ^ 0x36;
    update(&m->inner, pad, 64);
    init(&m->outer);
    for (int i = 0; i < 64; i++) pad[i] = k[i] ^ 0x5c;
    update(&m->outer, pad, 64);
}

static void hmac_run(const struct hmac *m, const unsigned char *msg, size_t n, unsigned char out[32]) {
    struct sha s = m->inner;
    unsigned char inner[32];
    update(&s, msg, n);
    finish(&s, inner);
    s = m->outer;
    update(&s, inner, 32);
    finish(&s, out);
}

static void pbkdf2(const char *pw, size_t pwlen, const unsigned char *salt, size_t slen, long iter, unsigned char out[32]) {
    struct hmac m;
    hmac_init(&m, (const unsigned char *)pw, pwlen);
    unsigned char first[64 + 4], u[32];
    memcpy(first, salt, slen);
    memcpy(first + slen, "\0\0\0\1", 4);
    hmac_run(&m, first, slen + 4, u);
    memcpy(out, u, 32);
    for (long i = 1; i < iter; i++) {
        hmac_run(&m, u, 32, u);
        for (int k = 0; k < 32; k++) out[k] ^= u[k];
    }
}

static int hexval(int c) { return c >= '0' && c <= '9' ? c - '0' : c >= 'a' && c <= 'f' ? c - 'a' + 10 : -1; }

static long parse_hex(const char *s, unsigned char *out, size_t max) {
    size_t n = strlen(s);
    if (n % 2 || n / 2 > max) return -1;
    for (size_t i = 0; i < n; i += 2) {
        int hi = hexval(s[i]), lo = hexval(s[i + 1]);
        if (hi < 0 || lo < 0) return -1;
        out[i / 2] = (unsigned char)(hi << 4 | lo);
    }
    return (long)(n / 2);
}

static long parse_iter(const char *s) {
    char *end;
    long v = strtol(s, &end, 10);
    if (end == s || *end || v < 1 || v > 1000000) return -1;
    return v;
}

static int read_line(char **line, size_t *cap, size_t *len) {
    ssize_t got = getline(line, cap, stdin);
    if (got < 0) return 0;
    *len = (size_t)got;
    if (*len > 0 && (*line)[*len - 1] == '\n') (*line)[--*len] = 0;
    return 1;
}

static int usage(const char *why) {
    fprintf(stderr, "pbkdf2: %s\nusage: pbkdf2 hash ITERATIONS SALTHEX | pbkdf2 verify\n", why);
    return 2;
}

int main(int argc, char **argv) {
    char *line = NULL, *pw = NULL;
    size_t cap = 0, pcap = 0, len, pwlen;
    unsigned char salt[64], dk[32];
    if (argc == 4 && strcmp(argv[1], "hash") == 0) {
        long iter = parse_iter(argv[2]), slen = parse_hex(argv[3], salt, sizeof salt);
        if (iter < 0 || slen < 0) return usage("bad iteration count or salt");
        if (!read_line(&pw, &pcap, &pwlen)) return usage("no password");
        pbkdf2(pw, pwlen, salt, (size_t)slen, iter, dk);
        printf("pbkdf2-sha256$%ld$%s$", iter, argv[3]);
        for (int i = 0; i < 32; i++) printf("%02x", dk[i]);
        printf("\n");
        free(pw);
        return 0;
    }
    if (argc != 2 || strcmp(argv[1], "verify") != 0) return usage("unknown command");
    if (!read_line(&line, &cap, &len) || !read_line(&pw, &pcap, &pwlen)) return usage("expected a stored hash and a password");
    char *f[4], *p = line;
    for (int i = 0; i < 4; i++) {
        f[i] = p;
        p = strchr(p, '$');
        if (i < 3) {
            if (!p) return usage("malformed stored hash");
            *p++ = 0;
        } else if (p) {
            return usage("malformed stored hash");
        }
    }
    unsigned char want[33];
    long iter = parse_iter(f[1]), slen = parse_hex(f[2], salt, sizeof salt);
    if (strcmp(f[0], "pbkdf2-sha256") != 0 || iter < 0 || slen < 0 || strlen(f[3]) != 64 || parse_hex(f[3], want, 32) != 32)
        return usage("malformed stored hash");
    pbkdf2(pw, pwlen, salt, (size_t)slen, iter, dk);
    unsigned diff = 0;
    for (int i = 0; i < 32; i++) diff |= (unsigned)(dk[i] ^ want[i]);
    free(line);
    free(pw);
    if (diff) {
        fprintf(stderr, "pbkdf2: wrong password\n");
        return 1;
    }
    printf("ok\n");
    return 0;
}
`,
	"html_escape": `#include <stdio.h>
#include <string.h>

static int unreserved(int c) {
    return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~';
}

int main(int argc, char **argv) {
    int html = argc == 2 && strcmp(argv[1], "html") == 0, url = argc == 2 && strcmp(argv[1], "url") == 0;
    if (!html && !url) {
        fprintf(stderr, "usage: html_escape html|url\n");
        return 2;
    }
    int c;
    while ((c = getchar()) != EOF) {
        if (url) {
            if (unreserved(c)) putchar(c);
            else printf("%%%02X", c);
            continue;
        }
        switch (c) {
        case '&': fputs("&amp;", stdout); break;
        case '<': fputs("&lt;", stdout); break;
        case '>': fputs("&gt;", stdout); break;
        case '"': fputs("&quot;", stdout); break;
        case '\'': fputs("&#39;", stdout); break;
        default: putchar(c);
        }
    }
    return 0;
}
`,
}

func TestAppsecChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := appsecSolutions
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

	pb := ref["pbkdf2"]
	// One HMAC instead of the iterated, XORed chain.
	failsWith(t, "pbkdf2", mutate(pb, "    for (long i = 1; i < iter; i++) {", "    for (long i = 1; i < 1 && i < iter; i++) {"), "the standard's test vector", "")
	// Long keys cut to 64 bytes instead of hashed.
	failsWith(t, "pbkdf2", mutate(pb, "    if (klen > 64) {", "    if (klen > 64) klen = 64;\n    if (0) {"), "a password longer than a block", "")
	// Verifying with a fixed iteration count.
	failsWith(t, "pbkdf2", mutate(pb, "    pbkdf2(pw, pwlen, salt, (size_t)slen, iter, dk);\n    unsigned diff = 0;", "    pbkdf2(pw, pwlen, salt, (size_t)slen, iter > 0 ? 20000 : iter, dk);\n    unsigned diff = 0;"), "the stored iteration count", "")
	// Trimming spaces from passwords.
	failsWith(t, "pbkdf2", mutate(pb, "    if (*len > 0 && (*line)[*len - 1] == '\\n') (*line)[--*len] = 0;", "    while (*len > 0 && ((*line)[*len - 1] == '\\n' || (*line)[*len - 1] == ' ')) (*line)[--*len] = 0;"), "spaces in a password", "")
	// No cap on the iteration count.
	failsWith(t, "pbkdf2", mutate(pb, "v < 1 || v > 1000000", "v < 1"), "too many iterations", "")

	he := ref["html_escape"]
	// The single quote left alone.
	failsWith(t, "html_escape", mutate(he, "        case '\\'': fputs(\"&#39;\", stdout); break;\n", ""), "an attribute", "")
	// Bytes printed through a signed char.
	failsWith(t, "html_escape", mutate(he, "            else printf(\"%%%02X\", c);", "            else printf(\"%%%02X\", (unsigned)(signed char)c);"), "high bytes", "")
	// A space as + (form encoding) instead of %20.
	failsWith(t, "html_escape", mutate(he, "            if (unreserved(c)) putchar(c);", "            if (c == ' ') putchar('+');\n            else if (unreserved(c)) putchar(c);"), "a URL value", "")
	// Lower-case hex.
	failsWith(t, "html_escape", mutate(he, "            else printf(\"%%%02X\", c);", "            else printf(\"%%%02x\", c);"), "a URL value", "")
	// Reading into a char: byte 255 looks like EOF.
	failsWith(t, "html_escape", mutate(he, "    int c;\n    while ((c = getchar()) != EOF) {", "    signed char sc;\n    int c;\n    while ((c = sc = (signed char)getchar()) != EOF) {"), "every byte", "")
}
