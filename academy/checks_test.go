package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestChecksAreWellFormed(t *testing.T) {
	ids := map[string]bool{}
	for _, ch := range checks {
		if ch.ID == "" || ids[ch.ID] || strings.ContainsAny(ch.ID, " /\\") {
			t.Errorf("bad or duplicate check ID %q", ch.ID)
		}
		ids[ch.ID] = true
		switch id, _, ok := conceptAt(ch.Concept); {
		case ch.Project && ch.Concept != "":
			t.Errorf("%s: a project joins a whole stage, so it names no concept", ch.ID)
		case !ch.Project && (!ok || id != ch.Stage):
			t.Errorf("%s: concept %q is not in Stage %d", ch.ID, ch.Concept, ch.Stage)
		}
		if !strings.HasSuffix(ch.File, ".c") || strings.ContainsAny(ch.File, "/\\") {
			t.Errorf("%s: file %q must be a plain .c name", ch.ID, ch.File)
		}
		if ch.Title == "" || ch.Task == "" || ch.Starter == "" || len(ch.Cases) == 0 || len(ch.Hints) == 0 {
			t.Errorf("%s: title, task, starter, cases and hints are all required", ch.ID)
		}
		names := map[string]bool{}
		for _, c := range ch.Cases {
			c = c.resolved()
			if c.Name == "" || names[c.Name] {
				t.Errorf("%s: missing or duplicate case name %q", ch.ID, c.Name)
			}
			names[c.Name] = true
			if c.WantExit != 0 && c.Want != "" {
				t.Errorf("%s/%s: an error case should expect no standard output", ch.ID, c.Name)
			}
		}
		if ch.Harness != "" && strings.Contains(ch.Starter, "main(") {
			t.Errorf("%s: the checker supplies main, so the starter must not have one", ch.ID)
		}
	}
	total := 0
	for _, id := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 17} {
		if len(checksFor(id)) == 0 {
			t.Errorf("Stage %d has no checked exercises", id)
		}
		total += len(checksFor(id))
	}
	if _, ok := checkByID("hello"); !ok || total != len(checks) {
		t.Error("lookup helpers")
	}
}

func TestOutputDiff(t *testing.T) {
	for _, tc := range []struct{ want, got, problem string }{
		{"a\nb\n", "a\nb\n", ""},
		{"a\nb\n", "a  \r\nb\n\n\n", ""}, // invisible differences are ignored
		{"a\nb\n", "a\nc\n", `line 2 should be "b", but was "c"`},
		{"a\nb\n", "a\n", `output stopped early: line 2 should be "b"`},
		{"a\n", "a\nb\n", `extra output from line 2: "b"`},
		{"", "oops\n", `printed "oops", but nothing was expected on standard output`},
	} {
		if p := outputDiff(tc.want, tc.got); p != tc.problem {
			t.Errorf("outputDiff(%q, %q) = %q, want %q", tc.want, tc.got, p, tc.problem)
		}
	}
}

func TestCheckLogAndStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), checkLogFileName)
	if ev, err := readCheckLog(path); err != nil || ev != nil {
		t.Fatalf("missing log: %v %v", ev, err)
	}
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	for _, e := range []CheckEvent{
		{ID: "swap", At: now, Kind: checkEventRun, Built: true, Passed: 3, Total: 5},
		{ID: "swap", At: now.Add(time.Minute), Kind: checkEventHint, Hint: 1},
		{ID: "swap", At: now.Add(2 * time.Minute), Kind: checkEventRun, Built: true, Passed: 5, Total: 5, OK: true, HintsSeen: 1},
		{ID: "hello", At: now, Kind: checkEventRun, Built: true, Passed: 1, Total: 1, OK: true},
	} {
		if err := appendCheckEvent(path, e); err != nil {
			t.Fatal(err)
		}
	}
	// A torn last line (a crash mid-write) is skipped, not fatal.
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString(`{"id":"swap","kind":"ru`)
	f.Close()

	events, err := readCheckLog(path)
	if err != nil || len(events) != 4 {
		t.Fatalf("read %d events, err %v", len(events), err)
	}
	st := checkStatuses(events)
	if s := st["swap"]; s.Runs != 2 || s.Best != 5 || !s.Passed || s.OnOwn || s.Hints != 1 || !s.LastOK {
		t.Errorf("swap status = %+v (passed with a hint: not on its own)", s)
	}
	if s := st["hello"]; !s.Passed || !s.OnOwn {
		t.Errorf("hello status = %+v (passed with no hints)", s)
	}
}

func TestEnsureStarterNeverOverwrites(t *testing.T) {
	ch, _ := checkByID("fix_average")
	path := filepath.Join(t.TempDir(), "Stage 01", ch.File)
	if created, err := ensureStarter(path, ch); err != nil || !created {
		t.Fatalf("first call: created=%v err=%v", created, err)
	}
	os.WriteFile(path, []byte("my work"), 0o644)
	if created, err := ensureStarter(path, ch); err != nil || created {
		t.Fatalf("second call: created=%v err=%v", created, err)
	}
	if b, _ := os.ReadFile(path); string(b) != "my work" {
		t.Errorf("the learner's file was overwritten: %q", b)
	}
}

func TestLabsDir(t *testing.T) {
	docs := func() (string, error) { return "/home/ana/Documents", nil }
	env := map[string]string{}
	get := func(k string) string { return env[k] }
	if d := labsDirFor("/data/reg.json", get, docs); d != filepath.Join("/home/ana/Documents", labsFolderName) {
		t.Errorf("default: %s", d)
	}
	env["ACADEMY_HOME"] = "/srv/academy"
	if d := labsDirFor("/data/reg.json", get, docs); d != filepath.Join("/srv/academy", labsFolderName) {
		t.Errorf("ACADEMY_HOME: %s", d)
	}
	env["ACADEMY_LABS"] = "/tmp/labs"
	if d := labsDirFor("/data/reg.json", get, docs); d != "/tmp/labs" {
		t.Errorf("ACADEMY_LABS: %s", d)
	}
}

