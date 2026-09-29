# Agent hooks: Claude Code, OpenCode, Antigravity CLI

Research for #4 (part of #1). Sources fetched 2026-09-30. Versions seen: Claude Code docs mention features up to v2.1.271; OpenCode v1.18.33 (latest release 2026-09-28, source read on the default branch 2026-09-29); Antigravity CLI docs as published at `antigravity.google/docs`.

## Question

What can each agent system's hook mechanism do today, and how should TuiSnip wire six behaviours so that one set of scripts under `scripts/` serves all three?

1. Block Edit/Write on files with a `// Code generated ... DO NOT EDIT.` header, on `migrations/**`, and on other protected globs.
2. After Edit/Write of a `.go` file, run `golangci-lint fmt` and `go vet` (optionally lint) on that file or its package.
3. Block shell commands that edit files (python/sed/perl/awk/node heredocs), and block `git add` and `git commit`.
4. Stop (end-of-turn) hook running `go build ./... && go vet ./...`.
5. How hooks are committed and shared.
6. One set of scripts serving all systems.

## Recommendation

**Policy lives in four agent-agnostic shell scripts with a single contract (exit 0 = pass, exit 2 = violation, reason on stderr). Each agent system gets one thin adapter that turns its stdin JSON into that contract and maps the result back. Where a system has native deny rules that can be committed, use them too.**

### Core scripts (agent-agnostic)

```
scripts/agent/
  guard-path.sh        <path>      exit 2 if path matches scripts/agent/protected-paths, or the file on disk has the Go generated header
  guard-command.sh     (stdin)     exit 2 on git add / git commit, or on shell file edits (sed -i, perl -i, python/node/awk heredocs, cat/tee > file)
  check-go-file.sh     <path>      no-op unless *.go; golangci-lint fmt <file>; go vet ./<pkgdir>; exit 2 with findings
  verify-build.sh                  go build ./... && go vet ./...; exit 2 with output tail on failure
  protected-paths                  one glob per line: migrations/**, plus whatever else the standards ticket adds
  adapters/claude.sh    <event>    reads Claude Code stdin JSON, calls a core script
  adapters/antigravity.sh <event>  reads Antigravity stdin JSON, calls a core script, prints the JSON decision
.opencode/plugins/tuisnip-guards.ts  OpenCode adapter; shells out to the same core scripts
```

- The generated-code test is the Go convention verbatim: `^// Code generated .* DO NOT EDIT\.$`, which "must appear before the first non-comment, non-blank text in the file" (`go help generate`). Check the file on disk, so Edit and a Write that overwrites both get blocked.
- The command goes to `guard-command.sh` on stdin, never argv, so heredocs and quotes arrive unmangled.
- `check-go-file.sh` formats only the one file (`golangci-lint fmt file1.go` is supported) and vets its package directory. Lint on edit is optional: the golangci-lint docs say files passed to `run` "must come from the same package", so linting the package dir is the natural unit.
- jq is needed by the two shell adapters. It is not in `mise.toml` today; add it there.

### Per-system wiring

**Claude Code**: `.claude/settings.json`, committed.

```json
{
  "permissions": {
    "deny": ["Edit(migrations/**)", "Bash(git add *)", "Bash(git commit *)"]
  },
  "hooks": {
    "PreToolUse": [
      { "matcher": "Edit|Write|NotebookEdit", "hooks": [{ "type": "command", "command": "${CLAUDE_PROJECT_DIR}/scripts/agent/adapters/claude.sh", "args": ["pre-edit"] }] },
      { "matcher": "Bash", "hooks": [{ "type": "command", "command": "${CLAUDE_PROJECT_DIR}/scripts/agent/adapters/claude.sh", "args": ["pre-bash"] }] }
    ],
    "PostToolUse": [
      { "matcher": "Edit|Write", "hooks": [{ "type": "command", "command": "${CLAUDE_PROJECT_DIR}/scripts/agent/adapters/claude.sh", "args": ["post-edit"], "timeout": 120 }] }
    ],
    "Stop": [
      { "hooks": [{ "type": "command", "command": "${CLAUDE_PROJECT_DIR}/scripts/agent/adapters/claude.sh", "args": ["stop"], "timeout": 300 }] }
    ]
  }
}
```

