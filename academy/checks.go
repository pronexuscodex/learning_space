package main

// Checked exercises: C programs the learner writes in their own editor and
// the academy compiles and runs for real with the C compiler on their
// machine (runner.go), against test cases written here. A pass is evidence,
// not a tick: it records the compiler, the time, the cases passed and how
// many hints had been revealed (checklog.go).
//
// Each check names the concept it demonstrates. Hints are revealed one at a
// time and never contain the solution.

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Check is one checked exercise.
type Check struct {
	ID      string // stable: used in the log and by "academy -check ID"
	Stage   int
	Concept string // the Concept.Name it demonstrates
	Title   string
	Task    string // what to write
	File    string // the learner's file, in the stage's folder of the labs
	Starter string // the file's starting content (a skeleton, or a buggy program to fix)

	// Harness, when set, is a main() that is compiled together with the
	// learner's file, which then holds functions only. Each case passes
	// its Args to it.
	Harness string

	// Extra are the learner's other files for this check, beside File
	// (a header, another .c file), each created with its starter.
	Extra []CheckFile

	Project    bool          // a stage project, joining several concepts (Concept is then "")
	Leaks      bool          // also fail on memory never freed (where the leak checker works)
	Timeout    time.Duration // per case, when the default 2 s is too short
	Flags      []string      // extra compiler flags (such as -pthread)
	BuildHints []BuildHint   // explanations for compiler messages this check tends to cause

	Cases []CheckCase
	Hints []string
}

// CheckFile is one more file the learner writes for a check.
type CheckFile struct {
	Name    string
	Starter string
}

// BuildHint adds Note to the build report when the compiler's output
// contains Match.
type BuildHint struct {
	Match, Note string
}

// CheckCase is one test: run the program with Args and Stdin, then compare
// standard output (trailing spaces and blank lines aside) and the exit code.
type CheckCase struct {
	Name      string
	Args      []string
	Stdin     string
	Want      string
	WantExit  int
	Stderr    bool              // the program must also print a message to standard error
	Files     map[string]string // files in the folder the program runs in (paths may have folders)
	WantFiles map[string]string // files that must hold exactly this afterwards
	Explain   string            // shown when this case fails: the likely cause

	// Gen, when set, makes Stdin and Want when the case is first run.
	// Big generated inputs would otherwise be built every time the
	// program starts. Wrap it in generated so it runs once.
	Gen func() (stdin, want string)
}

// generated memoises gen for CheckCase.Gen.
func generated(gen func() (stdin, want string)) func() (string, string) {
	var once sync.Once
	var stdin, want string
	return func() (string, string) {
		once.Do(func() { stdin, want = gen() })
		return stdin, want
	}
}

// resolved returns c with Stdin and Want made by Gen, if it has one.
func (c CheckCase) resolved() CheckCase {
	if c.Gen != nil {
		c.Stdin, c.Want = c.Gen()
		c.Gen = nil
	}
	return c
}

