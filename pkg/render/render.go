// Package render builds the self-contained HTML dependency Deck.
package render

import (
	_ "embed"
	"html/template"
	"io"
	"sort"
	"time"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

//go:embed templates/deck.html
var deckHTML string

var deckTpl = template.Must(template.New("deck").Parse(deckHTML))

type rarityTier struct {
	Label       string
	Seal        string
	Downloads   string
	Explanation string
}

// The labels describe the existing assessment thresholds from ADR-0009.
var rarityTiers = [...]rarityTier{
	{string(types.RarityLegendary), "seal-legendary", "10,000 or more", "Highest weekly download band."},
	{string(types.RarityEpic), "seal-epic", "5,000 to 9,999", "Second-highest weekly download band."},
	{string(types.RarityRare), "seal-rare", "1,000 to 4,999", "Middle weekly download band."},
	{string(types.RarityCommon), "seal-common", "Fewer than 1,000 (known count)", "Includes verified zero downloads."},
	{"Unknown", "seal-unknown", "No verified rank", "Download data is unavailable or unverified."},
}

type atlasRow struct {
	rarityTier
	Count int
}

type deckCard struct {
	types.Dependency
	Tier      rarityTier
	TagLabel  string
	Published string
	Journal   [2]string
	Notices   []archiveNotice
}

type deckView struct {
	Cards     []deckCard
	Atlas     [len(rarityTiers)]atlasRow
	Legendary int
	Unknown   int
	TagCount  int
}

func rarityIndex(dep types.Dependency) int {
	unknown := len(rarityTiers) - 1
	if dep.WeeklyDownloads != nil && len(dep.Errors) == 0 {
		for i, tier := range rarityTiers[:unknown] {
			if tier.Label == string(dep.Rarity) {
				return i
			}
		}
	}
	return unknown
}

func tagLabel(tag types.Tag) string {
	switch tag {
	case types.TagDependencies:
		return "Runtime"
	case types.TagDevDependencies:
		return "Development"
	case types.TagOptionalDependencies:
		return "Optional"
	default:
		return ""
	}
}

func makeCard(dep types.Dependency, tier rarityTier) deckCard {
	card := deckCard{Dependency: dep, Tier: tier, TagLabel: tagLabel(dep.Tag), Published: dep.LastPublish}
	var journalDate string
	if published, err := time.Parse(time.RFC3339, dep.LastPublish); err == nil {
		journalDate = published.UTC().Format("02 Jan 2006")
		card.Published = journalDate
	}
	card.Journal = fieldJournal(dep, tier.Label, card.TagLabel, journalDate)
	card.Notices = archiveNotices(dep, tier.Label == "Unknown")
	return card
}

// Render writes deterministic, escaped Deck HTML to w.
func Render(deps []types.Dependency, w io.Writer) error {
	ordered := make([]types.Dependency, len(deps))
	copy(ordered, deps)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Name != ordered[j].Name {
			return ordered[i].Name < ordered[j].Name
		}
		return ordered[i].Tag < ordered[j].Tag
	})
	view := deckView{Cards: make([]deckCard, 0, len(ordered))}
	for i, tier := range rarityTiers {
		view.Atlas[i].rarityTier = tier
	}
	tags := make(map[types.Tag]struct{})
	for _, dep := range ordered {
		tier := rarityIndex(dep)
		view.Atlas[tier].Count++
		card := makeCard(dep, rarityTiers[tier])
		view.Cards = append(view.Cards, card)
		if card.TagLabel != "" {
			tags[dep.Tag] = struct{}{}
		}
		switch card.Tier.Label {
		case string(types.RarityLegendary):
			view.Legendary++
		case "Unknown":
			view.Unknown++
		}
	}
	view.TagCount = len(tags)
	return deckTpl.Execute(w, view)
}
