# Systems & AI Academy: Campus Registry

A terminal-based study companion for **self-learners**. It covers the core of a computer science degree plus modern AI in **16 stages**, and it assumes no prior knowledge. Every concept is explained from scratch and comes with **real-world exercises**.

It is written in Go, uses only the standard library, and keeps all your progress in one readable JSON file.

New here? Run the app and press **`0`** for the *Start Here* guide.

## The curriculum

| Track | Stage | Topic |
|-------|-------|-------|
| **F · Foundations** | 1 | Programming Fundamentals |
| | 2 | Data Structures & Algorithms |
| | 3 | Discrete Mathematics, Logic & Probability |
| | 4 | Digital Logic & Computer Architecture |
| **A · Systems** | 5 | The Iron Layer: compilers & the machine |
| | 6 | Operating Systems Internals & Memory |
| | 7 | Storage Engines, Databases & SQL |
| | 8 | Networks, the Web & Distributed Systems |
| **S · Software, Security & Theory** | 9 | Software Engineering & Professional Tools |
| | 10 | Security & Cryptography |
| | 11 | Theory of Computation & Complexity |
| | 12 | Programming Languages & Paradigms |
| **B · AI** | 13 | Mathematical Foundations: linear algebra & matrix calculus |
| | 14 | Probability, Statistics & Classical ML |
| | 15 | Neural Networks & Autograd from Scratch |
| | 16 | AI Infrastructure, CUDA & Inference |

That is **82 concepts** and **246 exercises**, plus a glossary, a resource library, lab blueprints and a quiz for every stage. Start Here also lists electives for afterwards: graphics, quantum computing, embedded systems and bioinformatics.

## How each concept is taught

Each concept moves from intuition to precision:

1. **💬 In plain words:** an everyday analogy with no jargon.
2. **🌍 Real life:** where you have already met the idea, such as the Ariane 5 overflow, GTA Online's quadratic loading screen, Kubernetes running on Raft, or the CrowdStrike outage.
3. **🔍 The details:** the precise technical explanation, with a diagram where one helps.
4. **◆ Mental model:** one sentence to remember.
5. **🏋 Exercises:** three per concept, from easy to real-world (see below).
6. **📖 Words to know:** any jargon in the card is defined automatically at the bottom.

![Concept card](docs/concept-card.png)

## Exercises

Every concept has a ladder of three exercises, and each exercise has a hint:

| Level | What it is | Example (Hash Tables) |
|-------|------------|------------------------|
| ● **Warm-up** | Pen and paper; no code needed | Hash "cat", "act" and "dog" into 5 buckets by hand; why do two of them collide? |
| ● **Practice** | A small program | Build a word counter on your own hash table and race it against the built-in one on a full book |
| ● **Real-world** | A task from a real situation | Find duplicate sign-ups among 1 million shop accounts in roughly O(n) |

The **Exercise Gym** in each stage's Study Hall shows the exercises, reveals hints only when you ask, and lets you tick exercises off. Your progress is saved (`exercises_done`) and shown in the ledger, and it counts towards overall campus progress.

![Exercise gym](docs/exercise-gym.png)

## Build

```sh
cd academy
go build -trimpath -ldflags="-s -w" -o academy .      # Linux / macOS
go build -trimpath -ldflags="-s -w" -o academy.exe .  # Windows
./academy
```

Requires Go 1.22+. Nothing is downloaded, because there are no dependencies.

On first run, the program creates `academy_campus_registry.json` next to the binary. To put the file somewhere else, pass `-registry path/to/file.json`. (With `go run .`, the file goes in the current directory, because the temporary build directory would be deleted.)

**Upgrading from the 7-stage version:** your existing registry file is migrated automatically. The original stages move to their new numbers (1→5, 2→6, 3→7, 4→8, 5→13, 6→15, 7→16), your labs, hours, readings and concept progress are kept, and the nine new stages are added. Commit once to save the upgraded file.

## Menu

| # | Action |
|---|--------|
| 0 | **Start Here**: what the academy covers, how to practise, where to begin and a study rhythm |
| 1 | **View Campus Ledger**: a colour card per stage, with reading, concept and exercise progress bars, labs and hours |
| 2 | **Enroll in a New Lab**: pick a track (`F`/`A`/`S`/`B`), a stage, then a name, notes and initial hours |
| 3 | **Log Study/Lab Hours**: add hours (`1.5`, `1h30m`, `45m`) with an optional note |
| 4 | **Advance Academic Status**: graduate stages or labs, mark literature as read, set compilation status |
| 5 | **Atomic Commit & Exit** |
| 6 | **Study Hall**: concepts, exercise gym, glossary, resources, lab blueprints and quiz |
| 7 | Checkpoint: commit and keep working |
| 8 | Exit without saving (asks for confirmation) |

