package buildbuild

import (
	"errors"
	"path"
)

var (
	UnhandledBuildDirective = errors.New("Unhandled build directive")
	MissingOpenParen        = errors.New("Missing open parenthesis")
	MissingCloseParen       = errors.New("Missing close parenthesis")
	DuplicateConfig         = errors.New("Duplicate CONFIG or CONFIG wasn't first")
	UnexpectedEOF           = errors.New("Unexpected end of file")
)

type ParseError struct {
	Err       error
	Token     string
	Builddesc string
}

func (e *ParseError) Error() string {
	return e.Builddesc + ": " + e.Err.Error() + " near " + e.Token
}

func panicOrEOF(s *Scanner) {
	err := s.Err()
	if err == nil {
		err = &ParseError{UnexpectedEOF, "", s.Filename}
	}
	panic(err)
}

func panicIfErr(s *Scanner) {
	if err := s.Err(); err != nil {
		panic(err)
	}
}

type ParseFunc func(string, *Scanner, []string) ParseFunc

func (ops *GlobalOps) ParseDirective(srcdir string, s *Scanner, flavors []string) ParseFunc {
	if !s.Scan() {
		panicIfErr(s)
		return nil
	}
	dname := s.Text()
	if !s.Scan() {
		panicOrEOF(s)
	}
	if s.Text() != "(" {
		panic(&ParseError{MissingOpenParen, s.Text(), s.Filename})
	}
	confseen := ops.Config.Seen
	ops.Config.Seen = true
	if dname == "CONFIG" {
		if confseen {
			panic(&ParseError{DuplicateConfig, dname, s.Filename})
		}
		return ops.ParseConfig
	}
	if !confseen && ops.PostConfigFunc != nil {
		if err := ops.PostConfigFunc(ops); err != nil {
			panic(err)
		}
	}
	if dname == "COMPONENT" {
		return ops.ParseComponent
	}
	if defdesc := PluginDescriptors[dname]; defdesc != nil {
		dp := &DescParser{ops, defdesc}
		return dp.Parse
	}
	if defdesc := DefaultDescriptors[dname]; defdesc != nil {
		dp := &DescParser{ops, defdesc}
		return dp.Parse
	}
	panic(&ParseError{UnhandledBuildDirective, dname, s.Filename})
}

type DescParser struct {
	Ops     *GlobalOps
	DefDesc Descriptor
}

func (dp *DescParser) Parse(srcdir string, s *Scanner, flavors []string) ParseFunc {
	if !s.Scan() {
		panicOrEOF(s)
	}
	tname := s.Text()

	var args Args
	haveEnabled := args.Parse(s, dp.Ops.Config.Conditions)

	if flavors == nil {
		flavors = dp.Ops.Config.ActiveFlavors
	}
	descFlavors := args.Unflavored["flavors"]
	if len(descFlavors) == 0 {
		if len(args.Flavors) > 0 {
			// If we have flavored arguments, we need to split the descriptors.
			descFlavors = flavors
		} else {
			descFlavors = []string{""}
		}
	}
	delete(args.Unflavored, "flavors")

	for _, fl := range descFlavors {
		flargs := make(map[string][]string)
		for k, v := range args.Unflavored {
			flargs[k] = v
		}
		for k, v := range args.Flavors[fl] {
			flargs[k] = append(flargs[k], v...)
		}

		// If an enabled argument exists then skip this descriptor if
		// it's not currently set.
		if _, ok := flargs["enabled"]; haveEnabled && !ok {
			continue
		}
		delete(flargs, "enabled")

		onlyForFlavors := flavors
		if fl != "" {
			onlyForFlavors = []string{fl}
		}
		desc := dp.DefDesc.NewFromTemplate(s.Filename, tname, onlyForFlavors)
		desc = desc.Parse(dp.Ops, srcdir, flargs)
		dp.Ops.Descriptors = append(dp.Ops.Descriptors, desc)
	}
	return dp.Ops.ParseDescriptorEnd
}

func (ops *GlobalOps) ParseDescriptorEnd(srcdir string, s *Scanner, flavors []string) ParseFunc {
	if s.Text() != ")" {
		panic(&ParseError{MissingCloseParen, s.Text(), s.Filename})
	}
	return ops.ParseDirective
}

func (ops *GlobalOps) ParseComponent(srcdir string, s *Scanner, flavors []string) ParseFunc {
	var args Args
	args.Parse(s, ops.Config.Conditions)
	if args.Unflavored["flavors"] != nil {
		flavors = args.Unflavored["flavors"]
	}
	for _, comp := range args.Unflavored[""] {
		// XXX comp = normalizePath(srcdir, comp)
		comp = path.Join(srcdir, comp)
		err := ops.ReadComponent(comp, flavors)
		if err != nil {
			panic(err)
		}
	}
	return ops.ParseDescriptorEnd
}
