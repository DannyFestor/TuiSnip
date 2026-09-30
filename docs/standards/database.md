# Database

How TuiSnip stores its data in SQLite: the schema conventions, migrations, queries, connections, and start-up. Read it before touching `db/`, `sqlc.yaml`, or `internal/adapters/sqlite`. Where the adapter sits and what it may import is in [architecture](architecture.md). The entity rules the schema mirrors are in [`docs/spec/v1.md`](../spec/v1.md).

## Stack

- SQLite through `modernc.org/sqlite`, which compiles SQLite into the binary (3.53 at the time of writing). The system SQLite never matters. The one side effect: an `sqlite3` CLI older than 3.37 can't open the file, because the tables are `STRICT`.
- goose for migrations, sqlc for queries. No ORM.
- One database file, `tuisnip.db`, in the XDG data directory.

## Schema

| Table | Holds | Deleting a row removes |
|---|---|---|
| `folders` | the Folder tree, as an adjacency list on `parent_id` | its subfolders and their Snippets |
| `snippets` | Snippets; `folder_id` NULL means at the Root | its Fragments and Tag links |
| `fragments` | Fragments, ordered by `position` within their Snippet | nothing else |
| `tags` | Tags | their links; the Snippets stay |
| `snippet_tags` | which Snippet carries which Tag | nothing else |

The Root is not stored. It is the absence of a Folder, so `folder_id IS NULL` and `parent_id IS NULL` mean "at the Root".

### Conventions

- Every table is `STRICT`, so SQLite refuses a value of the wrong type rather than storing it.
- **IDs** are UUIDv7 as lowercase `TEXT`, generated in Go by the `IDGenerator`. The schema never generates one.
- **Timestamps** are `INTEGER` Unix nanoseconds, UTC. Every entity table has `created_at` and `updated_at`. Join tables have none.
- **Optional text** is `NOT NULL DEFAULT ''`. The domain doesn't distinguish "no Description" from an empty one, so NULL would only add a pointer.
- **Languages** are stored as chroma's canonical lexer name (`plaintext`, `Go`, `Bash`), the string `value.Language` holds.
- **Foreign keys** cascade on delete. Subtree delete is one `DELETE FROM folders WHERE id = ?`. When Trash arrives, deleting becomes a soft delete and the cascade only runs when something is removed for good.
- **Indexes**: one per foreign key that isn't already the leading column of a primary key or unique constraint.
- **Tag uniqueness** is on `name_key`, the case-folded form from `value.TagName.Key()`, never on `name` with `COLLATE NOCASE`. `NOCASE` folds only ASCII, so SQLite and memsearch would disagree about which Tags are the same.

### Constraints

The database is the last line of defence. It repeats every rule we know, except rules that look inside a string, such as "no commas in a Tag name".

| Rule | CHECK |
|---|---|
| IDs and ID references are UUIDs | `length(id) = 36` |
| titles and names are trimmed and not blank | `name <> '' AND name = trim(name, ' ' \|\| char(9, 10, 11, 12, 13))` |
| Tag keys, Languages, and Default Languages are not blank | `<> ''` |
| Fragment positions are not negative | `position >= 0` |
| timestamps are sane | `created_at > 0`, `updated_at >= created_at` |
| a Folder is not its own parent | `parent_id <> id` |

A CHECK may be weaker than the domain rule, never stricter. A stricter CHECK would reject a value the domain accepted, and the user would get a failed save with no explanation. `trim` above is an example: SQLite removes only the characters listed, while `strings.TrimSpace` removes all Unicode whitespace.

v1's "exactly one Fragment" is a product limit, not a data rule. The domain enforces it, and the schema allows many.

### Timestamp defaults

`created_at` and `updated_at` default to the current time, `CAST(unixepoch('subsec') * 1000000000 AS INTEGER)`. That default is only for rows written by hand, for example in the `sqlite3` CLI. It has millisecond precision and doesn't come from the `Clock`.

- Every insert query lists both timestamp columns, and the app always supplies them from the `Clock`. A row that took the default would carry a time the domain never saw, and its first save would fail the concurrency check.
- `created_at` is never updated.
- `updated_at` changes when:

| Entity | Bumped by | Not bumped by |
|---|---|---|
| Snippet | an edit in the Snippet pane: title, Description, Tags, Fragment content or Language | moving it; renaming, merging, or deleting a Tag it carries |
| Fragment | its content or Language changing (which also bumps the Snippet) | |
| Folder | a rename or a Default Language change | moving it |
| Tag | a rename; being the surviving Tag of a merge | |

## Migrations

