// Copyright 2019 Schibsted

package gobuild

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func executeWithLdFlagsAndPkg(ldflags []string, name string, args ...string) {
	args = appendFromEnv(args, "GOBUILD_FLAGS")
	args = append(args, ldflags...)
	executeWithPkg(name, args...)
}

func executeWithTestFlagsAndPkg(name string, args ...string) {
	args = appendFromEnv(args, "GOBUILD_FLAGS")
	args = appendFromEnv(args, "GOBUILD_TEST_FLAGS")
	executeWithPkg(name, args...)
}

func executeWithPkg(name string, args ...string) {
	if *pkg != "" {
		args = append(args, *pkg)
	}
	execute(name, args...)
}

func execute(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runWithBuildFlagsAndPkg(out io.Writer, name string, args ...string) error {
	args = appendFromEnv(args, "GOBUILD_FLAGS")
	if *pkg != "" {
		args = append(args, *pkg)
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func appendFromEnv(args []string, envname string) []string {
	for _, v := range strings.Fields(os.Getenv(envname)) {
		args = append(args, v)
	}
	return args
}
