# Build a local development binary as clockyDEV.exe (does not overwrite release clocky).
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$Version = if ($env:CLOCKY_DEV_VERSION) { $env:CLOCKY_DEV_VERSION } else { '0.0.0-dev' }
$Out = Join-Path $Root 'clockyDEV.exe'
$Ldflags = "-X github.com/LuizFer1/Clocky/internal/version.Version=$Version"

Write-Host "Building $Out (version $Version)..."
go build -ldflags $Ldflags -o $Out ./cmd/clocky
& $Out version
Write-Host ""
Write-Host "Run with: .\clockyDEV.exe <command>"