// referenceSolutions pass every case of their check.
var referenceSolutions = map[string]string{
	"linkedlist": `#include <stddef.h>
#include <stdlib.h>

struct node {
    int value;
    struct node *next;
};

struct node *push_front(struct node *head, int value) {
    struct node *n = malloc(sizeof *n);
    if (n == NULL) return NULL;
    n->value = value;
    n->next = head;
    return n;
}

struct node *reverse(struct node *head) {
    struct node *prev = NULL;
    while (head != NULL) {
        struct node *next = head->next;
        head->next = prev;
        prev = head;
        head = next;
    }
    return prev;
}

size_t list_length(const struct node *head) {
    size_t n = 0;
    for (; head != NULL; head = head->next) n++;
    return n;
}

void free_list(struct node *head) {
    while (head != NULL) {
        struct node *next = head->next;
        free(head);
        head = next;
    }
}
`,
	"hashmap": `#include <stdint.h>
#include <stdlib.h>
#include <string.h>

struct entry {
    char *key;
    int value;
    struct entry *next;
};

struct map {
    struct entry **buckets;
    size_t nbuckets, size;
};

static uint32_t hash(const char *s) {
    uint32_t h = 2166136261u;
    for (; *s; s++) {
        h ^= (unsigned char)*s;
        h *= 16777619u;
    }
    return h;
}

struct map *map_new(void) {
    struct map *m = malloc(sizeof *m);
    if (m == NULL) return NULL;
    m->nbuckets = 16;
    m->size = 0;
    m->buckets = calloc(m->nbuckets, sizeof *m->buckets);
    if (m->buckets == NULL) {
        free(m);
        return NULL;
    }
    return m;
}

static struct entry **find(const struct map *m, const char *key) {
    struct entry **e = &m->buckets[hash(key) % m->nbuckets];
    while (*e != NULL && strcmp((*e)->key, key) != 0) e = &(*e)->next;
    return e;
}

static int grow(struct map *m) {
    size_t n = m->nbuckets * 2;
    struct entry **b = calloc(n, sizeof *b);
    if (b == NULL) return -1;
    for (size_t i = 0; i < m->nbuckets; i++) {
        struct entry *e = m->buckets[i];
        while (e != NULL) {
            struct entry *next = e->next;
            size_t j = hash(e->key) % n;
            e->next = b[j];
            b[j] = e;
            e = next;
        }
    }
    free(m->buckets);
    m->buckets = b;
    m->nbuckets = n;
    return 0;
}

int map_put(struct map *m, const char *key, int value) {
    struct entry **e = find(m, key);
    if (*e != NULL) {
        (*e)->value = value;
        return 0;
    }
    if ((m->size + 1) * 4 > m->nbuckets * 3) {
        if (grow(m) != 0) return -1;
        e = find(m, key);
    }
    struct entry *n = malloc(sizeof *n);
    size_t len = strlen(key) + 1;
    char *k = malloc(len);
    if (n == NULL || k == NULL) {
        free(n);
        free(k);
        return -1;
    }
    memcpy(k, key, len);
    n->key = k;
    n->value = value;
    n->next = NULL;
    *e = n;
    m->size++;
    return 0;
}

int map_get(const struct map *m, const char *key, int *value) {
    struct entry **e = find(m, key);
    if (*e == NULL) return 0;
    *value = (*e)->value;
    return 1;
}

int map_remove(struct map *m, const char *key) {
    struct entry **e = find(m, key);
    if (*e == NULL) return 0;
    struct entry *dead = *e;
    *e = dead->next;
    free(dead->key);
    free(dead);
    m->size--;
    return 1;
}

size_t map_size(const struct map *m) { return m->size; }

void map_free(struct map *m) {
    for (size_t i = 0; i < m->nbuckets; i++) {
        struct entry *e = m->buckets[i];
        while (e != NULL) {
            struct entry *next = e->next;
            free(e->key);
            free(e);
            e = next;
        }
    }
    free(m->buckets);
    free(m);
}
`,
	"heap_sort": `#include <stddef.h>

static void sift_down(int *a, size_t n, size_t i) {
    for (;;) {
        size_t largest = i, l = 2 * i + 1, r = 2 * i + 2;
        if (l < n && a[l] > a[largest]) largest = l;
        if (r < n && a[r] > a[largest]) largest = r;
        if (largest == i) return;
        int t = a[i];
        a[i] = a[largest];
        a[largest] = t;
        i = largest;
    }
}

void heap_sort(int *a, size_t n) {
    if (n < 2) return;
    for (size_t i = n / 2; i-- > 0;) sift_down(a, n, i);
    for (size_t end = n - 1; end > 0; end--) {
        int t = a[0];
        a[0] = a[end];
        a[end] = t;
        sift_down(a, end, 0);
    }
}
`,
	"bfs": `#include <stdio.h>
#include <string.h>

static char grid[1000][1002];
static int dist[1000 * 1000];
static int queue[1000 * 1000];

int main(void) {
    int rows = 0, cols = 0, start = -1, goal = -1;
    while (rows < 1000 && fgets(grid[rows], sizeof grid[rows], stdin) != NULL) {
        size_t len = strcspn(grid[rows], "\n");
        grid[rows][len] = '\0';
        if (len == 0) break;
        cols = (int)len;
        rows++;
    }
    for (int r = 0; r < rows; r++) {
        for (int c = 0; c < cols; c++) {
            if (grid[r][c] == 'S') start = r * cols + c;
            if (grid[r][c] == 'E') goal = r * cols + c;
        }
    }
    if (start < 0 || goal < 0) {
        printf("-1\n");
        return 0;
    }
    memset(dist, -1, sizeof(int) * (size_t)(rows * cols));
    int head = 0, tail = 0;
    dist[start] = 0;
    queue[tail++] = start;
    static const int dr[] = {1, -1, 0, 0}, dc[] = {0, 0, 1, -1};
    while (head < tail) {
        int cur = queue[head++];
        if (cur == goal) {
            printf("%d\n", dist[cur]);
            return 0;
        }
        int r = cur / cols, c = cur % cols;
        for (int k = 0; k < 4; k++) {
            int nr = r + dr[k], nc = c + dc[k];
            if (nr < 0 || nc < 0 || nr >= rows || nc >= cols) continue;
            if (grid[nr][nc] == '#' || dist[nr * cols + nc] >= 0) continue;
            dist[nr * cols + nc] = dist[cur] + 1;
            queue[tail++] = nr * cols + nc;
        }
    }
    printf("-1\n");
    return 0;
}
`,
	"atomic_save": `#define _POSIX_C_SOURCE 200809L
#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

int save_file(const char *path, const char *data, size_t len) {
    char tmp[4096], dir[4096];
    if (snprintf(tmp, sizeof tmp, "%s.tmp", path) >= (int)sizeof tmp) return -1;
    const char *slash = strrchr(path, '/');
    if (slash == NULL) {
        snprintf(dir, sizeof dir, ".");
    } else {
        snprintf(dir, sizeof dir, "%.*s", (int)(slash - path), path);
    }
    int fd = open(tmp, O_WRONLY | O_CREAT | O_TRUNC, 0644);
    if (fd < 0) return -1;
    size_t off = 0;
    while (off < len) {
        ssize_t w = write(fd, data + off, len - off);
        if (w < 0) {
            close(fd);
            unlink(tmp);
            return -1;
        }
        off += (size_t)w;
    }
    if (fsync(fd) != 0 || close(fd) != 0) {
        unlink(tmp);
        return -1;
    }
    if (rename(tmp, path) != 0) {
        unlink(tmp);
        return -1;
    }
    int dfd = open(dir, O_RDONLY);
    if (dfd < 0) return -1;
    int r = fsync(dfd);
    close(dfd);
    return r == 0 ? 0 : -1;
}
`,
	"wal_kv": `#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#define MAXKEYS 1000

static char keys[MAXKEYS][64], vals[MAXKEYS][64];
static int used[MAXKEYS], n;

static int find(const char *k) {
    for (int i = 0; i < n; i++)
        if (used[i] && strcmp(keys[i], k) == 0) return i;
    return -1;
}

static void put(const char *k, const char *v) {
    int i = find(k);
    if (i < 0) {
        if (n == MAXKEYS) return;
        i = n++;
        snprintf(keys[i], sizeof keys[i], "%s", k);
        used[i] = 1;
    }
    snprintf(vals[i], sizeof vals[i], "%s", v);
}

static void del(const char *k) {
    int i = find(k);
    if (i >= 0) used[i] = 0;
}

static int append(FILE *log, const char *line) {
    if (fputs(line, log) == EOF || fflush(log) != 0) return -1;
    return fsync(fileno(log));
}

int main(void) {
    FILE *log = fopen("data.log", "a+");
    if (log == NULL) {
        perror("data.log");
        return 1;
    }
    rewind(log);
    char line[256], cmd[64], k[64], v[64], extra;
    long good = 0;
    while (fgets(line, sizeof line, log) != NULL) {
        if (strchr(line, '\n') == NULL) break; /* torn */
        good = ftell(log);
        int f = sscanf(line, "%63s %63s %63s", cmd, k, v);
        if (f == 3 && strcmp(cmd, "set") == 0) put(k, v);
        else if (f == 2 && strcmp(cmd, "del") == 0) del(k);
    }
    if (ftruncate(fileno(log), good) != 0) {
        perror("ftruncate");
        return 1;
    }
    fseek(log, 0, SEEK_END);
    while (fgets(line, sizeof line, stdin) != NULL) {
        int f = sscanf(line, "%63s %63s %63s %c", cmd, k, v, &extra);
        char rec[200];
        if (f == 3 && strcmp(cmd, "set") == 0) {
            snprintf(rec, sizeof rec, "set %s %s\n", k, v);
            if (append(log, rec) != 0) { perror("data.log"); return 1; }
            put(k, v);
            printf("ok\n");
        } else if (f == 2 && strcmp(cmd, "get") == 0) {
            int i = find(k);
            printf("%s\n", i < 0 ? "(nil)" : vals[i]);
        } else if (f == 2 && strcmp(cmd, "del") == 0) {
            snprintf(rec, sizeof rec, "del %s\n", k);
            if (append(log, rec) != 0) { perror("data.log"); return 1; }
            del(k);
            printf("ok\n");
        } else {
            printf("error\n");
        }
    }
    fclose(log);
    return 0;
}
`,
	"safe_join": `#include <stddef.h>
#include <string.h>

int safe_join(const char *base, const char *name, char *out, size_t size) {
    size_t blen = strlen(base), len = blen;
    size_t starts[4096];
    int depth = 0;
    if (name[0] == '/' || size == 0 || blen + 1 > size) return -1;
    memcpy(out, base, blen);
    const char *p = name;
    while (*p != '\0') {
        const char *end = strchr(p, '/');
        size_t clen = end ? (size_t)(end - p) : strlen(p);
        if (clen == 0 || (clen == 1 && p[0] == '.')) {
            /* empty or ".": skip */
        } else if (clen == 2 && p[0] == '.' && p[1] == '.') {
            if (depth == 0) return -1;
            len = starts[--depth];
        } else {
            if (depth == 4096 || len + 1 + clen + 1 > size) return -1;
            starts[depth++] = len;
            out[len] = '/';
            memcpy(out + len + 1, p, clen);
            len += 1 + clen;
        }
        p += clen;
        if (*p == '/') p++;
    }
    if (depth == 0) return -1;
    out[len] = '\0';
    return 0;
}
`,
	"copy_string": `#include <stddef.h>

size_t copy_string(char *dst, size_t size, const char *src) {
    size_t n = 0;
    while (src[n] != '\0') {
        if (n + 1 < size) dst[n] = src[n];
        n++;
    }
    if (size > 0) dst[n < size ? n : size - 1] = '\0';
    return n;
}
`,
	"bits": `#include <stdint.h>
#include <string.h>

int popcount(unsigned x) {
    int n = 0;
    while (x != 0) {
        x &= x - 1;
        n++;
    }
    return n;
}

unsigned rotate_left(unsigned x, int n) {
    if (n == 0) {
        return x;
    }
    return (x << n) | (x >> (32 - n));
}

int is_power_of_two(unsigned x) { return x != 0 && (x & (x - 1)) == 0; }

void float_parts(float f, int *sign, int *exponent, unsigned *mantissa) {
    uint32_t bits;
    memcpy(&bits, &f, sizeof bits);
    *sign = (int)(bits >> 31);
    *exponent = (int)((bits >> 23) & 0xFF);
    *mantissa = bits & 0x7FFFFF;
}
`,
	"calc": `#include <ctype.h>
#include <stdio.h>
#include <string.h>

static const char *p;
static int bad;

static void skip(void) {
    while (*p == ' ' || *p == '\t') p++;
}

static long long expr(void);

static long long factor(void) {
    skip();
    if (*p == '-') {
        p++;
        return -factor();
    }
    if (*p == '(') {
        p++;
        long long v = expr();
        skip();
        if (*p != ')') {
            bad = 1;
            return 0;
        }
        p++;
        return v;
    }
    if (!isdigit((unsigned char)*p)) {
        bad = 1;
        return 0;
    }
    long long v = 0;
    while (isdigit((unsigned char)*p)) v = v * 10 + (*p++ - '0');
    return v;
}

static long long term(void) {
    long long v = factor();
    for (;;) {
        skip();
        char op = *p;
        if (bad || (op != '*' && op != '/')) return v;
        p++;
        long long r = factor();
        if (op == '*') {
            v *= r;
        } else if (r == 0) {
            bad = 1;
            return 0;
        } else {
            v /= r;
        }
    }
}

static long long expr(void) {
    long long v = term();
    for (;;) {
        skip();
        char op = *p;
        if (bad || (op != '+' && op != '-')) return v;
        p++;
        long long r = term();
        v = op == '+' ? v + r : v - r;
    }
}

int main(void) {
    char line[8192];
    while (fgets(line, sizeof line, stdin) != NULL) {
        line[strcspn(line, "\n")] = '\0';
        p = line;
        bad = 0;
        long long v = expr();
        skip();
        if (bad || *p != '\0') {
            printf("error\n");
        } else {
            printf("%lld\n", v);
        }
    }
    return 0;
}
`,
	"run": `#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <sys/wait.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "usage: run PROGRAM ARGS...\n");
        return 2;
    }
    fflush(stdout);
    pid_t pid = fork();
    if (pid < 0) {
        perror("fork");
        return 1;
    }
    if (pid == 0) {
        execvp(argv[1], &argv[1]);
        perror(argv[1]);
        _exit(127);
    }
    int status;
    if (waitpid(pid, &status, 0) < 0) {
        perror("waitpid");
        return 1;
    }
    if (WIFEXITED(status)) {
        printf("exit %d\n", WEXITSTATUS(status));
    } else if (WIFSIGNALED(status)) {
        printf("signal %d\n", WTERMSIG(status));
    }
    return 0;
}
`,
	"rawcopy": `#define _POSIX_C_SOURCE 200809L
#include <errno.h>
#include <stdio.h>
#include <unistd.h>

int main(void) {
    char buf[8192];
    for (;;) {
        ssize_t n = read(0, buf, sizeof buf);
        if (n == 0) return 0;
        if (n < 0) {
            if (errno == EINTR) continue;
            perror("read");
            return 1;
        }
        ssize_t off = 0;
        while (off < n) {
            ssize_t w = write(1, buf + off, (size_t)(n - off));
            if (w < 0) {
                if (errno == EINTR) continue;
                perror("write");
                return 1;
            }
            off += w;
        }
    }
}
`,
	"counter": `#define _POSIX_C_SOURCE 200809L
#include <pthread.h>
#include <stdlib.h>

struct shared {
    pthread_mutex_t lock;
    long count;
    long per_thread;
};

static void *worker(void *arg) {
    struct shared *s = arg;
    for (long i = 0; i < s->per_thread; i++) {
        pthread_mutex_lock(&s->lock);
        s->count++;
        pthread_mutex_unlock(&s->lock);
    }
    return NULL;
}

long count_parallel(int threads, long per_thread) {
    struct shared s = {PTHREAD_MUTEX_INITIALIZER, 0, per_thread};
    if (threads <= 0) return 0;
    pthread_t *t = malloc(sizeof *t * (size_t)threads);
    if (t == NULL) return -1;
    int started = 0;
    for (; started < threads; started++) {
        if (pthread_create(&t[started], NULL, worker, &s) != 0) break;
    }
    for (int i = 0; i < started; i++) pthread_join(t[i], NULL);
    free(t);
    return s.count;
}
`,
	"upper_server": `#define _POSIX_C_SOURCE 200809L
#include <unistd.h>

void serve(int fd) {
    char buf[4096];
    ssize_t n;
    while ((n = read(fd, buf, sizeof buf)) > 0) {
        for (ssize_t i = 0; i < n; i++) {
            if (buf[i] >= 'a' && buf[i] <= 'z') buf[i] = (char)(buf[i] - 'a' + 'A');
        }
        ssize_t off = 0;
        while (off < n) {
            ssize_t w = write(fd, buf + off, (size_t)(n - off));
            if (w <= 0) return;
            off += w;
        }
    }
}
`,
	"http_request": `#include <ctype.h>
#include <stdio.h>
#include <string.h>
#include <strings.h>

static int fail(const char *why) {
    fprintf(stderr, "bad request: %s\n", why);
    return 1;
}

static int read_line(char *buf, size_t size) {
    if (fgets(buf, (int)size, stdin) == NULL) return 0;
    size_t n = strlen(buf);
    if (n == 0 || buf[n - 1] != '\n') return -1; /* too long, or no newline */
    buf[--n] = '\0';
    if (n > 0 && buf[n - 1] == '\r') buf[--n] = '\0';
    return 1;
}

int main(void) {
    static char line[8192], method[16], path[8000], version[16], host[8000];
    int r = read_line(line, sizeof line);
    if (r <= 0) return fail("no request line");
    char extra;
    if (sscanf(line, "%15s %7999s %15s %c", method, path, version, &extra) != 3) return fail("request line");
    for (char *m = method; *m; m++) {
        if (!isupper((unsigned char)*m)) return fail("method");
    }
    if (path[0] != '/' || strcmp(version, "HTTP/1.1") != 0) return fail("path or version");
    int have_host = 0;
    for (;;) {
        r = read_line(line, sizeof line);
        if (r <= 0) return fail("the head never ended");
        if (line[0] == '\0') break;
        char *colon = strchr(line, ':');
        if (colon == NULL) return fail("header without a colon");
        *colon = '\0';
        if (strcasecmp(line, "host") == 0) {
            char *v = colon + 1;
            while (*v == ' ' || *v == '\t') v++;
            size_t n = strlen(v);
            while (n > 0 && (v[n - 1] == ' ' || v[n - 1] == '\t')) v[--n] = '\0';
            snprintf(host, sizeof host, "%s", v);
            have_host = 1;
        }
    }
    if (!have_host) return fail("no Host header");
    printf("%s %s %s\n", method, path, host);
    return 0;
}
`,
	"types": `#include <stdio.h>

int main(void) {
    printf("%d\n", 7 / 2);
    printf("%d\n", 7 % 2);
    printf("%.1f\n", 7.0 / 2);
    printf("%d\n", (int)3.99);
    printf("%d\n", 'A' + 1);
    printf("%zu\n", sizeof(int) * 8);
    return 0;
}
`,
	"scores": `#include <stdio.h>

struct student {
    char name[32];
    int score;
};

int main(void) {
    struct student s[100];
    int n = 0;
    while (n < 100 && scanf("%31s %d", s[n].name, &s[n].score) == 2) {
        n++;
    }
    if (n == 0) {
        printf("no students\n");
        return 0;
    }
    int best = 0;
    long sum = 0;
    for (int i = 0; i < n; i++) {
        sum += s[i].score;
        if (s[i].score > s[best].score) {
            best = i;
        }
    }
    printf("best: %s %d\n", s[best].name, s[best].score);
    printf("average: %.1f\n", (double)sum / n);
    return 0;
}
`,
	"intvec": `#include <stddef.h>
#include <stdlib.h>

struct intvec {
    int *data;
    size_t len;
    size_t cap;
};

void vec_init(struct intvec *v) {
    v->data = NULL;
    v->len = 0;
    v->cap = 0;
}

int vec_push(struct intvec *v, int x) {
    if (v->len == v->cap) {
        size_t cap = v->cap ? v->cap * 2 : 4;
        int *tmp = realloc(v->data, cap * sizeof *tmp);
        if (tmp == NULL) {
            return -1;
        }
        v->data = tmp;
        v->cap = cap;
    }
    v->data[v->len++] = x;
    return 0;
}

void vec_free(struct intvec *v) {
    free(v->data);
    vec_init(v);
}
`,
	"count": `#include <ctype.h>
#include <stdio.h>

int main(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "usage: count FILE\n");
        return 2;
    }
    FILE *f = fopen(argv[1], "rb");
    if (f == NULL) {
        perror(argv[1]);
        return 1;
    }
    long lines = 0, words = 0, bytes = 0;
    int in_word = 0, c;
    while ((c = fgetc(f)) != EOF) {
        bytes++;
        if (c == '\n') {
            lines++;
        }
        if (isspace((unsigned char)c)) {
            in_word = 0;
        } else if (!in_word) {
            in_word = 1;
            words++;
        }
    }
    fclose(f);
    printf("%ld %ld %ld\n", lines, words, bytes);
    return 0;
}
`,
	"stack": `#include "stack.h"

void stack_init(struct stack *s) { s->top = 0; }

int stack_push(struct stack *s, int x) {
    if (s->top == STACK_MAX) {
        return -1;
    }
    s->items[s->top++] = x;
    return 0;
}

int stack_pop(struct stack *s, int *out) {
    if (s->top == 0) {
        return -1;
    }
    *out = s->items[--s->top];
    return 0;
}
`,
	"wordfreq": `#include <ctype.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct entry {
    char *word;
    long count;
};

static int cmp(const void *a, const void *b) {
    const struct entry *x = a, *y = b;
    if (x->count != y->count) {
        return x->count < y->count ? 1 : -1;
    }
    return strcmp(x->word, y->word);
}

static void die(void) {
    fprintf(stderr, "out of memory\n");
    exit(1);
}

int main(void) {
    struct entry *e = NULL;
    size_t n = 0, cap = 0;
    char *buf = NULL;
    size_t len = 0, bcap = 0;
    int c;
    do {
        c = getchar();
        if (c != EOF && c < 128 && isalpha(c)) {
            if (len + 1 >= bcap) {
                bcap = bcap ? bcap * 2 : 16;
                char *t = realloc(buf, bcap);
                if (t == NULL) die();
                buf = t;
            }
            buf[len++] = (char)tolower(c);
            continue;
        }
        if (len == 0) {
            continue;
        }
        buf[len] = '\0';
        size_t i = 0;
        while (i < n && strcmp(e[i].word, buf) != 0) {
            i++;
        }
        if (i == n) {
            if (n == cap) {
                cap = cap ? cap * 2 : 16;
                struct entry *t = realloc(e, cap * sizeof *t);
                if (t == NULL) die();
                e = t;
            }
            e[n].word = malloc(len + 1);
            if (e[n].word == NULL) die();
            memcpy(e[n].word, buf, len + 1);
            e[n].count = 0;
            n++;
        }
        e[i].count++;
        len = 0;
    } while (c != EOF);
    if (n > 0) {
        qsort(e, n, sizeof *e, cmp);
    }
    for (size_t i = 0; i < n && i < 10; i++) {
        printf("%ld %s\n", e[i].count, e[i].word);
    }
    for (size_t i = 0; i < n; i++) {
        free(e[i].word);
    }
    free(e);
    free(buf);
    return 0;
}
`,
	"hello": `#include <stdio.h>

int main(void) {
    printf("Hello, world!\n");
    return 0;
}
`,
	"fizzbuzz": `#include <stdio.h>

int main(void) {
    for (int i = 1; i <= 100; i++) {
        if (i % 15 == 0) {
            printf("FizzBuzz\n");
        } else if (i % 3 == 0) {
            printf("Fizz\n");
        } else if (i % 5 == 0) {
            printf("Buzz\n");
        } else {
            printf("%d\n", i);
        }
    }
    return 0;
}
`,
	"temperature": `double c_to_f(double c) { return c * 9.0 / 5.0 + 32.0; }
double f_to_c(double f) { return (f - 32.0) * 5.0 / 9.0; }
`,
	"swap": `void swap(int *a, int *b) {
    int t = *a;
    *a = *b;
    *b = t;
}
`,
	"letters": `#include <ctype.h>
#include <stdio.h>

int main(void) {
    long counts[26] = {0};
    int c;
    while ((c = getchar()) != EOF) {
        if (c < 128 && isalpha(c)) {
            counts[tolower(c) - 'a']++;
        }
    }
    for (int i = 0; i < 26; i++) {
        if (counts[i] > 0) {
            printf("%c %ld\n", 'a' + i, counts[i]);
        }
    }
    return 0;
}
`,
	"safe_number": `#include <ctype.h>
#include <errno.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>

int main(void) {
    char line[256];
    if (fgets(line, sizeof line, stdin) == NULL) {
        fprintf(stderr, "error: no input\n");
        return 1;
    }
    char *end;
    errno = 0;
    long v = strtol(line, &end, 10);
    if (end == line) {
        fprintf(stderr, "error: not a number\n");
        return 1;
    }
    while (*end != '\0' && isspace((unsigned char)*end)) {
        end++;
    }
    if (*end != '\0') {
        fprintf(stderr, "error: not a number\n");
        return 1;
    }
    if (errno == ERANGE || v < INT_MIN || v > INT_MAX) {
        fprintf(stderr, "error: out of range\n");
        return 1;
    }
    printf("ok %d\n", (int)v);
    return 0;
}
`,
	"fix_average": `#include <stdio.h>

double average(const int *a, int n) {
    int sum = 0;
    for (int i = 0; i < n; i++) {
        sum += a[i];
    }
    return (double)sum / n;
}

int main(void) {
    int scores[] = {90, 72, 85, 60, 99};
    printf("%.1f\n", average(scores, 5));
    int pair[] = {1, 2};
    printf("%.1f\n", average(pair, 2));
    return 0;
}
`,
}

