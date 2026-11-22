# Depdeck v1.1.0 candidate notes

These notes cover unreleased changes on `main` since the
[published v1.0.0 release](https://github.com/pragatikumari0927/depdeck/releases/tag/v1.0.0).
The candidate adds Pixel Archive, a self-contained HTML Deck with an offline
Rarity Atlas, dependency filters, Field Journals, and Archive Notices for
missing npm data.

- CLI help explains the path and output options. Manifest errors and plain
  Check verdicts tell you what to do next. Scripts can use the same exit codes
  and schema 1.1 JSON, including the Check reason codes.
- The three stdio MCP tool descriptions explain project assessments, standalone
  npm lookups, and the default publish-age and license-declaration policy.
- Empty Checks now return an empty dependencies array in CLI JSON and MCP.
- Release archives use Go 1.27.4. Preflight scans for known vulnerabilities
  with that compiler before packaging.
- Windows, macOS, and Linux archives support amd64 and arm64. Compare your
  archive's SHA-256 digest with the matching entry in checksums.txt before
  extracting it. The minimum for source builds is Go 1.26.11. Use the current
  security patch for your Go line.

See the [README](README.md) for usage and the [changelog](CHANGELOG.md) for
the full list of changes.
