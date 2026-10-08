package main

// Dictionary: words that sit one level deeper (how machines, databases,
// distributed systems, models and defences really work), and the "Go
// deeper" notes that take a word past its plain meaning: how it works
// underneath, why it exists, and how experienced engineers think about it.

var vocabDeep = []Word{
	// ---- Systems & concurrency
	{"Page fault", []string{"page faults", "major fault", "minor fault"}, "Systems & concurrency", 6,
		"What happens when a program touches a piece of its memory that is not in RAM right now. The CPU stops, the operating system finds or loads that page, and the program continues as if nothing happened.",
		"Asking a librarian for a book that is in the basement store: you wait while they fetch it, then you carry on reading.",
		`$ /usr/bin/time -v ./program
  Major (requiring I/O) page faults: 12
  Minor (reclaiming a frame) page faults: 4810`,
		"Page faults are normal, not errors. A segmentation fault is a page fault the operating system refuses to fix, because the address was never yours.",
		[]string{"Virtual memory", "Segmentation fault", "Kernel", "TLB"}},
	{"Context switch", []string{"context switching"}, "Systems & concurrency", 6,
		"The operating system pausing one thread and resuming another on the same CPU core: it saves the registers of one and loads the other's.",
		"A chef with one stove cooking three dishes: put one pan aside, note where you were, pick up the next.",
		`$ vmstat 1
 r  b ... in   cs   us sy id
 2  0 ... 812 2410  9  3 88    # cs = context switches per second`,
		"Switches are cheap one at a time (microseconds) but add up: thousands of threads fighting over a few cores spend their time switching, not working.",
		[]string{"Thread", "Process", "Interrupt", "Kernel"}},
	{"Interrupt", []string{"interrupts", "IRQ", "hardware interrupt"}, "Systems & concurrency", 4,
		"A signal from hardware (a key press, a network packet, a timer) that makes the CPU stop what it is doing and run a small handler in the operating system.",
		"A doorbell: you put down what you are doing, answer the door, then go back to it.",
		`$ cat /proc/interrupts      # Linux: how many interrupts each device raised
  0:   45  IO-APIC  2-edge  timer
 24: 8812  PCI-MSI  eth0`,
		"Interrupt handlers must be very short; the real work is handed off to be done later, or the machine stops responding.",
		[]string{"Kernel", "Context switch", "System call"}},
	{"Copy-on-write", []string{"COW", "copy on write"}, "Systems & concurrency", 6,
		"Sharing data between two users until one of them changes it; only then is a private copy made. Operating systems use it to make fork() fast.",
		"Two students sharing one textbook until one wants to write in the margins; then they get their own copy.",
		`pid_t pid = fork();   // child shares the parent's memory pages...
if (pid == 0) x = 1;  // ...until it writes: then that page is copied`,
		"Copy-on-write saves memory only while nobody writes. A forked process that rewrites most of its memory ends up copying it all anyway.",
		[]string{"Process", "Virtual memory", "Page fault"}},
	{"Memory-mapped file", []string{"mmap", "memory mapping"}, "Systems & concurrency", 6,
		"A file made to look like part of a program's memory: reading the bytes reads the file, and the operating system loads pages as they are touched.",
		"Instead of photocopying a book page by page, you are given the book open on your desk; the pages you turn to appear when you need them.",
		`int fd = open("data.bin", O_RDONLY);
char *p = mmap(NULL, size, PROT_READ, MAP_PRIVATE, fd, 0);
char first = p[0];      // the OS reads that page from disk on demand`,
		"An I/O error while reading a mapped file arrives as a crash (SIGBUS), not as an error code, so mmap suits files you trust to be there.",
		[]string{"Virtual memory", "Page fault", "File descriptor"}},
	{"TLB", []string{"translation lookaside buffer"}, "Systems & concurrency", 4,
		"A tiny, very fast cache inside the CPU that remembers recent translations from virtual addresses to physical memory.",
		"A speed-dial list for the few numbers you call most, so you rarely open the full phone book.",
		`$ perf stat -e dTLB-load-misses ./program
     1,204,551  dTLB-load-misses`,
		"Jumping randomly around a huge amount of memory causes TLB misses on top of cache misses; huge pages (2 MB instead of 4 KB) are the usual fix.",
		[]string{"Virtual memory", "Cache", "Page fault"}},
	{"Cache line", []string{"cache lines"}, "Systems & concurrency", 4,
		"The unit the CPU cache moves memory in: usually 64 bytes. Reading one byte brings in its whole 64-byte neighbourhood.",
		"Shopping by the box, not by the item: ask for one egg and you carry home the whole carton.",
		`int a[1024][1024];
for (i...) for (j...) sum += a[i][j];   // fast: walks each line in order
for (j...) for (i...) sum += a[i][j];   // slow: a new line for every read`,
		"Layout decides speed: data used together should sit together. Arrays of structs and structs of arrays can differ several times in speed.",
		[]string{"Cache", "False sharing", "Memory"}},
	{"False sharing", []string{}, "Systems & concurrency", 6,
		"Two threads slowing each other down because the different variables they write happen to sit in the same cache line, which the cores then pass back and forth.",
		"Two people writing in different boxes of the same form, but there is only one pen-holder: each must wait for the form to be handed over.",
		`struct { long a; long b; } counters;   // a and b share a cache line
// thread 1 does counters.a++, thread 2 does counters.b++: both crawl
struct { alignas(64) long a; alignas(64) long b; } fixed;`,
		"The code looks perfectly parallel and correct; only a profiler shows the cost. Pad hot per-thread data to separate cache lines.",
		[]string{"Cache line", "Thread", "Race condition"}},
	{"Branch prediction", []string{"branch predictor", "branch misprediction"}, "Systems & concurrency", 4,
		"The CPU guessing which way an if will go before it knows, so it can keep working ahead. A wrong guess throws away that work.",
		"A waiter bringing your usual coffee as you walk in; if you order tea instead, the coffee is poured away.",
		`// Summing only large values of an array:
for (i...) if (v[i] >= 128) sum += v[i];
// can run several times faster on sorted data: the branch becomes predictable`,
		"Predictable branches are almost free and random ones are costly, so the same code can run at very different speeds on different data.",
		[]string{"Cache line", "Cache"}},
	{"Two's complement", []string{"twos complement", "signed integer representation"}, "Types & data", 0,
		"The way computers store negative integers: flip every bit of the positive number and add one. The same adding circuit then works for positive and negative numbers.",
		"A car's odometer rolling back from 0000 to 9999: going one below zero lands on the largest value.",
		`int8_t x = -1;        // stored as 1111 1111
int8_t y = 127 + 1;   // overflows: wraps to -128 (1000 0000) in practice`,
		"In C, overflowing a signed integer is undefined behaviour, so the compiler may assume it never happens; do not rely on the wrap-around.",
		[]string{"Bit", "Integer overflow", "Undefined behavior"}},
	{"Calling convention", []string{"ABI", "application binary interface"}, "Memory & C", 5,
		"The agreed rules for how a function call works in machine code: which registers carry the arguments, where the return value goes, who cleans up the stack. The ABI is the full set of such rules for a platform.",
		"The rules of a relay race: where the baton is handed over and with which hand, so runners from different teams can still work together.",
		`// x86-64 Linux (System V ABI): the first integer arguments go in
// rdi, rsi, rdx, rcx, r8, r9; the result comes back in rax.
long add(long a, long b) { return a + b; }   // lea rax, [rdi+rsi]; ret`,
		"Libraries compiled with different ABIs cannot call each other safely, even when the source code matches; that is why C is the common language between languages.",
		[]string{"Stack frame", "Register", "Linker"}},

	// ---- Data & databases
	{"Write-ahead log", []string{"WAL", "journal", "journaling"}, "Data & databases", 7,
		"A log where a database writes what it is about to change before changing the data itself. After a crash it replays the log, so no committed change is lost.",
		"Writing in your diary \"moving the sofa to the left wall\" before moving it; if you are interrupted halfway, the note tells you what to finish.",
		`-- PostgreSQL: every change is written to the WAL first
SHOW wal_level;            -- replica
SELECT pg_current_wal_lsn();`,
		"The log only protects you if it really reaches the disk: a database that skips fsync to look fast can lose committed data in a power cut.",
		[]string{"Transaction", "Index", "Replication"}},
	{"LSM tree", []string{"log-structured merge tree"}, "Data & databases", 7,
		"A storage design that turns writes into fast appends: new data goes to memory, is flushed to sorted files, and background work merges those files. Used by RocksDB, LevelDB and Cassandra.",
		"Throwing incoming mail into today's tray, then sorting the trays into the filing cabinet in a quiet hour.",
		`write  → memtable (memory) → flushed to a sorted file on disk
read   → memtable, then the newest files, then older ones
compaction merges files in the background`,
		"Writes are cheap, but a read may have to look in several files, and compaction can suddenly use a lot of disk bandwidth.",
		[]string{"B-tree", "Bloom filter", "Write-ahead log"}},
	{"Bloom filter", []string{"bloom filters"}, "Data & databases", 7,
		"A tiny data structure that answers \"is this item in the set?\" with either \"definitely not\" or \"probably yes\", using far less memory than storing the items.",
		"A bouncer who remembers a few features of every guest: if nothing matches, you are certainly not on the list; if everything matches, you are probably on it.",
		`add(x):    set bits h1(x), h2(x), h3(x)
maybe(x):  are all three bits set?   // false positives possible,
                                     // false negatives never`,
		"You cannot delete items from a plain Bloom filter, and it fills up: past its planned size, almost every answer becomes \"probably yes\".",
		[]string{"Hash function", "LSM tree", "Hash table"}},
	{"Join", []string{"SQL join", "inner join", "left join"}, "Data & databases", 7,
		"Combining rows from two tables that match on a column, such as orders with the customers who placed them.",
		"Matching guest names on a seating plan with names on the meal-choice list to know what each table eats.",
		`SELECT c.name, o.total
FROM orders o
JOIN customers c ON c.id = o.customer_id;`,
		"A join without the right condition, or without an index on the joined column, can multiply rows or scan whole tables; check the plan with EXPLAIN.",
		[]string{"Index", "Normalization", "SQL"}},
	{"Normalization", []string{"normal form", "normalize"}, "Data & databases", 7,
		"Designing tables so each fact is stored once: customers in one table, orders in another, linked by an id, instead of copying a customer's address into every order.",
		"Keeping one contact card per friend, rather than writing their address on every birthday reminder.",
		`-- one fact, one place
customers(id, name, address)
orders(id, customer_id, total)    -- refers to customers.id`,
		"Duplicated facts drift apart (two addresses for one customer). Denormalizing on purpose for speed is fine, as long as you know which copy is the truth.",
		[]string{"Join", "Index", "Database"}},
	{"Replication", []string{"replica", "replicas", "leader-follower"}, "Networks & web", 8,
		"Keeping copies of the same data on several machines, so it survives a failure and more readers can be served.",
		"Keeping a copy of your important documents at a friend's house as well as your own.",
		`primary  ──writes──▶  WAL  ──streamed──▶  replica 1
                                  └──────▶  replica 2   (read-only)`,
		"Replicas can lag behind: a user who writes and then reads from a replica may not see their own change yet.",
		[]string{"Consensus", "Eventual consistency", "Write-ahead log"}},
	{"Consensus", []string{"consensus algorithm", "Raft", "Paxos"}, "Networks & web", 8,
		"Getting several machines to agree on one value, or one order of events, even when some of them crash or messages are lost. Raft and Paxos are the classic algorithms.",
		"A committee that only records a decision once a majority has signed the minutes, so no single absent member can block or rewrite it.",
		`Raft: the leader appends an entry, sends it to the followers,
and commits it once a majority (3 of 5) has stored it.`,
		"Consensus tolerates crashed machines, not lying ones, and it needs a majority to be reachable: a 5-node cluster stops accepting writes when 3 are down.",
		[]string{"Quorum", "Replication", "Leader election"}},
	{"Quorum", []string{"majority quorum"}, "Networks & web", 8,
		"The minimum number of machines that must agree for a decision to count, usually a majority. Any two majorities overlap, so they cannot make conflicting decisions.",
		"A club rule that a vote needs more than half of the members present: two different \"majorities\" must share at least one member.",
		`5 nodes → quorum = 3
write to 3, read from 3: at least one node has the latest value`,
		"An even number of nodes buys nothing: 4 nodes tolerate one failure, exactly like 3.",
		[]string{"Consensus", "Replication"}},
	{"Leader election", []string{}, "Networks & web", 8,
		"How a group of machines chooses one of them to coordinate, and chooses a new one when the leader fails.",
		"A team picking a captain, and quickly picking another when the captain leaves the pitch.",
		`Raft: a follower that hears nothing from the leader for a random
timeout becomes a candidate and asks the others for votes;
a majority of votes makes it leader for a new term.`,
		"Two leaders at once (split brain) is the classic disaster; terms or fencing tokens make sure an old leader's commands are rejected.",
		[]string{"Consensus", "Quorum"}},
	{"CAP theorem", []string{"CAP"}, "Networks & web", 8,
		"When the network splits a distributed system in two (a partition), each part must choose between staying available and staying consistent; it cannot fully do both.",
		"Two shop branches lose their phone line: either both keep selling and may oversell the last item, or one stops selling until the line is back.",
		`partition happens → choose:
  CP: refuse some requests, never return stale data
  AP: answer everything, reconcile differences later`,
		"CAP is only about behaviour during a partition. Most of the time the real trade-off is latency against consistency.",
		[]string{"Eventual consistency", "Consensus", "Replication"}},
	{"Eventual consistency", []string{"eventually consistent"}, "Networks & web", 8,
		"A guarantee that, if writes stop, all copies of the data will end up the same, without promising when, or what a reader sees in the meantime.",
		"News reaching every village by word of mouth: everyone will hear it, but not at the same moment.",
		`write "likes = 10" to replica A
read  from replica B → 9   (not yet updated)
a moment later        → 10`,
		"\"Eventually\" can mean seconds or, during an outage, much longer. Design screens and logic that cope with briefly stale values.",
		[]string{"Replication", "CAP theorem"}},
	{"Two-phase commit", []string{"2PC"}, "Networks & web", 8,
		"A protocol that makes a transaction across several databases all-or-nothing: first everyone votes on whether they can commit, then the coordinator tells everyone the result.",
		"A wedding: the officiant asks both partners first; only if both say \"I do\" is the marriage declared.",
		`phase 1: coordinator → "prepare?"   each participant → "yes" / "no"
phase 2: all yes → "commit"         any no  → "abort"`,
		"If the coordinator crashes between the phases, participants that voted yes are stuck waiting; that blocking is why many systems prefer other designs.",
		[]string{"Transaction", "Consensus"}},
	{"Lamport clock", []string{"logical clock", "vector clock", "happens-before"}, "Networks & web", 8,
		"A counter, instead of real time, used to order events across machines: each machine increments it, and messages carry it, so a reply is always numbered after the request.",
		"Numbering letters in a long correspondence: you always reply with a number higher than the letter you received, even if your watches disagree.",
		`on local event:   clock += 1
on send:          clock += 1; attach clock
on receive(t):    clock = max(clock, t) + 1`,
		"Wall clocks on different machines drift, so timestamps alone cannot say which of two events happened first.",
		[]string{"Consensus", "Replication"}},

	// ---- Algorithms
	{"Dynamic programming", []string{"DP"}, "Algorithms", 2,
		"Solving a problem by solving its smaller overlapping sub-problems once, storing their answers, and building the big answer from them.",
		"Climbing stairs and writing on each step how many ways there are to reach it, so you never recount the steps below.",
		`// ways to climb n stairs taking 1 or 2 steps at a time
ways[0] = 1; ways[1] = 1;
for (int i = 2; i <= n; i++) ways[i] = ways[i-1] + ways[i-2];`,
		"It only helps when sub-problems repeat. The hard part is not the code but finding what the sub-problem is.",
		[]string{"Recursion", "Big O", "Greedy algorithm"}},
	{"Greedy algorithm", []string{"greedy"}, "Algorithms", 2,
		"An algorithm that always takes the choice that looks best right now, never reconsidering. Fast and simple; correct only for problems with the right structure.",
		"Paying with the largest coin that still fits: it works with euros, but with coins of 1, 3 and 4, making 6 greedily gives 4+1+1 instead of 3+3.",
		`// Dijkstra's shortest paths is greedy: always settle the closest
// unvisited node next. It is correct because edge weights are >= 0.`,
		"Always test a greedy idea against a small counterexample before trusting it, or prove why the local choice is safe.",
		[]string{"Dynamic programming", "Graph"}},
	{"Amortized analysis", []string{"amortized", "amortised"}, "Algorithms", 2,
		"Measuring the average cost per operation over a long sequence, when most operations are cheap and a rare one is expensive.",
		"A monthly bus pass: one big payment, but averaged over every trip, each ride is cheap.",
		`// A growing array doubles its size when full. One append may copy
// n items, but n appends copy fewer than 2n items in total:
// O(1) amortized per append.`,
		"Amortized is not \"every call is fast\": a single append can still pause noticeably, which matters in real-time code.",
		[]string{"Big O", "Dynamic array"}},

	// ---- AI & ML
	{"Softmax", []string{}, "AI & ML", 15,
		"A function that turns a list of scores into probabilities: each becomes positive, larger scores get larger shares, and they all add up to 1.",
		"Turning raw votes into percentages, except that a small lead is exaggerated: the top choice gets most of the pie.",
		`import numpy as np
z = np.array([2.0, 1.0, 0.1])
p = np.exp(z - z.max()) / np.exp(z - z.max()).sum()   # [0.66, 0.24, 0.10]`,
		"Subtract the largest score before exponentiating, or exp() overflows to infinity on big scores.",
		[]string{"Neural network", "Attention", "Gradient descent"}},
	{"Attention", []string{"self-attention", "attention mechanism"}, "AI & ML", 15,
		"The mechanism that lets each word (token) in a model look at all the others and decide which ones matter for it, as a weighted mix.",
		"Reading \"the animal didn't cross the road because it was tired\": to understand \"it\", you glance back at \"animal\", not at \"road\".",
		`scores  = Q @ K.T / sqrt(d)       # how well each query matches each key
weights = softmax(scores)         # a probability per pair
out     = weights @ V             # a weighted mix of the values`,
		"The cost grows with the square of the sequence length, which is why long contexts are expensive and why FlashAttention and KV caches exist.",
		[]string{"Transformer", "Softmax", "Embedding", "KV cache"}},
	{"Transformer", []string{"transformers"}, "AI & ML", 15,
		"The neural-network architecture behind today's language models: layers of attention and small feed-forward networks, processing all tokens of a sequence in parallel.",
		"A meeting where everyone hears everyone at once and updates their notes each round, instead of a message passed person to person down a line.",
		`tokens → embeddings + positions
  → [attention → feed-forward] × N layers
  → probabilities for the next token`,
		"A transformer predicts likely next tokens; it does not look facts up. Fluent output can be confidently wrong, so check what matters.",
		[]string{"Attention", "Tokenizer", "Embedding"}},
	{"KV cache", []string{"key-value cache"}, "AI & ML", 16,
		"Saved attention keys and values of the tokens already processed, so a language model generating text one token at a time does not recompute the whole past for each new token.",
		"Keeping your notes from the start of a long meeting instead of re-reading the full transcript before every sentence you say.",
		`without cache: token n recomputes keys/values for all n tokens
with cache:    token n computes its own, reads the rest from memory
memory ≈ 2 × layers × tokens × hidden size × bytes per number`,
		"The cache can become larger than the model for long contexts or many users; paging it, as PagedAttention does, is how servers fit more users.",
		[]string{"Attention", "Transformer", "Quantization"}},
	{"Quantization", []string{"quantized", "int8", "4-bit"}, "AI & ML", 16,
		"Storing a model's numbers with fewer bits (8 or 4 instead of 16 or 32), so it needs less memory and runs faster, at the cost of a little accuracy.",
		"Saving a photo as a smaller JPEG: most of the picture survives, and fine details may blur.",
		`weight 0.731 → stored as the int8 value 93 with scale 0.00786
93 × 0.00786 ≈ 0.731  (close, not exact)`,
		"Quality drops are uneven: some layers or tasks suffer much more than others. Always measure the quantized model on your own task.",
		[]string{"Floating point", "Neural network", "KV cache"}},
	{"Bias-variance tradeoff", []string{"bias and variance"}, "AI & ML", 14,
		"The tension between a model too simple to capture the pattern (high bias, underfitting) and one so flexible it learns the noise (high variance, overfitting).",
		"Describing a friend's handwriting: \"it's writing\" is too vague to recognise them (bias); memorising one letter's coffee stain is too specific (variance).",
		`degree-1 polynomial:  misses the curve        (high bias)
degree-15 polynomial: wiggles through noise   (high variance)
degree 3:             follows the trend       (balanced)`,
		"More data mainly cures variance, not bias; a model that is too simple stays wrong however much data it sees.",
		[]string{"Overfitting", "Regularization", "Cross-validation"}},
	{"Cross-validation", []string{"k-fold", "k-fold cross-validation"}, "AI & ML", 14,
		"Estimating how a model will do on new data by training it several times, each time holding out a different slice of the data for testing.",
		"Practising for an exam by testing yourself on a different chapter each evening, instead of always re-reading the one you already know.",
		`from sklearn.model_selection import cross_val_score
scores = cross_val_score(model, X, y, cv=5)   # 5 folds
print(scores.mean(), scores.std())`,
		"Any step that learns from the data (scaling, choosing features) must happen inside each fold; doing it on all the data first leaks the test set.",
		[]string{"Overfitting", "Bias-variance tradeoff"}},
	{"Precision and recall", []string{"recall", "F1 score"}, "AI & ML", 14,
		"Two ways to score a classifier. Precision: of the items it flagged, how many were right. Recall: of the items it should have flagged, how many it found.",
		"A fisherman's net: precision is how much of the catch is fish rather than boots; recall is how many of the fish in the lake ended up in the net.",
		`flagged 10 emails as spam, 8 really spam  → precision 0.8
there were 20 spam emails in total         → recall    0.4`,
		"Accuracy misleads on rare events: a fraud detector that never flags anything is 99.9% \"accurate\" and useless. Decide which mistake costs more.",
		[]string{"Overfitting", "Cross-validation"}},

	// ---- Security
	{"Defense in depth", []string{"layered security", "defence in depth"}, "Security", 17,
		"Protecting a system with several independent layers, so that when one fails (and one will), another still stops the attacker.",
		"A castle with a moat, walls, guards and a locked keep: crossing the moat does not open the treasury.",
		`input validation  +  parameterised queries  +  least-privilege DB user
+  network rules  +  monitoring and alerts`,
		"Layers must be truly independent: five checks that all trust the same login token are one layer, not five.",
		[]string{"Least privilege", "Threat model", "Sandbox"}},
	{"Nonce", []string{"number used once", "IV", "initialization vector"}, "Security", 10,
		"A value that must never be used twice with the same key: it keeps identical messages from producing identical ciphertexts, and stops old messages from being replayed.",
		"A raffle ticket number: each one is printed once, so a copied ticket is spotted immediately.",
		`nonce := make([]byte, gcm.NonceSize())
rand.Read(nonce)                          // fresh for every message
ct := gcm.Seal(nil, nonce, plaintext, nil)`,
		"Reusing a nonce with AES-GCM can reveal plaintexts and even let attackers forge messages. Generate it randomly, or from a counter that never repeats.",
		[]string{"Encryption", "Salt", "Public-key cryptography"}},
	{"Side channel", []string{"side-channel attack", "timing attack"}, "Security", 10,
		"Leaking secrets not through the data a program outputs, but through how it behaves: how long it takes, how much power it draws, which memory it touches.",
		"Guessing a safe's combination by listening to the clicks of the dial rather than reading the numbers.",
		`// leaks how many leading characters match, through timing
if (strcmp(input, secret) == 0) ...
// constant-time comparison instead:
if (CRYPTO_memcmp(input, secret, len) == 0) ...`,
		"Code that is logically correct can still leak. Compare secrets in constant time, and use vetted crypto libraries rather than your own.",
		[]string{"Encryption", "Hash function", "Threat model"}},
	{"TLS handshake", []string{"HTTPS handshake"}, "Security", 8,
		"The first few messages of a secure connection, where client and server agree on keys and the server proves its identity with a certificate, before any real data is sent.",
		"Two spies meeting: they exchange a code phrase to prove who they are, then agree on a cipher for the rest of the conversation.",
		`$ openssl s_client -connect example.com:443 -brief
Protocol version: TLSv1.3
Peer certificate: CN = example.com
Verification: OK`,
		"Encryption without verifying the certificate protects you from nobody in the middle; never disable certificate checks to make an error go away.",
		[]string{"TLS", "Public-key cryptography", "Encryption"}},

	// ---- Tools & workflow
	{"Tail latency", []string{"p99", "p99 latency", "percentile latency"}, "Tools & workflow", 9,
		"The slowest responses of a service, measured as percentiles: p99 is the time within which 99% of requests finish. The average hides them.",
		"A bus that is usually on time but one day in a hundred is an hour late: the average looks fine, and you remember the hour.",
		`p50 =  12 ms   (half of requests are faster)
p99 = 380 ms   (1 in 100 is slower than this)
a page that makes 100 calls hits a p99 response almost every time`,
		"When one user action fans out to many calls, the slowest call decides the experience, so p99 matters more than the mean.",
		[]string{"Latency", "SLO", "Observability"}},
	{"SLO", []string{"service level objective", "SLA", "SLI", "error budget"}, "Tools & workflow", 9,
		"A target for how reliable a service should be, such as \"99.9% of requests succeed within 300 ms over 30 days\". The gap below 100% is the error budget teams can spend on change.",
		"A train company promising that 99 of 100 trains arrive within five minutes, and slowing down its track works when it misses that.",
		`99.9% availability over 30 days → 43 minutes of failure allowed
budget spent early → slow down risky releases until it recovers`,
		"100% is the wrong target: it is impossible, and chasing it freezes all change. Pick what users actually notice.",
		[]string{"Tail latency", "Observability"}},
	{"Observability", []string{"telemetry", "tracing", "metrics"}, "Tools & workflow", 9,
		"How well you can tell what a running system is doing from the outside, using its metrics (numbers over time), logs (events) and traces (one request's path through many services).",
		"A car dashboard plus a trip log plus a dashcam: together they tell you not just that something went wrong, but where and why.",
		`metric: http_requests_total{status="500"}  rising
trace:  request 7f3a spent 2.1 s in payments-db
log:    "connection pool exhausted" at 14:02:11`,
		"Add context when you instrument (request id, user-facing operation), or you will have millions of data points and no answers.",
		[]string{"Tail latency", "SLO", "Debugging"}},
	{"Flaky test", []string{"flaky", "flakiness"}, "Tools & workflow", 9,
		"A test that sometimes passes and sometimes fails on the same code, usually because of timing, shared state, the network or randomness.",
		"A smoke alarm that goes off once a week for no reason: soon nobody reacts, including the day there is a real fire.",
		`time.Sleep(100 * time.Millisecond)   // hoping the server is ready: flaky
waitUntilReady(server, 5*time.Second) // waiting for the real signal`,
		"Re-running until green hides the problem. A flaky test is a bug in the test or in the code; find the race or the hidden dependency.",
		[]string{"Unit test", "Race condition", "Continuous integration"}},
}

