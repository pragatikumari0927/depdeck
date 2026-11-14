# Depdeck 1.0.0

Depdeck scans the direct npm Dependencies declared in a root `package.json`.
It creates a self-contained HTML Deck, emits JSON for scripts, and offers a
policy Check for CI. The stdio MCP server exposes the same assessment through
`list_deck`, `get_card`, and `check_deck`.

- Cards include npm weekly downloads, Rarity, publish recency, license, and rule Flavor.
- Registry failures appear on degraded Cards so a partial assessment stays readable.
- Windows, macOS, and Linux archives target amd64 and arm64. Verify the archive
  against `checksums.txt` before extracting it.
- Source builds require Go 1.26.9 or later. The Homebrew formula builds from the tagged source.

See the [README](README.md) for installation and usage, and the
[changelog](CHANGELOG.md) for the development history.