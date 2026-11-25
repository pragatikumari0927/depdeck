package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type missingRegistry struct{}

func (missingRegistry) Fetch(_ context.Context, name, version string) (*types.Dependency, error) {
	return &types.Dependency{Name: name, Version: version, Errors: []types.FetchError{{Source: "npm", Kind: types.FetchNotFound, Detail: "missing"}}}, nil
}

func TestJSONAndCheckExitWithInjectedRegistry(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"missing":"^1.0.0"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	opt := pipeline.Options{Registry: missingRegistry{}}
	var out, diagnostics bytes.Buffer
	if code := Run(context.Background(), []string{"scan", "--json", root}, opt, &out, &diagnostics); code != 0 {
		t.Fatalf("scan exit = %d; %s", code, diagnostics.String())
	}
	var envelope types.Envelope
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.SchemaVersion != "1.1" || envelope.Tool != "scan" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if _, err := os.Stat(filepath.Join(root, "deck.html")); !os.IsNotExist(err) {
		t.Fatalf("JSON scan wrote HTML: %v", err)
	}
	out.Reset()
	diagnostics.Reset()
	if code := Run(context.Background(), []string{"check", "--json", root}, opt, &out, &diagnostics); code != 3 {
		t.Fatalf("check exit = %d; %s", code, diagnostics.String())
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Tool != "check" {
		t.Fatalf("check envelope = %+v", envelope)
	}
}

func TestCLIRejectsAIAndConflictingOutputs(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := Run(context.Background(), []string{"scan", "--flavor=ai", "missing"}, pipeline.Options{Registry: missingRegistry{}}, &out, &diagnostics); code != 2 {
		t.Fatalf("ai exit = %d", code)
	}
	if !bytes.Contains(diagnostics.Bytes(), []byte("v1.1")) {
		t.Fatalf("ai diagnostic = %s", diagnostics.String())
	}
	out.Reset()
	diagnostics.Reset()
	if code := Run(context.Background(), []string{"scan", "--json", "--out", "deck.html"}, pipeline.Options{Registry: missingRegistry{}}, &out, &diagnostics); code != 2 {
		t.Fatalf("conflicting output exit = %d", code)
	}
}

func TestHelpAndUsageStreams(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		code   int
		useErr bool
		want   []string
		omit   string
	}{
		{"help", []string{"--help"}, 0, false, []string{"depdeck scan [flags] [path]", "package.json in the current directory", "Try depdeck scan to build your first Deck."}, ""},
		{"no_args", nil, 2, true, []string{"depdeck scan [flags] [path]", "depdeck mcp"}, ""},
		{"scan_help", []string{"scan", "--help"}, 0, true, []string{"Usage: depdeck scan [flags] [path]", "Put flags before the path.", "This mode skips the HTML Deck.", "deck.html beside package.json", "Special Moves", "Skip reading and writing the local npm cache."}, ""},
		{"check_help", []string{"check", "-h"}, 0, true, []string{"Usage: depdeck check [flags] [path]", "Write assessment JSON to standard output."}, "-out"},
		{"conflicting_outputs", []string{"scan", "--json", "--out", "-"}, 2, true, []string{"--json and --out cannot be used together.", "Use --json for assessment data or --out for HTML."}, ""},
		{"bad_flag", []string{"scan", "--unknown"}, 2, true, []string{"flag provided but not defined", "Usage: depdeck scan [flags] [path]"}, ""},
		{"flag_after_path", []string{"scan", ".", "--json"}, 2, true, []string{"Put flags before the path."}, ""},
		{"bad_command", []string{"scan\nother"}, 2, true, []string{`Unknown command "scan\nother". Run depdeck --help`}, ""},
		{"bad_flavor", []string{"scan", "--flavor=other"}, 2, true, []string{`invalid Flavor mode "other", use rule or none`}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(t.Context(), tc.args, pipeline.Options{Registry: missingRegistry{}}, &stdout, &stderr)
			output, other := stdout.String(), stderr.String()
			if tc.useErr {
				output, other = other, output
			}
			if code != tc.code || other != "" {
				t.Fatalf("exit = %d, stdout = %q, stderr = %q", code, &stdout, &stderr)
			}
			for _, snippet := range tc.want {
				if !strings.Contains(output, snippet) {
					t.Errorf("output missing %q: %q", snippet, output)
				}
			}
			if tc.omit != "" && strings.Contains(output, tc.omit) {
				t.Errorf("output contains %q: %q", tc.omit, output)
			}
		})
	}
}

func TestWarningsQuoteControlCharacters(t *testing.T) {
	var stderr bytes.Buffer
	warnings([]types.Dependency{{
		Name:   "pkg\x1b[2J\nforged",
		Errors: []types.FetchError{{Source: "npm", Kind: types.FetchParse}},
	}}, &stderr)
	if got := stderr.String(); got != "Warning: npm data unavailable for \"pkg\\x1b[2J\\nforged\".\n" {
		t.Fatalf("warning = %q", got)
	}
}
