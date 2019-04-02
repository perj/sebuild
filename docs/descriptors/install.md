# Scripts, Configuration and Other Files - INSTALL

     INSTALL(regress/indexer/conf
             srcs[isearch.conf.in]
             conf[isearch.conf]
     )
    
     INSTALL(bin
             srcs[create_proschema.pgsql.in]
             scripts[create_proschema.pgsql]
     )
    
Installs files into the destination directory specified in the name. There are
different installation arguments described below.

A thing worth noting is that you can specify files that are copied straight
from the source directory and also configuration files that have been compiled
somehow (like .in). Sebuild will figure out which source file is built where
and where to copy it from.

## Arguments

### conf

    conf[some.conf]

Any kind of generic non-executable file, e.g. configuration files.

### scripts

    scripts[index_ctl]

Specifies executable shell scripts and such. Used to set the executable flag.

### php

    php[random.php]

PHP code to be installed and syntax checked using the PHP lint command.
Only valid in the INSTALL descriptor.

### python

    python[random.py]

Installs the python script but also compiles it, which produces .pyc and .pyo
files as well as checking the file syntax.

### symlink
Symlinks can also be created, using this syntax:

	INSTALL(/dir
		symlink[dst:src]
	)

The symlink `dst` is created in `build/flavor/dir` pointing to the `src`.
If `src` is not a full path it needs to be relative to the installation
directory, like normal for symlinks. There's no check that it points to any
valid file or directory. Symlinks are currently rebuilt based on the target
file, hopefully this will be fixed sometime in the future.
