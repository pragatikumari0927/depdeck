package mcpface

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type missingRegistry struct{}

func (missingRegistry) Fetch(_ context.Context, name, version string) (*types.Dependency, error) {
	return &types.Dependency{Name: name, Version: version, Errors: []types.FetchError{{Source: "npm", Kind: "not_found", Detail: "missing"}}}, nil
}

func TestSDKToolsAndStandaloneCard(t *testing.T) {
	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	server := New(pipeline.Options{Registry: missingRegistry{}})
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = clientSession.Close(); _ = serverSession.Wait() }()
	listed, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	wantDescriptions := map[string]string{
		"list_deck":   "List assessment data for direct dependencies in a project's root package.json. Use path for the project directory or manifest, tag to select one dependency group, and flavor to choose rule or none. Results include Fetch errors when npm data is unavailable.",
		"get_card":    "Look up facts for one public npm package by name, such as react or @scope/package. Use flavor to choose rule or none. Call list_deck to check project membership. If npm data is unavailable, the Card includes a Fetch error.",
		"check_deck":  "Evaluate direct dependencies against the default publish-age and license-declaration policy. Returns the verdict, policy, flagged dependencies, and assessed dependencies with any Fetch errors. Chaos reflects publish age, and Rarity reflects weekly downloads. Use a vulnerability scanner for security findings.",
		"render_deck": "Create a self-contained Pixel Archive HTML Deck from a project's direct dependencies using depdeck's canonical renderer. Use when the user wants the visual Deck, rather than recreating HTML from assessment data. This may contact npm. Writes a new .html file under the project directory and never overwrites an existing file; out is an optional relative path and defaults to deck.html. Returns the relative output path, Card count, and byte size.",
	}
	if len(listed.Tools) != len(wantDescriptions) {
		t.Fatalf("tools = %+v", listed.Tools)
	}
	for _, tool := range listed.Tools {
		if want := wantDescriptions[tool.Name]; want == "" || tool.Description != want {
			t.Errorf("%s description = %q, want %q", tool.Name, tool.Description, want)
		}
		if tool.Name == "render_deck" {
			encoded, err := json.Marshal(tool.InputSchema)
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Properties           map[string]json.RawMessage `json:"properties"`
				Required             []string                   `json:"required"`
				AdditionalProperties *bool                      `json:"additionalProperties"`
			}
			if err := json.Unmarshal(encoded, &schema); err != nil {
				t.Fatal(err)
			}
			if len(schema.Properties) != 3 || len(schema.Required) != 1 || schema.Required[0] != "path" ||
				schema.AdditionalProperties == nil || *schema.AdditionalProperties {
				t.Errorf("render_deck schema = %s", encoded)
			}
			for _, property := range []string{"path", "out", "flavor"} {
				if _, ok := schema.Properties[property]; !ok {
					t.Errorf("render_deck schema omits %q: %s", property, encoded)
				}
			}
		}
	}
	for _, tc := range []struct {
		tool string
		args map[string]any
		want string
	}{
		{"list_deck", map[string]any{"path": " "}, "set path to a project directory or a package.json file"},
		{"check_deck", map[string]any{"path": " "}, "set path to a project directory or a package.json file"},
		{"render_deck", map[string]any{"path": " "}, "set path to a project directory or a package.json file"},
		{"get_card", map[string]any{"name": "../private"}, "use an npm package name such as react or @scope/package"},
		{"list_deck", map[string]any{"path": t.TempDir()}, "could not assess the project, check that path points to a readable package.json"},
		{"check_deck", map[string]any{"path": t.TempDir()}, "could not assess the project, check that path points to a readable package.json"},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: tc.tool, Arguments: tc.args})
			if err != nil || !result.IsError || len(result.Content) != 1 {
				t.Fatalf("%s = %+v, error = %v", tc.tool, result, err)
			}
			if text, ok := result.Content[0].(*mcp.TextContent); !ok || text.Text != tc.want {
				t.Errorf("%s returned %+v, want %q", tc.tool, result.Content, tc.want)
			}
		})
	}
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "get_card", Arguments: map[string]any{"name": "missing"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("npm 404 is a tool error: %+v", result)
	}
	b, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(b, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["tool"] != "get_card" || envelope["schema_version"] != "1.1" {
		t.Fatalf("envelope = %s", b)
	}
	card := envelope["data"].(map[string]any)
	if _, ok := card["version"]; ok {
		t.Fatalf("standalone Card has project version: %s", b)
	}
	if _, ok := card["tag"]; ok {
		t.Fatalf("standalone Card has project tag: %s", b)
	}
	if _, ok := card["errors"]; !ok {
		t.Fatalf("missing npm errors: %s", b)
	}
	bad, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "get_card", Arguments: map[string]any{"name": "missing", "flavor": "ai"}})
	if err != nil {
		t.Fatal(err)
	}
	if !bad.IsError {
		t.Fatalf("ai must be a tool error: %+v", bad)
	}
}

type panicRegistry struct{}

func (panicRegistry) Fetch(context.Context, string, string) (*types.Dependency, error) {
	panic("private path and transport details")
}

func TestPanicInAssessmentIsSanitized(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "package.json"), []byte(`{"dependencies":{"pkg":"*"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	server := New(pipeline.Options{Registry: panicRegistry{}})
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close(); _ = serverSession.Wait() }()

	for _, call := range []*mcp.CallToolParams{
		{Name: "get_card", Arguments: map[string]any{"name": "pkg"}},
		{Name: "list_deck", Arguments: map[string]any{"path": project}},
	} {
		result, err := session.CallTool(ctx, call)
		if err != nil || !result.IsError || len(result.Content) != 1 {
			t.Fatalf("%s = %+v, error = %v", call.Name, result, err)
		}
		content, ok := result.Content[0].(*mcp.TextContent)
		if !ok || strings.Contains(content.Text, "private") {
			t.Fatalf("%s returned panic details: %+v", call.Name, result.Content)
		}
	}
}
