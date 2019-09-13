// Copyright 2019 Schibsted

// Packge link is a wrapper will parse rsp files for non-GNU ld versions.  GNU
// ld supports them natively and we just exec it if we detect GNU.
package link

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Main(args ...string) {
	ldout, err := exec.Command("ld", "-v").Output()
	if err != nil || !bytes.Contains(ldout, []byte("GNU")) {
		args = expandRsp(args)
	}
	bin, err := exec.LookPath(args[0])
	if err == nil {
		err = syscall.Exec(bin, args, os.Environ())
	}
	fmt.Fprintf(os.Stderr, "Failed to run %s: %s\n", args[0], err)
	os.Exit(1)
}

func expandRsp(args []string) []string {
	newargs := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasSuffix(arg, ".rsp") {
			newargs = append(newargs, arg)
			continue
		}
		arg = strings.TrimPrefix(arg, "@")
		// I've successfully passed over 1G of command line on OS X, so this should work.
		f, err := os.Open(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open %s: %s\n", arg, err)
			os.Exit(1)
		}
		s := bufio.NewScanner(f)
		s.Split(bufio.ScanWords)
		for s.Scan() {
			newargs = append(newargs, s.Text())
		}
		if err := s.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to scan %s: %s\n", arg, err)
			os.Exit(1)
		}
		f.Close()
	}
	return newargs
}
