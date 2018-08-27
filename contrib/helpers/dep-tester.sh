#!/bin/sh
# Copyright 2018 Schibsted

# This is a brute force ninja dependency tester.
# We generate the ninja files, extract all the targets from them and then build each target in a fresh build tree.
# Use this once in a while to make sure that all the dependencies are good enough to at least see if the final targets
# build correctly.

test -z "$BUILDPATH" && BUILDPATH=build
: ${BUILDBUILD:=seb}

rm -rf $BUILDPATH

$BUILDBUILD $BUILDBUILD_ARGS

ninja -f $BUILDPATH/build.ninja -n -t targets | grep -vE '(analyse|gotest|gobench)' | sed 's/:[^:]\{1,\}$//' > $BUILDPATH/test-targets

err=0
for a in `cat $BUILDPATH/test-targets` ; do
	echo TESTING $a
	rm -rf $BUILDPATH
	$BUILDBUILD $BUILDBUILD_ARGS
	ninja -f $BUILDPATH/build.ninja $a
	if [ $? != "0" ] ; then
		echo DEPENDENCY FAIL FOR $a >&2
		err=1
	else
		test -t 1 && printf '\033[3A\033[J'
	fi
done
test -t 1 && printf '\033[3B'
exit $err
