# Threat model

## Scope

Depdeck accepts a local `package.json`, reads local cache data, calls the public npm registry, and writes JSON or self-contained HTML. The CLI and stdio MCP server share an assessment pipeline. Depdeck does not install packages or execute package scripts.

## Assets and boundaries

- User files and local cache data must remain private and intact.
- Registry data must not control local command execution or escape the output document.
- CLI and MCP output must represent incomplete data clearly.
- The network boundary is between the local process and the public npm registry. Registry responses are untrusted.
- The filesystem boundary includes the manifest, cache directory, and requested output path. Local files may be malformed or hostile.

## Threats and controls

ADR-0012 locks five threats and their mitigations. Keep each threat covered:

| Threat | Control | Evidence |
|---|---|---|
| Prompt injection through package descriptions | Ignore descriptions, strip control characters from license text, and report fixed transport diagnostics. | [Registry regressions](../pkg/registry/registry_test.go), [Flavor implementation](../pkg/flavor/rule.go) |
| Stored XSS in the HTML deck | Escape interpolated values with `html/template`. Use a static script and self-contained assets. | [Renderer tests](../pkg/render/render_test.go), [Deck template](../pkg/render/templates/deck.html) |
| Secret leakage through MCP configuration | Keep credentials out of configs. The product has no private-registry authentication. | [Read-only CI secret scan](../.github/workflows/test.yml), [security policy](SECURITY.md) |
| Cache poisoning | Hash per-package cache keys, enforce TTLs, validate cached bodies, and cache only successful or not-found responses. | [Cache tests](../pkg/cache/cache_test.go), [registry tests](../pkg/registry/registry_test.go) |
| MCP output used as agent instructions | Return typed data and fixed operational diagnostics. Keep package fields non-authoritative. | [MCP adapter tests](../internal/mcp/mcp_test.go), [stdio process tests](../cmd/depdeck/main_test.go) |
| Oversized or deeply nested manifest | Require a regular file and JSON object. Bound manifest reads to 4 MiB. | [Parser tests](../pkg/parser/parser_test.go), [process tests](../cmd/depdeck/main_test.go) |
| Oversized, malformed, or slow registry response | Bound bodies, timeouts, and concurrent requests. Validate metadata and counts before applying facts. | [Registry fixture tests](../pkg/registry/registry_test.go) |
| Redirect to an untrusted host | Restrict redirect origin and hops. Keep redirect URLs out of Fetch error details. | [Registry redirect tests](../pkg/registry/registry_test.go) |
| Cache file exhausts memory or supplies invalid data | Check file type, encoded size, decoded body size, entry structure, and freshness. | [Cache limit and corruption tests](../pkg/cache/cache_test.go) |
| Package data injects HTML or terminal control sequences | Escape template values and quote package names in CLI warnings. | [Renderer tests](../pkg/render/render_test.go), [CLI tests](../internal/cli/cli_test.go) |
| MCP peer sends an oversized frame or invalid arguments | Limit inbound frames to 16 MiB and validate tool arguments against explicit schemas. | [Stdio process tests](../cmd/depdeck/main_test.go) |
| Output path overwrites the manifest | Compare file identities before assessment and before atomically replacing the destination. | [CLI process tests](../cmd/depdeck/main_test.go) |

## Residual risks

The public registry can be unavailable or return incomplete data. Depdeck preserves Fetch errors in an assessment where the contract allows it. Cache hashes identify entries. They do not authenticate data.

The invoking user and MCP host govern filesystem access. OS permissions and mounts determine whether a path resolves to a local or remote filesystem. The file-identity checks cover existing aliases. They do not guard against hostile concurrent filesystem changes by an actor with write access. The declared version is preserved in output, so project owners should keep credentials out of manifest declarations.

Update this model when a new input, output, trust boundary, or network behavior is added. Link each new mitigation to a test or review check.