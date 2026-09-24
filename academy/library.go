package main

// The offline library: free, legally hosted PDFs from the curriculum
// (direct PDF links, arXiv papers and classic texts) downloaded into a
// folder beside the registry, so they can be read without hunting for
// them in a browser. Only the standard library is used: net/http honours
// HTTPS_PROXY, and os/exec opens the system PDF viewer.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	libraryDirName  = "academy_library"
	maxDownloadSize = 200 << 20 // 200 MB: far above any book here, well below a runaway response
	partPattern     = ".academy-download-*.part"
)

// LibraryDoc is one downloadable PDF.
type LibraryDoc struct {
	Stage int
	Kind  string // Book, Paper, Article, Classic …
	Title string
	Page  string // the link as listed in the curriculum
	PDF   string // the direct PDF address
}

// freeBooks are whole textbooks whose authors or publishers give the PDF
// away at an official address (checked 2026-09-24). Page is the book's
// home page as listed in the stage's resources, when it is listed.
var freeBooks = []LibraryDoc{
	{1, "Book", "Think Python (2nd edition)", "https://greenteapress.com/wp/think-python-2e/", "https://greenteapress.com/thinkpython2/thinkpython2.pdf"},
	{2, "Book", "Algorithms by Jeff Erickson", "https://jeffe.cs.illinois.edu/teaching/algorithms/", "https://jeffe.cs.illinois.edu/teaching/algorithms/book/Algorithms-JeffE.pdf"},
	{3, "Book", "Mathematics for Computer Science", "https://courses.csail.mit.edu/6.042/spring18/", "https://courses.csail.mit.edu/6.042/spring18/mcs.pdf"},
	{8, "Book", "Beej's Guide to Network Programming", "https://beej.us/guide/bgnet/", "https://beej.us/guide/bgnet/pdf/bgnet_usl_c_1.pdf"},
	{13, "Book", "Mathematics for Machine Learning", "https://mml-book.github.io/", "https://mml-book.github.io/book/mml-book.pdf"},
	{13, "Book", "Linear Algebra Done Right (4th edition)", "https://linear.axler.net/", "https://linear.axler.net/LADR4e.pdf"},
	{14, "Book", "An Introduction to Statistical Learning (Python edition)", "https://www.statlearning.com/", "https://hastie.su.domains/ISLP/ISLP_website.pdf"},
}

// pdfURL returns the direct PDF address for a link, if it has one: links
// that already end in .pdf, and arXiv abstract pages (arxiv.org/abs/ID →
// arxiv.org/pdf/ID).
func pdfURL(link string) (string, bool) {
	u, err := url.Parse(link)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", false
	}
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	if host == "arxiv.org" && strings.HasPrefix(u.Path, "/abs/") {
		id := strings.TrimPrefix(u.Path, "/abs/")
		if id == "" {
			return "", false
		}
		return "https://arxiv.org/pdf/" + id, true
	}
	if strings.HasSuffix(strings.ToLower(u.Path), ".pdf") {
		return link, true
	}
	return "", false
}

// hasPDF reports whether a resource link has a PDF in the library.
func hasPDF(link string) bool {
	if link == "" {
		return false
	}
	if _, ok := pdfURL(link); ok {
		return true
	}
	for _, b := range freeBooks {
		if b.Page == link {
			return true
		}
	}
	return false
}

// libraryDocs lists every downloadable PDF in the curriculum, by stage,
// without duplicates.
func libraryDocs() []LibraryDoc {
	var docs []LibraryDoc
	seen := map[string]bool{}
	add := func(stage int, kind, title, page string) {
		pdf, ok := pdfURL(page)
		if !ok || seen[pdf] {
			return
		}
		seen[pdf] = true
		if kind == "Site" || kind == "Tool" {
			kind = "Reference"
		}
		docs = append(docs, LibraryDoc{Stage: stage, Kind: kind, Title: title, Page: page, PDF: pdf})
	}
	for _, b := range freeBooks {
		seen[b.PDF] = true
		docs = append(docs, b)
	}
	for id, g := range curriculum {
		for _, r := range g.Resources {
			add(id, r.Kind, r.Title, r.URL)
		}
		if cg, _, ok := classicFor(id); ok {
			for _, cr := range []ClassicRead{cg.Anchor, cg.Classic, cg.Source} {
				add(id, "Classic", cr.Title, cr.URL)
			}
		}
	}
	sort.SliceStable(docs, func(i, j int) bool {
		if docs[i].Stage != docs[j].Stage {
			return docs[i].Stage < docs[j].Stage
		}
		return docs[i].Title < docs[j].Title
	})
	return docs
}

// slug turns a title into a short, safe file-name part.
func slug(s string, maxLen int) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
		if b.Len() >= maxLen {
			break
		}
	}
	return strings.Trim(b.String(), "-")
}

// fileName is where a document is stored inside the library folder.
func (d LibraryDoc) fileName() string {
	name := slug(d.Title, 60)
	if name == "" {
		name = "document"
	}
	if d.Stage > 0 {
		return fmt.Sprintf("stage%02d-%s.pdf", d.Stage, name)
	}
	return name + ".pdf"
}

// libraryDir is the library folder for the registry at path.
func libraryDir(path string) string { return filepath.Join(filepath.Dir(path), libraryDirName) }

