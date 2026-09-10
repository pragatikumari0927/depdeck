Go 1.22+ (see go.mod).
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
Post-edit hooks skip goimports and golangci-lint when those binaries are missing.
On Windows, `go test -race` needs a 64-bit gcc (not `C:\MinGW`). Race runs in CI.
