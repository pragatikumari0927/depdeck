package render

import (
	"crypto/sha256"
	_ "embed"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

//go:embed artwork/vite.svg
var viteSVG string

type artworkPath struct {
	D    string
	Fill string
}

type artworkSymbol struct {
	ID      string
	ViewBox string
	Paths   []artworkPath
}

type artworkRef struct {
	ID   string
	Kind string
}

// The only curated mark is keyed by its complete npm package identity.
var curatedArtwork = map[string]artworkSymbol{"vite": mustReadViteArtwork()}

func mustReadViteArtwork() artworkSymbol {
	dec := xml.NewDecoder(strings.NewReader(viteSVG))
	var symbol artworkSymbol
	rootSeen := false
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic("invalid embedded Vite SVG: " + err.Error())
		}
		switch node := tok.(type) {
		case xml.StartElement:
			switch node.Name.Local {
			case "svg":
				if rootSeen {
					panic("invalid embedded Vite SVG: nested svg")
				}
				rootSeen = true
				for _, attr := range node.Attr {
					switch attr.Name.Local {
					case "viewBox":
						if attr.Value != "0 0 23 14" {
							panic("invalid embedded Vite SVG: unexpected viewBox")
						}
						symbol.ViewBox = attr.Value
					case "xmlns":
						if attr.Value != "http://www.w3.org/2000/svg" {
							panic("invalid embedded Vite SVG: unexpected namespace")
						}
					default:
						panic("invalid embedded Vite SVG: unexpected root attribute " + attr.Name.Local)

					}
				}
			case "path":
				if !rootSeen {
					panic("invalid embedded Vite SVG: path before root")
				}
				path := artworkPath{}
				for _, attr := range node.Attr {
					switch attr.Name.Local {
					case "d":
						path.D = attr.Value
					case "fill":
						path.Fill = attr.Value
					default:
						panic("invalid embedded Vite SVG: unexpected path attribute " + attr.Name.Local)
					}
				}
				if path.D == "" || (path.Fill != "#FFFFFF" && path.Fill != "#863BFF") {
					panic("invalid embedded Vite SVG: unexpected path attributes")
				}
				symbol.Paths = append(symbol.Paths, path)
			default:
				panic("invalid embedded Vite SVG: unsupported element " + node.Name.Local)
			}
		case xml.CharData:
			if strings.TrimSpace(string(node)) != "" {
				panic("invalid embedded Vite SVG: unexpected text")
			}
		}
	}
	if !rootSeen || symbol.ViewBox == "" || len(symbol.Paths) != 3 {
		panic("invalid embedded Vite SVG: incomplete symbol")
	}
	symbol.ID = "art-vite"
	return symbol
}

func artworkFor(name string) (artworkRef, artworkSymbol) {
	if symbol, ok := curatedArtwork[name]; ok {
		return artworkRef{ID: symbol.ID, Kind: name}, symbol
	}
	digest := sha256.Sum256([]byte(name))
	id := fmt.Sprintf("art-fallback-%x", digest)
	return artworkRef{ID: id, Kind: "generated"}, artworkSymbol{
		ID: id, ViewBox: "0 0 16 16", Paths: []artworkPath{{D: pixelPath(digest), Fill: "#F1F5FB"}},
	}
}

// pixelPath draws a mirrored 8-by-8 motif from the complete package-name hash.
// It is original decoration and does not claim to describe package behavior.
func pixelPath(digest [32]byte) string {
	var path strings.Builder
	for y := 0; y < 8; y++ {
		for x := 0; x < 4; x++ {
			bit := byte(1 << uint(x))
			if digest[y]&bit == 0 {
				continue
			}
			for _, px := range []int{x * 2, (7 - x) * 2} {
				fmt.Fprintf(&path, "M%d %dh2v2h-2z", px, y*2)
			}
		}
	}
	if path.Len() == 0 {
		path.WriteString("M6 6h4v4H6z")
	}
	return path.String()
}
