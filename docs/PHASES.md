# PHASES

Roadmap and git workflow. Tickets implement against [SPEC.md](SPEC.md). This file keeps later work out of v1.

## Status

v1 in progress. The baseline records 7 of the original 13 tickets as merged. New v1 work is tracked in tickets 14–16. Update this count only after work is merged.

## v1 — current

Scope lock: [SPEC.md](SPEC.md). Why: [DECISIONS.md](DECISIONS.md).

- Commands: `scan`, `check`, `mcp`
- Ecosystem: npm only
- Flavor: `rule` and `none` (no `ai` runtime)
- Output: single `deck.html`, self-contained, byte-deterministic
- Test seam: pipeline result, fixture npm, no live registry
- Completion plan: `.scratch/depdeck-v1/COMPLETION_PLAN.md`
- Release readiness requires tickets 14–16 plus the original product tickets.

## v1.1 — next

- `--flavor=ai` runtime via Ollama (ADR-0005 reserved the flag)
- Individual card PNG export via chromedp, falling back to HTML when Chrome is missing
- `install_size` as a Card stat (requires ADR to unfreeze `pkg/types`)
- `depdeck doctor` — scan mcp.json, rule files, cache for leak categories
- Decide whether `.agents/skills/` is tracked or ignored. If ignored,
  track `skills-lock.json` and add `npx skills install
  --frozen-lockfile` to `DEVELOPMENT.md`. Standard pattern: ignore the
  directory like `node_modules/`, track the lock like
  `package-lock.json`. Triage report: 2026-09-11.

## v2 — deferred

- GitHub enrichment — second source in `pkg/registry`
- Multi-ecosystem parsers (go.mod, Cargo.toml, pyproject.toml)
- `depdeck battle` — compare two decks
- Monorepo CI check — fan-out across workspaces
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

- All original product tickets and v1 tickets 14–16 merged
- Go 1.26.9 and official MCP SDK v1.8.0 recorded in module metadata
- Security and acceptance evidence from ticket 15
- Module path renamed
- README with a working GIF demo
- GitHub binary archives and source-built Homebrew formula
- `go install` works from the canonical path
- `CHANGELOG.md` generated from commit history via `git-cliff`. Format: Keep a Changelog. Regenerate on every release; do not hand-edit released sections.