// checks are all checked exercises, in teaching order.
var checks = []Check{
	{
		ID: "hello", Stage: 1, Concept: "Your First C Program: Source, Compiler, Executable",
		Title: "Hello, world",
		Task:  "Write a program that prints exactly: Hello, world! followed by a newline, and returns 0 from main.",
		File:  "hello.c",
		Starter: `/* Checked exercise: hello
 * Print exactly "Hello, world!" and a newline, then return 0 from main.
 * Save, then run the check (Study Hall, or: academy -check hello).
 */
`,
		Cases: []CheckCase{{Name: "prints the greeting", Want: "Hello, world!\n",
			Explain: "Check the spelling, the comma, the exclamation mark and the \\n at the end of the string."}},
		Hints: []string{
			"Every C program starts at int main(void) { ... }.",
			"#include <stdio.h> gives you printf. A string ends its line with \\n.",
		},
	},
	{
		ID: "types", Stage: 1, Concept: "Values, Types & Variables",
		Title: "Predict, then print",
		Task:  "Before you run anything, write down what you expect each of these to print. Then write a program that prints them, one per line, with the format given: 7 / 2 (%d), 7 % 2 (%d), 7.0 / 2 (%.1f), (int)3.99 (%d), 'A' + 1 (%d) and sizeof(int) * 8 (%zu). Compare the output with your predictions.",
		File:  "types.c",
		Starter: `/* Checked exercise: types
 * Predict first! Then print, one per line:
 *   7 / 2          with %d
 *   7 % 2          with %d
 *   7.0 / 2        with %.1f
 *   (int)3.99      with %d
 *   'A' + 1        with %d
 *   sizeof(int) * 8   with %zu (the number of bits in an int)
 */
`,
		Cases: []CheckCase{{Name: "six values", Want: "3\n1\n3.5\n3\n66\n32\n",
			Explain: "Line 1 is 3 because int / int throws the fraction away. A cast cuts the fraction off rather than rounding, and 'A' is the number 65."}},
		Hints: []string{
			"Each line is one printf with the format shown, such as printf(\"%d\\n\", 7 / 2);.",
			"sizeof gives a size_t, which is why it needs %zu rather than %d.",
		},
	},
	{
		ID: "fizzbuzz", Stage: 1, Concept: "Control Flow: Decisions & Loops",
		Title: "FizzBuzz",
		Task:  "Print the numbers 1 to 100, one per line, but print Fizz for multiples of 3, Buzz for multiples of 5 and FizzBuzz for multiples of both.",
		File:  "fizzbuzz.c",
		Starter: `/* Checked exercise: fizzbuzz
 * Print 1 to 100, one per line: Fizz for multiples of 3, Buzz for
 * multiples of 5, FizzBuzz for multiples of both.
 */
`,
		Cases: []CheckCase{{Name: "1 to 100", Want: fizzBuzzWant(100),
			Explain: "Look at the first line that differs. If 15 printed Fizz, test \"both\" before testing 3 or 5 alone. If it starts at 0 or stops at 99, check the loop bounds."}},
		Hints: []string{
			"a % b is the remainder of a divided by b, so n % 3 == 0 means n is a multiple of 3.",
			"Order matters: the first matching if wins, so check the most specific case (both 3 and 5) first.",
			"for (int i = 1; i <= 100; i++) runs exactly 100 times, from 1 to 100.",
		},
	},
	{
		ID: "temperature", Stage: 1, Concept: "Functions & Decomposition",
		Title: "Temperature conversion functions",
		Task:  "Write two functions, double c_to_f(double c) and double f_to_c(double f), and no main: the checker supplies its own main and calls them. F = C × 9/5 + 32.",
		File:  "temperature.c",
		Starter: `/* Checked exercise: temperature
 * Write these two functions. Do not write main: the checker has its own.
 *   double c_to_f(double c);   Celsius to Fahrenheit: F = C * 9/5 + 32
 *   double f_to_c(double f);   Fahrenheit to Celsius
 */
`,
		Harness: `#include <stdio.h>
#include <string.h>

double c_to_f(double c);
double f_to_c(double f);

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    if (strcmp(argv[1], "round-trip") == 0) {
        for (double c = -50.0; c <= 50.0; c += 0.5) {
            double back = f_to_c(c_to_f(c));
            double d = back > c ? back - c : c - back;
            if (d > 1e-9) { printf("%.2f came back as %.6f\n", c, back); return 0; }
        }
        printf("ok\n");
        return 0;
    }
    double x;
    if (sscanf(argv[2], "%lf", &x) != 1) return 2;
    printf("%.2f\n", strcmp(argv[1], "c_to_f") == 0 ? c_to_f(x) : f_to_c(x));
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "c_to_f(0)", Args: []string{"c_to_f", "0"}, Want: "32.00\n"},
			{Name: "c_to_f(100)", Args: []string{"c_to_f", "100"}, Want: "212.00\n",
				Explain: "132.00 means 9 / 5 was computed in integers, which gives 1. Write 9.0 / 5.0."},
			{Name: "c_to_f(-40)", Args: []string{"c_to_f", "-40"}, Want: "-40.00\n"},
			{Name: "c_to_f(37)", Args: []string{"c_to_f", "37"}, Want: "98.60\n"},
			{Name: "f_to_c(212)", Args: []string{"f_to_c", "212"}, Want: "100.00\n",
				Explain: "Subtract 32 first, then multiply by 5/9 (in floating point, 5.0 / 9.0)."},
			{Name: "there and back", Args: []string{"round-trip"}, Want: "ok\n",
				Explain: "f_to_c(c_to_f(c)) should give c back for every c: the two formulas must be exact inverses."},
		},
		Hints: []string{
			"A function definition looks like its prototype with a body: double c_to_f(double c) { return ...; }",
			"In C, 9 / 5 is 1: both are ints. Make one of them a double (9.0) to keep the fraction.",
			"For f_to_c, undo the steps of c_to_f in reverse order: first subtract 32, then undo the multiplication.",
		},
	},
	{
		ID: "swap", Stage: 1, Concept: "Memory, Addresses & Pointers",
		Title: "Swap through pointers",
		Task:  "Write void swap(int *a, int *b), which exchanges the two ints that a and b point to. No main: the checker calls it.",
		File:  "swap.c",
		Starter: `/* Checked exercise: swap
 * Write void swap(int *a, int *b): afterwards *a holds the old *b and *b
 * holds the old *a. Do not write main: the checker has its own.
 */
`,
		Harness: `#include <stdio.h>
#include <string.h>

void swap(int *a, int *b);

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    int x = 3, y = 7;
    if (strcmp(argv[1], "basic") == 0) { swap(&x, &y); printf("%d %d\n", x, y); }
    else if (strcmp(argv[1], "equal") == 0) { x = 5; y = 5; swap(&x, &y); printf("%d %d\n", x, y); }
    else if (strcmp(argv[1], "negative") == 0) { x = -1; y = 2; swap(&x, &y); printf("%d %d\n", x, y); }
    else if (strcmp(argv[1], "twice") == 0) { swap(&x, &y); swap(&x, &y); printf("%d %d\n", x, y); }
    else if (strcmp(argv[1], "same") == 0) { x = 4; swap(&x, &x); printf("%d\n", x); }
    else return 2;
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "swap 3 and 7", Args: []string{"basic"}, Want: "7 3\n",
				Explain: "If nothing changed, the function swapped copies: assign through the pointers, *a = ..."},
			{Name: "equal values", Args: []string{"equal"}, Want: "5 5\n"},
			{Name: "negative numbers", Args: []string{"negative"}, Want: "2 -1\n"},
			{Name: "swapping twice restores", Args: []string{"twice"}, Want: "3 7\n"},
			{Name: "a variable with itself", Args: []string{"same"}, Want: "4\n",
				Explain: "0 here means the XOR or add/subtract trick: when a and b point to the same int, it destroys the value. Use a temporary variable."},
		},
		Hints: []string{
			"*a is the int that a points to; you can read it and assign to it.",
			"To exchange two values you need somewhere to keep one of them for a moment.",
		},
	},
	{
		ID: "letters", Stage: 1, Concept: "Collections: Lists, Dictionaries & Strings",
		Title: "Letter counts",
		Task:  "Read all of standard input and print how often each letter a–z occurs, ignoring case, one line per letter that occurs at least once, in alphabetical order, as: the letter, a space, the count. Ignore everything that is not an ASCII letter.",
		File:  "letters.c",
		Starter: `/* Checked exercise: letters
 * Read all of standard input. For each letter a-z that occurs (upper or
 * lower case), print one line: the letter, a space, the count. In
 * alphabetical order; ignore every other character.
 * Try it: echo "Hello, World" | ./letters
 */
`,
		Cases: []CheckCase{
			{Name: "Hello, World!", Stdin: "Hello, World!\n", Want: "d 1\ne 1\nh 1\nl 3\no 2\nr 1\nw 1\n",
				Explain: "Upper and lower case count as the same letter; punctuation and spaces are not counted."},
			{Name: "empty input", Stdin: "", Want: "",
				Explain: "With no input, print nothing at all."},
			{Name: "accents and digits are ignored", Stdin: "Café 42 naïve\n", Want: "a 2\nc 1\ne 1\nf 1\nn 1\nv 1\n",
				Explain: "é and ï are not ASCII letters: in UTF-8 each is two bytes above 127. Passing such a byte, stored in a plain char, to isalpha or tolower is undefined behaviour; use an int from getchar, or cast to unsigned char."},
			{Name: "a long input", Gen: generated(func() (string, string) { return strings.Repeat("ab", 100000), "a 100000\nb 100000\n" }),
				Explain: "Read character by character (getchar until EOF) instead of into a fixed-size buffer."},
		},
		Hints: []string{
			"getchar() returns the next byte as an int, or EOF at the end of input. Keep it in an int, not a char.",
			"An array int counts[26] = {0} has one counter per letter; c - 'a' turns 'a'..'z' into 0..25.",
			"tolower and isalpha (ctype.h) expect a value of unsigned char or EOF; bytes of an accented letter are not ASCII letters.",
		},
	},
	{
		ID: "scores", Stage: 1, Concept: "Structs: Grouping Related Data",
		Title: "Best score and average",
		Task:  "Read lines of the form NAME SCORE from standard input: a name of at most 31 letters, a whole score from 0 to 100, and at most 100 lines. Keep them in an array of struct student. Print best: NAME SCORE for the first student with the highest score, then average: and the mean score with one decimal. With no input at all, print no students.",
		File:  "scores.c",
		Starter: `/* Checked exercise: scores
 * Read lines "NAME SCORE" (name: at most 31 letters; score: 0 to 100;
 * at most 100 lines) into an array of struct student. Print:
 *   best: NAME SCORE      the first student with the highest score
 *   average: X.X          the mean, with one decimal
 * With no input at all, print: no students
 * Try it: printf 'ada 90\nbob 72\n' | ./scores
 */
`,
		Cases: []CheckCase{
			{Name: "three students", Stdin: "ada 90\nbob 72\ncy 85\n", Want: "best: ada 90\naverage: 82.3\n",
				Explain: "The average of 90, 72 and 85 is 82.33…: divide as doubles, (double)sum / count, and print with %.1f."},
			{Name: "a tie goes to the first", Stdin: "ann 80\nben 80\ncal 79\n", Want: "best: ann 80\naverage: 79.7\n",
				Explain: "Replace the best student only when a score is strictly higher (>, not >=)."},
			{Name: "one student with zero", Stdin: "zed 0\n", Want: "best: zed 0\naverage: 0.0\n",
				Explain: "Start the best from the first student read, not from a score of 0 or 1."},
			{Name: "no input", Stdin: "", Want: "no students\n",
				Explain: "Check the count before dividing by it."},
			{Name: "a 31-letter name", Stdin: "abcdefghijklmnopqrstuvwxyzabcde 100\n", Want: "best: abcdefghijklmnopqrstuvwxyzabcde 100\naverage: 100.0\n",
				Explain: "31 letters need 32 bytes with the '\\0'. A width in the format (%31s) stops scanf writing past the array."},
			{Name: "a hundred students", Stdin: hundredStudents(), Want: "best: s99 99\naverage: 49.5\n"},
		},
		Hints: []string{
			"scanf(\"%31s %d\", name, &score) == 2 reads one line's two fields; the 31 stops it writing past a 32-byte name.",
			"Keep a count of students read and the index of the best so far.",
			"The mean is a double: (double)sum / count.",
		},
	},
	{
		ID: "intvec", Stage: 1, Concept: "Dynamic Memory: malloc, free & Leaks",
		Title: "A growable array",
		Task:  "Write vec_init, vec_push and vec_free for the struct intvec in the starter file: a growable array of ints whose data lives on the heap and doubles in capacity when full. No main: the checker calls your functions, a million times in one case, and also fails if any memory is never freed.",
		File:  "intvec.c",
		Leaks: true,
		Starter: `/* Checked exercise: intvec
 * A growable array of ints. Keep the struct exactly as it is, write the
 * three functions, and do not write main: the checker has its own.
 */
#include <stddef.h>
#include <stdlib.h>

struct intvec {
    int *data;     /* the elements, on the heap (NULL while empty) */
    size_t len;    /* how many are in use */
    size_t cap;    /* how many fit before data must grow */
};

/* vec_init makes v empty, without allocating anything yet. */
void vec_init(struct intvec *v);

/* vec_push appends x, doubling the capacity when data is full.
 * It returns 0, or -1 if memory runs out (v is then unchanged). */
int vec_push(struct intvec *v, int x);

/* vec_free releases the memory and leaves v empty, as after vec_init. */
void vec_free(struct intvec *v);
`,
		Harness: `#include <stdio.h>
#include <string.h>

struct intvec { int *data; size_t len; size_t cap; };
void vec_init(struct intvec *v);
int vec_push(struct intvec *v, int x);
void vec_free(struct intvec *v);

/* scrub overwrites old stack memory, so no stale copy of a pointer can
 * hide a leak from the leak checker. */
static int scrub(void) {
    volatile char junk[8192];
    for (int i = 0; i < 8192; i++) junk[i] = 0;
    return junk[100];
}

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    struct intvec v;
    vec_init(&v);
    if (strcmp(argv[1], "empty") == 0) {
        printf("%zu %zu %s\n", v.len, v.cap, v.data == NULL ? "null" : "allocated");
    } else if (strcmp(argv[1], "three") == 0) {
        vec_push(&v, 10); vec_push(&v, 20); vec_push(&v, 30);
        printf("%zu:", v.len);
        for (size_t i = 0; i < v.len; i++) printf(" %d", v.data[i]);
        printf("\n");
    } else if (strcmp(argv[1], "million") == 0) {
        long long sum = 0;
        int bad = 0;
        for (int i = 0; i < 1000000; i++) bad |= vec_push(&v, i) != 0;
        for (size_t i = 0; i < v.len; i++) { sum += v.data[i]; bad |= v.data[i] != (int)i; }
        printf("%zu %lld %s\n", v.len, sum, !bad && v.cap >= v.len ? "ok" : "wrong");
    } else if (strcmp(argv[1], "free") == 0) {
        vec_push(&v, 5);
        vec_free(&v);
        printf("%zu %zu %s\n", v.len, v.cap, v.data == NULL ? "null" : "dangling");
    } else if (strcmp(argv[1], "reuse") == 0) {
        vec_push(&v, 1);
        vec_free(&v);
        vec_push(&v, 7);
        printf("%zu: %d\n", v.len, v.data[0]);
    } else if (strcmp(argv[1], "leak") == 0) {
        for (int i = 0; i < 100; i++) vec_push(&v, i);
        vec_free(&v);
        printf("freed%s\n", scrub() ? "?" : "");
    } else return 2;
    vec_free(&v);
    return 0;
}
`,
		Cases: []CheckCase{
			{Name: "starts empty", Args: []string{"empty"}, Want: "0 0 null\n",
				Explain: "vec_init should set data to NULL and len and cap to 0, and allocate nothing."},
			{Name: "three pushes", Args: []string{"three"}, Want: "3: 10 20 30\n"},
			{Name: "a million pushes", Args: []string{"million"}, Want: "1000000 499999500000 ok\n",
				Explain: "Grow with realloc when len reaches cap; start from a small capacity (such as 4) when cap is 0, then double it. Growing by one each time is very slow."},
			{Name: "free leaves it empty", Args: []string{"free"}, Want: "0 0 null\n",
				Explain: "After free(v->data), set v->data to NULL and both counts to 0, so nothing points at freed memory."},
			{Name: "usable after free", Args: []string{"reuse"}, Want: "1: 7\n"},
			{Name: "frees everything", Args: []string{"leak"}, Want: "freed\n",
				Explain: "Memory was still allocated when the program ended: vec_free must free(v->data). (Where the leak checker does not work, this case cannot catch it.)"},
		},
		Hints: []string{
			"realloc(NULL, n) behaves like malloc(n), so the first push needs no special case if cap starts at 0.",
			"Keep the old pointer until realloc succeeds: void *tmp = realloc(v->data, new_cap * sizeof *v->data); if (tmp == NULL) return -1;",
			"Every case calls vec_free at the end: if the leak checker complains, vec_free is not freeing data.",
		},
	},
	{
		ID: "safe_number", Stage: 1, Concept: "Safe Input & Undefined Behavior",
		Title: "Read a number safely",
		Task:  "Read one line from standard input. If it holds a whole number that fits in an int (spaces before or after are fine), print ok and the number, and return 0. Otherwise print nothing to standard output, print a message to standard error, and return 1.",
		File:  "safe_number.c",
		Starter: `/* Checked exercise: safe_number
 * Read one line. If it is a whole number that fits in an int (spaces
 * around it are fine), print "ok N" and return 0. Otherwise print a
 * message to stderr (not stdout) and return 1: empty input, letters,
 * "12abc", and numbers too big for an int are all errors.
 */
`,
		Cases: []CheckCase{
			{Name: "42", Stdin: "42\n", Want: "ok 42\n"},
			{Name: "-17", Stdin: "-17\n", Want: "ok -17\n"},
			{Name: "spaces around", Stdin: "  8  \n", Want: "ok 8\n",
				Explain: "strtol skips leading spaces; after the number, allow trailing spaces and the newline."},
			{Name: "the largest int", Stdin: "2147483647\n", Want: "ok 2147483647\n"},
			{Name: "letters", Stdin: "abc\n", WantExit: 1, Stderr: true,
				Explain: "strtol sets its end pointer to the start of the text when it finds no digits at all."},
			{Name: "number then letters", Stdin: "12abc\n", WantExit: 1, Stderr: true,
				Explain: "Check that only spaces and the newline remain after the number: strtol's end pointer tells you where it stopped."},
			{Name: "an empty line", Stdin: "\n", WantExit: 1, Stderr: true},
			{Name: "no input at all", Stdin: "", WantExit: 1, Stderr: true,
				Explain: "fgets returns NULL at the end of input: check its return value."},
			{Name: "one more than the largest int", Stdin: "2147483648\n", WantExit: 1, Stderr: true,
				Explain: "Convert with strtol into a long, then check it against INT_MIN and INT_MAX (limits.h) before narrowing to int."},
			{Name: "far too big", Stdin: "99999999999999999999\n", WantExit: 1, Stderr: true,
				Explain: "When the number does not even fit in a long, strtol sets errno to ERANGE: set errno = 0 before calling it."},
		},
		Hints: []string{
			"Read the line with fgets into a char array, and check what fgets returns.",
			"strtol(text, &end, 10) converts and sets end to the first character it did not use. Compare end with text to see whether there were any digits.",
			"Errors go to stderr with fprintf(stderr, ...); return 1 from main to report failure.",
		},
	},
	{
		ID: "count", Stage: 1, Concept: "Files & Streams: Reading and Writing Data",
		Title: "Count lines, words and bytes",
		Task:  "Write a program that opens the file named by its first argument and prints three numbers like wc: the lines (newline characters), the words (runs of characters that are not white space) and the bytes. Without an argument, print a usage message to standard error and return 2; if the file cannot be opened, report why on standard error (perror) and return 1. Files may hold any bytes, including zero bytes.",
		File:  "count.c",
		Starter: `/* Checked exercise: count
 * Usage: ./count FILE
 * Print LINES WORDS BYTES, like wc: lines are '\n' characters, words are
 * runs of non-white-space characters, bytes are all bytes.
 * No argument: usage message on stderr, return 2.
 * Cannot open the file: perror, return 1.
 */
`,
		Cases: []CheckCase{
			{Name: "four lines", Args: []string{"a.txt"}, Files: map[string]string{"a.txt": "one two\nthree\n\nfour five six\n"}, Want: "4 6 29\n",
				Explain: "An empty line still counts as a line; words are separated by any white space."},
			{Name: "no newline at the end", Args: []string{"b.txt"}, Files: map[string]string{"b.txt": "last line"}, Want: "0 2 9\n",
				Explain: "Like wc, count newline characters: a file without a final newline has one fewer line than it seems."},
			{Name: "an empty file", Args: []string{"c.txt"}, Files: map[string]string{"c.txt": ""}, Want: "0 0 0\n"},
			{Name: "tabs and spaces", Args: []string{"d.txt"}, Files: map[string]string{"d.txt": "  a\tb  \n"}, Want: "1 2 8\n",
				Explain: "isspace (ctype.h) treats spaces, tabs and newlines alike."},
			{Name: "zero bytes", Args: []string{"e.bin"}, Files: map[string]string{"e.bin": "\x00\xff a\n"}, Want: "1 2 5\n",
				Explain: "fgets and strlen stop at a zero byte. Read with fgetc (or fread) and count every byte; pass a byte to isspace as an unsigned char."},
			{Name: "a big file", Args: []string{"f.txt"}, Files: map[string]string{"f.txt": strings.Repeat("x\n", 100000)}, Want: "100000 100000 200000\n"},
			{Name: "a missing file", Args: []string{"nope.txt"}, WantExit: 1, Stderr: true,
				Explain: "Check fopen's result for NULL and report it with perror(argv[1])."},
			{Name: "no argument", WantExit: 2, Stderr: true,
				Explain: "Check argc before reading argv[1]: without an argument, argv[1] is NULL."},
		},
		Hints: []string{
			"int c; while ((c = fgetc(f)) != EOF) visits every byte, zero bytes included.",
			"A word starts when a non-space character follows a space (or the start of the file): keep one flag saying whether you are inside a word.",
			"perror(argv[1]) prints the file name and the reason, such as No such file or directory.",
		},
	},
	{
		ID: "stack", Stage: 1, Concept: "Headers, the Preprocessor & make",
		Title: "A stack module: header and implementation",
		Task:  "Write a small module in two files. stack.h declares STACK_MAX (64), struct stack, and stack_init, stack_push and stack_pop, behind an include guard. stack.c includes it and defines the three functions. The checker's own main includes stack.h twice, on purpose, and links against your stack.c.",
		File:  "stack.c",
		Extra: []CheckFile{{Name: "stack.h", Starter: `/* Checked exercise: stack (the interface)
 * Declare here, behind an include guard (#ifndef STACK_H ... #endif):
 *   #define STACK_MAX 64
 *   struct stack { int items[STACK_MAX]; int top; };
 *   void stack_init(struct stack *s);
 *   int  stack_push(struct stack *s, int x);    0, or -1 when full
 *   int  stack_pop(struct stack *s, int *out);  0, or -1 when empty
 * The checker includes this header twice, on purpose.
 */
`}},
		Starter: `/* Checked exercise: stack (the implementation)
 * #include "stack.h" and define stack_init, stack_push and stack_pop.
 * Do not write main: the checker has its own.
 */
`,
		Harness: `#include <stdio.h>
#include <string.h>
#include "stack.h"
#include "stack.h" /* twice, on purpose: the include guard must make this harmless */

int main(int argc, char **argv) {
    if (argc < 2) return 2;
    struct stack s;
    stack_init(&s);
    int x;
    if (strcmp(argv[1], "lifo") == 0) {
        stack_push(&s, 1); stack_push(&s, 2); stack_push(&s, 3);
        while (stack_pop(&s, &x) == 0) printf("%d ", x);
        printf("\n");
    } else if (strcmp(argv[1], "empty") == 0) {
        x = 42;
        int r = stack_pop(&s, &x);
        printf("%d %d\n", r, x);
    } else if (strcmp(argv[1], "full") == 0) {
        int ok = 0;
        for (int i = 0; i < STACK_MAX; i++) ok += stack_push(&s, i) == 0;
        printf("%d %d %d\n", ok == STACK_MAX, stack_push(&s, 99), STACK_MAX);
    } else if (strcmp(argv[1], "independent") == 0) {
        struct stack t;
        stack_init(&t);
        stack_push(&s, 1); stack_push(&t, 2);
        stack_pop(&s, &x); printf("%d ", x);
        stack_pop(&t, &x); printf("%d\n", x);
    } else return 2;
    return 0;
}
`,
		BuildHints: []BuildHint{
			{"redefinition of", "Something is defined twice: the checker includes stack.h twice, so the header needs an include guard (#ifndef STACK_H, #define STACK_H, ..., #endif)."},
			{"stack.h: No such file", "Create stack.h in the same folder as stack.c."},
			{"implicit declaration of function", "A function is used before it is declared: put its prototype in stack.h."},
			{"undefined reference", "The linker found no definition: define every function declared in stack.h in stack.c."},
		},
		Cases: []CheckCase{
			{Name: "last in, first out", Args: []string{"lifo"}, Want: "3 2 1\n"},
			{Name: "pop from an empty stack", Args: []string{"empty"}, Want: "-1 42\n",
				Explain: "Return -1 and leave *out untouched when the stack is empty."},
			{Name: "full at STACK_MAX", Args: []string{"full"}, Want: "1 -1 64\n",
				Explain: "Exactly STACK_MAX (64) pushes succeed; the next returns -1 instead of writing past items."},
			{Name: "two stacks are separate", Args: []string{"independent"}, Want: "1 2\n",
				Explain: "Keep all state inside the struct: a global or static counter would be shared by every stack."},
		},
		Hints: []string{
			"The guard: #ifndef STACK_H on the first line, #define STACK_H on the second, #endif on the last.",
			"top can be the number of items: push writes items[top++], pop reads items[--top].",
		},
	},
	{
		ID: "fix_average", Stage: 1, Concept: "Debugging & Reading Errors",
		Title: "Fix the average",
		Task:  "The file already holds a program with two bugs. Find and fix them so that it prints 81.2 and then 1.5. Read what the checker reports first: the sanitizer points at one of the bugs.",
		File:  "fix_average.c",
		Starter: `#include <stdio.h>

/* Checked exercise: fix_average
 * average returns the mean of the n values in a. This program should
 * print 81.2 and then 1.5, but it has two bugs. Run the check first
 * and read the report: it points you at one of them.
 */
double average(const int *a, int n) {
    int sum = 0;
    for (int i = 0; i <= n; i++) {
        sum += a[i];
    }
    return sum / n;
}

int main(void) {
    int scores[] = {90, 72, 85, 60, 99};
    printf("%.1f\n", average(scores, 5));
    int pair[] = {1, 2};
    printf("%.1f\n", average(pair, 2));
    return 0;
}
`,
		Cases: []CheckCase{{Name: "both averages", Want: "81.2\n1.5\n",
			Explain: "A read past the end of the array means the loop runs once too often. 81.0 and 1.0 mean the division throws the fraction away."}},
		Hints: []string{
			"An array of n elements has valid indexes 0 to n - 1.",
			"sum / n with two ints is integer division, even if the result is returned as a double.",
		},
	},
	{
		ID: "wordfreq", Stage: 1, Project: true,
		Title:   "Stage 1 project: word frequencies",
		Task:    "Read text from standard input and print its ten most frequent words, most frequent first, as COUNT WORD. A word is a run of ASCII letters, compared in lower case (The and the are one word); everything else separates words. Equal counts are listed in alphabetical order, and with fewer than ten different words print them all. Words can be any length and the input any size, so use memory you allocate, and free all of it. This joins the whole stage: loops, functions, strings, structs, dynamic memory and input.",
		File:    "wordfreq.c",
		Leaks:   true,
		Timeout: 5 * time.Second,
		Starter: `/* Stage 1 project: wordfreq
 * Read text from stdin; print the 10 most frequent words, most frequent
 * first, as "COUNT WORD". A word is a run of ASCII letters, in lower
 * case. Ties: alphabetical order. Any word length, any input size: use
 * malloc/realloc, and free everything before main returns.
 * Try it: ./wordfreq < some_book.txt
 */
`,
		Cases: []CheckCase{
			{Name: "a sentence", Stdin: "The cat and the hat. The END!\n", Want: "3 the\n1 and\n1 cat\n1 end\n1 hat\n",
				Explain: "Lower-case every letter before comparing, and treat punctuation as a separator."},
			{Name: "ties in alphabetical order", Stdin: "b a c b a c d\n", Want: "2 a\n2 b\n2 c\n1 d\n",
				Explain: "Sort by count (highest first), then by the word with strcmp."},
			{Name: "only the top ten", Stdin: twelveWords(), Want: "12 aa\n11 bb\n10 cc\n9 dd\n8 ee\n7 ff\n6 gg\n5 hh\n4 ii\n3 jj\n"},
			{Name: "no input", Stdin: "", Want: "",
				Explain: "With no words your array is still NULL, and qsort(NULL, 0, …) is undefined behaviour: sort only when there is something to sort."},
			{Name: "no letters at all", Stdin: "123 456 !!!\n", Want: "",
				Explain: "Digits and punctuation only separate words: there are none here, so print nothing (and do not sort an empty array)."},
			{Name: "upper and lower case", Stdin: "Apple apple APPLE banana\n", Want: "3 apple\n1 banana\n"},
			{Name: "a 5000-letter word", Stdin: strings.Repeat("a", 5000) + " x " + strings.Repeat("a", 5000) + "\n", Want: "2 " + strings.Repeat("a", 5000) + "\n1 x\n",
				Explain: "A fixed-size word buffer overflows here: grow the buffer for the current word with realloc."},
			{Name: "200,000 words", Gen: generated(func() (string, string) {
				return manyWords(), "2000 aa\n2000 ab\n2000 ac\n2000 ad\n2000 ae\n2000 af\n2000 ag\n2000 ah\n2000 ai\n2000 aj\n"
			}),
				Explain: "Every word appears 2000 times, so the ten shown are the first ten alphabetically."},
		},
		Hints: []string{
			"Read with getchar into a growing buffer for the current word; when a non-letter arrives, the word is finished.",
			"Keep an array of struct { char *word; long count; }, grown with realloc, and look each finished word up in it (a linear search is fast enough here).",
			"qsort with a comparison that returns a positive number when a should come after b: higher counts first, then strcmp.",
			"Before returning, free every word you copied and then the array itself.",
		},
	},
}

