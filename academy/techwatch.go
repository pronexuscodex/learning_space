package main

// Tech watch: a method for keeping up with the field, a curated set of
// high-signal feeds (RSS and Atom, read with encoding/xml), a watch log of
// saved items with why they matter, and a personal technology radar.

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Feed is one news source.
type Feed struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Topic string `json:"topic"`
	Why   string `json:"why,omitempty"`
}

// watchTopics are the topics feeds and saved items are filed under.
var watchTopics = []string{"Research", "Systems", "Languages", "Databases", "Security", "AI", "Engineering", "Community"}

// curatedFeeds are primary or long-respected sources, checked 2026-09-24.
var curatedFeeds = []Feed{
	{"arXiv: Machine Learning (cs.LG)", "https://rss.arxiv.org/rss/cs.LG", "Research", "Every new ML paper, daily. Skim titles; open only what connects to your stage."},
	{"arXiv: Operating Systems (cs.OS)", "https://rss.arxiv.org/rss/cs.OS", "Research", "New OS research: a few papers a day, many readable after Stage 6."},
	{"arXiv: Cryptography and Security (cs.CR)", "https://rss.arxiv.org/rss/cs.CR", "Research", "New security research."},
	{"LWN.net", "https://lwn.net/headlines/rss", "Systems", "The best technical reporting on the Linux kernel and free software."},
	{"Linux kernel releases", "https://www.kernel.org/feeds/kdist.xml", "Systems", "Every stable and long-term kernel release, straight from kernel.org."},
	{"Brendan Gregg", "https://www.brendangregg.com/blog/rss.xml", "Systems", "Performance analysis and observability from the author of Systems Performance."},
	{"Julia Evans", "https://jvns.ca/atom.xml", "Systems", "Friendly, curious deep dives into networking, Linux and debugging."},
	{"Dan Luu", "https://danluu.com/atom.xml", "Systems", "Long, data-driven essays on hardware, latency and how software is built."},
	{"The Go Blog", "https://go.dev/blog/feed.atom", "Languages", "Official Go release notes and design write-ups."},
	{"Rust Blog", "https://blog.rust-lang.org/feed.xml", "Languages", "Official Rust releases and announcements."},
	{"PostgreSQL news", "https://www.postgresql.org/news.rss", "Databases", "Official PostgreSQL releases and security fixes."},
	{"Krebs on Security", "https://krebsonsecurity.com/feed/", "Security", "Investigative reporting on real breaches and cybercrime."},
	{"Schneier on Security", "https://www.schneier.com/feed/atom/", "Security", "Security thinking from one of the field's best-known cryptographers."},
	{"Hugging Face blog", "https://huggingface.co/blog/feed.xml", "AI", "Open models, datasets and practical ML engineering."},
	{"Simon Willison", "https://simonwillison.net/atom/everything/", "AI", "Hands-on, sceptical notes on LLMs and tools, with working code."},
	{"Martin Fowler", "https://martinfowler.com/feed.atom", "Engineering", "Software design, refactoring and architecture."},
	{"The Cloudflare Blog", "https://blog.cloudflare.com/rss/", "Engineering", "Deep technical posts on networks, outages and the internet's plumbing."},
	{"Hacker News (front page)", "https://news.ycombinator.com/rss", "Community", "What programmers are discussing today. High volume: skim, don't scroll."},
	{"Lobsters", "https://lobste.rs/rss", "Community", "A smaller, more technical link community."},
}

// FeedItem is one headline.
type FeedItem struct {
	Feed    string
	Topic   string
	Title   string
	Link    string
	Date    time.Time
	Summary string
}

// Watch-log statuses and radar rings.
const (
	WatchToRead = "to read"
	WatchRead   = "read"
	WatchTried  = "tried"
	WatchDone   = "dropped"
)

var watchStatuses = []string{WatchToRead, WatchRead, WatchTried, WatchDone}
var radarRings = []string{"adopt", "trial", "assess", "hold"}
var radarMeaning = map[string]string{
	"adopt":  "use it by default",
	"trial":  "use it on something real but low-risk",
	"assess": "explore it; worth understanding",
	"hold":   "do not start with it now",
}

// WatchItem is one saved item in the watch log.
type WatchItem struct {
	ID     int       `json:"id"`
	Added  time.Time `json:"added"`
	Title  string    `json:"title"`
	URL    string    `json:"url,omitempty"`
	Source string    `json:"source,omitempty"`
	Topic  string    `json:"topic"`
	Why    string    `json:"why"` // why it matters to me
	Status string    `json:"status"`
	Ring   string    `json:"ring,omitempty"` // radar ring, if placed
	Note   string    `json:"note,omitempty"` // what I learned or tried
}

// allFeeds is the curated list plus the learner's own feeds.
func (r *Registry) allFeeds() []Feed {
	return append(append([]Feed{}, curatedFeeds...), r.Feeds...)
}

