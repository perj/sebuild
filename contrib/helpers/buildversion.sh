#!/bin/sh
# Copyright 2018 Schibsted

export LANG="C"
revno=`svn info $1 2>/dev/null | grep "Last Changed Rev:" | cut -f4 -d' '`
if [ -z "$revno" ]; then
	if [ -f .buildversion.cache ]; then
		read cached_sha1 cached_revno < .buildversion.cache
	fi
	sha1=`git log --format=%H -1 2>/dev/null`
	if [ "$cached_sha1" = "$sha1" ]; then
		revno=$cached_revno
	else
		revno=`git rev-list --count HEAD 2>/dev/null`
		if [ -z "$revno" ]; then
			# Fallback for git < 1.8
			revno=`git rev-list HEAD | wc -l 2>/dev/null`
		fi
		[ -n "$revno" -a "$revno" != 0 ] && echo $sha1 $revno > .buildversion.cache
	fi
fi
if [ -z "$revno" ]; then
	exit 1
else
	echo $revno
fi
