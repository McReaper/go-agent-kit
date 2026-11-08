<div align="center">
    <img src="docs/media/logo.svg" alt="Rumblestrip logo" width="160" height="160"/>
    <h1>Rumblestrip</h1>
    <h3><em>Your AI agent writes the Go. Rumblestrip keeps it on the road.</em></h3>
</div>

<p align="center">
    <a href="https://github.com/McReaper/rumblestrip-go/tags"><img src="https://img.shields.io/github/v/tag/McReaper/rumblestrip-go?sort=semver&filter=v*&label=version" alt="Latest version"/></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="License: MIT"/></a>
    <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26"/>
    <img src="https://img.shields.io/badge/Claude_Code-plugin-D97757" alt="Claude Code plugin"/>
</p>

<p align="center">
    <a href="#-what-you-get">What you get</a> ·
    <a href="#%EF%B8%8F-how-a-change-goes">How a change goes</a> ·
    <a href="#-set-it-up">Set it up</a> ·
    <a href="#-docs">Docs</a> ·
    <a href="#-license">License</a>
</p>

A rumble strip is the grooved edge of a road: drift onto it and it rumbles, so you steer back. Rumblestrip does that for an AI agent writing Go: every rule it crosses rumbles with an error that says the fix, and nothing merges until every check passes. You read one thing: what the app should do.

## ✨ What you get

- **You approve the what, never the how** — the agent writes the spec first, short scenarios in plain English. You read and approve those; the code is the checks' to judge.
- **About twenty checks the agent can't switch off** — tests, the spec, lint, architecture, dead code, nil safety, known vulnerabilities, secrets, the workflows. They live in Rumblestrip, pinned by version, out of the agent's reach.
- **A rumble before CI** — in Claude Code, Rumblestrip checks each package the agent edits, refuses a gesture past the rules (a force push, editing CI, skipping hooks), and runs every check before the agent stops.
- **One line that says where things stand** — on each pull request: the checks running, the ones that fail, waiting for you, merged.
- **Merges by itself** — once every check passes and you've approved what's yours, Rumblestrip merges the spec, then the build, and credits you as co-author.

## 🛣️ How a change goes

1. **The spec.** The agent writes what the app will do and opens a pull request, labelled `human-review`.
2. **Your approval.** You read the scenarios and approve. That's all you read.
3. **The build.** The agent writes the code in a second pull request, stacked on the spec, and the checks run on it. If one fails, the agent fixes it.
4. **The merge.** Once every check passes, Rumblestrip merges both.

> [!NOTE]
> <details>
> <summary><b>🔍 Click to see the whole path</b>: who does what, and where the checks stand</summary>
>
> ```mermaid
> flowchart LR
>     you(["You"]) -->|"ask for a feature"| agent["AI agent"]
>     agent -->|"1 · writes"| spec["Spec pull request<br/>scenarios in plain English"]
>     you -.->|"2 · read and approve"| spec
>     agent -->|"3 · writes"| build["Build pull request<br/>the code, stacked on the spec"]
>     spec --> checks{{"~20 checks<br/>tests, spec, lint, security…"}}
>     build --> checks
>     checks -->|"one fails: the error says the fix"| agent
>     checks -->|"4 · all pass, you approved"| merge["Rumblestrip merges<br/>the spec, then the build"]
>     merge --> main[("main")]
> ```
>
> The agent never merges, and can't change the checks or the rules GitHub enforces: only you can.
>
> </details>

## 📥 Set it up

In a Go repository on GitHub, with Go and the `gh` CLI:

```sh
go install github.com/McReaper/rumblestrip-go/cmd/rumblestrip@latest
rumblestrip init --owner <your GitHub login>   # the workflows and settings: commit and push them
rumblestrip init app --name <agent name>       # once: the agent's GitHub App; install it on the repository
rumblestrip init github --repo <you>/<repo>    # the rules GitHub enforces, with the App's RUMBLESTRIP_APP_CLIENT_ID and RUMBLESTRIP_APP_KEY
```

Then open Claude Code in the repository and ask for a feature. A new app can start from the [template](template).

## 📚 Docs

- [The guide](docs/guide.md) — reading a pull request, approving, when something is stuck.
- [Setup in detail](docs/setup.md) — what each command sets, and what the agent can't do.
- [The rules](docs/rumblestrip.md) — every criterion, the check that holds it, and its proof.
- [Contributing](CLAUDE.md).

## 📄 License

MIT — see [`LICENSE`](LICENSE).