// needCompiler skips runner tests where there is nothing to run them with.
func needCompiler(t *testing.T) {
	t.Helper()
	if !runnerSupported {
		t.Skip("the runner needs Linux, macOS or a BSD")
	}
	if _, err := findCompiler(os.Getenv); err != nil {
		t.Skip("no C compiler on this machine")
	}
}

// referenceExtras are the reference versions of checks' other files.
var referenceExtras = map[string]map[string]string{
	"stack": {"stack.h": `#ifndef STACK_H
#define STACK_H

#define STACK_MAX 64

struct stack {
    int items[STACK_MAX];
    int top;
};

void stack_init(struct stack *s);
int stack_push(struct stack *s, int x);
int stack_pop(struct stack *s, int *out);

#endif
`},
}

// runSource writes src as the check's file (and the reference versions of
// its other files) and runs the check.
func runSource(t *testing.T, id, src string) CheckRun {
	t.Helper()
	return runSourceWith(t, id, src, referenceExtras[id])
}

// runSourceWith is runSource with the other files given.
func runSourceWith(t *testing.T, id, src string, extras map[string]string) CheckRun {
	t.Helper()
	ch, ok := checkByID(id)
	if !ok {
		t.Fatalf("no check %q", id)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, ch.File)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, content := range extras {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run, err := runCheck(ch, path)
	if err != nil {
		t.Fatalf("%s: %v", id, err)
	}
	return run
}

func TestReferenceSolutionsPass(t *testing.T) {
	needCompiler(t)
	for _, ch := range checks {
		src, ok := referenceSolutions[ch.ID]
		if !ok {
			t.Errorf("%s has no reference solution in the tests", ch.ID)
			continue
		}
		run := runSource(t, ch.ID, src)
		if !run.OK() {
			t.Errorf("%s: reference solution failed (built=%v)\n%s", ch.ID, run.Built, run.BuildLog)
			for _, c := range run.Cases {
				if !c.Pass {
					t.Errorf("  %s: %s\n  stderr: %s", c.Case.Name, c.Problem, c.Stderr)
				}
			}
		}
		if run.Compiler == "" {
			t.Errorf("%s: compiler version not recorded", ch.ID)
		}
	}
}

// failsWith runs src and checks that the named case fails with a problem
// containing want.
func failsWith(t *testing.T, id, src, caseName, want string) CheckRun {
	t.Helper()
	run := runSource(t, id, src)
	if run.OK() {
		t.Fatalf("%s: a wrong program passed", id)
	}
	for _, c := range run.Cases {
		if c.Case.Name == caseName {
			if c.Pass || !strings.Contains(c.Problem, want) {
				t.Errorf("%s/%s: pass=%v problem=%q, want it to mention %q", id, caseName, c.Pass, c.Problem, want)
			}
			return run
		}
	}
	t.Fatalf("%s: no case %q ran (built=%v)\n%s", id, caseName, run.Built, run.BuildLog)
	return run
}

func TestWrongProgramsFailWithReasons(t *testing.T) {
	needCompiler(t)
	old := checkCaseTimeout
	checkCaseTimeout = time.Second
	defer func() { checkCaseTimeout = old }()

	// The XOR trick zeroes a variable swapped with itself.
	failsWith(t, "swap", "void swap(int *a, int *b) { *a ^= *b; *b ^= *a; *a ^= *b; }\n",
		"a variable with itself", `line 1 should be "4", but was "0"`)

	// Integer division: 9 / 5 is 1.
	failsWith(t, "temperature", "double c_to_f(double c) { return 9 / 5 * c + 32; }\ndouble f_to_c(double f) { return (f - 32) * 5.0 / 9.0; }\n",
		"c_to_f(100)", `"212.00", but was "132.00"`)

	// An endless loop is stopped by the time limit.
	failsWith(t, "hello", "int main(void) { for (;;) {} }\n", "prints the greeting", "did not finish")

	// A crash is reported as one (without sanitizers it is a signal).
	run := runSource(t, "hello", "#include <stdio.h>\nint main(void) { int *p = NULL; printf(\"%d\\n\", *p); return 0; }\n")
	if run.OK() || !run.Built || run.Cases[0].Pass {
		t.Fatal("a NULL dereference passed")
	}
	if p := run.Cases[0].Problem; !strings.Contains(p, "Sanitizer") && !strings.Contains(p, "crashed") {
		t.Errorf("NULL dereference: problem = %q", p)
	}

	// The unfixed starter fails: the sanitizer catches the read past the end.
	ch, _ := checkByID("fix_average")
	run = runSource(t, "fix_average", ch.Starter)
	if run.OK() {
		t.Fatal("the buggy starter passed")
	}
	if run.Sanitizers && !strings.Contains(run.Cases[0].Problem, "AddressSanitizer") {
		t.Errorf("starter with sanitizers: problem = %q, want an AddressSanitizer report", run.Cases[0].Problem)
	}

	// An error printed to stdout instead of stderr is caught.
	failsWith(t, "safe_number", "#include <stdio.h>\nint main(void) { printf(\"error\\n\"); return 1; }\n",
		"letters", "nothing was expected on standard output")

	// A warning is an error: the build fails and says why.
	run = runSource(t, "hello", "#include <stdio.h>\nint main(void) { int unused; printf(\"Hello, world!\\n\"); return 0; }\n")
	if run.Built || !strings.Contains(run.BuildLog, "unused") {
		t.Errorf("a warning should fail the build: built=%v log=%q", run.Built, run.BuildLog)
	}

	// A main in a harness check is explained.
	run = runSource(t, "swap", "void swap(int *a, int *b) { int t = *a; *a = *b; *b = t; }\nint main(void) { return 0; }\n")
	if run.Built || !strings.Contains(run.BuildLog, "must not define main") {
		t.Errorf("main in a harness check: built=%v log=%q", run.Built, run.BuildLog)
	}
}

// A program that starts a background child cannot outlive the check.
func TestRunnerKillsTheProcessGroup(t *testing.T) {
	needCompiler(t)
	old := checkCaseTimeout
	checkCaseTimeout = time.Second
	defer func() { checkCaseTimeout = old }()
	marker := filepath.Join(t.TempDir(), "survived")
	src := `#include <stdio.h>
#include <unistd.h>
int main(void) {
    if (fork() == 0) {
        sleep(2);
        FILE *f = fopen("` + marker + `", "w");
        if (f) fclose(f);
        return 0;
    }
    for (;;) {}
}
`
	failsWith(t, "hello", src, "prints the greeting", "did not finish")
	time.Sleep(2500 * time.Millisecond)
	if _, err := os.Stat(marker); err == nil {
		t.Error("the forked child outlived the check")
	}
}

// The program sees none of the learner's environment variables.
func TestRunnerEnvironmentIsMinimal(t *testing.T) {
	needCompiler(t)
	t.Setenv("ACADEMY_TEST_SECRET", "hunter2")
	run := runSource(t, "hello", `#include <stdio.h>
#include <stdlib.h>
int main(void) {
    printf("%s\n", getenv("ACADEMY_TEST_SECRET") ? "leaked" : "Hello, world!");
    return 0;
}
`)
	if !run.OK() {
		t.Errorf("the environment leaked into the program: %+v", run.Cases)
	}
}

func TestSanitizerReport(t *testing.T) {
	asan := `=================================================================
==2979==ERROR: AddressSanitizer: stack-buffer-overflow on address 0x7f3b7aa00054 at pc 0x5596975613b5 bp 0x7ffc703e9550 sp 0x7ffc703e9540
READ of size 4 at 0x7f3b7aa00054 thread T0
    #0 0x5596975613b4 in average /home/ana/Academy Labs/Stage 01/fix_average.c:11
    #1 0x55969756164b in main /home/ana/Academy Labs/Stage 01/fix_average.c:18
`
	if got, want := sanitizerReport(asan), "AddressSanitizer found a stack-buffer-overflow, READ of size 4, in average, fix_average.c:11"; got != want {
		t.Errorf("ASan: got %q, want %q", got, want)
	}
	ubsan := "/home/ana/Academy Labs/Stage 01/hello.c:2:33: runtime error: load of null pointer of type 'int'\n    #0 0x5653919c6244 in main\n"
	if got, want := sanitizerReport(ubsan), "UndefinedBehaviorSanitizer: load of null pointer of type 'int', at hello.c:2:33"; got != want {
		t.Errorf("UBSan: got %q, want %q", got, want)
	}
	if got := sanitizerReport("warning: something else\n"); got != "" {
		t.Errorf("no report: got %q", got)
	}
}

// The Study Hall screens: open a check, reveal a hint, run it, go back.
func TestCheckedExercisesScreen(t *testing.T) {
	needCompiler(t)
	dir := t.TempDir()
	t.Setenv("ACADEMY_LABS", filepath.Join(dir, "labs"))
	input := strings.Join([]string{
		"1",                               // hello
		"2",                               // reveal hint 1
		"1",                               // run: the starter has no main, so it does not build
		"3",                               // back
		fmt.Sprint(len(checksFor(1)) + 1), // back to the Study Hall
	}, "\n") + "\n"
	var out strings.Builder
	a := &App{reg: seedRegistry(), path: filepath.Join(dir, "reg.json"), width: 80,
		con: &console{in: bufio.NewReader(strings.NewReader(input)), out: &out}}
	if err := a.checkedExercises(1); err != nil {
		t.Fatalf("checked exercises: %v\n%s", err, out.String())
	}
	text := stripANSI(out.String())
	for _, want := range []string{"Created", "hello.c", "Hint 1", "It did not compile", "academy -check hello"} {
		if !strings.Contains(text, want) {
			t.Errorf("screen lacks %q:\n%s", want, text)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "labs", "Stage 01", "hello.c")); err != nil {
		t.Errorf("starter not created: %v", err)
	}
	events, _ := readCheckLog(checkLogPath(a.path))
	if len(events) != 2 || events[0].Kind != checkEventHint || events[1].Kind != checkEventRun || events[1].HintsSeen != 1 || events[1].Built {
		t.Errorf("log = %+v, want a hint then a failed run that saw 1 hint", events)
	}
}

func TestCheckCommand(t *testing.T) {
	needCompiler(t)
	dir := t.TempDir()
	t.Setenv("ACADEMY_LABS", filepath.Join(dir, "labs"))
	reg := filepath.Join(dir, "reg.json")
	var out strings.Builder
	if code := checkCommand(reg, "nope", &out, 80); code != 2 {
		t.Errorf("unknown ID: exit %d", code)
	}
	if code := checkCommand(reg, "swap", &out, 80); code != 1 || !strings.Contains(out.String(), "Created") {
		t.Errorf("first run should create the file and stop: exit %d\n%s", code, out.String())
	}
	file := filepath.Join(dir, "labs", "Stage 01", "swap.c")
	os.WriteFile(file, []byte(referenceSolutions["swap"]), 0o644)
	out.Reset()
	if code := checkCommand(reg, "swap", &out, 80); code != 0 || !strings.Contains(out.String(), "All 5 cases pass") {
		t.Errorf("solution: exit %d\n%s", code, out.String())
	}
	out.Reset()
	if code := checkCommand(reg, "list", &out, 80); code != 0 || !strings.Contains(out.String(), "passed on your own") {
		t.Errorf("list: exit %d\n%s", code, out.String())
	}
}

// Where the sanitizers are missing, checks still build and run without them.
func TestRunnerWithoutSanitizers(t *testing.T) {
	needCompiler(t)
	real, _ := findCompiler(os.Getenv)
	if len(real.pre) > 0 {
		t.Skip("CC is a wrapper command")
	}
	fake := filepath.Join(t.TempDir(), "nosan-cc")
	script := "#!/bin/sh\nfor a in \"$@\"; do case \"$a\" in -fsanitize=*) echo 'cc: error: unsupported option' >&2; exit 1;; esac; done\nexec " + real.path + " \"$@\"\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CC", fake)
	run := runSource(t, "swap", referenceSolutions["swap"])
	if run.Sanitizers || !run.OK() {
		t.Errorf("without sanitizers: sanitizers=%v ok=%v\n%s", run.Sanitizers, run.OK(), run.BuildLog)
	}
	var out strings.Builder
	writeCheckRun(&out, run, 80)
	if !strings.Contains(out.String(), "no memory checks") {
		t.Errorf("the report should say the memory checks were off:\n%s", out.String())
	}
}

