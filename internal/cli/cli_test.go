package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type missingRegistry struct{}

func (missingRegistry) Fetch(_ context.Context, name, version string) (*types.Dependency, error) {
	return &types.Dependency{Name: name, Version: version, Errors: []types.FetchError{{Source: "npm", Kind: "not_found", Detail: "missing"}}}, nil
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

func TestWarningsQuoteControlCharacters(t *testing.T) {
	var stderr bytes.Buffer
	warnings([]types.Dependency{{
		Name:   "pkg\x1b[2J\nforged",
		Errors: []types.FetchError{{Source: "npm", Kind: "parse"}},
	}}, &stderr)
	if got := stderr.String(); got != "Warning: npm facts unavailable for \"pkg\\x1b[2J\\nforged\"\n" {
		t.Fatalf("warning = %q", got)
	}
}
