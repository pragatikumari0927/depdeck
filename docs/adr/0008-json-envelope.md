# Same JSON envelope for CLI `--json` and MCP

CLI `--json` and MCP successful results share one envelope: `schema_version`, `generated_at` (RFC3339 UTC), `generated_by` (`depdeck x.y.z`), `tool`, `data`. Dependency objects inside `data` are the shared contract. HTML stays byte-identical and must not include `generated_at`; JSON is not byte-deterministic. That is intentional.

**Status:** accepted

## `tool` names stay Face-specific

| CLI `--json` `tool` | MCP `tool` |
| --- | --- |
| `scan` | `list_deck` |
| `check` | `check_deck` |
| (no v1 command) | `get_card` |

Do not rename CLI verbs to MCP names or the reverse. Parsers that only need facts read `data` and ignore `tool`. Mapping lives in `AGENTS.md`.

## `data` shapes

- `scan` / `list_deck`: `{ "dependencies": [ ...Dependency ] }`
- `check` / `check_deck`: `{ "pass", "flagged", "reason", "policy" }`
- `get_card`: one Dependency object (fields at `data`, not nested under `dependencies`)

## Versioning

`schema_version` is `"1.0"`. Minor = additive; major = breaking. Consumers check the major component only.

## jq cost

One extra `.data` level in shell pipelines. Document the two common recipes in the README when the CLI exists (`rarity == Legendary`, `flagged | length`). No `--both`, no flattening flag in v1.
