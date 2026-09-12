# ADR-0015: Grok Build is primary agent; Cursor secondary; rules sync via script

**Date:** 2026-09-12
**Status:** accepted
**Deciders:** project

## Context

Cursor Pro quota was depleted mid-v1 (both Cursor Models and Other Models
at 100%). Grok Build (SuperGrok) was already installed and is the active
agent CLI. Cursor remains installed and functional; quota resets monthly.

## Decision

Grok Build is the primary agent CLI. Cursor is secondary, retained.

- `.cursor/rules/*.mdc` is the source of truth for rules. `.grok/rules/*.md`
  is generated from it via `scripts/sync-rules.ps1`. Grok Build does not
  read `.mdc`; every `.md` in `.grok/rules/` loads unconditionally.
- `.cursor/mcp.json` (JSON) and `.grok/config.toml` (TOML) are both
  maintained. Duplication is intentional; formats differ and neither
  tool reads the other's config.
- `.agents/skills/` is native to Grok. `.cursor/skills/` loads via Grok's
  Cursor compatibility layer.
- `AGENTS.md` remains the cross-tool contract read by both.
- `.grok/hooks/` is empty. The Go post-edit chain lives as instruction
  text inside `golang-hooks.mdc` and its `.grok` copy. No real hook
  infrastructure for v1.

## Consequences

- Rule changes go in `.mdc` only. Run `pwsh -File scripts/sync-rules.ps1`
  after any edit. The generated `.md` files are not hand-edited.
- Adding an MCP server requires updating both `.cursor/mcp.json` and
  `.grok/config.toml`.
- Cursor quota reset (Oct 4) restores Cursor as a working fallback with
  zero migration cost — the configs are parallel, not replacement.
