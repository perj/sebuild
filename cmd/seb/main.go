// Copyright 2018-2019 Schibsted

// Tool for compiling projects with ninja.
//
// Please see seb.1.ronn.md and https://schibsted.github.io/sebuild for more
// information about this tool.
package main

//go:generate go-bindata -nomemcopy -ignore Builddesc -ignore cmd.*\.go$ -prefix ../../ ../../internal/... ../../rules/...

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	copy_analyse "github.com/schibsted/sebuild/internal/cmd/copy-analyse"
	go_install "github.com/schibsted/sebuild/internal/cmd/go-install"
	"github.com/schibsted/sebuild/internal/cmd/gobuild"
	gperf_enum "github.com/schibsted/sebuild/internal/cmd/gperf-enum"
	header_install "github.com/schibsted/sebuild/internal/cmd/header-install"
	"github.com/schibsted/sebuild/internal/cmd/in"
	"github.com/schibsted/sebuild/internal/cmd/invars"
	"github.com/schibsted/sebuild/internal/cmd/link"
	python_install "github.com/schibsted/sebuild/internal/cmd/python-install"
	"github.com/schibsted/sebuild/internal/cmd/ronn"
	"github.com/schibsted/sebuild/internal/cmd/touch"
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

type ArrayFlag []string

func (a *ArrayFlag) Set(v string) error {
	*a = append(*a, v)
	return nil
}

func (a *ArrayFlag) String() string {
	return strings.Join(*a, ", ")
}

var (
	noexec  bool
	install bool
	topdir  string
)

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "-tool" {
		mainTool()
	}
	ops := buildbuild.NewGlobalOps()
	ops.BuildPlugin = BuildPlugin
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [--] [<ninja-args>]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.BoolVar(&ops.Options.Debug, "debug", false, "Enable debug output")
	flag.BoolVar(&ops.Options.Quiet, "quiet", false, "Silence default output")
	flag.Var(SetFlag(ops.Options.WithFlavors), "with-flavor", "Only generate this flavor. Can be used multiple times. Usually not needed as each flavor is also a ninja pseudo-target.")
	flag.Var(SetFlag(ops.Options.WithoutFlavors), "without-flavor", "Don't generate this flavor. Can be used multiple times.")
	flag.Var(SetFlag(ops.Config.Conditions), "condition", "Add build condition. Can be used multiple times.")
	flag.BoolVar(&noexec, "noexec", false, "Don't execute ninja")
	flag.BoolVar(&install, "install", false, "Install ninja runtime into $HOME/.seb/")
	flag.StringVar(&topdir, "topdir", "", "Set top directory manually instead of scanning for Builddesc.top")
	flag.Var((*ArrayFlag)(&ops.Config.Configvars), "configvars", "Add a configvars file. These are read before configvars files in CONFIG. Can be used multiple times.")
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
			// Either nocgo condition or CGO_ENABLED=0 env sets both of them.
			if ops.Config.Conditions["nocgo"] {
				os.Setenv("CGO_ENABLED", "0")
			} else if os.Getenv("CGO_ENABLED") == "0" {
				ops.Config.Conditions["nocgo"] = true
			}

			bnpath := filepath.Join(ops.Config.Buildpath, "build.ninja")
			f, err := os.Open(bnpath)
			if err != nil {
				return nil
			}
			// build.ninja already exists. If the header match then
			// we just exec ninja and let it re-invoke us if needed.
			// If we get any errors here try the long path.
			defer f.Close()
			var ourb bytes.Buffer
			fmt.Fprintf(&ourb, "# %s\n", buildbuild.BuildBuildArgs(os.Args))
			fmt.Fprintf(&ourb, "# %s\n", buildbuild.BuildtoolDir())
			ours := ourb.Bytes()
			theirs := make([]byte, len(ours))
			if _, err := io.ReadFull(f, theirs); err != nil {
				return nil
			}
			if !bytes.Equal(ours, theirs) {
				if ops.Options.Debug {
					fmt.Fprintf(os.Stderr, "ninja.build arguments mismatch, `%s' != `%s'\n", ours, theirs)
				}
				return nil
			}

			ninja, err := FindNinja()
			if err != nil {
				return nil
			}
			return RunNinja(ninja, ops.Config.Buildpath)
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

	log.Fatal(RunNinja("", ops.Config.Buildpath))
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

func mainTool() {
	switch os.Args[2] {
	case "gobuild":
		gobuild.Main(os.Args[3:]...)
	case "link":
		link.Main(os.Args[3:]...)
	case "go-install":
		go_install.Main(os.Args[3:]...)
	case "header-install":
		header_install.Main(os.Args[3:]...)
	case "python-install":
		python_install.Main(os.Args[3:]...)
	case "ronn":
		ronn.Main(os.Args[3:]...)
	case "in":
		in.Main(os.Args[3:]...)
	case "touch":
		touch.Main(os.Args[3:]...)
	case "gperf-enum":
		gperf_enum.Main(os.Args[3:]...)
	case "copy-analyse":
		copy_analyse.Main(os.Args[3:]...)
	case "invars":
		invars.Main(os.Args[3:]...)
	default:
		fmt.Fprintf(os.Stderr, "Unknown tool %q.\n", os.Args[2])
		os.Exit(1)
	}
	os.Exit(0)
}
