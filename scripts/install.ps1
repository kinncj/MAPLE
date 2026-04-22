param(
    [string]$Version    = "",
    [string]$InstallDir = "$env:USERPROFILE\.tools\maple\bin",
    [switch]$SkipRtk
)

$ErrorActionPreference = "Stop"
$Repo    = "kinncj/MAPLE"
$RtkRepo = "rtk-ai/rtk"

# ── Resolve maple version (semver sort, not /releases/latest) ──────────────────
if (-not $Version) {
    $releases = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases?per_page=100"
    $Version = $releases |
        Where-Object { $_.tag_name -match '^v\d+\.\d+\.\d+$' -and -not $_.prerelease -and -not $_.draft } |
        Sort-Object { [System.Version]($_.tag_name -replace '^v','') } |
        Select-Object -Last 1 -ExpandProperty tag_name
}
if (-not $Version) {
    Write-Error "Could not determine latest version. Pass -Version vX.Y.Z"
    exit 1
}

# ── Install maple ──────────────────────────────────────────────────────────────
$Archive = "maple-windows-amd64.zip"
$Url     = "https://github.com/$Repo/releases/download/$Version/$Archive"

Write-Host "Installing maple $Version (windows/amd64)"
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
    Write-Host "✓ Installed maple $Version"
    Write-Host ""

    # ── Install RTK token optimizer ────────────────────────────────────────────
    # Reduces LLM token consumption 60-90% via PreToolUse hook interception.
    # On Windows without WSL, falls back to CLAUDE.md instruction injection.
    if (-not $SkipRtk) {
        if (Get-Command rtk -ErrorAction SilentlyContinue) {
            Write-Host "✓ rtk already installed"
        } else {
            Write-Host "Installing rtk token optimizer..."
            try {
                $rtkReleases = Invoke-RestMethod "https://api.github.com/repos/$RtkRepo/releases?per_page=100"
                $rtkVersion  = ($rtkReleases | Sort-Object { [version]($_.tag_name -replace '^v','') } | Select-Object -Last 1).tag_name
                $rtkArchive  = "rtk-x86_64-pc-windows-msvc.zip"
                $rtkUrl      = "https://github.com/$RtkRepo/releases/download/$rtkVersion/$rtkArchive"
                $rtkZip      = Join-Path $tmp $rtkArchive
                $rtkExtract  = Join-Path $tmp "rtk-extract"

                Invoke-WebRequest $rtkUrl -OutFile $rtkZip -UseBasicParsing
                Expand-Archive $rtkZip -DestinationPath $rtkExtract -Force

                # locate rtk.exe anywhere in the extracted tree (handles subdirectory layouts)
                $rtkBin = Get-ChildItem -Path $rtkExtract -Filter "rtk.exe" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
                if ($null -eq $rtkBin) {
                    Write-Host "~ rtk: binary not found in archive — install manually: https://github.com/rtk-ai/rtk"
                } else {
                    Copy-Item $rtkBin.FullName (Join-Path $InstallDir "rtk.exe") -Force
                    Write-Host "✓ Installed rtk $rtkVersion"
                }
            } catch {
                Write-Host "~ rtk install skipped — download failed (install manually: https://github.com/rtk-ai/rtk)"
            }
        }
        Write-Host ""
    }
} finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

# ── PATH reminder ──────────────────────────────────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    Write-Host "Add to PATH (run once in PowerShell as user):"
    Write-Host "  [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$InstallDir', 'User')"
    Write-Host ""
}

Write-Host "Verify with: maple --version"
