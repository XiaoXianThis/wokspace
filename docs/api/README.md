# api

Contracts between apps. This is the only source-level agreement apps share.

## Rules

- A contract may be prose or a machine-readable schema (`openapi.yaml`, `*.proto`, JSON Schema, ...). When a schema exists, the schema wins.
- Code generated from a schema lives in the app that uses it, not here.
- When a contract changes, update the provider and every consumer in the same task. Record breaking changes in `docs/decisions/`.
- Kebab-case file names, starting with the provider app's name, such as `server-http.yaml`. No subdirectories.

## Contracts

| contract | provider | consumers |
|---|---|---|
