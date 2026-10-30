package archtest

import (
	"fmt"
	"strings"
)

type kind int

const (
	unplaced kind = iota
	command
	contract
	domain
	infra
	slice
	platform
	library
)

type place struct {
	kind   kind
	module string
}

func classify(module, pkg string) place {
	rel, ok := strings.CutPrefix(pkg, module+"/")
	if !ok {
		return place{kind: unplaced}
	}
	parts := strings.Split(rel, "/")
	switch {
	case parts[0] == "cmd":
		return place{kind: command}
	case parts[0] == "pkg":
		return place{kind: library}
	case len(parts) >= 2 && parts[0] == "internal" && parts[1] == "platform":
		return place{kind: platform}
	case len(parts) == 2 && parts[0] == "internal":
		return place{kind: contract, module: parts[1]}
	case len(parts) >= 4 && parts[0] == "internal" && parts[2] == "internal":
		return place{kind: inner(parts[3]), module: parts[1]}
	}
	return place{kind: unplaced}
}

func inner(name string) kind {
	switch name {
	case "domain":
		return domain
	case "infra":
		return infra
	}
	return slice
}

func forbidden(module string, from place, path string) (string, string, bool) {
	to := classify(module, path)
	switch from.kind {
	case domain:
		return domainImport(to, path)
	case slice:
		return sliceImport(from, to, path)
	case platform:
		return platformImport(to, path)
	case unplaced, command, contract, infra, library:
		return "", "", false
	}
	return "", "", false
}

func domainImport(to place, path string) (string, string, bool) {
	if to.kind == library || to.kind == domain || isStandard(path) {
		return "", "", false
	}
	return "domain-import", fmt.Sprintf("E1: the domain imports %s: it may import the standard library and the module's pkg/ only, "+
		"so its rules run without any framework or I/O", path), true
}

func sliceImport(from, to place, path string) (string, string, bool) {
	switch {
	case to.module != from.module:
		return "", "", false
	case to.kind == slice:
		return "slice-import", fmt.Sprintf("E2: a slice imports another slice, %s: slices change independently; "+
			"move what both need into the domain, or call the other through the contract", path), true
	case to.kind == infra:
		return "slice-infra", fmt.Sprintf("E2: a slice imports %s: declare the interface the slice needs and let the contract "+
			"pass the infra in, so the slice tests without I/O", path), true
	}
	return "", "", false
}

func platformImport(to place, path string) (string, string, bool) {
	if to.module == "" {
		return "", "", false
	}
	return "platform-import", fmt.Sprintf("E1: the platform imports module %s: shared code knows no module, "+
		"or it couples every module to that one", path), true
}

// WHY: a standard library path has no dot in its first element; every
// module path that can be downloaded does.
func isStandard(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}
