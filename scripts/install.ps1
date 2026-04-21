param(
    [string]$Version = "",
    [string]$InstallDir = "$env:USERPROFILE\.tools\maple\bin"
)

$ErrorActionPreference = "Stop"
$Repo = "kinncj/AI-Squad"

if (-not $Version) {
    $latest = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $latest.tag_name
}

$Archive = "maple-windows-amd64.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$Archive"

Write-Host "Installing maple $Version"
Write-Host "  -> $InstallDir\maple.exe"
Write-Host ""

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    $archivePath = Join-Path $tmp $Archive
    Invoke-WebRequest $Url -OutFile $archivePath -UseBasicParsing
    Expand-Archive $archivePath -DestinationPath $tmp -Force
    Copy-Item (Join-Path $tmp "maple.exe") (Join-Path $InstallDir "maple.exe") -Force
} finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "✓ Installed maple $Version"
Write-Host ""
Write-Host "Add to PATH (run once in PowerShell as user):"
Write-Host "  [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$InstallDir', 'User')"
Write-Host ""
Write-Host "Verify with: maple --version"
