# wok

- Purpose: the `wok` CLI (`init`, `new`, `dev`, `build`, `test`, `install`, `run`, `ls`, `check`) and the wokspace template it embeds.
- Stack: Go 1.22+, standard library only.
- Depends on: none. The install scripts in `apps/site` download this app's release binaries.

## Tasks

Commands run in `apps/wok/`. Leave a command empty to use what `wok` detects; `wok ls wok` shows the result.

| task | command | windows |
|---|---|---|
| install | go mod download | |
| dev | go run . | |
| test | go vet ./... && go test ./... | |
| build | go build -trimpath -ldflags "-s -w" -o "$WOK_OUT" . | |
| cross | go run ./internal/tools/crossbuild "$WOK_OUT" | |
| snapshot | goreleaser release --snapshot --clean | |

`wok run wok dev -- <args>` runs the CLI from source, for example `wok run wok dev -- check`.

## Layout

```
main.go              entry point; embeds template/
internal/cli/        command parsing and output
internal/project/    finding the wokspace root and its apps; LayoutVersion
internal/tasks/      task tables, detection from project files, $WOK_* expansion
internal/runner/     running commands through sh or cmd, parallel output
internal/scaffold/   init and new
internal/check/      wok check
internal/ui/         colors
template/            the skeleton written by wok init
.goreleaser.yaml     release builds
```

## Conventions

- **No third-party dependencies.** wok must stay a single small binary that builds anywhere.
- **The template is the spec.** `template/AGENTS.md` is the layout spec. The repository root `AGENTS.md` must stay byte-for-byte identical to it; CI checks this. Change both together.
- **Template dotfiles** are stored as `dot-<name>` (for example `template/dot-gitignore`) so they do not apply to this repository. `scaffold.DiskName` maps them back.
- **Skeleton = template directories with a README.md.** `wok check` derives the skeleton from the embedded template, so adding a skeleton directory means adding it to `template/` with a README.
- **Layout version.** When a template change requires existing wokspaces to change, bump `project.LayoutVersion` and the `<!-- wokspace-layout: N -->` marker in both AGENTS.md files.
- **Cross-platform.** Platform code lives in `*_windows.go` and `*_unix.go` / `*_other.go` files. `go vet` must pass with `GOOS=windows` too.
- **Releasing.** Push a tag `vX.Y.Z` at the repository root. The release workflow publishes the binaries to GitHub Releases, the only distribution channel.
- **Release artifacts** are raw binaries named `wok_<os>_<arch>[.exe]` plus `checksums.txt`. The install scripts in `apps/site/public/` depend on these names.
