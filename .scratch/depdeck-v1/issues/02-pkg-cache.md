Status: ready-for-agent

# pkg/cache — on-disk HTTP cache

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Per-package files under `os.UserCacheDir()/depdeck/` (overridable root for tests). ETag, TTL classes, no cache of 429/5xx/timeout. `--no-cache` equivalent: a `Nop` or `Enabled: false` store that never reads or writes.

## Files

- Create `pkg/cache/cache.go` (and `_test.go`)

## Interface

```go
package cache

type Class string // "ok" | "not_found"

type Entry struct {
    ETag      string
    Status    int
    FetchedAt time.Time
    Body      []byte
}

type Store struct { Root string; Disabled bool }

func (s *Store) Get(source, name string, class Class) (Entry, bool, error)
func (s *Store) Put(source, name string, class Class, e Entry) error
func Key(source, name string, class Class) string // sha256 hex filename
func Fresh(e Entry, class Class, now time.Time) bool // 24h for ok, 1h for not_found
```

Put is not called for 429, 5xx, or transport errors (callers must not Put those). 304 handling is the registry client: Get expired-with-etag, then Put refresh FetchedAt on 304.

## Acceptance criteria

- [ ] Key is hex sha256 of `source + ":" + name + ":" + class`; same inputs → same filename
- [ ] Fresh: 200-class older than 24h is not fresh; 404-class older than 1h is not fresh
- [ ] Disabled store Get always miss; Put is no-op
- [ ] Two packages are two files (not one monolith)
- [ ] Tests use a temp Root, never the real user cache

## Blocked by

- `.scratch/depdeck-v1/issues/01-pkg-types.md` (sequencing; cache does not import types)

## ADRs

0004, 0003 (cache hits do not take the fetch semaphore — enforced in registry ticket)
