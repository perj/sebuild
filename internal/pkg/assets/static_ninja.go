// Copyright 2018-2019 Schibsted

package assets

const StaticNinja = `
# Static rules for certain targets. This file is included once per flavor,
# unlike rules.ninja, which is a global include.

# Create a build version include file and a handy depedency if for things
# that need to rebuild if the build version changes.
build $incdir/build_version_$buildversion.h: build_version
build $incdir/build_version.h: install_conf $incdir/build_version_$buildversion.h

# Generate the variables for compiling .in files.
inconf = $buildtools/in.conf
build $inconf: inconfig $configvars $inconfig | $buildvars
`
