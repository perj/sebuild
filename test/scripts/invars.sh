#!/usr/bin/env bash

if [ -z "$buildtooldir" ]; then
	echo "buildtooldir environment variable unset in invars.sh." >&2
	exit 1
fi
source "$buildtooldir/scripts/invars.sh" "$@"
depend "$buildtooldir/scripts/invars.sh"

setval foovar fooval
