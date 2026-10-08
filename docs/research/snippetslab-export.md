# SnippetsLab export formats and data model

Research for [#176](https://github.com/DannyFestor/TuiSnip/issues/176), part of map [#175](https://github.com/DannyFestor/TuiSnip/issues/175). Sources checked 2026-10-08, against the SnippetsLab 2.7 manual (last updated 2026-10-02/03). No local SnippetsLab data was read.

`M/` below stands for `https://www.renfei.org/snippets-lab/manual/mac/`.

## Question

What export formats does SnippetsLab offer (JSON, library bundle, plain files, other), and where does it keep its library on disk? For each readable source, which parts of the data model survive: Snippets, Fragments (several per Snippet), Folders and nesting, Tags, notes/descriptions, Languages, dates, Smart Groups, favorites? Name what TuiSnip could read without SnippetsLab running.

## Answer

- **Formats.** Library > Export offers four: SnippetsLab Library (`.snippetslablibrary`), JSON Document, XML Document, and Plain Text Files (Zip Archive). Only the library format carries attachments. (M/import-and-export.html, section 8.2)
- **Best source for TuiSnip: the JSON export.** It is the only format with a published schema, and it keeps every Fragment of a Snippet, the Folder tree, Tags, Fragment notes, Languages, dates, Trash state, Smart Groups and favorites. (M/tips-and-tricks/json-import.html; M/release-notes/macos-2.1.html)
- **Second source: the automatic backups.** The manual says every automatic backup includes a JSON export of the library (M/faq.html, "Will I be locked into SnippetsLab?"). The default backup folder and the file name inside each backup are documented only by third parties (see [Library and backup locations](#library-and-backup-locations)).
- **The live library is not a practical source.** Its internal layout is undocumented. Third parties report keyed-archive binary plists inside the bundle. One Go project parses them, but only partly (see [Live library bundle](#live-library-bundle-snippetslablibrary)).
- **The `lab` CLI and MCP server need the app running.** Both send requests to the running SnippetsLab app and launch it if needed. (M/tips-and-tricks/lab.html, DESCRIPTION; M/agentic-tools.html)
- **Multi-Fragment Snippets do export.** The JSON format requires `fragments` as an array with at least one entry. Each Fragment has its own title, Language, notes and content. Exporting them is not the hard part. Fitting them into TuiSnip's one-Fragment v1 is.

## Findings

### SnippetsLab's data model (official)

- **Snippet.** A snippet has a title and holds one or more Fragments. It is in at most one folder. Snippets with no folder appear in *Uncategorized*. It can have any number of tags. It can be pinned and locked. (M/essentials.html; JSON spec, "Snippet")
- **Fragment.** "A snippet can contain multiple fragments, each with its own language, notes, and content." Fragments are ordered and can have a title. (M/essentials.html; JSON spec, "Fragment")
- **Notes live on the Fragment, not the Snippet.** "Add notes to a fragment for context about its content." Notes are rich text: bold, italic, underline, code, mark, strikethrough, sub/superscript and links, with ranges in UTF-16 code units. There is no snippet-level description field. (M/essentials.html; JSON spec, "Note Attribute"; M/release-notes/macos-2.6.html)
- **Folder.** Folders nest without a depth limit. Deleting a folder deletes its subfolders and moves their snippets to Uncategorized. A folder may have a default language. Without one, new snippets use the app-wide default. (M/essentials.html)
- **Tag.** "Tag names are case-insensitive, but their original capitalization is preserved." A tag may have one of 12 accent colors. (M/essentials.html; JSON spec, "Tag")
- **Smart Group.** A saved query. In JSON its rules are an NSPredicate format string, e.g. `ANY parts.language.displayName ==[cd] "Markdown"`. (JSON spec, "Smart Group" and "Predicate")
- **Favorites.** These are sidebar shortcuts to folders, tags and smart groups, not to snippets. The JSON calls them `shortcuts`. Pinning is the per-snippet equivalent. (M/essentials.html; JSON spec, "Shortcut")
- **Trash.** Deleted snippets keep a `dateDeleted`. The Trash is never emptied automatically. (M/essentials.html; JSON spec, "Snippet")
- **Language.** Each Fragment's Language is a Pygments lexer class name, such as `PythonLexer`. (JSON spec, "Discussion: Language")
- **Dates.** Snippets and Fragments each have `dateCreated` and `dateModified`, as ISO 8601 with time zone and no fractional seconds. A snippet's modification date changes only when its title, content, notes, fragments or attachments change. Tag, language, folder, pin, lock and Trash changes leave it alone. (JSON spec, "Discussion: Date"; M/essentials.html)
- **Attachments.** Files attached per Fragment: up to 100 per Fragment, each under 100 MB. (M/essentials.html)
- **IDs.** Every entity has a UUID. `snippetslab://` links point at snippet and fragment UUIDs. (JSON spec, "Discussion: UUID"; M/advanced.html)

### What survives in each source

| Data | Library bundle | JSON export | Backup JSON | XML export | Plain-text zip | `lab` CLI |
|---|---|---|---|---|---|---|
| Snippet title, UUID | yes | yes | yes | yes | ? | yes |
| Several Fragments per Snippet | yes | yes, ordered array | yes | yes | ? | yes |
| Fragment title | yes | yes | yes | yes | ? | ? |
| Fragment content | yes | yes | yes | yes | yes | yes |
| Fragment Language | yes | yes (lexer class) | yes | yes | file extension only | yes |
| Fragment notes | yes | yes, plus rich-text ranges | yes | yes | ? | yes |
| Folder and nesting | yes | yes (`children`) | yes | yes | ? | folder path |
| Folder default language | yes | yes | yes | ? | no | no |
| Tags | yes | yes, plus color | yes | yes | ? | yes |
| Smart Groups | yes | yes (NSPredicate) | yes | ? | no | no |
| Favorites (shortcuts) | yes | yes | yes | ? | no | no |
| Pinned, locked | yes | yes | yes | ? | no | no |
| Created/modified dates | yes | yes | yes | yes | ? | modified shown, created sortable |
| Trash (`dateDeleted`) | yes | yes | yes | ? | ? | no, Trash excluded |
| Attachments | yes | no | file names at most | no | Markdown snippets only | names, for search |
| Needs app running | no | no | no | no | no | **yes** |

Sources for each column:

- **Library bundle.** Import "including its folders, smart groups, snippets, tags, and attachments" (M/import-and-export.html, 8.1). It is the format to use "to keep a complete copy" (8.2).
- **JSON export.** The spec covers every row marked *yes*. "Exported files contain some additional keys that aren't listed below." "Attachments aren't imported from JSON files." (JSON spec, "General Information") Version 2.1 notes: "JSON and XML exports now contains all library data, including UUID and folder hierarchy." (M/release-notes/macos-2.1.html)
- **Backup JSON.** Official: "Every automatic backup includes a JSON export of your library." (M/faq.html) The extra root keys (`app`, `name`, `schema`, `date`) and the `contents.attachments` array come from a secondary source, mcp-snippetslab's decoder ([BackupSnippetRepository.swift](https://github.com/teran/mcp-snippetslab/blob/master/Sources/mcp-snippetslab/Infrastructure/BackupSnippetRepository.swift)). No source says whether the backup JSON matches Library > Export > JSON byte for byte.
- **XML export.** Only the 1.4 and 2.1 release notes describe it, and no schema is published. A *yes* in this column rests on the 2.1 note's "all library data"; *?* means that note does not name the field.
- **Plain-text zip.** "Snippets now have appropriate file extensions when exported as files or as a zip archive." (M/release-notes/macos-1.7.html) When a Markdown snippet is exported as plain text, its attachments are included (M/release-notes/macos-2.3.html). The zip's file layout is undocumented: whether folders become directories, and how a multi-Fragment Snippet is split into files.
- **`lab` CLI.** `search --brief` returns "UUID, title, folder path, tags, languages, fragment count, total content length, date modified". `fetch --format json` returns "the snippet and its metadata … All fragments are included". (M/tips-and-tricks/lab.html) Agents get "the library that is currently open in SnippetsLab, excluding the Trash." (M/agentic-tools.html)

### Library and backup locations

Official facts:

- When iCloud Sync is on, "it keeps a single library in iCloud". It does not appear as a folder in iCloud Drive. (M/sync-and-backup.html, 7.1.1; M/faq.html)
- With iCloud Sync off, the user can move the library or open another `.snippetslablibrary` from Settings > Sync. (M/sync-and-backup.html, 7.2)
- Backups run every 20 minutes over the past 6 hours, hourly over 24 hours, daily over a month, weekly over 6 months, and monthly over 2 years. Automatic backups are on by default. The backup folder can be moved. "Backups in the default location can be lost if the app is deleted or a disk cleanup utility removes them." (M/sync-and-backup.html, 7.3)
- The manual gives no default path for either the library or the backups.

Secondary sources (third-party code, not confirmed by the vendor):

- **Default local library:** `~/Library/Containers/com.renfei.SnippetsLab/Data/Library/Application Support/com.renfei.SnippetsLab/main.snippetslablibrary`. ([SnipKit constants.go](https://github.com/lemoony/snipkit/blob/main/internal/managers/snippetslab/constants.go))
- **Custom library path:** stored under the key `User DesignatedLibraryPathString` in `~/Library/Containers/com.renfei.SnippetsLab/Data/Library/Preferences/com.renfei.SnippetsLab.plist`. (same file; [SnipKit docs](https://github.com/lemoony/snipkit/blob/main/docs/managers/snippetslab.md))
- **iCloud library:** `~/Library/Mobile Documents/iCloud~com~renfei~SnippetsLab/main.snippetslablibrary`. ([mcp-snippetslab README](https://github.com/teran/mcp-snippetslab)) SnipKit's docs say the Containers path above applies when iCloud is on. The two sources conflict, and no primary source settles it.
- **Default backup folder:** `~/Library/Containers/com.renfei.SnippetsLab/Data/Library/Application Support/Backups/`. Two independent projects agree on it: mcp-snippetslab and [vscode-snippetslab](https://github.com/sisoe24/vscode-snippetslab). Per mcp-snippetslab, each backup is a `<date>.snippetslab-backup` directory holding `library.json`. Sorting the directory names finds the newest.
- The app is sandboxed. It ships only through the Mac App Store; the Setapp build ended on 2026-05-01. (M/faq.html; M/tips-and-tricks/migrate-from-setapp.html) That is why its data sits under `~/Library/Containers/com.renfei.SnippetsLab/`.

### Live library bundle (`.snippetslablibrary`)

The format is undocumented. What third parties report:

- It is a directory bundle holding `Database/Snippets/<UUID>.data` (one file per snippet), `Database/folders.data`, `tags.data`, `smart groups.data`, `shortcuts.data` and `deletion.data`. ([SnipKit testdata](https://github.com/lemoony/snipkit/tree/main/internal/managers/snippetslab/testdata))
- Each file is an NSKeyedArchiver binary plist with keys like `com.renfei.SnippetsLab.Key.SnippetTitle`, `…SnippetParts`, `…SnippetPartContent`, `…SnippetPartLanguage` and `…SnippetTagUUIDs`. SnipKit (Go, `howett.net/plist`) walks the raw `$objects` graph. It reads only title, UUID, tag UUIDs, and the first Fragment's content and Language. ([parser.go](https://github.com/lemoony/snipkit/blob/main/internal/managers/snippetslab/parser.go), last changed 2024-08-12)
- mcp-snippetslab reports that `NSKeyedUnarchiver` can't decode these files without SnippetsLab's own classes (`SLSnippet`, …). It reads the backup JSON instead and calls that "the only programmatically readable source". Walking the raw plist, as SnipKit does, avoids that problem. The cost is depending on private, unversioned keys.

## What TuiSnip could read without SnippetsLab running

1. **A JSON export the user makes** (Library > Export > JSON Document). It has a documented schema, covers the whole model except attachments, and works for libraries kept on iCloud too.
2. **The newest automatic backup's JSON.** It needs no user action and is at most about 20 minutes old while the user edits. But its path and file name are undocumented, the folder can be moved, and backups can be turned off.
3. **The live bundle**, as a last resort: an undocumented private format, likely different for iCloud libraries, which could break with any app update.

XML has no schema. The plain-text zip loses structure. The `lab` CLI and MCP server need the app.

## Mapping gaps against TuiSnip's model

These apply to every structured source. They are for the import-design tickets to decide; this note does not decide them.

- **Fragments.** SnippetsLab Snippets can have several Fragments. TuiSnip v1 allows exactly one (GLOSSARY.md, Fragment). Import needs a policy: keep the first, join them, split them into several Snippets, or refuse.
- **Description.** SnippetsLab has no snippet-level description. Its notes are per Fragment and rich text, so they have to be mapped onto TuiSnip's single plain-text Description.
- **Fragment title.** TuiSnip has no equivalent.
- **Language.** SnippetsLab stores Pygments lexer class names such as `PythonLexer`, which need a table mapping them to TuiSnip Languages.
- **Folder default language.** It is optional in SnippetsLab, but every TuiSnip Folder has a Default Language.
- **Uncategorized** maps to the Root. A folder reference that doesn't resolve also lands in Uncategorized on SnippetsLab's own import (JSON spec, "Discussion: UUID").
- **Tags.** Case-insensitive in both apps. Tag colors have no equivalent.
- **Trash.** Maps to TuiSnip's Trash.
- **No TuiSnip equivalent:** Smart Groups, favorites, pinned, locked, attachments.
- **Dates.** ISO 8601 strings. SnippetsLab's `dateModified` ignores tag, folder and Trash changes.

## Open questions for the human

- What do the XML export and the plain-text zip actually look like: folders as directories, and how multi-Fragment Snippets are named? Exporting a small throwaway library would answer this.
- Is a backup's `library.json` the same as Library > Export > JSON Document? Which keys does it add beyond the published import schema?
- Where does the live library sit when iCloud Sync is on: under `~/Library/Containers/…` or `~/Library/Mobile Documents/iCloud~com~renfei~SnippetsLab/…`? The secondary sources disagree.
