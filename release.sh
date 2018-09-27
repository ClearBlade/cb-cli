#!/bin/bash

if [ $# -ne 1 ] ; then
    echo "wrong number of arguments"
    exit 1
fi

version=$1

mkdir -p release_binaries
chmod 777 release_binaries
rm -f release_binaries/*
export PATH=$PATH:$GOPATH/bin

#
#  Should put all this in a fancy loop, but since there's only
#  five or so iterations, why bother. -swm
#

################################################################################

echo -n "Building linux/amd64... "
export GOOS=linux
export GOARCH=amd64
go build > build.out 2>&1
if [ $? -ne 0 ] ; then
    echo "build failed"
    cat build.out
    exit 1
fi
tar zcvf release_binaries/cb-cli-${version}-linux-amd64.tar.gz cb-cli README.md LICENSE > tar.out 2>&1
if [ $? -ne 0 ] ; then
    echo "tar failed:"
    cat tar.out
    exit 1
fi
echo "Success"

################################################################################

echo -n "Building linux/arm64... "
export GOOS=linux
export GOARCH=arm64
go build > build.out 2>&1
if [ $? -ne 0 ] ; then
    echo "build failed"
    cat build.out
    exit 1
fi
tar zcvf release_binaries/cb-cli-${version}-linux-arm64.tar.gz cb-cli README.md LICENSE > tar.out 2>&1
if [ $? -ne 0 ] ; then
    echo "tar failed:"
    cat tar.out
    exit 1
fi
echo "Success"

################################################################################

echo -n "Building linux/arm32... "
export GOOS=linux
export GOARCH=arm
go build > build.out 2>&1
if [ $? -ne 0 ] ; then
    echo "build failed"
    cat build.out
    exit 1
fi
tar zcvf release_binaries/cb-cli-${version}-linux-arm32.tar.gz cb-cli README.md LICENSE > tar.out 2>&1
if [ $? -ne 0 ] ; then
    echo "tar failed:"
    cat tar.out
    exit 1
fi
echo "Success"

################################################################################

echo -n "Building linux/386... "
export GOOS=linux
export GOARCH=386
go build > build.out 2>&1
if [ $? -ne 0 ] ; then
    echo "build failed"
    cat build.out
    exit 1
fi
tar zcvf release_binaries/cb-cli-${version}-linux-386.tar.gz cb-cli README.md LICENSE > tar.out 2>&1
if [ $? -ne 0 ] ; then
    echo "tar failed:"
    cat tar.out
    exit 1
fi
echo "Success"

################################################################################

echo -n "Building MacOSX/64... "
export GOOS=darwin
export GOARCH=amd64
go build > build.out 2>&1
if [ $? -ne 0 ] ; then
    echo "build failed"
    cat build.out
    exit 1
fi
tar zcvf release_binaries/cb-cli-${version}-MacOSX64.tar.gz cb-cli README.md LICENSE > tar.out 2>&1
if [ $? -ne 0 ] ; then
    echo "tar failed:"
    cat tar.out
    exit 1
fi
echo "Success"

################################################################################

echo -n "Building Windows/64... "
export GOOS=windows
export GOARCH=amd64
go build -o cb-cli.exe > build.out 2>&1
if [ $? -ne 0 ] ; then
    echo "build failed"
    cat build.out
    exit 1
fi
echo "Done"

echo -n "Zipping Windows binaries... "

zip release_binaries/cb-cli-${version}-Windows.zip cb-cli.exe README.md LICENSE > zip.out 2>&1
if [ $? -ne 0 ] ; then
    echo "zip failed:"
    cat tar.out
    exit 1
fi
chmod 666 release_binaries/*
echo "Success"

