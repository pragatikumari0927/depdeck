# ADR-0027: Read direct resolved versions from npm lockfiles

**Date:** 2026-11-25
**Status:** Accepted
**Supersedes:** the deferred lockfile-resolution implementation in [ADR-0002](0002-directs-only-deck.md)

## Context

The shared Dependency contract already has the optional `resolved_version`
field. The root `package.json` remains the project's Roster and the only source
of Card membership and declared versions. A root npm lockfile can add useful
installation context without changing the direct-only Deck contract.

## Decision

- Read the adjacent root `package-lock.json` and support npm lockfile formats 1,
  2, and 3.
- Populate `resolved_version` only for a direct dependency already present in
  the root manifest. For v1 use the root `dependencies` entries. For v2/v3 use
  the exact `packages["node_modules/<name>"]` entry. Do not use nested entries,
  infer a version from registry facts, or add Cards from the lockfile.
- Keep `version` as the declared range or exact version. A lockfile entry never
  changes it.
- Limit lockfile input to 64 MiB. Missing, malformed, unsupported, unmatched,
  or unreadable lockfiles leave `resolved_version` empty and do not prevent a
  valid manifest from being read.
- Do not read Yarn or pnpm lockfiles in this change. Do not change JSON schema,
  CLI, MCP, assessment, or registry behavior.

## Consequences

The parser exposes a lockfile pin when npm's supported lock format provides an
unambiguous root entry. This value is context only; it is not the version used
for npm fact selection, which continues to follow the declared range and the
existing registry selection rules. Lockfile format support can be added later
only with focused fixtures and a successor decision if it changes this scope.

## Verification

Offline tests cover v1, v2, and v3 lockfiles, scoped packages, root-only matching,
preservation of declared versions, malformed and unsupported lockfiles,
unmatched entries, and invalid direct entries. The shared assessment pipeline
continues to derive Cards from the manifest Roster only.
