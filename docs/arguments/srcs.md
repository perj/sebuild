# srcs

    srcs[foo.c bar.cc]

Source files to build a program, library, module, tool or configuration file.
Seb automatically knows what to do with the following extensions:

* c - C source
* cc, cxx - C++ source
* go - Go source
* gperf.enum - enumerated gperf source
* in - to be processed by in.pl
* ll - C++ lex
* yy - C++ yacc

## Go Sources

Go sources are a bit special. When used, they will be compiled into a Go
c-archive that is then linked into the final binary.

To access your go code, use `import "C"` and then add `//export` to the
functions you wish to access from C. They will then be declared in `gosrc.h`
which can be included from your C source. This also works for any
packages you import, with the caveat that they're not added to `gosrc.h`
and must be declared manually in C.
The Go package(s) have access to all local headers for the C program.

Several restrictions apply to what functions can be exported, see
`go doc cmd/cgo`.

The Go runtime stops working when a process forks, and can't be reinitialized.
To workaround this in programs that fork at launch, the automatic loading of
the Go runtime at program initialization has been disabled. Instead you must
manually load it. To do that call the symbol names like this as a function
taking argc, argv and environ:

	echo _rt0_`go env GOARCH`_`go env GOOS`_lib

(this disabling the autostart of Go runtime should probably not be enabled by
default.)
