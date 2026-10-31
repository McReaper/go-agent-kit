---
name: new-module
description: How to add a module (a bounded context with its own domain, slices and storage) to this app. Use when a request brings words and rules the existing modules don't own, when asked to create a module, or when a feature would make one module import another's insides.
---

# A new module

A module owns a part of the app's domain: its words, its rules, its data. Other code reaches it only through its contract. Nested `internal/` makes the compiler refuse anything else.

## When it is a new module, not a new feature

- The request brings words the existing glossaries don't define, and rules that don't touch theirs.
- Its data is its own: a separate directory, table or file.
- Otherwise it is a feature of an existing module: load the `new-feature` skill.

When unsure, ask: a module is costly to merge back.

## The skeleton, in its own pull request

One module per pull request: the skeleton first, its features after, each with the `new-feature` skill.

```
internal/<m>/<m>.go            the contract: Open, Close, Register(s, j), Commands(j, out)
internal/<m>/GLOSSARY.md       each domain type: what it means, what makes one
internal/<m>/<m>_test.go       the golden test of its tools (testdata/tools.golden.json)
internal/<m>/internal/domain/  pure: types, their Parse functions, the rules
internal/<m>/internal/infra/   its storage, behind the interfaces its slices declare
```

Follow `internal/notes/`:

- The contract is the only package outside the module may import. It wires the slices and the infra, and holds no rule.
- The domain imports the standard library's pure packages and `pkg/` only: no I/O, no clock, no environment, no log.
- Every domain type that holds an identifier or a constrained value has a `Parse` function; outside the domain, nothing converts to it.
- `GLOSSARY.md` defines every exported domain type; a test fails when one is missing.
- Wire the module in `cmd/<app>/main.go`: open it under the data directory, pass it the journal.

## What the module never does

- Import another module's insides: it calls that module's contract.
- Put code in `internal/platform/` that names a module: the platform is shared by all.
