package render

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/pragatikumari0927/depdeck/pkg/stats"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

func render(t *testing.T, deps []types.Dependency) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := Render(deps, &buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDeterminismSame(t *testing.T) {
	t.Parallel()
	deps := []types.Dependency{
		{Name: "zebra", Version: "1.0.0", Tag: types.TagDependencies},
		{Name: "apple", Version: "2.0.0", Tag: types.TagDevDependencies},
	}
	a := render(t, deps)
	b := render(t, deps)
	if !bytes.Equal(a, b) {
		t.Fatal("same slice produced different bytes")
	}
}

func TestSortTagTiebreaker(t *testing.T) {
	t.Parallel()
	opt := types.Dependency{Name: "dup", Version: "1.0.0", Tag: types.TagOptionalDependencies}
	prod := types.Dependency{Name: "dup", Version: "2.0.0", Tag: types.TagDependencies}
	fwd := render(t, []types.Dependency{opt, prod})
	rev := render(t, []types.Dependency{prod, opt})
	if !bytes.Equal(fwd, rev) {
		t.Fatal("same name different tag produced order-dependent bytes")
	}
	di := bytes.Index(fwd, []byte(`data-tag="dependencies"`))
	oi := bytes.Index(fwd, []byte(`data-tag="optionalDependencies"`))
	if di < 0 || oi < 0 || di > oi {
		t.Fatalf("tag order dependencies=%d optional=%d; want dependencies first", di, oi)
	}
}

func TestDeterminismShuffle(t *testing.T) {
	t.Parallel()
	zebra := types.Dependency{Name: "zebra", Version: "1.0.0", Tag: types.TagDependencies}
	apple := types.Dependency{Name: "apple", Version: "2.0.0", Tag: types.TagDevDependencies}
	fwd := render(t, []types.Dependency{zebra, apple})
	rev := render(t, []types.Dependency{apple, zebra})
	if !bytes.Equal(fwd, rev) {
		t.Fatal("shuffled slice produced different bytes")
	}
	ai := bytes.Index(fwd, []byte("apple"))
	zi := bytes.Index(fwd, []byte("zebra"))
	if ai < 0 || zi < 0 || ai > zi {
		t.Fatalf("apple index %d zebra index %d; want apple before zebra", ai, zi)
	}
	in := []types.Dependency{zebra, apple}
	_ = render(t, in)
	if in[0].Name != "zebra" || in[1].Name != "apple" {
		t.Fatal("Render mutated caller slice")
	}
}

func TestXSSName(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:    "foo<script>&bar",
		Version: "1.0.0",
		Tag:     types.TagDependencies,
	}})
	if !bytes.Contains(out, []byte("foo&lt;script&gt;&amp;bar")) {
		t.Fatal("escaped name missing")
	}
	if bytes.Contains(out, []byte("foo<script>")) {
		t.Fatal("raw name payload present")
	}

	alert := render(t, []types.Dependency{{
		Name:    "<script>alert(1)</script>",
		Version: "1.0.0",
		Tag:     types.TagDependencies,
	}})
	if bytes.Contains(alert, []byte("<script>alert(1)</script>")) {
		t.Fatal("raw alert payload present")
	}
	if !bytes.Contains(alert, []byte("&lt;script&gt;alert(1)&lt;/script&gt;")) {
		t.Fatal("escaped alert name missing")
	}
}

func TestXSSFlavor(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:       "safe",
		Version:    "1.0.0",
		Tag:        types.TagDependencies,
		FlavorText: "foo<script>&bar",
	}})
	if !bytes.Contains(out, []byte("foo&lt;script&gt;&amp;bar")) {
		t.Fatal("escaped flavor missing")
	}
	if bytes.Contains(out, []byte("foo<script>")) {
		t.Fatal("raw flavor payload present")
	}
}

func TestXSSMove(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:        "safe",
		Version:     "1.0.0",
		Tag:         types.TagDependencies,
		SpecialMove: "foo<script>&bar",
	}})
	if !bytes.Contains(out, []byte("foo&lt;script&gt;&amp;bar")) {
		t.Fatal("escaped special move missing")
	}
	if bytes.Contains(out, []byte("foo<script>")) {
		t.Fatal("raw special move payload present")
	}
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	for _, deps := range [][]types.Dependency{nil, {}} {
		var buf bytes.Buffer
		if err := Render(deps, &buf); err != nil {
			t.Fatal(err)
		}
		out := buf.Bytes()
		if !bytes.Contains(out, []byte("<!DOCTYPE html>")) {
			t.Fatal("missing doctype")
		}
		if bytes.Contains(out, []byte(`class="card`)) {
			t.Fatal("empty input emitted a card")
		}
	}
}

