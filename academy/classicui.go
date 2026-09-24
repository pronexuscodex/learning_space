package main

// Classic Mode screens: the mode toggle, each stage's classic corner, the
// magazine-style type-in lab, and the lab notebook.

import (
	"fmt"
	"strings"
	"time"
)

// addNote appends an entry to the lab notebook.
func (a *App) addNote(stage int, ref, kind, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	a.mutate(func(r *Registry) {
		r.Notebook = append(r.Notebook, NotebookEntry{
			At: time.Now().UTC().Truncate(time.Second), Stage: stage, Ref: ref, Kind: kind, Text: text,
		})
	})
}

// classicBadge is shown on the dashboard when Classic Mode is on.
func classicBadge() string { return sty.Bold(sty.Yellow("▣ CLASSIC MODE")) }

// toggleClassicMode explains Classic Mode and switches it on or off.
func (a *App) toggleClassicMode() error {
	a.mu.Lock()
	on := a.reg.ClassicMode
	a.mu.Unlock()

	a.println(heading("CLASSIC MODE · learn like it's 1985", sty.Yellow, a.cols()))
	a.println("")
	rules := `Strong learners of the 1980s and 1990s had few books, no search engine
and no answers on tap, so they read deeply, typed programs in by hand,
predicted before running, and struggled before asking. Classic Mode
brings those habits back on purpose, and keeps the modern help for
when it is truly needed.

- Struggle clock: an exercise's hint stays locked for 10 minutes (warm-up), 30 minutes (practice) or 45 minutes (real-world) after you first open it. If you are truly stuck, write down what you tried and it unlocks early.
- Lab notebook: before an exercise you write a plan or prediction; afterwards you record what actually happened.
- Classic corner: every Study Hall has an anchor book to read cover to cover, a classic text from the field's history, and real source code to read.
- Type-in lab: every stage has a short listing to predict, type in by hand (never paste), run, and compare with the real output.
- Suggested habits: keep a paper notebook next to you, and work in offline blocks with only man pages and saved documentation.`
	for _, l := range wrap(rules, a.cols()-4, "  ") {
		a.println(l)
	}
	a.println("")

	label := "Turn Classic Mode ON?"
	if on {
		label = "Classic Mode is ON. Turn it OFF?"
	}
	yes, err := a.con.confirm(label)
	if err != nil || !yes {
		return err
	}
	a.mutate(func(r *Registry) { r.ClassicMode = !on })
	if on {
		a.con.ok("Classic Mode is off. Hints are available immediately again. Your notebook is kept.")
	} else {
		a.con.ok("Classic Mode is on. Open any Study Hall and try its Classic corner first.")
	}
	return nil
}

// renderRead prints one classic-corner entry.
func (a *App) renderRead(label string, color func(string) string, r ClassicRead) {
	w := a.cols()
	meta := r.Author
	if r.Year > 0 {
		meta += fmt.Sprintf(", %d", r.Year)
	}
	a.println("")
	for i, l := range wrap(r.Title, w-visibleLen(label)-4, "") {
		if i == 0 {
			a.printf("  %s %s\n", color(sty.Bold(label)), sty.Bold(l))
		} else {
			a.printf("  %s %s\n", strings.Repeat(" ", visibleLen(label)), sty.Bold(l))
		}
	}
	for _, l := range wrap(meta, w-4, "    ") {
		a.println(sty.Gray(l))
	}
	if r.URL != "" {
		a.printf("    %s\n", sty.Under(sty.Cyan(r.URL)))
		if hasPDF(r.URL) {
			for _, l := range wrap("⬇ PDF: download it from the Study Hall's PDF library [9] or the Library [l]", a.cols()-6, "    ") {
				a.println(sty.Magenta(l))
			}
		}
	} else {
		a.printf("    %s\n", sty.Gray("no free official copy; try a library, or a used copy"))
	}
	for _, l := range wrap(r.Why, w-8, "    ") {
		a.println(l)
	}
	for i, l := range wrap(r.How, w-16, "") {
		lead := "            "
		if i == 0 {
			lead = sty.Yellow("How to read") + " "
		}
		a.printf("    %s%s\n", lead, sty.Gray(l))
	}
}

