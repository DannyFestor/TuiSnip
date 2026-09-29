# Clipboard strategy

Research for [#3](https://github.com/DannyFestor/TuiSnip/issues/3). Sources checked 2026-09-30.

## Question

How should TuiSnip write to the system clipboard (Copy) and read from it (pasting into a new Snippet, with Language auto-guess) on macOS and Linux? The candidates are OSC 52, shelling out to `pbcopy`/`pbpaste`, `wl-copy`/`wl-paste` and `xclip`/`xsel`, and the Go libraries `golang.design/x/clipboard` and `atotto/clipboard`.

## Recommendation

**Write path (Copy).** Put a `Clipboard` port (consumer-defined, in `app`) in front of two adapters:

1. `nativeClipboard`: shells out to a platform tool, with a `context` timeout on every call.
   - macOS: `pbcopy`. Set `LANG`/`LC_CTYPE` to a UTF-8 locale in the child environment when the parent has none. Without one, `pbcopy` falls back to the C encoding and mangles non-ASCII text.
   - Linux, Wayland (`WAYLAND_DISPLAY` set): `wl-copy`.
   - Linux, X11 (`DISPLAY` set): `xclip -in -selection clipboard`, then `xsel --input --clipboard`.
2. `osc52Clipboard`: returns Bubble Tea v2's `tea.SetClipboard(content)` and writes nothing itself.

**Read path.** No clipboard read in v1.
- Content enters a new Snippet through bracketed paste (`tea.PasteMsg`). Bubble Tea v2 enables bracketed paste by default. This works in every terminal listed below and over SSH, with no permission prompt.
- An explicit "new Snippet from clipboard" action, if the product spec wants one, reads through `nativeClipboard`: `pbpaste`, `wl-paste --no-newline`, `xclip -out -selection clipboard`, `xsel --output --clipboard`.
- OSC 52 read is never used by default. Every terminal that supports it prompts the user. Several terminals ignore it. Under default tmux it is ignored.

**Fallback order (write):**
- Remote session: `osc52Clipboard` only. A native tool would write the remote host's clipboard.
- Local session: `nativeClipboard` first. Fall back to `osc52Clipboard` when no tool is found or the tool exits non-zero or times out.
- Never emit OSC 52 alongside a successful native write. iTerm2 ships with OSC 52 off and shows a "clipboard access denied" nag on each attempt.

**Detection.** Resolve the strategy once at startup from the environment. Do not probe the terminal.
- Remote: `SSH_TTY` or `SSH_CONNECTION` is non-empty.
- Linux: pick the display from `WAYLAND_DISPLAY`, then `DISPLAY`. Pick the tool with `exec.LookPath`.
- The terminal cannot confirm OSC 52: a write gets no acknowledgement. DA1 attribute `52` is advertised inconsistently; iTerm2 advertises it even when it will prompt. So "OSC 52 is available" is an assumption, never a detected fact. After an OSC 52 Copy, the status line should say "sent to terminal" rather than "copied".

## Findings

### OSC 52 protocol

- The format is `OSC 52 ; Pc ; Pd ST`. `Pc` selects the clipboard (`c`), primary (`p`), and so on. `Pd` is base64 (RFC 4648). `Pd = ?` asks the terminal to reply with the contents. Any other non-base64 `Pd` clears the selection. xterm gates the whole sequence behind `allowWindowOps`. https://invisible-island.net/xterm/ctlseqs/ctlseqs.html (section "Operating System Commands", `Ps = 5 2`)
- The spec defines no success reply for a write, so the sender cannot tell whether a write landed. Same source.

### Terminal support matrix (defaults, current releases)

| Terminal | OSC 52 write | OSC 52 read | Source |
|---|---|---|---|
| iTerm2 | Off. Pref "Applications in terminal may access clipboard" (`AllowClipboardAccess`) defaults to `NO`. Writes are dropped, and each attempt triggers a nag. | Ask each time (`iTermPasteboardReporter` `.askEachTime`) | https://iterm2.com/documentation-preferences-general.html ; https://github.com/gnachman/iTerm2/blob/642e677c75b0/sources/Settings/iTermPreferences.m#L726 ; https://github.com/gnachman/iTerm2/blob/642e677c75b0/sources/PTYSession/PTYSession.m#L20208-L20236 ; https://github.com/gnachman/iTerm2/blob/642e677c75b0/sources/Pasting/PasteboardReporter.swift |
| Terminal.app | None (see note) | None (see note) | https://can-i-use-terminal.github.io/features/osc52copy.html (secondary; see Open questions) |
| Ghostty | `allow` | `ask` | https://ghostty.org/docs/config/reference#clipboard-read ; https://github.com/ghostty-org/ghostty/blob/4ddf1f79d349/src/config/Config.zig#L2458-L2459 |
| kitty 0.49.1 | Allowed (`write-clipboard write-primary`) | Ask (`read-clipboard-ask read-primary-ask`) | https://github.com/kovidgoyal/kitty/blob/v0.49.1/kitty/options/definition.py#L3082 ; https://sw.kovidgoyal.net/kitty/conf/#opt-kitty.clipboard_control |
| WezTerm | Allowed (no config gate in the handler) | Ignored (`QuerySelection(_) => {}`) | https://github.com/wezterm/wezterm/blob/cab25161054c/term/src/terminalstate/performer.rs#L788-L799 |
| Alacritty 0.17 | Allowed (`terminal.osc52 = "OnlyCopy"`) | Off by default (needs `CopyPaste` or `OnlyPaste`) | https://alacritty.org/config-alacritty.html |
| GNOME Terminal / VTE | Ignored (`VTE_OSC_XTERM_SET_XSELECTION` is in the unhandled case list) | Ignored | https://gitlab.gnome.org/GNOME/vte/-/blob/master/src/vteseq.cc (search `SET_XSELECTION`) ; https://gitlab.gnome.org/GNOME/vte/-/issues/2495 (open) |
| tmux 3.7c | Ignored by default. App OSC 52 is honoured only when `set-clipboard on`; the default is `external`. | Also needs `set-clipboard on`. Even then, default `get-clipboard buffer` returns tmux's newest paste buffer, not the system clipboard. | https://github.com/tmux/tmux/blob/3.7c/input.c#L3233 ; https://github.com/tmux/tmux/blob/3.7c/options-table.c#L414-L520 ; https://github.com/tmux/tmux/blob/3.7c/tmux.1#L4471-L4540 |
| tmux passthrough (`DCS tmux; … ST`) | Off by default (`allow-passthrough off`) | n/a | https://github.com/tmux/tmux/blob/3.7c/tmux.1#L5691 ; https://github.com/tmux/tmux/blob/3.7c/options-table.c |

Notes:
- Terminal.app is closed source, and Apple publishes no escape-sequence reference. The only statement found is the secondary can-i-use-terminal table, which lists it as unsupported. It was not verified here.
- iTerm2's read-side comment in `VT100Terminal.m` ("read access is not implemented due to security issues") is stale. The query is now routed to `terminalReportPasteboard` and the permission-gated `PasteboardReporter`. https://github.com/gnachman/iTerm2/blob/642e677c75b0/sources/VT100/VT100Terminal.m#L1533 and https://github.com/gnachman/iTerm2/blob/642e677c75b0/sources/VT100/VT100Terminal.m#L2710-L2720
- tmux 3.6 added OSC 52 detection through the DA1 report. tmux 3.7 added `get-clipboard`. The CHANGES entry says "the default is off", but the option table's default index is `buffer`. https://github.com/tmux/tmux/blob/master/CHANGES
- iTerm2 advertises DA1 `52` whenever the pref is on, and also when the pref was never set and it "will actually prompt". https://github.com/gnachman/iTerm2/blob/642e677c75b0/sources/PTYSession/PTYSession.m#L20208-L20223 ; https://github.com/gnachman/iTerm2/blob/642e677c75b0/sources/VT100/VT100Output.m#L1108

### Native tools

- `pbcopy`/`pbpaste` take their encoding from the locale environment. With no locale they use the C encoding, and `man 1 pbcopy` (macOS 26) recommends UTF-8.
- `wl-copy` forks and serves the data in the background by default (`--foreground` turns this off), so the clipboard outlives TuiSnip. https://github.com/bugaevc/wl-clipboard/blob/master/data/wl-clipboard.1
- Without the wlroots data-control protocol, `wl-copy` pops up a transparent surface. If the compositor does not focus it, `wl-copy` hangs, which is why every call needs a timeout. Same source, section BUGS.
- `xclip` forks into the background by default (`-silent`) and serves requests until another client takes the selection. https://github.com/astrand/xclip/blob/master/xclip.1

### Go libraries

- **Bubble Tea v2.0.10** (`charm.land/bubbletea/v2`) exposes `SetClipboard`, `SetPrimaryClipboard`, `ReadClipboard` and `ClipboardMsg`, all over OSC 52. https://github.com/charmbracelet/bubbletea/blob/v2.0.10/clipboard.go
  - It writes the sequence raw through `ansi.SetSystemClipboard`, with no tmux or screen passthrough wrapping. https://github.com/charmbracelet/bubbletea/blob/v2.0.10/tea.go#L820-L830
  - `tea.Raw` is available for a wrapped sequence (`ansi.TmuxPassthrough`) if one is ever needed. https://github.com/charmbracelet/bubbletea/blob/v2.0.10/raw.go
  - Bracketed paste arrives as `PasteMsg`. https://github.com/charmbracelet/bubbletea/blob/v2.0.10/paste.go
- **golang.design/x/clipboard v0.11.0** (2026-09-26): the premise that it needs CGO is out of date.
  - Since v0.8.0 (2026-06-07) it is Cgo-free on every desktop platform: purego on macOS, and pure-Go X11 and Wayland data-control on Linux. https://github.com/golang-design/clipboard/releases/tag/v0.8.0 ; https://github.com/golang-design/clipboard/blob/v0.11.0/README.md
  - On X11 and Wayland the writing process owns the data: "When it exits, the data goes with it, unless a clipboard manager has kept a copy". Same README, "Platform notes".
  - Wayland needs a compositor with data-control. GNOME before 49 goes through XWayland. Same README.
  - `Init` fails without a display, e.g. over SSH. Same README, "Quick start".
- **atotto/clipboard** (last commit 2026-03-26, docs only) shells out to `pbcopy`/`pbpaste` and, in order, to `wl-copy`/`wl-paste` (when `WAYLAND_DISPLAY` is set), `xclip`, `xsel` and termux.
  - It picks the tool once in `init()` and stores it in package globals.
  - It takes no context or timeout.
  - Setting the exported `Primary` global permanently truncates the arg slices.
  - It does no locale handling for `pbcopy`.
  - Sources: https://github.com/atotto/clipboard/blob/master/clipboard_unix.go and https://github.com/atotto/clipboard/blob/master/clipboard_darwin.go

## Alternatives rejected

- **OSC 52 as the primary write path.** It fails silently under stock iTerm2 (off, with a nag), Terminal.app, GNOME Terminal and default tmux, and success can never be detected. It stays as the remote and last-resort path.
- **OSC 52 read.** Every terminal that supports it prompts (iTerm2, Ghostty, kitty). WezTerm, VTE and default Alacritty ignore it. tmux returns its own buffer. It also needs an async reply with a timeout. Bracketed paste covers the use case with none of these problems.
- **golang.design/x/clipboard.** On Linux, a Copy made just before quitting TuiSnip would vanish on exit unless a clipboard manager runs. `wl-copy` and `xclip` avoid this by forking. The library also went through a large rewrite (v0.8 to v0.11 in under four months) and adds purego and an X11 codec to the dependency tree for something three small `exec` calls cover. It is worth reconsidering only if image or MIME clipboard support becomes a goal.
- **atotto/clipboard.** It runs the same tools we would run, but its package-global `init()` state, lack of a context/timeout, and missing locale handling clash with the hexagonal and testability standards. Owning the adapter behind the port costs about the same and stays fakeable.
- **Wrapping OSC 52 in tmux DCS passthrough.** It needs `allow-passthrough on`, which is off by default, just as plain OSC 52 needs `set-clipboard on`. Either way the user must configure tmux, and `set -g set-clipboard on` is the option tmux documents for this.

## Open questions for the human

1. **Terminal.app on macOS 26.** Terminal.app is closed source. Can you confirm by hand that `printf '\e]52;c;%s\a' "$(printf hi | base64)"` does nothing? It only matters for the remote-session path.
2. **Scope of remote use.** Is running TuiSnip over SSH a supported v1 scenario? If not, `osc52Clipboard` could shrink to a last-resort fallback, or be dropped.
3. **Config override.** Should `config.toml` carry a `clipboard` key (`auto` | `native` | `osc52`) so users can force a backend, e.g. inside tmux over SSH? Name and placement belong to the config and keybindings ticket.
4. **"New Snippet from clipboard".** Does the product spec want an explicit clipboard-read action, or is bracketed paste into the editor enough for v1? This decides whether the read half of `nativeClipboard` ships now.
5. **tmux guidance.** Should the README tell tmux users to `set -g set-clipboard on`? That is only needed for the OSC 52 path.
