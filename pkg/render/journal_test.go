package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

func TestFieldJournal(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		dep  types.Dependency
		want [2]string
	}{
		{
			"legendary recent",
			types.Dependency{Tag: types.TagDependencies, Rarity: types.RarityLegendary, WeeklyDownloads: new(12800),
				LastPublish: "2026-03-13T00:00:00Z", Chaos: new(0.0), License: "MIT"},
			[2]string{
				"npm recorded 12800 downloads for the assessed week. This entry carries the Legendary seal.",
				"Filed with the runtime dependencies. The assessed version has a recorded publish date of 13 Mar 2026. This date falls in the most recent publish-age band. The npm record lists MIT as the license.",
			},
		},
		{
			"epic middle age band",
			types.Dependency{Tag: types.TagDevDependencies, Rarity: types.RarityEpic, WeeklyDownloads: new(5500),
				LastPublish: "2023-04-01T00:00:00Z", Chaos: new(0.6), License: "Apache-2.0"},
			[2]string{
				"npm recorded 5500 downloads for the assessed week. This entry carries the Epic seal.",
				"Filed with the development dependencies. The assessed version has a recorded publish date of 01 Apr 2023, two to under five years before assessment. The npm record lists Apache-2.0 as the license.",
			},
		},
		{
			"rare oldest age band",
			types.Dependency{Tag: types.TagOptionalDependencies, Rarity: types.RarityRare, WeeklyDownloads: new(2400),
				LastPublish: "2019-01-01T00:00:00Z", Chaos: new(1.0)},
			[2]string{
				"npm recorded 2400 downloads for the assessed week. This entry carries the Rare seal.",
				"Filed with the optional dependencies. The assessed version has a recorded publish date of 01 Jan 2019, at least five years before assessment. The npm record lists no license.",
			},
		},
		{
			"common known zero and missing date",
			types.Dependency{Tag: types.TagDependencies, Rarity: types.RarityCommon, WeeklyDownloads: new(0), Chaos: new(1.0)},
			[2]string{
				"npm recorded zero downloads for the assessed week. This entry carries the Common seal.",
				"Filed with the runtime dependencies. The assessed version has no verified publish date. The npm record lists no license.",
			},
		},
		{
			"common singular count",
			types.Dependency{Rarity: types.RarityCommon, WeeklyDownloads: new(1), LastPublish: "2024-10-02T00:00:00Z"},
			[2]string{
				"npm recorded one download for the assessed week. This entry carries the Common seal.",
				"The dependency group is unavailable. The assessed version has a recorded publish date of 02 Oct 2024. The npm record lists no license.",
			},
		},
		{
			"unknown without downloads",
			types.Dependency{Tag: types.TagDevDependencies, LastPublish: "invalid date", License: "ISC"},
			[2]string{
				"The seal is Unknown because the weekly download count is unavailable.",
				"Filed with the development dependencies. The assessed version has no verified publish date. The npm record lists ISC as the license.",
			},
		},
		{
			"unknown with downloads",
			types.Dependency{Tag: types.TagDependencies, WeeklyDownloads: new(17)},
			[2]string{
				"npm recorded 17 downloads for the assessed week. The seal is Unknown because the rank could not be verified.",
				"Filed with the runtime dependencies. The assessed version has no verified publish date. The npm record lists no license.",
			},
		},
		{
			"degraded with stale fields",
			types.Dependency{Tag: types.TagOptionalDependencies, WeeklyDownloads: new(0), Rarity: types.RarityCommon,
				LastPublish: "2019-01-01T00:00:00Z", License: "stale-license",
				Errors: []types.FetchError{{Kind: string(types.FetchNetwork)}}},
			[2]string{
				"The seal is Unknown. This scan could not verify weekly downloads.",
				"Filed with the optional dependencies. The publish date and license could not be verified.",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deps := []types.Dependency{tc.dep}
			out := render(t, deps)
			want := []byte("<p>" + tc.want[0] + "</p><p>" + tc.want[1] + "</p>")
			if !bytes.Contains(journalSection(t, out), want) {
				t.Fatalf("unexpected Journal: %s", journalSection(t, out))
			}
			if !bytes.Equal(out, render(t, deps)) {
				t.Fatal("repeat render changed")
			}
		})
	}
}

func TestFieldJournalLongAndHostileText(t *testing.T) {
	t.Parallel()
	flavor := strings.Repeat("A fixture caption. ", 40) + "<script>flavor</script>"
	dep := types.Dependency{
		Tag: types.TagDependencies, WeeklyDownloads: new(10), Rarity: types.RarityCommon,
		FlavorText: flavor, License: "MIT <script>license</script>",
	}
	journal := journalSection(t, render(t, []types.Dependency{dep}))
	for _, want := range []string{
		strings.Repeat("A fixture caption. ", 40) + "&lt;script&gt;flavor&lt;/script&gt;",
		"MIT &lt;script&gt;license&lt;/script&gt;",
	} {
		if !bytes.Contains(journal, []byte(want)) {
			t.Errorf("missing escaped text: %q", want)
		}
	}
	if bytes.Contains(journal, []byte("<script>")) {
		t.Fatal("Journal contained raw markup")
	}

	dep.License = strings.Repeat("MIT OR ", 30) + "ISC"
	out := render(t, []types.Dependency{dep})
	journal = journalSection(t, out)
	if !bytes.Contains(journal, []byte("See the License field for the full npm declaration.")) ||
		bytes.Contains(journal, []byte(dep.License)) || !bytes.Contains(out, []byte("<dd>"+dep.License+"</dd>")) {
		t.Fatal("long declaration should appear in the License field")
	}
}

func journalSection(t *testing.T, out []byte) []byte {
	t.Helper()
	_, after, found := bytes.Cut(out, []byte(`<section class="journal">`))
	if !found {
		t.Fatal("missing Journal")
	}
	journal, _, found := bytes.Cut(after, []byte("</section>"))
	if !found {
		t.Fatal("unclosed Journal")
	}
	return journal
}
