Status: ready-for-agent

# pkg/registry — npm client, bound 8, cache-aware

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Fetch npm packument + weekly downloads for a name. Semaphore 8 around **HTTP only**. Cache Get before taking a slot; Put after 200/304/404. Base URL from config, default `https://registry.npmjs.org`. Downloads API may be `https://api.npmjs.org/downloads/point/last-week/{name}` — document both URLs in one place (constructor options), not scattered.

Tests: httptest fixture server; never live npm.

## Files

- Create `pkg/registry/registry.go` (and `_test.go`)

## Interface

```go
package registry

type Client struct { /* HTTP client, Store, BaseURL, DownloadsURL, Sem *semaphore */ }

func New(store *cache.Store, opts ...Option) *Client

type Packument struct {
    Name     string
    Latest   string
    Versions map[string]VersionMeta // license, time
    Time     map[string]string      // modified / version timestamps
}

func (c *Client) Fetch(ctx context.Context, name string) (Packument, int weeklyDownloads, err types.FetchError /* zero Kind means success */)
```

Prefer returning `(Packument, downloads, []types.FetchError)` with empty errors on success. Do not abort sibling fetches (caller runs errgroup). On 404 return FetchError Source `npm`, Kind `not_found`. Timeouts → `timeout`. 429 → `rate_limit`.

If-None-Match when cache has ETag but TTL expired. 304: refresh FetchedAt, serve body.

v1 dist-tags.latest vs range matching can live here or in stats; keep Fetch returning raw packument + downloads.

## Acceptance criteria

- [ ] At most 8 in-flight HTTP requests (test with a slow fixture)
- [ ] Cache hit does not start HTTP (assert server request count 0)
- [ ] 404 returns FetchError source `npm` kind `not_found`, not a Go `error` for “package missing”
- [ ] 500/timeout: Kind `network` or `timeout`; nothing Put in cache
- [ ] Tests use httptest only

## Blocked by

- `.scratch/depdeck-v1/issues/01-pkg-types.md`
- `.scratch/depdeck-v1/issues/02-pkg-cache.md`
- `.scratch/depdeck-v1/issues/13-testdata.md` (fixture contract must be ready first)

## ADRs

0003, 0004

## V1 contract amendment

This amendment overrides the earlier `Fetch` signature and version-selection text.

Follow [ADR-0017](../../../docs/adr/0017-go-runtime-and-npm-version-ranges.md) and [docs/THREAT_MODEL.md](../../../docs/THREAT_MODEL.md). `Fetch` accepts the declared dependency version range and returns the enriched dependency facts plus typed fetch errors. Select the best npm version that satisfies the declaration, using the pinned semver library. Add compatibility tests for common npm range forms. Keep package fetch failures as dependency data; do not fail the whole scan. Bound response bodies, use timeouts, cap HTTP concurrency at eight, and reject unsafe redirects. Cache reads happen before taking an HTTP slot. Ticket 13 fixtures must be ready before this ticket is implemented.
