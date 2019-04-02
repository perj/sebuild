# Build Tool Progams - TOOL_PROG

     TOOL_PROG(tool1
             srcs[tool1.c]
             specialsrcs[mkrules:rules.txt:rules.h]
             deps[tool1.c:rules.h]
             libs[platform_util]
     )

`TOOL_PROG` is just like [PROG](prog.md), with two differences. Tools get
installed under obj/flavor/tools/. Additionally they don't get compiled
by default but only if something depends on them, e.g. via
[ruledeps](config.md#ruledeps).
