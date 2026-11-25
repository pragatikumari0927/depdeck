package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pragatikumari0927/depdeck/pkg/cache"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "testdata", "registry", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func testStore(t *testing.T) *cache.Store {
	t.Helper()
	s, err := cache.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func testClient(t *testing.T, srv *httptest.Server, store *cache.Store) *Client {
	t.Helper()
	return New(store, srv.Client(),
		WithBaseURL(srv.URL),
		WithDownloadsURL(srv.URL+"/downloads/point/last-week"),
	)
}

func muxOK(t *testing.T, packument, downloads []byte, etag string, hits *atomic.Int32) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			hits.Add(1)
		}
		if len(r.URL.Path) >= len("/downloads/point/last-week/") &&
			r.URL.Path[:len("/downloads/point/last-week/")] == "/downloads/point/last-week/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(downloads)
			return
		}
		if etag != "" {
			w.Header().Set("ETag", etag)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(packument)
	})
}

func Test200PopulatesAndCaches(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(muxOK(t, fixture(t, "react-packument.json"), fixture(t, "react-downloads.json"), "", &hits))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)
	frozen := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return frozen }

	dep, err := c.Fetch(context.Background(), "react", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 0 {
		t.Fatalf("Errors = %+v", dep.Errors)
	}
	if dep.Name != "react" {
		t.Fatalf("Name = %q", dep.Name)
	}
	if dep.License != "MIT" {
		t.Fatalf("License = %q", dep.License)
	}
	if dep.LastPublish != "2022-06-14T19:45:00Z" {
		t.Fatalf("LastPublish = %q", dep.LastPublish)
	}
	if dep.WeeklyDownloads == nil || *dep.WeeklyDownloads != 10000 {
		t.Fatalf("WeeklyDownloads = %v", dep.WeeklyDownloads)
	}
	_, hit, err := store.Get(sourceNPM, "react", cache.ClassOK, frozen)
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("expected packument cache write")
	}
	_, hit, err = store.Get(sourceNPMDL, "react", cache.ClassOK, frozen)
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("expected downloads cache write")
	}
}

func Test200ETag304Revalidate(t *testing.T) {
	t.Parallel()
	packument := fixture(t, "react-packument.json")
	downloads := fixture(t, "react-downloads.json")
	const etag = `"abc"`
	var packHits atomic.Int32
	var sawINM atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= len("/downloads/point/last-week/") &&
			r.URL.Path[:len("/downloads/point/last-week/")] == "/downloads/point/last-week/" {
			_, _ = w.Write(downloads)
			return
		}
		n := packHits.Add(1)
		if n == 1 {
			w.Header().Set("ETag", etag)
			_, _ = w.Write(packument)
			return
		}
		if r.Header.Get("If-None-Match") == etag {
			sawINM.Store(true)
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)
	t0 := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	now := t0
	c.now = func() time.Time { return now }

	first, err := c.Fetch(context.Background(), "react", "*")
	if err != nil {
		t.Fatal(err)
	}
	now = t0.Add(24*time.Hour + time.Second)
	second, err := c.Fetch(context.Background(), "react", "*")
	if err != nil {
		t.Fatal(err)
	}
	if packHits.Load() != 2 {
		t.Fatalf("packument hits = %d want 2", packHits.Load())
	}
	if !sawINM.Load() {
		t.Fatal("second call must send If-None-Match")
	}
	if first.License != second.License || first.LastPublish != second.LastPublish {
		t.Fatalf("304 body mismatch: %+v vs %+v", first, second)
	}
	_, hit, err := store.Get(sourceNPM, "react", cache.ClassOK, now)
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("304 must refresh TTL")
	}
}

func Test404NotFoundCached1h(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)
	frozen := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return frozen }

	dep, err := c.Fetch(context.Background(), "missing-pkg", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchNotFound || dep.Errors[0].Source != sourceNPM {
		t.Fatalf("Errors = %+v", dep.Errors)
	}
	_, hit, err := store.Get(sourceNPM, "missing-pkg", cache.ClassNotFound, frozen)
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("404 must be cached")
	}
	_, hit, err = store.Get(sourceNPM, "missing-pkg", cache.ClassNotFound, frozen.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("404 must still be fresh at 1h")
	}
	_, hit, err = store.Get(sourceNPM, "missing-pkg", cache.ClassNotFound, frozen.Add(time.Hour+time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("404 must miss after 1h")
	}
}

