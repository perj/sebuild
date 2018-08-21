// Copyright 2018 Schibsted

package collect_tests

import (
	"strings"

	"github.com/schibsted/sebuild/pkg/buildbuild"
)

type CollectTests struct {
	Ops *buildbuild.GlobalOps
}

func init() {
	buildbuild.InitPlugin("collect_tests", &CollectTests{})
}

func (plug *CollectTests) Startup(ops *buildbuild.GlobalOps) error {
	plug.Ops = ops
	buildbuild.PluginSpecialSrcs["collect_go_tests"] = plug.CompileCollectGoTests
	return nil
}

func (plug *CollectTests) CompileCollectGoTests(desc buildbuild.Descriptor, tname, rule string, srcs []string, destdir, srcdir string, extraargs []string, options map[string]bool) buildbuild.Descriptor {
	return &CollectDesc{desc, tname, srcs, destdir, srcdir, extraargs}
}

type CollectDesc struct {
	buildbuild.Descriptor

	tname           string
	srcs            []string
	destdir, srcdir string
	extraargs       []string
}

func (c *CollectDesc) Finalize(ops *buildbuild.GlobalOps) {
	c.Descriptor.Finalize(ops)

	var gotests []string
	for _, src := range c.srcs {
		gotests = append(gotests, ops.ResolveCollectedVar(src))
	}
	eas := append(c.extraargs, "gotest_list = "+strings.Join(gotests, " "))
	opts := map[string]bool{"emptysrcs": true}
	c.AddTarget(c.tname, "gotest_to_rdep", nil, c.destdir, c.srcdir, eas, opts)
}
