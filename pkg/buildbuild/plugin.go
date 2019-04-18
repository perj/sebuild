// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"fmt"
	gobuild "go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Plugin interface {
	Startup(ops *GlobalOps) error
}

var LoadedPlugins = map[string]Plugin{}

var (
	// Descriptor defaults defined by plugins.
	PluginDescriptors = map[string]Descriptor{}
	// Params added by plugins that are allowed in all descriptors.
	PluginGeneralParams = map[string]ParseGeneralParam{}
	PluginLinkerParams  = map[string]ParseLinkerParam{}
	// File extensions added by plugins.
	PluginGeneralExtensions = map[string]CompileGeneralSrcFunc{}
	PluginLinkerExtensions  = map[string]CompileLinkerSrcFunc{}
	// Specialsrcs rules that plugin will process.
	PluginSpecialSrcs = map[string]CompileSpecialSrcFunc{}
)

var (
	NoSuchPlugin = errors.New("Plugin doesn't exist in binary")
)

func InitPlugin(pkg string, plug Plugin) {
	if LoadedPlugins[pkg] != nil {
		panic(`Duplicate plugin "` + pkg + `"`)
	}
	LoadedPlugins[pkg] = plug
}

func (ops *GlobalOps) TempDirWithPlugins(plugs []string) string {
	tmpdir, err := ioutil.TempDir("", "build-build")
	if err != nil {
		panic(err)
	}
	abs, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	os.Symlink(abs, filepath.Join(tmpdir, "cwd"))
	impname := filepath.Join(tmpdir, "imports.go")
	impfile, err := os.OpenFile(impname, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintf(impfile, "package main\n\n")
	if err != nil {
		panic(err)
	}
	for _, p := range plugs {
		// Could use vendor here, but feels like abusing it.
		p = filepath.Join("cwd", p)
		_, err := fmt.Fprintf(impfile, "import _ \"./%s\"\n", p)
		if err != nil {
			panic(err)
		}
	}
	impfile.Close()
	return tmpdir
}

func (ops *GlobalOps) RecompileWithPlugins() {
	if !ops.Options.Quiet {
		fmt.Fprintf(os.Stderr, "Recompiling seb with plugins (will fail if source is not available)\n")
	}
	var gofiles []string
	if mypkg, err := gobuild.Import(modpath+"/cmd/seb", "", 0); err == nil {
		for _, f := range mypkg.GoFiles {
			gofiles = append(gofiles, filepath.Join(mypkg.Dir, f))
		}
	} else {
		d := BuildtoolDir()
		pattern := filepath.Join(d, "cmd", "seb", "*.go")
		gofiles, err = filepath.Glob(pattern)
		if err != nil {
			panic(err)
		}
	}
	tmpdir := ops.TempDirWithPlugins(ops.Config.Plugins)
	defer os.RemoveAll(tmpdir)
	for _, f := range gofiles {
		abs, err := filepath.Abs(f)
		if err != nil {
			panic(err)
		}
		dst := filepath.Join(tmpdir, filepath.Base(f))
		err = os.Symlink(abs, dst)
		if err != nil {
			panic(err)
		}
	}

	binpath := path.Join(ops.Config.Buildpath, "obj/_build_build")
	binabs, err := filepath.Abs(binpath)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "build", "-o", binabs)
	cmd.Dir = tmpdir
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	for idx, str := range cmd.Env {
		// Workaround weird go invocation of CC, it just picks the first word, ignoring others.
		// Assume the last word not starting with - is the one we want
		if strings.HasPrefix(str, "CC=") {
			args := strings.Split(str[3:], " ")
			var i int
			for i = len(args) - 1; i > 0; i-- {
				if len(args[i]) > 0 && args[i][0] != '-' {
					break
				}
			}
			cmd.Env[idx] = "CC=" + strings.Join(args[i:], " ")
			break
		}
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func (ops *GlobalOps) LoadPlugin(bd, ppath string) (Plugin, error) {
	pkg := path.Base(ppath)
	if plug := LoadedPlugins[pkg]; plug != nil {
		return plug, nil
	}
	if ops.BuildPlugin != nil {
		err := ops.BuildPlugin(ops, ppath)
		if err != nil {
			return nil, &ParseError{err, bd, ppath}
		}
		if plug := LoadedPlugins[pkg]; plug != nil {
			return plug, nil
		}
	}
	return nil, &ParseError{NoSuchPlugin, ppath, bd}
}

func (ops *GlobalOps) StartPlugin(bd, ppath string) error {
	plug, err := ops.LoadPlugin(bd, ppath)
	if err != nil {
		return err
	}

	return plug.Startup(ops)
}
