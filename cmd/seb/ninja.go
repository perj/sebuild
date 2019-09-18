// Copyright 2018 Schibsted

package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func FindNinja() (string, error) {
	return exec.LookPath("ninja")
}

func RunNinja(ninja, bp string) error {
	if ninja == "" {
		var err error
		ninja, err = FindNinja()
		if err != nil {
			return err
		}
	}
	stamp := filepath.Join(bp, "stamp")
	ioutil.WriteFile(stamp, nil, 0666)
	bnpath := filepath.Join(bp, "build.ninja")
	argv := append([]string{"ninja", "-f", bnpath}, flag.Args()...)
	return syscall.Exec(ninja, argv, os.Environ())
}
