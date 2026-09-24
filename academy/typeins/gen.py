"""Regenerate ../typeins.go from the listings in this folder.

Runs every listing (python3, gcc, sh and git required) and embeds its
source and real output, so the app never shows untested code.
Usage: python3 typeins/gen.py && gofmt -w typeins.go
"""
import subprocess, os, tempfile, json

def run(cmd, cwd=None):
    return subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd, check=True).stdout.rstrip("\n")

os.chdir(os.path.dirname(os.path.abspath(__file__)))
here = os.getcwd()
tb = os.path.join(tempfile.gettempdir(), "academy_tb")
tf = os.path.join(tempfile.gettempdir(), "academy_tf")
out = {}
for f in sorted(os.listdir(".")):
    if f.endswith(".py") and f[0] == "s":
        out[f] = run(f"python3 {f}")
out["s00_cpu.c"] = run(f"gcc -Wall -Wextra -std=c17 -o {tb} s00_cpu.c && {tb}")
out["s01_collatz.c"] = run(f"gcc -Wall -Wextra -std=c17 -o {tb} s01_collatz.c && {tb}")
out["s05_bits.c"] = run(f"gcc -Wall -o {tb} s05_bits.c && {tb}")
fork = run(f"gcc -Wall -o {tf} s06_fork.c && {tf}").splitlines()
out["s06_fork.c"] = "\n".join(f"hello from process {1234 + i}" for i in range(len(fork)))  # PIDs vary per run
d = tempfile.mkdtemp()
out["s09_git.sh"] = run(f"sh {here}/s09_git.sh", cwd=d)

