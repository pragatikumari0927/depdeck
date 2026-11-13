# Depdeck v1 completion plan

Date: 2026-11-02 (Asia/Calcutta)

Status: ready for implementation review. This file is a plan. It does not mark tickets complete or authorize publishing.

Repository: `C:\Users\Aryan Singh\dev\depdeck`.
Paths below are relative to this repository. Existing tickets remain in `.scratch/depdeck-v1/issues/`.

## 1. Outcome and settled decisions

Deliver a release-ready v1 with a working CLI, a stdio MCP server, one shared assessment pipeline, current documentation, security evidence, and tested release packages. Keep publishing as a separate final action.

The user settled these choices during the planning interview:

| Decision | Agreed result |
|---|---|
| Documentation | Move the full project guidance into `docs`. Keep short root entry files and compatibility pointers. |
| Agent discovery | Keep tool configs, hooks, rules, and installed skills at their discovery paths. |
| MCP | Use `github.com/modelcontextprotocol/go-sdk v1.8.0`; raise the Go minimum to `1.25.0`. |
| Version ranges | Use `github.com/Masterminds/semver/v3 v3.5.0` with npm compatibility tests. |
| Project license | MIT. Preserve third-party licenses and notices. |
| Omitted CLI path | `scan` and `check` use the current directory. A missing or invalid manifest returns exit 2. |
| Check data | Add the assessed Dependencies. Preserve `pass`, `flagged`, `reason`, and `policy`. |
| Standalone Cards | `get_card` omits the Roster-only `version` and `tag` fields. Do not invent project membership. |
| Security review | Include repository controls and an official Codex Security review stage. |
| Service unavailable | Allow security tests, CI checks, and an independent manual review as a documented fallback. Never claim an official scan ran. |
| Homebrew | Use a source-built formula. Also provide prebuilt archives on GitHub. |

Keep the v1 scope: npm, root Roster, direct Dependencies, rule/none Flavor, static HTML, and three MCP tools. Keep lockfile resolution, workspace walking, GitHub enrichment, AI Flavor, a web server, and Deck comparison out of this release.

The current baseline has seven Go packages and passing package tests. It has no pipeline, CLI, MCP adapter, or executable entry point. The npm client is currently untracked. Existing cache and agent-tool changes belong to the current working tree; preserve them and review their diff before including them in a commit.

## 2. Ticket order and document migration

Use the existing ticket numbers. Repair stale acceptance criteria before implementing against them. Do not rebuild a package only because its ticket still says `ready-for-agent`.

| Order | Tickets | Required result |
|---|---|---|
| A | New 14: documentation and contract alignment | Move guidance, add the missing documents, record successor ADRs, and repair ticket interfaces. |
| B | 01, 02, 03, 05, 09, 10 | Recheck the existing types, cache, parser, stats, renderer, and template against the amended contract. Preserve working behavior. |
| C | 13 and 04 | Finish canonical fixtures, then repair and complete the npm client and version selection. |
| D | 06 | Add Flavor mode handling. Preserve the deterministic rule engine. |
| E | 07 | Implement and test the shared pipeline. |
| F | 08 and 11 | Implement the CLI and MCP adapters against the same pipeline. |
| G | 12 | Add executable wiring and process tests. |
| H | New 15: security and acceptance | Complete security controls, end-to-end tests, CI, and review evidence. |
| I | New 16: release readiness | Complete install support, release archives, Homebrew formula, demo, changelog, and release preflight. |

Create the three additional tickets only for the added scope. Put references to this plan in them. Keep the original thirteen tickets as the owners of their product work. Use one logical change per commit and the repository's small-PR convention. A ticket is resolved only when its criteria and evidence are recorded. Update the merged-ticket count only after the corresponding work is merged.

### Canonical document layout

Move `SPEC.md`, `CONTEXT.md`, `DECISIONS.md`, `DESIGN.md`, `DEVELOPMENT.md`, `PHASES.md`, `RULES.md`, and `SECURITY.md` to the same filenames under `docs/`.