The adapter extracts `.tool_input.file_path` or `.tool_input.command` and passes the core exit code straight through, because Claude Code's native contract already is "exit 2, stderr to Claude". On `stop` it exits 0 when `.stop_hook_active` is true. The deny rules are the hard layer for the fixed globs and git. The hook adds the checks that need content (generated header) or regexes (shell edits).

**OpenCode**: `.opencode/plugins/tuisnip-guards.ts` plus `opencode.json` permissions, both committed.

- `tool.execute.before`: for `edit` and `write` use `output.args.filePath`. For `apply_patch`, parse the paths out of `output.args.patchText` (`*** Add File:`, `*** Update File:`, `*** Move to:`, `*** Delete File:`). For `bash` use `output.args.command`. Run the core script with Bun `$` without throwing on non-zero exit, and `throw new Error(stderr)` on exit 2.
- `tool.execute.after`: for `edit`, `write` and `apply_patch` on `.go` files, run `check-go-file.sh` and append the findings to `output.output` so the model sees them.
- `event` with `session.idle`: run `verify-build.sh`. On failure, send the output back with `client.session.prompt(...)`, with the plugin keeping its own per-session retry cap.
- `opencode.json`: `permission.edit: {"migrations/*": "deny"}` and `permission.bash: {"git add *": "deny", "git commit *": "deny"}`. Leave `formatter` off (it is off by default) so formatting happens only through `check-go-file.sh`.

**Antigravity CLI**: `.agents/hooks.json`, committed.

```json
{
  "tuisnip-guards": {
    "PreToolUse":  [{ "matcher": "write_to_file|replace_file_content|multi_replace_file_content|run_command",
                      "hooks": [{ "command": "scripts/agent/adapters/antigravity.sh pre", "timeout": 10 }] }],
    "PostToolUse": [{ "matcher": "write_to_file|replace_file_content|multi_replace_file_content",
                      "hooks": [{ "command": "scripts/agent/adapters/antigravity.sh post", "timeout": 120 }] }],
    "PreInvocation": [{ "command": "scripts/agent/adapters/antigravity.sh pre-invocation" }],
    "Stop": [{ "command": "scripts/agent/adapters/antigravity.sh stop", "timeout": 300 }]
  }
}
```

The adapter reads `.toolCall.args.TargetFile` or `.toolCall.args.CommandLine`. On `pre` it prints `{"decision":"deny","reason":...}`. PostToolUse output is fixed at `{}`, so edit-time findings have no direct route back to the model. `post` therefore writes them to a per-conversation state file, and `pre-invocation` injects them as `{"injectSteps":[{"ephemeralMessage":...}]}`. On `stop` it prints `{"decision":"continue","reason":...}` on failure, capped by `executionNum`. Permission deny rules cannot be committed for this system (they live only in `~/.gemini/antigravity-cli/settings.json`), so the hook is the only shared layer.

## Capabilities matrix

