package main

import (
	"fmt"
	"runtime"
	"testing"
)

func TestGoArch(t *testing.T) {
	fmt.Println(runtime.GOARCH)
}
