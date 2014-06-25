#!/usr/bin/env perl
#
# !! ONLY FOR INTERNAL TESTING - IT'S VERY, VERY HACKY !!
# !! IT SENDS AN HACKY DHCP-REQUEST BROADCAST DATAGRAM WITH 'Requested IP Address': 01.01.01.01' !!
#
#
# * executed from jenkins
# * uses sudo for install / service management (hint: jenkins sudo - disable 'requiretty' in sudo config)
#
#
# Usage:
#   runner.pl -package-dir <PATH>
#     tests the latest package (with the highest version) under <PATH>
#
#   runner.pl -jenkins-url http://<IP>:<PORT>/job/<JOB>
#     tests the artifacts from the last succesful build
# 
#
#
# CPAN Modules:
#   * YAML::XS: config file parsing
#   * autodie: autodie on error
#   * Test::Most: test infra
#
#
use v5.14;
use strict;
use warnings;
use diagnostics;
use Config;
use local::lib;
use autodie;
use YAML::XS qw/LoadFile Load/;
use Test::Most;
use Socket;
use File::Basename;
use File::Temp qw/tempdir/;
use Getopt::Long;
use LWP::Simple;


# args
my $package_dir_flag; # local directory with packages (--package-dir-flag <BASE_DIR>/build-output/)
my $jenkins_url_flag; # jenkins url (--jenkins-url http://<IP>:<PORT>/job/<JOB>)
GetOptions("package-dir:s" => \$package_dir_flag,
           "jenkins-url:s" => \$jenkins_url_flag);


my $last_package_path;
if(defined($package_dir_flag)){
    $last_package_path = find_last_package_from_dir($package_dir_flag);
}elsif(defined($jenkins_url_flag)){
    $last_package_path = find_package_from_jenkins($jenkins_url_flag);
}else{
    say "parameter  -package-dir or -jenkins-url missing";
    exit(1);
}




# if any test fail - STOP
die_on_fail;

# cd script base dir
chdir(dirname(__FILE__));

# load cmds from platform depend config
my $cmds = LoadFile(config_file_name());

# replace <PKG> pattern with the last package in the 'install' cmd
$cmds->{install} =~ s/<PKG>/$last_package_path/;

# ready to test!
eval{
  isnt(exec_get_ec($cmds->{print_version}), 0, 'lsleases should not be installed');
  is(exec_get_ec($cmds->{install}), 0, 'installation should return exit code: 0');
  
  if($cmds->{start_after_install}){
      is(exec_get_ec($cmds->{start}), 0, 'startup should return code: 0');
      sleep(1);
  }
  
  is(exec_get_ec($cmds->{print_version}), 0, 'lsleases should be installed and running');
  is(exec_get_stdout($cmds->{list_leases}), '', 'should be empty after installation');
  send_dummy_dhcp_request();
  is(exec_get_stdout($cmds->{list_leases}), "1.1.1.1          08:00:27:f2:97:5a  testhost\n", 'dummy should be in the output');
  is(exec_get_ec($cmds->{stop}), 0, 'stopping lsleases service should return code: 0');
  sleep(1);
  isnt(exec_get_ec($cmds->{list_leases}), 0, 'list leases should return an error if no server instance running');
  is(exec_get_ec($cmds->{manpagecheck}), 0, "manpage should be installed");
  is(exec_get_ec($cmds->{remove}), 0, 'remove should return exit code: 0');
}; 
if($@){
  say "TEST FAILED - CLEANUP";
  exec_get_ec($cmds->{stop});
  sleep(1);
  exec_get_ec($cmds->{remove});
}
        
done_testing();


#
# sends an dummy dhcp request
#
sub send_dummy_dhcp_request{
    my $dummy_host_name = "testhost";
    my $host_name_option = sprintf("%02x%s", length($dummy_host_name), unpack("H*", $dummy_host_name));
    my $dummy_dhcp_request = "010106002ccab3380000000000000000000000000000000000000000080027f2975a" .
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
		"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
		"638253633501033204". "01010101" ."0c".$host_name_option."3712011c02030f06770c2c2f1a792a79f921fc2aff0000000" .
		"000000000000000000000000000";

    
    socket(SOCKET, AF_INET, SOCK_DGRAM, getprotobyname("udp")) or die "Couldn't create raw socket: $!";
    my $broadcast = sockaddr_in(67, INADDR_BROADCAST);
    setsockopt(SOCKET, SOL_SOCKET, SO_BROADCAST, 1);
    send(SOCKET, pack("H*", $dummy_dhcp_request), 0, $broadcast);
    close SOCKET;
}

