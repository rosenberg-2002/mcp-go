# MCP-Go Cross-Platform Build Script
# Usage: .\build.ps1

$ErrorActionPreference = "Stop"
$OutputDir = "dist"
$Module = "./cmd/server"
$AppName = "mcp-server"

# Clean previous builds
if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Path $OutputDir | Out-Null

Write-Host "`n🔨 Building MCP-Go for all platforms...`n" -ForegroundColor Cyan

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
    @{ GOOS = "darwin";  GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "darwin";  GOARCH = "arm64"; Ext = "" },
    @{ GOOS = "linux";   GOARCH = "amd64"; Ext = "" }
)

foreach ($target in $targets) {
    $outFile = "$OutputDir/$AppName-$($target.GOOS)-$($target.GOARCH)$($target.Ext)"
    Write-Host "  Building $($target.GOOS)/$($target.GOARCH)... " -NoNewline

    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    $env:CGO_ENABLED = "0"

    go build -ldflags="-s -w" -o $outFile $Module

    if ($LASTEXITCODE -eq 0) {
        $size = [math]::Round((Get-Item $outFile).Length / 1MB, 2)
        Write-Host "✅ ($size MB)" -ForegroundColor Green
    } else {
        Write-Host "❌ FAILED" -ForegroundColor Red
    }
}

# Reset environment
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host "`n✅ All builds complete! Output in ./$OutputDir/`n" -ForegroundColor Green
Get-ChildItem $OutputDir | Format-Table Name, @{N="Size (MB)";E={[math]::Round($_.Length/1MB, 2)}} -AutoSize
