## deps

    deps[config.h mkrules.c:sections.h]

Extra dependencies that Sebuild can't figure out by itself. Two forms:

* `<depender>:<dependency>`
  specifies that `<depender>` can't be built until `<dependency>`
  has been built.
* `<global dependency>`
  specifies a dependency for all source files in this descriptor.

This is often required to make sure a dependency exists before it's used.
