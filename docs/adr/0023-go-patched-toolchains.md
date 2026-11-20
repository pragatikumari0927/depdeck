# ADR-0023: Use patched Go toolchains for source and release builds

**Date:** 2026-11-20
**Status:** accepted
**Supersedes:** [ADR-0020](0020-go-security-minimum.md).

## Decision

Set the minimum Go compiler to 1.26.11 and keep Go 1.26 language compatibility.
Test the minimum and current stable versions in CI. Use the current security
patch when building from source. On the Go 1.27 line, use 1.27.4 or a later
security patch.

Use Go 1.27.4 for the current release archives. Preflight verifies the exact
compiler, disables automatic toolchain switching, and runs govulncheck under
that compiler before packaging. Record the compiler embedded in the staged
Homebrew executable. Recheck the upstream Go security releases before the
next release.

## Rationale

The Go 1.26.9 source scan reached certificate verification used by depdeck's
HTTPS client. Go 1.26.10 fixes the affected certificate trust behavior.
Go 1.26.11 also includes the November fixes to the standard library and build
toolchain. Source scans with 1.26.11 and 1.27.4 reported no findings.
[Certificate advisory](https://pkg.go.dev/vuln/GO-2026-6809) ·
[Go release history](https://go.dev/doc/devel/release)

## Consequences

The minimum affects builds from current source. The MCP SDK and semver
versions from ADR-0017 continue to apply. CLI, MCP, JSON, HTML, and archive
formats retain their existing contracts.
