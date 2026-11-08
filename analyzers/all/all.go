// Package all lists Rumblestrip's analyzers, for rumblestrip-vet and rumblestrip check alike.
package all

import (
	"github.com/McReaper/rumblestrip-go/analyzers/boundedlist"
	"github.com/McReaper/rumblestrip-go/analyzers/comments"
	"github.com/McReaper/rumblestrip-go/analyzers/ctxroot"
	"github.com/McReaper/rumblestrip-go/analyzers/domainconv"
	"github.com/McReaper/rumblestrip-go/analyzers/goroutine"
	"github.com/McReaper/rumblestrip-go/analyzers/logreturn"
	"github.com/McReaper/rumblestrip-go/analyzers/nesting"
	"github.com/McReaper/rumblestrip-go/analyzers/pkgname"
	"github.com/McReaper/rumblestrip-go/analyzers/sqlconst"
	"github.com/McReaper/rumblestrip-go/analyzers/testassert"
	"github.com/McReaper/rumblestrip-go/analyzers/typedid"
	"golang.org/x/tools/go/analysis"
)

// Analyzers returns a new instance of every analyzer of Rumblestrip.
func Analyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		boundedlist.New(),
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
