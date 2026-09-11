# RULES.md

Index and operating contract for `.cursor/rules/*.mdc`.
Rule bodies live in those files. This file does not restate them.

## Index

| File | Layer | Fires on | Purpose |
|------|-------|----------|---------|
| `common-security.mdc` | common | Always | Untrusted input, MCP output safety |
| `architecture.mdc` | golang | Always | Package boundaries, frozen types |
| `golang-coding-style.mdc` | golang | `**/*.go`, `**/go.mod`, `**/go.sum` | Formatting, interface size |
| `golang-patterns.mdc` | golang | `**/*.go`, `**/go.mod`, `**/go.sum` | Point-of-use interfaces, concurrency, pure functions |
| `golang-hooks.mdc` | golang | `**/*.go`, `**/go.mod`, `**/go.sum` | Post-edit automation chain |
| `golang-testing.mdc` | golang | `**/*.go`, `**/go.mod`, `**/go.sum` | Table-driven tests, coverage |
| `cli-contract.mdc` | golang | `cmd/**/*.go`, `internal/cli/**/*.go`, `internal/mcp/**/*.go` | Process status, stdout/stderr, color |

## Meta-rules

1. Every `.mdc` declares `description`, `globs` (if scoped), `alwaysApply`. Missing any field means the rule never fires. This is the most repeated failure mode in the project.
2. Rules are derived from `DECISIONS.md` and `SPEC.md`. When a rule and a source of truth disagree, fix the rule. Do not edit the source.
3. Keep each rule under 500 lines. Split by concern, not by size.
4. Add a rule only when the agent repeats a mistake. No speculative rules.
5. Never duplicate content across rule files. One rule owns a topic; others point at it.
6. After any agent edit to `.cursor/rules/`, read the first five lines of the edited file and confirm all three frontmatter fields are present before committing.
7. Two-layer structure: `common-*` first (language-agnostic), then `golang-*` (Go extensions). A `.go` file receives both layers.

## Ignore files

`.cursorignore` blocks artifacts, caches, secrets, and OS noise. Never add source, docs, rules, ADRs, tickets, or test fixtures. Verify path safety before committing any change. On Windows, `FileInfo.Extension` equals the leading-dot name for dotfiles — the "no extension" check means `Extension == Name`.

## Skill selection

See `SKILLS.md` for the curated skill list and divergence pins. The rules here are separate from skills — rules are `.mdc` files loaded by Cursor; skills are `SKILL.md` files loaded by the agent on invocation.

## Cross-tool contract

`AGENTS.md` is the tool-agnostic contract read by Claude Code, Codex, Cursor, and Grok. `.cursor/rules/*.mdc` are Cursor-specific. When the two overlap, `AGENTS.md` is the human-facing summary and the `.mdc` files are the mechanically-enforced version.
