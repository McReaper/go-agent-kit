# CLAUDE.md

The anti-drift kit for AI-written Go. `docs/kit.md` is the kit; read it before changing a rule.

## Layout

- `lint/` — the curated golangci-lint config apps run (never their own).
- `analyzers/<name>/` — custom `go/analysis` analyzers, each tested with `analysistest`.
- `archtest/` — architecture rules over `go/packages`, called from one test in each app.
- `cmd/kit/` — the kit's CLI: `lint`, `guard`, `contract`.
- `trial/` — the trial project, its own module: `trial/clean` passes every gate; `trial/cases/<rule>/` holds one deliberate violation per rule, annotated where it must fire.
- `proof/` — tests that run every tool on `trial/` and compare with the annotations.
- `docs/adr/` — the kit's decisions.

## Rules

- No config from memory: a rule enters `docs/kit.md` only once its proof case passes.
- A proof case asserts the exact finding (tool, file, line, message pattern), not just a non-zero exit.
- The kit's checks are Go, not scripts in another language.
- Everything in English. Conventional commits.

## Commands

- Test everything: `go test ./...` (the proofs need the tools listed in `proof/tools.go` on the PATH).
- The base `go` here may be older; `go.mod`'s `go` line makes it fetch the right toolchain.