func Test429RateLimitNotCached(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)
	frozen := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return frozen }

	dep, err := c.Fetch(context.Background(), "ratey", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchRateLimit {
		t.Fatalf("Errors = %+v", dep.Errors)
	}
	p := filepath.Join(store.Root, cache.Key(sourceNPM, "ratey", cache.ClassOK))
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("429 created cache file: %v", err)
	}
	p = filepath.Join(store.Root, cache.Key(sourceNPM, "ratey", cache.ClassNotFound))
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("429 created 404 cache file: %v", err)
	}
}

func TestTimeoutNotCached(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)
	c.timeout = 30 * time.Millisecond
	frozen := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return frozen }

	dep, err := c.Fetch(context.Background(), "slow", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchTimeout {
		t.Fatalf("Errors = %+v", dep.Errors)
	}
	p := filepath.Join(store.Root, cache.Key(sourceNPM, "slow", cache.ClassOK))
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("timeout created cache file: %v", err)
	}
}

func TestMalformedJSONParse(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fixture(t, "malformed.json"))
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)

	dep, err := c.Fetch(context.Background(), "bad", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse {
		t.Fatalf("Errors = %+v", dep.Errors)
	}
	p := filepath.Join(store.Root, cache.Key(sourceNPM, "bad", cache.ClassOK))
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("parse error created cache file: %v", err)
	}
}

