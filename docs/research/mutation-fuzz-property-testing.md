# Mutation, fuzz, and property-based testing in Go

Research for [#5](https://github.com/DannyFestor/TuiSnip/issues/5). Sources checked 2026-09-30. Local toolchain for flag checks: go1.27.1 darwin/arm64.

## Question

Which mutation testing tools exist for Go, how mature and how slow are they, and how do they fit build tags and a nightly or manual CI job? How do Go's native fuzzing and property-based testing with `pgregory.net/rapid` work, and which TuiSnip code does each suit?

## Recommendation

Adopt all three. Each one comes in when the code it targets exists.

| Technique | Tool | Adopt when | Runs on PRs | Runs deep |
|---|---|---|---|---|
| Fuzzing | Native `go test -fuzz` | With the first parser: the TOML config and keybinding loader in the scaffold | Yes. The seed corpus runs as ordinary unit tests under plain `go test` | Nightly, `-fuzztime` per target |
| Property-based | `pgregory.net/rapid` v1.3.0 | When Folder tree logic lands (create, move Folder or Snippet to a Folder or the Root) | Yes, 100 checks per property (the default) | Nightly, `-rapid.checks=10000` |
| Mutation | gremlins v0.6.0 (ooze as fallback) | After the walking skeleton, once `internal/domain` and `internal/app` have real tests | Never | Nightly and `workflow_dispatch`, report only |

**Build tags.** Fuzz and rapid tests are ordinary unit tests in the black-box `_test` packages beside the code. They need no tag, because their normal-mode run is cheap and deterministic. A property that drives use cases through in-memory SQLite belongs in `test/feature/` under the `feature` tag, following the existing tiers. Mutation testing gets no test-file tag with gremlins. It is a separate CLI, so `go test ./...` never runs it. `--tags feature` makes gremlins' per-mutant test runs include the feature tier where that tier sits in the same package. (If "mutation behind its own tag" has to mean a literal `//go:build mutation` file, pick ooze; see Open questions.)

**Makefile targets.**

- `test-fuzz`: loop over every `Fuzz*` found with `go test -list '^Fuzz' ./...` and run `go test -run='^$' -fuzz="^$$name$$" -fuzztime=$(FUZZTIME) $$pkg` one at a time. `-fuzz` accepts exactly one package and one fuzz test per invocation. Default `FUZZTIME=30s` locally.
- `test-property-deep`: `go test ./... -tags feature -run 'Property' -rapid.checks=10000`. This needs a naming convention: property tests end in `Property`, for example `TestFolderTreeMoveProperty`.
- `test-mutation`: `gremlins unleash --tags feature --exclude-files '<generated-file regex>' --output mutation-report.json ./...`. Keep the thresholds at 0 (the default), so the run reports and never fails.

**CI schedule.** Add one `nightly.yml` with `schedule` (cron) plus `workflow_dispatch` and three independent jobs:

1. fuzz: a matrix over fuzz targets, `FUZZTIME=5m` each, on ubuntu amd64. Coverage-guided fuzzing needs amd64 or arm64. On failure, upload `testdata/fuzz/` as an artifact. A human commits the crasher as a regression seed.
2. property-deep: `make test-property-deep`. On failure, upload `testdata/rapid/`.
3. mutation: `make test-mutation`. Upload the JSON report. The job is never a required check.

Pin gremlins in `mise.toml`. It ships versioned release binaries, so it doesn't touch `go.mod`.

## Findings

### Fuzzing (native, `testing.F`)

**What it is.** You write a function that takes arbitrary bytes or strings and must never misbehave, for example by panicking, hanging, or breaking a round trip. The Go toolchain mutates inputs and watches code coverage to steer toward new branches. When it finds a failing input, it shrinks the input and saves it. [1]

**Mechanics.**
- Signature `func FuzzXxx(f *testing.F)` in a `_test.go` file. The fuzz test calls `f.Fuzz(func(t *testing.T, ...))` exactly once. Arguments are limited to `string`, `[]byte`, sized ints and uints, floats, and `bool`. [1]
- The seed corpus is `f.Add(...)` calls plus files in `testdata/fuzz/FuzzXxx/`. Plain `go test` runs every seed entry as a regular test. [1]
- The generated corpus lives in `$GOCACHE/fuzz` and is only used in fuzzing mode. [1]
- A failing input is written to `testdata/fuzz/FuzzXxx/` and "will now be run by default with `go test`, serving as a regression test". [1]
- `-fuzz` "must match exactly one package within the main module, and regexp must match exactly one fuzz test". `-fuzztime` defaults to running forever. `-fuzzminimizetime` defaults to 60s. [2]
- Coverage instrumentation, which is what makes the corpus grow, exists only on AMD64 and ARM64. [1]
- Targets "should be fast and deterministic" and must not depend on global state, because workers run them in parallel. [1]
- Continuous fuzzing: OSS-Fuzz supports native Go fuzz tests. Go's own continuous-fuzzing support is still tracked as golang/go#50192. [1]

**Suits TuiSnip for:**
- The TOML config and keybinding parser (hand-edited, so malformed input is the normal case). Properties: no panic; every accepted config round-trips; every rejection returns an error, never a zero config.
- Language auto-guessing on clipboard paste (arbitrary text in, never panics, always returns a known Language).
- The Search fuzzy matcher on arbitrary query and content strings (no panic, no out-of-range index on multi-byte runes).
- Future importers (SnippetsLab export) when they land on the roadmap.

### Property-based testing (`pgregory.net/rapid`)

**What it is.** Example tests assert fixed cases, such as "moving Folder A under B gives this tree". A property test states a rule that must hold for every input, such as "no sequence of moves ever creates a cycle". The library generates hundreds of random inputs or operation sequences. When the rule breaks, it shrinks the failure to a minimal reproduction. [3]

**Mechanics.**
- `rapid.Check(t, func(t *rapid.T) {...})` with typed generators (`rapid.Int()`, `rapid.SliceOf(...)`, `rapid.Make[T]()`). [3]
- State machine testing: `t.Repeat(map[string]func(*rapid.T))` runs a random sequence of named actions against the real implementation and a simple model. The `""` key is an invariant checked after every action. [3][4]
- Defaults: 100 checks, 30 steps per `Repeat`, 30s of shrinking. The flags are `-rapid.checks`, `-rapid.seed`, `-rapid.failfile`, `-rapid.nofailfile`, `-rapid.shrinktime`, and `-rapid.v`, with `RAPID_*` environment variables as equivalents. [5]
- A failure writes a fail file to `testdata/rapid/<TestName>/`, and later runs automatically replay it. [6]
- `rapid.MakeFuzz` turns a rapid property into an `f.Fuzz` target, so one property can also be coverage-guided fuzzed. [3]
- Rapid's README positions it against the alternatives: it beats `testing.F` at structured data and state machines, has a simpler API than gopter with better shrinking, and `testing/quick` lacks both shrinking and state machines. [3]

**Suits TuiSnip for:**
- Folder tree invariants, as a state machine over create Folder, move Folder, move Snippet to a Folder or the Root, and delete. The invariants: the tree stays strict (no cycles, one parent or the Root); every Snippet is in at most one Folder; the Root is never a Folder; and the one-Fragment invariant holds.
- Tags: adding and removing Tags keeps the many-to-many relation consistent with a map model.
- Search: an exact title is always found; results are a subset of all Snippets; ranking is stable for identical input.
- Feature tier (`feature` tag): a Snippet saved through the use cases and in-memory SQLite loads back equal.

### Mutation testing

**What it is.** The tool makes small deliberate bugs ("mutants") in your code, for example flipping `<` to `<=` or `&&` to `||`, then runs your tests against each one. If the tests still pass, the mutant "lived", which shows the tests execute that line without actually checking its behaviour. Coverage says a line ran; mutation testing says whether a test would notice it being wrong. The cost is one test run per mutant, which is why it runs nightly and never on PRs. [7][8]

**gremlins specifics** [7][9][10]:
- Uses coverage to mark uncovered mutants `NOT COVERED` and skip them. Statuses are KILLED, LIVED, NOT COVERED, TIMED OUT, and NOT VIABLE.
- By default, runs only the tests of the mutant's own package. `--integration` runs the whole suite and is "generally much slower". It halves workers and test CPU.
- `--tags` passes build tags to `go`. `--diff <ref>` limits mutants to changed code (for a fast local run). `--exclude-files` takes regexes; by default only test files are excluded, so generated sqlc, mockery, and enum files need an explicit rule.
- `--threshold-efficacy` (exit 10) and `--threshold-mcover` (exit 11) both default to 0, which means never fail. `--output` writes machine-readable JSON.
- Config file: `.gremlins.yaml`.
- Scale: the README says it suits "smallish" modules and "doesn't work very well on very big Go modules" because runs "can take hours". TuiSnip is small.
- Known bug, open: integration mode reports killed mutants as LIVED, because it passes no `-count=1` and hits the test cache ([#272](https://github.com/go-gremlins/gremlins/issues/272), 2026-03-13). Avoid `-i` until it is fixed.

**ooze specifics** [8]:
- A library you call from a `//go:build mutation` test file: `go test -tags=mutation`. It adds a `go.mod` dependency.
- Runs the full test command for every mutant. The default is `go test -count=1 ./...`, overridable with `WithTestCommand`, for example to add `-tags=feature`.
- `WithMinimumThreshold` defaults to 1.0, so any live mutant fails the run.
- Since 2026-08 it respects `//go:build` constraints during source discovery.

### Tool maturity (as of 2026-09-30)

| Tool | Kind | Latest release | Last commit | 2026 commits | Stars / open issues | Verdict |
|---|---|---|---|---|---|---|
| [gremlins](https://github.com/go-gremlins/gremlins) | Mutation CLI | v0.6.0, 2025-12-06 | 2026-03-30 | 4 | 434 / 47 | Adopt. Released, pinnable, feature-rich; activity is slowing |
| [ooze](https://github.com/gtramontina/ooze) | Mutation library | v0.3.1 tag, 2023-05-04 (no GitHub releases) | 2026-08-19 | 20 | 290 / 15 | Fallback. Active again, but untagged since 2023, so pseudo-version pins |
| [gomu](https://github.com/sivchari/gomu) | Mutation CLI | v0.2.1, 2026-05-29 | 2026-09-29 | 78 | 45 / 18 | Watch. Very active, skips generated files by default; repo created 2025-07, no build-tag flag documented |
| [avito-tech/go-mutesting](https://github.com/avito-tech/go-mutesting) | Mutation CLI | v2.3.1, 2025-12-26 | 2025-12-26 | 0 | 274 / 25 | Reject. Dormant in 2026 |
| [zimmski/go-mutesting](https://github.com/zimmski/go-mutesting) | Mutation CLI | v1.2, 2021-06-10 | 2021-06-10 | 0 | 677 / 43 | Reject. Abandoned upstream of the avito fork |
| [rapid](https://github.com/flyingmutant/rapid) | Property-based | v1.3.0, 2026-04-30 | 2026-09-04 | active | 889 / 19 | Adopt. Stable v1, MPL-2.0 |
| [gopter](https://github.com/leanovate/gopter) | Property-based | v0.2.8, 2020-06-15 | pushed 2026-04-20 | n/a | n/a | Reject. No release in six years; rapid's README compares against it |
| Native fuzzing | Stdlib | Ships with Go (1.18+) | n/a | n/a | n/a | Adopt |

The release, commit, star, and issue figures come from the GitHub API (`gh api repos/<owner>/<repo>`, `/releases`, `/tags`, `/commits`), queried 2026-09-30.

## Alternatives rejected

- **go-mutesting (avito, zimmski):** no 2026 activity; the zimmski original is abandoned.
- **gomu as primary:** the most active tool, but 14 months old with a small user base and no documented build-tag support. TuiSnip's `feature` tier needs build tags. Revisit in 2027.
- **ooze as primary:** it fits a literal build-tag gate, but it has no release since 2023, so Dependabot cannot track tagged versions. It also puts a test dependency in `go.mod` and reruns the full suite per mutant.
- **gopter and `testing/quick`:** gopter has no release since 2020; `testing/quick` has no shrinking and no state machines. [3]
- **Mutation testing as a PR gate:** the map rules this out. Both gremlins thresholds default to 0, which matches.
- **OSS-Fuzz:** built for public, high-impact projects; the nightly fuzz job covers TuiSnip's scale.

## Open questions for the human

1. Does "mutation testing behind its own tag" require a literal `//go:build mutation` file? If yes, use ooze. If a separate Makefile target and nightly job satisfy it, use gremlins as recommended.
2. Should rapid fail files (`testdata/rapid/`) and fuzz crashers (`testdata/fuzz/`) be committed as regression cases after triage, or stay CI artifacts only? Committing matches Go's intended fuzz workflow. [1]
3. Is the `Property` suffix for rapid tests acceptable? It lets `test-property-deep` select them with `-run`.
4. Neither gremlins v0.6.0 nor ooze has been checked against Go 1.27 here. Verify with one local run at adoption time.

## Sources

1. Go Fuzzing: https://go.dev/doc/security/fuzz/
2. `go help testflag` (Go 1.27.1), also at https://pkg.go.dev/cmd/go#hdr-Testing_flags
3. rapid README: https://github.com/flyingmutant/rapid and https://pkg.go.dev/pgregory.net/rapid
4. rapid state machine example: https://github.com/flyingmutant/rapid/blob/master/example_statemachine_test.go
5. rapid flags and defaults: https://github.com/flyingmutant/rapid/blob/master/engine.go
6. rapid fail-file location: https://github.com/flyingmutant/rapid/blob/master/persist.go
7. gremlins README: https://github.com/go-gremlins/gremlins
8. ooze README: https://github.com/gtramontina/ooze
9. gremlins `unleash` flags: https://github.com/go-gremlins/gremlins/blob/main/docs/docs/usage/commands/unleash/index.md
10. gremlins workers and config: https://github.com/go-gremlins/gremlins/blob/main/docs/docs/usage/commands/unleash/workers.md and https://github.com/go-gremlins/gremlins/blob/main/docs/docs/usage/configuration.md
