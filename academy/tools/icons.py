#!/usr/bin/env python3
"""Generate every icon asset from assets/icon.svg and assets/icon-small.svg.

  python3 tools/icons.py        # from the academy/ directory

Writes:
  assets/png/icon-<size>.png       16–1024 px (small sizes from icon-small.svg)
  assets/academy.ico               Windows icon, 16–256 px
  assets/social-preview.png        1280×640 card for GitHub's social preview
  rsrc_windows_amd64.syso          the icon as a Windows resource; Go's linker
  rsrc_windows_arm64.syso          embeds it into academy.exe automatically

Needs Python 3 and Chromium (headless). No Go tools or packages: the .syso
files are COFF objects with a .rsrc section, written by hand below.
"""
import glob, os, shutil, struct, subprocess, sys, tempfile

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ASSETS = os.path.join(ROOT, "assets")
BIG = [1024, 512, 256, 128, 64, 48]
SMALL = [32, 24, 16]
ICO_SIZES = [16, 24, 32, 48, 64, 128, 256]


def chrome():
    if os.environ.get("CHROME"):
        return os.environ["CHROME"]
    for pattern in ["/opt/pw-browsers/chromium_headless_shell-*/chrome-linux/headless_shell",
                    "/opt/pw-browsers/chromium-*/chrome-linux/chrome"]:
        hits = sorted(glob.glob(pattern))
        if hits:
            return hits[-1]
    for name in ["chromium", "chromium-browser", "google-chrome"]:
        if shutil.which(name):
            return shutil.which(name)
    sys.exit("Chromium not found; set CHROME=/path/to/chrome")


def screenshot(html, width, height, out):
    tmp = tempfile.mkdtemp(prefix="academy-icon-")
    page = os.path.join(tmp, "page.html")
    with open(page, "w") as f:
        f.write(html)
    subprocess.run([chrome(), "--no-sandbox", "--hide-scrollbars", "--force-device-scale-factor=1",
                    "--allow-file-access-from-files", "--default-background-color=00000000",
                    f"--window-size={width},{height}", f"--screenshot={out}", "file://" + page],
                   check=True, capture_output=True)


def render_png(svg, size, out):
    src = "file://" + os.path.join(ASSETS, svg)
    html = (f'<html><body style="margin:0;background:transparent">'
            f'<img src="{src}" width="{size}" height="{size}" style="display:block"></body></html>')
    screenshot(html, size, size, out)


def social_preview(out):
    src = "file://" + os.path.join(ASSETS, "icon.svg")
    html = f"""<html><body style="margin:0;width:1280px;height:640px;background:linear-gradient(135deg,#1e1e2e,#11111b);
display:flex;align-items:center;gap:64px;padding:0 96px;box-sizing:border-box;font-family:'DejaVu Sans',sans-serif">
<img src="{src}" width="300" height="300">
<div><div style="font-size:64px;font-weight:700;color:#cdd6f4;line-height:1.1">Systems &amp; AI<br>Academy</div>
<div style="margin-top:24px;font-size:28px;color:#94e2d5">From how computers work and C<br>to operating systems, security and AI</div>
<div style="margin-top:28px;font-size:22px;color:#7f849c;font-family:'DejaVu Sans Mono',monospace;white-space:nowrap">17 stages · spaced repetition · offline</div></div>
</body></html>"""
    screenshot(html, 1280, 640, out)


def build_ico(pngs):
    """An ICO file whose images are PNGs (supported since Windows Vista)."""
    header = struct.pack("<HHH", 0, 1, len(pngs))
    offset = 6 + 16 * len(pngs)
    entries, blobs = b"", b""
    for size, data in pngs:
        dim = 0 if size >= 256 else size
        entries += struct.pack("<BBBBHHII", dim, dim, 0, 0, 1, 32, len(data), offset)
        blobs += data
        offset += len(data)
    return header + entries + blobs


# --- Windows resources --------------------------------------------------------
RT_ICON, RT_GROUP_ICON = 3, 14
LANG_EN_US = 0x0409
MACHINES = {"amd64": (0x8664, 3), "arm64": (0xAA64, 2)}  # (COFF machine, ADDR32NB relocation type)


