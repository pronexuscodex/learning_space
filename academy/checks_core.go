package main

// Checked exercises for data structures (Stage 2), storage (Stage 7) and
// secure code (Stage 17): the skills interviews and real systems test.

import (
	"fmt"
	"strings"
	"time"
)

var coreChecks = []Check{
	{
		ID: "linkedlist", Stage: 2, Concept: "Arrays, Linked Lists, Stacks & Queues",
		Title: "A linked list: push, reverse, free",
		Task:  "Write four functions for the struct node in the starter, and no main: push_front (a new node at the head), reverse (in place, by relinking, with a loop), list_length, and free_list. The checker builds lists of up to a million nodes, and fails if any node is never freed.",
		File:  "linkedlist.c",
		Leaks: true,
		Starter: `/* Checked exercise: linkedlist
 * Keep the struct as it is; write the four functions. No main.
 */
#include <stddef.h>
#include <stdlib.h>

struct node {
    int value;
    struct node *next;
};

/* push_front returns the new head: a node holding value, in front of
 * head. If memory runs out it returns NULL and head is unchanged. */
struct node *push_front(struct node *head, int value);

/* reverse relinks the list back to front and returns the new head. */
struct node *reverse(struct node *head);

size_t list_length(const struct node *head);

/* free_list frees every node. */
void free_list(struct node *head);
`,
		Harness: `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct node { int value; struct node *next; };
struct node *push_front(struct node *head, int value);
struct node *reverse(struct node *head);
size_t list_length(const struct node *head);
void free_list(struct node *head);

static int scrub(void) {
    volatile char junk[8192];
    for (int i = 0; i < 8192; i++) junk[i] = 0;
    return junk[7];
}

static void print(const struct node *h) {
    if (h == NULL) printf("(empty)");
    for (; h != NULL; h = h->next) printf("%d ", h->value);
    printf("\n");
}

static struct node *build(int n) {
    struct node *h = NULL;
    for (int i = n; i >= 1; i--) h = push_front(h, i);
    return h;
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    struct node *h = NULL;
    if (strcmp(argv[1], "five") == 0) {
        h = build(5);
        print(h);
    } else if (strcmp(argv[1], "reverse") == 0) {
        h = reverse(build(5));
        print(h);
    } else if (strcmp(argv[1], "edges") == 0) {
        print(reverse(NULL));
        h = reverse(build(1));
        print(h);
        printf("%zu %zu\n", list_length(NULL), list_length(h));
    } else if (strcmp(argv[1], "million") == 0) {
        h = reverse(build(1000000));
        long long sum = 0;
        int ordered = 1, expect = 1000000;
        for (const struct node *p = h; p != NULL; p = p->next) {
            sum += p->value;
            ordered &= p->value == expect--;
        }
        printf("%zu %lld %s\n", list_length(h), sum, ordered ? "ok" : "wrong order");
    } else if (strcmp(argv[1], "free") == 0) {
        free_list(build(1000));
        printf("freed%s\n", scrub() ? "?" : "");
        return 0;
    } else return 2;
    free_list(h);
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "five pushes", Args: []string{"five"}, Want: "1 2 3 4 5\n",
				Explain: "push_front puts the new node in front: new->next = head, then return new."},
			{Name: "reverse five", Args: []string{"reverse"}, Want: "5 4 3 2 1\n",
				Explain: "Walk the list with three pointers (prev, current, next): save next, point current back at prev, step on."},
			{Name: "empty and single lists", Args: []string{"edges"}, Want: "(empty)\n1\n0 1\n"},
			{Name: "a million nodes", Args: []string{"million"}, Want: "1000000 500000500000 ok\n",
				Explain: "A recursive reverse a million calls deep overflows the stack: use a loop."},
			{Name: "frees every node", Args: []string{"free"}, Want: "freed\n",
				Explain: "free_list must free each node, saving next before freeing the node that holds it."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Draw the boxes and arrows for a three-node list before you write reverse.",
			"Never use a node after freeing it: struct node *next = h->next; free(h); h = next;",
		},
	},
	{
		ID: "hashmap", Stage: 2, Concept: "Hash Tables",
		Title: "A hash map that grows",
		Task:  "Write a hash map from strings to ints: map_new, map_put (insert or replace, copying the key), map_get, map_remove, map_size and map_free. Define struct map yourself in your file. Use chaining or open addressing, and grow the table when it gets full, so a hundred thousand keys stay fast. No main: the checker calls your functions.",
		File:  "hashmap.c",
		Leaks: true,
		Starter: `/* Checked exercise: hashmap
 * Define struct map here (the checker only ever holds a pointer to it),
 * and write these functions. No main.
 *   struct map *map_new(void);                       NULL if out of memory
 *   int map_put(struct map *m, const char *key, int value);
 *       insert or replace; copy the key; 0, or -1 if out of memory
 *   int map_get(const struct map *m, const char *key, int *value);
 *       1 and *value if found, else 0
 *   int map_remove(struct map *m, const char *key);  1 if removed, else 0
 *   size_t map_size(const struct map *m);
 *   void map_free(struct map *m);                    frees everything
 * Grow the table when it is about 3/4 full, so lookups stay O(1).
 */
#include <stddef.h>
`,
		Harness: `#include <stdio.h>
#include <string.h>

struct map;
struct map *map_new(void);
int map_put(struct map *m, const char *key, int value);
int map_get(const struct map *m, const char *key, int *value);
int map_remove(struct map *m, const char *key);
size_t map_size(const struct map *m);
void map_free(struct map *m);

static int scrub(void) {
    volatile char junk[8192];
    for (int i = 0; i < 8192; i++) junk[i] = 0;
    return junk[9];
}

static void show(const struct map *m, const char *k) {
    int v = -999;
    if (map_get(m, k, &v)) printf("%s=%d ", k, v);
    else printf("%s=none ", k);
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    struct map *m = map_new();
    if (m == NULL) return 3;
    if (strcmp(argv[1], "basic") == 0) {
        map_put(m, "apple", 3);
        map_put(m, "pear", 5);
        show(m, "apple"); show(m, "pear"); show(m, "plum");
        printf("size=%zu\n", map_size(m));
    } else if (strcmp(argv[1], "replace") == 0) {
        map_put(m, "k", 1);
        map_put(m, "k", 2);
        show(m, "k");
        printf("size=%zu\n", map_size(m));
    } else if (strcmp(argv[1], "remove") == 0) {
        map_put(m, "a", 1); map_put(m, "b", 2); map_put(m, "c", 3);
        int first = map_remove(m, "b");  /* two statements: C does not fix the */
        int second = map_remove(m, "b"); /* order it evaluates arguments in */
        printf("%d %d ", first, second);
        show(m, "a"); show(m, "b"); show(m, "c");
        printf("size=%zu\n", map_size(m));
    } else if (strcmp(argv[1], "copies") == 0) {
        char key[16] = "temp";
        map_put(m, key, 7);
        strcpy(key, "XXXX");
        show(m, "temp"); show(m, "XXXX");
        printf("\n");
    } else if (strcmp(argv[1], "collisions") == 0) {
        map_put(m, "ab", 1); map_put(m, "ba", 2); map_put(m, "", 3);
        show(m, "ab"); show(m, "ba"); show(m, "");
        printf("size=%zu\n", map_size(m));
    } else if (strcmp(argv[1], "many") == 0) {
        char k[32];
        int bad = 0;
        for (int i = 0; i < 100000; i++) {
            snprintf(k, sizeof k, "key%d", i);
            bad |= map_put(m, k, i) != 0;
        }
        for (int i = 0; i < 100000; i += 2) {
            snprintf(k, sizeof k, "key%d", i);
            bad |= map_remove(m, k) != 1;
        }
        for (int i = 0; i < 100000; i++) {
            int v = -1;
            snprintf(k, sizeof k, "key%d", i);
            int found = map_get(m, k, &v);
            bad |= i % 2 ? (!found || v != i) : found;
        }
        printf("%zu %s\n", map_size(m), bad ? "wrong" : "ok");
    } else if (strcmp(argv[1], "free") == 0) {
        char k[32];
        for (int i = 0; i < 1000; i++) {
            snprintf(k, sizeof k, "w%d", i);
            map_put(m, k, i);
        }
        map_free(m);
        printf("freed%s\n", scrub() ? "?" : "");
        return 0;
    } else return 2;
    map_free(m);
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "put and get", Args: []string{"basic"}, Want: "apple=3 pear=5 plum=none size=2\n"},
			{Name: "put replaces", Args: []string{"replace"}, Want: "k=2 size=1\n",
				Explain: "Look the key up first: putting an existing key replaces its value instead of adding a second entry."},
			{Name: "remove", Args: []string{"remove"}, Want: "1 0 a=1 b=none c=3 size=2\n"},
			{Name: "keys are copied", Args: []string{"copies"}, Want: "temp=7 XXXX=none\n",
				Explain: "Store a copy of the key (malloc and memcpy): the caller may reuse its buffer afterwards."},
			{Name: "collisions and the empty key", Args: []string{"collisions"}, Want: "ab=1 ba=2 =3 size=3\n",
				Explain: "Keys that hash alike must still be told apart with strcmp, and the empty string is a key like any other."},
			{Name: "a hundred thousand keys", Args: []string{"many"}, Want: "50000 ok\n",
				Explain: "A table that never grows turns every lookup into a long walk: double the buckets and rehash when it is about 3/4 full."},
			{Name: "frees everything", Args: []string{"free"}, Want: "freed\n",
				Explain: "map_free must free every key copy, every entry and the table itself."},
		},
		Hints: []string{
			"A simple, good hash for strings is FNV-1a: start at 2166136261u and, for each byte, xor it in, then multiply by 16777619u.",
			"Chaining: an array of bucket pointers, each the head of a small linked list of entries { char *key; int value; next }.",
			"To grow, allocate a table twice the size and move every entry into its new bucket (index = hash % new_size).",
		},
	},
	{
		ID: "heap_sort", Stage: 2, Concept: "Trees, Heaps & Sorting",
		Title:   "Heap sort",
		Task:    "Write void heap_sort(int *a, size_t n), which sorts the array in place in ascending order using a binary max-heap: build the heap, then repeatedly swap the largest to the end and sift down. No main: the checker sorts small arrays, edge cases and a million random numbers within the time limit, which a quadratic sort cannot do.",
		File:    "heap_sort.c",
		Timeout: 10 * time.Second,
		Starter: `/* Checked exercise: heap_sort
 * void heap_sort(int *a, size_t n);   sort a[0..n-1] ascending, in place
 * The heap lives in the array itself: the children of i are 2i+1 and
 * 2i+2. Build a max-heap (sift down from n/2 - 1 to 0), then n - 1 times
 * swap a[0] to the end and sift the new root down. No main.
 */
#include <stddef.h>
`,
		Harness: `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void heap_sort(int *a, size_t n);

static int cmp(const void *x, const void *y) {
    int a = *(const int *)x, b = *(const int *)y;
    return (a > b) - (a < b);
}

static void show(int *a, size_t n) {
    heap_sort(a, n);
    for (size_t i = 0; i < n; i++) printf("%d ", a[i]);
    printf("\n");
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "small") == 0) {
        int a[] = {5, 2, 9, 1, 5, 6, -3, 0};
        show(a, 8);
    } else if (strcmp(argv[1], "edges") == 0) {
        heap_sort(NULL, 0);
        int one[] = {42};
        show(one, 1);
        int two[] = {2, 1};
        show(two, 2);
        int same[] = {7, 7, 7, 7};
        show(same, 4);
    } else if (strcmp(argv[1], "ordered") == 0) {
        int up[] = {1, 2, 3, 4, 5, 6}, down[] = {6, 5, 4, 3, 2, 1};
        show(up, 6);
        show(down, 6);
    } else if (strcmp(argv[1], "million") == 0) {
        size_t n = 1000000;
        int *a = malloc(n * sizeof *a), *b = malloc(n * sizeof *b);
        if (a == NULL || b == NULL) return 3;
        unsigned x = 12345;
        for (size_t i = 0; i < n; i++) {
            x = x * 1103515245u + 12345u;
            a[i] = (int)(x >> 1) - (1 << 30);
        }
        memcpy(b, a, n * sizeof *a);
        heap_sort(a, n);
        qsort(b, n, sizeof *b, cmp);
        printf("%s\n", memcmp(a, b, n * sizeof *a) == 0 ? "sorted" : "not sorted");
        free(a);
        free(b);
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "eight numbers", Args: []string{"small"}, Want: "-3 0 1 2 5 5 6 9\n"},
			{Name: "empty, one, two, all equal", Args: []string{"edges"}, Want: "42\n1 2\n7 7 7 7\n",
				Explain: "With n == 0 there is nothing to do, and the array may be NULL: n/2 - 1 on a size_t of 0 wraps around to a huge number, so return first."},
			{Name: "already sorted, and reversed", Args: []string{"ordered"}, Want: "1 2 3 4 5 6\n1 2 3 4 5 6\n",
				Explain: "Sorted and reverse-sorted input are the classic worst cases for a naive quicksort; a heap does not care."},
			{Name: "a million random numbers", Args: []string{"million"}, Want: "sorted\n",
				Explain: "Heap sort is O(n log n): about 20 million steps here. Insertion or bubble sort would need around a trillion, far past the time limit."},
		},
		Hints: []string{
			"sift_down(a, n, i): while i has a child, pick the larger child; if it is bigger than a[i], swap and continue from there, else stop.",
			"Indexes are size_t, so loops that count down need care: for (size_t i = n / 2; i-- > 0;) runs i from n/2 - 1 down to 0.",
		},
	},
	{
		ID: "bfs", Stage: 2, Concept: "Graphs, BFS/DFS & Shortest Paths",
		Title:   "Shortest path through a maze",
		Task:    "Read a maze from standard input: lines of equal length, at most 1000 lines of at most 1000 characters, where # is a wall, . is open, S is the start and E is the end. Moving up, down, left or right, print the fewest steps from S to E, or -1 if E cannot be reached (or the maze has no S or no E).",
		File:    "bfs.c",
		Timeout: 10 * time.Second,
		Starter: `/* Checked exercise: bfs
 * Read the maze (# wall, . open, S start, E end; up to 1000 x 1000) from
 * stdin and print the fewest steps from S to E, or -1.
 * Breadth-first search finds shortest paths when every step costs the
 * same: a queue of cells, a distance per cell, each cell visited once.
 */
`,
		Cases: []CheckCase{
			{Name: "next door", Stdin: "SE\n", Want: "1\n"},
			{Name: "around a wall", Stdin: "S.#.\n.##.\n...E\n", Want: "5\n",
				Explain: "Count steps, not cells: the start itself is 0 steps away."},
			{Name: "walled off", Stdin: "S#E\n", Want: "-1\n",
				Explain: "When the queue empties without reaching E, there is no path: print -1."},
			{Name: "no S at all", Stdin: "..E\n", Want: "-1\n"},
			{Name: "the shortest, not the first found", Stdin: openMaze(50, 50, 0, 0, 0, 49), Want: "49\n",
				Explain: "Depth-first search finds a path, not the shortest one: explore in rings with a queue instead."},
			{Name: "a winding corridor", Stdin: serpentine(), Want: fmt.Sprint(serpentineSteps()) + "\n"},
			{Name: "a 1000 by 1000 maze", Stdin: openMaze(1000, 1000, 0, 0, 999, 999), Want: "1998\n",
				Explain: "A recursive search a million cells deep overflows the stack, and per-cell queues of your own are too slow: use one array as a queue, with a head and a tail index."},
		},
		Hints: []string{
			"Store the grid in a static char array of 1000 x 1001, and the distances in a static int array (static: too big for the stack).",
			"The queue never holds more than every cell once: an array of rows * cols cell numbers (row * cols + col) is enough.",
			"Mark a cell's distance when you add it to the queue, not when you take it out, or the same cell is queued many times.",
		},
	},
	{
		ID: "atomic_save", Stage: 7, Concept: "Durability: Page Cache, fsync & Atomic Rename",
		Title: "Save a file atomically",
		Task:  "Write int save_file(const char *path, const char *data, size_t len): it must replace the file at path with exactly len bytes of data so that, even if the power fails half way, the file holds either the old contents or the new ones, never a mix. Write a temporary file in the same folder, fsync it, rename it over path, then fsync the folder. Return 0, or -1 on any error, leaving no temporary file behind. No main. (The checker can verify the result and that nothing is left behind, but it cannot pull the power: read your code against the four steps in the starter.)",
		File:  "atomic_save.c",
		Starter: posixDefine + `/* Checked exercise: atomic_save
 * int save_file(const char *path, const char *data, size_t len);
 *   1. write data to a temporary file IN THE SAME FOLDER as path
 *      (rename is only atomic within one file system)
 *   2. fsync it, then close it
 *   3. rename(temp, path): readers see the old file or the new, never half
 *   4. open the folder and fsync it, so the rename itself is durable
 * Return 0, or -1 on any error, removing the temporary file.
 * This academy saves your progress exactly this way (main.go).
 */
#include <fcntl.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
`,
		Harness: posixDefine + `#include <dirent.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int save_file(const char *path, const char *data, size_t len);

static int by_name(const void *a, const void *b) { return strcmp((const char *)a, (const char *)b); }

/* entries lists a folder: how many entries, and their names in order. */
static void entries(const char *dir) {
    DIR *d = opendir(dir);
    if (d == NULL) { printf("[no folder %s]\n", dir); return; }
    static char names[64][256];
    int n = 0;
    struct dirent *e;
    while ((e = readdir(d)) != NULL) {
        if (strcmp(e->d_name, ".") == 0 || strcmp(e->d_name, "..") == 0) continue;
        if (n < 64) snprintf(names[n], sizeof names[n], "%s", e->d_name);
        n++;
    }
    closedir(d);
    int shown = n < 64 ? n : 64;
    qsort(names, (size_t)shown, sizeof names[0], by_name);
    printf("%s: %d entr%s:", dir, n, n == 1 ? "y" : "ies");
    for (int i = 0; i < shown; i++) printf(" %s", names[i]);
    printf("\n");
}

static void cat(const char *path) {
    FILE *f = fopen(path, "rb");
    if (f == NULL) { printf("[no file %s]\n", path); return; }
    int c, n = 0;
    while ((c = fgetc(f)) != EOF) {
        if (c >= 32 && c < 127) putchar(c); else printf("\\x%02x", c);
        n++;
    }
    fclose(f);
    printf(" (%d bytes)\n", n);
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "create") == 0) {
        printf("%d\n", save_file("out.txt", "hello", 5));
        cat("out.txt");
        entries(".");
    } else if (strcmp(argv[1], "replace") == 0) {
        printf("%d\n", save_file("out.txt", "new", 3));
        cat("out.txt");
        entries(".");
    } else if (strcmp(argv[1], "binary") == 0) {
        printf("%d\n", save_file("bin.dat", "\0\1\0\377", 4));
        cat("bin.dat");
    } else if (strcmp(argv[1], "empty") == 0) {
        printf("%d\n", save_file("out.txt", "", 0));
        cat("out.txt");
    } else if (strcmp(argv[1], "subfolder") == 0) {
        printf("%d\n", save_file("sub/out.txt", "inside", 6));
        cat("sub/out.txt");
        entries("sub");
        entries(".");
    } else if (strcmp(argv[1], "missing") == 0) {
        printf("%d\n", save_file("no/such/folder/out.txt", "x", 1));
        entries(".");
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "creates the file", Args: []string{"create"}, Want: "0\nhello (5 bytes)\n.: 1 entry: out.txt\n",
				Explain: "Only out.txt may be left: the temporary file must be renamed away (or removed on failure)."},
			{Name: "replaces the old contents", Args: []string{"replace"}, Files: map[string]string{"out.txt": "the old, longer contents\n"}, Want: "0\nnew (3 bytes)\n.: 1 entry: out.txt\n",
				Explain: "Writing into the old file directly would leave a mix of old and new if it failed half way, and a shorter new file would keep the old tail."},
			{Name: "binary data", Args: []string{"binary"}, Want: "0\n\\x00\\x01\\x00\\xff (4 bytes)\n",
				Explain: "Write exactly len bytes with write (or fwrite): strlen and fputs stop at the first zero byte."},
			{Name: "empty data", Args: []string{"empty"}, Want: "0\n (0 bytes)\n"},
			{Name: "a file in a subfolder", Args: []string{"subfolder"}, Files: map[string]string{"sub/keep.txt": "keep me"}, Want: "0\ninside (6 bytes)\nsub: 2 entries: keep.txt out.txt\n.: 1 entry: sub\n",
				Explain: "Create the temporary file in the folder of path (sub/), not in the current folder or /tmp: rename cannot move a file atomically between folders on different file systems.",
			},
			{Name: "a folder that does not exist", Args: []string{"missing"}, Want: "-1\n.: 0 entries:\n",
				Explain: "When anything fails, return -1 and remove whatever temporary file you created."},
		},
		Hints: []string{
			"The temporary name: path plus a suffix, such as snprintf(tmp, sizeof tmp, \"%s.tmp\", path), keeps it in the same folder.",
			"open(tmp, O_WRONLY | O_CREAT | O_TRUNC, 0644), then write in a loop until all len bytes are written, then fsync(fd) and close(fd).",
			"The folder is everything before the last '/' (or \".\" if there is none): open it with O_RDONLY and fsync that descriptor.",
		},
	},
	{
		ID: "wal_kv", Stage: 7, Concept: "Write-Ahead Logging & Recovery",
		Title: "A key-value store with a write-ahead log",
		Task:  "Write a program that reads commands from standard input: set KEY VALUE, get KEY and del KEY (keys and values are single words of at most 63 characters). Before answering ok to a set or del, append the command as a line to data.log and flush it to disk (fflush, then fsync). At start, replay data.log to rebuild the store; a last line without a newline was torn by a crash: ignore it, and cut it off before appending. get prints the value or (nil); a malformed command prints error.",
		File:  "wal_kv.c",
		Starter: posixDefine + `/* Checked exercise: wal_kv
 * Commands on stdin, one per line:
 *   set KEY VALUE   append "set KEY VALUE\n" to data.log, fflush, fsync,
 *                   THEN print ok
 *   get KEY         print the value, or (nil)
 *   del KEY         append "del KEY\n" to data.log (flushed), print ok
 *   anything else   print error
 * At start: replay data.log line by line. A last line with no '\n' is a
 * torn write from a crash: ignore it, and truncate the file to the end
 * of the last complete line before appending.
 */
#include <stdio.h>
#include <string.h>
#include <unistd.h>
`,
		Cases: []CheckCase{
			{Name: "a fresh store", Stdin: "set a 1\nget a\nget b\n", Want: "ok\n1\n(nil)\n", WantFiles: map[string]string{"data.log": "set a 1\n"},
				Explain: "Every set is appended to data.log, as a line, before ok is printed."},
			{Name: "recovery replays the log", Stdin: "get a\nget b\n", Files: map[string]string{"data.log": "set a 1\nset b 2\nset a 3\ndel b\n"}, Want: "3\n(nil)\n",
				Explain: "Apply the log from the start: later records win, and del removes a key."},
			{Name: "appends, never rewrites", Stdin: "set b 2\n", Files: map[string]string{"data.log": "set a 1\n"}, Want: "ok\n", WantFiles: map[string]string{"data.log": "set a 1\nset b 2\n"},
				Explain: "Open the log for appending (\"a\"): opening it with \"w\" erases every earlier record."},
			{Name: "a torn last record", Stdin: "get b\nset c 3\nget c\n", Files: map[string]string{"data.log": "set a 1\nset b 2"}, Want: "(nil)\nok\n3\n", WantFiles: map[string]string{"data.log": "set a 1\nset c 3\n"},
				Explain: "\"set b 2\" has no newline: the crash happened while writing it, so it never happened. Truncate the file to the end of \"set a 1\\n\" (ftruncate) before appending, or the next record joins it on one line."},
			{Name: "deletes are logged too", Stdin: "del a\nget a\n", Files: map[string]string{"data.log": "set a 1\n"}, Want: "ok\n(nil)\n", WantFiles: map[string]string{"data.log": "set a 1\ndel a\n"}},
			{Name: "bad commands", Stdin: "fly away\nset onlykey\nget\nset a b c\n", Want: "error\nerror\nerror\nerror\n"},
			{Name: "five hundred keys", Stdin: manyKVCommands(), Want: strings.Repeat("ok\n", 500) + "v250\n(nil)\n"},
		},
		Hints: []string{
			"sscanf(line, \"%63s %63s %63s %c\", cmd, key, value, &extra) returns how many fields it filled: use the count to spot malformed commands.",
			"While replaying, keep track of the offset just after the last complete line; if the file is longer, ftruncate(fileno(log), offset).",
			"fflush moves the data from your buffer to the kernel; fsync(fileno(log)) moves it to the disk. ok only after both.",
		},
	},
	{
		ID: "safe_join", Stage: 17, Concept: "Handling Untrusted Input",
		Title: "Keep a file name inside its folder",
		Task:  "Write int safe_join(const char *base, const char *name, char *out, size_t size): name came from outside (a URL, an upload, a zip file). Canonicalise it first (drop empty and . components, let .. remove the previous component), then refuse it if it is absolute, climbs above base, or names no file at all. Write base/canonical-name into out and return 0, or return -1 if it is refused or does not fit in size bytes. No main.",
		File:  "safe_join.c",
		Starter: `/* Checked exercise: safe_join
 * int safe_join(const char *base, const char *name, char *out, size_t size);
 *   "report.pdf"          -> base/report.pdf
 *   "a/./b//c.txt"        -> base/a/b/c.txt      (canonicalise first)
 *   "a/../b.txt"          -> base/b.txt
 *   "../etc/passwd"       -> -1                  (climbs out of base)
 *   "/etc/passwd"         -> -1                  (absolute)
 *   "", ".", "a/.."       -> -1                  (no file named)
 * Return 0 with the result in out, or -1 (also when it does not fit in
 * size bytes, '\0' included). Never write past out[size - 1].
 */
#include <stddef.h>
#include <string.h>
`,
		Harness: `#include <stdio.h>
#include <string.h>

int safe_join(const char *base, const char *name, char *out, size_t size);

static void try(const char *name, size_t size) {
    char buf[8192 + 16];
    memset(buf, '#', sizeof buf);
    int r = safe_join("/srv/files", name, buf, size);
    int overrun = 0;
    for (size_t i = size; i < sizeof buf; i++) overrun |= buf[i] != '#';
    if (overrun) printf("WROTE PAST THE BUFFER\n");
    else if (r == 0 && memchr(buf, '\0', size) != NULL) printf("%s\n", buf);
    else if (r == 0) printf("NOT TERMINATED\n");
    else printf("refused\n");
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "ordinary") == 0) {
        try("report.pdf", 4096);
        try("2026/notes.txt", 4096);
        try("...", 4096);
    } else if (strcmp(argv[1], "canonical") == 0) {
        try("a/./b//c.txt", 4096);
        try("a/../b.txt", 4096);
        try("dir/", 4096);
    } else if (strcmp(argv[1], "escapes") == 0) {
        try("../etc/passwd", 4096);
        try("a/../../etc/passwd", 4096);
        try("a/b/../../../x", 4096);
        try("..", 4096);
    } else if (strcmp(argv[1], "absolute") == 0) {
        try("/etc/passwd", 4096);
        try("//etc/passwd", 4096);
    } else if (strcmp(argv[1], "nothing") == 0) {
        try("", 4096);
        try(".", 4096);
        try("a/..", 4096);
    } else if (strcmp(argv[1], "small") == 0) {
        try("report.pdf", 22); /* exactly enough: 21 characters and '\0' */
        try("report.pdf", 21); /* one byte short */
        try("report.pdf", 0);
    } else if (strcmp(argv[1], "long") == 0) {
        char name[6001];
        memset(name, 'n', 6000);
        name[6000] = '\0';
        char buf[8192];
        int r = safe_join("/srv/files", name, buf, sizeof buf);
        printf("%d %zu\n", r, r == 0 ? strlen(buf) : 0);
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "ordinary names", Args: []string{"ordinary"}, Want: "/srv/files/report.pdf\n/srv/files/2026/notes.txt\n/srv/files/...\n",
				Explain: "Only a component that is exactly \"..\" goes up: \"...\" is an ordinary file name."},
			{Name: "canonicalised first", Args: []string{"canonical"}, Want: "/srv/files/a/b/c.txt\n/srv/files/b.txt\n/srv/files/dir\n",
				Explain: "Drop empty and \".\" components, and let \"..\" remove the component before it, then build the result."},
			{Name: "escapes are refused", Args: []string{"escapes"}, Want: "refused\nrefused\nrefused\nrefused\n",
				Explain: "Checking for \"../\" at the start is not enough: \"a/../../etc\" climbs out from the middle. Refuse a \"..\" when there is no component left to remove."},
			{Name: "absolute names are refused", Args: []string{"absolute"}, Want: "refused\nrefused\n"},
			{Name: "names of no file are refused", Args: []string{"nothing"}, Want: "refused\nrefused\nrefused\n"},
			{Name: "the output size is respected", Args: []string{"small"}, Want: "/srv/files/report.pdf\nrefused\nrefused\n",
				Explain: "Check the length before copying, counting the '\\0', and never write past out[size - 1]: with size 0, write nothing at all."},
			{Name: "a 6000-character name", Args: []string{"long"}, Want: "0 6011\n"},
		},
		Hints: []string{
			"Walk name component by component (strchr for the next '/', or an index), appending each kept component to out after a '/'.",
			"To handle \"..\", remember where each component starts in out (a small array of offsets), and cut out back to the previous start.",
			"Build into out only after checking there is room: len_so_far + 1 + component_length + 1 <= size.",
		},
	},
	{
		ID: "copy_string", Stage: 17, Concept: "Writing Memory-Safe Code",
		Title: "A bounded string copy",
		Task:  "Write size_t copy_string(char *dst, size_t size, const char *src), like OpenBSD's strlcpy: copy at most size - 1 characters of src into dst, always end dst with '\\0' when size is not 0, never write past dst[size - 1], and return strlen(src), so the caller can tell whether it was cut short (the result is >= size). No main.",
		File:  "copy_string.c",
		Starter: `/* Checked exercise: copy_string
 * size_t copy_string(char *dst, size_t size, const char *src);
 * Like strlcpy: copies at most size - 1 bytes, always terminates dst
 * (unless size is 0), never writes past dst[size - 1], and returns
 * strlen(src): if that is >= size, the copy was cut short.
 */
#include <stddef.h>
`,
		Harness: `#include <stdio.h>
#include <string.h>

size_t copy_string(char *dst, size_t size, const char *src);

static void try(size_t size, const char *src) {
    char buf[64];
    memset(buf, '#', sizeof buf);
    char *dst = buf + 8;
    size_t r = copy_string(dst, size, src);
    int clean = 1;
    for (int i = 0; i < 8; i++) clean &= buf[i] == '#';
    for (size_t i = 8 + size; i < sizeof buf; i++) clean &= buf[i] == '#';
    if (!clean) printf("%zu WROTE OUTSIDE dst\n", r);
    else if (size == 0) printf("%zu untouched\n", r);
    else if (memchr(dst, '\0', size) == NULL) printf("%zu NOT TERMINATED\n", r);
    else printf("%zu [%s]\n", r, dst);
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "fits") == 0) { try(16, "hello"); try(6, "hello"); }
    else if (strcmp(argv[1], "cut") == 0) { try(4, "hello"); try(1, "hello"); }
    else if (strcmp(argv[1], "zero") == 0) try(0, "hello");
    else if (strcmp(argv[1], "empty") == 0) try(8, "");
    else if (strcmp(argv[1], "long") == 0) {
        static char big[100001];
        memset(big, 'a', 100000);
        try(8, big);
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "it fits", Args: []string{"fits"}, Want: "5 [hello]\n5 [hello]\n",
				Explain: "Size 6 holds the 5 letters and the '\\0' exactly."},
			{Name: "cut short", Args: []string{"cut"}, Want: "5 [hel]\n5 []\n",
				Explain: "Copy size - 1 characters, then write the '\\0' at dst[size - 1]; still return the full length of src."},
			{Name: "size 0", Args: []string{"zero"}, Want: "5 untouched\n",
				Explain: "With size 0 there is no room even for the '\\0': write nothing (size - 1 on a size_t of 0 wraps to a huge number)."},
			{Name: "an empty source", Args: []string{"empty"}, Want: "0 []\n"},
			{Name: "a 100,000-character source", Args: []string{"long"}, Want: "100000 [aaaaaaa]\n"},
		},
		Hints: []string{
			"Count the length of src first (or as you go), and copy only while i + 1 < size.",
			"This is the function to reach for instead of strncpy, which neither always terminates nor tells you it cut the string short.",
		},
	},
}

