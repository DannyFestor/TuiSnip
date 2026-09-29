# golangci-lint v2 catalogue for TuiSnip

Research for #6, feeding the linter grilling ticket (#11). Researched 2026-09-30.

## Question

Which golangci-lint v2 linters and formatters suit a strict, SOLID-minded Go codebase (one responsibility per function, no magic numbers, small consumer-side interfaces, generated code excluded), with what settings? How do the v2 config schema, `golangci-lint fmt`, and changed-files-only runs work?

## Version facts

- Current release: **v2.14.0** (changelog date 2026-09-23, GitHub release 2026-09-24). It is also what `mise` installs locally. Source: https://github.com/golangci/golangci-lint/blob/v2.14.0/CHANGELOG.md
- **Deprecated in v2.x, each replaced by a new linter name** (`default: all` still enables the old names and prints a deprecation warning, so the config must disable them explicitly; verified locally):
  - `wsl` became `wsl_v5` (v2.2.0).
  - `gomodguard` became `gomodguard_v2` (v2.12.0, new config shape).
  - `exhaustruct` became `exhaustruct_v5` (v2.13.0, new config keys: `enforce-patterns`, `ignore-patterns`, `optional-patterns`, `explicit-mode`).
- **Changed in v2.0.0** (https://golangci-lint.run/docs/product/migration-guide/):
  - `gosimple` and `stylecheck` merged into `staticcheck`.
  - Renames: `gomnd` to `mnd`, `goerr113` to `err113`, `vetshadow` to `govet` (`shadow` analyzer).
  - Removed: `tenv` (use `usetesting`), `exportloopref`, `execinquery`, `exhaustivestruct`, `golint`, `interfacer`, `maligned`, `scopelint`, `deadcode`, `structcheck`, `varcheck`, `ifshort`, `nosnakecase`.
  - `gofmt`, `gofumpt`, `goimports` and `gci` moved from `linters` to `formatters`, and `golines` was added there.
  - No exclusions apply by default any more.
  - v1 `issues.exclude-*` options became `linters.exclusions.{paths,presets,generated,rules}`.
- The `govet` modernize suite is its own `modernize` linter. v2.13 renamed its `waitgroup` analyzer to `waitgroupgo` and removed `fmtappendf`.

## Recommendation

**Gist:** `default: all` minus a short, reasoned `disable` list. That leaves 92 linters on. The complexity linters (cyclop, gocognit, funlen, nestif) run tighter than their defaults. Interface discipline comes from ireturn, iface and interfacebloat, and error discipline from wrapcheck, err113, errorlint and nilnil. Generated files are excluded with `generated: strict`, and tests get a small set of relaxations. Formatting is gofumpt plus goimports only, run by lefthook on staged files. CI lints the whole repo.

The config below passes `golangci-lint config verify` on v2.14.0.

```yaml
version: "2"

run:
  relative-path-mode: gomod
  modules-download-mode: readonly
  build-tags:
    - feature
    - e2e

linters:
  default: all
  disable:
    # Deprecated in v2.x, superseded by the *_v5 / *_v2 linter of the same name.
    - exhaustruct
    - gomodguard
    - wsl
    # Target libraries TuiSnip does not use.
    - arangolint
    - clickhouselint
    - ginkgolinter
    - promlinter
    - protogetter
    - spancheck
    - zerologlint
    # Duplicates of an enabled linter.
    - gocyclo
    - maintidx
    - gomodguard_v2
    - godoclint
    # Opinion-only, no defect signal (see catalogue).
    - decorder
    - gosmopolitan
    - goheader
    - lll
    - nlreturn
    - noinlineerr
    - prealloc
    - tagalign
    - wsl_v5

  settings:
    cyclop:
      max-complexity: 10
    depguard:
      rules:
        domain:
          list-mode: strict
          files:
            - "**/internal/domain/**"
          allow:
            - $gostd
            - github.com/DannyFestor/TuiSnip/internal/domain
        app:
          list-mode: lax
          files:
            - "**/internal/app/**"
          deny:
            - pkg: github.com/DannyFestor/TuiSnip/internal/adapters
              desc: use cases depend on ports, never on adapters
        everywhere:
          files:
            - $all
          deny:
            - pkg: math/rand$
              desc: use math/rand/v2
            - pkg: github.com/pkg/errors
              desc: use the standard errors package
    dupl:
      threshold: 100
    errcheck:
      check-type-assertions: true
      check-blank: true
    exhaustive:
      check:
        - switch
        - map
    exhaustruct_v5:
      enforce-patterns:
        - ^github\.com/DannyFestor/TuiSnip/.+
      allow-empty-returns: true
    funlen:
      lines: 50
      statements: 30
    gocognit:
      min-complexity: 15
    gocritic:
      enabled-tags:
        - diagnostic
        - style
        - performance
        - experimental
        - opinionated
      disabled-checks:
        - hugeParam
        - whyNoLint
    govet:
      enable-all: true
      disable:
        - fieldalignment
    iface:
      enable:
        - identical
        - unused
        - opaque
        - unexported
        - unusedmethod
    importas:
      no-unaliased: true
    interfacebloat:
      max: 5
    ireturn:
      allow:
        - anon
        - error
        - empty
        - stdlib
        - generic
        - bubbletea.*\.Model$
    nestif:
      min-complexity: 4
    nolintlint:
      require-explanation: true
      require-specific: true
    paralleltest:
      check-cleanup: true
    revive:
      enable-default-rules: true
      rules:
        - name: argument-limit
          arguments: [4]
        - name: function-result-limit
          arguments: [3]
        - name: flag-parameter
        - name: early-return
        - name: deep-exit
        - name: confusing-naming
        - name: import-shadowing
        - name: modifies-value-receiver
        - name: unused-receiver
        - name: bare-return
        - name: use-any
    sloglint:
      no-global: all
      context: scope
      static-msg: true
      kv-only: true
    tagliatelle:
      case:
        rules:
          toml: snake
    unparam:
      check-exported: true
    varnamelen:
      ignore-type-assert-ok: true
      ignore-map-index-ok: true
      ignore-chan-recv-ok: true
      ignore-names:
        - id
        - db
        - tx

  exclusions:
    generated: strict
    warn-unused: true
    presets:
      - comments
    rules:
      - path: _test\.go
        linters:
          - dupl
          - err113
          - exhaustruct_v5
          - funlen
          - goconst
          - varnamelen
          - wrapcheck

formatters:
  enable:
    - gofumpt
    - goimports
  settings:
    gofumpt:
      module-path: github.com/DannyFestor/TuiSnip
      extra:
        group-params: true
        clothe-returns: true
    goimports:
      local-prefixes:
        - github.com/DannyFestor/TuiSnip
  exclusions:
    generated: strict

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

### Exclusion strategy

- **Generated code.** Set `linters.exclusions.generated: strict` and `formatters.exclusions.generated: strict`. Strict mode matches exactly `^// Code generated .* DO NOT EDIT\.$` before the first non-comment text (https://go.dev/s/generatedcode), which is the header sqlc, mockery and stringer emit and the one the map's "never edit" rule keys on. The linter default is already `strict`. The formatter default is `lax`, so set it explicitly. Generated files are still type-checked; only their issues are dropped. `golangci-lint fmt` left a generated file untouched (verified locally). No path excludes are needed.
- **Tests.** Test files stay linted, including `testpackage`, which enforces the settled black-box `_test` packages. The relaxations are the linters whose premise does not hold for table tests: `dupl`, `funlen`, `goconst`, `exhaustruct_v5`, `err113`, `wrapcheck` and `varnamelen`. `mnd` ignores `_test.go` by design. `run.build-tags: [feature, e2e]` is required, otherwise `test/feature` and `test/e2e` are never linted.
- **Presets.** Only `comments` is on. It silences "exported X should have comment" from revive and ST1000/ST1020-22 from staticcheck, which fits the WHY-only comment rule. `common-false-positives` is not used because it blanket-silences gosec G204/G304. Those are exactly the `$EDITOR` launch and config-file read, and each deserves one reasoned `//nolint:gosec // ...` at the call site. `std-error-handling` is not used because it silences unchecked `Close` errors.
- **`nolint` hygiene.** nolintlint requires a named linter and an explanation, and reports unused directives (`allow-unused` defaults to false).

### Formatting and changed-files runs

- `golangci-lint fmt [paths...]` rewrites files using only the `formatters` section. It takes explicit file arguments. `--diff` prints the diff instead and exits 1 when anything would change (both verified locally). `golangci-lint run` also reports unformatted files as `File is not properly formatted (gofumpt)`, so CI needs no separate fmt step. Source: https://golangci-lint.run/docs/welcome/quick-start/#formatting
- **Pre-commit hook (lefthook), formatting only staged files:**

  ```yaml
  pre-commit:
    commands:
      fmt:
        glob: "*.go"
        run: golangci-lint fmt {staged_files}
        stage_fixed: true
  ```

  `{staged_files}` expands to the staged paths after `glob` filtering. `stage_fixed: true` re-runs `git add` on those files and fails the hook if that fails. Source: https://github.com/evilmartians/lefthook/blob/master/docs/configuration/stage_fixed.md
- **Linting changed code only.** Linters need whole packages for type information, so `run` cannot be scoped to files the way `fmt` can. Instead it filters reported issues:
  - `--new-from-rev=REV`: issues introduced after REV.
  - `--new-from-merge-base=main`: issues introduced since the merge base with `main`.
  - `--new-from-patch=PATH`: issues inside a patch file.
  - `--whole-files`: report anywhere in the touched files.
  - `-n/--new`: unstaged plus untracked changes, else `HEAD~`. The CLI help warns CI setups against it.

  TuiSnip is greenfield with zero existing debt, so **CI lints the whole repo** and no new-issue filter is needed. `--new-from-rev=HEAD` is an optional quick local pre-push check.

## Catalogue

FP = false-positive risk (L/M/H) for this codebase. Settings links: `https://golangci-lint.run/docs/linters/configuration/#<name>`.

### Linters named in the ticket

| Linter | What it checks | Key settings | FP | Rec | Reasoning | Source |
|---|---|---|---|---|---|---|
| staticcheck | SA bugs, S simplifications, ST style, QF quick fixes (gosimple and stylecheck merged in v2) | `checks` (default `all` minus ST1000/1003/1016/1020-22), `initialisms` | L | enable | Highest signal-to-noise; defaults already drop doc-comment and naming nits | https://staticcheck.dev/docs/checks/ |
| gosec | Security: G204 subprocess with variable, G304 file path from variable, G301/G302/G306 permissions, and others | `includes`/`excludes`, `severity`, `confidence` | M | enable | Will flag the `$EDITOR` exec and XDG config read; handle with per-site nolint, not a preset | https://github.com/securego/gosec |
| errcheck | Unchecked errors | `check-type-assertions`, `check-blank`, `exclude-functions` | L | enable with settings | Both flags on: `_ = f()` and unchecked `x.(T)` are hidden failures | https://github.com/kisielk/errcheck |
| govet | `go vet` analyzers; `enable-all` adds shadow, nilness, unusedwrite, fieldalignment and others | `enable-all`, `disable`, `settings.shadow.strict` | L (M with shadow) | enable with settings | `enable-all` minus fieldalignment; shadow non-strict catches `err` shadowing | https://pkg.go.dev/cmd/vet |
| govet/fieldalignment | Structs that would be smaller or scan fewer pointer bytes if fields were reordered | (analyzer on/off) | H (readability) | skip | Reorders fields by size rather than meaning; the analyzer's own doc warns compact order can cause false sharing. A snippet TUI is not memory-bound | https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/fieldalignment |
| revive | ~100 configurable rules (golint successor) | `enable-default-rules`, `rules[].arguments`, `enable-all-rules`, `confidence` | M | enable with settings | Defaults plus SRP-relevant extras: `flag-parameter` (a bool param means two behaviours), `argument-limit` 4, `function-result-limit` 3, `early-return`, `deep-exit`, `import-shadowing`, `unused-receiver`. Skip rules that duplicate dedicated linters (`add-constant`/mnd, `cognitive-complexity`/gocognit, `function-length`/funlen, `line-length-limit`) | https://github.com/mgechev/revive/blob/HEAD/RULES_DESCRIPTIONS.md |
| gocritic | 100+ checks tagged diagnostic/style/performance/experimental/opinionated | `enabled-tags`, `disabled-checks`, per-check `settings` | M | enable with settings | All tags on; drop `hugeParam` (Bubble Tea models are value receivers by convention) and `whyNoLint` (nolintlint covers it) | https://go-critic.com/overview.html |
| mnd | Magic numbers in args, case, condition, operation, return, assign | `checks`, `ignored-numbers`, `ignored-functions`, `ignored-files` (tests always ignored) | M | enable | Directly enforces "no magic numbers"; 0 and 1 are always allowed | https://github.com/tommy-muehle/go-mnd |
| gocognit | Cognitive complexity (penalises nesting and breaks in linear flow) | `min-complexity` (default 30; docs recommend 10-20) | L | enable with settings | 15 as a starting point; best single proxy for "one responsibility" | https://github.com/uudashr/gocognit |
| cyclop | Cyclomatic complexity per function, optional package average | `max-complexity` (default 10), `package-average` | L | enable with settings | Keep 10; supersedes gocyclo | https://github.com/bkielbasa/cyclop |
| funlen | Lines and statements per function | `lines` (60), `statements` (40), `ignore-comments` | M | enable with settings | 50/30 proposed; relaxed in tests for table cases | https://github.com/ultraware/funlen |
| nestif | Deeply nested `if` complexity | `min-complexity` (default 5) | L | enable with settings | 4 pushes guard clauses | https://github.com/nakabonne/nestif |
| wrapcheck | Errors returned from other packages, including interface methods, must be wrapped | `ignore-sigs`, `extra-ignore-sigs`, `ignore-package-globs`, `ignore-interface-regexps`, `report-internal-errors` | M | enable | Adds context at every hexagonal boundary; see open question on port errors | https://github.com/tomarrell/wrapcheck |
| errorlint | `==` on errors, type assertions on errors, `%v` instead of `%w` | `errorf`, `errorf-multi`, `asserts`, `comparison` | L | enable | Correct `errors.Is/As` and wrapping | https://codeberg.org/polyfloyd/go-errorlint |
| exhaustive | Missing enum members in `switch` (and optionally map literals) | `check`, `default-signifies-exhaustive`, `default-case-required` | L | enable with settings | `check: [switch, map]`; high value with go-enum/stringer types. `default` does not count as exhaustive | https://github.com/nishanths/exhaustive |
| exhaustruct_v5 | Struct literals missing fields | `enforce-patterns`, `ignore-patterns`, `optional-patterns`, `allow-empty*`, `explicit-mode`, `//exhaustruct:enforce` | H | enable with settings | Scoped to our own module's types so third-party config structs (lipgloss, slog, `exec.Cmd`) are not flagged; `allow-empty-returns` for `return Snippet{}, err`. Off in tests | https://github.com/GaijinEntertainment/go-exhaustruct |
| ireturn | Functions returning interfaces ("accept interfaces, return concrete types") | `allow` or `reject` lists; keywords `anon, error, empty, stdlib, generic` plus regexes | M | enable with settings | Fits small consumer-side interfaces. Fires on methods that implement a third-party interface (verified), so Bubble Tea's `Update() (tea.Model, tea.Cmd)` needs the `bubbletea.*\.Model$` allow entry (regex verified) | https://github.com/butuzov/ireturn |
| interfacebloat | Interfaces with too many methods | `max` (default 10) | L | enable with settings | 5 as the ceiling; ISP | https://github.com/sashamelentyev/interfacebloat |
| gochecknoglobals | Package-level `var`s (exempts `Err*` errors, `_`, `version`, `regexp.MustCompile`, `//go:embed`) | none | M | enable | Forces keymaps and styles to be injected, which suits configurable keybindings; lipgloss style vars will need restructuring | https://github.com/leighmcculloch/gochecknoglobals |
| paralleltest | Missing `t.Parallel()` in tests and subtests | `ignore-missing`, `ignore-missing-subtests`, `check-cleanup` | M | enable with settings | Surfaces hidden shared state; `check-cleanup` catches `defer` with parallel. Tests using `t.Setenv` or shared SQLite need a reasoned nolint | https://github.com/kunwardeep/paralleltest |
| tparallel | Parallel called only at the top level or only in subtests; `defer` instead of `t.Cleanup` | none | L | enable | Complements paralleltest | https://github.com/moricho/tparallel |
| thelper | Test helpers missing `t.Helper()`, wrong param position or name | `test/benchmark/tb/fuzz.{first,name,begin}` | L | enable | Correct failure line numbers | https://github.com/kulti/thelper |
| testifylint | testify misuse (`assert` vs `require`, expected/actual order, `len`, `error-is-as`, ...) | `enable-all`, `disable`, per-checker options | L | enable | Does nothing unless testify is adopted; if it is, the defaults are good | https://github.com/Antonboom/testifylint |
| godot | Comments end with a period | `scope` (declarations), `period`, `capital`, `exclude` | L | enable | Cheap, auto-fixable | https://github.com/tetafro/godot |
| dupl | Duplicate code fragments by token count | `threshold` (default 150) | M | enable with settings | 100 enforces "no duplication"; off in tests (table tests look alike) | https://github.com/mibk/dupl |
| unparam | Params that are unused or always receive the same value | `check-exported` | M | enable with settings | `check-exported: true` is valid for a binary nobody imports; the docs warn of editor false positives when run on a subdirectory | https://github.com/mvdan/unparam |
| nilnil | `return nil, nil` for pointer, interface, map, chan or func plus error | `only-two`, `detect-opposite`, `checked-types` | L | enable | Forces a sentinel error such as `ErrSnippetNotFound` instead of a nil value | https://github.com/Antonboom/nilnil |
| nolintlint | Malformed, unexplained or unused `//nolint` | `require-explanation`, `require-specific`, `allow-unused`, `allow-no-explanation` | L | enable with settings | Every suppression is named and justified | https://golangci-lint.run/docs/linters/configuration/#nolintlint |
| depguard | Import allow/deny lists per file glob | `rules.<name>.{list-mode, files, allow, deny}`; vars `$gostd`, `$all`, `$test` | L | enable with settings | Blunt hexagonal rules (domain imports only stdlib and domain; app never imports adapters) plus bans. **Without rules it allows only `$gostd` everywhere**, so rules are mandatory | https://github.com/OpenPeeDeeP/depguard |
| gomodguard_v2 | Module-level allow/block lists and version constraints in go.mod | `allowed`, `blocked`, `local-replace-directives` | L | skip | Overlaps depguard, go-arch-lint and Dependabot for a single-module app | https://github.com/ryancurrah/gomodguard |
| importas | Enforced import aliases | `alias[]`, `no-unaliased`, `no-extra-aliases` | L | enable with settings | Pin `tea`/`lipgloss` aliases once the Bubble Tea import path is fixed | https://github.com/julz/importas |
| varnamelen | Name length proportional to scope | `max-distance` (5), `min-name-length` (3), `ignore-names`, `ignore-decls`, `check-receiver/return/type-param` | M | enable with settings | Supports self-documenting names; `ctx`, `t`, `b`, `tb` always ignored; add `id`, `db`, `tx`; off in tests | https://github.com/blizzy78/varnamelen |

### Formatters

| Formatter | What it does | Key settings | FP | Rec | Reasoning | Source |
|---|---|---|---|---|---|---|
| gofumpt | Stricter superset of gofmt | `module-path`, `extra.{group-params, clothe-returns, balance-calls}` | L | enable with settings | Settled; `group-params` and `clothe-returns` fit the style (no naked returns) | https://github.com/mvdan/gofumpt |
| goimports | gofmt plus import add/remove and grouping | `local-prefixes` | L | enable with settings | Settled; `local-prefixes` puts TuiSnip imports in their own group | https://pkg.go.dev/golang.org/x/tools/cmd/goimports |
| gci | Fully configurable import sections and order | `sections`, `custom-order`, `no-inline-comments` | L | skip | Overlaps goimports; running both makes them fight over grouping | https://github.com/daixiang0/gci |
| golines | Wraps long lines | `max-len` (100), `tab-len`, `shorten-comments`, `chain-split-dots` | M | skip | Not in the settled set; rewrites aggressively. Revisit with lll if a line limit is wanted | https://github.com/golangci/golines |
| gofmt | Standard formatter (`simplify`, `rewrite-rules`) | `simplify`, `rewrite-rules` | L | skip | gofumpt already applies it | https://pkg.go.dev/cmd/gofmt |

### Other linters worth enabling (on via `default: all`)

| Linter | What it checks | Key settings | FP | Rec | Reasoning | Source |
|---|---|---|---|---|---|---|
| iface | Interface pollution: `identical`, `unused`, `unusedmethod`, `opaque` (returns an interface with one concrete implementation), `unexported` | `enable` (default only `identical`), per-analyzer `exclude` | M | enable with settings | All analyzers; enforces interfaces at the consumer | https://github.com/uudashr/iface |
| inamedparam | Interface methods with unnamed params | `skip-single-param` | L | enable | Self-documenting ports | https://github.com/macabu/inamedparam |
| err113 | Dynamic `errors.New` in returns; `==` on errors | none | M | enable | Pushes sentinel errors plus `%w`; off in tests | https://github.com/Djarvur/go-err113 |
| errname | Sentinels named `ErrX`, types named `XError` | none | L | enable | Naming consistency | https://github.com/Antonboom/errname |
| nilerr, nilnesserr | Returning nil after checking `err != nil`, or returning the wrong error | none | L | enable | Real bugs | https://github.com/gostaticanalysis/nilerr |
| forcetypeassert | Unchecked `x.(T)` | none | L | enable | Pairs with errcheck `check-type-assertions` | https://github.com/gostaticanalysis/forcetypeassert |
| goconst | Repeated string literals | `min-len`, `min-occurrences`, `ignore-string-values` | M | enable | String side of "no magic values"; off in tests | https://github.com/jgautheron/goconst |
| testpackage | Tests not in a `_test` package | `skip-regexp`, `allow-packages` | L | enable | Enforces the settled black-box tests | https://github.com/maratori/testpackage |
| usetesting | `os.MkdirTemp`/`context.Background` in tests instead of `t.TempDir`/`t.Context` | none | L | enable | Replaces the removed `tenv` | https://github.com/ldez/usetesting |
| testableexamples | Examples without `// Output:` | none | L | enable | Examples stay executable | https://github.com/maratori/testableexamples |
| gochecknoinits | `init()` functions | none | L | enable | Explicit wiring in `main` | https://golangci-lint.run/docs/linters/configuration/#gochecknoinits |
| godox | `TODO`/`FIXME` comments | `keywords` | M | enable | Work lives in GitHub Issues | https://github.com/matoous/godox |
| forbidigo | Forbidden identifiers (default `fmt.Print*`, `print`, `println`) | `forbid[]`, `analyze-types` | L | enable | stdout belongs to Bubble Tea; logging goes through slog | https://github.com/ashanbrown/forbidigo |
| sloglint | slog style: global logger, context, static messages, kv vs attr | `no-global`, `context`, `static-msg`, `kv-only`/`attr-only`, `key-naming-case` | L | enable with settings | Injected logger, static messages, one argument style | https://github.com/go-simpler/sloglint |
| tagliatelle | Struct tag key case | `case.rules.<tag>` | L | enable with settings | `toml: snake` for the hand-edited config | https://github.com/ldez/tagliatelle |
| contextcheck, containedctx, noctx | Non-inherited contexts, contexts stored in structs, calls without a context | none | L | enable | Correct cancellation through use cases | https://github.com/kkHAIKE/contextcheck |
| sqlclosecheck, rowserrcheck | Unclosed `sql.Rows`/`Stmt`, unchecked `Rows.Err` | none | L | enable | Matters for hand-written adapters next to sqlc | https://github.com/ryanrolds/sqlclosecheck |
| unqueryvet | `SELECT *` in Go SQL strings | none | L | enable | Cheap; sqlc queries live in `.sql` files | https://golangci-lint.run/docs/linters/configuration/#unqueryvet |
| modernize | Suggests newer language and stdlib forms (`min`, `slices`, range-over-int, ...) | `disable` | L | enable | Go 1.27 codebase, auto-fixable | https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize |
| recvcheck | Mixed pointer and value receivers | none | L | enable | Consistency | https://github.com/raeperd/recvcheck |
| funcorder | Constructor after type, exported methods before unexported | `constructor`, `struct-method`, `function`, `alphabetical` | L | enable | Predictable file layout for humans and agents | https://github.com/manuelarte/funcorder |
| embeddedstructfieldcheck | Embedded fields first, followed by a blank line | none | L | enable | Readability | https://github.com/manuelarte/embeddedstructfieldcheck |
| nonamedreturns, nakedret | Named returns / naked returns in long functions | `report-error-in-defer` | L | enable | Explicit returns | https://github.com/firefart/nonamedreturns |
| whitespace | Leading or trailing blank lines in blocks | `multi-if`, `multi-func` | L | enable | Auto-fixable | https://github.com/ultraware/whitespace |
| Low-noise correctness set: asasalint, asciicheck, bidichk, bodyclose, canonicalheader, copyloopvar, dogsled, dupword, durationcheck, errchkjson, exptostd, fatcontext, gocheckcompilerdirectives, gochecksumtype, gomoddirectives, goprintffuncname, grouper, ineffassign, intrange, iotamixing, loggercheck, makezero, mirror, misspell, musttag, nosprintfhostport, perfsprint, predeclared, reassign, unconvert, unused, usestdlibvars, wastedassign | Mostly bug or clarity checks with near-zero noise; many are no-ops until the relevant API is used | defaults | L | enable | They cost almost nothing and catch real mistakes | https://golangci-lint.run/docs/linters/ |

### Skipped

| Linter | What it checks | FP | Rec | Reasoning | Source |
|---|---|---|---|---|---|
| exhaustruct, gomodguard, wsl | Deprecated names | - | skip | Replaced by the `_v5` / `_v2` versions | https://github.com/golangci/golangci-lint/blob/v2.14.0/CHANGELOG.md |
| gocyclo | Cyclomatic complexity | L | skip | Duplicates cyclop | https://github.com/fzipp/gocyclo |
| maintidx | Maintainability index | M | skip | Opaque composite of the metrics cyclop, gocognit and funlen already enforce | https://github.com/yagipy/maintidx |
| godoclint | Godoc conventions | M | skip | Pending the doc-comment decision; overlaps godot and revive | https://github.com/godoc-lint/godoc-lint |
| lll | Line length | M | skip | Open question; gofumpt does not wrap lines, so lll alone just makes people hand-wrap | https://golangci-lint.run/docs/linters/configuration/#lll |
| wsl_v5, nlreturn | Blank-line placement rules | H (churn) | skip | Pure style, noisy, no defect signal; open question | https://github.com/bombsimon/wsl |
| noinlineerr | Bans `if err := f(); err != nil` | M | skip | Bans an idiomatic scoping form | https://github.com/AlwxSin/noinlineerr |
| prealloc | Slices that could be preallocated | M | skip | Premature optimisation | https://github.com/alexkohler/prealloc |
| decorder, tagalign, goheader | Declaration order, tag alignment, file headers | L | skip | Opinion without payoff (funcorder covers layout; there is no license header policy) | https://golangci-lint.run/docs/linters/ |
| gosmopolitan | i18n anti-patterns | M | skip | i18n is out of scope | https://github.com/xen0n/gosmopolitan |
| arangolint, clickhouselint, ginkgolinter, promlinter, protogetter, spancheck, zerologlint | Library-specific | - | skip | Libraries TuiSnip does not use | https://golangci-lint.run/docs/linters/ |

## Open questions for the human

1. **`default: all` vs an explicit `enable` list.** `all` picks up new linters on every version bump, but mise pins the version, so bumps are deliberate and new linters show up as a failing PR. An explicit list is quieter and easier to miss things with. Which do you want?
2. **Thresholds.** funlen 50 lines / 30 statements (default 60/40), cyclop 10, gocognit 15 (docs recommend 10-20), nestif 4, interfacebloat 5, revive `argument-limit` 4 and `function-result-limit` 3, dupl 100. Tighter or looser?
3. **exhaustruct_v5 scope.** Options: our module's types only (proposed), every type, explicit mode (`//exhaustruct:enforce` opt-in), or skip. It conflicts with Go's useful-zero-value idiom and with sqlc `Params` structs that have optional fields.
4. **Doc comments.** The `comments` preset stops requiring doc comments on exported identifiers, in line with WHY-only comments. Keep that, or require godoc on exported API (drop the preset, enable godoclint)?
5. **wrapcheck on port errors.** Should app code wrap every error returned by a port interface, or add the ports to `ignore-interface-regexps` so the adapter wraps once and the use case passes it through?
6. **gochecknoglobals vs lipgloss styles and default keymaps.** Enforce and inject them (proposed), or exclude the UI adapter package?
7. **paralleltest.** Enforce `t.Parallel` everywhere (proposed), or `ignore-missing-subtests`? Feature tests sharing an in-memory SQLite need per-test databases or nolint.
8. **Line length.** No limit (proposed), lll at 120, or golines as a third formatter?
9. **Whitespace style.** Adopt wsl_v5/nlreturn or not (proposed: not)?
10. **gosec G204/G304.** One reasoned nolint at each call site (proposed) or the `common-false-positives` preset?
11. **Build tags.** `feature` and `e2e` are listed. What is the mutation-test tag name, and should that code be linted?
12. **Bubble Tea import path.** The ireturn allow regex and the importas aliases depend on the major version chosen in the Bubble Tea ticket.
13. **unparam `check-exported: true`.** Accept possible editor false positives in exchange for full coverage?
14. **lefthook partial staging.** How lefthook handles a partially staged file with `stage_fixed` was not researched here; confirm in the lefthook ticket before relying on it.
