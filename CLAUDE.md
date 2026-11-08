# CLAUDE.md

Rumblestrip: the rails AI-written Go stays on. `docs/rumblestrip.md` holds every rule and its proof; read it before changing a rule, and `docs/adr/` before reopening a decision.

## Layout

- `lint/golangci.yml` — the curated golangci-lint config apps run (never their own).
- `analyzers/<name>/` + `cmd/rumblestrip-vet` — Rumblestrip's go/analysis analyzers, each tested with analysistest; `analyzers/internal/layout` says which part of the architecture a package is.
- `guard/` + `cmd/rumblestrip` — `rumblestrip guard`, the diff checks.
- `spec/` — the spec: Gherkin scenarios under an app's `features/`, what a change does to them, and `rumblestrip spec`, which shows each step can fail.
- `gesture/` — what `rumblestrip hook pre-tool` refuses before the agent runs it.
- `setup/` — `rumblestrip init`: the files an app keeps, its GitHub settings, the agent's App; `setup/files/` holds them, the app rulesets included.
- `.github/workflows/rumblestrip-check.yml` and `rumblestrip-approval.yml` — the checks an app's workflows call: one job per check, the review sheet, Check, the merge; their header comments say how they run.
- `plugin/` + `.claude-plugin/marketplace.json` — the Claude Code plugin apps turn on: skills, rules, hooks; `plugin/bin/rumblestrip` runs rumblestrip at the version the app's CI pins.
- `archtest/` — architecture rules over go/packages.
- `internal/platform/git` — the one place Rumblestrip starts git.
- `template/` — the app template (`gonew github.com/McReaper/rumblestrip-go/template`), its own module, and the trial every proof starts from: passes every gate. `trial/cases/<case>/` — one deliberate violation, the files it adds or replaces, and `want.txt` (`<tool>/<rule> <file>:<line> <regexp>`; `!isolated`, `!delete <file>`, `!expect-none <tool>`).
- `proof/` — runs every tool on the trial and compares with each `want.txt`; also checks `docs/rumblestrip.md` against the cases and `tools.txt`.
- `tools.txt` — every tool's pinned version; CI installs exactly these.

## Rules

- No config from memory: a rule enters `docs/rumblestrip.md` only once a case proves it, with the exact finding (tool, file, line, message).
- A case that runs one tool only: mark it `!isolated` if it replaces a file others build on (go.mod).
- Never run go-mutesting on `template`: it rewrites the source while tests run. The proof runs it on copies.
- Rumblestrip passes its own rules: `golangci-lint run --config lint/golangci.yml ./...` and rumblestrip-vet, for GOOS linux and windows.
- Moving the `go` or `toolchain` line follows ADR 1; `go get x@latest` can move it silently, so check go.mod after any `go get`.
- Everything in English. Conventional commits. Checks are Go, not scripts in another language.
- People read `README.md`, `docs/guide.md` and `docs/setup.md`: plain words, no criterion IDs, within the word budgets `TestTheHumanDocsStayShort` enforces. The criteria go in `docs/rumblestrip.md`.
- This repository's own ruleset is `platform/ruleset-rumblestrip.json`, applied by hand: `gh api repos/McReaper/rumblestrip-go/rulesets --method POST --input platform/ruleset-rumblestrip.json`. Its CI job stays `kit`, the context it requires.

## Commands

- Everything: `go test ./...` (needs the tools in `tools.txt` on the PATH; zizmor through `ZIZMOR=<path>` if pip put it elsewhere). The race proof is Linux only.
- One proof: `go test ./proof/ -run TestGuardReportsEveryCase`.
