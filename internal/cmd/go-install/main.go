// Copyright 2019 Schibsted

package go_install

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func Main(args ...string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s -tool go-install <in> <out>\n", os.Args[0])
		os.Exit(1)
	}
	in := args[0]
	out := args[1]

	fmtout, err := exec.Command("gofmt", "-l", in).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run gofmt: %s\n", err)
		os.Exit(1)
	}
	fmtout = bytes.TrimSpace(fmtout)
	if len(fmtout) > 0 {
		cmd := exec.Command("gofmt", in)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		os.Exit(1)
	}
	absin, err := filepath.Abs(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to determine full path of %s: %s", in, err)
		os.Exit(1)
	}

	inf, err := os.Open(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open input %s: %s", in, err)
		os.Exit(1)
	}
	outf, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open output %s: %s", out, err)
		os.Exit(1)
	}
	_, err = fmt.Fprintf(outf, "//line %s:1\n", absin)
	if err == nil {
		_, err = io.Copy(outf, inf)
	}
	outf.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output to %s: %s", out, err)
		os.Remove(out)
		os.Exit(1)
	}
}
