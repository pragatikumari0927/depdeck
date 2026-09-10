Status: ready-for-agent

# internal/cli — scan, check, flags, exit codes

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

CLI Face: `scan` and `check`. `--json` vs HTML. `--flavor`, `--no-cache`, `--out`. Exit codes. Does not implement HTML itself — calls `render`.

## Files

- Create `internal/cli/cli.go` (and `_test.go`)

## Interface

```go
package cli

func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int
```

Flags: `--json`, `--flavor` (default `rule`), `--no-cache`, `--out` (default deck.html next to package.json; `-` = stdout HTML). Path argument = package.json or directory.

Exit: 0 ok; 1 check failed (flagged); 2 usage / `--flavor=ai`; 3 every fetch failed on check.

`--json`: envelope on stdout, no HTML file. Default scan: write HTML, one line on stderr `Wrote deck.html (N cards)`.

## Acceptance criteria

- [ ] `--flavor=ai` → exit 2; stderr contains `ai` and `v1.1`; no `deck.html`; no rule Flavor JSON
- [ ] `scan --json` → stdout is envelope `tool`=`scan`; no deck.html created
- [ ] `scan` default: file next to package.json not cwd (test with path outside cwd)
- [ ] `check` pass → 0; any flagged → 1; all Dependencies have Errors → 3
- [ ] `--no-cache` passed through (pipeline/cache Disabled)

## Blocked by

- `.scratch/depdeck-v1/issues/07-internal-pipeline.md`
- `.scratch/depdeck-v1/issues/09-pkg-render.md`
- `.scratch/depdeck-v1/issues/10-templates-deck-html.md`

## ADRs

0001, 0005, 0006, 0008, 0010
