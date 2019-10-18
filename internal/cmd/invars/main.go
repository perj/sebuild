// Copyright 2019 Schibsted

package invars

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/schibsted/sebuild/v2/internal/pkg/assets"
)

func Main(args ...string) {
	var buildvars, invars, out string
	flagset := flag.NewFlagSet("invars", flag.ExitOnError)
	flagset.StringVar(&buildvars, "buildvars", "", "Buildvars file (buildvars.ninja), required.")
	flagset.StringVar(&invars, "invars", "", "Main invars file from CONFIG. If unset SEBUILD_INVARS_SH will be used, or the version stored in the binary failing that.")
	flagset.StringVar(&out, "out", "", "Output file. If used will also create a dependency file with .d appended.")
	flagset.Usage = func() {
		fmt.Fprintf(flagset.Output(), "Usage: %s -tool invars [options] <scripts...>\n", os.Args[0])
		flagset.PrintDefaults()
	}
	flagset.Parse(args)
	if buildvars == "" {
		fmt.Fprintf(os.Stderr, "-buildvars is required.\n")
		os.Exit(1)
	}
	if invars == "" {
		invars = os.Getenv("SEBUILD_INVARS_SH")
	}
	var w io.Writer
	var depw *bufio.Writer
	if out != "" {
		outf, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open output: %s\n", err)
			os.Exit(1)
		}
		defer outf.Close()
		w = outf
		depf, err := os.OpenFile(out+".d", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open depfile: %s\n", err)
			os.Remove(out)
			os.Exit(1)
		}
		defer depf.Close()
		depw = bufio.NewWriter(depf)
		fmt.Fprintf(depw, "%s: ", out)
	} else {
		w = os.Stdout
	}

	// Create the entire script first to more easily check for errors.
	var script bytes.Buffer
	if out != "" {
		fmt.Fprintf(&script, "export depfile=%s.d\n", out)
	}
	if err := copyScript(&script, buildvars, depw); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to copy buildvars: %s\n", err)
		if out != "" {
			os.Remove(out)
			os.Remove(out + ".d")
		}
		os.Exit(1)
	}
	scripts := flagset.Args()
	if invars == "" {
		script.WriteString(assets.InvarsSh)
	} else {
		scripts = append([]string{invars}, scripts...)
	}
	for _, f := range scripts {
		if err := copyScript(&script, f, depw); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy script: %s\n", err)
			if out != "" {
				os.Remove(out)
				os.Remove(out + ".d")
			}
			os.Exit(1)
		}
	}
	if depw != nil {
		if err := depw.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush depfile: %s\n", err)
			os.Remove(out)
			os.Remove(out + ".d")
			os.Exit(1)
		}
	}
	cmd := exec.Command("bash")
	cmd.Stdin = &script
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if out != "" {
			os.Remove(out)
			os.Remove(out + ".d")
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func copyScript(w *bytes.Buffer, pth string, depw *bufio.Writer) error {
	f, err := os.Open(pth)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	w.WriteRune('\n')
	if depw != nil {
		fmt.Fprintf(depw, " %s", pth)
	}
	return nil
}
