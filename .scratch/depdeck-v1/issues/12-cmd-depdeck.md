Status: resolved

# cmd/depdeck — wiring

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

`main` that dispatches CLI vs `mcp` subcommand (or `depdeck mcp` running the MCP stdio server). Wires default cache dir and registry URLs. No business logic.

## Files

- Create `cmd/depdeck/main.go`
- Create `go.mod` module path `depdeck` (or `github.com/.../depdeck` if remote exists — use `depdeck` until git remote exists)

## Interface

- `depdeck scan [path]`
- `depdeck check [path]`
- `depdeck mcp` — stdio MCP

`main` calls `os.Exit(cli.Main(os.Args[1:], ...))`.

## Acceptance criteria

- [x] `go build -o depdeck ./cmd/depdeck` succeeds
- [x] `depdeck --help` exits 0, no command exits 2 with usage, and `scan` without a path uses the current directory
- [x] Binary contains no third Source besides npm in default wiring

## Blocked by

- `.scratch/depdeck-v1/issues/08-internal-cli.md`
- `.scratch/depdeck-v1/issues/11-internal-mcp.md`

## ADRs

0001

## V1 contract amendment

This amendment overrides the earlier module-path fallback and Go-version target.

Use the repository's canonical remote module path and require Go 1.25.0. Pin the official MCP Go SDK at v1.8.0. `scan` and `check` without a path use the current directory. Include build identity in envelopes; use `depdeck dev` for unversioned builds. Follow [ADR-0017](../../../docs/adr/0017-go-runtime-and-npm-version-ranges.md) and [ADR-0019](../../../docs/adr/0019-cli-path-and-output-contract.md).

## Completion evidence

Added `cmd/depdeck` with signal cancellation, build identity, and dispatch for `scan`, `check`, and `mcp`. Updated the module and internal imports to `github.com/pragatikumari0927/depdeck`. Production assessment uses the default public npm endpoints.

- `go test ./... -count=1` passed on Windows amd64 with Go 1.25.0 and Go 1.27.0. The suite built and ran the real executable and exercised help, version, current-directory and file paths, exit codes 0–3, JSON, HTML output, write failures, and MCP stdio. Seeded-cache tests recorded zero outbound attempts. Fresh-fetch tests used a separate fixture subprocess.
- Isolated `go install ./cmd/depdeck` passed and the installed executable reported `depdeck dev`.
- `go vet ./...` and `golangci-lint run ./...` passed on Go 1.27.0. Windows race testing requires 64-bit gcc, which was unavailable here.