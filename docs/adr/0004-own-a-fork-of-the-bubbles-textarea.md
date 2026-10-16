# TuiSnip owns a fork of the Bubbles textarea

The Content textarea is TuiSnip's own copy of the Bubbles v2.2.1 `textarea`, in `internal/adapters/tui/textarea`, with its tests, its MIT licence notice, and its own copies of upstream's internal `runeutil` and `memoization`. The stock textarea can't keep tabs or highlight while editing, and no other component does both. Decided in [Can an existing component keep tabs and highlight while editing?](https://github.com/DannyFestor/TuiSnip/issues/205#issuecomment-6181629638) and [How the editor behaves in the edit overlay and Capture](https://github.com/DannyFestor/TuiSnip/issues/207#issuecomment-6184772817).

The fork exists for four changes against upstream. None of them is made yet. Each has its own issue:

- The sanitiser keeps tabs ([#223](https://github.com/DannyFestor/TuiSnip/issues/223)). Upstream turns every tab into four spaces on typing, pasting, `SetValue`, and `InsertString`, and nothing outside the module can pass the option that stops it.
- One tab-aware width function, used by wrap, the cursor, and rendering ([#223](https://github.com/DannyFestor/TuiSnip/issues/223)). `uniseg.StringWidth("\t")` is 0.
- `maxLines` becomes a field ([#224](https://github.com/DannyFestor/TuiSnip/issues/224)). Upstream it is a package constant of 10,000.
- The [bubbles#1030](https://github.com/charmbracelet/bubbles/pull/1030) highlighter patch, a `LineHighlighter` that returns styled ranges per line ([#225](https://github.com/DannyFestor/TuiSnip/issues/225)).

## Considered Options

- **Wait for upstream.** bubbles#1030 has had no maintainer response since August 2026, and the maintainers' own attempt (#657) was closed unmerged. No issue or PR proposes keeping tabs, and keeping them goes against the reason the sanitiser exists (bubbles#263).
- **Adopt another editor.** ripple has a highlighting hook, but it writes four spaces per tab, targets Bubble Tea v1, and has no licence. tview's `TextArea` and femto keep tabs but are tcell widgets. micro's editor packages are `internal`.
- **Write an editor.** The stock textarea already has wrapping with a cache, the viewport, the virtual cursor, selection, the KeyMap, and the `Styles` that `input/styles.go` maps onto, plus about 3,000 lines of tests. Writing that again costs more than four changes.

## Consequences

- TuiSnip opens no upstream PRs. Upstream changes are read and ported by hand.
- The fork follows the repo's lint and code standards, and meeting them changed its public API. Model methods take pointer receivers. `Selection()` returns `(Range, bool)`. The clipboard paste and copy (`Paste`, `CopySelection`, their bindings, and `Model.Err`) and `Blink` are gone, because TuiSnip receives pastes as `tea.PasteMsg` and never touches the system clipboard from the textarea. Each exported type has its own file.
- Upstream's tests now run as parallel black-box tests in `textarea_test`.
- So a diff against upstream shows all of this along with the four changes, and an upstream fix has to be rewritten for the fork. It can't be applied as a patch.
- The fork imports `github.com/mattn/go-runewidth` and `github.com/rivo/uniseg` directly, so `.go-arch-lint.yml` grants them to the `tui` component.
- The edit overlay and Capture keep the stock KeyMap, unconfigurable as before.
