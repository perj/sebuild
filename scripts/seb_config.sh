#!/bin/sh

if command -v ronn >/dev/null; then
	echo "ronn"
else
	echo "No ronn binary. Man pages will not be built." >&2
fi
