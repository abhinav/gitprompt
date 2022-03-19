//go:build tools
// +build tools

package gitprompt

import (
	_ "golang.org/x/lint/golint"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
