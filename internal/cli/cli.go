// Package cli presents the shared assessment as local commands.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/cache"
	"github.com/pragatikumari0927/depdeck/pkg/flavor"
	"github.com/pragatikumari0927/depdeck/pkg/render"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

// Main runs a human-facing command and returns its process exit code.
func Main(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return Run(context.Background(), args, pipeline.Options{}, stdout, stderr)
}

// Run executes scan or check with the caller's context and assessment options.
func Run(ctx context.Context, args []string, opt pipeline.Options, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, "usage: depdeck scan|check [flags] [path] or depdeck mcp")
		return 2
	}
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		_, _ = fmt.Fprintln(stdout, "usage: depdeck scan|check [flags] [path] or depdeck mcp")
		return 0
	}
	if args[0] == "--version" || args[0] == "version" {
		_, _ = fmt.Fprintln(stdout, "depdeck "+types.BuildVersion)
		return 0
	}
	command := args[0]
	if command != "scan" && command != "check" {
		_, _ = fmt.Fprintf(stderr, "unknown command %q\n", command)
		return 2
	}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonMode := fs.Bool("json", false, "write JSON envelope")
	mode := fs.String("flavor", "rule", "rule or none")
	noCache := fs.Bool("no-cache", false, "bypass response cache")
	var output *string
	if command == "scan" {
		output = fs.String("out", "", "HTML output path, or - for stdout")
	}
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() > 1 {
		_, _ = fmt.Fprintln(stderr, "expected at most one path")
		return 2
	}
	if command == "scan" && *jsonMode {
		conflict := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "out" {
				conflict = true
			}
		})
		if conflict {
			_, _ = fmt.Fprintln(stderr, "--json and --out cannot be used together")
			return 2
		}
	}
	opt.Flavor = flavor.Mode(*mode)
	if err := opt.Flavor.Validate(); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if *noCache {
		if opt.Cache == nil {
			var err error
			opt.Cache, err = cache.New("")
			if err != nil {
				_, _ = fmt.Fprintln(stderr, err)
				return 2
			}
		}
		store := *opt.Cache
		store.Disabled = true
		opt.Cache = &store
	}
	path := "."
	if fs.NArg() == 1 {
		path = fs.Arg(0)
	}
	if command == "check" {
		result, err := pipeline.Check(ctx, path, opt)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		warnings(result.Dependencies, stderr)
		if *jsonMode {
			if err := json.NewEncoder(stdout).Encode(types.Wrap("check", result, time.Now())); err != nil {
				_, _ = fmt.Fprintln(stderr, err)
				return 2
			}
		} else {
			_, _ = fmt.Fprintf(stdout, "%s: %s (%d flagged)\n", verdict(result.Pass), result.Reason, len(result.Flagged))
		}
		if result.Reason == "all_fetches_failed" {
			return 3
		}
		if !result.Pass {
			return 1
		}
		return 0
	}
	target := *output
	var manifestInfo os.FileInfo
	if !*jsonMode && target != "-" {
		manifest, err := pipeline.ManifestPath(path)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		manifestInfo, err = os.Stat(manifest)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		path = manifest
		if target == "" {
			target = filepath.Join(filepath.Dir(manifest), "deck.html")
		}
		if err := checkOutputTarget(target, manifestInfo); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
	}
	deps, err := pipeline.Run(ctx, path, opt)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	warnings(deps, stderr)
	if *jsonMode {
		if err := json.NewEncoder(stdout).Encode(types.Wrap("scan", types.DeckData{Dependencies: deps}, time.Now())); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		return 0
	}
	if *output == "-" {
		if err := render.Render(deps, stdout); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		return 0
	}
	if err := writeHTML(target, manifestInfo, deps); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	_, _ = fmt.Fprintf(stderr, "Wrote %s (%d cards)\n", target, len(deps))
	return 0
}

func verdict(pass bool) string {
	if pass {
		return "pass"
	}
	return "fail"
}

func warnings(deps []types.Dependency, stderr io.Writer) {
	for _, d := range deps {
		if len(d.Errors) > 0 {
			_, _ = fmt.Fprintf(stderr, "Warning: npm facts unavailable for %q\n", d.Name)
		}
	}
}

func checkOutputTarget(target string, manifest os.FileInfo) error {
	info, err := os.Stat(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect output %q: %w", target, err)
	}
	if os.SameFile(manifest, info) {
		return fmt.Errorf("output %q would overwrite the input manifest", target)
	}
	return nil
}

func writeHTML(target string, manifest os.FileInfo, deps []types.Dependency) error {
	f, err := os.CreateTemp(filepath.Dir(target), ".depdeck-*.html")
	if err != nil {
		return fmt.Errorf("create output %s: %w", target, err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if err := render.Render(deps, f); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := checkOutputTarget(target, manifest); err != nil {
		return err
	}
	if err := replaceOutput(f.Name(), target); err != nil {
		return fmt.Errorf("replace output %s: %w", target, err)
	}
	return nil
}
