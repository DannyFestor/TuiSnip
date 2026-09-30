# Linting

`.golangci.yml` is the source of truth. This document explains the rules whose intent isn't obvious from the config. The catalogue and the rejected alternatives are in [the research](../research/golangci-lint-catalogue.md).

Run `golangci-lint run` to lint and `golangci-lint fmt` to format. CI lints the whole repo on every PR.

## Everything on, by exclusion

`default: all` turns on every linter, and `disable` lists the ones that are off, each with a reason. A golangci-lint upgrade can bring new linters, and those show up as a failing PR. That's intended: the version is pinned in `mise.toml`, so an upgrade is always deliberate.

A linter is only disabled when it is deprecated, targets a library TuiSnip doesn't use, duplicates an enabled linter, or has no defect signal (formatting taste). Don't disable a linter to make a PR pass.

## Small functions, happy path

The thresholds enforce one responsibility per function:

| Linter | Limit |
|---|---|
| funlen | 50 lines, 30 statements |
| cyclop | complexity 10 |
| gocognit | cognitive complexity 15 |
| nestif | 2 |
| revive `argument-limit` | 4 parameters |
| revive `function-result-limit` | 3 results |
| interfacebloat | 5 methods |
| dupl | 100 tokens |

nestif at 2 lets through one `if` inside another. An `if`/`else` nested inside an `if`, or three levels deep, fails. Together with revive's `early-return`, `indent-error-flow`, and `superfluous-else`, it forces guard clauses and early returns:

```go
// ❌ nested
if cfg != nil {
	if cfg.Editor != "" {
		return cfg.Editor
	} else {
		return defaultEditor
	}
}

// ✅ happy path
if cfg == nil || cfg.Editor == "" {
	return defaultEditor
}
return cfg.Editor
```

More than four constructor arguments means the Action should be split, or should take an `<Action>Deps` struct (see [architecture](architecture.md)).

## Struct literals are complete

exhaustruct checks every struct literal of a TuiSnip type, sqlc `Params` included. Each field is named, even when its value is zero:

```go
// ❌ FolderID and Language missing
snippet.CreateInput{Title: "curl json", Content: body}

// ✅
snippet.CreateInput{Title: "curl json", Content: body, FolderID: nil, Language: domain.LanguageGo}
```

When a field is added, every literal that doesn't set it fails lint, so no caller silently leaves it at zero. Types from other modules aren't checked. Test files are exempt.

## No globals, no hidden state

- gochecknoglobals and gochecknoinits: no package-level variables, no `init()`. Lip Gloss styles and default keymaps come from constructors (`DefaultKeyMap()`, `NewStyles(theme)`) and are injected, because keybindings and themes are configurable.
- forbidigo:
  - bans `fmt.Print*`, `print`, and `println`, because stdout belongs to the TUI
  - bans `time.Sleep` outside tests
  - bans `context.Background` and `context.TODO` outside `cmd/`, because there is one root context from `main`
- containedctx: no `context.Context` fields in structs. The one exception is the TUI root model (see [architecture](architecture.md)).
- nonamedreturns: no named results, except where a deferred function changes the returned error (`defer func() { err = errors.Join(err, rows.Close()) }()`).

## Errors

- wrapcheck: every error from another package or interface is wrapped, with no interface exceptions. Format: `"<pkg>.<Symbol>: %w"`.
- err113: no dynamic errors (`errors.New` inside a function). Declare sentinel `Err…` variables and wrap them.
- errorlint: compare errors with `errors.Is` and `errors.As`.
- nilnil: a function never returns `nil, nil`.

## Imports

depguard covers what go-arch-lint can't, because go-arch-lint always allows the standard library:

- **core-purity**: non-test code in `internal/domain` and `internal/app` may import only the standard library and `internal/domain`. Even there, `database/sql`, `os/...`, `net/...`, `log`, and `log/slog` are denied. depguard matches by prefix, so `os` also covers `os/exec`. Test files are excluded so black-box tests can use testify.
- **repo-wide**:
  - `github.com/mattn/go-sqlite3`: CGO; use `modernc.org/sqlite`
  - `github.com/pkg/errors`
  - `io/ioutil`
  - `math/rand` and `math/rand/v2`: use `crypto/rand`
  - testify `suite`: suites can't run in parallel

importas pins `charm.land/bubbletea/v2` to `tea`.

## Tests

- testpackage: tests are black-box `_test` packages.
- paralleltest: every test and subtest calls `t.Parallel()`. Feature tests each open their own `:memory:` database.
- testifylint, with all checks on: use `require` when the rest of the test depends on the result, `assert` otherwise.
- Relaxed in `_test.go` files: dupl, err113, exhaustruct, funlen, goconst, varnamelen, wrapcheck. Their premise doesn't hold for table tests.
- `run.build-tags` lists `feature` and `e2e`. Without them, `test/feature` and `test/e2e` are never linted.

## Comments and naming

- The `comments` preset turns off "exported X should have a comment". Names document the code; comments explain why.
- staticcheck runs every check except the doc-comment ones, so ST1003 (`ID` not `Id`, no underscores) and ST1016 (consistent receiver names) are on.
- revive `exported` flags stutter such as `search.SearchIndex`.
- godox fails on `TODO`, `FIXME`, and `BUG`. Open an issue instead.
- godot: comments end with a period.
- funcorder: a type comes first, then its constructors, then exported methods, then unexported methods.

## `nolint`

A `//nolint` names its linter and gives a reason: `//nolint:gosec // G204: the editor command comes from the user's own config`. nolintlint rejects directives with no linter or no reason, and directives that suppress nothing. No preset silences whole gosec rules, so each process launch (G204) and file read from a variable path (G304) gets its own reasoned directive.

## Generated code

`generated: strict` skips files with the `// Code generated ... DO NOT EDIT.` header, for both linters and formatters. Such files are still type-checked. Fix the generator's input, never the output.

## Formatting

Formatting uses gofumpt, goimports (TuiSnip imports grouped last), and golines at 120 columns. golines doesn't touch comments. Formatting runs only through `golangci-lint fmt`, and lefthook runs it on staged files before each commit. `golangci-lint run` also reports unformatted files, so CI needs no separate formatting step. There's no line-length linter: golines wraps what it can, and the complexity limits keep the rest short.
