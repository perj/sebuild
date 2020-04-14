// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

// Operations on different build descriptors.
// A number of default descriptors are defined, when a descriptor is matched, that default
// one is copied and the Parse function called.
// It will then fill in the rest of the struct, including the target map.
type GeneralDesc struct {
	Destdir       string          // Destination for the target.
	TargetOptions map[string]bool // Default options for the target (see Descriptor.AddTarget)

	Srcdir    string
	Builddesc string

	TargetName     string
	OnlyForFlavors []string

	Targets          map[string]*Target
	Srcdirs          map[string]string   // The directory where we've seen the first reference to a source to a source file. If the source file is not generated by something else it will be in this directory.
	Deps             map[string][]string // src->dependency dependencies.
	Srcopts          map[string][]string
	Gendeps          []string // Global dependencies.
	CollectTargetVar []string

	Extravars []string
	Buildvars map[string][]string
}

type Descriptor interface {
	NewFromTemplate(bd, tname string, flavors []string) Descriptor
	GetBuilddesc() string

	Parse(ops *GlobalOps, realsrcdir string, args map[string][]string) Descriptor
	Finalize(ops *GlobalOps)

	CompileSrc(srcdir, src, srcbase, ext string)

	AddTarget(tname, rule string, srcs []string, destdir, srcdir string, extraargs []string, options map[string]bool) *Target
	AddMultiTarget(tnames []string, target *Target)
	AllTargets() map[string]*Target

	ResolveSrcs(ops *GlobalOps, tname string, srcs ...string) []string
	ResolveDeps(ops *GlobalOps, tname string) []string
	ResolveOrderDeps(target *Target) []string

	ValidForFlavor(flavor string) bool
	DefaultObjectDir() string
	OutputHeader(w io.Writer, objdir string)
}

type Finalizer func(*GlobalOps, Descriptor)

type Analyser struct {
	TargetName     string
	OnlyForFlavors []string
}

var DefaultDescriptors = map[string]Descriptor{
	"PROG":          &ProgTemplate,
	"GOPROG":        &GoprogTemplate,
	"GOMODULE":      &GomoduleTemplate,
	"GOTEST":        &GotestTemplate,
	"TOOL_PROG":     &ToolProgTemplate,
	"TOOL_INSTALL":  &ToolInstallTemplate,
	"LIB":           &LibTemplate,
	"LINKERSET_LIB": &LinkersetLibTemplate,
	"MODULE":        &ModuleTemplate,
	"INSTALL":       &InstallTemplate,
}

var (
	BadSrcoptsFormat  = errors.New("Bad srcopts format, need srcopts[src:opt]")
	BadSpecialSrcs    = errors.New("Bad specialsrcs format, need specialsrcs[rule:src,...:target] or specialsrcs[rule:src,...:target:var=val...]")
	UnhandledArgument = errors.New("Unknown argument")
)

func (defdesc *GeneralDesc) NewFromTemplate(bd, tname string, flavors []string) *GeneralDesc {
	ret := new(GeneralDesc)
	*ret = *defdesc
	// Initialize all the maps
	ret.Targets = make(map[string]*Target)
	ret.Srcdirs = make(map[string]string)
	ret.Deps = make(map[string][]string)
	ret.Srcopts = make(map[string][]string)
	ret.Buildvars = make(map[string][]string)
	ret.Builddesc = bd
	ret.TargetName = tname
	ret.OnlyForFlavors = flavors
	return ret
}

type ParseGeneralParam func(ops *GlobalOps, g *GeneralDesc, args []string)

func (g *GeneralDesc) GetBuilddesc() string {
	return g.Builddesc
}

