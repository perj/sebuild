// Copyright 2019 Schibsted

package header_install

import (
	"fmt"
	"io"
	"os"
)

func Main(args ...string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s -tool header-install <in> <out>\n", os.Args[0])
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
	_, err = fmt.Fprintf(outf, "# line 1 \"%s\"\n", in)
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
