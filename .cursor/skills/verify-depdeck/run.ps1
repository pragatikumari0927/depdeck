# Verify depdeck.exe against the nine-assertion map in SKILL.md.
# Arg 1: package.json path or its directory.
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$FixturePath
)

$ErrorActionPreference = 'Continue'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$WorkRoot = Join-Path $env:TEMP ("depdeck-verify-{0}" -f $PID)
$Bin = Join-Path $WorkRoot 'depdeck.exe'
$IsoDir = Join-Path $WorkRoot 'fixture'
$JsonDir = Join-Path $WorkRoot 'json-only'
$FlaggedIsoDir = Join-Path $WorkRoot 'flagged'
$EmptyDir = Join-Path $RepoRoot 'testdata\verify\empty'
$failures = 0

function Write-StepResult {
    param(
        [string]$Step,
        [string]$Name,
        [string]$Command,
        [bool]$Pass,
        [int]$ExitCode,
        [string]$Stdout,
        [string]$Stderr,
        [string]$Note
    )
    $tag = if ($Pass) { 'PASS' } else { 'FAIL' }
    if (-not $Pass) { $script:failures++ }
    Write-Output ("{0} {1}: {2}" -f $tag, $Step, $Name)
    Write-Output ("  command: {0}" -f $Command)
    Write-Output ("  exit: {0}" -f $ExitCode)
    if ($Stdout) {
        Write-Output '  stdout:'
        Write-Output $Stdout
    }
    if (-not $Pass -or $Note) {
        if ($Stderr) {
            Write-Output '  stderr:'
            Write-Output $Stderr
        }
    }
    if ($Note) { Write-Output ("  note: {0}" -f $Note) }
}

function Invoke-Depdeck {
    param(
        [string[]]$Args,
        [string]$WorkDir,
        [hashtable]$ExtraEnv
    )
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $Bin
    foreach ($a in $Args) { [void]$psi.ArgumentList.Add($a) }
    $psi.WorkingDirectory = $WorkDir
    $psi.UseShellExecute = $false
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.CreateNoWindow = $true
    if ($ExtraEnv) {
        foreach ($k in $ExtraEnv.Keys) {
            $psi.Environment[$k] = [string]$ExtraEnv[$k]
        }
    }
    $p = [System.Diagnostics.Process]::Start($psi)
    $out = $p.StandardOutput.ReadToEnd()
    $err = $p.StandardError.ReadToEnd()
    $p.WaitForExit()
    [pscustomobject]@{
        ExitCode = $p.ExitCode
        Stdout   = $out
        Stderr   = $err
        Command  = ('{0} {1}' -f $Bin, ($Args -join ' '))
    }
}

function Resolve-FixtureFile {
    param([string]$Path)
    $full = Resolve-Path -LiteralPath $Path -ErrorAction SilentlyContinue
    if (-not $full) { return $null }
    $item = Get-Item -LiteralPath $full.Path
    if ($item.PSIsContainer) {
        $pkg = Join-Path $item.FullName 'package.json'
        if (Test-Path -LiteralPath $pkg) { return $pkg }
        return $null
    }
    return $item.FullName
}

$srcPkg = Resolve-FixtureFile $FixturePath
if (-not $srcPkg) {
    Write-Output "FAIL 0: fixture not found: $FixturePath"
    exit 1
}

$flaggedSrc = Join-Path $RepoRoot 'testdata\verify\flagged\package.json'
if (-not (Test-Path -LiteralPath $flaggedSrc)) {
    Write-Output "FAIL 0: flagged fixture not found: $flaggedSrc"
    exit 1
}
if (-not (Test-Path -LiteralPath $EmptyDir)) {
    Write-Output "FAIL 0: empty fixture not found: $EmptyDir"
    exit 1
}

New-Item -ItemType Directory -Path $IsoDir, $JsonDir, $FlaggedIsoDir -Force | Out-Null
Copy-Item -LiteralPath $srcPkg -Destination (Join-Path $IsoDir 'package.json')
Copy-Item -LiteralPath $srcPkg -Destination (Join-Path $JsonDir 'package.json')
Copy-Item -LiteralPath $flaggedSrc -Destination (Join-Path $FlaggedIsoDir 'package.json')

