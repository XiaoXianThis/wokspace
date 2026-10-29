# apps

Each subdirectory is an independent app that follows its own language's conventions.

## Rules

- Ask the user before creating an app. Use a kebab-case name.
- Create it with `wok new <name>`. Without wok, create `apps/<name>/AGENTS.md` from the template below and add the app to the list.
- Before working in an app, read its `AGENTS.md`.
- Apps never reference each other's source. Allowed dependencies are listed in the root `AGENTS.md`, section 4.

## Apps

| app | purpose | stack | depends on |
|---|---|---|---|
| wok | the `wok` CLI and the embedded wokspace template | Go | none |
| site | the website at wokspace.wxian.cc and the install scripts | static HTML | wok (release binaries) |

## App AGENTS.md template

`wok new` copies the block below, replacing `<name>`. Edit it to change what new apps start with.

```markdown
# <name>

- Purpose:
- Stack:
- Depends on: none

## Tasks

Commands run in `apps/<name>/`. Leave a command empty to use what `wok` detects; `wok ls <name>` shows the result.

| task | command | windows |
|---|---|---|
| install | | |
| dev | | |
| test | | |
| build | | |

## Conventions
```
