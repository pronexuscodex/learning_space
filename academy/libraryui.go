package main

// The Library screen: download curriculum PDFs into academy_library/ and
// open them in the system viewer, without leaving the academy.

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ownDownloads lists PDFs in the library folder that are not in the
// catalogue (the learner's own downloads), sorted by name.
func ownDownloads(dir string, catalogue []LibraryDoc) []LibraryDoc {
	known := map[string]bool{}
	for _, d := range catalogue {
		known[d.fileName()] = true
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "*.pdf"))
	sort.Strings(matches)
	var out []LibraryDoc
	for _, m := range matches {
		name := filepath.Base(m)
		if known[name] {
			continue
		}
		title := strings.TrimSuffix(name, ".pdf")
		out = append(out, LibraryDoc{Stage: NoStage, Kind: "Yours", Title: title})
	}
	return out
}

// libraryEntries is what the Library screen lists: the catalogue (for one
// stage, or all), then the learner's own downloads.
func (a *App) libraryEntries(stage int) []LibraryDoc {
	all := libraryDocs()
	var docs []LibraryDoc
	for _, d := range all {
		if stage == AllStages || d.Stage == stage {
			docs = append(docs, d)
		}
	}
	// The learner's own PDF links, unless the catalogue already has them.
	have := map[string]bool{}
	for _, d := range docs {
		have[d.PDF] = true
	}
	a.mu.Lock()
	mine := a.reg.myLibraryDocs(stage)
	a.mu.Unlock()
	for _, d := range mine {
		if !have[d.PDF] {
			have[d.PDF] = true
			docs = append(docs, d)
		}
	}
	order := func(stage int) int { // general items go last
		if stage == NoStage {
			return 1 << 10
		}
		return stage
	}
	sort.SliceStable(docs, func(i, j int) bool { return order(docs[i].Stage) < order(docs[j].Stage) })
	if stage == AllStages {
		docs = append(docs, ownDownloads(libraryDir(a.path), append(all, mine...))...)
	}
	return docs
}

// renderLibrary lists the documents with their download state.
func (a *App) renderLibrary(stage int, docs []LibraryDoc) {
	w := a.cols()
	dir := libraryDir(a.path)
	title := "LIBRARY · free books and papers, saved on your computer"
	if stage != AllStages {
		title = fmt.Sprintf("LIBRARY · Stage %d", stage)
	}
	a.println(heading(title, sty.Magenta, w))
	a.println("")
	for _, l := range wrap("Every document here is free and legally hosted by its authors, publisher or archive (arXiv). Downloads go to one folder beside your registry, so they stay with your progress and can be read offline.", w-4, "  ") {
		a.println(sty.Gray(l))
	}
	if len(docs) == 0 {
		a.println("")
		a.con.note("No downloadable PDFs for this stage. Its resource library [4] links to web books and courses.")
		return
	}
	have := 0
	lastStage := -1
	for i, d := range docs {
		if d.Stage != lastStage {
			label := fmt.Sprintf("Stage %d", d.Stage)
			if d.Stage == NoStage {
				label = "General and your own downloads"
			} else {
				a.mu.Lock()
				if _, st := a.reg.findStage(d.Stage); st != nil {
					label += " · " + st.Title
				}
				a.mu.Unlock()
			}
			a.printf("\n  %s\n", sty.Bold(sty.Magenta(truncate(label, w-2))))
			lastStage = d.Stage
		}
		num := sty.Cyan(fmt.Sprintf("%3d.", i+1))
		mark, state := sty.Gray("⬇"), sty.Gray(d.Kind)
		if size, ok := d.downloaded(dir); ok {
			have++
			mark = sty.Green("✓")
			state = sty.Gray(d.Kind+" · ") + sty.Green(humanBytes(size))
		}
		room := w - 9 - visibleLen(state) - 2
		lines := wrap(d.Title, max(12, room), "")
		for j, l := range lines {
			if j == 0 {
				a.printf("  %s %s %s", num, mark, sty.Bold(l))
				if j == len(lines)-1 {
					a.printf("  %s", state)
				}
				a.println("")
				continue
			}
			line := "         " + sty.Bold(l)
			if j == len(lines)-1 {
				line += "  " + state
			}
			a.println(line)
		}
	}
	a.println("")
	for _, l := range flow([]string{
		sty.Green(fmt.Sprintf("%d of %d saved", have, len(docs))),
		sty.Gray("folder: ") + dir,
	}, sty.Gray(" · "), w, "  ") {
		a.println(l)
	}
}

// library is the interactive Library screen. stage 0 lists everything.
func (a *App) library(stage int) error {
	dir := libraryDir(a.path)
	cleanPartials(dir)
	for {
		docs := a.libraryEntries(stage)
		a.paged(false, func() { a.renderLibrary(stage, docs) })
		a.println("")
		s, err := a.con.readLine(promptLabel("Number to download or open · u your own PDF link · o open folder", "q back"))
		if err != nil {
			return err
		}
		s = strings.ToLower(strings.TrimSpace(s))
		switch {
		case s == "" || isCancel(s) || s == "b":
			return nil
		case s == "o":
			a.openPath(dir, true)
		case s == "u":
			if err := a.downloadOwn(); err != nil && !errors.Is(err, errCancel) {
				return err
			}
		default:
			n, convErr := strconv.Atoi(strings.TrimPrefix(s, "#"))
			if convErr != nil || n < 1 || n > len(docs) {
				a.con.warn("Choose a number from 1 to %d, u, o or q.", len(docs))
				continue
			}
			if err := a.libraryItem(docs[n-1]); err != nil && !errors.Is(err, errCancel) {
				return err
			}
		}
	}
}

