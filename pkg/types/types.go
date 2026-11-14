// Package types defines the shared JSON contract for depdeck.
package types

import "time"

// SchemaVersion is the first supported shipped contract version.
const SchemaVersion = "1.1"

// BuildVersion is set by the executable at build time.
var BuildVersion = "dev"

// Wrap makes the common CLI and MCP JSON envelope.
func Wrap(tool string, data any, now time.Time) Envelope {
	return Envelope{SchemaVersion: SchemaVersion, GeneratedAt: now.UTC().Format(time.RFC3339), GeneratedBy: "depdeck " + BuildVersion, Tool: tool, Data: data}
}

// Tag identifies the package.json section that lists a dependency.
type Tag string

// These values are the supported dependency sections.
const (
	TagDependencies         Tag = "dependencies"
	TagDevDependencies      Tag = "devDependencies"
	TagOptionalDependencies Tag = "optionalDependencies"
)

// Rarity is a popularity tier based on weekly downloads.
type Rarity string

// These values are the v1 popularity tiers.
const (
	RarityCommon    Rarity = "Common"
	RarityRare      Rarity = "Rare"
	RarityEpic      Rarity = "Epic"
	RarityLegendary Rarity = "Legendary"
)

// FlavorSource identifies the rule used to create Flavor text.
type FlavorSource string

// These values identify the v1 Flavor rules.
const (
	FlavorRuleFull     FlavorSource = "rule_full"
	FlavorRuleNameOnly FlavorSource = "rule_name_only"
)

// FetchKind identifies why an npm fact could not be fetched or parsed.
type FetchKind string

// These values are the supported v1 Fetch error kinds.
const (
	FetchNotFound  FetchKind = "not_found"
	FetchTimeout   FetchKind = "timeout"
	FetchRateLimit FetchKind = "rate_limit"
	FetchNetwork   FetchKind = "network"
	FetchParse     FetchKind = "parse"
)

// FetchError records a failure to read one npm fact.
type FetchError struct {
	Source string `json:"source"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

// Dependency contains a project's declared dependency and its npm facts.
type Dependency struct {
	Name            string       `json:"name"`
	Version         string       `json:"version"`
	ResolvedVersion string       `json:"resolved_version,omitempty"`
	Tag             Tag          `json:"tag"`
	WeeklyDownloads *int         `json:"weekly_downloads,omitempty"`
	Rarity          Rarity       `json:"rarity,omitempty"`
	Chaos           *float64     `json:"chaos,omitempty"`
	LastPublish     string       `json:"last_publish,omitempty"`
	License         string       `json:"license,omitempty"`
	FlavorText      string       `json:"flavor_text,omitempty"`
	SpecialMove     string       `json:"special_move,omitempty"`
	FlavorSource    FlavorSource `json:"flavor_source,omitempty"`
	Errors          []FetchError `json:"errors,omitempty"`
}

// Card contains the assessed facts for a package queried outside a Roster.
type Card struct {
	Name            string       `json:"name"`
	WeeklyDownloads *int         `json:"weekly_downloads,omitempty"`
	Rarity          Rarity       `json:"rarity,omitempty"`
	Chaos           *float64     `json:"chaos,omitempty"`
	LastPublish     string       `json:"last_publish,omitempty"`
	License         string       `json:"license,omitempty"`
	FlavorText      string       `json:"flavor_text,omitempty"`
	SpecialMove     string       `json:"special_move,omitempty"`
	FlavorSource    FlavorSource `json:"flavor_source,omitempty"`
	Errors          []FetchError `json:"errors,omitempty"`
}

// StandaloneCard projects an assessed Dependency without its Roster fields.
func (d Dependency) StandaloneCard() Card {
	return Card{
		Name: d.Name, WeeklyDownloads: d.WeeklyDownloads, Rarity: d.Rarity,
		Chaos: d.Chaos, LastPublish: d.LastPublish, License: d.License,
		FlavorText: d.FlavorText, SpecialMove: d.SpecialMove,
		FlavorSource: d.FlavorSource, Errors: d.Errors,
	}
}

// Envelope wraps CLI and MCP tool data.
type Envelope struct {
	SchemaVersion string `json:"schema_version"`
	GeneratedAt   string `json:"generated_at"`
	GeneratedBy   string `json:"generated_by"`
	Tool          string `json:"tool"`
	Data          any    `json:"data"`
}

// DeckData contains the dependencies in a project Deck.
type DeckData struct {
	Dependencies []Dependency `json:"dependencies"`
}

// CheckPolicy contains the thresholds used by a Check.
type CheckPolicy struct {
	ChaosScoreMax     *float64 `json:"chaos_score_max"`
	AgeYearsMax       *float64 `json:"age_years_max"`
	RequireLicense    bool     `json:"require_license"`
	FailOnFetchErrors bool     `json:"fail_on_fetch_errors"`
}

// CheckData contains the assessment and policy result.
type CheckData struct {
	Pass         bool         `json:"pass"`
	Flagged      []Dependency `json:"flagged"`
	Reason       string       `json:"reason"`
	Policy       CheckPolicy  `json:"policy"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
}
