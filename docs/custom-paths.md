## Custom library paths ##

If you have libraries in custom places, e.g. `/usr/pgsql-x.x`, the current best
way to find them is to create a `compile.sh` to setup the paths:

    export LIBRARY_PATH=/usr/pgsql-x.x/lib:$LIBRARY_PATH
    export CPATH=/usr/pgsql-x.x/include

Maybe even

    export PATH=/usr/pgsql-x.x/bin:$PATH

After you've set those variables run seb from the script.

Additional paths should be allowed to be specified in configvars files in the
future.
