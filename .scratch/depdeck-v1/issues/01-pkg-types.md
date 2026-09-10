Status: ready-for-agent

# pkg/types — frozen contract

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

The shared Dependency, FetchError, envelope, Check payload, and enum constants. No I/O. CLI `--json` and MCP `data` use these types only.

## Files

- Create `pkg/types/types.go` (and `_test.go`)

## Interface

```go
package types

const SchemaVersion = "1.0"

type Tag string // "dependencies" | "devDependencies" | "optionalDependencies"
type Rarity string // "Common" | "Rare" | "Epic" | "Legendary"
type FlavorSource string // "rule_full" | "rule_name_only"
type FetchKind string // "not_found" | "timeout" | "rate_limit" | "network" | "parse"

type FetchError struct {
    Source string `json:"source"` // v1 always "npm"
    Kind   string `json:"kind"`
    Detail string `json:"detail"`
}

type Dependency struct {
    Name             string       `json:"name"`
    Version          string       `json:"version"`
    ResolvedVersion  string       `json:"resolved_version,omitempty"`
    Tag              Tag          `json:"tag"`
    WeeklyDownloads  int          `json:"weekly_downloads,omitempty"`
    Rarity           Rarity       `json:"rarity,omitempty"`
    Chaos            *float64     `json:"chaos,omitempty"`
    LastPublish      string       `json:"last_publish,omitempty"` // RFC3339 date or empty
    License          string       `json:"license,omitempty"`
    FlavorText       string       `json:"flavor_text,omitempty"`
    FlavorSource     FlavorSource `json:"flavor_source,omitempty"`
    Errors           []FetchError `json:"errors,omitempty"`
}

type Envelope struct {
    SchemaVersion string `json:"schema_version"`
    GeneratedAt   string `json:"generated_at"`
    GeneratedBy   string `json:"generated_by"`
    Tool          string `json:"tool"`
    Data          any    `json:"data"`
}

type DeckData struct {
    Dependencies []Dependency `json:"dependencies"`
}

type CheckPolicy struct {
    ChaosScoreMax      *float64 `json:"chaos_score_max"`
    AgeYearsMax        *float64 `json:"age_years_max"`
    RequireLicense     bool     `json:"require_license"`
    FailOnFetchErrors  bool     `json:"fail_on_fetch_errors"`
}

type CheckData struct {
    Pass    bool         `json:"pass"`
    Flagged []Dependency `json:"flagged"`
    Reason  string       `json:"reason"`
    Policy  CheckPolicy  `json:"policy"`
}
```

No `Stars` or `Issues` fields (GitHub is v2). `Chaos` pointer so 0.0 vs omitted is distinguishable if needed; omitempty on pointer nil. v1 `resolved_version` stays empty until a later lockfile ticket.

## Acceptance criteria

- [ ] JSON of a Dependency with empty Rarity, zero downloads, nil Chaos, empty Errors omits those keys (`omitempty`)
- [ ] FetchError with Source `"npm"` round-trips
- [ ] Envelope marshals `schema_version`, `generated_at`, `generated_by`, `tool`, `data`
- [ ] No GitHub-only fields on Dependency

## Blocked by

None - can start immediately

## ADRs

0001, 0002, 0003, 0005, 0007, 0008, 0009, 0010, 0011
