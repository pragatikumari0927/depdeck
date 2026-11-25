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

## v1.1.0 release closure

The v1.1.0 implementation and repository updates are merged to `main` but are
not tagged or published. Published v1.0.0 remains the latest release. As of
[#73](https://github.com/pragatikumari0927/depdeck/pull/73), which finalized
the closure documentation after
[#71](https://github.com/pragatikumari0927/depdeck/pull/71) and
[#72](https://github.com/pragatikumari0927/depdeck/pull/72), the
[post-merge test workflow](https://github.com/pragatikumari0927/depdeck/actions/runs/45532168280)
passed on that closure commit.

- Pixel Archive adds the Rarity Atlas, group filters, Field Journals, Archive
  Notices, deterministic package artwork, and the `render_deck` MCP tool.
- CLI and MCP wording, empty Check results, and manifest version selection
  have focused regressions. The assessment and JSON/MCP contracts remain stable.
- New Unix cache directories use owner-only permissions. Release archives
  include the Vite artwork attribution, MIT license, and trademark disclaimer.
- Browser regression coverage checks generated HTML offline in Chromium.
- Source builds require Go 1.26.11; release artifacts use Go 1.27.4, with
  vulnerability scans in release preflight. See
  [ADR-0023](adr/0023-go-patched-toolchains.md).

Before creating a v1.1.0 tag, run release preflight on the exact commit that
will receive the tag and retain the workflow URL in the release record.

The [changelog](../CHANGELOG.md) keeps the prepared v1.1.0 entries under
Unreleased, with a target publication date of 2026-11-25. If publication
slips, record the actual date when the version heading is added. The
[v1.1.0 release notes](../RELEASE_NOTES.md) describe the prepared release.
Follow the [release guide](RELEASING.md) for artifact preparation. Tagging,
publication, and the Homebrew tap update require separate approval.

## Deferred ideas

These ideas have no assigned release version.

- AI-generated Flavor. `rule` and `none` are supported. The `ai` flag still
  returns an error.
- Individual Card PNG export, an `install_size` stat, and a `depdeck doctor`
  command for local configuration diagnostics.
- GitHub enrichment, additional ecosystems, Yarn and pnpm lockfile support,
  `depdeck battle` for Deck comparison, workspace traversal, and graph analysis.

## Contributing

Use the [contributing guide](../CONTRIBUTING.md) to open a focused PR. Follow
the [agent workflow](agents/README.md) for worktree and local setup, and the
[testing guide](TESTING.md) for recording evidence.
