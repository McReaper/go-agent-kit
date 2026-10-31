# notes

Short notes for people and agents, served as MCP tools (`notes mcp`) and CLI commands (`notes add-note`, `notes list-notes`). It is the anti-drift kit's template: an app starts as a copy of it.

## Start an app from it

```sh
go run golang.org/x/tools/cmd/gonew@latest github.com/McReaper/go-agent-kit/template@<kit version> example.com/yourapp
```

Then, once:

1. In `.github/CODEOWNERS`, replace `@OWNER` with your login.
2. In `.github/workflows/kit.yml`, pin both SHAs to the kit version you started from.
3. Apply the kit's ruleset and deploy its Claude Code settings: the kit's `platform/README.md` says how.
4. Enable the pre-commit hook: `git config core.hooksPath .githooks`.
5. Rename the notes module, or keep it as the reference while you build your first one, then delete it.

AGENTS.md is what the agent reads; the kit's `docs/kit.md` is why each rule exists.
