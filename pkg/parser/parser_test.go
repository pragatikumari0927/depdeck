package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

func writePkg(t *testing.T, dir, body string) string {
	t.Helper()
	p := filepath.Join(dir, "package.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func names(deps []types.Dependency) map[string]types.Dependency {
	m := make(map[string]types.Dependency, len(deps))
	for _, d := range deps {
		m[d.Name] = d
	}
	return m
}

func TestSingleDependency(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{
		"dependencies": { "react": "^18.2.0" }
	}`)
	got, err := ParseFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
	if got[0].Name != "react" || got[0].Version != "^18.2.0" || got[0].Tag != types.TagDependencies {
		t.Fatalf("got %+v", got[0])
	}
	if got[0].ResolvedVersion != "" {
		t.Fatalf("ResolvedVersion=%q want empty", got[0].ResolvedVersion)
	}
}

func TestParseFileReadsDirectResolutionsFromPackageLockV3(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifest := writePkg(t, dir, `{"dependencies":{"react":"^18.0.0","@scope/pkg":"~2.0.0"},"devDependencies":{"parent":"^1.0.0"}}`)
	lock := `{"name":"app","lockfileVersion":3,"packages":{"":{"name":"app"},"node_modules/react":{"version":"18.3.1"},"node_modules/@scope/pkg":{"version":"2.4.0"},"node_modules/parent":{"version":"1.2.0"},"node_modules/parent/node_modules/react":{"version":"17.0.0"},"node_modules/transitive":{"version":"9.9.9"}}}`
	if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	byName := names(got)
	if len(got) != 3 {
		t.Fatalf("lockfile changed direct Card count: %+v", got)
	}
	for name, want := range map[string]string{"react": "18.3.1", "@scope/pkg": "2.4.0", "parent": "1.2.0"} {
		if byName[name].ResolvedVersion != want {
			t.Errorf("%s resolved_version = %q, want %q", name, byName[name].ResolvedVersion, want)
		}
	}
	if byName["react"].Version != "^18.0.0" || byName["@scope/pkg"].Version != "~2.0.0" {
		t.Errorf("lockfile overwrote declared versions: %+v", got)
	}
	if _, ok := byName["transitive"]; ok {
		t.Fatal("lockfile added a transitive Card")
	}
}

func TestParseFileReadsDirectResolutionsFromPackageLockV1(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifest := writePkg(t, dir, `{"dependencies":{"react":"^18.0.0"}}`)
	lock := `{"name":"app","lockfileVersion":1,"dependencies":{"react":{"version":"18.2.0","requires":{"shared":"^1.0.0"}},"shared":{"version":"1.4.0"}}}`
	if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "react" || got[0].ResolvedVersion != "18.2.0" || got[0].Version != "^18.0.0" {
		t.Fatalf("parsed dependencies = %+v", got)
	}
}

func TestParseFileReadsDirectResolutionsFromPackageLockV2(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifest := writePkg(t, dir, `{"dependencies":{"react":"^18.0.0"}}`)
	lock := `{"name":"app","lockfileVersion":2,"packages":{"":{"name":"app"},"node_modules/react":{"version":"18.2.0"},"node_modules/parent/node_modules/react":{"version":"17.0.0"}},"dependencies":{"react":{"version":"18.2.0"}}}`
	if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ResolvedVersion != "18.2.0" || got[0].Version != "^18.0.0" {
		t.Fatalf("parsed dependencies = %+v", got)
	}
}

func TestParseFileIgnoresMalformedOrUnmatchedPackageLock(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "malformed", body: `{"lockfileVersion":3,"packages":`},
		{name: "unmatched", body: `{"lockfileVersion":3,"packages":{"node_modules/other":{"version":"9.9.9"}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			manifest := writePkg(t, dir, `{"dependencies":{"react":"^18.0.0"}}`)
			if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := ParseFile(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || got[0].ResolvedVersion != "" || got[0].Version != "^18.0.0" {
				t.Fatalf("invalid or unmatched lock changed parsed Roster: %+v", got)
			}
		})
	}
}

func TestParseFileIgnoresUnsupportedOrInvalidPackageLockEntries(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "unsupported version", body: `{"lockfileVersion":4,"packages":{"node_modules/react":{"version":"18.2.0"}}}`},
		{name: "null direct entry", body: `{"lockfileVersion":3,"packages":{"node_modules/react":null}}`},
		{name: "empty resolved version", body: `{"lockfileVersion":3,"packages":{"node_modules/react":{"version":""}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			manifest := writePkg(t, dir, `{"dependencies":{"react":"^18.0.0"}}`)
			if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := ParseFile(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || got[0].Version != "^18.0.0" || got[0].ResolvedVersion != "" {
				t.Fatalf("unsupported or invalid lock entry changed Roster: %+v", got)
			}
		})
	}
}

func TestParseFileOversizedPackageLockDoesNotBreakRoster(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifest := writePkg(t, dir, `{"dependencies":{"react":"^18.0.0"}}`)
	lockPath := filepath.Join(dir, "package-lock.json")
	f, err := os.Create(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(maxPackageLockSize + 1); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := ParseFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "react" || got[0].Version != "^18.0.0" || got[0].ResolvedVersion != "" {
		t.Fatalf("oversized lockfile changed parsed Roster: %+v", got)
	}
}

func TestPrecedenceDepsOverDev(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{
		"dependencies": { "lodash": "^4.0.0" },
		"devDependencies": { "lodash": "^3.0.0" }
	}`)
	got, err := ParseFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
	if got[0].Tag != types.TagDependencies || got[0].Version != "^4.0.0" {
		t.Fatalf("got %+v", got[0])
	}
}

func TestOptionalOnly(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{
		"optionalDependencies": { "fsevents": "^2.3.0" }
	}`)
	got, err := ParseFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Tag != types.TagOptionalDependencies {
		t.Fatalf("got %+v", got)
	}
}

func TestWorkspaceRootOnly(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "testdata", "workspace", "package.json")
	got, err := ParseFile(root)
	if err != nil {
		t.Fatal(err)
	}
	by := names(got)
	if _, ok := by["root-only"]; !ok {
		t.Fatalf("missing root-only: %+v", got)
	}
	if _, ok := by["nested-only"]; ok {
		t.Fatal("nested-only leaked from workspace walk")
	}
	for _, d := range got {
		if d.ResolvedVersion != "" {
			t.Fatalf("ResolvedVersion set on %s", d.Name)
		}
	}
}

func TestMalformedJSON(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{ not json`)
	if _, err := ParseFile(p); err == nil {
		t.Fatal("expected error")
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	if _, err := ParseFile(filepath.Join(t.TempDir(), "package.json")); err == nil {
		t.Fatal("expected error")
	}
}

func TestEmptyDeps(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{"name":"empty"}`)
	got, err := ParseFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("len=%d want 0", len(got))
	}
}

func TestNullDependencyVersionIsPreservedAsEmptyForRegistryValidation(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{"dependencies":{"null-version":null,"empty-version":"","whitespace-version":"  "}}`)
	got, err := ParseFile(p)
	if err != nil {
		t.Fatal(err)
	}
	byName := names(got)
	for _, name := range []string{"null-version", "empty-version", "whitespace-version"} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("version entry %q was lost: %+v", name, got)
		}
	}
	if byName["null-version"].Version != "" {
		t.Fatalf("null version = %q, want empty for registry validation", byName["null-version"].Version)
	}
}

