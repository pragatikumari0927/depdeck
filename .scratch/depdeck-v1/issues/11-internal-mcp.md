Status: resolved

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

- [x] Successful JSON has schema_version, generated_at, generated_by, tool, data
- [x] get_card without path; 404 fixture → isError false, errors kind not_found, source npm
- [x] flavor ai in arguments → protocol/schema error (isError), not rule Flavor
- [x] list_deck tool name is `list_deck` not `scan`
- [x] All-fetch-failed list_deck is still isError false
- [x] Tests use pipeline fixture, no live npm

## Blocked by

- `.scratch/depdeck-v1/issues/07-internal-pipeline.md`

## ADRs

0001, 0003, 0005, 0007, 0008

## V1 contract amendment

This amendment overrides earlier conflicting SDK, schema, and Card shape details.

Use `github.com/modelcontextprotocol/go-sdk` v1.8.0. CLI and MCP use envelope schema 1.1 and the same pipeline. For JSON-RPC success, npm fetch failures remain successful tool results with dependency errors. Keep `isError` for invalid arguments or tool failures. `CheckData` includes `Dependencies`. Standalone `get_card` omits `version`, `tag`, and `resolved_version`. Enforce SDK frame/input limits. Follow [ADR-0017](../../../docs/adr/0017-go-runtime-and-npm-version-ranges.md) and [ADR-0018](../../../docs/adr/0018-assessment-and-envelope-shapes.md).

## Completion evidence

Implemented `list_deck`, `get_card`, and `check_deck` over the shared pipeline with the pinned Go SDK. The tools advertise explicit schemas, reject invalid arguments as tool errors, and preserve npm failures in successful assessment envelopes. The standalone Card omits the three Roster-only fields. Stdio uses the SDK's 16 MiB inbound limit.

- `go test ./... -count=1` passed on Windows amd64 with Go 1.25.0 and Go 1.27.0. Real-binary tests used seeded cache fixtures and recorded zero outbound attempts. A test subprocess used localhost HTTP fixtures to exercise npm 404, 429, malformed metadata, all-fetch-failed assessments, and request cancellation.
- Protocol tests verified initialization, exactly three tools, schema and argument rejection, JSON text and structured content parity, panic sanitization, malformed JSON, oversized input, and clean EOF shutdown.
- `go vet ./...` and `golangci-lint run ./...` passed on Go 1.27.0. Windows race testing requires 64-bit gcc, which was unavailable here.