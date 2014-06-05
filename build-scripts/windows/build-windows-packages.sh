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
export GOOS=windows

BUILD_ROOT=$BUILD_DIR/$GOOS/root/$BUILD_ARCH/lsleases

cd $BUILD_DIR

# delete old build artifacts
rm -rf $BUILD_ROOT

# build code
go build -v -o $BUILD_ROOT/lsleases.exe

# build help
pandoc -s -t html MANUAL.md -o $BUILD_ROOT/manual.html
pandoc MANUAL.md -o $BUILD_ROOT/manual.txt

# copy [start|stop]-server / list-leases scripts
cp windows/start-server.bat $BUILD_ROOT
cp windows/stop-server.bat $BUILD_ROOT
cp windows/list-leases.bat $BUILD_ROOT

# create zip
echo "create zip in $PACKAGE_DIR/$BUILD_ARCH"
(cd $BUILD_ROOT/..; zip -r $PACKAGE_DIR/$BUILD_ARCH/lsleases_$VERSION.$BUILD_ARCH.zip lsleases)