func TestMax8InFlight(t *testing.T) {
	t.Parallel()
	var inFlight, maxFlight atomic.Int32
	release := make(chan struct{})
	started := make(chan struct{}, 16)
	body := []byte(`{"name":"x","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"MIT"}},"time":{"1.0.0":"2020-01-01T00:00:00.000Z"}}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= len("/downloads/point/last-week/") &&
			r.URL.Path[:len("/downloads/point/last-week/")] == "/downloads/point/last-week/" {
			_, _ = w.Write([]byte(`{"downloads":1}`))
			return
		}
		n := inFlight.Add(1)
		for {
			old := maxFlight.Load()
			if n <= old || maxFlight.CompareAndSwap(old, n) {
				break
			}
		}
		started <- struct{}{}
		<-release
		inFlight.Add(-1)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)

	var wg sync.WaitGroup
	errCh := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		name := "pkg-" + string(rune('a'+i))
		go func() {
			defer wg.Done()
			_, err := c.Fetch(context.Background(), name, "*")
			if err != nil {
				errCh <- err
			}
		}()
	}
	for i := 0; i < 8; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for 8 in-flight")
		}
	}
	select {
	case <-started:
		t.Fatal("9th request entered handler while 8 held the semaphore")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
	if maxFlight.Load() > 8 {
		t.Fatalf("max in-flight = %d want <= 8", maxFlight.Load())
	}
}

func TestCacheHitZeroHTTP(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(muxOK(t, fixture(t, "react-packument.json"), fixture(t, "react-downloads.json"), "", &hits))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)
	frozen := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return frozen }

	if _, err := c.Fetch(context.Background(), "react", "*"); err != nil {
		t.Fatal(err)
	}
	n := hits.Load()
	if _, err := c.Fetch(context.Background(), "react", "*"); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != n {
		t.Fatalf("cache hit issued HTTP: before=%d after=%d", n, hits.Load())
	}
}

func TestDisabledAlwaysFetches(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(muxOK(t, fixture(t, "react-packument.json"), fixture(t, "react-downloads.json"), "", &hits))
	t.Cleanup(srv.Close)
	store := testStore(t)
	store.Disabled = true
	c := testClient(t, srv, store)

	if _, err := c.Fetch(context.Background(), "react", "*"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Fetch(context.Background(), "react", "*"); err != nil {
		t.Fatal(err)
	}
	if hits.Load() < 2 {
		t.Fatalf("disabled cache hits = %d want >= 2", hits.Load())
	}
	ents, err := os.ReadDir(store.Root)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(ents) != 0 {
		t.Fatalf("disabled cache wrote %d files", len(ents))
	}
}

func Test5xxNetworkNotCached(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)

	dep, err := c.Fetch(context.Background(), "boom", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchNetwork {
		t.Fatalf("Errors = %+v", dep.Errors)
	}
	p := filepath.Join(store.Root, cache.Key(sourceNPM, "boom", cache.ClassOK))
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("5xx created cache file: %v", err)
	}
}

func TestEncodeNameScoped(t *testing.T) {
	t.Parallel()
	got := encodeName("@types/node")
	if got != "@types%2Fnode" {
		t.Fatalf("encodeName = %q", got)
	}
}

func TestFetchRejectsUnsafePackageNamesBeforeHTTP(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	c := testClient(t, srv, testStore(t))
	for _, name := range []string{"../outside", "pkg?x=1", "pkg#fragment", "has space", "@scope/../pkg"} {
		dep, err := c.Fetch(context.Background(), name, "*")
		if err != nil {
			t.Fatal(err)
		}
		if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse {
			t.Errorf("Fetch(%q) errors = %+v", name, dep.Errors)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("invalid names made %d HTTP requests", calls.Load())
	}
}

func TestFetchRejectsCrossOriginRedirect(t *testing.T) {
	t.Parallel()
	var targetCalls atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		targetCalls.Add(1)
		_, _ = w.Write(fixture(t, "react-packument.json"))
	}))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/react", http.StatusFound)
	}))
	defer origin.Close()
	c := testClient(t, origin, testStore(t))
	dep, err := c.Fetch(context.Background(), "react", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchNetwork {
		t.Fatalf("redirect errors = %+v", dep.Errors)
	}
	if targetCalls.Load() != 0 {
		t.Fatalf("cross-origin target received %d requests", targetCalls.Load())
	}
}

func TestFetchCapsSameOriginRedirectsWithCustomPolicy(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		hop := r.URL.Query().Get("hop")
		if hop == "" {
			hop = "0"
		}
		if hop == "12" {
			_, _ = w.Write(fixture(t, "react-packument.json"))
			return
		}
		hopNumber, _ := strconv.Atoi(hop)
		http.Redirect(w, r, "/react?hop="+strconv.Itoa(hopNumber+1), http.StatusFound)
	}))
	defer server.Close()

	client := server.Client()
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return nil }
	c := New(testStore(t), client, WithBaseURL(server.URL), WithDownloadsURL(server.URL+"/downloads/point/last-week"))
	dep, err := c.Fetch(context.Background(), "react", "*")
	if err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got > 10 {
		t.Fatalf("same-origin redirect policy allowed %d requests, want at most 10", got)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchNetwork {
		t.Fatalf("redirect errors = %+v; requests = %d", dep.Errors, calls.Load())
	}
}

func TestFetchSelectsHighestVersionInDeclaredRange(t *testing.T) {
	t.Parallel()
	body := []byte(`{"name":"react","dist-tags":{"latest":"2.0.0"},"versions":{"1.4.0":{"license":"MIT"},"1.9.0":{"license":"ISC"},"2.0.0":{"license":"Apache-2.0"}},"time":{"modified":"2025-01-01T00:00:00Z","1.4.0":"2021-01-01T00:00:00Z","1.9.0":"2022-02-03T04:05:06Z","2.0.0":"2025-01-01T00:00:00Z"}}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/downloads/point/last-week/react" {
			_, _ = w.Write([]byte(`{"downloads":0}`))
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	c := testClient(t, srv, testStore(t))
	dep, err := c.Fetch(context.Background(), "react", "^1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if dep.Name != "react" || dep.Version != "^1.0.0" || dep.ResolvedVersion != "" {
		t.Fatalf("project identity changed: %+v", dep)
	}
	if dep.License != "ISC" || dep.LastPublish != "2022-02-03T04:05:06Z" {
		t.Fatalf("facts not selected from highest compatible release: %+v", dep)
	}
	if dep.WeeklyDownloads == nil || *dep.WeeklyDownloads != 0 || len(dep.Errors) != 0 {
		t.Fatalf("valid zero downloads was not preserved: %+v", dep)
	}
}

func TestFetchRejectsIncompatibleAndUnsupportedDeclarations(t *testing.T) {
	t.Parallel()
	body := []byte(`{"name":"react","dist-tags":{"latest":"2.0.0"},"versions":{"2.0.0":{"license":"MIT"}},"time":{"2.0.0":"2025-01-01T00:00:00Z"}}`)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	c := testClient(t, srv, testStore(t))
	dep, err := c.Fetch(context.Background(), "react", "^1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchNotFound {
		t.Fatalf("no matching release errors = %+v", dep.Errors)
	}
	dep, err = c.Fetch(context.Background(), "react", "^1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchNotFound {
		t.Fatalf("cached no-match errors = %+v", dep.Errors)
	}
	dep, err = c.Fetch(context.Background(), "react", "https://example.invalid/pkg.tgz")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse {
		t.Fatalf("unsupported declaration errors = %+v", dep.Errors)
	}
	if calls.Load() != 1 {
		t.Fatalf("unsupported declaration made network request; calls=%d", calls.Load())
	}
}

func TestFetchRejectsEmptyAndWhitespaceDeclarationsWithoutRegistryRequests(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`{"name":"pkg","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"MIT"}}}`))
	}))
	defer srv.Close()
	c := testClient(t, srv, testStore(t))
	for _, declared := range []string{"", " ", "\t\n"} {
		dep, err := c.Fetch(t.Context(), "pkg", declared)
		if err != nil {
			t.Fatal(err)
		}
		if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse {
			t.Errorf("declaration %q errors = %+v, want parse error", declared, dep.Errors)
		}
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("empty declaration made %d registry requests", got)
	}
}

