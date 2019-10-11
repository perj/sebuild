seb(1) -- build your project
============================

## SYNOPSIS

`seb` [options] [--] [ninja-arguments]
`seb` `-tool` tool [tool-options]

## DESCRIPTION

The seb tool generates build.ninja files from special Builddesc files that
describe how various components of the build should be handled.

This is a very brief description. For information on how to write Builddesc
files refer to the documentation site, https://schibsted.github.io/sebuild

## OPTIONS

`--with-flavor` string

  Only generate the specified flavors, multiple --with-flavor options are
  allowed to select many flavors. Usually not needed as each flavor is also a
  ninja pseudo-target, thus

    seb flavor

  should work just as well.

`--without-flavor` string

  Exclude the specified flavors from the build.

`--quiet`

  Suppress program output with `--quiet`.

`--debug`

  Enable debug output with `--debug`.

`--condition` string

  Add an active build condition, which can be used to select what files
  or flags are active.

`-tool` tool

  Invoke an internal tool. These are usually invoked by ninja and are not
  considered stable.
  The current available tools are `asset`, `copy-analyse`, `go-install`,
  `gperf-enum`, `header-install`, `in`, `invars`, `link`, `python-install`,
  `ronn` and `touch`.

  This flag can only be used as the first argument given to `seb`. The rest of
  the arguments are passed to the tool rather than parsed as seb or ninja
  flags.

## SEE ALSO

There should be an extensive set of markdown documentation files bundled with
this tool. They're also available at https://schibsted.github.io/sebuild
