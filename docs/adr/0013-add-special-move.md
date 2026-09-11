# ADR-0013: Add SpecialMove to Dependency

**Date:** 2026-09-11
**Status:** accepted
**Deciders:** project

## Context

`SPEC.md` lists `special_move` on the Core types Dependency shape. `pkg/types/types.go` does not have the field. ADR-0005 requires a distinct move string separate from flavor text. The field is documented but unimplemented.

## Decision

Add `SpecialMove` to `Dependency` in `pkg/types`. This is a one-time unfreeze of the frozen package.

```
SpecialMove string `json:"special_move,omitempty"`
```

Placement is immediately before `FlavorSource`, matching SPEC ordering. The field is omitted when flavor is `none`, same rule as `FlavorText`. No other field changes.

## Alternatives Considered

### Leave types.go unchanged
- **Pros:** Frozen package stays untouched
- **Cons:** SPEC and DESIGN.md describe a field the compiled contract cannot carry
- **Why not:** Faces and the flavor engine have no place to put the move string

### Fold the move into `flavor_text`
- **Pros:** No struct change
- **Cons:** One string cannot stay distinct from flavor text (ADR-0005)
- **Why not:** Template and JSON consumers cannot split move from caption reliably

## Consequences

### Positive
- SPEC and `types.go` align on the Dependency shape.
- `pkg/types` returns to frozen after this commit.

### Negative
- One unfreeze of a package that was marked frozen.

### Risks
- Further “just one field” unfreezes. Mitigation: freeze again this commit; next shape change needs a new ADR.

### Intentional remaining divergences
- `Chaos` is `*float64` in Go so omitempty can drop unknown; SPEC lists a number.
- `FetchError.Kind` is `string` on the struct; allowed values stay those in ADR-0003. These are Go idioms, not drift.
