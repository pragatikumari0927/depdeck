# Rarity from npm weekly downloads

Rarity vocabulary is `Common | Rare | Epic | Legendary`. Unknown (omitted) when fetch failed — not Common (ADR-0003). v1 ranks by **npm weekly downloads**, not GitHub stars: GitHub is a v2 Source. The original pitch numbers still apply as cutoffs.

**Status:** accepted

## Cutoffs (weekly downloads)

- `>= 10000` → Legendary
- `>= 5000` → Epic
- `>= 1000` → Rare
- else (successful fetch) → Common

Inclusive lower bounds; first match from the top. Degraded Cards omit Rarity.
