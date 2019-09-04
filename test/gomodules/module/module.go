package main

import (
	"context"
	"log"
	"net"
	"strings"
)

func Test() string {
	// Use the net package to try to explicitly use cgo
	addrs, err := net.DefaultResolver.LookupHost(context.Background(), "localhost")
	if err != nil {
		log.Fatal(err)
	}
	return strings.Join(addrs, ", ")
}
