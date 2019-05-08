#!/usr/local/bin/bash
#
set -euo pipefail

ARTIFACT_NAME=$1

echo "build"
build() {
    MODULE=$1
    LDFLAGS="-X main.version=${LSLEASES_VERSION} -X github.com/j-keck/lsleases/pkg/daemon.version=${LSLEASES_VERSION}"
    GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GO111MODULES=on go build -ldflags "$LDFLAGS" -v $MODULE
}
build "github.com/j-keck/lsleases/cmd/lsleases"
build "github.com/j-keck/lsleases/cmd/lsleasesd"


echo "setup package hierarchy"
mkdir -vp usr/local/etc/rc.d usr/local/bin usr/local/man/man1

echo "copy files"
cp -v lsleases  usr/local/bin
cp -v lsleasesd usr/local/bin

cp -v lsleases.1 usr/local/man/man1
cp -v build/freebsd/lsleasesd.init usr/local/etc/rc.d/lsleasesd

echo "update version to ${LSLEASES_VERSION}"
sed -i.bak s/LSLEASES_VERSION/${LSLEASES_VERSION}/ build/freebsd/manifest/+MANIFEST

echo "create package"
pkg create -v -r . -m build/freebsd/manifest
mv -v "lsleases-${LSLEASES_VERSION}.txz" ${ARTIFACT_NAME}
