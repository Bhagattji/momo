# momo install script
# Usage: irm https://raw.githubusercontent.com/Bhagattji/momo/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"
$repo = "Bhagattji/momo"
$installDir = "$env:USERPROFILE\.momo\bin"
$exePath = Join-Path $installDir "momo.exe"

Write-Host ""
Write-Host "  __  __                  " -ForegroundColor Cyan
Write-Host " |  \/  | ___  _ __ ___   " -ForegroundColor Cyan
Write-Host " | |\/| |/ _ \| '_ \` _ \ " -ForegroundColor Cyan
Write-Host " | |  | | (_) | | | | | |  " -ForegroundColor Cyan
Write-Host " |_|  |_|\___/|_| |_| |_| " -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$downloadAsset = "momo_windows_amd64.zip"

if (Test-Path $exePath) {
    try {
        $currentVer = (& $exePath --version 2>$null | Select-Object -First 1)
        if ($currentVer) {
            Write-Host "  momo $currentVer is already installed." -ForegroundColor Yellow
            $confirm = Read-Host "  Update to latest? (y/N)"
            if ($confirm -notmatch '^[Yy]') {
                Write-Host "  Cancelled." -ForegroundColor DarkGray
                return
            }
        }
    } catch {
        Write-Host "  Existing binary unreadable, will reinstall." -ForegroundColor Yellow
    }
}

Write-Host "  Checking latest release..." -ForegroundColor DarkGray
try {
    $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest" -Headers @{"User-Agent"="momo-installer"}
} catch {
    Write-Error "  Failed to fetch release info. Check your internet connection."
    Write-Error "  $_"
    exit 1
}

$downloadUrl = $null
$assetExt = ".zip"
$assetName = "momo_windows_amd64.zip"

foreach ($asset in $release.assets) {
    $name = $asset.name
    if ($name -eq $assetName) {
        $downloadUrl = $asset.browser_download_url
        break
    }
    if ($name -like "*windows*amd64*" -and -not $downloadUrl) {
        $downloadUrl = $asset.browser_download_url
        $assetName = $name
        if ($name -like "*.tar.gz") { $assetExt = ".tar.gz" }
    }
}

if (-not $downloadUrl) {
    foreach ($asset in $release.assets) {
        if ($asset.name -like "*.zip" -or $asset.name -like "*.tar.gz") {
            $downloadUrl = $asset.browser_download_url
            $assetName = $asset.name
            if ($assetName -like "*.tar.gz") { $assetExt = ".tar.gz" }
            break
        }
    }
}

if (-not $downloadUrl) {
    Write-Error "No Windows binary found in release $($release.tag_name)"
    Write-Host "Available assets:" -ForegroundColor DarkGray
    foreach ($asset in $release.assets) {
        Write-Host "  - $($asset.name)" -ForegroundColor DarkGray
    }
    exit 1
}

$tag = $release.tag_name
Write-Host "  Downloading momo $tag ..." -ForegroundColor Yellow

$tempDir = Join-Path $env:TEMP "momo-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
$archivePath = Join-Path $tempDir "momo$assetExt"

try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing
} catch {
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
    Write-Error "Download failed: $_"
    exit 1
}

try {
    Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force
    $extracted = Get-ChildItem -Path $tempDir -Recurse -Filter "momo.exe" | Select-Object -First 1
    if (-not $extracted) {
        $extracted = Get-ChildItem -Path $tempDir -Recurse -Filter "momo" | Select-Object -First 1
    }
    if (-not $extracted) {
        Write-Error "momo binary not found in archive"
        exit 1
    }
    if (Test-Path $exePath) {
        Remove-Item $exePath -Force
    }
    Move-Item $extracted.FullName $exePath -Force
} finally {
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathDir = (Split-Path $exePath -Parent).TrimEnd('\')
if ($userPath -notlike "*$pathDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$pathDir", "User")
    Write-Host "  Added $pathDir to user PATH." -ForegroundColor Green
}

$env:Path = "$env:Path;$pathDir"

$ver = & $exePath --version 2>$null
Write-Host ""
Write-Host "  momo $ver installed successfully!" -ForegroundColor Green
Write-Host "  Binary: $exePath" -ForegroundColor DarkGray
Write-Host ""
Write-Host "  Open a NEW terminal and type:" -ForegroundColor Cyan
Write-Host "    momo" -ForegroundColor White
Write-Host ""