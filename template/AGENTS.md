# AGENTS.md

notes keeps short notes for people and agents, served as MCP tools (`notes mcp`) and as CLI commands. It is built on Rumblestrip, github.com/McReaper/rumblestrip-go: Rumblestrip's checks judge every change, and only the owner can change them.

## Where code goes

- `features/`: the spec, the one thing the owner reads: Gherkin scenarios of what the app does, in `notes.feature`, and the steps that run them on the built program, in `notes_test.go`. A scenario tagged `@wip` is approved, not built yet.
- `cmd/notes/`: wiring only: the config, the journal, the modules, the two surfaces. Its `smoke_test.go` builds the program and calls every tool and command with its example.
- `internal/notes/`: the notes module's contract, the only package other code imports. Its words are in `internal/notes/GLOSSARY.md`.
- `internal/notes/internal/domain/`: the rules, pure: no I/O, no clock, no log.
- `internal/notes/internal/add/`, `internal/notes/internal/list/`: one operation each, a vertical slice: its `Args`, its `Op` (name, description for agents, effect, example) and its tests. Slices don't import each other.
- `internal/notes/internal/infra/`: files and other I/O, behind the interfaces the slices declare.
- `internal/platform/`: shared by every module: the op binding, the journal, the config. It knows no module.
- `pkg/slug/`: a library other programs may import: its exported API is a contract.

Rumblestrip's commands and the rules no check explains come with its Claude Code plugin, `rumblestrip`, which the repository's Claude Code settings, written by `rumblestrip init`, turn on.
