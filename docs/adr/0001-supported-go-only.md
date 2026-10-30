# 1. Build on a supported Go release, its latest patch

## Context

On go1.25.5 the trial called 29 known vulnerabilities, almost all in the standard library; on go1.25.14, the last 1.25 patch, still 11: Go 1.25 is out of support, and fixes land in the two newest minors only. `go get pkg@latest` also raised the `go` line twice while building the kit, silently switching the toolchain.

## Decision

- The `go` line is the newest Go the pinned golangci-lint can analyze (it must be built with a Go at least as new); the `toolchain` line is that release's latest patch.
- The release must be one Go supports. govulncheck fails the build on any vulnerability the code calls.
- Moving either line is a protected change (`kit guard`, `protected-gomod`): it needs the owner's approval.

## Consequences

A Go release reaching end of life fails the gate through govulncheck before it ships; moving to the next one is a reviewed change, together with golangci-lint.
