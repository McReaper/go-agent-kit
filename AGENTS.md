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

- Run `./check.sh` before you say a change is done. It runs exactly what CI runs.
- Comments start with `WHY:`; doc comments only on the contract's exported API.

## Protected changes

CI refuses these until the owner approves them; `kit guard` reports them as `protected-…`:

- changing, renaming or deleting an existing test file, or anything under `testdata/` (new test files are free);
- a new `TestMain` or `init` in a test;
- the `go` or `toolchain` line of go.mod;
- `.kit/` and `check.sh`, which belong to the kit.

If check.sh fails only on protected findings, the change needs the owner's approval: say which files and why in your report.
