package main

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

	// Beginner layer, filled in from explainers.go.
	Analogy string // the idea in everyday terms, no jargon
	Example string // where this shows up in the real world
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

// guideFor returns the Study Hall content for a stage, if any.
func guideFor(stageID int) (StageGuide, bool) {
	g, ok := curriculum[stageID]
	return g, ok
}

var curriculum = map[int]StageGuide{
	// -----------------------------------------------------------------
	1: {
		Overview: `Everything above this stage rests on how source code becomes machine
instructions, and how those instructions move bytes between registers,
caches and memory. You will build a working model of the machine a C
compiler targets, and of the compiler itself.`,
		Concepts: []Concept{
			{
				Name:    "The Compilation Pipeline",
				Summary: "Source code is lowered step by step until it is bytes the CPU can run.",
				Body: `A compiler is a chain of translations. The lexer turns characters into
tokens. The parser turns tokens into an abstract syntax tree (AST).
Semantic analysis resolves names and checks types. The AST is lowered to
an intermediate representation (IR), which the optimizer rewrites, and
the backend selects instructions and allocates registers to produce
assembly.

The assembler turns assembly into an object file: machine code plus a
symbol table and relocations (holes to patch). The linker merges object
files, resolves symbols across them and fixes up addresses. At run time
the loader maps the executable's segments into memory and jumps to the
entry point.`,
				Diagram: `source.c ─▶ [lexer] ─▶ tokens ─▶ [parser] ─▶ AST ─▶ [IR gen] ─▶ IR
   ─▶ [optimizer] ─▶ [codegen] ─▶ .s ─▶ [assembler] ─▶ .o ─▶ [linker] ─▶ a.out`,
				MentalModel: "Each stage translates to a language that is slightly closer to the metal.",
				TryIt:       "Compile the same function with gcc -S -O0 and -O2 (or paste it into godbolt.org) and explain every instruction that disappeared.",
			},
			{
				Name:    "Registers, the Stack & Calling Conventions",
				Summary: "The ABI is the contract that lets separately compiled code call each other.",
				Body: `The CPU computes on a small set of registers. The calling convention
(on Linux x86-64, the System V AMD64 ABI) fixes where arguments and results go:
- integer arguments in rdi, rsi, rdx, rcx, r8, r9, and the rest on the stack;
- the return value in rax;
- rbx, rbp and r12 to r15 are callee-saved, so a function must restore them;
- the stack must be 16-byte aligned at every call instruction.

The call instruction pushes the return address. A function prologue
reserves a frame for locals and spilled registers. Local buffers and the
return address share the stack, and that is exactly why buffer overflows
can hijack control flow.`,
				Diagram: `high addr ┌────────────────────┐
          │  caller's frame    │
          ├────────────────────┤
          │  return address    │ ◀─ pushed by call
          │  saved rbp         │ ◀─ rbp (frame base)
          │  locals / spills   │
low addr  └────────────────────┘ ◀─ rsp   (stack grows down ↓)`,
				MentalModel: "A function call is a push, a jump, and an agreement about who saves which register.",
				TryIt:       "Write a function in x86-64 assembly that sums an int array, and call it from C.",
			},
			{
				Name:    "Integers & Floating Point in Bits",
				Summary: "Numbers are fixed-width bit patterns, and that has consequences.",
				Body: `Signed integers use two's complement: -x is ~x + 1, so the same adder
handles signed and unsigned arithmetic. Unsigned overflow wraps. In C,
signed overflow is undefined behaviour, and optimizers exploit that
assumption.

IEEE 754 floats store a sign, a biased exponent and a mantissa. Most
decimals cannot be represented exactly (0.1 + 0.2 != 0.3), and float
addition is not associative. That last fact returns in Stage 7: a
parallel GPU reduction sums in a different order and can give slightly
different results from a sequential loop.`,
				MentalModel: "Every number type is a finite code; learn where the code breaks.",
				TryIt:       "Print the bit patterns of -1, INT_MIN, 0.1f and 1e30f + 1.0f - 1e30f, and explain each one.",
			},
			{
				Name:    "The Memory Hierarchy & Locality",
				Summary: "Where your data lives often matters more than how many instructions you run.",
				Body: `Rough latencies on a modern CPU: register ~1 cycle, L1 ~4 cycles,
L2 ~12, L3 ~40, DRAM hundreds of cycles. Caches move data in 64-byte
lines, so touching one byte brings in its 63 neighbours.

Programs are fast when they reuse data soon (temporal locality) and use
neighbouring data (spatial locality). Walking a row-major 2D array
column by column can be ten times slower than walking it row by row, with
identical instruction counts. The same idea returns in Stage 7 as the
roofline model and GPU memory coalescing.`,
				Diagram: `  CPU ─ regs ─ L1 (32K) ─ L2 (1M) ─ L3 (tens of MB) ─ DRAM (GBs) ─ SSD
  fast · tiny ◀──────────────────────────────────────────────▶ slow · huge`,
				MentalModel: "Memory is a pyramid; performance is how rarely you fall down it.",
				TryIt:       "Time row-major vs column-major traversal of a 4096×4096 int matrix.",
			},
			{
				Name:    "Parsing: Recursive Descent & Pratt",
				Summary: "Turning a flat token stream into a tree that respects precedence.",
				Body: `A recursive-descent parser has one function per grammar rule, and each
function consumes tokens and returns an AST node. Operator precedence
is the hard part. Pratt parsing (precedence climbing) solves it neatly:
each token gets a binding power, and a loop keeps extending the left
expression while the next operator binds more tightly than the current
level.

Once you have an AST, everything else (type checking, interpretation,
code generation) is a tree walk.`,
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
		},
	},

	// -----------------------------------------------------------------
	2: {
		Overview: `The operating system is the program that shares the hardware among
programs that do not trust each other. Learn how it virtualises the CPU
and memory, and what really happens on a system call, a page fault and a
context switch.`,
		Concepts: []Concept{
			{
				Name:    "Processes & the Address Space",
				Summary: "Every process believes it owns all of memory.",
				Body: `A process is a running program: an address space, one or more threads,
and kernel state such as open files and credentials. Each process gets a
private virtual address space with a conventional layout. ASLR
randomises the base addresses so attackers cannot predict them.

fork() clones a process, exec() replaces its image with a new program,
and wait() collects a child's exit status. Every shell is built from
those three calls.`,
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
- copy-on-write fork;
- memory-mapped files and swapping;
- segfaults, when the access is truly invalid.`,
				MentalModel: "Memory is a lookup table that the kernel fills in only when you trip over a hole.",
				TryIt:       "malloc 1 GiB, check RSS in /proc/self/status, touch every 4096th byte, and check again.",
			},
			{
				Name:    "System Calls & Privilege",
				Summary: "The one controlled doorway from user code into the kernel.",
				Body: `User code runs in an unprivileged CPU mode. To open a file or send a
packet it executes the syscall instruction, which switches to kernel mode
at a fixed entry point with arguments in registers. The kernel validates
everything (pointers from user space are hostile until proven otherwise),
does the work and returns.

A syscall costs on the order of hundreds of nanoseconds, more with
speculative-execution mitigations. That cost is why buffered I/O exists
and why batching interfaces such as io_uring matter.`,
				MentalModel: "A syscall is a function call across a trust boundary, so it is checked and not free.",
				TryIt:       "Run strace -c on ls and on a simple Go program, and compare the syscall counts.",
			},
			{
				Name:    "Scheduling & Context Switches",
				Summary: "How one CPU gives many threads the illusion of running at once.",
				Body: `A timer interrupt periodically returns control to the kernel, whose
scheduler picks the next runnable thread. Linux used CFS for years and
has used EEVDF since 6.6. A context switch saves one thread's registers
and restores another's. Switching between processes may also switch page
tables and lose TLB entries.

Threads in one process share an address space, so switching between them
is cheaper, but they also share every bug. Green threads (goroutines,
async runtimes) do the switching in user space and avoid the kernel
entirely.`,
				MentalModel: "Concurrency is an illusion made of interrupts and saved registers.",
				TryIt:       "Write a coroutine library with a hand-written assembly context switch and round-robin scheduling.",
			},
			{
				Name:    "Synchronisation & Memory Ordering",
				Summary: "Shared memory without rules is a data race.",
				Body: `Two threads doing count++ lose updates because it is a load, an add and
a store. Atomic read-modify-write instructions (compare-and-swap,
fetch-add) are the hardware primitive underneath. A mutex is an atomic
fast path plus a futex so that waiting threads sleep in the kernel.

CPUs and compilers reorder memory operations. Acquire and release
semantics (and fences) define which reorderings other threads can
observe. A deadlock needs all four Coffman conditions: mutual exclusion,
hold-and-wait, no preemption and circular wait. Break any one of them.`,
				MentalModel: "Without a happens-before edge, another thread may see anything.",
				TryIt:       "Build a spinlock with atomics, then a futex-based mutex, and benchmark both under contention.",
			},
		},
		Resources: []Resource{
			{"Book", "Operating Systems: Three Easy Pieces (free)", "https://pages.cs.wisc.edu/~remzi/OSTEP/", "The clearest OS textbook; do its homework simulators."},
			{"Course", "MIT 6.1810 Operating System Engineering (xv6)", "https://pdos.csail.mit.edu/6.1810/", "Labs hack on a small, real Unix kernel."},
			{"Paper", "What Every Programmer Should Know About Memory", "https://people.freebsd.org/~lstewart/articles/cpumemory.pdf", "Drepper's deep dive into caches, the TLB and NUMA."},
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
		},
	},

	// -----------------------------------------------------------------
	3: {
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
		},
		Resources: []Resource{
			{"Course", "CMU 15-445 Database Systems", "https://15445.courses.cs.cmu.edu/", "Andy Pavlo's lectures and the BusTub project."},
			{"Book", "Designing Data-Intensive Applications", "https://dataintensive.net/", "The map of the whole storage and distributed-data landscape."},
			{"Article", "Let's Build a Simple Database", "https://cstack.github.io/db_tutorial/", "A SQLite clone in C, step by step."},
			{"Book", "Build Your Own Database", "https://build-your-own.org/", "B+tree, KV store and SQL layer from scratch."},
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
	4: {
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
			{"Replicated KV (capstone A)", "Put the Stage 3 storage engine behind the Raft log: a fault-tolerant KV service.",
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
	5: {
		Overview: `Deep learning is linear algebra plus the chain rule, run at scale. This
stage builds fluent intuition for matrices as transformations and for
derivatives of functions of vectors and matrices, so that every
backward pass in Stage 6 is something you can derive on paper.`,
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
			{"Video", "3Blue1Brown: Essence of Linear Algebra", "https://www.3blue1brown.com/topics/linear-algebra", "Geometric intuition first; watch it before any textbook."},
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
				[]string{"Finite-difference engine", "Relative-error metric", "Test matmul, softmax and CE", "Reuse in Stage 6"}},
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
	6: {
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
				TryIt:       "Build a scalar Value type with + × tanh and backward(), and check it against the Stage 5 gradient checker.",
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
O(n²) in sequence length, the bottleneck that FlashAttention (Stage 7)
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
	7: {
		Overview: `This stage is where the other six meet. The memory hierarchy (Stage 1),
virtual memory (Stage 2) and the math and autograd (Stages 5 and 6)
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
Stage 2, applied to attention. This removes fragmentation and lets
requests share prompt prefixes. Quantisation (int8, int4) cuts bytes
everywhere.`,
				MentalModel: "An LLM server is an operating system whose scarce resource is KV memory.",
				TryIt:       "Compute the KV-cache size for a Llama-style 8B model (32 layers, 8 KV heads, head_dim 128, fp16) at 8k context.",
			},
		},
		Resources: []Resource{
			{"Book", "Programming Massively Parallel Processors", "", "The CUDA textbook (Hwu, Kirk & El Hajj)."},
			{"Site", "CUDA C++ Programming Guide", "https://docs.nvidia.com/cuda/cuda-c-programming-guide/", "The official execution and memory model."},
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
			{"Which Stage 2 idea does PagedAttention borrow?", "Paging: a per-sequence block table maps logical KV positions to fixed-size physical blocks."},
		},
	},
}