// fizzBuzzWant is the expected output of FizzBuzz from 1 to n.
func fizzBuzzWant(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0:
			b.WriteString("FizzBuzz\n")
		case i%3 == 0:
			b.WriteString("Fizz\n")
		case i%5 == 0:
			b.WriteString("Buzz\n")
		default:
			fmt.Fprintf(&b, "%d\n", i)
		}
	}
	return b.String()
}

// hundredStudents is the input "s00 0" … "s99 99".
func hundredStudents() string {
	var b strings.Builder
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&b, "s%02d %d\n", i, i)
	}
	return b.String()
}

// twelveWords has twelve words, "aa" twelve times down to "ll" once.
func twelveWords() string {
	var b strings.Builder
	for i := 0; i < 12; i++ {
		w := strings.Repeat(string(rune('a'+i)), 2)
		for n := 0; n < 12-i; n++ {
			b.WriteString(w + " ")
		}
	}
	return b.String() + "\n"
}

// manyWords is 200,000 words: "aa" to "jj", 2,000 times each, mixed up.
func manyWords() string {
	var b strings.Builder
	for k := 0; k < 200000; k++ {
		i := k * 7 % 100
		b.WriteByte(byte('a' + i/10))
		b.WriteByte(byte('a' + i%10))
		if k%12 == 11 {
			b.WriteString(".\n")
		} else {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// The checks of every file, in teaching order: by stage, and within a
// stage in the order they are written.
func init() {
	checks = append(checks, coreChecks...)
	checks = append(checks, systemsChecks...)
	checks = append(checks, foundationsChecks...)
	checks = append(checks, softwareChecks...)
	sort.SliceStable(checks, func(i, j int) bool { return checks[i].Stage < checks[j].Stage })
}

// checkByID finds a check.
func checkByID(id string) (Check, bool) {
	for _, c := range checks {
		if c.ID == id {
			return c, true
		}
	}
	return Check{}, false
}

// checksFor lists a stage's checks, in order.
func checksFor(stage int) []Check {
	var out []Check
	for _, c := range checks {
		if c.Stage == stage {
			out = append(out, c)
		}
	}
	return out
}

// checksForConcept lists the checks that demonstrate a concept.
func checksForConcept(name string) []Check {
	var out []Check
	for _, c := range checks {
		if c.Concept == name {
			out = append(out, c)
		}
	}
	return out
}
