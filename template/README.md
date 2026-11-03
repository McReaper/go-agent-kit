# notes

Short notes for people and agents, served as MCP tools (`notes mcp`) and CLI commands (`notes add-note`, `notes list-notes`). It is the anti-drift kit's template: an app starts as a copy of it.

## Start an app from it

```sh
go run golang.org/x/tools/cmd/gonew@latest github.com/McReaper/go-agent-kit/template@<kit version> example.com/yourapp
cd yourapp && git init
go run github.com/McReaper/go-agent-kit/cmd/kit@<kit version> init --owner <your login>
```

`kit init` writes the four files the repository keeps for the kit: the CI that calls its merge check, the code owners of the harness, Dependabot's settings, and the Claude Code settings that turn the kit's plugin on. Then, as the owner, once:

1. Push the repository to GitHub, and run `kit init github --repo <owner>/<repo>`: the rulesets, auto-merge, immutable releases.
2. If the agent has no GitHub App yet, `kit init app --name <name>`, and install the App on the repository.
3. Replace `LICENSE`: gonew copies the kit's.
4. Rename the notes module, or keep it as the reference while you build your first one, then delete it.

The owner reads `features/`, the spec, and each pull request's review sheet; AGENTS.md is what the agent reads; the kit's `docs/kit.md` is why each rule exists.
