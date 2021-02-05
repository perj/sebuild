// Copyright 2019 Schibsted

package gobuild

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func buildCArchive(mode string, ldflags []string) {
	// go build links an executable to extract the symbols. If this is a plugin there'll be
	// unresolved symbols. Ignore now, handle in final link.
	// Only works with GNU ld right now.
	if runtime.GOOS != "darwin" {
		str := os.Getenv("CGO_LDFLAGS")
		os.Setenv("CGO_LDFLAGS", "-Wl,--unresolved-symbols=ignore-in-object-files "+str)
	}

	if mode == "piclib" {
		// -a to build standard libs with -shared
		executeWithLdFlagsAndPkg(ldflags, "go", "build", "-pkgdir", abspkgdir+"/gopkg_piclib", "-installsuffix=piclib", "-buildmode=c-archive", "-gcflags=-shared", "-asmflags=-shared", "-a", "-o", absout)
	} else {
		executeWithLdFlagsAndPkg(ldflags, "go", "build", "-pkgdir", abspkgdir+"/gopkg_lib", "-installsuffix=lib", "-buildmode=c-archive", "-o", absout)
	}

	// If there weren't any exports the header won't be created, but we expect it to be there.
	header := strings.TrimSuffix(absout, ".a") + ".h"
	f, _ := os.OpenFile(header, os.O_WRONLY|os.O_CREATE, 0666)
	if f != nil {
		f.Close()
	}

	if *libNoInit && runtime.GOOS != "darwin" {
		// Try to disable auto-start of go runtime. We want to be able to fork.
		// Don't know how to do it on Darwin right now.
		symbol := "_rt0_" + runtime.GOARCH + "_" + runtime.GOOS + "_lib"
		err := exec.Command("objcopy", "--rename-section", ".init_array=go_init", "--globalize-symbol="+symbol, absout).Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
