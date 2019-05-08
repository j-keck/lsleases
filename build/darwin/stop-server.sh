#!/bin/sh
BASE_PATH=$(dirname $0)
${BASE_PATH}/lsleases -x
if [ $? != 0 ]; then
    echo "<HIT ENTER>"
    read DUMMY
fi
