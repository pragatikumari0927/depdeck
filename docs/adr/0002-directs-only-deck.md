# Directs-only Deck from the Roster

The Deck is the project's Roster (`package.json` directs), not the lockfile tree. Transitive risk is real (event-stream, ua-parser-js, colors.js) but is a different product; competing with `npm audit` / Socket / Snyk would make depdeck a worse scanner and would explode HTML (47 cards vs ~800). The lockfile, when present, only supplies **Resolved version**; it never adds Cards.

**Status:** accepted

## Scope

- On the Deck, tagged: `dependencies`, `devDependencies`, `optionalDependencies`
- Off the Deck: `peerDependencies`, `bundledDependencies`, all transitives (v1 non-goal)

## v1 ticket batch

Lockfile parsing (Resolved version) is **deferred** with GitHub: first implementation issues ship `resolved_version` always omitted. The field stays on the contract.

## Shared contract

- `version` — **always** the declared Roster range/exact (`^18.2.0`). Never “whichever is available.”
- `resolved_version` — lockfile pin (`18.2.0`), omitted from JSON when empty

## v1 range fallback (no lockfile)

Prefer `dist-tags.latest`. Only if `latest` fails the declared range, pick the highest matching non-prerelease from the registry versions map. Defer a full semver-library resolver unless that heuristic fails in the wild.
