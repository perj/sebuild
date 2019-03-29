# Libraries - LIB and LINKERSET_LIB

     LIB(platform_util
             srcs[strings.c memory.c bits.c]
             includes[charmap.h memory.h bits.h]
             libs[hunspell]
     )

`LIB` describes how to build a library. Libraries end up in obj/flavor/lib/ and
are only built when something links with them. An interesting thing to note
here is that Sebuild generates targets for normal and PIC targets for all
shared libraries and the PIC targets get only built for the libraries that get
linked into modules.

`LINKERSET_LIB` is similar to lib. Instead of creating a normal library
however, it creates a partial linked object file.  This has the effect of
including all the symbols in the final binary, instead of only the referenced
ones, which is required to access otherwise unreferences linkerset entries.

## Arguments

`includes` and `libs` are valid arguments here. The includes argument
specifies which include files are the external interface for this library and
those get installed in obj/flavor/includes/. The libs argument works like
elsewhere, but internally it's a bit special.
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
