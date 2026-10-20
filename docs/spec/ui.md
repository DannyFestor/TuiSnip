# Main screen and navigation

How the main screen is laid out and how the keyboard and mouse move through it. Domain terms come from [`GLOSSARY.md`](../../GLOSSARY.md). Binding names and default keys are in [config](config.md#bindings); what the screens do is in [the v1 spec](v1.md). The throwaway prototype this was decided on lives on the `prototype/main-screen-layout` branch (variant D).

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
 Copied y Copy · e edit · E external editor · n new · m move · c duplicate · d delete · ? help
```

- The main screen has four Panes, numbered left to right: **1 Folders** above **2 Tags** in the left column, then **3 Snippet list**, then **4 Snippet pane**. Each Pane's title carries its number.
- **Widths** are proportions of the terminal width: the left column about 22%, the Snippet list about 30%, and the Snippet pane the rest.
- **The focused Pane grows by 8 columns.** They come from the Snippet pane, or from the other two columns when the Snippet pane has focus.
- **The left column** splits its height two-thirds to one-third. The tall Pane is the focused one of Folders and Tags, or the one holding the Browse selection when focus is elsewhere. Folders is tall at start-up.
- **The Snippet list title** shows the Browse selection's path (`Root / go`) or Tag (`# go`) and the sort order: `by title`, `by last updated`, or `by creation date`. Cycling the order keeps the cursor on the selected Snippet. With the Trash as the Browse selection the title is `Trash`, with no sort order.
- **The `▪ Trash` row** is always the last row of the Folders Pane, after the tree, with its count of Trash items (a trashed Folder counts as 1). `G`/`end` lands on it.
- **The Trash list** lists the Trash items newest deleted first. A Snippet row looks as it does in Browse. A Folder row reads `▸ docker`, with its Snippet count (`15 Snippets`) in the right-hand column. `s` doesn't apply, and the remembered Sort order is untouched.
- **A trashed Snippet in the Snippet pane** shows in full, read-only: its meta line names the path of the Folder it was in, and the dates line starts with `deleted 2026-10-14`, so a narrow Pane cuts off the other dates first. **A trashed Folder** shows its name, the path it was in, `1 subfolder · 15 Snippets`, and `deleted 2026-10-14`, then below a separator an outline of its subtree: subfolders indented as in the Folders Pane, with their Snippet titles in title order. The counts and the outline leave out Snippets and subfolders that are separate Trash items. The outline scrolls like Content.
- **The Snippet list's right-hand column** shows the first Fragment's Language, followed by the count of the other Fragments when there are any: `Go +2`.
- **The Snippet pane** shows one Fragment at a time, the Shown Fragment, and its meta line names that Fragment's Language. A tab strip under the header separator labels each Fragment by its Fragment title, or by its Language when it has none, and marks the Shown Fragment. The strip shows when the Snippet has two or more Fragments, or when its only Fragment has a title. A strip wider than the Pane is cut off with `…`. Showing a different Snippet goes back to its first Fragment, unless Search revealed it on another.

```
│Deploy                                  │
│Root / go · YAML                        │
│created 2026-09-06 · updated 2026-09-27 │
│────────────────────────────────────────│
│ main.go │ YAML │ README                │
│   1 │ on: push                         │
```
- **The status line** under the Panes shows the last message on the left and the focused Scope's main Bindings on the right (see [Status hint](#status-hint)).
- **Below 80×24** TuiSnip shows only the focused Pane, as when zoomed, with a hint that the terminal is too small for all four. Focus still moves between Panes.
- The proportions, the growth, and the minimum size are named constants in `tui`.

### Browse selection

The Browse selection is the Folder, the Root, the Tag, or the Trash whose contents the Snippet list shows. The Trash is entered and left like a Folder.

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
| Folders | `N` new Folder inside the selected one (at the Root when the Root is selected), typed in place on a new row below the cursor · `r` rename in place · `d` delete · `m` move · `space` collapse · `L` Default Language |
| Tags | `N` new Tag, typed in place; it exists with a count of 0 until a Snippet carries it · `r` rename in place · `d` delete |
| Snippet list | `y` Copy · `e` edit · `E` external editor · `m` move · `c` duplicate · `d` delete · `s` cycle sort |
| Snippet list, with the Trash as the Browse selection (`trash` Scope) | `y` Copy · `r` restore · `d` delete permanently · `D` empty Trash |
| Snippet pane | `y` Copy · `e` edit · `E` external editor · `w` wrap · `]` next Fragment · `[` previous Fragment |

- In-place entry commits with `enter` and cancels with `esc` (the `name_input` Scope). A blank name is refused. A new Folder becomes the Browse selection once it is created.
- A new Tag is typed on a row placed where the name will sort among the Tags, in the order the Tags Pane lists them. The row moves as the name is typed, and an empty name sorts first. The new Tag becomes the Browse selection once it is created. A name that already exists, compared case-insensitively, is refused with "Tag go already exists", which names the existing Tag with its own spelling, not the typed one. A comma is refused with "Tag name contains a comma", and a blank name with "Tag name is blank". A refused name stays in the row to be fixed.
- Renaming a Tag onto another Tag's name merges the two. The other Tag survives with the typed spelling, every Snippet that carried either Tag carries it once, and the cursor moves onto it.
- Deleting a Tag no Snippet carries happens at once. Otherwise a confirmation titled "Delete Tag" asks first, default No: `Permanently delete Tag "go" and remove it from 4 Snippets? This cannot be undone. [y/N]`. A count of one reads "1 Snippet". After a delete the cursor keeps its row. When the deleted Tag held the Browse selection, the Tag now under the cursor takes it, or the Folders take it back when no Tag is left.
- `c` in the Snippet list copies the Snippet under the cursor into the same Folder, with the same Title, Description, and Tags, and every Fragment with its title, content, and Language, in the same order. The cursor moves onto the copy.
- `y` and `E` in the Snippet list and the Snippet pane act on the Shown Fragment. `e` edits the first Fragment, and an external edit returns to the edit overlay on the Fragment it edited.
- `]` and `[` in the Snippet pane show the next and the previous Fragment, wrapping around at either end, and scroll back to the top.
- `d` in the Snippet list moves the Snippet to the Trash at once: `Moved Snippet "Prune everything" to the Trash`. `d` in Folders does the same for the Folder and its subtree: `Moved Folder "docker" to the Trash`, and the Folder's parent becomes the Browse selection. After a delete the cursor keeps its row: the Snippet below takes it, or the one above when the last row was deleted.
- `N`, `r`, `m`, `space`, `L`, and `d` do nothing on the `▪ Trash` row.
- In the Trash, `y` copies the Shown Fragment of a trashed Snippet. On a trashed Folder it does nothing, and the status line says "Only a Snippet can be copied". `r` restores the Trash item under the cursor ([v1: Rename and delete](v1.md#rename-and-delete)). The cursor keeps its row, the Browse selection stays on the Trash, and the restored Trash item isn't revealed.
- `d` in the Trash deletes the Trash item under the cursor for good, after a confirmation titled "Delete permanently", default No: `Permanently delete Snippet "Prune everything"? This cannot be undone. [y/N]`, or `Permanently delete Folder "docker", 1 subfolder, and 15 Snippets? This cannot be undone. [y/N]`, which leaves out a zero count: `Permanently delete Folder "docker" and 15 Snippets? This cannot be undone. [y/N]`. Then it says `Permanently deleted Snippet "Prune everything"` or `Permanently deleted Folder "docker"`. The cursor keeps its row, as after a restore. `D` empties the Trash after a confirmation titled "Empty Trash", default No: `Permanently delete 3 Folders and 20 Snippets in the Trash? This cannot be undone. [y/N]`, then says `Emptied the Trash: 2 Folders, 20 Snippets` ([v1: Rename and delete](v1.md#rename-and-delete)). Both do nothing in an empty Trash.
- A trashed Snippet in the Snippet pane keeps the `snippet_pane` Bindings for Copy, wrap, `]`/`[`, and scrolling. `e` and `E` do nothing, and the status line says "Restore the Snippet to edit it". On a trashed Folder, `y` in the Snippet pane says "Only a Snippet can be copied", as in the list. Beyond that, a trashed Folder's outline only scrolls and zooms, so the [status hint](#status-hint) and [Help](#help) leave out the `snippet_pane` Bindings there.
- `m` in Folders or the Snippet list opens the [Folder picker](#other-overlays) for the row under the cursor.
  - A moved Folder keeps the cursor and the Browse selection in its new place, and its new parent expands if it was collapsed. When a Tag held the Browse selection, the Folders take it back.
  - A moved Snippet leaves the Browse selection as it is. When the Snippet leaves the listed Folder or the Root, the cursor keeps its row as after a delete. When it is still listed, as under a Tag, the cursor stays on it and only its Folder path changes.
- `L` opens the Language picker, titled "Default Language of" and the Folder's name, on the Default Language of the Folder under the cursor. The picked Language becomes that Folder's Default Language for the Snippets created in it from then on. Snippets already in it keep their Language. `L` does nothing on the Root, which is always `plaintext`.
- `n` (new Snippet) and `p` (Capture) work from every Pane and use the Browse selection:
  - A Folder or the Root: the Snippet goes there, with that Folder's Default Language.
  - A Tag: the Snippet goes to the Root and already carries the Tag, so it shows up in the list being looked at.
  - The Trash: the Snippet goes to the Root, in plain text. On save, the Root becomes the Browse selection.
- `w` wraps long lines in the Snippet pane, and pressing it again cuts them off at the Pane's edge. A wrapped line keeps its number on its first row, and each continuation row shows `↪` in the line-number gutter, so the gutter keeps its width and the `│` separator lines up.

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

- The left 40% holds the query and the results, each with its Folder path. The right side previews the highlighted result the way the Snippet pane shows it, on the Fragment the result opens on ([v1.md: Search](v1.md#search)).
- An empty query lists the Snippets of the current Browse selection, in the current sort order.
- `↑` `↓` (also `ctrl+p` `ctrl+n`, `ctrl+k` `ctrl+j`) move through the results. Everything else typed goes to the query.
- `enter` reveals the Snippet: the Browse selection becomes its Folder or the Root, the Snippet list selects it, focus moves to the Snippet pane, and the pane shows the previewed Fragment.
- `ctrl+y` copies the previewed Fragment and closes the popup. `copy.quit_after` applies as for any Copy.
- `esc` closes the popup and changes nothing.

## Edit overlay

`e`, `n`, `p`, and a returning external editor open the edit overlay. It covers about 90% of the screen, and the main screen stays visible around it. While it is open the Panes don't respond: save or cancel first.

```
╭ Editing • ─────────────────────────────────────────────────────────────╮
│  Title          Deploy                                                 │
│  Description    Build and ship                                         │
│› Tags           #ci   (enter or ctrl+t to edit)                        │
│   main.go │ YAML │ README                                              │
│  Fragment title main.go                                                │
│  Language       Go   (enter or ctrl+l to pick)                         │
│  Content        enter or ↓ to edit                                     │
│┃   1 package main                                                      │
│┃                                                                       │
│ctrl+s save · esc cancel · ↑/↓ field · ctrl+t Tags · ctrl+l Language    │
╰────────────────────────────────────────────────────────────────────────╯
```

- Fields, top to bottom: Title, Description, Tags, Fragment title, Language, Content. The overlay opens on Title. `•` in the title marks unsaved changes.
- **Fragments.** Fragment title, Language and Content edit the Shown Fragment. The overlay opens on the first Fragment, or on the Fragment an external edit returned. Above the Fragment title row sits the same tab strip as in the Snippet pane, shown for two or more Fragments or for one titled Fragment. The Fragment title row is always shown. A typed Fragment title relabels its tab as it is typed, and a Fragment without one is labelled by its Language.
  - `alt+right` / `alt+left` switch to the next / previous Fragment, wrapping around at either end, and a click on a tab switches to it. Each Fragment keeps its unsaved edits while another is shown.
  - `ctrl+n` adds an empty, untitled Fragment right after the Shown Fragment and switches to it. Its Language is the Default Language of the Snippet's Folder, or of the Folder a new Snippet goes to. The Fragment title never sets the Language.
  - `ctrl+x` removes the Shown Fragment after a confirmation titled "Remove Fragment", default No: `Remove Fragment "main.go"? [y/N]`, naming the Fragment by its tab label. The last Fragment can't be removed: the status line says "A Snippet needs a Fragment; delete the Snippet instead".
  - `alt+shift+left` / `alt+shift+right` move the Shown Fragment one place left / right. They do nothing at either end.
  - These keys work outside the Content textarea. `esc` leaves the textarea first.
  - Saving writes every Fragment in the order the strip shows. Cancelling discards every Fragment change, after the usual [y/N].
  - A save refused for a Fragment's field, such as a Fragment title longer than 200 characters ("Fragment title is longer than 200 characters"), switches to that Fragment and marks the field.
- `↓` / `↑` (also `tab` / `shift+tab`) move between fields.
- `enter` on a text field moves to the next field. On Tags it opens the Tag editor, on Language the Language picker, and on Content it enters the textarea.
- `ctrl+t` and `ctrl+l` open the Tag editor and the Language picker from any field.
- **Inside Content:**
  - `esc` leaves the textarea but keeps the overlay open, and a second `esc` cancels.
  - `↑` on the first line moves to Language.
  - `tab` indents the cursor's line by four spaces. `shift+tab` removes one leading tab from a line that starts with a tab, and otherwise up to four leading spaces. Either way the cursor stays on the same character.
  - **A paste that would make Content longer than 10,000 lines** isn't inserted, because the textarea would cut it at 10,000 lines. The count is the lines Content would have after the paste: its current lines, minus the line breaks in a selection the paste replaces, plus the line breaks in the paste, since the paste's first line joins the cursor's line. The textarea turns every carriage return into a line break, so each `\r` counts as one, and a `\r\n` line ending counts as two. A paste that brings Content to exactly 10,000 lines goes in. The status line says "Paste would make Content longer than 10,000 lines; use ctrl+e to edit in $EDITOR", naming the first key the user configured for `open_in_editor` in the `content` Scope (`ctrl+e` by default). With no key bound it says only "Paste would make Content longer than 10,000 lines".
- `ctrl+s` saves, closes the overlay, selects the Snippet in the list, and shows it in the Snippet pane. `esc` outside Content cancels, asking [y/N] first if anything changed.
- **A save over a Snippet that another TuiSnip changed** is refused. A confirmation titled "Changed elsewhere" asks "This Snippet changed in another TuiSnip. Reload it and discard your changes? [y/N]", default No.
  - Yes closes the overlay, discards the edits, selects the Snippet in the list, and shows the stored version in the Snippet pane. `e` edits it again.
  - No keeps the overlay open with the edits. Every save stays refused, asking again, until the Snippet is reloaded.
- **A save whose Snippet or Folder is gone**, moved to the Trash or deleted in another TuiSnip, asks in a confirmation titled "Gone elsewhere", default No. An edit asks "This Snippet was moved to the Trash or deleted in another TuiSnip. Save your changes as a new Snippet at the Root? [y/N]". A New Snippet or Capture into a Folder that is gone asks `Folder "docker" was moved to the Trash or deleted in another TuiSnip. Save the Snippet at the Root instead? [y/N]`.
  - Yes saves everything the overlay holds as a new Snippet at the Root, closes the overlay, makes the Root the Browse selection with the cursor on the new Snippet, and says `Saved "Prune everything" as a new Snippet at the Root`.
  - No keeps the overlay open with the edits. The next save asks again.
- `ctrl+e` opens the Shown Fragment in the external editor from any field. The temp file is named from its Language.
- **Content over 10,000 lines** is read-only, because the textarea would cut it at 10,000 lines. That covers a stored Snippet opened with `e`, and externally edited content, which lands as an unsaved change. Lines are counted the way a paste into empty Content is, so exactly 10,000 lines stay editable. The Content field says "Longer than 10,000 lines: read-only here, edit with ctrl+e ($EDITOR)", naming the first key the user configured for `open_in_editor` in the `content` Scope (`ctrl+e` by default). With no key bound it says only "Longer than 10,000 lines: read-only here". Saving keeps the content byte for byte. When the external editor returns it with 10,000 lines or fewer, it is editable again.

### Tabs in Content

Content keeps its tab characters, and Content with tabs is editable in the Content textarea.

- **Tab stops** sit every 4 cells: a tab runs to the next multiple of 4 cells, so `a\tb` shows `b` at cell 4. On a soft-wrapped row the cells count from the start of the line, not of the row. The Content textarea, the Snippet pane, the Search preview, and the Trash preview all use the same tab stops.
- **Tabs draw as spaces**, with no visible mark.
- **The cursor** on a tab covers the whole tab.
- **Movement:** one `←` / `→` press moves over a tab. `↑` / `↓` keep the cell column and land on the tab covering it.
- **Pasted, captured, loaded, and externally edited tabs** stay tabs. The `tab` key still indents with four spaces ([Edit overlay](#edit-overlay)).

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

- The top line names the Tags this Snippet carries, in the Tags Pane's order. Every Tag is listed with its Snippet count, and ✓ marks the Tags this Snippet carries. "No Tags yet." shows until a Tag exists.
- Typing filters the list. When the typed name matches no Tag exactly, compared case-insensitively, a `+ create "…"` row closes the list.
- `enter` toggles the highlighted Tag, or creates and adds the typed name on the create row. Names are trimmed, and a comma is refused with "Tag name contains a comma" on the status line.
- A created Tag is listed with `new` in place of its count. Toggled off, it stays listed until the Tag editor closes.
- `esc` closes the Tag editor and keeps the toggles as unsaved changes in the edit overlay, and `ctrl+s` saves them with the rest. A created Tag is stored with the Snippet in the same save, so cancelling the edit overlay leaves no Tag behind.

## Other overlays

The Language picker, the Folder picker, and the Tag editor share one shape: a filter line on top and a list below, moved with the same keys as the Search popup. `enter` picks and closes, except in the Tag editor, where it toggles. `esc` closes.

Typing filters the list, ignoring case and the space around the filter: the entries that start with the filter come first, then the ones that contain it, each in list order.

- The **Language picker** shows the curated `languages` list when config sets one, in the config's order, and `ctrl+a` switches to every Language and back, keeping the filter. Every Language is listed by name, ignoring case. The picker opens on the current Language, and says "No Language matches." when the filter matches none. From the edit overlay it is titled "Pick a Language", and the picked Language is an unsaved change.
- The **Folder picker** is titled "Move", the Snippet's title or the Folder's name, and "to". It lists the Root and every Folder by its path, and opens on the Folder the Snippet or Folder is in. It greys out a moving Folder's own subtree, which `enter` doesn't pick.
- **Confirmations** (default No) open over whatever is showing. A confirmation shows the user's configured keys: the first `yes` key and the first `no` key in brackets, `[y/N]` by default. The No key is upper-cased to mark it as the default only when it is one printable character, so a first No key of `shift+tab` shows as written. A question wider than the screen wraps onto more rows.
- Overlays stack: the Tag editor, the Language picker, and a confirmation can open over the edit overlay.

## Help

`?` (`help`) opens help centred over the main screen. It opens from the main screen only: with all four Panes showing, with a Pane zoomed, and below 80×24. While an overlay is open or a Folder or Tag name is typed in place, the key belongs to that Scope, so `?` types a `?` into a text field and does nothing elsewhere.

```
╭ Help ────────────────────────────────────╮
│q              quit                 y Copy│
│?              help                 e edit│
│/              search               s sort│
│z              zoom                       │
│n              new                        │
│p              Capture                    │
│tab            next Pane                  │
│shift+tab      previous Pane              │
│l, right       Pane to the right          │
│h, left        Pane to the left           │
│1              Folders                    │
│2              Tags                       │
│3              Snippet list               │
│4              Snippet pane               │
│enter          open                       │
│esc            back                       │
│j, down        down                       │
│k, up          up                         │
│g, home        first row                  │
│G, end         last row                   │
│pgdown, ctrl+d page down                  │
│pgup, ctrl+u   page up                    │
╰──────────────────────────────────────────╯
```

- Help lists the Bindings active on the main screen: the `global` Bindings in the left column, and the focused Pane's Scope in the right. On a trashed Folder in the Snippet pane the right column is empty, because the outline takes no `snippet_pane` Bindings, only the `global` ones such as scrolling and zoom.
- Each row shows every key the user configured for the Binding, joined with `, `, then the Binding's label. A Binding with no keys is left out.
- Help closes with its own `help` key or the `global` `back` key, `?` and `esc` by default. Every other key does nothing.
- A terminal too small for help cuts it to fit: a column that doesn't fit is dropped, and rows past the bottom are cut off.

Labels are fixed. The `global` Bindings are labelled:

| Binding | Label |
|---|---|
| `quit` | quit |
| `help` | help |
| `search` | search |
| `zoom` | zoom |
| `new_snippet` | new |
| `capture` | Capture |
| `focus_next` / `focus_prev` | next Pane / previous Pane |
| `focus_right` / `focus_left` | Pane to the right / Pane to the left |
| `focus_folders` `focus_tags` `focus_list` `focus_snippet` | Folders, Tags, Snippet list, Snippet pane |
| `open` / `back` | open / back |
| `down` / `up` | down / up |
| `top` / `bottom` | first row / last row |
| `page_down` / `page_up` | page down / page up |

A pane Scope's Bindings carry the labels the [Status hint](#status-hint) gives them, and `collapse` is labelled collapse.

## Status hint

The right end of the status line lists the focused Scope's main Bindings with the user's actual keys. Labels are fixed. When the line is too narrow, entries drop from the right, but `?` help always stays.

| Scope | Hint |
|---|---|
| `folders` | open · new Folder · rename · delete · move · zoom · search · help |
| `tags` | open · new Tag · rename · delete · zoom · search · help |
| `snippet_list` | Copy · edit · external editor · new · move · duplicate · delete · sort · zoom · search · help |
| `trash` | Copy · restore · delete permanently · empty Trash · zoom · search · help |
| `snippet_pane` | Copy · edit · external editor · wrap · next Fragment · previous Fragment · zoom · search · help |
| `snippet_pane`, on a trashed Folder | zoom · search · help |
| `editor` | save · cancel · field · Tags · Language · external editor · next Fragment · previous Fragment · add Fragment · remove Fragment · move Fragment left · move Fragment right |
| `content` | save · leave · indent · dedent · external editor |
| `search` | move · reveal · Copy · close |
| `picker` | move · pick · close |
| `picker`, in the Tag editor | move · toggle / create · close |
| `name_input` | save · cancel |
| `confirm` | yes · no |
| help (no Scope) | close · close, naming the first `help` key and the first `back` key: `? close · esc close` by default |

## Mouse

On by default. `mouse = false` in config turns it off ([config](config.md#keys)).

| Action | Effect |
|---|---|
| Click in a Pane | focuses it |
| Click a row | selects it, as moving the cursor there would; in Folders or Tags it becomes the Browse selection |
| Click a label in the Snippet pane's tab strip | shows that Fragment |
| Click a label in the edit overlay's tab strip | makes that Fragment the Shown Fragment |
| Click a Folder's `▸`/`▾` | collapses or expands it, and selects it as a click on its row would |
| Double-click a row | same as `enter` |
| Wheel | scrolls the Pane under the pointer without moving focus |
| Click an item in an overlay | picks it (toggles it in the Tag editor) |
| Click outside an overlay | same as `esc`, so a stray click only ever cancels |
| Click a field in the edit overlay | focuses the field; a click below the field rows enters Content. The textarea cursor is not placed by clicking in v1 |

While the mouse is on, the terminal's own click-and-drag selection needs a modifier: Shift in most Linux terminals, Option in iTerm2, fn in Terminal.app. The README says so.

## Empty main screen

With no Snippets in the Browse selection, the Snippet list shows "No Snippets here." followed by the main Bindings: new Snippet, Capture, Search, new Folder, help. With no Snippets at all, the Snippet pane shows the same hint. An empty Trash shows "The Trash is empty." in the Snippet list and the Snippet pane, with no Binding hints.
