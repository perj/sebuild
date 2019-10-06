// Copyright 2018 Schibsted

package buildbuild

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

type CompileGeneralSrcFunc func(g *GeneralDesc, srcdir, src, srcbase string)
type CompileLinkerSrcFunc func(l *LinkDesc, srcdir, src, srcbase string)
type CompileSpecialSrcFunc func(desc Descriptor, tname, rule string, srcs []string, destdir, srcdir string, extraargs []string, options map[string]bool) Descriptor

var (
	UnknownSourceExtension = errors.New("Invalid source extension (not all extensions can be used in all Descriptors)")
	EnumWithoutGperf       = errors.New("Extension .enum without .gperf.enum")
	SourceNeedsLinker      = errors.New("Incompatible source type for this descriptor")
)

func (ops *GlobalOps) DefaultCompiler() {
	ops.CC = "cc"
	ops.CXX = "c++"
	ops.CompilerFlavor = "gcc"
}

// This is like AddTarget, but allow plugins to intercept in case they do some special processing.
// It also splits tname on comma and sets a multi target if more than one element.
func CompileSpecial(desc Descriptor, tname, rule string, srcs []string, destdir, srcdir string, extraargs []string, options map[string]bool) Descriptor {
	if len(srcs) == 1 && srcs[0] == "" {
		if options == nil {
			options = make(map[string]bool)
		}
		options["emptysrcs"] = true
		srcs = nil
	}
	if f := PluginSpecialSrcs[rule]; f != nil {
		return f(desc, tname, rule, srcs, destdir, srcdir, extraargs, options)
	}
	tnames := strings.Split(tname, ",")
	target := desc.AddTarget(tnames[0], rule, srcs, destdir, srcdir, extraargs, options)
	if len(tnames) > 1 {
		desc.AddMultiTarget(tnames, target)
	}
	return desc
}

// Sources in the srcs argument are compiled based on their file extension.
// What extensions are allowed depend on the descriptor. The only extension
// allowed anywhere is .in, which generates a file with the in.pl script.
//
// Descriptors that link a binary allow several more extensions:
// .c, .cc, .cxx, .ll, .yy, .gperf, .gperf.enum, .xs and .go
//
// Plugins might add additional extensions.
//
// The produced objects are put in the object directory of the descriptor,
// where they can be referenced by other arguments or used in the descriptor
// target.
//
// If a source generates a C file or similar, it is typically compiled all
// the way to an object file and then added to the linked objects.
// Other sources, e.g. in files, have to be used manually by other arguments
// or they will be discarded.
func CompileSrc(desc Descriptor, srcdir, src string) {
	ext := path.Ext(src)
	if len(ext) < 1 {
		ext = "."
	}
	srcbase := strings.TrimSuffix(src, ext)
	desc.CompileSrc(srcdir, src, srcbase, ext)
}

func (g *GeneralDesc) CompileSrc(srcdir, src, srcbase, ext string) {
	if f := PluginGeneralExtensions[ext[1:]]; f != nil {
		f(g, srcdir, src, srcbase)
		return
	}
	switch ext[1:] {
	case "in":
		g.CompileIn(srcdir, src, srcbase)
	default:
		panic(&ParseError{UnknownSourceExtension, src, g.Builddesc})
	}
}

func (l *LinkDesc) CompileSrc(srcdir, src, srcbase, ext string) {
	if f := PluginLinkerExtensions[ext[1:]]; f != nil {
		f(l, srcdir, src, srcbase)
		return
	}
	switch ext[1:] {
	case "c":
		l.CompileC(srcdir, src, srcbase)
	case "cc", "cxx":
		l.CompileCXX(srcdir, src, srcbase)
	case "yy":
		l.CompileYY(srcdir, src, srcbase)
	case "ll":
		l.CompileLL(srcdir, src, srcbase)
	case "gperf":
		l.CompileGperf(srcdir, src, srcbase)
	case "enum":
		l.CompileEnum(srcdir, src, srcbase)
	case "xs":
		l.CompileXs(srcdir, src, srcbase)
	case "go":
		l.CompileGo(srcdir, src, srcbase)
	default:
		l.GeneralDesc.CompileSrc(srcdir, src, srcbase, ext)
		return
	}
}

// Redirected by test
var findCompilerRun = (*exec.Cmd).Run

