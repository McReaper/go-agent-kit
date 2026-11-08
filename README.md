# Rumblestrip

An AI agent writes your Go code. Rumblestrip keeps it on the road: every change goes through checks the agent can't switch off, and nothing merges until it passes them. You read one thing: what the app should do.

## How a change goes

1. **The spec first.** The agent writes what the app will do, as short scenarios in plain English (`features/*.feature`), and opens a pull request.
2. **You approve it.** It's the only thing you read. The pull request carries the `human-review` label until you do.
3. **The build.** The agent writes the code in a second pull request. About twenty checks run on it: the tests, the spec, lint, security, and more.
4. **The merge.** Once every check passes, Rumblestrip merges the spec, then the build. If a check fails, the agent fixes it; you do nothing.

On every pull request, Rumblestrip's comment says in one line where things stand.

## Set it up

In a Go repository on GitHub, with Go and the `gh` CLI:

```sh
go install github.com/McReaper/rumblestrip-go/cmd/rumblestrip@latest
rumblestrip init --owner <your GitHub login>   # the workflows and settings: commit and push them
rumblestrip init app --name <agent name>       # once: the agent's GitHub App, made in your browser; install it on the repository
rumblestrip init github --repo <you>/<repo>    # the rules GitHub enforces, with the App's RUMBLESTRIP_APP_CLIENT_ID and RUMBLESTRIP_APP_KEY
```

Then open Claude Code in the repository and ask for a feature. A new app can start from the [template](template).

## More

- [The guide](docs/guide.md): reading a pull request, approving, when something is stuck.
- [Setup in detail](docs/setup.md): what each command sets, and what the agent can't do.
- [The rules](docs/rumblestrip.md): every criterion, the check that holds it, and its proof.
- [Contributing](CLAUDE.md).