func TestSkipPeerAndBundled(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writePkg(t, dir, `{
		"dependencies": { "keep": "1.0.0" },
		"peerDependencies": { "react": "^18.0.0" },
		"bundledDependencies": ["keep"]
	}`)
	got, err := ParseFile(p)
	if err != nil {
		t.Fatal(err)
	}
	by := names(got)
	if _, ok := by["react"]; ok {
		t.Fatal("peerDependencies emitted")
	}
	if len(got) != 1 || got[0].Name != "keep" {
		t.Fatalf("got %+v", got)
	}
}

func TestEmptyJSONObject(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{}`)
	got, err := ParseFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("len=%d want 0", len(got))
	}
}

func TestParseOrderStableByTagThenName(t *testing.T) {
	t.Parallel()
	p := writePkg(t, t.TempDir(), `{
		"dependencies": {"zeta":"1", "alpha":"2"},
		"devDependencies": {"test-z":"3", "test-a":"4"},
		"optionalDependencies": {"opt-z":"5", "opt-a":"6"}
	}`)
	for i := 0; i < 20; i++ {
		got, err := ParseFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, dep := range got {
			names = append(names, dep.Name)
		}
		want := []string{"alpha", "zeta", "test-a", "test-z", "opt-a", "opt-z"}
		if !reflect.DeepEqual(names, want) {
			t.Fatalf("ParseFile order = %v, want %v", names, want)
		}
	}
}

func TestDirectoryIsError(t *testing.T) {
	t.Parallel()
	if _, err := ParseFile(t.TempDir()); err == nil {
		t.Fatal("directory path must error")
	}
}

func TestOversizedManifestIsRejected(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "package.json")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(`{}`)); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if _, err := f.Write(bytes.Repeat([]byte(" "), 4<<20)); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseFile(p); err == nil {
		t.Fatal("oversized manifest must be rejected")
	}
}

func TestManifestAtSizeLimitIsAccepted(t *testing.T) {
	content := []byte(`{"dependencies":{"pkg":"1.0.0"}}`)
	content = append(content, bytes.Repeat([]byte(" "), maxManifestSize-len(content))...)
	p := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(p, content, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ParseFile(p)
	if err != nil || len(got) != 1 || got[0].Name != "pkg" {
		t.Fatalf("manifest at byte limit = %+v, error = %v", got, err)
	}
}

func FuzzParseFile(f *testing.F) {
	for _, seed := range []string{`{}`, `null`, `{"dependencies":{"pkg":"*"}}`, "{bad json", strings.Repeat(" ", 128)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > maxManifestSize+1 {
			t.Skip()
		}
		p := filepath.Join(t.TempDir(), "package.json")
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		deps, err := ParseFile(p)
		if err == nil {
			for i := 1; i < len(deps); i++ {
				if deps[i-1].Tag > deps[i].Tag || (deps[i-1].Tag == deps[i].Tag && deps[i-1].Name > deps[i].Name) {
					t.Fatalf("dependencies are not sorted: %+v", deps)
				}
			}
		}
	})
}
