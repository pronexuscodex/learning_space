package main

// Checked exercises for the systems stages: bits and parsing (Stage 5),
// processes, system calls and threads (Stage 6), and sockets and HTTP
// (Stage 8). They use only C and POSIX, so they run on Linux, macOS and
// the BSDs alike.

import (
	"strings"
	"time"
)

// posixDefine is the first line of every POSIX check: with -std=c17,
// glibc hides fork, read and sockets unless a program asks for POSIX.
const posixDefine = "#define _POSIX_C_SOURCE 200809L /* ask for POSIX: fork, read, sockets… */\n"

var systemsChecks = []Check{
	{
		ID: "bits", Stage: 5, Concept: "Integers & Floating Point in Bits",
		Title: "Bit tricks and the bits of a float",
		Task:  "Write four functions, and no main: popcount (how many bits are 1), rotate_left (bits that fall off the left come back on the right), is_power_of_two, and float_parts, which splits a float into the raw sign, exponent and mantissa fields of IEEE 754.",
		File:  "bits.c",
		Starter: `/* Checked exercise: bits
 * Write these four functions; the checker has its own main.
 *   int popcount(unsigned x);              how many bits of x are 1
 *   unsigned rotate_left(unsigned x, int n);  n from 0 to 31; bits leaving
 *                                          the left come back on the right
 *   int is_power_of_two(unsigned x);       1 for 1, 2, 4, ...; else 0
 *   void float_parts(float f, int *sign, int *exponent, unsigned *mantissa);
 *       the raw IEEE 754 fields: 1 sign bit, 8 exponent bits (biased by
 *       127), 23 mantissa bits
 */
`,
		Harness: `#include <stdio.h>
#include <string.h>

int popcount(unsigned x);
unsigned rotate_left(unsigned x, int n);
int is_power_of_two(unsigned x);
void float_parts(float f, int *sign, int *exponent, unsigned *mantissa);

static void parts(float f) {
    int s = -1, e = -1;
    unsigned m = 0xFFFFFFFFu;
    float_parts(f, &s, &e, &m);
    printf("%d %d %06x\n", s, e, m);
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "popcount") == 0) {
        printf("%d %d %d %d %d\n", popcount(0), popcount(1), popcount(0xFFFFFFFFu), popcount(0x80000001u), popcount(0xF0F0u));
    } else if (strcmp(argv[1], "rotate") == 0) {
        printf("%08x %08x %08x\n", rotate_left(0x80000001u, 1), rotate_left(0x12345678u, 8), rotate_left(0xDEADBEEFu, 0));
    } else if (strcmp(argv[1], "power") == 0) {
        unsigned xs[] = {0, 1, 2, 3, 64, 96, 0x80000000u};
        for (int i = 0; i < 7; i++) printf("%d", is_power_of_two(xs[i]));
        printf("\n");
    } else if (strcmp(argv[1], "float") == 0) {
        parts(1.0f);
        parts(-2.5f);
        parts(0.1f);
        parts(0.0f);
        parts(-0.0f);
    } else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "popcount", Args: []string{"popcount"}, Want: "0 1 32 2 8\n",
				Explain: "Test the lowest bit with x & 1, then shift right; or clear the lowest set bit with x &= x - 1 until x is 0."},
			{Name: "rotate_left, including by 0", Args: []string{"rotate"}, Want: "00000003 34567812 deadbeef\n",
				Explain: "x >> (32 - n) with n = 0 shifts by 32, which is undefined behaviour for a 32-bit unsigned: handle n == 0 on its own."},
			{Name: "is_power_of_two", Args: []string{"power"}, Want: "0110101\n",
				Explain: "A power of two has exactly one bit set, so x & (x - 1) is 0, but so is 0 itself: rule it out first."},
			{Name: "the fields of five floats", Args: []string{"float"}, Want: "0 127 000000\n1 128 200000\n0 123 4ccccd\n0 0 000000\n1 0 000000\n",
				Explain: "Copy the float's bits into a uint32_t with memcpy (a pointer cast breaks C's aliasing rules), then sign = bits >> 31, exponent = (bits >> 23) & 0xFF, mantissa = bits & 0x7FFFFF. -0.0 has its sign bit set."},
		},
		Hints: []string{
			"Use unsigned for bit work: shifting a signed int's sign bit is not portable.",
			"memcpy(&bits, &f, sizeof bits) with uint32_t bits (stdint.h) gives you the float's raw bits.",
			"0.1 has no exact binary form, which is why its mantissa is the repeating pattern 4ccccd.",
		},
	},
	{
		ID: "calc", Stage: 5, Concept: "Parsing: Recursive Descent & Pratt",
		Title: "A calculator: recursive descent",
		Task:  "Read expressions from standard input, one per line, and print each one's value, or error. Expressions hold whole numbers, + - * /, parentheses and unary minus, with the usual precedence and left-to-right order (10 - 4 - 3 is 3). Division truncates towards zero, as in C. A malformed line, an empty line or a division by zero prints error. Use long long for the values.",
		File:  "calc.c",
		Starter: `/* Checked exercise: calc
 * For each line of stdin, print the value of the expression, or "error".
 *   numbers, + - * /, ( ), unary minus, spaces anywhere
 *   * and / bind tighter than + and -; same level: left to right
 *   division truncates towards zero; division by zero is an error
 * A recursive-descent parser has one function per grammar rule:
 *   expr   = term   { ("+" | "-") term }
 *   term   = factor { ("*" | "/") factor }
 *   factor = number | "(" expr ")" | "-" factor
 */
`,
		Cases: []CheckCase{
			{Name: "precedence", Stdin: "1 + 2 * 3\n(1 + 2) * 3\n2 * (3 + 4) * 5\n", Want: "7\n9\n70\n",
				Explain: "* and / belong one level down the grammar from + and -: expr calls term, term calls factor."},
			{Name: "left to right", Stdin: "10 - 4 - 3\n100 / 10 / 5\n", Want: "3\n2\n",
				Explain: "10 - (4 - 3) = 9 comes from recursing on the right; loop instead: value = term; while the next token is + or -, combine it with the next term."},
			{Name: "unary minus", Stdin: "-3 * -(2 + 1)\n--4\n", Want: "9\n4\n"},
			{Name: "division truncates towards zero", Stdin: "-7 / 2\n7 / -2\n", Want: "-3\n-3\n"},
			{Name: "errors", Stdin: "7 / 0\n1 +\n(1 + 2\n1 2\n\n2 * )\n", Want: "error\nerror\nerror\nerror\nerror\nerror\n",
				Explain: "After a complete expression, only the end of the line may follow; any other leftover input (\"1 2\") is an error."},
			{Name: "big numbers", Stdin: "123456789 * 1000\n", Want: "123456789000\n"},
			{Name: "deep nesting", Stdin: strings.Repeat("(", 500) + "42" + strings.Repeat(")", 500) + "\n", Want: "42\n"},
		},
		Hints: []string{
			"Keep a position in the current line, and a function that skips spaces and looks at the next character.",
			"Report an error through a flag (or a return code) that every level checks, rather than printing from deep inside.",
			"Read each line with fgets into a buffer big enough for the longest line you accept, and treat a line that did not fit as an error.",
		},
	},
	{
		ID: "run", Stage: 6, Concept: "Processes & the Address Space",
		Title: "Run a program: fork, exec, wait",
		Task:  "./run PROGRAM ARGS... starts the program as a child process with fork and execvp, waits for it, and then prints exit N (it exited with status N) or signal N (a signal killed it). If exec fails, the child reports why on standard error and exits with status 127. With no arguments, print a usage message to standard error and return 2.",
		File:  "run.c",
		Starter: posixDefine + `/* Checked exercise: run
 * Usage: ./run PROGRAM ARGS...
 * fork; in the child, execvp(argv[1], &argv[1]); if that returns,
 * perror and _exit(127). In the parent, waitpid, then print
 *   exit N     (WIFEXITED, WEXITSTATUS)
 *   signal N   (WIFSIGNALED, WTERMSIG)
 * No arguments: usage on stderr, return 2.
 */
`,
		Cases: []CheckCase{
			{Name: "exit 0", Args: []string{"/bin/sh", "-c", "exit 0"}, Want: "exit 0\n"},
			{Name: "exit 3", Args: []string{"/bin/sh", "-c", "exit 3"}, Want: "exit 3\n",
				Explain: "waitpid's status is not the exit code itself: use WIFEXITED(status) and WEXITSTATUS(status)."},
			{Name: "killed by a signal", Args: []string{"/bin/sh", "-c", "kill -TERM $$"}, Want: "signal 15\n",
				Explain: "When WIFSIGNALED(status) is true, WTERMSIG(status) is the signal: SIGTERM is 15."},
			{Name: "arguments are passed on", Args: []string{"/bin/echo", "one", "two"}, Want: "one two\nexit 0\n",
				Explain: "execvp takes the program and the whole argument vector, &argv[1], which ends with NULL already."},
			{Name: "a program that does not exist", Args: []string{"/no/such/program"}, Want: "exit 127\n", Stderr: true,
				Explain: "execvp only returns when it fails: then perror and _exit(127) in the child. Using exit instead of _exit would flush the parent's buffered output a second time."},
			{Name: "no arguments", WantExit: 2, Stderr: true},
		},
		Hints: []string{
			"fork returns 0 in the child, the child's process ID in the parent, and -1 on failure.",
			"waitpid(pid, &status, 0) waits for that one child; the macros in sys/wait.h decode status.",
			"Call fflush(stdout) before fork if you have printed anything, or the child inherits the unflushed buffer and prints it again.",
		},
	},
	{
		ID: "rawcopy", Stage: 6, Concept: "System Calls & Privilege",
		Title: "Copy with system calls only",
		Task:  "Copy all of standard input to standard output using only the read and write system calls (unistd.h), not stdio, through a buffer of 4096 bytes or more. Handle short writes (write may write fewer bytes than asked), and report errors with perror and status 1. The input may hold any bytes.",
		File:  "rawcopy.c",
		Starter: posixDefine + `/* Checked exercise: rawcopy
 * Copy stdin (file descriptor 0) to stdout (1) with read and write only.
 * read returns how many bytes it read, 0 at the end, -1 on error.
 * write may write fewer bytes than you asked: loop until all are written.
 * See your system calls: strace -e trace=read,write ./rawcopy < file
 */
`,
		Cases: []CheckCase{
			{Name: "a line of text", Stdin: "hello, system calls\n", Want: "hello, system calls\n"},
			{Name: "no input", Stdin: "", Want: ""},
			{Name: "half a megabyte of every byte value", Stdin: allBytes(512 << 10), Want: allBytes(512 << 10),
				Explain: "Write exactly as many bytes as read returned, not the buffer's size, and nothing stops at a zero byte."},
		},
		Hints: []string{
			"ssize_t n = read(0, buf, sizeof buf); n bytes are now in buf.",
			"Keep an offset into the buffer and call write again for what is left until it is all written.",
			"read can fail with errno == EINTR when a signal arrives: try again in that case.",
		},
	},
	{
		ID: "counter", Stage: 6, Concept: "Synchronisation & Memory Ordering",
		Title: "Threads and a mutex",
		Task:  "Write long count_parallel(int threads, long per_thread): start that many POSIX threads, have each add 1 to one shared counter per_thread times, wait for them all, and return the total. Without a lock, increments are lost; protect the counter with a pthread mutex. No main: the checker calls it.",
		File:  "counter.c",
		Flags: []string{"-pthread"},
		Starter: posixDefine + `/* Checked exercise: counter
 * long count_parallel(int threads, long per_thread);
 * Start the threads (pthread_create), each adding 1 to a shared counter
 * per_thread times; join them all (pthread_join); return the counter.
 * count++ is a load, an add and a store: two threads can lose updates.
 * Guard it with a pthread_mutex_t. The checker supplies main.
 */
#include <pthread.h>
`,
		Harness: posixDefine + `#include <stdio.h>
#include <stdlib.h>

long count_parallel(int threads, long per_thread);

int main(int argc, char **argv) {
    if (argc < 3) return 2;
    printf("%ld\n", count_parallel(atoi(argv[1]), atol(argv[2])));
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "one thread", Args: []string{"1", "1000"}, Want: "1000\n"},
			{Name: "four threads", Args: []string{"4", "100000"}, Want: "400000\n",
				Explain: "A smaller total means lost updates: two threads read the same value and both stored value + 1. Lock the mutex around the increment."},
			{Name: "eight threads", Args: []string{"8", "50000"}, Want: "400000\n"},
			{Name: "no threads", Args: []string{"0", "1000"}, Want: "0\n"},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"pthread_create(&tid[i], NULL, worker, arg) starts worker(arg) in a new thread; keep the thread IDs to join them.",
			"Each thread needs per_thread and a pointer to the shared state: pass a struct through the void * argument.",
			"Lock and unlock around count++ only; doing the whole loop under the lock is correct too, but runs the threads one at a time.",
		},
	},
	{
		ID: "upper_server", Stage: 8, Concept: "Sockets & I/O Multiplexing",
		Title: "A TCP server: shout it back",
		Task:  "Write void serve(int fd): fd is a connected TCP socket. Read everything the client sends until it closes its side, and send each byte back with a–z turned into A–Z, until the client is done. The checker opens a real socket on 127.0.0.1, accepts the connection, calls serve, and sends text, nothing, and 300 KB through it.",
		File:  "upper_server.c",
		Starter: posixDefine + `/* Checked exercise: upper_server
 * void serve(int fd);  fd is a connected TCP socket.
 * Loop: read a chunk; uppercase a-z; write ALL of it back (write may
 * send less than you asked); stop when read returns 0 (the client closed
 * its side). The checker supplies main, the listening socket and the
 * client, on 127.0.0.1.
 */
#include <unistd.h>
`,
		Harness: posixDefine + `#include <arpa/inet.h>
#include <netinet/in.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/wait.h>
#include <unistd.h>

void serve(int fd);

static size_t make_input(const char *kind, char *buf, size_t cap) {
    if (strcmp(kind, "hello") == 0) {
        const char *s = "hello, world\nsecond line: 42!\n";
        memcpy(buf, s, strlen(s));
        return strlen(s);
    }
    if (strcmp(kind, "empty") == 0) return 0;
    size_t n = strcmp(kind, "big") == 0 ? 300000 : 256;
    for (size_t i = 0; i < n && i < cap; i++)
        buf[i] = strcmp(kind, "big") == 0 ? (char)(i % 61 == 60 ? '\n' : 'a' + i % 26) : (char)i;
    return n;
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    signal(SIGPIPE, SIG_IGN);
    static char in[300000], out[400000];
    size_t n = make_input(argv[1], in, sizeof in);

    int ls = socket(AF_INET, SOCK_STREAM, 0);
    struct sockaddr_in a;
    memset(&a, 0, sizeof a);
    a.sin_family = AF_INET;
    a.sin_addr.s_addr = htonl(0x7f000001); /* 127.0.0.1 */
    a.sin_port = 0;                         /* any free port */
    socklen_t al = sizeof a;
    if (ls < 0 || bind(ls, (struct sockaddr *)&a, sizeof a) < 0 || listen(ls, 1) < 0 ||
        getsockname(ls, (struct sockaddr *)&a, &al) < 0) { perror("setting up the socket"); return 3; }

    pid_t server = fork();
    if (server == 0) {
        int fd = accept(ls, NULL, NULL);
        if (fd < 0) _exit(4);
        serve(fd);
        close(fd);
        _exit(0);
    }
    close(ls);
    int c = socket(AF_INET, SOCK_STREAM, 0);
    if (c < 0 || connect(c, (struct sockaddr *)&a, sizeof a) < 0) { perror("connect"); return 3; }
    pid_t writer = fork();
    if (writer == 0) {
        size_t off = 0;
        while (off < n) {
            ssize_t w = write(c, in + off, n - off);
            if (w <= 0) _exit(5);
            off += (size_t)w;
        }
        shutdown(c, SHUT_WR);
        _exit(0);
    }
    /* No shutdown here: it acts on the socket itself, so it would also end
     * the writer child's side before it had written everything. */
    size_t got = 0;
    ssize_t r;
    while ((r = read(c, out + got, sizeof out - got)) > 0) got += (size_t)r;
    close(c);
    waitpid(writer, NULL, 0);
    int st = 0;
    waitpid(server, &st, 0);

    if (strcmp(argv[1], "hello") == 0 || strcmp(argv[1], "empty") == 0) {
        printf("%.*s", (int)got, out);
        printf("[%zu bytes]\n", got);
        return 0;
    }
    int ok = got == n;
    for (size_t i = 0; ok && i < n; i++) {
        unsigned char ch = (unsigned char)in[i];
        unsigned char want = ch >= 'a' && ch <= 'z' ? (unsigned char)(ch - 'a' + 'A') : ch;
        ok = (unsigned char)out[i] == want;
    }
    printf("%zu %s\n", got, ok ? "ok" : "wrong");
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "two lines", Args: []string{"hello"}, Want: "HELLO, WORLD\nSECOND LINE: 42!\n[30 bytes]\n",
				Explain: "Only a-z change; digits, punctuation and newlines are sent back as they are."},
			{Name: "the client sends nothing", Args: []string{"empty"}, Want: "[0 bytes]\n",
				Explain: "read returns 0 at once when the client closes without sending: return from serve."},
			{Name: "300 KB in pieces", Args: []string{"big"}, Want: "300000 ok\n",
				Explain: "A large transfer arrives in many pieces: keep reading until read returns 0, and write back exactly the bytes each read returned."},
			{Name: "every byte value", Args: []string{"bytes"}, Want: "256 ok\n",
				Explain: "Bytes above 127 must come back unchanged: test the range 'a' to 'z' yourself, or pass toupper an unsigned char."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"ssize_t n = read(fd, buf, sizeof buf): n can be anything from 1 to sizeof buf, whatever has arrived.",
			"Write in a loop, like rawcopy: write may send only part of the buffer.",
			"The checker writes and reads at the same time, so you can write each chunk back as soon as you have read it.",
		},
	},
	{
		ID: "http_request", Stage: 8, Concept: "HTTP, DNS & How the Web Works",
		Title: "Read an HTTP request",
		Task:  "Read the head of one HTTP/1.1 request from standard input: the request line, then header lines, up to the first empty line (lines end in \\r\\n; accept a bare \\n too). If it is valid, print METHOD PATH HOST and return 0. Valid means: the request line is METHOD SP PATH SP HTTP/1.1, with an upper-case method and a path starting with /; every header line holds a colon; and there is a Host header, whose name matches in any case and whose value has the spaces around it trimmed. Anything else: print nothing to standard output, a message to standard error, and return 1.",
		File:  "http_request.c",
		Starter: `/* Checked exercise: http_request
 * Read an HTTP/1.1 request head from stdin, up to the first empty line:
 *   GET /index.html HTTP/1.1\r\n
 *   Host: example.com\r\n
 *   \r\n
 * Valid: print "METHOD PATH HOST", return 0. Otherwise: nothing on stdout,
 * a message on stderr, return 1. Header names ignore case; trim the
 * spaces around the Host value; ignore anything after the empty line.
 * Accept request lines and header lines up to 8000 bytes.
 */
`,
		Cases: []CheckCase{
			{Name: "a simple GET", Stdin: "GET /index.html HTTP/1.1\r\nHost: example.com\r\n\r\n", Want: "GET /index.html example.com\n"},
			{Name: "Host in any case, with spaces", Stdin: "GET / HTTP/1.1\r\nUser-Agent: academy\r\nhost:    example.org   \r\nAccept: */*\r\n\r\n", Want: "GET / example.org\n",
				Explain: "Compare header names without regard to case (strncasecmp, or tolower each byte), and trim spaces and tabs around the value."},
			{Name: "a POST with a body", Stdin: "POST /form HTTP/1.1\r\nHost: h.test\r\nContent-Length: 3\r\n\r\nabc", Want: "POST /form h.test\n",
				Explain: "The head ends at the first empty line; whatever follows is the body, which this exercise ignores."},
			{Name: "bare newlines", Stdin: "GET /a HTTP/1.1\nHost: lf.test\n\n", Want: "GET /a lf.test\n"},
			{Name: "a long path", Stdin: "GET /" + strings.Repeat("p", 7000) + " HTTP/1.1\r\nHost: long.test\r\n\r\n", Want: "GET /" + strings.Repeat("p", 7000) + " long.test\n",
				Explain: "The request line can be 8000 bytes: size your line buffer for it."},
			{Name: "no Host header", Stdin: "GET / HTTP/1.1\r\nAccept: */*\r\n\r\n", WantExit: 1, Stderr: true,
				Explain: "HTTP/1.1 requires a Host header: without one the request is invalid."},
			{Name: "a broken request line", Stdin: "GET\r\nHost: x\r\n\r\n", WantExit: 1, Stderr: true},
			{Name: "a path without /", Stdin: "GET index.html HTTP/1.1\r\nHost: x\r\n\r\n", WantExit: 1, Stderr: true},
			{Name: "another HTTP version", Stdin: "GET / HTTP/1.0\r\nHost: x\r\n\r\n", WantExit: 1, Stderr: true},
			{Name: "a header without a colon", Stdin: "GET / HTTP/1.1\r\nHost example.com\r\n\r\n", WantExit: 1, Stderr: true},
			{Name: "the head never ends", Stdin: "GET / HTTP/1.1\r\nHost: x\r\n", WantExit: 1, Stderr: true,
				Explain: "Without the empty line the head is incomplete: end of input before it is an error."},
		},
		Hints: []string{
			"Read line by line with fgets into a buffer of 8192 bytes, and strip the trailing \\n and an \\r before it.",
			"sscanf(line, \"%15s %7999s %15s\", method, path, version) == 3 splits the request line; then check each part.",
			"Remember whether you saw Host, and only print once you reach the empty line.",
		},
	},
}

// allBytes is n bytes cycling through every value from 0 to 255.
func allBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i * 7 % 256)
	}
	return string(b)
}
