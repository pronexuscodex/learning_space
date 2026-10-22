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
}
