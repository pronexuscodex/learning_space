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
	waybackCDX = "https://web.archive.org/cdx/search/cdx?output=json&fl=timestamp&filter=statuscode:200&filter=mimetype:application/pdf&limit=-1&url="
	waybackRaw = "https://web.archive.org/web/"
)

var errNoArchive = errors.New("the Internet Archive has no saved copy")

const archiveAgent = "academy-library/1.0 (self-study; one copy for personal reading)"

// archiveURL returns a URL that serves the Wayback Machine's most recent
// saved copy of link, byte for byte (the "id_" form, without the archive's
// page furniture). It asks the availability API first, then the capture
// index (the availability API sometimes answers "nothing" for pages it
// has), and as a last resort returns the form that makes the Wayback
// Machine redirect to its newest capture itself. Whatever it returns is
// still checked by the download: HTTPS, a real PDF, the size limit.
func archiveURL(ctx context.Context, client *http.Client, link string) string {
	if ts, err := archiveAvailable(ctx, client, link); err == nil {
		return waybackRaw + ts + "id_/" + link
	}
	if ts, err := archiveIndexed(ctx, client, link); err == nil {
		return waybackRaw + ts + "id_/" + link
	}
	return waybackRaw + "2id_/" + link
}

// archiveGet fetches a small JSON answer from the archive into v.
func archiveGet(ctx context.Context, client *http.Client, endpoint string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", archiveAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w (HTTP %d)", errNoArchive, resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(v); err != nil {
		return fmt.Errorf("%w (unreadable answer)", errNoArchive)
	}
	return nil
}

// archiveAvailable asks the availability API for the closest capture.
func archiveAvailable(ctx context.Context, client *http.Client, link string) (string, error) {
	var answer struct {
		Snapshots struct {
			Closest struct {
				Available bool   `json:"available"`
				Status    string `json:"status"`
				Timestamp string `json:"timestamp"`
			} `json:"closest"`
		} `json:"archived_snapshots"`
	}
	if err := archiveGet(ctx, client, waybackAPI+url.QueryEscape(link), &answer); err != nil {
		return "", err
	}
	c := answer.Snapshots.Closest
	if !c.Available || c.Status != "200" || !isDigits(c.Timestamp) {
		return "", errNoArchive
	}
	return c.Timestamp, nil
}

// archiveIndexed asks the capture index for the newest capture of link
// that was saved as a PDF. The answer is a header row, then one row per
// capture: [["timestamp"],["20240102030405"]].
func archiveIndexed(ctx context.Context, client *http.Client, link string) (string, error) {
	var rows [][]string
	if err := archiveGet(ctx, client, waybackCDX+url.QueryEscape(link), &rows); err != nil {
		return "", err
	}
	if len(rows) < 2 || len(rows[len(rows)-1]) == 0 || !isDigits(rows[len(rows)-1][0]) {
		return "", errNoArchive
	}
	return rows[len(rows)-1][0], nil
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
	n, aerr := downloadPDF(ctx, client, archiveURL(ctx, client, link), dest, limit, progress)
	if aerr != nil {
		return 0, false, fmt.Errorf("%w; the Internet Archive's copy failed too: %v", err, aerr)
	}
	return n, true, nil
}
