@echo off
setlocal

:: unicli - Auto Deploy to GitHub
:: Usage: deploy-to-github.cmd [version]
::   If no version provided, auto-increments patch version

set REPO=neko233-com/unicli
set BRANCH=main

set CURRENT_VERSION=
if exist "version.txt" (
    set /p CURRENT_VERSION=<version.txt
)

if "%~1"=="" (
    if "!CURRENT_VERSION!"=="" (
        set NEW_VERSION=v1.0.0
    ) else (
        for /f "tokens=1,2,3 delims=." %%a in ("!CURRENT_VERSION!") do (
            set /a PATCH=%%c+1
            set NEW_VERSION=v%%a.%%b.!PATCH!
        )
    )
) else (
    set NEW_VERSION=%~1
    if "!NEW_VERSION:~0,1!" neq "v" set NEW_VERSION=v!NEW_VERSION!
)

echo ========================================
echo   unicli Auto Deploy to GitHub
echo ========================================
echo Repository:  %REPO%
echo New Version: %NEW_VERSION%
echo Branch:      %BRANCH%
echo ========================================
echo.

echo [1/6] Checking git status...
git status --porcelain >nul 2>&1
if errorlevel 1 (
    echo Initializing git repository...
    git init -b %BRANCH% >nul 2>&1
)

echo [2/6] Auto-adding all files...
git add -A

echo [3/6] Checking remote...
git remote get-url origin >nul 2>&1
if errorlevel 1 (
    echo Please add remote first:
    echo   git remote add origin https://github.com/%REPO%.git
    exit /b 1
)

echo [4/6] Updating version to %NEW_VERSION%...
echo %NEW_VERSION% > version.txt
powershell -ExecutionPolicy Bypass -File "scripts\update-version.ps1" "%NEW_VERSION%"

echo [5/6] Committing changes...
git add -A
git commit -m "chore: release %NEW_VERSION%"

echo [6/6] Pushing tag %NEW_VERSION%...
git tag -a %NEW_VERSION% -m "Release %NEW_VERSION%" 2>nul
if errorlevel 1 (
    git tag -d %NEW_VERSION% 2>nul
    git tag -a %NEW_VERSION% -m "Release %NEW_VERSION%"
)

git push origin %BRANCH%
if errorlevel 1 (
    echo ERROR: Failed to push branch
    exit /b 1
)

git push origin %NEW_VERSION%
if errorlevel 1 (
    echo ERROR: Failed to push tag
    exit /b 1
)

echo.
echo ========================================
echo   Deploy Complete!
echo ========================================
echo Tag %NEW_VERSION% pushed.
echo Check: https://github.com/%REPO%/actions
echo ========================================

endlocal