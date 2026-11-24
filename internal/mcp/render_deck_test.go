package mcpface

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type renderFixtureRegistry struct{}

func (renderFixtureRegistry) Fetch(_ context.Context, name, version string) (*types.Dependency, error) {
	downloads := 1200
	chaos := 0.2
	return &types.Dependency{
		Name: name, Version: version, WeeklyDownloads: &downloads,
		LastPublish: "2026-11-01T00:00:00Z", License: "MIT", Chaos: &chaos,
	}, nil
}

func TestRenderDeckToolWritesPixelArchiveFromOfflineAssessment(t *testing.T) {
	ctx := context.Background()
	project := t.TempDir()
	manifest := filepath.Join(project, "package.json")
	if err := os.WriteFile(manifest, []byte(`{"dependencies":{"fixture-pkg":"^1.0.0"}}`), 0600); err != nil {
		t.Fatal(err)
	}

	clientSession, serverSession := connectTestSessions(t, pipeline.Options{
		Registry: renderFixtureRegistry{}, Now: func() time.Time { return time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC) },
	})
	defer func() { _ = clientSession.Close(); _ = serverSession.Wait() }()

	listed, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 4 {
		t.Fatalf("tool count = %d, want 4: %+v", len(listed.Tools), listed.Tools)
	}

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "render_deck", Arguments: map[string]any{"path": project},
	})
	if err != nil || result.IsError {
		t.Fatalf("render_deck result = %+v, error = %v", result, err)
	}
	encoded, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		SchemaVersion string `json:"schema_version"`
		Tool          string `json:"tool"`
		Data          struct {
			Output string `json:"output"`
			Cards  int    `json:"cards"`
			Bytes  int    `json:"bytes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.SchemaVersion != types.SchemaVersion || envelope.Tool != "render_deck" ||
		envelope.Data.Output != "deck.html" || envelope.Data.Cards != 1 || envelope.Data.Bytes <= 0 {
		t.Fatalf("render envelope = %s", encoded)
	}
	htmlBytes, err := os.ReadFile(filepath.Join(project, "deck.html"))
	if err != nil {
		t.Fatal(err)
	}
	if len(htmlBytes) != envelope.Data.Bytes {
		t.Fatalf("reported HTML size = %d, actual = %d", envelope.Data.Bytes, len(htmlBytes))
	}
	html := string(htmlBytes)
	for _, marker := range []string{"Rarity Atlas", "Field Journal", "fixture-pkg", "<title>depdeck"} {
		if !strings.Contains(html, marker) {
			t.Errorf("rendered Pixel Archive is missing %q", marker)
		}
	}
	if strings.Contains(html, "http://") || strings.Contains(html, "https://") ||
		strings.Contains(html, "registry.npmjs.org") || strings.Contains(html, "api.npmjs.org") {
		t.Fatal("generated Deck contains an external URL")
	}

	second, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "render_deck", Arguments: map[string]any{"path": project, "out": "copy.html"},
	})
	if err != nil || second.IsError {
		t.Fatalf("second render = %+v, error = %v", second, err)
	}
	secondHTML, err := os.ReadFile(filepath.Join(project, "copy.html"))
	if err != nil {
		t.Fatal(err)
	}
	if string(secondHTML) != html {
		t.Fatal("identical fixture assessments produced different HTML")
	}
	repeated, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "render_deck", Arguments: map[string]any{"path": project},
	})
	if err != nil || repeated == nil || !repeated.IsError {
		t.Fatalf("repeated render to existing default output = %+v, error = %v", repeated, err)
	}
	unchanged, err := os.ReadFile(filepath.Join(project, "deck.html"))
	if err != nil || string(unchanged) != html {
		t.Fatalf("repeated render changed the existing Deck: %v", err)
	}
}

func TestRenderDeckToolRefusesUnsafeOrExistingOutput(t *testing.T) {
	ctx := context.Background()
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "package.json"), []byte(`{"dependencies":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(filepath.Dir(project), "outside.html")
	if err := os.WriteFile(outside, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}

	clientSession, serverSession := connectTestSessions(t, pipeline.Options{Registry: renderFixtureRegistry{}})
	defer func() { _ = clientSession.Close(); _ = serverSession.Wait() }()

	for _, tc := range []struct {
		name string
		out  string
	}{
		{name: "absolute", out: outside},
		{name: "traversal", out: "../outside.html"},
		{name: "wrong_extension", out: "deck.txt"},
		{name: "missing_directory", out: "missing/deck.html"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
				Name: "render_deck", Arguments: map[string]any{"path": project, "out": tc.out},
			})
			if err != nil || result == nil || !result.IsError {
				t.Fatalf("unsafe output result = %+v, error = %v", result, err)
			}
		})
	}
	if err := os.Mkdir(filepath.Join(project, "directory.html"), 0700); err != nil {
		t.Fatal(err)
	}
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "render_deck", Arguments: map[string]any{"path": project, "out": "directory.html"},
	})
	if err != nil || result == nil || !result.IsError {
		t.Fatalf("filesystem output failure = %+v, error = %v", result, err)
	}

	outsideDir := t.TempDir()
	if err := os.Symlink(outsideDir, filepath.Join(project, "escape")); err == nil {
		t.Run("symlink_escape", func(t *testing.T) {
			result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
				Name: "render_deck", Arguments: map[string]any{"path": project, "out": "escape/deck.html"},
			})
			if err != nil || result == nil || !result.IsError {
				t.Fatalf("symlink output result = %+v, error = %v", result, err)
			}
			if _, err := os.Stat(filepath.Join(outsideDir, "deck.html")); !os.IsNotExist(err) {
				t.Fatalf("symlink target was written: %v", err)
			}
		})
	}

	if err := os.WriteFile(filepath.Join(project, "exists.html"), []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "render_deck", Arguments: map[string]any{"path": project, "out": "exists.html"},
	})
	if err != nil || result == nil || !result.IsError {
		t.Fatalf("existing output result = %+v, error = %v", result, err)
	}
	contents, err := os.ReadFile(filepath.Join(project, "exists.html"))
	if err != nil || string(contents) != "keep me" {
		t.Fatalf("existing output changed: %q, error = %v", contents, err)
	}
	contents, err = os.ReadFile(outside)
	if err != nil || string(contents) != "keep me" {
		t.Fatalf("outside file changed: %q, error = %v", contents, err)
	}
}