def build_syso(pngs, arch):
    """A COFF object holding a .rsrc section with RT_ICON images and one
    RT_GROUP_ICON (ID 1), the layout the Go linker merges into a PE file."""
    machine, reloc_type = MACHINES[arch]
    icons = [(i + 1, size, data) for i, (size, data) in enumerate(pngs)]
    group = struct.pack("<HHH", 0, 1, len(icons))
    for icon_id, size, data in icons:
        dim = 0 if size >= 256 else size
        group += struct.pack("<BBBBHHIH", dim, dim, 0, 0, 1, 32, len(data), icon_id)

    def directory(n_ids):
        return struct.pack("<IIHHHH", 0, 0, 0, 0, 0, n_ids)

    # Layout: root dir → type dirs → name dirs → language dirs → data entries → data.
    types = [(RT_ICON, [(i, d) for i, _, d in icons]), (RT_GROUP_ICON, [(1, group)])]
    root_size = 16 + 8 * len(types)
    type_dirs_size = sum(16 + 8 * len(items) for _, items in types)
    n_leaves = sum(len(items) for _, items in types)
    lang_dirs_size = n_leaves * (16 + 8)
    data_entries_off = root_size + type_dirs_size + lang_dirs_size
    data_off = data_entries_off + n_leaves * 16

    out = bytearray(directory(len(types)))
    relocs = []
    type_dir_off = root_size
    for type_id, items in types:
        out += struct.pack("<II", type_id, 0x80000000 | type_dir_off)
        type_dir_off += 16 + 8 * len(items)
    lang_dir_off = root_size + type_dirs_size
    leaf = 0
    for type_id, items in types:
        out += directory(len(items))
        for name_id, _ in items:
            out += struct.pack("<II", name_id, 0x80000000 | (lang_dir_off + leaf * 24))
            leaf += 1
    for i in range(n_leaves):
        out += directory(1)
        out += struct.pack("<II", LANG_EN_US, data_entries_off + i * 16)
    blobs = [data for _, items in types for _, data in items]
    cursor = data_off
    placed = []
    for data in blobs:
        placed.append(cursor)
        cursor = (cursor + len(data) + 7) & ~7
    for i, data in enumerate(blobs):
        relocs.append(len(out))  # the OffsetToData field is an RVA: relocate it
        out += struct.pack("<IIII", placed[i], len(data), 0, 0)
    assert len(out) == data_off
    for i, data in enumerate(blobs):
        out += b"\0" * (placed[i] - len(out))
        out += data
    out += b"\0" * ((-len(out)) % 8)

    raw_ptr = 20 + 40
    reloc_ptr = raw_ptr + len(out)
    sym_ptr = reloc_ptr + 10 * len(relocs)
    coff = struct.pack("<HHIIIHH", machine, 1, 0, sym_ptr, 1, 0, 0)
    section = struct.pack("<8sIIIIIIHHI", b".rsrc", 0, 0, len(out), raw_ptr, reloc_ptr, 0,
                          len(relocs), 0, 0x40000040)  # initialized data, readable
    reloc_table = b"".join(struct.pack("<IIH", r, 0, reloc_type) for r in relocs)
    symbol = struct.pack("<8sIhHBB", b".rsrc", 0, 1, 0, 3, 0)  # section 1, static
    strings = struct.pack("<I", 4)
    return coff + section + bytes(out) + reloc_table + symbol + strings


def main():
    png_dir = os.path.join(ASSETS, "png")
    os.makedirs(png_dir, exist_ok=True)
    for size in BIG:
        render_png("icon.svg", size, os.path.join(png_dir, f"icon-{size}.png"))
    for size in SMALL:
        render_png("icon-small.svg", size, os.path.join(png_dir, f"icon-{size}.png"))
    pngs = [(s, open(os.path.join(png_dir, f"icon-{s}.png"), "rb").read()) for s in ICO_SIZES]
    with open(os.path.join(ASSETS, "academy.ico"), "wb") as f:
        f.write(build_ico(pngs))
    for arch in MACHINES:
        with open(os.path.join(ROOT, f"rsrc_windows_{arch}.syso"), "wb") as f:
            f.write(build_syso(pngs, arch))
    social_preview(os.path.join(ASSETS, "social-preview.png"))
    print("icons written:", ", ".join(f"{s}px" for s in sorted(BIG + SMALL)),
          "+ academy.ico + social-preview.png + rsrc_windows_{amd64,arm64}.syso")


if __name__ == "__main__":
    main()