// deeper holds the "Go deeper" notes, keyed by dictionary word: how the
// thing works underneath, where it came from, and how experts think
// about it. Paragraphs are separated by blank lines.
var deeper = map[string]string{
	"Pointer": `A pointer is just a number: the address of a byte in your program's virtual address space. Its type (int *, char *) tells the compiler how many bytes to read at that address and how far p + 1 moves (sizeof the type), which is why pointer arithmetic on an int * steps 4 bytes at a time.

Every language that hides pointers still uses them underneath. A Python list holds pointers to objects, a Java object variable is a pointer, and a Go slice is a small struct holding a pointer, a length and a capacity. Learning pointers in C is learning what all of those really are.

Experts think about ownership: for every pointer, who is responsible for freeing what it points to, and how long is it valid? Most pointer bugs (leaks, use-after-free, double free) are ownership bugs, which is exactly what Rust's borrow checker turns into compile errors.`,
	"Stack": `The call stack is one contiguous region of memory per thread, typically 8 MB on Linux. Calling a function moves the stack pointer register down to make room for its frame (return address, saved registers, local variables), and returning moves it back up. That is why allocating on the stack costs almost nothing: it is one subtraction.

It is also why returning the address of a local variable is a bug: the frame is reused by the next call. And why deep recursion crashes: each level takes a frame, and the stack has a fixed size.

Security history lives here: a buffer overflow on the stack can overwrite the saved return address and redirect the program. Stack canaries, non-executable stacks and address randomisation (ASLR) were all invented to stop that.`,
	"Heap": `The heap is memory you ask for while the program runs (malloc in C, new in C++, every object in Python or Java). An allocator manages it: it keeps lists of free blocks of different sizes and asks the operating system for more pages when it runs out (brk or mmap on Linux).

Heap allocation is far slower than the stack (finding a free block, bookkeeping, thread safety), and freed blocks can leave gaps that waste memory, which is called fragmentation. High-performance code reuses objects, allocates in bulk, or uses arenas: one big block handed out piece by piece and freed all at once.

Garbage-collected languages still use a heap; the collector decides when to free. Understanding allocation explains why creating millions of tiny objects in a hot loop is slow in any language.`,
	"Process": `A process is a running program plus everything the operating system keeps for it: its own virtual address space, open files, current directory, user identity and at least one thread. Isolation between processes is enforced by the hardware's memory management unit, which is why one crashing program cannot corrupt another.

On Unix, new processes are made by fork() (copy me, cheaply, with copy-on-write) followed by exec() (replace my program with another one). Your shell does that for every command you type.

Processes are the unit of isolation and of failure: when you want two pieces of software unable to break each other, put them in different processes. Containers are processes with extra isolation (namespaces) and resource limits (cgroups), not small virtual machines.`,
	"Thread": `Threads are several flows of execution inside one process, sharing its memory but each with its own registers and stack. The operating system schedules threads, not processes, onto CPU cores.

Sharing memory is what makes threads fast and dangerous: any two threads can touch the same variable at the same time. Without synchronisation (mutexes, atomics, channels) you get race conditions, and modern CPUs make it worse by reordering memory operations unless told not to.

There are other models: event loops (one thread, many tasks, as in Node.js), green threads or goroutines (many lightweight tasks on a few OS threads, as in Go), and processes with message passing (as in Erlang). Choosing a model is choosing which bugs you want to make impossible.`,
	"Dictionary": `A hash table turns a key into a number with a hash function, uses that number to pick a slot in an array, and stores the value there. Lookups take constant time on average because they jump straight to the right slot instead of searching.

Two keys can land in the same slot (a collision). Tables handle that by chaining (a small list per slot) or by open addressing (trying the next slots). As the table fills, collisions grow, so it resizes, typically doubling and moving every entry, which is amortised over many inserts.

Real-world twists: a predictable hash function lets attackers send keys that all collide and slow a server to a crawl, so languages randomise their hashes per process. That is also why iterating over a Python dict or a Go map does not follow any order you should rely on.`,
	"Cache": `Caches exist because of the memory gap: a CPU can do hundreds of operations in the time one read from RAM takes. Each core has small, very fast L1 and L2 caches, and the cores share a larger L3. Data moves between them in 64-byte cache lines.

Caches work because programs have locality: they reuse data they just used (temporal locality) and use data next to it (spatial locality). Code that walks arrays in order is fast; code that chases pointers all over memory is slow, even when both do the same number of operations.

The same idea repeats at every scale: the operating system's page cache keeps disk data in RAM, browsers and CDNs cache web pages, and databases cache hot rows. And the same hard problem follows it everywhere: knowing when a cached copy is out of date.`,
	"Race condition": `A race condition is a bug whose outcome depends on the timing of events you do not control. The classic one is two threads doing count++ at the same time: each reads the old value, adds one and writes back, so one increment is lost.

Races are hard because they are rare and timing-dependent: the test passes a thousand times and fails in production. Tools help: Go's -race flag and ThreadSanitizer for C and C++ detect unsynchronised access while the program runs.

They are not only about threads. Two web requests updating the same row, a check-then-act on a file (does it exist? then open it), or two machines writing to the same key are all races. The fix is always to make the check and the action one atomic step: a lock, a transaction, an atomic operation or a compare-and-swap.`,
	"Deadlock": `A deadlock happens when two or more tasks each hold something and wait for something another one holds, so nobody can continue. Four conditions must all hold: exclusive locks, holding while waiting, no forced release, and a cycle of waiting.

Break any one condition and deadlocks become impossible. The most practical rule is a global lock order: if every piece of code takes lock A before lock B, a cycle can never form.

Databases detect deadlocks by looking for cycles in their wait-for graph and abort one transaction, which then must be retried. That is why application code talking to a database should be ready to retry a transaction.`,
	"Encryption": `Modern encryption has two families. Symmetric ciphers (AES, ChaCha20) use one shared secret key and are extremely fast. Asymmetric cryptography (RSA, elliptic curves) uses a key pair and is slow, so real systems such as TLS use asymmetric cryptography only to agree on a symmetric key, then switch to the fast cipher.

Encryption alone hides data but does not stop tampering. Modern ciphers are authenticated (AES-GCM, ChaCha20-Poly1305): they detect any change to the ciphertext. Using an unauthenticated mode, or reusing a nonce, is how real systems get broken even with a strong cipher.

The engineer's rule is Kerckhoffs's principle: the system must stay secure even if everything except the key is public. Never invent your own cipher or protocol; use a well-reviewed library and spend your effort on protecting the keys.`,
	"Hash function": `A hash function maps any input to a fixed-size output. Two very different kinds share the name. Non-cryptographic hashes (used by hash tables) only need to be fast and spread values evenly. Cryptographic hashes (SHA-256, BLAKE3) must also be one-way and collision-resistant: nobody can find an input for a given output, or two inputs with the same output.

Cryptographic hashes are everywhere: Git names every commit by a hash of its content, downloads are checked against published SHA-256 sums (this academy's releases are), and blockchains chain blocks with them.

Passwords need something else: a deliberately slow, salted hash such as Argon2, bcrypt or scrypt. A fast hash like SHA-256 lets an attacker try billions of guesses per second on a stolen password database.`,
	"Virtual memory": `Every process sees its own private address space, as if it had the machine's memory to itself. The CPU's memory management unit translates each virtual address to a physical one using page tables kept by the operating system, in pages of usually 4 KB, and the TLB caches recent translations.

This one idea gives isolation (a process cannot even name another's memory), lazy loading (pages are filled in on first touch, through a page fault), sharing (the same library's code is mapped into many processes once), copy-on-write for fork, and memory-mapped files.

When RAM runs out, the operating system can move rarely used pages to disk (swap). It keeps programs running, but a machine that swaps heavily slows to a crawl, which is why servers often prefer to kill a process rather than swap.`,
	"System call": `A system call is the only way a program can ask the kernel to do something privileged: open a file, send a packet, start a process. The program puts a number and arguments in registers and executes a special instruction (syscall on x86-64); the CPU switches to kernel mode, runs the kernel's handler, then returns.

The crossing costs much more than a normal function call, so efficient programs make fewer, bigger system calls: buffered I/O collects many small writes into one write(). On Linux, strace shows every system call a program makes, which is one of the best debugging tools there is.

System calls are also a security boundary. Sandboxes such as seccomp work by allowing a process only the few system calls it needs, so even a compromised program can do little harm.`,
	"Big-O notation": `Big O describes how the cost of an algorithm grows as the input grows, ignoring constant factors: O(n) doubles when the input doubles, O(n²) quadruples, and O(log n) barely moves. It answers "what happens at scale?", not "how fast is it today?".

Constants still matter in practice. An O(n) scan of an array often beats an O(log n) tree lookup for small n, because the array is contiguous and cache-friendly. And O describes the worst case unless stated otherwise: quicksort is O(n log n) on average but O(n²) in its worst case.

Use it to spot danger, such as a loop inside a loop over the same large list, or a database query inside a loop over rows. Then measure: profilers, not formulas, tell you where the time really goes.`,
	"TCP": `TCP turns the internet's unreliable packet delivery into a reliable, ordered stream of bytes. It numbers every byte, the receiver acknowledges what arrived, and anything not acknowledged in time is sent again. A three-way handshake (SYN, SYN-ACK, ACK) sets up the numbering before any data flows.

TCP also protects the network: flow control stops a fast sender from flooding a slow receiver, and congestion control slows everyone down when packets start getting lost, which is what keeps the internet from collapsing under load.

Its guarantees have costs. A lost packet holds up everything behind it (head-of-line blocking), and a new connection needs a round trip before data flows. QUIC, which carries HTTP/3, rebuilds these guarantees on top of UDP to avoid both.`,
	"Gradient descent": `Training a neural network means finding weights that make a loss function small. Gradient descent repeatedly computes the gradient (the direction in which the loss rises fastest) and takes a small step the other way. The learning rate sets the step size: too big and training diverges, too small and it crawls.

Real training uses stochastic gradient descent: each step uses a small random batch of examples rather than the whole dataset, which is far cheaper and whose noise even helps escape poor regions. Optimisers such as Adam adapt the step size per weight using running averages of past gradients.

Gradients come from backpropagation, the chain rule applied backwards through the network. Autograd libraries (PyTorch, JAX, micrograd) record every operation during the forward pass so they can replay it backwards automatically.`,
	"Embedding": `An embedding represents a word, image or user as a list of numbers (a vector), learned so that similar things end up close together. In a language model, every token in the vocabulary starts as a row of a big embedding matrix; training adjusts the rows until they are useful.

Distances and directions carry meaning: related words are near each other, and some relationships appear as consistent directions in the space. Search engines and recommendation systems store embeddings and find "nearest neighbours" to answer queries by meaning rather than by exact words.

Embeddings inherit whatever is in their training data, biases included. They are also not interpretable one number at a time: meaning lives in the geometry of the whole space, not in any single coordinate.`,
	"Injection": `SQL injection happens when a program builds a query by pasting user input into SQL text, so specially crafted input changes the query itself. Input such as ' OR '1'='1 can turn "find this user" into "find every user", and worse.

The fix is structural, not cosmetic: parameterised queries send the SQL and the data separately, so the database never interprets data as code. Escaping strings by hand is fragile and has failed many times; ORMs help only when you do not bypass them with raw string-built queries.

It has been a top web vulnerability for over twenty years because the root cause, mixing code and data in one string, keeps reappearing in new forms: command injection in shells, cross-site scripting in HTML, and prompt injection in AI systems.`,
	"Buffer overflow": `A buffer overflow writes past the end of an array into whatever memory comes next. In C nothing stops it: the language trusts you to stay within bounds. Functions like strcpy and gets, which do not know the destination's size, have caused countless vulnerabilities.

On the stack, an overflow can overwrite a function's saved return address, so returning jumps to code of the attacker's choosing. Defences stack up in layers: stack canaries detect the overwrite, non-executable memory stops injected code from running, and ASLR makes addresses hard to guess. Attackers answered with techniques such as return-oriented programming, which reuses existing code.

The lasting fixes are bounds checking and memory-safe languages. Major vendors report that most of their serious security bugs are memory-safety bugs, which is why Rust, Go and others are increasingly preferred for new systems code.`,
	"Compiler": `A compiler translates source code into another language, usually machine code, in stages: lexing (characters into tokens), parsing (tokens into a syntax tree), semantic analysis (types and names), optimisation on an intermediate representation, and code generation for a specific CPU.

Optimisers are why C is fast and why undefined behaviour is dangerous. The compiler is allowed to assume your program never does anything undefined, and it uses that assumption to delete checks and reorder code. Code that "worked" without optimisation can break at -O2 for exactly this reason.

Compiler Explorer (godbolt.org) shows the machine code your source becomes, line by line. Ten minutes there teaches more about what code really costs than hours of guessing.`,
	"Transaction": `A transaction groups several database changes so they succeed or fail together. The guarantees are called ACID: atomicity (all or nothing), consistency (rules hold before and after), isolation (concurrent transactions do not see each other's half-done work) and durability (once committed, it survives a crash, thanks to the write-ahead log).

Isolation comes in levels, and most databases do not default to the strictest one. Under weaker levels, anomalies such as lost updates or write skew can occur, so read your database's documentation on its default and use SELECT ... FOR UPDATE or serializable isolation when correctness depends on it.

Across several databases or services there is no single transaction. Patterns such as two-phase commit, sagas (a sequence of steps, each with an undo step) and idempotent retries fill that gap, each with trade-offs.`,
	"Model": `A neural network is a function built from many simple layers. Each layer multiplies its input by a matrix of weights, adds a bias and applies a non-linear function such as ReLU. Without that non-linearity, any stack of layers would collapse into a single matrix multiplication and could only learn straight-line relationships.

All the knowledge is in the weights, often billions of numbers, which start random and are adjusted by gradient descent to reduce a loss on training data. The architecture (convolutions for images, attention for sequences) decides which patterns are easy to learn.

Networks find statistical patterns in their training data; they do not understand in a human sense, and they can fail unpredictably on inputs unlike that data. Evaluating carefully on realistic data matters more than the size of the model.`,
}
