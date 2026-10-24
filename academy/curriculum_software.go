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
command line and shell scripts, Git, build systems and dependencies,
testing, debugging tools, design, and shipping to real users. Each lesson
looks under the hood with strace, Git's own object files, make, gcov,
gdb and Linux namespaces.`,
		Outcomes: []string{
			"Work comfortably in a terminal and automate chores with scripts",
			"Collaborate through Git and pull requests like a professional team",
			"Test, structure and deploy software that others can rely on",
			"Explain what Git stores, and what a build system rebuilds and why",
			"Find a bug with a debugger, a core dump or git bisect instead of guessing",
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
			{"Exit status", "The number a program returns when it ends: 0 for success, anything else for failure."},
			{"Hash (Git object ID)", "The SHA-1 of an object's content, which Git uses as its name: the same content always gets the same ID."},
			{"Dependency graph", "Which files or packages each one needs; a build system walks it to decide what to rebuild."},
			{"Lockfile", "A file recording the exact version of every dependency, so every build uses the same ones."},
			{"Code coverage", "Which lines and branches of a program its tests actually ran."},
			{"Backtrace", "The chain of function calls that led to the current point, as a debugger prints it."},
			{"git bisect", "A binary search through history for the commit that introduced a bug."},
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
reproducibly.

Some commands are programs on disk (ls is /usr/bin/ls); others are
built into the shell (cd, echo, export), because they must change the
shell itself: a separate program cannot change its parent's current
directory. type ls tells you which kind a command is.`,
				Diagram: `you type:  ls -l | wc -l
shell:     fork ─▶ child: search PATH for "ls" ─▶ exec /usr/bin/ls
           fork ─▶ child: search PATH for "wc" ─▶ exec /usr/bin/wc
           ls's standard output ──pipe──▶ wc's standard input`,
				UnderTheHood: `strace shows the shell finding ls. With PATH set to
/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/bin, bash checked each
folder in order:

  newfstatat("/usr/local/sbin/ls")   ENOENT, no such file
  newfstatat("/usr/local/bin/ls")    ENOENT
  newfstatat("/usr/sbin/ls")         ENOENT
  newfstatat("/usr/bin/ls")          found: a 142,312-byte program
  execve("/usr/bin/ls", ["ls"], … 4 vars …)

That is all PATH is: a list of folders searched left to right, so an
earlier folder wins. The environment travels with every execve: in a
normal session here, 152 variables and 8,278 bytes were copied into
each program started. The kernel caps arguments plus environment at
getconf ARG_MAX, 2,097,152 bytes, which is why rm * fails with
"Argument list too long" in a folder of a million files.`,
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
				Name:    "Shell Scripting & Automation",
				Summary: "A script is a recipe of commands: quoting, exit codes and failing loudly make it safe.",
				Body: `A shell script is a text file of commands, run top to bottom, with
variables, if, for and functions. It turns a chore you do by hand into
one command that does it the same way every time.

Three rules prevent most script bugs:
- Quote every variable: "$file", not $file. Unquoted, the shell splits
  the value at spaces and expands wildcards, so a file named "my
  file.txt" becomes two arguments, "my" and "file.txt".
- Check exit statuses. Every command returns a number: 0 means success,
  anything else failure. if, && and || test it; $? holds the last one.
- Fail loudly: start scripts with set -euo pipefail. -e stops at the
  first failing command, -u treats an unset variable as an error
  instead of an empty string, and pipefail makes a pipeline fail if
  any part of it fails, not just the last.

The classic disaster is rm -rf "$DIR/" with DIR unset: without -u it
becomes rm -rf /. For anything longer than a page, or with real data
structures, switch to a proper language; shell is glue.`,
				Diagram: `#!/usr/bin/env bash
set -euo pipefail          # stop on errors, unset names, broken pipes
src=$1                     # the first argument
for f in "$src"/*.log; do  # quoted: names with spaces stay whole
  gzip -- "$f" || echo "could not compress $f" >&2
done`,
				UnderTheHood: `Word splitting, measured in bash. With f="my file.txt" and that file
present:

  ls $f       ls: cannot access 'my': No such file or directory
              ls: cannot access 'file.txt': No such file or directory
              exit status 2
  ls "$f"     my file.txt
              exit status 0

Exit statuses as bash saw them:

  exit 0           → 0
  exit 3           → 3
  false            → 1
  kill -9 $$       → 137   (128 + 9: killed by SIGKILL)
  false | true     → 0, but 1 with set -o pipefail

Without pipefail, a pipeline's status is only its last command's, so
a script that runs "curl … | tar x" carries on happily when the
download fails.`,
				MentalModel: "Quote everything, check every status, and make the script stop at the first surprise.",
				TryIt:       "Create a file whose name contains a space, then compare for f in $(ls) and for f in * with echo \"[$f]\" in each loop.",
				Analogy: `A recipe for a kitchen assistant who follows it literally. "Add the
contents of the box" means one thing if the box is labelled "flour" and
quite another if the label is missing and the assistant grabs whatever
is nearest. set -u is an assistant who stops and asks.`,
				Example: `In 2015 a bug in Steam's Linux client ran rm -rf "$STEAMROOT/"* with
STEAMROOT empty, deleting users' entire home folders. Shell scripts
still run most build systems, CI pipelines, Docker images and server
start-up, which is why tools such as ShellCheck exist to catch exactly
these mistakes.`,
				Exercises: trio(
					"What does this print when name is unset: echo \"Hello, $name!\"? And with set -u at the top?",
					"Without -u: \"Hello, !\". With -u, bash stops with \"name: unbound variable\".",
					"Write a script that takes a folder as its argument, checks that it exists (exit 1 with a message on stderr otherwise), and renames every .jpeg file in it to .jpg, coping with spaces in names. Test it on files named \"a b.jpeg\" and \"-x.jpeg\".",
					"Loop with for f in \"$dir\"/*.jpeg, use \"${f%.jpeg}.jpg\" for the new name, and mv -- so a name starting with - is not taken for an option.",
					"Automate a real chore: a script, run daily by cron (or a systemd timer, or Task Scheduler), that fetches something you check by hand (a web page, a file listing, disk usage), compares it with yesterday's copy, and only tells you when it changed.",
					"Save each day's copy with the date in its name; diff -q returns 1 when the files differ.",
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
commits with clear messages, and never commit passwords or keys.

Underneath, Git is a content-addressed store of three kinds of object,
each named by the SHA-1 hash of its content: blobs (a file's bytes),
trees (a folder: names pointing to blobs and other trees) and commits
(a tree, its parents, the author and the message). A branch is just a
file holding one commit's ID. Because names are hashes of content,
identical files are stored once, and changing anything in history
changes every ID after it.`,
				Diagram: `main:     A ── B ─────────── E   (merge commit)
                ╲           ╱
feature:         C ──── D ─╯

commit ─▶ tree ─▶ blob "hello\n"
  │        └────▶ tree src/ ─▶ blobs …
  └─▶ parent commit`,
				UnderTheHood: `A one-file repository, opened up. git hash-object hello.txt (holding
"hello" and a newline) printed ce013625030ba8dba906f756967f9e9ca394464a,
and so did a plain SHA-1 of the header "blob 6", a zero byte and the 6
bytes of content:

  printf 'blob 6\0hello\n' | sha1sum
  ce013625030ba8dba906f756967f9e9ca394464a

After the first commit, .git/objects held exactly three files: that
blob, a tree ("100644 blob ce0136… hello.txt") and the commit:

  tree aaa96ced2d9a1c8e72c56b253a0e2fe78393feb7
  author Ana <ana@example.com> 1767225600 +0000
  committer Ana <ana@example.com> 1767225600 +0000

  First commit

.git/refs/heads/main was a 41-byte file: the commit's ID and a newline.
In a repository of 1,000 files, changing one added exactly 4 objects: the
new blob, its folder's tree, the root tree and the commit; the other 999
blobs were shared. After a third commit changed every file, the loose
objects took 9,136 KB; git gc packed them, storing similar versions as
deltas, into 172 KB.`,
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
					"Put one of your earlier projects in a Git repository, make at least 5 meaningful commits, and use git log, git diff and git switch --detach to travel back to an old version and return. Then use git cat-file -p on HEAD, its tree and one blob.",
					"git log --oneline --graph shows the history compactly; git rev-parse HEAD^{tree} gives the tree's ID.",
					`Make an open-source contribution: find a project with a "good first issue" or a typo in its documentation, fork it, fix it on a branch and open a pull request.`,
					"Read the project's CONTRIBUTING file first and follow its conventions.",
				),
			},
			{
				Name:    "Build Systems & Dependencies",
				Summary: "Rebuild only what changed, and pin exactly what you depend on.",
				Body: `A build system turns source into programs and knows the dependency
graph: which outputs come from which inputs. make, the classic, reads
rules of the form "target: prerequisites" followed by a command, and
rebuilds a target when any prerequisite is newer than it (by file
modification time). Modern tools (Ninja, Bazel, Go's and Rust's own
builders) compare content hashes instead, and run independent steps in
parallel.

Two mistakes break incremental builds: a missing dependency (a header
not listed, so changing it rebuilds nothing and the program is a mix
of old and new code), and a dependency cycle, which no order can
satisfy.

Libraries you use are dependencies too. Semantic versioning
(MAJOR.MINOR.PATCH) promises that minor and patch releases do not break
you; a lockfile (package-lock.json, Cargo.lock, go.sum) records the
exact versions and their hashes, so every machine builds the same thing
and a tampered download is refused.`,
				Diagram: `app ◀── main.o ◀── main.c
 ▲  ◀── f1.o   ◀── f1.c, util.h
 └───── util.o ◀── util.c, util.h
touch util.h  ⇒  every .o that lists util.h, then app`,
				UnderTheHood: `A Makefile building 42 C files (each including util.h) into one
program, timed on a 4-core machine:

  full build                               1,165 ms
  nothing changed     "'app' is up to date"    5 ms
  one .c file touched   2 commands (compile, link)   62 ms
  util.h touched        43 commands               1,112 ms
  full build with make -j8                    302 ms

make -n prints what would run without running it. Touching one source
file rebuilt one object and relinked; touching the header that every
file includes rebuilt everything, which is why widely included headers
should change rarely. Running jobs in parallel made the full build four
times faster on four cores: the dependency graph says which steps are
independent.`,
				MentalModel: "A build is a graph: rebuild a node only when one of its inputs changed.",
				TryIt:       "Write a Makefile for a three-file C program, then run make, touch one file, run make -n, and predict what it will rebuild before you look.",
				Analogy: `Cooking a meal from several dishes. If you change the sauce recipe, you
remake the sauce and every dish that uses it, but not the salad. A
lockfile is the shopping list that names the exact brand and size of
every ingredient, so the dish tastes the same in any kitchen.`,
				Example: `In 2016 the removal of left-pad, an 11-line npm package, broke the builds
of thousands of projects, including React, within hours. In 2024 the
xz-utils backdoor reached Linux distributions through a trusted
dependency's release files. Google builds most of its code with Bazel,
whose content hashes let it skip any step whose exact inputs were built
before, even on another machine.`,
				Exercises: trio(
					"A Makefile says app: main.o util.o, and main.o: main.c. main.c includes util.h, which the Makefile does not mention. You change util.h. What does make rebuild, and what goes wrong?",
					"Nothing that includes it: main.o is not rebuilt, so app mixes old and new definitions. Generate header dependencies with gcc -MMD.",
					"Write a Makefile for a C project with three source files and one header, using pattern rules and automatic header dependencies (-MMD -MP and -include $(DEPS)). Show that touching the header rebuilds exactly the files that include it.",
					"$(SRCS:.c=.o) builds the object list; $@ is the target and $< the first prerequisite.",
					"Pick a project with a lockfile (package-lock.json, Cargo.lock or go.sum), find one dependency's exact version and hash in it, and check the hash against the registry. Then explain what would happen if the downloaded file were altered.",
					"npm view <package>@<version> dist.integrity prints the same SHA-512 that package-lock.json records.",
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
of random inputs to find cases you did not think of.

Coverage tools report which lines and branches the tests ran. Low
coverage proves that code is untested; high coverage does not prove it
is correct, because a line can run with every input but the one that
breaks it.`,
				Diagram: `        ╱ end-to-end ╲        few, slow, the whole app
       ╱ integration  ╲       some: pieces together
      ╱   unit tests    ╲     many, fast, one function each
     ────────────────────`,
				UnderTheHood: `A C file with average(a, b) = (a + b) / 2 and a clamp function, tested
by hand with average(2, 4), average(-3, 3) and clamp(5, 0, 10): all
correct. gcc --coverage and gcov reported:

  Lines executed: 46.15% of 13
  Branches taken at least once: 25.00% of 12
  clamp marked 1*: run, but some of its branches never taken

Then a million random pairs, each checked against the same formula
computed in 64-bit arithmetic:

  first failure at test 14:
    average(318608798, 2100852393) = -937753052, want 1209730595
  250,239 of 1,000,000 random tests fail

a + b overflows whenever the sum leaves the int range, about one input
in four. Built with -fsanitize=undefined, the run reported it at that same
test: "signed integer overflow: 318608798 + 2100852393 cannot be
represented in type 'int'". Three hand-picked tests saw nothing; random
tests found it in fourteen tries.`,
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
					"Write that age function test-first in C: write one failing test (a small main that compares the result and returns non-zero on a mismatch), make it pass, and repeat until all 8 cases pass. Then measure your coverage with gcc --coverage and gcov.",
					"A test can be one line: if (age(...) != 34) { printf(\"line %d failed\\n\", __LINE__); fails++; }",
					"Set up continuous integration for one of your GitHub repositories (for example with GitHub Actions) so the tests run on every push, and add a status badge to the README.",
					"A workflow file in .github/workflows/ that checks out the code and runs the tests.",
				),
			},
			{
				Name:    "Debugging Tools: gdb, Core Dumps & git bisect",
				Summary: "Stop guessing: a debugger shows where it broke, and bisect finds the commit that broke it.",
				Body: `When a program crashes or misbehaves, three tools replace guesswork:
- A debugger (gdb, lldb) runs the program under control. bt prints the
  backtrace, the chain of calls that led to the crash; frame and print
  show any function's local variables at that moment; break and step
  stop and walk through code. Build with -g so the debugger knows the
  source lines, and -O0 so variables are not optimised away.
- A core dump is the crashed process's memory written to a file. With
  ulimit -c unlimited (or systemd-coredump), you can run gdb ./prog core
  after the fact, even for a crash on a user's machine.
- git bisect is a binary search through history. Mark one commit good
  and one bad; Git checks out the middle, you test it, and it halves the
  range. With a script that returns 0 for good, git bisect run does it
  all automatically.

Underneath them all is the scientific method: reproduce the bug
reliably, make a hypothesis, test it, and change one thing at a time.`,
				Diagram: `1000 commits:  good ●──────────────────────────● bad
bisect:        test 500 ─▶ test 750 ─▶ test 625 ─▶ … (10 tests)
                                                  ▲
                                   the first bad commit`,
				UnderTheHood: `A program that walked an array of user pointers with i <= n instead of
i < n crashed: "Segmentation fault", exit status 139. Under gdb:

  Program received signal SIGSEGV, Segmentation fault.
  #0  __strlen_evex () at ../sysdeps/x86_64/multiarch/strlen-evex-base.S:81
  #1  name_length (u=0x0) at crash.c:4
  #2  total (users=0x7fffffffc710, n=2) at crash.c:5
  #3  main () at crash.c:9
  (gdb) frame 2
  (gdb) print i          $1 = 2
  (gdb) print users[i]   $2 = (struct user *) 0x0

The crash was inside the C library's strlen, but the backtrace shows
who called it with a NULL pointer, and print i shows the loop had run one
step too far. Core dumps were off on this machine: ulimit -c reported 0.

For history, a script made 1,000 commits with a bug introduced in
commit 637. git bisect run, given only "the first commit was good, the
last is bad" and a one-line test, ran the test 10 times and reported
commit 637 as the first bad commit: log₂ 1000 ≈ 10.`,
				MentalModel: "Bisect the time (git bisect) and the space (backtrace), then read the variables.",
				TryIt:       "Build a program that crashes with gcc -g -O0, run it under gdb, and use bt, frame and print to explain the crash in one sentence.",
				Analogy: `A detective at a crime scene (the core dump) reconstructing the chain of
events from the evidence left behind, and the children's game of
guessing a number between 1 and 1,000 with "higher or lower": ten
questions are always enough.`,
				Example: `Linux kernel developers bisect regressions across thousands of commits
between releases, and git bisect run is how a single developer can find
which of a month's changes broke a test overnight. Browsers and operating
systems collect crash reports, essentially trimmed core dumps with
backtraces, from millions of users to find the crashes that matter most.`,
				Exercises: trio(
					"git bisect starts with 4,000 commits between good and bad. At most how many tests does it need? And for 1,000,000 commits?",
					"⌈log₂ 4000⌉ = 12; ⌈log₂ 1,000,000⌉ = 20.",
					"In C, write a program with a bug that only appears for some inputs (for example an off-by-one in a loop). Debug it in gdb with a breakpoint (break file.c:LINE), a condition (condition 1 i == 99), step and print, without adding any printf.",
					"run args, then bt when it stops; watch var stops whenever a variable changes.",
					"In a project with real history (yours, or a clone of an open-source one), find when a behaviour changed using git bisect run and a small test script, and write up which commit it was and why it broke.",
					"The script must exit 0 for good, 1 to 124 for bad, and 125 to skip a commit that does not build.",
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
safe because tests exist.

In C, an interface is a header plus an opaque type: the header declares
struct order; and the functions that use it, while only the .c file
knows its fields, so no caller can depend on them. A table of function
pointers is how C does polymorphism; it is what object-oriented
languages build for you as a vtable.`,
				Diagram: `┌──────────────────────────────────┐
│ interface / API     (what)       │
├──────────────────────────────────┤
│ business logic      (rules)      │
├──────────────────────────────────┤
│ storage / network   (how)        │
└──────────────────────────────────┘
each layer only talks to the layer below it`,
				UnderTheHood: `What does an interface cost? A C program called a square's area
function 200 million times, directly and through a struct of function
pointers whose contents the compiler could not see:

  direct call            1.28 ns per call
  through the ops table  2.91 ns per call

objdump shows the difference: call 12d0 <square_area> for the direct
call, call *(%rax) for the other, a jump to an address loaded from
memory. The extra 1.6 ns is that load and an indirect branch the CPU
must predict. Unless it is in the innermost loop of a hot path, an
abstraction costs nothing you could measure; what costs real time is
the code people cannot understand well enough to change.`,
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
					"Restructure your bill splitter or budget tracker in C so that the calculation logic lives in its own .c file with a header, with no printf or scanf, then write unit tests for that logic alone.",
					"Pure functions take data in and return data; the interface layer (main.c) calls them and does the printing.",
					"Your budget tool reads CSV files from your bank. Design it so that supporting a second bank's different CSV format means writing one small new component, not changing the rest. Implement both.",
					"The adapter pattern: each bank gets a parser that outputs the same Transaction structure; in C, a struct of function pointers selects the parser.",
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
has actually been tested, and blameless post-mortems after incidents.

A container is not a virtual machine: it is an ordinary process that the
kernel gives its own view of the system with namespaces (its own process
IDs, mounts, network and host name) and its own limits with cgroups
(memory, CPU). That is why containers start in milliseconds and share
the host's kernel.`,
				Diagram: `host kernel
├── process: your shell          (the host's namespaces)
└── container: a process tree with its own
      pid namespace   (it sees itself as PID 1)
      mnt namespace   (its own / from the image)
      net namespace   (its own interfaces)
      cgroup          (memory.max, cpu.max)`,
				UnderTheHood: `unshare, run as root, starts a process in new namespaces, the same
kernel feature Docker uses:

  unshare --pid --fork --mount-proc sh -c 'echo $$; ps -e'
  1
      PID COMMAND
        1 sh
        2 ps

Inside the new PID namespace the shell was process 1, and ps saw only
two processes; outside, ps -e listed 79. unshare --uts let a shell set
its host name to box-in-a-box while the machine itself stayed "vm".
/proc/self/cgroup showed the shell running these experiments inside a
memory cgroup set up by the machine's own supervisor.

What goes into the box matters too. "hello" compiled three ways:

  C, linked to the shared C library    15,960 bytes   needs libc.so.6
  C, gcc -static                      785,360 bytes   needs nothing
  Go                                2,209,303 bytes   needs nothing

A static binary runs in an empty container image; a dynamic one needs
the right shared libraries inside it.`,
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
					"Start FROM an official language image and COPY your code in; for a static C or Go binary, FROM scratch works.",
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
			{"Why is cd built into the shell rather than a program in /usr/bin?", "A program runs in a child process and cannot change its parent's current directory; only the shell itself can change its own."},
			{"Why should a shell script write \"$file\" rather than $file?", "Unquoted, the shell splits the value at spaces and expands wildcards, so a name with a space becomes several arguments."},
			{"What does set -euo pipefail do?", "-e stops at the first failing command, -u makes unset variables an error, and pipefail makes a pipeline fail if any of its commands fails."},
			{"What is a Git blob's ID, and why do identical files cost nothing extra?", "The SHA-1 of \"blob <size>\", a zero byte and the content. Identical content has the same ID, so Git stores it once."},
			{"What is a branch, physically?", "A small file in .git/refs/heads holding the ID of one commit; committing on the branch rewrites that file."},
			{"Why does touching a widely included header rebuild almost everything?", "Every object file that lists the header as a prerequisite is now older than it, so make recompiles each of them."},
			{"What does a lockfile guarantee?", "Every build uses the same exact versions of every dependency, with their hashes checked, so builds are reproducible and a tampered download is refused."},
			{"Can 100% line coverage miss a bug?", "Yes: every line can run with inputs that work, while the input that breaks it (an overflow, an edge case) is never tried."},
			{"How many tests does git bisect need for 1,000 commits?", "About 10, since each test halves the range: ⌈log₂ 1000⌉ = 10."},
			{"What makes a container different from a virtual machine?", "A container is a process tree on the host's kernel, isolated with namespaces and limited with cgroups; a virtual machine runs its own kernel on emulated hardware."},
		},
	},

	// -----------------------------------------------------------------
	10: {
		Overview: `Security is about keeping systems doing what they should, even when
someone clever is trying to make them do something else. This stage
teaches how attackers think, the most common vulnerabilities, the
cryptography that protects every password, message and payment, and how
defenders keep networks, code and builds safe, and respond when
something goes wrong. Only ever practise attacks on systems you own or
that explicitly invite it: CTFs, deliberately vulnerable labs, and your
own virtual machines.`,
		Outcomes: []string{
			"Threat-model an app and prioritise its biggest risks",
			"Find and fix common web vulnerabilities (injection, XSS, broken access control)",
			"Use hashing, encryption and signatures correctly, without inventing your own crypto",
			"Explain how memory-safety bugs in C become exploits, and prevent them",
			"Harden a server and its network, and keep secrets and dependencies under control",
			"Detect an attack in logs and run a calm, documented incident response",
			"Read a research paper, a CVE and a CVSS score, and prioritise what to patch",
			"Find an unknown bug by fuzzing and disclose it responsibly, within the law",
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
			{"Buffer overflow", "Writing past the end of a buffer, overwriting whatever memory comes next."},
			{"ASLR", "Address space layout randomisation: code and data load at random addresses, so attackers cannot predict them."},
			{"IDOR", "Insecure direct object reference: changing an ID in a request reaches someone else's data."},
			{"Port scan", "Probing a machine to find which network services answer."},
			{"Supply-chain attack", "Compromising software you depend on so the attack arrives through a trusted channel."},
			{"Incident response", "The planned steps for detecting, containing and recovering from an attack."},
			{"CVE", "Common Vulnerabilities and Exposures: a unique public ID for one known vulnerability, such as CVE-2021-44228."},
			{"CWE", "Common Weakness Enumeration: the class a vulnerability belongs to, such as CWE-89 (SQL injection) or CWE-416 (use-after-free)."},
			{"CVSS", "Common Vulnerability Scoring System: a 0.0 to 10.0 severity score computed from a vector of metrics."},
			{"Zero-day", "A vulnerability the vendor does not yet know about, so no patch exists."},
			{"Fuzzing", "Feeding a program huge numbers of malformed inputs under a sanitizer to find inputs that crash it."},
			{"Coordinated disclosure", "Reporting a vulnerability privately to the vendor and giving them time to fix it before publishing."},
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
				UnderTheHood: `Threat modelling this academy itself. Its assets: your progress, notes,
reflections and career files; the C programs it compiles and runs; and
the releases you install. Its entry points: the files it reads, the web
pages and PDFs it downloads, the code you write, and the installer.

Walking those against what can go wrong found a real problem. Every file
the academy saved was created with mode 0644: readable by every account
on the machine, including notes and job applications. Nothing had ever
gone wrong, which is exactly why no test had noticed. The fix was one
line (the files are now 0600, owner only) and a test that keeps it so.
Other entries came out fine and stay on the list to re-check: releases
are verified against SHA256SUMS before installing, and checked
exercises run only your own code, in a fresh temporary folder, with a
time limit.`,
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
				UnderTheHood: `The same lookup two ways, in Python with SQLite, on a table of 1,000
users, given the name nobody' OR '1'='1:

  "select * from users where name = '" + name + "'"    1,000 rows
  "select * from users where name = ?", (name,)            0 rows

Pasted into the query text, the quote ends the string and OR '1'='1'
becomes part of the condition, true for every row. Passed as a
parameter, the whole thing is only ever a name, and no user has that
name. On the HTML side, Python's html.escape turns
<img src=x onerror="alert(1)"> into
&lt;img src=x onerror=&quot;alert(1)&quot;&gt;, which a browser shows as
text instead of running. In both cases the fix is the same: data goes
through a channel that cannot be read as code.`,
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
				Diagram: `"hunter2" + salt "x9Qa" ─▶ [ Argon2, slow on purpose ] ─▶ $argon2id$…Kq7
login: hash the typed password with the same salt, then compare`,
				UnderTheHood: `What one guess costs, on one core of this machine, in Python's hashlib:

  stored as                            one guess    10 million guesses
  SHA-256 with a salt                  0.5 µs       5 seconds
  scrypt (n = 2^14, r = 8, p = 1)      44 ms        5 days
  PBKDF2-SHA256, 600,000 iterations    153 ms       18 days

A fast hash is designed to be fast, so a leaked table of SHA-256
password hashes lets an attacker try every common password against
every account in minutes. A deliberately slow, salted hash makes the
same work take days per account, and attackers' GPUs are fast at
SHA-256 but slowed by scrypt's memory use. Your login pays the cost once
(a sixth of a second); someone trying billions of guesses pays it every
time. OWASP recommends 600,000 PBKDF2-SHA256 iterations, or Argon2id.`,
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
				Diagram: `Alice                                Bob
  │ encrypt with Bob's PUBLIC key       │
  │ ─────────── ciphertext ───────────▶ │ decrypt with Bob's PRIVATE key
  │                                     │
  │ ◀──────── message + signature ──────│ signed with Bob's PRIVATE key
  │ verify with Bob's PUBLIC key        │`,
				UnderTheHood: `Why the mode of a cipher matters, measured with openssl: 1 MB made of
one 16-byte block repeated (like a flat area of an image), encrypted
with AES-128 in two modes:

  mode                     ciphertext blocks   distinct
  ECB (each block alone)        65,536             1
  CTR (with a counter)          65,536        65,536

With ECB, equal plaintext blocks give equal ciphertext blocks, so the
picture's outline survives encryption; CTR, GCM and every modern mode
mix in a counter or nonce, and the output looks random. Cost is no
excuse either: AES-256-GCM ran at 10.8 GB/s on one core here, and an
ECDSA P-256 signature took 30 µs to make and 88 µs to check, against
RSA-2048's 0.39 ms to sign. Use an authenticated mode such as AES-GCM
from a well-tested library, never ECB, and never a cipher of your own.`,
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
				UnderTheHood: `Certificate checks, tested with openssl against a local server whose
certificate was signed by a private test authority:

  client trusts                       asks for       result
  the test authority                  localhost      Verify return code: 0 (ok)
  only the system's usual roots       localhost      error 20: unable to get
                                                     local issuer certificate
  the test authority                  example.com    error 62: hostname mismatch

A correct client refuses both failures: a certificate from an
authority it does not trust, and a valid certificate for a different
name. Code that turns verification off to make an error go away (curl
-k, verify=False, InsecureSkipVerify) accepts anyone at all as the
server, which is exactly what TLS exists to prevent. A full TLS 1.3
handshake cost about 0.8 ms of CPU here: there is no performance reason
to skip it either.`,
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
			{
				Name:    "Memory-Safety Bugs: Overflows & Use-After-Free",
				Summary: "When C trusts a length it should have checked.",
				Body: `C lets you write past the end of an array, read memory you already
freed, or use a pointer to something that no longer exists. The language
does not stop you; the behaviour is undefined. To an attacker, that is
an opportunity: a stack buffer overflow can overwrite the saved return
address, so when the function returns, the CPU jumps wherever the
attacker chose. Use-after-free and out-of-bounds reads leak secrets or
hijack control in the same way.

Defences come in layers. Compilers add stack canaries (a secret value
checked before returning), the operating system randomises where code
and data live (ASLR) and marks the stack non-executable (NX). But these
only make exploitation harder. The real fixes are bounded functions
(snprintf, fgets, never gets), checking every length, testing with
sanitizers and fuzzers, and choosing memory-safe languages (Rust, Go,
Java, Python) where the bug cannot happen.`,
				Diagram: `char name[8]; strcpy(name, input);    higher addresses ▶

 before: [ b o b \0 . . . . ][ canary ][ saved rbp ][ return addr ]
 attack: [ A A A A A A A A ][ A A A A  ][ A A A A A ][ A A A A A A  ]
                              ▲ changed: the program aborts at ret`,
				UnderTheHood: `Take main with char name[8] and strcpy(name, argv[1]), built with
gcc 13 on x86-64 Linux (gcc -O0, Intel syntax). The compiler lays out
the stack frame like this:

  mov  rax, QWORD PTR fs:40      ; load the secret canary
  mov  QWORD PTR -8[rbp], rax    ; store it just above name
  lea  rdx, -16[rbp]             ; name lives at rbp-16 .. rbp-9
  call strcpy@PLT
  mov  rdx, QWORD PTR -8[rbp]    ; before returning: reload it
  sub  rdx, QWORD PTR fs:40      ; still the same?
  call __stack_chk_fail@PLT      ; (only reached if it changed)

Run it with growing arguments:

  ./a.out alice              hello, alice
  ./a.out AAAA…A (16 A's)    *** stack smashing detected ***
                             terminated  (exit 134)
  -fno-stack-protector, 40   Segmentation fault  (exit 139)

Without the canary, gdb shows the crash on main's final ret: the saved
return address has become 0x4141414141414141 ("AAAAAAAA"), which is not
a valid address, so the CPU refuses to jump. An attacker writes a real
address there instead, and ret obeys.

Now the surprise: 8 letters (9 bytes with the final zero) is not
detected. The canary is random on every run, but glibc always makes
its lowest byte 00, so the stray terminating zero overwrites 00 with 00.
AddressSanitizer is not fooled:

  gcc -fsanitize=address ...; ./a.out AAAAAAAA
  ERROR: AddressSanitizer: stack-buffer-overflow ... WRITE of size 9

The canary catches big overflows when the function returns;
the sanitizer catches every out-of-bounds byte at the moment it happens.`,
				MentalModel: "Every buffer has a size; every write must prove it fits.",
				TryIt:       "Compile a program that copies argv[1] into char name[8] with strcpy, once normally and once with -fsanitize=address, and run both with a 20-character argument.",
				Analogy: `A form with eight boxes for your surname, printed right above the
"Pay to" line. Write a long enough name and your extra letters spill into
the payment line. A careful clerk (bounds checking) refuses the form; a
careless one pays whoever the spill-over names.`,
				Example: `The 1988 Morris worm spread partly through a buffer overflow in the
fingerd service, which read a request with gets(). Heartbleed (2014) was
a missing length check in OpenSSL that let anyone read up to 64 KB of a
server's memory per request, including private keys and passwords.
Microsoft and Google's Chromium team have each reported that about 70%
of their serious security bugs are memory-safety bugs.`,
				Exercises: trio(
					`Why is gets(buf) impossible to use safely, while fgets(buf, sizeof buf, stdin) is fine?`,
					"gets cannot know how big buf is; fgets is told the size and stops there.",
					"Write a function that copies a string into a fixed-size buffer and can never overflow, returning an error when the input is too long. Test it with strings of length 0, 7, 8 and 100, compiled with -fsanitize=address,undefined.",
					"snprintf(dst, size, \"%s\", src) returns the length it wanted to write; compare it with size.",
					"Solve the first levels of the pwn.college \"Memory Errors\" module or picoCTF's buffer-overflow challenges, then write up for each: the bug, why it was exploitable, and the one-line fix.",
					"These platforms exist so you can practise legally. Never try this on systems you do not own.",
				),
			},
			{
				Name:    "Access Control: Who May Do What",
				Summary: "Logged in is not the same as allowed.",
				Body: `Authentication proves who someone is; access control (authorization)
decides what they may do. Broken access control tops the OWASP Top 10
because it is easy to forget: a page hides the Delete button from normal
users, but the server still accepts the delete request from anyone who
sends it. An insecure direct object reference (IDOR) is the classic
case: /invoices/1042 returns someone else's invoice because the server
checks only that you are logged in.

Rules that work: enforce every check on the server, never only in the
interface; deny by default and allow explicitly; check ownership on
every object, not just the role; give each account and service the least
privilege it needs; and log every refused request, because a run of 403
errors is often an attack in progress. Role-based access control (RBAC)
groups permissions into roles such as viewer, editor and admin.`,
				Diagram: `request ─▶ who are you?          ── no ─▶ 401 Unauthorized
               │ yes
               ▼
           may YOU do THIS       ── no ─▶ 403 Forbidden (log it)
           to THIS object?
               │ yes
               ▼
             do it`,
				UnderTheHood: `Unix file permissions, tested by a program that drops to the
unprivileged nobody account before trying to read three files:

  file mode                  as nobody     as root
  0644 (everyone may read)   read          read
  0640 (owner and group)     refused       read
  0600 (owner only)          refused       read

The kernel checks every open, every time, which is complete mediation
in practice; and root passes every check, which is why services run as
their own unprivileged accounts. The same idea applies inside
applications: a web request for invoice 1042 must be checked against
the logged-in user on the server, not trusted because the page only
showed that user their own invoice numbers.`,
				MentalModel: "For every request ask three things: who is this, what are they doing, and do they own it?",
				TryIt:       "In any web app you built, log in as one user, copy a URL containing an ID, then open it logged in as another user. Does the server refuse?",
				Analogy: `A hotel key card. The front desk checks your passport once
(authentication). After that, the lock on each door checks whether your
card belongs to that room (authorization). A hotel whose locks only
check "is this a real hotel card?" lets every guest into every room.`,
				Example: `In 2019 the US title-insurance company First American Financial exposed
about 885 million documents, including bank account numbers and Social
Security numbers, because anyone with a link to one document could
change the number in the URL and read others. No login was needed.`,
				Exercises: trio(
					`A page shows "Edit" only to admins, but the server's /edit endpoint does not check the role. Is the app secure? Explain.`,
					"Anyone can send the request directly, with curl or the browser's developer tools.",
					"Add role-based access control to a small web app: viewer, editor and admin roles, a single function that decides every permission, a default of deny, and tests that prove an editor cannot delete and a viewer cannot edit someone else's post.",
					"Keep one can(user, action, object) function and call it on every request; test the denials, not just the successes.",
					"Complete the access-control labs in PortSwigger's Web Security Academy, including at least one IDOR and one privilege-escalation lab, and write a short checklist you would use to review someone else's API.",
					"The labs run in your browser against PortSwigger's own servers, so they are legal to attack.",
				),
			},
			{
				Name:    "Network Attacks & Defences",
				Summary: "Sniffing, spoofing, scanning and flooding, and the walls against them.",
				Body: `On a shared network, anyone can try to read (sniff) or change traffic.
On open Wi-Fi without encryption, a neighbour can see what you send. An
attacker can pretend to be the router (ARP spoofing) or answer DNS
questions with fake addresses, placing themselves in the middle. TLS
defeats this, as long as nobody clicks through certificate warnings.

Before attacking, attackers map: a port scan (nmap) finds which services
answer, and each open port is a door to check. Availability attacks
(denial of service) flood a service until real users cannot get in.

Defences: expose as few services as possible, and put a firewall in
front that denies by default. Separate networks into zones
(segmentation) so one compromised laptop cannot reach the database.
Encrypt everything in transit, patch internet-facing software first, and
watch the traffic for anything unusual.`,
				Diagram: `internet ─▶ [ firewall: 443 only ] ─▶ web server   (zone 1)
                                          │
                                   [ firewall ]
                                          │
                                          ▼
                                       database     (zone 2)

a port scan from outside finds only 443; the database is never
reachable from the internet`,
				UnderTheHood: `Rate limiting as a defence, measured in a private network namespace with
nftables, against a local server receiving new connections as fast as
one client could open them:

  firewall                                        accepted a second
  no limit                                              14,101
  limit rate over 100/second drop (new only)                96

One rule held new connections to the limit, while connections already
open were untouched. That is how servers keep a flood of connection
attempts, or one misbehaving client, from using up everything; real
deployments limit per source address and in several layers (the
network, the load balancer, the application), since a single limit for
everyone also locks out legitimate users during a flood.`,
				MentalModel: "Every open port is a door; close the ones you do not use and lock the rest.",
				TryIt:       "Run nmap -sV scanme.nmap.org (the Nmap project allows scanning that host) and nmap localhost, and explain every open port you find.",
				Analogy: `A castle: a moat and one gate (the firewall), guards who check everyone
at the gate (authentication), an inner keep for the treasure
(segmentation), and lookouts on the walls (monitoring). A secret tunnel
nobody remembers is the forgotten open port.`,
				Example: `In 2016 the Mirai botnet took over hundreds of thousands of cameras and
routers that still used factory-default passwords, then flooded the DNS
provider Dyn, making sites such as Twitter and Netflix unreachable for
many users for hours. In 2010 the Firesheep browser extension showed
that anyone on café Wi-Fi could hijack other people's Facebook and
Twitter sessions, which pushed big sites to use HTTPS everywhere.`,
				Exercises: trio(
					"Your laptop is on café Wi-Fi. Which of these can the person at the next table read: the websites you visit over HTTPS, the pages' contents, your DNS lookups, an old plain-HTTP site you log into?",
					"HTTPS hides contents but not which server you talk to; plain HTTP hides nothing.",
					"Capture your own traffic with Wireshark while loading an HTTP page and an HTTPS page. Find the DNS query, the TCP handshake and the TLS handshake, and show that the HTTPS page's content is unreadable.",
					"Use the display filters dns, tcp.flags.syn==1 and tls.handshake.",
					"Harden a small Linux server or VM you own: firewall denying all inbound traffic except SSH (keys only, no passwords) and your web app, automatic security updates, and fail2ban. Scan it from outside with nmap before and after and compare.",
					"ufw or nftables for the firewall; in /etc/ssh/sshd_config set PasswordAuthentication no.",
				),
			},
			{
				Name:    "Secrets & the Software Supply Chain",
				Summary: "Your code is only as safe as its keys, dependencies and build.",
				Body: `A modern program is mostly other people's code: hundreds of packages,
a compiler, a build server and an update channel. A supply-chain attack
compromises any of these, and the malicious code arrives through a
channel you trust. Attackers publish look-alike packages
(typosquatting), take over abandoned ones, or break into build systems.

Secrets are the other half: API keys, database passwords and signing
keys. Committed to a repository, a secret is public forever, even if a
later commit deletes it, and bots scan public repositories for keys
within minutes.

Defences: keep secrets out of code (environment variables, a secrets
manager), scan for them before pushing, revoke and rotate any that leak.
Pin dependency versions with a lockfile, update them deliberately, and
watch for security advisories. Sign and checksum what you release, and
build it in a clean, automated pipeline. This academy's own releases
publish SHA256 checksums and signed build-provenance attestations for
exactly this reason.`,
				Diagram: `your code ──┐
packages ───┼──▶ build server ──▶ release file ──▶ user's machine
compiler ───┘        ▲                 ▲
  each arrow is a place to attack; checksums and signatures
  let the user check that nothing was swapped on the way`,
				UnderTheHood: `Can you check that a release was built from its published source?
Building this academy twice, measured with SHA-256:

  build                                               same binary?
  twice in the same folder, same flags                yes
  from a copy of the source in another folder         no
  both with -trimpath and -buildvcs=false             yes

The copy differed only because Go stamps the Git commit (and whether
the tree had uncommitted changes) into the binary, and the copy had no
Git folder. With paths and that stamp left out, the builds were
byte-for-byte identical. That is a reproducible build: anyone can
rebuild a release from its source and compare the hash, so a tampered
build server cannot slip in a different binary unnoticed.`,
				MentalModel: "Trust is transitive: everything you depend on can act as you.",
				TryIt:       "Run your language's audit tool (npm audit, pip-audit or govulncheck) on a project, and read one advisory in full.",
				Analogy: `Poisoning the flour at the mill instead of breaking into every bakery.
The bakers did nothing wrong except trust their supplier, which is why
careful bakeries check each delivery's seal.`,
				Example: `In 2020 attackers inserted a backdoor into the build of SolarWinds'
Orion software, and about 18,000 customers installed the signed,
official update. In 2024 a developer noticed that SSH logins had become
half a second slower and uncovered a backdoor in xz-utils, planted over
two years by a contributor who had gained the maintainer's trust. In
2016 attackers found cloud credentials in a private GitHub repository
used by Uber's engineers and stole data on 57 million people.`,
				Exercises: trio(
					"You committed an API key, then deleted it in the next commit. Is it safe now? What must you do?",
					"It is still in Git history and possibly already scraped. Revoke it and issue a new key.",
					"Add secret scanning and dependency auditing to one of your repositories: a pre-commit hook or CI step with gitleaks, plus your language's audit tool, and make CI fail on a finding.",
					"Test it on a throwaway branch with a made-up key in a format the scanner's rules list, then delete the branch.",
					"Verify a real release end to end: download an academy archive, check it against SHA256SUMS, then run gh attestation verify on it. Write down which attacks each check stops, and which it cannot.",
					"A checksum from the same server as the file proves integrity, not origin; the signed attestation proves who built it.",
				),
			},
			{
				Name:    "Detection & Incident Response",
				Summary: "Assume someone will get in; notice fast and respond calmly.",
				Body: `Prevention eventually fails, so defenders also need to detect and
respond. Detection starts with logs: logins (successes and failures),
permission denials, administrator actions and unusual network traffic,
sent to one place where they cannot be quietly deleted. Alerts watch for
patterns such as many failed logins, a login from a new country followed
by a password change, or a server suddenly sending gigabytes out.

When something is found, incident response follows the classic NIST
cycle: prepare (plans, contacts, backups you have actually tested),
detect and analyse, contain (isolate machines, disable accounts),
eradicate and recover, then learn in a blameless post-mortem. Keep
notes with timestamps from the first minute, and preserve evidence:
wiping a machine first destroys the record of how the attacker got in.`,
				Diagram: `prepare ─▶ detect & analyse ─▶ contain ─▶ eradicate & recover
   ▲                                                   │
   │                                                   ▼
   └────────── improve defences ◀──────────── lessons learned`,
				UnderTheHood: `File-integrity monitoring, the simplest detection there is: record a
SHA-256 for every program, then compare later. A baseline of this
machine's /usr/bin:

  files hashed        707
  data read           661 MB
  time taken          3.6 s

A few seconds a day finds any changed system program, but only if the
baseline is kept where an intruder on the machine cannot rewrite it
(another machine, or read-only storage), and only if someone looks at
the differences. Tools such as AIDE and Tripwire do exactly this, and
package managers can do the same for their own files (dpkg --verify,
rpm -V).`,
				MentalModel: "You cannot respond to what you never logged.",
				TryIt:       "On Linux, run journalctl -u ssh (or read /var/log/auth.log) on a machine with SSH open to the internet, and count the failed login attempts from strangers.",
				Analogy: `A smoke detector and a fire drill. The detector (logging and alerts)
turns a hidden fire into a loud one; the drill (the response plan) means
that when it sounds, nobody wastes the first ten minutes arguing about
who should call whom.`,
				Example: `In 1986 Cliff Stoll, a Berkeley astronomer turned systems administrator,
followed up a 75-cent accounting error and uncovered a spy selling
military data, a story he told in The Cuckoo's Egg. In the 2017 Equifax
breach, attackers went unnoticed for about 76 days, partly because an
expired certificate had silently switched off a device that inspected
outgoing traffic.`,
				Exercises: trio(
					"Put these steps in order and say what goes wrong if you skip each: contain, lessons learned, detect, recover, prepare.",
					"Skipping preparation means improvising under stress; skipping lessons means it happens again.",
					"Write a small log analyser for an SSH or web-server log: count failed logins per IP address, flag any address with more than 20 failures in 10 minutes, and print a short report. Test it on a real or sample log file.",
					"Parse the timestamp, IP and outcome from each line into a record, then slide a 10-minute window per IP.",
					"Write an incident-response plan for a small online shop: who is on call, how to reach them, the first 30 minutes step by step for \"customer passwords may have leaked\", what to tell customers and when, and a template for the post-mortem. Then run it as a 20-minute tabletop exercise with a friend.",
					"Many countries require reporting some breaches within a deadline (72 hours under the EU's GDPR); put that in the plan.",
				),
			},
			{
				Name:    "Reading Research Papers",
				Summary: "The frontier of computing lives in papers; learn to read one without drowning in it.",
				Body: `Textbooks are years behind; the newest ideas, in security and
everywhere else, arrive first as papers. Reading one is a skill, not a
talent, and it is not read front to back.

The three-pass method (Keshav): first pass, five minutes, read the
title, abstract, introduction, section headings and conclusion, and
decide whether to go on. Second pass, up to an hour, read the whole
thing but skip the heavy proofs, studying the figures and tables, which
carry most of the result, and marking terms to look up. Third pass, for
a paper you must truly understand, rebuild it: re-derive the argument,
re-implement the core, and note where you would have done it
differently. Most papers deserve only the first pass.

Read actively: a paper makes a claim, offers evidence and compares with
what came before. Ask what exactly is new, what the threat model or
assumptions are, what the evaluation does not cover, and whether you
believe the result. Find papers through Google Scholar, arXiv and a
venue's proceedings; follow a paper backwards through its references and
forwards through "cited by". Preprints are not peer-reviewed, and a
single paper is a claim, not a settled fact.`,
				Diagram: `pass 1 (5 min)   title, abstract, intro, headings, conclusion: read on?
pass 2 (1 hr)    whole paper, skip proofs, study figures and tables
pass 3 (hours)   re-derive and re-implement; find hidden assumptions
                 most papers: stop after pass 1`,
				UnderTheHood: `The best test of a paper is to reproduce one of its numbers. Burton
Bloom's 1970 paper predicts the false-positive rate of his filter; in the
modern form, (1 − e^(−kn/m))^k. A Bloom filter of 100,000 bits with 7
hash functions and 10,000 keys, built in Python and probed with 200,000
keys that were never added:

  measured false positives      0.79%
  the formula                   0.82%

Fifty years later, the claim holds on a laptop in a few lines of code.
When a reproduction does not match, that is worth a careful look: a
misread parameter, an assumption the paper states in passing (here,
independent hash functions), or occasionally an error in the paper.`,
				MentalModel: "A paper is an argument to interrogate, read in passes, not a chapter to absorb front to back.",
				TryIt:       "Give a famous short paper (Dijkstra's \"Go To Statement Considered Harmful\", or the original RSA paper) a five-minute first pass and write down its one claim.",
				Analogy: `Watching a match before studying the players. First you see the whole
game and the final score (abstract and conclusion); only if it matters
do you rewatch in slow motion to see exactly how each goal was built
(the method and proofs).`,
				Example: `Every technology in this degree began as a paper: Shannon's 1948 theory
of communication, Codd's 1970 relational model, Diffie and Hellman's
1976 key exchange, and the 2017 "Attention Is All You Need", which
started the large-language-model era. Security research is published at
venues such as USENIX Security, IEEE S&P ("Oakland"), ACM CCS and NDSS,
and the talks at Black Hat and DEF CON.`,
				Exercises: trio(
					"For a paper of your choice, write one sentence each for: its claim, its evidence, and what it does not address.",
					"The abstract gives the claim; the evaluation section gives the evidence and, read carefully, its limits.",
					"Give three papers a first pass in 15 minutes total and rank them by how much you would gain from a second pass. Pick one and do the second pass, listing every term you had to look up.",
					"Figures and tables first: they usually carry the result more clearly than the prose.",
					"Re-implement the core idea of a short systems or security paper (a data structure, an algorithm, a small protocol) from its description alone, then compare your version with the authors' code if they released it.",
					"If your implementation behaves differently, you have found either a bug of yours or an unstated assumption of theirs; both are worth writing down.",
				),
			},
			{
				Name:    "Reading CVEs, CWEs & Advisories",
				Summary: "The shared language defenders use to name, score and prioritise known vulnerabilities.",
				Body: `When a vulnerability becomes public it gets a CVE ID, such as
CVE-2021-44228 (Log4Shell): a unique name everyone can refer to. The
record names the affected product and versions, links to the fix, and
usually cites a CWE, the class of weakness (CWE-79 cross-site scripting,
CWE-89 SQL injection, CWE-416 use-after-free, CWE-787 out-of-bounds
write): the same bugs this stage has been teaching.

Severity is scored with CVSS, from 0.0 to 10.0. The base score is
computed from a vector string such as
CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H: how the vulnerability is
reached (Attack Vector: Network, Adjacent, Local, Physical), how hard
(Attack Complexity), what privileges and user interaction it needs, and
its impact on confidentiality, integrity and availability. A high score
is not the same as high risk to you: a 9.8 in software you do not run,
or that is not reachable from your network, may matter less than a 6.5
on your public login page.

Prioritise with more than the number. CISA's Known Exploited
Vulnerabilities (KEV) catalogue lists what attackers are using right
now, and EPSS estimates the probability a vulnerability will be
exploited. "Patch the highest CVSS first" is a worse rule than "patch
what is being exploited and reachable first".`,
				Diagram: `CVE-2021-44228                  the name (year + number)
  └ CWE-917  expression-language injection   the weakness class
  └ CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H  = 10.0   the score
  └ affected: Log4j 2.0-beta9 .. 2.14.1;  fixed in 2.15.0 (then 2.17.1)
  └ on CISA KEV: yes  → being exploited: patch now`,
				UnderTheHood: `CVSS v3.1 base scores computed from the specification's formulas, in a
few lines of Python, for three vectors:

  vulnerability                     vector                          score
  Log4Shell (CVE-2021-44228)        AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H  10.0
  Heartbleed (CVE-2014-0160)        AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N   7.5
  a local bug needing a user click  AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H   7.8

The first two match the NVD's published scores. The third shows the
score's limits: a local bug that needs someone to open a file scores
above Heartbleed, which leaked private keys from a sixth of the web's
secure servers. The score measures the bug's properties, not whether
anyone is exploiting it, which is what CISA's catalogue of known
exploited vulnerabilities adds.`,
				MentalModel: "A CVE names the bug, a CWE names its family, CVSS scores it, and KEV tells you if it is on fire.",
				TryIt:       "Look up CVE-2021-44228 on the NVD and read its CVSS vector field by field.",
				Analogy: `A disease registry. The CVE is the case number, the CWE is the diagnosis
("it's a kind of pneumonia"), CVSS is the severity triage score, and the
KEV catalogue is the public-health alert that this strain is actively
spreading right now.`,
				Example: `Log4Shell (CVE-2021-44228) scored 10.0 and was in a logging library
buried inside thousands of applications, so the hard part was not the
score but finding everywhere it was used. The Equifax breach exploited
CVE-2017-5638 in Apache Struts, a patch that had been available for
months: the vulnerability was known, the patching was not done.`,
				Exercises: trio(
					"Two vulnerabilities score CVSS 9.8 and 6.1. One is in a library on your internet-facing server; the other needs an attacker already logged in as an admin. Which do you patch first, and why is the score not enough?",
					"Reachability and context beat the raw number: an exploitable, reachable 6.1 can outrank an unreachable 9.8.",
					"Parse a CVSS v3.1 base vector yourself and compute the base score, then check it against the official calculator for three real CVEs.",
					"The base score combines the exploitability metrics and the impact metrics with a defined formula; the specification gives the exact weights.",
					"Take a project with dependencies, list its known vulnerabilities with your language's audit tool, and for each decide patch now, patch soon or accept, writing down the reasoning (reachable? exploited in the wild per KEV? a fix available?).",
					"A vulnerability in code you never call may be a lower priority than a medium one on a hot path; record why, so the decision can be reviewed.",
				),
			},
			{
				Name:    "Vulnerability Research & Coordinated Disclosure",
				Summary: "How researchers find a flaw nobody knew about, and get it fixed without causing harm.",
				Body: `A zero-day is a vulnerability the vendor does not yet know about, so
there is no patch. Finding one before an attacker does, and getting it
fixed, is the work of a security researcher, and it is a legitimate,
well-paid career. It is also bounded by law and ethics: in most
countries, testing a system you do not own or have permission to test is
a crime (the US Computer Fraud and Abuse Act, the UK Computer Misuse
Act), whatever your intention. Everything below is practised only on
your own systems, on deliberately vulnerable targets built for it, or
under the explicit scope of a bug-bounty or engagement.

How researchers find unknown bugs, all of which you have met in this
degree:
- fuzzing: feed a program millions of malformed or random inputs under a
  sanitizer (AddressSanitizer, UBSan) and watch for crashes. Modern
  fuzzers (AFL++, libFuzzer) guide themselves by code coverage. This is
  how most memory-safety bugs are now found.
- code audit: read the source for the patterns from this stage, missing
  length checks, integer overflow, unchecked return values, trust of
  user input.
- diffing patches: compare a program before and after a fix to see what
  changed, which can reveal a related, still-unpatched bug.

A crash is not yet a finding. You minimise the input to the smallest
that still triggers it, work out the root cause (which line, which
assumption), and assess impact: can it be made to disclose memory, or
only to crash? That assessment is where judgement, and restraint, matter.

Then you disclose, coordinated. Report privately to the vendor (or
through a CERT or a bug-bounty platform), give them a reasonable time to
fix it (90 days is the common norm, as Google Project Zero uses), and
publish only after a fix is available, or the deadline passes with none.
Selling to the highest bidder, or publishing a working exploit against
unpatched software people rely on, causes the harm the whole field
exists to prevent.`,
				Diagram: `find ─▶ reproduce ─▶ minimise ─▶ root cause ─▶ assess impact
                                                   │
                                                   ▼
          report privately to the vendor ─▶ they fix it ─▶ CVE issued
                              │  (≈90 days)            │
                              └─ no fix in time ───────┴─▶ publish`,
				UnderTheHood: `Finding and minimising a bug in a toy parser written for the purpose,
with a one-character off-by-one in its bounds check (a length of 32 is
let through for a 32-slot buffer that also needs a terminator):

  random inputs tried before the first crash        73
  size of that input                                118 bytes
  after minimising (dropping pieces while it
    still crashes)                                   33 bytes: a length
                                                     of 32, then 32 bytes

The minimised input is the report: it names the exact boundary, lets the
maintainer reproduce it in seconds, and shows the fix (> 32 must be
>= 32). Coordinated disclosure sends that, privately, to the maintainer,
with a deadline for the fix, instead of a working attack published for
everyone.`,
				MentalModel: "Find it, prove it minimally, report it privately, and let the fix ship before the world hears.",
				TryIt:       "Read a published Project Zero write-up end to end and note the timeline from report to fix to disclosure.",
				Analogy: `Finding a faulty lock on a neighbour's door. The responsible thing is to
tell them quietly and give them time to change it, not to pick it to
prove you can, and not to post the technique while every door in the
street still has that lock.`,
				Example: `Google's Project Zero finds zero-days in widely used software and
publishes after a 90-day deadline, which pushed the industry toward
timely patching. OSS-Fuzz runs fuzzers continuously against hundreds of
open-source projects and has reported tens of thousands of bugs to their
maintainers. Bug-bounty programmes (HackerOne, Bugcrowd, and vendors'
own) pay researchers for exactly this work, under rules that make it
legal; some single high-impact reports have paid over $100,000.`,
				Exercises: trio(
					"Put these in the right order and say why each matters: publish, minimise the input, report to the vendor, confirm the crash reproduces, assess impact.",
					"Reproduce, minimise, assess, report, then publish only after a fix or the deadline: skipping \"report first\" is what makes disclosure irresponsible.",
					"Write a small parser in C (for a config format, or a simple image header), then fuzz it with libFuzzer or AFL++ under -fsanitize=address,undefined. Minimise the first crashing input by hand to the fewest bytes that still crash it, find the root cause, and fix it. Keep the crashing input as a regression test.",
					"A libFuzzer target is one function, int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size); build with -fsanitize=fuzzer,address.",
					"Practise the full loop legally: pick a target that invites it (a pwn.college or picoCTF challenge, or a program with an open bug-bounty scope you have read), find an issue, and write it up as a disclosure report: summary, affected version, reproduction steps, impact, and suggested fix.",
					"A good report lets the maintainer reproduce the bug in minutes; that is what gets it fixed, and what earns a bounty.",
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
			{"Site", "OverTheWire: Bandit", "https://overthewire.org/wargames/bandit/", "A beginner wargame played over SSH; teaches the Linux command line as a security tool."},
			{"Course", "pwn.college (free)", "https://pwn.college/", "Arizona State University's hands-on course, from the command line to memory-corruption exploits."},
			{"Site", "OWASP Cheat Sheet Series", "https://cheatsheetseries.owasp.org/", "Short, practical guides for defending against each class of vulnerability."},
			{"Book", "Nmap Network Scanning (free online chapters)", "https://nmap.org/book/", "The official guide to port scanning, by Nmap's author."},
			{"Site", "CISA Known Exploited Vulnerabilities Catalog", "https://www.cisa.gov/known-exploited-vulnerabilities-catalog", "Vulnerabilities attackers are exploiting right now: what to patch first."},
		},
		Blueprints: []Blueprint{
			{"Password Manager", "A local, encrypted password vault with a master password, key derivation (Argon2) and authenticated encryption.",
				[]string{"Key derivation from the master password", "Encrypt/decrypt vault (AES-GCM)", "Add, get and generate passwords", "Threat model document"}},
			{"Secure Login Service", "A web login system with hashed passwords, sessions, rate limiting, TOTP 2FA and a security review.",
				[]string{"Sign-up and login", "Session cookies", "Rate limiting", "TOTP 2FA", "Self-audit against OWASP Top 10"}},
			{"CTF Journey", "Solve 30 beginner capture-the-flag challenges and write up how you solved each one.",
				[]string{"Web challenges", "Crypto challenges", "Forensics challenges", "Write-ups"}},
			{"Home Security Lab", "Two virtual machines on a private network: a hardened server and an attacker box, with centralised logs and an alert for brute-force logins.",
				[]string{"Isolated VM network", "Hardened SSH and firewall", "Central log collection", "Brute-force alert", "Attack yourself and read the logs"}},
			{"Fuzz a C Parser", "Fuzz a small C parser you wrote (for example a CSV or INI reader) with libFuzzer and AddressSanitizer, fix every crash, and keep the crashing inputs as regression tests.",
				[]string{"Write the parser", "Fuzz target", "Triage and fix crashes", "Regression tests in CI"}},
		},
		Quiz: []Question{
			{"What does the CIA triad stand for?", "Confidentiality, Integrity and Availability."},
			{"What is the correct fix for SQL injection?", "Parameterised queries (prepared statements), so input is never interpreted as SQL."},
			{"Why use bcrypt or Argon2 instead of SHA-256 for passwords?", "They are deliberately slow and salted, which makes mass guessing expensive."},
			{"With public-key crypto, which key signs a message and which verifies it?", "The private key signs; anyone can verify with the matching public key."},
			{"How can a stack buffer overflow give an attacker control of a program?", "By overwriting the saved return address, so the function returns to code the attacker chose."},
			{"What is the difference between a 401 and a 403 response?", "401: we do not know who you are (authenticate). 403: we know who you are, and you may not do this."},
			{"You leaked an API key in a public commit and deleted it. What now?", "Revoke it and issue a new one; it stays in Git history and has probably been scraped already."},
			{"Why should you not wipe a hacked server immediately?", "It destroys the evidence of how the attacker got in; isolate it and investigate first."},
			{"What is the three-pass method for reading a paper?", "Pass 1: title, abstract, intro, headings and conclusion, to decide whether to read on. Pass 2: the whole paper, skipping proofs, studying figures. Pass 3: re-derive and re-implement it."},
			{"What do a CVE ID, a CWE and a CVSS score each tell you?", "The CVE names a specific vulnerability; the CWE names its class of weakness; the CVSS score (0 to 10) rates its severity."},
			{"Why is the highest CVSS score not always what to patch first?", "A high score in software you do not run, or that an attacker cannot reach, can matter less than a lower-scored, reachable, actively exploited bug (see the CISA KEV catalogue)."},
			{"What is a zero-day?", "A vulnerability the vendor does not yet know about, so no patch exists; finding and responsibly reporting one before attackers use it is the work of a security researcher."},
			{"What is coordinated disclosure, and why the delay before publishing?", "Report the flaw privately to the vendor and give them a reasonable time (often 90 days) to fix it, publishing only after a fix exists, so attackers do not get a working exploit against unpatched systems."},
			{"Why is testing a system you do not own a crime even with good intentions?", "Laws such as the US CFAA and the UK Computer Misuse Act forbid unauthorised access regardless of intent; practise only on your own systems, deliberately vulnerable targets, or within an explicit bug-bounty or engagement scope."},
			{"A search box builds its SQL by joining strings. What does a name such as x' OR '1'='1 do, and what is the fix?", "It turns the condition true for every row; pass the input as a query parameter, so it can only ever be a value."},
			{"Why store passwords with PBKDF2, scrypt or Argon2 rather than SHA-256?", "SHA-256 is fast, so a leaked table can be guessed at millions a second; a deliberately slow, salted hash makes each guess cost milliseconds."},
			{"What does encrypting with AES in ECB mode leak?", "Equal plaintext blocks give equal ciphertext blocks, so patterns in the data survive encryption."},
			{"A TLS client reports \"hostname mismatch\". What does that mean, and is it safe to switch verification off?", "The certificate is valid but for another name; switching verification off would accept any server, which defeats TLS."},
		},
	},

	// -----------------------------------------------------------------
	11: {
		Overview: `What can computers do at all, and what can they do efficiently? This
stage explores the mathematical limits of computation: simple machines
that recognise patterns and the regex engines built on them, grammars
and stacks, the Turing machine that models every computer, problems that
no program can ever solve, the famous P vs NP question, reductions
between problems, and the heuristics that tame hard problems in
practice. The theory is measured, too: a regex that takes exponential
time, a DFA that is 2ⁿ times bigger than its NFA, a 5-state Turing
machine that runs for 47 million steps, and a Sudoku solved as logic.`,
		Outcomes: []string{
			"Use regular expressions and grammars confidently, knowing their limits",
			"Explain why some problems can never be solved by any program",
			"Recognise hard (NP-complete) problems and choose practical approaches",
			"Explain why some regex engines can hang and others cannot",
			"Turn a puzzle into a SAT problem, and judge a heuristic's answer against the optimum",
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
			{"NFA", "A nondeterministic finite automaton: it may have several possible next states, and accepts if any path does."},
			{"Backtracking", "Trying one choice, and undoing it to try the next when it fails; can take exponential time."},
			{"Reduction", "Turning instances of one problem into instances of another, so a solver for the second solves the first."},
			{"SAT", "The satisfiability problem: can the variables of a formula be set so that it is true? The first problem proved NP-complete."},
			{"Approximation ratio", "How far, at worst, an algorithm's answer can be from the optimum, such as \"at most twice the best\"."},
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
				UnderTheHood: `The pattern ^(a+)+$ matches strings of a's. Given a string of n a's
followed by a b, which cannot match, Python's re module (a backtracking
engine) tried every way of splitting the a's between the two +'s
before giving up:

  n = 16    0.004 s
  n = 18    0.018 s
  n = 20    0.064 s
  n = 22    0.200 s
  n = 24    0.847 s     about 4 times slower for every 2 more a's

At that rate n = 40 would take years. GNU grep -E, which runs the
pattern as an automaton, answered in 2.7 ms for n = 24, and still 2.7 ms
for n = 100,000: it reads each character once, keeping track of every
state the pattern could be in. Same pattern, same answer; the
difference is the algorithm, and it is the subject of the next lesson.`,
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
					`In C, implement a DFA simulator (a transition table int next[STATES][2], and an array of accepting states) and run both the "ends in ab" machine and your even-1s machine on 20 test strings.`,
					"The whole simulation is one loop: state = next[state][symbol] for each character.",
					"Write a log scanner that uses regular expressions to extract timestamps, IP addresses and status codes from a web-server log, and reports the number of errors per hour, with grep -E and awk, or in C with regcomp and regexec from regex.h.",
					"In regex.h, regexec fills an array of regmatch_t with the start and end of each parenthesised group.",
				),
			},
			{
				Name:    "NFAs, DFAs & How Regex Engines Work",
				Summary: "Nondeterminism is free in theory and costs either memory or time in practice; good engines never backtrack.",
				Body: `A nondeterministic finite automaton (NFA) may have several possible next
states for a symbol, or none, and may jump without reading anything (an
ε-move). It accepts if any path through it ends in an accepting state.
Thompson's construction (1968) turns any regular expression into an NFA
with at most about two states per symbol of the pattern.

There are two honest ways to run an NFA:
- simulate it: keep the set of states it could be in, and update the
  whole set for each input character. Time is O(pattern × text), never
  more, whatever the input. RE2, Go's regexp, Rust's regex crate and
  grep work this way.
- convert it to a DFA with the subset construction: each DFA state is a
  set of NFA states. Matching is then one table lookup per character,
  but the DFA can have up to 2ⁿ states for an n-state NFA, so engines
  build it lazily, only the states the text actually reaches.

Backtracking engines (Perl, Python, Java, JavaScript) instead try one
path at a time and back up on failure. That supports backreferences
like (a+)\1, which no automaton can, at the price of exponential worst
cases.`,
				Diagram: `regex a(b|c)*  ─Thompson─▶  NFA  ─subset construction─▶  DFA
                                                 {each DFA state =
simulate the NFA: a set of states per character   a set of NFA states}
backtrack:        one path at a time, undo on failure`,
				UnderTheHood: `The language "the n-th symbol from the end is 1" has an NFA with n + 1
states: wait in state 0, guess that this 1 is the one, then count n − 1
more symbols. A short program ran the subset construction on it and
counted the DFA states it produced:

  n =  1      2 NFA states      2 DFA states
  n =  3      4                 8
  n =  5      6                32
  n = 10     11             1,024
  n = 16     17            65,536

Exactly 2ⁿ: the DFA must remember the last n symbols, every one of the
2ⁿ combinations, because any of them could turn out to be the n-th from
the end. This is not an accident of the construction; it is proven that
no smaller DFA exists. The NFA simulation stays at 17 states for n = 16,
and that is why engines simulate, or build only the part of the DFA
they need.`,
				MentalModel: "Track every possible state at once, and the input can never make you go back.",
				TryIt:       "Draw the Thompson NFA for (a|b)*abb, then run the subset construction by hand and count the DFA states you get.",
				Analogy: `Searching a maze. A backtracking engine sends one explorer, who walks
down a corridor and, at a dead end, walks all the way back to try the
next one. An NFA simulation floods the maze with water: every corridor
fills at once, step by step, and you only watch where the water has
reached.`,
				Example: `Google built RE2 so that users could type regular expressions into Code
Search without being able to hang its servers, and Go's regexp package
and Rust's regex crate follow its design. Cloudflare's 2019 outage was
a backtracking regex; after it, the company moved its firewall rules to
an engine with guaranteed linear time. Lexer generators such as lex
and flex compile token patterns into DFAs ahead of time.`,
				Exercises: trio(
					"How many states does the Thompson NFA for a|b have? And for (a|b)*?",
					"Each symbol gets two states joined by its edge; | adds a new start and end with ε-moves to both branches (6 states); * adds two more with a loop back (8).",
					"In C, simulate an NFA for a pattern you build by hand: store the states and their transitions in arrays, keep the current set as a bool array, and update it for each input character, following ε-moves. Run it on the input that makes ^(a+)+$ hang and time it.",
					"Compute the ε-closure (every state reachable by ε-moves) each time you add a state to the set.",
					"Write a lazy DFA: start from the NFA, create each DFA state (a set of NFA states) only when the input first reaches it, cache it in a hash table, and count how many of the possible states a real text actually creates.",
					"Represent a set of NFA states as a bitmask (or a sorted list) so that equal sets hash to the same DFA state.",
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
simplest to the most powerful.

A recursive-descent parser uses the call stack as its pushdown stack:
each level of nesting is one more function call. That makes nesting
depth a real limit, and a hostile input that is nothing but opening
brackets can crash a parser that does not set one.`,
				Diagram: `recursively enumerable     ◀ Turing machines
  └ context-sensitive
      └ context-free         ◀ pushdown automata (code, JSON, HTML)
          └ regular          ◀ finite automata, regexes`,
				UnderTheHood: `How deep can real JSON parsers nest? Each was given n "[" followed by
n "]":

  Python json    depth 990 ok, depth 1,000: RecursionError
                 (its recursion limit is 1,000 calls)
  Go json        depth 10,000 ok, depth 10,001:
                 "exceeded max depth"
  Node.js        depth 100,000 ok

Three parsers, three limits, and each one is a pushdown automaton
running out of stack, or refusing to before it does. Go's limit is a
deliberate rule against stack exhaustion from hostile input; Python's
is its general recursion limit. No regular expression could check
these inputs at all, at any depth, because matching [ to ] needs to
count, and a finite automaton cannot.`,
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
					"In C, write a recursive-descent parser and evaluator for: Expr → Term (('+'|'-') Term)*, Term → Factor (('*'|'/') Factor)*, Factor → number | '(' Expr ')'. Give it a depth limit, and test it on 100,000 opening brackets.",
					"One function per rule, each consuming tokens; pass the depth down and fail cleanly when it passes the limit.",
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
				UnderTheHood: `How long can a tiny Turing machine run before it stops? The "busy
beaver" question asks, for machines with n states and two symbols,
started on a blank tape, which halting one runs longest. This
5-state machine, found by Marxen and Buntrock in 1989:

  A: 0 → 1 R B    1 → 1 L C
  B: 0 → 1 R C    1 → 1 R B
  C: 0 → 1 R D    1 → 0 L E
  D: 0 → 1 L A    1 → 1 L D
  E: 0 → 1 R halt 1 → 0 L A

was simulated in C, its tape an array:

  halted after 47,176,870 steps, 4,098 ones, 12,289 cells visited
  0.21 s at gcc -O2; 1.59 s with sanitizers

Ten rules, 47 million steps. In 2024 a collaborative project proved
with the Coq proof assistant that no 5-state machine that halts runs
longer: BB(5) = 47,176,870. For 6 states the known champion runs for
more steps than could be written in decimal with every atom in the
observable universe, and the general question is undecidable, which is
the next lesson.`,
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
					"In C, write a Turing-machine simulator (the tape as an array with the head starting in the middle, the rules as a table rule[state][symbol]) and program it to increment a binary number.",
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

trouble(trouble): if it halts, it loops; if it loops, it halts. Absurd.`,
				UnderTheHood: `Undecidability is not only about contrived programs. This loop:

  while (x != 1) x = x % 2 ? 3 * x + 1 : x / 2;

halts for every starting x anyone has ever tried, but nobody can prove
it halts for all of them: that is the Collatz conjecture, open since
1937. A C program ran it for every x below 10 million in 3.2 seconds:

  longest:  8,400,511 takes 685 steps to reach 1
  highest:  starting from 6,631,675, x climbs to 60,342,610,919,632

Testing more numbers can never settle it, and a perfect halting checker
would. Real tools therefore approximate. gcc -fanalyzer, given a
function that frees a pointer only when argc > 5 and then uses it:

  warning: use after 'free' of 'p' [CWE-416]
  warning: dereference of possibly-NULL 'p' [CWE-690]

It reasoned about both paths without running them, and also pointed
out the unchecked malloc. On other programs it will miss bugs or raise
false alarms: no tool can avoid both.`,
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
					`Write a "loop detector" in C that runs a function in a child process with a time limit (fork, alarm, waitpid) and reports "probably infinite". Construct a function it wrongly flags, and explain why no time limit can be right for all programs.`,
					"A long but finite computation beats any fixed limit.",
					"Run a static analyser on one of your projects (gcc -fanalyzer or clang --analyze for C, go vet, or clippy) and classify three of its warnings as a true bug, a false alarm or style. Explain why the tool cannot be perfect.",
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
				UnderTheHood: `The travelling-salesperson problem by brute force in C: random cities,
the first one fixed, every order of the others tried:

  cities   tours          time
       8        5,040     under 1 ms
      10      362,880     0.025 s
      11    3,628,800     0.197 s
      12   39,916,800     2.27 s

Each extra city multiplies the time by the number of cities: about 17.6
million tours a second, so 20 cities (19! ≈ 1.2 × 10¹⁷ tours) would
take around 220 years. Checking a proposed tour, adding up n distances
and comparing with a target length, takes microseconds at any size.
That gap, between finding and checking, is exactly the difference
between P and NP, if they are different.`,
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
					"In C, solve the travelling-salesperson problem for 8 cities by brute force (all permutations) and with the nearest-neighbour heuristic. Compare route lengths and running times, then try 11 and 12 cities.",
					"With the start fixed, 11 cities give 10! = 3,628,800 routes.",
					"Plan a delivery route through 15 real addresses in your town (approximate coordinates are fine). Implement nearest-neighbour plus 2-opt improvement, and compare it with the order you would intuitively drive.",
					"2-opt repeatedly reverses a segment of the route whenever that makes the route shorter.",
				),
			},
			{
				Name:    "Reductions & NP-Completeness",
				Summary: "Solve a new problem by translating it into one you already have a solver for.",
				Body: `A reduction turns any instance of problem A into an instance of problem
B, quickly, so that the answers agree. Then a solver for B also solves
A, and B is at least as hard as A. Reductions are how hardness is
proved and how hard problems are solved in practice.

In 1971 Stephen Cook (and independently Leonid Levin) proved that every
problem in NP reduces to SAT, the question of whether a Boolean formula
can be made true. SAT was the first NP-complete problem; since then
thousands more have been shown NP-complete by reducing a known one to
them: 3-SAT, graph colouring, clique, Hamiltonian cycle, subset sum,
Sudoku of any size.

The practical side is just as important. SAT solvers based on DPLL and
its modern descendant, conflict-driven clause learning (CDCL), solve
real formulas with millions of variables. So translating a scheduling,
verification or puzzle problem into SAT is often the fastest way to
solve it, even though the worst case is exponential.`,
				Diagram: `Sudoku grid ──encode──▶ SAT formula ──solver──▶ true/false assignment
                 │                                          │
 variable x(r,c,d) = "cell (r,c) holds digit d"   decode ─▶ filled grid
 one digit per cell; each row, column and box holds each digit once`,
				UnderTheHood: `A Sudoku becomes SAT with one variable per (row, column, digit): 729
variables. The rules become clauses:

  every cell holds at least one digit            81 clauses
  … and at most one (no two digits together)  2,916
  every row, column and box has each digit       243
  … and at most once                           8,748
  total                                       11,988, plus one per given

A short DPLL solver in Python (unit propagation, then branch on a
literal from the shortest clause) took the puzzle Arto Inkala published
in 2012 as "the world's hardest Sudoku", 21 givens:

  solved in 3.1 s: 863 decisions, 30,771 unit propagations
  812753649 943682175 675491283 154237896 369845721 …

the published solution. The 60 empty cells allow 9⁶⁰ ≈ 1.8 × 10⁵⁷ grids
to brute-force; unit propagation, which fills in every forced value
for free, cut the search to 863 choices.`,
				MentalModel: "If you can translate it into SAT, a SAT solver can probably solve it.",
				TryIt:       "Write the clauses that say \"exactly one of x1, x2, x3 is true\" in CNF, and check them against all 8 assignments.",
				Analogy: `Translating a letter into a language that has an excellent translator,
rather than learning to write that letter yourself. The reduction is
the translation; the SAT solver is the expert. Proving hardness works
the same way in reverse: if your new problem can be used to answer a
problem everyone already knows is hard, yours is hard too.`,
				Example: `Hardware companies verify chip designs by reducing "can this circuit
ever reach a bad state?" to SAT. Linux package managers such as
openSUSE's zypper (libsolv) and Fedora's dnf solve package
dependencies as SAT problems. In 2016 a SAT solver settled the Boolean
Pythagorean triples problem, open for decades, with a proof 200
terabytes long.`,
				Exercises: trio(
					"Encode \"A and B cannot both be on duty, and at least one of A, B, C must be\" as clauses in conjunctive normal form.",
					"(¬A ∨ ¬B) ∧ (A ∨ B ∨ C).",
					"In C, write a DPLL SAT solver that reads a formula in the DIMACS CNF format, with unit propagation and backtracking, and test it on small satisfiable and unsatisfiable formulas, then on a Sudoku you encode yourself.",
					"Keep the assignment in an array indexed by variable, and undo the assignments of a branch when it fails.",
					"Reduce a real scheduling problem you care about (exam timetables, a sports league, shifts at work) to SAT: choose the variables, write the constraints as clauses, solve it with your solver or a free one such as MiniSat or CaDiCaL, and decode the answer.",
					"\"At most one of n\" takes n(n−1)/2 pairwise clauses; for large n, look up the sequential-counter encoding.",
				),
			},
			{
				Name:    "Approximation & Heuristics for Hard Problems",
				Summary: "When the exact answer costs centuries, take a provably close one, or a usually excellent one, in milliseconds.",
				Body: `For NP-hard optimisation problems, three practical strategies:
- approximation algorithms, with a guarantee: the answer is never worse
  than some factor of the optimum. Picking both ends of any uncovered
  edge gives a vertex cover at most twice the smallest; Christofides'
  algorithm gives a metric TSP tour at most 1.5 times the shortest.
- heuristics and local search, with no guarantee but often excellent
  results: start from a reasonable answer (nearest neighbour for TSP)
  and keep making small improvements (2-opt: reverse a segment if that
  shortens the tour) until none helps.
- exact solvers that are fast in practice: branch and bound,
  integer-programming and SAT solvers, which prune most of the search
  on real inputs.

Some problems resist all three: for some, even approximating within any
constant factor is NP-hard. And when a problem must be solved exactly,
problem size is the lever: an exponential algorithm is fine for 20
items and hopeless for 60.`,
				Diagram: `cost                                        exact search
  ▲                                       ╱
  │                                    ╱
  │                              ╱ ─ ─ ─ ─ ─   heuristic: fast, good,
  │                   ╱ ─ ─ ─ ─                no guarantee
  │        ╱ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─   approximation: fast,
  └───────────────────────────────────────▶      within a factor
                                     problem size`,
				UnderTheHood: `The same random 11-city instances solved exactly (3,628,800 tours each)
and with two heuristics, six layouts:

  nearest neighbour        +0.0% to +17.6% above the optimum
  nearest neighbour, 2-opt +0.0% to +1.0%, optimal in 5 of 6

Nearest neighbour alone was once 17.6% too long; 2-opt then fixed
almost all of it. On 1,000 cities, where no exact answer is known:

  nearest neighbour   289,951
  after 2-opt         251,417   13.3% shorter, 7 passes, 0.145 s

A brute-force search of 1,000 cities would need 999! tours, a number
with 2,565 digits. A heuristic answer in a seventh of a second, close
to optimal on every case that could be checked, is why routing,
scheduling and chip design are solved problems in practice and open
problems in theory.`,
				MentalModel: "Good enough, fast, and with a known bound beats perfect and never.",
				TryIt:       "Draw five points on paper, find the shortest tour by trying them all, then run nearest neighbour from each starting point and compare.",
				Analogy: `Packing a car boot. Nobody computes the optimal arrangement of every
bag; you put the big ones in first, then fill the gaps, then swap two
bags if that makes room. You end up close to the best packing in a
minute instead of the perfect one next year.`,
				Example: `UPS's ORION routes its drivers with heuristics, not exact optimisation,
and airlines plan crews with integer programming solvers that are
exponential in the worst case but fast on their data. The Concorde TSP
solver has found exact optimal tours for instances with 85,900 cities,
using branch and cut, while 2-opt and Lin–Kernighan heuristics get
within a few percent in seconds.`,
				Exercises: trio(
					"The greedy vertex cover takes any edge with neither end covered and adds both ends. Why is the result at most twice the size of the smallest cover?",
					"The chosen edges share no vertices, and any cover must contain at least one end of each of them; greedy takes two.",
					"In C, implement nearest neighbour and 2-opt for TSP, and compare both with the exact answer from brute force on 50 random instances of 10 cities: report the average and the worst ratio.",
					"Fix the random seed for each instance so that both methods see the same cities.",
					"Take a real knapsack-style decision (packing for a trip by weight and value, or a budget across projects), solve it exactly with dynamic programming, and compare with the greedy \"best value per kilogram first\" heuristic. Find an instance where greedy is far from the optimum.",
					"Greedy fails when one heavy, valuable item would fit only if a few lighter ones were left out.",
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
			{"Why can ^(a+)+$ take exponential time in Python but not in grep?", "Python's engine backtracks through every way of splitting the a's between the two +'s; grep runs the pattern as an automaton, tracking all possible states at once, one pass over the text."},
			{"Why can converting an NFA to a DFA blow up, and how do engines avoid it?", "Each DFA state is a set of NFA states, and there can be 2ⁿ sets. Engines simulate the NFA directly, or build only the DFA states the text reaches."},
			{"What can a backtracking regex engine do that an automaton cannot?", "Backreferences such as (a+)\\1, which need to remember an unbounded amount of text."},
			{"Why does a recursive-descent parser need a nesting limit?", "Each level of nesting is a function call; an input of nothing but opening brackets would exhaust the stack and crash it."},
			{"What is BB(5), and why can't BB(n) be computed in general?", "47,176,870: the most steps any halting 5-state, 2-symbol Turing machine takes from a blank tape. Computing it in general would solve the halting problem."},
			{"What does it mean to reduce problem A to problem B?", "Translate every instance of A, quickly, into an instance of B with the same answer, so a solver for B solves A, and B is at least as hard as A."},
			{"Why was SAT the first NP-complete problem?", "The Cook–Levin theorem (1971) showed that any problem in NP can be translated into a Boolean formula that is satisfiable exactly when the answer is yes."},
			{"What does unit propagation do in a SAT solver?", "When a clause has only one unassigned literal left and the rest are false, that literal must be true: it is set at once, often triggering more forced values."},
			{"What is the difference between an approximation algorithm and a heuristic?", "An approximation algorithm guarantees its answer is within a proven factor of the optimum; a heuristic usually does well but promises nothing."},
		},
	},

	// -----------------------------------------------------------------
	12: {
		Overview: `Every programming language is a set of design choices about how to
express computation. This stage teaches the big ideas behind all of
them (paradigms, type systems, recursion and closures, generics, error
handling, memory management, execution models and concurrency), so that
you can learn any new language in days instead of months. Each idea is
shown in the languages built around it, and measured: C, Java, OCaml,
Haskell, Rust, Go, Erlang, JavaScript, Python and SQL, all run for these
lessons.`,
		Outcomes: []string{
			"Recognise the paradigm and design choices behind any language",
			"Use types, memory models and concurrency models deliberately",
			"Pick up a new programming language quickly",
			"Write recursive and higher-order code that does not overflow the stack",
			"Choose between exceptions, result types and error codes, and between generics and interfaces",
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
			{"Type inference", "The compiler working out the types you did not write, as OCaml and Haskell do for whole programs."},
			{"Closure", "A function together with the variables it captured from where it was created."},
			{"Tail call", "A call that is the very last thing a function does, which a compiler can turn into a jump that uses no stack."},
			{"Generics", "Code written once for many types, such as a list of T; compiled as one copy per type or as one shared copy."},
			{"Result type", "A return value that is either a success with a value or an error, which the caller must handle (Rust's Result, Haskell's Either)."},
			{"Actor", "An isolated process with its own state that communicates only by sending messages, as in Erlang."},
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
values that can be passed around (Haskell, OCaml, Erlang, Clojure, and
map/filter/reduce in most languages). Declarative languages describe
what you want, not how to get it (SQL, HTML, regex).

Most modern languages mix paradigms. The skill is choosing the right
style for each part of a program.`,
				Diagram: `one question: the sum of the squares of the even numbers to 1,000,000

C        for (i = 1; i <= n; i++) if (i % 2 == 0) total += i * i;
Java     LongStream.rangeClosed(1, n).filter(i -> i % 2 == 0)
                   .map(i -> i * i).sum()
Haskell  sum [x * x | x <- [1 .. n], even x]
Erlang   lists:sum([X * X || X <- lists:seq(1, N), X rem 2 =:= 0])
SQL      SELECT sum(i * i) FROM n WHERE i % 2 = 0`,
				UnderTheHood: `All of the versions above were run for this lesson, and all printed
166667166667000000. The OCaml one, written with lists:

  List.init 1_000_000 (fun i -> i + 1)
  |> List.filter (fun i -> i mod 2 = 0)
  |> List.map (fun i -> i * i)
  |> List.fold_left ( + ) 0

crashed with "Fatal error: exception Stack_overflow". In OCaml 4.14,
List.init and List.filter are tail-recursive but List.map is not: on
500,000 elements it needed 500,000 stack frames, more than the 8 MB stack
holds. Rewritten with Seq, a lazy sequence that produces one element at
a time, the same pipeline printed the right answer.

Laziness goes further in Haskell, where every list is lazy:

  take 5 (filter even (map (^ 2) [1 ..]))   →   [4,16,36,64,100]

map over the infinite list [1 ..] finishes, because only the five
squares asked for are ever computed. A paradigm is a choice of what the
language does for you, and of its costs.`,
				MentalModel: "Paradigms are lenses; use the one that makes the problem simplest.",
				TryIt:       "Sum the squares of the even numbers in a list three ways: a for loop in C, a stream in Java, and a list comprehension in Haskell or Erlang.",
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
					"Implement a shopping cart twice: in Java with classes (Cart, Item, methods that change the cart), and in OCaml or Haskell with only pure functions that return a new cart. Write the same five tests for both and compare which was easier to test.",
					"In the functional version, add_item cart item returns a new cart; nothing is ever modified.",
					"Take a program that processes a CSV of orders with nested loops and mutable counters, and rewrite the processing as a pipeline of pure functions (parse → filter → group → summarise) in a functional language (OCaml, Haskell, Elixir) or with JavaScript's map, filter and reduce, keeping its tests passing.",
					"Each stage takes data and returns new data, with no global variables.",
				),
			},
			{
				Name:    "Type Systems",
				Summary: "Rules that catch mistakes before your code runs.",
				Body: `A type system classifies values and checks that operations make sense.
Static typing checks before the program runs (Go, Rust, Java, OCaml,
Haskell, TypeScript); dynamic typing checks while it runs (Python,
JavaScript, Ruby, Erlang). Strong versus weak typing is about implicit
conversions, such as what "5" + 3 does. Type inference lets the
compiler work out types for you, and generics let one function work
safely over many types (List<T>).

Richer types can rule out whole classes of bugs. Optional types replace
null, which Tony Hoare called his "billion-dollar mistake". Sum types
(enums that carry data) force you to handle every case.`,
				Diagram: `                   checked before running      checked while running
no surprises       Rust, OCaml, Haskell        Python, Ruby
(strong)           Java, Go                    Erlang
converts quietly   C (int ↔ pointer, casts)    JavaScript: "5" + 3 = "53"
(weak)`,
				UnderTheHood: `"5" + 3, run in three dynamic languages:

  JavaScript   "53"      and "5" * 3 is 15: it converts quietly
  Python       TypeError: can only concatenate str (not "int") to str
  Ruby         TypeError: no implicit conversion of Integer into String

OCaml inferred every type in this file without a single annotation
(ocamlc -i prints them):

  let f x = x + 1                  val f : int -> int
  let g x y = if x then y else []  val g : bool -> 'a list -> 'a list
  let compose f g x = f (g x)      val compose : ('a -> 'b) -> ('c -> 'a) -> 'c -> 'b

A traffic light as a sum type, with a case forgotten:

  OCaml   Warning 8 [partial-match]: this pattern-matching is not
          exhaustive. Here is an example of a case that is not matched:
          Yellow
  Rust    error[E0004]: non-exhaustive patterns: Light::Yellow not covered

And null, in C: a lookup that returned NULL for an unknown name, passed
straight to strlen, compiled without a warning and ended in
"Segmentation fault", exit status 139. Rust's Option and Haskell's
Maybe make that impossible: the "no answer" case is a value the
compiler makes you handle.`,
				MentalModel: "Types are tests that the compiler writes and runs for free.",
				TryIt:       "Write let f x = x + 1 and let g x = x ^ \"!\" in OCaml's toplevel (ocaml), read the types it infers, then try f \"a\".",
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
					`Python raises TypeError for "5" + 3 and gives "555" for "5" * 3; JavaScript converts, giving "53" and 15.`,
					"Model a traffic light as a sum type (Red | Yellow | Green) with a next function in OCaml, Rust or Haskell, and an Option-returning lookup function. Show that forgetting a case, or using the lookup's result without checking for None, is caught by the compiler.",
					"OCaml's match warns (make it an error with -w +8); Rust's match refuses to compile; in Haskell compile with -Wall.",
					"Take a small untyped JavaScript project (yours or an open-source one), convert it to TypeScript step by step with strict mode on, and record every real bug or unclear piece of code the types exposed.",
					"Start at the edges: function signatures for inputs and outputs.",
				),
			},
			{
				Name:    "Recursion, Closures & Higher-Order Functions",
				Summary: "Functions as values: passed in, returned, and remembering where they were made.",
				Body: `A higher-order function takes functions as arguments or returns them:
map, filter and fold (reduce) replace most loops over collections. A
closure is a function together with the variables it captured where it
was created; calling make_counter() twice gives two counters, each with
its own hidden count. In C, which has no closures, the same thing takes
a function pointer plus a void * pointer to the captured data, passed
along by hand; qsort and pthread_create work this way.

Functional languages loop by recursion, which raises the question of the
stack. A tail call, a call that is the very last thing a function does,
can reuse the current stack frame, so a tail-recursive loop runs in
constant stack space. OCaml, Haskell, Erlang and Scheme guarantee this;
C compilers often do it at -O2; Java, Python and JavaScript in practice
do not, so deep recursion there overflows. The standard fix is an
accumulator: carry the result along as an argument, so nothing is left
to do after the recursive call.`,
				Diagram: `sum 3 = 3 + sum 2                 sum_acc 0 3 = sum_acc 3 2
      = 3 + (2 + sum 1)                        = sum_acc 5 1
      = 3 + (2 + (1 + sum 0))                  = sum_acc 6 0
      = 6   (every frame waits)                 = 6   (a plain loop)`,
				UnderTheHood: `The same recursive sum, sum n = n + sum (n − 1), pushed until it broke:

  Python      n = 900 fine; n = 1,000: RecursionError (its limit is 1,000)
  JavaScript  the deepest that worked, by bisection: 9,642 (Node.js 22)
  OCaml       n = 100,000 fine; n = 1,000,000: Stack_overflow
  OCaml, with an accumulator, sum_acc (acc + n) (n − 1):
              n = 100,000,000 → 5000000050000000, in constant stack

The tail-recursive version made a hundred million calls in the stack
space of one, because OCaml compiles a tail call into a jump. This is the
same reason List.map overflowed in the first lesson: it is not tail-
recursive in OCaml 4.14 (OCaml 5.1 rewrote it so that it is).

Closures, measured in OCaml:

  let make_counter () = let count = ref 0 in fun () -> incr count; !count
  a() twice, then a() → 3; a fresh b() → 1

Each closure carries its own count: the function's code plus a pointer to
its captured variables, which is exactly what a C programmer builds by
hand with a function pointer and a void *.`,
				MentalModel: "A closure is code plus the variables it remembers; a tail call is a loop in disguise.",
				TryIt:       "In JavaScript or OCaml, write make_counter, create two counters and call them in turn; then write the same in C with a struct holding a count and a function pointer.",
				Analogy: `A recipe card with a note clipped to it, "use the oven at Mum's house":
the recipe (the code) plus what it needs from where it was written (the
captured variables). Tail recursion is passing a running total down a
line of people instead of having each one wait to add their number on
the way back.`,
				Example: `JavaScript is built on closures: every event handler and callback that
remembers a variable from the code around it is one. React's hooks are
closures over a component's state. MapReduce and Spark are higher-order
functions (map, then reduce) run over clusters, and Erlang servers run
for years as tail-recursive loops that never grow the stack.`,
				Exercises: trio(
					"What does this print, and why? let a = make_counter () and b = make_counter () in a (); a (); print (b ())",
					"1: b has its own count, captured when it was created; a's two calls do not touch it.",
					"In OCaml or Haskell, implement map, filter and fold yourself, first with plain recursion and then tail-recursively with an accumulator (reversing at the end). Run both on a list of 10 million elements and explain which survives.",
					"The accumulator version builds the result backwards, so reverse it once at the end: still O(n).",
					"In C, write map, filter and fold over an array of ints that take a function pointer and a void *context, and use the context to filter by a threshold chosen at run time and to count calls. Then write the same in JavaScript with closures, and compare what each language had to make explicit.",
					"The context pointer is the closure's environment: in C you allocate and pass it; in JavaScript the language does it for you.",
				),
			},
			{
				Name:    "Generics & Polymorphism",
				Summary: "One piece of code for many types: copied per type, or shared through a table of functions.",
				Body: `Polymorphism means one interface for many types. There are two ways to
compile it:
- monomorphisation (static dispatch): the compiler makes a separate copy
  of the generic code for every type it is used with, and can optimise
  each copy fully. C++ templates, Rust generics and (mostly) Go generics
  work this way. Fast, but more code and longer builds.
- dynamic dispatch: one shared copy calls the type-specific parts through
  a table of function pointers (a vtable): Java interfaces, C++ virtual
  functions, Rust's dyn Trait, Go interfaces, and C's qsort with its
  comparison function. Smaller, flexible, a little slower per call.

Java implements generics by erasure: List<String> and List<Integer> are
the same class at run time, and the type parameters exist only for the
compiler. Haskell's type classes and Rust's traits group the operations a
type must provide (Ord, Display) and pick the implementation by type.`,
				Diagram: `monomorphised (Rust, C++)            dynamic dispatch (Java, Go, C)
largest::<i32>   ─▶ its own code      largest(xs) ─▶ calls cmp through
largest::<f64>   ─▶ its own code                     a pointer, one copy
largest::<char>  ─▶ its own code                     for every type`,
				UnderTheHood: `A generic Rust function, largest<T: PartialOrd + Copy>, called with i32,
f64 and char, compiled to an object file, and nm listed three
functions named gen::largest: one copy per type.

Java's erasure, from a two-line program:

  new ArrayList<String>().getClass() == new ArrayList<Integer>().getClass()
  → true; the class is java.util.ArrayList, for both

The cost of not specialising: sorting 10 million random ints at -O2,

  C qsort, comparison through a function pointer    1.67 s
  C++ std::sort, comparison inlined by the template  0.88 s

The same algorithm family and the same data. qsort calls its
comparison function 200-odd million times through a pointer, and
cannot inline it; std::sort is compiled anew for int, and its
comparison becomes one instruction.`,
				MentalModel: "Copy the code per type for speed, or share it through a function table for flexibility.",
				TryIt:       "Write a generic largest function in Rust (or a template in C++), call it with two types, and find the two copies with nm on the object file.",
				Analogy: `A cookie cutter for every shape you need (monomorphisation) against one
adjustable cutter that you reconfigure by looking up a chart each time
(dynamic dispatch). The first is faster on the production line; the
second takes up less room in the drawer and works for shapes invented
later.`,
				Example: `Go waited until version 1.18 (2022) to add generics, after a decade of
debate about exactly these trade-offs. Rust programs are known for long
compile times partly because of monomorphisation. Java's erasure kept
old bytecode compatible when generics arrived in 2004, which is why a
Java program cannot ask what T is at run time.`,
				Exercises: trio(
					"After erasure, what can a Java method with a List<T> parameter not do that a Rust function with a Vec<T> can? Give one example.",
					"It cannot create a new T, or check whether the list holds Strings: T is gone at run time. Rust has a separate copy that knows T.",
					"In Rust, write a trait Shape with area(), two structs that implement it, and two functions that sum the areas of a list: one generic (fn total<T: Shape>(v: &[T])) and one with Vec<Box<dyn Shape>>. Explain which can mix circles and squares in one list, and why.",
					"Only the dyn version holds different types together; the generic one needs all elements to be the same T.",
					"Benchmark C's qsort against a sort written for int only (or C++ std::sort) on 10 million numbers, then look at both in objdump or Compiler Explorer and find the indirect call that the specialised version does not have.",
					"In the qsort loop, look for call *%rax or call *%rcx: the comparison function, reached through a pointer.",
				),
			},
			{
				Name:    "Error Handling: Exceptions, Results & Error Codes",
				Summary: "Every language must answer: how does a function say it failed, and who must notice?",
				Body: `Four designs, all in wide use:
- error codes: the function returns a special value (-1, NULL) and puts
  the reason somewhere (C's errno). Cheap, but easy to ignore silently.
- exceptions: the function throws, and the error unwinds the stack to
  the nearest handler (Java, Python, C#, OCaml). The happy path stays
  clean, but any call may throw, and nothing in its signature says so
  (except Java's checked exceptions).
- result types: the function returns either a value or an error, and the
  type system makes the caller deal with it (Rust's Result and its ?
  operator, Haskell's Either, Swift's throws). Explicit and checked.
- multiple returns: Go's value, err := f(), checked with if err != nil.

Erlang adds a fifth philosophy, "let it crash": processes are cheap and
isolated, so a process that hits an unexpected error simply dies, and a
supervisor restarts it in a known good state. Whatever the design, the
rules are the same: never silently ignore an error, add context as it
travels up, and handle it where you can actually do something about it.`,
				Diagram: `C        fd = open(path, O_RDONLY);  if (fd < 0) perror(path);
Java     try { read(path); } catch (IOException e) { … }
Rust     let text = fs::read_to_string(path)?;   // ? returns the Err
Go       data, err := os.ReadFile(path); if err != nil { return err }
Erlang   let the process crash; its supervisor restarts it`,
				UnderTheHood: `The same failure, opening a file that does not exist, in three styles:

  C       open() returned -1, errno 2, "No such file or directory"
  Go      open /no/such/file: no such file or directory
  Rust    total("40", "two") → Err(ParseIntError { kind: InvalidDigit })
          total("40", "2")   → Ok(42)

What does each style cost? Five million calls, half of them failing,
after the JIT had warmed up:

  Java, returning -1 for the failure          2.1 ns per call
  Java, throwing IllegalArgumentException   291.1 ns per call
  OCaml, raising an exception                 4.4 ns per call

A Java exception records the stack trace when it is created, which is
most of its cost; OCaml's exceptions are a jump to the handler, cheap
enough to use for control flow. So in Java, exceptions are for
exceptional cases; in OCaml, raising Not_found from a lookup is
ordinary style.`,
				MentalModel: "Make failure impossible to ignore, and handle it where you can act on it.",
				TryIt:       "In Rust, write a function that parses two numbers from strings with ? and returns Result<i64, ParseIntError>; call it with good and bad input and print both results.",
				Analogy: `A delivery that cannot be made. Error codes: a note through the door
that you may or may not read. Exceptions: the parcel is sent back up the
chain until someone signs for it. Result types: the courier will not
leave until you say what to do with it. Let it crash: the van is
replaced and the route restarts from the depot.`,
				Example: `Ericsson's AXD301 telephone switch, written in Erlang with supervisors,
was reported to reach nine nines of availability. Go's explicit if err !=
nil is the most discussed feature of the language. A famous 2014 Apple
bug ("goto fail") skipped a TLS certificate check because of a single
duplicated line in C error-handling code.`,
				Exercises: trio(
					"In C, what goes wrong in: FILE *f = fopen(path, \"r\"); fgets(buf, sizeof buf, f);? Fix it.",
					"fopen returns NULL when the file cannot be opened, and fgets(…, NULL) crashes: check f first and report errno with perror.",
					"Write the same small config-file parser (key = value lines, with a line number in every error) twice: in Rust with a custom error enum and ?, and in Go with fmt.Errorf(\"line %d: %w\", …). Compare how each version shows where an error came from.",
					"In Rust, implement std::fmt::Display for the error enum; in Go, errors.Is and errors.As look through wrapped errors.",
					"In Erlang (or Elixir), write a worker process that crashes on bad input and a supervisor that restarts it, then send it a mix of good and bad messages and show that the service keeps running.",
					"spawn_link links the two processes; process_flag(trap_exit, true) lets the supervisor receive the crash as a message.",
				),
			},
			{
				Name:    "Memory Management",
				Summary: "Who cleans up the memory your program no longer needs?",
				Body: `Programs allocate memory while they run. With manual management (C,
C++), the programmer frees it, and mistakes cause leaks (memory never
freed), use-after-free and double-free bugs, a major source of security
holes.

Garbage collection (Java, Go, JavaScript, OCaml, Haskell, Erlang)
automatically finds unreachable objects and frees them, at the cost of
some CPU time and occasional pauses. Reference counting (Python, Swift)
frees an object as soon as nothing refers to it, but needs extra help
with cycles. Rust's ownership system checks at compile time that every
value has exactly one owner and that references never outlive it: memory
safety without a garbage collector.`,
				Diagram: `manual (C)        malloc ──────────────────── free   you decide when
refcount (Python) +1 per new reference, −1 per dropped one; at 0, freed
tracing GC (Go)   from the roots, mark what is reachable; sweep the rest
ownership (Rust)  one owner; freed when the owner goes out of scope`,
				UnderTheHood: `Reference counting, watched in CPython with sys.getrefcount:

  a new list x                1 reference
  after y = x                 2
  after del y                 1

Two objects pointing at each other, then deleted, are never freed by
counting alone (each still has a count of 1); with the automatic
collector disabled, gc.collect() then found and freed exactly 2 objects.

Rust refused to compile a use after a move:

  error[E0382]: borrow of moved value: name
    let moved = name;      value moved here
    println!("{} {}", name, moved);   value borrowed here after move

Go's collector, traced with GODEBUG=gctrace=1, on a program that
allocated 20 million 72-byte nodes and kept one in a thousand: 641
collections in about 2.1 seconds, each with stop-the-world pauses of
around 0.02 and 0.07 ms either side of 4 ms of concurrent marking, the
live heap held near 2 MB, against more than 1.4 GB allocated in all.`,
				MentalModel: "Every allocation needs an owner who knows when it is finished.",
				TryIt:       "In Python, use sys.getrefcount and the gc module to watch an object's reference count change, then make a cycle and see who frees it.",
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
					"In C, write a program with a memory leak and a use-after-free, find both with AddressSanitizer or Valgrind, and fix them. Then write the same program in Rust and read the compiler's error for the use-after-free.",
					"gcc -g -fsanitize=address prog.c && ./a.out; in Rust, the borrow checker refuses the program before it runs.",
					"A long-running Python web service grows from 200 MB to 4 GB over a week. Reproduce a similar leak in a small script (for example an ever-growing cache), find it with tracemalloc, and fix it with a bounded cache.",
					"functools.lru_cache(maxsize=…) or an explicit eviction policy.",
				),
			},
			{
				Name:    "Interpreters, Compilers & Runtimes",
				Summary: "Different ways of turning source code into a running program.",
				Body: `An interpreter reads your program and executes it directly (shell
scripts, simple interpreters). A compiler translates it ahead of time
into machine code (C, Go, Rust, OCaml, Haskell): fast to run, slower to
build. Many languages compile to bytecode for a virtual machine (Java's
JVM, Python's CPython VM, Erlang's BEAM, .NET's CLR), which makes the
program portable across CPUs.

Just-in-time (JIT) compilers watch the running program and compile its
hot paths to machine code, getting the best of both worlds (V8 for
JavaScript, the JVM's HotSpot, PyPy). The runtime provides services such
as garbage collection, threads and the standard library, and its start-up
time is part of every run.`,
				Diagram: `C, Go, Rust     source ─▶ [compiler] ─▶ machine code ───────────▶ CPU
Java, C#, JS    source ─▶ [compiler] ─▶ bytecode ─▶ [VM + JIT] ─▶ CPU
CPython, Ruby   source ─▶ [compiler] ─▶ bytecode ─▶ [VM] ───────▶ CPU
shell scripts   source ─────────────────▶ [interpreter] ────────▶ CPU`,
				UnderTheHood: `The same naive fib(35), about 30 million calls, everywhere, timed with
start-up included:

  C, gcc -O2                    45 ms    native code
  Go                            67 ms    native code
  OCaml, ocamlopt               73 ms    native code
  Haskell, ghc -O2              88 ms    native code
  JavaScript, Node.js 22       196 ms    JIT (V8), start-up 43 ms
  Java, HotSpot                304 ms    JIT, start-up 53 ms
  Java, -Xint (no JIT)         923 ms    bytecode interpreter
  Ruby 3.3                   1,323 ms    bytecode VM
  Erlang                     1,365 ms    of which 1,179 ms is starting the VM
  Python 3.11                1,962 ms    bytecode VM
  Guile Scheme, interpreted  5,731 ms

The JIT made the same Java bytecode three times faster than the JVM's own
interpreter. Python's dis module shows what CPython runs instead of
machine code:

  LOAD_FAST n · LOAD_CONST 2 · COMPARE_OP < · POP_JUMP_FORWARD_IF_FALSE
  LOAD_GLOBAL fib · LOAD_FAST n · LOAD_CONST 1 · BINARY_OP - · CALL 1 …

Each of those instructions is a trip through the interpreter's loop;
each line of the C version is one or two machine instructions.`,
				MentalModel: "Interpreted versus compiled describes an implementation, not a language.",
				TryIt:       "Use Python's dis module to see the bytecode of a small function, and javap -c to see the JVM bytecode of the same function in Java.",
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
					`In C, write a tiny stack-based bytecode VM (PUSH, ADD, MUL, PRINT) and a compiler that turns expressions such as "2 * (3 + 4)" into that bytecode.`,
					"Compile the syntax tree in post-order: children first, then the operator.",
					"Benchmark the same CPU-heavy function (for example naive Fibonacci of 35) in at least four languages from different rows of the diagram above, measure each runtime's empty start-up separately, and explain the differences using the words interpreter, bytecode, JIT and compiler.",
					"Run each several times and take the median; subtract start-up before comparing.",
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
				Diagram: `threads      ┌ thread ┐ ┌ thread ┐   share memory, guard it with locks
             └───┬────┘ └───┬────┘
                 └─ memory ─┘
actors       (proc) ──msg──▶ (proc) ──msg──▶ (proc)   nothing shared
event loop   one thread: run a callback, wait for the next ready event`,
				UnderTheHood: `What a concurrent task costs, measured for 100,000 of each:

  Erlang processes    spawned in 322 ms, 2,688 bytes each;
                      every one pinged and answered in 241 ms
  Go goroutines       started in 343 ms, about 2,800 bytes each;
                      every one woken and finished in 53 ms
  POSIX threads       70 µs each to create and join, one at a time,
                      and each reserves an 8 MiB stack (Stage 8)

An Erlang process or a goroutine starts with a stack of a few kilobytes
that grows when needed, and the language's own scheduler switches
between them without entering the kernel. That is why a server can give
each connection its own process or goroutine where it could never give
each one an operating-system thread. (At a million, both still worked:
about 2.7 GB in all, though touching that much fresh memory took this
virtual machine 8 to 21 seconds.)`,
				MentalModel: "Prefer passing messages over sharing state; when you must share, lock carefully.",
				TryIt:       "Download 20 web pages one after another, then concurrently with goroutines or JavaScript's Promise.all, and compare the times.",
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
					"Write a producer–consumer program twice: in Go (one goroutine generates numbers, three workers square them over channels, one collects the results) and in Erlang (the same with processes and messages).",
					"In Go, close the channel to tell the workers to stop; in Erlang, send each worker a stop message.",
					"Build a link checker that crawls a site you own (or a local test site) concurrently, with at most 10 requests in flight at a time, and reports broken links, in Go or with JavaScript's async/await.",
					"A buffered channel of size 10 (Go) or a small pool of workers pulling from a queue (JavaScript) limits concurrency so you do not overload the server.",
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
			{"Why does Haskell's take 5 (map (^2) [1 ..]) finish although [1 ..] is infinite?", "Haskell is lazy: elements are computed only when needed, so only the first five squares are ever produced."},
			{"What does type inference do?", "The compiler works out the types you did not write from how values are used, as OCaml does for f x = x + 1 (int -> int)."},
			{"What is a closure?", "A function together with the variables it captured where it was created, so that two closures from the same code can each keep their own state."},
			{"Why does a tail-recursive function not overflow the stack in OCaml or Erlang?", "The recursive call is the last thing it does, so the compiler turns it into a jump that reuses the current stack frame."},
			{"What is the difference between monomorphisation and dynamic dispatch?", "Monomorphisation compiles a separate copy of generic code for each type; dynamic dispatch shares one copy that calls type-specific functions through a table of pointers."},
			{"What does Java's type erasure mean?", "Generic type parameters exist only at compile time: List<String> and List<Integer> are the same class at run time."},
			{"Why are Java exceptions expensive and OCaml exceptions cheap?", "A Java exception records a stack trace when it is created; an OCaml exception is little more than a jump to the handler."},
			{"What does Rust's ? operator do?", "If the Result is an Err, it returns that error from the current function at once; otherwise it unwraps the Ok value."},
			{"Why can a server give each connection an Erlang process or a goroutine, but not an OS thread?", "They start with a stack of a few kilobytes and are switched by the language's own scheduler; an OS thread reserves megabytes and needs the kernel to switch."},
		},
	},
}
