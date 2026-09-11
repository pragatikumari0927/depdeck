# PHASES

Roadmap and git workflow. Tickets implement against [SPEC.md](SPEC.md). This file keeps later work out of v1.

## Status

v1 in progress. 7 of 13 tickets merged.

## v1 — current

Scope lock: [SPEC.md](SPEC.md). Why: [DECISIONS.md](DECISIONS.md).

- Commands: `scan`, `check`, `mcp`
- Ecosystem: npm only
- Flavor: `rule` and `none` (no `ai` runtime)
- Output: single `deck.html`, self-contained, byte-deterministic
- Test seam: pipeline result, fixture npm, no live registry

## v1.1 — next

- `--flavor=ai` runtime via Ollama (ADR-0005 reserved the flag)
- Individual card PNG export via chromedp, falling back to HTML when Chrome is missing
- `install_size` as a Card stat (requires ADR to unfreeze `pkg/types`)
- `depdeck doctor` — scan mcp.json, rule files, cache for leak categories

## v2 — deferred

- GitHub enrichment — second source in `pkg/registry`
- Multi-ecosystem parsers (go.mod, Cargo.toml, pyproject.toml)
- `depdeck battle` — compare two decks
- Monorepo CI check — fan-out across workspaces
- Module path rename to `github.com/pragatikumari0927/depdeck`
- Graph engineering — see `docs/adr/0015-graph-engineering-deferred.md` when written

## Git workflow

- Trunk-based. `main` always shippable.
- Conventional Commits. One logical change per commit.
- One commit per ticket, not per batch.
- Worktree parallel variants: `git worktree add ../depdeck-<variant> -b <branch>`. One agent per worktree; never two on the same tree.
- Post-edit hooks run scoped tests. Pre-commit runs full-tree `go test ./... -race` on Linux CI.
- PR under 400 lines. Bigger means the ticket was too large.
- Squash to `main`. PR description becomes the commit message.

## Release gate — before tagging v1.0.0

- All 13 tickets merged
- Module path renamed
- README with a working GIF demo
- GoReleaser config producing GitHub Release + Homebrew tap
- `go install` works from the canonical path
- `CHANGELOG.md` generated from commit history via `git-cliff`. Format: Keep a Changelog. Regenerate on every release; do not hand-edit released sections.
