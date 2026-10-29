# 0002 A frozen skeleton with a README in every directory

- Status: Accepted
- Date: 2026-10-29

## Context

The first design let directories "grow" on demand from a list in `AGENTS.md`. Agents had to read the list, decide whether a directory was allowed, and create it. Directories that existed nowhere were easy to miss.

## Decision

`wok init` creates every skeleton directory up front, and the skeleton never changes without user approval. Git does not track empty directories, so each skeleton directory has a `README.md`. The README states the directory's purpose and rules, which puts the rules where the agent is about to write.

Content inside skeleton directories still grows freely, within a nesting limit: two levels of subdirectories below a leaf skeleton directory, none inside `docs/` categories.

The skeleton is defined by the template itself: every template directory with a README is a skeleton directory. `wok check` derives the skeleton from the template it embeds.

## Alternatives

- **Grow on demand.** Rejected for the reason above.
- **`.gitkeep` placeholders.** Rejected: an empty marker carries no rules.

## Consequences

- A new wokspace has 29 files before any app exists.
- Adding a skeleton directory is a change to the spec: a new README in the template and a layout version bump if existing wokspaces must follow.
