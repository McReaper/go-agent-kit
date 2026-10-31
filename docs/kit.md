# Anti-drift kit — maintainable, evolvable AI-written Go

The Go counterpart of the .NET anti-drift kit. Target: **Go CLIs, MCP servers and libraries**.

Every tool and setting this document names was run on the template (`template/`, the app apps start from with `gonew`): it passes all of them with no finding, and each case under `trial/cases/<case>` holds one deliberate violation and the exact findings it must produce (`want.txt`). `go test ./...` in this repository reruns them all; a tool upgrade that stops reporting a rule fails it. A row marked *milestone 2* is decided but not yet built or proven.

## Goal and method

Read far less of the code AI produces, and only what matters: quality is carried by the project's structure and checked by machines.

1. Step 1 — the criteria for maintainability, evolvability and security, and what threatens them when AI writes and tests everything. Unchanged from the .NET kit.
2. Step 2 — for each criterion, the tactical solutions in Go, each proven on the trial project. Done, coverage check included.
3. Step 3 — assemble the kit: a reusable base and the rules of each application. Plan below (milestone 2).

## What changes when AI writes and tests everything

Five constraints make the criteria harder to hold than with a human team; every solution in step 2 must withstand them.

- **C1 — Volume outruns review.** A requirement no machine checks is not checked.
- **C2 — The agent optimises for green.** It bypasses any rule it can change: tests, thresholds, warning suppressions.
- **C3 — No memory between sessions.** Conventions, decisions and domain vocabulary must live in the repository.
- **C4 — Limited context.** The agent sees part of the code and reinvents what it doesn't find.
- **C5 — Slow drift.** Each change looks acceptable; their sum degrades the code.

Go note: C2 has Go-specific doors (a custom build tag, a `_dir`, a `TestMain`), listed under *Go traps* below; C4 is eased by Go's small language and explicit imports, nothing else changes.

## Maintainability criteria

Maintainable = understanding, fixing and checking the existing code is cheap. Eleven criteria, each with its typical drift under AI, here in Go terms.

- **M1 — Local simplicity.** Short, shallow functions and files, bounded complexity. *Drift:* long functions, stacked conditions, defensive code no one needs.
- **M2 — Readability.** Explicit names, visible intent, no gratuitous indirection. *Drift:* decorative abstractions, comments that paraphrase the code, `notes.NotesService`.
- **M3 — Consistency.** One problem, one solution everywhere: errors, logs, validation, data access, format. *Drift:* each session brings its pattern or its library (logrus beside slog, testify beside cmp).
- **M4 — One home per business rule.** *Drift:* reimplementation for not having found the existing one.
- **M5 — Minimality.** No dead code, no speculative option, no abstraction without a second use. *Drift:* over-engineering, refactoring leftovers, an exported function nothing calls.
- **M6 — Verifiability.** Behaviour is pinned by tests that really fail when it breaks. *Drift:* hollow tests, written to pass, or changed to pass.
- **M7 — Diagnosability.** In production, why it breaks is known. *Drift:* `_ = err`, `fmt.Errorf("%v", err)` that loses the chain, an error logged at every level.
- **M8 — Dependency hygiene.** Few modules, chosen, current and safe. *Drift:* `go get pkg@latest` that silently raises the `go` line, versions remembered from training, invented module paths.
- **M9 — Traced intent.** The why of decisions (ADR, glossary) is in the repository. *Drift:* decisions forgotten between sessions, then undone.
- **M10 — Testability.** Code tests without infrastructure: dependencies passed in, no hidden clock, randomness or package state. *Drift:* `time.Now()` in rules, `http.DefaultClient`, package-level variables and `init()`.
- **M11 — Technical robustness.** Errors, nil, contexts, goroutines and resources are right. *Drift:* unchecked errors and type assertions, `go func()` nobody waits for, `context.Background()` deep in the code, bodies never closed.

## Evolvability criteria

Evolvable = adding or changing a feature stays cheap as the code grows.

- **E1 — Explicit boundaries.** Delimited modules, dependency direction enforced, a domain without I/O. *Drift:* shortcuts across layers to get to green.
- **E2 — Locality of change.** What changes together lives together. *Drift:* diffs spread out, shared code changed for one use.
- **E3 — Additive extension.** A new feature mostly adds code. *Drift:* editing existing code by convenience, hence regressions.
- **E4 — Explicit domain model.** Domain concepts are types with a shared vocabulary. *Drift:* `string` ids everywhere, rules scattered in handlers.
- **E5 — Stable contracts.** MCP tools, CLI flags, exported library APIs evolve without breaking their consumers. *Drift:* a tool argument or a function signature changed in passing.
- **E6 — Refactorability.** Restructure without rewriting tests, because they test behaviour. *Drift:* mock-based tests that freeze the design.
- **E7 — Predictable structure.** Where a feature goes, and what to reuse, is known without reading everything. *Drift:* a `utils` package, a new location invented per session.

