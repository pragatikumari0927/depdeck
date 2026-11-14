# ADR-0018: Assessment and envelope shapes

**Date:** 2026-11-02
**Status:** accepted

## Decision

Add `Dependencies` to Check data. Keep the existing `pass`, `flagged`, `reason`, and `policy` fields. Keep project Dependency fields and JSON tags. A standalone MCP Card omits project-only `version`, `tag`, and `resolved_version` fields.

Use schema version `1.1` as the first supported shipped baseline. The documented MCP 1.0 shape was not shipped. Consumers check the major version only. Keep the envelope `{schema_version, generated_at, generated_by, tool, data}`.

## Reason

Check results need the assessed dependencies to show registry fetch errors. A standalone Card has npm facts but no project membership. Schema 1.1 records the first supported shared contract without changing a previously shipped contract.

## Evidence and maintenance

Ticket 01 owns types and JSON tests. Tickets 07 and 11 own shared pipeline and MCP behavior. Change the major version before removing required fields from a shipped schema.
