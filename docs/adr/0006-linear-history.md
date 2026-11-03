# 6. History stays linear: an app squashes, the kit's owner picks

## Decision

Every repository the kit protects keeps a linear history (`required_linear_history`). An app's ruleset allows squash merges only, and the auto-merge `kit-check.yml` turns on squashes. The kit's own ruleset allows a rebase or a squash: its owner merges by hand and picks. No rule caps a pull request's commits.

## Why

- A squash lands one commit per pull request, the change its `kit` check passed: reverting a change reverts one commit, and `git bisect` never stops on an intermediate commit no check ran on. A rebase merge would land each of the agent's commits, the fixes after a red check included.
- A cap on a pull request's commits would hold the same line only by having the agent rewrite a branch it already pushed after each red check; the squash holds it with no rule to maintain.
- No kit check reads the default branch's history: each compares the pull request with its base. Merge commits are left out for the history's reader alone.
- The kit's pull requests sometimes carry several commits, each worth its own message; its owner keeps them with a rebase merge.
