# Installation arguments

These arguments are valid in INSTALL and TOOL_INSTALL only. They mostly
copy files into specific places in the build directory. The source
files might be either from the source tree or generated intermediate files.

## conf

    conf[some.conf]

Configuration files to be installed.  Only valid in `INSTALL`.

## scripts

    scripts[index_ctl]

Scripts to be installed.  Only valid in `INSTALL`

## php

    php[random.php]

PHP code to be installed and syntax checked using the PHP lint command.
Only valid in the INSTALL descriptor.
