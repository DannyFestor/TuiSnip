# Testing

How to write tests in this repo. The tiers and where they live are in [architecture](architecture.md#tests). The test linters (testpackage, paralleltest, testifylint, and the `_test.go` relaxations) are in [linting](linting.md#tests). Production-code conventions are in [code](code.md).

## Choosing the tier

| Tier | Where | Tag | Covers |
|---|---|---|---|
| Unit | beside the code, `<pkg>_test` | none | one type: value objects, entities, one Action against mocked capabilities, one adapter |
| Feature | `test/feature/` | `feature` | Actions wired through `bootstrap` against a temporary SQLite file |
| e2e | `test/e2e/` | `e2e` | the whole TUI driven by teatest |
| Platform | beside the code, in `*_platform_test.go` | `platform` plus the OS (`darwin && platform`) | one adapter against the real OS tool, such as `pbcopy` |

Put a test in the lowest tier that can observe the behaviour.

Platform tests touch state the user owns, like the system clipboard, so they only run when asked: `make test-platform` locally, and in CI's `test-macos` job. They restore what they change, and they don't run in parallel. Anything stateful is a feature test, such as whether a moved Snippet shows up in its new Folder. Unit tests have no repository to hold state; see [Test doubles](#test-doubles).

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
func (h *Home) Start(t *testing.T, tool ClipboardTool) *bootstrap.App {
	t.Helper()

	app, err := h.Open(t, tool)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, app.Close()) })

	return app
}
```

A test builder that more than one test file in a package uses lives in that package's `helpers_test.go`: `insertSnippet` in `sqlite`, `browsingModel` and `filedIn` in `tui`, `opened` and `browsing` in `mainscreen`, `create` in `test/feature`, `seedNestedFolders` in `test/e2e`. A builder both tiers use lives in `test/testapp`. A builder only one file uses stays in that file. A fake with its own type keeps its own file, such as `fakeTools` in clipboard's `fake_tools_test.go`.

### export_test.go

A package's tests are black-box (testpackage), so they see only its exported API. An `export_test.go` in the package itself (`package config`) may expose an unexported fact to the package's own black-box tests, such as the list of key names `config` accepts. It never exposes a way around the API, such as a constructor that skips validation. It is compiled only into the package's tests, so nothing else can reach it.

## Test doubles

### What to cover

Test every behaviour a ticket or standard decided, and every branch that chooses between outcomes: Root or Folder, roll back or commit, refuse start-up or continue, log or stay quiet. Don't test a branch that only passes an I/O error up (`if err != nil { return fmt.Errorf("…: %w", err) }` after `Commit`, `Close`, or `os.Remove`). Reaching it needs a fake that fails on cue, which this section rules out, and wrapcheck and errcheck already make sure the error is wrapped and not dropped.

Coverage is reported and has no threshold. A branch that shows red is fine when it is pass-through, and a gap when it is a decision. Mutation testing (`make test-mutation`) finds the second kind: a mutant that lives on a decision branch needs a test. Before opening the PR for a ticket that adds decision logic, run `make test-mutation-changed` and add the missing tests in the same PR. It mutation-tests the packages changed since `origin/main`, uncommitted changes included, skipping a package whose parent's run already covers it, and does nothing when no Go code changed. Each package gets its own `mutation-report-<package>.json`. `make test-mutation PKG=./internal/adapters/sqlite` runs one package.

### Doubles

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

- Build the app with `test/testapp`, which the e2e tier shares: `testapp.Start(t, testapp.RecordingTool)`, or `testapp.NewHome(t)` when the test writes a `config.toml` or blocks a directory first. Each test gets its own HOME in `t.TempDir()`, so parallel tests share nothing and run with production's paths and pragmas. Why not `:memory:`: [database](database.md#tests).
- The clipboard is faked at the OS boundary, not at the Copy capability. `bootstrap.Options` takes `Environ`, `LookPath`, and `GOOS`. The harness passes `darwin` and a `LookPath` that resolves `pbcopy` to a script that records its stdin (`RecordingTool`), exits 1 (`FailingTool`), or isn't there (`NoTool`). The real backend switch and the clipboard adapter run on every OS.
- Drive and check behaviour **only through Actions**. Create a Snippet with `snippet.Create`, then read it back with `browse.SnippetsInFolder`. Raw SQL would tie the tier to the schema, and this tier exists to test behaviour.
- Seed with `testapp.SeedFolder`, `testapp.SeedTag`, and `testapp.SeedSnippet`. Each runs the Create Action and returns the entity as a browse Action reads it back, so it compares equal to what later reads return. Read one entity back with `testapp.StoredFolder`, `testapp.StoredTag`, `testapp.StoredSnippet`, or `testapp.FindSnippet` when it may be gone. `bootstrap.App` exposes only Actions, so there is no way around them.

## TUI unit tests

- Build settings with `testsettings.Default(t)`, the keys the app starts with. A test about a remapped key changes that one entry in `Keys`. Never import `config` or write the default keys out by hand.
- Build key presses with `test/keypress`. Never write a local helper or a `tea.KeyPressMsg` literal for a key it covers.
- Test the `tui` model with mocked Actions. Send it messages through `Update` and run the returned commands in the test, so every step is synchronous and the order never depends on timing.
- Test a stack member (an Overlay) through `test/overlaytest`, which pushes it onto a real Overlay stack. Its tests see the confirmations it opens, the outcomes it reports, and whether it closes, all without Model.
- Check what the user sees: the rendered `View()` with ANSI stripped, or the messages the model emitted (`tea.SetClipboard`, `tea.QuitMsg`). Not the model's fields.
- A small number of golden-file snapshots guard the layout of the main screens. Each is the final `View()`, ANSI stripped, at 120×40, saved in `testdata/<Test>.golden` and compared with `golden.RequireEqual` from `github.com/charmbracelet/x/exp/golden`. The data comes from mocks, so dates and IDs never change. A golden of teatest's output stream would hold every intermediate frame, which depends on when asynchronous messages arrive. After an intended layout change, regenerate with `go test ./internal/adapters/tui/... -update` and review the snapshot diff in the PR.
- The colour schemes are the one exception to stripping ANSI: `TestDarkSchemeLayout` and `TestLightSchemeLayout` keep the escape codes, because colour is what they guard. `golden.RequireEqual` quotes the codes in its diff.
- Add a golden only for layout. Behaviour checks use substring assertions, because a golden breaks on every visual change.
- One teatest test runs the model inside a real Bubble Tea program, to catch what the synchronous driver can't, such as a model that never quits.

## e2e tests

- Drive the `bootstrap`-built TUI with teatest, and wait for a substring of the last full frame (`waitForFrame`). Bubble Tea writes only the cells that changed, so a word rarely reaches the output stream in one piece, and `teatest.WaitFor` on the stream times out.
- Wait for text that only the next state shows. "Reclaim" is already on screen in the edit overlay's Description field before the save finishes. "Root · plaintext" in the Snippet pane only shows once it has.
- e2e tests check flows, not layout. The layout goldens live with the `tui` unit tests.

## Fuzz and property tests

Both are ordinary unit tests: no tag, and their seed corpus or default 100 checks run under plain `go test`.

- **Fuzz** (`testing.F`) parsers of hand-edited or foreign input: the config and keybinding loader, the Search matcher on arbitrary strings. A crasher found by the weekly job is committed to `testdata/fuzz/` as a regression seed.
- **Property** (`pgregory.net/rapid`) rules that must hold for every sequence, such as Folder moves never creating a cycle. A property that needs SQLite goes in the feature tier.

The weekly deep runs, the mutation report, and the Makefile targets are described in [the research](../research/mutation-fuzz-property-testing.md).
