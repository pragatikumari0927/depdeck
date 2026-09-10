Status: ready-for-agent

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

- [ ] `go build -o depdeck ./cmd/depdeck` succeeds
- [ ] `depdeck --help` or scan without args exits 2 with usage (not panic)
- [ ] Binary contains no third Source besides npm in default wiring

## Blocked by

- `.scratch/depdeck-v1/issues/08-internal-cli.md`
- `.scratch/depdeck-v1/issues/11-internal-mcp.md`

## ADRs

0001
