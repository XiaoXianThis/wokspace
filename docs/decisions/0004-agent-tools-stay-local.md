# 0004 Agent tool files stay local

- Status: Accepted
- Date: 2026-10-29

## Context

Every agent tool has its own directory and entry file: `.kiro/`, `.cursor/`, `.claude/`, `CLAUDE.md`, `GEMINI.md`. Committing them ties the repository to particular tools and duplicates the rules in several formats.

## Decision

- `AGENTS.md` is the only committed rule entry point. Most tools read it directly.
- Agent tool directories and entry files are ignored at every depth. Each person sets them up locally, for example a one-line `CLAUDE.md` containing `@AGENTS.md`.
- Configs a team wants to share (MCP servers, hooks, custom agents) are kept as standard versions in `work/agent/configs/<tool>/` and copied locally.
- Tools that write specs into their own directory (Kiro writes `.kiro/specs/`) either link that directory to `work/specs/` (a symlink, or a directory junction on Windows, which needs no admin rights) or copy final specs there. `work/specs/` uses Kiro's spec file names, so both work.

## Alternatives

- **Commit each tool's directory.** Rejected for the reasons above.
- **Symlink entry files into the repository.** Rejected: symlinks need admin rights or developer mode on Windows.

## Consequences

- A new contributor's agent setup is one or two local steps, documented in `docs/README.md`.
- Shared configs can drift from local copies. That is accepted in exchange for a tool-neutral repository.
