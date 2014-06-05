#!/bin/sh
#
# script generate platform depend packages 
#


THIS_SCRIPT_PATH=$(readlink -f "$0")
# export env vars for slave scripts
export BASE_DIR=$(dirname $THIS_SCRIPT_PATH)
export PACKAGE_DIR=$BASE_DIR/build-output
export BUILD_DIR=$BASE_DIR/build-work


# create package dir
test -d $PACKAGE_DIR/i386 || mkdir -p $PACKAGE_DIR/i386
test -d $PACKAGE_DIR/amd64 || mkdir -p $PACKAGE_DIR/amd64

# clone into working dir
git ls-remote -h $BUILD_DIR >/dev/null 2>&1 || git clone . $BUILD_DIR

case "$1" in
    Linux)
        cp -a build-scripts/debian $BUILD_DIR
        cp -a build-scripts/rpmbuild $BUILD_DIR

        echo 
        echo '###############################################################################################################'
        echo '# 32bit rpm'
        (cd $BUILD_DIR; rpmbuild -bb --target i386 rpmbuild/SPECS/lsleases.spec)

        echo 
        echo '###############################################################################################################'
        echo '# 64bit rpm'
        (cd $BUILD_DIR; rpmbuild -bb --target amd64 rpmbuild/SPECS/lsleases.spec)



        echo 
        echo '###############################################################################################################'
        echo '# 32bit deb'
        export GOARCH=386
        export DEB_HOST_ARCH=i386
        export DEB_BUILD_OPTIONS=nocheck
        (cd $BUILD_DIR; fakeroot $BUILD_DIR/debian/rules binary)
        (cd $BUILD_DIR; fakeroot $BUILD_DIR/debian/rules clean)

        echo 
        echo '###############################################################################################################'
        echo '# 64bit deb'
        export GOARCH=amd64
        export DEB_HOST_ARCH=amd64
        export DEB_BUILD_OPTIONS=nocheck
        (cd $BUILD_DIR; fakeroot $BUILD_DIR/debian/rules binary)
        (cd $BUILD_DIR; fakeroot $BUILD_DIR/debian/rules clean)

        ;;
    FreeBSD)
        cp -a build-scripts/freebsd $BUILD_DIR

        echo 
        echo '###############################################################################################################'
        echo '# 32bit pkg'
        $BUILD_DIR/freebsd/build-freebsd-packages.sh i386

        echo 
        echo '###############################################################################################################'
        echo '# 64bit pkg'
        $BUILD_DIR/freebsd/build-freebsd-packages.sh amd64
        
        ;;
    Windows)
        cp -a build-scripts/windows $BUILD_DIR

        echo 
        echo '###############################################################################################################'
        echo '# 32bit exe'
        $BUILD_DIR/windows/build-windows-packages.sh i386

        echo 
        echo '###############################################################################################################'
        echo '# 64bit exe'
        $BUILD_DIR/windows/build-windows-packages.sh amd64
        ;;
 
    *)
        echo "unsupported platform - supported: [Linux,FreeBSD|Windows]"
esac

rm -rf $BUILD_DIR

