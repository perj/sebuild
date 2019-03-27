# Linker arguments

All descriptors using the linker accepts these arguments. The standard
descriptors that use these are PROG, MODULE, LIB and TOOL_PROG.

## cflags, cwarnflags, conlyflags, cxxflags

    cflags[-O3 -funroll-loops]
    cwarnflags[]
    conlyflags[-Wno-dump-warning]
    cxxflags[-fno-exceptions]

Override the default compilation (optimization settings mostly) and warning
flags. Don't do this unless you really can't fix the code.

## copts

    copts[\`php-config --includes\` -fno-strict-aliasing]

Options passed to C and C++ compilers. Mostly optimizations, warnings and such.
Does not override the default flags.* [Linker Specific Arguments - 

## ldopts

    ldopts[-fprofile-arcs]

Options passed to linker. Does not override the default flags.
