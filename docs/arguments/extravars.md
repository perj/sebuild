## extravars

    extravars[morevars.ninja]

Extra ninja variables needed to build this descriptor. `extravars` will be
included only in this build descriptor and can therefore reference generated
paths. You can put rules in these files as well, despite the `extravars` name,
although if you use the same file in multiple descriptors this might cause
a `duplicate rule definition` ninja error.
