# Schibsted Search Engineering Build #

Copyright 2018 Schibsted

## Intro ##

This is a tool to generate a Ninja build solution and compile using ninja.

Ninja is a tool developed by Evan Martin at Google and is used to build Chrome
as fast as possible. As a Make replacement it's slightly quirky, it doesn't
really have much logic other than handling dependencies correctly, and you need
an external tool to generate ninja files for ninja to be usable. It's possible
to write them by hand, but it gets very chatty and unreadable fast.

The big advantage of ninja is that it forces you to think about what you're
doing and it is bloody fast.

## Origins ##

This software was originally developed by members of the Schibsted Coordination
Team to build what was known as the Blocket Platform, which we maintained
within the organization from 2010 to 2017. Ninja had just reached version 1.0,
and we wanted something that could build our tree faster than our existing GNU
Make solution.

At that point in time, tools like Bazel and Tundra weren't publicly available,
so the options were slim.

Initially a perl-script was produced. This was eventually rewritten into Go.

This documentation was originally written for internal use, and while it's been
edited for the public release, it will no doubt still reflect those origins.

### Authors ###

The original design and implementation was done by
[Artur Grabowski](https://github.com/art4711).

Prior to the public release the Go version was primarily written by
[Per Johansson](https://github.com/perj) with contributions from
[Eddy Jansson](https://github.com/eloj),
Artur Grabowski,
[Daniel Gustafsson](https://github.com/danielgustafsson),
[Claudio Jeker](https://github.com/cjeker),
[Fredrik Kuivinen](https://github.com/frekui),
[Daniel Melani](https://github.com/dmelani) and
[Sergio Visinoni](https://github.com/piffio).

## About this documentation ##

* "The basics" describes how to setup and use the build system day to day.

* "High level concepts" is a mess, but describes dependencies, libraries and
  how the file tree is organized.

* "Writing Builddesc files" and on describes how to actually write Builddesc
  files to build new components.

## The basics ##

You'll need to have ninja in your $PATH, you can get ninja from
[github](https://github.com/ninja-build/ninja).

If you want to install through packages, beware that there are other unrelated
software projects out there called 'ninja'.

If this document gets out of date, the tools we have are made roughly around
the time of ninja 1.0, and known to work up to ninja 1.8.

### How to run ###

The basic way to build a tree is simply to invoke `seb`.

    $ seb

This will compile everything you need.

The first thing seb does is finding the top directory of your project. To do
this it walks the directory tree upwards and looks for a Builddesc.top or
Builddesc file. If a Builddesc.top is found, that is used as the top level
directory, otherwise the topmost Builddesc is used. Since Builddescs are
somewhat recursive it was determined that using the topmost one makes most
sense.

After changing to this directory, seb will generate ninja files and then
execute ninja.

If you changed some file and want to recompile, just run `seb` again. It costs
50ms extra to run seb compared to just ninja directly, it's a lot less typing
and you'll never lose dependencies. Reading this sentence cost you more time
than you'd ever save skipping seb.

If you're not sure if you changed anything, you can still recompile and see
what happens. An empty compilation of the platform code is 100ms. Since seb
finds the top directory you should be able to run it from anywhere within the
tree.

## High level concepts ##

### What to build ###

Components of the system are described by Builddesc files. Those files get
parsed by the `seb` program and translated into ninja files. Seb does most of
the heavy lifting for figuring out dependencies.

### Dependencies ###

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

### Libraries ###

Most dependencies on generic components are expressed through libraries. If a
source file is used by multiple binaries, it needs to end up in a library. This
includes highly localized sources like for example the files shared by
`redis-server` and `redis-cli`. The libraries we build are all static, so there
shouldn't be much size or performance overhead. But one thing this means is
that you can't have defines controlling certain behaviors of the build. It is
often trivial to fix the existing defines, so it shouldn't much of a problem.

### Build directory ###

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

## Writing Builddesc files ##

### General syntax ###

A Builddesc file contains one or more build descriptors, and maybe some
comments. Example:

    # Let's build libplatform_foo
    LIB(platform_foo
       srcs[src1.c src2.cc]
       includes[foo.h]
    )
    # Build program "bar" with libplatform_foo
    PROG(bar
       srcs[main.c]
       libs[platform_foo]
    )

Comments start with '#' and continue to the end of the line.

Build descriptors have the general form of:

    DESC(name argument1[element1 element2 element3] argument2[el])

`DESC` is the descriptor, for example: `CONFIG`, `COMPONENT`, `PROG`,
`TOOL_PROG`, `LIB`, `MODULE`, `INSTALL`, `TOOL_INSTALL`. These will be
described later on. There are a few more descriptors than these and plugins can
also add more descriptors.

All descriptors except `COMPONENT` and `CONFIG` require a name and at least one
argument. Name and arguments are separated by spaces.

Arguments are usually lists of elements. The elements are separated by
whitespace. Some arguments can only contain one element, it will become obvious
later in the documentation which ones those are. Arguments are things like:
srcs, includes, libs, deps, etc.

### Descriptors ###

#### Global configuation - CONFIG ####

To configure what's being built and how you need something like this:

    CONFIG(
        buildversion_script[scripts/build/buildversion.sh]
        flavors[dev release]
        configvars[
            scripts/build/config.ninja
	]
        rules[
            scripts/build/gcc.ninja
            scripts/build/rules.ninja
        ]
        extravars[scripts/build/toolrules.ninja scripts/build/static.ninja]
        buildpath[build]
        buildvars[copts cflags cwarnflags conlyflags cxxflags]
        ruledeps[in:$inconf,$intool]
        prefix:prod[opt/blocket]
        config_script[scripts/build/config_script.sh]
    )

This configures some aspects of the build system.

* `buildversion_script` - script that outputs one number which is the version
  of what's being built. It's highly recommended that the version number is
  unique for at least every commit to your repository.
* `compiler` - Override the compiler used, set it to the C compiler, C++ one
  will be guessed with some heuristics.
* `flavors` - Various build environments needed to build your site. The usual
  is to build dev and release.
* `configvars` - Global ninja variables. Can be used to set some ninja variables
  such as the default gopath. Passed to invars.sh to also generate variables
  there, must thus contain only variable assignments.
* `rules` - Global compilation rules. These ninja files gets included globally.
* `extravars` - Per flavor-included ninja files. This means they can depend on
  the variables defined in the flavor files. Can be flavored.
* `buildpath` - Where the build files are put. See other sections of this
  document to see how files are organized.
* `buildvars` - attributes in other build descriptors that are copied into
  ninja files as variables. As in this example, those are various variables we
  want to be able to specify in build decsriptors that override default
  variables.
* `ruledeps` - Per-rule dependencies. Targets built with a certain rule will
  depend on those additional target. In this example everything built with `in`
  will also depend on `$inconf` and `$intool`.
* `prefix:flavor` - Set a prefix for the installed files for the specified
  flavor.
* `config_script` - Run a script whenever seb is run and parse its
  output as variables or conditions.
* `compiler_rule_dir` - Directory containing variables for a compiler variant
  (such as gcc or clang).
* `flavor_rule_dir` - Directory containing default variables for a flavor, in
  addition to those set here.
* `compiler_flavor_rule_dir` - Directory for variables specific to both
  compiler and flavor, if any.

The defaults are hard-coded to have buildpath set to `build` and one flavor -
`dev`. The environment variable `BUILDPATH` can also be used to set the
buildpath, but the config option has precedence.  Additionally,
`compiler_rule_dir`, `flavor_rule_dir` and `compiler_flavor_rule_dir` defaults
to respective subfolder insider the seb rules directory.

There can be only one `CONFIG` directive in the whole build system and it has
to be before any `COMPONENT` directives. In practice this means that it has to
be first in the first Builddesc or not at all.

##### Config script #####

The script is run when seb is run, and the output is parsed.  Any line
containing a equal sign (`=`) is added as a ninja variable.  This can be used
to e.g. run php-config only when seb is run instead of each call to the
compiler.

Any non-empty line output without an equal sign will be considered a condition
to activate, conditions are described in more detail further down.

##### Compiler #####

You can switch the compiler used in CONFIG. The available choices are gcc or
clang:

	compiler[clang]

You will need respective compiler to be installed, obviously.

#### Top level - COMPONENT ####

The top level Builddesc contains something like this:

    COMPONENT(flavors[dev release] [
         bin/foo
         bin/bar
         regress
         scripts/db
    ])

`COMPONENT` is a special build descriptor that doesn't really follow the
general rules. It doesn't have a name and has one argument with no name and a
list of directories where we will find further Builddesc files describing other
components of our build. It is possible for one or more of those directories to
contain a Builddesc with `COMPONENT` in it, but be very careful with that. It's
much more transparent and readable to have a long list of everything built
instead of a magic tree where everything is hidden.

The only other argumnet for `COMPONENT` build descriptors is `flavors`. See
flavors documentation below for the semantics of that. If `flavors` is omitted,
the components will be included in all flavors.

#### Normal programs - PROG ####

A program we want installed as part of the distribution can look something like
this:

    PROG(prog1
        srcs[
            prog1_parser.yy prog1_lex.ll prog1.cc sock_util.c parse_node.cc
        ]
        libs[platform_prog1]
    )

This specifies that our program is called "prog1", the binary ends up in bin/,
it has a bunch of source files and links to one library.

The interesting part here is that the sources are in four different languages.
We will figure out how to build everything and generate the right build
directives for ninja. We will also figure out that since at least one of our
source files is C++, the final linking of the binary will be done correctly.

Another thing worth noting is that one of our source files is yacc and those
always generate an include file. It's quite likely that lots of other source
files will include it. Instead of specifying a dependency on it, `seb` will
generate a general dependency so that no other files are compiled until yacc
has generated our include file. This might not be the most performance optimal
thing to do, but it's the most convenient.

#### Build tools - TOOL_PROG ####

     TOOL_PROG(tool1
             srcs[tool1.c]
             specialsrcs[mkrules:rules.txt:rules.h]
             deps[tool1.c:rules.h]
             libs[platform_util]
     )

`TOOL_PROG` is just like `PROG`, with two differences. Tools get installed
under obj/flavor/tools/ and instead of like `PROG` - being part of the default
build target tools get built only if something depends on them.

#### Go programs - GOPROG ####

    GOPROG(foo
    )

When used, the current directory will be compiled into a go binary.  You can
add the parameter nocgo[] to disable cgo for this program only. Other common
parameters also work, for example `libs[]` can be useful when compiling a Go
program linking with C libraries.

To build go programs, you should add a gopath variable in a configvars file.
This will be used to set the GOPATH environment variable while building, and
will override any such set by the environment.  If not set in configvars, it
will however default to the GOPATH environment variable.

Unlike the normal GOPATH, the gopath config variable can use paths relative to
the project root, they'll be changed into full paths by gobuild.sh and
invars.sh.

#### Go tests - GOTEST ####

    GOTEST(foo
    )

Add this descriptor to the same Buildesc as GOPROG and any other go packages
where you wish to use ninja to run `go test`. The advantage of using ninja here
is that it adds include paths etc. for cgo tests that can otherwise be difficult
to figure out.

These tests aren't built by the default ninja run, for each GOTEST descriptor a
target `build/flavor/gotest/name` is created which you can give to ninja to run
the test. In addition, you can also run `go test -cover` to generat an html file
with this target:

    build/<flavor>/gocover/<name>-coverage.html

And also `go test -bench` with

    build/<flavor>/gobench/<name>

You can override the default of running all benchmarks
by adding a `benchflags` argument to the `GOTEST` directive
and putting a regexp there, possibly also adding additional
flags.

#### Build libraries - LIB and LINKERSET_LIB ####

     LIB(platform_util
             srcs[strings.c memory.c bits.c]
             includes[charmap.h memory.h bits.h]
             libs[hunspell]
     )

`LIB` describes how to build a library. Libraries end up in obj/flavor/lib/ and
are only built when something links with them. An interesting thing to note
here is that `seb` generates targets for normal and PIC targets for all shared
libraries and the PIC targets get only built for the libraries that get linked
into modules.

`LINKERSET_LIB` is similar to lib. Instead of creating a normal library
however, it creates a partial linked object file.  This has the effect of
including all the symbols in the final binary, instead of only the referenced
ones, which is required to access otherwise unreferences linkerset entries.

`includes` and `libs` are valid parameters here. The includes argument
specifies which include files are the external interface for this library and
those get installed in obj/flavor/includes/. The libs argument is special.
Since we're building a static library, it can't be linked with other libraries
(unless they are static, but there's madness there). Instead `libs` create
dependencies that make programs built with this library to also link with those
other libraries.  So if you link your program with `platform_util` from this
example, you don't need to specify that it needs to link with hunspell, that
will happen automatically.

Another thing worth noting is that programs that link to this library will get
an automatic dependency on the includes from this library, so the programs
source files won't be compiled until the include files for this library are
installed. This is also recursive, the program will depend on all the include
files from the libraries this library depends on.

If you wish to install the headers with a prefix directory, you can use the
`incprefix` parameter for this. For example, to add the prefix `sbp/` to each
header path, use `incprefix[sbp]`.

#### Modules - MODULE ####

     MODULE(php_module1
            srcs[php_module1.c]
            libs[platform_core]
            copts[-DCOMPILE_DL_MOD=1 $php_module_cflags]
     )
    
`MODULE` describes how to build various shared object modules. Those end up in
modules/. php, perl and apache modules are all built the same way. All modules
are added to the default build target.

#### Configuration and other files - INSTALL ####

     INSTALL(regress/indexer/conf
             srcs[isearch.conf.in]
             conf[isearch.conf]
     )
    
     INSTALL(bin
             srcs[create_proschema.pgsql.in]
             scripts[create_proschema.pgsql]
     )
    
Installs files into the destination directory specified in the name. There are
different installation arguments, "scripts" specifies executable scripts and
such, "conf" specifies non-executable files, "python" should only be used for
python scripts (they will be compiled after install). A thing worth noting is
that you can specify files that are copied straight from the source directory
and also configuration files that have been compiled somehow (like .in).  Seb
will figure out which source file is built where and where to copy it from.

Symlinks can also be created, using this syntax:

	INSTALL(/dir
		symlink[dst:src]
	)

The symlink `dst` is created in `build/flavor/dir` pointing to the `src`.
If `src` is not a full path it needs to be relative to the installation
directory, like normal for symlinks. There's no check that it points to any
valid file or directory. Symlinks are currently rebuilt based on the target
file, hopefully this will be fixed sometime in the future.

#### Non binary build tools - TOOL_INSTALL ####

     TOOL_INSTALL(perl_xs_compiler
             srcs[perl_xs_compiler.in]
             scripts[perl_xs_compiler]
     )

`TOOL_INSTALL` is just like `INSTALL` with one major difference: non binary
tools get installed under tool/ and the target argument is the script name
instead of the target directory. This descriptor is used to install build tools
that are either static scripts or autogenerated scripts from a `.in` file.
The `name` argument is currently ignored since the destination directory is
always `tool/`.

### Build descriptor arguments ###

#### enabled ####

Can be used to disable the descriptor unless a certain flavor or condition
matches. Use with an empty value.

E.g. `enabled::foo[]` will enable this describor if the `foo` condition is set,
but it will otherwise be disabled. You can use enabled multiple times to
specify different conditions.

#### flavors ####

    flavors[dev]

Limits the building of this descriptor to certain flavors. See below for more
explaination of flavors. Superseded by the enabled argument which does much
of the same and more.

#### srcs ####

    srcs[foo.c bar.cc]

Source files to build a program, library, module, tool or configuration file.
Seb automatically knows what to do with the following extensions:

* c - C source
* cc, cxx - C++ source
* go - Go source
* gperf.enum - enumerated gperf source
* in - to be processed by in.pl
* ll - C++ lex
* yy - C++ yacc

#### srcopts ####

You can add compile flags or other optiones on a per source bases.
Use the format `srcopts[src:flags]` in the Builddesc. For example, to
add `-Wno-error` while compiling buggy.c use `srcopts[buggy.c:-Wno-error]`.

This works for intermediate sources as well, if buggy.c was generated
(e.g. from buggy.ll), the flag is still applied when compiling the C file.

The value is added as the variable `$srcopts` to any ninja rule that
has the given source. Thus it's up to each rule to determine the semantics.
Currently it's setup for C/C++ compile rules to add srcopts as compilation
flags.

#### Go sources ####

Go sources are a bit special. When used, they will be compiled into a Go
c-archive that is then linked into the final binary.

To access your go code, use `import "C"` and then add `//export` to the
functions you wish to access from C. They will then be declared in `gosrc.h`
which can be included from your C source. This also works for any
packages you import, with the caveat that they're not added to `gosrc.h`
and must be declared manually in C.
The Go package(s) have access to all local headers for the C program.

Several restrictions apply to what functions can be exported, see
`go doc cmd/cgo`.

The Go runtime stops working when a process forks, and can't be reinitialized.
To workaround this in programs that fork at launch, the automatic loading of
the Go runtime at program initialization has been disabled. Instead you must
manually load it. To do that call the symbol names like this as a function
taking argc, argv and environ:

	echo _rt0_`go env GOARCH`_`go env GOOS`_lib

(this disabling the autostart of Go runtime should probably not be enabled by
default.)

#### libs ####

    libs[platform_search z stdc++]

In `LIB` this establishes the other libraries that this library depends on.
This makes everything else link to those other library when they link to this
library.

For all others this is a list of libraries that need to be linked to. `MODULE`
automatically links to the PIC versions of libraries if necessary. As a special
case, if the name of the library contains a `/` or `$` the script assumes that
it's a special case library from a full path or a predefined variable and links
to it directly instead of through -l.

#### Standard buildvars ####

These are standard in that they're commonly used and when defined, should be
used for the purpose stated. They're not necessarily required or even defined
per default.

##### cflags, cwarnflags, conlyflags, cxxflags #####

    cflags[-O3 -funroll-loops]
    cwarnflags[]
    conlyflags[-Wno-dump-warning]
    cxxflags[-fno-exceptions]

Override the default compilation (optimization settings mostly) and warning
flags. Don't do this unless you really can't fix the code.

##### copts #####

    copts[\`php-config --includes\` -fno-strict-aliasing]

Options passed to C and C++ compilers. Mostly optimizations, warnings and such.
Does not override the default flags.

##### ldopts #####

    ldopts[-fprofile-arcs]

Options passed to linker. Does not override the default flags.

#### conf ####

    conf[some.conf]

Configuration files to be installed.  Only valid in `INSTALL`.

#### scripts ####

    scripts[index_ctl]

Scripts to be installed.  Only valid in `INSTALL`

#### incdirs ####

    incdirs[/usr/include/postgresql]

Paths where include files can be found. What you'd normally add with -I, but
without -I. The reason this isn't part of copts is because without special
handling this could end up with very long lines of repeated include
directories, now the script can resolve include paths and get the ordering
right. You can use relative paths here which are relative to the Builddesc
containing `incdirs`.

#### specialsrcs ####

    specialsrcs[charmap:swedish.txt:charmap-swedish
                charmap:russian.txt:charmap-russian
                charmap:swedish-utf8.txt:charmap-swedish-utf8
    ]           
    
This argument specifies a method to compile a source file where the automatic
file type detection can't figure out how to compile. Each element in is of the
form: `<command>:<src>[,<src>]*:<target>[:<extravars>[,<extravars>]*]`

`command` is the ninja rule to build the target, `src` is a comma separated
list of sources, `target` is the resulting output. `extravars` is a comma
separated list of extra variables added to the build directive for ninja.

Notice that specialsrcs is mostly useless on its own, the script will not pick
up and do anything with the output from the script. You will need to specify
additional arguments that use the output somehow (usually `conf` or `srcs`).

#### extravars ####

    extravars[morevars.ninja]

Extra ninja variables needed to build this descriptor. `extravars` will be
included only in this build descriptor and can therefore reference generated
paths. Since rules are global and `extranvars` can be included multiple times
it can't define rules. See special section below to see how to deal with this.

#### deps ####

    deps[config.h mkrules.c:sections.h]

Extra dependencies that the script can't figure out by itself. Two forms:

* `<depender>:<dependency>`
  specifies that `<depender>` can't be built until `<dependency>`
  has been built.
* `<global dependency>`
  specifies a dependency for all source files in this descriptor.

#### srcdir ####

    srcdir[hunspell-1.2.11/src/hunspell]

Specifies a different directory where the source files can be found relative to
where the Builddesc is. Source files can always be given as a path, but if all
source files are somewhere deeper inside the tree, this makes `srcs` shorter.

#### destdir ####

    destdir[lib/perl5]

Normally PROG and MODULE installs in predefined target directories, but in rare
cases you might need to override this, for example for perl modules or to
install to libexec/ instead of bin/.

There are some special values for the first step in `destdir`. You write these
like a normal directory path but they will be translated by seb into other
paths:

* `dest_inc`: The default destination for includes.
* `dest_bin`: Same as bin/ but not hard coded. Used by default by PROG.
* `dest_tool`: The tool directory inside obj/flavor.
* `dest_lib`: The lib directory inside obj/flavor.
* `dest_mod`: Same as modules/. Used by default by MODULE.
* `destroot`: The build/flavor/prefix root directory.
* `flavorroot`: The build/flavor root directory without any prefix.
* `builddir`: The root intermediate directory, build/obj/flavor.
* `objdir`, `obj`: Translated to the intermediate objects directory for the
  current descriptor.

#### collect_target_var ####

    INSTALL(conf/a
        conf[x.conf y.conf]
        collect_target_var[all_confs a_confs]
    )
    INSTALL(conf/b
        conf[x.conf]
        collect_target_var[all_confs]
    INSTALL(conf/csums
        specialsrcs[
            checksum:$all_confs:allconfs.csum
            checksum:$a_confs:aconfs.csum
        ]
        conf[allconfs.csum aconfs.csum]
    )

Collects the names of final targets into one or more variables that can later
be used as a sources in other descriptors.

In this example `$all_confs` will contain `$destroot/conf/a/x.conf`,
`$destroot/conf/a/y.conf` and `$destroot/conf/b/x.conf`, `$a_confs` will
contain `$destroot/conf/a/x.conf` and `$destroot/conf/y.conf`. Notice that the
variables will be expanded when generating the ninja files to get proper
dependencies (and because ninja treats variables as one file, never a list).

As a special rule, all targets generated by GOTEST are collected like with
`collect_target_var`, but in special variables:

 * `$_gotest` collects test targets
 * `$_gobench` collects benchmark targets
 * `$_gocover` collects coverage targets (the html output)

This makes it easier to invoke all registers go tests.

#### PHP ####

    php[random.php]

PHP code to be installed and syntax checked using the PHP lint command.
Only valid in the INSTALL descriptor.

### Separate top-level Builddesc ###

When you want to do a project that both builds separately and can be part of a
larger project, you can't have a CONFIG in your Builddesc file, and possibly
you also want different COMPONENTs. To solve this, at the top-level only, the
file `Builddesc.top` is preferred over Builddesc. You can put your CONFIG etc.
there. To also include Builddesc, add `.` as a component.

Example Builddesc.top:

```
CONFIG(
	flavors[dev release]
)

COMPONENT([
	.
])
```

## Builddesc fragments and includes ##

In some cases we need to build a binary multiple times with optional
components. In that case it's useful to have partial Builddesc files included
in the main Builddesc. The INCLUDE[] argument does this. The included fragment
contains arguments to the build descriptor.

Example:

* in modules/mod_a/Builddesc.inc:

           srcs[mod_a.c request.c extrafunctions.c]
           libs[platform_util]
 
* in regress/mod_a/Builddesc:

         MODULE(mod_a
                INCLUDE[../../modules/mod_a/Builddesc.inc]
                srcs[local.c]
         )
   
Two things are worth noting here. The same arguments can be present in both the
included fragment and in the main descriptor.  Source files paths for the
fragment are relative to where the fragment lives.

The path specified in `INCLUDE` can either be relative to the top directory as
all other paths or if it starts with a `./` or `../` will be relative to the
directory of the Builddesc where the `INCLUDE` argument is.

## Flavors ##

The build system supports building various flavors of the same code. The
typical configuration will build 'dev' and 'release' versions of the code.

There are three ways the flavors can be differentiated between each other.

### .in flavors ###

The typical way to generate the variables for .in files is to allow the script
that generates the input to include the file with various build variables
defined. These will be defined inside `build/obj/<flavor>/buildvars.ninja` and
include definitions for the flavor being built and various directories.

### COMPONENT flavors ###

`COMPONENT` descriptors can contain the argument `flavors` which contains the
list of flavors this component will only be built for. The default is to build
for all flavors.

If you want some components to be only built for dev, you can do:

    COMPONENT(flavors[dev]
     [
   	regress/component1
   	regress/component2
     ]
    )

### flavor modifiers on arguments ###

Almost all arguments can be modfied with the flavor with a colon and the flavor
name after. Like this:

    srcs[a.c b.c]
    srcs:dev[c.c]

The result will be a merge between the unflavored arguments and the arguments
for this flavor (`srcs[a.c b.c c.c]` in the example).

### flavors argument ###

All build descriptors (except `CONFIG`) also support a `flavors` argument. This
argument lists the flavors where that build descriptor applies.

## Conditions ##

In addition to flavors, build conditions can also be used to select what's
built.  A condition is a simple truth check -- the presence or absence of a
condition is checked. If all checks are true the rule is added (merged) into
the build result. If any condition is false the rule is skipped.  Possible
example could look like this:

	LIB(platform_random
		srcs::linux,x86_64[*.c]
		includes[rand.h]
	)

Sources are only added if both conditions `linux` and `x86_64` are set.  Two
conditions are always added: lowercase output of `uname` and output of `uname
-m`. For example `linux` and `x86_64`.

Conditions can also be added in `CONFIG` like so:

	CONFIG(
		conditions[a b]
	)

adds conditions a and b. They can also be passed on the command line, running
`seb -cond foo` adds the condition `foo`.

Finally they can also be added by the config script, see the separate section.

Conditions can use letters, numbers and underscore (`_`), no other characters
are allowed.

## Various special rules and definition files ##

### build_version.h ###

The version of the build is always written with a `#define BUILD_VERSION` into
the `build_version.h` file. The build version is generated by the script in
`buildversion_script` CONFIG option. If you include this file in your
source you'll have to add an explicit dependency on it.

Another way to get the build version is through .in files. There is an
automatic dependency there and you don't need to do anything.

### invars.sh ###

The variables for .in files are generated into `build/obj/<flavor>/tools/in.conf`
with a script. The generated file must be of the form `KEY=value` and nothing
else, no comments, no extra whitespace, no nothing. This allows us to include
it from both Makefiles and ninja files.

The variables are generated by the shell script in
`scripts/invars.sh`.

### static.ninja ###

The file `rules/static.ninja` contains certain static targets we need to build
and don't really have other ways to express them. Currently it contains the
rules for `build_version.h` and `in.conf`.

### rules.ninja ###

`rules.ninja` contain the main rules for compilation (like how to invoke
compilers, how to link, etc.). Read the ninja documentation before changing
anything in `rules.ninja`. Be careful about failing rules which just simply
redirect some script output into the $out file. Make sure to remove the $out
file if the script fails. If you don't, the next build can get confused and not
rebuild the file even though it failed the previous time. make removes the file
for you, ninja doesn't. There are some examples in there how to do this, but
the safest is probably to wrap everything in a script.

### Various variables like incdir & co ###

There's a bunch of variables defined inside the ninja files that can be
relevant. It will work to use them in Builddesc since they will get expanded in
the generated ninja file.

* `$incdir` - directory with all the includes installed. Mostly used in
  dependencies if you need to depend on a particular include file:
  `deps[$incdir/build_version.h]`
* `$libdir` - directory with all the libraries.
* `$builddir` - the root of all ninja and intermediate files.
  (`build/obj/dev` or such).
* `$destroot` - the root installation directory. (`build/dev` or such).
* `$buildtools` - directory where the tools are if you need an explicit
  dependency or refer to a tool in a extravars file.

### Multiple targets with the same name ###

Since seb resolves dependencies and build rules according to the names of
generated files, special care needs to be taken when you end up with multiple
rules creating several different files with the same names in different
directories. In general this can only happen when you do something like this:

    INSTALL(conf
        srcs[conf.txt.in]
        conf[conf.txt]
        specialsrcs[magic:conf.txt:foobar]
    )
  
In the srcs argument `conf.txt.in` generates `conf.txt`, but the `conf`
argument does the same. Then the `specialsrcs` depends on `conf.txt` and it's
not really clear which one. The only safe heuristics we can apply to resolve
this is when one of the targets is an install target (`conf` or `scripts`). In
that case the dependency chain gets transformed from:

    $src/conf.txt.in -> $objdir/conf.txt -> $conf/conf.txt

to:

    $src/conf.txt.in -> $objdir/TMP_BUILDconf.txt -> $conf/conf.txt

So if you look in the object directory for the generated file, it might have a
different name.

### Writing your own rules ###

One behavior of ninja that is important to understand is that rules are global
while variables can be local (read up ninja documentation to understand
variable scoping). This means that all rules files can only be included once.
If you compile a tool that will then be used inside a rule you must therefore
do something like this:

 * `Builddesc`:

          TOOL_PROG(mytool
              srcs[...]
          )
          INSTALL(conf
              conf[fooconf]
              specialsrcs[myrule:fooconf:in.foo]
              extravars[myvars.ninja]
          )

 * `<global>/rules.ninja`:

          rule myrule
              command=$mytool $in $out

 * `myvars.ninja`:

          mytool=$buildtools/mytool

 * `<topdir>/Builddesc`
          CONFIG(... ruledeps[... myrule:$mytool] ...)

This uses the fact that rules have to defined globally, but the variables that
the rules are using are scoped. So we can get away with having a local scope on
the definition of the tool we use.

Notice the special top level CONFIG addition. This tells ninja that all targets
built with the new rule should depend on the tool. If the tool changes,
everything built with the tool should be rebuilt. If you depend on a tool and a
configuration file you can specify multiple dependencies for the rule separated
by a comma, like this `ruledeps[...  myrule:$mytool,$myconfigfile ...]`.

### Globbing ###

Whenever an argument accepts real source files those source files can also be
given with a `*`-glob. The build script will keep dependencies correct and
react when the source directory changes. Globbing only works on real source
files that exist in source directories. Any files generated by intermediate
steps will still have to be specified with full name. The current arguments
that support globs are: `srcs`, the source part of `specialsrcs`, `conf` and
`scripts`. Example:

    TOOL_PROG(tpc
        srcs[*.yy *.ll *.cc]
        libs[pcre platform_util]
        copts[-Wno-unused]
    )

Globbing is guaranteed to remove duplicate source files and preserve order.

## Custom library paths ##

If you have libraries in custom places, e.g. `/usr/pgsql-x.x`, the best way to
find them is to create a `compile.sh` to setup the paths:

    export LIBRARY_PATH=/usr/pgsql-x.x/lib:$LIBRARY_PATH
    export CPATH=/usr/pgsql-x.x/include

Maybe even

    export PATH=/usr/pgsql-x.x/bin:$PATH

After you've set those variables run seb from the script.

## Static analyser ##

You can run the clang static analyser on your code. This is setup
automatically, but disabled by default since it's very slow. Run

    seb build/<flavor>/analyse

to generate a report in that directory.
