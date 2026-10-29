# Launch the Rust workspace from any working directory.
# Usage: powershell -File .\run.ps1 <command> [args]
$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path

function Show-Usage {
    Write-Host @"
usage: run.ps1 <command> [args]

  build                 cargo build --workspace
  test                  cargo test --workspace
  run <name> [--new]    cargo run -p qsm -- <name> [--new]
  view <name> [args]    cargo run -p qsm-view --features view -- <name> [args]

QSM_DATA_ROOT selects the data directory. When unset, these commands use
$root\data.
"@
    exit 2
}

if ($args.Count -lt 1) {
    Show-Usage
}

$commandName = $args[0]
$rest = @()
if ($args.Count -gt 1) {
    $rest = $args[1..($args.Count - 1)]
}

switch ($commandName) {
    'build' { }
    'test' { }
    'run' { }
    'view' { }
    default {
        Write-Host "run.ps1: unknown command $commandName" -ForegroundColor Red
        Show-Usage
    }
}

$caller = (Get-Location).Path

function Resolve-QsmPath([string]$value, [string]$fallback) {
    if ([string]::IsNullOrEmpty($value)) {
        return $fallback
    }
    if (-not [System.IO.Path]::IsPathRooted($value)) {
        return (Join-Path $caller $value)
    }
    return $value
}

$env:QSM_HOME = $root
$env:QSM_DATA_ROOT = Resolve-QsmPath $env:QSM_DATA_ROOT (Join-Path $root 'data')
$env:QSM_CONFIG_ROOT = Resolve-QsmPath $env:QSM_CONFIG_ROOT (Join-Path $root 'simulation-configs')

if ($commandName -eq 'run' -and $rest.Count -lt 1) {
    Write-Host 'usage: run.ps1 run <simulation name> [--new]'
    exit 2
}
if ($commandName -eq 'view' -and $rest.Count -lt 1) {
    Write-Host 'usage: run.ps1 view <simulation name> [--screenshot <png>]'
    exit 2
}

$cargoArgs = switch ($commandName) {
    'build' { @('build', '--workspace') + $rest }
    'test' { @('test', '--workspace') + $rest }
    'run' { @('run', '-p', 'qsm', '--') + $rest }
    'view' { @('run', '-p', 'qsm-view', '--features', 'view', '--') + $rest }
}

Write-Host "cargo $($cargoArgs -join ' ')"
Write-Host "data: $env:QSM_DATA_ROOT"

Push-Location (Join-Path $root 'rust')
try {
    & cargo @cargoArgs
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}
