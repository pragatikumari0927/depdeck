# Security policy

Depdeck reads a local npm manifest and requests public npm registry data. It creates a local HTML file. It does not install packages or run package scripts.

## Secure coding

- Treat manifest names, versions, registry responses, cache files, and command arguments as untrusted input.
- Bound `package.json` input to 4 MiB, metadata response reads to 128 MiB, retained metadata to 32 MiB, downloads responses to 256 KiB, encoded metadata cache entries to 48 MiB, and encoded downloads cache entries to 384 KiB. Use request timeouts and a fixed concurrency limit.
- Use HTTPS for registry requests. Do not forward credentials or user headers to another host.
- Do not execute shell commands from package data. Do not install dependencies during a scan.
- Escape package data before writing HTML. Keep output paths explicit and report write failures.
- MCP `render_deck` writes only a new relative `.html` file under the selected project. It does not overwrite files or create directories; review the destination before calling the tool.
- Keep secrets and personal data out of logs, fixtures, issues, and commits.
- Keep dependencies pinned or bounded by the module lock data. Review dependency changes.

## Review and release gates

For release review, run the checks in [TESTING.md](TESTING.md), inspect dependency and secret scan results, and complete an independent security review. Request the official Codex Security review when account access permits it. If the service is unavailable, record that fact and use the documented fallback. Never report a review that did not run.

Only release supported, reviewed builds. See [RELEASING.md](RELEASING.md). Keep third-party notices with distributions.

## Reporting

Use the private reporting option on the repository's
[Security Advisories page](https://github.com/pragatikumari0927/depdeck/security/advisories)
when available. Otherwise, ask a maintainer for a private channel before
sharing details. Keep suspected vulnerabilities out of public issues.
Include the affected command or MCP tool, depdeck version or commit,
reproduction steps, observed and expected behavior, and impact.

Update this policy when controls, supported releases, or reporting channels change.
