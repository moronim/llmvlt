$ErrorActionPreference = "Stop"

$Repo = "moronim/llmvlt"
$Binary = "llmvlt.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "llmvlt"

function Write-Info($msg) { Write-Host "==> " -ForegroundColor Green -NoNewline; Write-Host $msg }
function Write-Err($msg)  { Write-Host "error: " -ForegroundColor Red -NoNewline; Write-Host $msg; exit 1 }

function Test-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Write-Err "Missing required command: $Name"
    }
}

$RawArch = if ($env:PROCESSOR_ARCHITECTURE) {
    $env:PROCESSOR_ARCHITECTURE
} elseif ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
} else {
    "unknown"
}

$Arch = switch ($RawArch.ToUpper()) {
    "AMD64"  { "amd64" }
    "X64"    { "amd64" }
    "X86"    { Write-Err "32-bit Windows is not supported."; "" }
    "ARM64"  { Write-Err "Windows ARM64 is not supported yet. Build from source."; "" }
    default  { Write-Err "Unsupported architecture: $RawArch"; "" }
}

Write-Info "Checking latest version..."
try {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{ "User-Agent" = "llmvlt-installer" }
    $Version = $Release.tag_name
} catch {
    Write-Err "Could not fetch latest release. Check https://github.com/$Repo/releases"
}

$AssetName = "llmvlt-windows-$Arch.exe"
$Url = "https://github.com/$Repo/releases/download/$Version/$AssetName"

$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("llmvlt-install-" + [guid]::NewGuid().ToString("N"))
$null = New-Item -ItemType Directory -Path $TmpDir -Force
$TmpFile = Join-Path $TmpDir $Binary
$DestFile = Join-Path $InstallDir $Binary

Write-Info "Downloading llmvlt $Version for windows/$Arch..."
Write-Host "  $Url" -ForegroundColor DarkGray

try {
    Invoke-WebRequest -Uri $Url -OutFile $TmpFile -UseBasicParsing
} catch {
    Write-Err "Download failed. Check https://github.com/$Repo/releases/tag/$Version"
}

try {
    if (Get-Command Unblock-File -ErrorAction SilentlyContinue) {
        Unblock-File -Path $TmpFile -ErrorAction SilentlyContinue
    }
} catch {}

try {
    & $TmpFile --version | Out-Null
} catch {
    try {
        & $TmpFile --help | Out-Null
    } catch {
        Write-Err "Downloaded binary failed to execute."
    }
}

if (-not (Test-Path $InstallDir)) {
    $null = New-Item -ItemType Directory -Path $InstallDir -Force
}

Copy-Item -Path $TmpFile -Destination $DestFile -Force
Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
$PathParts = @()
if ($UserPath) { $PathParts = $UserPath -split ';' | Where-Object { $_ } }

if ($PathParts -notcontains $InstallDir) {
    $NewUserPath = if ([string]::IsNullOrWhiteSpace($UserPath)) { $InstallDir } else { "$UserPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $NewUserPath, "User")
    if (($env:Path -split ';' | Where-Object { $_ }) -notcontains $InstallDir) {
        $env:Path = "$env:Path;$InstallDir"
    }
    Write-Info "Added $InstallDir to your PATH."
}

Write-Info "Installed llmvlt $Version to $DestFile"
Write-Host ""
Write-Host "  Run " -NoNewline
Write-Host "llmvlt --help" -ForegroundColor Green -NoNewline
Write-Host " to get started."
Write-Host ""
