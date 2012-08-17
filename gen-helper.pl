#!/usr/bin/perl
#
# This script embeds the helper App Engine app's source into the testing package.
#

use strict;
use FindBin qw($Bin);
open(my $fh, "$Bin/app/helper/helper.go") or die "opening app/helper/helper.go: $!";
my $slurp = do { local $/; <$fh> };
die "helper.go contains backticks and Brad's lazy.\n" if $slurp =~ /`/;
open(my $out, ">$Bin/gen-helpersource.go") or die "opening gen-helpersource.go: $!";
print $out "package appenginetesting;\nvar helperSource = `$slurp`;\n";
close($out) or die;
