# ADR-0025: Render the canonical Deck through MCP

**Date:** 2026-11-23
**Status:** accepted

## Context

The three original MCP tools return structured assessment data. They do not
render HTML. Agent clients can mistake that data for a request to create their
own page, which can produce a visual result that differs from depdeck's
Pixel Archive renderer. The original ADR-0007 statement that MCP has three
tools remains historically accurate, but it no longer defines the complete
tool inventory.

## Decision

Add `render_deck` as an additive MCP tool. It runs the shared pipeline and
`pkg/render.Render`, so the generated file uses the same Cards, Rarity Atlas,
Field Journal, Archive Notices, filters, escaping, and deterministic HTML as
`depdeck scan`.

Inputs are `path` (a project directory or `package.json` file), optional
`out` (a relative `.html` path under the project, default `deck.html`), and
optional `flavor` (`rule` or `none`). The tool creates a new file only. It
does not overwrite an existing file, create parent directories, or write
outside the project. It returns the relative output path, Card count, and
byte size in the existing `{schema_version, generated_at, generated_by,
tool, data}` envelope. npm fetch errors remain ordinary degraded Cards and
appear as Archive Notices in the Deck.

The tool has an intentional filesystem side effect when called. It uses
`os.Root` to constrain output paths to the project and exclusive file creation
to refuse overwrites. The existing assessment tools and their schemas,
outputs, and error behavior do not change. The schema major version remains
1; this adds a tool and a new `data` shape without removing existing fields.

## Consequences

- Agents can ask depdeck to create its own canonical visual Deck instead of
  rebuilding HTML from MCP assessment results.
- MCP clients must display the new `render_deck` tool and its input schema.
- The project setup guide must explain the output path and no-overwrite rule.
- `render_html` remains an unimplemented historical name; the implemented
  tool is `render_deck`.
