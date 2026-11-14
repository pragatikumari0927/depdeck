package cache

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestGetMissingMiss(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	_, hit, err := s.Get("npm", "left-pad", ClassOK, now)
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("expected miss")
	}
}

func TestPutThenGetHit(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	fetched := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	want := Entry{ETag: `"abc"`, Status: 200, FetchedAt: fetched, Body: []byte(`{"name":"react"}`)}
	if err := s.Put("npm", "react", ClassOK, want); err != nil {
		t.Fatal(err)
	}
	got, hit, err := s.Get("npm", "react", ClassOK, fetched)
	if err != nil {
		t.Fatal(err)
	}
	if !hit {
		t.Fatal("expected hit")
	}
	if got.ETag != want.ETag || got.Status != want.Status || string(got.Body) != string(want.Body) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestLoadStaleReturnsETag(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	fetched := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	want := Entry{ETag: `"abc"`, Status: 200, FetchedAt: fetched, Body: []byte(`{"name":"react"}`)}
	if err := s.Put("npm", "react", ClassOK, want); err != nil {
		t.Fatal(err)
	}
	staleAt := fetched.Add(24*time.Hour + time.Second)
	_, hit, err := s.Get("npm", "react", ClassOK, staleAt)
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("Get must miss after 24h")
	}
	got, ok, err := s.Load("npm", "react", ClassOK)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Load must return stale entry")
	}
	if got.ETag != want.ETag || string(got.Body) != string(want.Body) {
		t.Fatalf("Load = %+v want %+v", got, want)
	}
}

