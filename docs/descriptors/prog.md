# Normal programs - PROG

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
