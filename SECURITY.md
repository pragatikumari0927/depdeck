# Security

v1 threat model. Mitigations are locked by the cited ADR. Catalog decision: [docs/adr/0012-security-threat-model.md](docs/adr/0012-security-threat-model.md).

## Prompt injection via package descriptions

**Attack:** Registry description, README, or other package metadata is treated as instructions by an agent that reads MCP or Flavor output.

**Mitigation:** All registry and package text is untrusted. Strip control characters. Disable markdown links. Refuse executable-looking content. Sanitize before MCP output or HTML. v1 rule Flavor is name + stats templates, not raw descriptions.

**ADR:** [0005](docs/adr/0005-v1-flavor-rule-or-none.md), [0012](docs/adr/0012-security-threat-model.md)

## Stored XSS in the HTML deck

**Attack:** A package name, license, Flavor string, or other fetched field is written into `deck.html` and executes in the browser.

**Mitigation:** `html/template` only. Never `text/template`. Never `template.HTML` or `template.JS` on fetched strings. Filter script is a static template string.

**ADR:** [0006](docs/adr/0006-self-contained-deck-html.md), [0012](docs/adr/0012-security-threat-model.md)

## Secret leakage in MCP config

**Attack:** Tokens or credentials stored in MCP or app config files are logged, copied into envelopes, or committed.

**Mitigation:** Secrets come from environment variables only. Never from config files. v1 has no private-registry auth.

**ADR:** [0012](docs/adr/0012-security-threat-model.md)

## Cache poisoning

**Attack:** A write to the on-disk HTTP cache (or a swapped 200 body) is reused as npm metadata on the next scan or check.

**Mitigation:** Per-package `sha256` keys under the user cache dir. TTL 24h / 304 refresh / 404 1h. Do not cache 429, 5xx, timeouts, or network errors. `--no-cache` skips reads and writes. Treat cached bodies as untrusted registry text.

**ADR:** [0004](docs/adr/0004-on-disk-http-cache.md), [0012](docs/adr/0012-security-threat-model.md)

## MCP output as injection vector

**Attack:** An agent treats `list_deck` / `get_card` / `check_deck` fields (`flavor_text`, names, `detail`) as instructions rather than data.

**Mitigation:** MCP tool output is data, not authoritative. `isError` is tool failure only. npm 404 and fetch errors stay in `data` with `errors`.

**ADR:** [0007](docs/adr/0007-mcp-three-tools.md), [0012](docs/adr/0012-security-threat-model.md)

## Reporting

Report privately via GitHub Security Advisories on this repository. Do not open a public issue for an unreleased vulnerability.

Include: affected command or MCP tool, depdeck version or commit, reproduction steps, observed vs expected behavior, and impact (XSS, injection, secret exposure, cache integrity).
