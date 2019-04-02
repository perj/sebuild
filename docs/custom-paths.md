# Custom Library Paths

If you have libraries in custom places, e.g. `/usr/pgsql-x.x`, these can be
setup in a [confivars files](descriptors/config.md#configvars) or other ninja
file. Add the variable `ldopts` as such:

    ldopts=-L /usr/pgsql-x.x/lib

For header files the best way might be to use the
[incdirs argument](arguments/incdirs.md) in the descriptor that needs
to find these. Another option is to set cflags in an
[extravars files](descriptors/config.md#extravars).

For more information about how the compiler and linker are invoked, see
[Compiler and Linker Flags](compiler-flags.md).
