# ADR-0026: Offline package artwork

**Status:** Accepted
**Date:** 2026-11-23
**Supersedes:** the no-per-card-artwork clause in [ADR-0006](0006-self-contained-deck-html.md)

## Context

Pixel Archive Cards currently reuse the rarity emblem as their main illustration.
This makes unrelated packages look alike. ADR-0006 keeps the HTML Deck
self-contained, deterministic, escaped, and under 150,000 uncompressed bytes
for 47 Cards. It also excludes per-card logos. This decision makes a narrow,
reviewed exception to that exclusion without changing the other renderer
contracts.

Package names and npm metadata are untrusted input. They must never select
artwork by substring, supply SVG markup, or enter executable HTML, CSS, or
JavaScript. Artwork is decorative; the visible package name remains the
accessible identity of each Card. Rarity remains an independent assessment
field and keeps its existing seal.

## Decision

- Use a small catalog keyed by exact, canonical npm package identity. Begin with
  the `vite` package only. Do not infer identity from keywords, prefixes, or
  similar names.
- Bundle reviewed SVG source locally. Inline it through a reusable symbol
  definition and reference that symbol from Cards. Do not request images,
  fonts, or other assets at runtime.
- Record the upstream source, source revision or blob identifier, license, and
  trademark notice beside the artwork. Vite's repository is MIT-licensed, but
  its name and logo remain their owners' marks; use the mark only to identify the
  exact package and do not imply endorsement.
- For all other names, draw a compact original pixel motif from a stable hash
  of the complete package name. The fallback is decorative and does not encode
  package semantics. Scoped names are treated as complete identities.
- Generate only controlled SVG paths, symbol IDs, and attributes. Keep names in
  ordinary `html/template` text contexts. Never cast fetched or package data to
  trusted HTML, CSS, JavaScript, or SVG types.
- Deduplicate symbols per rendered Deck. Identical assessed inputs must produce
  byte-identical HTML with no time-dependent or random values.
- Keep the HTML under ADR-0006's 150,000-byte limit for 47 Cards. If the
  catalog or fallback exceeds this limit, reduce or reject artwork additions;
  do not relax the limit in this ADR.

## Consequences

- The renderer has a small catalog and a fallback-art generator. Neither
  affects Rarity, Chaos, statistics, policy, JSON, CLI, or MCP output.
- New catalog entries need exact-name tests, provenance and rights review, and
  a measured size check. Artwork with unclear permission or identity remains
  excluded.
- Logos may be altered only as required for safe, compact, self-contained SVG
  embedding; provenance must describe such changes.
- SVG symbols are decorative and hidden from assistive technology. Visible
  package names and separate Rarity labels and seals remain available.

## Alternatives considered

- Keep using only rarity emblems: rejected because it does not distinguish
  package identity visually.
- Fetch package logos at render or browser time: rejected because it breaks
  offline use, determinism, and the self-contained Deck contract.
- Match by substring or package keywords: rejected because it mislabels
  unrelated and attacker-controlled package names.
- Generate only a deterministic fallback: safe and compact, but misses a
  recognizable mark for the one explicitly curated package.
- Add a large third-party logo catalog: deferred until each asset has a clear
  license and provenance record and the size budget is measured.

## Verification

The implementation must test exact catalog matching, fallback rendering,
scoped and similar names, malicious names, deterministic bytes, no remote
assets, unchanged Rarity, and the 47-Card size limit. Browser evidence should
show both the curated symbol and generated fallback at Card scale.
