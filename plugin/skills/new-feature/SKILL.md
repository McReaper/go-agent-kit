---
name: new-feature
description: How to add or change a feature of this app, spec first. The spec is the Gherkin scenarios in features/, the one thing the owner reads and approves; an operation is one MCP tool and its CLI command, from one arguments struct. Use when asked to add a feature, a tool, a command or an operation, or to change what an existing one does or accepts.
---

# A feature, spec first

The owner reads the spec and nothing else: the scenarios in `features/*.feature`. They approve what the app does there. The code, its unit tests and the steps that run the scenarios are judged by the machines: `rumblestrip check`, `rumblestrip mutate`.

## 1. Write the spec, open its pull request

On a branch from the default branch, `spec/<feature>`, add or change scenarios in `features/<module>.feature`:

- one scenario per behaviour a user would notice, named as that behaviour;
- in the user's words: what they do and what they see, not flags or JSON;
- every `Then` says what it expects with a value, a quoted text or a number: `Then it is refused with "the limit is 80"`. Rumblestrip alters each value and checks the scenario then fails; a `Then` with no value is refused;
- a new scenario is tagged `@wip`: approved, not built yet. The spec's test leaves it out.

Push the branch, then open the pull request with the spec alone: `rumblestrip pr create --title "spec(<module>): <the behaviour, in the user's words>" --body-file <file>`. It refuses a title the check would refuse. The description says in one line what the feature is for. If the request leaves something open, a limit or what is refused, write the scenario you would choose and say so there: the owner decides. Write `@wip` in backticks there, as any `@name`: GitHub notifies the account a bare one names. Once they approve it, the pull request merges by itself.

## 2. Build it on top

Branch from the spec branch, `feat/<feature>`, push it, and open its pull request on the spec branch while the spec waits: `rumblestrip pr create --base spec/<feature> --title "feat(<module>): …" --body-file <file>`. It stacks it on the spec's pull request in the same go: the check fails a pull request on another branch that is in no stack. One opened otherwise gets stacked with `rumblestrip pr stack`, on its branch. Once the spec merges, GitHub rebases the build onto the default branch, and its check runs again.

- Remove the `@wip` tag of each scenario you build: that is free. Any other change to the spec is a spec change the owner approves again.
- Write the steps in `features/*_test.go`, like the ones there: one method per step, bound by a pattern that takes any value (`"([^"]*)"`, `(\d+)`), driving the built program from the command line or over MCP.
- Write the slice, its unit tests first (section 3). Each test fails on something it checks, not only on an error.

The pull request merges by itself once its checks pass.

## 3. The slice

In a new slice, `internal/<module>/internal/<feature>/`, follow `internal/notes/internal/list/`:

- `Args`: its `json` tags are the MCP arguments and the CLI flags, its `jsonschema` tags the descriptions agents read. An optional argument is a pointer (`*int`, `*string`), nil when not given. The binding knows `string`, `int`, `bool`, `[]string` and those two pointers: another type fails `TestEveryToolIsACommandWithTheSameArguments`, and teaching the binding a new one is a platform change. An operation that returns a list takes a `limit` with a default and a maximum, read with `op.Limit`, never a copy of it.
- `Out`: what the operation returns.
- The interface it needs from the infra, declared in the slice, named for what it does (`Saver`, `Reader`).
- `New`, then `Op()`: the name (`verb_noun`, snake_case), a description written for an agent (what it does, what it refuses), the effect (`op.ReadOnly`, `op.Additive`, `op.Destructive`), and an `Example`: one realistic call, which agents read in the tool's schema. The binding refuses an operation without one, and journals every operation that isn't read-only.
- `Run`: it turns every input into a domain type through the domain's `Parse` function, never by conversion.
- Its tests, in the black-box package `<feature>_test`: fakes of the interfaces the slice declares, never a mock library.

Then wire it in the module's contract, `internal/<module>/<module>.go`: `Register` and `Commands`. Run `go test ./internal/<module>/ -update` to record the tools MCP clients see in `testdata/tools.golden.json`.

## 4. Check before you stop

- `rumblestrip check`: every gate. It runs `rumblestrip spec`, which runs each built scenario alone, then once per value its `Then` steps expect, that value altered: it must pass as written and fail altered. A step that passes whatever it is told checks nothing.
- `rumblestrip mutate`: the new lines' mutants must die. A mutant that lives is a behaviour no test pins: add the test, or delete the code if nothing needs it.
- `rumblestrip review`: the sheet the owner reads. On the build's pull request it says there is nothing for them.

## When something is wrong

- **A unit test is wrong**: fix it in the pull request where you find it, and say why in its description.
- **A scenario is wrong**, or can't hold: never bend the code or the steps around it. Change it on the spec branch, or in a new spec pull request once it merged, and say why in one line: the owner approves it again.
