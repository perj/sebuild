# Builddesc fragments and includes - INCLUDE

The INCLUDE argument allows to put part of a Builddesc in a different file.
This is useful when you want to build the same binary multiple times with
just a few different optional sources, e.g. linkersets.

In that case it's useful to have partial Builddesc files included
in the main Builddesc. The INCLUDE[] argument does this. The included fragment
contains arguments to the build descriptor.

Example:

* in modules/mod_a/Builddesc.inc:

           srcs[mod_a.c request.c extrafunctions.c]
           libs[platform_util]
 
* in regress/mod_a/Builddesc:

         MODULE(mod_a
                INCLUDE[../../modules/mod_a/Builddesc.inc]
                srcs[local.c]
         )
   
Two things are worth noting here. The same arguments can be present in both the
included fragment and in the main descriptor.  Source files paths for the
fragment are relative to where the fragment lives.

The path specified in `INCLUDE` can either be relative to the top directory as
all other paths or if it starts with a `./` or `../` will be relative to the
directory of the Builddesc where the `INCLUDE` argument is.
