# Changelog

All notable changes to the Systems & AI Academy are listed here. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
versions follow [Semantic Versioning](https://semver.org/).

## [1.7.0] - 2026-10-22

### Changed
After a walk through the app as a new learner, at 40, 80 and 160 columns:
- **Menu keys work on pages opened from the main menu.** "All screens & keys" promised "type any key to go there", but `d`, `j`, `g`, `k`, `f` and `/` scrolled the page instead. On those pages (also Start Here, the ledger, the progress report and achievements), a menu key now goes straight to its screen; read them with Space, the arrows and `q`.
- **Leaving a concept card early no longer asks straight away whether you understand it.** You see the key idea and how far you read, then choose: read the rest (from where you stopped), mark it understood and explain it, or come back later.
- **`q` quits at the main menu** (also `quit` and `exit`; `5` still works), as it means back everywhere else.
- **Today's workout skips steps with nothing to do,** such as a review with nothing due or a recall with nothing learned yet, with one line each instead of asking "Ready?" first.
- **Prompts use the screen's width** and break only between choices, never inside one ("t / today's word"); a short hint such as (y/n) stays on the question's line.
- **The stage page shows its long overview once.** Coming back to it after the glossary or a quiz shows the title, mastery and the concept list.
- **Fewer keystrokes:** "Study a concept" asks for a number from the list already on screen, and the two-digit "Back" items are gone from the stage menu and the career center (`q` goes back).
- **Tidy screen is on by default for new learners:** every action starts on a clean screen, so the scrollback stays readable.
- **Polish:**
  - The pager's status line names the page, as man does (`START HERE · lines 1–29 of 266`).
  - Search lists every match through the pager (up to 60) instead of only the best 15, and labels career results "career".
  - The Study Hall stage list uses a compact `concepts · exercises` column, so full stage titles fit at 80 columns.
  - Plainer wording: "in progress" instead of "ACTIVE RESEARCH", hours shown as "8.5 h" instead of "8.50h", real plurals instead of "lab(s)" and "match(es)", and the banner reads "Systems & AI Academy".
  - Development builds show a short `dev-<commit>` instead of a long Go pseudo-version.
- **No more "Enter more · b back · /word find · q stop · or type your choice" line under each page** on Windows and with `ACADEMY_PAGER=inline`. A long lesson now just stops at the end of a page, and Enter shows the next one. The pager's keys still work when typed, and questions after a lesson (a choice, y/n, your explanation) are asked as before. The full-screen pager on Linux and macOS is unchanged.
- **A refreshed logo.** The same chip and prompt, redrawn flat and bolder, with two pins a side, so it stays clear from 16 px to the 1024 px app icon. Each part has its own colour from the app's theme: the chip, the prompt `>` and the cursor `_`. New in `assets/logo/`: the mark alone, side-by-side and stacked versions with the `academy` wordmark (for dark and light backgrounds), one-colour black and white, and a one-page usage guide. The app icon, `academy.ico`, `academy.exe`'s icon and the social preview are regenerated from it; the social preview now says 21 stages.

### Added
- **C from zero, completed: four new Stage 1 lessons.** Each comes with analogies, a diagram, an under-the-hood section checked against real gcc output, exercises, glossary words and quiz questions:
  - **Structs:** fields, `.` and `->`, passing structs, padding and alignment;
  - **Dynamic Memory:** malloc, calloc, realloc and free; ownership; leaks, double frees; the growable array;
  - **Files & Streams:** modes, fgets, fgetc and fread, perror, buffering, fclose;
  - **Headers, the Preprocessor & make:** separate compilation, include guards, static, the linker's symbols, macros, Makefiles.

  Stage 1 now has twelve concepts.
- **Checked exercises: your C, compiled and tested for real.** Stage 1 has twelve, one per concept: `hello`, `types`, `fizzbuzz`, `temperature`, `swap`, `letters`, `scores`, `intvec`, `safe_number`, `count`, `stack` and `fix_average`. A project joins them all: `wordfreq`, the ten most frequent words of any text. You write the program in your own editor, in `Documents/Academy Labs/Stage 01/`. The academy then compiles it with your C compiler (`-std=c17 -Wall -Wextra -Werror`, plus AddressSanitizer and UBSan where available), runs it against the test cases, and explains each failure with the line that differs, the exit code, the sanitizer's report, or the usual cause.
  - **Where:** Study Hall → Checked exercises, or `academy -check ID` from another terminal (`academy -check list` shows them all).
  - **The tests catch what matters:**
    - memory never freed (the leak checker, for `intvec` and the project, where it works);
    - a header without an include guard (the checker includes it twice);
    - zero bytes in a file;
    - a fixed-size buffer meeting a 5,000-letter word.
  - **Build hints:** errors like "redefinition of" and "undefined reference" come with what to fix.
  - **What's next** suggests a check as soon as you mark its concept as understood, and the project once the whole stage is.
  - **Evidence, not a tick.** Each run is appended to `academy_checks.jsonl` with the time, the compiler and the cases passed. A check shows as passed on your own, passed with hints, or how many cases pass. Concept cards show the check for their concept. Hints come one at a time and never give the solution.
  - **Limits:**
    - each run uses a fresh temporary folder and none of your environment variables;
    - each case has a 2-second limit, after which the whole process group is killed;
    - output is capped at 1 MB.
  - **Not a sandbox:** the academy only ever runs your own code. Checked exercises need Linux, macOS or a BSD (on Windows, use WSL).
- **Checked exercises in the systems stages.** Seven more, in C and POSIX only, so they run the same on Linux, macOS and the BSDs:
  - Stage 5: `bits` (bit tricks and the raw fields of a float) and `calc` (a recursive-descent calculator);
  - Stage 6: `run` (fork, exec and wait), `rawcopy` (read and write only) and `counter` (threads and a mutex);
  - Stage 8: `upper_server` (a TCP server tested over a real loopback socket) and `http_request` (parsing an HTTP/1.1 request head).
- **Checked exercises for data structures, storage and secure code.** Eight more:
  - Stage 2: `linkedlist`, `hashmap` (it must grow), `heap_sort` (a million numbers in the time limit) and `bfs` (the shortest path through a maze of up to 1000 × 1000);
  - Stage 7: `atomic_save` (temporary file, fsync, rename, fsync the folder) and `wal_kv` (a key-value store with a write-ahead log that survives a torn record);
  - Stage 17: `safe_join` (canonicalise, then check) and `copy_string` (a bounded copy like strlcpy).

- **Stages 2, 3 and 4 completed, to the depth of Stage 1.** Ten new lessons, and every lesson in the three stages now has an *Under the hood* section built on real measurements (compiler output, timings, the bytes in memory), and exercises in C, the language you know from Stage 1, instead of Python:
  - **Stage 2, Data Structures & Algorithms** (5 → 9 lessons): new *Recursion & Divide and Conquer*, *Searching: Linear & Binary Search*, *Greedy Algorithms* and *Dynamic Programming*;
  - **Stage 3, Discrete Mathematics** (5 → 8): new *Modular Arithmetic & Number Theory*, *Recurrences: The Cost of Recursion* and *Graph Theory: Degrees, Trees & Paths*;
  - **Stage 4, Computer Architecture** (5 → 8): new *Building an ALU: Adders, Subtraction & Flags*, *Finite State Machines & Sequential Logic* and *Instruction Sets & Machine Code*.

  Their glossaries, outcomes and quizzes grow with them (new quiz questions are added at the end, so your review cards keep their history).
- **Stages 5 to 8 completed, to the same depth.** Eleven new lessons, and every lesson in the four stages now has an *Under the hood* section measured on a real Linux machine, with C exercises instead of Python:
  - **Stage 5, Computer Systems** (5 → 8 lessons): new *Linking, Loading & Shared Libraries*, *What the Optimizer Does* and *Profiling & Measuring Performance*; measured with `nm`, `readelf`, `objdump`, gcc at each `-O` level, Cachegrind and Callgrind;
  - **Stage 6, Operating Systems** (5 → 8): new *Memory Allocators: How malloc Works*, *File Descriptors, Files & Pipes* and *Signals*; measured with strace and `/proc` (ASLR, lazy allocation and 262,143 page faults, brk against mmap, the cost of a system call and a context switch, a data race at `-O0` and `-O1`);
  - **Stage 7, Storage Engines** (6 → 8): new *Pages, Records & the Buffer Pool* and *Query Planning & Join Algorithms*; measured with fsync timings and SQLite (its file header, B-tree depth from `dbstat`, commit cost per journal mode, an MVCC snapshot, `EXPLAIN QUERY PLAN` before and after an index, a nested loop against an automatic index);
  - **Stage 8, Networks** (6 → 9): new *IP Addressing, Routing & NAT*, *Partitioning & Consistent Hashing* and *Timeouts, Retries & Idempotency*; packets captured on loopback and decoded byte by byte, the routing table, Nagle against delayed ACKs (43 ms against 0.07 ms), SYN retransmissions, and simulations of Raft elections, consistent hashing and retry storms.

- **Stages 9, 11 and 12 completed, to the same depth.** Nine new lessons, an *Under the hood* section in every lesson of the three stages, and exercises in the languages each idea belongs to:
  - **Stage 9, Software Engineering** (5 → 8 lessons): new *Shell Scripting & Automation*, *Build Systems & Dependencies* and *Debugging Tools: gdb, Core Dumps & git bisect*; measured with strace (bash searching PATH), Git's own object files (a blob ID recomputed with sha1sum), make's incremental and parallel builds, gcov coverage against a million random tests, gdb, git bisect over 1,000 commits, and Linux namespaces with unshare;
  - **Stage 11, Theory of Computation** (5 → 8): new *NFAs, DFAs & How Regex Engines Work*, *Reductions & NP-Completeness* and *Approximation & Heuristics for Hard Problems*; measured with Python's backtracking regex against grep, the subset construction's 2ⁿ states, JSON nesting limits in three languages, the 5-state busy beaver (47,176,870 steps), a Sudoku solved as SAT, and 2-opt against the optimum;
  - **Stage 12, Programming Languages** (5 → 8): new *Recursion, Closures & Higher-Order Functions*, *Generics & Polymorphism* and *Error Handling: Exceptions, Results & Error Codes*. Its examples and exercises now use the languages built around each idea (C, Java, OCaml, Haskell, Rust, Go, Erlang, JavaScript and SQL, with Python where reference counting or dynamic typing is the point), all run for these lessons: one task in six languages, type inference and exhaustiveness errors, recursion limits, monomorphisation, the cost of exceptions, Go's garbage collector, fib(35) in eleven runtimes, and 100,000 Erlang processes and goroutines.

  The academy now has 152 concepts and 456 exercises.
- **Stage 10, Security: three new lessons.** *Reading Research Papers* (the three-pass method), *Reading CVEs, CWEs & Advisories* (CVE IDs, CWE classes, CVSS scoring, and prioritising with the CISA KEV catalogue over the raw score) and *Vulnerability Research & Coordinated Disclosure* (how researchers find an unknown bug by fuzzing, code audit and patch diffing, and report it responsibly within the law). Exercises use the legal practice targets the stage already points to (pwn.college, picoCTF, OWASP Juice Shop) and real tools (fuzzers, sanitizers, the NVD). New glossary terms (CVE, CWE, CVSS, zero-day, fuzzing, coordinated disclosure), two new outcomes and six appended quiz questions. New check `cvss`: a CVSS v3.1 base-score calculator, checked against the official formula on every combination of the eight base metrics. The academy now has 155 concepts and 465 exercises.
- **Checked exercises for Stages 2 to 4.** Eighteen more, each timed on inputs where only the efficient idea passes:
  - Stage 2: `duplicates` (O(n log n), not O(n²)), `hanoi`, `bsearch` (the first match, 100,000 lookups), `intervals` (greedy by end time) and `edit_distance` (a 2,000 × 2,000 table), so every Stage 2 lesson has one;
  - Stage 3: `sat_count`, `bitsets`, `binomial` (exact up to C(67, 33)), `modpow` (exponents up to 10¹⁸), `inversions`, `euler` and `birthday`;
  - Stage 4: `bases`, `nand_gates` (the checker counts your NAND gates), `alu` (zero, negative, carry and overflow flags), `number_fsm`, `cpu` (an emulator) and `rv_decode` (real RISC-V machine code).

- **Checked exercises for Stages 5 to 8.** Eleven more:
  - Stage 5: `linkage` (names kept private with `static`, linked against the checker's own) and `checked_add` (overflow checks that cannot overflow);
  - Stage 6: `arena` (a bump allocator), `pipeline` (`A | B` with pipe, dup2 and fork) and `graceful` (SIGTERM and SIGINT, and a blocking read interrupted);
  - Stage 7: `buffer_pool` (LRU with pins and dirty pages), `bloom` (a Bloom filter's false-positive rate) and `hash_join` (100,000 rows each);
  - Stage 8: `routes` (longest-prefix match), `retry` (backoff with jitter) and `ring` (consistent hashing).

- **Checked exercises for Stages 9, 11 and 12.** Nine more: `git_blob`, `make_order` and `bisect`; `nfa_match`, `turing` and `dpll`; `higher_order`, `refcount` and `stack_vm`.

  That makes 67 in all (adding Stage 10's `cvss`). Checks can now give the program files in subfolders and examine the files it leaves behind.
- **Mastery counts real code.** A concept with a checked exercise reaches "mastered" only once that check has passed (with or without hints). Concepts you had mastered before stay at "retained" until you pass their check. Where checks cannot run (Windows), nothing changes.
- **Concept prerequisites.** Concepts now list the earlier concepts they build on (248 links across 150 concepts in all 21 stages). The list sits at the top of each card, with a tick for each one you have marked as understood.
  - **What's next and Today's workout** use them. When the next concept builds on one you have not marked yet, they send you to the first missing step and say which concept it leads to. They do not stop you reading anything.
  - **Tests check the graph:** every name is a real concept, every prerequisite comes earlier and agrees with the stage prerequisites (so there can be no cycle), and every concept from Stage 1 on lists at least one.
- **Get hired: the career center** (`e`). It turns learning into the three things companies hire on: foundations, finished projects and interview skill.
  - **Five target roles:** Backend, Systems / Infrastructure, SRE / Platform, Security and ML / AI Engineer. For each: the work, who hires, the job titles to search for, the interview process, what the interviews test, the stages that prepare you, the portfolio projects employers value, resume skills and advice.
  - **A readiness score:** foundations, portfolio, interview practice and job-search kit, with the most useful next steps.
  - **Portfolio projects:** start them as labs with one key, with a checklist of what makes a project count.
  - **Coding interview patterns:** an eight-step method and 16 patterns with 32 practice problems written for the academy. Each problem has a three-step hint ladder instead of a model answer, and a rubric to score yourself.
  - **System design:** a seven-step framework, building blocks, numbers to know, and 9 classic questions with what a strong design covers and the usual mistake.
  - **Behavioral interviews:** STAR and 12 common questions, with a story bank you write.
  - **Timed mock interviews** (coding, system design, behavioral), scored and kept in a history.
  - **A resume builder** that writes a bullet for each of your projects.
  - **A job search section:** a seven-item kit, an applications tracker with a response rate, and tips.
  - **An export,** `academy_career.md`.
  - **Woven into the rest of the academy:**
    - Practised interview patterns add two cards each to the Daily Review.
    - Problems solved with hints come back after three days to redo alone.
    - Today's workout can practise an interview problem.
    - What's next suggests interview practice.
    - Search finds patterns, design questions and roles.
    - The home screen, share card and progress report show your goal and readiness.
    - Eight career achievements are added.
- **A full-screen pager, like `man` and `less`.** Long screens fill the terminal with only a status line at the bottom (`lines 24–46 of 129 35%`, `(END)`).
  - Keys act at once, without Enter: Space (next page), Enter or ↓ (one line), b (back a page), g / G (start / end), /word (search, with every match highlighted), n / N (next / previous match), h (every key) and q (leave).
  - The normal screen comes back afterwards, with what you were reading in view.
  - Typing a number or menu key leaves and answers what comes next.
  - Resizing the window redraws the page at once, re-wrapped to the new width.
  - On Windows, or with `ACADEMY_PAGER=inline`, the line pager is used as before.

### Fixed
- **The academy starts at once again.** The checked exercises' big test inputs (a million buffer-pool accesses, 100,000-row tables, a 1000 × 1000 maze) were built every time the program started, which had grown to almost 3 seconds; they are now made the first time a check runs. A test keeps it that way, and another checks that every starter compiles as it is, so your first build fails only for what you have not written yet.
- The link check (`-check-links` and the weekly Links workflow) failed whenever a site's own server was failing, such as Complexity Zoo answering 503 for hours. A server error (5xx) that persists after the retry is now a warning, like a site that does not answer; a 404, a moved page or a fake PDF still fails.
- A full correctness pass over the teaching content. Every coding drill was solved, the reference solutions were checked, and every quiz answer, exercise hint, type-in program and number in the lessons was recomputed. Four wordings were fixed:
  - the dictionary's List example could be read as `append` returning the new list;
  - Commit and Migration still described saving and schema versions from before 1.6.0;
  - the Stage 5 glossary said the CPU can do maths only on registers.

## [1.6.2] - 2026-10-15

1.6.1 was published from the commit just before its changes were merged, so its programs are 1.6.0 with a new version number. 1.6.2 is the release with them.

### Added
- **A better pager for long screens**, closer to `less`:
  - **b** goes back a page, and **<** and **>** jump to the start and the end;
  - **/word** finds a word in the screen you are reading (a concept card, Start Here, the ledger, a dictionary entry…), highlights every match and jumps to it; **/** again goes to the next match, and the search wraps to the first match at the end;
  - the status line shows how far through you are (`31%`, `end`) and, while searching, which match you are on (`“stack” 2/7`);
  - it fits narrow terminals with shorter forms;
  - the shortcuts page (`?`) lists the new keys.

### Fixed
- **Paging works like `man` and `less`.** Pressing Enter at "Enter for more" used to leave that prompt in the middle of the text, splitting a sentence across two pages with the prompt between them. The prompt is now erased once answered, so the next page continues the text seamlessly. Output sent to a file or pipe is unchanged.
- **Windows:** the classic console (outside Windows Terminal) now processes escape sequences, so colours, clearing the screen and erasing the pager's prompt work there too.
- A release can no longer be built from the wrong commit: the release script refuses a version with no section in this changelog.

## [1.6.1] - 2026-10-14

Published from the wrong commit: the same programs as 1.6.0. Its intended changes are in 1.6.2.

## [1.6.0] - 2026-10-08

### Added
- **Download the whole Library at once.** Press **`a`** in the Library to download every document you don't have yet, with a summary at the end; Ctrl+C stops it. Your copies stay in your Documents folder, so the books are there even if a link later breaks.
- **Broken links fall back to the Internet Archive.** When a book's or paper's link is dead or no longer returns a PDF, the Library and `-fetch-library` download the Wayback Machine's saved copy of that link instead, and say so. `-check-links` reports a broken PDF link whose archived copy works, so maintainers know learners are still covered, and the weekly Links workflow checks that the archive fallback works.
- **A deeper dictionary: 324 words, up from 278.** 46 new words for real engineering work, such as page fault, context switch, TLB, copy-on-write, false sharing, two's complement, calling convention, write-ahead log, LSM tree, Bloom filter, replication, consensus, quorum, the CAP theorem, two-phase commit, Lamport clocks, dynamic programming, amortized analysis, softmax, attention, transformers, the KV cache, quantization, the bias-variance tradeoff, precision and recall, defense in depth, nonces, side channels, the TLS handshake, tail latency, SLOs, observability and flaky tests.
- **Go deeper** on 22 core words (pointer, stack, heap, process, thread, dictionary/hash table, cache, race condition, deadlock, encryption, hash function, virtual memory, system call, Big-O, TCP, gradient descent, embedding, injection, buffer overflow, compiler, transaction, model): a few paragraphs on how it works underneath, its trade-offs and its classic failure modes.

### Fixed
- The installers no longer fail when the newest release is published without its files (for example with a malformed tag such as `v.1.6.0`). They install the newest release that has its files and checksums, and say so.
- Stage 15's classic reading, the LeNet paper, pointed at yann.lecun.com, which has refused connections for weeks. It now links to the paper's permanent DOI record, and the reading notes where the authors' free copy lives.
- The link check (`-check-links` and the weekly Links workflow) failed on any momentary network error. It now asks a failing link again a few seconds later, and reports a site that stays unreachable as a warning. It still fails on links that need fixing: a 404, a moved page, or a "PDF" that isn't one.
- **Closing the terminal window no longer loses your session.** On macOS and Linux, closing the window (SIGHUP) quit without saving, which is now the usual way to leave an app opened from Launchpad or the app menu.
- Prompt instructions were silently dropped when the question was long. For example, "How sure are you?" appeared without "1 guessing · 2 fairly sure · 3 certain". They now appear on their own line above the question.
- The Study Hall marked every stage "in progress" (◐), even stages you had not started. It now shows your real state (locked, ready, in progress, understood, mastered), with a legend, as the Roadmap does.

### Changed
- **A simpler home screen.** It shows three numbers (streak, today's minutes, cards due), the stage you are on, one call to action and a short menu of 11 everyday actions instead of 26. Start Here stays in the menu for your first few concepts.
  - `?` now opens **All screens & keys**: the full grouped menu, then every key. Type any key on that page to go there.
  - Every key still works from the home screen.
  - Prefer everything at once? Settings → *Home screen: detailed* brings back the full dashboard and menu.
  - The detailed numbers are always in the Progress report, under *At a glance*.
- **Progress saves automatically** after every action. There is no more "commit", "checkpoint" or "uncommitted changes". One backup is made per session, so the 10 kept backups cover your last 10 sessions instead of the last 10 saves. "Exit without saving" became **Undo this session**, which restores your progress exactly as it was when you opened the academy.
- The home screen has one call to action: today's workout until it is done, then the next step. The registry path is no longer printed at every start (`academy -where` shows it).
- Plainer menu labels: Start a lab project, Log study hours, Update progress, Ledger; "Which stage?" instead of "Stage ID"; study time shown as hours with one decimal; proper plurals ("1 card", "3 cards").
- Long lists no longer make you press Enter page after page before you can choose. At any "Enter for more" pause, type the item's number (or any menu key) and it is used straight away.

## [1.5.0] - 2026-10-07

### Added
- **Think first**: features that make you work things out yourself before reaching for the answer, and a Start Here section explaining why, based on learning research.
  - **Today's workout** (`j`): one guided daily session of warm-up review, blank-page recall, one new concept, a mixed exercise from a different topic, and a short reflection ("Did you ask an AI before trying yourself?"). Steps are saved as you go, and the dashboard shows today's progress.
  - **The hint ladder**: stuck on an exercise? Rung 1 gives the questions an engineer asks, rung 2 says where to look it up (concept card, glossary words in the task, reading pointer, man pages and specs), and only rung 3 shows the hint.
  - **Honest solve log**: after each exercise, say how you got there. Anything solved with the hint or with someone else's help (an AI, a forum answer) comes back three days later, to redo from a blank page.
  - **Blank-page recall**: write everything you remember about a concept, then compare with the card and score yourself. Weak recalls return the next day, strong ones weeks later. It is in the workout and on every understood concept.
  - **Calibration**: say how sure you are before a review answer appears. The Progress report shows whether your confidence can be trusted.
  - A *Thinking for yourself* section in the Progress report, a *Solved myself* line on the share card, and five achievements: First Principles, Blank Page, Made It Mine, In Training and Know What You Know.
- **Settings** (`o`): colour themes (default, high contrast, colour-blind friendly, monochrome), plain symbols instead of emoji, and confidence ratings on or off, alongside Classic Mode and tidy screen.

### Fixed
- The review card's title no longer runs past the edge of narrow terminals when the concept name is long.

## [1.4.0] - 2026-09-29

### Added
- **Share card** (`k` on the menu, or `academy -card`): your progress on one screenshot-ready card, with concepts, exercises, stages, study time, streak, achievements and a bar per track, plus the project's address. Post it with #LearnInPublic.

## [1.3.0] - 2026-09-29

### Added
- **One-command installer** for Windows (`install.ps1`) and for macOS and Linux (`install.sh`). No administrator rights are needed, and the checksum is verified before anything is installed. It puts `academy` on your PATH and adds the app with its icon to the Start Menu, Launchpad or the applications menu. On Windows it also appears in Settings → Apps, where it can be uninstalled. Running the same command again upgrades it.
- `academy -where` shows where the program, your progress, backups and PDFs are.

### Changed
- An installed academy (installer, `/usr/local/bin`, `Program Files` and similar) keeps your progress in your per-user data folder, like one installed by a package manager. A registry that an older version left beside the program keeps being used.
- **Library PDFs now live where you can find them**: `Documents/Academy Library`, with one folder per stage (`Stage 04 - Digital Logic & Computer Architecture/…`) and a `My PDFs` folder for PDFs saved from your own links. Files are named after the document, not a code. The Library screen shows the folder's full path, and `o` opens it.
- PDFs downloaded by earlier versions (the flat `academy_library` folder beside the registry) are moved into the new folders and renamed the first time the Library opens. Nothing is downloaded again.
- `ACADEMY_LIBRARY` chooses another folder; with `ACADEMY_HOME` set, the Library sits inside it.

### Fixed
- PDFs you put in the Library folder yourself, in any sub-folder, are listed as **Yours** and can be opened or deleted from the app.

## [1.2.0] - 2026-09-28

### Added
- **Track D, Defensive Security**: four new stages, 20 concepts and 60 exercises about protecting real systems.
  - Stage 17, Secure Coding & Code Review
  - Stage 18, Identity, Access & Cloud Security
  - Stage 19, Network Defence & Monitoring
  - Stage 20, Incident Response, Forensics & Privacy

  Each stage has a glossary, free resources, lab blueprints, a quiz, a classic corner and a type-in whose output comes from running it. Existing registries gain the new stages automatically.
- A **Links** workflow checks every resource and classic-corner link on curriculum changes and weekly, and downloads the start of every Library PDF to confirm it really is a PDF. Sites that refuse automated checks (HTTP 401, 403, 429) are reported as warnings, not failures.
- Stage 10, Security & Cryptography, doubles from 5 to 10 concepts: memory-safety bugs (overflows and use-after-free, with real compiler output), access control, network attacks and defences, secrets and the software supply chain, and detection and incident response. Each has an analogy, real incidents, a diagram and three exercises built on legal practice sites (pwn.college, PortSwigger, OverTheWire, picoCTF).
- Stage 10 also gains glossary terms, five free resources, two lab blueprints (a home security lab, fuzzing a C parser) and four quiz questions.
- Nine dictionary words: broken access control, man-in-the-middle, firewall, denial of service, phishing, two-factor authentication, secrets management, fuzzing and incident response.

### Fixed
- Every stage's PDF library now has something to download: Stages 4, 10, 11, 12, 17, 18 and 20 were empty, and now have free official PDFs (NIST guides, Turing 1936, the BeyondCorp paper and more). The Stage 14 PDF had moved away, so Think Stats replaces it.
- Narrow terminals (40 to 60 columns): the exercise card title, the Classic Mode struggle-clock message, blueprint milestones, the Exercise gym heading and the resource library header no longer run past the edge of the screen.
- Security audit: links typed into the tech-watch log must be http(s) links, because "Open in your browser" passes them to the system, which on Windows can start programs. Links saved earlier are checked again before opening.
- Security audit: imported resource files may no longer contain hidden control codes or right-to-left override characters, which could overwrite or reorder text on screen. The same rule now filters feeds.
- Backups of one registry are no longer listed, or pruned, as another's when two registries share a folder (such as `academy.json` and `academy-old.json`).
- Stage pages with ten or more concepts keep the concept names aligned.
- Ctrl+C during a Library download now cancels just that download, instead of quitting the app.
- A download stops with a clear message when the server sends nothing for 30 seconds, instead of waiting up to 15 minutes.

### Changed
- CI checks for known vulnerabilities with govulncheck, and Dependabot keeps the workflow actions up to date.
- CI: the random-input "monkey" test runs offline, so a slow website can no longer make it fail.
- Release archives are reproducible: every file carries the commit's timestamp, a fixed owner and a fixed order, so rebuilding a commit gives the same checksums.
- Publishing a release from the website no longer builds it twice; the second workflow run stops when the files are already attached.

## [1.1.0] - 2026-09-25

### Added
- An app icon: a CPU chip with a terminal prompt, in the app's colours. The Windows `academy.exe` now shows it in Explorer and the taskbar; the READMEs and a GitHub social-preview card use it too. `tools/icons.py` regenerates every size from the SVG.
- Install with Homebrew (macOS, Linux) or Scoop (Windows). This repository is the tap and the bucket, and each release updates them automatically.
- Signed build-provenance attestations for every release archive, checkable with `gh attestation verify`.
- `ACADEMY_HOME` sets the folder for the registry, backups and PDFs.

### Changed
- Package-manager installs (Homebrew, Scoop, winget, Nix) keep your progress in a per-user data folder instead of next to the binary, because upgrades replace the install folder. Downloaded copies still keep it next to the binary, as before.

## [1.0.1] - 2026-09-24

### Fixed
- Ctrl+D now also stops the live focus timer, so pressing Ctrl+D (twice at most) always saves and exits.
- The downloads now include the MIT `LICENSE`.

### Changed
- CI: the random-input "monkey" test no longer reports false failures. It clears a half-typed line before ending a session, and it matches Go's real crash output instead of words the lessons teach, such as SIGSEGV.
- CI and release workflows use `actions/checkout@v5` and `actions/setup-go@v6` (Node.js 24).

## [1.0.0] - 2026-09-24

The first public release: a complete, offline study companion for
self-learners, in one executable with no dependencies.

### Curriculum
- 17 stages in four tracks, 93 concepts and 279 exercises, from how computers work to AI infrastructure.
- **Stage 0, How Computers Work**: bits and bytes, the CPU's fetch–decode–execute cycle, memory and storage, from source code to a running process, the operating system, I/O and interrupts, networks, and booting.
- **Stage 1, Programming Fundamentals in C**: the first steps in C, including pointers, safe input, undefined behaviour, and debugging with gdb and sanitizers.
- Every concept has an everyday analogy, a real-life example, the technical details (often with a diagram), a mental model, and three exercises (warm-up, practice, real-world) with hints.
- **🔧 Under the hood** sections for Stages 0 and 1, showing what the machine really does, taken from real compiler output.
- For every stage: a glossary, verified resources, lab blueprints, a quiz and a mastery check. Concepts link to related ideas across stages, with a pointer to where to read next.

### Remembering what you learn
- Daily Review with spaced repetition (an SM-2 style scheduler), interleaved across stages.
- A five-level mastery ladder per concept: new, understood, practised, retained, mastered.
- "In your own words" notes, compared against the key idea during reviews.

### Learning tools
- **What's next**: the best next steps, each one key away.
- **Roadmap**: every stage's state and what it builds on.
- **Programmer's dictionary**: 272 jargon words, each with a plain meaning, an analogy, real code and the usual mistake; a word of the day; and a review deck for words.
- **Search** across concepts, glossary terms, resources, blueprints, classics, the dictionary and your own resources.
- **Classic Mode**: a struggle clock before hints, a lab notebook, classic readings, and magazine-style type-ins whose expected output comes from real runs.
- **Focus timer**, **weekly goals**, a **progress report** (activity calendar, weekly trend, deck health), and 23 **achievements**.

### Resources
- **Library**: download free, legally hosted books and papers (including arXiv papers) as PDFs, then open them in your PDF viewer. Downloads are HTTPS-only, checked to be real PDFs, and written atomically.
- **My resources**: add your own finds, and share them as a JSON file.
- **Tech watch**: a fundamentals-first method for keeping up, live headlines from 19 curated RSS/Atom feeds, a watch log and a personal tech radar.
- **Export** your notes and progress to Markdown.

### Reliability
- One human-readable JSON registry, written atomically (temp file, fsync, rename), with automatic rotating backups and `-backups` / `-restore`.
- Older registries are migrated automatically and never lose progress.
- A key-by-key line editor with Ctrl+L, Backspace, Ctrl+U and Ctrl+W, and every screen adapts to the terminal width (40–200 columns) and redraws when the window is resized.
- Tested with unit tests, the race detector, a pseudo-terminal layout check at many widths, and a random-input "monkey" test.

### Platforms
- Release builds for Linux (amd64, arm64), macOS (Intel, Apple silicon), Windows (amd64, arm64) and FreeBSD (amd64), with SHA256 checksums.
- `-version` and a full `-h` help screen.