func TestFetchDownloadFailureIsDegraded(t *testing.T) {
	t.Parallel()
	body := fixture(t, "react-packument.json")
	for _, downloads := range []string{`{}`, `{"downloads":null}`, `{"downloads":"5"}`, `{"downloads":1.5}`, `{"downloads":-1}`} {
		t.Run(downloads, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/downloads/point/last-week/react" {
					_, _ = w.Write([]byte(downloads))
					return
				}
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			c := testClient(t, srv, testStore(t))
			dep, err := c.Fetch(context.Background(), "react", "*")
			if err != nil {
				t.Fatal(err)
			}
			if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse {
				t.Fatalf("invalid download count errors = %+v", dep.Errors)
			}
			if dep.License != "" || dep.WeeklyDownloads != nil {
				t.Fatalf("invalid downloads must degrade the whole Card: %+v", dep)
			}
		})
	}
}

func TestPartialMetadataAndDownloadsFailuresStayUnknown(t *testing.T) {
	for _, tc := range []struct {
		name            string
		metadataStatus  int
		downloadsStatus int
		wantKind        types.FetchKind
	}{
		{name: "metadata failure", metadataStatus: http.StatusServiceUnavailable, wantKind: types.FetchNetwork},
		{name: "downloads failure after metadata", metadataStatus: http.StatusOK, downloadsStatus: http.StatusInternalServerError, wantKind: types.FetchNetwork},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metadata := fixture(t, "react-packument.json")
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/downloads/point/last-week/react" {
					if tc.downloadsStatus != http.StatusOK {
						http.Error(w, "fixture failure", tc.downloadsStatus)
						return
					}
					_, _ = w.Write([]byte(`{"downloads":1200}`))
					return
				}
				if tc.metadataStatus != http.StatusOK {
					http.Error(w, "fixture failure", tc.metadataStatus)
					return
				}
				_, _ = w.Write(metadata)
			}))
			defer srv.Close()
			dep, err := testClient(t, srv, testStore(t)).Fetch(t.Context(), "react", "^18.0.0")
			if err != nil {
				t.Fatal(err)
			}
			if len(dep.Errors) != 1 || dep.Errors[0].Kind != tc.wantKind {
				t.Fatalf("fetch errors = %+v, want %s", dep.Errors, tc.wantKind)
			}
			if dep.License != "" || dep.WeeklyDownloads != nil || dep.Rarity != "" || dep.Chaos != nil {
				t.Fatalf("incomplete npm response fabricated assessed facts: %+v", dep)
			}
		})
	}
}