| Behaviour | Claude Code | OpenCode | Antigravity CLI |
| :- | :- | :- | :- |
| 1. Block edits: generated header, `migrations/**`, globs | **Yes.** PreToolUse on `Edit\|Write`, `tool_input.file_path`, exit 2 blocks, and "even a JSON `permissionDecision` of `"allow"` can't override it" [C1][C2]. Native `Edit(migrations/**)` deny rules also cover Bash `sed`/`tee`/redirect targets [C5]. | **Yes.** `tool.execute.before`, `output.args.filePath`, throw to block [O1]. `apply_patch` carries paths inside `patchText` instead [O3]. `permission.edit` glob deny [O4]. | **Yes.** PreToolUse on `write_to_file\|replace_file_content\|multi_replace_file_content`, `toolCall.args.TargetFile`, output `"decision":"deny"` [A1]. |
| 2. Post-edit `golangci-lint fmt` + `go vet` with feedback | **Yes.** PostToolUse can't undo the edit, but exit 2 "shows stderr to Claude" [C2]. `decision:"block"` / `additionalContext` add text next to the result [C3]. | **Yes.** `tool.execute.after` gets a mutable `output.output` [O2]. The built-in `gofmt` formatter is off by default; custom formatters take `$FILE` [O5]. | **Partial.** PostToolUse output "Returns an empty JSON object `{}`", so there is no feedback field. PreInvocation `injectSteps` / `ephemeralMessage` can carry findings into the next model call [A1]. |
| 3. Block shell file edits, `git add` / `git commit` | **Yes.** PreToolUse `Bash`, `tool_input.command` [C1]. Deny rules `Bash(git commit *)` match any subcommand in `&&` chains and `$()` [C5], but "isn't a security boundary" (e.g. `git -C . push` escapes) [C5]. Deny rules don't catch "a Python or Node script that opens files itself" [C5], hence the hook. | **Yes.** `tool.execute.before` with `input.tool === "bash"`, `output.args.command` [O1]. `permission.bash` patterns such as `"git commit *": "deny"` [O4]. | **Yes.** PreToolUse on `run_command`, `toolCall.args.CommandLine`, `"decision":"deny"` [A1]. `command(...)` deny rules exist, but only in the global settings file [A2]. |
| 4. End-of-turn `go build && go vet` | **Yes.** Stop hook; exit 2 or `decision:"block"` prevents stopping. Loop guard `stop_hook_active`, plus a built-in 8-consecutive-continuation cap [C4]. | **Workaround.** `session.idle` is a notification event (`event?: (...) => Promise<void>`) and cannot veto the stop [O1][O2]. The plugin can re-prompt through `client.session.prompt` [O6]. The loop cap is ours to build. | **Yes.** Stop output `"decision":"continue"` re-enters the loop and `reason` is injected [A1]. There is no documented loop cap; `executionNum` is available to build one [A1]. |
| 5. Commit and share | `.claude/settings.json` "can be committed to the repo"; `.claude/settings.local.json` is personal [C6]. Hooks are held back until the workspace-trust dialog is accepted, and run untrusted under `-p` [C7]. | `.opencode/plugins/` is "automatically loaded at startup"; `opencode.json` is the project config; npm deps go in `.opencode/package.json` and are installed with `bun install` [O1][O7]. | `.agents/hooks.json` "at your project root" [A1]. Nothing documented about trust or approval for workspace hooks. |
| 6. One script set for all | Command hooks run any executable; `${CLAUDE_PROJECT_DIR}` gives the repo root; exec form via `args` [C8]. | The plugin is TS, but gets Bun `$` to shell out to `scripts/` [O1]. | Handlers are shell `command` strings; `type` "Currently only `"command"` is supported" [A1]. |

## Findings

**Claude Code** [C1–C8]
- Matchers: exact names or a `|` list; any other character switches to an unanchored JS regex, so `Edit.*` also matches `NotebookEdit`.
- The `if` field takes one permission-rule pattern (`"Edit(*.ts)"`, `"Bash(git *)"`) and works only on tool events; on other events a hook with `if` never runs. `Edit(src/**)` anchors at the cwd since v2.1.214. The docs call `if` best-effort and say to "use the permission system rather than a hook to enforce a hard allow or deny". Don't use `if` as the gate. Match broadly and decide in the script.
- Exit 1 is non-blocking. A missing or non-executable script exits 127 and "leaves the gate silently disabled". Policy hooks must use exit 2.
- A timed-out PreToolUse command hook does not block. Keep guards fast.
- Edit input: `file_path`, `old_string`, `new_string`, `replace_all`. Write input: `file_path`, `content`. Both paths are absolute.
- Permission path rules apply to `Edit(...)`/`Read(...)` only. A `Write(...)` path rule "is accepted but never consulted" (v2.1.210+), so write `Edit(migrations/**)`.
- `${CLAUDE_PROJECT_DIR}` stays on the main checkout when Claude enters a worktree, while the `cwd` input follows the worktree. Scripts should `cd "$(jq -r .cwd)"` before running `go` so they check the worktree that was actually edited.
- On Bash calls, PostToolUse may get `tool_response.bashEditDiff` (v2.1.269+, beta). The docs say not to use it to enforce a policy.

