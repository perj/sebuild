// Copyright 2018 Schibsted

package buildbuild

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	gobuild "go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var (
	CantFindBuildtools = errors.New("Can't find directory containing buildtools. Use seb -install to install them manually.")
)

const modpath = "github.com/schibsted/sebuild"

var mkpathCache = make(map[string]struct{})

func mkpath(p ...string) {
	fullp := path.Join(p...)
	part := ""
	for _, step := range strings.Split(fullp, "/") {
		part += step + "/"
		if _, ok := mkpathCache[part]; !ok {
			os.Mkdir(part, 0777)
			mkpathCache[part] = struct{}{}
		}
	}
}

// Buildtooldir calls the package level BuildtoolDir. Kept for v1
// compatibility only.
func (ops *GlobalOps) BuildtoolDir() string {
	return BuildtoolDir()
}

// Buildtooldir figures out what directory contains the sebuild ninja runtime.
// First it tries to import the go package and check for sources, if that
// fails it looks for a directory based on the binary path. If that also fails
// it checks $HOME/.seb/
func BuildtoolDir() string {
	if p := os.Getenv("BUILDTOOLDIR"); p != "" {
		return p
	}

	// First check GOPATH.
	pkg, err := gobuild.Import(modpath+"/cmd/seb", "", gobuild.FindOnly)
	if err == nil {
		return filepath.Dir(filepath.Dir(pkg.Dir))
	}

	// Else derive from binary
	binp, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(err)
	}

	basep := filepath.Dir(binp)
	cands := []string{basep}
	realp, err := filepath.EvalSymlinks(basep)
	if err == nil {
		cands = append(cands, realp)
	}
	for _, p := range cands {
		if filepath.Base(p) == "bin" {
			// If directory containing binary is called "bin" we might have src there as well.
			srcp := filepath.Join(p, "../src/"+modpath)
			_, err := os.Stat(srcp)
			if err == nil {
				return srcp
			}
			// Also check for share/seb
			srcp = filepath.Join(p, "../share/seb")
			_, err = os.Stat(srcp)
			if err == nil {
				return srcp
			}
		}
		// If directory is called "sebuild" we're probably running from a source directory.
		if filepath.Base(p) == "sebuild" {
			// Check for rules.ninja just in case.
			_, err := os.Stat(filepath.Join(p, "rules/rules.ninja"))
			if err == nil {
				return filepath.Dir(p)
			}
		}
	}
	// Finally fallback to $HOME/.seb
	// XXX should probably check the version there somehow.
	if d := os.Getenv("HOME"); d != "" {
		_, err := os.Stat(filepath.Join(d, ".seb/rules/rules.ninja"))
		if err == nil {
			return filepath.Join(d, ".seb")
		}
	}
	panic(CantFindBuildtools)
}

func (ops *GlobalOps) GodepsStamp() string {
	return path.Join(ops.Config.Buildpath, "obj/_go/.stamp")
}

func (ops *GlobalOps) StatRulePath(pth string) bool {
	// Special handle $buildtooldir. Not super happy about it
	// but it's the only solution I can think of.
	if strings.HasPrefix(pth, "$buildtooldir") {
		pth = BuildtoolDir() + pth[len("$buildtooldir"):]
	}
	_, err := os.Stat(pth)
	return err == nil
}

// Strips args of anything after -- and returns it joined.
// In the future might remove some arguments before -- as well.
func BuildBuildArgs(args []string) string {
	for idx, arg := range args {
		if arg == "--" {
			args = args[:idx]
			break
		}
	}
	return strings.Join(args, " ")
}

