# Fuzzy matcher for Search

Research for [#39](https://github.com/DannyFestor/TuiSnip/issues/39). Sources checked 2026-10-01. Local toolchain for the scratch benchmarks and fuzz runs: go1.27.1 darwin/arm64, Apple M1 Pro.

## Question

Which Go fuzzy-matching library should back the `memsearch` adapter? Search matches title, Description, and Tags fuzzily and Fragment content as a literal substring. Each field's score is weighted (title ×4, Tag ×3, Description ×2, content ×1), and a Snippet's score is its best weighted field score ([spec](../spec/v1.md#search)). Compare candidates on maintenance, rune and multi-byte safety, scoring and match positions, performance on a few thousand Snippets, licence, dependencies, and API fit for weighting several fields.

## Recommendation

**Use fzf's matcher, `github.com/junegunn/fzf/src/algo` at v0.74.4, imported only by `memsearch`.**

- `algo.FuzzyMatchV2` scores title, Description, and each Tag.
- `algo.ExactMatchNaive` scores content. It is a literal substring match and returns a score on the same scale, so the content ×1 weight compares fairly with the fuzzy fields.

It is the only candidate whose scores can be multiplied by the spec's weights as they are. A match scores the same in a short field and a long one, and a matching field never scores below zero. sahilm/fuzzy, the obvious alternative, subtracts a point per unmatched byte, so long fields score lower and often below zero, and a ×4 weight on a negative score sinks it further (see Findings).

**go-arch-lint.** Add the vendor and grant it to `memsearch` only:

```yaml
vendors:
  fuzzy-matcher: { in: [ github.com/junegunn/fzf/src/algo, github.com/junegunn/fzf/src/util ] }

deps:
  memsearch:
    mayDependOn: [ memsearch ]
    canUse: [ fuzzy-matcher ]
```

`src/util` is listed because the matcher's input type is `util.Chars` (built with `util.ToChars`) and its optional scratch buffer is `util.Slab`. Granting the two packages by exact path keeps the rest of the fzf module (TUI, shell integration) out of reach.

**Usage rules for `memsearch`.** These come from reading the v0.74.4 source:

1. Call `algo.Init("default")` exactly once, for example in the constructor behind a `sync.Once`. `Init` fills the character-class and bonus tables. The package has no `init()` that does it, and without it no word-boundary or camelCase bonus is ever given. [3]
2. With `caseSensitive=false`, pass the pattern already lowercased. With `normalize=true`, also pass it through `algo.NormalizeRunes`. fzf's own `pattern.go` does both. [3][4]
3. Pass `nil` for the slab, or one slab per query. A `util.Slab` is a reused scratch buffer and must not be shared between concurrent queries, and the index is read under an `RWMutex` read lock ([code standard](../standards/code.md)).
4. Positions are rune indexes into the field, returned in descending order (`[7 4 1 0]` for one match in the scratch run). The highlighter sorts them and maps them to byte offsets or grapheme cells itself.
5. Content smart case follows the spec: `caseSensitive` is true only when the query contains an uppercase letter.

**Fallback.** If importing a package from an application module is not acceptable (Open question 1), use sahilm/fuzzy v0.1.3 and have `memsearch` turn its raw score into a length-independent one before weighting. Writing the matcher in-repo is the third choice.

## Findings

### Candidates at a glance

| Candidate | Latest release | Last commit | Licence | Match positions | Score usable for weighting | New modules in TuiSnip's graph | Verdict |
|---|---|---|---|---|---|---|---|
| [junegunn/fzf](https://github.com/junegunn/fzf) `src/algo` | v0.74.4, 2026-09-12 | `src/algo` 2026-08-10; repo pushed 2026-09-30 | MIT | Yes, rune indexes | Yes | `junegunn/go-shellwords`, `mattn/go-isatty` | Adopt |
| [sahilm/fuzzy](https://github.com/sahilm/fuzzy) | v0.1.3, 2026-06-11 | 2026-06-24 | MIT | Yes, byte offsets | No, length-biased and can be negative | None (already required by `charm.land/bubbles/v2`) | Fallback |
| [lithammer/fuzzysearch](https://github.com/lithammer/fuzzysearch) | v1.1.8, 2023-05-15 | 2026-09-01 (Dependabot only) | MIT | No | Levenshtein distance, lower is better | `golang.org/x/text` | Reject |
| [reinhrst/fzf-lib](https://github.com/reinhrst/fzf-lib) | tag v1.0.0-beta1 (no GitHub releases) | 2025-02-15 | GitHub reports `NOASSERTION` | n/a | n/a | n/a | Reject |
| [ktr0731/go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) | n/a | pushed 2026-09-21 | MIT | n/a | n/a | n/a | Reject: an interactive finder UI, not a matcher library |
| [schollz/closestmatch](https://github.com/schollz/closestmatch) | n/a | pushed 2022-09-13 | MIT | No | n/a | n/a | Reject: bag-of-words closest match, dormant |
| In-repo matcher | n/a | n/a | MIT (ours) | Our choice | Our choice | None | Third choice |

The release, commit, licence, star, and issue figures come from the GitHub API (`gh api repos/<owner>/<repo>`, `/releases`, `/tags`, `/commits`), queried 2026-10-01. The importer counts come from pkg.go.dev: sahilm/fuzzy is imported by 322 packages, `fzf/src/algo` by 36. [1][2]

### junegunn/fzf `src/algo`

- **Importable.** The package sits at `src/algo`, not under `internal/`, so Go allows the import. pkg.go.dev lists it under MIT. [2][3]
- **Not a published library.** The fzf module is an application, at v0.x, and its README does not offer `src/algo` as a public API. No stability promise was found. In practice the `FuzzyMatchV2` signature is identical in 0.30.0 (2022-04-04), v0.56.0 (2024-10-27), v0.66.0, and v0.74.4: `(caseSensitive, normalize, forward bool, input *util.Chars, pattern []rune, withPos bool, slab *util.Slab) (Result, *[]int)`. [3]
- **Maintenance.** 14 releases since 2025-10-01, and 17 commits touched `src/algo/algo.go` in the same period, the latest a rune-mode prefilter and its equivalence tests (2026-08). The release cadence means frequent Dependabot bumps. [1]
- **Algorithm.** V2 is a modified Smith-Waterman that finds the highest-scoring alignment. It gives bonuses at word boundaries, camelCase, and after delimiters, a gap penalty, and a consecutive-match bonus. It runs in O(nm) when a match exists and O(n) when none does. `normalize=true` folds Latin diacritics, so `et` matches `Été`. [3]
- **Scores ignore field length.** In the scratch run, `dock` scored 114 against both `docker` and `docker` followed by 200 more characters. Points come from the matched span, not the whole field, so a weight multiplies like for like.
- **Global state.** `Init(scheme)` writes package-level tables and variables, so it must run once before the first match and never concurrently with matching. [3]
- **Dependencies.** `src/util` imports `github.com/mattn/go-isatty`, `github.com/rivo/uniseg`, `github.com/junegunn/go-shellwords`, and `golang.org/x/sys`. After `go mod tidy` in a scratch module, those four were the only indirect requirements; tcell and the rest of fzf's `go.mod` were pruned. The Charm v2 modules already require `uniseg` and `x/sys`, so the new modules are `go-shellwords` (MIT) and `go-isatty` (MIT). `x/sys` is BSD-3-Clause. [3][5]

### sahilm/fuzzy

- **Maintenance.** v0.1.3 (2026-06-11) added iterator sources (`FindFromIter`). It has one open issue. [1]
- **Already in the graph.** `charm.land/bubbles/v2` v2.2.1 requires `github.com/sahilm/fuzzy v0.1.3` for its `list` component, so adopting it adds no module. [6]
- **Dependencies.** Standard library only. Its `go.mod` lists only a test dependency (`kylelemons/godebug`) and `golang.org/x` tooling. [7]
- **Positions.** `Match.MatchedIndexes` holds byte offsets into the string, which suit slicing a Go string directly. [7]
- **Scoring.** It applies bonuses for the first character, camelCase, separators, and adjacency, then subtracts one point for every byte of the string that was not matched (`penalty := len(match.MatchedIndexes) - len(cleanMatchStr)`). The scratch run scored `dock` at 73 against `docker` and at -128 against `docker` followed by 200 more characters. A Description is longer than a title, so it loses on length alone, and a negative score multiplied by a weight drops further. [7]
- **Multi-byte bug, open.** [Issue #30](https://github.com/sahilm/fuzzy/issues/30) (2026-09-07) reports that both penalties count bytes where they mean characters, so a CJK or accented candidate ranks below an ASCII one of the same length. A fix has been proposed and is not merged. [8]
- **NUL.** Matching stops at the first NUL rune in the candidate, on purpose. [7]

### lithammer/fuzzysearch

- No release since v1.1.8 (2023-05-15). Recent commits are Dependabot bumps. [1]
- `RankMatch` returns a Levenshtein distance, so lower is better and nothing marks which characters matched. Highlighting would need a second pass. [9]
- It pulls in `golang.org/x/text` for Unicode folding and normalisation. [9]

### Writing a matcher in-repo

A subsequence match over runes with boundary bonuses is short to write, has no dependencies, and would be designed for the weight table. The cost is the scoring itself. fzf's comments document years of tuning: boundary bonus against gap penalty, consecutive-chunk bonuses, first-character multiplier. Reproducing that quality is work the issue does not need. This doc has no measured line count or quality comparison for an in-repo matcher.

### Rune and multi-byte safety (fuzz runs)

Scratch fuzz targets ran each matcher on arbitrary `(pattern, text)` pairs and checked that it did not panic and that every returned position was in range.

| Target | Time | Execs | Result |
|---|---|---|---|
| sahilm `fuzzy.Find` | 60 s | 3.76 M | Pass |
| fzf `FuzzyMatchV2` (case-insensitive, normalize) | 60 s | 13.4 M | Pass |
| fzf `ExactMatchNaive` (both case modes) | 60 s | 17.0 M | Pass |

One finding: in sahilm, an invalid UTF-8 byte in the pattern matches an invalid byte in the text, because both decode to `U+FFFD`. The returned offset can then point at a lone continuation byte. It is in range and does not panic, but the highlighter must not assume every offset starts a valid rune when the text is not valid UTF-8.

### Performance

The scratch benchmark used a synthetic corpus: titles of 4 words, Descriptions of 20 words, 3 Tags per Snippet. Each op matched one query against every field of every Snippet, unsorted. That is the work for one keystroke.

| Snippets | Query | sahilm `FindNoSort` | fzf `FuzzyMatchV2` (slab reused) |
|---|---|---|---|
| 1000 | `d` | 1.05 ms | 0.15 ms |
| 1000 | `gitreb` | 1.15 ms | 0.80 ms |
| 5000 | `d` | 5.33 ms | 0.81 ms |
| 5000 | `dcp` | 5.51 ms | 3.99 ms |
| 5000 | `gitreb` | 5.79 ms | 4.08 ms |

Both stay under a 16 ms frame at 5000 Snippets on this machine. The fzf runs held the fields as prebuilt `util.Chars`. `memsearch` should build them once when it loads the index, not on every keystroke. Content substring matching was not benchmarked.

## Alternatives rejected

- **sahilm/fuzzy as primary:** its score falls with field length and goes negative, so it breaks the spec's weighting unless the adapter undoes the length penalty. That correction depends on the library's internal formula, and issue #30 would change that formula. It stays as the fallback because it adds no module.
- **lithammer/fuzzysearch:** no positions for highlighting, a distance metric rather than a match-quality score, and no release since 2023.
- **reinhrst/fzf-lib:** a repackaging of fzf's algorithm. Little use, no commit since 2025-02, and GitHub cannot detect its licence.
- **go-fuzzyfinder, go-fzf, and similar:** interactive finders with their own UI, not matcher libraries.
- **closestmatch:** a different problem (closest whole string) and dormant since 2022.
- **In-repo matcher:** a reasonable third choice, but it means owning scoring quality that fzf already provides.

## Open questions for the human

1. Is it acceptable to import a package from fzf, an application module at v0.x with no stated API promise? The signature has not changed since 2022, but nothing guarantees it. If not, use sahilm/fuzzy with a score correction in `memsearch`.
2. Should Search fold diacritics (`normalize=true`, so `cafe` finds `Café`)? The spec does not say. The recommendation assumes yes for the fuzzy fields and no for content, which the spec defines as a literal substring.
3. Do the spec weights apply to raw fzf scores, or should the adapter scale scores by query length first? Raw scores grow with the number of matched characters, and that number is the same for every field of one query, so raw scores look comparable. Nothing here tested this beyond the examples above.
4. Dependabot will open a PR for most of fzf's frequent releases. Should the module get a lower update frequency or a group in `.github/dependabot.yml`?

## Sources

1. GitHub API, queried 2026-10-01: `gh api repos/{junegunn/fzf,sahilm/fuzzy,lithammer/fuzzysearch,reinhrst/fzf-lib,ktr0731/go-fuzzyfinder,schollz/closestmatch}` with `/releases`, `/tags`, `/commits`
2. pkg.go.dev: https://pkg.go.dev/github.com/junegunn/fzf/src/algo and https://pkg.go.dev/github.com/sahilm/fuzzy
3. fzf `src/algo/algo.go` (algorithm notes, `Init`, `FuzzyMatchV2`, `ExactMatchNaive`): https://github.com/junegunn/fzf/blob/v0.74.4/src/algo/algo.go, plus the same file at tags `0.30.0`, `v0.56.0`, and `v0.66.0` for the signature history
4. fzf `src/pattern.go` (lowercasing and `NormalizeRunes` before matching): https://github.com/junegunn/fzf/blob/v0.74.4/src/pattern.go
5. fzf `go.mod` and `src/util` imports: https://github.com/junegunn/fzf/blob/v0.74.4/go.mod, https://github.com/junegunn/fzf/tree/v0.74.4/src/util
6. bubbles v2.2.1 `go.mod` and `list/list.go`: https://proxy.golang.org/charm.land/bubbles/v2/@v/v2.2.1.mod
7. sahilm/fuzzy v0.1.3 source: https://github.com/sahilm/fuzzy/blob/v0.1.3/fuzzy.go and https://github.com/sahilm/fuzzy/blob/v0.1.3/go.mod
8. sahilm/fuzzy issue #30: https://github.com/sahilm/fuzzy/issues/30
9. lithammer/fuzzysearch v1.1.8 source: https://github.com/lithammer/fuzzysearch/blob/v1.1.8/fuzzy/fuzzy.go
