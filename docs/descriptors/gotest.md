# Go tests - GOTEST

    GOTEST(foo
    )

Uses `go test -c` to produce a binary called foo.test that's by default
installed to the /gotest destination directory. This binary can be executed to
run the tests, but do note that some go unit tests might assume that they're
run from a specific directory or have other expectations.

In addition, you can also run the go tests directly via ninja. The advantage of
using ninja here to run go test is that it adds include paths etc. for cgo
tests that can otherwise be difficult to figure out.

To allow to run go test via ninja, for each GOTEST descriptor a target
`build/flavor/gotest/name` is created which you can give to ninja to run the
test. In addition, you can also run `go test -cover` to generat an html file
with this target:

    build/<flavor>/gocover/<name>-coverage.html

And also `go test -bench` with

    build/<flavor>/gobench/<name>

You can override the default of running all benchmarks by adding a `benchflags`
argument to the `GOTEST` directive and putting a regexp there, possibly also
adding additional flags.

The gobuild tool used by GOTEST checks a number of ninja variables when
executed. These are described on the [GOPROG page](goprog.md).

As an advanced feature, GOTEST targets are also automatically collected in
variables, see the
[collect_target_var argument](../arguments/collect-target-var.md) for more more
information about this.
