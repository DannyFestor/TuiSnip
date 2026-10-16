# Trash is a soft delete read through live views

Deleting a Snippet or Folder sets `deleted_at` on that one row, the Trash item. The rows inside a trashed Folder stay untouched. They are in the Trash because an ancestor is. Every Library read that leaves the Trash out selects from the views `live_snippets` and `live_folders`. Those views use a recursive CTE to drop any row whose own `deleted_at`, or an ancestor's, is set. Export and the Import emptiness counts read the base tables, because the Library includes the Trash. Decided in [How the database records a Trash item and what the migration does](https://github.com/DannyFestor/TuiSnip/issues/208).

## Considered Options

- **A `trash_items` table beside unchanged rows.** Every read would need an anti-join against it, and whether a row is trashed would live in two places.
- **Shadow tables** (`trashed_snippets`, `trashed_fragments`, …). They duplicate the schema and the Tag links, and restoring means copying rows back.
- **Trash membership stored on every row of a trashed subtree.** Queries get a plain `WHERE`, but trash, restore, and move would all have to keep a whole subtree's column in step. Working it out from the ancestors stores nothing that can go stale, and a Library this size keeps the recursive CTE cheap.

## Consequences

- An adapter test reads every query in `db/queries/` and fails when a SELECT names `snippets`, `folders`, or `snippet_tag` without being on its allowlist. That allowlist covers writes, the views, the Trash queries, and the whole-Library reads, each with a reason. Tag counts join `live_snippets` instead of reading `snippet_tag` alone.
- The finders the existing Actions use read through the views, so a Trash item looks deleted to them. The domain entities don't know about the Trash. A `trash` app package owns the Trash item read model and its Actions.
- The foreign keys still cascade. Permanently deleting a Folder first **detaches** every separate Trash item inside its subtree, such as a Snippet deleted before its Folder, so that the cascade doesn't take it along. Detaching sets the item's `folder_id`/`parent_id` to NULL and saves the Folder's path as a JSON array of names in `former_folder_path`/`former_parent_path`. A non-empty former path is how restore knows to say "its Folder was deleted", and how the Trash shows where the item was. Export carries the same fields.
