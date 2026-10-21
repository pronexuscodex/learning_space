package main

// The full-screen pager, used when keys can be read one at a time (Linux,
// macOS and the BSDs at a real terminal). It works like man and less: the
// text fills the screen, nothing but a status line sits at the bottom,
// keys act at once without Enter, and the screen is restored afterwards.
// Elsewhere the line pager in pager.go is used.

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

const (
	altScreenOn  = "\x1b[?1049h\x1b[?25l" // alternate screen, cursor hidden
	altScreenOff = "\x1b[?25h\x1b[?1049l"
)

// Key names returned by readKey for keys that send escape sequences.
const (
	keyUp       = "up"
	keyDown     = "down"
	keyPageUp   = "pgup"
	keyPageDown = "pgdn"
	keyHome     = "home"
	keyEnd      = "end"
	keyEsc      = "esc"
)

// fullScreenOK reports whether the full-screen pager can be used.
// ACADEMY_PAGER=inline keeps long screens in the normal scrollback, paged
// with a prompt under each page, as before.
func (a *App) fullScreenOK() bool {
	return a.con.raw && a.con.screen && os.Getenv("TERM") != "dumb" && os.Getenv("ACADEMY_PAGER") != "inline"
}

// rawMode switches the terminal to single keys, through the test hook
// when one is set.
func (c *console) rawMode() (func(), error) {
	if c.enterRaw != nil {
		return c.enterRaw()
	}
	return enableRawInput(int(os.Stdin.Fd()))
}

// readKey reads one key press: a rune, or the name of a special key.
func (c *console) readKey() (rune, string, error) {
	r, _, err := c.in.ReadRune()
	if err != nil {
		return 0, "", err
	}
	if r != keyEscape {
		return r, "", nil
	}
	if c.in.Buffered() == 0 {
		return 0, keyEsc, nil // a lone Esc press
	}
	next, _, err := c.in.ReadRune()
	if err != nil || (next != '[' && next != 'O') {
		return 0, keyEsc, nil
	}
	var params strings.Builder
	for c.in.Buffered() > 0 {
		b, err := c.in.ReadByte()
		if err != nil {
			break
		}
		if b >= 0x40 && b <= 0x7e { // the final byte
			switch {
			case b == 'A':
				return 0, keyUp, nil
			case b == 'B':
				return 0, keyDown, nil
			case b == 'H':
				return 0, keyHome, nil
			case b == 'F':
				return 0, keyEnd, nil
			case b == '~':
				switch params.String() {
				case "5":
					return 0, keyPageUp, nil
				case "6":
					return 0, keyPageDown, nil
				case "1", "7":
					return 0, keyHome, nil
				case "4", "8":
					return 0, keyEnd, nil
				}
			}
			return 0, "", nil // another key (left, right, F1…): ignored
		}
		params.WriteByte(b)
	}
	return 0, "", nil
}

// fullPager is the state of one full-screen paging session.
type fullPager struct {
	a      *App
	render func()
	lines  []string
	width  int // the width the lines were laid out for
	top    int
	find   pagerSearch
	msg    string // shown in the status line until the next key
}

// pageFullScreen shows lines in the full-screen pager. It returns false,
// having shown nothing, when the terminal cannot switch to single keys.
// After it, with finalPause off, the part of the text last on screen is
// printed again, so the question that follows has it in view.
func (a *App) pageFullScreen(lines []string, finalPause bool, render func()) bool {
	c := a.con
	restore, err := c.rawMode()
	if err != nil {
		return false
	}
	leave := func() {
		fmt.Fprint(c.out, altScreenOff)
		restore()
	}
	// A signal (Ctrl+C, closing the window) restores the screen too.
	termMu.Lock()
	termRestore = leave
	termMu.Unlock()

	fmt.Fprint(c.out, altScreenOn)
	p := &fullPager{a: a, render: render, lines: lines, width: a.cols()}
	p.find.current = -1
	prefill := p.run()
	restoreTerminal()

	if !finalPause {
		top, end := p.top, min(p.top+p.bodyRows(), len(p.lines))
		a.println(strings.Join(p.lines[top:end], "\n"))
	}
	if prefill != "" {
		c.prefill = []rune(prefill)
	}
	return true
}

