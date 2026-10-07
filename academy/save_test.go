package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Saving is automatic: every action is saved, one backup is made per
// session (not per save), and "undo this session" restores the file
// exactly as it was when the academy opened.
func TestAutosaveBackupOncePerSessionAndUndo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, registryFileName)
	if err := atomicWriteJSON(path, seedRegistry()); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)

	reg, _, err := loadRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{reg: reg, path: path, con: &console{in: bufio.NewReader(strings.NewReader("")), out: &bytes.Buffer{}}, width: 80}
	a.rememberOriginal()

	a.autosave() // nothing changed: nothing written
	if backups, _ := listBackups(path); len(backups) != 0 {
		t.Fatalf("an unchanged session must not make a backup, got %d", len(backups))
	}
	for i := 0; i < 5; i++ {
		a.mutate(func(r *Registry) { r.TidyScreen = !r.TidyScreen })
		a.autosave()
		saved, _, err := loadRegistry(path)
		if err != nil {
			t.Fatal(err)
		}
		if saved.TidyScreen != a.reg.TidyScreen {
			t.Fatalf("save %d: the change is not on disk", i)
		}
	}
	if backups, _ := listBackups(path); len(backups) != 1 {
		t.Fatalf("five saves in one session must make exactly one backup, got %d", len(backups))
	}

	if err := a.undoSession(); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("undo must restore the registry exactly as it was")
	}
}

// Undoing the very first session (no registry before) removes the file.
func TestUndoFirstSessionRemovesTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), registryFileName)
	a := &App{reg: seedRegistry(), path: path, con: &console{out: &bytes.Buffer{}}, width: 80}
	a.rememberOriginal()
	if err := a.commit(); err != nil {
		t.Fatal(err)
	}
	if err := a.undoSession(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("the registry created this session must be removed, stat err = %v", err)
	}
}
