Status: ready-for-agent

# pkg/render — html/template, byte-deterministic

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Render a Deck to HTML bytes. `html/template` only. No timestamps, random IDs, or `Date.now()` in generated output. Filter JS is a static template string. Never `template.HTML`/`JS` on Dependency fields.

## Files

- Create `pkg/render/render.go` (and `_test.go`)
- Embed or load `templates/deck.html` (created in ticket 10)

## Interface

```go
package render

func Deck(deps []types.Dependency) ([]byte, error)
```

## Acceptance criteria

- [ ] **Byte-determinism:** `Deck(deps)` twice → `bytes.Equal`; same for shuffled-then-sorted stable name order (document sort by Name)
- [ ] Input description `<script>alert(1)</script>` appears escaped in output, not as a tag
- [ ] Output contains no RFC3339 “generated” timestamps
- [ ] Degraded Card (empty Rarity) still rendered (grey/no-data class or equivalent in template)

## Blocked by

- `.scratch/depdeck-v1/issues/01-pkg-types.md`
- `.scratch/depdeck-v1/issues/10-templates-deck-html.md`

## ADRs

0006, 0003
