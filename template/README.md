# notes

Short notes for people and agents, served as MCP tools (`notes mcp`) and CLI commands (`notes add-note`, `notes list-notes`). It is Rumblestrip's template: an app starts as a copy of it, then grows its own modules.

## Start an app from it

```sh
go run golang.org/x/tools/cmd/gonew@latest github.com/McReaper/rumblestrip-go/template@<Rumblestrip version> example.com/yourapp
cd yourapp && git init
```

Then follow Rumblestrip's [setup](https://github.com/McReaper/rumblestrip-go#set-it-up), and:

- replace `LICENSE`, which gonew copied from Rumblestrip;
- keep the notes module as the reference while you build your first one, then delete it.

You read `features/`, the spec. The agent reads `AGENTS.md`.
