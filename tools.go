//go:build tools

package main

import (
	_ "github.com/golang/mock/mockgen"
	_ "golang.org/x/tools/cmd/godoc"
	_ "golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment"
	_ "mvdan.cc/gofumpt"
)
