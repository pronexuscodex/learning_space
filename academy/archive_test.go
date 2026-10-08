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
// saved copy is downloaded instead, with the same checks; when there is
// no saved copy, the original error is reported.
func TestArchiveFallback(t *testing.T) {
	pdf := []byte("%PDF-1.7\nsaved by the archive\n%%EOF\n")
	mux := http.NewServeMux()
	mux.HandleFunc("/gone.pdf", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	mux.HandleFunc("/never-archived.pdf", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	mux.HandleFunc("/available", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("url"), "never-archived") {
			io.WriteString(w, `{"archived_snapshots":{}}`)
			return
		}
		io.WriteString(w, `{"archived_snapshots":{"closest":{"status":"200","available":true,"url":"http://web.archive.org/web/20240102030405/x","timestamp":"20240102030405"}}}`)
	})
	mux.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/web/20240102030405id_/") {
			http.NotFound(w, r)
			return
		}
		w.Write(pdf)
	})
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
	srv.StartTLS()
	defer srv.Close()
	client := srv.Client()
	client.CheckRedirect = newDownloadClient().CheckRedirect

	oldAPI, oldRaw := waybackAPI, waybackRaw
	waybackAPI, waybackRaw = srv.URL+"/available?url=", srv.URL+"/web/"
	defer func() { waybackAPI, waybackRaw = oldAPI, oldRaw }()

	dir := t.TempDir()
	dest := filepath.Join(dir, "book.pdf")
	n, archived, err := downloadWithFallback(context.Background(), client, srv.URL+"/gone.pdf", dest, 1<<20, nil)
	if err != nil || !archived || n != int64(len(pdf)) {
		t.Fatalf("broken original: n=%d archived=%v err=%v", n, archived, err)
	}
	if got, _ := os.ReadFile(dest); string(got) != string(pdf) {
		t.Fatal("the archive's copy was not saved")
	}

	_, archived, err = downloadWithFallback(context.Background(), client, srv.URL+"/never-archived.pdf", filepath.Join(dir, "x.pdf"), 1<<20, nil)
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
