// Copyright 2018 Schibsted

package buildbuild

import (
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

type LinkDesc struct {
	GeneralDesc

	Picrules bool
	Link     string

	Incdirs  []string
	Objs     []string
	Libs     []string
	Includes []string
	Incdeps  []string // Include file dependecies. Either include files we install into the global include directory or include files generated by us.
	GoSrc    []string

	NoAnalyse   bool
	DontAnalyse map[string]bool

	IncdepsGenerated []string
}

func (tmpl *LinkDesc) NewFromTemplate(bd, tname string, flavors []string) *LinkDesc {
	ret := new(LinkDesc)
	*ret = *tmpl
	ret.GeneralDesc = *tmpl.GeneralDesc.NewFromTemplate(bd, tname, flavors)
	ret.DontAnalyse = make(map[string]bool)
	return ret
}

type ParseLinkerParam func(l *LinkDesc, args []string)

// Return keys handled by LinkDesc.Parse to pass to GenericParse
func LinkerExtra(extra ...string) []string {
	extra = append(extra, "incdirs", "no_analyse", "libs")
	for k := range PluginLinkerParams {
		extra = append(extra, k)
	}
	return extra
}

func (l *LinkDesc) LinkerParse(srcdir string, args map[string][]string) {
	if srcdir != "" {
		// Always add the source directory to include path.
		idx := sort.SearchStrings(l.Incdirs, srcdir)
		if idx >= len(l.Incdirs) || l.Incdirs[idx] != srcdir {
			l.Incdirs = append(l.Incdirs, "")
			copy(l.Incdirs[idx+1:], l.Incdirs[idx:])
			l.Incdirs[idx] = srcdir
		}
	}
	for _, inc := range args["incdirs"] {
		inc = NormalizePath(srcdir, inc)
		if inc != "" {
			idx := sort.SearchStrings(l.Incdirs, inc)
			if idx >= len(l.Incdirs) || l.Incdirs[idx] != inc {
				l.Incdirs = append(l.Incdirs, "")
				copy(l.Incdirs[idx+1:], l.Incdirs[idx:])
				l.Incdirs[idx] = inc
			}
		}
	}

	if args["no_analyse"] != nil {
		l.NoAnalyse = true
	}

	l.Libs = append(l.Libs, args["libs"]...)

	for pv, pfun := range PluginLinkerParams {
		if len(args[pv]) > 0 {
			pfun(l, args[pv])
		}
	}
}

// Some targets (GOPROG) might need incdeps without
// actually setting cc
func (l *LinkDesc) FinalizeIncdeps(ops *GlobalOps) {
	tname := l.TargetName
	islib := l.TargetOptions["lib"]
	l.IncdepsGenerated = append(l.IncdepsGenerated, l.ResolveIncdeps(ops)...)
	l.IncdepsGenerated = append(l.IncdepsGenerated, l.ResolveSrcs(ops, tname, l.Incdeps...)...)
	if islib {
		l.IncdepsGenerated = append(l.IncdepsGenerated, "$builddir/depend_includes_"+tname)
	}
}

func (l *LinkDesc) FinalizeCC(ops *GlobalOps) {
	if len(l.Objs) > 0 {
		l.FinalizeIncdeps(ops)
		ops.VersionChecks["cc"] = ops.FindCompilerCC
	}
}

func (l *LinkDesc) FinalizeAnalyse(ops *GlobalOps) {
	objs := l.SuffixedObjs(".analyse", func(base string) bool { return !l.DontAnalyse[base] })

	if len(objs) > 0 {
		tname := l.TargetName
		srcdir := l.Srcdir

		l.AddTarget(tname+".target_analyze", "copy_analyse", objs, "obj", "", nil, nil)

		an := &Analyser{
			TargetName:     path.Join(srcdir, tname, tname+".target_analyze"),
			OnlyForFlavors: l.OnlyForFlavors,
		}
		ops.Analyses = append(ops.Analyses, an)
	}
}

func (l *LinkDesc) FinalizeGoSrcs(ops *GlobalOps, mode string) []string {
	if len(l.GoSrc) == 0 {
		return nil
	}

	l.AddTarget("go/main.go", "goaddmain", []string{"$buildtooldir/main.go"}, "objdir", "", nil, nil)
	l.AddTarget("go", "phony", []string{"go/main.go"}, "objdir", "", nil, nil)

	eas := []string{"depfile=$objdir/go/depfile", "gomode=" + mode}
	opts := map[string]bool{"incdeps": true}
	if mode == "lib" {
		objs, llibs := ops.ResolveLibsOurStaticAsLib(l.Libs)
		elibs := ops.ResolveLibsExternal(l.Libs)
		objs = append(objs, llibs...)
		objs = append(objs, elibs...)
		eas = append(eas, "ldlibs="+strings.Join(objs, " "))
		opts["libdeps"] = true
	} else {
		objs, llibs := ops.ResolveLibsOurPicAsLib(l.Libs)
		elibs := ops.ResolveLibsExternal(l.Libs)
		objs = append(objs, llibs...)
		objs = append(objs, elibs...)
		eas = append(eas, "ldlibs="+strings.Join(objs, " "))
		opts["piclibdeps"] = true
	}
	lib := l.AddTarget("gosrc.a", "gobuildlib", []string{"$objdir/go"}, "objdir", "", eas, opts)
	lib.Deps = append(lib.Deps, l.GoSrc...)
	lib.IncdepsExcept["$objdir/gosrc.h"] = true

	l.AddTarget("gosrc.h", "phony", []string{"$objdir/gosrc.a"}, "objdir", "", nil, nil)
	l.IncdepsGenerated = append(l.IncdepsGenerated, "$objdir/gosrc.h")
	return []string{"$objdir/gosrc.a"}
}

func (l *LinkDesc) SuffixedObjs(suff string, filter func(base string) bool) []string {
	objs := make([]string, 0, len(l.Objs))
	for _, o := range l.Objs {
		if filter == nil || filter(o) {
			objs = append(objs, o+suff)
		}
	}
	return objs
}

func (l *LinkDesc) ResolveIncdeps(ops *GlobalOps) []string {
	var ret []string
	for _, libname := range l.Libs {
		if lib := ops.Libs[libname]; lib != nil {
			ret = append(ret, "$builddir/depend_includes_"+libname)
		}
	}
	return ret
}

func (l *LinkDesc) ResolveDeps(ops *GlobalOps, tname string) []string {
	ret := l.GeneralDesc.ResolveDeps(ops, tname)
	if l.Targets[tname].Options["libdeps"] {
		ret = append(ret, ops.ResolveLibsOurStatic(l.Libs)...)
	}
	if l.Targets[tname].Options["piclibdeps"] {
		ret = append(ret, ops.ResolveLibsOurPic(l.Libs)...)
	}
	return ret
}

func (l *LinkDesc) ResolveOrderDeps(target *Target) []string {
	deps := l.GeneralDesc.ResolveOrderDeps(target)

	if target.Options["incdeps"] {
		for _, d := range l.IncdepsGenerated {
			if !target.IncdepsExcept[d] {
				deps = append(deps, d)
			}
		}
	}
	return deps
}

func (l *LinkDesc) OutputHeader(w io.Writer, objdir string) {
	l.GeneralDesc.OutputHeader(w, objdir)
	if len(l.Objs) > 0 {
		fmt.Fprintf(w, "includes =")
		for _, inc := range l.Incdirs {
			fmt.Fprintf(w, " -I %s", inc)
		}
		fmt.Fprintf(w, " -I $objdir\n")
	}
}
