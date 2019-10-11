// Copyright 2019 Schibsted

package asset

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/schibsted/sebuild/internal/pkg/assets"
)

func Main(args ...string) {
	var out string
	flagset := flag.NewFlagSet("asset", flag.ExitOnError)
	flagset.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -tool asset [options] [asset]\n", os.Args[0])
		flagset.PrintDefaults()
	}
	flagset.StringVar(&out, "out", "", "Output file instead of stdout.")
	flagset.Parse(args)

	w := os.Stdout
	if out != "" {
		f, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open outfile: %s\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}
	wb := bufio.NewWriter(w)
	var assets = map[string]string{
		"invars.sh":            assets.InvarsSh,
		"main.go":              assets.MainGo,
		"compiler/clang.ninja": assets.CompilerClangNinja,
		"compiler/gcc.ninja":   assets.CompilerGccNinja,
		"defaults.ninja":       assets.DefaultsNinja,
		"flavor/dev.ninja":     assets.FlavorDevNinja,
		"flavor/gcov.ninja":    assets.FlavorGcovNinja,
		"flavor/release.ninja": assets.FlavorReleaseNinja,
		"rules.ninja":          assets.RulesNinja,
		"static.ninja":         assets.StaticNinja,
	}
	ecode := 0
	var err error
	switch flagset.NArg() {
	case 0:
		keys := make([]string, 0, len(assets))
		for k := range assets {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintln(wb, k)
		}
		err = wb.Flush()
	case 1:
		a, ok := assets[flagset.Arg(0)]
		if ok {
			fmt.Fprint(wb, a)
			err = wb.Flush()
		} else {
			fmt.Fprintf(os.Stderr, "No such asset %q\n", flagset.Arg(0))
			ecode = 1
		}
	default:
		flagset.Usage()
		ecode = 1
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	if ecode != 0 && out != "" {
		os.Remove(out)
	}
	os.Exit(ecode)
}
