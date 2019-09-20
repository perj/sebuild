// Copyright 2019 Schibsted

package gperf_enum

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s -tool gperf-enum <mode> <src> <gperf-file>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nMode is either enum or source.\n")
	os.Exit(1)
}

func isIdentifier(name string) bool {
	for i, ch := range name {
		if unicode.IsLetter(ch) || ch == '_' || i > 0 && unicode.IsDigit(ch) {
			continue
		}
		return false
	}
	return len(name) > 0
}

func Main(args ...string) {
	if len(args) != 3 {
		usage()
	}
	sourceMode := false
	switch args[0] {
	case "enum":
	case "source":
		sourceMode = true
	default:
		usage()
	}

	inf, err := os.Open(args[1])
	out := args[2]
	outname := filepath.Base(out)
	ext := filepath.Ext(outname)
	if ext != ".gperf" {
		fmt.Fprintf(os.Stderr, "Must use a gperf file with a .gperf extension.\n")
		os.Exit(1)
	}
	name := strings.TrimSuffix(outname, ext)
	if !isIdentifier(name) {
		fmt.Fprintf(os.Stderr, "Gperf file base must be a valid identifier.\n")
		os.Exit(1)
	}
	var prefix string
	for _, part := range strings.Split(name, "_") {
		if len(part) > 0 {
			r, _ := utf8.DecodeRune([]byte(part))
			prefix += string(unicode.ToUpper(r))
		}
	}
	prefix += "_"

	startRE, err := regexp.Compile(`GPERF_ENUM(_NOCASE)?[[:space:]]*\(` + name + `(?:;([^;]*))?\)`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create regexp: %s\n", err)
		os.Exit(1)
	}
	started := false
	type tmplField struct {
		String string
		Field  string
		Value  string
		Extra  string
	}
	tmplData := struct {
		Name       string
		Prefix     string
		DeclExtra  string
		NoCase     bool
		SourceMode bool
		Fields     []tmplField
	}{Name: name, Prefix: prefix, SourceMode: sourceMode}

	caseRE := regexp.MustCompile(`GPERF_CASE\("((?:[^"\\]|\\.)*)"([^)]*)\)`)
	caseValueRE := regexp.MustCompile(`GPERF_CASE_VALUE\(([^,]+),[[:space:]]*"((?:[^"\\]|\\.)*)"([^)]*)\)`)
	s := bufio.NewScanner(inf)
	var lineno int
	for s.Scan() {
		lineno++
		var field tmplField
		field.Value = fmt.Sprint(lineno)
		if sourceMode {
			if !started {
				matches := startRE.FindSubmatch(s.Bytes())
				if matches == nil {
					continue
				}
				tmplData.NoCase = len(matches[1]) > 0
				tmplData.DeclExtra = string(matches[2])
				if !strings.HasSuffix(tmplData.DeclExtra, ";") {
					tmplData.DeclExtra += ";"
				}
				started = true
				continue
			}
			// Stop on the next enum start (or EOF).
			if bytes.Contains(s.Bytes(), []byte("GPERF_ENUM")) {
				break
			}
			if m1 := caseRE.FindSubmatch(s.Bytes()); m1 != nil {
				field.String = string(m1[1])
				field.Extra = string(m1[2])
			} else if m2 := caseValueRE.FindSubmatch(s.Bytes()); m2 != nil {
				field.Value = string(m2[1])
				field.String = string(m2[2])
				field.Extra = string(m2[3])
			} else {
				continue
			}
		} else {
			field.String = s.Text()
		}
		field.Field = strings.ReplaceAll(strings.ToUpper(field.String), "-", "_")
		tmplData.Fields = append(tmplData.Fields, field)
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to scan %s: %s\n", args[1], err)
		os.Exit(1)
	}

	tmpl, err := template.New("").Parse(tmplSource)
	if err != nil {
		panic(err)
	}

	var outbuf bytes.Buffer
	err = tmpl.Execute(&outbuf, tmplData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run template: %s\n", err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(out, outbuf.Bytes(), 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write %s: %s\n", out, err)
		os.Remove(out)
		os.Exit(1)
	}
}

var tmplSource = `
struct {{.Name}}_rec;
enum {{.Name}} {
	{{$.Prefix}}NONE,
{{range .Fields}}	{{$.Prefix}}{{.Field}} = {{.Value}},
{{end -}}
};
struct {{.Name}}_rec {
	const char *name;
	enum {{.Name}} val;
	{{.DeclExtra}}
};

%struct-type
%define hash-function-name {{.Name}}_hash
%define lookup-function-name {{.Name}}_lookup
%readonly-tables
%enum
%compare-strncmp
%define string-pool-name {{.Name}}_strings
%define word-array-name {{.Name}}_words
{{if .NoCase}}%ignore-case{{end}}

%{
#include <string.h>

{{if not .SourceMode}}
#define MAX_{{$.Prefix}}WORD {{len .Fields}}
{{end}}

#ifndef register
#define register
#endif

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wmissing-declarations"
#if defined __clang__
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wstatic-in-inline"
#endif

%}

%%
{{range .Fields -}}
{{.String}}, {{$.Prefix}}{{.Field}}{{.Extra}}
{{end -}}
%%

static
enum {{.Name}} lookup_{{.Name}} (const char *str, int len) __attribute__((unused));

static
enum {{.Name}} lookup_{{.Name}} (const char *str, int len) {
	if (len == -1)
		len = strlen (str);
	const struct {{.Name}}_rec *val = {{.Name}}_lookup (str, len);

	if (val)
		return val->val;
	return {{$.Prefix}}NONE;
}

static
int lookup_{{.Name}}_int (const char *str, int len) __attribute__((unused));

static
int lookup_{{.Name}}_int (const char *str, int len) {
	return (int)lookup_{{.Name}} (str, len);
}

#if defined __clang__
#pragma clang diagnostic pop
#endif
#pragma GCC diagnostic pop
{{if .SourceMode}}
#ifndef GPERF_ENUM
#define GPERF_ENUM(x)
#define GPERF_ENUM_NOCASE(x)
#define GPERF_CASE(...) __LINE__
#define GPERF_CASE_VALUE(v, ...) (v)
#define GPERF_CASE_NONE 0
#endif
{{end}}
`
