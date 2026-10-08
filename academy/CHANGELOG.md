# Changelog

All notable changes to the Systems & AI Academy are listed here. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Fixed
- Stage 15's classic reading, the LeNet paper, pointed at yann.lecun.com, which has refused connections for weeks. It now links to the paper's permanent DOI record, and the reading notes where the authors' free copy lives.
- The link check (`-check-links` and the weekly Links workflow) failed on any momentary network error. It now retries a link that gives no answer, and reports a site that stays unreachable as a warning. It still fails on links that need fixing: a 404, a moved page, or a "PDF" that isn't one.
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
