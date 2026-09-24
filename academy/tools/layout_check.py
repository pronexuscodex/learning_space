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
  "9\r",                    # back to main
  "n\r", "q\r",             # what's next, then cancel
  "/\r", "memory cache\r", "1\r", "\r"*3, "n\r",  # search → open first hit (a concept) → don't mark
  "f\r", "\r", "1\r", "a goal that is rather long for a narrow terminal\r", "p", "p", "s", # focus timer, stop at once
  "p\r", "\r"*3,            # progress report (paged)
  "x\r",                    # export notes
  "5\r",                    # commit & exit
]

def run(cols, rows=60):
    reg = os.path.join(tempfile.gettempdir(), f"academy-layout-{cols}.json")
    if os.path.exists(reg): os.remove(reg)
    pid, fd = pty.fork()
    if pid == 0:
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
    for k in KEYS:
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
