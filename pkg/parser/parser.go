// Package parser reads direct dependencies from one package.json manifest.
package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type manifest struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

const maxManifestSize = 4 << 20
const maxPackageLockSize = 64 << 20

type packageLockHeader struct {
	LockfileVersion int             `json:"lockfileVersion"`
	Packages        json.RawMessage `json:"packages"`
	Dependencies    json.RawMessage `json:"dependencies"`
}

type packageLockEntry struct {
	Version string `json:"version"`
}

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
	resolved := packageLockVersions(filepath.Join(filepath.Dir(path), "package-lock.json"), out)
	for i := range out {
		out[i].ResolvedVersion = resolved[out[i].Name]
	}
	return out, nil
}

// packageLockVersions reads only direct package entries from supported npm
// package-lock formats. A missing, malformed, unsupported, or oversized lock
// file leaves resolved versions empty and does not prevent reading the Roster.
func packageLockVersions(path string, deps []types.Dependency) map[string]string {
	versions := make(map[string]string)
	if len(deps) == 0 {
		return versions
	}
	f, err := os.Open(path)
	if err != nil {
		return versions
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxPackageLockSize {
		return versions
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxPackageLockSize+1))
	if err != nil || len(raw) > maxPackageLockSize {
		return versions
	}
	var header packageLockHeader
	if err := json.Unmarshal(raw, &header); err != nil {
		return versions
	}
	switch header.LockfileVersion {
	case 1:
		var entries map[string]packageLockEntry
		if len(header.Dependencies) == 0 || json.Unmarshal(header.Dependencies, &entries) != nil {
			return versions
		}
		for _, dep := range deps {
			if entry, ok := entries[dep.Name]; ok && entry.Version != "" {
				versions[dep.Name] = entry.Version
			}
		}
	case 2, 3:
		var entries map[string]packageLockEntry
		if len(header.Packages) == 0 || json.Unmarshal(header.Packages, &entries) != nil {
			return versions
		}
		for _, dep := range deps {
			key := "node_modules/" + dep.Name
			if entry, ok := entries[key]; ok && entry.Version != "" {
				versions[dep.Name] = entry.Version
			}
		}
	}
	return versions
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
