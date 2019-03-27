# Sub components - COMPONENT

The top level Builddesc contains something like this:

    COMPONENT(flavors[dev release] [
         bin/foo
         bin/bar
         regress
         scripts/db
    ])

`COMPONENT` is a special build descriptor that doesn't really follow the
general rules. It doesn't have a name and has one argument with no name and a
list of directories where we will find further Builddesc files describing other
components of our build. It is possible for one or more of those directories to
contain a Builddesc with `COMPONENT` in it, but be very careful with that. It's
much more transparent and readable to have a long list of everything built
instead of a magic tree where everything is hidden.

The only other argumnet for `COMPONENT` build descriptors is `flavors`. See
flavors documentation below for the semantics of that. If `flavors` is omitted,
the components will be included in all flavors.

