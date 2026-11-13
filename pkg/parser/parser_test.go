package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
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
