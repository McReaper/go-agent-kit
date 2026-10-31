---
name: new-feature
description: How to add an operation to a module of this app, tests first. An operation is one MCP tool and its CLI command, from one arguments struct. Use when asked to add a feature, a tool, a command or an operation, or to change what an existing one does or accepts.
---

# A new operation, tests first

The tests are written before the code and read as the behaviour the owner approves: the review sheet (`kit review`) shows each test's name as a sentence. Name each one for what a user would notice.

## 1. Say what it does before writing anything

In the pull request's description, write the operation's contract:

- its name (`verb_noun`, snake_case: `add_note`, `list_notes`);
- its arguments, the optional ones, and their limits;
- its result;
- its effect: `op.ReadOnly`, `op.Additive` (adds, never overwrites) or `op.Destructive` (may overwrite or delete);
- what it refuses, and why.

If the request leaves one of these open, ask before going on.

## 2. Write the tests, see them fail

In a new slice, `internal/<module>/internal/<feature>/`, write `<feature>_test.go` in the black-box package `<feature>_test`:

- one test per behaviour, named as its sentence: `TestListRefusesALimitOutOfRange` reads "List refuses a limit out of range";
- fakes of the interfaces the slice declares, never a mock library;
- every test fails on something it checks, not only on an error.

Run them. Each must fail, and for the reason you expect, not a compile error you didn't mean. If the owner asked to see tests first, open the pull request as a draft with the tests alone, and stop.

## 3. Write the slice

Follow `internal/notes/internal/list/`:

- `Args`: its `json` tags are the MCP arguments and the CLI flags, its `jsonschema` tags the descriptions agents read. An optional argument is a pointer (`*int`, `*string`), nil when not given. The binding knows `string`, `int`, `bool`, `[]string` and those two pointers: another type fails `TestEveryToolIsACommandWithTheSameArguments`, and teaching the binding a new one is a platform change. An operation that returns a list takes a limit with a default and a maximum.
- `Out`: what the operation returns.
- The interface it needs from the infra, declared in the slice, named for what it does (`Saver`, `Reader`).
- `New`, then `Op()`: the name, a description written for an agent (what it does, what it refuses), the effect. The binding journals every operation that isn't read-only.
- `Run`: it turns every input into a domain type through the domain's `Parse` function, never by conversion.

Then wire it in the module's contract, `internal/<module>/<module>.go`: `Register` and `Commands`.

## 4. Read the contract change

Run `go test ./internal/<module>/ -update`, then read the diff of `testdata/tools.golden.json`: it is what MCP clients will see, and the owner reviews it.

## 5. Check before you stop

- `kit check`: every gate.
- `kit mutate`: the new lines' mutants must die. A mutant that lives is a behaviour no test pins: add the test, or delete the code if nothing needs it.
- `kit review`: read the sheet as the owner will.

## When a test is wrong

Fix the test, and say in your report what changed in it and why. The owner approves every changed test before merge. Never bend the code to fit a test you believe is wrong.
