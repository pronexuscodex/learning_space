package main

// The Internet Archive fallback. Links rot: a university page moves, a
// personal site goes down. When a Library PDF's original link fails, the
// academy asks the Internet Archive's Wayback Machine for its most recent
// saved copy and downloads that instead, with the same checks (HTTPS
// only, a real PDF, a size limit). This keeps the books reachable without
// redistributing them ourselves, which many of their licences forbid.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Wayback Machine endpoints (variables so tests can use a local server).
var (
	waybackAPI = "https://archive.org/wayback/available?url="
	waybackRaw = "https://web.archive.org/web/"
)

var errNoArchive = errors.New("the Internet Archive has no saved copy")

// archivedCopy returns a URL that serves the Wayback Machine's most recent
// saved copy of link, byte for byte (the "id_" form, without the archive's
// page furniture).
func archivedCopy(ctx context.Context, client *http.Client, link string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, waybackAPI+url.QueryEscape(link), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "academy-library/1.0 (self-study; one copy for personal reading)")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w (HTTP %d)", errNoArchive, resp.StatusCode)
	}
	var answer struct {
		Snapshots struct {
			Closest struct {
				Available bool   `json:"available"`
				Status    string `json:"status"`
				Timestamp string `json:"timestamp"`
			} `json:"closest"`
		} `json:"archived_snapshots"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&answer); err != nil {
		return "", fmt.Errorf("%w (unreadable answer)", errNoArchive)
	}
	c := answer.Snapshots.Closest
	if !c.Available || c.Status != "200" || !isDigits(c.Timestamp) {
		return "", errNoArchive
	}
	return waybackRaw + c.Timestamp + "id_/" + link, nil
}

func isDigits(s string) bool {
	return s != "" && strings.Trim(s, "0123456789") == ""
}

// downloadWithFallback downloads link, and when the original fails for a
// reason the archive can help with (gone, moved, not a PDF any more, site
// down), downloads the Internet Archive's copy instead. fromArchive tells
// which one was saved. A cancelled download is not retried.
func downloadWithFallback(ctx context.Context, client *http.Client, link, dest string, limit int64, progress progressFunc) (n int64, fromArchive bool, err error) {
	n, err = downloadPDF(ctx, client, link, dest, limit, progress)
	if err == nil || errors.Is(err, errCancelled) || errors.Is(err, errTooLarge) || errors.Is(err, errNotHTTPS) || ctx.Err() != nil {
		return n, false, err
	}
	alt, aerr := archivedCopy(ctx, client, link)
	if aerr != nil {
		return 0, false, err // report the original problem
	}
	n, aerr = downloadPDF(ctx, client, alt, dest, limit, progress)
	if aerr != nil {
		return 0, false, fmt.Errorf("%v; the Internet Archive's copy failed too: %v", err, aerr)
	}
	return n, true, nil
}
