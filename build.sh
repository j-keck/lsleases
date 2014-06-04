#!/bin/sh

buildLinuxPackages(){
  echo '#################################################################################'
  echo 32bit rpm
  rpmbuild -bb --target i386 rpmbuild/SPECS/lsleases.spec

  echo '#################################################################################'
  echo 64bit rpm
  rpmbuild -bb --target amd64 rpmbuild/SPECS/lsleases.spec


  echo '#################################################################################'
  echo 32bit deb
  export GOARCH=386
  export DEB_HOST_ARCH=i386
  fakeroot ./debian/rules binary
  fakeroot ./debian/rules clean


  echo '#################################################################################'
  echo 64bit deb
  export GOARCH=amd64
  export DEB_HOST_ARCH=amd64
  fakeroot ./debian/rules binary
  fakeroot ./debian/rules clean
}

buildFreeBSDPackages(){
  echo '#################################################################################'
  echo "32bit pkg"   
}

case $(uname) in
    Linux)
      buildLinuxPackages
      ;;
    FreeBSD)
        buildFreeBSDPackages
        ;;
    *)
        echo "unsupported platform"
esac


