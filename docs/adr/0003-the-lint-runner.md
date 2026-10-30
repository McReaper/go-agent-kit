# 3. golangci-lint plus kitvet, per target system

## Decision

The kit's rules run as two tools: golangci-lint with `lint/golangci.yml`, and `kitvet`, a multichecker of the kit's own go/analysis analyzers. Both run once per target GOOS. `kit lint` (milestone 2) will run them together with the config embedded, so an app never reads a config of its own.

## Why not a custom golangci-lint build with module plugins

It would merge the outputs, but it rebuilds golangci-lint for every kit change and ties the kit to its plugin API; two binaries that each print `file:line: message` are simpler to pin and to prove.
