The v1 minimum is Go 1.26.9 (see ADR-0020). Use a patched stable Go toolchain for release builds.
go install golang.org/x/tools/cmd/goimports@v0.50.0
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
These tool versions require a recent Go toolchain; the project language minimum
remains the version in go.mod. Put the Go binary directory (normally
`%USERPROFILE%\go\bin` on Windows) on PATH.
Post-edit hooks skip `goimports` and `golangci-lint` when missing,
print a one-time warning with the install command, and continue.
`-race` requires a 64-bit gcc. Verify with `gcc -dumpmachine` — it
must print `x86_64-w64-mingw32`. If it does not, run tests without
`-race` locally and rely on CI (`.github/workflows/test.yml` runs
it on ubuntu-latest).

## Required tools

Use `golangci-lint run` as the single local static-analysis pass. `.golangci.yml`
explicitly enables govet, staticcheck, errcheck, ineffassign, unused, and revive.
Do not run standalone vet/staticcheck again locally. CI's separate `go vet`
is retained as a dependency-free baseline alongside Linux race tests.

The Codex post-edit adapter is `scripts/codex-go-checks.cjs`. It handles
`PostToolUse` for `apply_patch`, `Edit`, and `Write`, formats only existing
Go files named by that edit, then lints and tests their packages. It does not
scan or format all dirty files. Shell-based file writes and deletion-only edits
require the agent to run the affected package checks explicitly.
Run `node scripts/codex-go-checks.test.cjs` to verify file selection without
formatting or testing the working tree.

- gofmt (ships with Go)
- CodeRabbit CLI — `coderabbit auth login`. Review before merge:
    `coderabbit review --committed --agent`. Config in `.coderabbit.yaml`.
- ripgrep — `rg` on PATH. Cursor's Shell tool resolves it to `%LOCALAPPDATA%\Programs\cursor\resources\app\node_modules\@vscode\ripgrep\bin\rg.exe` when Cursor is installed user-scoped (the default on Windows). Any system install (Chocolatey, WinGet) works as fallback.
