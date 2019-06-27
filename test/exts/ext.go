// Copyright 2018 Schibsted

package main

import (
	"strings"

	"github.com/schibsted/sebuild/pkg/buildbuild"
)

type Ext struct {
	ops *buildbuild.GlobalOps
}

func init() {
	buildbuild.InitPlugin("exts", &Ext{})
}

func (ext *Ext) Startup(ops *buildbuild.GlobalOps) error {
	ext.ops = ops

	buildbuild.PluginGeneralExtensions["foo"] = ext.CompileFoo
	buildbuild.PluginDescriptors["FOO"] = &FooDesc{
		GeneralDesc: buildbuild.GeneralDesc{
			Destdir: "dest_lib",
		},
	}
	ops.Config.Rules = append(ops.Config.Rules, "foo.ninja")
	return nil
}

func (ext *Ext) CompileFoo(g *buildbuild.GeneralDesc, srcdir, src, srcbase string) {
	g.AddTarget(srcbase+".bar", "foo-cp", []string{src}, "obj", srcdir, nil, nil)
}

type FooDesc struct {
	buildbuild.GeneralDesc
}

func (ft *FooDesc) NewFromTemplate(bd, tname string, flavors []string) buildbuild.Descriptor {
	return &FooDesc{
		GeneralDesc: *ft.GeneralDesc.NewFromTemplate(bd, tname, flavors),
	}
}

func (f *FooDesc) Parse(ops *buildbuild.GlobalOps, realsrcdir string, args map[string][]string) buildbuild.Descriptor {
	desc := f.GenericParse(f, ops, realsrcdir, args, []string{"foos"})

	// Synonym to srcs, except no glob because lazy.
	var objs []string
	for _, foo := range args["foos"] {
		buildbuild.CompileSrc(desc, f.Srcdir, foo)
		objs = append(objs, strings.TrimSuffix(foo, ".foo")+".bar")
	}

	opts := map[string]bool{"all": true}
	desc.AddTarget(f.TargetName, "concat", objs, f.Destdir, f.Srcdir, nil, opts)
	return desc
}

func (f *FooDesc) Finalize(ops *buildbuild.GlobalOps) {
	f.GeneralDesc.Finalize(ops)
	println("Finalizing foo")
}
