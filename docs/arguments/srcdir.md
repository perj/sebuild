## Set the Source Directory - srcdir

    srcdir[hunspell-1.2.11/src/hunspell]

Specifies a different directory where the source files can be found relative to
where the Builddesc is. Source files can always be given as a path, but if all
source files are somewhere deeper inside the tree, this makes `srcs` shorter.

This affects all paths in [srcs](srcs.md), [specialsrcs](specialsrcs.md),
[includes](../descriptors/lib.md#arguments) as well as the
[GOPROG](../descriptors/goprog.md) and [GOTEST](../descriptors/gotest.md)
descriptors.
