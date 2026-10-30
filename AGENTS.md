# AGENTS.md

A notes app: a CLI and an MCP server.

## Layout

- `cmd/notes` — wiring only.
- `internal/notes` — the notes module's contract: the only package other modules use.
- `internal/notes/internal/domain` — pure rules: no I/O, no clock, no logging.
- `internal/notes/internal/<feature>` — one package per operation (`add`, `list`): its `Args`, `Out`, `Run`, and `Op()` binding it to an MCP tool and a CLI command through `internal/platform/op`.
- `internal/notes/internal/infra` — the file store.
- `internal/platform` — shared code that knows no module.

## Rules

- Run `./check.sh` before you say a change is done. It must pass.
- Comments start with `WHY:`; doc comments only on the contract's exported API.
