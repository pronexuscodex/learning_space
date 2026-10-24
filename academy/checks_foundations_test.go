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
		if _, dup := referenceSolutions[id]; dup {
			panic("two reference solutions for check " + id)
		}
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
	"bases": `#include <ctype.h>
#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

int main(void) {
    char line[128];
    while (fgets(line, sizeof line, stdin) != NULL) {
        line[strcspn(line, "\r\n")] = '\0';
        const char *p = line;
        unsigned base = 10;
        if (p[0] == '0' && (p[1] == 'x' || p[1] == 'X')) { base = 16; p += 2; }
        else if (p[0] == '0' && (p[1] == 'b' || p[1] == 'B')) { base = 2; p += 2; }
        uint64_t v = 0;
        for (; *p != '\0'; p++) {
            unsigned d = isdigit((unsigned char)*p) ? (unsigned)(*p - '0') : (unsigned)(tolower((unsigned char)*p) - 'a' + 10);
            v = v * base + d;
        }
        uint32_t x = (uint32_t)v;
        printf("%" PRIu32 " 0x%" PRIx32 " 0b", x, x);
        int top = 31;
        while (top > 0 && !(x >> top & 1)) top--;
        for (int i = top; i >= 0; i--) putchar((x >> i & 1) ? '1' : '0');
        putchar('\n');
    }
    return 0;
}
`,
	"nand_gates": `int nand(int a, int b);

int not_(int a) { return nand(a, a); }

int and_(int a, int b) {
    int x = nand(a, b);
    return nand(x, x);
}

int or_(int a, int b) { return nand(not_(a), not_(b)); }

int xor_(int a, int b) {
    int m = nand(a, b);
    return nand(nand(a, m), nand(b, m));
}

int full_add(int a, int b, int cin, int *cout) {
    int ab = xor_(a, b);
    *cout = or_(and_(a, b), and_(cin, ab));
    return xor_(ab, cin);
}
`,
	"alu": `#include <stdint.h>

struct flags { int z, n, c, v; };

uint32_t alu(uint32_t a, uint32_t b, char op, struct flags *f) {
    uint32_t r;
    f->c = f->v = 0;
    switch (op) {
    case '+':
        r = a + b;
        f->c = r < a;
        f->v = (int)(((a ^ r) & (b ^ r)) >> 31);
        break;
    case '-':
        r = a - b;
        f->c = a < b;
        f->v = (int)(((a ^ b) & (a ^ r)) >> 31);
        break;
    case '&': r = a & b; break;
    case '|': r = a | b; break;
    default:  r = a ^ b; break;
    }
    f->z = r == 0;
    f->n = (int)(r >> 31);
    return r;
}
`,
	"number_fsm": `#include <ctype.h>
#include <stdio.h>
#include <string.h>

enum state { START, SIGN, INT, DOT_NO_INT, FRAC, EXP, EXP_SIGN, EXP_INT, ERROR };

static char line[10100];

int main(void) {
    while (fgets(line, sizeof line, stdin) != NULL) {
        line[strcspn(line, "\n")] = '\0';
        enum state s = START;
        for (const char *p = line; *p != '\0' && s != ERROR; p++) {
            int c = (unsigned char)*p, digit = isdigit(c), sign = c == '+' || c == '-', e = c == 'e' || c == 'E';
            switch (s) {
            case START:      s = sign ? SIGN : digit ? INT : c == '.' ? DOT_NO_INT : ERROR; break;
            case SIGN:       s = digit ? INT : c == '.' ? DOT_NO_INT : ERROR; break;
            case INT:        s = digit ? INT : c == '.' ? FRAC : e ? EXP : ERROR; break;
            case DOT_NO_INT: s = digit ? FRAC : ERROR; break;
            case FRAC:       s = digit ? FRAC : e ? EXP : ERROR; break;
            case EXP:        s = sign ? EXP_SIGN : digit ? EXP_INT : ERROR; break;
            case EXP_SIGN:   s = digit ? EXP_INT : ERROR; break;
            case EXP_INT:    s = digit ? EXP_INT : ERROR; break;
            case ERROR:      break;
            }
        }
        printf("%s\n", s == INT || s == FRAC || s == EXP_INT ? "valid" : "invalid");
    }
    return 0;
}
`,
	"cpu": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

enum op { LOAD, ADD, SUB, JZ, JMP, PRINT, HALT };
struct insn { enum op op; int a; long long b; };

static struct insn code[1000];

static int reg(const char *s) { return s[0] == 'r' ? atoi(s + 1) & 3 : 0; }

