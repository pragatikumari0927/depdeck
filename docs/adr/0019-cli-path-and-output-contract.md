# ADR-0019: CLI path and output contract

**Date:** 2026-11-02
**Status:** accepted

## Decision

`scan` and `check` use the current directory when no path is given. They read `package.json` from that directory. A missing or invalid manifest returns exit code 2. Use the shared JSON envelope for `--json`; keep machine-readable output on stdout and diagnostics on stderr. HTML output uses the requested path and reports write failures.

## Reason

The current directory is the least surprising default for a local project tool. Stable output streams and exit codes support scripts and automation.

## Evidence and maintenance

Ticket 08 owns command behavior and process tests. Ticket 12 owns executable wiring. Update this decision if CLI defaults or output compatibility changes.
