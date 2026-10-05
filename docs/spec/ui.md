# Main screen and navigation

How the main screen is laid out and how the keyboard and mouse move through it. Domain terms come from [`CONTEXT.md`](../../CONTEXT.md). Binding names and default keys are in [config](config.md#bindings); what the screens do is in [the v1 spec](v1.md). The throwaway prototype this was decided on lives on the `prototype/main-screen-layout` branch (variant D).

## Layout

```
╭ 1 Folders ──────────╮╭ 3 Root / go · by title ───╮╭ 4 Snippet ─────────────────────────────╮
│◆ Root              2││errors.Join a slice      Go││Graceful HTTP shutdown                  │
│▾ docker            2││Graceful HTTP shutdown   Go││Root / go · Go · #go                    │
│    compose         1││                           ││created 2026-09-06 · updated 2026-09-27 │
│▾ go                2││                           ││Stop accepting, drain, exit             │
│    testing         1││                           ││────────────────────────────────────────│
│  shell             2││                           ││   1 │ func run(ctx context.Context, …  │
│                     ││                           ││   2 │     errc := make(chan error, 1)  │
╰─────────────────────╯│                           ││   3 │     go func() { errc <- srv.Li…  │
╭ 2 Tags ─────────────╮│                           ││   4 │                                  │
│# docker            3││                           ││   5 │     select {                     │
│# go                4││                           ││   6 │     case err := <-errc:          │
│# unused            0││                           ││   7 │         return err               │
╰─────────────────────╯╰───────────────────────────╯╰────────────────────────────────────────╯
 Copied                                         y copy · e edit · n new · s sort · z zoom · / search · ? help
```

- The main screen has four Panes, numbered left to right: **1 Folders** above **2 Tags** in the left column, then **3 Snippet list**, then **4 Snippet pane**. Each Pane's title carries its number.
- **Widths** are proportions of the terminal width: the left column about 22%, the Snippet list about 30%, and the Snippet pane the rest.
- **The focused Pane grows by 8 columns.** They come from the Snippet pane, or from the other two columns when the Snippet pane has focus.
- **The left column** splits its height two-thirds to one-third. The tall Pane is the focused one of Folders and Tags, or the one holding the Browse selection when focus is elsewhere. Folders is tall at start-up.
- **The Snippet list title** shows the Browse selection's path (`Root / go`) or Tag (`# go`) and the sort order: `by title`, `by last updated`, or `by creation date`. Cycling the order keeps the cursor on the selected Snippet.
- **The status line** under the Panes shows the last message on the left and the focused Scope's main Bindings on the right (see [Status hint](#status-hint)).
- **Below 80×24** TuiSnip shows only the focused Pane, as when zoomed, with a hint that the terminal is too small for all four. Focus still moves between Panes.
- The proportions, the growth, and the minimum size are named constants in `tui`.

### Browse selection

The Browse selection is the Folder, the Root, or the Tag whose Snippets the Snippet list shows.

- Moving focus never changes it. Tabbing through Tags on the way to the Snippet list keeps showing the Folder's Snippets.
- Moving the cursor in Folders or Tags, clicking a row there, or pressing `enter` on it makes that row the Browse selection.
- Folders and Tags each keep their own cursor, so switching between them loses neither place.

### Zoom

`z` zooms the focused Pane to fill the screen, and pressing it again restores the layout. Only one Pane shows while zoomed, so zooming Folders shows Folders without Tags. Moving focus while zoomed shows the newly focused Pane zoomed. Zoom is not remembered across restarts.

## Keyboard navigation

These Bindings live in the `global` Scope, active whenever a Pane has focus and no overlay is open.

| Keys | Effect |
|---|---|
| `tab` / `shift+tab` | next / previous Pane, in the order 1 → 2 → 3 → 4, wrapping around |
| `l` `→` / `h` `←` | Pane to the right / left. Folders and Tags count as one column: right from either goes to the Snippet list, and left from the Snippet list returns to whichever of them holds the Browse selection. No wrapping. |
| `1` `2` `3` `4` | focus that Pane directly |
| `enter` | drill in: from Folders or Tags to the Snippet list, making the row the Browse selection; from the Snippet list to the Snippet pane. Does nothing in the Snippet pane. |
| `esc` | back out: from the Snippet pane to the Snippet list, from the Snippet list to the left Pane holding the Browse selection. Does nothing in Folders or Tags. |
| `j` `↓` / `k` `↑` | move the cursor down / up; in the Snippet pane, scroll |
| `g` `home` / `G` `end` | first / last row; top / bottom in the Snippet pane |
| `pgdown` `ctrl+d` / `pgup` `ctrl+u` | a page down / up |

`space` collapses or expands the Folder under the cursor. TuiSnip remembers collapsed Folders in its [state file](config.md#state-file).

## Pane Bindings

| Pane | Bindings |
|---|---|
| Folders | `N` new Folder inside the selected one (at the Root when the Root is selected), typed in place on a new row below the cursor · `r` rename in place · `d` delete · `m` move · `space` collapse |
| Tags | `N` new Tag, typed in place; it exists with a count of 0 until a Snippet carries it · `r` rename in place · `d` delete |
| Snippet list | `y` Copy · `e` edit · `E` external editor · `m` move · `c` duplicate · `d` delete · `s` cycle sort |
| Snippet pane | `y` Copy · `e` edit · `E` external editor · `w` wrap |

- In-place entry commits with `enter` and cancels with `esc` (the `name_input` Scope). A blank name is refused. A new Folder becomes the Browse selection once it is created.
- A new Tag whose name already exists, compared case-insensitively, is refused with "Tag go already exists". A comma is refused too.
- `n` (new Snippet) and `p` (Capture) work from every Pane and use the Browse selection:
  - A Folder or the Root: the Snippet goes there, with that Folder's Default Language.
  - A Tag: the Snippet goes to the Root and already carries the Tag, so it shows up in the list being looked at.

## Search popup

`/` opens Search as a popup centred over the main screen, about 80% of its width and height.

```
╭ Search · 3 results ─────────────╮╭ Preview ──────────────────────────────╮
│/ test                           ││Table test skeleton                    │
│                                 ││Root / go / testing · Go · #go #testing│
│Table test skeleton go / testing ││───────────────────────────────────────│
│Meeting notes template      Root ││   1 │ func TestParse(t *testing.T) {  │
│Find large files           shell ││   2 │     t.Parallel()                │
╰─────────────────────────────────╯╰───────────────────────────────────────╯
```

- The left 40% holds the query and the results, each with its Folder path. The right side previews the highlighted result the way the Snippet pane shows it.
- An empty query lists the Snippets of the current Browse selection, in the current sort order.
- `↑` `↓` (also `ctrl+p` `ctrl+n`, `ctrl+k` `ctrl+j`) move through the results. Everything else typed goes to the query.
- `enter` reveals the Snippet: the Browse selection becomes its Folder or the Root, the Snippet list selects it, and focus moves to the Snippet pane.
- `ctrl+y` copies the highlighted result and closes the popup. `copy.quit_after` applies as for any Copy.
- `esc` closes the popup and changes nothing.

## Edit overlay

`e`, `n`, `p`, and a returning external editor open the edit overlay. It covers about 90% of the screen, and the main screen stays visible around it. While it is open the Panes don't respond: save or cancel first.

```
╭ Editing • ─────────────────────────────────────────────────────────────╮
│  Title       Prune everything                                          │
│  Description Reclaim disk space                                        │
│› Tags        #docker #oneliner   (enter or ctrl+t to edit)             │
│  Language    Bash   (enter or ctrl+l to pick)                          │
│  Content     enter or ↓ to edit                                        │
│┃   1 docker system prune --all --volumes --force                       │
│┃                                                                       │
│ctrl+s save · esc cancel · ↑/↓ field · ctrl+t Tags · ctrl+l Language    │
╰────────────────────────────────────────────────────────────────────────╯
```

- Fields, top to bottom: Title, Description, Tags, Language, Content. The overlay opens on Title. `•` in the title marks unsaved changes.
- `↓` / `↑` (also `tab` / `shift+tab`) move between fields.
- `enter` on a text field moves to the next field. On Tags it opens the Tag editor, on Language the Language picker, and on Content it enters the textarea.
- `ctrl+t` and `ctrl+l` open the Tag editor and the Language picker from any field.
- **Inside Content:**
  - `esc` leaves the textarea but keeps the overlay open, and a second `esc` cancels.
  - `↑` on the first line moves to Language.
  - `tab` indents and `shift+tab` removes one indent level from the cursor's line. An indent level is four spaces. `shift+tab` on a line with fewer leading spaces removes those, and the cursor stays on the same character.
  - **A paste that would make Content longer than 10,000 lines** isn't inserted, because the textarea would cut it at 10,000 lines. The count is the lines Content would have after the paste: its current lines, minus the line breaks in a selection the paste replaces, plus the line breaks in the paste, since the paste's first line joins the cursor's line. The textarea turns every carriage return into a line break, so each `\r` counts as one, and a `\r\n` line ending counts as two. A paste that brings Content to exactly 10,000 lines goes in. The status line says "Paste would make Content longer than 10,000 lines; use ctrl+e to edit in $EDITOR", naming the first key the user configured for `open_in_editor` in the `content` Scope (`ctrl+e` by default). With no key bound it says only "Paste would make Content longer than 10,000 lines".
- `ctrl+s` saves, closes the overlay, selects the Snippet in the list, and shows it in the Snippet pane. `esc` outside Content cancels, asking [y/N] first if anything changed.
- `ctrl+e` opens the Fragment in the external editor from any field.

### Content with tabs

The stock Bubbles textarea replaces every tab with four spaces, on typing, pasting, and loading alike, and has no option to keep them. Letting it edit such content would silently rewrite the indentation. In v1:

- **Content that contains a tab** shows highlighted and read-only in the Content field, with "Contains tabs: read-only here, edit with ctrl+e ($EDITOR)", naming the first key the user configured for `open_in_editor` in the `content` Scope (`ctrl+e` by default). The other fields stay editable, and saving keeps the content byte for byte.
- **The tab key in editable content** indents the cursor's line with four spaces. That is visible, so nothing is silently changed.
- **A paste containing tabs** isn't inserted. The status line says "Pasted text contains tabs; use ctrl+e to edit in $EDITOR", naming the first key the user configured for `open_in_editor` in the `content` Scope (`ctrl+e` by default). With no key bound it says only "Pasted text contains tabs".
- **Captured or externally edited content with tabs** lands read-only as an unsaved change and saves unchanged.

A TuiSnip-owned editor component that keeps tabs and highlights while editing is on the [roadmap](../../ROADMAP.md).

### Tag editor

`ctrl+t`, or `enter` on Tags, opens the Tag editor over the edit overlay.

```
╭──────────────────────────────────────────╮
│ Tags  docker, oneliner                   │
│ filter or new Tag: te                    │
│                                          │
│ ✓ testing                             1  │
│ + create "te"                       new  │
│                                          │
│ enter toggle / create · ↑/↓ move · esc   │
╰──────────────────────────────────────────╯
```

- Every Tag is listed with its Snippet count, and ✓ marks the Tags this Snippet carries.
- Typing filters the list. When the typed name matches no Tag exactly, compared case-insensitively, a `+ create "…"` row closes the list.
- `enter` toggles the highlighted Tag, or creates and adds the typed name on the create row. Names are trimmed, and a comma is refused.
- `esc` closes the Tag editor and keeps the toggles as unsaved changes in the edit overlay, and `ctrl+s` saves them with the rest.

## Other overlays

The Language picker, the Folder picker, and the Tag editor share one shape: a filter line on top and a list below, moved with the same keys as the Search popup. `enter` picks and closes, except in the Tag editor, where it toggles. `esc` closes.

- The **Language picker** shows the curated `languages` list when config sets one, and `ctrl+a` switches to every Language and back.
- The **Folder picker** includes the Root and greys out a moving Folder's own subtree.
- **Confirmations** (default No) and **help** (`?`) open over whatever is showing. A confirmation shows the user's configured keys: the first `yes` key and the first `no` key in brackets, `[y/N]` by default. The No key is upper-cased to mark it as the default only when it is one printable character, so a first No key of `shift+tab` shows as written. Help lists the Bindings active where it was opened.
- Overlays stack: the Tag editor, the Language picker, and a confirmation can open over the edit overlay.

## Status hint

The right end of the status line lists the focused Scope's main Bindings with the user's actual keys. Labels are fixed. When the line is too narrow, entries drop from the right, but `?` help always stays.

| Scope | Hint |
|---|---|
| `folders` | open · new Folder · rename · delete · zoom · search · help |
| `tags` | open · new Tag · rename · delete · zoom · search · help |
| `snippet_list` | Copy · edit · new · sort · zoom · search · help |
| `snippet_pane` | Copy · edit · wrap · zoom · search · help |
| `editor` | save · cancel · field · Language |
| `content` | save · leave · indent · dedent |
| `search` | move · reveal · Copy · close |
| `picker` | move · pick · close |
| `name_input` | save · cancel |
| `confirm` | yes · no |

## Mouse

On by default. `mouse = false` in config turns it off ([config](config.md#keys)).

| Action | Effect |
|---|---|
| Click in a Pane | focuses it |
| Click a row | selects it, as moving the cursor there would; in Folders or Tags it becomes the Browse selection |
| Click a Folder's `▸`/`▾` | collapses or expands it |
| Double-click a row | same as `enter` |
| Wheel | scrolls the Pane under the pointer without moving focus |
| Click an item in an overlay | picks it (toggles it in the Tag editor) |
| Click outside an overlay | same as `esc`, so a stray click only ever cancels |
| Click a field in the edit overlay | focuses the field; the textarea cursor is not placed by clicking in v1 |

While the mouse is on, the terminal's own click-and-drag selection needs a modifier: Shift in most Linux terminals, Option in iTerm2, fn in Terminal.app. The README says so.

## Empty main screen

With no Snippets in the Browse selection, the Snippet list shows "No Snippets here." followed by the main Bindings: new Snippet, Capture, Search, new Folder, help. With no Snippets at all, the Snippet pane shows the same hint.