func TestFindCompilerHonoursCC(t *testing.T) {
	needCompiler(t)
	env := map[string]string{"CC": "env gcc"}
	cc, err := findCompiler(func(k string) string { return env[k] })
	if err != nil || filepath.Base(cc.path) != "env" || len(cc.pre) != 1 || cc.pre[0] != "gcc" {
		t.Errorf("CC=%q: got %+v, %v", env["CC"], cc, err)
	}
	env["CC"] = "no-such-compiler-here"
	if cc, err := findCompiler(func(k string) string { return env[k] }); err != nil || len(cc.pre) != 0 {
		t.Errorf("a CC that is not installed should fall back to cc: %+v, %v", cc, err)
	}
}

func TestWrapPath(t *testing.T) {
	p := "/tmp/academy-labs-ootp9n8d/Stage 01/fix_average.c"
	for _, w := range []int{10, 20, 34, 80} {
		lines := wrapPath(p, w)
		if strings.Join(lines, "") != p {
			t.Errorf("width %d: %q does not rejoin to the path", w, lines)
		}
		for _, l := range lines {
			if visibleLen(l) > w {
				t.Errorf("width %d: line %q is too wide", w, l)
			}
		}
	}
	if got := wrapPath(p, 34); len(got) != 2 || got[0] != "/tmp/academy-labs-ootp9n8d/" {
		t.Errorf("should break after a separator: %q", got)
	}
}

