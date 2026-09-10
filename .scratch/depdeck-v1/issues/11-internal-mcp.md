Status: ready-for-agent

# internal/mcp — three tools, envelope, schemas

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

MCP Face: `list_deck`, `get_card`, `check_deck`. Same envelope as CLI. Explicit JSON Schema. `isError` only for bad args/crashes. npm 404 is success with errors. `get_card` has no path.

## Files

- Create `internal/mcp/mcp.go` (and `_test.go`)

## Interface

```go
package mcp

func New(opt pipeline.Options) /* MCP server or handler */
// Tools: list_deck {path, tag?, flavor?}
//        get_card {name, flavor?}
//        check_deck {path}
```

list_deck data = DeckData; check_deck data = CheckData; get_card data = one Dependency. flavor enum in schema: rule, none only.

Descriptions exactly as ADR-0007.

## Acceptance criteria

- [ ] Successful JSON has schema_version, generated_at, generated_by, tool, data
- [ ] get_card without path; 404 fixture → isError false, errors kind not_found, source npm
- [ ] flavor ai in arguments → protocol/schema error (isError), not rule Flavor
- [ ] list_deck tool name is `list_deck` not `scan`
- [ ] All-fetch-failed list_deck is still isError false
- [ ] Tests use pipeline fixture, no live npm

## Blocked by

- `.scratch/depdeck-v1/issues/07-internal-pipeline.md`

## ADRs

0001, 0003, 0005, 0007, 0008
