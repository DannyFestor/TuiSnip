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
| pre-commit | `scripts/refuse-commit-to-main.sh`, `golangci-lint fmt` on staged Go files with `stage_fixed`, shellcheck on staged shell scripts |
| commit-msg | `cog verify` |
| pre-push | `scripts/refuse-push-to-main.sh`, `golangci-lint run` on the packages changed since `origin/main`, the unit tests. The feature and e2e tests run only with `TUISNIP_SLOW_TESTS=1`. |

The pre-push jobs are piped, so the first failure stops the rest.

pre-push lints whole packages, not single files, because golangci-lint type-checks a package as a unit. `--new-from-rev` was rejected: it reports only issues on changed lines, so it misses issues in old lines that a change affects. The unit tests run on every package, because a change can break the tests of the packages that import it. The feature and e2e tests can take minutes, so they wait for CI unless asked for.

`stage_fixed` was tested with partially staged files. lefthook hides the unstaged hunks while the hook runs and puts them back afterwards, so they are never formatted or staged. When the formatter changes a line that also has an unstaged edit, lefthook aborts the commit with `conflict while merging unstaged changes` and restores the worktree and index. Stage the whole file and commit again.

Rejected: husky (needs Node) and pre-commit (needs Python). lefthook is a single binary that mise pins.

**cocogitto 7.0.0** checks Conventional Commits. `cog verify --file` checks the commit-msg hook's message file, and `cog verify "<message>"` checks a plain string, which suits the PR-title check in CI. The default commit types are used, so there is no `cog.toml`. Only `verify` is used. Changelogs and version bumps are left to the release process.

The hook passes `--ignore-fixup-commits`, because `rebase --autosquash` folds those commits away before a PR. The first commits predate the convention, so the history is never checked, only new messages.

Rejected: commitlint (needs Node). committed and a hand-written regex script were passed over by preference: cocogitto parses the full spec, so no parsing code has to be maintained here.

## Agent hooks

