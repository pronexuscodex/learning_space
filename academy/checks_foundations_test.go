package main

import (
	"strings"
	"testing"
	"time"
)

// Reference solutions for the Foundations checks (Stages 2 to 4); they
// join referenceSolutions, so TestReferenceSolutionsPass runs them.
func init() {
	for id, src := range foundationsSolutions {
		referenceSolutions[id] = src
	}
}

var foundationsSolutions = map[string]string{
	"duplicates": `#include <stdio.h>
#include <stdlib.h>

struct item { int value; int pos; };

static int by_value_then_pos(const void *a, const void *b) {
    const struct item *x = a, *y = b;
    if (x->value != y->value) return (x->value > y->value) - (x->value < y->value);
    return (x->pos > y->pos) - (x->pos < y->pos);
}

int main(void) {
    int n;
    if (scanf("%d", &n) != 1 || n < 0) return 1;
    struct item *a = malloc((size_t)(n > 0 ? n : 1) * sizeof *a);
    if (a == NULL) return 1;
    for (int i = 0; i < n; i++) {
        if (scanf("%d", &a[i].value) != 1) return 1;
        a[i].pos = i;
    }
    qsort(a, (size_t)n, sizeof *a, by_value_then_pos);
    int best = -1, value = 0;
    for (int i = 1; i < n; i++) {
        if (a[i].value == a[i - 1].value && (i < 2 || a[i - 2].value != a[i].value)) {
            if (best < 0 || a[i].pos < best) {
                best = a[i].pos;
                value = a[i].value;
            }
        }
    }
    if (best < 0) printf("none\n");
    else printf("%d\n", value);
    free(a);
    return 0;
}
`,
	"hanoi": `#include <stdio.h>

static void hanoi(int n, char from, char to, char via) {
    if (n == 0) return;
    hanoi(n - 1, from, via, to);
    printf("%d %c %c\n", n, from, to);
    hanoi(n - 1, via, to, from);
}

int main(void) {
    int n;
    if (scanf("%d", &n) != 1 || n < 0 || n > 20) return 1;
    hanoi(n, 'A', 'C', 'B');
    return 0;
}
`,
	"bsearch": `#include <stdio.h>
#include <stdlib.h>

/* first index whose value is >= x, or n */
static int lower_bound(const int *a, int n, int x) {
    int low = 0, high = n;
    while (low < high) {
        int mid = low + (high - low) / 2;
        if (a[mid] < x) low = mid + 1;
        else high = mid;
    }
    return low;
}

int main(void) {
    int n, q;
    if (scanf("%d", &n) != 1 || n < 0) return 1;
    int *a = malloc((size_t)(n > 0 ? n : 1) * sizeof *a);
    if (a == NULL) return 1;
    for (int i = 0; i < n; i++)
        if (scanf("%d", &a[i]) != 1) return 1;
    if (scanf("%d", &q) != 1) return 1;
    while (q-- > 0) {
        int x;
        if (scanf("%d", &x) != 1) return 1;
        int i = lower_bound(a, n, x);
        printf("%d\n", i < n && a[i] == x ? i : -1);
    }
    free(a);
    return 0;
}
`,
	"intervals": `#include <stdio.h>
#include <stdlib.h>

struct meeting { int start, end; };

static int by_end(const void *a, const void *b) {
    const struct meeting *x = a, *y = b;
    return (x->end > y->end) - (x->end < y->end);
}

int main(void) {
    int n;
    if (scanf("%d", &n) != 1 || n < 0) return 1;
    struct meeting *m = malloc((size_t)(n > 0 ? n : 1) * sizeof *m);
    if (m == NULL) return 1;
    for (int i = 0; i < n; i++)
        if (scanf("%d %d", &m[i].start, &m[i].end) != 2) return 1;
    qsort(m, (size_t)n, sizeof *m, by_end);
    int count = 0, last = -1;
    for (int i = 0; i < n; i++) {
        if (m[i].start >= last) {
            count++;
            last = m[i].end;
        }
    }
    printf("%d\n", count);
    free(m);
    return 0;
}
`,
	"edit_distance": `#include <stdio.h>
#include <string.h>

static char a[2100], b[2100];
static int prev[2100], cur[2100];

static int min3(int x, int y, int z) {
    int m = x < y ? x : y;
    return m < z ? m : z;
}

int main(void) {
    while (fgets(a, sizeof a, stdin) != NULL && fgets(b, sizeof b, stdin) != NULL) {
        a[strcspn(a, "\n")] = '\0';
        b[strcspn(b, "\n")] = '\0';
        int n = (int)strlen(a), m = (int)strlen(b);
        for (int j = 0; j <= m; j++) prev[j] = j;
        for (int i = 1; i <= n; i++) {
            cur[0] = i;
            for (int j = 1; j <= m; j++)
                cur[j] = min3(prev[j] + 1, cur[j - 1] + 1, prev[j - 1] + (a[i - 1] != b[j - 1]));
            memcpy(prev, cur, (size_t)(m + 1) * sizeof cur[0]);
        }
        printf("%d\n", prev[m]);
    }
    return 0;
}
`,
	"sat_count": `#include <stdio.h>

static int lits[4096];

int main(void) {
    int n, m, len = 0;
    if (scanf("%d %d", &n, &m) != 2) return 1;
    for (int c = 0; c < m; c++) {
        int l;
        do {
            if (scanf("%d", &l) != 1 || len >= 4096) return 1;
            lits[len++] = l;
        } while (l != 0);
    }
    long count = 0;
    for (long row = 0; row < (1L << n); row++) {
        int all = 1, any = 0;
        for (int i = 0; i < len && all; i++) {
            int l = lits[i];
            if (l == 0) {
                all = any;
                any = 0;
            } else {
                int v = l > 0 ? l : -l;
                int bit = (int)(row >> (v - 1) & 1);
                if (bit == (l > 0)) any = 1;
            }
        }
        if (all) count++;
    }
    printf("%ld\n", count);
    return 0;
}
`,
	"bitsets": `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static uint64_t read_set(void) {
    char line[1024];
    uint64_t s = 0;
    if (fgets(line, sizeof line, stdin) == NULL) return 0;
    char *p = line, *end;
    for (;;) {
        long v = strtol(p, &end, 10);
        if (end == p) break;
        if (v >= 0 && v < 64) s |= (uint64_t)1 << v;
        p = end;
    }
    return s;
}

static void print_set(uint64_t s) {
    if (s == 0) {
        printf("-\n");
        return;
    }
    const char *sep = "";
    for (int i = 0; i < 64; i++) {
        if (s >> i & 1) {
            printf("%s%d", sep, i);
            sep = " ";
        }
    }
    printf("\n");
}

int main(void) {
    uint64_t a = read_set(), b = read_set();
    print_set(a | b);
    print_set(a & b);
    print_set(a & ~b);
    print_set(b & ~a);
    printf("%s\n", (a & b) == a ? "subset" : "not a subset");
    return 0;
}
`,
	"binomial": `#include <stdio.h>

static unsigned long long c[68][68];

int main(void) {
    for (int n = 0; n < 68; n++) {
        c[n][0] = c[n][n] = 1;
        for (int k = 1; k < n; k++) c[n][k] = c[n - 1][k - 1] + c[n - 1][k];
    }
    int n, k;
    while (scanf("%d %d", &n, &k) == 2) {
        if (n < 0 || n > 67 || k < 0 || k > n) return 1;
        printf("%llu\n", c[n][k]);
    }
    return 0;
}
`,
	"modpow": `#include <stdint.h>
#include <stdio.h>

int main(void) {
    unsigned long long a, e, m;
    while (scanf("%llu %llu %llu", &a, &e, &m) == 3) {
        uint64_t result = 1 % m, base = a % m;
        while (e > 0) {
            if (e & 1) result = result * base % m;
            base = base * base % m;
            e >>= 1;
        }
        printf("%llu\n", (unsigned long long)result);
    }
    return 0;
}
`,
	"inversions": `#include <stdio.h>
#include <stdlib.h>

static long long sort_count(int *a, int *tmp, int n) {
    if (n < 2) return 0;
    int h = n / 2;
    long long count = sort_count(a, tmp, h) + sort_count(a + h, tmp, n - h);
    int i = 0, j = h, k = 0;
    while (i < h && j < n) {
        if (a[i] <= a[j]) tmp[k++] = a[i++];
        else {
            tmp[k++] = a[j++];
            count += h - i;
        }
    }
    while (i < h) tmp[k++] = a[i++];
    while (j < n) tmp[k++] = a[j++];
    for (int x = 0; x < n; x++) a[x] = tmp[x];
    return count;
}

int main(void) {
    int n;
    if (scanf("%d", &n) != 1 || n < 0) return 1;
    int *a = malloc((size_t)(n > 0 ? n : 1) * sizeof *a);
    int *tmp = malloc((size_t)(n > 0 ? n : 1) * sizeof *tmp);
    if (a == NULL || tmp == NULL) return 1;
    for (int i = 0; i < n; i++)
        if (scanf("%d", &a[i]) != 1) return 1;
    printf("%lld\n", sort_count(a, tmp, n));
    free(a);
    free(tmp);
    return 0;
}
`,
	"euler": `#include <stdio.h>
#include <stdlib.h>

static int parent[100001], degree[100001];

static int find(int v) {
    while (parent[v] != v) {
        parent[v] = parent[parent[v]];
        v = parent[v];
    }
    return v;
}

int main(void) {
    int n, m;
    if (scanf("%d %d", &n, &m) != 2 || n < 1 || n > 100000) return 1;
    for (int v = 1; v <= n; v++) parent[v] = v;
    for (int i = 0; i < m; i++) {
        int u, v;
        if (scanf("%d %d", &u, &v) != 2) return 1;
        degree[u]++;
        degree[v]++;
        parent[find(u)] = find(v);
    }
    int root = 0, odd[2], odds = 0;
    for (int v = 1; v <= n; v++) {
        if (degree[v] == 0) continue;
        if (root == 0) root = find(v);
        else if (find(v) != root) {
            printf("none\n");
            return 0;
        }
        if (degree[v] % 2 == 1) {
            if (odds < 2) odd[odds] = v;
            odds++;
        }
    }
    if (odds == 0) printf("circuit\n");
    else if (odds == 2) printf("path %d %d\n", odd[0], odd[1]);
    else printf("none\n");
    return 0;
}
`,
	"birthday": `#include <stdio.h>

int main(void) {
    int k;
    while (scanf("%d", &k) == 1) {
        double all_different = 1.0;
        for (int i = 0; i < k && i < 365; i++) all_different *= (365.0 - i) / 365.0;
        if (k > 365) all_different = 0.0;
        printf("%.6f\n", 1.0 - all_different);
    }
    return 0;
}
`,
}