// The new checks catch the mistakes they were written for.
func TestNewChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)

	// >= instead of >: the last of the tied students wins.
	tie := strings.Replace(referenceSolutions["scores"], "s[i].score > s[best].score", "s[i].score >= s[best].score", 1)
	failsWith(t, "scores", tie, "a tie goes to the first", `"best: ann 80", but was "best: ben 80"`)

	// No free in vec_free: the leak checker reports it, where it works.
	noFree := strings.Replace(referenceSolutions["intvec"], "    free(v->data);\n", "", 1)
	if cc, _ := findCompiler(os.Getenv); sanitizersWork(cc) && leaksWork(cc) {
		failsWith(t, "intvec", noFree, "frees everything", "LeakSanitizer found memory that was never freed")
	} else {
		t.Log("the leak checker does not work here: skipping the leak case")
	}

	// A header without an include guard, included twice by the checker.
	noGuard := strings.NewReplacer("#ifndef STACK_H\n", "", "#define STACK_H\n", "", "#endif\n", "").Replace(referenceExtras["stack"]["stack.h"])
	run := runSourceWith(t, "stack", referenceSolutions["stack"], map[string]string{"stack.h": noGuard})
	if run.Built || !strings.Contains(run.BuildLog, "include guard") {
		t.Errorf("a header without a guard: built=%v, log should explain the guard:\n%s", run.Built, run.BuildLog)
	}

	// A missing stack.h is reported as a missing file, not a crash.
	ch, _ := checkByID("stack")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "stack.c"), []byte(referenceSolutions["stack"]), 0o644)
	if _, err := runCheck(ch, filepath.Join(dir, "stack.c")); err == nil || !strings.Contains(err.Error(), "stack.h") {
		t.Errorf("missing stack.h: err = %v", err)
	}

	// fgets and strlen stop at a zero byte.
	fgetsCount := `#include <ctype.h>
#include <stdio.h>
#include <string.h>

int main(int argc, char **argv) {
    if (argc < 2) { fprintf(stderr, "usage\n"); return 2; }
    FILE *f = fopen(argv[1], "r");
    if (f == NULL) { perror(argv[1]); return 1; }
    char line[4096];
    long lines = 0, words = 0, bytes = 0;
    int in = 0;
    while (fgets(line, sizeof line, f) != NULL) {
        size_t n = strlen(line);
        bytes += (long)n;
        for (size_t i = 0; i < n; i++) {
            if (line[i] == '\n') lines++;
            if (isspace((unsigned char)line[i])) in = 0;
            else if (!in) { in = 1; words++; }
        }
    }
    fclose(f);
    printf("%ld %ld %ld\n", lines, words, bytes);
    return 0;
}
`
	failsWith(t, "count", fgetsCount, "zero bytes", `"1 2 5"`)

	// A fixed-size word buffer overflows on a 5000-letter word.
	fixed := `#include <ctype.h>
#include <stdio.h>
#include <string.h>

int main(void) {
    char word[64];
    int len = 0, c;
    char seen[100][64];
    long count[100];
    int n = 0;
    do {
        c = getchar();
        if (c != EOF && c < 128 && isalpha(c)) { word[len++] = (char)tolower(c); continue; }
        if (len == 0) continue;
        word[len] = '\0';
        int i = 0;
        while (i < n && strcmp(seen[i], word) != 0) i++;
        if (i == n) { strcpy(seen[n], word); count[n] = 0; n++; }
        count[i]++;
        len = 0;
    } while (c != EOF);
    for (int i = 0; i < n && i < 10; i++) printf("%ld %s\n", count[i], seen[i]);
    return 0;
}
`
	// With the sanitizers it is reported as a bug at its line; without
	// them (a compiler lacking their runtime) it can only crash or misbehave.
	run = runSource(t, "wordfreq", fixed)
	for _, c := range run.Cases {
		if c.Case.Name == "a 5000-letter word" {
			if c.Pass || (run.Sanitizers && !strings.Contains(c.Problem, "Sanitizer")) {
				t.Errorf("fixed word buffer: pass=%v sanitizers=%v problem=%q", c.Pass, run.Sanitizers, c.Problem)
			}
		}
	}
}

