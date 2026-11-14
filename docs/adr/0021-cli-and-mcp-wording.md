# ADR-0021: Explain CLI Check results and MCP tool purpose

**Date:** 2026-11-15
**Status:** accepted
**Supersedes:** The tool descriptions in ADR-0007. Clarifies plain Check output under ADR-0019.

## Decision

Use the tool descriptions in [SPEC](../SPEC.md#mcp-surface) for discovery.
`list_deck` assesses direct dependencies declared by a project. `get_card`
looks up one public npm package independently of project membership.
`check_deck` evaluates the default publish-age and license-declaration policy.
Chaos reflects publish age. Rarity reflects weekly downloads. Fetch errors
indicate unavailable facts.

Plain Check writes one verdict line to stdout in natural language. It
distinguishes a pass, an empty project, partial Fetch errors, policy flags,
and failure to fetch data for every dependency. Scripts can use the exit
status or `--json` and its `data.reason` field.

Keep the CLI commands, flags, streams and exit codes, MCP tool names and
input schemas, JSON schema 1.1, reason codes, and assessment policy. Report
Fetch errors as data and operational failures as tool errors.

## Compatibility

Plain Check stdout changes wording. Integrations that parse the old line can
use the existing JSON or exit-status interface. MCP tool descriptions change
in discovery metadata.

## Verification

Use offline fixtures to verify Check output, discovery descriptions, and
tool errors. Preserve the existing regressions for degraded Cards, the empty
Check array, and structured content.
