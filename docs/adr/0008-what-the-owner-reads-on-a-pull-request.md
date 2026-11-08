# 8. What the owner reads on a pull request

## Decision

- Each gate of `kit check` runs in a CI job of its own, all at once, named for the owner. `Check`, the job the ruleset requires, passes once every one did.
- The owner's approval is a commit status, `Kit / Owner approval`, which the review sheet's job posts and the ruleset requires. No gate fails for want of it.
- The kit's comment opens with a status line that each job rewrites as the pull request moves on. The run has one summary, Check's: every check with what it checks, then what each failing gate wants and found. Each finding with a line is an annotation on it.
- The commit the merge job lands is the pull request's title, a conventional commit of 72 characters at most, and its description. The App's bot is its author, the owner its co-author.

## Why

- One job reported `kit / kit / kit — Failing after 3m` for a failed test and for a spec awaiting approval alike. The owner opened the log to learn whose move it was, and the log was written for the agent.
- A check that fails for want of approval reads as broken. A pending status reads as waiting, in the kit's words.
- GitHub shows no text of a workflow's own in the checks list. A commit status shows its description there, and the kit's comment is on the pull request's page.
- The approval status counts only from GitHub Actions, and from the one job allowed to write statuses, which runs none of the change's code. Code under test could read a job's token.
- GitHub counts a skipped check as passed. `Check` never skips, and the condition that skips an edit of the description alone is on the caller's job.
- When it squashes, GitHub adds every author of the branch's commits as co-author, the bot itself included, and keeps the commits' trailers. The merge job writes the message itself.
- Each gate's job installs only that gate's tool. A run is about twenty jobs: on GitHub's free plan, at most twenty run at once, so a stack's two runs queue.