func TestLeakSummary(t *testing.T) {
	report := `=================================================================
==456==ERROR: LeakSanitizer: detected memory leaks

Direct leak of 16 byte(s) in 1 object(s) allocated from:
    #0 0x7f4a760fd9c7 in malloc ../../../../src/libsanitizer/asan/asan_malloc_linux.cpp:69
    #1 0x7f4a760fd9c8 in realloc ../../../../src/libsanitizer/asan/asan_malloc_linux.cpp:85
    #2 0x55e92009b19e in vec_push /home/ana/Academy Labs/Stage 01/intvec.c:21
    #3 0x55e92009b1ff in main /tmp/academy-check-1/check_main.c:17

SUMMARY: AddressSanitizer: 16 byte(s) leaked in 1 allocation(s).
`
	want := "LeakSanitizer found memory that was never freed: 16 byte(s) leaked in 1 allocation(s), allocated in vec_push, intvec.c:21"
	if got := sanitizerReport(report); got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

// The systems checks catch the mistakes they were written for. (A missing
// mutex is left out: a data race loses updates only most of the time.)
func TestSystemsChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := referenceSolutions

	// Raw wait status instead of WEXITSTATUS: 3 << 8 is 768.
	raw := strings.Replace(ref["run"], `printf("exit %d\n", WEXITSTATUS(status));`, `printf("exit %d\n", status);`, 1)
	failsWith(t, "run", raw, "exit 3", `"exit 3", but was "exit 768"`)

	// Recursing on the right makes - and / group from the right.
	rightAssoc := strings.Replace(ref["calc"], "static long long expr(void) {\n    long long v = term();\n    for (;;) {\n        skip();\n        char op = *p;\n        if (bad || (op != '+' && op != '-')) return v;\n        p++;\n        long long r = term();\n        v = op == '+' ? v + r : v - r;\n    }\n}",
		"static long long expr(void) {\n    long long v = term();\n    skip();\n    char op = *p;\n    if (bad || (op != '+' && op != '-')) return v;\n    p++;\n    long long r = expr();\n    return op == '+' ? v + r : v - r;\n}", 1)
	if rightAssoc == ref["calc"] {
		t.Fatal("calc right-recursion mutation did not apply")
	}
	failsWith(t, "calc", rightAssoc, "left to right", `"3", but was "9"`)

	// Writing the whole buffer instead of the bytes read.
	whole := strings.Replace(ref["rawcopy"], "write(1, buf + off, (size_t)(n - off))", "write(1, buf + off, sizeof buf - (size_t)off)", 1)
	failsWith(t, "rawcopy", whole, "a line of text", "")

	// One read is not the whole stream.
	once := strings.Replace(ref["upper_server"], "while ((n = read(fd, buf, sizeof buf)) > 0) {", "if ((n = read(fd, buf, sizeof buf)) > 0) {", 1)
	failsWith(t, "upper_server", once, "300 KB in pieces", `"300000 ok"`)

	// Header names are case-insensitive.
	exact := strings.Replace(ref["http_request"], `strcasecmp(line, "host")`, `strcmp(line, "Host")`, 1)
	failsWith(t, "http_request", exact, "Host in any case, with spaces", "")

	// Shifting by 32 is undefined: the sanitizer says so, where it is on.
	shift := strings.Replace(ref["bits"], "    if (n == 0) {\n        return x;\n    }\n", "", 1)
	run := runSource(t, "bits", shift)
	if run.Sanitizers {
		for _, c := range run.Cases {
			if c.Case.Name == "rotate_left, including by 0" && (c.Pass || !strings.Contains(c.Problem, "shift")) {
				t.Errorf("shift by 32: pass=%v problem=%q", c.Pass, c.Problem)
			}
		}
	}
}

