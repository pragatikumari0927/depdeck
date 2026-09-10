Status: ready-for-agent

# pkg/flavor — rule and none; ai is an error

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Deterministic Flavor text. Mode `ai` returns a typed error (CLI maps to exit 2). Templates use downloads and Chaos/recency, never GitHub issues.

## Files

- Create `pkg/flavor/flavor.go` and `pkg/flavor/rule.go` (and `_test.go`)

## Interface

```go
package flavor

type Mode string // "rule" | "none" | "ai"

var ErrAINotInV1 = errors.New(`--flavor=ai is planned for v1.1; v1 supports rule and none`)

func Apply(mode Mode, dep types.Dependency) (types.Dependency, error)
// none: clear FlavorText and FlavorSource
// ai: return ErrAINotInV1, dep unchanged
// rule: if len(Errors)>0 → name_only hash templates, FlavorSource rule_name_only
//       else → full templates using WeeklyDownloads/Chaos/License, FlavorSource rule_full
```

Same Dependency in → same FlavorText out (table tests).

## Acceptance criteria

- [ ] Apply(`ai`, dep) returns ErrAINotInV1 and empty FlavorText (no rule fallback)
- [ ] ErrAINotInV1.Error() contains `ai` and `v1.1`
- [ ] Degraded (errors set) → FlavorSource `rule_name_only`; FlavorText does not mention issues/GitHub
- [ ] Full card FlavorText may mention downloads or recency, never “issues”
- [ ] Two Apply calls with equal dep → equal FlavorText
- [ ] Stars==0 with empty Errors still `rule_full` (not name_only)

## Blocked by

- `.scratch/depdeck-v1/issues/01-pkg-types.md`

## ADRs

0005, 0003
