package main

// Connections turn isolated facts into a web of knowledge, which is one of
// the strongest aids to long-term memory (elaboration). Every concept lists
// related concepts in other stages and one precise "go deeper" pointer into
// a resource from its stage's library or reading list. Pointers name
// chapters or lectures by title rather than guessing page numbers.

// links holds a concept's cross-references.
type links struct {
	Related []string // names of concepts elsewhere in the curriculum
	Read    string   // where to go deeper, in a verified resource
}

// stagePrereqs lists the stages each stage builds on.
var stagePrereqs = map[int][]int{
	1: {}, 2: {1}, 3: {1}, 4: {1, 3},
	5: {1, 2, 4}, 6: {5}, 7: {2, 6}, 8: {6},
	9: {1}, 10: {3, 8}, 11: {2, 3}, 12: {1, 5},
	13: {3}, 14: {3, 13}, 15: {13, 14}, 16: {5, 6, 15},
}

// conceptLinks is keyed by Concept.Name.
var conceptLinks = map[string]links{
	// ---- Stage 1
	"Values, Types & Variables": {[]string{"Integers & Floating Point in Bits", "Type Systems"},
		"CS50: the first weeks on data types and variables. Step through assignments in Python Tutor to watch each variable change."},
	"Control Flow: Decisions & Loops": {[]string{"Logic & Boolean Algebra", "The CPU: Datapath & Instruction Cycle"},
		"CS50: the lectures on conditions and loops. Trace your own loops in Python Tutor."},
	"Functions & Decomposition": {[]string{"Registers, the Stack & Calling Conventions", "Software Design & Abstraction", "Proof & Induction"},
		"How to Design Programs: the design recipe for functions. SICP chapter 1, Building Abstractions with Procedures."},
	"Collections: Lists, Dictionaries & Strings": {[]string{"Arrays, Linked Lists, Stacks & Queues", "Hash Tables"},
		"CS50: the lectures on arrays and data structures. Exercism's exercises on lists and maps."},
	"Debugging & Reading Errors": {[]string{"Testing & Reliability", "The Command Line & Your Environment"},
		"CS50: its debugging sections. Step through failing code in Python Tutor."},

	// ---- Stage 2
	"Big-O: Measuring Growth": {[]string{"Counting & Combinatorics", "Complexity Classes: P, NP & Hard Problems"},
		"MIT 6.006: the opening lectures on asymptotic analysis. Grokking Algorithms, chapter 1 (binary search and Big O)."},
	"Arrays, Linked Lists, Stacks & Queues": {[]string{"The Memory Hierarchy & Locality", "Registers, the Stack & Calling Conventions"},
		"Sedgewick & Wayne's Algorithms: the section on bags, queues and stacks. Animate them on VisuAlgo."},
	"Hash Tables": {[]string{"Hashing & Password Storage", "B-Trees"},
		"Sedgewick & Wayne's Algorithms: the section on hash tables. VisuAlgo's hash-table animation."},
	"Trees, Heaps & Sorting": {[]string{"B-Trees", "Proof & Induction", "Trees, Ensembles & Clustering"},
		"Sedgewick & Wayne's Algorithms: the sorting chapter and priority queues. VisuAlgo's sorting and heap animations."},
	"Graphs, BFS/DFS & Shortest Paths": {[]string{"Sets, Functions & Relations", "Complexity Classes: P, NP & Hard Problems", "Computational Graphs & Reverse-Mode Autodiff"},
		"Jeff Erickson's Algorithms: the chapters on graph search and shortest paths. VisuAlgo for BFS and DFS."},

	// ---- Stage 3
	"Logic & Boolean Algebra": {[]string{"Logic Gates & Circuits", "Control Flow: Decisions & Loops"},
		"Book of Proof: the chapter on logic. The MIT 6.042J notes on propositions."},
	"Sets, Functions & Relations": {[]string{"SQL & the Relational Model", "Hash Tables"},
		"Book of Proof: the chapters on sets, relations and functions."},
	"Proof & Induction": {[]string{"Functions & Decomposition", "Trees, Heaps & Sorting", "The Halting Problem & Undecidability"},
		"Book of Proof: the chapters on direct proof, proof by contradiction and mathematical induction."},
	"Counting & Combinatorics": {[]string{"Big-O: Measuring Growth", "Hashing & Password Storage"},
		"Book of Proof: the chapter on counting. MIT 6.042J's counting lectures."},
	"Probability for Programmers": {[]string{"Statistical Inference & Experiments", "Probability, Softmax & Cross-Entropy"},
		"Seeing Theory: Basic Probability and Compound Probability. Khan Academy's probability unit."},

	// ---- Stage 4
	"Binary & Number Representation": {[]string{"Integers & Floating Point in Bits", "Values, Types & Variables"},
		"Petzold's Code: the early chapters on binary and hexadecimal, with the companion site's interactive figures."},
	"Logic Gates & Circuits": {[]string{"Logic & Boolean Algebra", "Binary & Number Representation"},
		"Nand2Tetris: project 1 (Boolean logic) and project 2 (Boolean arithmetic)."},
	"Memory: Latches, Registers & RAM": {[]string{"The Memory Hierarchy & Locality", "Durability: Page Cache, fsync & Atomic Rename"},
		"Nand2Tetris: project 3 (memory: flip-flops, registers and RAM)."},
	"The CPU: Datapath & Instruction Cycle": {[]string{"The Compilation Pipeline", "Interpreters, Compilers & Runtimes"},
		"Nand2Tetris: projects 4 and 5 (machine language and computer architecture). Ben Eater's 8-bit computer videos."},
	"Pipelining, Branch Prediction & Parallelism": {[]string{"The Memory Hierarchy & Locality", "Scheduling & Context Switches", "The GPU Execution Model"},
		"Patterson & Hennessy's Computer Organization and Design: the chapters on pipelining and the memory hierarchy."},

	// ---- Stage 5
	"The Compilation Pipeline": {[]string{"Parsing: Recursive Descent & Pratt", "Interpreters, Compilers & Runtimes", "The CPU: Datapath & Instruction Cycle"},
		"CS:APP chapter 7 (Linking). Crafting Interpreters: the chapter A Map of the Territory."},
	"Registers, the Stack & Calling Conventions": {[]string{"Functions & Decomposition", "Processes & the Address Space"},
		"CS:APP chapter 3 (Machine-Level Representation of Programs), especially its section on procedures and the stack."},
	"Integers & Floating Point in Bits": {[]string{"Binary & Number Representation", "Values, Types & Variables"},
		"CS:APP chapter 2 (Representing and Manipulating Information)."},
	"The Memory Hierarchy & Locality": {[]string{"The Roofline Model: Compute- vs Memory-Bound", "GPU Memory Hierarchy & Coalescing", "Memory: Latches, Registers & RAM"},
		"CS:APP chapter 6 (The Memory Hierarchy) and its cache lab. Drepper's memory paper for the full depth."},
	"Parsing: Recursive Descent & Pratt": {[]string{"Grammars & Pushdown Automata", "The Compilation Pipeline"},
		"Crafting Interpreters: the chapters Representing Code and Parsing Expressions."},

	// ---- Stage 6
	"Processes & the Address Space": {[]string{"Registers, the Stack & Calling Conventions", "Scheduling & Context Switches"},
		"OSTEP: the chapters on the process abstraction and the process API (fork, exec, wait)."},
	"Virtual Memory & Paging": {[]string{"KV Cache, Batching & PagedAttention", "Memory: Latches, Registers & RAM"},
		"OSTEP: the chapters on address spaces, paging and TLBs."},
	"System Calls & Privilege": {[]string{"The Command Line & Your Environment", "Sockets & I/O Multiplexing"},
		"OSTEP: the chapter on limited direct execution. The Linux Programming Interface on system calls."},
	"Scheduling & Context Switches": {[]string{"Concurrency Models", "Pipelining, Branch Prediction & Parallelism"},
		"OSTEP: the scheduling chapters (multi-level feedback queue and proportional share)."},
	"Synchronisation & Memory Ordering": {[]string{"Concurrency Models", "Transactions, Isolation & MVCC"},
		"OSTEP: the concurrency chapters on locks, condition variables and common concurrency bugs."},

	// ---- Stage 7
	"Durability: Page Cache, fsync & Atomic Rename": {[]string{"Memory: Latches, Registers & RAM", "Write-Ahead Logging & Recovery"},
		"Build Your Own Database: its chapters on durability and fsync. DDIA chapter 3 (Storage and Retrieval)."},
	"B-Trees": {[]string{"Trees, Heaps & Sorting", "SQL & the Relational Model"},
		"CMU 15-445: the lectures on B+tree indexes. Use The Index, Luke for how indexes behave from SQL."},
	"LSM-Trees": {[]string{"Write-Ahead Logging & Recovery", "B-Trees"},
		"The LevelDB implementation notes. DDIA chapter 3 on SSTables and LSM-trees."},
	"Write-Ahead Logging & Recovery": {[]string{"Durability: Page Cache, fsync & Atomic Rename", "Consensus & Replication (Raft)"},
		"CMU 15-445: the lectures on database logging and recovery (including ARIES)."},
	"Transactions, Isolation & MVCC": {[]string{"Synchronisation & Memory Ordering", "Time, Clocks & Ordering"},
		"DDIA chapter 7 (Transactions). CMU 15-445's concurrency-control lectures."},
	"SQL & the Relational Model": {[]string{"Sets, Functions & Relations", "B-Trees", "Web Vulnerabilities: Injection & XSS"},
		"Use The Index, Luke. CMU 15-445's opening lectures on the relational model and SQL."},

	// ---- Stage 8
	"The Layered Stack & Encapsulation": {[]string{"HTTP, DNS & How the Web Works", "Binary & Number Representation"},
		"High Performance Browser Networking: the Networking 101 chapters on latency, TCP and UDP."},
	"TCP: Handshake, Flow & Congestion Control": {[]string{"Sockets & I/O Multiplexing", "TLS & Authentication on the Web"},
		"High Performance Browser Networking: Building Blocks of TCP. TCP/IP Illustrated for the full detail."},
	"Sockets & I/O Multiplexing": {[]string{"System Calls & Privilege", "Concurrency Models"},
		"Beej's Guide to Network Programming, from socket() through to poll()."},
	"Time, Clocks & Ordering": {[]string{"Transactions, Isolation & MVCC", "Consensus & Replication (Raft)"},
		"Lamport's Time, Clocks, and the Ordering of Events (on your reading list). The MIT 6.5840 lectures."},
	"Consensus & Replication (Raft)": {[]string{"Write-Ahead Logging & Recovery", "Time, Clocks & Ordering"},
		"The Raft paper and visualisation at raft.github.io. The MIT 6.5840 Raft labs."},
	"HTTP, DNS & How the Web Works": {[]string{"TLS & Authentication on the Web", "Web Vulnerabilities: Injection & XSS", "The Layered Stack & Encapsulation"},
		"High Performance Browser Networking: the HTTP chapters. Watch real requests in your browser's developer tools."},

	// ---- Stage 9
	"The Command Line & Your Environment": {[]string{"Processes & the Address Space", "System Calls & Privilege"},
		"The Missing Semester: the lectures on the shell, shell tools and scripting."},
	"Version Control with Git": {[]string{"Hashing & Password Storage", "Graphs, BFS/DFS & Shortest Paths"},
		"Pro Git chapters 1–3 (Getting Started, Git Basics, Git Branching). Practise on Learn Git Branching."},
	"Testing & Reliability": {[]string{"Debugging & Reading Errors", "Proof & Induction"},
		"The Missing Semester: the Metaprogramming lecture (build systems, continuous integration and tests)."},
	"Software Design & Abstraction": {[]string{"Functions & Decomposition", "Paradigms: Imperative, Object-Oriented & Functional"},
		"Ousterhout's A Philosophy of Software Design (on your reading list). Refactoring.Guru for patterns."},
	"From Code to Production": {[]string{"Testing & Reliability", "Thinking Like an Attacker: Threat Modelling"},
		"Google's SRE book: the chapters on monitoring distributed systems and postmortem culture. The Twelve-Factor App."},

	// ---- Stage 10
	"Thinking Like an Attacker: Threat Modelling": {[]string{"Software Design & Abstraction", "From Code to Production"},
		"Anderson's Security Engineering (on your reading list): the early chapters on opponents and psychology. OWASP Top 10:2025."},
	"Web Vulnerabilities: Injection & XSS": {[]string{"SQL & the Relational Model", "HTTP, DNS & How the Web Works"},
		"PortSwigger Web Security Academy: the SQL injection and cross-site scripting topics."},
	"Hashing & Password Storage": {[]string{"Hash Tables", "Counting & Combinatorics"},
		"Crypto 101: the chapter on hash functions, including its section on password storage."},
	"Encryption: Symmetric & Public-Key": {[]string{"Counting & Combinatorics", "Complexity Classes: P, NP & Hard Problems"},
		"Crypto 101: the chapters on block ciphers, public-key encryption and signatures. Cryptopals set 1."},
	"TLS & Authentication on the Web": {[]string{"HTTP, DNS & How the Web Works", "Encryption: Symmetric & Public-Key"},
		"Crypto 101: the chapter on SSL and TLS. High Performance Browser Networking: the chapter on TLS."},

	// ---- Stage 11
	"Finite Automata & Regular Expressions": {[]string{"Parsing: Recursive Descent & Pratt", "Logic & Boolean Algebra"},
		"MIT 18.404J: the first lectures on finite automata and regular expressions. Practise on regex101."},
	"Grammars & Pushdown Automata": {[]string{"Parsing: Recursive Descent & Pratt", "Arrays, Linked Lists, Stacks & Queues"},
		"MIT 18.404J: the lectures on context-free grammars and pushdown automata. Simulate them in JFLAP."},
	"Turing Machines & Computability": {[]string{"The CPU: Datapath & Instruction Cycle", "Interpreters, Compilers & Runtimes"},
		"MIT 18.404J: the lectures on Turing machines and the Church–Turing thesis. Turing's 1936 paper (on your reading list)."},
	"The Halting Problem & Undecidability": {[]string{"Proof & Induction", "Testing & Reliability"},
		"MIT 18.404J: the lectures on decidability and undecidability."},
	"Complexity Classes: P, NP & Hard Problems": {[]string{"Big-O: Measuring Growth", "Graphs, BFS/DFS & Shortest Paths", "Encryption: Symmetric & Public-Key"},
		"MIT 18.404J: the lectures on time complexity and NP-completeness. The Complexity Zoo's Petting Zoo for a gentle tour."},

	// ---- Stage 12
	"Paradigms: Imperative, Object-Oriented & Functional": {[]string{"Functions & Decomposition", "Software Design & Abstraction"},
		"Dan Grossman's Programming Languages course (functional programming in ML and Racket, OOP in Ruby). Hughes' Why Functional Programming Matters."},
	"Type Systems": {[]string{"Values, Types & Variables", "Logic & Boolean Algebra"},
		"The Rust Book: the chapter on enums and pattern matching. Grossman's course on static versus dynamic typing."},
	"Memory Management": {[]string{"Processes & the Address Space", "Virtual Memory & Paging"},
		"The Rust Book: the chapter on ownership. Crafting Interpreters: the chapter on garbage collection."},
	"Interpreters, Compilers & Runtimes": {[]string{"The Compilation Pipeline", "The CPU: Datapath & Instruction Cycle"},
		"Crafting Interpreters, part III: A Bytecode Virtual Machine."},
	"Concurrency Models": {[]string{"Synchronisation & Memory Ordering", "Sockets & I/O Multiplexing", "The GPU Execution Model"},
		"The Rust Book: the chapter on fearless concurrency."},

	// ---- Stage 13
	"Matrices as Linear Maps": {[]string{"Attention & the Transformer", "Sets, Functions & Relations"},
		"3Blue1Brown's Essence of Linear Algebra: the chapters on linear transformations and matrix multiplication."},
	"Eigenvectors & the SVD": {[]string{"Matrices as Linear Maps", "Trees, Ensembles & Clustering"},
		"3Blue1Brown on eigenvectors and eigenvalues. Gilbert Strang's 18.06 lectures on the SVD. The SVD section of Linear Algebra Done Right."},
	"Gradients, Jacobians & the Matrix Chain Rule": {[]string{"Computational Graphs & Reverse-Mode Autodiff", "Matrices as Linear Maps"},
		"Parr & Howard's The Matrix Calculus You Need for Deep Learning. The Matrix Cookbook's derivatives section."},
	"Probability, Softmax & Cross-Entropy": {[]string{"Probability for Programmers", "Linear & Logistic Regression"},
		"Mathematics for Machine Learning: the chapter Probability and Distributions."},
	"Optimisation: From Gradient Descent to Adam": {[]string{"Linear & Logistic Regression", "The Training Loop & Generalisation"},
		"Mathematics for Machine Learning: the chapter Continuous Optimization."},

	// ---- Stage 14
	"Descriptive Statistics & Distributions": {[]string{"Probability for Programmers", "Evaluating Models Honestly"},
		"Seeing Theory: Basic Probability and Probability Distributions. StatQuest's videos on histograms and distributions."},
	"Statistical Inference & Experiments": {[]string{"Probability for Programmers", "Testing & Reliability"},
		"Seeing Theory: Frequentist Inference. StatQuest on p-values and statistical power."},
	"Linear & Logistic Regression": {[]string{"Optimisation: From Gradient Descent to Adam", "Probability, Softmax & Cross-Entropy"},
		"An Introduction to Statistical Learning: the chapters Linear Regression and Classification."},
	"Trees, Ensembles & Clustering": {[]string{"Trees, Heaps & Sorting", "Eigenvectors & the SVD"},
		"An Introduction to Statistical Learning: the chapters Tree-Based Methods and Unsupervised Learning."},
	"Evaluating Models Honestly": {[]string{"Statistical Inference & Experiments", "Testing & Reliability", "The Training Loop & Generalisation"},
		"An Introduction to Statistical Learning: the chapter Resampling Methods (cross-validation). The scikit-learn guide on model evaluation."},

	// ---- Stage 15
	"Computational Graphs & Reverse-Mode Autodiff": {[]string{"Gradients, Jacobians & the Matrix Chain Rule", "Graphs, BFS/DFS & Shortest Paths"},
		"Karpathy's Zero to Hero: the micrograd lecture. The CS231n notes on backpropagation."},
	"MLPs, Activations & Initialisation": {[]string{"Matrices as Linear Maps", "Linear & Logistic Regression"},
		"Karpathy's Zero to Hero: the makemore lectures on MLPs, and on activations and gradients. The CS231n neural-network notes."},
	"Residual Connections & Normalisation": {[]string{"MLPs, Activations & Initialisation", "Descriptive Statistics & Distributions"},
		"Karpathy's Zero to Hero: the makemore lecture on activations, gradients and BatchNorm. The Deep Learning book's chapter on optimisation."},
	"Attention & the Transformer": {[]string{"Matrices as Linear Maps", "Probability, Softmax & Cross-Entropy"},
		"The Illustrated Transformer. Karpathy's Let's build GPT lecture. Attention Is All You Need."},
	"The Training Loop & Generalisation": {[]string{"Evaluating Models Honestly", "Optimisation: From Gradient Descent to Adam"},
		"The CS231n notes on training neural networks (learning and evaluation). Karpathy's nanoGPT training script."},

	// ---- Stage 16
	"The GPU Execution Model": {[]string{"Pipelining, Branch Prediction & Parallelism", "Concurrency Models"},
		"The CUDA C++ Programming Guide: the programming-model and hardware-implementation chapters. GPU MODE's early lectures."},
	"GPU Memory Hierarchy & Coalescing": {[]string{"The Memory Hierarchy & Locality", "Memory: Latches, Registers & RAM"},
		"Simon Boehm's CUDA matmul worklog: the steps on global-memory coalescing and shared-memory blocking."},
	"The Roofline Model: Compute- vs Memory-Bound": {[]string{"The Memory Hierarchy & Locality", "Big-O: Measuring Growth"},
		"Horace He's Making Deep Learning Go Brrrr From First Principles. The Roofline paper (on your reading list)."},
	"Kernel Fusion & FlashAttention": {[]string{"Attention & the Transformer", "GPU Memory Hierarchy & Coalescing"},
		"The FlashAttention paper. Horace He's post on operator fusion."},
	"KV Cache, Batching & PagedAttention": {[]string{"Virtual Memory & Paging", "Attention & the Transformer"},
		"The PagedAttention (vLLM) paper. llama.cpp's source for a real KV cache."},
}

// stageOfConcept finds which stage a concept belongs to (0 if unknown).
func stageOfConcept(name string) int {
	for id, g := range curriculum {
		for _, c := range g.Concepts {
			if c.Name == name {
				return id
			}
		}
	}
	return 0
}