// libraryItem downloads a document, or offers to open, refresh or delete
// one that is already saved.
func (a *App) libraryItem(d LibraryDoc) error {
	dir := libraryDir(a.path)
	path := filepath.Join(dir, d.fileName())
	if d.Kind == "Yours" {
		path = filepath.Join(dir, d.Title+".pdf")
	}
	a.printf("\n  %s\n", sty.Bold(d.Title))
	if d.Page != "" {
		a.printf("  %s\n", sty.Under(sty.Cyan(d.Page)))
	}
	if _, err := os.Stat(path); err != nil {
		if d.PDF == "" {
			return nil
		}
		if err := a.download(d.PDF, path); err != nil {
			return nil // already reported
		}
		return a.offerOpen(path)
	}

	options := []string{"Open in your PDF viewer", "Show where it is saved", "Delete it from the library", "Back"}
	if d.PDF != "" {
		options = []string{"Open in your PDF viewer", "Show where it is saved", "Download again (get the latest version)", "Delete it from the library", "Back"}
	}
	choice, err := a.con.promptChoice("This document is saved", options)
	if err != nil {
		return err
	}
	switch options[choice] {
	case "Open in your PDF viewer":
		a.openPath(path, false)
	case "Show where it is saved":
		a.con.note("%s", path)
	case "Download again (get the latest version)":
		// The old copy is replaced only once the new one is complete.
		if err := a.download(d.PDF, path); err == nil {
			return a.offerOpen(path)
		}
	case "Delete it from the library":
		ok, err := a.con.confirm("Delete " + filepath.Base(path) + "?")
		if err != nil || !ok {
			return err
		}
		if err := os.Remove(path); err != nil {
			a.con.fail("Could not delete: %v", err)
		} else {
			a.con.ok("Deleted. You can download it again any time.")
		}
	}
	return nil
}

// download fetches link into path with a live progress line.
func (a *App) download(link, path string) error {
	host := link
	if u, err := url.Parse(link); err == nil {
		host = u.Host
	}
	a.con.note("Downloading from %s … (Ctrl+C cancels)", host)
	var progress progressFunc
	if a.con.screen {
		progress = func(done, total int64) {
			w := a.cols()
			line := sty.Magenta("⬇ ") + sty.Bold(humanBytes(done))
			if total > 0 {
				line += sty.Gray(" of "+humanBytes(total)) + " " + bar(int(done/1024), int(total/1024), max(6, min(24, w-40)), sty.Magenta) + " " + pct(int(done/1024), int(total/1024))
			}
			fmt.Fprint(a.con.out, "\r\x1b[K  "+truncateStyled(line, w-3))
		}
	}
	ctx, stop := context.WithCancelCause(context.Background())
	defer stop(nil)
	a.onInterrupt(func() { stop(errCancelled) })
	defer a.onInterrupt(nil)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	start := time.Now()
	n, err := downloadPDF(ctx, newDownloadClient(), link, path, maxDownloadSize, progress)
	if progress != nil {
		fmt.Fprint(a.con.out, "\r\x1b[K")
	}
	if errors.Is(err, errCancelled) {
		a.con.warn("Download cancelled. Nothing was saved.")
		return err
	}
	if err != nil {
		a.con.fail("Download failed: %v", err)
		for _, l := range wrap("If you are offline or behind a strict firewall, try again later, or open the link in a browser. The library keeps nothing half-downloaded.", a.cols()-6, "    ") {
			a.println(sty.Gray(l))
		}
		return err
	}
	a.con.ok("Saved %s (%s in %s) to %s", filepath.Base(path), humanBytes(n), time.Since(start).Round(100*time.Millisecond), filepath.Dir(path))
	return nil
}

// offerOpen asks whether to open a freshly downloaded file.
func (a *App) offerOpen(path string) error {
	yes, err := a.con.confirm("Open it now in your PDF viewer?")
	if err != nil || !yes {
		return err
	}
	a.openPath(path, false)
	return nil
}

// openPath opens a file or folder with the system's default app.
func (a *App) openPath(path string, folder bool) {
	if folder {
		if err := os.MkdirAll(path, 0o755); err != nil {
			a.con.fail("%v", err)
			return
		}
	}
	if err := openInViewer(path); err != nil {
		a.con.warn("Could not open a viewer here (%v). The file is at:", err)
		a.println("    " + path)
		return
	}
	what := "Opened in your PDF viewer."
	if folder {
		what = "Opened the library folder."
	}
	a.con.ok("%s The academy keeps running here.", what)
}

// downloadOwn saves a PDF from a link the learner provides.
func (a *App) downloadOwn() error {
	for _, l := range wrap("Paste an https:// link to a PDF you are allowed to download, such as an arXiv paper (arxiv.org/abs/… works too), a course handout or an open textbook.", a.cols()-6, "  ") {
		a.println(sty.Gray(l))
	}
	link, err := a.con.promptText("PDF link", 500, true)
	if err != nil {
		return err
	}
	if pdf, ok := pdfURL(link); ok {
		link = pdf
	}
	if u, err := url.Parse(link); err != nil || u.Scheme != "https" || u.Host == "" {
		a.con.warn("Only https:// links are downloaded.")
		return nil
	}
	name, err := a.con.promptText("Save it as (a short title)", maxNameLen, true)
	if err != nil {
		return err
	}
	file := slug(name, 60)
	if file == "" {
		a.con.warn("Use letters or digits in the title.")
		return nil
	}
	path := filepath.Join(libraryDir(a.path), file+".pdf")
	if _, err := os.Stat(path); err == nil {
		ok, err := a.con.confirm(filepath.Base(path) + " exists. Replace it?")
		if err != nil || !ok {
			return err
		}
	}
	if err := a.download(link, path); err != nil {
		return nil
	}
	return a.offerOpen(path)
}
