# SPEC — depdeck v1

Frozen v1 technical contract. Claims are checkable against code. Index of why: [DECISIONS.md](DECISIONS.md). Successor ADRs 0016–0019 update the original decisions for v1. Changing a marked claim requires changing the cited ADR first.

## Non-goals

v1 does not implement: GitHub enrichment; lockfile resolution (`resolved_version` stays empty); workspace walk; `--flavor=ai` runtime (flag errors); `depdeck serve`; MCP `render_html` / `audit` / `search_deck` / `compare_decks`; private-registry auth; transitives, `peerDependencies`, or `bundledDependencies` on the Deck. The registry selects a version that matches the declared npm range.

Requires ADR-0002, ADR-0005, ADR-0007, ADR-0011, and DECISIONS.md v2 to change.

## Package map

| Package | Responsibility |
|---------|----------------|
| `pkg/types` | Shared Dependency, FetchError, Envelope, Check types |
| `pkg/cache` | On-disk HTTP cache |
| `pkg/parser` | Root `package.json` Roster |
| `pkg/registry` | npm fetch, range selection, semaphore 8, cache-aware |
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

## Runtime flow

`depdeck scan [path]` executes:

  1. Parse     internal/cli reads <path>/package.json via pkg/parser
  2. Cache     pkg/cache checks for cached npm responses
  3. Fetch     pkg/registry fetches uncached deps, bounded at 8 concurrent
  4. Stats     pkg/stats.Apply computes Rarity, Chaos per dep
  5. Flavor    pkg/flavor.Apply fills SpecialMove, FlavorText, FlavorSource
  6. Render    pkg/render.Render writes deck.html (unless --json)

`depdeck check [path]` is identical through step 5, then evaluates
pkg/stats.Flagged and exits 1 if any dependency is flagged.

Both CLI and MCP call the same pipeline. See ADR-0001.

## Core types

Checkable against `pkg/types`. JSON names:

- Dependency: `name`, `version` (always declared range/exact), `resolved_version` (omitempty), `tag`, `weekly_downloads` (omitempty), `rarity` (omitempty), `chaos` (omitempty; Go representation is *float64 so omitempty distinguishes unknown from zero.), `last_publish` (omitempty), `license` (omitempty), `flavor_text` (omitempty), `special_move` (omitempty), `flavor_source` (omitempty), `errors` (omitempty).
- No `stars` or `issues` fields.
- `tag`: `dependencies` | `devDependencies` | `optionalDependencies`.
- `rarity`: `Common` | `Rare` | `Epic` | `Legendary` or omitted.
- `flavor_source`: `rule_full` | `rule_name_only` or omitted.
- FetchError: `source`, `kind`, `detail`. `source` is always `"npm"` in v1. `kind`: `not_found` | `timeout` | `rate_limit` | `network` | `parse`. Go representation is a string field with exported constants; SPEC enum lists the permitted values.
- Empty Rarity is not Common.
- Check data contains `pass`, `flagged`, `reason`, `policy`, and the full assessed `dependencies` list.
- Standalone MCP Card output contains npm facts and errors. It omits project-only `version`, `tag`, and `resolved_version`.
- Version matching uses `github.com/Masterminds/semver/v3` v3.5.0. Test supported npm range forms against npm behavior. Do not silently replace an unsupported range with `latest`.

Requires ADR-0001, ADR-0002, ADR-0003, ADR-0011.

## Envelope

Successful CLI `--json` and MCP results:

```json
{
  "schema_version": "1.1",
  "generated_at": "<RFC3339 UTC>",
  "generated_by": "depdeck x.y.z",
  "tool": "<name>",
  "data": {}
}
```

| `tool` (CLI) | `tool` (MCP) | `data` |
|--------------|--------------|--------|
| `scan` | `list_deck` | `{ "dependencies": [ Dependency, ... ] }` |
| `check` | `check_deck` | `{ "pass", "flagged", "reason", "policy", "dependencies" }` |
| (none) | `get_card` | one standalone Card, without project-only fields |

`schema_version` is `"1.1"`, the first supported shipped baseline. The earlier documented MCP 1.0 shape was not shipped. Consumers check the major component only. JSON is not byte-identical across runs (`generated_at`). HTML must not contain `generated_at`.

Requires ADR-0008 to change.

## CLI surface

