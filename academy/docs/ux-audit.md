# Terminal UI audit (October 2026)

An audit of the academy's terminal interface, followed by the fixes made for it.
- **Base commit:** `3750efe`, which includes the home-screen change from pronexuscodex/learning_space#64.
- **Method:** reading the code, running the real binary in a pseudo-terminal (pyte, at the sizes given), rendering screens through `tools/screenshots.py`, and running the test suite.
- **What a finding needs:** a file, a function, and a way to reproduce it. Anything not reproduced is listed as suspected.
- **Screenshots:** no screenshots came with the request, so the README screenshots (`docs/*.png`, rendered from the real binary) and live captures were used instead.

Severity:
- **High:** blocks or misleads the main daily workflow.
- **Medium:** a learner can misread something, or some readers can't read it.
- **Low:** polish.

## 1. Verified defects

### V1 · High · regression: `t` still treated as a menu key on menu pages
- **Where:** `pagerfull.go:118` (`homeKeys`).
- **Cause:** pronexuscodex/learning_space#64 removed the Tidy screen and its `t` key, but `homeKeys` still contains `t`.
- **Reproduce:** open Start Here (`0`) and press `t`. The page closes, and the menu answers `"t" is not a menu option`.
- **Missing test:** none checks that `homeKeys` matches the menu.

### V2 · High: the recommended next activity is off-screen on a standard terminal
- **Where:** `app.go` `printMenu`, `printDashboard` and `workoutLine`.
- **Cause:** the full home is 42 lines at 80 columns and 41 at 40 columns. The "➜ Today's workout … (press j)" line, or the next-step hint, sits near the top, above a menu of about 25 lines.
- **Reproduce:** open the academy on an 80×24 terminal with the demo data. The visible screen holds only the menu and the prompt, with no dashboard and no next step. Capture: `before-home-80x24.txt`, Appendix A.

### V3 · Medium: essential instructions are drawn in low-contrast grey
- **Where:** `keybar.go` `barLine` (the key descriptions) and `console.go` `promptLabel` (hints such as `(q cancel)`).
- **Cause:** both use `sty.Gray`, which is SGR 90 ("bright black"). How readable that is depends entirely on the terminal's palette. WCAG contrast ratios from published palettes:

| Palette | SGR 90 on background |
|---|---|
| xterm default | 5.24:1 |
| Windows Terminal "Campbell" (Windows default) | 4.31:1 (below the 4.5:1 AA minimum) |
| Dracula | 3.03:1 |
| Solarized Dark (colour 8 is the background colour) | 1.00:1, so invisible |

`sty.Gray` has 277 call sites in all. Decorative grey is acceptable, but grey on *how to navigate and cancel* is not. The High contrast theme fixes this (it maps 90 to 37), but nothing points readers to it.

### V4 · Medium: "mastered" and ★ mean three different things
- **Concept mastery** (`retention.go:250`): evidence-based. It needs the concept studied, every exercise done, reviews spaced at least `masteredDays` apart, and the checks passed.
- **Stage "mastered"** (`goals.go:75–90`, legend in `app.go` `pickStage`): one passed mastery exam, or graduation.
- **★ GRADUATED** in the Ledger (`app.go:157`) and **(N★)** for graduated labs on the dashboard (`app.go:201`).
- **Reproduce with the demo data:**
  - The Study Hall marks Stage 0 ★ "mastered".
  - The Ledger shows it "◐ IN PROGRESS".
  - The dashboard counts it in neither "Stages 0/21" (graduated only) nor its concepts, of which only 4 of 8 are mastered.
- The same word is claiming mastery without concept-level evidence.

### V5 · Medium: dashboard labels that contradict other screens
- **"Texts 1/64":** the Ledger calls the same numbers "Reading".
- **"Last 14d" and the streak:** the 14-day line counts logged study hours only (`stats`, `main.go:383–405`), but the streak also counts reviews and workout steps. The demo shows "▲ 15-day streak" next to a 14-day line with empty days, which reads like a contradiction.
- **"Stages 0/21":** the number counts graduated stages, but the label doesn't say so.

### V6 · Medium: no automatic emoji fallback
- **Where:** `settings.go` `applySettings` and `main.go`. `plainText` and `plainWriter` exist, but only the manual setting enables them.
- **Problem:** the Linux virtual console (`TERM=linux`) has no emoji glyphs, and `runeWidth` counts emoji as two columns. So 🔥 💡 🔧 show as blanks or wrong characters, and the boxes that contain them go out of line.

### V7 · Low: High contrast theme isn't presented as the fix for grey text
- **Where:** `thinking.go` `themes`.
- **Problem:** its description, "brighter text, no faint grey", doesn't tell a reader with grey hints they can't read that this theme is the remedy.

## 2. Checked and found sound (no change needed)

