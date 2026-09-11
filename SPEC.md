# SPEC — depdeck v1

Frozen technical contract. Claims are checkable against code. Index of why: [DECISIONS.md](DECISIONS.md). Changing a marked claim requires changing the cited ADR first.

## Non-goals

v1 does not implement: GitHub enrichment; lockfile pinning (`resolved_version` always omitted this batch); workspace walk; `--flavor=ai` runtime (flag errors); `depdeck serve`; MCP `render_html` / `audit` / `search_deck` / `compare_decks`; private-registry auth; transitives, `peerDependencies`, or `bundledDependencies` on the Deck.

Requires ADR-0002, ADR-0005, ADR-0007, ADR-0011, and DECISIONS.md v2 to change.

## Package map

| Package | Responsibility |
|---------|----------------|
| `pkg/types` | Shared Dependency, FetchError, Envelope, Check types |
| `pkg/cache` | On-disk HTTP cache |
| `pkg/parser` | Root `package.json` Roster |
| `pkg/registry` | npm fetch, semaphore 8, cache-aware |
| `pkg/stats` | Rarity, Chaos, Check flagging |
| `pkg/flavor` | `rule` / `none`; `ai` is an error |
| `pkg/render` | HTML bytes from Dependencies |
| `internal/pipeline` | parse → fetch → stats → flavor |
| `internal/cli` | `scan`, `check`, flags, exit codes |
| `internal/mcp` | `list_deck`, `get_card`, `check_deck` |
| `cmd/depdeck` | Wiring only |
| `pkg/render/templates/` | deck.html, embedded via go:embed |
| `testdata/` | Fixture npm responses |

`internal/pipeline` has no `os.Stdout`, `flag.Parse`, or `os.Exit`. Requires ADR-0001 to change.

## Core types

Checkable against `pkg/types`. JSON names:

- Dependency: `name`, `version` (always declared range/exact), `resolved_version` (omitempty), `tag`, `weekly_downloads` (omitempty), `rarity` (omitempty), `chaos` (omitempty; Go representation is *float64 so omitempty distinguishes unknown from zero.), `last_publish` (omitempty), `license` (omitempty), `flavor_text` (omitempty), `special_move` (omitempty), `flavor_source` (omitempty), `errors` (omitempty).
- No `stars` or `issues` fields.
- `tag`: `dependencies` | `devDependencies` | `optionalDependencies`.
- `rarity`: `Common` | `Rare` | `Epic` | `Legendary` or omitted.
- `flavor_source`: `rule_full` | `rule_name_only` or omitted.
- FetchError: `source`, `kind`, `detail`. `source` is always `"npm"` in v1. `kind`: `not_found` | `timeout` | `rate_limit` | `network` | `parse`. Go representation is a string field with exported constants; SPEC enum lists the permitted values.
- Empty Rarity is not Common.

Requires ADR-0001, ADR-0002, ADR-0003, ADR-0011.

## Envelope

Successful CLI `--json` and MCP results:

```json
{
  "schema_version": "1.0",
  "generated_at": "<RFC3339 UTC>",
  "generated_by": "depdeck x.y.z",
  "tool": "<name>",
  "data": {}
}
```

| `tool` (CLI) | `tool` (MCP) | `data` |
|--------------|--------------|--------|
| `scan` | `list_deck` | `{ "dependencies": [ Dependency, ... ] }` |
| `check` | `check_deck` | `{ "pass", "flagged", "reason", "policy" }` |
| (none) | `get_card` | one Dependency object |

`schema_version` is `"1.0"`. Consumers check the major component only. JSON is not byte-identical across runs (`generated_at`). HTML must not contain `generated_at`.

Requires ADR-0008 to change.

## CLI surface

Commands: `scan [path]`, `check [path]`. Path is `package.json` or its directory.

Flags: `--json`, `--flavor` (default `rule`), `--no-cache`, `--out` (default `deck.html` next to that `package.json`; `-` writes HTML to stdout).

- `--json`: envelope on stdout; no HTML file.
- Default `scan`: write HTML; one line on stderr `Wrote deck.html (N cards)`.
- No `--both`.
- `--flavor=ai`: exit `2`; stderr contains `ai` and `v1.1`; no rule Flavor; no HTML.

Exit codes: `0` success / check pass; `1` check flagged; `2` usage or `--flavor=ai`; `3` check when every Dependency has fetch Errors.

