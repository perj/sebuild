#!/usr/bin/env perl
# Copyright 2018 Schibsted

use strict;
use Getopt::Long;
use File::Path;
use File::Find;
use Cwd;
use POSIX qw(strftime);

my $finalize = 0;

GetOptions("finalize" => \$finalize);

my $output = Cwd::abs_path(shift);

File::Path::remove_tree($output, { keep_root => 1 });
mkdir($output);

sub generate_index($) {
	my $srcdir = $_[0];

	my %index = ();

	File::Find::find({wanted => sub {
		next if !/\.html$/;

		my $src = $_;

		open SRC, '<', $src or die "Failed to open $src";
		my @srcdata = <SRC>;
		close SRC;

		my @titles = map /^<title>(.*)<\/title>$/, @srcdata;
		my %meta = map /^<!-- BUG([^ ]+) (.*) -->$/, @srcdata;
		if (@titles) {
			$meta{'src'} = $src;
			while ( (my $k, my $v) = each %meta) {
				$v =~ s/</&lt;/g;
				$v =~ s/>/&gt;/g;
				$meta{$k} = $v;
			}
			push @{$index{$titles[0]}}, \%meta;
		}

	}}, $srcdir);

	open DST, '>', "$srcdir/index.html" or die "Failed to open $output/index.html for output";

	my $now = strftime "%a %b %e %Y at %H:%M:%S", gmtime;
	my $reports = 0;

	print DST "<html><head><title>clang analyze report, $now</title></head>\n<body>\n";
	print DST "<p>Report generated $now</p>\n";
	print DST "<dl>\n";
	foreach my $k ( sort keys %index ) {
		my $res = "<dt>$k</dt>\n<dd><ol>\n";
		my @metas = sort { $a->{'LINE'} <=> $b->{'LINE'} } @{$index{$k}};
		foreach my $meta (@metas) {
			my $link .=
				'<li><span title="'.$meta->{'CATEGORY'}.': '.$meta->{'TYPE'}.'">' .
				'<a href="'.$meta->{'src'}.'#EndPath">line '.$meta->{'LINE'}."</a>: ".$meta->{'DESC'}."</span></li>\n";
			$res .= $link;
			++$reports;
		}
		$res .= "</ol></dd>\n";
		print DST $res;
	}

	if ($reports == 0) {
		print DST "<dt>No errors reported.</dt>\n";
	}

	print DST "</dl>\n";
	print DST "</body>\n</html>\n";

	close(DST);

	return $reports;
}

File::Find::find({wanted => sub {
		next if !/\.html$/;

		my $src = $_;
		my $dst = $src;

		open SRC, '<', $src or die "Failed to open $src";
		my @srcdata = <SRC>;
		close SRC;

		my $false_positive = 0;
		/^report-/ && do {
			my $title;
			my $buglnum;
			my $line = 0;
			for (@srcdata) {
				last if $false_positive;
				$line++;
				$title = $1 if /^<title>(.*)<\/title>$/;
				$buglnum = $1 if /^<!-- BUGLINE (.*) -->$/;
				$false_positive = 1 if $buglnum && /id="LN$buglnum".*TAILQ_REMOVE/;
				$false_positive = 1 if $buglnum && /id="LN$buglnum".*RB_GENERATE_STATIC/;
				$false_positive = 1 if $buglnum && /id="LN$buglnum".*ALLOC_PERMANENT_ZVAL/;
			}
			if ($title) {
				$title =~ s,/,__,g;
				$dst =~ s/^report/$title/;
			}
		};

		next if $false_positive;

		open DST, '>', "$output/$dst" or die "Failed to open $output/$dst";
		print DST @srcdata;
		close(DST);

	}
}, @ARGV) if @ARGV;

if ($finalize) {
	my $reports = generate_index($output);
	if ($reports > 0) {
		print STDERR "\e[31;1mTotal of " . $reports . " reports\e[0m\n";
		print STDERR "mongoose -r $output\n";
		print STDERR "to launch web server, or browse to file://localhost/$output/index.html\n";
		exit(1);
	}
}

