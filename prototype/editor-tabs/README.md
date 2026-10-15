# PROTOTYPE: an editor that keeps tabs and highlights while editing

Throwaway. Never merged. It answers the wayfinder ticket
[An editor that keeps tabs and highlights while editing](https://github.com/DannyFestor/TuiSnip/issues/206).
The verdict is in the ticket's resolution comment.

A standalone Go module, so the repo's `go build ./...` and lint skip it. Run from this directory:

```
go run .                  # built-in sample: tabs, tab-aligned columns, wide runes, a block comment
go run . -kib 256         # a generated Go Fragment at the 256 KiB Fragment cap (~10,900 lines)
go run . -lines 60000     # short lines far past the old 10,000-line cap
go run . -file <path>     # any file, Language from its name
go test -v -run Timing .  # the timing table in timing.txt
```

## What it is

`textarea/` is the Bubbles v2.2.1 textarea (MIT, see `textarea/LICENSE-bubbles`) with
[bubbles#1030](https://github.com/charmbracelet/bubbles/pull/1030) applied (`highlighter.go`) and these changes:

- The sanitiser keeps tabs (`runeutil.ReplaceTabs("\t")` in `san()`).
- `tabs.go`: one tab-aware width (`widthFrom`) and renderer (`expandFrom`), used by wrap, `LineInfo`,
  vertical cursor movement, the render loop, and the highlighter's cell ranges.
- `wrap` keeps whitespace runes instead of rebuilding them as spaces, which turned every tab into one space.
- `MaxLines` is a field; 0 means no cap.
- `WindowedRender` draws only rows near the viewport; the rest become blank rows.
- Knobs for the reactions: tab width, tab stops, cursor on a tab, per-cell movement, tab glyph.

`highlight.go` adapts chroma to the `LineHighlighter` hook. `main.go` is the app: F-keys flip the knobs,
the bottom lines show the cursor state and timings.

Not tab-aware: mouse selection (`selection.go`).

## Timings

`timing.txt`: 120×40 area, 10 keypresses averaged, on an Apple Silicon Mac. Taken before the
`background` highlight mode existed; that mode moves the `lex` column off the keypress.
