# Overview

This is a [ninja](https://ninja-build.org/) generator. It generates a
ninja build solution from Builddesc files describing your project. It will
also run ninja to compile your project.

# Installation

To use sebuild, go get the seb binary. You will need to have a recent version
of go installed, we currently support Go 1.13 and higher.

	go get github.com/schibsted/sebuild/cmd/seb

The master branch of this project always correspond to the latest release,
it will correspond to a release tagged revision.

In case you obtain only the binary and discard the source, you have to use

	seb -install

to install the required runtime files. By default the source downloaded by
go get is however used. Seb will prompt you to do this if it can't find the
files.

If you wish to change the code or contribute, see
[CONTRIBUTING.md](CONTRIBUTING.md).

# Usage

You need a Builddesc file.  This file describes your buildables. A very simple
one would be to compile all C files in current directory into a program:

```
PROG(myprog
	srcs[*.c]
)
```

When you run seb with no arguments, this will produce the file
`build/dev/bin/myprog`.

For more details and instructions see the
[documentation site](https://schibsted.github.io/sebuild) and the
[seb man page](cmd/seb/seb.1.ronn.md).

## LICENSE

Copyright 2018 Schibsted

Licensed under the MIT License, you may not use this code except in compliance
with the License. The full license is included in the file LICENCE.

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
