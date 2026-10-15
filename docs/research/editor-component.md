# Editor component: can an existing one keep tabs and highlight?

Research for [#205](https://github.com/DannyFestor/TuiSnip/issues/205), part of the Settle in map [#201](https://github.com/DannyFestor/TuiSnip/issues/201). Sources checked 2026-10-15 against the Go module cache (`charm.land/bubbles/v2@v2.2.1`, `charm.land/bubbletea/v2@v2.1.0`, `charm.land/lipgloss/v2@v2.0.6`, `github.com/alecthomas/chroma/v2@v2.27.0`), the upstream repositories on GitHub, and their issues, PRs, discussions, and releases.

## Question

Before TuiSnip writes its own editor component, what does the field offer? Does the Bubbles v2 textarea keep tab characters, or has an option landed or been proposed? What is the state of bubbles#1030 (syntax highlighting in the textarea)? Does the 10,000-line cap still exist, and is it configurable? Which Go TUI text-editing components keep tabs and support highlighting, and could one be adopted or forked? If TuiSnip forks or writes its own, which parts of the stock textarea are worth keeping?

## Recommendation

**No existing component does both. Fork the Bubbles v2 textarea (MIT) into TuiSnip rather than adopting another library or starting from nothing.**

- The stock textarea turns every tab into four spaces on every input path, with no option, and nobody upstream has proposed one. The hard-coded `maxLines = 10000` is still there on `main`, and still not configurable.
- bubbles#1030 is open, unreviewed by any maintainer, and mergeable against `main`. Its `LineHighlighter` hook is the right shape for chroma, but it does nothing about tabs, and it has sat since August with no maintainer response.
- No other Go TUI editing component is on Bubble Tea v2 with a permissive licence, a maintainer, tab preservation, and a highlighting hook. The ones that keep tabs are tcell widgets or applications; the one Bubble Tea component with a highlighting hook also replaces tabs, has no licence, and targets Bubble Tea v1.
- A fork of v2.2.1 keeps the wrapping, cursor, viewport, selection, key map, styles, and 3,100 lines of tests, and changes four things: the sanitiser option (`runeutil.ReplaceTabs` already exists, it's just not reachable), a tab-aware width function in the wrap and cursor math, `maxLines` as a field, and the #1030 highlighter patch, which applies cleanly to v2.2.1 because `main` has no textarea changes since that tag.

## Findings

### Tabs in the Bubbles v2 textarea (v2.2.1 and `main`)

Every path that puts text into the textarea goes through `insertRunesFromUserInput`, which calls the rune sanitiser first ([textarea.go#L543-L547](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L543-L547)):

```go
// Clean up any special characters in the input provided by the
// clipboard. This avoids bugs due to e.g. tab characters and
// whatnot.
runes = m.san().Sanitize(runes)
```

- `SetValue` calls `Reset` then `InsertString`; `InsertString` and `InsertRune` call `insertRunesFromUserInput`; a typed key with `Text` and a `pasteMsg` call it too ([#L524-L541](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L524-L541), [#L1433-L1442](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L1433-L1442)). So loading, typing, and pasting all sanitise.
- The sanitiser is `runeutil.NewSanitizer()` with no options, and its default is `replaceTab: []rune("    ")`, four spaces ([textarea.go#L858-L865](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L858-L865), [internal/runeutil/runeutil.go#L23-L33](https://github.com/charmbracelet/bubbles/blob/v2.2.1/internal/runeutil/runeutil.go#L23-L33)). An option `runeutil.ReplaceTabs(string)` exists, but `runeutil` is an `internal` package, the `rsan` field is unexported, and `san()` never passes options. There is no way to keep tabs from outside the module.
- The Tab key itself never reaches the buffer. Bubble Tea v2 leaves `Key.Text` empty for `KeyTab` ([bubbletea key.go#L297-L301](https://github.com/charmbracelet/bubbletea/blob/v2.1.0/key.go#L297-L301)), and the textarea's default branch does nothing when `Text` is empty. None of the 32 bindings in `textarea.KeyMap` is bound to tab. TuiSnip's `indentedLine` in `internal/adapters/tui/editoverlay/content_line.go` inserts four spaces itself.
- `main` is two commits ahead of v2.2.1 and both are dependency bumps; the compare lists no change under `textarea/` or `internal/runeutil/` ([compare v2.2.1...main](https://github.com/charmbracelet/bubbles/compare/v2.2.1...main), head `c3bde0d1` dated 2026-10-01). The sanitiser call is at the same line on `main`.
- Nothing upstream proposes keeping tabs. The GitHub search for `textarea tab` over issues, PRs, and discussions returns no feature request, and the open-PR list for textarea (17 PRs on 2026-10-15) has none about tabs. The only related item is [bubbles#263](https://github.com/charmbracelet/bubbles/issues/263) (open since 2022-10-02, labelled bug): control characters, "most noticeable when inserting a Tab character", break width computation and can panic cursor navigation. The sanitiser is the fix that was chosen for it. That matters for a fork: `uniseg.StringWidth("\t")` and `ansi.StringWidth("\t")` are both 0 (checked locally with `rivo/uniseg` and `charmbracelet/x/ansi` from the module graph), so a tab let through today's wrap and cursor code would be placed as zero cells wide.

The spec's claim in `docs/spec/ui.md` ("replaces every tab with four spaces, on typing, pasting, and loading alike, and has no option to keep them") holds for v2.2.1 and `main`.

### The line cap

- `maxLines = 10000` is a package constant ([textarea.go#L36-L37](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L36-L37)), with the comment `// XXX: in v2, make max lines dynamic and default max lines configurable.` It is the same on `main`.
- `insertRunesFromUserInput` drops lines past the cap ([#L581-L585](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L581-L585)), `New` and `Reset` allocate `value` with capacity `maxLines`, and the wrap cache is sized to it.
- It came from [bubbles#672](https://github.com/charmbracelet/bubbles/pull/672) (2024-11-14, merged), which "hard codes the maximum lines to 10000 as a stop gap" after [#671](https://github.com/charmbracelet/bubbles/issues/671) (a large JSON file was cut because `MaxHeight` doubled as the line limit). The PR's v2 to-do list, "The maximum lines should be configurable", was not done in v2.0.0 through v2.2.1.
- The configurable caps are different things: `MaxHeight` (default 99, TuiSnip sets 0) bounds the visible rows and, when `MaxContentHeight` is 0, blocks input once the logical line count reaches it; `MaxContentHeight` (v2.1.0, with `DynamicHeight`) bounds visual rows ([#L310-L332](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L310-L332)). `CharLimit` defaults to 0 (no limit). Neither lifts `maxLines`.
- Related: [bubbles#1023](https://github.com/charmbracelet/bubbles/issues/1023) (open, 2026-07-30) reports that the textarea re-renders all text on every update and performs badly with a lot of it. `view()` iterates every logical line and `Update` calls `view()` on every message ([#L1453-L1455](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go#L1453-L1455)). A fork that lifts the cap inherits this.

The spec's claim that the textarea "would cut it at 10,000 lines" holds.

### bubbles#1030: syntax highlighting in the textarea

[bubbles#1030](https://github.com/charmbracelet/bubbles/pull/1030), "feat(textarea): add an optional line highlighter", by joestump.

- **State on 2026-10-15:** open, not a draft, `mergeable: MERGEABLE`, `mergeStateStatus: CLEAN`. No labels, no review decision, no maintainer comment. Opened 2026-08-04; the only activity since is a "COMMENTED" review on 2026-08-26 from `joestump-agent`, the author's own agent account. The PR body says it was posted on the author's behalf by an AI model.
- Its companion feature request [bubbles#1043](https://github.com/charmbracelet/bubbles/issues/1043) (2026-08-26, labelled enhancement) has no comments. Charm's [CONTRIBUTING.md](https://github.com/charmbracelet/.github/blob/main/CONTRIBUTING.md) asks for a Discussion before a feature PR; the author filed an issue instead and says so in the PR. No maintainer has engaged with either.
- An earlier attempt by a maintainer, [bubbles#657](https://github.com/charmbracelet/bubbles/pull/657) "textarea(v2): highlight, formatter, suggestions" (caarlos0, 2024-11-01), was closed unmerged on 2026-05-16 with no comments.
- **Approach:** 507 additions over `textarea/highlighter.go`, `textarea/highlighter_test.go`, and `textarea/textarea.go`. It adds

  ```go
  type LineHighlighter interface {
      Highlight(lineIdx int, line []rune) []lipgloss.Range
  }
  func (m *Model) SetHighlighter(h LineHighlighter)
  ```

  The renderer asks the hook once per logical line per render, gets rune-offset ranges, intersects them with each wrapped segment, normalises (sorts and de-overlaps) them, and renders the segment with `lipgloss.StyleRanges` ([highlighter.go on the PR branch](https://github.com/joestump/bubbles/blob/feat/textarea-line-highlighter/textarea/highlighter.go)). The cursor cell and a selected segment are never range-styled. With no highlighter set, the render path is unchanged. It also fixes `wrappedBase` drifting by one rune per wrapped segment whose trailing space was trimmed, a bug that already affects selection spans.
- `lipgloss.StyleRanges`, `NewRange`, and `Range` exist in lipgloss v2.0.6 ([ranges.go](https://github.com/charmbracelet/lipgloss/blob/v2.0.6/ranges.go)); `StyleRanges` cuts by grapheme cluster through `ansi.Cut` and "Ranges should not overlap".
- **Fit for chroma:** a chroma `Token` is `{Type, Value}` with no offset ([lexer.go#L83-L86](https://github.com/alecthomas/chroma/blob/v2.27.0/lexer.go#L83-L86)), so an adapter tokenises the whole Fragment, walks the tokens accumulating rune offsets, splits them at newlines, and maps each line's tokens to ranges styled from a chroma style entry. The hook's per-line, per-render contract means the adapter should cache the token ranges by content and re-tokenise on change.
- **What it does not do:** nothing about tabs, nothing about `maxLines`. A sibling PR by the same author, [#1031](https://github.com/charmbracelet/bubbles/pull/1031) `ByteOffset`/`SetCursorByteOffset` (also open, unreviewed), is unrelated to either.

### Candidate components

| Component | Framework | Licence | Last push | Keeps tabs | Highlighting | Verdict |
|---|---|---|---|---|---|---|
| [charmbracelet/bubbles `textarea`](https://github.com/charmbracelet/bubbles/tree/v2.2.1/textarea) v2.2.1 | Bubble Tea v2 | MIT | 2026-10-01 | No, four spaces | No; #1030 open | Fork base |
| [allisonhere/ripple](https://github.com/allisonhere/ripple) | Bubble Tea v1 (bubbletea v1.3.10, bubbles v1.0.0) | None (no LICENSE file, GitHub reports none) | 2026-08-14 | No, four spaces | Yes, `StyleKey`/`Style` hook | Rejected |
| [rivo/tview `TextArea`](https://github.com/rivo/tview/blob/master/textarea.go) | tcell | MIT | 2026-08-11 | Yes, rendered as `TabSize` spaces | No, whole-widget `SetTextStyle` only | Rejected |
| [pgavlin/femto](https://github.com/pgavlin/femto) | tview/tcell | MIT | 2023-02-25 | Yes (micro buffer, `tabsize` setting) | Yes, micro syntax files | Rejected |
| [micro-editor/micro](https://github.com/micro-editor/micro) | tcell | MIT | 2026-10-14 | Yes | Yes | Not importable |
| [gopad-dev/gopad](https://github.com/gopad-dev/gopad) | Bubble Tea v2 alpha (2024-10 pseudo-version, pre-`charm.land`) | Apache-2.0 | 2024-12-03 | Unverified | Yes, tree-sitter | Rejected |
| [charmbracelet/crush](https://github.com/charmbracelet/crush) | Bubble Tea v2 | FSL-1.1-MIT | 2026-10-15 | No (stock textarea) | Only in read-only views | No component to take |
| [mistakenelf/teacup](https://github.com/mistakenelf/teacup) | Bubble Tea v1 | MIT | 2024-03-08 | n/a | `code` bubble is a chroma viewer | Not an editor |

Details:

- **ripple** is the only Bubble Tea component found with a highlighting hook: `Options.StyleKey func(offset int) string` classifies runes and `Options.Style func(key, text string) string` renders each run, layered under selection and cursor ([ripple.go](https://github.com/allisonhere/ripple/blob/main/ripple.go), `Options` struct). But its `sanitizeInput` writes `"    "` for every `\t`, the same loss as Bubbles (the doc comment above it says "a tab jumps to the next tab stop"; the code does not). Its `go.mod` requires the v1 modules, the repo has no licence file, one star, and the `View(opts Options) string` signature with func-typed styles doesn't match the Bubbles `Styles` pattern TuiSnip's `input` package maps onto. Rejected on licence and tabs alone.
- **tview `TextArea`** is the one maintained widget that stores `\t` ("Tab: Insert a tab character (\t). It will be rendered like [TabSize] spaces", [textarea.go#L131-L132](https://github.com/rivo/tview/blob/master/textarea.go#L131-L132)). Its only text styling is `SetTextStyle`, `SetSelectedStyle`, and `SetPlaceholderStyle`, all whole-widget `tcell.Style`. It draws to a `tcell.Screen`, so it can't be composed into a Bubble Tea `View() string`. Its tab rendering and cursor mapping over tabs are worth reading when writing TuiSnip's own.
- **femto** extracts micro v1.4.1's buffer and cell view as a tview widget. It keeps tabs (`buffer.go` reads the `tabsize` setting), highlights with micro's regex syntax definitions, and depends on `zyedidia/micro v1.4.1` (2019). `go 1.12`, last push 2023-02-25, README: "The shape of the femto API is a work-in-progress". tcell-bound and dormant.
- **micro** keeps its buffer, display, and actions under `internal/`, so they can't be imported. Only `pkg/highlight` (its regex highlighter) is public. TuiSnip already has chroma, so there is nothing to take.
- **gopad** is an application. Its `go.mod` pins `github.com/charmbracelet/bubbletea/v2 v2.0.0-alpha.1.0.20241017211530-...` (the pre-release import path, superseded by `charm.land/bubbletea/v2`), replaces lipgloss and bubblezone with its own forks, and parses with `go.gopad.dev/go-tree-sitter`, whose repository carries `allocator.c`/`allocator.h` and a `src/` tree, so it is a CGO binding (the toolchain already rejects tree-sitter bindings for CGO). Last push 2024-12-03.
- **crush** (Charm's own agent) uses the stock textarea v2.2.1 for its prompt and runs chroma only on read-only content (`internal/ui/common/highlight.go` formats with `terminal16m`). Its licence is FSL-1.1-MIT, not MIT. There is no reusable editor in it, which is itself a data point: Charm hasn't needed in-textarea highlighting for their flagship app.
- The GitHub repository searches (`bubbletea text editor`, `topic:bubbletea editor`, `"syntax highlighting" editor bubbletea`, language Go) otherwise return applications with embedded editors (justwrite, atlas.ed, committed, mdtt, lazysql), none packaged as a component and none on `charm.land` v2.

### What a fork of the stock textarea keeps

`textarea/` at v2.2.1 is 2,115 lines in `textarea.go`, 248 in `selection.go`, and 3,166 lines of tests. Worth keeping as is:

- **Wrapping and layout:** `wrap` (word wrap over rune slices, width from `uniseg`), `memoizedWrap` with `internal/memoization` (a fork copies that package), `LineInfo` (row and column offsets of the cursor within wrapped lines), `repositionView`, and the `viewport` the textarea embeds. This is also where the open upstream bugs sit (#231, #333, #566, #567, #887 are open issues about reflow and cursor placement; #913 and #1051 are open fix PRs), and a fork inherits them along with the code.
- **Cursor:** `Cursor() *tea.Cursor` for the real terminal cursor, and the `bubbles/cursor` virtual cursor with blink handling, plus `SetVirtualCursor`.
- **Editing and key map:** the 32-binding `KeyMap` with `DefaultKeyMap()`, word movement and deletion, transpose, case changes, page movement, and `Paste`/`CopySelection` commands (TuiSnip disables the clipboard ones in `input.NewContentArea`).
- **Selection** (`selection.go`, v2.2.0): keyboard and pointer selection, `BeginSelection`/`ExtendSelection`/`EndSelection`, `PositionAt`, `SelectedText`, `DeleteSelection`. TuiSnip's `content_line.go` already relies on `ClearSelection` and `SetCursorColumn`.
- **Styles:** `Styles`/`StyleState` with `DefaultStyles(isDark)`, prompt and `SetPromptFunc`, `ShowLineNumbers`, placeholder, `EndOfBufferCharacter`, `DynamicHeight`/`MinHeight`/`MaxContentHeight`. `internal/adapters/tui/input/styles.go` maps `look.Styles` onto these today, so the adapter code survives a fork unchanged.
- **Tests:** `textarea_test.go` (2,883 lines) and `selection_test.go` pin the behaviour above and would catch regressions from the tab change.

What changes:

1. `san()` builds the sanitiser with `runeutil.ReplaceTabs("\t")` (the option exists; the fork owns `runeutil` so it can be reached).
2. A display-width function that expands a tab to the next tab stop replaces the bare `uniseg.StringWidth` calls in `wrap`, `LineInfo`, `Cursor`, and `view`, with the same width used in all four. bubbles#263 is the record of what goes wrong when they disagree. Rendering also has to emit spaces for the tab, since the terminal's own tab behaviour isn't what the viewport expects.
3. `maxLines` becomes a field (the upstream `XXX` comment asks for exactly that); the `value` capacity and wrap-cache size follow it. #1023's full re-render per update stays; whether it is acceptable at TuiSnip's Fragment sizes is a prototype question.
4. The #1030 patch (`highlighter.go`, the `view()` hunks, and the `wrappedBase` fix) on top, with a chroma adapter outside the component.

`textinput` builds its own sanitiser with `ReplaceTabs(" ")` and `ReplaceNewlines(" ")` ([textinput.go#L299-L303](https://github.com/charmbracelet/bubbles/blob/v2.2.1/textinput/textinput.go#L299-L303)); a fork that changes only `textarea`'s `san()` leaves single-line inputs alone.

## Alternatives rejected

- **Waiting for upstream.** Neither tabs nor the line cap has an open proposal, and the one highlighting PR has had no maintainer response in ten weeks; the maintainer's own earlier attempt was closed unmerged after eighteen months. There is nothing to wait for.
- **Adopting ripple, tview, femto, or gopad.** See the table: v1 and no licence; tcell; tcell and dormant; CGO and dormant.
- **Writing from scratch.** The stock component's wrap, cursor, viewport, and selection code carries four years of fixes and 3,100 lines of tests, all MIT. The tab change touches four functions in it.
- **Post-processing the rendered `View()` string with chroma.** The PR body for #1030 names why: the renderer has already emitted ANSI for cursor, selection, and line styles, so a second pass has to parse escapes it didn't write.

## Open questions for the human

- Should the fork live under `internal/adapters/tui/` as a package (`go-arch-lint` would need a component for it) or as a vendored module directory? The former keeps the standard-library-only core rule and the existing `tui` layer edges.
- Tab width: a fixed 4, or a per-Language setting? The prototype ticket on the map is the place to decide; this research only establishes that the width must be one function used by wrap, cursor, and render alike.
- Whether to send the tab and `maxLines` changes upstream as separate PRs once they exist. The `XXX` comment and #672's to-do list suggest a configurable line cap would be accepted; tabs would need a Discussion first per CONTRIBUTING.md.

## Sources

- Bubbles textarea v2.2.1: https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go, https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/selection.go, https://github.com/charmbracelet/bubbles/blob/v2.2.1/internal/runeutil/runeutil.go
- Bubbles `main` vs v2.2.1: https://github.com/charmbracelet/bubbles/compare/v2.2.1...main
- Bubbles releases: https://github.com/charmbracelet/bubbles/releases (v2.1.0 dynamic height, v2.2.0 selection, v2.2.1 word-left fix)
- Bubble Tea v2 key semantics: https://github.com/charmbracelet/bubbletea/blob/v2.1.0/key.go
- lipgloss v2 ranges: https://github.com/charmbracelet/lipgloss/blob/v2.0.6/ranges.go
- chroma Token: https://github.com/alecthomas/chroma/blob/v2.27.0/lexer.go
- bubbles#1030: https://github.com/charmbracelet/bubbles/pull/1030; branch source: https://github.com/joestump/bubbles/tree/feat/textarea-line-highlighter/textarea
- bubbles#1043: https://github.com/charmbracelet/bubbles/issues/1043
- bubbles#657: https://github.com/charmbracelet/bubbles/pull/657
- bubbles#263: https://github.com/charmbracelet/bubbles/issues/263
- bubbles#671, #672: https://github.com/charmbracelet/bubbles/issues/671, https://github.com/charmbracelet/bubbles/pull/672
- bubbles#1023: https://github.com/charmbracelet/bubbles/issues/1023
- bubbles#1031: https://github.com/charmbracelet/bubbles/pull/1031
- Charm CONTRIBUTING: https://github.com/charmbracelet/.github/blob/main/CONTRIBUTING.md
- ripple: https://github.com/allisonhere/ripple (go.mod, README.md, ripple.go)
- tview TextArea: https://github.com/rivo/tview/blob/master/textarea.go
- femto: https://github.com/pgavlin/femto (README.md, go.mod, buffer.go)
- micro: https://github.com/micro-editor/micro (tree: `internal/`, `pkg/highlight`)
- gopad: https://github.com/gopad-dev/gopad/blob/main/go.mod, https://github.com/gopad-dev/go-tree-sitter
- crush: https://github.com/charmbracelet/crush (go.mod, LICENSE.md, internal/ui/common/highlight.go)
- teacup: https://github.com/mistakenelf/teacup
- TuiSnip: `docs/spec/ui.md` "Content with tabs", `internal/adapters/tui/input/input.go`, `internal/adapters/tui/editoverlay/content_line.go`
