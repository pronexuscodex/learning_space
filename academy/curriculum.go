package main

import "fmt"

// This file is the Study Hall knowledge base: static teaching content that
// ships inside the binary. It is deliberately kept out of the JSON registry,
// which only holds *your* progress (which concepts you have studied).
//
// Text conventions (see reflow in ui.go):
//   - a blank line starts a new paragraph;
//   - a line beginning with "- " is a bullet;
//   - any other line break is soft and gets re-wrapped to the terminal width.
// Diagrams are printed verbatim.

// Concept is one idea worth understanding deeply within a stage.
type Concept struct {
	Name        string
	Summary     string // one-line hook
	Body        string // the explanation
	Diagram     string // optional ASCII diagram
	MentalModel string // the sentence to remember
	TryIt       string // a small hands-on exercise

	// UnderTheHood goes one level down: what the machine actually does
	// (the bits, the instructions, the operating system) when this idea
	// runs. Optional; lines indented by two spaces are shown as code.
	UnderTheHood string

	// Beginner layer. Set inline for newer stages, or merged in from
	// explainers.go and exercises.go for the original ones.
	Analogy   string     // the idea in everyday terms, no jargon
	Example   string     // where this shows up in the real world
	Exercises []Exercise // warm-up → practice → real-world, each with a hint
}

// Exercise levels, in the order they are always presented.
const (
	LevelWarmUp    = "warm-up"
	LevelPractice  = "practice"
	LevelRealWorld = "real-world"
)

// Exercise is one graded practice task with a hint.
type Exercise struct {
	Level string
	Task  string
	Hint  string
}

// trio builds the standard three-exercise ladder for a concept:
// a warm-up (no code), a practice task, and a real-world project.
func trio(warm, warmHint, practice, practiceHint, real, realHint string) []Exercise {
	return []Exercise{
		{LevelWarmUp, warm, warmHint},
		{LevelPractice, practice, practiceHint},
		{LevelRealWorld, real, realHint},
	}
}

// Term is one piece of jargon explained in plain English.
type Term struct {
	Word    string
	Meaning string
}

// Resource is an external course, book, article, tool or video.
type Resource struct {
	Kind  string // Course, Book, Paper, Article, Video, Tool, Site
	Title string
	URL   string
	Note  string
}

// Blueprint is a suggested lab that can be enrolled with one keystroke.
type Blueprint struct {
	Name       string
	Brief      string
	Milestones []string
}

// Question is a self-check flashcard.
type Question struct {
	Q string
	A string
}

// StageGuide is the full Study Hall content for one stage.
type StageGuide struct {
	Overview   string
	Outcomes   []string // "after this stage you can…", from explainers.go
	Glossary   []Term   // from explainers.go
	Concepts   []Concept
	Resources  []Resource
	Blueprints []Blueprint
	Quiz       []Question
}

// curriculum is every stage's guide, keyed by stage ID. Each track's
// content lives in its own file.
var curriculum = mergeGuides(machineGuides, foundationsGuides, systemsAndAIGuides, softwareGuides, dataGuides, securityGuides)

// lastStageID is the highest stage number in the curriculum. Nothing else
// should assume how many stages there are.
func lastStageID() int {
	last := 0
	for id := range curriculum {
		last = max(last, id)
	}
	return last
}

// mergeGuides combines per-file guide maps; a duplicate stage ID is a
// programming error.
func mergeGuides(parts ...map[int]StageGuide) map[int]StageGuide {
	out := map[int]StageGuide{}
	for _, part := range parts {
		for id, g := range part {
			if _, dup := out[id]; dup {
				panic(fmt.Sprintf("curriculum: stage %d defined twice", id))
			}
			out[id] = g
		}
	}
	return out
}

// guideFor returns the Study Hall content for a stage, if any.
func guideFor(stageID int) (StageGuide, bool) {
	g, ok := curriculum[stageID]
	return g, ok
}