Above the menu, a dashboard shows hours, labs, stages, texts, concepts and exercises, an overall progress bar, a 14-day activity sparkline and your study streak.

To cancel a prompt, type `q` at number prompts or `:q` at text prompts. Ctrl-D and Ctrl-C/SIGTERM also commit before exiting.

![Stages](docs/stages.png)

## Study Hall

For every stage:

- **What you'll be able to do:** concrete outcomes, shown before you begin.
- **Concepts:** the six-part cards described above. Mark each one as understood once you could explain it to a friend.
- **Exercise gym:** warm-up → practice → real-world, with hints and check-off.
- **Glossary:** 8–12 words explained simply.
- **Resource library:** free courses, books, papers, videos and tools with links, such as CS50, MIT OCW courses, OSTEP, CMU 15-445, PortSwigger Academy, ISL, and Karpathy's *Zero to Hero*.
- **Lab blueprints:** larger projects with milestones (a text adventure, a route planner, a CPU emulator, a password manager, a regex engine, Raft, GPT from scratch…). You can enroll any of them as a lab in one step.
- **Self-check quiz:** flashcards with self-scoring.

Concepts link across stages. For example, logic gates (Stage 4) become the CPU; the memory hierarchy (Stage 5) returns as the roofline model (Stage 16); and paging (Stage 6) returns as PagedAttention.

![Ledger](docs/ledger.png)

## Colours

Colour is turned on automatically when output goes to a terminal. It is turned off by `-no-color`, by setting `NO_COLOR`, by `TERM=dumb`, or when output is piped. Text wraps to `$COLUMNS` (60–110 columns, default 80).

## Tests

```sh
go test ./...
```

The tests check:
- **Curriculum integrity:** every seeded stage has a guide; every concept has an analogy, a real-life example and a warm-up → practice → real-world exercise ladder with hints; no orphaned explainer or exercise sets.
- **The v1 → v2 migration and stage merging.**
- **Mechanics:** text wrapping, hour parsing, the atomic-write round trip and the activity streak.

## Source layout

One package, one build target:

| File | Contents |
|------|----------|
| `main.go` | schema, migration, storage layer (atomic write), entry point |
| `console.go` | sanitised, line-based input over `bufio` |
| `ui.go` | ANSI styling, progress bars, text wrapping, sparklines |
| `app.go` | dashboard, ledger and all ledger actions |
| `studyhall.go` | concept reader, exercise gym, glossary, resources, blueprints, quizzes, Start Here |
| `curriculum.go` | types, plus Tracks A and B (stages 5–8, 13, 15, 16) |
| `curriculum_foundations.go` | Track F (stages 1–4) |
| `curriculum_software.go` | Track S (stages 9–12) |
| `curriculum_data.go` | Stage 14 |
| `explainers.go` | analogies, real-life examples and glossaries for the original stages; Start Here |
| `exercises.go` | exercises for the original stages |

## Data safety

- **Atomic writes:** the program writes a temporary file in the same directory, fsyncs it, renames it over the registry, then fsyncs the directory. A crash leaves either the old file or the new one, never a half-written file.
- **Corrupt-file guard:** if the registry won't parse or fails validation, the program refuses to start rather than overwrite it.
- **Input rigor:** the program removes control characters, trims whitespace, and rejects input that is too long instead of cutting it off. It re-prompts when a number can't be parsed. Hours must be finite, non-negative, and at most 24 per log entry. Duplicate lab names within a stage are rejected, and so are unknown menu options.

## Schema (abridged)

```json
{
  "schema_version": 2,
  "last_commit": "2026-09-24T11:02:38Z",
  "next_lab_id": 2,
  "tracks": [{
    "id": "F", "name": "Foundations of Computing",
    "stages": [{
      "id": 2, "title": "Data Structures & Algorithms",
      "status": "Active Research",
      "required_literature": [{ "title": "The Algorithm Design Manual", "author": "Steven Skiena", "kind": "Book", "read": true }],
      "concepts_studied": ["Hash Tables"],
      "exercises_done": ["Hash Tables #1", "Hash Tables #3"],
      "labs": [{
        "id": 1, "name": "Route Planner",
        "architecture_notes": "Dijkstra and A* on an OpenStreetMap extract",
        "hours_logged": 4.75,
        "compilation_status": "Tests Passing",
        "status": "Mastered/Graduated",
        "enrolled_at": "2026-09-24T11:02:38Z",
        "hour_log": [{ "at": "2026-09-24T11:02:38Z", "hours": 1.5, "note": "initial hours" }]
      }]
    }]
  }]
}
```
