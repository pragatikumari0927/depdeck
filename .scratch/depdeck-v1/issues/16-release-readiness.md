Status: ready-for-agent

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

- [ ] Builds succeed for documented targets and archives contain required files.
- [ ] Checksums verify and clean-machine smoke tests pass.
- [ ] Homebrew builds from source and installs the CLI.
- [ ] License, third-party notices, changelog, release notes, and demo are ready.
- [ ] Ticket 15 evidence is complete and no release-blocking finding remains.
- [ ] This ticket does not publish artifacts. Publishing needs a separate final instruction.