// openMaze is a maze with no walls, S and E at the given cells.
func openMaze(rows, cols, sr, sc, er, ec int) string {
	var b strings.Builder
	for r := 0; r < rows; r++ {
		row := []byte(strings.Repeat(".", cols))
		if r == sr {
			row[sc] = 'S'
		}
		if r == er {
			row[ec] = 'E'
		}
		b.Write(row)
		b.WriteByte('\n')
	}
	return b.String()
}

// serpentine is a 21 by 21 maze whose only path winds through every
// row: walls on alternate rows, with the gap at alternating ends.
func serpentine() string {
	const n = 21
	var b strings.Builder
	for r := 0; r < n; r++ {
		row := []byte(strings.Repeat(".", n))
		if r%2 == 1 {
			for c := range row {
				row[c] = '#'
			}
			if r%4 == 1 {
				row[n-1] = '.'
			} else {
				row[0] = '.'
			}
		}
		if r == 0 {
			row[0] = 'S'
		}
		if r == n-1 {
			row[0] = 'E'
		}
		b.Write(row)
		b.WriteByte('\n')
	}
	return b.String()
}

// serpentineSteps is the length of the only path through serpentine():
// found by a breadth-first search here, so the expected answer is
// computed, not typed in.
func serpentineSteps() int {
	lines := strings.Split(strings.TrimSpace(serpentine()), "\n")
	rows, cols := len(lines), len(lines[0])
	dist := make([]int, rows*cols)
	for i := range dist {
		dist[i] = -1
	}
	var queue []int
	for r, l := range lines {
		if c := strings.IndexByte(l, 'S'); c >= 0 {
			dist[r*cols+c] = 0
			queue = append(queue, r*cols+c)
		}
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		r, c := cur/cols, cur%cols
		if lines[r][c] == 'E' {
			return dist[cur]
		}
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nr, nc := r+d[0], c+d[1]
			if nr < 0 || nc < 0 || nr >= rows || nc >= cols || lines[nr][nc] == '#' || dist[nr*cols+nc] >= 0 {
				continue
			}
			dist[nr*cols+nc] = dist[cur] + 1
			queue = append(queue, nr*cols+nc)
		}
	}
	return -1
}

// manyKVCommands sets k0…k499, then gets k250 and a missing key.
func manyKVCommands() string {
	var b strings.Builder
	for i := 0; i < 500; i++ {
		fmt.Fprintf(&b, "set k%d v%d\n", i, i)
	}
	b.WriteString("get k250\nget k500\n")
	return b.String()
}
