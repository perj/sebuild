// Copyright 2018 Schibsted

// Tool for compiling projects with ninja.
//
// Please see seb.1.ronn.md and COMPILING.md for more information
// about this tool.
package main

//go:generate go-bindata -nomemcopy -prefix ../../ ../../internal/tools ../../rules/...

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/schibsted/sebuild/pkg/buildbuild"
)

type SetFlag map[string]bool

func (s SetFlag) Set(v string) error {
	s[v] = true
	return nil
}

func (s SetFlag) String() string {
	var keys []string
	for k := range s {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

var (
	noexec  bool
	install bool
	topdir  string
)

func main() {
	ops := buildbuild.NewGlobalOps()
	// Disable BuildPlugin for now, it's too buggy on 1.8beta1
	//ops.BuildPlugin = BuildPlugin
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [--] [<ninja-args>]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.BoolVar(&ops.Options.Debug, "debug", false, "Enable debug output")
	flag.BoolVar(&ops.Options.Quiet, "quiet", false, "Silence default output")
	flag.Var(SetFlag(ops.Options.WithFlavors), "with-flavor", "Only generate this flavor (can be used multiple times). Usually not needed as each flavor is also a ninja pseudo-target.")
	flag.Var(SetFlag(ops.Options.WithoutFlavors), "without-flavor", "Don't generate this flavor (can be used multiple times)")
	flag.Var(SetFlag(ops.Config.Conditions), "condition", "Add build condition (can be used multiple times)")
	flag.BoolVar(&noexec, "noexec", false, "Don't execute ninja")
	flag.BoolVar(&install, "install", false, "Install ninja runtime into $HOME/.seb/")
	flag.StringVar(&topdir, "topdir", "", "Set top directory manually instead of scanning for Builddesc.top")
	flag.Parse()

	if install {
		installTools(ops.Options.Quiet)
		return
	}

	if topdir == "" {
		var err error
		topdir, _, err = FindTopdir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if topdir != "" {
		// Print make-style cwd message, to help editors like vim.
		if !ops.Options.Quiet {
			fmt.Printf("%s: Entering directory `%s'\n", filepath.Base(os.Args[0]), topdir)
		}
		err := os.Chdir(topdir)
		if err != nil {
			log.Fatal(err)
		}
	}

	if !noexec && os.Getenv("BUILD_BUILD_FROM_NINJA") == "" {
		ops.PostConfigFunc = func(ops *buildbuild.GlobalOps) error {
			bnpath := filepath.Join(ops.Config.Buildpath, "build.ninja")
			_, err := os.Stat(bnpath)
			if err != nil {
				return nil
			}
			// build.ninja already exists. Let's just exec ninja and let it
			// re-invoke us if needed.
			ninja, err := FindNinja()
			if err != nil {
				return nil
			}
			return RunNinja(ninja, bnpath)
		}
	}

	err := ops.ReadComponent("", nil)
	if err != nil {
		log.Fatal(err)
	}
	if !ops.Options.Quiet && len(ops.Config.ActiveFlavors) != len(ops.Config.AllFlavors) {
		fmt.Printf("Building only requested flavor(s): %s\n",
			strings.Join(ops.Config.ActiveFlavors, ", "))
	}
	err = ops.RunFinalizers()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range ops.VersionChecks {
		err := f()
		if err != nil {
			log.Fatal(err)
		}
	}
	err = ops.OutputTop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Success, lets run ninja, unless they run us.
	if noexec || os.Getenv("BUILD_BUILD_FROM_NINJA") != "" {
		return
	}

	bnpath := filepath.Join(ops.Config.Buildpath, "build.ninja")
	log.Fatal(RunNinja("", bnpath))
}

func installTools(quiet bool) {
	dir := os.Getenv("HOME")
	if dir == "" {
		fmt.Fprintln(os.Stderr, "$HOME is unset, can't install.")
		os.Exit(1)
	}
	dir = filepath.Join(dir, ".seb")
	err := RestoreAssets(dir, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !quiet {
		fmt.Printf("Seb tools installed to %s\n", dir)
	}
}
