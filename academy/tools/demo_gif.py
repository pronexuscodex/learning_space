#!/usr/bin/env python3
"""Make docs/demo.gif, the animated tour at the top of the README.

Every frame is the real app: it runs in a pseudo-terminal on the demo
registry, exactly as tools/screenshots.py does, and each screen is
rendered by headless Chromium at the same size, then joined into a GIF.

Usage (Linux, with Chromium and Pillow):
  go build -o academy . && python3 tools/demo_gif.py
"""
import os, re, shutil, subprocess, sys, tempfile

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import screenshots as shots  # noqa: E402

from PIL import Image  # noqa: E402

COLS, ROWS = 80, 30

# (caption, keys, first line regex, last line regex, seconds on screen)
SCENES = [
    ("21 stages, from how a CPU works to building AI", [], r"Hours ", r"└─", 3.0),
    ("today's workout: one guided session a day", ["j"], r"TODAY'S WORKOUT", r"Ready\?", 3.5),
    ("the Study Hall: pick a stage", ["6"], r"STUDY HALL", r"\[16\]", 2.5),
    ("every concept: analogy, example, diagram", ["6", "1", "1", "5"], r"┏━", None, 3.5),
    ("under the hood: the real bytes and commands", ["6", "1", "1", "6"], r"🔧 Under the hood", None, 3.0),
    ("real exercises, three per concept", ["6", "0", "2", "3", "3"], r"REAL-WORLD", r"\(y/n\)|›", 3.0),
    ("stuck? questions first, then where to look, then the hint", ["6", "0", "2", "3", "3", "2", "2"], r"🧭 Rung 1", r"this exercise\.", 4.0),
    ("spaced repetition, so you keep what you learn", ["9"], r"DAILY REVIEW", r"Your answer", 3.0),
    ("a dictionary of every programming word", ["d", "type casting"], r"┏━ Type casting", r"┗", 3.0),
    ("free books and papers, downloaded in the app", ["l"], r"LIBRARY", None, 3.0),
    ("your progress, ready to share", ["k"], r"╭", r"learning_space", 4.0),
]


def frame(chrome, demo, n, caption, keys, first, last, out_dir):
    tmp = tempfile.mkdtemp(prefix="academy-gif-")
    reg = os.path.join(tmp, "reg.json")
    shutil.copy(demo, reg)
    lines = shots.terminal_lines(shots.run_session(COLS, keys, reg))
    start = next((i for i, l in enumerate(lines) if re.search(first, shots.plain(l))), None)
    if start is None:  # the scene spans several keys: look from the first one
        shutil.copy(demo, reg)
        lines = shots.terminal_lines(shots.run_session(COLS, keys, reg, 0))
        start = next((i for i, l in enumerate(lines) if re.search(first, shots.plain(l))), None)
    if start is None:
        raise SystemExit(f"scene {n}: start marker {first!r} not found")
    if start > 0 and set(shots.plain(lines[start - 1]).strip()) == {"═"}:
        start -= 1
    end = min(len(lines), start + ROWS)
    if last:
        for i in range(start + 1, len(lines)):
            if re.search(last, shots.plain(lines[i])):
                end = min(i + 1, start + ROWS)
                break
    chosen = [l for l in lines[start:end]
              if not re.search(r"press Enter to continue|Enter for more", shots.plain(l))]
    chosen += [""] * (ROWS - len(chosen))  # every frame the same size
    body = "\n".join(shots.to_html(l) for l in chosen)
    page = os.path.join(tmp, "frame.html")
    with open(page, "w") as f:
        f.write(shots.PAGE.format(cols=COLS, title=caption, body=body))
    width = int(COLS * 8.43) + 36 + 36 + 4
    height = ROWS * 17 + 30 + 24 + 36 + 14
    png = os.path.join(out_dir, f"{n:02d}.png")
    subprocess.run([chrome, "--no-sandbox", "--hide-scrollbars", "--force-device-scale-factor=1",
                    "--disable-lcd-text", f"--window-size={width},{height}", f"--screenshot={png}",
                    "file://" + page], check=True, capture_output=True)
    return png


def main():
    chrome = shots.find_chrome()
    demo = os.path.join(tempfile.mkdtemp(prefix="academy-demo-"), "demo.json")
    subprocess.run(["go", "test", "-count=1", "-run", "TestWriteDemoRegistry", "."], cwd=shots.ROOT,
                   check=True, env={**os.environ, "ACADEMY_DEMO_OUT": demo}, capture_output=True)
    out_dir = tempfile.mkdtemp(prefix="academy-frames-")
    images, durations = [], []
    for n, (caption, keys, first, last, secs) in enumerate(SCENES):
        png = frame(chrome, demo, n, caption, keys, first, last, out_dir)
        # One shared-looking palette per frame keeps the colours crisp.
        images.append(Image.open(png).convert("RGB").quantize(colors=64, method=Image.Quantize.MEDIANCUT))
        durations.append(int(secs * 1000))
        print(f"  frame {n}: {caption}")
    target = os.path.join(shots.DOCS, "demo.gif")
    images[0].save(target, save_all=True, append_images=images[1:], duration=durations, loop=0,
                   optimize=True, disposal=1)
    print(f"  {target}: {os.path.getsize(target) // 1024} KB, {sum(durations) / 1000:.0f} s")


if __name__ == "__main__":
    main()
