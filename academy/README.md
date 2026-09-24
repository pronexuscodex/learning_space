# Systems & AI Academy — Campus Registry

A terminal-based ledger and study companion for a multi-stage curriculum across low-level systems and from-scratch AI.
It is written in Go, uses only the standard library, and keeps all its data in one flat JSON file.

## Build

```sh
cd academy
go build -trimpath -ldflags="-s -w" -o academy .   # Linux / macOS
go build -trimpath -ldflags="-s -w" -o academy.exe . # Windows
./academy
```

Requires Go 1.22+. Nothing else is downloaded, because there are no dependencies.

On first run, the program creates `academy_campus_registry.json` next to the binary and fills in the curriculum.
To put the file somewhere else, pass `-registry path/to/file.json`.
(With `go run .`, the file goes in the current directory, because the temporary build directory would be deleted.)

![Concept card](docs/concept-card.png)

## Menu

| # | Action |
|---|--------|
| 1 | **View Campus Ledger**: a colour card per stage, with reading and concept progress bars, labs, compilation badges and cumulative hours |
| 2 | **Enroll in a New Lab**: pick a track (`A`/`B`), a stage ID, then a name, architecture notes and initial hours |
| 3 | **Log Study/Lab Hours**: add hours (`1.5`, `1h30m`, `45m`) with an optional note to an existing lab |
| 4 | **Advance Academic Status**: toggle a stage or lab between *Active Research* and *Mastered/Graduated*, mark literature as read, or set compilation status |
| 5 | **Atomic Commit & Exit** |
| 6 | **Study Hall**: concepts, resources, lab blueprints and quizzes (see below) |
| 7 | Checkpoint: commit and keep working |
| 8 | Exit without saving (asks for confirmation) |

Above the menu, a dashboard shows total hours, labs, graduated stages, texts read, concepts studied, an overall campus progress bar, a 14-day activity sparkline and your current study streak.

To cancel a prompt, type `q` at number prompts or `:q` at text prompts. Ctrl-D and Ctrl-C/SIGTERM also commit before exiting.

## Study Hall

Every stage has built-in teaching material, compiled into the binary (`curriculum.go`):

- **Concepts (5 per stage, 35 in total):** each is a plain-language explanation, with an ASCII diagram where one helps, a *mental model* to remember and a *try it* exercise. You can mark concepts as understood; that progress is stored in the registry (`concepts_studied`) and shown in the ledger.
- **Resource library:** free courses, books, papers, articles, videos and tools with links, grouped by type. Examples: OSTEP, MIT 6.1810 and 6.5840, CMU 15-445, Crafting Interpreters, Karpathy's *Zero to Hero*, 3Blue1Brown, and Simon Boehm's CUDA matmul article.
- **Lab blueprints (3–4 per stage):** suggested projects with milestones, such as a custom malloc, an LSM-tree engine, Raft, an autograd engine, GPT from scratch or a CUDA SGEMM ladder. You can enroll any of them as a lab in one step.
- **Self-check quiz:** flashcards. You answer in your head, reveal the answer, then score yourself.

Concepts also link across stages. For example, the memory hierarchy in Stage 1 comes back as the roofline model in Stage 7, and paging in Stage 2 comes back as PagedAttention.

![Ledger](docs/ledger.png)

## Colours

Colour is turned on automatically when output goes to a terminal. It is turned off by `-no-color`, by setting `NO_COLOR`, by `TERM=dumb`, or when output is piped. Text wraps to `$COLUMNS` (60–110 columns, default 80).

## Tests

```sh
go test ./...
```

The tests check the curriculum's integrity, text wrapping, hour parsing, the atomic-write round trip, loading older registry files, and the activity streak.

## Source layout

One package, one build target:

| File | Contents |
|------|----------|
| `main.go` | schema, storage layer (atomic write), entry point |
| `console.go` | sanitised, line-based input over `bufio` |
| `ui.go` | ANSI styling, progress bars, text wrapping, sparklines |
| `app.go` | dashboard, ledger and all ledger actions |
| `studyhall.go` | concept reader, resource library, blueprints, quizzes |
| `curriculum.go` | the Study Hall knowledge base |

## Data safety

- **Atomic writes:** the program writes a temporary file in the same directory, fsyncs it, renames it over the registry, then fsyncs the directory. A crash leaves either the old file or the new one, never a half-written file.
- **Corrupt-file guard:** if the registry won't parse or fails validation, the program refuses to start rather than overwrite it.
- **Input rigor:** the program removes control characters, trims whitespace, and rejects input that is too long instead of cutting it off. It re-prompts when a number can't be parsed. Hours must be finite, non-negative, and at most 24 per log entry. Duplicate lab names within a stage are rejected, and so are unknown menu options.

## Schema (abridged)

```json
{
  "schema_version": 1,
  "last_commit": "2026-09-24T11:02:38Z",
  "next_lab_id": 2,
  "tracks": [{
    "id": "B", "name": "Advanced AI & Hardware Stack",
    "stages": [{
      "id": 6, "title": "Neural Architectures & Autograd from Scratch",
      "status": "Active Research",
      "required_literature": [{ "title": "Deep Learning", "author": "Goodfellow, Bengio & Courville", "kind": "Book", "read": true }],
      "concepts_studied": ["Computational Graphs & Reverse-Mode Autodiff"],
      "labs": [{
        "id": 1, "name": "Custom Autograd Engine",
        "architecture_notes": "Reverse-mode tape",
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
