// Copyright 2019 Schibsted

// Package gobuild provides the gobuild tool for sebuild.
//
// This package handles a lot of nuances about building go packages that are
// compatible with sebuild's other build modes. It also creates proper
// dependencies such that ninja can determine if it has to re-compile Go
// binaries or not.
//
// Normally, with go modules, relative paths are used to build, this fits well
// into using the top level directory for all things. Across modules and when
// using GOPATH we however cd into the source directory. A make-compatible
// diagnostic is in that case print to help editors such as vim to detect what
// directory error messages point to.
//
// The go tool will itself change the current directory. This means that many
// of the paths we give it have to be absolute paths. Effort is thus spent
// converting to those.
package gobuild

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	flagset = flag.NewFlagSet("gobuild", flag.ExitOnError)
	inpath  string
	outpath string
	depfile string
	pkg     = flagset.String("pkg", "", "Explicit go package name. If unset the relative path is used.")
	cflags  = flagset.String("cflags", "", "C flags to use with cgo.")
	ldflags = flagset.String("ldflags", "", "Linker flags to use with cgo. Can contain objects or flags.")
	mode    = flagset.String("mode", "prog", "Type of output. One of prog, prog-nocgo, module, test-prog, lib, piclib, test, bench, cover, cover_html")
	pkgdir  = flagset.String("pkgdir", "", "Directory to store compiled standard packages. Only used when custom versions are needed.")

	absin     string
	absout    string
	absdep    string
	abspkgdir string
	objs      []string
)

