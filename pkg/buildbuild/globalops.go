// Copyright 2018 Schibsted

// This package creates ninja build files from source Builddesc files.  The
// package documentation is currently a work in progress. Normally you do not
// use this package directly but rather use the seb tool to invoke it.
package buildbuild

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Global build configuration.
//
// After the initial parsing we'll only have one GlobalOps. Each ops descriptor
// will contain any possible flavorings. Before finalize we create flavor_ops
// for each flavor.  The way to consolidate targets that are untainted by a
// flavor is to:
//
// * Create a global flavor
//
// * Install unflavored includes and libs in the global flavor
//
// * Install includes and libs in each flavor
//
// * Link final targets in the flavored directories only.
//
// * Create $tool_ variables automagically
//
// Resolving flavor should be a matter of walking dependencies and seeing if
// the dependency is flavored. Flavored dependencies spread the flavor to
// whatever depends on them. Flavor taint is triggered on:
//
// * specialsrcs: Unknown rules (or in).
//
// * .in: The quintessential flavoring.
//
// * extravars: Always
//
// * deps, copts, cflags, cwarnflags, conlyflags, cxxflags, indirs: If contain $
type GlobalOps struct {
	// Command line options.
	Options struct {
		WithFlavors    map[string]bool
		WithoutFlavors map[string]bool
		Debug          bool
		Quiet          bool
	}
	// Result of parsing CONFIG.
	Config        Config
	FlavorConfigs map[string]*FlavorConfig

	// If non-nil, called after parsing CONFIG.
	PostConfigFunc func(ops *GlobalOps) error

	// List of Builddesc files that were parsed to generate this. Might contain duplicates.
	Builddescs []string

	// Cached output of BuildversionScript
	Buildversion string

	// All defined descriptors, some are for all flavors, some are duplicated for each flavor.
	Descriptors []Descriptor
	// Descriptors for internal libraries (dummy or not). XXX what about flavored libs?
	Libs map[string]LibDescriptor
	// Some descriptors also generate analyse targets, stored here.
	Analyses []*Analyser

	// Version checks, e.g. to verify C compiler.
	VersionChecks  map[string]func() error
	CC             string
	CXX            string
	CompilerFlavor string

	// Targets can be collected in variables and then used in other targets.
	CollectedVars map[string][]string

	// Callback to build plugins. As of go 1.8beta1, plugins can only be loaded from "main" package.
	// See https://github.com/golang/go/issues/18120
	BuildPlugin func(ops *GlobalOps, ppath string) error
}

func NewGlobalOps() *GlobalOps {
	ops := &GlobalOps{}
	ops.Options.WithFlavors = make(map[string]bool)
	ops.Options.WithoutFlavors = make(map[string]bool)
	ops.DefaultConfig()
	ops.DefaultCompiler()
	ops.CollectedVars = make(map[string][]string)

	ops.VersionChecks = make(map[string]func() error)

	return ops
}

func (ops *GlobalOps) RunFinalizers() (err error) {
	defer func() {
		p := recover()
		if perr, ok := err.(error); ok {
			err = perr
		} else if p != nil {
			panic(p)
		}
	}()
	for _, desc := range ops.Descriptors {
		desc.Finalize(ops)
	}
	return
}

// Expands globs relative to a source directory.
// Strings that aren't globs or don't match will be returned unchanged.
// If we get a glob match we register the directory as a dependency for
// rebuilding the build files.
func (ops *GlobalOps) GlobDir(srcdir string, srcs []string) []string {
	// Make sure to remove the / in the TrimPrefix below.
	if srcdir != "" && !strings.HasSuffix(srcdir, "/") {
		srcdir += "/"
	}

	var ret []string
	filter := make(map[string]bool)
	for _, src := range srcs {
		ops.RegisterGlob(srcdir, src)
		globs, err := filepath.Glob(srcdir + src)
		if err != nil {
			log.Print("Warning: Glob failed:", err)
			continue
		}
		if len(globs) == 0 {
			// filepath.Glob will return an empty slice if no source was found,
			// even if there wasn't a glob in src.
			globs = []string{srcdir + src}
		}
		for _, f := range globs {
			f = strings.TrimPrefix(f, srcdir)
			// Add each file only once, but try to do it in a stable fashion.
			if !filter[f] {
				ret = append(ret, f)
				filter[f] = true
			}
		}
	}
	return ret
}

