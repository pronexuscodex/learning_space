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
	"strings"
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

	Cases []CheckCase
	Hints []string
}

// CheckCase is one test: run the program with Args and Stdin, then compare
// standard output (trailing spaces and blank lines aside) and the exit code.
type CheckCase struct {
	Name     string
	Args     []string
	Stdin    string
	Want     string
	WantExit int
	Stderr   bool   // the program must also print a message to standard error
	Explain  string // shown when this case fails: the likely cause
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
			{Name: "a long input", Stdin: strings.Repeat("ab", 100000), Want: "a 100000\nb 100000\n",
				Explain: "Read character by character (getchar until EOF) instead of into a fixed-size buffer."},
		},
		Hints: []string{
			"getchar() returns the next byte as an int, or EOF at the end of input. Keep it in an int, not a char.",
			"An array int counts[26] = {0} has one counter per letter; c - 'a' turns 'a'..'z' into 0..25.",
			"tolower and isalpha (ctype.h) expect a value of unsigned char or EOF; bytes of an accented letter are not ASCII letters.",
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
