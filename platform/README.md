# platform

What Rumblestrip puts on GitHub and in Claude Code so the agent can't change the rules that judge it (criterion T2). The owner applies these; the agent can't. `rumblestrip init` does most of it (see the repository's README); this page says what each part holds.

## The agent's identity: a GitHub App

`rumblestrip init app --name <name>` creates it from Rumblestrip's manifest, through the owner's browser, and saves its private key in the owner's configuration directory. Repository permissions: Contents and Pull requests read and write; Checks, Actions and Commit statuses read; Metadata read. **Workflows and Administration: no access**, so the agent can't change CI or the protection rules, even through a pull request. No webhook. Installed on the chosen repositories only.

The agent acts as the app through `rumblestrip app exec -- <command>`, with `RUMBLESTRIP_APP_CLIENT_ID`, `RUMBLESTRIP_APP_KEY` (the private key's path, outside every repository) and `RUMBLESTRIP_APP_INSTALLATION` set. `rumblestrip app installations` lists the installation ids. Rumblestrip's merge job merges as an App too: any App with Contents and Pull requests write, one the owner already has included.

## An app's GitHub settings: `rumblestrip init github`

`rumblestrip init github --repo OWNER/REPO`, as the owner, applies `setup/files/ruleset.json` and `ruleset-tags.json` (creating each, or bringing an existing one of the same name to Rumblestrip's), deletes merged branches, allows squash merges, and makes releases immutable. It stores the Client ID and key of the App the merge job merges as (`RUMBLESTRIP_APP_CLIENT_ID` and `RUMBLESTRIP_APP_KEY`, from those variables or `--client-id` and `--key`) as the repository's secrets, for Actions and for Dependabot, once GitHub says that App is installed on the repository with Contents and Pull requests write; it checks the App before it changes anything.

- The default branch: a pull request, squash-merged; no review required but the code owners' on the harness, and every approval stays on record; two checks, both reported by GitHub Actions only: `Pull request / Check`, every gate passed, and `Rumblestrip / Owner approval`, which holds the spec, a new dependency, the go line and a library's API until the owner approved the content they have; no `update` rule, so Rumblestrip's merge job lands the change once both hold.
- Tags and releases: only the owner creates, moves or deletes a tag (the admin role bypasses, and pushes release tags from its own clone); once published, a release's tag and assets can't change. A release of Rumblestrip tags one commit twice, `vX` and `template/vX`: the template is a module of its own, and `gonew` finds it at a version only through its tag.

## Rumblestrip's own default branch: `ruleset-rumblestrip.json`

Rumblestrip judges every app, so its owner merges each change to it: a code owner (`*`) approves the last commit, the `kit` check passes (its own CI's job), and only the owner updates the branch, through a pull request. Apply it by hand: `gh api repos/OWNER/Rumblestrip/rulesets --method POST --input platform/ruleset-rumblestrip.json`.

## The merge check: `rumblestrip-check.yml`

An app's `.github/workflows/rumblestrip.yml` and `rumblestrip-approval.yml`, written by `rumblestrip init`, call `.github/workflows/rumblestrip-check.yml` and `rumblestrip-approval.yml` of this repository at the commit their `uses` lines pin; Dependabot moves that pin to each new release tag. `rumblestrip.yml` runs on `pull_request` (an edit counts when it changed the title or the base), as the `Pull request` job:

- one job per gate of `rumblestrip check`, all at once, plus the title and `rumblestrip mutate`, each named for the owner (`Tests`, `Spec`, `Known vulnerabilities`), with an annotation on each finding's line; a failing one hands Check its section of the run's one summary: what it wants and what it found;
- `Review sheet`: Rumblestrip's comment, the `human-review` label, and the `Rumblestrip / Owner approval` status, from the trees of the commits a code owner approved, so a change that is the owner's is approved once, and stays approved while its files hold the content approved. It is the one job allowed to post statuses, and runs none of the change's code;
- `Check`, the required one: every gate's job passed. It never skips (GitHub counts a skipped check as passed), and rewrites the comment's status line: which checks failed, or what comes next;
- `Merge`.

The merge job merges a pull request into the default branch once Check passed and the owner's approval holds, through GitHub's asynchronous merge: GitHub's auto-merge takes no pull request of a stack. The commit is the pull request's title and description, with the owner as co-author and Rumblestrip's release as a `Built-with` trailer. It merges as the repository's App: after a merge by the workflow's own token, GitHub deletes no branch, and moves a stack's next pull request onto the default branch without rebasing it or running its check. A pull request on another branch must be in a stack, whose next pull request GitHub rebases onto the default branch once the one below merged. The owner's approval runs `rumblestrip-approval.yml`, which re-runs the review sheet's job of the approved commit's run, then Check and Merge after it.

## Claude Code: `claude/managed-settings.json`

The repository's `.claude/settings.json`, written by `rumblestrip init`, turns Rumblestrip's plugin on, makes the agent's commits and pull requests name Rumblestrip rather than Claude Code (`attribution`), and holds the deny rules a plugin can't carry: a force push without a lease, any force push to `main`, `--no-verify`, merging, changing a ruleset, editing `.rumblestrip/`, `.github/`, `.claude/`; `--force-with-lease` stays allowed on the agent's own branches, so it can rebase its pull request. The agent could edit that file; deploy `claude/managed-settings.json`, the same, as the machine's managed settings (`C:\Program Files\ClaudeCode\managed-settings.json` on Windows, `/Library/Application Support/ClaudeCode/` on macOS, `/etc/claude-code/` on Linux), which no other settings override. Deny rules and hooks are a speed bump; the platform is the arbiter.
