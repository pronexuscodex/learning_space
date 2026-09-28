package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLinkCheckClassifiesAnswers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
		case "/bots-not-welcome":
			w.WriteHeader(http.StatusForbidden)
		case "/gone":
			http.NotFound(w, r)
		case "/head-refused":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case "/real.pdf":
			w.Write([]byte("%PDF-1.7 ..."))
		case "/fake.pdf":
			w.Write([]byte("<html>Sorry, moved</html>"))
		case "/head-404":
			if r.Method == http.MethodHead {
				http.NotFound(w, r)
			}
		}
	}))
	defer srv.Close()
	for path, want := range map[string]string{"/ok": "ok", "/bots-not-welcome": "guarded", "/gone": "failed", "/head-refused": "ok", "/head-404": "ok"} {
		l := linkResult{URL: srv.URL + path}
		l.Status, l.Err = checkLink(srv.Client(), l.URL)
		got := "failed"
		if l.ok() {
			got = "ok"
		} else if l.guarded() {
			got = "guarded"
		}
		if got != want {
			t.Errorf("%s: %s (HTTP %d, %v), want %s", path, got, l.Status, l.Err, want)
		}
	}
	// Library files must really be PDFs, not error pages served as 200 OK.
	if _, err := checkPDF(srv.Client(), srv.URL+"/real.pdf"); err != nil {
		t.Errorf("real PDF: %v", err)
	}
	if _, err := checkPDF(srv.Client(), srv.URL+"/fake.pdf"); err == nil {
		t.Error("an HTML page must not pass as a PDF")
	}
	// Classic-corner links and every Library PDF are checked too.
	pdfs := 0
	for _, l := range allResourceLinks() {
		if l.PDF {
			pdfs++
		}
	}
	if pdfs != len(libraryDocs()) {
		t.Errorf("link check covers %d of %d Library PDFs", pdfs, len(libraryDocs()))
	}
	found := false
	for _, l := range allResourceLinks() {
		if l.URL == classicGuides[10].Classic.URL {
			found = true
		}
	}
	if !found {
		t.Error("classic-corner links should be part of the link check")
	}
}
