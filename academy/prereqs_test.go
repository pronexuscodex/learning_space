package main

import (
	"strings"
	"testing"
	"time"
)

// stageClosure is every stage that stage id builds on, directly or not.
func stageClosure(id int) map[int]bool {
	seen := map[int]bool{}
	var walk func(int)
	walk = func(s int) {
		for _, p := range stagePrereqs[s] {
			if !seen[p] {
				seen[p] = true
				walk(p)
			}
		}
	}
	walk(id)
	return seen
}

// TestConceptPrereqs checks the prerequisite graph: every name is a
// concept, there are no duplicates, and every prerequisite comes earlier
// (same stage and earlier, or in a stage this one builds on). That last
// rule also rules out cycles: a cycle would need an edge that goes forward.
func TestConceptPrereqs(t *testing.T) {
	seen := map[string]int{}
	for id, g := range curriculum {
		for _, c := range g.Concepts {
			if prev, dup := seen[c.Name]; dup {
				t.Errorf("concept %q is in stage %d and stage %d: names must be unique", c.Name, prev, id)
			}
			seen[c.Name] = id
		}
	}
	for name, reqs := range conceptPrereqs {
		id, idx, ok := conceptAt(name)
		if !ok {
			t.Errorf("conceptPrereqs has %q, which is not a concept", name)
			continue
		}
		closure := stageClosure(id)
		dup := map[string]bool{}
		for _, p := range reqs {
			if dup[p] {
				t.Errorf("%q lists %q twice", name, p)
			}
			dup[p] = true
			pid, pidx, ok := conceptAt(p)
			switch {
			case !ok:
				t.Errorf("%q builds on %q, which is not a concept", name, p)
			case p == name:
				t.Errorf("%q builds on itself", name)
			case pid == id && pidx >= idx:
				t.Errorf("%q (Stage %d, #%d) builds on %q, which comes later in the same stage (#%d)", name, id, idx+1, p, pidx+1)
			case pid != id && !closure[pid]:
				t.Errorf("%q (Stage %d) builds on %q (Stage %d), a stage that Stage %d does not build on: fix the edge or stagePrereqs", name, id, p, pid, id)
			}
		}
	}
}

// Every concept after the first in Stage 2 onwards should say what it
// builds on; a concept with no prerequisites there is probably an omission.
func TestConceptPrereqsCoverage(t *testing.T) {
	for id, g := range curriculum {
		if id < 1 {
			continue
		}
		for _, c := range g.Concepts {
			if len(conceptPrereqs[c.Name]) == 0 && c.Name != "Thinking Like an Attacker: Threat Modelling" {
				t.Errorf("Stage %d: %q has no prerequisites listed", id, c.Name)
			}
		}
	}
}

func TestNextToLearnStartsAtTheBeginning(t *testing.T) {
	r := seedRegistry()
	id, idx, unblocks, ok := r.nextToLearn()
	if !ok || id != 0 || idx != 0 || unblocks != "" {
		t.Fatalf("fresh registry: got stage %d concept %d unblocks %q ok %v, want stage 0 concept 0", id, idx, unblocks, ok)
	}
}

// A learner who jumped to Stage 6 is sent back to the deepest missing
// prerequisite of the next concept, and told which concept it unblocks.
func TestNextToLearnDetoursToMissingPrereq(t *testing.T) {
	r := seedRegistry()
	_, s := r.findStage(6)
	s.setStudied("Processes & the Address Space", true)

	id, idx, unblocks, ok := r.nextToLearn()
	if !ok {
		t.Fatal("no next concept")
	}
	got := curriculum[id].Concepts[idx].Name
	// Virtual Memory needs Pointers (Stage 1), which needs Memory & Storage (Stage 0).
	if got != "Memory & Storage: Where Data Lives" || unblocks != "Virtual Memory & Paging" {
		t.Fatalf("got %q (Stage %d) unblocking %q", got, id, unblocks)
	}
	if sid, sidx, ok := r.nextConcept(); !ok || sid != id || sidx != idx {
		t.Errorf("nextConcept = %d/%d, want the same detour %d/%d", sid, sidx, id, idx)
	}

	var why string
	for _, sg := range r.nextSteps(time.Now()) {
		if sg.Kind == SuggestConcept {
			why = sg.Title + " · " + sg.Why
		}
	}
	if !strings.Contains(why, "Memory & Storage") || !strings.Contains(why, "Virtual Memory & Paging") {
		t.Errorf("suggestion does not explain the detour: %q", why)
	}

	// With that chain done, the next blocker is the concept's other
	// prerequisite, The Memory Hierarchy & Locality, followed down to its
	// deepest unmet root (Stage 0's first concept).
	_, s0 := r.findStage(0)
	s0.setStudied("Memory & Storage: Where Data Lives", true)
	_, s1 := r.findStage(1)
	s1.setStudied("Memory, Addresses & Pointers", true)
	for _, p := range r.prereqsOf("Virtual Memory & Paging") {
		if want := p.Name != "The Memory Hierarchy & Locality"; p.Met != want {
			t.Errorf("%s: met = %v, want %v", p.Name, p.Met, want)
		}
	}
	if p := r.firstUnmet("Virtual Memory & Paging"); p != "Bits & Bytes: Everything Is a Number" {
		t.Errorf("deepest blocker = %q, want Bits & Bytes (via the memory hierarchy chain)", p)
	}
	// The earliest stage begun is the current stage (currentStage), so the
	// learner now carries on through Stage 0 in order.
	id, idx, unblocks, _ = r.nextToLearn()
	if got := curriculum[id].Concepts[idx].Name; id != 0 || got != "Bits & Bytes: Everything Is a Number" || unblocks != "" {
		t.Errorf("after the detour: got %q (Stage %d) unblocking %q", got, id, unblocks)
	}
}

// When the next concept of a stage that is already under way has all its
// prerequisites, it is suggested as is.
func TestNextToLearnNoDetourWhenMet(t *testing.T) {
	r := seedRegistry()
	_, s0 := r.findStage(0)
	for _, c := range curriculum[0].Concepts {
		s0.setStudied(c.Name, true)
	}
	_, s1 := r.findStage(1)
	s1.setStudied("Your First C Program: Source, Compiler, Executable", true)
	id, idx, unblocks, ok := r.nextToLearn()
	if got := curriculum[id].Concepts[idx].Name; !ok || id != 1 || got != "Values, Types & Variables" || unblocks != "" {
		t.Errorf("got %q (Stage %d) unblocking %q", got, id, unblocks)
	}
}

func TestPrereqsOfReportsState(t *testing.T) {
	r := seedRegistry()
	_, s := r.findStage(1)
	s.setStudied("Functions & Decomposition", true)
	ps := r.prereqsOf("Memory, Addresses & Pointers")
	if len(ps) != 2 {
		t.Fatalf("got %d prerequisites, want 2", len(ps))
	}
	for _, p := range ps {
		want := p.Name == "Functions & Decomposition"
		if p.Met != want {
			t.Errorf("%s: met = %v, want %v", p.Name, p.Met, want)
		}
	}
}
