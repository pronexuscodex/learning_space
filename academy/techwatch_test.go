package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Example</title>
<item><title>Go 1.99 released &amp; faster</title><link>https://example.org/go199</link>
<pubDate>Tue, 22 Sep 2026 10:00:00 +0000</pubDate>
<description><![CDATA[<p>Big <b>news</b>&nbsp;today.</p>]]></description></item>
<item><title>Link only in guid</title><guid isPermaLink="true">https://example.org/guid-post</guid>
<pubDate>Mon, 21 Sep 2026 10:00:00 GMT</pubDate></item>
<item><title>Evil` + "\x1b[2J" + ` title</title><link>javascript:alert(1)</link></item>
<item><title></title><link>https://example.org/untitled</link></item>
</channel></rss>`

const sampleAtom = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom"><title>Atom Example</title>
<entry><title>Paged attention, explained</title>
<link rel="self" href="https://example.org/self"/><link rel="alternate" href="https://example.org/paged"/>
<updated>2026-09-23T08:00:00Z</updated><summary>Virtual memory for KV caches.</summary></entry>
<entry><title type="html">Second &lt;em&gt;post&lt;/em&gt;</title><link href="https://example.org/second"/>
<published>2026-09-20T08:00:00+02:00</published><content type="html">&lt;p&gt;Body&lt;/p&gt;</content></entry>
</feed>`

const sampleRDF = `<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/" xmlns:dc="http://purl.org/dc/elements/1.1/">
<channel><title>RDF</title></channel>
<item><title>An RDF item</title><link>https://example.org/rdf</link><dc:date>2026-09-19T00:00:00Z</dc:date></item>
</rdf:RDF>`

func TestParseFeeds(t *testing.T) {
	f := Feed{Title: "Ex", Topic: "Languages"}
	items, err := parseFeed([]byte(sampleRSS), f)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("RSS items = %d: %+v", len(items), items)
	}
	if items[0].Title != "Go 1.99 released & faster" || items[0].Summary != "Big news today." || items[0].Date.Day() != 22 {
		t.Errorf("RSS item 0 = %+v", items[0])
	}
	if items[1].Link != "https://example.org/guid-post" || items[1].Date.IsZero() {
		t.Errorf("GUID fallback or GMT date failed: %+v", items[1])
	}
	if items[2].Link != "" || strings.ContainsRune(items[2].Title, 0x1b) {
		t.Errorf("hostile item not neutralised: %+v", items[2])
	}

	items, err = parseFeed([]byte(sampleAtom), f)
	if err != nil || len(items) != 2 {
		t.Fatalf("Atom: %v %+v", err, items)
	}
	if items[0].Link != "https://example.org/paged" || items[0].Summary != "Virtual memory for KV caches." {
		t.Errorf("Atom alternate link or summary: %+v", items[0])
	}
	if items[1].Title != "Second post" {
		t.Errorf("HTML in an Atom title should be stripped: %q", items[1].Title)
	}
	if items[1].Summary != "Body" || items[1].Date.IsZero() {
		t.Errorf("Atom content fallback or published date: %+v", items[1])
	}

	items, err = parseFeed([]byte(sampleRDF), f)
	if err != nil || len(items) != 1 || items[0].Date.IsZero() {
		t.Fatalf("RDF: %v %+v", err, items)
	}

	latin := []byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><rss><channel><item><title>Caf\xe9</title></item></channel></rss>")
	if items, err := parseFeed(latin, f); err != nil || len(items) != 1 || items[0].Title != "Café" {
		t.Fatalf("latin-1: %v %+v", err, items)
	}
	if _, err := parseFeed([]byte("<html><body>not a feed</body></html>"), f); err == nil {
		t.Fatal("an HTML page is not a feed")
	}
	if _, err := parseFeed([]byte("plain text"), f); err == nil {
		t.Fatal("plain text is not a feed")
	}
}