func TestGetOKAfter24hMiss(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	fetched := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	e := Entry{ETag: `"e"`, Status: 200, FetchedAt: fetched, Body: []byte("ok")}
	if err := s.Put("npm", "old", ClassOK, e); err != nil {
		t.Fatal(err)
	}
	_, hit, err := s.Get("npm", "old", ClassOK, fetched.Add(24*time.Hour+time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("expected miss after 24h")
	}
}

func TestGet404After1hMiss(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	fetched := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	e := Entry{ETag: "", Status: 404, FetchedAt: fetched, Body: []byte("not found")}
	if err := s.Put("npm", "missing-pkg", ClassNotFound, e); err != nil {
		t.Fatal(err)
	}
	_, hit, err := s.Get("npm", "missing-pkg", ClassNotFound, fetched.Add(time.Hour+time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("expected miss after 1h")
	}
}

func TestPut429NotWritten(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	err := s.Put("npm", "ratey", ClassOK, Entry{Status: 429, FetchedAt: now, Body: []byte("slow down")})
	if err != nil {
		t.Fatal(err)
	}
	_, hit, err := s.Get("npm", "ratey", ClassOK, now)
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("429 must not be cached")
	}
	name := Key("npm", "ratey", ClassOK)
	if _, err := os.Stat(filepath.Join(s.Root, name)); !os.IsNotExist(err) {
		t.Fatalf("429 created file: %v", err)
	}
}

func TestConcurrentEight(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	var wg sync.WaitGroup
	errCh := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		name := "pkg-" + string(rune('a'+i))
		go func() {
			defer wg.Done()
			e := Entry{ETag: `"x"`, Status: 200, FetchedAt: now, Body: []byte(name)}
			if err := s.Put("npm", name, ClassOK, e); err != nil {
				errCh <- err
				return
			}
			got, hit, err := s.Get("npm", name, ClassOK, now)
			if err != nil {
				errCh <- err
				return
			}
			if !hit || string(got.Body) != name {
				errCh <- errMismatch
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

var errMismatch = errors.New("get mismatch")

func TestDisabled(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	s.Disabled = true
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	if err := s.Put("npm", "x", ClassOK, Entry{Status: 200, FetchedAt: now, Body: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	_, hit, err := s.Get("npm", "x", ClassOK, now)
	if err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Fatal("disabled Get must miss")
	}
}

func TestKeyStable(t *testing.T) {
	t.Parallel()
	a := Key("npm", "react", ClassOK)
	b := Key("npm", "react", ClassOK)
	if a == "" || a != b {
		t.Fatalf("Key unstable: %q %q", a, b)
	}
	if Key("npm", "react", ClassNotFound) == a {
		t.Fatal("class must change Key")
	}
}

func TestFreshWindows(t *testing.T) {
	t.Parallel()
	fetched := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	e := Entry{FetchedAt: fetched}
	if !Fresh(e, ClassOK, fetched.Add(24*time.Hour)) {
		t.Fatal("ok at exactly 24h should be fresh")
	}
	if Fresh(e, ClassOK, fetched.Add(24*time.Hour+time.Second)) {
		t.Fatal("ok after 24h should be stale")
	}
	if !Fresh(e, ClassNotFound, fetched.Add(time.Hour)) {
		t.Fatal("404 at exactly 1h should be fresh")
	}
	if Fresh(e, ClassNotFound, fetched.Add(time.Hour+time.Second)) {
		t.Fatal("404 after 1h should be stale")
	}
}

func TestNewEmptyUsesUserCacheDir(t *testing.T) {
	t.Parallel()
	s, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	base, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(base, "depdeck")
	if s.Root != want {
		t.Fatalf("Root = %q want %q", s.Root, want)
	}
}

func TestGetCorruptJSON(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	if err := os.MkdirAll(s.Root, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(s.Root, Key("npm", "bad", ClassOK))
	if err := os.WriteFile(p, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	if _, hit, err := s.Get("npm", "bad", ClassOK, now); err != nil || hit {
		t.Fatalf("corrupt cache must be a miss, hit=%v err=%v", hit, err)
	}
}

func TestGetOversizedCacheEntryIsMiss(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		source string
		class  Class
		limit  int64
	}{
		{name: "metadata", source: "npm", class: ClassOK, limit: 48 << 20},
		{name: "downloads", source: "npm-downloads", class: ClassOK, limit: 384 << 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := testStore(t)
			if err := os.MkdirAll(s.Root, 0o755); err != nil {
				t.Fatal(err)
			}
			p := filepath.Join(s.Root, Key(tc.source, "large", tc.class))
			f, err := os.Create(p)
			if err != nil {
				t.Fatal(err)
			}
			if err := f.Truncate(tc.limit + 1); err != nil {
				_ = f.Close()
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			_, hit, err := s.Get(tc.source, "large", tc.class, time.Now())
			if err != nil || hit {
				t.Fatalf("oversized cache must be a miss, hit=%v err=%v", hit, err)
			}
		})
	}
}

func TestPutReplacesExistingEntry(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	first := Entry{Status: 200, FetchedAt: now, Body: []byte("first")}
	second := Entry{Status: 200, FetchedAt: now.Add(time.Minute), Body: []byte("second")}
	if err := s.Put("npm", "replace", ClassOK, first); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("npm", "replace", ClassOK, second); err != nil {
		t.Fatal(err)
	}
	got, hit, err := s.Get("npm", "replace", ClassOK, second.FetchedAt)
	if err != nil || !hit || string(got.Body) != "second" {
		t.Fatalf("replacement = %+v hit=%v err=%v", got, hit, err)
	}
}

func TestPut5xxNotWritten(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	if err := s.Put("npm", "boom", ClassOK, Entry{Status: 503, FetchedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Root, Key("npm", "boom", ClassOK))); !os.IsNotExist(err) {
		t.Fatalf("5xx created file: %v", err)
	}
}

func TestOversizedDecodedDownloadsCacheIsMiss(t *testing.T) {
	store := testStore(t)
	now := time.Now()
	entry := Entry{Status: 200, FetchedAt: now, Body: make([]byte, maxDownloadsBodySize+1)}
	if err := store.Put("npm-downloads", "pkg", ClassOK, entry); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(store.path("npm-downloads", "pkg", ClassOK))
	if err != nil || info.Size() > maxDownloadsEntrySize {
		t.Fatalf("expected fixture within encoded limit: %v, %v", info, err)
	}
	if _, hit, err := store.Load("npm-downloads", "pkg", ClassOK); err != nil || hit {
		t.Fatalf("oversized decoded cache entry: hit=%v, error=%v", hit, err)
	}
}

func TestFutureCacheEntryIsNotFresh(t *testing.T) {
	now := time.Now()
	entry := Entry{FetchedAt: now.Add(time.Hour)}
	for _, class := range []Class{ClassOK, ClassNotFound} {
		if Fresh(entry, class, now) {
			t.Errorf("future entry is fresh for %s", class)
		}
	}
}
