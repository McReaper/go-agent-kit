// Command kitvet runs the kit's own analyzers: the rules golangci-lint has
// no linter for. Usage: kitvet [-json] ./...
package main

import (
	"github.com/McReaper/go-agent-kit/analyzers/all"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(all.Analyzers()...)
}
