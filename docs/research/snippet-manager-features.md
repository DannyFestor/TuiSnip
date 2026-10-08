# Features of other snippet managers

Research for [#177](https://github.com/DannyFestor/TuiSnip/issues/177) (map [#175](https://github.com/DannyFestor/TuiSnip/issues/175)). Sources checked 2026-10-08, primary sources only: each product's site, manual, changelog, README, and source code.

| Product | Kind | Version seen | Maintenance |
|---|---|---|---|
| SnippetsLab (SL) | macOS app | 2.7, 2026-09-17 | Active |
| Raycast Snippets (RC) | macOS/Windows/iOS launcher feature | Raycast 2.7, 2026-10-07 | Active |
| massCode (MC) | Electron app, macOS/Windows/Linux | v6.1.0, 2026-10-08 | Active, about monthly |
| Lepton (LP) | Electron GitHub Gist client | v2.0.0, 2026-07-05; v2.0.1-beta.3, 2026-10-05 | Revived in 2026 after five years without a stable release |
| Cacher (CA) | Cloud SaaS with desktop apps | Desktop 2.47.9, 2025-08-05 | Site live; CLI and plugins stale (CLI v1.1.1, 2024-09-17) |
| pet (PT) | Go CLI | v1.0.1, 2024-12-08; 21 unreleased commits on `main` | Low activity, last commit 2026-03-13 |
| nap (NP) | Go TUI | v0.1.1, 2022-11-21; 51 unreleased commits on `main` | Dormant, last commit 2023-10-30 |

For pet and nap, this document describes `main`, because the README of each tells users to install from it. The rows mark features that only exist on `main`.

## Question

Which features do comparable snippet managers offer that TuiSnip v1 lacks? Group them by theme, note how common each is, and flag what a single-user terminal tool would benefit from.

Each feature below is marked against [`docs/spec/v1.md`](../spec/v1.md) and [`ROADMAP.md`](../../ROADMAP.md):

- **has**: TuiSnip v1 does it.
- **ROADMAP**: planned, with the milestone.
- **new**: neither. "v1 non-goal" means the spec lists it as a non-goal but the roadmap doesn't schedule it.

"Count" is how many of the seven products have the feature. **Fit** is a judgement on a single-user terminal tool: **good**, **maybe**, or **poor** (desktop-only, team-only, or outside a snippet manager).

## Answer

TuiSnip v1 already matches the common core: a folder tree, tags, Description, fuzzy Search over content, highlighting, an external editor, Copy, and configurable Bindings. The roadmap covers most of what the desktop apps add next: Fragments, Trash, favourites, JSON export, a SnippetsLab importer, `get`/`add` subcommands, language guessing, `#tag` filters, and themes.

The new features worth ranking, by fit:

1. **Placeholders filled in at Copy time** (pet, Raycast; 2/7). pet's `<param=default>` with a prompt is the terminal reference. The v1 spec lists templates as a non-goal and the roadmap does not schedule them.
2. **Shell integration** (pet; 1/7, plus nap's stdout lookup). A picker that prints the chosen Snippet to stdout, a keybinding recipe that puts it on the shell prompt, and direct execution (pet `exec`, Cacher Run Server; 2/7).
3. **Agent access over MCP** (SnippetsLab 2.7, massCode; 2/7, both shipped in 2026). Search, get, and create. Neither lets an agent edit or delete. SnippetsLab's `lab` CLI also prints JSON and ships an Agent Skill file.
4. **Recently used ranking** (massCode ranks palette hits by recent use and keeps a Recent list; 1/7). A v1 non-goal.
5. **Richer query language** (SnippetsLab: AND/OR/NOT, phrases, wildcards, `in:` filters; massCode: `#tag`, `/Folder`; 2/7). `#tag` is on the roadmap. Folder and Language filters inside the query, and boolean operators, are not.
6. **Smart groups**, saved rule-based searches in the sidebar (SnippetsLab; 1/7).
7. **Sync**, by far the most common gap (5/7 sync in some form). The cheap forms are Gist sync (pet, Lepton, Cacher) and keeping the library in plain files that the user syncs with git or Syncthing (massCode, pet, nap). A v1 non-goal.
8. **Version history** (Cacher; massCode via git on its vault; 1/7 built in). A v1 non-goal.
9. **Formatting through an external formatter** (SnippetsLab: clang-format, Prettier, swift-format; massCode: Prettier; 2/7).
10. **Rendered Markdown** in notes or Markdown Snippets (4/7). In a TUI this would mean rendering the Description or a Markdown Fragment with glamour.

## Findings

### Organisation

| Feature | Who | Count | TuiSnip | Fit |
|---|---|---|---|---|
| Nested folders | SL [SL2], MC [MC2], CA nested labels [CA2]; NP flat folders [NP2] | 4 | has | |
| Tags | SL [SL2], RC [RC1], MC [MC3], LP [LP1], CA labels [CA2], PT [PT1]; NP in data model, UI key disabled [NP2] | 6 | has | |
| Per-folder default language | SL [SL2], MC [MC2] | 2 | has | |
| Description or notes | SL per fragment [SL2], MC [MC8], CA Markdown description [CA3], PT `description` [PT1] | 4 | has | |
| Sort order choice | MC [MC15 v5.7.0], LP [LP3], PT `sortby` [PT1] | 3 | has | |
| Several Fragments per Snippet | SL [SL2], MC [MC5], CA [CA3], LP (source code only) [LP5] | 4 | ROADMAP v2.0 | |
| Trash | SL [SL2], MC [MC4] | 2 | ROADMAP v1.1 | |
| Favourites or pins | SL favourites and pinned Snippets [SL2], RC pins [RC1], MC Favorites [MC4], CA star [CA5], LP pinned tags (source code only) [LP5] | 5 | ROADMAP Later | |
| Smart groups (saved rule-based searches, nested AND/OR) | SL [SL2] | 1 | new | good |
| Coloured tags | SL [SL9 2.4], LP [LP3], CA [CA2] | 3 | new | maybe |
| Hide unused tags | SL [SL2] | 1 | new | maybe |
| Lock a Snippet against edits | SL [SL9 2.4] | 1 | new | maybe |
| Manual reordering of Snippets | NP `J`/`K`, `main` only [NP2] | 1 | new | maybe |
| Folder icons | SL [SL2], MC [MC2] | 2 | new | poor |
| File attachments | SL, up to 100 per fragment [SL2], CA [CA14] | 2 | new | poor |
| Team or shared libraries | RC [RC7], CA [CA3] | 2 | new | poor |

### Retrieval

| Feature | Who | Count | TuiSnip | Fit |
|---|---|---|---|---|
| Search over title and content | all seven | 7 | has | |
| Fuzzy matching | SL toggle [SL2], MC palette [MC7], NP [NP2], PT through fzf [PT1] | 4 | has | |
| Filter by tag | SL `in:` [SL2], RC [RC1], MC `#tag` [MC7], CA [CA4], PT `--tag` [PT2] | 5 | ROADMAP v1.3 (`#tag`) | |
| Scope Search to a Folder | SL `in:` [SL2], MC `/Folder` and list search in the selected folder [MC6][MC7] | 2 | ROADMAP v1.3 (current Folder or Tag) | |
| Filter by Language in the query | SL `in:` [SL2] | 1 | new | good |
| Boolean operators, phrases, wildcards | SL `&` `\|` `!` `*` `"…"` [SL2], CA quoted exact match [CA4] | 2 | new (v1 has no query syntax) | good |
| Whole-word or case options | SL [SL2][SL9 2.5] | 1 | partly: smart case is in v1 | maybe |
| Recently used list or ranking | MC [MC7][MC15 v5.12.0] | 1 | new (v1 non-goal) | good |
| Command palette | SL Quick Actions [SL4], MC [MC7] | 2 | new (v1 has a help Overlay) | maybe |
| Back and forward navigation history | MC [MC15 v5.12.0] | 1 | new | maybe |
| Link to a Snippet by URL | SL `snippetslab://` [SL4], RC deeplinks [RC6] | 2 | new | maybe, as `tuisnip get <id>` |
| Shell keybinding that puts a Snippet on the prompt | PT recipe for bash, zsh, fish [PT1] | 1 | new | good |
| Run a Snippet directly | PT `exec` [PT2], CA Run Server [CA15] | 2 | new | good |
| Global hotkey or quick-access window | SL Assistant [SL3], RC [RC1], CA tray app [CA5] | 3 | new | poor: a terminal app has no system-wide hotkey; the tmux-window workflow covers it |
| Paste into the active app | SL [SL3], RC [RC1] | 2 | new | poor |
| Text expansion by keyword | RC [RC1] | 1 | new | poor: needs a system-wide input hook |

### Editing and viewing

| Feature | Who | Count | TuiSnip | Fit |
|---|---|---|---|---|
| Highlighted read view | SL [SL1], MC [MC1], LP [LP1], CA [CA3], NP [NP2] | 5 | has | |
| Line numbers | SL [SL2], NP [NP2] | 2 | has | |
| External editor | PT [PT1], NP [NP2] | 2 | has | |
| Highlighting while editing | SL [SL2], MC CodeMirror [MC15 v6.0.0], LP CodeMirror [LP1] | 3 | ROADMAP Later (own editor component) | |
| Automatic Language detection | SL on-device, top 50 languages [SL4], LP by GitHub API [LP2] | 2 | ROADMAP Later | |
| Code formatting | SL clang-format, Prettier, swift-format, optional on paste [SL2]; MC Prettier [MC8] | 2 | new | good, by piping through a configured formatter |
| Rendered Markdown | SL Markdown Snippets [SL5], MC Notes space [MC1], LP [LP1], CA [CA13] | 4 | new | maybe, with glamour |
| Version history | CA per-file diffs and restore [CA6]; MC points to git on the vault [MC10] | 1 | new (v1 non-goal) | maybe |
| Append the clipboard to a Snippet | NP `p` [NP2] | 1 | new | maybe |
| Recover a draft after a failed save | LP [LP4 v2.0.1-beta.2] | 1 | new | maybe |
| Built-in AI editing | MC AI Assistant [MC13], SL Apple Writing Tools [SL9 2.5] | 2 | new | poor: MCP covers this from outside |
| Export a Snippet as an image | MC [MC1] | 1 | new | poor |
| Live HTML preview, JSON graph, Jupyter rendering | MC [MC8], LP [LP1], CA [CA13] | 3 | new | poor |

### Variables and placeholders

| Feature | Who | Count | TuiSnip | Fit |
|---|---|---|---|---|
| Prompted parameters with defaults | PT `<param>`, `<param=default>`, a parameter used twice is filled once [PT1][PT2]; RC `{argument name="…" default="…"}` [RC2] | 2 | new (v1 non-goal) | good |
| A choice of defaults for one parameter | PT `<p=\|_a_\|\|_b_\|>`, picked with arrow keys [PT1] | 1 | new | good |
| Dynamic values (date, time, UUID, clipboard, cursor) | RC [RC2] | 1 | new | maybe |
| Embed another Snippet | RC `{snippet name="…"}` [RC2] | 1 | new | maybe |
| Modifiers on values (`uppercase`, `trim`, `json-stringify`, `percent-encode`) | RC [RC2] | 1 | new | maybe |
| Skip the prompt and fill parameters at the shell prompt | PT `--raw` and a Ctrl-N recipe, `main` only [PT1] | 1 | new | maybe |

SnippetsLab, massCode, Lepton, and nap have no Snippet placeholders. Cacher has them only inside its VS Code extension, through VS Code's own snippet syntax [CA10]. massCode's `{{variables}}` exist only in its HTTP client [MC15 v5.10.0].

### Sync, import, and export

| Feature | Who | Count | TuiSnip | Fit |
|---|---|---|---|---|
| Backup | SL tiered automatic backups, every 20 minutes up to monthly for 2 years [SL6] | 1 | partly: v1 backs up before each migration and keeps 3 | maybe: periodic backups are new |
| JSON export and import | SL [SL7], RC [RC4], CA export [CA8] | 3 | ROADMAP v1.1 | |
| Importer from SnippetsLab | MC [MC11] | 1 | ROADMAP Later | |
| Importers from other apps | SL Quiver, CodeBox [SL7]; RC TextExpander, aText, Espanso, PhraseExpress [RC1]; MC VS Code snippets, Raycast [MC11] | 3 | new | maybe: VS Code and pet formats suit terminal users |
| Export as plain files | SL zip of text files [SL7], NP stdout [NP1] | 2 | new | maybe |
| Cloud sync between machines | SL iCloud [SL6], RC Cloud Sync, Pro only [RC3], CA [CA1], LP and PT through Gist [LP1][PT1] | 5 | new (v1 non-goal) | maybe |
| GitHub Gist import | SL [SL7], MC public Gists [MC11], CA [CA9] | 3 | new | maybe |
| GitHub Gist publish or sync | SL publish and update [SL7], LP as its only store [LP1], CA two-way for personal Snippets [CA7], PT whole-file sync by modification time, no merge [PT2] | 4 | new | maybe |
| GitLab Snippets sync | PT [PT1] | 1 | new | maybe |
| Library kept as plain files the user syncs | MC Markdown vault, one `.md` per Snippet, watched for outside changes [MC9][MC10]; PT TOML files [PT1]; NP one file per Snippet plus `snippets.json` [NP2] | 3 | new (v1 stores SQLite) | maybe: changes the storage design |
| Repair after sync conflicts | MC Vault Doctor [MC15 v5.7.0] | 1 | new | only with file-based storage |

None of the seven documents end-to-end encryption. Raycast states data is encrypted at rest and in transit [RC3]. massCode warns against storing secrets in the vault [MC9]. Encryption is a v1 non-goal.

### Agent and CLI access

| Feature | Who | Count | TuiSnip | Fit |
|---|---|---|---|---|
| Print a Snippet from the command line | NP `nap <name>`, fuzzy, raw when piped [NP2]; PT `search` prints the pick [PT2]; SL `lab search` [SL8] | 3 | ROADMAP v1.2 (`get`) | |
| Create a Snippet from stdin or a file | NP `nap folder/name.lang < file` [NP1]; PT `new` [PT1]; SL `lab create` [SL8]; CA `snippets:add`, from the clipboard if no file [CA9] | 4 | ROADMAP v1.2 (`add`) | |
| JSON output | SL `lab` [SL8] | 1 | new | good, as a `get` flag |
| Interactive picker that prints to stdout | PT `search` [PT2]; NP composes with gum: `nap $(nap list \| gum filter)` [NP1] | 2 | new | good |
| Copy from the command line | PT `clip` [PT2] | 1 | new | good |
| MCP server | SL `lab mcp`: search, get, list filters, create when write access is on; no edit or delete [SL8]; MC HTTP MCP on localhost with a token: search, get, create; no delete [MC12] | 2 | new | good |
| Agent Skill file | SL `lab skill show` [SL8] | 1 | new | good |
| One-step setup for coding agents | SL: Claude Code, Codex, Cursor, Gemini CLI, Copilot, Antigravity [SL8]; MC documents setup for Codex, Claude Code, Cursor, Copilot [MC12] | 2 | new | maybe |
| Local HTTP API | MC Integration API [MC16] | 1 | new | poor: a CLI serves the same need |
| Shell completions | PT zsh [PT2] | 1 | new | good, once subcommands exist |
| Editor plugins (VS Code, IntelliJ) | MC [MC1], CA [CA10] | 2 | new | poor |
| Launcher integrations (Raycast, Alfred) | SL [SL4], MC [MC1], CA [CA10] | 3 | new | poor: macOS-only |
| Browser clipper, Slack | MC [MC16], CA [CA1] | 2 | new | poor |

### Other

| Feature | Who | Count | TuiSnip | Fit |
|---|---|---|---|---|
| Configurable keys | LP [LP3] | 1 | has | |
| Themes beyond light and dark | SL at least 18 [SL11], MC [MC14], LP 13 [LP4 v2.0.0], NP [NP2], CA paid tier [CA12] | 5 | ROADMAP Later | |
| Remember the last selection between runs | NP last folder and Snippet, `main` only [NP2] | 1 | ROADMAP Later | |
| Settings screen | LP [LP4 v2.0.1-beta.2] | 1 | ROADMAP Later | |
| Language statistics dashboard | LP [LP1] | 1 | new | poor |
| Public sharing pages and embeds | CA [CA11]; LP share links (source code only) [LP5] | 2 | new | poor |

## Discrepancies found while checking

- nap's README documents `nap fuzzy` and `f` to move a Snippet. In the source, `fuzzy` is not a subcommand and the move key is `R` [NP2].
- pet's README says `auto_sync` works for Gist, GitHub Enterprise, and GitLab. The source only reads `[Gist].auto_sync` [PT2]. This comes from reading the code, not from running it.
- Raycast's free Teams limit of 30 shared Snippets appeared only in a search summary and could not be confirmed on a page.

## Open questions for the human

1. Placeholders are a v1 non-goal. Should they go on the roadmap, and if so with pet's `<param=default>` syntax (its users could import their snippets unchanged) or Raycast's `{argument}` syntax, which has room for dynamic values?
2. The roadmap's `get` and `add` cover the basic CLI. Should the same milestone add a picker that prints to stdout, `--json`, and a shell keybinding recipe, or should those wait?
3. Should agent access be an MCP server, a CLI with JSON output plus an Agent Skill, or both? SnippetsLab ships both. Both products limit agents to search, get, and create.
4. Sync is the most common gap. Is it still out of scope for a single-user tool, or is "Gist sync" or "export to plain files in a git repo" small enough to schedule?

## Sources

SnippetsLab, manual pages under `https://www.renfei.org/snippets-lab/manual/mac/`:

- [SL1] https://www.renfei.org/snippets-lab/
- [SL2] https://www.renfei.org/snippets-lab/manual/mac/essentials.html
- [SL3] https://www.renfei.org/snippets-lab/manual/mac/assistant.html
- [SL4] https://www.renfei.org/snippets-lab/manual/mac/advanced.html
- [SL5] https://www.renfei.org/snippets-lab/manual/mac/markdown.html
- [SL6] https://www.renfei.org/snippets-lab/manual/mac/sync-and-backup.html
- [SL7] https://www.renfei.org/snippets-lab/manual/mac/import-and-export.html
- [SL8] https://www.renfei.org/snippets-lab/manual/mac/agentic-tools.html
- [SL9] Release notes: https://www.renfei.org/snippets-lab/manual/mac/release-notes/macos-2.4.html, `macos-2.5.html`, `macos-2.6.html`, `macos-2.7.html`
- [SL10] https://www.renfei.org/snippets-lab/manual/mac/faq.html
- [SL11] https://www.renfei.org/snippets-lab/manual/mac/themes.html

Raycast:

- [RC1] https://manual.raycast.com/snippets
- [RC2] https://manual.raycast.com/dynamic-placeholders
- [RC3] https://manual.raycast.com/cloud-sync
- [RC4] https://manual.raycast.com/import-export
- [RC5] https://manual.raycast.com/ai/ai-extensions
- [RC6] https://developers.raycast.com/information/lifecycle/deeplinks
- [RC7] https://manual.raycast.com/teams/shared-features
- [RC8] https://www.raycast.com/changelog

massCode:

- [MC1] https://github.com/massCodeIO/massCode (README)
- [MC2] https://masscode.io/documentation/code/folders.html
- [MC3] https://masscode.io/documentation/code/tags.html
- [MC4] https://masscode.io/documentation/code/library.html
- [MC5] https://masscode.io/documentation/code/fragments.html
- [MC6] https://masscode.io/documentation/search.html
- [MC7] https://masscode.io/documentation/command-palette.html
- [MC8] https://masscode.io/documentation/code/snippets.html
- [MC9] https://masscode.io/documentation/storage.html
- [MC10] https://masscode.io/documentation/sync.html
- [MC11] https://masscode.io/documentation/imports.html
- [MC12] https://masscode.io/documentation/mcp.html
- [MC13] https://masscode.io/documentation/ai.html
- [MC14] https://masscode.io/documentation/themes.html
- [MC15] Releases: https://github.com/massCodeIO/massCode/releases (tags v5.7.0, v5.10.0, v5.12.0, v6.0.0, v6.1.0)
- [MC16] https://masscode.io/documentation/clipper.html

Lepton:

- [LP1] https://github.com/hackjutsu/Lepton#readme
- [LP2] https://github.com/hackjutsu/Lepton/blob/master/wiki/faq.md
- [LP3] https://github.com/hackjutsu/Lepton/blob/master/wiki/configuration.md
- [LP4] Releases: https://github.com/hackjutsu/Lepton/releases (tags v2.0.0, v2.0.1-beta.2)
- [LP5] Source: https://github.com/hackjutsu/Lepton/blob/master/app/utilities/i18n/locales/en.js, https://github.com/hackjutsu/Lepton/blob/master/app/reducers/reducer_pinned_tags.js

Cacher:

- [CA1] https://www.cacher.io/features
- [CA2] https://www.cacher.io/docs/getting-started/organizing-with-labels
- [CA3] https://www.cacher.io/docs/getting-started/creating-your-first-snippet
- [CA4] https://www.cacher.io/docs/getting-started/using-your-snippets
- [CA5] https://www.cacher.io/docs/guides/keyboard-shortcuts
- [CA6] https://www.cacher.io/docs/guides/snippets/snippet-file-history
- [CA7] https://www.cacher.io/docs/guides/snippets/syncing-with-github-gist
- [CA8] https://www.cacher.io/docs/guides/snippets/exporting-snippets
- [CA9] https://github.com/CacherApp/cacher-cli
- [CA10] https://www.cacher.io/docs/integrations/editor-plugins/visual-studio-code, https://www.cacher.io/docs/integrations/raycast, https://github.com/CacherApp/alfred-cacher
- [CA11] https://www.cacher.io/docs/getting-started/sharing-snippets
- [CA12] https://www.cacher.io/pricing
- [CA13] https://www.cacher.io/docs/guides/render-modes/markdown, https://www.cacher.io/docs/guides/render-modes/jupyter-notebooks
- [CA14] https://www.cacher.io/docs/guides/snippets/file-attachments
- [CA15] https://github.com/CacherApp/cacher-run-server

pet:

- [PT1] https://github.com/knqyf263/pet#readme
- [PT2] Source on `main`: https://github.com/knqyf263/pet/tree/main (`cmd/search.go`, `cmd/exec.go`, `cmd/clip.go`, `cmd/new.go`, `cmd/edit.go`, `cmd/util.go`, `dialog/params.go`, `dialog/view.go`, `sync/sync.go`, `config/config.go`, `misc/completions/zsh/_pet`)
- [PT3] Releases and unreleased commits: https://github.com/knqyf263/pet/releases, `gh api repos/knqyf263/pet/compare/v1.0.1...main`

nap:

- [NP1] https://github.com/maaslalani/nap#readme
- [NP2] Source on `main`: https://github.com/maaslalani/nap/tree/main (`main.go`, `snippet.go`, `model.go`, `keys.go`, `list.go`, `config.go`, `state.go`, `editor.go`)
- [NP3] Releases and unreleased commits: https://github.com/maaslalani/nap/releases, `gh api repos/maaslalani/nap/compare/v0.1.1...main`
