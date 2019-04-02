# Origins

This software was originally developed by members of the Schibsted Coordination
Team to build what was known as the Blocket Platform, which we maintained
within the organization from 2010 to 2017. Ninja had just reached version 1.0,
and we wanted something that could build our tree faster than our existing GNU
Make solution.

At that point in time, tools like Bazel and Tundra weren't publicly available,
so the options were slim.

Initially a perl-script was produced. This was eventually rewritten into Go.

This documentation was originally written for internal use, and while it's been
edited for the public release, it will no doubt still reflect those origins.

## Original Authors

The original design and implementation was done by
[Artur Grabowski](https://github.com/art4711).

Prior to the public release the Go version was primarily written by
[Per Johansson](https://github.com/perj) with contributions from
[Eddy Jansson](https://github.com/eloj),
Artur Grabowski,
[Daniel Gustafsson](https://github.com/danielgustafsson),
[Claudio Jeker](https://github.com/cjeker),
[Fredrik Kuivinen](https://github.com/frekui),
[Daniel Melani](https://github.com/dmelani) and
[Sergio Visinoni](https://github.com/piffio).
