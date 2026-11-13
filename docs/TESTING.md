# Testing

Use the Go version in `go.mod`. Run tests without network access. Tests must use local fixtures or an HTTP test server. Do not depend on live npm data.

## Checks

Run these commands from the repository root:

```sh
go test ./...
go vet ./...
go test -race ./...
```

Run the race check where the Go toolchain and compiler support it. Record any platform limit. Run the repository lint and security checks described in [DEVELOPMENT.md](DEVELOPMENT.md) and [SECURITY.md](SECURITY.md).

## Acceptance coverage

- Parser: valid and invalid manifests, omitted and empty fields, unsupported declarations, and size limits.
- Cache: round trips, expiry, corruption, bounded reads, and write failures.
- Registry: semver selection compatible with common npm ranges, malformed data, timeouts, redirects, body limits, and fetch errors.
- Pipeline: all check outcomes and incomplete assessments, with a fixed clock and no network dependency.
- CLI: current-directory default, exit codes, stdout and stderr separation, JSON envelope, HTML output, and write failures.
- MCP: all three tools, schema major-version use, structured content, tool errors, and a preserved assessment when npm fetch fails.
- HTML: hostile package names and versions do not create markup or script.

For each ticket, record commands and results in its completion notes. Do not claim a check passed if it did not run.
