# Testing

Use a patched Go toolchain that meets the minimum in `go.mod`. Application
tests use local fixtures and may start loopback HTTP servers. Download missing
Go modules before running tests. Application tests do not depend on live npm.

## Checks

Run these commands from the repository root:

```sh
go test ./...
go vet ./...
go test -race ./...
```

Run the race check where the Go toolchain and compiler support it. Record any platform limit. Run the repository lint and security checks described in [DEVELOPMENT.md](DEVELOPMENT.md) and [SECURITY.md](SECURITY.md).

The [CI workflow](../.github/workflows/test.yml) tests the minimum and stable Go
versions on Linux, Windows, and macOS. The Linux jobs run race tests and the
stable Linux job records statement coverage. The security job runs the Node
tests, Go lint, vulnerability scanning, and secret scanning.
The `quality-gate` job depends on the complete test matrix, browser job, and
security job. It runs even when a dependency fails or is skipped, and succeeds
only when all three job results are `success`. Configure branch protection to
require `quality-gate` after confirming that check appears on a pull request.

## Acceptance coverage

- Parser: valid and invalid manifests, omitted and empty fields, unsupported declarations, and size limits.
- Cache: round trips, expiry, corruption, bounded reads, and write failures.
- Registry: semver selection compatible with common npm ranges, malformed data, timeouts, redirects, body limits, and fetch errors.
- Pipeline: all check outcomes and incomplete assessments, with a fixed clock and no network dependency.
- CLI: current-directory default, exit codes, stdout and stderr separation, JSON envelope, HTML output, and write failures.
- MCP: all four tools, schema major-version use, structured content, tool errors, a preserved assessment when npm fetch fails, and canonical Deck rendering with offline fixtures.
- HTML: hostile package names and versions do not create markup or script.

For each ticket, record commands and results in its completion notes. Do not claim a check passed if it did not run.

## Pixel Archive verification

This historical record covers the Pixel Archive feature branch on 2026-11-14
with Go 1.27.4, Node 24.19.0, and Chromium 157 on Windows. The race suite used
WSL Linux and GCC 16.2.0. The executable read a seeded cache based on the
six-card Pixel Archive v2 reference. The fixture values are sample npm facts.
Use a CI run for the commit under review to verify later changes.

| Check | Result |
| --- | --- |
| `go test -json ./... -count=1` | 11 packages and 162 tests and subtests passed, with zero skips. Includes renderer, CLI process, and MCP tests. |
| `go test -race ./... -count=1` | All 11 packages passed under WSL Linux. |
| `go vet ./...`, `golangci-lint run ./...`, `go mod verify`, `git diff --check` | Passed. golangci-lint 2.14.0 reported zero issues. |
| `node --test scripts/codex-go-checks.test.cjs scripts/deck-filter.test.cjs` | Four tests passed. |
| Built CLI and MCP stdio | CLI `scan` and `check` data matched MCP `list_deck` and `check_deck`. `get_card` preserved its standalone shape. The degraded Card retained its structured `not_found` error. |
| Browser, local HTML file | All five seals rendered. Atlas counts were 2 / 1 / 1 / 1 / 1 across all Tag filters. The Field Journal and Archive Notice matched the fixture facts. Hostile names, versions, and license strings stayed inert. The Deck made zero external requests. |
| Responsive layout | No horizontal overflow at 320, 390, 768, 1024, or 1440 px. Keyboard focus, reduced motion, forced colors, and the accessibility tree were checked in Chromium. |
| Determinism and size | Repeated CLI renders were byte-identical. The six-Card Deck was 31,685 bytes. The 47-Card CLI stress fixture was 96,372 bytes, and the repository size fixture was 107,101 bytes. Both 47-Card files were below 150,000 bytes. |

The CLI process suite also covers a successful package lookup followed by a
downloads 404 or 500 and a registry record dated in the future. It checks
the HTML messages and the original JSON error kinds and publish date.

## Pixel Archive browser regression

The `browser` job in the test workflow runs an offline Chromium regression on
pushes and pull requests. It is a separate status check and is not currently
configured as a required branch-protection check. The job uses Node 24.22.3,
Playwright 1.65.0 to install its matching Chromium build, and the Go 1.26.11
minimum toolchain. It adds one Ubuntu runner and downloads Chromium plus its
system libraries; the exact duration depends on GitHub-hosted runner cache and
package download time.

