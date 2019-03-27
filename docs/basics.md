# The Basics

You'll need to have ninja in your $PATH, you can get ninja from
[github](https://github.com/ninja-build/ninja).

If you want to install through packages, beware that there are other unrelated
software projects out there called 'ninja'.

If this document gets out of date, the tools we have are made roughly around
the time of ninja 1.0, and known to work up to ninja 1.8.

## How to run

The basic way to build a tree is simply to invoke `seb`.

    $ seb

This will compile everything you need.

The first thing seb does is to find the top directory of your project. To do
this it walks the directory tree upwards and looks for a Builddesc.top or
Builddesc file. If a Builddesc.top is found, that is used as the top level
directory, otherwise the topmost Builddesc is used. Since Builddescs are
somewhat recursive it was determined that using the topmost one makes most
sense.

After changing to this directory, seb will generate ninja files and then
execute ninja.

If you changed some file and want to recompile, just run `seb` again. It costs
50ms extra to run seb compared to just ninja directly, it's a lot less typing
and you'll never lose dependencies. Reading this sentence cost you more time
than you'd ever save skipping seb.

If you're not sure if you changed anything, you can still recompile and see
what happens. An empty compilation of the platform code is 100ms. Since seb
finds the top directory you should be able to run it from anywhere within the
tree.
