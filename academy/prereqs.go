package main

// Concept prerequisites: which concepts each concept builds on. They make
// "what should I learn next, and why?" answerable concept by concept. The
// next-concept suggestion sends the learner to an unmet prerequisite first
// and says which concept it unblocks, and every concept card lists what it
// builds on.
//
// Rules, checked by TestConceptPrereqs:
//   - every name is a concept in the curriculum;
//   - a prerequisite comes earlier in the default order: an earlier
//     concept of the same stage, or a concept in a stage that this stage
//     builds on (stagePrereqs, followed transitively). Following the stages
//     in order therefore never meets a missing prerequisite, and the graph
//     cannot contain a cycle.
//
// Only edges that matter are listed. Within a stage the concepts are
// already in teaching order, so most edges cross stages.

// conceptPrereqs is keyed by Concept.Name, like conceptLinks.
var conceptPrereqs = map[string][]string{
	// ---- Stage 0
	"From Source Code to Running Program":         {"The CPU at Work: Fetch, Decode, Execute", "Memory & Storage: Where Data Lives"},
	"The Operating System: The Machine's Manager": {"From Source Code to Running Program"},
	"Input & Output: Keys, Screens & Interrupts":  {"The Operating System: The Machine's Manager"},
	"Booting: From Power Button to Login":         {"The Operating System: The Machine's Manager"},

	// ---- Stage 1
	"Your First C Program: Source, Compiler, Executable": {"From Source Code to Running Program"},
	"Values, Types & Variables":                          {"Bits & Bytes: Everything Is a Number", "Your First C Program: Source, Compiler, Executable"},
	"Control Flow: Decisions & Loops":                    {"Values, Types & Variables"},
	"Functions & Decomposition":                          {"Control Flow: Decisions & Loops"},
	"Memory, Addresses & Pointers":                       {"Memory & Storage: Where Data Lives", "Functions & Decomposition"},
	"Collections: Lists, Dictionaries & Strings":         {"Memory, Addresses & Pointers"},
	"Structs: Grouping Related Data":                     {"Collections: Lists, Dictionaries & Strings"},
	"Dynamic Memory: malloc, free & Leaks":               {"Memory, Addresses & Pointers", "Structs: Grouping Related Data"},
	"Safe Input & Undefined Behavior":                    {"Collections: Lists, Dictionaries & Strings"},
	"Files & Streams: Reading and Writing Data":          {"Safe Input & Undefined Behavior"},
	"Headers, the Preprocessor & make":                   {"Functions & Decomposition", "Structs: Grouping Related Data"},
	"Debugging & Reading Errors":                         {"Functions & Decomposition"},

	// ---- Stage 2
	"Big-O: Measuring Growth":               {"Control Flow: Decisions & Loops"},
	"Recursion & Divide and Conquer":        {"Big-O: Measuring Growth", "Functions & Decomposition"},
	"Arrays, Linked Lists, Stacks & Queues": {"Big-O: Measuring Growth", "Collections: Lists, Dictionaries & Strings", "Dynamic Memory: malloc, free & Leaks"},
	"Searching: Linear & Binary Search":     {"Arrays, Linked Lists, Stacks & Queues", "Recursion & Divide and Conquer"},
	"Hash Tables":                           {"Arrays, Linked Lists, Stacks & Queues"},
	"Trees, Heaps & Sorting":                {"Arrays, Linked Lists, Stacks & Queues", "Recursion & Divide and Conquer", "Searching: Linear & Binary Search"},
	"Graphs, BFS/DFS & Shortest Paths":      {"Arrays, Linked Lists, Stacks & Queues", "Trees, Heaps & Sorting"},
	"Greedy Algorithms":                     {"Trees, Heaps & Sorting", "Graphs, BFS/DFS & Shortest Paths"},
	"Dynamic Programming":                   {"Recursion & Divide and Conquer", "Greedy Algorithms"},

	// ---- Stage 3
	"Logic & Boolean Algebra":              {"Control Flow: Decisions & Loops"},
	"Sets, Functions & Relations":          {"Logic & Boolean Algebra"},
	"Proof & Induction":                    {"Logic & Boolean Algebra", "Control Flow: Decisions & Loops"},
	"Counting & Combinatorics":             {"Sets, Functions & Relations"},
	"Modular Arithmetic & Number Theory":   {"Proof & Induction", "Values, Types & Variables"},
	"Recurrences: The Cost of Recursion":   {"Proof & Induction", "Counting & Combinatorics", "Functions & Decomposition"},
	"Graph Theory: Degrees, Trees & Paths": {"Sets, Functions & Relations", "Proof & Induction"},
	"Probability for Programmers":          {"Counting & Combinatorics"},

	// ---- Stage 4
	"Binary & Number Representation":               {"Bits & Bytes: Everything Is a Number"},
	"Logic Gates & Circuits":                       {"Logic & Boolean Algebra", "Binary & Number Representation"},
	"Building an ALU: Adders, Subtraction & Flags": {"Logic Gates & Circuits", "Binary & Number Representation"},
	"Memory: Latches, Registers & RAM":             {"Logic Gates & Circuits"},
	"Finite State Machines & Sequential Logic":     {"Memory: Latches, Registers & RAM", "Logic & Boolean Algebra"},
	"The CPU: Datapath & Instruction Cycle":        {"Memory: Latches, Registers & RAM", "The CPU at Work: Fetch, Decode, Execute", "Building an ALU: Adders, Subtraction & Flags", "Finite State Machines & Sequential Logic"},
	"Instruction Sets & Machine Code":              {"The CPU: Datapath & Instruction Cycle", "Binary & Number Representation"},
	"Pipelining, Branch Prediction & Parallelism":  {"The CPU: Datapath & Instruction Cycle"},

	// ---- Stage 5
	"The Compilation Pipeline":                   {"Headers, the Preprocessor & make", "From Source Code to Running Program"},
	"Registers, the Stack & Calling Conventions": {"Functions & Decomposition", "The CPU: Datapath & Instruction Cycle"},
	"Integers & Floating Point in Bits":          {"Binary & Number Representation", "Values, Types & Variables"},
	"The Memory Hierarchy & Locality":            {"Memory & Storage: Where Data Lives", "Memory: Latches, Registers & RAM"},
	"Linking, Loading & Shared Libraries":        {"The Compilation Pipeline", "Headers, the Preprocessor & make"},
	"What the Optimizer Does":                    {"The Compilation Pipeline", "Integers & Floating Point in Bits", "Safe Input & Undefined Behavior"},
	"Profiling & Measuring Performance":          {"The Memory Hierarchy & Locality", "Big-O: Measuring Growth"},
	"Parsing: Recursive Descent & Pratt":         {"The Compilation Pipeline", "Trees, Heaps & Sorting"},

	// ---- Stage 6
	"Processes & the Address Space":       {"The Operating System: The Machine's Manager", "The Compilation Pipeline"},
	"Virtual Memory & Paging":             {"Processes & the Address Space", "Memory, Addresses & Pointers", "The Memory Hierarchy & Locality"},
	"Memory Allocators: How malloc Works": {"Virtual Memory & Paging", "Dynamic Memory: malloc, free & Leaks"},
	"System Calls & Privilege":            {"The Operating System: The Machine's Manager", "Registers, the Stack & Calling Conventions"},
	"File Descriptors, Files & Pipes":     {"System Calls & Privilege", "Processes & the Address Space", "Files & Streams: Reading and Writing Data"},
	"Signals":                             {"Processes & the Address Space", "File Descriptors, Files & Pipes"},
	"Scheduling & Context Switches":       {"Processes & the Address Space", "Input & Output: Keys, Screens & Interrupts"},
	"Synchronisation & Memory Ordering":   {"Scheduling & Context Switches"},

	// ---- Stage 7
	"Durability: Page Cache, fsync & Atomic Rename": {"System Calls & Privilege", "Files & Streams: Reading and Writing Data"},
	"Pages, Records & the Buffer Pool":              {"Durability: Page Cache, fsync & Atomic Rename", "The Memory Hierarchy & Locality", "Virtual Memory & Paging"},
	"B-Trees":                                       {"Trees, Heaps & Sorting", "The Memory Hierarchy & Locality"},
	"LSM-Trees":                                     {"B-Trees", "Durability: Page Cache, fsync & Atomic Rename"},
	"Write-Ahead Logging & Recovery":                {"Durability: Page Cache, fsync & Atomic Rename"},
	"Transactions, Isolation & MVCC":                {"Write-Ahead Logging & Recovery", "Synchronisation & Memory Ordering"},
	"SQL & the Relational Model":                    {"Sets, Functions & Relations", "B-Trees"},
	"Query Planning & Join Algorithms":              {"SQL & the Relational Model", "B-Trees", "Hash Tables", "Big-O: Measuring Growth"},

	// ---- Stage 8
	"The Layered Stack & Encapsulation":         {"Networks: How Computers Talk"},
	"IP Addressing, Routing & NAT":              {"The Layered Stack & Encapsulation", "Binary & Number Representation"},
	"TCP: Handshake, Flow & Congestion Control": {"The Layered Stack & Encapsulation"},
	"Sockets & I/O Multiplexing":                {"TCP: Handshake, Flow & Congestion Control", "System Calls & Privilege"},
	"Time, Clocks & Ordering":                   {"The Layered Stack & Encapsulation"},
	"Consensus & Replication (Raft)":            {"Time, Clocks & Ordering"},
	"Partitioning & Consistent Hashing":         {"Consensus & Replication (Raft)", "Hash Tables"},
	"HTTP, DNS & How the Web Works":             {"TCP: Handshake, Flow & Congestion Control"},
	"Timeouts, Retries & Idempotency":           {"HTTP, DNS & How the Web Works", "Sockets & I/O Multiplexing"},

	// ---- Stage 9
	"The Command Line & Your Environment":           {"The Operating System: The Machine's Manager"},
	"Shell Scripting & Automation":                  {"The Command Line & Your Environment"},
	"Version Control with Git":                      {"The Command Line & Your Environment"},
	"Build Systems & Dependencies":                  {"Version Control with Git", "Headers, the Preprocessor & make"},
	"Testing & Reliability":                         {"Functions & Decomposition"},
	"Debugging Tools: gdb, Core Dumps & git bisect": {"Version Control with Git", "Testing & Reliability", "Debugging & Reading Errors"},
	"Software Design & Abstraction":                 {"Functions & Decomposition", "Testing & Reliability"},
	"From Code to Production":                       {"Version Control with Git", "Testing & Reliability"},

	// ---- Stage 10
	"Web Vulnerabilities: Injection & XSS":            {"Thinking Like an Attacker: Threat Modelling", "HTTP, DNS & How the Web Works"},
	"Hashing & Password Storage":                      {"Hash Tables"},
	"Encryption: Symmetric & Public-Key":              {"Counting & Combinatorics"},
	"TLS & Authentication on the Web":                 {"Encryption: Symmetric & Public-Key", "HTTP, DNS & How the Web Works"},
	"Memory-Safety Bugs: Overflows & Use-After-Free":  {"Safe Input & Undefined Behavior", "Dynamic Memory: malloc, free & Leaks", "Registers, the Stack & Calling Conventions"},
	"Access Control: Who May Do What":                 {"TLS & Authentication on the Web"},
	"Network Attacks & Defences":                      {"TCP: Handshake, Flow & Congestion Control", "TLS & Authentication on the Web"},
	"Secrets & the Software Supply Chain":             {"Thinking Like an Attacker: Threat Modelling", "Encryption: Symmetric & Public-Key"},
	"Detection & Incident Response":                   {"Thinking Like an Attacker: Threat Modelling"},
	"Reading Research Papers":                         {"Big-O: Measuring Growth"},
	"Reading CVEs, CWEs & Advisories":                 {"Thinking Like an Attacker: Threat Modelling", "Web Vulnerabilities: Injection & XSS"},
	"Vulnerability Research & Coordinated Disclosure": {"Memory-Safety Bugs: Overflows & Use-After-Free", "Reading CVEs, CWEs & Advisories"},

	// ---- Stage 11
	"Finite Automata & Regular Expressions":        {"Sets, Functions & Relations"},
	"NFAs, DFAs & How Regex Engines Work":          {"Finite Automata & Regular Expressions", "Sets, Functions & Relations"},
	"Grammars & Pushdown Automata":                 {"Finite Automata & Regular Expressions", "Arrays, Linked Lists, Stacks & Queues"},
	"Turing Machines & Computability":              {"Grammars & Pushdown Automata"},
	"The Halting Problem & Undecidability":         {"Turing Machines & Computability", "Proof & Induction"},
	"Complexity Classes: P, NP & Hard Problems":    {"Turing Machines & Computability", "Big-O: Measuring Growth"},
	"Reductions & NP-Completeness":                 {"Complexity Classes: P, NP & Hard Problems", "Logic & Boolean Algebra"},
	"Approximation & Heuristics for Hard Problems": {"Complexity Classes: P, NP & Hard Problems", "Greedy Algorithms"},

	// ---- Stage 12
	"Paradigms: Imperative, Object-Oriented & Functional": {"Functions & Decomposition"},
	"Type Systems": {"Values, Types & Variables"},
	"Recursion, Closures & Higher-Order Functions":      {"Paradigms: Imperative, Object-Oriented & Functional", "Recursion & Divide and Conquer"},
	"Generics & Polymorphism":                           {"Type Systems", "Trees, Heaps & Sorting"},
	"Error Handling: Exceptions, Results & Error Codes": {"Type Systems", "Debugging & Reading Errors"},
	"Memory Management":                                 {"Dynamic Memory: malloc, free & Leaks", "Safe Input & Undefined Behavior"},
	"Interpreters, Compilers & Runtimes":                {"The Compilation Pipeline"},
	"Concurrency Models":                                {"Paradigms: Imperative, Object-Oriented & Functional"},

	// ---- Stage 13
	"Vectors, Dot Products & Similarity":           {"Sets, Functions & Relations"},
	"Matrices as Linear Maps":                      {"Sets, Functions & Relations", "Vectors, Dot Products & Similarity"},
	"Linear Systems, Least Squares & Conditioning": {"Matrices as Linear Maps"},
	"Floating Point for Numerical Computing":       {"Bits & Bytes: Everything Is a Number", "Values, Types & Variables"},
	"Eigenvectors & the SVD":                       {"Matrices as Linear Maps"},
	"Gradients, Jacobians & the Matrix Chain Rule": {"Matrices as Linear Maps"},
	"Probability, Softmax & Cross-Entropy":         {"Probability for Programmers", "Floating Point for Numerical Computing"},
	"Optimisation: From Gradient Descent to Adam":  {"Gradients, Jacobians & the Matrix Chain Rule"},

	// ---- Stage 14
	"Descriptive Statistics & Distributions": {"Probability for Programmers"},
	"Bayes' Theorem & Base Rates":            {"Probability for Programmers", "Descriptive Statistics & Distributions"},
	"Statistical Inference & Experiments":    {"Descriptive Statistics & Distributions"},
	"The Bootstrap & Resampling":             {"Statistical Inference & Experiments"},
	"Bias, Variance & Regularisation":        {"Linear & Logistic Regression", "Linear Systems, Least Squares & Conditioning"},
	"Linear & Logistic Regression":           {"Descriptive Statistics & Distributions", "Optimisation: From Gradient Descent to Adam", "Probability, Softmax & Cross-Entropy"},
	"Trees, Ensembles & Clustering":          {"Linear & Logistic Regression"},
	"Evaluating Models Honestly":             {"Linear & Logistic Regression", "Bayes' Theorem & Base Rates"},

	// ---- Stage 15
	"Computational Graphs & Reverse-Mode Autodiff": {"Gradients, Jacobians & the Matrix Chain Rule"},
	"MLPs, Activations & Initialisation":           {"Computational Graphs & Reverse-Mode Autodiff", "Matrices as Linear Maps"},
	"Residual Connections & Normalisation":         {"MLPs, Activations & Initialisation"},
	"Convolutions & Weight Sharing":                {"MLPs, Activations & Initialisation", "Matrices as Linear Maps"},
	"Tokenisation & Embeddings":                    {"Vectors, Dot Products & Similarity", "Collections: Lists, Dictionaries & Strings"},
	"Attention & the Transformer":                  {"Residual Connections & Normalisation", "Probability, Softmax & Cross-Entropy", "Tokenisation & Embeddings"},
	"Sampling & Decoding":                          {"Attention & the Transformer", "Probability, Softmax & Cross-Entropy"},
	"The Training Loop & Generalisation":           {"MLPs, Activations & Initialisation", "Evaluating Models Honestly", "Optimisation: From Gradient Descent to Adam"},

	// ---- Stage 16
	"The GPU Execution Model":                          {"Pipelining, Branch Prediction & Parallelism", "Scheduling & Context Switches"},
	"GPU Memory Hierarchy & Coalescing":                {"The GPU Execution Model", "The Memory Hierarchy & Locality"},
	"The Roofline Model: Compute- vs Memory-Bound":     {"GPU Memory Hierarchy & Coalescing"},
	"Kernel Fusion & FlashAttention":                   {"The Roofline Model: Compute- vs Memory-Bound", "Attention & the Transformer"},
	"KV Cache, Batching & PagedAttention":              {"Attention & the Transformer", "Virtual Memory & Paging"},
	"Quantisation & Number Formats":                    {"The Roofline Model: Compute- vs Memory-Bound", "Floating Point for Numerical Computing"},
	"Parallelism Across GPUs: Data, Tensor & Pipeline": {"The Training Loop & Generalisation", "The GPU Execution Model"},
	"Profiling & Benchmarking Honestly":                {"The Roofline Model: Compute- vs Memory-Bound", "Profiling & Measuring Performance"},

	// ---- Stage 17
	"Secure Design Principles":                              {"Thinking Like an Attacker: Threat Modelling", "Access Control: Who May Do What"},
	"Handling Untrusted Input":                              {"Web Vulnerabilities: Injection & XSS"},
	"Writing Memory-Safe Code":                              {"Memory-Safety Bugs: Overflows & Use-After-Free"},
	"Security Code Review":                                  {"Handling Untrusted Input", "Writing Memory-Safe Code"},
	"Automated Security Testing":                            {"Writing Memory-Safe Code"},
	"Safe Parsing: Lengths, Limits & Binary Formats":        {"Handling Untrusted Input", "Memory-Safety Bugs: Overflows & Use-After-Free", "Bits & Bytes: Everything Is a Number"},
	"Finding Secrets: Patterns, Entropy & Git History":      {"Secrets & the Software Supply Chain", "Probability for Programmers"},
	"Integrity Checks: Hashes, Checksums & Signed Releases": {"Hashing & Password Storage", "Secrets & the Software Supply Chain"},

	// ---- Stage 18
	"Strong Authentication: MFA & Passkeys":                     {"TLS & Authentication on the Web", "Encryption: Symmetric & Public-Key"},
	"Sessions, Tokens & Single Sign-On":                         {"Strong Authentication: MFA & Passkeys", "HTTP, DNS & How the Web Works"},
	"Authorization Design: Roles, Attributes & Least Privilege": {"Access Control: Who May Do What"},
	"Cloud Security & Shared Responsibility":                    {"Authorization Design: Roles, Attributes & Least Privilege"},
	"Zero Trust: Never Trust the Network Alone":                 {"Strong Authentication: MFA & Passkeys", "Network Attacks & Defences"},
	"Rate Limiting & Credential Stuffing Defences":              {"Strong Authentication: MFA & Passkeys", "Hashing & Password Storage"},
	"Policy as Code: Writing & Testing Authorization Rules":     {"Authorization Design: Roles, Attributes & Least Privilege"},
	"Secrets Management & Short-Lived Credentials":              {"Cloud Security & Shared Responsibility", "Secrets & the Software Supply Chain"},

	// ---- Stage 19
	"Firewalls & Network Segmentation":            {"Network Attacks & Defences"},
	"Logging & Telemetry":                         {"Detection & Incident Response"},
	"Intrusion Detection: Signatures & Anomalies": {"Logging & Telemetry", "Probability for Programmers"},
	"Detection Engineering with MITRE ATT&CK":     {"Intrusion Detection: Signatures & Anomalies"},
	"Security Monitoring & Alert Triage":          {"Intrusion Detection: Signatures & Anomalies"},
	"Network Flows & Packet Capture":              {"Firewalls & Network Segmentation", "TCP: Handshake, Flow & Congestion Control"},
	"Asset Inventory & Attack Surface":            {"Firewalls & Network Segmentation", "Sockets & I/O Multiplexing"},
	"Tamper-Evident Logs & Trustworthy Time":      {"Logging & Telemetry", "Hashing & Password Storage", "Time, Clocks & Ordering"},

	// ---- Stage 20
	"Incident Response Planning":                             {"Detection & Incident Response"},
	"Containment, Eradication & Recovery":                    {"Incident Response Planning"},
	"Digital Forensics Basics":                               {"Incident Response Planning", "Logging & Telemetry"},
	"Blameless Post-Mortems":                                 {"Incident Response Planning"},
	"Privacy & Data Protection":                              {"Thinking Like an Attacker: Threat Modelling"},
	"Backups & Recovery Testing":                             {"Containment, Eradication & Recovery", "Hashing & Password Storage"},
	"Building an Incident Timeline":                          {"Digital Forensics Basics", "Tamper-Evident Logs & Trustworthy Time"},
	"Anonymisation, Pseudonymisation & Differential Privacy": {"Privacy & Data Protection", "Probability for Programmers"},
}

