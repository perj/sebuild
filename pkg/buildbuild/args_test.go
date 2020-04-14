// Copyright 2018 Schibsted

package buildbuild

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

type argTests struct {
	haveEnabled bool
	input       string
	expArgs     *Args
	expErr      error
}

var argstests = []argTests{
	{false, ``, &Args{}, nil},
	{false, `)`, &Args{}, nil},
	{false, `[]`, &Args{Unflavored: map[string][]string{"": []string{}}}, nil},
	{false, `[foo]`, &Args{Unflavored: map[string][]string{"": []string{"foo"}}}, nil},
	{false, `foo[bar baz] x[y]`, &Args{Unflavored: map[string][]string{
		"foo": []string{"bar", "baz"}, "x": []string{"y"},
	}}, nil},
	{false, `a[b] c:f[d] e::testcond[g] h::other[i] j:f:testcond,!other[k]`, &Args{Unflavored: map[string][]string{
		"a": []string{"b"}, "e": []string{"g"},
	}, Flavored: map[string]map[string][]string{
		"c": map[string][]string{"f": []string{"d"}},
		"j": map[string][]string{"f": []string{"k"}},
	}, Flavors: map[string]map[string][]string{
		"f": map[string][]string{
			"c": []string{"d"}, "j": []string{"k"},
		},
	}}, nil},
	{false, `a[b] a[c] a:f[d] a:f[e]`, &Args{Unflavored: map[string][]string{
		"a": []string{"b", "c"},
	}, Flavored: map[string]map[string][]string{
		"a": map[string][]string{
			"f": []string{"d", "e"},
		},
	}, Flavors: map[string]map[string][]string{
		"f": map[string][]string{
			"a": []string{"d", "e"},
		},
	}}, nil},
	{false, `foo`, nil, &ParseError{UnexpectedEOF, "", "test"}},
	{false, `foo:`, nil, &ParseError{UnexpectedEOF, "", "test"}},
	{false, `foo:bar`, nil, &ParseError{UnexpectedEOF, "", "test"}},
	{false, `foo:bar:`, nil, &ParseError{UnexpectedEOF, "", "test"}},
	{false, `foo:bar:baz`, nil, &ParseError{UnexpectedEOF, "", "test"}},
	{false, `foo:bar:baz[`, nil, &ParseError{UnexpectedEOF, "", "test"}},
	{false, `foo bar`, nil, &ParseError{MissingOpenBracket, "bar", "test"}},
	{false, `a[b[]] b[c [ ] ] c[d[ ]]`, &Args{Unflavored: map[string][]string{
		"a": []string{"b[]"}, "b": []string{"c", "[", "]"}, "c": []string{"d[", "]"},
	}}, nil},
	{true, `enabled[]`, &Args{Unflavored: map[string][]string{
		"enabled": []string{},
	}}, nil},
	{true, `enabled::none[]`, nil, nil},
}

func testParseArgs(args *Args, s *Scanner, conds map[string]bool) (haveEnabled bool, err error) {
	defer func() {
		p := recover()
		if p != nil {
			err = p.(error)
		}
	}()
	checkConds := func(condstr string) bool {
		for _, c := range strings.Split(condstr, ",") {
			condval, c := parseCond(c)
			if conds[c] != condval {
				return false
			}
		}
		return true
	}
	return args.Parse(s, checkConds), nil
}

func TestArgs(t *testing.T) {
	conds := map[string]bool{"testcond": true}

	for _, tst := range argstests {
		if tst.expArgs == nil {
			tst.expArgs = new(Args)
		}
		if tst.expArgs.Unflavored == nil {
			tst.expArgs.Unflavored = make(map[string][]string)
		}
		if tst.expArgs.Flavored == nil {
			tst.expArgs.Flavored = make(map[string]map[string][]string)
		}
		if tst.expArgs.Flavors == nil {
			tst.expArgs.Flavors = make(map[string]map[string][]string)
		}
		s := NewScanner(ioutil.NopCloser(strings.NewReader(tst.input)), "test")
		var args Args
		haveEnabled, err := testParseArgs(&args, s, conds)
		if haveEnabled != tst.haveEnabled {
			t.Error(tst.input, ": haveEnabled mismatch, got", haveEnabled)
		}
		if !reflect.DeepEqual(err, tst.expErr) {
			if tst.expErr == nil {
				t.Error(tst.input, ": Expected no error, got", err)
			} else {
				t.Error(tst.input, ": Expected error", tst.expErr, "got", err)
			}
		} else if !reflect.DeepEqual(&args, tst.expArgs) {
			t.Errorf("%v: Expected args %#v, got %#v", tst.input, tst.expArgs, args)
		}
	}
}
