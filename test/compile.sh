#!/bin/sh
# Copyright 2018 Schibsted

set -xe
test -z "$BUILDPATH" && BUILDPATH=build

CC='env cc' seb -condition cfoo -condition cbar
touch Builddesc # to make ninja invoke seb.
ninja -f $BUILDPATH/build.ninja

grep '# Flavors: regress' $BUILDPATH/build.ninja
grep '# Conditions:.*cbar, .*cbaz, .*cfoo' $BUILDPATH/build.ninja

grep -q gopath $BUILDPATH/regress/collect_test/hejsan.txt
grep -q gopath $BUILDPATH/regress/collect_test/other.txt
grep -q bar $BUILDPATH/obj/regress/lib/test

ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gotest/gopath
ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gocover/gopath-coverage.html

# This sleep is unfortunately necessary because ninja doesn't
# recognize timestamp differences smaller than one second.
sleep 1

# Test that we don't rebuild anything if nothing has changed
if (ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gocover/gopath-coverage.html | grep -q 'coverage to') ; then echo "gopath-coverage.html rebuilt without changes" ; exit 1 ; fi

# Then test that we do rebuild if something has changed
touch go/src/gopath/main_test.go
if ! (ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gocover/gopath-coverage.html | grep -q 'coverage to') ; then echo "gopath-coverage.html not rebuilt with changes" ; exit 1 ; fi

# Test of enabled argument
test -f $BUILDPATH/obj/regress/lib/enabled
! test -f $BUILDPATH/obj/regress/lib/disabled

[ -n "$NODEPTEST" ] || CC='env cc' BUILDBUILD_ARGS="-condition cfoo -condition cbar" ../contrib/helpers/dep-tester.sh
