# Logo

The Academy mark is a CPU chip with a terminal prompt inside: the hardware you learn from the bottom up, and the shell where you do the work. The cursor is you, waiting to type the next line.

## Files

| File | Use it for |
|---|---|
| `../icon.svg` | The app icon: the mark on its dark tile. Master for every PNG, `academy.ico` and the Windows resources. |
| `../icon-small.svg` | The app icon at 16–32 px: no pins, heavier lines. `tools/icons.py` uses it for those sizes. |
| `academy-symbol-dark.svg` | The mark alone, on dark grounds. |
| `academy-symbol-light.svg` | The mark alone, on light grounds. |
| `academy-horizontal-dark.svg`, `academy-horizontal-light.svg` | Mark and wordmark side by side: headers, banners, slides. |
| `academy-stacked-dark.svg`, `academy-stacked-light.svg` | Mark above the wordmark: square spaces, title slides, stickers. |
| `academy-symbol-black.svg`, `academy-symbol-white.svg` | One colour: print, embossing, places that allow a single ink. |

Every file is plain SVG made of filled paths: no fonts, no strokes, no gradients, no filters. After changing `icon.svg` or `icon-small.svg`, run `python3 tools/icons.py` from `academy/` to regenerate the PNGs, the `.ico`, the Windows resources and the social preview.

## Colours

The mark uses the app's own theme, Catppuccin: Mocha on dark grounds, Latte on light grounds. Each colour has one job: the chip, the prompt and the cursor.

| Part | Dark grounds | Light grounds |
|---|---|---|
| Tile and dark ground | `#1e1e2e` (30, 30, 46) | `#eff1f5` (239, 241, 245) |
| Chip and pins | `#cba6f7` (203, 166, 247) | `#8839ef` (136, 57, 239) |
| Prompt `>` | `#94e2d5` (148, 226, 213) | `#179299` (23, 146, 153) |
| Cursor `_` | `#f9e2af` (249, 226, 175) | `#c07510` (192, 117, 16) |
| Wordmark | `#cdd6f4` (205, 214, 244) | `#4c4f69` (76, 79, 105) |

Every part reaches at least 3:1 contrast against its ground (`#1e1e2e` or white), and the wordmark at least 8:1. The light cursor is darker than Latte's yellow (`#df8e1d`, only 2.6:1 on white) for that reason.

## Wordmark

`academy` in lowercase, set in DejaVu Sans Mono Bold with letter spacing tightened by 3 %, and converted to outlines. A monospaced face because the app lives in a terminal. In the horizontal lockup the word runs from the top of the chip body (the top of the `d`) to its bottom (the tail of the `y`), and the gap between the mark and the word is 48 units on the 256-unit grid. DejaVu fonts are free to use, including in logos (Bitstream Vera licence).

## Clear space and minimum size

- Keep clear space on every side at least as wide as one pin's length plus the gap around it: a quarter of the mark's height. Nothing else, not text or edges, goes inside it.
- Below 48 px use the app icon, not the bare mark. Below 32 px use `icon-small.svg`, which is drawn for those sizes.
- The horizontal lockup needs at least 120 px of width. Under that, use the mark alone.

## Don't

- Recolour the parts, swap their colours, or add gradients, shadows or glows.
- Use the dark-ground files on light grounds or the reverse; pick the matching file instead.
- Stretch, rotate, or outline the mark.
- Retype the wordmark in another font, or change its case.
- Put the bare mark on a busy photo; use the app icon with its tile instead.
