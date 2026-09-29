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
	"time"
)

func TestPDFURL(t *testing.T) {
	cases := map[string]string{
		"https://arxiv.org/abs/1706.03762":       "https://arxiv.org/pdf/1706.03762",
		"https://www.arxiv.org/abs/2205.14135v2": "https://arxiv.org/pdf/2205.14135v2",
		"https://example.org/papers/Thing.PDF":   "https://example.org/papers/Thing.PDF",
		"https://example.org/book/":              "",
		"http://example.org/insecure.pdf":        "",
		"https://arxiv.org/abs/":                 "",
		"https://arxiv.org/list/cs.LG/recent":    "",
	}
	for in, want := range cases {
		got, ok := pdfURL(in)
		if got != want || ok != (want != "") {
			t.Errorf("pdfURL(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
}

func TestLibraryCatalogue(t *testing.T) {
	docs := libraryDocs()
	if len(docs) < 20 {
		t.Fatalf("expected at least 20 documents, got %d", len(docs))
	}
	names := map[string]bool{}
	for i, d := range docs {
		if !strings.HasPrefix(d.PDF, "https://") || d.Title == "" || d.Stage < 0 || d.Stage > lastStageID() {
			t.Errorf("bad document %+v", d)
		}
		key := strings.ToLower(d.relPath()) // Windows and macOS ignore case
		if names[key] {
			t.Errorf("duplicate file name %s", d.relPath())
		}
		names[key] = true
		if i > 0 && docs[i-1].Stage > d.Stage {
			t.Error("documents should be ordered by stage")
		}
	}
	// Every free book whose home page is a stage resource is listed under that stage.
	for _, b := range freeBooks {
		for id, g := range curriculum {
			for _, r := range g.Resources {
				if r.URL == b.Page && id != b.Stage {
					t.Errorf("%s is a Stage %d resource but filed under Stage %d", b.Title, id, b.Stage)
				}
			}
		}
		if !hasPDF(b.Page) {
			t.Errorf("hasPDF(%s) = false", b.Page)
		}
	}
	want := filepath.Join("Stage 03 - Discrete Mathematics, Logic & Probability", "Mathematics for Computer Science!.pdf")
	if got := (LibraryDoc{Stage: 3, Title: "Mathematics for Computer Science!"}).relPath(); got != want {
		t.Errorf("relPath = %q, want %q", got, want)
	}
	if slug("../../etc/passwd", 60) != "etc-passwd" {
		t.Error("slug must not keep path separators")
	}
}

func TestDownloadPDF(t *testing.T) {
	pdf := []byte("%PDF-1.7\n" + strings.Repeat("x", 100000) + "\n%%EOF\n")
	mux := http.NewServeMux()
	mux.HandleFunc("/ok.pdf", func(w http.ResponseWriter, r *http.Request) { w.Write(pdf) })
	mux.HandleFunc("/page.pdf", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("<html>moved</html>")) })
	mux.HandleFunc("/gone.pdf", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	mux.HandleFunc("/insecure.pdf", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://example.invalid/x.pdf", http.StatusFound)
	})
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.ErrorLog = log.New(io.Discard, "", 0) // closed-connection noise
	srv.StartTLS()
	defer srv.Close()
	client := srv.Client()
	client.CheckRedirect = newDownloadClient().CheckRedirect
	dir := t.TempDir()
	ctx := context.Background()

	dest := filepath.Join(dir, "lib", "ok.pdf")
	var calls int
	var last int64
	n, err := downloadPDF(ctx, client, srv.URL+"/ok.pdf", dest, 1<<20, func(done, total int64) { calls++; last = done })
	if err != nil || n != int64(len(pdf)) {
		t.Fatalf("download: n=%d err=%v", n, err)
	}
	if got, _ := os.ReadFile(dest); string(got) != string(pdf) {
		t.Fatal("saved file differs from the served PDF")
	}
	if calls == 0 || last != int64(len(pdf)) {
		t.Fatalf("progress not reported to completion: calls=%d last=%d", calls, last)
	}

	for name, want := range map[string]error{
		"/page.pdf":     errNotPDF,
		"/gone.pdf":     errBadStatus,
		"/insecure.pdf": errNotHTTPS,
	} {
		_, err := downloadPDF(ctx, client, srv.URL+name, filepath.Join(dir, "lib", "bad.pdf"), 1<<20, nil)
		if !errors.Is(err, want) {
			t.Errorf("%s: err = %v, want %v", name, err, want)
		}
	}
	if _, err := downloadPDF(ctx, client, srv.URL+"/ok.pdf", filepath.Join(dir, "lib", "big.pdf"), 1000, nil); !errors.Is(err, errTooLarge) {
		t.Errorf("size limit: err = %v", err)
	}
	if _, err := downloadPDF(ctx, client, "http://example.org/x.pdf", filepath.Join(dir, "x.pdf"), 1<<20, nil); !errors.Is(err, errNotHTTPS) {
		t.Errorf("plain http: err = %v", err)
	}

	// Nothing but the good file is left behind: no failed files, no partials.
	entries, _ := os.ReadDir(filepath.Join(dir, "lib"))
	if len(entries) != 1 || entries[0].Name() != "ok.pdf" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("library folder holds %v", names)
	}
}

