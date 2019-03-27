# specialsrcs

    specialsrcs[charmap:swedish.txt:charmap-swedish
                charmap:russian.txt:charmap-russian
                charmap:swedish-utf8.txt:charmap-swedish-utf8
    ]           
    
This argument specifies a method to compile a source file where the automatic
file type detection can't figure out how to compile. Each element in is of the
form: `<command>:<src>[,<src>]*:<target>[:<extravars>[,<extravars>]*]`

`command` is the ninja rule to build the target, `src` is a comma separated
list of sources, `target` is the resulting output. `extravars` is a comma
separated list of extra variables added to the build directive for ninja.

Notice that specialsrcs is mostly useless on its own, the script will not pick
up and do anything with the output from the script. You will need to specify
additional arguments that use the output somehow (usually `conf` or `srcs`).

## Special commands

Plugins are allowed to hijack specialsrcs by redirecting any specific commands
to themselves. Such specialsrcs should be documented in the plugin
documentation. Sebase does not currently have any builtin special rules here.

## Common Commands

The rules.ninja file bundles with sebuild contains some rules that are meant to
be used as commands in specialsrcs. You can further add your own either locally
via the [extravars argument](extravars.md) or in the [CONFIG
descriptor](../descriptors/config.md) via the `rules` argument there.

### concat

Simply concatenates the sources into the target, can also be used to copy files.

### gperf_switch

The source should be a C or C++ file. Inside these you can have
`GPERF_ENUM(x)` or `GPERF_ENUM_NOCASE(x)` markers to start a gperf switch.
Below you add `switch (lookup_x(str, strlen(str)))` and inside you can use
`case GPERF_CASE("foo"):` to match str vs. "foo" via gperf.
The target header file must have the name `x.h` to match the arguemnt given
to ´GPREF_ENUM`. This header files contains definitions for the required macros
and also the inline lookup function.

### symlink

Used to create a symlink. Doesn't have a source, instead tell it the path
for the symlink target via a `target` extravar.
Typically you set a custom directory for the target of this specialsrc,
see the [destdir argument](destdir.md) for valid prefixes.

### protocc, protoc-c

Used to run protobuf generating tools on a `.proto` file. The target file
name should match what's created by protoc.

### download

The source file should have one line with a URL and one line with the
sha256 sum of the file the URL specifies. It will be downloaded and checked.

### unzip, untar

Can be used to extract files from an archive, e.g. one downloaded via the
download command. Set the `file` extravar to the inside archive path to
be extracted.

### bunzip2

Runs bunzip2 to decompress a file.

### ronn

Runs the ronn command to generate man pages as used to generate the
seb man page.