- Files live in `db/migrations/`, embedded by `db/migrations/embed.go`. Only `internal/adapters/sqlite` may import that package.
- Names are sequential: `00006_add_trash.sql`. Timestamped names are not used, because goose refuses a migration older than one already applied.
- One migration per table or per change. Indexes go in the migration of their table.
- **Up only.** There are no `-- +goose Down` sections. The app only migrates forward. To undo something, write a new migration, and to recover data, restore a [backup](#backups).
- Migrations are editable until the first release tag (v0.1). After that they're append-only, because an installed database has already applied them.
- goose wraps each migration in a transaction, where `PRAGMA foreign_keys` has no effect. A migration that rebuilds a table (create, copy, drop, rename) needs `-- +goose NO TRANSACTION` and has to switch foreign keys off and on itself.

## Queries

- One query file per table in `db/queries/` (`snippets.sql`, `fragments.sql`). A query that reads several tables goes in the file of the table it returns.
- sqlc generates `internal/adapters/sqlite/sqlcgen/` from `sqlc.yaml` with `make generate:sql`. Never edit the output.
- Queries arrive with the Action that needs them. The adapter method that runs them is shaped by the Action's capability interface, not by the table.
- `sqlc.yaml` maps every ID column to `sqltype.ID` and every timestamp column to `sqltype.Timestamp` with wildcard column overrides (`*.id`, `*.created_at`). A new ID or timestamp column with a different name needs its own override.
- `sqltype` (`internal/adapters/sqlite/sqltype`) holds the `Scan`/`Value` types sqlc generates against. The standard library's `uuid.UUID` has neither method. A value that fails `Scan` becomes an error wrapping `domain.ErrCorruptRecord`.
- `emit_empty_slices` and `emit_pointers_for_null_types` are on. `emit_interface` is off, because nothing mocks `sqlcgen`: adapter tests run against real SQLite.
- `sqlc verify` needs sqlc Cloud and is not used. The generated-code drift check in CI covers the output.

## Reading entities

`sqlite` rebuilds entities with the same `domain` constructors the Actions use. There is no constructor that skips validation. An unexported function per entity (`snippetFromRows(snippet, fragments, tags)`) does the conversion. It wraps any failure in `domain.ErrCorruptRecord` and logs the row's ID.

Loading many Snippets, for the Search index or a Folder listing, runs one query per table (Snippets, Fragments, Tag links) in one read transaction and stitches them together in Go. It never runs one query per Snippet.

## Writing and transactions

- Each capability method writes one whole aggregate in one transaction inside the adapter. `Insert(ctx, snippet)` writes the Snippet, its Fragments, and its Tag links together. Actions never see a transaction.
- **Optimistic concurrency.** Saving an edited Snippet updates `WHERE id = ? AND updated_at = ?` with the `updated_at` it loaded. If no row changed, the adapter looks the ID up again: it returns `domain.ErrNotFound` if the row is gone and `domain.ErrConflict` if it's there. Only Snippet saves are checked. Folder and Tag changes are last-write-wins.
- **Tag merge.** Renaming a Tag onto an existing Tag's name keeps the existing Tag's ID, gives it the spelling the user typed, moves the renamed Tag's links over with `INSERT OR IGNORE`, and deletes the renamed Tag, all in one transaction.
- **Folder cycles.** The domain rejects moving a Folder into its own subtree. The adapter's move method checks again with a recursive query inside its write transaction and returns `domain.ErrFolderCycle`. The second check exists because two instances can each pass the domain check with a tree the other is changing. Together they would leave a loop cut off from the Root.
- The delete confirmation's counts come from a recursive query over the subtree.

## Errors

- `sql.ErrNoRows` becomes `domain.ErrNotFound`.
- Constraint failures are `*sqlite.Error`. `Code()` is SQLite's extended code: `SQLITE_CONSTRAINT_UNIQUE` (2067), `SQLITE_CONSTRAINT_PRIMARYKEY` (1555, even though the message says "UNIQUE constraint failed"), `SQLITE_CONSTRAINT_FOREIGNKEY` (787), and `SQLITE_CONSTRAINT_CHECK` (275). Constants are in `modernc.org/sqlite/lib`. A failure the adapter expects, such as a Tag name clash, is converted to a sentinel. Any other failure is a bug and returns wrapped as it is.
- Wrapping follows [architecture](architecture.md#errors): `"sqlite.Repository.<Method>: %w"`.

## Connections

Every connection gets the same pragmas through the DSN. modernc applies `_pragma` values to each new pooled connection:

| Setting | Value | Why |
|---|---|---|
| `foreign_keys` | `1` | SQLite leaves it off by default, and the cascades depend on it |
| `journal_mode` | `WAL` | several instances share the file; readers don't block the writer |
| `synchronous` | `NORMAL` | safe with WAL, and far fewer fsyncs than `FULL` |
| `busy_timeout` | `5000` ms | how long an instance waits for the write lock before giving up with an error |
| `_txlock` | `immediate` | every write transaction takes the write lock at `BEGIN`, so two instances queue instead of failing halfway through |

Each value is a named constant in the adapter.

## Start-up

`sqlite` opens the database in this order:

1. Take an exclusive `flock` on `tuisnip.db.lock` beside the database. Two instances starting together (a tmux restore, say) then migrate one after the other. The kernel drops the lock if the process dies.
2. Open the database, creating the file if it doesn't exist.
3. Compare the database's goose version with the newest embedded migration. If the database is newer, refuse to start: `The database was created by a newer TuiSnip (schema 7; this build knows 5). Upgrade TuiSnip.` goose itself would skip it silently, and an older build writing through an older model could lose data.
4. If an existing database has pending migrations, [back it up](#backups). If the backup fails, refuse to start.
5. Run the pending migrations.
6. Release the lock.

### Backups

- Before migrating an existing database, `VACUUM INTO` a copy in `backups/` beside the database: `tuisnip-<UTC timestamp>-schema<from version>.db`. `VACUUM INTO` gives a consistent copy even while other instances hold the file in WAL mode.
- Keep the latest 3 and delete older ones.
- A new database, or one with nothing pending, gets no backup.
- `tuisnip --paths` prints the backup directory.

## Tests

- Feature tests open a real database file in `t.TempDir()`, with production's exact pragmas and WAL, through `bootstrap`. Plain `:memory:` gives every pooled connection its own empty database. Shared-cache memory databases either leak between tests or lock differently from production.
- The adapter's own tests (row conversion, error mapping, the cycle re-check, start-up ordering) use the same temp-file setup.
