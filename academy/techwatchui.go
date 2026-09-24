package main

// Screens for tech watch: the method, live headlines, the watch log, the
// radar, and sources.

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// techWatchMethod is the guide to keeping up without drowning.
const techWatchMethod = `Tech watch (the French call it "veille technologique") is the habit of
following what changes in your field on purpose, a little and often, and
turning what you find into knowledge instead of noise. Done well, it is
what keeps a strong programmer strong for decades.

FUNDAMENTALS FIRST: WHAT MAKES YOU DANGEROUS
Trends sit on top of fundamentals, and fundamentals barely change. C is
from 1972, Unix from 1969, SQL from the 1970s, TCP/IP from 1974, and they
still run the world (the Lindy effect: the longer something has lasted,
the longer it is likely to last). Someone who truly knows memory,
pointers, processes, networks, data structures and how a compiler thinks
can learn any new framework in a week, and can tell which trends are real.
That is what "bulletproof like C" means here: not a language that
protects you, but deep knowledge of the layer everything else rests on.
- Rule of thumb: spend about 80% of your time on the academy's stages and 20% on tech watch.
- A trend is worth learning when you can say which fundamental it builds on. FlashAttention is the memory hierarchy (Stage 5); PagedAttention is virtual memory (Stage 6); a new database is B-trees, logs and transactions (Stage 7).

THE WEEKLY LOOP (30 TO 45 MINUTES)
- Skim (10 min): open Latest headlines. Read titles only. Your only question: does this connect to something I am learning or building?
- Triage (5 min): for each interesting item, choose ignore, save, or read now. Save at most three. Everything else can go; the important things come back.
- Read deeply (15 to 20 min): read one item properly, the original source if possible: the paper, the release notes, the maintainer's post, not a summary of a summary.
- Write it down (5 min): in the watch log, answer "why does this matter to me?" in one or two sentences. If you cannot, it probably does not, yet.
- Try one thing (optional): a ten-line experiment beats ten articles. Move the item to "tried" and note what happened.

EVERY MONTH: UPDATE YOUR RADAR
Place what you have read on your personal technology radar, an idea popularised by Thoughtworks:
- Adopt: use it by default.
- Trial: use it on something real but low-risk.
- Assess: worth understanding; explore it.
- Hold: do not start with it now.
Your radar is personal: it reflects your goals, not what is popular.

CHOOSING SOURCES
- Prefer primary sources: specifications, release notes, papers, and maintainers' own blogs.
- Prefer a few high-signal sources to many noisy ones. Ten feeds you actually read beat a hundred you scroll past.
- Community sites (Hacker News, Lobsters) are good for spotting what people discuss. Always go to the original link, and read the comments with care.
- Add your own feeds in Sources, and remove any source that makes you anxious rather than curious.

QUESTIONS THAT CUT THROUGH HYPE
- Which problem does it solve, and for whom? Do I have that problem?
- What is it built on? Can I explain it using fundamentals I know?
- Is there a paper, a benchmark or real production use, or only a launch post?
- Who benefits if I believe it? Who maintains it, and will they in five years?
- What does it cost: complexity, dependencies, lock-in, money?
Remember the hype cycle: the peak of inflated expectations comes before the trough of disillusionment, and real value usually shows up later and quieter.

WHAT TO AVOID
- Doom-scrolling: tech watch has a timer. Use the focus timer [f].
- FOMO: you do not need every new tool. Missing a trend costs little; weak fundamentals cost a lot.
- Collecting without reading: a saved item you never read is not knowledge. Empty your "to read" list or drop items honestly.`

