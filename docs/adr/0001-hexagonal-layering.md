# Hexagonal layering with consumer-owned interfaces and independent concerns

TuiSnip keeps a stdlib-only `domain`, Actions in `internal/app`, and adapters in `internal/adapters`, with dependencies pointing inward and a single composition root in `internal/bootstrap`. Two choices in this shape will look odd to someone used to other hexagonal Go projects, and both are deliberate.

There is no `ports` package. Each Action declares the interfaces it needs in its own package: one-method capabilities, combined by embedding into `<Action>Repository`, `<Action>Clipboard`, or `<Action>Index`. A shared ports package would give one place to find every outside dependency, but its interfaces grow to serve every caller, and every adapter would import it. Go satisfies interfaces implicitly, so adapters import only `domain`, and `bootstrap` wiring them together is the compile-time check.

The packages under `internal/app` (`snippet`, `folder`, `tag`, `browse`, `search`) never import each other, and go-arch-lint gives each its own component to enforce that. When one concern needs data another concern owns, it declares an interface and `bootstrap` passes the same adapter. This duplicates a few small interfaces. The alternative is concerns calling each other, which leads to import cycles and Actions that can't be tested alone.

## Consequences

- go-arch-lint's deep scan stays off. It treats passing an adapter into a component as a dependency on that adapter, which is the exact edge this layering inverts. The TUI's own interfaces (remembered state, editor command) are filled by driven adapters, so deep scan would flag them even on the TUI alone.
- Rules in detail: [docs/standards/architecture.md](../standards/architecture.md).
