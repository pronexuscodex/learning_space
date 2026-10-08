package main

// Settings: colour themes, plain symbols, and the practice switches, in
// one screen (o on the main menu).

import (
	"io"
	"strings"
	"unicode/utf8"
)

// plainSymbols are the replacements used when emoji are switched off, for
// terminals and fonts that draw them badly. Each keeps the emoji's width,
// so boxes and columns stay aligned.
var plainSymbols = map[rune]string{
	'💡': "!", '📝': "=", '📓': "=", '📕': "=", '📖': "=", '📚': "=", '📜': "=",
	'🔧': "#", '🧱': "#", '🔒': "x", '🔗': "&", '🔎': "?", '🔍': "?", '🧭': ">",
	'🏆': "*", '📁': "/", '💬': "\"", '🌍': "@", '🏋': "+", '😀': ":)", '🔤': "Aa",
	'⏱': "~", '⏸': "|", '⌨': "#", '⏎': "<", '⬇': "v",
}

// plainText replaces emoji with plain characters of the same width.
func plainText(s string) string {
	if !strings.ContainsFunc(s, isEmoji) {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == 0xfe0f: // emoji presentation selector: drop it
		case isEmoji(r):
			rep, ok := plainSymbols[r]
			if !ok {
				rep = "*"
			}
			b.WriteString(rep)
			if pad := runeWidth(r) - utf8.RuneCountInString(rep); pad > 0 {
				b.WriteString(strings.Repeat(" ", pad))
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isEmoji(r rune) bool {
	if r >= 0x1f000 && r <= 0x1faff {
		return true
	}
	_, ok := plainSymbols[r]
	return ok
}

// plainWriter applies plainText to everything written through it.
type plainWriter struct{ w io.Writer }

func (p plainWriter) Write(b []byte) (int, error) {
	if _, err := io.WriteString(p.w, plainText(string(b))); err != nil {
		return 0, err
	}
	return len(b), nil
}

// applySettings puts the display settings into effect.
func (a *App) applySettings() {
	a.mu.Lock()
	st := a.reg.Settings
	a.mu.Unlock()
	sty.theme = st.Theme
	base := a.con.out
	if pw, ok := base.(plainWriter); ok {
		base = pw.w
	}
	if st.PlainSymbols {
		a.con.out = plainWriter{base}
	} else {
		a.con.out = base
	}
}

// settingsScreen lets the learner change display and practice settings.
func (a *App) settingsScreen() error {
	for {
		a.mu.Lock()
		st := a.reg.Settings
		classic, tidy := a.reg.ClassicMode, a.reg.TidyScreen
		a.mu.Unlock()
		onOff := func(b bool) string {
			if b {
				return sty.Green("on")
			}
			return sty.Gray("off")
		}
		themeName := themes[0].Name
		for _, t := range themes {
			if t.ID == st.Theme {
				themeName = t.Name
			}
		}
		homeName := "simple"
		if st.DetailedHome {
			homeName = "detailed"
		}
		symbols := "emoji 💡"
		if st.PlainSymbols {
			symbols = "plain characters"
		}
		a.println(heading("SETTINGS", sty.Cyan, a.cols()))
		choice, err := a.con.promptChoice("Change which?", []string{
			"Colour theme: " + sty.Bold(themeName),
			"Symbols: " + sty.Bold(symbols) + sty.Gray(" · plain if emoji look broken"),
			"Ask how sure I am before review answers: " + onOff(!st.NoConfidence),
			"Classic Mode: " + onOff(classic) + sty.Gray(" · struggle clock, notebook, type-ins"),
			"Tidy screen: " + onOff(tidy) + sty.Gray(" · each action on a clean screen"),
			"Home screen: " + sty.Bold(homeName) + sty.Gray(" · simple, or every number and menu item"),
			"Back",
		})
		if err != nil || choice == 6 {
			if err == errCancel {
				return nil
			}
			return err
		}
		switch choice {
		case 0:
			if err := a.chooseTheme(); err != nil && err != errCancel {
				return err
			}
		case 1:
			a.mutate(func(r *Registry) { r.Settings.PlainSymbols = !r.Settings.PlainSymbols })
			a.applySettings()
			a.con.ok("Symbols changed.")
		case 2:
			a.mutate(func(r *Registry) { r.Settings.NoConfidence = !r.Settings.NoConfidence })
			a.con.ok("Saved.")
		case 3:
			if err := a.toggleClassicMode(); err != nil && err != errCancel {
				return err
			}
		case 4:
			a.mutate(func(r *Registry) { r.TidyScreen = !r.TidyScreen })
			a.con.ok("Saved.")
		case 5:
			a.mutate(func(r *Registry) { r.Settings.DetailedHome = !r.Settings.DetailedHome })
			a.con.ok("Home screen changed. Every key works either way; ? lists them all.")
		}
	}
}

// chooseTheme shows every theme with a colour sample and applies the pick.
func (a *App) chooseTheme() error {
	opts := make([]string, len(themes))
	for i, t := range themes {
		p := Style{on: sty.on, theme: t.ID}
		sample := p.Green("● done") + " " + p.Yellow("● due") + " " + p.Red("● missed") + " " + p.Cyan("● info") + " " + p.Gray("hint")
		opts[i] = p.Bold(t.Name) + p.Gray(" · "+t.About) + "\n" + sample
	}
	if !sty.on {
		a.con.note("Colours are off in this terminal (NO_COLOR, -no-color or not a terminal), so themes have no visible effect here.")
	}
	choice, err := a.con.promptChoice("Theme", opts)
	if err != nil {
		return err
	}
	id := themes[choice].ID
	a.mutate(func(r *Registry) { r.Settings.Theme = id })
	a.applySettings()
	a.con.ok("Theme set to %s.", themes[choice].Name)
	return nil
}
