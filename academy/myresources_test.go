package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAddResourceValidatesAndDedupes(t *testing.T) {
	reg := seedRegistry()
	now := time.Now()
	good := MyResource{Stage: 5, Kind: "Book", Title: "  Modern C  ", URL: "https://example.org/modernc.pdf", Note: "Free C book"}
	id, err := reg.addResource(good, now)
	if err != nil || id != 1 || reg.MyResources[0].Title != "Modern C" {
		t.Fatalf("add: id=%d err=%v %+v", id, err, reg.MyResources)
	}
	if _, err := reg.addResource(MyResource{Stage: 1, Kind: "Site", Title: "Other title", URL: "https://example.org/modernc.pdf/"}, now); err == nil {
		t.Fatal("the same link must be rejected as a duplicate")
	}
	bad := []MyResource{
		{Kind: "Book"},
		{Kind: "Book", Title: strings.Repeat("x", maxNameLen+1)},
		{Kind: "Book", Title: "x", Stage: 17},
		{Kind: "Spaceship", Title: "x"},
		{Kind: "Book", Title: "x", URL: "ftp://example.org/x"},
		{Kind: "Book", Title: "x", URL: "javascript:alert(1)"},
		{Kind: "Book", Title: "evil\x1b[2J"},
	}
	for _, b := range bad {
		if _, err := reg.addResource(b, now); err == nil {
			t.Errorf("accepted invalid resource %+v", b)
		}
	}
	if id, _ := reg.addResource(MyResource{Kind: "Video", Title: "A talk"}, now); id != 2 {
		t.Fatalf("next id = %d", id)
	}
	if got := reg.myResourcesFor(5); len(got) != 1 || got[0].Title != "Modern C" {
		t.Fatalf("myResourcesFor(5) = %+v", got)
	}
	if got := reg.myResourcesFor(0); len(got) != 2 || got[0].Stage != 0 {
		t.Fatalf("myResourcesFor(0) should list all, general first: %+v", got)
	}
	if hits := reg.myResourceHits("modern book"); len(hits) != 1 || hits[0].Kind != HitMine || hits[0].Concept != 1 {
		t.Fatalf("hits = %+v", hits)
	}
	docs := reg.myLibraryDocs(0)
	if len(docs) != 1 || docs[0].PDF != "https://example.org/modernc.pdf" {
		t.Fatalf("library docs = %+v", docs)
	}
	if md := reg.exportMarkdown(now); !strings.Contains(md, "## My resources") || !strings.Contains(md, "[Modern C](https://example.org/modernc.pdf)") {
		t.Fatal("export should list my resources")
	}
}

func TestResourceFileRoundTrip(t *testing.T) {
	src := seedRegistry()
	now := time.Now()
	src.addResource(MyResource{Stage: 2, Kind: "Course", Title: "Algorithms course", URL: "https://example.org/algo"}, now)
	src.addResource(MyResource{Stage: 0, Kind: "Podcast", Title: "A podcast"}, now)
	data, err := exportResources(src.MyResources)
	if err != nil {
		t.Fatal(err)
	}
	var file resourceFile
	json.Unmarshal(data, &file)
	if file.Format != resourceFileFormat || len(file.Resources) != 2 || file.Resources[0].ID != 0 {
		t.Fatalf("export = %s", data)
	}

	dst := seedRegistry()
	dst.addResource(MyResource{Stage: 2, Kind: "Course", Title: "Mine already", URL: "https://example.org/algo"}, now)
	added, dupes, problems, err := dst.importResources(data, now)
	if err != nil || added != 1 || dupes != 1 || len(problems) != 0 {
		t.Fatalf("import: added=%d dupes=%d problems=%v err=%v", added, dupes, problems, err)
	}

	// A plain array works too, and invalid entries are reported, not added.
	plain := []byte(`[{"stage": 3, "kind": "Book", "title": "Proofs"}, {"stage": 99, "kind": "Book", "title": "Bad"}]`)
	added, _, problems, err = dst.importResources(plain, now)
	if err != nil || added != 1 || len(problems) != 1 || !strings.Contains(problems[0], "Bad") {
		t.Fatalf("plain import: added=%d problems=%v err=%v", added, problems, err)
	}
	if _, _, _, err := dst.importResources([]byte("not json"), now); err == nil {
		t.Fatal("garbage must be refused")
	}
}