Commands: `scan [path]`, `check [path]`. Path is optional. Without it, use the current directory and read its `package.json`. A supplied path can name `package.json` or its directory. A missing or invalid manifest returns exit `2`.

Flags: `--json`, `--flavor` (default `rule`), `--no-cache`, `--out` (default `deck.html` next to that `package.json`; `-` writes HTML to stdout).

- `--json`: envelope on stdout; no HTML file.
- Default `scan`: write HTML; one line on stderr `Wrote deck.html (N cards)`.
- No `--both`.
- `--flavor=ai`: exit `2`; stderr contains `ai` and `v1.1`; no rule Flavor; no HTML.

Exit codes: `0` success / check pass; `1` check flagged; `2` usage or `--flavor=ai`; `3` check when every Dependency has fetch Errors.

Requires ADR-0001, ADR-0005, ADR-0006, ADR-0010.

## MCP surface

Tools only: `list_deck`, `get_card`, `check_deck`. Use the official Go SDK v1.8.0 and explicit `inputSchema`. `flavor` enum is `rule` | `none` (`ai` is schema failure, not runtime fallback). The minimum Go version is `1.26.9` per ADR-0020.

- `list_deck`: `{ path, tag?, flavor? }`
- `get_card`: `{ name, flavor? }` — npm only; no Roster path
- `check_deck`: `{ path }`

The CLI and MCP use the same pipeline. MCP project paths follow the caller's working directory. Standalone Card responses omit `version`, `tag`, and `resolved_version`.

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

Reads the root `package.json` only. Rejects manifest files larger than 4 MiB. Ignores `workspaces` and nested manifests. Emits directs from `dependencies`, `devDependencies`, `optionalDependencies`, in Tag order and sorted by name within each Tag. Skips `peerDependencies` and `bundledDependencies`. Duplicate name: first Tag in order `dependencies`, `devDependencies`, `optionalDependencies`. No lockfile read. `version` is the declared string. `resolved_version` is empty.

Requires ADR-0002, ADR-0011.

## Registry

For npm metadata, select `dist-tags.latest` only when it is published and satisfies the declared range. Otherwise select the highest matching stable version. A prerelease is selected only when `latest` explicitly names it and the range allows it under npm's same major, minor, and patch rule. Read `license` and `last_publish` from the selected version; `last_publish` means its publish time at assessment, not the document modification time. Preserve the declared version in `version`; leave `resolved_version` empty. Unsupported declarations produce npm `parse` errors; no match produces npm `not_found`.

Bound successful HTTP JSON bodies to 32 MiB for metadata and 256 KiB for downloads. Classify non-200 responses by status without reading their bodies. A 304 refreshes a valid stale metadata entry. Reject cross-origin and scheme-changing redirects, and allow at most ten same-origin redirects.

Requires ADR-0017, ADR-0004.

## Stats

Weekly downloads → Rarity: `>=10000` Legendary, `>=5000` Epic, `>=1000` Rare, else Common (successful fetch only). A valid zero is present as `weekly_downloads: 0`. An unknown count is omitted. Failure to fetch or parse downloads degrades the Card; unknown values are not zero.

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

Root: `os.UserCacheDir()/depdeck/` unless overridden. One file per package: filename `sha256(source + ":" + name + ":" + ttl_class)`. Value JSON `{etag, status, fetched_at, body}`. Read encoded metadata entries up to 48 MiB and downloads entries up to 384 KiB. Oversized or malformed cache entries are misses.

TTL: 200+body 24h; 304 resets 24h; 404 1h. Do not cache 429, 5xx, timeouts, network errors. `--no-cache` skips reads and writes. `check` in CI reads cache. Tests never use the real user cache dir. Cache is keyed by package name, not project.

Requires ADR-0004.

## Test seam

One seam: `pipeline.Run` and `pipeline.Check` return `[]Dependency` / `CheckData`. Tests inject fixture HTTP and a temp cache. Assert observable JSON, Check pass/flagged/policy, HTML bytes/escaping/path, CLI exit codes, MCP `isError` vs `errors`. Tests never dial `registry.npmjs.org` or `api.npmjs.org`.

Parser negative: `testdata/workspace/` root vs nested `package.json` — only root Roster names.

Render: call `Deck` twice; `bytes.Equal`.

Requires ADR-0001, ADR-0004, ADR-0006, ADR-0011.
