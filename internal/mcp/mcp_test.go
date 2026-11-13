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
	if len(listed.Tools) != 3 {
		t.Fatalf("tools = %+v", listed.Tools)
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
