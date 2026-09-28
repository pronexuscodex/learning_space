package main

import (
	"bytes"
	"flag"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionAndUsage(t *testing.T) {
	old := version
	defer func() { version = old }()
	version = "v9.8.7"
	if appVersion() != "v9.8.7" || !strings.HasPrefix(versionLine(), "academy v9.8.7 (go") {
		t.Fatalf("version = %q, line = %q", appVersion(), versionLine())
	}
	version = ""
	if v := appVersion(); v == "" || strings.Contains(v, "+dirty") || strings.HasPrefix(v, "v0.0.0-") {
		t.Fatalf("fallback version = %q", v)
	}
	var b bytes.Buffer
	fs := flag.NewFlagSet("academy", flag.ContinueOnError)
	fs.Bool("version", false, "print the version, then exit")
	usage(&b, fs)()
	for _, want := range []string{"Usage:", "-version", "Examples:", "academy -restore 1"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("usage lacks %q", want)
		}
	}
}

func TestWhereText(t *testing.T) {
	reg := filepath.FromSlash("/data/academy/" + registryFileName)
	got := whereText(filepath.FromSlash("/bin/academy"), reg)
	for _, want := range []string{
		"Program:  " + filepath.FromSlash("/bin/academy") + "\n",
		"Progress: " + reg + "\n",
		"Backups:  " + backupDir(reg) + "\n",
		"PDFs:     " + libraryDir(reg) + "\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("whereText missing %q in:\n%s", want, got)
		}
	}
}