func (ops *GlobalOps) FindCompilerCC() error {
	candidates := ops.Config.Compiler
	envcc := os.Getenv("CC")
	if envcc != "" {
		candidates = append(candidates, envcc)
	}
	candidates = append(candidates, "gcc", "clang", "cc")

	ops.CompilerFlavor = ""

	minGcc := comparableVersion("4.8")
	minClang := comparableVersion("3.4")
	versionRE := regexp.MustCompile(`(\S+) version ([0-9]+\.[0-9]+)`)
	for _, cc := range candidates {
		minv := ""
		if col := strings.IndexRune(cc, ':'); col >= 0 {
			minv = cc[col+1:]
			cc = cc[:col]
		}
		ccv := append(strings.Split(cc, " "), "-v")
		cmd := exec.Command(ccv[0], ccv[1:]...)
		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output
		err := findCompilerRun(cmd)
		if err != nil {
			continue
		}

		match := versionRE.FindStringSubmatch(output.String())
		if match == nil {
			continue
		}

		c := match[1]
		v := comparableVersion(match[2])

		if minv != "" && v < comparableVersion(minv) {
			continue
		}
		if c == "gcc" && v < minGcc {
			continue
		}
		if c == "clang" && v < minClang {
			continue
		}

		ops.CompilerFlavor = c
		ops.CC = cc
		if idx := strings.Index(cc, "gcc"); idx >= 0 {
			ops.CXX = cc[:idx] + "g++" + cc[idx+3:]
		} else if idx := strings.Index(cc, "cc"); idx >= 0 {
			ops.CXX = cc[:idx] + "c++" + cc[idx+2:]
		} else if idx := strings.Index(cc, "-"); idx >= 0 { // clang-4.0 -> clang++-4.0
			ops.CXX = cc[:idx] + "++" + cc[idx:]
		} else {
			ops.CXX = cc + "++"
		}

		if ops.Options.Debug {
			fmt.Printf("Compilers detected: %s / %s (%s %s)\n", ops.CC, ops.CXX, c, v)
		}
		break
	}

	if ops.CompilerFlavor == "" {
		return errors.New("Couldn't find a compatible compiler")
	}
	return nil
}

func comparableVersion(v string) string {
	var vlong strings.Builder
	for _, vpart := range strings.Split(v, ".") {
		fmt.Fprintf(&vlong, "%4s", vpart)
	}
	return vlong.String()
}

func (l *LinkDesc) CompileC(srcdir, src, srcbase string) {
	opts := map[string]bool{"incdeps": true}
	l.AddTarget(srcbase+".o", "cc", []string{src}, "obj", srcdir, nil, opts)
	if l.Picrules {
		l.AddTarget(srcbase+".pic_o", "cc", []string{src}, "obj", srcdir, []string{"picflag=-fPIC"}, opts)
	}

	if l.NoAnalyse {
		l.DontAnalyse[srcbase] = true
	} else {
		l.AddTarget(srcbase+".analyse", "cc_analyse", []string{src}, "obj", srcdir, nil, opts)
	}
	l.Objs = append(l.Objs, srcbase)
}

func (l *LinkDesc) CompileCXX(srcdir, src, srcbase string) {
	l.Link = "linkxx"
	opts := map[string]bool{"incdeps": true}
	l.AddTarget(srcbase+".o", "cxx", []string{src}, "obj", srcdir, nil, opts)
	if l.Picrules {
		l.AddTarget(srcbase+".pic_o", "cxx", []string{src}, "obj", srcdir, []string{"picflag=-fPIC"}, opts)
	}

	if l.NoAnalyse {
		l.DontAnalyse[srcbase] = true
	} else {
		l.AddTarget(srcbase+".analyse", "cxx_analyse", []string{src}, "obj", srcdir, nil, opts)
	}
	l.Objs = append(l.Objs, srcbase)
}

func (l *LinkDesc) CompileYY(srcdir, src, srcbase string) {
	l.Link = "linkxx"

	l.AddTarget(srcbase+".cc", "yaccxx", []string{src}, "obj", srcdir, nil, nil)
	l.AddTarget(srcbase+".hh", "phony", []string{srcbase + ".cc"}, "obj", "", nil, nil)
	l.Incdeps = append(l.Incdeps, srcbase+".hh")

	l.CompileCXX(srcdir, srcbase+".cc", srcbase)
}

func (l *LinkDesc) CompileLL(srcdir, src, srcbase string) {
	l.Link = "linkxx"

	l.AddTarget(srcbase+".cc", "flexx", []string{src}, "obj", srcdir, nil, nil)
	l.CompileCXX(srcdir, srcbase+".cc", srcbase)
}

func (l *LinkDesc) CompileGperf(srcdir, src, srcbase string) {
	l.AddTarget(srcbase+".h", "gperf", []string{src}, "obj", srcdir, nil, nil)
	l.Incdeps = append(l.Incdeps, srcbase+".h")
}

func (l *LinkDesc) CompileEnum(srcdir, src, srcbase string) {
	l.AddTarget(srcbase, "gperf_enum", []string{src}, "obj", srcdir, nil, nil)

	gp := strings.TrimSuffix(srcbase, ".gperf")
	if gp == srcbase {
		panic(&ParseError{EnumWithoutGperf, src, l.Builddesc})
	}
	l.CompileGperf("", srcbase, gp)
}

func (g *GeneralDesc) CompileIn(srcdir, src, srcbase string) {
	g.AddTarget(srcbase, "in", []string{src}, "obj", srcdir, nil, nil)
}

func (l *LinkDesc) CompileXs(srcdir, src, srcbase string) {
	l.AddTarget(srcbase+".c", "xs", []string{src}, "obj", srcdir, nil, nil)
	l.CompileC("", srcbase+".c", srcbase)
}

func (l *LinkDesc) CompileGo(srcdir, src, srcbase string) {
	l.AddTarget(path.Join("go", src), "copy_go", []string{src}, "obj", srcdir, nil, nil)
	l.GoSrc = append(l.GoSrc, path.Join("$objdir/go", src))
}
