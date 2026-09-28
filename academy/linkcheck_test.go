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
		}
	}))
	defer srv.Close()
	for path, want := range map[string]string{"/ok": "ok", "/bots-not-welcome": "guarded", "/gone": "failed", "/head-refused": "ok"} {
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
	// Classic-corner links are checked too.
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