Keep a short root pointer for each moved file. Each pointer identifies the canonical document and tells agents to read it. Keep the root `AGENTS.md` as the discoverable entry file. Its required startup instructions must appear before its links. It points to the authority list and the shared agent workflow. Keep the root `SECURITY.md` reporting pointer so existing ADR links and repository tooling still resolve.

Add these documents. Give each document one clear purpose:

| Document | Content and maintenance trigger |
|---|---|
| `docs/README.md` | Reading order, authority order, links, and document ownership. Update when the document map changes. |
| `docs/SECURITY.md` | Secure coding rules, supported release policy, reporting, review gates, secret handling, and scanner exceptions. Update when controls or reporting change. |
| `docs/THREAT_MODEL.md` | Assets, entry points, trust boundaries, the five existing threats, added resource/input threats, mitigations, residual risks, and linked tests. Update with boundary changes. |
| `docs/TESTING.md` | Setup, test commands, fixture rules, CLI/MCP acceptance, race-detector limits, and evidence requirements. Update with behavior or verification changes. |
| `docs/RELEASING.md` | Versioning, install checks, archive and formula preparation, credentials, preflight, publication, and rollback. Update with release workflow changes. |
| `docs/agents/README.md` | Shared workflow: authority lookup, ticket use, scope, checks, completion evidence, and handoff. |
| `docs/agents/CODEX.md` | Codex discovery, scoped hooks, Windows constraints, repository permissions, and official security-review setup. |
| `docs/agents/REVIEW.md` | Correctness and security review checklist, severity, evidence, exceptions, and merge criteria. Link to the policies rather than copying them. |
| `docs/agents/HANDOFF.md` | Short handoff template: ticket, commit/branch, changes, tests, open risks, and next action. Store actual handoffs in the existing handoff directory. |

Do not add a second architecture or API specification. Keep architecture and interfaces in `docs/SPEC.md`; keep domain definitions alone in `docs/CONTEXT.md`. Keep tickets in `.scratch`, not in `docs`.

Keep `.cursor/rules`, `.grok/rules`, `.codex/hooks.json`, tool configuration, hook scripts, and skill bundles in place. Move narrative guidance only. Cursor rule files remain the rule source. Regenerate Grok copies with the existing sync script.

Update active references in agent instructions, Cursor rules, CodeRabbit configuration, review commands, scripts, tickets, and test instructions. Resolve links relative to the document that contains them. Preserve accepted ADR text and historical handoffs. Root pointers preserve their old links. Check pointer targets and require exactly one canonical copy of each policy.

### Record contract changes first

Use successor ADRs in the next available numbers; do not edit accepted ADRs. Record four grouped decisions:

1. Documentation location and the security-policy/threat-model split. Preserve the five threats from ADR-0012.
2. Go 1.25, the two pinned runtime dependencies, npm range semantics, and unsupported declarations. This replaces the no-semver-library restriction in ADR-0002.
3. Check data, incomplete-assessment behavior, standalone Card shape, and schema version `1.1`. Preserve project Dependency JSON fields and tags.
4. Omitted CLI paths and the clarified CLI output/exit rules below.

Update the decision index, canonical specification, roadmap, PRD, and affected tickets after these decisions are recorded. In particular, align tickets 03/07 on file-path parsing, 06 on the Flavor API, 09 on the existing writer-based renderer, 10 on the embedded template path, 11 on tool versus protocol errors, and 12 on omitted paths. Remove ticket 13's current dependency on ticket 04. Prepare fixtures first, then make ticket 04 depend on that fixture work. This removes the tracker/order contradiction.

## 3. Complete the assessment core

### Shared types and public interfaces

Preserve the existing `Dependency` fields and their JSON tags. Add `Dependencies []Dependency` with `json:"dependencies,omitempty"` to `CheckData`. Keep `flagged` as an empty array when nothing is flagged. Use schema version `1.1` on both Faces; consumers still check major version 1. Record 1.1 as the first supported shipped schema baseline. The absent MCP server has never shipped its documented 1.0 standalone shape. Correcting that unreleased shape is not a claim that removing required fields is a minor-compatible change. Any future removal from a shipped shape requires a major schema change.

