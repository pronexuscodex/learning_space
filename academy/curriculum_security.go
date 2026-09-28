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

	// -----------------------------------------------------------------
	18: {
		Overview: `Most breaches today do not break code; they log in. Stolen passwords,
tricked employees, over-powerful accounts and misconfigured cloud
storage cause a large share of real incidents. This stage covers how
people and programs prove who they are, how permissions are designed so
one stolen account cannot do much damage, how to secure the cloud
accounts that now run most software, and the zero-trust idea that no
network location is trusted by itself.`,
		Outcomes: []string{
			"Choose phishing-resistant authentication: MFA, hardware keys and passkeys",
			"Explain sessions, OAuth 2.0 and OpenID Connect, and validate tokens correctly",
			"Design roles and permissions around least privilege, and review them regularly",
			"Secure a cloud account using the shared responsibility model",
			"Explain zero trust and how it differs from trusting the office network",
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
					"Add TOTP two-factor authentication to a small web app you built: enrolment with a QR code, verification with a time window, and one-time recovery codes stored hashed.",
					"Use a well-tested library for TOTP (RFC 6238); never invent the maths yourself.",
					"Do an account-security audit of your own digital life: every account that can reset others (email, phone carrier, password manager), its second factor, and its recovery options. Upgrade the weakest three.",
					"Your email account can reset almost everything else, so protect it first and best.",
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
