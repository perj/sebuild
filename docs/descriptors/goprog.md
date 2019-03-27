# Go Programs - GOPROG

    GOPROG(foo
    )

When used, the current directory will be compiled into a go binary.  You can
add the parameter nocgo[] to disable cgo for this program only. Other common
parameters also work, for example `libs[]` can be useful when compiling a Go
program linking with C libraries.

To build go programs, you should add a gopath variable in a configvars file.
This will be used to set the GOPATH environment variable while building, and
will override any such set by the environment.  If not set in configvars, it
will however default to the GOPATH environment variable.

Unlike the normal GOPATH, the gopath config variable can use paths relative to
the project root, they'll be changed into full paths by gobuild.sh and
invars.sh.
