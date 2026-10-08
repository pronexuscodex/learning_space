package main

// Resource link verification. Links rot, so the academy ships a checker:
// run "academy -check-links" on any machine with internet access to
// confirm that every resource URL still resolves.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

// resourcesVerifiedOn is when the resource URLs were last checked against
// live sources while the curriculum was written.
const resourcesVerifiedOn = "2026-09-24"

// retryDelay is the pause before asking an unreachable link again.
var retryDelay = 5 * time.Second

// linkResult is the outcome of checking one URL.
type linkResult struct {
	Stage  int
	Title  string
	URL    string
	Status int
	Err    error
	PDF    bool // a Library download: the file itself must be a PDF

	Archived bool // the original failed, but the Internet Archive's copy is a PDF
}

// ok reports whether the link resolved to a successful page.
func (l linkResult) ok() bool { return l.Err == nil && l.Status >= 200 && l.Status < 400 }

// guarded reports a server that answered but refuses automated clients
// (401, 403, 418, 429): the page is there, a person with a browser can
// open it, so it is a warning, not a dead link.
func (l linkResult) guarded() bool {
	return l.Err == nil && (l.Status == 401 || l.Status == 403 || l.Status == 418 || l.Status == 429)
}

// unreachable reports a link that got no HTTP answer at all, even after a
// retry: the site is down, slow or blocked from here. Nothing says the
// link is wrong, so it is a warning; a red check must mean something to
// fix (a 404, a moved page, a "PDF" that is not one).
func (l linkResult) unreachable() bool {
	return l.Err != nil && l.Status == 0 && !errors.Is(l.Err, errNotPDF)
}

// allResourceLinks lists every resource and classic-corner reading that
// has a URL.
func allResourceLinks() []linkResult {
	var out []linkResult
	for id, g := range curriculum {
		for _, r := range g.Resources {
			if r.URL != "" {
				out = append(out, linkResult{Stage: id, Title: r.Title, URL: r.URL})
			}
		}
		if cg, ok := classicGuides[id]; ok {
			for _, r := range []ClassicRead{cg.Anchor, cg.Classic, cg.Source} {
				if r.URL != "" {
					out = append(out, linkResult{Stage: id, Title: r.Title, URL: r.URL})
				}
			}
		}
	}
	for _, d := range libraryDocs() {
		out = append(out, linkResult{Stage: d.Stage, Title: "PDF: " + d.Title, URL: d.PDF, PDF: true})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Stage != out[j].Stage {
			return out[i].Stage < out[j].Stage
		}
		return out[i].Title < out[j].Title
	})
	return out
}

// checkLink requests a URL, trying HEAD first and falling back to GET for
// servers that reject HEAD.
func checkLink(client *http.Client, url string) (int, error) {
	for _, method := range []string{http.MethodHead, http.MethodGet} {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("User-Agent", "academy-link-check/1.0")
		resp, err := client.Do(req)
		if err != nil {
			if method == http.MethodHead {
				continue
			}
			return 0, err
		}
		io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		resp.Body.Close()
		if method == http.MethodHead && resp.StatusCode >= 400 { // some servers answer HEAD wrongly; ask again with GET
			continue
		}
		return resp.StatusCode, nil
	}
	return 0, fmt.Errorf("no response")
}

// checkPDF fetches the start of a Library file and checks that it really
// is a PDF, not an HTML error page served with a 200 status.
func checkPDF(client *http.Client, url string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "academy-link-check/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	head := make([]byte, 5)
	if _, err := io.ReadFull(resp.Body, head); err != nil || !bytes.Equal(head, []byte("%PDF-")) {
		// Say what came instead, so the report is something to act on.
		return resp.StatusCode, fmt.Errorf("%w; got %q from %s", errNotPDF, resp.Header.Get("Content-Type"), resp.Request.URL)
	}
	return resp.StatusCode, nil
}

// runLinkCheck checks every resource link concurrently, prints a report,
// and returns the number of failures.
func runLinkCheck(out io.Writer) int {
	links := allResourceLinks()
	client := &http.Client{Timeout: 20 * time.Second}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i := range links {
		wg.Add(1)
		go func(l *linkResult) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if l.PDF {
				l.Status, l.Err = checkPDF(client, l.URL)
			} else {
				l.Status, l.Err = checkLink(client, l.URL)
			}
		}(&links[i])
	}
	wg.Wait()

	// Servers have bad moments (a timeout, a refused connection, an error
	// page for a few seconds): ask again once, a little later, before
	// reporting anything. Only a failure that repeats is reported.
	for i := range links {
		if l := &links[i]; !l.ok() && !l.guarded() {
			time.Sleep(retryDelay)
			if l.PDF {
				l.Status, l.Err = checkPDF(client, l.URL)
			} else {
				l.Status, l.Err = checkLink(client, l.URL)
			}
		}
	}

	// A broken Library PDF: does the Internet Archive's copy work? Learners
	// are then still served, though the link should be fixed.
	for i := range links {
		if l := &links[i]; l.PDF && !l.ok() && !l.guarded() {
			if _, err := checkPDF(client, archiveURL(context.Background(), client, l.URL)); err == nil {
				l.Archived = true
			}
		}
	}

	failed, guarded, unreachable := 0, 0, 0
	for _, l := range links {
		mark := sty.Green("✓")
		detail := sty.Gray(fmt.Sprintf("%d", l.Status))
		if l.guarded() {
			guarded++
			mark = sty.Yellow("⚠")
			detail = sty.Yellow(fmt.Sprintf("HTTP %d: reachable, but refuses automated checks", l.Status))
		} else if l.unreachable() {
			unreachable++
			mark = sty.Yellow("⚠")
			detail = sty.Yellow("no answer, twice (site down, slow or blocked): " + l.Err.Error())
		} else if !l.ok() {
			failed++
			mark = sty.Red("✗")
			if l.Err != nil {
				detail = sty.Red(l.Err.Error())
			} else {
				detail = sty.Red(fmt.Sprintf("HTTP %d", l.Status))
			}
			if l.Archived {
				detail += sty.Yellow(" · learners still get it: the Internet Archive's copy works")
			}
		}
		fmt.Fprintf(out, "  %s %s %s %s\n", mark, sty.Gray(fmt.Sprintf("stage %2d", l.Stage)), truncate(l.Title, 48), detail)
		if !l.ok() {
			fmt.Fprintf(out, "      %s\n", sty.Gray(l.URL))
		}
	}
	fmt.Fprintf(out, "\n  %d of %d links OK", len(links)-failed-guarded-unreachable, len(links))
	if guarded > 0 {
		fmt.Fprintf(out, "; %d refuse automated checks (open them in a browser)", guarded)
	}
	if unreachable > 0 {
		fmt.Fprintf(out, "; %d did not answer (try them again later)", unreachable)
	}
	if failed > 0 {
		fmt.Fprintf(out, "; %d failed (a firewall or proxy can also cause failures)", failed)
	}
	fmt.Fprintln(out)
	return failed
}
