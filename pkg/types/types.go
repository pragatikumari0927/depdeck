package types

const SchemaVersion = "1.0"

type Tag string

const (
	TagDependencies         Tag = "dependencies"
	TagDevDependencies      Tag = "devDependencies"
	TagOptionalDependencies Tag = "optionalDependencies"
)

type Rarity string

const (
	RarityCommon    Rarity = "Common"
	RarityRare      Rarity = "Rare"
	RarityEpic      Rarity = "Epic"
	RarityLegendary Rarity = "Legendary"
)

type FlavorSource string

const (
	FlavorRuleFull     FlavorSource = "rule_full"
	FlavorRuleNameOnly FlavorSource = "rule_name_only"
)

type FetchKind string

const (
	FetchNotFound  FetchKind = "not_found"
	FetchTimeout   FetchKind = "timeout"
	FetchRateLimit FetchKind = "rate_limit"
	FetchNetwork   FetchKind = "network"
	FetchParse     FetchKind = "parse"
)

type FetchError struct {
	Source string `json:"source"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

type Dependency struct {
	Name            string       `json:"name"`
	Version         string       `json:"version"`
	ResolvedVersion string       `json:"resolved_version,omitempty"`
	Tag             Tag          `json:"tag"`
	WeeklyDownloads int          `json:"weekly_downloads,omitempty"`
	Rarity          Rarity       `json:"rarity,omitempty"`
	Chaos           *float64     `json:"chaos,omitempty"`
	LastPublish     string       `json:"last_publish,omitempty"`
	License         string       `json:"license,omitempty"`
	FlavorText      string       `json:"flavor_text,omitempty"`
	FlavorSource    FlavorSource `json:"flavor_source,omitempty"`
	Errors          []FetchError `json:"errors,omitempty"`
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
	ChaosScoreMax     *float64 `json:"chaos_score_max"`
	AgeYearsMax       *float64 `json:"age_years_max"`
	RequireLicense    bool     `json:"require_license"`
	FailOnFetchErrors bool     `json:"fail_on_fetch_errors"`
}

type CheckData struct {
	Pass    bool         `json:"pass"`
	Flagged []Dependency `json:"flagged"`
	Reason  string       `json:"reason"`
	Policy  CheckPolicy  `json:"policy"`
}
