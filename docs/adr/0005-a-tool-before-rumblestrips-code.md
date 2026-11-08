# 5. A rule is an existing tool's before it is Rumblestrip's code

## Decision

Rumblestrip writes a check in Go only when no tool it already runs can hold the rule, proven on the rule's trial case. Each time a tool can, Rumblestrip's code goes.

## What replaced Rumblestrip's code, and what didn't

| Rule | Tried | Result |
| --- | --- | --- |
| A function copied from another module (M4) | dupl at 35 tokens, in golangci-lint | Replaces `rumblestrip sweep`'s copied-function rule: catches the case, nothing on the template |
| Keys, certificates and env files committed (S1) | a gitleaks path rule, in Rumblestrip's gitleaks config | Replaces the guard's sensitive-file rule |
| `t.Skip`, `testing.Short` (T2) | forbidigo, in golangci-lint | Replaces the guard's skip rule |
| Mutation on changed lines (M6) | gremlins | Kept `rumblestrip mutate` over go-mutesting: on Windows gremlins reports every mutant not covered |
| Architecture rules (E1, E2) | go-arch-lint | Kept `archtest`: on Windows its mapping folds every package into one component |
| The guard's other bypasses: `//nolint`, `//go:linkname`, cgo, hidden directories, unknown build tags, a generated header outside `gen/`, `TestMain` | semgrep | Caught them all, but needs Python in every app's CI: kept in the guard, about 50 lines |
| The review sheet, size and module rules | Danger | Needs Node and rules in JavaScript: kept in Go, the sheet cut to the spec and the owner's approvals |
| rumblestrip-vet's analyzers: comments, nesting, test assertions, typed ids, goroutines, root contexts, package names, logged-and-returned errors, SQL built at run time, domain conversions, bounded lists | golangci-lint v2.14's linters | None holds the rule: revive's nesting counts `if` only, gosec misses SQL built through a variable, no linter knows the template's layout |

## Why

Every line Rumblestrip writes is a line someone maintains and an agent can misread; a tool the ecosystem maintains carries its own tests. A tool that needs another runtime in every app's CI costs more than the lines it saves.