// downloaded reports whether the document is already in the library, and
// its size.
func (d LibraryDoc) downloaded(dir string) (int64, bool) {
	st, err := os.Stat(filepath.Join(dir, d.fileName()))
	if err != nil || !st.Mode().IsRegular() {
		return 0, false
	}
	return st.Size(), true
}

// cleanPartials removes downloads that were interrupted.
func cleanPartials(dir string) {
	matches, _ := filepath.Glob(filepath.Join(dir, partPattern))
	for _, m := range matches {
		os.Remove(m)
	}
}

// Download errors a learner can act on.
var (
	errNotHTTPS  = errors.New("only https:// links are downloaded")
	errNotPDF    = errors.New("the server did not send a PDF (the file may have moved; open the link in a browser)")
	errTooLarge  = errors.New("the file is larger than the 200 MB limit")
	errBadStatus = errors.New("the server refused the download")
)

// newDownloadClient is an HTTP client that follows at most 10 redirects,
// and only to other https:// addresses.
func newDownloadClient() *http.Client {
	return &http.Client{
		Transport: http.DefaultTransport, // honours HTTPS_PROXY
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "https" {
				return errNotHTTPS
			}
			return nil
		},
	}
}

// progressFunc reports bytes received so far and the total (-1 if unknown).
type progressFunc func(done, total int64)

// progressReader counts bytes as they are read.
type progressReader struct {
	r           io.Reader
	done, total int64
	report      progressFunc
	last        time.Time
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.done += int64(n)
	if p.report != nil && (time.Since(p.last) > 100*time.Millisecond || err == io.EOF) {
		p.last = time.Now()
		p.report(p.done, p.total)
	}
	return n, err
}

// downloadPDF fetches link into dest. The file is streamed into a
// temporary file in the same folder and renamed into place only after it
// has been checked to be a complete PDF, so an interrupted or bad download
// never leaves a broken file behind.
func downloadPDF(ctx context.Context, client *http.Client, link, dest string, limit int64, progress progressFunc) (n int64, err error) {
	u, err := url.Parse(link)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return 0, errNotHTTPS
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "academy-library/1.0 (self-study; one copy for personal reading)")
	req.Header.Set("Accept", "application/pdf,*/*;q=0.5")
	resp, err := client.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			if errors.Is(ue.Err, errNotHTTPS) {
				return 0, errNotHTTPS
			}
			return 0, fmt.Errorf("could not reach %s: %v", u.Host, ue.Err)
		}
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%w (HTTP %d)", errBadStatus, resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return 0, errTooLarge
	}

	// The first bytes must be the PDF signature, whatever the headers say.
	head := make([]byte, 5)
	if _, err := io.ReadFull(resp.Body, head); err != nil || !bytes.Equal(head, []byte("%PDF-")) {
		return 0, errNotPDF
	}

	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}
	tmp, err := os.CreateTemp(dir, partPattern)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
		}
	}()

	body := &progressReader{r: io.MultiReader(bytes.NewReader(head), resp.Body), total: resp.ContentLength, report: progress}
	n, err = io.Copy(tmp, io.LimitReader(body, limit+1))
	if err != nil {
		return 0, fmt.Errorf("download interrupted: %w", err)
	}
	if n > limit {
		return 0, errTooLarge
	}
	if resp.ContentLength > 0 && n != resp.ContentLength {
		return 0, fmt.Errorf("download incomplete: got %d of %d bytes", n, resp.ContentLength)
	}
	if err = tmp.Sync(); err != nil {
		return 0, err
	}
	if err = tmp.Close(); err != nil {
		return 0, err
	}
	if err = os.Chmod(tmp.Name(), 0o644); err != nil {
		return 0, err
	}
	if err = os.Rename(tmp.Name(), dest); err != nil {
		return 0, err
	}
	return n, nil
}

// openCommand is the platform's "open this file with its default app".
func openCommand(path string) *exec.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path)
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		return exec.Command("xdg-open", path)
	}
}

// openInViewer starts the system PDF viewer without waiting for it.
func openInViewer(path string) error {
	cmd := openCommand(path)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait() // reap it; the viewer's exit status does not matter
	return nil
}

// humanBytes formats a size for people.
func humanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}

// fetchLibrary implements -fetch-library: download every document that is
// not in the library yet. It returns the exit code.
func fetchLibrary(regPath string, out io.Writer) int {
	dir := libraryDir(regPath)
	cleanPartials(dir)
	client := newDownloadClient()
	failed := 0
	for _, d := range libraryDocs() {
		dest := filepath.Join(dir, d.fileName())
		if size, ok := d.downloaded(dir); ok {
			fmt.Fprintf(out, "  = %-60s already here (%s)\n", truncate(d.Title, 60), humanBytes(size))
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		n, err := downloadPDF(ctx, client, d.PDF, dest, maxDownloadSize, nil)
		cancel()
		if err != nil {
			failed++
			fmt.Fprintf(out, "  ✗ %-60s %v\n", truncate(d.Title, 60), err)
			continue
		}
		fmt.Fprintf(out, "  ✓ %-60s %s\n", truncate(d.Title, 60), humanBytes(n))
	}
	fmt.Fprintf(out, "Library: %s\n", dir)
	if failed > 0 {
		fmt.Fprintf(out, "%d download(s) failed. Links move; run -check-links, or open the page in a browser.\n", failed)
		return 1
	}
	return 0
}
