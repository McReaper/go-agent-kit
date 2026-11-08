# go-agent-kit

An anti-drift kit for Go code written by AI agents: the rules that keep a codebase maintainable, evolvable and safe are enforced by machines. The owner reads the spec, Gherkin scenarios of what the app does, and a few lines per pull request; never the diff.

Target: Go CLIs, MCP servers and libraries.

## Use it in a repository

An app keeps four files for the kit and installs one GitHub App; the rest ships with the kit.

```sh
go run github.com/McReaper/go-agent-kit/cmd/kit@<kit version> init --owner <your login>   # kit.yml, kit-approval.yml, CODEOWNERS, dependabot.yml, .claude/settings.json
kit init app --name <name>              # once, if no GitHub App acts for the agent yet: one from the kit's manifest
kit init github --repo <owner>/<repo>   # as the owner, with the App's KIT_APP_CLIENT_ID and KIT_APP_KEY: the rulesets, merged branches deleted, immutable releases, the App the kit merges as
```

A new app starts from the [template](template) (`gonew`), which `kit init` then sets up.

## What it ships

- [`plugin/`](plugin) — the Claude Code plugin the repository's settings turn on: the skills (a feature spec first, a module, a secret), the kit's rules in every session, and the hooks: a gesture past the kit refused before it runs, each edit's package checked, the full check before the agent stops.
- `cmd/kit` — one CLI: `check` (every gate, the same on the agent's machine and in CI), `spec` (each step of the spec can fail), `guard` (bypasses, what is the owner's to approve, size), `review` (the short sheet the owner reads), `mutate` (the tests must kill the change's mutants), `sweep` (what only the whole module shows), `secrets` (the staged lines), `init`, `app` (acts as the agent's GitHub App), `hook` (Claude Code's hooks).
- `.github/workflows/kit-check.yml` — the merge check apps call: a job per gate, the review sheet and the owner's approval status, Check, the merge; `kit-approval.yml` re-runs the review sheet on the owner's approval.
- `lint/` — the curated golangci-lint and gitleaks configs; an app's own are never read. `cmd/kitvet` — eleven analyzers for rules no linter has. `archtest` — the layout rules the compiler can't see.
- [`platform/`](platform) — what the owner applies by hand: the kit repository's own ruleset, the agent's managed settings.

## How it is proven

- [`docs/kit.md`](docs/kit.md) — the kit: the constraints, the criteria, the solution for each, the coverage check and what stays uncovered. [`docs/adr/`](docs/adr) — its decisions: supported Go only, one library per need, the lint runner, mutation on a copy, a tool before the kit's code.
- Every rule is proven on the template: it passes every tool, and each case under `trial/cases` holds one deliberate violation with the exact findings it must produce. [`proof/`](proof) reruns them all; `go test ./...` fails if a rule stops firing. `tools.txt` pins every tool at the version the proofs ran with.
