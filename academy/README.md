# Systems & AI Academy — Campus Registry

A terminal-based ledger for tracking a multi-stage curriculum across low-level systems and from-scratch AI.
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

## Menu

| # | Action |
|---|--------|
| 1 | **View Campus Ledger**: both tracks, every stage, reading checklist, labs, and cumulative hours |
| 2 | **Enroll in a New Lab**: pick a track (`A`/`B`), a stage ID, then a name, architecture notes, and initial hours |
| 3 | **Log Study/Lab Hours**: add hours (`1.5`, `1h30m`, `45m`) with an optional note to an existing lab |
| 4 | **Advance Academic Status**: toggle a stage or lab between *Active Research* and *Mastered/Graduated*, mark books or papers as read, or set a lab's compilation status |
| 5 | **Atomic Commit & Exit** |
| 6 | Checkpoint: commit and keep working |
| 7 | Exit without saving (asks for confirmation) |

To cancel a prompt, type `q` at number prompts or `:q` at text prompts.
Ctrl-D (end of input) and Ctrl-C/SIGTERM also commit before exiting.

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
