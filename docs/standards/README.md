# Standards

How code is written in this repo. Read the matching standard before writing or reviewing code in its area.

| Standard | Covers | Read before |
|---|---|---|
| [architecture](architecture.md) | Hexagonal layers, package layout, Actions, capabilities, error wrapping, where context and logging go, the test tiers, TUI components | adding a package, an Action, or a TUI component, or importing across layers |
| [code](code.md) | Conventions no linter checks: validation into value objects, context, concurrency, logging, comments, generated code | writing any Go |
| [testing](testing.md) | Choosing a tier, naming, table tests, test doubles, feature, e2e, fuzz, and property tests | writing a test |
| [linting](linting.md) | Why the non-obvious golangci-lint rules exist, and the `nolint` format | adding a `//nolint` or changing `.golangci.yml` |
| [database](database.md) | Schema conventions, migrations, sqlc queries, transactions, connections, start-up | touching `db/`, `sqlc.yaml`, or `internal/adapters/sqlite` |

Related documents:

- [`CONTEXT.md`](../../CONTEXT.md): the domain terms. They are binding for identifiers, test names, and log keys.
- [`docs/spec/`](../spec/): what the product does. [v1](../spec/v1.md), [config](../spec/config.md), [UI](../spec/ui.md).
- [`docs/toolchain.md`](../toolchain.md): each tool and library, why it was chosen, and what was rejected.
- [`docs/adr/`](../adr/): decisions that are hard to reverse.