// addFeed validates and stores one of the learner's feeds.
func (r *Registry) addFeed(f Feed) error {
	f.Title, f.URL = strings.TrimSpace(f.Title), strings.TrimSpace(f.URL)
	u, err := url.Parse(f.URL)
	switch {
	case f.Title == "" || len([]rune(f.Title)) > maxNameLen:
		return errors.New("the feed needs a short title")
	case err != nil || u.Scheme != "https" || u.Host == "":
		return errors.New("the feed link must start with https://")
	case !contains(watchTopics, f.Topic):
		return fmt.Errorf("topic must be one of %s", strings.Join(watchTopics, ", "))
	}
	for _, have := range r.allFeeds() {
		if strings.EqualFold(strings.TrimRight(have.URL, "/"), strings.TrimRight(f.URL, "/")) {
			return fmt.Errorf("already following it as %q", have.Title)
		}
	}
	r.Feeds = append(r.Feeds, f)
	return nil
}

// saveWatch adds an item to the watch log and returns its ID.
func (r *Registry) saveWatch(w WatchItem, now time.Time) (int, error) {
	w.Title, w.Why = strings.TrimSpace(w.Title), strings.TrimSpace(w.Why)
	switch {
	case w.Title == "":
		return 0, errors.New("a title is required")
	case w.Why == "":
		return 0, errors.New("say in a few words why it matters to you; that is the point of the log")
	case len([]rune(w.Why)) > maxNotesLen:
		return 0, fmt.Errorf("keep it under %d characters", maxNotesLen)
	case !contains(watchTopics, w.Topic):
		return 0, fmt.Errorf("topic must be one of %s", strings.Join(watchTopics, ", "))
	}
	for _, have := range r.Watch {
		if w.URL != "" && have.URL == w.URL {
			return 0, fmt.Errorf("already in your log as #%d", have.ID)
		}
	}
	for _, have := range r.Watch {
		w.ID = max(w.ID, have.ID)
	}
	w.ID++
	w.Added = now.UTC().Truncate(time.Second)
	if w.Status == "" {
		w.Status = WatchToRead
	}
	r.Watch = append(r.Watch, w)
	return w.ID, nil
}

// watchByID returns a pointer into the log.
func (r *Registry) watchByID(id int) *WatchItem {
	for i := range r.Watch {
		if r.Watch[i].ID == id {
			return &r.Watch[i]
		}
	}
	return nil
}

// radar groups the placed items by ring.
func (r *Registry) radar() map[string][]WatchItem {
	out := map[string][]WatchItem{}
	for _, w := range r.Watch {
		if w.Ring != "" {
			out[w.Ring] = append(out[w.Ring], w)
		}
	}
	for _, ring := range radarRings {
		sort.SliceStable(out[ring], func(i, j int) bool { return out[ring][i].Topic < out[ring][j].Topic })
	}
	return out
}