// systemsAndAIGuides holds Track A (stages 5–8) and the original Track B
// stages (13, 15, 16).
var systemsAndAIGuides = map[int]StageGuide{
	// -----------------------------------------------------------------
	5: {
		Overview: `Everything above this stage rests on how source code becomes machine
instructions, and how those instructions move bytes between registers,
caches and memory. You will follow a C program through the compiler,
the assembler, the linker and the loader; learn the calling convention
every compiled function obeys; see where integer and floating-point
arithmetic break; watch what the optimizer does to your code (and what
undefined behaviour lets it do); learn to measure performance honestly;
and write the parser at the front of every compiler. Each lesson is
checked against the real tools: gcc, objdump, nm, valgrind.`,
		Concepts: []Concept{
			{
				Name:    "The Compilation Pipeline",
				Summary: "Source code is lowered step by step until it is bytes the CPU can run.",
				Body: `A compiler is a chain of translations. The preprocessor pastes in every
#include and expands macros. The lexer turns characters into tokens. The
parser turns tokens into an abstract syntax tree (AST). Semantic
analysis resolves names and checks types. The AST is lowered to an
intermediate representation (IR), which the optimizer rewrites, and the
backend selects instructions and allocates registers to produce
assembly.

The assembler turns assembly into an object file: machine code plus a
symbol table and relocations (holes to patch). The linker merges object
files, resolves symbols across them and fixes up addresses. At run time
the loader maps the executable's segments into memory and jumps to the
entry point.

The command gcc is only a driver that runs these programs in turn. You
can stop it after any step and look at the result:

  gcc -E hello.c          preprocessed C, to the screen
  gcc -S hello.c          assembly, in hello.s
  gcc -c hello.c          an object file, hello.o
  gcc hello.o -o hello    linked into a program`,
				Diagram: `source.c ─▶ [lexer] ─▶ tokens ─▶ [parser] ─▶ AST ─▶ [IR gen] ─▶ IR
   ─▶ [optimizer] ─▶ [codegen] ─▶ .s ─▶ [assembler] ─▶ .o ─▶ [linker] ─▶ a.out`,
				UnderTheHood: `Following a six-line hello.c through gcc 13 on Linux:

  gcc -E hello.c | wc -l       819

The preprocessed file is 819 lines: #include <stdio.h> pasted in the
declarations of the whole standard I/O library before your six lines.
gcc -v shows the programs the driver runs: cc1 (the compiler proper),
as (the assembler) and collect2, which runs ld (the linker).

gcc -O2 -S hello.c shows the compiler already rewriting your code:

  lea   rdi, .LC0[rip]     ; the address of "hello, world"
  call  puts@PLT
  xor   eax, eax           ; return 0

printf("hello, world\n") became puts("hello, world"): a format string
with no % and a final newline means exactly what puts does, more
cheaply. The compiler may make any change that the C standard says
cannot be told apart from your program.`,
				MentalModel: "Each stage translates to a language that is slightly closer to the metal.",
				TryIt:       "Compile the same function with gcc -S -O0 and -O2 (or paste it into godbolt.org) and explain every instruction that disappeared.",
			},
			{
				Name:    "Linking, Loading & Shared Libraries",
				Summary: "The linker stitches object files into one program; the loader and the dynamic linker bring it to life.",
				Body: `Every object file carries a symbol table: the names it defines (its
functions and global variables) and the names it uses but does not
define. The linker's job is to match every use to exactly one
definition, then patch the addresses into the code (relocation).

Two errors follow from that rule, and every C programmer meets both:
- undefined reference to f: something calls f, but no object file or
  library given to the linker defines it;
- multiple definition of x: two object files both define a global x.

The cure for the second is linkage. A global variable or function is
visible to every file (external linkage) unless you mark it static,
which makes it private to its file (internal linkage). Shared state that
only one file needs should always be static.

Libraries come in two kinds. A static library (libm.a) is an archive of
object files; the linker copies the ones you need into your program. A
shared library (libc.so.6 on Linux, .dylib on macOS, .dll on Windows) is
loaded at run time by the dynamic linker (ld-linux.so on Linux), so many
programs share one copy in memory, and fixing a bug in the library fixes
every program that uses it. Calls to shared functions go through small
stubs (the PLT) and a table of addresses (the GOT) that the dynamic
linker fills in.`,
				Diagram: `  main.o                 libc.so.6
  ┌──────────────┐       ┌──────────────┐
  │ T main       │       │ T puts       │
  │ U puts ──────┼───────┼─▶            │   resolved when the program
  └──────────────┘       └──────────────┘   starts, by ld-linux.so
          │ ld (linker)
          ▼
        a.out: main's code, plus a PLT stub that jumps to puts`,
				UnderTheHood: `nm lists an object file's symbols. For hello.c compiled with -O2:

  nm hello.o
  0000000000000000 T main         ; T: defined here, in the code (text)
                   U puts         ; U: undefined, needed from elsewhere

and readelf -r shows the hole the linker must fill:

  000000000013  R_X86_64_PLT32   puts - 4

"at byte 0x13 of the code, write the distance to puts". After linking,
ldd lists the shared libraries the program will need:

  libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
  /lib64/ld-linux-x86-64.so.2

linux-vdso.so.1 also appears: a tiny library the kernel maps into every
process so that calls such as clock_gettime need no system call. Linked
normally, the program is 15,960 bytes; linked with -static, which copies
the C library in, it is 785,360 bytes.`,
				MentalModel: "Every name must be defined exactly once: static keeps a name private to its file, and the linker connects the rest.",
				TryIt:       "Compile hello.c with -c, run nm and readelf -r on hello.o, then link it and run ldd on the program. Then link it with -static and compare the sizes.",
				Analogy: `A book written by several authors, each chapter saying "see the table
in chapter 7". The linker is the editor who assembles the chapters and
fills in the real page numbers. A shared library is the public library
in town: instead of every household buying its own encyclopedia, everyone
borrows the same copy, and when it is corrected, everyone sees the fix.`,
				Example: `When the Heartbleed bug in OpenSSL was fixed in 2014, servers needed only
an updated libssl shared library and a restart, not new builds of every
program. "DLL hell" on Windows and missing .so errors on Linux are what
happens when a program cannot find the exact library it was linked
against. Tools such as Valgrind and faketime work by loading their own
code into a program with the dynamic linker.`,
				Exercises: trio(
					"In nm's output, what do T, U, D and B mean? Which letter would you expect for main, for printf, for an initialised global int and for a global array of 100 ints with no initialiser?",
					"T: code defined here; U: undefined, needed from elsewhere; D: initialised data; B: zeroed data (bss). main is T, printf is U, the initialised int is D, the array is B.",
					"Create util.c (a function and a counter) and main.c that uses them. Build util.c into a static library (ar rcs libutil.a util.o), link with -L. -lutil, then build it as a shared library (gcc -shared -fPIC -o libutil.so util.c) and run the program with LD_LIBRARY_PATH=. set. Compare the two programs with nm and ldd.",
					"With the static library, util's code is inside the program and ldd does not mention it; with the shared one, nm shows the function as U and ldd lists libutil.so.",
					"Count the mallocs of a program you did not write: build a shared library whose malloc counts calls and forwards to the real one, found with dlsym(RTLD_NEXT, \"malloc\"), and run ls with LD_PRELOAD pointing at it (Linux).",
					"Print the count from a function marked __attribute__((destructor)), which runs at exit. Careful: dlsym and printf may call malloc themselves, so guard against recursion with a flag.",
				),
			},
			{
				Name:    "Registers, the Stack & Calling Conventions",
				Summary: "The ABI is the contract that lets separately compiled code call each other.",
				Body: `The CPU computes on a small set of registers. The calling convention
(on Linux and macOS x86-64, the System V AMD64 ABI) fixes where
arguments and results go:
- integer and pointer arguments in rdi, rsi, rdx, rcx, r8, r9, and the
  rest on the stack;
- floating-point arguments in xmm0 to xmm7;
- the return value in rax (or xmm0 for floating point);
- rbx, rbp and r12 to r15 are callee-saved, so a function must restore
  them;
- the stack must be 16-byte aligned at every call instruction.

ARM64 (Apple silicon, phones, Raspberry Pi) has its own convention with
the same ideas: arguments in x0 to x7, the result in x0. Windows x64
uses yet another (rcx, rdx, r8, r9). That is why a library compiled for
one platform cannot be called from another.

The call instruction pushes the return address. A function prologue
reserves a frame for locals and spilled registers. Local buffers and the
return address share the stack, and that is exactly why buffer overflows
can hijack control flow (Stage 10).`,
				Diagram: `high addr ┌────────────────────┐
          │  caller's frame    │
          │  7th, 8th … args   │ ◀─ the arguments past the sixth
          ├────────────────────┤
          │  return address    │ ◀─ pushed by call
          │  saved rbp         │ ◀─ rbp (frame base)
          │  locals / spills   │
low addr  └────────────────────┘ ◀─ rsp   (stack grows down ↓)`,
				UnderTheHood: `A function with seven long arguments, compiled with gcc -O2:

  long seven(long a, long b, long c, long d, long e, long f, long g)
  { return a + b + c + d + e + f + g; }

  seven:
      add   rdi, rsi             ; a + b
      add   rdi, rdx             ; + c
      add   rdi, rcx             ; + d
      add   rdi, r8              ; + e
      lea   rax, [rdi+r9]        ; + f, into the return register
      add   rax, QWORD PTR 8[rsp] ; + g: the seventh comes from the stack
      ret

The first six arrive in registers exactly as the ABI lists them; the
seventh sits just above the return address, at rsp + 8. And a caller
compiled in the same file, long call7(void) { return seven(1, 2, 3, 4,
5, 6, 7); }, compiled to just mov eax, 28: the optimizer inlined the
call and added the constants itself (What the Optimizer Does).`,
				MentalModel: "A function call is a push, a jump, and an agreement about who saves which register.",
				TryIt:       "Write a function in x86-64 assembly that sums an int array, and call it from C.",
			},
			{
				Name:    "Integers & Floating Point in Bits",
				Summary: "Numbers are fixed-width bit patterns, and that has consequences.",
				Body: `Signed integers use two's complement: -x is ~x + 1, so the same adder
handles signed and unsigned arithmetic (Building an ALU, Stage 4).
Unsigned overflow wraps. In C, signed overflow is undefined behaviour,
and optimizers exploit that assumption.

IEEE 754 floats store a sign bit, a biased exponent and a mantissa: a
float has 1 + 8 + 23 bits, a double 1 + 11 + 52. The value is
(−1)^sign × 1.mantissa × 2^(exponent − bias), with bias 127 for float
and 1023 for double. So a float has about 7 significant decimal digits
and a double about 16. Most decimals cannot be represented exactly
(0.1 + 0.2 != 0.3), large floats cannot hold small changes, and float
addition is not associative: (a + b) + c can differ from a + (b + c).

Practical rules: never compare floats with ==; never store money in
floats (use integer cents); and remember that the order of a sum
matters, which returns in Stage 16, where a parallel GPU reduction sums
in a different order and can give slightly different results.`,
				Diagram: `float 0.1f = 0x3dcccccd
 sign   exponent      mantissa
  0     01111011      10011001100110011001101
  +     123 − 127 = −4   1.6000000238… × 2⁻⁴ = 0.100000001490116…`,
				UnderTheHood: `Printed from C on a test machine:

  0.1f is stored as        3dcccccd  (sign 0, exponent 123, mantissa 4ccccd)
  0.1 + 0.2                0.30000000000000004
  0.1 + 0.2 == 0.3         0 (false)
  (1e16 + -1e16) + 1.0     1
  1e16 + (-1e16 + 1.0)     0      ; the 1.0 is lost next to 1e16
  1e8f + 1.0f == 1e8f      1      ; a float cannot hold 100000001
  0.10 added 1000 times    99.999999999998593

The last line is why money is counted in integer cents: every 0.10 is
a little off, and the errors add up. In the second sum, −1e16 + 1.0
rounds back to −1e16, because a double has only about 16 significant
digits, so the 1 vanishes before it can be added.`,
				MentalModel: "Every number type is a finite code; learn where the code breaks.",
				TryIt:       "Print the bit patterns of -1, INT_MIN, 0.1f and 1e30f + 1.0f - 1e30f, and explain each one.",
			},
			{
				Name:    "The Memory Hierarchy & Locality",
				Summary: "Where your data lives often matters more than how many instructions you run.",
				Body: `Rough latencies on a modern CPU: register ~1 cycle, L1 ~4 cycles,
L2 ~12, L3 ~40, DRAM hundreds of cycles (Stage 4 measured them: 1.6 ns
in L1 against 207 ns in DRAM on one machine). Caches move data in
64-byte lines, so touching one byte brings in its 63 neighbours.

Programs are fast when they reuse data soon (temporal locality) and use
neighbouring data (spatial locality). C stores a 2D array row by row
(row-major), so walking it column by column touches a new cache line on
every access, with identical instruction counts. The same idea returns
in Stage 16 as the roofline model and GPU memory coalescing.

Designing for the cache:
- loop over arrays in the order they are stored;
- prefer arrays of values to linked structures of pointers;
- keep the data a hot loop needs together ("struct of arrays" when a
  loop uses only some fields of many records);
- work on blocks that fit in cache (tiling), as fast matrix
  multiplication does.`,
				Diagram: `  CPU ─ regs ─ L1 (32K) ─ L2 (1M) ─ L3 (tens of MB) ─ DRAM (GBs) ─ SSD
  fast · tiny ◀──────────────────────────────────────────────▶ slow · huge`,
				UnderTheHood: `Summing a 4096 × 4096 int matrix (64 MB) with gcc -O2 on a test
machine, best of three runs:

  row by row          7.8 ms
  column by column  155   ms     20 times slower, same instructions

Valgrind's cache simulator (valgrind --tool=cachegrind) counts why, on a
1024 × 1024 matrix (4 MB):

  row by row        66,915 L1 read misses      one per 64-byte line
  column by column  1,049,955 L1 read misses   almost one per access

Row order misses once per 16 ints, because a line holds 16 of them;
column order jumps 4 KB between accesses and misses nearly every time.
The instruction counts of the two loops are the same, about 5.2 million
each: counting instructions would never have found this.`,
				MentalModel: "Memory is a pyramid; performance is how rarely you fall down it.",
				TryIt:       "Time row-major vs column-major traversal of a 4096×4096 int matrix, then run both under valgrind --tool=cachegrind on a smaller one and compare the D1 misses.",
			},
			{
				Name:    "What the Optimizer Does",
				Summary: "-O2 rewrites your program: it inlines, folds, deletes and vectorises, keeping only what the C standard promises.",
				Body: `The C standard describes an abstract machine, and the compiler may
produce any code that behaves the same as far as the standard can tell
(the "as if" rule). With -O2, gcc and clang use that freedom fully:
- inlining: a small function's body replaces the call;
- constant folding and propagation: work on known values is done at
  compile time;
- dead code elimination: results nobody uses, and stores nobody reads,
  disappear;
- loop optimisations: invariant work moves out of the loop, loops are
  unrolled, and at -O3 (or with the right flags) vectorised with SIMD;
- strength reduction: a multiplication in a loop becomes an addition.

Undefined behaviour is where this bites. The compiler assumes your
program never does anything undefined: no signed overflow, no
out-of-bounds access, no use of a pointer after free, no NULL
dereference. It optimises on that assumption, so a check written after
the undefined step, such as if (x + 1 < x) to detect overflow, can be
deleted as impossible.

Practical rules: debug at -O0 (or -Og), ship at -O2; turn on the
sanitizers while testing; write overflow checks that cannot overflow
(__builtin_add_overflow, or compare with INT_MAX - b first); and in a
benchmark, use the result, or the work may vanish.`,
				Diagram: `static int square(int x) { return x * x; }
int fold(void) { return square(7) + 1; }

   -O0:  call square ; add eax, 1 ; ret        (as written)
   -O2:  mov eax, 50 ; ret                      (inlined, then folded)`,
				UnderTheHood: `What gcc 13 -O2 made of a few small functions:

  int fold(void) { return square(7) + 1; }      mov eax, 50
  long call7(void) { return seven(1,…,7); }     mov eax, 28
  int always(int x) { return x + 1 > x; }       mov eax, 1
  int always_u(unsigned x) { return x + 1 > x; } cmp edi, -1 ; setne al

For a signed int, x + 1 > x "cannot" be false, because the only case
where it would be (x = INT_MAX) overflows, which is undefined. For an
unsigned int, overflow is defined to wrap, so the compiler must test.

A loop doubling an array of floats stayed one float at a time at -O2
(gcc 13) and became movups / addps, four floats per instruction, at
-O3. For a loop summing 1 to n, gcc unrolled it, and clang replaced the
loop with the formula n(n + 1)/2: no loop at all.

And a security trap: a function that clears a password with memset just
before returning compiled at -O2 with no memset at all. The array is
never read again, so the store is dead. explicit_bzero (glibc and the
BSDs) is guaranteed to stay, and it does: the -O2 code calls
__explicit_bzero_chk.`,
				MentalModel: "The optimizer keeps what the standard promises, nothing more: write code whose meaning is defined, and measure what you think it does.",
				TryIt:       "Paste fold, always and always_u into Compiler Explorer (godbolt.org), compare -O0 with -O2, then add -fsanitize=undefined and call always(INT_MAX).",
				Analogy: `An editor who rewrites your essay to be half as long, promising it says
exactly the same thing. They keep that promise, as long as you did not
rely on anything the rules of grammar forbid; if you did, they may cut
the paragraph you were counting on.`,
				Example: `In 2009 a Linux kernel driver (the tun driver) read a field through a
pointer and only then checked the pointer for NULL. gcc reasoned that a
pointer already dereferenced cannot be NULL and deleted the check, which
turned a crash into an exploitable hole (CVE-2009-1897). The kernel now
builds with -fno-delete-null-pointer-checks. Benchmarks whose result is
never used are another classic: the optimizer deletes the work, and the
program appears infinitely fast.`,
				Exercises: trio(
					"Predict what gcc -O2 makes of int f(void) { int x = 6; return x * 7; }, and of a loop that adds 1 to a local variable a million times and never uses it.",
					"mov eax, 42 (constant folding). The loop disappears entirely: its result is never used, so it is dead code.",
					"Write a function that sums an array of floats. Compare its assembly at -O2 and -O3 (look for addps or vaddps), then add -ffast-math and explain why the compiler needed permission before it would vectorise the sum.",
					"Vectorising adds the floats in a different order, and float addition is not associative, so the result could change. -ffast-math (or -fassociative-math) gives the compiler permission to reorder.",
					"A security review says a function clears a password with memset before returning, but the compiled program contains no memset. Reproduce it with gcc -O2 -S, then fix it with explicit_bzero (or memset_s where available) and confirm the clearing is still there at -O2.",
					"The buffer is never read after the memset, so dead-store elimination removes it. explicit_bzero is defined so that it cannot be removed.",
				),
			},
			{
				Name:    "Profiling & Measuring Performance",
				Summary: "Measure before you optimise, and know exactly what your measurement measured.",
				Body: `Optimising without measuring is guessing. The first step is always to
find where the time goes, because it is rarely where you think, and by
Amdahl's law (Stage 4) speeding up anything else barely helps.

What to measure:
- wall-clock time (what a user waits) against CPU time (user and system),
  which time ./program prints as real, user and sys; a big gap means
  waiting: disk, network, locks or other programs;
- the steady state against the first run: the first run pays for page
  faults, cold caches and file reads;
- the spread: repeat, and report the minimum or the median, not one run.

Tools:
- a timer around the code: clock_gettime(CLOCK_MONOTONIC), never the
  wall-calendar clock, which can jump;
- sampling profilers (perf on Linux, Instruments on macOS) interrupt the
  program thousands of times a second and record where it was: cheap,
  and good for "which function is hot";
- instrumenting profilers and simulators (gprof, valgrind's callgrind
  and cachegrind) count every call, instruction or cache miss: exact,
  but 20 to 100 times slower;
- flame graphs show sampled call stacks, so you can see at a glance
  which paths through the program take the time.

Benchmark pitfalls: the optimizer deleting unused work, CPU frequency
changing under load, other programs on the machine, and inputs too small
to show the real growth.`,
				Diagram: `  time ./program
  real   2.00s   ← wall clock: what the user waited
  user   0.50s   ← CPU time running your code
  sys    0.10s   ← CPU time in the kernel for you
         ─────
         1.40s waiting: disk, network, locks, sleep`,
				UnderTheHood: `The first run lies. A C program that read and incremented one byte in
every 64 of a freshly malloc'd 256 MB array, three times over, measured
on a test machine (a virtual machine, where page faults are especially
expensive):

  pass 1   1,528.9 ms    131,070 page faults
  pass 2      27.2 ms          0 page faults
  pass 3      26.9 ms          0 page faults

malloc only reserved the addresses; the first touch of each 4 KiB page
faulted into the kernel twice (the read mapped a shared page of zeros,
the write then gave the page its own memory). Timing only the first pass
would have blamed the loop for the kernel's work.

Instruction counts lie too. callgrind counted about 5.2 million
instructions for both the row-order and the column-order matrix sums of
the last lesson, which differ 20 times in speed: only cachegrind's
cache-miss counts explain it. clock_getres reported a 1 ns resolution
for CLOCK_MONOTONIC there, but each call itself took about 20 ns:
time whole loops, never single operations.`,
				MentalModel: "Profile first, change one thing, measure again, and trust the steady state, not the first run.",
				TryIt:       "Run time on one of your programs with a large input, then valgrind --tool=callgrind and callgrind_annotate on the same program with a smaller one, and find the function with the most instructions.",
				Analogy: `A doctor runs tests before operating: you do not remove the organ you
guess is sick. And a stopwatch at a race measures the runners, not how
long the crowd took to sit down: decide what you are timing before you
press start.`,
				Example: `Donald Knuth's line "premature optimization is the root of all evil"
(1974) continues: "yet we should not pass up our opportunities in that
critical 3%", and finding that 3% is what profilers are for. Brendan
Gregg invented flame graphs in 2011 while working on performance
problems at Joyent, and later used them at Netflix; the same picture,
under names such as flame chart, is now in tools from Linux perf to
browser developer tools.`,
				Exercises: trio(
					"time ./program prints real 2.0s, user 0.5s and sys 0.1s. What was the program doing for the other 1.4 seconds, and would a faster CPU help?",
					"Waiting: for the disk, the network, a lock or a sleep (or for other programs to give it the CPU). A faster CPU would hardly help; the waiting is the problem.",
					"Write a small benchmark harness in C: run a function once to warm up, then five times, timing each with clock_gettime(CLOCK_MONOTONIC), and print the minimum and the median. Make sure the compiler cannot delete the work, then use it on the row-order and column-order matrix sums.",
					"Sort the five times to find the median. Use the result: accumulate it into a global, or assign it to a volatile variable, so the work is not dead code.",
					"Profile a real program: run valgrind --tool=callgrind on your wordfreq (Stage 1) or edit_distance (Stage 2) solution with a large input, open the result with callgrind_annotate, speed up the function with the most instructions, and measure again with your harness.",
					"Compile with -g so callgrind_annotate can show source lines, and keep -O2, so you profile the code that will actually run.",
				),
			},
			{
				Name:    "Parsing: Recursive Descent & Pratt",
				Summary: "Turning a flat token stream into a tree that respects precedence.",
				Body: `A recursive-descent parser has one function per grammar rule, and each
function consumes tokens and returns an AST node. For arithmetic:

  expr   = term   (('+' | '-') term)*
  term   = factor (('*' | '/') factor)*
  factor = NUMBER | '(' expr ')' | '-' factor

expr calls term, term calls factor, and factor calls expr again for a
parenthesised expression: the recursion in the code mirrors the nesting
in the input. Looping (rather than recursing) for the repeated part
makes - and / group from the left, so 10 - 4 - 3 is (10 - 4) - 3.

Operator precedence is the hard part when there are many levels. Pratt
parsing (precedence climbing) solves it neatly: each token gets a
binding power, and a loop keeps extending the left expression while the
next operator binds more tightly than the current level.

Once you have an AST, everything else (type checking, interpretation,
code generation) is a tree walk.`,
				Diagram: `  2 + 3 * 4 - 1             expr
                           ┌──┴──┐
                         expr  '-'  1
                        ┌──┴──┐
                        2 '+' term
                             ┌─┴─┐
                             3 '*' 4`,
				UnderTheHood: `gcc and clang both parse C with hand-written recursive-descent parsers,
so nesting really does cost stack. Feeding each compiler
int f(void) { return ((…(1)…)); } with more and more parentheses:

  depth       gcc 13                        clang
  200         compiles                      compiles
  1,000       compiles                      error: bracket nesting level
                                            exceeded maximum of 256
  100,000     internal compiler error:      (same error)
              Segmentation fault

Clang protects itself with a limit (raise it with -fbracket-depth=N).
gcc has none, and its recursion overflowed cc1's own stack: run with
ulimit -s unlimited, the same gcc compiled the 100,000-deep file without
complaint. Your parser faces the same choice: the calc check feeds yours 500 levels of parentheses,
deep enough to matter, but well within an 8 MiB stack.`,
				MentalModel: "Grammar rules become functions; precedence becomes a number.",
				TryIt:       "Write a Pratt parser for + - * / ^ and unary minus, and print the fully parenthesised tree.",
			},
		},
		Resources: []Resource{
			{"Book", "Computer Systems: A Programmer's Perspective (CS:APP)", "https://csapp.cs.cmu.edu/", "Its labs (bomb, attack, cache, malloc) are some of the best exercises available."},
			{"Book", "Crafting Interpreters (free online)", "https://craftinginterpreters.com/", "Builds a tree-walker and a bytecode VM from scratch."},
			{"Course", "Nand2Tetris", "https://www.nand2tetris.org/", "From NAND gates to a working computer, assembler, VM and compiler."},
			{"Tool", "Compiler Explorer", "https://godbolt.org/", "Shows the assembly for any snippet, across many compilers and flags."},
			{"Video", "Ben Eater: 8-bit breadboard computer", "https://eater.net/8bit", "Shows at the level of wires what a clock, a bus and microcode do."},
			{"Site", "x86 instruction reference", "https://www.felixcloutier.com/x86/", "A searchable copy of the Intel manual."},
			{"Site", "Agner Fog's optimization manuals", "https://www.agner.org/optimize/", "Instruction latencies and microarchitecture details."},
		},
		Blueprints: []Blueprint{
			{"Tree-Walking Interpreter", "A Lox-style language with a lexer, Pratt parser, environments, closures and classes.",
				[]string{"Lexer with line numbers in errors", "Pratt expression parser", "Statements and scoping", "Closures", "Classes and inheritance"}},
			{"Tiny C Compiler to x86-64", "Compile a C subset (ints, functions, if/while, pointers) to GAS assembly that follows the System V ABI.",
				[]string{"return <const>", "Arithmetic expressions", "Local variables on the stack", "Control flow", "Function calls following the ABI", "Pointers and arrays"}},
			{"Cache Latency Probe", "A pointer-chasing benchmark over growing working sets that shows the L1/L2/L3/DRAM cliffs.",
				[]string{"Randomised pointer chain", "Timing harness with warm-up", "Plot latency against working-set size", "Identify each cache size"}},
		},
		Quiz: []Question{
			{"Why is column-by-column traversal of a C 2D array slower than row-by-row?", "C arrays are row-major. Column traversal jumps a whole row per access, so each access touches a new 64-byte cache line and wastes the rest of it."},
			{"What does the linker do that the compiler cannot?", "It resolves symbols across separately compiled object files and patches addresses (relocations) into one final layout."},
			{"Under System V AMD64, where is the first integer argument and where is the return value?", "The first argument is in rdi; the return value is in rax."},
			{"Why can an optimizer assume that x + 1 > x for a signed int in C?", "Signed overflow is undefined behaviour, so the compiler may assume it never happens."},
			{"What do the linker errors \"undefined reference\" and \"multiple definition\" mean?", "Undefined reference: something uses a name that no object file or library defines. Multiple definition: two object files both define the same global name; making one of them static (private to its file) fixes it."},
			{"What is the difference between a static and a shared library?", "A static library's code is copied into the program at link time. A shared library is loaded at run time by the dynamic linker and shared by every program that uses it, so an update to it reaches all of them."},
			{"Where does the seventh integer argument of a function go on x86-64 Linux?", "On the stack: the first six go in rdi, rsi, rdx, rcx, r8 and r9, and the rest are pushed by the caller, just above the return address."},
			{"Why does 0.1 + 0.2 not equal 0.3 in floating point?", "None of the three has an exact binary representation, so each is rounded, and the rounded sum of two of them is not the rounded third."},
			{"Why can a memset that clears a password disappear at -O2?", "The buffer is never read afterwards, so the store is dead and the optimizer removes it. explicit_bzero is guaranteed to stay."},
			{"Why should a benchmark warm up before timing?", "The first run pays one-off costs, such as page faults, cold caches and loading files, that the steady state does not."},
			{"Why can two loops with the same instruction count differ 20 times in speed?", "They touch memory in different orders: one uses each cache line fully, the other misses the cache on almost every access."},
		},
	},

	// -----------------------------------------------------------------
	6: {
		Overview: `The operating system is the program that shares the hardware among
programs that do not trust each other. Learn how it gives every process
a private computer: its own address space, backed lazily by pages; a
heap that malloc carves out of memory the kernel hands over; files,
pipes and sockets behind small integers; signals that interrupt it; a
share of the CPU, switched in and out by the scheduler; and the rules
that keep threads from corrupting what they share. Every lesson is
measured with strace, /proc and C programs on a real Linux system.`,
		Concepts: []Concept{
			{
				Name:    "Processes & the Address Space",
				Summary: "Every process believes it owns all of memory.",
				Body: `A process is a running program: an address space, one or more threads,
and kernel state such as open files, the current directory and
credentials. Each process gets a private virtual address space with a
conventional layout. ASLR randomises the base addresses so attackers
cannot predict them.

Three system calls build every process tree:
- fork() clones the calling process; both copies return from it, the
  child with 0 and the parent with the child's PID;
- exec() (execvp and friends) replaces the process's program with a new
  one, keeping its PID and open files;
- wait() (waitpid) collects a child's exit status; until then, a child
  that has exited lingers as a "zombie".

A shell runs a command by forking, exec-ing the command in the child,
and waiting in the parent. The exit status packs either the exit code
or the signal that killed the process; WIFEXITED, WEXITSTATUS and
WTERMSIG unpack it.`,
				Diagram: `0x7fff… ┌───────────────────┐
        │ stack        ↓    │  locals, return addrs
        │        …          │
        │ mmap region       │  shared libs, mapped files
        │        …          │
        │ heap         ↑    │  malloc / brk
        │ .bss              │  zeroed globals
        │ .data             │  initialised globals
        │ .text             │  code (r-x)
0x0000… └───────────────────┘  unmapped: NULL deref faults`,
				UnderTheHood: `/proc/self/maps lists the current process's address space. Running
grep for the heap and the stack twice gave different addresses each
time:

  run 1   5605c0adb000-…  [heap]     7ffcbbcfb000-…  [stack]
  run 2   55f624ba2000-…  [heap]     7ffeac7ad000-…  [stack]

That is ASLR: every run gets a randomised layout.

Processes are not free. A C program that started /bin/true 1,000 times
measured, on a test machine (a virtual machine):

  fork + exec + waitpid      1,554 µs per program
  fork + waitpid (no exec)     257 µs per child

Most of the cost is exec: loading a new program, mapping its libraries
and starting it from scratch. fork alone is cheaper because it copies
only the page tables, not the memory, which is shared copy-on-write
(next lesson).`,
				MentalModel: "A process is an illusion of a private computer, maintained by the kernel.",
				TryIt:       "Run cat /proc/self/maps twice and compare the addresses (ASLR in action).",
			},
			{
				Name:    "Virtual Memory & Paging",
				Summary: "The MMU translates every address, and the kernel handles its failures.",
				Body: `Virtual addresses are translated to physical ones through page tables
(on x86-64: 4 levels and 4 KiB pages, with 2 MiB and 1 GiB huge pages).
The TLB caches recent translations; a TLB miss costs a walk of several
memory accesses.

When a translation is missing or not permitted, the CPU raises a page
fault and the kernel decides what to do. That one mechanism implements:
- lazy allocation (memory is only backed when first touched);
- copy-on-write fork (parent and child share pages until one writes);
- memory-mapped files and swapping;
- segfaults, when the access is truly invalid.

So two numbers describe a process's memory: its virtual size (VmSize,
every address it has reserved) and its resident set (VmRSS, the pages
actually in RAM). They can differ enormously, and only the second costs
RAM.`,
				Diagram: `virtual address 0x00007f3a12345678
  ┌─────────┬─────────┬─────────┬─────────┬──────────────┐
  │ L4 idx  │ L3 idx  │ L2 idx  │ L1 idx  │ offset (12b) │
  └────┬────┴────┬────┴────┬────┴────┬────┴──────────────┘
       ▼         ▼         ▼         ▼
     table  ─▶ table  ─▶ table  ─▶ table ─▶ physical page + offset
   (a TLB hit skips all four lookups)`,
				UnderTheHood: `Measured from /proc/self/status inside a C program:

  at start                   VmSize    2,696 kB   VmRSS    1,412 kB
  after malloc(1 GiB)        VmSize 1,051,276 kB  VmRSS    1,548 kB
  after writing one byte
  in every 4 KiB page        VmSize 1,051,276 kB  VmRSS 1,050,120 kB

malloc reserved a gigabyte of addresses and used almost no memory. The
loop that touched each page caused 262,143 minor page faults (by
getrusage): one per page, each one the kernel finding a free page of
RAM, zeroing it and mapping it in. That is why a program can "allocate"
more than the machine has, and why Linux's out-of-memory killer acts
only later, when the pages are actually touched.`,
				MentalModel: "Memory is a lookup table that the kernel fills in only when you trip over a hole.",
				TryIt:       "malloc 1 GiB, check VmRSS in /proc/self/status, touch every 4096th byte, and check again.",
			},
			{
				Name:    "Memory Allocators: How malloc Works",
				Summary: "malloc is ordinary library code that carves big chunks from the kernel into small blocks, and fragmentation is its enemy.",
				Body: `The kernel only hands out memory in pages, through two system calls:
brk moves the end of the heap, and mmap maps a fresh region anywhere.
malloc, in the C library, turns those pages into blocks of any size:
- each block has a small header recording its size, just before the
  pointer you get, so free knows how much to release;
- freed blocks go on free lists, often one per size class, to be reused;
- neighbouring free blocks are merged (coalesced), so the heap does not
  splinter into pieces too small to use;
- large requests (128 KiB by default in glibc) get their own mmap, so
  freeing them returns memory to the kernel at once.

Fragmentation is the cost: free memory scattered in small holes that no
request fits. Allocators fight it with size classes and coalescing, and
programs fight it by allocating fewer, larger blocks.

For many short-lived objects there is a simpler design: an arena (or
bump allocator). Take one big block, hand out pieces by advancing a
pointer, and free everything at once. Compilers, games and web servers
use arenas per compilation, frame or request: allocation is a few
instructions and there is nothing to leak.`,
				Diagram: `  arena: one block, a bump pointer, freed all at once

  ┌──────┬─────────┬───┬──────────┬───────────────────────┐
  │ a 24 │ b 100   │ c │ d 64     │        free           │
  └──────┴─────────┴───┴──────────┴───────────────────────┘
                                  ▲ next allocation starts here
                                    (rounded up to 16 bytes)`,
				UnderTheHood: `strace -e trace=brk,mmap,munmap on a C program shows what malloc asks
the kernel for (glibc 2.39):

  1,000 × malloc(100)    one brk call, growing the heap by 135,168 bytes
  malloc(1 MiB)          mmap(NULL, 1052672, …)   its own mapping
  free of that block     munmap(…, 1052672)       straight back

A thousand small allocations cost a single system call: malloc took a
large slice and carved it up itself. The 1 MiB block got a private
mapping of 1 MiB plus one 4 KiB page for its header, and went back to
the kernel the moment it was freed.

Each 100-byte block really occupied 112 bytes (consecutive blocks were
112 bytes apart; malloc_usable_size reported 104 usable). And glibc
adapts: after that 1 MiB block was freed, it raised its mmap threshold,
so a later 129 KiB request came from the heap with brk, not from mmap.`,
				MentalModel: "malloc is a retailer: it buys memory wholesale from the kernel and sells it by the byte.",
				TryIt:       "Run a program that mallocs 1,000 small blocks and one large one under strace -e trace=brk,mmap,munmap, and match each system call to the line that caused it.",
				Analogy: `A wholesale warehouse and a corner shop. The kernel sells memory only by
the pallet (pages); malloc is the shopkeeper who buys pallets and sells
single items, keeps returned items on the shelf to sell again, and only
sends a whole pallet back when it was never split. An arena is a
packed lunch: you take everything you need in one box and throw the box
away at the end of the day.`,
				Example: `Firefox, Redis and many game engines replace the default malloc with
allocators such as jemalloc or mimalloc for speed and lower
fragmentation, and Google's tcmalloc runs much of its server fleet.
Compilers (including gcc's obstacks) and Go's and Zig's standard
libraries use arenas. Heap exploits (Stage 10) work by corrupting the
headers and free lists described here.`,
				Exercises: trio(
					"A program mallocs blocks of 24, 100 and 7 bytes in a row, and the allocator rounds every block, header included, up to a multiple of 16 with an 8-byte header. How many bytes of heap do the three use?",
					"24 + 8 = 32; 100 + 8 = 108, rounded to 112; 7 + 8 = 15, rounded to 16: 160 bytes for 131 asked for.",
					"In C, write an arena: struct arena { char *base; size_t size, used; }, arena_init (one malloc), arena_alloc(a, n) that returns 16-byte-aligned pointers or NULL when full, and arena_reset. Use it to build a linked list of a million nodes and free them all with one reset.",
					"Round used up with (used + 15) & ~(size_t)15 before handing out a block. Check that n plus the padding fits in what is left before advancing.",
					"Write a tiny malloc replacement with a free list: one big region from mmap, a header of the block size before each block, first-fit search, splitting of large free blocks and coalescing of neighbours on free. Test it with random malloc/free sequences and compare its speed with the C library's.",
					"Keep the free list sorted by address, so coalescing only has to look at the previous and next free blocks.",
				),
			},
			{
				Name:    "System Calls & Privilege",
				Summary: "The one controlled doorway from user code into the kernel.",
				Body: `User code runs in an unprivileged CPU mode. To open a file or send a
packet it executes the syscall instruction, which switches to kernel mode
at a fixed entry point with arguments in registers. The kernel validates
everything (pointers from user space are hostile until proven otherwise),
does the work and returns.

A syscall costs on the order of a hundred nanoseconds, more with
speculative-execution mitigations, and far more if it has to wait for a
disk or the network. That cost is why buffered I/O exists (fwrite and
printf collect bytes and make one write call per buffer) and why
batching interfaces such as io_uring matter.

The C library wraps each syscall in a function (read, write, open,
fork); errors come back as -1 with the reason in errno. strace shows
every syscall a program makes, with its arguments and result: the
single most useful tool for finding out what a program is really doing.`,
				Diagram: `user mode          │  kernel mode
                   │
read(fd, buf, n) ──┼─▶ syscall entry ─▶ check fd, check buf
  (libc wrapper)   │                    copy data into buf
     ◀─────────────┼── return n (or -1 and errno)`,
				UnderTheHood: `Timed in C on a test machine (a virtual machine):

  syscall(SYS_getppid)          about 101 ns per call
  an ordinary function call     about 0.9 ns

A hundred times the cost of a function call, for a syscall that does
almost nothing. So the block size of I/O decides its speed. Copying a
1 MB file with read and write:

  1-byte buffer       about 2 million syscalls     374 ms
  4,096-byte buffer   258 reads, 256 writes          4.6 ms

(strace -c counted the calls; the extra reads are the end-of-file read
and the dynamic linker loading libc.) With 64 KiB blocks it was no
faster, 5.5 ms: once the syscalls are few, starting the program
dominates.`,
				MentalModel: "A syscall is a function call across a trust boundary, so it is checked and not free.",
				TryIt:       "Run strace -c on ls and on a simple C program, and compare the syscall counts.",
			},
			{
				Name:    "File Descriptors, Files & Pipes",
				Summary: "Files, pipes, sockets and terminals all hide behind small integers, and redirection is just rearranging them.",
				Body: `A file descriptor is a small integer that indexes the process's table
of open files. open returns the lowest free number; read, write and
close take it. By convention 0 is standard input, 1 standard output and
2 standard error. The same calls work on files, pipes, sockets and
terminals: "everything is a file" means everything is a descriptor.

Three calls make the shell possible:
- dup2(old, new) makes new refer to the same open file as old: that is
  how > file and 2>&1 work, by rearranging descriptors before exec;
- pipe(fds) creates a one-way channel: bytes written to fds[1] come out
  of fds[0], held in a kernel buffer;
- descriptors survive exec, so a child started by fork and exec inherits
  them (unless opened with O_CLOEXEC).

So ls | wc -l is: create a pipe; fork a child that dup2s the pipe's
write end onto 1 and execs ls; fork another that dup2s the read end
onto 0 and execs wc; close every pipe end in the parent; wait for both.
Forgetting to close an unused write end is the classic bug: the reader
never sees end-of-file, because a writer still exists, and it waits
forever.`,
				Diagram: `ls | wc -l

  ls  fd 1 ──▶ [ pipe: kernel buffer, 64 KiB ] ──▶ fd 0  wc -l
  fd 2 ──▶ terminal                                fd 1 ──▶ terminal`,
				UnderTheHood: `ls -l /proc/self/fd < /dev/null | cat shows ls's own descriptors:

  0 -> /dev/null
  1 -> pipe:[21586]          ; the pipe to cat
  2 -> pipe:[21586]
  3 -> /proc/7323/fd          ; the directory ls itself opened

strace -f shows the shell implementing echo hi > out.txt:

  openat(AT_FDCWD, "out.txt", O_WRONLY|O_CREAT|O_TRUNC, 0666) = 3
  dup2(3, 1)                 ; standard output is now out.txt
  …                          ; echo writes to fd 1
  dup2(10, 1)                ; put back the original, saved in fd 10

and fcntl(fds[1], F_GETPIPE_SZ) on a fresh pipe reported 65,536 bytes:
a writer can get 64 KiB ahead of its reader before write blocks.`,
				MentalModel: "A descriptor is a number in a table; redirection is copying one entry over another before exec.",
				TryIt:       "Run ls -l /proc/self/fd, then ls -l /proc/self/fd > /tmp/x and cat /tmp/x: which entry changed, and why?",
				Analogy: `The numbered lockers at a station. Each locker (descriptor) holds a
ticket to something: a file, a pipe, the terminal. Programs only ever
say "put this in locker 1"; the shell, before the program arrives,
decides what locker 1 actually leads to.`,
				Example: `Every pipeline you type in a shell (grep error log | sort | uniq -c),
every web server's sockets and every log rotation relies on
descriptors. The classic "too many open files" error is a process
running out of descriptors (ulimit -n), usually because something opens
files or sockets and never closes them.`,
				Exercises: trio(
					"A program has descriptors 0, 1 and 2 open and then calls open twice and close(3) once, then open again. Which numbers did the three opens return?",
					"3, 4, then 3 again: open always returns the lowest free descriptor.",
					"In C, run ls -l | wc -l yourself: create a pipe, fork two children, dup2 the right ends onto 1 and 0, execvp the two commands, close every unused end, and wait for both. Then remove one close and watch the program hang.",
					"The parent must close both ends too. If any copy of the write end stays open, wc never sees end-of-file.",
					"Find a descriptor leak: write a program that opens a file in a loop without closing it, run it until open fails, and print errno with strerror. Then check the limit with ulimit -n and watch the count grow in ls /proc/<pid>/fd.",
					"open fails with EMFILE (\"Too many open files\") at the soft limit, often 1,024.",
				),
			},
			{
				Name:    "Signals",
				Summary: "Asynchronous notifications that interrupt a process: how it is told to stop, crashes, and hears about its children.",
				Body: `A signal is a small numbered notification the kernel delivers to a
process, interrupting whatever it was doing. Some come from other
processes (kill), some from the terminal (Ctrl-C sends SIGINT), and some
from the process's own mistakes (SIGSEGV for a bad memory access,
SIGFPE for an integer division by zero, SIGPIPE for writing to a pipe
nobody reads).

For each signal a process can take the default action (usually
terminate), ignore it, or install a handler with sigaction. Two signals
can never be caught or ignored: SIGKILL and SIGSTOP. That is why kill -9
always works, and why it should be the last resort: the program gets no
chance to save or clean up. SIGTERM, the polite request, is what
service managers send first.

A handler runs in the middle of anything, even inside malloc or printf,
so it may only call async-signal-safe functions and should do almost
nothing. The standard pattern is to set a flag and let the main loop
act on it:

  static volatile sig_atomic_t stop;
  static void on_term(int sig) { (void)sig; stop = 1; }
  …
  struct sigaction sa = {0};
  sa.sa_handler = on_term;
  sigaction(SIGTERM, &sa, NULL);
  while (!stop) { /* work */ }
  /* save, close, exit cleanly */`,
				Diagram: `  kill -TERM pid
       │
       ▼                     main loop:  work ─ work ─ work ─ check stop ─▶ save, exit 0
  kernel interrupts ─▶ on_term(): stop = 1 ─┘ (return to where it was)`,
				UnderTheHood: `A shell reports a process killed by signal n with exit status 128 + n.
Measured in bash on Linux:

  yes | head -1           yes exits with 141   SIGPIPE (13): its reader quit
  kill -TERM of a sleep    143                 SIGTERM (15)
  kill -KILL of a sleep    137                 SIGKILL  (9), "Killed"

yes kept writing after head had read its one line and exited; the next
write to the pipe raised SIGPIPE, whose default action terminated it.
That is how pipelines shut down early producers, and why a network
server must ignore SIGPIPE (or use MSG_NOSIGNAL), or a client that
disconnects at the wrong moment kills the whole server.

kill -l on Linux lists 62 signals: the 31 classic ones plus the
"real-time" signals 34 to 64, which queue instead of merging.`,
				MentalModel: "A signal is a tap on the shoulder at any moment: note it in a flag, and act on it when it is safe.",
				TryIt:       "Write a C program that loops printing a counter, catches SIGINT with sigaction, and prints a summary instead of dying when you press Ctrl-C; press Ctrl-\\ to see SIGQUIT's default action.",
				Analogy: `A fire alarm in an office. Whatever you are doing, it interrupts you.
Sensible people do not start packing the whole office on the spot; they
note it and head for the exit in an orderly way. kill -9 is the
building being demolished with you inside: no time to save anything.`,
				Example: `systemd and Docker stop services with SIGTERM and wait (10 seconds by
default in docker stop) before sending SIGKILL: services that handle
SIGTERM can finish requests and flush data. Database servers flush
their logs on SIGTERM, nginx reloads its configuration on SIGHUP, and
every crash report you have seen ("Segmentation fault (core dumped)")
started as a SIGSEGV.`,
				Exercises: trio(
					"A program dies and the shell says its exit status was 139. Which signal killed it, and what does that usually mean?",
					"139 − 128 = 11, SIGSEGV: an invalid memory access, such as a NULL or dangling pointer.",
					"In C, write a program that counts up in a loop until it receives SIGTERM or SIGINT, then prints how far it got and exits with status 0. Test it with kill -TERM and Ctrl-C, and confirm that kill -KILL still ends it at once.",
					"Use the flag pattern: the handler only sets a volatile sig_atomic_t; the loop checks it and does the printing.",
					"Make a server shut down gracefully: on SIGTERM it stops accepting new connections, finishes the requests in progress, flushes its log and exits, all within a deadline. Test it by sending SIGTERM in the middle of a burst of requests.",
					"Accept with a timeout (or poll) so the loop sees the flag even when no connection arrives; keep a count of requests in progress, and exit when it reaches 0 or the deadline passes.",
				),
			},
			{
				Name:    "Scheduling & Context Switches",
				Summary: "How one CPU gives many threads the illusion of running at once.",
				Body: `A timer interrupt periodically returns control to the kernel, whose
scheduler picks the next runnable thread. Linux used CFS for years and
has used EEVDF since 6.6. A context switch saves one thread's registers
and restores another's. Switching between processes may also switch page
tables and lose TLB entries, and a thread moved to another CPU arrives
with cold caches.

A thread is in one of three broad states: running, runnable (waiting for
a CPU) or blocked (waiting for I/O, a lock or a timer). Most threads,
most of the time, are blocked: the scheduler's job is to give the CPUs to
whatever is runnable, fairly, while keeping interactive tasks
responsive. nice values and priorities bias that choice.

Threads in one process share an address space, so switching between them
is cheaper, but they also share every bug. Green threads (goroutines,
async runtimes) do the switching in user space and avoid the kernel
entirely.`,
				Diagram: `CPU 0  │ A ███████ │ B ████ │ A ██████ │ C ███ │ A …
                    ▲         ▲          ▲
          timer interrupt or blocking call: save A's registers,
          pick the next runnable thread, restore its registers`,
				UnderTheHood: `Two processes passing a byte back and forth through two pipes force a
context switch on every hop. On a test machine (a virtual machine):

  both free to run on any CPU      36.4 µs per round trip
  both pinned to CPU 0 (taskset)    2.85 µs per round trip

Pinned, each round trip is two switches on one CPU. Unpinned, each
wake-up crossed CPUs: the sleeping process was woken by another CPU,
which had to interrupt it (an inter-processor interrupt), and in a
virtual machine that is expensive. Where threads run, and which CPU
wakes them, can matter more than the switch itself.`,
				MentalModel: "Concurrency is an illusion made of interrupts and saved registers.",
				TryIt:       "Write the two-pipe ping-pong in C, time 100,000 round trips, then run it again under taskset -c 0 and explain the difference.",
			},
			{
				Name:    "Synchronisation & Memory Ordering",
				Summary: "Shared memory without rules is a data race.",
				Body: `Two threads doing count++ lose updates because it is a load, an add and
a store, and the other thread can run in between. Atomic read-modify-
write instructions (compare-and-swap, fetch-add) are the hardware
primitive underneath. A mutex is an atomic fast path plus a futex so
that waiting threads sleep in the kernel.

In C, a data race (two threads touching the same variable, at least one
writing, without synchronisation) is undefined behaviour. The compiler
may assume it never happens, and keep a shared variable in a register
for a whole loop.

CPUs and compilers reorder memory operations. Acquire and release
semantics (and fences) define which reorderings other threads can
observe: a release store publishes everything written before it to any
thread that does an acquire load of it. A deadlock needs all four
Coffman conditions: mutual exclusion, hold-and-wait, no preemption and
circular wait. Break any one of them, for example by always taking
locks in the same order.`,
				Diagram: `thread A: load count (5)  add 1  store 6
thread B:     load count (5)  add 1  store 6     two increments, one lost
                                                   (count should be 7)`,
				UnderTheHood: `Two threads each incremented a shared long 10 million times. Results
from one test machine:

  plain long, gcc -O0         10.6 to 11.0 million of 20 million
  plain long, gcc -O1         exactly 10,000,000, every time
  atomic_fetch_add            20,000,000 in about 290 ms
  pthread mutex around ++     20,000,000 in about 1.8 s

At -O0 the race lost about half the increments. At -O1 it lost exactly
half, because gcc compiled each thread's loop into one load at the start
and one store at the end, a single "+= 10,000,000": the threads'
results overwrote each other whole. That is legal, because a data race
is undefined: the compiler owes you nothing. Only the atomic and the
mutex were correct, and correctness cost speed: contended atomics are
70 times slower than the plain increment, and a contended mutex six
times slower again.`,
				MentalModel: "Without a happens-before edge, another thread may see anything.",
				TryIt:       "Build a spinlock with atomics, then a futex-based mutex, and benchmark both under contention.",
			},
		},
		Resources: []Resource{
			{"Book", "Operating Systems: Three Easy Pieces (free)", "https://pages.cs.wisc.edu/~remzi/OSTEP/", "The clearest OS textbook; do its homework simulators."},
			{"Course", "MIT 6.1810 Operating System Engineering (xv6), OCW archive", "https://ocw.mit.edu/courses/6-1810-operating-system-engineering-fall-2023/", "Lectures and labs that hack on xv6, a small, real Unix-like kernel."},
			{"Paper", "What Every Programmer Should Know About Memory", "https://www.akkadia.org/drepper/cpumemory.pdf", "Drepper's deep dive into caches, the TLB and NUMA."},
			{"Site", "OSDev Wiki", "https://wiki.osdev.org/", "For writing your own kernel from the boot sector up."},
			{"Book", "The Linux Programming Interface", "https://man7.org/tlpi/", "A definitive reference for syscalls."},
			{"Site", "Linux Kernel Labs", "https://linux-kernel-labs.github.io/", "Guided exercises on real kernel modules."},
		},
		Blueprints: []Blueprint{
			{"Custom malloc", "A general-purpose allocator with segregated free lists, splitting and coalescing, and mmap for large blocks.",
				[]string{"Bump allocator", "Free list with first-fit", "Coalescing via boundary tags", "Size-class bins", "Benchmark against glibc"}},
			{"Unix Shell", "A shell with fork/exec/wait, pipes, redirection and job control.",
				[]string{"Run one command", "Pipes a | b | c", "Redirection < > >>", "Background jobs and signals", "Built-ins: cd, exit, jobs"}},
			{"User-Space Threads", "Green threads with an assembly context switch and a round-robin scheduler.",
				[]string{"Context switch in asm", "Spawn and yield", "Preemption via SIGALRM", "Channels or mutexes"}},
			{"xv6 Kernel Hacking", "Extend xv6 with new system calls and copy-on-write fork.",
				[]string{"Add a trace syscall", "Print the page table", "Lazy allocation", "COW fork"}},
		},
		Quiz: []Question{
			{"What happens on the first write to a page after fork()?", "The page is shared read-only (copy-on-write). The write faults, the kernel copies the page and remaps it writable, and the write is retried."},
			{"Why does the TLB exist?", "A page-table walk needs several dependent memory accesses. The TLB caches recent virtual-to-physical translations so most accesses skip the walk."},
			{"What do threads share that processes do not?", "The address space: heap, globals, code and file descriptors. Each thread still has its own stack and registers."},
			{"Name the four conditions required for deadlock.", "Mutual exclusion, hold-and-wait, no preemption and circular wait."},
			{"Why can a process have a virtual size of 1 GB but use only 2 MB of RAM?", "Pages are only backed by RAM when first touched: malloc reserved the addresses, but the kernel allocates each page lazily, on its first page fault."},
			{"Where does malloc get memory from, and why is a thousand small mallocs only one system call?", "From the kernel, with brk (growing the heap) or mmap (a new mapping, for large blocks). It asks for a large slice at once and carves small blocks out of it itself."},
			{"Why is copying a file one byte at a time so slow?", "Every read and write is a system call, about a hundred times the cost of a function call, so a 1 MB copy makes two million of them. A buffer of a few KiB cuts that to a few hundred."},
			{"What do dup2 and pipe do, and how does a shell use them for a | b?", "pipe creates a channel with a read end and a write end; dup2 makes one descriptor refer to another's open file. The shell dup2s the write end onto a's standard output and the read end onto b's standard input before exec-ing each."},
			{"Why does a reader of a pipe hang forever if a write end is left open?", "End-of-file only arrives when every copy of the write end is closed; while one is open, the reader waits for more data."},
			{"Which two signals can never be caught or ignored, and why is kill -9 a last resort?", "SIGKILL and SIGSTOP. kill -9 gives the program no chance to save data or clean up; send SIGTERM first."},
			{"What may a signal handler safely do?", "Very little: set a volatile sig_atomic_t flag, or call async-signal-safe functions such as write. It can interrupt the program inside malloc or printf, so calling those is unsafe."},
			{"Why can an unsynchronised counter shared by two threads lose a whole thread's work at -O1?", "A data race is undefined behaviour, so the compiler may keep the counter in a register for the whole loop and store it once at the end, overwriting the other thread's total."},
		},
	},

	// -----------------------------------------------------------------
	7: {
		Overview: `A database is a data structure that survives crashes. Learn how bytes
are laid out on disk, how indexes trade write cost for read cost, and how
durability is actually guaranteed. Hint: this very program uses the
fsync-and-rename trick.`,
		Concepts: []Concept{
			{
				Name:    "Durability: Page Cache, fsync & Atomic Rename",
				Summary: "write() returning does not mean your data is on disk.",
				Body: `write() copies data into the kernel page cache and returns. The disk may
see it seconds later, or never if the power fails. fsync() forces the
file's data and metadata to stable storage. A new directory entry also
needs an fsync of the directory.

The classic atomic-update recipe is: write a temp file, fsync it, rename
it over the target, then fsync the directory. rename is atomic, so
readers see the old file or the new one, never a mix. SSDs add their own
layer: flash is erased in large blocks, so the drive's firmware
constantly remaps and rewrites data (write amplification).`,
				MentalModel: "Durable means it survived a power cut, not that the call returned.",
				TryIt:       "Read atomicWriteJSON in this program's main.go and explain why each step exists.",
			},
			{
				Name:    "B-Trees",
				Summary: "Wide, shallow, page-sized trees: the default index for decades.",
				Body: `A B-tree node is a whole disk page (4 to 16 KiB) holding hundreds of
sorted keys, so the fan-out is huge and a billion keys need only about
3 to 4 levels, which means 3 to 4 page reads per lookup. Inserts go into
a leaf and split it when full, pushing a separator key up.

In a B+tree, values live only in the leaves and the leaves are linked, so
range scans are sequential. B-trees update in place and are read-optimised.
Postgres, SQLite and InnoDB use them.`,
				Diagram: `                 [ 40 | 80 ]
          ┌──────────┼──────────┐
   [10|20|30]   [50|60|70]   [90|95]      ◀ leaves hold values,
       ↔            ↔            ↔           linked for range scans`,
				MentalModel: "Make the tree as wide as a disk page so it never gets deep.",
				TryIt:       "Insert keys 1..10,000 into a B+tree with fan-out 4 and print its height after each split.",
			},
			{
				Name:    "LSM-Trees",
				Summary: "Turn random writes into sequential ones, then clean up later.",
				Body: `A log-structured merge tree sends every write to an append-only WAL and
an in-memory sorted memtable. When the memtable is full it is flushed as
an immutable, sorted SSTable file. Background compaction merges SSTables
and drops overwritten or deleted keys (tombstones).

A read checks the memtable, then SSTables from newest to oldest. Bloom
filters let it skip files that definitely do not hold the key.
LSM-trees are write-optimised; RocksDB, LevelDB and Cassandra use them.
Each design trades off read, write and space amplification (the RUM
conjecture).`,
				Diagram: `write ─▶ WAL (append)  +  memtable (sorted, in RAM)
                              │ flush when full
                              ▼
  L0   [sst] [sst] [sst]                 newest, may overlap
  L1   [ sst  sst  sst  sst ]            compacted, disjoint
  L2   [ sst sst sst sst sst sst sst … ] 10× larger per level`,
				MentalModel: "Write now in order, sort out the mess later in the background.",
				TryIt:       "Work out the write amplification of leveled compaction with a size ratio of 10 and 4 levels.",
			},
			{
				Name:    "Write-Ahead Logging & Recovery",
				Summary: "Log the intent before touching the data, so crashes can be replayed.",
				Body: `The WAL rule: before a modified page reaches disk, the log record that
describes the change must already be durable. After a crash, recovery
replays the log to redo committed work and undo uncommitted work (ARIES
is the canonical algorithm).

Checkpoints bound how much log must be replayed. Group commit batches
many transactions into one fsync, which is the main reason databases can
commit thousands of transactions per second on hardware that manages only
hundreds of fsyncs per second.`,
				MentalModel: "The log is the truth; the data files are a cache of it.",
				TryIt:       "Kill -9 your KV store in the middle of a write loop 100 times and verify that recovery never loses an acknowledged write.",
			},
			{
				Name:    "Transactions, Isolation & MVCC",
				Summary: "Making concurrent readers and writers look sequential, or nearly so.",
				Body: `ACID stands for atomicity, consistency, isolation and durability.
Isolation levels trade correctness for concurrency, and each level allows
certain anomalies: dirty reads, non-repeatable reads, phantoms and write
skew.

MVCC keeps several versions of each row, stamped with transaction IDs, so
readers see a consistent snapshot without blocking writers. Snapshot
isolation still permits write skew. True serializability needs extra
checks (SSI) or locking (2PL).`,
				MentalModel: "Every transaction reads a photo of the past and argues about the future at commit.",
				TryIt:       "Reproduce write skew (the on-call doctors example) in Postgres at REPEATABLE READ, then fix it with SERIALIZABLE.",
			},
			{
				Name:    "SQL & the Relational Model",
				Summary: "The language nearly every application uses to ask questions of its data.",
				Body: `The relational model stores data in tables of rows and columns. Each row
is identified by a primary key, and rows link to other tables through
foreign keys. SQL is declarative: you describe the result you want, and
the database's query planner decides how to produce it.

The core statements: SELECT … FROM … WHERE filters rows; JOIN combines
tables; GROUP BY aggregates (COUNT, SUM, AVG); INSERT, UPDATE and DELETE
change data. Normalisation removes duplication so that each fact is
stored once. Indexes (the B-trees from this stage) make lookups fast,
and EXPLAIN shows the plan the database chose.`,
				Diagram: `customers                  orders
┌────┬──────┐              ┌────┬─────────────┬───────┐
│ id │ name │   1 ── many  │ id │ customer_id │ total │
├────┼──────┤              ├────┼─────────────┼───────┤
│ 1  │ Ana  │ ◀─────────── │ 10 │ 1           │ 25.00 │
│ 2  │ Ben  │              │ 11 │ 1           │ 12.50 │
└────┴──────┘              └────┴─────────────┴───────┘
SELECT c.name, SUM(o.total) FROM customers c
JOIN orders o ON o.customer_id = c.id GROUP BY c.name;`,
				MentalModel: "Store each fact once, and answer questions with joins.",
				TryIt:       "Create the two tables above in SQLite and run the query.",
				Analogy: `A well-organised spreadsheet workbook where each sheet holds one kind of
thing (customers, orders) and a customer number links them, instead of
retyping the customer's address on every single order.`,
				Example: `Almost every app you use (banking, shopping, social media, and your
phone's own messages and contacts, stored in SQLite) keeps its data in
relational databases and queries it with SQL. SQL dates from the 1970s
and is still among the most widely used languages in developer surveys.`,
				Exercises: trio(
					`Given students(id, name) and grades(student_id, course, grade), describe in plain English what "SELECT name FROM students JOIN grades ON grades.student_id = students.id WHERE grade < 50" returns.`,
					"Students with at least one failing grade, possibly listed more than once.",
					"Design tables for a library (books, members, loans), create them in SQLite with sample data, and write queries for overdue books, the most borrowed book, and members who have never borrowed anything.",
					`"Never borrowed" needs a LEFT JOIN … WHERE loans.id IS NULL.`,
					"Import a public CSV dataset (city bike trips, films, weather) into SQLite or PostgreSQL, answer five questions you care about with SQL, then add an index for your slowest query and compare EXPLAIN output and timing before and after.",
					".import in the sqlite3 shell loads a CSV file directly.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "CMU 15-445 Database Systems", "https://15445.courses.cs.cmu.edu/", "Andy Pavlo's lectures and the BusTub project."},
			{"Book", "Designing Data-Intensive Applications", "https://dataintensive.net/", "The map of the whole storage and distributed-data landscape (2nd edition published 2026)."},
			{"Article", "Let's Build a Simple Database", "https://cstack.github.io/db_tutorial/", "A SQLite clone in C, step by step."},
			{"Book", "Build Your Own Database", "https://build-your-own.org/database/", "B+tree, KV store and SQL layer from scratch, in Go."},
			{"Site", "SQLite architecture", "https://www.sqlite.org/arch.html", "A small, real, well-documented engine."},
			{"Article", "LevelDB implementation notes", "https://github.com/google/leveldb/blob/main/doc/impl.md", "SSTable format and compaction in a few pages."},
			{"Site", "Use The Index, Luke", "https://use-the-index-luke.com/", "How indexes behave from the SQL side."},
		},
		Blueprints: []Blueprint{
			{"Bitcask-Style KV Store", "An append-only log with an in-memory hash index, crash recovery and merge compaction.",
				[]string{"Append-only data file with CRC", "In-memory keydir", "Recovery by scanning the log", "Compaction", "Crash tests with kill -9"}},
			{"LSM-Tree Engine", "Skiplist memtable, WAL, SSTables with an index block, bloom filters and leveled compaction.",
				[]string{"Memtable and WAL", "SSTable writer and reader", "Bloom filters", "Merge iterator", "Leveled compaction"}},
			{"B+Tree on Pages", "A disk-backed B+tree with fixed 4 KiB pages and an LRU buffer pool.",
				[]string{"Page layout (slotted pages)", "Buffer pool with LRU", "Insert with splits", "Range-scan cursor", "Delete with merges"}},
		},
		Quiz: []Question{
			{"Why do LSM-trees use bloom filters?", "Without them a lookup for a missing key could read every SSTable. A bloom filter answers 'definitely not here' from memory."},
			{"Why is write() plus close() not enough for durability?", "The data can sit in the page cache. You must fsync the file, and the directory too after creating or renaming a file."},
			{"What is write amplification?", "Bytes physically written to storage divided by bytes the application wrote. Compaction, page rewrites and SSD garbage collection all increase it."},
			{"How does group commit improve throughput?", "Many transactions share one fsync of the log, so the fixed cost of a flush is spread across all of them."},
		},
	},

	// -----------------------------------------------------------------
	8: {
		Overview: `Programs on different machines can only talk through an unreliable
network with no shared clock. Learn the protocol stack, how to write
servers that handle thousands of connections, and how replicated systems
agree on anything at all.`,
		Concepts: []Concept{
			{
				Name:    "The Layered Stack & Encapsulation",
				Summary: "Each layer wraps the layer above in its own header.",
				Body: `The link layer (Ethernet, Wi-Fi) moves frames to the next hop. The
network layer (IP) routes packets across networks on a best-effort basis:
they may be lost, duplicated or reordered. The transport layer turns that
into something useful. TCP gives a reliable, ordered byte stream; UDP
gives bare datagrams. Applications (HTTP, DNS, your own protocol) sit on
top.

On the wire, a packet is headers nested like envelopes, each read by
one layer.`,
				Diagram: `┌──────────┬────────┬─────────┬────────────────────┬─────┐
│ Ethernet │  IP    │  TCP    │  HTTP payload      │ FCS │
│  header  │ header │ header  │                    │     │
└──────────┴────────┴─────────┴────────────────────┴─────┘`,
				MentalModel: "Reliability is built in software, on top of hardware that makes no promises.",
				TryIt:       "Capture a curl request with tcpdump -X and label every header byte by hand.",
			},
			{
				Name:    "TCP: Handshake, Flow & Congestion Control",
				Summary: "A reliable, ordered stream built from lost and reordered packets.",
				Body: `Connection set-up exchanges initial sequence numbers: SYN, SYN-ACK, ACK.
Every byte is numbered. The receiver acknowledges, and missing data is
retransmitted after a timeout or three duplicate ACKs.

Flow control (the receive window) protects the receiver. Congestion
control protects the network: slow start doubles the window every round
trip, then additive-increase/multiplicative-decrease (Reno, CUBIC) or
model-based pacing (BBR) takes over. Because everything is one ordered
stream, a single lost packet stalls all the data behind it (head-of-line
blocking), which is one reason QUIC exists.`,
				Diagram: `client                       server
  │ ── SYN seq=x ────────────▶ │
  │ ◀──── SYN-ACK seq=y ack=x+1│
  │ ── ACK ack=y+1 ──────────▶ │   connection ESTABLISHED`,
				MentalModel: "TCP is a sliding window that grows until the network pushes back.",
				TryIt:       "Use tc netem to add 2% loss and 100 ms delay, and measure the throughput of a bulk transfer.",
			},
			{
				Name:    "Sockets & I/O Multiplexing",
				Summary: "Serving 10,000 connections without 10,000 threads.",
				Body: `Server lifecycle: socket, bind, listen, then accept in a loop. With one
thread per connection, memory and context switches become the limit.
Non-blocking sockets plus a readiness API (epoll on Linux, kqueue on BSD)
let one thread wait on many sockets and service whichever are ready.
This is the event-loop design.

io_uring goes further: you submit operations into a shared ring buffer
and later collect completions, which avoids a syscall per operation.
Go's runtime hides an epoll-based netpoller behind blocking-looking
goroutines.`,
				MentalModel: "Do not wait on one socket; ask the kernel which of all of them are ready.",
				TryIt:       "Build an echo server with epoll and load it with 10k connections.",
			},
			{
				Name:    "Time, Clocks & Ordering",
				Summary: "In a distributed system, 'before' is something you build, not something you read.",
				Body: `Machines' physical clocks drift and are corrected (NTP) in jumps, so
timestamps from different machines cannot reliably order events. A
Lamport clock is a counter that is incremented on every event and
maxed-and-incremented on every received message. It guarantees that if
A happened before B, then clock(A) < clock(B). The converse does not
hold.

Vector clocks keep one counter per node, and they can tell whether two
events are causally ordered or truly concurrent.`,
				MentalModel: "Causality is carried by messages, not by wall clocks.",
				TryIt:       "Simulate three nodes exchanging messages and show vector clocks detecting a concurrent write.",
			},
			{
				Name:    "Consensus & Replication (Raft)",
				Summary: "Getting a majority to agree on one ordered log, despite failures.",
				Body: `A replicated state machine applies the same log of commands on every
node, so every node reaches the same state. Raft splits consensus into
three parts:
- leader election, with randomised timeouts and monotonically increasing terms;
- log replication, where the leader appends and followers must match the previous entry;
- safety, where an entry is committed once a majority stores it, and only up-to-date nodes can win elections.

A cluster of 2f+1 nodes survives f failures. During a partition, the
minority side cannot make progress. That is the CAP trade-off in
practice.`,
				MentalModel: "A value is decided the moment a majority has written it down.",
				TryIt:       "Explore raft.github.io's visualisation: kill the leader and watch the election.",
			},
			{
				Name:    "HTTP, DNS & How the Web Works",
				Summary: "What really happens when you type a web address and press Enter.",
				Body: `First, DNS translates a name such as example.com into an IP address by
asking a chain of servers (your resolver, a root server, the .com
servers, then the domain's own servers) and caching the answer. Your
browser opens a TCP (or QUIC) connection to that address, sets up TLS
for encryption (Stage 10), and sends an HTTP request: a method (GET,
POST), a path, headers and sometimes a body.

The server replies with a status code (200 OK, 404 Not Found, 500 Server
Error), headers and content. The HTML refers to CSS, JavaScript and
images, each of which triggers more requests. CDNs cache content close
to users, cookies keep you logged in, and REST/JSON APIs use the same
HTTP for mobile apps.`,
				Diagram: `browser ── "example.com?" ───────▶ DNS resolver ──▶ IP address
browser ── TCP + TLS handshake ─────────────────▶ server
browser ── GET /index.html ─────────────────────▶ server
browser ◀── 200 OK + HTML ─────────────────────── server
browser ── GET /style.css, /app.js, images … ───▶ often a CDN`,
				MentalModel: "Name → address → connection → request → response, repeated for every resource on the page.",
				TryIt:       "Open your browser's developer tools (F12), go to the Network tab, reload a page and count the requests.",
				Analogy: `Looking up a friend's number in your contacts (DNS), calling them (the
connection), asking for something specific (the request) and getting an
answer (the response). A CDN is a chain store with a branch in every
city, so you do not have to travel to the head office.`,
				Example: `In October 2021, Facebook, Instagram and WhatsApp vanished for about six
hours after a configuration change withdrew the network routes to
Facebook's DNS servers, so the rest of the internet could no longer find
them. CDNs such as Cloudflare and Akamai serve a large share of all web
traffic.`,
				Exercises: trio(
					"Match each status code to its meaning: 200, 301, 404, 500 and 503.",
					"2xx success, 3xx redirect, 4xx your mistake, 5xx the server's.",
					"Use curl -v to fetch a web page and identify the connection, the TLS handshake, the request headers and the response headers. Then use dig or nslookup to look up the domain's DNS records.",
					"curl -v https://example.com prints every step.",
					"Write a tiny HTTP server from raw sockets (no web framework) that serves files from a folder with correct status codes and Content-Type headers, and open it in your browser.",
					`Read the request line ("GET /path HTTP/1.1"), then send "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n" followed by the file.`,
				),
			},
		},
		Resources: []Resource{
			{"Book", "Beej's Guide to Network Programming", "https://beej.us/guide/bgnet/", "The friendliest introduction to the sockets API."},
			{"Course", "MIT 6.5840 Distributed Systems", "https://pdos.csail.mit.edu/6.824/", "Labs: MapReduce, Raft, sharded KV."},
			{"Site", "The Raft consensus site", "https://raft.github.io/", "The paper, an interactive visualisation and implementations."},
			{"Course", "Gossip Glomers (Fly.io + Maelstrom)", "https://fly.io/dist-sys/", "Distributed-systems challenges tested with a network simulator."},
			{"Book", "High Performance Browser Networking (free)", "https://hpbn.co/", "TCP, TLS, HTTP/2 and QUIC performance explained."},
			{"Site", "Jepsen analyses", "https://jepsen.io/", "How real databases break under network partitions."},
		},
		Blueprints: []Blueprint{
			{"Epoll HTTP/1.1 Server", "A non-blocking, single-threaded HTTP server built on epoll, benchmarked with wrk.",
				[]string{"Blocking echo server", "Non-blocking + epoll loop", "HTTP request parsing", "Keep-alive", "Benchmark and profile"}},
			{"Raft Implementation", "Leader election, log replication and persistence, tested under simulated partitions.",
				[]string{"Election with terms", "AppendEntries replication", "Commit index and apply", "Persistence and restart", "Partition test harness"}},
			{"Replicated KV (capstone A)", "Put the Stage 7 storage engine behind the Raft log: a fault-tolerant KV service.",
				[]string{"Client protocol", "Raft-backed state machine", "Snapshots", "Linearizable reads", "Chaos testing"}},
		},
		Quiz: []Question{
			{"Why does TCP need a three-way handshake rather than two messages?", "Each side must choose an initial sequence number and have the other side confirm it. The third message also protects against stale, duplicated SYNs."},
			{"How many Raft nodes do you need to tolerate 2 failures?", "Five (2f+1), so that a majority of 3 remains."},
			{"Why does epoll scale better than one thread per connection?", "One thread waits on thousands of sockets. There is no per-connection stack memory and no context switch for each idle connection."},
			{"What does a Lamport clock guarantee, and what does it not?", "If A happened before B, then L(A) < L(B). But L(A) < L(B) does not imply that A happened before B; they may be concurrent."},
		},
	},

	// -----------------------------------------------------------------
	13: {
		Overview: `Deep learning is linear algebra plus the chain rule, run at scale. This
stage builds fluent intuition for matrices as transformations and for
derivatives of functions of vectors and matrices, so that every
backward pass in Stage 15 is something you can derive on paper.`,
		Concepts: []Concept{
			{
				Name:    "Matrices as Linear Maps",
				Summary: "A matrix is a function that transforms space.",
				Body: `Read a matrix by its columns: column j is where the j-th basis vector
lands. Then Ax is x's coordinates used to mix those columns.
Multiplying matrices composes the transformations, which is why the
shapes chain: (m×n)(n×p) = (m×p).

The determinant is the factor by which volumes scale (0 means some
dimension is squashed). The rank is the dimension of the output space.
A neural network layer y = Wx + b is exactly such a map, plus a shift.`,
				MentalModel: "Columns are where the axes go; multiplication is composition.",
				TryIt:       "Draw the 2×2 matrices for a rotation, a shear and a projection, and compute each one's determinant.",
			},
			{
				Name:    "Eigenvectors & the SVD",
				Summary: "Every matrix is a rotation, then a stretch, then a rotation.",
				Body: `An eigenvector keeps its direction under A: Av = λv. The singular value
decomposition A = UΣVᵀ applies to every matrix. Vᵀ rotates the input,
Σ scales along the axes, and U rotates the result.

Keeping only the largest k singular values gives the best rank-k
approximation. That idea underlies PCA and compression, and it is also
the idea behind LoRA fine-tuning, which learns a low-rank update ΔW = BA
instead of a full matrix.`,
				MentalModel: "The singular values tell you how much each direction matters.",
				TryIt:       "Compress a greyscale image with rank-k SVD for k = 5, 20 and 50, and compare quality with storage.",
			},
			{
				Name:    "Gradients, Jacobians & the Matrix Chain Rule",
				Summary: "Backprop is the chain rule with shapes that must line up.",
				Body: `For a scalar loss L and a vector x, the gradient ∂L/∂x has the same
shape as x. For a vector function y = f(x), the Jacobian J holds every
∂yᵢ/∂xⱼ. The chain rule composes Jacobians. In practice you never build
them: you propagate the upstream gradient ∂L/∂y through each operation.

The identities to know by heart, for y = Wx:
- ∂L/∂x = Wᵀ · (∂L/∂y)
- ∂L/∂W = (∂L/∂y) · xᵀ
- ∂L/∂b = ∂L/∂y (summed over the batch)

Shape-checking is your best debugger. If the dimensions do not match,
the formula is wrong.`,
				Diagram: `  x (n)  ──[ W (m×n) ]──▶  y (m)  ──▶ … ──▶  L (scalar)
  ∂L/∂x = Wᵀ ∂L/∂y  ◀──────── ∂L/∂y (m) ◀──── upstream gradient`,
				MentalModel: "Gradients flow backwards through the transpose.",
				TryIt:       "Derive ∂L/∂W for L = ‖Wx − t‖² by hand, then verify it with finite differences.",
			},
			{
				Name:    "Probability, Softmax & Cross-Entropy",
				Summary: "Turning scores into probabilities and measuring surprise.",
				Body: `Softmax maps logits z to a distribution: pᵢ = exp(zᵢ) / Σ exp(zⱼ).
Cross-entropy loss −log p_target is the model's surprise at the true
label. Minimising it is maximum-likelihood estimation.

Combined, the gradient is simply p − y (the predicted distribution minus
the one-hot target). For numerical stability, subtract max(z) before
exponentiating. The result is unchanged, because softmax is invariant to
shifts, and overflow disappears. The log-sum-exp trick does the same job
for log-probabilities.`,
				MentalModel: "The gradient of softmax plus cross-entropy is prediction minus truth.",
				TryIt:       "Implement softmax naively and compute softmax([1000, 1001]); then fix it with the max-subtraction trick.",
			},
			{
				Name:    "Optimisation: From Gradient Descent to Adam",
				Summary: "Walking downhill in a million dimensions.",
				Body: `Gradient descent updates θ ← θ − η∇L. Too large a learning rate η
diverges; too small a rate crawls. Stochastic minibatches give noisy but
cheap gradient estimates. Momentum averages past gradients to push
through ravines.

Adam keeps running estimates of each parameter's gradient mean and
variance and scales every step individually. AdamW decouples weight
decay from that scaling. Neural-network loss landscapes are non-convex,
yet in practice SGD-family methods find good minima.`,
				MentalModel: "The learning rate is the single most important hyperparameter; tune it first.",
				TryIt:       "Minimise the Rosenbrock function with SGD, momentum and Adam, and plot the paths.",
			},
		},
		Resources: []Resource{
			{"Video", "3Blue1Brown: Essence of Linear Algebra", "https://www.3blue1brown.com/lessons/eola-preview/", "Geometric intuition first; watch it before any textbook."},
			{"Course", "MIT 18.06 Linear Algebra (Gilbert Strang)", "https://ocw.mit.edu/courses/18-06-linear-algebra-spring-2010/", "The classic lecture series with problem sets."},
			{"Paper", "The Matrix Calculus You Need for Deep Learning", "https://explained.ai/matrix-calculus/", "Parr & Howard; exactly the derivatives backprop needs."},
			{"Book", "Mathematics for Machine Learning (free)", "https://mml-book.github.io/", "Linear algebra, calculus and probability, aimed at ML."},
			{"Book", "Linear Algebra Done Right (free)", "https://linear.axler.net/", "A rigorous, proof-based treatment."},
			{"Site", "The Matrix Cookbook", "https://www.math.uwaterloo.ca/~hwolkowi/matrixcookbook.pdf", "A lookup table of matrix identities and derivatives."},
		},
		Blueprints: []Blueprint{
			{"Matrix Library from Scratch", "Dense matrices with matmul (naive, then cache-blocked), transpose, LU and solve, benchmarked against BLAS.",
				[]string{"Matrix type and matmul", "Cache-blocked matmul", "LU decomposition", "Linear solve", "Benchmark vs numpy"}},
			{"Gradient Checker", "Compare analytic gradients with central finite differences, reporting relative error.",
				[]string{"Finite-difference engine", "Relative-error metric", "Test matmul, softmax and CE", "Reuse in Stage 15"}},
			{"PCA via Power Iteration", "Find the top principal components of MNIST using power iteration and deflation.",
				[]string{"Load and centre the data", "Power iteration", "Deflation for k components", "Visualise the eigen-digits"}},
		},
		Quiz: []Question{
			{"For y = Wx with W of shape m×n, what are the shape and formula of ∂L/∂W?", "Shape m×n: (∂L/∂y)·xᵀ, an outer product of an m-vector and an n-vector."},
			{"Why subtract max(z) before a softmax?", "To avoid exp overflow. Softmax is invariant to adding a constant to every logit, so the result is unchanged."},
			{"What does the number of non-zero singular values tell you?", "The rank of the matrix."},
			{"What is the gradient of cross-entropy with respect to the logits after softmax?", "p − y: the predicted probabilities minus the one-hot target."},
		},
	},

	// -----------------------------------------------------------------
	15: {
		Overview: `Build the machinery of modern AI yourself: an automatic differentiation
engine, then the layers that sit on it, up to a transformer. Once you
have written backward() by hand, frameworks stop being magic.`,
		Concepts: []Concept{
			{
				Name:    "Computational Graphs & Reverse-Mode Autodiff",
				Summary: "Record the forward pass, then replay it backwards with the chain rule.",
				Body: `Each operation creates a node that remembers its inputs and a local
backward rule. The forward pass builds a directed acyclic graph (DAG).
backward() visits the nodes in reverse topological order, and each node
adds its contribution to its inputs' gradients (+=, because a value used
twice receives gradient from both paths).

Reverse mode costs roughly one extra forward pass, however many
parameters there are. That is why it wins for a single scalar loss with
millions of weights. Forward mode would need one pass per input.`,
				Diagram: `x ──┐
    (×)─▶ a ──┐
w ──┘         (+)─▶ y ─▶ L = (y − t)²
b ────────────┘
backward:  ∂L/∂y = 2(y − t)
           ∂L/∂a = ∂L/∂y     ∂L/∂b = ∂L/∂y
           ∂L/∂w = ∂L/∂a·x   ∂L/∂x = ∂L/∂a·w`,
				MentalModel: "Every op knows its own local derivative; the graph multiplies them together.",
				TryIt:       "Build a scalar Value type with + × tanh and backward(), and check it against the Stage 13 gradient checker.",
			},
			{
				Name:    "MLPs, Activations & Initialisation",
				Summary: "Why stacking linear layers needs nonlinearity and careful random numbers.",
				Body: `A stack of linear layers collapses into a single linear map. Nonlinear
activations (ReLU, GELU, tanh) are what give depth its power.

The scale of the initial weights matters. If activations shrink or grow
at every layer, gradients vanish or explode. Xavier/Glorot
initialisation (for tanh) and Kaiming/He initialisation (for ReLU) choose
the weight variance so that activation variance stays roughly constant
through depth.`,
				MentalModel: "Without nonlinearity, depth is an illusion; with bad initialisation, it cannot be trained.",
				TryIt:       "Train a 10-layer MLP with N(0,1) initialisation and then with Kaiming, and plot each layer's activation standard deviation.",
			},
			{
				Name:    "Residual Connections & Normalisation",
				Summary: "The two tricks that make very deep networks trainable.",
				Body: `A residual block computes x + f(x). The identity path gives gradients a
highway straight back to early layers, so a block only has to learn a
correction.

Normalisation keeps activations well-scaled. BatchNorm normalises over
the batch; LayerNorm (and RMSNorm) normalise over each token's features,
which suits sequences and variable batch sizes. Transformers usually put
the norm before each sub-layer (pre-norm) for stable training.`,
				MentalModel: "Residuals make identity the default; normalisation keeps the signal in range.",
				TryIt:       "Train a 20-layer MLP with and without residual connections and compare the loss curves.",
			},
			{
				Name:    "Attention & the Transformer",
				Summary: "Every token looks at every other token and takes a weighted mix.",
				Body: `Project the inputs into queries, keys and values: Q = XWq, K = XWk,
V = XWv. Then Attention(Q,K,V) = softmax(QKᵀ / √d) V. The scores say how
much each token should attend to each other token. Dividing by √d keeps
their variance near 1, so the softmax does not saturate.

A causal mask stops a token from seeing the future, which is what makes
next-token prediction possible. Multi-head attention runs several of
these in parallel subspaces. A transformer block is attention plus an
MLP, each wrapped in a residual connection and a norm. The cost is
O(n²) in sequence length, the bottleneck that FlashAttention (Stage 16)
attacks.`,
				Diagram: `tokens ─▶ embed ─▶ ┌──────────────────────────────┐ × N
                   │ x + Attn(Norm(x))            │
                   │ x + MLP(Norm(x))             │
                   └──────────────────────────────┘ ─▶ Norm ─▶ logits`,
				MentalModel: "Attention is a learned, differentiable lookup table.",
				TryIt:       "Compute single-head causal attention for 4 tokens by hand with d = 2.",
			},
			{
				Name:    "The Training Loop & Generalisation",
				Summary: "Loss curves tell you what is wrong, if you know how to read them.",
				Body: `One step: sample a minibatch, run forward, compute the loss, run
backward, take an optimiser step, zero the gradients. Keep a validation
split. Falling training loss with rising validation loss means
overfitting, which you counter with more data, weight decay, dropout or
a smaller model.

Use learning-rate warm-up and then cosine decay. Before scaling
anything, overfit one tiny batch to zero loss. If you cannot, there is a
bug.`,
				MentalModel: "First make it overfit, then make it generalise.",
				TryIt:       "Deliberately overfit 32 MNIST images to 100% accuracy, then add weight decay and watch validation accuracy.",
			},
		},
		Resources: []Resource{
			{"Video", "Karpathy: Neural Networks, Zero to Hero", "https://karpathy.ai/zero-to-hero.html", "From micrograd to GPT, coded live. The spine of this stage."},
			{"Tool", "micrograd", "https://github.com/karpathy/micrograd", "A complete scalar autograd engine in about 150 lines."},
			{"Tool", "nanoGPT", "https://github.com/karpathy/nanoGPT", "A minimal, readable GPT training code base."},
			{"Book", "Deep Learning (Goodfellow et al., free)", "https://www.deeplearningbook.org/", "The theory reference."},
			{"Course", "Stanford CS231n notes", "https://cs231n.github.io/", "The backprop and optimisation notes are superb."},
			{"Article", "The Illustrated Transformer", "https://jalammar.github.io/illustrated-transformer/", "A visual walk through attention."},
			{"Paper", "Attention Is All You Need", "https://arxiv.org/abs/1706.03762", "The original transformer paper."},
			{"Paper", "Automatic Differentiation in ML: a Survey", "https://arxiv.org/abs/1502.05767", "Forward mode, reverse mode and implementation strategies."},
			{"Tool", "tinygrad", "https://github.com/tinygrad/tinygrad", "A small framework whose whole stack, down to the kernels, is readable."},
		},
		Blueprints: []Blueprint{
			{"Scalar Autograd Engine", "A micrograd-style engine: a Value node, the ops, a topological-sort backward pass and a tiny MLP trained on a toy dataset.",
				[]string{"Value with + × and backward", "tanh, exp, pow", "Topological sort", "Neuron/Layer/MLP", "Gradient-check every op"}},
			{"Tensor Autograd + MNIST", "Tensor autograd with broadcasting and matmul backward, and a fused softmax-CE; train an MLP on MNIST to over 97%.",
				[]string{"Tensor type and broadcasting", "matmul backward", "Fused softmax + cross-entropy", "SGD and Adam", "97%+ on MNIST"}},
			{"GPT from Scratch", "A character-level transformer with causal self-attention, trained on tiny Shakespeare using your own engine.",
				[]string{"Tokenizer and data loader", "Causal self-attention", "Transformer block", "Training loop with LR schedule", "Sampling"}},
		},
		Quiz: []Question{
			{"Why are gradients accumulated with += rather than =?", "A value used in several places receives gradient along every path; the multivariate chain rule sums them."},
			{"Why divide by √d in scaled dot-product attention?", "Dot products of d-dimensional vectors have variance about d. Scaling keeps it near 1 so the softmax does not saturate and gradients do not vanish."},
			{"Why is reverse mode preferred over forward mode for training?", "There is one scalar output and millions of inputs. Reverse mode gets every gradient in one backward pass; forward mode needs one pass per input."},
			{"What does it mean if you cannot overfit a single small batch?", "There is almost certainly a bug: in the gradients, the data pipeline, the loss or the learning rate."},
		},
	},

	// -----------------------------------------------------------------
	16: {
		Overview: `This stage is where the other six meet. The memory hierarchy (Stage 5),
virtual memory (Stage 6) and the math and autograd (Stages 13 and 15)
collide on a GPU. Learn why AI performance is mostly about moving bytes,
and how kernels, attention and inference servers are designed around
that.`,
		Concepts: []Concept{
			{
				Name:    "The GPU Execution Model",
				Summary: "Tens of thousands of threads, executed 32 at a time in lockstep.",
				Body: `A CUDA kernel launches a grid of thread blocks. Each block runs on one
streaming multiprocessor (SM), and its threads can share fast on-chip
memory and synchronise. The hardware executes threads in warps of 32
that share one instruction stream (SIMT). If threads in a warp take
different branches, the paths run one after the other (divergence).

GPUs hide memory latency not with big caches but by switching to another
ready warp. High occupancy means enough warps in flight to keep the
arithmetic units fed.`,
				MentalModel: "A GPU is a throughput machine; latency is hidden by always having other work ready.",
				TryIt:       "Write a vector-add kernel, try block sizes of 32, 128, 256 and 1024, and measure the bandwidth achieved.",
			},
			{
				Name:    "GPU Memory Hierarchy & Coalescing",
				Summary: "HBM is fast but far away; shared memory is small but close.",
				Body: `From fastest to slowest: registers (per thread), shared memory/L1
(on-chip, per SM, hundreds of KB), L2, then HBM (TB/s of bandwidth but
hundreds of cycles of latency).

When the 32 threads of a warp read 32 consecutive words, the hardware
coalesces them into a few wide transactions. Strided or scattered
access wastes most of each transaction. Tiling a matmul through shared
memory loads each input tile once and reuses it many times. That tiling
turns a memory-bound kernel into a compute-bound one.`,
				Diagram: `  regs ◀─▶ shared mem / L1 (per SM) ◀─▶ L2 ◀─▶ HBM (device DRAM)
  ~1 cy        ~30 cy                  ~200 cy     ~500 cy · TB/s
  coalesced:  t0 t1 t2 … t31 → [ one 128-byte transaction ]
  strided:    t0 . . t1 . . t2 …  → up to 32 separate transactions`,
				MentalModel: "Neighbouring threads should touch neighbouring bytes.",
				TryIt:       "Write a matrix transpose kernel naively, then with a shared-memory tile, and compare GB/s.",
			},
			{
				Name:    "The Roofline Model: Compute- vs Memory-Bound",
				Summary: "Arithmetic intensity tells you which wall you will hit.",
				Body: `Arithmetic intensity (AI) = FLOPs performed / bytes moved from memory.
The achievable performance is min(peak FLOP/s, memory bandwidth × AI).
The ridge point, peak ÷ bandwidth, is often 100 or more FLOPs per byte on
modern GPUs.

Elementwise ops (add, GELU, the arithmetic inside layernorm) have an AI
of about 1 or less, so they are memory-bound. Large matmuls are
compute-bound. LLM decoding at batch size 1 reads every weight once per
token, so its speed is roughly bandwidth ÷ model size. A 7B model in fp16
(about 14 GB) on a GPU with 1 TB/s tops out near 70 tokens per second,
however many FLOPs the GPU has.`,
				Diagram: `perf │           ________________  peak compute
     │          /
     │         /  ◀ memory-bound: perf = BW × AI
     │        /
     │       /
     └──────┴──────────────────────▶ arithmetic intensity (FLOP/byte)
          ridge`,
				MentalModel: "Count the bytes before you count the FLOPs.",
				TryIt:       "Predict the tokens per second of a 13B int8 model on your GPU from its bandwidth, then measure it with llama.cpp.",
			},
			{
				Name:    "Kernel Fusion & FlashAttention",
				Summary: "The fastest byte is the one that never goes to HBM.",
				Body: `Run separately, every op writes its output to HBM and the next op reads
it back. Fusion merges the ops into one kernel so intermediates stay in
registers or shared memory.

Naive attention materialises the n×n score matrix in HBM. FlashAttention
tiles Q, K and V into blocks that fit in on-chip memory and uses an
online softmax (a running max and a running sum, rescaled as each new
block arrives). The full matrix is never stored. The result is exact,
not an approximation; it is just far less memory traffic. The idea is
IO-awareness.`,
				MentalModel: "Reorganise the algorithm around the memory hierarchy, not the other way round.",
				TryIt:       "Implement the online softmax in plain code and verify that it matches the two-pass version to within 1e-6.",
			},
			{
				Name:    "KV Cache, Batching & PagedAttention",
				Summary: "Inference serving is a memory-management problem.",
				Body: `During autoregressive decoding, the keys and values of past tokens are
cached so they are not recomputed. KV bytes per token = 2 × layers ×
kv_heads × head_dim × bytes_per_value. The cache grows with every
generated token, and for long contexts it can exceed the model weights.

Batching many requests reuses each weight read across all of them,
which raises arithmetic intensity. Continuous batching adds and removes
requests at every step. PagedAttention (vLLM) stores the KV cache in
fixed-size blocks mapped through a block table: virtual memory from
Stage 6, applied to attention. This removes fragmentation and lets
requests share prompt prefixes. Quantisation (int8, int4) cuts bytes
everywhere.`,
				MentalModel: "An LLM server is an operating system whose scarce resource is KV memory.",
				TryIt:       "Compute the KV-cache size for a Llama-style 8B model (32 layers, 8 KV heads, head_dim 128, fp16) at 8k context.",
			},
		},
		Resources: []Resource{
			{"Book", "Programming Massively Parallel Processors", "", "The CUDA textbook (Hwu, Kirk & El Hajj)."},
			{"Site", "CUDA Programming Guide", "https://docs.nvidia.com/cuda/cuda-programming-guide/", "The official execution and memory model."},
			{"Article", "How to Optimize a CUDA Matmul Kernel", "https://siboehm.com/articles/22/CUDA-MMM", "Simon Boehm's step-by-step SGEMM ladder towards cuBLAS speed."},
			{"Article", "Making Deep Learning Go Brrrr", "https://horace.io/brrr_intro.html", "Horace He on compute, memory and overhead as bottlenecks."},
			{"Video", "GPU MODE lectures", "https://github.com/gpu-mode/lectures", "A community lecture series on CUDA, Triton and kernels."},
			{"Paper", "FlashAttention", "https://arxiv.org/abs/2205.14135", "IO-aware exact attention (Dao et al.)."},
			{"Paper", "PagedAttention / vLLM", "https://arxiv.org/abs/2309.06180", "Memory management for LLM serving (Kwon et al.)."},
			{"Tool", "llm.c", "https://github.com/karpathy/llm.c", "GPT-2 training in plain C and CUDA."},
			{"Tool", "llama.cpp", "https://github.com/ggml-org/llama.cpp", "CPU/GPU inference with quantisation; great for reading code."},
		},
		Blueprints: []Blueprint{
			{"CUDA SGEMM Ladder", "Matmul from naive, to coalesced, to shared-memory tiling, to register blocking; plot GFLOP/s against cuBLAS.",
				[]string{"Naive kernel", "Coalesced access", "Shared-memory tiling", "2D register blocking", "Within 80% of cuBLAS"}},
			{"Fused Softmax / LayerNorm Kernel", "A single-pass fused kernel using warp shuffles; report bandwidth achieved as a percentage of peak.",
				[]string{"Naive multi-kernel version", "Warp-level reductions", "Online softmax", "Bandwidth vs roofline"}},
			{"Tiny LLM Inference Engine", "Load a small model's weights and run decoding with a KV cache and sampling; compare tokens per second with the roofline prediction.",
				[]string{"Weight loader", "Forward pass on CPU", "KV cache", "GPU kernels", "int8 quantisation"}},
			{"Paged KV-Cache Allocator", "A block-table allocator for KV memory with prefix sharing and copy-on-write, as in vLLM.",
				[]string{"Fixed-size block pool", "Per-sequence block tables", "Prefix sharing with refcounts", "Copy-on-write on divergence"}},
		},
		Quiz: []Question{
			{"Estimate the decode speed limit at batch size 1 for a 13B int8 model on an 800 GB/s GPU.", "About 13 GB of weights are read per token: 800 / 13 ≈ 60 tokens per second at most."},
			{"Why does batching raise decode throughput?", "Each weight read from HBM serves every sequence in the batch, which multiplies arithmetic intensity."},
			{"What does FlashAttention avoid writing to HBM?", "The n×n attention score and probability matrices. It uses tiling plus an online softmax."},
			{"Which Stage 6 idea does PagedAttention borrow?", "Paging: a per-sequence block table maps logical KV positions to fixed-size physical blocks."},
		},
	},
}