func (g *GeneralDesc) GenericParse(desc Descriptor, ops *GlobalOps, realsrcdir string, args map[string][]string, extra []string) Descriptor {
	eks := make(map[string]bool, len(extra))
	for _, k := range extra {
		eks[k] = true
	}
	keys := make(map[string]bool)
	for k := range args {
		if !eks[k] {
			keys[k] = true
		}
	}

	for _, inc := range args["INCLUDE"] {
		// It looks like realsrcdir is supposed to be the current directory we work in,
		// and desc.Srcdir should be the Builddesc srcdir (modified by srcdir param).
		// When we include files, we need to reset both to the included path.
		// Handle this by recursing INCLUDES before using srcdir, and resetting it
		// directly after INCLUDE is done.
		inc = NormalizePath(realsrcdir, inc)
		s, err := ops.OpenBuilddesc(inc)
		if err != nil {
			panic(err)
		}
		var incargs Args
		incargs.Parse(s, ops.CheckConditions)
		s.Close()
		parentbd := g.Builddesc
		g.Builddesc = inc
		desc.Parse(ops, path.Dir(inc), incargs.Unflavored)
		g.Builddesc = parentbd
	}
	delete(keys, "INCLUDE")

	srcdir := realsrcdir
	if len(args["srcdir"]) > 0 {
		srcdir = NormalizePath(realsrcdir, args["srcdir"][0])
	}
	delete(keys, "srcdir")
	g.Srcdir = srcdir

	if len(args["destdir"]) > 0 {
		g.Destdir = args["destdir"][0]
	}
	delete(keys, "destdir")

	for _, ev := range args["extravars"] {
		// XXX normalizePath
		g.Extravars = append(g.Extravars, path.Join(realsrcdir, ev))
	}
	delete(keys, "extravars")

	for _, bv := range ops.Config.Buildvars {
		g.Buildvars[bv] = append(g.Buildvars[bv], args[bv]...)
		delete(keys, bv)
	}

	g.CollectTargetVar = append(g.CollectTargetVar, args["collect_target_var"]...)
	delete(keys, "collect_target_var")

	for pv, pfun := range PluginGeneralParams {
		if len(args[pv]) > 0 {
			pfun(ops, g, args[pv])
		}
		delete(keys, pv)
	}

	for _, dep := range args["deps"] {
		if col := strings.IndexRune(dep, ':'); col >= 0 {
			d := dep[:col]
			s := dep[col+1:]
			g.Deps[d] = append(g.Deps[d], s)
		} else {
			g.Gendeps = append(g.Gendeps, dep)
		}
	}
	delete(keys, "deps")

	for _, srcopt := range args["srcopts"] {
		col := strings.IndexRune(srcopt, ':')
		if col < 0 {
			panic(&ParseError{BadSrcoptsFormat, srcopt, g.Builddesc})
		}
		s := srcopt[:col]
		g.Srcopts[s] = append(g.Srcopts[s], srcopt[col+1:])
	}
	delete(keys, "srcopts")

	for _, spsrc := range args["specialsrcs"] {
		// rule:srcs:target[:extra-vars] srcs and extra-vars comma separated
		spargs := strings.SplitN(spsrc, ":", 4)
		if len(spargs) < 3 {
			panic(&ParseError{BadSpecialSrcs, spsrc, g.Builddesc})
		}
		sprule := spargs[0]
		spsrcs := strings.Split(spargs[1], ",")
		sptarg := spargs[2]
		spextra := []string{}
		if len(spargs) >= 4 {
			spextra = strings.Split(spargs[3], ",")
		}
		// Store targets in the object directory, an install target should be added separately if needed.
		desc = CompileSpecial(desc, sptarg, sprule, ops.GlobDir(srcdir, spsrcs), "obj", realsrcdir, spextra, nil)
	}
	delete(keys, "specialsrcs")

	for _, src := range ops.GlobDir(srcdir, args["srcs"]) {
		CompileSrc(desc, srcdir, src)
	}
	delete(keys, "srcs")

	if len(keys) > 0 {
		var karr []string
		for k := range keys {
			karr = append(karr, k)
		}
		sort.Strings(karr)
		panic(&ParseError{UnhandledArgument, strings.Join(karr, ", "), g.Builddesc})
	}
	return desc
}

