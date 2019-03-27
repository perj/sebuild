# incdirs

    incdirs[/usr/include/postgresql]

Paths where include files can be found. What you'd normally add with -I, but
without -I. The reason this isn't part of copts is because without special
handling this could end up with very long lines of repeated include
directories, now the script can resolve include paths and get the ordering
right. You can use relative paths here which are relative to the Builddesc
containing `incdirs`.