// techWatch is the tech watch hub.
func (a *App) techWatch() error {
	for {
		a.println(heading("TECH WATCH · keep up with the field, without drowning", sty.Green, a.cols()))
		a.mu.Lock()
		toRead := 0
		for _, w := range a.reg.Watch {
			if w.Status == WatchToRead {
				toRead++
			}
		}
		total, placed := len(a.reg.Watch), len(a.reg.radar())
		feeds := len(a.reg.allFeeds())
		a.mu.Unlock()
		choice, err := a.con.promptChoice("Tech watch", []string{
			sty.Bold("How to keep up") + sty.Gray(" · the method, fundamentals first"),
			sty.Bold("Latest headlines") + sty.Gray(fmt.Sprintf(" · live from %d sources", feeds)),
			"Watch log" + sty.Gray(fmt.Sprintf(" · %d saved, %d to read", total, toRead)),
			"My tech radar" + sty.Gray(fmt.Sprintf(" · adopt · trial · assess · hold (%d rings used)", placed)),
			"Sources" + sty.Gray(" · the feeds, and add your own"),
			"Save something by hand" + sty.Gray(" · a talk, a release, an idea"),
			"Back",
		})
		if err != nil {
			return err
		}
		switch choice {
		case 0:
			a.paged(true, func() {
				a.println("")
				for _, l := range wrap(techWatchMethod, a.cols()-4, "  ") {
					if isHeadingLine(strings.TrimSpace(l)) {
						a.println(sty.Bold(sty.Green(l)))
					} else {
						a.println(l)
					}
				}
			})
		case 1:
			err = a.headlinesScreen()
		case 2:
			err = a.watchLog()
		case 3:
			a.paged(true, a.renderRadar)
		case 4:
			err = a.sourcesScreen()
		case 5:
			err = a.saveWatchFlow(FeedItem{})
		default:
			return nil
		}
		if errors.Is(err, errCancel) {
			a.con.note("Cancelled.")
		} else if err != nil {
			return err
		}
	}
}

// pickTopic asks for a topic; allowAll adds "All topics" first.
func (a *App) pickTopic(label string, allowAll bool) (string, error) {
	options := append([]string{}, watchTopics...)
	if allowAll {
		options = append([]string{"All topics"}, options...)
	}
	i, err := a.con.promptChoice(label, options)
	if err != nil {
		return "", err
	}
	if allowAll && i == 0 {
		return "", nil
	}
	return options[i], nil
}

// headlinesScreen fetches feeds and lets the learner triage items.
func (a *App) headlinesScreen() error {
	topic, err := a.pickTopic("Which topic?", true)
	if err != nil {
		return err
	}
	a.mu.Lock()
	var feeds []Feed
	for _, f := range a.reg.allFeeds() {
		if topic == "" || f.Topic == topic {
			feeds = append(feeds, f)
		}
	}
	since := a.reg.WatchVisited
	a.mu.Unlock()
	if len(feeds) == 0 {
		a.con.note("No sources for that topic yet. Add one in Sources.")
		return nil
	}
	a.con.note("Fetching %d feed(s)…", len(feeds))
	results := fetchFeeds(newDownloadClient(), feeds, 15*time.Second)
	items := headlines(results, 8)
	failed := 0
	for _, r := range results {
		if r.Err != nil {
			failed++
		}
	}
	now := time.Now()
	a.mutate(func(r *Registry) { r.WatchVisited = now.UTC().Truncate(time.Second) })
	if len(items) == 0 {
		a.con.fail("No headlines: %d of %d source(s) failed.", failed, len(feeds))
		for i, r := range results {
			if r.Err != nil && i < 4 {
				a.println(sty.Gray("    " + truncate(r.Feed.Title+": "+r.Err.Error(), a.cols()-6)))
			}
		}
		for _, l := range wrap("You may be offline or behind a firewall. The method, watch log, radar and sources all work offline.", a.cols()-6, "    ") {
			a.println(sty.Gray(l))
		}
		return nil
	}
	if len(items) > 60 {
		items = items[:60]
	}
	for {
		a.paged(false, func() { a.renderHeadlines(items, since, failed, len(feeds)) })
		a.println("")
		s, err := a.con.readLine(promptLabel("Number to open · Enter back", "timebox: skim titles, save at most three"))
		if err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" || isCancel(s) {
			return nil
		}
		n, convErr := strconv.Atoi(s)
		if convErr != nil || n < 1 || n > len(items) {
			a.con.warn("Choose 1 to %d, or press Enter.", len(items))
			continue
		}
		if err := a.headlineItem(items[n-1]); err != nil && !errors.Is(err, errCancel) {
			return err
		}
	}
}

