#!/usr/bin/env perl
#
# script builds packages under 'build-output'
#
#
# to build platform depend packages, it uses platform specific tools:
#   freebsd     : 'pkg create'
#   linux/debian: 'debhelper'
#   linux/redhat: 'rpmbuild'
#   windows     : 'NSIS'
#
# the corresponding tools need be installed.
# 
#
#
#
#
# Usage:
#   build.pl
#     * creates packages for the current platform
#
#   build.pl -commit <COMMIT>
#     * creates packages for the given <COMMIT> and current platform
#
#   build.pl -commit <COMMIT> -platform <PLATFORM> -platform <PLATFORM> ...
#     * creates packages for the given <COMMIT> and <PLATFORM>
#       <PLATFORM> can be: freebsd, linux/debian, linux/redhat or windows
#
#
# 
#
# CPAN Modules:
#  * Path::Class
#  * File::Copy::Recursive
#
#
#
# Build structure:
#
#  lsleases ($base_dir)
#    |
#    +-- build-output ($build_output): created packages
#    |
#    +-- build-work ($build_work): working directory for build
#          |
#          +-- gopath ($build_gopath): temporary GOPATH
#          |     |
#          |     +-- src
#          |          |
#          |          +-- lsleases ($build_dir): src dir
#          |          |
#          |          +-- github.com/... : dependencies
#          |
#          +-- lsleases ($package_root): package root 
#                     
#
#
#
# Exported Env per build:
#
#   BUILD_OUTPUT : build-output
#   BUILD_WORK   : build-work
#   GOPATH       : build-work/gopath
#   BUILD_DIR    : build-work/gopath/src/lsleases
#   PACKAGE_ROOT : build-work/lsleases
#   VERSION      : <VERSION>
#   GOOS         : os depend (freebsd, linux, windows)
#   GOARCH       : arch depend (386, amd64)
#   BUILD_ARCH   : arch depend (i386, amd64)
#
#
use v5.14;
use strict;
use warnings;
use diagnostics;
use Config;
use autodie;
use local::lib;
use Path::Class;
use File::Path qw/make_path remove_tree/;
use File::Copy::Recursive qw/dircopy/;
use Getopt::Long;


my $base_dir = file(__FILE__)->parent->parent->absolute;
my $build_output = "$base_dir/build-output";
my $build_work = "$base_dir/build-work";
my $build_gopath = "$build_work/gopath";
my $build_dir = "$build_gopath/src/lsleases";
my $package_root = "$build_work/lsleases";

my @target_platforms;
my $commit;
GetOptions("platform:s" => \@target_platforms,
           "commit:s"   => \$commit);
