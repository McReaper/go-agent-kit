// Command rumblestrip-vet runs Rumblestrip's own analyzers: the rules golangci-lint has
// no linter for. Usage: rumblestrip-vet [-json] ./...
package main

import (
	"github.com/McReaper/rumblestrip-go/analyzers/all"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(all.Analyzers()...)
}