| Area | How it was checked | Result |
|---|---|---|
| Nested styles losing their outer colour after an inner reset | Captured raw output of 18 screens and searched for unstyled text after a reset | 0 cases |
| Colour codes that bypass the theme | `grep` for raw SGR outside `ui.go` | Only reverse video (pager status line, search highlight), which doesn't depend on the theme |
| Colour-only status | Read every `●`, `○`, `✓`, `◐`, `–`, `★` site | Every state also has its own shape or words (`checksui.go` `checkMark`, `workout.go`, `studyhall.go`) |
| Settings persistence and invalid values | `thinking_test.go:162–180` (round trip; an unknown theme becomes the default) | Covered |
| Width from 40 to 120 columns | `tools/layout_check.py 40 60 80 120` | 0 overflowing lines |
| Random input, resizes and editing keys | `tools/monkey.py` | 0 crashes or hangs |
| Progress totals | `conceptProgress`, `exerciseProgress` and `stageMastery` count only names present in the curriculum | Stale names can't push totals past 100% |

## 3. Suspected, by design, or out of scope

- **S1 · URLs split at column 0.** In the pager, a URL longer than the screen continues at column 0. *By design:* indenting the continuation would put spaces inside the URL and break copy and paste.
- **S2 · Pager keys on menu pages.** On pages opened from the menu, `j k d f g` act as menu keys instead of pager keys. The pager's `h` page documents this, but it goes against `less` habits. A future option: menu keys only after `:`.
- **S3 · Two mastery systems.** Concept evidence and the stage exam deserve one model. Only the wording is fixed here (V4).
- **S4 · Home screen length.** The full home is longer than 24 rows by the owner's explicit choice (#64), which conflicts with this brief's "move detailed statistics into the progress view". It wasn't changed: V2 makes the next step visible instead, and *Home screen: simple* remains available.
- **S5 · Ctrl+D exits from any prompt.** It's documented, and readline behaves the same way. Not changed.
- **S6 · Flicker.** Not measured: no physical terminal or recorder was available.
- **S7 · Windows.** Line-mode input, the console and VT handling were not tested; no Windows terminal was available. CI runs the unit tests on Windows only.
- **S8 · Palette detection.** Solarized Dark can't be detected reliably (`COLORFGBG` is rarely set). V3 and V7 are mitigated instead.

## 4. Plan, ordered by impact and then by regression risk

| # | Fixes | Change | Risk | Test |
|---|---|---|---|---|
| 1 | V1 | Remove `t` from `homeKeys` | Very low | Every key in `homeKeys` must open something on the menu |
| 2 | V2 | Move the next-step line under the menu, just above the key bar and prompt | Low (one line moves) | The next step is among the last lines of the home at 40, 80 and 110 columns |
| 3 | V3 | Key-bar descriptions and prompt hints use the terminal's normal foreground instead of SGR 90 | Low (colour only) | Neither emits SGR 90 |
| 4 | V4, V5 | Dashboard reads "Graduated N/21", "Labs N (N graduated)", "Reading", "Hours 14d"; the stage legend says "passed" instead of "mastered" | Low (wording) | Rendered dashboard and legend checked |
| 5 | V6 | Emoji fall back to plain symbols automatically on `TERM=linux`; Settings says why | Low | `TERM=linux` turns on plain symbols without changing the saved setting |
| 6 | V7 | The High contrast description names the grey-text problem | Very low | — |
| 7 | — | Tests of the Settings screen: invalid input, cancelling, toggling and saving | None | New test |

Not changed, deliberately: no new UI framework, no key changes beyond V1's leftover, and no removed features.

## Appendix A: before

80×24, opened with the demo data. The whole screen:

```
  │ [n] What's next                 [m] Roadmap · all 21 stages
  │ [d] Dictionary                  [/] Search · find any topic
  │
  │ PRACTISE & TRACK
  │ [f] Focus timer · Pomodoro      [2] Start a lab project
  │ [3] Log study hours             [4] Update progress · graduate, mark read
  │ [1] Ledger                      [g] Weekly goals
  │ [p] Progress report             [a] Achievements
  │ [k] Share card · post your progress
  │
  │ CAREER
  │ [e] Get hired · roles, portfolio, interviews, job search
  │
  │ RESOURCES
  │ [l] Library · download PDFs     [+] My resources · your finds
  │ [w] Tech watch                  [x] Export notes · to Markdown
  │
  │ SETTINGS & EXIT
  │ [o] Settings                    [?] All screens & keys
  │ [c] Classic Mode off            [q] Exit
  │ [8] Undo this session
  └────────────────────────────────────────────────────────────────────────────
  key+Enter open  ? every key  0 start here  q quit  ↑ last choice  ^L redraw
  academy ›
```

## Appendix B: after

The same 80×24 terminal and demo data. The bottom of the screen:

```
  │ [c] Classic Mode off            [q] Exit
  │ [8] Undo this session
  └────────────────────────────────────────────────────────────────────────────
  ➜ Today's workout ○ ○ ○ ○ ○ (press j)
  key+Enter open  ? every key  0 start here  q quit  ↑ last choice  ^L redraw
  academy ›
```

