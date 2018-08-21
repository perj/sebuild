#!/bin/sh
# Copyright 2018 Schibsted

PYTHON="$(command -v python3 || command -v python)"

install -m0644 "$@" || exit 1

"$PYTHON" -m py_compile "$2" && "$PYTHON" -O -m py_compile "$2"
s=$?
test $s != 0 && rm -f "$2"
exit $s
