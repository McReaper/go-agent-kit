# notes

Short notes for people and agents, served as MCP tools (`notes mcp`) and CLI commands (`notes add-note`, `notes list-notes`). It is Rumblestrip's template: an app starts as a copy of it.

## Start an app from it

```sh
go run golang.org/x/tools/cmd/gonew@latest github.com/McReaper/rumblestrip-go/template@<Rumblestrip version> example.com/yourapp
cd yourapp && git init
go run github.com/McReaper/rumblestrip-go/cmd/rumblestrip@<Rumblestrip version> init --owner <your login>
```

`rumblestrip init` writes the four files the repository keeps for Rumblestrip: the CI that calls its merge check, the code owners of the harness, Dependabot's settings, and the Claude Code settings that turn Rumblestrip's plugin on. Then, as the owner, once:

1. If the agent has no GitHub App yet, `rumblestrip init app --name <name>`. Install the App, that one or one you have, on the repository.
2. Push the repository to GitHub, and run `rumblestrip init github --repo <owner>/<repo>` with the App's `RUMBLESTRIP_APP_CLIENT_ID` and `RUMBLESTRIP_APP_KEY`: the rulesets, merged branches deleted, immutable releases, and the App Rumblestrip merges as.
3. Replace `LICENSE`: gonew copies Rumblestrip's.
4. Rename the notes module, or keep it as the reference while you build your first one, then delete it.

The owner reads `features/`, the spec, and each pull request's review sheet; AGENTS.md is what the agent reads; Rumblestrip's `docs/rumblestrip.md` is why each rule exists.
