# v1 Flavor is `rule` or `none`; `ai` errors

`--flavor` default is `rule`. `none` strips personality text. `--flavor=ai` is a known flag that **fails** with a clear message (*planned for v1.1; v1 supports `rule` and `none`*), CLI exit code 2, MCP structured error — never silent fallback to `rule`. That keeps JSON honest (`flavor_text` without a model vs template) and lets v1.1 turn the flag on without an API change.

**Status:** accepted

## Rule engine

Pick template tier from `len(errors) > 0`, not from zero stats. `full` uses name + stats; `name_only` uses name hash. Same Dependency struct → same Flavor. Warm vs cold cache with successful fetch → same Flavor. npm up vs down → different Flavor (full vs name-only); do not “fix” that by forcing name-only.

`flavor_source` is `rule_full` | `rule_name_only` (omit when Flavor is `none`).

Rule templates must match v1 metrics: weekly downloads and publish recency (Chaos). Do not mention GitHub issues or “buried under a mountain of issues.”
