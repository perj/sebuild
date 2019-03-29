#### Dynamic Modules - MODULE ####

     MODULE(php_module1
            srcs[php_module1.c]
            libs[platform_core]
            copts[-DCOMPILE_DL_MOD=1 $php_module_cflags]
     )
    
`MODULE` describes how to build various shared object modules. Those end up in
modules/. Php, perl and apache modules are all built the same way. All modules
are added to the default build target.

The MODULE descriptor is fairly similar to [PROG](prog.md) in that it creates
a binary in the destination directory. The main difference is the kind of
binary created.
