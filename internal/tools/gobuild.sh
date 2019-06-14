#!/usr/bin/env bash
# Copyright 2018 Schibsted

# Brute force wrapper script to integrate building go code into our
# build system. The go tools are amazing for building go code, but
# they are a bit provincial. There is a bigger world out there and
# sometimes you need to cooperate with others, you can't just build
# your own perfect utopia and pretend that the world outside doesn't
# exist.

# We expect the GOPATH variable to be set pointing to all go
# code. This is done in config.ninja. We will build the package main in
# directory $2 to the binary in $3, while writing out dependencies
# into the file $4. The dependencies will be as correct as we can make
# them so that we don't have to run go build again unless we have to
# because even with a simple hello world program it's 20x slower than
# running our build system.

PKG=$1
IN=$2
OUT=$3
DEPFILE=$4
cflags="$5"
ldflags="$6"
mode="$7"
pkgdir="$8"


ABSIN="$(cd "${IN:-.}" 2>/dev/null ; pwd)"

# Convert relative paths to absolute, since go will change directory.
CGO_CFLAGS=""
for f in $cflags; do
	[[ "$f" =~ ^[-/] ]] || f="$PWD/$f"
	CGO_CFLAGS="$CGO_CFLAGS $f"
done
export CGO_CFLAGS
CGO_LDFLAGS=""
objs=""
for f in $ldflags; do
	[[ "$f" =~ ^[-/] ]] || f="$PWD/$f"
	if [[ "$f" =~ \.o$ ]]; then
		objs="$objs $f"
	else
		CGO_LDFLAGS="$CGO_LDFLAGS $f"
	fi
done
export CGO_LDFLAGS

if [ "$objs" != "" ]; then
	# If we have object files, we want to only pass them to the final link,
	# otherwise they might be added twice. Unfortunately, they might reference
	# new library objects, so we have to pass all the libraries as well.
	EXTLDFLAGS=(-ldflags "-extldflags \"$objs $CGO_LDFLAGS\"")
else
	EXTLDFLAGS=()
fi

# Convert GOPATH to absolute, since go demands it. Also figure out pkg name if unset.
orig_IFS="$IFS"
IFS=":"
gopath=""
for p in $GOPATH; do
	ABSP="$(cd "$p" 2>/dev/null && pwd)"
	[ -z "$ABSP" ] && continue
	gopath="$gopath:$ABSP"
done
IFS="$orig_IFS"
# Strip initial :
GOPATH="${gopath:1}"

# If we have an explicit GOPATH then disable go modules.
# Probably will want to drop support for GOPATH quite soon, but that should be
# a major release so it will have to wait for that.
if [ -n "$GOPATH" ]; then
	export GO111MODULE=off
fi

if [ -z "$mode" ]; then
	mode="prog"
fi

if [ "$mode" = "test" ]; then
	[ -z "$PKG" ] && cd "$ABSIN"
	exec go test $GOBUILD_FLAGS $GOBUILD_TEST_FLAGS $PKG
fi
if [ "$mode" = "bench" ]; then
	[ -z "$PKG" ] && cd "$ABSIN"
	exec go test $GOBUILD_FLAGS $GOBUILD_TEST_FLAGS $PKG -bench $4
fi
if [ "$mode" = "cover_html" ]; then
	exec go tool cover -html=$IN -o "$OUT"
fi

# XXX Is there a better way? Except for GNU readlink which we can't assume.
out="$(cd "$(dirname "$OUT")" ; pwd)/$(basename "$OUT")"
depfile="$(cd "$(dirname "$DEPFILE")" ; pwd)/$(basename "$DEPFILE")"

if [ -z "$PKG" ] && [ -n "$(cd "$ABSIN" 2>/dev/null && go env GOMOD 2>/dev/null)" ]; then
	PKG="./$IN"
fi
[ -z "$PKG" ] && cd "$ABSIN" > /dev/null

# Do the deps file async to speed it up slightly.
# It's waited for at the end as long as the compile worked.
( echo -n "$OUT: " ; go list $GOBUILD_FLAGS -deps -f '{{$dir:=.Dir}}{{range .GoFiles}}{{$dir}}/{{.}} {{end}}{{range .CgoFiles}}{{$dir}}/{{.}} {{end}}{{range .HFiles}}{{$dir}}/{{.}} {{end}}{{range .CFiles}}{{$dir}}/{{.}} {{end}}{{range .TestGoFiles}}{{$dir}}/{{.}} {{end}}' $PKG ) > "$depfile" &

case "$mode" in
	cover)
		go test $GOBUILD_FLAGS -coverprofile="$out" $GOBUILD_TEST_FLAGS $PKG
	;;
	prog-nocgo)
		CGO_ENABLED=0
		export CGO_ENABLED
		go build $GOBUILD_FLAGS -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
	;;
	test-prog)
		go test -c $GOBUILD_FLAGS -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
	;;
	""|prog)
		go build $GOBUILD_FLAGS -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
	;;
	*)
		# go build links an executable to extract the symbols. If this is a plugin there'll be
		# unresolved symbols. Ignore now, handle in final link.
		if [ "$(go env GOOS)" != darwin ]; then
			CGO_LDFLAGS="-Wl,--unresolved-symbols=ignore-in-object-files $CGO_LDFLAGS"
		fi
		BUILDFLAGS="-i -pkgdir $(cd "$pkgdir" ; pwd)/gopkg_$mode -installsuffix=$mode $GOBUILD_FLAGS"

		if [ "$mode" = "piclib" ]; then
			# -a to build standard libs with -shared
			go build $BUILDFLAGS -buildmode=c-archive -gcflags='-shared' -asmflags='-shared' -a -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
		else
			go build $BUILDFLAGS -buildmode=c-archive -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
		fi
		# If there weren't any exports the header won't be created, but we expected it to be there.
		touch "${OUT%.a}.h"

		if [ "$(go env GOOS)" != darwin ]; then
			# Try to disable auto-start of go runtime. We want to be able to fork.
			# Don't know how to do it on Darwin right now.
			objcopy --rename-section .init_array=go_init --globalize-symbol="_rt0_$(go env GOARCH)_$(go env GOOS)_lib" "$out"
		fi
	;;
esac

# Wait for depfile generator.
wait
