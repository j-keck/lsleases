#!/bin/sh
#

BUILD_ARCH=$1
case $BUILD_ARCH in
    i386)
        export GOARCH=386
        ;;
    amd64)
        export GOARCH=amd64
        ;;
    *)
        echo "unsupported arch"
        exit 1
esac
export GOOS=freebsd

BUILD_ROOT=$BUILD_DIR/$GOOS/root/$BUILD_ARCH

cd $BUILD_DIR

# delete old build artifacts
rm -rf $BUILD_ROOT

# build code
go build -v -o $BUILD_ROOT/usr/local/bin/lsleases

# build manpage
mkdir -p $BUILD_ROOT/usr/local/man/man1
pandoc -s -t man MANUAL.md -o $BUILD_ROOT/usr/local/man/man1/lsleases.1

# copy init script
mkdir -p $BUILD_ROOT/usr/local/etc/rc.d
cp freebsd/lsleases.init $BUILD_ROOT/usr/local/etc/rc.d/lsleases

# create package
echo "create package in $PACKAGE_DIR/$BUILD_ARCH"
pkg create -r $BUILD_ROOT -m freebsd/manifest -o $PACKAGE_DIR/$BUILD_ARCH