func TestOwnDownloadsAndPartials(t *testing.T) {
	dir := t.TempDir()
	cat := libraryDocs()
	put := func(rel string) {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte("%PDF-"), 0o644)
	}
	put(cat[0].relPath())
	put(filepath.Join(myPDFsFolder, "my notes.pdf"))
	put(filepath.Join("Arranged by me", "Exam prep", "past paper.PDF")) // a folder the learner made
	put(filepath.Join(stageFolder(1), ".academy-download-123.part"))
	own := ownDownloads(dir, cat)
	if len(own) != 2 || own[0].Title != "past paper" || own[1].Title != "my notes" {
		t.Fatalf("own downloads = %+v", own)
	}
	if own[1].relPath() != filepath.Join(myPDFsFolder, "my notes.pdf") {
		t.Errorf("own download path = %q", own[1].relPath())
	}
	if _, ok := cat[0].downloaded(dir); !ok {
		t.Fatal("catalogue file should count as downloaded")
	}
	cleanPartials(dir)
	if _, err := os.Stat(filepath.Join(dir, stageFolder(1), ".academy-download-123.part")); !os.IsNotExist(err) {
		t.Fatal("partial download should be removed, in any sub-folder")
	}
}

func TestLibraryFolderLocation(t *testing.T) {
	reg := filepath.FromSlash("/data/academy/r.json")
	docs := func() (string, error) { return filepath.FromSlash("/home/ada/Documents"), nil }
	noDocs := func() (string, error) { return "", os.ErrNotExist }
	env := func(kv ...string) func(string) string {
		return func(k string) string {
			for i := 0; i+1 < len(kv); i += 2 {
				if kv[i] == k {
					return kv[i+1]
				}
			}
			return ""
		}
	}
	for _, c := range []struct {
		name string
		got  string
		want string
	}{
		{"Documents by default", libraryDirFor(reg, env(), docs), "/home/ada/Documents/Academy Library"},
		{"ACADEMY_LIBRARY wins", libraryDirFor(reg, env("ACADEMY_LIBRARY", filepath.FromSlash("/pdfs"), "ACADEMY_HOME", filepath.FromSlash("/study")), docs), "/pdfs"},
		{"inside ACADEMY_HOME", libraryDirFor(reg, env("ACADEMY_HOME", filepath.FromSlash("/study")), docs), "/study/Academy Library"},
		{"no home folder: beside the registry", libraryDirFor(reg, env(), noDocs), "/data/academy/Academy Library"},
	} {
		if c.got != filepath.FromSlash(c.want) {
			t.Errorf("%s: %s, want %s", c.name, c.got, filepath.FromSlash(c.want))
		}
	}
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	os.MkdirAll(filepath.Join(home, ".config"), 0o755)
	os.WriteFile(filepath.Join(home, ".config", "user-dirs.dirs"), []byte("# written by xdg-user-dirs-update\nXDG_DOCUMENTS_DIR=\"$HOME/Dokumente\"\n"), 0o644)
	if got := xdgDocuments(home); got != filepath.Join(home, "Dokumente") {
		t.Errorf("xdgDocuments = %q", got)
	}
}

