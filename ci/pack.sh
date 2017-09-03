#!/bin/bash


if [ "${GIMME_OS}" = "windows" ]; then
    zip remoteRotator-v$TRAVIS_TAG-$GIMME_OS-$GIMME_ARCH.zip remoteRotator.exe
else 
    tar -cvzf remoteRotator-v$TRAVIS_TAG-$GIMME_OS-$GIMME_ARCH.tar.gz remoteRotator
fi
