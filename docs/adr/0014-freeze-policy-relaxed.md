# ADR-0014: Relax pkg/types freeze for additive fields

**Date:** 2026-09-11
**Status:** accepted
**Deciders:** project

## Context

`pkg/types` has been unfrozen twice in the first week for additive field changes (SpecialMove via ADR-0013, and earlier type additions during initial build). The "frozen, ADR required for any change" rule was designed to prevent renames and removals — the failure mode where a downstream consumer's code stops compiling. Additive fields do not cause that failure.

## Decision

`pkg/types` is frozen against removals, renames, and changes to the type or json tag of existing fields. These require an ADR. Additive fields — new struct fields with omitempty — do not require an ADR.

## Alternatives Considered

### Keep a full freeze (ADR for every field)
- **Pros:** Every contract change is indexed
- **Cons:** Additive omitempty fields still need a one-time unfreeze each time
- **Why not:** That is the failure mode this ADR removes; compile-safe additions do not need a per-field ADR

### Unfreeze the package entirely
- **Pros:** Fastest iteration
- **Cons:** Renames and removals can break parser, registry, stats, flavor without a recorded decision
- **Why not:** Those changes still need an ADR and a migration plan

## Consequences

### Positive
- Downstream packages (parser, registry, stats, flavor) cannot be broken without an ADR.
- New fields discovered during implementation (e.g. a registry client discovering a new npm metadata field worth exposing) can be added directly.

### Negative
- Additive fields can land without an index row unless someone chooses to record them.

### Risks
- Silent contract growth. Mitigation: omitempty only; renames, removals, and type/tag edits still need an ADR plus a downstream migration plan.

ADR-0013 remains the precedent for the additive case, but no longer needs a per-field ADR going forward.
