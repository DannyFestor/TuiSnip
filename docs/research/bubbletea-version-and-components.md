# Bubble Tea, Bubbles, Lip Gloss: version and component pattern

Research for [#2](https://github.com/DannyFestor/TuiSnip/issues/2). Checked 2026-09-30 against the Go module proxy, the upstream repos, and their release notes.

## Question

Which major versions of Bubble Tea, Bubbles, and Lip Gloss should TuiSnip build on? What is the current idiomatic way to build small, reusable, independently testable components (model composition, message routing, `key.Binding`, teatest)? How do chroma-highlighted Fragments render in a textarea or viewport, and what does the `$EDITOR` shell-out look like?

## Recommendation

- **Use the v2 line of all three libraries, under the `charm.land` vanity paths:** `charm.land/bubbletea/v2` v2.0.10, `charm.land/bubbles/v2` v2.2.1, `charm.land/lipgloss/v2` v2.0.6. v1 gets no more development. Bubbles v2 needs Bubble Tea v2 and Lip Gloss v2, so all three upgrade together.
- **Use `github.com/charmbracelet/x/exp/teatest/v2` for the e2e tier.** It has no tagged releases (pseudo-versions only), so pin one commit and let Dependabot bump it.
- **Component pattern:**
  - Each TUI component is a concrete struct with `Update(tea.Msg) (T, tea.Cmd)` and `View() string`.
  - Only the root model implements `tea.Model` and returns a `tea.View`.
  - The root sends key messages to the focused child only. It broadcasts sizing and tick messages to every child and combines their commands with `tea.Batch`.
  - Each component owns a `KeyMap` struct of `key.Binding` values that it builds from config, and implements `help.KeyMap`.
- **Highlighting:** tokenise with chroma, format with `terminal16m`, and hand the ANSI string to `viewport.SetContent`. Bubble Tea v2 downsamples colour automatically. Pick the chroma style from `tea.BackgroundColorMsg.IsDark()`.
- **The Bubbles textarea cannot highlight syntax today.** It has no per-token styling hook, and the upstream PR that would add one ([bubbles#1030](https://github.com/charmbracelet/bubbles/pull/1030)) is open and unreviewed. See Open questions.
- **`$EDITOR`:** write the Fragment to a temp file, build the command with `github.com/charmbracelet/x/editor` (it splits `EDITOR="code -w"` into command and arguments), run it with `tea.ExecProcess`, and read the file back in the callback message.

## Findings

### Versions (as of 2026-09-30)

| Module | Latest | Released | Source |
|---|---|---|---|
| `charm.land/bubbletea/v2` | v2.0.10 | 2026-09-24 | https://proxy.golang.org/charm.land/bubbletea/v2/@latest, https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.10 |
| `charm.land/bubbles/v2` | v2.2.1 | 2026-08-24 | https://proxy.golang.org/charm.land/bubbles/v2/@latest, https://github.com/charmbracelet/bubbles/releases/tag/v2.2.1 |
| `charm.land/lipgloss/v2` | v2.0.6 | 2026-08-11 | https://proxy.golang.org/charm.land/lipgloss/v2/@latest, https://github.com/charmbracelet/lipgloss/releases/tag/v2.0.6 |
| `github.com/charmbracelet/x/exp/teatest/v2` | `v2.0.0-20260927004216-9c77d672503d` (no tags) | 2026-09-27 | https://proxy.golang.org/github.com/charmbracelet/x/exp/teatest/v2/@latest |
| `github.com/charmbracelet/x/editor` | v0.2.0 | 2026-01-09 | https://proxy.golang.org/github.com/charmbracelet/x/editor/@latest |
| `github.com/alecthomas/chroma/v2` | v2.27.0 | 2026-06-17 | https://proxy.golang.org/github.com/alecthomas/chroma/v2/@latest |

- None of `charm.land/{bubbletea,bubbles,lipgloss}/v3` exist on the proxy. v2 is the current major.
- Bubble Tea v2.0.0 was released on 2026-02-24 and has had ten patch releases since: https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.0
- The last v1 releases are bubbletea v1.3.10 (2025-09-17) and lipgloss v1.1.0 (2025-03-12). Bubbles v1.0.0 (2026-02-09) calls itself "just an honorary release of Bubbles v1": https://proxy.golang.org/github.com/charmbracelet/bubbletea/@latest, https://github.com/charmbracelet/bubbles/releases/tag/v1.0.0
- The module paths moved to the vanity domain (`charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`): https://github.com/charmbracelet/bubbletea/blob/v2.0.10/UPGRADE_GUIDE_V2.md#import-paths
- "Bubbles v2 requires Bubble Tea v2 and Lip Gloss v2. Upgrade all three together": https://github.com/charmbracelet/bubbles/blob/v2.2.1/UPGRADE_GUIDE_V2.md
- Minimum Go versions on the upstream `main` branches are 1.26.0 (bubbletea), 1.25.0 (bubbles), 1.26.7 (lipgloss), and 1.25 (chroma v2.27.0). The map's Go 1.27+ covers all of them: https://github.com/charmbracelet/bubbletea/blob/main/go.mod, https://github.com/charmbracelet/lipgloss/blob/main/go.mod

### v2 API facts that shape components

- `tea.Model.View()` returns `tea.View`, not `string`. Terminal features are declared as fields on the View (`AltScreen`, `MouseMode`, `Cursor`, `WindowTitle`, `KeyboardEnhancements`, ...), not set with startup options or commands: https://github.com/charmbracelet/bubbletea/blob/v2.0.10/UPGRADE_GUIDE_V2.md#the-big-idea-declarative-views
- Key presses arrive as `tea.KeyPressMsg`. `tea.KeyMsg` is now an interface that covers both press and release. Space stringifies as `"space"`: https://github.com/charmbracelet/bubbletea/blob/v2.0.10/UPGRADE_GUIDE_V2.md#key-messages
- Paste arrives as its own message types, not as a flag on a key message. This matters for Language guessing on paste: https://github.com/charmbracelet/bubbletea/blob/v2.0.10/UPGRADE_GUIDE_V2.md#paste-messages
- Lip Gloss v2 is "pure" and leaves terminal I/O to Bubble Tea. Bubble Tea downsamples any ANSI colour to the detected profile automatically: https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.0
- `AdaptiveColor` is gone. Components take an explicit `isDark`. The recommended source for it is `tea.RequestBackgroundColor` in `Init` followed by `tea.BackgroundColorMsg.IsDark()`: https://github.com/charmbracelet/bubbles/blob/v2.2.1/UPGRADE_GUIDE_V2.md#4-light-and-dark-styles
- Bubbles components expose getter/setter methods for size (`SetWidth`, `SetHeight`) instead of exported fields. `DefaultKeyMap` is now a function that returns a fresh value instead of a mutable global: https://github.com/charmbracelet/bubbles/blob/v2.2.1/UPGRADE_GUIDE_V2.md#2-global-patterns
- Bubbles components still render `View() string`. Only the program's root returns `tea.View`: https://github.com/charmbracelet/bubbles/blob/v2.2.1/viewport/viewport.go, https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go

### Composition and message routing

- In the official `split-editors` example:
  - Child textareas are concrete values updated with `m.inputs[i], cmd = m.inputs[i].Update(msg)`, and the commands are combined with `tea.Batch`.
  - Children are sized from `tea.WindowSizeMsg` through `SetWidth`/`SetHeight`.
  - Bindings are toggled with `key.Binding.SetEnabled`.
  - The root composes child `View()` strings with `lipgloss.JoinHorizontal`.

  https://github.com/charmbracelet/bubbletea/blob/v2.0.10/examples/split-editors/main.go
- The `composable-views` example switches on a focus state to decide which child receives key messages, and routes each child's own tick messages to that child: https://github.com/charmbracelet/bubbletea/blob/v2.0.10/examples/composable-views/main.go
- The real terminal cursor belongs to the root. `textarea.Model.Cursor()` returns a `*tea.Cursor` relative to the textarea, and the root has to offset it by the component's position before setting `View.Cursor`: https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go, https://github.com/charmbracelet/bubbletea/blob/v2.0.10/examples/split-editors/main.go

### Keybindings

- `key.NewBinding(key.WithKeys(...), key.WithHelp(...))`, `SetKeys`, `SetEnabled`, `Unbind`, and a generic `key.Matches[Key fmt.Stringer](k Key, b ...Binding)`. Because `SetKeys` works at runtime, bindings can be built from the TOML config: https://github.com/charmbracelet/bubbles/blob/v2.2.1/key/key.go
- `help.KeyMap` is `ShortHelp() []key.Binding` plus `FullHelp() [][]key.Binding`. A component's `KeyMap` struct that implements it gets the help bar for free: https://github.com/charmbracelet/bubbles/blob/v2.2.1/help/help.go

### Testing (teatest v2)

- teatest v2 lives at `github.com/charmbracelet/x/exp/teatest/v2`, a separate module inside the `x` monorepo under `exp/`. Its go.mod requires `charm.land/bubbletea/v2 v2.0.0-rc.1`, and MVS resolves that to whatever newer version TuiSnip requires: https://github.com/charmbracelet/x/blob/main/exp/teatest/v2/go.mod
- API:
  - `NewTestModel(tb, m, WithInitialTermSize(w,h), WithProgramOptions(...))`
  - `tm.Send`, `tm.Type`, `tm.Quit`
  - `WaitFor(tb, tm.Output(), cond, WithDuration, WithCheckInterval)`
  - `tm.FinalModel(tb)` for type-asserting on final state
  - `tm.FinalOutput(tb)` plus `RequireEqualOutput` for golden files, updated with `-update`. It uses the system `diff`.
  - The default size is 80x24.

  https://github.com/charmbracelet/x/blob/main/exp/teatest/v2/teatest.go
- The x repo's README lists teatest under `exp`, i.e. experimental with no compatibility promise: https://github.com/charmbracelet/x/blob/main/README.md
- Components with the `Update/View string` shape can be unit-tested without teatest: call `Update` with a constructed `tea.KeyPressMsg` and assert on the returned value, the returned Cmd's message, or `View()`. This follows from the value-returning signatures above. Keep teatest for the e2e tier, as the map says.

### Chroma rendering

- `quick.Highlight` shows the pipeline: get the lexer (`lexers.Get`, falling back to `lexers.Analyse`, then `lexers.Fallback`), then `chroma.Coalesce`, `formatters.Get`, `styles.Get`, and `Tokenise`/`Format` into an `io.Writer`: https://github.com/alecthomas/chroma/blob/v2.27.0/quick/quick.go
- Terminal formatters are `terminal`, `terminal8`, `terminal16`, `terminal256`, and `terminal16m`: https://github.com/alecthomas/chroma/blob/v2.27.0/formatters/tty_indexed.go, https://github.com/alecthomas/chroma/blob/v2.27.0/formatters/tty_truecolour.go
- Paired light/dark styles ship with chroma (`github`/`github-dark`, `catppuccin-latte`/`catppuccin-mocha`): https://github.com/alecthomas/chroma/tree/v2.27.0/styles
- `viewport.SetContent` accepts ANSI-styled strings. v2 adds `SoftWrap`, `LeftGutterFunc` (line numbers), `SetHighlights` (useful for Search match marking), and `StyleLineFunc`: https://github.com/charmbracelet/bubbles/blob/v2.2.1/UPGRADE_GUIDE_V2.md#viewport. The official `glamour` example renders pre-styled ANSI in a viewport: https://github.com/charmbracelet/bubbletea/tree/v2.0.10/examples/glamour
- The textarea can only style whole regions: `StyleState` has `Base`, `Text`, `LineNumber`, `CursorLine`, `Placeholder`, `Prompt`, and `Selection`, with no per-token hook. It stores the value as runes and styles each line with a single `Text` style, so feeding it ANSI would corrupt the buffer: https://github.com/charmbracelet/bubbles/blob/v2.2.1/textarea/textarea.go
- Upstream has an open feature request, [bubbles#1043](https://github.com/charmbracelet/bubbles/issues/1043) (2026-08-26), and an open PR, [bubbles#1030](https://github.com/charmbracelet/bubbles/pull/1030) (2026-08-04, no review yet). Together they propose `SetHighlighter(LineHighlighter)`, where `Highlight(lineIdx int, line []rune) []lipgloss.Range` is intersected with soft-wrap segments at render time.
- Lip Gloss v2 already has `lipgloss.Range`, `NewRange`, and `StyleRanges`, so chroma token offsets map directly onto that proposed hook: https://github.com/charmbracelet/lipgloss/blob/v2.0.6/ranges.go

### `$EDITOR` shell-out

- `tea.ExecProcess(c *exec.Cmd, fn ExecCallback) Cmd` pauses the Program, gives the terminal to the child, and resumes afterwards. `ExecCallback` is `func(error) Msg`. `tea.Exec` takes any `ExecCommand`: https://github.com/charmbracelet/bubbletea/blob/v2.0.10/exec.go
- The official example is `exec.Command(cmp.Or(os.Getenv("EDITOR"), "vim"))` passed through `ExecProcess`, returning an `editorFinishedMsg{err}`: https://github.com/charmbracelet/bubbletea/blob/v2.0.10/examples/exec/main.go. That call does not split `EDITOR="code -w"`.
- `x/editor.Command(app, path, opts...)` splits `$EDITOR` with `strings.Fields`, defaults to `nano`, and supports `LineNumber`/`AtPosition` for vi/vim/nvim/nano/emacs/kak/gedit/code/hx. It refuses to run under Snap. It does not read `$VISUAL`: https://github.com/charmbracelet/x/blob/main/editor/editor.go

## Alternatives rejected

- **Bubble Tea/Bubbles/Lip Gloss v1.** Development has stopped (last bubbletea v1 release 2025-09-17), the Bubbles v1.0.0 release notes point to v2, and new Bubbles features (viewport gutter and soft wrap, textarea selection) are v2-only. Starting a new project on v1 means a guaranteed migration later.
- **teatest v1 (`github.com/charmbracelet/x/exp/teatest`).** It targets Bubble Tea v1.
- **Raw `exec.Command(os.Getenv("EDITOR"))` as in the upstream example.** It breaks on `EDITOR` values that include flags (`code -w`, `subl -w`), which are common on macOS. `x/editor` fixes this and adds jump-to-line.
- **Post-processing `textarea.View()` output to inject chroma colours.** bubbles#1043 describes why this fails: it breaks at soft-wrap boundaries, on wide or ZWJ graphemes, and where it collides with the virtual cursor.
- **Styling through the `compat.HasDarkBackground()` package.** It does blocking I/O outside the event loop (https://github.com/charmbracelet/bubbles/blob/v2.2.1/UPGRADE_GUIDE_V2.md#quick-use-compat-package). `tea.BackgroundColorMsg` is the supported path.

## Open questions for the human

1. **Highlighting while editing.** The map settles on "in-TUI textarea with syntax highlighting (chroma) as default", but the stock textarea cannot do that yet. Options:
   - (a) v1 shows highlighted Fragments in a viewport and edits them in a plain textarea, with the editing component behind a TuiSnip-owned interface so the highlighter can be added when bubbles#1030 merges.
   - (b) Vendor or fork the MIT textarea into `internal/` with the #1030-style `LineHighlighter` applied, and own that code.
   - (c) Depend on the PR branch through a `replace` directive.

   Which one?
2. **`$VISUAL`.** Should `$VISUAL` take precedence over `$EDITOR`, as the Unix convention has it? `x/editor` ignores `$VISUAL`, so honouring it would need a small wrapper.
3. **teatest pinning.** The recommendation puts an unversioned `exp/` module in the e2e tier, pinned to one pseudo-version and bumped by Dependabot. Is that acceptable, or should the e2e tier use a different test harness?
