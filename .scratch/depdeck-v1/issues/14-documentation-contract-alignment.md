Status: ready-for-human

# 14 — Documentation and contract alignment

**Parent:** [v1 completion plan](../COMPLETION_PLAN.md)

Move canonical project guidance into `docs/`, retain root pointers, add the security, threat, testing, release, and shared agent documents, and update ticket interfaces to match the settled v1 decisions.

## Acceptance criteria

- [x] All eight project guidance files have one canonical copy in `docs/` and a short root pointer.
- [x] Root `AGENTS.md` gives required startup steps and points to canonical project docs.
- [x] `docs/README.md`, `SECURITY.md`, `THREAT_MODEL.md`, `TESTING.md`, `RELEASING.md`, and shared agent guides exist.
- [x] Successor ADRs record documentation/security, Go and npm versions, response shapes, and CLI defaults.
- [x] Decision index, PRD, roadmap, active rules/config references, and tickets use canonical paths and agreed contracts.
- [x] Ticket 13 is ready before ticket 04; ticket dependencies agree with the completion order.
- [x] Local Markdown links resolve. Generated Grok rules match Cursor sources where generated copies change.
- [x] Accepted ADRs, historical handoffs, and package sources match the pre-move SHA-256 baseline.
- [x] Verification results are recorded below.

## Verification

- Local Markdown links: 58 files checked; no missing targets.
- `git diff --check`: passed.
- Cursor-to-Grok rule sync: generated pairs match.
- Pre-move SHA-256 snapshot: accepted ADRs, historical docs, and package files unchanged. Temporary snapshot removed.
- Go product tests were not run. This ticket changes documentation, ticket text, and rule references only.

This ticket is ready for review. It is not marked merged.
