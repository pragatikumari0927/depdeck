Status: ready-for-agent

# pkg/parser — root Roster only

## Parent

`.scratch/depdeck-v1/PRD.md`

## What to build

Read one `package.json`. Directs only, Tag precedence, ignore `workspaces` and nested manifests. Do not parse lockfiles (deferred). `ResolvedVersion` left empty.

## Files

- Create `pkg/parser/parser.go` (and `_test.go`)
- Create `testdata/workspace/package.json` and `testdata/workspace/packages/pkg-a/package.json` for the negative test

## Interface

```go
package parser

func ParseFile(path string) ([]types.Dependency, error) // path to package.json or directory containing it
```

Each Dependency has Name, Version (declared range/exact), Tag, empty enrichment.

Tag precedence if the same name appears in more than one object: `dependencies`, then `devDependencies`, then `optionalDependencies`. Skip `peerDependencies` and `bundledDependencies`. Ignore `workspaces`.

## Acceptance criteria

- [ ] Parses dependencies, devDependencies, optionalDependencies into Tags
- [ ] Does not emit peerDependencies or bundledDependencies
- [ ] Duplicate name kept once; first Tag in precedence order
- [ ] **Negative:** `testdata/workspace/` root lists `root-only`; nested `packages/pkg-a` lists `nested-only`. ParseFile(root) returns only root-only names — never `nested-only`
- [ ] No lockfile read; ResolvedVersion always `""`

## Blocked by

- `.scratch/depdeck-v1/issues/01-pkg-types.md`

## ADRs

0002, 0011

## V1 contract amendment

This amendment overrides the earlier directory-accepting `ParseFile` signature.

`ParseFile` accepts a path to a `package.json` file only. The CLI and pipeline resolve a directory to its `package.json`. Add a file-size limit and tests for a missing file, invalid JSON, and oversized input. Follow [docs/SPEC.md](../../../docs/SPEC.md) and [docs/THREAT_MODEL.md](../../../docs/THREAT_MODEL.md).