@target_platforms = (osflavor()) unless($#target_platforms >= 0);

say "=" x 80;
say "# prepare";


#
say "- create package output dir";
make_path("$build_output"); 


#
say "- recreate working dir: $build_dir";
remove_tree($build_dir);
make_path($build_dir);


#
say qq{- clone repo into "$build_dir"};
system(qq{git clone -l -q "$base_dir" "$build_dir"});

#
say "- cd $build_dir";
chdir($build_dir);

#
if(defined($commit)){
    say "- checkout $commit";
    system("git checkout $commit") && die "unable to checkout commit: $commit";
}

# extract version from new cloned project
my $version = extractVersion("$build_dir/lsleases.go");


#
set_env("BUILD_OUTPUT", $build_output);
set_env("BUILD_WORK", $build_work);
set_env("GOPATH", $build_gopath);
set_env("BUILD_DIR", $build_dir);
set_env("PACKAGE_ROOT", $package_root);
set_env("VERSION", $version); 




for my $target_platform(@target_platforms){
    my $os = shift([split(q^/^, $target_platform)]);
    set_env("GOOS", $os);

    #
    say "- get go dependencies";
    say `go get -v -d`;


    for my $arch(<i386 amd64>){
        set_env("GOARCH", ($arch eq 'i386' ? '386' : $arch));
        set_env("BUILD_ARCH", $arch);

        
        say("=" x 80);
        say("# building Version: $version, for $target_platform, arch: $arch");

        recreate_package_root();

        build_freebsd($arch)     if($target_platform eq "freebsd");
        build_debian($arch)      if($target_platform eq "linux/debian");
        build_redhat($arch    )  if($target_platform eq "linux/redhat");
        
        build_windows_zip($arch) if($target_platform eq "windows");
        build_windows_exe($arch) if($target_platform eq "windows");    
    }
}

say "=" x 80;
say "# cleanup working dir";
chdir($base_dir);
remove_tree($build_work);






#
# build freebsd package
#
sub build_freebsd{
    my $arch = shift;

    dircopy("$build_dir/build-scripts/freebsd", "freebsd");
    

    #
    say "- build code";
    system("go build -v -o $package_root/usr/local/bin/lsleases") && die "build error";

    #
    say "- generate man page";
    make_path("$package_root/usr/local/man/man1");
    system("pandoc -s -t man MANUAL.md -o $package_root/usr/local/man/man1/lsleases.1") && die "man page error";

    #
    say "- copy init script";
    make_path("$package_root/usr/local/etc/rc.d");
    system("cp -va freebsd/lsleases.init $package_root/usr/local/etc/rc.d/lsleases") && die "init script error";

    #
    say "- update version";
    system("sed -i .bak 's/version:.*/version: $version/' freebsd/manifest/+MANIFEST") && die "update version error";

    #
    say "- build package";
    system("pkg create -r $package_root -m freebsd/manifest -o $build_output") && die "packaging error";

    #
    say "- add osname / arch to package";
    system("(cd $build_output && mv -v lsleases-${version}.txz lsleases-${version}_freebsd_${arch}.txz)") && die "rename error";
}


sub build_debian{
    my $arch = shift;
    
    dircopy("$build_dir/build-scripts/debian", "debian");
 
    
    set_env("DEB_HOST_ARCH", $arch);
    set_env("DEB_BUILD_OPTIONS", "nocheck");

    #
    dircopy("debian", "$package_root/debian");

    
    #
    say ". build code";
    system("go build -v -o $package_root/usr/bin/lsleases") && die "build error";

    
    #
    say "cd $package_root";
    chdir($package_root);
    
    # 
    say "- call debian/rules";
    system("fakeroot debian/rules binary") && die "packaging error (binary)";
    system("fakeroot debian/rules clean") && die "packaging error (clean)";

    say "- cd $build_dir";
    chdir($build_dir);
}

sub build_redhat{
    my $arch = shift;

    dircopy("$build_dir/build-scripts/redhat", "redhat");

    #
    say "- build code";
    system("go build -v -o $package_root/usr/bin/lsleases") && die "build error";

    #
    say "- generate man page";
    make_path("$package_root/usr/share/man/man1");
    system("pandoc -s -t man MANUAL.md -o $package_root/usr/share/man/man1/lsleases.1") && die "man page error";

    #
    say "- copy sysvinit script";
    make_path("$package_root/etc/init.d");
    system("install -m 0755 redhat/lsleases.init $package_root/etc/init.d/lsleases") && die "init script error";

    #
    say "- copy defaults script";
    make_path("$package_root/etc/default");
    system("install -m 0755 redhat/lsleases.default $package_root/etc/default/lsleases") && die "defaults script error";

    #
    say "- call rpmbuild";
    system("rpmbuild -bb --buildroot $package_root --target $arch redhat/lsleases.spec") && die "packaging error";

    #
    say "- move package";
    system("mv $build_output/$arch/* $build_output && rmdir $build_output/$arch") && die "move package error";
}

sub convert_bat_files{
    my $dir = shift;

    for my $bat_file(<$dir/*.bat>){
      # read content
      open SRC_FH, "<", $bat_file;
      my @content = <SRC_FH>;
      close SRC_FH;

      # convert \r to \r\n
      @content = map{$_ =~ s/\R/\015\012/; $_} @content;

      # write content
      open DST_FH, ">", $bat_file;
      print DST_FH $_ for @content;
      close DST_FH;
    }

}
sub build_windows_zip{
    my $arch = shift;

    say("-" x 80);
    say("# build zip");

    dircopy("$build_dir/build-scripts/windows", "windows");

    #
    say "- build code";
    system(qq{go build -v -o "$package_root/lsleases/lsleases.exe"}) && die "build error";

    #
    say "- generate help";
    system(qq{pandoc -s -S --toc --toc-depth=1 -t html MANUAL.md -o "$package_root/lsleases/manual.html"}) && die "generate help error";

    #
    say "- copy helper scripts";
    system(qq{cp -v windows/list-leases.bat "${package_root}/lsleases"}) && die "cp list-leases.bat error";
    system(qq{cp -v windows/clear-leases.bat "${package_root}/lsleases"}) && die "cp clear-leases.bat error";
    system(qq{cp -v windows/start-server.bat "${package_root}/lsleases"}) && die "cp start-server.bat error";
    system(qq{cp -v windows/stop-server.bat "${package_root}/lsleases"}) && die "cp stop-server.bat error";
   
    #    
    say "- convert line endings from helper scripts";
    convert_bat_files("${package_root}/lsleases");
    
    #
    say "- create zip";
    chdir($package_root);
    system(qq{zip -r "${build_output}/lsleases_${version}_win_standalone_${arch}.zip" lsleases}) && die "create zip error";
    chdir($build_dir);
}


sub build_windows_exe{
    my $arch = shift;

    say("-" x 80);
    say("# build installer exe");

    dircopy("$build_dir/build-scripts/windows", "windows");


    #
    say "- build code";
    system(qq{go build -v -o "$package_root/lsleases.exe"}) && die "build error";


    #
    say "- generate help";
    system(qq{pandoc -s -S --toc --toc-depth=1 -t html MANUAL.md -o "$package_root/manual.html"}) && die "generate help error";
    

    #
    say "- copy LICENSE";
    system(qq{cp -va LICENSE "$package_root"}) && die "copy LICENSE error";

    #
    say "- copy nsis script";
    system(qq{cp -va windows/installer.nsi "$package_root"}) && die "copy nsis script error";


    #
    say "- copy helper scripts";
    system(qq{cp -v windows/list-leases.bat "${package_root}"}) && die "cp list-leases.bat error";
    system(qq{cp -v windows/clear-leases.bat "${package_root}"}) && die "cp clear-leases.bat error";
    system(qq{cp -v windows/start-server.bat "${package_root}"}) && die "cp start-server.bat error";
    system(qq{cp -v windows/stop-server.bat "${package_root}"}) && die "cp stop-server.bat error";
    system(qq{cp -v windows/start-service.bat "${package_root}"}) && die "cp start-service.bat error";
    system(qq{cp -v windows/stop-service.bat "${package_root}"}) && die "cp stop-service.bat error";
    system(qq{cp -v windows/restart-service.bat "${package_root}"}) && die "cp restart-service.bat error";        
   
    #    
    say "- convert line endings from helper scripts";
    convert_bat_files("${package_root}/lsleases");
    

    #
    say "- copy nssm.exe (service wrapper)";
    system(qq{cp -va windows/${arch}/nssm.exe "$package_root"}) && die "copy nssm error";

    
    #
    say "- create package";
    chdir($package_root);
    my $makensis_flag = (osflavor() eq "windows" ? "/NOCD" : "-NOCD");
    system(qq{makensis $makensis_flag "$build_dir/windows/installer.nsi"}) && die "nsis error";
    chdir($build_dir);
}








#
# log and set env
#
sub set_env{
    my $name = shift;
    my $value = shift;

    say "- set env: $name = $value";
    $ENV{$name} = $value;
}

#
#
#
sub recreate_package_root{
    say "- recreate package root: $package_root";
    remove_tree($package_root);
    make_path($package_root);
}



#
# extracts the version from lsleases.go
#
sub extractVersion {
    my $file = shift;
    open(my $fh, "<$file");
    my ($version_line) = grep /.*VERSION.*/, <$fh>;
    close($fh);

    die "version line not found" if(! defined $version_line);
    $version_line =~ /VERSION\s*=\s*"(.*)"/;
    die "version not found" if(! defined $1);

    return $1;
}

#
# returns the osflavor: freebsd, windows, linux/debian, linux/redhat
#
sub osflavor{
    my $osname = $Config{osname};

    return "windows" if($osname eq "MSWin32");
    return "freebsd" if($osname eq "freebsd");

    if($osname eq "linux"){
        return "linux/redhat" if( -e "/etc/redhat-release");
        return "linux/debian" if(-e "/etc/debian_version");        
    }

    die "unsupported platform: $osname";
}
