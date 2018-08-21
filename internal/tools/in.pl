#!/usr/bin/perl
# Copyright 2018 Schibsted

use strict;

my %sub = ();
my $conffile = shift;
if ($conffile) {
	open my $conf, "<", $conffile or die "open $conffile: $!";
	while (<$conf> =~ /([a-z_A-Z0-9]+)=(.*)/) {
		$sub{$1} = $2;
	}
	close $conf;
}

my $infile;
my $outfile;

my $oldout = "";
my $newout = "";

my $IN;

if ($infile = shift) {
	open $IN, "<", $infile or die "open $infile: $!";
	$outfile = shift or die $!;
	$oldout = do { local(@ARGV, $/) = $outfile; <> } if -e $outfile;
} else {
	$IN = *STDIN;
}

my $end = '';
while (<$IN>) {
	s/%$sub{'BUILD_STAGE'}_START%\n//;
	s/%ANY_START%\n//;

	/%([A-Z]+)_START%\n/ and $end = $1;
	s/%${end}_END%\n// and $end = '';
	next if $end;

	s/%[A-Z]+_END%\n//;

	my $res = '';
	# For every segment that doesn't contain a '%' ...
	while (m/\G([^%]*)/msgc) {
		$res .= $1;
		# If what follows next is a '%' ID '%' sequence, then output the substitution for the ID.
		if (m/\G%([a-zA-Z][a-zA-Z0-9_]*?)%/gc && defined $1) {
			die "undefined variable >$1< -> $_" if (!defined $sub{$1});
			$res .= $sub{$1};
		} elsif (m/\G(%+)/gc) {
			# If we see one or more '%' at this point, just add them to the output.
			$res .= $1;
		}
	}

	# Handle the 'double-escape' which is sometimes used in the wild.
	$res =~ s/%%/%/g;

	$newout .= $res;
}

if ($outfile) {
	if ($oldout eq "" || !($oldout eq $newout)) {
		open my $OUT, ">", $outfile;
		print $OUT $newout;
		close $OUT;
	}
} else {
	print $newout;
}
