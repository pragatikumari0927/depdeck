Status: resolved

# 16 — Release readiness

**Parent:** [v1 completion plan](../COMPLETION_PLAN.md)

Prepare a release-ready v1. Publishing remains a separate final action.

## Scope

- Add the MIT license and preserve third-party notices.
- Provide reproducible GitHub binary archives for supported Windows, macOS, and Linux targets, with checksums.
- Add a source-built Homebrew formula and validate it against release source.
- Verify clean-machine install and CLI/MCP startup behavior.
- Prepare version, changelog, release notes, demo, and release preflight.
- Complete ticket 15 before release readiness is marked complete.

## Acceptance criteria

- [x] Builds succeed for documented targets and archives contain required files.
- [x] Checksums verify and clean-machine smoke tests pass.
- [x] Homebrew builds from source and installs the CLI.
- [x] License, third-party notices, changelog, release notes, and demo are ready.
- [x] Ticket 15 security and acceptance evidence is complete.
- [x] Confirm access to the production Homebrew tap before authorizing release.
- [x] This ticket does not publish artifacts. Publishing needs a separate final instruction.

## Preflight evidence

Verified 2026-11-13 at candidate `a1147d8285a28610bc2d2bcd578afd9f5576d673` on `codex/depdeck-v1-rc`.

- [Product and security CI](https://github.com/pragatikumari0927/depdeck/actions/runs/43351609987) passed all seven jobs, including Linux race tests, both Windows and macOS toolchain jobs, lint, govulncheck, and Gitleaks. The [ticket 15 evidence](15-security-and-acceptance.md) records the independent review and offline fixture coverage.
- [Archive preflight](https://github.com/pragatikumari0927/depdeck/actions/runs/43351610051/job/131643742975) passed. Official GoReleaser OSS 2.19.0 binaries were verified against upstream SHA-256. Windows and Linux snapshot builds used Go 1.27.4, `--snapshot --skip=publish`, and an offline Go module cache during compilation. All six archive SHA-256 values matched between the two hosts. Verified archive members, embedded target metadata, and the extracted Windows amd64 and Linux amd64 executables. Artifacts are in the local candidate worktree's ignored `dist/`.
- [macOS Homebrew preflight](https://github.com/pragatikumari0927/depdeck/actions/runs/43351610051/job/131643743252) passed. A disposable tap installed the generated formula from a loopback-served `git archive` of the candidate and ran `brew test`. The installed CLI and stdio MCP server reported version `1.0.0`, returned the expected empty assessment, and shut down cleanly.
- Isolated `go install ./cmd/depdeck` passed on Windows and hosted Linux. Smoke tests checked build identity, CLI error status, schema 1.1 Check data, MCP initialization, all three tools, an assessment, and EOF shutdown.
- git-cliff 2.14.2 generated the committed Unreleased changelog and an untagged `1.0.0` preview from Conventional Commits. The MIT license and notices are in every archive. Notices include the Go runtime and all ten modules linked across the supported platforms.
- The [README demo](../../../docs/assets/deck-demo.gif) was captured in headless Chrome from six CLI-rendered Cards with seeded fixture metadata and a cached 404. ImageMagick verified three distinct 940 × 489 GIF frames showing All, development, and optional Tags. Local README, release notes, and runbook links resolve.

### Snapshot SHA-256

```text
01072cd9af6687cff8f1c95116c67e0f4087350f2316988b64f43feee4898440  depdeck_1.0.0-rc-a1147d8_darwin_amd64.tar.gz
7e4b087936920934b20708e7e539cc9303fa612182826e527e617fe64ec49cf1  depdeck_1.0.0-rc-a1147d8_darwin_arm64.tar.gz
f9aa1c8d0d024f478c77f8b228d4423617b4552e9dcadae63e177961c8115a6c  depdeck_1.0.0-rc-a1147d8_linux_amd64.tar.gz
cc6c401350cb8b54811218d29173cfbc9a7868d779de0a18cfe8dd354729b076  depdeck_1.0.0-rc-a1147d8_linux_arm64.tar.gz
c138cd5699219b83d59aa1619b86db166f3b3696b8fc2497a7799270491a0e36  depdeck_1.0.0-rc-a1147d8_windows_amd64.zip
0eaafe226a71244622801f8b7e31836b3c61bf6bd989fa69892b666f318c3a18  depdeck_1.0.0-rc-a1147d8_windows_arm64.zip
```

### Publication preflight

- On 2026-11-13, created the public [Homebrew tap](https://github.com/pragatikumari0927/homebrew-tap) with default branch `main`. GitHub reported `permissions.push: true` for the authenticated owner account, and a [README-only commit](https://github.com/pragatikumari0927/homebrew-tap/commit/b4fbfeed9ea03937fd74c3dd75973a44327ac236) verified an actual write.
- Verified that the tap tree contained only `README.md` and had zero tags. Depdeck had no `v1.0.0` tag or GitHub release. The staged formula remains in the candidate workflow.
- The release operator must use a separately scoped tap credential when publication is authorized. The public Homebrew and versioned Go installation checks follow publication. The candidate remains on its own branch.