func TestDegraded(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:    "left-behind",
		Version: "0.0.1",
		Tag:     types.TagOptionalDependencies,
		Errors:  []types.FetchError{{Source: "npm", Kind: string(types.FetchTimeout)}},
	}})
	if !bytes.Contains(out, []byte("left-behind")) {
		t.Fatal("degraded card missing")
	}
	if !bytes.Contains(out, []byte("card--no-data")) {
		t.Fatal("missing card--no-data")
	}
	if !bytes.Contains(out, []byte("The registry took too long")) || bytes.Contains(out, []byte("timeout")) {
		t.Fatal("expected a plain-language notice without the raw error kind")
	}
}

func TestDownloadsShowsKnownZeroAndOmitsUnknown(t *testing.T) {
	t.Parallel()
	zero := 0
	known := render(t, []types.Dependency{{
		Name: "zero", Version: "1.0.0", Tag: types.TagDependencies, WeeklyDownloads: &zero,
	}})
	if !bytes.Contains(known, []byte("<dt>Weekly downloads</dt><dd>0</dd>")) {
		t.Fatalf("known zero downloads are not shown: %s", known)
	}
	unknown := render(t, []types.Dependency{{Name: "unknown", Version: "1.0.0", Tag: types.TagDependencies}})
	if !bytes.Contains(unknown, []byte("<dt>Weekly downloads</dt><dd class=\"unavailable\">Unavailable</dd>")) {
		t.Fatal("unknown downloads should show Unavailable")
	}
}

func TestCardRaritySeals(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		label     string
		seal      string
		downloads *int
		rarity    types.Rarity
	}{
		{"Common", "seal-common", new(0), types.RarityCommon},
		{"Rare", "seal-rare", new(1000), types.RarityRare},
		{"Epic", "seal-epic", new(5000), types.RarityEpic},
		{"Legendary", "seal-legendary", new(10000), types.RarityLegendary},
		{"Unknown", "seal-unknown", nil, ""},
	} {
		t.Run(tc.label, func(t *testing.T) {
			dep := types.Dependency{
				Name: "fixture", WeeklyDownloads: tc.downloads, Rarity: tc.rarity,
			}
			_, card, found := bytes.Cut(render(t, []types.Dependency{dep}), []byte(`<div class="deck">`))
			if !found {
				t.Fatal("missing Deck")
			}
			want := fmt.Sprintf(`<span class="rarity-label">%s</span><svg class="rarity-seal" viewBox="0 0 12 12" aria-hidden="true" focusable="false"><use href="#%s"></use></svg>`, tc.label, tc.seal)
			if !bytes.Contains(card, []byte(want)) {
				t.Fatalf("Card missing decorative %s seal and text label", tc.label)
			}
		})
	}
}

func TestArchiveNotices(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		kind        types.FetchKind
		title       string
		explanation string
		nextStep    string
	}{
		{types.FetchNotFound, "Registry data not found", "npm did not return a required record for this Card. The package, a compatible version, or its download count may be missing.", "Check the package name and declared version, or try another scan later."},
		{types.FetchNetwork, "The npm request failed", "Depdeck could not complete the request to npm.", "Check your connection or try again later."},
		{types.FetchTimeout, "The registry took too long", "npm did not respond in time.", "Try another scan."},
		{types.FetchRateLimit, "The registry needs a moment", "npm is limiting requests.", "Wait a moment before scanning again."},
		{types.FetchParse, "The record could not be read", "Depdeck could not read the registry response or interpret the declared version.", "Check the declared version, then run a new scan."},
		{`raw-unknown-kind"><script>unknown-kind</script>`, "Registry details unavailable", "Depdeck could not verify the registry details for this Card.", "Try another scan later."},
	} {
		t.Run(string(tc.kind), func(t *testing.T) {
			fetch := types.FetchError{
				Source: "private-host",
				Kind:   string(tc.kind),
				Detail: `Get "https://private.example": open C:\private\registry-token.json <script>raw-detail</script>`,
			}
			deps := []types.Dependency{{
				Name: "@scope/a-very-long-package-name", Version: "^1.0.0", Tag: types.TagDevDependencies,
				Errors: []types.FetchError{fetch, fetch},
			}}
			out := render(t, deps)
			if !bytes.Equal(out, render(t, deps)) {
				t.Fatal("repeated render changed the HTML")
			}
			notice := regexp.MustCompile(`(?s)<div class="notice" role="note" aria-label="Archive Notice">.*?</h4>\s*<p>.*?</p>`).Find(out)
			for _, want := range []string{
				`<span class="notice-crest" aria-hidden="true">!</span>`,
				"<h4>" + tc.title + "</h4>", "<p>" + tc.explanation + " " + tc.nextStep + "</p>",
			} {
				if !bytes.Contains(notice, []byte(want)) {
					t.Errorf("notice missing %q", want)
				}
			}
			if bytes.Count(out, []byte(`class="notice"`)) != 1 {
				t.Fatal("duplicate fetch errors should produce one notice")
			}
			for _, want := range []string{`data-rarity="Unknown"`, "Unavailable</dd>", "@scope/a-very-long-package-name"} {
				if !bytes.Contains(out, []byte(want)) {
					t.Errorf("degraded Card missing %q", want)
				}
			}
			for _, forbidden := range []string{
				"raw-detail", "private-host", "private.example", "registry-token.json", "unknown-kind",
				">" + string(tc.kind) + "<", "<button", "role=\"alert\"",
			} {
				haystack := out
				if forbidden == "<button" {
					haystack = notice
				}
				if bytes.Contains(haystack, []byte(forbidden)) {
					t.Errorf("unexpected %q in output", forbidden)
				}
			}
			if deps[0].Errors[0] != fetch {
				t.Fatal("renderer changed the structured fetch error")
			}
		})
	}
}