func TestNPMRangeCompatibility(t *testing.T) {
	t.Parallel()
	body := []byte(`{"name":"pkg","dist-tags":{"latest":"9.0.0"},"versions":{"0.0.0":{"license":"zero"},"0.0.1":{"license":"zero-patch"},"0.2.5":{"license":"caret-zero"},"1.0.0-beta.2":{"license":"beta"},"1.0.0":{"license":"stable"},"1.2.3":{"license":"patch"},"1.4.0":{"license":"hyphen"},"1.4.1":{"license":"tilde-patch"},"1.9.0":{"license":"highest"}},"time":{}}`)
	cases := []struct{ rangeText, want string }{
		{"1.2.3", "1.2.3"},
		{"1.2", "1.2.3"},
		{"1.x", "1.9.0"},
		{"^0.2.0", "0.2.5"},
		{"~1.2.0", "1.2.3"},
		{">=1.2.0 <2.0.0", "1.9.0"},
		{"^0.2.0 || ^1.0.0", "1.9.0"},
		{"1.0.0 - 1.4.0", "1.4.0"},
		{"~0.0.0", "0.0.1"},
		{"^1.0.0-beta.1", "1.9.0"},
	}
	for _, tc := range cases {
		t.Run(tc.rangeText, func(t *testing.T) {
			constraintText, err := npmRange(tc.rangeText)
			if err != nil {
				t.Fatal(err)
			}
			dep, err := parsePackument("pkg", tc.rangeText, constraintText, body)
			if err != nil {
				t.Fatal(err)
			}
			if dep.License == "" {
				t.Fatal("selected version has no fixture license")
			}
			var selected string
			for version, meta := range map[string]string{
				"0.0.1": "zero-patch", "0.2.5": "caret-zero", "1.0.0": "stable",
				"1.2.3": "patch", "1.4.0": "hyphen", "1.4.1": "tilde-patch", "1.9.0": "highest",
			} {
				if dep.License == meta {
					selected = version
					break
				}
			}
			if selected != tc.want {
				t.Fatalf("selected version %q, want %q (license %q)", selected, tc.want, dep.License)
			}
		})
	}
}

func TestFetchRejectsMismatchedMetadataIdentity(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"name":"other","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"MIT"}}}`))
	}))
	t.Cleanup(srv.Close)
	c := testClient(t, srv, testStore(t))
	dep, err := c.Fetch(context.Background(), "react", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse {
		t.Fatalf("mismatched metadata errors = %+v", dep.Errors)
	}
}

func TestFetchKeepsHTTPStatusWhenErrorBodyIsOversized(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		status int
		kind   types.FetchKind
	}{
		{http.StatusNotFound, types.FetchNotFound},
		{http.StatusTooManyRequests, types.FetchRateLimit},
		{http.StatusServiceUnavailable, types.FetchNetwork},
	} {
		t.Run(strconv.Itoa(tc.status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write(make([]byte, maxMetadataBody+1))
			}))
			defer srv.Close()
			c := testClient(t, srv, testStore(t))
			dep, err := c.Fetch(context.Background(), "package", "*")
			if err != nil {
				t.Fatal(err)
			}
			if len(dep.Errors) != 1 || dep.Errors[0].Kind != tc.kind {
				t.Fatalf("status %d: errors = %+v", tc.status, dep.Errors)
			}
		})
	}
}

func TestVersionSelectionPrereleaseRule(t *testing.T) {
	t.Parallel()
	body := []byte(`{"name":"pkg","dist-tags":{"latest":"2.0.0-beta.2"},"versions":{"1.9.0":{"license":"stable"},"2.0.0-beta.1":{"license":"beta-one"},"2.0.0-beta.2":{"license":"beta-two"}}}`)
	for _, tc := range []struct{ declared, want string }{
		{"*", "stable"},
		{"^2.0.0-beta.1", "beta-two"},
		{"2.0.0-beta.2", "beta-two"},
		{"^2.0.0-beta.3", ""},
	} {
		constraint, err := npmRange(tc.declared)
		if err != nil {
			t.Fatal(err)
		}
		dep, err := parsePackument("pkg", tc.declared, constraint, body)
		if tc.want == "" {
			if !errors.Is(err, errNoCompatibleVersion) {
				t.Fatalf("%q: error = %v", tc.declared, err)
			}
			continue
		}
		if err != nil || dep.License != tc.want {
			t.Fatalf("%q: dependency = %+v, error = %v", tc.declared, dep, err)
		}
	}
}