Add a standalone Card response type in `pkg/types`. It contains the existing enrichment fields and Fetch errors, but no `version`, `tag`, or `resolved_version`. Convert from the assessed Dependency in one place. Test that shared fact fields are identical. This avoids changing project Dependency JSON tags, including the case of an empty declared version.

Keep one envelope format: schema version, UTC RFC3339 generation time, build identity, Face-specific tool name, and data. Pass the build version into the adapters. Use `depdeck dev` for an unversioned development build.

Use these core entry points:

```go
flavor.Apply(mode flavor.Mode, dep types.Dependency) (types.Dependency, error)
registry.Client.Fetch(ctx context.Context, name, declared string) (*types.Dependency, error)
pipeline.Run(ctx context.Context, path string, opt pipeline.Options) ([]types.Dependency, error)
pipeline.Check(ctx context.Context, path string, opt pipeline.Options) (types.CheckData, error)
pipeline.GetCard(ctx context.Context, name string, opt pipeline.Options) (types.Dependency, error)
```

Pipeline options carry Flavor, cache, an injected registry interface, and an injected clock. The registry interface has the Fetch method above. Production supplies real dependencies; tests supply fixture HTTP, temporary cache storage, and a fixed clock. Normalize an empty Flavor option to `rule`.

### Tickets 04 and 13: npm, version selection, and cache

Consolidate npm fixtures under the existing ticket's `testdata/npm/` target. Move the useful current registry fixtures there and update all readers. Add licensed, unlicensed, stale, zero-download, range-fallback, malformed, missing, rate-limit, server-error, timeout, and failed-download cases. Document that fixture dates and counts are test data. Use local HTTP handlers for response status and delay.

Cache raw metadata by package name, not the selected version. Two projects can select different versions from the same response. Keep metadata and download cache entries separate internally; every public Fetch error still uses Source `npm`.

For a project Dependency, preserve the original Name, declared Version, and Tag. Do not overwrite its identity with a returned npm name. Reject mismatched metadata identity as a parse failure. Never populate `resolved_version` in v1.

Support stable exact versions, partial versions, wildcards, caret/tilde ranges, comparator sets, AND/OR groups, and hyphen ranges. Use the pinned semver library behind an npm compatibility layer. Normalize range grammar and numeric bounds there. Do not claim complete npm package-spec support.

Select `dist-tags.latest` when that published version satisfies the declaration. Otherwise choose the highest matching non-prerelease version. Exclude prereleases unless `latest` is a prerelease explicitly allowed by the declared range. Apply npm's same-major/minor/patch prerelease rule. Cover the known `~0.0.0` difference and prerelease matching with checked-in parity fixtures from npm's documented behavior. A missing compatible version produces an npm `not_found` Fetch error. Read License and LastPublish from the selected version, using `time[selectedVersion]` for LastPublish. Do not substitute the document's modified timestamp or a different major version's publish time. Document this assessment-date meaning in the successor ADR and glossary.

Treat an empty declaration as wildcard for selection while preserving its original declared string. Treat npm aliases, tags, workspace references, file paths, Git references, and URLs as unsupported declarations. Keep their Cards with npm `parse` errors. Do not fetch or follow their locations. The standalone query uses wildcard selection internally and exposes no declared version.

Repair download handling: zero is known only after a valid successful response. Require a present, non-null, non-negative integer that fits the supported 64-bit targets. Missing, null, fractional, negative, overflowing, or malformed counts produce an npm parse Fetch error. An absent, expired, or failed download response must not become zero popularity. Refresh downloads even when metadata is fresh. If either required fetch fails, return a degraded Card under the existing all-or-nothing enrichment rule. Omit unknown stats and use name-only Flavor. This preserves v1 behavior without adding partial-enrichment policy.

Acquire the eight-slot semaphore only around each HTTP request and response-body read. Release it before cache writes. Cache hits use no slot. Use a ten-second timeout per network request and honor cancellation while waiting for a slot. No automatic retry loop in v1.

