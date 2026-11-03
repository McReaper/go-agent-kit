# platform

What the kit puts on GitHub and in Claude Code so the agent can't change the rules that judge it (criterion T2). The owner applies these; the agent can't.

## The agent's identity: a GitHub App

Repository permissions: Contents and Pull requests read and write; Checks, Actions and Commit statuses read; Metadata read. **Workflows and Administration: no access**, so the agent can't change CI or the protection rules, even through a pull request. No webhook. Installed on the chosen repositories only.

The agent acts as the app through `kit app exec -- <command>`, with `KIT_APP_CLIENT_ID`, `KIT_APP_KEY` (the private key's path, outside every repository) and `KIT_APP_INSTALLATION` set. `kit app installations` lists the installation ids.

## An app's default branch: `ruleset.json`

Apply it once per repository, as the owner, and allow auto-merge and the deletion of merged branches in the repository's settings:

```sh
gh api repos/OWNER/REPO/rulesets --method POST --input platform/ruleset.json
gh api repos/OWNER/REPO --method PATCH -F allow_auto_merge=true -F delete_branch_on_merge=true
```

- a pull request, squash-merged; no review required but the code owners' on the harness, and every approval stays on record;
- the `kit` check passing, reported by GitHub Actions only: it holds the spec, a new dependency, the go line and a library's API until the owner approved the content it has;
- no `update` rule: once both hold, GitHub's auto-merge, which `kit-check.yml` turns on, lands the change.

## The kit's own default branch: `ruleset-kit.json`

The kit judges every app, so its owner merges each change to it: a code owner (`*`) approves the last commit, the `kit` check passes, and only the owner updates the branch, through a pull request.

## Tags and releases: `ruleset-tags.json` and immutable releases

Contents write lets the app push tags and publish releases: a `v*` tag runs an app's release workflow, and a published release's assets are what its users download. Apply both, as the owner:

```sh
gh api repos/OWNER/REPO/rulesets --method POST --input platform/ruleset-tags.json
gh api repos/OWNER/REPO/immutable-releases --method PUT
```

- only the owner creates, moves or deletes a tag (the admin role bypasses; it pushes release tags from its own clone);
- once published, a release's tag and assets can't change, so the app can't swap a binary under a release users already trust.

## Code owners: the template's `.github/CODEOWNERS`

An app started from the template has it; replace `@OWNER` with the owner's login. It names the harness: `.github/`, `.claude/`, `.kit/`, AGENTS.md, CLAUDE.md, the ADRs, `renovate.json` and the glossaries. The spec stays out of it, so that building a scenario, removing its `@wip` tag, needs no one: the kit check holds the spec.

## The merge check: `kit-check.yml`

The app's CI calls `.github/workflows/kit-check.yml` of this repository, pinned by SHA, on `pull_request` (an edit counts when it moved the base), `pull_request_review` and pushes to the default branch. It runs `kit check --ci` with the trees of the commits a code owner approved: a change that is the owner's passes once approved, and stays approved while its files hold the content approved, through later commits and rebases; touching one reopens it. It posts the review sheet, and turns GitHub's auto-merge on for every pull request into the default branch.

## Claude Code: `claude/settings.json`

Deny rules (a force push without a lease, any force push to `main`, `--no-verify`, merging, editing `.kit/`, `.github/`, `.claude/`; `--force-with-lease` stays allowed on the agent's own branches, so it can rebase its pull request) and the hooks: `kit hook post-edit` after each edit, `kit hook stop` before the agent stops. Deploy it as the agent's managed settings, which the agent can't override (see Claude Code's managed settings documentation for the file's place on each system); deny rules are a speed bump, the platform is the arbiter.