Requires ADR-0001, ADR-0005, ADR-0006, ADR-0010.

## MCP surface

Tools only: `list_deck`, `get_card`, `check_deck`. Explicit `inputSchema`. `flavor` enum is `rule` | `none` (`ai` is schema failure, not runtime fallback).

- `list_deck`: `{ path, tag?, flavor? }`
- `get_card`: `{ name, flavor? }` — npm only; no Roster path
- `check_deck`: `{ path }`

`isError` only for malformed args or crashes. npm 404 / 429 / timeout / all-fetch-failed: success with `errors` on Dependencies.

Descriptions:

- `list_deck`: List every direct dependency of a project as Cards. Use for project-wide questions.
- `get_card`: Fetch a single package's Card from npm. Use for package-specific questions.
- `check_deck`: Return which dependencies are flagged as risky, and an overall pass/fail. Use before shipping or for CI gating.

Requires ADR-0007, ADR-0003, ADR-0005.

## Check policy

Returned as `data.policy`. v1 defaults:

| Field | Value | Flag when |
|-------|--------|-----------|
| `chaos_score_max` | `0.5` | Chaos greater than 0.5 |
| `age_years_max` | `null` | not enforced |
| `require_license` | `true` | license empty after successful fetch |
| `fail_on_fetch_errors` | `false` | never flag solely for fetch errors |

Degraded Cards (non-empty `errors`) are not flagged for Chaos, age, or license.

Requires ADR-0010, ADR-0003.

## Parser

Reads the root `package.json` only. Ignores `workspaces` and nested manifests. Emits directs from `dependencies`, `devDependencies`, `optionalDependencies`. Skips `peerDependencies` and `bundledDependencies`. Duplicate name: first Tag in order `dependencies`, `devDependencies`, `optionalDependencies`. No lockfile read. `version` is the declared string. `resolved_version` is empty.

Requires ADR-0002, ADR-0011.

## Stats

Weekly downloads → Rarity: `>=10000` Legendary, `>=5000` Epic, `>=1000` Rare, else Common (successful fetch only).

Chaos from last npm publish vs now: within 2 years `0.0`; within 5 years `0.6`; older or missing last-publish on successful fetch `1.0`. Omitted on degraded Cards.

Requires ADR-0009, ADR-0010.

## Flavor

Modes: `rule` (default), `none`. Mode `ai` returns an error whose message contains `ai` and `v1.1`. CLI maps that to exit `2`. No silent `rule` fallback.

`rule` + `len(errors) > 0` → `flavor_source` `rule_name_only` (name hash). Else `rule_full` (name + stats). Zero downloads is not degraded. `none` clears `flavor_text` and `flavor_source`. Templates do not mention GitHub issues.

Requires ADR-0005.

## Render

One self-contained HTML file. `html/template` only. No `template.HTML` / `template.JS` on fetched strings. Filter `<script>` is a static template string. Same Dependency slice (stable sort by `name`) → identical bytes. No timestamps or random IDs. Default path: `deck.html` next to scanned `package.json`. Target under 150 KB uncompressed for 47 Cards. No `depdeck serve`. No per-card logos in v1.

Requires ADR-0006.

## Cache

Root: `os.UserCacheDir()/depdeck/` unless overridden. One file per package: filename `sha256(source + ":" + name + ":" + ttl_class)`. Value JSON `{etag, status, fetched_at, body}`.

TTL: 200+body 24h; 304 resets 24h; 404 1h. Do not cache 429, 5xx, timeouts, network errors. `--no-cache` skips reads and writes. `check` in CI reads cache. Tests never use the real user cache dir. Cache is keyed by package name, not project.

Requires ADR-0004.

## Test seam

One seam: `pipeline.Run` and `pipeline.Check` return `[]Dependency` / `CheckData`. Tests inject fixture HTTP and a temp cache. Assert observable JSON, Check pass/flagged/policy, HTML bytes/escaping/path, CLI exit codes, MCP `isError` vs `errors`. Tests never dial `registry.npmjs.org` or `api.npmjs.org`.

Parser negative: `testdata/workspace/` root vs nested `package.json` — only root Roster names.

Render: call `Deck` twice; `bytes.Equal`.

Requires ADR-0001, ADR-0004, ADR-0006, ADR-0011.