func (g *GeneralDesc) Finalize(ops *GlobalOps) {
	// To keep the collected vars stable between runs, we need to always
	// append to it in the same order. Thus we need to sort the target
	// keys.
	var targets []string
	for tname := range g.Targets {
		targets = append(targets, tname)
	}
	sort.Strings(targets)

	for _, cv := range g.CollectTargetVar {
		for _, tname := range targets {
			target := g.Targets[tname]
			if target.Options["all"] {
				tgt := path.Join(target.ResolveDest(), tname)
				ops.CollectedVars[cv] = append(ops.CollectedVars[cv], tgt)
			}
		}
	}
	for _, tname := range targets {
		target := g.Targets[tname]
		if cv := target.CollectAs; cv != "" {
			tgt := path.Join(target.ResolveDest(), tname)
			ops.CollectedVars[cv] = append(ops.CollectedVars[cv], tgt)
		}
	}
}

func (ops *GlobalOps) ResolveCollectedVar(src string) string {
	if !strings.HasPrefix(src, "$") {
		return src
	}
	if cv, ok := ops.CollectedVars[src[1:]]; ok {
		return strings.Join(cv, " ")
	}
	return src
}

func (g *GeneralDesc) ResolveSrcdir(src, tname string) string {
	// Check for full path or variable.
	if src != "" && strings.ContainsAny(src[0:1], "/$") {
		return ""
	}

	// Check if the source file is a target and use that directory
	// unless the source file is named the same as the target.
	if g.Targets[src] != nil && src != tname {
		return g.Targets[src].ResolveDest()
	}

	// Entry might not exist, but "" is the default we want.
	return g.Srcdirs[src]
}

func (g *GeneralDesc) ResolveSrcs(ops *GlobalOps, tname string, srcs ...string) []string {
	var ret []string

	for _, src := range srcs {
		src = ops.ResolveCollectedVar(src)
		ret = append(ret, path.Join(g.ResolveSrcdir(src, tname), src))
	}

	return ret
}

func (g *GeneralDesc) ResolveDeps(ops *GlobalOps, tname string) []string {
	var ret []string

	ret = append(ret, g.ResolveSrcs(ops, tname, g.Deps[tname]...)...)
	ret = append(ret, g.ResolveSrcs(ops, tname, g.Targets[tname].Deps...)...)
	gendeps := append([]string{}, g.Gendeps...)
	for idx, gd := range gendeps {
		if gd == tname {
			gendeps = append(gendeps[0:idx], gendeps[idx+1:]...)
			break
		}
	}
	ret = append(ret, g.ResolveSrcs(ops, tname, gendeps...)...)
	for _, src := range g.Targets[tname].Sources {
		ret = append(ret, g.ResolveSrcs(ops, src, g.Deps[src]...)...)
	}

	// depend on the compilation tool.
	rule := g.Targets[tname].Rule
	ret = append(ret, ops.Config.Ruledeps[rule]...)
	return ret
}

func (g *GeneralDesc) ResolveOrderDeps(target *Target) []string {
	return nil
}

func (g *GeneralDesc) OutputHeader(w io.Writer, objdir string) {
	fmt.Fprintf(w, "objdir=$builddir/%s\n", objdir)
	for _, ev := range g.Extravars {
		fmt.Fprintf(w, "include %s\n", ev)
	}
	for bv, v := range g.Buildvars {
		if len(v) > 0 {
			fmt.Fprintf(w, "%s=%s\n", bv, strings.Join(v, " "))
		}
	}
}

func (g *GeneralDesc) DefaultObjectDir() string {
	return path.Join(g.Srcdir, g.TargetName)
}

func (g *GeneralDesc) ValidForFlavor(flavor string) bool {
	if len(g.OnlyForFlavors) == 0 {
		return true
	}
	for _, fl := range g.OnlyForFlavors {
		if flavor == fl {
			return true
		}
	}
	return false
}