// bodyRows is how many text rows fit above the status line.
func (p *fullPager) bodyRows() int { return max(3, termHeight()-1) }

func (p *fullPager) maxTop() int { return max(0, len(p.lines)-p.bodyRows()) }

func (p *fullPager) scroll(n int) { p.top = min(max(0, p.top+n), p.maxTop()) }

// relayout lays the text out again after the terminal changes width,
// keeping the reader at the same point in it.
func (p *fullPager) relayout() {
	w := p.a.cols()
	if w == p.width || p.render == nil {
		return
	}
	at := float64(p.top) / float64(max(1, len(p.lines)))
	p.lines = renderLines(p.a, p.render)
	p.width = w
	p.top = min(int(at*float64(len(p.lines))), p.maxTop())
	if p.find.query != "" {
		p.find = newPagerSearch(p.lines, p.find.query)
	}
}

// run handles keys until the reader leaves. It returns a key typed to
// answer whatever comes next, if that is how the reader left. The state
// changes under the console lock, which the resize handler also takes to
// redraw at once when the window changes size.
func (p *fullPager) run() string {
	c := p.a.con
	c.setOnResize(func() { p.relayout(); p.draw() })
	defer c.setOnResize(nil)
	for {
		c.mu.Lock()
		p.relayout()
		p.draw()
		p.msg = ""
		c.mu.Unlock()
		r, name, err := c.readKey()
		if err != nil {
			return ""
		}
		c.mu.Lock()
		leave, typed := p.key(r, name)
		c.mu.Unlock()
		if leave {
			return typed
		}
	}
}

// key applies one key. It reports whether the reader leaves, and the key
// to pass on when it was typed to answer what comes next.
func (p *fullPager) key(r rune, name string) (bool, string) {
	body := p.bodyRows()
	switch {
	case r == 'q' || r == 'Q' || name == keyEsc:
		return true, ""
	case r == ' ' || r == 'f' || r == 0x06 || r == 0x16 || name == keyPageDown:
		p.scroll(body)
	case r == '\n' || r == '\r' || r == 'j' || r == 0x0e || r == 0x05 || name == keyDown:
		p.scroll(1)
	case r == 'k' || r == 0x10 || r == 0x19 || name == keyUp:
		p.scroll(-1)
	case r == 'b' || r == 0x02 || name == keyPageUp:
		p.scroll(-body)
	case r == 'd' || r == 0x04:
		p.scroll(body / 2)
	case r == 'u' || r == 0x15:
		p.scroll(-body / 2)
	case r == 'g' || r == '<' || name == keyHome:
		p.top = 0
	case r == 'G' || r == '>' || name == keyEnd:
		p.top = p.maxTop()
	case r == '/':
		if q, ok := p.ask("/"); ok && q != "" {
			p.find = newPagerSearch(p.lines, q)
			p.jump(p.find.next(p.top))
		}
	case (r == 'n' || r == 'N') && p.find.query != "":
		if r == 'n' {
			p.jump(p.find.next(p.top))
		} else {
			p.jump(p.find.prev(p.top))
		}
	case r == 'h' || r == 'H' || r == '?':
		p.help()
	case r == keyCtrlL || name != "" || r == 0:
		// redraw / an unused special key
	case unicode.IsPrint(r):
		// A number or a menu key: leave, and let it answer what comes
		// next (the reader goes on typing there).
		return true, string(r)
	}
	return false, ""
}

// jump moves to a search result, with the match on the second row.
func (p *fullPager) jump(at int, note string) {
	if at < 0 {
		p.msg = note
		return
	}
	p.top = min(max(0, at-1), p.maxTop())
	p.msg = note
}

// draw paints the screen: the text, then the status line.
func (p *fullPager) draw() {
	p.top = min(p.top, p.maxTop()) // the window may have grown
	var b strings.Builder
	b.WriteString("\x1b[H")
	body := p.bodyRows()
	for i := 0; i < body; i++ {
		b.WriteString("\x1b[2K")
		if n := p.top + i; n < len(p.lines) {
			line := p.lines[n]
			if p.find.query != "" {
				line = highlightMatches(line, p.find.query)
			}
			b.WriteString(line)
		} else {
			b.WriteString(sty.Gray("~"))
		}
		b.WriteString("\r\n")
	}
	b.WriteString("\x1b[2K" + p.status())
	fmt.Fprint(p.a.con.out, b.String())
}

