# go-agent-kit

An anti-drift kit for Go code written by AI agents: the rules that keep a codebase maintainable, evolvable and safe are enforced by machines, not by reading every diff.

Target: Go CLIs, MCP servers and libraries.

- [`docs/kit.md`](docs/kit.md) — the kit: the constraints, the criteria, the solution for each, the coverage check and what stays uncovered.
- [`docs/adr/`](docs/adr) — its decisions: supported Go only, one library per need, the lint runner, mutation on a copy.
- Every rule is proven on the [template](template), the app apps start from: it passes every tool, and each case under `trial/cases` holds one deliberate violation with the exact findings it must produce. [`proof/`](proof) reruns them all; `go test ./...` fails if a rule stops firing.

What it ships today:

- `lint/golangci.yml` — the curated golangci-lint config.
- `cmd/kitvet` — nine analyzers for rules no linter has (comments, nesting, test assertions, typed ids, goroutines, root contexts, package names, logged-and-returned errors, SQL built at run time).
- `cmd/kit guard` — the diff checks: bypasses, protected tests and go line, size.
- `archtest` — the layout rules the compiler can't see.
- `tools.txt` — every tool at the version the proofs ran with.

Status: milestone 1 (the doc and its proofs). Milestone 2 assembles the platform rules, the template and reference slice, and the skills.
