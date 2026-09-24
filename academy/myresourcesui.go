package main

// Screens for the learner's own resources.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// renderMyResource prints one of the learner's resources.
func (a *App) renderMyResource(n int, m MyResource) {
	w := a.cols()
	num := "   "
	if n > 0 {
		num = fmt.Sprintf("%2d.", n)
	}
	head := sty.Cyan(num) + " " + sty.Yellow("★") + " " + sty.Bold(truncate(m.Title, w-14-len(m.Kind))) + sty.Gray(" · "+m.Kind)
	a.printf("    %s\n", head)
	if m.URL != "" {
		a.printf("         %s\n", sty.Under(sty.Cyan(m.URL)))
		if hasPDF(m.URL) {
			a.printf("         %s\n", sty.Magenta("⬇ PDF in the Library [l]"))
		}
	}
	for _, l := range wrap(m.Note, w-10, "         ") {
		a.println(sty.Gray(l))
	}
}

// renderMyResources lists all of the learner's resources by stage.
func (a *App) renderMyResources(list []MyResource) {
	a.println(heading("MY RESOURCES · what you found, beside what the academy gives you", sty.Yellow, a.cols()))
	a.println("")
	for _, l := range wrap("Add the books, courses, videos, papers and sites you discover. They appear in each stage's resource library (marked ★), in search, in the Library when the link is a PDF, and in your notes export. Share them with friends as a JSON file.", a.cols()-4, "  ") {
		a.println(sty.Gray(l))
	}
	if len(list) == 0 {
		a.println("")
		a.con.note("Nothing yet. Press a to add your first resource.")
		return
	}
	last := -1
	for i, m := range list {
		if m.Stage != last {
			label := "General (no stage)"
			if m.Stage > 0 {
				a.mu.Lock()
				if _, s := a.reg.findStage(m.Stage); s != nil {
					label = fmt.Sprintf("Stage %d · %s", m.Stage, s.Title)
				}
				a.mu.Unlock()
			}
			a.printf("\n  %s\n", sty.Bold(sty.Yellow(truncate(label, a.cols()-2))))
			last = m.Stage
		}
		a.renderMyResource(i+1, m)
	}
}

// myResources is the interactive screen.
func (a *App) myResources() error {
	for {
		a.mu.Lock()
		list := a.reg.myResourcesFor(0)
		a.mu.Unlock()
		a.paged(false, func() { a.renderMyResources(list) })
		a.println("")
		s, err := a.con.readLine(promptLabel("a add · number to edit or delete · i import · e export", "q back"))
		if err != nil {
			return err
		}
		s = strings.ToLower(strings.TrimSpace(s))
		switch {
		case s == "" || isCancel(s):
			return nil
		case s == "a":
			err = a.addResourceFlow(0)
		case s == "i":
			err = a.importResourcesFlow()
		case s == "e":
			err = a.exportResourcesFlow(list)
		default:
			n, convErr := strconv.Atoi(s)
			if convErr != nil || n < 1 || n > len(list) {
				a.con.warn("Choose a, i, e, a number from the list, or q.")
				continue
			}
			err = a.editResourceFlow(list[n-1])
		}
		if errors.Is(err, errCancel) {
			a.con.note("Cancelled — nothing changed.")
		} else if err != nil {
			return err
		}
	}
}

// askResource prompts for every field, using cur as the defaults.
func (a *App) askResource(cur MyResource, editing bool) (MyResource, error) {
	keep := func(label, value string) string {
		if editing && value != "" {
			return fmt.Sprintf("%s (Enter keeps %q)", label, truncate(value, 30))
		}
		return label
	}
	title, err := a.con.promptText(keep("Title", cur.Title), maxNameLen, !editing)
	if err != nil {
		return cur, err
	}
	if title != "" {
		cur.Title = title
	}
	for {
		link, err := a.con.promptText(keep("Link, https://… (Enter for none)", cur.URL), 500, false)
		if err != nil {
			return cur, err
		}
		if link == "" {
			break
		}
		probe := cur
		probe.URL = link
		if err := validateResource(MyResource{Title: "x", Kind: "Book", URL: link}); err != nil {
			a.con.warn("%v", err)
			continue
		}
		cur = probe
		break
	}
	options := append([]string{}, resourceKinds...)
	label := "What kind of resource is it?"
	if editing {
		label += fmt.Sprintf(" (now: %s)", cur.Kind)
	}
	k, err := a.con.promptChoice(label, options)
	if err != nil {
		return cur, err
	}
	cur.Kind = resourceKinds[k]

	a.mu.Lock()
	def := cur.Stage
	if !editing && def == 0 {
		if s := a.reg.currentStage(); s != nil {
			def = s.ID
		}
	}
	a.mu.Unlock()
	for {
		s, err := a.con.readLine(promptLabel(fmt.Sprintf("Stage 1–16, or 0 for general (Enter = %d)", def), "q cancel"))
		if err != nil {
			return cur, err
		}
		if isCancel(s) {
			return cur, errCancel
		}
		if s == "" {
			cur.Stage = def
			break
		}
		if n, err := strconv.Atoi(s); err == nil && n >= 0 && n <= 16 {
			cur.Stage = n
			break
		}
		a.con.warn("Enter a number from 0 to 16.")
	}
	note, err := a.con.promptText(keep("Why is it good? A note for future you (Enter to skip)", cur.Note), maxNotesLen, false)
	if err != nil {
		return cur, err
	}
	if note != "" {
		cur.Note = note
	}
	return cur, nil
}

