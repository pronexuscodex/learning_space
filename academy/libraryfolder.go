package main

// Where the Library keeps downloaded PDFs, and how it arranges them.
//
// PDFs are the learner's own files, so they live where people look for
// documents, not next to the registry (which a package manager may put in
// a hidden folder):
//
//	Documents/Academy Library/
//	    Stage 00 - How Computers Work/
//	        Computer Science from the Bottom Up.pdf
//	    Stage 04 - Digital Logic & Computer Architecture/
//	        First Draft of a Report on the EDVAC.pdf
//	    My PDFs/                 (PDFs saved from your own links)
//
// $ACADEMY_LIBRARY chooses another folder; with $ACADEMY_HOME set, the
// library sits inside it, next to the registry, as -h promises.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

const (
	libraryFolderName = "Academy Library"
	myPDFsFolder      = "My PDFs"
	legacyLibraryName = "academy_library" // before 1.3: a flat folder beside the registry
)

// libraryDir is the Library folder for the registry at regPath.
func libraryDir(regPath string) string {
	return libraryDirFor(regPath, os.Getenv, documentsDir)
}

func libraryDirFor(regPath string, getenv func(string) string, docs func() (string, error)) string {
	if d := getenv("ACADEMY_LIBRARY"); d != "" {
		return d
	}
	if home := getenv("ACADEMY_HOME"); home != "" {
		return filepath.Join(home, libraryFolderName)
	}
	if d, err := docs(); err == nil {
		return filepath.Join(d, libraryFolderName)
	}
	return filepath.Join(filepath.Dir(regPath), libraryFolderName)
}

// documentsDir is the user's Documents folder: the XDG setting on Linux
// and BSD when there is one, and ~/Documents everywhere else.
func documentsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		if d := xdgDocuments(home); d != "" {
			return d, nil
		}
	}
	return filepath.Join(home, "Documents"), nil
}

// xdgDocuments reads XDG_DOCUMENTS_DIR from ~/.config/user-dirs.dirs,
// where Linux desktops record the (possibly translated) Documents folder.
func xdgDocuments(home string) string {
	cfg := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(cfg) {
		cfg = filepath.Join(home, ".config")
	}
	f, err := os.Open(filepath.Join(cfg, "user-dirs.dirs"))
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		v, ok := strings.CutPrefix(strings.TrimSpace(sc.Text()), "XDG_DOCUMENTS_DIR=")
		if !ok {
			continue
		}
		v = strings.Trim(v, `"`)
		v = strings.Replace(v, "$HOME", home, 1)
		if filepath.IsAbs(v) && filepath.Clean(v) != filepath.Clean(home) {
			return filepath.Clean(v)
		}
	}
	return ""
}

var (
	stageTitlesOnce sync.Once
	stageTitles     map[int]string
)

// stageFolder is the sub-folder for a stage's PDFs.
func stageFolder(stage int) string {
	if stage == NoStage {
		return myPDFsFolder
	}
	stageTitlesOnce.Do(func() {
		stageTitles = map[int]string{}
		for _, t := range seedRegistry().Tracks {
			for _, s := range t.Stages {
				stageTitles[s.ID] = s.Title
			}
		}
	})
	name := fmt.Sprintf("Stage %02d", stage)
	if title := cleanName(stageTitles[stage], 70); title != "" {
		name += " - " + title
	}
	return name
}

// cleanName turns a title into a file or folder name that is valid on
// Windows, macOS and Linux: no path separators, no characters Windows
// forbids, no control codes, no trailing dots or spaces, not a reserved
// device name, and at most max characters.
func cleanName(s string, max int) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == ':':
			b.WriteString(" -")
		case strings.ContainsRune(`/\<>"|?*`, r), unsafeRune(r):
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	name := strings.Join(strings.Fields(b.String()), " ")
	if r := []rune(name); len(r) > max {
		name = strings.TrimSpace(string(r[:max]))
	}
	name = strings.TrimRight(name, ". ")
	switch strings.ToUpper(strings.SplitN(name, ".", 2)[0]) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "LPT1", "LPT2", "LPT3":
		name = "_" + name
	}
	return name
}

// relPath is where a document lives inside the Library folder.
func (d LibraryDoc) relPath() string {
	if d.File != "" {
		return d.File
	}
	name := cleanName(d.Title, 100)
	if name == "" {
		name = "Document"
	}
	return filepath.Join(stageFolder(d.Stage), name+".pdf")
}

// walkPDFs lists every PDF under dir (relative paths, sorted).
func walkPDFs(dir string) []string {
	var out []string
	filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !e.IsDir() && strings.EqualFold(filepath.Ext(p), ".pdf") {
			if rel, err := filepath.Rel(dir, p); err == nil {
				out = append(out, rel)
			}
		}
		return nil
	})
	sort.Strings(out)
	return out
}

// migrateLegacyLibrary moves PDFs from the old flat folder beside the
// registry into the arranged Library, and reports how many it moved.
// Files that already exist at the destination are left where they were.
func migrateLegacyLibrary(regPath, dir string) (int, error) {
	old := filepath.Join(filepath.Dir(regPath), legacyLibraryName)
	if filepath.Clean(old) == filepath.Clean(dir) {
		return 0, nil
	}
	entries, err := os.ReadDir(old)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	byLegacyName := map[string]LibraryDoc{}
	for _, d := range libraryDocs() {
		byLegacyName[d.legacyFileName()] = d
	}
	moved := 0
	for _, e := range entries {
		name := e.Name()
		src := filepath.Join(old, name)
		if strings.HasPrefix(name, ".academy-download-") {
			os.Remove(src) // an interrupted download
			continue
		}
		if !e.Type().IsRegular() || !strings.EqualFold(filepath.Ext(name), ".pdf") {
			continue
		}
		rel := filepath.Join(myPDFsFolder, name)
		if d, ok := byLegacyName[name]; ok {
			rel = d.relPath()
		}
		dest := filepath.Join(dir, rel)
		if _, err := os.Stat(dest); err == nil {
			continue
		}
		if err := moveFile(src, dest); err != nil {
			return moved, err
		}
		moved++
	}
	os.Remove(old) // only succeeds once it is empty
	return moved, nil
}

// moveFile renames src to dest, copying when they are on different drives.
func moveFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(dest), partPattern)
	if err != nil {
		return err
	}
	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), dest); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	in.Close()
	return os.Remove(src)
}
