# Configuration

TuiSnip reads one hand-edited TOML file, `config.toml`, and keeps remembered UI state in a separate file it owns. This document is the reference for both. Domain terms come from [`CONTEXT.md`](../../CONTEXT.md). The default file, with a comment on every key, is [`embeds/config/default.toml`](../../embeds/config/default.toml).

## Location

| File | Path | Owner |
|---|---|---|
| Config | `$XDG_CONFIG_HOME/tuisnip/config.toml`, falling back to `~/.config/tuisnip/config.toml` | the user |
| State | `$XDG_STATE_HOME/tuisnip/state.toml`, falling back to `~/.local/state/tuisnip/state.toml` | TuiSnip |

The same paths apply on macOS and Linux. v1 has no `--config` flag and no environment variable of its own. Tests point `XDG_*` at a temporary directory. `tuisnip --paths` prints both paths.

## Lifecycle

- On every start, if `config.toml` is missing, TuiSnip creates the directory (`0755`) and writes the embedded default file byte for byte (`0644`), comments included.
- It creates the file with `O_CREATE|O_EXCL`. When two instances start at once, the loser reads the winner's file.
- TuiSnip never modifies an existing `config.toml`. Changes take effect on the next start.

## Loading

1. Decode the embedded default file.
2. Decode the user's file over it. A key missing from the user's file keeps its default without a warning, so a file written by an older version keeps working.
3. Validate the result. If anything is invalid, refuse to start.

The embedded file is the only place defaults live. A unit test asserts that it decodes and validates with no warnings.

### Validation

Validation runs in two passes, so the user sees every mistake from one start.

1. **Decode** into a raw struct of TOML primitives: strings, booleans, and string arrays. Syntax errors and wrong types (`theme = 3`) stop here, one at a time, with the line.
2. **Parse** each raw value into its type: go-enum `Parse…` functions for enums, `value.NewLanguage` for Languages, the key-string validator for Bindings. Every error is collected with `errors.Join` and named by its key path.

On failure TuiSnip prints every error to stderr and exits non-zero:

```
tuisnip: invalid config ~/.config/tuisnip/config.toml:
  theme: "blue" is not one of auto, light, dark
  languages[2]: "golang" is not a Language
  bindings.snippet_list.copy: "shift+y" never matches; write "Y"
  bindings.global.search: "/" is also bound to bindings.snippet_list.edit
```

### Unknown keys

- An unknown key outside `[bindings.*]` is logged at Warn and ignored. That lets an older build read a file written for a newer one.
- An unknown scope or Binding name inside `[bindings.*]` is an error. A misspelt Binding would otherwise silently keep its old key.

## Keys