// addResourceFlow adds a resource, defaulting to the given stage.
func (a *App) addResourceFlow(stage int) error {
	a.println(heading("ADD A RESOURCE", sty.Yellow, a.cols()))
	for _, l := range wrap("Good resources explain why, not only how; come from people who know the field (authors, maintainers, universities); and have exercises. Official documentation beats a random blog post.", a.cols()-4, "  ") {
		a.println(sty.Gray(l))
	}
	m, err := a.askResource(MyResource{Stage: stage}, false)
	if err != nil {
		return err
	}
	var id int
	var addErr error
	a.mutate(func(r *Registry) { id, addErr = r.addResource(m, time.Now()) })
	if addErr != nil {
		a.con.warn("Not added: %v", addErr)
		return nil
	}
	a.con.ok("Added %s (#%d). Commit to keep it.", sty.Bold(m.Title), id)
	return nil
}

// editResourceFlow edits or deletes one resource.
func (a *App) editResourceFlow(m MyResource) error {
	a.println("")
	a.renderMyResource(0, m)
	choice, err := a.con.promptChoice("What do you want to do?", []string{"Edit it", "Delete it", "Back"})
	if err != nil || choice == 2 {
		return err
	}
	if choice == 1 {
		ok, err := a.con.confirm(fmt.Sprintf("Delete %q?", m.Title))
		if err != nil || !ok {
			return err
		}
		a.mutate(func(r *Registry) {
			for i := range r.MyResources {
				if r.MyResources[i].ID == m.ID {
					r.MyResources = append(r.MyResources[:i], r.MyResources[i+1:]...)
					break
				}
			}
		})
		a.con.ok("Deleted.")
		return nil
	}
	updated, err := a.askResource(m, true)
	if err != nil {
		return err
	}
	if err := validateResource(updated); err != nil {
		a.con.warn("Not saved: %v", err)
		return nil
	}
	a.mutate(func(r *Registry) {
		for i := range r.MyResources {
			if r.MyResources[i].ID == m.ID {
				r.MyResources[i] = updated
			}
		}
	})
	a.con.ok("Saved.")
	return nil
}

// resourcesFilePath is the default shared-file location.
func (a *App) resourcesFilePath() string {
	return filepath.Join(filepath.Dir(a.path), "academy_resources.json")
}

// exportResourcesFlow writes the learner's resources to a shareable file.
func (a *App) exportResourcesFlow(list []MyResource) error {
	if len(list) == 0 {
		a.con.note("Nothing to export yet.")
		return nil
	}
	data, err := exportResources(list)
	if err != nil {
		return err
	}
	path := a.resourcesFilePath()
	if err := atomicWriteFile(path, data); err != nil {
		a.con.fail("Export failed: %v", err)
		return nil
	}
	a.con.ok("Exported %d resource(s) to %s. Send that file to a friend; they import it with i.", len(list), path)
	return nil
}

// importResourcesFlow reads a shared file and adds its new resources.
func (a *App) importResourcesFlow() error {
	def := a.resourcesFilePath()
	path, err := a.con.promptText(fmt.Sprintf("File to import (Enter = %s)", filepath.Base(def)), 500, false)
	if err != nil {
		return err
	}
	if path == "" {
		path = def
	}
	info, err := os.Stat(path)
	if err != nil {
		a.con.fail("Cannot read %s: %v", path, err)
		return nil
	}
	if info.Size() > 5<<20 {
		a.con.fail("That file is larger than 5 MB; it is not a resource list.")
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		a.con.fail("Cannot read %s: %v", path, err)
		return nil
	}
	var added, dupes int
	var problems []string
	var impErr error
	a.mutate(func(r *Registry) { added, dupes, problems, impErr = r.importResources(data, time.Now()) })
	if impErr != nil {
		a.con.fail("%v", impErr)
		return nil
	}
	a.con.ok("Imported %d new resource(s); %d you already had.", added, dupes)
	for i, p := range problems {
		if i == 5 {
			a.con.warn("…and %d more skipped.", len(problems)-5)
			break
		}
		a.con.warn("Skipped %s", p)
	}
	return nil
}
