// Copyright 2018 Schibsted

//+build !go1.8

package main

import "github.com/schibsted/sebuild/pkg/buildbuild"

func BuildPlugin(ops *buildbuild.GlobalOps, ppath string) error {
	// Return nil is ok since the plugin still won't be found in the parent code.
	return nil
}
