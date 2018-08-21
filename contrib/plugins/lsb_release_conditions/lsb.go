// Copyright 2018 Schibsted

// This plugin runs `lsb_release` and adds condition flags to the build based on its output.
//
// This is all best explained by the following table, which shows what's added for a
// releases of Ubuntu 16.04, Debian/Jessie, CentOS 7 and CentOS 6.
//
// 	'ubuntu', 'ubuntu16', 'ubuntu16_04', 'xenial'
// 	'debian', 'debian8', 'debian8_7', 'jessie'
// 	'centos', 'centos7', 'centos7_3', 'core'
// 	'centos', 'centos6', 'centos6_8', 'final'
//
// An example use case would be to set `cxxflags` if NOT building on Ubuntu:
//
// 	cxxflags::!ubuntu[-std=gnu++0x -D_GLIBCXX_USE_CXX11_ABI=0]
//
package lsb_release_conditions

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
	"unicode"

	"github.com/schibsted/sebuild/pkg/buildbuild"
)

type Plugin struct {
	Ops *buildbuild.GlobalOps
}

func init() {
	buildbuild.InitPlugin("lsb_release_conditions", &Plugin{})
}

func (p *Plugin) Startup(ops *buildbuild.GlobalOps) error {
	p.Ops = ops

	lsb, err := exec.Command("lsb_release", "-ric").Output()
	if err != nil {
		lsb = []byte{}
	}
	s := bufio.NewScanner(bytes.NewReader(lsb))

	var id, release, verMaj, verMin, codename string
	for s.Scan() {
		line := strings.SplitN(s.Text(), ":", 2)

		if len(line) < 2 {
			continue
		}

		k := line[0]
		v := strings.ToLower(strings.TrimSpace(line[1]))

		switch k {
		case "Distributor ID":
			id = v
		case "Release":
			release = v
			if dot := strings.IndexRune(release, '.'); dot > 0 {
				verMaj = release[:dot]
				verMin = release[dot+1:]
				if dot := strings.IndexRune(verMin, '.'); dot >= 0 {
					verMin = verMin[:dot]
				}
			}
		case "Codename":
			codename = strings.Map(func(r rune) rune {
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					return r
				}
				return -1
			}, v)
		}
	}
	if id != "" {
		// e.g 'centos', 'centos7', 'centos7_3' and 'core'
		ops.Config.Conditions[id] = true
		ops.Config.Conditions[id+verMaj] = true
		if verMin != "" {
			ops.Config.Conditions[id+verMaj+"_"+verMin] = true
		}
		if codename != "" {
			ops.Config.Conditions[codename] = true
		}
	} else {
		ops.Config.Conditions["no_lsb_release"] = true
	}
	return nil
}
