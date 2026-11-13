package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/cache"
	"github.com/pragatikumari0927/depdeck/pkg/registry"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

const mixedRoster = `{"dependencies":{"fresh":"^1.0.0"},"devDependencies":{"missing":"^1.0.0"}}`

type wireEnvelope struct {
	SchemaVersion string          `json:"schema_version"`
	GeneratedAt   string          `json:"generated_at"`
	GeneratedBy   string          `json:"generated_by"`
	Tool          string          `json:"tool"`
	Data          json.RawMessage `json:"data"`
}

func TestReleaseBinaryOffline(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "depdeck.exe")
	build := exec.Command("go", "build", "-ldflags", "-X github.com/pragatikumari0927/depdeck/pkg/types.BuildVersion=1.0.0-test", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build depdeck: %v\n%s", err, output)
	}

	var outbound atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		outbound.Add(1)
		http.Error(w, "unexpected external request", http.StatusBadGateway)
	}))
	defer proxy.Close()
	home := t.TempDir()
	cacheBase := filepath.Join(home, "Library", "Caches")
	env := []string{
		"HOME=" + home, "USERPROFILE=" + home,
		"LOCALAPPDATA=" + cacheBase, "XDG_CACHE_HOME=" + cacheBase,
		"HTTP_PROXY=" + proxy.URL, "HTTPS_PROXY=" + proxy.URL, "NO_PROXY=",
	}
	t.Cleanup(func() {
		if count := outbound.Load(); count != 0 {
			t.Errorf("seeded-cache processes attempted %d outbound requests", count)
		}
	})
	store, err := cache.New(filepath.Join(cacheBase, "depdeck"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	put := func(source, name string, class cache.Class, status int, body string) {
		t.Helper()
		if err := store.Put(source, name, class, cache.Entry{Status: status, FetchedAt: now, Body: []byte(body)}); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"fresh", "unlicensed"} {
		license := "MIT"
		if name == "unlicensed" {
			license = ""
		}
		put("npm", name, cache.ClassOK, 200, packument(name, license, now))
		put("npm-downloads", name, cache.ClassOK, 200, `{"downloads":0}`)
	}
	put("npm", "missing", cache.ClassNotFound, 404, "")

	project := manifest(t, "project with spaces", mixedRoster)
	command := func(ctx context.Context, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.Dir = project
		cmd.Env = append(cmd.Environ(), env...)
		return cmd
	}
	t.Run("cli", func(t *testing.T) {
		testManifestOutputGuard(t, command, project)
		cases := []struct {
			name string
			args []string
			code int
			want string
		}{
			{"help", []string{"--help"}, 0, "depdeck mcp"},
			{"version", []string{"--version"}, 0, "depdeck 1.0.0-test"},
			{"usage", nil, 2, ""},
			{"mcp_usage", []string{"mcp", "extra"}, 2, ""},
			{"write_failure", []string{"scan", "--out", project}, 2, ""},
			{"invalid_manifest", []string{"scan", manifest(t, "invalid", "{")}, 2, ""},
			{"missing_manifest", []string{"scan", t.TempDir()}, 2, ""},
			{"null_manifest", []string{"check", manifest(t, "null", "null")}, 2, ""},
			{"ai", []string{"scan", "--flavor=ai"}, 2, ""},
			{"stdout_html", []string{"scan", "--out", "-"}, 0, "<!DOCTYPE html>"},
			{"empty", []string{"check", "--json", manifest(t, "empty", "{}")}, 0, "empty_roster"},
			{"flagged", []string{"check", "--json", manifest(t, "flagged", `{"dependencies":{"unlicensed":"*"}}`)}, 1, "policy_failed"},
			{"degraded", []string{"check", "--json", manifest(t, "degraded", `{"dependencies":{"missing":"*"}}`)}, 3, "all_fetches_failed"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				stdout, stderr := cliResult(t, command, tc.code, tc.args...)
				if tc.want != "" && !strings.Contains(stdout, tc.want) {
					t.Fatalf("stdout = %q, want %q", stdout, tc.want)
				}
				if tc.code == 2 && (stdout != "" || stderr == "") {
					t.Fatalf("stdout = %q, stderr = %q", stdout, stderr)
				}
			})
		}
		scan, stderr := cliResult(t, command, 0, "scan", "--json")
		if !strings.Contains(stderr, "missing") {
			t.Fatalf("missing npm warning: %q", stderr)
		}
		deck := decodeData[types.DeckData](t, []byte(scan), "scan", "1.0.0-test")
		if len(deck.Dependencies) != 2 || deck.Dependencies[0].WeeklyDownloads == nil ||
			*deck.Dependencies[0].WeeklyDownloads != 0 || len(deck.Dependencies[1].Errors) != 1 {
			t.Fatalf("scan data = %+v", deck)
		}
		check, _ := cliResult(t, command, 0, "check", "--json", filepath.Join(project, "package.json"))
		assessment := decodeData[types.CheckData](t, []byte(check), "check", "1.0.0-test")
		if !assessment.Pass || assessment.Reason != "partial_fetch_errors" || len(assessment.Dependencies) != 2 {
			t.Fatalf("check data = %+v", assessment)
		}
		target := filepath.Join(project, "deck.html")
		if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("JSON scan created HTML: %v", err)
		}
		_, stderr = cliResult(t, command, 0, "scan")
		html, err := os.ReadFile(target)
		if err != nil || !strings.Contains(string(html), "fresh") || !strings.Contains(stderr, "Wrote deck.html (2 cards)") {
			t.Fatalf("default Deck: %v, stderr = %q", err, stderr)
		}
	})

	t.Run("mcp", func(t *testing.T) {
		ctx, session := connectMCP(t, command)
		if got := session.InitializeResult().ServerInfo.Version; got != "1.0.0-test" {
			t.Fatalf("MCP server version = %q", got)
		}
		list, err := session.ListTools(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string]string{"check_deck": "path", "get_card": "name", "list_deck": "path"}
		for _, tool := range list.Tools {
			raw, err := json.Marshal(tool.InputSchema)
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Type                 string   `json:"type"`
				Required             []string `json:"required"`
				AdditionalProperties *bool    `json:"additionalProperties"`
			}
			key := expected[tool.Name]
			if err := json.Unmarshal(raw, &schema); err != nil || key == "" || schema.Type != "object" ||
				len(schema.Required) != 1 || schema.Required[0] != key || schema.AdditionalProperties == nil ||
				*schema.AdditionalProperties {
				t.Fatalf("unexpected tool %s schema %s: %v", tool.Name, raw, err)
			}
			delete(expected, tool.Name)
		}
		if len(expected) != 0 {
			t.Fatalf("missing tools: %v", expected)
		}

		data := callData(ctx, t, session, "list_deck", map[string]any{"path": ".", "tag": "dependencies", "flavor": "none"})
		deck := decodeData[types.DeckData](t, data, "list_deck", "1.0.0-test")
		if len(deck.Dependencies) != 1 || deck.Dependencies[0].Name != "fresh" ||
			deck.Dependencies[0].FlavorText != "" || deck.Dependencies[0].SpecialMove != "" {
			t.Fatalf("filtered Deck = %+v", deck)
		}
		data = callData(ctx, t, session, "get_card", map[string]any{"name": "fresh"})
		card := decodeData[map[string]json.RawMessage](t, data, "get_card", "1.0.0-test")
		for _, field := range []string{"version", "resolved_version", "tag"} {
			if _, present := card[field]; present {
				t.Fatalf("standalone Card includes %s: %s", field, data)
			}
		}
		if string(card["name"]) != `"fresh"` || string(card["flavor_source"]) != `"rule_full"` ||
			string(card["weekly_downloads"]) != "0" || string(card["license"]) != `"MIT"` {
			t.Fatalf("standalone Card = %s", data)
		}
		data = callData(ctx, t, session, "get_card", map[string]any{"name": "missing"})
		missing := decodeData[types.Card](t, data, "get_card", "1.0.0-test")
		if len(missing.Errors) != 1 || missing.Errors[0].Kind != "not_found" || missing.Errors[0].Source != "npm" {
			t.Fatalf("missing Card = %+v", missing)
		}
		data = callData(ctx, t, session, "check_deck", map[string]any{"path": "."})
		check := decodeData[types.CheckData](t, data, "check_deck", "1.0.0-test")
		if !check.Pass || check.Reason != "partial_fetch_errors" || len(check.Dependencies) != 2 {
			t.Fatalf("MCP Check = %+v", check)
		}
	})
}