func TestCleanNameIsSafeEverywhere(t *testing.T) {
	for in, want := range map[string]string{
		"NIST SP 800-207: Zero Trust Architecture": "NIST SP 800-207 - Zero Trust Architecture",
		"../../etc/passwd":                         ".. .. etc passwd",
		`a\b/c<d>e"f|g?h*i`:                        "a b c d e f g h i",
		"trailing dots...":                         "trailing dots",
		"CON":                                      "_CON",
		"Networks, Sockets, & Distributed Topology": "Networks, Sockets, & Distributed Topology",
		"bell\x07 and\u202eflip":                    "bell and flip",
	} {
		if got := cleanName(in, 100); got != want {
			t.Errorf("cleanName(%q) = %q, want %q", in, got, want)
		}
	}
	if got := cleanName(strings.Repeat("x", 300), 100); len(got) != 100 {
		t.Errorf("long names are cut to 100, got %d", len(got))
	}
}

func TestLegacyLibraryIsMovedAndArranged(t *testing.T) {
	root := t.TempDir()
	reg := filepath.Join(root, "academy_campus_registry.json")
	old := filepath.Join(root, legacyLibraryName)
	os.MkdirAll(old, 0o755)
	doc := libraryDocs()[0]
	os.WriteFile(filepath.Join(old, doc.legacyFileName()), []byte("%PDF-catalogue"), 0o644)
	os.WriteFile(filepath.Join(old, "my-own-paper.pdf"), []byte("%PDF-own"), 0o644)
	os.WriteFile(filepath.Join(old, ".academy-download-9.part"), []byte("%PD"), 0o644)
	lib := filepath.Join(root, "Documents", libraryFolderName)

	n, err := migrateLegacyLibrary(reg, lib)
	if err != nil || n != 2 {
		t.Fatalf("moved %d, err %v", n, err)
	}
	if b, _ := os.ReadFile(filepath.Join(lib, doc.relPath())); string(b) != "%PDF-catalogue" {
		t.Error("the catalogue PDF should be in its stage folder with a readable name")
	}
	if b, _ := os.ReadFile(filepath.Join(lib, myPDFsFolder, "my-own-paper.pdf")); string(b) != "%PDF-own" {
		t.Error("the learner's own PDF should be in My PDFs")
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("the empty old folder should be removed")
	}
	if n, err := migrateLegacyLibrary(reg, lib); n != 0 || err != nil {
		t.Errorf("a second run should do nothing: %d, %v", n, err)
	}
}

func TestDownloadStallAndCancel(t *testing.T) {
	// The server sends the PDF signature, then nothing.
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100000")
		w.Write([]byte("%PDF-1.7\n"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
	srv.StartTLS()
	defer srv.Close()
	client := srv.Client()
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib")

	defer func(old time.Duration) { stallTimeout = old }(stallTimeout)
	stallTimeout = 300 * time.Millisecond
	start := time.Now()
	_, err := downloadPDF(context.Background(), client, srv.URL+"/slow.pdf", filepath.Join(lib, "a.pdf"), 1<<20, nil)
	if err == nil || !strings.Contains(err.Error(), "sent nothing") {
		t.Fatalf("stalled download: got %v, want a stall error", err)
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("stall detected after %s", d)
	}

	// Ctrl+C cancels with errCancelled, long before the stall timeout.
	stallTimeout = time.Minute
	ctx, cancel := context.WithCancelCause(context.Background())
	time.AfterFunc(200*time.Millisecond, func() { cancel(errCancelled) })
	start = time.Now()
	_, err = downloadPDF(ctx, client, srv.URL+"/slow.pdf", filepath.Join(lib, "b.pdf"), 1<<20, nil)
	if !errors.Is(err, errCancelled) {
		t.Fatalf("cancelled download: got %v, want errCancelled", err)
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("cancel took %s", d)
	}

	// Neither left anything behind.
	if left, _ := os.ReadDir(lib); len(left) != 0 {
		t.Errorf("files left in the library: %v", left)
	}
}

// Every stage's "PDF library" entry must lead somewhere: no empty screens.
func TestEveryStageHasLibraryPDFs(t *testing.T) {
	count := map[int]int{}
	for _, d := range libraryDocs() {
		count[d.Stage]++
	}
	for id := range curriculum {
		if count[id] == 0 {
			t.Errorf("stage %d has no downloadable PDF in its library", id)
		}
	}
}
