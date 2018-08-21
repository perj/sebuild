#!/bin/sh
# Copyright 2018 Schibsted

IN="$1"
OUT="$2"

rm -f "$OUT"
echo "# line 1 \"$IN\"" > "$OUT"
exec cat "$IN" >> "$OUT"