func packument(name, license string, published time.Time) string {
	return fmt.Sprintf(`{"name":%q,"dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":%q}},"time":{"1.0.0":%q}}`,
		name, license, published.UTC().Format(time.RFC3339))
}

func manifest(t *testing.T, name, content string) string {
	t.Helper()
	project := filepath.Join(t.TempDir(), name)
	if err := os.Mkdir(project, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "package.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return project
}

func cliResult(t *testing.T, command func(context.Context, ...string) *exec.Cmd, want int, args ...string) (string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	var diagnostics bytes.Buffer
	cmd := command(ctx, args...)
	cmd.Stderr = &diagnostics
	output, err := cmd.Output()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("run %v: %v", args, err)
		}
		code = exit.ExitCode()
	}
	if code != want {
		t.Fatalf("%v exit = %d, want %d; stdout=%s stderr=%s", args, code, want, output, &diagnostics)
	}
	return string(output), diagnostics.String()
}

func connectMCP(t *testing.T, command func(context.Context, ...string) *exec.Cmd) (context.Context, *mcp.ClientSession) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	cmd := command(ctx, "mcp")
	var diagnostics bytes.Buffer
	cmd.Stderr = &diagnostics
	client := mcp.NewClient(&mcp.Implementation{Name: "depdeck-test", Version: "dev"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("initialize MCP: %v\nstderr: %s", err, &diagnostics)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Errorf("MCP shutdown: %v", err)
		}
		if cmd.ProcessState == nil || !cmd.ProcessState.Success() || diagnostics.Len() != 0 {
			t.Errorf("MCP EOF shutdown: %v; stderr: %s", cmd.ProcessState, &diagnostics)
		}
	})
	return ctx, session
}

