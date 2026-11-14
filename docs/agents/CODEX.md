# Codex agent guide

Codex must read root `AGENTS.md` first. Keep this file for Codex-specific behavior. Keep hooks and configuration in `.codex/` and other tool-discovery locations.

Use the repository permissions. Do not read or write outside approved workspace paths. Preserve unrelated changes. On Windows, prefer PowerShell commands that operate on literal paths. Do not chain destructive file operations.

Before work, inspect the relevant ticket and canonical docs. Before handoff, report changed paths, verification, and known risks. Use the review checklist in [REVIEW.md](REVIEW.md).

For v1, request the official Codex Security review if the account supports it. If it does not, record the limitation and complete the fallback in [SECURITY.md](../SECURITY.md). Never claim that a scan ran when it did not.