| Key | Type | Default | Meaning |
|---|---|---|---|
| `editor` | string | `""` | External editor command. See [Editor](#editor). |
| `theme` | `auto` \| `light` \| `dark` | `auto` | `auto` follows the terminal background and uses dark if the terminal does not report one. |
| `languages` | array of Language names | `[]` | The Languages the Language picker offers. Empty means every Language. |
| `copy.clipboard` | `auto` \| `native` \| `osc52` | `auto` | How Copy reaches the clipboard. See [Clipboard](#clipboard). |
| `copy.trim_trailing_newline` | bool | `true` | Copy drops one trailing newline from the Fragment. |
| `copy.quit_after` | bool | `false` | TuiSnip quits after a successful Copy. |
| `mouse` | bool | `true` | Mouse clicks and the wheel work in the TUI. While on, the terminal's own text selection needs a modifier key. See [the UI spec](ui.md#mouse). |
| `bindings.<scope>.<name>` | array of key strings | see the default file | The keys of one Binding. See [Bindings](#bindings). |

Values an Action needs that the user doesn't set, such as a Search result limit, are named constants in `bootstrap`. They are not config.

The Snippet order is not config. The cycle key changes it, and the state file remembers it.

### Editor

TuiSnip uses the first non-empty value of:

1. `editor` in config
2. `$VISUAL`
3. `$EDITOR`
4. the first of `nvim`, `vim`, `vi`, `nano` found on `PATH`

It runs the command the way git runs `core.editor`, as `sh -c '<editor> "$@"' tuisnip <file>`. Values with arguments, such as `code --wait`, work the same from config and from the environment. If nothing resolves, the status line says "No editor found; set `editor` in config.toml". The editing flow itself is in [the v1 spec](v1.md#edit-in-an-external-editor).

### Languages

Each entry is a chroma lexer name, parsed by `value.NewLanguage`. An unknown name is an error. When the list is non-empty, the Language picker shows only those Languages and has a toggle to reveal all of them. A Fragment keeps its Language even if the list doesn't include it.

### Clipboard

- `auto` writes with the platform tool (`pbcopy`, `wl-copy`, `xclip`, `xsel`). Over SSH, or when no tool is installed, it sends OSC 52 instead.
- `native` only uses a platform tool. If none is installed, Copy fails and the status line says "No clipboard tool found (pbcopy, wl-copy, xclip, xsel)".
- `osc52` always sends OSC 52 and never runs a tool.

Capture always reads with the platform tool. OSC 52 can't read.

## Bindings

A Binding maps keys to one thing the user can do. Bindings are grouped by **Scope**, the part of the screen that has focus. Each Scope is a TOML table, and each Binding is a name mapped to an array of key strings:

```toml
[bindings.snippet_list]
copy = ["y", "ctrl+y"]
delete = []
```

- Setting a Binding replaces only that Binding's keys. The others keep their defaults.
- `[]` unbinds it.
- Arrays only. A bare string is a type error.
- Help labels are fixed in code and not configurable.

### Scopes

| Scope | Active when | Text entry |
|---|---|---|
| `global` | a Pane has focus and no overlay is open | no |
| `folders` | the Folders Pane has focus | no |
| `tags` | the Tags Pane has focus | no |
| `snippet_list` | the Snippet list has focus | no |
| `snippet_pane` | the Snippet pane has focus | no |
| `editor` | the edit overlay is open, outside the Content textarea | yes |
| `content` | the edit overlay's Content textarea is entered | yes |
| `search` | the Search popup is open | yes |
| `picker` | the Language picker, the Folder picker, or the Tag editor is open | yes |
| `confirm` | a y/N confirmation is open | no |

Help closes with its own `help` key or with `esc`, so it has no Scope.

`global` Bindings are active only alongside the four pane Scopes: `folders`, `tags`, `snippet_list`, and `snippet_pane`. Text-entry Scopes and `confirm` get none of them, so typing `q` in the editor types a `q`. How the Panes are laid out and moved between is in [the UI spec](ui.md).

### Names and defaults

| Scope | Binding: default keys |
|---|---|
| `global` | `quit` q · `help` ? · `search` / · `zoom` z · `new_snippet` n · `capture` p · `focus_next` tab · `focus_prev` shift+tab · `focus_right` l, right · `focus_left` h, left · `focus_folders` 1 · `focus_tags` 2 · `focus_list` 3 · `focus_snippet` 4 · `open` enter · `back` esc · `down` j, down · `up` k, up · `top` g, home · `bottom` G, end · `page_down` pgdown, ctrl+d · `page_up` pgup, ctrl+u |
| `folders` | `new_folder` N · `rename` r · `delete` d · `move` m · `collapse` space |
| `tags` | `new_tag` N · `rename` r · `delete` d |
| `snippet_list` | `copy` y · `edit` e · `open_in_editor` E · `move` m · `duplicate` c · `delete` d · `cycle_sort` s |
| `snippet_pane` | `copy` y · `edit` e · `open_in_editor` E · `wrap` w |
| `editor` | `save` ctrl+s · `cancel` esc · `next_field` down, tab · `prev_field` up, shift+tab · `open_field` enter · `pick_language` ctrl+l · `edit_tags` ctrl+t · `open_in_editor` ctrl+e |
| `content` | `save` ctrl+s · `leave` esc · `indent` tab · `dedent` shift+tab · `pick_language` ctrl+l · `edit_tags` ctrl+t · `open_in_editor` ctrl+e |
| `search` | `down` down, ctrl+n, ctrl+j · `up` up, ctrl+p, ctrl+k · `accept` enter · `copy` ctrl+y · `cancel` esc |
| `picker` | `down` down, ctrl+n, ctrl+j · `up` up, ctrl+p, ctrl+k · `accept` enter · `cancel` esc · `show_all_languages` ctrl+a |
| `confirm` | `yes` y · `no` n, esc, enter |

- `new_snippet` and `capture` act on the Browse selection from any Pane. A Folder or the Root receives the Snippet. With a Tag, the Snippet goes to the Root carrying that Tag.
- `rename` and `delete` act on the row under the cursor in `folders` or `tags`. `move` exists only for Folders.
- `focus_right` and `focus_left` treat Folders and Tags as one column. `open` and `back` drill in and out. The details are in [the UI spec](ui.md#keyboard-navigation).
- `up`, `down`, `top`, `bottom`, `page_up`, and `page_down` move the cursor in Folders, Tags, and the Snippet list, and scroll the Snippet pane.
- `editor.open_field` moves on from a text field, opens the Tag editor or the Language picker on those fields, and enters the textarea on Content.
- `search.accept` reveals the highlighted Snippet in its Folder and focuses the Snippet pane. `search.copy` copies it and closes the popup.
- `picker.accept` picks and closes, except in the Tag editor, where it toggles the highlighted Tag or creates the typed one.
- `picker.show_all_languages` only acts in the Language picker when `languages` is set.
- In `confirm`, `enter` picks the default, No.

### Key strings

Key strings use the names Bubble Tea v2 produces, because `key.Matches` compares strings exactly:

- one printable character: `y`, `?`, `/`, `E`
- a named key: `enter`, `esc`, `space`, `tab`, `backspace`, `delete`, `insert`, `up`, `down`, `left`, `right`, `home`, `end`, `pgup`, `pgdown`, `f1` to `f12`
- modifiers joined with `+`: `ctrl`, `alt`, `shift`, `meta`, `super`, `hyper`, for example `ctrl+s` or `shift+tab`

The validator writes modifiers in Bubble Tea's order (`ctrl+alt+shift+meta+hyper+super`), so `shift+ctrl+up` is accepted as `ctrl+shift+up`. It rejects, with a hint:

| Written | Error |
|---|---|
| `" "` | use `space` |
| `shift+e` | Bubble Tea reports a shifted letter as the capital; write `E` |
| `escape` | use `esc` |
| an unknown name | not a key |

The validator is a fuzz target alongside the config parser (`docs/standards/testing.md`).

### Conflicts

TuiSnip refuses to start when:

- two Bindings in the same Scope share a key
- a `global` Binding shares a key with a Binding in `folders`, `tags`, `snippet_list`, or `snippet_pane`
- a text-entry Scope binds a printable key without a modifier, since it would swallow typing. `esc`, `enter`, `tab`, and `ctrl+…` are fine.

The same key in two Scopes that are never active together, such as `d` in `folders` and `snippet_list`, or `save` in `editor` and `content`, is fine.

### ctrl+c

`ctrl+c` quits from every Scope and is not configurable, so a config mistake can't trap the user. It still asks for confirmation when there are unsaved changes. Binding `ctrl+c` anywhere is an error.

## State file

The state file holds what TuiSnip remembers between runs. In v1 that is two keys:

```toml
[snippet_list]
sort = "title"   # title | updated | created

[folders]
collapsed = ["0192f1d4-7b3e-7c1a-9f00-3c5e8a2b4d61"]   # ids of collapsed Folders
```

- An id in `collapsed` whose Folder no longer exists is ignored and dropped on the next write.

- TuiSnip writes it to a temporary file in the same directory and renames it over the old one. With several instances the last writer wins.
- A missing, unreadable, or invalid state file is logged at Warn and replaced by defaults. It never stops start-up, because the user doesn't edit it.
- Unknown keys are ignored.

## Implementation

- `github.com/BurntSushi/toml` decodes both files. `MetaData.Undecoded()` lists unknown keys without failing the decode, which is what the Warn rule needs. It is granted to `config` and `state` in `.go-arch-lint.yml`.
- `embeds/config` embeds `default.toml`. Only `internal/adapters/config` imports it, under an alias, because both packages are named `config`.
- `internal/adapters/config` produces a plain `config.Config`. `bootstrap` converts it into `tui.Settings`, one `KeyMap` per Scope plus the theme and the other settings. `tui` never imports `config`.
- Scope and Binding names, the theme, the clipboard backend, and the sort order are go-enum types with `--marshal`.
