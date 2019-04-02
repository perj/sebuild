## Per Source Options - srcopts

You can add compile flags or other optiones on a per source bases.
Use the format `srcopts[src:flags]` in the Builddesc. For example, to
add `-Wno-error` while compiling buggy.c use `srcopts[buggy.c:-Wno-error]`.

This works for intermediate sources as well, if buggy.c was generated
(e.g. from buggy.ll), the flag is still applied when compiling the C file.

The value is added as the variable `$srcopts` to any ninja rule that
has the given source. Thus it's up to each rule to determine the semantics.
Currently it's setup for C/C++ compile rules to add srcopts as compilation
flags.
