# AGENTS.md

notes keeps short notes for people and agents, served as MCP tools (`notes mcp`) and as CLI commands. It is built on the anti-drift kit, github.com/McReaper/go-agent-kit: the kit's checks judge every change, and only the owner can change them.

## Where code goes

- `features/`: the spec, the one thing the owner reads: Gherkin scenarios of what the app does, in `notes.feature`, and the steps that run them on the built program, in `notes_test.go`. A scenario tagged `@wip` is approved, not built yet.
- `cmd/notes/`: wiring only: the config, the journal, the modules, the two surfaces. Its `smoke_test.go` builds the program and calls every tool and command with its example.
- `internal/notes/`: the notes module's contract, the only package other code imports. Its words are in `internal/notes/GLOSSARY.md`.
- `internal/notes/internal/domain/`: the rules, pure: no I/O, no clock, no log.
- `internal/notes/internal/add/`, `internal/notes/internal/list/`: one operation each, a vertical slice: its `Args`, its `Op` (name, description for agents, effect, example) and its tests. Slices don't import each other.
- `internal/notes/internal/infra/`: files and other I/O, behind the interfaces the slices declare.
- `internal/platform/`: shared by every module: the op binding, the journal, the config. It knows no module.
- `pkg/slug/`: a library other programs may import: its exported API is a contract.

## Commands

- `go test ./...`: the tests.
- `kit check`: every gate the merge check runs, the same as CI. Run it before you stop.
- `kit mutate`: the tests must kill the mutants of the lines you changed.
- `kit spec`: each built scenario passes alone, and fails once a value its `Then` steps expect is altered (`kit check` runs it).
- `kit review`: the review sheet the owner reads: the spec's changes and what else is theirs to approve. Read it before you open a pull request.

## Rules no check explains

- The spec comes first, in its own pull request the owner approves; the feature is built on top. A scenario you believe is wrong: never bend the code or the steps to it; the `new-feature` skill says how to change it.
- Code only tests reach: delete it with its tests.
- Never silence a check (`//nolint`, `t.Skip`, a build tag): fix the code, or stop and ask.
- A new operation: load the `new-feature` skill. A new module: the `new-module` skill.
- Tags are lower-case: the tags module normalizes them, never the caller.
