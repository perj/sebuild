// Copyright 2018 Schibsted

package main

import "os"

func main() {
	println(os.Getenv("GOPATH"))
}