func TestMCPHandlesConcurrentAssessmentCalls(t *testing.T) {
	ctx := context.Background()
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "package.json"), []byte(`{"dependencies":{"fixture-pkg":"^1.0.0"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	clientSession, serverSession := connectTestSessions(t, pipeline.Options{Registry: renderFixtureRegistry{}})
	defer func() { _ = clientSession.Close(); _ = serverSession.Wait() }()

	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for i := range 16 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tool := "list_deck"
			args := map[string]any{"path": project}
			if i%2 == 1 {
				tool = "check_deck"
			}
			result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: args})
			if err != nil {
				errs <- err
				return
			}
			if result == nil || result.IsError {
				errs <- fmt.Errorf("%s returned an MCP error: %+v", tool, result)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func FuzzCleanDeckOutput(f *testing.F) {
	for _, seed := range []string{"", "deck.html", "../outside.html", `C:\private\deck.html`, "nested/deck.HTML", "name.txt", "a/../../b.html"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, requested string) {
		cleaned, err := cleanDeckOutput(requested)
		if err != nil {
			return
		}
		if !filepath.IsLocal(cleaned) || !strings.EqualFold(filepath.Ext(cleaned), ".html") {
			t.Fatalf("accepted output path is not local HTML: %q", cleaned)
		}
		again, err := cleanDeckOutput(cleaned)
		if err != nil || again != cleaned {
			t.Fatalf("output normalization is not stable: %q => %q, %v", cleaned, again, err)
		}
	})
}

func connectTestSessions(t *testing.T, opt pipeline.Options) (*mcp.ClientSession, *mcp.ServerSession) {
	t.Helper()
	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(opt).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "render-deck-test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	return clientSession, serverSession
}
