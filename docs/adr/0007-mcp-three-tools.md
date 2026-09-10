# v1 MCP: three tools, one envelope, data vs isError

MCP exposes `list_deck`, `get_card`, and `check_deck`. No HTML, audit, AI, search, or compare tools. `get_card(name)` talks to npm only — no Roster path, no `not_on_roster`. Project membership is `list_deck(path)` plus filter.

**Status:** accepted

## Envelope

Every successful tool result:

`{ "schema_version": "1.0", "tool": "<name>", "data": { ... } }`

`flavor` on `list_deck` / `get_card` is JSON Schema enum `rule` | `none` (not `ai` — validation failure, not runtime). Declare `inputSchema` explicitly.

## Errors

`isError` is tool failure (bad args, panic). npm 404/429/timeout and “every fetch failed” are **normal** results with `errors` on Dependencies.

## `check_deck` data

`pass`, `flagged` (Dependency objects), `reason`, and `policy` (`chaos_score_max`, `age_years_max`, `require_license`, `fail_on_fetch_errors`). Null policy fields are not enforced in v1.

## Descriptions (agent-facing)

- `list_deck`: List every direct dependency of a project as Cards. Use for project-wide questions.
- `get_card`: Fetch a single package's Card from npm. Use for package-specific questions.
- `check_deck`: Return which dependencies are flagged as risky, and an overall pass/fail. Use before shipping or for CI gating.
