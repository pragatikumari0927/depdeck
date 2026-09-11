---
name: verify-depdeck
description: Prove depdeck.exe CLI behavior (scan, check, deck.html, --json, exit codes), not go test. Reach for this after CLI or pipeline changes, or before claiming scan, check, HTML, or JSON output works.
---

# Verify depdeck

Drive the **binary** the way a user does. `go test ./pkg/...` is not this skill.

`cmd/depdeck` is the user path. If that package is missing, doctor fails — do not substitute library tests.

MCP (`depdeck mcp`) exists in `SPEC.md`. This map does not prove it. Do not invent a ninth gate.

## When to invoke

- After a change to `cmd/depdeck`, `internal/cli`, `internal/pipeline`, or render/JSON output
- Before claiming `scan`, `check`, `deck.html`, `--json`, `--flavor`, or exit codes work
- When a PR needs live proof, not unit coverage

## Launch

From the repo root:

```powershell
go build -o depdeck.exe ./cmd/depdeck
```

**Ready:** `depdeck.exe` exists and running it without a valid command prints usage to stderr (or a mapped command runs). There is no server to keep alive.

**Teardown:** delete the run temp dir and the `depdeck.exe` this run built. Do not delete the script's stdout transcript.

## Doctor

Read-only. Run first if anything looks off.

1. `cmd/depdeck` exists (at least `cmd/depdeck/main.go`). If absent, **fail** with that path — do not invent a `go test` stand-in.
2. `go build -o depdeck.exe ./cmd/depdeck` exits 0.
3. Fixtures exist: `testdata/verify/package.json` (clean), `testdata/verify/flagged/package.json`, `testdata/verify/empty/` (no manifest).

`rg` is not required.

## Drive

```powershell
pwsh -NoProfile -File .cursor/skills/verify-depdeck/run.ps1 testdata/verify
```

Arg 1 is a `package.json` or its directory. The script copies the fixture into `$env:TEMP\depdeck-verify-<pid>` so `deck.html` never lands in the repo.

## Nine-assertion feature map

Nine assertions, in order. Each: user action, exact command, observable end state. `run.ps1` prints `PASS` or `FAIL` with the command and observed exit/streams. `FAIL` includes **raw stderr**, not a paraphrase.

### 1. scan writes deck.html next to the package.json

User scans a project. Cards appear as `deck.html` beside that manifest, not in CWD.

```text
.\depdeck.exe scan <isolated-fixture-dir>
```

**Proof:** `<isolated-fixture-dir>\deck.html` exists after exit 0.

### 2. deck.html is byte-identical across two scans of the same fixture

Same inputs, two scans, same bytes. HTML must not contain `generated_at`.

```text
.\depdeck.exe scan <isolated-fixture-dir>
.\depdeck.exe scan <isolated-fixture-dir>
```

**Proof:** SHA256 of the two `deck.html` files match.

### 3. scan --json writes a valid envelope to stdout

```text
.\depdeck.exe scan --json <isolated-fixture-dir>
```

**Proof:** stdout parses as JSON with keys `schema_version`, `generated_by`, `generated_at`, `tool`, `data`. Consumers check the major of `schema_version` only (prefix `1.`).

### 4. scan --json does not write deck.html

```text
.\depdeck.exe scan --json <fresh-isolated-dir>
```

**Proof:** no `deck.html` next to that `package.json` after the command.

### 5a. check testdata/verify/ exits 0 (no deps flagged)

```text
.\depdeck.exe check testdata/verify/
```

**Proof:** exit `0` exactly. Any other value is FAIL. If exit `3`, dump raw stderr and note that SPEC maps all-fetch-failed to 3 — cannot prove exit 0 without network.

### 5b. check testdata/verify/flagged/ exits 1 (left-pad flagged)

```text
.\depdeck.exe check testdata/verify/flagged/
```

**Proof:** exit `1` exactly (`left-pad` chaos 1.0 > 0.5; `lodash` is the healthy control). Any other value is FAIL. If exit `3`, dump raw stderr and note that SPEC maps all-fetch-failed to 3 — cannot prove exit 1 without network.

### 6. --flavor=ai exits 2 with ai and v1.1

```text
.\depdeck.exe scan --flavor=ai <isolated-fixture-dir>
```

**Proof:** exit `2`. stderr contains `ai` and `v1.1`. No `deck.html`. Planned message: `--flavor=ai is planned for v1.1; v1 supports rule and none`.

### 7. missing manifest (empty dir) exits 2

```text
.\depdeck.exe scan testdata/verify/empty/
```

**Proof:** exit `2` (usage/parse). The directory exists and has no `package.json`. Not exit 3 — that code is `check` when every dependency has fetch errors (`SPEC.md`).

### 8. NO_COLOR=1 produces no ANSI in stderr

```text
$env:NO_COLOR = '1'; .\depdeck.exe scan <isolated-fixture-dir>
```

**Proof:** stderr contains no ESC `[` CSI sequences (`\x1b[`).

## Evidence

Proof is the script's **stdout**. Each step line names:

- the exact command
- observed exit code
- stdout/stderr (raw stderr on FAIL)

Do not treat a unit-test pass as a substitute. Cleanup must not delete this transcript.

## Cleanup

Remove only what this run created: `$env:TEMP\depdeck-verify-<pid>`, any `deck.html` inside it, and the `depdeck.exe` the script built. Never kill by process name. Leave the stdout transcript.

## Helpers

```powershell
pwsh -NoProfile -File .cursor/skills/verify-depdeck/run.ps1 testdata/verify
```

Optional path: a `package.json` file or a directory that contains one.

Keep the map honest as the CLI changes: `/maintain-verification-skill`.
