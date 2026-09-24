package main

// The beginner layer of the Study Hall. It assumes no prior knowledge: every
// concept gets an everyday analogy and a real-world example, and every
// stage gets a plain-English glossary and a list of concrete outcomes.
// init() merges it into the curriculum, and a test checks that nothing is
// missing.

// plain holds the beginner explanations for one concept.
type plain struct {
	Analogy string
	Example string
}

// stageExtras holds the per-stage beginner material.
type stageExtras struct {
	Outcomes []string
	Glossary []Term
}

func init() {
	for id, g := range curriculum {
		for i := range g.Concepts {
			c := &g.Concepts[i]
			if p, ok := plainWords[c.Name]; ok {
				c.Analogy = p.Analogy
				c.Example = p.Example
			}
			if ex, ok := conceptExercises[c.Name]; ok && len(c.Exercises) == 0 {
				c.Exercises = ex
			}
		}
		if x, ok := extras[id]; ok {
			g.Outcomes = x.Outcomes
			g.Glossary = x.Glossary
		}
		curriculum[id] = g
	}
}

// plainWords is keyed by Concept.Name.
var plainWords = map[string]plain{
	// ---------------- Stage 5: The Iron Layer ----------------
	"The Compilation Pipeline": {
		Analogy: `Imagine translating a novel into Japanese with a team of specialists
working in a line. The first splits the text into words, the next works
out how each sentence is built, another tightens the wording, and the last
writes the final Japanese. Each specialist only needs to understand the
output of the one before them. A compiler is exactly that assembly line,
except the final language is the CPU's own numeric instructions.`,
		Example: `When an iPhone app is built in Xcode, the Clang/LLVM compiler runs these
exact stages. When you open a web page, Chrome's V8 engine compiles the
site's JavaScript into machine code while the page loads, which is why
modern web apps can feel as fast as desktop programs.`,
	},
	"Registers, the Stack & Calling Conventions": {
		Analogy: `Picture a cook in a kitchen. Registers are the few things the cook can
hold in their hands at once. The stack is a pile of plates: every time a
new task starts, a fresh plate goes on top for that task's ingredients,
and it comes off when the task is done. The calling convention is the
house rule every cook follows, for example "hand over ingredients in this
order, and put the finished dish on the left counter". Because everyone
follows the same rule, any cook can hand work to any other.`,
		Example: `When Python's NumPy calls fast code written in C, both sides rely on the
same calling convention. That is also why a C library compiled ten years
ago still works with today's programs. And when a program crashes with
"stack overflow" (usually from runaway recursion), the pile of plates has
literally hit the ceiling.`,
	},
	"Integers & Floating Point in Bits": {
		Analogy: `An old car's odometer has six digits. Drive past 999,999 and it rolls
over to 000,000. That is integer overflow. Floating-point numbers are like
scientific notation with a fixed number of digits: you can write
1.234 × 10³⁰, but you cannot keep every digit of that huge number, so
some precision is always rounded away.`,
		Example: `In 1996 the Ariane 5 rocket self-destructed about 40 seconds after
launch because a 64-bit floating-point value was squeezed into a 16-bit
integer and overflowed. In 2014 "Gangnam Style" passed YouTube's 32-bit
view-counter limit of 2,147,483,647, and YouTube had to move to 64-bit
counters. Banking software stores money as whole cents, never as floats,
precisely because 0.1 + 0.2 is not exactly 0.3 in binary.`,
	},
	"The Memory Hierarchy & Locality": {
		Analogy: `Think of studying for an exam. The notes on your desk (the CPU cache)
take a second to grab. The bookshelf across the room (RAM) takes a
minute. The library across town (the disk) takes an hour. And when you
fetch something, you bring back the whole shelf section, not just one
page. That is why working through material in order is so much faster
than jumping around.`,
		Example: `Game engines such as Unity's DOTS store all game objects of one kind side
by side in memory ("data-oriented design"). They get large speed-ups
without doing any less work, just by being cache-friendly. The same
effect is why a plain array usually beats a linked list in real programs.`,
	},
	"Parsing: Recursive Descent & Pratt": {
		Analogy: `At school you learned that in 3 + 4 × 2 you multiply before you add
(PEMDAS/BODMAS). A parser is a program that knows those rules and turns
the flat line of symbols into a tree: the × sits lower in the tree, so it
is worked out first. It is the same as diagramming sentences in grammar
class, but for code.`,
		Example: `Every time you type a formula such as =SUM(A1:A5)*2 into Excel or Google
Sheets, a parser builds a tree from it before anything is calculated.
Code editors use parsers (for example Tree-sitter) for syntax
highlighting and "go to definition", and your browser parses every
JSON response it receives.`,
	},

	// ---------------- Stage 6: Operating Systems ----------------
	"Processes & the Address Space": {
		Analogy: `Every process is a tenant in an apartment building where all the
apartments have the same floor plan. Apartment "3B" exists in every
building, but they are different rooms. Two programs can both use
"address 1000" and never collide. The kernel is the landlord: it hands out
the apartments and stops tenants from walking into each other's.`,
		Example: `Google Chrome runs each website in its own process ("site isolation").
If one tab crashes, or a malicious page tries to read your banking tab's
memory, the walls between processes stop it.`,
	},
	"Virtual Memory & Paging": {
		Analogy: `A hotel front desk gives you a key card for "room 12", but behind the
scenes the desk maps "your room 12" to whichever real room is free.
Rooms are not even cleaned and prepared until you first walk in (lazy
allocation). If the hotel is full, rarely used guests' luggage goes to
storage in the basement (swap) and comes back when they return.`,
		Example: `This is why your laptop can keep more apps open than it has RAM, and why
switching back to a tab you have not touched for hours sometimes
stutters: its memory pages are being fetched back. Phones do the same
job more aggressively and quietly close background apps when memory runs
low.`,
	},
	"System Calls & Privilege": {
		Analogy: `A bank does not let customers walk into the vault. You go to the teller
window and fill in a slip (the arguments). The teller checks your ID and
does the job for you (the kernel). Every trip to the window takes time,
so a sensible customer does several things per visit, which is batching.`,
		Example: `Saving a file, opening a web page and playing a sound all go through
system calls. Docker containers and browser sandboxes (seccomp) improve
security by limiting which system calls a program is even allowed to
make.`,
	},
	"Scheduling & Context Switches": {
		Analogy: `A chess grandmaster plays thirty opponents at once. They make one move at
a board and walk on to the next. Every opponent feels they have a
dedicated player, but it is really one person switching very quickly.
Before leaving each board, the grandmaster remembers the position; that
is the context switch.`,
		Example: `Your computer runs hundreds of threads on a handful of CPU cores. Music
keeps playing smoothly while a big program compiles because the
scheduler makes sure the audio thread gets its turn on time. Go's
goroutines let a single web server juggle tens of thousands of
connections this way.`,
	},
	"Synchronisation & Memory Ordering": {
		Analogy: `Two people update the same shared bank balance at the same moment. Both
read $100, both add $50, and both write back $150, so $50 has vanished.
A mutex is like the single key to a petrol-station restroom: only the
person holding it can go in. A deadlock is two cars on a one-lane bridge,
each waiting forever for the other to back up.`,
		Example: `Race conditions are not just academic. A race in the software of the
Therac-25 radiation machine (1985–87) caused fatal overdoses, and a
race condition in alarm software contributed to the 2003 blackout in
the north-eastern US and Canada.`,
	},

	// ---------------- Stage 7: Storage Engines ----------------
	"Durability: Page Cache, fsync & Atomic Rename": {
		Analogy: `Telling a waiter your order (write) is not the same as seeing it printed
on the kitchen ticket (fsync). If the waiter trips on the way, your order
is lost. Atomic rename is like painting a new shop sign on a fresh board
and swapping it in one motion: passers-by see either the old sign or the
new one, never a half-painted mess.`,
		Example: `When your phone dies while an app is saving, a well-built app keeps the
old file intact instead of a corrupted half-file, because it saves to a
temporary file and renames it. Text editors such as Vim save this way,
and so does this program: see atomicWriteJSON in main.go.`,
	},
	"B-Trees": {
		Analogy: `Think of a printed phone book or dictionary with tabbed sections. The
tabs say A–F, G–M and so on. Inside each section, headers narrow it down
further. You can find any one of millions of names in three or four
flips, without reading every page.`,
		Example: `When you search your phone's contacts, or filter an online shop by
price, a B-tree index inside the database finds the rows without
scanning everything. Every iPhone and Android phone ships with SQLite,
which is B-tree based, and PostgreSQL and MySQL use B-trees by default.`,
	},
	"LSM-Trees": {
		Analogy: `A busy post room drops every incoming letter into an in-tray, which is
very fast. Every evening someone sorts the trays into the filing
cabinets (compaction). To find a letter, you check today's tray first,
then the cabinets. Writing is quick; the tidying happens later.`,
		Example: `Systems that receive enormous volumes of writes use LSM storage. Discord
stores trillions of chat messages in ScyllaDB, and Meta runs RocksDB
underneath many of its services. Both are LSM-based.`,
	},
	"Write-Ahead Logging & Recovery": {
		Analogy: `An accountant writes every transaction in a journal, in pen, before
moving any money. If the office floods and the spreadsheets are lost, the
balances can be rebuilt by replaying the journal from the beginning. A
ship's captain keeps a logbook for the same reason.`,
		Example: `If a bank's server crashes one second after it says your transfer has
gone through, the transfer is still there after the restart, because
the commit was written to the write-ahead log first. PostgreSQL also
streams its WAL to backup servers to keep them in sync.`,
	},
	"Transactions, Isolation & MVCC": {
		Analogy: `Two fans click "buy" on the last concert seat at the same instant.
Transactions guarantee that exactly one of them gets it. MVCC is like
handing every reader a photocopy of the seating chart as it was when
they started looking, so readers never have to wait for people who are
in the middle of buying.`,
		Example: `Airline booking, hotel reservations and bank transfers all rely on
transactions. A transfer is "take from A" plus "add to B", and it must
happen completely or not at all. Money must never vanish or be created
halfway.`,
	},

	// ---------------- Stage 8: Networks ----------------
	"The Layered Stack & Encapsulation": {
		Analogy: `Sending a letter: you write it (the application), put it in an envelope
addressed to a person (the port, via TCP or UDP), the post office adds a
routing label for the city and street (IP), and trucks carry the sacks
(the link layer). Each worker reads only their own label and never opens
your letter.`,
		Example: `When you load a website on café Wi-Fi, your request is HTTP inside TCP,
inside IP, inside Wi-Fi frames. The café's router only ever looks at the
IP label to decide where to send the packet. That layering is why the
same web page works over Wi-Fi, 5G or fibre.`,
	},
	"TCP: Handshake, Flow & Congestion Control": {
		Analogy: `A careful phone call where the listener says "got it" after every
sentence, and the speaker repeats anything that was not acknowledged.
Congestion control is like merging onto a motorway: start slowly, speed
up while the road is clear, and ease off as soon as you see brake
lights.`,
		Example: `Video calls freeze on bad Wi-Fi because packets are lost and must be
resent. Big downloads start slowly and then speed up; that is slow start.
Google created BBR congestion control and the QUIC protocol (used by
HTTP/3) to cut these delays.`,
	},
	"Sockets & I/O Multiplexing": {
		Analogy: `A restaurant waiter with twenty tables. The bad approach is to stand at
one table until those diners decide (one thread per connection). The
good approach is to watch the whole room and go to whichever table
raises a hand (epoll). One waiter can then serve the entire room.`,
		Example: `The nginx web server handles huge numbers of connections with only a few
worker processes because it uses epoll. Node.js's event loop works the
same way. In 2012, WhatsApp described running about two million
connections on a single server.`,
	},
	"Time, Clocks & Ordering": {
		Analogy: `In a busy group chat, messages sometimes appear out of order, and
everyone's phone clock is slightly off. But a reply always comes after
the message it replies to. That "reply-to" chain is what Lamport clocks
capture: the order of cause and effect, not wall-clock time.`,
		Example: `Google's Spanner database uses GPS receivers and atomic clocks (TrueTime)
to keep clock uncertainty tiny. On New Year's 2017, a leap second made
time appear to go backwards inside Cloudflare's DNS software and caused
an outage.`,
	},
	"Consensus & Replication (Raft)": {
		Analogy: `A committee taking minutes. The chair (the leader) proposes each
decision, and it becomes official once a majority of members have signed
it. If the chair disappears, the members hold an election and carry on.
Nobody needs everyone present, just a majority.`,
		Example: `Kubernetes, which runs a large share of the world's cloud software,
keeps all of its cluster state in etcd, and etcd uses Raft. If one of
three etcd servers dies, the cluster keeps working. CockroachDB and
Consul are built on Raft as well.`,
	},

	// ---------------- Stage 13: Mathematical Foundations ----------------
	"Matrices as Linear Maps": {
		Analogy: `Think of the tools in a photo editor: stretch, rotate, skew, flip. Each
is a matrix applied to the position of every pixel. Doing one tool after
another is multiplying the matrices together.`,
		Example: `Every 3D video game multiplies millions of vertex positions by matrices
every frame to move the camera and the characters. In AI, each layer of a
neural network is a matrix that turns one list of numbers (say, pixel
brightness) into another list of more useful numbers (features).`,
	},
	"Eigenvectors & the SVD": {
		Analogy: `Describing a face to a sketch artist: you start with the most important
features (face shape, eye spacing) and leave out freckles. The SVD ranks
the "directions" in your data from most to least important, so you can
keep the few that matter and drop the rest.`,
		Example: `Matrix factorisation of the huge users × movies ratings table was a key
technique in the Netflix Prize (2006–2009) for movie recommendations.
Today, LoRA uses the same low-rank idea to let people fine-tune large AI
models on a single consumer GPU.`,
	},
	"Gradients, Jacobians & the Matrix Chain Rule": {
		Analogy: `Gears: if gear A turns gear B twice as fast, and B turns C three times as
fast, then A turns C six times as fast. The chain rule simply multiplies
these rates along the chain. A gradient is like the arrow on a hiking map
that points up the steepest slope from where you stand.`,
		Example: `Training a large language model means computing, trillions of times,
how much each of its billions of weights affected the error. That is
the chain rule applied through every layer. Engineers use the same
sensitivity idea to ask "if this input changes a little, how much does
the output move?"`,
	},
	"Probability, Softmax & Cross-Entropy": {
		Analogy: `A talent show's raw judge scores get turned into percentages that add up
to 100%. That is softmax. Cross-entropy measures how surprised you were:
if the model gave only 1% to the right answer, that is a big surprise and
a big penalty; if it gave 90%, only a small one.`,
		Example: `Your phone keyboard's next-word suggestions come from scores for every
possible word, turned into probabilities with softmax. Spam filters
output a probability that a message is spam, and they are trained with
cross-entropy loss.`,
	},
	"Optimisation: From Gradient Descent to Adam": {
		Analogy: `Walking down a mountain in thick fog. You cannot see the valley, but you
can feel the slope under your feet, so you step downhill. The learning
rate is your stride length: huge strides overshoot the valley and tiny
ones take forever. Momentum is like a rolling ball that keeps going
through small bumps.`,
		Example: `Every AI model you have used, from photo tagging to video
recommendations to chatbots, was trained by this exact loop: measure the
error, compute the slope, take a step, and repeat millions of times.`,
	},

	// ---------------- Stage 15: Neural Networks & Autograd ----------------
	"Computational Graphs & Reverse-Mode Autodiff": {
		Analogy: `A soup tastes too salty. You work backwards through the recipe: how much
salt did the stock add, how much did the soy sauce add, and how much
came from the stock cube inside the stock? Tracing blame backwards,
step by step, is exactly what backpropagation does for a model's error.`,
		Example: `When you call loss.backward() in PyTorch, or use TensorFlow or JAX, this
is what runs. It is the engine underneath essentially every modern AI
model, and in this stage you build your own.`,
	},
	"MLPs, Activations & Initialisation": {
		Analogy: `Stacking magnifying glasses gives you just one stronger magnifying glass.
Stacking linear layers is the same: without an activation function in
between, a deep network is really a single layer. Initialisation is like
setting the volume for an orchestra: too loud and it becomes noise, too
quiet and you hear nothing by the last row.`,
		Example: `Deep networks were notoriously hard to train until better activations
(ReLU, used by AlexNet in 2012) and better initialisation schemes made
depth practical. That is part of why deep learning took off in the
2010s.`,
	},
	"Residual Connections & Normalisation": {
		Analogy: `A game of telephone through a hundred people distorts the message
badly. A residual connection also passes the original written note down
the line, so each person only adds small corrections. Normalisation is
like a volume limiter that keeps everyone speaking at a normal level.`,
		Example: `ResNet (2015) used residual connections to train networks 152 layers
deep and won the ImageNet competition that year. Every modern transformer
language model uses residual connections and normalisation in every
block.`,
	},
	"Attention & the Transformer": {
		Analogy: `Take the sentence "The animal didn't cross the street because it was
too tired." To understand "it", you look back and link it to "animal".
Attention lets every word look at every other word and decide which ones
matter. It works like a search: the query is what I am looking for, the
keys are the labels on the other words, and the values are what they
actually say.`,
		Example: `Transformers power modern machine translation and chat assistants, and
AlphaFold 2 used attention to predict protein structures, work that
shared the 2024 Nobel Prize in Chemistry.`,
	},
	"The Training Loop & Generalisation": {
		Analogy: `A student who memorises last year's exam answers (overfitting) versus
one who actually understands the subject. The validation set is a mock
exam the student has never seen. It is the only honest measure of what
they have learned.`,
		Example: `Medical-imaging models have been caught "cheating". A well-known example
is a skin-lesion classifier that learned to associate the rulers doctors
place next to suspicious moles with cancer, instead of learning what the
lesion itself looks like. It scored well in the lab and would fail
elsewhere.`,
	},

	// ---------------- Stage 16: AI Infrastructure ----------------
	"The GPU Execution Model": {
		Analogy: `A CPU is a few brilliant professors who can each handle anything. A GPU
is a stadium of thousands of students each doing simple arithmetic at
the same moment. The students sit in rows of 32 (warps) that must do the
same step together; if half a row needs to do something different, the
other half waits.`,
		Example: `The same chips that draw video-game frames, where millions of pixels can
each be computed independently, are the chips that train and run AI
models. That overlap is how NVIDIA went from a gaming company to the
centre of the AI industry.`,
	},
	"GPU Memory Hierarchy & Coalescing": {
		Analogy: `Moving house: one truck trip carrying 32 boxes from the same room,
versus 32 separate trips to pick up boxes scattered around town.
Coalesced access is the single trip. Shared memory is the workbench right
next to you, where you keep what you are using right now.`,
		Example: `Optimised libraries such as cuBLAS and FlashAttention get most of their
speed from arranging memory accesses cleverly, not from doing less
arithmetic. The same maths can be ten times faster just by reading
memory in a better order.`,
	},
	"The Roofline Model: Compute- vs Memory-Bound": {
		Analogy: `A factory can be limited by how fast the workers assemble parts
(compute) or by how fast the delivery trucks bring parts in (memory
bandwidth). If the trucks are the bottleneck, hiring more workers
achieves nothing. The roofline tells you which of the two is holding you
back.`,
		Example: `When you run a language model on your own laptop (for example with
llama.cpp), tokens per second follow memory bandwidth, not processing
power. That is why Macs with high-bandwidth unified memory became popular
for running models locally.`,
	},
	"Kernel Fusion & FlashAttention": {
		Analogy: `Cooking dinner: instead of chopping the vegetables, putting them in the
fridge, taking them out to fry, putting them back, and taking them out
again to serve, you do every step at the counter in one go. The
fridge is GPU main memory; every trip to it costs time.`,
		Example: `FlashAttention is now used in most large-model training and serving
software; PyTorch includes fused attention kernels in
scaled_dot_product_attention. It is one of the main reasons today's
long context windows are affordable.`,
	},
	"KV Cache, Batching & PagedAttention": {
		Analogy: `While reading a long novel, you keep notes so you do not have to re-read
every earlier page before each new sentence; that is the KV cache.
PagedAttention stores those notes on fixed-size index cards filed
wherever there is space, with an index to find them, instead of needing
one giant empty shelf reserved in advance.`,
		Example: `Chat services answer many users at once on the same GPUs by batching
their requests together. The PagedAttention paper, which introduced the
vLLM serving system, reports 2–4× higher throughput than earlier serving
systems such as FasterTransformer and Orca, at the same latency, mostly
by no longer wasting KV-cache memory.`,
	},
}

