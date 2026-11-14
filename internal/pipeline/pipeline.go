// Package pipeline assesses dependencies for both CLI and MCP.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/pragatikumari0927/depdeck/pkg/cache"
	"github.com/pragatikumari0927/depdeck/pkg/flavor"
	"github.com/pragatikumari0927/depdeck/pkg/parser"
	"github.com/pragatikumari0927/depdeck/pkg/registry"
	"github.com/pragatikumari0927/depdeck/pkg/stats"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

// Fetcher supplies npm facts for one declared dependency.
type Fetcher interface {
	Fetch(context.Context, string, string) (*types.Dependency, error)
}

// Options supplies the assessment dependencies and Flavor setting.
type Options struct {
	Flavor   flavor.Mode
	Cache    *cache.Store
	Registry Fetcher
	Now      func() time.Time
}

func (o Options) ready() (Options, error) {
	if err := o.Flavor.Validate(); err != nil {
		return o, err
	}
	if o.Flavor == "" {
		o.Flavor = flavor.Rule
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Registry == nil {
		if o.Cache == nil {
			var err error
			o.Cache, err = cache.New("")
			if err != nil {
				return o, err
			}
		}
		o.Registry = registry.New(o.Cache, nil)
	}
	return o, nil
}

// ManifestPath resolves a project directory or manifest file.
func ManifestPath(path string) (string, error) {
	if path == "" {
		path = "."
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat project path %s: %w", path, err)
	}
	if info.IsDir() {
		return filepath.Join(path, "package.json"), nil
	}
	return path, nil
}

// Run assesses one root Roster.
func Run(ctx context.Context, path string, opt Options) ([]types.Dependency, error) {
	opt, err := opt.ready()
	if err != nil {
		return nil, err
	}
	manifest, err := ManifestPath(path)
	if err != nil {
		return nil, err
	}
	deps, err := parser.ParseFile(manifest)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(deps, func(i, j int) bool {
		if deps[i].Name != deps[j].Name {
			return deps[i].Name < deps[j].Name
		}
		return deps[i].Tag < deps[j].Tag
	})
	now := opt.Now()
	jobs := make(chan int)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error
	workers := len(deps)
	if workers > 8 {
		workers = 8
	}
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				result, err := assess(ctx, deps[i], now, opt)
				if err != nil {
					once.Do(func() { firstErr = err })
					continue
				}
				deps[i] = result
			}
		}()
	}
	for i := range deps {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return deps, nil
}

func assess(ctx context.Context, original types.Dependency, now time.Time, opt Options) (result types.Dependency, err error) {
	// Run calls assess in worker goroutines, outside the Face's recovery boundary.
	defer func() {
		if recover() != nil {
			result = types.Dependency{}
			err = errors.New("dependency assessment failed")
		}
	}()
	facts, err := opt.Registry.Fetch(ctx, original.Name, original.Version)
	if err != nil {
		return original, fmt.Errorf("fetch %s: %w", original.Name, err)
	}
	if facts == nil {
		return original, fmt.Errorf("fetch %s: nil facts", original.Name)
	}
	var published time.Time
	if facts.LastPublish != "" {
		published, _ = time.Parse(time.RFC3339, facts.LastPublish)
	}
	weekly := 0
	if facts.WeeklyDownloads != nil {
		weekly = *facts.WeeklyDownloads
	}
	out := stats.Apply(original, weekly, published, now, facts.License, facts.Errors)
	if err := flavor.ApplyMode(&out, opt.Flavor); err != nil {
		return original, err
	}
	return out, nil
}

// Check assesses one Roster and applies the default v1 policy.
func Check(ctx context.Context, path string, opt Options) (types.CheckData, error) {
	deps, err := Run(ctx, path, opt)
	if err != nil {
		return types.CheckData{}, err
	}
	policy := stats.DefaultPolicy()
	pass, flagged, _ := stats.Flagged(deps, policy)
	if flagged == nil {
		flagged = []types.Dependency{}
	}
	reason := "pass"
	failed := 0
	for _, dep := range deps {
		if len(dep.Errors) > 0 {
			failed++
		}
	}
	switch {
	case len(deps) > 0 && failed == len(deps):
		reason, pass = "all_fetches_failed", false
		flagged = []types.Dependency{}
	case len(flagged) > 0:
		reason, pass = "policy_failed", false
	case failed > 0:
		reason = "partial_fetch_errors"
	case len(deps) == 0:
		reason = "empty_roster"
	}
	return types.CheckData{Pass: pass, Flagged: flagged, Reason: reason, Policy: policy, Dependencies: deps}, nil
}

// GetCard assesses one package without project membership.
func GetCard(ctx context.Context, name string, opt Options) (types.Dependency, error) {
	opt, err := opt.ready()
	if err != nil {
		return types.Dependency{}, err
	}
	return assess(ctx, types.Dependency{Name: name}, opt.Now(), opt)
}
