package main

// Terminal presentation helpers: ANSI styling, progress bars, text reflow
// and sparklines. Everything degrades to plain text when colour is off
// (NO_COLOR set, TERM=dumb, output not a terminal, or -no-color).

import (
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Style emits ANSI SGR sequences only when enabled.
type Style struct{ on bool }

// sty is the process-wide style, configured once in main.
var sty Style

func (s Style) paint(code, text string) string {
	if !s.on || text == "" {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func (s Style) Bold(t string) string    { return s.paint("1", t) }
func (s Style) Dim(t string) string     { return s.paint("2", t) }
func (s Style) Italic(t string) string  { return s.paint("3", t) }
func (s Style) Under(t string) string   { return s.paint("4", t) }
func (s Style) Red(t string) string     { return s.paint("31", t) }
func (s Style) Green(t string) string   { return s.paint("32", t) }
func (s Style) Yellow(t string) string  { return s.paint("33", t) }
func (s Style) Blue(t string) string    { return s.paint("34", t) }
func (s Style) Magenta(t string) string { return s.paint("35", t) }
func (s Style) Cyan(t string) string    { return s.paint("36", t) }
func (s Style) Gray(t string) string    { return s.paint("90", t) }

// colorEnabled decides whether stdout should receive ANSI colour.
func colorEnabled(forceOff bool) bool {
	if forceOff || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// trackColor gives each track a consistent accent colour.
func trackColor(id string) func(string) string {
	switch id {
	case "F":
		return sty.Green
	case "A":
		return sty.Cyan
	case "S":
		return sty.Yellow
	case "B":
		return sty.Magenta
	default:
		return sty.Blue
	}
}

// Layout width limits: below minWidth the terminal wraps our lines; above
// maxWidth text gets hard to read, so we stop growing.
const (
	minWidth = 40
	maxWidth = 110
)

// termWidth returns the width to lay content out in: the terminal's live
// width (so a resize takes effect on the next screen), else $COLUMNS,
// else 80, clamped to [minWidth, maxWidth].
func termWidth() int {
	w, _, ok := terminalSize(int(os.Stdout.Fd()))
	if !ok {
		if env, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && env > 0 {
			w = env
		} else {
			w = 80
		}
	}
	return max(minWidth, min(w-1, maxWidth)) // -1: never touch the last column, which makes some terminals wrap
}

// stripANSI removes SGR escape sequences so visible width can be measured.
func stripANSI(s string) string {
	if !strings.Contains(s, "\x1b[") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7e) {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// runeWidth is how many terminal columns a rune occupies: 2 for wide
// characters (CJK and most emoji), 0 for combining marks, else 1.
func runeWidth(r rune) int {
	switch {
	case unicode.Is(unicode.Mn, r) || r == 0x200d || (r >= 0xfe00 && r <= 0xfe0f):
		return 0
	case r >= 0x1100 && r <= 0x115f, r >= 0x2e80 && r <= 0xa4cf, r >= 0xac00 && r <= 0xd7a3,
		r >= 0xf900 && r <= 0xfaff, r >= 0xfe30 && r <= 0xfe4f, r >= 0xff00 && r <= 0xff60,
		r >= 0xffe0 && r <= 0xffe6, r >= 0x1f300 && r <= 0x1f64f, r >= 0x1f680 && r <= 0x1f6ff,
		r >= 0x1f900 && r <= 0x1faff, r >= 0x20000 && r <= 0x3fffd:
		return 2
	}
	return 1
}

// visibleLen is the number of terminal columns a string occupies.
func visibleLen(s string) int {
	n := 0
	for _, r := range stripANSI(s) {
		n += runeWidth(r)
	}
	return n
}

// padRight pads s with spaces to n visible columns.
func padRight(s string, n int) string {
	if gap := n - visibleLen(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

// truncate shortens plain text to n columns, ending with an ellipsis.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if visibleLen(s) <= n {
		return s
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		if w+runeWidth(r) > n-1 {
			break
		}
		b.WriteRune(r)
		w += runeWidth(r)
	}
	return b.String() + "…"
}

// flow lays items out left to right, separated by sep, starting a new
// line (with indent) whenever the next item would pass width.
func flow(items []string, sep string, width int, indent string) []string {
	var lines []string
	line := ""
	for _, it := range items {
		if it == "" {
			continue
		}
		switch {
		case line == "":
			line = indent + it
		case visibleLen(line)+visibleLen(sep)+visibleLen(it) <= width:
			line += sep + it
		default:
			lines = append(lines, line)
			line = indent + it
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// hardBreak splits a word that is wider than width into pieces.
func hardBreak(word string, width int) []string {
	if width < 1 || visibleLen(word) <= width {
		return []string{word}
	}
	var parts []string
	var b strings.Builder
	w := 0
	for _, r := range word {
		if w+runeWidth(r) > width {
			parts = append(parts, b.String())
			b.Reset()
			w = 0
		}
		b.WriteRune(r)
		w += runeWidth(r)
	}
	return append(parts, b.String())
}

// bar renders a width-cell progress bar for done/total.
func bar(done, total, width int, color func(string) string) string {
	if total <= 0 {
		return sty.Gray(strings.Repeat("─", width))
	}
	filled := int(math.Round(float64(done) / float64(total) * float64(width)))
	filled = max(0, min(filled, width))
	return color(strings.Repeat("█", filled)) + sty.Gray(strings.Repeat("░", width-filled))
}

// pct formats done/total as a whole percentage.
func pct(done, total int) string {
	if total == 0 {
		return "0%"
	}
	return strconv.Itoa(int(math.Round(100*float64(done)/float64(total)))) + "%"
}

// sparkline maps values onto eighth-block glyphs; zero days show as a dot.
func sparkline(vals []float64) string {
	const ticks = "▁▂▃▄▅▆▇█"
	glyphs := []rune(ticks)
	peak := 0.0
	for _, v := range vals {
		peak = math.Max(peak, v)
	}
	var b strings.Builder
	for _, v := range vals {
		if v <= 0 || peak == 0 {
			b.WriteString(sty.Gray("·"))
			continue
		}
		idx := int(math.Ceil(v/peak*float64(len(glyphs)))) - 1
		b.WriteString(sty.Green(string(glyphs[max(0, min(idx, len(glyphs)-1))])))
	}
	return b.String()
}

// preMark starts a logical line that must be shown verbatim (code).
const preMark = "\x00"

// reflow turns authored text into logical lines. Blank lines separate
// paragraphs, lines starting with "- " are bullets, ALL-CAPS lines are
// headings, lines indented by two or more spaces are code (kept exactly
// as written), and other newlines are soft (joined with a space).
func reflow(text string) []string {
	var out []string
	cur := ""
	flush := func() {
		if cur != "" {
			out = append(out, cur)
			cur = ""
		}
	}
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case line != "" && strings.HasPrefix(raw, "  "):
			flush()
			out = append(out, preMark+strings.TrimRight(raw[2:], " "))
		case line == "":
			flush()
			out = append(out, "")
		case strings.HasPrefix(line, "- "):
			flush()
			cur = line
		case isHeadingLine(line):
			flush()
			out = append(out, line)
		case cur == "":
			cur = line
		default:
			cur += " " + line
		}
	}
	flush()
	return out
}

// isHeadingLine reports whether a line is an ALL-CAPS section heading.
func isHeadingLine(line string) bool {
	return line == strings.ToUpper(line) && strings.ContainsAny(line, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
		!strings.ContainsAny(line, "abcdefghijklmnopqrstuvwxyz")
}

// wrap word-wraps text to width columns, prefixing each line with indent.
// Bullets get a hanging indent so continuation lines align with the text.
func wrap(text string, width int, indent string) []string {
	var lines []string
	for _, logical := range reflow(text) {
		if logical == "" {
			lines = append(lines, "")
			continue
		}
		if code, ok := strings.CutPrefix(logical, preMark); ok {
			for _, part := range hardBreak(code, width-visibleLen(indent)-2) {
				lines = append(lines, indent+"  "+part)
			}
			continue
		}
		first, rest := indent, indent
		if strings.HasPrefix(logical, "- ") {
			first = indent + "• "
			rest = indent + "  "
			logical = strings.TrimPrefix(logical, "- ")
		}
		line, prefix := "", first
		for _, long := range strings.Fields(logical) {
			for _, word := range hardBreak(long, width-visibleLen(rest)) {
				if line != "" && visibleLen(prefix)+visibleLen(line)+1+visibleLen(word) > width {
					lines = append(lines, prefix+line)
					line, prefix = "", rest
				}
				if line == "" {
					line = word
				} else {
					line += " " + word
				}
			}
		}
		if line != "" {
			lines = append(lines, prefix+line)
		}
	}
	return lines
}

// dedent strips the common leading whitespace from a block (for diagrams).
func dedent(block string) []string {
	lines := strings.Split(strings.Trim(block, "\n"), "\n")
	common := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		lead := len(l) - len(strings.TrimLeft(l, " \t"))
		if common < 0 || lead < common {
			common = lead
		}
	}
	for i, l := range lines {
		if len(l) >= common && common > 0 {
			lines[i] = l[common:]
		}
	}
	return lines
}

// banner is the startup logo; the subtitle wraps on narrow terminals.
func banner(width int) string {
	art := []string{
		"   ▄▀█ █▀▀ ▄▀█ █▀▄ █▀▀ █▀▄▀█ █▄█",
		"   █▀█ █▄▄ █▀█ █▄▀ ██▄ █░▀░█ ░█░",
	}
	sub := ""
	for _, l := range wrap("Systems & AI · Campus Registry · "+appVersion(), width-3, "   ") {
		sub += sty.Dim(l) + "\n"
	}
	return "\n" + sty.Bold(sty.Cyan(art[0])) + "\n" + sty.Bold(sty.Magenta(art[1])) + "\n" + sub
}

// heading renders a double-ruled section title in the given colour.
func heading(title string, color func(string) string, width int) string {
	rule := strings.Repeat("═", width)
	out := "\n" + color(rule)
	for _, l := range wrap(title, width, "  ") {
		out += "\n" + color(sty.Bold(l))
	}
	return out + "\n" + color(rule)
}
