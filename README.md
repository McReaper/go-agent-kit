# go-agent-kit

An anti-drift kit for Go code written by AI agents: the rules that keep a codebase maintainable, evolvable and safe are enforced by machines, not by reading every diff.

Target: Go CLIs, MCP servers and libraries.

- [`docs/kit.md`](docs/kit.md) — the kit: the constraints, the criteria, the solution for each, the coverage check and what stays uncovered.
- [`docs/adr/`](docs/adr) — its decisions: supported Go only, one library per need, the lint runner, mutation on a copy.
- Every rule is proven on the [template](template), the app apps start from: it passes every tool, and each case under `trial/cases` holds one deliberate violation with the exact findings it must produce. [`proof/`](proof) reruns them all; `go test ./...` fails if a rule stops firing.

What it ships:

- `lint/golangci.yml` — the curated golangci-lint config; an app's own is never read.
- `cmd/kitvet` — eleven analyzers for rules no linter has: comments, nesting, test assertions, typed ids, goroutines, root contexts, package names, logged-and-returned errors, SQL built at run time, domain types built by conversion, unbounded lists.
- `cmd/kit` — one CLI: `check` (every gate, the same on the agent's machine and in CI), `guard` (bypasses, protected changes, size), `sweep` (what only the whole module shows), `mutate` (the tests must kill the change's mutants), `review` (the sheet the owner reads), `sync` (brings an app's kit-owned files to a kit version), `app` (acts as the agent's GitHub App), `hook` (Claude Code's hooks).
- `archtest` — the layout rules the compiler can't see.
- [`template/`](template) — the app an app starts from, with `gonew`: the reference slices, AGENTS.md, the skills, the smoke test on the built program.
- [`platform/`](platform) — what the owner applies on GitHub and in Claude Code: the rulesets, the App's permissions, the agent's settings; `.github/workflows/kit-check.yml`, the merge check apps call.
- `tools.txt` — every tool at the version the proofs ran with.

Status: milestones 1 and 2 are done: every .NET solution has its Go counterpart, proven, and the outside review's evolutions are in. Next: the kit on a real project, measuring each gate failure and how the agent resolved it.
