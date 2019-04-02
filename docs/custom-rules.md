# Writing Your Own Rules

Rules are what ninja use to convert input files into output, they are basically
shell commands. An example rule might look like this:

    rule gperf
        command = gperf -L ANSI-C --output-file=$out $in
        description = gperf $out

When the gperf rule is used the gperf command is run to create `$out` from `$in`.
For more details about this see the
[Ninja documentation](https://ninja-build.org/manual.html).

## Where to Put Your Rules

Ninja rules can only be defined once. This means you have to be a bit careful
about where you define them. Easiest is to use the
[rules config argument](descriptors/config.md#rules) to name a file that
contain all your custom rules. There is however no restriction on using any
of the other arguments naming ninja files, such as
[extravars](arguments/extravars.md), if you're sure it will only be included
once.

## Remove the Output on Failures

You have to be careful that if your rule command fails, the output file must
be manually removed. Many standard commands does this automatically, but
several will not. Thus your command might need to do this manually, like this:

    command = cat > $out < $in || ( rm -f "$out" ; exit 1 )

## Using Rules

Custom rules are primarily used via the
[specialsrcs argument](arguments/specialsrcs.md). With it you name the rule to
use as well as the input and output files. You can also add additional local
ninja variables if you wish, which can be used in the rule.

## Dependencies

If the command your custom rule use depends on files in the local repository
you should setup a dependency on those files for when your custom rule is used.
That's done using the
[ruledeps config argument](descriptors/config.md#ruledeps). For example when
you have the rule

    rule myrule
        command = myscript.sh $out $in

In the CONFIG descriptor your should add

    ruledeps[
            myrule:myscript.sh
    ]

It might be a good practice to use a Ninja variable to make sure they're kept
in sync. You can set the variable in the rules file and used it in ruledeps:

    myrule_tool=myscript.sh
    rule myrule
        command = $myrule_tool $out $in

    ...

    ruledeps[
            myrule:$myrule_tool
    ]

