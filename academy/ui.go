package main

// Terminal presentation helpers: ANSI styling, progress bars, text reflow
// and sparklines. Everything degrades to plain text when colour is off
// (NO_COLOR set, TERM=dumb, output not a terminal, or -no-color).

import (
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
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
	case "A":
		return sty.Cyan
	case "B":
		return sty.Magenta
	default:
		return sty.Blue
	}
}

// termWidth reads $COLUMNS, defaulting to 80 and clamping to a readable range.
func termWidth() int {
	w, err := strconv.Atoi(os.Getenv("COLUMNS"))
	if err != nil || w <= 0 {
		w = 80
	}
	return max(60, min(w, 110))
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

// visibleLen is the number of runes a string occupies on screen.
func visibleLen(s string) int { return utf8.RuneCountInString(stripANSI(s)) }

// padRight pads s with spaces to n visible columns.
func padRight(s string, n int) string {
	if gap := n - visibleLen(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

// truncate shortens plain text to n runes, ending with an ellipsis.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
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

// reflow turns authored text into logical lines. Blank lines separate
// paragraphs, lines starting with "- " are bullets, and other newlines are
// soft (joined with a space).
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
		case line == "":
			flush()
			out = append(out, "")
		case strings.HasPrefix(line, "- "):
			flush()
			cur = line
		case cur == "":
			cur = line
		default:
			cur += " " + line
		}
	}
	flush()
	return out
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
		first, rest := indent, indent
		if strings.HasPrefix(logical, "- ") {
			first = indent + "• "
			rest = indent + "  "
			logical = strings.TrimPrefix(logical, "- ")
		}
		line, prefix := "", first
		for _, word := range strings.Fields(logical) {
			if line != "" && visibleLen(prefix)+utf8.RuneCountInString(line)+1+utf8.RuneCountInString(word) > width {
				lines = append(lines, prefix+line)
				line, prefix = "", rest
			}
			if line == "" {
				line = word
			} else {
				line += " " + word
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

// banner is the startup logo.
func banner() string {
	art := []string{
		"   ▄▀█ █▀▀ ▄▀█ █▀▄ █▀▀ █▀▄▀█ █▄█",
		"   █▀█ █▄▄ █▀█ █▄▀ ██▄ █░▀░█ ░█░",
	}
	return "\n" + sty.Bold(sty.Cyan(art[0])) + "\n" + sty.Bold(sty.Magenta(art[1])) + "\n" +
		sty.Dim("   Systems & AI · Campus Registry · zero dependencies") + "\n"
}

// heading renders a double-ruled section title in the given colour.
func heading(title string, color func(string) string, width int) string {
	rule := strings.Repeat("═", width)
	return "\n" + color(rule) + "\n" + color(sty.Bold("  "+title)) + "\n" + color(rule)
}
