# Root-only workspaces; Tag is one enum

v1 scans the **root** `package.json` only. An npm/pnpm `workspaces` field is ignored; nested package.json files are not walked. No `--workspaces` flag. That matches directs-only: the Deck is one Roster, not a monorepo dump.

**Tag** is a **single** string enum: `dependencies` | `devDependencies` | `optionalDependencies`. npm puts a name on one of those lists (if it appears on two, last-write-wins in npm’s own merge is not our problem — parser takes the first list found in that order: `dependencies`, then `devDependencies`, then `optionalDependencies`, and does not emit duplicates). Filter APIs take one Tag, not an array.

**Status:** accepted
