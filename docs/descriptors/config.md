# Global configuation - CONFIG

There can be only one `CONFIG` directive in the whole build system and it has
to be before any `COMPONENT` directives. In practice this means that it has to
be first in the first Builddesc or not at all.

Sebuild is meant to have sane defaults and perhaps you will not need a
CONFIG directive at all, its entire contents is optional. The most common
options to add are `configvars`, `config_script` and `flavors`.

A config directive with full contents will look something like this.

	CONFIG(
		buildversion_script[sebuild/buildversion.sh]
		compiler[gcc]
		flavors[dev release]
		configvars[
			sebuild/config.ninja
		]
		rules[
			sebuild/gcc.ninja
			sebuild/rules.ninja
		]
		extravars[
			scripts/build/toolrules.ninja
			scripts/build/static.ninja
		]
		buildpath[build]
		buildvars[toolflags]
		ruledeps[
			in:$inconf,$intool,$configvars
		]
		prefix:release[usr/local]
		config_script[
			sebuild/config_script.sh
		]
	)

CONFIG contains custom arguments, described here. The list of arguments
on the index page does not apply to CONFIG. As a special exception, you _can_
use the [INCLUDE argument](../arguments/include.md) in CONFIG.

## `buildversion_script`
Script that outputs one number which is the version of what's being built. It's
highly recommended that the version number is unique for at least every commit
to your repository.
Defaults to `git rev-list HEAD 2>/dev/null|wc -l|xargs` (xargs remove
any formatting wc adds.)

## `compiler`

Override the compiler used, set it to the C compiler, C++ one will be guessed
with some heuristics. Defaults to CC env variables or by testing a few common
compilers if unset.

## `flavors`
Various build environments needed to build your site. The usual is to build dev
and release.

Defaults to `dev` only.

## `configvars`
A list of file names, relative paths.

The files should contain global ninja variables. This can be used to set some
ninja variables such as the default gopath. Passed to invars.sh to also
generate variables there, must thus contain only variable assignments, no
rules.

## `rules`
A list of file names, relative paths.

Global compilation rules. These ninja files gets included globally.
Defaults to empty list. The rules.ninja bundled with sebuild is however always
included as well, regardless of this value.

## `extravars`
A list of file names, relative paths.

Per flavor-included ninja files. This means they can depend on the variables
defined in the flavor files. Can be flavored.

## `buildpath`
Where the build files are put. See other sections of the documentation to see
how files are organized.

Defaults to the BUILDPATH environment variable or to `build` if unset.

## `buildvars`
Attributes in other build descriptors that are copied into ninja files as
variables. The variable names here can be added as arguments in any descriptor
and will be available to rules in your custom rules.ninja.

### `ruledeps`
Per-rule dependencies. Targets built with a certain rule will depend on those
additional target. In this example everything built with `in` will also depend
on `$inconf` and `$intool`.

There exists some default ruledeps, entries you put here will add to those. The
defaults are:

* `in:$inconf,$intool,$configvars`

### `prefix:flavor`
Set a prefix for the installed files for the specified flavor.
This argument must be flavored.

### `config_script`
Run a script whenever seb is generating ninja files and parse its output as
variables or conditions.

### `compiler_rule_dir`
Directory containing variables for a compiler variant (such as gcc or clang).
In this directory should be a file named `compiler.ninja` (e.g. gcc.ninja)
that will be a global include based on compiler used. Typically sets the
`warncompiler` ninja variable.

Defaults to the bundled `rules/compiler` directory.

### `flavor_rule_dir`
Directory containing default variables for a flavor, in addition to those set
here.
In this directory should be a file named `flavor.ninja` (e.g. dev.ninja) that
will be a flavor specific include. Typically sets variables such as `cflags`
and `cwarnflags`.

Defaults to the bundled `rules/flavor` directory.

### `compiler_flavor_rule_dir`
Directory for variables specific to both compiler and flavor, if any.
Included mostly for completeness, works similar to the arguments above.

Defaults to the bundled `rules/compiler-flavor` directory which is empty.

## Config Script

The script is run when seb is generating ninja files, and the output is parsed.
Any line containing a equal sign (`=`) is added as a ninja variable. This can
be used to e.g. run php-config only when seb is run instead of each call to the
compiler.

Any non-empty line output without an equal sign will be considered a condition
to activate, conditions are described in more detail further down.

## Compiler

You can switch the compiler used in CONFIG. The available choices are gcc or
clang:

	compiler[clang]

You will need respective compiler to be installed, obviously.