int main(void) {
    char line[128], word[16], x[32], y[32];
    int count = 0;
    while (count < 1000 && fgets(line, sizeof line, stdin) != NULL) {
        x[0] = y[0] = '\0';
        if (sscanf(line, "%15s %31s %31s", word, x, y) < 1) continue;
        struct insn *in = &code[count++];
        if (strcmp(word, "LOAD") == 0) { in->op = LOAD; in->a = reg(x); in->b = strtoll(y, NULL, 10); }
        else if (strcmp(word, "ADD") == 0) { in->op = ADD; in->a = reg(x); in->b = reg(y); }
        else if (strcmp(word, "SUB") == 0) { in->op = SUB; in->a = reg(x); in->b = reg(y); }
        else if (strcmp(word, "JZ") == 0) { in->op = JZ; in->a = reg(x); in->b = atoi(y); }
        else if (strcmp(word, "JMP") == 0) { in->op = JMP; in->b = atoi(x); }
        else if (strcmp(word, "PRINT") == 0) { in->op = PRINT; in->a = reg(x); }
        else in->op = HALT;
    }
    long long r[4] = {0, 0, 0, 0};
    long steps = 0;
    int pc = 0;
    while (pc >= 0 && pc < count) {
        if (++steps > 1000000) {
            printf("too many steps\n");
            return 0;
        }
        struct insn in = code[pc++];
        switch (in.op) {
        case LOAD:  r[in.a] = in.b; break;
        case ADD:   r[in.a] += r[in.b]; break;
        case SUB:   r[in.a] -= r[in.b]; break;
        case JZ:    if (r[in.a] == 0) pc = (int)in.b; break;
        case JMP:   pc = (int)in.b; break;
        case PRINT: printf("%lld\n", r[in.a]); break;
        case HALT:  return 0;
        }
    }
    return 0;
}
`,
	"rv_decode": `#include <stdint.h>
#include <stdio.h>

int main(void) {
    unsigned int w;
    while (scanf("%x", &w) == 1) {
        uint32_t opcode = w & 0x7f, rd = (w >> 7) & 0x1f, f3 = (w >> 12) & 7;
        uint32_t rs1 = (w >> 15) & 0x1f, rs2 = (w >> 20) & 0x1f, f7 = w >> 25;
        int32_t imm = (int32_t)w >> 20;
        const char *name = NULL;
        if (opcode == 0x33) {
            if (f7 == 0 && f3 == 0) name = "add";
            else if (f7 == 0x20 && f3 == 0) name = "sub";
            else if (f7 == 0 && f3 == 4) name = "xor";
            else if (f7 == 0 && f3 == 6) name = "or";
            else if (f7 == 0 && f3 == 7) name = "and";
            if (name != NULL) {
                printf("%s x%u, x%u, x%u\n", name, (unsigned)rd, (unsigned)rs1, (unsigned)rs2);
                continue;
            }
        } else if (opcode == 0x13) {
            if (f3 == 0) name = "addi";
            else if (f3 == 4) name = "xori";
            else if (f3 == 6) name = "ori";
            else if (f3 == 7) name = "andi";
            if (name != NULL) {
                printf("%s x%u, x%u, %d\n", name, (unsigned)rd, (unsigned)rs1, (int)imm);
                continue;
            }
        }
        printf("unknown\n");
    }
    return 0;
}
`,
	"linkage": `static int count;

static void bump(int by) { count += by; }

void tick(void) { bump(1); }
int ticks(void) { return count; }
void reset(void) { count = 0; }
`,
	"checked_add": `#include <limits.h>

int checked_add(int a, int b, int *sum) {
    if (b > 0 ? a > INT_MAX - b : a < INT_MIN - b) return 0;
    *sum = a + b;
    return 1;
}

int checked_mul(int a, int b, int *product) {
    long long p = (long long)a * b;
    if (p > INT_MAX || p < INT_MIN) return 0;
    *product = (int)p;
    return 1;
}
`,
	"arena": `#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

struct arena {
    char *base;   /* the block, from one malloc */
    size_t size;  /* its size in bytes */
    size_t used;  /* bytes handed out so far, padding included */
};

int arena_init(struct arena *a, size_t size) {
    a->base = malloc(size);
    a->size = a->base ? size : 0;
    a->used = 0;
    return a->base != NULL;
}

void *arena_alloc(struct arena *a, size_t n) {
    uintptr_t p = (uintptr_t)(a->base + a->used);
    p = (p + 15) & ~(uintptr_t)15;
    size_t start = p - (uintptr_t)a->base;
    if (start > a->size || n > a->size - start) return NULL;
    a->used = start + n;
    return a->base + start;
}

