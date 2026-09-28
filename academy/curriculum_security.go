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
from outside, memory-safe habits, reviewing code with an attacker's eye,
and automated tools that look for trouble on every change.`,
		Outcomes: []string{
			"Apply classic secure-design principles such as least privilege and fail-safe defaults",
			"Handle untrusted input safely: canonicalise, validate against an allowlist, encode on output",
			"Write C that is hard to exploit, and know when to choose a memory-safe language",
			"Review a change for security problems with a repeatable checklist",
			"Add static analysis, fuzzing and dependency scanning to a CI pipeline",
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
					"Take a small program that reads a configuration file and runs with defaults if the file is missing. Change it so that a missing, unreadable or invalid security setting makes it refuse to start, with a clear error.",
					"Fail closed: a missing setting must never fall back to 'allow everything'.",
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
				MentalModel: "Check the canonical form, not the spelling the attacker chose.",
				TryIt:       "Type in and run this stage's safe_join program, then add a test name of your own that you expect to be refused.",
				Analogy: `A post room that X-rays every parcel. First it unwraps the parcel to see
what is really inside (canonicalise), then compares it with the list of
things the company ordered (allowlist), and only then passes it on,
relabelled for the right department (encode).`,
				Example: `In 2021 a change to path handling in Apache HTTP Server 2.4.49
(CVE-2021-41773) let requests with encoded ../ sequences escape the web
folder and read other files on the server. The first fix, 2.4.50, was
incomplete (CVE-2021-42013), and a full fix arrived in 2.4.51. This
academy's own resource importer refuses control codes and bidirectional
text markers, which could otherwise disguise what you are importing.`,
				Exercises: trio(
					`Why must a program normalise "files/a/../../etc/passwd" before checking that it starts with "files/"?`,
					"The unnormalised string does start with files/, yet it points outside.",
					"Write safe_join(base, name) in your language with tests for: a normal name, a nested name, ../ escapes, an absolute path, an empty name and a very long name. Then extend it to refuse symbolic links that point outside the folder.",
					"Resolve the real path (realpath) after joining, then compare it with the real path of the base folder.",
					"Choose one program you wrote that accepts input (a form, a CLI, a file importer). Draw its trust boundaries, list every input, and for each write the allowlist rule, the size limit and the output encoding. Fix at least one gap you find.",
					"Include the inputs people forget: environment variables, file names inside archives, headers.",
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
					"Write a one-page security review checklist for your own projects, then use it to review one of your older programs and record every finding with its severity.",
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
		},
	},
}
