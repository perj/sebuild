// Copyright 2019 Schibsted

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/schibsted/sebuild/pkg/buildbuild"
)

// BuildPlugin compiles a sebuild plugin in the given directory.
// This function is in the main package primarily because early in the
// plugin support it didn't work to load plugins from non-main packages.
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

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", binabs)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = ppath

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
		return err
	}
	_, err = plugin.Open(binpath)
	if err != nil && strings.Contains(err.Error(), "previous failure") {
		if !ops.Options.Quiet {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, `Trying again due to "previous failure".`)
		}
		err = buildbuild.ErrNeedReExec
	}
	return err
}