func TestDistinctArchiveNotices(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{Name: "fixture", Errors: []types.FetchError{
		{Kind: string(types.FetchTimeout)},
		{Kind: string(types.FetchNetwork)},
		{Kind: string(types.FetchTimeout)},
	}}})
	first := bytes.Index(out, []byte("<h4>The registry took too long</h4>"))
	second := bytes.Index(out, []byte("<h4>The npm request failed</h4>"))
	if bytes.Count(out, []byte(`class="notice"`)) != 2 || first < 0 || second <= first {
		t.Fatal("expected one note per distinct issue, in first-seen order")
	}
}

func TestCardStatsAndFlavor(t *testing.T) {
	t.Parallel()
	dep := types.Dependency{
		Name: "zero", Version: "^1.0.0", Tag: types.TagDependencies, WeeklyDownloads: new(0),
		Rarity: types.RarityCommon, Chaos: new(0.0), LastPublish: "2026-03-13T00:00:00Z",
		License: "MIT", FlavorText: "A fixture caption.", SpecialMove: "Fixture Move",
	}
	out := render(t, []types.Dependency{dep})
	for _, want := range []string{
		"<dt>Chaos</dt><dd>0</dd>", "<dt>Assessed version published</dt><dd>13 Mar 2026</dd>", "<dt>License</dt><dd>MIT</dd>",
		"npm recorded zero downloads", "Fixture Move", "A fixture caption.",
	} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("missing %q", want)
		}
	}
	if bytes.Contains(out, []byte("Archive Notice")) {
		t.Fatal("complete record showed a notice")
	}
	dep.License = ""
	out = render(t, []types.Dependency{dep})
	if !bytes.Contains(out, []byte("Some details are unavailable")) {
		t.Fatal("missing notice for incomplete statistics")
	}
}