void arena_reset(struct arena *a) { a->used = 0; }

void arena_destroy(struct arena *a) {
    free(a->base);
    a->base = NULL;
    a->size = a->used = 0;
}
`,
	"pipeline": `#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

static pid_t start(char **argv, int fd, int onto, int fds[2]) {
    pid_t pid = fork();
    if (pid == 0) {
        dup2(fd, onto);
        close(fds[0]);
        close(fds[1]);
        execvp(argv[0], argv);
        perror(argv[0]);
        _exit(127);
    }
    return pid;
}

int main(int argc, char **argv) {
    int bar = 0;
    for (int i = 1; i < argc; i++)
        if (strcmp(argv[i], "|") == 0) { bar = i; break; }
    if (bar <= 1 || bar == argc - 1) {
        fprintf(stderr, "usage: pipeline A ARGS... '|' B ARGS...\n");
        return 2;
    }
    argv[bar] = NULL;
    int fds[2];
    if (pipe(fds) != 0) { perror("pipe"); return 1; }
    pid_t a = start(&argv[1], fds[1], 1, fds);
    pid_t b = start(&argv[bar + 1], fds[0], 0, fds);
    close(fds[0]);
    close(fds[1]);
    int status = 0, bstatus = 0;
    waitpid(a, &status, 0);
    waitpid(b, &bstatus, 0);
    return WIFEXITED(bstatus) ? WEXITSTATUS(bstatus) : 128 + WTERMSIG(bstatus);
}
`,
	"graceful": `#define _POSIX_C_SOURCE 200809L
#include <signal.h>
#include <string.h>

static volatile sig_atomic_t got;

static void on_stop(int sig) { got = sig; }

int install_stop_handlers(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof sa);
    sa.sa_handler = on_stop;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = 0;
    if (sigaction(SIGTERM, &sa, NULL) != 0) return -1;
    if (sigaction(SIGINT, &sa, NULL) != 0) return -1;
    return 0;
}

int stop_signal(void) { return got; }
`,
	"buffer_pool": `#include <stdio.h>
#include <string.h>

#define MAXFRAMES 4096

struct frame { long page; int pins, dirty; long used; };

int main(void) {
    struct frame pool[MAXFRAMES];
    int frames, n = 0;
    char mode[16], op[4];
    long page, clock = 0, hits = 0, misses = 0, writes = 0;
    if (scanf("%d %15s", &frames, mode) != 2 || frames < 1 || frames > MAXFRAMES) return 1;
    int trace = strcmp(mode, "trace") == 0;
    while (scanf("%3s %ld", op, &page) == 2) {
        struct frame *f = NULL;
        for (int i = 0; i < n; i++)
            if (pool[i].page == page) { f = &pool[i]; break; }
        if (op[0] == 'u') {
            if (f != NULL && f->pins > 0) f->pins--;
            continue;
        }
        clock++;
        if (f != NULL) {
            hits++;
        } else {
            misses++;
            if (n < frames) {
                f = &pool[n++];
                if (trace) printf("load %ld\n", page);
            } else {
                for (int i = 0; i < n; i++)
                    if (pool[i].pins == 0 && (f == NULL || pool[i].used < f->used)) f = &pool[i];
                if (f == NULL) {
                    if (trace) printf("no frame for %ld\n", page);
                    continue;
                }
                if (f->dirty) {
                    writes++;
                    if (trace) printf("write %ld, ", f->page);
                }
                if (trace) printf("evict %ld, load %ld\n", f->page, page);
            }
            f->page = page;
            f->pins = 0;
            f->dirty = 0;
        }
        f->used = clock;
        if (op[0] == 'w') f->dirty = 1;
        if (op[0] == 'p') f->pins++;
    }
    printf("hits %ld, misses %ld, writes %ld\n", hits, misses, writes);
    return 0;
}
`,
	"bloom": `#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

struct bloom {
    unsigned char *bits;  /* m bits, (m + 7) / 8 bytes */
    size_t m;             /* number of bits */
    int k;                /* number of hash functions */
};

static uint64_t fnv1a(const char *key, uint64_t h) {
    for (const unsigned char *p = (const unsigned char *)key; *p; p++) {
        h ^= *p;
        h *= 1099511628211u;
    }
    return h;
}

