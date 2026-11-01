# CLAUDE.md

The anti-drift kit for AI-written Go. `docs/kit.md` is the kit; read it before changing a rule, and `docs/adr/` before reopening a decision.

## Layout

- `lint/golangci.yml` — the curated golangci-lint config apps run (never their own).
- `analyzers/<name>/` + `cmd/kitvet` — the kit's go/analysis analyzers, each tested with analysistest; `analyzers/internal/layout` says which part of the architecture a package is.
- `guard/` + `cmd/kit` — `kit guard`, the diff checks.
- `archtest/` — architecture rules over go/packages.
- `internal/platform/git` — the one place the kit starts git.
- `template/` — the app template (`gonew github.com/McReaper/go-agent-kit/template`), its own module, and the trial every proof starts from: passes every gate. `trial/cases/<case>/` — one deliberate violation, the files it adds or replaces, and `want.txt` (`<tool>/<rule> <file>:<line> <regexp>`; `!isolated`, `!delete <file>`, `!expect-none <tool>`).
- `proof/` — runs every tool on the trial and compares with each `want.txt`; also checks `docs/kit.md` against the cases and `tools.txt`.
- `tools.txt` — every tool's pinned version; CI installs exactly these.

## Rules

- No config from memory: a rule enters `docs/kit.md` only once a case proves it, with the exact finding (tool, file, line, message).
- A case that runs one tool only: mark it `!isolated` if it replaces a file others build on (go.mod).
- Never run go-mutesting on `template`: it rewrites the source while tests run. The proof runs it on copies.
- The kit passes its own rules: `golangci-lint run --config lint/golangci.yml ./...` and kitvet, for GOOS linux and windows.
- Moving the `go` or `toolchain` line follows ADR 1; `go get x@latest` can move it silently, so check go.mod after any `go get`.
- Everything in English. Conventional commits. Checks are Go, not scripts in another language.

## Commands

- Everything: `go test ./...` (needs the tools in `tools.txt` on the PATH; zizmor through `ZIZMOR=<path>` if pip put it elsewhere). The race proof is Linux only.
- One proof: `go test ./proof/ -run TestGuardReportsEveryCase`.
