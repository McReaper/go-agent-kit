# 4. Mutation testing with go-mutesting, on a copy, on changed code

## Context

gremlins has `--diff` and thresholds but swaps operators only, and on Windows reports every mutant as not covered. go-mutesting removes statements, branches and expressions as well, and runs on Windows; it writes each mutant over the source file while the tests run, and rapid saves failing inputs under `testdata/`.

## Decision

`kit mutate` (milestone 2) runs go-mutesting on a copy of the module, on the lines a change touches, and fails under a score calibrated on the reference slice. Equivalent mutants are why the gate is a score, not zero survivors.
