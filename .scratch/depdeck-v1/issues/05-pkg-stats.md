Status: ready-for-agent

# pkg/stats — Rarity and Chaos

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Pure functions: weekly downloads → Rarity; last publish → Chaos pointer; Check policy evaluation. No HTTP.

## Files

- Create `pkg/stats/stats.go` (and `_test.go`)

## Interface

```go
package stats

func Rarity(weeklyDownloads int) types.Rarity
// >=10000 Legendary, >=5000 Epic, >=1000 Rare, else Common

func Chaos(lastPublish time.Time, now time.Time) float64
// <2y → 0.0; <5y → 0.6; else 1.0

func DefaultPolicy() types.CheckPolicy
// ChaosScoreMax 0.5, AgeYearsMax nil, RequireLicense true, FailOnFetchErrors false

func Apply(dep types.Dependency, weekly int, lastPublish time.Time, license string, fetchErrs []types.FetchError) types.Dependency
// if fetchErrs non-empty: omit Rarity/Chaos/License/WeeklyDownloads; keep Errors; do not treat as Common
// else set WeeklyDownloads, Rarity, Chaos, LastPublish, License

func Flagged(deps []types.Dependency, p types.CheckPolicy) (pass bool, flagged []types.Dependency, reason string)
```

Flag when Chaos > max (if Chaos non-nil), or require_license and License empty on a **successful** fetch (no errors). Degraded: not flagged unless FailOnFetchErrors.

## Acceptance criteria

- [ ] Table test Rarity cutoffs including 999, 1000, 4999, 5000, 9999, 10000
- [ ] Chaos 1y → 0.0, 3y → 0.6, 6y → 1.0
- [ ] Apply with not_found errors: Rarity empty, Chaos nil, not Common
- [ ] Flagged: missing license on success → flagged; missing license with errors → not flagged when FailOnFetchErrors false
- [ ] Chaos 0.6 with max 0.5 → flagged

## Blocked by

- `.scratch/depdeck-v1/issues/01-pkg-types.md`

## ADRs

0003, 0009, 0010
