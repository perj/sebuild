// Copyright 2018-2019 Schibsted

package assets

const DefaultsNinja = `
# Default values for variables often overridden by flavor or compiler rule
# files.
cflags=-g -pipe -O2 -D_GNU_SOURCE -fvisibility=hidden -fstack-protector
# $warncompiler doesn't really work here as the compiler rule file is included
# after this one, but included as an example.
cwarnflags=-Wall -Wshadow -Wwrite-strings -Wpointer-arith -Wcast-align -Wsign-compare -Wformat-security -Wmissing-declarations $warncompiler
conlyflags=
cxxflags=
analyser_flags=--analyze -Xanalyzer -analyzer-output=html -Xanalyzer -analyzer-disable-checker -Xanalyzer deadcode.DeadStores
`
