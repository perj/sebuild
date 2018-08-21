#!/bin/bash
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

READLINK="$(type -p greadlink readlink | head -1)"
ABSIN="$($READLINK -f "${IN:-.}")"

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
	ABSP=$($READLINK -f $p)
	[ -z "$ABSP" ] && continue
	[ -z "$PKG" ] && [ "${ABSIN#$ABSP/src/}" != "$ABSIN" ] && PKG="${ABSIN#$ABSP/src/}"
	gopath="$gopath:$ABSP"
done
IFS="$orig_IFS"
# Strip initial :
GOPATH="${gopath:1}"

if [ -z "$mode" ]; then
	mode="prog"
fi

BUILDFLAGS="-i -pkgdir $($READLINK -f $pkgdir)/gopkg_$mode -installsuffix=$mode"

if [ "$mode" = "test" ]; then
	[ -z "$PKG" ] && cd "$ABSIN"
	exec go test $GOBUILD_TEST_FLAGS $PKG
fi
if [ "$mode" = "bench" ]; then
	[ -z "$PKG" ] && cd "$ABSIN"
	exec go test $PKG -bench $4
fi
if [ "$mode" = "cover_html" ]; then
	exec go tool cover -html=$IN -o $OUT
fi

out="$($READLINK -f $OUT)"

[ -z "$PKG" ] && pushd "$ABSIN" > /dev/null

if [ "$mode" = "cover" ]; then
	go test -coverprofile=$out $PKG
	[ -z "$PKG" ] && popd > /dev/null
else
	if [ "$mode" = prog-nocgo ]; then
		CGO_ENABLED=0
		export CGO_ENABLED
		go build $BUILDFLAGS -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
		[ -z "$PKG" ] && popd > /dev/null
	elif [ -z "$mode" ] || [ "$mode" = prog ]; then
		go build $BUILDFLAGS -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
		[ -z "$PKG" ] && popd > /dev/null
	else
		# go build links an executable to extract the symbols. If this is a plugin there'll be
		# unresolved symbols. Ignore now, handle in final link.
		CGO_LDFLAGS="-Wl,--unresolved-symbols=ignore-in-object-files $CGO_LDFLAGS"
		if [ "$mode" = "piclib" ]; then
			# -a to build standard libs with -shared
			go build $BUILDFLAGS -buildmode=c-archive -gcflags='-shared' -asmflags='-shared' -a -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
		else
			go build $BUILDFLAGS -buildmode=c-archive -o "$out" "${EXTLDFLAGS[@]}" $PKG || exit 1
		fi
		[ -z "$PKG" ] && popd > /dev/null
		# If there weren't any exports the header won't be created, but we expected it to be there.
		touch "$($READLINK -f ${OUT%.a}.h)"
	fi
fi

(cd $ABSIN ; echo -n "$OUT: " ; go list -f "${PKG:-.}"' {{range .Deps}}{{.}} {{end}}' $PKG | xargs go list -f '{{$dir:=.Dir}}{{range .GoFiles}}{{$dir}}/{{.}}{{"\n"}}{{end}}{{range .CgoFiles}}{{$dir}}/{{.}}{{"\n"}}{{end}}{{range .HFiles}}{{$dir}}/{{.}}{{"\n"}}{{end}}{{range .CFiles}}{{$dir}}/{{.}}{{"\n"}}{{end}}{{range .TestGoFiles}}{{$dir}}/{{.}}{{"\n"}}{{end}}') | tr "\n" " " > $DEPFILE || exit 1
