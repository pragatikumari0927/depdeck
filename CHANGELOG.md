# Changelog

This project follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Rarity, Chaos, policy per ADR-0009/0010 (b156a23)
- Dependency, Rarity, Tag, FetchError structs (420b921)
- Add SpecialMove per ADR-0013 (ec1c8b5)
- Rule-based generator per ADR-0005 (0dd7244)
- On-disk HTTP cache per ADR-0004 (e0665ab)
- Root package.json reader per ADR-0011 (9d90d48)
- Self-contained deck.html per ADR-0006 (1e584ee)
- Render(deps, io.Writer) per ADR-0006 (b55c8af)
- Assemble depdeck v1 release candidate (f8b485e)

### Changed

- Initial commit before agent setup (fa0a2b8)
- CI vet step, DECISIONS index row, platform-aware -race hook (6463007)
- Add SPEC, CONTEXT, AGENTS, ADRs 0001-0011, agent scaffold (eb2e7aa)
- Normalize DECISIONS.md Locked table to 3-column format (1362078)
- Add go.mod (module depdeck, Go 1.22) (6346d2b)
- Align Core types with Go representation (ebfb8a0)
- Add card tokens, typography, layout (ca9c9bc)
- Relax pkg/types freeze to allow additive fields (bc2b33c)
- Note ParseFile takes a file path in 07-pipeline (0665240)
- Add .cursorignore for artifacts, caches, secrets (f78f3af)
- Move deck.html under pkg/render for embed (ce3de98)
- Tighten ADR immutability instruction (81b62ef)
- Update package map for pkg/render/templates move (15213c9)
- Add runtime flow section (7371437)
- Add PHASES.md roadmap and git workflow (c0eb380)
- Remove dead extends preamble from Go rules (91301d8)
- Add RULES.md rule index; track archived ponytail rule (f6c66a1)
- Add outbound-HTTP context timeout pattern (6fe1350)
- Fix section levels; add concurrency pattern (f1e8a80)
- Add .gitignore; pin Engram writes to depdeck (ed6075a)
- Align hooks, testing, DEVELOPMENT on skip and race (b62de14)
- Ignore local Cursor state and npm residue (660d5c4)
- Log deferred decision on .agents/skills tracking (0fab3c3)
- Record ripgrep install paths for Windows (3dfeeee)
- Route shell search through ripgrep (dc78e3c)
- Index shell-search rule (6d70f19)
- Add verify-depdeck CLI verification skill (e8ec477)
- Add .md rule wrappers for Cursor .mdc rules (d1a477b)
- Migrate rules to full-content copies; enforce rg (215c3c6)
- Record Grok Build as primary agent (ADR-0015) (f22352c)
- Add README for depdeck v1 (04683b4)
- Redesign README with diagrams, badges and honest v1 status (97850a8)
- Enable Dependabot for Go and Actions (cb75397)
- Untrack local agent config folders (633b255)
- Add banner and designed diagram images to README (3a89dc6)
- Mark unbuilt features as planned (c4964d8)
- Update Go prerequisite to 1.26.9 (e65d04c)
- Bump actions/setup-go from 5 to 7 (710336d)
- Close v1 security and acceptance ticket (4262e03)

### Fixed

- Wrap errors with %w; clean tmp on rename fail (1f5aaee)
- Require Go 1.26.9 to clear 8 stdlib vulnerabilities (91ae97c)
