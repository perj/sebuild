#!/bin/sh
# Copyright 2018 Schibsted

# This script is necessary because there is no way to convince swig
# to output the module file with another filename. So we create a
# temporary directory of junk, generate everything in there and then
# rename as necessary.

T=`mktemp -d`
fail() {
	rm -rf $T
	exit 1
}

IN=$1
OUT=$2
shift 2

NAME=`grep '%module' $IN | awk '{print $2}'`

swig "$@" -outdir $T -o $T/garbage $IN || fail

mv $T/$NAME.* $OUT || fail

rm -rf $T
exit 0
