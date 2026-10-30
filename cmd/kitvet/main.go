// Command kitvet runs the kit's own analyzers: the rules golangci-lint has
// no linter for. Usage: kitvet [-json] ./...
package main

import (
	"github.com/McReaper/go-agent-kit/analyzers/comments"
	"github.com/McReaper/go-agent-kit/analyzers/ctxroot"
	"github.com/McReaper/go-agent-kit/analyzers/goroutine"
	"github.com/McReaper/go-agent-kit/analyzers/logreturn"
	"github.com/McReaper/go-agent-kit/analyzers/nesting"
	"github.com/McReaper/go-agent-kit/analyzers/pkgname"
	"github.com/McReaper/go-agent-kit/analyzers/sqlconst"
	"github.com/McReaper/go-agent-kit/analyzers/testassert"
	"github.com/McReaper/go-agent-kit/analyzers/typedid"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		comments.New(),
		ctxroot.New(),
		goroutine.New(),
		logreturn.New(),
		nesting.New(),
		pkgname.New(),
		sqlconst.New(),
		testassert.New(),
		typedid.New(),
	)
}
