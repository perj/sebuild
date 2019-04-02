# Multiple Targets With the Same Name

Since Sebuild resolves dependencies and build rules according to the names of
generated files, special care needs to be taken when you end up with multiple
rules creating several different files with the same names in different
directories. In general this can only happen when you do something like this:

    INSTALL(conf
        srcs[conf.txt.in]
        conf[conf.txt]
        specialsrcs[magic:conf.txt:foobar]
    )
  
In the srcs argument `conf.txt.in` generates `conf.txt`, but the `conf`
argument does the same. Then the `specialsrcs` depends on `conf.txt` and it's
not really clear which one. The only safe heuristics we can apply to resolve
this is when one of the targets is an install target (`conf` or `scripts`). In
that case the dependency chain gets transformed from:

    $src/conf.txt.in -> $objdir/conf.txt -> $conf/conf.txt

to:

    $src/conf.txt.in -> $objdir/TMP_BUILDconf.txt -> $conf/conf.txt

So if you look in the object directory for the generated file, it might have a
different name.