Coding agents get the same guardrails as git hooks, but earlier: before an edit lands, after it, and before the agent ends its turn. The policy lives in bash scripts under `scripts/agent/`, and each agent system gets a thin adapter. The full spec is the [Agent hooks design resolution](https://github.com/DannyFestor/TuiSnip/issues/13#issuecomment-5904919452).

| Script | Runs | Does |
|---|---|---|
| `guard-path.sh` | before an edit | Denies generator output, asks before edits to guardrail files (both listed in `protected-paths`), denies files with a generated-code header |
| `guard-go-comments.sh` | before a Go edit | Denies the first submission of a comment the edit adds, so the agent checks it against the [comment rule](standards/code.md#comments). The same comment submitted again in the session asks the user (passes in OpenCode, which can't ask from a plugin). Directives, go-enum `// ENUM(` declarations, `// Output:`, `doc.go`, and generated files are exempt |
| `guard-command.sh` | before a shell command | Denies hook bypass (`--no-verify`, `commit -n`, `core.hooksPath`, `LEFTHOOK`, `LEFTHOOK_EXCLUDE`, `LEFTHOOK_CONFIG`), force-push, pushes to `main`, destructive git commands, and shell file writes, including copies, moves, links, and patches into the repo |
| `check-go-file.sh` | after a Go edit | `golangci-lint fmt` on the file, then `go vet` and the full `golangci-lint run` on its package, so lint findings arrive with the edit that causes them and not at the end of the turn. Lint is skipped while vet fails, and waits for any other golangci-lint run instead of failing on its lock |
| `verify-build.sh` | at end of turn | build, go-arch-lint, lint, and short unit tests on the packages changed since `origin/main`. Skips an unchanged Go diff, and blocks at most 3 times in a row. |

Every core script exits 0 to pass, 2 to deny, 3 to ask, with the way forward on stderr.

| System | Wiring | Adapter |
|---|---|---|
| Claude Code | `.claude/settings.json` | `scripts/agent/adapters/claude.sh` |
| OpenCode | `.opencode/plugins/tuisnip-guards.ts`, `opencode.json` | the plugin itself, plus `scripts/agent/adapters/opencode.sh` for the comment guard |
| Antigravity | `.agents/hooks.json` | `scripts/agent/adapters/antigravity.sh` (best effort, untested against a live agent) |

**jq 1.8.2** reads the hook input JSON in the bash adapters. It is pinned in `mise.toml` because the hooks don't work without it, so a fresh clone needs `mise install` before an agent edits anything.

`scripts/agent/protected-paths` is the single list of path rules. `make generate-agent-rules` turns it into OpenCode's `permission.edit` block in `opencode.json`, so OpenCode prompts for the ask tier itself, and the CI drift check catches a stale file.

The guards keep agents from making mistakes, but they aren't a security boundary. A command nested in `sh -c '...'` isn't inspected, and the deny on shell file writes exists so file changes go through the edit tools, where the post-edit format and vet run.

`scripts/agent/hooks_test.go` pipes recorded hook input from `testdata/` through the adapters in a throwaway git repo, with `go`, `golangci-lint`, and `go-arch-lint` stubbed. It runs with the unit tests. bats was rejected so the repo keeps one test runner.

Rejected: Claude Code permission rules alongside the hooks (a second copy of the path rules), and OpenCode's plugin `permission.ask` hook (legacy v1 API, and it only overrides a check that already happens).

### Shell pitfalls

Each of these has cost an agent a retry:

- Quote globs and URLs that contain `?`, such as `--include='*.go'`. zsh fails on a glob that matches nothing.
- macOS has no `timeout`.
- `tee` test output to a file, then grep the file. Rerunning the suite to grep it a second way repeats the whole run.
- Redirect to absolute paths. `guard-command.sh` can't resolve a relative redirect after `cd`, even into `/tmp`.

## CI and release

GitHub Actions runs every check the hooks run, plus the ones too slow for them. Each job calls a Makefile target, so a failing job can be reproduced locally with the same command.

| Workflow | Trigger | Jobs |
|---|---|---|
| `ci.yml` | every PR, every push to `main` | `lint` (golangci-lint and `fmt --diff`), `arch-lint`, `test` (unit, feature, e2e), `build (<os>, <arch>)` for darwin/linux × amd64/arm64, `govulncheck`, `generated` (`make generate` drift and `make fix-check`), `shellcheck` |
| `pr-title.yml` | PR opened, edited, or updated | `pr-title`: `cog verify` on the title |
| `weekly.yml` | Mondays at 03:00 UTC, and by hand | `fuzz` (5 minutes per target), `property-deep`, `mutation`. Report only, never required. |
| `release.yml` | `v*` tags | goreleaser |

The deep tests run weekly, not nightly. Daily runs would mostly search code that hasn't changed, and the cached Go build cache keeps the fuzz corpus from one run to the next.

`.github/actions/setup` installs the tools from `mise.toml` with `jdx/mise-action` and caches the Go build and module caches per job. The mise version is pinned there, because Dependabot can't update it and an unpinned mise changes under CI without a commit.

The tests run only on Ubuntu. The builds cross-compile there too, which works because nothing uses CGO. A macOS test job comes with the first platform-specific code, the clipboard adapter.

The PR-title check is its own workflow, so editing a title re-runs that check and nothing else. It runs `cog verify` without `--ignore-fixup-commits`, because the title becomes the squash commit on `main`. The title reaches the script through an env var, never interpolated, so a crafted title can't inject shell.

Third-party actions are pinned by commit SHA with the version in a comment. A tag can be moved to other code. Dependabot updates the SHA and the comment together.

The drift check runs `git add --intent-to-add` before `git diff --exit-code`, so a file a generator newly creates fails the check too, not only a changed one.

**goreleaser 2.18.2** builds tar.gz archives for darwin/linux × amd64/arm64 and `checksums.txt`, and uses GitHub's generated release notes. There is no signing and no Homebrew tap. `goreleaser release --snapshot --clean` tries a release locally without a tag.

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
- **shellcheck 0.11.0** lints the bash in `scripts/`. pre-commit checks the staged scripts, and CI checks every tracked script with `make lint-shell`. golangci-lint doesn't look at shell, and the hooks that guard everything else are shell.
- **govulncheck 1.8.0** checks the dependencies against the Go vulnerability database in CI (`make vulncheck`). It reports only vulnerabilities in code the module calls, so an unused vulnerable function doesn't fail the build.

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
| `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2` | `tui` | The current major of all three, which upgrade together. v1 gets no more development. [Research](research/bubbletea-version-and-components.md) | the v1 line (development stopped, new Bubbles features are v2-only) |
| `github.com/charmbracelet/x/ansi` | `tui` | Width and truncation with an ellipsis for ANSI-styled text. Lip Gloss builds on it but only truncates without an ellipsis. Granted with the `tui-framework` vendor. | measuring with `len` or `utf8.RuneCountInString` (wrong for wide runes and escape codes) |
| `github.com/charmbracelet/x/exp/teatest/v2` | `tui` tests | Runs a model inside a real Bubble Tea program. No tagged releases, so `go.mod` pins one commit and Dependabot bumps it. | teatest v1 (targets Bubble Tea v1) |
| `github.com/charmbracelet/x/exp/golden` | `tui` tests | The golden-file helper teatest itself uses: compares against `testdata/<Test>.golden` and rewrites it on `-update` | a hand-written golden helper |
| `modernc.org/sqlite` | `sqlite` | SQLite without CGO, so cross-compiling needs no C toolchain | `github.com/mattn/go-sqlite3` (CGO), denied by depguard |
| `github.com/pressly/goose/v3` | `sqlite` | A 50/50 call with golang-migrate, which would have worked as well. goose reads the migrations from an `fs.FS`, and `GetDBVersion` plus `ListSources` give the newer-schema check at start-up. | Atlas: more than this app needs |
| `github.com/BurntSushi/toml` | `config`, `state` | Decodes the config and reports unknown keys | `github.com/pelletier/go-toml/v2`: it reports unknown keys only as a `StrictMissingError` return, and research couldn't confirm the struct is still populated when that error comes back |
| `github.com/junegunn/fzf` (`src/algo`, `src/util`) | `memsearch` | `FuzzyMatchV2` for title, Description, and Tags, `ExactMatchNaive` for content. Its scores come from the matched span, not the field length, so the Search weights compare like for like. Folds Latin diacritics, and returns match positions as rune indexes. Adds `junegunn/go-shellwords` and `mattn/go-isatty` (both MIT). [Research](research/fuzzy-matcher.md) | `sahilm/fuzzy`: it subtracts a point per unmatched byte, so the same match scores 73 in a short field and -128 in a long one, and it ranks multi-byte text lower ([sahilm/fuzzy#30](https://github.com/sahilm/fuzzy/issues/30)). `lithammer/fuzzysearch`: no match positions, no release since 2023. An in-repo matcher: would reimplement fzf's Smith-Waterman scoring |

fzf is an application at v0.x, and `src/algo` has no API promise. The `FuzzyMatchV2` signature hasn't changed since 0.30.0 (2022), and the import stays inside `memsearch`, so a breaking bump costs an adapter change, which is accepted. `algo.Init` writes package globals, so `memsearch` calls it once, and every query gets its own slab. Dependabot needs no special rule: a v0.x bump counts as minor, so fzf's frequent releases land in the weekly grouped PR.

The **goose 3.28.0** CLI is pinned for people, not for scripts. `goose -dir db/migrations -s create <name> sql` writes the next sequentially numbered migration, and `goose sqlite3 <file> status` shows which migrations a database has applied. The app itself runs migrations through the library.

A module enters `go.mod` with the first code that imports it, because `go mod tidy` drops the rest. When this was written, only chroma and testify were in. Their indirect modules (regexp2 for chroma's lexers, yaml for testify's `assert`) aren't listed.
