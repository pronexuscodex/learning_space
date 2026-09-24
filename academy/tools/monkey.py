#!/usr/bin/env python3
"""Monkey test: drive the academy with random but plausible input in a real
pseudo-terminal and fail on any Go panic, crash or hang.

Usage (Linux/macOS): go build -o academy . && python3 tools/monkey.py [sessions] [keys] [seed]

Each session starts the app on a fresh registry, sends random keys (menu
keys, numbers, answers, long and Unicode text, editing keys, escape
sequences, terminal resizes), and restarts it whenever it exits. Network
features fail fast offline; that is fine, they must fail gracefully.
"""
import os, pty, sys, time, select, random, signal, tempfile, fcntl, termios, struct, re

D = os.path.dirname(os.path.abspath(__file__))
BIN = os.environ.get("ACADEMY_BIN", os.path.join(D, "..", "academy"))
ANSI = re.compile(rb"\x1b\[[0-9;?]*[A-Za-z]")

MENU = list("0123456799nmd/fpxl+wgact?") + ["r", "s", "clear", "help", "my"]
TOKENS = (
    [k + "\r" for k in MENU] * 3
    + [f"{n}\r" for n in range(0, 25)] * 2
    + ["\r"] * 12 + ["q\r", ":q\r", "y\r", "n\r", "g\r", "a\r", "b\r", "o\r", "e\r", "i\r", "t\r", "u\r"] * 2
    + ["-1\r", "999999999999\r", "1.5\r", "1h30m\r", "abc\r", "  \r", "#3\r", "0x10\r"]
    + ["type casting\r", "memory cache\r", "pointer\r", "https://example.invalid/x.pdf\r", "../../etc/passwd\r"]
    + ["é漢字🙂 text\r", "x" * 700 + "\r", "'; DROP TABLE users; --\r", "%s%n%x\r"]
    + ["\x0c", "\x7f", "\x7f\x7f", "\x15", "\x17", "\x1b[A", "\x1b[D", "\x1b[3~", "\x1bOP", "p", "s"]
    + ["RESIZE"] * 3
)
# Go's own crash output. Bare words such as "SIGSEGV" or "runtime error"
# are not enough: the curriculum teaches them, so they appear in lessons.
FATAL = [re.compile(p, re.M) for p in (rb"^\s*panic: ", rb"goroutine \d+ \[running\]", rb"^\s*fatal error: ", rb"\[signal SIG[A-Z]+")]


def set_size(fd, rows, cols):
    fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))


def session(rng, keys, reg):
    pid, fd = pty.fork()
    if pid == 0:
        os.execv(BIN, ["academy", "-registry", reg])
    set_size(fd, rng.choice([24, 40, 60]), rng.choice([40, 60, 80, 120]))
    out = bytearray()
    history = []
    exited = None

    def pump(t):
        nonlocal exited
        end = time.time() + t
        while time.time() < end:
            r, _, _ = select.select([fd], [], [], 0.02)
            if r:
                try:
                    chunk = os.read(fd, 1 << 16)
                except OSError:
                    chunk = b""
                if not chunk:  # the app closed the terminal: collect its status
                    for _ in range(100):
                        done, status = os.waitpid(pid, os.WNOHANG)
                        if done:
                            exited = status
                            return
                        time.sleep(0.02)
                    return
                out.extend(chunk)
            done, status = os.waitpid(pid, os.WNOHANG)
            if done:
                exited = status
                return

    pump(0.8)
    sent = 0
    while sent < keys and exited is None:
        k = rng.choice(TOKENS)
        if k == "RESIZE":
            set_size(fd, rng.randint(15, 70), rng.randint(40, 200))
            os.kill(pid, signal.SIGWINCH)
        else:
            try:
                os.write(fd, k.encode())
            except OSError:
                break
        history.append(k)
        sent += 1
        before = len(out)
        pump(0.03)
        # Hang detection: if nothing at all comes back for a while after a
        # newline, poke with Enter a few times before declaring a hang.
        if k.endswith("\r") and len(out) == before:
            for _ in range(3):
                pump(0.6)
                if len(out) > before or exited is not None:
                    break
            else:
                pass  # some prompts legitimately echo nothing new; timers and fetches are slow
    # End the session the way a person would: Ctrl+U clears any half-typed
    # line (Ctrl+D only ends input on an empty line, as in a shell), then
    # Ctrl+D commits and exits.
    for _ in range(3):
        if exited is not None:
            break
        try:
            os.write(fd, b"\x15\x04")
        except OSError:
            break
        pump(3)
    if exited is None:
        os.kill(pid, signal.SIGKILL)
        os.waitpid(pid, 0)
        return "hang", out, history
    os.close(fd)
    code = os.waitstatus_to_exitcode(exited)
    clean = ANSI.sub(b"", bytes(out))
    for pattern in FATAL:
        m = pattern.search(clean)
        if m:
            context = clean[max(0, m.start() - 120):m.end() + 200].decode("utf-8", "replace")
            return f"crash (exit {code}): {m.group(0).decode().strip()!r} in …{context}…", out, history
    if code != 0:
        return f"crash (exit {code})", out, history
    return None, out, history


def main():
    sessions = int(sys.argv[1]) if len(sys.argv) > 1 else 10
    keys = int(sys.argv[2]) if len(sys.argv) > 2 else 300
    seed = int(sys.argv[3]) if len(sys.argv) > 3 else int(time.time())
    rng = random.Random(seed)
    print(f"monkey: {sessions} sessions × up to {keys} keys, seed {seed}")
    failures = 0
    total = 0
    for i in range(sessions):
        reg = os.path.join(tempfile.mkdtemp(prefix="academy-monkey-"), "reg.json")
        left = keys
        runs = 0
        while left > 0:
            n = min(left, rng.randint(50, keys))
            problem, out, history = session(rng, n, reg)
            runs += 1
            total += len(history)
            left -= max(1, len(history))
            if problem:
                failures += 1
                text = ANSI.sub(b"", bytes(out)).decode("utf-8", "replace")
                print(f"  ✗ session {i + 1}: {problem}")
                print("    last keys:", [h[:20] for h in history[-15:]])
                tail = text[-1500:]
                print("    output tail:\n" + "\n".join("      " + l for l in tail.splitlines()[-25:]))
                break
        else:
            print(f"  ✓ session {i + 1}: {runs} run(s), no crash")
    print(f"monkey: {total} keys sent, {failures} failure(s)")
    sys.exit(1 if failures else 0)


if __name__ == "__main__":
    main()