// ago formats how long ago a time was.
func ago(t, now time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := now.Sub(t)
	switch {
	case d < time.Hour:
		return "now"
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 60*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	return t.Format("Jan 2006")
}

// renderHeadlines lists headlines with their source and age.
func (a *App) renderHeadlines(items []FeedItem, since time.Time, failed, total int) {
	w := a.cols()
	now := time.Now()
	a.println(heading("LATEST HEADLINES", sty.Green, w))
	status := fmt.Sprintf("%d headlines from %d source(s)", len(items), total-failed)
	if failed > 0 {
		status += fmt.Sprintf(" · %d unreachable", failed)
	}
	a.printf("  %s\n", sty.Gray(status))
	for i, it := range items {
		mark := " "
		if !since.IsZero() && it.Date.After(since) {
			mark = sty.Green("•")
		}
		meta := sty.Gray(truncate(it.Feed, 22) + " · " + ago(it.Date, now))
		num := sty.Cyan(fmt.Sprintf("%3d.", i+1))
		room := w - 8
		a.printf("  %s %s %s\n", num, mark, sty.Bold(truncate(it.Title, room)))
		a.printf("        %s\n", meta)
	}
	if !since.IsZero() {
		a.printf("\n  %s\n", sty.Gray("• new since your last visit"))
	}
}

// headlineItem shows one item and offers the triage actions.
func (a *App) headlineItem(it FeedItem) error {
	w := a.cols()
	a.println("")
	for _, l := range wrap(it.Title, w-4, "  ") {
		a.println(sty.Bold(l))
	}
	a.printf("  %s\n", sty.Gray(it.Feed+" · "+it.Topic+" · "+ago(it.Date, time.Now())))
	if it.Link != "" {
		a.printf("  %s\n", sty.Under(sty.Cyan(it.Link)))
	}
	if it.Summary != "" {
		for _, l := range wrap(it.Summary, w-6, "    ") {
			a.println(sty.Gray(l))
		}
	}
	options := []string{"Save to my watch log (say why it matters)", "Open in your browser", "Back"}
	pdf, isPDF := pdfURL(it.Link)
	if isPDF {
		options = []string{"Save to my watch log (say why it matters)", "Open in your browser", "Download the PDF to the Library", "Back"}
	}
	choice, err := a.con.promptChoice("What now?", options)
	if err != nil {
		return err
	}
	switch options[choice] {
	case "Save to my watch log (say why it matters)":
		return a.saveWatchFlow(it)
	case "Open in your browser":
		if it.Link == "" {
			a.con.warn("This item has no link.")
			return nil
		}
		if err := openInViewer(it.Link); err != nil {
			a.con.warn("Could not open a browser here (%v). The link is above.", err)
		} else {
			a.con.ok("Opened in your browser. The academy keeps running here.")
		}
	case "Download the PDF to the Library":
		doc := LibraryDoc{Stage: NoStage, Kind: "Paper", Title: it.Title, Page: it.Link, PDF: pdf}
		path := filepath.Join(libraryDir(a.path), doc.fileName())
		if err := a.download(pdf, path); err == nil {
			return a.offerOpen(path)
		}
	}
	return nil
}

// saveWatchFlow saves a headline (or a hand-entered item) to the log.
func (a *App) saveWatchFlow(it FeedItem) error {
	item := WatchItem{Title: it.Title, URL: it.Link, Source: it.Feed, Topic: it.Topic}
	if item.Title == "" {
		title, err := a.con.promptText("What is it? (a title)", maxNameLen, true)
		if err != nil {
			return err
		}
		item.Title = title
		link, err := a.con.promptText("Link, https://… (Enter for none)", 500, false)
		if err != nil {
			return err
		}
		item.URL = link
	}
	if item.Topic == "" {
		topic, err := a.pickTopic("Topic", false)
		if err != nil {
			return err
		}
		item.Topic = topic
	}
	for _, l := range wrap("Why does it matter to you? Connect it to something you know: \"This is Stage 6's virtual memory applied to GPU memory.\"", a.cols()-6, "  ") {
		a.println(sty.Gray(l))
	}
	why, err := a.con.promptText("Why it matters", maxNotesLen, true)
	if err != nil {
		return err
	}
	item.Why = why
	var id int
	var saveErr error
	a.mutate(func(r *Registry) { id, saveErr = r.saveWatch(item, time.Now()) })
	if saveErr != nil {
		a.con.warn("Not saved: %v", saveErr)
		return nil
	}
	a.con.ok("Saved as #%d in your watch log (to read). Commit to keep it.", id)
	return nil
}

// watchLog lists saved items and lets the learner update them.
func (a *App) watchLog() error {
	for {
		a.mu.Lock()
		items := append([]WatchItem{}, a.reg.Watch...)
		a.mu.Unlock()
		if len(items) == 0 {
			a.con.note("Your watch log is empty. Save items from Latest headlines, or save something by hand.")
			return nil
		}
		// To read first, then read, tried, dropped; newest first within each.
		rank := map[string]int{WatchToRead: 0, WatchRead: 1, WatchTried: 2, WatchDone: 3}
		for i := range items {
			for j := i + 1; j < len(items); j++ {
				ri, rj := rank[items[i].Status], rank[items[j].Status]
				if rj < ri || (rj == ri && items[j].Added.After(items[i].Added)) {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
		a.paged(false, func() { a.renderWatchLog(items) })
		a.println("")
		s, err := a.con.readLine(promptLabel("Number to update · Enter back", "q back"))
		if err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" || isCancel(s) {
			return nil
		}
		n, convErr := strconv.Atoi(s)
		if convErr != nil || n < 1 || n > len(items) {
			a.con.warn("Choose 1 to %d, or press Enter.", len(items))
			continue
		}
		if err := a.updateWatchItem(items[n-1].ID); err != nil && !errors.Is(err, errCancel) {
			return err
		}
	}
}

func watchStatusStyle(status string) func(string) string {
	switch status {
	case WatchToRead:
		return sty.Yellow
	case WatchRead:
		return sty.Cyan
	case WatchTried:
		return sty.Green
	}
	return sty.Gray
}

// renderWatchLog prints the log.
func (a *App) renderWatchLog(items []WatchItem) {
	w := a.cols()
	a.println(heading("WATCH LOG", sty.Green, w))
	last := ""
	for i, it := range items {
		if it.Status != last {
			a.printf("\n  %s\n", sty.Bold(watchStatusStyle(it.Status)(strings.ToUpper(it.Status))))
			last = it.Status
		}
		ring := ""
		if it.Ring != "" {
			ring = sty.Magenta(" ◎ " + it.Ring)
		}
		a.printf("  %s %s%s\n", sty.Cyan(fmt.Sprintf("%3d.", i+1)), sty.Bold(truncate(it.Title, w-10-visibleLen(ring))), ring)
		a.printf("       %s\n", sty.Gray(truncate(it.Topic+" · "+it.Added.Local().Format("2006-01-02")+" · "+it.Source, w-8)))
		for _, l := range wrap("Why: "+it.Why, w-8, "       ") {
			a.println(l)
		}
		if it.Note != "" {
			for _, l := range wrap("Learned: "+it.Note, w-8, "       ") {
				a.println(sty.Green(l))
			}
		}
	}
}

// updateWatchItem changes an item's status, radar ring or note.
func (a *App) updateWatchItem(id int) error {
	a.mu.Lock()
	it := a.reg.watchByID(id)
	var cur WatchItem
	if it != nil {
		cur = *it
	}
	a.mu.Unlock()
	if it == nil {
		return nil
	}
	a.println("")
	a.printf("  %s\n", sty.Bold(cur.Title))
	if cur.URL != "" {
		a.printf("  %s\n", sty.Under(sty.Cyan(cur.URL)))
	}
	choice, err := a.con.promptChoice("Update", []string{
		"Mark as read", "Mark as tried (I experimented with it)", "Drop it (not for me, or not now)",
		"Place it on my radar", "Write what I learned", "Open the link", "Delete it", "Back",
	})
	if err != nil {
		return err
	}
	set := func(fn func(w *WatchItem)) {
		a.mutate(func(r *Registry) {
			if w := r.watchByID(id); w != nil {
				fn(w)
			}
		})
	}
	switch choice {
	case 0:
		set(func(w *WatchItem) { w.Status = WatchRead })
		a.con.ok("Marked as read. Consider writing one line on what you learned.")
	case 1:
		set(func(w *WatchItem) { w.Status = WatchTried })
		a.con.ok("Marked as tried. That is the step most people skip.")
	case 2:
		set(func(w *WatchItem) { w.Status = WatchDone })
		a.con.ok("Dropped. Letting go is part of a healthy watch.")
	case 3:
		var options []string
		for _, r := range radarRings {
			options = append(options, fmt.Sprintf("%s %s", sty.Bold(strings.ToUpper(r[:1])+r[1:]), sty.Gray("· "+radarMeaning[r])))
		}
		options = append(options, "Remove from the radar")
		i, err := a.con.promptChoice("Which ring?", options)
		if err != nil {
			return err
		}
		ring := ""
		if i < len(radarRings) {
			ring = radarRings[i]
		}
		set(func(w *WatchItem) { w.Ring = ring })
		a.con.ok("Radar updated.")
	case 4:
		note, err := a.con.promptText("What did you learn or try?", maxNotesLen, true)
		if err != nil {
			return err
		}
		set(func(w *WatchItem) { w.Note = note })
		a.con.ok("Noted.")
	case 5:
		if cur.URL == "" {
			a.con.warn("No link saved for this item.")
		} else if err := openInViewer(cur.URL); err != nil {
			a.con.warn("Could not open a browser here (%v).", err)
		}
	case 6:
		ok, err := a.con.confirm("Delete it from the log?")
		if err != nil || !ok {
			return err
		}
		a.mutate(func(r *Registry) {
			for i := range r.Watch {
				if r.Watch[i].ID == id {
					r.Watch = append(r.Watch[:i], r.Watch[i+1:]...)
					break
				}
			}
		})
		a.con.ok("Deleted.")
	}
	return nil
}

// renderRadar prints the personal technology radar.
func (a *App) renderRadar() {
	a.mu.Lock()
	rings := a.reg.radar()
	a.mu.Unlock()
	w := a.cols()
	a.println(heading("MY TECH RADAR", sty.Magenta, w))
	for _, l := range wrap("Place items from your watch log in a ring (open the log, choose an item, then \"Place it on my radar\"). Review it monthly: things move inwards as you gain confidence, and outwards when they disappoint.", w-4, "  ") {
		a.println(sty.Gray(l))
	}
	colors := map[string]func(string) string{"adopt": sty.Green, "trial": sty.Cyan, "assess": sty.Yellow, "hold": sty.Red}
	for _, ring := range radarRings {
		a.println("")
		a.printf("  %s\n", colors[ring](sty.Bold("◎ "+strings.ToUpper(ring))))
		for _, l := range wrap(radarMeaning[ring], w-6, "    ") {
			a.println(sty.Gray(l))
		}
		if len(rings[ring]) == 0 {
			a.printf("    %s\n", sty.Gray("(empty)"))
			continue
		}
		for _, it := range rings[ring] {
			a.printf("    %s %s\n", sty.Gray(padRight(it.Topic, 12)), truncate(it.Title, w-18))
		}
	}
}

// sourcesScreen lists feeds and manages the learner's own.
func (a *App) sourcesScreen() error {
	for {
		a.mu.Lock()
		mine := append([]Feed{}, a.reg.Feeds...)
		a.mu.Unlock()
		a.paged(false, func() {
			w := a.cols()
			a.println(heading("SOURCES", sty.Green, w))
			for _, topic := range watchTopics {
				first := true
				for _, f := range append(append([]Feed{}, curatedFeeds...), mine...) {
					if f.Topic != topic {
						continue
					}
					if first {
						a.printf("\n  %s\n", sty.Bold(sty.Green(topic)))
						first = false
					}
					tag := ""
					for _, m := range mine {
						if m.URL == f.URL {
							tag = sty.Yellow(" ★ yours")
						}
					}
					a.printf("    %s%s\n", sty.Bold(truncate(f.Title, w-14)), tag)
					a.printf("      %s\n", sty.Gray(truncate(f.URL, w-8)))
					if f.Why != "" {
						for _, l := range wrap(f.Why, w-8, "      ") {
							a.println(sty.Gray(l))
						}
					}
				}
			}
		})
		a.println("")
		s, err := a.con.readLine(promptLabel("a add a feed · r remove one of yours · Enter back", "q back"))
		if err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "", "q", ":q":
			return nil
		case "a":
			for _, l := range wrap("Paste the address of an RSS or Atom feed (often /feed, /rss, /atom.xml or /index.xml on a blog). Prefer primary sources: the project's own blog or release notes.", a.cols()-6, "  ") {
				a.println(sty.Gray(l))
			}
			link, err := a.con.promptText("Feed address (https://…)", 500, true)
			if err != nil {
				return err
			}
			title, err := a.con.promptText("A short name for it", maxNameLen, true)
			if err != nil {
				return err
			}
			topic, err := a.pickTopic("Topic", false)
			if err != nil {
				return err
			}
			var addErr error
			a.mutate(func(r *Registry) { addErr = r.addFeed(Feed{Title: title, URL: link, Topic: topic}) })
			if addErr != nil {
				a.con.warn("Not added: %v", addErr)
			} else {
				a.con.ok("Following %s. It appears in Latest headlines under %s.", title, topic)
			}
		case "r":
			if len(mine) == 0 {
				a.con.note("You have not added any feeds. The curated ones cannot be removed, but you can ignore them by topic.")
				continue
			}
			var names []string
			for _, f := range mine {
				names = append(names, f.Title)
			}
			names = append(names, "Back")
			i, err := a.con.promptChoice("Remove which feed?", names)
			if err != nil || i == len(mine) {
				continue
			}
			a.mutate(func(r *Registry) { r.Feeds = append(r.Feeds[:i], r.Feeds[i+1:]...) })
			a.con.ok("Removed %s.", mine[i].Title)
		default:
			a.con.warn("Choose a, r, or press Enter.")
		}
	}
}