static size_t position(const struct bloom *b, const char *key, int j) {
    uint64_t h1 = fnv1a(key, 14695981039346656037u);
    uint64_t h2 = fnv1a(key, 0x9E3779B97F4A7C15u) | 1;
    return (size_t)((h1 + (uint64_t)j * h2) % b->m);
}

int bloom_init(struct bloom *b, size_t m, int k) {
    b->bits = calloc((m + 7) / 8, 1);
    b->m = m;
    b->k = k;
    return b->bits != NULL;
}

void bloom_add(struct bloom *b, const char *key) {
    for (int j = 0; j < b->k; j++) {
        size_t i = position(b, key, j);
        b->bits[i / 8] |= (unsigned char)(1u << (i % 8));
    }
}

int bloom_maybe(const struct bloom *b, const char *key) {
    for (int j = 0; j < b->k; j++) {
        size_t i = position(b, key, j);
        if (!((b->bits[i / 8] >> (i % 8)) & 1)) return 0;
    }
    return 1;
}

void bloom_free(struct bloom *b) {
    free(b->bits);
    b->bits = NULL;
}
`,
	"hash_join": `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#define BUCKETS (1 << 18)

struct row { long long key; char word[16]; int next; };

static size_t bucket(long long key) {
    return (size_t)(((uint64_t)key * 0x9E3779B97F4A7C15u) >> 46);
}

int main(void) {
    int n, m;
    if (scanf("%d", &n) != 1 || n < 0) return 1;
    struct row *a = malloc(((size_t)n + 1) * sizeof *a);
    int *head = malloc(BUCKETS * sizeof *head);
    if (a == NULL || head == NULL) return 1;
    for (size_t i = 0; i < BUCKETS; i++) head[i] = -1;
    for (int i = 0; i < n; i++)
        if (scanf("%lld %15s", &a[i].key, a[i].word) != 2) return 1;
    for (int i = n - 1; i >= 0; i--) {   /* backwards, so each chain is in A's order */
        size_t h = bucket(a[i].key);
        a[i].next = head[h];
        head[h] = i;
    }
    if (scanf("%d", &m) != 1) return 1;
    for (int j = 0; j < m; j++) {
        long long key;
        char word[16];
        if (scanf("%lld %15s", &key, word) != 2) return 1;
        for (int i = head[bucket(key)]; i != -1; i = a[i].next)
            if (a[i].key == key) printf("%lld %s %s\n", key, a[i].word, word);
    }
    free(a);
    free(head);
    return 0;
}
`,
	"routes": `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct node { int child[2]; int hop; };

static struct node *trie;
static int nodes, capacity;
static char (*hops)[64];

static int new_node(void) {
    if (nodes == capacity) {
        capacity = capacity ? capacity * 2 : 1024;
        trie = realloc(trie, (size_t)capacity * sizeof *trie);
        if (trie == NULL) exit(1);
    }
    trie[nodes].child[0] = trie[nodes].child[1] = -1;
    trie[nodes].hop = -1;
    return nodes++;
}

static int parse(const char *s, uint32_t *addr, int *len) {
    unsigned a, b, c, d;
    int n = sscanf(s, "%u.%u.%u.%u/%d", &a, &b, &c, &d, len);
    if (n < 4 || a > 255 || b > 255 || c > 255 || d > 255) return 0;
    *addr = (uint32_t)a << 24 | (uint32_t)b << 16 | (uint32_t)c << 8 | d;
    return n;
}

int main(void) {
    int n;
    char line[128], hop[64];
    if (fgets(line, sizeof line, stdin) == NULL || sscanf(line, "%d", &n) != 1 || n < 0) return 1;
    hops = malloc(((size_t)n + 1) * sizeof *hops);
    if (hops == NULL) return 1;
    new_node();
    for (int i = 0; i < n; i++) {
        uint32_t addr;
        int len;
        char prefix[64];
        if (fgets(line, sizeof line, stdin) == NULL || sscanf(line, "%63s %63s", prefix, hop) != 2) return 1;
        if (parse(prefix, &addr, &len) != 5 || len < 0 || len > 32) return 1;
        int at = 0;
        for (int bit = 0; bit < len; bit++) {
            int b = addr >> (31 - bit) & 1;
            if (trie[at].child[b] == -1) {
                int fresh = new_node();
                trie[at].child[b] = fresh;
            }
            at = trie[at].child[b];
        }
        strcpy(hops[i], hop);
        trie[at].hop = i;
    }
    while (fgets(line, sizeof line, stdin) != NULL) {
        uint32_t addr;
        int len;
        if (parse(line, &addr, &len) != 4) continue;
        int at = 0, best = trie[0].hop;
        for (int bit = 0; bit < 32 && at != -1; bit++) {
            at = trie[at].child[addr >> (31 - bit) & 1];
            if (at != -1 && trie[at].hop != -1) best = trie[at].hop;
        }
        puts(best == -1 ? "unreachable" : hops[best]);
    }
    free(trie);
    free(hops);
    return 0;
}
`,
	"retry": `int attempt(void *op);