// The Stage 2, 7 and 17 checks catch the mistakes they were written for.
func TestCoreChecksCatchTypicalMistakes(t *testing.T) {
	needCompiler(t)
	ref := referenceSolutions
	mutate := func(id, from, to string) string {
		t.Helper()
		out := strings.Replace(ref[id], from, to, 1)
		if out == ref[id] {
			t.Fatalf("%s: the mutation %q did not apply", id, from)
		}
		return out
	}

	// A recursive reverse a million calls deep runs out of stack.
	recursive := mutate("linkedlist", "struct node *reverse(struct node *head) {\n    struct node *prev = NULL;\n    while (head != NULL) {\n        struct node *next = head->next;\n        head->next = prev;\n        prev = head;\n        head = next;\n    }\n    return prev;\n}",
		"struct node *reverse(struct node *head) {\n    if (head == NULL || head->next == NULL) return head;\n    struct node *rest = reverse(head->next);\n    head->next->next = head;\n    head->next = NULL;\n    return rest;\n}")
	failsWith(t, "linkedlist", recursive, "a million nodes", "")

	// A table that never grows is far too slow for 100,000 keys.
	old := checkCaseTimeout
	checkCaseTimeout = time.Second
	defer func() { checkCaseTimeout = old }()
	failsWith(t, "hashmap", mutate("hashmap", "if ((m->size + 1) * 4 > m->nbuckets * 3) {", "if (0) {"), "a hundred thousand keys", "did not finish")

	// Keeping the caller's pointer instead of a copy of the key.
	noCopy := mutate("hashmap", "    memcpy(k, key, len);\n    n->key = k;", "    free(k);\n    n->key = (char *)key;")
	noCopy = strings.Replace(noCopy, "            free(e->key);\n", "", 1)
	noCopy = strings.Replace(noCopy, "    free(dead->key);\n", "", 1)
	failsWith(t, "hashmap", noCopy, "keys are copied", "")

	// Insertion sort cannot sort a million numbers in time.
	insertion := `#include <stddef.h>

void heap_sort(int *a, size_t n) {
    for (size_t i = 1; i < n; i++) {
        int x = a[i];
        size_t j = i;
        while (j > 0 && a[j - 1] > x) {
            a[j] = a[j - 1];
            j--;
        }
        a[j] = x;
    }
}
`
	failsWith(t, "heap_sort", insertion, "a million random numbers", "did not finish")

	// Counting cells instead of steps.
	failsWith(t, "bfs", mutate("bfs", "    dist[start] = 0;", "    dist[start] = 1;"), "around a wall", `"5", but was "6"`)

	// link instead of rename leaves the temporary file behind.
	failsWith(t, "atomic_save", mutate("atomic_save", "if (rename(tmp, path) != 0) {", "if (link(tmp, path) != 0) {"), "creates the file", "2 entries")

	// Opening the log with "w" forgets everything at start.
	failsWith(t, "wal_kv", mutate("wal_kv", `fopen("data.log", "a+")`, `fopen("data.log", "w+")`), "recovery replays the log", `"3", but was "(nil)"`)

	// Not cutting off the torn record glues the next one onto it.
	failsWith(t, "wal_kv", mutate("wal_kv", "    if (ftruncate(fileno(log), good) != 0) {", "    if (0 && ftruncate(fileno(log), good) != 0) {"), "a torn last record", "in the file data.log")

	// Checking only for a leading "../" misses the rest.
	naive := `#include <stdio.h>
#include <string.h>

int safe_join(const char *base, const char *name, char *out, size_t size) {
    if (name[0] == '/' || strncmp(name, "../", 3) == 0) return -1;
    if ((size_t)snprintf(out, size, "%s/%s", base, name) >= size) return -1;
    return 0;
}
`
	failsWith(t, "safe_join", naive, "escapes are refused", "")
	failsWith(t, "safe_join", naive, "canonicalised first", "")

	// strncpy does not terminate a string it cuts short.
	strncpyCopy := `#include <stddef.h>
#include <string.h>

size_t copy_string(char *dst, size_t size, const char *src) {
    strncpy(dst, src, size);
    return strlen(src);
}
`
	failsWith(t, "copy_string", strncpyCopy, "cut short", "NOT TERMINATED")
}

