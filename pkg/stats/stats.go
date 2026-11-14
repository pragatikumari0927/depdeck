// Package stats derives npm risk facts and applies the v1 Check policy.
package stats

import (
	"time"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

const year = 365.25 * 24 * time.Hour

// Rarity maps weekly downloads to the v1 popularity tier.
func Rarity(weeklyDownloads int) types.Rarity {
	switch {
	case weeklyDownloads >= 10000:
		return types.RarityLegendary
	case weeklyDownloads >= 5000:
		return types.RarityEpic
	case weeklyDownloads >= 1000:
		return types.RarityRare
	default:
		return types.RarityCommon
	}
}

// Chaos maps publish age to the v1 maintenance score.
func Chaos(lastPublish time.Time, now time.Time) float64 {
	if lastPublish.IsZero() {
		return 1.0
	}
	age := now.Sub(lastPublish)
	if age < 2*year {
		return 0.0
	}
	if age < 5*year {
		return 0.6
	}
	return 1.0
}

// DefaultPolicy returns the v1 Check thresholds.
func DefaultPolicy() types.CheckPolicy {
	maxChaos := 0.5
	return types.CheckPolicy{
		ChaosScoreMax:     &maxChaos,
		RequireLicense:    true,
		FailOnFetchErrors: false,
	}
}

// Apply clears stale facts and adds the latest successful enrichment or errors.
func Apply(dep types.Dependency, weekly int, lastPublish, now time.Time, license string, fetchErrs []types.FetchError) types.Dependency {
	out := dep
	if len(fetchErrs) > 0 {
		out.Errors = fetchErrs
		out.Rarity = ""
		out.Chaos = nil
		out.WeeklyDownloads = nil
		out.License = ""
		out.LastPublish = ""
		out.FlavorText = ""
		out.SpecialMove = ""
		out.FlavorSource = ""
		return out
	}
	out.Errors = nil
	out.WeeklyDownloads = &weekly
	out.Rarity = Rarity(weekly)
	c := Chaos(lastPublish, now)
	out.Chaos = &c
	if lastPublish.IsZero() {
		out.LastPublish = ""
	} else {
		out.LastPublish = lastPublish.UTC().Format(time.RFC3339)
	}
	out.License = license
	return out
}

// Flagged returns dependencies that fail a known v1 policy check.
func Flagged(deps []types.Dependency, p types.CheckPolicy) (bool, []types.Dependency, string) {
	var flagged []types.Dependency
	for _, d := range deps {
		if len(d.Errors) > 0 {
			if p.FailOnFetchErrors {
				flagged = append(flagged, d)
			}
			continue
		}
		if p.ChaosScoreMax != nil && d.Chaos != nil && *d.Chaos > *p.ChaosScoreMax {
			flagged = append(flagged, d)
			continue
		}
		if p.RequireLicense && d.License == "" {
			flagged = append(flagged, d)
		}
	}
	return len(flagged) == 0, flagged, ""
}
