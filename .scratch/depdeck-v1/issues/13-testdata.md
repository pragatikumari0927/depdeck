Status: ready-for-agent

# testdata — fixture npm responses

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Canonical fixture bodies for httptest: 200 packument + downloads, 404, 429, 500, packument with license missing, old last-publish, high weekly downloads. Workspace fixtures if not already added in parser ticket. README in testdata explaining no live registry.

## Files

- Create `testdata/npm/README.md`
- Create `testdata/npm/react-200.json` (and downloads json)
- Create `testdata/npm/missing-404.json` (or status-only)
- Create `testdata/npm/unlicensed-200.json`
- Create `testdata/npm/stale-publish-200.json`
- Ensure `testdata/workspace/` from parser ticket remains the workspace negative case

## Interface

Registry tests load these files. Do not invent a second fixture layout in pkg/registry.

## Acceptance criteria

- [ ] At least one 200 packument with weekly downloads ≥ 10000 (Legendary)
- [ ] At least one 200 with no license
- [ ] At least one 200 with last publish older than 5 years
- [ ] Documented 404 and 500 fixtures
- [ ] README: tests must not call registry.npmjs.org

## Blocked by

- `.scratch/depdeck-v1/issues/01-pkg-types.md`

## ADRs

0003, 0004, 0009, 0010

## V1 contract amendment

This amendment replaces the earlier blocker and fixture sequencing.

Remove the dependency on ticket 04. Complete and review these fixture contracts before implementation of ticket 04. Include representative npm ranges and registry cases for successful metadata, no matching version, malformed JSON, 404, 429, 500, timeout, redirect, oversized body, missing license, old publish date, and high download count. Tests must use local files or an HTTP test server. Never call the live registry.
