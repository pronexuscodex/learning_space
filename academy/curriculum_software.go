package main

// Track S: Software, Security & Theory (stages 9–12). How professionals
// build and protect software, what computers can and cannot do in
// principle, and the ideas behind every programming language.

var softwareGuides = map[int]StageGuide{
	// -----------------------------------------------------------------
	9: {
		Overview: `Writing code that works once is programming. Writing code that keeps
working as it grows, alongside other people, for years, is software
engineering. This stage covers the everyday professional toolkit: the
command line, Git, testing, design and shipping to real users.`,
		Outcomes: []string{
			"Work comfortably in a terminal and automate chores with scripts",
			"Collaborate through Git and pull requests like a professional team",
			"Test, structure and deploy software that others can rely on",
		},
		Glossary: []Term{
			{"Terminal", "A text window where you type commands to the computer."},
			{"Shell", "The program that reads and runs your terminal commands (bash, zsh, PowerShell)."},
			{"Repository", "A project folder tracked by Git, including its full history."},
			{"Commit", "A saved snapshot of a project, with a message describing the change."},
			{"Branch", "An independent line of development in Git."},
			{"Unit test", "A small automated check of one function's behaviour."},
			{"CI", "Continuous Integration: automatically building and testing every change."},
			{"Container", "A package holding an app plus everything it needs to run, so it runs the same everywhere."},
			{"Refactoring", "Improving code's structure without changing what it does."},
		},
		Concepts: []Concept{
			{
				Name:    "The Command Line & Your Environment",
				Summary: "The terminal is the programmer's workshop.",
				Body: `A shell (bash, zsh, PowerShell) runs programs when you type commands.
The file system is a tree: a path such as /home/ana/projects is
absolute, and ./notes.txt is relative to where you are. The core
commands are ls, cd, mkdir, cp, mv, rm, cat, less, grep (search text)
and find.

Programs read standard input and write standard output, so the pipe |
chains small tools together: cat log.txt | grep ERROR | wc -l counts the
error lines. Environment variables such as PATH configure programs, and
package managers (apt, brew, winget, pip, npm) install software
reproducibly.`,
				MentalModel: "Small tools, joined by pipes, do big jobs.",
				TryIt:       "Use one pipeline to find the 10 most common words in a text file.",
				Analogy: `A graphical interface is a menu with pictures. The command line is
talking to the chef directly: less hand-holding, but you can ask for
exactly what you want, and write the request down to repeat it tomorrow
(a script).`,
				Example: `Cloud servers, Docker containers and CI systems are all driven from the
command line. Operations engineers automate deployments with shell
scripts, and data scientists use grep, sort and uniq to explore gigabyte
logs that no spreadsheet can even open.`,
				Exercises: trio(
					`Using only commands, create a folder "practice", make three empty files in it, list them, rename one and delete another.`,
					"mkdir, touch (New-Item on Windows), ls, mv and rm.",
					"Write a shell script that backs up a folder into a dated archive (such as backup-2026-09-24.tar.gz) and keeps only the 5 newest backups.",
					"$(date +%F), tar -czf, and ls -t | tail -n +6 to find the old ones.",
					"Get a real web-server access log (many sample logs are published online) and, with a single pipeline, list the top 10 IP addresses and count the 404 errors.",
					"awk '{print $1}' | sort | uniq -c | sort -rn | head",
				),
			},
			{
				Name:    "Version Control with Git",
				Summary: "A time machine and collaboration tool for your code.",
				Body: `Git records snapshots of your project called commits. Each has a
message, an author and a pointer to its parent, forming a history you
can browse and restore. Branches are cheap, movable labels, so you can
try an idea without touching the main line. Merging combines branches,
and a conflict happens when two people changed the same lines.

Remotes (GitHub, GitLab) share repositories, and pull requests let
teammates review changes before they are merged. Good habits: small
commits with clear messages, and never commit passwords or keys.`,
				Diagram: `main:     A ── B ─────────── E   (merge commit)
                ╲           ╱
feature:         C ──── D ─╯`,
				MentalModel: "Commit early, commit often, and write messages for your future self.",
				TryIt:       "Create a repository, make a branch, change the same line on both branches, and resolve the merge conflict.",
				Analogy: `Save points in a video game, plus parallel universes. You can branch off
to try a risky strategy, and if it works, bring it back into the main
timeline.`,
				Example: `Linus Torvalds wrote Git in 2005 to manage development of the Linux
kernel, which now has thousands of contributors per release. Today
almost every software company uses Git, and GitHub reports over 100
million developers.`,
				Exercises: trio(
					"Explain in your own words the difference between git add, git commit and git push.",
					"Staging area → local history → remote server.",
					"Put one of your earlier projects in a Git repository, make at least 5 meaningful commits, and use git log, git diff and git switch --detach to travel back to an old version and return.",
					"git log --oneline --graph shows the history compactly.",
					`Make an open-source contribution: find a project with a "good first issue" or a typo in its documentation, fork it, fix it on a branch and open a pull request.`,
					"Read the project's CONTRIBUTING file first and follow its conventions.",
				),
			},
			{
				Name:    "Testing & Reliability",
				Summary: "Code you have not tested is code that does not work yet.",
				Body: `A unit test calls one function with known inputs and checks the output.
Integration tests check that pieces work together, and end-to-end tests
drive the whole app as a user would. Good tests cover normal cases, edge
cases (empty input, zero, huge values) and error cases.

Automated test suites run on every change through continuous integration
(CI), so bugs are caught before users see them. Test-driven development
writes the failing test first. Property-based testing generates hundreds
of random inputs to find cases you did not think of.`,
				MentalModel: "A test is a claim about your code that a machine re-checks forever.",
				TryIt:       "Write five tests for is_palindrome, including an empty string and a single character.",
				Analogy: `A pilot's pre-flight checklist: the same checks every time, however
experienced the pilot, because skipping one costs too much.`,
				Example: `In 1999 NASA lost the Mars Climate Orbiter because one team's software
produced pound-force seconds while another expected newton-seconds: an
interface mismatch that an integration check could have caught. In 2012
Knight Capital lost about $440 million in 45 minutes after old code was
accidentally reactivated during a deployment.`,
				Exercises: trio(
					"For a function that computes a person's age from their birth date, list 8 test cases, including tricky ones.",
					"A birthday today, 29 February, a date in the future, very old dates.",
					"Write that age function test-first: write one failing test, make it pass, and repeat until all 8 cases pass.",
					"Use pytest, or your language's built-in test tool.",
					"Set up continuous integration for one of your GitHub repositories (for example with GitHub Actions) so the tests run on every push, and add a status badge to the README.",
					"A workflow file in .github/workflows/ that checks out the code and runs the tests.",
				),
			},
			{
				Name:    "Software Design & Abstraction",
				Summary: "Managing complexity so programs stay understandable as they grow.",
				Body: `Complexity is the main enemy of large software. Abstraction hides
details behind simple interfaces: callers of save_order() should not
need to know about SQL. Modules with a lot of functionality behind a
small interface are easier to use and to change.

Useful principles: separate concerns (interface, logic, storage), avoid
duplication (but do not abstract too early), keep coupling low and
cohesion high, and prefer clear names to clever code. Design patterns
are named solutions to recurring problems (observer, strategy, adapter).
Refactoring improves structure without changing behaviour, and it is
safe because tests exist.`,
				Diagram: `┌──────────────────────────────────┐
│ interface / API     (what)       │
├──────────────────────────────────┤
│ business logic      (rules)      │
├──────────────────────────────────┤
│ storage / network   (how)        │
└──────────────────────────────────┘
each layer only talks to the layer below it`,
				MentalModel: "A good module is a deep box with a small lid.",
				TryIt:       "Refactor a 100-line script into functions for input, processing and output, keeping its behaviour identical.",
				Analogy: `A car: the driver uses a steering wheel and pedals (the interface) and
never needs to understand fuel injection (the implementation). A
mechanic can swap the engine for an electric motor and the driver's
interface barely changes.`,
				Example: `This very program separates storage (main.go), input (console.go),
display (ui.go) and teaching content (curriculum files). Amazon famously
split its systems into services with strict interfaces so that thousands
of teams could work independently.`,
				Exercises: trio(
					"Pick an app you use daily and sketch its likely modules (login, feed, messaging, storage). Which ones need to know about each other?",
					"Draw boxes and arrows; fewer arrows is better.",
					"Restructure your bill splitter or budget tracker so that the calculation logic has no print() or input() calls, then write unit tests for that logic alone.",
					"Pure functions take data in and return data; the interface layer calls them.",
					"Your budget tool reads CSV files from your bank. Design it so that supporting a second bank's different CSV format means writing one small new component, not changing the rest. Implement both.",
					"The adapter pattern: each bank gets a parser that outputs the same Transaction structure.",
				),
			},
			{
				Name:    "From Code to Production",
				Summary: "Shipping software that people can rely on.",
				Body: `Real software must be built, packaged, deployed and kept running.
Dependencies are pinned in lockfiles so builds are reproducible.
Containers (Docker) bundle an app with its environment, so "it works on
my machine" works everywhere. CI/CD pipelines test and deploy every
change automatically.

In production you need logs (what happened), metrics (how much, how
fast) and alerts (tell a human when something is wrong). Reliable teams
use code review, gradual rollouts, feature flags, backups whose restore
has actually been tested, and blameless post-mortems after incidents.`,
				MentalModel: "If it is not monitored, you do not know it works; if the backup is not tested, you do not have one.",
				TryIt:       "Put a tiny web app in a Docker container and run it locally.",
				Analogy: `A restaurant is more than a recipe. It needs suppliers (dependencies),
a standard kitchen setup (containers), health inspections (tests and
review), and a manager watching how the evening is going (monitoring).`,
				Example: `In 2017 GitLab accidentally deleted a production database and discovered
that several of its backup methods had silently failed. It livestreamed
the recovery and published an unusually honest post-mortem. In July 2024
a faulty CrowdStrike update crashed about 8.5 million Windows machines
worldwide, which is why staged rollouts matter.`,
				Exercises: trio(
					`List 10 things that could go wrong between "it works on my laptop" and "it works for 10,000 users".`,
					"Think about configuration, load, data, network, versions and people.",
					"Write a Dockerfile for one of your programs, build the image and run it. Then add a CI workflow that builds the image on every push.",
					"Start FROM an official language image and COPY your code in.",
					"Deploy a small web app to a free hosting tier, add a health-check endpoint and uptime monitoring that emails you when it goes down, then break it on purpose and confirm that the alert arrives.",
					"Many free uptime monitors can check a URL every few minutes.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "The Missing Semester of Your CS Education (MIT)", "https://missing.csail.mit.edu/", "Shell, editors, Git, debugging and more; exactly what courses skip."},
			{"Book", "Pro Git (free)", "https://git-scm.com/book/en/v2", "The complete Git book."},
			{"Tool", "Learn Git Branching", "https://learngitbranching.js.org/", "Interactive, visual Git exercises."},
			{"Site", "Refactoring.Guru", "https://refactoring.guru/", "Design patterns and refactorings, explained with pictures."},
			{"Book", "Google's Site Reliability Engineering books (free)", "https://sre.google/books/", "How Google runs production systems."},
			{"Site", "The Twelve-Factor App", "https://12factor.net/", "A short checklist for building deployable web services."},
		},
		Blueprints: []Blueprint{
			{"Dotfiles & Automation Kit", "A Git repository of your shell configuration and scripts that sets up a new machine with one command.",
				[]string{"Shell config and aliases", "Useful scripts", "Install script", "Test on a fresh VM or container"}},
			{"Production-Ready To-Do API", "A small web API with tests, CI, a Docker image, logging, a health check and a real deployment.",
				[]string{"CRUD endpoints", "Unit and integration tests", "CI pipeline", "Docker image", "Deploy and monitor"}},
			{"Open-Source Contribution", "Land three merged pull requests in real open-source projects.",
				[]string{"Find a project", "Fix documentation", "Fix a small bug", "Add a test or feature"}},
		},
		Quiz: []Question{
			{"What does the pipe | do in a shell?", "It sends the standard output of one command into the standard input of the next."},
			{"What is a Git merge conflict?", "Two branches changed the same lines differently, so Git needs a human to choose the result."},
			{"Why run tests in CI and not only on your laptop?", "Every change gets checked automatically in a clean environment, before it reaches users."},
			{"What is the difference between refactoring and rewriting?", "Refactoring changes structure in small steps without changing behaviour, protected by tests; rewriting replaces the code wholesale."},
		},
	},

	// -----------------------------------------------------------------
	10: {
		Overview: `Security is about keeping systems doing what they should, even when
someone clever is trying to make them do something else. This stage
teaches how attackers think, the most common vulnerabilities, and the
cryptography that protects every password, message and payment. Only
ever practise attacks on systems you own or that explicitly invite it.`,
		Outcomes: []string{
			"Threat-model an app and prioritise its biggest risks",
			"Find and fix common web vulnerabilities (injection, XSS, broken access control)",
			"Use hashing, encryption and signatures correctly, without inventing your own crypto",
		},
		Glossary: []Term{
			{"Vulnerability", "A weakness that an attacker can exploit."},
			{"Threat model", "A structured list of what you protect, from whom, and how it could be attacked."},
			{"Authentication", "Proving who you are (password, key, fingerprint)."},
			{"Authorization", "Deciding what an authenticated user is allowed to do."},
			{"Hash", "A one-way fingerprint of data; the same input always gives the same output."},
			{"Salt", "A random value added to a password before hashing so equal passwords get different hashes."},
			{"Encryption", "Scrambling data so only someone with the key can read it."},
			{"Certificate", "A signed document binding a public key to a name, such as a website domain."},
			{"Phishing", "Tricking people into revealing secrets through fake messages or websites."},
		},
		Concepts: []Concept{
			{
				Name:    "Thinking Like an Attacker: Threat Modelling",
				Summary: "Ask what you are protecting, from whom, and how it could fail.",
				Body: `Security goals are often summarised as the CIA triad: confidentiality
(only the right people can read), integrity (nobody can tamper
undetected) and availability (the system keeps working). Threat modelling
asks: what are the assets, who are the attackers, what can they reach
(the attack surface), and what is the worst that could happen?

Principles: least privilege (give each part only the access it needs),
defence in depth (several independent layers), fail securely, and never
rely on keeping the design secret. The weakest link is often human:
phishing and reused passwords cause a large share of breaches.`,
				MentalModel: "Assume the attacker knows your system, and protect what matters most first.",
				TryIt:       "Write a threat model for your own email account: assets, attackers, entry points and defences.",
				Analogy: `Securing a house: locks on the doors (authentication), a safe for
valuables (encryption), cameras (logging), and not giving every
tradesperson a master key (least privilege). A burglar only needs one
open window.`,
				Example: `The 2013 Target breach began with credentials stolen from a heating and
air-conditioning contractor and ended with about 40 million payment
cards stolen. The 2017 Equifax breach exposed data on around 147 million
people through an unpatched web framework (Apache Struts).`,
				Exercises: trio(
					"For a school's online grade system, list three threats each against confidentiality, integrity and availability.",
					"Who would want to read grades, change them, or knock the system offline?",
					"Turn on two-factor authentication for your important accounts, check your email address on haveibeenpwned.com, and replace every reused password with one generated by a password manager.",
					"Start with your email account, because it can reset everything else.",
					"Write a one-page threat model for a small online shop (customer accounts, payments, admin panel): assets, attackers, the top 5 risks ranked by likelihood × impact, and one mitigation for each.",
					"Payment data and admin access are the crown jewels.",
				),
			},
			{
				Name:    "Web Vulnerabilities: Injection & XSS",
				Summary: "Never mix untrusted data with code.",
				Body: `Most attacks happen when user input is treated as code. In SQL
injection, a query built as "SELECT * FROM users WHERE name = '" + input
+ "'" lets an attacker type ' OR '1'='1 and read every row. The fix is
parameterised queries.

In cross-site scripting (XSS), showing user text on a page without
escaping it lets attackers run JavaScript in other users' browsers. The
fix is escaping output, plus a Content Security Policy. Other classics
are CSRF (tricking a logged-in browser into sending a request), broken
access control (changing ?user_id=5 to 6 shows someone else's data) and
insecure deserialisation. The OWASP Top 10 lists the most common.`,
				Diagram: `intended:  SELECT * FROM users WHERE name = 'alice'
input:     ' OR '1'='1
injected:  SELECT * FROM users WHERE name = '' OR '1'='1'   ◀ always true`,
				MentalModel: "Data is data and code is code; keep them in separate channels.",
				TryIt:       "Complete the SQL injection labs on PortSwigger's free Web Security Academy.",
				Analogy: `A cheque that reads "Pay ____ to the bearer". If you let a stranger fill
the blank with "£10, and also everything else in the account", a naive
bank clerk reads it as an instruction. A parameterised query is a form
where the blank can only ever hold an amount.`,
				Example: `The 2015 TalkTalk breach in the UK exposed the data of about 157,000
customers through SQL injection, and the company was fined £400,000. The
"Samy" XSS worm (2005) spread to over a million MySpace profiles in
under a day.`,
				Exercises: trio(
					`Explain why escaping "<script>" as "&lt;script&gt;" stops XSS.`,
					"The browser then displays the text instead of running it.",
					"Build a tiny login form backed by SQLite that is deliberately vulnerable to SQL injection, exploit it yourself, then fix it with a parameterised query and show that the exploit no longer works.",
					`cursor.execute("SELECT … WHERE name = ?", (name,))`,
					"Run OWASP Juice Shop (a deliberately insecure online shop) locally and complete five challenges. For each, write down the vulnerability, how you exploited it, and how to fix it.",
					"Juice Shop starts with a single Docker command and has a built-in score board.",
				),
			},
			{
				Name:    "Hashing & Password Storage",
				Summary: "Store fingerprints, not passwords.",
				Body: `A cryptographic hash such as SHA-256 turns any input into a fixed-size
fingerprint. It is one-way (you cannot get the input back) and
collision-resistant (you cannot find two inputs with the same hash).

Websites should never store passwords, only hashes. But fast hashes are
dangerous, because attackers can try billions of guesses per second.
Proper password storage uses a deliberately slow, salted hash designed
for the job: bcrypt, scrypt or Argon2. The salt is a random value per
user, so identical passwords get different hashes and precomputed
"rainbow tables" are useless. Hashes also verify downloads, and they
underpin Git and blockchains.`,
				Diagram: `"hunter2" + salt "x9Qa" ──▶ [ Argon2, slow on purpose ] ──▶ $argon2id$…Kq7
login: hash the typed password with the same salt, then compare`,
				MentalModel: "Make guessing expensive for attackers but cheap enough for honest logins.",
				TryIt:       `Hash "hello" and "hello!" with SHA-256 and compare: one character changes everything.`,
				Analogy: `A fingerprint identifies a person, but you cannot rebuild the person from
it. Salting is like each bank adding its own secret smudge pattern to
every fingerprint, so a list stolen from one bank is useless at another.`,
				Example: `In 2012 LinkedIn stored passwords as unsalted SHA-1 hashes. After a
breach most were cracked, and in 2016 a dump of about 117 million
credentials was offered for sale. Modern frameworks such as Django use
salted, slow password hashes by default.`,
				Exercises: trio(
					`Why is storing SHA-256(password) without a salt a bad idea when two users both choose "password123"?`,
					"Their hashes are identical, so cracking one cracks both.",
					"Write a sign-up and login program that stores users in a JSON file with bcrypt or Argon2 hashes. Confirm that the same password produces different stored hashes for two users.",
					"pip install bcrypt; then bcrypt.hashpw and bcrypt.checkpw.",
					"Verify a real download: take a Linux ISO or any software release that publishes SHA-256 checksums, compute the checksum yourself and compare. Explain which attack this protects against and which it does not.",
					"sha256sum (Linux), shasum -a 256 (macOS) or Get-FileHash (Windows).",
				),
			},
			{
				Name:    "Encryption: Symmetric & Public-Key",
				Summary: "Keeping secrets over an open network.",
				Body: `Symmetric encryption (AES) uses the same secret key to lock and unlock.
It is fast, but both sides need the key. Public-key cryptography (RSA,
elliptic curves) gives each person a key pair: anyone can encrypt with
your public key, but only your private key decrypts.

Key exchange (Diffie–Hellman) lets two strangers agree on a shared secret
over a public channel. Digital signatures prove who sent a message and
that it was not altered. Real systems combine these: public-key methods
to agree on a key, then fast symmetric encryption for the data. Never
invent your own cryptography; use well-reviewed libraries.`,
				Diagram: `Alice                                       Bob
  │ encrypt with Bob's PUBLIC key              │
  │ ────────────── ciphertext ───────────────▶ │ decrypt with Bob's PRIVATE key
  │                                            │
  │ ◀─────────── message + signature ──────────│ signed with Bob's PRIVATE key
  │ verify with Bob's PUBLIC key               │`,
				MentalModel: "Public key locks, private key unlocks; private key signs, public key verifies.",
				TryIt:       "Implement a Caesar cipher and break it by trying all 26 keys, and see why real ciphers need enormous key spaces.",
				Analogy: `A padlock you hand out freely (the public key): anyone can snap it shut
on a box, but only you hold the key that opens it (the private key). A
signature is like a wax seal that only you can make but everyone can
recognise.`,
				Example: `WhatsApp and Signal use end-to-end encryption (the Signal Protocol), so
even the companies cannot read your messages. Phone software updates are
digitally signed, so your phone refuses any update the manufacturer did
not sign.`,
				Exercises: trio(
					`A Caesar cipher shifted each letter by 3. Decrypt "KHOOR ZRUOG".`,
					"Shift each letter back by 3.",
					"Break a Vigenère cipher: encrypt a long English paragraph with a 5-letter key, then recover the key using letter-frequency analysis on every 5th letter.",
					"In English, 'E' is the most common letter.",
					`Generate an SSH key pair and use it to push to GitHub (or log in to a server) without a password. Then sign a Git commit and see GitHub mark it "Verified".`,
					"ssh-keygen -t ed25519, then add the .pub file in your GitHub settings.",
				),
			},
			{
				Name:    "TLS & Authentication on the Web",
				Summary: "The padlock in your browser, and proving who you are.",
				Body: `HTTPS is HTTP inside TLS. When you connect, the server presents a
certificate: its public key, signed by a certificate authority (CA) that
your browser already trusts. The browser verifies the chain of
signatures and checks that the name matches. Then both sides run a key
exchange to agree on symmetric keys, which protect the confidentiality
and integrity of the whole session.

Authentication proves who a user is: a password plus a second factor (an
authenticator app or a hardware security key), or passkeys, which use
public-key cryptography so there is no password to steal or phish.
Sessions and tokens (cookies, JWTs) remember that you logged in, and
must be protected too.`,
				MentalModel: `The padlock means "private and untampered, with this server", not that the site is honest.`,
				TryIt:       "Click the padlock in your browser on three sites and read who issued each certificate and when it expires.",
				Analogy: `A passport check: the certificate is the passport, the CA is the
government that issued it, and your browser is the border officer who
knows what a genuine passport looks like. A phishing site can hold a
perfectly genuine passport; it just belongs to someone else.`,
				Example: `Let's Encrypt, launched in 2015, issues free certificates, and it helped
HTTPS go from a minority of page loads to the large majority. Google
reported that after it required hardware security keys for its
employees, it had no confirmed phishing takeovers of their work
accounts.`,
				Exercises: trio(
					`A friend says: "The site has a padlock, so it's safe to type my bank password." What does the padlock actually guarantee, and what does it not?`,
					"Encryption to that domain, not that the domain belongs to your bank.",
					"Use openssl s_client -connect example.com:443 to inspect a real TLS handshake and certificate chain, and identify the issuer, the validity dates and the key type.",
					"Pipe the certificate to openssl x509 -noout -text for a readable version.",
					"Add login to a small web app with properly hashed passwords, secure session cookies (HttpOnly, Secure, SameSite), rate limiting on failed logins, and TOTP two-factor codes that work with authenticator apps.",
					"TOTP is a public standard (RFC 6238); use a well-tested library for it.",
				),
			},
		},
		Resources: []Resource{
			{"Site", "OWASP Top 10:2025", "https://owasp.org/Top10/2025/0x00_2025-Introduction/", "The current list of the most critical web-application security risks."},
			{"Course", "PortSwigger Web Security Academy (free)", "https://portswigger.net/web-security", "Hands-on labs for every major web vulnerability."},
			{"Site", "picoCTF", "https://picoctf.org/", "Beginner-friendly capture-the-flag security challenges."},
			{"Site", "Cryptopals crypto challenges", "https://cryptopals.com/", "Learn cryptography by breaking it, step by step."},
			{"Book", "Crypto 101 (free)", "https://www.crypto101.io/", "An introductory book on cryptography for programmers."},
			{"Tool", "Have I Been Pwned", "https://haveibeenpwned.com/", "Check whether your accounts appear in known breaches."},
		},
		Blueprints: []Blueprint{
			{"Password Manager", "A local, encrypted password vault with a master password, key derivation (Argon2) and authenticated encryption.",
				[]string{"Key derivation from the master password", "Encrypt/decrypt vault (AES-GCM)", "Add, get and generate passwords", "Threat model document"}},
			{"Secure Login Service", "A web login system with hashed passwords, sessions, rate limiting, TOTP 2FA and a security review.",
				[]string{"Sign-up and login", "Session cookies", "Rate limiting", "TOTP 2FA", "Self-audit against OWASP Top 10"}},
			{"CTF Journey", "Solve 30 beginner capture-the-flag challenges and write up how you solved each one.",
				[]string{"Web challenges", "Crypto challenges", "Forensics challenges", "Write-ups"}},
		},
		Quiz: []Question{
			{"What does the CIA triad stand for?", "Confidentiality, Integrity and Availability."},
			{"What is the correct fix for SQL injection?", "Parameterised queries (prepared statements), so input is never interpreted as SQL."},
			{"Why use bcrypt or Argon2 instead of SHA-256 for passwords?", "They are deliberately slow and salted, which makes mass guessing expensive."},
			{"With public-key crypto, which key signs a message and which verifies it?", "The private key signs; anyone can verify with the matching public key."},
		},
	},

	// -----------------------------------------------------------------
	11: {
		Overview: `What can computers do at all, and what can they do efficiently? This
stage explores the mathematical limits of computation: simple machines
that recognise patterns, the Turing machine that models every computer,
problems that no program can ever solve, and the famous P vs NP question.`,
		Outcomes: []string{
			"Use regular expressions and grammars confidently, knowing their limits",
			"Explain why some problems can never be solved by any program",
			"Recognise hard (NP-complete) problems and choose practical approaches",
		},
		Glossary: []Term{
			{"Automaton", "An abstract machine that reads input and moves between states."},
			{"Regular expression", "A pattern language for matching text, such as \\d{3}-\\d{4}."},
			{"Grammar", "A set of rules describing which strings belong to a language."},
			{"Turing machine", "A simple theoretical machine that can compute anything any computer can."},
			{"Decidable", "Solvable by a program that always finishes with a yes/no answer."},
			{"Polynomial time", "Running time bounded by a power of the input size, such as n² or n³; considered 'efficient'."},
			{"NP", "Problems whose proposed solutions can be checked quickly."},
			{"NP-complete", "The hardest problems in NP; a fast solution to one would solve them all."},
			{"Heuristic", "A practical rule of thumb that usually finds good answers, without guarantees."},
		},
		Concepts: []Concept{
			{
				Name:    "Finite Automata & Regular Expressions",
				Summary: "The simplest machines, and the pattern language they power.",
				Body: `A finite automaton is a machine with a fixed number of states. It reads
its input one symbol at a time and moves between states; if it finishes
in an accepting state, the input matches.

Regular expressions describe exactly the same patterns, using sequence,
alternation (a|b) and repetition (a*). They are perfect for tokens, IDs
and simple formats. But they cannot count without limit, so no regular
expression can check that brackets are balanced to any depth. Real regex
engines add features (such as backreferences) beyond the theory, and some
patterns take exponential time on hostile input.`,
				Diagram: `accepts strings over {a, b} that end in "ab":

          a            b
 ─▶ (S0) ───▶ (S1) ───▶ ((S2))      S0 on b → S0     S1 on a → S1
                                     S2 on a → S1     S2 on b → S0`,
				MentalModel: "If you only need to remember a fixed amount, a finite automaton (or a regex) can do it.",
				TryIt:       "Write a regex for UK postcodes or US ZIP+4 codes and test it on regex101.com.",
				Analogy: `A turnstile is either locked or unlocked. A coin unlocks it, and pushing
through locks it again. It remembers nothing beyond its current state,
exactly like a finite automaton.`,
				Example: `Search-and-replace in editors, form validation for emails and phone
numbers, and the lexers in compilers all use regular expressions. In
July 2019 a single badly written regex caused a global Cloudflare outage
by driving its servers' CPUs to 100% (catastrophic backtracking).`,
				Exercises: trio(
					"Draw a finite automaton that accepts binary strings containing an even number of 1s.",
					`Two states are enough: "even so far" and "odd so far".`,
					`Implement a DFA simulator (states, transitions, accepting states) and run both the "ends in ab" machine and your even-1s machine on 20 test strings.`,
					"A dictionary from (state, symbol) to the next state.",
					"Write a log scanner that uses regular expressions to extract timestamps, IP addresses and status codes from a web-server log, and reports the number of errors per hour.",
					`Named groups such as (?P<ip>\d+\.\d+\.\d+\.\d+) keep the code readable.`,
				),
			},
			{
				Name:    "Grammars & Pushdown Automata",
				Summary: "Adding a stack lets machines understand nesting.",
				Body: `A context-free grammar describes languages with nesting, using rules
such as Expr → Expr + Term | Term. Programming languages, JSON, HTML and
arithmetic are all (mostly) context-free. The matching machine is a
pushdown automaton: a finite automaton plus a stack, which lets it
remember how many brackets are open. Parsers (Stage 5 builds one) are
practical pushdown automata.

The Chomsky hierarchy orders these classes of languages from the
simplest to the most powerful.`,
				Diagram: `recursively enumerable     ◀ Turing machines
  └ context-sensitive
      └ context-free         ◀ pushdown automata (code, JSON, HTML)
          └ regular          ◀ finite automata, regexes`,
				MentalModel: "Finite memory handles patterns; a stack handles nesting.",
				TryIt:       `Write a grammar for arithmetic with +, * and parentheses, and derive "2 * (3 + 4)" by hand.`,
				Analogy: `Russian nesting dolls: to know that every doll is closed again, you must
remember how many are open, which is a stack of open dolls. A machine
with only a fixed number of states cannot do that for any depth.`,
				Example: `This is why "you can't parse HTML with regex" is a famous programmer
joke; browsers use real parsers. Tools such as ANTLR and Tree-sitter
generate parsers from grammars, for compilers and for code editors.`,
				Exercises: trio(
					`Using the rule S → ( S ) S | ε, derive "(()())".`,
					`ε means "empty". Expand the leftmost S each time.`,
					"Write a recursive-descent parser and evaluator for: Expr → Term (('+'|'-') Term)*, Term → Factor (('*'|'/') Factor)*, Factor → number | '(' Expr ')'.",
					"One function per rule, each consuming tokens.",
					"Write a checker that validates brackets, braces and quotes in a JSON-like config file and reports the line and column of the first error, as editors do when they underline a missing bracket.",
					"A stack of open symbols, each stored with its position.",
				),
			},
			{
				Name:    "Turing Machines & Computability",
				Summary: "A simple model that captures everything any computer can compute.",
				Body: `In 1936 Alan Turing described a machine with an infinite tape, a
read/write head and a finite table of rules. Despite its simplicity, it
can compute anything a modern computer can. The Church–Turing thesis
states that it captures everything we mean by "computable".

A system that can simulate a Turing machine is Turing-complete: Python,
C, and even surprising things such as PowerPoint animations and Magic:
The Gathering card rules. A universal Turing machine can run any other
machine given its description, which is the idea behind the
stored-program computer: programs are just data.`,
				Diagram: `tape:  … │ 1 │ 0 │ 1 │ 1 │ _ │ _ │ …
                 ▲
                head        state: q2
rule:  (q2, reads 1) → write 0, move right, go to state q3`,
				MentalModel: "Every computer is the same machine in different clothes; only speed and memory differ.",
				TryIt:       "Run a Turing-machine simulator online and watch a binary-increment machine add 1 to 1011.",
				Analogy: `A clerk with an endless paper strip, a pencil and a rulebook that says
"if you are in mood 3 and see a 1, erase it, write 0, step right and
switch to mood 5". Painfully slow, but given enough paper and time, the
clerk can compute anything your laptop can.`,
				Example: `Turing's ideas came from the foundations of mathematics, and during the
Second World War he helped break the Enigma cipher at Bletchley Park.
Researchers have shown PowerPoint (2017) and Magic: The Gathering (2019)
to be Turing-complete.`,
				Exercises: trio(
					"Trace, by hand, a Turing machine that flips every bit (0 ↔ 1) until it reaches a blank, on the input 1010.",
					"One working state plus a halt state is enough.",
					"Write a Turing-machine simulator (the tape as a dictionary, the rules as a table) and program it to increment a binary number.",
					"Move to the rightmost bit, then carry leftwards: 1 → 0 and continue; 0 → 1 and halt; blank → 1 and halt.",
					`Research one "accidentally Turing-complete" system (PowerPoint, Magic: The Gathering, Minecraft…) and explain in half a page which of its features provide memory and which provide conditional branching, the two ingredients of computation.`,
					"Look for something that stores state and something that chooses what happens next.",
				),
			},
			{
				Name:    "The Halting Problem & Undecidability",
				Summary: "Some questions no program can ever answer.",
				Body: `Turing proved that no program can decide, for every program and input,
whether it will eventually halt or run forever. The proof is by
contradiction. Suppose halts(P, x) existed. Build a program that does the
opposite of whatever halts predicts about itself: it halts exactly when
it does not.

By Rice's theorem, every non-trivial question about what a program does
(as opposed to how its text looks) is undecidable in general. That is why
antivirus tools, compilers and bug finders use approximations: for
arbitrary programs, they can never be both sound (never missing a
problem) and complete (never raising a false alarm).`,
				Diagram: `def trouble(p):
    if halts(p, p):      # ask the "oracle" about p run on itself
        loop_forever()
    else:
        return

trouble(trouble)  →  if it halts, it loops; if it loops, it halts.  Contradiction.`,
				MentalModel: "Programs cannot fully predict programs, so tools must approximate.",
				TryIt:       "Explain the halting proof to a friend, using only the 'trouble' program above.",
				Analogy: `"This sentence is false." If it is true, it is false; if it is false, it
is true. The halting proof builds the same self-referential trap for any
claimed perfect program checker.`,
				Example: `This is why your editor's warnings sometimes miss real bugs or flag code
that is fine, and why no antivirus can detect all malware. Static
analysers such as Infer (Meta) and Error Prone (Google) deliberately
trade some precision for usefulness.`,
				Exercises: trio(
					`Which can a program decide, and which not in general: "does this file contain the word print?", "does this program ever print anything?", "does this program have more than 100 lines?"`,
					"Questions about behaviour are undecidable in general; questions about text are easy.",
					`Write a "loop detector" that runs a function with a time limit and reports "probably infinite". Construct a function it wrongly flags, and explain why no time limit can be right for all programs.`,
					"A long but finite computation beats any fixed limit.",
					"Run a static analyser on one of your projects (pylint, mypy, go vet or clippy) and classify three of its warnings as a true bug, a false alarm or style. Explain why the tool cannot be perfect.",
					"Undecidability forces a trade-off between missed bugs and false alarms.",
				),
			},
			{
				Name:    "Complexity Classes: P, NP & Hard Problems",
				Summary: "Which problems can be solved quickly, and the million-dollar question.",
				Body: `P is the class of problems solvable in polynomial time, such as sorting
and shortest paths. NP is the class where a proposed solution can be
checked quickly, like verifying a filled-in Sudoku. NP-complete problems
(SAT, the travelling-salesperson decision problem, graph colouring,
knapsack) are the hardest in NP: a fast algorithm for any one of them
would give fast algorithms for all of them.

Whether P = NP is unknown, and the Clay Mathematics Institute offers
$1,000,000 for the answer. In practice, NP-hard problems are tackled with
approximations, heuristics and solvers that work well on real inputs.
Much of modern cryptography relies on problems that are believed to be
hard to solve but are easy to check.`,
				Diagram: `┌───────────────── NP: easy to CHECK ─────────────────┐
│ ┌──── P: easy to SOLVE ────┐                        │
│ │ sorting, shortest paths, │   NP-complete:         │
│ │ matching, primality      │   SAT, TSP, knapsack,  │
│ └──────────────────────────┘   graph colouring      │
└─────────────────────────────────────────────────────┘
  if P = NP, the two boxes would be the same`,
				MentalModel: "Checking an answer is often easy even when finding it is hard.",
				TryIt:       "Solve a Sudoku by backtracking and count how many positions the search tries.",
				Analogy: `Finding a four-digit lock combination by trying every option is slow;
checking that "7392" opens it takes a second. NP problems are those
where checking is easy. Whether finding is always genuinely hard is the
open question.`,
				Example: `Delivery companies solve huge routing problems (relatives of the
travelling-salesperson problem) with heuristics every day; UPS has said
its ORION routing system saves about 100 million miles a year. Airline
crew scheduling and chip layout rely on solvers for NP-hard problems.`,
				Exercises: trio(
					"For a Sudoku, explain why checking a filled grid is fast, while solving a blank one may require trying many possibilities.",
					"Checking means looking at each row, column and box once.",
					"Solve the travelling-salesperson problem for 8 cities by brute force (all permutations) and with the nearest-neighbour heuristic. Compare route lengths and running times, then try 11 cities.",
					"With the start fixed, 11 cities give 10! = 3,628,800 routes.",
					"Plan a delivery route through 15 real addresses in your town (approximate coordinates are fine). Implement nearest-neighbour plus 2-opt improvement, and compare it with the order you would intuitively drive.",
					"2-opt repeatedly reverses a segment of the route whenever that makes the route shorter.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "MIT 18.404J Theory of Computation (OCW)", "https://ocw.mit.edu/courses/18-404j-theory-of-computation-fall-2020/", "Michael Sipser's own lectures."},
			{"Tool", "regex101", "https://regex101.com/", "Test and understand regular expressions interactively."},
			{"Tool", "JFLAP", "https://www.jflap.org/", "Build and simulate automata, grammars and Turing machines."},
			{"Site", "Complexity Zoo", "https://complexityzoo.net/Complexity_Zoo", "An encyclopaedia of complexity classes, for the curious."},
		},
		Blueprints: []Blueprint{
			{"Regex Engine", "Build your own regular-expression matcher: parse the pattern, compile it to an NFA, and simulate it without exponential backtracking.",
				[]string{"Pattern parser", "Thompson NFA construction", "NFA simulation", "Character classes and anchors", "Benchmark against backtracking"}},
			{"Automata Playground", "A tool that loads DFAs, NFAs and Turing machines from files, runs them and shows every step.",
				[]string{"DFA runner", "NFA → DFA conversion", "Turing machine runner", "Step-by-step trace output"}},
			{"Sudoku & SAT Solver", "Solve Sudoku by translating it into a SAT problem and writing a DPLL solver.",
				[]string{"Backtracking Sudoku solver", "CNF encoding of Sudoku", "DPLL SAT solver", "Compare speeds"}},
		},
		Quiz: []Question{
			{"Why can no regular expression match all balanced bracket strings?", "It would need to count unboundedly deep nesting, but finite automata have only a fixed number of states."},
			{"What does the halting problem say?", "No program can correctly decide, for every program and input, whether it halts."},
			{"What makes a problem NP-complete?", "It is in NP (solutions can be checked quickly), and every NP problem reduces to it."},
			{"What do we do in practice with NP-hard problems?", "Use heuristics, approximation algorithms, or solvers that work well on real inputs."},
		},
	},

	// -----------------------------------------------------------------
	12: {
		Overview: `Every programming language is a set of design choices about how to
express computation. This stage teaches the big ideas behind all of
them (paradigms, type systems, memory management, execution models and
concurrency), so that you can learn any new language in days instead of
months.`,
		Outcomes: []string{
			"Recognise the paradigm and design choices behind any language",
			"Use types, memory models and concurrency models deliberately",
			"Pick up a new programming language quickly",
		},
		Glossary: []Term{
			{"Paradigm", "A style of programming, such as imperative, object-oriented or functional."},
			{"Object", "A bundle of data together with the functions (methods) that act on it."},
			{"Pure function", "A function whose output depends only on its inputs, with no side effects."},
			{"Static typing", "Types are checked before the program runs."},
			{"Garbage collection", "Automatic reclaiming of memory that the program no longer uses."},
			{"Interpreter", "A program that runs source code directly, step by step."},
			{"Bytecode", "A compact, portable instruction format run by a virtual machine."},
			{"JIT", "Just-in-time compiler: compiles frequently run code to machine code while the program runs."},
			{"Concurrency", "Structuring a program as several tasks that make progress independently."},
		},
		Concepts: []Concept{
			{
				Name:    "Paradigms: Imperative, Object-Oriented & Functional",
				Summary: "Different ways of organising the same computation.",
				Body: `Imperative programming gives step-by-step commands that change state (C,
early BASIC). Object-oriented programming bundles data with the methods
that act on it into objects, using encapsulation, inheritance and
polymorphism (Java, C#, Python). Functional programming builds programs
from pure functions that do not change state, and treats functions as
values that can be passed around (Haskell, Clojure, and map/filter/reduce
in most languages). Declarative languages describe what you want, not
how to get it (SQL, HTML, regex).

Most modern languages mix paradigms. The skill is choosing the right
style for each part of a program.`,
				MentalModel: "Paradigms are lenses; use the one that makes the problem simplest.",
				TryIt:       "Sum the squares of the even numbers in a list three ways: a for loop, a class, and map/filter/reduce.",
				Analogy: `Giving directions. Imperative: "walk 200 m, turn left, go two blocks".
Declarative: "meet me at the library", and you find your own way.
Object-oriented: asking a taxi driver (an object that knows how to drive)
to take you there.`,
				Example: `React, which powers the web interfaces of Facebook and Instagram,
popularised a functional, declarative style for user interfaces. Trading
firms such as Jane Street build on the functional language OCaml. Java's
object-oriented style runs much of the world's enterprise software, and
Android apps were historically written in Java.`,
				Exercises: trio(
					"Classify each as imperative, object-oriented, functional or declarative: a SQL query, a recipe, a spreadsheet formula, a class Dog with a bark() method.",
					"Does it say how, say what, or bundle data with behaviour?",
					"Implement a shopping cart twice: once with classes (Cart, Item) and once with only pure functions over lists and dictionaries. Compare which is easier to test.",
					"The pure version returns a new cart instead of modifying the old one.",
					"Take a script that processes a CSV of orders with nested loops and mutable counters, and rewrite the processing as a pipeline of pure functions (parse → filter → group → summarise), keeping its tests passing.",
					"Each stage takes data and returns new data, with no global variables.",
				),
			},
			{
				Name:    "Type Systems",
				Summary: "Rules that catch mistakes before your code runs.",
				Body: `A type system classifies values and checks that operations make sense.
Static typing checks before the program runs (Go, Rust, Java,
TypeScript); dynamic typing checks while it runs (Python, JavaScript,
Ruby). Strong versus weak typing is about implicit conversions, such as
what "5" + 3 does. Type inference lets the compiler work out types for
you, and generics let one function work safely over many types
(List<T>).

Richer types can rule out whole classes of bugs. Optional types replace
null, which Tony Hoare called his "billion-dollar mistake". Sum types
(enums that carry data) force you to handle every case.`,
				MentalModel: "Types are tests that the compiler writes and runs for free.",
				TryIt:       "Add type hints to a Python program and run mypy to find mismatches.",
				Analogy: `Electrical plugs are shaped so that you cannot push a kettle's plug into
a phone-charger socket. The shape (the type) prevents a dangerous mistake
before any current flows.`,
				Example: `Tony Hoare introduced the null reference in 1965 and later called it his
"billion-dollar mistake". TypeScript, which adds static types to
JavaScript, is now used by a large share of web developers, and several
companies have reported that it would have caught a significant fraction
of their production bugs.`,
				Exercises: trio(
					`Predict the result in Python and in JavaScript: "5" + 3 and "5" * 3. Which language is stricter here?`,
					`Python raises TypeError for "5" + 3; JavaScript converts, giving "53" and 15.`,
					"Model a traffic light as a sum type (Red | Yellow | Green) in Rust, TypeScript or Python (Enum plus match), with a next() function. Show that forgetting a case is caught by the compiler or type checker.",
					"Rust's match, or a TypeScript switch with a never check, both detect missing cases.",
					"Take a small untyped JavaScript or Python project (yours or an open-source one), add types step by step (TypeScript, or mypy --strict), and record every real bug or unclear piece of code the types exposed.",
					"Start at the edges: function signatures for inputs and outputs.",
				),
			},
			{
				Name:    "Memory Management",
				Summary: "Who cleans up the memory your program no longer needs?",
				Body: `Programs allocate memory while they run. With manual management (C,
C++), the programmer frees it, and mistakes cause leaks (memory never
freed), use-after-free and double-free bugs, a major source of security
holes.

Garbage collection (Java, Go, Python, JavaScript) automatically finds
unreachable objects and frees them, at the cost of some CPU time and
occasional pauses. Reference counting (Python, Swift) frees an object as
soon as nothing refers to it, but needs extra help with cycles. Rust's
ownership system checks at compile time that every value has exactly one
owner and that references never outlive it: memory safety without a
garbage collector.`,
				MentalModel: "Every allocation needs an owner who knows when it is finished.",
				TryIt:       "In Python, use sys.getrefcount and the gc module to watch an object's reference count change.",
				Analogy: `A library. With manual management, you must return every book yourself:
forget and the shelves empty out; return one twice and the records are
chaos. Garbage collection is a librarian who periodically collects the
books nobody is reading. Rust's ownership is a strict checkout desk that
will not let you leave until every loan is accounted for.`,
				Example: `Microsoft and Google's Chrome team have each reported that about 70% of
their serious security bugs are memory-safety problems. That is why
Android and the Linux kernel have started accepting Rust code, and why
US government agencies have urged a move to memory-safe languages.`,
				Exercises: trio(
					"Explain what goes wrong in each case: forgetting to free memory, freeing it twice, and using it after freeing it.",
					"A leak, a double free and a use-after-free.",
					"In C, write a program with a memory leak and a use-after-free, find both with AddressSanitizer or Valgrind, and fix them.",
					"gcc -g -fsanitize=address prog.c && ./a.out",
					"A long-running Python web service grows from 200 MB to 4 GB over a week. Reproduce a similar leak in a small script (for example an ever-growing cache), find it with tracemalloc, and fix it with a bounded cache.",
					"functools.lru_cache(maxsize=…) or an explicit eviction policy.",
				),
			},
			{
				Name:    "Interpreters, Compilers & Runtimes",
				Summary: "Different ways of turning source code into a running program.",
				Body: `An interpreter reads your program and executes it directly (shell
scripts, simple interpreters). A compiler translates it ahead of time
into machine code (C, Go, Rust): fast to run, slower to build. Many
languages compile to bytecode for a virtual machine (Java's JVM, Python's
CPython VM, .NET's CLR), which makes the program portable across CPUs.

Just-in-time (JIT) compilers watch the running program and compile its
hot paths to machine code, getting the best of both worlds (V8 for
JavaScript, the JVM's HotSpot, PyPy). The runtime provides services such
as garbage collection, threads and the standard library.`,
				Diagram: `source ─▶ [compiler] ─▶ machine code ──────────────▶ CPU   (C, Go, Rust)
source ─▶ [compiler] ─▶ bytecode ─▶ [VM + JIT] ────▶ CPU   (Java, C#, JavaScript)
source ─────────────────────────▶ [interpreter] ───▶ CPU   (shell scripts)`,
				MentalModel: "Interpreted versus compiled describes an implementation, not a language.",
				TryIt:       "Use Python's dis module to see the bytecode of a small function.",
				Analogy: `A live interpreter translating a speech sentence by sentence (an
interpreter), versus a translated book prepared in advance (a compiler).
A JIT is an interpreter who notices you keep repeating the same phrases
and writes down their translations for instant reuse.`,
				Example: `V8's JIT (in Chrome and Node.js) is why JavaScript went from slow
scripting to running full applications such as Google Docs in the
browser. WebAssembly lets C, C++ and Rust code run in browsers at
near-native speed; Figma uses it.`,
				Exercises: trio(
					"For Python (CPython), Go and JavaScript in Chrome, say whether each is usually compiled ahead of time, run as bytecode by a VM, or JIT-compiled.",
					"CPython: a bytecode VM. Go: compiled to native code. V8: JIT.",
					`Write a tiny stack-based bytecode VM (PUSH, ADD, MUL, PRINT) and a compiler that turns expressions such as "2 * (3 + 4)" into that bytecode.`,
					"Compile the syntax tree in post-order: children first, then the operator.",
					"Benchmark the same CPU-heavy function (for example naive Fibonacci of 30) in CPython, PyPy (if available), Node.js and Go, and explain the differences using the words interpreter, JIT and compiler.",
					"Run each several times and take the median.",
				),
			},
			{
				Name:    "Concurrency Models",
				Summary: "How languages let programs do many things at once, safely.",
				Body: `Concurrency is structuring a program as independent tasks; parallelism
is actually running them at the same time on several cores. The main
models:
- threads with shared memory and locks (Java, C++): powerful, but prone to races and deadlocks;
- message passing between isolated processes (Erlang actors, Go channels): "don't communicate by sharing memory; share memory by communicating";
- async/await with an event loop (JavaScript, Python asyncio, Rust): great for waiting on many network requests;
- data parallelism: the same operation on lots of data at once (GPUs, Stage 16).

Choosing the right model matters more than micro-optimising.`,
				MentalModel: "Prefer passing messages over sharing state; when you must share, lock carefully.",
				TryIt:       "Download 20 web pages one after another, then concurrently with asyncio or goroutines, and compare the times.",
				Analogy: `A kitchen. Several cooks sharing one chopping board need rules (locks).
Cooks passing dishes through a hatch (channels) avoid collisions. A
single waiter who takes an order, then serves other tables while the
kitchen cooks, is async/await.`,
				Example: `WhatsApp's servers are written in Erlang, whose actor model handles
millions of connections, and Discord uses Elixir, which runs on the same
virtual machine. Go's goroutines and channels power Docker and
Kubernetes.`,
				Exercises: trio(
					"Is each one concurrency, parallelism or both: one chef juggling three dishes; three chefs each cooking one dish; a single-core computer playing music while you browse?",
					"Parallelism needs things happening at literally the same instant.",
					"Write a producer–consumer program: one task generates numbers, three workers square them, and one task collects the results, using channels (Go) or queues (Python).",
					"Close the channel, or send a sentinel value, to tell the workers to stop.",
					"Build a link checker that crawls a site you own (or a local test site) concurrently, with at most 10 requests in flight at a time, and reports broken links.",
					"A semaphore or a fixed worker pool limits concurrency so you do not overload the server.",
				),
			},
		},
		Resources: []Resource{
			{"Course", "Programming Languages (Dan Grossman, Coursera)", "https://www.coursera.org/learn/programming-languages", "Functional programming, types and language design, beautifully taught."},
			{"Book", "Crafting Interpreters (free online)", "https://craftinginterpreters.com/", "Build a tree-walker and a bytecode VM with a garbage collector."},
			{"Book", "The Rust Programming Language (free)", "https://doc.rust-lang.org/book/", "Learn ownership, types and fearless concurrency."},
			{"Paper", "Why Functional Programming Matters (Hughes)", "https://www.cse.chalmers.se/~rjmh/Papers/whyfp.html", "A classic, readable argument for functional style."},
			{"Book", "Programming Languages: Application and Interpretation (free)", "https://www.plai.org/", "Understand languages by implementing them."},
		},
		Blueprints: []Blueprint{
			{"Language with a Bytecode VM", "Design a small language with functions and closures, compile it to bytecode, and run it on your own VM with a garbage collector.",
				[]string{"Lexer and parser", "Bytecode compiler", "Stack VM", "Closures", "Mark-and-sweep GC"}},
			{"Same App, Three Paradigms", "Build one small app (a library catalogue) in an OO style, a functional style and a procedural style, and write up the trade-offs.",
				[]string{"Procedural version", "Object-oriented version", "Functional version", "Comparison write-up"}},
			{"Concurrent Web Crawler", "A polite, concurrent crawler with rate limiting and graceful shutdown, written with channels or async/await.",
				[]string{"Fetch and parse links", "Worker pool", "Rate limiting per host", "Graceful shutdown", "Report"}},
		},
		Quiz: []Question{
			{"What is a pure function?", "A function whose result depends only on its inputs and that has no side effects."},
			{"What is the difference between static and dynamic typing?", "Static typing checks types before running; dynamic typing checks them while running."},
			{"How does Rust achieve memory safety without a garbage collector?", "Its ownership and borrowing rules are checked at compile time."},
			{"What is the difference between concurrency and parallelism?", "Concurrency is structuring work as independent tasks; parallelism is executing them simultaneously."},
		},
	},
}
