// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
