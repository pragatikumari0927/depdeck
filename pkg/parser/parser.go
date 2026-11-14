// Package parser reads direct dependencies from one package.json manifest.
package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type manifest struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

const maxManifestSize = 4 << 20

// ParseFile reads a manifest file and returns direct dependencies in stable order.
func ParseFile(path string) ([]types.Dependency, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("parse %q: manifest must be a regular file", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	info, err = f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("parse %q: manifest must be a regular file", path)
	}
	if info.Size() > maxManifestSize {
		return nil, fmt.Errorf("parse %s: manifest exceeds %d byte limit", path, maxManifestSize)
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxManifestSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(raw) > maxManifestSize {
		return nil, fmt.Errorf("parse %s: manifest exceeds %d byte limit", path, maxManifestSize)
	}
	var m *manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if m == nil {
		return nil, fmt.Errorf("parse %q: manifest must be a JSON object", path)
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
	names := make([]string, 0, len(bucket))
	for name := range bucket {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, types.Dependency{
			Name:    name,
			Version: bucket[name],
			Tag:     tag,
		})
	}
	return out
}
