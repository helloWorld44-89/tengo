# tengo installer for Windows
# Usage: powershell -ExecutionPolicy Bypass -c "irm https://raw.githubusercontent.com/helloWorld44-89/tengo/main/install.ps1 | iex"

$ErrorActionPreference = 'Stop'

$Repo      = "helloWorld44-89/tengo"
$Binary    = "tengo.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\tengo"

function Write-Info  { param($msg) Write-Host "  • $msg" -ForegroundColor Cyan }
function Write-Ok    { param($msg) Write-Host "  ✓ $msg" -ForegroundColor Green }
function Write-Warn  { param($msg) Write-Host "  ! $msg" -ForegroundColor Yellow }
function Write-Fail  { param($msg) Write-Host "  ✗ $msg" -ForegroundColor Red; exit 1 }

# ── Detect architecture ───────────────────────────────────────────────────────

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { Write-Fail "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

# Only amd64 has a prebuilt binary right now.
if ($arch -ne "amd64") {
    Write-Fail "No prebuilt binary for Windows/$arch yet. Build from source: https://github.com/$Repo"
}

$Asset = "tengo-windows-$arch.exe"

Write-Host ""
Write-Host "  tengo installer" -ForegroundColor White
Write-Host ""
Write-Info "Platform: windows/$arch"
Write-Info "Install directory: $InstallDir"

# ── Download ──────────────────────────────────────────────────────────────────

$Url  = "https://github.com/$Repo/releases/latest/download/$Asset"
$Dest = Join-Path $InstallDir $Binary

Write-Info "Downloading latest release..."

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

try {
    Invoke-WebRequest -Uri $Url -OutFile $Dest -UseBasicParsing
} catch {
    Write-Fail "Download failed: $_"
}

# ── Verify ────────────────────────────────────────────────────────────────────

try {
    $Ver = & $Dest -version 2>&1
} catch {
    Write-Fail "Downloaded binary failed to run: $_"
}

Write-Ok "Installed: $Ver"

# ── Add to PATH ───────────────────────────────────────────────────────────────

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")

if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    Write-Ok "Added $InstallDir to your PATH"
    Write-Warn "Restart your terminal for PATH changes to take effect."
} else {
    Write-Info "$InstallDir is already in PATH"
}

# ── Completion hint ───────────────────────────────────────────────────────────

Write-Host ""
Write-Host "  Optional — add shell completions (Git Bash / WSL):" -ForegroundColor DarkGray
Write-Host "    bash:  echo 'source <(tengo -completion bash)' >> ~/.bashrc" -ForegroundColor DarkGray
Write-Host ""