Keep the 24-hour success TTL, one-hour missing-package TTL, ETag refresh, and complete no-cache bypass. Use atomic cache writes. Treat malformed cache entries as misses and refetch; do not turn cache corruption into fake facts. Return a contextual operational error for unreadable or unwritable cache storage. Never silently discard a cache write failure. Permit only successful body entries and missing-package entries. Test replacement on Windows as well as first write.

### Tickets 05 and 06: stats, Check, and Flavor

Keep the existing Rarity and Chaos thresholds and default policy. Clear old errors and enrichment before applying a new result. A valid zero-download result remains full data and Common. A Fetch error means unknown enrichment.

Add the Mode type and an `ErrAINotInV1` sentinel. `rule` calls the existing deterministic engine; `none` clears FlavorText, SpecialMove, and FlavorSource; `ai` returns an error containing `ai` and `v1.1` without running the engine. Reject other modes. Validate modes before manifest reads or network activity.

For Check, return the assessed Dependencies and apply the default policy only to known data. Keep degraded Cards out of `flagged` when `fail_on_fetch_errors` is false. Define `reason` in this precedence order:

1. Nonempty Roster and every Card degraded: `all_fetches_failed`, pass false, flagged empty.
2. At least one policy flag: `policy_failed`, pass false.
3. Some Fetch errors, no policy flags: `partial_fetch_errors`, pass true.
4. Empty Roster: `empty_roster`, pass true.
5. Otherwise: `pass`, pass true.

Do not describe Check as a vulnerability scan or license-compatibility check. Return the policy that actually ran.

### Ticket 07: pipeline

Resolve a directory to its root `package.json`. Keep `parser.ParseFile` file-only. Read one Roster. Sort by Name and then Tag before assessment and output. Use at most eight worker goroutines; do not start one unbounded goroutine per Dependency. A package Fetch error must not cancel siblings. Fatal local errors and caller cancellation return a contextual Go error.

Capture one assessment time per Run or Check. Merge enrichment into the original Roster entry, calculate stats, then apply Flavor. Check must assess once, not fetch a second time. GetCard uses the same enrichment, stats, and Flavor helper without a Roster. Keep stdout, process exits, flag parsing, and MCP protocol handling outside this package.

## 4. Complete both Faces and executable wiring

### Tickets 08, 09, and 10: CLI and HTML

Keep `cli.Main(args, stdin, stdout, stderr) int` as the public CLI entry point. Add an internal dependency-injected runner for tests. Use the standard `flag` package. Accept one optional path after command flags; an omitted path means the current directory. Reject extra positional arguments and unknown flags.

Support `scan`, `check`, and top-level `mcp`. Top-level no command is usage exit 2. Explicit help returns 0. `scan` supports `--json`, `--flavor`, `--no-cache`, and `--out`. `check` supports the first three; reject `--out` there. Reject `scan --json --out ...` to prevent conflicting outputs. Keep flag-before-path syntax documented and tested.

Default scan writes `deck.html` beside the scanned manifest. An explicit relative output path is relative to the caller's current directory. `--out -` writes HTML to stdout. `--json` writes one envelope and creates no HTML. Render to a temporary sibling file and replace the target only after a successful write. Preserve an existing output on failure.

Keep the current `render.Render(deps, io.Writer)` interface and embedded template. Do not restore the stale ticket's byte-slice API. Keep html/template escaping, static filter JavaScript, stable ordering, self-contained assets, and the under-150-KB target for 47 short Cards. Show unknown downloads as unknown rather than rendering the Go zero value on degraded Cards. Label displayed statistics. Omit empty Flavor elements in none mode. Verify tag filters actually hide and restore Cards in a browser.

Data goes to stdout; diagnostics go to stderr. Default scan reports `Wrote <path> (N cards)` on stderr. Check prints a compact verdict to stdout unless JSON is requested. Fetch warnings identify a Dependency without exposing raw transport or secret data. Honor NO_COLOR; v1 may use plain text throughout.

