// Copyright 2018 Schibsted

package buildbuild

import (
	"os"
	"path"
)

const Builddesc = "Builddesc"

func (ops *GlobalOps) OpenBuilddesc(file string) (s *Scanner, err error) {
	var bdfile *os.File
	bdfile, err = os.Open(file)
	if err != nil {
		return
	}
	ops.Builddescs = append(ops.Builddescs, file)
	s = NewScanner(bdfile, file)
	return
}

func (ops *GlobalOps) OpenComponent(dir string) (s *Scanner, err error) {
	bdpath := path.Join(dir, Builddesc)
	if dir == "" {
		s, err = ops.OpenBuilddesc(bdpath + ".top")
		if !os.IsNotExist(err) {
			return
		}
	}
	return ops.OpenBuilddesc(bdpath)
}

func (ops *GlobalOps) ReadComponent(dir string, flavors []string) (err error) {
	s, err := ops.OpenComponent(dir)
	if err != nil {
		return
	}
	defer s.Close()
	defer func() {
		p := recover()
		if perr, ok := p.(error); ok {
			err = perr
		} else if p != nil {
			panic(p)
		}
	}()

	next := ops.ParseDirective
	for next != nil {
		next = next(dir, s, flavors)
	}
	return nil
}
