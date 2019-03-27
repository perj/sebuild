## libs

    libs[platform_search z stdc++]

In `LIB` this establishes the other libraries that this library depends on.
This makes everything else link to those other library when they link to this
library.

For all others this is a list of libraries that need to be linked to. `MODULE`
automatically links to the PIC versions of libraries if necessary. As a special
case, if the name of the library contains a `/` or `$` the script assumes that
it's a special case library from a full path or a predefined variable and links
to it directly instead of through -l.

