# Working on the Systems & AI Academy

The app lives in `academy/` (Go, standard library only). Read
`CONTRIBUTING.md` for how to build, test and ship; this file holds the
design rules every change must keep.

## Keyboard first, like the terminals of the 1980s

The academy should make learners fluent at driving a computer from the
keyboard alone, the way people learned on VT100s, vi, Emacs and the BBSs:
fast, without a mouse, with keys that become muscle memory. Every screen
is a chance to practise the keys they will use for the rest of their
careers in shells, editors and pagers.

Rules for every screen and feature:

- **Everything works from the keyboard.** No action may need a mouse,
  and none may need the arrow keys only: offer the letter keys too.
- **Use the classic bindings, the same everywhere.** Readers use the keys
  of `less` and `man` (Space/b, j/k, d/u, g/G, /word, n/N, q, h or ?).
  Text input uses the Emacs and readline keys of the shell (Ctrl-U,
  Ctrl-W, Ctrl-L, Ctrl-D today; Ctrl-A, Ctrl-E, Ctrl-K, ←/→ and ↑/↓
  history are the known gaps to close). One key means one thing across
  the app: `q` goes back or quits, `?` shows the keys, `/` searches.
- **Single keys act at once where it is safe** (menus, the pager);
  anything destructive asks first.
- **Discoverable, then fast.** Every screen shows its main keys, the
  "All screens & keys" page and the pager's `h` list them all, and the
  hints say where the same key works outside the academy (man, less,
  bash, vi).
- **Teach the keys, not only use them.** Where a lesson or exercise
  touches a tool (the shell, an editor, git, a pager), name the keys and
  have the learner use them.
- **Works on any terminal:** 40 columns and up, plain symbols available,
  no colour required to understand anything. `tools/layout_check.py` and
  `tools/monkey.py` must stay clean.

## Content rules

- Concept names never change, and quiz questions are only appended:
  progress and review cards are keyed by them.
- Every lesson's "Under the hood" is built on real measurements, made on
  programs and data written for the purpose, and reported honestly
  (say when a number is a simulation or comes from a datasheet).
- Security stages (10, 17–20) are strictly defensive: protecting systems
  you own, on legal practice targets only.
- Every checked C exercise has a reference solution and mutation tests in
  its `checks_*_test.go`, passing with gcc and clang.
