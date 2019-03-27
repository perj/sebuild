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

