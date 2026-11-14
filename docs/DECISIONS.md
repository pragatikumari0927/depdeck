# Decisions

Index only. Reasoning lives in `docs/adr/`.

## Locked

| ID | Decision | ADR |
|----|----------|-----|
| D-001 | Two faces (CLI + MCP), one pipeline | [0001](adr/0001-two-faces-same-core.md) |
| D-002 | Deck contains direct dependencies only | [0002](adr/0002-directs-only-deck.md) |
| D-003 | Degraded cards: keep the card, attach fetch error | [0003](adr/0003-degraded-cards-and-fetch.md) |
| D-004 | On-disk HTTP cache, ETag, three TTL classes | [0004](adr/0004-on-disk-http-cache.md) |
| D-005 | v1 flavor: rule and none only; ai errors | [0005](adr/0005-v1-flavor-rule-or-none.md) |
| D-006 | One self-contained deck.html, byte-deterministic | [0006](adr/0006-self-contained-deck-html.md) |
| D-007 | MCP: list_deck, get_card, check_deck | [0007](adr/0007-mcp-three-tools.md) |
| D-008 | Shared JSON envelope across CLI and MCP | [0008](adr/0008-json-envelope.md) |
| D-009 | Rarity from npm weekly downloads | [0009](adr/0009-rarity-thresholds.md) |
| D-010 | v1 Check policy numbers | [0010](adr/0010-check-policy-v1.md) |
| D-011 | Root-only parser; Tag is one enum | [0011](adr/0011-root-roster-and-tag-enum.md) |
| D-012 | Security threat model: five threats and mitigations | [0012](adr/0012-security-threat-model.md) |
| D-013 | SpecialMove added to Dependency; pkg/types unfrozen for one field | [0013](adr/0013-add-special-move.md) |
| D-014 | pkg/types freeze relaxed: additive fields allowed without ADR | [0014](adr/0014-freeze-policy-relaxed.md) |
| D-015 | Grok Build is primary agent; Cursor secondary; rules sync via script | [0015](adr/0015-grok-build-primary-agent.md) |
| D-016 | Canonical docs and security guidance | [0016](adr/0016-canonical-docs-and-security-guidance.md) |
| D-017 | Go 1.25 and pinned npm/MCP dependencies | [0017](adr/0017-go-runtime-and-npm-version-ranges.md) |
| D-018 | Assessment, standalone Card, and schema 1.1 | [0018](adr/0018-assessment-and-envelope-shapes.md) |
| D-019 | CLI current-directory defaults and output contract | [0019](adr/0019-cli-path-and-output-contract.md) |
| D-020 | Keep the upstream Go security minimum | [0020](adr/0020-go-security-minimum.md) |

## Standing conventions

- Pipeline is a library: no `os.Stdout`, `flag.Parse`, or `os.Exit`.
- v1 `FetchError.Source` is always `"npm"`; do not invent a second Source.
- HTML is byte-identical for the same input; JSON is not (`generated_at` on the envelope).
- Tests never call live npm; use fixtures and `--no-cache`.
- Tag is a single string enum (`dependencies` | `devDependencies` | `optionalDependencies`).
- `--flavor=ai` errors (CLI exit 2); never silent fallback to `rule`.

## v2

Deferred: GitHub enrichment, lockfile pinning, workspace walk, `--flavor=ai` runtime, `depdeck serve`, audit/compare MCP tools, private-registry auth, transitives on the Deck.
