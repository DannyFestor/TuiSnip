# go-arch-lint for TuiSnip's hexagonal layout

Resolves [#7](https://github.com/DannyFestor/TuiSnip/issues/7). Researched 2026-09-30 against go-arch-lint v1.19.0. The draft config below was run with v1.19.0 on Go 1.27.1 against a throwaway project in `/tmp` that mirrors the planned layout.

## Question

What can go-arch-lint express and enforce (components, vendors, common components, deep scan, allowed edges)? How should `.go-arch-lint.yml` look for `internal/domain`, `internal/app` (with ports), `internal/adapters/{sqlite,clipboard,tui,cli,config}` and `cmd/tuisnip`? How does it run locally and in CI, how is it maintained, and how does depguard cover the rules go-arch-lint cannot express?

## Recommendation

**Tool:** go-arch-lint v1.19.0 for layer edges, plus depguard (inside golangci-lint) for the bans go-arch-lint can't express, mainly standard-library imports. go-arch-lint is actively maintained (four releases between July and September 2026), so no replacement is needed.

**Draft `.go-arch-lint.yml`** (checked: every legal edge passes, and each deliberate violation below was reported):

```yaml
version: 3
workdir: .

allow:
  depOnAnyVendor: false          # every third-party import must be granted via canUse
  deepScan: false                # see "Deep scan" finding: on globally it flags legitimate port injection into app
  ignoreNotFoundComponents: true # scaffold phase only; flip to false once every adapter directory exists

exclude:
  - test                         # test/feature, test/e2e

excludeFiles:
  - "^.*_test\\.go$"

vendors:
  sqlite-driver: { in: [ modernc.org/sqlite, modernc.org/sqlite/** ] }
  migrations:    { in: [ github.com/pressly/goose/v3, github.com/pressly/goose/v3/** ] }
  tui-framework: { in: [ github.com/charmbracelet/**, charm.land/** ] }
  highlighting:  { in: [ github.com/alecthomas/chroma/v2, github.com/alecthomas/chroma/v2/** ] }

components:
  domain:    { in: internal/domain/** }
  app:       { in: internal/app/** }
  ports:     { in: internal/app/ports }   # more specific match wins over app/**
  sqlite:    { in: internal/adapters/sqlite/** }
  clipboard: { in: internal/adapters/clipboard/** }
  tui:       { in: internal/adapters/tui/** }
  cli:       { in: internal/adapters/cli/** }
  config:    { in: internal/adapters/config/** }
  main:      { in: cmd/tuisnip }

commonComponents:
  - domain

deps:
  domain:    { mayDependOn: [ domain ] }
  ports:     { mayDependOn: [ ports ] }
  app:       { mayDependOn: [ app, ports ] }
  sqlite:
    mayDependOn: [ sqlite, ports ]
    canUse: [ sqlite-driver, migrations ]
  clipboard: { mayDependOn: [ clipboard, ports ] }
  tui:
    deepScan: true               # driving adapter: may be handed use cases, never a driven adapter
    mayDependOn: [ tui, app, ports ]
    canUse: [ tui-framework, highlighting ]
  cli:
    deepScan: true
    mayDependOn: [ cli, app, ports ]
  config:    { mayDependOn: [ config ] }
  main:
    anyProjectDeps: true         # composition root
    anyVendorDeps: true
```

Each component lists itself in `mayDependOn`. That's required, not decoration: without it, `sqlite` importing its own sqlc subpackage `sqlite/db` was reported as a violation.

**Complementary depguard rules** (`.golangci.yml`, golangci-lint v2 syntax). go-arch-lint always allows the standard library, so depguard is the only guard against `database/sql` in the core. It also gives a second, blunt check on core purity and bans the CGO SQLite driver everywhere:

```yaml
linters:
  enable: [depguard]
  settings:
    depguard:
      rules:
        core-purity:
          list-mode: strict
          files: [ "**/internal/domain/**", "**/internal/app/**" ]
          allow:
            - $gostd
            - github.com/DannyFestor/TuiSnip/internal/domain
            - github.com/DannyFestor/TuiSnip/internal/app
          deny:
            - pkg: database/sql
              desc: persistence belongs in internal/adapters/sqlite behind a port
            - pkg: os/exec
              desc: process spawning belongs in an adapter
            - pkg: "log$"
              desc: use log/slog
        repo-wide:
          list-mode: lax
          files: [ $all ]
          deny:
            - pkg: github.com/mattn/go-sqlite3
              desc: CGO driver; use modernc.org/sqlite
            - pkg: github.com/pkg/errors
              desc: use standard errors
            - pkg: io/ioutil
              desc: deprecated; use io and os
```

Declare the rules explicitly. With depguard enabled and no rules, golangci-lint only allows `$gostd` in every file.

**Install:** go-arch-lint has no mise registry shortname. Following the map's settled rule, use the `go:` backend in `mise.toml`:

```toml
"go:github.com/fe3dback/go-arch-lint" = "1.19.0"
```

Building it with the project's own Go toolchain matters here. The tool loads packages through `golang.org/x/tools/go/packages`, and binaries built against an older toolchain have failed to load the standard library (issue #75). Go 1.27 support only arrived in v1.18.0. Don't pin anything older than 1.18.0.

**Run:** add a Makefile target `arch-lint: ; go-arch-lint check`. It exits 1 on any warning, so CI needs no extra flags. In CI, add a step after the existing mise tool-install step that runs `make arch-lint` in the PR workflow, next to golangci-lint. `go-arch-lint graph` writes an SVG of the component graph, which is useful for `docs/`, but it isn't a gate.

## Findings

### What it can express

| Feature | Behaviour | Source |
|---|---|---|
| Components | Named sets of packages, selected by directory globs. `*` matches one level; `**` matches the base directory and every level below it. | [syntax](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/docs/syntax/README.md), [glob.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/common/path/glob.go) |
| Overlapping globs | When several components match a package, the one matching the fewest files wins (then the deeper path, then the longer name). That is why `ports` can be carved out of `app/**`. Verified locally. | [holder.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/project/holder/holder.go) |
| Allowed edges | `mayDependOn` is an explicit allow-list per component. A component's own packages are **not** implicitly allowed to import each other. Verified locally. | [allowed_project_imports.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/spec/assembler/allowed_project_imports.go), [checker_imports.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/checker/checker_imports.go) |
| Common components | `commonComponents` may be imported by every component. | [syntax](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/docs/syntax/README.md) |
| Vendors | Named groups of module import paths with globs. `canUse` grants them per component, `commonVendors` grants them everywhere, `depOnAnyVendor`/`anyVendorDeps` switch the check off. The project's own config lists both `pkg` and `pkg/**`, so this draft does the same. | [syntax](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/docs/syntax/README.md), [.go-arch-lint.yml](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/.go-arch-lint.yml) |
| Escape hatches | `anyProjectDeps` and `anyVendorDeps` per component, meant for the DI/main component. | [syntax](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/docs/syntax/README.md) |
| Unmapped packages | A Go file inside `workdir` that no component matches produces a warning, so new packages can't slip through unclassified. | [checker_imports.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/checker/checker_imports.go) |
| Exclusions | `exclude` takes directories. `excludeFiles` takes filename regexes, and a matching file excludes its whole package. | [v3.json schema](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/schema/v3.json) |
| Missing directories | A component glob with no matching directory is a hard error unless `ignoreNotFoundComponents: true`. Verified locally. | [v3.json schema](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/schema/v3.json) |
| Deep scan | AST analysis of constructor/function parameters with interface types. It reports when a concrete type from component X is injected into component Y and Y may not depend on X. It works both for interfaces declared locally (consumer-defined) and for imported `ports` interfaces. Verified locally. It can be set globally or per component. | [checker_deepscan.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/checker/checker_deepscan.go), [deepscan fixtures](https://github.com/fe3dback/go-arch-lint/tree/v1.19.0/internal/services/checker/deepscan/test/project/internal) |
| Output | ASCII or `--json`. Exit code 0 means clean, 1 means warnings. `graph` renders an SVG. A pre-commit hook and a Docker image are published. | [README](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/README.md), [graph docs](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/docs/graph/README.md), [.pre-commit-hooks.yaml](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/.pre-commit-hooks.yaml) |

### What it cannot express

- **Standard-library bans.** `ImportTypeStdLib` returns "allowed" unconditionally ([checker_imports.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/checker/checker_imports.go)). depguard covers this with `$gostd` plus `deny` entries ([depguard README](https://github.com/OpenPeeDeeP/depguard/blob/v2/README.md), [golangci-lint reference](https://github.com/golangci/golangci-lint/blob/HEAD/.golangci.reference.yml)).
- **Global deep scan with dependency inversion.** With `deepScan: true` globally, `app.NewCreateSnippet(sqlite.NewStore())` in `main` was reported as `Dependency sqlite -\-> app not allowed`. Deep scan treats runtime injection as a dependency, which is exactly the edge hexagonal architecture inverts. The fix is to enable it only on driving adapters (`tui`, `cli`). Verified locally.
- **Deep scan only runs after import checks pass.** The composite checker stops after the first checker that reports anything ([checker_composite.go](https://github.com/fe3dback/go-arch-lint/blob/v1.19.0/internal/services/checker/checker_composite.go)), so fixing an import warning can reveal new deep-scan warnings.
- **Rules on test files.** The project's own config and ours exclude `_test.go`. Test packages are unconstrained by go-arch-lint, and depguard can target them with `$test` if needed.
- **Deny-lists and wildcard edges.** `mayDependOn` takes component names only. Everything is allow-list.
- **Windows path handling** is reported broken (open issue [#79](https://github.com/fe3dback/go-arch-lint/issues/79), 2026-01-14). This doesn't matter here because Windows is out of scope.

### Maintenance status (as of 2026-09-30)

- Latest release **v1.19.0, 2026-09-07**. Earlier ones: v1.18.0 (2026-08-21, Go 1.27 support), v1.17.0 (2026-08-05), v1.16.0 (2026-07-09), v1.15.0 (2026-05-04). [releases](https://github.com/fe3dback/go-arch-lint/releases)
- Last commit **2026-09-07** (`bf473af`, merge of PR #90). [commits](https://github.com/fe3dback/go-arch-lint/commits/master)
- Not archived, MIT licensed, about 580 stars, 18 open issues. Most open issues are feature requests from 2020 to 2024. Recent releases are community PRs merged by the maintainer. [repo](https://github.com/fe3dback/go-arch-lint), [issues](https://github.com/fe3dback/go-arch-lint/issues)
- Release binaries for darwin and linux on amd64 and arm64, with `checksums.txt`. The package is also in the aqua registry. [v1.19.0 assets](https://github.com/fe3dback/go-arch-lint/releases/tag/v1.19.0), [aqua registry](https://github.com/aquaproj/aqua-registry/blob/main/pkgs/fe3dback/go-arch-lint/registry.yaml)
- depguard v2.0.0 was released 2023-03-25; the repo was last pushed 2026-09-22. It ships inside golangci-lint, so its release cadence doesn't matter here. [depguard](https://github.com/OpenPeeDeeP/depguard)

## Alternatives rejected

- **arch-go** ([repo](https://github.com/arch-go/arch-go)). Active: v2.1.2 released 2026-02-03, last push 2026-09-24. It has dependency, content and naming rules plus compliance thresholds. Rejected because it checks packages against patterns and doesn't model named components or DI injection. The map already settled on go-arch-lint, and go-arch-lint is healthy.
- **go-cleanarch** ([repo](https://github.com/roblaszczak/go-cleanarch)). Last release v1.2.1 (2021-02-18), last push 2021-11-08. Stale, and it assumes fixed layer directory names.
- **depguard only.** It can express "domain imports only stdlib", but every edge becomes a file-glob rule with no component model, no unmapped-package warning and no DI check. It's a better fit as the blunt complement.

## Open questions for the human

1. **Ports placement:** the draft assumes the use-case interfaces live in `internal/app/ports`, a separate component so driven adapters can implement ports without seeing use cases. Is that the intended layout, or should ports sit beside each use case (which would force `sqlite` to `mayDependOn: app`)?
2. **Adapter vendors not yet chosen:** clipboard library, TOML/XDG library for `config`, and the CLI parser for `cli`. Each needs a `vendors` entry plus `canUse` once chosen. Until then `depOnAnyVendor: false` makes such an import fail.
3. **Bubble Tea module path:** the `tui-framework` glob covers both `github.com/charmbracelet/**` and `charm.land/**`. Narrow it once the Bubble Tea research ticket settles the major version.
4. **Keybinding config into the TUI:** should `tui` be allowed to depend on `config`, or should `main` translate config into plain values or domain types? The draft forbids the edge.
5. **mockery output location:** if mocks live in non-test packages (e.g. `internal/app/ports/mocks`), they need a component. If they're generated into `_test.go` files, `excludeFiles` already covers them.
6. **`go:` vs `aqua:` backend:** `go:` follows the settled rule and builds with the project toolchain. `aqua:fe3dback/go-arch-lint` would give checksum-verified release binaries instead. Keep `go:`?
