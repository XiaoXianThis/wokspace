# Install wok on Windows.
#
#   irm https://wokspace.wxian.cc/install.ps1 | iex
#
# Environment:
#   WOK_VERSION      release tag to install, such as v0.1.0 (default: latest)
#   WOK_INSTALL_DIR  where to put wok.exe (default: %LOCALAPPDATA%\Programs\wok)
#   WOK_DOWNLOAD_BASE  download from this URL instead of GitHub releases (a mirror)
$ErrorActionPreference = 'Stop'

$repo = 'XiaoXianThis/wokspace'
$dir = if ($env:WOK_INSTALL_DIR) { $env:WOK_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\wok' }

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
  'AMD64' { 'amd64' }
  'ARM64' { 'arm64' }
  default { throw "wok: unsupported architecture $($env:PROCESSOR_ARCHITECTURE)" }
}
$asset = "wok_windows_$arch.exe"
$base = if ($env:WOK_DOWNLOAD_BASE) {
  $env:WOK_DOWNLOAD_BASE
} elseif ($env:WOK_VERSION) {
  "https://github.com/$repo/releases/download/$($env:WOK_VERSION)"
} else {
  "https://github.com/$repo/releases/latest/download"
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("wok-" + [System.Guid]::NewGuid())
New-Item -ItemType Directory -Force $tmp | Out-Null
try {
  Write-Host "Downloading $asset"
  Invoke-WebRequest -UseBasicParsing -Uri "$base/$asset" -OutFile (Join-Path $tmp $asset)
  Invoke-WebRequest -UseBasicParsing -Uri "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

  $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { ($_ -split '\s+')[1] -eq $asset }
  if (-not $line) { throw "wok: $asset is not listed in checksums.txt" }
  $expected = ($line -split '\s+')[0]
  $actual = (Get-FileHash -Algorithm SHA256 (Join-Path $tmp $asset)).Hash.ToLower()
  if ($expected -ne $actual) { throw "wok: checksum mismatch for $asset" }

  New-Item -ItemType Directory -Force $dir | Out-Null
  Move-Item -Force (Join-Path $tmp $asset) (Join-Path $dir 'wok.exe')
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (-not ($userPath -split ';' | Where-Object { $_ -eq $dir })) {
  [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ';' + $dir).TrimStart(';'), 'User')
  Write-Host "Added $dir to your user PATH. Open a new terminal to use wok."
}
& (Join-Path $dir 'wok.exe') version
