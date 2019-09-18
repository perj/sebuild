// Copyright 2019 Schibsted

package ronn

import (
	"fmt"
	"os"
	"os/exec"
)

func Main(args ...string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s -tool ronn <in> <out>\n", os.Args[0])
		os.Exit(1)
	}
	in := args[0]
	out := args[1]

	bin, err := exec.LookPath("ronn")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ronn not installed, missing manpage %s\n", in)
		os.Exit(0)
	}
	outf, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open output %s: %s\n", out, err)
		os.Exit(1)
	}
	cmd := exec.Command(bin, "--pipe", "--roff", in)
	cmd.Stdout = outf
	err = cmd.Run()
	outf.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run ronn: %s\n", err)
		os.Remove(out)
		os.Exit(1)
	}
}