$cmdMain = Join-Path $RepoRoot 'cmd\depdeck\main.go'
if (-not (Test-Path -LiteralPath $cmdMain)) {
    Write-Output "FAIL 0: doctor: cmd/depdeck missing ($cmdMain). Do not substitute go test."
    Remove-Item -LiteralPath $WorkRoot -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

Push-Location $RepoRoot
try {
    $build = & go build -o $Bin ./cmd/depdeck 2>&1
    $buildExit = $LASTEXITCODE
} finally {
    Pop-Location
}
if ($buildExit -ne 0 -or -not (Test-Path -LiteralPath $Bin)) {
    Write-Output 'FAIL 0: go build -o depdeck.exe ./cmd/depdeck'
    Write-Output '  stderr:'
    Write-Output ($build | Out-String)
    Remove-Item -LiteralPath $WorkRoot -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# 1. scan writes deck.html
$r1 = Invoke-Depdeck -Args @('scan', $IsoDir) -WorkDir $IsoDir
$deck1 = Join-Path $IsoDir 'deck.html'
$ok1 = ($r1.ExitCode -eq 0) -and (Test-Path -LiteralPath $deck1)
Write-StepResult -Step 1 -Name 'scan writes deck.html next to package.json' -Command $r1.Command -Pass $ok1 -ExitCode $r1.ExitCode -Stdout $r1.Stdout -Stderr $r1.Stderr -Note $(if (-not (Test-Path -LiteralPath $deck1)) { "missing $deck1" } else { $null })

# 2. byte-identical second scan
$hash1 = $null
if (Test-Path -LiteralPath $deck1) { $hash1 = (Get-FileHash -LiteralPath $deck1 -Algorithm SHA256).Hash }
$r2 = Invoke-Depdeck -Args @('scan', $IsoDir) -WorkDir $IsoDir
$hash2 = $null
if (Test-Path -LiteralPath $deck1) { $hash2 = (Get-FileHash -LiteralPath $deck1 -Algorithm SHA256).Hash }
$ok2 = ($r2.ExitCode -eq 0) -and $hash1 -and ($hash1 -eq $hash2)
Write-StepResult -Step 2 -Name 'deck.html byte-identical across two scans' -Command $r2.Command -Pass $ok2 -ExitCode $r2.ExitCode -Stdout $r2.Stdout -Stderr $r2.Stderr -Note $(if ($hash1 -and $hash2 -and $hash1 -ne $hash2) { "hash1=$hash1 hash2=$hash2" } else { $null })

# 3. scan --json envelope
$r3 = Invoke-Depdeck -Args @('scan', '--json', $IsoDir) -WorkDir $IsoDir
$envOk = $false
$envNote = $null
try {
    $obj = $r3.Stdout | ConvertFrom-Json
    $need = @('schema_version', 'generated_by', 'generated_at', 'tool', 'data')
    $missing = @($need | Where-Object { -not ($obj.PSObject.Properties.Name -contains $_) })
    if ($missing.Count -eq 0) { $envOk = $true } else { $envNote = "missing keys: $($missing -join ', ')" }
} catch {
    $envNote = $_.Exception.Message
}
$ok3 = ($r3.ExitCode -eq 0) -and $envOk
Write-StepResult -Step 3 -Name 'scan --json envelope on stdout' -Command $r3.Command -Pass $ok3 -ExitCode $r3.ExitCode -Stdout $r3.Stdout -Stderr $r3.Stderr -Note $envNote

# 4. scan --json does not write deck.html
$r4 = Invoke-Depdeck -Args @('scan', '--json', $JsonDir) -WorkDir $JsonDir
$jsonDeck = Join-Path $JsonDir 'deck.html'
$ok4 = ($r4.ExitCode -eq 0) -and -not (Test-Path -LiteralPath $jsonDeck)
Write-StepResult -Step 4 -Name 'scan --json does not write deck.html' -Command $r4.Command -Pass $ok4 -ExitCode $r4.ExitCode -Stdout $r4.Stdout -Stderr $r4.Stderr -Note $(if (Test-Path -LiteralPath $jsonDeck) { "wrote $jsonDeck" } else { $null })

# 5a. check clean fixture → exit 0
$r5a = Invoke-Depdeck -Args @('check', $IsoDir) -WorkDir $IsoDir
$ok5a = ($r5a.ExitCode -eq 0)
$note5a = $null
if ($r5a.ExitCode -eq 3) {
    $note5a = 'exit 3 per SPEC — cannot prove exit 0/1 without network. Re-run when npm is reachable.'
}
Write-StepResult -Step '5a' -Name 'check clean fixture exits 0' -Command $r5a.Command -Pass $ok5a -ExitCode $r5a.ExitCode -Stdout $r5a.Stdout -Stderr $r5a.Stderr -Note $note5a

# 5b. check flagged fixture → exit 1
$r5b = Invoke-Depdeck -Args @('check', $FlaggedIsoDir) -WorkDir $FlaggedIsoDir
$ok5b = ($r5b.ExitCode -eq 1)
$note5b = $null
if ($r5b.ExitCode -eq 3) {
    $note5b = 'exit 3 per SPEC — cannot prove exit 0/1 without network. Re-run when npm is reachable.'
}
Write-StepResult -Step '5b' -Name 'check flagged fixture exits 1' -Command $r5b.Command -Pass $ok5b -ExitCode $r5b.ExitCode -Stdout $r5b.Stdout -Stderr $r5b.Stderr -Note $note5b

# 6. --flavor=ai
$r6 = Invoke-Depdeck -Args @('scan', '--flavor=ai', $IsoDir) -WorkDir $IsoDir
$hasAi = $r6.Stderr -match 'ai'
$hasV11 = $r6.Stderr -match 'v1\.1'
$ok6 = ($r6.ExitCode -eq 2) -and $hasAi -and $hasV11
Write-StepResult -Step 6 -Name '--flavor=ai exits 2 with ai and v1.1' -Command $r6.Command -Pass $ok6 -ExitCode $r6.ExitCode -Stdout $r6.Stdout -Stderr $r6.Stderr -Note $(if (-not $ok6) { "ai=$hasAi v1.1=$hasV11" } else { $null })

# 7. missing package.json → exit 2
$r7 = Invoke-Depdeck -Args @('scan', $EmptyDir) -WorkDir $EmptyDir
$ok7 = ($r7.ExitCode -eq 2)
Write-StepResult -Step 7 -Name 'missing package.json exits 2' -Command $r7.Command -Pass $ok7 -ExitCode $r7.ExitCode -Stdout $r7.Stdout -Stderr $r7.Stderr -Note $(if ($r7.ExitCode -eq 3) { 'exit 3 is check all-fetch-failed, not a missing manifest' } else { $null })

# 8. NO_COLOR=1 — no ANSI in stderr
$r8 = Invoke-Depdeck -Args @('scan', $IsoDir) -WorkDir $IsoDir -ExtraEnv @{ NO_COLOR = '1' }
$hasAnsi = $r8.Stderr -match "`e\["
if (-not $hasAnsi) { $hasAnsi = $r8.Stderr -match [char]27 + '\[' }
$ok8 = -not $hasAnsi
Write-StepResult -Step 8 -Name 'NO_COLOR=1 produces no ANSI in stderr' -Command $r8.Command -Pass $ok8 -ExitCode $r8.ExitCode -Stdout $r8.Stdout -Stderr $r8.Stderr -Note $(if ($hasAnsi) { 'CSI sequence found in stderr' } else { $null })

Remove-Item -LiteralPath $WorkRoot -Recurse -Force -ErrorAction SilentlyContinue
exit $(if ($failures -gt 0) { 1 } else { 0 })