// status is the bottom line, in reverse video like man's. It has longer
// and shorter forms for the width of the terminal.
func (p *fullPager) status() string {
	cols := max(20, termWidth()) - 1
	end := min(p.top+p.bodyRows(), len(p.lines))
	where := fmt.Sprintf("lines %d–%d of %d %d%%", p.top+1, end, len(p.lines), end*100/max(1, len(p.lines)))
	if end == len(p.lines) {
		where = fmt.Sprintf("(END) lines %d–%d of %d", p.top+1, end, len(p.lines))
	}
	if p.find.current >= 0 && len(p.find.matches) > 0 {
		where = fmt.Sprintf("“%s” %d/%d · %s", truncate(p.find.query, 16), p.find.current+1, len(p.find.matches), where)
	}
	if p.msg != "" {
		where = p.msg
	}
	keys := []string{
		"Space next page · ↑↓ scroll · b back · / find · h help · q quit",
		"Space next · b back · / find · h help · q quit",
		"Space next · q quit",
		"q quit",
	}
	if p.find.query != "" && len(p.find.matches) > 0 {
		keys[0] = "Space next page · n/N next/previous match · h help · q quit"
		keys[1] = "n next · N previous · h help · q quit"
		keys[2] = "n next · q quit"
	}
	text := " " + where
	for _, k := range keys {
		if t := " " + where + " · " + k + " "; visibleLen(t) <= cols {
			text = t
			break
		}
	}
	text = truncate(text, cols)
	text += strings.Repeat(" ", max(0, cols-visibleLen(text)))
	if !sty.on {
		return text
	}
	return "\x1b[7m" + text + "\x1b[27m"
}

// ask reads a line in the status line (the text to find), like less.
// Enter accepts it; Esc or an empty line cancels.
func (p *fullPager) ask(label string) (string, bool) {
	c := p.a.con
	var buf []rune
	show := func() {
		fmt.Fprint(c.out, "\r\x1b[2K"+label+string(buf)+"\x1b[?25h")
	}
	show()
	defer fmt.Fprint(c.out, "\x1b[?25l")
	for {
		r, name, err := c.readKey()
		switch {
		case err != nil || name == keyEsc || r == keyCtrlC:
			return "", false
		case r == '\n' || r == '\r':
			return strings.TrimSpace(string(buf)), true
		case r == keyBackspace || r == keyDelete:
			if len(buf) == 0 {
				return "", false
			}
			buf = buf[:len(buf)-1]
		case r == keyCtrlU:
			buf = buf[:0]
		case name == "" && unicode.IsPrint(r):
			if len(buf) < 60 {
				buf = append(buf, r)
			}
		}
		show()
	}
}

// pagerHelp is the page h shows: every key, like less's own help.
var pagerHelp = []string{
	"READING A LONG SCREEN",
	"",
	"  Space  f  PgDn      next page",
	"  b  PgUp             previous page",
	"  Enter  j  ↓         one line down",
	"  k  ↑                one line up",
	"  d  /  u             half a page down / up",
	"  g  <  Home          the start",
	"  G  >  End           the end",
	"  /word               find a word (every match is highlighted)",
	"  n  /  N             next / previous match",
	"  q  Esc              quit and continue",
	"  a number or key     quit, and use it as your answer to what comes next",
	"",
	"These are the keys of man and less, so they work there too.",
	"",
	"Press any key to go back.",
}

// help shows the keys until any key is pressed.
func (p *fullPager) help() {
	var b strings.Builder
	b.WriteString("\x1b[H\x1b[2J")
	for i, l := range pagerHelp {
		if i == 0 {
			l = sty.Bold(sty.Cyan("  " + l))
		}
		b.WriteString(l + "\r\n")
	}
	fmt.Fprint(p.a.con.out, b.String())
	if _, _, err := p.a.con.readKey(); err == io.EOF {
		return
	}
	fmt.Fprint(p.a.con.out, "\x1b[2J")
}

// renderLines runs a screen's render function and returns its lines.
func renderLines(a *App, render func()) []string {
	var buf strings.Builder
	out := a.con.out
	a.con.out = &buf
	render()
	a.con.out = out
	return strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
}
