# Changelog

All notable changes to the Systems & AI Academy are listed here. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
versions follow [Semantic Versioning](https://semver.org/).

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