CLI exit mapping:

| Condition | Exit |
|---|---:|
| Successful scan, passing Check, or help | 0 |
| At least one policy flag | 1 |
| Invalid input/mode, malformed manifest, or local operational failure | 2 |
| Check of a nonempty Roster where every fetch failed | 3 |

Exit 3 takes precedence over policy status. A scan may successfully return degraded Cards. An empty Roster is not total fetch failure. Validate ai before fetching or opening an output file.

### Ticket 11: MCP server

Use the approved SDK's stdio transport and explicit schemas. Expose exactly:

| Tool | Input | Envelope data |
|---|---|---|
| list_deck | path; optional Tag and Flavor | DeckData, with an optional Tag filter applied after assessment |
| get_card | name; optional Flavor | Standalone Card projection |
| check_deck | path | CheckData, including assessed Dependencies |

Use the exact agent-facing descriptions from the existing contract. Path is required for project MCP tools. Tag is one known enum value. Flavor is `rule` or `none`. Reject unknown input properties. Return the envelope in structured content and the matching JSON text block for text-only clients.

Valid tool calls with invalid arguments produce `isError: true`. Malformed JSON-RPC and unknown tools produce protocol errors. npm 404, 429, timeout, malformed metadata, and all-fetch-failed assessments remain successful tool results with Fetch errors in data. Recover handler panics into a sanitized tool error. EOF closes the server normally. Keep stdout protocol-only and diagnostics on stderr.

Use the SDK's bounded stdio reader with an explicit 16-MiB frame limit. Propagate cancellation. Share read-only options and a safe registry client; do not mutate global Flavor or cache settings per request. Test actual protocol initialization, tool listing, schemas, calls, failures, cancellation, and shutdown.

### Ticket 12: executable

Keep main limited to dependency wiring, signal cancellation, build-version injection, and process exit. Dispatch `mcp` to its stdio server; dispatch the human commands to the CLI. The only production network destinations are the public npm metadata and download endpoints. HTTP endpoint injection exists for tests, not as a user-provided manifest setting.

Rename the module to `github.com/pragatikumari0927/depdeck` and update internal imports. Move this existing release requirement out of the roadmap's conflicting v2 list. Add a version option that reports the build identity used in envelopes.

## 5. Security and verification gates

Implement security with the feature work. New ticket 15 collects the complete evidence; it is not a last-minute sanitizer pass.

The threat model must cover hostile registry/cache text, stored HTML injection, agent prompt injection, credential leakage, cache integrity, path/output writes, redirects, and resource exhaustion. State that hash-based cache filenames are not authentication. State that v1 is a local tool, not an authenticated network service. The invoking user and MCP host control local file access through OS and host permissions.

Validate requested package identities before forming a URL. Escape URL path components. Keep production hosts fixed and validate redirect destinations. Do not execute package content, install Dependencies, follow declaration URLs, fetch remote artwork, or insert registry descriptions into Flavor. Sanitize control characters and error details before either Face returns them. Refuse link or executable markup in displayed metadata; keep names, versions, and license facts distinct from instructions. Preserve html/template escaping as a separate control.

Set explicit input limits: 2 MiB for a manifest, 32 MiB for metadata, 256 KiB for downloads, and 16 MiB for an MCP frame. Reject oversized input with a clear parse/tool error and bounded memory use. Bound cache-file reads before JSON/base64 decoding: 48 MiB for encoded metadata entries and 384 KiB for encoded download entries. Then check decoded bodies against their own limits. Treat cached bodies as untrusted and apply the same metadata checks. Keep secrets in environment variables or release secret storage, never in repository config or result payloads.

### Required test matrix

