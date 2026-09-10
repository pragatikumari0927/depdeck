# Decisions

Index only. Reasoning lives in `docs/adr/`.

## Locked

| ADR | Title |
|-----|-------|
| [0001](docs/adr/0001-two-faces-same-core.md) | Two faces, same core |
| [0002](docs/adr/0002-directs-only-deck.md) | Directs-only Deck from the Roster |
| [0003](docs/adr/0003-degraded-cards-and-fetch.md) | Degraded Cards: keep the Card, structured fetch errors |
| [0004](docs/adr/0004-on-disk-http-cache.md) | On-disk HTTP cache |
| [0005](docs/adr/0005-v1-flavor-rule-or-none.md) | v1 Flavor is rule or none; ai errors |
| [0006](docs/adr/0006-self-contained-deck-html.md) | Human Deck is one self-contained deck.html |
| [0007](docs/adr/0007-mcp-three-tools.md) | v1 MCP: three tools, one envelope |
| [0008](docs/adr/0008-json-envelope.md) | Same JSON envelope for CLI --json and MCP |
| [0009](docs/adr/0009-rarity-thresholds.md) | Rarity from npm weekly downloads |
| [0010](docs/adr/0010-check-policy-v1.md) | v1 Check policy numbers |
| [0011](docs/adr/0011-root-roster-and-tag-enum.md) | Root-only workspaces; Tag is one enum |
| D-012 | Security threat model: five threats and mitigations | [0012](docs/adr/0012-security-threat-model.md) |

## Standing conventions

- Pipeline is a library: no `os.Stdout`, `flag.Parse`, or `os.Exit`.
- v1 `FetchError.Source` is always `"npm"`; do not invent a second Source.
- HTML is byte-identical for the same input; JSON is not (`generated_at` on the envelope).
- Tests never call live npm; use fixtures and `--no-cache`.
- Tag is a single string enum (`dependencies` | `devDependencies` | `optionalDependencies`).
- `--flavor=ai` errors (CLI exit 2); never silent fallback to `rule`.

## v2

Deferred: GitHub enrichment, lockfile pinning, workspace walk, `--flavor=ai` runtime, `depdeck serve`, audit/compare MCP tools, private-registry auth, transitives on the Deck.