# stage -> (file, saved name, run command, predict question, lesson)
T = {
 0: ("s00_cpu.c", "cpu.c", "gcc -Wall -Wextra -std=c17 cpu.c -o cpu && ./cpu", "Trace the program on paper before running it: what number will it print, and how many instructions will the tiny CPU execute before it halts? (Hint: the loop body is 7 instructions.)", "The loop runs 5 times × 7 instructions = 35, plus 3 to print and halt: 38 steps, printing 15. Notice that the program and its data share the same memory (the von Neumann design), and that a loop is nothing but a conditional jump back to an earlier address. A real CPU does exactly this, billions of times per second."),
 1: ("s01_collatz.c", "collatz.c", "gcc -Wall -Wextra -std=c17 collatz.c -o collatz && ./collatz", "How many steps does 27 take to reach 1, and what is the highest value it reaches on the way? Write your guesses down first. Why is n a long rather than an int?", "Tiny rules can produce surprisingly long journeys: 27 climbs past 9,000 before falling to 1. Nobody has proved that every starting number reaches 1 (the Collatz conjecture). n is a long because some starting numbers climb far beyond what a 32-bit int can hold, and in C a signed overflow is undefined behaviour. The size line depends on your machine: 8 bytes on 64-bit Linux and macOS, 4 on Windows."),
 2: ("s02_bsearch.py", "bsearch.py", "python3 bsearch.py", "In a sorted list of one million numbers, how many guesses does binary search need to find 1? 500,000? A number that is not there?", "About log₂(1,000,000) ≈ 20 guesses at most, and the exact middle is found on the very first guess."),
 3: ("s03_sieve.py", "sieve.py", "python3 sieve.py", "Roughly how many primes are there below 10,000? (Hint: the prime number theorem says about n / ln n.)", "10,000 / ln 10,000 ≈ 1,086, a decent estimate of the true 1,229. The sieve was a favourite benchmark in early-1980s computer magazines."),
 4: ("s04_adder.py", "adder.py", "python3 adder.py", "What does an 8-bit adder give for 200 + 100, and for 255 + 1? What happens to the carry?", "300 doesn't fit in 8 bits: the adder keeps 300 − 256 = 44 and raises the carry flag. This is overflow, built from nothing but AND, OR and XOR."),
 5: ("s05_bits.c", "bits.c", "gcc bits.c -o bits && ./bits", "Predict every line before running: 255 + 1 in 8 bits, 127 + 1 in signed 8 bits, whether 0.1 + 0.2 equals 0.3, and the two sizes.", "Integers wrap around and decimals are approximated in binary. The sizes shown are typical of modern 64-bit machines, and C allows them to differ elsewhere."),
 6: ("s06_fork.c", "fork.c", "gcc fork.c -o fork && ./fork", "How many lines will this print? (Each fork() clones every process that runs it.)", "Two forks give 2 × 2 = 4 processes, so 4 lines, each with a different process ID. The order can change from run to run, because the scheduler decides who goes first."),
 7: ("s07_index.py", "index.py", "python3 index.py", "What will the query plan say before and after the index is created?", "Without an index the database must SCAN every row; with one it SEARCHes the B-tree directly. The exact wording varies between SQLite versions (older ones say SCAN TABLE)."),
 8: ("s08_echo.py", "echo.py", "python3 echo.py", "What exactly will the client print?", "A real TCP connection on your own machine: bind, listen, accept on one side; connect, send, receive on the other."),
 9: ("s09_git.sh", "git_history.sh", "mkdir scratch && cd scratch && sh ../git_history.sh", "How many commits will the main line have, how many will the idea branch have, and how many lines will notes.txt have at the end?", "A branch is just a label on a commit. Switching back to the main line hides the experiment's change without losing it. (Needs Git 2.23+ for git switch.)"),
 10: ("s10_inject.py", "inject.py", "python3 inject.py", "Which users will each query return when the attacker types ' OR '1'='1 into both fields?", "Glued-together SQL lets input become code, so every user leaks. With parameters, the input is only ever data, and nobody is literally called ' OR '1'='1."),
 11: ("s11_regex.py", "match.py", "python3 match.py", "For each of the six lines, predict True or False before running.", "o.d doesn't match \"hello world\": no o is followed by one character and then a d. Tracing the recursion by hand is the best way to see why."),
 12: ("s12_lisp.py", "lisp.py", "python3 lisp.py", "What will each of the three programs evaluate to?", "Tokenise, parse, evaluate: the whole pipeline of a language in a few dozen lines. Adding define and lambda turns it into a real Lisp."),
 13: ("s13_power.py", "power.py", "python3 power.py", "Which direction will the vector settle on, and what eigenvalue will the estimate approach?", "Repeated multiplication amplifies the strongest direction, (1, 1)/√2 with eigenvalue 3. The same idea underlies Google's original PageRank."),
 14: ("s14_line.py", "line.py", "python3 line.py", "The data is roughly y = 2x + 1. What slope and intercept will the formulas find, and what will the errors add up to?", "Least squares lands close to the truth despite the noise, and its errors always sum to (essentially) zero, which is a built-in property of the fit."),
 15: ("s15_autograd.py", "autograd.py", "python3 autograd.py", "With w = 2, x = 3 and b = 1, what are L, dL/dw, dL/dx and dL/db? Work them out with the chain rule first.", "y = 7 and L = y² = 49, so dL/dy = 2y = 14; then dL/dw = 14·x = 42, dL/dx = 14·w = 28 and dL/db = 14. You have just done backpropagation by hand."),
 16: ("s16_roofline.py", "roofline.py", "python3 roofline.py", "For a 7B model in 16-bit on a 100 GB/s laptop, what is the most tokens per second it can produce? And in 4-bit?", "Speed is bandwidth ÷ bytes per token: quantising to 4 bits makes the same model four times faster on the same hardware."),
}

def gostr(s):
    assert "`" not in s
    return "`" + s + "`"

lines = []
for sid in sorted(T):
    f, name, cmd, predict, lesson = T[sid]
    code = open(f).read().rstrip("\n")
    lang = {"py": "Python", "c": "C", "sh": "Shell"}[f.rsplit(".", 1)[1]]
    lines.append(f"""\t{sid}: {{
\t\tFile:     {json.dumps(name)},
\t\tLang:     {json.dumps(lang)},
\t\tRun:      {json.dumps(cmd)},
\t\tPredict:  {json.dumps(predict, ensure_ascii=False)},
\t\tLesson:   {json.dumps(lesson, ensure_ascii=False)},
\t\tCode:     {gostr(code)},
\t\tExpected: {gostr(out[f])},
\t}},""")

src = """package main

// Code generated from the tested type-in programs; the listings and their
// expected output were produced by actually running each program. Do not
// edit by hand: change the program, re-run it, and regenerate.

// typeIns holds one magazine-style listing per stage.
var typeIns = map[int]TypeIn{
""" + "\n".join(lines) + "\n}\n"
open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "typeins.go"), "w").write(src)
print("wrote", len(T), "type-ins")