func TestMissingLatestFallsBackToPublishedStableVersion(t *testing.T) {
	for _, versions := range []string{
		`{"1.1.0":{"license":"ISC"},"1.4.0":{"license":"MIT"},"1.8.0-beta.1":{}}`,
		`{"1.1.0":{"license":"ISC"},"1.4.0":{"license":"MIT"},"1.8.0-beta.1":{},"1.9.0":null}`,
	} {
		body := []byte(`{"name":"pkg","dist-tags":{"latest":"1.9.0"},"versions":` + versions +
			`,"time":{"1.4.0":"2025-01-01T00:00:00Z"}}`)
		dep, err := parsePackument("pkg", "^1", "^1", body)
		if err != nil || dep.License != "MIT" || dep.LastPublish != "2025-01-01T00:00:00Z" || dep.Version != "^1" {
			t.Fatalf("published fallback: %+v, %v", dep, err)
		}
	}
}

func TestFetchDoesNotReflectMalformedRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "%zz-sensitive-fixture")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()
	dep, err := testClient(t, srv, testStore(t)).Fetch(t.Context(), "pkg", "*")
	if err != nil || len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchNetwork ||
		dep.Errors[0].Detail != "npm request failed" {
		t.Fatalf("unexpected redirect result: %+v, %v", dep, err)
	}
}

func TestLicenseDropsControlCharacters(t *testing.T) {
	license := licenseString([]byte(`"M\u001bI\u202eT"`))
	if license != "MIT" {
		t.Fatalf("license = %q, want MIT", license)
	}
}

func TestFetchStreamsMetadataLargerThan32MiB(t *testing.T) {
	t.Parallel()
	const padding = maxMetadataBody + 1
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/downloads/point/last-week/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"downloads":12345}`))
			return
		}
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", `"vite-large"`)
		_, _ = io.WriteString(w, `{"name":"vite","readme":"`)
		if err := writeRepeat(w, 'a', padding); err != nil {
			t.Errorf("write padding: %v", err)
			return
		}
		_, _ = io.WriteString(w, `","dist-tags":{"latest":"1.2.3"},"versions":{"1.2.3":{"license":"MIT","description":"bundler"}},"time":{"1.2.3":"2024-06-01T12:00:00.000Z"}}`)
	}))
	t.Cleanup(srv.Close)
	store := testStore(t)
	c := testClient(t, srv, store)
	frozen := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return frozen }

	dep, err := c.Fetch(context.Background(), "vite", "^1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 0 {
		t.Fatalf("Errors = %+v", dep.Errors)
	}
	if dep.Name != "vite" || dep.Version != "^1.0.0" || dep.License != "MIT" {
		t.Fatalf("card = %+v", dep)
	}
	if dep.LastPublish != "2024-06-01T12:00:00Z" {
		t.Fatalf("LastPublish = %q", dep.LastPublish)
	}
	if dep.WeeklyDownloads == nil || *dep.WeeklyDownloads != 12345 {
		t.Fatalf("WeeklyDownloads = %v", dep.WeeklyDownloads)
	}

	entry, hit, err := store.Get(sourceNPM, "vite", cache.ClassOK, frozen)
	if err != nil || !hit {
		t.Fatalf("cache hit = %v, err = %v", hit, err)
	}
	if int64(len(entry.Body)) > 4<<10 {
		t.Fatalf("cached slim body is %d bytes, want a small retained document", len(entry.Body))
	}
	if !bytes.Contains(entry.Body, []byte(`"license":"MIT"`)) || bytes.Contains(entry.Body, []byte(`"readme"`)) {
		t.Fatalf("cached body should be slimmed: %s", entry.Body)
	}
	if hits.Load() != 1 {
		t.Fatalf("metadata hits = %d, want 1", hits.Load())
	}

	// Fresh cache must not re-fetch, and must still produce a real card.
	dep, err = c.Fetch(context.Background(), "vite", "^1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 0 || dep.License != "MIT" || dep.LastPublish != "2024-06-01T12:00:00Z" {
		t.Fatalf("cached refetch = %+v", dep)
	}
	if hits.Load() != 1 {
		t.Fatalf("cache hit re-fetched metadata: hits = %d", hits.Load())
	}
}