// Big generated inputs are made when a case first runs (Gen), not when
// the program starts: building them all at start-up once took 2.8 s.
func TestBigCaseInputsAreGeneratedLazily(t *testing.T) {
	for _, ch := range checks {
		for _, c := range ch.Cases {
			if len(c.Stdin)+len(c.Want) > 100<<10 {
				t.Errorf("%s/%s: %d bytes built at start-up; use Gen: generated(…)", ch.ID, c.Name, len(c.Stdin)+len(c.Want))
			}
		}
	}
}

// Every starter compiles as it is, with the runner's flags, so a learner's
// first build fails only for what they have not written yet (a missing
// main or function shows up at link time), never for the starter itself:
// an unused static helper is an error under -Werror.
func TestStartersCompile(t *testing.T) {
	needCompiler(t)
	cc, _ := findCompiler(os.Getenv)
	dir := t.TempDir()
	for _, ch := range checks {
		src := filepath.Join(dir, ch.ID+".c")
		if err := os.WriteFile(src, []byte(ch.Starter), 0o644); err != nil {
			t.Fatal(err)
		}
		args := append([]string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-c", src, "-o", src + ".o"}, ch.Flags...)
		if out, err := cc.command(context.Background(), args...).CombinedOutput(); err != nil {
			t.Errorf("%s: the starter does not compile:\n%s", ch.ID, out)
		}
	}
}
