# AGENTS.md

notes keeps short notes for people and agents, served as MCP tools (`notes mcp`) and as CLI commands. It is built on the anti-drift kit, github.com/McReaper/go-agent-kit: the kit's checks judge every change, and only the owner can change them.

## Commands

- `go test ./...`: the tests.
- `kit check`: every gate the merge check runs, the same as CI. Run it before you stop.
- `kit mutate`: the tests must kill the mutants of the lines you changed.
- `kit review`: the review sheet the owner reads. Read it before you open a pull request: it is what they will judge.

## Where code goes

- `cmd/notes/`: wiring only: the config, the journal, the modules, the two surfaces.
- `internal/notes/`: the notes module's contract, the only package other code imports. Its words are in `internal/notes/GLOSSARY.md`.
- `internal/notes/internal/domain/`: the rules, pure: no I/O, no clock, no log.
- `internal/notes/internal/add/`, `internal/notes/internal/list/`: one operation each, a vertical slice: its `Args`, its `Op` (name, description for agents, effect) and its tests. Slices don't import each other.
- `internal/notes/internal/infra/`: files and other I/O, behind the interfaces the slices declare.
- `internal/platform/`: shared by every module: the op binding, the journal, the config. It knows no module.
- `pkg/slug/`: a library other programs may import: its exported API is a contract.

## Rules no check explains

- A test you believe is wrong: fix the test, and say in your report what changed in it and why. The owner approves every changed test before merge; bending the code to fit a wrong test is the one thing that costs more.
- Code only tests reach: delete it with its tests.
- Never silence a check (`//nolint`, `t.Skip`, a build tag): fix the code, or stop and ask.
- A new operation: load the `new-feature` skill. A new module: the `new-module` skill.
