package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range monitorSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var monitorSolutions = map[string]string{
	"log_chain": `#define _POSIX_C_SOURCE 200809L
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

static void sha256_hex(const unsigned char *data, size_t n, char hex[65]) {
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
    for (int i = 0; i < 8; i++) snprintf(hex + 8 * i, 9, "%08x", (unsigned)h[i]);
}

/* next = SHA-256(prev + "\n" + msg), in hex. */
static void chain_link(const char prev[65], const char *msg, size_t len, char next[65]) {
    unsigned char *buf = malloc(65 + len);
    if (!buf) exit(1);
    memcpy(buf, prev, 64);
    buf[64] = '\n';
    if (len) memcpy(buf + 65, msg, len);
    sha256_hex(buf, 65 + len, next);
    free(buf);
}

static int is_hex64(const char *s, size_t len) {
    if (len < 65 || s[64] != ' ') return 0;
    for (int i = 0; i < 64; i++)
        if (!((s[i] >= '0' && s[i] <= '9') || (s[i] >= 'a' && s[i] <= 'f'))) return 0;
    return 1;
}

int main(int argc, char **argv) {
    int seal = argc == 2 && strcmp(argv[1], "seal") == 0;
    int verify = (argc == 2 || argc == 3) && strcmp(argv[1], "verify") == 0;
    if (!seal && !verify) {
        fprintf(stderr, "usage: log_chain seal | log_chain verify [NEWEST]\n");
        return 2;
    }
    char prev[65], next[65];
    memset(prev, '0', 64);
    prev[64] = 0;
    char *line = NULL;
    size_t cap = 0;
    ssize_t got;
    long n = 0;
    while ((got = getline(&line, &cap, stdin)) != -1) {
        size_t len = (size_t)got;
        if (len > 0 && line[len - 1] == '\n') len--;
        n++;
        if (seal) {
            chain_link(prev, line, len, next);
            printf("%s ", next);
            fwrite(line, 1, len, stdout);
            putchar('\n');
        } else {
            if (!is_hex64(line, len)) {
                fprintf(stderr, "log_chain: line %ld: not HASH MESSAGE\n", n);
                free(line);
                return 1;
            }
            chain_link(prev, line + 65, len - 65, next);
            if (memcmp(next, line, 64) != 0) {
                fprintf(stderr, "log_chain: line %ld: the hash does not match\n", n);
                free(line);
                return 1;
            }
        }
        memcpy(prev, next, 65);
    }
    free(line);
    if (verify) {
        if (argc == 3 && (n == 0 || strcmp(prev, argv[2]) != 0)) {
            fprintf(stderr, "log_chain: the newest hash is not %s: lines are missing at the end\n", argv[2]);
            return 1;
        }
        printf("ok %ld lines\n", n);
    }
    return 0;
}
`,
	"multi_match": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MAX_PATTERNS 5000
#define MAX_BYTES 40000

static int (*next)[256];
static int *fail, *order;
static long *hits;
static int nstates = 1;

int main(int argc, char **argv) {
    if (argc != 2) {
        fprintf(stderr, "usage: multi_match PATTERNS\n");
        return 2;
    }
    FILE *f = fopen(argv[1], "r");
    if (!f) {
        perror(argv[1]);
        return 1;
    }
    next = malloc((MAX_BYTES + 1) * sizeof *next);
    fail = calloc(MAX_BYTES + 1, sizeof *fail);
    order = malloc((MAX_BYTES + 1) * sizeof *order);
    hits = calloc(MAX_BYTES + 1, sizeof *hits);
    static char pats[MAX_PATTERNS + 1][66];
    static int end[MAX_PATTERNS + 1];
    if (!next || !fail || !order || !hits) return 1;
    memset(next[0], -1, sizeof next[0]);
    int np = 0, bytes = 0;
    while (fgets(pats[np], sizeof pats[np], f)) {
        size_t len = strcspn(pats[np], "\n");
        if (pats[np][len] != '\n' && !feof(f)) {
            fprintf(stderr, "multi_match: pattern %d is longer than 64 bytes\n", np + 1);
            return 1;
        }
        pats[np][len] = 0;
        if (len == 0 || len > 64 || np == MAX_PATTERNS || bytes + (int)len > MAX_BYTES) {
            fprintf(stderr, "multi_match: pattern %d is empty or over a limit\n", np + 1);
            return 1;
        }
        bytes += (int)len;
        int v = 0;
        for (size_t k = 0; k < len; k++) {
            unsigned char c = (unsigned char)pats[np][k];
            if (next[v][c] < 0) {
                memset(next[nstates], -1, sizeof next[0]);
                next[v][c] = nstates++;
            }
            v = next[v][c];
        }
        end[np++] = v;
    }
    fclose(f);
    /* Breadth first: fail links, and the missing moves filled in. */
    int head = 0, tail = 0;
    for (int c = 0; c < 256; c++) {
        if (next[0][c] < 0) next[0][c] = 0;
        else order[tail++] = next[0][c];
    }
    while (head < tail) {
        int v = order[head++];
        for (int c = 0; c < 256; c++) {
            int u = next[v][c];
            if (u < 0) {
                next[v][c] = next[fail[v]][c];
            } else {
                fail[u] = next[fail[v]][c];
                order[tail++] = u;
            }
        }
    }
    /* One pass over the text, counting visits to each state. */
    unsigned char buf[1 << 16];
    size_t got;
    int v = 0;
    while ((got = fread(buf, 1, sizeof buf, stdin)) > 0)
        for (size_t i = 0; i < got; i++) {
            v = next[v][buf[i]];
            hits[v]++;
        }
    /* A visit to a state is also a visit to its fail state: push counts
     * down the fail links, deepest states first. */
    for (int i = tail - 1; i >= 0; i--) hits[fail[order[i]]] += hits[order[i]];
    long total = 0;
    for (int i = 0; i < np; i++) {
        printf("%ld %s\n", hits[end[i]], pats[i]);
        total += hits[end[i]];
    }
    printf("total %ld\n", total);
    free(next);
    free(fail);
    free(order);
    free(hits);
    return 0;
}
`,
	"auth_detect": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define WINDOW 600
#define TABLE 8192

struct user { char name[64]; long *fails; int nfails, cap; long last_alert; int reported; };
struct addr { char name[64]; int *users; long *last_fail; int n, cap; long last_alert; };

static struct user users[TABLE];
static struct addr addrs[TABLE];
static int user_used[TABLE], addr_used[TABLE];

static unsigned hash(const char *s) {
    unsigned h = 2166136261u;
    while (*s) h = (h ^ (unsigned char)*s++) * 16777619u;
    return h;
}

static int find_user(const char *name) {
    for (unsigned i = hash(name) % TABLE;; i = (i + 1) % TABLE) {
        if (!user_used[i]) {
            user_used[i] = 1;
            strcpy(users[i].name, name);
            users[i].last_alert = -1;
            return (int)i;
        }
        if (strcmp(users[i].name, name) == 0) return (int)i;
    }
}

static int find_addr(const char *name) {
    for (unsigned i = hash(name) % TABLE;; i = (i + 1) % TABLE) {
        if (!addr_used[i]) {
            addr_used[i] = 1;
            strcpy(addrs[i].name, name);
            addrs[i].last_alert = -1;
            return (int)i;
        }
        if (strcmp(addrs[i].name, name) == 0) return (int)i;
    }
}

int main(void) {
    char line[256], a[64], u[64], r[8], extra;
    long t, prev = -1, alerts = 0;
    int lineno = 0;
    while (fgets(line, sizeof line, stdin)) {
        lineno++;
        if (sscanf(line, "%ld %63s %63s %7s %c", &t, a, u, r, &extra) != 4 || (strcmp(r, "ok") != 0 && strcmp(r, "fail") != 0) || t < 0) {
            fprintf(stderr, "auth_detect: line %d: expected TIME ADDRESS USER ok|fail\n", lineno);
            return 1;
        }
        if (t < prev) {
            fprintf(stderr, "auth_detect: line %d: time goes backwards\n", lineno);
            return 1;
        }
        prev = t;
        struct user *us = &users[find_user(u)];
        struct addr *ad = &addrs[find_addr(a)];
        if (strcmp(r, "ok") == 0) {
            if (us->last_alert >= 0 && !us->reported && t - us->last_alert <= WINDOW) {
                printf("%ld success-after-alert %s %s\n", t, u, a);
                us->reported = 1;
                alerts++;
            }
            continue;
        }
        /* The user's failures in the window. */
        if (us->nfails == us->cap) {
            us->cap = us->cap ? 2 * us->cap : 8;
            us->fails = realloc(us->fails, us->cap * sizeof *us->fails);
            if (!us->fails) return 1;
        }
        us->fails[us->nfails++] = t;
        int keep = 0;
        for (int i = 0; i < us->nfails; i++)
            if (us->fails[i] > t - WINDOW) us->fails[keep++] = us->fails[i];
        us->nfails = keep;
        if (us->nfails >= 5 && (us->last_alert < 0 || t >= us->last_alert + WINDOW)) {
            printf("%ld single-account %s\n", t, u);
            us->last_alert = t;
            us->reported = 0;
            alerts++;
        }
        /* Different users failing from this address in the window. */
        int ui = (int)(us - users), found = 0, distinct = 0;
        for (int i = 0; i < ad->n; i++)
            if (ad->users[i] == ui) {
                ad->last_fail[i] = t;
                found = 1;
            }
        if (!found) {
            if (ad->n == ad->cap) {
                ad->cap = ad->cap ? 2 * ad->cap : 8;
                ad->users = realloc(ad->users, ad->cap * sizeof *ad->users);
                ad->last_fail = realloc(ad->last_fail, ad->cap * sizeof *ad->last_fail);
                if (!ad->users || !ad->last_fail) return 1;
            }
            ad->users[ad->n] = ui;
            ad->last_fail[ad->n++] = t;
        }
        for (int i = 0; i < ad->n; i++)
            if (ad->last_fail[i] > t - WINDOW) distinct++;
        if (distinct >= 10 && (ad->last_alert < 0 || t >= ad->last_alert + WINDOW)) {
            printf("%ld many-accounts %s\n", t, a);
            ad->last_alert = t;
            alerts++;
        }
    }
    printf("alerts %ld\n", alerts);
    for (int i = 0; i < TABLE; i++) {
        free(users[i].fails);
        free(addrs[i].users);
        free(addrs[i].last_fail);
    }
    return 0;
}
`,
}

func TestMonitorChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := monitorSolutions
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

	chain := ref["log_chain"]
	// Each message hashed alone: no chain at all.
	alone := mutate(chain, "    memcpy(buf, prev, 64);\n    buf[64] = '\\n';\n    if (len) memcpy(buf + 65, msg, len);\n    sha256_hex(buf, 65 + len, next);",
		"    (void)prev;\n    if (len) memcpy(buf, msg, len);\n    sha256_hex(buf, len, next);")
	failsWith(t, "log_chain", alone, "the starter's example", "")
	// Chaining that skips the newline between hash and message.
	failsWith(t, "log_chain", mutate(chain, "    sha256_hex(buf, 65 + len, next);", "    memmove(buf + 64, buf + 65, len);\n    sha256_hex(buf, 64 + len, next);"), "the starter's example", "")
	// NEWEST ignored.
	failsWith(t, "log_chain", mutate(chain, "if (argc == 3 && (n == 0 || strcmp(prev, argv[2]) != 0)) {", "if (argc == 4) {"), "the end cut off", "")
	// Hex accepted and compared in either case.
	anyCase := mutate(chain, "(s[i] >= 'a' && s[i] <= 'f')", "(s[i] >= 'a' && s[i] <= 'f') || (s[i] >= 'A' && s[i] <= 'F')")
	anyCase = mutate(anyCase, "if (memcmp(next, line, 64) != 0) {", "if (strncasecmp(next, line, 64) != 0) {")
	anyCase = mutate(anyCase, "#include <string.h>\n", "#include <string.h>\n#include <strings.h>\n")
	failsWith(t, "log_chain", anyCase, "upper-case hex", "")
	// A fixed line buffer.
	fixed := mutate(chain, "    while ((got = getline(&line, &cap, stdin)) != -1) {", "    static char fixedbuf[4096];\n    (void)cap;\n    while (fgets(fixedbuf, sizeof fixedbuf, stdin) && (line = fixedbuf, got = (ssize_t)strlen(fixedbuf), 1)) {")
	fixed = strings.ReplaceAll(fixed, "                free(line);\n", "")
	fixed = mutate(fixed, "    free(line);\n    if (verify) {", "    if (verify) {")
	failsWith(t, "log_chain", fixed, "a very long message", "")

	mm := ref["multi_match"]
	// No counts through fail links: shorter signatures inside longer ones are missed.
	failsWith(t, "multi_match", mutate(mm, "    for (int i = tail - 1; i >= 0; i--) hits[fail[order[i]]] += hits[order[i]];\n", ""), "one inside another", "")
	// Back to the root after every match: overlaps are missed.
	failsWith(t, "multi_match", mutate(mm, "            v = next[v][buf[i]];\n            hits[v]++;", "            v = next[v][buf[i]];\n            hits[v]++;\n            if (v == end[0]) v = 0;"), "overlaps", "")
	// Text read as a C string: everything after a zero byte is lost.
	failsWith(t, "multi_match", mutate(mm, "        for (size_t i = 0; i < got; i++) {", "        for (size_t i = 0; i < got && buf[i]; i++) {"), "binary data", "")
	// Blank signatures accepted.
	failsWith(t, "multi_match", mutate(mm, "if (len == 0 || len > 64 ||", "if (len > 64 ||"), "a blank signature", "")
	// Each signature searched separately, with strstr: one pass per pattern.
	naive := mutate(mm, "    long total = 0;\n    for (int i = 0; i < np; i++) {\n        printf(\"%ld %s\\n\", hits[end[i]], pats[i]);\n        total += hits[end[i]];\n    }",
		`    long total = 0;
    rewind(stdin);
    static char text[1 << 25];
    size_t tl = fread(text, 1, sizeof text - 1, stdin);
    text[tl] = 0;
    for (int i = 0; i < np; i++) {
        long c = 0;
        for (const char *p = text; (p = strstr(p, pats[i])); p++) c++;
        printf("%ld %s\n", c, pats[i]);
        total += c;
    }`)
	naive = mutate(naive, "    while ((got = fread(buf, 1, sizeof buf, stdin)) > 0)", "    (void)v;\n    (void)end;\n    while (0 && (got = fread(buf, 1, sizeof buf, stdin)) > 0)")
	failsWith(t, "multi_match", naive, "five thousand signatures", "did not finish")

	auth := ref["auth_detect"]
	// A window that includes its edge.
	failsWith(t, "auth_detect", mutate(auth, "if (us->fails[i] > t - WINDOW) us->fails[keep++]", "if (us->fails[i] >= t - WINDOW) us->fails[keep++]"), "the window's edge", "")
	// Counting failures from an address, not different users.
	failsWith(t, "auth_detect", mutate(auth, "            if (ad->last_fail[i] > t - WINDOW) distinct++;", "            if (ad->last_fail[i] > t - WINDOW) distinct += us->nfails;"), "many failures, one account", "")
	// No quiet period: an alert at every failure after the fifth.
	failsWith(t, "auth_detect", mutate(auth, "if (us->nfails >= 5 && (us->last_alert < 0 || t >= us->last_alert + WINDOW)) {", "if (us->nfails >= 5) {"), "many failures, one account", "")
	// The success rule missing.
	failsWith(t, "auth_detect", mutate(auth, "if (us->last_alert >= 0 && !us->reported && t - us->last_alert <= WINDOW) {", "if (0) {"), "success after an alert", "")
	// Reported at every success, not once per alert.
	failsWith(t, "auth_detect", mutate(auth, "                us->reported = 1;\n", ""), "success after an alert", "")
	// Out-of-order times accepted.
	failsWith(t, "auth_detect", mutate(auth, "        if (t < prev) {", "        if (0 && t < prev) {"), "time going backwards", "")
}
