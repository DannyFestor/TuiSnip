# Roadmap

What TuiSnip does in v1 is fixed in [`docs/spec/v1.md`](docs/spec/v1.md). This file orders what comes after, ranked by one test: does it get Danny off SnippetsLab? Next and Soon are GitHub milestones named after their theme. Later is unscheduled, and Drop records what was ruled out and why.

## Next

### Move in

Leave SnippetsLab in one move.

- Several Fragments per Snippet. Built before the SnippetsLab Import, so Import keeps every Fragment.
- Export and Import of TuiSnip's own data, as a backup.
- SnippetsLab Import, from its JSON export.

## Soon

In this order.

### Settle in

Live in TuiSnip day to day once it holds the only copy.

- Trash. Deleted Snippets and Folders go to the Trash and can be restored. Deleting a Folder moves its whole subtree there.
- A TuiSnip-owned editor component replacing the stock Bubbles textarea. It keeps tab characters, so content with tabs becomes editable in the TUI, and it highlights syntax while editing. The stock textarea turns every tab into four spaces, and upstream highlighting (bubbles#1030) is unreviewed. A fifth of the SnippetsLab library's Fragments contain tabs.

### Find it

- Search scoped to the current Folder or Tag.
- `#tag` filters in the query.
- A Language filter in the query.
- Search hits ranked by recent use.

### From the shell

CLI subcommands as thin adapters over the application layer.

- `tuisnip get` to print a Snippet.
- `tuisnip add` to create one from stdin.
- A picker that prints the chosen Snippet to stdout.
- A shell keybinding that puts a Snippet on the prompt.

## Later

Unscheduled.

- Placeholders filled in when a Snippet is used.
- Favorites and pins.
- Agent access: an MCP server, `--json` output, an Agent Skill.
- Running a Snippet from the shell.
- FTS5 trigram search for Content behind the existing Search port. Title, Description, and Tags stay fuzzy-matched with fzf ([ADR 0002](docs/adr/0002-search-splits-fuzzy-fields-from-content.md)). Only if Search is slow on the moved-in library.
- A cached in-memory Search index, refreshed after writes, if loading every Snippet on each query is too slow.
- A cap on the number of Search hits.
- Boolean, phrase, and wildcard query syntax.
- Language guessing via chroma: the title's file extension (`lexers.Match`), then content analysis (`lexers.Analyse`). Needs a research ticket on accuracy for short Snippets first.
- Settings screen. The preferred design writes a separate `settings.toml` layered over the hand-edited `config.toml`, so the app never modifies a file a human edits. Comment-preserving writes to `config.toml` are the alternative. Research before deciding.
- Remembered UI state in the state file: zoomed Pane, last Browse selection.
- Colour themes beyond the built-in light and dark schemes.
- Placing the textarea cursor by clicking.
- Multi-select for bulk move, tag, and delete.

## Drop

- Smart Groups. The SnippetsLab library has none.
- Sync through a Gist or plain files under git. TuiSnip runs on one machine, and sync would change the storage design.
- Homebrew tap. TuiSnip is a personal tool with no releases.
