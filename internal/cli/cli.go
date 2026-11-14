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

const usage = `Usage:
  depdeck scan [flags] [path]   Build an HTML Deck or write JSON.
  depdeck check [flags] [path]  Check direct dependencies against the default policy.
  depdeck mcp                   Start the MCP server over standard input and output.

If you omit path, depdeck reads package.json in the current directory.
Run depdeck scan --help or depdeck check --help for flags.
`

// Main runs a human-facing command and returns its process exit code.
func Main(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return Run(context.Background(), args, pipeline.Options{}, stdout, stderr)
}

// Run executes scan or check with the caller's context and assessment options.
func Run(ctx context.Context, args []string, opt pipeline.Options, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		_, _ = fmt.Fprint(stdout, usage)
		return 0
	}
	if args[0] == "--version" || args[0] == "version" {
		_, _ = fmt.Fprintln(stdout, "depdeck "+types.BuildVersion)
		return 0
	}
	command := args[0]
	if command != "scan" && command != "check" {
		_, _ = fmt.Fprintf(stderr, "Unknown command %q. Run depdeck --help for available commands.\n", command)
		return 2
	}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "Usage: depdeck %s [flags] [path]\n", command)
		_, _ = fmt.Fprintln(fs.Output(), "Use a project directory or the path to package.json. Put flags before the path.")
		fs.PrintDefaults()
	}
	jsonHelp := "Write assessment JSON to standard output."
	var output *string
	if command == "scan" {
		jsonHelp += " This mode skips the HTML Deck."
		output = fs.String("out", "", "Write the Deck to this file. The default is deck.html beside package.json. Use - for standard output.")
	}
	jsonMode := fs.Bool("json", false, jsonHelp)
	mode := fs.String("flavor", "rule", "Use rule for fixed Card text and Special Moves, or none to omit both.")
	noCache := fs.Bool("no-cache", false, "Skip reading and writing the local npm cache.")
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() > 1 {
		_, _ = fmt.Fprintln(stderr, "Use one project directory or package.json path. Put flags before the path.")
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
			_, _ = fmt.Fprintln(stdout, checkSummary(result))
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

func checkSummary(result types.CheckData) string {
	switch result.Reason {
	case "pass":
		return "Check passed. No dependencies met the policy's flagging criteria."
	case "empty_roster":
		return "Check passed. No direct dependencies to assess."
	case "partial_fetch_errors":
		return "Check passed for assessed dependencies. npm data was unavailable for some dependencies."
	case "policy_failed":
		if len(result.Flagged) == 1 {
			return "Check failed. The policy flagged 1 dependency."
		}
		return fmt.Sprintf("Check failed. The policy flagged %d dependencies.", len(result.Flagged))
	case "all_fetches_failed":
		return "Check failed. npm data was unavailable for every dependency."
	}
	verdict := "failed"
	if result.Pass {
		verdict = "passed"
	}
	return fmt.Sprintf("Check %s. Flagged dependencies: %d.", verdict, len(result.Flagged))
}

func warnings(deps []types.Dependency, stderr io.Writer) {
	for _, d := range deps {
		if len(d.Errors) > 0 {
			_, _ = fmt.Fprintf(stderr, "Warning: Could not assess %q using npm data. Its Card has a Fetch error.\n", d.Name)
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
