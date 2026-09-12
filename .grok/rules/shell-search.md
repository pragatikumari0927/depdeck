# Shell search — ripgrep only

When running terminal commands, use `rg` (ripgrep) for content search
and `Get-ChildItem` only when listing is required. Never use
`Select-String` — it is the PowerShell equivalent of grep and is
slower and less precise than `rg`.

## Blocked

- `Select-String` — use `rg` instead
- `findstr` — use `rg` instead
- `Get-ChildItem -Recurse | Select-String` — use `rg PATTERN .`
- `Get-ChildItem` for content search — use `rg --files`

## Required

- Always pass an explicit path to `rg`. `rg PATTERN .` or
  `rg PATTERN ./src` — never bare `rg PATTERN`. Bare invocations can
  treat stdin as input and hang the Shell tool.
- `rg --files` to list files
- `rg -l` to list matching files
- `rg -c` to count matches per file
- `rg -t go` for Go files only (use type filters over globs when
  possible)
- `rg --json` only when a downstream tool consumes it

## Examples

Wrong:

    Get-ChildItem -Recurse -Filter *.go | Select-String "func Apply"

Right:

    rg "func Apply" --type go .

Wrong:

    Select-String -Path . -Pattern "TODO"

Right:

    rg "TODO" .

## Verify

When the rule is active, this search over recent transcripts should
return zero matches for `Select-String` or `findstr`:

    rg "Select-String|findstr" .cursor/rules/

The rule itself mentions the blocked strings, so the check is only
useful over non-rule files. Use it on a fixture directory.
