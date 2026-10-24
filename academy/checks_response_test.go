package main

import (
	"strings"
	"testing"
	"time"
)

func init() {
	for id, src := range responseSolutions {
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
		referenceSolutions[id] = src
	}
}

var responseSolutions = map[string]string{
	"backup_verify": `#define _POSIX_C_SOURCE 200809L
#include <dirent.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>

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

static int hash_file(const char *path, char hex[65]) {
    FILE *f = fopen(path, "rb");
    if (!f) return 0;
    struct sha s = {{0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19}, {0}, 0, 0};
    static unsigned char buf[65536];
    size_t got;
    while ((got = fread(buf, 1, sizeof buf, f)) > 0) update(&s, buf, got);
    fclose(f);
    uint64_t bits = s.total * 8;
    unsigned char pad = 0x80, zero = 0, len[8];
    update(&s, &pad, 1);
    while (s.used != 56) update(&s, &zero, 1);
    for (int i = 0; i < 8; i++) len[i] = (unsigned char)(bits >> (56 - 8 * i));
    update(&s, len, 8);
    for (int i = 0; i < 8; i++) snprintf(hex + 8 * i, 9, "%08x", (unsigned)s.h[i]);
    return 1;
}

struct entry { char hash[65]; char *path; };

static char **found;
static size_t nfound, capfound;

static void walk(const char *dir, const char *prefix) {
    DIR *d = opendir(dir);
    if (!d) return;
    struct dirent *e;
    while ((e = readdir(d))) {
        if (strcmp(e->d_name, ".") == 0 || strcmp(e->d_name, "..") == 0) continue;
        char full[4096], rel[4096];
        snprintf(full, sizeof full, "%s/%s", dir, e->d_name);
        snprintf(rel, sizeof rel, "%s%s%s", prefix, *prefix ? "/" : "", e->d_name);
        struct stat st;
        if (lstat(full, &st) != 0) continue;
        if (S_ISDIR(st.st_mode)) {
            walk(full, rel);
        } else if (S_ISREG(st.st_mode)) {
            if (nfound == capfound) {
                capfound = capfound ? 2 * capfound : 64;
                found = realloc(found, capfound * sizeof *found);
                if (!found) exit(1);
            }
            found[nfound++] = strdup(rel);
        }
    }
    closedir(d);
}

static int by_string(const void *a, const void *b) { return strcmp(*(char *const *)a, *(char *const *)b); }

static int safe_path(const char *p) {
    if (*p == 0 || *p == '/') return 0;
    while (*p) {
        const char *slash = strchr(p, '/');
        size_t n = slash ? (size_t)(slash - p) : strlen(p);
        if (n == 0 || (n == 1 && p[0] == '.') || (n == 2 && p[0] == '.' && p[1] == '.')) return 0;
        p += n;
        if (*p == '/') {
            p++;
            if (*p == 0) return 0;
        }
    }
    return 1;
}

int main(int argc, char **argv) {
    if (argc != 3) {
        fprintf(stderr, "usage: backup_verify MANIFEST DIR\n");
        return 2;
    }
    FILE *m = fopen(argv[1], "r");
    if (!m) {
        perror(argv[1]);
        return 1;
    }
    struct entry *es = NULL;
    size_t n = 0, cap = 0, lcap = 0;
    char *line = NULL;
    ssize_t got;
    while ((got = getline(&line, &lcap, m)) != -1) {
        size_t len = (size_t)got;
        if (len > 0 && line[len - 1] == '\n') line[--len] = 0;
        int ok = len > 66 && line[64] == ' ' && line[65] == ' ';
        for (int i = 0; ok && i < 64; i++)
            ok = (line[i] >= '0' && line[i] <= '9') || (line[i] >= 'a' && line[i] <= 'f');
        if (!ok || !safe_path(line + 66)) {
            fprintf(stderr, "backup_verify: line %zu of the manifest is not a safe HASH  PATH\n", n + 1);
            return 1;
        }
        if (n == cap) {
            cap = cap ? 2 * cap : 64;
            es = realloc(es, cap * sizeof *es);
            if (!es) return 1;
        }
        memcpy(es[n].hash, line, 64);
        es[n].hash[64] = 0;
        es[n].path = strdup(line + 66);
        n++;
    }
    free(line);
    fclose(m);
    long ok = 0, changed = 0, missing = 0, extra = 0;
    for (size_t i = 0; i < n; i++) {
        char full[4096], hex[65];
        struct stat st;
        snprintf(full, sizeof full, "%s/%s", argv[2], es[i].path);
        if (lstat(full, &st) != 0 || !S_ISREG(st.st_mode) || !hash_file(full, hex)) {
            printf("missing %s\n", es[i].path);
            missing++;
        } else if (strcmp(hex, es[i].hash) != 0) {
            printf("changed %s\n", es[i].path);
            changed++;
        } else {
            ok++;
        }
    }
    walk(argv[2], "");
    char **listed = malloc((n ? n : 1) * sizeof *listed);
    if (!listed) return 1;
    for (size_t i = 0; i < n; i++) listed[i] = es[i].path;
    if (n > 1) qsort(listed, n, sizeof *listed, by_string);
    if (nfound > 1) qsort(found, nfound, sizeof *found, by_string);
    for (size_t i = 0; i < nfound; i++) {
        if (!n || !bsearch(&found[i], listed, n, sizeof *listed, by_string)) {
            printf("extra %s\n", found[i]);
            extra++;
        }
        free(found[i]);
    }
    printf("ok %ld changed %ld missing %ld extra %ld\n", ok, changed, missing, extra);
    for (size_t i = 0; i < n; i++) free(es[i].path);
    free(es);
    free(listed);
    free(found);
    return 0;
}
`,
	"timeline": `#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct event { long long t; long pos; char *host; char *msg; };
struct clock { char *host; long skew; };

static long long days_from_civil(long long y, int m, int d) {
    y -= m <= 2;
    long long era = (y >= 0 ? y : y - 399) / 400;
    long long yoe = y - era * 400;
    long long doy = (153 * (m + (m > 2 ? -3 : 9)) + 2) / 5 + d - 1;
    long long doe = yoe * 365 + yoe / 4 - yoe / 100 + doy;
    return era * 146097 + doe - 719468;
}

static void civil_from_days(long long z, long long *y, int *m, int *d) {
    z += 719468;
    long long era = (z >= 0 ? z : z - 146096) / 146097;
    long long doe = z - era * 146097;
    long long yoe = (doe - doe / 1460 + doe / 36524 - doe / 146096) / 365;
    long long doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    long long mp = (5 * doy + 2) / 153;
    *d = (int)(doy - (153 * mp + 2) / 5 + 1);
    *m = (int)(mp < 10 ? mp + 3 : mp - 9);
    *y = yoe + era * 400 + (*m <= 2);
}

static int leap(long long y) { return (y % 4 == 0 && y % 100 != 0) || y % 400 == 0; }

static int digits(const char *s, int n, int *out) {
    int v = 0;
    for (int i = 0; i < n; i++) {
        if (s[i] < '0' || s[i] > '9') return 0;
        v = v * 10 + (s[i] - '0');
    }
    *out = v;
    return 1;
}

/* Parses an RFC 3339 time into microseconds since 1970 (UTC). */
static int parse_time(const char *s, long long *out) {
    int y, mo, d, h, mi, se, oh = 0, om = 0;
    if (strlen(s) < 20 || !digits(s, 4, &y) || s[4] != '-' || !digits(s + 5, 2, &mo) || s[7] != '-' || !digits(s + 8, 2, &d) ||
        s[10] != 'T' || !digits(s + 11, 2, &h) || s[13] != ':' || !digits(s + 14, 2, &mi) || s[16] != ':' || !digits(s + 17, 2, &se))
        return 0;
    static const int mdays[] = {31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31};
    if (y < 1970 || mo < 1 || mo > 12 || d < 1 || d > mdays[mo - 1] + (mo == 2 && leap(y)) || h > 23 || mi > 59 || se > 59) return 0;
    const char *p = s + 19;
    long long micro = 0;
    if (*p == '.') {
        p++;
        int n = 0;
        long long scale = 100000;
        while (p[n] >= '0' && p[n] <= '9') {
            if (n < 6) micro += (p[n] - '0') * scale;
            scale /= 10;
            n++;
        }
        if (n < 1 || n > 9) return 0;
        p += n;
    }
    int sign = 0;
    if (p[0] == 'Z' && p[1] == 0) {
        sign = 0;
    } else if ((p[0] == '+' || p[0] == '-') && digits(p + 1, 2, &oh) && p[3] == ':' && digits(p + 4, 2, &om) && p[6] == 0) {
        sign = p[0] == '+' ? 1 : -1;
        if (oh > 14 || om > 59 || oh * 60 + om > 14 * 60) return 0;
    } else {
        return 0;
    }
    long long secs = days_from_civil(y, mo, d) * 86400 + h * 3600 + mi * 60 + se - sign * (oh * 3600 + om * 60);
    *out = secs * 1000000 + micro;
    return 1;
}

static int by_time(const void *a, const void *b) {
    const struct event *x = a, *y = b;
    if (x->t != y->t) return x->t < y->t ? -1 : 1;
    return x->pos < y->pos ? -1 : x->pos > y->pos;
}

static int fail(long n) {
    fprintf(stderr, "timeline: line %ld is malformed\n", n);
    return 1;
}

int main(void) {
    struct event *ev = NULL;
    struct clock *cl = NULL;
    size_t ne = 0, ce = 0, nc = 0, cc = 0, cap = 0;
    char *line = NULL;
    ssize_t got;
    long n = 0;
    while ((got = getline(&line, &cap, stdin)) != -1) {
        n++;
        if (got > 0 && line[got - 1] == '\n') line[got - 1] = 0;
        char *sp1 = strchr(line, ' ');
        if (!sp1) return fail(n);
        *sp1 = 0;
        char *rest = sp1 + 1, *sp2 = strchr(rest, ' ');
        if (strcmp(line, "clock") == 0) {
            char *end;
            if (!sp2) return fail(n);
            *sp2 = 0;
            long skew = strtol(sp2 + 1, &end, 10);
            if (end == sp2 + 1 || *end) return fail(n);
            for (size_t i = 0; i < nc; i++)
                if (strcmp(cl[i].host, rest) == 0) return fail(n);
            if (nc == cc) {
                cc = cc ? 2 * cc : 8;
                cl = realloc(cl, cc * sizeof *cl);
                if (!cl) return 1;
            }
            cl[nc].host = strdup(rest);
            cl[nc++].skew = skew;
            continue;
        }
        if (sp2) *sp2 = 0;
        long long t;
        if (!parse_time(rest, &t)) return fail(n);
        if (ne == ce) {
            ce = ce ? 2 * ce : 1024;
            ev = realloc(ev, ce * sizeof *ev);
            if (!ev) return 1;
        }
        ev[ne].t = t;
        ev[ne].pos = n;
        ev[ne].host = strdup(line);
        ev[ne].msg = strdup(sp2 ? sp2 + 1 : "");
        ne++;
    }
    free(line);
    for (size_t i = 0; i < ne; i++)
        for (size_t j = 0; j < nc; j++)
            if (strcmp(ev[i].host, cl[j].host) == 0) ev[i].t -= (long long)cl[j].skew * 1000000;
    if (ne > 1) qsort(ev, ne, sizeof *ev, by_time);
    for (size_t i = 0; i < ne; i++) {
        long long t = ev[i].t, secs = t / 1000000, micro = t % 1000000, days = secs / 86400, rem = secs % 86400, y;
        int m, d;
        civil_from_days(days, &y, &m, &d);
        printf("%04lld-%02d-%02dT%02lld:%02lld:%02lld.%06lldZ %s %s\n", y, m, d, rem / 3600, rem / 60 % 60, rem % 60, micro, ev[i].host, ev[i].msg);
        free(ev[i].host);
        free(ev[i].msg);
    }
    for (size_t j = 0; j < nc; j++) free(cl[j].host);
    free(ev);
    free(cl);
    return 0;
}
`,
	"k_anon": `#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct spec { int col; int prefix; long width; };
struct group { long size; char *key; };

static int split(char *line, char **f, int max) {
    int n = 0;
    for (char *p = line;; ) {
        if (n == max) return -1;
        f[n++] = p;
        char *c = strchr(p, ',');
        if (!c) break;
        *c = 0;
        p = c + 1;
    }
    return n;
}

static int by_string(const void *a, const void *b) { return strcmp(*(char *const *)a, *(char *const *)b); }

static int by_group(const void *a, const void *b) {
    const struct group *x = a, *y = b;
    if (x->size != y->size) return x->size < y->size ? -1 : 1;
    return strcmp(x->key, y->key);
}

static int bad(const char *why) {
    fprintf(stderr, "k_anon: %s\n", why);
    return 1;
}

int main(int argc, char **argv) {
    char *end;
    if (argc < 3) {
        fprintf(stderr, "usage: k_anon K COLUMN[:N|/W]...\n");
        return 2;
    }
    long k = strtol(argv[1], &end, 10);
    if (end == argv[1] || *end || k < 1) return bad("K must be a whole number of 1 or more");
    char *line = NULL, *fields[256], *header[256];
    size_t cap = 0;
    ssize_t got = getline(&line, &cap, stdin);
    if (got <= 0) return bad("no header line");
    if (line[got - 1] == '\n') line[got - 1] = 0;
    char *hline = strdup(line);
    int ncols = split(hline, header, 256);
    if (ncols < 0) return bad("too many columns");
    int nspec = argc - 2;
    struct spec *sp = calloc((size_t)nspec, sizeof *sp);
    if (!sp) return 1;
    for (int i = 0; i < nspec; i++) {
        char name[256];
        snprintf(name, sizeof name, "%s", argv[i + 2]);
        char *colon = strchr(name, ':'), *slash = strchr(name, '/');
        if (colon || slash) {
            char *num = (colon ? colon : slash) + 1;
            long v = strtol(num, &end, 10);
            if (end == num || *end || v < 1) return bad("a SPEC's N or W must be a whole number of 1 or more");
            if (colon) sp[i].prefix = (int)v;
            else sp[i].width = v;
            *(colon ? colon : slash) = 0;
        }
        sp[i].col = -1;
        for (int c = 0; c < ncols; c++)
            if (strcmp(header[c], name) == 0) sp[i].col = c;
        if (sp[i].col < 0) return bad("unknown column");
    }
    char **keys = NULL;
    size_t nrows = 0, kcap = 0;
    while ((got = getline(&line, &cap, stdin)) != -1) {
        if (got > 0 && line[got - 1] == '\n') line[got - 1] = 0;
        if (split(line, fields, 256) != ncols) return bad("a row has the wrong number of fields");
        char key[4096];
        size_t kl = 0;
        for (int i = 0; i < nspec; i++) {
            const char *v = fields[sp[i].col];
            char part[512];
            if (sp[i].width) {
                char *e;
                long long x = strtoll(v, &e, 10);
                if (e == v || *e || x < 0 || *v == '-' || *v == '+') return bad("a value is not a whole number");
                long long lo = x / sp[i].width * sp[i].width;
                snprintf(part, sizeof part, "%lld-%lld", lo, lo + sp[i].width - 1);
            } else if (sp[i].prefix) {
                snprintf(part, sizeof part, "%.*s", sp[i].prefix, v);
            } else {
                snprintf(part, sizeof part, "%s", v);
            }
            kl += (size_t)snprintf(key + kl, sizeof key - kl, "%s%s", i ? "," : "", part);
            if (kl >= sizeof key) return bad("a row is too long");
        }
        if (nrows == kcap) {
            kcap = kcap ? 2 * kcap : 1024;
            keys = realloc(keys, kcap * sizeof *keys);
            if (!keys) return 1;
        }
        keys[nrows++] = strdup(key);
    }
    free(line);
    if (nrows > 1) qsort(keys, nrows, sizeof *keys, by_string);
    struct group *small = NULL;
    size_t nsmall = 0, groups = 0;
    long smallest = 0, at_risk = 0;
    for (size_t i = 0; i < nrows;) {
        size_t j = i;
        while (j < nrows && strcmp(keys[j], keys[i]) == 0) j++;
        long size = (long)(j - i);
        groups++;
        if (groups == 1 || size < smallest) smallest = size;
        if (size < k) {
            small = realloc(small, (nsmall + 1) * sizeof *small);
            if (!small) return 1;
            small[nsmall].size = size;
            small[nsmall++].key = keys[i];
            at_risk += size;
        }
        i = j;
    }
    if (nsmall > 1) qsort(small, nsmall, sizeof *small, by_group);
    printf("rows %zu\ngroups %zu\nsmallest %ld\nat risk %ld\n", nrows, groups, smallest, at_risk);
    for (size_t i = 0; i < nsmall; i++) printf("%ld %s\n", small[i].size, small[i].key);
    for (size_t i = 0; i < nrows; i++) free(keys[i]);
    free(keys);
    free(small);
    free(sp);
    free(hline);
    return 0;
}
`,
}

func TestResponseChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := responseSolutions
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

	bv := ref["backup_verify"]
	// Manifest paths trusted: .. reads outside the folder.
	failsWith(t, "backup_verify", mutate(bv, "if (!ok || !safe_path(line + 66)) {", "if (!ok || !(safe_path(line + 66) || 1)) {"), "a path that leaves the folder", "")
	// Only checking that each file exists.
	failsWith(t, "backup_verify", mutate(bv, "        } else if (strcmp(hex, es[i].hash) != 0) {", "        } else if (0 && strcmp(hex, es[i].hash) != 0) {"), "a file emptied", "")
	// Files read as text: hashing stops at the first zero byte.
	failsWith(t, "backup_verify", mutate(bv, "    while ((got = fread(buf, 1, sizeof buf, f)) > 0) update(&s, buf, got);", "    while ((got = fread(buf, 1, sizeof buf, f)) > 0) update(&s, buf, strnlen((char *)buf, got));"), "binary files", "")
	// Not looking inside subfolders.
	failsWith(t, "backup_verify", mutate(bv, "            walk(full, rel);", "            (void)0;"), "extra files in order", "")

	tl := ref["timeline"]
	// The offset added instead of subtracted.
	failsWith(t, "timeline", mutate(tl, "- sign * (oh * 3600 + om * 60);", "+ sign * (oh * 3600 + om * 60);"), "offsets both ways", "")
	// A fraction read as a plain number: .5 becomes 5 microseconds.
	failsWith(t, "timeline", mutate(tl, "            if (n < 6) micro += (p[n] - '0') * scale;", "            if (n < 6) micro = micro * 10 + (p[n] - '0');\n            (void)scale;"), "fractions", "")
	// Every year divisible by 4 a leap year.
	failsWith(t, "timeline", mutate(tl, "static int leap(long long y) { return (y % 4 == 0 && y % 100 != 0) || y % 400 == 0; }", "static int leap(long long y) { return y % 4 == 0; }"), "not a leap year", "")
	// The clock error added instead of subtracted.
	failsWith(t, "timeline", mutate(tl, "ev[i].t -= (long long)cl[j].skew * 1000000;", "ev[i].t += (long long)cl[j].skew * 1000000;"), "clock errors", "")
	// February 30 accepted.
	failsWith(t, "timeline", mutate(tl, "d > mdays[mo - 1] + (mo == 2 && leap(y))", "d > mdays[mo - 1] + (mo == 2 ? 2 : 0) + 0 * leap(y)"), "February 30", "")

	ka := ref["k_anon"]
	// Keys without separators: ab+c and a+bc collide.
	failsWith(t, "k_anon", mutate(ka, `"%s%s", i ? "," : "", part);`, `"%s%s", i ? "" : "", part);`), "keys that run together", "")
	// The band's top one too high.
	failsWith(t, "k_anon", mutate(ka, "lo, lo + sp[i].width - 1);", "lo, lo + sp[i].width);"), "bands", "")
	// Counting groups at risk instead of rows.
	failsWith(t, "k_anon", mutate(ka, "            at_risk += size;", "            at_risk += 1;"), "generalised", "")
	// Short rows accepted.
	failsWith(t, "k_anon", mutate(ka, "if (split(line, fields, 256) != ncols) return bad(", "if (split(line, fields, 256) > ncols) return bad("), "a row with too few fields", "")
}
