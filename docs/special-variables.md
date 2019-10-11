# Various variables like incdir & co

There's a bunch of variables defined inside the ninja files that can be
relevant. It will work to use them in Builddesc since they will get expanded in
the generated ninja file.

* `$incdir` - directory with all the includes installed. Mostly used in
  dependencies if you need to depend on a particular include file:
  `deps[$incdir/build_version.h]`
* `$libdir` - directory with all the libraries.
* `$builddir` - the root of all ninja and intermediate files.
  (`build/obj/dev` or such).
* `$destroot` - the root installation directory. (`build/dev` or such).
* `$buildtools` - directory where the tools are if you need an explicit
  dependency or refer to a tool in a extravars file.
* `$objdir` - Set for each descriptor. The directory where intermediate
  files are placed.
* `$buildflavor` - The flavor name.
* `$buildversion` - The build version number calculated from the buildversion
  script.

All active condtions are also ninja variables with the value 1.
