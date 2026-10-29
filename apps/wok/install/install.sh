#!/bin/sh
# Install wok on macOS or Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/XiaoXianThis/wokspace/main/apps/wok/install/install.sh | sh
#
# Environment:
#   WOK_VERSION      release tag to install, such as v0.1.0 (default: latest)
#   WOK_INSTALL_DIR  where to put the binary (default: ~/.local/bin)
#   WOK_DOWNLOAD_BASE  download from this URL instead of GitHub releases (a mirror)
set -eu

repo="XiaoXianThis/wokspace"
dir="${WOK_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "wok: unsupported OS $(uname -s); on Windows use install.ps1" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "wok: unsupported architecture $(uname -m)" >&2; exit 1 ;;
esac

asset="wok_${os}_${arch}"
if [ -n "${WOK_DOWNLOAD_BASE:-}" ]; then
  base="$WOK_DOWNLOAD_BASE"
elif [ -n "${WOK_VERSION:-}" ]; then
  base="https://github.com/$repo/releases/download/$WOK_VERSION"
else
  base="https://github.com/$repo/releases/latest/download"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $asset"
curl -fsSL -o "$tmp/$asset" "$base/$asset"
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt"

expected="$(awk -v f="$asset" '$2 == f { print $1 }' "$tmp/checksums.txt")"
if [ -z "$expected" ]; then
  echo "wok: $asset is not listed in checksums.txt" >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp/$asset" | awk '{ print $1 }')"
else
  actual="$(shasum -a 256 "$tmp/$asset" | awk '{ print $1 }')"
fi
if [ "$expected" != "$actual" ]; then
  echo "wok: checksum mismatch for $asset" >&2
  exit 1
fi

mkdir -p "$dir"
chmod +x "$tmp/$asset"
mv "$tmp/$asset" "$dir/wok"
echo "Installed $("$dir/wok" version) to $dir/wok"

case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "Add $dir to your PATH, for example: export PATH=\"$dir:\$PATH\"" ;;
esac
