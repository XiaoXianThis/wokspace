# {{name}}

One sentence about the project.

> This file is the repository home page. GitHub shows `docs/README.md` automatically, so the root has no README.

## Layout

This repository is a [wokspace](https://wokspace.wxian.cc).

```
AGENTS.md        rules for agents; the entry point for every agent
apps/            independent projects, each in its own language's layout
build/           final build output (not committed)
docs/            long-lived documentation (this directory)
work/            specs, scripts, tests, assets, infrastructure, scratch space
```

## Commands

```
wok ls                 list apps and their tasks
wok dev <app>          start an app
wok build --all        build every app into build/
wok check              check the layout
```

## Documentation

(List documents here as they are added.)

## Agent setup

Files that belong to one agent tool are ignored by git, so each person sets them up locally. Shared standard versions live in `work/agent/configs/`. Run the commands below at the repository root.

- **Tools that read AGENTS.md** (Kiro, Codex, Cursor, GitHub Copilot, ...): nothing to do.
- **Claude Code**: create `CLAUDE.md` containing the single line `@AGENTS.md`.
- **Gemini CLI**: set its context file name to `AGENTS.md`.
- **Kiro specs (optional)**: Kiro writes specs to `.kiro/specs/`. Link it to `work/specs/` before Kiro creates its first spec, and specs become part of the repository.

  macOS / Linux:

  ```sh
  mkdir -p .kiro
  ln -s ../work/specs .kiro/specs
  ```

  Windows PowerShell (a directory junction needs no admin rights):

  ```powershell
  New-Item -ItemType Directory -Force .kiro | Out-Null
  New-Item -ItemType Junction -Path .kiro\specs -Target (Resolve-Path work\specs)
  ```

  Without the link, agents copy final specs to `work/specs/` as `AGENTS.md` requires.
