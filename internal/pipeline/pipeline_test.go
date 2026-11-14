package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type fakeRegistry struct {
	mu    sync.Mutex
	calls map[string]int
}

func (f *fakeRegistry) Fetch(_ context.Context, name, declared string) (*types.Dependency, error) {
	f.mu.Lock()
	f.calls[name]++
	f.mu.Unlock()
	if name == "missing" {
		return &types.Dependency{Name: name, Version: declared, Errors: []types.FetchError{{Source: "npm", Kind: "not_found", Detail: "missing"}}}, nil
	}
	n := 0
	return &types.Dependency{Name: name, Version: declared, WeeklyDownloads: &n, License: "MIT", LastPublish: "2026-01-01T00:00:00Z"}, nil
}

func TestRunAndCheckShareOneAssessment(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"zeta":"^1.0.0","missing":"*"},"devDependencies":{"alpha":"2.0.0"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	f := &fakeRegistry{calls: map[string]int{}}
	opt := Options{Registry: f, Now: func() time.Time { return time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) }}
	deps, err := Run(context.Background(), root, opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 3 || deps[0].Name != "alpha" || deps[1].Name != "missing" || deps[2].Name != "zeta" {
		t.Fatalf("sorted roster = %+v", deps)
	}
	if deps[1].Version != "*" || deps[1].Tag != types.TagDependencies || len(deps[1].Errors) != 1 || deps[1].FlavorSource != types.FlavorRuleNameOnly {
		t.Fatalf("degraded identity or flavor = %+v", deps[1])
	}
	if deps[2].Version != "^1.0.0" || deps[2].WeeklyDownloads == nil || deps[2].Rarity != types.RarityCommon {
		t.Fatalf("enrichment = %+v", deps[2])
	}
	check, err := Check(context.Background(), root, opt)
	if err != nil {
		t.Fatal(err)
	}
	if !check.Pass || check.Reason != "partial_fetch_errors" || len(check.Dependencies) != 3 || len(check.Flagged) != 0 {
		t.Fatalf("check = %+v", check)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, name := range []string{"alpha", "missing", "zeta"} {
		if f.calls[name] != 2 {
			t.Fatalf("%s fetched %d times; want once per assessment", name, f.calls[name])
		}
	}
}

func TestAIRejectedBeforeManifestRead(t *testing.T) {
	f := &fakeRegistry{calls: map[string]int{}}
	_, err := Run(context.Background(), filepath.Join(t.TempDir(), "missing.json"), Options{Flavor: "ai", Registry: f})
	if err == nil {
		t.Fatal("ai mode must fail before manifest read")
	}
	if len(f.calls) != 0 {
		t.Fatalf("ai mode fetched %d packages", len(f.calls))
	}
}
