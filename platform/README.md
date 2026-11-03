# platform

What the kit puts on GitHub and in Claude Code so the agent can't change the rules that judge it (criterion T2). The owner applies these; the agent can't. `kit init` does most of it (see the repository's README); this page says what each part holds.

## The agent's identity: a GitHub App

`kit init app --name <name>` creates it from the kit's manifest, through the owner's browser, and saves its private key in the owner's configuration directory. Repository permissions: Contents and Pull requests read and write; Checks, Actions and Commit statuses read; Metadata read. **Workflows and Administration: no access**, so the agent can't change CI or the protection rules, even through a pull request. No webhook. Installed on the chosen repositories only.

The agent acts as the app through `kit app exec -- <command>`, with `KIT_APP_CLIENT_ID`, `KIT_APP_KEY` (the private key's path, outside every repository) and `KIT_APP_INSTALLATION` set. `kit app installations` lists the installation ids.

## An app's GitHub settings: `kit init github`

`kit init github --repo OWNER/REPO`, as the owner, applies `kitinit/files/ruleset.json` and `ruleset-tags.json` (creating each, or bringing an existing one of the same name to the kit's), allows auto-merge and the deletion of merged branches, and makes releases immutable.

- The default branch: a pull request, squash-merged; no review required but the code owners' on the harness, and every approval stays on record; the `kit / kit` check passing, reported by GitHub Actions only, which holds the spec, a new dependency, the go line and a library's API until the owner approved the content they have; no `update` rule, so GitHub's auto-merge, which `kit-check.yml` turns on, lands the change once both hold.
- Tags and releases: only the owner creates, moves or deletes a tag (the admin role bypasses, and pushes release tags from its own clone); once published, a release's tag and assets can't change.

## The kit's own default branch: `ruleset-kit.json`

The kit judges every app, so its owner merges each change to it: a code owner (`*`) approves the last commit, the `kit` check passes, and only the owner updates the branch, through a pull request. Apply it by hand: `gh api repos/OWNER/go-agent-kit/rulesets --method POST --input platform/ruleset-kit.json`.

## The merge check: `kit-check.yml`

An app's `.github/workflows/kit.yml`, written by `kit init`, calls `.github/workflows/kit-check.yml` of this repository at the commit its one `uses` line pins; Dependabot moves that pin to each new release tag. It runs on `pull_request` (an edit counts when it moved the base), `pull_request_review` and pushes to the default branch: `kit check --ci` with the trees of the commits a code owner approved, so a change that is the owner's passes once approved, and stays approved while its files hold the content approved; it posts the review sheet, and turns GitHub's auto-merge on for every pull request into the default branch.

## Claude Code: `claude/managed-settings.json`

The repository's `.claude/settings.json`, written by `kit init`, turns the kit's plugin on and holds the deny rules a plugin can't carry: a force push without a lease, any force push to `main`, `--no-verify`, merging, changing a ruleset, editing `.kit/`, `.github/`, `.claude/`; `--force-with-lease` stays allowed on the agent's own branches, so it can rebase its pull request. The agent could edit that file; deploy `claude/managed-settings.json`, the same, as the machine's managed settings (`C:\Program Files\ClaudeCode\managed-settings.json` on Windows, `/Library/Application Support/ClaudeCode/` on macOS, `/etc/claude-code/` on Linux), which no other settings override. Deny rules and hooks are a speed bump; the platform is the arbiter.
