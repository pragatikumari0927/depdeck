# CLI and MCP Contract

Source of truth: `SPEC.md` sections "CLI surface" and "MCP surface".

- Data to stdout. Progress, warnings, errors to stderr. Never mix.
- Exit codes: 0 success, 1 check-flag, 2 bad args, 3 total network failure.
- `NO_COLOR` env var disables all ANSI codes.
- Every data-producing command supports `--json`.
- `--flavor=ai` errors with code 2. Never silent fallback.
- MCP `isError: true` is reserved for tool failure, not data absence.
