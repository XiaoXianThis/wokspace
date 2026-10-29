# 0001 A workspace at the top, not a project

- Status: Accepted
- Date: 2026-10-29

## Context

Language toolchains (npm, Cargo, Go, Gradle) expect their project at the repository root. Agents work across languages, and without firm rules they create files wherever is convenient. Over time the root fills with configs, notes, scripts, and debug output from several stacks.

## Decision

The root is a workspace with three committed files (`AGENTS.md`, `.gitignore`, `.gitattributes`) and four directories:

- `apps/` for independent projects, each in its own native layout;
- `build/` for final output (only its README is committed);
- `docs/` for long-lived facts;
- `work/` for everything else.

The root is frozen. A whitelist `.gitignore` (`/*` followed by `!/entry` lines) keeps any other root entry out of git, including every agent tool's private directory. Optional entries that platforms require at the root (`.github/`, `LICENSE`, `.gitmodules`, `.gitlab-ci.yml`) are added to the whitelist only after the user approves.

The project home page lives in `docs/README.md`, which GitHub shows on the repository page, so the root needs no README.

## Alternatives

- **A root README and `.editorconfig`.** Rejected: GitHub finds `docs/README.md`, and formatting belongs to each app's own tools.
- **A deny list in `.gitignore`.** Rejected: it can never list every file an agent might create.
- **A whitelist inside `work/` too.** Rejected: files in a wrong place in `work/` are usually real work. Ignoring them silently would lose it, so `wok check` reports them instead.

## Consequences

- Stray root files never reach the repository, but they can still exist locally. `wok check` lists them.
- Platforms that only read a root README (some git hosts other than GitHub) need the README moved back to the root.
