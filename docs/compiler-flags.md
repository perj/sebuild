# Compiler and Linker Flags

The flags used when the C/C++ compiler and linker are invoked can be
controlled by a number of ninja variables. The variables ending with
`flags` is meant to be set globally such as in a
[configvars file](descriptors/config.md#configvars) while the ones
ending with `opts` are meant to be set locally in a Builddesc.
These are then combined into the final flags used.

Note that if a flavor override sets a variable then that overrides
the configvars value. You have to use an
[extravars](descriptors/config.md#extravars) file instead in that case.

You can however override the flag arguments as well in a Builddesc,
in case the defaults do not work in that particular case. Another
option might however be to use the [srcopts argument](arguments/srcopts.md)
to set flags for one particular source file.

## Ninja Variables

### cc
Names the C compiler. Default value is decided based on the config
directive and detected installed compilers.

### cxx
Names the C++ compiler. Default value is based on the cc variable.

### cflags
Despite the name, this variable is used when compiling both C and C++ files.

Defaults to `-g -pipe -O2 -D_GNU_SOURCE -fvisibility=hidden -fstack-protector`
although that might be overwritten based on flavor.

### cwarnflags
Same a cflags but meant to contain only warning flags. Ued when compiling both
C and C++ files.

Defaults to `-Wall -Wshadow -Wwrite-strings -Wpointer-arith -Wcast-align -Wsign-compare -Wformat-security -Wmissing-declarations` but might be overridden by the flavor.

### flavor_cflags
The value of the flavored [cflags config argument](descriptors/config.md#cflags)
if any. Used when compiling both C and C++ files.

### cxxflags
Used when compiling C++ files only. Empty by default.

### copts
Set locally in the Builddesc with the [copts argument](arguments/linker-args.md#copts).
Empty by default.

### srcopts
Set locally in the Builddesc with the [srcsopts argument](arguments/srcopts.md).
Empty by default.

### ldopts
Used when running the linker. Can for example contain additional -L flags.
Empty by default and must be set globally despite the name.

### analyzer_flags
Used when running the static analyzer. Can't be overridden in Builddesc but
can be set in configvars and similar. It usually don't have to be changed.

Defaults to `--analyze -Xanalyzer -analyzer-output=html -Xanalyzer -analyzer-disable-checker -Xanalyzer deadcode.DeadStores`

## Compiler Specific Variables
Some ninja variables are set by including a file based on the compiler used.
It's mandatory that these files exist, but the directory where they are found
can be controlled via the
[compile_rule_dir config argument](descriptors/config.md#compiler_rule_dir).

### clang
Sets two values.

```
warncompiler=-Wno-parentheses-equality
w_no_self_assign=-Wno-self-assign
```

The `w_no_self_assign` variable can be used in the
[copts argument](arguments/linker-args.md) or [srcopts](arguments/srcopts.md)
if you wish to remove that warning.

The `warncompiler` value is added when compiling the dev flavor, see below.

### gcc
Sets a single value

```
warncompiler=-Wlogical-op
```

The `warncompiler` value is added when compiling the dev flavor, see below.

## Flavor Overrides

Some flavor names cause different default compiler flags. You might want
to consider this when naming your flavors. Or ignore it by changing the
[flavor_rule_dir config argument](descriptors/config.md#flavor_rule_dir).

These are included after configvars due to being flavor specific. Thus
they override values set in configvars. Use the
[extravars config argument](descriptors/config.md#extravars) instead if you
wish to override these without using the flavor_rule_dir argument.

### dev

Changes cwarnflags to add `-Werror` and also include the `warncompiler`
ninja variable set in the compiler specific file, see above.

### release
Changes `cflags` to `-g -pipe -O2 -D_GNU_SOURCE -fvisibility=hidden -fstack-protector`
and changes `cwarnflags` to simply be `-Wall`, disabling -Werror and several other
values.

### gcov
Uses the same `cwarnflags` as dev, but also adds two special variables also used
when compiling C and C++ files.

```
gcov_copts=-fprofile-arcs -ftest-coverage
gcov_ldopts=-fprofile-arcs -ftest-coverage -lgcov
```

## Compiler and Flavor Overrides

Finally it's possible to set variables on both the compiler and flavor used.
There are however no defaults for this, but if you set a directory with
the [compiler_flavor_rule_dir](descriptors/config.md#compiler_flavor_rule_dir)
you can add files there called `compiler-flavor.ninja` (e.g. `gcc-dev.ninja`)
which will be included if both the compiler and flavor match.
