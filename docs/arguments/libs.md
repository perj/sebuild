## libs

    libs[platform_search z stdc++]

Libraries the descriptor depends on. This can be either libraries created via
the [LIB and LINKERSET_LIB descriptors](../descriptors/lib.md) or external
libraries that will be passed with the `-l` flag to the linker.

While it's mostly transparent, in `LIB` this establishes the other libraries
that this library depends on.  This makes everything else link to those other
library when they link to this library.

`MODULE` automatically links to the PIC versions of libraries if necessary. As
a special case, if the name of the library contains a `/` or `$` the script
assumes that it's a special case library from a full path or a predefined
variable and links to it directly instead of through -l.
