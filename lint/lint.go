// Package lint holds the kit's golangci-lint config, embedded so that kit
// check applies it whatever an app carries.
package lint

import _ "embed"

// Config is lint/golangci.yml.
//
//go:embed golangci.yml
var Config []byte
