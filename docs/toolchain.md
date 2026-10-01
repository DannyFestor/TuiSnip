# Toolchain

Every tool and library the repo depends on, why it's here, and what was rejected. Versions are pinned in `mise.toml` and `go.mod`. The standards say how to use each tool; this file only records the choice.

A rejected alternative is listed only when a checked fact rules it out. Where a choice came down to preference, the row says so.

## Go and mise

**Go 1.27.1**, pinned in `mise.toml`. UUIDs come from the standard library `uuid` package added in 1.27. `github.com/google/uuid` is not used.

**mise** pins every tool, so a fresh clone runs `mise install` and gets the versions CI uses.

Tools that load Go packages (golangci-lint, go-arch-lint, mockery, govulncheck) use mise's `go:` backend, which builds them from source with the pinned Go. A prebuilt binary is compiled with whatever Go its release used, and it refuses to load a module whose `go` directive is newer than that. gremlins, go-enum, and goose aren't in mise's registry, so they use the `go:` backend too.

## Make

The Makefile holds every task. Go has no built-in task runner like `npm run` or `composer run`, and Make is the usual choice in Go repos. It comes preinstalled on macOS (3.81) and on the Ubuntu 24.04 runner (4.3), so it needs nothing before `mise install`.

macOS ships GNU Make 3.81. It reads an escaped colon (`generate\:sql`) literally in `.PHONY` and in prerequisite lists, so target names use hyphens (`generate-sql`), never colons.

Rejected: Task, just, and mage. mise could pin any of them, but none comes preinstalled.

## Git hooks

**lefthook 2.1.15** runs the local hooks from `lefthook.yml`. A fresh clone runs `lefthook install` once after `mise install`.

| Hook | Job |
|---|---|
| pre-commit | `golangci-lint fmt` on staged Go files with `stage_fixed`, shellcheck on staged shell scripts |
| commit-msg | `cog verify` |
| pre-push | `scripts/refuse-push-to-main.sh` |

`stage_fixed` was tested with partially staged files. lefthook hides the unstaged hunks while the hook runs and puts them back afterwards, so they are never formatted or staged. When the formatter changes a line that also has an unstaged edit, lefthook aborts the commit with `conflict while merging unstaged changes` and restores the worktree and index. Stage the whole file and commit again.

Rejected: husky (needs Node) and pre-commit (needs Python). lefthook is a single binary that mise pins.

**cocogitto 7.0.0** checks Conventional Commits. `cog verify --file` checks the commit-msg hook's message file, and `cog verify "<message>"` checks a plain string, which suits the PR-title check in CI. The default commit types are used, so there is no `cog.toml`. Only `verify` is used. Changelogs and version bumps are left to the release process.

The hook passes `--ignore-fixup-commits`, because `rebase --autosquash` folds those commits away before a PR. The first commits predate the convention, so the history is never checked, only new messages.

Rejected: commitlint (needs Node). committed and a hand-written regex script were passed over by preference: cocogitto parses the full spec, so no parsing code has to be maintained here.

## Code generation

| Tool | Why | Rejected |
|---|---|---|
| sqlc 1.31.1 | Type-checked Go from plain SQL, no ORM. Generates against `sqltype.ID` and `sqltype.Timestamp`. | ORMs, ruled out while charting. `sqlc verify` stays out of CI because it needs sqlc Cloud. |
| mockery 3.8.0 | testify mocks with the typed `EXPECT()` API, written into `<pkg>_test` and configured once in `.mockery.yml`. Also chosen to try something other than gomock. | go.uber.org/mock (its own gomock framework, not testify, and flags on every `go:generate` directive); moq and counterfeiter (function-field fakes with no expectations) |
| go-enum 0.9.5 | `String`, `Parse…`, `IsValid`, and text marshalling for enums read from config, the database, and the state file | stringer, which only generates `String` |
| `internal/tools/languagegen` | Writes chroma's canonical lexer names into `value`, so every chroma change shows up in the drift check | a hand-maintained list (drifts from chroma silently); `value` importing chroma (core-purity forbids it); a `//go:generate` directive in `value` (`generate-enums` would run it before `generate-languages` and then again, with paths relative to `internal/domain/value`) |

mockery refuses a config with an empty `packages` map. `.mockery.yml` lists `internal/app/snippet` with no interfaces until the first mock is needed.

## Linting

- **golangci-lint 2.14.0**, `default: all`. The catalogue and the rejected linters are in [the research](research/golangci-lint-catalogue.md).
- **go-arch-lint 1.19.0** enforces the layer edges. arch-go was rejected because it matches packages against patterns and has no model of named components or injected dependencies. Details in [the research](research/go-arch-lint.md).

## Testing

| Tool | Why | Rejected |
|---|---|---|
| testify `assert` and `require` | `require` stops the test, `assert` continues. testifylint checks the idioms, and mockery's mocks are testify mocks, so one library covers both. `suite` is denied because suites can't run in parallel. | matryer/is (no release since 2023); google/go-cmp (compares values, no `require`/`assert` split); gotest.tools/v3 (no idiom linter). testifylint checks only testify. |
| `pgregory.net/rapid` | Property tests with shrinking and state machines. Enters `go.mod` with the first property test. | gopter (no release since 2020), `testing/quick` (no shrinking, no state machines) |
| gremlins 0.6.0 | Mutation testing as a separate CLI, so it needs no build tag and never touches `go.mod` | go-mutesting (dormant or abandoned); ooze is the fallback; gomu is being watched. See [the research](research/mutation-fuzz-property-testing.md). |

`make test-property-deep` sets `RAPID_CHECKS` rather than passing `-rapid.checks`. Every test binary that doesn't import rapid rejects the flag and fails, and rapid reads the environment variable as its default.

Avoid gremlins' `-i` until [gremlins#272](https://github.com/go-gremlins/gremlins/issues/272) is fixed. It reports killed mutants as lived.

## Libraries

| Module | Used by | Why | Rejected |
|---|---|---|---|
| `github.com/alecthomas/chroma/v2` | `tui`, `languagegen` | The standard Go highlighter, pure Go, and the one glamour uses. Provides the Language list and each Language's filename patterns. v3 is in alpha: its upgrade can rename or drop Languages, which shows up in the drift check and needs a data migration for stored Fragments. | tree-sitter bindings (`smacker/go-tree-sitter`, `tree-sitter/go-tree-sitter`): both need CGO |
| `modernc.org/sqlite` | `sqlite` | SQLite without CGO, so cross-compiling needs no C toolchain | `github.com/mattn/go-sqlite3` (CGO), denied by depguard |
| `github.com/pressly/goose/v3` | `sqlite` | A 50/50 call with golang-migrate, which would have worked as well. goose reads the migrations from an `fs.FS`, and `GetDBVersion` plus `ListSources` give the newer-schema check at start-up. | Atlas: more than this app needs |
| `github.com/BurntSushi/toml` | `config`, `state` | Decodes the config and reports unknown keys | `github.com/pelletier/go-toml/v2`: it reports unknown keys only as a `StrictMissingError` return, and research couldn't confirm the struct is still populated when that error comes back |

A module enters `go.mod` with the first code that imports it, because `go mod tidy` drops the rest. When this was written, only chroma and testify were in.
