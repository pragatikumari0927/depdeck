package stats

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"depdeck/pkg/types"
)

func TestRarity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		weekly int
		want   types.Rarity
	}{
		{0, types.RarityCommon},
		{999, types.RarityCommon},
		{1000, types.RarityRare},
		{4999, types.RarityRare},
		{5000, types.RarityEpic},
		{9999, types.RarityEpic},
		{10000, types.RarityLegendary},
		{1_000_000, types.RarityLegendary},
	}
	for _, tc := range cases {
		if got := Rarity(tc.weekly); got != tc.want {
			t.Errorf("Rarity(%d) = %q, want %q", tc.weekly, got, tc.want)
		}
	}
}

func TestChaos(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		age  time.Duration
		want float64
	}{
		{0, 0.0},
		{time.Duration(1.9 * float64(year)), 0.0},
		{2 * year, 0.6},
		{time.Duration(4.9 * float64(year)), 0.6},
		{5 * year, 1.0},
		{time.Duration(5.1 * float64(year)), 1.0},
		{1 * year, 0.0},
		{3 * year, 0.6},
		{6 * year, 1.0},
	}
	for _, tc := range cases {
		got := Chaos(now.Add(-tc.age), now)
		if got != tc.want {
			t.Errorf("Chaos(age=%v) = %v, want %v", tc.age, got, tc.want)
		}
	}
	if got := Chaos(time.Time{}, now); got != 1.0 {
		t.Errorf("Chaos(zero) = %v, want 1.0", got)
	}
}

func TestDefaultPolicy(t *testing.T) {
	t.Parallel()
	p := DefaultPolicy()
	if p.ChaosScoreMax == nil || *p.ChaosScoreMax != 0.5 {
		t.Errorf("ChaosScoreMax = %v, want 0.5", p.ChaosScoreMax)
	}
	if p.AgeYearsMax != nil {
		t.Errorf("AgeYearsMax = %v, want nil", p.AgeYearsMax)
	}
	if !p.RequireLicense {
		t.Error("RequireLicense = false, want true")
	}
	if p.FailOnFetchErrors {
		t.Error("FailOnFetchErrors = true, want false")
	}
}

func TestApplyDegradedOmitsStats(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	errs := []types.FetchError{{Source: "npm", Kind: string(types.FetchNotFound), Detail: "missing"}}
	in := types.Dependency{Name: "left-pad", Version: "1.0.0", Tag: types.TagDependencies}
	got := Apply(in, 10000, now.Add(-year), now, "MIT", errs)
	if got.Rarity != "" {
		t.Errorf("Rarity = %q, want empty", got.Rarity)
	}
	if got.Chaos != nil {
		t.Errorf("Chaos = %v, want nil", got.Chaos)
	}
	if got.WeeklyDownloads != 0 {
		t.Errorf("WeeklyDownloads = %d, want 0", got.WeeklyDownloads)
	}
	if got.License != "" {
		t.Errorf("License = %q, want empty", got.License)
	}
	if got.LastPublish != "" {
		t.Errorf("LastPublish = %q, want empty", got.LastPublish)
	}
	if len(got.Errors) != 1 || got.Errors[0].Kind != string(types.FetchNotFound) {
		t.Errorf("Errors = %+v, want not_found kept", got.Errors)
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if strings.Contains(s, `"rarity"`) {
		t.Errorf("degraded JSON must omit rarity: %s", s)
	}
}

func TestApplySuccessSetsStats(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	pub := now.Add(-year)
	in := types.Dependency{Name: "react", Version: "^18.2.0", Tag: types.TagDependencies}
	got := Apply(in, 10000, pub, now, "MIT", nil)
	if got.WeeklyDownloads != 10000 {
		t.Errorf("WeeklyDownloads = %d, want 10000", got.WeeklyDownloads)
	}
	if got.Rarity != types.RarityLegendary {
		t.Errorf("Rarity = %q, want Legendary", got.Rarity)
	}
	if got.Chaos == nil || *got.Chaos != 0.0 {
		t.Errorf("Chaos = %v, want 0.0", got.Chaos)
	}
	if got.LastPublish != pub.UTC().Format(time.RFC3339) {
		t.Errorf("LastPublish = %q, want RFC3339", got.LastPublish)
	}
	if got.License != "MIT" {
		t.Errorf("License = %q, want MIT", got.License)
	}
	zeroPub := Apply(in, 0, time.Time{}, now, "MIT", nil)
	if zeroPub.LastPublish != "" {
		t.Errorf("zero lastPublish LastPublish = %q, want empty", zeroPub.LastPublish)
	}
	if zeroPub.Chaos == nil || *zeroPub.Chaos != 1.0 {
		t.Errorf("zero lastPublish Chaos = %v, want 1.0", zeroPub.Chaos)
	}
	if zeroPub.Rarity != types.RarityCommon {
		t.Errorf("zero downloads Rarity = %q, want Common", zeroPub.Rarity)
	}
}

func TestFlagged(t *testing.T) {
	t.Parallel()
	p := DefaultPolicy()
	chaos06 := 0.6

	licenseMissing := types.Dependency{Name: "nolicense", Version: "1.0.0", Tag: types.TagDependencies, License: ""}
	pass, flagged, _ := Flagged([]types.Dependency{licenseMissing}, p)
	if pass || !flaggedHas(flagged, "nolicense") {
		t.Errorf("empty license on success: pass=%v flagged=%v", pass, names(flagged))
	}

	degradedLicense := types.Dependency{
		Name:    "gone",
		Version: "1.0.0",
		Tag:     types.TagDependencies,
		Errors:  []types.FetchError{{Source: "npm", Kind: string(types.FetchNotFound), Detail: "404"}},
	}
	pass, flagged, _ = Flagged([]types.Dependency{degradedLicense}, p)
	if !pass || flaggedHas(flagged, "gone") {
		t.Errorf("degraded empty license: pass=%v flagged=%v", pass, names(flagged))
	}

	stale := types.Dependency{Name: "stale", Version: "1.0.0", Tag: types.TagDependencies, License: "MIT", Chaos: &chaos06}
	pass, flagged, _ = Flagged([]types.Dependency{stale}, p)
	if pass || !flaggedHas(flagged, "stale") {
		t.Errorf("chaos 0.6 > 0.5: pass=%v flagged=%v", pass, names(flagged))
	}

	degradedNilChaos := types.Dependency{
		Name:    "unknown",
		Version: "1.0.0",
		Tag:     types.TagDependencies,
		Errors:  []types.FetchError{{Source: "npm", Kind: string(types.FetchTimeout), Detail: "timeout"}},
	}
	pass, flagged, _ = Flagged([]types.Dependency{degradedNilChaos}, p)
	if !pass || flaggedHas(flagged, "unknown") {
		t.Errorf("degraded nil chaos: pass=%v flagged=%v", pass, names(flagged))
	}
}

func flaggedHas(deps []types.Dependency, name string) bool {
	for _, d := range deps {
		if d.Name == name {
			return true
		}
	}
	return false
}

func names(deps []types.Dependency) []string {
	out := make([]string, len(deps))
	for i, d := range deps {
		out[i] = d.Name
	}
	return out
}
