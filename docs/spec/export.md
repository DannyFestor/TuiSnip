# Export format

`tuisnip export <file>` writes the whole Library to one JSON document, as a backup TuiSnip can restore. This document is the reference for that file. Domain terms come from [`GLOSSARY.md`](../../GLOSSARY.md). The command itself is in [v1: Command line](v1.md#command-line).

## Contents

The file holds everything in the database and nothing outside it: every Folder, every Tag (Tags no Snippet carries included), and every Snippet with its Fragments in order, each with its ID and dates. The Trash is part of it: each Trash item carries its deletion date and, once detached, the path it was in, so importing the file gives back the same Trash items, in the same order, with the same restore destinations. Neither `config.toml` nor the state file (Sort order, Collapsed Folders) is part of it.

## Document

```json
{
  "format": "tuisnip-export",
  "version": 2,
  "folders": [
    {
      "id": "0194c3a0-0000-7000-8000-000000000001",
      "parent_id": null,
      "name": "scripts",
      "default_language": "Bash",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z",
      "deleted_at": null,
      "former_parent_path": []
    }
  ],
  "tags": [
    {
      "id": "0194c3a0-0000-7000-8000-000000000002",
      "name": "Shell",
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z"
    }
  ],
  "snippets": [
    {
      "id": "0194c3a0-0000-7000-8000-000000000003",
      "folder_id": "0194c3a0-0000-7000-8000-000000000001",
      "title": "loop",
      "description": "",
      "tag_ids": [
        "0194c3a0-0000-7000-8000-000000000002"
      ],
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-02T09:00:00.000000000Z",
      "fragments": [
        {
          "id": "0194c3a0-0000-7000-8000-000000000004",
          "title": "loop.sh",
          "language": "Bash",
          "content": "for f in *; do echo \"$f\"; done\n",
          "created_at": "2026-03-01T09:30:00.123456789Z",
          "updated_at": "2026-03-02T09:00:00.000000000Z"
        }
      ],
      "deleted_at": null,
      "former_folder_path": []
    },
    {
      "id": "0194c3a0-0000-7000-8000-000000000005",
      "folder_id": null,
      "title": "prune",
      "description": "",
      "tag_ids": [],
      "created_at": "2026-03-01T09:30:00.123456789Z",
      "updated_at": "2026-03-01T09:30:00.123456789Z",
      "fragments": [
        {
          "id": "0194c3a0-0000-7000-8000-000000000006",
          "title": "",
          "language": "Bash",
          "content": "docker system prune\n",
          "created_at": "2026-03-01T09:30:00.123456789Z",
          "updated_at": "2026-03-01T09:30:00.123456789Z"
        }
      ],
      "deleted_at": "2026-03-03T18:00:00.000000000Z",
      "former_folder_path": [
        "docker",
        "cleanup"
      ]
    }
  ]
}
```

| Field | Holds |
|---|---|
| `format` | always `tuisnip-export`, so a reader can tell the file from any other JSON |
| `version` | the format version, `2`. It is separate from the database schema version. |
| `folders`, `tags`, `snippets` | flat lists, linked by ID; the Trash items and everything inside a trashed Folder included |
| `parent_id`, `folder_id` | the Folder's parent or the Snippet's Folder; `null` means at the Root. A Trash item keeps its Folder for as long as that Folder exists, in the Trash or not. |
| `deleted_at` | when the record was moved to the Trash; `null` when it isn't a Trash item. It is set only on the Trash item itself: a Snippet or subfolder inside a trashed Folder keeps `null`, and a Snippet deleted before its Folder keeps its own date. |
| `former_parent_path`, `former_folder_path` | the path a detached Trash item was in, as Folder names, Root first; `[]` unless the item was detached when its Folder was permanently deleted. Names are listed one by one because a name can contain `/`. |
| `default_language`, `language` | chroma's canonical lexer name, as stored ([database](../standards/database.md#conventions)) |
| `tag_ids` | the IDs of the Tags the Snippet carries, by Tag name ignoring case |
| `fragments` | the Snippet's Fragments, in the order the user set |
| `title` on a Fragment | the Fragment title; `""` when it has none |

An empty list is written as `[]`, never `null`. Every Folder and Snippet carries `deleted_at` and its former path, even when it isn't in the Trash.

## Determinism

Exporting an unchanged Library twice gives byte-identical files.

- Each list is sorted by created date, then ID.
- Timestamps are UTC RFC 3339 with all nine fractional digits, so every timestamp has the same width.
- The JSON is indented with two spaces and ends with a newline. `<`, `>` and `&` are written as they are, not escaped.

## Writing

- The file is written to a temporary file in the target's directory, then renamed over the target. An existing file is replaced without asking, and a failed Export leaves it untouched and no temporary file behind.
- The file is readable by its owner only (`0600`), because Snippets often hold secrets.
- The target's directory must exist. A missing one fails the Export instead of being created, so a mistyped path doesn't scatter backups.
- The Library is read in one read transaction, so an open TUI writing at the same time can't leave a Snippet pointing at a Folder the file lacks. There is no lock against a running TUI.
- A record the database holds but the domain rejects, such as a Fragment whose Language chroma no longer knows, fails the Export and names the record. Browse skips such a record, but a backup that silently drops one couldn't be restored.
- `-` is refused as the file, with usage and exit 2. Writing to stdout isn't supported.

## Reading

`tuisnip import <file>` ([v1: Command line](v1.md#command-line)) reads the file back.

- A file whose `format` isn't `tuisnip-export` is refused as not being an Export.
- A field the format doesn't have is refused, so a mistyped key such as `parent-id` can't drop its value silently.
- Each record is built through the same checks the TUI applies. A bad record is named by its kind and ID (`snippet 0194…: value: title is longer than 200 characters`), and every bad record is reported at once.
- A Trash item is checked like a live record. A deletion date before the created date, or one that doesn't parse, fails the Import and names the record. So does a former path on a record outside the Trash or on a Trash item that still has its Folder.

## Versioning

Import reads versions 1 and 2. Version 1 has no `deleted_at` or former paths, so a version 1 file reads as a Library with an empty Trash, and a version 1 file carrying either field is refused like any field the format lacks. Import refuses a newer version and asks for an upgrade, so an older TuiSnip asks for one on a version 2 file.
