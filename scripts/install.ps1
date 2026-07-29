$repo = "Bhagattji/momo"
$asset = "momo-windows-amd64.exe"
$installDir = "$env:USERPROFILE\bin"
$exePath = "$installDir\momo.exe"

Write-Host "Installing momo from $repo..." -Fore Green

New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$api = "https://api.github.com/repos/$repo/releases/latest"
$release = Invoke-RestMethod -Uri $api -Headers @{Accept="application/vnd.github.v3+json"}
$url = $release.assets | Where-Object { $_.name -eq $asset } | Select-Object -ExpandProperty browser_download_url

if (-not $url) {
    Write-Error "Asset $asset not found in latest release"
    exit 1
}

Write-Host "Downloading $url..." -Fore Cyan
Invoke-WebRequest -Uri $url -OutFile $exePath

$env:PATH += ";$installDir"
[Environment]::SetEnvironmentVariable("PATH", [Environment]::GetEnvironmentVariable("PATH","User") + ";$installDir", "User")

Write-Host "Installed to $exePath" -Fore Green
Write-Host "Run 'momo --version' to verify" -Fore Cyan
& $exePath --version