func callData(ctx context.Context, t *testing.T, session *mcp.ClientSession, name string, args map[string]any) []byte {
	t.Helper()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil || result.IsError || len(result.Content) != 1 {
		t.Fatalf("%s = %+v, error = %v", name, result, err)
	}
	structured, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	var textData, structuredData any
	if !ok || json.Unmarshal([]byte(text.Text), &textData) != nil ||
		json.Unmarshal(structured, &structuredData) != nil || !reflect.DeepEqual(textData, structuredData) {
		t.Fatalf("%s text and structured content differ: %+v / %s", name, result.Content, structured)
	}
	return structured
}

func decodeData[T any](t *testing.T, raw []byte, tool, version string) T {
	t.Helper()
	var env wireEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	if env.SchemaVersion != types.SchemaVersion || env.Tool != tool || env.GeneratedBy != "depdeck "+version {
		t.Fatalf("unexpected envelope: %s", raw)
	}
	if _, err := time.Parse(time.RFC3339, env.GeneratedAt); err != nil || !strings.HasSuffix(env.GeneratedAt, "Z") {
		t.Fatalf("invalid generation time: %s", raw)
	}
	var data T
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode data in %s: %v", raw, err)
	}
	return data
}

// TestFixtureProcess runs the production dispatcher in a child with local endpoints.
func TestFixtureProcess(t *testing.T) {
	endpoint := os.Getenv("DEPDECK_TEST_FIXTURE_URL")
	if endpoint == "" {
		return
	}
	store, err := cache.New(os.Getenv("DEPDECK_TEST_CACHE"))
	if err != nil {
		t.Fatal(err)
	}
	store.Disabled = true
	client := registry.New(store, &http.Client{Transport: &http.Transport{}},
		registry.WithBaseURL(endpoint), registry.WithDownloadsURL(endpoint+"/downloads"))
	for i, arg := range os.Args {
		if arg == "--" {
			os.Exit(dispatch(t.Context(), os.Args[i+1:], pipeline.Options{Registry: client}))
		}
	}
	t.Fatal("missing child command")
}

