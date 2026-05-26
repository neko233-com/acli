$ErrorActionPreference = 'Stop'

$Version = $args[0]

if (-not $Version) {
    Write-Host "Usage: update-version.ps1 <version>"
    exit 1
}

$Version = if ($Version -notlike 'v*') { "v$Version" } else { $Version }

# Update cmd/misc.go
$file = "cmd\misc.go"
$content = Get-Content $file -Raw
$content = $content -replace 'currentVersion := "[^"]+"', "currentVersion := `"$Version`""
Set-Content -Path $file -Value $content -NoNewline

Write-Host "Version updated to $Version"