**OpenCode** [O1–O7]
- Hook signatures (from `packages/plugin/src/index.ts`): `"tool.execute.before"(input: {tool, sessionID, callID}, output: {args})` and `"tool.execute.after"(input: {tool, sessionID, callID, args}, output: {title, output, metadata})`. In `session/tools.ts`, `before` runs ahead of the tool's `execute`, so a throw stops the call.
- Tool arg names from source: `edit` takes `filePath`, `oldString`, `newString`; `write` takes `filePath`, `content`.
- Some models edit through `apply_patch` rather than `edit`. A guard that checks only `edit`/`write` misses those edits [O3].
- OpenCode reads `AGENTS.md`, falling back to `CLAUDE.md`. Its "Claude Code compatibility" covers rules and skills only, not `.claude/settings.json` hooks [O8]. OpenCode needs its own adapter.

**Antigravity CLI** [A1–A3]
- Hooks are real and documented. There are five events (`PreToolUse`, `PostToolUse`, `PreInvocation`, `PostInvocation`, `Stop`), with the same config shape in the CLI, the IDE and Antigravity 2.0.
- Config keys are named hook groups with an optional `enabled`. The default timeout is 30 s.
- Input uses camelCase fields with a `toolCall: {name, args}` object; tool args are PascalCase (`TargetFile`, `CommandLine`, `Cwd`).
- **Undocumented**: exit-code semantics, the hook's working directory, environment variables, and what happens when PreToolUse prints nothing. `decision` is marked **Required**, and `"allow"` "Automatically allows the tool execution", so an adapter that answers `allow` for clean calls would silently bypass the user's permission prompts. See open questions.
- `workspacePaths` in the input gives the repo root without relying on cwd.
- It reads `AGENTS.md` and `GEMINI.md` at workspace and directory scope, plus `.agents/rules/*.md` [A3]. The settled "AGENTS.md only" rule holds.

