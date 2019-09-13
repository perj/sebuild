// Copyright 2018 Schibsted

package buildbuild

import (
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestParseConfigAll(t *testing.T) {
	r := strings.NewReader(`
flavors[a b c]
configvars[config.ninja]
rules[rules.ninja]
extravars[extra.ninja]
compiler_rule_dir[$buildtooldir/rules/compiler]
flavor_rule_dir[$buildtooldir/rules/flavor]
compiler_flavor_rule_dir[$buildtooldir/rules/compiler-flavor]
ruledeps[q:qtool,ztool]
INCLUDE[testinclude]
config_script[./config_script.sh]
buildvars[foo]
compiler[cc]
extensions[exts]
conditions[pbuild]
buildversion_script[bv.sh]
buildpath[buildpath]
prefix:a[apref]
extravars:a[avars.ninja]
cflags:a[aflags]
`)
	s := NewScanner(ioutil.NopCloser(r), "test")

	parseConfigOpenBuilddesc = func(ops *GlobalOps, path string) (s *Scanner, err error) {
		r := strings.NewReader(`
configvars[config2.ninja]
rules[rules2.ninja]
extravars[extra2.ninja]
ruledeps[q:xtool]
config_script[./overridden.sh]
buildversion_script[overridden.sh]
buildpath[overridden]
buildvars[foo2]
compiler[cc2]
`)
		s = NewScanner(ioutil.NopCloser(r), path)
		return
	}
	defer func() {
		parseConfigOpenBuilddesc = (*GlobalOps).OpenBuilddesc
	}()

	ops := NewGlobalOps()
	ops.ParseConfig("", s, nil)

	march := runtime.GOARCH
	if march == "amd64" {
		march = "x86_64"
	}

	c := &ops.Config
	e := &Config{
		Conditions:            map[string]bool{"pbuild": true, runtime.GOOS: true, march: true},
		Buildparams:           nil,
		AllFlavors:            map[string]bool{"a": true, "b": true, "c": true},
		ActiveFlavors:         []string{"a", "b", "c"},
		Plugins:               []string{"exts"},
		Configvars:            []string{"config.ninja", "config2.ninja"},
		Rules:                 []string{"rules.ninja", "rules2.ninja"},
		Extravars:             []string{"extra.ninja", "extra2.ninja"},
		CompilerRuleDir:       "$buildtooldir/rules/compiler",
		FlavorRuleDir:         "$buildtooldir/rules/flavor",
		CompilerFlavorRuleDir: "$buildtooldir/rules/compiler-flavor",
		Ruledeps: map[string][]string{
			"q":  []string{"qtool", "ztool", "xtool"},
			"in": []string{"$inconf", "$configvars"},
		},
		Buildvars:          []string{"foo2", "foo"},
		Compiler:           []string{"cc", "cc2"},
		BuildversionScript: "bv.sh",
		Buildpath:          "buildpath",
		ConfigScript:       "./config_script.sh",
		GodepsRule:         "godeps",
	}

	if !reflect.DeepEqual(c, e) {
		t.Errorf("Config didn't match expected, got:")
		t.Errorf("%#v", c)
	}

	fc := ops.FlavorConfigs["a"]
	fe := &FlavorConfig{
		Prefix:    "apref",
		Extravars: []string{"avars.ninja"},
		Cflags:    "aflags",
	}

	if !reflect.DeepEqual(fc, fe) {
		t.Errorf("Flavor config didn't match expected, got:")
		t.Errorf("%#v", fc)
	}
}
