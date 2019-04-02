## Set the Destination Directory - destdir

    destdir[lib/perl5]

Normally PROG and MODULE installs in predefined target directories, but in some
cases you might need to override this, for example for perl modules or to
install to libexec/ instead of bin/.

There are some special values for the first step in `destdir`. You write these
like a normal directory path but they will be translated by Sebuild into other
paths:

* `dest_inc`: The default destination for includes inside obj/flavor.
* `dest_bin`: Same as bin/ but not hard coded. Used by default by PROG.
* `dest_tool`: The tool directory inside obj/flavor.
* `dest_lib`: The lib directory inside obj/flavor.
* `dest_mod`: Same as modules/. Used by default by MODULE.
* `destroot`: The build/flavor/prefix root directory.
* `flavorroot`: The build/flavor root directory without the
  [prefix](../descriptors/config.md#prefix) if any.
* `builddir`: The root intermediate directory, build/obj/flavor.
* `objdir`, `obj`: Translated to the intermediate objects directory for the
  current descriptor.
