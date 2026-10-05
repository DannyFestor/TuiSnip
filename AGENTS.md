## Code standards

Before writing or reviewing Go, read the matching standard. The index in `docs/standards/README.md` says which one applies.

## Workflow

Edit files with the edit tool, branch before the first commit, and give each PR a Conventional Commit title. The hooks that enforce this and the shell pitfalls are in `docs/toolchain.md`.

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues for `DannyFestor/TuiSnip` via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