## Cross-cutting criteria and out of scope

- **T1 — No drift.** No criterion degrades as the project grows (answer to C5).
- **T2 — Guardrails out of reach.** The agent can't change the rules, thresholds and tests that judge it (answer to C2).
- **T3 — Sustainable review.** What's left to review per PR is small and targeted (answer to C1).

Out of scope for now: performance, accessibility.

## Security criteria

Eleven criteria besides M8. The typical drift: the agent weakens a protection to pass a test or get unstuck.

- **S1 — No secret in code.** *Drift:* a key "to test", copied into a fixture.
- **S2 — Untrusted input under control.** Validation, parameterised queries, paths, outbound URLs, archives, subprocesses. *Drift:* SQL built with `+`, a path joined from input, `exec.Command(userInput)`.
- **S3 — Explicit authorisation.** For an MCP server: every tool declares what it does to its environment. *Drift:* a side-effecting tool registered like a read-only one.
- **S4 — Standard authentication.** *Out of target* (no network authentication in a local CLI or stdio MCP server).
- **S5 — Safe crypto and transport.** *Drift:* `InsecureSkipVerify: true`, `math/rand` for a token, MD5, a `godebug` line that re-enables an insecure default.
- **S6 — Sensitive data contained.** Never in logs or errors. *Drift:* a whole request logged.
- **S7 — Safe configuration by default.** *Drift:* `net/http/pprof` imported in a shipped binary. Its HTTP part (CORS, headers) is *out of target*.
- **S8 — Abuse resistance.** Bounded sizes, lists and decompression. *Drift:* a list with no limit, `io.Copy` from a gzip stream.
- **S9 — Least-privilege infrastructure.** *Out of target.*
- **S10 — Supply chain.** Locked versions, pinned actions, SBOM. *Drift:* a workflow action on a moving tag, a module of unknown origin.
- **S11 — Audit trail.** *Drift:* a new destructive tool with no trace.

## What Go changes

