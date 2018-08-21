// Copyright 2018 Schibsted

package main

import (
	"flag"
	"os"
	"os/exec"
	"syscall"
)

func FindNinja() (string, error) {
	return exec.LookPath("ninja")
}

func RunNinja(ninja, bnpath string) error {
	if ninja == "" {
		var err error
		ninja, err = FindNinja()
		if err != nil {
			return err
		}
	}
	argv := append([]string{"ninja", "-f", bnpath}, flag.Args()...)
	return syscall.Exec(ninja, argv, os.Environ())
}
