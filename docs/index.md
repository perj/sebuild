# Search Engineering Build

## Intro

This is a tool to generate a Ninja build solution and compile using ninja.

Ninja is a tool developed by Evan Martin at Google and is used to build Chrome
as fast as possible. As a Make replacement it's slightly quirky, it doesn't
really have much logic other than handling dependencies correctly, and you need
an external tool to generate ninja files for ninja to be usable. It's possible
to write them by hand, but it gets very chatty and unreadable fast.

The big advantage of ninja is that it forces you to think about what you're
doing and it is bloody fast.

## Index

### Concepts

* [The Basics](basics.md) describes how to setup and use the build system day
  to day.

* [High Level Concepts](high-level-concepts.md) is a mess, but describes
  dependencies, libraries and how the file tree is organized.

* [Writing Builddesc Files](writing-builddescs.md) and on describes how to
  actually write Builddesc files to build new components.

### Flavors and Conditions

* [Flavors](flavors.md) are different sets of output. For example you might
  wish to only build some programs in a development flavor, or use different
  compilation flags for dev and release.

* [Conditions](conditions.md) lets you select what to build on a global basis.
  This can be used to enable platform dependent code or an optional link to a
  particular system library.

### Descriptors

Each descriptor is described on its own page. The first two primarily exist
in the top level Builddesc, although COMPONENT can be used recursively.

* [Global Configuration - CONFIG](descriptors/config.md)
* [Sub components - COMPONENT](descriptors/component.md)

The below descriptors all generate some kind of output in the build/flavor
directory.

* [Normal Programs - PROG](descriptors/prog.md)
* [Go Program - GOPROG](descriptors/goprog.md)
* [Go Tests - GOTEST](descriptors/gotest.md)
* [Dynamic Modules - MODULE](descriptors/module.md)
* [Go Dynamic Modules - GOMODULE](descriptors/gomodule.md)
* [Scripts, Configuration and Other Files - INSTALL](descriptors/install.md)

The next set is for generating intermediate libraries and tool files in the
build/obj/flavor directory.

* [Libraries - LIB and LINKERSET_LIB](descriptors/lib.md)
* [Build Tool Programs - TOOL_PROG](descriptors/tool-prog.md)
* [Other Build Tools - TOOL_INSTALL](descriptors/tool-install.md)

### Descriptor Arguments

These arguments apply to all descriptors except CONFIG and COMPONENT.
If an argument only applies to one or some desciptors (again except CONFIG and
COMPONENT), that is noted specifically.

* [Including Builddesc fragments - INCLUDE](arguments/include.md)
* [Enabling Descriptors - enabled](arguments/enabled.md)
* [Limiting a Descriptor to Certain Flavors - flavors](arguments/flavors.md)
* [Specifying Sources - srcs](arguments/srcs.md)
* [Depending on Libraries - libs](arguments/libs.md)
* [Using Specialized Sources - specialsrcs](arguments/specialsrcs.md)
* [Setting Specific Options - srcopts](arguments/srcopts.md)
* [Finding Header Files - incdirs](arguments/incdirs.md)
* [Adding Extra Ninja Variables or Rules - extravars](arguments/extravars.md)
* [Adding Manual Dependencies - deps](arguments/deps.md)
* [Changing the Source Directory - srcdir](arguments/srcdir.md)
* [Changing the Destination Directory - destdir](arguments/destdir.md)
* [Collecting Targets in a Variable - collect_target_var](arguments/collect-target-var.md)
* [Linker Specific Arguments - cflags, cwarnflags, conlyflags, cxxflags, copts, no_analyse](arguments/linker-args.md)
* [Install Specific Arguments - conf, scripts, php, python, symlink](descriptors/install.md#arguments)
* [Library Specific Arguments - includes, libs, incprefix](descriptors/lib#arguments)

### Customizing Sebuild

* [Compiler and Linker Flags](compiler-flags.md) tells you how to modify the default
  compiler flags used by Sebuild.
* [Custom Library Paths](custom-paths.md) describe how to find libraries and
  tools in custom locations.
* [Custom Rules](custom-rules.md) describe how to add more rules that can be
  used in specialsrcs arguments.
* [Plugins](plugins.md) handles adding custom descriptors, arguments or
  specialsrcs.

### Advanced Topics

* [Separate Top-level Builddesc](separate-builddesc-top.md) tells you how to
  use sub projects that have their own CONFIG descriptor.
* [Special Files](special-files.md) contains some low-level details about
  files and rules that Sebuild uses.
* [Special Variables](special-variables.md) has some details about used Ninja
  variables.
* [Globbing](globbing.md) describes how globs work in srcs etc.
* [Static Analyser](static-analyser.md) tells you how to run the Clang static
  analyser on your code.
* [Target Name Conflict](target-name-conflict.md) describes an issue that might
  come up if you have sources and intermediate files with the same file name.

### Historical Notes

* [2.0 Upgrade Notes](upgrade-notes/2.0.md)
* [Tool Origins](origins.md)
