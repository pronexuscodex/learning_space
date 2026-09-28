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
}
