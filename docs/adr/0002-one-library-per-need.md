# 2. One library per need

## Decision

| Need | Library | Denied |
| --- | --- | --- |
| MCP server | `github.com/modelcontextprotocol/go-sdk` | `mark3labs/mcp-go` |
| CLI | `github.com/spf13/cobra` | `urfave/cli` |
| Logging | `log/slog` | `log`, logrus, zap, zerolog |
| Errors | `errors`, `fmt.Errorf` with `%w` | `pkg/errors` |
| Test comparison | `github.com/google/go-cmp/cmp` | testify |
| Property tests | `pgregory.net/rapid` and `testing.F` | — |
| Goroutine leaks in tests | `go.uber.org/goleak` | — |
| Fakes | hand-written, implementing the interface | `golang/mock`, `go.uber.org/mock` |

depguard enforces the denied column (`lint/golangci.yml`).

## Consequences

An agent that reaches for another library gets a lint error naming the one to use.
