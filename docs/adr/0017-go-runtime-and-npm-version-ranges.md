# ADR-0017: Go runtime and npm version ranges

**Date:** 2026-11-02
**Status:** accepted

## Decision

Set the minimum Go version to `1.25.0`. Use the official `github.com/modelcontextprotocol/go-sdk` at v1.8.0 for MCP. Use `github.com/Masterminds/semver/v3` at v3.5.0 for version range checks.

Test common npm range forms against npm behavior. Preserve the declared range if no version matches. Do not silently treat an unsupported declaration as `latest`. Document unsupported forms and return a useful error or unresolved result.

## Reason

The official SDK provides the protocol implementation. A maintained semver library reduces custom range parsing risk. Pinning versions makes the planned v1 dependency set reviewable.

## Evidence and maintenance

Ticket 04 owns registry behavior and compatibility tests. Ticket 12 owns the module toolchain. Revisit if the SDK or required Go version changes.
