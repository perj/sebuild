// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
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

	ErrNeedReExec = errors.New("need to re-execute binary")
)

func InitPlugin(pkg string, plug Plugin) {
	if LoadedPlugins[pkg] != nil {
		panic(`Duplicate plugin "` + pkg + `"`)
	}
	LoadedPlugins[pkg] = plug
}

func (ops *GlobalOps) LoadPlugin(bd, ppath string) (Plugin, error) {
	pkg := path.Base(ppath)
	if plug := LoadedPlugins[pkg]; plug != nil {
		return plug, nil
	}
	if ops.BuildPlugin != nil {
		err := ops.BuildPlugin(ops, ppath)
		if err != nil {
			if err == ErrNeedReExec {
				ops.ReExec()
			}
			return nil, &ParseError{err, bd, ppath}
		}
		if plug := LoadedPlugins[pkg]; plug != nil {
			return plug, nil
		}
	}
	return nil, &ParseError{NoSuchPlugin, ppath, bd}
}

func (ops *GlobalOps) ReExec() {
	bin, err := exec.LookPath(os.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to re-exec %q", os.Args[0])
		return
	}
	if err := syscall.Exec(bin, os.Args, syscall.Environ()); err != nil {
		panic(err)
	}
}

func (ops *GlobalOps) StartPlugin(bd, ppath string) error {
	plug, err := ops.LoadPlugin(bd, ppath)
	if err != nil {
		return err
	}

	return plug.Startup(ops)
}
