param(
    [Parameter(Position = 0)]
    [string]$Version = "latest"
)

$ErrorActionPreference = 'Stop'

# unicli - Windows installer
# Usage: irm https://raw.githubusercontent.com/neko233-com/unicli/main/scripts/install.ps1 | iex
# Or:    irm .../install.ps1 -OutFile install.ps1; .\install.ps1 v1.0.0

$BinaryName = "unicli"
$Repo = "neko233-com/unicli"
$InstallDir = Join-Path $env:LOCALAPPDATA $BinaryName
$Asset = "${BinaryName}-windows-amd64.exe"

function Get-NormalizedVersion([string]$Value) {
    $v = $Value.Trim()
    while ($v.StartsWith('v') -or $v.StartsWith('V')) { $v = $v.Substring(1) }
    return $v
}

if ($Version -eq "latest" -or [string]::IsNullOrWhiteSpace($Version)) {
    $url = "https://github.com/$Repo/releases/latest/download/$Asset"
    $versionLabel = "latest"
} else {
    $Version = Get-NormalizedVersion $Version
    $url = "https://github.com/$Repo/releases/download/v$Version/$Asset"
    $versionLabel = "v$Version"
}

$dest = Join-Path $InstallDir "$BinaryName.exe"

Write-Host "Installing ${BinaryName} $versionLabel for windows/amd64..."
Write-Host "Downloading $url..."

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Invoke-WebRequest -Uri $url -OutFile $dest

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "Added $InstallDir to user PATH."
}

Write-Host ""
Write-Host "Installed to $dest"
Write-Host "Run: unicli --help"
