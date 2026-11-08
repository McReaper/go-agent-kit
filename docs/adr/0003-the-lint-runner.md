# 3. golangci-lint plus rumblestrip-vet, per target system

## Decision

Rumblestrip's rules run as two tools: golangci-lint with `lint/golangci.yml`, and `rumblestrip-vet`, a multichecker of Rumblestrip's own go/analysis analyzers. Both run once per target GOOS. `rumblestrip check` runs them together with the config embedded, so an app never reads a config of its own.

## Why not a custom golangci-lint build with module plugins

It would merge the outputs, but it rebuilds golangci-lint for every rumblestrip change and ties Rumblestrip to its plugin API; two binaries that each print `file:line: message` are simpler to pin and to prove.
