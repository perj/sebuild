# srcs

    srcs[foo.c bar.cc]

Source files to build a program, library, module, tool or configuration file.
Sebuild automatically knows what to do with the following extensions:

* c - C source
* cc, cxx - C++ source
* go - Go source
* gperf - gperf source.
* gperf.enum - enumerated gperf source.
* in - to be processed as an in-file.
* ll - C++ lex
* yy - C++ yacc

Additional extensions can be added by plugins. Failing that
[specialsrcs](specialsrcs.md) can be used instead to manually specify what rule
compiles the source and what it generates.

The sources generate intermediate files that are put in the build/obj/ directory
matching the source path and flavor. They are further added again to be part
of the final product, unless anything else is noted below.

Most sources are only recognized in descriptors that use the linker, see
[Linker Specific Arguments](linker-args.md).
The main exception is the in extension which is often used in the
[INSTALL](../descriptors/install.md) descriptor and elsewhere.

## C and C++ Sources

These are compiled into object files, replacing the `c`, `cc` or `cxx` extension
with `o`. Those are further linked into the final product.

## In Sources

Also called in-files. A special script is run on these, using the
`build/obj/flavor/tools/in.conf` file to replace `%VARIABLE%` with the value of
`VARIABLE` from that file. It's an error to have undefined variables in the
in-files. To generate a `%` character use `%%`.

Entire sections can also be enabled/disabled based on the current flavor.
Text between `%FLAVOR_START%` and `%FLAVOR_END%` marker lines are removed
unless FLAVOR matches the uppercased version of the current flavore. All
markers are also removed.

The generate file has the same file name but with `.in` removed. This file is
NOT further processed but can be used e.g. in a
[conf](../descriptors/install.md#conf) or
[scripts](../descriptors/install.md#scripts) argument.

The script for generating the in.conf file is set using the `inconfig`
ninja variable. It defaults to `$buildtooldir/scripts/invars.sh`. If you change
it (e.g. via a [configvars file](../descriptors/config.md#configvars)), make
sure to source the original to access its functions and the configvars values.
Use a pair of lines like

    source $buildtooldir/scripts/invars.sh "$@"
    depend $buildtooldir/scripts/invars.sh

## Go Sources

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

## Gperf Sources

These are compiled with gperf to generate header files. The flag `-L ANSI-C`
is added but furter options should be put in the gperf file itself. The generated
file has the gperf extension removed and .h added.

## Gperf Enum Sources

No longer really needed, as an improved [specialsrcs](specialsrcs.md) is available,
but these can still be used to generate a Gperf header file that match on the strings
contained in this file. It's suggested to inspect the generated header to see how
these works. The generated file has the .gperf.enum extension removed and .h added.

## Flex and yacc Sources

The extensions `.ll` and `.yy` are passed to flex and bison respectively. They
will be given flags to generate C++ code.
