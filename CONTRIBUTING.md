# Contributing to sebuild

You're welcome to contribute to this project.

## Copyright

By default, we require contributors to accept this extra condition. It's cut
from the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0), but
as this project is released with the MIT license, this is the license referred
to:

Unless You explicitly state otherwise, any Contribution intentionally submitted
for inclusion in the Work by You to the Licensor shall be under the terms and
conditions of the License, without any additional terms or conditions.
Notwithstanding the above, nothing herein shall supersede or modify the terms
of any separate license agreement you may have executed with Licensor regarding
such Contributions.

## Process

We prefer that you open a Github issue before starting to work on a
contribution.  This allows us to comment and agree or disagree before any work
is committed.  Make sure to look through the existing issues to make sure it
has not already been reported.

Once you have a patch you feel should be included in the main project, open
a Pull Request. We will review it on a best effort basis.

## Building sebuild for development or packaging

To build the seb binary from the root of the source directory, issue:

```bash
$ ./compile.sh
```

This will first do `go install ./cmd/seb` and then use that binary to
build seb again to `build/dev/bin/seb`.

To also run the included tests, use `RUNTESTS=1 ./compile.sh`.
