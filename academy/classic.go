package main

// Classic Mode: the habits of strong 1980s and 1990s CS learners, kept
// where they help and made explicit:
//   - one anchor book per stage, read cover to cover;
//   - classic texts from the field's history, read in the original;
//   - real, small source code, read like literature;
//   - magazine-style type-in listings: predict, type by hand, run, compare;
//   - productive struggle before hints, recorded in a lab notebook.
// Every URL here was checked against live web search results; the type-in
// listings and their expected output come from running them (typeins.go).

import "time"

// ClassicRead is a book, paper or codebase in a stage's classic corner.
type ClassicRead struct {
	Title  string
	Author string
	Year   int    // 0 when not meaningful (a living codebase)
	URL    string // empty when there is no free official copy
	Why    string
	How    string // how to read it
}

// TypeIn is a short listing to predict, type in by hand, and run.
type TypeIn struct {
	File     string // suggested file name
	Lang     string
	Run      string // command to run it
	Predict  string // what to predict before running
	Lesson   string // what the result teaches
	Code     string
	Expected string // actual output from running the listing
}

// ClassicGuide is one stage's classic corner.
type ClassicGuide struct {
	Anchor  ClassicRead // the book to read cover to cover
	Classic ClassicRead // a classic text from the field's history
	Source  ClassicRead // real code to read
}

// classicFor returns a stage's classic corner and type-in.
func classicFor(stageID int) (ClassicGuide, TypeIn, bool) {
	g, ok := classicGuides[stageID]
	t, ok2 := typeIns[stageID]
	return g, t, ok && ok2
}

// Struggle clock: how long to wrestle with an exercise before its hint
// unlocks in Classic Mode.
var struggleMinutes = map[string]int{
	LevelWarmUp:    10,
	LevelPractice:  30,
	LevelRealWorld: 45,
}

// hintUnlocked reports whether an exercise's hint may be shown in Classic
// Mode, and how long remains otherwise. The clock runs from when the
// exercise was first opened; writing down what you tried when truly stuck
// unlocks it early.
func hintUnlocked(opened time.Time, level string, stuckLogged bool, now time.Time) (bool, time.Duration) {
	if stuckLogged || opened.IsZero() {
		return stuckLogged, 0
	}
	need := time.Duration(struggleMinutes[level]) * time.Minute
	elapsed := now.Sub(opened)
	if elapsed >= need {
		return true, 0
	}
	return false, need - elapsed
}

// NotebookEntry is one line in the learner's lab notebook.
type NotebookEntry struct {
	At    time.Time `json:"at"`
	Stage int       `json:"stage"`
	Ref   string    `json:"ref"`  // exercise key, or "type-in"
	Kind  string    `json:"kind"` // plan, tried, stuck, observed, prediction
	Text  string    `json:"text"`
}

// Notebook entry kinds.
const (
	NotePlan       = "plan"
	NoteTried      = "tried"
	NoteStuck      = "stuck"
	NoteObserved   = "observed"
	NotePrediction = "prediction"
)

// notebookFor returns a stage's notebook entries, optionally for one ref.
func (r *Registry) notebookFor(stage int, ref string) []NotebookEntry {
	var out []NotebookEntry
	for _, e := range r.Notebook {
		if e.Stage == stage && (ref == "" || e.Ref == ref) {
			out = append(out, e)
		}
	}
	return out
}

// stuckLogged reports whether the learner has recorded being stuck on ref.
func (r *Registry) stuckLogged(ref string) bool {
	for _, e := range r.Notebook {
		if e.Ref == ref && e.Kind == NoteStuck {
			return true
		}
	}
	return false
}