- **Trivial in Go** (the language or toolchain holds it): E1's three projects per module (nested `internal/` makes another module's domain a compile error), DI lifetimes (no container: wiring is plain code the compiler checks), reflection-hidden dependencies (no dynamic loading: imports are complete), formatting (gofmt), lock files (`go.sum` and the checksum database), ReDoS (RE2 is linear), polymorphic deserialisation (none in `encoding/json`), native fuzzing, unused imports and variables (compile errors).
- **Different, not harder: rules don't live in the compiler.** Go has no warnings: a problem is a compile error or a vet/analyzer finding, and `go test` already runs vet's high-confidence checks. The one place .NET needed the compiler (E1) is still a compile error in Go. Everything else arrives seconds later, through a hook that lints the edited package, and CI arbitrates either way.
- **Harder in Go:** mutation tools have less (no per-test coverage analysis nor mutant schemata, so each mutant is a recompile; gremlins does not run on Windows); visibility is per package, nothing like `sealed`; no nullability (nilaway fills part of it); test names are identifiers (sentences go in `t.Run`); doc comments are Go culture (the comment rule allows them on contracts); golangci-lint has no config inheritance (the kit's config is applied whole); `if err != nil` inflates every M1 metric; foreign JSON in the app's own code fights the snake_case rule.
- **Go traps with no .NET row** (each proven in `guard-bypass`, `guard-protected`, `m11-robustness`, `t2-gomod`): a custom `//go:build` tag hides a file from every lint and test; `_x`, `.x` and `testdata/` directories are never built; a `*_windows.go` file is unlinted on Linux CI unless lint runs per GOOS; a `// Code generated … DO NOT EDIT.` header turns linters off; `//go:linkname`, `//go:debug` and `import "C"` reach past the checks; a new `TestMain` can skip a package's tests; a `godebug` line in go.mod re-enables insecure defaults; `go get pkg@latest` silently raises the `go` line (it happened twice while building this kit).

## Target architecture

```
cmd/<app>/                         wiring only
internal/<m>/                      a module's contract: the only package other modules import
internal/<m>/internal/domain/      pure: the standard library and pkg/ only, no I/O, no clock, no log
internal/<m>/internal/<feature>/   a vertical slice: Run(ctx, Args) (Out, error) and its tool and command
internal/<m>/internal/infra/       I/O adapters, the only place besides platform that starts processes
internal/platform/                 the shared kernel (MCP/CLI binding, paths); knows no module
pkg/<lib>/                         libraries
```

The template is built this way and compiles: the contract wires its slices and infra, each slice declares the interface it needs. One `Args` struct per operation: its `json` tags are the MCP arguments, the CLI flags are derived from them (`template/internal/platform/op`), so the two surfaces can't drift.

## Step 2 — Tactical solutions

**Blocking** = a CI gate; **partial** = a gate on part of it, the rest guided; **guided** = reference example, skill or template.

| Criterion | Level | Left to review |
| --- | --- | --- |
| T2 — Guardrails out of reach | Blocking | The rare rule changes |
| M6 — Verifiability | Blocking | Test scenarios and properties, not their code; equivalent mutants |
| S1 — No secret | Blocking | Low-entropy secrets |
| M10 — Testability | Blocking | Nothing |
| M11 — Robustness | Blocking | Concurrency the tests don't exercise |
| M1 — Local simplicity | Blocking | Nothing |
| E1 — Boundaries | Blocking | Nothing |
| T1 — No drift | Blocking | Nothing |
| T3 — Sustainable review | Blocking | n/a |
| E5 — Stable contracts | Blocking | Intended contract changes |
| M8 — Dependencies | Blocking | Every new module |
| S10 — Supply chain | Blocking | Nothing beyond M8 |
| S5 — Crypto and transport | Blocking | Nothing |
| S2 — Untrusted input | Partial | Business validation; a program name taken from input |
| S3 — Authorisation (MCP) | Partial | Which effect a tool declares |
| S7 — Safe defaults | Partial | Nothing in target |
| S8 — Abuse resistance | Partial | The limits chosen |
| S6 — Sensitive data | Partial | What new data gets logged |
| M3 — Consistency | Partial | Every new pattern |
| M5 — Minimality | Partial | Speculative abstractions |
| M4 — One home per rule | Partial | A rule rewritten in another form |
| M7 — Diagnosability | Partial | Whether an event deserves a log |
| E2 — Locality | Partial | Diffs that leave their module |
| E6 — Refactorability | Partial | Nothing |
| M2 — Readability | Partial | Names, by sample |
| E4 — Domain model | Partial | The model itself |
| S11 — Audit | Guided | Which actions are journaled |
| E3 — Additive extension | Guided | Changes to shared code |
| E7 — Predictable structure | Partial | Where new features go |
| M9 — Traced intent | Guided | ADRs and `// WHY:` comments |
| S4, S9 | Out of target | — |

The rules ship in this repository: `lint/golangci.yml` (golangci-lint 2.14), `cmd/kitvet` (the kit's nine analyzers), `cmd/kit guard`, `archtest`, and `tools.txt` (every tool's pinned version). Apps never carry a lint config of their own.

### T2 — Guardrails out of reach

The workstation gives fast feedback but can be bypassed; the Git platform is the arbiter, and the agent never holds the rights to change it.

| Solution | Tool | What it blocks | Proof |
| --- | --- | --- | --- |
| A separate identity for the agent: push to its branches, no merge, no admin | A GitHub App with contents and pull requests only, no workflows or administration permission; `kit app exec -- <command>` mints its hour-long installation token into the command's environment, never printed, and commits as `<app>[bot]` | The agent merging alone, editing protection rules, or changing a workflow | `internal/platform/ghapp` tests; the live round is milestone 2 |
| One command, the same on the agent's machine and in CI | `kit check`: golangci-lint with the kit's embedded config and the kit's analyzers per GOOS, archtest, sweep, deadcode, nilaway, govulncheck, the tests (with `-race` on Linux), the guard, and the files no Go gate reads: actionlint and zizmor on the workflows, shellcheck on the scripts (`.sh`, or a shell shebang for git hooks), gitleaks on the change's commits and on what isn't committed yet; an app's own lint config is never read (crash tests ran a local check that missed a gate CI ran) | A local pass that CI then fails; a config weakened in the app; a workflow, script or secret an app's CI never checked | `TestKitCheckRunsEveryGateAndFailsOnAny`, `TestKitCheckFailsOnAWorkflowAScriptAndASecret` |
| A secret stopped before it enters the history | The template's `.githooks/pre-commit`: `kit guard`, then gitleaks on the staged lines; enabled per clone with `git config core.hooksPath .githooks` | A token committed, which only a history rewrite removes | `TestThePreCommitHookStopsAStagedSecret` |
| Protected paths need the owner's approval | `platform/CODEOWNERS` (tests, testdata, `go.mod`, `.github/`, `.kit/`, `.claude/`, AGENTS.md, ADRs) and `platform/ruleset.json`: a code owner approves the last commit, stale approvals dismissed, the `kit` check required, and only the owner updates the branch, through a pull request, so the app can't merge even an approved change. The app's CI calls `kit-check.yml`, pinned by SHA | Changes to config, thresholds, CI, hooks, `.claude/`, `go.mod`, existing tests; the agent merging | `TestTheRulesetLetsOnlyTheOwnerMergeAnApprovedCheckedChange`, `TestTheAppCodeOwnersProtectWhatJudgesTheAgent`; applying them is the owner's |
| Rules shipped by the kit, pinned by version | `lint/golangci.yml`, `cmd/kitvet`, `tools.txt`; an app's own `.golangci.yml` is never read | A rule weakened in the app repository | all cases |
| Lint and vet per target system | CI runs golangci-lint and kitvet with `GOOS=linux` and `GOOS=windows` | A file built for one system escaping the checks | `m11-robustness` (`clean_windows.go`) |
| Diff scan for bypasses | `kit guard`: `//nolint`, `#nosec`, `//lint:ignore`, `//go:linkname`, `//go:debug`, `import "C"`, unknown build tags, `_`/`.` directories, a generated header outside `gen/`, `t.Skip`, `testing.Short` | Silencing a finding or hiding a file; a normal change gets nothing | `guard-bypass`, `guard-clean` |
| The guard reads the working tree, not only commits: edits and files git doesn't track yet | `kit guard` against base's merge base | A local check passing before the commit that CI then rejects | `guard-uncommitted` |
| Tests: new test files are free, any change to an existing one is protected | `kit guard`: a modified, renamed or deleted `_test.go` or `testdata/` file, a new `TestMain` or `init` in a test. An addition counts too: an added line can cancel an assertion (a reassignment, an early return). Locally they are reported for the owner, not failures: a guard that failed on them made an agent bend the code to fit a wrong test rather than ask (`r2-envtest`). The merge check (`--ci`) holds them until a code owner approved the last commit (`--approved`); a weakened assertion also shows as surviving mutants | A test weakened or removed to pass, without an incentive to bend the code instead | `guard-protected` |
| The owner reads what matters, not the diff | `kit review`: the review sheet, posted on the pull request by `kit-check.yml` and updated on each push: the changes that need the owner's approval, test scenarios added, removed or changed, read as sentences (TestAddReportsTheStoreFailure reads "Add reports the store failure", a subtest follows a colon), golden snapshots changed, a library's API changes (apidiff against the base), new modules, `go.mod` lines added, shared code (platform, `pkg/`, `cmd/`, a module's contract), rules (CI, the agent's settings and instructions, ADRs) and the `// WHY:` comments added; labels to match | A protected change, a removed scenario or a contract change lost in a large diff | `t2-review-sheet`, `TestKitReviewWritesTheSheetForThePullRequest` |
| Nothing loaded or reached past the type system | depguard denies `unsafe`, `plugin`; `reflect` in the domain; gomoddirectives forbids `replace` to a local path and `godebug` | Dependencies hidden from the checks | `t2-directives`, `t2-gomod` |
| Agent-side guardrails | `platform/claude/settings.json`, deployed as managed settings: deny rules (force push, `--no-verify`, `gh pr merge`, editing `.kit/`, `.github/`, `.claude/`), `kit hook post-edit` (each edit's package linted, findings back to the agent through exit code 2) and `kit hook stop` (the full `kit check` before the agent stops, once) | Writing a protected path, `--no-verify`, `push --force`: a speed bump, not a guarantee; rules learned at the end instead of at each edit | `TestTheAgentSettingsDenyBypassesAndRunTheKitHooks`, `TestThePostEditHookTellsTheAgentWhatItsEditBroke`, `TestTheStopHookLetsTheAgentStopOnItsSecondTry` |

Left to review: PRs touching a protected path. A PR runs its own workflow version: CODEOWNERS on `.github/` is what stops a weakened CI from merging.

### M6 — Verifiability

Not whether code is covered, but whether tests fail when the code breaks.

| Solution | Tool | What it blocks | Proof |
| --- | --- | --- | --- |
| Mutation on changed code | `kit mutate`: go-mutesting (statement, branch, expression removal and operator swaps) on a copy of the module, since it rewrites the source in place while tests run; only mutants starting on a line the change adds count, and not those that only drop the return of an `if err != nil` block; fails under `--min`, 0.8. `kit-check.yml` runs it on every pull request | New code that can be broken without any test failing | `m6-weak-test`, `TestKitMutateFailsAChangeItsTestsDontPin` |
| Tests must assert something | `kitvet testassert`: a test whose every failure sits under `err != nil` or a nil check | A test that only checks that the code ran | `e6-tests` |
| No mocks | depguard denies `golang/mock`, `go.uber.org/mock`, testify; fakes implement the interface | Tests that freeze calls instead of behaviour | mechanism in `m3-consistency` |
| Properties and fuzzing | `pgregory.net/rapid`, `testing.F` | Code that only passes the planned cases | `template` (`slug`, `domain`) |
| Golden files | stdlib golden pattern with `-update`, protected by the guard | Partial assertions on large outputs | `template` (`tools.golden.json`), `guard-protected` |
| Tests first | The template's `new-feature` skill: the contract, then tests that must fail first, then the slice; the owner may ask to see the tests alone | Tests written to fit the code | guided (Pass 4) |

No coverage gate: mutation already counts uncovered code. A mutant counts on the line it starts: removing a block whose first line the change left alone doesn't count.

The threshold comes from the template. Its tests kill 76 of the 81 mutants that count; the survivors are equivalent: `return -2` where `strings.Map` drops any negative, `> MaxASCII` where DEL is neither letter nor digit, `iota + 2` for an enum whose values mean nothing, a sort comparison the store's file order never reaches. So 0.8 leaves room for one equivalent mutant in five. Go adds a kind .NET doesn't have: 29 more mutants only drop the return of an `if err != nil` block, and live unless a test makes a disk or system call fail. Counted, they would take the infra from 91% to 55%; the compiler (an unused `err`) and errcheck already refuse a dropped error, so `kit mutate` lists them apart (ADR 4). A refusal (`errors.Is(err, fs.ErrExist)`, a sentinel error) is a behaviour and counts.

### Pass 1 — Linters

`lint/golangci.yml`, plus `cmd/kitvet` for rules no linter has. `issues.uniq-by-line: false`: golangci-lint's default keeps one finding per line, and an agent that fixes one then meets the next.

| Criterion | Rules and tools | What they block | Proof |
| --- | --- | --- | --- |
| M1 | gocognit ≤ 10; revive function-length ≤ 30 lines, argument-limit ≤ 4, file-length-limit ≤ 300; `kitvet nesting` ≤ 2 (every control statement counts: revive's rule counts only `if`) | Long, deep or tangled functions | `m1-simplicity` |
| M2 | `kitvet comments` (`// WHY:` only; doc comments on a contract's exported API and a command's package doc; a `Copyright` or `SPDX-License-Identifier:` header above the package clause), staticcheck ST*, misspell, predeclared, `kitvet pkgname` (stutter), gofumpt and goimports | Paraphrasing comments, mixed naming | `m2-readability`, `e7-names`, `fmt-formatting` |
| M3 | depguard one-library-per-need (slog, stdlib errors, go-cmp, go-sdk, cobra), forbidigo `fmt.Print*`, modernize, tagliatelle (snake_case json; another program's format is decoded in its own casing, in infra or the platform) | Two ways to do one thing | `m3-consistency` |
| M5 | unused, unparam, wastedassign, unconvert, interfacebloat ≤ 5 | Dead code, unused parameters | `m5-minimality` |
| M5 | `kit sweep`, over the whole module: code only tests reach, methods included (deadcode never reports a method, which let a crash test leave one alive), and names a module's inside exports though only their package uses them | Code kept alive by its tests; a package's inside leaking out | `m5-only-tests` |
| M10 | depguard `pure-domain` (no `os`, `io/fs`, `net`, `os/exec`, `database/sql`, `log/slog`, `reflect`, randomness), forbidigo `time.Now/Since/Until` and `os.Getenv` in the domain, gochecknoglobals, gochecknoinits | Hidden clock, environment, I/O, state | `m10-testability` |
| M11 | errcheck (type assertions, blank), govet (all but fieldalignment, shadow), forcetypeassert, nilerr, nilnil, nilnesserr, containedctx, fatcontext, bodyclose, noctx, sqlclosecheck, rowserrcheck, exhaustive, makezero, durationcheck, reassign, musttag; `kitvet goroutine` (`go` only in platform), `kitvet ctxroot` (no new root context outside main) | Unchecked errors, leaked resources, orphan goroutines | `m11-robustness` |
| M11 | nilaway | A nil that flows to a dereference | `m11-nil` |
| M11 | `go test -race` (Linux CI: it needs cgo) | Data races the tests exercise | `m11-race` |
| S2 | gosec G107, G110, G204 (outside infra and platform), G304, G305; `kitvet sqlconst` (the query of every `database/sql` call is a constant: gosec's G201/G202 miss quoted values and every `*Context` method) | SQL injection, path traversal, zip slip, decompression bombs | `s-security` |
| S5 | gosec G401, G402, G404, G501; forbidigo `InsecureSkipVerify` | Weak crypto, TLS off | `s-security` |
| S7 | gosec G108 | pprof exposed | `s-security` |
| T2 | bidichk, asciicheck, gocheckcompilerdirectives | Trojan Source, homoglyph names, directives disabled by a space | `s-security`, `t2-directives` |

### Pass 2 — Architecture tests

`archtest.Test(t, dir)`, called from one test of the app; the compiler already keeps other modules out of a module's `internal/`.

| Criterion | Rule | What it blocks | Proof |
| --- | --- | --- | --- |
| E1 | The domain imports the standard library and `pkg/` only (third-party included) | A framework in the domain | `e1-boundaries` |
| E2 | A slice imports no other slice, and no infra (the contract passes it in) | Coupled slices, slices that need I/O to test | `e1-boundaries` |
| E1 | Platform imports no module | Shared code coupled to one module | `e1-boundaries` |
| E7 | Every package sits in the layout | A package invented outside it | `e1-boundaries` |
| E4 | `kitvet typedid`: no id of a bare string or integer in the domain; `kitvet domainconv`: a domain type is built by its Parse function, never by conversion outside the domain (a crash test validated ids in one session and converted them in the next) | Ids of different things swapped; unchecked values in the model | `e4-typed-ids` |
| E6 | testpackage (black-box `_test` packages, `export_test.go` and `internal_test.go` included, which it skips by default), no mocks | Tests tied to internals | `e6-tests` |
| S3 | An operation's effect (`ReadOnly`, `Additive`, `Destructive`) has no default: registering one without it fails; it becomes the MCP tool's annotations | A side-effecting tool advertised as harmless | `template` (`op_test`, `tools.golden.json`) |
| S8 | `kitvet boundedlist`: a slice whose `Out` holds a list (a slice or a map) has an integer `limit` in its `Args` (a crash test without the kit shipped a search returning every match) | A call that returns every row | `s8-unbounded-list` |

### Pass 3 — CI checks

| Criterion | Rule and tool | What it blocks | Proof |
| --- | --- | --- | --- |
| S1 | gitleaks on the tree and the history | A secret committed, even removed later | `s1-secrets` |
| S1 | `kit guard`: env files, keys and certificates (`.env`, `.pem`, `.key`, `.pfx`, `.p12`, `.jks`, `id_ed25519`…) refused; `.env.example` and public keys allowed | A secret file committed whatever its content | `guard-sensitive` |
| M8, S10 | govulncheck: only vulnerabilities the code calls | A known vulnerability shipped | `m8-vulnerable-dep` |
| M8 | `go mod tidy -diff`, gomoddirectives, `kit guard` on the `go`/`toolchain` lines | Untidy or silently moved modules | `m8-untidy`, `t2-gomod`, `guard-gomod` |
| M8 | The `go` line follows the newest Go the pinned golangci-lint supports, the `toolchain` line its latest patch, on a supported release | Building on an unsupported Go | see below |
| S10 | zizmor and actionlint on workflows: pinned actions, no `pull_request_target` with PR code, no template injection, least permissions | A workflow that leaks or runs untrusted code | `ci-workflow` |
| S10 | cyclonedx-gomod: an SBOM of what the binary links | An unknown shipped dependency | `TestSBOMListsTheModulesTheBinaryLinks` |
| E5 | MCP tools: a golden snapshot of `tools/list` (names, schemas, annotations), protected by the guard; libraries: apidiff against the base | A contract changed in passing | `template` (`tools.golden.json`), `e5-api-break` |
| M5 | deadcode from `cmd/`: functions only tests reach | Code nothing ships | `m5-deadcode` |
| M4 | dupl ≥ 100 tokens | Copied rules | `m4-duplication` |
| T1 | Absolute: zero findings everywhere, every gate | Slow degradation | `template` |
| T3 | `kit guard`: ≤ 400 changed production lines, one module per PR | PRs too big to review | `guard-size` |
| Scripts | shellcheck | Word splitting, globbing | `sh-script` |
| Scripts | PSScriptAnalyzer at warning severity, on `.ps1`, `.psm1` and `.psd1`, in pwsh or Windows PowerShell | Aliases, `Write-Host`, `Invoke-Expression`, a password taken as plain text | `ps-script` |
| M8, S10 | Renovate, the template's `renovate.json`: an update waits 7 days (`minimumReleaseAge`), `go mod tidy` runs after it, the `go` and `toolchain` lines stay the owner's, actions are pinned by digest, the kit's two pins move together. `kit check` runs Renovate's own validator, `--strict` | A misspelled option, which Renovate's server ignores without a word to the repository; a dependency released yesterday | `m8-renovate`, `TestRenovatesValidatorFindsNothingInTheTemplate` |

The Go version rule comes from the trial: on go1.25.5, govulncheck found 29 vulnerabilities the code calls; on go1.25.14, the last 1.25 patch, still 11, because Go 1.25 is out of support and fixes land in 1.26 and 1.27 only. The trial moved to go 1.26.0 with toolchain go1.26.10, and x/text and x/sys took their fixes: none left.

### M7 — Diagnosability

| Solution | Tool | What it blocks | Proof |
| --- | --- | --- | --- |
| One logger | depguard (`log`, logrus, zap, zerolog denied) | Mixed log formats | `m3-consistency` |
| Structured logs | sloglint: attributes only, static messages, snake_case keys, no global logger, context variants | `slog.Info("…", "k", v)` with no context | `m7-diagnosability` |
| Errors keep their chain | errorlint (`%w`, `errors.Is`) | `err == ErrX`, `%v` on an error | `m7-diagnosability` |
| Logged once | `kitvet logreturn`: an error logged and returned in one block | The same failure logged at every level | `m7-diagnosability` |
| One message, one place | `kit sweep`: a log message written in two places | A log line that can't say which code ran (.NET's unique `EventId`) | `m7-duplicate-messages` |
| No panic, exit or recover in the code | forbidigo: `panic` and `recover` in platform only, `os.Exit` in main only | Crashes and swallowed panics | `m7-diagnosability` |
| No log in the domain | depguard `pure-domain` | Logic and diagnostics mixed | `m10-testability` |

### Pass 4 — Guided: the template

An app starts as a copy of the template: `gonew github.com/McReaper/go-agent-kit/template@<version> example.com/app`. It is the code every proof runs on, so what an app starts from passes every gate. Rules stay in the gates; these show and explain them.

| Criterion | What the template holds | Its test |
| --- | --- | --- |
| E3, E7 | The reference slices (`add`, `list`), their module's contract, domain and infra, the platform; AGENTS.md says where code goes, and a test fails on any path it names that doesn't exist | `template/agents_test.go` |
| E4 | A glossary per module: every exported domain type defined; a test fails on one missing | `template/internal/notes/glossary_test.go` |
| S11 | The op binding journals every run of an operation that isn't read-only, one JSON line in the data directory: the operation, its arguments, its error; binding an operation without a journal is refused | `template/internal/platform/op`, `…/journal` tests |
| S8 | A list takes a bounded limit with a default (`list_notes`: 1 to 100, 20 when not given); `kitvet boundedlist` holds every slice to it | `template/internal/notes/internal/list` tests |
| S6 | The `secrets` skill: a `secret.Secret` type that reads `[redacted]` through fmt, slog, errors, JSON and the journal; its code is extracted from the skill and run through every gate | `TestTheSecretsSkillsCodePassesEveryGate` |
| M10 | Configuration read once at startup, checked, with a message naming the setting (`NOTES_DATA_DIR` must be absolute) | `template/internal/platform/config` tests |
| M2 | An optional argument is a pointer: nil when not given, a CLI flag left out stays nil, so "not given" differs from zero | `template/internal/platform/op` tests |
| M6, C2 | The `new-feature` skill: the contract first, tests as sentences that must fail first, the slice, the golden diff, `kit check`, `kit mutate`, `kit review`; a wrong test is fixed and reported, never bent around. The `new-module` skill: when a module, its skeleton in its own pull request | AGENTS.md's skill test |
| T2 | `.github/CODEOWNERS` and the CI that calls `kit-check.yml` pinned by SHA; the pre-commit hook | `TestTheAppCodeOwnersProtectWhatJudgesTheAgent`, `TestThePreCommitHookStopsAStagedSecret` |

## Coverage check

**Criterion → solution.** Every criterion points to at least one proven solution, or to milestone 2, or is out of target (S4, S9, S7's HTTP part).

**File type → gate.**

| Files | What checks them | Status |
| --- | --- | --- |
| Go production code | golangci-lint, kitvet, archtest, nilaway, deadcode, govulncheck, mutation, per GOOS | Proven |
| Tests, golden files, fuzz corpus | testassert, testpackage, thelper, usetesting; the guard's protected-test rules | Proven |
| Generated code | `exclusions.generated: disable`; the guard allows the header in `gen/`, `*_gen.go`, `*.pb.go` only | Proven |
| `go.mod`, `go.sum` | gomoddirectives, `go mod tidy -diff`, govulncheck, the guard's go-line rule; CODEOWNERS | Proven |
| Workflows | actionlint, zizmor, in `kit check` | Proven |
| Shell scripts, git hooks | shellcheck, in `kit check` | Proven |
| Secrets in any file | gitleaks in `kit check` and the pre-commit hook; the guard's sensitive files | Proven |
| The kit's config in an app (`.golangci.yml`, `tools.txt`, hooks, `.claude/`) | Never read from the app; CODEOWNERS, managed settings | Proven; managed settings are the owner's to deploy |
| Markdown (AGENTS.md, glossaries, skills) | AGENTS.md's paths and skills exist; every domain type is in its glossary; a skill's code passes every gate | Proven (template) |
| ADRs | CODEOWNERS; the review sheet lists them | Proven |
| PowerShell | PSScriptAnalyzer, in `kit check` | Proven |
| `renovate.json` | Renovate's validator, in `kit check`; CODEOWNERS | Proven |
| Dockerfiles, IaC | hadolint, Checkov when they appear | Out of target until then |

**What stays uncovered**, each a review point or a future rule:

- Semantic duplication (M4): dupl sees tokens, not a rule rewritten another way.
- Equivalent mutants (M6): judged by a human, which is why the gate is a score.
- Races in paths no test runs (M11): the race detector sees only what executes.
- A subprocess whose program name comes from input (S2): gosec G204 reports variable arguments, not a variable program (pinned in `s-security`).
- Low-entropy secrets (S1): gitleaks misses `"correct horse"`.
- SQL built outside `database/sql` calls, through a helper (S2): `sqlconst` checks the call site only.
- Code bent to fit a wrong test (M6, C2): since changing a test needs approval, an agent may change the code instead (a crash test made a Windows app read a Linux variable to satisfy a test with a Linux assumption). Guard v2 reports a changed test locally instead of failing on it, and in the rerun the agent fixed the test and said why (`r4-envtest`); review catches the rest.
- Whether an event deserves a log (M7), names (M2), speculative abstractions (M5).
- What Renovate's rules do once on its server (S10): the validator proves the config is well-formed, not that a `matchDepTypes` matches; the first update pull requests show it.

## Step 3 — Assembly (milestone 2)

The kit's repository is itself the arbiter: everything in it goes through the owner's full review, like the reference slice. Every .NET tactical solution is built first, the ones step 2 marks *milestone 2* included, and trimmed later by use. One PR per package, each usable alone:

1. **Guard v2.** Locally, protected changes are reported, not failures: a guard that failed on them made an agent bend the code to fit a wrong test rather than ask (crash test `r2-envtest`). Bypasses stay blocking. In CI, protected changes fail until a code owner approved the last commit. Sensitive files (`.env`, keys, certificates) are refused.
2. **Analyzers from the crash tests.** Functions and methods only tests reach (deadcode never reports a method); a domain type built by conversion outside the domain; duplicate log messages; exported names used only in their package.
3. **`kit check`.** One command, the same locally and in CI: lint per GOOS, kitvet, archtest, dead code, nilaway, govulncheck, tests, guard.
4. **`kit mutate`.** go-mutesting on a copy, changed lines only, a threshold calibrated on the reference slice.
5. **`kit review`.** The review sheet posted on the PR: protected changes, test scenarios added or removed, read as sentences, contract changes (golden snapshot, apidiff), new modules, shared code touched, `// WHY:` added; path labels.
6. **Platform.** The bot account; CODEOWNERS and the ruleset; reusable workflows apps pin by SHA; Claude Code hooks, deny rules and managed settings; pre-commit hooks.
7. **Template and reference slice** (`gonew`): a `Secret` type that redacts in logs (S6), bounded list arguments (S8), a journal of every non-read-only operation in the op binding (S11), configuration validated at startup, optional arguments, AGENTS.md and the test of its paths and commands, a glossary per module, the new-feature (tests first) and new-module skills. The owner reads the slice in full; it calibrates the thresholds.
8. **Supply chain.** Renovate, proven by its validator; PSScriptAnalyzer for PowerShell.
9. **Crash tests.** The environment test rerun on guard v2; the same chain of features with and without the kit; a platform round under the bot (merge, weaken the CI, edit a workflow). Each attempt is kept as a `crash/<round>-<name>` branch.
10. **Final review** with the dotnet-agent-harness checklist.

## Principles

Kept from the .NET kit, with Go notes.

- A rule that matters breaks the build; documentation only explains why. *Go:* "the build" is CI and the hook, since analyzers don't run in `go build`.
- Every error message is written for the agent: the rule, why, how to fix. *Go:* every message of kitvet, the guard and archtest names its criterion and the fix; depguard and forbidigo carry the criterion in their message.
- Examples beat instructions: a reference slice in the template.
- Absolute thresholds on every PR's new code; a ratchet (`new-from-merge-base`) only to adopt the kit on legacy code.
- The generic base ships as a pinned, versioned dependency, out of the agent's reach; app-specific rules stay in the app, protected.
- One target: Go CLIs, MCP servers and libraries; same architecture, no profiles.
- Architecture: a modular monolith, one repository per system, a module per bounded context: contract, pure domain, vertical slices and infra, with nested `internal/` as the boundary.
- DRY applies to business rules inside a module's domain, not to the plumbing of slices.
- Comments `//` are forbidden unless they start with `WHY:`; doc comments only on a contract's exported API and a command's package doc.
- T2 and M6 are the foundations.
- Tests first: the agent writes tests, the owner reviews them, they are locked, then implemented in another pass.
- Review moves to specs, properties and architecture rules, and is bounded in size per PR.
- Feedback loop: any defect that passes the gates adds a rule (this kit's own proofs found seven: the module counter, `uniq-by-line`, revive's nesting, gosec's SQL rules, testassert and `TestMain`, `slug.Make`'s nesting, nilaway's stderr).
- Drift: tools first, no detection agent at first.
- A check enters the base only if it is generic, with no app-specific configuration.
- Platform: GitHub. The agent works under a separate identity, without merge or admin rights.

## Existing work studied

- [Harness engineering for coding agent users](https://martinfowler.com/articles/harness-engineering.html) (Birgitta Böckeler): guides and sensors match the three levels.
- [Harness engineering: leveraging Codex in an agent-first world](https://openai.com/en/index/harness-engineering/) (OpenAI): fixed layers per domain, dependency direction checked by custom linters.
- [Modular Monolith with DDD](https://github.com/kgrzybek/modular-monolith-with-ddd): the .NET reference; layers inside modules where this kit puts slices.
- [dotnet-agent-harness](https://github.com/ryan75195/dotnet-agent-harness): checklist for the final review.