// Each Foundations check fails the typical wrong answer, for the reason
// its lesson teaches.
func TestFoundationsChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := foundationsSolutions
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

	// Comparing every pair is correct but O(n²).
	nested := `#include <stdio.h>

static int a[200000];

int main(void) {
    int n;
    if (scanf("%d", &n) != 1) return 1;
    for (int i = 0; i < n; i++) scanf("%d", &a[i]);
    for (int j = 1; j < n; j++)
        for (int i = 0; i < j; i++)
            if (a[i] == a[j]) { printf("%d\n", a[j]); return 0; }
    printf("none\n");
    return 0;
}
`
	failsWith(t, "duplicates", nested, "200,000 different numbers", "did not finish")
	// x - y as a comparison overflows for numbers far apart.
	failsWith(t, "duplicates", mutate("duplicates", "return (x->value > y->value) - (x->value < y->value);", "return x->value - y->value;"), "extremes and negatives", "")

	// Swapping the pegs in the second call breaks the plan.
	failsWith(t, "hanoi", mutate("hanoi", "    hanoi(n - 1, via, to, from);", "    hanoi(n - 1, via, from, to);"), "two disks", "")

	// Stopping at any match, not the first.
	anyMatch := mutate("bsearch", "        if (a[mid] < x) low = mid + 1;\n        else high = mid;",
		"        if (a[mid] == x) return mid;\n        if (a[mid] < x) low = mid + 1;\n        else high = mid;")
	failsWith(t, "bsearch", anyMatch, "all the same", "")
	// A linear scan per lookup is too slow.
	failsWith(t, "bsearch", mutate("bsearch", "    int low = 0, high = n;\n    while (low < high) {", "    for (int i = 0; i < n; i++) if (a[i] >= x) return i;\n    return n;\n    int low = 0, high = n;\n    while (low < high) {"), "200,000 numbers, 100,000 lookups", "did not finish")

	// Greedy by start time, and an overlap test that rejects back-to-back meetings.
	byStart := mutate("intervals", "return (x->end > y->end) - (x->end < y->end);", "return (x->start > y->start) - (x->start < y->start);")
	failsWith(t, "intervals", byStart, "one long meeting", `"2", but was "1"`)
	failsWith(t, "intervals", mutate("intervals", "if (m[i].start >= last) {", "if (m[i].start > last) {"), "back to back", "")

	// The plain recursion is exponential.
	recursive := `#include <stdio.h>
#include <string.h>

static char a[2100], b[2100];

static int dist(const char *x, const char *y) {
    if (*x == '\0') return (int)strlen(y);
    if (*y == '\0') return (int)strlen(x);
    int best = dist(x + 1, y + 1) + (*x != *y);
    int del = dist(x + 1, y) + 1, ins = dist(x, y + 1) + 1;
    if (del < best) best = del;
    if (ins < best) best = ins;
    return best;
}

int main(void) {
    while (fgets(a, sizeof a, stdin) != NULL && fgets(b, sizeof b, stdin) != NULL) {
        a[strcspn(a, "\n")] = '\0';
        b[strcspn(b, "\n")] = '\0';
        printf("%d\n", dist(a, b));
    }
    return 0;
}
`
	failsWith(t, "edit_distance", recursive, "2,000 characters each", "did not finish")

	// Numbering variables from bit k instead of bit k - 1.
	failsWith(t, "sat_count", mutate("sat_count", "(row >> (v - 1) & 1)", "(row >> v & 1)"), "the last variable", "")

	// 1 << v is an int shift: wrong from bit 31 on.
	failsWith(t, "bitsets", mutate("bitsets", "s |= (uint64_t)1 << v;", "s |= 1 << v;"), "high members", "")

	// The factorial formula wraps from 21!, and multiplying first overflows at 67.
	factorial := `#include <stdio.h>

int main(void) {
    int n, k;
    while (scanf("%d %d", &n, &k) == 2) {
        unsigned long long f[68] = {1};
        for (int i = 1; i <= n; i++) f[i] = f[i - 1] * i;
        printf("%llu\n", f[n] / (f[k] * f[n - k]));
    }
    return 0;
}
`
	failsWith(t, "binomial", factorial, "just past 20!", "")
	multiplicative := `#include <stdio.h>

int main(void) {
    int n, k;
    while (scanf("%d %d", &n, &k) == 2) {
        unsigned long long c = 1;
        for (int i = 1; i <= k; i++) c = c * (unsigned long long)(n - k + i) / (unsigned long long)i;
        printf("%llu\n", c);
    }
    return 0;
}
`
	failsWith(t, "binomial", multiplicative, "the largest", "")

	// Multiplying e times, and forgetting to reduce a first.
	failsWith(t, "modpow", mutate("modpow", "        while (e > 0) {\n            if (e & 1) result = result * base % m;\n            base = base * base % m;\n            e >>= 1;\n        }", "        for (unsigned long long i = 0; i < e; i++) result = result * base % m;"), "huge exponents", "did not finish")
	failsWith(t, "modpow", mutate("modpow", "base = a % m;", "base = a;"), "a bigger than m", "")

	// An int counter overflows, and counting every pair is O(n²).
	failsWith(t, "inversions", strings.NewReplacer("static long long sort_count", "static int sort_count", "long long count =", "int count =", "%lld", "%d").Replace(ref["inversions"]), "200,000 in reverse", "")
	pairs := `#include <stdio.h>

static int a[200000];

int main(void) {
    int n;
    if (scanf("%d", &n) != 1) return 1;
    for (int i = 0; i < n; i++) scanf("%d", &a[i]);
    long long count = 0;
    for (int i = 0; i < n; i++)
        for (int j = i + 1; j < n; j++)
            if (a[i] > a[j]) count++;
    printf("%lld\n", count);
    return 0;
}
`
	failsWith(t, "inversions", pairs, "200,000 in reverse", "did not finish")

	// Degrees alone, without checking that the edges are connected.
	failsWith(t, "euler", mutate("euler", "        else if (find(v) != root) {\n            printf(\"none\\n\");\n            return 0;\n        }\n", ""), "two islands", `"none", but was "circuit"`)

	// Integer division makes every factor 0 or 1.
	failsWith(t, "birthday", mutate("birthday", "(365.0 - i) / 365.0", "(365 - i) / 365"), "twenty-three people", "")
}
