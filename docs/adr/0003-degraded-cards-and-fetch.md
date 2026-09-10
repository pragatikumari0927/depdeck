# Degraded Cards: keep the Card, structured fetch errors

A failed registry call is not a missing Roster member. The pipeline keeps the Dependency, attaches `errors[]` (`source`, `kind`: `not_found` | `timeout` | `rate_limit` | `network` | `parse`, `detail`), and omits unknown enrichment (Rarity is null, not Common). Rule Flavor still runs from `name`; AI Flavor does not.

**Status:** accepted

## Concurrency

Bound of 8 **per Source**, around the **network call only**. Cache hits do not take a slot. v1 has one Source (npm); a global semaphore would later stall npm behind GitHub's rate limit.

## Adapter translation

- CLI `check`: fetch errors are warnings. Exit 3 only if every fetch failed (pipeline broken), not on partial npm hiccups.
- MCP `get_card`: 404 is a successful tool result with `errors`. MCP errors are for malformed arguments or crashes.

## Registry

Base URL is a reserved config point; v1 hardcodes `registry.npmjs.org`. Do not scatter that URL. Auth for private registries is deferred.

## `FetchError.Source` in v1

Every FetchError `source` is `"npm"`. The field exists for v2 GitHub enrichment. Do not invent a second Source, a dummy `"unknown"`, or omit `source` on v1 errors.
