package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// The Windows icon resources must stay valid COFF objects with a .rsrc
// section for the right machine, or academy.exe loses its icon (or the
// Windows build fails). Regenerate them with tools/icons.py.
func TestWindowsIconResources(t *testing.T) {
	for file, machine := range map[string]uint16{
		"rsrc_windows_amd64.syso": 0x8664,
		"rsrc_windows_arm64.syso": 0xAA64,
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("%s: %v (run python3 tools/icons.py)", file, err)
		}
		if len(data) < 60 {
			t.Fatalf("%s is truncated", file)
		}
		if got := binary.LittleEndian.Uint16(data[0:2]); got != machine {
			t.Errorf("%s: machine %#x, want %#x", file, got, machine)
		}
		if n := binary.LittleEndian.Uint16(data[2:4]); n != 1 {
			t.Errorf("%s: %d sections, want 1", file, n)
		}
		if name := bytes.TrimRight(data[20:28], "\x00"); string(name) != ".rsrc" {
			t.Errorf("%s: section %q, want .rsrc", file, name)
		}
		if !bytes.Contains(data, []byte("\x89PNG\r\n\x1a\n")) {
			t.Errorf("%s holds no PNG icon images", file)
		}
	}
	ico, err := os.ReadFile("assets/academy.ico")
	if err != nil {
		t.Fatal(err)
	}
	if len(ico) < 6 || binary.LittleEndian.Uint16(ico[2:4]) != 1 || binary.LittleEndian.Uint16(ico[4:6]) != 7 {
		t.Fatalf("academy.ico should be an icon file with 7 sizes")
	}
}
