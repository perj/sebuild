#!/bin/sh
# Copyright 2018 Schibsted

IN="$1"
OUT="$2"

if [ -n "`gofmt -l $IN`" ]; then
	gofmt $IN
	exit 1
fi
rm -f "$OUT"
echo "//line $PWD/$IN:1" > "$OUT"
exec cat "$IN" >> "$OUT"
