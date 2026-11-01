# AGENTS.md

notes keeps short notes for people and agents, served as MCP tools (`notes mcp`) and as CLI commands. It is built on the anti-drift kit, github.com/McReaper/go-agent-kit: the kit's checks judge every change, and only the owner can change them.

## Commands

- `go test ./...`: the tests.
- `kit check`: every gate the merge check runs, the same as CI. Run it before you stop.
- `kit mutate`: the tests must kill the mutants of the lines you changed.
- `kit review`: the review sheet the owner reads. Read it before you open a pull request: it is what they will judge.

## Where code goes

- `cmd/notes/`: wiring only: the config, the journal, the modules, the two surfaces. Its tests run the built program: `smoke_test.go` calls every tool and command with its example, `journey_test.go` holds one user journey per feature.
- `internal/notes/`: the notes module's contract, the only package other code imports. Its words are in `internal/notes/GLOSSARY.md`.
- `internal/notes/internal/domain/`: the rules, pure: no I/O, no clock, no log.
- `internal/notes/internal/add/`, `internal/notes/internal/list/`: one operation each, a vertical slice: its `Args`, its `Op` (name, description for agents, effect, example) and its tests. Slices don't import each other.
- `internal/notes/internal/infra/`: files and other I/O, behind the interfaces the slices declare.
- `internal/platform/`: shared by every module: the op binding, the journal, the config. It knows no module.
- `pkg/slug/`: a library other programs may import: its exported API is a contract.

## Rules no check explains

- A test you believe is wrong, or a contract the feature changes: a protected change is expected when it's justified. The `new-feature` skill says which pull request it goes in. Never bend the code to fit a wrong test.
- Code only tests reach: delete it with its tests.
- Never silence a check (`//nolint`, `t.Skip`, a build tag): fix the code, or stop and ask.
- A new operation: load the `new-feature` skill. A new module: the `new-module` skill.
