# Contributing to depdeck

Depdeck turns the direct npm dependencies in a project's `package.json` into
an HTML Deck, CLI output, or results for an MCP client.

## Before you start

- For a bug report, include your OS, the output of `depdeck --version`, and a
  small example that reproduces the problem. Remove credentials and private
  package names before sharing it.
- Report suspected vulnerabilities privately through GitHub Security Advisories.
  See the [security reporting guide](docs/SECURITY.md).
- Use the [specification](docs/SPEC.md) to check the expected behavior. The
  [documentation map](docs/README.md) points to the architecture decisions, test
  guidance, and local agent setup.

## Build and test

Install Go 1.26.9 or later and run these commands from the repository root.

```sh
go mod download
go build ./cmd/depdeck
go test ./...
```

The first command downloads Go modules. Application tests use local fixtures
and loopback HTTP servers. Keep new tests independent of the public npm service.
Run the race suite where your Go toolchain and compiler support it.

```sh
go test -race ./...
```

For lint setup and platform notes, see the [development guide](docs/DEVELOPMENT.md)
and [test guide](docs/TESTING.md).

## Open a pull request

1. Start a branch from the latest `main`. Keep the change scoped to one issue
   or one logical improvement.
2. Use Conventional Commits, such as `docs: clarify the scan example` or
   `fix(registry): preserve verified download counts`.
3. Open a pull request against `main`. Explain the change, link its issue if
   one exists, and list the checks you ran. Include screenshots for visible
   Deck changes.
4. Wait for CI and maintainer review. A maintainer merges the pull request.

Keep local planning files and agent configuration out of commits. See the
[agent workflow](docs/agents/README.md) for local setup.
