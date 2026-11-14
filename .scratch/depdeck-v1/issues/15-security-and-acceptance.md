Status: resolved

# 15 — Security and acceptance

**Parent:** [v1 completion plan](../COMPLETION_PLAN.md)

Complete v1 security controls and acceptance evidence after the product interfaces are implemented.

## Scope

- Review the threat model and link each mitigation to code, tests, or a release check.
- Add tests for path handling, bounded inputs and responses, redirects, cache corruption, output escaping, and MCP inputs.
- Run CI, tests, lint, dependency review, secret checks, and available vulnerability scans.
- Request the official Codex Security review if account access permits. If unavailable, record that status and complete the independent manual review and other checks.
- Fix release-blocking findings. Record exceptions with owner, impact, and follow-up.

## Acceptance criteria

- [x] Required checks pass, or a platform limitation explains each unavailable check.
- [x] No unresolved critical or high security findings remain.
- [x] Review evidence names tools, versions, and date. It does not claim a scan that did not run.
- [x] Threat-model mitigations link to tests or controls.
- [x] Manual review evidence and accepted residual risks are recorded.

## Completion evidence

Reviewed 2026-11-13 on local base `64eaac5bf5e4ccb3ac2696145460a71eec98e688` plus the working tree. The candidate includes the implementation of tickets 11 and 12. The following evidence applies to that local candidate.

- Registry selection skips absent and null `latest` records and falls back to the highest matching published stable release. Scan rejects output paths with the input manifest's file identity before assessment and before replacement. Real-binary tests cover relative and absolute paths and a hard-link alias, and verify the manifest bytes are preserved.
- Security regressions cover decoded cache bounds, future cache timestamps, hostile terminal names, malformed redirect diagnostics, control characters in licenses, invalid manifests, and MCP input and frame limits. The [threat model](../../../docs/THREAT_MODEL.md) links controls to their evidence. Browser Tag filters now hide the Cards they exclude.
- Windows amd64 passed `go test ./... -count=1` with Go 1.25.0 and 1.27.4. Go 1.27.4 also passed `go vet ./...`, `go mod verify`, and golangci-lint 2.14.0. The real-binary CLI/MCP suite uses an isolated seeded cache and asserts zero outbound attempts. The separate HTTP fixture process uses loopback only.
- Linux amd64 passed `go test -race ./... -count=1` for all 11 packages under Kali WSL with GCC and the SHA-256-verified Go 1.27.4 toolchain. Module downloads were disabled during the suite. Native Windows race testing was unavailable without MinGW GCC.
- Node 24.19.0 passed the hook-selection and Tag-filter tests. Headless Chrome 157.0.8103.52 opened a CLI-generated fixture Deck at 500 × 800 and 1400 × 900. It verified computed Card visibility for all Tags and All, the one- and six-column layouts, and escaping of a hostile package name. A renderer regression checks the 47-Card size target.
- govulncheck 1.7.0 initially reported standard-library findings with Go 1.27.0, plus Windows advisory GO-2026-5024 in the indirect `golang.org/x/sys` module. Updated the module to 0.44.0 without raising the earlier Go 1.25.0 test baseline. The normal gating scan with Go 1.27.4 exited 0 with no vulnerabilities found. Database `https://vuln.go.dev`, last updated 2026-11-11 02:51:11 UTC. Release builds must use a patched Go toolchain.
- Gitleaks 8.30.1, built from the checksum-verified Go module, initially flagged the same synthetic API-key example in two skill files. The local checkout replaced the current copies with a placeholder. The candidate preserves the removal of those tracked agent skills from upstream. The [history exceptions](../../../.gitleaksignore) name only the two exact historical fingerprints. Redacted history (44 commits from HEAD) and working-tree source reruns exited 0 with no leaks found.
- Independent reviewers separate from the implementation agent covered filesystem authority and input handling, npm/cache/HTML/MCP boundaries, dependency manifests, hooks, CI permissions, and MCP configuration. No confirmed High or Critical finding remained in that review. Official Codex Security local and Cloud scans were unavailable in this session. The independent source review, regression suite, and scanner evidence are the documented fallback.
- [CI](../../../.github/workflows/test.yml) now has Windows, Linux, and macOS jobs on the minimum and stable Go toolchains, Linux race tests, lint, and scanner gates. The action SHAs were checked against GitHub, permissions are read-only, and checkout credentials are not persisted. Actionlint 1.7.12 passed locally.

### Hosted acceptance

GitHub Actions passed all seven jobs on 2026-11-13 for candidate commit `f8b485e4f6cece17b29c680f636ebfe97f805d2f` on `codex/depdeck-v1-rc`. The branch uses the upstream Go 1.26.9 security minimum and tests it alongside stable. [Workflow run](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881)

| Gate | Successful jobs |
|---|---|
| Linux race suite | [Go 1.26.9](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881/job/131638470762), [stable](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881/job/131638470864) |
| Windows offline suite | [Go 1.26.9](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881/job/131638471144), [stable](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881/job/131638470884) |
| macOS offline suite | [Go 1.26.9](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881/job/131638471095), [stable](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881/job/131638471078) |
| Security | [Lint, Node checks, govulncheck, and redacted Gitleaks](https://github.com/pragatikumari0927/depdeck/actions/runs/43349998881/job/131638470859) |

### Residual risks

- Project/MCP host operator: OS permissions and mounts govern access to paths, including remote mounts. The static alias guard does not protect against an actor who can rename paths concurrently. Manifest declarations are preserved in output and must not contain credentials.
- Agent configuration owner: optional Cursor and Grok MCP configurations contain floating package versions. The shipped Go CLI and MCP server do not use them. Review the agent-side pins separately before relying on those configurations in a shared environment.