**Community practice**
- `go-gui-org/go-gui` is the closest match to our target. [`.claude/settings.json`](https://github.com/go-gui-org/go-gui/blob/HEAD/.claude/settings.json) wires one script as PreToolUse `pre` and PostToolUse `post` (timeout 180), plus a Stop gate (timeout 300).
  - [`.claude/hooks/go-edit-check.sh`](https://github.com/go-gui-org/go-gui/blob/HEAD/.claude/hooks/go-edit-check.sh) blocks direct `go.sum` edits. After an edit it runs `golangci-lint run --fix` and `go test -short` on the edited package only ("recursing ... makes each edit wait on the whole tree"), and uses exit 2 because "Exit 1 is a non-blocking error the model never sees".
  - [`.claude/hooks/stop-gate.sh`](https://github.com/go-gui-org/go-gui/blob/HEAD/.claude/hooks/stop-gate.sh) limits itself to packages with changed `.go` files, and skips the run when the Go diff fingerprint matches the last pass. It caps blocks at 3 per session and notes that it exists to catch "edits made through Bash (sed, gofmt -w, go generate)".
- `openshift/gcp-project-operator` is a cautionary example. [`.claude/hooks/pre-edit.sh`](https://github.com/openshift/gcp-project-operator/blob/HEAD/.claude/hooks/pre-edit.sh) blocks `zz_generated.*` and mock files, but it reads the path from `$1` instead of stdin, exits 1 (non-blocking), and is not registered in [`.claude/settings.json`](https://github.com/openshift/gcp-project-operator/blob/HEAD/.claude/settings.json) (only SessionStart and Stop are). The guard never runs. The same settings file shows the permission-layer pattern: `ask` on `Bash(git commit *)`, and `deny` on `--no-verify` and force-push variants.
- Takeaways for our design: take the path from stdin JSON, exit 2, scope work to the edited package, and fingerprint the Stop gate so no-op turns are free.

## Alternatives rejected

- **Per-system scripts in `.claude/hooks/`, `.opencode/`, `.agents/`.** This triplicates the policy and contradicts the settled "thin wrappers over `scripts/`".
- **A Go program (`go run ./scripts/agenthook`) as the single adapter.** One language and testable, but each hook pays for `go run` and breaks whenever the module fails to resolve, which is exactly when the Stop gate matters. Worth revisiting if the shell adapters grow.
- **Relying on Claude Code's `if` field or on deny rules alone.** Both are best-effort by the docs' own statement. Deny rules can't see generated headers or catch interpreters writing files.
- **OpenCode's built-in `formatter`.** It runs gofmt, not `golangci-lint fmt` (gofumpt and goimports), which contradicts the settled formatter. It also can't report vet findings.
- **Running `golangci-lint run` on every edit.** Too slow on a very strict ruleset. Lint belongs to the Stop gate or lefthook/CI; edit-time is fmt + vet only.
- **Blocking every Stop until the build passes, with no cap.** Claude Code caps at 8 on its own; OpenCode and Antigravity don't. Every adapter keeps an explicit cap.

## Open questions for the human

1. **Antigravity no-op answer.** PreToolUse `decision` is required and `"allow"` auto-approves. Should clean calls answer `"ask"` (which respects "Always Allow" but may prompt more), or should we first test whether empty output falls through to the normal permission flow? This needs a hands-on `agy` test; the docs don't say.
2. **Antigravity working directory.** The docs' examples use `./scripts/lint.sh`, which implies the project root, but this is undocumented. Is a quick `agy` check acceptable, or should the command strings resolve the root some other way? The adapter itself can use `workspacePaths[0]`.
3. **Writes that create a new file carrying the generated header** (hand-writing a would-be generated file). Should these be blocked too (check `content`/`CodeContent`), or only edits to existing generated files?
4. **Scope of `protected-paths` beyond `migrations/**`.** Candidates: `go.sum`, `.github/workflows/**`, `mise.toml`, `.claude/settings.json` and `.agents/hooks.json` themselves. Which of these?
5. **Stop gate: `go vet` only, or also `golangci-lint run --new-from-rev=HEAD`?** Is lint at end of turn wanted, given the strict ruleset?
6. **Is jq as a pinned mise tool acceptable** as a hard dependency of the agent hooks?

## Sources

- [C1] Claude Code hooks reference, PreToolUse input (Bash/Write/Edit) and decision control: https://code.claude.com/docs/en/hooks#pretooluse
- [C2] Exit code output and "Exit code 2 behavior per event": https://code.claude.com/docs/en/hooks#exit-code-output
- [C3] PostToolUse decision control: https://code.claude.com/docs/en/hooks#posttooluse-decision-control
- [C4] Stop input and decision control: https://code.claude.com/docs/en/hooks#stop ; loop cap guidance: https://code.claude.com/docs/en/hooks-guide (section "Stop hook hits the block cap")
- [C5] Permissions: compound commands, Bash rule limits, file path rules: https://code.claude.com/docs/en/permissions
- [C6] Hook locations: https://code.claude.com/docs/en/hooks#hook-locations
- [C7] Workspace trust: https://code.claude.com/docs/en/hooks#workspace-trust
- [C8] Command hook fields, exec vs shell form, `${CLAUDE_PROJECT_DIR}`, worktrees: https://code.claude.com/docs/en/hooks#command-hook-fields ; https://code.claude.com/docs/en/hooks#reference-scripts-by-path ; protected-files example: https://code.claude.com/docs/en/hooks-guide (section "Block edits to protected files")
- [O1] OpenCode plugins: https://opencode.ai/docs/plugins/
- [O2] Plugin hook types: https://github.com/anomalyco/opencode/blob/HEAD/packages/plugin/src/index.ts ; hook invocation: https://github.com/anomalyco/opencode/blob/HEAD/packages/opencode/src/session/tools.ts ; tool params: `packages/opencode/src/tool/edit.ts`, `write.ts`
- [O3] OpenCode tools (`apply_patch` and `patchText`): https://opencode.ai/docs/tools/
- [O4] OpenCode permissions: https://opencode.ai/docs/permissions/
- [O5] OpenCode formatters: https://opencode.ai/docs/formatters/
- [O6] OpenCode SDK `session.prompt`: https://opencode.ai/docs/sdk/
- [O7] OpenCode config precedence and `.opencode` dirs: https://opencode.ai/docs/config/
- [O8] OpenCode rules and Claude Code compatibility: https://opencode.ai/docs/rules/
- [A1] Antigravity hooks: https://antigravity.google/docs/hooks (markdown at `/docs/hooks.md`)
- [A2] Antigravity permissions (CLI rules in `~/.gemini/antigravity-cli/settings.json`): https://antigravity.google/docs/permissions
- [A3] Antigravity rules: https://antigravity.google/docs/rules
- Go generated-code header: `go help generate` (Go toolchain)
- golangci-lint `fmt` / `run` path arguments: https://golangci-lint.run/docs/welcome/quick-start/
