# decisions

Architecture decision records: what was decided, why, and which options were rejected.

## Rules

- File names are `NNNN-slug.md`, numbered from `0001`, such as `0001-use-postgres.md`.
- Start from a copy of `0000-template.md`.
- Never delete a record or rewrite its decision. When a decision is reversed, add a new record and set the old one's status to "Superseded by NNNN".
- No subdirectories.

## Records

| record | status |
|---|---|
| [0001 A workspace at the top, not a project](0001-workspace-at-the-top.md) | Accepted |
| [0002 A frozen skeleton with a README in every directory](0002-frozen-skeleton.md) | Accepted |
| [0003 An optional CLI in Go with no config file](0003-optional-cli.md) | Accepted |
| [0004 Agent tool files stay local](0004-agent-tools-stay-local.md) | Accepted |
