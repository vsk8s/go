#!/usr/bin/perl

# Extract list of golang architectures by looking at control file.

package main;

use warnings;
use strict;

use Dpkg::Control::Info;

my $control = Dpkg::Control::Info->new();
my $p = $control->get_pkg_by_name("golang-go");
printf("%s\n", $p->{Architecture});
