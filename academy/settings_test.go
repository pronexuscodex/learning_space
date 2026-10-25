package main

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// The Settings screen as a learner uses it: wrong answers are refused
// with a reason and change nothing, toggles take effect, q leaves without
// an error, and the choices are still there after a restart.
func TestSettingsScreen(t *testing.T) {
	path := filepath.Join(t.TempDir(), registryFileName)
	if err := atomicWriteJSON(path, seedRegistry()); err != nil {
		t.Fatal(err)
	}
	reg, _, err := loadRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	keys := strings.Join([]string{
		"abc", // not a number
		"99",  // out of range
		"",    // empty
		"5",   // Home screen: everything → simple
		"6",   // Key bar: on → off
		"1",   // Colour theme…
		"q",   // …cancelled: the theme stays
		"q",   // leave Settings
	}, "\n") + "\n"
	a := &App{reg: reg, path: path, con: &console{in: bufio.NewReader(strings.NewReader(keys)), out: &out}, width: 80}
	a.rememberOriginal()
	saved := sty
	defer func() { sty = saved }()

	if err := a.settingsScreen(); err != nil {
		t.Fatalf("leaving Settings with q must not be an error: %v", err)
	}
	screen := stripANSI(out.String())
	for _, want := range []string{`"abc" is not a whole number`, "choose a number between 1 and 7", "Home screen changed", "Key bar: off"} {
		if !strings.Contains(screen, want) {
			t.Errorf("Settings screen lacks %q:\n%s", want, screen)
		}
	}
	if !a.reg.Settings.SimpleHome || !a.reg.Settings.HideKeyBar || !a.con.hideBar {
		t.Errorf("toggles not applied: %+v, bar hidden %v", a.reg.Settings, a.con.hideBar)
	}
	if a.reg.Settings.Theme != ThemeDefault {
		t.Errorf("cancelling the theme list changed the theme to %q", a.reg.Settings.Theme)
	}

	a.autosave()
	back, _, err := loadRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	if !back.Settings.SimpleHome || !back.Settings.HideKeyBar || back.Settings.Theme != ThemeDefault {
		t.Errorf("after a restart the settings are %+v", back.Settings)
	}
}
