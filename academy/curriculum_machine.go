package main

// Stage 0: How Computers Work. The first stage, before any programming: a
// guided tour of the whole machine, from bits and the CPU to the operating
// system, devices, networks and booting, so that everything after it has
// somewhere to hang.

var machineGuides = map[int]StageGuide{
	0: {
		Overview: `Before writing a single line of code, it helps to know what a computer
actually is. This stage is a guided tour of the whole machine. You will
follow a keypress from your finger to the screen, a program from a text
file to a running process, and a web page from a server across the world
to your eyes.

A computer is a machine that follows instructions, very quickly and
completely literally. Everything it stores (text, photos, music,
programs) is numbers, and every number is a pattern of bits: switches
that are on or off. On top of that sit a few big ideas, each one hiding
the details of the layer below: the CPU runs instructions, memory holds
them, the operating system shares the machine between programs, and
networks connect machines together.

You do not need to memorise any of this. The goal is a mental map, so
that when later stages zoom in (C in Stage 1, logic gates in Stage 4,
the operating system in Stage 6, networks in Stage 8) you already know
where each piece fits.`,
		Outcomes: []string{
			"Explain how text, images, sound and programs are all stored as bits and bytes",
			"Describe what the CPU does in each fetch–decode–execute cycle",
			"Explain what happens between typing a program, running it, and seeing its output",
			"Trace a keypress, a web request and a boot sequence through the hardware and the operating system",
		},
		Glossary: []Term{
			{"Bit", "The smallest piece of information: a single 0 or 1."},
			{"Byte", "A group of 8 bits, able to hold 256 different values (0 to 255)."},
			{"Binary", "Counting with only two digits, 0 and 1, the number system computers use."},
			{"CPU", "The processor: the chip that fetches, decodes and executes instructions."},
			{"Register", "A tiny, very fast storage slot inside the CPU, holding what it is working on right now."},
			{"Clock", "The CPU's metronome; each tick lets it take the next step. 3 GHz means 3 billion ticks per second."},
			{"RAM", "The computer's working memory: fast, large, and wiped when the power goes off."},
			{"Storage", "Where data survives with the power off: an SSD or hard disk, organised into files."},
			{"Operating system", "The software (Linux, Windows, macOS, Android) that manages the hardware and runs programs."},
			{"Kernel", "The core of the operating system, running with full control of the hardware."},
			{"Process", "A running program, with its own memory, managed by the operating system."},
			{"Interrupt", "A signal from a device that makes the CPU pause what it is doing and handle an event."},
		},
		Concepts: []Concept{
			{
				Name:    "Bits & Bytes: Everything Is a Number",
				Summary: "Everything a computer stores (text, colours, sound, programs) is numbers, written in bits.",
				Body: `A computer is built from billions of tiny electronic switches called
transistors. Each one is either off or on, and we write that as 0 or 1:
a bit. Why only two states? Because electronics are noisy. Telling "low
voltage" from "high voltage" is easy and reliable even when the signal
wobbles; telling ten different levels apart would not be.

With more bits you can count higher, exactly as with more decimal
digits. Binary is base 2: each position is worth twice the one to its
right (1, 2, 4, 8, 16…), so 1011 is 8 + 0 + 2 + 1 = 11. Eight bits make
a byte, which can hold 256 patterns (0 to 255). Programmers often write
bytes in hexadecimal (base 16, digits 0–9 and A–F) because two hex
digits are exactly one byte: 0xFF is 255.

Bits only mean something when you agree how to read them:
- Text: each character gets a number. In ASCII, "A" is 65; Unicode extends this to every script and emoji, and UTF-8 stores those numbers as 1 to 4 bytes.
- Colours: a pixel is usually three bytes, for red, green and blue. #FF8800 is orange.
- Images: a grid of pixels. A 12-megapixel photo is 36 million bytes before compression.
- Sound: the air pressure measured 44,100 times per second, each measurement a number (CD quality).
- Programs: the CPU's instructions are numbers too, stored in memory next to the data.

The same bytes can mean completely different things. The two bytes 0x48
0x69 are the text "Hi", or the number 18,537 or 26,952 (depending on
byte order), or part of a colour. The type decides, which is why types
matter so much when you program.`,
				Diagram: `  one byte = 8 bits        0  1  0  0  0  0  0  1
  place values             128 64 32 16  8  4  2  1
  value                    64 + 1 = 65
  read as a number  → 65
  read as ASCII     → 'A'
  read as red level → a dark red (65 of 255)`,
				UnderTheHood: `Inside the chip, a bit is a voltage on a wire: roughly 0 volts for 0,
and around 1 volt for 1 in a modern processor. A transistor is a switch
that one voltage can turn on or off, so transistors can control other
transistors, and that is enough to build logic gates (Stage 4), then
adders, then memory cells, then a whole CPU.

Negative whole numbers are usually stored in two's complement: the
highest bit counts as negative, so in one byte 11111111 is -1 and
10000000 is -128. Numbers with fractions use floating point, a binary
version of scientific notation, which is why 0.1 + 0.2 is not exactly
0.3 in most languages.`,
				MentalModel: "Bits have no meaning on their own; the reader decides whether they are a number, a letter, a colour or an instruction.",
				TryIt:       `On Linux or macOS, run: echo -n "Hi!" | xxd   (or hexdump -C). You will see 48 69 21: the bytes for H, i and !. Then look up those numbers in an ASCII table.`,
				Analogy: `Morse code: only two symbols, dot and dash, yet you can send any
message, as long as both sides agree on the code. Or a light switch
panel: eight switches can be set in 256 different ways, and an
agreed chart says what each pattern means.`,
				Example: `Every photo on your phone is millions of numbers: one byte each for the
red, green and blue of every pixel. When you "brighten" it, the app
adds to those numbers. A song on your headphones is a list of about
44,100 numbers per second per ear. The text you are reading now is a
list of numbers, one or more per character.`,
				Exercises: trio(
					"Convert these by hand: binary 1010 and 11111111 to decimal; decimal 13 and 200 to binary; hex 0x2A and 0x10 to decimal.",
					"Write the place values 128 64 32 16 8 4 2 1 above the bits. Hex digits are worth 16 and 1.",
					"Using an ASCII table, decode the bytes 72 101 108 108 111, then encode your first name as decimal and hex bytes. How many bytes does an emoji take in UTF-8?",
					"72 is H. Most emoji take 4 bytes in UTF-8; you can check with echo -n '😀' | xxd.",
					"Your phone camera takes 12-megapixel photos. How many bytes would one photo take uncompressed (3 bytes per pixel)? Compare with a real JPEG from your phone and explain the difference.",
					"12,000,000 × 3 = 36 million bytes (about 36 MB). JPEG compression throws away detail the eye barely notices, usually leaving 2–5 MB.",
				),
			},
			{
				Name:    "The CPU at Work: Fetch, Decode, Execute",
				Summary: "The processor repeats one tiny loop, billions of times a second: fetch an instruction, decode it, execute it.",
				Body: `The CPU (central processing unit) is the part that actually does things.
Inside it are:
- registers: a few dozen tiny, extremely fast storage slots holding the values being worked on;
- the ALU (arithmetic logic unit): the circuit that adds, subtracts, compares and combines bits;
- the control unit: it reads instructions and tells everything else what to do;
- the program counter: a special register holding the address of the next instruction.

A program is a list of instructions in memory, each a number that means
something like "load this value", "add these two registers", "store
this result" or "jump to that address". The CPU runs them with a loop
that never stops while the computer is on:
- Fetch: read the instruction at the address in the program counter.
- Decode: work out which operation it is and what it applies to.
- Execute: do it (add, compare, load, store…), then move the program counter to the next instruction.

Decisions and loops are just jumps. A comparison sets a flag, and a
conditional jump either continues to the next instruction or changes the
program counter to another address. Everything a program does, from a
spreadsheet formula to a video game, is built from these simple steps.

A clock keeps everything in step. A 3 GHz processor ticks 3 billion
times per second, and modern CPUs overlap and parallelise instructions
(pipelining, several cores), so they complete billions of instructions
per second. The set of instructions a CPU understands is its instruction
set architecture: x86-64 in most PCs, ARM64 in phones and newer Macs,
RISC-V in a growing number of chips. That is why a program compiled for
one does not run on the other without translation.`,
				Diagram: `        ┌──────────────── CPU ────────────────┐
        │  program counter: 104               │
        │  registers: r1=7 r2=5               │
        │  ALU: r1 + r2 → 12                  │
        └──────┬──────────────────▲───────────┘
          fetch│ address 104      │instruction
               ▼                  │
        ┌──────────── memory ─────┴───────────┐
        │ 100: load r1   104: add r1, r2      │
        │ 108: store r1  112: jump if ≠ 0 100 │
        └─────────────────────────────────────┘`,
				UnderTheHood: `This is the instruction that "x = x + y" might become on an x86-64 PC,
shown as the bytes in memory and as assembly language (a readable name
for each instruction):

  48 01 d8        add rax, rbx     ; rax = rax + rbx

Three bytes: 48 says "64-bit operands", 01 means "add register to
register", and d8 encodes which two registers. The CPU's decoder turns
those bytes into control signals that route the two registers through
the adder and back. Real CPUs split each instruction into several stages
and work on many instructions at once, guessing which way jumps will go
(branch prediction) so the pipeline stays full. Stage 4 builds this.`,
				MentalModel: "A CPU is a very fast, very obedient clerk: read the next instruction, do exactly that, move on, forever.",
				TryIt:       "Type in this stage's Classic corner program (a tiny CPU simulated in C), predict what it prints, and then trace its first seven steps on paper.",
				Analogy: `A cook following a recipe card by card. The recipe (program) sits in a
binder (memory). The cook reads the next card (fetch), understands it
("whisk two eggs", decode), does it (execute), and moves to the next
card, unless the card says "if the sauce is lumpy, go back to card 4"
(a conditional jump). The cook's hands and the counter are the
registers: only a few things fit, but they are right there.`,
				Example: `When you press play on a video, the CPU runs millions of instructions to
read the file and prepare the stream, then hands the decoding of each
frame to specialised circuits. A 3 GHz laptop core has about a third
of a nanosecond per clock tick: in that time light travels only about
10 centimetres.`,
				Exercises: trio(
					"Put the steps in order and name the component that does each: add two numbers; read the instruction at address 200; realise it is an ADD; store the result in a register; move to address 204.",
					"Fetch (control unit, program counter), decode (control unit), execute (ALU), write back (register), advance the program counter.",
					"Using this stage's tiny CPU type-in, write a new program for it that computes 7 × 3 by repeated addition, and test it.",
					"Keep a counter at 3, add 7 to a sum and subtract 1 from the counter each time around the loop, and use JNZ to repeat.",
					"Find out which instruction set your own devices use (your laptop, your phone, a Raspberry Pi if you have one), and explain why an app built for one cannot simply be copied to another.",
					"On Linux or macOS, uname -m prints x86_64 or arm64 (aarch64). The machine code for one instruction set means nothing to the other.",
				),
			},
			{
				Name:    "Memory & Storage: Where Data Lives",
				Summary: "Fast, small memory close to the CPU; big, slow storage further away; and caches in between.",
				Body: `Memory (RAM) is one enormous row of numbered bytes. Each byte has an
address, a plain number from 0 up to billions. When a program uses a
variable, the CPU is reading and writing bytes at some address. RAM is
fast and directly addressable, but volatile: it forgets everything when
the power goes off.

Storage (an SSD or hard disk) keeps data with the power off, in files
organised by a filesystem. It is far bigger and cheaper per byte, and
far slower. Programs must be copied from storage into memory before the
CPU can run them, which is what "loading" means.

The gap in speed is enormous, so computers use a hierarchy, each level
smaller and faster than the one below. Approximate times to fetch one
piece of data:
- a register: well under a nanosecond (a billionth of a second);
- L1 cache: about 1 ns; L2: a few ns; L3: about 10–20 ns;
- main memory (RAM): about 100 ns;
- an SSD: roughly 10–100 microseconds;
- a spinning hard disk: roughly 5–10 milliseconds;
- a round trip across an ocean over the internet: about 100 milliseconds.

Caches work because programs have locality: they tend to reuse data they
used recently and data stored next to it. The CPU fetches memory in
blocks (cache lines, usually 64 bytes), so walking through an array in
order is fast, and jumping around randomly is slow, even when the
amount of work looks the same.`,
				Diagram: `  faster, smaller, pricier per byte
   ▲   registers        < 1 ns        bytes
   │   L1 / L2 / L3     1–20 ns       KB → MB
   │   RAM              ~100 ns       GB
   │   SSD              ~10–100 µs    hundreds of GB → TB
   ▼   hard disk        ~5–10 ms      TB
  slower, bigger, cheaper per byte`,
				UnderTheHood: `Scale the times up so that one nanosecond becomes one second. Then an L1
cache hit takes a second, reading RAM takes about a minute and a half,
an SSD read takes about a day, a hard-disk seek takes a few months, and
a round trip across the Atlantic takes about three years. This is why
performance work is mostly about moving data less, and why the same
idea returns in Stage 5 (the memory hierarchy), Stage 7 (databases) and
Stage 16 (AI inference is limited by memory bandwidth).

Each running program also sees virtual memory: the addresses it uses
are private and are translated by the CPU and the operating system into
real RAM addresses, page by page (usually 4 KB pages). That is how two
programs can both use address 0x1000 without colliding (Stage 6).`,
				MentalModel: "Registers are your hands, cache is the desk, RAM is the bookshelf, the disk is the warehouse across town.",
				TryIt:       "Check your own machine: on Linux run free -h and lscpu | grep -i cache; on macOS, sysctl hw.memsize and sysctl -a | grep cachesize. How big is each level, and how much faster is the smallest one?",
				Analogy: `Cooking again. What you are using right now is in your hands
(registers). Ingredients for this dish are on the counter (cache). The
fridge holds this week's food (RAM): a few steps away. The supermarket
holds everything (storage): a trip. A good cook, like a good program,
brings what the next steps need to the counter before starting.`,
				Example: `When your laptop "freezes" after you open too many browser tabs, RAM is
full, and the operating system is moving memory out to the much slower
disk and back (swapping). When a game shows a loading screen, it is
copying gigabytes of textures from storage into RAM and GPU memory,
because it cannot draw from the disk fast enough.`,
				Exercises: trio(
					"Put in order from fastest to slowest and give a rough time for each: RAM, L1 cache, SSD, register, a hard disk, a request to a server on another continent.",
					"Register < L1 < RAM (~100 ns) < SSD (~tens of µs) < hard disk (~ms) < intercontinental round trip (~100 ms).",
					"Using the one-nanosecond-becomes-one-second scale, work out how long 1,000 RAM reads would take on that scale, and compare with 1,000 L1 cache hits.",
					"1,000 × 100 s ≈ 28 hours versus 1,000 × 1 s ≈ 17 minutes: the same work, about 100 times slower.",
					"Watch your computer's memory while you work (Activity Monitor, Task Manager or htop): open 30 browser tabs and a video, and record what happens to used memory, cached memory and swap. Explain what you see.",
					"'Cached' memory is the operating system keeping recently used file data in RAM; it is released when programs need it. Swap growing means RAM is full.",
				),
			},
			{
				Name:    "From Source Code to Running Program",
				Summary: "Text you write becomes machine code in a file; the operating system loads that file into memory and runs it as a process.",
				Body: `A program begins as source code: text in a language people can read,
such as C. The CPU cannot run text. A compiler translates the source
into machine code for a specific instruction set and writes an
executable file. That file holds the machine code, the program's fixed
data (such as the text it prints), and a header describing how to load
it. On Linux the format is called ELF, on Windows PE, and on macOS
Mach-O.

When you run a program, a chain of events follows:
- The shell (or the icon you double-click) asks the operating system to start the executable.
- The operating system creates a new process: a private address space, and an entry in its table of running programs.
- The loader maps the file's code and data into that address space, loads any shared libraries it needs (such as the C library), and sets up a stack.
- The CPU jumps to the program's entry point. For C, a little startup code runs first and then calls main.
- The program runs, asking the operating system for anything outside its own memory (files, the screen, the network) through system calls.
- When main returns, the process exits with a status code (0 means success), and the operating system frees its memory.

Interpreted languages add one step: the executable that runs is the
interpreter (python3, node), and it reads your script and carries it
out. Python first compiles the script to bytecode, instructions for
its own virtual machine, and then runs those.`,
				Diagram: `  hello.c ──compiler──▶ ./hello  (machine code on disk)
                           │ you run it
                           ▼
  kernel: new process ▶ load code + data ▶ set up stack
                           │ jump to main
                           ▼
  running process ──system calls──▶ kernel (files, screen, network)
                           │ main returns
                           ▼
  exit(0) ▶ memory freed`,
				UnderTheHood: `You can watch every step on Linux:

  file ./hello              # ELF 64-bit LSB executable, x86-64 ...
  ldd ./hello               # the shared libraries it needs (libc)
  strace ./hello            # every system call it makes

strace shows execve (the kernel starting the program), the loader
mapping libc into memory with mmap, then a single write(1, "Hello,
world!\n", 14) for your printf, and exit_group(0) at the end. Your whole
program, seen from the operating system, is a handful of requests.`,
				MentalModel: "An executable is a frozen program on disk; a process is that program thawed into memory and running.",
				TryIt:       "Compile hello.c (Stage 1's first concept), then run file ./hello, ls -l ./hello and, on Linux, strace ./hello. Find the write call that prints your text.",
				Analogy: `A recipe book on a shelf (the executable) versus a meal being cooked
(the process). Opening the book and cooking is the loader's job:
clear a space on the counter (memory), fetch the equipment the recipe
needs (libraries), then start at step one (main). Two people can cook
the same recipe at once in two kitchens: two processes from one file.`,
				Example: `When you double-click an app, your operating system does exactly this
chain. The apps on your phone's home screen are executable files in
storage. The spinning icon while one opens is the loader copying and
linking its code, and the app's data being read in.`,
				Exercises: trio(
					"Order these events: main starts; the process exits with status 0; you type ./hello; the operating system creates a process; the loader maps the code into memory; printf asks the kernel to write to the screen.",
					"Type → create process → load → main → system call for output → exit.",
					"Compile hello.c twice, once normally and once with gcc -static. Compare the file sizes with ls -l and the ldd output, and explain the difference.",
					"A static executable copies the C library into the file itself; a dynamic one asks the loader to map the shared library at start-up.",
					"On Linux, run strace -c ls (or dtruss on macOS, with care) and list the five most frequent system calls ls makes. For each, say in one sentence what it asks the kernel to do.",
					"Typical ones: openat (open a file or directory), getdents64 (read directory entries), fstat (file details), mmap (map memory), write (print).",
				),
			},
			{
				Name:    "The Operating System: The Machine's Manager",
				Summary: "The operating system shares one machine between many programs, safely and fairly.",
				Body: `Your computer runs hundreds of programs at once on a handful of CPU
cores, with one set of memory chips and one disk. The operating system
(Linux, Windows, macOS, Android, iOS) makes that possible. Its core, the
kernel, is the only software with full control of the hardware.

The CPU helps enforce this with privilege levels. Ordinary programs run
in user mode, where they cannot touch devices or other programs' memory
directly. The kernel runs in kernel mode. When a program needs
something (open a file, send a packet, get more memory, start another
program), it makes a system call: a special instruction that switches
the CPU into kernel mode and runs kernel code on the program's behalf.

The kernel's main jobs:
- Processes and scheduling: it gives each running program a turn on a CPU core, a few milliseconds at a time, switching between them so quickly that they all seem to run at once.
- Memory: it gives each process its own virtual address space and stops programs from reading each other's memory. A program that tries gets stopped (on Unix, a segmentation fault).
- Files: it turns blocks on a disk into named files and folders, with permissions saying who may read and write them.
- Devices: through drivers, it talks to keyboards, screens, disks, network cards and cameras, so programs can use them without knowing the hardware details.

Everything else (the desktop, the shell, the browser) is an ordinary
program running on top.`,
				Diagram: `  user mode     [ browser ] [ editor ] [ shell ] [ your program ]
                    │ system calls: open, read, write, fork, mmap …
  ─ ─ ─ ─ ─ ─ ─ ─ ─ ▼ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
  kernel mode   [ scheduler ] [ memory ] [ filesystems ] [ drivers ]
                    │
  hardware      [ CPU cores ] [ RAM ] [ disk ] [ network ] [ screen ]`,
				UnderTheHood: `A context switch is the kernel saving everything about one program (its
registers, including the program counter, and which memory map it
uses) and loading another program's saved state, so the CPU simply
carries on where the other program left off. It takes on the order of a
few microseconds, and happens thousands of times per second.

A hardware timer fires regularly and interrupts whatever is running,
which is how the kernel gets control back even from a program stuck in
an infinite loop. Stage 6 looks inside all of this: processes, virtual
memory, system calls, scheduling and synchronisation.`,
				MentalModel: "Programs ask; the kernel decides. Nothing touches the hardware without going through it.",
				TryIt:       "Open your system monitor (top or htop on Linux and macOS, Task Manager on Windows). Count the processes, find which uses most CPU and most memory, and note how many CPU cores you have.",
				Analogy: `A hotel. Guests (programs) have private rooms (memory) they cannot enter
each other's, a front desk (system calls) for anything they need, and
staff (the kernel) who alone have master keys, run the lifts
(scheduling), manage the storerooms (files), and deal with the plumbers
and electricians (drivers).`,
				Example: `When an app on your phone crashes, the other apps keep running: the
operating system isolated it. When you plug in a USB stick and it
appears as a drive, a driver recognised the device and a filesystem
driver made its contents look like files and folders.`,
				Exercises: trio(
					"For each action, say whether the program can do it alone or needs a system call: add two numbers, read a file, print to the screen, call its own function, open a network connection.",
					"Anything that touches something outside the process's own memory and CPU registers (files, the screen, the network) needs the kernel.",
					"Write (or reuse) a C program that loops forever. Run it, then watch it in top, change its priority with renice, and stop it with kill. Explain what each tool asked the kernel to do.",
					"kill sends a signal; the default action for SIGTERM ends the process. renice changes how the scheduler weighs it.",
					"Investigate why your computer feels slow at a busy moment: using the system monitor, decide whether the bottleneck is CPU, memory, disk or network, and back it up with numbers.",
					"CPU near 100% on all cores, memory full with swap growing, disk busy near 100%, or network saturated: each points to a different fix.",
				),
			},
			{
				Name:    "Input & Output: Keys, Screens & Interrupts",
				Summary: "Devices signal the CPU with interrupts; drivers translate; the operating system delivers events to programs.",
				Body: `A computer that could only compute would be useless; it has to talk to
the world through devices: keyboards, mice, touchscreens, screens,
speakers, disks, network cards. Each device has a controller, a small
chip that speaks its own language, and the operating system has a
driver for it.

Devices tell the CPU that something happened with interrupts. An
interrupt is an electrical signal that makes the CPU finish its current
instruction, save where it was, and jump to the kernel's handler for
that device. The handler does the minimum (for example, reads which key
arrived), and the CPU returns to what it was doing. The alternative,
polling, means asking the device over and over whether anything
happened, which wastes time when nothing does.

For large transfers (a disk read, a network packet, a frame of video),
devices use DMA (direct memory access): they copy data straight into
memory, and interrupt the CPU only when the whole transfer is done.

Follow one keypress, pressing "a" in a text editor:
- The keyboard's controller scans its grid of switches, notices the key, and reports it. USB keyboards are asked for reports by the computer, typically every 1 to 8 milliseconds.
- The USB host controller receives the report and raises an interrupt.
- The kernel's handler and the keyboard driver turn the key code into an event: "key a pressed".
- The event is delivered to the program that has focus (through the window system on a desktop, or the terminal).
- The editor adds "a" to its text, and asks for the window to be redrawn.
- The graphics system and GPU draw the new frame, and at the display's next refresh (every 16.7 ms at 60 Hz) the letter appears.`,
				Diagram: `  key ─▶ keyboard controller ─▶ USB host controller
                                     │ interrupt (IRQ)
                                     ▼
  CPU ▶ kernel interrupt handler ▶ keyboard driver
                                     │ "key a pressed"
                                     ▼
  window system / terminal ▶ editor adds 'a' ▶ redraw
                                     │
                                     ▼
  GPU draws the frame ▶ shown at the next screen refresh`,
				UnderTheHood: `Programs never see interrupts directly. A C program reading the keyboard
calls read(0, buf, n) on standard input; if no key has arrived, the
kernel puts the process to sleep, runs other programs, and wakes it
when the keyboard interrupt delivers data. That is why a waiting
program uses no CPU at all.

The terminal adds a layer of its own: by default it collects a whole
line and lets you edit it with Backspace before your program sees it.
This academy switches that off (raw mode) so that keys like Ctrl+L reach
it immediately, which is exactly the kind of low-level detail Stage 6
explains.`,
				MentalModel: "Devices interrupt, drivers translate, the kernel delivers, programs react.",
				TryIt:       "On Linux, run cat /proc/interrupts twice, a few seconds apart, while typing and moving the mouse. Which lines' counts grow? (On other systems, just count how many devices your system settings list.)",
				Analogy: `A restaurant kitchen. Checking the dining room every few seconds for new
customers (polling) wastes the chef's time. A bell at the door
(an interrupt) lets the chef keep cooking until someone actually
arrives. A waiter who carries whole trays from the kitchen to the
table without the chef's help is DMA.`,
				Example: `Gaming mice and keyboards advertise "1000 Hz polling": the computer asks
them for updates every millisecond instead of every 8, shaving a few
milliseconds off the delay. The whole journey from pressing a key to
seeing it on screen typically takes tens of milliseconds, and most of
that is waiting for the next screen refresh.`,
				Exercises: trio(
					"Explain the difference between polling and interrupts, and give one situation where polling is actually the better choice.",
					"Polling wastes work when events are rare, but is simple and has predictable timing when events are very frequent (very fast network cards sometimes poll).",
					"Estimate the keypress-to-screen delay on a 60 Hz display: keyboard polling (up to 8 ms), processing (about 1 ms) and waiting for the next refresh (up to 16.7 ms). What are the best and worst cases, and what changes at 144 Hz?",
					"Add up the minimum and the maximum of each part. At 144 Hz a refresh comes every 6.9 ms.",
					"Write your own 'what happens when I press a key' for a phone touchscreen, from finger to pixel, naming each hardware and software layer. Compare it with the keyboard path above.",
					"The touch controller measures changes in capacitance on a grid and reports coordinates; the rest (interrupt, driver, event, app, redraw) follows the same pattern.",
				),
			},
			{
				Name:    "Networks: How Computers Talk",
				Summary: "Data travels in small packets, passed from router to router, and protocols in layers make it reliable and meaningful.",
				Body: `Computers talk by sending packets: small chunks of data, usually up to
about 1,500 bytes, each with an address label. Protocols are the
agreed rules, and they come in layers, each relying on the one below:
- Link: getting bits across one hop, by Wi-Fi radio, an Ethernet cable or fibre.
- Internet (IP): every device has an IP address, and routers pass each packet one hop closer to its destination, like sorting offices forwarding a letter.
- Transport: TCP turns unreliable packets into a reliable, ordered stream, numbering the bytes and resending anything lost. UDP skips that, for speed. Port numbers pick which program on the machine receives the data (443 for secure web servers).
- Application: the conversation itself. HTTP for the web, with TLS encrypting it (HTTPS); SMTP for email; DNS for names.

Follow one web page, typing example.com and pressing Enter:
- DNS: your computer asks a DNS resolver for the IP address behind the name example.com.
- TCP: it opens a connection to that address on port 443, with a three-way handshake (SYN, SYN-ACK, ACK).
- TLS: the two sides agree on encryption keys, and the server proves its identity with a certificate.
- HTTP: the browser sends GET / and the server replies with the page's HTML.
- The browser reads the HTML, fetches the images, styles and scripts it mentions (more requests), and draws the page.

Each step takes at least one round trip, and light in fibre needs about
5 milliseconds per 1,000 km, so the distance to the server matters as
much as your bandwidth.`,
				Diagram: `  your laptop                          server
  HTTP  GET /                   ⇄      HTTP  200 OK
  TLS   encrypt                 ⇄      TLS   decrypt
  TCP   port 443, seq numbers   ⇄      TCP
  IP    to 93.184.x.x ▶ router ▶ router ▶ IP
  Link  Wi-Fi / Ethernet               Ethernet`,
				UnderTheHood: `You can watch each layer with standard tools:

  nslookup example.com        # DNS: name → IP address
  ping -c 3 example.com       # round-trip time, in milliseconds
  traceroute example.com      # every router on the way (tracert on Windows)
  curl -v https://example.com # the TCP connection, TLS handshake and HTTP

In a program, all of this sits behind a socket: the operating system
object you connect, write to and read from (Stage 8). The kernel's
network stack does the TCP bookkeeping, and the network card moves the
packets into memory with DMA.`,
				MentalModel: "Packets are letters, IP addresses are postal addresses, routers are sorting offices, and TCP is the registered-mail service that makes sure nothing is lost.",
				TryIt:       "Run traceroute (tracert on Windows) to a website in another country, count the hops, and use ping to measure the round-trip time. Estimate how far away the server is from the time alone.",
				Analogy: `Sending a long letter by postcards: you cut it into numbered postcards
(packets), each with the address (IP). Post offices (routers) pass each
one along, perhaps by different routes. The recipient puts them back in
order and asks you to resend any that went missing (TCP). DNS is the
address book that turns a name into a postal address.`,
				Example: `Loading a typical news website means dozens of DNS lookups, TCP and TLS
connections, and often more than a hundred HTTP requests for text,
images, fonts, scripts and adverts, which is why pages load faster when
content is served from a nearby data centre (a CDN).`,
				Exercises: trio(
					"Match each to its layer and job: IP address, port 443, Wi-Fi, TCP retransmission, DNS, HTTP GET.",
					"Link: Wi-Fi. Internet: IP address. Transport: port, retransmission. Application: DNS, HTTP.",
					"Use curl -v https://example.com (or your browser's developer tools, Network tab) to identify the DNS result, the TLS version, the HTTP status code and the response size. Write each down with where you found it.",
					"In curl -v, lines starting with * are connection details (IP, TLS), > is your request and < is the server's response.",
					"Load a big website with your browser's developer tools open (Network tab). Count the requests, find the largest file and the slowest request, and suggest two ways the site could load faster.",
					"Look at the total transferred size, how many requests go to other domains, and the waterfall: long gaps before the first byte usually mean server or network distance.",
				),
			},
			{
				Name:    "Booting: From Power Button to Login",
				Summary: "Firmware wakes the hardware, a boot loader loads the kernel, and the kernel starts everything else.",
				Body: `When you press the power button, memory is empty and nothing is loaded.
Starting up is a relay race, each runner starting the next:
- Power on: the power supply stabilises, and the CPU starts executing at a fixed address that points into firmware, a program stored in a flash chip on the motherboard (UEFI on modern PCs, the older BIOS before it).
- Firmware: it runs self-tests, initialises memory and devices, and finds a bootable disk. On UEFI systems it reads a boot loader file from a small disk partition (the EFI system partition).
- Boot loader: GRUB on many Linux systems, Windows Boot Manager, or iBoot on Apple devices. It loads the operating-system kernel into memory and jumps to it.
- Kernel: it sets up memory management and interrupts, detects hardware and loads drivers, mounts the main (root) filesystem, and starts the first user program, which gets process ID 1 (systemd on most Linux systems, launchd on macOS).
- Services and login: process 1 starts everything else (networking, the display, background services), then the login screen or shell.

Phones and modern PCs add secure boot: each stage checks the
cryptographic signature of the next before running it, so malware
cannot insert itself early in the chain.`,
				Diagram: `  power ▶ CPU reset ▶ firmware (UEFI): self-test, find boot disk
                          │
                          ▼
  boot loader (GRUB) ▶ loads the kernel into memory
                          │
                          ▼
  kernel: drivers, memory, mounts / ▶ process 1 (systemd)
                          │
                          ▼
  services ▶ login screen ▶ your desktop`,
				UnderTheHood: `On Linux you can read the story of your last boot:

  systemd-analyze            # time spent in firmware, loader, kernel, userspace
  systemd-analyze blame      # which services took longest
  journalctl -b | head -50   # the kernel's first messages from this boot
  ps -p 1                    # process 1

The kernel's own first messages (dmesg) list the CPU, memory and every
device it found, in order.`,
				MentalModel: "Booting is a relay: firmware hands to the boot loader, the boot loader to the kernel, the kernel to process 1, and process 1 to everything else.",
				TryIt:       "On Linux, run systemd-analyze and systemd-analyze blame. How long did firmware, the boot loader, the kernel and user space each take, and which service was slowest?",
				Analogy: `Opening a shop in the morning. The alarm and lights come on
automatically (firmware), the manager arrives with the keys and the
plan for the day (boot loader), the manager unlocks and turns on the
tills, fridges and lights (the kernel initialising hardware), then calls
in the staff (process 1 starting services), and finally the doors open
to customers (the login screen).`,
				Example: `A phone that "boots in 20 seconds" is running this whole relay, with
every stage verified by secure boot. A smart TV, a car's infotainment
system and a Wi-Fi router all boot the same way, usually with a Linux
kernel inside.`,
				Exercises: trio(
					"Put in order and say what each does: boot loader, login screen, firmware, process 1, kernel, power button.",
					"Power → firmware → boot loader → kernel → process 1 → services → login.",
					"On Linux, find your process 1 and the kernel version (ps -p 1, uname -r), and read the first twenty lines of dmesg (sudo dmesg | head -20). Name three pieces of hardware the kernel reported.",
					"dmesg may need sudo. Look for lines naming the CPU, the amount of memory, and storage or USB devices.",
					"Measure and reduce your boot time: record systemd-analyze on a Linux machine (a virtual machine is fine), disable one unneeded service safely, reboot and measure again. On other systems, time from power button to usable desktop with a stopwatch and compare with and without startup apps.",
					"Only disable services you understand (systemctl disable), and write down how to re-enable them. Startup apps are listed in Task Manager (Windows) or Login Items (macOS).",
				),
			},
		},
		Resources: []Resource{
			{"Course", "CS50 Week 0 (Harvard, free)", "https://cs50.harvard.edu/x/weeks/0/", "Binary, representing text, colour, images and sound, and algorithms, before any code."},
			{"Video", "Crash Course Computer Science (PBS, free)", "https://www.youtube.com/watch?v=O5nskjZ_GoI", "A 40-episode tour from early computing to transistors, CPUs, operating systems and the internet; start with episode 1."},
			{"Course", "Nand2Tetris (From NAND to Tetris)", "https://www.nand2tetris.org/", "Build a whole computer yourself, from logic gates to an operating system, in a simulator."},
			{"Book", "Computer Science from the Bottom Up (free)", "https://www.bottomupcs.com/", "Binary, the CPU, the operating system and how programs are built, explained from the ground up. The PDF is in the Library [l]."},
			{"Book", "Code: The Hidden Language of Computer Hardware and Software", "https://www.codehiddenlanguage.com/", "Charles Petzold's gentle, story-like path from Morse code to a working computer."},
			{"Article", "What happens when you type google.com into your browser and press Enter?", "https://github.com/alex/what-happens-when", "A famous, community-written answer that follows one keypress through every layer."},
			{"Article", "Write your Own Virtual Machine (LC-3)", "https://www.jmeiners.com/lc3-vm/", "Build a simulator for a small real computer in about 250 lines of C, step by step."},
		},
		Blueprints: []Blueprint{
			{"Binary & Hex Explorer (in C)", "A command-line tool that shows any number, character or string as decimal, binary and hex, and decodes bytes back into text.",
				[]string{"Number → binary and hex", "Characters and strings → their byte values", "Bytes → text (ASCII, then UTF-8)", "Two's complement view of negative numbers", "Read input safely and reject bad values"}},
			{"Tiny CPU Emulator (in C)", "Extend this stage's type-in into a small computer: more instructions, a real input instruction, and a debugger that shows every step.",
				[]string{"Add MUL, JZ and an INPUT instruction", "Load programs from a text file", "Single-step mode printing registers and memory", "Write three programs for it (maximum, multiplication, countdown)", "Detect and report invalid instructions and addresses"}},
			{"What Happens When… (a lab report)", "Write your own detailed answer to \"what happens when I type a URL and press Enter\", with evidence from real tools at every layer.",
				[]string{"Keypress path, from the keyboard to the browser", "DNS lookup captured with nslookup or dig", "TCP and TLS seen with curl -v", "Routers seen with traceroute", "Browser developer-tools timeline, explained step by step"}},
		},
		Quiz: []Question{
			{"Why do computers use binary rather than decimal?", "Two voltage levels (off and on) are easy to tell apart reliably even with electrical noise; ten levels would not be."},
			{"What happens in one fetch–decode–execute cycle?", "The CPU reads the instruction at the program counter's address, works out what it means, carries it out, and moves the program counter to the next instruction (or jumps)."},
			{"About how much slower is reading main memory than an L1 cache hit, and why do caches work?", "Roughly 100 times (about 100 ns versus about 1 ns). They work because programs reuse recent data and data stored nearby (locality)."},
			{"What is the difference between an executable and a process?", "An executable is a file of machine code on disk; a process is a running instance of it in memory, managed by the operating system. One executable can run as many processes."},
			{"What is a system call?", "A request from a program to the kernel for something only the kernel may do, such as reading a file, drawing on the screen or opening a network connection."},
			{"What is an interrupt, and why is it better than polling for a keyboard?", "A signal that makes the CPU pause and handle a device event. The CPU can do other work (or sleep) until a key arrives, instead of checking constantly."},
			{"Name the main steps when your browser loads https://example.com.", "DNS lookup, TCP connection (handshake), TLS handshake, HTTP request and response, then the browser fetches the page's other files and draws it."},
			{"What runs first when you press the power button?", "Firmware (UEFI or BIOS) stored on the motherboard, which initialises the hardware and starts the boot loader, which loads the kernel."},
		},
	},
}
