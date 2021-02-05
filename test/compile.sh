#!/bin/sh
# Copyright 2018 Schibsted

set -xe
export BUILDPATH=build

CC="cc -std=gnu11" seb -condition cfoo -condition cbar
touch Builddesc # to make ninja invoke seb.
ninja -f $BUILDPATH/build.ninja

grep '# Flavors: regress, prod' $BUILDPATH/build.ninja
grep '# Conditions:.*cbar, .*cbaz, .*cfoo' $BUILDPATH/build.ninja

grep -q goarch $BUILDPATH/regress/collect_test/hejsan.txt
grep -q goarch $BUILDPATH/regress/collect_test/other.txt
grep -q bar $BUILDPATH/obj/regress/lib/test
grep -q fooval $BUILDPATH/regress/regress/infile/infile

# Check regress flavor for darwin 386 build
grep -q rt0_386_openbsd $BUILDPATH/regress/bin/goarch
# Check prod flavor is executable.
$BUILDPATH/prod/bin/goarch

$BUILDPATH/regress/bin/build_ctest
$BUILDPATH/regress/bin/gosrc_test
$BUILDPATH/regress/bin/gosrc_test_noinit

# Load the built go module.
$BUILDPATH/regress/bin/loader $BUILDPATH/regress/modules/gomod.so

ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gotest/goarch
ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gocover/goarch-coverage.html

# This sleep is unfortunately necessary because ninja doesn't
# recognize timestamp differences smaller than one second.
sleep 1

# Test that we don't rebuild anything if nothing has changed
if (ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gocover/goarch-coverage.html | grep -q 'coverage to') ; then echo "goarch-coverage.html rebuilt without changes" ; exit 1 ; fi

# Then test that we do rebuild if something has changed
touch goarch/main_test.go
if ! (ninja -f $BUILDPATH/build.ninja $BUILDPATH/regress/gocover/goarch-coverage.html | grep -q 'coverage to') ; then echo "goarch-coverage.html not rebuilt with changes" ; exit 1 ; fi

# Test of enabled argument
test -f $BUILDPATH/obj/regress/lib/enabled
! test -f $BUILDPATH/obj/regress/lib/disabled

# Test touching in future
seb $BUILDPATH/regress/regress/touchtest
sleep 1 # Bump stamp source.
seb $BUILDPATH/regress/regress/touchtest | grep -q 'no work to do'

[ -n "$NODEPTEST" ] || CC="cc -std=gnu11" BUILDBUILD_ARGS="-condition cfoo -condition cbar" ../contrib/helpers/dep-tester.sh
