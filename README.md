# go-agent-kit

An anti-drift kit for Go code written by AI agents: the rules that keep a codebase maintainable, evolvable and safe are enforced by machines, not by reading every diff.

Target: Go CLIs, MCP servers and libraries.

- [`docs/kit.md`](docs/kit.md) — the kit: the constraints, the criteria, the solution for each, and the coverage check.
- Every rule the kit names is proven on [`trial/`](trial), a small project with one deliberate violation per rule; [`proof/`](proof) runs each tool on it and fails if a rule stops firing.

Status: milestone 1 (the doc and its proofs) in progress.