At 40×24 the last three lines are the same next step, the key bar and the prompt. The README screenshots were retaken with `tools/screenshots.py`. `docs/menu.png`, `settings.png`, `key-bar.png`, `roadmap.png` and `workout.png` show the changes. The previous versions are in git history at `3750efe`.

## 5. What changed

| Finding | Files | Change |
|---|---|---|
| V1 | `pagerfull.go` | `t` removed from `homeKeys`. Committed separately into #64, the PR that caused it. |
| V2 | `app.go` | New `nextStepLine()`; the full home prints it after the menu, so it's the last line before the key bar and prompt |
| V3 | `keybar.go`, `console.go` | Key-bar descriptions and prompt hints use the normal foreground, with no SGR 90 |
| V4, V5 | `app.go`, `goals.go` | Dashboard reads "Stages graduated", "(N graduated)", "Reading", "Hours, 14 days"; `stageStateNames` says "passed" |
| V6 | `settings.go`, `main.go` | `noEmojiTerminal()` and `Settings.plainSymbolsOn()`. On `TERM=linux`, plain symbols switch on and Settings says why; the saved choice is unchanged |
| V7 | `thinking.go` | High contrast description: "choose this if grey hints are hard to read" |
| Docs | `README.md`, `CHANGELOG.md`, `tools/screenshots.py`, `docs/*.png` | Updated; the menu shot now ends at the next-step line |

No keys changed apart from removing the leftover `t`. No feature was removed, and there's no new dependency.

## 6. Verification report

Run on Linux with Go 1.24 and gcc on 2026-10-25, after all the changes.

| Item | Command or method | Result |
|---|---|---|
| Formatting | `gofmt -l .` | PASSED (no files listed) |
| Static checks | `go vet ./...` | PASSED |
| Full test suite (166 test functions and their subtests, including every C check, compiled with gcc) | `go test -timeout 30m ./...` | PASSED in 313 s |
| New: menu-page keys match the menu (V1) | `TestHomeKeysAreMenuKeys` | PASSED. Fails on `3750efe` with `homeKeys has 't'` |
| New: next step beside the prompt at 40, 80 and 110 columns, for new and demo learners (V2) | `TestNextStepIsBesideThePrompt` | PASSED. Fails before: the last line was the menu border |
| New: no grey or faint navigation text, in all four themes (V3) | `TestNavigationTextIsNotGrey` | PASSED. 15 failures on the old code |
| New: dashboard words and stage legend (V4, V5) | `TestDashboardWords` | PASSED |
| New: emoji fallback on `TERM=linux` only, saved choice untouched (V6) | `TestPlainSymbolsOnTheLinuxConsole`, plus a live pty run (`TERM=linux` shows "Aa", xterm shows 🔤) | PASSED |
| New: Settings screen with invalid input (`abc`, `99`, empty), toggles, a cancelled theme list, q to leave, and the values after a restart | `TestSettingsScreen` | PASSED |
| Every theme and its status colours | `TestThemesRemapColours` (existing) and `TestNavigationTextIsNotGrey` | PASSED |
| Widths 40, 60, 80, 120 | `python3 tools/layout_check.py 40 60 80 120` | PASSED (0 overflowing lines) |
| Random keys, resizes and editing keys | `python3 tools/monkey.py 30 300 20261025` | PASSED (9,000 keys, 0 failures) |
| Before and after captures | pty capture at 80×24 and 40×24 (Appendices A and B); screenshots retaken | Done |
| C checks with clang | `CC=clang go test -run 'Check\|Reference\|Starter\|Mistake'` | NOT TESTED here. No C code changed; CI runs the suite on macOS (clang). |
| Windows console and line-mode input | — | BLOCKED: no Windows terminal here; CI covers only the unit tests on Windows |
| Flicker and redraw cost on a real terminal | — | NOT TESTED: no physical terminal or recorder |
| Contrast on real palettes | Computed from published palette colours, not measured on screen | Computed only |

Nothing failed in the final run. One test failure during development came from my own test: the long-hint case had no words for it to find. The test was fixed; the code was not at fault.

## 7. Remaining issues and next steps

1. **One mastery model (S3).** A stage could earn its star when its concepts reach mastery (evidence) or its check passes, and show both on every screen. That's a design decision for the owner.
2. **Pager keys on menu pages (S2).** Consider letting `j k d f g` page on menu pages too, with menu keys after `:`. It changes documented behaviour, so it needs the owner's agreement.
3. **Grey elsewhere (V3, remainder).** About 270 decorative grey texts remain. In particular the dashboard labels ("Study", "Labs" and so on) and the item hints in the menu are still SGR 90. Next would be a theme-level choice: a *Default (bright)* palette that maps 90 to the normal foreground and keeps bold for hierarchy.
4. **Full home height (S4).** If the owner later prefers the brief's direction, *Home screen: simple* already shows a short dashboard. Making it the default is a one-line change, reversed in #64 at the owner's request.
5. **Windows (S7).** Run `tools/layout_check.py` equivalents on Windows Terminal, including line mode, and record screenshots.
6. **Terminal recordings.** For future reviews, add an asciinema-style capture to `tools/` for the main flows.
