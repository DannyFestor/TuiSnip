## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues for `DannyFestor/TuiSnip` via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

## Code standards

Before writing or reviewing Go, read the matching file in `docs/standards/`:

- `architecture.md`: where code goes, Actions, interfaces, errors
- `code.md`: conventions no linter checks (validation, context, logging, generated code)
- `testing.md`: choosing a tier, naming, test doubles
- `linting.md`: why a linter rule exists, before adding a `//nolint`
