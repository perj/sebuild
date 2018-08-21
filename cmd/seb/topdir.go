// Copyright 2018 Schibsted

package main

import (
	"errors"
	"os"
	"path/filepath"
)

var NoBuilddescFound = errors.New("No Builddesc found")

// Walking upwards, find either the first directory containing Builddesc.top
// or the last one containing Builddesc.
// Also return the directory of the first Builddesc found, to be able to
// warn if that wasn't processed.
func FindTopdir() (topdir, firstbd string, err error) {
	topdir, err = os.Getwd()
	if err != nil {
		return
	}
	cwd := topdir
	topbd := ""
	for {
		_, serr := os.Stat(filepath.Join(topdir, "Builddesc.top"))
		if serr == nil {
			if topdir == cwd {
				topdir = ""
			}
			return
		}
		_, serr = os.Stat(filepath.Join(topdir, "Builddesc"))
		if serr == nil {
			if firstbd != "" {
				firstbd = topdir
			}
			topbd = topdir
		}
		if topdir == "/" {
			break
		}
		topdir = filepath.Dir(topdir)
	}
	if topbd != "" {
		topdir = topbd
		if topdir == cwd {
			topdir = ""
		}
	} else {
		topdir = ""
		err = NoBuilddescFound
	}
	return
}