// classicCorner shows a stage's anchor book, classic text, source code and
// type-in, then offers the type-in lab and the notebook.
func (a *App) classicCorner(stageID int) error {
	cg, ti, ok := classicFor(stageID)
	if !ok {
		a.con.note("No classic corner for this stage yet.")
		return nil
	}
	for {
		a.mu.Lock()
		_, s := a.reg.findStage(stageID)
		done := s.TypeInDone
		notes := len(a.reg.notebookFor(stageID, ""))
		on := a.reg.ClassicMode
		a.mu.Unlock()

		a.println("")
		head := []string{sty.Bold(sty.Yellow("CLASSIC CORNER")), sty.Yellow("read deeply, type it in, predict first")}
		if on {
			head = append(head, classicBadge())
		}
		for _, l := range flow(head, sty.Gray(" · "), a.cols(), "  ") {
			a.println(l)
		}
		a.renderRead("📕 Anchor book", sty.Yellow, cg.Anchor)
		a.renderRead("📜 Classic text", sty.Magenta, cg.Classic)
		a.renderRead("🔎 Read the source", sty.Cyan, cg.Source)
		status := sty.Gray("not done yet")
		if done {
			status = sty.Green("✓ done")
		}
		a.printf("\n  %s %s  %s\n", sty.Green(sty.Bold("⌨ Type-in")), sty.Bold(ti.File)+sty.Gray(" ("+ti.Lang+")"), status)
		a.println("")

		choice, err := a.con.promptChoice("Classic corner", []string{
			"Type-in lab: predict, type it by hand, run, compare",
			fmt.Sprintf("Lab notebook %s", sty.Gray(fmt.Sprintf("(%d entries for this stage)", notes))),
			"Back",
		})
		if err != nil {
			return err
		}
		switch choice {
		case 0:
			if err := a.typeInLab(stageID, ti); err != nil {
				return err
			}
		case 1:
			a.paged(true, func() { a.showNotebook(stageID) })
		default:
			return nil
		}
	}
}

// renderListing prints code like a magazine listing, with line numbers.
func (a *App) renderListing(code string) {
	lines := strings.Split(code, "\n")
	width := len(fmt.Sprint(len(lines)))
	a.println("  " + sty.Gray("┌"+strings.Repeat("─", a.cols()-4)))
	room := a.cols() - 5 - width // after "  │ " + number + " "
	for i, l := range lines {
		// Long lines wrap with a ↪ marker, like a printed listing, so no
		// code is ever hidden; type the pieces as one line.
		for j, part := range hardBreak(l, room) {
			num := fmt.Sprintf("%*d", width, i+1)
			if j > 0 {
				num = strings.Repeat(" ", width-1) + "↪"
			}
			a.printf("  %s %s %s\n", sty.Gray("│"), sty.Gray(num), sty.Green(part))
		}
	}
	a.println("  " + sty.Gray("└"+strings.Repeat("─", a.cols()-4)))
}

// typeInLab walks through one type-in: predict, type, run, compare.
func (a *App) typeInLab(stageID int, ti TypeIn) error {
	a.println("")
	a.printf("  %s %s\n", sty.Bold(sty.Green("TYPE-IN LAB ·")), sty.Bold(ti.File)+sty.Gray(" ("+ti.Lang+")"))
	a.renderListing(ti.Code)

	a.printf("\n  %s\n", sty.Bold("1. Predict first."))
	for _, l := range wrap(ti.Predict, a.cols()-8, "     ") {
		a.println(l)
	}
	prediction, err := a.con.promptText("Your prediction (Enter to skip)", maxNotesLen, false)
	if err != nil {
		return err
	}
	a.addNote(stageID, "type-in", NotePrediction, prediction)

	a.printf("\n  %s\n", sty.Bold("2. Type it in by hand."))
	for _, l := range wrap(fmt.Sprintf("Save it as %s. Do not copy and paste: typing every character is how you read every character. Typos are part of the lesson.", ti.File), a.cols()-8, "     ") {
		a.println(l)
	}
	a.printf("\n  %s\n", sty.Bold("3. Run it."))
	a.printf("     %s\n", sty.Cyan(ti.Run))

	show, err := a.con.confirm("Ran it? Show the expected output to compare")
	if err != nil {
		return err
	}
	if !show {
		a.con.note("Come back when you have run it.")
		return nil
	}
	a.printf("\n  %s\n", sty.Bold("4. Compare. Expected output:"))
	for _, l := range strings.Split(ti.Expected, "\n") {
		a.printf("     %s\n", sty.Cyan(l))
	}
	for _, l := range wrap(ti.Lesson, a.cols()-8, "     ") {
		a.println(sty.Yellow(l))
	}
	observed, err := a.con.promptText("What happened? Did it match your prediction? Any typos? (Enter to skip)", maxNotesLen, false)
	if err != nil {
		return err
	}
	a.addNote(stageID, "type-in", NoteObserved, observed)

	yes, err := a.con.confirm("Mark this type-in as done?")
	if err != nil || !yes {
		return err
	}
	a.mutate(func(r *Registry) {
		_, s := r.findStage(stageID)
		s.TypeInDone = true
	})
	a.con.ok("Type-in done. For extra credit, change one thing and predict the new output before running it.")
	return nil
}

// showNotebook prints a stage's lab notebook, oldest first.
func (a *App) showNotebook(stageID int) {
	a.mu.Lock()
	entries := a.reg.notebookFor(stageID, "")
	a.mu.Unlock()

	a.println("")
	a.printf("  %s\n", sty.Bold(sty.Yellow("LAB NOTEBOOK")))
	if len(entries) == 0 {
		a.con.note("Empty so far. In Classic Mode, exercises and type-ins ask you for plans, predictions and observations.")
		return
	}
	for _, e := range entries {
		a.printf("  %s %s %s\n", sty.Gray(formatTime(e.At)), sty.Bold(sty.Yellow(strings.ToUpper(e.Kind))), sty.Gray(e.Ref))
		for _, l := range wrap(e.Text, a.cols()-8, "    ") {
			a.println(l)
		}
	}
}
