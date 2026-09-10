Status: ready-for-agent

# PRD: depdeck v1

## Problem Statement

I want to see the dependencies I actually chose as a small, shareable deck of Pokémon-style cards — and I want agents to query the same facts without shelling out to a CLI and praying the JSON still matches. Transitive scanners already exist; I do not want another worse Socket. I do want one command that writes `deck.html` next to `package.json`, and MCP tools that return the same Cards.

## Solution

depdeck is two Faces over one pipeline: parse the root Roster, fetch npm with a bounded concurrent cache, compute stats/Rarity/Chaos/Flavor, then either write byte-identical HTML, print an envelope JSON, or answer MCP tools. Check flags risky directs using a published policy, without failing CI on a single npm blip.

## User Stories

1. As a developer, I want `depdeck scan` on a project path so that I get `deck.html` beside that project's `package.json`.
2. As a developer who ran scan from `$HOME`, I want the file in the project, not in my home directory, so that I can gitignore or commit it in the right repo.
3. As a developer, I want a one-line stderr confirmation of how many Cards were written so that I know it worked without dumping HTML in the terminal.
4. As a developer, I want `--out -` so that I can redirect HTML in a pipeline.
5. As a developer, I want `--json` to skip HTML and print the envelope on stdout so that scripts can consume Cards.
6. As a developer, I want a second invocation for HTML when I also want JSON, so that v1 stays simple (no `--both`).
7. As a developer, I want `--flavor none` so that Cards have no personality text.
8. As a developer, I want `--flavor rule` (default) so that Flavor comes from metrics when present.
9. As a developer, I want `--flavor ai` to fail with exit code 2 and a v1.1 message so that I never silently get rule Flavor.
10. As a developer, I want `--no-cache` so that tests never read or write the disk cache.
11. As a CI job, I want `depdeck check` to read the cache so that I am not flaky on npm uptime.
12. As a CI job, I want check to exit 0 when the Deck passes policy so that I can gate merges.
13. As a CI job, I want check to exit 1 when any Dependency is flagged so that I notice risk.
14. As a CI job, I want check to exit 3 only when every fetch failed so that one 404 does not fail the build.
15. As a CI job, I want `--flavor ai` to fail before any fetch so that CI never waits on a missing model.
16. As an agent, I want `list_deck` so that I can ask project-wide questions about directs.
17. As an agent, I want `get_card(name)` so that I can ask about a package with no project path.
18. As an agent, I want `check_deck` so that I get `pass`, `flagged`, `reason`, and `policy` without inventing thresholds.
19. As an agent, I want npm 404 as a successful tool result with `errors` so that I do not retry a missing package forever.
20. As an agent, I want malformed arguments as `isError` so that I can fix the call.
21. As an agent, I want `schema_version` and to survive `"1.1"` if I check major `1` only.
22. As an agent, I want `generated_by` so that I can tell which depdeck build produced a payload.
23. As an agent, I want `flavor` enum `rule`|`none` in inputSchema so that `ai` fails validation in v1.
24. As an agent, I want `list_deck` tag filter so that I can see only `devDependencies`.
25. As an HTML viewer, I want client-side Tag toggles so that I can hide devDependencies without rescan.
26. As an HTML viewer, I want degraded Cards to look grey/"no data", not Common, so that I do not trust fake Rarity.
27. As a security-conscious user, I want npm descriptions HTML-escaped so that a malicious package cannot XSS `deck.html`.
28. As a git user, I want two scans of the same input to produce byte-identical HTML so that diffs show real Card changes.
29. As a git user, I accept JSON is not byte-identical because `generated_at` is on the envelope.
30. As a jq user, I want documented recipes for Legendary Cards and flagged count so that `.data` nesting is bearable.
31. As a workspace user, I want v1 to read only the root `package.json` so that I do not get a 400-card monorepo dump.
32. As a lockfile user, I want `resolved_version` when a lockfile exists so that the Card shows what is installed.
33. As a no-lockfile user, I want `version` as the declared range and fetch via dist-tags.latest (then highest matching) so that Cards still enrich.
34. As a Roster author, I want optionalDependencies on the Deck, tagged, so that fsevents-like Cards exist.
35. As a Roster author, I do not want peer or bundled dependencies on the Deck.
36. As a user of three repos that all depend on react, I want one user-cache fetch of react metadata so that I do not triple-hit npm.
37. As a Windows user, I want cache filenames that are hashes so that path length and case are not issues.
38. As an npm-down user, I want the Card kept with a fetch error and name-only Flavor so that the Roster is still visible.
39. As a JSON consumer, I want `flavor_source` `rule_full`|`rule_name_only` so that I know which template ran.
40. As a JSON consumer, I want `version` always declared and `resolved_version` omitted when empty.
41. As a future GitHub-enrichment user, I want concurrency bounded per Source so that GitHub rate limits do not stall npm.
42. As a README reader, I want the pitch "47 Cards vs 800 installs" to stay true because transitives are out of scope.