// conceptAt locates a concept by name: its stage and its index there.
func conceptAt(name string) (stageID, idx int, ok bool) {
	for id, g := range curriculum {
		for i, c := range g.Concepts {
			if c.Name == name {
				return id, i, true
			}
		}
	}
	return 0, 0, false
}

// Prereq is one prerequisite of a concept and whether it is met.
type Prereq struct {
	Name  string
	Stage int
	Met   bool // marked as understood
}

// prereqsOf lists a concept's prerequisites with their state.
func (r *Registry) prereqsOf(name string) []Prereq {
	var out []Prereq
	for _, p := range conceptPrereqs[name] {
		id, _, ok := conceptAt(p)
		if !ok {
			continue
		}
		met := true // a stage missing from the registry cannot block anything
		if _, s := r.findStage(id); s != nil {
			met = s.hasStudied(p)
		}
		out = append(out, Prereq{p, id, met})
	}
	return out
}

// firstUnmet returns the concept to learn before name: its earliest unmet
// prerequisite, followed down to one whose own prerequisites are all met.
// It returns "" when every prerequisite of name is met. Prerequisites
// always come earlier in the curriculum, so the walk ends.
func (r *Registry) firstUnmet(name string) string {
	for _, p := range r.prereqsOf(name) {
		if !p.Met {
			if deeper := r.firstUnmet(p.Name); deeper != "" {
				return deeper
			}
			return p.Name
		}
	}
	return ""
}

// nextToLearn is the next concept to learn: the first one not yet
// understood in the current stage, or, when that one builds on concepts
// not yet understood, the first of those. unblocks names the concept the
// detour leads back to ("" when there is no detour).
func (r *Registry) nextToLearn() (stage, concept int, unblocks string, ok bool) {
	s := r.currentStage()
	if s == nil {
		return 0, 0, "", false
	}
	g, found := guideFor(s.ID)
	if !found {
		return 0, 0, "", false
	}
	for i, c := range g.Concepts {
		if s.hasStudied(c.Name) {
			continue
		}
		if p := r.firstUnmet(c.Name); p != "" {
			if id, idx, ok := conceptAt(p); ok {
				return id, idx, c.Name, true
			}
		}
		return s.ID, i, "", true
	}
	return 0, 0, "", false
}