#
# finds the last package (with the highest version) from local directory
#   * package file name layout: lsleases-<MAJOR>.<MINOR>(.dev)?_xxxx.....
#     
sub find_last_package_from_dir{
    my $path = shift;

    my @packages = <$path/*>;
    my $matched_package_pattern =  sprintf(".*%s.%s\$", arch(), platform_depend_package_suffix());    
    my @platform_depend_packages = grep(/$matched_package_pattern/, @packages);
    
    my $extract_version = sub {
        #
        # extract version and format as:
        #   aaaaaaaaaa.bbbbbbbbbbc
        # where
        #   aaaaaaaaaa: major version
        #   bbbbbbbbbb: minor version
        #   c         : 0 if dev version, else 1
        #
        my $file_name = shift;
        $file_name =~ /.*lsleases[-|_](\d+)\.(\d+)(\.([a-z]+))?_.*/;
        my ($major, $minor, $dev_suffix) = ($1, $2, $4);
        my $has_not_dev_suffix = (defined $dev_suffix && $dev_suffix =~ m/dev/ ? 0 : 1);
        return sprintf("%010d.%010d%1d", $major, $minor, $has_not_dev_suffix);
    };
    
    my ($last_package) = sort { $extract_version->($a) < $extract_version->($b) } @platform_depend_packages;

    die "no matching package found" if(! defined($last_package));
    return $last_package;
}

#
# finds the platform depend build artifact
#
sub find_package_from_jenkins{
    my $jenkins_url = shift;

    my $raw_json = get("$jenkins_url/lastSuccessfulBuild/api/json");
    my $json = Load($raw_json);

    # builds the absolute artifact path
    my @artifact_path = map{
        $jenkins_url . "/lastSuccessfulBuild/artifact/" . $_->{relativePath}
    } @{$json->{artifacts}};

    
    my $matched_package_pattern =  sprintf(".*%s.%s\$", arch(), platform_depend_package_suffix());
    my ($matched_package) = grep(/$matched_package_pattern/, @artifact_path);

    die "no matching package found" if(! defined($matched_package));

    my @path_segments = split(q^/^, $matched_package);
    my $file_name = $path_segments[-1];

    my $tmp_dir = tempdir(CLEANUP => 1);
    my $download_file_path = "$tmp_dir/$file_name";
    
    # download the file
    say "download $matched_package to $download_file_path ...";
    getstore($matched_package, $download_file_path) || die $!;
    
    return $download_file_path;
}

#
# executes a given command and returns the exit code
#
sub exec_get_ec {
    say "execute: " . join(", ", @_);
    my $code = system(@_);
    say "  code: $code";
    return $code;
}

#
# executes a given command and returns the stdout output
#
sub exec_get_stdout {
    say "execute: " . join(", ", @_);
    my $out = `@_`;
    say "  out: $out";
    return $out;
}


#
# returns the osname: freebsd / linux / windows
#
sub osname{
    my $osname = $Config{osname};

    return "windows" if($osname eq "MSWin32");

    return $osname;
}

#
# returns the host arch: i386 / amd64
#
sub arch{
    # $Config{archname} under windows not usefull
    #   * 32bit: $Config{longsize} == 4
    #   * 64bit: $Config{longsize} == 8
    return $Config{longsize} == 4 ? "i386" : "amd64";
}


#
# returns the platform depend package suffix: rpm / deb / txz / msi
#
sub platform_depend_package_suffix {
    return "rpm" if( -e "/etc/redhat-release");
    return "deb" if(-e "/etc/debian_version");

    my $osname = osname();
    return "txz" if($osname eq "freebsd");
    return "zip" if($osname eq "windows");

    fatal("unsupported os");   
}


#
# returns the config file name
#
sub config_file_name{
    my $platform_depend_package_suffix = platform_depend_package_suffix();
    return "linux-debian.yaml" if($platform_depend_package_suffix eq "deb");
    return "linux-redhat.yaml" if($platform_depend_package_suffix eq "rpm");

    return osname() . ".yaml";
}
