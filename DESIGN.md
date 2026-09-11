# DESIGN — depdeck v1 Card tokens

Token file for Card and Deck faces. Template applies values. Thresholds: ADR-0009. HTML bytes: ADR-0006. Flavor modes: ADR-0005. Vocabulary: CONTEXT.md. Contract fields: SPEC.md Core types.

Subjective hexes, stacks, and lengths below are **v1** choices, not permanent.

## Card anatomy

Top to bottom on one Card. One line each.

- Rarity indicator — color swatch plus label `Common` | `Rare` | `Epic` | `Legendary`; field `rarity`. Empty `rarity` gets no hue and no Common stand-in (ADR-0003).
- Name and version — fields `name`, `version`.
- Stats row — weekly downloads, age, chaos, license; fields `weekly_downloads`, `last_publish` (age is derived), `chaos`, `license`.
- Flavor text — field `flavor_text`.
- Special move — field `special_move`; omit when flavor is `none` (no `special_move`, empty `flavor_source`).
- Error indicator — render when `errors` is non-empty.

## Rarity tokens

One named color per tier. Assignment cutoffs live in ADR-0009; this file does not copy them. Mid-chroma so the same hex reads as border or label on light and dark.

- `rarity-common` `#64748B` (v1)
- `rarity-rare` `#2563EB` (v1)
- `rarity-epic` `#7C3AED` (v1)
- `rarity-legendary` `#B45309` (v1)

## Typography

- `font-html` (v1): `system-ui, "Segoe UI", Roboto, Helvetica, Arial, sans-serif` — HTML Deck only; no external font loads (ADR-0006).
- `font-terminal` (v1): `ui-monospace, "Cascadia Mono", Consolas, monospace`

## Layout

Target: 47 Cards readable in one desktop page; one column on mobile.

- `card-width` 220px (v1)
- `card-gap` 12px (v1)
- `breakpoint-stack` 640px (v1) — single column below this width
- `breakpoint-wide` 1280px (v1) — six-column grid; 47 Cards ≈ eight rows

Self-contained Deck file: ADR-0006.

## Terminal tokens

lipgloss color names, same rarity hues, shifted for terminal contrast (v1).

- `term-rarity-common` `gray`
- `term-rarity-rare` `blue`
- `term-rarity-epic` `magenta`
- `term-rarity-legendary` `yellow`

`NO_COLOR`: strip all ANSI. Rarity is the text label only.

## Error token

- `error-ink` `#B91C1C` (v1) — error indicator when `errors` is non-empty.

## Flavor voice

Flavor is a deterministic caption of that Card’s own fields; it invents no maintainer, purpose, or GitHub facts. ADR-0005 distinguishes `rule`, `none`, and `ai`.
