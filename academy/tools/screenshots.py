#!/usr/bin/env python3
"""Regenerate the README screenshots from the real app.

Each shot starts the academy in a pseudo-terminal on a copy of a demo
registry (written by the TestWriteDemoRegistry test), types the keys that
open a screen, converts the ANSI colour output to HTML, and has headless
Chromium render it to docs/<name>.png.

Usage (Linux, with Chromium available; CHROME can point at the binary):
  go build -o academy . && python3 tools/screenshots.py [name ...]
"""
import os, pty, sys, time, select, re, html, shutil, subprocess, tempfile, fcntl, termios, struct, glob

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
BIN = os.environ.get("ACADEMY_BIN", os.path.join(ROOT, "academy"))
DOCS = os.path.join(ROOT, "docs")


def find_chrome():
    if os.environ.get("CHROME"):
        return os.environ["CHROME"]
    for pattern in ["/opt/pw-browsers/chromium_headless_shell-*/chrome-linux/headless_shell",
                    "/opt/pw-browsers/chromium-*/chrome-linux/chrome"]:
        hits = sorted(glob.glob(pattern))
        if hits:
            return hits[-1]
    for name in ["chromium", "chromium-browser", "google-chrome", "headless_shell"]:
        path = shutil.which(name)
        if path:
            return path
    sys.exit("Chromium not found; set CHROME=/path/to/chrome")


# name: (columns, keys to type, first line regex, last line regex or None,
#        max lines[, which key's output to start from: -1 = the last])
SHOTS = {
    "menu": (80, [], r"Hours ", r"└─", 40),
    "stages": (80, ["6"], r"STUDY HALL", r"\[16\]", 40),
    "concept-card": (80, ["6", "1", "1", "5"], r"┏━", r"◆ Mental model", 70),
    "under-the-hood": (80, ["6", "1", "1", "6"], r"🔧 Under the hood", r"◆ Mental model", 30),
    "exercise-gym": (80, ["6", "0", "2", "3", "3"], r"REAL-WORLD", r"\(y/n\)|›", 30),
    "daily-review": (80, ["9"], r"DAILY REVIEW", r"Your answer", 30),
    "mastery": (80, ["6", "0"], r"Stage 0 · How Computers Work", r"8\. ", 45),
    "ledger": (80, ["1"], r"TRACK F", r"Stage 2 ·", 45),
    "classic-corner": (80, ["6", "0", "8"], r"CLASSIC CORNER", r"Type-in", 40),
    "type-in-lab": (80, ["6", "0", "8", "1"], r"TYPE-IN LAB", r"1\. Predict", 50),
    "connections": (80, ["6", "1", "1", "5"], r"🔗 Connects to", r"📝|📖", 16),
    "dictionary": (80, ["d", "type casting"], r"┏━ Type casting", r"┗", 45),
    "roadmap": (80, ["m"], r"ROADMAP", r"more sense later", 40),
    "progress": (80, ["p"], r"PROGRESS REPORT", r"Achievements", 45),
    "library": (80, ["l"], r"LIBRARY", r"Stage 2 ·", 30),
    "techwatch": (80, ["w", "4"], r"MY TECH RADAR", r"press Enter", 30),
    "start-here": (80, ["0"], r"WHY WE START WITH C", r"GET A C COMPILER", 20),
    "struggle-clock": (80, ["c", "y", "6", "1", "2", "5", "2", "y"], r"PRACTICE", r"learning happens", 30, -2),
}

SGR_COLORS = {"30": "#45475a", "31": "#f38ba8", "32": "#a6e3a1", "33": "#f9e2af", "34": "#89b4fa",
              "35": "#cba6f7", "36": "#94e2d5", "37": "#cdd6f4", "90": "#7f849c"}
CSI = re.compile(r"\x1b\[([0-9;?]*)([A-Za-z])")


def run_session(cols, keys, registry, from_key=-1):
    pid, fd = pty.fork()
    if pid == 0:
        os.environ["TERM"] = "xterm-256color"
        os.environ.pop("NO_COLOR", None)
        os.execv(BIN, ["academy", "-registry", registry])
    fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", 200, cols, 0, 0))
    out = bytearray()

    def pump(t):
        end = time.time() + t
        while time.time() < end:
            r, _, _ = select.select([fd], [], [], 0.03)
            if r:
                try:
                    out.extend(os.read(fd, 1 << 16))
                except OSError:
                    return

    pump(1.2)
    marks = []
    for k in keys:
        marks.append(len(out))
        os.write(fd, (k + "\r").encode())
        pump(0.5)
    mark = marks[from_key] if marks else 0  # no keys: the start-up screen itself
    pump(0.6)
    try:
        os.kill(pid, 9)
        os.waitpid(pid, 0)
    except OSError:
        pass
    return bytes(out[mark:]).decode("utf-8", "replace")


