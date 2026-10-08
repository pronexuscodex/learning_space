package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// When a Library PDF's original link is broken, the Internet Archive's
// saved copy is downloaded instead, with the same checks. The copy is
// found through the availability API, else the capture index, else the
// Wayback Machine's own redirect to its newest capture; when there is no
// saved copy at all, the original error is reported.
func TestArchiveFallback(t *testing.T) {
	pdf := []byte("%PDF-1.7\nsaved by the archive\n%%EOF\n")
	mux := http.NewServeMux()
	for _, p := range []string{"/gone.pdf", "/indexed.pdf", "/redirect-only.pdf", "/never-archived.pdf"} {
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	}
	mux.HandleFunc("/available", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Query().Get("url"), "/gone.pdf") {
			io.WriteString(w, `{"archived_snapshots":{}}`) // the API sometimes says this for saved pages
			return
		}
		io.WriteString(w, `{"archived_snapshots":{"closest":{"status":"200","available":true,"url":"http://web.archive.org/web/20240102030405/x","timestamp":"20240102030405"}}}`)
	})
	mux.HandleFunc("/cdx", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Query().Get("url"), "/indexed.pdf") {
			io.WriteString(w, `[]`)
			return
		}
		io.WriteString(w, `[["timestamp"],["20210101000000"],["20240102030405"]]`)
	})
	mux.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/web/20240102030405id_/"):
			w.Write(pdf)
		case strings.HasPrefix(r.URL.Path, "/web/2id_/") && strings.HasSuffix(r.URL.Path, "/redirect-only.pdf"):
			http.Redirect(w, r, "/web/20240102030405id_/"+strings.TrimPrefix(r.URL.Path, "/web/2id_/"), http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	})
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
	srv.StartTLS()
	defer srv.Close()
	client := srv.Client()
	client.CheckRedirect = newDownloadClient().CheckRedirect

	oldAPI, oldCDX, oldRaw := waybackAPI, waybackCDX, waybackRaw
	waybackAPI, waybackCDX, waybackRaw = srv.URL+"/available?url=", srv.URL+"/cdx?url=", srv.URL+"/web/"
	defer func() { waybackAPI, waybackCDX, waybackRaw = oldAPI, oldCDX, oldRaw }()

	dir := t.TempDir()
	dest := filepath.Join(dir, "book.pdf")
	for _, name := range []string{"gone.pdf", "indexed.pdf", "redirect-only.pdf"} {
		os.Remove(dest)
		n, archived, err := downloadWithFallback(context.Background(), client, srv.URL+"/"+name, dest, 1<<20, nil)
		if err != nil || !archived || n != int64(len(pdf)) {
			t.Fatalf("%s: n=%d archived=%v err=%v", name, n, archived, err)
		}
		if got, _ := os.ReadFile(dest); string(got) != string(pdf) {
			t.Fatalf("%s: the archive's copy was not saved", name)
		}
	}

	_, archived, err := downloadWithFallback(context.Background(), client, srv.URL+"/never-archived.pdf", filepath.Join(dir, "x.pdf"), 1<<20, nil)
	if !errors.Is(err, errBadStatus) || archived {
		t.Fatalf("no saved copy: the original error must be reported, got archived=%v err=%v", archived, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "x.pdf")); !os.IsNotExist(err) {
		t.Fatal("a failed download must leave no file")
	}

	// A refused (non-HTTPS) link is never sent to the archive.
	if _, archived, err := downloadWithFallback(context.Background(), client, "http://example.invalid/a.pdf", filepath.Join(dir, "y.pdf"), 1<<20, nil); !errors.Is(err, errNotHTTPS) || archived {
		t.Fatalf("plain http must be refused outright, got archived=%v err=%v", archived, err)
	}
}

// TestArchiveLive checks the real Internet Archive, through the same code
// the Library uses. It needs the network, so it runs only when asked:
// ACADEMY_LIVE_ARCHIVE=1 (the weekly Links workflow does).
func TestArchiveLive(t *testing.T) {
	if os.Getenv("ACADEMY_LIVE_ARCHIVE") == "" {
		t.Skip("set ACADEMY_LIVE_ARCHIVE=1 to check the real Internet Archive")
	}
	const link = "https://jeffe.cs.illinois.edu/teaching/algorithms/book/Algorithms-JeffE.pdf"
	client := newDownloadClient()
	ctx := context.Background()
	ts, err := archiveAvailable(ctx, client, link)
	t.Logf("availability API: %q %v", ts, err)
	ts, err = archiveIndexed(ctx, client, link)
	t.Logf("capture index: %q %v", ts, err)
	for _, alt := range []string{archiveURL(ctx, client, link), waybackRaw + "2id_/" + link} {
		if status, err := checkPDF(client, alt); err != nil {
			t.Errorf("%s: HTTP %d: %v", alt, status, err)
		} else {
			t.Logf("✓ %s is a PDF", alt)
		}
	}
}
