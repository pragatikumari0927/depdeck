# v1 security threat model

The pipeline, HTML deck, and MCP Face all handle untrusted registry text. Without a locked threat list, Faces will grow ad-hoc sanitizers or none.

**Date:** 2026-09-11
**Status:** accepted

## Context

v1 fetches npm metadata, caches it on disk, renders `deck.html`, and returns the same Dependency objects over MCP. Package text can inject into agents, execute in a browser, or persist as a poisoned cache entry. Secrets must not live in config files.

## Decision

The five threats and their mitigations in [SECURITY.md](../../SECURITY.md) are locked for v1.

## Alternatives Considered

### Trust npm metadata
- **Pros:** Less sanitizing code
- **Cons:** Descriptions and names are attacker-controlled
- **Why not:** XSS and prompt injection become default

### Duplicate the threat list in the Cursor rule
- **Pros:** Agents see full text without opening SECURITY.md
- **Cons:** Drift between rule and doc
- **Why not:** SECURITY.md is the single catalog

## Consequences

### Positive
- One catalog for attack, mitigation, and reporting
- Existing ADRs (0004, 0005, 0006, 0007) stay the mitigation locks; this ADR locks the set

### Negative
- [DECISIONS.md](../../DECISIONS.md) is not updated in this change

### Risks
- Agents may skip SECURITY.md if they only load the short rule — the rule names the file as source of truth