func TestFixtureHTTPProcesses(t *testing.T) {
	var requests atomic.Int32
	started, cancelled := make(chan struct{}, 1), make(chan struct{}, 1)
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/fresh":
			_, _ = fmt.Fprint(w, packument("fresh", "MIT", time.Now()))
		case "/downloads/fresh":
			_, _ = fmt.Fprint(w, `{"downloads":10000}`)
		case "/rate":
			http.Error(w, "rate limited", http.StatusTooManyRequests)
		case "/broken":
			_, _ = fmt.Fprint(w, "{")
		case "/slow":
			started <- struct{}{}
			<-r.Context().Done()
			cancelled <- struct{}{}
		default:
			http.NotFound(w, r)
		}
	}))
	defer fixture.Close()

	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	project := manifest(t, "fixture project", mixedRoster)
	cacheDir := t.TempDir()
	command := func(ctx context.Context, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, binary, append([]string{"-test.run=^TestFixtureProcess$", "--"}, args...)...)
		cmd.Dir = project
		cmd.Env = append(cmd.Environ(), "DEPDECK_TEST_FIXTURE_URL="+fixture.URL, "DEPDECK_TEST_CACHE="+cacheDir)
		return cmd
	}

	raw, _ := cliResult(t, command, 0, "scan", "--json", "--no-cache")
	deck := decodeData[types.DeckData](t, []byte(raw), "scan", "dev")
	if len(deck.Dependencies) != 2 || deck.Dependencies[0].Rarity != types.RarityLegendary ||
		len(deck.Dependencies[1].Errors) != 1 || deck.Dependencies[1].Errors[0].Kind != "not_found" {
		t.Fatalf("unexpected fixture scan: %s", raw)
	}
	ctx, session := connectMCP(t, command)
	failed := manifest(t, "all unavailable", `{"dependencies":{"missing":"*","rate":"*","broken":"*"}}`)
	rawDeck := callData(ctx, t, session, "list_deck", map[string]any{"path": failed})
	deck = decodeData[types.DeckData](t, rawDeck, "list_deck", "dev")
	rawCheck := callData(ctx, t, session, "check_deck", map[string]any{"path": failed})
	check := decodeData[types.CheckData](t, rawCheck, "check_deck", "dev")
	if check.Pass || check.Reason != "all_fetches_failed" || len(check.Flagged) != 0 || len(check.Dependencies) != 3 || len(deck.Dependencies) != 3 {
		t.Fatalf("all-fetch-failed results: %s and %s", rawDeck, rawCheck)
	}
	for i, kind := range []string{"parse", "not_found", "rate_limit"} {
		for _, dep := range []types.Dependency{deck.Dependencies[i], check.Dependencies[i]} {
			if len(dep.Errors) != 1 || dep.Errors[0].Source != "npm" || dep.Errors[0].Kind != kind {
				t.Fatalf("expected %s fetch error, got %+v", kind, dep)
			}
		}
	}

	before := requests.Load()
	for _, test := range []struct {
		name string
		args any
	}{
		{"get_card", map[string]any{"name": "fresh", "flavor": "ai"}},
		{"get_card", map[string]any{"name": "fresh", "path": "."}},
		{"get_card", map[string]any{"name": "../bad"}},
		{"list_deck", map[string]any{"path": ".", "tag": "peers"}},
		{"list_deck", map[string]any{"path": ".", "flavor": nil}},
		{"list_deck", map[string]any{"path": 3}},
		{"check_deck", map[string]any{}},
		{"check_deck", map[string]any{"path": " "}},
	} {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: test.name, Arguments: test.args})
		if err != nil || !result.IsError {
			t.Fatalf("invalid arguments %s %v: %+v, %v", test.name, test.args, result, err)
		}
	}
	if result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "unknown"}); err == nil || result != nil {
		t.Fatalf("unknown tool: %+v, %v", result, err)
	}
	if requests.Load() != before {
		t.Fatal("invalid calls accessed fixture endpoints")
	}

	callCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := session.CallTool(callCtx, &mcp.CallToolParams{Name: "get_card", Arguments: map[string]any{"name": "slow"}})
		done <- err
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("slow fixture was never called")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled tool call: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("MCP call did not cancel")
	}
	select {
	case <-cancelled:
	case <-time.After(3 * time.Second):
		t.Fatal("MCP cancellation did not reach the fixture request")
	}
	callData(ctx, t, session, "get_card", map[string]any{"name": "fresh"})
}

func TestMCPRejectsBadFrames(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, frame, diagnostic string
	}{
		{"oversized", `{"jsonrpc":"2.0","method":"` + strings.Repeat("x", 16<<20) + "\"}\n", "maximum line length"},
		{"malformed", "{invalid}\n", "malformed JSON-RPC input"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "-test.run=^TestFixtureProcess$", "--", "mcp")
			cmd.Env = append(cmd.Environ(), "DEPDECK_TEST_FIXTURE_URL=http://127.0.0.1:1", "DEPDECK_TEST_CACHE="+t.TempDir())
			cmd.Stdin = strings.NewReader(tc.frame)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			stdout, err := cmd.Output()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 2 || !strings.Contains(stderr.String(), tc.diagnostic) {
				t.Fatalf("invalid frame: %v, stdout length %d, stderr %q", err, len(stdout), &stderr)
			}
			if tc.name == "oversized" {
				if len(stdout) != 0 {
					t.Fatalf("oversized frame wrote %d stdout bytes", len(stdout))
				}
				return
			}
			var response struct {
				Version string             `json:"jsonrpc"`
				ID      json.RawMessage    `json:"id"`
				Error   struct{ Code int } `json:"error"`
			}
			if err := json.Unmarshal(stdout, &response); err != nil ||
				response.Version != "2.0" || string(response.ID) != "null" || response.Error.Code != -32700 {
				t.Fatalf("JSON-RPC parse error = %q (%v)", stdout, err)
			}
		})
	}
}

func testManifestOutputGuard(t *testing.T, command func(context.Context, ...string) *exec.Cmd, project string) {
	t.Helper()
	input := filepath.Join(project, "package.json")
	original, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(project, "manifest alias.json")
	if err := os.Link(input, alias); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"package.json", input, alias} {
		stdout, stderr := cliResult(t, command, 2, "scan", "--out", target)
		if stdout != "" || !strings.Contains(stderr, "would overwrite the input manifest") {
			t.Fatalf("output guard for %q: stdout=%q, stderr=%q", target, stdout, stderr)
		}
	}
	for _, name := range []string{input, alias} {
		got, err := os.ReadFile(name)
		if err != nil || !bytes.Equal(got, original) {
			t.Fatalf("input or alias changed: %s, %v", name, err)
		}
	}
}