func (ops *GlobalOps) OutputTop() (err error) {
	toppath := ops.Config.Buildpath
	mkpath(toppath)

	defer func() {
		p := recover()
		if perr, ok := p.(error); ok {
			err = perr
		} else if p != nil {
			panic(p)
		}
	}()

	topfile, err := os.OpenFile(path.Join(toppath, "build.ninja"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer topfile.Close()
	// Using bufio.Writer allows us to skip error checking until Flush.
	w := bufio.NewWriter(topfile)

	fmt.Fprintf(w, "# %s\n", BuildBuildArgs(os.Args))
	fmt.Fprintf(w, "# %s\n", BuildtoolDir())
	fmt.Fprintf(w, "# Flavors: %s\n", strings.Join(ops.Config.ActiveFlavors, ", "))
	conds := make([]string, 0, len(ops.Config.Conditions))
	for c := range ops.Config.Conditions {
		conds = append(conds, c)
	}
	sort.Strings(conds)
	fmt.Fprintf(w, "# Conditions: %s\n", strings.Join(conds, ", "))

	// While we usually use $buildpath to refer to the top path
	// ninja treats $builddir specially so set it as well.
	fmt.Fprintf(w, "builddir=%s\n", toppath)
	fmt.Fprintf(w, "buildpath=%s\n", toppath)
	fmt.Fprintf(w, "cc=%s\n", ops.CC)
	fmt.Fprintf(w, "cxx=%s\n", ops.CXX)

	// Copy some environment variables, then allow configvars ninja files
	// to override them. The reason to do it this way is that dependencies
	// don't work with environment variables. Changing a configvars file
	// does trigger rebuilds properly.
	fmt.Fprintf(w, "gopath=$$GOPATH\n")
	fmt.Fprintf(w, "gobuild_flags=$$GOBUILD_FLAGS\n")
	fmt.Fprintf(w, "gobuild_test_flags=$$GOBUILD_TEST_FLAGS\n")
	fmt.Fprintf(w, "cgo_enabled=$$CGO_ENABLED\n")

	fmt.Fprintf(w, "build_build = %s\n", BuildBuildArgs(os.Args))
	fmt.Fprintf(w, "buildtooldir=%s\n", BuildtoolDir())
	fmt.Fprintf(w, "inconfig = $buildtooldir/scripts/invars.sh\n")
	cv := strings.TrimSpace(strings.Join(ops.Config.Configvars, " "))
	if cv == "" {
		cv = "/dev/null"
	}
	fmt.Fprintf(w, "configvars = %s\n", cv)
	fmt.Fprintf(w, "include $buildtooldir/rules/defaults.ninja\n")
	for _, bp := range ops.Config.Buildparams {
		fmt.Fprintln(w, bp)
	}
	if ops.Config.CompilerRuleDir != "" {
		pth := ops.Config.CompilerRuleDir + "/" + ops.CompilerFlavor + ".ninja"
		if ops.StatRulePath(pth) {
			fmt.Fprintf(w, "include %s\n", pth)
		}
	}
	for _, cv := range ops.Config.Configvars {
		fmt.Fprintf(w, "include %s\n", cv)
	}
	for _, r := range ops.Config.Rules {
		fmt.Fprintf(w, "include %s\n", r)
	}
	fmt.Fprintf(w, "include $buildtooldir/rules/rules.ninja\n")
	if len(ops.Config.Godeps) > 0 {
		fmt.Fprintf(w, "build %s: %s %s\n", ops.GodepsStamp(), ops.Config.GodepsRule,
			strings.Join(ops.Config.Godeps, " "))
		mkpath(toppath, "obj/_go")
	}

	for _, f := range ops.Config.ActiveFlavors {
		ops.OutputFlavor(toppath, f)
		fmt.Fprintf(w, "subninja %s/obj/%s/build.ninja\n", toppath, f)
	}
	sort.Strings(ops.Builddescs)
	prevbd := ""
	for _, bd := range ops.Builddescs {
		if bd != prevbd {
			fmt.Fprintf(w, "build %s: phony\n", bd)
		}
		prevbd = bd
	}

	fmt.Fprintf(w, "build %s/build.ninja: generate_ninjas", toppath)
	for _, bd := range ops.Builddescs {
		if bd != prevbd {
			fmt.Fprint(w, " ", bd)
		}
		prevbd = bd
	}
	for _, pd := range ops.PluginDeps() {
		fmt.Fprint(w, " ", filepath.Join(pd.ppath, pd.Name()))
	}
	w.WriteByte('\n')

	fmt.Fprintf(w, "build all: phony %s\n", strings.Join(ops.Config.ActiveFlavors, " "))
	fmt.Fprintf(w, "default all\n")

	return w.Flush()
}

func (ops *GlobalOps) SetBuildversion() {
	data, err := exec.Command("sh", "-c", ops.Config.BuildversionScript).Output()
	if err != nil {
		panic(err)
	}
	ops.Buildversion = strings.TrimSpace(string(data))
}

func (ops *GlobalOps) OutputFlavor(topdir, flavor string) {
	destdir := path.Join(topdir, flavor)
	builddir := path.Join(topdir, "obj", flavor)
	buildvars := path.Join(builddir, "buildvars.ninja")

	flavorConf := ops.FlavorConfigs[flavor]
	prefix := ""
	if flavorConf != nil {
		prefix = flavorConf.Prefix
	}

	mkpath(destdir, prefix)
	mkpath(builddir)

	subninjas := make(map[string]bool)
	var defaults []string
	for _, desc := range ops.Descriptors {
		if !desc.ValidForFlavor(flavor) {
			continue
		}

		objbase := desc.DefaultObjectDir()
		objdir := objbase
		for i := 0; subninjas[objdir]; i++ {
			objdir = fmt.Sprint(objbase, i)
		}
		subninjas[objdir] = true
		defs := ops.OutputDescriptor(desc, builddir, objdir)
		defaults = append(defaults, defs...)
	}

	flfile, err := os.OpenFile(path.Join(builddir, "build.ninja"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(flfile)
	// Buildvars is separate here because we need to be able to include it from invars.sh
	fmt.Fprintf(w, "buildvars=%s\n", buildvars)
	fmt.Fprintf(w, "include $buildvars\n")

	if ops.Config.FlavorRuleDir != "" {
		pth := ops.Config.FlavorRuleDir + "/" + flavor + ".ninja"
		if ops.StatRulePath(pth) {
			fmt.Fprintf(w, "include %s\n", pth)
		}
	}
	if ops.Config.CompilerFlavorRuleDir != "" {
		pth := ops.Config.CompilerFlavorRuleDir + "/" + ops.CompilerFlavor + "-" + flavor + ".ninja"
		if ops.StatRulePath(pth) {
			fmt.Fprintf(w, "include %s\n", pth)
		}
	}
	var evs []string
	if flavorConf != nil {
		evs = append(evs, flavorConf.Extravars...)
	}
	evs = append(evs, ops.Config.Extravars...)
	for _, ev := range evs {
		fmt.Fprintf(w, "include %s\n", ev)
	}
	fmt.Fprintf(w, "include $buildtooldir/rules/static.ninja\n")
	for sn := range subninjas {
		fmt.Fprintf(w, "subninja %s/%s.ninja\n", builddir, sn)
	}

	fmt.Fprintf(w, "build %s/analyse: final_analyse", builddir)
	for _, an := range ops.Analyses {
		if len(an.OnlyForFlavors) > 0 {
			found := false
			for _, fl := range an.OnlyForFlavors {
				if flavor == fl {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		fmt.Fprintf(w, " %s/%s", builddir, an.TargetName)
	}
	w.WriteByte('\n')

	fmt.Fprintf(w, "build %s: phony %s\n", flavor, strings.Join(defaults, " "))

	if err := w.Flush(); err != nil {
		panic(err)
	}
	flfile.Close()

	// We need to be careful when writing the buildvars file. Since something (usually inconf) depends on buildvars
	// we don't want to rewrite it with the same data and trigger a rebuild.
	var bvbuf bytes.Buffer
	fmt.Fprintf(&bvbuf, "buildpath=%s\n", topdir)
	fmt.Fprintf(&bvbuf, "flavorroot=%s\n", destdir)
	fmt.Fprintf(&bvbuf, "destprefix=%s\n", prefix)
	fmt.Fprintf(&bvbuf, "destroot=%s\n", path.Join(destdir, prefix))
	fmt.Fprintf(&bvbuf, "builddir=%s\n", builddir)
	fmt.Fprintf(&bvbuf, `buildtools=$builddir/tools
incdir=$builddir/include
libdir=$builddir/lib
dest_bin=$destroot/bin
dest_mod=$destroot/modules
dest_lib=$destroot/lib
`)
	fmt.Fprintf(&bvbuf, "buildversion=%s\n", ops.Buildversion)
	fmt.Fprintf(&bvbuf, "buildflavor=%s\n", flavor)
	if flavorConf != nil {
		fmt.Fprintf(&bvbuf, "flavor_cflags=%s\n", flavorConf.Cflags)
	}
	fmt.Fprintf(&bvbuf, "\n")
	conds := make([]string, 0, len(ops.Config.Conditions))
	for c := range ops.Config.Conditions {
		conds = append(conds, c)
	}
	sort.Strings(conds)
	for _, c := range conds {
		fmt.Fprintf(&bvbuf, "%s=1\n", c)
	}

	oldbv, _ := ioutil.ReadFile(buildvars)
	if bytes.Compare(oldbv, bvbuf.Bytes()) == 0 {
		return
	}
	err = ioutil.WriteFile(buildvars, bvbuf.Bytes(), 0666)
	if err != nil {
		panic(err)
	}
}

func (ops *GlobalOps) OutputDescriptor(desc Descriptor, builddir, objdir string) (defaults []string) {
	mkpath(builddir, objdir)
	ninjaname := path.Join(builddir, objdir+".ninja")
	descfile, err := os.OpenFile(ninjaname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(descfile)

	desc.OutputHeader(w, objdir)

	multiTargets := make(map[string]bool)
	alltargets := desc.AllTargets()
	for tname, target := range alltargets {
		deps := desc.ResolveDeps(ops, tname)

		if len(target.Sources) == 0 && len(deps) == 0 && !target.Options["emptysrcs"] {
			continue
		}

		rule := target.Rule
		dest := path.Join(target.ResolveDest(), tname)
		if len(target.MultiTarget) > 0 {
			// Check already processed.
			if multiTargets[tname] {
				continue
			}
			dests := make([]string, len(target.MultiTarget))
			for i, t := range target.MultiTarget {
				multiTargets[t] = true
				dests[i] = path.Join(target.ResolveDest(), t)
			}
			dest = strings.Join(dests, " ")
		}
		orderDeps := desc.ResolveOrderDeps(target)
		srcs := desc.ResolveSrcs(ops, tname, target.Sources...)

		fmt.Fprintf(w, "build %s: %s ", dest, rule)
		fmt.Fprint(w, strings.Join(srcs, " "))

		if len(deps) > 0 {
			fmt.Fprint(w, " | ", strings.Join(deps, " "))
		}
		if len(orderDeps) > 0 {
			fmt.Fprint(w, " || ", strings.Join(orderDeps, " "))
		}
		fmt.Fprintln(w)
		for _, ea := range target.Extraargs {
			arr := strings.SplitN(ea, "=", 2)
			if len(arr) != 2 {
				fmt.Fprint(w, "    ", ea, "\n")
				continue
			}
			arr[1] = ops.ResolveCollectedVar(strings.TrimSpace(arr[1]))
			fmt.Fprint(w, "    ", arr[0], "=", arr[1], "\n")
		}
		if len(target.Srcopts) > 0 {
			fmt.Fprint(w, "    srcopts=", strings.Join(target.Srcopts, " "), "\n")
		}
		if target.Options["all"] {
			fmt.Fprintf(w, "default %s\n", dest)
			defaults = append(defaults, dest)
		}
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	descfile.Close()
	return defaults
}
