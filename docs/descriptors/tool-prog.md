# Build Tool Progams - TOOL_PROG

     TOOL_PROG(tool1
             srcs[tool1.c]
             specialsrcs[mkrules:rules.txt:rules.h]
             deps[tool1.c:rules.h]
             libs[platform_util]
     )

`TOOL_PROG` is just like `PROG`, with two differences. Tools get installed
under obj/flavor/tools/ and instead of like `PROG` - being part of the default
build target tools get built only if something depends on them.
