# Project phases

Use [SPEC.md](SPEC.md) for the current contract and [DECISIONS.md](DECISIONS.md)
for the accepted architecture decisions.

## Published v1.0.0

Depdeck v1.0.0 was published on 2026-11-13. It includes `scan`, `check`, and
the stdio MCP server for direct npm dependencies. It produces self-contained
HTML Decks with the original Card layout. The release provides six archives
and a checksum manifest. [GitHub release](https://github.com/pragatikumari0927/depdeck/releases/tag/v1.0.0)

The [archived completion plan](https://github.com/pragatikumari0927/depdeck/blob/510113e561b931b1d888b247026322b082d8c30d/.scratch/depdeck-v1/COMPLETION_PLAN.md)
and [roadmap at the v1.0.0 tag](https://github.com/pragatikumari0927/depdeck/blob/v1.0.0/docs/PHASES.md)
preserve the original ticket list and release criteria. The
[publication record](https://github.com/pragatikumari0927/depdeck/pull/23) covers archive verification and the
public Go and Homebrew install checks.

## Unreleased on main

- Pixel Archive adds Rarity seals, the Rarity Atlas, dependency filters,
  Field Journals, and Archive Notices. [PR #24](https://github.com/pragatikumari0927/depdeck/pull/24) · [PR #26](https://github.com/pragatikumari0927/depdeck/pull/26)
- CLI guidance and MCP tool descriptions are clearer. Empty Checks return
  `data.dependencies: []` in CLI JSON and MCP. [PR #39](https://github.com/pragatikumari0927/depdeck/pull/39) · [PR #42](https://github.com/pragatikumari0927/depdeck/pull/42) · [PR #38](https://github.com/pragatikumari0927/depdeck/pull/38)
- The minimum Go compiler for source builds is 1.26.11. The release archive
  preflight uses Go 1.27.4. [ADR-0023](adr/0023-go-patched-toolchains.md)

The [changelog](../CHANGELOG.md) and [candidate notes](../RELEASE_NOTES.md)
record the changes proposed for v1.1.0. The [candidate PR](https://github.com/pragatikumari0927/depdeck/pull/49)
records the hosted preflight evidence. Follow the [release guide](RELEASING.md)
for approval and publication.

## Deferred ideas

These ideas have no assigned release version.

- AI-generated Flavor. `rule` and `none` are supported. The `ai` flag still
  returns an error.
- Individual Card PNG export, an `install_size` stat, and a `depdeck doctor`
  command for local configuration diagnostics.
- GitHub enrichment, additional ecosystems, `depdeck battle` for Deck
  comparison, workspace traversal, and graph analysis.

## Contributing

Use the [contributing guide](../CONTRIBUTING.md) to open a focused PR. Follow
the [agent workflow](agents/README.md) for worktree and local setup, and the
[testing guide](TESTING.md) for recording evidence.
