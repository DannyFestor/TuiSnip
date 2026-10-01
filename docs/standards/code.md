# Code

This document holds the conventions no linter checks, and the reasons behind the ones a linter enforces in a way that isn't obvious. Read it before writing Go in this repo.

- Layering, Actions, constructors, interfaces, error wrapping, where logging and context go: [architecture](architecture.md).
- Everything golangci-lint enforces (function size, no globals, `nolint` format, formatting): [linting](linting.md).
- Tests: [testing](testing.md).

## Vocabulary

[`CONTEXT.md`](../../CONTEXT.md) is binding for code. Types, Actions, fields, test names, and log keys use its terms: `Snippet`, `Fragment`, `Folder`, `Root`, `Tag`, `Capture`, `Copy`. Each entry's `_Avoid_` list names the words to leave out of identifiers.

```go
// ❌
func (r *Repository) ListUnfiled(ctx context.Context) ([]domain.Snippet, error)
type ClipItem struct{}

// ✅
func (r *Repository) SnippetsAtRoot(ctx context.Context) ([]domain.Snippet, error)
```

The Root is spelled `Root` in code, as in `AtRoot()` and `folder.MoveToRoot`. In Go it is the zero `domain.FolderID`, never a nil pointer, and code asks `AtRoot()` rather than comparing IDs. A new concept gets its `CONTEXT.md` entry before it gets code.

## Files and packages

- One exported type per file, named after the type in snake_case: `snippet.Create` in `create.go`, `snippet.DefaultLanguageFinder` in `default_language_finder.go`. The type's constructors and methods sit in the same file, in funcorder order. Its tests go in the mirroring `_test.go` file (`create_test.go`).
- Every package has a `doc.go` whose package comment says why the package exists: what it owns and what it keeps out.

```go
// Package clipboard owns the platform clipboard tools, so that no other package
// shells out and the Copy backend is chosen in one place.
package clipboard
```

## Actions

### No ambient inputs

An Action's behaviour depends only on its Input and its injected dependencies. Time, IDs, a configured limit, anything that varies arrives through the constructor or the Input. Same Input plus same dependencies gives the same result, which is why a fixed `Clock` and `IDGenerator` make even `snippet.Create` repeatable in tests.

```go
// ❌ the Action picks its own limit and reads the clock
func (q *Query) Run(ctx context.Context, in QueryInput) ([]Hit, error) {
	return q.index.Search(ctx, in.Text, 50, time.Now())
}

// ✅ the limit comes from bootstrap, the time from Clock
func (q *Query) Run(ctx context.Context, in QueryInput) ([]Hit, error) {
	return q.index.Search(ctx, in.Text, q.limit, q.clock.Now())
}
```

