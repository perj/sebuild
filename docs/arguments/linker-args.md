# Linker Specific Arguments

All descriptors using the linker accepts these arguments. The standard
descriptors that use these are [PROG](../descriptors/prog.md),
[MODULE](../descriptors/module.md), [LIB](../descriptors/lib.md)
[LINKERSET_LIB](../descriptors/lib.md) and
[TOOL_PROG](../descriptors/tool-prog.md).

## cflags, cwarnflags, conlyflags, cxxflags

    cflags[-O3 -funroll-loops]
    cwarnflags[]
    conlyflags[-Wno-dump-warning]
    cxxflags[-fno-exceptions]

These set the corresponding ninja variables. Those are then used in the
rules files when compiling C and C++ files.

The default values are listed on the page about
[compiler flags](../compiler-flags.md).

## copts

    copts[\`php-config --includes\` -fno-strict-aliasing]

Options passed to C and C++ compilers. Mostly optimizations, warnings and such.
Does not override the default flags.

Instead of doing like the example shows consider setting a variable to the
output of php-config in the [config_script](../descriptors/config.md#config_script)
and then use that variable. E.g.

    copts[$php_config_output]

Then php-config isn't run for each compilation but only when generating the ninja
files.

## no_analyse

Disables the static analyser for this descriptor. Useful when it's made from
third party sources and you won't modify it regardless. Use with an empty
value.

    no_analyse[]

## go_noinit

If this is used together with Go files in [srcs](srcs.md) the Go runtime won't
be initialized automatically. This is needed if you have code that forks during
startup as the Go runtime can't survive forks.

If using this you must manually load the Go runtime before calling any Go
functions. To do that call the symbol named like this as a function
taking argc, argv and environ:

	echo _rt0_`go env GOARCH`_`go env GOOS`_lib

The sister-project [sebase](https://github.com/schibsted/sebase) has a wrapper
called `init_go_runtime` for making this more convenient.
