// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func checkTargets(t *testing.T, expt, got map[string]*Target) {
	if len(expt) != len(got) {
		var gk []string
		for k := range got {
			gk = append(gk, k)
		}
		t.Errorf("len mismatch, %v != %v, got keys %v", len(expt), len(got), gk)
	}
	for k, tgt := range expt {
		if tgt.IncdepsExcept == nil {
			tgt.IncdepsExcept = make(map[string]bool)
		}
		if !reflect.DeepEqual(got[k], tgt) {
			t.Errorf("Target %s didn't match expected", k)
			t.Errorf("wanted: %#v", tgt)
			t.Errorf("got   : %#v", got[k])
		}
	}
}

func TestCompileC(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileC("testdir", "src.c", "src")
	expt := map[string]*Target{
		"src.o": {
			Rule:    "cc",
			Sources: []string{"src.c"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
		"src.analyse": {
			Rule:    "cc_analyse",
			Sources: []string{"src.c"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if len(desc.Objs) == 0 {
		t.Error("No objects")
	}
}

func TestCompileCPic(t *testing.T) {
	desc := (&ProgDesc{LinkDesc: LinkDesc{Picrules: true}}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileC("testdir", "src.c", "src")
	expt := map[string]*Target{
		"src.o": {
			Rule:    "cc",
			Sources: []string{"src.c"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
		"src.pic_o": {
			Rule:      "cc",
			Sources:   []string{"src.c"},
			Destdir:   "obj",
			Extraargs: []string{"picflag=-fPIC"},
			Options:   map[string]bool{"incdeps": true},
		},
		"src.analyse": {
			Rule:    "cc_analyse",
			Sources: []string{"src.c"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
	}
	checkTargets(t, expt, desc.AllTargets())
}

func TestCompileCXXPic(t *testing.T) {
	desc := (&ProgDesc{LinkDesc: LinkDesc{Picrules: true}}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileCXX("testdir", "src.cc", "src")
	expt := map[string]*Target{
		"src.o": {
			Rule:    "cxx",
			Sources: []string{"src.cc"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
		"src.pic_o": {
			Rule:      "cxx",
			Sources:   []string{"src.cc"},
			Destdir:   "obj",
			Extraargs: []string{"picflag=-fPIC"},
			Options:   map[string]bool{"incdeps": true},
		},
		"src.analyse": {
			Rule:    "cxx_analyse",
			Sources: []string{"src.cc"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if desc.Link != "linkxx" {
		t.Error("linker wasn't changed")
	}
	if len(desc.Objs) == 0 {
		t.Error("No objects")
	}
}

func TestCompileNoAnalyse(t *testing.T) {
	desc := (&ProgDesc{LinkDesc: LinkDesc{NoAnalyse: true}}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileC("testdir", "src.c", "src")
	desc.CompileCXX("testdir", "src2.cc", "src2")
	expt := map[string]*Target{
		"src.o": {
			Rule:    "cc",
			Sources: []string{"src.c"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
		"src2.o": {
			Rule:    "cxx",
			Sources: []string{"src2.cc"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if !desc.DontAnalyse["src"] {
		t.Error("DontAnalyse wasn't set")
	}
	if !desc.DontAnalyse["src2"] {
		t.Error("DontAnalyse wasn't set")
	}
}

func TestCompileYY(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileYY("testdir", "src.yy", "src")
	expt := map[string]*Target{
		"src.cc": {
			Rule:    "yaccxx",
			Sources: []string{"src.yy"},
			Destdir: "obj",
		},
		"src.hh": {
			Rule:    "phony",
			Sources: []string{"src.cc"},
			Destdir: "obj",
		},
		"src.o": {
			Rule:    "cxx",
			Sources: []string{"src.cc"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
		"src.analyse": {
			Rule:    "cxx_analyse",
			Sources: []string{"src.cc"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if desc.Link != "linkxx" {
		t.Error("linker wasn't changed")
	}
	if !reflect.DeepEqual(desc.Incdeps, []string{"src.hh"}) {
		t.Error("Not added to incdeps")
	}
}

func TestCompileLL(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileLL("testdir", "src.ll", "src")
	expt := map[string]*Target{
		"src.cc": {
			Rule:    "flexx",
			Sources: []string{"src.ll"},
			Destdir: "obj",
		},
		"src.o": {
			Rule:    "cxx",
			Sources: []string{"src.cc"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
		"src.analyse": {
			Rule:    "cxx_analyse",
			Sources: []string{"src.cc"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if desc.Link != "linkxx" {
		t.Error("linker wasn't changed")
	}
}

func TestCompileGperf(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileGperf("testdir", "src.gperf", "src")
	expt := map[string]*Target{
		"src.h": {
			Rule:    "gperf",
			Sources: []string{"src.gperf"},
			Destdir: "obj",
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if !reflect.DeepEqual(desc.Incdeps, []string{"src.h"}) {
		t.Error("Not added to incdeps")
	}
}

func TestCompileEnum(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileEnum("testdir", "src.gperf.enum", "src.gperf")
	expt := map[string]*Target{
		"src.gperf": {
			Rule:    "gperf_enum",
			Sources: []string{"src.gperf.enum"},
			Destdir: "obj",
		},
		"src.h": {
			Rule:    "gperf",
			Sources: []string{"src.gperf"},
			Destdir: "obj",
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if !reflect.DeepEqual(desc.Incdeps, []string{"src.h"}) {
		t.Error("Not added to incdeps")
	}

	defer func() {
		err := recover()
		if err == nil || err.(*ParseError).Err != EnumWithoutGperf {
			t.Error("Didn't get EnumWithoutGperf error")
		}
	}()
	desc.CompileEnum("testdir", "src.enum", "src")
}

func TestCompileIn(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileIn("testdir", "src.in", "src")
	expt := map[string]*Target{
		"src": {
			Rule:    "in",
			Sources: []string{"src.in"},
			Destdir: "obj",
		},
	}
	checkTargets(t, expt, desc.AllTargets())
}

func TestCompileXS(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileXs("testdir", "src.xs", "src")
	expt := map[string]*Target{
		"src.c": {
			Rule:    "xs",
			Sources: []string{"src.xs"},
			Destdir: "obj",
		},
		"src.o": {
			Rule:    "cc",
			Sources: []string{"src.c"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
		"src.analyse": {
			Rule:    "cc_analyse",
			Sources: []string{"src.c"},
			Destdir: "obj",
			Options: map[string]bool{"incdeps": true},
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if len(desc.Objs) == 0 {
		t.Error("No objects")
	}
}

func TestCompileGo(t *testing.T) {
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil).(*ProgDesc)

	desc.CompileGo("testdir", "src.go", "src")
	expt := map[string]*Target{
		"go/src.go": {
			Rule:    "copy_go",
			Sources: []string{"src.go"},
			Destdir: "obj",
		},
	}
	checkTargets(t, expt, desc.AllTargets())
	if !reflect.DeepEqual(desc.GoSrc, []string{"$objdir/go/src.go"}) {
		t.Error("GoSrc wasn't set")
	}
}

func TestFindCompilerCC(t *testing.T) {
	vmap := map[string]string{
		"x":     "nope",
		"y":     "cc version 1.0",
		"gcc":   "gcc version 4.7.0",
		"clang": "clang version 3.3.2 (tags/RELEASE_33/dot2-final)",
		"cc":    "gcc version 4.8.0",
	}
	var order []string
	notFound := errors.New("Not found")
	defer func() {
		findCompilerRun = (*exec.Cmd).Run
	}()
	findCompilerRun = func(cmd *exec.Cmd) error {
		cc := filepath.Base(cmd.Path)
		order = append(order, cc)
		out, ok := vmap[cc]
		if !ok {
			return notFound
		}
		io.WriteString(cmd.Stdout, out)
		return nil
	}
	ops := NewGlobalOps()
	ops.Config.Compiler = []string{"cc:9.0", "x", "y:2"}
	os.Setenv("CC", "z")
	err := ops.FindCompilerCC()
	if err != nil {
		t.Error("Expected no error, got", err)
	}
	if !reflect.DeepEqual(order, []string{"cc", "x", "y", "z", "gcc", "clang", "cc"}) {
		t.Error("Tested compiler order didn't match expected, got", order)
	}
}

func TestCompileSpecial(t *testing.T) {
	called := false
	PluginSpecialSrcs["test"] = func(desc Descriptor, tname, rule string, srcs []string, destdir, srcdir string, extraargs []string, options map[string]bool) Descriptor {
		called = true
		return desc
	}
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil)
	desc = CompileSpecial(desc, "testtarget", "test", []string{"src"}, "obj", "testdir", nil, nil)
	if !called {
		t.Error("Expected plugin to be called")
	}
	desc = CompileSpecial(desc, "testtarget2", "test2", []string{"src2"}, "obj", "testdir", nil, nil)
	delete(PluginSpecialSrcs, "test")
	expt := map[string]*Target{
		"testtarget2": {
			Rule:    "test2",
			Sources: []string{"src2"},
			Destdir: "obj",
		},
	}
	checkTargets(t, expt, desc.AllTargets())
}

func TestCompileSrc(t *testing.T) {
	called := false
	PluginGeneralExtensions["test"] = func(g *GeneralDesc, srcdir, src, srcbase string) {
		if srcbase != "src" {
			t.Error("srcbase != src")
		}
		called = true
	}
	desc := (&ProgDesc{}).NewFromTemplate("Builddesc", "test", nil)
	CompileSrc(desc, "testdir", "src.test")
	if !called {
		t.Error("Expected plugin to be called")
	}
	CompileSrc(desc, "testdir", "src.in")
	delete(PluginGeneralExtensions, "test")
	expt := map[string]*Target{
		"src": {
			Rule:    "in",
			Sources: []string{"src.in"},
			Destdir: "obj",
		},
	}
	checkTargets(t, expt, desc.AllTargets())
}
