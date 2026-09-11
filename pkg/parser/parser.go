package parser

import (
	"encoding/json"
	"fmt"
	"os"

	"depdeck/pkg/types"
)

type manifest struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

func ParseFile(path string) ([]types.Dependency, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	seen := map[string]struct{}{}
	var out []types.Dependency
	out = appendBucket(out, seen, m.Dependencies, types.TagDependencies)
	out = appendBucket(out, seen, m.DevDependencies, types.TagDevDependencies)
	out = appendBucket(out, seen, m.OptionalDependencies, types.TagOptionalDependencies)
	if out == nil {
		out = []types.Dependency{}
	}
	return out, nil
}

func appendBucket(out []types.Dependency, seen map[string]struct{}, bucket map[string]string, tag types.Tag) []types.Dependency {
	for name, version := range bucket {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, types.Dependency{
			Name:    name,
			Version: version,
			Tag:     tag,
		})
	}
	return out
}