The test builds a temporary Deck by running
`scripts/browser-fixture/main.go` through the production Go HTML renderer. The
fixture is synthetic, includes the curated `vite` artwork, generated fallback
artwork, all five Rarity states, all three dependency groups, degraded Cards,
and hostile package text. It does not contact npm or use a project cache. The
test renders it twice and compares the output bytes
before opening the first Deck as a local file in Chromium.

The browser check covers:

- No HTTP or HTTPS browser requests.
- Runtime, Development, Optional, and All Cards filters, including hidden
  Card layout and unchanged Rarity Atlas totals.
- The five Atlas counts, semantic landmarks, labeled filter controls, and
  keyboard-visible focus.
- The curated Vite symbol and generated fallback both have visible geometry at
  Card size, and the Rarity seal remains separate.
- Hostile package names and versions as escaped text.
- 100% and 200% root text size at 320, 390, 768, 1024, and 1440 CSS pixels.

Run the check locally with Node 24 and a Chromium executable:

```sh
go run ./scripts/browser-fixture > /tmp/depdeck-browser-fixture.html
node scripts/verify-deck-reflow.cjs /path/to/chromium
```

The browser test generates its own disposable fixture files; the first command
is optional and lets a developer inspect the same fixture manually. Use
`node_modules/.bin/playwright` from the pinned version to install Chromium as
shown in `.github/workflows/test.yml`. On Windows, pass the path to a local
Chrome or Chromium executable.
GitHub's Ubuntu runner uses AppArmor to restrict the unprivileged user
namespaces that Chromium's sandbox needs. The workflow sets
`kernel.apparmor_restrict_unprivileged_userns=0` on the disposable runner before
the browser step, so CI runs Chromium with its sandbox on. If a local Linux
machine shows "No usable sandbox", use the same setting or an AppArmor profile
for the Chromium binary. Set `DEPDECK_CHROMIUM_NO_SANDBOX=1` only as a last
resort on a disposable machine.

This automation runs Chromium only. It does not claim Firefox, Safari,
screen-reader, or automated WCAG audit coverage. The existing Go renderer tests
remain responsible for the complete HTML contract and escaping regressions.

The [desktop screenshot](assets/pixel-archive-desktop.png) and
[mobile screenshot](assets/pixel-archive-mobile.png) are crops of the
CLI-generated Deck. The [current README demo](assets/deck-demo.gif) captures
the Go renderer at 1280 by 1150 CSS pixels. Its offline assessed fixture has
five sample Cards: `react` (Legendary, Runtime), `chalk` (Rare, Runtime),
`typescript` (Epic, Development), `tiny-helper` (Common, Optional), and
`unlisted-demo` (Unknown, Optional). The facts and dates are illustrative, not
live npm results. The capture shows All, Runtime, Development, and Optional
filters; Atlas totals remain one Card per rarity.

Chromium verification confirmed all five Atlas rows, visible Card counts of
5, 2, 1, and 2 for All, Runtime, Development, and Optional, unchanged Atlas
totals after filtering, no horizontal overflow at 1280 pixels, and no external
requests (the only request was the local `file:` document). The GIF is 800 by
719 pixels. This recorded browser run predates package artwork and its fixture
contains no known package such as `vite`; it does not verify package identity
art. Current `main` includes a curated Vite mark and a deterministic original
fallback, separate from the Rarity seal. Browser verification covered Chromium.
Firefox, Safari, and screen readers were outside this run.

<img src="assets/pixel-archive-mobile.png" alt="Mobile Deck with collection summary, Rarity Atlas, Tag filters, and the first Card" width="375">

## Fuzzing and assessment benchmark

Short fuzz runs exercise the public manifest parser, npm metadata parser, and
MCP output-path validator. Longer campaigns can increase `-fuzztime`:

```sh
go test ./pkg/parser -run '^$' -fuzz FuzzParseFile -fuzztime 30s
go test ./pkg/registry -run '^$' -fuzz FuzzParsePackument -fuzztime 30s
go test ./internal/mcp -run '^$' -fuzz FuzzCleanDeckOutput -fuzztime 30s
```

The 47-package shared-pipeline benchmark uses a loopback npm fixture. It reports
Go allocation totals (`B/op`), not process RSS, and does not include HTML
rendering or operating-system process startup:

```sh
go test ./internal/pipeline -run '^$' -bench BenchmarkPipelineColdWarmScan47 -benchmem -count=5 -benchtime=1x
```