## Implementation Decisions

- One pipeline, two thin Faces (CLI, MCP). Pipeline is a library: no stdout, flags, or process exit. CLI maps errors to exit codes; MCP maps to `isError` vs data `errors`.
- Shared Dependency contract for CLI `--json` and MCP `data`. Envelope `{schema_version, generated_at, generated_by, tool, data}`. CLI `tool` is `scan`|`check`; MCP is `list_deck`|`get_card`|`check_deck`.
- Deck = root Roster directs only. Tags: one enum. Lockfile pins resolved version only. Transitives, peers, bundled: out. Workspaces: root-only.
- npm Source, URL reserved as config, v1 hardcoded registry.npmjs.org. Concurrency 8 per Source around network only. Keep Card + `FetchError` array on failure.
- User cache dir, per-package files, TTL 24h/304-refresh/404-1h, no cache of 429/5xx/timeout. `--no-cache` skips read and write. CI check uses cache.
- Flavor v1: `rule`|`none`; `ai` errors. Template tier from `len(errors)>0`, not zero stars. `flavor_source` on the contract.
- HTML: one file next to package.json, html/template, static filter JS, <150KB target, no serve, no `--both`.
- Rarity from weekly downloads (ADR-0009). Check policy (ADR-0010). Tag and workspaces (ADR-0011).
- Test seam: run the pipeline against a fixture npm/cache (no live registry in tests). CLI and MCP are adapters over that seam.

## Testing Decisions

- Good tests assert observable contract: Dependency JSON, Check pass/flagged/policy, HTML file path and escaping, exit codes, MCP isError vs data errors. Do not assert private helper names.
- Pipeline tests with a fake Source and a temp cache dir cover fetch, degrade, Flavor tiers, Rarity cutoffs, Check policy.
- Parser tests: Tag first-wins, optional included, peer skipped, workspaces ignored, lockfile pin vs range.
- Cache tests: TTL classes, no slot on hit, `--no-cache`, ETag 304.
- CLI tests: `--json` skips HTML; default writes beside package.json; `--flavor ai` exit 2; check exit 0/1/3.
- MCP tests: envelope, get_card without path, schema enum reject ai, 404 is success with errors.
- No live npm in tests (`--no-cache` + fixture server).
- No prior test suite in-repo yet; this PRD is the first feature.

## Out of Scope

- Transitive / lockfile-as-deck, peer/bundled, `--workspaces`, GitHub Source, `ai` Flavor, `depdeck serve`, `--both`, `--refresh`, private registry auth, `render_html`/`search_deck`/`compare_decks` MCP tools, identicons/logos, full semver library (v1 latest-then-highest heuristic only).

## Further Notes

Glossary: `CONTEXT.md`. Decisions: `docs/adr/0001`–`0011`. Grill trail: `docs/rambles/2026-09-10-depdeck.md`. Issue tracker: this file. Seams: one — the pipeline result (Deck of Dependencies + Check). Adapters should stay thin enough that HTML/MCP tests are few.
