#!/usr/bin/perl -w
# Copyright 2018 Schibsted

use strict;

open INF, "-|", "bzcat ${ARGV[0]}" or die;

print "
struct section
{
	const char *name;
	int start;
	int end;
} sections[] = {
";

while (<INF>) {
	/^\@\@\t([0-9A-F]+)\t([^\t]+)\t([0-9A-F]+)$/ or next;
	$2 eq "Unassigned" and next;
	print "\t{\"$2\", 0x$1, 0x$3},\n";
}

print "\t{NULL}\n};\n";

close INF;

exit 0;
