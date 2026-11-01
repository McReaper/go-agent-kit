# 4. Mutation testing with go-mutesting, on a copy, on changed code

## Context

gremlins has `--diff` and thresholds but swaps operators only, and on Windows reports every mutant as not covered. go-mutesting removes statements, branches and expressions as well, and runs on Windows; it writes each mutant over the source file while the tests run, and rapid saves failing inputs under `testdata/`.

Go handles errors in code: every `if err != nil { return …err… }` is a branch go-mutesting removes, and that mutant lives unless a test makes the call fail. On the template they were 29 of 110 mutants, 24 of them in the infra and the platform, where failing a disk or a file system call in a test rarely pays. Meanwhile the compiler refuses an `err` never used, and errcheck and ineffassign refuse one dropped.

## Decision

`kit mutate` runs go-mutesting on a copy of the module, the files git sees (tracked and new, ignored ones left out, so build outputs and databases aren't copied), in the packages a change touches. Only mutants that start on a line the change adds count, so old code's survivors don't fail a new change. A mutant that only removes the return of an `if err != nil` block isn't counted, and the report says how many there were. It fails under `--min`, 0.8, calibrated on the template (see `docs/kit.md`, M6). Equivalent mutants are why the gate is a score, not zero survivors.

## Consequences

- A mutant counts on the line it starts: removing a block whose first line the change left alone isn't counted.
- An error's return goes untested by mutation; a refusal (`errors.Is(err, fs.ErrExist)`, a sentinel error) is a behaviour and stays counted.
- Each mutant is a recompile plus the package's tests, so the gate runs in the merge check, not in the agent's stop hook.
