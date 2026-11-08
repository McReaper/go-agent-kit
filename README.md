# Rumblestrip

Rumblestrip keeps Go code written by AI agents on the road: the rules that keep a codebase maintainable, evolvable and safe are enforced by machines. The owner reads the spec, Gherkin scenarios of what the app does, and a few lines per pull request; never the diff.

Target: Go CLIs, MCP servers and libraries.

## Use it in a repository

An app keeps four files for Rumblestrip and installs one GitHub App; the rest ships with Rumblestrip.

```sh
go run github.com/McReaper/rumblestrip-go/cmd/rumblestrip@<Rumblestrip version> init --owner <your login>   # rumblestrip.yml, rumblestrip-approval.yml, CODEOWNERS, dependabot.yml, .claude/settings.json
rumblestrip init app --name <name>              # once, if no GitHub App acts for the agent yet: one from Rumblestrip's manifest
rumblestrip init github --repo <owner>/<repo>   # as the owner, with the App's RUMBLESTRIP_APP_CLIENT_ID and RUMBLESTRIP_APP_KEY: the rulesets, merged branches deleted, immutable releases, the App Rumblestrip merges as
```

A new app starts from the [template](template) (`gonew`), which `rumblestrip init` then sets up.

## What it ships

- [`plugin/`](plugin) — the Claude Code plugin the repository's settings turn on: the skills (a feature spec first, a module, a secret), Rumblestrip's rules in every session, and the hooks: a gesture past Rumblestrip refused before it runs, each edit's package checked, the full check before the agent stops.
- `cmd/rumblestrip` — one CLI: `check` (every gate, the same on the agent's machine and in CI), `spec` (each step of the spec can fail), `guard` (bypasses, what is the owner's to approve, size), `review` (the short sheet the owner reads), `mutate` (the tests must kill the change's mutants), `sweep` (what only the whole module shows), `secrets` (the staged lines), `init`, `app` (acts as the agent's GitHub App), `hook` (Claude Code's hooks).
- `.github/workflows/rumblestrip-check.yml` — the merge check apps call: a job per gate, the review sheet and the owner's approval status, Check, the merge; `rumblestrip-approval.yml` re-runs the review sheet on the owner's approval.
- `lint/` — the curated golangci-lint and gitleaks configs; an app's own are never read. `cmd/rumblestrip-vet` — eleven analyzers for rules no linter has. `archtest` — the layout rules the compiler can't see.
- [`platform/`](platform) — what the owner applies by hand: Rumblestrip repository's own ruleset, the agent's managed settings.

## How it is proven

- [`docs/rumblestrip.md`](docs/rumblestrip.md) — Rumblestrip: the constraints, the criteria, the solution for each, the coverage check and what stays uncovered. [`docs/adr/`](docs/adr) — its decisions: supported Go only, one library per need, the lint runner, mutation on a copy, a tool before Rumblestrip's code.
- Every rule is proven on the template: it passes every tool, and each case under `trial/cases` holds one deliberate violation with the exact findings it must produce. [`proof/`](proof) reruns them all; `go test ./...` fails if a rule stops firing. `tools.txt` pins every tool at the version the proofs ran with.
