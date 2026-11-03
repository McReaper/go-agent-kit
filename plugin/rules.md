This repository is built on the anti-drift kit (github.com/McReaper/go-agent-kit): its checks judge every change, and only the owner can change them.

Commands (`kit` runs the version the repository pins):
- `kit check`: every gate the merge check runs, the same as CI. Run it before you stop.
- `kit spec`: each built scenario of `features/` passes alone, and fails once a value its `Then` steps expect is altered.
- `kit mutate`: the tests must kill the mutants of the lines you changed.
- `kit review`: the sheet the owner reads: the spec's changes and what else is theirs to approve.

Rules no check explains:
- The owner reads the spec, `features/*.feature`, and nothing else. The spec comes first, in its own pull request; the feature is built on top. Load the `kit:new-feature` skill for any feature, the `kit:new-module` skill for a new module, `kit:secrets` for a secret.
- A scenario you believe is wrong: never bend the code or the steps to it; change it on the spec branch.
- Code only tests reach: delete it with its tests.
- Never silence a check (`//nolint`, `t.Skip`, a build tag): fix the code, or stop and ask.
- A pull request merges by itself once its checks pass: never merge, never push to the default branch.
