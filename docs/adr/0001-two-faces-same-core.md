# Two faces, same core

Depdeck is both a human CLI (cards, HTML) and an MCP server (agent queries). Wrapping MCP around the CLI would shell out, parse stdout, and drift; making the CLI the only product would leave agents without a forced machine-readable path. We share one pipeline (parse → fetch → stats → flavor → Dependencies) and two thin adapters.

**Status:** accepted

## Consequences

- `pkg/types` is the shared contract: CLI `--json` and MCP tool responses are the same struct, one schema, one version number.
- Exit codes are a CLI concern. The pipeline returns `error`; MCP maps that to a structured tool response.
- The pipeline must be callable from tests with no `os.Stdout`, `flag.Parse`, or `os.Exit`.
- `--flavor` defaults to `rule` on both faces. MCP does not silently invoke a local model. `none` means the same thing in both Faces.

If any of these are violated, the Faces drift and MCP returns things the CLI never would.
