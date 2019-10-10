// Copyright 2018 Schibsted

package buildbuild

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// The CONFIG descriptor must be the first one encountered either in
// Builddesc.top or the top level Builddesc. It contains global configuration
// for the build, for example valid flavors and enabled plugins.
//
// For many projects, the defaults are fine, and the CONFIG descriptor
// can then be skipped.
//
// Arguments:
//
// buildversion_script - script that outputs one number which is the version
// of what's being built. It's highly recommended that the version number is
// unique for at least every commit to your repository.
//
// compiler - Override the compiler used, set it to the C compiler, C++ one
// will be guessed with some heuristics.
//
// flavors - Various build environments needed to build your site. The usual
// is to build prod and regress.
//
// rules - Global compilation rules. These ninja files gets included globally.
//
// configvars - Global ninja variables. These ninja files gets included
// globally, and are also passed as arguments to the invars script.
//
// extravars - Per flavor-included ninja files. This means they can depend on
// the variables defined in the flavor files. Can be flavored.
//
// buildpath - Where the build files are put. See other sections of this
// document to see how files are organized.
//
// buildvars - attributes in other build descriptors that are copied into ninja
// files as variables. As in this example, those are various variables we want
// to be able to specify in build decsriptors that override default variables.
//
// ruledeps - Per-rule dependencies. Targets built with a certain rule will
// depend on those additional target. In this example everything built with
// the in rule will also depend on $inconf.
//
// prefix:flavor - Set a prefix for the installed files for the specified
// flavor. Must be flavored.
//
// config_script - Run a script whenever build-build is run and parse its
// output as variables or conditions.
//
// cflags:flavor - CFLAGS for a flavor. Must be flavored.
//
// compiler_rule_dir, flavor_rule_dir, compiler_flavor_rule_dir -
// Directories containing ninja files included based on current compiler
// and/or flavor.
type Config struct {
	Seen bool

	Conditions  map[string]bool
	Buildparams []string

	AllFlavors    map[string]bool
	ActiveFlavors []string // Flavors left after filtering --with-flavors and --without-flavors.

	Plugins    []string
	Configvars []string // Files with ninja variables, available to invars.
	Rules      []string // Files with ninja rules.
	Extravars  []string
	Invars     []string

	CompilerRuleDir       string
	FlavorRuleDir         string
	CompilerFlavorRuleDir string

	Ruledeps   map[string][]string // Additional dependencies for particular rules.
	Buildvars  []string
	Compiler   []string
	Godeps     []string
	GodepsRule string

	BuildversionScript string
	Buildpath          string
	ConfigScript       string

	BuiltinInvars string
}

type FlavorConfig struct {
	Prefix    string
	Extravars []string
	Cflags    string
}

var (
	RuledepsError            = errors.New("Bad ruledeps argument")
	FlavoredConfigUnknownArg = errors.New("Unrecognized flavored argument in CONFIG")
	ConfigMustBeFlavored     = errors.New("CONFIG argument must be flavored")
	ConfigUnknownArg         = errors.New("Unrecognized argument in CONFIG")
	BadFlavor                = errors.New("Flavor does not exist")
)

func (ops *GlobalOps) DefaultConfig() {
	ops.Config.Conditions = make(map[string]bool)
	ops.Config.Ruledeps = make(map[string][]string)

	ops.Config.AllFlavors = map[string]bool{"dev": true}
	ops.Config.ActiveFlavors = []string{"dev"}

	ops.Config.Buildpath = os.Getenv("BUILDPATH")
	if ops.Config.Buildpath == "" {
		ops.Config.Buildpath = "build"
	}
	ops.Config.BuildversionScript = "git rev-list HEAD 2>/dev/null|wc -l|xargs"
	ops.Config.CompilerRuleDir = "$buildtooldir/rules/compiler"
	ops.Config.FlavorRuleDir = "$buildtooldir/rules/flavor"
	ops.Config.CompilerFlavorRuleDir = "$buildtooldir/rules/compiler-flavor"
	ops.Config.GodepsRule = "godeps"

	ops.Config.Conditions[runtime.GOOS] = true
	if runtime.GOARCH == "amd64" {
		ops.Config.Conditions["x86_64"] = true
	} else {
		ops.Config.Conditions[runtime.GOARCH] = true
	}

	ops.Config.Ruledeps["in"] = []string{"$inconf", "$configvars"}

	ops.FlavorConfigs = make(map[string]*FlavorConfig)
}

