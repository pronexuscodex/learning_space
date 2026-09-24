package main

// The learner's own resources: books, courses, videos, papers and sites
// they found themselves, stored in the registry and shown beside the
// curriculum's own. They can be shared as a JSON file.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// MyResource is one resource the learner added.
type MyResource struct {
	ID    int       `json:"id"`
	Stage int       `json:"stage"` // 0: general, not tied to a stage
	Kind  string    `json:"kind"`
	Title string    `json:"title"`
	URL   string    `json:"url,omitempty"`
	Note  string    `json:"note,omitempty"`
	Added time.Time `json:"added"`
}

// resourceKinds are the kinds a learner can choose from.
var resourceKinds = []string{"Book", "Course", "Video", "Article", "Paper", "Tool", "Site", "Podcast", "Newsletter"}

// validateResource checks a resource before it is saved or imported.
func validateResource(m MyResource) error {
	switch {
	case strings.TrimSpace(m.Title) == "":
		return errors.New("a title is required")
	case len([]rune(m.Title)) > maxNameLen:
		return fmt.Errorf("the title is longer than %d characters", maxNameLen)
	case len([]rune(m.Note)) > maxNotesLen:
		return fmt.Errorf("the note is longer than %d characters", maxNotesLen)
	case m.Stage < 0 || m.Stage > 16:
		return fmt.Errorf("stage %d does not exist (use 1–16, or 0 for general)", m.Stage)
	case !contains(resourceKinds, m.Kind):
		return fmt.Errorf("kind %q is not one of %s", m.Kind, strings.Join(resourceKinds, ", "))
	}
	if m.URL != "" {
		u, err := url.Parse(m.URL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return errors.New("the link must start with https:// (or http://)")
		}
	}
	for _, s := range []string{m.Title, m.URL, m.Note} {
		if strings.ContainsAny(s, "\x00\x1b") {
			return errors.New("control characters are not allowed")
		}
	}
	return nil
}

// sameResource reports whether two resources are duplicates: the same
// link, or the same title in the same stage when neither has a link.
func sameResource(a, b MyResource) bool {
	if a.URL != "" || b.URL != "" {
		return strings.EqualFold(strings.TrimRight(a.URL, "/"), strings.TrimRight(b.URL, "/"))
	}
	return a.Stage == b.Stage && strings.EqualFold(a.Title, b.Title)
}

// addResource validates and stores a resource, returning its new ID.
func (r *Registry) addResource(m MyResource, now time.Time) (int, error) {
	m.Title, m.URL, m.Note = strings.TrimSpace(m.Title), strings.TrimSpace(m.URL), strings.TrimSpace(m.Note)
	if err := validateResource(m); err != nil {
		return 0, err
	}
	for _, have := range r.MyResources {
		if sameResource(have, m) {
			return 0, fmt.Errorf("you already have this resource (%q)", have.Title)
		}
	}
	for _, have := range r.MyResources {
		m.ID = max(m.ID, have.ID)
	}
	m.ID++
	m.Added = now.UTC().Truncate(time.Second)
	r.MyResources = append(r.MyResources, m)
	return m.ID, nil
}

// myResourcesFor returns the learner's resources for a stage (0: all),
// sorted by stage then title.
func (r *Registry) myResourcesFor(stage int) []MyResource {
	var out []MyResource
	for _, m := range r.MyResources {
		if stage == 0 || m.Stage == stage {
			out = append(out, m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Stage != out[j].Stage {
			return out[i].Stage < out[j].Stage
		}
		return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title)
	})
	return out
}

// resourceFile is the shareable JSON format.
type resourceFile struct {
	Format    string       `json:"format"`
	Resources []MyResource `json:"resources"`
}

const resourceFileFormat = "academy-resources/1"

// exportResources renders resources as a shareable JSON document.
func exportResources(list []MyResource) ([]byte, error) {
	out := make([]MyResource, len(list))
	for i, m := range list {
		m.ID = 0 // IDs are local to one registry
		out[i] = m
	}
	data, err := json.MarshalIndent(resourceFile{Format: resourceFileFormat, Resources: out}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// importResources adds every valid, new resource from a shared file and
// reports how many were added, how many were duplicates, and the problems
// with the rest. A plain JSON array of resources is accepted too.
func (r *Registry) importResources(data []byte, now time.Time) (added, dupes int, problems []string, err error) {
	var file resourceFile
	if err := json.Unmarshal(data, &file); err != nil || file.Resources == nil {
		var list []MyResource
		if err2 := json.Unmarshal(data, &list); err2 != nil {
			return 0, 0, nil, fmt.Errorf("not a resource file: %v", err)
		}
		file.Resources = list
	}
	if len(file.Resources) > 1000 {
		return 0, 0, nil, errors.New("the file holds more than 1,000 resources")
	}
	for i, m := range file.Resources {
		m.ID = 0
		if _, err := r.addResource(m, now); err != nil {
			if strings.HasPrefix(err.Error(), "you already have") {
				dupes++
			} else {
				problems = append(problems, fmt.Sprintf("#%d %q: %v", i+1, m.Title, err))
			}
			continue
		}
		added++
	}
	return added, dupes, problems, nil
}

// myResourceHits finds the learner's resources matching every query word.
func (r *Registry) myResourceHits(query string) []Hit {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil
	}
	var hits []Hit
	for _, m := range r.MyResources {
		text := strings.ToLower(m.Title + " " + m.Note + " " + m.Kind + " " + m.URL)
		all := true
		for _, w := range words {
			all = all && strings.Contains(text, w)
		}
		if all {
			snippet := m.Note
			if snippet == "" {
				snippet = m.URL
			}
			hits = append(hits, Hit{Kind: HitMine, Stage: m.Stage, Title: m.Title, Snippet: snippet, Concept: m.ID, score: 4})
		}
	}
	return hits
}

// myLibraryDocs turns the learner's PDF links into library documents.
func (r *Registry) myLibraryDocs(stage int) []LibraryDoc {
	var docs []LibraryDoc
	for _, m := range r.myResourcesFor(stage) {
		if pdf, ok := pdfURL(m.URL); ok {
			docs = append(docs, LibraryDoc{Stage: m.Stage, Kind: m.Kind + " · yours", Title: m.Title, Page: m.URL, PDF: pdf})
		}
	}
	return docs
}
