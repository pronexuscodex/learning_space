package main

// Track F: Foundations of Computing (stages 1–4). The starting point for a
// complete beginner: programming in C, data structures & algorithms,
// discrete mathematics, and how the hardware itself works.

var foundationsGuides = map[int]StageGuide{
	// -----------------------------------------------------------------
	1: {
		Overview: `Programming is telling a very fast, very literal machine exactly what to
do. This stage teaches the building blocks every language shares
(values, variables, decisions, loops, functions and collections) in C,
then what real C programs are made of: structs, memory you allocate and
free yourself, files, and programs split across several files. C is the
language that lets you see what is really happening underneath: every
variable has a size and an address, nothing is hidden, and the compiler
turns your text into real machine instructions you can run.

Why start with C? It is small (you can learn the whole language), it is
where operating systems, databases, language runtimes and embedded
devices are written, and every idea you learn in it (memory, pointers,
the stack) explains what Python, Go or JavaScript do for you behind the
scenes. Harvard's CS50 moves to C in its first week, and Stanford's
CS107, Berkeley's CS61C and CMU's 15-213 teach systems in C. C also lets
you make mistakes other languages hide, so this stage teaches you to
write it defensively from day one: that is what makes you solid, not
the language itself.`,
		Outcomes: []string{
			"Write, compile, run and debug small C programs from scratch",
			"Explain what a variable, an array, a string and a pointer are in memory",
			"Break a problem into functions and read compiler errors and warnings calmly",
			"Recognise undefined behaviour and use the compiler's warnings and sanitizers to catch it",
			"Organise data with structs, manage heap memory with malloc and free without leaks, and read and write files safely",
			"Split a program into modules with headers, and build it with make",
		},
		Glossary: []Term{
			{"Source file", "The text file you write (hello.c) before the compiler turns it into a program."},
			{"Compiler", "The program (gcc or clang) that translates C source code into machine code the CPU can run."},
			{"Executable", "The runnable file the compiler produces, such as ./hello."},
			{"Variable", "A named piece of memory holding a value of a fixed type, such as int total = 42;."},
			{"Type", "The kind and size of a value (int, double, char), which decides what you can do with it."},
			{"printf", "C's standard function for printing formatted text, such as printf(\"%d\\n\", total);."},
			{"Function", "A named, reusable piece of code that takes inputs and returns an output."},
			{"Array", "A fixed-size row of values of the same type, stored side by side in memory."},
			{"Pointer", "A variable that holds the memory address of another value."},
			{"Undefined behaviour", "An operation C gives no meaning to (such as reading past an array); anything may happen."},
			{"Warning", "A message from the compiler about code that is legal but probably wrong. Treat it as an error."},
			{"Segfault", "The crash you get when a program touches memory it is not allowed to."},
			{"Struct", "A type that groups named fields into one value, such as struct point { int x, y; }."},
			{"Heap", "Memory you request while the program runs (malloc) and must give back yourself (free)."},
			{"Memory leak", "Memory that was allocated and never freed, so the program keeps growing."},
			{"Stream", "A FILE * the C library reads or writes through a buffer: a file, or stdin, stdout and stderr."},
			{"Header file", "A .h file of declarations that other .c files include to use a module."},
			{"Preprocessor", "The step before compiling that handles #include, #define and #ifndef."},
			{"Makefile", "The file that tells make how to build a program, and which files each step depends on."},
		},
		Concepts: []Concept{
			{
				Name:    "Your First C Program: Source, Compiler, Executable",
				Summary: "You write text, the compiler turns it into machine code, and the operating system runs it.",
				Body: `A C program starts as a plain text file, such as hello.c. It cannot run
yet: the CPU understands only machine code. A compiler (gcc or clang)
reads your source file, checks it, and translates it into an
executable, a file of machine instructions the operating system can
load and run.

Every C program starts at a function called main. #include <stdio.h>
brings in the declarations for standard input and output, including
printf. main returns an int to the operating system: 0 means success.

Get a compiler first. On Linux, install gcc (sudo apt install
build-essential, or your distribution's equivalent). On macOS, run
xcode-select --install to get clang. On Windows, install WSL (the
Windows Subsystem for Linux) and use Linux's gcc, or MSYS2. Then compile
with warnings turned on, every time:

  gcc -Wall -Wextra -std=c17 -g hello.c -o hello
  ./hello

-Wall -Wextra turns on warnings, -std=c17 picks the language standard,
-g adds debugging information, and -o names the output file.`,
				Diagram: `  hello.c          your source text
     │ preprocessor  pastes in #include files
     ▼ compiler      C → assembly
     │ assembler     assembly → machine code
     ▼ linker        + the C library (printf…)
  ./hello          the executable you run`,
				UnderTheHood: `The executable gcc produces is an ELF file (on Linux): a header, your
machine code in a section called .text, fixed data such as "Hello,
world!\n" in .rodata, and a table of symbols. main is not where it
starts. The real entry point is _start, a few instructions from the C
runtime, which calls __libc_start_main, which sets things up and calls
your main, then passes its return value to exit:

  nm hello | grep -E ' _start| main'     # T _start, T main
  strace ./hello                          # execve ... write(1, "Hello, world!\n", 14) ... exit_group(0)

Your printf becomes a single write system call of 14 bytes to file
descriptor 1 (standard output). printf itself does not talk to the
screen; it fills a buffer, and the C library hands the buffer to the
kernel. When output goes to a terminal the buffer is flushed at every
newline; when it goes to a file or a pipe it is flushed only when full
or at exit, which is why output order can surprise you when you
redirect a program that also prints to stderr.`,
				MentalModel: "Source is a recipe; the compiler cooks it into a program; you run the program, not the recipe.",
				TryIt: `Save this as hello.c, compile it with gcc -Wall -Wextra -std=c17 -g hello.c -o hello, run ./hello, then remove the semicolon and read the error.

  #include <stdio.h>

  int main(void) {
      printf("Hello, world!\n");
      return 0;
  }`,
				Analogy: `Writing a letter in English to someone who reads only Morse code. You
write the letter (source code), a translator converts it into dots and
dashes (the compiler), and only the Morse version is delivered (the
executable). If your English has a grammar mistake, the translator
refuses and tells you the line (a compile error).`,
				Example: `Every app on your phone started as source code that a compiler turned
into machine code for its processor. The Linux kernel, Git, SQLite,
Python's own interpreter (CPython) and the firmware in your car and
washing machine are all C programs, compiled exactly like hello.c.`,
				Exercises: trio(
					`Label each part of hello.c: the #include line, main, the printf call, the \n, and return 0. Then explain what "gcc hello.c -o hello" produces, and why you type ./hello rather than hello.c to run it.`,
					`\n is the newline character. The compiler's output is a new file: the executable.`,
					`Write a program that prints your name, your age next year, and a 5-line triangle of stars. Compile it with -Wall -Wextra, and fix every warning until there are none.`,
					`printf("%d\n", age + 1); prints a number. Warnings are the compiler pointing at likely bugs.`,
					`Install a compiler on your own machine (gcc, clang, or WSL on Windows), then write a two-file program: greet.c with a function greet(), and main.c that calls it. Compile both into one executable and explain what each step of the build did.`,
					`gcc -Wall -Wextra main.c greet.c -o greet. Put the declaration void greet(void); in a header greet.h and #include it from both files.`,
				),
			},
			{
				Name:    "Values, Types & Variables",
				Summary: "A variable is a named piece of memory with a fixed type and size.",
				Body: `In C, every variable is declared with a type before it is used, and the
type fixes its size in memory:

  int count = 3;         // a whole number, usually 4 bytes
  double price = 9.99;   // a floating-point number, usually 8 bytes
  char grade = 'A';      // one byte; 'A' is stored as the number 65

sizeof(int) tells you the size on your machine. Assignment (count =
count + 1) means "work out the right-hand side, then store it in the
variable on the left": an instruction, not an equation.

Types matter for every operation. 7 / 2 is 3 in C, because both are
ints and integer division throws the fraction away; 7.0 / 2 is 3.5. A
cast converts explicitly: (double)7 / 2 is 3.5, and (int)3.99 is 3
(the fraction is cut off, not rounded). printf needs the matching
format: %d for int, %f for double, %c for char, %s for strings.

A local variable that you never assigned holds leftover garbage, not 0.
Always initialise your variables.`,
				Diagram: `  memory (one byte per box)
  address:  1000 1001 1002 1003   1004 ...
            [    int count = 3   ][ char grade = 'A' (65) ]
            └──── 4 bytes ──────┘ └─ 1 byte ─┘`,
				UnderTheHood: `Here is what gcc -O0 makes of int count = 3; inside main (x86-64,
Intel syntax):

  mov DWORD PTR -8[rbp], 3     ; store the 4-byte value 3, 8 bytes below the frame pointer

A local variable is just a slot on the stack at a fixed distance from
the frame pointer (rbp); its name is gone after compiling. DWORD means
4 bytes: the type decided the size. In memory, 3 is stored as the bytes
03 00 00 00, lowest byte first, because x86 and ARM are little-endian,
and -1 is ff ff ff ff, because negative numbers use two's complement.
With optimisation (-O2) the variable often never touches memory at all:
it lives only in a register, or disappears entirely if the compiler can
work out the value in advance.`,
				MentalModel: "A variable is a labelled box of a fixed size; the type says how to read the bits inside.",
				TryIt:       `Print sizeof(char), sizeof(int), sizeof(long), sizeof(double), then 7 / 2, 7.0 / 2, (int)3.99 and 'A' + 1 (with %d), and explain each result.`,
				Analogy: `Labelled jars of fixed sizes in a kitchen: a small jar for salt (a
char), a bigger one for flour (a double). Pour a litre into the small
jar and it overflows. Types are like units, too: adding 3 metres to 2
seconds makes no sense, and C's type checks stop many such mistakes
before the program ever runs.`,
				Example: `In 2020 scientists officially renamed several human genes, such as SEPT1
and MARCH1, because Excel kept silently converting their names into
dates: a type problem. In 1996 the Ariane 5 rocket destroyed itself 37
seconds after launch because a 64-bit floating-point value was
converted into a 16-bit integer that could not hold it.`,
				Exercises: trio(
					`Without running anything, predict the values: int x = 5; x = x * 2; x = x - 3; then 17 / 5, 17 % 5, 17.0 / 5 and (int)2.9. Then check with printf.`,
					"Integer division drops the fraction; % gives the remainder; a cast to int cuts off the fraction.",
					`Write a temperature converter: read a Celsius value with scanf("%lf", &c), print the Fahrenheit value to one decimal place (%.1f), and check scanf's return value so typing "hot" prints a polite error instead of garbage.`,
					`scanf returns how many values it read: if (scanf("%lf", &c) != 1) { ... }. F = C × 9 / 5 + 32; watch out for integer division.`,
					`Prices in a shop's system must never be stored as double. Write functions that parse "12.34" into 1234 cents (a long) and print 1234 back as "12.34", and test them on 0.10, 19.99, 1000.00 and 0.05.`,
					`Split at the dot yourself or read two integers with sscanf(text, "%ld.%2ld", &euros, &cents). Printing: printf("%ld.%02ld", c / 100, c % 100).`,
				),
			},
			{
				Name:    "Control Flow: Decisions & Loops",
				Summary: "if chooses, loops repeat, and together they can express any procedure.",
				Body: `Code normally runs top to bottom. if runs a block only when a condition
is true (non-zero, in C), with else if and else for the alternatives.
Conditions combine with && (and), || (or) and ! (not). switch picks
between many constant cases; remember each case needs a break.

A loop repeats a block. for (int i = 0; i < n; i++) runs n times, with
i from 0 to n − 1; while (condition) repeats until the condition becomes
false; do { } while (condition) always runs at least once. break leaves
a loop early; continue skips to the next iteration.

The classic bugs are off-by-one errors (i <= n instead of i < n),
infinite loops whose condition never changes, and if (x = 5), which
assigns instead of comparing. Always use braces, even for one line, and
let -Wall warn you about the rest.`,
				Diagram: `            ┌─────────────────────────────┐
            ▼                             │
 start ─▶ [ i < 10 ? ] ──yes──▶ [ work; i++ ]
              │
              no
              ▼
            done`,
				UnderTheHood: `The CPU has no if and no for. It has compare instructions, which set
flags, and jumps, which change the program counter. gcc -O2 turns a loop
that sums an array into, in essence:

  loop:  add eax, DWORD PTR [rdi]   ; s += *a
         add rdi, 4                 ; move to the next int (4 bytes)
         cmp rdi, rdx               ; reached the end?
         jne loop                   ; if not, jump back

and it may turn a simple if into no jump at all: clamp(x), which returns
100 when x > 100, becomes a compare and a cmovle ("conditional move"),
because a jump the CPU guesses wrongly costs around 15–20 cycles. That
guessing is branch prediction (Stage 4): loops whose conditions follow a
pattern run much faster than ones that jump unpredictably.`,
				MentalModel: "Sequence, choice and repetition: the three moves that build every program.",
				TryIt:       `Write FizzBuzz in C: print 1 to 100, but "Fizz" for multiples of 3, "Buzz" for multiples of 5 and "FizzBuzz" for both. Use % and a for loop.`,
				Analogy: `A recipe: "if the dough is sticky, add flour" is a decision, and "stir
until smooth" is a loop. A recipe that said "stir until the cake is
baked", without ever putting it in the oven, would have you stirring
forever. That is an infinite loop.`,
				Example: `A thermostat's firmware (very often written in C) is a loop with a
decision inside: every few seconds, if the room is colder than the
target, turn the heating on, otherwise turn it off. Traffic lights,
washing-machine programs and your router's "retry three times" are all
loops and conditions.`,
				Exercises: trio(
					"How many times does for (int i = 2; i < 10; i += 3) run, and with which values of i? And for (int i = 10; i > 0; i /= 2)?",
					"Trace it on paper: 2, 5, 8 for the first; 10, 5, 2, 1 for the second (integer division).",
					`Write a number-guessing game in C: pick a random number from 1 to 100 with srand(time(NULL)) and rand() % 100 + 1, let the user guess, reply "higher" or "lower", and count the attempts.`,
					"A while loop that ends when the guess is right. Include <stdlib.h> for rand and <time.h> for time.",
					`A gym charges €30 a month. Students get 20% off, and members of more than 2 years get another €5 off, but the price never drops below €15. Write int price_cents(int is_student, int years) and test it on five different people with assert.`,
					"Work in cents so there are no floating-point errors. Apply each rule with an if, then the floor: if (p < 1500) p = 1500;. #include <assert.h>.",
				),
			},
			{
				Name:    "Functions & Decomposition",
				Summary: "Name a piece of work once, reuse it everywhere, and break big problems into small ones.",
				Body: `A function takes inputs (parameters), does some work and returns an
output:

  double area(double width, double height) {
      return width * height;
  }

C needs to know a function's signature before it is called, so either
define it above main or declare a prototype first (double area(double,
double);). In bigger programs, prototypes go in a header file (.h) and
the code in a .c file.

C passes arguments by value: the function gets a copy. Changing a
parameter inside the function does not change the caller's variable
(to do that, you pass its address, as you will see with pointers). Each
call gets its own local variables on the stack, which disappear when
the function returns.

Decomposition is the core skill of programming. Split "build a to-do
app" into "load tasks", "add a task" and "save tasks", then split those
again until every piece is obvious. A function that calls itself
(recursion) is solving a smaller copy of the same problem.`,
				Diagram: `  stack (grows down)
  ┌────────────────────────┐
  │ main:  w = 3, h = 4    │
  ├────────────────────────┤
  │ area:  width = 3 (copy)│  ← gone when area returns
  │        height = 4      │
  └────────────────────────┘`,
				UnderTheHood: `On x86-64 Linux (the System V calling convention), the first six integer
arguments travel in registers (rdi, rsi, rdx, rcx, r8, r9) and the
result comes back in rax (eax for an int). With -O2,
int add(int a, int b) { return a + b; } compiles to just:

  lea eax, [rdi+rsi]    ; eax = a + b
  ret                   ; return to the caller

call pushes the return address onto the stack and jumps; ret pops it and
jumps back. Without optimisation (-O0) each function also builds a stack
frame (push rbp; mov rbp, rsp) and gives its locals slots in it, which
is exactly what the debugger shows you with bt. Pass-by-value is simply
this: the callee receives a copy of the value in a register or on the
stack, and has no idea where the caller's variable lives.`,
				MentalModel: "If you cannot name it, you have not understood it; if it is long, split it.",
				TryIt:       "Write double c_to_f(double c) and double f_to_c(double f), and check with assert that converting there and back returns the original value (within 1e-9).",
				Analogy: `A coffee machine: water and beans go in (inputs), you press one button,
and coffee comes out (the output). You do not need to know how the pump
works. And C's pass-by-value is like handing someone a photocopy: they
can scribble on it, but your original stays clean.`,
				Example: `The C standard library is a set of functions like this: strlen, qsort,
fopen. Every "Log in with Google" button calls a small set of shared
functions, and large programs such as SQLite are built from thousands
of small, well-named C functions.`,
				Exercises: trio(
					`Predict the output, then run it: void bump(int n) { n = n + 1; } then in main: int x = 5; bump(x); printf("%d\n", x);. Explain why.`,
					"C passes a copy. bump changes its own n, not main's x.",
					`Write int is_palindrome(const char *s), ignoring spaces, punctuation and capital letters (use isalnum and tolower from <ctype.h>), and test it on "A man, a plan, a canal: Panama".`,
					"Two indexes, one from each end, skipping characters that are not letters or digits, moving towards the middle.",
					`Decompose a "split the restaurant bill" program into functions (read the items, assign items to people, add tax and tip, round to cents so the shares still add up to the total), with prototypes in bill.h and code in bill.c, then implement it.`,
					"Work in integer cents. Rounding each share can lose a cent, so give leftover cents to someone explicitly.",
				),
			},
			{
				Name:    "Memory, Addresses & Pointers",
				Summary: "Every value lives at an address; a pointer is a variable that holds one.",
				Body: `Memory is one long row of numbered bytes. Every variable lives at some
address. The & operator gives a variable's address, and a pointer is a
variable that stores an address:

  int x = 5;
  int *p = &x;    // p holds the address of x
  *p = 7;         // follow the pointer: x is now 7

The * in a declaration (int *p) says "p is a pointer to an int"; the *
in an expression (*p) means "go to that address" (dereferencing).

Pointers let a function change the caller's variables, which is why
scanf needs &: void swap(int *a, int *b) swaps two ints in place. They
also let you pass large data without copying it, and they are how
arrays, strings and dynamic memory work.

NULL is a pointer that points nowhere; dereferencing it crashes the
program (a segfault). Never return the address of a local variable: it
lives on the stack and disappears when the function returns.`,
				Diagram: `  address   1000          1008
           ┌──────────┐  ┌──────────┐
           │ x = 7    │◀─│ p = 1000 │
           └──────────┘  └──────────┘
            int x         int *p  (p points to x)`,
				UnderTheHood: `A pointer is nothing more than a number: an address, 8 bytes on a
64-bit machine. Dereferencing is a single load instruction.
int get(int *p) { return *p; } becomes:

  mov eax, DWORD PTR [rdi]   ; read 4 bytes from the address in rdi
  ret

The addresses you print are virtual: the CPU's memory-management unit
translates them to physical RAM through page tables that the operating
system maintains (Stage 6). They also change between runs, because the
system loads the stack, heap and libraries at random positions (address
space layout randomisation) to make attacks harder. A NULL dereference
crashes because the kernel deliberately leaves the page at address 0
unmapped: the CPU raises a fault, and the kernel sends your process
SIGSEGV.`,
				MentalModel: "A pointer is a street address: & asks where something lives, * goes there.",
				TryIt:       `Print a variable's address with printf("%p\n", (void *)&x), then write swap(int *a, int *b), and watch both in Python Tutor's C mode (pythontutor.com/c.html), which draws pointers as arrows.`,
				Analogy: `A house and its address. The house (the value) stays where it is; you
can write its address on a card (a pointer) and give the card to
anyone, and they can go and repaint the house. Copying the card is
cheap; copying the house is not. A card with no address on it (NULL)
leads nowhere, and trying to visit it ends badly.`,
				Example: `When you pass a 4 GB video to an editing program's functions, they pass
a pointer to it (8 bytes), never the video itself. Linked lists, trees,
your browser's DOM and the operating system's process table are all
built from pointers linking pieces of memory together.`,
				Exercises: trio(
					`Given int a = 3; int *p = &a; *p = *p + 1; int b = *p * 2;, what are a and b? Draw the boxes and the arrow.`,
					"*p is a, so a becomes 4, and b is 8.",
					`Write void min_max(const int *arr, int n, int *min, int *max) that finds both the smallest and largest value in one pass and "returns" them through the two pointers. Test it on an array with negative numbers.`,
					"Start both at arr[0], then loop from index 1. Write the results with *min = ...; *max = ...;.",
					`Write a function that splits "2026-09-24" into three ints through pointers (int parse_date(const char *s, int *y, int *m, int *d)), returning 0 on success and -1 on malformed input, and test it on five good and five bad dates.`,
					`sscanf(s, "%d-%d-%d", y, m, d) == 3, then check the ranges yourself (month 1–12, day 1–31). Note there is no & here: y is already a pointer.`,
				),
			},
			{
				Name:    "Collections: Lists, Dictionaries & Strings",
				Summary: "Arrays, strings and structs are C's building blocks; lists and dictionaries are built from them.",
				Body: `An array is a fixed-size row of values of the same type, side by side in
memory, indexed from 0:

  int scores[5] = {90, 72, 85, 60, 99};
  scores[0]                 // 90; the last valid index is 4

C does not check bounds: scores[5] reads whatever memory comes next.
Always carry the length with the array (arrays passed to functions
decay into a pointer to their first element, so the function cannot
know the length unless you pass it).

A string in C is an array of char ending in a zero byte ('\0'), which
marks the end. "hi" takes 3 bytes: 'h', 'i', '\0'. strlen counts up to
the zero; strcmp compares; snprintf builds strings safely within a
size. Forgetting room for the '\0' is a classic bug.

A struct groups named fields into one value (struct point { int x; int
y; };). The flexible collections of other languages are built from
these parts: a dynamic list is an array that is reallocated when full,
and a dictionary is a hash table of structs, which you will build
yourself in Stage 2.`,
				Diagram: `  char name[6] = "Ada";
  ┌─────┬─────┬─────┬─────┬─────┬─────┐
  │ 'A' │ 'd' │ 'a' │ \0  │  ?  │  ?  │
  └─────┴─────┴─────┴─────┴─────┴─────┘
    [0]   [1]   [2]   [3]   [4]   [5]    strlen = 3, size = 6`,
				UnderTheHood: `a[i] is arithmetic: the address of a plus i times the element size. For
an int array, int at(int *a, long i) { return a[i]; } compiles to one
instruction:

  mov eax, DWORD PTR [rdi+rsi*4]   ; load from a + i*4

No bounds check exists anywhere, which is why C is fast, and why an
index past the end silently reads the neighbouring memory. A string
literal such as "hi" is stored once in the read-only data section as the
bytes 68 69 00; writing to it crashes. Structs are laid out in order,
with padding so each field sits at an address its type prefers:
struct { char c; int x; } takes 8 bytes, not 5, because x must start at
a multiple of 4. Arrays of structs are contiguous, which makes them
cache-friendly (Stage 5).`,
				MentalModel: "An array is numbered lockers in a row; a string is the same, ending with an empty locker marked \\0.",
				TryIt:       "Count how often each letter a–z appears in a line of text, using an int counts[26] array indexed by c - 'a', then print the five most common letters.",
				Analogy: `A row of numbered lockers: you go straight to locker 7, but if you ask
for locker 100 in a row of 10, C will open whatever is behind the wall.
A string is a sentence written on a strip of paper with a full stop
('\0') marking where it ends; lose the full stop and the reader carries
on into whatever is written next.`,
				Example: `The 2014 Heartbleed bug in OpenSSL let attackers read up to 64 KB of a
server's memory, including passwords and private keys, because the code
trusted a length the attacker sent and copied that many bytes from a
buffer without checking its real size. The same arrays and lengths you
use here, with one missing check.`,
				Exercises: trio(
					`How many bytes does char word[] = "hello"; take, and what is strlen(word)? What is wrong with char s[5] = "hello";?`,
					"6 bytes (5 letters plus '\\0'); strlen is 5. s[5] has no room for the terminating zero.",
					`Define struct student { char name[32]; int scores[5]; int n; }, fill an array of five students, compute each one's average, and print them sorted from highest to lowest average with qsort.`,
					"qsort needs a comparison function: int cmp(const void *a, const void *b) that casts to const struct student *.",
					`Read a CSV of bank transactions (date,category,amount) line by line with fgets, total the spending per category in a fixed array of structs (at most 32 categories), and print the three largest categories. Reject lines that are too long or malformed instead of crashing.`,
					"strchr or strtok to split at commas; strcmp to find the category; check that each line ends with '\\n' to detect lines longer than your buffer.",
				),
			},
			{
				Name:    "Structs: Grouping Related Data",
				Summary: "A struct bundles named fields into one value, so a program can talk about a student instead of three loose variables.",
				Body: `Real data comes in groups: a student has a name, an age and an average.
A struct gives the group one name and one type:

  struct student {
      char   name[32];
      int    age;
      double average;
  };

  struct student s = {"Ada", 36, 91.5};
  s.age = 37;                       // . reaches a field

A struct is a value like any other. Assigning it (s2 = s) copies every
field, passing it to a function gives the function a copy, and a
function can return one. When the struct is big, or the function must
change the caller's struct, pass a pointer instead and reach the fields
with ->:

  void birthday(struct student *p) { p->age++; }   // p->age means (*p).age
  birthday(&s);

An array of structs is a table of records (struct student class[30]),
and typedef gives a type a shorter name (typedef struct student
Student;). Compare structs field by field: == does not compile for
structs, and memcmp is unreliable because of the padding bytes the
compiler may add between fields (see below).`,
				Diagram: `  struct student s                         sizeof = 48 on x86-64
  ┌──────────────────────┬────────┬─────────┬──────────────┐
  │ name[32]  "Ada\0…"   │ age 37 │ padding │ average 91.5 │
  └──────────────────────┴────────┴─────────┴──────────────┘
   offset 0               32       36        40`,
				UnderTheHood: `The compiler turns every field into a fixed offset from the start of
the struct. With the layout above, p->age is "the 4 bytes at p + 32",
and int get_age(struct student *p) { return p->age; } compiles with
gcc -O2 to:

  mov eax, DWORD PTR [rdi+32]   ; load 4 bytes, 32 past the pointer
  ret

Each field starts at a multiple of its own alignment (4 for an int, 8
for a double), so the compiler inserts padding: average cannot start at
36, so it starts at 40. Field order therefore changes the size.
struct { char a; double b; char c; } takes 24 bytes; reordered as
{ double b; char a; char c; } it takes 16. offsetof(type, field), from
stddef.h, tells you where any field sits.`,
				MentalModel: "A struct is a form with labelled boxes; the compiler knows how far down the form each box sits.",
				TryIt:       "Define struct student, print sizeof(struct student) and offsetof(struct student, average) with %zu, then reorder the fields and predict the new size before you print it.",
				Analogy: `A library's index card for each book: title, author, year and shelf on
one card. You hand someone the whole card, not four scraps of paper,
and everyone knows where on the card the year is written.`,
				Example: `Every file on your disk is described by a struct inside the operating
system: on Linux and macOS, struct stat holds a file's size, owner,
permissions and timestamps, and it is what ls -l prints. Network packet
headers, image file headers and the entities in a game are all structs
laid out field by field.`,
				Exercises: trio(
					`What is sizeof(struct { char c; int x; char d; }) on a typical 64-bit machine, and how would you reorder the fields to make it smaller?`,
					"12: c at 0, x at 4 (it must be aligned to 4), d at 8, then padding up to a multiple of 4. With x first it becomes 8.",
					`Define struct point { double x, y; } and write double distance(struct point a, struct point b) and struct point midpoint(struct point a, struct point b). Then write void translate(struct point *p, double dx, double dy) and explain why it takes a pointer when the other two do not.`,
					"distance uses sqrt from math.h (link with -lm on Linux). translate must change the caller's point, so it needs its address: p->x += dx.",
					`Model a small library: struct book { char title[64]; char author[64]; int year; int on_loan; }. Read up to 100 books from a file of lines title;author;year;on_loan, then print the books on loan and the oldest book. Put the parsing of one line in its own function that returns 0 or -1.`,
					"fgets each line, strchr to find each ';', copy each field with a length check, and strtol the numbers. An array of 100 structs and a count is all the storage you need.",
				),
			},
			{
				Name:    "Dynamic Memory: malloc, free & Leaks",
				Summary: "When the size is only known while the program runs, ask for memory on the heap, and give back every byte you take.",
				Body: `Local variables live on the stack: fast, automatic, freed when the
function returns, and fixed in size when you write the program. When
the size is only known at run time (a file of unknown length, a list
that keeps growing), you ask for memory on the heap:

  int *a = malloc(n * sizeof *a);     // room for n ints
  if (a == NULL) { /* out of memory: report it and stop */ }
  /* … use a[0] … a[n - 1] … */
  free(a);                            // give it back, exactly once
  a = NULL;

The rules:
- check every malloc, calloc and realloc for NULL;
- calloc(n, size) is malloc plus zeroed memory, and the C library checks n * size for overflow;
- realloc(p, new_size) grows or shrinks a block and may move it, so keep the old pointer until it succeeds: tmp = realloc(p, size); if (tmp != NULL) p = tmp;
- free every block exactly once: forgetting is a leak, freeing twice is a double free, and using it afterwards is a use-after-free;
- for every allocation, decide which function owns it and is responsible for freeing it.

The growable array is the workhorse of C programs: keep a pointer, a
length and a capacity, and when the length reaches the capacity, double
the capacity with realloc. Because the capacity doubles, each element is
copied fewer than two times on average, so appending stays fast however
long the array gets.`,
				Diagram: `  stack (automatic)              heap (malloc / free)
  ┌──────────────────┐          ┌────┬────┬────┬────┐
  │ int *data ───────┼────────▶ │ 10 │ 20 │ 30 │  ? │   len 3, cap 4
  │ size_t len = 3   │          └────┴────┴────┴────┘
  │ size_t cap = 4   │
  └──────────────────┘
   gone when the function        stays until free(data)
   returns`,
				UnderTheHood: `malloc is ordinary library code, not a system call. The C library's
allocator keeps lists of free chunks of different sizes, and each chunk
has a small header just before the pointer you are given, recording its
size. Small requests come from the heap region, which the allocator
grows with the brk system call when it runs out; large ones (glibc's
default threshold is 128 KiB) get their own pages straight from the
kernel with mmap, and free returns those with munmap. That is why
strace shows only a handful of brk calls for millions of small mallocs.

Writing past the end of a block, or freeing it twice, corrupts those
headers, which is why heap bugs crash far from where they happened and
why they become security holes (Stage 10). The address sanitizer puts
poisoned red zones around every block to catch the overflow at the
exact line, and its leak checker lists every block still allocated
when the program exits, with the line that allocated it.`,
				MentalModel: "Every malloc is a loan: the code that borrows it must know who pays it back, and when.",
				TryIt:       "Read numbers until end of input into a growing array (capacity 4 at first, doubled with realloc), print their sum and free the array. Run it with -fsanitize=address; then delete the free, run it again, and read the leak report.",
				Analogy: `Booking a table at a restaurant. The stack is the seat you take while
you eat: when you leave, it is free again. The heap is a table you
reserve for as long as you like: it stays reserved until you cancel.
Forget to cancel and the restaurant slowly fills with empty reserved
tables (a leak); cancel twice and you may cancel someone else's booking
(a double free).`,
				Example: `A long-running program that leaks a little memory on every request
grows until the system runs out and kills it: browsers, games and
servers have all shipped leaks that only showed after hours of use.
Valgrind, the sanitizers' leak checker and heap profilers exist to find
them, and garbage collection and Rust's ownership rules (Stage 12)
exist largely to make these mistakes impossible.`,
				Exercises: trio(
					`What is wrong with each line? int *p = malloc(10); p[9] = 1;   char *s = malloc(strlen(t)); strcpy(s, t);   free(p); free(p);`,
					"malloc(10) is 10 bytes, room for only two 4-byte ints; strlen leaves no room for the '\\0'; and freeing twice is undefined behaviour.",
					`Write char *join(const char *a, const char *b), which returns a newly allocated string holding a, a space and b, and test it, freeing every result. Then write char *repeat(const char *s, int n).`,
					"Allocate strlen(a) + 1 + strlen(b) + 1 bytes: the space and the terminating zero. The caller frees the result: say so in a comment above the function.",
					`Read a text file of unknown length line by line into a growing array of strings (each line its own allocated copy), print the lines in reverse order, and free everything. Make it work for an empty file, a one-line file and a million-line file, with no leaks reported by -fsanitize=address.`,
					"An array of char * that you realloc when full, and a copy of each line made with malloc(len + 1) and memcpy. Free each line, then the array. getline (POSIX) reads lines of any length.",
				),
			},
			{
				Name:    "Safe Input & Undefined Behavior",
				Summary: "C trusts you completely, so check every input, every size and every return value.",
				Body: `C gives you power, not protection. Many mistakes are undefined
behaviour: the C standard says nothing about what happens, so the
program may crash, seem to work, or corrupt data silently, and the
compiler is allowed to assume they never happen. The usual suspects:
- reading or writing outside an array;
- using an uninitialised variable;
- dereferencing NULL, or memory that has been freed;
- signed integer overflow (INT_MAX + 1);
- writing a string without room for its '\0'.

Being "bulletproof" in C is a discipline, not a feature:
- read input with fgets into a buffer of known size, then convert it with strtol or strtod and check that the whole input was used;
- never use gets (removed from the language) and avoid scanf("%s") without a width;
- check every return value (fopen can return NULL, malloc can fail);
- pass sizes with arrays, and prefer bounded functions such as snprintf;
- compile with -Wall -Wextra, and while developing add -fsanitize=address,undefined, which makes most of these bugs crash loudly at the exact line.`,
				Diagram: `  input
    │ fgets into buf[64]          too long? → reject, ask again
    ▼
  strtol: a whole number? in range?    no → reject, ask again
    │ yes
    ▼
  use it`,
				UnderTheHood: `Undefined behaviour is not just "the program might crash". The compiler
is allowed to assume it never happens, and optimises on that basis. This
check for overflow:

  int overflow_check(int x) { return x + 1 < x; }

compiles with gcc -O2 to:

  xor eax, eax     ; return 0, always
  ret

Signed overflow is undefined, so x + 1 < x "can never be true", and the
check is deleted. (Write x == INT_MAX instead.) This is why "it worked
in debug mode" proves nothing. The address sanitizer catches memory
bugs by surrounding every array and allocation with poisoned "red
zones" and checking every load and store against a shadow map of
memory, at the cost of running about twice as slowly: a good trade
while you develop and test.`,
				MentalModel: "Never trust input, never assume a size, never ignore a return value.",
				TryIt:       `Compile int a[3] = {0}; a[3] = 1; first normally, then with gcc -g -fsanitize=address,undefined, and compare what each run tells you.`,
				Analogy: `A kitchen knife with no guard: in skilled hands it is the best tool
there is, and careless hands get cut. Professional chefs do not use
blunt knives; they use habits (cut away from yourself, keep it clean,
put it down flat). Defensive C is those habits for code.`,
				Example: `The 1988 Morris worm, one of the first internet worms, spread partly
through a buffer overflow in the fingerd server, which read input with
gets. Decades later, Microsoft and Google have each reported that about
70% of their serious security bugs are memory-safety problems of
exactly these kinds, which is why the habits in this concept matter.`,
				Exercises: trio(
					`For each line, say whether it is undefined behaviour and why: int x; printf("%d", x);   char s[4]; strcpy(s, "four");   int *p = NULL; *p = 1;   unsigned char c = 255; c++;`,
					"The last one is fine: unsigned arithmetic wraps around by definition. The first three are undefined.",
					`Write int read_int(const char *prompt, int min, int max) that uses fgets and strtol, rejects empty input, trailing junk ("12abc"), out-of-range numbers and overlong lines, and asks again until the input is valid.`,
					"strtol sets an end pointer: after skipping trailing whitespace it must point at '\\0'. Check errno == ERANGE and the min/max range.",
					`Take one of your earlier programs (the bill splitter or the CSV reader), compile it with -fsanitize=address,undefined, and feed it hostile input: empty lines, 10,000-character lines, negative numbers, missing fields. Fix every crash the sanitizers find, and write down each bug you found.`,
					`Generate hostile input with python3 -c "print('A' * 10000)" | ./program. Each sanitizer report names the line and the kind of bug.`,
				),
			},
			{
				Name:    "Files & Streams: Reading and Writing Data",
				Summary: "A file is a stream of bytes: open it, read or write it a piece at a time, check every step, and close it.",
				Body: `Programs keep data between runs in files. The C library gives you
streams (FILE *), which buffer reads and writes for speed:

  FILE *f = fopen("notes.txt", "r");    // "r" read, "w" write (empties it first), "a" append
  if (f == NULL) { perror("notes.txt"); return 1; }
  char line[256];
  while (fgets(line, sizeof line, f) != NULL) {
      /* use line */
  }
  if (ferror(f)) perror("notes.txt");
  fclose(f);

Three streams are always open: stdin, stdout and stderr. That is why the
same loop can read a file or a pipe: fgets(line, sizeof line, stdin)
reads whatever the user or another program sends.

The tools: fgetc and fputc for one byte, fgets and fputs for lines,
fprintf for formatted text, and fread and fwrite for raw bytes. Each
reports failure in its return value (NULL, EOF or a short count), and
perror prints why, using errno ("No such file or directory"). A text
file is just bytes, and so is a photo: the difference is how you read
them. strlen stops at the first zero byte, so binary data needs fread
and an explicit length. (On Windows, text mode also translates line
endings: open with "rb" or "wb" to get the bytes exactly.)

When writing, data waits in the stream's buffer until it is flushed.
fclose (or fflush) hands it to the operating system and reports a
failed write, so check what fclose returns too. Even then it sits in
the kernel's page cache, not necessarily on the disk (Stage 7).`,
				Diagram: `  your program          C library               kernel           disk
  fprintf(f, …) ──▶  FILE buffer (a few KiB) ──▶ page cache ──▶ ▓▓▓▓
                      flushed by fflush,          write()        fsync
                      fclose or exit`,
				UnderTheHood: `A FILE is a struct inside the C library holding a file descriptor (the
small integer the kernel handed out when the file was opened; 0, 1 and
2 are stdin, stdout and stderr) and a buffer. fopen makes the openat
system call. Reading a 1 MB file with fgets makes about 260 read system
calls, not a million, because each read fills a 4 KiB buffer that fgets
then serves from:

  strace -c -e trace=read ./program big.txt     # count the read calls

When main returns, exit flushes and closes every open stream, which is
why output still appears when you forget fclose, and why it can be lost
when the program crashes before exiting.`,
				MentalModel: "Open, check, loop over reads or writes, check, close; stdin is just a file that is already open.",
				TryIt:       "Write a program that copies one file to another with fgetc and fputc, reporting errors with perror. Copy a photo with it and compare the two files with cmp, then run it under strace -c and count the read calls.",
				Analogy: `A library's reading room. You fill in a request slip (fopen) and get a
numbered ticket (the FILE *). The librarian brings pages in batches
rather than one word at a time (the buffer). When you are done you
return the book (fclose); if you leave without returning it, someone
tidies up at closing time (exit), unless the building burns down first
(a crash).`,
				Example: `grep, sort and wc read stdin and write stdout, which is what lets you
join them with pipes (Stage 9). Configuration files, logs, CSV exports
and saved games are read and written exactly this way. This academy's
own progress file is written to a temporary file, synced and renamed
(Stage 7), so a crash never leaves it half written.`,
				Exercises: trio(
					`What does fopen("out.txt", "w") do to an existing out.txt, and how is "a" different? What should a program do when fopen returns NULL?`,
					`"w" empties the file first; "a" adds to its end. Report why with perror and stop, or try something else: never carry on with a NULL FILE *.`,
					`Write a program that prints the longest line of the file named by its first argument, with the line's number. Handle a missing argument, a file that cannot be opened and an empty file, each with a clear message on stderr and a non-zero exit status.`,
					"Keep a copy of the longest line so far and its number. A line longer than your buffer arrives in several pieces: decide what you do with those, and say so in a comment.",
					`Give the budget tracker a save file: write an array of transaction structs to a text file, one per line, read it back into a second array, and check that the two match. A truncated or corrupted file must fail with a clear message, never fill the array with garbage.`,
					"fprintf one line per struct with a fixed field order; read with fgets and sscanf, and check that sscanf returned the number of fields you expected.",
				),
			},
			{
				Name:    "Headers, the Preprocessor & make",
				Summary: "Split a program into files that compile separately, share declarations through headers, and let make rebuild only what changed.",
				Body: `Real programs are many files. Each .c file is compiled on its own into
an object file (.o), and the linker joins the objects into one
executable. For that to work, each file must see the declarations of
what it uses from the others. That is what headers are for:

  /* stack.h: the interface */
  #ifndef STACK_H
  #define STACK_H
  struct stack { int items[64]; int top; };
  void stack_push(struct stack *s, int x);
  #endif

  /* stack.c: the implementation, and main.c: a user */
  #include "stack.h"

  gcc -Wall -Wextra -c stack.c     # → stack.o
  gcc -Wall -Wextra -c main.c      # → main.o
  gcc stack.o main.o -o app        # link

#include, #define and #ifndef are preprocessor directives: before the
compiler proper sees your code, the preprocessor pastes in included
files, replaces macros and keeps or drops lines. The #ifndef … #endif
include guard makes a header safe to include twice. Put declarations in
headers and definitions in .c files, and mark helpers that only one
file uses as static, so they stay private to it.

make automates the build. A Makefile lists each target, the files it
depends on, and the command that makes it:

  app: main.o stack.o
      gcc main.o stack.o -o app
  %.o: %.c stack.h
      gcc -Wall -Wextra -g -c $< -o $@

make compares file times and reruns only the commands whose inputs
changed, which turns minutes into seconds on big projects. The indented
command lines must start with a real Tab character, not spaces.`,
				Diagram: `  stack.h  (included by both)
     │
  stack.c ──▶ gcc -c ──▶ stack.o ──┐
  main.c  ──▶ gcc -c ──▶ main.o  ──┴──▶ linker ──▶ app`,
				UnderTheHood: `gcc -E main.c shows the program after preprocessing: every #include is
replaced by the whole text of the header (stdio.h alone expands to
about 800 lines). Each object file has a symbol table: nm stack.o lists
stack_push with T (defined here, in the code section), and nm main.o
lists it with U (undefined: needed from elsewhere). The linker matches
every U with exactly one T. Two definitions of one name give "multiple
definition"; none gives "undefined reference", the most common linker
error, which usually means a .c file is missing from the command.

Macros are text substitution, not functions: with #define SQUARE(x) x*x,
SQUARE(1+2) becomes 1+2*1+2, which is 5. Write ((x)*(x)), or better, a
static inline function, which the compiler checks like any other.`,
				MentalModel: "A header is a promise (this exists, and looks like this); a .c file keeps it; the linker checks every promise was kept exactly once.",
				TryIt:       "Split one of your programs into main.c, a module .c and its .h. Build it by hand with gcc -c and link it, then write a Makefile. Run make twice (the second does nothing), touch one .c file and watch make rebuild only that file. Then run nm on each .o.",
				Analogy: `A restaurant menu and its kitchen. The menu (the header) promises what
can be ordered and how to ask for it; the kitchen (the .c file) makes
it. Waiters (other files) only need the menu. Before opening, the
manager (the linker) checks that every dish on the menu has exactly one
station that cooks it.`,
				Example: `The Linux kernel has tens of thousands of .c files and is built with
make: change one file and only what depends on it is rebuilt. Every C
library you install (SQLite, zlib, OpenSSL) ships as headers plus
compiled code: you include its header and link its library, as in
gcc app.c -lsqlite3.`,
				Exercises: trio(
					`Which goes in the header and which in the .c file: a function prototype, a function body, a struct type that callers use, a static helper function, the definition of a global variable? Why does a header need an include guard?`,
					"Prototypes and shared struct types in the header; bodies, static helpers and variable definitions in the .c file (the header gets only an extern declaration). Without a guard, including it twice defines the struct twice.",
					`Turn the temperature functions into a module: temperature.h with the prototypes and an include guard, temperature.c with the code, and main.c printing a table. Build it with a Makefile whose first target is the program, plus a clean target.`,
					".PHONY: clean, then clean: rm -f *.o and the program. Make each .o depend on the header too, so editing it rebuilds them.",
					`Give your largest C program a real structure: one module per job (input, storage, reports), each with a header that exposes only what the others need, a Makefile with -Wall -Wextra -Werror, and a test target that builds and runs a small test program. Count how many functions you could make static.`,
					"A test target can build test_storage.c against storage.o and run it. If one module needs another's internals, the interface between them is probably wrong.",
				),
			},
			{
				Name:    "Debugging & Reading Errors",
				Summary: "Bugs are normal; finding them is a skill you can learn.",
				Body: `There are three kinds of problems. Compile errors stop the program from
being built: read the first error first (later ones are often caused by
it), and look at the line it names and the line before. Warnings are
legal code that is probably wrong: treat every warning as an error.
Runtime problems happen while it runs: a segfault, a sanitizer report,
or, worst of all, a silently wrong answer.

Use the scientific method: reproduce the bug reliably, form a
hypothesis, test it, fix it, then add a test so it never comes back.
Your tools:
- printf debugging: print the values you think you know, to stderr with fprintf(stderr, ...);
- gdb (or lldb on macOS): run the program under a debugger, stop at a line (break), step through it (next, step), print variables (print x), and see the call stack after a crash (bt, for backtrace);
- sanitizers (-fsanitize=address,undefined) to catch memory and undefined-behaviour bugs at the exact line;
- explaining your code line by line out loud (rubber-duck debugging), which finds a surprising number of bugs.`,
				UnderTheHood: `-g adds debug information (in a format called DWARF) that maps every
machine-code address back to a file, a line and variable names. That is
how gdb can say "main.c:12" when the CPU only knows an address.

A breakpoint is a trick: gdb replaces the first byte of the instruction
at that line with 0xCC, the int3 instruction. When the CPU reaches it,
it traps into the kernel, which stops your program and wakes gdb
(through the ptrace system call). gdb puts the original byte back,
shows you the state, and continues when you ask. A backtrace (bt) is
gdb walking the chain of stack frames, reading each saved return
address and looking up which function and line it belongs to.`,
				MentalModel: "The computer did exactly what you said; find where that differs from what you meant.",
				TryIt:       "Compile a program that crashes with -g, run it under gdb (gdb ./program, then run), and after the crash type bt to see exactly which function and line it died in.",
				Analogy: `A doctor diagnosing symptoms: gather evidence, form a hypothesis, test
one idea at a time, and do not prescribe before you have found the
cause. The compiler's warnings are the routine check-up that catches
problems before they hurt.`,
				Example: `In 1947, engineers on the Harvard Mark II computer found a real moth
stuck in a relay and taped it into their logbook as the "first actual
case of bug being found". Grace Hopper helped make the story famous.
Today, the Linux kernel and Chrome run sanitizers and fuzzers
continuously, around the clock, to find memory bugs before attackers
do.`,
				Exercises: trio(
					`Read this compiler output and explain it in one sentence each: "error: expected ';' before 'return'", "warning: unused variable 'total'", and "warning: format '%d' expects argument of type 'int', but argument 2 has type 'double'".`,
					"The first names the line after the missing semicolon; the third means printf will print garbage.",
					"Take a working program of yours and plant three bugs (a missing semicolon, an off-by-one in a loop, and a wrong printf format). Wait a day, then find and fix them using only the compiler's messages, gdb and printf.",
					"Fix the first compile error, rebuild, and repeat. For runtime bugs, set a breakpoint before the loop and step.",
					`A user reports "the program crashes when I type a very long name". Reproduce the crash in your own code with a 1,000-character input, find the exact line with the address sanitizer and gdb, fix it with a bounded read, and add a test that proves it stays fixed.`,
					`python3 -c "print('x' * 1000)" | ./program reproduces it. The fix is usually fgets with sizeof buf, plus a check for lines that did not fit.`,
				),
			},
		},
		Resources: []Resource{
			{"Course", "CS50: Introduction to Computer Science (Harvard, free)", "https://cs50.harvard.edu/x/", "Weeks 1–5 teach C, memory and pointers, with excellent lectures and problem sets."},
			{"Book", "Beej's Guide to C Programming (free)", "https://beej.us/guide/bgc/", "A friendly, complete, free tour of C for people who can already code a little. The PDF is in the Library [l]."},
			{"Book", "Modern C (free, Jens Gustedt)", "https://gustedt.gitlabpages.inria.fr/modern-c/", "A rigorous, up-to-date book on C as it is written today, including the C23 standard."},
			{"Tool", "Python Tutor in C mode (C Tutor)", "https://pythontutor.com/c.html", "Runs your C code step by step and draws memory, arrays and pointers as boxes and arrows."},
			{"Tool", "Compiler Explorer", "https://godbolt.org/", "Shows the machine code your C compiles into, line by line."},
			{"Site", "cppreference: C reference", "https://en.cppreference.com/w/c", "Precise, searchable documentation for the C language and its standard library."},
			{"Site", "Exercism: C track", "https://exercism.org/tracks/c", "Free C exercises with automatic tests and optional mentor feedback."},
			{"Site", "Advent of Code", "https://adventofcode.com/", "Yearly puzzles that are great for practising loops, arrays and parsing in C."},
		},
		Blueprints: []Blueprint{
			{"Personal Budget Tracker (in C)", "A command-line tool that records income and spending in a file, and prints monthly summaries by category, using structs, arrays and careful file I/O.",
				[]string{"Add and list transactions (a struct per entry)", "Save to and load from a text file with fopen and fgets", "Monthly totals by category", "Import a bank CSV, rejecting malformed lines", "Clean under -Wall -Wextra and -fsanitize=address,undefined"}},
			{"Text Adventure Game (in C)", "A small interactive story with rooms, items and choices, driven by loops, conditions, structs and arrays.",
				[]string{"Rooms as an array of structs with exits", "Move between rooms with safe input", "Pick up and use items", "Win and lose conditions", "Save and load the game"}},
			{"Quiz App (in C)", "A flashcard quiz that loads questions from a file, shuffles them, keeps score and remembers your weak spots.",
				[]string{"Load questions into an array of structs", "Ask and check answers (case-insensitive)", "Shuffle with a Fisher–Yates loop", "Repeat missed questions", "Store history between runs"}},
		},
		Quiz: []Question{
			{"What does the compiler do, and why can't you run hello.c directly?", "It translates C source text into machine code and produces an executable. The CPU runs only machine code, never source text."},
			{"What is 7 / 2 in C, and how do you get 3.5?", "3, because both are ints and integer division drops the fraction. Use 7.0 / 2 or (double)7 / 2."},
			{"Why does scanf need &x instead of x?", "C passes arguments by value; scanf needs x's address (a pointer) so that it can write the value into x."},
			{"How many bytes does the string \"cat\" take in C, and why?", "4: three characters plus the terminating '\\0' that marks the end of the string."},
			{"What is undefined behaviour, and name two causes.", "An operation the C standard gives no meaning to, so anything may happen. For example: reading past the end of an array, dereferencing NULL, signed overflow, using an uninitialised variable."},
			{"What is the difference between s.age and p->age?", "s.age reaches a field of a struct value; p->age reaches a field through a pointer to a struct, and means (*p).age."},
			{"Why should you not write p = realloc(p, n)?", "If realloc fails it returns NULL and leaves the old block untouched; assigning NULL to p loses the only pointer to that block, which leaks it. Use a temporary and assign only on success."},
			{"Why check what fclose returns when you have been writing a file?", "Buffered data is only handed to the operating system when the stream is flushed, so a full disk or an I/O error may only show up then, as fclose returning EOF."},
			{"What does an include guard prevent, and what does \"undefined reference\" at link time usually mean?", "It stops a header's contents being processed twice in one file. An undefined reference means the linker found no definition for a name: usually a .c file (or a library) is missing from the command."},
			{"Which compiler flags should you use while learning, and why?", "-Wall -Wextra for warnings, -g for debugging information, and -fsanitize=address,undefined to catch memory and undefined-behaviour bugs at the exact line."},
		},
	},

	// -----------------------------------------------------------------
	2: {
		Overview: `A data structure is a way of organising data; an algorithm is a recipe
for working with it. Together they decide whether your program answers
in a millisecond or in a week. This stage gives you the classic toolbox,
in C, so you see where every byte lives: Big-O to reason about speed
before you write a line, recursion and divide and conquer, arrays and
lists, searching, hash tables, trees and heaps, graphs, and the two big
problem-solving strategies, greedy algorithms and dynamic programming.
Each lesson has a checked exercise: the academy compiles your solution
and times it on inputs big enough that only the efficient idea passes.`,
		Outcomes: []string{
			"Predict how a program will slow down as its data grows, and measure it",
			"Write recursive solutions, and know when a loop is safer",
			"Choose the right data structure for a problem, and explain what it costs in memory",
			"Solve classic problems (searching, sorting, routing, scheduling) and explain why your solution is efficient",
			"Recognise when a greedy choice is safe and when a problem needs dynamic programming",
		},
		Glossary: []Term{
			{"Algorithm", "A precise, step-by-step recipe for solving a problem."},
			{"Data structure", "A way of organising data in memory so that certain operations are fast."},
			{"Big-O", "Notation for how the work grows with input size, such as O(n) or O(n²)."},
			{"Array", "Items stored side by side in memory, accessed by position."},
			{"Pointer", "A value that holds the memory address of something else."},
			{"Hash function", "A function that turns a key into a number, used to pick a storage slot."},
			{"Tree", "A hierarchy of nodes where each node has children and there are no loops."},
			{"Graph", "A set of nodes connected by edges, such as cities and roads or people and friendships."},
			{"Recursion", "Solving a problem by having a function call itself on smaller versions of it."},
			{"Base case", "The smallest version of a recursive problem, answered directly so the calls stop."},
			{"Divide and conquer", "Split a problem into smaller parts, solve each, and combine the answers (merge sort, binary search)."},
			{"Binary search", "Finding an item in sorted data by halving the range that can hold it at every step."},
			{"Invariant", "A statement that stays true at every step of a loop, used to reason that the loop is correct."},
			{"Greedy algorithm", "An algorithm that makes the choice that looks best right now and never revisits it."},
			{"Dynamic programming", "Solving a problem by solving each smaller subproblem once, storing the answers and reusing them."},
			{"Memoization", "Caching a function's results so that a repeated call returns the stored answer instead of recomputing it."},
		},
		Concepts: []Concept{
			{
				Name:    "Big-O: Measuring Growth",
				Summary: "How does the work grow when the input grows?",
				Body: `Instead of timing code on one machine, describe how its cost grows with
the input size n, ignoring constant factors:
- O(1), constant: array lookup by index.
- O(log n), halving each step: binary search.
- O(n), looking at everything once.
- O(n log n): efficient sorting.
- O(n²), comparing every pair: fine for 1,000 items, painful for 1,000,000.
- O(2ⁿ), trying every subset: hopeless beyond about 50 items.

To find the Big-O of code, count how many times the innermost line runs
as a function of n. Two nested loops over the same data are usually n²;
a loop that halves its range each time is log n; a loop that does
O(log n) work n times is n log n. Keep only the fastest-growing term and
drop constants: 3n² + 10n + 7 is O(n²).

Big-O usually describes the worst case, and it hides constant factors.
Those constants can be large: the same O(n) loop can be a hundred times
slower over data scattered in memory than over an array (see below).
Memory use (space complexity) matters too: a solution that needs an
n × n table is impossible for n = 100,000 whatever its speed.`,
				Diagram: ` n            log n        n       n log n            n²
 10               3       10            33          100
 1,000           10    1,000        10,000    1,000,000
 1,000,000       20    10⁶       2 × 10⁷           10¹²  (days of work)`,
				UnderTheHood: `Measured with gcc -O2 on one test machine (a 2.8 GHz Xeon), searching
an array of distinct numbers for a duplicate:

  n          nested loops (n²)   qsort, then scan neighbours (n log n)
  10,000          15 ms                0.2 ms
  100,000      1,900 ms                2.7 ms
  1,000,000    not run (about 3 min)   31 ms

Ten times more data made the nested loops a hundred times slower, and
the sorted version only eleven to fourteen times slower. That is n²
against n log n.

Constants matter as soon as the growth rates are equal. Summing 10
million ints is O(n) whichever way they are stored, but on the same
machine it took about 6 ms in an array, about 50 ms following a linked
list whose nodes sat in allocation order, and about 2.5 s following the
same list after its links were shuffled. Every hop to a scattered node
is a cache miss that waits on RAM (The Memory Hierarchy & Locality,
Stage 5), while the array streams through memory the CPU fetches ahead.`,
				MentalModel: "Constants matter in practice, but the growth rate decides who wins at scale.",
				TryIt:       "Write a nested-loop duplicate finder and a qsort-based one in C, time both with clock_gettime on 10,000 and 100,000 numbers, and check that your times grow like the table above.",
				Analogy: `Finding a word in a paper dictionary. Reading every page from the start
is O(n). Opening it in the middle, deciding which half the word is in
and halving again is O(log n): about 17 flips for 100,000 words.`,
				Example: `In 2021 a player discovered that GTA Online's loading screen spent
minutes parsing a 10 MB JSON file with an accidentally quadratic
algorithm. His fix cut loading times by about 70%, and Rockstar shipped
it and paid him a bounty.`,
				Exercises: trio(
					"Give the Big-O of: printing every item; printing every pair of items; finding a word by repeatedly halving a sorted list; reading the first item; for (i = 1; i < n; i *= 2) for (j = 0; j < n; j++) work();",
					"n, n², log n, 1, and n log n: the outer loop runs about log₂ n times, and each pass runs the inner loop n times.",
					"Write two versions of has_duplicates(const int *a, size_t n) in C, one with nested loops and one that copies the array, sorts it with qsort and compares neighbours. Time both with clock_gettime(CLOCK_MONOTONIC) at n = 1,000, 10,000 and 100,000, and put the results in a table.",
					"Fill the array with distinct numbers (the worst case: no early exit). A qsort comparison function must return negative, zero or positive: (x > y) - (x < y) avoids the overflow of x - y.",
					"A school emails every parent by, for each of 50,000 students, scanning the whole parent list for a match. It takes hours. State the Big-O and rewrite it in C to be fast.",
					"It is O(n²). Sort both lists by student ID and walk them together like a zip (O(n log n)), or put the parents in a hash table keyed by student ID (O(n) on average).",
				),
			},
			{
				Name:    "Recursion & Divide and Conquer",
				Summary: "Solve a problem by solving smaller copies of it, and know what every call costs.",
				Body: `A recursive function calls itself on a smaller version of its own
problem. It needs two parts: a base case, answered directly, and a
recursive case that moves towards it.

  long sum_to(long n) {
      if (n == 0) return 0;          // base case: nothing left to add
      return n + sum_to(n - 1);      // a smaller copy of the same problem
  }

To trust a recursive function, do not trace every call. Check that the
base case is right, and that if the smaller call returns the right
answer, this call does too. That is proof by induction (Stage 3).

Divide and conquer is recursion that splits the work: divide the input
into parts, solve each part recursively, and combine the answers. Merge
sort splits an array in half, sorts each half and merges them: log₂ n
levels of splitting, n work per level, O(n log n) in total. Binary
search keeps only the half that can hold the answer. Fast power computes
xⁿ from x^(n/2) squared, in at most about 2 log₂ n multiplications
instead of n.

Every call takes a frame on the stack, which is small (8 MiB by default
on Linux, often 1 MiB on Windows). Recursion as deep as the input is long (a
list of a million nodes, a path through a million maze cells) runs out
of stack and crashes; recursion as deep as log n (halving) never will.
Any recursion can be rewritten as a loop with an explicit stack of your
own, on the heap.`,
				Diagram: `  sum_to(3)                          the stack while sum_to(0) runs
  └ 3 + sum_to(2)                    ┌──────────────────┐
        └ 2 + sum_to(1)              │ main             │
              └ 1 + sum_to(0)        │ sum_to  n = 3    │
                    └ returns 0      │ sum_to  n = 2    │
  then 1, 3 and 6 come back up       │ sum_to  n = 1    │
                                     │ sum_to  n = 0    │ ◀ top
                                     └──────────────────┘`,
				UnderTheHood: `gcc -O0 compiles sum_to into a real call that saves its argument in a
16-byte stack frame:

  sum_to:
      push  rbp
      mov   rbp, rsp
      sub   rsp, 16
      mov   QWORD PTR -8[rbp], rdi      ; keep n for after the call
      ...
      call  sum_to
      mov   rdx, QWORD PTR -8[rbp]      ; n again
      add   rax, rdx                    ; n + sum_to(n - 1)

With the return address and the saved rbp, each call costs 32 bytes, so
the default 8 MiB stack holds 262,144 of them. On a test machine,
sum_to(260000) returned and sum_to(262000) crashed with a segmentation
fault: the stack overflowed into the guard page below it.

gcc -O2 is cleverer: it notices that the pending additions can be done
on the way down, and turns the recursion into a loop with no calls at
all, so sum_to(10000000) works. Do not rely on it: at -O0, under a
sanitizer, or in a slightly different function, the calls are real.`,
				MentalModel: "Trust the smaller call, get the base case right, and watch the depth.",
				TryIt:       "Compile sum_to with -O0, call it with 100000, 250000 and 300000, and find where it crashes. Then compile with -O2 and run gcc -O2 -S to see the loop gcc wrote instead.",
				Analogy: `In a dark cinema you want your row number. You ask the person in front,
who asks the person in front of them, and so on until someone in the
first row says "1" (the base case). Each answer comes back adding one.
Nobody needs to see the whole cinema; everyone trusts the answer from
the row in front.`,
				Example: `Every tool that walks a folder tree (find, du, a file explorer's search)
is recursive, because folders contain folders. Compilers parse nested
expressions recursively (Parsing: Recursive Descent & Pratt, Stage 5),
and the site Stack Overflow is named after what happens when recursion
never reaches its base case.`,
				Exercises: trio(
					"Trace sum_to(4) on paper: the calls going down and the values coming back up. Then delete the base case: what happens when you run it?",
					"4 + (3 + (2 + (1 + 0))) = 10. Without the base case the calls never stop: each one takes another 32 bytes of stack until the stack runs out and the program crashes with a segmentation fault.",
					"Write long long power(long long x, unsigned n) by halving (xⁿ = (x^(n/2))², times x when n is odd), and count its multiplications for n = 1,000 against a simple loop. Then write void hanoi(int n, char from, char to, char via), which prints every move of the Towers of Hanoi, and count the moves for n = 1 to 20.",
					"power needs about 2 × log₂ n multiplications (about 20 for n = 1,000). Hanoi: move n − 1 disks out of the way, move the biggest, move the n − 1 back on top; the count is 2ⁿ − 1.",
					"Write a du-like program: given a folder, print the total size of every file under it, recursing into subfolders with opendir and readdir (POSIX). Then make it survive a folder nested 10,000 levels deep by replacing the recursion with an explicit stack of paths.",
					"Skip \".\" and \"..\". Use lstat, not stat, so a symbolic link that points back up the tree cannot loop forever. The explicit stack is an array of malloc'd path strings that you grow with realloc.",
				),
			},
			{
				Name:    "Arrays, Linked Lists, Stacks & Queues",
				Summary: "The basic ways to line data up, and their trade-offs.",
				Body: `An array stores items side by side in memory. Access by index is
instant and cache-friendly, but inserting in the middle shifts
everything after it (O(n)). A growable array (your intvec from Stage 1,
Python lists, Go slices) doubles its capacity when full, so appending is
O(1) on average.

A linked list chains nodes with pointers:

  struct node { int value; struct node *next; };

Inserting is O(1) once you hold the node, but reaching the i-th item is
O(n), and the nodes are scattered in memory, which is slow for the
cache. A stack is last-in, first-out (push and pop), used for undo,
function calls and matching brackets. A queue is first-in, first-out,
used for print jobs, task scheduling and breadth-first search.

In C both are easiest on top of an array. A stack is an array and a top
index. A queue is an array used as a ring: a head index where items
leave and a tail index where they arrive, both wrapping around to 0 at
the end, so nothing is ever shifted.`,
				Diagram: `array:   [ 7 | 3 | 9 | 4 ]              index i → jump straight there
linked:  7 → 3 → 9 → 4 → ∅              walk from the head
stack:   push ▶ [ 4 | 9 | 3 | 7 ]        pop takes the 4  (last in, first out)
queue:   in ▶ [ 4 | 9 | 3 | 7 ] ▶ out    7 leaves first   (first in, first out)`,
				UnderTheHood: `On x86-64, struct node { int value; struct node *next; } is 16 bytes:
4 for the int, 4 of padding so the pointer is 8-aligned, and 8 for the
pointer. Each malloc(16) costs more than 16. In glibc, three nodes
allocated one after another sat 32 bytes apart, because every block
carries an 8-byte size header and is rounded up to a multiple of 16
(malloc_usable_size reported 24 usable bytes). A million-node list of
ints therefore takes about 32 MB; an array of the same ints takes 4 MB.

Speed differs even more. Summing 10 million ints with gcc -O2 on a test
machine took about 6 ms in an array, about 50 ms walking a list whose
nodes were in allocation order, and about 2.5 seconds walking a list
whose links had been shuffled, the state a list reaches after many
inserts and deletes. The array is read in order, so the CPU fetches the
next cache lines before they are needed: well under a nanosecond per
int. Each shuffled node is a cache miss, and often a page-table miss as
well, that waits on RAM: about 250 ns per node here.`,
				MentalModel: "Stack = pile of undo steps; queue = waiting line; array = numbered seats.",
				TryIt:       `Use a stack (an array and a top index) to check whether the brackets in "(a[b]{c})" are balanced, and print where the first mismatch is in "(a[b)c]".`,
				Analogy: `A stack of plates (you always take from the top), a queue at the post
office (first come, first served), numbered cinema seats (an array: walk
straight to seat 14) and a treasure hunt where each clue points to the
next one (a linked list).`,
				Example: `Ctrl+Z in every editor pops a stack of changes, and the browser's back
button pops a stack of pages. Print spoolers, customer-support ticket
systems and message queues such as RabbitMQ and Kafka are queues. The
Linux kernel links thousands of its structures (processes, open files)
with its own doubly linked list, struct list_head.`,
				Exercises: trio(
					"Push 1, 2 and 3 onto a stack, pop once, push 4, then pop twice. What is left? Repeat with a queue, using enqueue and dequeue.",
					"The stack is left holding 1; the queue is left holding 4. Draw the boxes on paper after every step.",
					"In C, write a stack of ints (push, pop, peek, empty) on a growable array, and a queue of ints as a ring buffer (enqueue, dequeue) that doubles its capacity when full. Test both with a million operations under -fsanitize=address.",
					"For the ring: head is where items leave, tail where they arrive, and index = (head + i) % capacity. When it is full, copy the items in order into a new array twice as big and reset head to 0.",
					"Build undo and redo for a tiny line editor in C: commands to append a line, delete the last line and print the text; every edit can be undone, and undone edits can be redone until a new edit is made.",
					"Two stacks of edits: undo pops from one and pushes onto the other; redo does the reverse; a new edit clears the redo stack. An edit records enough to reverse it (for a deletion, the deleted line).",
				),
			},
			{
				Name:    "Searching: Linear & Binary Search",
				Summary: "Looking at everything takes n steps; halving a sorted range takes log₂ n.",
				Body: `Linear search checks every item in turn: O(n), and it works on any
data. If the data is sorted, binary search does far better. Keep a range
[low, high] that must contain the target if it is present, look at the
middle, and throw away the half that cannot hold it:

  int find(const int *a, int n, int x) {
      int low = 0, high = n - 1;        // x, if present, is in a[low..high]
      while (low <= high) {
          int mid = low + (high - low) / 2;
          if (a[mid] == x) return mid;
          if (a[mid] < x) low = mid + 1;   // x is right of mid
          else            high = mid - 1;  // x is left of mid
      }
      return -1;                        // the range is empty: not there
  }

Each step halves the range, so a million sorted items need at most 20
comparisons and a billion need 30. The hard part is the edges: is high
inclusive, does the loop stop at low < high or low <= high, and does
each branch shrink the range? State the invariant in a comment and check
every branch against it.

The same idea finds boundaries, not just items. lower_bound returns the
first position whose value is at least x, which also tells you where to
insert x and, with duplicates, where their run starts. And you can
binary search on the answer: the smallest capacity that passes a test,
the first version that has a bug, the square root of a number, as long
as the test is false up to some point and true after it.`,
				Diagram: `find 23 in  2   5   8   12  16  23  38  56  72  91
index       0   1   2   3   4   5   6   7   8   9
step 1: low 0, high 9, mid 4 → 16 < 23, so low = 5
step 2: low 5, high 9, mid 7 → 56 > 23, so high = 6
step 3: low 5, high 6, mid 5 → 23 found at index 5`,
				UnderTheHood: `The textbook midpoint, (low + high) / 2, overflows when low + high
exceeds INT_MAX. With low = 1,500,000,000 and high = 2,000,000,000,
gcc's undefined-behaviour sanitizer stops the program with:

  runtime error: signed integer overflow: 1500000000 + 2000000000
  cannot be represented in type 'int'

and without it the sum wraps to a negative number, so mid is a negative
index. low + (high - low) / 2 never overflows. This exact bug sat in
Java's standard binary search for nine years before Joshua Bloch
reported it in 2006: it only shows on arrays of over a billion items,
which nobody tested until such arrays became common.

The C library has a generic binary search, bsearch from stdlib.h, which
takes the same kind of comparison function as qsort. It returns a
pointer to some matching element, not necessarily the first one, which
is why you write lower_bound yourself when duplicates matter.`,
				MentalModel: "Keep a range that must hold the answer, and halve it until it is one item or empty.",
				TryIt:       "Write find and a linear search, and run both on a sorted array of 10 million ints for 1,000 random targets. Then test find on an empty array, a one-item array, and targets smaller and larger than every item.",
				Analogy: `The higher-or-lower guessing game. Someone picks a number from 1 to 100;
you guess 50, hear "lower", guess 25, and so on. Seven guesses always
suffice, because 2⁷ = 128 is more than 100.`,
				Example: `git bisect binary-searches your history: you mark one commit good and a
later one bad, and it checks out the middle commit for you to test,
finding the commit that introduced a bug among 1,000 in about 10 steps.
Database indexes (B-Trees, Stage 7) and the lookup tables inside
compilers and network routers are binary searches over sorted data.`,
				Exercises: trio(
					"How many higher-or-lower guesses do you need, at most, for a number from 1 to 1,000? From 1 to 1,000,000? From 1 to 4 billion?",
					"10, 20 and 32: the smallest k with 2ᵏ at least the count, since each guess halves what is left.",
					"Write int lower_bound(const int *a, int n, int x), the first index whose value is at least x (n if there is none), and test it against a linear scan on 10,000 random sorted arrays, including empty arrays, duplicates, and x below or above every value.",
					"Keep the invariant: everything before low is less than x, everything from high on is at least x. Loop while low < high with high = n at first; at the end low == high is the answer.",
					"Versions 1 to n of a program: early ones are good, and from some version on every one is bad. Given a function int is_bad(int version), find the first bad version with at most about log₂ n calls, and count your calls for n = 1,000,000. Then use git bisect on a real repository.",
					"Binary search on the answer: if is_bad(mid), the first bad one is mid or earlier (high = mid); otherwise it is later (low = mid + 1).",
				),
			},
			{
				Name:    "Hash Tables",
				Summary: "Near-instant lookup by key: the most useful data structure in practice.",
				Body: `A hash function turns a key (such as "alice@mail.com") into a number,
which picks a bucket in an array. Lookup, insert and delete are O(1) on
average. Two keys can land in the same bucket (a collision). Tables
handle this by chaining (a small list per bucket) or by open addressing
(probing for the next free slot). When the table gets too full, it grows
and rehashes everything.

A simple, good hash for strings is FNV-1a:

  #include <stdint.h>

  uint64_t hash(const char *s) {
      uint64_t h = 14695981039346656037ULL;   // 64-bit FNV offset basis
      for (; *s != '\0'; s++) {
          h ^= (unsigned char)*s;
          h *= 1099511628211ULL;               // 64-bit FNV prime
      }
      return h;
  }

and the bucket is h % nbuckets. (uint64_t, not unsigned long, which is
only 32 bits on Windows.) Keep the load (items ÷ buckets) below
about 3/4, doubling the table when it passes. Because the table doubles,
each item is rehashed only a constant number of times on average, so
inserting stays O(1).

A good hash spreads keys evenly. A predictable one can be attacked with
keys that all collide, degrading every operation to O(n). This is
called hash flooding.`,
				Diagram: `"alice" ─hash─▶ 2      bucket 0 [                ]
"bob"   ─hash─▶ 5      bucket 2 [ alice → carol  ]  ◀ collision, chained
"carol" ─hash─▶ 2      bucket 5 [ bob            ]`,
				UnderTheHood: `Why so many hash tables use a power of two for their size: gcc -O2
turns h % 8 on an unsigned h into a single AND,

  and   eax, 7

but h % n, with n only known at run time, needs a division instruction,

  xor   edx, edx
  div   esi              ; the remainder lands in edx

and a division costs many times more than an AND (tens of cycles on
many CPUs), often more than the hash itself. With a power-of-two size, the bucket is h & (size - 1). The
catch: that keeps only the low bits of h, so the hash must mix its high
bits down into the low ones, as FNV-1a's multiply does.

With signed ints, h % 8 takes six instructions instead of one, because
C requires -9 % 8 to be -1, not 7: one more reason to make hash values
unsigned.`,
				MentalModel: "Compute where it lives instead of searching for it.",
				TryIt:       "Implement a hash map with chaining and resizing in C, and print the longest chain and the average chain length after inserting 100,000 words.",
				Analogy: `A coat check: your ticket number tells the attendant exactly which hook
your coat is on. Nobody searches through every coat.`,
				Example: `Python dictionaries, JavaScript objects, database indexes on IDs, DNS
caches and Git (which names every file version by its hash) all rely on
hashing. In 2011 researchers showed they could freeze many web
frameworks by sending form fields built to collide, and languages then
started randomising their hash functions.`,
				Exercises: trio(
					`Using hash(key) = (sum of letter positions, a = 1) mod 5, place "cat", "act" and "dog" into 5 buckets. What happens with "cat" and "act"?`,
					"Anagrams have the same letter sum, so they collide: cat and act both land in bucket 4 (24 mod 5), dog in bucket 1 (26 mod 5). A good hash depends on the order of the letters too.",
					"Write a word-frequency counter in C on your own hash table (chaining, FNV-1a, doubling at 3/4 full): read a book from Project Gutenberg, lowercase each word, and print the 20 most frequent words.",
					"Copy each key into the table with malloc; never keep a pointer into the line buffer, which the next fgets overwrites. To find the top 20, copy the entries into an array and qsort it by count.",
					"An online shop must find duplicate sign-ups whose emails differ only in capitalisation or surrounding spaces. Find all duplicates among 1 million emails in roughly O(n), in C.",
					"Normalise each email (trim the spaces, lowercase it) into a new string, and look that up in your hash table: if it is already there, it is a duplicate.",
				),
			},
			{
				Name:    "Trees, Heaps & Sorting",
				Summary: "Keeping data ordered so that searching, ranking and sorting are fast.",
				Body: `A binary search tree keeps smaller keys on the left and larger keys on
the right, so search, insert and delete take O(log n), provided the tree
stays balanced (red-black and AVL trees rebalance themselves). A heap
keeps the smallest (or largest) item on top: O(1) to peek and O(log n)
to insert or remove. Heaps are the basis of priority queues.

A heap needs no pointers at all. Store it in an array: the children of
the item at index i are at 2i + 1 and 2i + 2, and its parent is at
(i − 1) / 2. To insert, append at the end and swap the item up while it
is smaller than its parent; to remove the top, move the last item to
the top and swap it down with its smaller child.

Simple sorts (insertion, selection) are O(n²), though insertion sort is
the fastest on tiny or nearly sorted arrays. Merge sort and heap sort
guarantee O(n log n). Quicksort is O(n log n) on average and very fast
in practice, but O(n²) on unlucky input unless it picks its pivot well.
No comparison-based sort can beat O(n log n) in the worst case. A sort
is stable when equal items keep their original order, which matters
when you sort by one field after another.`,
				Diagram: `          8            binary search tree: left < node < right
        ╱   ╲
       3     10        search for 6: 8 → 3 → 6  (3 steps, not 7)
      ╱ ╲      ╲
     1   6      14

  min-heap as an array:  [ 1 | 3 | 2 | 7 | 4 ]   children of i: 2i+1, 2i+2
                           0   1   2   3   4`,
				UnderTheHood: `"Provided the tree stays balanced" hides a trap. On a test machine,
inserting the keys 0 to 99,999 into a plain binary search tree in random
order built a tree 39 levels deep in 25 ms. Inserting the same keys in
sorted order built a "tree" 100,000 levels deep, a linked list in
disguise, and took 9.9 seconds: every insert walked the whole list.
Sorted input is common (IDs, timestamps), which is why real systems use
self-balancing trees, or B-Trees on disk (Stage 7).

The heap's index arithmetic is cheap: 2 * i + 1 compiles to a single
instruction, lea rax, 1[rdi+rdi]. A heap of a million items is only 20 levels deep, so a push
or pop does at most about 20 swaps, and heap sort needs no extra memory
at all: it sorts inside the array it is given.`,
				MentalModel: "Sorted data turns searching into halving.",
				TryIt:       "Insert 100,000 keys into a plain binary search tree in C, once in random order and once sorted, and print each tree's height and build time.",
				Analogy: `A tournament bracket is a tree. Hospital emergency triage is a heap: the
most urgent patient is seen next, not whoever arrived first. A sorted
library shelf lets you jump straight to the right spot.`,
				Example: `The folders on your computer form a tree. Route planners use heaps
inside Dijkstra's algorithm. Python's built-in sort (Timsort) is a hybrid
of merge sort and insertion sort, tuned for real data that is often
partly sorted already. Operating-system schedulers and timers keep
their next jobs in priority queues.`,
				Exercises: trio(
					"Insert 5, 2, 8, 1, 9 and 3 into an empty binary search tree and draw it. Then insert 1, 2, 3, 4, 5 into a fresh tree. What shape do you get, and why is it bad?",
					`Sorted input makes a "stick": a linked list in disguise, where every search walks every node.`,
					"In C, implement a min-heap of ints in an array (push, pop, and a sift-down), and use it to write heap_sort. Test it against qsort on 1,000 random arrays, including empty and one-item arrays.",
					"Children of i are at 2i + 1 and 2i + 2. Heap sort: build a max-heap in place, then repeatedly swap the top to the end and sift the new top down over the shrinking heap.",
					"An emergency-room program in C must always show the most urgent patient next (severity 1 to 5, 1 most urgent), and among equal severities, whoever arrived first. Implement it with a heap and simulate 20 arrivals and treatments.",
					"A heap of structs ordered by (severity, arrival number): compare severity first, then the arrival counter, which also makes the order stable.",
				),
			},
			{
				Name:    "Graphs, BFS/DFS & Shortest Paths",
				Summary: "Networks of things and connections, and how to explore them.",
				Body: `A graph is a set of nodes connected by edges, which can be directed
(one-way) and weighted (distances or costs). It is stored as an
adjacency list (each node's neighbours) or as a matrix.

Breadth-first search (BFS) explores in rings using a queue and finds the
path with the fewest hops. Depth-first search (DFS) dives deep using a
stack or recursion; it detects cycles and produces topological orders
(orderings of tasks with dependencies). Dijkstra's algorithm finds the
shortest weighted paths using a priority queue, and A* adds a heuristic
that aims towards the goal. Many problems become easy once you see the
graph: dependencies, maps, social networks, the web, and grids, where
each cell is a node joined to its four neighbours.

BFS in C needs little more than arrays:

  dist[start] = 0;  queue[tail++] = start;
  while (head < tail) {
      int u = queue[head++];
      for (each neighbour v of u)
          if (dist[v] == -1) {          // not seen yet
              dist[v] = dist[u] + 1;
              queue[tail++] = v;
          }
  }

Each node enters the queue once and each edge is looked at once (twice
if undirected), so BFS is O(V + E).`,
				Diagram: `A ── 4 ── B        BFS from A (by hops):   A │ B C │ D
│         │
1         1        Dijkstra from A (by weight):
│         │          A = 0, C = 1, D = 3, B = 4
C ── 2 ── D`,
				UnderTheHood: `The choice of representation decides what fits in memory. An adjacency
matrix needs V² cells whatever the edges: for a million nodes that is
10¹² cells, 125 GB even at one bit each. Most real graphs are sparse (a
road junction joins three or four roads), so an adjacency list, V + E
entries, holds the same million-node graph in a few tens of megabytes.

A compact C layout is two arrays: all the edges sorted by their source
node, and an offset array saying where each node's edges start (the
"compressed sparse row" layout). Walking a node's neighbours is then a
scan of consecutive memory, which is fast for the cache, and BFS needs
only a dist array and one queue array of V ints, no malloc per node.`,
				MentalModel: "If things connect to things, draw the graph first.",
				TryIt:       "Solve a maze stored in a text file with BFS in C and print the shortest route, by keeping each cell's parent and walking back from the end.",
				Analogy: `A city map: intersections are nodes, roads are edges and travel times
are weights. BFS is a rumour spreading through friends of friends. DFS is
exploring a cave by always taking the next unexplored tunnel and
backtracking at dead ends.`,
				Example: `Google Maps finds routes with shortest-path algorithms on the road graph.
LinkedIn's "2nd-degree connection" is BFS. Package managers such as npm
and pip use topological sorting to install dependencies in the right
order, and make uses it to decide what to rebuild. Google's original
PageRank treated the web as a graph of links.`,
				Exercises: trio(
					"Draw a graph of five friends and who knows whom. Who is two hops from you? Is there a cycle?",
					"Do BFS by hand: ring 1, then ring 2. A cycle is a path that comes back to where it started without reusing an edge.",
					"In C, read a graph as lines of \"a b w\" (an edge from a to b with weight w), store it as adjacency lists, and implement BFS, DFS and Dijkstra. Test them on the small graph in the diagram.",
					"BFS uses a queue, DFS a stack (or recursion), and Dijkstra a min-heap of (distance, node) pairs; skip a popped pair whose distance is larger than the best already known.",
					"Given a list of university courses and their prerequisites, print an order in which a student can take every course, or report that the prerequisites form an impossible cycle.",
					"Topological sort (Kahn's algorithm): count each course's unmet prerequisites, queue the courses with none, and as you take each one, decrement the courses that depend on it. If some are never queued, there is a cycle.",
				),
			},
			{
				Name:    "Greedy Algorithms",
				Summary: "Take the choice that looks best right now, and never look back: fast, but only correct when you can prove it.",
				Body: `A greedy algorithm builds its answer one choice at a time, always taking
the option that looks best at that moment, and never undoing a choice.
It is usually simple and fast (often a sort followed by one pass), but
it is only correct for problems where the local best choice is always
part of some best overall answer.

Interval scheduling is the classic success. Given meetings with start
and end times, to fit the most into one room, sort them by end time and
take each meeting that starts after the last one you took ends. Taking
the meeting that ends first always leaves the most room for the rest,
and an exchange argument proves it: any best schedule can swap its first
meeting for the earliest-ending one without losing anything.

Other greedy algorithms you will meet: Dijkstra's shortest paths (always
settle the closest unsettled node), Kruskal's and Prim's minimum
spanning trees (always add the cheapest safe edge), and Huffman coding
(always merge the two rarest symbols).

Greedy fails when an early choice blocks a better one later. Making 6
from coins of 1, 3 and 4, greedy takes 4, then 1, then 1: three coins,
where 3 + 3 needs two. When you cannot prove the greedy choice safe, try
a counterexample on small inputs, and reach for dynamic programming.`,
				Diagram: `time     0  1  2  3  4  5  6  7  8  9  10
A        ├────────┤                       ✓ taken
B              ├────────┤                 ✗ overlaps A
C                    ├────────┤           ✓ taken
D                          ├────────┤     ✗ overlaps C
E                                ├─────┤  ✓ taken
sorted by end time: A, C and E fit, the most possible`,
				UnderTheHood: `Huffman coding is the greedy algorithm inside zip, gzip and PNG (their
DEFLATE format uses it). Count how often each symbol appears, then
repeatedly merge the two rarest symbols or groups into one, until one
tree remains; each symbol's code is its path from the root, so frequent
symbols get short codes. For "abracadabra" (a ×5, b ×2, r ×2, c ×1,
d ×1) a program doing exactly that gives:

  8-bit ASCII              11 × 8 = 88 bits
  fixed 3-bit codes        11 × 3 = 33 bits   (5 symbols need 3 bits)
  Huffman                  23 bits

The total is the sum of the weights of every merge (2 + 4 + 6 + 11 = 23),
and no prefix code can do better: merging the two rarest first is
provably safe, which is what makes this greedy algorithm correct. The
merging uses a min-heap, so building the code is O(k log k) for k
symbols.`,
				MentalModel: "Greedy is a hypothesis: prove the choice is safe, or find the counterexample.",
				TryIt:       "Write the coin-change greedy in C and run it for every amount from 1 to 100 with coins {1, 3, 4} and with coins {1, 5, 10, 25}. For which coin set does it ever use more coins than necessary? Check against a brute force.",
				Analogy: `A cashier giving change with the largest coin that fits, again and
again. With euro or dollar coins this always gives the fewest coins; with
a strange set of coins it can fail, and the cashier would never know.`,
				Example: `Every zip file, PNG image and gzip-compressed web page uses Huffman
codes. Network routers run Dijkstra (as in the OSPF protocol) to choose
paths, and utilities lay cable or pipe networks as minimum spanning
trees. Calendar apps that pack meetings into the fewest rooms use the
interval-scheduling idea.`,
				Exercises: trio(
					"Make 30 cents with coins of 25, 10 and 1 greedily. How many coins does greedy use, and what is the best you can do?",
					"Greedy takes 25 and then five 1s: six coins. Three 10s make 30 with three coins, so greedy is not optimal for this coin set.",
					"In C, read n meetings (start and end times), sort them with qsort by end time and choose the most non-overlapping meetings greedily. Then, for n up to 15, compare your answer against a brute force that tries every subset, on 1,000 random tests.",
					"A meeting may start exactly when the previous one ends. The brute force loops over every bitmask from 0 to 2ⁿ − 1 and checks that the chosen meetings do not overlap.",
					"Build a Huffman code in C: count the byte frequencies of a file, build the tree with a min-heap, print each byte's code, and compute the compressed size in bits. Compare it with the size of the same file after gzip -9.",
					"Each heap entry is a node with a weight; pop two, push a new node whose weight is their sum. The compressed size is the sum, over bytes, of frequency × code length. gzip also replaces repeated strings (LZ77), so it usually beats Huffman alone.",
				),
			},
			{
				Name:    "Dynamic Programming",
				Summary: "Solve each smaller subproblem once, store the answer, and build the big answer from them.",
				Body: `Some recursive solutions solve the same subproblem again and again.
The Fibonacci recursion fib(n) = fib(n − 1) + fib(n − 2) computes fib(38)
twice, fib(37) three times, and so on: exponential work for a linear
amount of information. Dynamic programming solves each subproblem once
and stores the answer. Top-down, that is memoization: keep the recursion
and cache each result. Bottom-up, it is tabulation: fill a table from
the smallest subproblems up, in a loop.

It applies when a problem has optimal substructure (the best answer is
built from best answers to subproblems) and overlapping subproblems.
The recipe:
- define the state: what does best[i] (or best[i][j]) mean, in words;
- write the recurrence that builds it from smaller states;
- set the base cases;
- fill the table in an order where every state needed is already done.

Fewest coins to make an amount, with any coin set:

  best[0] = 0;
  for (int v = 1; v <= amount; v++) {
      best[v] = INF;                       // INF: not possible (yet)
      for (int i = 0; i < ncoins; i++)
          if (coin[i] <= v && best[v - coin[i]] + 1 < best[v])
              best[v] = best[v - coin[i]] + 1;
  }

For coins {1, 3, 4} this finds 3 + 3 for 6, which greedy missed. Edit
distance (the fewest insertions, deletions and substitutions turning one
word into another), the longest common subsequence and the knapsack
problem are all filled tables like this.`,
				Diagram: `edit distance from "kitten" to "sitting": each cell is the distance
between the first i letters of one and the first j of the other

         ε  s  i  t  t  i  n  g
     ε   0  1  2  3  4  5  6  7
     k   1  1  2  3  4  5  6  7
     i   2  2  1  2  3  4  5  6
     t   3  3  2  1  2  3  4  5
     t   4  4  3  2  1  2  3  4
     e   5  5  4  3  2  2  3  4
     n   6  6  5  4  3  3  2  3      → answer 3 (bottom right)

cell = min(above + 1, left + 1, diagonal + (letters differ ? 1 : 0))`,
				UnderTheHood: `Counting the calls, compiled with gcc: the plain recursive fib(40) made
331,160,281 calls; the memoized version, with an array of 41 cached
results, made 79. Same answer (102,334,155), four million times fewer
calls.

Memory needs the same care as time. A full edit-distance table for two
strings of 100,000 characters has 10¹⁰ cells: 40 GB of ints. But each
row only needs the row above it, so keeping two rows of 100,001 ints
(about 800 KB) gives the same distance. Keeping the whole table is only
necessary when you also want to reconstruct the edits, by walking back
from the bottom-right corner.

The diff tool that compares two files, and git diff, are built on this
problem: the longest common subsequence of the two files' lines. Git
uses Myers' algorithm, a refinement that is fast when the files are
similar.`,
				MentalModel: "Name the subproblem in words, write how it builds on smaller ones, and fill the table in order.",
				TryIt:       "Write fib in C three ways (plain recursion, memoized, a loop with two variables), count the calls of the first two for n = 30 and 40, and time all three.",
				Analogy: `Working out a long sum in a notebook: once you have written down 37 × 12
on a line, you never compute it again, you look it up. Dynamic
programming is the discipline of keeping that notebook, and filling it
in an order where every line you need is already written.`,
				Example: `Spell checkers suggest "did you mean" words by edit distance. Biologists
align DNA and protein sequences with dynamic programming (the
Needleman–Wunsch and Smith–Waterman algorithms). diff, git diff and
merge tools find the longest common subsequence of two files, and route
planners and speech recognisers fill tables of best partial answers.`,
				Exercises: trio(
					"How many ways are there to climb a staircase of 5 steps if you can take 1 or 2 steps at a time? Of 6 steps?",
					"8 and 13. ways(n) = ways(n − 1) + ways(n − 2): your last move was either 1 step or 2. ways(1) = 1 and ways(2) = 2.",
					"In C, compute the fewest coins for every amount up to 10,000 with coins {1, 3, 4}, and print the coins used for 6, 10 and 9,999. Then compare your answers with the greedy choice for every amount, and count where they differ.",
					"Keep a second array, last[v], the coin chosen for v; to list the coins, repeat v -= last[v] until v is 0.",
					`Write a "did you mean?" program in C: read a dictionary file of words, then for each word typed on standard input, print the dictionary word with the smallest edit distance, keeping only two rows of the table.`,
					"Two arrays of length m + 1, swapped after each row. You can stop early on a word once every value in the current row exceeds the best distance found so far.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "MIT 6.006 Introduction to Algorithms (OCW)", "https://ocw.mit.edu/courses/6-006-introduction-to-algorithms-spring-2020/", "Full lectures, notes and problem sets."},
			{"Book", "Algorithms by Jeff Erickson (free)", "https://jeffe.cs.illinois.edu/teaching/algorithms/", "A clear, rigorous and free textbook."},
			{"Site", "Algorithms, 4th ed. (Sedgewick & Wayne)", "https://algs4.cs.princeton.edu/", "Code and explanations for every classic algorithm."},
			{"Tool", "VisuAlgo", "https://visualgo.net/", "Animations of sorting, trees, heaps and graph algorithms."},
			{"Site", "LeetCode", "https://leetcode.com/", "Thousands of practice problems, as used in job interviews."},
			{"Site", "Project Euler", "https://projecteuler.net/", "Maths-flavoured programming puzzles that reward efficient algorithms."},
		},
		Blueprints: []Blueprint{
			{"Data Structures Library", "Implement a dynamic array, linked list, hash map, binary heap and balanced BST from scratch, with tests and benchmarks.",
				[]string{"Dynamic array", "Linked list", "Hash map with resizing", "Binary heap", "Balanced BST (AVL or red-black)"}},
			{"Route Planner", "Load a real road network (such as an OpenStreetMap extract), then find shortest routes with Dijkstra and A*.",
				[]string{"Parse the graph", "BFS by hops", "Dijkstra by distance", "A* with a straight-line heuristic", "Compare nodes explored"}},
			{"Autocomplete Engine", "Suggest words as you type using a trie, ranked by frequency, like a search box.",
				[]string{"Build a trie", "Prefix search", "Rank by frequency with a heap", "Handle typos (edit distance)"}},
		},
		Quiz: []Question{
			{"Why can an O(n²) algorithm be fine for 1,000 items but not for 1,000,000?", "1,000² is a million steps (fast), but 1,000,000² is a trillion steps (hours or days)."},
			{"When is a linked list better than an array?", "When you insert or delete often in the middle and already hold the node; for most other uses, arrays win."},
			{"What makes a hash table fast, and what can make it slow?", "The key's hash picks its bucket directly (O(1) on average). Many collisions, from a bad hash or crafted keys, make it O(n)."},
			{"When would you use BFS rather than Dijkstra?", "When every edge has the same cost; BFS then finds the fewest-hop path more simply."},
			{"What two parts does every recursive function need?", "A base case that is answered directly, and a recursive case that calls itself on a smaller problem, moving towards the base case."},
			{"Why does a recursive function crash on a list of a million items when a loop does not?", "Every call takes a stack frame, and the stack is small (8 MiB by default on Linux), so a million nested calls overflow it. A loop uses the same few variables throughout."},
			{"How many comparisons does binary search need, at most, for a million sorted items?", "About 20, because 2²⁰ is just over a million and each comparison halves the range."},
			{"Why is mid = low + (high − low) / 2 safer than (low + high) / 2?", "low + high can exceed the largest int and overflow; high − low never does."},
			{"Give a coin set for which the greedy change-making algorithm is wrong.", "Coins {1, 3, 4} for 6: greedy gives 4 + 1 + 1, three coins, but 3 + 3 uses two."},
			{"What two properties make a problem suitable for dynamic programming?", "Optimal substructure (the best answer is built from best answers to subproblems) and overlapping subproblems (the same subproblems recur)."},
			{"What is the difference between memoization and tabulation?", "Memoization keeps the recursion and caches each result (top-down); tabulation fills a table in a loop from the smallest subproblems up (bottom-up)."},
		},
	},

	// -----------------------------------------------------------------
	3: {
		Overview: `Discrete mathematics is the maths of separate, countable things, which
is exactly what computers work with. It teaches you to reason precisely,
to prove that code is correct, to count possibilities, to compute with
remainders (the arithmetic of hashing and cryptography), to work out
what a recursive algorithm costs, to reason about networks, and to think
clearly about chance. You do not need calculus for this stage, and
every idea comes with a C program that puts it to work.`,
		Outcomes: []string{
			"Reason precisely with logic, sets and proofs, and state a loop invariant",
			"Count possibilities to judge whether brute force is feasible, without overflowing",
			"Compute with remainders: gcd, modular powers and inverses, as hashing and cryptography do",
			"Turn a recursive algorithm into a recurrence and solve it",
			"Prove simple facts about graphs and use them to check data",
			"Think correctly about probability, and spot when intuition is wrong",
		},
		Glossary: []Term{
			{"Proposition", "A statement that is either true or false."},
			{"Set", "A collection of distinct items, such as {1, 2, 3}."},
			{"Function", "In maths, a rule that gives exactly one output for each input."},
			{"Proof", "An argument showing that a statement is always true, with no cases left unchecked."},
			{"Induction", "Proving a statement for 1, then showing that each case implies the next."},
			{"Permutation", "An ordered arrangement of items."},
			{"Combination", "A selection of items where order does not matter."},
			{"Probability", "A number from 0 to 1 measuring how likely something is."},
			{"Modulo", "The remainder after division: 17 mod 5 = 2. Arithmetic mod n wraps around like a clock."},
			{"Greatest common divisor", "The largest number dividing two numbers exactly; Euclid's algorithm finds it quickly."},
			{"Prime", "A whole number above 1 whose only divisors are 1 and itself."},
			{"Recurrence", "An equation that defines a sequence, or a cost, in terms of smaller cases, such as T(n) = 2T(n/2) + n."},
			{"Degree", "The number of edges touching a node in a graph."},
			{"Expected value", "The long-run average outcome of a random process."},
		},
		Concepts: []Concept{
			{
				Name:    "Logic & Boolean Algebra",
				Summary: "True, false, and the rules for combining them.",
				Body: `A proposition is a statement that is true or false. You combine
propositions with AND (∧), OR (∨), NOT (¬) and IMPLIES (→), and a truth
table lists every case. De Morgan's laws are the key identities:
¬(A ∧ B) = ¬A ∨ ¬B, and ¬(A ∨ B) = ¬A ∧ ¬B.

"A → B" is only false when A is true and B is false, and it is not the
same as "B → A", a very common reasoning error. Its contrapositive,
"¬B → ¬A", is the same statement in disguise. Quantifiers say "for all"
(∀) and "there exists" (∃), and negating one flips it: "not every test
passed" means "some test failed".

In C, && and || are logic with one twist: they short-circuit. In
a && b, b is only evaluated if a is true, so

  if (p != NULL && p->count > 0)

is safe, and swapping the two sides would crash on NULL. A formula over
n variables has 2ⁿ rows in its truth table: checking every row is easy
for 10 variables (1,024 rows) and hopeless for 100. Deciding whether
some row makes a formula true is the famous SAT problem, which Stage 11
shows is as hard as a whole class of problems. Stage 4 builds these
operations out of transistors.`,
				Diagram: `A  B │ A AND B │ A OR B │ A → B
0  0 │    0    │   0    │   1
0  1 │    0    │   1    │   1
1  0 │    0    │   1    │   0
1  1 │    1    │   1    │   1`,
				UnderTheHood: `Compilers apply De Morgan's laws for you. These two functions compile
to exactly the same instructions with gcc -O2:

  int can_enter(int age, int member)  { return !(age < 18 || !member); }
  int can_enter2(int age, int member) { return age >= 18 && member; }

      cmp    edi, 17
      setg   dl          ; dl = (age > 17)
      xor    eax, eax
      test   esi, esi
      setne  al          ; al = (member != 0)
      and    eax, edx    ; both
      ret

No jumps: both sides are cheap and safe to evaluate, so the compiler
computes them and ANDs the bits. But for p != NULL && *p > 0 it keeps
the jump (test rdi, rdi; je), because evaluating *p when p is NULL would
crash: the short-circuit is part of the program's meaning, not an
optimisation.`,
				MentalModel: "Every condition in code is a small logical formula, so simplify it on paper first.",
				TryIt:       `Simplify "not (age < 18 or not member)" with De Morgan's laws, then check it with a truth table, and compile both versions with gcc -O2 -S to compare.`,
				Analogy: `A security door opens only if you have a badge AND know the PIN. A café
gives a discount if you are a student OR a senior. And "if it rains, the
ground is wet" does not mean that a wet ground proves it rained: someone
may have watered the lawn.`,
				Example: `Advanced search (cats AND NOT dogs), spreadsheet filters and database
WHERE clauses are Boolean logic. Tools called SAT solvers check logical
formulas with millions of variables to verify that chips and software
behave correctly, and package managers use them to find versions of
every dependency that fit together.`,
				Exercises: trio(
					`Build the truth table for "(A OR B) AND NOT (A AND B)". Which everyday phrase describes it?`,
					`It is true when exactly one input is true: "either, but not both" (exclusive or, which C writes for bits as ^).`,
					"In C, write int f(int a, int b, int c) for any formula, and a loop over the 8 rows (bits of the numbers 0 to 7) that prints its truth table. Use it to verify both De Morgan laws, by checking that the two sides agree on every row.",
					"Row r gives a = (r >> 2) & 1, b = (r >> 1) & 1, c = r & 1. Two formulas are equivalent if they agree on all 8 rows.",
					`A cinema's rule: "Children under 12 need an adult, unless the film is rated U, and nobody under 18 may see an 18-rated film." Write it as one Boolean expression, then as a C function, and test it on every case.`,
					"Name the inputs first (age band, with adult, rating), write the table of cases, then derive the expression from it.",
				),
			},
			{
				Name:    "Sets, Functions & Relations",
				Summary: "The vocabulary for describing collections and connections precisely.",
				Body: `A set is a collection of distinct items, such as {1, 2, 3}. The
operations are union (A ∪ B: in either), intersection (A ∩ B: in both),
difference (A − B) and complement. A set of n items has 2ⁿ subsets.

A relation connects items from sets, such as "is enrolled in" between
students and courses. A function is a relation where each input has
exactly one output. Functions can be one-to-one (injective), onto
(surjective) or both (bijective, which means perfectly reversible).
Equivalence relations group items into classes (for example, the same
remainder mod 3), and orders rank them (≤). SQL databases are built on
exactly this relational mathematics.

In C, a set of small numbers (0 to 63) fits in one 64-bit integer: bit
i is 1 when i is in the set. Then every set operation is one
instruction:

  uint64_t a = (1ULL << 3) | (1ULL << 5);   // {3, 5}
  a | b       union             a & b    intersection
  a & ~b      difference        a ^ b    symmetric difference
  (a & b) == a                  a is a subset of b

Bigger sets of numbers use an array of such words (a bitset) or a
sorted array; sets of strings use a hash table (Stage 2).`,
				Diagram: `┌──────── A ────────┐
│  A only   ┌───────┼──────── B ───────┐
│           │ A ∩ B │  B only          │
└───────────┼───────┘                  │
            └──────────────────────────┘`,
				UnderTheHood: `With sets as bits, a whole set operation is one CPU instruction on 64
members at once. gcc -O2 compiles a | b to a single or, a & ~b to and
with not (or one andn instruction on CPUs that have it), and the size of
a set, __builtin_popcountll(a), to one popcnt instruction when built
with -mpopcnt or -march=native (without it, gcc calls a small library
function, __popcountdi2, since older x86-64 CPUs lack the instruction).

That is why bitsets are everywhere in systems code: the Linux kernel
tracks which CPUs a process may run on as a bitmask (taskset prints it
in hex), file permissions are bits (rwx = 4 + 2 + 1, so 0755 is three
small sets), and select() describes the files it waits on as a set of
bits, fd_set.`,
				MentalModel: "Name the sets and the relations, and the problem usually becomes clear.",
				TryIt:       "Represent the sets {1, 3, 5, 7} and {3, 4, 5} as uint64_t bitmasks in C, and print their union, intersection and both differences as lists of numbers.",
				Analogy: `Guest lists for two parties: the union is everyone invited to either,
and the intersection is the people invited to both. A function is like
the codes on a vending machine: each code gives exactly one snack.`,
				Example: `"People you may know" and "customers who bought this also bought" are
built on set intersections. A SQL JOIN combines relations. Sync tools use
set difference to work out which files are new or missing, and
permission systems check whether the permissions you have are a superset
of those you need.`,
				Exercises: trio(
					"With A = {1, 2, 3, 4} and B = {3, 4, 5}, compute A ∪ B, A ∩ B, A − B and B − A. How many subsets does A have?",
					"A ∪ B = {1, 2, 3, 4, 5}, A ∩ B = {3, 4}, A − B = {1, 2}, B − A = {5}. A has 2⁴ = 16 subsets, counting {} and A itself.",
					"In C, a function is a table: int f[10] maps 0–9 to 0–9. Write functions that report whether f is injective, surjective and bijective, and for a bijection, build its inverse table and check that applying both gives back every input.",
					"Injective: no two inputs share an output (mark the outputs seen in a second array). Surjective onto 0–9: every value is some output. On a finite set of equal size, one implies the other.",
					"You have two exports of newsletter subscribers from different years, one email per line. In C, produce three lists (unsubscribed, new and loyal), comparing emails without regard to case.",
					"Lowercase every email, sort both lists with qsort and strcmp, and walk them together like a merge: equal means loyal, otherwise the smaller one is only in its own list.",
				),
			},
			{
				Name:    "Proof & Induction",
				Summary: "How to be certain, not just confident, that something always works.",
				Body: `A proof is an argument that leaves no case unchecked. A direct proof
starts from what you know and derives the claim. A proof by
contradiction assumes the opposite and reaches an impossibility (this is
how you show that √2 is irrational). A single counterexample disproves
a "for all" claim, however many examples support it.

Mathematical induction proves a claim for every n. Show that it holds in
the base case, then show that if it holds for n, it holds for n + 1,
like falling dominoes. Recursion and induction are mirror images: a
recursive function is correct if its base case is right and each call is
right whenever the smaller calls are.

A loop invariant is induction applied to code: a statement that is true
before the loop, stays true after every iteration, and, together with
the exit condition, gives the result. For a sum:

  long total = 0;
  for (int i = 0; i < n; i++)
      total += a[i];      // invariant: total = a[0] + … + a[i]
  // on exit, i == n, so total = a[0] + … + a[n − 1]

Tests check some inputs; a proof covers all of them.`,
				Diagram: `base case         step: if domino n falls, domino n + 1 falls
▌ ▌ ▌ ▌ ▌ ▌  ▶  ╱ ╱ ╱ ╱ ╱ ╱   …so every domino falls
1 2 3 4 5 6`,
				UnderTheHood: `For a function of one 32-bit int you can actually test every input. A
C program checked whether "x < 0 ? -x : x" is never negative, for all
4,294,967,296 values of x, in 2.2 seconds on a test machine, and found
exactly one counterexample: x = -2,147,483,648 (INT_MIN), whose negation
does not fit in an int. For the same reason, abs(INT_MIN) is undefined
behaviour in C.

For two 32-bit inputs the same exhaustive test would take 2.2 s × 2³²,
about 300 years; for a 64-bit input, far longer than the age of the
universe. Beyond the smallest inputs, only proof covers every case,
which is why tools such as Frama-C and CBMC prove C code correct, and
why the seL4 kernel's correctness is a machine-checked proof rather than
a test suite.`,
				MentalModel: "Prove the first case, prove each case gives the next, and every case is done.",
				TryIt:       "Prove by induction that 1 + 2 + … + n = n(n + 1)/2, then check it in C for every n up to 100,000 with long long, and find the first n where it overflows a 32-bit int.",
				Analogy: `Climbing a ladder: if you can get onto the first rung, and from any rung
you can reach the next one, you can climb as high as you like.`,
				Example: `Amazon Web Services has used the formal specification language TLA+ to
check the designs of services such as S3 and DynamoDB, catching subtle
bugs that testing had missed. The seL4 operating-system kernel is
mathematically proven correct and is used in safety-critical systems.`,
				Exercises: trio(
					`Find the flaw: "All horses are the same colour. One horse is trivially one colour. If any n horses share a colour, so do n + 1, because the first n and the last n overlap."`,
					"Check the step from n = 1 to n = 2: the first horse and the last horse are two groups of one that do not overlap, so nothing links their colours.",
					"Write binary search in C, state its loop invariant in a comment, and add an assert that checks the invariant on every iteration. Run it on 10,000 random sorted arrays.",
					`The invariant: "if the target is in the array, it lies between low and high". The assert can scan a[0..low-1] and a[high+1..n-1] to confirm the target is not there (slow, but only while testing).`,
					"Your bank app rounds each of 3 interest payments to the cent. Prove, or disprove with a counterexample, that rounding each payment gives the same total as rounding their sum.",
					"Try three payments of 0.4 cents each: each rounds to 0, totalling 0, but their sum is 1.2 cents, which rounds to 1. One counterexample is enough to disprove it.",
				),
			},
			{
				Name:    "Counting & Combinatorics",
				Summary: "How many ways? The maths behind passwords, probabilities and brute force.",
				Body: `The product rule: if one choice can be made in a ways and another in b
ways, there are a × b combinations. Permutations count ordered
arrangements: n! ways to order n items, and n!/(n − k)! ways to order k
of them. Combinations count unordered selections: C(n, k) = n! / (k!(n −
k)!). Each row of Pascal's triangle holds the C(n, k), and each number
is the sum of the two above it: C(n, k) = C(n − 1, k − 1) + C(n − 1, k).

The pigeonhole principle: put more than n items into n boxes, and some
box holds two. That is why hash collisions must exist, and why no
lossless compressor can shrink every file. Counting tells you whether
"try every possibility" takes seconds or longer than the age of the
universe.

In code, counts grow faster than integer types. 20!, about 2.4 × 10¹⁸,
is the largest factorial that fits in an unsigned 64-bit integer, so
computing C(n, k) as n! / (k!(n − k)!) silently goes wrong from n = 21. Compute it multiplicatively instead, dividing as you go:

  unsigned long long c = 1;
  for (int i = 1; i <= k; i++)
      c = c * (n - k + i) / i;   // exact: c is C(n - k + i, i) here

which stays exact as long as the intermediate product fits.`,
				Diagram: `Pascal's triangle: each number is the sum of the two above it
row 0:            1
row 1:          1   1
row 2:        1   2   1
row 3:      1   3   3   1
row 4:    1   4   6   4   1        C(4, 2) = 6`,
				UnderTheHood: `Unsigned arithmetic in C does not fail when it overflows: it wraps
around modulo 2⁶⁴, silently. A loop multiplying factorials into an
unsigned long long printed:

  20! = 2432902008176640000
  21! = 14197454024290336768     (the true value is 51090942171709440000)

The second number is 21! minus 2 × 2⁶⁴: wrong, and nothing warns you.
For signed ints, overflow is undefined behaviour instead, which
-fsanitize=undefined catches. gcc and clang also offer
__builtin_mul_overflow(a, b, &result), which returns true when the
product does not fit, usually at the cost of one extra instruction that
tests the CPU's overflow flag: cheap insurance for counting code.`,
				MentalModel: "Count before you brute-force, and check the count fits its type.",
				TryIt:       "Count the 4-digit PINs with no repeated digit, first with a formula and then with a C program that loops over all 10,000 PINs.",
				Analogy: `Outfits: 3 shirts × 4 pairs of trousers × 2 pairs of shoes gives 24
outfits. Picking 3 pizza toppings from 10 does not care about order
(combinations); awarding gold, silver and bronze does (permutations).`,
				Example: `A 4-digit PIN has 10⁴ = 10,000 options, which is why cards lock after
three wrong tries. An 8-character password drawn from 94 symbols has
about 6 × 10¹⁵ options, but people pick words, so attackers use
dictionaries instead. The chance of winning a 6-from-49 lottery is 1 in
C(49, 6) = 13,983,816.`,
				Exercises: trio(
					"How many 3-letter codes can you make from A to Z if letters may repeat? And if they may not? How many ways to choose 3 letters, order ignored?",
					"26 × 26 × 26 = 17,576; 26 × 25 × 24 = 15,600; and 15,600 / 3! = 2,600 = C(26, 3).",
					"In C, write unsigned long long choose(int n, int k) the multiplicative way, check it against Pascal's triangle for every n up to 60, and find the first n where the factorial formula gives a wrong answer.",
					"Fill a 61 × 61 table with C(n, k) = C(n − 1, k − 1) + C(n − 1, k). The factorial formula breaks at n = 21, when 21! no longer fits in 64 bits.",
					"A website requires 8-character passwords of lowercase letters. An attacker can try 10 billion guesses per second offline. How long to try them all? And for 12 characters mixing upper case, lower case and digits? Write the calculation as a small C program, using double for the counts.",
					"26⁸ ≈ 2.1 × 10¹¹ takes about 20 seconds; 62¹² ≈ 3.2 × 10²¹ takes about 10,000 years.",
				),
			},
			{
				Name:    "Modular Arithmetic & Number Theory",
				Summary: "Arithmetic that wraps around like a clock: the maths of hashing, checksums and cryptography.",
				Body: `a mod n is the remainder when a is divided by n: 17 mod 5 = 2. Two
numbers are congruent mod n when they have the same remainder, and you
can add, subtract and multiply first and take the remainder at the end,
or take it at every step and get the same answer:

  (a + b) mod n = ((a mod n) + (b mod n)) mod n
  (a × b) mod n = ((a mod n) × (b mod n)) mod n

That is what keeps numbers small. To compute 7^1000000 mod 13, square
and multiply, reducing mod 13 each time (fast power, from Stage 2):
about 40 multiplications, and no number ever exceeds 13 × 13.

The greatest common divisor comes from Euclid's algorithm, over 2,000
years old and still the fastest simple way: gcd(a, b) = gcd(b, a mod b),
and gcd(a, 0) = a. When gcd(a, n) = 1, a has an inverse mod n (a number
x with a × x ≡ 1), found by the extended version of the same algorithm;
that is how RSA computes its private key. A prime has no divisors but 1
and itself, and the sieve of Eratosthenes finds every prime below n by
crossing out multiples.

C's % is not quite mod: it follows the sign of the left side, so
-7 % 3 is -1, not 2. For a true mod of a possibly negative a, use
((a % n) + n) % n.`,
				Diagram: `a clock is arithmetic mod 12
            12
       11        1         9 o'clock + 5 hours = 2 o'clock
    10              2      (9 + 5) mod 12 = 2
     9      •       3
     8              4      gcd(48, 18):  48 mod 18 = 12
       7         5                       18 mod 12 = 6
            6                            12 mod 6  = 0   → 6`,
				UnderTheHood: `Your computer already does modular arithmetic on every unsigned
operation. C defines unsigned ints to wrap around modulo 2³² (or 2⁶⁴):
4294967295u + 1u is 0, and the CPU's add instruction simply drops the
carry out of the top bit. That is why hash functions such as FNV-1a are
written with unsigned types: the multiply is meant to wrap.

Euclid's algorithm is fast because the numbers shrink at least as fast
as the Fibonacci numbers grow (Lamé's theorem), and consecutive
Fibonacci numbers are its worst case. A test program computing
gcd(1836311903, 1134903170), the 46th and 45th Fibonacci numbers, took
44 steps, the most for any pair below 2³¹. Its cost grows with the
number of digits, not with the numbers themselves.

When numbers get close to the limit, the product needs a wider type:
for a, b < n below 2³², (a × b) mod n is exact in uint64_t, but for n
near 2⁶³ the product would overflow, and cryptography libraries use
128-bit products or special multiplication tricks.`,
				MentalModel: "Reduce early and often: the remainder carries everything you need.",
				TryIt:       "In C, print -7 % 3, 7 % -3 and ((-7 % 3) + 3) % 3, then write gcd and test it on 48 and 18, on consecutive Fibonacci numbers, and on a number and 0.",
				Analogy: `A clock: 9 o'clock plus 5 hours is 2 o'clock, not 14 o'clock. Days of
the week work mod 7: 100 days after a Monday is a Wednesday, because
100 = 14 × 7 + 2, and the 14 whole weeks change nothing.`,
				Example: `ISBN check digits, the Luhn check on credit-card numbers and IBAN bank
codes are all mod-n checksums that catch typos. Hash tables pick a
bucket with h mod size. RSA and Diffie–Hellman, which protect HTTPS
connections, are modular exponentiation with numbers of hundreds of
digits, and the security of RSA rests on how hard it is to factor a
product of two large primes.`,
				Exercises: trio(
					"Compute by hand: 17 mod 5, 3¹⁰ mod 7 (reduce as you go), gcd(48, 18) with Euclid's algorithm, and the day of the week 100 days after a Monday.",
					"2; 3² = 9 ≡ 2, so 3¹⁰ = (3²)⁵ ≡ 2⁵ = 32 ≡ 4; gcd is 6 (48 → 18 → 12 → 6 → 0); 100 = 14 × 7 + 2, so two days later: Wednesday.",
					"In C, write uint64_t mod_pow(uint64_t a, uint64_t e, uint64_t n) by repeated squaring, check it against a slow loop for small e, then compute 2^(10^18) mod 1,000,000,007. Write the sieve of Eratosthenes and count the primes below 10 million.",
					"Keep every intermediate below n, so the product of two of them fits in 64 bits when n < 2³². There are 664,579 primes below 10 million; the sieve is a char array of 10 million flags.",
					"Validate credit-card numbers with the Luhn algorithm in C: from the right, double every second digit (subtracting 9 when the result exceeds 9), add everything, and the number is valid when the total mod 10 is 0. Test it on the published test numbers, such as 4539 1488 0343 6467.",
					"Read the number as a string, skip spaces, and walk it from the last digit to the first with a flag that alternates.",
				),
			},
			{
				Name:    "Recurrences: The Cost of Recursion",
				Summary: "Write down what a recursive algorithm costs in terms of itself, then solve the equation.",
				Body: `The running time of a recursive algorithm is naturally recursive too.
Merge sort sorts two halves and merges them in linear time, so its cost
T(n) satisfies T(n) = 2T(n/2) + n, with T(1) = 1. A recurrence like this
says what the cost is in terms of smaller costs; solving it gives the
Big-O.

Three ways to solve one:
- unroll it: T(n) = n + 2(n/2) + 4(n/4) + … has log₂ n levels of n
  work each, so T(n) is about n log₂ n;
- draw the recursion tree: each node is one call, labelled with the work
  it does outside its children, and the total is the sum over all nodes;
- guess the answer and prove it by induction (Proof & Induction).

A few patterns cover most code:

  T(n) = T(n − 1) + 1        O(n)        a loop written as recursion
  T(n) = T(n − 1) + n        O(n²)       selection sort
  T(n) = T(n/2) + 1          O(log n)    binary search
  T(n) = 2T(n/2) + n         O(n log n)  merge sort
  T(n) = 2T(n − 1) + 1       O(2ⁿ)       Towers of Hanoi
  T(n) = T(n − 1) + T(n − 2) O(1.62ⁿ)    plain recursive Fibonacci

The master theorem handles T(n) = aT(n/b) + f(n) in general, by
comparing f(n) with n^(log_b a): whichever grows faster dominates, and
when they match, a log factor appears.`,
				Diagram: `recursion tree for T(n) = 2T(n/2) + n

level 0              n                   work n
level 1        n/2       n/2             work n
level 2     n/4  n/4  n/4  n/4           work n
  …                  …                     …
level log₂ n  1 1 1 1 … 1 1 1 1          work n
                                  total: n × log₂ n`,
				UnderTheHood: `The recurrence predicts real counts closely. A merge sort in C that
counted its comparisons on random arrays gave:

  n            comparisons     n log₂ n
  1,024              8,965        10,240
  32,768           450,080       491,520
  1,048,576     19,645,391    20,971,520

The count stays just below n log₂ n (a merge of two runs never needs
more comparisons than items, and often stops early), and grows by a
factor of about 43.6 from the second row to the third, where n grew 32
times: the extra log factor, exactly as T(n) = 2T(n/2) + n says.

Recurrences also predict stack depth. The depth of merge sort's
recursion is log₂ n, about 20 calls for a million items, so it never
overflows the stack. Quicksort's depth is log n on average but n in the
worst case, which is why careful implementations recurse into the
smaller part first and loop over the larger one: that caps the depth at
log₂ n.`,
				MentalModel: "Cost of a call = work done here + cost of the calls it makes; sum it level by level.",
				TryIt:       "Add a comparison counter to a merge sort in C and print the count for n = 1,000, 10,000 and 100,000 next to n × log₂(n). Then count the calls of the plain recursive fib for n = 20 to 30 and watch each one grow by about 1.6 times.",
				Analogy: `A company where every manager splits a job in two and hands each half
to a junior, then spends an hour combining their results. To find the
total hours, you do not follow one person; you add up the hours level
by level down the organisation chart: the recursion tree.`,
				Example: `Every divide-and-conquer algorithm you will meet has a recurrence:
binary search, merge sort, quicksort, fast Fourier transforms (used in
audio, image compression and fast multiplication), and Karatsuba
multiplication, which multiplies huge numbers in about n^1.58 steps
instead of n², as big-number libraries such as GMP do.`,
				Exercises: trio(
					"Solve by unrolling: T(n) = T(n − 1) + 2 with T(0) = 0, and T(n) = T(n/2) + 1 with T(1) = 0 (n a power of 2). What does each describe?",
					"T(n) = 2n: a loop doing constant work per step. T(n) = log₂ n: halving, as in binary search.",
					"In C, write merge sort with a global comparison counter, and the plain recursive fib with a call counter. Tabulate both against the formulas n log₂ n and fib's own recurrence C(n) = C(n − 1) + C(n − 2) + 1.",
					"The calls of fib(n) satisfy C(0) = C(1) = 1 and C(n) = C(n − 1) + C(n − 2) + 1, which works out to 2 fib(n + 1) − 1: 331,160,281 for n = 40.",
					"Count the inversions of a list (pairs that are out of order, a measure of how unsorted it is, used to compare rankings) in O(n log n) by adapting merge sort, and check it against the O(n²) count on random lists. Write its recurrence.",
					"While merging, when an item from the right half is taken before the remaining items of the left half, it forms an inversion with each of them: add the number of items left in the left half. T(n) = 2T(n/2) + n.",
				),
			},
			{
				Name:    "Graph Theory: Degrees, Trees & Paths",
				Summary: "The mathematics of networks: simple counting facts that let you check, bound and route.",
				Body: `A graph is a set of vertices and a set of edges between pairs of them
(Stage 2 shows how to store and search one; this lesson is about what
you can prove). The degree of a vertex is the number of edges touching
it. The handshake lemma: the degrees add up to twice the number of
edges, because every edge has two ends. So the number of odd-degree
vertices is always even.

A tree is a connected graph with no cycles. Every tree with n vertices
has exactly n − 1 edges, between any two vertices there is exactly one
path, and adding any edge creates exactly one cycle. A spanning tree of
a connected graph keeps every vertex and just enough edges to stay
connected.

An Euler path uses every edge exactly once. Euler showed in 1736, with
the bridges of Königsberg, that a connected graph has an Euler circuit
(one that returns to its start) exactly when every vertex has even
degree, and an Euler path that is not a circuit exactly when two
vertices have odd degree, which are then its two ends. A bipartite graph
splits its vertices into two sides with every edge crossing between
them; that is possible exactly when the graph has no cycle of odd
length. A directed acyclic graph (DAG) is one with no directed cycle,
which is what any set of dependencies must be.`,
				Diagram: `Königsberg in 1736: four land masses and seven bridges

  A ═ B   two bridges        degrees:  A (the island) 5
  A ═ C   two bridges                  B 3,  C 3,  D 3
  A ─ D   one bridge
  B ─ D   one bridge         four odd-degree vertices, more than two,
  C ─ D   one bridge         so no walk crosses every bridge exactly once`,
				UnderTheHood: `Programs use these facts to check data and detect impossible requests.
GNU make treats a Makefile as a dependency graph that must be a DAG. A
Makefile where a needs b and b needs a gets this answer from GNU Make
4.3:

  make: Circular b <- a dependency dropped.

It finds the cycle the way a depth-first search does: it reaches a
target that is still being worked on further up its own path.
Git's commit history is a DAG too: a commit can have two parents (a
merge), but never itself as an ancestor.

Graph colouring, giving adjacent vertices different colours, is how
compilers assign registers: variables that are alive at the same time
are joined by an edge and must get different registers, and a register
allocator colours that interference graph. The four-colour theorem (any
map needs at most four colours) was first proved in 1976 by Appel and
Haken with a computer checking nearly 2,000 configurations: the first
major theorem whose proof was too long to check by hand.`,
				MentalModel: "Count the ends of the edges: degrees, parity and n − 1 tell you what is possible before you search.",
				TryIt:       "In C, read a list of edges, count each vertex's degree, check that the degrees sum to twice the number of edges, and print the odd-degree vertices: can you draw this graph without lifting your pen?",
				Analogy: `Drawing a house shape without lifting your pen: it works only if you
start at one of the two corners where an odd number of lines meet. A
family tree is a tree: everyone has exactly one path back to a common
ancestor, and adding a marriage between cousins creates a cycle.`,
				Example: `Ethernet switches run the Spanning Tree Protocol so that a network with
redundant cables forwards frames along a tree and never in loops. Road
gritters and postal routes are planned as Euler paths (the Chinese
postman problem), build systems and package managers depend on DAGs, and
scheduling exams so that no student has two at once is graph colouring.`,
				Exercises: trio(
					"A graph has 6 vertices with degrees 3, 3, 2, 2, 1 and 1. How many edges does it have? Could a graph have degrees 3, 3, 3, 2 and 2?",
					"The degrees sum to 12, so 6 edges. 3 + 3 + 3 + 2 + 2 = 13 is odd, and degrees always sum to an even number (twice the edges): impossible.",
					"In C, read an undirected graph (n, then edges as pairs) and report whether it is a tree, by checking that it has exactly n − 1 edges and is connected (a BFS from vertex 0 reaches all n vertices).",
					"Either check alone is not enough: n − 1 edges can hide a cycle plus a disconnected vertex, and a connected graph can have extra edges. Both together mean a tree.",
					"Decide in C whether a set of exams can be scheduled in two time slots so that no student sits two exams at once: build the conflict graph and try to 2-colour it with BFS, printing the two slots or an odd cycle that makes it impossible.",
					"Colour the start vertex 0 and each neighbour the opposite colour; a neighbour that already has the same colour as you means an odd cycle. Do it from every uncoloured vertex, for graphs in several pieces.",
				),
			},
			{
				Name:    "Probability for Programmers",
				Summary: "Reasoning about chance, and why intuition is often wrong.",
				Body: `Probability measures how likely an event is, from 0 (impossible) to 1
(certain). For equally likely outcomes, P = favourable ÷ total.
Independent events multiply: two heads in a row is ½ × ½. The expected
value is the long-run average, and it adds up across events even when
they are not independent (linearity of expectation).

Conditional probability P(A|B) is the chance of A given that B happened,
and Bayes' rule flips it: P(A|B) = P(B|A) · P(A) / P(B). Intuition fails
often: in a room of just 23 people, there is about a 50% chance that two
share a birthday. The complement is the trick: the chance that all 23
birthdays differ is (365/365) × (364/365) × … × (343/365), about 0.493,
so the chance of a shared one is about 0.507.

Programmers use probability for randomised algorithms, hashing, load
balancing, A/B tests, simulations and machine learning (Stage 14). When
the formula is hard, simulate: run the experiment a million times with
a random number generator and count. In C, rand() is weak and rand() % n
is slightly biased; for simulations use a better generator, such as
xorshift or PCG.`,
				Diagram: `chance that at least two of k people share a birthday
  k = 10    ██▍                  11.7%
  k = 23    ██████████▏          50.7%
  k = 41    ██████████████████   90.3%
  k = 70    ████████████████████ 99.9%`,
				UnderTheHood: `The formula and a simulation agree. In C, the exact probability for 23
people is 1 − (365 × 364 × … × 343) / 365²³ = 0.507297. A simulation of
a million rooms of 23 people with a xorshift random generator found a
shared birthday in 50.7349% of them.

Why rand() % n is biased: rand() returns one of RAND_MAX + 1 equally
likely values, and unless n divides that count, the low remainders get
one extra value each. glibc's RAND_MAX is 2,147,483,647, so the bias of
rand() % 6 is tiny; but some C libraries (Microsoft's among them) have
RAND_MAX = 32,767, and rand() % 1000 there gives each of 0 to 767 a
chance of 33/32768 and each of 768 to 999 a chance of 32/32768: values
below 768 come up about 3% more often. Rejecting the top, uneven part of
the range removes the bias.`,
				MentalModel: "When intuition and arithmetic disagree, simulate.",
				TryIt:       "Compute the birthday probability for k = 1 to 70 exactly in C, then simulate k = 23 a million times with xorshift and compare.",
				Analogy: `A forecast of "70% chance of rain" means that on days like this, it
rained about 7 times out of 10. And a medical test that is "99%
accurate" can still produce mostly false alarms when the disease is
rare, because healthy people vastly outnumber sick ones.`,
				Example: `Spam filters apply Bayes' rule to the words in an email. A/B tests
decide whether a new design is really better or just lucky. The birthday
paradox is why hash functions need long outputs to avoid collisions: a
64-bit hash has a 50% chance of a collision after only about 5 billion
items.`,
				Exercises: trio(
					"What is the chance of rolling a total of 7 with two dice? And a total of 12? What is the expected total?",
					"List all 36 outcomes: 6/36 for a 7 and 1/36 for a 12. The expected total is 3.5 + 3.5 = 7, by linearity of expectation.",
					`In C, simulate the Monty Hall problem 1,000,000 times with a xorshift random generator and measure the win rate for "stay" and for "switch".`,
					"Randomise the car's door and your first pick; the host opens a goat door that you did not pick. Switching wins exactly when your first pick was wrong: 2/3 of the time.",
					"A disease affects 1 in 1,000 people. A test catches 99% of cases but also flags 5% of healthy people. If you test positive, what is the chance you are sick? Work it out with Bayes' rule, then simulate 1 million people in C.",
					"About 2%: 0.99 × 0.001 / (0.99 × 0.001 + 0.05 × 0.999) ≈ 0.0194. The false positives from the healthy majority swamp the true positives.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "MIT 6.042J Mathematics for Computer Science (OCW)", "https://ocw.mit.edu/courses/6-042j-mathematics-for-computer-science-fall-2010/", "The standard course; the lecture notes are a full free textbook."},
			{"Book", "Book of Proof (free)", "https://richardhammack.github.io/BookOfProof/", "A gentle, clear introduction to proofs."},
			{"Course", "Khan Academy: Statistics & Probability", "https://www.khanacademy.org/math/statistics-probability", "Friendly videos and exercises on probability."},
			{"Tool", "Seeing Theory", "https://seeing-theory.brown.edu/", "Interactive visual introduction to probability."},
		},
		Blueprints: []Blueprint{
			{"Truth-Table & Logic Toolkit", "A program that parses Boolean formulas, prints truth tables, checks equivalence and simplifies expressions.",
				[]string{"Formula parser", "Truth tables", "Equivalence checker", "Find a satisfying assignment", "Simplification rules"}},
			{"Probability Simulator", "Monte Carlo simulations of classic puzzles (birthdays, Monty Hall, dice, card games), compared with exact formulas.",
				[]string{"Random experiments", "Exact formulas", "Convergence plots", "Your own new puzzle"}},
			{"Password Strength Estimator", "Estimate how long an offline attacker would need to crack a password, accounting for dictionaries and patterns.",
				[]string{"Brute-force counting", "Dictionary check", "Common substitutions", "Time-to-crack report"}},
		},
		Quiz: []Question{
			{"State De Morgan's laws.", "NOT (A AND B) = (NOT A) OR (NOT B), and NOT (A OR B) = (NOT A) AND (NOT B)."},
			{"What are the two parts of a proof by induction?", "A base case (the claim holds for the first value) and an inductive step (if it holds for n, it holds for n + 1)."},
			{"How many ways can you choose 2 people from 5, if order does not matter?", "C(5, 2) = 10."},
			{"Why can a 99%-accurate test give mostly false positives?", "When the condition is rare, the small false-positive rate applied to the huge healthy group outnumbers the true positives."},
			{"Why is p != NULL && p->x > 0 safe, but p->x > 0 && p != NULL is not?", "&& short-circuits: the right side is only evaluated when the left is true. In the second order, p->x is read before p is checked."},
			{"What is a loop invariant?", "A statement that is true before a loop and stays true after every iteration; together with the exit condition, it proves what the loop computes."},
			{"Why does computing C(n, k) as n! / (k!(n − k)!) fail in C from n = 21?", "21! no longer fits in 64 bits, and unsigned arithmetic wraps silently. Multiply and divide as you go instead."},
			{"What is -7 % 3 in C, and how do you get the mathematical remainder?", "-1, because C's % takes the sign of the left side; ((a % n) + n) % n gives 2."},
			{"What does Euclid's algorithm compute, and how?", "The greatest common divisor, using gcd(a, b) = gcd(b, a mod b) until the second number is 0."},
			{"What does the recurrence T(n) = 2T(n/2) + n solve to, and which algorithm has it?", "O(n log n): log₂ n levels with n work each. Merge sort."},
			{"How many edges does a tree with n vertices have?", "Exactly n − 1."},
			{"When does a connected graph have an Euler circuit?", "Exactly when every vertex has even degree."},
			{"In a room of 23 people, what is the chance that two share a birthday?", "About 50.7%: the chance that all differ, 365/365 × 364/365 × … × 343/365, is about 49.3%."},
		},
	},

	// -----------------------------------------------------------------
	4: {
		Overview: `How does a lump of silicon run your code? This stage climbs from a
single switch to a working CPU: binary numbers, logic gates, an ALU that
adds, subtracts and compares, memory circuits, state machines, the
instruction cycle, the machine code your compiler writes, and the tricks
that make modern chips fast. Every step is checked against the real
thing: the bytes gcc emits, the flags the CPU sets, the time a cache miss
costs. After it, the machine is no longer magic. It is a clever
arrangement of on/off switches.`,
		Outcomes: []string{
			"Explain how numbers, text and images are stored as bits, and read a hex dump",
			"Design small digital circuits, including an adder, from NAND gates up",
			"Explain how a CPU subtracts and compares with one adder, and what the carry and overflow flags mean",
			"Model a controller as a finite state machine, in hardware or in C",
			"Describe what a CPU does on every clock tick, and read the machine code your compiler produces",
			"Explain why some code runs faster: caches, pipelines and branch prediction",
		},
		Glossary: []Term{
			{"Bit", "A single binary digit: 0 or 1."},
			{"Byte", "8 bits, able to hold a value from 0 to 255."},
			{"Hexadecimal", "Base-16 notation (0–9, A–F) used as shorthand for binary; 0xFF = 255."},
			{"Transistor", "A tiny electrically controlled switch, the building block of chips."},
			{"Logic gate", "A circuit that computes a logical function (AND, OR, NOT) on bits."},
			{"Clock", "A signal that ticks billions of times a second to keep the chip's parts in step."},
			{"ALU", "Arithmetic Logic Unit: the part of the CPU that does arithmetic and comparisons."},
			{"ISA", "Instruction Set Architecture: the list of instructions a CPU understands (x86-64, ARM, RISC-V)."},
			{"Core", "One independent processing unit; modern chips have several."},
			{"Two's complement", "The way CPUs store negative integers: −x is the bits of x inverted, plus one."},
			{"Flags", "Bits the ALU sets after an operation (zero, negative, carry, overflow) that conditional jumps test."},
			{"Finite state machine", "A system that is always in one of a fixed set of states and moves between them on inputs."},
			{"Opcode", "The part of a machine instruction that says which operation to perform."},
			{"Endianness", "The order of a number's bytes in memory; x86 and most ARM systems store the lowest byte first (little-endian)."},
		},
		Concepts: []Concept{
			{
				Name:    "Binary & Number Representation",
				Summary: "Everything in a computer (numbers, text, images) is patterns of 0s and 1s.",
				Body: `A bit is one binary digit. Binary is base 2, so 1011 = 8 + 0 + 2 + 1 = 11.
Eight bits make a byte (0 to 255). Hexadecimal (base 16, with digits
0–9 and A–F) is shorthand for 4 bits per digit: 0xFF = 255, and a byte
is always exactly two hex digits.

Text is numbers too. In ASCII, 'A' = 65, and Unicode (usually stored as
UTF-8) extends this to every writing system and to emoji, using one to
four bytes per character. Colours are three bytes (red, green, blue), so
#FF8800 is orange. Meaning comes from interpretation: the same bytes can
be a number, a letter, a pixel or an instruction.

Converting by hand: to read binary, add the place values of the 1s; to
write a number in binary, repeatedly divide by 2 and collect the
remainders, which come out from the lowest bit up. In C:

  for (int i = 31; i >= 0; i--)
      putchar((x >> i & 1) ? '1' : '0');    // x unsigned, bit 31 first

printf prints hex with %x, but has no standard binary format, so you
write that loop yourself. Numbers bigger than one byte also have a byte
order in memory (endianness): see below.`,
				Diagram: `bit value:  128  64  32  16   8   4   2   1
bits:         0   1   0   0   0   0   0   1   = 64 + 1 = 65 = 'A'
hex:          └─── 4 ───┘   └─── 1 ───┘       = 0x41`,
				UnderTheHood: `What the bytes really are, from od -An -tx1 (or xxd) on Linux:

  printf 'A é 😀' | od -An -tx1
   41 20 c3 a9 20 f0 9f 98 80

'A' is the single byte 41, the space is 20, é takes two bytes (c3 a9)
and the emoji takes four (f0 9f 98 80): that is UTF-8. A string's
length in bytes (what strlen counts) is not its number of characters.

Byte order: storing the 32-bit number 0x01020304 and printing its bytes
in memory order, on an x86-64 machine, gives 04 03 02 01, lowest byte
first. That is little-endian, used by x86 and almost every ARM system;
network protocols send the highest byte first (big-endian), which is
why C has htonl and ntohl to convert. And interpretation is everything:
the byte 200 read as an int8_t is -56, the same bits as uint8_t are
200.`,
				MentalModel: "Bits have no meaning until you decide how to read them.",
				TryIt:       "Convert your age and your birth year to binary and hex by hand, then check them with a C program that prints the bits with the loop above and the hex with %x.",
				Analogy: `Morse code turns letters into dots and dashes; binary turns everything
into 0s and 1s. A row of light switches, each on or off, can encode any
number, as long as everyone agrees what each switch is worth.`,
				Example: `Web colour codes such as #1DA1F2 are hex RGB values. An IPv4 address
such as 192.168.1.1 is a 32-bit number written as four bytes. The emoji
😀 is Unicode code point U+1F600, stored as 4 bytes in UTF-8.`,
				Exercises: trio(
					"Convert 13, 100 and 255 to binary, and convert 0x2A and 0x10 to decimal. How many hex digits does a 32-bit number need?",
					"1101, 1100100 and 11111111; 0x2A = 42 and 0x10 = 16. Each hex digit is 4 bits, so 8 digits.",
					"In C, write void to_binary(uint32_t x, char out[33]) and uint32_t from_binary(const char *s) without library conversions, and test that from_binary(to_binary(x)) == x for a million random x, and for 0 and 4294967295.",
					"Write the bits from bit 31 down, then terminate the string. from_binary: for each character, x = x * 2 + (c - '0').",
					`Look at a small PNG and a text file byte by byte (xxd on most systems, or od -An -tx1 anywhere) and identify the PNG's "magic number" header and any readable text in either file. Then write your own hex dumper in C: offset, 16 bytes in hex, and the printable characters.`,
					`PNG files start with 89 50 4E 47 0D 0A 1A 0A; the bytes 50 4E 47 spell "PNG". Read the file with fread in blocks of 16 bytes, and print a byte as a character only if isprint() says so.`,
				),
			},
			{
				Name:    "Logic Gates & Circuits",
				Summary: "Tiny switches combined into circuits that compute.",
				Body: `A transistor is an electrically controlled switch. Wire a few together
and you get logic gates: NOT, AND, OR, XOR and NAND. A NAND gate takes
four transistors in CMOS, the technology of almost every chip, and NAND
alone is universal: any circuit can be built from NAND gates only.

  NOT a     = a NAND a                            1 NAND
  a AND b   = NOT (a NAND b)                      2 NANDs
  a OR b    = (NOT a) NAND (NOT b)                3 NANDs
  a XOR b   = with m = a NAND b:
              (a NAND m) NAND (b NAND m)          4 NANDs

Combinational circuits compute their outputs from their current inputs
only. A half adder uses XOR for the sum bit and AND for the carry. A
full adder also takes a carry in, and chaining 32 of them, each carry
feeding the next, adds two 32-bit numbers (a ripple-carry adder). A
multiplexer selects one of several inputs: out = (a AND s) OR (b AND NOT
s). Modern processors contain tens of billions of transistors.`,
				Diagram: `A ─┬─────▶ [ XOR ] ──▶ sum        A B │ sum carry
B ─┼─┬───▶ [     ]                0 0 │  0    0
   │ │                            0 1 │  1    0
   └─┼───▶ [ AND ] ──▶ carry      1 0 │  1    0
     └───▶ [     ]                1 1 │  0    1   (1 + 1 = 10 in binary)`,
				UnderTheHood: `C's bitwise operators are gates, 32 or 64 of them side by side. gcc -O2
compiles unsigned mix(unsigned a, unsigned b) { return a ^ b; } to a
single instruction,

  xor   eax, esi

which is 32 XOR gates working in parallel in one clock cycle. And the
multiplexer is an instruction too: int pick(int c, int a, int b) {
return c ? a : b; } compiles to

  test  edi, edi
  mov   eax, esi
  cmove eax, edx        ; if c was zero, take b instead

with no jump: the CPU computes both inputs and a multiplexer picks one.

A ripple-carry adder is slow because the carry may have to travel
through every bit, one gate delay after another. Real ALUs use
carry-lookahead adders, which compute all the carries at once with more
gates, so a 64-bit add finishes in one cycle.`,
				MentalModel: "Arithmetic is just logic, and logic is just switches.",
				TryIt:       "Build a half adder and a full adder in a free logic simulator (Digital or Logisim Evolution), then chain four into a 4-bit adder and add 7 + 9.",
				Analogy: `Two light switches wired one after the other (both must be on for the
bulb to light) make AND. Wired side by side (either one will do), they
make OR. A hallway light with a switch at each end, where flipping either
switch toggles the light, behaves like XOR.`,
				Example: `Every chip in your phone and laptop is built from gates like these, tens
of billions of transistors at a time. Minecraft players have even built
working computers out of redstone logic gates.`,
				Exercises: trio(
					"Build NOT, AND and OR using only NAND gates, on paper. How many NAND gates does each need?",
					"NOT a = a NAND a (1). AND = NOT(a NAND b) (2). OR = (NOT a) NAND (NOT b) (3).",
					"In C, write int nand(int a, int b), then not_, and_, or_ and xor_ using only nand, then a full adder from those, and add two 8-bit numbers bit by bit (ripple carry). Verify it against + for every pair from 0 to 255.",
					"Full adder: sum = a XOR b XOR cin; cout = (a AND b) OR (cin AND (a XOR b)). The carry out of bit i is the carry into bit i + 1.",
					"Design the logic for a car's seat-belt warning: beep if the engine is on AND (the driver's belt is off OR (the passenger seat is occupied AND the passenger's belt is off)). Write the truth table, simplify, and simulate it in C for all 16 cases.",
					"Name each sensor as an input bit, loop over the 16 values of a 4-bit counter, and print the row and the result.",
				),
			},
			{
				Name:    "Building an ALU: Adders, Subtraction & Flags",
				Summary: "One adder does addition, subtraction and comparison; four flag bits tell the program what happened.",
				Body: `An ALU (arithmetic logic unit) takes two numbers and an operation code,
and produces a result and a few flag bits. The surprising part is how
little it needs: the adder from the last lesson does nearly everything.

Subtraction is addition. CPUs store negative numbers in two's
complement: −x is the bits of x inverted, plus one, so in 8 bits −1 is
11111111 and −128 is 10000000. Then a − b = a + (NOT b) + 1, which is
the adder with b's bits inverted and the carry-in set to 1. The same
circuit adds signed and unsigned numbers: only the interpretation of the
result differs.

After each operation the ALU sets flags, which conditional jumps read:
- Z, zero: the result is 0;
- N (or S), negative: the top bit of the result is 1;
- C, carry: an unsigned result did not fit (a carry out of the top bit,
  or for subtraction, a borrow; ARM records the opposite of the borrow);
- V (or O), overflow: a signed result did not fit (two positives gave a
  negative, or two negatives a positive).

A comparison is a subtraction that keeps only the flags: cmp a, b
computes a − b and throws the result away. "a < b" is then a question
about the flags: for unsigned numbers, did it borrow (C)? For signed
numbers, is N different from V?`,
				Diagram: `8-bit examples       result    Z N C V
 127 + 1             10000000  0 1 0 1   signed overflow: 127 + 1 ≠ −128
 255 + 1             00000000  1 0 1 0   unsigned carry: 255 + 1 ≠ 0
   5 − 7             11111110  0 1 1 0   borrow; as signed it is −2, fine
−128 − 1             01111111  0 0 0 1   signed overflow: gives +127`,
				UnderTheHood: `The same comparison instruction serves signed and unsigned code; only
the flag test after it differs. gcc -O2 compiles

  int lt_s(int a, int b)           { return a < b; }
  int lt_u(unsigned a, unsigned b) { return a < b; }

to

  cmp   edi, esi        cmp   edi, esi
  setl  al              setb  al

setl ("less") tests N ≠ V, the signed rule; setb ("below") tests the
carry flag, the unsigned rule. That is why comparing a signed with an
unsigned number in C is a trap: the compiler converts the signed one,
and −1 < 1u is false.

The overflow flag is also there for you to read. gcc's
__builtin_add_overflow(a, b, &r) compiles to an add followed by setno,
"set if no overflow", one instruction that reads V: checking every
signed addition costs almost nothing.`,
				MentalModel: "Subtract by adding the complement; compare by subtracting and reading the flags.",
				TryIt:       "In C, print -1 < 1u, (unsigned)-1, and (int8_t)128, and explain each result in terms of two's complement. Then compile both lt_s and lt_u with gcc -O2 -S.",
				Analogy: `A car's odometer: roll it back one from 000000 and it shows 999999. Two's
complement treats that 999999 as −1: counting down from zero wraps
around to the top. The carry flag is the odometer telling you it wrapped;
the overflow flag is it telling you a signed reading jumped from the
largest positive to the largest negative.`,
				Example: `Every if, for and while in compiled code ends in a compare and a jump
that reads these flags. Integer-overflow bugs are a whole class of
security holes: a size computed as count × size that wraps around
allocates a small buffer that is then overrun, which is why calloc
checks the multiplication for you.`,
				Exercises: trio(
					"In 8-bit two's complement, write 5, −5, −1 and −128 in binary. What do 01111111 + 1 and 11111111 + 1 give, and which flags do they set?",
					"00000101, 11111011, 11111111, 10000000. 127 + 1 = 10000000 sets N and V; 255 + 1 = 00000000 sets Z and C.",
					"In C, write uint8_t add8(uint8_t a, uint8_t b, int *c, int *v) and sub8 the same way, computing the flags from the bits (not by testing the answer in a wider type), and check them against a wider-type calculation for all 65,536 pairs.",
					"Carry: the 9th bit of a + b. Overflow: a and b have the same sign bit and the result's sign bit differs. sub8 is add8(a, ~b) with a carry-in of 1, and its borrow is the inverse of that carry.",
					"Find an integer-overflow bug: write a function that allocates count elements of size bytes with malloc(count * size), and show that for count = 2⁶¹ + 1 and size = 8 (with 64-bit size_t) the product wraps around to just 8 bytes. Fix it with __builtin_mul_overflow, or with calloc.",
					"(2⁶¹ + 1) × 8 = 2⁶⁴ + 8, which wraps to 8. Print count * size as an unsigned 64-bit number to see it. The check is: if (__builtin_mul_overflow(count, size, &bytes)) return NULL;",
				),
			},
			{
				Name:    "Memory: Latches, Registers & RAM",
				Summary: "How circuits remember.",
				Body: `Combinational circuits forget as soon as their inputs change. Feed a
gate's output back into its input and you get a latch, which holds one
bit. Two NOR gates, each feeding the other, make an SR latch: S sets
the stored bit to 1, R resets it to 0, and with both at 0 it holds
whatever it last stored. A flip-flop updates only on a clock tick,
keeping the whole chip in lockstep. Groups of flip-flops form registers
(for example, 64 bits).

SRAM (used in CPU caches and registers) keeps each bit in a small
flip-flop circuit of about six transistors: fast but large. DRAM (main
memory) stores each bit as charge in one transistor and a tiny
capacitor, which leaks, so every row must be refreshed every 64
milliseconds: dense and cheap, but slower. Flash memory (in SSDs) traps
charge in an insulated gate, so it survives power-off.

Memory is addressed: a decoder turns an address into a single selected
row, and the bits of that row are read out or written. That is why
reading any address takes the same time in principle ("random access").
In practice the distance matters a great deal, as the next section
measures.`,
				Diagram: `SR latch: two NOR gates, each fed by the other's output

  Q  = NOR(R, nQ)          S R │ Q
  nQ = NOR(S, Q)           0 0 │ keep (memory)
                           1 0 │ 1    (set)
  nQ is "not Q"            0 1 │ 0    (reset)
                           1 1 │ not allowed`,
				UnderTheHood: `The speed gap between SRAM and DRAM is measurable from C. A program
that follows a chain of pointers through a shuffled array (so the CPU
cannot guess the next address) timed each load on a test machine:

  working set     time per load    where the data lives
  16 KiB              1.6 ns       L1 cache (SRAM, next to the core)
  256 KiB             4.0 ns       L2 cache (SRAM)
  4 MiB              28.5 ns       L3 cache (SRAM, shared)
  256 MiB           207 ns         DRAM, across the memory bus

Over a hundred times slower from L1 to DRAM, for the same instruction.
That gap, and the caches built to hide it, are the subject of The Memory
Hierarchy & Locality in Stage 5.`,
				MentalModel: "Feedback turns logic into memory, and the clock keeps it in step.",
				TryIt:       "Build an SR latch from two NOR gates in a simulator and watch it hold its value. Then write the pointer-chasing program in C and measure your own machine's L1, L2, L3 and DRAM latencies.",
				Analogy: `A light switch stays where you put it after you let go (a latch), while
a doorbell only rings while you hold the button (combinational). DRAM is
like a leaky bucket that must be topped up regularly, or the water (your
data) drains away.`,
				Example: `When your laptop sleeps, it keeps powering the RAM so it can wake
instantly. When it hibernates, it copies RAM to the SSD, because RAM
forgets without power. The "Rowhammer" attack flips bits in DRAM by
rapidly reading neighbouring rows.`,
				Exercises: trio(
					"Why does an unsaved document vanish in a power cut, while files on the SSD survive? Use the words volatile and non-volatile.",
					"RAM is volatile: DRAM's capacitors lose their charge within milliseconds without power and refresh. Flash is non-volatile: its trapped charge stays for years.",
					"In C, simulate an SR latch: two NOR gates updated in turn (first nQ = NOR(S, Q), then Q = NOR(R, nQ)) until the outputs stop changing, and show that set, reset and hold behave correctly over a sequence of inputs.",
					"int nor(int a, int b) { return !(a | b); } and loop until neither output changed in a full pass. Start from Q = 0, nQ = 1.",
					`Laptops advertise "16 GB RAM, 512 GB SSD". Explain to a friend what each one is for, then measure how much RAM your browser uses with 1, 10 and 30 tabs open.`,
					"Use Task Manager (Windows), Activity Monitor (macOS) or htop (Linux).",
				),
			},
			{
				Name:    "Finite State Machines & Sequential Logic",
				Summary: "A system that is always in one of a few states, and moves between them as inputs arrive.",
				Body: `A finite state machine (FSM) has a fixed set of states, a starting
state, and a transition rule: given the current state and an input,
which state comes next. Some states may be accepting (the input so far
is valid) or produce outputs.

In hardware, an FSM is a register that holds the current state plus a
combinational circuit that computes the next state from it and the
inputs; on every clock tick the register loads the next state. A
counter is the simplest: a register and an adder that adds 1. A traffic
light controller, a vending machine and the control unit at the heart
of a CPU are all FSMs.

In software, FSMs are everywhere input arrives one piece at a time:
parsers, protocol handlers, games and user interfaces. The clean way to
write one in C is a state variable and a switch:

  enum state { START, SIGN, DIGITS, ERROR };
  enum state s = START;
  for (const char *p = text; *p != '\0'; p++) {
      int digit = isdigit((unsigned char)*p);
      switch (s) {
      case START:  s = *p == '-' ? SIGN : digit ? DIGITS : ERROR; break;
      case SIGN:   s = digit ? DIGITS : ERROR; break;
      case DIGITS: s = digit ? DIGITS : ERROR; break;
      case ERROR:  break;
      }
  }
  int is_integer = (s == DIGITS);      // the accepting state

Drawing the states as circles and the transitions as arrows, before
writing code, is the fastest way to find the cases you forgot.`,
				Diagram: `an FSM that accepts integers such as 42 or -7

             '-'               digit
  START ──────────▶ SIGN ──────────▶ DIGITS ──┐
    │                                 ▲  ▲    │ digit
    └──────────── digit ──────────────┘  └────┘
  any other character, from any state ──▶ ERROR, which it never leaves
  the input is an integer if it ends in DIGITS`,
				UnderTheHood: `A switch over a state variable compiles to a jump table, the software
version of the hardware's next-state logic. For a seven-state switch,
gcc -O2 produced:

  cmp    edi, 6
  ja     .L15                          ; state out of range
  lea    rdx, .L17[rip]                ; the table of jump offsets
  mov    edi, edi
  movsx  rax, DWORD PTR [rdx+rdi*4]    ; this state's entry
  add    rax, rdx
  notrack jmp rax                      ; go straight to its case

so moving to the right case takes the same few instructions whether
there are 7 states or 700: no chain of comparisons.

TCP, the protocol under the web, is specified as a state machine
(CLOSED, LISTEN, SYN-SENT, ESTABLISHED, FIN-WAIT and others), and on
Linux, ss -t shows each connection's current state. Regular expressions
are compiled into state machines too (Stage 11).`,
				MentalModel: "Draw the states and arrows first; the code is then a switch that follows the drawing.",
				TryIt:       "Draw the FSM for a turnstile (states locked and unlocked, inputs coin and push), then write it in C as a switch, reading c and p characters from standard input and printing the state after each.",
				Analogy: `A turnstile at a station: it is locked until you insert a coin, then
unlocked until you push through, then locked again. Pushing when it is
locked does nothing; a second coin when it is unlocked is wasted. Two
states, two inputs, and a rule for every combination.`,
				Example: `Traffic lights, lifts, washing machines and vending machines are run by
state machines. Network protocols (TCP), text parsers, the lexer at the
front of every compiler, game characters' behaviours and the "loading,
loaded, error" states of an app's screen are all FSMs.`,
				Exercises: trio(
					"Draw the FSM for a traffic light that cycles red → red-and-amber → green → amber → red on each tick, with a pedestrian button that, when pressed during green, moves to amber on the next tick instead of waiting.",
					"Four states in a cycle; the button only changes the transition out of green. Write the transition table first: one row per (state, input).",
					"In C, write a vending machine FSM: a drink costs 70 cents; inputs are coins of 10, 20 and 50 and a cancel key; it dispenses and gives change when 70 or more is inserted, and refunds on cancel. Read inputs from standard input and print every state change.",
					"The state is the amount inserted so far (0 to 60, in tens); each coin moves to a higher state, or to dispense-and-change.",
					"Write a C program that strips comments from C source code with a state machine: states for normal code, inside a string, inside a character literal, after a /, inside a /* */ comment and inside a // comment. Test it on a file containing \"/* not a comment */\" in a string.",
					"The tricky transitions: a backslash inside a string escapes the next character, and a * inside a block comment only ends it when followed by /.",
				),
			},
			{
				Name:    "The CPU: Datapath & Instruction Cycle",
				Summary: "Fetch, decode, execute, billions of times per second.",
				Body: `A CPU repeats one cycle. It fetches the next instruction from memory (at
the address held in the program counter), decodes it, executes it in the
ALU (arithmetic logic unit) using registers, and writes back the result.

The datapath is the hardware the data flows through: the register file,
the ALU, the connection to memory, and the multiplexers that choose which
register feeds which input. The control unit is a finite state machine
that, for each instruction, sets those multiplexers and tells the ALU
which operation to do.

Instructions are just numbers, defined by the instruction set
architecture (ISA), such as x86-64, ARM or RISC-V. Branch instructions
change the program counter, which is how ifs and loops exist in
hardware: a loop is a compare that sets flags, and a jump back that
happens only if a flag says so. The clock (for example 3 GHz, or 3
billion ticks a second) sets the pace, and modern cores complete several
instructions per tick.

An emulator is this cycle written in software: an array for memory, a
few variables for registers, and a loop that fetches, decodes with a
switch, and executes.`,
				Diagram: `   ┌───────┐    ┌────────┐    ┌─────────┐    ┌───────────┐
┌─▶│ FETCH │───▶│ DECODE │───▶│ EXECUTE │───▶│ WRITEBACK │───┐
│  └───────┘    └────────┘    └─────────┘    └───────────┘   │
│   PC → memory   which op?     ALU / memory   result → reg  │
└──────────────── PC = next instruction ◀────────────────────┘`,
				UnderTheHood: `The instructions really are bytes in your program file. gcc -O2 compiled
int add(int a, int b) { return a + b; } and objdump -d shows:

  f3 0f 1e fa          endbr64
  8d 04 37             lea    eax, [rdi+rsi*1]
  c3                   ret

The CPU fetches f3 0f 1e fa (a marker for control-flow protection,
which does nothing else), then 8d 04 37: opcode 8d (lea, "load effective
address", used here simply as a fast add), and 04 37 encode the
registers. The arguments arrive in edi and esi, and the result goes in
eax, as the calling convention (Stage 5) says. Then c3, ret: fetch the
return address from the stack and continue from there.

A modern core does not run these strictly one at a time: it fetches and
decodes several per cycle, splits them into simpler internal operations,
and executes them out of order when their inputs are ready. The
programmer-visible model, one instruction after another, is kept as a
promise, not as a description of the hardware.`,
				MentalModel: "A CPU is a loop that reads a number, does what the number says, and moves on.",
				TryIt:       "Compile a small C function with gcc -O2 -c, look at it with objdump -d -M intel, and find each instruction's bytes. Then change the function and see which bytes change.",
				Analogy: `A cook following a recipe card: read the next step (fetch), understand it
(decode), do it (execute), put the result on the counter (write back),
then move on to the next step, unless the step says "go back to step 3".`,
				Example: `In 2020 Apple moved the Mac from Intel (x86) chips to its own ARM chips:
a change of ISA. That is why apps had to be recompiled, or translated by
Rosetta 2. Phones, Raspberry Pis and most embedded devices run ARM, and
RISC-V is an open ISA that anyone can implement.`,
				Exercises: trio(
					"A 3 GHz CPU completes on average 2 instructions per clock tick. How many instructions per second is that, and how long is one tick?",
					"6 billion per second; one tick is 1 / 3,000,000,000 s ≈ 0.33 ns.",
					"In C, write an emulator for a tiny CPU with 4 registers and the instructions LOAD r value, ADD rd rs, SUB rd rs, JZ r target, JMP target, PRINT r and HALT, read as text. Then write a program for it that counts down from 10.",
					"Parse the program into an array of structs first; then a loop over a program counter, with a switch on the opcode. A jump sets pc instead of adding 1.",
					"Look up the CPU in your laptop or phone: its ISA, core count, clock speed and cache sizes. Explain which of those numbers matter most for gaming, video editing and web browsing.",
					"lscpu on Linux, About This Mac on macOS, Task Manager → Performance on Windows.",
				),
			},
			{
				Name:    "Instruction Sets & Machine Code",
				Summary: "How an instruction is packed into bits, and why x86 and RISC-V pack them so differently.",
				Body: `An instruction set architecture defines the instructions, the
registers, and the exact bit patterns that encode each instruction. The
assembler turns text such as addi x1, x0, 5 into those bits; a
disassembler (objdump) turns them back.

RISC-V, an open ISA, shows the idea cleanly. Every instruction is 32
bits, with fields at fixed positions:

  bits:   31        20 19   15 14  12 11    7 6       0
  I-type: [ imm 11:0  ][  rs1 ][funct3][  rd  ][ opcode ]
  R-type: [funct7][rs2][  rs1 ][funct3][  rd  ][ opcode ]

The opcode (bits 0–6) says the format: 0010011 for arithmetic with an
immediate (I-type), 0110011 for register-to-register (R-type). funct3,
and for R-type funct7, pick the operation (add or sub, and, or, xor).
rd is the destination register and rs1, rs2 the sources, 5 bits each
for 32 registers. The immediate is a 12-bit two's-complement number, so
decoding it means sign-extending bit 31: -1 is all ones.

Decoding is shifts and masks:

  uint32_t opcode = w & 0x7f;
  uint32_t rd     = (w >> 7)  & 0x1f;
  uint32_t funct3 = (w >> 12) & 0x7;
  uint32_t rs1    = (w >> 15) & 0x1f;
  int32_t  imm    = (int32_t)w >> 20;    // sign-extends (see below)

Shifting a negative int right is arithmetic (it copies the sign bit) in
gcc and clang, though the C standard leaves it to the implementation;
the portable alternative is to shift unsigned and subtract 4096 when bit
11 of the field is set.

RISC designs (RISC-V, ARM) use simple, fixed-size instructions that are
easy to decode; x86, a CISC design from 1978 extended ever since, uses
variable-length instructions from 1 to 15 bytes.`,
				Diagram: `addi x1, x0, 5  =  0x00500093

  000000000101  00000  000  00001  0010011
  imm = 5       rs1=x0 add  rd=x1  opcode: I-type arithmetic`,
				UnderTheHood: `llvm-mc, LLVM's assembler, shows the real encodings. For RISC-V:

  addi x1, x0, 5      93 00 50 00    = 0x00500093
  addi x2, x1, -1     13 81 f0 ff    = 0xfff08113
  add  x3, x1, x2     b3 81 20 00    = 0x002081b3
  sub  x3, x1, x2     b3 81 20 40    = 0x402081b3

The bytes come out lowest first (RISC-V is little-endian), so the word
is read right to left. add and sub differ in a single bit, bit 30 (the
funct7 field), and -1's immediate is twelve 1 bits at the top of the
word, the f f f in 0xfff08113.

x86-64 is the opposite. objdump of three gcc-compiled instructions:

  c3                               ret                     1 byte
  8d 04 37                         lea  eax, [rdi+rsi*1]   3 bytes
  48 b8 88 77 66 55 44 33 22 11    movabs rax, 0x1122334455667788
                                                           10 bytes

A decoder cannot even know where the next instruction starts until it
has decoded this one, which is why x86 chips spend so many transistors
on decoding, while every RISC-V instruction starts at a multiple of 4.`,
				MentalModel: "An instruction is a number with fields; decoding is shifting and masking.",
				TryIt:       "Decode 0x00500093 and 0x402081b3 by hand into their fields, then check yourself: write each field out in binary and compare with the diagram.",
				Analogy: `A postcode or a vehicle registration number: positions in the string
mean different things (region, district, street), so you can decode one
by cutting it at fixed places. RISC-V instructions are like that. x86
instructions are more like a sentence, whose meaning you only know once
you have read up to the full stop.`,
				Example: `Apple's M-series chips decode fixed-width ARM instructions, which helps
them decode eight or more per cycle. Emulators such as QEMU, JIT
compilers in browsers and the JVM, debuggers, and malware analysts all
encode or decode machine code. RISC-V's openness lets universities and
companies design their own CPUs without a licence.`,
				Exercises: trio(
					"Split 0x002081b3 into its fields: opcode (bits 0–6), rd (7–11), funct3 (12–14), rs1 (15–19), rs2 (20–24), funct7 (25–31). Which instruction is it?",
					"opcode 0110011 (R-type), rd 3, funct3 000, rs1 1, rs2 2, funct7 0000000: add x3, x1, x2.",
					"In C, write a decoder for the RISC-V instructions add, sub, and, or, xor, addi, andi, ori and xori, given 32-bit words in hex, printing them like addi x1, x0, 5. Test it with the encodings above, including negative immediates.",
					"Check the opcode first, then funct3, then (for R-type) funct7. Get the immediate with an arithmetic right shift of the word as int32_t, which keeps the sign.",
					"Write the reverse: a tiny assembler in C for the same nine instructions, reading lines like add x3, x1, x2 and printing the 32-bit words in hex. Check that your decoder turns its output back into the same lines, for a whole program.",
					"Encoding is the decoder backwards: word = funct7 << 25 | rs2 << 20 | rs1 << 15 | funct3 << 12 | rd << 7 | opcode, with (imm & 0xfff) << 20 for I-type.",
				),
			},
			{
				Name:    "Pipelining, Branch Prediction & Parallelism",
				Summary: "The tricks that make modern CPUs fast.",
				Body: `Pipelining overlaps the stages of instructions like an assembly line:
while one instruction executes, the next is decoded and a third is
fetched. Branch prediction guesses which way an if will go, to keep the
pipeline full; a wrong guess throws away the work in flight and costs
roughly 15 to 20 cycles. Out-of-order execution runs independent
instructions early, and caches (covered again in Stage 5) hide slow
memory.

From the mid-2000s, clock speeds stopped rising because of heat, so
chips added more cores instead. Software must run in parallel to use
them, and Amdahl's law says the part that must run serially limits the
overall speed-up: if a fraction s of the work is serial, no number of
cores can make it more than 1/s times faster.

Compilers help too. They remove branches where they can, replacing a
short if with a conditional move (cmov), and they vectorise loops with
SIMD instructions that process 16, 32 or 64 bytes at once. Whether your
code has a branch at all can depend on the optimisation flags, which is
why performance claims must be measured, not assumed.`,
				Diagram: `time →      1    2    3    4    5    6
instr 1:   [F]  [D]  [E]  [W]
instr 2:        [F]  [D]  [E]  [W]
instr 3:             [F]  [D]  [E]  [W]     one instruction finishes per tick`,
				UnderTheHood: `The famous experiment, re-run on a test machine: sum the bytes ≥ 128 of
a 32 MB array, first in random order, then after sorting it.

  compiled with                                 random    sorted
  gcc -O2 -fno-if-conversion -fno-tree-vectorize  148 ms     20 ms
  gcc -O0                                         264 ms     89 ms
  gcc -O2                                          14 ms     15 ms

With a real branch, random data makes the predictor wrong about half
the time, and the loop is seven times slower than on sorted data, where
the branch goes the same way for millions of iterations. But plain
gcc -O2 removes the branch altogether: objdump shows pcmpgtb, a SIMD
instruction that compares 16 bytes at once and turns the if into
arithmetic on masks. No branch, nothing to mispredict, and the order of
the data no longer matters.`,
				MentalModel: "Keep every part of the chip busy: overlap, predict and parallelise, and measure which one the compiler already did.",
				TryIt:       "Run the sorted-versus-random experiment in C with -O0, -O2, and -O2 -fno-if-conversion -fno-tree-vectorize, and use objdump to find the branch (or its absence) in each.",
				Analogy: `A laundromat: washing, drying and folding different loads at the same
time is far faster than finishing each load before starting the next.
More cores are like more washing machines, but if one step has to happen
alone, extra machines do not help (Amdahl's law).`,
				Example: `One of the most upvoted questions on Stack Overflow asks why processing
a sorted array is faster than an unsorted one. The answer is branch
prediction. The Spectre and Meltdown vulnerabilities (2018) abused
speculative execution to leak secrets.`,
				Exercises: trio(
					"A program is 90% parallelisable. What is the maximum speed-up with 10 cores? With unlimited cores?",
					"Amdahl's law: 1 / (0.1 + 0.9/N) gives about 5.3× for 10 cores and 10× at most.",
					"Run the sorted-versus-unsorted branch experiment in C and report the times with three sets of flags: -O0, -O2, and -O2 -fno-if-conversion -fno-tree-vectorize. Explain each result from the assembly.",
					"Use an array of 32 million bytes and repeat each run several times. In the objdump output, look for a conditional jump inside the loop, a cmov, or SIMD instructions such as pcmpgtb.",
					"Take a slow batch job (for example converting 500 images, or hashing 500 files) and split it across all your CPU cores with POSIX threads or by running several processes. Measure the speed-up and explain why it is less than the number of cores.",
					"Give each thread a contiguous slice of the files. Look for the serial parts: listing the files, reading from disk, combining the results.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "Nand2Tetris", "https://www.nand2tetris.org/", "Build a whole computer from NAND gates up; perfect for this stage."},
			{"Book", "Code: The Hidden Language of Computer Hardware and Software", "https://www.codehiddenlanguage.com/", "Charles Petzold's gentle, story-like path from Morse code to CPUs."},
			{"Video", "Ben Eater: building an 8-bit computer", "https://eater.net/8bit", "Watch a CPU get built on breadboards, wire by wire."},
			{"Tool", "Digital (logic simulator)", "https://github.com/hneemann/Digital", "A free simulator for designing and testing circuits."},
			{"Tool", "Logisim Evolution", "https://github.com/logisim-evolution/logisim-evolution", "Another popular free logic simulator."},
		},
		Blueprints: []Blueprint{
			{"8-bit CPU in a Simulator", "Design a small CPU (ALU, registers, program counter, control unit) in a logic simulator and run a program on it.",
				[]string{"ALU with add/sub", "Register file", "Program counter and memory", "Control logic", "Run a countdown program"}},
			{"CPU Emulator (CHIP-8)", "Write an emulator for the classic CHIP-8 virtual machine and run old games on it.",
				[]string{"Fetch–decode–execute loop", "Registers and memory", "Graphics output", "Keyboard input", "Timers and sound"}},
			{"Binary & Encoding Explorer", "A tool that shows any value as binary, hex, ASCII/UTF-8, a signed or unsigned integer, and a float.",
				[]string{"Integer conversions", "Text encodings", "Float bit fields", "Hex dump of files"}},
		},
		Quiz: []Question{
			{"What is 0x1F in decimal?", "31: 1 × 16 + 15."},
			{"Why is NAND called a universal gate?", "Any logic circuit, including NOT, AND and OR, can be built from NAND gates alone."},
			{"What are the four steps of the instruction cycle?", "Fetch, decode, execute and write back."},
			{"Why did CPUs switch from faster clocks to more cores?", "Higher clock speeds produced too much heat and used too much power, so manufacturers added parallel cores instead."},
			{"How many bytes does UTF-8 use for 'A', for é, and for an emoji such as 😀?", "One (0x41), two (c3 a9) and four (f0 9f 98 80)."},
			{"How does a CPU compute a − b with an adder?", "It adds a, the inverted bits of b, and a carry-in of 1: a + (NOT b) + 1, which is a + (−b) in two's complement."},
			{"What is the difference between the carry flag and the overflow flag?", "Carry means an unsigned result did not fit (a carry or borrow out of the top bit); overflow means a signed result did not fit (for example, two positives gave a negative)."},
			{"Why does DRAM need refreshing, and SRAM not?", "DRAM stores each bit as charge in a capacitor that leaks away within milliseconds; SRAM holds each bit in a flip-flop that keeps its state as long as it has power."},
			{"What makes up a finite state machine?", "A fixed set of states, a starting state, and a transition rule giving the next state for each state and input (plus, optionally, accepting states or outputs)."},
			{"How does a RISC-V decoder find the destination register of an instruction?", "It is always bits 7 to 11: (w >> 7) & 0x1f."},
			{"Why might sorting an array make no difference to the speed of an if inside a loop?", "The compiler may have removed the branch, with a conditional move or SIMD instructions, so there is nothing to mispredict."},
		},
	},
}
