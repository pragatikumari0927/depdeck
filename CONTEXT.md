# Depdeck

Depdeck is the context for turning a project's declared dependencies into a deck of Pokémon-style cards, and for answering the same facts to humans and to agents.

## Language

**Dependency**:
One named package that belongs on the Deck because the project chose it. Direct only — never a transitive.
_Avoid_: package, module, library, transitive (when meaning a Deck member)

**Roster**:
The project's statement of intent: names and ranges listed in `package.json`. The Deck is built from the Roster, not from the lockfile tree.
_Avoid_: lockfile, install tree, node_modules

**Tag**:
Exactly one Roster list a Dependency came from: `dependencies`, `devDependencies`, or `optionalDependencies` (single string, not an array). Parser order if a name appears twice: `dependencies`, then `devDependencies`, then `optionalDependencies` — first wins, no duplicate Cards.
_Avoid_: tags, kind, group

**Declared version**:
The range or exact version written on the Roster. Contract field: `version`.
_Avoid_: version_range, spec, constraint (when meaning this field)

**Resolved version**:
The version actually installed, taken from the lockfile when one exists (`resolved_version` in the shared contract). Absent when there is no lockfile.
_Avoid_: installed version, pinned version (use this term)

**Card**:
The Pokémon-style presentation of one Dependency.
_Avoid_: tile, widget, badge

**Deck**:
The collection of Cards for one project. The human Face writes it as one self-contained HTML file beside `package.json` (`deck.html` by default).
_Avoid_: report, listing, dashboard, live preview

**Flavor**:
The personality text applied to a Card. v1 settings are `rule` and `none`. `ai` is reserved and rejected, not silently turned into `rule`.
_Avoid_: blurb, copy, lore (until Flavor includes lore)

**Flavor source**:
Which template tier produced Flavor: `rule_full` (stats present) or `rule_name_only` (fetch failed; name hash only). Distinct from Fetch error: errors say the Source failed; Flavor source says which template ran.
_Avoid_: flavor_mode (that is the `--flavor` setting)

**Face**:
A thin adapter over the shared assessment: the human CLI or the agent MCP. Neither Face owns the assessment.
_Avoid_: frontend, UI, wrapper CLI (when meaning the MCP Face)

**Fetch error**:
A structured failure to enrich a Dependency from one Source. The Card still exists; unknown stats stay unknown (not faked as Common or zero-as-rank).
_Avoid_: exception, 404 (when meaning this attached fact)

**Source**:
Where enrichment comes from. v1 has npm; GitHub is reserved for later. Fetch concurrency is bounded per Source, and only a network call consumes a slot.
_Avoid_: registry, API (when meaning this named origin)

**Rarity**:
The Card's popularity tier after a successful npm fetch: `Common`, `Rare`, `Epic`, or `Legendary`, from weekly downloads (`>=10000` / `>=5000` / `>=1000` / else). Omitted when fetch failed — not Common.
_Avoid_: rank, star rank (v1 does not use GitHub stars)

**Chaos**:
A 0.0–1.0 staleness score from last npm publish (within 2 years = 0.0, within 5 years = 0.6, older = 1.0). Omitted on degraded Cards.
_Avoid_: risk, health (Check is the verdict; Chaos is one input)

**Check**:
A Face-specific verdict over a Deck: pass/fail plus flagged Dependencies. Fetch errors do not fail Check unless every fetch failed (CLI) or as policy `fail_on_fetch_errors` (default false).
_Avoid_: audit, scan (scan produces the Deck; Check judges it)
