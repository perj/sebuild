// Copyright 2018 Schibsted

package buildbuild

import "strings"

type ProgDesc struct {
	LinkDesc
}

func (tmpl *ProgDesc) NewFromTemplate(bd, tname string, flavors []string) Descriptor {
	return &ProgDesc{
		LinkDesc: *tmpl.LinkDesc.NewFromTemplate(bd, tname, flavors),
	}
}

func (p *ProgDesc) Parse(ops *GlobalOps, realsrcdir string, args map[string][]string) Descriptor {
	desc := p.GenericParse(p, ops, realsrcdir, args, LinkerExtra())
	p.LinkerParse(realsrcdir, args)
	return desc
}

func (p *ProgDesc) Finalize(ops *GlobalOps) {
	p.FinalizeCC(ops)

	prog := p.TargetName
	objs := p.SuffixedObjs(".o", nil)

	goobj := p.FinalizeGoSrcs(ops, "lib")
	objs = append(objs, goobj...)
	objs = append(objs, ops.ResolveLibsOurStatic(p.Libs)...)

	ldlibs := ops.ResolveLibsExternal(p.Libs)
	link := ops.ResolveLibsLinker(p.Link, p.Libs)
	eas := []string{"ldlibs=" + strings.Join(ldlibs, " ")}
	p.AddTarget(prog, link, objs, p.Destdir, "", eas, p.TargetOptions)

	p.FinalizeAnalyse(ops)
	p.GeneralDesc.Finalize(ops)
}

var ProgTemplate = ProgDesc{
	LinkDesc: LinkDesc{
		GeneralDesc: GeneralDesc{
			Destdir:       "dest_bin",
			TargetOptions: map[string]bool{"all": true},
		},
		Picrules: false,
		Link:     "link",
	},
}

var ToolProgTemplate = ProgDesc{
	LinkDesc: LinkDesc{
		GeneralDesc: GeneralDesc{
			Destdir: "dest_tool",
		},
		Picrules: false,
		Link:     "link",
	},
}
