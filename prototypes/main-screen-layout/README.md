# PROTOTYPE: main screen layout and navigation

Throwaway. Answers [Main screen layout and navigation](https://github.com/DannyFestor/TuiSnip/issues/16). Lives only on the `prototype/main-screen-layout` branch.

Three structurally different main screens over the same fake in-memory data. Nothing is saved, copied, or written.

```sh
go -C prototypes/main-screen-layout run .   # starts on D; -variant=A|B|C|D, -empty for the empty-screen hint
```

`` ` `` and `~` cycle variants at any time outside text entry. The yellow bar is prototype chrome and shows the active Scope and state.

| | Layout | Navigation | Sidebar | Search |
|---|---|---|---|---|
| A | three fixed columns | `tab`/`shift+tab` cycle panes, `enter` moves right | one cursor across Folders and Tags | input atop the Snippet list |
| B | sidebar, list stacked above Snippet pane | `h`/`l` hop panes, `enter` moves right | tabs, `tab` switches Folders ⇄ Tags | input atop the Snippet list |
| D (revision) | Folders over Tags, list, Snippet pane; the active left pane is tall; the focused pane widens a little | `tab`/`shift+tab` cycle all four; `h`/`l`/`←`/`→` treat Folders and Tags as one column (left returns to the one filling the list); `1`–`4` jump; `enter` drills in, `esc` backs out | separate stacked panes, each keeps its cursor; focus alone never changes the list, moving the cursor or `enter` does | centred popup: results left, highlighted preview right; `enter` reveals the Snippet in its Folder |
| C | focused pane widens | `enter`/`l` drill in, `esc`/`h` back out, `1`/`2`/`3` jump | Folders and Tags stacked, `tab` hops | centred palette with preview; `enter` reveals the Snippet in its Folder |

Shared by all: edit mode (`e`; `↑`/`↓` between fields, `esc` leaves Content before cancelling, `tab`/`shift+tab` indent and dedent inside Content), overlays (`?` help, `m` Folder picker, `d` confirmations, `ctrl+l` Language picker), `z` maximize, `w` wrap, in-place Folder/Tag rename (`r`) and new Folder (`N`).
