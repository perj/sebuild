// Copyright 2018 Schibsted

package buildbuild

import (
	"errors"
	"strings"
)

// Arguments are a key and a list of elements. After the key you use a
// bracket pair [] and put the white space separated elements inside.
// Many arguments can only contain one element, those will either join
// a list into a single string, or use only the first element.
// Some arguments take no values, but the brackets are still required.
// Arguments are things like: srcs, includes, libs, deps, etc.
//
// After the key you can use a colon : and a flavor. If used, the argument
// applies to that flavor only. You can then but another color and a comma
// separated list of conditions that need to match. Conditions can be
// negated by prefixing them with an exclamation sign !.
// You can leave the flavor empty if you wish to apply conditions without
// chosing a flavor.
//
// Examples: srcs[a.c b.c] copts:prof[-pg] libs::linux[rt]
type Args struct {
	Unflavored map[string][]string
	Flavored   map[string]map[string][]string
	Flavors    map[string]map[string][]string
}

var (
	MissingOpenBracket   = errors.New("Missing open bracket in parameter")
	ConditionsNotAllowed = errors.New("Conditions are not allowed here")
)

// Parse from s until a ) token is found, indicating the end of a descriptor.
// Fills in *args with the arguments found, if the conditions set on the
// argument match the ones given (if no conditions are set it's always a
// match).
//
// Returns true if a `enabled` argument was found, regardless if it was saved
// or not. Returning true and then not finding enabled in the arguments would
// imply the descriptor should be skipped.
func (args *Args) Parse(s *Scanner, conditions map[string]bool) (haveEnabled bool) {
	*args = Args{
		make(map[string][]string),
		make(map[string]map[string][]string),
		make(map[string]map[string][]string),
	}
	for s.Scan() {
		if s.Text() == ")" {
			break
		}
		var key, flavor, cond string
		// Allow empty key, used by COMPONENT
		if s.Text() != "[" {
			// Not empty, format key:flavor:conditions[values...]
			for _, ptr := range []*string{&key, &flavor, &cond} {
				if s.Text() != ":" {
					*ptr = s.Text()
					if !s.Scan() {
						panicOrEOF(s)
					}
					if s.Text() != ":" {
						break
					}
				}
				if !s.Scan() {
					panicOrEOF(s)
				}
			}
			if key == "enabled" {
				haveEnabled = true
			}
		}
		if s.Text() != "[" {
			panic(&ParseError{MissingOpenBracket, s.Text(), s.Filename})
		}
		var value []string
		level := 1
		s.scannerSpecials = argsSpecials
		s.scannerNewword = true
		for {
			if !s.Scan() {
				panicOrEOF(s)
			}
			switch s.Text() {
			case "[":
				level++
			case "]":
				level--
			}
			if level <= 0 {
				break
			}
			// Bit of a hack, we need to make sure to not treat " [" the same as "["
			if s.scannerNewword {
				value = append(value, s.Text())
			} else {
				value[len(value)-1] += s.Text()
			}
			s.scannerNewword = false
		}
		s.scannerSpecials = builddescSpecials

		if cond != "" && conditions == nil {
			panic(&ParseError{ConditionsNotAllowed, cond, s.Filename})
		}
		if cond != "" && !CheckConditions(cond, conditions) {
			continue
		}

		if flavor == "" {
			args.Unflavored[key] = append(args.Unflavored[key], value...)
			if args.Unflavored[key] == nil {
				args.Unflavored[key] = make([]string, 0)
			}
		} else {
			m := args.Flavored[key]
			if m == nil {
				m = make(map[string][]string)
				args.Flavored[key] = m
			}
			m[flavor] = append(m[flavor], value...)
			m = args.Flavors[flavor]
			if m == nil {
				m = make(map[string][]string)
				args.Flavors[flavor] = m
			}
			m[key] = append(m[key], value...)
		}
	}
	panicIfErr(s)
	return
}

func CheckConditions(condstr string, conditions map[string]bool) bool {
	for _, cond := range strings.Split(condstr, ",") {
		condval := true
		if strings.HasPrefix(cond, "!") {
			condval = false
			cond = cond[1:]
		}
		if conditions[cond] != condval {
			return false
		}
	}
	return true
}
