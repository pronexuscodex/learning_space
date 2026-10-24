package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range identitySolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var identitySolutions = map[string]string{
	"totp": `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define ROTL(x, n) (((x) << (n)) | ((x) >> (32 - (n))))

/* SHA-1 of a message that fits in memory. */
static void sha1(const unsigned char *data, size_t n, unsigned char out[20]) {
    uint32_t h[5] = {0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0};
    size_t total = ((n + 8) / 64 + 1) * 64;
    unsigned char *m = calloc(total, 1);
    if (!m) exit(1);
    if (n) memcpy(m, data, n);
    m[n] = 0x80;
    uint64_t bits = (uint64_t)n * 8;
    for (int i = 0; i < 8; i++) m[total - 1 - i] = (unsigned char)(bits >> (8 * i));
    for (size_t off = 0; off < total; off += 64) {
        uint32_t w[80];
        for (int t = 0; t < 16; t++)
            w[t] = (uint32_t)m[off + 4 * t] << 24 | (uint32_t)m[off + 4 * t + 1] << 16 | (uint32_t)m[off + 4 * t + 2] << 8 | m[off + 4 * t + 3];
        for (int t = 16; t < 80; t++) w[t] = ROTL(w[t - 3] ^ w[t - 8] ^ w[t - 14] ^ w[t - 16], 1);
        uint32_t a = h[0], b = h[1], c = h[2], d = h[3], e = h[4];
        for (int t = 0; t < 80; t++) {
            uint32_t f, k;
            if (t < 20) { f = (b & c) | (~b & d); k = 0x5A827999; }
            else if (t < 40) { f = b ^ c ^ d; k = 0x6ED9EBA1; }
            else if (t < 60) { f = (b & c) | (b & d) | (c & d); k = 0x8F1BBCDC; }
            else { f = b ^ c ^ d; k = 0xCA62C1D6; }
            uint32_t tmp = ROTL(a, 5) + f + e + k + w[t];
            e = d;
            d = c;
            c = ROTL(b, 30);
            b = a;
            a = tmp;
        }
        h[0] += a;
        h[1] += b;
        h[2] += c;
        h[3] += d;
        h[4] += e;
    }
    free(m);
    for (int i = 0; i < 5; i++)
        for (int j = 0; j < 4; j++) out[4 * i + j] = (unsigned char)(h[i] >> (24 - 8 * j));
}

static void hmac_sha1(const unsigned char *key, size_t klen, const unsigned char *msg, size_t mlen, unsigned char out[20]) {
    unsigned char k[64] = {0}, buf[64 + 64], inner[20];
    if (klen > 64) sha1(key, klen, k);
    else memcpy(k, key, klen);
    for (int i = 0; i < 64; i++) buf[i] = k[i] ^ 0x36;
    memcpy(buf + 64, msg, mlen);
    sha1(buf, 64 + mlen, inner);
    for (int i = 0; i < 64; i++) buf[i] = k[i] ^ 0x5c;
    memcpy(buf + 64, inner, 20);
    sha1(buf, 64 + 20, out);
}

/* Base32 to bytes; returns the byte count, or -1 for a bad character. */
static int base32(const char *s, unsigned char *out) {
    size_t len = strlen(s);
    while (len > 0 && s[len - 1] == '=') len--;
    if (len < 16) return -1;
    uint32_t bits = 0;
    int n = 0, count = 0;
    for (size_t i = 0; i < len; i++) {
        int c = s[i], v;
        if (c >= 'A' && c <= 'Z') v = c - 'A';
        else if (c >= 'a' && c <= 'z') v = c - 'a';
        else if (c >= '2' && c <= '7') v = c - '2' + 26;
        else return -1;
        bits = bits << 5 | (uint32_t)v;
        n += 5;
        if (n >= 8) {
            out[count++] = (unsigned char)(bits >> (n - 8));
            n -= 8;
        }
    }
    return count;
}

static unsigned code_at(const unsigned char *key, int klen, uint64_t step) {
    unsigned char msg[8], mac[20];
    for (int i = 0; i < 8; i++) msg[i] = (unsigned char)(step >> (56 - 8 * i));
    hmac_sha1(key, (size_t)klen, msg, 8, mac);
    int o = mac[19] & 15;
    uint32_t v = (uint32_t)(mac[o] & 0x7f) << 24 | (uint32_t)mac[o + 1] << 16 | (uint32_t)mac[o + 2] << 8 | mac[o + 3];
    return v % 1000000;
}

static int all_digits(const char *s, size_t max) {
    size_t n = strlen(s);
    if (n == 0 || n > max) return 0;
    for (size_t i = 0; i < n; i++)
        if (s[i] < '0' || s[i] > '9') return 0;
    return 1;
}

int main(void) {
    char line[512], secret[301], tstr[31], cstr[31];
    unsigned char key[200];
    int lineno = 0;
    while (fgets(line, sizeof line, stdin)) {
        lineno++;
        int got = sscanf(line, "%300s %30s %30s", secret, tstr, cstr);
        if (got <= 0) continue;
        int klen = got >= 2 ? base32(secret, key) : -1;
        if (klen < 0 || !all_digits(tstr, 13)) {
            fprintf(stderr, "totp: line %d: expected SECRET TIME [CODE]\n", lineno);
            return 1;
        }
        unsigned long long t = strtoull(tstr, NULL, 10);
        if (t > 1000000000000ULL) {
            fprintf(stderr, "totp: line %d: time out of range\n", lineno);
            return 1;
        }
        uint64_t step = t / 30;
        if (got == 2) {
            printf("%06u\n", code_at(key, klen, step));
            continue;
        }
        if (strlen(cstr) != 6 || !all_digits(cstr, 6)) {
            fprintf(stderr, "totp: line %d: a code is exactly 6 digits\n", lineno);
            return 1;
        }
        unsigned want = (unsigned)strtoul(cstr, NULL, 10);
        if (code_at(key, klen, step) == want) printf("ok 0\n");
        else if (step > 0 && code_at(key, klen, step - 1) == want) printf("ok -1\n");
        else if (code_at(key, klen, step + 1) == want) printf("ok +1\n");
        else printf("no\n");
    }
    return 0;
}
`,
	"jwt_verify": `#include <stdint.h>
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

static void sha256(const unsigned char *data, size_t n, unsigned char out[32]) {
    uint32_t h[8] = {0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19};
    size_t total = ((n + 8) / 64 + 1) * 64;
    unsigned char *m = calloc(total, 1);
    if (!m) exit(1);
    if (n) memcpy(m, data, n);
    m[n] = 0x80;
    uint64_t bits = (uint64_t)n * 8;
    for (int i = 0; i < 8; i++) m[total - 1 - i] = (unsigned char)(bits >> (8 * i));
    for (size_t off = 0; off < total; off += 64) {
        uint32_t w[64];
        for (int t = 0; t < 16; t++)
            w[t] = (uint32_t)m[off + 4 * t] << 24 | (uint32_t)m[off + 4 * t + 1] << 16 | (uint32_t)m[off + 4 * t + 2] << 8 | m[off + 4 * t + 3];
        for (int t = 16; t < 64; t++) {
            uint32_t s0 = ROTR(w[t - 15], 7) ^ ROTR(w[t - 15], 18) ^ (w[t - 15] >> 3);
            uint32_t s1 = ROTR(w[t - 2], 17) ^ ROTR(w[t - 2], 19) ^ (w[t - 2] >> 10);
            w[t] = s1 + w[t - 7] + s0 + w[t - 16];
        }
        uint32_t a = h[0], b = h[1], c = h[2], d = h[3], e = h[4], f = h[5], g = h[6], hh = h[7];
        for (int t = 0; t < 64; t++) {
            uint32_t t1 = hh + (ROTR(e, 6) ^ ROTR(e, 11) ^ ROTR(e, 25)) + ((e & f) ^ (~e & g)) + K[t] + w[t];
            uint32_t t2 = (ROTR(a, 2) ^ ROTR(a, 13) ^ ROTR(a, 22)) + ((a & b) ^ (a & c) ^ (b & c));
            hh = g;
            g = f;
            f = e;
            e = d + t1;
            d = c;
            c = b;
            b = a;
            a = t1 + t2;
        }
        h[0] += a;
        h[1] += b;
        h[2] += c;
        h[3] += d;
        h[4] += e;
        h[5] += f;
        h[6] += g;
        h[7] += hh;
    }
    free(m);
    for (int i = 0; i < 8; i++)
        for (int j = 0; j < 4; j++) out[4 * i + j] = (unsigned char)(h[i] >> (24 - 8 * j));
}

static void hmac_sha256(const unsigned char *key, size_t klen, const unsigned char *msg, size_t mlen, unsigned char out[32]) {
    unsigned char k[64] = {0}, inner[32];
    if (klen > 64) sha256(key, klen, k);
    else memcpy(k, key, klen);
    unsigned char *buf = malloc(64 + mlen + 32);
    if (!buf) exit(1);
    for (int i = 0; i < 64; i++) buf[i] = k[i] ^ 0x36;
    memcpy(buf + 64, msg, mlen);
    sha256(buf, 64 + mlen, inner);
    for (int i = 0; i < 64; i++) buf[i] = k[i] ^ 0x5c;
    memcpy(buf + 64, inner, 32);
    sha256(buf, 64 + 32, out);
    free(buf);
}

static int b64val(int c) {
    if (c >= 'A' && c <= 'Z') return c - 'A';
    if (c >= 'a' && c <= 'z') return c - 'a' + 26;
    if (c >= '0' && c <= '9') return c - '0' + 52;
    if (c == '-') return 62;
    if (c == '_') return 63;
    return -1;
}

/* base64url without padding; returns the length, or -1. */
static long b64url(const char *s, size_t n, unsigned char *out) {
    if (n % 4 == 1) return -1;
    uint32_t bits = 0;
    int have = 0;
    long len = 0;
    for (size_t i = 0; i < n; i++) {
        int v = b64val((unsigned char)s[i]);
        if (v < 0) return -1;
        bits = bits << 6 | (uint32_t)v;
        have += 6;
        if (have >= 8) {
            out[len++] = (unsigned char)(bits >> (have - 8));
            have -= 8;
        }
    }
    return len;
}

#define MAX_FIELDS 16

struct field { char key[32]; char str[256]; long long num; int is_num; };

/* A compact JSON object of string and number values; -1 if malformed
 * or if a key appears twice. */
static int parse_object(const char *s, size_t n, struct field *f) {
    size_t i = 0;
    int count = 0;
    if (n < 2 || s[0] != '{' || s[n - 1] != '}') return -1;
    i = 1;
    if (i == n - 1) return 0;
    for (;;) {
        if (count == MAX_FIELDS || s[i] != '"') return -1;
        size_t k = ++i;
        while (i < n && s[i] != '"') i++;
        if (i >= n || i - k >= sizeof f[count].key) return -1;
        memcpy(f[count].key, s + k, i - k);
        f[count].key[i - k] = 0;
        for (int j = 0; j < count; j++)
            if (strcmp(f[j].key, f[count].key) == 0) return -1;
        i++;
        if (i >= n || s[i] != ':') return -1;
        i++;
        if (i < n && s[i] == '"') {
            size_t v = ++i;
            while (i < n && s[i] != '"' && s[i] != '\\') i++;
            if (i >= n || s[i] != '"' || i - v >= sizeof f[count].str) return -1;
            memcpy(f[count].str, s + v, i - v);
            f[count].str[i - v] = 0;
            f[count].is_num = 0;
            i++;
        } else {
            size_t v = i;
            long long num = 0;
            while (i < n && s[i] >= '0' && s[i] <= '9') {
                if (num > 100000000000000LL) return -1;
                num = num * 10 + (s[i] - '0');
                i++;
            }
            if (i == v) return -1;
            f[count].num = num;
            f[count].is_num = 1;
        }
        count++;
        if (i == n - 1) return count;
        if (s[i] != ',') return -1;
        i++;
    }
}

static const struct field *get(const struct field *f, int n, const char *key) {
    for (int i = 0; i < n; i++)
        if (strcmp(f[i].key, key) == 0) return &f[i];
    return NULL;
}

static int invalid(const char *why) {
    fprintf(stderr, "invalid: %s\n", why);
    return 1;
}

int main(int argc, char **argv) {
    char *end;
    if (argc != 5) {
        fprintf(stderr, "usage: jwt_verify SECRET ISSUER AUDIENCE NOW\n");
        return 2;
    }
    long long now = strtoll(argv[4], &end, 10);
    if (end == argv[4] || *end) {
        fprintf(stderr, "usage: NOW is a number of seconds\n");
        return 2;
    }
    static char tok[8192];
    if (!fgets(tok, sizeof tok, stdin)) return invalid("no token");
    size_t len = strcspn(tok, "\r\n");
    tok[len] = 0;
    char *d1 = strchr(tok, '.'), *d2 = d1 ? strchr(d1 + 1, '.') : NULL;
    if (!d1 || !d2 || strchr(d2 + 1, '.')) return invalid("a token has exactly three parts");
    static unsigned char header[8192], payload[8192], sig[8192];
    long hl = b64url(tok, (size_t)(d1 - tok), header);
    long pl = b64url(d1 + 1, (size_t)(d2 - d1 - 1), payload);
    long sl = b64url(d2 + 1, strlen(d2 + 1), sig);
    if (hl < 0 || pl < 0 || sl < 0) return invalid("not base64url");
    struct field hf[MAX_FIELDS], pf[MAX_FIELDS];
    int hn = parse_object((const char *)header, (size_t)hl, hf);
    int pn = parse_object((const char *)payload, (size_t)pl, pf);
    if (hn < 0 || pn < 0) return invalid("malformed or duplicated JSON");
    const struct field *alg = get(hf, hn, "alg");
    if (!alg || alg->is_num || strcmp(alg->str, "HS256") != 0) return invalid("alg must be HS256");
    unsigned char mac[32];
    hmac_sha256((const unsigned char *)argv[1], strlen(argv[1]), (const unsigned char *)tok, (size_t)(d2 - tok), mac);
    unsigned diff = sl == 32 ? 0 : 1;
    for (int i = 0; i < 32; i++) diff |= (unsigned)(mac[i] ^ (sl == 32 ? sig[i] : 0));
    if (diff) return invalid("bad signature");
    const struct field *sub = get(pf, pn, "sub"), *iss = get(pf, pn, "iss"), *aud = get(pf, pn, "aud");
    const struct field *exp = get(pf, pn, "exp"), *nbf = get(pf, pn, "nbf");
    if (!sub || sub->is_num || !iss || iss->is_num || !aud || aud->is_num || !exp || !exp->is_num)
        return invalid("sub, iss, aud and exp are required");
    if (strcmp(iss->str, argv[2]) != 0) return invalid("wrong issuer");
    if (strcmp(aud->str, argv[3]) != 0) return invalid("wrong audience");
    if (now >= exp->num) return invalid("expired");
    if (nbf && (!nbf->is_num || now < nbf->num)) return invalid("not yet valid");
    printf("valid sub=%s exp=%lld\n", sub->str, exp->num);
    return 0;
}
`,
	"policy_eval": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct rule { int allow, line; char pat[3][256]; };

/* * matches any run of characters; linear, by remembering the last star. */
static int match(const char *p, const char *s) {
    const char *star = NULL, *mark = s;
    while (*s) {
        if (*p == '*') {
            star = p++;
            mark = s;
        } else if (*p == *s) {
            p++;
            s++;
        } else if (star) {
            p = star + 1;
            s = ++mark;
        } else {
            return 0;
        }
    }
    while (*p == '*') p++;
    return *p == 0;
}

static int matches(const struct rule *r, char req[3][256]) {
    return match(r->pat[0], req[0]) && match(r->pat[1], req[1]) && match(r->pat[2], req[2]);
}

int main(int argc, char **argv) {
    if (argc != 2) {
        fprintf(stderr, "usage: policy_eval POLICY\n");
        return 2;
    }
    FILE *f = fopen(argv[1], "r");
    if (!f) {
        perror(argv[1]);
        return 1;
    }
    struct rule *rules = NULL;
    size_t n = 0, cap = 0;
    char line[1100], effect[16], extra;
    int lineno = 0;
    while (fgets(line, sizeof line, f)) {
        lineno++;
        size_t lead = strspn(line, " \t\r\n");
        if (line[lead] == 0 || line[lead] == '#') continue;
        if (n == cap) {
            cap = cap ? 2 * cap : 64;
            rules = realloc(rules, cap * sizeof *rules);
            if (!rules) return 1;
        }
        struct rule *r = &rules[n];
        int got = sscanf(line, "%15s %255s %255s %255s %c", effect, r->pat[0], r->pat[1], r->pat[2], &extra);
        if (got != 4 || (strcmp(effect, "allow") != 0 && strcmp(effect, "deny") != 0)) {
            fprintf(stderr, "policy_eval: %s line %d: expected allow|deny PRINCIPAL ACTION RESOURCE\n", argv[1], lineno);
            fclose(f);
            free(rules);
            return 1;
        }
        r->allow = strcmp(effect, "allow") == 0;
        r->line = lineno;
        n++;
    }
    fclose(f);
    char req[3][256];
    while (fgets(line, sizeof line, stdin)) {
        if (sscanf(line, "%255s %255s %255s %c", req[0], req[1], req[2], &extra) != 3) {
            fprintf(stderr, "policy_eval: a request is PRINCIPAL ACTION RESOURCE\n");
            free(rules);
            return 1;
        }
        const struct rule *decided = NULL;
        for (size_t i = 0; i < n && !decided; i++)
            if (!rules[i].allow && matches(&rules[i], req)) decided = &rules[i];
        for (size_t i = 0; i < n && !decided; i++)
            if (rules[i].allow && matches(&rules[i], req)) decided = &rules[i];
        if (decided) printf("%s (line %d)\n", decided->allow ? "allow" : "deny", decided->line);
        else printf("deny (no rule)\n");
    }
    free(rules);
    return 0;
}
`,
}

func TestIdentityChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := identitySolutions
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

	totp := ref["totp"]
	// Codes printed without their leading zeros.
	failsWith(t, "totp", mutate(totp, `printf("%06u\n", code_at(key, klen, step));`, `printf("%u\n", code_at(key, klen, step));`), "the standard's test vectors", "")
	// The step sent as 4 bytes: wrong once it passes 2^32.
	failsWith(t, "totp", mutate(totp, "for (int i = 0; i < 8; i++) msg[i] = (unsigned char)(step >> (56 - 8 * i));", "for (int i = 0; i < 8; i++) msg[i] = i < 4 ? 0 : (unsigned char)((uint32_t)step >> (56 - 8 * i));"), "after 2038", "")
	// The top bit not cleared.
	failsWith(t, "totp", mutate(totp, "uint32_t v = (uint32_t)(mac[o] & 0x7f) << 24", "uint32_t v = (uint32_t)mac[o] << 24"), "the top bit", "")
	// No window: only the current step.
	failsWith(t, "totp", mutate(totp, `        else if (step > 0 && code_at(key, klen, step - 1) == want) printf("ok -1\n");
        else if (code_at(key, klen, step + 1) == want) printf("ok +1\n");
`, ""), "the time window", "")
	// A long key cut to 64 bytes instead of hashed.
	failsWith(t, "totp", mutate(totp, "if (klen > 64) sha1(key, klen, k);\n    else memcpy(k, key, klen);", "memcpy(k, key, klen > 64 ? 64 : klen);"), "a long secret", "")
	// Upper case only.
	failsWith(t, "totp", mutate(totp, "        else if (c >= 'a' && c <= 'z') v = c - 'a';\n", ""), "lower case and padding", "")
	// Any character accepted as Base32.
	failsWith(t, "totp", mutate(totp, "        else return -1;\n        bits = bits << 5", "        else v = c & 31;\n        bits = bits << 5"), "not Base32", "")

	jwt := ref["jwt_verify"]
	// The token chooses its own algorithm.
	trustsAlg := mutate(jwt, `    if (!alg || alg->is_num || strcmp(alg->str, "HS256") != 0) return invalid("alg must be HS256");`, `    if (!alg || alg->is_num) return invalid("no alg");`)
	trustsAlg = mutate(trustsAlg, `    if (diff) return invalid("bad signature");`, `    if (diff && strcmp(alg->str, "none") != 0) return invalid("bad signature");`)
	failsWith(t, "jwt_verify", trustsAlg, "alg none", "")
	// Any algorithm name, as long as the HS256 signature checks out.
	failsWith(t, "jwt_verify", mutate(jwt, `strcmp(alg->str, "HS256") != 0`, `strncmp(alg->str, "HS", 2) != 0`), "another algorithm", "")
	// No audience check.
	failsWith(t, "jwt_verify", mutate(jwt, `    if (strcmp(aud->str, argv[3]) != 0) return invalid("wrong audience");
`, ""), "the wrong audience", "")
	// Expired only after exp.
	failsWith(t, "jwt_verify", mutate(jwt, "if (now >= exp->num)", "if (now > exp->num)"), "expired", "")
	// The standard base64 alphabet.
	failsWith(t, "jwt_verify", mutate(mutate(jwt, "if (c == '-') return 62;", "if (c == '+') return 62;"), "if (c == '_') return 63;", "if (c == '/') return 63;"), "every base64url character", "")
	// A long key cut to 64 bytes instead of hashed.
	failsWith(t, "jwt_verify", mutate(jwt, "if (klen > 64) sha256(key, klen, k);\n    else memcpy(k, key, klen);", "memcpy(k, key, klen > 64 ? 64 : klen);"), "a long key", "")
	// Duplicate keys allowed: the last one wins.
	failsWith(t, "jwt_verify", mutate(jwt, `        for (int j = 0; j < count; j++)
            if (strcmp(f[j].key, f[count].key) == 0) return -1;
`, `        for (int j = 0; j < count; j++)
            if (strcmp(f[j].key, f[count].key) == 0) f[j].key[0] = 0;
`), "a key given twice", "")

	pol := ref["policy_eval"]
	// The first matching rule decides, whatever it says.
	failsWith(t, "policy_eval", mutate(pol, `        for (size_t i = 0; i < n && !decided; i++)
            if (!rules[i].allow && matches(&rules[i], req)) decided = &rules[i];
        for (size_t i = 0; i < n && !decided; i++)
            if (rules[i].allow && matches(&rules[i], req)) decided = &rules[i];`, `        for (size_t i = 0; i < n && !decided; i++)
            if (matches(&rules[i], req)) decided = &rules[i];`), "deny wins, wherever it is", "")
	// A star that must match at least one character.
	failsWith(t, "policy_eval", mutate(pol, "    while (*p == '*') p++;\n    return *p == 0;", "    return *p == 0 && p[-1] != '*';"), "wildcards", "")
	// A star that never gives characters back.
	failsWith(t, "policy_eval", mutate(pol, "            p = star + 1;\n            s = ++mark;", "            (void)mark;\n            return 0;"), "wildcards", "")
	// Naive recursive matching.
	naive := mutate(pol, "static int match(const char *p, const char *s) {\n    const char *star = NULL, *mark = s;", `static int match(const char *p, const char *s) {
    if (*p == '*') return match(p + 1, s) || (*s && match(p, s + 1));
    if (!*p) return !*s;
    return *s == *p && match(p + 1, s + 1);
}

static int unused_match(const char *p, const char *s) {
    const char *star = NULL, *mark = s;`)
	naive = mutate(naive, "static int matches(", "int unused_reference(const char *p, const char *s);\nint unused_reference(const char *p, const char *s) { return unused_match(p, s); }\n\nstatic int matches(")
	failsWith(t, "policy_eval", naive, "a pathological pattern", "did not finish")
	// Line numbers counting only rules.
	failsWith(t, "policy_eval", mutate(pol, "        r->line = lineno;", "        r->line = (int)n + 1;"), "comments count as lines", "")
	// A bad rule skipped instead of refused.
	failsWith(t, "policy_eval", mutate(pol, `            fprintf(stderr, "policy_eval: %s line %d: expected allow|deny PRINCIPAL ACTION RESOURCE\n", argv[1], lineno);
            fclose(f);
            free(rules);
            return 1;`, "            continue;"), "a rule that is not allow or deny", "")
}
