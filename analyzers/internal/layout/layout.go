// Package layout tells, from a package's import path, which part of the
// kit's architecture it belongs to:
//
//	cmd/<app>                        main: wiring only
//	internal/<m>                     a module's contract
//	internal/<m>/internal/domain     the module's pure domain
//	internal/<m>/internal/<feature>  a vertical slice, or infra
//	internal/platform/...            the shared kernel
package layout

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strings"
)

var (
	domainPath   = regexp.MustCompile(`(^|/)internal/[^/]+/internal/domain(/|$)`)
	platformPath = regexp.MustCompile(`(^|/)internal/platform(/|$)`)
	innerPath    = regexp.MustCompile(`(^|/)internal/[^/]+/internal/([^/]+)$`)
)

// IsMain reports whether pkg is a main package.
func IsMain(pkg *types.Package) bool {
	return pkg.Name() == "main"
}

// IsContract reports whether pkg is visible outside its module: a module's
// contract, the platform, or a library package. A package under a second
// internal directory is a module's inside, never a contract.
func IsContract(pkg *types.Package) bool {
	return !IsMain(pkg) && strings.Count("/"+pkg.Path()+"/", "/internal/") <= 1
}

// IsDomain reports whether pkg is part of a module's pure domain.
func IsDomain(pkg *types.Package) bool {
	return domainPath.MatchString(pkg.Path())
}

// IsSlice reports whether pkg is a vertical slice: a module's inside that is
// neither its domain nor its infra.
func IsSlice(pkg *types.Package) bool {
	m := innerPath.FindStringSubmatch(pkg.Path())
	return m != nil && m[2] != "domain" && m[2] != "infra"
}

// IsPlatform reports whether pkg is part of the shared kernel.
func IsPlatform(pkg *types.Package) bool {
	return platformPath.MatchString(pkg.Path())
}

// IsTest reports whether f is a _test.go file.
func IsTest(fset *token.FileSet, f *ast.File) bool {
	return strings.HasSuffix(fset.File(f.Pos()).Name(), "_test.go")
}