func TestFetchFeedsAndHeadlines(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(sampleRSS)) })
	mux.HandleFunc("/atom", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(sampleAtom)) })
	mux.HandleFunc("/gone", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.ErrorLog = log.New(io.Discard, "", 0) // closed-connection noise
	srv.StartTLS()
	defer srv.Close()
	feeds := []Feed{
		{Title: "R", URL: srv.URL + "/rss", Topic: "Languages"},
		{Title: "A", URL: srv.URL + "/atom", Topic: "AI"},
		{Title: "G", URL: srv.URL + "/gone", Topic: "AI"},
		{Title: "H", URL: "http://example.org/feed", Topic: "AI"},
	}
	results := fetchFeeds(srv.Client(), feeds, 5*time.Second)
	if results[0].Err != nil || results[1].Err != nil || len(results[0].Items) != 3 {
		t.Fatalf("good feeds failed: %+v", results[:2])
	}
	if results[2].Err == nil || results[3].Err == nil {
		t.Fatal("404 and plain http must fail")
	}
	all := headlines(results, 2)
	if len(all) != 4 {
		t.Fatalf("perFeed limit: got %d headlines", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i].Date.After(all[i-1].Date) {
			t.Fatal("headlines must be newest first")
		}
	}
	if all[0].Title != "Paged attention, explained" {
		t.Fatalf("newest = %q", all[0].Title)
	}
}

func TestWatchLogRadarAndFeeds(t *testing.T) {
	reg := seedRegistry()
	now := time.Now()
	if _, err := reg.saveWatch(WatchItem{Title: "PagedAttention", Topic: "AI"}, now); err == nil {
		t.Fatal("an item without 'why it matters' must be refused")
	}
	id, err := reg.saveWatch(WatchItem{Title: "PagedAttention", URL: "https://example.org/p", Topic: "AI", Why: "Stage 6 virtual memory, applied to GPUs"}, now)
	if err != nil || id != 1 || reg.Watch[0].Status != WatchToRead {
		t.Fatalf("save: %v %+v", err, reg.Watch)
	}
	if _, err := reg.saveWatch(WatchItem{Title: "dup", URL: "https://example.org/p", Topic: "AI", Why: "x"}, now); err == nil {
		t.Fatal("the same link twice must be refused")
	}
	if _, err := reg.saveWatch(WatchItem{Title: "x", Topic: "Gossip", Why: "x"}, now); err == nil {
		t.Fatal("unknown topics must be refused")
	}
	reg.watchByID(1).Ring = "trial"
	if r := reg.radar(); len(r["trial"]) != 1 || len(r["adopt"]) != 0 {
		t.Fatalf("radar = %+v", r)
	}
	if !reg.watchedRecently(now) || reg.watchedRecently(now.Add(8*24*time.Hour)) {
		t.Fatal("watchedRecently should cover exactly the last 7 days")
	}

	if err := reg.addFeed(Feed{Title: "Mine", URL: "https://example.org/feed.xml", Topic: "Systems"}); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []Feed{
		{Title: "", URL: "https://example.org/a", Topic: "Systems"},
		{Title: "x", URL: "http://example.org/a", Topic: "Systems"},
		{Title: "x", URL: "https://example.org/a", Topic: "Nope"},
		{Title: "dup", URL: "https://example.org/feed.xml/", Topic: "Systems"},
		{Title: "dup curated", URL: curatedFeeds[0].URL, Topic: "Research"},
	} {
		if err := reg.addFeed(bad); err == nil {
			t.Errorf("accepted bad feed %+v", bad)
		}
	}
	if len(reg.allFeeds()) != len(curatedFeeds)+1 {
		t.Fatal("allFeeds should include the learner's feed")
	}
}

func TestCuratedFeedsAreSound(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range curatedFeeds {
		if !strings.HasPrefix(f.URL, "https://") || f.Title == "" || f.Why == "" || !contains(watchTopics, f.Topic) || seen[f.URL] {
			t.Errorf("bad curated feed %+v", f)
		}
		seen[f.URL] = true
	}
	for _, topic := range watchTopics {
		n := 0
		for _, f := range curatedFeeds {
			if f.Topic == topic {
				n++
			}
		}
		if n == 0 {
			t.Errorf("topic %s has no curated feed", topic)
		}
	}
}

func TestWatchSuggestion(t *testing.T) {
	reg := seedRegistry()
	now := time.Now()
	has := func() bool {
		for _, s := range reg.nextSteps(now) {
			if s.Kind == SuggestWatch {
				return true
			}
		}
		return false
	}
	if has() {
		t.Fatal("beginners should not be pushed to tech watch")
	}
	_, s := reg.findStage(1)
	g, _ := guideFor(1)
	for _, c := range g.Concepts {
		s.setStudied(c.Name, true)
	}
	reg.ReviewHistory[now.Format("2006-01-02")] = 50 // keep the list short: nothing else due today
	for id := range reg.Reviews {
		delete(reg.Reviews, id)
	}
	reg.saveWatch(WatchItem{Title: "x", Topic: "AI", Why: "y"}, now)
	if has() {
		t.Fatal("no watch suggestion right after a watch session")
	}
}