// watchedRecently reports whether the learner saved a watch item in the
// last seven days.
func (r *Registry) watchedRecently(now time.Time) bool {
	for _, w := range r.Watch {
		if now.Sub(w.Added) < 7*24*time.Hour {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Feed parsing
// ---------------------------------------------------------------------------

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Date        string `xml:"date"` // dc:date (RSS 1.0)
	Description string `xml:"description"`
}

type rssDoc struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
	Items []rssItem `xml:"item"` // RSS 1.0 (RDF) puts items beside the channel
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	ID        string     `xml:"id"`
	Updated   string     `xml:"updated"`
	Published string     `xml:"published"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
}

type atomDoc struct {
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

var dateLayouts = []string{
	time.RFC1123Z, time.RFC1123, time.RFC3339, time.RFC3339Nano, time.RFC822Z, time.RFC822,
	"Mon, 2 Jan 2006 15:04:05 -0700", "Mon, 2 Jan 2006 15:04:05 MST", "2 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 GMT", "2006-01-02T15:04:05Z0700", "2006-01-02 15:04:05", "2006-01-02",
}

// parseFeedDate understands the date formats feeds use in practice.
func parseFeedDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

var tagRE = regexp.MustCompile(`<[^>]*>`)

// plainSummary turns an HTML description into one short line of text.
func plainSummary(s string, n int) string {
	s = tagRE.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.Join(strings.Fields(s), " ")
	return truncate(stripControl(s), n)
}

// stripControl removes control characters, including escape sequences a
// hostile feed could use to repaint the terminal.
func stripControl(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r < 0xa0) {
			return -1
		}
		return r
	}, s)
}

// latin1Reader converts ISO-8859-1 (and, closely enough, Windows-1252)
// feeds to UTF-8.
func latin1Reader(charset string, in io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "iso-8859-1", "latin1", "latin-1", "windows-1252", "us-ascii", "ascii":
		data, err := io.ReadAll(in)
		if err != nil {
			return nil, err
		}
		var b bytes.Buffer
		for _, c := range data {
			b.WriteRune(rune(c))
		}
		return &b, nil
	}
	return nil, fmt.Errorf("unsupported charset %q", charset)
}

// parseFeed reads RSS 2.0, RSS 1.0 (RDF) or Atom.
func parseFeed(data []byte, feed Feed) ([]FeedItem, error) {
	// Stray control bytes are common in real feeds and make XML invalid;
	// drop them (keeping tab, newline and carriage return).
	// Work on raw bytes: the feed may not be UTF-8 yet.
	cleaned := make([]byte, 0, len(data))
	for _, c := range data {
		if (c < 0x20 && c != '\t' && c != '\n' && c != '\r') || c == 0x7f {
			continue
		}
		cleaned = append(cleaned, c)
	}
	data = cleaned
	newDecoder := func() *xml.Decoder {
		d := xml.NewDecoder(bytes.NewReader(data))
		d.Strict = false
		d.Entity = xml.HTMLEntity
		d.CharsetReader = latin1Reader
		return d
	}
	// Find the root element to know the format.
	var root string
	d := newDecoder()
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, fmt.Errorf("not a feed: %v", err)
		}
		if se, ok := tok.(xml.StartElement); ok {
			root = strings.ToLower(se.Name.Local)
			break
		}
	}
	clean := func(s string) string {
		s = html.UnescapeString(tagRE.ReplaceAllString(html.UnescapeString(s), ""))
		return strings.Join(strings.Fields(stripControl(s)), " ")
	}
	var items []FeedItem
	switch root {
	case "rss", "rdf":
		var doc rssDoc
		if err := newDecoder().Decode(&doc); err != nil {
			return nil, fmt.Errorf("bad RSS: %v", err)
		}
		for _, it := range append(doc.Channel.Items, doc.Items...) {
			link := strings.TrimSpace(it.Link)
			if link == "" && strings.HasPrefix(strings.TrimSpace(it.GUID), "http") {
				link = strings.TrimSpace(it.GUID) // some feeds carry the link only as a permalink GUID
			}
			date := parseFeedDate(it.PubDate)
			if date.IsZero() {
				date = parseFeedDate(it.Date)
			}
			items = append(items, FeedItem{Feed: feed.Title, Topic: feed.Topic, Title: clean(it.Title), Link: link,
				Date: date, Summary: plainSummary(it.Description, 400)})
		}
	case "feed":
		var doc atomDoc
		if err := newDecoder().Decode(&doc); err != nil {
			return nil, fmt.Errorf("bad Atom: %v", err)
		}
		for _, e := range doc.Entries {
			link := ""
			for _, l := range e.Links {
				if l.Rel == "" || l.Rel == "alternate" {
					link = l.Href
					break
				}
			}
			if link == "" && len(e.Links) > 0 {
				link = e.Links[0].Href
			}
			date := parseFeedDate(e.Published)
			if date.IsZero() {
				date = parseFeedDate(e.Updated)
			}
			summary := e.Summary
			if summary == "" {
				summary = e.Content
			}
			items = append(items, FeedItem{Feed: feed.Title, Topic: feed.Topic, Title: clean(e.Title), Link: strings.TrimSpace(link),
				Date: date, Summary: plainSummary(summary, 400)})
		}
	default:
		return nil, fmt.Errorf("not an RSS or Atom feed (root element <%s>)", root)
	}
	var out []FeedItem
	for _, it := range items {
		if it.Title == "" || !utf8.ValidString(it.Title) {
			continue
		}
		if u, err := url.Parse(it.Link); it.Link != "" && (err != nil || (u.Scheme != "https" && u.Scheme != "http")) {
			it.Link = "" // never pass odd schemes (javascript:, file:) to a browser
		}
		out = append(out, it)
	}
	return out, nil
}

const maxFeedSize = 5 << 20

// fetchFeed downloads and parses one feed.
func fetchFeed(ctx context.Context, client *http.Client, feed Feed) ([]FeedItem, error) {
	u, err := url.Parse(feed.URL)
	if err != nil || u.Scheme != "https" {
		return nil, errNotHTTPS
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "academy-techwatch/1.0 (personal feed reader)")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml;q=0.9, */*;q=0.5")
	resp, err := client.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			return nil, fmt.Errorf("could not reach %s: %v", u.Host, ue.Err)
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, u.Host)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxFeedSize {
		return nil, errors.New("feed larger than 5 MB")
	}
	return parseFeed(data, feed)
}

// FeedResult is the outcome of fetching one feed.
type FeedResult struct {
	Feed  Feed
	Items []FeedItem
	Err   error
}

// fetchFeeds fetches feeds concurrently (six at a time) with a per-feed
// timeout, keeping the order of the input.
func fetchFeeds(client *http.Client, feeds []Feed, timeout time.Duration) []FeedResult {
	results := make([]FeedResult, len(feeds))
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup
	for i, f := range feeds {
		wg.Add(1)
		go func(i int, f Feed) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			items, err := fetchFeed(ctx, client, f)
			results[i] = FeedResult{Feed: f, Items: items, Err: err}
		}(i, f)
	}
	wg.Wait()
	return results
}

// headlines merges results, newest first, keeping at most perFeed items
// from each feed so one busy source cannot drown the rest.
func headlines(results []FeedResult, perFeed int) []FeedItem {
	var all []FeedItem
	for _, r := range results {
		items := append([]FeedItem{}, r.Items...)
		sort.SliceStable(items, func(i, j int) bool { return items[i].Date.After(items[j].Date) })
		if len(items) > perFeed {
			items = items[:perFeed]
		}
		all = append(all, items...)
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].Date.After(all[j].Date) })
	return all
}