long jitter(long max);
void wait_ms(long ms);

int with_retries(void *op, int max_attempts, long base_ms, long cap_ms) {
    long d = base_ms < cap_ms ? base_ms : cap_ms;
    for (int i = 1; i <= max_attempts; i++) {
        int r = attempt(op);
        if (r != 1) return r;
        if (i == max_attempts) break;
        wait_ms(jitter(d));
        d = d > cap_ms / 2 ? cap_ms : d * 2;
    }
    return 1;
}
`,
	"ring": `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static uint32_t fnv1a(const char *s) {
    uint32_t h = 2166136261u;
    for (; *s; s++) {
        h ^= (unsigned char)*s;
        h *= 16777619u;
    }
    return h;
}

struct point { uint32_t hash; int node; };

static int by_hash(const void *x, const void *y) {
    const struct point *a = x, *b = y;
    if (a->hash != b->hash) return a->hash < b->hash ? -1 : 1;
    return a->node - b->node;
}

int main(void) {
    int v, n, keys;
    if (scanf("%d %d", &v, &n) != 2 || v < 1 || n < 1) return 1;
    char (*names)[64] = malloc((size_t)n * sizeof *names);
    struct point *pts = malloc((size_t)n * v * sizeof *pts);
    if (names == NULL || pts == NULL) return 1;
    int count = 0;
    for (int i = 0; i < n; i++) {
        if (scanf("%63s", names[i]) != 1) return 1;
        for (int j = 0; j < v; j++) {
            char buf[96];
            snprintf(buf, sizeof buf, "%s#%d", names[i], j);
            pts[count].hash = fnv1a(buf);
            pts[count].node = i;
            count++;
        }
    }
    qsort(pts, (size_t)count, sizeof *pts, by_hash);
    if (scanf("%d", &keys) != 1) return 1;
    char key[64];
    for (int k = 0; k < keys && scanf("%63s", key) == 1; k++) {
        uint32_t h = fnv1a(key);
        int lo = 0, hi = count;
        while (lo < hi) {
            int mid = lo + (hi - lo) / 2;
            if (pts[mid].hash < h) lo = mid + 1;
            else hi = mid;
        }
        if (lo == count) lo = 0;
        puts(names[pts[lo].node]);
    }
    free(names);
    free(pts);
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
	// x - y as a comparison overflows for numbers far apart. Only the
	// sanitizers see it for certain: without them, what a wrapped
	// comparison does to the sort depends on the C library's qsort.
	wrapped := mutate("duplicates", "return (x->value > y->value) - (x->value < y->value);", "return x->value - y->value;")
	if runSource(t, "duplicates", wrapped).Sanitizers {
		failsWith(t, "duplicates", wrapped, "extremes and negatives", "")
	}

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

	// A signed int prints the largest value as -1.
	failsWith(t, "bases", mutate("bases", `printf("%" PRIu32 " 0x%" PRIx32 " 0b", x, x);`, `printf("%d 0x%" PRIx32 " 0b", (int)x, x);`), "the largest", `but was "-1`)

	// AND as NOT(NAND) with a second NAND call is one gate too many.
	failsWith(t, "nand_gates", mutate("nand_gates", "    int x = nand(a, b);\n    return nand(x, x);", "    return nand(nand(a, b), nand(a, b));"), "and", "NAND gates: 3")
	// XOR from AND, OR and NOT works but costs far more gates.
	failsWith(t, "nand_gates", mutate("nand_gates", "    int m = nand(a, b);\n    return nand(nand(a, m), nand(b, m));", "    return or_(and_(a, not_(b)), and_(not_(a), b));"), "xor", "NAND gates: 9")

	// The carry flag computed for subtraction like addition.
	failsWith(t, "alu", mutate("alu", "        f->c = a < b;", "        f->c = r > a;\n        f->c = 0;"), "subtraction", "C0")
	// Overflow from the result's sign alone.
	failsWith(t, "alu", mutate("alu", "        f->v = (int)(((a ^ r) & (b ^ r)) >> 31);", "        f->v = (int)(r >> 31);"), "addition", "")

	// A dot alone accepted as a number.
	failsWith(t, "number_fsm", mutate("number_fsm", "s == INT || s == FRAC || s == EXP_INT", "s == INT || s == FRAC || s == EXP_INT || s == DOT_NO_INT"), "not numbers", "")
	// An exponent sign accepted without digits.
	failsWith(t, "number_fsm", mutate("number_fsm", "s == INT || s == FRAC || s == EXP_INT", "s == INT || s == FRAC || s == EXP_INT || s == EXP_SIGN"), "exponents", "")

	// No step limit: the endless loop never ends.
	failsWith(t, "cpu", mutate("cpu", "        if (++steps > 1000000) {", "        if (++steps < 0) {"), "an endless loop", "did not finish")
	// int registers overflow.
	failsWith(t, "cpu", strings.NewReplacer("long long r[4]", "int r[4]", `printf("%lld\n", r[in.a])`, `printf("%d\n", r[in.a])`, "r[in.a] = in.b;", "r[in.a] = (int)in.b;").Replace(ref["cpu"]), "big numbers", "")

	// A global counter clashes with the checker's own count at link time.
	run := runSource(t, "linkage", strings.Replace(ref["linkage"], "static int count;", "int count;", 1))
	if run.Built || !strings.Contains(run.BuildLog, "Make yours static") {
		t.Errorf("linkage: a global count must fail to link with the static hint (built=%v)\n%s", run.Built, run.BuildLog)
	}

	// Checking after the addition is undefined behaviour, which the sanitizer stops.
	late := mutate("checked_add", "    if (b > 0 ? a > INT_MAX - b : a < INT_MIN - b) return 0;\n    *sum = a + b;\n    return 1;",
		"    int s = a + b;\n    if ((a > 0 && b > 0 && s < 0) || (a < 0 && b < 0 && s >= 0)) return 0;\n    *sum = s;\n    return 1;")
	if run := runSource(t, "checked_add", late); run.Sanitizers {
		for _, c := range run.Cases {
			if c.Case.Name == "additions at the edges" && (c.Pass || !strings.Contains(c.Problem, "overflow")) {
				t.Errorf("checked_add: an addition before the check must be reported as overflow: pass=%v problem=%q", c.Pass, c.Problem)
			}
		}
	}
	// Forgetting that INT_MIN * -1 overflows.
	failsWith(t, "checked_add", mutate("checked_add", "    long long p = (long long)a * b;\n    if (p > INT_MAX || p < INT_MIN) return 0;\n    *product = (int)p;\n    return 1;",
		"    if (a != 0 && (b > INT_MAX / a || b < INT_MIN / a) && a > 0) return 0;\n    *product = a * b;\n    return 1;"), "multiplications at the edges", "")

	// The immediate read without sign extension.
	failsWith(t, "rv_decode", mutate("rv_decode", "int32_t imm = (int32_t)w >> 20;", "int32_t imm = (int32_t)(w >> 20);"), "assembled by LLVM", `"addi x2, x1, 4095"`)
	// funct7 ignored: sub decodes as add, and junk as an instruction.
	failsWith(t, "rv_decode", mutate("rv_decode", "            if (f7 == 0 && f3 == 0) name = \"add\";", "            if (f7 != 0x20 && f3 == 0) name = \"add\";"), "not ours", "")

	// Aligning the offset alone is fine here, but used + n wraps around.
	failsWith(t, "arena", mutate("arena", "if (start > a->size || n > a->size - start) return NULL;", "if (start + n > a->size) return NULL;"), "a request too big to add", "")
	// No alignment at all.
	failsWith(t, "arena", mutate("arena", "    p = (p + 15) & ~(uintptr_t)15;\n", ""), "sizes 1, 3, 24, 100 and 7", "multiple of 16")
	// A malloc per allocation is not an arena (and leaks, where the leak checker runs).
	failsWith(t, "arena", mutate("arena", "    a->used = start + n;\n    return a->base + start;", "    a->used = start + n;\n    return malloc(n);"), "sizes 1, 3, 24, 100 and 7", "")
	// Forgetting to free the block leaks it.
	if run := runSource(t, "arena", mutate("arena", "    free(a->base);\n", "")); run.Leaks && run.OK() {
		t.Error("arena: a block never freed must be reported as a leak")
	}

	// The parent keeps the write end open: cat never sees end-of-file.
	failsWith(t, "pipeline", mutate("pipeline", "    close(fds[0]);\n    close(fds[1]);\n    int status", "    close(fds[0]);\n    int status"), "echo hello | cat", "did not finish")
	// Waiting for A before starting B deadlocks on a full pipe.
	failsWith(t, "pipeline", mutate("pipeline", "    pid_t b = start(&argv[bar + 1], fds[0], 0, fds);\n", "    int early;\n    waitpid(a, &early, 0);\n    pid_t b = start(&argv[bar + 1], fds[0], 0, fds);\n"), "more than the pipe holds", "did not finish")
	// Returning A's status instead of B's.
	failsWith(t, "pipeline", mutate("pipeline", "return WIFEXITED(bstatus) ? WEXITSTATUS(bstatus) : 128 + WTERMSIG(bstatus);", "return WEXITSTATUS(status);"), "B's exit status", "expected 1")

	// signal() restarts the blocking read.
	failsWith(t, "graceful", mutate("graceful", "    sa.sa_flags = 0;", "    sa.sa_flags = SA_RESTART;"), "from another process, while blocked in read", "restarted")
	// SIGINT forgotten: Ctrl-C still kills the program.
	failsWith(t, "graceful", mutate("graceful", "    if (sigaction(SIGINT, &sa, NULL) != 0) return -1;\n", ""), "SIGINT", "")
	// A handler that exits straight away gives the program no chance to clean up.
	failsWith(t, "graceful", strings.Replace(mutate("graceful", "static void on_stop(int sig) { got = sig; }", "static void on_stop(int sig) { got = sig; _exit(0); }"), "#include <string.h>", "#include <string.h>\n#include <unistd.h>", 1), "SIGTERM", "")

	// Evicting the oldest loaded page (FIFO) instead of the least recently used.
	failsWith(t, "buffer_pool", mutate("buffer_pool", "        f->used = clock;\n        if (op[0] == 'w')", "        if (f->used == 0 || misses + hits == 1) f->used = clock;\n        if (op[0] == 'w')"), "three frames", "")
	// Ignoring pins.
	failsWith(t, "buffer_pool", mutate("buffer_pool", "if (pool[i].pins == 0 && (f == NULL", "if ((f == NULL"), "a pinned page stays", "")
	// Writing back every evicted page, dirty or not.
	failsWith(t, "buffer_pool", mutate("buffer_pool", "                if (f->dirty) {", "                if (1) {"), "three frames", "")

	// One hash for all k positions: the false-positive rate soars.
	failsWith(t, "bloom", mutate("bloom", "    return (size_t)((h1 + (uint64_t)j * h2) % b->m);", "    (void)h2;\n    (void)j;\n    return (size_t)(h1 % b->m);"), "the false-positive rate", "")
	// A hash of the first two characters only.
	failsWith(t, "bloom", mutate("bloom", "    for (const unsigned char *p = (const unsigned char *)key; *p; p++) {", "    for (const unsigned char *p = (const unsigned char *)key; *p && p < (const unsigned char *)key + 2; p++) {"), "the false-positive rate", "")
	// m / 8 bytes is one byte short for 1,001 bits.
	if run := runSource(t, "bloom", mutate("bloom", "calloc((m + 7) / 8, 1)", "calloc(m / 8, 1)")); run.Sanitizers && run.OK() {
		t.Error("bloom: m / 8 bytes must overflow for 1,001 bits")
	}
	// add and maybe disagree on the bit order.
	failsWith(t, "bloom", mutate("bloom", "        if (!((b->bits[i / 8] >> (i % 8)) & 1)) return 0;", "        if (!((b->bits[i / 8] >> (7 - i % 8)) & 1)) return 0;"), "no false negatives", "")

	// A nested loop is correct but far too slow.
	nestedJoin := mutate("hash_join", "        for (int i = head[bucket(key)]; i != -1; i = a[i].next)", "        for (int i = 0; i < n; i++)")
	failsWith(t, "hash_join", nestedJoin, "100,000 rows each", "did not finish")
	// Chains built forwards come out in reverse order.
	failsWith(t, "hash_join", mutate("hash_join", "for (int i = n - 1; i >= 0; i--) {", "for (int i = 0; i < n; i++) {"), "the example", "")
	// A signed modulo gives a negative bucket for a negative key.
	if run := runSource(t, "hash_join", mutate("hash_join", "    return (size_t)(((uint64_t)key * 0x9E3779B97F4A7C15u) >> 46);", "    return (size_t)(long long)(key % BUCKETS);")); run.OK() {
		t.Error("hash_join: a negative bucket index must not pass")
	}

	// Only the deepest node reached counts, not the last route passed.
	failsWith(t, "routes", mutate("routes", "            if (at != -1 && trie[at].hop != -1) best = trie[at].hop;", "            if (at != -1) best = trie[at].hop;"), "the most specific route wins", "")
	// The first route for a prefix wins instead of the later one.
	failsWith(t, "routes", mutate("routes", "        trie[at].hop = i;", "        if (trie[at].hop == -1) trie[at].hop = i;"), "the later route wins", "")
	// Comparing every address with every route.
	linear := `#include <stdint.h>
#include <stdio.h>

static uint32_t net[100000], mask[100000];
static int len[100000];
static char hop[100000][64];

int main(void) {
    int n;
    char line[128], prefix[64];
    unsigned a, b, c, d;
    if (fgets(line, sizeof line, stdin) == NULL || sscanf(line, "%d", &n) != 1) return 1;
    for (int i = 0; i < n; i++) {
        if (fgets(line, sizeof line, stdin) == NULL || sscanf(line, "%63s %63s", prefix, hop[i]) != 2) return 1;
        sscanf(prefix, "%u.%u.%u.%u/%d", &a, &b, &c, &d, &len[i]);
        mask[i] = len[i] == 0 ? 0 : 0xFFFFFFFFu << (32 - len[i]);
        net[i] = ((uint32_t)a << 24 | (uint32_t)b << 16 | (uint32_t)c << 8 | d) & mask[i];
    }
    while (fgets(line, sizeof line, stdin) != NULL) {
        if (sscanf(line, "%u.%u.%u.%u", &a, &b, &c, &d) != 4) continue;
        uint32_t addr = (uint32_t)a << 24 | (uint32_t)b << 16 | (uint32_t)c << 8 | d;
        int best = -1;
        for (int i = 0; i < n; i++)
            if ((addr & mask[i]) == net[i] && (best == -1 || len[i] >= len[best])) best = i;
        puts(best == -1 ? "unreachable" : hop[best]);
    }
    return 0;
}
`
	if run := runSource(t, "routes", linear); !run.Built {
		t.Fatalf("routes: the linear version did not build\n%s", run.BuildLog)
	} else {
		for _, c := range run.Cases {
			if c.Case.Name != "100,000 routes" && !c.Pass {
				t.Errorf("routes: the linear version should pass %q: %s", c.Case.Name, c.Problem)
			}
		}
	}
	failsWith(t, "routes", linear, "100,000 routes", "did not finish")
	// /0 with a 32-bit shift: undefined, and on x86 a mask of all ones.
	failsWith(t, "routes", strings.Replace(linear, "len[i] == 0 ? 0 : 0xFFFFFFFFu << (32 - len[i])", "0xFFFFFFFFu << (32 - len[i])", 1), "/32 and /0", "")

	// Computing base × 2^(i-1) afresh each time overflows a long.
	overflow := mutate("retry", "        wait_ms(jitter(d));", "        long e = base_ms * (1L << (i - 1));\n        if (e > cap_ms) e = cap_ms;\n        wait_ms(jitter(e));")
	if run := runSource(t, "retry", overflow); run.Sanitizers {
		for _, c := range run.Cases {
			if c.Case.Name == "seventy attempts" && (c.Pass || !strings.Contains(c.Problem, "UndefinedBehaviorSanitizer")) {
				t.Errorf("retry: doubling past LONG_MAX must be reported as undefined behaviour: pass=%v problem=%q", c.Pass, c.Problem)
			}
		}
	}
	// Waiting after the last attempt too.
	failsWith(t, "retry", mutate("retry", "        if (i == max_attempts) break;\n", ""), "out of attempts", "")
	// Retrying permanent failures.
	failsWith(t, "retry", mutate("retry", "        if (r != 1) return r;", "        if (r == 0) return r;"), "a permanent failure is not retried", "")
	// Backoff without jitter.
	failsWith(t, "retry", mutate("retry", "        wait_ms(jitter(d));", "        wait_ms(d);"), "two failures, then success", "")

	// No wrap-around past the largest point.
	failsWith(t, "ring", mutate("ring", "        if (lo == count) lo = 0;\n", ""), "three nodes", "")
	// A linear scan for each key.
	failsWith(t, "ring", mutate("ring", "        int lo = 0, hi = count;\n        while (lo < hi) {\n            int mid = lo + (hi - lo) / 2;\n            if (pts[mid].hash < h) lo = mid + 1;\n            else hi = mid;\n        }",
		"        int lo = 0;\n        while (lo < count && pts[lo].hash < h) lo++;"), "100,000 points", "did not finish")
}
