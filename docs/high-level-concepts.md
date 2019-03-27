# High level concepts

## What to build

Components of the system are described by Builddesc files. Those files get
parsed by the `seb` program and translated into ninja files. Seb does most of
the heavy lifting for figuring out dependencies.

## Dependencies ##

The important thing to understand about the build system is that everything
must have proper dependencies. You can't get away with knowing that Make will
run the compilation in some order that happens to work most of the time. This
is important because it allows ninja to optimize the compilation very hard and
make it faster than you'd imagine was possible, and it allows ninja to only
rebuild the things you need.

To make sure that you don't cheat on your dependencies, there's a certain
amount of randomness in the build process. Things will not compile in the same
order every time. This has a disadvantage though - if you have multiple errors
in your compilation and fix one error, you might get the next error before you
see if your fix worked or didn't work. It's slightly annoying, but worth the
advantages.

How to express dependencies is described in the documentation of how to write
`Builddesc` files.

## Libraries ##

Most dependencies on generic components are expressed through libraries. If a
source file is used by multiple binaries, it needs to end up in a library. This
includes highly localized sources like for example the files shared by
`redis-server` and `redis-cli`. The libraries we build are all static, so there
shouldn't be much size or performance overhead. But one thing this means is
that you can't have defines controlling certain behaviors of the build. It is
often trivial to fix the existing defines, so it shouldn't much of a problem.

## Build directory ##

Everything during the build is contained into the build/ directory at top of
the tree, or wherever you wish by setting the `BUILDPATH` environment variable.
If you believe that you messed something up really badly and think that you
need a fresh build, nuke that directory and rebuild.

The organization of the build directory requires some explanation:

    $ ls build
    build.ninja obj dev <flavor0> <flavorN>
    $ 

The `build.ninja` file is the top level ninja file that describes how to build
everything. It includes various global configuration and rules files, contains
a target to rebuild the ninja files (with proper dependencies) and finally
includes the flavors we'll build.  The default flavor is `dev`. A flavor is
simply a named configuration, think 'dev' and 'release'.

As a convenience, near the top it contains a comment with enabled flavors and
conditions, allowing you to check that they match your expectations.

The obj directory also contains a subdirectory for each flavor. In there the
ninja files and intermediate build files are kept. More about that below.

    $ ls build/dev/
    bin  include modules regress
    $

This is the destination directory for the dev flavor. The goal here is
that this directory should contain all the final files used in the build. It
can for example be passed directly to your rpm or deb creation tool.

* `bin` and `modules` are the destination directories to where executables and
  modules end up.

* `regress` is the target for our INSTALL descriptors where we install various
  regress test configuration.

    $ ls build/obj/dev/
    bin  build.ninja  buildvars.ninja  include  lib  modules regress scripts
    tools  vendor

This is a rough mirror of the source tree where we drop compiled object files
and such. The idea is that nothing except ninja and the compiler should ever
touch things in here. That's why you don't need to know what's specifically in
there. Regardless, some of the subdirectories are special.

* `lib` contains all the libraries we've built that we're linking against.

* `include` contains all the include files that our libraries provide as an
  external API. This is mostly to avoid an include path hell. Instead of having
  an -I option to gcc for every library you happen to use and conflicting
  include files, all includes get put into that directory and if you manage to
  create two include files with the same name, instead of randomly including
  the wrong file half the time ninja will yell at you. You wouldn't want a
  ninja yelling at you, would you?

* `tools` are binaries that aren't part of the destination build proper; most
  likely other build-tools which are used to produce the build.
