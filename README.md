# TuiSnip

TuiSnip is a terminal application for code snippets. It runs on macOS and Linux. It gives a SnippetsLab-like workflow in the terminal.

With TuiSnip, you save a Snippet, file it in a Folder, and give it Tags. You find it again with Search, and you copy it to the clipboard. The keyboard controls all operations.

## Status

TuiSnip is before v0.1. The binary does not do anything yet.

The v0.1 milestone is a walking skeleton. In it, you create a Snippet at the Root, find it with Search, and copy it, all through the TUI.

- The v1 specification is in [`docs/spec/v1.md`](docs/spec/v1.md).
- The order of later work is in [`ROADMAP.md`](ROADMAP.md).
- The domain terms are in [`CONTEXT.md`](CONTEXT.md). Examples are Snippet, Folder, Root, and Tag.

## Requirements

- macOS or Linux.
- [git](https://git-scm.com/).
- [mise](https://mise.jdx.dev/). mise installs Go and all other tools at the versions that CI uses.
- GNU Make. macOS and most Linux distributions include it.

The build does not use CGO. You do not need a C compiler.

## Build from source

1. Clone the repository:

   ```sh
   git clone https://github.com/DannyFestor/TuiSnip.git
   cd TuiSnip
   ```

2. Install the tools:

   ```sh
   mise install
   ```

3. Build the binary:

   ```sh
   make build
   ```

The binary is `bin/tuisnip`.

## Set up for development

Do the steps in [Build from source](#build-from-source) first. Then install the git hooks:

```sh
lefthook install
```

Do this one time for each clone. The hooks do these checks:

| Hook | Checks |
|---|---|
| pre-commit | Stops a commit on `main`. Formats the staged Go files. Runs shellcheck on the staged shell scripts. |
| commit-msg | Makes sure that the message obeys [Conventional Commits](https://www.conventionalcommits.org/). |
| pre-push | Stops a push to `main`. Lints the changed packages. Runs the unit tests. |

To run the feature and e2e tests before a push too, set `TUISNIP_SLOW_TESTS=1`.

The agent hooks for Claude Code, OpenCode, and Antigravity need jq. `mise install` installs it. Run `mise install` before you let a coding agent work in the repository.

### Make targets

`make help` shows all targets. These are the most frequent:

| Target | Does |
|---|---|
| `make build` | Builds `bin/tuisnip`. |
| `make run` | Starts TuiSnip on your own config and data. `ARGS=--paths` passes flags. |
| `make test` | Runs the unit, feature, and e2e tests. |
| `make test-unit` | Runs only the unit tests. |
| `make test-platform` | Runs the unit tests plus those that use the real clipboard. They overwrite it and then restore it. |
| `make snapshot` | Builds the release archives into `dist/` without publishing them. |
| `make lint` | Runs golangci-lint. |
| `make arch-lint` | Checks the layer rules with go-arch-lint. |
| `make fmt` | Formats the Go code. |
| `make generate` | Runs all code generators. |

Do not edit generated files. Change the source of the generator, then run `make generate`. CI fails when a generated file is out of date.

## Contribute

- GitHub squash-merges each pull request. The pull request title becomes the commit message, so it must obey Conventional Commits.
- CI must pass before a merge.
- The coding standards are in [`docs/standards/`](docs/standards/README.md). Start from its index before you change code.
- The tools, and the reasons for each tool, are in [`docs/toolchain.md`](docs/toolchain.md).
- Coding agents start from [`AGENTS.md`](AGENTS.md).

## License

MIT. Refer to [`LICENSE`](LICENSE).
