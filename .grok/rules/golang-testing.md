# Go Testing

## Framework

Use the standard `go test` with **table-driven tests**.

## Race Detection

Run tests with `-race` on Linux and macOS. On Windows, gate on
`gcc -dumpmachine` — see "Race detector on Windows" below.

```bash
go test -race ./...
```

## Coverage

```bash
go test -cover ./...
```

### Race detector on Windows

`-race` requires cgo and a 64-bit C compiler. Linux and macOS have it
by default. On Windows, run `gcc -dumpmachine`; if it does not print
`x86_64-w64-mingw32`, run tests without `-race` and rely on CI. CI
runs `go test -race ./...` on ubuntu-latest.

## Reference

See skill: `golang-testing` for detailed Go testing patterns and helpers.
