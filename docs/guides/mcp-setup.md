# Connect depdeck over MCP

[Model Context Protocol (MCP)](https://modelcontextprotocol.io/docs/getting-started/intro)
lets a coding agent discover and call tools. Depdeck uses **stdio**, short for
standard input and output. Your client launches `depdeck mcp` as a local
child process and exchanges MCP messages with it while the session is open.

Running `depdeck mcp` directly can look idle. It is waiting for protocol
messages, so use your client's MCP tool view to check the connection. Press
Ctrl+C if you started it manually. Depdeck needs no listening port or MCP
server URL. An assessment can look up package names in the public npm registry.
A saved HTML Deck from `depdeck scan` opens and filters offline. Your coding
client's own settings govern where it processes the tool results.

## Install depdeck first

Choose a [published archive](https://github.com/pragatikumari0927/depdeck/releases/tag/v1.0.0)
or use Go to install the published v1.0.0 executable. If you need Go, follow
the [official Go installation guide](https://go.dev/doc/install) and use a
patched toolchain.
For current source builds, see the [README source install](../../README.md#try-the-pixel-archive-from-main).

```sh
go version
go install github.com/pragatikumari0927/depdeck/cmd/depdeck@v1.0.0
```

`PATH` is the list of directories a process searches for executables. If
`depdeck --version` succeeds in your client's environment, continue to
[Codex](#codex-cli) or choose [another client](#other-coding-agents-and-editors).
If you installed with Go and the command cannot be found, use the platform
steps below.

### Native Windows (PowerShell)

```powershell
$depdeckBin = go env GOBIN
if (-not $depdeckBin) {
    $depdeckGoPath = ((go env GOPATH) -split ';', 2)[0]
    $depdeckBin = Join-Path $depdeckGoPath 'bin'
}
$depdeckExe = Join-Path $depdeckBin 'depdeck.exe'
$env:Path = "$depdeckBin;$env:Path"
depdeck --version
```

If you extracted an archive, use its actual directory for `$depdeckBin`.
Native Windows clients need the Windows executable. Set the path explicitly
if an IDE started earlier cannot see the updated `PATH`.

### WSL, Linux, and macOS

Run installation and setup from the environment that will launch the server.

- In **WSL**, install depdeck using Linux Go inside the distro. Use Linux paths
  when the MCP process runs there. For a Windows drive, a project path may
  look like `/mnt/c/path/to/project`.
- On **Linux**, use the Linux executable and POSIX paths such as `/home/…`.
- On **macOS**, use the macOS executable and POSIX paths such as `/Users/…`.
  If you use the published Homebrew install, `command -v depdeck` gives its path.

```sh
depdeck_bin="$(go env GOBIN)"
if [ -z "$depdeck_bin" ]; then
  depdeck_bin="$(go env GOPATH | cut -d: -f1)/bin"
fi
export PATH="$depdeck_bin:$PATH"
depdeck --version
```

If you extracted an archive, use its actual directory for `depdeck_bin`. These
PATH changes apply to the current terminal. Start your client from that
environment or use an absolute executable path. With a remote workspace or
WSL integration, check where the client launches the server before choosing
the executable and its project paths.

## Codex CLI

Run these commands to register depdeck and inspect the entry. The add command
writes to Codex's user configuration, normally `~/.codex/config.toml`
(`%USERPROFILE%\.codex\config.toml` on native Windows).

```sh
codex mcp add depdeck -- depdeck mcp
codex mcp list
codex mcp get depdeck
```

If the client cannot find `depdeck` on PATH, register the absolute executable
path resolved in the install step. If you already added the entry, remove
that entry first.

```powershell
codex mcp add depdeck -- "$depdeckExe" mcp
```

```sh
codex mcp add depdeck -- "$depdeck_bin/depdeck" mcp
```

If you used Homebrew or an extracted archive, substitute its absolute
executable path in the same command. Open a fresh `codex` session in your
project and enter `/mcp`. See [Verify the tools](#verify-the-tools) for the
expected inventory. The `list` and `get` commands inspect registrations.

For a trusted workspace-specific setup, add this table to the project's
`.codex/config.toml` and open Codex in that workspace. Choose a single scope
for depdeck to avoid conflicting entries.

```toml
[mcp_servers.depdeck]
command = "depdeck"
args = ["mcp"]
```

For the Codex IDE extension, open its gear menu, choose **MCP servers**, then
**Add server** and **STDIO**. Set the command to the executable and the argument
to `mcp`. Save and restart the extension. See the
[official Codex MCP guide](https://developers.openai.com/codex/mcp).

Remove a user entry with `codex mcp remove depdeck`, then run `codex mcp list`.
For the project-file option, remove only the `[mcp_servers.depdeck]` table.

## Generic stdio JSON

Use this example for clients that read an `mcpServers` object. Merge the entry
into the client's existing object. Set `command` to the executable and pass
each argument as its own array item. Choose `stdio` or **local command** if
the client asks for a transport.

```json
{
  "mcpServers": {
    "depdeck": {
      "command": "depdeck",
      "args": ["mcp"]
    }
  }
}
```

For an absolute path, replace only the `command` value. JSON needs escaped
Windows backslashes, for example `"C:\\path\\to\\depdeck.exe"`. In a terminal,
quote a path that contains spaces. Keep filesystem paths local to the
environment that runs the server.

## Other coding agents and editors

Use one of the following configurations, then open the client's MCP view or
a new session. Approve the resulting launch if your client prompts you to
trust the workspace or server. Other clients can use the same executable
and argument in their local stdio settings.

### Claude Code

From the project you want to assess, use **local** scope. Claude Code stores
this per-project registration in its personal configuration. A shared project
configuration uses `.mcp.json` and has its own approval prompt.

```sh
claude mcp add --transport stdio --scope local depdeck -- depdeck mcp
claude mcp list
claude mcp get depdeck
```

Open a Claude Code session in that project and enter `/mcp`. Remove the local
entry from the same project with `claude mcp remove --scope local depdeck`.
See [Claude Code MCP configuration](https://code.claude.com/docs/en/mcp).

### Cursor

Add the [generic JSON entry](#generic-stdio-json) to `.cursor/mcp.json` in
the project, or to `~/.cursor/mcp.json` for user scope. Confirm the server
is enabled in Cursor's customization controls and check its available tools
in the Agent panel. See [Cursor MCP configuration](https://cursor.com/docs/mcp).

### Grok Build CLI

The following adds a user-scoped server. For an approved project-local
configuration, add `--scope project` before `--` from the project directory.
Grok Build writes that entry to `.grok/config.toml`.

```sh
grok mcp add depdeck -- depdeck mcp
grok mcp list
grok mcp doctor depdeck
```

Remove the user entry with `grok mcp remove depdeck`. For a project entry,
remove its `[mcp_servers.depdeck]` table from the project file. See the
[official Grok Build MCP guide](https://docs.x.ai/build/features/mcp-servers).

### Hermes Agent

These commands edit and check the current Hermes profile's configuration
(`~/.hermes/config.yaml` for the default profile).

```sh
hermes mcp add depdeck --command depdeck --args mcp
hermes mcp list
hermes mcp test depdeck
```

For manual configuration, merge this entry under `mcp_servers` in that YAML.

```yaml
mcp_servers:
  depdeck:
    command: depdeck
    args: [mcp]
```

Remove it with `hermes mcp remove depdeck`. See the
[Hermes command reference](https://hermes-agent.nousresearch.com/docs/reference/cli-commands)
and [MCP configuration reference](https://hermes-agent.nousresearch.com/docs/reference/mcp-config-reference).

### GitHub Copilot in VS Code and Copilot CLI

In recent VS Code, run **MCP: Add Server** and choose **Command (stdio)**.
Enter `depdeck mcp` as the launch command and `depdeck` as the name. Choose
**Copilot Global** for user scope or workspace-root `.mcp.json` when offered.
The Copilot harness uses `mcpServers`. For its user file, `mcp-config.json`
under the Copilot home directory, use this entry. Set the command to an
absolute path if necessary.

```json
{
  "mcpServers": {
    "depdeck": {
      "type": "stdio",
      "command": "depdeck",
      "args": ["mcp"],
      "tools": ["list_deck", "get_card", "check_deck"]
    }
  }
}
```

The older editor-managed `.vscode/mcp.json` uses a top-level `servers`
object. Use the client's current configuration destination and schema.
See [VS Code's MCP configuration reference](https://code.visualstudio.com/docs/agents/reference/mcp-configuration).

### Gemini CLI

From your project, register a stdio server in the project settings.

```sh
gemini mcp add --scope project depdeck depdeck mcp
gemini mcp list
```

Gemini CLI writes the `mcpServers` entry to `.gemini/settings.json`. In a
session, `/mcp` lists server status and tools. See the
[Gemini CLI MCP guide](https://github.com/google-gemini/gemini-cli/blob/main/docs/tools/mcp-server.md).

### OpenCode

For OpenCode v2, add the local command from your project and check its status.

```sh
opencode mcp add depdeck -- depdeck mcp
opencode mcp list
```

For manual v2 configuration, put `"depdeck": {"type": "local", "command":
["depdeck", "mcp"]}` under `mcp.servers` in `opencode.jsonc`. In v1, server
names go directly under `mcp`. See the
[OpenCode MCP guide](https://opencode.ai/docs/mcp-servers/) and
[v1-to-v2 migration guide](https://opencode.ai/docs/migrate-v1/).

### More desktop clients

| Client | Where to configure a local server |
| --- | --- |
| Claude Desktop | Open **Settings → Developer → Edit Config**. Merge the generic JSON entry in `claude_desktop_config.json` and restart the app. [Local server guide](https://modelcontextprotocol.io/docs/develop/connect-local-servers) |
| Zed | Choose **Settings → AI → MCP Servers → Add Server → Add Local Server**. Use command `depdeck` and argument `mcp`. Zed stores it under `context_servers`. [Zed guide](https://zed.dev/docs/ai/mcp) |
| Cline | In the extension's MCP Servers panel, open **Configure MCP Servers** and merge the generic entry in `cline_mcp_settings.json`. [Cline MCP guide](https://docs.cline.bot/mcp/mcp-overview) |
| Windsurf | In an existing Windsurf installation, use its MCP settings or merge the generic entry in `~/.codeium/windsurf/mcp_config.json`. [Windsurf config location](https://code.visualstudio.com/docs/agents/reference/mcp-configuration) |
| Devin Local | Merge the generic JSON entry in the project's `.devin/mcp_config.json`, or use the user config path for your platform. Run `devin mcp list` in the CLI to inspect the registration. [Devin MCP guide](https://docs.devin.ai/cli/extensibility/mcp/configuration) |

For any client configured by file, remove the `depdeck` entry from the scope
where you added it and reload the client. Keep the other server definitions.

## Copyable setup prompt

Paste this into a coding agent that can run terminal commands.

```text
Help me connect depdeck to my coding client as a local stdio MCP server.
Use depdeck's docs/guides/mcp-setup.md for setup. Identify my client and the
OS and environment where it launches child processes. Check Go, depdeck
--version, and the resolved executable path.

Show me the exact install or registration command and the config file and
scope it will change. Ask me before installing anything or writing a
persistent configuration. After approval, add only the depdeck entry.
Use the absolute executable if PATH is unavailable. Launch it with the
single argument mcp. Keep the existing server entries.

After reloading the client, verify that depdeck exposes list_deck, get_card,
and check_deck. For a smoke test, create a disposable project with an empty
package.json and call list_deck and check_deck using its absolute path.
Avoid live npm for this smoke test. Report the results, any step I must do
in the client, and how to remove the new configuration.
```

## Verify the tools

Look for these three tools under the `depdeck` entry. Some clients show a
server prefix before each name. Use an absolute path to a project directory
or its `package.json` for project operations.

| Tool | Required input | What it returns |
| --- | --- | --- |
| `list_deck` | `path` | Assessment of the root project's direct dependencies. Optionally supply `tag` or `flavor`. |
| `get_card` | npm package `name` | Public npm facts for one package, independent of any project. Optionally supply `flavor`. |
| `check_deck` | `path` | Default policy verdict, flagged and assessed dependencies, and any Fetch errors on their Cards. |

The tag values are `dependencies`, `devDependencies`, and
`optionalDependencies`. Flavor accepts `rule` and `none`. CLI JSON and MCP use
schema `1.1`. The [output contract](../SPEC.md#envelope) explains the data
shapes and the [MCP contract](../SPEC.md#mcp-surface) lists the arguments.
For an empty Check, builds from `main` return `data.dependencies: []`.
The published `v1.0.0` release omits that field for empty Checks.

A project with `{}` in `package.json` can verify the two project tools
without npm access. On a source checkout, run the repository's offline
tests to verify all three tool calls, including `get_card`, without relying
on live npm.

```sh
go test ./cmd/depdeck ./internal/mcp -count=1
```

### Prompts to try

1. “Use depdeck's `list_deck` on the absolute path to this project's root.
   Summarize its direct runtime, development, and optional dependencies.
   Tell me which npm facts are unavailable.”
2. “Use depdeck's `get_card` for `react`. Explain its weekly downloads,
   Rarity, publish date, and license declaration.”
3. “Run depdeck's `check_deck` on the absolute path to this project's root.
   Explain which Cards triggered the default publish-age or license
   policy, and list incomplete npm lookups separately.”

Check flags assessed Cards with Chaos above `0.5` or a missing license
declaration after a successful lookup. Chaos reflects publish age and Rarity
reflects weekly downloads. A Fetch error means facts are missing. Use a
vulnerability scanner for security findings.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| Executable missing or `ENOENT` | Run `depdeck --version` where the client launches subprocesses. Use the absolute Windows or POSIX executable path. A GUI may have inherited an earlier PATH. |
| Server exits or the client waits for a URL | Select local stdio and check the executable and its single `mcp` argument. Inspect the client's MCP diagnostics. `scan` and `check` are one-shot CLI commands. |
| Looks idle in a terminal | The process waits for protocol messages. Start the tool from the MCP client and check the tool inventory there. |
| Server configured but tools missing | Check the scope, enable the server, review any trust or policy prompt, and start a new session or reload the client. |
| Project path or manifest error | Use an absolute directory or `package.json` path accessible from the server's filesystem. Recheck Windows versus WSL paths. |
| Tool responds with Fetch errors | The connection worked. Check access to public npm and explain unavailable facts separately from policy flags. |
| Agent only accepts a remote URL | Use a client that launches stdio processes, or run `depdeck scan --json .` or `depdeck check --json .` in an accessible terminal. |
