// Command browser-fixture writes a deterministic, offline Pixel Archive Deck
// for browser regression tests.
package main

import (
	"os"

	"github.com/pragatikumari0927/depdeck/pkg/render"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

func main() {
	deps := []types.Dependency{
		{
			Name: "runtime-legendary", Version: "^3.0.0", Tag: types.TagDependencies,
			WeeklyDownloads: intPtr(18000), Rarity: types.RarityLegendary,
			Chaos: floatPtr(0.0), LastPublish: "2026-10-01T00:00:00Z", License: "MIT",
			FlavorText: "A sample runtime Card with verified fixture facts.", SpecialMove: "Fixture Star",
		},
		{
			Name: "vite", Version: "~7.0.0", Tag: types.TagDependencies,
			WeeklyDownloads: intPtr(42), Rarity: types.RarityCommon,
			Chaos: floatPtr(0.6), LastPublish: "2025-09-12T00:00:00Z", License: "ISC",
			FlavorText: "A sample Card from the Runtime group.", SpecialMove: "Small Step",
		},
		{
			Name: "<script>alert('fixture')</script>", Version: "<img src=x onerror=alert(1)>", Tag: types.TagDependencies,
			Errors: []types.FetchError{{Source: "npm", Kind: types.FetchNetwork}},
		},
		{
			Name: "development-epic", Version: "^2.0.0", Tag: types.TagDevDependencies,
			WeeklyDownloads: intPtr(7200), Rarity: types.RarityEpic,
			Chaos: floatPtr(0.0), LastPublish: "2026-11-01T00:00:00Z", License: "MIT",
			FlavorText: "A sample Card from the Development group.", SpecialMove: "Build Spark",
		},
		{
			Name: "development-unknown", Version: "^1.0.0", Tag: types.TagDevDependencies,
			Errors: []types.FetchError{{Source: "npm", Kind: types.FetchTimeout}},
		},
		{
			Name: "optional-rare", Version: "^4.0.0", Tag: types.TagOptionalDependencies,
			WeeklyDownloads: intPtr(2400), Rarity: types.RarityRare,
			Chaos: floatPtr(0.6), LastPublish: "2025-05-08T00:00:00Z", License: "BSD-2-Clause",
			FlavorText: "A sample Card from the Optional group.", SpecialMove: "Side Quest",
		},
		{
			Name: "optional-unknown", Version: "^1.0.0", Tag: types.TagOptionalDependencies,
			Errors: []types.FetchError{{Source: "npm", Kind: types.FetchNotFound}},
		},
	}
	if err := render.Render(deps, os.Stdout); err != nil {
		_, _ = os.Stderr.WriteString("render browser fixture: " + err.Error() + "\n")
		os.Exit(1)
	}
}

func intPtr(value int) *int { return &value }

func floatPtr(value float64) *float64 { return &value }
