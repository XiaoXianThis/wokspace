# 0003 An optional CLI in Go with no config file

- Status: Accepted
- Date: 2026-10-29

## Context

People want `cargo init`-style creation and `npm run`-style commands from the root. The layout must also keep working without any tool, since agents and humans can follow Markdown rules directly.

## Decision

`wok` is an accelerator, never a requirement. Everything it does can be derived from the layout conventions, and it keeps no state of its own:

- `init` and `new` write files any person could write by hand.
- Task commands come from the task table in each app's `AGENTS.md`, which agents fill in when they create the app. Empty cells fall back to detection from project files (`package.json`, `Cargo.toml`, `go.mod`, ...).
- `$WOK_ROOT`, `$WOK_APP`, and `$WOK_OUT` are expanded by wok on every platform, so one command works under both `sh` and `cmd`.
- `check` automates the self-check that `AGENTS.md` describes.

There is no `wok.toml`. The app template for `wok new` is read from the project's own `apps/README.md`, so a team changes it by editing Markdown.

wok is written in Go with only the standard library: one static binary per platform, simple cross-compilation, and good child-process control on Windows.

## Alternatives

- **A root `package.json` with `npm run dev:<app>`.** Rejected: it breaks the frozen root and makes every wokspace depend on Node.
- **`make` or `just`.** Rejected: neither is installed by default on every platform, and both put a file at the root.
- **Rust.** A good fit too. Go was chosen for simpler process and signal handling across platforms.

## Consequences

- Without wok, a person runs the same command inside `apps/<app>/` and substitutes `$WOK_*` by hand.
- Detection covers common setups only. Anything else goes in the task table.