func TestRarityAtlas(t *testing.T) {
	t.Parallel()
	fixtures := []struct {
		name        string
		seal        string
		downloads   string
		explanation string
	}{
		{"Legendary", "seal-legendary", "10,000 or more", "Highest weekly download band."},
		{"Epic", "seal-epic", "5,000 to 9,999", "Second-highest weekly download band."},
		{"Rare", "seal-rare", "1,000 to 4,999", "Middle weekly download band."},
		{"Common", "seal-common", "Fewer than 1,000 (known count)", "Includes verified zero downloads."},
		{"Unknown", "seal-unknown", "No verified rank", "Download data is unavailable or unverified."},
	}
	downloads := []int{0, 999, 1000, 4999, 5000, 9999, 10000}
	tags := []types.Tag{types.TagDependencies, types.TagDevDependencies, types.TagOptionalDependencies}
	deps := make([]types.Dependency, 0, len(downloads)+4)
	for i := range downloads {
		deps = append(deps, types.Dependency{
			Name: fmt.Sprintf("fixture-%d", i), Tag: tags[i%len(tags)],
			WeeklyDownloads: &downloads[i], Rarity: stats.Rarity(downloads[i]),
		})
	}
	deps = append(deps,
		types.Dependency{Name: "unverified", Tag: types.TagOptionalDependencies},
		types.Dependency{Name: "missing-count", Rarity: types.RarityCommon},
		types.Dependency{Name: "failed", Rarity: types.RarityCommon, WeeklyDownloads: &downloads[0],
			Errors: []types.FetchError{{Kind: string(types.FetchTimeout)}}},
		types.Dependency{Name: "hostile-tier", Rarity: types.Rarity(`"><script>alert(1)</script>`), WeeklyDownloads: &downloads[0]},
	)

	for _, tc := range []struct {
		name   string
		deps   []types.Dependency
		counts [5]int
	}{
		{"empty", nil, [5]int{}},
		{"one known zero", deps[:1], [5]int{0, 0, 0, 1, 0}},
		{"mixed", deps, [5]int{1, 2, 2, 2, 4}},
		{"all unknown", deps[len(downloads):], [5]int{0, 0, 0, 0, 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := render(t, tc.deps)
			if !bytes.Equal(out, render(t, tc.deps)) {
				t.Fatal("same collection produced different HTML")
			}
			atlas := regexp.MustCompile(`(?s)<section class="rarity-atlas panel".*?</section>`).Find(out)
			if !bytes.Contains(atlas, []byte("Counts cover all Cards in the Deck")) {
				t.Fatal("missing full-Deck count explanation")
			}
			if bytes.Count(atlas, []byte(`class="atlas-row"`)) != len(fixtures) {
				t.Fatal("expected five Atlas rows")
			}
			lastRow := -1
			for i, want := range fixtures {
				count := tc.counts[i]
				row := regexp.MustCompile(`(?s)<li class="atlas-row" data-rarity="` + want.name + `">.*?</li>`).Find(atlas)
				index := bytes.Index(atlas, row)
				if len(row) == 0 || index <= lastRow {
					t.Fatalf("missing or out-of-order Atlas row for %s", want.name)
				}
				lastRow = index
				label := "Cards"
				if count == 1 {
					label = "Card"
				}
				for _, text := range []string{want.downloads, want.explanation, fmt.Sprintf(">%d %s</span>", count, label), `href="#` + want.seal + `"`} {
					if !bytes.Contains(row, []byte(text)) {
						t.Errorf("%s Atlas row missing %q", want.name, text)
					}
				}
				if got := bytes.Count(out, []byte(`<symbol id="`+want.seal+`"`)); got != 1 {
					t.Errorf("%s symbol definitions = %d, want 1", want.name, got)
				}
				if got := bytes.Count(out, []byte(`href="#`+want.seal+`"`)); got != count*2+1 {
					t.Errorf("%s references = %d, want %d for the seal, illustration and Atlas", want.name, got, count*2+1)
				}
			}
			if bytes.Contains(out, []byte("alert(1)")) {
				t.Fatal("unrecognized Rarity leaked into rendered markup")
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()
	err := Render([]types.Dependency{{Name: "x", Version: "1", Tag: types.TagDependencies}}, errWriter{err: io.ErrClosedPipe})
	if err != io.ErrClosedPipe {
		t.Fatalf("err=%v want %v", err, io.ErrClosedPipe)
	}
}

func TestNoGeneratedAt(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{Name: "x", Version: "1", Tag: types.TagDependencies}})
	if bytes.Contains(out, []byte("generated_at")) {
		t.Fatal("generated_at in HTML")
	}
}

type errWriter struct {
	err error
}

func (w errWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestDeckSizeFor47Cards(t *testing.T) {
	deps := make([]types.Dependency, 47)
	downloads := []int{0, 1000, 5000, 10000}
	for i := range deps {
		weekly := &downloads[i%len(downloads)]
		deps[i] = types.Dependency{
			Name:    fmt.Sprintf("@fixture/package-%d-long-enough-to-wrap-on-mobile", i),
			Version: "^1.0.0", Tag: types.TagDependencies,
			WeeklyDownloads: weekly, Rarity: stats.Rarity(*weekly),
			License: "Apache-2.0", LastPublish: "2026-03-13T00:00:00Z",
			FlavorText:  "A dependable entry in the collection, with room to grow.",
			SpecialMove: "Dependency Cascade",
		}
		if i%5 == 0 {
			deps[i] = types.Dependency{
				Name: deps[i].Name, Version: deps[i].Version, Tag: types.TagOptionalDependencies,
				Errors: []types.FetchError{{Kind: string(types.FetchNotFound)}},
			}
		}
	}
	var out bytes.Buffer
	if err := Render(deps, &out); err != nil {
		t.Fatal(err)
	}
	t.Logf("47-card Deck: %d uncompressed bytes", out.Len())
	if out.Len() >= 150_000 {
		t.Fatalf("47-card Deck is %d bytes", out.Len())
	}
}