func TestFetchRejectsRetainedMetadataOverLimit(t *testing.T) {
	t.Parallel()
	const retained = maxMetadataBody + 1
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/downloads/point/last-week/") {
			_, _ = w.Write([]byte(`{"downloads":1}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"name":"pkg","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"`)
		if err := writeRepeat(w, 'B', retained); err != nil {
			t.Errorf("write license: %v", err)
			return
		}
		_, _ = io.WriteString(w, `"}},"time":{"1.0.0":"2024-01-01T00:00:00Z"}}`)
	}))
	t.Cleanup(srv.Close)
	dep, err := testClient(t, srv, testStore(t)).Fetch(context.Background(), "pkg", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse ||
		dep.Errors[0].Detail != "npm response exceeds size limit" {
		t.Fatalf("errors = %+v, want a size-limit parse error", dep.Errors)
	}
	if dep.License != "" || dep.WeeklyDownloads != nil {
		t.Fatalf("oversized retained metadata must not populate facts: %+v", dep)
	}
}

func TestSlimMetadataEnforcesReadAndRetainLimits(t *testing.T) {
	t.Parallel()
	padding := strings.Repeat("a", 64)
	raw := `{"name":"pkg","readme":"` + padding + `","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"MIT"}},"time":{"1.0.0":"2024-01-01T00:00:00Z"}}`
	if _, err := slimMetadata(strings.NewReader(raw), 32, 1024); !errors.Is(err, errResponseTooLarge) {
		t.Fatalf("read limit: err = %v, want errResponseTooLarge", err)
	}

	slim, err := slimMetadata(strings.NewReader(raw), int64(len(raw)), 1024)
	if err != nil {
		t.Fatal(err)
	}
	var p packument
	if err := json.Unmarshal(slim, &p); err != nil {
		t.Fatal(err)
	}
	if p.Name != "pkg" || p.DistTags.Latest != "1.0.0" || p.Time["1.0.0"] != "2024-01-01T00:00:00Z" {
		t.Fatalf("slim packument = %+v", p)
	}
	if p.Versions["1.0.0"] == nil || licenseString(p.Versions["1.0.0"].License) != "MIT" {
		t.Fatalf("slim license = %+v", p.Versions)
	}
	if bytes.Contains(slim, []byte("readme")) {
		t.Fatalf("slim body retained unused fields: %s", slim)
	}

	license := strings.Repeat("C", 64)
	heavy := `{"name":"pkg","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"` + license + `"}}}`
	if _, err := slimMetadata(strings.NewReader(heavy), int64(len(heavy)), 32); !errors.Is(err, errResponseTooLarge) {
		t.Fatalf("retain limit: err = %v, want errResponseTooLarge", err)
	}
}

func TestFetchInvalidMetadataJSONIsParseError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/downloads/point/last-week/") {
			_, _ = w.Write([]byte(`{"downloads":1}`))
			return
		}
		_, _ = w.Write([]byte(`{"name":`))
	}))
	t.Cleanup(srv.Close)
	dep, err := testClient(t, srv, testStore(t)).Fetch(context.Background(), "pkg", "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(dep.Errors) != 1 || dep.Errors[0].Kind != types.FetchParse {
		t.Fatalf("errors = %+v, want parse", dep.Errors)
	}
	if dep.Errors[0].Detail == "npm request failed" || dep.Errors[0].Detail == "npm response exceeds size limit" {
		t.Fatalf("detail = %q, want a JSON parse message", dep.Errors[0].Detail)
	}
}

func writeRepeat(w io.Writer, ch byte, n int64) error {
	buf := bytes.Repeat([]byte{ch}, 1<<20)
	var written int64
	for written < n {
		chunk := buf
		if remain := n - written; int64(len(chunk)) > remain {
			chunk = chunk[:remain]
		}
		if _, err := w.Write(chunk); err != nil {
			return err
		}
		written += int64(len(chunk))
	}
	return nil
}

func FuzzParsePackument(f *testing.F) {

	f.Add("pkg", "^1.0.0", []byte(`{"name":"pkg","dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"MIT"}}}`))
	f.Add("@scope/pkg", "latest", []byte(`null`))
	f.Fuzz(func(t *testing.T, name, declared string, body []byte) {
		if int64(len(body)) > maxMetadataBody {
			t.Skip()
		}
		rangeText, err := npmRange(declared)
		if err != nil {
			return
		}
		dep, err := parsePackument(name, declared, rangeText, body)
		if err == nil && (dep.Name != name || dep.Version != declared) {
			t.Fatalf("successful parse changed requested identity: %+v", dep)
		}
	})
}
