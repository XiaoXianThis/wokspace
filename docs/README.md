# Wokspace

**An agent-first repository layout for any stack. One wok, any stack.**

[中文](README.zh-cn.md) · [Website](https://wokspace.wxian.cc) · [Spec (AGENTS.md)](../AGENTS.md) · [Decisions](decisions/README.md)

Package managers like npm, Cargo, Go, and Gradle put one language's project at the top of the repository. That suits a person working in one stack. Agents work across stacks, and without firm rules they scatter files and directories until the repository is a mess.

A **wokspace** puts a workspace at the top instead of a project:

```
AGENTS.md        the rules every agent reads first
apps/            independent projects, each in its own language's native layout
build/           final build output (not committed)
docs/            long-lived facts: architecture, decisions, guides, API contracts
work/            everything else: specs, scripts, tests, assets, infra, scratch space
```

- **A frozen root.** Three committed files and four directories. A whitelist `.gitignore` keeps everything else out of git, including every agent tool's private directory (`.cursor/`, `.kiro/`, `.claude/`, ...).
- **A frozen skeleton.** Every directory exists from day one, and each has a README with its rules, so an agent always knows where a file goes before it writes one.
- **Any stack, side by side.** Each app in `apps/` is a normal Go, Rust, Node, Python, .NET, or Java project. Apps never share source; they depend on each other's build output or on contracts in `docs/api/`.
- **Works without a tool.** The rules are plain Markdown. The `wok` CLI only speeds things up: it creates wokspaces, runs app tasks from the root, and checks the rules.
- **Windows, macOS, and Linux.** Naming rules, line endings, and commands are chosen to work everywhere.

## Install

macOS / Linux:

```sh
curl -fsSL https://wokspace.wxian.cc/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://wokspace.wxian.cc/install.ps1 | iex
```

The install scripts verify SHA-256 checksums. Prebuilt binaries for every platform are also on the [releases page](https://github.com/XiaoXianThis/wokspace/releases).

## Quick start

```sh
wok init my-project          # the full skeleton, plus git init
cd my-project
wok new web                  # apps/web/AGENTS.md, listed in apps/README.md
cd apps/web && npm create vite@latest . && cd ../..
wok dev web                  # runs `npm run dev` in apps/web/
```

Or ask your agent: "Create a web app with Vite." The rules in `AGENTS.md` tell it to run `wok new web`, initialize the app inside `apps/web/`, fill in the task table, and run `wok check` before it finishes.

## Commands

```
wok init [dir]                 create a wokspace (--existing adds missing files to a non-empty directory)
wok new <app>                  create apps/<app>/ with its AGENTS.md
wok dev [app...]               start apps; several run in parallel with labeled output
wok build [app...] [--all]     build apps into build/<app>/
wok test [app...] [--all]      test apps
wok install [app...] [--all]   install app dependencies
wok run <app> <task> [-- args] run any task
wok ls [app]                   list apps and the command behind each task
wok check                      check the layout against AGENTS.md
```

wok works from any directory inside a wokspace. Inside `apps/<app>/`, the app name can be left out.

## Tasks: convention first, config if needed

There is no wok config file. Each app's `AGENTS.md` has a task table that humans, agents, and wok all read:

```markdown
| task | command | windows |
|---|---|---|
| install | | |
| dev | | |
| test | cargo test --all-features | |
| build | cargo build --release && cp target/release/api "$WOK_OUT" | cargo build --release && copy target\release\api.exe "$WOK_OUT" |
```

An empty command means "use what wok detects". `wok ls` shows the result and where each command came from.

| project file | install | dev | test | build |
|---|---|---|---|---|
| `package.json` (npm, pnpm, yarn, or bun, chosen from the lock file) | `<pm> install` | `<pm> run dev` | `<pm> run test` | `<pm> run build` |
| `deno.json` | `deno install` | `deno task dev` | `deno task test` | `deno task build` |
| `Cargo.toml` | `cargo fetch` | `cargo run` | `cargo test` | `cargo build --release` |
| `go.mod` | `go mod download` | `go run .` | `go test ./...` | `go build -o "$WOK_OUT" .` |
| `*.sln`, `*.csproj`, `*.fsproj` | `dotnet restore` | `dotnet run` | `dotnet test` | `dotnet publish -c Release -o "$WOK_OUT"` |
| `gradlew`, `build.gradle(.kts)` | | | `gradlew test` | `gradlew build` |
| `pom.xml` | `mvn dependency:resolve` | | `mvn test` | `mvn package` |
| `pyproject.toml` + `uv.lock` | `uv sync` | | `uv run pytest` | `uv build --out-dir "$WOK_OUT"` |
| `pyproject.toml` + `poetry.lock` | `poetry install` | | `poetry run pytest` | `poetry build` |

Every `package.json` script and `deno.json` task becomes a task too, so `wok run web lint` works without any table row.

Commands run in the app directory, through `sh` on macOS and Linux and `cmd` on Windows. Before a command runs, wok replaces `$WOK_ROOT`, `$WOK_APP`, and `$WOK_OUT` (the app's `build/<app>/` directory) on every platform, and also sets them as environment variables.

## Agents

Most agent tools read `AGENTS.md` directly: Kiro, Codex, Cursor, GitHub Copilot, and others. For Claude Code, create a local `CLAUDE.md` containing `@AGENTS.md`. Agent tool directories stay out of git; configs a team wants to share go to `work/agent/configs/`. See the agent setup section in a new wokspace's `docs/README.md`.

## This repository

This repository is itself a wokspace:

| app | what it is |
|---|---|
| [`apps/wok`](../apps/wok/AGENTS.md) | the Go CLI, with the template in `apps/wok/template/` |
| [`apps/site`](../apps/site/AGENTS.md) | the website at [wokspace.wxian.cc](https://wokspace.wxian.cc) and the install scripts |

The root `AGENTS.md` is the spec, and it is identical to the one `wok init` writes.

## License

[MIT](../LICENSE)
