# platform

What the kit puts on GitHub and in Claude Code so the agent can't change the rules that judge it (criterion T2). The owner applies these; the agent can't.

## The agent's identity: a GitHub App

Repository permissions: Contents and Pull requests read and write; Checks, Actions and Commit statuses read; Metadata read. **Workflows and Administration: no access**, so the agent can't change CI or the protection rules, even through a pull request. No webhook. Installed on the chosen repositories only.

The agent acts as the app through `kit app exec -- <command>`, with `KIT_APP_CLIENT_ID`, `KIT_APP_KEY` (the private key's path, outside every repository) and `KIT_APP_INSTALLATION` set. `kit app installations` lists the installation ids.

## The default branch: `ruleset.json`

Apply it once per repository, as the owner:

```sh
gh api repos/OWNER/REPO/rulesets --method POST --input platform/ruleset.json
```

- a pull request, approved by a code owner on its last commit (stale approvals dismissed);
- the `kit` check passing;
- only the owner updates the branch, through a pull request (`update` rule, bypass by pull request only): the app can't merge even an approved change.

## Code owners: `CODEOWNERS`

Copy it to the app's `.github/CODEOWNERS`, with `@OWNER` replaced: tests, testdata, `go.mod`, `.github/`, `.kit/`, `.claude/`, AGENTS.md and the ADRs need the owner's approval.

## The merge check: `kit-check.yml`

The app's CI calls `.github/workflows/kit-check.yml` of this repository, pinned by SHA, on `pull_request`, `pull_request_review` and pushes to the default branch. It runs `kit check --ci`, approved once a code owner approved the last commit.

## Claude Code: `claude/settings.json`

Deny rules (force pushes, `--no-verify`, merging, editing `.kit/`, `.github/`, `.claude/`) and the hooks: `kit hook post-edit` after each edit, `kit hook stop` before the agent stops. Deploy it as the agent's managed settings, which the agent can't override (see Claude Code's managed settings documentation for the file's place on each system); deny rules are a speed bump, the platform is the arbiter.
