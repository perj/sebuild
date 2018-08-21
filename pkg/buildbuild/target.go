// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"path"
	"strings"
)

// Targets we want to generate.
type Target struct {
	Rule      string   // Rule to generate the target.
	Sources   []string // Source files, path unresolved.
	Deps      []string
	Srcopts   []string
	Destdir   string          // Destination directory, might be an alias defined in DestLookup
	Extraargs []string        // Various extra arguments like "cflags = -O0"
	Options   map[string]bool // Various output options, notably "all" makes this a default target.
	CollectAs string

	IncdepsExcept map[string]bool
}

var (
	MultipleDefinedTarget = errors.New("Multiple defined target. Rename your generated intermediate files.")
)

func (desc *GeneralDesc) AddTarget(tname, rule string, srcs []string, destdir, srcdir string, extraargs []string, options map[string]bool) *Target {
	var srcopts []string
	for _, src := range srcs {
		if opt := desc.Srcopts[src]; len(opt) > 0 {
			srcopts = append(srcopts, opt...)
		}
	}

	if desc.Targets[tname] != nil {
		// First we have to make sure that no target has already been renamed according
		// to the rules below. If there was and we have three targets with the same
		// name we can't resolve that. Rename your intermediate files.
		tmpname := "TMP_BUILD" + tname
		if desc.Targets[tmpname] != nil {
			panic(&ParseError{MultipleDefinedTarget, tname, desc.Builddesc})
		}

		if strings.HasPrefix(rule, "install") {
			// If our rule is install, we need to rename the old target so that
			// all dependencies are on us.
			for idx, src := range srcs {
				if src == tname {
					srcs[idx] = tmpname
				}
			}

			desc.Targets[tmpname] = desc.Targets[tname]
			delete(desc.Targets, tname)
			desc.Srcdirs[tmpname] = desc.Srcdirs[tname]
			delete(desc.Srcdirs, tname)
			desc.Deps[tmpname] = desc.Deps[tname]
			delete(desc.Deps, tname)
			desc.Srcopts[tmpname] = desc.Srcopts[tname]
			delete(desc.Srcopts, tname)
		} else if strings.HasPrefix(desc.Targets[tname].Rule, "install") {
			// if the already defined rule is install, we can change our
			// name since no one will depend on us having the right name.
			for idx, src := range desc.Targets[tname].Sources {
				if src == tname {
					desc.Targets[tname].Sources[idx] = tmpname
				}
			}
			tname = tmpname
		} else {
			// If neither of the targets are install, we can't handle it.
			panic(&ParseError{MultipleDefinedTarget, tname, desc.Builddesc})
		}
	}

	target := &Target{
		Rule:          rule,
		Sources:       srcs,
		Srcopts:       srcopts,
		Extraargs:     extraargs,
		Destdir:       destdir,
		Options:       options,
		IncdepsExcept: make(map[string]bool),
	}
	desc.Targets[tname] = target
	if desc.Srcdirs == nil {
		desc.Srcdirs = make(map[string]string)
	}
	for _, src := range srcs {
		desc.Srcdirs[src] = srcdir
	}
	return target
}

func (g *GeneralDesc) AllTargets() map[string]*Target {
	return g.Targets
}

var DestLookup = map[string]string{
	"obj":        "$objdir/",
	"objdir":     "$objdir/",
	"dest_inc":   "$incdir/",
	"dest_bin":   "$dest_bin/",
	"dest_tool":  "$buildtools/",
	"dest_lib":   "$libdir/",
	"dest_mod":   "$dest_mod/",
	"destroot":   "$destroot/",
	"builddir":   "$builddir/",
	"flavorroot": "$flavorroot/",
}

func (target *Target) ResolveDest() string {
	dest := target.Destdir

	parts := strings.SplitN(dest, "/", 2)
	first := DestLookup[parts[0]]
	if first != "" {
		parts[0] = first
	} else {
		parts = append([]string{"$destroot"}, parts...)
	}
	return path.Join(parts...)
}
