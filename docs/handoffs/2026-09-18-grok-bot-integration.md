# Handoff — 2026-09-18 Grok Bot integration

Status snapshot for the next Grok Build session. Constitution: `AGENTS.md`. Map: `PHASES.md`, `SPEC.md`. History: `DECISIONS.md`, `docs/adr/`. This file is status only.

## Session summary

This session wired Cursor Grok Bot into Grok Build using [grok-bot-cli](https://github.com/ScriptedAlchemy/grok-bot-cli) `0.9.0`. `gbot-install` does not target Grok Build (hosts are claude, codex, cursor), so MCP was registered with `grok mcp add` against the real stdio entry `dist/mcp/mcp-grok-bot-b8c2461e.mjs`. `AGENT_BUNDLE_PLUGIN_ROOT` was omitted after `grok mcp doctor grok-bot` reported a healthy handshake and 12 tools without it. The `talk-to-grok-bot` skill was copied to the user Grok skills root. `AGENTS.md` already carries the README’s General-bot send/thread note (working tree, uncommitted). Codex auto-reply was left out: `gbot codex` needs a Unix control socket and is unsupported on native Windows.

Ticket `04-registry` was not started. Live `gbot` traffic is blocked: `listAgents` and `send` return `417 Expectation Failed` from the Cursor Grok Bot gateway. `gbot doctor` had already shown a usable app session and present gateway auth; the 417 is not a Grok Build config defect. Do not debug Grok Build for it. Wait until the Grok Bot computer finishes reconnecting or updating, then retry `gbot bots list`.

## Completed this session

- `grok-bot-cli` upgraded 0.8.0 → 0.9.0 (`npm install --global grok-bot-cli@0.9.0`). CLI: `C:\Users\Aryan Singh\AppData\Roaming\npm\gbot.ps1`. Package: `C:\Users\Aryan Singh\AppData\Roaming\npm\node_modules\grok-bot-cli\`.
- Resolved MCP entry (do not hardcode on upgrade): `C:\Users\Aryan Singh\AppData\Roaming\npm\node_modules\grok-bot-cli\dist\mcp\mcp-grok-bot-b8c2461e.mjs` (sibling `mcp-grok-bot-b8c2461e-flight.mjs` is not the server).
- Backup before MCP write: `C:\Users\Aryan Singh\.grok\config.toml.bak-gbot-20260918-100649`.
- User config `C:\Users\Aryan Singh\.grok\config.toml` gained `[mcp_servers.grok-bot]`: `command = "node"`, `args` = that `.mjs`, `enabled = true`. No `cwd`, no `AGENT_BUNDLE_PLUGIN_ROOT`.
- `grok mcp doctor grok-bot`: command found, server started, handshake OK (protocol 2025-11-25), 12 tools discovered.
- Skill: `C:\Users\Aryan Singh\.grok\skills\talk-to-grok-bot\SKILL.md` (copied from the npm package `dist/skills/talk-to-grok-bot/`).
- Repo pointer: `AGENTS.md` section **Grok Bot** (modified, not committed). Send to `General` with `gbot_send`; read with `gbot_thread`; do not wait on the turn.

## Blocked

`gbot listAgents failed: 417 Expectation Failed` (same on `gbot bots list`, `gbot groups list`, `gbot skills list`, `gbot send General`).

- Cursor Grok Bot gateway issue, not Grok Build config. Doctor already succeeded; MCP handshake already succeeded.
- Not fixable from this side.
- Workaround: wait for the Grok Bot computer to finish reconnecting or updating, then retry `gbot bots list`. If still 417, wait and retry. Do not chase Grok Build, Headroom, or `config.toml`.

## Current project state

Tickets done (feat commits; `PHASES.md` says 7 of 13 merged):

- `420b921` feat(types): Dependency, Rarity, Tag, FetchError structs
- `ec1c8b5` feat(types): add SpecialMove per ADR-0013
- `e0665ab` feat(cache): on-disk HTTP cache per ADR-0004
- `9d90d48` feat(parser): root package.json reader per ADR-0011
- `b156a23` feat(stats): Rarity, Chaos, policy per ADR-0009/0010
- `0dd7244` feat(flavor): rule-based generator per ADR-0005
- `1e584ee` feat(templates): self-contained deck.html per ADR-0006
- `b55c8af` feat(render): Render(deps, io.Writer) per ADR-0006

Tickets pending (issue files under `.scratch/depdeck-v1/issues/`):

- `04-pkg-registry.md` (`04-registry`)
- `07-internal-pipeline.md` (`07-pipeline`)
- `08-internal-cli.md` (`08-cli`)
- `11-internal-mcp.md` (`11-mcp`)
- `12-cmd-depdeck.md` (`12-cmd`)
- `13-testdata.md` (sixth remaining ticket; not in the original next-session list)

Docs complete: `SPEC.md`, `DECISIONS.md`, `CONTEXT.md`, `DESIGN.md`, `SECURITY.md`, `DEVELOPMENT.md`, `RULES.md`, `PHASES.md`, `AGENTS.md`. ADRs `0001`–`0015` in `docs/adr/`. Latest: `f22352c` docs(adr): Grok Build primary (ADR-0015).

Working tree (do not treat as ticket work): `M AGENTS.md`, `M .gitignore`, `M skills-lock.json`; untracked `.agents/skills/*`, `.grok/config.toml`, `.grok/skills/`, `.claude/`, `agent/`.

MCP from `grok inspect` (names only; do not copy URLs or tokens):

- User config, enabled: firecrawl, engram, headroom, playwright, grok-bot, github, context7
- Project config: graft (enabled); context-mode (`enabled = false`, absent from the inspect list)
- Also present from plugins/compat: cloudflare-*, railway, tavily, wmux, beyondbeings, chrome-devtools, and disabled/compat leftovers (21st, redesign)

Skills recently installed for this harness:

- `task-observer` — `.grok/skills/task-observer` and `.agents/skills/task-observer`
- `karpathy-guidelines` — `.grok/skills/karpathy-guidelines`
- `verify-depdeck` — `.cursor/skills/verify-depdeck` (`e8ec477`)
- `unlazy` — `.grok/skills/unlazy`
- `graft` — `.grok/skills/graft`
- `talk-to-grok-bot` — `~/.grok/skills/talk-to-grok-bot`

## Next steps for the fresh session

1. Start a **new** Grok Build session so `grok-bot` MCP tools enter the catalog (`search_tool` / `use_tool` as `grok-bot__gbot_send`, `grok-bot__gbot_thread`).
2. Run `grok mcp doctor grok-bot` and confirm 12 tools still green.
3. Retry `gbot bots list`. If still 417, wait and retry. Do not debug Grok Build.
4. Begin ticket `04-registry`:
   - Read `.scratch/depdeck-v1/issues/04-pkg-registry.md` (`Status: ready-for-agent`). There is no separate session-notes file; the ticket body is the plan source (ADRs 0003, 0004; blocked by 01-types and 02-cache, both merged).
   - Shape: plan → agent → verify → debug.
   - Create `pkg/registry/` with httptest only; never live npm.

## Context the next agent needs

- Repo: `C:\Users\Aryan Singh\dev\depdeck`
- Primary agent: Grok Build (ADR-0015). Cursor secondary.
- Rules: `.cursor/rules/*.mdc` is source of truth; sync to `.grok/rules/*.md` with `pwsh -File scripts/sync-rules.ps1`. Do not hand-edit the generated `.md`.
- Shell search: `rg` with an explicit path. Never `Select-String` (`shell-search.mdc`).
- Workflow: plan → agent → verify → debug.
- Commits: Conventional Commits, one per ticket. Do not push without approval.
- JSON envelope and authorities: `AGENTS.md`.
- Grok Bot usage: `AGENTS.md` **Grok Bot** plus `talk-to-grok-bot`. Manual send/thread on Windows.

Suggested skills to load: `task-observer` (session start), `graft` (before hunting registry code), `karpathy-guidelines`, `tdd` / `golang-testing` for `04-registry`, `talk-to-grok-bot` only if coordinating with a named bot.

## Rollback

If grok-bot MCP misbehaves after restart:

```powershell
Copy-Item "$env:USERPROFILE\.grok\config.toml.bak-gbot-20260918-100649" `
  "$env:USERPROFILE\.grok\config.toml" -Force
```
