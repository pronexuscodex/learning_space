package main

// Code generated from the tested type-in programs; the listings and their
// expected output were produced by actually running each program. Do not
// edit by hand: change the program, re-run it, and regenerate.

// typeIns holds one magazine-style listing per stage.
var typeIns = map[int]TypeIn{
	0: {
		File:    "cpu.c",
		Lang:    "C",
		Run:     "gcc -Wall -Wextra -std=c17 cpu.c -o cpu && ./cpu",
		Predict: "Trace the program on paper before running it: what number will it print, and how many instructions will the tiny CPU execute before it halts? (Hint: the loop body is 7 instructions.)",
		Lesson:  "The loop runs 5 times × 7 instructions = 35, plus 3 to print and halt: 38 steps, printing 15. Notice that the program and its data share the same memory (the von Neumann design), and that a loop is nothing but a conditional jump back to an earlier address. A real CPU does exactly this, billions of times per second.",
		Code: `/* TINY CPU -- a type-in in the spirit of 1980s magazine listings.
 * A pretend computer with 32 memory cells and one register (acc).
 * Instructions are numbers: opcode * 100 + address.
 *   1xx LOAD  acc = mem[xx]     4xx STORE mem[xx] = acc
 *   2xx ADD   acc += mem[xx]    5xx JNZ   jump to xx if acc != 0
 *   3xx SUB   acc -= mem[xx]    600 PRINT acc      000 HALT
 * The program below adds 5 + 4 + 3 + 2 + 1. */
#include <stdio.h>

int main(void) {
    int mem[32] = {
        120, 221, 420,      /* 0: sum = sum + n            */
        121, 322, 421,      /* 3: n = n - 1                */
        500,                /* 6: if n != 0 go back to 0   */
        120, 600, 0,        /* 7: print sum, then halt     */
    };
    mem[20] = 0;            /* sum */
    mem[21] = 5;            /* n   */
    mem[22] = 1;            /* the constant 1 */

    int pc = 0, acc = 0, steps = 0;
    for (;;) {
        int instr = mem[pc];            /* FETCH  */
        int op = instr / 100;           /* DECODE */
        int addr = instr % 100;
        pc++;
        steps++;
        if (steps <= 7)
            printf("step %d: pc=%d instr=%03d acc=%d\n", steps, pc - 1, instr, acc);
        switch (op) {                   /* EXECUTE */
        case 1: acc = mem[addr]; break;
        case 2: acc += mem[addr]; break;
        case 3: acc -= mem[addr]; break;
        case 4: mem[addr] = acc; break;
        case 5: if (acc != 0) pc = addr; break;
        case 6: printf("OUTPUT: %d\n", acc); break;
        case 0: printf("halted after %d steps\n", steps); return 0;
        }
    }
}`,
		Expected: `step 1: pc=0 instr=120 acc=0
step 2: pc=1 instr=221 acc=0
step 3: pc=2 instr=420 acc=5
step 4: pc=3 instr=121 acc=5
step 5: pc=4 instr=322 acc=5
step 6: pc=5 instr=421 acc=4
step 7: pc=6 instr=500 acc=4
OUTPUT: 15
halted after 38 steps`,
	},
	1: {
		File:    "collatz.c",
		Lang:    "C",
		Run:     "gcc -Wall -Wextra -std=c17 collatz.c -o collatz && ./collatz",
		Predict: "How many steps does 27 take to reach 1, and what is the highest value it reaches on the way? Write your guesses down first. Why is n a long rather than an int?",
		Lesson:  "Tiny rules can produce surprisingly long journeys: 27 climbs past 9,000 before falling to 1. Nobody has proved that every starting number reaches 1 (the Collatz conjecture). n is a long because some starting numbers climb far beyond what a 32-bit int can hold, and in C a signed overflow is undefined behaviour. The size line depends on your machine: 8 bytes on 64-bit Linux and macOS, 4 on Windows.",
		Code: `/* COLLATZ -- a type-in in the spirit of 1980s magazine listings.
 * Start with n. If n is even, halve it; if odd, make it 3n + 1.
 * Count the steps until n reaches 1. */
#include <stdio.h>

int main(void) {
    long n = 27;
    int steps = 0;
    long peak = n;

    while (n != 1) {
        if (n % 2 == 0) {
            n = n / 2;
        } else {
            n = 3 * n + 1;
        }
        steps++;
        if (n > peak) {
            peak = n;
        }
    }
    printf("steps: %d\n", steps);
    printf("highest value: %ld\n", peak);
    printf("sizeof(long) on this machine: %zu bytes\n", sizeof n);
    return 0;
}`,
		Expected: `steps: 111
highest value: 9232
sizeof(long) on this machine: 8 bytes`,
	},
	2: {
		File:    "bsearch.py",
		Lang:    "Python",
		Run:     "python3 bsearch.py",
		Predict: "In a sorted list of one million numbers, how many guesses does binary search need to find 1? 500,000? A number that is not there?",
		Lesson:  "About log₂(1,000,000) ≈ 20 guesses at most, and the exact middle is found on the very first guess.",
		Code: `# BINARY SEARCH -- count how many guesses it takes.

def search(items, target):
    low, high = 0, len(items) - 1
    guesses = 0
    while low <= high:
        mid = (low + high) // 2
        guesses += 1
        if items[mid] == target:
            return mid, guesses
        if items[mid] < target:
            low = mid + 1
        else:
            high = mid - 1
    return -1, guesses

numbers = list(range(1, 1_000_001))      # 1 .. 1,000,000
for target in (1, 500_000, 999_999, 1_000_001):
    index, guesses = search(numbers, target)
    print(target, "->", "found" if index >= 0 else "missing", "in", guesses, "guesses")`,
		Expected: `1 -> found in 19 guesses
500000 -> found in 1 guesses
999999 -> found in 19 guesses
1000001 -> missing in 20 guesses`,
	},
	3: {
		File:    "sieve.py",
		Lang:    "Python",
		Run:     "python3 sieve.py",
		Predict: "Roughly how many primes are there below 10,000? (Hint: the prime number theorem says about n / ln n.)",
		Lesson:  "10,000 / ln 10,000 ≈ 1,086, a decent estimate of the true 1,229. The sieve was a favourite benchmark in early-1980s computer magazines.",
		Code: `# SIEVE OF ERATOSTHENES -- the classic benchmark of home-computer
# magazines. Cross out multiples; whatever survives is prime.

LIMIT = 10_000
is_prime = [True] * LIMIT
is_prime[0] = is_prime[1] = False
for i in range(2, int(LIMIT ** 0.5) + 1):
    if is_prime[i]:
        for multiple in range(i * i, LIMIT, i):
            is_prime[multiple] = False

primes = [i for i, p in enumerate(is_prime) if p]
print("primes below", LIMIT, ":", len(primes))
print("last five:", primes[-5:])`,
		Expected: `primes below 10000 : 1229
last five: [9931, 9941, 9949, 9967, 9973]`,
	},
	4: {
		File:    "adder.py",
		Lang:    "Python",
		Run:     "python3 adder.py",
		Predict: "What does an 8-bit adder give for 200 + 100, and for 255 + 1? What happens to the carry?",
		Lesson:  "300 doesn't fit in 8 bits: the adder keeps 300 − 256 = 44 and raises the carry flag. This is overflow, built from nothing but AND, OR and XOR.",
		Code: `# RIPPLE-CARRY ADDER -- addition built only from logic gates.

def AND(a, b): return a & b
def XOR(a, b): return a ^ b
def OR(a, b):  return a | b

def full_adder(a, b, carry_in):
    s = XOR(XOR(a, b), carry_in)
    carry_out = OR(AND(a, b), AND(carry_in, XOR(a, b)))
    return s, carry_out

def add8(x, y):
    carry, result = 0, 0
    for bit in range(8):                     # least significant bit first
        a = (x >> bit) & 1
        b = (y >> bit) & 1
        s, carry = full_adder(a, b, carry)
        result |= s << bit
    return result, carry

for x, y in ((12, 30), (200, 100), (255, 1)):
    total, carry = add8(x, y)
    print(f"{x} + {y} = {total} (carry out {carry})  bits {total:08b}")`,
		Expected: `12 + 30 = 42 (carry out 0)  bits 00101010
200 + 100 = 44 (carry out 1)  bits 00101100
255 + 1 = 0 (carry out 1)  bits 00000000`,
	},
	5: {
		File:    "bits.c",
		Lang:    "C",
		Run:     "gcc bits.c -o bits && ./bits",
		Predict: "Predict every line before running: 255 + 1 in 8 bits, 127 + 1 in signed 8 bits, whether 0.1 + 0.2 equals 0.3, and the two sizes.",
		Lesson:  "Integers wrap around and decimals are approximated in binary. The sizes shown are typical of modern 64-bit machines, and C allows them to differ elsewhere.",
		Code: `/* BITS -- what numbers really are inside the machine. */
#include <stdio.h>

int main(void) {
    unsigned char small = 255;
    small = small + 1;                       /* 8 bits: wraps around */
    printf("255 + 1 in 8 bits = %d\n", small);

    signed char s = 127;
    s = (signed char)(s + 1);
    printf("127 + 1 in signed 8 bits = %d\n", s);

    double a = 0.1 + 0.2;
    printf("0.1 + 0.2 == 0.3 ? %s\n", a == 0.3 ? "yes" : "no");
    printf("0.1 + 0.2 = %.17f\n", a);

    printf("sizeof(int) = %zu bytes, sizeof(double) = %zu bytes\n",
           sizeof(int), sizeof(double));
    return 0;
}`,
		Expected: `255 + 1 in 8 bits = 0
127 + 1 in signed 8 bits = -128
0.1 + 0.2 == 0.3 ? no
0.1 + 0.2 = 0.30000000000000004
sizeof(int) = 4 bytes, sizeof(double) = 8 bytes`,
	},
	6: {
		File:    "fork.c",
		Lang:    "C",
		Run:     "gcc fork.c -o fork && ./fork",
		Predict: "How many lines will this print? (Each fork() clones every process that runs it.)",
		Lesson:  "Two forks give 2 × 2 = 4 processes, so 4 lines, each with a different process ID. The order can change from run to run, because the scheduler decides who goes first.",
		Code: `/* FORK -- one program becomes many processes. */
#include <stdio.h>
#include <sys/wait.h>
#include <unistd.h>

int main(void) {
    fork();                  /* 1 process becomes 2 */
    fork();                  /* each of those becomes 2 */
    printf("hello from process %d\n", getpid());
    fflush(stdout);
    while (wait(NULL) > 0)   /* parents wait for their children */
        ;
    return 0;
}`,
		Expected: `hello from process 1234
hello from process 1235
hello from process 1236
hello from process 1237`,
	},
	7: {
		File:    "index.py",
		Lang:    "Python",
		Run:     "python3 index.py",
		Predict: "What will the query plan say before and after the index is created?",
		Lesson:  "Without an index the database must SCAN every row; with one it SEARCHes the B-tree directly. The exact wording varies between SQLite versions (older ones say SCAN TABLE).",
		Code: `# INDEX -- watch the database change its plan when an index appears.
import sqlite3

db = sqlite3.connect(":memory:")
db.execute("CREATE TABLE people (id INTEGER PRIMARY KEY, email TEXT, city TEXT)")
db.executemany("INSERT INTO people (email, city) VALUES (?, ?)",
               [(f"user{i}@example.com", f"city{i % 50}") for i in range(10_000)])

query = "SELECT id FROM people WHERE email = 'user9999@example.com'"

def plan():
    return [row[-1] for row in db.execute("EXPLAIN QUERY PLAN " + query)]

print("before index:", plan())
db.execute("CREATE INDEX idx_email ON people (email)")
print("after index: ", plan())
print("result:", db.execute(query).fetchall())`,
		Expected: `before index: ['SCAN people']
after index:  ['SEARCH people USING COVERING INDEX idx_email (email=?)']
result: [(10000,)]`,
	},
	8: {
		File:    "echo.py",
		Lang:    "Python",
		Run:     "python3 echo.py",
		Predict: "What exactly will the client print?",
		Lesson:  "A real TCP connection on your own machine: bind, listen, accept on one side; connect, send, receive on the other.",
		Code: `# ECHO -- a server and a client talking over a real TCP socket.
import socket
import threading

server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.bind(("127.0.0.1", 0))            # port 0: let the OS choose a free port
server.listen(1)
port = server.getsockname()[1]

def serve():
    conn, _ = server.accept()
    data = conn.recv(1024)
    conn.sendall(b"echo: " + data.upper())
    conn.close()

threading.Thread(target=serve).start()

client = socket.create_connection(("127.0.0.1", port))
client.sendall(b"hello, 1985")
print(client.recv(1024).decode())
client.close()
server.close()`,
		Expected: `echo: HELLO, 1985`,
	},
	9: {
		File:    "git_history.sh",
		Lang:    "Shell",
		Run:     "mkdir scratch && cd scratch && sh ../git_history.sh",
		Predict: "How many commits will the main line have, how many will the idea branch have, and how many lines will notes.txt have at the end?",
		Lesson:  "A branch is just a label on a commit. Switching back to the main line hides the experiment's change without losing it. (Needs Git 2.23+ for git switch.)",
		Code: `#!/bin/sh
# GIT -- a whole history in a scratch folder. Run it in an empty directory.
set -e
git init -q history && cd history
git config user.name "Type In" && git config user.email "typein@example.com"
echo "first line" > notes.txt
git add notes.txt && git commit -q -m "Add notes"
echo "second line" >> notes.txt
git commit -q -am "Extend notes"
git switch -q -c idea
echo "an experiment" >> notes.txt
git commit -q -am "Try an idea"
git switch -q -
echo "commits on the main line:  $(git rev-list --count HEAD)"
echo "commits on the idea branch: $(git rev-list --count idea)"
echo "lines in notes.txt here:    $(wc -l < notes.txt | tr -d ' ')"`,
		Expected: `commits on the main line:  2
commits on the idea branch: 3
lines in notes.txt here:    2`,
	},
	10: {
		File:    "inject.py",
		Lang:    "Python",
		Run:     "python3 inject.py",
		Predict: "Which users will each query return when the attacker types ' OR '1'='1 into both fields?",
		Lesson:  "Glued-together SQL lets input become code, so every user leaks. With parameters, the input is only ever data, and nobody is literally called ' OR '1'='1.",
		Code: `# INJECTION -- the same login, written badly and written well.
import sqlite3

db = sqlite3.connect(":memory:")
db.execute("CREATE TABLE users (name TEXT, password TEXT)")
db.executemany("INSERT INTO users VALUES (?, ?)",
               [("alice", "s3cret"), ("bob", "hunter2"), ("carol", "letmein")])

typed_name = "' OR '1'='1"              # what an attacker types
typed_password = "' OR '1'='1"

unsafe = ("SELECT name FROM users WHERE name = '" + typed_name +
          "' AND password = '" + typed_password + "'")
print("unsafe query returns:", db.execute(unsafe).fetchall())

safe = "SELECT name FROM users WHERE name = ? AND password = ?"
print("safe query returns:  ", db.execute(safe, (typed_name, typed_password)).fetchall())`,
		Expected: `unsafe query returns: [('alice',), ('bob',), ('carol',)]
safe query returns:   []`,
	},
	11: {
		File:    "match.py",
		Lang:    "Python",
		Run:     "python3 match.py",
		Predict: "For each of the six lines, predict True or False before running.",
		Lesson:  "o.d doesn't match \"hello world\": no o is followed by one character and then a d. Tracing the recursion by hand is the best way to see why.",
		Code: `# MATCH -- a tiny regular-expression matcher, in the spirit of the
# famous 30-line matchers of the Unix tradition. Supports:
#   c  a literal character      .  any character
#   ^  start of text            $  end of text
#   *  zero or more of the previous character

def match(pattern, text):
    if pattern.startswith("^"):
        return match_here(pattern[1:], text)
    for start in range(len(text) + 1):   # try every starting position
        if match_here(pattern, text[start:]):
            return True
    return False

def match_here(pattern, text):
    if pattern == "":
        return True
    if len(pattern) >= 2 and pattern[1] == "*":
        return match_star(pattern[0], pattern[2:], text)
    if pattern == "$":
        return text == ""
    if text and (pattern[0] == "." or pattern[0] == text[0]):
        return match_here(pattern[1:], text[1:])
    return False

def match_star(c, rest, text):
    while True:                          # try zero, one, two... copies of c
        if match_here(rest, text):
            return True
        if not text or (c != "." and text[0] != c):
            return False
        text = text[1:]

tests = [("^ab*c$", "ac"), ("^ab*c$", "abbbc"), ("^ab*c$", "abd"),
         ("o.d", "hello world"), ("^h.*d$", "hello world"), ("x", "hello")]
for pattern, text in tests:
    print(f"{pattern:8} {text!r:15} {match(pattern, text)}")`,
		Expected: `^ab*c$   'ac'            True
^ab*c$   'abbbc'         True
^ab*c$   'abd'           False
o.d      'hello world'   False
^h.*d$   'hello world'   True
x        'hello'         False`,
	},
	12: {
		File:    "lisp.py",
		Lang:    "Python",
		Run:     "python3 lisp.py",
		Predict: "What will each of the three programs evaluate to?",
		Lesson:  "Tokenise, parse, evaluate: the whole pipeline of a language in a few dozen lines. Adding define and lambda turns it into a real Lisp.",
		Code: `# LISP -- a calculator language in a few dozen lines:
# read the text, parse it into nested lists, evaluate them.
import operator

def tokenize(src):
    return src.replace("(", " ( ").replace(")", " ) ").split()

def parse(tokens):
    token = tokens.pop(0)
    if token == "(":
        expr = []
        while tokens[0] != ")":
            expr.append(parse(tokens))
        tokens.pop(0)                    # drop ")"
        return expr
    try:
        return int(token)
    except ValueError:
        return token                     # a symbol such as + or max

ENV = {"+": operator.add, "-": operator.sub, "*": operator.mul,
       "max": max, "min": min}

def evaluate(expr):
    if isinstance(expr, int):
        return expr
    if isinstance(expr, str):
        return ENV[expr]
    op, *args = [evaluate(e) for e in expr]
    return op(*args)

for program in ["(+ 1 2)", "(* 2 (+ 3 4))", "(max 3 (- 10 2) (* 2 2))"]:
    print(program, "=>", evaluate(parse(tokenize(program))))`,
		Expected: `(+ 1 2) => 3
(* 2 (+ 3 4)) => 14
(max 3 (- 10 2) (* 2 2)) => 8`,
	},
	13: {
		File:    "power.py",
		Lang:    "Python",
		Run:     "python3 power.py",
		Predict: "Which direction will the vector settle on, and what eigenvalue will the estimate approach?",
		Lesson:  "Repeated multiplication amplifies the strongest direction, (1, 1)/√2 with eigenvalue 3. The same idea underlies Google's original PageRank.",
		Code: `# POWER ITERATION -- find a matrix's strongest direction by
# multiplying by it over and over.

A = [[2.0, 1.0],
     [1.0, 2.0]]

def multiply(m, v):
    return [m[0][0] * v[0] + m[0][1] * v[1],
            m[1][0] * v[0] + m[1][1] * v[1]]

v = [1.0, 0.0]                           # any starting vector
for step in range(1, 21):
    w = multiply(A, v)
    length = (w[0] ** 2 + w[1] ** 2) ** 0.5
    v = [w[0] / length, w[1] / length]
    if step in (1, 2, 5, 20):
        estimate = sum(a * b for a, b in zip(multiply(A, v), v))
        print(f"step {step:2}: direction ({v[0]:.4f}, {v[1]:.4f})  eigenvalue ~ {estimate:.4f}")`,
		Expected: `step  1: direction (0.8944, 0.4472)  eigenvalue ~ 2.8000
step  2: direction (0.7809, 0.6247)  eigenvalue ~ 2.9756
step  5: direction (0.7100, 0.7042)  eigenvalue ~ 3.0000
step 20: direction (0.7071, 0.7071)  eigenvalue ~ 3.0000`,
	},
	14: {
		File:    "line.py",
		Lang:    "Python",
		Run:     "python3 line.py",
		Predict: "The data is roughly y = 2x + 1. What slope and intercept will the formulas find, and what will the errors add up to?",
		Lesson:  "Least squares lands close to the truth despite the noise, and its errors always sum to (essentially) zero, which is a built-in property of the fit.",
		Code: `# LEAST SQUARES -- the best straight line through some points,
# using the textbook formulas (no libraries).

xs = [1, 2, 3, 4, 5]
ys = [3.1, 4.9, 7.2, 8.8, 11.0]          # roughly y = 2x + 1, with noise

n = len(xs)
mean_x = sum(xs) / n
mean_y = sum(ys) / n
slope = (sum((x - mean_x) * (y - mean_y) for x, y in zip(xs, ys)) /
         sum((x - mean_x) ** 2 for x in xs))
intercept = mean_y - slope * mean_x
print(f"best line: y = {slope:.2f}x + {intercept:.2f}")

errors = [y - (slope * x + intercept) for x, y in zip(xs, ys)]
print("errors:", [round(e, 2) for e in errors])
print("sum of errors:", round(sum(errors), 10))`,
		Expected: `best line: y = 1.97x + 1.09
errors: [0.04, -0.13, 0.2, -0.17, 0.06]
sum of errors: 0.0`,
	},
	15: {
		File:    "autograd.py",
		Lang:    "Python",
		Run:     "python3 autograd.py",
		Predict: "With w = 2, x = 3 and b = 1, what are L, dL/dw, dL/dx and dL/db? Work them out with the chain rule first.",
		Lesson:  "y = 7 and L = y² = 49, so dL/dy = 2y = 14; then dL/dw = 14·x = 42, dL/dx = 14·w = 28 and dL/db = 14. You have just done backpropagation by hand.",
		Code: `# AUTOGRAD -- a value that remembers how it was made, so gradients
# can flow backwards through the chain rule.

class Value:
    def __init__(self, data, parents=(), backward=lambda: None):
        self.data, self.grad = data, 0.0
        self.parents, self._backward = parents, backward

    def __add__(self, other):
        out = Value(self.data + other.data, (self, other))
        def backward():
            self.grad += out.grad
            other.grad += out.grad
        out._backward = backward
        return out

    def __mul__(self, other):
        out = Value(self.data * other.data, (self, other))
        def backward():
            self.grad += other.data * out.grad
            other.grad += self.data * out.grad
        out._backward = backward
        return out

    def backward(self):
        order, seen = [], set()
        def visit(v):
            if v not in seen:
                seen.add(v)
                for p in v.parents:
                    visit(p)
                order.append(v)
        visit(self)
        self.grad = 1.0
        for v in reversed(order):
            v._backward()

w, x, b = Value(2.0), Value(3.0), Value(1.0)
y = w * x + b          # y = 7
L = y * y              # L = 49
L.backward()
print("L =", L.data)
print("dL/dw =", w.grad, " dL/dx =", x.grad, " dL/db =", b.grad)`,
		Expected: `L = 49.0
dL/dw = 42.0  dL/dx = 28.0  dL/db = 14.0`,
	},
	16: {
		File:    "roofline.py",
		Lang:    "Python",
		Run:     "python3 roofline.py",
		Predict: "For a 7B model in 16-bit on a 100 GB/s laptop, what is the most tokens per second it can produce? And in 4-bit?",
		Lesson:  "Speed is bandwidth ÷ bytes per token: quantising to 4 bits makes the same model four times faster on the same hardware.",
		Code: `# ROOFLINE -- how fast can a language model possibly talk?
# At batch size 1, every generated token reads every weight once,
# so tokens per second <= memory bandwidth / model size.

models = [("7B, 16-bit", 7e9, 2.0), ("7B, 4-bit", 7e9, 0.5), ("70B, 4-bit", 70e9, 0.5)]
devices = [("laptop, 100 GB/s", 100e9), ("GPU, 1000 GB/s", 1000e9)]

for name, params, bytes_per_param in models:
    size_gb = params * bytes_per_param / 1e9
    ceilings = ", ".join(f"{dev}: {bw / (params * bytes_per_param):5.1f} tok/s"
                         for dev, bw in devices)
    print(f"{name:11} ({size_gb:4.1f} GB)  {ceilings}")`,
		Expected: `7B, 16-bit  (14.0 GB)  laptop, 100 GB/s:   7.1 tok/s, GPU, 1000 GB/s:  71.4 tok/s
7B, 4-bit   ( 3.5 GB)  laptop, 100 GB/s:  28.6 tok/s, GPU, 1000 GB/s: 285.7 tok/s
70B, 4-bit  (35.0 GB)  laptop, 100 GB/s:   2.9 tok/s, GPU, 1000 GB/s:  28.6 tok/s`,
	},
}
