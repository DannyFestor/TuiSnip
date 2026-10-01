# Testing

How to write tests in this repo. The tiers and where they live are in [architecture](architecture.md#tests). The test linters (testpackage, paralleltest, testifylint, and the `_test.go` relaxations) are in [linting](linting.md#tests). Production-code conventions are in [code](code.md).

## Choosing the tier

| Tier | Where | Tag | Covers |
|---|---|---|---|
| Unit | beside the code, `<pkg>_test` | none | one type: value objects, entities, one Action against mocked capabilities, one adapter |
| Feature | `test/feature/` | `feature` | Actions wired through `bootstrap` against a temporary SQLite file |
| e2e | `test/e2e/` | `e2e` | the whole TUI driven by teatest |

Put a test in the lowest tier that can observe the behaviour. Anything stateful is a feature test, such as whether a moved Snippet shows up in its new Folder. Unit tests have no repository to hold state; see [Test doubles](#test-doubles).

## Naming

| Kind | Name | Example |
|---|---|---|
| Unit | `Test<Type>_<Method>` | `TestCreate_Run`, `TestTitle_New` |
| Unit subtest | a behaviour phrase | `"rejects blank title"` |
| Feature | the user-visible behaviour | `TestCaptureFilesSnippetAtRoot` |
| e2e | the screen flow | `TestSearchThenCopy` |
| Property | ends in `Property` | `TestFolderTreeMoveProperty` |
| Fuzz | `Fuzz<Func>` | `FuzzParseKeyBinding` |

`make test-property-deep` selects property tests by the `Property` suffix. A property test without it never gets the deep run. Names use [`CONTEXT.md`](../../CONTEXT.md) terms.

## Structure

Every test and subtest calls `t.Parallel()` (paralleltest), so nothing a test touches may be shared.

- **Context**: use `t.Context()`. It's cancelled when the test ends.
- **Helpers** call `t.Helper()` first, so failures point at the caller.
- **Teardown** goes in `t.Cleanup`, not `defer`, so helpers can register it.
- **`require` or `assert`**: use `require` when the rest of the test can't continue without the result, such as a setup step or an error that must be nil before you read the value. Use `assert` for the checks after that.

### Tables

Use a table when cases share arrange and act and differ only in data. Otherwise write separate tests. A table is a slice, so cases run and fail in a stable order:

```go
func TestTitle_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "trims surrounding space", raw: "  curl json ", want: "curl json"},
		{name: "rejects blank title", raw: "   ", wantErr: value.ErrBlankTitle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := value.NewTitle(tt.raw)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
```

### Shared setup

testify `suite` is denied, because suites can't run in parallel. Replace a suite's `SetupTest` with a helper that returns what the test needs and registers its own cleanup:

```go
func newTestApp(t *testing.T) *bootstrap.App {
	t.Helper()

	app, err := bootstrap.New(t.Context(), bootstrap.Options{DatabasePath: filepath.Join(t.TempDir(), "tuisnip.db")})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, app.Close()) })
	return app
}
```

`bootstrap`'s API arrives with the walking skeleton. The example shows the helper pattern, not that API.

## Test doubles

| Need | Double |
|---|---|
| a capability in an Action unit test (`Inserter`, `CaptureClipboard`, `QueryIndex`) | mockery mock |
| time | `testkit.FixedClock`, or `testkit.ManualClock` when the test moves time with `Advance` |
| IDs | `testkit.SequentialIDs`, a deterministic UUIDv7 sequence |
| a valid entity to start from | a `testkit` builder |
| state that persists across calls | the feature tier, against real SQLite |

A hand-written repository fake would be a second database that slowly stops matching SQLite. Behaviour that needs state belongs in the feature tier.

### Mocks

mockery generates testify mocks from `.mockery.yml` into `mocks_test.go` beside the consumer (`make generate-mocks`). Build them with the generated `NewMock…(t)` constructor, which asserts expectations on cleanup. Set expectations with the typed `EXPECT()` API, not the string-based `On("Insert", …)`:

```go
repo := NewMockCreateRepository(t)
repo.EXPECT().Insert(mock.Anything, mock.AnythingOfType("domain.Snippet")).Return(nil)
```

### testkit

`internal/testkit` holds the fixed `Clock`, the sequential `IDGenerator`, and entity builders. depguard allows it only in `_test.go` files.

A builder takes a Spec whose zero fields mean "use a valid default". It builds through the real constructors and fails the test if they reject the Spec, so a builder never produces an entity the domain wouldn't:

```go
snippet := testkit.Snippet(t, testkit.SnippetSpec{Title: "curl json"})
```

Set only the fields the test is about. A test that states every field hides which one matters.

## Feature tests

- Build the app with `newTestApp(t)`. Each test gets its own database file in `t.TempDir()`, so parallel tests share nothing and run with production's pragmas. Why not `:memory:`: [database](database.md#tests).
- Drive and check behaviour **only through Actions**. Create a Snippet with `snippet.Create`, then read it back with `browse.SnippetsInFolder`. Raw SQL would tie the tier to the schema, and this tier exists to test behaviour.

## e2e tests

- Drive the `bootstrap`-built TUI with teatest, and wait on output with `teatest.WaitFor` and a substring check.
- A small number of golden-file snapshots guard the layout of the main screens: a saved copy of the rendered screen in `testdata/<Test>.golden`, compared with `teatest.RequireEqualOutput(t, tm.FinalOutput(t))`. Render them at a fixed terminal size (`teatest.WithInitialTermSize(120, 40)`) with the ASCII colour profile, so they don't depend on the machine running them. After an intended layout change, regenerate with `go test -tags e2e ./test/e2e/... -update` and review the snapshot diff in the PR.
- Add a golden only for layout. Flow checks use `WaitFor`, because a golden breaks on every visual change.

## Fuzz and property tests

Both are ordinary unit tests: no tag, and their seed corpus or default 100 checks run under plain `go test`.

- **Fuzz** (`testing.F`) parsers of hand-edited or foreign input: the config and keybinding loader, the Search matcher on arbitrary strings. A crasher found by the weekly job is committed to `testdata/fuzz/` as a regression seed.
- **Property** (`pgregory.net/rapid`) rules that must hold for every sequence, such as Folder moves never creating a cycle. A property that needs SQLite goes in the feature tier.

The weekly deep runs, the mutation report, and the Makefile targets are described in [the research](../research/mutation-fuzz-property-testing.md).