// extras is keyed by stage ID.
var extras = map[int]stageExtras{
	5: {
		Outcomes: []string{
			"Explain what really happens between typing code and the CPU running it",
			"Read simple assembly output and see why one version of a function is faster",
			"Build an interpreter for your own small programming language",
		},
		Glossary: []Term{
			{"CPU", "The chip that executes instructions, one tiny step at a time, billions of times per second."},
			{"Instruction", "One basic command the CPU understands, such as 'add these two numbers' or 'jump to here'."},
			{"Register", "A tiny, ultra-fast storage slot inside the CPU; the CPU can only do maths on values held in registers."},
			{"Assembly", "A human-readable spelling of machine instructions, such as 'add rax, rbx'."},
			{"Compiler", "A program that translates code you write (C, Go, Rust) into machine instructions."},
			{"Linker", "The tool that stitches separately compiled pieces into one runnable program."},
			{"Cache", "A small, fast copy of recently used memory kept close to the CPU."},
			{"Byte", "8 bits. A bit is a single 0 or 1, and a byte can hold a number from 0 to 255."},
			{"ABI", "Application Binary Interface: the agreed rules for how compiled functions call each other."},
			{"AST", "Abstract Syntax Tree: code turned into a tree that shows its structure."},
		},
	},
	6: {
		Outcomes: []string{
			"Explain what the operating system does for every program you run",
			"Understand crashes such as 'segmentation fault' and 'out of memory'",
			"Write your own shell and your own memory allocator",
		},
		Glossary: []Term{
			{"Kernel", "The core of the operating system. It runs with full control of the hardware and referees all programs."},
			{"Process", "A running program with its own private memory."},
			{"Thread", "One line of execution inside a process; threads in the same process share memory."},
			{"Virtual address", "The address a program uses. The hardware translates it to a real location in RAM."},
			{"Page", "A fixed-size chunk of memory (usually 4 KiB), the unit the OS manages memory in."},
			{"System call", "A request from a program to the kernel, such as 'open this file' or 'send this data'."},
			{"Context switch", "Pausing one thread and resuming another on the same CPU core."},
			{"Mutex", "A lock that lets only one thread at a time into a critical piece of code."},
			{"Race condition", "A bug where the result depends on which thread happens to run first."},
		},
	},
	7: {
		Outcomes: []string{
			"Explain how databases find one row among billions in milliseconds",
			"Explain how data survives crashes and power cuts",
			"Build your own crash-safe key-value database",
		},
		Glossary: []Term{
			{"Database", "A program that stores data safely and lets you find it quickly."},
			{"Index", "An extra structure, like a book's index, that makes lookups fast without scanning everything."},
			{"fsync", "The system call that forces data out of temporary memory and onto the physical disk."},
			{"WAL", "Write-Ahead Log: a diary of every change, written before the change itself."},
			{"SSTable", "Sorted String Table: an unchangeable file of sorted key-value pairs."},
			{"Compaction", "Background tidying that merges files and throws away old or deleted data."},
			{"Transaction", "A group of changes that happens completely or not at all."},
			{"ACID", "Atomicity, Consistency, Isolation, Durability: the four promises a transaction makes."},
			{"SQL", "Structured Query Language: the standard language for asking relational databases questions."},
			{"Primary key", "A column (or columns) that uniquely identifies each row in a table."},
			{"JOIN", "A SQL operation that combines rows from two tables that share a matching value."},
		},
	},
	8: {
		Outcomes: []string{
			"Trace exactly what happens when you open a web page",
			"Write servers that handle thousands of users at once",
			"Explain how systems such as Kubernetes stay running when machines fail",
		},
		Glossary: []Term{
			{"Packet", "A small chunk of data sent across a network with an address label on it."},
			{"IP address", "The address of a machine on a network, such as 192.168.1.10."},
			{"Port", "A number that picks which program on a machine gets the data (the web usually uses 443)."},
			{"Protocol", "The agreed rules of a conversation between machines (TCP, HTTP, DNS)."},
			{"Latency", "How long one message takes to arrive: a delay, measured in milliseconds."},
			{"Bandwidth", "How much data can flow per second, like the width of a pipe."},
			{"Socket", "The program's handle for one network connection, which it reads from and writes to."},
			{"Node", "One machine (or process) in a distributed system."},
			{"Consensus", "Getting several machines to agree on the same value even when some fail."},
			{"DNS", "Domain Name System: the internet's phone book, turning names like example.com into IP addresses."},
			{"HTTP", "The request/response protocol browsers and apps use to fetch web pages and data."},
			{"CDN", "Content Delivery Network: servers around the world that keep copies of content close to users."},
		},
	},
	13: {
		Outcomes: []string{
			"Read the maths in machine-learning papers without fear",
			"Derive by hand how a neural network layer learns",
			"Understand what 'training' actually does with the numbers",
		},
		Glossary: []Term{
			{"Vector", "An ordered list of numbers, such as [3, 1, 4], which can be seen as an arrow or a point in space."},
			{"Matrix", "A grid of numbers that transforms vectors: rotating, stretching or mixing them."},
			{"Dot product", "Multiply two vectors element by element and add up the results; it measures how aligned they are."},
			{"Derivative", "How fast something changes when you nudge its input a tiny bit."},
			{"Gradient", "All the derivatives together: the arrow pointing towards the steepest increase."},
			{"Probability distribution", "A list of possible outcomes and how likely each is; the chances add up to 1."},
			{"Logit", "A raw score from a model before it is turned into a probability."},
			{"Learning rate", "How big a step the model takes each time it adjusts its weights."},
		},
	},
	15: {
		Outcomes: []string{
			"Build a neural network library from nothing, including backpropagation",
			"Train models that recognise handwritten digits",
			"Explain, and build, a small GPT-style language model",
		},
		Glossary: []Term{
			{"Neuron", "A tiny function: a weighted sum of its inputs passed through an activation function."},
			{"Weight", "A number the model learns; knowledge is stored in millions or billions of them."},
			{"Loss", "A single number measuring how wrong the model's predictions are; training makes it smaller."},
			{"Backpropagation", "Working out, backwards through the network, how each weight contributed to the loss."},
			{"Activation function", "A simple non-linear function (such as ReLU) applied after each layer."},
			{"Batch", "A handful of training examples processed together in one step."},
			{"Token", "A piece of text, such as a word or part of a word, that a language model reads and writes."},
			{"Embedding", "A list of numbers that represents a token (or anything else) in a way the model can compute with."},
		},
	},
	16: {
		Outcomes: []string{
			"Explain why AI needs GPUs and what makes a GPU program fast",
			"Predict how fast a model will run on given hardware, before you run it",
			"Build pieces of a real LLM inference engine",
		},
		Glossary: []Term{
			{"GPU", "A processor with thousands of simple cores, built for doing the same maths on lots of data at once."},
			{"CUDA", "NVIDIA's platform and language extension for writing programs that run on GPUs."},
			{"Kernel", "Here, a function that runs on the GPU across thousands of threads (not the OS kernel)."},
			{"HBM", "High Bandwidth Memory: the GPU's main memory, very fast but still far slower than on-chip memory."},
			{"FLOP", "A floating-point operation, such as one multiply or one add; FLOP/s measures compute speed."},
			{"Throughput", "Total work done per second, for example tokens per second across all users."},
			{"Quantisation", "Storing numbers with fewer bits (for example 8 or 4 instead of 16) to save memory and bandwidth."},
			{"Inference", "Using a trained model to produce answers, as opposed to training it."},
		},
	},
}

