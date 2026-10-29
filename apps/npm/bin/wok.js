#!/usr/bin/env node
// Runs the wok binary for this platform. On first use, downloads the release
// that matches this package's version from GitHub, verifies its SHA-256
// checksum, and caches it.
//
// Environment:
//   WOK_BINARY     use this wok binary instead of downloading one
//   WOK_CACHE_DIR  cache directory (default: ~/.cache/wokspace)
//   WOK_DOWNLOAD_BASE  download from this URL instead of GitHub releases (a mirror)
'use strict';

const crypto = require('node:crypto');
const fs = require('node:fs');
const http = require('node:http');
const https = require('node:https');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

const pkg = require('../package.json');
const repo = 'XiaoXianThis/wokspace';

function fail(message) {
  console.error(`wok: ${message}`);
  process.exit(1);
}

function assetName() {
  const goos = { darwin: 'darwin', linux: 'linux', win32: 'windows' }[process.platform];
  const goarch = { x64: 'amd64', arm64: 'arm64' }[process.arch];
  if (!goos || !goarch) fail(`unsupported platform ${process.platform}/${process.arch}`);
  return `wok_${goos}_${goarch}${goos === 'windows' ? '.exe' : ''}`;
}

function get(url, redirects = 5) {
  const client = url.startsWith('http:') ? http : https;
  return new Promise((resolve, reject) => {
    client
      .get(url, { headers: { 'User-Agent': `wokspace-npm/${pkg.version}` } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location && redirects > 0) {
          res.resume();
          resolve(get(new URL(res.headers.location, url).toString(), redirects - 1));
          return;
        }
        if (res.statusCode !== 200) {
          res.resume();
          reject(new Error(`GET ${url}: HTTP ${res.statusCode}`));
          return;
        }
        const chunks = [];
        res.on('data', (c) => chunks.push(c));
        res.on('end', () => resolve(Buffer.concat(chunks)));
        res.on('error', reject);
      })
      .on('error', reject);
  });
}

async function download(asset, dest) {
  const base = (
    process.env.WOK_DOWNLOAD_BASE || `https://github.com/${repo}/releases/download/v${pkg.version}`
  ).replace(/\/+$/, '');
  console.error(`wok: downloading ${asset} v${pkg.version}`);
  const [binary, sums] = await Promise.all([get(`${base}/${asset}`), get(`${base}/checksums.txt`)]);
  const line = sums
    .toString('utf8')
    .split('\n')
    .map((l) => l.trim().split(/\s+/))
    .find((parts) => parts[1] === asset);
  if (!line) throw new Error(`${asset} is not listed in checksums.txt`);
  const actual = crypto.createHash('sha256').update(binary).digest('hex');
  if (actual !== line[0]) throw new Error(`checksum mismatch for ${asset}`);

  fs.mkdirSync(path.dirname(dest), { recursive: true });
  const tmp = `${dest}.${process.pid}.tmp`;
  fs.writeFileSync(tmp, binary, { mode: 0o755 });
  fs.renameSync(tmp, dest);
}

async function main() {
  let bin = process.env.WOK_BINARY;
  if (!bin) {
    const asset = assetName();
    const cache = process.env.WOK_CACHE_DIR || path.join(os.homedir(), '.cache', 'wokspace');
    bin = path.join(cache, pkg.version, asset);
    if (!fs.existsSync(bin)) {
      try {
        await download(asset, bin);
      } catch (err) {
        fail(`${err.message}\nInstall wok another way: https://github.com/${repo}#install`);
      }
    }
  }
  const res = spawnSync(bin, process.argv.slice(2), { stdio: 'inherit' });
  if (res.error) fail(res.error.message);
  if (res.signal) process.kill(process.pid, res.signal);
  process.exit(res.status === null ? 1 : res.status);
}

main();
