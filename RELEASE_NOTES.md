# Depdeck v1.1.0 candidate notes

These notes cover unreleased changes on `main` since the
[published v1.0.0 release](https://github.com/pragatikumari0927/depdeck/releases/tag/v1.0.0).
The candidate adds Pixel Archive, a self-contained HTML Deck with an offline
Rarity Atlas, dependency filters, Field Journals, and Archive Notices for
missing npm data. It also adds `render_deck`, an MCP tool that creates this
canonical Deck inside a project without overwriting an existing file.

Cards also carry package identity artwork independently of their Rarity seal.
The current catalog includes the reviewed Vite mark; other package names use
deterministic original pixel artwork. The generated Deck stays self-contained
and works offline. Release archives and the Homebrew installation include the
Vite attribution and MIT license in `THIRD_PARTY_NOTICES.txt`.

- CLI help explains the path and output options. Manifest errors and plain
  Check verdicts tell you what to do next. Scripts can use the same exit codes
  and schema 1.1 JSON, including the Check reason codes.
- The stdio MCP server provides four tools: `list_deck` and `check_deck` assess
  a project's direct dependencies, `get_card` looks up one npm package, and
  `render_deck` creates the canonical HTML Deck. The assessment tools explain
  publish-age and license-declaration policy results without treating them as
  vulnerability findings.
- Empty Checks now return an empty dependencies array in CLI JSON and MCP.
- New Unix cache directories use owner-only permissions. Existing directories
  and cache contents are left as they are.
- Offline Chromium browser regression checks cover filtering, layout, scaling,
  accessibility basics, escaping, and external requests. They do not claim
  Firefox, Safari, or screen-reader coverage.
- The README now uses the project logo and a current Go-rendered Pixel Archive
  sample. The documented sample data is illustrative; it is not a live npm scan.
- Release archives use Go 1.27.4. Preflight scans for known vulnerabilities
  with that compiler before packaging.
- Windows, macOS, and Linux archives support amd64 and arm64. Compare your
  archive's SHA-256 digest with the matching entry in checksums.txt before
  extracting it. The minimum for source builds is Go 1.26.11. Use the current
  security patch for your Go line.

See the [README](README.md) for usage and the [changelog](CHANGELOG.md) for
the full list of changes.
