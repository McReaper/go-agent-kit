# The guide

You decide what the app does. The agent decides how, and Rumblestrip checks it. This page is all you need on a pull request.

## Reading a pull request

**Rumblestrip's comment**, at the top, starts with one line:

| It says | It means | You |
| --- | --- | --- |
| ⏳ The checks are running | Rumblestrip is checking the change | wait |
| ✅ Every check passes. **Waiting for you** | Something below is yours to approve | read it, approve |
| ✅ Every check passes. Rumblestrip merges it now | Nothing is yours | nothing |
| ✅ Every check passes. It merges once the pull request below it in the stack has | A build waiting for its spec | nothing |
| ❌ 2 checks fail: Tests, Spec | The agent's code is wrong | nothing: the agent fixes it |
| ✅ Merged into `main` as 051208a | Done | nothing |
| ❌ GitHub refused the merge: … | A rule GitHub enforces said no | read the reason |

Under **For you**, the comment lists what you approve: the spec's scenarios first, then anything else that is yours.

**The label** `human-review` marks the pull requests waiting for you: filter by it.

**The checks**, at the bottom of the conversation: one per check (Tests, Spec, Lint…); **Check**, which passes once they all do; and **Rumblestrip / Owner approval**, which waits for you. The run's summary lists every check with what it checks, then what each failing one found.

## Approving

Read the scenarios the comment lists, in the **Files** tab. If they say what you want, approve the pull request (Review changes → Approve); Rumblestrip checks again and merges it. If not, tell the agent what to change: the changed scenarios need your approval again.

Your approval holds what you read. A file of yours changed afterwards needs it again; the rest doesn't.

Besides the spec, these are yours:

- a new dependency, or a change to Go's version;
- the agent's instructions and the CI: `AGENTS.md`, `CLAUDE.md`, `.github/`, `.claude/`, `.rumblestrip/`, the ADRs;
- an incompatible change to a library's API (`pkg/`);
- a change of more than 400 lines of code.

Everything else, the tests included, the checks judge.

## When something is stuck

- **A check stays red** after the agent's push: open Check's summary. It says which check, what it wants, and what it found: that's what to tell the agent.
- **"A job of Rumblestrip did not finish"**: GitHub's runner failed, not the change. Re-run the failed jobs.
- **"No App to merge as"**: the repository gives Rumblestrip no App to merge with. Run `rumblestrip init github` again with the App's variables.
- **"The merge is still pending"**: re-run the Merge job.
- **"Not in a stack"**: the agent didn't stack the pull request on the one below it. Tell it.
