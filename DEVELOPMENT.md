Go 1.22+ (see go.mod).
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
Post-edit hooks skip `goimports` and `golangci-lint` when missing,
print a one-time warning with the install command, and continue.
`-race` requires a 64-bit gcc. Verify with `gcc -dumpmachine` — it
must print `x86_64-w64-mingw32`. If it does not, run tests without
`-race` locally and rely on CI (`.github/workflows/test.yml` runs
it on ubuntu-latest).

## Required tools

- gofmt (ships with Go)
- CodeRabbit CLI — `coderabbit auth login`. Review before merge:
    `coderabbit review --committed --agent`. Config in `.coderabbit.yaml`.
- ripgrep — `rg` on PATH. Cursor's Shell tool resolves it to `%LOCALAPPDATA%\Programs\cursor\resources\app\node_modules\@vscode\ripgrep\bin\rg.exe` when Cursor is installed user-scoped (the default on Windows). Any system install (Chocolatey, WinGet) works as fallback.
