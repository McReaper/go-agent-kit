This repository is built on Rumblestrip (github.com/McReaper/rumblestrip-go): its checks judge every change, and only the owner can change them.

Commands (`rumblestrip` runs the version the repository pins):
- `rumblestrip check`: every gate the merge check runs, the same as CI. Run it before you stop.
- `rumblestrip spec`: each built scenario of `features/` passes alone, and fails once a value its `Then` steps expect is altered.
- `rumblestrip mutate`: the tests must kill the mutants of the lines you changed.
- `rumblestrip review`: the sheet the owner reads: the spec's changes and what else is theirs to approve.
- `rumblestrip pr create`: opens the current branch's pull request, stacked on its base's when that isn't the default branch.

Rules no check explains:
- The owner reads the spec, `features/*.feature`, and nothing else. The spec comes first, in its own pull request; the feature is built on top. Load the `rumblestrip:new-feature` skill for any feature, the `rumblestrip:new-module` skill for a new module, `rumblestrip:secrets` for a secret.
- A scenario you believe is wrong: never bend the code or the steps to it; change it on the spec branch.
- Code only tests reach: delete it with its tests.
- Never silence a check (`//nolint`, `t.Skip`, a build tag): fix the code, or stop and ask.
- A pull request's title is a conventional commit of 72 characters at most, `feat(notes): a blank title is refused`: it becomes the commit's subject on the default branch, and its description the commit's body.
- A pull request merges by itself once its checks pass: never merge, never push to the default branch.
