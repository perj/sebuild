# Static Analyser

You can run the clang static analyser on your code. This is setup
automatically, but disabled by default since it's very slow. Run

    seb build/<flavor>/analyse

to generate a report in that directory. It will contain a very basic
index.html file you can open in your browser. If working remotely,
download the entire folder first.
