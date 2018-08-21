#!/usr/bin/env perl
# Copyright 2018 Schibsted

# This wrapper will parse rsp files for non-GNU ld versions.
# GNU ld supports them natively and we just exec it if we detect GNU.

use strict;

my $ld_v = `ld -v 2>&1`;
exec @ARGV if $ld_v =~ /GNU/;

my @args = ();

for (@ARGV) {
	/.rsp$/ and do {
		s/^@//;
		# I've successfully passed over 1G of command line on OS X, so this should work.
		open(my $rsp, '<', $_) or die "Failed to open $_";
		push @ARGV, split(/[[:space:]]+/, join(" ", <$rsp>));
		close($rsp);
		next;
	};

	push @args, $_;
}

exec @args;
