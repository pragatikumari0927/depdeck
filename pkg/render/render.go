// Package render builds the self-contained HTML dependency Deck.
package render

import (
	_ "embed"
	"html/template"
	"io"
	"sort"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

//go:embed templates/deck.html
var deckHTML string

var deckTpl = template.Must(template.New("deck").Parse(deckHTML))

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
	return deckTpl.Execute(w, types.DeckData{Dependencies: ordered})
}
