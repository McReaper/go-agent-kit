// Package all lists the kit's analyzers, for kitvet and kit check alike.
package all

import (
	"github.com/McReaper/go-agent-kit/analyzers/comments"
	"github.com/McReaper/go-agent-kit/analyzers/ctxroot"
	"github.com/McReaper/go-agent-kit/analyzers/domainconv"
	"github.com/McReaper/go-agent-kit/analyzers/goroutine"
	"github.com/McReaper/go-agent-kit/analyzers/logreturn"
	"github.com/McReaper/go-agent-kit/analyzers/nesting"
	"github.com/McReaper/go-agent-kit/analyzers/pkgname"
	"github.com/McReaper/go-agent-kit/analyzers/sqlconst"
	"github.com/McReaper/go-agent-kit/analyzers/testassert"
	"github.com/McReaper/go-agent-kit/analyzers/typedid"
	"golang.org/x/tools/go/analysis"
)

// Analyzers returns a new instance of every analyzer of the kit.
func Analyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		comments.New(),
		ctxroot.New(),
		domainconv.New(),
		goroutine.New(),
		logreturn.New(),
		nesting.New(),
		pkgname.New(),
		sqlconst.New(),
		testassert.New(),
		typedid.New(),
	}
}
