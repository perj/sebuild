#!/usr/bin/perl
# Copyright 2018 Schibsted

use strict;

my $source = ($ARGV[0] eq '-s');
shift if $source;

my $output = shift;
close(STDOUT);
open(STDOUT, '>', $output) or die "Can't open $output for writing";

$output =~ /^.*?([^\/]+)\.([^\.]+)$/ or die "can't parse extension from $output";
my $name = $1;

my $prefix = uc ($name);
$prefix =~ s/([^_])[^_]*(_|$)/\1/g;
$prefix .= '_';

print "struct ${name}_rec;\n";
print "enum $name {\n\t\t${prefix}NONE";

my $decl_extra;
my @fields;
my @extra;
my $num = 0;
my $started = 0;
my $lnum = 0;
my $nocase;
while (<>) {
	chomp;
	$lnum++;

	my $field;
	my $extra;
	my $value = $lnum;

	if ($source) {
		if (!$started) {
			$started = 1 if /GPERF_ENUM(_NOCASE)?\($name(;(.*))?\)/;
			$nocase = "%ignore-case" if $1 ne '';
			$decl_extra = $3;
			$decl_extra .= ';' if $decl_extra !~ /;$/;
			next;
		}
		last if /GPERF_ENUM/;

		if (/GPERF_CASE\("(([^\\"]|\\.)*)"([^)]*)/) {
			$field = $1;
			$extra = $3;
		} elsif (/GPERF_CASE_VALUE\(([^,]+), "(([^\\"]|\\.)*)"([^)]*)/) {
			$value = $1;
			$field = $2;
			$extra = $4;
		} else {
			next;
		}
	} else {
		$field = $_;
		$extra = '';
	}

	push @fields, $field;
	push @extra, $extra;
	$field = uc($field);
	$field =~ tr/-/_/;
	print ", $prefix" . $field . " = $value";
	$num++;
}

print <<EOT
};
struct ${name}_rec {
	const char *name;
	enum ${name} val;
	$decl_extra
};

\%struct-type
\%define hash-function-name ${name}_hash
\%define lookup-function-name ${name}_lookup
\%readonly-tables\n
\%enum
\%compare-strncmp
\%define string-pool-name ${name}_strings
\%define word-array-name ${name}_words
$nocase

%{
EOT
;
print "#define MAX_${prefix}WORD $num\n" if (!$source);
print <<EOT

#ifndef register
#define register
#endif

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wmissing-declarations"
#if defined __clang__
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wstatic-in-inline"
#endif

%}
EOT
;

print "\%\%\n";

for my $i (0 .. $#fields) {
	my $v = uc($fields[$i]);
	$v =~ tr/-/_/;
	print $fields[$i] . ", $prefix" . $v . $extra[$i] . "\n";
}

print <<EOT
\%\%

static
enum $name lookup_$name (const char *str, int len) __attribute__((unused));

static
enum $name lookup_$name (const char *str, int len) {
	if (len == -1)
		len = strlen (str);
	const struct ${name}_rec *val = ${name}_lookup (str, len);

	if (val)
		return val->val;
	return ${prefix}NONE;
}

static
int lookup_${name}_int (const char *str, int len) __attribute__((unused));

static
int lookup_${name}_int (const char *str, int len) {
	return (int)lookup_$name (str, len);
}

#if defined __clang__
#pragma clang diagnostic pop
#endif
#pragma GCC diagnostic pop
EOT
;

if ($source) {
	print <<EOT

#ifndef GPERF_ENUM
#define GPERF_ENUM(x)
#define GPERF_ENUM_NOCASE(x)
#define GPERF_CASE(...) __LINE__
#define GPERF_CASE_VALUE(v, ...) (v)
#define GPERF_CASE_NONE 0
#endif
EOT
	;
}
;