func Main(args ...string) {
	flagset.Usage = func() {
		fmt.Fprintf(flagset.Output(), "Usage: %s -tool gobuild [options] <in> <out> <depfile>\n", os.Args[0])
		flagset.PrintDefaults()
	}
	flagset.Parse(args)
	inpath = flagset.Arg(0)
	outpath = flagset.Arg(1)
	depfile = flagset.Arg(2)
	if inpath == "" {
		flagset.Usage()
		os.Exit(2)
	}
	if outpath == "" && needOutpath(*mode) {
		flagset.Usage()
		os.Exit(2)
	}
	if depfile == "" && needDepfile(*mode) {
		flagset.Usage()
		os.Exit(2)
	}

	if err := setAbsPaths(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	setRelPkg()
	if *pkg == "" && *mode != "cover_html" {
		// Ignore errors here, they'll be detected later.
		os.Chdir(absin)
		fmt.Printf("gobuild: Entering directory `%s'\n", absin)
	}

	if depfile != "" {
		depf, err := os.OpenFile(absdep, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		depdone := make(chan error)
		go func() {
			depdone <- writeDepfile(depf)
		}()
		defer func() {
			err := <-depdone
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}()
	}

	var ldflags []string
	if len(objs) > 0 {
		var extldflags strings.Builder
		extldflags.WriteString(`-extldflags "`)
		for _, obj := range objs {
			extldflags.WriteString(obj)
			extldflags.WriteString(" ")
		}
		extldflags.WriteString(os.Getenv("CGO_LDFLAGS"))
		ldflags = append(ldflags, "-ldflags", extldflags.String())
	}

	switch *mode {
	case "prog":
		executeWithLdFlagsAndPkg(ldflags, "go", "build", "-o", absout)
	case "prog-nocgo":
		os.Setenv("CGO_ENABLED", "0")
		executeWithLdFlagsAndPkg(ldflags, "go", "build", "-o", absout)
	case "test-prog":
		executeWithLdFlagsAndPkg(ldflags, "go", "test", "-c", "-o", absout)
	case "module":
		executeWithLdFlagsAndPkg(ldflags, "go", "build", "-o", absout, "-buildmode=plugin")
	case "test":
		executeWithTestFlagsAndPkg("go", "test")
	case "bench":
		bench := flagset.Arg(1)
		if bench == "" {
			bench = "."
		}
		executeWithTestFlagsAndPkg("go", "test", "-bench", bench)
	case "cover":
		executeWithTestFlagsAndPkg("go", "test", "-coverprofile="+absout)
	case "cover_html":
		checkGomodGopath()
		execute("go", "tool", "cover", "-html="+inpath, "-o", outpath)
	case "lib", "piclib":
		buildCArchive(*mode, ldflags)
	default:
		fmt.Fprintf(os.Stderr, "Unknown build mode %q\n", *mode)
		os.Exit(1)
	}
}

func needOutpath(mode string) bool {
	switch mode {
	case "test", "bench":
		return false
	}
	return true
}

func needDepfile(mode string) bool {
	switch mode {
	case "test", "bench", "cover_html":
		return false
	}
	return true
}

// setAbsPath sets absin, absout, absdep, abspkgdir, objs, CGO_CFLAGS,
// CGO_LDFLAGS, GOPATH to absolute paths.  It uses the globals inpath, output,
// cflags, ldflags, pkgdir and the GOPATH environment variable.
func setAbsPaths() (err error) {
	absin, err = filepath.Abs(inpath)
	if err != nil {
		return err
	}
	absout, err = filepath.Abs(outpath)
	if err != nil {
		return err
	}
	absdep, err = filepath.Abs(depfile)
	if err != nil {
		return err
	}
	if *pkgdir != "" {
		abspkgdir, err = filepath.Abs(*pkgdir)
		if err != nil {
			return err
		}
	}

	setAbsList("CGO_CFLAGS", " ", *cflags, false)
	setAbsList("CGO_LDFLAGS", " ", *ldflags, true)
	setAbsPathList("GOPATH", ":", os.Getenv("GOPATH"))
	return nil
}

func setAbsList(env, sep, list string, setobjs bool) {
	var liststr strings.Builder
	for _, elem := range strings.Fields(list) {
		if elem == "" {
			continue
		}
		if !strings.HasPrefix(elem, "-") {
			newelem, err := filepath.Abs(elem)
			// Errors are not fatal, just don't update.
			if err == nil {
				elem = newelem
			}
		}
		if setobjs && strings.HasSuffix(elem, ".o") {
			objs = append(objs, elem)
			continue
		}
		if liststr.Len() > 0 {
			liststr.WriteString(sep)
		}
		liststr.WriteString(elem)
	}
	os.Setenv(env, liststr.String())
}

func setAbsPathList(env, sep, list string) {
	var liststr strings.Builder
	for _, elem := range filepath.SplitList(list) {
		if elem == "" {
			continue
		}
		if !strings.HasPrefix(elem, "-") {
			newelem, err := filepath.Abs(elem)
			// Errors are not fatal, just don't update.
			if err == nil {
				elem = newelem
			}
		}
		if liststr.Len() > 0 {
			liststr.WriteString(sep)
		}
		liststr.WriteString(elem)
	}
	os.Setenv(env, liststr.String())
}

// setRelPkg might set *pkg to a relative if it's currently empty.
// In some cases this can't be done, and *pkg is unchanged empty string.
func setRelPkg() {
	if *pkg != "" {
		return
	}
	// Using a relative path helps the error messages, since we don't have to
	// change the current directory, but it won't work if it's not within the
	// same module.
	gomod, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return
	}
	gomod = bytes.TrimSpace(gomod)
	// Don't try this without go modules.
	if len(gomod) == 0 {
		return
	}
	// Check for cross module source directory. We can't use relative path with that.
	srcmodcmd := exec.Command("go", "env", "GOMOD")
	srcmodcmd.Dir = absin
	srcmod, err := srcmodcmd.Output()
	if err != nil {
		return
	}
	srcmod = bytes.TrimSpace(srcmod)
	if bytes.Equal(gomod, srcmod) {
		*pkg = "./" + inpath
	}
}

func checkGomodGopath() {
	// On Go 1.11 and 1.12 only, auto modules might fail here due to being
	// run from outside gopath but working on files inside. Go 1.11 is not
	// supported.
	if os.Getenv("GOPATH") == "" {
		return
	}
	modmode := os.Getenv("GO111MODULE")
	if modmode != "" && modmode != "auto" {
		return
	}
	gover, err := exec.Command("go", "version").Output()
	if err != nil {
		return
	}
	if !bytes.Contains(gover, []byte("go1.12")) {
		return
	}
	os.Setenv("GO111MODULE", "off")
}
