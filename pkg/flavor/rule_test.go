package flavor

import (
	"strings"
	"testing"

	"depdeck/pkg/types"
)

func forbidIssues(t *testing.T, s string) {
	t.Helper()
	low := strings.ToLower(s)
	if strings.Contains(low, "issues") || strings.Contains(low, "github") {
		t.Errorf("flavor mentions issues/GitHub: %q", s)
	}
}

func TestApplyTwiceIdentical(t *testing.T) {
	t.Parallel()
	in := types.Dependency{
		Name:            "react",
		WeeklyDownloads: 10000,
		Rarity:          types.RarityLegendary,
		Chaos:           ptr(0.0),
	}
	a := in
	b := in
	Apply(&a)
	Apply(&b)
	if a.FlavorText != b.FlavorText || a.SpecialMove != b.SpecialMove || a.FlavorSource != b.FlavorSource {
		t.Fatalf("same input diverged: %+v vs %+v", a, b)
	}
	forbidIssues(t, a.FlavorText)
	forbidIssues(t, a.SpecialMove)
}

func TestApplyNameOnly(t *testing.T) {
	t.Parallel()
	dep := types.Dependency{
		Name: "left-pad",
		Errors: []types.FetchError{
			{Source: "npm", Kind: string(types.FetchNotFound), Detail: "404"},
		},
	}
	Apply(&dep)
	if dep.FlavorSource != types.FlavorRuleNameOnly {
		t.Fatalf("FlavorSource = %q, want %q", dep.FlavorSource, types.FlavorRuleNameOnly)
	}
	if dep.FlavorText == "" || dep.SpecialMove == "" {
		t.Fatalf("name_only left empty text=%q move=%q", dep.FlavorText, dep.SpecialMove)
	}
	forbidIssues(t, dep.FlavorText)
	forbidIssues(t, dep.SpecialMove)
}

func TestApplyLegendary(t *testing.T) {
	t.Parallel()
	dep := types.Dependency{
		Name:            "lodash",
		WeeklyDownloads: 10000,
		Rarity:          types.RarityLegendary,
		Chaos:           ptr(0.0),
	}
	Apply(&dep)
	if dep.FlavorSource != types.FlavorRuleFull {
		t.Fatalf("FlavorSource = %q, want %q", dep.FlavorSource, types.FlavorRuleFull)
	}
	blob := dep.FlavorText + " " + dep.SpecialMove
	if !strings.Contains(blob, "Legendary") {
		t.Fatalf("legendary tier not hit: text=%q move=%q", dep.FlavorText, dep.SpecialMove)
	}
	forbidIssues(t, dep.FlavorText)
	forbidIssues(t, dep.SpecialMove)
}

func TestApplyHundredSameName(t *testing.T) {
	t.Parallel()
	seen := map[string]struct{}{}
	for i := 0; i < 100; i++ {
		dep := types.Dependency{
			Name:            "express",
			WeeklyDownloads: 5000,
			Rarity:          types.RarityEpic,
			Chaos:           ptr(0.6),
		}
		Apply(&dep)
		key := dep.FlavorText + "\x00" + dep.SpecialMove + "\x00" + string(dep.FlavorSource)
		seen[key] = struct{}{}
		forbidIssues(t, dep.FlavorText)
		forbidIssues(t, dep.SpecialMove)
	}
	if len(seen) != 1 {
		t.Fatalf("100 calls produced %d unique outputs, want 1", len(seen))
	}
}

func TestApplyZeroDownloadsStillFull(t *testing.T) {
	t.Parallel()
	dep := types.Dependency{Name: "tiny-pkg", WeeklyDownloads: 0}
	Apply(&dep)
	if dep.FlavorSource != types.FlavorRuleFull {
		t.Fatalf("FlavorSource = %q, want %q", dep.FlavorSource, types.FlavorRuleFull)
	}
	if dep.FlavorText == "" || dep.SpecialMove == "" {
		t.Fatalf("full left empty text=%q move=%q", dep.FlavorText, dep.SpecialMove)
	}
	forbidIssues(t, dep.FlavorText)
	forbidIssues(t, dep.SpecialMove)
}

func ptr(v float64) *float64 { return &v }
