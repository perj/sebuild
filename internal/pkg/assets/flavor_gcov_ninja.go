// Copyright 2018-2019 Schibsted

package assets

const FlavorGcovNinja = `
cwarnflags=-Wall -Werror -Wshadow -Wwrite-strings -Wpointer-arith -Wcast-align -Wsign-compare -Wformat-security -Wmissing-declarations $warncompiler
# Additional flags for building with gcov support
gcov_copts=-fprofile-arcs -ftest-coverage
gcov_ldopts=-fprofile-arcs -ftest-coverage -lgcov
`
