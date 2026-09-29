package main

import (
	"os"
	"testing"
)

// TestMain keeps every test's Library downloads in a temporary folder, so
// no test ever touches the real Documents folder.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "academy-test-library-")
	if err != nil {
		panic(err)
	}
	os.Setenv("ACADEMY_LIBRARY", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
