# Writing Your Own Rules

One behavior of ninja that is important to understand is that rules are global
while variables can be local (read up ninja documentation to understand
variable scoping). This means that all rules files can only be included once.
If you compile a tool that will then be used inside a rule you must therefore
do something like this:

 * `Builddesc`:

          TOOL_PROG(mytool
              srcs[...]
          )
          INSTALL(conf
              conf[fooconf]
              specialsrcs[myrule:fooconf:in.foo]
              extravars[myvars.ninja]
          )

 * `<global>/rules.ninja`:

          rule myrule
              command=$mytool $in $out

 * `myvars.ninja`:

          mytool=$buildtools/mytool

 * `<topdir>/Builddesc`
          CONFIG(... ruledeps[... myrule:$mytool] ...)

This uses the fact that rules have to defined globally, but the variables that
the rules are using are scoped. So we can get away with having a local scope on
the definition of the tool we use.

Notice the special top level CONFIG addition. This tells ninja that all targets
built with the new rule should depend on the tool. If the tool changes,
everything built with the tool should be rebuilt. If you depend on a tool and a
configuration file you can specify multiple dependencies for the rule separated
by a comma, like this `ruledeps[...  myrule:$mytool,$myconfigfile ...]`.