func (ops *GlobalOps) RegisterGlob(srcdir, src string) {
	// If there's a glob, we need to re-run if the directory contents change.
	// Thus we mark the directory as a Builddesc
	if strings.ContainsAny(filepath.Base(src), "*?[]") {
		d := filepath.Dir(src)
		ops.RegisterGlob(srcdir, d)
		// XXX The + "/" here is probably not needed.
		globs, err := filepath.Glob(filepath.Join(srcdir, d) + "/")
		if err != nil {
			log.Print("Warning: Glob failed:", err)
			return
		}
		ops.Builddescs = append(ops.Builddescs, globs...)
	}
}

func (ops *GlobalOps) ReExec() {
	if _, ok := os.LookupEnv("BUILDTOOLDIR"); !ok {
		btp := BuildtoolDir()
		os.Setenv("BUILDTOOLDIR", btp)
	}
	binpath := filepath.Join(ops.Config.Buildpath, "obj", "_build_build")
	if err := syscall.Exec(binpath, os.Args, syscall.Environ()); err != nil {
		panic(err)
	}
}

type PluginDep struct {
	ppath string
	os.FileInfo
}

func (ops *GlobalOps) PluginDeps() (deps []PluginDep) {
	for _, ppath := range ops.Config.Plugins {
		/* Too slow
		cmd := exec.Command("go", "list", "-f", ` {{$dir := .Dir}}
			{{- .Dir}}
			{{- range .GoFiles}} {{$dir}}/{{.}}{{end}}
			{{- range .CgoFiles}} {{$dir}}/{{.}}{{end}}
			{{- range .CFiles}} {{$dir}}/{{.}}{{end}}
			{{- range .HFiles}} {{$dir}}/{{.}}{{end}}`,
			"./"+ppath)
		cmd.Stderr = os.Stderr
		files, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		deps = append(deps, strings.Fields(string(files))...)
		*/
		df, err := os.Open(ppath)
		if err != nil {
			panic(err)
		}
		infos, err := df.Readdir(0)
		if err != nil {
			panic(err)
		}
		df.Close()
		for _, info := range infos {
			n := info.Name()
			if strings.HasPrefix(n, ".") || strings.HasSuffix(n, "~") {
				continue
			}
			deps = append(deps, PluginDep{ppath, info})
		}
	}
	return
}

func (ops *GlobalOps) MaybeReExec() {
	binpath := filepath.Join(ops.Config.Buildpath, "obj", "_build_build")
	if binpath == os.Args[0] {
		return
	}
	theirinfo, err := os.Stat(binpath)
	if err != nil {
		// If they don't exist then we proceed ourselves.
		return
	}
	me, err := exec.LookPath(os.Args[0])
	if err != nil {
		var tmp error
		me, tmp = filepath.EvalSymlinks("/proc/self/exe")
		if tmp != nil {
			// Return the original error, /proc/self is just a fallback.
			panic(err)
		}
	}
	if me == binpath {
		return
	}
	myinfo, err := os.Stat(me)
	if err != nil {
		panic(err)
	}
	if !theirinfo.ModTime().After(myinfo.ModTime()) {
		// If we're newer than them, then we can't trust them to be up to date.
		return
	}
	// Check plugins as well.
	deps := ops.PluginDeps()
	for _, pi := range deps {
		if !theirinfo.ModTime().After(pi.ModTime()) {
			return
		}
	}
	ops.ReExec()
}

func NormalizePath(basedir, p string) string {
	if !strings.HasPrefix(p, ".") {
		return p
	}
	return filepath.Join(basedir, p)
}