The same reason keeps timeouts out of Actions (see [Context](#context)).

### Inputs carry raw values

`<Action>Input` fields are the primitives an adapter holds: the text the user typed, the ID the TUI selected. The Action parses them into [value objects](#value-objects) and reports every problem at once. The TUI and the future CLI pass text through and get identical validation.

```go
type CreateInput struct {
	Title    string
	FolderID domain.FolderID
	Language string
	Content  string
}
```

An ID is already a domain type when the TUI holds it, so Inputs carry it as one. The zero `FolderID` means the Root.

Optional parameters are Input or `<Action>Deps` fields, never functional options (`WithX(...)`). exhaustruct makes every caller state every field, so there is no hidden default.

## Domain types

### Entities

Snippet, Fragment, Folder, and Tag keep their fields unexported and have value semantics. A validating constructor builds them, accessors read them, and a mutator returns a new value. When a mutator's rule can fail, it returns the new value and an error, and the original is untouched either way. Values are also safe to hand to concurrent `tea.Cmd`s.

```go
// ❌ any package can build a Snippet with no Fragment
type Snippet struct {
	Title     string
	Fragments []Fragment
}

// ✅
type Snippet struct {
	title     value.Title
	fragments []Fragment
	updatedAt time.Time
}

func (s Snippet) Rename(title value.Title, now time.Time) Snippet {
	s.title = title
	s.updatedAt = now
	return s
}
```

`Rename` can't fail because a `value.Title` is valid by construction. Entity invariants that span fields ("exactly one Fragment") are checked in `domain` and reported with `domain` sentinels.

- Each entity's ID is its own type over `uuid.UUID` (`domain.SnippetID`, `domain.FragmentID`, `domain.FolderID`), so one can't be passed where another belongs. A constructor rejects the Nil UUID as the entity's own ID.
- `NewSnippet` and `NewFragment` take every field positionally. revive's `argument-limit` is lifted for `internal/domain` only, because the distinct value and ID types already catch most swapped arguments. The same constructor serves `snippet.Create` and the `sqlite` rebuild from rows.
- The constructors also repeat the schema's timestamp CHECKs (after the epoch, storable as `int64` nanoseconds, `updatedAt >= createdAt`), so the domain never accepts a value a save would reject.

### Value objects

A value with a normalisation or validation rule gets its own type in `internal/domain/value`: `value.Title`, `value.Description`, `value.Content`, `value.FolderName`, `value.TagName`, `value.Language`. A value with no rule stays a plain type. A length cap counts as a rule: Title (200 runes), FolderName (200 runes), Description (2,000 runes), and Content (256 KiB) are capped, and the schema repeats each cap ([database](database.md#constraints)). Description and Content are stored exactly as given, because whitespace in code matters and an all-blank Description is the user's choice.

Each value object wraps an unexported field, so the constructor is the only way to get one:

```go
// ❌ value.Title("   ") compiles anywhere and skips the rule
type Title string

// ✅
type Title struct{ value string }

func NewTitle(raw string) (Title, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Title{}, ErrBlankTitle
	}
	return Title{value: trimmed}, nil
}

func (t Title) String() string { return t.value }
```

- The constructor normalises, and the normalised form is the one stored. `Title` and `FolderName` are trimmed. `TagName` is trimmed, keeps the user's spelling, and has `Key()` returning the case-folded form. SQLite's unique index and memsearch both compare on `Key()`. `Language` is parsed from the known set by exact, case-sensitive match: chroma's canonical lexer names, generated into `value` because `value` may not import chroma. A chroma upgrade that renames or drops a name shows up in the drift check's diff and needs a data migration for stored Fragments ([database](database.md#conventions)).
- `value` imports only the standard library, and `domain` imports `value`. The sentinel for a value rule lives in `value` beside the rule, prefixed `value:` (`errors.New("value: title is blank")`). `domain` keeps `ErrNotFound`, `ErrConflict`, `ErrMissingDependency`, `ErrCorruptRecord`, and the entity-invariant sentinels.

### Validation

Every input rule from [`docs/spec/v1.md`](../spec/v1.md) lives in `domain` or `domain/value`, and only there. An adapter that wants early feedback calls the value constructor. The TUI does this to hint while the user types.

An Action reports every broken rule at once. It wraps each failure in a `domain.FieldError` naming the Input field, joins them, and wraps the result with its own name:

```go
func (c *Create) parse(in CreateInput) (value.Title, value.Language, error) {
	title, titleErr := value.NewTitle(in.Title)
	language, languageErr := value.NewLanguage(in.Language)
	return title, language, errors.Join(
		domain.OnField(domain.FieldTitle, titleErr),
		domain.OnField(domain.FieldLanguage, languageErr),
	)
}
```

`domain.OnField` returns nil for a nil error, so `errors.Join` drops the fields that parsed. The TUI marks each field from `domain.FieldErrors(err)`, which walks the error tree. It uses `errors.AsType[domain.FieldError]` when it needs only the first one. `Field` is a go-enum.

Data that doesn't come from the user is checked at its own boundary:

- `config` and `state` validate their files against their own schemas. Those rules aren't domain rules.
- `sqlite` rebuilds entities from rows through the same value constructors. There is no constructor that skips validation. A row that fails becomes an error wrapping `domain.ErrCorruptRecord`, and the log line carries the row's ID. A bug in an older build then shows up as an error, not as a broken Snippet on screen.

## Errors

Wrapping and sentinels: [architecture](architecture.md#errors). Beyond that:

- Production code never panics. Every failure is a returned error, including the "can't happen" branch of a switch, which exhaustive forces you to write. forbidigo bans `panic` outside tests. `Must…` helpers exist only in `_test.go` files. There's no `recover` of our own, because Bubble Tea already recovers and restores the terminal.
- Unpack typed errors with `errors.AsType[T](err)`. `modernize` flags `errors.As` with a target variable.

## Context

Root context, the parameter position, and the one `containedctx` exception: [architecture](architecture.md#context).

### Deadlines

The adapter that launches an outside process owns its deadline, declared as a named constant in the adapter. `wl-copy` can hang forever when the compositor doesn't focus it, and only the clipboard adapter knows that.

```go
func (t *Tool) Copy(ctx context.Context, text string) error {
	ctx, cancel := context.WithTimeout(ctx, copyToolTimeout)
	defer cancel()
	// ...
}
```

- Actions never set a timeout. They have no ambient inputs, and a deadline is policy about the outside world.
- SQLite queries get no per-call timeout. The database is local, and lock waits are bounded by the `busy_timeout` pragma.
- The `$EDITOR` run gets no timeout, because the user is in it.

### Cancellation

An Action doesn't check `ctx.Err()` itself. It passes `ctx` to every capability and returns the wrapped error when one fails. An Action that writes more than once does it in one transaction, so a cancel rolls back and leaves nothing half-written. The TUI treats `errors.Is(err, context.Canceled)` as quiet: no status message, and a Debug log line.

### No context values

A context carries cancellation and deadlines only. Loggers, IDs, and settings are passed explicitly. forbidigo bans `context.WithValue`.

## Concurrency

Bubble Tea runs each `tea.Cmd` on its own goroutine, and that is the only concurrency in production code. There are no `go` statements. If fan-out is ever needed, use `errgroup` with the Action's `ctx`.

Anything a Cmd can reach may be called concurrently: the memsearch index, repositories, the clipboard adapter. Make each safe for concurrent use with a `sync.RWMutex` field (never a package-level one), and cover it with a test that `go test -race` exercises.

## Logging

Who logs and how the logger arrives: [architecture](architecture.md#logging). What to log:

| Level | When | Example |
|---|---|---|
| Error | an error the TUI shows the user | save failed with `ErrConflict` |
| Warn | a fallback was taken | OSC 52 instead of a clipboard tool; an unknown config key ignored |
| Info | a lifecycle event | start, resolved paths, migrations applied, shutdown |
| Debug | adapter detail | clipboard tool chosen, command run, a cancelled context |

- Attributes are typed (`slog.String`, `slog.Int`, `slog.Any`), and keys are snake_case constants. sloglint enforces all three (`attr-only`, `no-raw-keys`, `key-naming-case`), plus static messages and the `…Context` methods.
- Log IDs, counts, paths, and Languages. Snippet content, Descriptions, and clipboard text stay out of every log line, because snippets often hold secrets.

```go
// ❌
log.InfoContext(ctx, "captured "+text)

// ✅
log.InfoContext(ctx, "snippet captured", slog.String(keySnippetID, id.String()), slog.Int(keyLength, len(text)))
```

## Comments

Names say what the code does. A comment says why it has to be this way: the constraint, the bug it avoids, the reason for an odd choice. A package comment in `doc.go` says why the package exists.

```go
// ❌ restates the code
// Trim the title.
trimmed := strings.TrimSpace(raw)

// ✅ explains a choice a reader would question
// wl-copy hangs when the compositor never focuses its surface, so every run
// gets a deadline.
ctx, cancel := context.WithTimeout(ctx, copyToolTimeout)
```

## Generated code

Never edit a file whose header says `// Code generated ... DO NOT EDIT.`. Change the source and regenerate. Files generated by go-enum declare package-level maps and `ErrInvalid…` variables. They're the one place globals exist, and golangci-lint skips generated files.

| Target | Runs | Source |
|---|---|---|
| `make generate-sql` | `sqlc generate` | `db/queries/*.sql`, `sqlc.yaml` |
| `make generate-mocks` | `mockery` | `.mockery.yml` |
| `make generate-enums` | `go generate ./...` | a `//go:generate go-enum` directive in the file declaring the enum |
| `make generate-languages` | the Language generator | chroma's lexer registry, into `internal/domain/value/languages_gen.go` |
| `make generate` | all four, in that order | |

Enums use go-enum only. It generates `String`, `Parse…`, `IsValid`, and, with `--marshal`, text marshalling, which the enums read from config, the database, and the state file need:

```go
//go:generate go-enum --marshal

// ENUM(title, folder_name, tag_name, language)
type Field string
```

mockery is configured once in `.mockery.yml`, not per interface. It writes `mocks_test.go` beside the consumer, in package `<pkg>_test` (see [testing](testing.md#test-doubles)).

CI runs `make generate` and fails if `git diff` shows changes.

### go fix

`make fix` runs `go fix -tags feature,e2e ./...`. Run it after a Go or dependency upgrade. `modernize` already reports every other `go fix` rewrite, but not `inline`, which follows a library's `//go:fix inline` directives from a deprecated function to its replacement. CI runs `go fix -diff -tags feature,e2e ./...`, which exits non-zero on any pending rewrite. After a Dependabot bump, that PR fails and shows the exact rewrite.