def terminal_lines(text):
    """Apply the few control sequences the app uses; keep SGR colours."""
    lines, cur = [], ""
    i = 0
    while i < len(text):
        ch = text[i]
        if ch == "\x1b":
            m = CSI.match(text, i)
            if m:
                params, cmd = m.groups()
                if cmd == "m":
                    cur += m.group(0)
                elif cmd == "J":
                    lines, cur = [], ""
                elif cmd == "K":
                    pass
                i = m.end()
                continue
            i += 1
            continue
        if ch == "\n":
            lines.append(cur)
            cur = ""
        elif ch == "\r":
            if i + 1 < len(text) and text[i + 1] == "\n":
                pass
            else:
                cur = ""
        elif ch == "\x07":
            pass
        else:
            cur += ch
        i += 1
    if cur:
        lines.append(cur)
    return lines


def plain(line):
    return CSI.sub("", line)


def to_html(line):
    out, style = [], {}
    pos = 0
    def emit(s):
        if not s:
            return
        css = []
        if "color" in style:
            css.append(f"color:{style['color']}")
        if style.get("bold"):
            css.append("font-weight:700")
        if style.get("dim"):
            css.append("opacity:.72")
        if style.get("italic"):
            css.append("font-style:italic")
        if style.get("under"):
            css.append("text-decoration:underline")
        text = html.escape(s)
        out.append(f'<span style="{";".join(css)}">{text}</span>' if css else text)
    for m in CSI.finditer(line):
        emit(line[pos:m.start()])
        pos = m.end()
        if m.group(2) != "m":
            continue
        for p in (m.group(1) or "0").split(";"):
            if p in ("0", ""):
                style = {}
            elif p == "1":
                style["bold"] = True
            elif p == "2":
                style["dim"] = True
            elif p == "3":
                style["italic"] = True
            elif p == "4":
                style["under"] = True
            elif p in SGR_COLORS:
                style["color"] = SGR_COLORS[p]
    emit(line[pos:])
    return "".join(out)


PAGE = """<!doctype html><html><head><meta charset="utf-8"><style>
html,body{{margin:0;background:#11111b}}
.win{{margin:18px;background:#1e1e2e;border-radius:10px;box-shadow:0 8px 30px rgba(0,0,0,.45);display:inline-block}}
.bar{{height:30px;display:flex;align-items:center;padding:0 12px;gap:8px;color:#7f849c;font:12px 'DejaVu Sans',sans-serif}}
.dot{{width:12px;height:12px;border-radius:6px;display:inline-block}}
pre{{margin:0;padding:6px 18px 18px;color:#cdd6f4;font:14px/17px 'DejaVu Sans Mono','Noto Color Emoji',monospace;width:{cols}ch}}
</style></head><body><div class="win"><div class="bar"><span class="dot" style="background:#f38ba8"></span><span class="dot" style="background:#f9e2af"></span><span class="dot" style="background:#a6e3a1"></span><span style="margin-left:10px">academy — {title}</span></div><pre>{body}</pre></div></body></html>"""


def shoot(name, chrome, demo):
    cols, keys, first, last, limit, *rest = SHOTS[name]
    from_key = rest[0] if rest else -1
    tmp = tempfile.mkdtemp(prefix="academy-shot-")
    reg = os.path.join(tmp, "reg.json")
    shutil.copy(demo, reg)
    lines = terminal_lines(run_session(cols, keys, reg, from_key))
    start = next((i for i, l in enumerate(lines) if re.search(first, plain(l))), None)
    if start is None:
        raise SystemExit(f"{name}: start marker {first!r} not found")
    # Keep a heading's rule line above it, if there is one.
    if start > 0 and set(plain(lines[start - 1]).strip()) == {"═"}:
        start -= 1
    end = min(len(lines), start + limit)
    if last:
        for i in range(start + 1, len(lines)):
            if re.search(last, plain(lines[i])):
                end = min(i + 1, start + limit)
                break
    chosen = [l for l in lines[start:end] if not re.search(r"press Enter to continue|Enter for more", plain(l))]
    while chosen and not plain(chosen[-1]).strip():
        chosen.pop()
    body = "\n".join(to_html(l) for l in chosen)
    page = os.path.join(tmp, "shot.html")
    open(page, "w").write(PAGE.format(cols=cols, title=name.replace("-", " "), body=body))
    width = int(cols * 8.43) + 36 + 36 + 4
    height = len(chosen) * 17 + 30 + 24 + 36 + 14
    target = os.path.join(DOCS, name + ".png")
    subprocess.run([chrome, "--no-sandbox", "--hide-scrollbars", "--force-device-scale-factor=1", "--disable-lcd-text",
                    f"--window-size={width},{height}", f"--screenshot={target}", "file://" + page],
                   check=True, capture_output=True)
    print(f"  {name}.png  {len(chosen)} lines")


def main():
    chrome = find_chrome()
    demo = os.path.join(tempfile.mkdtemp(prefix="academy-demo-"), "demo.json")
    subprocess.run(["go", "test", "-count=1", "-run", "TestWriteDemoRegistry", "."], cwd=ROOT, check=True,
                   env={**os.environ, "ACADEMY_DEMO_OUT": demo}, capture_output=True)
    names = sys.argv[1:] or list(SHOTS)
    os.makedirs(DOCS, exist_ok=True)
    for n in names:
        shoot(n, chrome, demo)


if __name__ == "__main__":
    main()
