## Agent skills

### Issue tracker

Issues live as markdown files under `.scratch/<feature>/`. External PRs are not a triage surface. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

### JSON envelope (CLI `--json` and MCP)

Same wrapper: `{schema_version, generated_at, generated_by, tool, data}`. Parse `data` by shape; map `tool` names:

- `scan` → `list_deck`
- `check` → `check_deck`
- MCP-only: `get_card` (no CLI single-card command in v1)

Consumers must check the **major** component of `schema_version` only (e.g. prefix `1.`), not exact `"1.0"`.

## Authorities
- SPEC.md — the frozen technical contract
- DECISIONS.md — index of the decision trail
- CONTEXT.md — glossary
- docs/adr/ — per-decision reasoning

