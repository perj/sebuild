// Copyright 2018 Schibsted

package buildbuild

import (
	"path"
	"strings"
)

type LibDescriptor interface {
	IsDummyLib() bool
	LibName() string
	PiclibName() string
	NameAsLib() (name string, islib bool)
	NameAsPiclib() (name string, islib bool)
	LibDeps() []string
}

type LibDesc struct {
	LinkDesc
	LinkSet bool
}

func (tmpl *LibDesc) NewFromTemplate(bd, tname string, flavors []string) Descriptor {
	return &LibDesc{
		LinkDesc: *tmpl.LinkDesc.NewFromTemplate(bd, tname, flavors),
		LinkSet:  tmpl.LinkSet,
	}
}

func (l *LibDesc) Parse(ops *GlobalOps, realsrcdir string, args map[string][]string) Descriptor {
	desc := l.GenericParse(l, ops, realsrcdir, args, LinkerExtra("includes", "incprefix"))
	l.LinkerParse(realsrcdir, args)

	l.Includes = append(l.Includes, args["includes"]...)

	destInc := "dest_inc"
	if len(args["incprefix"]) > 0 {
		destInc = path.Join("dest_inc", args["incprefix"][0])
	}

	for _, inc := range args["includes"] {
		l.AddTarget(inc, "install_header", []string{inc}, destInc, l.Srcdir, nil, nil)
	}

	if ops.Libs == nil {
		ops.Libs = make(map[string]LibDescriptor)
	}
	ops.Libs[l.TargetName] = l
	return desc
}

func (l *LibDesc) Finalize(ops *GlobalOps) {
	libname := l.TargetName
	l.FinalizeCC(ops)

	// We can have dummy libs without any sources (for include dependencies).
	if l.LinkSet || len(l.Objs) > 0 {
		objs := l.SuffixedObjs(".o", nil)
		picobjs := l.SuffixedObjs(".pic_o", nil)

		if l.LinkSet {
			l.AddTarget("lib"+libname+".o", "partiallink", objs, l.Destdir, "", nil, l.TargetOptions)
			l.AddTarget("lib"+libname+"_pic.o", "partiallink", picobjs, l.Destdir, "", nil, l.TargetOptions)
		} else {
			l.AddTarget("lib"+libname+".a", "ar", objs, l.Destdir, "", nil, l.TargetOptions)
			l.AddTarget("lib"+libname+"_pic.a", "ar", picobjs, l.Destdir, "", nil, l.TargetOptions)
		}
	}

	l.Deps["depend_includes_"+libname] = l.ResolveIncdeps(ops)
	l.AddTarget("depend_includes_"+libname, "phony", l.Includes, "builddir", l.Srcdir, nil, nil)

	l.FinalizeAnalyse(ops)
	l.GeneralDesc.Finalize(ops)
}

func (l *LibDesc) IsDummyLib() bool {
	return len(l.Objs) == 0
}

func (l *LibDesc) LibName() string {
	if l.LinkSet {
		return "lib" + l.TargetName + ".o"
	}
	return "lib" + l.TargetName + ".a"
}

func (l *LibDesc) PiclibName() string {
	if l.LinkSet {
		return "lib" + l.TargetName + "_pic.o"
	}
	return "lib" + l.TargetName + "_pic.a"
}

func (l *LibDesc) NameAsLib() (string, bool) {
	if l.LinkSet {
		return l.LibName(), false
	}
	return "-l" + l.TargetName, true
}

func (l *LibDesc) NameAsPiclib() (string, bool) {
	if l.LinkSet {
		return l.PiclibName(), false
	}
	return "-l" + l.TargetName + "_pic", true
}

func (l *LibDesc) LibDeps() []string {
	return l.Libs
}

var LibTemplate = LibDesc{
	LinkDesc: LinkDesc{
		GeneralDesc: GeneralDesc{
			Destdir:       "dest_lib",
			TargetOptions: map[string]bool{"lib": true},
		},
		Picrules: true,
	},
	LinkSet: false,
}

var LinkersetLibTemplate = LibDesc{
	LinkDesc: LinkDesc{
		GeneralDesc: GeneralDesc{
			Destdir:       "dest_lib",
			TargetOptions: map[string]bool{"lib": true},
		},
		Picrules: true,
	},
	LinkSet: true,
}

// Topological sort of library dependencies. Implemented as
// a depth-first descent of the dependencies. The return value
// is in dependency order. This is only really necessary for
// resolving link order static libraries, but it also allows
// us to generate a deduplicated list of depedencies.
func (ops *GlobalOps) ResolveLibs(libs []string) []string {
	var ret []string
	resolved := make(map[string]bool)
	var resolver func(libs []string)

	// Recursive resolver function
	resolver = func(libs []string) {
		for _, l := range libs {
			if resolved[l] {
				continue
			}
			resolved[l] = true
			if desc := ops.Libs[l]; desc != nil {
				resolver(desc.LibDeps())
			}
			ret = append(ret, l)
		}
	}
	resolver(libs)
	return ret
}

// Resolve non-dummy libs that we build.
// Dummy libs are libs that don't have objs.
func (ops *GlobalOps) ResolveLibsOur(libs []string) []LibDescriptor {
	var ret []LibDescriptor

	libs = ops.ResolveLibs(libs)
	for i := len(libs) - 1; i >= 0; i-- {
		if desc := ops.Libs[libs[i]]; desc != nil && !desc.IsDummyLib() {
			ret = append(ret, desc)
		}
	}
	return ret
}

// Resolves all external libs
func (ops *GlobalOps) ResolveLibsExternal(libs []string) []string {
	var ret []string
	libs = ops.ResolveLibs(libs)
	for i := len(libs) - 1; i >= 0; i-- {
		if ops.Libs[libs[i]] == nil {
			if strings.ContainsAny(libs[i], "/$") {
				ret = append(ret, libs[i])
			} else {
				ret = append(ret, "-l"+libs[i])
			}
		}
	}
	return ret
}

// The following three functions resolve arguments to link with various libraries.

func (ops *GlobalOps) ResolveLibsOurStatic(libs []string) []string {
	var ret []string
	for _, lib := range ops.ResolveLibsOur(libs) {
		ret = append(ret, "$libdir/"+lib.LibName())
	}
	return ret
}

func (ops *GlobalOps) ResolveLibsOurPic(libs []string) []string {
	var ret []string
	for _, lib := range ops.ResolveLibsOur(libs) {
		ret = append(ret, "$libdir/"+lib.PiclibName())
	}
	return ret
}

func (ops *GlobalOps) ResolveLibsOurStaticAsLib(libs []string) ([]string, []string) {
	var objs, llibs []string
	for _, lib := range ops.ResolveLibsOur(libs) {
		name, islib := lib.NameAsLib()
		if islib {
			llibs = append(llibs, name)
		} else {
			objs = append(objs, "$libdir/"+name)
		}
	}
	return objs, llibs
}

func (ops *GlobalOps) ResolveLibsOurPicAsLib(libs []string) ([]string, []string) {
	var objs, llibs []string
	for _, lib := range ops.ResolveLibsOur(libs) {
		name, islib := lib.NameAsPiclib()
		if islib {
			llibs = append(llibs, name)
		} else {
			objs = append(objs, "$libdir/"+name)
		}
	}
	return objs, llibs
}
