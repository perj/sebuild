// This file is here to add a dependency on go-bindata
// While that is only used for go generate we want to track
// it in go.mod. This file prevents go mod -sync from removing
// it from there.
// Ref https://github.com/golang/go/issues/25922

//+build tools

package main

import (
	_ "github.com/go-bindata/go-bindata/go-bindata"
)
