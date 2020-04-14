# Conditions

In addition to flavors, build conditions can also be used to select what's
built.  A condition is a simple truth check -- the presence or absence of a
condition is checked. If all checks are true the rule is added (merged) into
the build result. If any condition is false the rule is skipped.  Possible
example could look like this:

	LIB(platform_random
		srcs::linux,x86_64[*.c]
		includes[rand.h]
	)

Sources are only added if both conditions `linux` and `x86_64` are set. Three
conditions are always added: lowercase output of `uname`, output of `uname
-m` and the compiler flavor. For example `linux`, `x86_64` and `gcc`.

Condition checks can be negated by prefixing them with a bang (`!`). Then
the check is successful if the condition is NOT set.

Conditions can also be added in `CONFIG` like so:

	CONFIG(
		conditions[a b]
	)

adds conditions a and b. They can also be passed on the command line, running
`seb -cond foo` adds the condition `foo`.

Finally they can also be added by the [config script](descriptors/config.md#config_script).
This is the main way to add conditions dynamically, based on the configuration
script output.

Conditions can use letters, numbers and underscore (`_`), no other characters
are allowed.

## nocgo condition

A special condition is `nocgo`. If used, the environment variable `CGO_ENABLED`
will be set to 0. The check is done in the other direction as well, using either
sets both. This is to sync Sebuild ninja file generation with go build
invocation.
