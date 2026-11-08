# 7. A spec and its build are a stack, and Rumblestrip merges them

## Decision

The build's pull request sits on the spec's branch, and the two are one GitHub stack. Rumblestrip merges a pull request into the default branch itself, once its check passed, through GitHub's asynchronous merge, as a GitHub App whose Client ID and key `rumblestrip init github` stores as the repository's secrets: the agent's, or any App with contents and pull requests write. It turns on no auto-merge, and no job of its own moves pull requests or deletes branches.

## Why

- GitHub's stacks move the next pull request onto the default branch once the one below merged: they rebase it, retarget it, and its check runs again. A pull request outside a stack stays on a merged branch, so the check fails one on another branch that is in no stack.
- GitHub's auto-merge takes no pull request of a stack through the API ("Auto-merge is not supported for stacked pull requests"), even where its web page offers it, and a pull request with auto-merge on can't join a stack. The asynchronous merge is the one GitHub names for a stack's pull request; the ruleset still decides when it runs.
- The repository deletes merged branches; GitHub did so after the asynchronous merge, and kept the stack's next pull request open.
- That holds for a merge by an App. After a merge by the workflow's own token, GitHub deleted no branch, and moved the stack's next pull request onto the default branch without rebasing it or running its check, so it sat in conflict; events of that token start no workflow.
- A check that passes on the owner's approval, run apart, leaves the run that failed for want of it on the commit, and GitHub holds the merge on that one. The approval re-runs the failed run instead, and that run's merge job merges.
- An hourly job that moved stacked pull requests and deleted their bases (rumblestrip v0.2.3) redid by hand what stacks do; it went.
