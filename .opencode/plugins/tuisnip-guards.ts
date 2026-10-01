// OpenCode adapter for the agent hooks in scripts/agent/. It only routes tool calls to the
// core scripts; the policy lives there. Ask rules come from the generated opencode.json.

const AGENT_DIR = "scripts/agent"
const EXIT_DENY = 2
const FILE_TOOLS = new Set(["edit", "write", "apply_patch"])
const GO_FILE = /\.go$/

type ToolArgs = {
  filePath?: string
  content?: string
  newString?: string
  patchText?: string
  command?: string
}

type ScriptResult = { status: number; message: string; stdout: string }

export const TuisnipGuards = async ({ client, $, directory, worktree }) => {
  const script = (name: string) => `${worktree}/${AGENT_DIR}/${name}`

  const runScript = async (name: string, args: string[], stdin = ""): Promise<ScriptResult> => {
    const result = await $`${script(name)} ${args} < ${new Response(stdin)}`.cwd(directory).quiet().nothrow()
    return { status: result.exitCode, message: result.stderr.toString().trim(), stdout: result.stdout.toString() }
  }

  const editedPaths = async (tool: string, args: ToolArgs): Promise<string[]> => {
    if (tool !== "apply_patch") {
      return [args.filePath ?? ""]
    }
    const result = await runScript("patch-paths.sh", [], args.patchText)
    return result.stdout.split("\n").filter(Boolean)
  }

  const guardEdit = async (tool: string, args: ToolArgs) => {
    const content = args.content ?? args.newString ?? args.patchText ?? ""
    for (const path of await editedPaths(tool, args)) {
      const result = await runScript("guard-path.sh", [path], content)
      if (result.status === EXIT_DENY) {
        throw new Error(result.message)
      }
    }
  }

  const guardCommand = async (args: ToolArgs) => {
    const result = await runScript("guard-command.sh", [], args.command)
    if (result.status === EXIT_DENY) {
      throw new Error(result.message)
    }
  }

  const goFileFindings = async (tool: string, args: ToolArgs): Promise<string[]> => {
    const findings: string[] = []
    for (const path of (await editedPaths(tool, args)).filter((path) => GO_FILE.test(path))) {
      const result = await runScript("check-go-file.sh", [path])
      if (result.status === EXIT_DENY) {
        findings.push(result.message)
      }
    }
    return findings
  }

  // session.idle cannot veto the stop, so a failing gate sends the agent back with a prompt.
  const verifyBuild = async (sessionID: string) => {
    const result = await runScript("verify-build.sh", [sessionID])
    if (result.status === EXIT_DENY) {
      await client.session.prompt({ path: { id: sessionID }, body: { parts: [{ type: "text", text: result.message }] } })
    }
  }

  return {
    "tool.execute.before": async (input, output) => {
      if (FILE_TOOLS.has(input.tool)) {
        await guardEdit(input.tool, output.args)
      }
      if (input.tool === "bash") {
        await guardCommand(output.args)
      }
    },
    "tool.execute.after": async (input, output) => {
      if (!FILE_TOOLS.has(input.tool)) {
        return
      }
      for (const finding of await goFileFindings(input.tool, input.args)) {
        output.output += `\n\n${finding}`
      }
    },
    event: async ({ event }) => {
      if (event.type === "session.idle") {
        await verifyBuild(event.properties.sessionID)
      }
    },
  }
}
