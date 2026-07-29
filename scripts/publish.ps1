<#
publish.ps1
Usage: run from repo root (PowerShell)
  .\scripts\publish.ps1 -RepoName "ORG/momo" -Visibility private -Push

This script:
 - checks for git and gh
 - initializes git if needed
 - creates initial commit
 - creates a private GitHub repo via `gh` if available, else prints instructions
 - pushes to origin/main

Be careful: this will publish code to the remote you specify. For closed-source, ensure repo is private and CI secrets are set.
#>
param(
    [string]$RepoName = "",
    [ValidateSet('private','public')][string]$Visibility = 'private',
    [switch]$Push
)

function Need-Command($name) {
    $p = Get-Command $name -ErrorAction SilentlyContinue
    return $p -ne $null
}

if (-not (Need-Command git)) {
    Write-Error "git is not installed or not in PATH. Install Git and re-run this script."
    exit 1
}

$root = Get-Location
Write-Host "Repo root: $root"

if (-not (Test-Path .git)) {
    Write-Host "Initializing git repository..."
    git init
} else {
    Write-Host ".git already exists; skipping git init"
}

Write-Host "Staging files..."
git add .

if (-not (git rev-parse --verify HEAD 2>$null)) {
    git commit -m "Initial private momo scaffold"
} else {
    Write-Host "A commit already exists; creating a new commit"
    git commit -m "Initial private momo scaffold" --allow-empty
}

git branch -M main

if ($RepoName -ne "") {
    if (Need-Command gh) {
        Write-Host "Creating GitHub repo via gh: $RepoName ($Visibility)"
        gh repo create $RepoName --$Visibility --source=. --remote=origin --push
    } else {
        Write-Host "gh CLI not found. Create a private repo named '$RepoName' on GitHub and add remote:"
        Write-Host "  git remote add origin git@github.com:$RepoName.git"
        Write-Host "  git push -u origin main"
        if ($Push) {
            Write-Host "Pushing to remote origin/main..."
            git push -u origin main
        }
    }
} else {
    Write-Host "No RepoName provided. To push manually, run:"
    Write-Host "  git remote add origin <git@github.com:ORG/repo.git>"
    Write-Host "  git push -u origin main"
}

Write-Host "Done. Reminder: Set repository secrets (GROQ_API_KEY, ANTHROPIC_API_KEY, GPG signing keys etc.) in repo settings before running CI that requires them."