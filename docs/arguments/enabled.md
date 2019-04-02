## enabled

Can be used to disable the descriptor unless a certain flavor or condition
matches. Use with an empty value.

E.g. `enabled::foo[]` will enable this describor if the `foo` condition is set,
but it will otherwise be disabled. You can use enabled multiple times to
specify different conditions and flavors.
