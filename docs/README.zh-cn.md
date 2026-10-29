# Wokspace

**为 Agent 设计的项目结构，适用于任何技术栈。One wok, any stack.**

[English](README.md) · [规范（AGENTS.md）](../AGENTS.md) · [决策记录](decisions/README.md)

npm、Cargo、Go、Gradle 这些工具都把某一种语言的工程放在仓库顶层。这适合只用一种技术栈的人。Agent 会同时在多种技术栈之间工作，没有明确规则时会到处建文件和目录，时间一长，仓库就乱了。

**wokspace** 的顶层是工作区，而不是某个工程：

```
AGENTS.md        所有 Agent 首先读取的规则
apps/            独立的子项目，各自使用自己语言的原生结构
build/           最终构建产物（不提交）
docs/            长期有效的事实：架构、决策、指南、接口契约
work/            其余一切：spec、脚本、测试、素材、基础设施、临时区
```

- **根目录冻结。** 只提交 3 个文件和 4 个目录。白名单式的 `.gitignore` 把其他东西都挡在 git 之外，包括各家 Agent 的专属目录（`.cursor/`、`.kiro/`、`.claude/` 等）。
- **骨架冻结。** 所有目录从第一天起就存在，每个目录都有一份写明规则的 README，Agent 在写文件之前就知道该放在哪里。
- **多技术栈并存。** `apps/` 里的每个 app 都是普通的 Go、Rust、Node、Python、.NET 或 Java 工程。app 之间不共享源码，只通过构建产物或 `docs/api/` 里的契约互相依赖。
- **不依赖工具也能运转。** 规则是纯 Markdown。`wok` 命令行只是加速器：创建 wokspace、在根目录运行 app 的任务、检查规则。
- **支持 Windows、macOS、Linux。** 命名规则、换行符和命令都按三个平台都能用的标准来定。

## 安装

macOS / Linux：

```sh
curl -fsSL https://raw.githubusercontent.com/XiaoXianThis/wokspace/main/apps/wok/install/install.sh | sh
```

Windows（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/XiaoXianThis/wokspace/main/apps/wok/install/install.ps1 | iex
```

有 Node.js 时可以免安装：`npx wokspace <命令>`。有 Go 时：`go install github.com/XiaoXianThis/wokspace/apps/wok@latest`。

各平台的预编译文件见 [Releases 页面](https://github.com/XiaoXianThis/wokspace/releases)。

## 快速开始

```sh
wok init my-project          # 生成完整骨架，并执行 git init
cd my-project
wok new web                  # 生成 apps/web/AGENTS.md，并登记到 apps/README.md
cd apps/web && npm create vite@latest . && cd ../..
wok dev web                  # 在 apps/web/ 中执行 npm run dev
```

也可以直接告诉 Agent："用 Vite 建一个 web 应用。" `AGENTS.md` 的规则会让它执行 `wok new web`，在 `apps/web/` 里初始化工程，填写任务表，并在结束前执行 `wok check`。

## 命令

```
wok init [dir]                 创建 wokspace（--existing 向非空目录补充缺失的文件）
wok new <app>                  创建 apps/<app>/ 及其 AGENTS.md
wok dev [app...]               启动 app；多个 app 并行运行，输出带标签
wok build [app...] [--all]     构建 app，产物输出到 build/<app>/
wok test [app...] [--all]      测试 app
wok install [app...] [--all]   安装 app 的依赖
wok run <app> <task> [-- args] 运行任意任务
wok ls [app]                   列出 app 以及每个任务背后的命令
wok check                      按 AGENTS.md 检查目录结构
```

在 wokspace 内的任何目录都能使用 wok。在 `apps/<app>/` 里时，可以省略 app 名。

## 任务：约定优先，必要时再配置

wok 没有配置文件。每个 app 的 `AGENTS.md` 里有一张任务表，人、Agent 和 wok 都读这张表：

```markdown
| task | command | windows |
|---|---|---|
| install | | |
| dev | | |
| test | cargo test --all-features | |
| build | cargo build --release && cp target/release/api "$WOK_OUT" | cargo build --release && copy target\release\api.exe "$WOK_OUT" |
```

命令留空表示"使用 wok 自动识别的命令"。`wok ls` 会显示最终结果，以及每条命令的来源。自动识别规则见[英文版](README.md#tasks-convention-first-config-if-needed)中的表格：支持 Node（npm、pnpm、yarn、bun）、Deno、Cargo、Go、.NET、Gradle、Maven，以及使用 uv 或 Poetry 的 Python 项目。`package.json` 的每个 script 和 `deno.json` 的每个 task 也会自动成为任务，所以不写表也能直接 `wok run web lint`。

命令在 app 目录中执行：macOS 和 Linux 通过 `sh`，Windows 通过 `cmd`。执行前，wok 会在所有平台上替换 `$WOK_ROOT`、`$WOK_APP` 和 `$WOK_OUT`（该 app 的 `build/<app>/` 目录），同时把它们设为环境变量。

## Agent

大多数 Agent 工具会直接读取 `AGENTS.md`，例如 Kiro、Codex、Cursor、GitHub Copilot。Claude Code 需要在本地建一个 `CLAUDE.md`，内容是 `@AGENTS.md`。Agent 工具的目录都不进 git；团队需要共享的配置放在 `work/agent/configs/`。详见新建 wokspace 中 `docs/README.md` 的 Agent setup 一节。

## 本仓库

本仓库本身就是一个 wokspace：

| app | 内容 |
|---|---|
| [`apps/wok`](../apps/wok/AGENTS.md) | Go 写的命令行工具，模板在 `apps/wok/template/` |
| [`apps/npm`](../apps/npm/AGENTS.md) | `wokspace` npm 包，负责下载命令行工具 |

根目录的 `AGENTS.md` 就是规范本身，和 `wok init` 生成的完全一致。

## 许可证

[MIT](../LICENSE)
