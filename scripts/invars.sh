#!/bin/bash
# Copyright 2018 Schibsted

if [ $# -lt 1 ] ; then
	echo "Must specify buildvars as argument." >&2
	exit 1;
fi

set -e

for f in "$@"; do
	. $f
done

if [ "$(command -v setval)" != "setval" ]; then
	setval() {
		local _k=$1
		shift
		eval "$_k=\"$*\""
		echo "$_k=$*"
	}
fi
if [ "$(command -v depend)" != "depend" ]; then
	if [ -n "$depfile" ]; then
		depend() {
			for dep in $@; do
				echo -n " ${dep}" >> ${depfile}
			done
		}
	else
		depend() {
			:
		}
	fi
fi

setval BUILD_STAGE `echo $buildflavor | tr [:lower:] [:upper:]`

setval READLINK $(type -p greadlink readlink | head -1)

oldifs="$IFS"
IFS=":"
gp=""
for p in $gopath; do
	ABSP=$($READLINK -f $p || true)
	[ -z "$ABSP" ] && continue
	gp="$gp:$ABSP"
done
# Strip initial :
setval GOPATH "${gp:1}"
IFS="$oldifs"

setval GOARCH $(go env GOARCH 2>/dev/null)
setval GOOS $(go env GOOS 2>/dev/null)
