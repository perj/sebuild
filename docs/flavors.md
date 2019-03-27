# Flavors

The build system supports building various flavors of the same code. The
typical configuration will build 'dev' and 'release' versions of the code.

There are three ways the flavors can be differentiated between each other.

## In-file flavors

The [srcs arguments](arguments/srcs.md) allow you to generate files by
simply replacing some variables. These variables or entire portions of the
in file can be flavor specific.

Sebuild runs a script to generate the variables for the in-files. This
script include the `build/obj/<flavor>/buildvars.ninja` file which
contain various flavor specific settings.

### COMPONENT flavors ###

`COMPONENT` descriptors can contain the argument `flavors` which contains the
list of flavors this component will only be built for. The default is to build
for all flavors.

If you want some components to be only built for dev, you can do:

    COMPONENT(flavors[dev]
     [
   	regress/component1
   	regress/component2
     ]
    )

### flavor modifiers on arguments ###

Almost all arguments can be modfied with the flavor with a colon and the flavor
name after. Like this:

    srcs[a.c b.c]
    srcs:dev[c.c]

The result will be a merge between the unflavored arguments and the arguments
for this flavor (`srcs[a.c b.c c.c]` in the example).
