#!/usr/bin/env bash
# Copyright 2018 Schibsted

SRCFILE=$1
DSTFILE=$2
ORGANIZATION=${3:-Search Engineering}
MANUAL=${4:-Schibsted Search}
RONN=ronn

if command -v $RONN >/dev/null; then
	$RONN --pipe --roff --organization="$ORGANIZATION" --manual="$MANUAL" $SRCFILE > $DSTFILE
else
	echo $RONN not installed, missing manpage $SRCFILE >/dev/stderr
fi
