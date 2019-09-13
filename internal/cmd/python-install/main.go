// Copyright 2019 Schibsted

package python_install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func Main(args ...string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s -tool python-install <in> <out>\n", os.Args[0])
		os.Exit(1)
	}
	in := args[0]
	out := args[1]

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
	_, err = io.Copy(outf, inf)
	outf.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output to %s: %s", out, err)
		os.Remove(out)
		os.Exit(1)
	}

	pbin, err := exec.LookPath("python3")
	if err != nil {
		pbin, err = exec.LookPath("python")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find python binary: %s\n", err)
		os.Remove(out)
		os.Exit(1)
	}
	err = exec.Command(pbin, "-m", "py_compile", out).Run()
	if err == nil {
		err = exec.Command(pbin, "-O", "-m", "py_compile", out).Run()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run py_compile: %s\n", err)
		os.Remove(out)
		os.Exit(1)
	}
}