// Redirected by test
var parseConfigOpenBuilddesc = (*GlobalOps).OpenBuilddesc

func (ops *GlobalOps) ParseConfig(srcdir string, s *Scanner, flavors []string) ParseFunc {
	var args Args
	args.Parse(s, nil)

	if args.Unflavored["flavors"] != nil {
		ops.Config.AllFlavors = make(map[string]bool)
		ops.Config.ActiveFlavors = nil
		for _, fl := range args.Unflavored["flavors"] {
			ops.Config.AllFlavors[fl] = true
			if len(ops.Options.WithFlavors) > 0 && !ops.Options.WithFlavors[fl] {
				continue
			}
			if len(ops.Options.WithoutFlavors) > 0 && ops.Options.WithoutFlavors[fl] {
				continue
			}
			ops.Config.ActiveFlavors = append(ops.Config.ActiveFlavors, fl)
		}
		delete(args.Unflavored, "flavors")
	}
	// Check that flavors are valid now that we have set them.
	for fl := range args.Flavors {
		if !ops.Config.AllFlavors[fl] {
			panic(&ParseError{BadFlavor, fl, s.Filename})
		}
	}

	// We need to handle local ninja files before the included ones. The reason for this is complex.
	// This is the only instance of the order of elements in attributes actually being significant.
	for _, ninja := range []struct {
		key  string
		conf *[]string
	}{
		{"configvars", &ops.Config.Configvars},
		{"rules", &ops.Config.Rules},
		{"extravars", &ops.Config.Extravars},
		{"godeps", &ops.Config.Godeps},
	} {
		for _, pth := range args.Unflavored[ninja.key] {
			// XXX Why not normalizePath ?
			pth = path.Join(srcdir, pth)
			*ninja.conf = append(*ninja.conf, pth)
		}
		delete(args.Unflavored, ninja.key)
	}

	// ruledeps format is <rule>:<dependency>[,<dependency>]*
	for _, dep := range args.Unflavored["ruledeps"] {
		depargs := strings.SplitN(dep, ":", 2)
		if len(depargs) < 2 {
			panic(&ParseError{RuledepsError, dep, s.Filename})
		}
		depsrcs := strings.Split(depargs[1], ",")
		ops.Config.Ruledeps[depargs[0]] = append(ops.Config.Ruledeps[depargs[0]], depsrcs...)
	}
	delete(args.Unflavored, "ruledeps")

	ops.Config.Compiler = append(ops.Config.Compiler, args.Unflavored["compiler"]...)
	delete(args.Unflavored, "compiler")

	for _, inc := range args.Unflavored["INCLUDE"] {
		inc = NormalizePath(srcdir, inc)
		s, err := parseConfigOpenBuilddesc(ops, inc)
		if err != nil {
			panic(err)
		}
		ops.ParseConfig(path.Dir(inc), s, flavors)
		s.Close()
	}
	delete(args.Unflavored, "INCLUDE")

	// Each appearance of these appends.
	ops.Config.Buildvars = append(ops.Config.Buildvars, args.Unflavored["buildvars"]...)
	delete(args.Unflavored, "buildvars")
	ops.Config.Plugins = append(ops.Config.Plugins, args.Unflavored["extensions"]...)
	delete(args.Unflavored, "extensions")
	ops.Config.Invars = append(ops.Config.Invars, args.Unflavored["invars"]...)
	delete(args.Unflavored, "invars")

	// These overwrite the previous value.
	for _, conf := range []struct {
		key  string
		conf *string
	}{
		{"buildversion_script", &ops.Config.BuildversionScript},
		{"buildpath", &ops.Config.Buildpath},
		{"config_script", &ops.Config.ConfigScript},
		{"compiler_rule_dir", &ops.Config.CompilerRuleDir},
		{"flavor_rule_dir", &ops.Config.FlavorRuleDir},
		{"compiler_flavor_rule_dir", &ops.Config.CompilerFlavorRuleDir},
		{"godeps_rule", &ops.Config.GodepsRule},
		{"builtin_invars", &ops.Config.BuiltinInvars},
	} {
		if args.Unflavored[conf.key] != nil {
			*conf.conf = strings.Join(args.Unflavored[conf.key], " ")
			delete(args.Unflavored, conf.key)
		}
	}

	for _, cond := range args.Unflavored["conditions"] {
		ops.Config.Conditions[cond] = true
	}
	delete(args.Unflavored, "conditions")

	// Parse the arguments needing a flavor.
	for _, fl := range ops.Config.ActiveFlavors {
		conf := new(FlavorConfig)
		ops.FlavorConfigs[fl] = conf
		flargs := args.Flavors[fl]
		if flargs == nil {
			continue
		}
		conf.Prefix = strings.Join(flargs["prefix"], " ")
		delete(flargs, "prefix")
		conf.Extravars = flargs["extravars"]
		delete(flargs, "extravars")
		conf.Cflags = strings.Join(flargs["cflags"], " ")
		delete(flargs, "cflags")
		for k := range flargs {
			panic(&ParseError{FlavoredConfigUnknownArg, k, s.Filename})
		}
	}
	for _, k := range []string{"prefix", "cflags"} {
		if args.Unflavored[k] != nil {
			panic(&ParseError{ConfigMustBeFlavored, k, s.Filename})
		}
	}

	// Check for unparsed arguments.
	for k := range args.Unflavored {
		panic(&ParseError{ConfigUnknownArg, k, s.Filename})
	}

	if ops.PostConfigFunc != nil {
		if err := ops.PostConfigFunc(ops); err != nil {
			panic(err)
		}
	}

	return ops.RunConfigScript
}

