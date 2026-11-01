# depdeck

> Turn the dependencies in your `package.json` into a deck of trading cards, and fail CI when one looks risky.

depdeck reads the direct dependencies of an npm project and looks each one up on the npm registry. Each dependency becomes a **card**:
- **Rarity** comes from weekly downloads.
- **Chaos** comes from how long ago the package was last published.
- Each card also shows its license and some flavor text.

The cards are rendered as a single self-contained `deck.html`. The same pipeline also powers a `check` command for CI and an MCP server for AI agents.

**Status: v1 in progress (7 of 13 tickets merged, per [PHASES.md](PHASES.md)).** The core packages are implemented and tested: `types`, `cache`, `parser`, `stats`, `flavor` and `render`. The npm registry client, pipeline, CLI, MCP server and `cmd/depdeck` entrypoint are still to come, so there is no runnable binary yet.

## Features (v1 scope)

- **`depdeck scan [path]`**: parse `package.json` (`dependencies`, `devDependencies`, `optionalDependencies`), fetch npm metadata (8 concurrent requests, on-disk HTTP cache), and write `deck.html`.
- **`depdeck check [path]`**: same pipeline. Exits `1` if any dependency breaks policy. The default policy is chaos ≤ 0.5 and a license is required.
- **`depdeck mcp`**: an MCP server with three tools: `list_deck`, `get_card`, `check_deck`.
- **`--json`**: a versioned JSON envelope (`schema_version: "1.0"`) instead of HTML.
- **Byte-deterministic HTML** that is self-contained (template embedded with `go:embed`).
- **Degraded cards**: npm 404, 429 or timeout errors are shown on the card instead of crashing the run.

### Card stats

| Stat | Rule |
|---|---|
| Rarity | ≥10k weekly downloads: Legendary · ≥5k: Epic · ≥1k: Rare · otherwise: Common |
| Chaos | last publish under 2 years ago: 0.0 · 2–5 years: 0.6 · over 5 years or unknown: 1.0 |
| Flavor | `rule` (default) or `none`; `--flavor=ai` is reserved for v1.1 and currently exits 2 |

### Exit codes
`0` success or check passed · `1` check flagged · `2` usage error or `--flavor=ai` · `3` every dependency failed to fetch

## Install

Requires Go 1.22+.

```bash
git clone https://github.com/pragatikumari0927/depdeck.git
cd depdeck
go test ./...            # runs the implemented packages against fixtures (no network)
# once cmd/depdeck lands:
go build -o depdeck ./cmd/depdeck
```

`go install github.com/pragatikumari0927/depdeck/cmd/depdeck@latest` will work after the planned module-path rename. The module is currently named `depdeck`.

## Usage (target v1 CLI)

```bash
depdeck scan ./my-app                # writes my-app/deck.html
depdeck scan ./my-app --json         # JSON envelope on stdout
depdeck scan . --out -               # HTML to stdout
depdeck check . --no-cache           # CI gate
depdeck mcp                          # stdio MCP server
```

## Project structure

```
pkg/
  types/      shared Dependency, FetchError, Envelope, CheckPolicy types
  cache/      on-disk HTTP cache
  parser/     root package.json roster
  stats/      rarity, chaos, check flagging
  flavor/     rule / none flavor text
  render/     deck.html renderer (templates/deck.html, embedded)
  registry/   (planned) npm client
internal/     (planned) pipeline, cli, mcp
cmd/depdeck/  (planned) entrypoint
testdata/     fixture package.json files
docs/adr/     15 architecture decision records
SPEC.md       frozen v1 technical contract
DESIGN.md, DECISIONS.md, PHASES.md, RULES.md, DEVELOPMENT.md, SECURITY.md
```

## Roadmap

- **v1.1**: `--flavor=ai` via Ollama, PNG card export, `install_size` stat, `depdeck doctor`.
- **v2**: GitHub enrichment, go.mod/Cargo.toml/pyproject parsers, `depdeck battle`, monorepo CI.

See [PHASES.md](PHASES.md).

## Contributing

Trunk-based workflow with Conventional Commits and one commit per ticket. PRs stay under 400 lines. Run `go test ./... -race` before pushing. Read [SPEC.md](SPEC.md) and [DEVELOPMENT.md](DEVELOPMENT.md) first.

## License

TBD.
