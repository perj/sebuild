#!/bin/sh
# Copyright 2018 Schibsted

set -e

: ${BUILDPATH:=build}
export BUILDPATH
go install ./cmd/seb

if [ -n "$GOPATH" ]; then
	export PATH="$GOPATH/bin:$PATH"
elif ! command -v seb > /dev/null; then
	export PATH="$HOME/go/bin:$PATH"
fi

seb "$@"
[ -z "$RUNTESTS" ] && exit 0

ninja -f "$BUILDPATH"/build.ninja "$BUILDPATH"/dev/gotest/buildbuild

cd test
./compile.sh