// startHere is the orientation guide shown from the main menu.
const startHere = `Welcome. This academy is built for self-learners. You do not need a
degree, prior experience or anyone's permission. Every idea is explained
from scratch, in plain words first, with real-world examples, before any
jargon appears, and every idea comes with exercises drawn from real
situations.

WHAT IT COVERS
Sixteen stages in four tracks, covering the core of a computer science
degree plus modern AI:
- Track F, Foundations: programming, data structures and algorithms, discrete maths, digital logic and computer architecture.
- Track A, Systems: compilers and the machine, operating systems, databases and storage, networks and the web.
- Track S, Software, Security & Theory: software engineering, security and cryptography, theory of computation, programming languages.
- Track B, AI: linear algebra and calculus, statistics and classical ML, neural networks, AI infrastructure and GPUs.

HOW EACH CONCEPT IS TAUGHT
- In plain words: an everyday analogy with no jargon at all.
- Real life: where you have already met this idea without knowing it.
- The details: the precise explanation, once the intuition is in place.
- Diagram: a picture of the mechanism, where one helps.
- Mental model: one sentence to carry around in your head.
- Exercises: three per concept, from easy to real-world (see below).
- Connects to: related ideas in other stages, so knowledge forms a web instead of a list.
- Go deeper: exactly which chapter or lecture of a verified resource to read next.
- Words to know: any jargon in the card, defined at the bottom.

HOW YOU WILL REMEMBER IT
Reading something once is not learning it; most of it fades within days.
This academy uses five techniques that research on memory consistently
supports:
- Retrieval practice: the Daily Review [9] asks you questions instead of showing you answers. The effort of recalling is what makes memories last.
- Spacing: each card comes back after 1 day, then 3, then about a week, then weeks and months, a little before you would forget it.
- Interleaving: reviews and mastery checks mix topics and stages, which trains you to recognise which idea applies.
- Your own words: when you mark a concept as understood, you explain it in your own words. Your explanation is shown back to you at review time.
- Connections: every concept links to related ones, so each new idea hooks onto something you already know.

MASTERY LEVELS
Each concept climbs a ladder: ○ new → ◔ understood → ◑ practised (two
exercises done) → ◕ retained (every review card remembered a week apart)
→ ● mastered (all exercises done and remembered three weeks apart). You
cannot reach mastered in a day, and that is the point: it measures what
you still know weeks later. Each stage also has a Mastery Check, a mixed
exam you pass at 80%.

CLASSIC MODE: LEARN LIKE IT'S 1985
Press c on the main menu. Classic Mode brings back the habits that made
the best 1980s and 1990s learners so strong:
- Read one anchor book per stage, cover to cover, plus a classic text from the field's history and real source code (see each Study Hall's Classic corner).
- Type programs in by hand, like magazine listings, and predict the output before running them.
- Struggle before you peek: hints unlock after 10, 30 or 45 minutes, or earlier if you write down what you tried.
- Keep a lab notebook of plans, predictions and what actually happened.
- Work in offline blocks with only man pages and saved docs, and keep a paper notebook beside you.
The modern help is still there; you just earn it first.

HOW TO PRACTISE
Every concept has three exercises, and each has a hint:
- Warm-up: pen and paper, or a few minutes of thinking. No code needed.
- Practice: a small program that makes the idea concrete.
- Real-world: a task from a real situation, such as a shop, a bank, your own files or a real dataset.

Open the Exercise Gym in any stage's Study Hall to see hints and tick
exercises off. Try each one before you peek at the hint.

WHERE TO START
- Complete beginner? Start at Stage 1, Programming Fundamentals, and follow Track F in order.
- After Track F, take Tracks A and S in either order.
- For AI, Stages 1 to 3 are enough preparation for Stage 13.
- Each stage shows what it builds on. If those stages feel shaky, a quick visit back first pays off.
- Stuck on a word? Open the Glossary in the stage's Study Hall.

A SIMPLE STUDY RHYTHM
- Start every session with the Daily Review [9]. It takes 5 to 15 minutes and protects everything you have learned.
- Then one new concept. Read it slowly, then close your eyes and explain it out loud.
- Do its warm-up the same day and its practice exercise the next.
- Mark a concept as understood only when you could explain it to a friend.
- Take the stage's quiz at the end of the week, and its Mastery Check when every concept is understood.
- Finish each stage with one real-world exercise or lab blueprint. Building is where it sticks.
- Log your hours. The streak and the activity chart are there to keep you going.
- Set weekly goals [g] you can keep on a bad week, and watch them on the dashboard. Consistency beats intensity.
- The Roadmap [m] shows every stage, what each builds on, and where you are.
- Not sure what to do? Press n for What's next: it picks your best next step and starts it for you.
- Study in focused blocks with the Focus timer [f]: 25 minutes on one task, phone away, then a short break. Sessions are logged automatically.
- Once a week, open the Progress report [p] to see your calendar, your trend against last week, and the cards you forget most.
- Looking for something? Press / to search every concept, glossary term, resource and classic.
- Press x to export your notes, explanations and progress to a Markdown file you can keep, print or share.

IF IT FEELS HARD
That is normal and means you are learning. Re-read the analogy, look at
the real-life example, then try one resource from the library. Nobody
understands this material on the first pass. You have unlimited time and
no exams.

BEYOND THE CORE
Once the sixteen stages feel comfortable, these electives build on them:
- Computer graphics: Scratchapixel (scratchapixel.com)
- Quantum computing: IBM Quantum Learning (learning.quantum.ibm.com)
- Embedded systems and electronics: Arduino docs (docs.arduino.cc)
- Bioinformatics: Rosalind problems (rosalind.info)
- Game development, robotics, human-computer interaction and computer vision all reuse the ideas you have learned here.

RESOURCES YOU CAN TRUST
Every resource link was reviewed on ` + resourcesVerifiedOn + `. Links
can move over time, so run "academy -check-links" on any computer with
internet access to re-verify all of them.

YOUR DATA
Everything you do, including review schedules and your own explanations,
is saved in one readable JSON file next to the program. Nothing is sent
anywhere.`
