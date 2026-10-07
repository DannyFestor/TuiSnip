## Code standards

Before writing or reviewing Go, start from the index in `docs/standards/README.md`.

## Workflow

From the first action, follow the workflow `docs/toolchain.md` enforces: edit files with the edit tool, branch before the first commit, and give each PR a Conventional Commit title. It also lists the shell pitfalls.

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues for `DannyFestor/TuiSnip` via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