var classicGuides = map[int]ClassicGuide{
	1: {
		Anchor: ClassicRead{"How to Design Programs", "Felleisen, Findler, Flatt & Krishnamurthi", 2001, "https://htdp.org/",
			"A complete, systematic method for turning a problem statement into a working program: the closest thing to a programming apprenticeship in a book.",
			"One section per session, in order. Do every exercise before turning the page; that is where the learning happens."},
		Classic: ClassicRead{"Go To Statement Considered Harmful", "Edsger W. Dijkstra", 1968, "https://homepages.cwi.nl/~storm/teaching/reader/Dijkstra68.pdf",
			"A short letter that helped launch structured programming, which is why your loops and ifs look the way they do.",
			"Read it twice: once for the argument, once asking how it applies to the loops you wrote this week."},
		Source: ClassicRead{"How to Write a Spelling Corrector", "Peter Norvig", 2007, "https://www.norvig.com/spell-correct.html",
			"A genuinely useful program in about 21 lines of Python, explained line by line by one of the field's best writers.",
			"Type the program in yourself, run it on a big text file, then read the essay's explanation of every line."},
	},
	2: {
		Anchor: ClassicRead{"Algorithms", "Jeff Erickson", 2019, "https://jeffe.cs.illinois.edu/teaching/algorithms/",
			"A free, rigorous and witty textbook, written for exactly this stage.",
			"Read the chapters on recursion, backtracking, dynamic programming and graphs in order; attempt the exercises with pen and paper."},
		Classic: ClassicRead{"A Note on Two Problems in Connexion with Graphs", "Edsger W. Dijkstra", 1959, "https://ir.cwi.nl/pub/9256/9256D.pdf",
			"Three pages that introduced Dijkstra's shortest-path algorithm, which still routes traffic today.",
			"Compare his description with the version you implemented in the graphs exercise. What did he call things?"},
		Source: ClassicRead{"algs4: the textbook's own code", "Robert Sedgewick & Kevin Wayne", 0, "https://github.com/kevin-wayne/algs4",
			"Clean, well-tested implementations of every classic algorithm.",
			"Read BinarySearch.java, MinPQ.java and DijkstraSP.java, and compare them line by line with your own versions."},
	},
	3: {
		Anchor: ClassicRead{"Book of Proof", "Richard Hammack", 2018, "https://richardhammack.github.io/BookOfProof/",
			"A free, gentle and complete introduction to logic, sets, counting and proof.",
			"One chapter a week. Do the odd-numbered exercises and check them against the solutions at the back."},
		Classic: ClassicRead{"Why Numbering Should Start at Zero (EWD 831)", "Edsger W. Dijkstra", 1982, "https://www.cs.utexas.edu/~EWD/transcriptions/EWD08xx/EWD831.html",
			"A short, proof-style argument about an everyday programming choice, which shows how precise reasoning settles debates.",
			"Reconstruct his argument from memory afterwards, in your own words."},
		Source: ClassicRead{"A Concrete Introduction to Probability (using Python)", "Peter Norvig", 2016, "https://github.com/norvig/pytudes/blob/main/ipynb/Probability.ipynb",
			"Probability built from counting, with short, exact Python for every idea.",
			"Run each cell, predict each answer before you see it, and redo the dice and card examples yourself."},
	},
	4: {
		Anchor: ClassicRead{"Code: The Hidden Language of Computer Hardware and Software", "Charles Petzold", 1999, "https://www.codehiddenlanguage.com/",
			"From flashlights and Morse code to a working CPU, one small step at a time. The second edition (2022) has an interactive companion site.",
			"Read it straight through like a novel, and rebuild each circuit in a logic simulator as you go."},
		Classic: ClassicRead{"First Draft of a Report on the EDVAC", "John von Neumann", 1945, "https://archive.org/details/vnedvac",
			"The first published description of the stored-program computer, the design nearly every computer still follows.",
			"Read the opening sections on the main organs of the machine and spot the CPU, memory and I/O you now know."},
		Source: ClassicRead{"Visual 6502", "visual6502.org team", 0, "https://github.com/trebonian/visual6502",
			"A transistor-level simulation of the 6502, the chip family behind the Apple II, Commodore 64 and NES, built from photographs of the real silicon.",
			"Run it in a browser, single-step a few instructions, and watch the transistors switch."},
	},
	5: {
		Anchor: ClassicRead{"Computer Systems: A Programmer's Perspective", "Bryant & O'Hallaron", 2015, "https://csapp.cs.cmu.edu/",
			"The programmer's view of the whole machine: bits, assembly, caches, linking. Its labs are legendary.",
			"Chapters 1–3 and 6–7 in order, doing the data, bomb and cache labs."},
		Classic: ClassicRead{"Reflections on Trusting Trust", "Ken Thompson", 1984, "https://www.cs.cmu.edu/~rdriley/487/papers/Thompson_1984_ReflectionsonTrustingTrust.pdf",
			"Thompson's Turing Award lecture: a compiler that secretly inserts a backdoor into programs, including into new copies of itself. Three pages that changed how we think about trust.",
			"Read it, then explain to someone why reading the source code of a program is not enough to trust it."},
		Source: ClassicRead{"chibicc: a small C compiler", "Rui Ueyama", 0, "https://github.com/rui314/chibicc",
			"Every commit adds one feature, carefully written to be read, from a compiler for a single number up to C11.",
			"Start at the first commit and step forward; after each commit, predict what the next one will add."},
	},
	6: {
		Anchor: ClassicRead{"Operating Systems: Three Easy Pieces", "Remzi & Andrea Arpaci-Dusseau", 2018, "https://pages.cs.wisc.edu/~remzi/OSTEP/",
			"A free, conversational and deep OS textbook, organised around virtualisation, concurrency and persistence.",
			"Read it in order, running the homework simulators at the end of each chapter."},
		Classic: ClassicRead{"The UNIX Time-Sharing System", "Dennis M. Ritchie & Ken Thompson", 1974, "https://sipb.mit.edu/iap/6.828/readings/ritchie78unix.pdf",
			"The paper that introduced Unix to the world. (This copy is the revised 1978 version.)",
			"List every design idea in it that you still use today: files, processes, pipes, the shell…"},
		Source: ClassicRead{"xv6: a teaching Unix", "MIT PDOS", 0, "https://github.com/mit-pdos/xv6-riscv",
			"A modern re-implementation of 1975's Unix Version 6, small enough to read completely.",
			"Read kernel/proc.c (processes), kernel/vm.c (page tables) and kernel/trap.c (system calls), then add a system call of your own."},
	},
	7: {
		Anchor: ClassicRead{"Designing Data-Intensive Applications", "Martin Kleppmann", 2017, "https://dataintensive.net/",
			"The map of storage, replication and consistency. A second edition was published in 2026.",
			"Read Part I closely; skim Parts II and III first, then return to them after Stage 8."},
		Classic: ClassicRead{"A Relational Model of Data for Large Shared Data Banks", "E. F. Codd", 1970, "https://rebelsky.cs.grinnell.edu/Courses/CS302/2007S/Readings/codd-1970.pdf",
			"The paper that invented relational databases, and with them the ideas behind SQL.",
			"Read sections 1 and 2; translate each of Codd's examples into a modern SQL table."},
		Source: ClassicRead{"SQLite's source tree (official mirror)", "D. Richard Hipp and contributors", 0, "https://github.com/sqlite/sqlite/tree/master/src",
			"One of the most widely deployed programs in the world, and famously well commented.",
			"Read the opening comment of btreeInt.h, which describes the on-disk B-tree format, then browse btree.c."},
	},
	8: {
		Anchor: ClassicRead{"High Performance Browser Networking", "Ilya Grigorik", 2013, "https://hpbn.co/",
			"Free, practical and precise about latency, TCP, TLS and HTTP.",
			"Read Part I (Networking 101) in full, then the HTTP chapters."},
		Classic: ClassicRead{"End-to-End Arguments in System Design", "Saltzer, Reed & Clark", 1984, "https://web.mit.edu/saltzer/www/publications/endtoend/endtoend.pdf",
			"Why the internet keeps the network simple and puts the smarts at the edges. One of the most cited systems papers ever.",
			"After reading, find one end-to-end check in an app you use, such as a download checksum."},
		Source: ClassicRead{"Redis's event loop (ae.c)", "Salvatore Sanfilippo and contributors", 0, "https://github.com/redis/redis/blob/unstable/src/ae.c",
			"A small, readable event loop that lets one thread serve thousands of connections.",
			"Read aeProcessEvents and follow a readable socket from epoll to its callback."},
	},
	9: {
		Anchor: ClassicRead{"The Pragmatic Programmer", "Andrew Hunt & David Thomas", 1999, "",
			"The craftsman's handbook: plain text, automation, testing, and taking responsibility for your code. (The 20th-anniversary edition is from 2019.)",
			"One short section a day, and apply one tip to a real project before reading the next."},
		Classic: ClassicRead{"No Silver Bullet: Essence and Accidents of Software Engineering", "Frederick P. Brooks, Jr.", 1986, "https://www.cs.unc.edu/techreports/86-020.pdf",
			"Why no tool or method gives a tenfold productivity jump. Still the sharpest essay on why software is hard.",
			"Sort the difficulties you have met this month into Brooks's essence versus accident."},
		Source: ClassicRead{"Git's first commit", "Linus Torvalds", 2005, "https://github.com/git/git/commit/e83c5163316f89bfbde7d9ab23ca2e25604af290",
			"\"Initial revision of git, the information manager from hell\": about 1,000 lines, and already content-addressed storage.",
			"Read the README and the tiny C files in that commit; find where an object's hash becomes its file name."},
	},
	10: {
		Anchor: ClassicRead{"Crypto 101", "Laurens Van Houtven", 0, "https://www.crypto101.io/",
			"A free book that teaches cryptography by breaking it.",
			"Read it in order, and do the matching Cryptopals challenges as you go."},
		Classic: ClassicRead{"Smashing the Stack for Fun and Profit", "Aleph One (Elias Levy)", 1996, "https://phrack.org/issues/49/14",
			"The Phrack article that taught a generation how buffer overflows work, and why memory safety matters.",
			"Read it alongside your Stage 5 notes on the stack; then find which modern defences (stack canaries, non-executable stacks, ASLR) break each step."},
		Source: ClassicRead{"OWASP Juice Shop", "Bjoern Kimminich and contributors", 0, "https://github.com/juice-shop/juice-shop",
			"A deliberately insecure web shop, so its source is a catalogue of real vulnerabilities.",
			"After solving a challenge, find the code responsible in the repository and write the fix."},
	},
	11: {
		Anchor: ClassicRead{"Introduction to the Theory of Computation", "Michael Sipser", 2012, "",
			"The clearest book on automata, computability and complexity. The MIT 18.404J lectures follow it.",
			"Watch the matching 18.404J lecture, then read the chapter and do its exercises."},
		Classic: ClassicRead{"Regular Expression Matching Can Be Simple And Fast", "Russ Cox", 2007, "https://swtch.com/~rsc/regexp/regexp1.html",
			"Explains Ken Thompson's 1968 automaton-based regex algorithm, and why many popular engines are exponentially slower.",
			"Run his pathological example in your language's regex engine and time it."},
		Source: ClassicRead{"A Regular Expression Matcher", "Brian Kernighan (on Rob Pike's code)", 2007, "https://www.cs.princeton.edu/courses/archive/spr09/cos333/beautiful.html",
			"About 30 lines of C by Rob Pike, written in 1998 for The Practice of Programming, and explained by Kernighan.",
			"Compare it with the Python type-in for this stage, then extend yours with + and ?."},
	},
	12: {
		Anchor: ClassicRead{"Crafting Interpreters", "Robert Nystrom", 2021, "https://craftinginterpreters.com/",
			"Two complete interpreters, a tree-walker and a bytecode VM with a garbage collector, built line by line.",
			"Type every line in yourself; never copy. By the end you will have written a language."},
		Classic: ClassicRead{"Recursive Functions of Symbolic Expressions and Their Computation by Machine, Part I", "John McCarthy", 1960, "https://www-formal.stanford.edu/jmc/recursive.html",
			"The paper that introduced Lisp: a language defined in terms of itself in a few pages.",
			"Find eval in the paper and compare it with the evaluate function in this stage's type-in."},
		Source: ClassicRead{"(How to Write a (Lisp) Interpreter (in Python))", "Peter Norvig", 2010, "https://www.norvig.com/lispy.html",
			"A working Scheme interpreter in about 100 lines, explained step by step.",
			"Extend your type-in with define, lambda and if, using Norvig's version only when stuck."},
	},
	13: {
		Anchor: ClassicRead{"Mathematics for Machine Learning", "Deisenroth, Faisal & Ong", 2020, "https://mml-book.github.io/",
			"Free, and written for exactly the maths that ML needs.",
			"Part I in order; skip the proofs on a first pass, but do the worked examples by hand."},
		Classic: ClassicRead{"What Every Computer Scientist Should Know About Floating-Point Arithmetic", "David Goldberg", 1991, "https://docs.oracle.com/cd/E19957-01/806-3568/ncg_goldberg.html",
			"Every matrix computation runs on floats; this is the classic account of where they go wrong.",
			"Read the sections on rounding error and cancellation, then find a case where your power-iteration type-in loses precision."},
		Source: ClassicRead{"How to Optimize GEMM", "Robert van de Geijn and FLAME contributors", 0, "https://github.com/flame/how-to-optimize-gemm/wiki",
			"Matrix multiplication, taken from a naive triple loop to near-peak speed one small step at a time.",
			"Follow the steps in order, predicting each speed-up before measuring it."},
	},
	14: {
		Anchor: ClassicRead{"An Introduction to Statistical Learning", "James, Witten, Hastie & Tibshirani", 2023, "https://www.statlearning.com/",
			"The friendliest serious statistics and ML textbook, free, with labs in Python or R.",
			"One chapter a week, doing its lab by typing the code yourself."},
		Classic: ClassicRead{"Statistical Modeling: The Two Cultures", "Leo Breiman", 2001, "https://projecteuclid.org/journals/statistical-science/volume-16/issue-3/Statistical-Modeling--The-Two-Cultures-with-comments-and-a/10.1214/ss/1009213726.full",
			"The inventor of random forests argues for judging models by prediction, not by assumptions. It shaped modern ML.",
			"Read the paper, then write down which culture each model in this stage belongs to."},
		Source: ClassicRead{"Data Science from Scratch: the book's code", "Joel Grus", 0, "https://github.com/joelgrus/data-science-from-scratch",
			"Statistics and ML implemented in plain Python, with no libraries hiding the ideas.",
			"Read scratch/statistics.py and scratch/linear_algebra.py, then reimplement one function without looking."},
	},
	15: {
		Anchor: ClassicRead{"Deep Learning", "Goodfellow, Bengio & Courville", 2016, "https://www.deeplearningbook.org/",
			"The standard reference for the theory behind neural networks, free online.",
			"Read Part II alongside Karpathy's lectures: the lecture for intuition, the book for precision."},
		Classic: ClassicRead{"Gradient-Based Learning Applied to Document Recognition", "LeCun, Bottou, Bengio & Haffner", 1998, "https://yann.lecun.com/exdb/publis/pdf/lecun-98.pdf",
			"The LeNet paper: convolutional networks reading handwritten digits (and bank cheques) in the 1990s.",
			"Read the introduction and the section describing LeNet-5, then sketch its layers from memory."},
		Source: ClassicRead{"micrograd", "Andrej Karpathy", 2020, "https://github.com/karpathy/micrograd",
			"A complete autograd engine in about 150 lines. Your type-in is a slice of it.",
			"Read micrograd/engine.py, then add tanh and exp to your type-in without looking."},
	},
	16: {
		Anchor: ClassicRead{"Programming Massively Parallel Processors", "Hwu, Kirk & El Hajj", 2022, "",
			"The standard CUDA textbook. The GPU MODE lectures follow it.",
			"Read the early chapters on the execution model and memory, writing each kernel yourself."},
		Classic: ClassicRead{"Hitting the Memory Wall: Implications of the Obvious", "Wulf & McKee", 1995, "https://libraopen.lib.virginia.edu/public_view/k35694375",
			"A five-page 1995 warning that processors would outrun memory. Every memory-bound AI workload today is that prediction coming true.",
			"Read it, then check your roofline type-in: which side of the roofline does LLM decoding sit on?"},
		Source: ClassicRead{"llm.c", "Andrej Karpathy", 2024, "https://github.com/karpathy/llm.c",
			"GPT-2 training in plain C and CUDA, with no framework in the way.",
			"Read train_gpt2.c from top to bottom and match each function to a concept from Stage 15."},
	},
}
