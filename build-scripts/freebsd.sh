#!/usr/local/bin/bash
#

set -euo pipefail

echo "setup package hierarchy"
mkdir -vp usr/local/etc/rc.d usr/local/bin usr/local/man/man1

echo "copy files"
cp -v lsleases usr/local/bin
cp -v lsleases.1 usr/local/man/man1
cp -v build-scripts/freebsd/lsleases.init usr/local/etc/rc.d/lsleases

echo "update version to ${LSLEASES_VERSION}"
sed -i.bak s/LSLEASES_VERSION/${LSLEASES_VERSION}/ build-scripts/freebsd/manifest/+MANIFEST

echo "create package"
pkg create -v -r . -m build-scripts/freebsd/manifest
