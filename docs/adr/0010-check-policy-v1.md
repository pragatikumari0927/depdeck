# v1 Check policy numbers

`check` / `check_deck` is meaningless without numbers. Policy is returned in the envelope (`data.policy`) so agents see what ran.

**Status:** accepted

## v1 defaults

| Field | v1 value | Flag when |
| --- | --- | --- |
| `chaos_score_max` | `0.5` | Chaos **greater than** 0.5 |
| `age_years_max` | `null` | Not enforced (staleness is Chaos) |
| `require_license` | `true` | License missing or empty after a successful fetch |
| `fail_on_fetch_errors` | `false` | Matches ADR-0003: partial fetch errors do not fail Check |

Degraded Cards: do not flag on Chaos, age, or license (those facts are unknown). They appear in `flagged` only if `fail_on_fetch_errors` is true (v1: never).

## Chaos (v1)

A 0.0–1.0 **Chaos** score from last-publish age on the npm document (not GitHub):

- last publish within 2 years → `0.0`
- within 5 years → `0.6`
- older, or no last-publish on a successful fetch → `1.0`

With `chaos_score_max` 0.5, packages unpublished for 2+ years are flagged. Missing last-publish on a **failed** fetch does not produce Chaos (omitted).
