Status: ready-for-agent

# internal/pipeline — orchestrate Deck

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

parse → fetch (bounded in registry) → stats → flavor. Returns `[]Dependency`. No stdout, flags, or os.Exit. Dist-tags.latest then highest matching version when no resolved pin (v1: no lockfile). Fixture HTTP only in tests.

## Files

- Create `internal/pipeline/pipeline.go` (and `_test.go`)

## Interface

```go
package pipeline

type Options struct {
    Flavor    flavor.Mode
    Cache     *cache.Store
    Registry  *registry.Client // optional; tests inject
}

func Run(ctx context.Context, packageJSONPath string, opt Options) ([]types.Dependency, error)
func Check(ctx context.Context, packageJSONPath string, opt Options) (types.CheckData, error)
```

Run: invalid flavor `ai` returns flavor.ErrAINotInV1 (do not fetch). Parse error is a Go error (adapter → CLI exit). Per-package fetch failures stay on Dependency.Errors. Check uses stats.DefaultPolicy / Flagged; FailOnFetchErrors false; all-fetch-failed is still a successful CheckData with errors on every dep (CLI maps exit 3).

## Acceptance criteria

- [ ] Run with fixture npm: Dependencies match parser names; Rarity set on 200s
- [ ] One 404 among many: that Card has errors; others enriched; Run err == nil
- [ ] Flavor ai: error is ErrAINotInV1; zero HTTP
- [ ] testdata/workspace: only root names (via parser)
- [ ] No os.Stdout/os.Exit in package
- [ ] Tests never dial registry.npmjs.org

## Blocked by

- `.scratch/depdeck-v1/issues/03-pkg-parser.md`
- `.scratch/depdeck-v1/issues/04-pkg-registry.md`
- `.scratch/depdeck-v1/issues/05-pkg-stats.md`
- `.scratch/depdeck-v1/issues/06-pkg-flavor.md`

## ADRs

0001, 0002, 0003, 0005, 0011
