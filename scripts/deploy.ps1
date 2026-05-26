param([string]$VersionArg = "")

$ErrorActionPreference = 'Stop'

$Repo = "neko233-com/unicli"
$Branch = "main"

function Normalize-Version([string]$v) {
    $v = $v.Trim()
    while ($v.StartsWith('v')) { $v = $v.Substring(1) }
    return $v
}

function Format-Version([string]$v) {
    return "v$(Normalize-Version $v)"
}

$current = ""
if (Test-Path "version.txt") {
    $current = (Get-Content "version.txt" -Raw).Trim()
}

if ($VersionArg) {
    $newVersion = Format-Version $VersionArg
} elseif (-not $current) {
    $newVersion = "v1.0.0"
} else {
    $base = Normalize-Version $current
    $parts = $base -split '\.'
    if ($parts.Length -ge 3) {
        $patch = [int]$parts[2] + 1
        $newVersion = "v$($parts[0]).$($parts[1]).$patch"
    } else {
        $newVersion = "v1.0.0"
    }
}

Write-Host "========================================"
Write-Host "  unicli Auto Deploy to GitHub"
Write-Host "========================================"
Write-Host "Repository:  $Repo"
Write-Host "New Version: $newVersion"
Write-Host "Branch:      $Branch"
Write-Host "========================================"
Write-Host ""

Write-Host "[1/6] Checking git..."
if (-not (Test-Path ".git")) {
    git init -b $Branch | Out-Null
}

Write-Host "[2/6] Auto-adding all files..."
git add -A

Write-Host "[3/6] Checking remote..."
$remote = git remote get-url origin 2>$null
if (-not $remote) {
    Write-Host "Please add remote first:"
    Write-Host "  git remote add origin https://github.com/$Repo.git"
    exit 1
}

Write-Host "[4/6] Updating version to $newVersion..."
Set-Content -Path "version.txt" -Value $newVersion -NoNewline

$file = "cmd\misc.go"
$content = Get-Content $file -Raw
$content = $content -replace 'currentVersion := "[^"]+"', "currentVersion := `"$newVersion`""
Set-Content -Path $file -Value $content -NoNewline

Write-Host "[5/6] Committing..."
git add -A
git commit -m "chore: release $newVersion"

Write-Host "[6/6] Pushing tag $newVersion..."
$prevEA = $ErrorActionPreference
$ErrorActionPreference = 'Continue'
git tag -d $newVersion 2>$null | Out-Null
$ErrorActionPreference = $prevEA

git tag -a $newVersion -m "Release $newVersion" 2>$null
if ($LASTEXITCODE -ne 0) {
    git tag -f -a $newVersion -m "Release $newVersion"
}

git push origin $Branch
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

git push origin $newVersion
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "========================================"
Write-Host "  Deploy Complete!"
Write-Host "========================================"
Write-Host "Tag $newVersion pushed."
Write-Host "Check: https://github.com/$Repo/actions"
Write-Host "========================================"
