# Agent entry

Use `task-observer` only for skill or workflow improvement work; it is not a required session-start step.

Read [the documentation map](docs/README.md), then follow the authority order in that file. The canonical project contract is [docs/SPEC.md](docs/SPEC.md); the decision index is [docs/DECISIONS.md](docs/DECISIONS.md); domain terms are in [docs/CONTEXT.md](docs/CONTEXT.md).

Issues are Markdown files in `.scratch/<feature>/`. Follow [the issue workflow](docs/agents/issue-tracker.md) and [triage labels](docs/agents/triage-labels.md). Keep tool configs, hooks, rules, and skills at their discovery paths.

For CLI JSON and MCP, use `{schema_version, generated_at, generated_by, tool, data}`. Parse `data` by shape. Map CLI tools `scan` to `list_deck`, and `check` to `check_deck`; MCP also has `get_card`. Check the major schema version only.

Use the `grok-bot` MCP tools to coordinate with the `General` bot when they are available. Send questions and handoffs with `gbot_send`; put who you are and what you need on the first line. Read replies with `gbot_thread`. Do not wait for a reply before continuing.