| Area | Acceptance evidence |
|---|---|
| Parser/types | Three Tags, precedence, root-only workspace behavior, declared Version preservation, JSON omissions, standalone projection, schema-major compatibility. |
| Registry/cache | Range selection and npm parity cases; scoped names; unsupported declarations; zero versus missing downloads; partial cache expiry; 200/304/404; 429/5xx/timeout; corrupt cache; write errors; disabled cache; eight-request cap; cancellation. |
| Stats/Flavor | Exact Rarity/Chaos boundaries; missing publish/license; clearing stale state; every Flavor mode; deterministic captions and moves; all five Check reasons; empty Roster. |
| Pipeline | One failed package among healthy siblings; total failure; no ai fetch; fixed clock; stable results under shuffled completion; one assessment per Check; no process I/O. |
| CLI process | All exits; omitted/file/directory paths; paths with spaces; output beside a manifest outside CWD; JSON-only mode; HTML stdout; failed output replacement; help/version; no ANSI. |
| HTML | Escaping every displayed field; no external assets; identical bytes; degraded values shown as unknown; none mode; 47-Card size target; responsive view and working Tag filters. |
| MCP process | Initialize/list/call; exact three tools; explicit schemas; enum/unknown-field rejection; standalone omission; JSON/structured parity; npm errors as data; panic recovery; oversized input; cancellation/EOF. |
| Security | Malicious names/license/error strings; poisoned cache; redirect attempts; oversized bodies; no secret output; no package execution. |
| Documentation | All local links resolve; canonical/pointer ownership; rule sync has no unintended diff; existing hook-selection test still passes. |

Use fixture HTTP and temporary caches for automated product tests. Never call live npm. Separate two kinds of process evidence. Build the release binary and test help, version, invalid arguments, ai rejection, missing manifests, and an empty Roster with no network. Also seed valid fresh metadata/download entries and cached 404 entries in an isolated cache. Drive real-binary scan, check, and MCP cases against those entries, including mixed and total failure. Set cache/home locations only in the child-process environment for each platform; never alter the host's environment or read its real cache. Use a deny/recording proxy or equivalent network guard and assert zero outbound attempts for seeded-cache cases.

For fresh-fetch and HTTP failure cases, use a helper process compiled from `cmd/depdeck` tests. Its `_test.go` setup injects fixture endpoints and a temporary cache, then calls the same command dispatcher and adapters as production. Do not add runtime endpoint environment variables or fixture flags to the release binary. Label helper-process results accurately; neither fixture nor seeded-cache tests prove live npm connectivity. Update the existing verification skill/script to drive these offline checks and report both evidence types.

CI runs on the minimum Go 1.25 toolchain and the current stable toolchain. Use Linux race tests; use ordinary Windows and macOS tests. On Windows, run local race tests only with the documented 64-bit compiler. Keep CI vet, the configured golangci-lint pass, and the Node hook test. Do not repeat standalone vet/staticcheck in local post-edit checks.

Pin govulncheck `v1.7.0` and Gitleaks `v8.30.1`. Run govulncheck in its normal gating mode; JSON/SARIF output alone must not determine success. Run a redacted secret scan of the release-candidate history and current intended source. Give scanner workflows read-only repository access. Do not expose release secrets to pull-request jobs. Verify checksums for downloaded tools and pin workflow actions to reviewed commit SHAs. Record scan time and vulnerability database information; a changing vulnerability database is expected.

### Official Codex Security review

Prepare `docs/THREAT_MODEL.md` as scan context. Use a local Codex Security scan for uncommitted candidate changes and a connected GitHub review for the exact release-candidate commit when access permits. Confirm repository/account access and record host, plugin, commit, scope, scan status, evidence, and findings. A local scan and a Cloud scan are different workflows; do not claim the Markdown file is automatically synchronized into Cloud's threat model. Supply or update that context explicitly.

Triage each finding against source and reproduce it where practical. Fix verified High/Critical issues before declaring release readiness. For lower-severity accepted risk, record the owner, reason, and follow-up. Store sensitive findings privately; keep only redacted evidence or a private advisory pointer in the Markdown tracker. Do not auto-publish findings or patches.

If the service is unavailable, record that fact and use the agreed fallback: security regression tests, vulnerability/secret scans, and an independent manual review against the threat model. The reviewer must be separate from the implementation author. Record the review scope and remaining limits. Do not label this fallback a Codex Security pass.

