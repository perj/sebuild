# Go Dynamic Modules - GOMODULE

    GOMODULE(foo
    )

Similar to [GOPROG](goprog.md) this uses go build to compile the directory into
a binary. Unlike goprog, this descriptor creates a dynamic loadable module
instead of a runnable progam. By default the module is created in the
`modules/` [destination directory](../arguments/destdir.md) and will have a
`.so` file suffix.

In the Go echosystem, these are known as plugins and you use the standard
[plugin](https://godoc.org/plugin) package to load them.

GOMODULE support most the same arguments as GOPROG does, so refer to
[GOPROG arguments](goprog.md#arguments) for descriptions of additional
arguments. Special conditions and ninja variables are described on that
page as well. Since Go doesn't support compiling these modules without
cgo the `nocgo` argument is not available. Other settings disabling cgo
are also ignored.
