// Copyright 2018 Schibsted

//+build go1.8

package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"plugin"

	"github.com/schibsted/sebuild/pkg/buildbuild"
)

func BuildPlugin(ops *buildbuild.GlobalOps, ppath string) error {
	binpath := path.Join(ops.Config.Buildpath, "obj/_plugins", ppath+".so")

	_, err := plugin.Open(binpath)
	if err == nil {
		return nil
	}

	binabs, err := filepath.Abs(binpath)
	if err != nil {
		return err
	}
	tmpdir := ops.TempDirWithPlugins([]string{ppath})
	defer os.RemoveAll(tmpdir)
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", binabs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpdir
	err = cmd.Run()
	if err != nil {
		return err
	}
	_, err = plugin.Open(binpath)
	return err
}
