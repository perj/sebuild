# Writing Builddesc Files

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
`TOOL_PROG`, `LIB`, `MODULE`, `INSTALL`, `TOOL_INSTALL`.
There are a few more descriptors than these and plugins can also add more
descriptors.

All descriptors except `COMPONENT` and `CONFIG` require a name and at least one
argument. Name and arguments are separated by spaces.

Arguments are usually lists of elements. The elements are separated by
whitespace. Some arguments can only contain one element, this is clarified
in the argument documentation. Arguments are things like: srcs, includes, libs,
deps, etc.

Descriptors and arguments are listed on the main [index page](README.md).
