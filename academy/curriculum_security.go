package main

// Track D, Defensive Security: building systems that resist attack, and
// noticing and recovering when an attack happens anyway. Every exercise
// is about protecting systems you own; offensive practice stays on the
// legal training platforms named in Stage 10.

var securityGuides = map[int]StageGuide{
	// -----------------------------------------------------------------
	17: {
		Overview: `Most security holes are ordinary bugs in ordinary code. This stage is
about writing code that stays safe when its input is hostile, and about
catching problems before they ship: design principles that make whole
classes of bugs impossible, careful handling of everything that comes
from outside, parsers that never trust a length, memory-safe habits,
keeping secrets out of code, verifying what you download, reviewing code
with an attacker's eye, and automated tools that look for trouble on
every change. Everything here was measured on programs written for the
purpose: a 17-byte name that silently made a user an administrator, a
blocklist that let 6 of 11 path escapes through, and a fuzzer that ran
106 million inputs without getting past a 4-byte header.`,
		Outcomes: []string{
			"Apply classic secure-design principles such as least privilege and fail-safe defaults",
			"Handle untrusted input safely: canonicalise, validate against an allowlist, encode on output",
			"Write C that is hard to exploit, and know when to choose a memory-safe language",
			"Review a change for security problems with a repeatable checklist",
			"Add static analysis, fuzzing and dependency scanning to a CI pipeline",
			"Parse binary formats without trusting their length fields, with limits on size and depth",
			"Find secrets in code and in history, and respond correctly when one leaks",
			"Verify downloads with checksums and signatures, and implement SHA-256 in C",
		},
		Glossary: []Term{
			{"Secure by default", "The safe setting is the one you get without doing anything."},
			{"Fail closed", "When something goes wrong, deny access rather than allow it."},
			{"Trust boundary", "A place where data passes from something you do not control to something you do."},
			{"Allowlist", "A list of what is permitted; everything else is refused."},
			{"Canonicalisation", "Turning input into its one standard form (a clean path, normalised text) before checking it."},
			{"Static analysis", "Tools that read source code to find bugs without running it (SAST)."},
			{"Fuzzing", "Feeding a program huge numbers of generated inputs to find crashes."},
			{"Code review", "Another engineer reading a change before it is merged."},
			{"umask", "The permission bits a Unix process removes from every file it creates; 022 makes new files readable by everyone."},
			{"ReDoS", "Regular-expression denial of service: an input that makes a backtracking regex take exponential time."},
			{"Type-length-value (TLV)", "A binary format in which each record is a type, a length, and that many bytes of value."},
			{"Stack canary", "A secret value placed before a function's return address and checked before it returns, to detect stack overflows."},
			{"FORTIFY_SOURCE", "A compiler and C library feature that replaces calls such as strcpy and memcpy with versions that check the destination's size."},
			{"Sanitizer", "A compiler option (AddressSanitizer, UBSan) that adds checks to a program so memory and undefined-behaviour bugs stop it with a report."},
			{"Secret scanning", "Searching code and its history for credentials, by their known formats and by how random they look."},
			{"Key rotation", "Replacing a credential with a new one and revoking the old, done at once when a secret leaks."},
			{"Checksum", "A short fingerprint of a file, such as its SHA-256 hash, published so downloads can be checked."},
			{"Digital signature", "Proof made with a private key that a file came from its holder and was not changed; anyone can check it with the public key."},
			{"Seed corpus", "Valid example inputs a fuzzer starts from, so it reaches the code behind headers and magic numbers."},
		},
		Concepts: []Concept{
			{
				Name:    "Secure Design Principles",
				Summary: "Fifty-year-old rules that still prevent most disasters.",
				Body: `In 1975 Jerome Saltzer and Michael Schroeder listed design principles
for protecting information, and they still hold. Economy of mechanism:
keep security code small enough to check. Fail-safe defaults: base
access on explicit permission, so a mistake denies rather than allows.
Complete mediation: check every access, every time, not just the first.
Open design: security must not depend on the design being secret, only
on keys. Separation of privilege: sensitive actions need two
independent conditions (two keys, two people). Least privilege: every
program and user gets only the rights it needs. Psychological
acceptability: if the secure way is painful, people route around it.

Modern practice adds secure defaults: ship with the safe setting on,
so that someone who never reads the manual is still protected.`,
				UnderTheHood: `Defaults, measured in C on Linux. With the usual umask of 022, the
permissions each way of creating a file gives:

  open(path, O_CREAT, 0666)     0644   readable by every user
  fopen(path, "w")              0644   the same
  open(path, O_CREAT, 0600)     0600   the owner only
  mkstemp(template)             0600   the owner only

fopen has no way to ask for less: a program that saves a token with
fopen leaves it readable by every account on the machine, unless it
sets the umask first or creates the file with open and mode 0600. The
system itself follows least privilege: /etc/passwd, which every
program reads, is 0644, while /etc/shadow, holding the password
hashes, is 0640, readable only by root and the shadow group.

One more honest measurement: the container these lessons were measured
in runs everything as root (uid 0), and every one of those permission
checks succeeds for root. Least privilege starts with not running as the
most privileged user.`,
				MentalModel: "Make the safe path the easy path, and make mistakes fail closed.",
				TryIt:       "Pick one app you use daily and list its default settings for sharing, passwords and updates. Which defaults protect you, and which would you have to know to change?",
				Analogy: `A well-designed building: fire doors that close by themselves when the
power fails (fail-safe default), a single staffed entrance where every
visitor is checked (complete mediation), and a cleaner's key that opens
offices but not the safe (least privilege).`,
				Example: `Log4Shell (2021) came from a logging library that, by default, looked up
and fetched addresses found inside log messages, so merely logging a
visitor's text could make a server load and run remote code. Log4j
2.16.0 fixed it by turning that feature off by default. In 2023 Amazon
S3 began blocking public access on all new storage buckets by default,
after years of data leaks from buckets made public by mistake.`,
				Exercises: trio(
					"Match each to a principle: (a) a building needs two keycards for the server room; (b) a firewall that blocks everything not listed; (c) an admin panel that checks your role only at login; (d) a cipher whose algorithm is published.",
					"(c) breaks complete mediation: roles can change during a session.",
					"In C, write save_token(path, token) that creates the file so only its owner can read it, whatever the umask, and refuses to follow a symbolic link or overwrite an existing file. Check the result with stat and print the mode.",
					"open(path, O_WRONLY | O_CREAT | O_EXCL | O_NOFOLLOW, 0600): O_EXCL refuses an existing file.",
					"Audit the defaults of a service you run or could run (a database, a web server, a router): bind address, default accounts and passwords, logging, update policy. Write down what you would change before exposing it to a network, and why.",
					"Many databases once listened on all network interfaces with no password by default.",
				),
			},
			{
				Name:    "Handling Untrusted Input",
				Summary: "Canonicalise, then check against an allowlist, then encode on the way out.",
				Body: `Everything that crosses a trust boundary is untrusted: request fields,
file names, uploaded files, environment variables, even the output of
another service or an AI model. Handle it in a fixed order.

First canonicalise: turn it into its one standard form, because
attackers use alternative spellings (../, %2e%2e, Unicode look-alikes)
to slip past checks. Then validate against an allowlist of what is
expected (type, length, range, format) and reject everything else;
blocklists of known-bad values always miss something. Better still,
parse input into a typed value (a date, a user ID, an enum) so the rest
of the program never sees raw text. Finally, encode data for wherever it
goes next: parameters for SQL, escaping for HTML, argument lists (never
a shell string) for commands. Put size limits on everything.`,
				Diagram: `outside (requests, files, env, other services)
   │
═══╪═══════════════ trust boundary ═══════════════
   ▼
canonicalise ─▶ allowlist check ─▶ parse to types ─▶ your code
                                                        │
where it goes next ◀─ encode: SQL params, HTML escape ◀─┘`,
				UnderTheHood: `Seventeen file names sent to a server folder, /srv/files, that decodes
%xx once and treats \ as / (as Windows does). Eleven of them really
escape the folder. Two ways of checking them:

                                  escapes let in   good names refused
  blocklist: no "../", no "/"         6 of 11             1
  decode, normalise, then check       0 of 11             0

The blocklist let ..%2f, %2e%2e/, %2e%2e%2f, .%2e/, ..\ and %2f through,
and still refused the harmless a/../b.txt. Checking the canonical form
needs no list of bad spellings at all.

Size limits matter as much. The regular expression ^(a+)+$ against n
a's followed by a !, in Python:

  n        16       20       24       26
  time    3 ms    48 ms    0.94 s    3.2 s

Every 2 more characters made it 4 times slower, because the engine
tries every way of splitting the a's between the two +s; ^a+$ answered
each in microseconds. Forty characters would take hours. Limit lengths,
avoid nested repetition, or use an engine that guarantees linear time,
such as Go's regexp or RE2.`,
				MentalModel: "Check the canonical form, not the spelling the attacker chose.",
				TryIt:       "Type in and run this stage's safe_join program, then add a test name of your own that you expect to be refused.",
				Analogy: `A post room that X-rays every parcel. First it unwraps the parcel to see
what is really inside (canonicalise), then compares it with the list of
things the company ordered (allowlist), and only then passes it on,
relabelled for the right department (encode).`,
				Example: `In 2021 a change to path handling in Apache HTTP Server 2.4.49
(CVE-2021-41773) let requests with encoded ../ sequences escape the web
folder and read other files on the server. The first fix, 2.4.50, was
incomplete (CVE-2021-42013), and a full fix arrived in 2.4.51. In 2019 a
single regular expression with nested repetition, deployed in a firewall
rule, took Cloudflare's network down for 27 minutes by using all of its
CPU. This academy's own resource importer refuses control codes and
bidirectional text markers, which could otherwise disguise what you are
importing.`,
				Exercises: trio(
					`Why must a program normalise "files/a/../../etc/passwd" before checking that it starts with "files/"?`,
					"The unnormalised string does start with files/, yet it points outside.",
					"Write safe_join(base, name) in C with tests for: a normal name, a nested name, ../ escapes, an absolute path, an empty name and a very long name. Then extend it to refuse symbolic links that point outside the folder.",
					"Resolve the real path (realpath) after joining, then compare it with the real path of the base folder.",
					"Choose one program you wrote that accepts input (a form, a CLI, a file importer). Draw its trust boundaries, list every input, and for each write the allowlist rule, the size limit and the output encoding. Fix at least one gap you find.",
					"Include the inputs people forget: environment variables, file names inside archives, headers.",
				),
			},
			{
				Name:    "Safe Parsing: Lengths, Limits & Binary Formats",
				Summary: "A length field is a claim made by the attacker: check it against what you actually received.",
				Body: `Binary formats (network protocols, images, archives, fonts) describe
themselves: a header says how long the next part is, how many items
follow, how deep the nesting goes. Every one of those numbers comes from
the sender, so a parser that believes them reads or writes wherever the
sender chooses.

The rules are few. Compare every length with the bytes that actually
remain before using it, and do the arithmetic so it cannot overflow
(check len > remaining, not offset + len > size, which can wrap round).
Read multi-byte numbers from unsigned bytes. Put limits on everything
the input controls: total size, number of records, nesting depth (a
recursive parser crashes on deep enough input). Validate the whole input
before acting on any of it, so a half-valid file never leaves half its
effects. And keep parsers small, separate and fuzzed, or write them in a
memory-safe language.`,
				Diagram: `record:  [type 1 byte][length 2 bytes][value: length bytes]
          ▲              ▲
          │              └ a claim: refuse if length > bytes remaining
          └ unknown type: skip it, using the same checked length`,
				UnderTheHood: `A heartbeat in the style of TLS, in C: the request is a two-byte length
and a payload, and the reply echoes the payload back. A request that
claims 200 bytes but carries 4 ("ping"):

  memcpy(reply, payload, claimed)       200 bytes back, 196 never sent,
                                        a password stored nearby among them
  if (claimed > received - 2) refuse    nothing sent back
  the same bug under AddressSanitizer   heap-buffer-overflow: READ of
                                        size 200, and the program stops

That is Heartbleed (2014) in fifteen lines: OpenSSL returned up to 64 KB
of its memory per request, private keys included, because of one missing
comparison.

Depth is a length too. A recursive C parser for nested brackets [[[…]]],
with Linux's default 8 MB stack: depth 10,000 parsed; depth 100,000, just
200 KB of input, crashed it with a segmentation fault. Python's json
module refused depth 1,000 with a RecursionError instead of crashing:
a limit, chosen in advance, is the difference.`,
				MentalModel: "Every number in the input is the attacker's claim; check it against what you received before you use it.",
				TryIt:       "Open any file in a hex viewer (xxd file.png | head) and find a length field in its header: for PNG, the 4 bytes before IHDR.",
				Analogy: `A delivery note that says "40 boxes" on a van holding 4. A careful
warehouse counts the boxes against the note before signing; a careless
one walks into the neighbour's storeroom to look for the other 36.`,
				Example: `Heartbleed (CVE-2014-0160) leaked memory from about a sixth of the
web's secure servers through an unchecked length. In 2023 a heap
overflow in libwebp's handling of lossless images (CVE-2023-4863), a
table built from attacker-chosen code lengths, was exploited through
images on the web and patched in every major browser. Both parsers were
written in C, and both bugs came down to a number from the input that
was trusted.`,
				Exercises: trio(
					"A record header says its value is 65,535 bytes long, and 10 bytes are left in the input. List three things a parser could do here, and say which one is right.",
					"Refuse the whole input: reading past the end, or trusting the rest, are both wrong.",
					"In C, write a parser for type-length-value records with a two-byte length, which refuses lengths past the end of the input and nesting deeper than 8, and validates everything before printing anything. Run it under AddressSanitizer on truncated and oversized inputs.",
					"Keep an offset and the bytes remaining; check length <= remaining before advancing.",
					"Fuzz a real file parser (for example a small PNG or WAV header reader you write, or an open-source one) with libFuzzer or go test -fuzz, starting from a few real files, and fix every crash it finds.",
					"Seed it with valid files, so the fuzzer gets past the header at once.",
				),
			},
			{
				Name:    "Writing Memory-Safe Code",
				Summary: "Prevent overflow bugs by language choice, habits and compiler help.",
				Body: `The strongest defence against memory-safety bugs is a language that
cannot have them: Rust, Go, Java, Python and C# check bounds and manage
lifetimes for you. Much new systems code is written in them for exactly
this reason.

When you do write C, make the machine help. Compile with warnings as
errors (-Wall -Wextra -Werror). Turn on the hardening that modern
distributions use: -D_FORTIFY_SOURCE (checked versions of risky
functions), -fstack-protector-strong (canaries), position-independent
executables for ASLR, and full RELRO. Use functions that take a size
(snprintf, fgets, strlcpy), never gets or sprintf. Pass lengths alongside
pointers, decide who frees each allocation, and set pointers to NULL
after freeing. Run every test under AddressSanitizer and
UndefinedBehaviorSanitizer, and fuzz anything that parses input.`,
				Diagram: `strongest ▲ memory-safe language: bounds and lifetimes checked
          │ safe C habits: sized APIs, one owner, NULL after free
          │ sanitizers and fuzzing while testing
weakest   │ hardening: canaries, FORTIFY, ASLR, RELRO`,
				UnderTheHood: `One C program with two classic overflows: strcpy of a user's name into
char name[16] inside struct { char name[16]; int is_admin; }, and into
a local char buf[16]. Built four ways, with gcc:

  built with                     into the struct      past the local
                                 (17 bytes)           buffer (40 bytes)
  no protection                  is_admin became 1    segfault
  -fstack-protector-strong       is_admin became 1    "stack smashing
                                                      detected", abort
  -D_FORTIFY_SOURCE=2            "buffer overflow     "buffer overflow
                                 detected", abort     detected", abort
  -fsanitize=address             is_admin became 1    stack-buffer-overflow
                                                      reported

The silent one is the dangerous one: 17 bytes made an ordinary user an
administrator, and neither the canary nor AddressSanitizer noticed,
because the write never left the struct. Only FORTIFY, which knew at
compile time that name holds 16 bytes, stopped it.

What they cost, best of three runs:

                               call-heavy loop   byte-pair encoding
  no protection                    1.16 s             0.123 s
  stack protector                  1.23 s             0.125 s
  stack protector + FORTIFY        1.17 s             0.136 s
  AddressSanitizer + UBSan         3.58 s             0.460 s

Hardening costs a few percent, so ship it; the sanitizers cost 3 to 4
times, so test with them, always.`,
				MentalModel: "Hardening makes exploits harder; only removing the bug makes you safe.",
				TryIt:       "Compile a small C program with gcc -Wall -Wextra -O2 -D_FORTIFY_SOURCE=2 -fstack-protector-strong, and check with checksec or readelf which protections the binary has.",
				Analogy: `Seatbelts, airbags and crumple zones (hardening) save lives, but the
best crash is the one prevented by a road with no sharp bend (a
memory-safe language). Careful driving (safe habits) matters on every
road.`,
				Example: `Google reported in 2024 that memory-safety bugs fell from 76% of
Android's vulnerabilities in 2019 to 24% in 2024, largely because new
code was written in memory-safe languages such as Rust and Kotlin while
old C and C++ code was left in place. In 2023 the US cybersecurity
agency CISA and partner agencies published "The Case for Memory Safe
Roadmaps", urging vendors to plan the same move.`,
				Exercises: trio(
					"Rank these from most to least effective at stopping buffer overflows: stack canaries, rewriting the parser in Rust, AddressSanitizer in tests, ASLR. Explain the order.",
					"Only one of them removes the bug; the others detect it or make it harder to exploit.",
					"Take a C program you wrote in Stage 1 and harden it: compile with warnings as errors and the hardening flags, run its tests under -fsanitize=address,undefined, and fix every warning and report.",
					"Treat each sanitizer report as a real bug, even if the program seemed to work.",
					"Write a small parser twice, once in C and once in a memory-safe language (Go or Rust), for a simple format such as key=value lines. Fuzz both for ten minutes and compare what each fuzzer finds and how the program fails.",
					"Go has built-in fuzzing (go test -fuzz); for C use libFuzzer with -fsanitize=fuzzer,address.",
				),
			},
			{
				Name:    "Finding Secrets: Patterns, Entropy & Git History",
				Summary: "How secret scanners work, why deleting a key is not enough, and what to do when one leaks.",
				Body: `A secret scanner reads code, configuration and history looking for
credentials. It uses three kinds of evidence. Known formats: providers
make their keys recognisable on purpose (AWS access keys start with
AKIA, GitHub tokens with ghp_, private keys with a BEGIN … PRIVATE KEY
line). Context: a quoted value assigned to a name such as password,
secret or token. And randomness: real keys look random, so their
characters are spread evenly.

Scan before the secret leaves your machine (a pre-commit hook, or push
protection on the server), and scan the whole history, not just the
latest files: git keeps every version. When a secret does leak, the
order matters: revoke and rotate it first, because it must be assumed
copied the moment it was pushed; then find out whether it was used;
rewriting history comes last. And a scanner, like every log, must never
print the secret it found.`,
				Diagram: `commit 1   api_key = "sk_live_…"         ◀ the secret is here forever
commit 2   api_key = os.environ["API_KEY"]
           working tree: clean     history: still holds commit 1
leak ─▶ revoke + rotate ─▶ check its use in logs ─▶ clean history`,
				UnderTheHood: `A key committed and then removed, in a fresh git repository:

  commit 1   config.py: api_key = "sk_test_4eC3…"
  commit 2   config.py: api_key = os.environ["API_KEY"]
  grep in the working tree      0 matches
  git log -p                    2 matches (added, then removed)

The key is still in every clone of the repository.

Why scanners match formats rather than just randomness: the Shannon
entropy of four 40-character strings, in bits per character (at most
5.3 for 40 characters):

  random base64 API key     4.68
  English phrase            4.38
  a file path               3.96
  random hex token          3.76

The hex token, a real secret, scores lower than an English sentence.
And in this academy's own source, 22 of the 367 long string literals
without spaces score 4.5 or more: alphabets and URLs, none of them
secrets. Entropy alone both misses and cries wolf, which is why
gitleaks and GitHub's push protection lead with the formats providers
deliberately make recognisable.`,
				MentalModel: "A pushed secret is a published secret: rotate first, clean up second.",
				TryIt:       "Run git log -p | grep -i -E 'password|secret|token' in one of your own repositories and read what comes back.",
				Analogy: `Shredding the letter after the photocopies were made. Once a key has
been pushed, removing it from the latest version is shredding the
original: the copies (every clone, every fork, every cache) still exist,
so the only real fix is to change the lock.`,
				Example: `GitHub's secret scanning works with over 200 token formats from
providers who agreed to make their keys recognisable and to revoke any
that are found in public. In 2022 Toyota revealed that an access key
for a customer database had sat in a public GitHub repository for five
years. Researchers have repeatedly shown that keys pushed to public
repositories are found by automated scanners within minutes.`,
				Exercises: trio(
					"A developer deletes an API key from a file and commits again. The repository is public. List, in order, what they must do next.",
					"Revoke and rotate the key, check its usage logs for abuse, then remove it from history.",
					"In C, write a scanner that reads a file and reports the line and column of AWS-style access keys (AKIA and 16 capital letters or digits), with no false match inside longer words, without ever printing the key itself.",
					"Check the character before and after the 20 characters, not just the 20.",
					"Add secret scanning to one of your repositories: a pre-commit hook (gitleaks protect, or your own script) and a CI step that scans the full history. Plant a fake key on a branch to prove it is caught, then remove the branch.",
					"Use an obviously fake key in the provider's format, such as an example from their documentation.",
				),
			},
			{
				Name:    "Integrity Checks: Hashes, Checksums & Signed Releases",
				Summary: "Check that what you run is exactly what its authors published.",
				Body: `A cryptographic hash such as SHA-256 turns any file into a 256-bit
fingerprint: any change, even one bit, gives a completely different
value, and nobody can construct two files with the same one. Publishing
a release's SHA-256 lets anyone check that their download is exactly the
published file.

A checksum only proves that two things match. If an attacker controls
the server, they replace both the file and its checksum. Signatures
close the gap: the publisher signs with a private key kept off the
server (GPG, minisign, Sigstore), and users check with the public key.
Build provenance (SLSA attestations) adds which source and which build
produced the file. For dependencies, lockfiles record each package's
hash (go.sum, package-lock.json, pip's --require-hashes), so a package
that changes after you chose it fails the build instead of running.`,
				Diagram: `release file ──sha256──▶ fingerprint ──compare──▶ published checksum
     │                       (file and checksum on one server can be
     │                        replaced together)
     └── signature, made with a private key kept off the server,
         checked with the publisher's public key`,
				UnderTheHood: `SHA-256 of a 1 MB file, then of the same file with one random bit
flipped, 1,000 times: on average 128.2 of the 256 output bits changed,
never fewer than 99. However small the change, the checksum is
unrecognisable, which is what makes it a fingerprint. It is also cheap:
256 MB hashed in 0.20 s here, 1.3 GB/s.

What you are trusting grows with every dependency. In this machine's
Python, the requests package brings 4 more packages with it
(certifi, charset_normalizer, idna and urllib3), and scikit-learn
brings 6, each of them code that runs with your permissions. The
academy itself uses nothing outside Go's standard library; Go projects
that do have dependencies record a hash of each in go.sum, and the go
command refuses a download that does not match.`,
				MentalModel: "A checksum proves the file matches; a signature proves who it matches.",
				TryIt:       "Download a release of a tool you use, together with its checksum file, and check it with sha256sum -c (shasum -a 256 -c on macOS).",
				Analogy: `A sealed envelope with the sender's wax seal. The seal's shape (the
checksum) shows the letter was not opened; only the sender's signet ring
(the private key) can make the seal, so a forger who rewrites the letter
cannot reseal it.`,
				Example: `In 2016 attackers broke into the Linux Mint website and replaced the
download link with a backdoored ISO. A checksum shown on the same
hacked website proves nothing in that situation; the release's GPG
signature, made with a key the attackers did not have, would not have
matched. This academy publishes SHA256 checksums and signed build
provenance with every release, and its installers check the checksum
before installing.`,
				Exercises: trio(
					"An attacker controls a download server that hosts both a file and its SHA-256 checksum. Does checking the checksum protect you? What does?",
					"No: they can change both. A signature from a key the attacker does not have does.",
					"In C, implement SHA-256 from the FIPS 180-4 standard and check it against sha256sum on files of 0, 55, 56, 64 and a million bytes.",
					"The padding needs a second block when 55 < length mod 64: test exactly those lengths.",
					"Verify a real release end to end: download a project's release, its checksum file and its signature (GPG, minisign or Sigstore), verify the signature on the checksum file and then the checksum of the release, and write down which key you trusted and why.",
					"The trust starts with how you got the public key: from the project's website over HTTPS, or a keyserver.",
				),
			},
			{
				Name:    "Security Code Review",
				Summary: "A second pair of eyes that knows where bugs hide.",
				Body: `Code review catches bugs that tests and tools miss, especially logic and
permission mistakes. A security-minded reviewer starts from the trust
boundaries: which lines touch input from outside, authentication,
authorization, cryptography, secrets, file paths or shell commands?
Those lines get the closest reading.

Questions to ask: What happens with empty, huge or malformed input? Is
every request checked for permission, including this new one? Does an
error path leak details or skip a check? Are secrets kept out of code
and logs? Is cryptography done by a well-known library, not by hand? Is
there a test for the dangerous case, not just the happy one?

Keep changes small (a few hundred lines at most) so they can be read
properly, use a written checklist, and require a second reviewer for
the most sensitive code.`,
				UnderTheHood: `Where would a reviewer look first in this academy's own code? Searching
its 39,862 lines of Go (tests left out) for the calls that cross a
trust boundary:

  filepath.Join          47    paths built from names
  http.Client, Request   19    the network
  os.Getenv              16    the environment
  os.Remove, RemoveAll   16    deleting files
  exec.Command            7    running other programs
  reading, writing, renaming files and decoding JSON: 35 more

140 calls on 136 lines, in 22 of the 77 files: 0.34% of the code. That
is where a security review starts, where tests for hostile input
belong, and where a change deserves a second reviewer. The other 99.7%
still matters, but it cannot be reached by an attacker except through
these lines.`,
				MentalModel: "Review the lines that cross a trust boundary as if an attacker wrote the input.",
				TryIt:       "Open any recent pull request in an open-source project you use and find the lines that handle input from outside. Would you have approved them?",
				Analogy: `A second pilot reading the pre-flight checklist aloud. The captain is
skilled, but a checklist and a second person catch the one step anyone
can skip when busy.`,
				Example: `Heartbleed was reviewed. The patch that added the TLS heartbeat
feature to OpenSSL was reviewed and committed at the end of 2011,
shipped in OpenSSL 1.0.1 in 2012, and the missing length check was only
found in 2014. One volunteer reviewer, a large change and no fuzzing let
it through. At Google, every change to the main code base must be
reviewed and approved by another engineer before it is merged.`,
				Exercises: trio(
					"List five kinds of line in a pull request that deserve extra scrutiny from a security point of view.",
					"Think: input, permissions, crypto, secrets, shell and file paths.",
					"Write a script (grep is enough) that lists the trust-boundary lines of one of your projects: file paths, environment variables, network calls, commands and parsers. Review each line it finds with a written checklist and record every finding with its severity.",
					"Severity: could it leak data, change data, or take the service down, and how easily?",
					"Review a real change: pick a small merged pull request in an open-source project that touches input handling, review it with your checklist, and write up what you would have asked. Compare your notes with the actual review comments.",
					"Search the project's pull requests for words such as path, parse, upload or auth.",
				),
			},
			{
				Name:    "Automated Security Testing",
				Summary: "Tools that check every change, so people can focus on judgement.",
				Body: `Automation finds the bugs that humans get bored looking for, on every
change. Static analysis (SAST) reads code for risky patterns: CodeQL,
Semgrep, go vet and compiler warnings. Dependency scanning compares your
dependencies with databases of known vulnerabilities; good tools, such
as govulncheck, report only the ones your code can actually reach.
Secret scanning stops keys from being committed. Fuzzing generates
inputs to crash parsers, ideally with sanitizers on. Dynamic testing
(DAST) probes a running application from outside.

Put the fast checks in CI so they run on every pull request and block
the merge when they fail, and run slower ones (long fuzzing campaigns)
on a schedule. Triage every finding: fix it, or record why it is a false
positive, so the tools stay trusted instead of ignored.`,
				Diagram: `pull request ─▶ build ─▶ tests ─▶ static analysis + secret scan
             ─▶ dependency scan (reachable code only) ─▶ merge
nightly ─▶ long fuzzing runs with sanitizers
        ─▶ every new crash becomes a regression test`,
				UnderTheHood: `Go's built-in fuzzer (go test -fuzz), on two small parsers with one
planted bug each:

  a length byte that is trusted             found in under a second
  an index bug behind the header "TLV2",    not found: 106 million
    fuzzing from the seed "hello, world"    inputs in 5 minutes
  the same bug, seeded with one valid file  found in under a second

The fuzzer ran about 350,000 inputs a second and mutates the ones that
reach new code, but matching four exact bytes by mutation is one chance
in about 4 billion per try. One real example file got past the header
at once. That is why fuzzing projects ship a seed corpus of valid
files, and why a fuzzer started from nothing finds only the shallow
bugs.`,
				MentalModel: "Every bug a tool can find should never reach a human reviewer.",
				TryIt:       "Run go vet ./... and govulncheck ./... on any Go project (or npm audit on a JavaScript one) and read what each reports.",
				Analogy: `Smoke detectors and sprinklers in every room: they never get tired or
distracted. Fire inspectors (reviewers) still visit, but they spend their
time on the problems detectors cannot see.`,
				Example: `Google's OSS-Fuzz service fuzzes open-source projects around the clock
and, by 2023, had helped find and fix over 10,000 vulnerabilities in
about 1,000 projects. Go added fuzzing to its standard test tool in Go
1.18 (2022). This academy's own CI runs vet, the race detector and
govulncheck on every change.`,
				Exercises: trio(
					"A dependency scanner reports a vulnerability in a library function that your code never calls. Is it urgent? What would you do?",
					"Unreachable does not mean irrelevant: note it, and update the dependency in normal maintenance.",
					"Add three automated checks to one of your repositories: a linter or static analyser, a dependency scan, and a secret scan, all running in CI on every pull request and failing the build on a real finding.",
					"GitHub Actions can run CodeQL, and gitleaks scans for secrets.",
					"Write a fuzz test for a parser you wrote (Go's go test -fuzz, or libFuzzer for C), run it for at least 30 minutes, fix what it finds, and keep every crashing input as a regression test.",
					"Start the fuzzer from a few valid example inputs (a seed corpus).",
				),
			},
		},
		Resources: []Resource{
			{"Book", "Building Secure and Reliable Systems (free)", "https://sre.google/books/building-secure-reliable-systems/", "Google's book on designing, writing and running systems that stay secure."},
			{"Site", "OWASP Application Security Verification Standard", "https://owasp.org/www-project-application-security-verification-standard/", "A detailed checklist of security requirements for applications."},
			{"Site", "SEI CERT C Coding Standard", "https://wiki.sei.cmu.edu/confluence/display/c/SEI+CERT+C+Coding+Standard", "Rules and examples for writing secure C."},
			{"Site", "OSS-Fuzz", "https://google.github.io/oss-fuzz/", "Continuous fuzzing for open-source projects, with guides for adding your own."},
			{"Course", "Go fuzzing tutorial", "https://go.dev/doc/tutorial/fuzz", "The official walkthrough of Go's built-in fuzzing."},
			{"Site", "CISA Secure by Design", "https://www.cisa.gov/securebydesign", "Principles for making products safe by default."},
		},
		Blueprints: []Blueprint{
			{"Secure Review Practice", "A personal security review checklist, used on five real changes in open-source projects, with written reviews.",
				[]string{"Write the checklist", "Review five merged pull requests", "Compare with the real reviews", "Refine the checklist"}},
			{"Harden a Small C Library", "Take a small C library (yours or open source), harden its build, run its tests under sanitizers, and fuzz its parser.",
				[]string{"Warnings as errors", "Hardening flags", "Sanitized test run", "Fuzz target", "Fix and add regression tests"}},
			{"Security Gates in CI", "A CI pipeline that blocks merges on static analysis, dependency and secret-scanning findings, with a documented triage process.",
				[]string{"Static analysis", "Dependency scan", "Secret scan", "Branch protection", "Triage notes"}},
		},
		Quiz: []Question{
			{"What does 'fail closed' mean?", "When a check cannot be completed, deny access rather than allow it."},
			{"Why canonicalise input before validating it?", "Attackers use alternative spellings (../, encodings) that pass a check on the raw text but mean something else."},
			{"Which is stronger against buffer overflows: stack canaries or a memory-safe language?", "A memory-safe language: it removes the bug; canaries only make exploiting it harder."},
			{"Why keep pull requests small?", "Reviewers can only read a few hundred lines carefully; large changes hide bugs."},
			{"A file is created with fopen(path, \"w\") under the usual umask of 022. Who can read it?", "Every user on the machine: the file gets mode 0644. Create secrets with open(..., 0600) or mkstemp."},
			{"Why can ^(a+)+$ take seconds on a 26-character input?", "A backtracking engine tries every way of splitting the a's between the two repetitions: the time doubles with every character."},
			{"What single check would have prevented Heartbleed?", "Comparing the claimed payload length with the bytes actually received, and refusing the request when it is larger."},
			{"A 17-byte name overflows a 16-byte field into is_admin inside the same struct. Which of a stack canary, AddressSanitizer and FORTIFY_SOURCE catches it?", "Only FORTIFY_SOURCE, which knows the field's size at compile time; the write never leaves the struct, so the canary and ASan see nothing."},
			{"A key was pushed to a public repository and deleted in the next commit. What comes first?", "Revoke and rotate the key: it is in the history and must be assumed copied already."},
			{"Why does a fuzzer need a seed corpus?", "Random mutation almost never produces exact header bytes or magic numbers; valid examples get it past them to the deeper code."},
			{"An attacker controls the server hosting both a file and its checksum. What still protects you?", "A signature made with a private key kept off that server, checked with the publisher's public key."},
		},
	},

	// -----------------------------------------------------------------
	18: {
		Overview: `Most breaches today do not break code; they log in. Stolen passwords,
tricked employees, over-powerful accounts and misconfigured cloud
storage cause a large share of real incidents. This stage covers how
people and programs prove who they are, how permissions are designed so
one stolen account cannot do much damage, how to secure the cloud
accounts that now run most software, and the zero-trust idea that no
network location is trusted by itself. The measurements are real ones,
made on programs written for the purpose: a JWT verifier that accepted a
forged administrator token with no signature at all, a per-IP limit that
did nothing against credential stuffing from 5,000 addresses, and three
years of team moves that doubled everyone's permissions.`,
		Outcomes: []string{
			"Choose phishing-resistant authentication: MFA, hardware keys and passkeys",
			"Explain sessions, OAuth 2.0 and OpenID Connect, and validate tokens correctly",
			"Design roles and permissions around least privilege, and review them regularly",
			"Secure a cloud account using the shared responsibility model",
			"Explain zero trust and how it differs from trusting the office network",
			"Defend logins against guessing and credential stuffing with limits, breach checks and MFA",
			"Write authorization policies as code, test every allow and deny, and evaluate them safely",
			"Keep secrets in a manager, prefer short-lived credentials, and implement TOTP and JWT checks in C",
		},
		Glossary: []Term{
			{"MFA", "Multi-factor authentication: proving who you are with two or more different kinds of evidence."},
			{"Passkey", "A login credential based on a key pair stored on your device; nothing reusable is typed or sent."},
			{"OAuth 2.0", "A standard for letting one app act on your behalf at another, with limited, revocable permission."},
			{"OpenID Connect", "A layer on top of OAuth 2.0 that tells an app who you are (single sign-on)."},
			{"JWT", "JSON Web Token: a signed token carrying claims such as who you are and when the token expires."},
			{"RBAC", "Role-based access control: permissions are given to roles, and people are given roles."},
			{"Shared responsibility", "The cloud provider secures its infrastructure; you secure your accounts, data and settings."},
			{"Zero trust", "Checking every request's identity and device, instead of trusting anything inside the network."},
			{"TOTP", "Time-based one-time password: a 6-digit code computed from a shared secret and the current 30-second step (RFC 6238)."},
			{"HMAC", "A keyed hash: a fingerprint of a message that only holders of the key can make or check."},
			{"Credential stuffing", "Trying username and password pairs leaked from one site against another, each account once."},
			{"Rate limiting", "Capping how many requests a client, account or address may make in a period, such as with a token bucket."},
			{"Token bucket", "A rate limiter that adds tokens at a steady rate up to a maximum; each request spends one, and an empty bucket refuses."},
			{"Privilege creep", "The slow growth of a person's access as they change roles and keep the old permissions."},
			{"Policy as code", "Authorization rules kept as files in version control, reviewed, tested and evaluated by an engine."},
			{"Explicit deny", "A rule that refuses a request even when another rule allows it; in IAM-style policies, deny always wins."},
			{"Short-lived credential", "A key or token that expires within minutes or hours, so a leaked copy is soon useless."},
			{"Workload identity", "A program proving what it is (a CI job, a cloud VM) to get short-lived credentials, instead of storing a long-lived key."},
		},
		Concepts: []Concept{
			{
				Name:    "Strong Authentication: MFA & Passkeys",
				Summary: "Stop stolen and phished passwords from being enough.",
				Body: `A password alone fails in predictable ways: it is reused, guessed,
leaked in someone else's breach, or typed into a fake site. Multi-factor
authentication adds a second kind of evidence: something you have (a
phone, a security key) or something you are (a fingerprint that unlocks
a key on your device).

Not all factors are equal. SMS codes can be stolen by SIM swapping, and
any code, including an authenticator app's, can be typed into a
convincing fake site. Push prompts can be approved by a tired user who
is flooded with them. Phishing-resistant methods (FIDO2 security keys
and passkeys) sign a challenge for the real website's address, so they
simply do not work on a look-alike domain.

Current guidance (NIST SP 800-63B) also says: allow long passphrases,
check new passwords against lists of breached ones, and stop forcing
routine password changes, which push people towards weak patterns.`,
				Diagram: `password only        ─▶ reused, leaked, phished
password + SMS code  ─▶ SIM swap, phishing
password + app code  ─▶ phishing: the code is typed into the fake site
security key/passkey ─▶ bound to the real site's name: phishing fails`,
				UnderTheHood: `A TOTP code, as authenticator apps make it (RFC 6238): HMAC-SHA1 of the
number of 30-second steps since 1970, keyed with a shared secret, cut
down to 6 digits. A Python implementation reproduced the standard's
test vectors exactly: at 59 seconds the 8-digit test code is 94287082.

How well does a 6-digit code resist guessing? Against a server that
accepts the codes for the current step and one either side, 3 million
random guesses at 5 per 30-second step:

  guesses that worked        13 of 3,000,000   (3 in a million expected)

At 5 tries per step, an attacker who already has the password gets
14,400 guesses a day: about a 4% chance of getting in each day, and
better than even within a month. So limit failures per account, and lock
or slow down after a handful. None of this helps against a phishing
site, which simply asks the user for the current code and uses it at
once; a passkey's signature covers the real site's name, so there is
nothing to type into a fake one.`,
				MentalModel: "The strongest factor is one that cannot be typed into the wrong website.",
				TryIt:       "Turn on a passkey or a security key for your email account, then list which of your other accounts still rely on a password alone.",
				Analogy: `A password is a spoken code word: anyone who overhears it can repeat it.
A passkey is a key cut for one specific lock; hold it up to a fake door
and it simply does not fit.`,
				Example: `In 2022 an attacker who had an Uber contractor's password sent repeated
MFA push requests, then messaged the contractor pretending to be IT
support until one was approved. The same year, Cloudflare employees
received convincing SMS phishing messages and some typed in their
passwords, but the attack failed because Cloudflare required hardware
security keys, which will not sign in to a fake site.`,
				Exercises: trio(
					"Rank from weakest to strongest against phishing: password only, SMS code, authenticator app code, push approval with number matching, hardware security key.",
					"Ask for each: can an attacker's fake site pass it on in real time?",
					"In C, implement TOTP (RFC 6238) from its definition, with SHA-1 and HMAC written by you, and check it against the standard's test vectors; then explain why production code should still use a well-tested library.",
					"Decode the Base32 secret, HMAC-SHA1 the 8-byte big-endian step counter, take 4 bytes at the offset given by the last nibble, mask the top bit, and take the remainder mod 1,000,000.",
					"Do an account-security audit of your own digital life: every account that can reset others (email, phone carrier, password manager), its second factor, and its recovery options. Upgrade the weakest three.",
					"Your email account can reset almost everything else, so protect it first and best.",
				),
			},
			{
				Name:    "Rate Limiting & Credential Stuffing Defences",
				Summary: "Slow down guessing, and notice when stolen passwords are being tried.",
				Body: `A login form is an oracle: it answers "right" or "wrong" to anyone.
Two attacks use it. Guessing tries many passwords (or MFA codes)
against one account; per-account limits stop it: a few failures, then
growing delays or a temporary lock. Credential stuffing tries pairs
leaked from other sites, each account once, from thousands of addresses;
per-account limits never trigger, and per-address limits barely slow it.

The defences against stuffing are different: refuse passwords found in
breach lists (Have I Been Pwned's range API checks one without sending
it), require MFA, and watch the pattern itself: many accounts, a high
failure ratio, unfamiliar addresses and devices. A token bucket is the
usual limiter: tokens refill at a steady rate up to a burst size, each
attempt spends one, and an empty bucket answers 429 Too Many Requests.
Locking accounts outright lets anyone lock out your users on purpose,
so prefer delays and step-up checks.`,
				Diagram: `token bucket: capacity 5, refill 1 every 12 s
  attempt ─▶ token left? ─yes─▶ check password
                         └no──▶ 429, try again later
guessing  : 1 account × many passwords   ─▶ per-account limit stops it
stuffing  : many accounts × 1 password   ─▶ breach check + MFA stop it`,
				UnderTheHood: `Credential stuffing, worked through: an attacker with 1,000,000 leaked
email and password pairs, 0.5% of them reused on your site, sending
1,000 login attempts a second for a day:

  defence                                  pairs tried   accounts lost
  none                                      1,000,000        5,000
  10 a minute per IP, 50 addresses            720,000        3,600
  10 a minute per IP, 5,000 addresses       1,000,000        5,000
  lock an account after 5 failures          1,000,000        5,000

Every attempt is a different account, tried once, so the lockout never
triggers, and addresses are cheap: residential proxy networks rent out
thousands. Rejecting breached passwords at sign-up, or requiring MFA,
takes the number of accounts lost to nearly zero. The per-account limit
is still essential against the opposite attack: it is what keeps a
6-digit code at the odds of the previous lesson. Without it, an
attacker tries codes as fast as the server answers, and one Python core
computes 367,000 codes a second.`,
				MentalModel: "Limit guessing per account; stop stuffing with breach checks and MFA, because it never guesses twice.",
				TryIt:       "Look up one of your old passwords' SHA-1 prefix in the Have I Been Pwned range API (only the first 5 hex digits are sent), and see whether it has appeared in a breach.",
				Analogy: `A cash machine keeps the card after three wrong PINs: that stops
someone guessing one card's PIN. It does nothing against a thief with a
stack of stolen cards and their PINs who tries each card once; for
that, the bank watches for unusual withdrawals.`,
				Example: `In 2019 State Farm reset passwords after a credential-stuffing attack on
its customer accounts. In 2023 the same technique against 23andMe
logged in to accounts with reused passwords and, through a
relatives-sharing feature, exposed the profiles of about 6.9 million
people. 23andMe then required two-step verification for all customers.`,
				Exercises: trio(
					"Why does locking an account after 5 failed logins not stop credential stuffing, and what could it do to your users?",
					"Each account is tried once; and anyone can lock out a victim by failing 5 times on purpose.",
					"In C, implement a token bucket: allow(bucket, now) returns 1 and spends a token when one is left, refilling at R tokens a second up to a capacity C. Test it with bursts, steady traffic and long pauses.",
					"Store the token count as a double and the time of the last refill; refill lazily on each call.",
					"Add per-account and per-address rate limiting to the login of a web app you built, plus a breached-password check on sign-up using the Have I Been Pwned range API, and log every refused attempt.",
					"Send only the first 5 hex digits of the password's SHA-1 hash, and compare the returned suffixes locally.",
				),
			},
			{
				Name:    "Sessions, Tokens & Single Sign-On",
				Summary: "How a login is remembered, and how one login works across many apps.",
				Body: `After you log in, the server has to remember it. Classic web apps use a
session: a random ID in a cookie (HttpOnly, Secure, SameSite), pointing
to state on the server that can be revoked at once. Many modern systems
use signed tokens instead, often JWTs, which carry claims (who, for which
app, until when) that any service can verify without asking a central
database. That convenience has a cost: a token is valid until it
expires, so keep lifetimes short and use refresh tokens to renew them.

OAuth 2.0 lets one app act for you at another with limited permission
(scopes), such as "read my calendar" without your password. OpenID
Connect adds identity on top, which gives single sign-on: log in once
with your company or Google account. When you accept a token, always
check its signature with the expected key and algorithm, its issuer,
its audience (it was meant for you) and its expiry.`,
				Diagram: `you ─▶ app: "log in"
app ─▶ identity provider: send the user here (OpenID Connect)
you ─▶ identity provider: log in with MFA
identity provider ─▶ app: signed ID token
                          (who, for which app, until when)
app: verify signature, issuer, audience, expiry ─▶ start your session`,
				UnderTheHood: `A signed token is only as strong as the code that checks it. An HS256
JWT with six claims came to 237 bytes (a random 256-bit session ID is 43
characters). Two verifiers, in Python, given four tokens:

  token                             takes alg from     decides alg and
                                    the token          checks iss, aud, exp
  genuine                           accepted           accepted
  forged: alg "none", role changed  ACCEPTED as        refused
    to admin, no signature          admin
  payload changed, old signature    refused            refused
  genuine but expired               accepted           refused

The first verifier did check signatures, but it let the token say how
to check it, and "none" means not at all. Several JWT libraries shipped
exactly this bug in 2015. Decide the algorithm and the key in your own
code, never from the token, and check issuer, audience and expiry
every time.`,
				MentalModel: "A token is a signed promise: check who signed it, who it is for, and when it expires.",
				TryIt:       "Paste a JWT from a tutorial into a decoder such as jwt.io and read its three parts. Note that anyone can read the claims; the signature only proves they were not changed.",
				Analogy: `A concert wristband. The ticket office (identity provider) checks your
ID once and gives you a band stamped for tonight's show (audience) and
tonight's date (expiry). Security staff inside only check the band, not
your ID again, so a band for the wrong venue or the wrong night must be
refused.`,
				Example: `In 2023 a China-based group known as Storm-0558 used a stolen Microsoft
consumer signing key to forge tokens that Microsoft's systems also
accepted for enterprise email accounts, reading the email of US
government agencies. A US review board found in 2024 that the key
should never have been accepted for those accounts: the services did not
correctly check which signing keys were allowed for which kind of token.`,
				Exercises: trio(
					"Why should an API that receives a JWT check the audience claim, even when the signature is valid?",
					"A valid token issued for a different app would otherwise be accepted by yours.",
					"Implement session login for a small web app twice: once with server-side sessions in a secure cookie, once with short-lived signed tokens plus a refresh token. Add logout to both and explain which one can revoke access instantly.",
					"Server-side sessions are deleted in one step; a token stays valid until it expires unless you keep a deny list.",
					"Add 'Log in with GitHub' (OAuth 2.0) to a small app using an established library, requesting the smallest scope you need, and write down each redirect and what it carries.",
					"Use the authorization code flow with PKCE; never put client secrets in browser code.",
				),
			},
			{
				Name:    "Authorization Design: Roles, Attributes & Least Privilege",
				Summary: "Grant the minimum, check on the server, and review access regularly.",
				Body: `Checking permissions on every request is the rule (Stage 10). This
concept is about designing the permissions themselves so mistakes stay
small. Role-based access control (RBAC) groups permissions into roles
such as viewer, editor and admin. Attribute-based control (ABAC) adds
conditions: the owner of the document, the user's department, the time
of day. Many teams write policies as code (Open Policy Agent, cloud IAM
policies) so they can be reviewed and tested like any other change.

Least privilege in practice: start new accounts and services with
nothing and add what they need; avoid permanent admin rights by granting
them just in time, for a limited period and with a reason; separate
duties so no single person can both make and approve a payment; and
review who has access at regular intervals, removing what is unused.
Service accounts and API keys need the same care as people.`,
				Diagram: `user ─▶ roles ─▶ permissions            (RBAC: coarse, simple)
      + attributes: owner? department? time? device?
                                          (ABAC: fine-grained)
policy as code ─▶ reviewed ─▶ tested ─▶ deployed`,
				UnderTheHood: `Privilege creep, simulated: 500 staff in 20 teams, each team needing 5
permissions. Every month 3% of people move team and are given the new
team's permissions. After three years:

                           permissions   not needed   most held
                           per person    any more     by anyone
  never reviewed              10.2          5.2           25
  reviewed every quarter       5.0          0.0            5

Nobody was ever given too much on purpose. Half of the access in the
unreviewed company is left over from old jobs, and the person holding
the most could act in five teams, the account an attacker most wants
to phish. Removing access is the step everyone forgets, so it has to be
a scheduled job with an owner, not a good intention.`,
				MentalModel: "Ask of every permission: who needs it, for how long, and who checks it is still needed?",
				TryIt:       "Type in and run this stage's rbac.py, then add an 'auditor' role that can read everything but change nothing, and test it.",
				Analogy: `A hospital: nurses can read the charts of patients on their ward, not
the whole hospital (attributes), a surgeon's theatre access is for the
day of the operation (just in time), and controlled drugs need two staff
to sign them out (separation of duties).`,
				Example: `In 2019 an attacker stole data on about 100 million Capital One
customers. A misconfigured web firewall could be tricked into fetching
cloud credentials for its own role, and that role had far more
permission than a firewall needed, including listing and reading the
storage buckets with customer data.`,
				Exercises: trio(
					"A reporting service needs to read orders once a night. What permissions should its account have, and what should it certainly not have?",
					"Read only, only the orders table, ideally only during the job's time window.",
					"Extend the rbac.py type-in: add an auditor role, an ABAC rule that editors may only edit documents in their own team, and a test for every allowed and every denied case.",
					"Write the denied cases first: they are the ones that protect you.",
					"Do an access review of a real system you use (a GitHub organisation, a shared drive, a cloud account): list every person and key with write or admin access, when each last used it, and remove or reduce at least one.",
					"GitHub shows organisation members' roles; cloud providers show when each access key was last used.",
				),
			},
			{
				Name:    "Policy as Code: Writing & Testing Authorization Rules",
				Summary: "Authorization rules as reviewed, tested files, evaluated the same way every time.",
				Body: `When permissions live in code scattered through an application, nobody
can answer "who can delete invoices?". Policy as code moves them into
rule files that an engine evaluates: AWS IAM policies, Open Policy
Agent (Rego), Amazon's Cedar, Kubernetes RBAC. Each request (who, which
action, which resource) goes to the engine, which answers allow or
deny.

The usual evaluation rules are worth knowing exactly: deny by default;
an explicit deny beats any allow; wildcards match whole families of
actions or resources, which is where most over-granting hides. Because
rules are files, they go through pull requests, and they get tests:
for every rule, a request that it must allow and one it must not. A
policy engine is also a parser of untrusted patterns, so its matching
must be linear: wildcards evaluated by naive backtracking can be made
to take seconds per request.`,
				Diagram: `request (principal, action, resource)
   │
   ▼  any deny rule matches?  ──yes──▶ DENY (explicit deny wins)
   │ no
   ▼  any allow rule matches? ──yes──▶ ALLOW
   │ no
   ▼  DENY (default)`,
				UnderTheHood: `Matching a wildcard pattern such as *a*a*a*a*a*a*b against a resource
name made of n a's, in C:

  n                       40        60        80
  naive recursion       0.11 s    2.3 s     23.5 s
  greedy, one pass      37 ns     74 ns     70 ns

The naive matcher tries every way of giving characters to every star;
the greedy one remembers only the last star and moves forward. One
pattern of 13 characters and an 80-character name kept a CPU busy for
23 seconds per request: a policy engine (or a file-path matcher, or a
router) that accepts patterns from users must match in linear time.

The other lesson is about wildcards themselves. Over-broad actions
such as storage:* or iam:* are where accidental access lives, which is
why tests that enumerate what a policy allows find problems that
reading it does not.`,
				MentalModel: "Default deny, explicit deny wins, and every rule has a test that it allows what it should and nothing more.",
				TryIt:       "Read one IAM policy from a cloud tutorial and write down every action its wildcards match; most people underestimate the count.",
				Analogy: `A building's access list kept at the security desk instead of in each
guard's head: written down, the same for every guard, changed only with
a signature, and checked by someone trying the doors that should stay
locked.`,
				Example: `Amazon open-sourced Cedar in 2023, a policy language designed so its
evaluator can be formally verified; AWS's IAM Access Analyzer uses
automated reasoning to prove what a policy can and cannot allow.
Kubernetes RBAC and Open Policy Agent are the usual tools for clusters
and microservices.`,
				Exercises: trio(
					"A policy allows storage:* on reports/* and denies storage:Delete* on reports/2024/*. Is storage:DeleteObject on reports/2024/q1.pdf allowed? And on reports/2025/q1.pdf?",
					"Explicit deny wins for 2024; for 2025 only the allow matches.",
					"In C, write an evaluator for allow and deny rules with * wildcards in the principal, action and resource, with default deny and deny winning, and a linear-time wildcard matcher. Test it with a pattern that would make naive backtracking slow.",
					"Remember the position of the last * and where it started matching; on a mismatch, give that star one more character.",
					"Write the authorization rules of a small app in Open Policy Agent (or Cedar), with a test for every allowed and denied case, and run the tests in CI so a pull request that loosens a rule fails unless its tests change too.",
					"opa test runs Rego tests; keep the policy and its tests in the same pull request.",
				),
			},
			{
				Name:    "Cloud Security & Shared Responsibility",
				Summary: "The provider secures the cloud; you secure what you put in it.",
				Body: `A cloud provider secures its data centres, hardware and core services.
Everything you configure is your responsibility: accounts and
permissions, network rules, what is public, encryption settings,
logging, and patching the software you run. Most cloud incidents are
configuration mistakes, not broken providers.

A secure baseline: protect the root or owner account with a hardware
key and never use it day to day; give people and services roles with
least privilege, not long-lived keys; block public access to storage
unless something is meant to be public; turn on audit logs (such as AWS
CloudTrail) and keep them where an attacker cannot delete them; encrypt
data at rest with managed keys; and define everything as code
(Terraform, CloudFormation) so changes are reviewed, repeatable and
easy to audit. Posture tools and CIS Benchmarks check the configuration
continuously.`,
				Diagram: `             provider secures          you secure
IaaS (VMs)   hardware, network,        OS patches, apps, data,
             hypervisor                accounts, firewall rules
SaaS (email) almost everything         accounts, MFA, sharing
                                       settings, your data`,
				UnderTheHood: `Encryption at rest costs almost nothing. openssl speed on this machine,
one core, AES-256-GCM (the cipher used by cloud disks and storage):

  block size     16 bytes    1 KB        8 KB
  throughput     0.7 GB/s    6.7 GB/s    10.8 GB/s

One core could encrypt a terabyte in about a minute and a half, thanks
to the AES instructions built into the CPU, so there is no performance
reason to store anything unencrypted. Turning it on is a setting, which
is exactly why it is the customer's job. What it does not do is stop
anyone with valid credentials: storage decrypts for whoever may read
it, so permissions and public-access settings remain the real lock, as
the misconfigured buckets in this lesson's examples show.`,
				MentalModel: "If you can configure it, you are responsible for how it is configured.",
				TryIt:       "In any cloud account you own (a free tier is enough), find the setting that blocks public access to storage and the audit log, and check that both are on.",
				Analogy: `Renting a flat in a well-guarded building. The landlord maintains the
front door, the lifts and the fire alarms. Leaving your own flat door
open, or giving copies of your key to strangers, is still your problem.`,
				Example: `In 2023 Toyota disclosed that a misconfigured cloud environment had
left the vehicle location data of about 2.15 million customers in Japan
publicly accessible for roughly a decade. Gartner predicted in 2019
that through 2025, 99% of cloud security failures would be the
customer's fault, mostly misconfiguration.`,
				Exercises: trio(
					"For each, say whether the provider or the customer is responsible: a disk failure in the data centre; a storage bucket set to public; an unpatched web server on a rented VM; a weak password on the admin account.",
					"If it is a setting you choose or software you install, it is yours.",
					"Write infrastructure as code (Terraform or your provider's tool) for a small, private storage bucket with public access blocked, encryption on, access logging on, and a role that can only read it. Review your own plan output before applying.",
					"terraform plan shows exactly what would change; read it like a code review.",
					"Harden a free-tier cloud account: hardware-key MFA on the owner account, a day-to-day user with limited rights, a budget alert, audit logging to a protected bucket, and a check with the provider's security advisor or a CIS Benchmark. Record what you changed.",
					"A budget alert is also a security alarm: stolen accounts are often used to mine cryptocurrency.",
				),
			},
			{
				Name:    "Secrets Management & Short-Lived Credentials",
				Summary: "Keep keys in a manager, hand out credentials that expire, and let workloads prove who they are.",
				Body: `A secrets manager (HashiCorp Vault, AWS Secrets Manager, Azure Key
Vault, Google Secret Manager) stores keys encrypted, hands them only to
identities allowed to read them, logs every read, and can rotate them on
a schedule. Better still is not storing a long-lived key at all. Cloud
VMs and containers get credentials from their role; a CI job proves
which repository and branch it is through OpenID Connect and receives
credentials for an hour (workload identity federation); people get
temporary elevated access with a reason.

Short lifetimes change what a leak means: a key valid for a year is
valid for six months on average after it leaks; a credential valid for
an hour is useless before most leaks are even noticed. Keep secrets out
of places they spread from: images, logs, URLs, crash reports and
environment variables, which every child process inherits. Encrypt data
with keys from a key management service (envelope encryption), so the
master key never leaves it.`,
				Diagram: `CI job ──OIDC token (repo, branch)──▶ cloud identity service
                                       │ checks the trust policy
       ◀── credentials valid for 1 hour ┘
app ──role identity──▶ secrets manager ──▶ database password
                       (logged, rotated)`,
				UnderTheHood: `How long a leaked credential stays useful, if it leaks at a random
moment of its life:

  credential                    still valid, on average, after a leak
  rotated yearly                  4,380 hours  (six months)
  rotated every 90 days           1,080 hours  (45 days)
  valid for 1 hour                    0.5 hours
  valid for 15 minutes                0.1 hours

Making them is cheap: one Python core signed 394,025 HMAC-SHA256
tokens a second, so issuing fresh hourly credentials to 10,000 services
costs a fraction of a second an hour.

Where an environment variable goes, measured with a dummy secret in a
shell: all 3 child programs started from it (sh, python3 and awk)
received it, and it was readable in /proc/PID/environ by that user and
by root. Every tool, plugin and crash reporter the process starts sees
it. A secrets manager's read is logged and scoped; an environment
variable is copied everywhere silently.`,
				MentalModel: "The best secret is one that expires before anyone can use a stolen copy.",
				TryIt:       "In a GitHub Actions workflow of yours, list every secret it uses and whether each could be replaced by OIDC federation with your cloud provider.",
				Analogy: `A hotel key card that stops working at checkout, instead of a metal
key that opens the room for as long as the lock is unchanged. Losing the
card on the last day costs nothing; losing the metal key means changing
the lock.`,
				Example: `In 2022 attackers who had stolen an employee's session found
hard-coded administrator credentials for Uber's privileged access
management system in a network share script, and from there reached
most of Uber's internal tools. GitHub Actions, GitLab CI and the major
cloud providers now support OIDC federation so that CI pipelines need no
stored cloud keys at all.`,
				Exercises: trio(
					"A deploy key with full access is stored as a CI secret and never rotated. List three ways it could leak and two changes that would make a leak far less harmful.",
					"Logs, a forked pull request workflow, a developer's laptop; then short-lived credentials and a smaller scope.",
					"In C, write a program that issues and checks short-lived tokens: subject, expiry, and an HMAC-SHA256 tag over both (reuse your sha256.c), refusing a token whose tag is wrong or whose expiry has passed.",
					"HMAC(K, m) = H((K ^ opad) || H((K ^ ipad) || m)), with the key padded to 64 bytes.",
					"Replace one stored long-lived cloud key in a CI pipeline of yours with OIDC federation (GitHub Actions to AWS, Azure or Google Cloud), restricted to one repository and branch, and delete the old key.",
					"The cloud side's trust policy should check the token's repository and ref claims, not just its issuer.",
				),
			},
			{
				Name:    "Zero Trust: Never Trust the Network Alone",
				Summary: "Verify every request by who is asking and from what device.",
				Body: `The old model was a castle with a moat: a firewall around the office
network, and anything inside trusted. It fails once one laptop inside is
compromised, and it never fitted remote work and cloud services anyway.

Zero trust removes trust based on network location. Every request to
every application is checked: who is the user (strong authentication),
is the device known and healthy (up to date, disk encrypted), and is
this particular access allowed by policy? Access is granted per
application, not to a whole network, and is re-checked continuously.
Traffic between services is authenticated and encrypted too (often
with mutual TLS). NIST described the architecture in SP 800-207 (2020).

Zero trust is a direction, not a product: teams move there step by step,
starting with strong identity and device checks for their most
important applications.`,
				Diagram: `castle and moat:  outside ─▶ [firewall] ─▶ inside: everything trusted
zero trust:       any network ─▶ access proxy ─▶ one application
                                  │ checks: user identity + MFA,
                                  │ device health, policy
                                  └ every request, every time`,
				UnderTheHood: `Zero trust checks every request. Can you afford it? Measured with
openssl on this machine:

  ECDSA P-256 signatures       33,333 a second   (30 µs each)
  ECDSA P-256 verifications    11,421 a second   (88 µs each)
  RSA-2048 signatures           2,572 a second   (0.39 ms each)
  RSA-2048 verifications       43,826 a second
  full TLS 1.3 handshakes      4,226 in 6 seconds, client and server
                               on the same 4 cores: about 0.8 ms of
                               client CPU each

Checking a token's or a device certificate's signature costs tens of
microseconds, and a whole mutual-TLS handshake under a millisecond,
paid once per connection rather than per request. The cost of zero
trust is in the engineering (an inventory of devices, an identity for
every service), not in the cryptography.`,
				MentalModel: "Being on the office network proves nothing; identity and device health prove something.",
				TryIt:       "Draw how you reach your email today. Where is identity checked, where is the device checked, and where does being on a certain network make a difference?",
				Analogy: `A hotel where your key card is checked at every door you open, not just
at the front entrance. Walking in behind a guest gets a stranger into
the lobby, and nowhere else.`,
				Example: `After the 2009 "Operation Aurora" attacks, Google started BeyondCorp,
moving its employees off a privileged corporate network so that every
application checks the user and device instead. It described the
approach publicly in 2014, and it became a model for zero-trust
products across the industry.`,
				Exercises: trio(
					"An attacker takes over one laptop on the office Wi-Fi. What can they reach in a castle-and-moat network, and what in a zero-trust one?",
					"In zero trust the laptop's network position grants nothing; only its user's permissions and its health count.",
					"Put a small internal web app behind an identity-aware proxy (for example oauth2-proxy in front of it), so that only logged-in users from your organisation can reach it, and confirm that the app itself is not reachable directly.",
					"Bind the app to localhost so the proxy is the only way in.",
					"Write a one-page zero-trust plan for a small organisation (a club, a family, a small business): its important applications, how each will check identity and device, and the order you would do it in.",
					"Start with email and the password manager: they unlock everything else.",
				),
			},
		},
		Resources: []Resource{
			{"Site", "NIST SP 800-63B: Digital Identity Guidelines", "https://pages.nist.gov/800-63-3/sp800-63b.html", "The US standard for passwords, MFA and authenticators, readable online."},
			{"Site", "FIDO Alliance: passkeys", "https://fidoalliance.org/passkeys/", "What passkeys are and how they resist phishing."},
			{"Site", "OAuth 2.0 Simplified (Aaron Parecki)", "https://www.oauth.com/", "A clear, free guide to OAuth flows and tokens."},
			{"Site", "OWASP Authentication Cheat Sheet", "https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html", "Practical rules for login, passwords and MFA."},
			{"Site", "AWS IAM security best practices", "https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html", "Least privilege and account hygiene in one cloud provider's words."},
			{"Paper", "NIST SP 800-207: Zero Trust Architecture", "https://csrc.nist.gov/pubs/sp/800/207/final", "The reference description of zero trust."},
			{"Tool", "Open Policy Agent", "https://www.openpolicyagent.org/docs/latest/", "Write, test and review authorization policies as code."},
		},
		Blueprints: []Blueprint{
			{"Login Service Done Right", "A login service with passkeys or TOTP, secure sessions, token refresh, rate limiting and an audit log of every login.",
				[]string{"Password + TOTP or passkey", "Secure session cookies", "Short-lived tokens with refresh", "Rate limiting", "Login audit log"}},
			{"Policy as Code", "Authorization rules for a small app written in Open Policy Agent, with a test for every allowed and denied case, running in CI.",
				[]string{"Roles and attributes", "Policy file", "Allow and deny tests", "CI check"}},
			{"Hardened Cloud Baseline", "A free-tier cloud account defined as code: protected owner account, least-privilege roles, private storage, audit logs and alerts.",
				[]string{"Owner account locked down", "Roles instead of keys", "Private encrypted storage", "Audit logging", "Budget and security alerts"}},
		},
		Quiz: []Question{
			{"Why are passkeys resistant to phishing?", "They sign a challenge bound to the real website's name, so they will not work on a look-alike domain."},
			{"Name four things to check on a JWT before trusting it.", "The signature (with the expected key and algorithm), the issuer, the audience and the expiry."},
			{"In the shared responsibility model, who is responsible for a storage bucket left public?", "The customer: it is a configuration they chose."},
			{"What does zero trust stop trusting?", "Network location: being inside the office network grants nothing by itself."},
			{"A server accepts TOTP codes for the current step and one either side, and allows 5 guesses per step. Roughly what chance does an attacker with the password have per day?", "About 4%: 14,400 guesses a day at 3 in a million each. Per-account limits and lockout are what keep it low."},
			{"Why does a per-account lockout not stop credential stuffing?", "Each account is tried once with its leaked password, so no account ever reaches the failure limit."},
			{"A JWT verifier reads the algorithm from the token's header. What can go wrong?", "A forged token can say alg none (or another algorithm) and be accepted without a valid signature: the verifier must decide the algorithm itself."},
			{"In an IAM-style policy, a request matches both an allow and a deny rule. What is the decision?", "Deny: an explicit deny always wins, and anything matching no rule is denied by default."},
			{"Why prefer a credential that lasts an hour to a key rotated every 90 days?", "A leaked copy is valid for about half an hour on average instead of about 45 days."},
			{"What is privilege creep, and what stops it?", "Access accumulating as people change roles and keep old permissions; regular access reviews that remove what is not needed."},
		},
	},

	// -----------------------------------------------------------------
	19: {
		Overview: `Defenders cannot stop every attack, but they can make networks hard to
move through and make attackers visible. This stage covers dividing a
network so one compromised machine cannot reach everything, collecting
the logs and network records that reveal what happened, detecting
attacks with signatures and anomaly detection, mapping detections to the
MITRE ATT&CK catalogue of attacker behaviour, and running the day-to-day
work of monitoring without drowning in alerts.`,
		Outcomes: []string{
			"Design firewall rules and network zones that deny by default",
			"Decide what to log, collect it centrally, and protect it",
			"Explain signature and anomaly detection, and why false alarms are inevitable",
			"Map an attack to MITRE ATT&CK techniques and write a detection for one",
			"Triage alerts with a runbook and tune noisy rules",
		},
		Glossary: []Term{
			{"Segmentation", "Dividing a network into zones so traffic between them can be controlled."},
			{"DMZ", "A zone for internet-facing servers, separated from the internal network."},
			{"Egress filtering", "Controlling which connections may leave a network, not just which may enter."},
			{"Telemetry", "The logs, metrics and network records a system produces about what it is doing."},
			{"IDS", "Intrusion detection system: watches traffic or hosts and raises alerts on suspicious activity."},
			{"SIEM", "Security information and event management: collects logs in one place, correlates them and raises alerts."},
			{"False positive", "An alert for something harmless. A false negative is an attack that raises no alert."},
			{"MITRE ATT&CK", "A public catalogue of attacker tactics and techniques, used to plan and measure detection."},
		},
		Concepts: []Concept{
			{
				Name:    "Firewalls & Network Segmentation",
				Summary: "Deny by default, and never let one compromised machine reach everything.",
				Body: `A firewall filters traffic by rules. Stateful firewalls remember
connections, so replies to traffic you started are allowed while
unsolicited traffic is not. Good rule sets deny by default and allow
only what is needed, in both directions: egress filtering stops a
compromised server from phoning home or sending data out.

Segmentation splits a network into zones with firewalls between them:
internet-facing servers in a DMZ, office laptops in another zone,
databases and management systems in the most protected one, reachable
only from the machines that need them. Administrators reach sensitive
zones through a single hardened entry point (a bastion host or VPN) with
strong authentication. A flat network, where everything can talk to
everything, lets one infected machine reach every other.`,
				Diagram: `internet ─▶ [fw] ─▶ DMZ: web servers
                     │ only the app port
                     ▼
office ───▶ [fw] ─▶ internal: app servers ─▶ [fw] ─▶ databases
admins ───▶ VPN / bastion with MFA ──────────────────▲`,
				MentalModel: "Draw who needs to talk to whom; everything else is blocked.",
				TryIt:       "List the listening ports on your computer (ss -tulpn on Linux, netstat -an on Windows or macOS) and decide which of them need to be reachable from other machines.",
				Analogy: `A ship divided into watertight compartments. One hole floods one
compartment, and the ship stays afloat. A ship with one big open hull
sinks from a single hole.`,
				Example: `The 2013 Target breach started with a heating contractor's login to a
vendor portal, and the attackers then reached the checkout systems that
held card data. A US Senate report in 2014 pointed to weak separation
between those networks. In 2017 the WannaCry worm spread through
Windows file sharing (port 445) across flat networks, disrupting
hospitals in the UK's National Health Service.`,
				Exercises: trio(
					"A web server in the DMZ needs to reach its database. Write the minimum firewall rules: which direction, which port, from where to where?",
					"Allow only the web server's address to the database port; deny everything else in both directions.",
					"Write an nftables or ufw rule set for a small server that allows SSH from one address range and HTTPS from anywhere, denies everything else inbound, and limits outbound traffic to DNS, NTP, HTTPS and package updates. Test it from another machine.",
					"Test the denials as carefully as the allows: from a second machine, try a port that should be closed.",
					"Draw the network of your home or a small office: every device, what it needs to reach, and a proposed set of zones (for example: work devices, guests, smart-home devices) with the rules between them. Implement the guest zone if your router supports it.",
					"Smart TVs and cameras rarely need to reach your laptop; put them in their own zone.",
				),
			},
			{
				Name:    "Logging & Telemetry",
				Summary: "Record the events that answer 'what happened?', and keep them safe.",
				Body: `When something goes wrong, logs are the only witness. Log the events
that matter for security: logins and failed logins, permission denials,
changes to accounts and permissions, administrator actions, and
configuration changes. At the network level, collect connection records
(flow logs), DNS queries and web proxy logs; they show which machine
talked to what, even when the content was encrypted.

Make logs useful: structured (JSON with named fields), with accurate
synchronised clocks (NTP) and time zones, and with enough context (user,
source address, request ID). Send them to a central store as they happen,
because attackers delete local logs, and restrict who can change or
delete them. Keep them long enough to investigate attacks discovered
months later. Never log secrets: passwords, tokens and full card
numbers turn your logs into a target.`,
				Diagram: `servers ──┐
laptops ──┼─▶ collector ─▶ central store
firewall ─┤                (append-only, access-controlled)
DNS ──────┘                     │
                                ├─▶ search during investigations
                                └─▶ detection rules ─▶ alerts`,
				MentalModel: "Log so that a stranger could reconstruct the story months later.",
				TryIt:       "On Linux, run journalctl -u ssh --since today (or read /var/log/auth.log); on Windows, open Event Viewer and find event ID 4625 (failed logon). What would an investigator want that is missing?",
				Analogy: `A ship's logbook, written as things happen, in ink, with the time of
every entry, and a copy sent ashore each day, so that a wreck, or a
dishonest captain, cannot erase the record.`,
				Example: `The security company FireEye discovered in 2020 that it had been
breached when an alert flagged a new phone registered for an employee's
two-factor login. Following that single logged event uncovered the
SolarWinds supply-chain attack, whose backdoored update had reached
thousands of organisations months earlier.`,
				Exercises: trio(
					"Which of these should be logged, and which must never be: failed logins, the password typed on a failed login, a role change, a session token, the source IP of an admin action?",
					"Log that a secret was used, never the secret itself.",
					"Add structured security logging to a small app you built: log in, log out, failed login, permission denied and role change as JSON lines with time, user, source address and outcome. Check that no password or token ever appears.",
					"Write a test that logs in with a known password and asserts that string is absent from the log output.",
					"Set up central logging for two machines or containers you own (for example with rsyslog or a small Loki or OpenSearch stack), then delete a local log file on one machine and show that the central copy still has it.",
					"Forward over an encrypted connection, and restrict who can delete from the central store.",
				),
			},
			{
				Name:    "Intrusion Detection: Signatures & Anomalies",
				Summary: "Spot known attacks by their fingerprints, and new ones by what is unusual.",
				Body: `An intrusion detection system (IDS) watches network traffic (Suricata,
Snort, Zeek) or activity on a host, and raises alerts. Signature
detection matches known patterns, such as a known piece of malware or an
exploit's request. It is precise, but it misses anything new. Anomaly
detection learns what is normal (which users log in when, how much data
a server usually sends) and flags departures. It can catch new attacks,
but it also flags every unusual but harmless event.

Base rates make this hard: real attacks are rare compared with normal
events. If one event in a million is an attack, even a detector that is
99% accurate raises about 10,000 false alarms for every real one. So
good detection combines signals, focuses on the behaviours attackers
cannot avoid, and is tuned continuously. An intrusion prevention system
(IPS) goes further and blocks what it detects, so its rules must be
especially precise.`,
				Diagram: `1,000,000 events, 1 of them an attack; detector 99% accurate:
  attack caught:        1
  false alarms:    ~10,000   (1% of 999,999 harmless events)
  → an alert is almost always harmless unless the rule is very precise`,
				MentalModel: "Rare events and imperfect detectors mean most alerts are false; design for that.",
				TryIt:       "Type in and run this stage's bruteforce.py. Then change the log so an attacker stays just under the limit, and notice that the detector never fires.",
				Analogy: `Airport security uses both: a list of banned items on the X-ray screen
(signatures), and trained officers who notice a passenger behaving
oddly (anomalies). The list misses new tricks; the officers stop many
innocent travellers.`,
				Example: `Martin Roesch released Snort in 1998 as a small open-source network
intrusion detection tool, and it became one of the most widely used. In
2000 Stefan Axelsson's paper "The base-rate fallacy and the difficulty
of intrusion detection" showed mathematically why false alarms dominate
unless detectors are extremely precise.`,
				Exercises: trio(
					"Work it out: 10,000 logins a day, 1 of which is an attack; your detector catches every attack but also flags 0.5% of normal logins. How many alerts a day, and how many are real?",
					"About 50 false alarms plus 1 real one: roughly one real alert in fifty.",
					"Extend the bruteforce.py type-in to read a real SSH or web server log, count distinct user names tried per address, and add a second rule for many different users from one address (password spraying).",
					"Spraying tries one common password on many accounts, so per-account counts stay low.",
					"Run Zeek or Suricata on a packet capture of your own traffic (captured with tcpdump or Wireshark), read the logs or alerts it produces, and explain three of them.",
					"Zeek's conn.log and dns.log are the most readable place to start.",
				),
			},
			{
				Name:    "Detection Engineering with MITRE ATT&CK",
				Summary: "Plan detections around what attackers do, not which tools they use.",
				Body: `Attackers change tools easily but their behaviour less so: after getting
in, they still need to run code, stay in, raise privileges, steal
credentials, move to other machines and take data out. MITRE ATT&CK
catalogues these tactics and hundreds of techniques observed in real
attacks, each with notes on how to detect and mitigate it.

Detection engineering treats detections like software. Pick a technique
that matters for your environment, find the log source that shows it,
write a rule (for example in Sigma, a vendor-neutral rule format),
document why it exists and what to do when it fires, then test it:
trigger the behaviour safely in a lab and confirm the alert appears.
Mapping your detections onto ATT&CK shows where your coverage has gaps.`,
				Diagram: `technique (e.g. new admin account created)
   ─▶ log source (account-change events)
   ─▶ rule (Sigma) + why it matters + what to do
   ─▶ test: create a test account in a lab ─▶ did the alert fire?
   ─▶ coverage map: which techniques are covered, which are not`,
				MentalModel: "A detection is untested code until you have seen it fire in a lab.",
				TryIt:       "Open attack.mitre.org, pick the technique 'Valid Accounts', and read its detection section. Which of your own systems would log it?",
				Analogy: `Burglars change their crowbars, but they all still have to get through
a door or window. Alarm sensors on the doors (behaviours) work against
every crowbar; a list of known crowbar brands (tools) does not.`,
				Example: `MITRE began building ATT&CK in 2013 from observations in its own test
network and released it publicly in 2015. Security teams and vendors
now use it as a common language for detections and threat reports, and
the Sigma project publishes thousands of shared detection rules mapped
to it.`,
				Exercises: trio(
					"Why is 'alert when an account is added to the administrators group' a more durable detection than 'alert when the file bad-tool.exe runs'?",
					"Renaming a file defeats the second; the attacker still needs admin rights.",
					"Write a detection for one technique in your own lab: for example, a new user account or a new scheduled task or cron job. Write it as a Sigma rule or a script over your logs, trigger it with a harmless test, and document what to do when it fires.",
					"Include in the rule's notes how you tested it and what a false positive looks like.",
					"Build a small coverage map: list ten ATT&CK techniques relevant to a small organisation, and for each, the log source you have, whether a detection exists, and whether it has been tested. Pick the biggest gap and close it.",
					"Start with techniques around valid accounts, remote services and exfiltration.",
				),
			},
			{
				Name:    "Security Monitoring & Alert Triage",
				Summary: "Turn alerts into decisions, without burning out the people.",
				Body: `A security operations team watches alerts, usually collected in a SIEM,
and decides for each: harmless, suspicious, or an incident. Good triage
is a routine: gather context (which user, which machine, what else
happened around that time), compare with what is normal for them, and
decide within a set time, following a written runbook for each alert
type.

Alert fatigue is the main enemy: when most alerts are noise, people
start clicking them away, and the real one gets missed. So every alert
must be actionable, have an owner and a runbook, and be reviewed: rules
that fire often without finding anything are tuned or removed.
Automation (SOAR) can gather context and handle routine responses, such
as resetting a password or isolating a laptop, so people spend their
time on judgement.`,
				Diagram: `alert ─▶ enrich (user, machine, recent events) ─▶ compare with normal
      ─▶ decide: benign │ suspicious: investigate │ incident: respond
      ─▶ afterwards: tune the rule if it was noise`,
				MentalModel: "An alert without a runbook and an owner is noise that will be ignored.",
				TryIt:       "Write a five-step runbook for the alert 'login from a new country followed by a password change'. What would you check first, and what would make you escalate?",
				Analogy: `A hospital's emergency department triages patients: a quick, standard
assessment sorts them so the most serious are seen first. If the alarm
on every bed beeped constantly for nothing, nurses would stop
listening, which is exactly what happens with noisy alerts.`,
				Example: `During the 2013 Target breach, the company's malware-detection system
raised alerts that were passed to its security team, but they were not
acted on in time, and around 40 million card numbers were stolen. The
detection worked; the process after the alert did not.`,
				Exercises: trio(
					"An alert fires 300 times a week and has never found a real incident. Give three options, and say which you would try first.",
					"Tune it (narrow the condition), add context to raise its precision, or remove it; do not just ignore it.",
					"Write runbooks for three alerts: many failed logins for one account, a new admin account, and unusual outbound data volume from a server. Each with: what it means, what to check, how to decide, and who to escalate to.",
					"A good runbook lets someone who did not write the rule handle the alert at 3 a.m.",
					"Run a week of monitoring on your own home lab or a server you own: collect alerts from your detections, triage each one with your runbooks, keep a log of decisions and time taken, and tune at least one noisy rule.",
					"Measure precision: the share of alerts that turned out to matter.",
				),
			},
		},
		Resources: []Resource{
			{"Site", "MITRE ATT&CK", "https://attack.mitre.org/", "The catalogue of attacker tactics and techniques, with detection notes."},
			{"Tool", "Zeek documentation", "https://docs.zeek.org/en/master/", "A network monitor that turns traffic into readable logs."},
			{"Tool", "Suricata", "https://suricata.io/", "An open-source intrusion detection and prevention engine."},
			{"Tool", "Wireshark User's Guide", "https://www.wireshark.org/docs/wsug_html_chunked/", "Capture and read network traffic."},
			{"Site", "nftables wiki", "https://wiki.nftables.org/", "The Linux firewall framework, with examples."},
			{"Tool", "Sigma detection rules", "https://github.com/SigmaHQ/sigma", "A vendor-neutral rule format and thousands of shared detections."},
			{"Site", "Security Onion documentation", "https://docs.securityonion.net/", "A free platform for network security monitoring and log management."},
		},
		Blueprints: []Blueprint{
			{"Segmented Home Lab", "A small virtual network with a DMZ, an internal zone and an admin zone, firewall rules between them, and tests that prove the denials.",
				[]string{"Draw the zones", "Firewall rules", "Bastion or VPN with MFA", "Test allowed and denied paths"}},
			{"Central Logging Stack", "Logs from several machines and the firewall collected centrally, protected from deletion, and searchable.",
				[]string{"Log shipping", "Central store", "Retention and access control", "Saved searches for investigations"}},
			{"Detection Pipeline", "Five tested detections mapped to ATT&CK, each with a runbook, running over your lab's logs.",
				[]string{"Choose techniques", "Write rules", "Trigger each in the lab", "Runbooks", "Tune after a week"}},
		},
		Quiz: []Question{
			{"Why filter outgoing traffic as well as incoming?", "So a compromised machine cannot easily connect out to attackers or send data away."},
			{"What is the difference between signature and anomaly detection?", "Signatures match known attack patterns; anomaly detection flags departures from normal behaviour."},
			{"Why do most alerts turn out to be false alarms?", "Real attacks are rare compared with normal events (the base rate), so even accurate detectors flag many harmless events."},
			{"Why send logs to a central store as they happen?", "Attackers often delete local logs; a protected central copy survives."},
		},
	},

	// -----------------------------------------------------------------
	20: {
		Overview: `Sooner or later every organisation has an incident. What separates a
bad day from a disaster is preparation: a plan people have practised,
backups that actually restore, evidence handled carefully enough to
learn what happened, and a culture that fixes causes instead of blaming
people. This stage also covers privacy and data protection: collecting
less personal data, protecting what you keep, and the legal duties that
apply when it leaks.`,
		Outcomes: []string{
			"Write an incident response plan with clear roles, severities and contacts",
			"Contain an incident, rebuild from known-good sources and restore from tested backups",
			"Collect and preserve digital evidence without destroying it",
			"Run a blameless post-mortem that leads to real fixes",
			"Apply data-protection principles and know when a breach must be reported",
		},
		Glossary: []Term{
			{"Incident commander", "The one person who coordinates an incident response and makes the calls."},
			{"Playbook", "Written, step-by-step instructions for handling a particular kind of incident."},
			{"Containment", "Stopping an incident from spreading or doing more damage, before fixing it."},
			{"Chain of custody", "A record of who handled evidence, when and how, so it can be trusted later."},
			{"Order of volatility", "Collect the evidence that disappears fastest (memory, network state) first."},
			{"Post-mortem", "A written review after an incident: what happened, why, and what will change."},
			{"Personal data", "Any information about an identifiable person: a name, an email, a location, an IP address."},
			{"Data minimisation", "Collecting and keeping only the personal data you actually need."},
		},
		Concepts: []Concept{
			{
				Name:    "Incident Response Planning",
				Summary: "Decide who does what before the emergency, and practise it.",
				Body: `In an incident, people are stressed, information is incomplete, and
decisions cannot wait. A plan written in calm times answers the
questions in advance. Who is the incident commander, who handles
communication, who keeps the timeline, who does the technical work? How
are incidents classified by severity, and who must be woken up for each
level? Who are the outside contacts: lawyers, insurers, the data
protection authority, law enforcement, key customers? How do we talk if
email and chat are compromised?

Playbooks cover the likely cases (ransomware, a stolen laptop, leaked
credentials, a compromised account), and tabletop exercises test them:
the team walks through a realistic scenario around a table and finds
the gaps before a real attack does. NIST's guide, SP 800-61, long
described the lifecycle as: prepare, detect and analyse, contain,
eradicate and recover, and learn. Its 2025 revision places incident
response inside NIST's wider Cybersecurity Framework.`,
				Diagram: `incident commander ── decides, coordinates
 ├─ technical lead ── investigates, contains, fixes
 ├─ communications ── staff, customers, regulators
 └─ scribe ────────── timeline: every action, with the time
severity 1: all hands now │ severity 2: working hours │ 3: ticket`,
				MentalModel: "The worst moment to write a plan is during the incident.",
				TryIt:       "For your own devices, write the first five things you would do if your email account were taken over, and the phone numbers or recovery codes you would need, printed on paper.",
				Analogy: `A fire drill. Nobody learns where the exits are while the building is
full of smoke; they learned it in the drill, and the fire marshal
already has a list of who must be accounted for.`,
				Example: `In 2017 the NotPetya malware destroyed the IT systems of the shipping
company Maersk within hours. Its teams rebuilt about 4,000 servers and
45,000 PCs in around ten days, helped by the one domain controller that
survived, in Ghana, because a power cut had taken it offline during the
attack. Luck saved them; a plan and offline backups would not have
needed it.`,
				Exercises: trio(
					"Name the four roles in an incident team and say why the incident commander should usually not do technical work.",
					"Someone must keep the overview and make decisions while others are deep in the details.",
					"Write an incident response plan for a small organisation (or your own home lab): roles, severity levels with examples, a contact list, how to communicate if email is down, and one full playbook for ransomware.",
					"Keep a printed copy: in a real incident, the plan's file server may be encrypted.",
					"Run a 30-minute tabletop exercise with a friend or classmate using your plan and a scenario such as 'a laptop with customer data was stolen from a car'. Record every question the plan could not answer, and update it.",
					"The facilitator reveals new facts every few minutes: 'the laptop was not encrypted'.",
				),
			},
			{
				Name:    "Containment, Eradication & Recovery",
				Summary: "Stop the spread, remove the attacker, and restore from known-good sources.",
				Body: `Containment comes first: isolate affected machines from the network
(keep them powered on for evidence), disable compromised accounts, and
block the attacker's addresses and domains. Then eradicate: find how
they got in and close it, remove their persistence (new accounts,
scheduled tasks, backdoors), and rotate every credential they may have
seen, including service accounts and API keys.

Recovery means rebuilding from known-good sources, not cleaning a
compromised machine and hoping: reinstall from trusted images and
restore data from backups taken before the compromise. Backups only
count if they restore: follow the 3-2-1 rule (three copies, on two kinds
of storage, one off-site), keep at least one copy offline or immutable
so ransomware cannot encrypt it, and test restores regularly. Watch
closely afterwards; attackers often try to return.`,
				Diagram: `contain:   isolate hosts, disable accounts, block addresses
eradicate: close the way in, remove backdoors, rotate secrets
recover:   rebuild from clean images, restore pre-attack backups
watch:     extra monitoring for the attacker returning`,
				MentalModel: "You cannot trust a machine an attacker controlled; rebuild it.",
				TryIt:       "Pick one important folder of yours, restore it from your backup to a different place, and check that the files open. If you cannot, you do not have a backup yet.",
				Analogy: `A kitchen with food poisoning: close it (contain), find and throw out
the contaminated stock and fix the broken fridge (eradicate), restock
from a trusted supplier (recover), and have inspectors visit more often
for a while (watch).`,
				Example: `In 2021 ransomware in Colonial Pipeline's business systems led it to
shut down the largest fuel pipeline on the US East Coast for several
days. The attackers had got in with the password of an old VPN account
that had no multi-factor authentication. In 2023 a ransomware attack
took the British Library's systems offline for months, and it later
published a detailed report on what went wrong so others could learn.`,
				Exercises: trio(
					"Why keep an infected machine powered on but disconnected, rather than switching it off?",
					"Memory holds evidence (running processes, network connections, keys) that is lost at power-off.",
					"Set up 3-2-1 backups for a machine or project you care about, with one copy that is offline or immutable, then do a full test restore to a different machine and time it.",
					"The restore time is your real recovery time; write it down.",
					"Write a ransomware playbook for a small business: the first hour minute by minute, how to decide what to rebuild, the order of recovery by business importance, and how to check backups are clean before restoring.",
					"Restore identity systems (accounts, passwords) first: everything else depends on them.",
				),
			},
			{
				Name:    "Digital Forensics Basics",
				Summary: "Collect evidence carefully enough to trust what it tells you.",
				Body: `Forensics answers what happened, when, and how. The first rule is to
avoid destroying what you are trying to study. Collect in order of
volatility: memory and network connections first (they vanish at
power-off), then running processes, then disk, then logs and backups
stored elsewhere (RFC 3227 describes this). Work on copies: take a
full image of a disk with a write blocker, record its hash, and analyse
the copy, so you can prove later that nothing changed.

Keep a chain of custody: who collected each item, when, how it was
stored and who handled it since. Build a timeline by merging file
times, log entries and network records, all converted to one time
zone. Free tools such as The Sleuth Kit and Autopsy (disks) and
Volatility (memory) do the heavy lifting. If the incident may end up in
court, involve professionals early.`,
				Diagram: `most volatile  ▲ memory, network connections
               │ running processes, logged-in users
               │ disk contents (image it, hash it)
least volatile │ remote logs, backups, archives`,
				MentalModel: "Every action you take changes the evidence; copy first, analyse the copy.",
				TryIt:       "Type in and run this stage's integrity.py, then change one byte of a file's content and predict its new fingerprint's first characters before running again.",
				Analogy: `Detectives at a crime scene photograph everything before touching
anything, wear gloves, and log every item into an evidence bag with a
label, so that months later the court can trust it has not been swapped
or contaminated.`,
				Example: `In 2005 the BTK serial killer in Kansas was identified partly because a
floppy disk he sent to the police held a deleted Word document whose
hidden metadata named a local church and the user "Dennis". RFC 3227,
published in 2002, is still the standard short guide to collecting
evidence in the right order.`,
				Exercises: trio(
					"Put these in the order you would collect them: a copy of the web server's disk, the current network connections, the cloud provider's access logs, the contents of memory.",
					"Memory and connections first: they disappear when the machine is switched off.",
					"Extend the integrity.py type-in to scan a real folder you own: store the baseline in a JSON file, then report new, modified and deleted files on each run. Test it by changing files yourself.",
					"Hash files in chunks so large files do not fill memory.",
					"Investigate your own machine: build a timeline of one afternoon from its logs (logins, installed software, browser history, file changes), all in one time zone, and write a one-page report of what happened when.",
					"On Linux, journalctl --since and find -newermt are a good start; on Windows, Event Viewer.",
				),
			},
			{
				Name:    "Blameless Post-Mortems",
				Summary: "Learn from every incident by fixing systems, not blaming people.",
				Body: `After an incident, write down what happened while memories are fresh:
the timeline, the impact, how it was detected and resolved, what went
well and what did not. Look for contributing factors, not one culprit:
the missing check, the confusing tool, the alert nobody owned, the
backup nobody had tested.

Blameless means assuming people acted reasonably given what they knew
and the tools they had. If one person's mistake could cause an outage,
the system allowed it, and that is what to fix. Blame makes people hide
mistakes, which destroys exactly the information you need. Every
post-mortem ends with action items, each with an owner and a date, and
someone checks later that they were done. Many companies publish their
post-mortems, and reading them is one of the fastest ways to learn.`,
				Diagram: `timeline ─▶ impact ─▶ contributing factors (not "who")
         ─▶ what went well ─▶ what went badly
         ─▶ action items: owner + date ─▶ follow-up review`,
				MentalModel: "Ask 'what made this mistake easy to make?', never 'who made it?'.",
				TryIt:       "Read one public post-mortem (Cloudflare's blog publishes many) and list its contributing factors and action items.",
				Analogy: `Air accident investigation. Investigators look for every factor behind
a crash (design, training, weather, procedures) rather than only blaming
the pilot, and aviation became extraordinarily safe because every
lesson is shared across the whole industry.`,
				Example: `In 2017 a GitLab engineer accidentally deleted a production database
while working late during an incident. Five different backup methods
turned out not to work, and about six hours of data were lost. GitLab
streamed the recovery live and published a blameless post-mortem that
focused on the broken backups, not on the engineer.`,
				Exercises: trio(
					`Rewrite this finding blamelessly: "Sam deployed the wrong config and caused the outage."`,
					"Say what allowed it: no review step, no config validation, no staged rollout.",
					"Write a post-mortem for a real failure you experienced (a lost file, a broken project, a missed deadline): timeline, impact, contributing factors, and three action items with dates.",
					"Look for at least three contributing factors; there is almost never only one.",
					"Read three published post-mortems from different companies and write a one-page comparison: common contributing factors, which action items look most effective, and one practice you will adopt.",
					"GitLab, Cloudflare and Google have all published detailed ones.",
				),
			},
			{
				Name:    "Privacy & Data Protection",
				Summary: "Collect less, protect what you keep, and know your legal duties.",
				Body: `Personal data is any information about an identifiable person, and
the safest personal data is the data you never collected. The EU's
General Data Protection Regulation (GDPR), a model for laws in many
countries, sets principles: a lawful reason for each use, collecting
only what is needed for a stated purpose (data minimisation), keeping
it accurate and no longer than needed, and securing it. People have
rights to see, correct and delete their data.

In engineering terms: design for privacy from the start, record what
personal data you hold and why, set retention periods and actually
delete, restrict access, and encrypt. Pseudonymisation (replacing names
with IDs) reduces risk but is still personal data; true anonymisation
is much harder than it looks, because combining a few attributes often
identifies a person. Under GDPR, a personal-data breach must usually be
reported to the regulator within 72 hours of becoming aware of it.`,
				Diagram: `collect only what you need ─▶ record why ─▶ protect it
   ─▶ keep only as long as needed ─▶ delete (backups too)
breach? ─▶ assess risk ─▶ tell the regulator within 72 h
       ─▶ tell the people affected if the risk to them is high`,
				MentalModel: "Data you do not have cannot leak.",
				TryIt:       "Ask one service you use for a copy of all the data it holds about you (most have a download-your-data page) and look through it. What surprises you?",
				Analogy: `A doctor's surgery keeps only the records it needs to treat you, locks
the filing cabinet, shreds old files on schedule, and must tell you and
the authorities if a folder goes missing.`,
				Example: `In 2023 Ireland's data protection regulator fined Meta €1.2 billion
for transferring Europeans' Facebook data to the US without adequate
protection. In 2008 researchers showed that the "anonymous" movie
ratings Netflix had released for a competition could be linked back to
real people using public reviews: removing names is not anonymisation.
This academy keeps your progress only on your own computer and sends
nothing anywhere.`,
				Exercises: trio(
					"Which of these are personal data: an email address, a customer ID number, the average age of all customers, an IP address, a photo of a crowd?",
					"If it can be linked to one person, even indirectly, it is personal data.",
					"Take a small app you built and write its data inventory: every personal field, why it is collected, where it is stored, who can access it, and when it is deleted. Remove one field you do not need and add a delete-my-account feature.",
					"Remember logs, backups and analytics: personal data hides in all of them.",
					"Write a breach-response checklist for a small online shop under GDPR (or your country's law): how to assess the risk, what to tell the regulator within 72 hours, when and how to tell customers, and a template notification email.",
					"Your national data protection authority publishes breach-reporting guidance and forms.",
				),
			},
		},
		Resources: []Resource{
			{"Paper", "NIST SP 800-61 Rev. 3: Incident Response Recommendations", "https://csrc.nist.gov/pubs/sp/800/61/r3/final", "The current US guide to incident response."},
			{"Paper", "RFC 3227: Guidelines for Evidence Collection and Archiving", "https://www.rfc-editor.org/rfc/rfc3227", "A short, classic guide to collecting evidence in the right order."},
			{"Tool", "Autopsy and The Sleuth Kit", "https://www.sleuthkit.org/autopsy/", "Free, open-source disk forensics."},
			{"Tool", "Volatility", "https://volatilityfoundation.org/", "The open-source framework for memory forensics."},
			{"Book", "Google SRE book: Postmortem Culture", "https://sre.google/sre-book/postmortem-culture/", "How Google runs blameless post-mortems."},
			{"Article", "GitLab's 2017 database outage post-mortem", "https://about.gitlab.com/blog/2017/02/10/postmortem-of-database-outage-of-january-31/", "A candid, blameless account of a real data-loss incident."},
			{"Site", "GDPR: the official text", "https://eur-lex.europa.eu/eli/reg/2016/679/oj", "The EU General Data Protection Regulation."},
		},
		Blueprints: []Blueprint{
			{"Incident Response Kit", "An incident response plan, three playbooks, a contact sheet and a post-mortem template, tested in a tabletop exercise.",
				[]string{"Roles and severities", "Playbooks: ransomware, stolen laptop, leaked key", "Printed contact sheet", "Tabletop exercise", "Updated plan"}},
			{"Tested Backups", "3-2-1 backups for a real project, with an immutable or offline copy and a monthly restore test with timings.",
				[]string{"Three copies, two media, one off-site", "Offline or immutable copy", "Automated schedule", "Restore test log"}},
			{"Forensics Lab", "A small virtual machine you misconfigure and 'attack' yourself, then investigate: memory capture, disk image, hashes, timeline and report.",
				[]string{"Capture memory", "Image and hash the disk", "Build a timeline", "Write the report", "Chain-of-custody log"}},
		},
		Quiz: []Question{
			{"Why should the incident commander usually not do hands-on technical work?", "They must keep the overview, coordinate people and make decisions."},
			{"What is the 3-2-1 backup rule?", "Three copies of the data, on two kinds of storage, with one copy off-site (and ideally one offline or immutable)."},
			{"Why collect memory before imaging the disk?", "Memory is the most volatile evidence: it is lost when the machine is switched off."},
			{"Within how long must a personal-data breach usually be reported to the regulator under GDPR?", "Within 72 hours of becoming aware of it."},
		},
	},
}
