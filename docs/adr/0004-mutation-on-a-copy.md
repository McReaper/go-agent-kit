# 4. Mutation testing with go-mutesting, on a copy, on changed code

## Context

gremlins has `--diff` and thresholds but swaps operators only, and on Windows reports every mutant as not covered. go-mutesting removes statements, branches and expressions as well, and runs on Windows; it writes each mutant over the source file while the tests run, and rapid saves failing inputs under `testdata/`.

## Decision

`kit mutate` runs go-mutesting on a copy of the module, the files git sees (tracked and new, ignored ones left out, so build outputs and databases aren't copied), in the packages a change touches. Only mutants that start on a line the change adds count, so old code's survivors don't fail a new change. It fails under `--min`, 0.8 until the reference slice calibrates it. Equivalent mutants are why the gate is a score, not zero survivors.

## Consequences

- A mutant counts on the line it starts: removing a block whose first line the change left alone isn't counted.
- Each mutant is a recompile plus the package's tests, so the gate runs in the merge check, not in the agent's stop hook.
