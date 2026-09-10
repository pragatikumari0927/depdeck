# On-disk HTTP cache, per package, user cache dir

Registry responses are cached on disk under `os.UserCacheDir()/depdeck/`, shared across projects and Faces. One file per package (`sha256(source + ":" + name + ":" + ttl_class)`), not a monolithic store (that would serialize the bounded concurrent fetches). The cache stores npm *responses* keyed by name; upgrading `react@18` to `@19` still uses the same metadata document.

**Status:** accepted

## TTL

- 200 with body: 24h
- 304: reset the 24h window, keep body
- 404: 1h
- Never cache: 429, 5xx, timeouts, network errors

Store ETag from every 200. After TTL expiry, send `If-None-Match`; 304 refreshes `fetched_at`; 200 replaces body.

## `--no-cache`

Bypasses reads and writes. Tests use `--no-cache` plus a fixture server (never live npm). `depdeck check` in CI **uses** the cache; it must not depend on npm uptime. Do not use a “TTL 0” mode.

## Non-goals

Do not key by project, invalidate on `package.json` change, or add `--refresh` for manifest edits.
