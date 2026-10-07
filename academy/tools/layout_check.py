#!/usr/bin/env python3
"""Drive every major screen through a real pseudo-terminal at several
widths and report any line wider than the terminal.

Usage (Linux/macOS): go build -o academy . && python3 tools/layout_check.py [widths...]
"""
import tempfile, os, pty, time, select, sys, re, fcntl, termios, struct, unicodedata
D = os.path.dirname(os.path.abspath(__file__))
ANSI = re.compile(r"\x1b\[[0-9;?]*[A-Za-z]")

def width(s):
    w = 0
    for ch in s:
        if unicodedata.combining(ch): continue
        w += 2 if unicodedata.east_asian_width(ch) in "WF" else 1
    return w

# Enroll a lab with a long name, then visit every major screen.
KEYS = [
  "2\r", "A\r", "5\r", "A really quite long lab name for testing\r", "Architecture notes that are long enough to need truncation on narrow screens\r", "3\r",
  "1\r", "\r"*3,            # ledger (paged)
  "0\r", "\r"*3,            # start here
  "?\r", "\r"*2,            # shortcuts
  "6\r", "5\r",             # study hall, stage 5 (diagrams)
  "1\r", "4\r", "\r", "n\r",  # concept: memory hierarchy (has diagram), don't mark
  "3\r", "\r"*2,            # glossary
  "4\r", "\r"*2,            # resources
  "5\r", "n\r",             # blueprints
  "8\r", "1\r", "\r", "n\r", "3\r",  # classic corner → type-in lab listing → skip prediction → not run → back
  "10\r",                   # back to main
  "n\r", "q\r",             # what's next, then cancel
  "/\r", "memory cache\r", "1\r", "\r"*3, "n\r",  # search → open first hit (a concept) → don't mark
  "f\r", "\r", "1\r", "a goal that is rather long for a narrow terminal\r", "p", "p", "s", # focus timer, stop at once
  "p\r", "\r"*3,            # progress report (paged)
  "x\r",                    # export notes
  "g\r", "\r", "\r", "\r",   # weekly goals: accept the suggestions
  "m\r", "\r"*3, "2\r", "10\r",  # roadmap → open stage 2's Study Hall → back
  "a\r", "\r"*3,            # achievements
  "k\r",                    # share card
  "j\r", "2\r", "1\r", "1\r", "y\r", "Everything in a computer is a number made of bits\r",  # workout: skip review, nothing to recall yet, learn the first concept
  "1\r", "2\r", "2\r", "2\r", "y\r", "3\r",  # practice: climb all three rungs, done, "with the hint"
  "1\r", "Choosing between bits and bytes\r", "n\r",  # reflect
  "p\r", "\r"*4,          # progress report, now with "Thinking for yourself"
  "o\r", "1\r", "3\r", "2\r", "6\r",  # settings: colour-blind theme, plain symbols (the rest of the run checks them)
  "l\r", "1\r", "q\r",  # library → try downloading the first PDF → back
  "d\r", "type casting\r", "1\r", "\r", "c\r", "4\r", "1\r", "\r", "a\r", "q\r", "q\r", "q\r",  # dictionary: word → related → categories → A–Z
  "+\r", "a\r", "Modern C, a free book on writing C well\r", "https://example.org/modern-c.pdf\r", "1\r", "5\r", "A long note that should wrap neatly at every terminal width\r", "q\r",  # my resources
  "w\r", "1\r", "q\r", "4\r", "q\r", "5\r", "q\r", "7\r",  # tech watch: method, radar, sources
  "9\r", "q\r",             # daily review intro, then stop
  "c\r", "y\r", "6\r", "1\r", "2\r", "5\r", "2\r", "y\r", "n\r", "n\r", "q\r", "10\r", "c\r", "y\r",  # Classic Mode: struggle clock message, then off again
  "5\r",                    # commit & exit
]

PAGER = re.compile(r"── (Enter for more · q to stop( · or type your choice)?|Enter more · q stop|end · press Enter to continue) ──")

def run(cols, rows=60):
    reg = os.path.join(tempfile.gettempdir(), f"academy-layout-{cols}.json")
    if os.path.exists(reg): os.remove(reg)
    pid, fd = pty.fork()
    if pid == 0:
        os.environ["ACADEMY_LIBRARY"] = tempfile.mkdtemp(prefix="academy-library-")  # never the real Documents
        os.execv(os.environ.get("ACADEMY_BIN", os.path.join(D, "..", "academy")), ["academy", "-registry", reg])
    fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))
    out = b""
    def read(t):
        nonlocal out
        end = time.time() + t
        while time.time() < end:
            r, _, _ = select.select([fd], [], [], 0.03)
            if r:
                try: out += os.read(fd, 1 << 16)
                except OSError: return
    read(1.0)
    pause = re.compile(r"(Enter for more · q to stop( · or type your choice)?|Enter more · q stop) ──\s*$")
    for k in KEYS:
        # A long screen may pause ("Enter for more"). If the script is not
        # about to press Enter anyway, press it first, as a person would,
        # so the keys stay in step with the screens.
        for _ in range(10):
            tail = ANSI.sub("", out[-400:].decode("utf-8", "replace"))
            if k == "\r" or not pause.search(tail):
                break
            os.write(fd, b"\r"); read(0.25)
        os.write(fd, k.encode()); read(0.25)
    read(1.0)
    text = ANSI.sub("", out.decode("utf-8", "replace")).replace("\r\n", "\n")
    # A bare \r redraws the line in place (live timer): keep what is left.
    text = "\n".join(l.split("\r")[-1] for l in text.split("\n"))
    if os.environ.get("LAYOUT_DUMP"):
        with open(os.environ["LAYOUT_DUMP"] + f".{cols}.txt", "w") as f: f.write(text)
    bad = []
    for line in text.split("\n"):
        if "http" in line:   # long URLs are left for the terminal to wrap
            continue
        if " › " in line and not line.rstrip().endswith("›"):  # echoed typing: the terminal wraps it
            continue
        # A pager prompt is an input point too: on a slow machine the next
        # keys can be typed while it is showing, and the terminal echoes
        # them after it. Measure only the app's own part of the line.
        m = PAGER.search(line)
        if m:
            line = line[:m.end()]
        if width(line) > cols:
            bad.append((width(line), line))
    return bad

cols_list = [int(c) for c in sys.argv[1:]] or [40, 50, 60, 80, 100, 130]
for cols in cols_list:
    bad = run(cols)
    print(f"== {cols} columns: {len(bad)} overflowing line(s)")
    seen = set()
    for w, l in bad:
        key = l.strip()[:40]
        if key in seen: continue
        seen.add(key)
        print(f"   [{w}] {l[:120]}")
        if len(seen) >= 12: print("   …"); break
