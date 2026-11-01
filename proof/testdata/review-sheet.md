<!-- kit review -->
## Review sheet

### Needs your approval

_the merge check waits for a code owner's approval of the last commit_

- `internal/notes/internal/add/add_test.go`: T2: a test changed: the owner approves it before merge; say in your report what changed in it and why
- `internal/notes/internal/domain/note_test.go`: T2: a test changed: the owner approves it before merge; say in your report what changed in it and why
- `internal/notes/testdata/tools.golden.json`: T2: a test changed: the owner approves it before merge; say in your report what changed in it and why

### Test scenarios

Added:

- Normalize keeps the first order once (`internal/tags/tags_test.go`)
- Normalize keeps the first order once: case folded (`internal/tags/tags_test.go`)

Removed:

- Parse title: refuses an empty title (`internal/notes/internal/domain/note_test.go`)

Changed:

- Add reports the store failure (`internal/notes/internal/add/add_test.go`)
- Parse title (`internal/notes/internal/domain/note_test.go`)

### Contract

_golden snapshots of what the app's users see_

- `internal/notes/testdata/tools.golden.json`: changed

### Library API

_apidiff against the base_

- `pkg/slug`: compatible: Join: added

### New modules

- `internal/tags`

### Dependencies

_go.mod lines added_

- `go.mod:14`: golang.org/x/sync v0.22.0

### Shared code

_what other modules or the app build on_

- `internal/tags/tags.go`: the tags module's contract, added
- `pkg/slug/slug.go`: library, changed

### Rules

_CI, the agent's settings and instructions, decision records_

- `AGENTS.md`: the agent's instructions, changed

### Decisions

_`// WHY:` comments added_

- `internal/tags/tags.go:8`: tags come from people typing; "Go" and "go" are one tag, and the order they gave is the order they read back.