func (ops *GlobalOps) RunConfigScript(srcdir string, s *Scanner, flavors []string) ParseFunc {
	if ops.Config.ConfigScript != "" {
		cmd := exec.Command("sh", "-c", ops.Config.ConfigScript)
		cmd.Stderr = os.Stderr
		cdata, err := cmd.Output()
		if err != nil {
			panic(fmt.Errorf("config_script[%s] failed: %v", ops.Config.ConfigScript, err))
		}
		cscan := bufio.NewScanner(bytes.NewReader(cdata))
		for cscan.Scan() {
			line := strings.TrimSpace(cscan.Text())
			if strings.ContainsRune(line, '=') {
				ops.Config.Buildparams = append(ops.Config.Buildparams, line)
			} else if line != "" {
				ops.Config.Conditions[line] = true
			}
		}
	}
	return ops.StartupPlugins
}

func (ops *GlobalOps) StartupPlugins(srcdir string, s *Scanner, flavors []string) ParseFunc {
	var missingPlugins []string
	for _, ppath := range ops.Config.Plugins {
		ppath = NormalizePath(srcdir, ppath)
		err := ops.StartPlugin(s.Filename, ppath)
		if err != nil {
			if perr, ok := err.(*ParseError); ok && perr.Err == NoSuchPlugin {
				missingPlugins = append(missingPlugins, ppath)
				continue
			}
			panic(err)
		}
	}
	if len(missingPlugins) > 0 {
		panic("Failed to load plugins: " + strings.Join(missingPlugins, ", "))
	}

	return ops.ParseDescriptorEnd
}
