package main

// Resource link verification. Links rot, so the academy ships a checker:
// run "academy -check-links" on any machine with internet access to
// confirm that every resource URL still resolves.

import (
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

// linkResult is the outcome of checking one URL.
type linkResult struct {
	Stage  int
	Title  string
	URL    string
	Status int
	Err    error
}

// ok reports whether the link resolved to a successful page.
func (l linkResult) ok() bool { return l.Err == nil && l.Status >= 200 && l.Status < 400 }

// allResourceLinks lists every resource that has a URL.
func allResourceLinks() []linkResult {
	var out []linkResult
	for id, g := range curriculum {
		for _, r := range g.Resources {
			if r.URL != "" {
				out = append(out, linkResult{Stage: id, Title: r.Title, URL: r.URL})
			}
		}
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
		if method == http.MethodHead && (resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusForbidden || resp.StatusCode >= 500) {
			continue
		}
		return resp.StatusCode, nil
	}
	return 0, fmt.Errorf("no response")
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
			l.Status, l.Err = checkLink(client, l.URL)
		}(&links[i])
	}
	wg.Wait()

	failed := 0
	for _, l := range links {
		mark := sty.Green("✓")
		detail := sty.Gray(fmt.Sprintf("%d", l.Status))
		if !l.ok() {
			failed++
			mark = sty.Red("✗")
			if l.Err != nil {
				detail = sty.Red(l.Err.Error())
			} else {
				detail = sty.Red(fmt.Sprintf("HTTP %d", l.Status))
			}
		}
		fmt.Fprintf(out, "  %s %s %s %s\n", mark, sty.Gray(fmt.Sprintf("stage %2d", l.Stage)), truncate(l.Title, 48), detail)
		if !l.ok() {
			fmt.Fprintf(out, "      %s\n", sty.Gray(l.URL))
		}
	}
	fmt.Fprintf(out, "\n  %d of %d links OK", len(links)-failed, len(links))
	if failed > 0 {
		fmt.Fprintf(out, "; %d failed (a firewall or proxy can also cause failures)", failed)
	}
	fmt.Fprintln(out)
	return failed
}
