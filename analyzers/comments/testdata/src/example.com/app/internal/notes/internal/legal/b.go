// SPDX-License-Identifier: MIT

// Package legal is inside a module: its doc is not a license. // want `M2: a comment says why`
package legal

// Copyright is not a header down here. // want `M2: a comment says why`
func B() int { return 2 }
