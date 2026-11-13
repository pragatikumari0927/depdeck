# Release process

This guide defines v1 release readiness. Publishing is a separate final action. Do not publish without an explicit release instruction.

## Before a release

1. Confirm all v1 tickets are complete and merged. Review the working tree.
2. Run the checks in [TESTING.md](TESTING.md). Review dependency, secret, and vulnerability scan output.
3. Complete an independent security review. Run the official Codex Security review if account access permits. If unavailable, record the reason and the fallback evidence.
4. Update the version, changelog, and release notes. Check the MIT license and third-party notices.
5. Build and inspect Windows, macOS, and Linux archives. Verify checksums and a clean-machine install smoke test.
6. Validate the source-built Homebrew formula against the tagged source. Keep binary archives on GitHub.
7. Confirm that the CLI and stdio MCP server start, report the build version, and return documented errors.

## Credentials and publication

Use repository-host release permissions only. Never commit tokens, signing keys, or registry credentials. Keep publishing as the final separate action after the release artifacts and notes are ready for review.

## Rollback

If an artifact is wrong, stop further distribution. Mark the release as withdrawn, publish corrected artifacts under a new version, and document the change. Do not replace an artifact under the same version.

Update this guide when packaging or release controls change.
