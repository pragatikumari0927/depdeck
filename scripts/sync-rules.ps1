# Regenerate .grok/rules/*.md from .cursor/rules/*.mdc
# Run from repo root: pwsh -File scripts/sync-rules.ps1

$ErrorActionPreference = 'Stop'
$src = Join-Path $PSScriptRoot '..\.cursor\rules'
$dst = Join-Path $PSScriptRoot '..\.grok\rules'

if (-not (Test-Path $dst)) {
    New-Item -ItemType Directory -Path $dst -Force | Out-Null
}

Get-ChildItem $src -Filter *.mdc | ForEach-Object {
    $name = $_.BaseName
    $body = Get-Content $_.FullName -Raw

    if ($body -match '(?s)^---\s*\r?\n.*?\r?\n---\s*\r?\n') {
        $body = $body -replace '(?s)^---\s*\r?\n.*?\r?\n---\s*\r?\n', ''
    }

    $out = Join-Path $dst "$name.md"
    Set-Content -Path $out -Value $body -NoNewline -Encoding utf8
    Write-Host "synced $name.mdc -> .grok\rules\$name.md"
}

Write-Host "Done."
