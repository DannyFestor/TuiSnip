# Roadmap

What TuiSnip does in v1 is fixed in [`docs/spec/v1.md`](docs/spec/v1.md). This file orders what comes after. The first three milestones exist on GitHub; later sections become milestones as they come closer.

## v0.1 Walking skeleton

A tooled repo where one Snippet can be created at the Root, found by Search, and copied, all through the TUI.

## v1.0

Everything in the v1 spec.

## v1.1

- Trash. Deleted Snippets and Folders go to the Trash and can be restored. Deleting a Folder moves its whole subtree there.
- JSON export and import, as a backup.

## v1.2

- CLI subcommands as thin adapters over the application layer: `tuisnip get` to print a Snippet, `tuisnip add` to create one from stdin.

## v1.3

- FTS5 full-text search behind the existing Search port.
- Search scoped to the current Folder or Tag.
- `#tag` filters in the query.

## v2.0

- Several Fragments per Snippet, shown as re-orderable tabs.

## Later

Unscheduled.

- Syntax highlighting while editing in the TUI, once the Bubbles textarea supports it (bubbles#1030).
- Language guessing via chroma: the title's file extension (`lexers.Match`), then content analysis (`lexers.Analyse`). Needs a research ticket on accuracy for short Snippets first.
- Settings screen. The preferred design writes a separate `settings.toml` layered over the hand-edited `config.toml`, so the app never modifies a file a human edits. Comment-preserving writes to `config.toml` are the alternative. Research before deciding.
- Remembered UI state in the state file: maximized pane, last selected Folder.
- Favorites and pins.
- SnippetsLab importer.
- Homebrew tap.
- Colour themes beyond the built-in light and dark schemes.
- Mouse support.
- Multi-select for bulk move, tag, and delete.
