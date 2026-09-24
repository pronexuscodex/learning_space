package main

// Track F: Foundations of Computing (stages 1–4). The starting point for a
// complete beginner: programming, data structures & algorithms, discrete
// mathematics, and how the hardware itself works.

var foundationsGuides = map[int]StageGuide{
	// -----------------------------------------------------------------
	1: {
		Overview: `Programming is telling a very fast, very literal machine exactly what to
do. This stage teaches the handful of building blocks every language
shares (values, variables, decisions, loops, functions and collections)
and, just as importantly, how to break a problem into steps and how to
fix code when it does not work. Any language will do; the examples use
Python because it reads almost like English.`,
		Outcomes: []string{
			"Write small, useful programs from scratch in any mainstream language",
			"Break a vague problem into clear steps and functions",
			"Read error messages calmly and debug methodically",
		},
		Glossary: []Term{
			{"Program", "A list of instructions a computer follows exactly, in order."},
			{"Variable", "A name that refers to a value, such as total = 42."},
			{"Type", "The kind of a value (number, text, true/false, list), which decides what you can do with it."},
			{"String", "A piece of text, such as \"hello\"."},
			{"Boolean", "A value that is either true or false."},
			{"Loop", "Code that repeats, either for each item in a collection or while a condition holds."},
			{"Function", "A named, reusable piece of code that takes inputs and returns an output."},
			{"Bug", "A mistake that makes a program behave differently from what you intended."},
			{"Traceback", "The error report listing which functions were running, and on which lines, when a crash happened."},
		},
		Concepts: []Concept{
			{
				Name:    "Values, Types & Variables",
				Summary: "A variable is a named box that holds a value of some type.",
				Body: `Programs work with values: numbers (42, 3.14), text ("hello"),
booleans (true/false) and collections. A value's type says what you can
do with it. You can add numbers and join strings, but "5" + 3 means
different things in different languages: Python raises an error,
JavaScript gives "53".

A variable is a name bound to a value. Assignment (x = x + 1) means
"work out the right-hand side, then store it under the name on the left".
It is an instruction, not a mathematical equation. Statically typed
languages (Go, Rust, Java) check types before the program runs;
dynamically typed ones (Python, JavaScript) check while it runs.`,
				MentalModel: "A variable is a label on a box, and the type says what may go in the box.",
				TryIt:       `In Python, run type(42), type("42"), int("42") + 1 and "42" * 2, and explain each result.`,
				Analogy: `Labelled jars in a kitchen: the "sugar" jar holds sugar. You can empty
it and refill it, but pour flour into the sugar jar and the next cake
goes wrong. Types are like units: adding 3 metres to 2 seconds makes no
sense, and a good type system refuses to let you try.`,
				Example: `In 2020 scientists officially renamed several human genes, such as SEPT1
and MARCH1, because Excel kept silently converting their names into
dates. That is a type problem. Spreadsheets that strip the leading zero
from phone numbers and postcodes make the same mistake: they treat text
as a number.`,
				Exercises: trio(
					`Without running anything, predict x after: x = 5, then x = x * 2, then x = x - 3. Then name the type of each of: 7, 7.0, "7", True.`,
					"Assignment works right to left: compute first, then store.",
					`Write a tip calculator: ask for the bill and the tip percentage, print the tip and the total rounded to 2 decimal places, and reply politely if someone types "twelve" instead of a number.`,
					"Convert the text input with float() inside try/except ValueError.",
					`A shop's spreadsheet export has prices such as "$1,299.00", "15" and "N/A". Write a function that converts each one to a whole number of cents, or reports it as missing, and test it on ten messy values.`,
					`Strip "$" and ",", multiply by 100 and round to an int. Never keep money in floats.`,
				),
			},
			{
				Name:    "Control Flow: Decisions & Loops",
				Summary: "if chooses, loops repeat, and together they can express any procedure.",
				Body: `Code normally runs top to bottom. An if statement runs a block only when a
condition is true, with else/elif for the alternatives. Conditions
combine with and, or and not.

A loop repeats a block. A for loop visits each item in a collection or a
range; a while loop repeats until its condition becomes false. The
classic bugs are off-by-one errors (one iteration too many or too few)
and infinite loops, where the condition never changes. With just
sequence, choice and repetition you can express any algorithm.`,
				Diagram: `            ┌─────────────────────────────┐
            ▼                             │
 start ─▶ [ i < 10 ? ] ──yes──▶ [ work; i = i + 1 ]
              │
              no
              ▼
            done`,
				MentalModel: "Sequence, choice and repetition: the three moves that build every program.",
				TryIt:       `Write FizzBuzz: print 1 to 100, but "Fizz" for multiples of 3, "Buzz" for multiples of 5 and "FizzBuzz" for both.`,
				Analogy: `A recipe: "if the dough is sticky, add flour" is a decision, and "stir
until smooth" is a loop. A recipe that said "stir until the cake is
baked", without ever putting it in the oven, would have you stirring
forever. That is an infinite loop.`,
				Example: `A thermostat is a loop with a decision inside: every minute, if the room
is colder than the target, turn the heating on, otherwise turn it off.
Traffic lights, washing-machine programs and your phone's "retry three
times" when joining Wi-Fi are all loops and conditions.`,
				Exercises: trio(
					"How many times does for i in range(2, 10, 3) run, and with which values of i?",
					"range(start, stop, step) stops before reaching stop: 2, 5, 8.",
					`Write a number-guessing game: pick a random number from 1 to 100, let the user guess, reply "higher" or "lower", and count the attempts.`,
					"A while loop that ends when the guess is right; random.randint(1, 100).",
					`A gym charges €30 a month. Students get 20% off, and members of more than 2 years get another €5 off, but the price never drops below €15. Write the price function and test it on five different people.`,
					"Write each rule as an if, then apply the floor with max(price, 15).",
				),
			},
			{
				Name:    "Functions & Decomposition",
				Summary: "Name a piece of work once, reuse it everywhere, and break big problems into small ones.",
				Body: `A function takes inputs (parameters), does some work and returns an
output. Good functions do one thing, have a clear name and do not
secretly change things outside themselves (side effects). That makes
them easy to test and reuse.

Decomposition is the core skill of programming. Split "build a to-do
app" into "load tasks", "add a task" and "save tasks", then split those
again until every piece is obvious. Each call gets its own local
variables. A function that calls itself (recursion) is solving a smaller
copy of the same problem.`,
				MentalModel: "If you cannot name it, you have not understood it; if it is long, split it.",
				TryIt:       "Write celsius_to_fahrenheit(c) and fahrenheit_to_celsius(f), and check that converting there and back returns the original value.",
				Analogy: `A coffee machine: water and beans go in (inputs), you press one button,
and coffee comes out (the output). You do not need to know how the pump
works. And nobody opens a café in one step: you set up the machine, the
till and the menu separately, then put them together.`,
				Example: `Every "Log in with Google" button calls the same small set of functions,
shared by millions of websites. App developers build big features by
combining small functions such as validate_email(), hash_password() and
send_email().`,
				Exercises: trio(
					"List the inputs and the output of these real-world 'functions': a vending machine, a calculator's √ button and a translation app.",
					"Ask what goes in, what comes out, and whether anything else in the world changes.",
					`Write is_palindrome(text), ignoring spaces, punctuation and capital letters, and test it on "A man, a plan, a canal: Panama".`,
					"Clean the text first, then compare it with its reverse.",
					`Decompose a "split the restaurant bill" app into functions (parse the items, assign items to people, add tax and tip, round to cents so the shares still add up to the total), then implement it.`,
					"Rounding each share can lose a cent, so give leftover cents to someone explicitly.",
				),
			},
			{
				Name:    "Collections: Lists, Dictionaries & Strings",
				Summary: "Real programs work on many values at once.",
				Body: `A list (array) keeps items in order and gives access by position,
counting from 0. A dictionary (map, hash table) looks values up by key,
as in phone_book["Ana"]. A set holds unique items and answers "is x in
here?" quickly. Strings are sequences of characters, with methods to
split, search and replace.

Most everyday programs follow the same shape: loop over a collection,
filter some items, transform them and summarise them (count, sum, group
by). Choosing the right collection often matters more than clever code.`,
				MentalModel: "Order matters? Use a list. Look up by name? Use a dictionary. Must be unique? Use a set.",
				TryIt:       "Count how often each word appears in a paragraph and print the 5 most common words.",
				Analogy: `A list is the numbered queue at a bakery. A dictionary is a coat check:
your ticket (the key) gets your coat (the value) back instantly. A set
is a guest list where each name appears once.`,
				Example: `Your phone's contacts app is a dictionary from names to numbers. A
Spotify playlist is a list, because order matters. "People who liked
this post" is a set, because nobody can like a post twice.`,
				Exercises: trio(
					"Choose a list, dictionary or set for each: a shopping list, word definitions, today's unique website visitors, and the pages of a book.",
					"Ask: does order matter? Do I look things up by name? Must items be unique?",
					"Given a list of (student, score) pairs, compute each student's average and print the students from highest to lowest average.",
					"A dictionary from each name to a list of that student's scores.",
					"Export (or invent) a CSV of your bank transactions, total the spending per category per month, and print the three biggest categories.",
					"Use the csv module and key a dictionary by (month, category).",
				),
			},
			{
				Name:    "Debugging & Reading Errors",
				Summary: "Bugs are normal; finding them is a skill you can learn.",
				Body: `An error message is a clue, not an insult. Read the last line first (what
went wrong), then the traceback (where it went wrong: file and line,
with the innermost call last). Syntax errors stop a program from
starting. Runtime errors, such as dividing by zero or a missing key,
crash it midway. Logic errors silently give wrong answers, and they are
the hardest to find.

Use the scientific method: reproduce the bug reliably, form a
hypothesis, test it by printing values or stepping through with a
debugger, fix it, then add a test so it never comes back. Explaining
your code line by line out loud ("rubber-duck debugging") finds a
surprising number of bugs.`,
				MentalModel: "The computer did exactly what you said; find where that differs from what you meant.",
				TryIt:       "Paste a buggy loop into pythontutor.com and watch every variable change, step by step.",
				Analogy: `A doctor diagnosing symptoms: gather evidence, form a hypothesis, test
one idea at a time, and do not prescribe before you have found the
cause.`,
				Example: `In 1947, engineers on the Harvard Mark II computer found a real moth
stuck in a relay and taped it into their logbook as the "first actual
case of bug being found". Grace Hopper helped make the story famous.
Today, large companies collect crash reports from millions of users to
find and fix bugs.`,
				Exercises: trio(
					`Read this error: "IndexError: list index out of range", raised on line 12 inside a loop over range(len(items) + 1). Explain in one sentence what went wrong.`,
					"A list of length n has indexes 0 to n − 1.",
					"Take a working program of yours and plant three bugs (a typo, an off-by-one and a wrong operator). Wait a day, then find and fix them using only error messages and print statements.",
					"Change one thing at a time and re-run after each change.",
					`A user reports "the app crashes when I leave the age field empty". Reproduce this in your tip calculator or guessing game, write a test that fails because of it, fix the bug, and confirm the test passes.`,
					`Empty input is "", and int("") raises ValueError.`,
				),
			},
		},
		Resources: []Resource{
			{"Course", "CS50: Introduction to Computer Science (Harvard, free)", "https://cs50.harvard.edu/x/", "The best-known beginner course; excellent lectures and problem sets."},
			{"Tool", "Python Tutor", "https://pythontutor.com/", "Visualises your code running step by step. Ideal for beginners."},
			{"Book", "How to Design Programs (free online)", "https://htdp.org/", "A systematic recipe for designing programs, not just typing them."},
			{"Book", "Structure and Interpretation of Computer Programs (free)", "https://sarabander.github.io/sicp/", "A classic; deep ideas about abstraction."},
			{"Site", "Exercism", "https://exercism.org/", "Free coding exercises in 70+ languages, with mentor feedback."},
			{"Site", "Advent of Code", "https://adventofcode.com/", "Yearly puzzles that are great for practising loops, collections and parsing."},
		},
		Blueprints: []Blueprint{
			{"Personal Budget Tracker", "A command-line tool that records income and spending, stores it in a file, and prints monthly summaries by category.",
				[]string{"Add and list transactions", "Save to and load from a file", "Monthly totals by category", "Import a bank CSV", "Tests for the calculations"}},
			{"Text Adventure Game", "A small interactive story with rooms, items and choices, driven by loops, conditions and dictionaries.",
				[]string{"Rooms as a dictionary", "Move between rooms", "Pick up and use items", "Win and lose conditions", "Save the game"}},
			{"Quiz App", "A flashcard quiz that loads questions from a file, shuffles them, keeps score and remembers your weak spots.",
				[]string{"Load questions", "Ask and check answers", "Score and summary", "Repeat missed questions", "Store history between runs"}},
		},
		Quiz: []Question{
			{"What does x = x + 1 mean in most programming languages?", "Compute x + 1 using the current value of x, then store the result back in x. It is an instruction, not an equation."},
			{"What is the difference between a for loop and a while loop?", "A for loop visits each item of a known collection or range; a while loop repeats until its condition becomes false."},
			{"Why split code into functions?", "To name ideas, reuse them, test them separately, and make big problems manageable."},
			{"When should you use a dictionary instead of a list?", "When you look values up by a key (such as a name or ID) rather than by position."},
		},
	},

	// -----------------------------------------------------------------
	2: {
		Overview: `A data structure is a way of organising data; an algorithm is a recipe
for working with it. Together they decide whether your program answers
in a millisecond or in a week. This stage gives you the classic toolbox,
and the language (Big-O) to reason about speed before you write a line
of code.`,
		Outcomes: []string{
			"Predict how a program will slow down as its data grows",
			"Choose the right data structure for a problem",
			"Solve classic problems (searching, sorting, routing) and explain why your solution is efficient",
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

Big-O usually describes the worst case. Memory use (space complexity)
matters too.`,
				Diagram: ` n            log n        n       n log n            n²
 10               3       10            33          100
 1,000           10    1,000        10,000    1,000,000
 1,000,000       20    10⁶       2 × 10⁷           10¹²  (days of work)`,
				MentalModel: "Constants matter in practice, but the growth rate decides who wins at scale.",
				TryIt:       "Time a nested-loop duplicate finder against a set-based one on 10,000 and 100,000 items.",
				Analogy: `Finding a word in a paper dictionary. Reading every page from the start
is O(n). Opening it in the middle, deciding which half the word is in
and halving again is O(log n): about 17 flips for 100,000 words.`,
				Example: `In 2021 a player discovered that GTA Online's loading screen spent
minutes parsing a 10 MB JSON file with an accidentally quadratic
algorithm. His fix cut loading times by about 70%, and Rockstar shipped
it and paid him a bounty.`,
				Exercises: trio(
					"Give the Big-O of: printing every item; printing every pair of items; finding a word by repeatedly halving a sorted list; reading the first item.",
					"Count how many times the innermost line runs as n grows.",
					"Write two versions of has_duplicates(items), one with nested loops and one with a set, and time both at n = 1,000, 10,000 and 100,000. Put the results in a table.",
					"time.perf_counter() before and after each call.",
					"A school emails every parent by, for each of 50,000 students, scanning the whole parent list for a match. It takes hours. State the Big-O and rewrite it to be fast.",
					"Build a dictionary from student ID to parent once: O(n) instead of O(n²).",
				),
			},
			{
				Name:    "Arrays, Linked Lists, Stacks & Queues",
				Summary: "The basic ways to line data up, and their trade-offs.",
				Body: `An array stores items side by side in memory. Access by index is
instant and cache-friendly, but inserting in the middle shifts
everything after it (O(n)). Dynamic arrays (Python lists, Go slices)
grow their capacity by a constant factor (often doubling) when full, so
appending is O(1) on average.

A linked list chains nodes with pointers. Inserting is O(1) once you
hold the node, but reaching the i-th item is O(n), and the scattered
memory is slow for the cache. A stack is last-in, first-out (push and
pop), used for undo, function calls and matching brackets. A queue is
first-in, first-out, used for print jobs, task scheduling and
breadth-first search.`,
				Diagram: `array:   [ 7 | 3 | 9 | 4 ]              index i → jump straight there
linked:  7 → 3 → 9 → 4 → ∅              walk from the head
stack:   push ▶ [ 4 | 9 | 3 | 7 ]        pop takes the 4  (last in, first out)
queue:   in ▶ [ 4 | 9 | 3 | 7 ] ▶ out    7 leaves first   (first in, first out)`,
				MentalModel: "Stack = pile of undo steps; queue = waiting line; array = numbered seats.",
				TryIt:       `Use a stack to check whether the brackets in "(a[b]{c})" are balanced.`,
				Analogy: `A stack of plates (you always take from the top), a queue at the post
office (first come, first served), numbered cinema seats (an array: walk
straight to seat 14) and a treasure hunt where each clue points to the
next one (a linked list).`,
				Example: `Ctrl+Z in every editor pops a stack of changes, and the browser's back
button pops a stack of pages. Print spoolers, customer-support ticket
systems and message queues such as RabbitMQ and Kafka are queues.`,
				Exercises: trio(
					"Push 1, 2 and 3 onto a stack, pop once, push 4, then pop twice. What is left? Repeat with a queue, using enqueue and dequeue.",
					"Draw the boxes on paper after every step.",
					"Implement a stack and a queue from scratch, without the language's built-in deque, and write tests for both.",
					"A queue can be built from two stacks.",
					"Build undo and redo for a tiny text editor: every edit can be undone, and undone edits can be redone until a new edit is made.",
					"Use two stacks, one for undo and one for redo. A new edit clears the redo stack.",
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

A good hash spreads keys evenly. A predictable one can be attacked with
keys that all collide, degrading every operation to O(n). This is
called hash flooding.`,
				Diagram: `"alice" ─hash─▶ 2      bucket 0 [                ]
"bob"   ─hash─▶ 5      bucket 2 [ alice → carol  ]  ◀ collision, chained
"carol" ─hash─▶ 2      bucket 5 [ bob            ]`,
				MentalModel: "Compute where it lives instead of searching for it.",
				TryIt:       "Implement a hash map with chaining and resizing, and count collisions as it grows.",
				Analogy: `A coat check: your ticket number tells the attendant exactly which hook
your coat is on. Nobody searches through every coat.`,
				Example: `Python dictionaries, JavaScript objects, database indexes on IDs, DNS
caches and Git (which names every file version by its hash) all rely on
hashing. In 2011 researchers showed they could freeze many web
frameworks by sending form fields built to collide, and languages then
started randomising their hash functions.`,
				Exercises: trio(
					`Using hash(key) = (sum of letter positions, a = 1) mod 5, place "cat", "act" and "dog" into 5 buckets. What happens with "cat" and "act"?`,
					"Anagrams have the same letter sum, so they collide.",
					"Write a word-frequency counter using your own hash table (no built-in dict) and compare its speed with the built-in dictionary on a full book from Project Gutenberg.",
					"Start with 8 buckets, and double when the average chain length exceeds 1.",
					"An online shop must find duplicate sign-ups whose emails differ only in capitalisation or surrounding spaces. Find all duplicates among 1 million emails in roughly O(n).",
					"Normalise each email (trim, lowercase) and use it as a dictionary key.",
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

Simple sorts (insertion, selection) are O(n²). Merge sort and heap sort
guarantee O(n log n). Quicksort is O(n log n) on average and very fast
in practice. No comparison-based sort can beat O(n log n) in the worst
case.`,
				Diagram: `          8            binary search tree: left < node < right
        ╱   ╲
       3     10        search for 6: 8 → 3 → 6  (3 steps, not 7)
      ╱ ╲      ╲
     1   6      14`,
				MentalModel: "Sorted data turns searching into halving.",
				TryIt:       "Implement merge sort and test it on 1,000 random lists against the built-in sort.",
				Analogy: `A tournament bracket is a tree. Hospital emergency triage is a heap: the
most urgent patient is seen next, not whoever arrived first. A sorted
library shelf lets you jump straight to the right spot.`,
				Example: `The folders on your computer form a tree. Route planners use heaps
inside Dijkstra's algorithm. Python's built-in sort (Timsort) is a hybrid
of merge sort and insertion sort, tuned for real data that is often
partly sorted already.`,
				Exercises: trio(
					"Insert 5, 2, 8, 1, 9 and 3 into an empty binary search tree and draw it. Then insert 1, 2, 3, 4, 5 into a fresh tree. What shape do you get, and why is it bad?",
					`Sorted input makes a "stick": a linked list in disguise.`,
					"Implement a min-heap with push and pop, and use it to sort a list (heap sort).",
					"Store the heap in an array: the children of i are at 2i + 1 and 2i + 2.",
					"An emergency-room app must always show the most urgent patient next (severity 1 to 5), and among equal severities, whoever arrived first. Implement it and simulate 20 arrivals.",
					"A heap of (severity, arrival_number, name) tuples.",
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
graph: dependencies, maps, social networks, the web.`,
				Diagram: `A ── 4 ── B        BFS from A (by hops):   A │ B C │ D
│         │
1         1        Dijkstra from A (by weight):
│         │          A = 0, C = 1, D = 3, B = 4
C ── 2 ── D`,
				MentalModel: "If things connect to things, draw the graph first.",
				TryIt:       "Solve a maze stored in a text file with BFS and print the shortest route.",
				Analogy: `A city map: intersections are nodes, roads are edges and travel times
are weights. BFS is a rumour spreading through friends of friends. DFS is
exploring a cave by always taking the next unexplored tunnel and
backtracking at dead ends.`,
				Example: `Google Maps finds routes with shortest-path algorithms on the road graph.
LinkedIn's "2nd-degree connection" is BFS. Package managers such as npm
and pip use topological sorting to install dependencies in the right
order. Google's original PageRank treated the web as a graph of links.`,
				Exercises: trio(
					"Draw a graph of five friends and who knows whom. Who is two hops from you? Is there a cycle?",
					"Do BFS by hand: ring 1, then ring 2.",
					"Implement BFS, DFS and Dijkstra on an adjacency list, and test them on the small graph in the diagram.",
					"BFS uses a queue, DFS a stack, and Dijkstra a min-heap.",
					"Given a list of university courses and their prerequisites, print an order in which a student can take every course, or report that the prerequisites form an impossible cycle.",
					"Topological sort (Kahn's algorithm): repeatedly take the courses that have no remaining prerequisites.",
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
		},
	},

	// -----------------------------------------------------------------
	3: {
		Overview: `Discrete mathematics is the maths of separate, countable things, which
is exactly what computers work with. It teaches you to reason precisely,
to prove that code is correct, to count possibilities and to think
clearly about chance. You do not need calculus for this stage.`,
		Outcomes: []string{
			"Reason precisely with logic, sets and proofs",
			"Count possibilities to judge whether brute force is feasible",
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
same as "B → A", a very common reasoning error. Quantifiers say "for all"
(∀) and "there exists" (∃). Every if statement you write is logic, and
Stage 4 builds these operations out of transistors.`,
				Diagram: `A  B │ A AND B │ A OR B │ A → B
0  0 │    0    │   0    │   1
0  1 │    0    │   1    │   1
1  0 │    0    │   1    │   0
1  1 │    1    │   1    │   1`,
				MentalModel: "Every condition in code is a small logical formula, so simplify it on paper first.",
				TryIt:       `Simplify "not (age < 18 or not member)" with De Morgan's laws, then check it with a truth table.`,
				Analogy: `A security door opens only if you have a badge AND know the PIN. A café
gives a discount if you are a student OR a senior. And "if it rains, the
ground is wet" does not mean that a wet ground proves it rained: someone
may have watered the lawn.`,
				Example: `Advanced search (cats AND NOT dogs), spreadsheet filters and database
WHERE clauses are Boolean logic. Tools called SAT solvers check logical
formulas with millions of variables to verify that chips and software
behave correctly.`,
				Exercises: trio(
					`Build the truth table for "(A OR B) AND NOT (A AND B)". Which everyday phrase describes it?`,
					`It is true when exactly one input is true: "either, but not both" (exclusive or).`,
					"Write a program that prints the truth table of any formula over A, B and C, given as a Python function, and use it to verify both De Morgan laws.",
					"itertools.product([False, True], repeat=3) gives every row.",
					`A cinema's rule: "Children under 12 need an adult, unless the film is rated U, and nobody under 18 may see an 18-rated film." Write it as one Boolean expression, then as a function, and test it on 8 cases.`,
					"Write the table of cases first, then derive the expression from it.",
				),
			},
			{
				Name:    "Sets, Functions & Relations",
				Summary: "The vocabulary for describing collections and connections precisely.",
				Body: `A set is a collection of distinct items, such as {1, 2, 3}. The
operations are union (A ∪ B: in either), intersection (A ∩ B: in both),
difference (A − B) and complement.

A relation connects items from sets, such as "is enrolled in" between
students and courses. A function is a relation where each input has
exactly one output. Functions can be one-to-one (injective), onto
(surjective) or both (bijective, which means perfectly reversible).
Equivalence relations group items into classes (for example, the same
remainder mod 3), and orders rank them (≤). SQL databases are built on
exactly this relational mathematics.`,
				Diagram: `┌──────── A ────────┐
│  A only   ┌───────┼──────── B ───────┐
│           │ A ∩ B │  B only          │
└───────────┼───────┘                  │
            └──────────────────────────┘`,
				MentalModel: "Name the sets and the relations, and the problem usually becomes clear.",
				TryIt:       "Using Python sets, find which friends follow both of two accounts, and which follow only one.",
				Analogy: `Guest lists for two parties: the union is everyone invited to either,
and the intersection is the people invited to both. A function is like
the codes on a vending machine: each code gives exactly one snack.`,
				Example: `"People you may know" and "customers who bought this also bought" are
built on set intersections. A SQL JOIN combines relations. Sync tools use
set difference to work out which files are new or missing.`,
				Exercises: trio(
					"With A = {1, 2, 3, 4} and B = {3, 4, 5}, compute A ∪ B, A ∩ B, A − B and B − A.",
					"A difference keeps only what is in the first set.",
					"Write functions that check whether a dictionary-based function is injective, surjective (onto a given set) or bijective.",
					"Injective means no two keys share a value.",
					"You have two CSV exports of newsletter subscribers from different years. Produce three lists (unsubscribed, new and loyal), comparing emails without regard to case.",
					"Normalise the emails into sets, then use difference and intersection.",
				),
			},
			{
				Name:    "Proof & Induction",
				Summary: "How to be certain, not just confident, that something always works.",
				Body: `A proof is an argument that leaves no case unchecked. A direct proof
starts from what you know and derives the claim. A proof by
contradiction assumes the opposite and reaches an impossibility (this is
how you show that √2 is irrational).

Mathematical induction proves a claim for every n. Show that it holds in
the base case, then show that if it holds for n, it holds for n + 1,
like falling dominoes. A loop invariant is induction applied to code: a
statement that is true before and after every iteration, which proves
the loop correct. Recursion and induction are mirror images of each
other.`,
				Diagram: `base case         step: if domino n falls, domino n + 1 falls
▌ ▌ ▌ ▌ ▌ ▌  ▶  ╱ ╱ ╱ ╱ ╱ ╱   …so every domino falls
1 2 3 4 5 6`,
				MentalModel: "Prove the first case, prove each case gives the next, and every case is done.",
				TryIt:       "Prove by induction that 1 + 2 + … + n = n(n + 1)/2, then check it in code for n up to 1,000.",
				Analogy: `Climbing a ladder: if you can get onto the first rung, and from any rung
you can reach the next one, you can climb as high as you like.`,
				Example: `Amazon Web Services has used the formal specification language TLA+ to
check the designs of services such as S3 and DynamoDB, catching subtle
bugs that testing had missed. The seL4 operating-system kernel is
mathematically proven correct and is used in safety-critical systems.`,
				Exercises: trio(
					`Find the flaw: "All horses are the same colour. One horse is trivially one colour. If any n horses share a colour, so do n + 1, because the first n and the last n overlap."`,
					"Check the step from n = 1 to n = 2. Do the two groups overlap?",
					"Write binary search, state its loop invariant in a comment, and add an assert that checks the invariant on every iteration.",
					`The invariant: "if the target is in the list, it lies between low and high".`,
					"Your bank app rounds each of 3 interest payments to the cent. Prove, or disprove with a counterexample, that rounding each payment gives the same total as rounding their sum.",
					"Try three payments of 0.005 each.",
				),
			},
			{
				Name:    "Counting & Combinatorics",
				Summary: "How many ways? The maths behind passwords, probabilities and brute force.",
				Body: `The product rule: if one choice can be made in a ways and another in b
ways, there are a × b combinations. Permutations count ordered
arrangements: n! ways to order n items. Combinations count unordered
selections: C(n, k) = n! / (k!(n − k)!).

The pigeonhole principle: put more than n items into n boxes, and some
box holds two. That is why hash collisions must exist, and why no
lossless compressor can shrink every file. Counting tells you whether
"try every possibility" takes seconds or longer than the age of the
universe.`,
				MentalModel: "Count before you brute-force.",
				TryIt:       "Count the 4-digit PINs with no repeated digit, first with a formula and then with a program.",
				Analogy: `Outfits: 3 shirts × 4 pairs of trousers × 2 pairs of shoes gives 24
outfits. Picking 3 pizza toppings from 10 does not care about order
(combinations); awarding gold, silver and bronze does (permutations).`,
				Example: `A 4-digit PIN has 10⁴ = 10,000 options, which is why cards lock after
three wrong tries. An 8-character password drawn from 94 symbols has
about 6 × 10¹⁵ options, but people pick words, so attackers use
dictionaries instead. The chance of winning a 6-from-49 lottery is 1 in
C(49, 6) = 13,983,816.`,
				Exercises: trio(
					"How many 3-letter codes can you make from A to Z if letters may repeat? And if they may not?",
					"26 × 26 × 26 versus 26 × 25 × 24.",
					"Write functions for n!, P(n, k) and C(n, k), and verify C(n, k) by generating every subset with itertools for small n.",
					"Use integers, not floats, so large results stay exact.",
					"A website requires 8-character passwords of lowercase letters. An attacker can try 10 billion guesses per second offline. How long to try them all? And for 12 characters mixing upper case, lower case and digits? Write the calculation as a small program.",
					"26⁸ ≈ 2 × 10¹¹ takes about 20 seconds; 62¹² ≈ 3.2 × 10²¹ takes about 10,000 years.",
				),
			},
			{
				Name:    "Probability for Programmers",
				Summary: "Reasoning about chance, and why intuition is often wrong.",
				Body: `Probability measures how likely an event is, from 0 (impossible) to 1
(certain). For equally likely outcomes, P = favourable ÷ total.
Independent events multiply: two heads in a row is ½ × ½. The expected
value is the long-run average.

Conditional probability P(A|B) is the chance of A given that B happened,
and Bayes' rule flips it: P(A|B) = P(B|A) · P(A) / P(B). Intuition fails
often: in a room of just 23 people, there is about a 50% chance that two
share a birthday. Programmers use probability for randomised
algorithms, hashing, load balancing, A/B tests and machine learning
(Stage 14).`,
				MentalModel: "When intuition and arithmetic disagree, simulate.",
				TryIt:       "Simulate the birthday problem 10,000 times for 23 people and compare the result with the formula.",
				Analogy: `A forecast of "70% chance of rain" means that on days like this, it
rained about 7 times out of 10. And a medical test that is "99%
accurate" can still produce mostly false alarms when the disease is
rare, because healthy people vastly outnumber sick ones.`,
				Example: `Spam filters apply Bayes' rule to the words in an email. A/B tests
decide whether a new design is really better or just lucky. The birthday
paradox is why hash functions need long outputs to avoid collisions.`,
				Exercises: trio(
					"What is the chance of rolling a total of 7 with two dice? And a total of 12?",
					"List all 36 outcomes: 6/36 for a 7 and 1/36 for a 12.",
					`Simulate the Monty Hall problem 100,000 times and measure the win rate for "stay" and for "switch".`,
					"Randomise the car's door and your first pick; the host opens a goat door that you did not pick.",
					"A disease affects 1 in 1,000 people. A test catches 99% of cases but also flags 5% of healthy people. If you test positive, what is the chance you are sick? Work it out with Bayes' rule and with a simulation of 1 million people.",
					"About 2%: false positives from the healthy majority swamp the true positives.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "MIT 6.042J Mathematics for Computer Science (OCW)", "https://ocw.mit.edu/courses/6-042j-mathematics-for-computer-science-fall-2010/", "The standard course; the lecture notes are a full free textbook."},
			{"Book", "Book of Proof (free)", "https://www.people.vcu.edu/~rhammack/BookOfProof/", "A gentle, clear introduction to proofs."},
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
		},
	},

	// -----------------------------------------------------------------
	4: {
		Overview: `How does a lump of silicon run your code? This stage climbs from a
single switch to a working CPU: binary numbers, logic gates, memory
circuits, the instruction cycle and the tricks that make modern chips
fast. After it, the machine is no longer magic. It is a clever
arrangement of on/off switches.`,
		Outcomes: []string{
			"Explain how numbers, text and images are stored as bits",
			"Design small digital circuits, including an adder",
			"Describe what a CPU does on every clock tick and why some code runs faster",
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
		},
		Concepts: []Concept{
			{
				Name:    "Binary & Number Representation",
				Summary: "Everything in a computer (numbers, text, images) is patterns of 0s and 1s.",
				Body: `A bit is one binary digit. Binary is base 2, so 1011 = 8 + 0 + 2 + 1 = 11.
Eight bits make a byte (0 to 255). Hexadecimal (base 16, with digits
0–9 and A–F) is shorthand for 4 bits per digit: 0xFF = 255.

Text is numbers too. In ASCII, 'A' = 65, and Unicode (usually stored as
UTF-8) extends this to every writing system and to emoji. Colours are
three bytes (red, green, blue), so #FF8800 is orange. Meaning comes from
interpretation: the same bytes can be a number, a letter, a pixel or an
instruction.`,
				Diagram: `bit value:  128  64  32  16   8   4   2   1
bits:         0   1   0   0   0   0   0   1   = 64 + 1 = 65 = 'A'`,
				MentalModel: "Bits have no meaning until you decide how to read them.",
				TryIt:       "Convert your age and your birth year to binary and hex by hand, then check with Python's bin() and hex().",
				Analogy: `Morse code turns letters into dots and dashes; binary turns everything
into 0s and 1s. A row of light switches, each on or off, can encode any
number, as long as everyone agrees what each switch is worth.`,
				Example: `Web colour codes such as #1DA1F2 are hex RGB values. An IPv4 address
such as 192.168.1.1 is a 32-bit number written as four bytes. The emoji
😀 is Unicode code point U+1F600, stored as 4 bytes in UTF-8.`,
				Exercises: trio(
					"Convert 13, 100 and 255 to binary, and convert 0x2A and 0x10 to decimal.",
					"Repeatedly subtract the largest power of 2 that fits.",
					"Write to_binary(n) and from_binary(s) without built-in conversions, and test them against bin() and int(s, 2) for every number from 0 to 10,000.",
					"Repeated division by 2 gives the bits from right to left.",
					`Open a small PNG and a text file in a hex viewer (xxd on Linux/macOS, or an online hex viewer) and identify the PNG's "magic number" header and any readable text in either file.`,
					`PNG files start with 89 50 4E 47; the bytes 50 4E 47 spell "PNG".`,
				),
			},
			{
				Name:    "Logic Gates & Circuits",
				Summary: "Tiny switches combined into circuits that compute.",
				Body: `A transistor is an electrically controlled switch. Wire a few together
and you get logic gates: NOT, AND, OR, XOR and NAND. NAND alone is
universal, meaning any circuit can be built from NAND gates only.

Combinational circuits compute their outputs from their current inputs.
A half adder uses XOR for the sum bit and AND for the carry. Chain full
adders together and you can add 64-bit numbers. A multiplexer selects
one of several inputs. Modern processors contain tens of billions of
transistors.`,
				Diagram: `A ─┬─────▶ [ XOR ] ──▶ sum        A B │ sum carry
B ─┼─┬───▶ [     ]                0 0 │  0    0
   │ │                            0 1 │  1    0
   └─┼───▶ [ AND ] ──▶ carry      1 0 │  1    0
     └───▶ [     ]                1 1 │  0    1   (1 + 1 = 10 in binary)`,
				MentalModel: "Arithmetic is just logic, and logic is just switches.",
				TryIt:       "Build a half adder and a full adder in a free logic simulator (Digital or Logisim Evolution), then chain four into a 4-bit adder.",
				Analogy: `Two light switches wired one after the other (both must be on for the
bulb to light) make AND. Wired side by side (either one will do), they
make OR. A hallway light with a switch at each end, where flipping either
switch toggles the light, behaves like XOR.`,
				Example: `Every chip in your phone and laptop is built from gates like these, tens
of billions of transistors at a time. Minecraft players have even built
working computers out of redstone logic gates.`,
				Exercises: trio(
					"Build NOT, AND and OR using only NAND gates, on paper.",
					"NOT A = A NAND A. Then AND = NOT(A NAND B).",
					"Write a program that simulates gates as functions, builds a full adder from them, and adds two 8-bit numbers bit by bit (ripple carry). Verify it against normal addition for every pair from 0 to 255.",
					"The carry out of bit i is the carry into bit i + 1.",
					"Design the logic for a car's seat-belt warning: beep if the engine is on AND (the driver's belt is off OR (the passenger seat is occupied AND the passenger's belt is off)). Write the truth table, simplify, and simulate it.",
					"Name each sensor as a Boolean input, then list all 16 cases.",
				),
			},
			{
				Name:    "Memory: Latches, Registers & RAM",
				Summary: "How circuits remember.",
				Body: `Combinational circuits forget as soon as their inputs change. Feed a
gate's output back into its input and you get a latch, which holds one
bit. A flip-flop updates only on a clock tick, keeping the whole chip in
lockstep. Groups of flip-flops form registers (for example, 64 bits).

SRAM (used in CPU caches) keeps each bit in a small flip-flop circuit of
about six transistors: fast but large. DRAM (main memory) stores each bit
as charge in a tiny capacitor that leaks, so it must be refreshed many
times a second: dense and cheap, but slower. Flash memory (in SSDs)
traps charge so that it survives power-off.`,
				MentalModel: "Feedback turns logic into memory, and the clock keeps it in step.",
				TryIt:       "Build an SR latch from two NOR gates in a simulator and watch it hold its value.",
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
					"Which kinds of memory need power to keep their data?",
					"Simulate an SR latch in code (update the two NOR gates repeatedly until the outputs are stable) and show that set, reset and hold all behave correctly.",
					"Loop until the outputs stop changing.",
					`Laptops advertise "16 GB RAM, 512 GB SSD". Explain to a friend what each one is for, then measure how much RAM your browser uses with 1, 10 and 30 tabs open.`,
					"Use Task Manager (Windows), Activity Monitor (macOS) or htop (Linux).",
				),
			},
			{
				Name:    "The CPU: Datapath & Instruction Cycle",
				Summary: "Fetch, decode, execute, billions of times per second.",
				Body: `A CPU repeats one cycle. It fetches the next instruction from memory (at
the address held in the program counter), decodes it, executes it in the
ALU (arithmetic logic unit) using registers, and writes back the result.

Instructions are just numbers, defined by the instruction set
architecture (ISA), such as x86-64, ARM or RISC-V. Branch instructions
change the program counter, which is how ifs and loops exist in
hardware. The clock (for example 3 GHz, or 3 billion ticks a second)
sets the pace, and modern cores complete several instructions per tick.`,
				Diagram: `   ┌───────┐    ┌────────┐    ┌─────────┐    ┌───────────┐
┌─▶│ FETCH │───▶│ DECODE │───▶│ EXECUTE │───▶│ WRITEBACK │───┐
│  └───────┘    └────────┘    └─────────┘    └───────────┘   │
│   PC → memory   which op?     ALU / memory   result → reg  │
└──────────────── PC = next instruction ◀────────────────────┘`,
				MentalModel: "A CPU is a loop that reads a number, does what the number says, and moves on.",
				TryIt:       "Write a program for the Little Man Computer (a toy CPU simulator) that adds two inputs.",
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
					"Write an emulator for a tiny CPU with 4 registers and the instructions LOAD, ADD, SUB, JUMP_IF_ZERO, PRINT and HALT. Then write a program for it that counts down from 10.",
					"A while loop over a program counter, with an if/elif on the opcode.",
					"Look up the CPU in your laptop or phone: its ISA, core count, clock speed and cache sizes. Explain which of those numbers matter most for gaming, video editing and web browsing.",
					"lscpu on Linux, About This Mac on macOS, Task Manager → Performance on Windows.",
				),
			},
			{
				Name:    "Pipelining, Branch Prediction & Parallelism",
				Summary: "The tricks that make modern CPUs fast.",
				Body: `Pipelining overlaps the stages of instructions like an assembly line:
while one instruction executes, the next is decoded and a third is
fetched. Branch prediction guesses which way an if will go, to keep the
pipeline full; a wrong guess wastes roughly 15 to 20 cycles.
Out-of-order execution runs independent instructions early, and caches
(covered again in Stage 5) hide slow memory.

From the mid-2000s, clock speeds stopped rising because of heat, so
chips added more cores instead. Software must run in parallel to use
them, and Amdahl's law says the part that must run serially limits the
overall speed-up.`,
				Diagram: `time →      1    2    3    4    5    6
instr 1:   [F]  [D]  [E]  [W]
instr 2:        [F]  [D]  [E]  [W]
instr 3:             [F]  [D]  [E]  [W]     one instruction finishes per tick`,
				MentalModel: "Keep every part of the chip busy: overlap, predict and parallelise.",
				TryIt:       "Time summing only the values ≥ 128 of a big array, sorted versus unsorted, and explain the difference.",
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
					"Run the sorted-versus-unsorted branch experiment in C, Go or Rust (Python hides the effect) and report the times.",
					"Use a large array (10 million or more elements) and repeat each run several times.",
					"Take a slow batch job (for example resizing a folder of 500 photos) and split it across all your CPU cores with multiprocessing or goroutines. Measure the speed-up and explain why it is less than the number of cores.",
					"Look for the serial parts: listing files, reading from disk, merging results.",
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
		},
	},
}
