# SAFE JOIN -- keep user-supplied file names inside one folder.
# A download server, an upload handler, a zip extractor: all of them
# join a folder with a name that came from outside.
import posixpath as path

BASE = "/srv/files"

def naive_join(name):
    return path.join(BASE, name)

def safe_join(name):
    full = path.normpath(path.join(BASE, name))       # 1. canonicalise
    if path.commonpath([full, BASE]) != BASE:          # 2. then check
        raise ValueError("outside " + BASE)
    return full

for name in ["report.pdf", "2026/notes.txt", "a/./b//c.txt",
             "../etc/passwd", "a/../../etc/passwd", "/etc/passwd"]:
    try:
        verdict = safe_join(name)
    except ValueError as e:
        verdict = "REFUSED (" + str(e) + ")"
    print(name)
    print("   naive:", naive_join(name))
    print("   safe: ", verdict)
