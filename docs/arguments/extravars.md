## extravars

    extravars[morevars.ninja]

Extra ninja variables needed to build this descriptor. `extravars` will be
included only in this build descriptor and can therefore reference generated
paths. Since rules are global and `extranvars` can be included multiple times
it can't define rules. See special section below to see how to deal with this.
