# site

- Purpose: the project website, served by GitHub Pages at wokspace.wxian.cc. It also serves the install scripts `install.sh` and `install.ps1`.
- Stack: static HTML, CSS, and plain JavaScript. No build tool, no dependencies.
- Depends on: wok, through its GitHub release assets `wok_<os>_<arch>[.exe]` and `checksums.txt`, which the install scripts download.

## Tasks

Commands run in `apps/site/`. Leave a command empty to use what `wok` detects; `wok ls site` shows the result.

| task | command | windows |
|---|---|---|
| install | | |
| dev | python3 -m http.server 8080 --directory public | python -m http.server 8080 --directory public |
| test | sh -n public/install.sh | powershell -NoProfile -Command "$null = [scriptblock]::Create((Get-Content -Raw public\install.ps1))" |
| build | cp -R public/. "$WOK_OUT" | xcopy /E /I /Y /Q public "$WOK_OUT" |

`dev` serves the site at http://localhost:8080 and needs Python 3. `build` copies `public/` to `build/site/`, which the Pages workflow deploys.

## Layout

```
public/index.html      Chinese home page (the default language)
public/en/index.html   English page
public/style.css       shared styles, light and dark
public/main.js         install command URLs, tabs, copy buttons
public/install.sh      installer for macOS and Linux
public/install.ps1     installer for Windows
```

## Conventions

- Chinese is the default language. Keep `index.html` and `en/index.html` in sync: same sections, same commands.
- The install commands on the page point at the site the page is served from (`main.js`), so they work on the Pages URL and on the custom domain. The HTML keeps a raw GitHub URL as the fallback.
- The custom domain is configured in the repository's Pages settings, not with a `CNAME` file.
- The install scripts depend on the release asset names produced by `apps/wok/.goreleaser.yaml`. Change them together.
