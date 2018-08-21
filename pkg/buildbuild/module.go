// Copyright 2018 Schibsted

package buildbuild

import "strings"

type ModuleDesc struct {
	LinkDesc
}

func (tmpl *ModuleDesc) NewFromTemplate(bd, tname string, flavors []string) Descriptor {
	return &ModuleDesc{
		LinkDesc: *tmpl.LinkDesc.NewFromTemplate(bd, tname, flavors),
	}
}

func (m *ModuleDesc) Parse(ops *GlobalOps, realsrcdir string, args map[string][]string) Descriptor {
	desc := m.GenericParse(m, ops, realsrcdir, args, LinkerExtra())
	m.LinkerParse(realsrcdir, args)
	return desc
}

func (m *ModuleDesc) Finalize(ops *GlobalOps) {
	m.FinalizeCC(ops)

	mod := m.TargetName + ".so"
	objs := m.SuffixedObjs(".pic_o", nil)

	goobj := m.FinalizeGoSrcs(ops, "piclib")
	objs = append(objs, goobj...)
	objs = append(objs, ops.ResolveLibsOurPic(m.Libs)...)

	ldlibs := ops.ResolveLibsExternal(m.Libs)
	eas := []string{"ldflags=-rdynamic -fPIC -shared", "ldlibs=" + strings.Join(ldlibs, " ")}
	m.AddTarget(mod, m.Link, objs, m.Destdir, "", eas, m.TargetOptions)

	m.FinalizeAnalyse(ops)
	m.GeneralDesc.Finalize(ops)
}

var ModuleTemplate = ModuleDesc{
	LinkDesc: LinkDesc{
		GeneralDesc: GeneralDesc{
			Destdir:       "dest_mod",
			TargetOptions: map[string]bool{"all": true},
		},
		Picrules: true,
		Link:     "link",
	},
}