## 6. Release readiness and completion evidence

New ticket 16 owns these deliverables:

- Root README with the pitch, install and usage commands, a working GIF from fixtures, sample Cards, JSON recipes, MCP setup, supported scope, and links to docs.
- MIT license for project-owned work, plus the applicable notices for redistributed dependencies and vendored material. Do not replace third-party skill licenses.
- Canonical module path and an isolated `go install ./cmd/depdeck` check. Prepare the public versioned install command for the release; verify it after the tag is published.
- GoReleaser OSS `v2.19.0` configuration for Windows, Linux, and macOS on amd64 and arm64. Use ZIP on Windows, tar.gz elsewhere, and SHA-256 checksums. Include the project license and applicable notices.
- Use the official prebuilt GoReleaser binary. Its own build requirement does not raise Depdeck's approved Go minimum.
- git-cliff `v2.14.2` configuration and generated Keep a Changelog output. Generate unreleased/release sections from Conventional Commits; do not hand-edit released sections.
- A source-built Homebrew formula for `pragatikumari0927/homebrew-tap`, the documented default tap. Build the command from tagged source. Include a version test. Prepare formula generation and update automation outside GoReleaser's deprecated formula publisher.
- Release preflight for tap access, tag/commit identity, checksums, least-privilege publishing credentials, and expected files. Keep the tap token separate from the default repository Actions token. Do not add quarantine-bypass hooks.

Validate packaging with `goreleaser check` and a snapshot run with publishing disabled. Use a verified disposable distribution directory before any clean step. Generate and inspect the changelog preview without creating a Git tag. Run native executable smoke tests and verify cross-target archive contents/checksums. On macOS, validate the formula against staged source and run its test; verify the public formula again after publication.

Completion evidence must identify the candidate commit, toolchain, resolved original tickets, new tickets, CI runs, offline process tests, HTML/browser checks, schema examples, security scan or fallback, package checksums, install checks, and documentation checks. Distinguish implemented, verified, merged, release-ready, and published states.

The final publication sequence is separate: approve the reviewed candidate, create the v1.0.0 tag, publish GitHub assets, publish the formula update, and verify public installation. If publication fails, retain the evidence and fix the failed step. Never move an already-published tag to hide a bad release. Use a patch release for a published defect.

### Definition of release-ready

All original tickets meet their amended criteria. Documentation/security/release tickets meet their criteria. The real CLI and MCP server pass offline acceptance. No verified High/Critical security finding remains open. The official security review or documented fallback is complete. Packages and the formula pass preflight. No unresolved implementation decision remains. Publishing itself has not occurred.

### Primary references verified for this plan

- Official MCP SDK version, Go minimum, and tool/stdio support: [v1.8.0 module](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/go.mod), [server documentation](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/docs/server.md), [transport](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/transport.go).
- Semver dependency and its compatibility boundary: [v3.5.0 documentation](https://github.com/Masterminds/semver/blob/v3.5.0/README.md), [constraint code](https://github.com/Masterminds/semver/blob/v3.5.0/constraints.go), [npm range and prerelease rules](https://github.com/npm/node-semver/blob/main/README.md).
- Security service setup and context: [local plugin](https://learn.chatgpt.com/docs/security/plugin), [Cloud setup](https://learn.chatgpt.com/docs/security/setup), [Cloud threat model](https://learn.chatgpt.com/docs/security/threat-model).
- Security tool compatibility: [govulncheck v1.7.0 module](https://raw.githubusercontent.com/golang/vuln/v1.7.0/go.mod), [Gitleaks v8.30.1 module](https://raw.githubusercontent.com/gitleaks/gitleaks/v8.30.1/go.mod).
- Release tooling: [GoReleaser v2.19.0](https://github.com/goreleaser/goreleaser/releases/tag/v2.19.0), [git-cliff v2.14.2](https://github.com/orhun/git-cliff/releases/tag/v2.14.2), [Homebrew formula-publisher deprecation](https://www.goreleaser.com/blog/goreleaser-v2.16/), [tap conventions](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap).
