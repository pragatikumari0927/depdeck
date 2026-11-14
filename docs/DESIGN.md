# DESIGN — depdeck Card and Deck tokens

Token file for Card and Deck faces. Template applies values. Thresholds: ADR-0009. HTML bytes: ADR-0006. Flavor modes: ADR-0005. Vocabulary: CONTEXT.md. Contract fields: SPEC.md Core types.

The first sections record the v1 baseline. The Pixel Archive section records the new HTML design system and its rollout.

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

## Pixel Archive v2 — Pixel Utility Pop

The HTML Deck uses an aligned Swiss grid, pixel-game emblems, and beveled
Windows 95/98 controls. The generated Deck includes a title bar, collection
summary, Rarity Atlas, Tag controls, and a status area.

### Typography, grid, and chrome

- Use the local system sans-serif stack for body copy at 14–16 px with at least
  1.4 line height. Use the local monospace `--font-display` stack for short,
  bold window titles, Atlas labels, and Card headings. Keep names and notes in
  mixed case and allow long names and versions to wrap. Draw pixels in artwork
  and borders so text remains clear at zoom.
- Use `--space-1` through `--space-4` (4, 8, 12, and 16 px) for insets and
  gaps. Align labels and values across Cards. Use `minmax(0, 1fr)` for tracks
  containing long text and keep page padding within the available width.
- Use `--archive-surface` (#16263B), `--archive-panel` (#101D31),
  `--archive-line` (#536A89), `--archive-ink` (#F1F5FB), and
  `--archive-muted` (#B9CADF) for the window and its panels. Use a compact
  title bar and square, beveled controls. Give buttons visible default, hover,
  pressed, and keyboard-focus states. Use `aria-pressed` for the active Tag
  filter. Window decoration stays non-interactive.

### Rarity and Cards

| Label | CSS custom property | Color | Seal |
| --- | --- | --- | --- |
| Legendary | `--rarity-legendary` | #FFDF79 | Gold crown with a star |
| Epic | `--rarity-epic` | #E0A6FF | Violet six-point crest |
| Rare | `--rarity-rare` | #87F2FF | Cyan faceted crystal |
| Common | `--rarity-common` | #D9E5F2 | Slate pixel shield |
| Unknown | `--rarity-unknown` | #FFD8B4 | Peach question mark with a detached pixel |

Define each original seal once as inline SVG and reuse it on Cards and in the
Atlas. Use the seal, a text label, and a colored leading border together. The
Atlas shows each seal, name, count, threshold, and a short explanation. It
includes zero counts and describes the full Deck while any Tag filter is active.
Use the weekly download cutoffs in [ADR-0009](adr/0009-rarity-thresholds.md). Known zero belongs to Common.
Unknown is an HTML label for a Card with no verified rank.

Use a square Card frame, a consistent neutral face, and a tier accent on the
heading. The illustration area uses an enlarged tier seal and a shared
geometric backdrop at a 2:1 aspect ratio. Keep package names and declared versions readable at full
length. Per-package artwork requires an approved design and licensing decision.
See [ADR-0006](adr/0006-self-contained-deck-html.md).

### Field Journal and Archive Notices

Use a quiet inset panel for the Field Journal. Separate playful Flavor from a
two short notes drawn from assessed fields. The first links weekly downloads
with the seal. The second records the dependency group, publish date and
recency at assessment, and license. Give missing facts plain labels. In
`--flavor none`, keep these notes and omit Flavor and Special Move. Use a bordered Archive Notice with a square caution emblem, a heading,
a plain-language explanation, and a next step where useful. Treat each distinct
message as a static, named note. Map the five fetch kinds and a neutral fallback
to fixed copy. Keep structured errors on the CLI and MCP Faces. Avoid treating unavailable data as zero or implying that Rarity,
Chaos, or license availability measures security.

### Responsive and accessible behavior

- At 320–639 px, use one Card column and stack the Atlas above the Deck.
  Allow the filter controls to wrap. From 640–1023 px, keep the Atlas above a
  two-column Deck. From 1024 px, use an Atlas sidebar and a Deck grid that
  uses two Card columns through 1279 px and three from 1280 px.
- Use a named Atlas section, visible text labels, and decorative SVGs hidden
  from assistive technology. Keep controls keyboard-operable and use a visible
  focus outline. Announce the filtered Card count with a polite live region.
  Check text contrast at 4.5:1 and essential control boundaries and focus
  indicators at 3:1. Keep transitions and scrolling effects off.
- Keep one deterministic HTML file under 150,000 uncompressed bytes for a
  populated 47-Card fixture. Embed CSS, the five seals, and static JavaScript
  once. Use local fonts and zero external asset requests. Use document
  scrolling and keep every Card available. Measure large-Deck layout before
  adding rendering optimizations. No external UI framework or added dependency.

### File-level plan

1. `pkg/render/render.go` builds the renderer-only view and full-Deck counts
   from assessed dependencies. `pkg/render/templates/deck.html` defines the
   five symbols, tier tokens, and Atlas. `pkg/render/render_test.go` checks
   cutoffs, known zero, Unknown, fixed references, determinism, and size.
2. Extend the renderer view and template for fixed Archive Notices and the
   Field Journal. Replace the existing raw-error assertion with checks for
   readable Notices. Keep scoring, fetching, and public response types intact.
3. Apply the planned chrome and responsive Card grid in the template.
   Extend `scripts/deck-filter.test.cjs` for the live visible count.
   Verify Go-rendered 0-, 1-, 6-, and 47-Card fixtures at 320, 390, 768, and
   1440 px. Check keyboard focus, contrast, reduced motion, scrolling, long
   names, hostile text, filter behavior, and zero external requests.
4. Record the full-layout checks in `docs/TESTING.md`. Update the README demo
   with a capture from the Go renderer once the full layout is verified.
