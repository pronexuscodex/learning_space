package main

// Version reporting. Release builds stamp the version in with
//   go build -ldflags "-X main.version=v1.0.0"
// Other builds fall back to the Go module version or the Git revision
// recorded by the Go toolchain, and finally to "dev".

import (
	"flag"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strings"
)

var version = "" // set at release build time

// appVersion is the version shown to users.
func appVersion() string {
	if version != "" {
		return version
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	// A real tagged module version, but not a local pseudo-version such
	// as v0.0.0-20260924…-77876a787eed+dirty.
	if v := bi.Main.Version; v != "" && v != "(devel)" && !strings.HasPrefix(v, "v0.0.0-") && !strings.Contains(v, "+dirty") {
		return v
	}
	rev, dirty := "", false
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if len(rev) >= 7 {
		v := "dev-" + rev[:7]
		if dirty {
			v += "-modified"
		}
		return v
	}
	return "dev"
}

// versionLine is the full -version output.
func versionLine() string {
	return fmt.Sprintf("academy %s (%s, %s/%s)", appVersion(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// usage prints the -h help.
func usage(out io.Writer, fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(out, `Systems & AI Academy %s
A terminal study companion for self-learners: 17 stages from how computers
work and C to operating systems, security and AI, with spaced repetition,
exercises, a dictionary, a PDF library and tech watch. Your progress lives
in one JSON file.

Usage:
  academy [flags]

Flags:
`, appVersion())
		fs.SetOutput(out)
		fs.PrintDefaults()
		fmt.Fprint(out, `
Examples:
  academy                                  start (the registry sits beside the binary)
  academy -registry ~/study/academy.json   keep your progress somewhere else
  academy -backups                         list automatic backups
  academy -restore 1                       restore the newest backup
  academy -fetch-library                   download every free PDF for offline study
  ACADEMY_HOME=~/study academy             keep the registry, backups and PDFs in ~/study

Inside the app, press 0 for Start Here and ? for every key.
`)
	}
}
