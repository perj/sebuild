// Copyright 2018 Schibsted

package buildbuild

import "strings"

type GoProgDesc struct {
	LinkDesc
	Pkg    string
	NoCgo  bool
	GOOS   string
	GOARCH string
}

type GoTestDesc struct {
	LinkDesc
	Pkg        string
	Benchflags string
}

func (tmpl *GoProgDesc) NewFromTemplate(bd, tname string, flavors []string) Descriptor {
	return &GoProgDesc{
		LinkDesc: *tmpl.LinkDesc.NewFromTemplate(bd, tname, flavors),
	}
}

func (tmpl *GoTestDesc) NewFromTemplate(bd, tname string, flavors []string) Descriptor {
	return &GoTestDesc{
		LinkDesc: *tmpl.LinkDesc.NewFromTemplate(bd, tname, flavors),
	}
}

func (g *GoProgDesc) Parse(ops *GlobalOps, realsrcdir string, args map[string][]string) Descriptor {
	desc := g.GenericParse(g, ops, realsrcdir, args, LinkerExtra("gopkg", "nocgo", "goos", "goarch"))
	g.LinkerParse(realsrcdir, args)
	g.Pkg = strings.Join(args["gopkg"], " ")
	g.NoCgo = args["nocgo"] != nil
	g.GOOS = strings.Join(args["goos"], " ")
	g.GOARCH = strings.Join(args["goarch"], " ")
	return desc
}

func AddGodeps(t *Target, ops *GlobalOps) {
	if len(ops.Config.Godeps) > 0 {
		t.Deps = append(t.Deps, ops.GodepsStamp())
	}
}

func (g *GoProgDesc) Finalize(ops *GlobalOps) {
	g.FinalizeIncdeps(ops)

	objs, llibs := ops.ResolveLibsOurStaticAsLib(g.Libs)
	elibs := ops.ResolveLibsExternal(g.Libs)
	objs = append(objs, llibs...)
	objs = append(objs, elibs...)
	eas := []string{"ldlibs=" + strings.Join(objs, " ")}
	if g.Pkg != "" {
		eas = append(eas, "gopkg="+g.Pkg)
	}
	if g.NoCgo {
		eas = append(eas, "gomode=prog-nocgo")
	} else {
		eas = append(eas, "gomode=prog")
	}
	if g.GOOS != "" {
		eas = append(eas, "goos="+g.GOOS)
	}
	if g.GOARCH != "" {
		eas = append(eas, "goarch="+g.GOARCH)
	}

	target := g.AddTarget(g.TargetName, "gobuild", []string{g.Srcdir}, g.Destdir, "", eas, g.TargetOptions)
	AddGodeps(target, ops)
	g.GeneralDesc.Finalize(ops)
}

func (g *GoTestDesc) Parse(ops *GlobalOps, realsrcdir string, args map[string][]string) Descriptor {
	desc := g.GenericParse(g, ops, realsrcdir, args, LinkerExtra("gopkg", "benchflags"))
	g.LinkerParse(realsrcdir, args)
	g.Pkg = strings.Join(args["gopkg"], " ")
	g.Benchflags = strings.Join(args["benchflags"], " ")
	return desc
}

func (g *GoTestDesc) Finalize(ops *GlobalOps) {
	name := g.TargetName
	g.FinalizeIncdeps(ops)

	objs, llibs := ops.ResolveLibsOurStaticAsLib(g.Libs)
	elibs := ops.ResolveLibsExternal(g.Libs)
	objs = append(objs, llibs...)
	objs = append(objs, elibs...)
	eas := []string{"ldlibs=" + strings.Join(objs, " ")}
	if g.Pkg != "" {
		eas = append(eas, "gopkg="+g.Pkg)
	}

	eas = append(eas, "gomode=test-prog")
	target := g.AddTarget(name+".test", "gobuild", []string{g.Srcdir}, g.Destdir, "", eas, g.TargetOptions)
	eas = eas[:len(eas)-1]
	AddGodeps(target, ops)

	opts := map[string]bool{"incdeps": true, "libdeps": true}
	target = g.AddTarget("gotest/"+name, "gotest", []string{g.Srcdir}, "destroot", "", eas, opts)
	AddGodeps(target, ops)
	target.CollectAs = "_gotest"

	target = g.AddTarget("gocover/"+name+"-coverage", "gocover", []string{g.Srcdir}, "destroot", "", eas, opts)
	AddGodeps(target, ops)

	target = g.AddTarget("gocover/"+name+"-coverage.html", "gocover_html", []string{"gocover/" + name + "-coverage"}, "destroot", "", eas, nil)
	target.CollectAs = "_gocover"

	if g.Benchflags != "" {
		eas = append(eas, "benchflags="+g.Benchflags)
	} else {
		eas = append(eas, "benchflags=.")
	}
	target = g.AddTarget("gobench/"+name, "gobench", []string{g.Srcdir}, "destroot", "", eas, opts)
	AddGodeps(target, ops)
	target.CollectAs = "_gobench"

	g.GeneralDesc.Finalize(ops)
}

var GoprogTemplate = GoProgDesc{
	LinkDesc: LinkDesc{
		GeneralDesc: GeneralDesc{
			Destdir:       "dest_bin",
			TargetOptions: map[string]bool{"all": true, "incdeps": true, "libdeps": true},
		},
	},
}

var GotestTemplate = GoTestDesc{
	LinkDesc: LinkDesc{
		GeneralDesc: GeneralDesc{
			Destdir:       "gotest",
			TargetOptions: map[string]bool{"all": true, "incdeps": true, "libdeps": true},
		},
	},
}
