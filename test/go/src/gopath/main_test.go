// Copyright 2018 Schibsted

package main

import (
	"os"
	"strings"
	"testing"
)

func TestGoPath(t *testing.T) {
	for _, p := range strings.Split(os.Getenv("GOPATH"), ":") {
		if p[0] != '/' {
			t.Errorf("Not a full path: %s", p)
		}
	}
}
