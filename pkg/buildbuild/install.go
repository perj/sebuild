// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"strings"
)

var (
	BadSymlinkFormat = errors.New("Bad symlink format, need symlink[dst:target]")
)

var InstallCommands = map[string]string{
	"conf":    "install_conf",   // Installs the file with 644 permissions, intended for configuration files.
	"scripts": "install_script", // Installs the file with 755 permissions, intended for executable scripts.
	"python":  "install_py",     // Runs the py_compile module on the source file(s) during installation.
	"php":     "install_php",    // Runs the php linter on the file during installation and fails if the file doesn't pass linting.
}

type InstallDesc struct {
	GeneralDesc

	Installs map[string][]string
	Symlinks [][2]string
}

func (tmpl *InstallDesc) NewFromTemplate(bd, tname string, flavors []string) Descriptor {
	return &InstallDesc{
		GeneralDesc: *tmpl.GeneralDesc.NewFromTemplate(bd, tname, flavors),
		Installs:    make(map[string][]string),
	}
}

func (id *InstallDesc) Parse(ops *GlobalOps, realsrcdir string, args map[string][]string) Descriptor {
	extra := []string{"symlink"}
	for inst := range InstallCommands {
		if len(args[inst]) > 0 {
			id.Installs[inst] = append(id.Installs[inst], ops.GlobDir(realsrcdir, args[inst])...)
		}
		extra = append(extra, inst)
	}

	desc := id.GenericParse(id, ops, realsrcdir, args, extra)

	// Symlinks the tgt to the src, given as tgt:src in the arguments.
	// Any relative path should be from the installation directory.
	for _, sym := range args["symlink"] {
		syms := strings.SplitN(sym, ":", 2)
		if len(syms) < 2 {
			panic(&ParseError{BadSymlinkFormat, sym, id.Builddesc})
		}
		id.Symlinks = append(id.Symlinks, [2]string{syms[0], syms[1]})
	}
	return desc
}

func (id *InstallDesc) Finalize(ops *GlobalOps) {
	destdir := id.Destdir
	if destdir == "" {
		destdir = id.TargetName
	}
	opts := map[string]bool{"all": true}
	for inst, command := range InstallCommands {
		for _, src := range id.Installs[inst] {
			id.AddTarget(src, command, []string{src}, destdir, id.Srcdir, nil, opts)
		}
	}

	opts = map[string]bool{"all": true, "emptysrcs": true}
	for _, syms := range id.Symlinks {
		id.AddTarget(syms[0], "symlink", nil, destdir, destdir, []string{"target=" + syms[1]}, opts)
	}

	id.GeneralDesc.Finalize(ops)
}

var ToolInstallTemplate = InstallDesc{
	GeneralDesc: GeneralDesc{
		Destdir: "dest_tool",
	},
}

var InstallTemplate = InstallDesc{}
