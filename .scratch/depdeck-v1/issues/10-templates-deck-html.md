Status: ready-for-agent

# templates/deck.html — the artifact

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Self-contained HTML template: inline CSS/JS, Tag filter as **static** script, Card layout, degraded/no-data styling. Target ~150KB for 47 short Flavor sentences. No per-card SVG logos. No `generated_at`.

## Files

- Create `templates/deck.html`

## Interface

Go template actions over `.` as a slice (or struct `{ Cards []types.Dependency }`). Fields referenced must exist on types.Dependency. Filter JS must not be built from `template.JS`.

## Acceptance criteria

- [ ] File exists and parses with `html/template`
- [ ] Contains a static `<script>` for Tag filter (no data interpolation inside that script)
- [ ] Has a hook for missing Rarity / errors (no-data)
- [ ] No `Date.now()`, `Math.random()`, or generated_at placeholders

## Blocked by

None - can start immediately (keep field names aligned with 01-pkg-types)

## ADRs

0006
