# npm

- Purpose: the `wokspace` npm package, so that `npx wokspace init` works without installing anything. It contains no Go code: `bin/wok.js` downloads the matching wok release binary on first use, verifies its checksum, caches it, and runs it.
- Stack: Node.js 18+, no dependencies.
- Depends on: wok, through its GitHub release assets `wok_<os>_<arch>[.exe]` and `checksums.txt` for the tag `v<package version>`.

## Tasks

Commands run in `apps/npm/`. Leave a command empty to use what `wok` detects; `wok ls npm` shows the result.

| task | command | windows |
|---|---|---|
| install | | |
| dev | | |
| test | npm test | |
| build | npm pack --pack-destination "$WOK_OUT" | |

To try the wrapper against a local wok build: `wok build wok`, then set `WOK_BINARY` to `build/wok/wok` (or `wok.exe`) and run `node bin/wok.js version`.

## Conventions

- `version` stays `0.0.0` in the repository. The release workflow sets it from the git tag before publishing, so the package version always equals the wok release it downloads.
- No dependencies and no install scripts. The download happens at first run, which also works where `npm install` scripts are disabled.
