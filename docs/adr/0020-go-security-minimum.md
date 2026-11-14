# ADR-0020: Keep the upstream Go security minimum

**Date:** 2026-11-13
**Status:** accepted
**Supersedes:** ADR-0017's Go minimum only.

## Decision

Keep Go 1.26.9 as the project minimum, as established by the upstream standard-library security update in commit `91ae97c`. The release candidate must not lower that minimum when it adds the CLI and MCP Faces. Test the minimum and current stable toolchains in CI. Build releases with a patched stable toolchain and run govulncheck against that toolchain.

The SDK and semver pins from ADR-0017 still apply.
